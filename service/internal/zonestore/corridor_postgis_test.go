//go:build postgis

package zonestore

// `FR-023` AC1 against PostGIS: given a local copy within the staleness bound
// and a route corridor, the corridor's zones come back from the local copy and
// no Turf API request is issued.
//
// IT IS BUILT ONLY UNDER THE `postgis` TAG, AND THAT IS NOT A SKIP. The default
// suite does not compile this file, so it cannot print `ok` having measured
// nothing; CI's `migrations` job builds it on purpose, against a scratch database
// it has just applied `migrations/` to, and fails if it fails. Built under the tag
// with no database named, it fails rather than skips.
//
// IT WRITES ROWS, SO IT REFUSES ANY DATABASE BUT A SCRATCH ONE: a host that is not
// loopback, or a `zone` or `sync_run` that already holds rows, is refused before
// anything is written, and what it wrote is deleted after.
//
// The "no Turf API request" half is also held without a database, structurally,
// by `internal/corridor/noturf_test.go`. Here it is watched in the act: every
// request through the process's default HTTP transport — the one the Turf
// adapter's client uses — is recorded and refused while the corridor resolves.

import (
	"cmp"
	"context"
	"errors"
	"net/http"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Caisesiume/TurfGPS/service/internal/corridor"
)

// testDatabaseEnv names the scratch database. It is a test's input and never
// the service's configuration.
const testDatabaseEnv = "TURFGPS_TEST_DATABASE_URL"

// probeHalfWidth is a probe distance, as part E's are in
// `migrations/0001_zone_store.verify.sql`. The product's half-width is chosen
// nowhere yet (`Architecture.md § What is unproven`, item 7) and this settles
// nothing about it.
const probeHalfWidth = 1000.0

// The route runs east along latitude 59.33 in Stockholm. The zones below sit
// due north of its midpoint, at offsets in latitude whose length in metres is
// the known distance `postgis-migration-protocol § Standing schema facts` asks a
// coordinate test to assert: a degree of latitude here is about 111.4 km.
var route = []corridor.Point{{Latitude: 59.33, Longitude: 18.02}, {Latitude: 59.33, Longitude: 18.10}}

func typeID(v int16) *int16 { return &v }

var (
	onRoute = corridor.Zone{
		ID: 9_000_001, Name: "on the route", Latitude: 59.33, Longitude: 18.06,
		DateCreated:    time.Date(2015, 6, 1, 0, 0, 0, 0, time.UTC),
		TotalTakeovers: 7, TakeoverPoints: 185, PointsPerHour: 3, TypeID: typeID(2),
	}
	// 0.0045 degrees north: about 500 m, inside the probe half-width.
	nearRoute = corridor.Zone{
		ID: 9_000_002, Name: "500 m north", Latitude: 59.3345, Longitude: 18.06,
		DateCreated:    time.Date(2021, 3, 14, 12, 0, 0, 0, time.UTC),
		TotalTakeovers: 11, TakeoverPoints: 125, PointsPerHour: 1, TypeID: nil,
	}
	// 0.0135 degrees north: about 1,500 m, outside the probe half-width and
	// inside twice it.
	offRoute = corridor.Zone{
		ID: 9_000_003, Name: "1500 m north", Latitude: 59.3435, Longitude: 18.06,
		DateCreated:    time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		TotalTakeovers: 2, TakeoverPoints: 65, PointsPerHour: 0, TypeID: nil,
	}
	// On the line a route written latitude-first would draw: returned only if
	// the axes were swapped, in which case the two above would not be.
	onSwappedRoute = corridor.Zone{
		ID: 9_000_004, Name: "on the swapped route", Latitude: 18.06, Longitude: 59.33,
		DateCreated:    time.Date(2012, 8, 20, 0, 0, 0, 0, time.UTC),
		TotalTakeovers: 40, TakeoverPoints: 95, PointsPerHour: 2, TypeID: typeID(5),
	}
)

