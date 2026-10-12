package main

import (
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
)

// Tests over the one mapping in this file that a compiler cannot check:
// newZoneSyncScheduler copying five fields out of `internal/config` into
// `zonesync.Config`.
//
// ---------------------------------------------------------------------------
// WHY A MAPPING NEEDS A TEST AT ALL. Read before editing.
//
// The five lines are `Interval`, `FetchTimeout`, `DatabaseTimeout`,
// `MergeTimeout` and `MinZoneRatio`, each assigned from the field of the same
// name. Four of the five are `time.Duration`. Cross any two of those four and
// every type still checks, every existing test still passes, and the service
// runs each operation on another one's budget — the merge on the short one, or
// a run that gives up on the endpoint after the interval it was supposed to
// wait. Nothing observable says so until an operator raises one budget and
// watches a different one move, which is the same defect the override block in
// `internal/config` carries and is tested for there.
//
// A mapping is exactly where this hides. It is the code with the least to say
// and the most to get silently wrong, and it is the last place anyone thinks to
// look, because five lines that each repeat a name read as correct on sight.
//
// HOW THE MAPPING IS MADE OBSERVABLE, GIVEN THAT THE SCHEDULER TELLS NOBODY
// WHAT IT WAS BUILT WITH. `zonesync.Scheduler` exposes none of these values —
// deliberately, and this test does not ask for a getter to be added so that it
// can be written. What the package does expose is its REFUSALS: newRunner
// checks the three budgets in a fixed order and names each one in the error it
// returns, with a comment at that loop saying it is ordered so that the same
// misconfiguration always names the same field, and NewScheduler names the
// interval. So the mapping is read backwards. Make exactly one field
// unacceptable, and the name in the refusal says which field the value
// actually reached. Cross two of them and the refusal names the other one.
//
// This binds the test to `internal/zonesync`'s refusal wording, which is a real
// cost and is the narrower one. The alternative was an exported accessor on
// Scheduler that exists only for this test — widening a package's API, on the
// boundary `FR-022` AC2 is enforced over, to observe a wiring detail. A wording
// change here fails loudly and is a one-line edit; an accessor would be
// permanent.
//
// WHAT THIS FILE DOES NOT TEST. It builds a `syncstore.Store` on the way past,
// because newZoneSyncScheduler does, and it asserts nothing whatever about it.
// That package is unexercised and says so in its own doc comment; nothing here
// changes that.
// ---------------------------------------------------------------------------

// unreachedPool is the pool argument, and it is a zero value on purpose.
//
// `syncstore.New` nil-checks its pool and stores it, and nothing this file
// calls goes on to use it: every case below is refused during construction, and
// the one case that is not refused never calls Run. No connection is opened and
// no address is dialled, which is the point — there is no database on this host
// and a test that pretended otherwise would be measuring the pretence.
//
// If the composition root ever starts touching the pool during construction,
// this panics rather than passing, which is the failure direction to want.
func unreachedPool() *pgxpool.Pool { return &pgxpool.Pool{} }

// wireable is a configuration newZoneSyncScheduler accepts whole.
//
// Every duration is distinct from every other, so a refusal naming a budget
// identifies one field rather than a set of fields that happen to share a
// figure. The values are arbitrary; only their being unlike each other matters.
func wireable() config.ZoneSync {
	return config.ZoneSync{
		DatabaseURL:      "postgres://localhost/turfgps",
		AllZonesURL:      "https://example.test/zones",
		Interval:         11 * time.Minute,
		FetchTimeout:     13 * time.Second,
		DatabaseTimeout:  17 * time.Second,
		MergeTimeout:     19 * time.Second,
		MaxResponseBytes: 1 << 20,
		MinZoneRatio:     0.9,
	}
}

