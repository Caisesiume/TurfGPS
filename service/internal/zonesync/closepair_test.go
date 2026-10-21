package zonesync

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
)

// `FR-032` (#30) AC1, the ingest half: a zone set holding two zones closer
// together than the placement guideline of `Architecture.md § Distance between
// zones` reaches the store as that many zones, each at its own coordinate —
// by the all-zones run and by the zone feed. What PostgreSQL then stores, and
// what the corridor query returns, is `internal/syncstore`'s postgis_test.go
// and `internal/zonestore`'s corridor_postgis_test.go.
//
// THE PAIRS CARRY NO FIGURE FOR THE GUIDELINE, AND NEED NONE. #30 forbids a
// separation assumed at any size, so the fixture is closer than every size at
// once: the first pair shares one coordinate, and the second is the nearest two
// distinct coordinates can be, one float64 step apart in latitude. A proximity
// de-duplication radius, a clustering bucket or a coordinate-keyed index drops
// one zone of a pair whatever figure it was sized to.

// placedZone is one zone of a close pair: an id and a coordinate, every other
// field of its record fixed.
type placedZone struct {
	id       int32
	lat, lon float64
}

// closePairs returns the two pairs, each pair at its own place.
func closePairs() [][2]placedZone {
	const (
		sharedLat, sharedLon = 59.3293, 18.0686
		steppedLat, stepLon  = 59.3301, 18.0750
	)

	return [][2]placedZone{
		{{id: 7001, lat: sharedLat, lon: sharedLon}, {id: 7002, lat: sharedLat, lon: sharedLon}},
		{{id: 7003, lat: steppedLat, lon: stepLon}, {id: 7004, lat: math.Nextafter(steppedLat, 90), lon: stepLon}},
	}
}

// record renders z as an all-zones record. Each coordinate is written as the
// shortest decimal that parses back to the same float64, so the step between
// the second pair survives the round trip.
func (z placedZone) record() string {
	return fmt.Sprintf(`{"id":%d,"name":"close %d","latitude":%s,"longitude":%s,`+
		`"dateCreated":"2020-01-02T03:04:05Z","totalTakeovers":3,"takeoverPoints":120,`+
		`"pointsPerHour":2,"region":{"id":141,"name":"Stockholm"}}`,
		z.id, z.id, strconv.FormatFloat(z.lat, 'f', -1, 64), strconv.FormatFloat(z.lon, 'f', -1, 64))
}

// closePairRecords renders every zone of pairs with render, in order.
func closePairRecords(pairs [][2]placedZone, render func(placedZone) string) string {
	var records []string

	for _, pair := range pairs {
		for _, z := range pair {
			records = append(records, render(z))
		}
	}

	return "[" + strings.Join(records, ",") + "]"
}

// assertEachPairWhole fails for every zone of pairs that got lacks, holds more
// than once, or holds at any coordinate but its own — the last being a pair
// merged into one zone that kept both ids.
func assertEachPairWhole(t *testing.T, stage string, pairs [][2]placedZone, got []Zone) {
	t.Helper()

	held := map[int32][]Zone{}
	for _, z := range got {
		held[z.ID] = append(held[z.ID], z)
	}

	for _, pair := range pairs {
		for _, want := range pair {
			switch zs := held[want.id]; {
			case len(zs) == 0:
				t.Errorf("%s dropped zone %d of the close pair %d/%d — FR-032 forbids assuming any minimum distance between two zones",
					stage, want.id, pair[0].id, pair[1].id)
			case len(zs) > 1:
				t.Errorf("%s holds zone %d of the close pair %d/%d %d times, want once", stage, want.id, pair[0].id, pair[1].id, len(zs))
			case zs[0].Latitude != want.lat || zs[0].Longitude != want.lon:
				t.Errorf("%s holds zone %d of the close pair %d/%d at (%v, %v), want its own (%v, %v) — the pair was merged",
					stage, want.id, pair[0].id, pair[1].id, zs[0].Latitude, zs[0].Longitude, want.lat, want.lon)
			}
		}
	}

	if want := 2 * len(pairs); len(got) != want {
		t.Errorf("%s holds %d zones for a set of %d", stage, len(got), want)
	}
}

func TestAClosePairReachesTheStoreAsTwoZones(t *testing.T) {
	t.Parallel()

	pairs := closePairs()
	n := 2 * len(pairs)

	t.Run("by the all-zones run", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{
			inspect: Staged{Rows: n, DistinctIDs: n, CurrentZoneRows: n},
			merged:  Merged{Inserted: n},
		}
		fetch := &fakeFetch{body: []byte(closePairRecords(pairs, placedZone.record)), status: 200}

		if got := testRunner(t, store, fetch.fetch).runOnce(t.Context()); got != OutcomeOK {
			t.Fatalf("the run reached %q, want %q", got, OutcomeOK)
		}

		store.mu.Lock()
		staged := store.staged
		store.mu.Unlock()

		if len(staged) != 1 {
			t.Fatalf("the store was handed %v, want one staging load", stagedSizes(staged))
		}

		assertEachPairWhole(t, "the staging load", pairs, staged[0])
	})

	t.Run("by the zone feed", func(t *testing.T) {
		t.Parallel()

		item := func(z placedZone) string {
			return `{"type":"zone","time":"2026-10-19T21:46:07+0000","zone":` + z.record() + `}`
		}

		store := &fakeFeedStore{now: feedStart}
		fetch := &fakeFeedFetch{body: []byte(closePairRecords(pairs, item))}

		newTestFeedPoller(t, store, fetch.fetch).poll(t.Context())

		store.mu.Lock()
		applied := store.applied
		store.mu.Unlock()

		if len(applied) != 1 {
			t.Fatalf("the store was handed %d feed batches, want 1", len(applied))
		}

		if len(applied[0].Refusals) != 0 {
			t.Errorf("the feed batch refused %+v, want every zone of both pairs storable", applied[0].Refusals)
		}

		assertEachPairWhole(t, "the feed batch", pairs, applied[0].Zones)
	})
}
