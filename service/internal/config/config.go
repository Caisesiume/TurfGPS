// Package config reads the service's runtime configuration from the process
// environment.
//
// The environment is the source rather than a file in this repository because
// `DEPLOYMENT.md § Where the deployment configuration lives` puts the host's
// arrangement outside the repository and forbids secret material inside it: the
// service's unit references an out-of-repo source, and this package reads what
// that source sets.
//
// TWO VALUES DELIBERATELY HAVE NO DEFAULT HERE, and that is the point of the
// package rather than an omission.
//
//   - The refresh interval. Its one home is `Architecture.md § Retrieving
//     zones`, and a default compiled into this package would be a second home
//     for it — one that goes wrong silently when that section moves.
//   - The all-zones endpoint. It is named by role in that same section and
//     reaches the service as a resolved URL, so no versioned API path appears
//     anywhere in this module. Binding the version by citation is `FR-019`'s
//     work, not this story's.
//
// Both therefore arrive from the environment or the sync does not run at all.
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
	// is `Architecture.md § Retrieving zones`.
	Interval time.Duration

	FetchTimeout     time.Duration
	DatabaseTimeout  time.Duration
	MergeTimeout     time.Duration
	MaxResponseBytes int64
	MinZoneRatio     float64
}

// LookupFunc reads one environment variable and reports whether it was set at
// all. It has the shape of os.LookupEnv so a test can supply its own.
type LookupFunc func(name string) (string, bool)

// LoadZoneSync reads the zone sync's configuration.
//
// It returns (nil, nil) when NONE of the three required variables is set, which
// is how a service with no sync configured still starts and serves — the
// property `NFR-003` binds and this story must not spend. It returns an error
// when SOME of them are set: partial configuration is a mistake rather than an
// intention, and a service that started quietly with a half-configured sync
// would be indistinguishable from one deliberately running without one.
func LoadZoneSync(lookup LookupFunc) (*ZoneSync, error) {
	required := []string{EnvDatabaseURL, EnvAllZonesURL, EnvInterval}

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

	if cfg.MaxResponseBytes, err = positiveBytes(lookup, EnvMaxResponseBytes, cfg.MaxResponseBytes); err != nil {
		return nil, err
	}

	if cfg.MinZoneRatio, err = ratio(lookup, EnvMinZoneRatio, cfg.MinZoneRatio); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validateEndpoint rejects an endpoint the fetcher could not use, at load time
// rather than on the first tick. A sync that starts and then fails once per
// interval against an unparseable URL reports a configuration error in the
// slowest way available to it.
func validateEndpoint(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("not a URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("scheme is %q, want http or https", u.Scheme)
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

func positiveBytes(lookup LookupFunc, name string, fallback int64) (int64, error) {
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

	if f <= 0 || f > 1 {
		return 0, fmt.Errorf("%s=%q is outside (0, 1]", name, raw)
	}

	return f, nil
}