func TestACorridorResolvesFromTheLocalCopyWithoutCallingTurf(t *testing.T) {
	ctx := t.Context()
	pool := openScratch(t)

	reader, err := NewReader(pool)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	// The Given: a local copy within the staleness bound. A run that completed
	// just now is within any bound, so the bound is not restated here.
	if _, err := pool.Exec(ctx, `INSERT INTO sync_run (started_at, completed_at, outcome) VALUES (now(), now(), 'ok')`); err != nil {
		t.Fatalf("recording a completed run: %v", err)
	}

	for _, z := range []corridor.Zone{onRoute, nearRoute, offRoute, onSwappedRoute} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO zone (id, name, latitude, longitude, date_created, total_takeovers,
			                  takeover_points, points_per_hour, type_id, region_id, region_name,
			                  first_seen_at, last_changed_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 141, 'Stockholm', now(), now())`,
			z.ID, z.Name, z.Latitude, z.Longitude, z.DateCreated, z.TotalTakeovers,
			z.TakeoverPoints, z.PointsPerHour, z.TypeID); err != nil {
			t.Fatalf("loading zone %d: %v", z.ID, err)
		}
	}

	currency, err := reader.Currency(ctx)
	if err != nil || !currency.EverSucceeded {
		t.Fatalf("the copy reports currency %+v, %v; the Given needs a completed run", currency, err)
	}

	// The When, with every request through the default transport recorded and
	// refused.
	transport := &refusingTransport{}
	previous := http.DefaultTransport
	http.DefaultTransport = transport

	t.Cleanup(func() { http.DefaultTransport = previous })

	got, err := corridor.Resolve(ctx, reader, route, probeHalfWidth)

	http.DefaultTransport = previous

	if err != nil {
		t.Fatalf("resolving the corridor: %v", err)
	}

	if sent := transport.sent(); len(sent) != 0 {
		t.Errorf("resolving the corridor issued %d HTTP request(s) %v: FR-023 AC1 requires the corridor to resolve from the local copy with no Turf API request",
			len(sent), sent)
	}

	assertZones(t, "at the probe half-width", got, onRoute, nearRoute)

	// The same corridor twice as wide takes in the zone 1,500 m off it, so it
	// was the distance that kept that zone out above and nothing else.
	wider, err := corridor.Resolve(ctx, reader, route, 2*probeHalfWidth)
	if err != nil {
		t.Fatalf("resolving the wider corridor: %v", err)
	}

	assertZones(t, "at twice the probe half-width", wider, onRoute, nearRoute, offRoute)
}

// assertZones compares by every field, so a column scanned into the wrong field
// is caught as well as a zone returned or missed.
func assertZones(t *testing.T, label string, got []corridor.Zone, want ...corridor.Zone) {
	t.Helper()

	byID := func(a, b corridor.Zone) int { return cmp.Compare(a.ID, b.ID) }
	slices.SortFunc(got, byID)
	slices.SortFunc(want, byID)

	if !slices.EqualFunc(got, want, sameZone) {
		t.Errorf("%s the corridor resolved to\n  %+v\nwant\n  %+v", label, got, want)
	}
}

func sameZone(a, b corridor.Zone) bool {
	sameType := (a.TypeID == nil && b.TypeID == nil) ||
		(a.TypeID != nil && b.TypeID != nil && *a.TypeID == *b.TypeID)

	return sameType && a.ID == b.ID && a.Name == b.Name &&
		a.Latitude == b.Latitude && a.Longitude == b.Longitude &&
		a.DateCreated.Equal(b.DateCreated) && a.TotalTakeovers == b.TotalTakeovers &&
		a.TakeoverPoints == b.TakeoverPoints && a.PointsPerHour == b.PointsPerHour
}

// openScratch opens the database testDatabaseEnv names through Open, so the
// pool carries the settings the service's does, and refuses one that is not a
// scratch database.
func openScratch(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv(testDatabaseEnv)
	if url == "" {
		t.Fatalf("%s is not set: this file is built only under the postgis tag, for a scratch database with migrations/ applied, and it fails rather than skips without one", testDatabaseEnv)
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("%s does not parse: %v", testDatabaseEnv, err)
	}

	switch cfg.ConnConfig.Host {
	case "127.0.0.1", "localhost", "::1":
	default:
		t.Fatalf("%s names host %q, not loopback; this test writes rows and runs only against a throwaway local instance", testDatabaseEnv, cfg.ConnConfig.Host)
	}

	pool, err := Open(t.Context(), url)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	t.Cleanup(pool.Close)

	var zones, runs int64
	if err := pool.QueryRow(t.Context(), `SELECT (SELECT count(*) FROM zone), (SELECT count(*) FROM sync_run)`).Scan(&zones, &runs); err != nil {
		t.Fatalf("counting what the database already holds: %v", err)
	}

	if zones != 0 || runs != 0 {
		t.Fatalf("the database holds %d zone(s) and %d sync run(s); this test writes and deletes rows and refuses anything but an empty, freshly migrated scratch copy", zones, runs)
	}

	// Registered after pool.Close, so it runs before it.
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for _, stmt := range []string{`DELETE FROM zone`, `DELETE FROM sync_run`} {
			if _, err := pool.Exec(ctx, stmt); err != nil {
				t.Errorf("cleaning up (%s): %v", stmt, err)
			}
		}
	})

	return pool
}

// refusingTransport records every request it is handed and sends none.
type refusingTransport struct {
	mu       sync.Mutex
	requests []string
}

func (r *refusingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.requests = append(r.requests, req.Method+" "+req.URL.String())

	return nil, errors.New("this test refuses every HTTP request")
}

func (r *refusingTransport) sent() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return slices.Clone(r.requests)
}
