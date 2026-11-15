package config

import (
	"strings"
	"testing"
)

// TestTheAvoidanceBoundHasNoDefault (#248): with nothing set there is no bound,
// so no leg is asked for again — never a value from here.
func TestTheAvoidanceBoundHasNoDefault(t *testing.T) {
	t.Parallel()

	got, err := LoadAccessAvoidance(lookupMap(nil))
	if err != nil || got != nil {
		t.Fatalf("LoadAccessAvoidance with nothing set gave %+v, %v; want no configuration and no error", got, err)
	}
}

// TestTheAvoidanceBoundArrivesWithItsOrigin: a value is taken only with its
// origin, and a malformed one is refused rather than read as another number.
func TestTheAvoidanceBoundArrivesWithItsOrigin(t *testing.T) {
	t.Parallel()

	const origin = "the deployment's choice, #249"

	got, err := LoadAccessAvoidance(lookupMap(map[string]string{
		EnvAccessAvoidanceAttempts:       " 3 ",
		EnvAccessAvoidanceAttemptsOrigin: origin,
	}))
	if err != nil || got == nil || got.Attempts != 3 || got.Origin != origin {
		t.Fatalf("LoadAccessAvoidance gave %+v, %v; want 3 attempts with their origin", got, err)
	}

	for name, env := range map[string]map[string]string{
		"a value without an origin": {EnvAccessAvoidanceAttempts: "3"},
		"a blank origin":            {EnvAccessAvoidanceAttempts: "3", EnvAccessAvoidanceAttemptsOrigin: " "},
		"an origin without a value": {EnvAccessAvoidanceAttemptsOrigin: origin},
		"a negative value":          {EnvAccessAvoidanceAttempts: "-1", EnvAccessAvoidanceAttemptsOrigin: origin},
		"a fraction":                {EnvAccessAvoidanceAttempts: "2.5", EnvAccessAvoidanceAttemptsOrigin: origin},
		"a word":                    {EnvAccessAvoidanceAttempts: "three", EnvAccessAvoidanceAttemptsOrigin: origin},
	} {
		got, err := LoadAccessAvoidance(lookupMap(env))
		if err == nil || got != nil {
			t.Errorf("%s: LoadAccessAvoidance gave %+v, %v; want a refusal", name, got, err)

			continue
		}

		if !strings.Contains(err.Error(), "TURFGPS_ACCESS_AVOIDANCE_ATTEMPTS") {
			t.Errorf("%s: the refusal %q names no variable", name, err)
		}
	}
}
