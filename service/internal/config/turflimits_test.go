package config

import (
	"strings"
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

// TestAScheduleShorterThanTheIntervalPlusTheFetchDeadlineIsRefused is AC2's
// first guard: a schedule that could ask the all-zones endpoint sooner than it
// permits does not load. Across a restart only the schedule holds, so that
// includes a schedule equal to the interval, which asks no more often than
// permitted and still breaks it there — the case
// TestTheAllZonesIntervalHoldsAcrossARestart in `internal/turf` runs. The
// shortest schedule that loads is the interval plus the fetch deadline.
func TestAScheduleShorterThanTheIntervalPlusTheFetchDeadlineIsRefused(t *testing.T) {
	t.Parallel()

	pairs := complete()
	pairs[EnvAllZonesInterval] = "7m"
	pairs[EnvFetchTimeout] = "41s"

	for _, schedule := range []string{"7m", "7m40.999s"} {
		pairs[EnvInterval] = schedule

		cfg, err := LoadZoneSync(env(pairs))
		if err == nil {
			t.Errorf("a schedule of %s under an all-zones interval of %s and a fetch deadline of %s loaded as %+v, want it refused",
				schedule, pairs[EnvAllZonesInterval], pairs[EnvFetchTimeout], cfg)

			continue
		}

		for _, name := range []string{EnvInterval, EnvAllZonesInterval, EnvFetchTimeout} {
			if !strings.Contains(err.Error(), name) {
				t.Errorf("the refusal reads %q, want it to name %s", err, name)
			}
		}
	}

	pairs[EnvInterval] = "7m41s"

	cfg, err := LoadZoneSync(env(pairs))
	if err != nil {
		t.Fatalf("a schedule of the all-zones interval plus the fetch deadline was refused with %v, want it loaded", err)
	}

	// The restart test runs the sync at ShortestInterval, so it must be the
	// shortest schedule that loads and not merely one that does.
	if shortest := ShortestInterval(cfg.AllZonesInterval, cfg.FetchTimeout); cfg.Interval != shortest {
		t.Errorf("the shortest schedule that loads is %s and ShortestInterval says %s, want them equal", cfg.Interval, shortest)
	}
}
