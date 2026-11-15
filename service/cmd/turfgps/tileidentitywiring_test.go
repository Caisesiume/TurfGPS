package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
	"github.com/Caisesiume/TurfGPS/service/internal/planner"
	"github.com/Caisesiume/TurfGPS/service/internal/valhalla"
)

// Tests over the tile identity the composition root plans against (#249
// AC5–AC7): the root's planner reads it through `internal/valhalla`'s
// TileIdentity, from the engine at config.EnvValhallaURL and the record at
// config.EnvTileIdentityFile, as `DEPLOYMENT.md § The tile build` ("How the
// service reads it") defines. The zone copy and the markings are injected, so
// each case is the identity's alone; `storewiring_postgis_test.go` plans
// through the store as well.

// loadedFrom is access markings that find nothing on any leg, loaded from the
// extract it names.
type loadedFrom string

func (l loadedFrom) ExtractSHA256(context.Context) (string, error) { return string(l), nil }

func (loadedFrom) Crossings(context.Context, []planner.Edge) ([]planner.Crossing, error) {
	return nil, nil
}

func (loadedFrom) Run(context.Context, []planner.Edge) (planner.RunMeasure, error) {
	return planner.RunMeasure{}, planner.ErrNoLeastRun
}

// aboveFloat is a SHA-256 whose stamp is above 2^53 and odd, so no float64
// holds it.
const aboveFloat = "fedcba9876543211" + "00112233445566778899aabbccddeeff0011223344556677"

// identityLookup configures the engine and, where record is not empty, the
// tile record.
func identityLookup(engine *fakeEngine, record string) config.LookupFunc {
	vars := map[string]string{config.EnvValhallaURL: engine.URL}
	if record != "" {
		vars[config.EnvTileIdentityFile] = record
	}

	return lookupOnly(vars)
}

// refusedFor plans journeyRequest through the root's planner and fails the
// test unless no journey is returned, no leg was routed, and the error wraps
// want.
func refusedFor(t *testing.T, name string, engine *fakeEngine, plans *planner.Planner, want error) {
	t.Helper()

	journey, err := plans.Plan(t.Context(), journeyRequest)
	if journey.Legs != nil || !errors.Is(err, want) || !errors.Is(err, planner.ErrAccessUnchecked) {
		t.Errorf("%s: Plan gave %d leg(s) and %v; want no journey and an error wrapping %v", name, len(journey.Legs), err, want)
	}

	if n := engine.routed(); n != 0 {
		t.Errorf("%s: the engine was asked %d leg(s), want none before the identity is read", name, n)
	}
}

// TestTheRootPlansAgainstTheTileIdentityItReads (#249 AC5): the planner the
// root builds reads the identity through the tile reader — the engine's
// verbose status, before the leg and again before the journey is returned,
// and the record — and plans where it is the markings' extract.
func TestTheRootPlansAgainstTheTileIdentityItReads(t *testing.T) {
	t.Parallel()

	engine := newFakeEngine(t, tileStamp(t, clearSHA256))
	plans := rootPlanner(t, identityLookup(engine, writeTileRecord(t, clearSHA256)), heldZoneCopy{}, loadedFrom(clearSHA256))

	journey, err := plans.Plan(t.Context(), journeyRequest)
	if err != nil || len(journey.Legs) != 1 {
		t.Fatalf("Plan gave %d leg(s) and %v; want the one leg, the tile identity read as %s", len(journey.Legs), err, clearSHA256)
	}

	if got := engine.asked(); got != "/status /route /trace_attributes /status" {
		t.Errorf("the engine was asked at %s, want its status read before the leg and again after it", got)
	}
}

// TestAMissingTileIdentityPlansNothing (#249 AC6, missing): an engine that
// does not answer, or answers a status without osm_changeset — how a zero
// stamp reaches the service — plans no journey, though the record and the
// markings agree.
func TestAMissingTileIdentityPlansNothing(t *testing.T) {
	t.Parallel()

	record := writeTileRecord(t, clearSHA256)

	noStamp := newFakeEngine(t, "")
	refusedFor(t, "a status without osm_changeset", noStamp,
		rootPlanner(t, identityLookup(noStamp, record), heldZoneCopy{}, loadedFrom(clearSHA256)), valhalla.ErrTileIdentityMissing)

	// Nothing listens at its address, so it is asked nothing either.
	stopped := httptest.NewServer(http.NotFoundHandler())
	gone := &fakeEngine{URL: stopped.URL}
	stopped.Close()

	refusedFor(t, "an engine that does not answer", gone,
		rootPlanner(t, identityLookup(gone, record), heldZoneCopy{}, loadedFrom(clearSHA256)), valhalla.ErrTileIdentityMissing)
}

