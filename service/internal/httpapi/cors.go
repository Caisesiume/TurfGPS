package httpapi

import (
	"net/http"
	"slices"
	"strings"
)

// The grant a preflight may carry, held here as explicit lists rather than
// echoed from what the preflight asks for (#224 AC2). They are what the
// client's calls use and nothing more: its one call today is a GET that sets no
// request header of its own (`web/src/api.ts`, #26), so the header list is
// empty. A client call that needs another method or header extends these lists
// in the same change, and a preflight asking for anything else is refused.
var (
	corsMethods        = []string{http.MethodGet}
	corsRequestHeaders = []string{}
)

// NewHandler returns the request surface of NewMux with cross-origin access
// granted to allowedOrigins and to no other origin.
//
// The client is deployed on Vercel and calls the API on its own subdomain,
// never through a Vercel rewrite (`docs/adr/ADR-0006-deployment-topology.md`
// § Decision, D4), so its calls are cross-origin and the browser reads none of
// their responses without this grant. The headers are set here and not by the
// proxy: a second `Access-Control-Allow-Origin` value fails the browser's
// check, so exactly one of the two may set them, and #224 places them in the
// service.
//
// allowedOrigins are taken as given, in the exact form a browser sends in
// `Origin`; `config.LoadAllowedOrigins` is what holds them to it. An empty list
// grants nothing to anyone, and `*` is never granted even if it is listed.
//
// NOTHING ON THIS PATH LOGS. A plan code travels in the request target
// (`DEPLOYMENT.md § Where the deployment configuration lives`) and a preflight
// carries the same target as the call it precedes, so the refusals below are
// answered without a word to the log rather than reported in one.
//
// opts are NewMux's. `POST /journeys` is not on the grant lists above because
// no client call makes it yet; the call that does adds POST and
// `Content-Type` to them in the same change, as those lists require.
func NewHandler(allowedOrigins []string, opts ...Option) http.Handler {
	origins := make(map[string]struct{}, len(allowedOrigins))

	for _, origin := range allowedOrigins {
		origins[origin] = struct{}{}
	}

	return &cors{next: NewMux(opts...), origins: origins}
}

type cors struct {
	next    http.Handler
	origins map[string]struct{}
}

func (c *cors) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// On every response, granted or not: whether a response carries a grant
	// depends on the request's Origin, so a cache must not hand one origin's
	// answer to another.
	w.Header().Add("Vary", "Origin")

	origin, granted := c.grantedOrigin(r)

	if r.Method == http.MethodOptions && origin != "" && r.Header.Get("Access-Control-Request-Method") != "" {
		c.preflight(w, r, origin, granted)

		return
	}

	if granted {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	c.next.ServeHTTP(w, r)
}

// grantedOrigin returns the request's origin and whether it is one this
// service grants. A request carrying more than one Origin value is not a
// browser's, and is granted nothing.
func (c *cors) grantedOrigin(r *http.Request) (string, bool) {
	values := r.Header.Values("Origin")
	if len(values) != 1 {
		return "", false
	}

	origin := values[0]
	if origin == "*" {
		return origin, false
	}

	_, ok := c.origins[origin]

	return origin, ok
}

// preflight answers a CORS preflight itself, so that no handler behind it ever
// sees one. It grants only when the origin is granted AND every method and
// request header asked for is on the lists above; otherwise it refuses with no
// `Access-Control-Allow-*` header at all, rather than granting part of what was
// asked.
func (c *cors) preflight(w http.ResponseWriter, r *http.Request, origin string, granted bool) {
	h := w.Header()
	h.Add("Vary", "Access-Control-Request-Method")
	h.Add("Vary", "Access-Control-Request-Headers")

	if !granted || !slices.Contains(corsMethods, r.Header.Get("Access-Control-Request-Method")) || !requestHeadersGranted(r) {
		w.WriteHeader(http.StatusForbidden)

		return
	}

	h.Set("Access-Control-Allow-Origin", origin)
	h.Set("Access-Control-Allow-Methods", strings.Join(corsMethods, ", "))

	if len(corsRequestHeaders) > 0 {
		h.Set("Access-Control-Allow-Headers", strings.Join(corsRequestHeaders, ", "))
	}

	w.WriteHeader(http.StatusNoContent)
}

// requestHeadersGranted reports whether every header name a preflight asks for
// is on corsRequestHeaders. Header names are case-insensitive.
func requestHeadersGranted(r *http.Request) bool {
	for _, line := range r.Header.Values("Access-Control-Request-Headers") {
		for name := range strings.SplitSeq(line, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}

			if !slices.ContainsFunc(corsRequestHeaders, func(allowed string) bool { return strings.EqualFold(allowed, name) }) {
				return false
			}
		}
	}

	return true
}
