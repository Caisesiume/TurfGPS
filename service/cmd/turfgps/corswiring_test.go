package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
)

// TestStartupRefusesAWildcardOrigin binds #224 AC1's refusal at the point main
// acts on it: the composition root builds no request surface from a `*`
// origin, and its error names the setting, which main logs before exiting —
// before anything listens.
func TestStartupRefusesAWildcardOrigin(t *testing.T) {
	t.Parallel()

	handler, err := newHandler(func(name string) (string, bool) {
		if name == config.EnvCORSAllowedOrigins {
			return "*", true
		}

		return "", false
	})
	if err == nil {
		t.Fatal("newHandler accepted a * origin, want a refusal")
	}

	if !strings.Contains(err.Error(), config.EnvCORSAllowedOrigins) {
		t.Errorf("the refusal %q does not name %s", err, config.EnvCORSAllowedOrigins)
	}

	if handler != nil {
		t.Error("newHandler returned a handler beside its refusal, want none")
	}
}

// TestStartupGrantsTheConfiguredOrigin binds the wiring AC1 depends on: the
// origin the environment names is the origin the served handler grants.
func TestStartupGrantsTheConfiguredOrigin(t *testing.T) {
	t.Parallel()

	const origin = "https://turfgps.example"

	handler, err := newHandler(func(name string) (string, bool) {
		if name == config.EnvCORSAllowedOrigins {
			return origin, true
		}

		return "", false
	})
	if err != nil {
		t.Fatalf("newHandler: %v", err)
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.Header.Set("Origin", origin)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q: the configured origin did not reach the served handler", got, origin)
	}
}
