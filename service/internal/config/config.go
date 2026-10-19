// Package config reads the service's runtime configuration from the process
// environment.
//
// The environment is the source rather than a file in this repository because
// `DEPLOYMENT.md § Where the deployment configuration lives` puts the host's
// arrangement outside the repository and forbids secret material inside it: the
// service's unit references an out-of-repo source, and this package reads what
// that source sets.
//
// FOUR VALUES DELIBERATELY HAVE NO DEFAULT HERE, and that is the point of the
// package rather than an omission.
//
//   - The refresh interval. Its one home is `Architecture.md § Retrieving
//     zones`, and a default compiled into this package would be a second home
//     for it — one that goes wrong silently when that section moves.
//   - The all-zones endpoint. It is named by role in that same section and
//     reaches the service as a resolved URL, so no API path is written in this
//     module. The VERSION that URL addresses is not this package's to check:
//     `FR-019`'s class is the paths that construct a Turf request together with
//     any configuration they draw a base path from, and `internal/turf` names
//     the version once, as APIVersion, and refuses an endpoint addressing any
//     other.
//   - The Turf API's two published rate limits (`NFR-001`): the general one,
//     whose home is `Architecture.md § Data sources and constraints`, and the
//     all-zones endpoint's own, whose home is `Architecture.md § Retrieving
//     zones`. They are the API's figures, not this service's, so they are
//     carried rather than chosen; `internal/turf` holds every request the
//     process sends to them.
//
// All four therefore arrive from the environment or the sync does not run at
// all.
package config

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Environment names. The prefix keeps the service's variables distinguishable
// in a unit file that also configures the host around it.
const (
	EnvDatabaseURL      = "TURFGPS_DATABASE_URL"
	EnvAllZonesURL      = "TURFGPS_ZONE_SYNC_ALL_ZONES_URL"
	EnvInterval         = "TURFGPS_ZONE_SYNC_INTERVAL"
	EnvFetchTimeout     = "TURFGPS_ZONE_SYNC_FETCH_TIMEOUT"
	EnvDatabaseTimeout  = "TURFGPS_ZONE_SYNC_DB_TIMEOUT"
	EnvMergeTimeout     = "TURFGPS_ZONE_SYNC_MERGE_TIMEOUT"
	EnvMaxResponseBytes = "TURFGPS_ZONE_SYNC_MAX_RESPONSE_BYTES"
	EnvMinZoneRatio     = "TURFGPS_ZONE_SYNC_MIN_ZONE_RATIO"

	// The Turf API's published limits, each as the shortest interval it permits
	// between two requests. TURF rather than ZONE_SYNC, because the limit is
	// the API's and binds every caller of it, of which the sync is the first.
	EnvRequestInterval  = "TURFGPS_TURF_MIN_REQUEST_INTERVAL"
	EnvAllZonesInterval = "TURFGPS_TURF_ALL_ZONES_MIN_INTERVAL"
)

