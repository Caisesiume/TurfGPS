package zonestore

import (
	"context"
	"fmt"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// Reader is the local copy's half of the planning gate's port: whether a copy
// is held is read here (#33).
var _ planner.ZoneCopy = (*Reader)(nil)

// heldSQL asks whether the local copy holds any zone at all — the one input of
// the gate `internal/planner` puts in front of planning (`FR-033`).
//
// IT READS `zone` AND NOTHING ELSE: not `sync_run`, and not the zone feed's
// state. A copy is held whatever refreshed it, however long ago, and with no
// refresh on record at all — `FR-033` plans from each of those — so reading
// "held" as "a refresh has completed" would refuse a copy the zone feed alone
// filled, and Currency is not this question's answer.
//
// IT DOES NOT WAIT ON A REFRESH. The merge writes `zone` under ROW EXCLUSIVE
// and row locks, neither of which blocks a plain read, and what this reads is
// the last committed copy: a refresh in flight is invisible to it until it
// commits, so a held copy stays held throughout one, and a first fill still in
// flight is no copy yet.
const heldSQL = `SELECT EXISTS (SELECT 1 FROM zone)`

// Held reports whether the local copy holds at least one zone.
//
// An error is a failure to read and never an answer: it is not "none held".
func (r *Reader) Held(ctx context.Context) (bool, error) {
	var held bool

	if err := r.pool.QueryRow(ctx, heldSQL).Scan(&held); err != nil {
		return false, fmt.Errorf("reading whether a zone copy is held: %w", err)
	}

	return held, nil
}
