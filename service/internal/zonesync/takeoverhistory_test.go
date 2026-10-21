package zonesync

import (
	"fmt"
	"reflect"
	"testing"
)

// `FR-031` (#31), where a zone's takeover history first passes.
//
// NEITHER CRITERION IS EXERCISED HERE, AND THIS FILE DOES NOT CLAIM TO. Both
// are conditioned on a signal derived from a zone's takeover history being
// computed, and this module computes none: takeover rate is owed to a later
// batch. A test that computed nothing and reported green would close #31 on
// the vacuous pass that story forbids.
//
// What is held is the input any such signal will read. A zone reporting
// takeovers and no `dateLastTaken`, no `currentOwner`, or neither reaches the
// store with its `totalTakeovers` whole — the same zone the record carrying
// both maps to. An ingest that read the absence as "never taken", by zeroing
// the count or refusing the record, would leave the later signal nothing but
// the conclusion FR-031 forbids, and fails here.

// historyTakeovers is a lifetime count of the size `Architecture.md § Volatile
// and optional fields` records on zones carrying no date at all.
const historyTakeovers = 318

// ownershipVariants are the record's ownership members: both present, and each
// absence the criteria name. Each is spliced in ahead of `region`.
var ownershipVariants = []struct {
	name    string
	members string
}{
	{"both present", `"dateLastTaken":"2026-10-19T22:00:00+0000","currentOwner":{"id":330825,"name":"Zinky"},`},
	{"no dateLastTaken", `"currentOwner":{"id":330825,"name":"Zinky"},`},
	{"no currentOwner", `"dateLastTaken":"2026-10-19T22:00:00+0000",`},
	{"neither", ``},
}

// historyRecord is one record reporting historyTakeovers, with ownership
// spliced in.
func historyRecord(ownership string) string {
	return fmt.Sprintf(`{"id":7101,"name":"Held before","latitude":59.3293,"longitude":18.0686,`+
		`"dateCreated":"2016-05-04T03:02:01Z","totalTakeovers":%d,"takeoverPoints":95,"pointsPerHour":1,`+
		`%s"region":{"id":141,"name":"Stockholm"}}`, historyTakeovers, ownership)
}

func TestAnAbsentOwnershipFieldLeavesTheTakeoverHistoryWhole(t *testing.T) {
	t.Parallel()

	// ingest returns the one zone the record reaches the store as, by each
	// path, or fails the case.
	ingest := map[string]func(t *testing.T, record string) Zone{
		"the all-zones run": func(t *testing.T, record string) Zone {
			t.Helper()

			store := &fakeStore{inspect: Staged{Rows: 1, DistinctIDs: 1, CurrentZoneRows: 1}, merged: Merged{Inserted: 1}}
			fetch := &fakeFetch{body: []byte("[" + record + "]"), status: 200}

			if got := testRunner(t, store, fetch.fetch).runOnce(t.Context()); got != OutcomeOK {
				t.Fatalf("the run reached %q, want %q", got, OutcomeOK)
			}

			store.mu.Lock()
			defer store.mu.Unlock()

			if len(store.staged) != 1 || len(store.staged[0]) != 1 {
				t.Fatalf("the store was handed %v, want one staging load of the one zone", stagedSizes(store.staged))
			}

			return store.staged[0][0]
		},
		"the zone feed": func(t *testing.T, record string) Zone {
			t.Helper()

			store := &fakeFeedStore{now: feedStart}
			fetch := &fakeFeedFetch{body: []byte(`[{"type":"zone","time":"2026-10-19T21:46:07+0000","zone":` + record + `}]`)}

			newTestFeedPoller(t, store, fetch.fetch).poll(t.Context())

			store.mu.Lock()
			defer store.mu.Unlock()

			if len(store.applied) != 1 || len(store.applied[0].Zones) != 1 {
				t.Fatalf("the store was handed %+v, want one batch holding the one zone", store.applied)
			}

			return store.applied[0].Zones[0]
		},
	}

	for path, reach := range ingest {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			baseline := reach(t, historyRecord(ownershipVariants[0].members))

			for _, v := range ownershipVariants {
				got := reach(t, historyRecord(v.members))

				if got.TotalTakeovers != historyTakeovers {
					t.Errorf("by %s, a zone reporting %d takeovers with %s reached the store with %d — the absence was read as a zone never taken, which FR-031 forbids",
						path, historyTakeovers, v.name, got.TotalTakeovers)
				}

				if !reflect.DeepEqual(got, baseline) {
					t.Errorf("by %s, the zone with %s reached the store as %+v, and with both present as %+v — the absence changed what was carried",
						path, v.name, got, baseline)
				}
			}
		})
	}
}
