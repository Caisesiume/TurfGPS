package config

import (
	"strings"
	"testing"
	"time"
)

// TestNoRoutingEngineConfiguredIsNotAnError: with neither variable set the
// service runs without journey planning, which is a configuration and not a
// fault.
func TestNoRoutingEngineConfiguredIsNotAnError(t *testing.T) {
	t.Parallel()

	for name, vars := range map[string]map[string]string{
		"nothing":       {},
		"an empty URL":  {EnvValhallaURL: ""},
		"only the sync": complete(),
	} {
		cfg, err := LoadValhalla(env(vars))
		if err != nil || cfg != nil {
			t.Errorf("with %s configured, LoadValhalla gave %+v, %v; want no engine and no error", name, cfg, err)
		}
	}
}

// TestTheRoutingEngineURLIsCarriedAsGiven: a well-formed base URL reaches the
// configuration unchanged, with the default timeout unless one is set.
func TestTheRoutingEngineURLIsCarriedAsGiven(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		"http://127.0.0.1:8002",
		"http://127.0.0.1:8002/",
		"http://[::1]:8002",
		"https://routing.example/valhalla",
	} {
		cfg, err := LoadValhalla(env(map[string]string{EnvValhallaURL: raw}))
		if err != nil {
			t.Errorf("%s=%q was refused: %v", EnvValhallaURL, raw, err)

			continue
		}

		if want := (Valhalla{URL: raw, Timeout: defaultValhallaTimeout}); *cfg != want {
			t.Errorf("%s=%q gave %+v, want %+v", EnvValhallaURL, raw, *cfg, want)
		}
	}

	cfg, err := LoadValhalla(env(map[string]string{EnvValhallaURL: "http://127.0.0.1:8002", EnvValhallaTimeout: "3s"}))
	if err != nil {
		t.Fatalf("a timeout of 3s was refused: %v", err)
	}

	if cfg.Timeout != 3*time.Second {
		t.Errorf("%s=3s gave a timeout of %v", EnvValhallaTimeout, cfg.Timeout)
	}
}

// TestAMalformedRoutingEngineConfigurationIsRefused: each malformed value is
// refused with an error naming the variable it came from, so the service
// stops at start rather than answering every journey with an error.
func TestAMalformedRoutingEngineConfigurationIsRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		vars  map[string]string
		names string
	}{
		{"no scheme", map[string]string{EnvValhallaURL: "127.0.0.1:8002"}, EnvValhallaURL},
		{"another scheme", map[string]string{EnvValhallaURL: "tcp://127.0.0.1:8002"}, EnvValhallaURL},
		{"no host", map[string]string{EnvValhallaURL: "http:///route"}, EnvValhallaURL},
		{"a query", map[string]string{EnvValhallaURL: "http://127.0.0.1:8002/?json={}"}, EnvValhallaURL},
		{"an empty query", map[string]string{EnvValhallaURL: "http://127.0.0.1:8002/?"}, EnvValhallaURL},
		{"a fragment", map[string]string{EnvValhallaURL: "http://127.0.0.1:8002/#route"}, EnvValhallaURL},
		{"not a URL", map[string]string{EnvValhallaURL: "http://127.0.0.1:%zz"}, EnvValhallaURL},
		{"a timeout that is not a duration", map[string]string{EnvValhallaURL: "http://127.0.0.1:8002", EnvValhallaTimeout: "10"}, EnvValhallaTimeout},
		{"a timeout that is not positive", map[string]string{EnvValhallaURL: "http://127.0.0.1:8002", EnvValhallaTimeout: "0s"}, EnvValhallaTimeout},
		{"a timeout without a URL", map[string]string{EnvValhallaTimeout: "3s"}, EnvValhallaURL},
	} {
		cfg, err := LoadValhalla(env(tc.vars))
		if err == nil {
			t.Errorf("%s: LoadValhalla accepted %v as %+v, want a refusal", tc.name, tc.vars, cfg)

			continue
		}

		if cfg != nil {
			t.Errorf("%s: LoadValhalla returned %+v beside its refusal, want none", tc.name, cfg)
		}

		if !strings.Contains(err.Error(), tc.names) {
			t.Errorf("%s: the refusal %q does not name %s", tc.name, err, tc.names)
		}
	}
}

// TestTheRefusalDoesNotEchoCredentials: main logs a refusal, so a URL carrying
// a password is refused without the password in the error, whichever check
// refuses it.
func TestTheRefusalDoesNotEchoCredentials(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		"ftp://user:hunter2@127.0.0.1:8002",
		"http://user:hunter2@127.0.0.1:%zz",
		"http://user:hunter2@127.0.0.1:8002/?q",
	} {
		_, err := LoadValhalla(env(map[string]string{EnvValhallaURL: raw}))
		if err == nil {
			t.Errorf("%q was accepted, want a refusal", raw)

			continue
		}

		if strings.Contains(err.Error(), "hunter2") {
			t.Errorf("the refusal %q quotes the URL's password", err)
		}
	}
}
