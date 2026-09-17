package config

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func env(pairs map[string]string) LookupFunc {
	return func(name string) (string, bool) {
		v, ok := pairs[name]

		return v, ok
	}
}

// complete is the smallest configuration that runs a sync.
//
// The interval here is a test's own figure and is deliberately unlike the one
// the endpoint permits, whose single home is `Architecture.md § Retrieving
// zones`. What is under test is that the value is carried through, not what the
// value is.
func complete() map[string]string {
	return map[string]string{
		EnvDatabaseURL: "postgres://localhost/turfgps",
		EnvAllZonesURL: "https://example.test/zones",
		EnvInterval:    "3m",
	}
}

// TestNoSyncConfiguredIsNotAnError binds the property `NFR-003` measures: the
// service starts and serves on a clean host with nothing beside it. A service
// that refused to start without a database and an endpoint would spend that
// property to gain a check the sync's own logging already makes.
func TestNoSyncConfiguredIsNotAnError(t *testing.T) {
	t.Parallel()

	cfg, err := LoadZoneSync(env(nil))
	if err != nil {
		t.Fatalf("an unconfigured sync returned %v, want no error", err)
	}

	if cfg != nil {
		t.Errorf("an unconfigured sync returned %+v, want nil", cfg)
	}
}

// TestPartialConfigurationIsRefused is the other half of that decision, and the
// reason it is not simply "absent means off". A half-configured sync that
// started quietly would be indistinguishable from one deliberately running
// without a sync, and the operator would learn about the typo from the copy
// never getting any fresher.
func TestPartialConfigurationIsRefused(t *testing.T) {
	t.Parallel()

	for _, missing := range []string{EnvDatabaseURL, EnvAllZonesURL, EnvInterval} {
		t.Run("without "+missing, func(t *testing.T) {
			t.Parallel()

			pairs := complete()
			delete(pairs, missing)

			cfg, err := LoadZoneSync(env(pairs))
			if err == nil {
				t.Fatalf("a configuration missing %s was accepted as %+v, want it refused", missing, cfg)
			}

			if !strings.Contains(err.Error(), missing) {
				t.Errorf("the refusal reads %q, want it to name %s", err, missing)
			}
		})
	}
}

// TestTheIntervalIsCarriedFromTheEnvironmentAndNowhereElse is the constraint
// this package exists for. The interval has one home and it is a document; this
// package must hold no default that could stand in for it.
func TestTheIntervalIsCarriedFromTheEnvironmentAndNowhereElse(t *testing.T) {
	t.Parallel()

	pairs := complete()
	pairs[EnvInterval] = "17m30s"

	cfg, err := LoadZoneSync(env(pairs))
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	if want := 17*time.Minute + 30*time.Second; cfg.Interval != want {
		t.Errorf("the interval loaded as %s, want %s", cfg.Interval, want)
	}

	// The same load with the variable empty must fail rather than fall back,
	// which is what says there is no compiled-in interval to fall back to.
	pairs[EnvInterval] = ""

	if cfg, err := LoadZoneSync(env(pairs)); err == nil {
		t.Errorf("an empty interval loaded as %+v, want it refused: a default here would be a second home for a figure that has one", cfg)
	}
}

// TestTheEndpointIsValidatedAtLoad catches a misconfiguration at start-up rather
// than once per interval for ever.
func TestTheEndpointIsValidatedAtLoad(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]string{
		"no scheme":      "example.test/zones",
		"a bare path":    "/zones",
		"a wrong scheme": "ftp://example.test/zones",

		// http is the dangerous row of this table rather than the obvious one:
		// unlike the three above it parses, resolves, and fetches. It was
		// accepted as readily as https, so the corpus that becomes the
		// authoritative `zone` geometry crossed the network in the clear
		// whenever the variable happened to be typed without the s — and every
		// assertion downstream of the wire accepts a rewritten body whose
		// coordinates are merely plausible.
		"plain http": "http://example.test/zones",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pairs := complete()
			pairs[EnvAllZonesURL] = value

			if cfg, err := LoadZoneSync(env(pairs)); err == nil {
				t.Errorf("the endpoint %q was accepted as %+v, want it refused", value, cfg)
			}
		})
	}
}

