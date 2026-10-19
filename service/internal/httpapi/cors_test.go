package httpapi

import (
	"bytes"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
)

const (
	clientOrigin  = "https://turfgps.example"
	previewOrigin = "https://preview.turfgps.example:8443"
)

func serveCORS(t *testing.T, origins []string, method, target string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, target, nil)
	for name, values := range header {
		for _, v := range values {
			req.Header.Add(name, v)
		}
	}

	rec := httptest.NewRecorder()
	NewHandler(origins).ServeHTTP(rec, req)

	return rec
}

func preflightHeader(origin, method, headers string) http.Header {
	h := http.Header{"Origin": {origin}, "Access-Control-Request-Method": {method}}
	if headers != "" {
		h.Set("Access-Control-Request-Headers", headers)
	}

	return h
}

// grants lists every Access-Control-Allow-* header a response carries.
func grants(rec *httptest.ResponseRecorder) []string {
	var found []string

	for name, values := range rec.Header() {
		if strings.HasPrefix(http.CanonicalHeaderKey(name), "Access-Control-Allow-") {
			found = append(found, name+": "+strings.Join(values, ", "))
		}
	}

	return found
}

func varies(rec *httptest.ResponseRecorder, name string) bool {
	for _, line := range rec.Header().Values("Vary") {
		for v := range strings.SplitSeq(line, ",") {
			if strings.EqualFold(strings.TrimSpace(v), name) {
				return true
			}
		}
	}

	return false
}

// TestAC1AConfiguredOriginIsEchoedAndNeverWildcarded binds #224 AC1: a request
// from each configured origin is granted to that origin by name, never by `*`,
// and says so in Vary.
func TestAC1AConfiguredOriginIsEchoedAndNeverWildcarded(t *testing.T) {
	t.Parallel()

	origins := []string{clientOrigin, previewOrigin}

	for _, origin := range origins {
		t.Run(origin, func(t *testing.T) {
			t.Parallel()

			rec := serveCORS(t, origins, http.MethodGet, "/", http.Header{"Origin": {origin}})

			if got := rec.Header().Values("Access-Control-Allow-Origin"); !slices.Equal(got, []string{origin}) {
				t.Errorf("Access-Control-Allow-Origin = %q, want exactly [%q]: the configured origin, echoed, and never *", got, origin)
			}

			if !varies(rec, "Origin") {
				t.Errorf("Vary = %q, want it to name Origin: the grant depends on the request's origin", rec.Header().Values("Vary"))
			}

			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want %d: the grant must not change what the route answers", rec.Code, http.StatusOK)
			}
		})
	}
}

// TestAC1AWildcardIsNeverSent holds the property AC1's refusal protects, at the
// handler rather than only at load: even an origin list that somehow carries
// `*` never puts `*` in a response.
func TestAC1AWildcardIsNeverSent(t *testing.T) {
	t.Parallel()

	for _, h := range []http.Header{
		{"Origin": {"*"}},
		preflightHeader("*", http.MethodGet, ""),
	} {
		method := http.MethodGet
		if h.Get("Access-Control-Request-Method") != "" {
			method = http.MethodOptions
		}

		rec := serveCORS(t, []string{"*"}, method, "/", h)

		if g := grants(rec); len(g) != 0 {
			t.Errorf("%s with Origin: * was answered with %q, want no grant", method, g)
		}
	}
}

// TestAC2TheGrantIsWhatTheClientUses pins the explicit lists to the client's
// calls. Its one call is a GET that sets no request header of its own (#26), so
// a preflight may be granted GET and no header. Widening either list is a
// deliberate change to this test, made beside the client call that needs it.
func TestAC2TheGrantIsWhatTheClientUses(t *testing.T) {
	t.Parallel()

	if want := []string{http.MethodGet}; !slices.Equal(corsMethods, want) {
		t.Errorf("corsMethods = %q, want %q: the methods the client's calls use, and no others", corsMethods, want)
	}

	if len(corsRequestHeaders) != 0 {
		t.Errorf("corsRequestHeaders = %q, want none: no client call sets a request header of its own", corsRequestHeaders)
	}
}

// TestAC2APreflightIsGrantedWhatTheClientUses binds #224 AC2's first half: a
// preflight from a configured origin asking for what the client's calls use is
// granted that origin, that method and those headers, from the lists held in
// this package.
func TestAC2APreflightIsGrantedWhatTheClientUses(t *testing.T) {
	t.Parallel()

	for _, method := range corsMethods {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			rec := serveCORS(t, []string{clientOrigin}, http.MethodOptions, "/",
				preflightHeader(clientOrigin, method, strings.Join(corsRequestHeaders, ", ")))

			if rec.Code != http.StatusNoContent {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
			}

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != clientOrigin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, clientOrigin)
			}

			allowed := strings.Split(rec.Header().Get("Access-Control-Allow-Methods"), ", ")
			if !slices.Contains(allowed, method) {
				t.Errorf("Access-Control-Allow-Methods = %q, want it to grant %s", allowed, method)
			}

			for _, name := range corsRequestHeaders {
				if !strings.Contains(strings.ToLower(rec.Header().Get("Access-Control-Allow-Headers")), strings.ToLower(name)) {
					t.Errorf("Access-Control-Allow-Headers = %q, want it to grant %s", rec.Header().Get("Access-Control-Allow-Headers"), name)
				}
			}

			for _, name := range []string{"Origin", "Access-Control-Request-Method", "Access-Control-Request-Headers"} {
				if !varies(rec, name) {
					t.Errorf("Vary = %q, want it to name %s", rec.Header().Values("Vary"), name)
				}
			}
		})
	}
}

