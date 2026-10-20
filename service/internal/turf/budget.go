package turf

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Limits are the Turf API's published rate limits, as this process honours
// them (`NFR-001`).
//
// NEITHER FIGURE IS WRITTEN IN THIS MODULE AND NEITHER HAS A DEFAULT. Both reach
// the client from configuration, and each has one home:
//
//   - Request is the general limit, which `Architecture.md § Data sources and
//     constraints` states per resource and for the whole API.
//   - AllZones is the all-zones endpoint's own, tighter limit, which
//     `Architecture.md § Retrieving zones` records with that endpoint.
//
// Each is held as the shortest interval permitted between two requests, which
// is the form a limit of one request per period takes.
type Limits struct {
	Request  time.Duration
	AllZones time.Duration
}

// errNotPermitted is returned, unsent, by a request the limits will not permit
// before its own deadline.
var errNotPermitted = errors.New("the Turf API's published limit does not permit this request before its deadline, so it was not sent")

// budget is THE PROCESS'S ONE BUDGET AGAINST THE TURF API, and it is a package
// variable because the limit is the API's and not any caller's.
//
// `NFR-001`'s Rationale is the whole argument: a limiter held per client
// instance, per session or per worker satisfies each caller and breaches the
// limit as soon as two of them run at once. A budget each Client constructed
// for itself would be exactly that, and a budget handed to each Client by the
// composition root would hold only for as long as the root never built two.
// Every Client in the process draws from this one, however many there are and
// whoever built them, which is what makes the sync and any request-path caller
// count against the same limit. `DEPLOYMENT.md` models one service instance, so
// the process is the whole system the limit is held across.
var budget = &registry{gates: make(map[gateKey]*gate)}

// gateKey names what one gate meters. path is empty for the gate every request
// to a host passes through, and is the endpoint's path for the all-zones gate.
type gateKey struct {
	host string
	path string
}

type registry struct {
	mu    sync.Mutex
	gates map[gateKey]*gate
}

// gate returns the process's gate for key, creating it on first use.
//
// THE LONGEST INTERVAL ASKED FOR WINS. Two Clients configured differently for
// one key would otherwise each be held to their own interval through a gate
// they share, and the shorter one would decide when the next request goes. The
// longer interval satisfies both.
func (r *registry) gate(key gateKey, interval time.Duration) *gate {
	r.mu.Lock()
	defer r.mu.Unlock()

	g, ok := r.gates[key]
	if !ok {
		g = &gate{slot: make(chan struct{}, 1)}
		r.gates[key] = g
	}

	if int64(interval) > g.interval.Load() {
		g.interval.Store(int64(interval))
	}

	return g
}

// gate admits one request at a time, and each no sooner than its interval after
// the previous one RETURNED.
//
// MEASURED FROM THE RETURN, NOT THE SEND, so that the spacing holds where the
// API measures it. The slot is held from the wait until the response headers
// arrive, and the next request may not leave until the interval has passed
// since then — so it reaches the API more than the interval after the previous
// one did, whatever the scheduler did with either goroutine. Measuring from the
// moment a request was permitted to leave would leave the gap between permission
// and sending, which no lock can see, on the wrong side of the limit.
//
// The cost is that requests through one gate never overlap before their
// headers arrive. Against a limit of one request per interval that is a
// property, not a loss.
type gate struct {
	// slot has capacity one; holding it is the right to the next request. Its
	// send and receive are also what order one holder's write to last before
	// the next holder's read.
	slot chan struct{}

	interval atomic.Int64

	// last is when the previous request through this gate returned. Only the
	// slot's holder reads or writes it.
	last time.Time
}

// take waits until the gate permits a request and returns the function that
// records the request as returned and admits the next one.
//
// A WAIT THAT WOULD OUTLAST ctx'S DEADLINE IS REFUSED AT ONCE rather than slept
// through, and nothing is sent. The request could not be made in time anyway,
// and refusing now hands its caller the reason instead of a deadline error that
// reads as a slow API.
func (g *gate) take(ctx context.Context, what string) (func(), error) {
	// First, because a select with both cases ready picks either: a free slot
	// would otherwise admit a request whose caller has already given up.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	select {
	case g.slot <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	wait := time.Until(g.last.Add(time.Duration(g.interval.Load())))

	if wait > 0 {
		if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) < wait {
			<-g.slot

			return nil, fmt.Errorf("%w: %s permits the next one in %s", errNotPermitted, what, wait.Round(time.Millisecond))
		}

		timer := time.NewTimer(wait)

		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			<-g.slot

			return nil, ctx.Err()
		}
	}

	return func() {
		g.last = time.Now()
		<-g.slot
	}, nil
}

