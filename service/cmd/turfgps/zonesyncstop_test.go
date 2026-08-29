package main

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
)

// Tests over the one thing startZoneSync tells main that main acts on: whether
// the sync ENDED ON ITS OWN. That boolean is the whole input to the exit status
// this process leaves behind, and it arrived with no assertion under it.
//
// ---------------------------------------------------------------------------
// WHAT THE BOOLEAN COSTS WHEN IT IS WRONG, IN EACH DIRECTION. Read before
// editing.
//
// FALSE REPORTED AS TRUE is the expensive one, and it is the direction the
// unconfigured case sits in. main maps true to os.Exit(1) so that a supervisor
// replaces a process whose sync is gone. A service started with no sync
// configured — which is the configuration `NFR-003` measures the image in, and
// the one every image test and every clean host starts in — has no sync to
// lose, and a waiter reporting true there makes every ordinary stop look like a
// casualty: the image exits 1 on a clean shutdown, and a runtime restarting on
// failure crash-loops it. Nothing in the logs contradicts that, because the
// drain completed and every field of the shutdown looks ordinary.
//
// TRUE REPORTED AS FALSE is the failure C2-06 was written to close: a process
// serving a static route that answers 200 while the only job it does has
// stopped, exiting 0 so that nothing replaces it.
//
// NO DATABASE IS REACHED BY ANY TEST HERE, and the construction that guarantees
// it is deliberate rather than incidental. The configured case cancels its
// context BEFORE startZoneSync is called, so `zonesync.Scheduler.Run` returns
// at the `ctx.Err()` check it opens its loop with — ahead of untilDue, which is
// the first thing that would touch a store. The DSN names port 1 as a second
// guard: if that ordering ever changes, the test fails on a refused connection
// rather than quietly finding a developer's own PostgreSQL and running the
// currency query against it. There is no database on this host (`DEP-01`), no
// pgxpool fake here, and no SQL sent.
//
// WHAT THIS FILE CANNOT ASSERT, STATED RATHER THAN LEFT AS A GAP. The `true`
// side — a sync that really died, mapped to exit 1 — is not reachable from any
// test-only input. `died` has two writers in main.go and one of them is dead
// code: `Scheduler.Run` returns nil on every path it has, as its own comment at
// that branch says, so the live writer is the recover that re-raises a panic
// out of the sync goroutine. No environment variable induces that panic, and
// startZoneSync builds its own scheduler rather than accepting one, so nothing
// here can inject a port that panics.
//
// TWO THINGS BLOCK IT, NOT ONE, AND A SEAM ALONE BUYS NOTHING. The injection
// point is the first: startZoneSync would have to accept a scheduler or a
// factory before any test could drive `died` true. The second is that the
// mapping itself is inline in `func main()` at main.go:148-151 and ends in a
// direct `os.Exit(1)`, which terminates the test binary at the moment the
// status becomes observable — so even with the seam built and a sync driven to
// its death, there is no value returned to anything and no assertion left alive
// to read one. Closure therefore needs main split into an int-returning `run()`
// whose status a test can compare, or the mapping asserted out of process via
// `exec.Command(os.Args[0], "-test.run=…")` under an env guard, IN ADDITION TO
// the injection point. Both are production changes this file is not licensed to
// make, and they are tracked together as `FW-22`. The gap is reported with the
// work, at its true cost, rather than papered over with a test that asserts the
// mapping against a copy of it.
// ---------------------------------------------------------------------------

// noSyncConfigured puts the environment in the state `NFR-003` measures: none
// of the three keys that turn the sync on.
//
// Empty rather than unset, because `config.LoadZoneSync` counts a key set to
// the empty string as unset and t.Setenv cannot remove one. That equivalence is
// the loader's, is tested in its own package, and is relied on here so this
// test does not have to mutate and restore the process environment by hand.
func noSyncConfigured(t *testing.T) {
	t.Helper()

	for _, key := range []string{config.EnvDatabaseURL, config.EnvAllZonesURL, config.EnvInterval} {
		t.Setenv(key, "")
	}
}

// syncConfigured turns the sync on with values that reach nothing.
//
// The endpoint is never fetched and the DSN is never dialled; see the note on
// port 1 in the file header. The interval is long enough that a scheduler which
// somehow reached its loop would sleep rather than attempt.
func syncConfigured(t *testing.T) {
	t.Helper()

	t.Setenv(config.EnvDatabaseURL, "postgres://turfgps@127.0.0.1:1/turfgps")
	t.Setenv(config.EnvAllZonesURL, "https://example.test/zones")
	t.Setenv(config.EnvInterval, "1h")
}

