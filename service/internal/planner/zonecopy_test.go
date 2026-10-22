package planner

import (
	"context"
	"errors"
	"testing"
)

// zoneCopy is a ZoneCopy that reports a copy held, or none, and never fails.
type zoneCopy bool

func (z zoneCopy) Held(context.Context) (bool, error) { return bool(z), nil }

const (
	copyHeld   zoneCopy = true
	noCopyHeld zoneCopy = false
)

// unreadableCopy is a ZoneCopy whose store cannot be read.
type unreadableCopy struct{ err error }

func (u unreadableCopy) Held(context.Context) (bool, error) { return false, u.err }

func straightLegs() *fakeRouting {
	return &fakeRouting{shape: func(from, to Point) []Point { return []Point{from, to} }}
}

var (
	gateOrigin      = Point{Latitude: 59.3293, Longitude: 18.0686}
	gateVia         = Point{Latitude: 57.7826, Longitude: 14.1618}
	gateDestination = Point{Latitude: 57.7089, Longitude: 11.9746}
)

// TestAJourneyIsPlannedFromAnyZoneCopyHeld is #33 AC1, AC2 and AC3 (`FR-024`,
// `FR-033`) at the gate: wherever a copy is held, planning completes and
// returns the journey.
//
// A refresh in progress, a copy as old as the staleness bound, and one older
// than it or with no refresh on record differ only in the store, and the gate
// reads none of that — ZoneCopy has no age and no refresh state to read. That
// each of those copies is reported held is the adapter's test,
// `internal/zonestore/held_postgis_test.go`, so the bound is not restated
// here either.
func TestAJourneyIsPlannedFromAnyZoneCopyHeld(t *testing.T) {
	t.Parallel()

	routing := straightLegs()

	p, err := New(routing, copyHeld)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	req := Request{Origin: gateOrigin, Intermediates: []Point{gateVia}, Destination: gateDestination}

	journey, err := p.Plan(t.Context(), req)
	if err != nil {
		t.Fatalf("Plan refused a journey with a zone copy held: %v; FR-024 and FR-033 plan from any copy held, whatever its age or refresh state", err)
	}

	if len(journey.Legs) != 2 || len(routing.asked) != 2 {
		t.Errorf("the journey has %d legs and %d were routed, want 2 of each", len(journey.Legs), len(routing.asked))
	}
}

// TestNoJourneyIsPlannedWhereNoZoneCopyIsHeld is #33 AC4 (`FR-033`): with no
// local copy of the zone set, the system plans no journey — no journey is
// returned and no leg is routed, so nothing of one is produced either.
func TestNoJourneyIsPlannedWhereNoZoneCopyIsHeld(t *testing.T) {
	t.Parallel()

	routing := straightLegs()

	p, err := New(routing, noCopyHeld)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	journey, err := p.Plan(t.Context(), Request{Origin: gateOrigin, Destination: gateDestination})
	if !errors.Is(err, ErrNoZoneCopy) {
		t.Errorf("Plan with no zone copy held returned error %v, want ErrNoZoneCopy", err)
	}

	if journey.Locations != nil || journey.Legs != nil {
		t.Errorf("Plan with no zone copy held returned the journey %+v, want none", journey)
	}

	if len(routing.asked) != 0 {
		t.Errorf("Plan with no zone copy held routed %+v, want nothing routed", routing.asked)
	}
}

// TestAnUnreadableZoneCopyIsNotReportedAsNone pins the edge between the gate's
// refusal and a failure: a store that could not be read has not been shown to
// hold nothing, so its error is never ErrNoZoneCopy, and no journey is planned
// on an answer nobody received.
func TestAnUnreadableZoneCopyIsNotReportedAsNone(t *testing.T) {
	t.Parallel()

	errStore := errors.New("store unreachable")
	routing := straightLegs()

	p, err := New(routing, unreadableCopy{err: errStore})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	journey, err := p.Plan(t.Context(), Request{Origin: gateOrigin, Destination: gateDestination})

	if !errors.Is(err, ErrZoneCopyUnreadable) || !errors.Is(err, errStore) {
		t.Errorf("Plan over an unreadable store returned error %v, want one wrapping ErrZoneCopyUnreadable and the store's error", err)
	}

	if errors.Is(err, ErrNoZoneCopy) {
		t.Errorf("Plan over an unreadable store returned %v, which reports no copy held: the store was never read", err)
	}

	if journey.Locations != nil || len(routing.asked) != 0 {
		t.Errorf("Plan over an unreadable store returned %+v and routed %+v, want neither", journey, routing.asked)
	}
}

func TestNewRefusesNoZoneCopy(t *testing.T) {
	t.Parallel()

	if _, err := New(straightLegs(), nil); err == nil {
		t.Error("New(routing, nil) returned no error, want a refusal")
	}
}
