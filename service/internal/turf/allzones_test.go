package turf

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func clientFor(t *testing.T, srv *httptest.Server, maxBytes int64) *Client {
	t.Helper()

	c, err := NewClient(srv.URL, maxBytes)
	if err != nil {
		t.Fatalf("building the client: %v", err)
	}

	c.setHTTPClient(srv.Client())

	return c
}

// TestASuccessfulFetchReturnsTheBodyAndItsStatus is the ordinary path.
func TestASuccessfulFetchReturnsTheBodyAndItsStatus(t *testing.T) {
	t.Parallel()

	const payload = `[{"id":1}]`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q, want application/json", got)
		}

		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	body, status, err := clientFor(t, srv, 1<<20).FetchAllZones(t.Context())
	if err != nil {
		t.Fatalf("fetching: %v", err)
	}

	if status != http.StatusOK {
		t.Errorf("status = %d, want 200", status)
	}

	if string(body) != payload {
		t.Errorf("body = %q, want %q", body, payload)
	}
}

// TestARefusedRequestCarriesItsStatusAndItsBody binds what
// `Architecture.md § The sync write path` asks http_status and response_bytes to
// do: separate a refused request from a body that would not parse. Both are
// http_error, and without the status the two are the same row.
func TestARefusedRequestCarriesItsStatusAndItsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer srv.Close()

	body, status, err := clientFor(t, srv, 1<<20).FetchAllZones(t.Context())
	if err == nil {
		t.Fatal("a 429 was returned as a success, want an error")
	}

	if status != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429 — a refused request is recognisable only by its status", status)
	}

	if len(body) == 0 {
		t.Error("the body was discarded, want it returned so response_bytes can be recorded")
	}

	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("the error reads %q, want it to quote what the endpoint said", err)
	}
}

// TestAnOversizedResponseIsRefusedAndNotTruncated is the one that matters most
// here.
//
// A body cut at the ceiling is either unparseable or — for a shape that happens
// to close cleanly — a short response that looks complete, which is exactly the
// truncated response the staging assertions exist to catch. Manufacturing one
// would be this service producing the failure it defends against.
//
// WHAT CARRIES THAT GUARANTEE IS THE ERROR, and this test asserts it rather than
// asserting a nil body. The refusal used to return no body at all, which left
// `response_bytes` NULL on the one run whose failure cause IS its size: the row
// an operator reads said a fetch failed and withheld the only figure that says
// why. So the bytes taken to establish the overrun come back, and what stops
// them reaching a parse is that err is non-nil — the same contract the refused
// request above already runs on, where the body is returned for exactly this
// reason. A caller cannot mistake them for an accepted response either: their
// length is past the ceiling, which no accepted body's ever is.
func TestAnOversizedResponseIsRefusedAndNotTruncated(t *testing.T) {
	t.Parallel()

	const ceiling = 64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", ceiling*4)))
	}))
	defer srv.Close()

	body, _, err := clientFor(t, srv, ceiling).FetchAllZones(t.Context())
	if err == nil {
		t.Fatal("an oversized response was accepted, want it refused")
	}

	if int64(len(body)) <= ceiling {
		t.Errorf("the refusal returned %d bytes, want more than the %d byte ceiling: response_bytes is recorded from this length, and a figure at or under the ceiling would read on the row as a response that fitted",
			len(body), ceiling)
	}
}

