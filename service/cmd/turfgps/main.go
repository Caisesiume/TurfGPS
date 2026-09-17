// Command turfgps runs the TurfGPS service as one long-lived server process,
// built as a single self-contained executable per NFR-003.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
	"github.com/Caisesiume/TurfGPS/service/internal/httpapi"
	"github.com/Caisesiume/TurfGPS/service/internal/syncstore"
	"github.com/Caisesiume/TurfGPS/service/internal/turf"
	"github.com/Caisesiume/TurfGPS/service/internal/zonestore"
	"github.com/Caisesiume/TurfGPS/service/internal/zonesync"
)

const (
	// defaultAddr is where the server listens. All interfaces, because the
	// process is expected to be the only thing in its container.
	defaultAddr = ":8080"

	// shutdownTimeout bounds the drain of in-flight requests once the process
	// has been asked to stop. It stays well inside a container runtime's usual
	// grace period, which kills the process outright when that period expires —
	// and that grace, not writeTimeout below, is what bounds it. The two answer
	// different questions: writeTimeout is how long one request may take, this
	// is how long the process may take to disappear before its supervisor stops
	// waiting, so neither derives from the other. A drain matching writeTimeout
	// would promise time no runtime lets it serve: it outlives even the
	// stopGrace the image harness signals with, so the runtime would kill the
	// process mid-drain — the same request lost, and lost at the runtime's
	// discretion rather than through the service's own error path.
	//
	// The cost is deliberate and it falls on the handler: a request still
	// running when this expires is severed. A solve streamed over tens of
	// seconds, per `Architecture.md § Response time and progressive results`,
	// sits inside writeTimeout and far outside this budget, so what carries one
	// across a restart is a client that can resume — raising this alone only
	// moves the loss to the runtime's kill.
	shutdownTimeout = 5 * time.Second

	// syncStopTimeout bounds the wait for the background sync to stop, for the
	// same reason shutdownTimeout above bounds the HTTP drain — and it is a
	// second budget rather than a share of that one because the two are waited
	// on in sequence, so what the runtime allows the process to disappear in is
	// spent by their sum and by neither alone.
	//
	// IT CANNOT BE LONG ENOUGH TO LET A RUN FINISH, and it does not try. The
	// sync's detached writes are budgeted in minutes — the terminal `sync_run`
	// update and the lock release each take the database timeout — so a wait
	// that saw them out is a wait no grace period outlives, and the process
	// would be killed mid-shutdown on every rolling deploy rather than choosing
	// where it stopped.
	//
	// The cost is deliberate and it falls on the run record rather than on a
	// request: a run still in flight when this expires is left at `running`,
	// which `Architecture.md § The sync write path` already gives a meaning to,
	// and the next attempt waits out the interval on it rather than fetching. A
	// loop between runs — the ordinary case, every minute of the interval bar
	// the seconds a run takes — stops at once and never spends this at all.
	syncStopTimeout = 3 * time.Second

	// readHeaderTimeout bounds how long a client may take to send its request
	// headers, so an idle connection cannot hold a server goroutine open.
	readHeaderTimeout = 10 * time.Second

	// readTimeout bounds the request as a whole, headers and body together.
	// readHeaderTimeout above stops applying once the headers are in, so
	// without this a client may send a well-formed header block and then
	// trickle a body for as long as it likes.
	readTimeout = 30 * time.Second

	// writeTimeout bounds the response. Go arms the write deadline as soon as
	// the request headers have been read, so this budget covers reading the
	// body and running the handler as well as writing the reply — which is why
	// it sits above readTimeout rather than beside it. A handler that
	// legitimately needs longer extends its own deadline through
	// http.ResponseController; raising this ceiling for one slow route would
	// relax it for every route. It is not a survival guarantee either:
	// shutdownTimeout above is far shorter and answers to the runtime rather
	// than to this budget, so a request still running when the process is asked
	// to stop is severed well before this bound.
	writeTimeout = 60 * time.Second

	// idleTimeout bounds how long a kept-alive connection may wait between
	// requests. Go falls back on readTimeout when this is unset, which is the
	// wrong budget for it: a connection waiting to be reused is not a slow
	// client, and closing healthy connections every 30 seconds would cost a
	// handshake per client per half-minute for nothing.
	idleTimeout = 120 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ln, err := net.Listen("tcp", defaultAddr)
	if err != nil {
		slog.Error("cannot listen", "addr", defaultAddr, "error", err)
		os.Exit(1)
	}

	stopSync, err := startZoneSync(ctx, stop)
	if err != nil {
		slog.Error("the zone sync is configured and could not start", "error", err)
		os.Exit(1)
	}

	serveErr := serve(ctx, ln)

	// Cancelled here, explicitly, rather than left to the deferred stop above.
	// serve returns for two reasons and only one of them has already cancelled
	// ctx: a listener that fails returns an error with the context still live,
	// the sync loop watching that context never finishes, and the wait below
	// blocks for good — so the exit this function ends with is unreachable and
	// the process survives as a dead listener beside a live sync loop. This
	// covers both paths; on the signal path ctx is already done and it is a
	// no-op.
	stop()

	// Called explicitly rather than deferred: the exit below skips defers, and
	// the whole purpose of this call is to let the sync's own shutdown finish.
	syncDied := stopSync()

	if serveErr != nil {
		slog.Error("service stopped", "error", serveErr)
		os.Exit(1)
	}

	// A sync that ended on its own took this process down with it, and the exit
	// status is the only part of that a supervisor reads. Zero here would report
	// the shutdown that just happened as the ordinary one — the drain completed,
	// so every field of it looks ordinary — and a runtime restarting only on
	// failure would leave the image running with no sync in it and a route that
	// answers 200 regardless. The log line above says what happened; this is what
	// makes something act on it.
	if syncDied {
		slog.Error("this process is exiting nonzero because its zone sync is gone, so that a supervisor replaces it")
		os.Exit(1)
	}
}

