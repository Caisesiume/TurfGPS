package main

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
	"github.com/Caisesiume/TurfGPS/service/internal/turf"
)

// Tests over what loadZoneFeed and startZoneFeed tell main: whether the feed ENDED ON ITS OWN,
// which main maps to the exit status, as `zonesyncstop_test.go` tests for the
// sync and for the reasons its header gives in both directions. As there, no
// database is reached: the configured case cancels its context before the
// feed starts, so FeedPoller.Run returns at the check that opens its loop,
// and the DSN names port 1 in case that ever changes.
//
// superviseBackgroundJob takes the loop it runs, so the `true` side the sync's
// file cannot reach — a job that really died — is reachable here.

func noFeedConfigured(t *testing.T) {
	t.Helper()

	for _, key := range []string{config.EnvZoneFeedURL, config.EnvZoneFeedInterval} {
		t.Setenv(key, "")
	}
}

// loadedFeed is loadZoneFeed over the process's environment, as main reads it.
func loadedFeed(t *testing.T) *zoneFeed {
	t.Helper()

	feed, err := loadZoneFeed(os.LookupEnv)
	if err != nil {
		t.Fatalf("the zone feed's configuration was refused: %v", err)
	}

	return feed
}

func TestAnUnconfiguredFeedIsNotReportedAsHavingDied(t *testing.T) {
	noFeedConfigured(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	stop := &countedStop{cancel: cancel}

	wait, err := startZoneFeed(ctx, stop.stop, loadedFeed(t))
	if err != nil {
		t.Fatalf("no zone feed is configured and startZoneFeed refused to start: %v — main maps that to exit 1, so a service with no feed would not start", err)
	}

	if died := wait(); died {
		t.Error("no zone feed was configured and the waiter reports that one died, want false — main would exit 1 on every clean shutdown")
	}

	if n := stop.calls.Load(); n != 0 {
		t.Errorf("no zone feed was configured and the root context was cancelled %d time(s), want 0", n)
	}
}

func TestAStoppedFeedIsNotReportedAsHavingDied(t *testing.T) {
	t.Setenv(config.EnvDatabaseURL, "postgres://turfgps@127.0.0.1:1/turfgps")
	t.Setenv(config.EnvZoneFeedURL, "https://example.test/"+turf.APIVersion+"/feeds/zone")
	t.Setenv(config.EnvZoneFeedInterval, "1h")
	t.Setenv(config.EnvRequestInterval, "250ms")

	ctx, cancel := context.WithCancel(t.Context())
	stop := &countedStop{cancel: cancel}

	cancel()

	wait, err := startZoneFeed(ctx, stop.stop, loadedFeed(t))
	if err != nil {
		t.Fatalf("a configured feed refused to start: %v", err)
	}

	if died := wait(); died {
		t.Error("a feed stopped by its context reports that it died, want false")
	}

	if n := stop.calls.Load(); n != 0 {
		t.Errorf("a feed stopped by its context cancelled the root context %d time(s), want 0", n)
	}
}

func TestAPartlyConfiguredFeedIsRefused(t *testing.T) {
	noFeedConfigured(t)
	t.Setenv(config.EnvZoneFeedInterval, "1h")

	if _, err := loadZoneFeed(os.LookupEnv); err == nil {
		t.Error("a feed with an interval and no endpoint was accepted, want a refusal main exits on")
	}
}

// TestAFeedEndpointOffTheCurrentVersionIsRefusedBeforeAnythingStarts (#238,
// N2): the feed's client refuses an endpoint not on the API's current version,
// and loadZoneFeed — which main calls before the listener and the sync start,
// and which opens nothing — is where that refusal comes from, so main exits on
// it with nothing running to stop.
func TestAFeedEndpointOffTheCurrentVersionIsRefusedBeforeAnythingStarts(t *testing.T) {
	t.Setenv(config.EnvDatabaseURL, "postgres://turfgps@127.0.0.1:1/turfgps")
	t.Setenv(config.EnvZoneFeedURL, "https://example.test/v4/feeds/zone")
	t.Setenv(config.EnvZoneFeedInterval, "1h")
	t.Setenv(config.EnvRequestInterval, "250ms")

	if feed, err := loadZoneFeed(os.LookupEnv); err == nil {
		t.Errorf("a feed endpoint on /v4 was accepted by loadZoneFeed (%+v), want the refusal main exits on before the sync starts", feed)
	}
}

// TestABackgroundJobThatEndsOnItsOwnTakesTheProcessDown: a loop that returns
// an error or panics cancels the root context and is reported as having died,
// so main exits nonzero and a supervisor replaces the process.
func TestABackgroundJobThatEndsOnItsOwnTakesTheProcessDown(t *testing.T) {
	for name, run := range map[string]func(context.Context) error{
		"returns":  func(context.Context) error { return errors.New("gone") },
		"panics":   func(context.Context) error { panic("gone") },
		"is asked": func(ctx context.Context) error { <-ctx.Done(); return nil },
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			stop := &countedStop{cancel: cancel}
			closed := false

			wait := superviseBackgroundJob(ctx, stop.stop, "test job", run, func() { closed = true })

			if name == "is asked" {
				cancel()
			}

			died := wait()

			if want := name != "is asked"; died != want {
				t.Errorf("a job that %s was reported died=%v, want %v", name, died, want)
			}

			if want := name != "is asked"; (stop.calls.Load() > 0) != want {
				t.Errorf("a job that %s cancelled the root context %d time(s)", name, stop.calls.Load())
			}

			if !closed {
				t.Errorf("a job that %s and stopped was not closed after it", name)
			}
		})
	}
}