// TestTheCallerDeadlineBoundsTheFetch binds the constraint that every external
// call is bounded, and that the bound is the caller's rather than one this
// package chose. A hung endpoint must not hold the sync open.
func TestTheCallerDeadlineBoundsTheFetch(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	defer close(release)

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	done := make(chan struct{})

	go func() {
		defer close(done)

		if _, _, err := clientFor(t, srv, 1<<20).FetchAllZones(ctx); err == nil {
			t.Error("a hung endpoint returned a success, want the deadline to end the call")
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the fetch outlived its deadline by seconds, so nothing bounds it")
	}
}

// TestAnEndpointIsRequired refuses a client that could not fetch anything, at
// construction rather than on the first tick.
func TestAnEndpointIsRequired(t *testing.T) {
	t.Parallel()

	if _, err := NewClient("", 1<<20); err == nil {
		t.Error("a client with no endpoint was built, want it refused")
	}

	if _, err := NewClient("https://example.test/zones", 0); err == nil {
		t.Error("a client with no response ceiling was built, want it refused")
	}

	// The other end of the same guard. A ceiling this size is not a large
	// ceiling, it is one read cannot compute with — see maxCeiling — and the
	// two failures it produces are a panic and a fetch that succeeds over an
	// empty body, neither of which is reachable from a constructor that
	// refuses it.
	if _, err := NewClient("https://example.test/zones", math.MaxInt64); err == nil {
		t.Error("a client with a response ceiling above what it can read within was built, want it refused")
	}
}

// TestAnHTTPSToHTTPRedirectIsRefused is the abuse case for the transport.
//
// net/http's default redirect policy follows https to http without comment, so
// an endpoint configured as https was only as encrypted as whatever answered
// it: an origin that has been compromised, hijacked by DNS, or merely
// misconfigured answers one 302 and the client fetches the corpus in the clear.
// Nothing downstream would catch it. The staging assertions of
// `Architecture.md § The sync write path` check the staged row count and the
// coordinate ranges, so a corpus rewritten in flight to merely plausible
// coordinates is staged, asserted, merged, and becomes the authoritative
// geometry every later query resolves against.
//
// WHAT THIS ASSERTS IS THAT THE PLAINTEXT BODY NEVER ARRIVES, not merely that an
// error is returned. An error beside a body that was fetched anyway would leave
// the bytes one careless caller away from the parse, so the payload the plain
// server holds is the thing the test looks for and must not find.
func TestAnHTTPSToHTTPRedirectIsRefused(t *testing.T) {
	t.Parallel()

	const substituted = `[{"id":1,"latitude":59.0,"longitude":18.0}]`

	reached := make(chan struct{}, 1)

	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		select {
		case reached <- struct{}{}:
		default:
		}

		_, _ = w.Write([]byte(substituted))
	}))
	defer plain.Close()

	secure := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, plain.URL, http.StatusFound)
	}))
	defer secure.Close()

	body, _, err := clientFor(t, secure, 1<<20).FetchAllZones(t.Context())
	if err == nil {
		t.Fatal("the downgrade to http was followed and the response accepted, want the redirect refused")
	}

	if strings.Contains(string(body), substituted) {
		t.Errorf("the plaintext body came back with the error: %q. An error beside the bytes still hands them to any caller that reads the body before the error", body)
	}

	select {
	case <-reached:
		t.Error("the plaintext endpoint was contacted, want the redirect refused before the request was made: the corpus must not cross the network in the clear even if the body is then discarded")
	default:
	}
}

// TestTheRedirectPolicyDecides pins the policy's decision table directly,
// because the end-to-end test above can only demonstrate the refusal and would
// pass just as well against a policy that refused every redirect — including
// the ordinary https-to-https one an endpoint uses to normalise a path.
//
// The hop limit is here for a reason that is easy to lose: a non-nil
// CheckRedirect REPLACES net/http's default policy rather than adding to it, so
// installing one that returned nil for every https hop would have removed the
// ten-redirect limit that was already in force and answered a downgrade with an
// unbounded loop.
func TestTheRedirectPolicyDecides(t *testing.T) {
	t.Parallel()

	hop := func(t *testing.T, raw string) *http.Request {
		t.Helper()

		req, err := http.NewRequest(http.MethodGet, raw, nil)
		if err != nil {
			t.Fatalf("building the hop %q: %v", raw, err)
		}

		return req
	}

	t.Run("an https hop is followed", func(t *testing.T) {
		t.Parallel()

		if err := httpsOnly(hop(t, "https://example.test/elsewhere"), nil); err != nil {
			t.Errorf("an https redirect was refused with %v, want it followed: refusing every redirect is not the guarantee this policy exists to give", err)
		}
	})

	for name, raw := range map[string]string{
		"http":           "http://example.test/elsewhere",
		"an odd scheme":  "gopher://example.test/elsewhere",
		"a schemeless h": "//example.test/elsewhere",
	} {
		t.Run(name+" is refused", func(t *testing.T) {
			t.Parallel()

			if err := httpsOnly(hop(t, raw), nil); err == nil {
				t.Errorf("the redirect to %q was followed, want it refused: the policy fails closed on anything not known to be encrypted", raw)
			}
		})
	}

	t.Run("the chain is finite", func(t *testing.T) {
		t.Parallel()

		via := make([]*http.Request, maxRedirects)
		if err := httpsOnly(hop(t, "https://example.test/elsewhere"), via); err == nil {
			t.Errorf("a chain of %d https hops was extended, want it refused: a non-nil CheckRedirect replaces net/http's default limit rather than adding to it, so this policy has to carry the bound itself", maxRedirects)
		}
	})
}