// startZoneSync starts the background job of `FR-022` and returns the function
// that waits for it to stop. That function reports whether the sync ended on its
// own rather than because it was asked to, which is the difference between a
// shutdown and a casualty and is not otherwise visible from outside.
//
// THIS IS THE COMPOSITION ROOT AND THE ONLY PLACE THE SYNC IS WIRED. The ports
// it satisfies are declared by `internal/zonesync`, and the adapters that
// satisfy them do not know each other; joining them is this function's whole
// job, which is why it is the one package the off-request-path invariant permits
// to import the worker.
//
// A SERVICE WITH NO SYNC CONFIGURED STARTS AND SERVES ANYWAY. `NFR-003` measures
// start-up and a first served request in an image carrying nothing else, and a
// service that refused to start without a database and an endpoint would spend
// that property to gain a check the sync's own logging already makes.
//
// Nor does it wait for the store to answer. The pool connects lazily and this
// function does not force it: a database that is down delays the copy's next
// refresh, and `FR-022` AC2 is precisely the requirement that no such failure
// reaches a request.
func startZoneSync(ctx context.Context, stop context.CancelFunc) (func() bool, error) {
	cfg, err := config.LoadZoneSync(os.LookupEnv)
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		slog.Info("no zone sync is configured, so the local zone copy will not be refreshed",
			"configure", config.EnvDatabaseURL+", "+config.EnvAllZonesURL+" and "+config.EnvInterval)

		// False rather than true: a sync that was never started has not died,
		// and this is the configuration `NFR-003` measures the image in.
		return func() bool { return false }, nil
	}

	pool, err := zonestore.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	scheduler, err := newZoneSyncScheduler(cfg, pool)
	if err != nil {
		pool.Close()

		return nil, err
	}

	stopped := make(chan struct{})

	// died is written by the goroutine below and read by the waiter returned at
	// the end, and the two are ordered by the close of stopped: every write to
	// it happens in a frame the deferred close is registered before, so the
	// close cannot run until they are done. The waiter reads it only on the
	// branch that saw that close — the timeout branch reads nothing because
	// without the close there is no edge to read it across, WHICH IS NOT THE
	// SAME AS THERE BEING NOTHING TO READ. A panicking run sets died in a
	// handler that still has a stack to log and a stop() to call before the
	// close can run, so "died, and not yet stopped" is a state that branch can
	// genuinely find; what it returns there is a statement about what it
	// observed rather than about what happened, which the branch itself says.
	var died bool

	go func() {
		defer close(stopped)

		// Registered after the close above so that it runs BEFORE it, and it is
		// here because this is the frame a panicking run was always going to
		// arrive in. runOnce recovers, records the run's row, and re-raises on
		// purpose — so that a panic is not swallowed by the store's bookkeeping
		// — and a bare goroutine is where a re-raise becomes the death of the
		// whole process. That death takes the HTTP server with it without ever
		// reaching srv.Shutdown: every in-flight request severed, by a failure
		// in the job whose own package doc says it is off the request path
		// INCLUDING WHEN IT FAILS.
		//
		// RECOVERING IS NOT THE SAME AS SURVIVING, AND THE FIRST VERSION OF THIS
		// TRADED ONE FAILURE FOR A QUIETER ONE. Keeping the process up after this
		// loop is gone keeps nothing that matters: nothing here restarts it, the
		// request surface is one static route that answers 200 whether the copy
		// is minutes or months stale, and the log line below would be the whole
		// of the signal. That is a service reporting healthy while the only
		// thing it was doing has stopped — indefinitely, and invisibly to
		// anything that polls it — where an ordinary crash would at least have
		// been replaced by a supervisor within seconds.
		//
		// So the root context is cancelled instead. serve returns, drains what is
		// in flight rather than severing it, and main exits nonzero so a
		// supervisor puts a live sync back. The panic still costs this process;
		// what it no longer costs is the in-flight requests, which is the whole
		// of what recovering here buys. The stack is worth capturing because a
		// panic raised outside runOnce has been logged nowhere else, and it is
		// still readable here: the runtime unwinds to this frame when this
		// deferred call RETURNS, not when recover returns, so the panicking
		// frames are on the stack for the whole of it.
		defer func() {
			if recovered := recover(); recovered != nil {
				died = true

				slog.Error("the zone sync panicked, so this process is stopping rather than serving on without it",
					"panic", recovered, "stack", string(debug.Stack()))

				stop()
			}
		}()

		// Run is declared to return an error and returns nil on every path it
		// has. The branch is kept because the signature is the port's and not
		// this file's to narrow, and it is treated as the panic above is rather
		// than justifying it: a loop that has ended is a loop that has ended,
		// however it says so.
		if err := scheduler.Run(ctx); err != nil {
			died = true

			slog.Error("the zone sync stopped, so this process is stopping rather than serving on without it", "error", err)

			stop()
		}
	}()

	slog.Info("the zone sync is running", "interval", cfg.Interval)

	return func() bool {
		select {
		case <-stopped:
		case <-time.After(syncStopTimeout):
			// Returns WITHOUT closing the pool, deliberately. Close blocks
			// until every connection is back, and the connection this sync
			// holds its session lock on is precisely the one that has not come
			// back — so closing here would restore the unbounded wait the
			// budget exists to remove. The process is leaving: its sockets go
			// with it, and PostgreSQL releases a session lock when the backend
			// holding it dies, which is why the lock is session-level and not a
			// lease.
			slog.Error("the zone sync did not stop within its budget, so this process is leaving without it",
				"budget", syncStopTimeout,
				"consequence", "a run in flight is left at running, and the next attempt waits out the interval on it rather than fetching")

			// False on this branch, and it is a statement about what was
			// observed rather than a guess: the sync has not stopped, so it
			// has not ended on its own, and this process is leaving for a
			// reason that was already decided before the wait began.
			return false
		}

		pool.Close()

		return died
	}, nil
}