// TestAnUnreadableTileIdentityPlansNothing (#249 AC6, unreadable): a record
// not named, named empty, absent, or holding anything but 64 lower-case
// hexadecimal digits and a newline plans no journey, though the engine
// answers.
func TestAnUnreadableTileIdentityPlansNothing(t *testing.T) {
	t.Parallel()

	engine := newFakeEngine(t, tileStamp(t, clearSHA256))

	malformed := filepath.Join(t.TempDir(), "tiles.sha256")
	if err := os.WriteFile(malformed, []byte(strings.ToUpper(clearSHA256)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	for name, lookup := range map[string]config.LookupFunc{
		"no record named": identityLookup(engine, ""),
		"a record named empty": lookupOnly(map[string]string{
			config.EnvValhallaURL: engine.URL, config.EnvTileIdentityFile: "",
		}),
		"an absent record":   identityLookup(engine, filepath.Join(t.TempDir(), "absent.sha256")),
		"a malformed record": identityLookup(engine, malformed),
	} {
		refusedFor(t, name, engine, rootPlanner(t, lookup, heldZoneCopy{}, loadedFrom(clearSHA256)), valhalla.ErrTileIdentityUnreadable)
	}
}

// TestAStaleTileIdentityPlansNothing (#249 AC6, stale): a record whose first
// 16 digits are not the engine's osm_changeset plans no journey — even where
// the record names the markings' extract, which is the hazard: tiles reloaded
// under a running engine, or a record left from another build. A record whose
// stamp is zero is one of them.
func TestAStaleTileIdentityPlansNothing(t *testing.T) {
	t.Parallel()

	engine := newFakeEngine(t, tileStamp(t, clearSHA256))

	for name, sha := range map[string]string{
		"a record for other tiles":     aboveFloat,
		"a record whose stamp is zero": strings.Repeat("0", 16) + clearSHA256[16:],
	} {
		refusedFor(t, name, engine,
			rootPlanner(t, identityLookup(engine, writeTileRecord(t, sha)), heldZoneCopy{}, loadedFrom(sha)), valhalla.ErrTileIdentityStale)
	}
}

// TestAStampAboveTwoToTheFiftyThirdIsReadExactlyThroughTheRoot (#249 AC7):
// where the stamp is above 2^53, the identity is the record's SHA-256 and the
// journey plans — the case a stamp carried through a float64 fails.
func TestAStampAboveTwoToTheFiftyThirdIsReadExactlyThroughTheRoot(t *testing.T) {
	t.Parallel()

	engine := newFakeEngine(t, tileStamp(t, aboveFloat))
	plans := rootPlanner(t, identityLookup(engine, writeTileRecord(t, aboveFloat)), heldZoneCopy{}, loadedFrom(aboveFloat))

	journey, err := plans.Plan(t.Context(), journeyRequest)
	if err != nil || len(journey.Legs) != 1 {
		t.Fatalf("with osm_changeset %s, Plan gave %d leg(s) and %v; want the journey, the identity read as %s",
			tileStamp(t, aboveFloat), len(journey.Legs), err, aboveFloat)
	}
}

// TestAnIdentityThatDiffersFromTheLoadPlansNothing (#249 AC7): a tile
// identity read whole, and markings loaded from another extract, plan no
// journey (ErrExtractMismatch); the two equal, it plans.
func TestAnIdentityThatDiffersFromTheLoadPlansNothing(t *testing.T) {
	t.Parallel()

	engine := newFakeEngine(t, tileStamp(t, clearSHA256))
	lookup := identityLookup(engine, writeTileRecord(t, clearSHA256))

	journey, err := rootPlanner(t, lookup, heldZoneCopy{}, loadedFrom(aboveFloat)).Plan(t.Context(), journeyRequest)
	if journey.Legs != nil || !errors.Is(err, planner.ErrExtractMismatch) {
		t.Errorf("with the tiles from %s and the markings from %s, Plan gave %d leg(s) and %v; want no journey and ErrExtractMismatch",
			clearSHA256, aboveFloat, len(journey.Legs), err)
	}

	if n := engine.routed(); n != 0 {
		t.Errorf("the engine was asked %d leg(s), want none", n)
	}

	journey, err = rootPlanner(t, lookup, heldZoneCopy{}, loadedFrom(clearSHA256)).Plan(t.Context(), journeyRequest)
	if err != nil || len(journey.Legs) != 1 {
		t.Errorf("with the tiles and the markings both from %s, Plan gave %d leg(s) and %v; want the journey", clearSHA256, len(journey.Legs), err)
	}
}
