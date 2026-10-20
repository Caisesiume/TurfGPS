package config

import (
	"strings"
	"testing"
	"time"
)

// feedComplete is the smallest configuration that runs the zone feed. The
// interval is this test's own figure and stands for no schedule: the feed's
// has its home in the Owner's decision on #20 and reaches the service from the
// environment only.
func feedComplete() map[string]string {
	return map[string]string{
		EnvDatabaseURL:      "postgres://localhost/turfgps",
		EnvZoneFeedURL:      "https://example.test/feed",
		EnvZoneFeedInterval: "13m",
		EnvRequestInterval:  "250ms",
	}
}

// TestNoFeedConfiguredIsNotAnError: neither of the feed's own variables set is
// a service running without the feed, whether or not the sync is configured.
func TestNoFeedConfiguredIsNotAnError(t *testing.T) {
	t.Parallel()

	for name, vars := range map[string]map[string]string{"nothing": {}, "only the sync": complete()} {
		cfg, err := LoadZoneFeed(env(vars))
		if err != nil || cfg != nil {
			t.Errorf("with %s configured, LoadZoneFeed gave %+v, %v; want no feed and no error", name, cfg, err)
		}
	}
}

// TestTheFeedIntervalIsCarriedAndNeverDefaulted: the interval reaches the
// configuration from the environment and from nowhere else — a feed endpoint
// set without it is refused, not filled in.
func TestTheFeedIntervalIsCarriedAndNeverDefaulted(t *testing.T) {
	t.Parallel()

	cfg, err := LoadZoneFeed(env(feedComplete()))
	if err != nil {
		t.Fatalf("a complete feed configuration was refused: %v", err)
	}

	want := ZoneFeed{
		DatabaseURL:      "postgres://localhost/turfgps",
		FeedURL:          "https://example.test/feed",
		Interval:         13 * time.Minute,
		FetchTimeout:     defaultFeedFetchTimeout,
		DatabaseTimeout:  defaultFeedDatabaseTimeout,
		MaxResponseBytes: defaultFeedMaxResponseBytes,
		RequestInterval:  250 * time.Millisecond,
	}

	if *cfg != want {
		t.Errorf("the feed configuration is %+v, want %+v", *cfg, want)
	}

	for _, missing := range ZoneFeedRequired() {
		vars := feedComplete()
		delete(vars, missing)

		if _, err := LoadZoneFeed(env(vars)); err == nil || !strings.Contains(err.Error(), missing) {
			t.Errorf("a feed configuration without %s gave %v, want a refusal naming it", missing, err)
		}
	}
}

func TestAMalformedFeedVariableIsRefused(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]string{
		EnvZoneFeedURL:              "http://example.test/feed",
		EnvZoneFeedInterval:         "0s",
		EnvZoneFeedFetchTimeout:     "soon",
		EnvZoneFeedDatabaseTimeout:  "-1m",
		EnvZoneFeedMaxResponseBytes: "1099511627776",
	} {
		vars := feedComplete()
		vars[name] = value

		if _, err := LoadZoneFeed(env(vars)); err == nil || !strings.Contains(err.Error(), name) {
			t.Errorf("%s=%q gave %v, want a refusal naming the variable", name, value, err)
		}
	}
}

func TestTheFeedTunablesAreOverridable(t *testing.T) {
	t.Parallel()

	vars := feedComplete()
	vars[EnvZoneFeedFetchTimeout] = "17s"
	vars[EnvZoneFeedDatabaseTimeout] = "19s"
	vars[EnvZoneFeedMaxResponseBytes] = "4096"

	cfg, err := LoadZoneFeed(env(vars))
	if err != nil {
		t.Fatalf("refused: %v", err)
	}

	if cfg.FetchTimeout != 17*time.Second || cfg.DatabaseTimeout != 19*time.Second || cfg.MaxResponseBytes != 4096 {
		t.Errorf("the overrides gave %+v", *cfg)
	}
}
