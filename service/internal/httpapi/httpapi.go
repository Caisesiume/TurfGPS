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
// The difficulty is that AC2 is satisfiable by absence. There is no plan handler
// yet, so a test asserting that the plan handler issues no all-zones request
// reports green while measuring nothing, and keeps reporting green on the day
// the handler lands wired the wrong way. The guard is therefore structural and
// it fails closed:
//
//   - The invariant is that NO package in this module except the composition
//     root may reach `internal/zonesync`, transitively. Wherever a future
//     handler lands, in this package or beside it, the invariant already covers
//     it — it does not have to be remembered.
//   - The check refuses rather than passes when it cannot find the packages it
//     ranges over, including this one. Deleting this package does not make the
//     check vacuous, it makes it red.
//
// The check lives beside the thing it guards, in
// `internal/zonesync/offrequestpath_test.go`.
package httpapi

import "net/http"

// NewMux returns the service's request surface.
//
// It takes no zone-sync dependency and must never acquire one. What a request
// may read of the sync's work is the currency `zonestore` records — a read of
// `sync_run`, which is what the system observed — and never the worker that
// produces it.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", root)

	return mux
}

// root answers the service's root path. It exists so the process is observably
// serving; `NFR-003` AC2 is measured against it.
func root(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	_, _ = w.Write([]byte("TurfGPS service\n"))
}