// TestTheTunablesHaveDefaultsAndAreOverridable separates the two kinds of value
// this package holds: the operational budgets are its own and carry defaults,
// while the interval and the endpoint are not and do not.
func TestTheTunablesHaveDefaultsAndAreOverridable(t *testing.T) {
	t.Parallel()

	cfg, err := LoadZoneSync(env(complete()))
	if err != nil {
		t.Fatalf("loading: %v", err)
	}

	if cfg.FetchTimeout <= 0 || cfg.DatabaseTimeout <= 0 || cfg.MergeTimeout <= 0 {
		t.Errorf("a budget defaulted to a non-positive value: %+v", cfg)
	}

	if cfg.MaxResponseBytes <= 0 {
		t.Errorf("the response ceiling defaulted to %d", cfg.MaxResponseBytes)
	}

	if cfg.MinZoneRatio <= 0 || cfg.MinZoneRatio > 1 {
		t.Errorf("the floor ratio defaulted to %v, outside (0, 1]", cfg.MinZoneRatio)
	}

	pairs := complete()
	pairs[EnvFetchTimeout] = "45s"
	pairs[EnvMinZoneRatio] = "0.5"
	pairs[EnvMaxResponseBytes] = "1024"

	cfg, err = LoadZoneSync(env(pairs))
	if err != nil {
		t.Fatalf("loading the overrides: %v", err)
	}

	if cfg.FetchTimeout != 45*time.Second || cfg.MinZoneRatio != 0.5 || cfg.MaxResponseBytes != 1024 {
		t.Errorf("the overrides loaded as %+v", cfg)
	}
}

// TestAMalformedTunableIsRefusedRatherThanDefaulted keeps a typo from reading as
// an intention. Falling back to the default here would run the sync on a budget
// the operator did not choose and did not know they had not chosen.
func TestAMalformedTunableIsRefusedRatherThanDefaulted(t *testing.T) {
	t.Parallel()

	for name, pair := range map[string][2]string{
		"a duration that is not one": {EnvFetchTimeout, "soon"},
		"a negative duration":        {EnvDatabaseTimeout, "-1s"},
		"a zero duration":            {EnvMergeTimeout, "0s"},
		"a ratio above one":          {EnvMinZoneRatio, "1.5"},
		"a ratio of zero":            {EnvMinZoneRatio, "0"},
		"a byte count that is not":   {EnvMaxResponseBytes, "lots"},

		// The ceiling was bounded below and not above, so this value loaded
		// and reached the adapter, where it is what breaks the read's own
		// arithmetic. It is refused here as a malformed tunable and not as a
		// large one: there is no reading of it the service can honour.
		"a byte count above what one response is read into": {EnvMaxResponseBytes, "9223372036854775807"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pairs := complete()
			pairs[pair[0]] = pair[1]

			if cfg, err := LoadZoneSync(env(pairs)); err == nil {
				t.Errorf("%s=%q was accepted as %+v, want it refused", pair[0], pair[1], cfg)
			}
		})
	}
}

// unoverridden is the configuration `complete()` loads, written out field by
// field so a test can say what a single override must and must not change.
//
// It names the default CONSTANTS rather than the figures behind them. Copying
// "5m" here would put a second home under a value whose home is the constant a
// few lines away in config.go, and the copy would go stale silently — the
// failure the constants exist to prevent, reintroduced by the test that checks
// them.
func unoverridden() ZoneSync {
	return ZoneSync{
		DatabaseURL:      complete()[EnvDatabaseURL],
		AllZonesURL:      complete()[EnvAllZonesURL],
		Interval:         3 * time.Minute, // complete()'s own interval, parsed.
		FetchTimeout:     defaultFetchTimeout,
		DatabaseTimeout:  defaultDatabaseTimeout,
		MergeTimeout:     defaultMergeTimeout,
		MaxResponseBytes: defaultMaxResponseBytes,
		MinZoneRatio:     defaultMinZoneRatio,
	}
}

// differences lists the fields on which two configurations disagree.
//
// It exists so the failure below names the field rather than printing two
// structs and leaving a reader to diff eight fields by eye. A swap reports as
// two lines that read as each other's mirror, which is the shape that says
// "swap" rather than "wrong value".
func differences(got, want ZoneSync) []string {
	var out []string

	for _, f := range []struct {
		name      string
		got, want any
	}{
		{"DatabaseURL", got.DatabaseURL, want.DatabaseURL},
		{"AllZonesURL", got.AllZonesURL, want.AllZonesURL},
		{"Interval", got.Interval, want.Interval},
		{"FetchTimeout", got.FetchTimeout, want.FetchTimeout},
		{"DatabaseTimeout", got.DatabaseTimeout, want.DatabaseTimeout},
		{"MergeTimeout", got.MergeTimeout, want.MergeTimeout},
		{"MaxResponseBytes", got.MaxResponseBytes, want.MaxResponseBytes},
		{"MinZoneRatio", got.MinZoneRatio, want.MinZoneRatio},
	} {
		if f.got != f.want {
			out = append(out, fmt.Sprintf("%s = %v, want %v", f.name, f.got, f.want))
		}
	}

	return out
}

// override is one environment variable, the value it is set to, and the single
// change that setting it may make.
type override struct {
	env   string
	raw   string
	apply func(*ZoneSync)
}

