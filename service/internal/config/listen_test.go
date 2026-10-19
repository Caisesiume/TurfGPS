package config

import (
	"strings"
	"testing"
)

func lookupListenAddr(value string, set bool) LookupFunc {
	return func(name string) (string, bool) {
		if name == EnvListenAddr && set {
			return value, true
		}

		return "", false
	}
}

// TestListenAddrFallsBackWhenUnset binds the half of #125 that is this
// package's: with the variable unset or empty, the address is the fallback the
// caller passed, and nothing in this package supplies one of its own.
func TestListenAddrFallsBackWhenUnset(t *testing.T) {
	t.Parallel()

	const fallback = "127.0.0.1:8080"

	for _, tc := range []struct {
		name string
		set  bool
	}{
		{"unset", false},
		{"empty", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := LoadListenAddr(lookupListenAddr("", tc.set), fallback)
			if err != nil {
				t.Fatalf("LoadListenAddr: %v, want no error", err)
			}

			if got != fallback {
				t.Errorf("LoadListenAddr = %q, want the fallback %q", got, fallback)
			}
		})
	}
}

// TestListenAddrTakesAnExplicitAddress binds the override #125 names: an
// operator who sets the variable gets exactly the address they wrote, the
// every-interface forms included, because widening is theirs to choose.
func TestListenAddrTakesAnExplicitAddress(t *testing.T) {
	t.Parallel()

	for _, value := range []string{
		"127.0.0.1:8080",
		"[::1]:8080",
		"127.0.0.1:0",
		":8080",
		"0.0.0.0:8080",
		"[::]:8080",
		"10.0.0.5:9000",
	} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			got, err := LoadListenAddr(lookupListenAddr(value, true), "127.0.0.1:8080")
			if err != nil {
				t.Fatalf("LoadListenAddr(%q): %v, want it accepted", value, err)
			}

			if got != value {
				t.Errorf("LoadListenAddr(%q) = %q, want the value as written", value, got)
			}
		})
	}
}

// TestListenAddrRefusesAMalformedValue binds #125's refusal: a value the
// listener could not bind as written, or would bind somewhere the setting does
// not say, is refused with the variable named, and never replaced by the
// fallback.
func TestListenAddrRefusesAMalformedValue(t *testing.T) {
	t.Parallel()

	for _, value := range []string{
		"8080",
		"127.0.0.1",
		"127.0.0.1:",
		"::1:8080",
		"localhost:8080",
		"turfgps.example:8080",
		"127.0.0.1:http",
		"127.0.0.1:65536",
		"127.0.0.1:-1",
		" 127.0.0.1:8080",
		"127.0.0.1:8080 ",
		"http://127.0.0.1:8080",
	} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			got, err := LoadListenAddr(lookupListenAddr(value, true), "127.0.0.1:8080")
			if err == nil {
				t.Fatalf("LoadListenAddr(%q) = %q with no error, want a refusal", value, got)
			}

			if !strings.Contains(err.Error(), EnvListenAddr) {
				t.Errorf("the refusal %q does not name %s", err, EnvListenAddr)
			}

			if got != "" {
				t.Errorf("LoadListenAddr(%q) returned %q beside its refusal, want none", value, got)
			}
		})
	}
}
