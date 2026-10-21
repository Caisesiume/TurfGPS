package config

import (
	"fmt"
	"time"
)

// The zone feed's configuration (#232), a block of its own beside the zone
// sync's: the feed runs on its own schedule, and either can be configured
// without the other.
//
// THE INTERVAL HAS NO DEFAULT, for the reason the refresh interval has none.
// Its home is the Owner's zone-sync decision on #20
// (`#issuecomment-6279336314`, item 2) until `Architecture.md § Retrieving
// zones` is amended to carry it, and a figure compiled in here would be a
// second home for it. Nor has the endpoint, which is named by role there and
// reaches the service as a resolved URL; `internal/turf` refuses one that does
// not address APIVersion.
//
// The database and the API's general limit are the sync's variables, read
// again here rather than duplicated: they name one store and one API whichever
// job reads them.
const (
	EnvZoneFeedURL              = "TURFGPS_ZONE_FEED_URL"
	EnvZoneFeedInterval         = "TURFGPS_ZONE_FEED_INTERVAL"
	EnvZoneFeedFetchTimeout     = "TURFGPS_ZONE_FEED_FETCH_TIMEOUT"
	EnvZoneFeedDatabaseTimeout  = "TURFGPS_ZONE_FEED_DB_TIMEOUT"
	EnvZoneFeedMaxResponseBytes = "TURFGPS_ZONE_FEED_MAX_RESPONSE_BYTES"
)

// Defaults for the feed's own operational tunables. None restates a fact
// recorded elsewhere, and none is the interval.
const (
	// defaultFeedFetchTimeout bounds one feed request end to end. It makes the
	// call finite; a feed response is small (99 items in about 38 kB, live on
	// 20 October 2026), so a minute is generous rather than tight.
	defaultFeedFetchTimeout = time.Minute

	// defaultFeedDatabaseTimeout bounds the latest-time read and the one write
	// transaction a poll makes.
	defaultFeedDatabaseTimeout = time.Minute

	// defaultFeedMaxResponseBytes caps one feed response in memory. The API
	// states no item count — "from 0 to x" — so this is a guard against a
	// runaway body and not a prediction: 16 MiB is some four hundred times the
	// response observed. A feed that outgrows it is refused, logged, and
	// retried on the next poll until the ceiling is raised.
	defaultFeedMaxResponseBytes int64 = 16 << 20
)

// ZoneFeed is the configuration the zone-feed poller needs.
type ZoneFeed struct {
	// DatabaseURL addresses the store of `Architecture.md § D4` — the sync's.
	DatabaseURL string

	// FeedURL is the resolved zone-feed endpoint.
	FeedURL string

	// Interval is how long passes between one poll and the next. Its home is
	// the Owner's decision cited on this file.
	Interval time.Duration

	FetchTimeout     time.Duration
	DatabaseTimeout  time.Duration
	MaxResponseBytes int64

	// RequestInterval is the Turf API's general limit, the sync's
	// EnvRequestInterval. Its home is `Architecture.md § Data sources and
	// constraints`.
	RequestInterval time.Duration
}

// LoadZoneFeed reads the zone feed's configuration.
//
// It returns (nil, nil) when neither of the feed's own variables — its
// endpoint and its interval — is set, which is a service running without the
// feed, as `NFR-003` permits. Once either is set, every variable the feed has
// no default for is required, the two it shares with the sync among them, and
// a missing one is an error rather than a feed quietly not running.
func LoadZoneFeed(lookup LookupFunc) (*ZoneFeed, error) {
	triggered := false

	for _, name := range []string{EnvZoneFeedURL, EnvZoneFeedInterval} {
		if v, ok := lookup(name); ok && v != "" {
			triggered = true
		}
	}

	if !triggered {
		return nil, nil
	}

	var unset []string

	for _, name := range ZoneFeedRequired() {
		if v, ok := lookup(name); !ok || v == "" {
			unset = append(unset, name)
		}
	}

	if len(unset) > 0 {
		return nil, fmt.Errorf("the zone feed is partially configured: %v missing — set all of %v to run it, or neither %s nor %s to run without it",
			unset, ZoneFeedRequired(), EnvZoneFeedURL, EnvZoneFeedInterval)
	}

	cfg := &ZoneFeed{
		FetchTimeout:     defaultFeedFetchTimeout,
		DatabaseTimeout:  defaultFeedDatabaseTimeout,
		MaxResponseBytes: defaultFeedMaxResponseBytes,
	}

	cfg.DatabaseURL, _ = lookup(EnvDatabaseURL)

	feed, _ := lookup(EnvZoneFeedURL)
	if err := validateEndpoint(feed); err != nil {
		return nil, fmt.Errorf("%s: %w", EnvZoneFeedURL, err)
	}

	cfg.FeedURL = feed

	var err error

	if cfg.Interval, err = positiveDuration(lookup, EnvZoneFeedInterval, 0); err != nil {
		return nil, err
	}

	if cfg.FetchTimeout, err = positiveDuration(lookup, EnvZoneFeedFetchTimeout, cfg.FetchTimeout); err != nil {
		return nil, err
	}

	if cfg.DatabaseTimeout, err = positiveDuration(lookup, EnvZoneFeedDatabaseTimeout, cfg.DatabaseTimeout); err != nil {
		return nil, err
	}

	if cfg.MaxResponseBytes, err = positiveBytes(lookup, EnvZoneFeedMaxResponseBytes, cfg.MaxResponseBytes, maxResponseBytesCeiling); err != nil {
		return nil, err
	}

	if cfg.RequestInterval, err = positiveDuration(lookup, EnvRequestInterval, 0); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ZoneFeedRequired names every variable the feed needs set and has no default
// for: its own endpoint and interval, and the database and general limit it
// shares with the sync.
func ZoneFeedRequired() []string {
	return []string{EnvDatabaseURL, EnvZoneFeedURL, EnvZoneFeedInterval, EnvRequestInterval}
}
