package main

import (
	"context"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
	"github.com/Caisesiume/TurfGPS/service/internal/syncstore"
	"github.com/Caisesiume/TurfGPS/service/internal/turf"
	"github.com/Caisesiume/TurfGPS/service/internal/zonestore"
	"github.com/Caisesiume/TurfGPS/service/internal/zonesync"
)

// startZoneFeed starts the zone feed's background job (#232, `FR-119`) and
// returns the function that waits for it to stop, reporting — as
// startZoneSync's does — whether it ended on its own.
//
// IT IS WIRED BESIDE THE SYNC AND NOT INSIDE IT. The feed has its own
// configuration block and its own schedule, and either job runs without the
// other: a service with no feed configured starts and serves, for `NFR-003`'s
// reason given on startZoneSync. It opens a pool of its own rather than
// sharing the sync's, so neither job's wiring depends on the other's being
// configured; the two meet only in the database, under the sync's advisory key
// (`internal/syncstore`'s ApplyFeed says how).
//
// Its stop is waited on alongside the sync's, inside the same budget, because
// main waits for the two at once: see main.
//
// feed is loadZoneFeed's, read before anything started; nil is no feed
// configured.
func startZoneFeed(ctx context.Context, stop context.CancelFunc, feed *zoneFeed) (func() bool, error) {
	if feed == nil {
		slog.Info("no zone feed is configured, so zones created since the last full sync will not reach the local copy until the next one",
			"configure", strings.Join(config.ZoneFeedRequired(), ", "))

		return func() bool { return false }, nil
	}

	pool, err := zonestore.Open(ctx, feed.cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	poller, err := newZoneFeedPoller(feed, pool)
	if err != nil {
		pool.Close()

		return nil, err
	}

	slog.Info("the zone feed is running", "interval", feed.cfg.Interval)

	return superviseBackgroundJob(ctx, stop, "zone feed", poller.Run, pool.Close), nil
}

// zoneFeed is the zone feed's configuration, read and checked, with the client
// it configures already built.
type zoneFeed struct {
	cfg    *config.ZoneFeed
	client *turf.FeedClient
}

// loadZoneFeed reads the feed's configuration and builds its client, and so
// refuses every configuration either refuses — an endpoint off the API's
// current version among them — without opening anything. main calls it before
// the listener and the sync start, so a feed configured wrongly stops the
// process before there is anything to shut down (#238, N2). Nil, nil is no feed
// configured.
func loadZoneFeed(lookup config.LookupFunc) (*zoneFeed, error) {
	cfg, err := config.LoadZoneFeed(lookup)
	if err != nil || cfg == nil {
		return nil, err
	}

	client, err := turf.NewFeedClient(cfg.FeedURL, cfg.MaxResponseBytes, turf.Limits{Request: cfg.RequestInterval})
	if err != nil {
		return nil, err
	}

	return &zoneFeed{cfg: cfg, client: client}, nil
}

func newZoneFeedPoller(feed *zoneFeed, pool *pgxpool.Pool) (*zonesync.FeedPoller, error) {
	store, err := syncstore.New(pool, slog.Default())
	if err != nil {
		return nil, err
	}

	cfg := feed.cfg

	return zonesync.NewFeedPoller(zonesync.FeedConfig{
		Fetch:           feed.client.FetchZoneFeed,
		Store:           store,
		Logger:          slog.Default().With("component", "zone_feed"),
		Interval:        cfg.Interval,
		FetchTimeout:    cfg.FetchTimeout,
		DatabaseTimeout: cfg.DatabaseTimeout,
	})
}

// superviseBackgroundJob runs a background loop the process owns, and returns
// the function main waits on. It is startZoneSync's supervision — the
// arguments for each step are there, beside the job they were written for —
// held as one function so the feed is supervised the same way:
//
//   - a loop that panics or returns cancels the root context, so the process
//     drains what is in flight and exits nonzero rather than serving on
//     without the job;
//   - the waiter reports whether the job ended on its own, and false when it
//     did not stop within syncStopTimeout, leaving without closing what close
//     releases, for the reason startZoneSync's waiter gives.
func superviseBackgroundJob(ctx context.Context, stop context.CancelFunc, name string, run func(context.Context) error, closeFn func()) func() bool {
	stopped := make(chan struct{})

	// Written by the goroutine and read by the waiter only after the close of
	// stopped, which orders the two; see startZoneSync's `died`.
	var died bool

	go func() {
		defer close(stopped)

		defer func() {
			if recovered := recover(); recovered != nil {
				died = true

				slog.Error("a background job panicked, so this process is stopping rather than serving on without it",
					"job", name, "panic", recovered, "stack", string(debug.Stack()))

				stop()
			}
		}()

		if err := run(ctx); err != nil {
			died = true

			slog.Error("a background job stopped, so this process is stopping rather than serving on without it", "job", name, "error", err)

			stop()
		}
	}()

	return func() bool {
		select {
		case <-stopped:
		case <-time.After(syncStopTimeout):
			slog.Error("a background job did not stop within its budget, so this process is leaving without it",
				"job", name, "budget", syncStopTimeout)

			return false
		}

		closeFn()

		return died
	}
}