// Defaults for the tunables that are this service's own operational choices
// rather than facts recorded elsewhere. Each is overridable, none of them
// restates a measurement, and none of them is the interval.
const (
	// defaultFetchTimeout bounds one all-zones request end to end. It exists to
	// make the call finite rather than to predict it: the transfer measured in
	// `Architecture.md § Retrieving zones` completes far inside this, and a
	// budget tight enough to be interesting would turn one slow day into a
	// permanently failing sync.
	defaultFetchTimeout = 5 * time.Minute

	// defaultDatabaseTimeout bounds each of the sync's short database
	// operations — the due-gate read, the two sync_run writes, the staging load
	// and its assertions.
	defaultDatabaseTimeout = 2 * time.Minute

	// defaultMergeTimeout bounds the one transaction that takes ROW EXCLUSIVE on
	// `zone`. It is a budget of its own rather than a share of the one above
	// because this is the operation whose overrun is felt outside the sync, so
	// it is the one an operator reaches for first.
	defaultMergeTimeout = 5 * time.Minute

	// defaultMaxResponseBytes caps what one response may put in memory. It is a
	// guard against a runaway or hostile body and never a prediction of the
	// corpus, whose size is measured in `Architecture.md § Retrieving zones` and
	// is not restated here.
	//
	// IT WAS 512 MiB, WHICH IS NOT A CEILING THIS PROCESS SURVIVES, and a
	// ceiling is only a guard if the process is still running when it is
	// reached. The body must be held whole while it is parsed and the parse
	// builds a second representation of the same response beside it, so a
	// response arriving at that ceiling exhausts the process before anything
	// refuses it — which takes down the request path the ceiling exists to
	// protect, the one outcome `FR-022` AC2 forbids, and takes it down in the
	// least legible way available.
	//
	// 128 MiB is this package's own choice in the sense the ratio below is: it
	// leaves the corpus that section measures a great deal of room to grow, and
	// it is small enough that a response reaching it is refused by an error
	// rather than by the kernel. Whether it still clears the corpus is a
	// question for that section, which is where the figure to compare it against
	// lives and where it will move.
	defaultMaxResponseBytes int64 = 128 << 20

	// maxResponseBytesCeiling is the largest ceiling this package accepts from
	// an operator. The variable was bounded below and not above, so every figure
	// from one byte past the default to MaxInt64 was taken as given.
	//
	// WHY AN UNBOUNDED CEILING IS UNSAFE IS NOT ARGUED HERE. `internal/turf`
	// sizes a buffer from this figure and reads through a limit of this figure
	// plus one, so it is that package's arithmetic an absurd value breaks, and
	// its own constant is where that is stated. This bound exists so the refusal
	// names the variable an operator set and the value they set it to, at load,
	// rather than arriving as a constructor failing over a number nothing names.
	//
	// The figure is 512 MiB, and the note above is the reason to stay far under
	// it rather than a reason it is safe: that note is why the default is not
	// 512 MiB, and its argument does not stop holding because the figure was
	// chosen by hand. This is where the package stops honouring the request, not
	// a size it promises the process survives.
	maxResponseBytesCeiling int64 = 512 << 20

	// defaultMinZoneRatio is the floor the staged row count is held to against
	// the rows already in `zone`, for the first assertion of `Architecture.md §
	// The sync write path`. That section requires a floor and states no figure,
	// so this is a choice of this package rather than a fact taken from it. It
	// leaves room for a corpus that shrinks slightly while refusing the
	// truncated response the assertion exists to catch — the failure whose cost
	// `Architecture.md § Absence is recorded and never acted on` argues is
	// unbounded in one direction.
	defaultMinZoneRatio = 0.9
)

// ZoneSync is the configuration one zone-sync worker needs.
type ZoneSync struct {
	// DatabaseURL addresses the store of `Architecture.md § D4`.
	DatabaseURL string

	// AllZonesURL is the resolved all-zones endpoint that
	// `Architecture.md § Retrieving zones` names by role.
	AllZonesURL string

	// Interval is how long must pass between one attempt and the next. Its home
	// is `Architecture.md § Retrieving zones`, and it is never shorter than
	// ShortestInterval(AllZonesInterval, FetchTimeout).
	Interval time.Duration

	FetchTimeout     time.Duration
	DatabaseTimeout  time.Duration
	MergeTimeout     time.Duration
	MaxResponseBytes int64
	MinZoneRatio     float64

	// RequestInterval is the Turf API's general limit, as the shortest interval
	// between two requests. Its home is `Architecture.md § Data sources and
	// constraints`.
	RequestInterval time.Duration

	// AllZonesInterval is the all-zones endpoint's own limit, as the shortest
	// interval between two requests to it. Its home is `Architecture.md §
	// Retrieving zones`.
	AllZonesInterval time.Duration
}

// LookupFunc reads one environment variable and reports whether it was set at
// all. It has the shape of os.LookupEnv so a test can supply its own.
type LookupFunc func(name string) (string, bool)

