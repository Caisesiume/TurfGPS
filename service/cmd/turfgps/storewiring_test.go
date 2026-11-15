package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
	"github.com/Caisesiume/TurfGPS/service/internal/httpapi"
)

// unreachableStore is a DSN nothing answers at: port 1 on loopback, as
// `zonesyncstop_test.go` uses for the same reason, with a short connect
// timeout so a refused dial is not waited on.
const unreachableStore = "postgres://turfgps@127.0.0.1:1/turfgps?connect_timeout=2&sslmode=disable"

// TestAnUnreachableStoreAnswersTheUnreadable503 (#249 AC2, unreachable): the
// planner the root builds on a job's pool reads the zone copy through it, and
// a store that does not answer refuses every plan with 503 saying the copy
// could not be read — never that none is held — before a leg is asked of the
// engine.
func TestAnUnreachableStoreAnswersTheUnreadable503(t *testing.T) {
	t.Parallel()

	syncCfg := wireable()
	syncCfg.DatabaseURL = unreachableStore

	pools, err := openJobPools(t.Context(), &syncCfg, nil)
	if err != nil {
		t.Fatalf("openJobPools: %v", err)
	}
	t.Cleanup(pools.close)

	engine := newFakeEngine(t, tileStamp(t, clearSHA256))
	lookup := identityLookup(engine, writeTileRecord(t, clearSHA256))

	plans, err := newPlanner(lookup, pools.request())
	if err != nil || plans == nil {
		t.Fatalf("newPlanner on the sync's pool gave %v, %v; want a planner", plans, err)
	}

	handler, err := newHandler(lookup, httpapi.WithPlanner(plans))
	if err != nil {
		t.Fatalf("newHandler: %v", err)
	}

	rec := postJourney(t, handler)
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "could not be read") {
		t.Errorf("POST /journeys over an unreachable store answered %d %q, want 503 saying the zone copy could not be read", rec.Code, rec.Body)
	}

	if got := engine.asked(); got != "" {
		t.Errorf("the engine was asked at %s, want nothing before the zone copy is read", got)
	}
}

// TestTheRequestPathReadsThroughAJobsPool (#249 AC1): the pool the planner
// reads through is the sync's where a sync is configured, the feed's where
// only the feed is, and none where neither is — never one of its own.
func TestTheRequestPathReadsThroughAJobsPool(t *testing.T) {
	t.Parallel()

	syncCfg := wireable()
	syncCfg.DatabaseURL = unreachableStore

	feed := &zoneFeed{cfg: &config.ZoneFeed{DatabaseURL: unreachableStore}}

	for name, tc := range map[string]struct {
		sync *config.ZoneSync
		feed *zoneFeed
	}{
		"both":     {&syncCfg, feed},
		"the sync": {&syncCfg, nil},
		"the feed": {nil, feed},
		"neither":  {nil, nil},
	} {
		pools, err := openJobPools(t.Context(), tc.sync, tc.feed)
		if err != nil {
			t.Fatalf("%s: openJobPools: %v", name, err)
		}

		want := pools.sync
		if want == nil {
			want = pools.feed
		}

		if got := pools.request(); got != want || (tc.sync == nil && tc.feed == nil) != (got == nil) {
			t.Errorf("with %s configured the request path reads through %p, want %p", name, got, want)
		}

		if (tc.sync != nil) != (pools.sync != nil) || (tc.feed != nil) != (pools.feed != nil) {
			t.Errorf("with %s configured the pools opened are sync %v and feed %v, want one per job configured", name, pools.sync != nil, pools.feed != nil)
		}

		pools.close()
	}
}
