package corridor

import (
	"cmp"
	"context"
	"errors"
	"math"
	"slices"
	"testing"
)

// recordingRepository answers the port from memory and records what it was
// asked.
type recordingRepository struct {
	zones []Zone
	err   error

	calls     int
	route     []Point
	halfWidth float64
}

func (r *recordingRepository) ZonesInCorridor(_ context.Context, route []Point, halfWidthMetres float64) ([]Zone, error) {
	r.calls++
	r.route = route
	r.halfWidth = halfWidthMetres

	return r.zones, r.err
}

var stockholm = []Point{{Latitude: 59.33, Longitude: 18.02}, {Latitude: 59.33, Longitude: 18.10}}

func TestResolveAsksThePortForTheCorridorItWasGiven(t *testing.T) {
	t.Parallel()

	want := []Zone{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
	repo := &recordingRepository{zones: want}

	got, err := Resolve(t.Context(), repo, stockholm, 1000)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if repo.calls != 1 || !slices.Equal(repo.route, stockholm) || repo.halfWidth != 1000 {
		t.Errorf("the port was asked %d time(s) for route %v at %v m, want once for %v at 1000 m",
			repo.calls, repo.route, repo.halfWidth, stockholm)
	}

	if !slices.EqualFunc(got, want, func(a, b Zone) bool { return a.ID == b.ID && a.Name == b.Name }) {
		t.Errorf("Resolve returned %v, want the port's answer %v unchanged", got, want)
	}
}

func TestResolveRefusesACorridorNobodyAskedFor(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		route     []Point
		halfWidth float64
	}{
		"zero half-width":        {stockholm, 0},
		"negative half-width":    {stockholm, -1000},
		"NaN half-width":         {stockholm, math.NaN()},
		"infinite half-width":    {stockholm, math.Inf(1)},
		"no route":               {nil, 1000},
		"one-point route":        {stockholm[:1], 1000},
		"latitude out of range":  {[]Point{{Latitude: 91, Longitude: 18}, {Latitude: 59, Longitude: 18}}, 1000},
		"longitude out of range": {[]Point{{Latitude: 59, Longitude: 18}, {Latitude: 59, Longitude: -180.5}}, 1000},
		"NaN coordinate":         {[]Point{{Latitude: 59, Longitude: 18}, {Latitude: math.NaN(), Longitude: 18}}, 1000},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := &recordingRepository{}

			if _, err := Resolve(t.Context(), repo, tc.route, tc.halfWidth); err == nil {
				t.Errorf("Resolve accepted route %v at %v m, want a refusal", tc.route, tc.halfWidth)
			}

			if repo.calls != 0 {
				t.Errorf("the port was asked %d time(s) for a corridor that should have been refused first", repo.calls)
			}
		})
	}
}

func TestResolveReportsThePortsFailure(t *testing.T) {
	t.Parallel()

	cause := errors.New("store unreachable")

	_, err := Resolve(t.Context(), &recordingRepository{err: cause}, stockholm, 1000)
	if !errors.Is(err, cause) {
		t.Errorf("Resolve returned %v, want an error wrapping %v", err, cause)
	}
}

func TestResolveRefusesNoRepository(t *testing.T) {
	t.Parallel()

	if _, err := Resolve(t.Context(), nil, stockholm, 1000); err == nil {
		t.Error("Resolve accepted a nil repository, want a refusal")
	}
}

// TestResolveReturnsBothZonesOfAClosePair is `FR-032` (#30) AC2 at the stage
// that selects a corridor's zones: two zones closer together than the
// placement guideline of `Architecture.md § Distance between zones` come back
// as two, each at its own coordinate. The pairs carry no figure for the
// guideline, and need none: the first shares one coordinate and the second is
// one float64 step apart, which is closer than a floor of any size. What the
// corridor query itself returns for such a pair is `internal/zonestore`'s
// corridor_postgis_test.go.
func TestResolveReturnsBothZonesOfAClosePair(t *testing.T) {
	t.Parallel()

	const stepped = 59.3301

	want := []Zone{
		{ID: 7001, Latitude: 59.3293, Longitude: 18.0686},
		{ID: 7002, Latitude: 59.3293, Longitude: 18.0686},
		{ID: 7003, Latitude: stepped, Longitude: 18.0750},
		{ID: 7004, Latitude: math.Nextafter(stepped, 90), Longitude: 18.0750},
	}

	got, err := Resolve(t.Context(), &recordingRepository{zones: slices.Clone(want)}, stockholm, 1000)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	got = slices.Clone(got)
	slices.SortFunc(got, func(a, b Zone) int { return cmp.Compare(a.ID, b.ID) })

	if !slices.EqualFunc(got, want, func(a, b Zone) bool {
		return a.ID == b.ID && a.Latitude == b.Latitude && a.Longitude == b.Longitude
	}) {
		t.Errorf("Resolve returned %+v for a corridor holding two close pairs, want all four zones at their own coordinates %+v — FR-032 forbids assuming any minimum distance between two zones",
			got, want)
	}
}