// LoadZoneSync reads the zone sync's configuration.
//
// It returns (nil, nil) when NONE of the required variables is set, which is
// how a service with no sync configured still starts and serves — the property
// `NFR-003` binds and this story must not spend. It returns an error when SOME
// of them are set: partial configuration is a mistake rather than an
// intention, and a service that started quietly with a half-configured sync
// would be indistinguishable from one deliberately running without one.
//
// The two limits are required with the sync because the sync is, so far, the
// only caller of the API: a sync that ran without them would have nothing to
// hold its requests to.
func LoadZoneSync(lookup LookupFunc) (*ZoneSync, error) {
	required := ZoneSyncRequired()

	var set, unset []string

	for _, name := range required {
		if v, ok := lookup(name); ok && v != "" {
			set = append(set, name)
		} else {
			unset = append(unset, name)
		}
	}

	switch len(set) {
	case 0:
		return nil, nil
	case len(required):
	default:
		return nil, fmt.Errorf("the zone sync is partially configured: %v set, %v missing — set all of them to run it, or none of them to run without it", set, unset)
	}

	cfg := &ZoneSync{
		FetchTimeout:     defaultFetchTimeout,
		DatabaseTimeout:  defaultDatabaseTimeout,
		MergeTimeout:     defaultMergeTimeout,
		MaxResponseBytes: defaultMaxResponseBytes,
		MinZoneRatio:     defaultMinZoneRatio,
	}

	cfg.DatabaseURL, _ = lookup(EnvDatabaseURL)

	allZones, _ := lookup(EnvAllZonesURL)
	if err := validateEndpoint(allZones); err != nil {
		return nil, fmt.Errorf("%s: %w", EnvAllZonesURL, err)
	}

	cfg.AllZonesURL = allZones

	var err error

	if cfg.Interval, err = positiveDuration(lookup, EnvInterval, 0); err != nil {
		return nil, err
	}

	if cfg.FetchTimeout, err = positiveDuration(lookup, EnvFetchTimeout, cfg.FetchTimeout); err != nil {
		return nil, err
	}

	if cfg.DatabaseTimeout, err = positiveDuration(lookup, EnvDatabaseTimeout, cfg.DatabaseTimeout); err != nil {
		return nil, err
	}

	if cfg.MergeTimeout, err = positiveDuration(lookup, EnvMergeTimeout, cfg.MergeTimeout); err != nil {
		return nil, err
	}

	if cfg.MaxResponseBytes, err = positiveBytes(lookup, EnvMaxResponseBytes, cfg.MaxResponseBytes, maxResponseBytesCeiling); err != nil {
		return nil, err
	}

	if cfg.MinZoneRatio, err = ratio(lookup, EnvMinZoneRatio, cfg.MinZoneRatio); err != nil {
		return nil, err
	}

	if cfg.RequestInterval, err = positiveDuration(lookup, EnvRequestInterval, 0); err != nil {
		return nil, err
	}

	if cfg.AllZonesInterval, err = positiveDuration(lookup, EnvAllZonesInterval, 0); err != nil {
		return nil, err
	}

	// REFUSED RATHER THAN CLAMPED, as every figure in this file is. `NFR-001`'s
	// Risk is that a refused all-zones request costs the next window rather than
	// a moment, and ShortestInterval says why the bound is the sum.
	if shortest := ShortestInterval(cfg.AllZonesInterval, cfg.FetchTimeout); cfg.Interval < shortest {
		return nil, fmt.Errorf("%s=%s is shorter than %s=%s plus %s=%s: across a restart the sync could ask the all-zones endpoint sooner than it permits after the previous request — lengthen the first, or shorten the fetch deadline",
			EnvInterval, cfg.Interval, EnvAllZonesInterval, cfg.AllZonesInterval, EnvFetchTimeout, cfg.FetchTimeout)
	}

	return cfg, nil
}

// ShortestInterval is the shortest sync schedule LoadZoneSync accepts: the
// all-zones interval, plus the deadline one all-zones fetch is bounded by.
//
// THE SUM, BECAUSE THE TWO HOLDS ON THAT ENDPOINT COUNT FROM DIFFERENT INSTANTS.
// Across a restart only the schedule holds, and it counts from the recorded
// start of the last run (`sync_run.started_at`), which is all a new process can
// read. `internal/turf`'s gate holds one process, and it counts from when the
// previous request RETURNED. A schedule of the interval alone satisfies each in
// its own terms and not both: every tick the gate holds a request back by the
// previous one's time to headers, so the request trails its recorded start by
// a little more each run, and a new process that waits the interval from the
// last start asks that much early.
//
// The fetch deadline closes the gap, because it bounds the request whatever the
// gate did: it is set as soon as the store has answered for the run's start,
// and before the gate is taken, so a request has returned within the deadline
// of its run's start. The next run's request, too, leaves only once the store
// has answered for its own start, so that answer is on both sides. A schedule
// at least the deadline longer than the interval therefore starts every run —
// in this process or in the next, both counted from the same recorded start —
// the interval after the previous request returned. The gate then never holds
// a scheduled request back, no request drifts behind its start, and nothing
// remains for a restart to miss.
//
// The send instant would close it too, and is not recorded anywhere a new
// process can read: `sync_run` dates a run by its start, and a column for the
// send is a migration.
func ShortestInterval(allZonesInterval, fetchTimeout time.Duration) time.Duration {
	return allZonesInterval + fetchTimeout
}