// TestAC2NothingIsGrantedToAnyOtherOrigin binds #224 AC2's second half: a
// preflight or an actual request from an origin that is not configured, or
// under a configuration that names none, carries no Access-Control-Allow-*
// header at all — and neither does a preflight from a configured origin asking
// for more than the lists hold. The near misses are the ones a reflected or
// prefix-matched origin would grant.
func TestAC2NothingIsGrantedToAnyOtherOrigin(t *testing.T) {
	t.Parallel()

	configured := []string{clientOrigin}

	cases := []struct {
		name    string
		origins []string
		method  string
		header  http.Header
	}{
		{"another origin, actual", configured, http.MethodGet, http.Header{"Origin": {"https://attacker.example"}}},
		{"another origin, preflight", configured, http.MethodOptions, preflightHeader("https://attacker.example", http.MethodGet, "")},
		{"a suffix of the configured origin", configured, http.MethodGet, http.Header{"Origin": {clientOrigin + ".attacker.example"}}},
		{"the configured host over http", configured, http.MethodGet, http.Header{"Origin": {"http://turfgps.example"}}},
		{"the configured host on another port", configured, http.MethodGet, http.Header{"Origin": {clientOrigin + ":8443"}}},
		{"the configured origin, uppercased", configured, http.MethodGet, http.Header{"Origin": {strings.ToUpper(clientOrigin)}}},
		{"the null origin", configured, http.MethodOptions, preflightHeader("null", http.MethodGet, "")},
		{"two Origin values", configured, http.MethodGet, http.Header{"Origin": {"https://attacker.example", clientOrigin}}},
		{"no origin configured, actual", nil, http.MethodGet, http.Header{"Origin": {clientOrigin}}},
		{"no origin configured, preflight", nil, http.MethodOptions, preflightHeader(clientOrigin, http.MethodGet, "")},
		{"configured origin, a method the client does not use", configured, http.MethodOptions, preflightHeader(clientOrigin, http.MethodDelete, "")},
		{"configured origin, a header the client does not send", configured, http.MethodOptions, preflightHeader(clientOrigin, http.MethodGet, "x-plan-code")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := serveCORS(t, tc.origins, tc.method, "/", tc.header)

			if g := grants(rec); len(g) != 0 {
				t.Errorf("the response carries %q, want no Access-Control-Allow-* header", g)
			}

			if tc.method == http.MethodOptions && rec.Code < 400 {
				t.Errorf("a refused preflight was answered %d, want a refusal status", rec.Code)
			}
		})
	}
}

// syncBuffer is a log sink the server's goroutines and the test may share.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

// TestAC3NoPlanCodeReachesTheLogThroughCORS binds #224 AC3: a preflight, a
// granted request and a refused request, each carrying a plan code in its
// request target, leave no log line carrying that code.
//
// It captures both of the service's log paths at their most verbose — the
// default slog logger at Debug, and the standard log package, which the
// server's own error log writes through — and serves through a real
// http.Server rather than a recorder, so the server's logging is inside what
// is measured. It then writes a canary through both paths and requires it, so
// a capture that silently caught nothing fails rather than passes.
//
// It is not parallel because it swaps those two process-wide loggers, and
// restores them before it returns.
func TestAC3NoPlanCodeReachesTheLogThroughCORS(t *testing.T) {
	const planCode = "PLANCODE7QX4K9"

	sink := &syncBuffer{}

	prevSlog, prevOut, prevFlags := slog.Default(), log.Writer(), log.Flags()

	t.Cleanup(func() {
		slog.SetDefault(prevSlog)
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	})

	slog.SetDefault(slog.New(slog.NewTextHandler(sink, &slog.HandlerOptions{Level: slog.LevelDebug})))

	srv := httptest.NewServer(NewHandler([]string{clientOrigin}))
	defer srv.Close()

	target := srv.URL + "/plans/" + planCode + "?code=" + planCode

	for _, tc := range []struct {
		name   string
		method string
		header http.Header
	}{
		{"preflight, granted", http.MethodOptions, preflightHeader(clientOrigin, http.MethodGet, "")},
		{"preflight, refused", http.MethodOptions, preflightHeader("https://attacker.example", http.MethodGet, "")},
		{"actual, granted", http.MethodGet, http.Header{"Origin": {clientOrigin}}},
		{"actual, refused", http.MethodGet, http.Header{"Origin": {"https://attacker.example"}}},
	} {
		req, err := http.NewRequestWithContext(t.Context(), tc.method, target, nil)
		if err != nil {
			t.Fatalf("%s: building the request: %v", tc.name, err)
		}

		req.Header = tc.header

		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}

		_ = resp.Body.Close()
	}

	srv.Close()

	if strings.Contains(sink.String(), planCode) {
		t.Errorf("the log carries the plan code %s:\n%s", planCode, sink.String())
	}

	slog.Debug("capture canary slog")
	log.Print("capture canary log")

	for _, canary := range []string{"capture canary slog", "capture canary log"} {
		if !strings.Contains(sink.String(), canary) {
			t.Fatalf("the capture did not record %q, so it would not have recorded a plan code either and the check above measured nothing", canary)
		}
	}
}