// limitedTransport passes every request the client sends through the process's
// budget before it reaches the network.
//
// IT IS THE TRANSPORT, NOT THE FETCH, THAT IS METERED, because the transport is
// where a request becomes one. A redirect the client follows is a second
// request to the API, issued inside one call to http.Client.Do; a limit taken
// once per call would let that hop through unmetered.
type limitedTransport struct {
	base   http.RoundTripper
	limits Limits

	// allZones is the configured all-zones endpoint. A request addressed to it
	// passes the all-zones gate as well as the API's. It is the zero key for a
	// client configured with no all-zones endpoint, and no request matches that
	// key, because every request names a host.
	allZones gateKey
}

// newLimitedTransport meters base. allZones is nil for a client configured with
// no all-zones endpoint, such as the SessionClient.
func newLimitedTransport(base http.RoundTripper, allZones *url.URL, limits Limits) *limitedTransport {
	// Never wrapped twice. A second layer would take the same gate's slot while
	// the first layer holds it, and wait for itself for ever.
	if inner, ok := base.(*limitedTransport); ok {
		base = inner.base
	}

	if base == nil {
		base = http.DefaultTransport
	}

	if tr, ok := base.(*http.Transport); ok {
		base = unretried(tr)
	}

	t := &limitedTransport{base: base, limits: limits}
	if allZones != nil {
		t.allZones = gateKey{host: hostKey(allZones), path: allZones.Path}
	}

	return t
}

// unretried returns a copy of tr that never re-sends a request it has written.
//
// A RE-SEND INSIDE THE TRANSPORT IS A REQUEST THE BUDGET NEVER SAW (#234, N1).
// net/http re-sends an idempotent GET, within the one RoundTrip limitedTransport
// meters, when a REUSED HTTP/1 connection drops after the server has read the
// request; its HTTP/2 transport re-sends on a fresh connection after some
// stream and connection errors. Either puts a second request on the wire
// inside one pass through the gate. With no connection kept alive there is no
// reused connection to re-send on, and with HTTP/1.1 alone there is no HTTP/2
// re-send. A request the transport never wrote is still retried, and never
// reached the API.
//
// The cost is a new connection, and its handshake, for every request, which a
// limit of one request per interval makes a small price.
//
// Only an *http.Transport can be configured so. Every client this package
// builds passes http.DefaultTransport; a test that hands setHTTPClient its
// server's client passes an *http.Transport too.
func unretried(tr *http.Transport) *http.Transport {
	c := tr.Clone()
	c.DisableKeepAlives = true

	var http1 http.Protocols
	http1.SetHTTP1(true)
	c.Protocols = &http1

	// A TLS configuration that offers h2 itself would have the server choose
	// it, and an HTTP/1-only transport cannot speak it. Clone copied the
	// configuration, so clearing the offer here changes no one else's.
	if c.TLSClientConfig != nil {
		c.TLSClientConfig.NextProtos = nil
	}

	return c
}

// RoundTrip takes the all-zones gate, when the request is addressed to that
// endpoint, and then the API's.
//
// ALWAYS IN THAT ORDER. A request holding the API's gate never waits on
// another, so no two requests can each hold what the other is waiting for.
func (t *limitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	host := hostKey(req.URL)

	if (gateKey{host: host, path: req.URL.Path}) == t.allZones {
		done, err := budget.gate(t.allZones, t.limits.AllZones).take(ctx, "the all-zones endpoint")
		if err != nil {
			return nil, refuse(req, err)
		}

		defer done()
	}

	done, err := budget.gate(gateKey{host: host}, t.limits.Request).take(ctx, "the API at "+host)
	if err != nil {
		return nil, refuse(req, err)
	}

	defer done()

	return t.base.RoundTrip(req)
}

// refuse closes the body of a request that will not be sent, which the
// http.RoundTripper contract requires of every path that returns an error.
func refuse(req *http.Request, err error) error {
	if req.Body != nil {
		_ = req.Body.Close()
	}

	return err
}

// hostKey is the host a request is metered against.
//
// THE WHOLE API SHARES ONE GATE, AND THAT IS A READING, RECORDED AS ONE. The
// general limit is stated per resource, and neither the API's own documentation
// nor `Architecture.md § Data sources and constraints` says what one resource
// is — an endpoint, a path, or something wider. One gate for every request to
// the API holds the limit under each of those readings, where a gate per path
// would breach it under the wider ones. What that costs is concurrency between
// requests to different endpoints, which is a latency and never a block.
// Narrowing it is this function and the key it returns, once the meaning is
// verified rather than inferred.
//
// The port is kept, so two servers on one host are two budgets, and a default
// port is dropped, so naming it does not make a second budget for one server.
func hostKey(u *url.URL) string {
	host := strings.ToLower(u.Hostname())

	switch port := u.Port(); {
	case port == "",
		u.Scheme == "https" && port == "443",
		u.Scheme == "http" && port == "80":
		return host
	default:
		return host + ":" + port
	}
}