// TestTheSchedulerIsWiredFromTheConfigurationFieldByField makes each of the
// five mapped fields unacceptable in turn and reads back which field the
// refusal names.
func TestTheSchedulerIsWiredFromTheConfigurationFieldByField(t *testing.T) {
	t.Parallel()

	// The whole-configuration case first. Without it, a mapping that dropped a
	// field entirely — leaving the zero value in `zonesync.Config` — would make
	// every subtest below fail on that one field's refusal and look like a
	// crossed pair rather than a missing line. It also states that the five
	// values this test then breaks are acceptable to begin with.
	whole := wireable()

	if _, err := newZoneSyncScheduler(&whole, unreachedPool()); err != nil {
		t.Fatalf("a fully valid configuration was refused: %v — every case below breaks exactly one field of it, so nothing below measures anything until this passes", err)
	}

	for _, c := range []struct {
		field string

		// spoil makes one field unacceptable and nothing else.
		spoil func(*config.ZoneSync)

		// refusal is the wording `internal/zonesync` uses for that field. It is
		// that package's, not this one's, and it is written out here because
		// the whole question this test asks is whether the field that was
		// spoiled is the field the refusal names.
		refusal string

		// alsoNamed is the wording that would appear instead if this field had
		// been crossed with another. Asserted absent, so the failure says which
		// crossing rather than only that something is wrong.
		alsoNamed []string
	}{
		{
			field:     "Interval",
			spoil:     func(c *config.ZoneSync) { c.Interval = 0 },
			refusal:   "the refresh interval is not positive",
			alsoNamed: []string{"timeout", "ratio"},
		},
		{
			field:     "FetchTimeout",
			spoil:     func(c *config.ZoneSync) { c.FetchTimeout = 0 },
			refusal:   "the fetch timeout is not positive",
			alsoNamed: []string{"the database timeout", "the merge timeout", "interval", "ratio"},
		},
		{
			field:     "DatabaseTimeout",
			spoil:     func(c *config.ZoneSync) { c.DatabaseTimeout = 0 },
			refusal:   "the database timeout is not positive",
			alsoNamed: []string{"the fetch timeout", "the merge timeout", "interval", "ratio"},
		},
		{
			field:     "MergeTimeout",
			spoil:     func(c *config.ZoneSync) { c.MergeTimeout = 0 },
			refusal:   "the merge timeout is not positive",
			alsoNamed: []string{"the fetch timeout", "the database timeout", "interval", "ratio"},
		},
		{
			field:     "MinZoneRatio",
			spoil:     func(c *config.ZoneSync) { c.MinZoneRatio = 0 },
			refusal:   "the minimum zone ratio",
			alsoNamed: []string{"timeout", "interval"},
		},
	} {
		t.Run(c.field, func(t *testing.T) {
			t.Parallel()

			cfg := wireable()
			c.spoil(&cfg)

			scheduler, err := newZoneSyncScheduler(&cfg, unreachedPool())
			if err == nil {
				t.Fatalf("config.ZoneSync.%s was made unacceptable and the scheduler was built anyway (%v), so that field reaches nothing: the mapping in newZoneSyncScheduler is not carrying it",
					c.field, scheduler != nil)
			}

			if !strings.Contains(err.Error(), c.refusal) {
				t.Errorf("config.ZoneSync.%s was made unacceptable and the refusal reads %q, want it to name %q — the refusal names the field the value actually reached, so a different name means %s is mapped onto the wrong field of zonesync.Config",
					c.field, err, c.refusal, c.field)
			}

			for _, other := range c.alsoNamed {
				if strings.Contains(err.Error(), other) {
					t.Errorf("config.ZoneSync.%s was made unacceptable and the refusal reads %q, which names %q: that is the field it was crossed with",
						c.field, err, other)
				}
			}
		})
	}
}

// TestTheEndpointAndTheCeilingReachTheClient covers the two fields the mapping
// passes POSITIONALLY rather than by name, which is the other shape a wiring
// test has to watch.
//
// They are a string and an int64, so the compiler catches a straight swap and
// this test is not there for that. It is there because both are silently
// droppable: a `turf.NewClient("", ...)` or a zero ceiling would be refused,
// and a client built from neither field would be built from nothing and never
// complain. Making each unacceptable in turn says the value travelled.
func TestTheEndpointAndTheCeilingReachTheClient(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		field   string
		spoil   func(*config.ZoneSync)
		refusal string
	}{
		{"AllZonesURL", func(c *config.ZoneSync) { c.AllZonesURL = "" }, "no all-zones endpoint"},
		{"MaxResponseBytes", func(c *config.ZoneSync) { c.MaxResponseBytes = 0 }, "response ceiling"},
	} {
		t.Run(c.field, func(t *testing.T) {
			t.Parallel()

			cfg := wireable()
			c.spoil(&cfg)

			if _, err := newZoneSyncScheduler(&cfg, unreachedPool()); err == nil {
				t.Errorf("config.ZoneSync.%s was made unacceptable and the scheduler was built anyway, so that field reaches no client", c.field)
			} else if !strings.Contains(err.Error(), c.refusal) {
				t.Errorf("config.ZoneSync.%s was made unacceptable and the refusal reads %q, want it to name %q", c.field, err, c.refusal)
			}
		})
	}
}
