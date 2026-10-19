package config

import (
	"slices"
	"strings"
	"testing"
)

func lookupOrigins(value string, set bool) LookupFunc {
	return func(name string) (string, bool) {
		if name == EnvCORSAllowedOrigins && set {
			return value, true
		}

		return "", false
	}
}

// TestNoOriginIsConfiguredByDefault binds the half of #224 AC2 that is about
// configuration: unset or empty, the variable names no origin, and nothing in
// this package supplies one in its place.
func TestNoOriginIsConfiguredByDefault(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		value string
		set   bool
	}{
		{"unset", "", false},
		{"empty", "", true},
		{"blank", "  ", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := LoadAllowedOrigins(lookupOrigins(tc.value, tc.set))
			if err != nil {
				t.Fatalf("LoadAllowedOrigins: %v, want no error", err)
			}

			if len(got) != 0 {
				t.Errorf("LoadAllowedOrigins = %q, want no origin: none was configured, and none may be defaulted", got)
			}
		})
	}
}

// TestConfiguredOriginsAreReadExactly binds #224 AC1's "read from
// configuration": every origin listed is returned, in the form it was written,
// which is the form a browser sends.
func TestConfiguredOriginsAreReadExactly(t *testing.T) {
	t.Parallel()

	got, err := LoadAllowedOrigins(lookupOrigins("https://turfgps.example, https://preview.turfgps.example:8443,https://[2001:db8::1]", true))
	if err != nil {
		t.Fatalf("LoadAllowedOrigins: %v, want no error", err)
	}

	want := []string{"https://turfgps.example", "https://preview.turfgps.example:8443", "https://[2001:db8::1]"}
	if !slices.Equal(got, want) {
		t.Errorf("LoadAllowedOrigins = %q, want %q", got, want)
	}
}

// TestAMalformedOriginIsRefusedNamingTheSetting binds #224 AC1's refusal: `*`
// is refused at load with an error naming the variable, and so is every other
// value that is not one exact https origin — a wildcard pattern, a scheme other
// than https, and each spelling a browser never sends, which would otherwise
// load cleanly and then never match.
func TestAMalformedOriginIsRefusedNamingTheSetting(t *testing.T) {
	t.Parallel()

	for _, value := range []string{
		"*",
		"https://turfgps.example,*",
		"https://*.vercel.app",
		"null",
		"http://turfgps.example",
		"turfgps.example",
		"https://",
		"https://Turfgps.example",
		"https://turfgps.example/",
		"https://turfgps.example/plans",
		"https://turfgps.example:443",
		"http://localhost:80",
		"https://turfgps.example:",
		"https://turfgps.example?",
		"https://turfgps.example#",
		"https://user@turfgps.example",
		"https://turfgps.example,",
		",https://turfgps.example",
		"https://turfgps.example,,https://preview.turfgps.example",
	} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			got, err := LoadAllowedOrigins(lookupOrigins(value, true))
			if err == nil {
				t.Fatalf("LoadAllowedOrigins(%q) = %q with no error, want a refusal", value, got)
			}

			if !strings.Contains(err.Error(), EnvCORSAllowedOrigins) {
				t.Errorf("the refusal %q does not name %s, so an operator cannot tell which setting to fix", err, EnvCORSAllowedOrigins)
			}

			if got != nil {
				t.Errorf("LoadAllowedOrigins(%q) returned %q beside its error, want nothing", value, got)
			}
		})
	}
}

// TestAnHTTPOriginOffLoopbackIsRefused binds the other half of the loopback
// allowance (N2): http is accepted on localhost, 127.0.0.1 and [::1] alone, so
// an http origin on any other host — a public one, or one that only spells a
// loopback name inside it — is still refused at load, naming the setting.
func TestAnHTTPOriginOffLoopbackIsRefused(t *testing.T) {
	t.Parallel()

	for _, value := range []string{
		"http://turfgps.example",
		"http://preview.turfgps.example:5173",
		"http://203.0.113.7:5173",
		"http://[2001:db8::1]:5173",
		"http://localhost.attacker.example:5173",
		"http://127.0.0.1.attacker.example",
		"http://localhost@attacker.example",
		"http://localhost,http://attacker.example",
	} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			got, err := LoadAllowedOrigins(lookupOrigins(value, true))
			if err == nil {
				t.Fatalf("LoadAllowedOrigins(%q) = %q with no error, want a refusal: http is accepted on loopback only", value, got)
			}

			if !strings.Contains(err.Error(), EnvCORSAllowedOrigins) {
				t.Errorf("the refusal %q does not name %s", err, EnvCORSAllowedOrigins)
			}

			if got != nil {
				t.Errorf("LoadAllowedOrigins(%q) returned %q beside its error, want nothing", value, got)
			}
		})
	}
}
