// Package httpapi is the service's request surface: every HTTP handler the
// service serves is registered here or under here.
//
// IT IS ALSO THE SEAM `FR-022` AC2 IS CHECKED AGAINST, and that is why it is a
// package of its own rather than a mux built inline in the main package.
//
// AC2 obliges a planning request to issue no all-zones request and to wait on no
// refresh. `FR-022`'s `Rationale` states why that is one obligation with AC1
// rather than a refinement of it: a job that runs on a schedule and can also be
// triggered by a request is not off the request path. The half a convenience
// trigger removes is this one, and it is removed in good faith by someone who
// never reads the record.
//
// The difficulty is that AC2 is satisfiable by absence. A test asserting that
// the plan handler issues no all-zones request measures only the handler it
// runs — today `POST /journeys` (#1), behind a fake routing provider and with
// no zone selection yet — and keeps reporting green on the day the handler
// acquires a dependency wired the wrong way. The guard is therefore structural
// and it fails closed:
//
//   - The invariant is that no package in this module may reach
//     `internal/zonesync` transitively except the composition root and
//     `internal/syncstore`, the sync's write-path adapter, which must name
//     the ports it implements — and that THIS package reaches neither of those
//     two, which is the leg the criterion is about. The exemption is argued at
//     the check rather than here. Wherever a future handler lands, in this
//     package or beside it, the invariant already covers it — it does not have
//     to be remembered.
//   - The check refuses rather than passes when it cannot find the packages it
//     ranges over, including this one. Deleting this package does not make the
//     check vacuous, it makes it red.
//
// The check lives beside the thing it guards, in
// `internal/zonesync/offrequestpath_test.go`.
package httpapi

import (
	"net/http"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// Option supplies the request surface with something a handler needs.
type Option func(*options)

type options struct {
	planner     *planner.Planner
	planTimeout time.Duration
}

// WithPlanner has `POST /journeys` plan through p. Without it, or with a nil
// p, that route answers 503: the composition root registers a planner only
// when a routing engine is configured (#6), and a route that is served and
// says so is easier to find than one that is missing.
func WithPlanner(p *planner.Planner) Option {
	return func(o *options) { o.planner = p }
}

// WithPlanTimeout bounds each `POST /journeys` by d, measured from the moment
// its handler is entered: the request body's read and every leg of the plan
// fall inside it, and a plan still routing when it expires is stopped and
// answered with 504. Without it, or with d not positive, a plan is bounded by
// nothing but the request's own context.
//
// The bound is there for the server's write budget, which Go arms as the
// handler is entered and which a plan routed one leg at a time can otherwise
// outlast with no answer written at all (#6). The composition root sets it
// below that budget, so the refusal is written while the connection can still
// carry it.
func WithPlanTimeout(d time.Duration) Option {
	return func(o *options) { o.planTimeout = d }
}

// NewMux returns the service's request surface.
//
// It takes no zone-sync dependency and must never acquire one. What a request
// may read of the sync's work is the currency `zonestore` records — a read of
// `sync_run`, which is what the system observed — and never the worker that
// produces it.
func NewMux(opts ...Option) *http.ServeMux {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", root)
	mux.HandleFunc("POST /journeys", planJourney(o.planner, o.planTimeout))

	return mux
}

// root answers the service's root path. It exists so the process is observably
// serving; `NFR-003` AC2 is measured against it.
func root(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	_, _ = w.Write([]byte("TurfGPS service\n"))
}