// countedStop is a root-cancel that records whether the composition root used
// it. Calling it is how the sync takes the process down, so a call is an
// observable event of its own and not merely a side effect.
type countedStop struct {
	calls  atomic.Int32
	cancel context.CancelFunc
}

func (s *countedStop) stop() {
	s.calls.Add(1)

	s.cancel()
}

// TestAnUnconfiguredSyncIsNotReportedAsHavingDied binds the branch C2-06 landed
// with no assertion under it: a sync that was never started has not died.
func TestAnUnconfiguredSyncIsNotReportedAsHavingDied(t *testing.T) {
	noSyncConfigured(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	stop := &countedStop{cancel: cancel}

	wait, err := startZoneSync(ctx, stop.stop)
	if err != nil {
		t.Fatalf("no zone sync is configured and startZoneSync refused to start: %v — main maps that refusal to exit 1, so a service with no sync would not start at all", err)
	}

	if wait == nil {
		t.Fatal("no zone sync is configured and startZoneSync returned no waiter, want one that reports false: main calls it unconditionally")
	}

	if died := wait(); died {
		t.Errorf("no zone sync was configured and the waiter reports that one died, want false — main maps true to exit 1, so the image `NFR-003` measures would report every clean shutdown as a casualty and crash-loop under a supervisor that restarts on failure")
	}

	if calls := stop.calls.Load(); calls != 0 {
		t.Errorf("no zone sync was configured and the root context was cancelled %d time(s), want 0 — cancelling it stops the HTTP server, so this is a sync that was never started taking the request surface down with it", calls)
	}
}

// TestAStoppedSyncIsNotReportedAsHavingDied binds the other direction of the
// same boolean: a sync that stopped BECAUSE IT WAS ASKED TO is a shutdown, not
// a casualty.
//
// It is the assertion that separates the two states main cannot otherwise tell
// apart. Set `died` unconditionally — the obvious simplification, since the
// goroutine sets it on every path that reaches its end abnormally — and every
// rolling deploy exits 1, every SIGTERM reads as a crash, and the exit status
// stops carrying any information at all. Nothing else in this suite notices.
func TestAStoppedSyncIsNotReportedAsHavingDied(t *testing.T) {
	syncConfigured(t)

	ctx, cancel := context.WithCancel(t.Context())

	// Cancelled before the sync is started, not after: see the file header.
	// This is the shutdown case with the race against a first attempt removed,
	// which is what keeps it off the database.
	cancel()

	stop := &countedStop{cancel: cancel}

	wait, err := startZoneSync(ctx, stop.stop)
	if err != nil {
		t.Fatalf("a configured zone sync could not be started: %v", err)
	}

	if died := wait(); died {
		t.Errorf("the zone sync was asked to stop and the waiter reports that it died on its own, want false — main maps true to exit 1, so an ordinary shutdown would be reported to a supervisor as a failure and the exit status would stop distinguishing the two")
	}

	if calls := stop.calls.Load(); calls != 0 {
		t.Errorf("the zone sync was asked to stop and it cancelled the root context %d time(s), want 0 — that cancel exists to take the process down when the sync dies, and a shutdown is not a death", calls)
	}
}

// TestAPartlyConfiguredSyncIsRefusedRatherThanRunWithout covers the other input
// to an exit 1, and it is the reachable half of that mapping.
//
// A refusal is what main turns into "the zone sync is configured and could not
// start" and exit 1. The failure it excludes is the quiet one: three keys where
// two were meant, the loader returning nil rather than an error, and a service
// that starts, serves, logs nothing alarming and never refreshes the copy — the
// operator having done everything but the third variable.
func TestAPartlyConfiguredSyncIsRefusedRatherThanRunWithout(t *testing.T) {
	noSyncConfigured(t)
	t.Setenv(config.EnvDatabaseURL, "postgres://turfgps@127.0.0.1:1/turfgps")

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	stop := &countedStop{cancel: cancel}

	wait, err := startZoneSync(ctx, stop.stop)
	if err == nil {
		t.Fatalf("a zone sync configured with %s alone was accepted (waiter present: %v), want a refusal — main exits 1 on one, and on the other it serves indefinitely without ever refreshing the copy",
			config.EnvDatabaseURL, wait != nil)
	}
}