// overrides pairs every overridable variable with the one field it owns.
//
// EVERY VALUE IS DISTINCT FROM EVERY OTHER AND FROM EVERY DEFAULT, which is the
// property the table turns on rather than a matter of taste. Three of these
// fields are `time.Duration` and two of the defaults behind them are the same
// figure, so a value shared between two rows would let a mapping that sent an
// override to the wrong field of the same type produce the expected struct
// anyway. The figures are arbitrary and mean nothing beyond being unlike each
// other; they are prime seconds so that no two can be confused by arithmetic.
func overrides() []override {
	return []override{
		{EnvFetchTimeout, "41s", func(c *ZoneSync) { c.FetchTimeout = 41 * time.Second }},
		{EnvDatabaseTimeout, "43s", func(c *ZoneSync) { c.DatabaseTimeout = 43 * time.Second }},
		{EnvMergeTimeout, "47s", func(c *ZoneSync) { c.MergeTimeout = 47 * time.Second }},
		{EnvMaxResponseBytes, "53", func(c *ZoneSync) { c.MaxResponseBytes = 53 }},
		{EnvMinZoneRatio, "0.59", func(c *ZoneSync) { c.MinZoneRatio = 0.59 }},
	}
}

// TestEachOverrideReachesItsOwnFieldAndNoOther discriminates the override block
// of LoadZoneSync, which nothing else does.
//
// ---------------------------------------------------------------------------
// WHY THE OVERRIDE TEST NEXT DOOR IS NOT THIS TEST. Read before editing.
//
// TestTheTunablesHaveDefaultsAndAreOverridable sets three variables and checks
// three fields. It cannot fail on the defect this block is actually exposed to:
// the block is five near-identical lines, each naming a variable and a field
// that happen to share a name, and every one of those variables is a string.
// Swap the variables on the database-timeout and merge-timeout lines and the
// compiler is satisfied, every existing test still passes — neither of those
// two is ever set — and the service runs the merge on the short budget and the
// short operations on the long one. Nothing observable says so until an
// operator raises one budget and watches the other move.
//
// So the assertion here is over the WHOLE configuration and not over the field
// under test. Checking only the overridden field would catch a variable read
// into no field at all, and miss the swap entirely, because a swap is only
// visible in the field that was not supposed to change.
//
// The subtests are one override at a time by design. Setting all five at once
// is the weaker arrangement — every field then differs from its default, so a
// mapping that crossed two of them still produces five changed fields and the
// crossing has to be spotted in the values. One at a time makes it structural:
// exactly one field may move, and a swap moves two.
// ---------------------------------------------------------------------------
func TestEachOverrideReachesItsOwnFieldAndNoOther(t *testing.T) {
	t.Parallel()

	// The unoverridden load is asserted first, because every subtest below is
	// stated as a difference from it. If this is wrong they are all measuring
	// against the wrong baseline, and the block they exist to discriminate
	// would be reported as sound while the defaults were the thing at fault.
	base, err := LoadZoneSync(env(complete()))
	if err != nil {
		t.Fatalf("loading with no overrides: %v", err)
	}

	if diff := differences(*base, unoverridden()); len(diff) != 0 {
		t.Fatalf("an unoverridden load disagrees with the defaults it is built from: %v", diff)
	}

	for _, o := range overrides() {
		t.Run(o.env, func(t *testing.T) {
			t.Parallel()

			want := unoverridden()
			o.apply(&want)

			pairs := complete()
			pairs[o.env] = o.raw

			got, err := LoadZoneSync(env(pairs))
			if err != nil {
				t.Fatalf("loading with %s=%q: %v", o.env, o.raw, err)
			}

			if diff := differences(*got, want); len(diff) != 0 {
				t.Errorf("setting %s=%q and nothing else changed the wrong field(s): %v — exactly one field may move, and two moving is the two variables having been crossed in the override block",
					o.env, o.raw, diff)
			}
		})
	}
}

// TestEveryOverrideAppliedAtOnceKeepsItsOwnField is the second arrangement of
// the same question, and it is here for the one defect the subtests above
// cannot reach: a line that reads the right variable into the right field but
// is later clobbered by another, which is invisible while only one variable is
// ever set.
func TestEveryOverrideAppliedAtOnceKeepsItsOwnField(t *testing.T) {
	t.Parallel()

	want := unoverridden()
	pairs := complete()

	for _, o := range overrides() {
		o.apply(&want)

		pairs[o.env] = o.raw
	}

	got, err := LoadZoneSync(env(pairs))
	if err != nil {
		t.Fatalf("loading with every override set: %v", err)
	}

	if diff := differences(*got, want); len(diff) != 0 {
		t.Errorf("with every variable set to a value unlike every other, the configuration loaded wrong: %v", diff)
	}
}
