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

// TestAScheduleShorterThanTheAllZonesIntervalIsRefused is AC2's first guard: a
// schedule that would ask the all-zones endpoint more often than it permits
// does not load. A schedule equal to the interval does, because it asks no more
// often than permitted.
func TestAScheduleShorterThanTheAllZonesIntervalIsRefused(t *testing.T) {
	t.Parallel()

	pairs := complete()
	pairs[EnvAllZonesInterval] = "7m"
	pairs[EnvInterval] = "6m59s"

	cfg, err := LoadZoneSync(env(pairs))
	if err == nil {
		t.Fatalf("a schedule of %s under an all-zones interval of %s loaded as %+v, want it refused", pairs[EnvInterval], pairs[EnvAllZonesInterval], cfg)
	}

	for _, name := range []string{EnvInterval, EnvAllZonesInterval} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("the refusal reads %q, want it to name %s", err, name)
		}
	}

	pairs[EnvInterval] = "7m"

	if _, err := LoadZoneSync(env(pairs)); err != nil {
		t.Errorf("a schedule equal to the all-zones interval was refused with %v, want it loaded", err)
	}
}
