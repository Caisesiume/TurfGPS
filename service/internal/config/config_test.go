package config

import (
	"strings"
	"testing"
	"time"
)

func env(pairs map[string]string) LookupFunc {
	return func(name string) (string, bool) {
		v, ok := pairs[name]

		return v, ok
	}
}

// complete is the smallest configuration that runs a sync.
//
// The interval here is a test's own figure and is deliberately unlike the one
// the endpoint permits, whose single home is `Architecture.md § Retrieving
// zones`. What is under test is that the value is carried through, not what the
// value is.
func complete() map[string]string {
	return map[string]string{
		EnvDatabaseURL: "postgres://localhost/turfgps",
		EnvAllZonesURL: "https://example.test/zones",
		EnvInterval:    "3m",
	}
}

// TestNoSyncConfiguredIsNotAnError binds the property `NFR-003` measures: the
// service starts and serves on a clean host with nothing beside it. A service
// that refused to start without a database and an endpoint would spend that
// property to gain a check the sync's own logging already makes.
func TestNoSyncConfiguredIsNotAnError(t *testing.T) {
	t.Parallel()

	cfg, err := LoadZoneSync(env(nil))
	if err != nil {
		t.Fatalf("an unconfigured sync returned %v, want no error", err)
	}

	if cfg != nil {
		t.Errorf("an unconfigured sync returned %+v, want nil", cfg)
	}
}

// TestPartialConfigurationIsRefused is the other half of that decision, and the
// reason it is not simply "absent means off". A half-configured sync that
// started quietly would be indistinguishable from one deliberately running
// without a sync, and the operator would learn about the typo from the copy
// never getting any fresher.
func TestPartialConfigurationIsRefused(t *testing.T) {
	t.Parallel()

	for _, missing := range []string{EnvDatabaseURL, EnvAllZonesURL, EnvInterval} {
		t.Run("without "+missing, func(t *testing.T) {
			t.Parallel()

			pairs := complete()
			delete(pairs, missing)

			cfg, err := LoadZoneSync(env(pairs))
			if err == nil {
				t.Fatalf("a configuration missing %s was accepted as %+v, want it refused", missing, cfg)
			}

			if !strings.Contains(err.Error(), missing) {
				t.Errorf("the refusal reads %q, want it to name %s", err, missing)
			}
		})
	}
}

// TestTheIntervalIsCarriedFromTheEnvironmentAndNowhereElse is the constraint
// this package exists for. The interval has one home and it is a document; this
// package must hold no default that could stand in for it.
func TestTheIntervalIsCarriedFromTheEnvironmentAndNowhereElse(t *testing.T) {
	t.Parallel()

	pairs := complete()
	pairs[EnvInterval] = "17m30s"

	cfg, err := LoadZoneSync(env(pairs))
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	if want := 17*time.Minute + 30*time.Second; cfg.Interval != want {
		t.Errorf("the interval loaded as %s, want %s", cfg.Interval, want)
	}

	// The same load with the variable empty must fail rather than fall back,
	// which is what says there is no compiled-in interval to fall back to.
	pairs[EnvInterval] = ""

	if cfg, err := LoadZoneSync(env(pairs)); err == nil {
		t.Errorf("an empty interval loaded as %+v, want it refused: a default here would be a second home for a figure that has one", cfg)
	}
}

// TestTheEndpointIsValidatedAtLoad catches a misconfiguration at start-up rather
// than once per interval for ever.
func TestTheEndpointIsValidatedAtLoad(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]string{
		"no scheme":      "example.test/zones",
		"a bare path":    "/zones",
		"a wrong scheme": "ftp://example.test/zones",

		// http is the dangerous row of this table rather than the obvious one:
		// unlike the three above it parses, resolves, and fetches. It was
		// accepted as readily as https, so the corpus that becomes the
		// authoritative `zone` geometry crossed the network in the clear
		// whenever the variable happened to be typed without the s — and every
		// assertion downstream of the wire accepts a rewritten body whose
		// coordinates are merely plausible.
		"plain http": "http://example.test/zones",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pairs := complete()
			pairs[EnvAllZonesURL] = value

			if cfg, err := LoadZoneSync(env(pairs)); err == nil {
				t.Errorf("the endpoint %q was accepted as %+v, want it refused", value, cfg)
			}
		})
	}
}

// TestTheTunablesHaveDefaultsAndAreOverridable separates the two kinds of value
// this package holds: the operational budgets are its own and carry defaults,
// while the interval and the endpoint are not and do not.
func TestTheTunablesHaveDefaultsAndAreOverridable(t *testing.T) {
	t.Parallel()

	cfg, err := LoadZoneSync(env(complete()))
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	if cfg.FetchTimeout <= 0 || cfg.DatabaseTimeout <= 0 || cfg.MergeTimeout <= 0 {
		t.Errorf("a budget defaulted to a non-positive value: %+v", cfg)
	}

	if cfg.MaxResponseBytes <= 0 {
		t.Errorf("the response ceiling defaulted to %d", cfg.MaxResponseBytes)
	}

	if cfg.MinZoneRatio <= 0 || cfg.MinZoneRatio > 1 {
		t.Errorf("the floor ratio defaulted to %v, outside (0, 1]", cfg.MinZoneRatio)
	}

	pairs := complete()
	pairs[EnvFetchTimeout] = "45s"
	pairs[EnvMinZoneRatio] = "0.5"
	pairs[EnvMaxResponseBytes] = "1024"

	cfg, err = LoadZoneSync(env(pairs))
	if err != nil {
		t.Fatalf("loading the overrides: %v", err)
	}

	if cfg.FetchTimeout != 45*time.Second || cfg.MinZoneRatio != 0.5 || cfg.MaxResponseBytes != 1024 {
		t.Errorf("the overrides loaded as %+v", cfg)
	}
}

// TestAMalformedTunableIsRefusedRatherThanDefaulted keeps a typo from reading as
// an intention. Falling back to the default here would run the sync on a budget
// the operator did not choose and did not know they had not chosen.
func TestAMalformedTunableIsRefusedRatherThanDefaulted(t *testing.T) {
	t.Parallel()

	for name, pair := range map[string][2]string{
		"a duration that is not one": {EnvFetchTimeout, "soon"},
		"a negative duration":        {EnvDatabaseTimeout, "-1s"},
		"a zero duration":            {EnvMergeTimeout, "0s"},
		"a ratio above one":          {EnvMinZoneRatio, "1.5"},
		"a ratio of zero":            {EnvMinZoneRatio, "0"},
		"a byte count that is not":   {EnvMaxResponseBytes, "lots"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pairs := complete()
			pairs[pair[0]] = pair[1]

			if cfg, err := LoadZoneSync(env(pairs)); err == nil {
				t.Errorf("%s=%q was accepted as %+v, want it refused", pair[0], pair[1], cfg)
			}
		})
	}
}
