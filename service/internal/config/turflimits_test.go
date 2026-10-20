package config

import (
	"testing"
	"time"
)

// TestTheLimitsAreCarriedFromTheEnvironmentAndNowhereElse binds #28's note
// that no figure from either limit may appear in a configuration default: each
// is read from its variable, and an empty variable is refused rather than
// filled in.
func TestTheLimitsAreCarriedFromTheEnvironmentAndNowhereElse(t *testing.T) {
	t.Parallel()

	pairs := complete()
	pairs[EnvRequestInterval] = "1500ms"
	pairs[EnvAllZonesInterval] = "150s"

	cfg, err := LoadZoneSync(env(pairs))
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	if want := 1500 * time.Millisecond; cfg.RequestInterval != want {
		t.Errorf("the request interval loaded as %s, want %s", cfg.RequestInterval, want)
	}

	if want := 150 * time.Second; cfg.AllZonesInterval != want {
		t.Errorf("the all-zones interval loaded as %s, want %s", cfg.AllZonesInterval, want)
	}

	for _, name := range []string{EnvRequestInterval, EnvAllZonesInterval} {
		pairs := complete()
		pairs[name] = ""

		if cfg, err := LoadZoneSync(env(pairs)); err == nil {
			t.Errorf("an empty %s loaded as %+v, want it refused: a default here would be a second home for a figure the API publishes", name, cfg)
		}
	}
}

// TestTheRetryIntervalIsTheAllZonesIntervalPlusTheFetchDeadline is AC2's
// first guard under the monthly schedule (#233): the spacing between two
// attempts is derived, never read, and never shorter than the all-zones
// interval plus the fetch deadline. Across a restart only the scheduler's gate
// holds, so a spacing of the interval alone asks no more often than permitted
// and still breaks it there — the case TestTheAllZonesIntervalHoldsAcrossARestart
// in `internal/turf` runs at ShortestInterval.
func TestTheRetryIntervalIsTheAllZonesIntervalPlusTheFetchDeadline(t *testing.T) {
	t.Parallel()

	pairs := complete()
	pairs[EnvAllZonesInterval] = "7m"
	pairs[EnvFetchTimeout] = "41s"

	cfg, err := LoadZoneSync(env(pairs))
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	if want := 7*time.Minute + 41*time.Second; cfg.RetryInterval != want {
		t.Errorf("under an all-zones interval of 7m and a fetch deadline of 41s the retry interval loaded as %s, want %s: their sum, which ShortestInterval says why", cfg.RetryInterval, want)
	}

	// The restart test runs the sync at ShortestInterval, so it must be the
	// spacing the loader derives and not merely one like it.
	if shortest := ShortestInterval(cfg.AllZonesInterval, cfg.FetchTimeout); cfg.RetryInterval != shortest {
		t.Errorf("the loader derives %s and ShortestInterval says %s, want them equal", cfg.RetryInterval, shortest)
	}
}
