package zonesync

import "fmt"

// Staged is what one query counts over the loaded staging table, together with
// the size of the table it is about to be merged into.
//
// It exists so the assertions of `Architecture.md § The sync write path` are
// decided by a pure function over counts rather than inside the store adapter.
// The counting needs a database; deciding what the counts mean does not, and
// separating them is what makes the decision testable at all — a merge refused
// for the wrong reason and a merge allowed through are both silent.
type Staged struct {
	// Rows is every row loaded into the staging table.
	Rows int

	// DistinctIDs counts the distinct non-null ids among them.
	DistinctIDs int

	// NullIDs counts rows whose id is null.
	NullIDs int

	// OutOfRange counts rows whose latitude or longitude is missing or outside
	// the permitted range.
	OutOfRange int

	// FutureCreated counts rows whose date_created is missing or later than the
	// run's own start instant.
	FutureCreated int

	// CurrentZoneRows is how many rows `zone` holds now, which is what the row
	// count is held to a floor of.
	CurrentZoneRows int
}

// Verify applies the four staging assertions and returns the reasons the staged
// response must not be merged. An empty result is a pass.
//
// It returns every reason rather than the first, because all four are computed
// by one query and a run that is refused should record what was wrong with the
// response and not merely that something was.
//
// minZoneRatio is the floor's fraction. `Architecture.md § The sync write path`
// requires the floor and states no figure; where the figure comes from, and why
// it is a choice rather than a fact, is on `config` alongside its default.
func (s Staged) Verify(minZoneRatio float64) []string {
	var reasons []string

	// 1. The row count against a floor of the current table's count. A
	// truncated response must not be merged at all: `Architecture.md § Absence
	// is recorded and never acted on` records what merging one costs and why
	// the next opportunity to repair it is a whole interval away.
	floor := int(minZoneRatio * float64(s.CurrentZoneRows))
	if s.Rows < floor {
		reasons = append(reasons, fmt.Sprintf("the response carries %d zones, below the floor of %d for the %d rows now held — a truncated response is not merged", s.Rows, floor, s.CurrentZoneRows))
	}

	// An empty response is the extreme of the same failure, and the floor alone
	// does not catch it: against an empty table the floor is zero and every
	// count clears it, so the first run of a fresh store would record `ok` for
	// a response that carried nothing.
	if s.Rows == 0 {
		reasons = append(reasons, "the response carries no zones")
	}

	// 2. No duplicate ids. The id is the merge's conflict target, so a
	// duplicate makes the merge's result depend on which of the two rows the
	// planner happens to reach last.
	if s.NullIDs > 0 {
		reasons = append(reasons, fmt.Sprintf("%d staged rows carry no id", s.NullIDs))
	}

	if uniques := s.Rows - s.NullIDs; uniques != s.DistinctIDs {
		reasons = append(reasons, fmt.Sprintf("%d staged rows carry only %d distinct ids", uniques, s.DistinctIDs))
	}

	// 3. Every coordinate within range. `zone` carries the same two checks as
	// constraints, so this is where they are reported rather than where a merge
	// aborts a transaction on them.
	if s.OutOfRange > 0 {
		reasons = append(reasons, fmt.Sprintf("%d staged rows carry a coordinate outside the permitted range", s.OutOfRange))
	}

	// 4. No date_created later than the run's own start instant. This is the
	// constraint that could not be a CHECK — its bound is the run's start, not
	// a constant — which is why it is asserted here, where it can be.
	if s.FutureCreated > 0 {
		reasons = append(reasons, fmt.Sprintf("%d staged rows carry a dateCreated later than this run's start", s.FutureCreated))
	}

	return reasons
}