func newZoneSyncScheduler(cfg *config.ZoneSync, pool *pgxpool.Pool) (*zonesync.Scheduler, error) {
	store, err := syncstore.New(pool, slog.Default())
	if err != nil {
		return nil, err
	}

	client, err := turf.NewClient(cfg.AllZonesURL, cfg.MaxResponseBytes)
	if err != nil {
		return nil, err
	}

	return zonesync.NewScheduler(zonesync.Config{
		Fetch:           client.FetchAllZones,
		Store:           store,
		Locker:          store,
		Logger:          slog.Default().With("component", "zone_sync"),
		Interval:        cfg.Interval,
		FetchTimeout:    cfg.FetchTimeout,
		DatabaseTimeout: cfg.DatabaseTimeout,
		MergeTimeout:    cfg.MergeTimeout,
		MinZoneRatio:    cfg.MinZoneRatio,
	})
}

// serve runs the HTTP server on ln until ctx is cancelled, then drains
// in-flight requests and returns. It takes ownership of ln and closes it.
func serve(ctx context.Context, ln net.Listener) error {
	// The request surface is a package of its own. `internal/httpapi` says why:
	// it is the seam `FR-022` AC2 is checked over, and a mux built inline here
	// would put every future handler inside the one package the invariant has
	// to permit to reach the sync worker.
	srv := &http.Server{
		Handler:           httpapi.NewMux(),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	served := make(chan error, 1)
	go func() { served <- srv.Serve(ln) }()

	slog.Info("listening", "addr", ln.Addr().String())

	select {
	case err := <-served:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	slog.Info("stopping", "drain", shutdownTimeout)

	drainCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(drainCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}
