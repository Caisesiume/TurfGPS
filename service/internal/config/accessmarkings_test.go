package config

import (
	"strings"
	"testing"
)

// TestTheBarrierToleranceHasNoDefault (#255): with nothing set there is no
// tolerance, and so no check that could pass a leg — never a value from here.
func TestTheBarrierToleranceHasNoDefault(t *testing.T) {
	t.Parallel()

	got, err := LoadAccessMarkings(lookupMap(nil))
	if err != nil || got != nil {
		t.Fatalf("LoadAccessMarkings with nothing set gave %+v, %v; want no configuration and no error", got, err)
	}
}

// TestTheBarrierToleranceArrivesWithItsOrigin: a value is taken only with the
// origin that accounts for it, and a malformed one is refused rather than read
// as some other number.
func TestTheBarrierToleranceArrivesWithItsOrigin(t *testing.T) {
	t.Parallel()

	const origin = "measured against the extract on 2026-11-14, #255"

	got, err := LoadAccessMarkings(lookupMap(map[string]string{
		EnvAccessBarrierTolerance:       " 2.5 ",
		EnvAccessBarrierToleranceOrigin: origin,
	}))
	if err != nil || got == nil || got.BarrierToleranceMetres != 2.5 || got.BarrierToleranceOrigin != origin {
		t.Fatalf("LoadAccessMarkings gave %+v, %v; want 2.5 m with its origin", got, err)
	}

	for name, env := range map[string]map[string]string{
		"a value without an origin": {EnvAccessBarrierTolerance: "2.5"},
		"a blank origin":            {EnvAccessBarrierTolerance: "2.5", EnvAccessBarrierToleranceOrigin: "  "},
		"an origin without a value": {EnvAccessBarrierToleranceOrigin: origin},
		"a negative value":          {EnvAccessBarrierTolerance: "-1", EnvAccessBarrierToleranceOrigin: origin},
		"NaN":                       {EnvAccessBarrierTolerance: "NaN", EnvAccessBarrierToleranceOrigin: origin},
		"infinity":                  {EnvAccessBarrierTolerance: "+Inf", EnvAccessBarrierToleranceOrigin: origin},
		"a unit":                    {EnvAccessBarrierTolerance: "2.5m", EnvAccessBarrierToleranceOrigin: origin},
	} {
		got, err := LoadAccessMarkings(lookupMap(env))
		if err == nil || got != nil {
			t.Errorf("%s: LoadAccessMarkings gave %+v, %v; want a refusal", name, got, err)

			continue
		}

		if !strings.Contains(err.Error(), "TURFGPS_ACCESS_BARRIER_TOLERANCE") {
			t.Errorf("%s: the refusal %q names no variable", name, err)
		}
	}
}

func lookupMap(env map[string]string) LookupFunc {
	return func(name string) (string, bool) {
		v, ok := env[name]

		return v, ok
	}
}
