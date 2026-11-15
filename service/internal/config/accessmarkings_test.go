package config

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

// metresText is v as the environment would carry it, in the fewest digits
// that read back as v exactly.
func metresText(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// TestABarrierToleranceBelowTheFloorIsRefused is #265 AC1: a tolerance below
// the floor, zero included, is refused even with its origin, and the refusal
// names the floor, its value and where it is derived.
func TestABarrierToleranceBelowTheFloorIsRefused(t *testing.T) {
	t.Parallel()

	floor := BarrierToleranceFloorMetres

	for _, raw := range []string{
		"0", "0.0", "-0",
		metresText(floor / 2),
		metresText(math.Nextafter(floor, 0)),
	} {
		got, err := LoadAccessMarkings(lookupMap(map[string]string{
			EnvAccessBarrierTolerance:       raw,
			EnvAccessBarrierToleranceOrigin: "measured against the extract, #265's test",
		}))
		if err == nil || got != nil {
			t.Errorf("a tolerance of %s m with its origin: LoadAccessMarkings gave %+v, %v; want a refusal below the floor of %v m", raw, got, err, floor)

			continue
		}

		for _, want := range []string{EnvAccessBarrierTolerance, "config.BarrierToleranceFloorMetres", metresText(floor), "accessmarkings.go"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("a tolerance of %s m: the refusal %q does not name %q", raw, err, want)
			}
		}
	}
}

// TestABarrierToleranceAtOrAboveTheFloorIsTakenUnchanged is #265 AC2: the floor
// itself, the next value above it, and the figure CI configures are each taken
// as they were written, with their origin.
func TestABarrierToleranceAtOrAboveTheFloorIsTakenUnchanged(t *testing.T) {
	t.Parallel()

	const origin = "measured against the extract, #265's test"

	floor := BarrierToleranceFloorMetres

	for _, raw := range []string{
		metresText(floor),
		metresText(math.Nextafter(floor, math.Inf(1))),
		"1.0",
	} {
		want, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			t.Fatal(err)
		}

		got, err := LoadAccessMarkings(lookupMap(map[string]string{
			EnvAccessBarrierTolerance:       raw,
			EnvAccessBarrierToleranceOrigin: origin,
		}))
		if err != nil || got == nil || got.BarrierToleranceMetres != want || got.BarrierToleranceOrigin != origin {
			t.Errorf("a tolerance of %s m: LoadAccessMarkings gave %+v, %v; want %v m with its origin, unchanged", raw, got, err, want)
		}
	}
}

// TestTheFloorCoversTheShapeEncodingAtEveryLatitude (#265 AC3): the floor is at
// least the farthest the shape encoding's rounding alone can move a point, at
// every latitude, measured with the spheroid's own radii rather than the bound
// the derivation uses for them. It checks the bound against one of its terms;
// the derivation is on the constant.
func TestTheFloorCoversTheShapeEncodingAtEveryLatitude(t *testing.T) {
	t.Parallel()

	const a = wgs84SemiMajorAxisMetres

	f := 1 / wgs84InverseFlattening
	e2 := f * (2 - f)
	half := legShapeQuantumDegrees / 2 * math.Pi / 180

	for q := -360; q <= 360; q++ {
		phi := float64(q) / 4 * math.Pi / 180
		w := 1 - e2*math.Sin(phi)*math.Sin(phi)
		meridian, primeVertical := a*(1-e2)/math.Pow(w, 1.5), a/math.Sqrt(w)

		if d := math.Hypot(meridian*half, primeVertical*math.Cos(phi)*half); d > BarrierToleranceFloorMetres {
			t.Errorf("at %v° the encoding's rounding moves a point %v m, beyond the floor of %v m", float64(q)/4, d, BarrierToleranceFloorMetres)
		}
	}
}

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