// ZoneSyncRequired names every variable the sync needs set and has no default
// for. Setting all of them runs the sync, and setting none runs the service
// without it.
func ZoneSyncRequired() []string {
	return []string{EnvDatabaseURL, EnvAllZonesURL, EnvInterval, EnvRequestInterval, EnvAllZonesInterval}
}

// validateEndpoint rejects an endpoint the fetcher could not use, at load time
// rather than on the first tick. A sync that starts and then fails once per
// interval against an unparseable URL reports a configuration error in the
// slowest way available to it.
//
// HTTPS IS REQUIRED, AND http IS REFUSED RATHER THAN ACCEPTED ALONGSIDE IT. The
// two were accepted equally, which made the transport a matter of how the
// variable happened to be typed. What arrives over it is the corpus that is
// merged into `zone`, and `zone` is the authoritative geometry every later
// query resolves against — so an unencrypted fetch is a body any observer on
// the path may rewrite before it is stored, and nothing downstream would
// notice: the staging assertions of `Architecture.md § The sync write path`
// check the staged row count and the coordinate ranges, which a substituted
// corpus with merely plausible coordinates satisfies. Refusing the scheme here
// is what keeps that from being a configuration typo's decision to make.
func validateEndpoint(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("not a URL: %w", err)
	}

	if u.Scheme != "https" {
		return fmt.Errorf("scheme is %q, want https: the zone corpus is merged into the authoritative store, so it is not fetched over a transport that permits it to be rewritten in flight", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("%q names no host", raw)
	}

	return nil
}

// positiveDuration reads a duration. A fallback of zero marks the variable
// required, which is how the interval is kept out of this package.
func positiveDuration(lookup LookupFunc, name string, fallback time.Duration) (time.Duration, error) {
	raw, ok := lookup(name)
	if !ok || raw == "" {
		if fallback <= 0 {
			return 0, fmt.Errorf("%s is required and was not set", name)
		}

		return fallback, nil
	}

	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s=%q is not a duration: %w", name, raw, err)
	}

	if d <= 0 {
		return 0, fmt.Errorf("%s=%q is not positive", name, raw)
	}

	return d, nil
}

// positiveBytes reads a byte count, and holds it inside (0, ceiling]. Both ends
// are refusals rather than clamps, for the reason every tunable in this file is
// refused rather than defaulted: a figure silently replaced is a figure the
// operator believes they set.
func positiveBytes(lookup LookupFunc, name string, fallback, ceiling int64) (int64, error) {
	raw, ok := lookup(name)
	if !ok || raw == "" {
		return fallback, nil
	}

	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s=%q is not a byte count: %w", name, raw, err)
	}

	if n <= 0 {
		return 0, fmt.Errorf("%s=%q is not positive", name, raw)
	}

	if n > ceiling {
		return 0, fmt.Errorf("%s=%q is above the %d bytes this service will read one response into", name, raw, ceiling)
	}

	return n, nil
}

func ratio(lookup LookupFunc, name string, fallback float64) (float64, error) {
	raw, ok := lookup(name)
	if !ok || raw == "" {
		return fallback, nil
	}

	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s=%q is not a number: %w", name, raw, err)
	}

	// Written as the range it accepts, not as the two it refuses: ParseFloat
	// reads "NaN" as a number, and NaN compares false against both bounds.
	if !(f > 0 && f <= 1) {
		return 0, fmt.Errorf("%s=%q is outside (0, 1]", name, raw)
	}

	return f, nil
}
