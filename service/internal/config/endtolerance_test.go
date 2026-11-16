package config

import (
	"strings"
	"testing"
)

// TestTheEndToleranceHasNoDefault (#250): with nothing set there is no
// tolerance, so no leg is served — never a value from here.
func TestTheEndToleranceHasNoDefault(t *testing.T) {
	t.Parallel()

	got, err := LoadEndTolerance(lookupMap(nil))
	if err != nil || got != nil {
		t.Fatalf("LoadEndTolerance with nothing set gave %+v, %v; want no configuration and no error", got, err)
	}
}

// TestTheEndToleranceArrivesWithItsOrigin (#250): a value is taken unchanged
// only with its origin, zero included as the strictest, and a partial or
// malformed one is refused rather than read as another number.
func TestTheEndToleranceArrivesWithItsOrigin(t *testing.T) {
	t.Parallel()

	const origin = "measured against the extract, the deployment's record"

	for raw, want := range map[string]float64{" 25.5 ": 25.5, "0": 0, "32053": 32053} {
		got, err := LoadEndTolerance(lookupMap(map[string]string{
			EnvEndTolerance:       raw,
			EnvEndToleranceOrigin: origin,
		}))
		if err != nil || got == nil || got.Metres != want || got.Origin != origin {
			t.Errorf("LoadEndTolerance(%q) gave %+v, %v; want %v m with its origin", raw, got, err, want)
		}
	}

	for name, env := range map[string]map[string]string{
		"a value without an origin": {EnvEndTolerance: "25"},
		"a blank origin":            {EnvEndTolerance: "25", EnvEndToleranceOrigin: " "},
		"an origin without a value": {EnvEndToleranceOrigin: origin},
		"a negative value":          {EnvEndTolerance: "-1", EnvEndToleranceOrigin: origin},
		"not a number":              {EnvEndTolerance: "NaN", EnvEndToleranceOrigin: origin},
		"an infinite value":         {EnvEndTolerance: "+Inf", EnvEndToleranceOrigin: origin},
		"a word":                    {EnvEndTolerance: "near", EnvEndToleranceOrigin: origin},
		"a unit":                    {EnvEndTolerance: "25m", EnvEndToleranceOrigin: origin},
	} {
		got, err := LoadEndTolerance(lookupMap(env))
		if err == nil || got != nil {
			t.Errorf("%s: LoadEndTolerance gave %+v, %v; want a refusal", name, got, err)

			continue
		}

		if !strings.Contains(err.Error(), EnvEndTolerance) {
			t.Errorf("%s: the refusal %q names no variable", name, err)
		}
	}
}
