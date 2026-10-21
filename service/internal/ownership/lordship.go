package ownership

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type wireRegion struct {
	RegionLord *wireOwner `json:"regionLord"`
}

// RegionLord determines once whether the user holds a region lordship
// anywhere, as one value covering every region (`FR-029`), from exactly one
// region response.
//
// IT TAKES NO REGION, ZONE OR JOURNEY, and that is the decision: the Region
// Lord takeover bonus applies globally, per `Architecture.md § Region lords`, so
// the status is a scan of one response for the user's id and never a join
// between zones and regions. A per-region status would under-apply the bonus at
// every stop outside the region the user leads.
//
// `regionLord` is absent on a region with no lord, observed against the live
// region response on 20 October 2026, and its `id` is the user's id as the
// player-data call returns it. How often this is re-resolved is not this
// function's to decide.
//
// AN ID THAT IS NOT POSITIVE IS REFUSED, NEVER COMPARED (#234, S1): a user id
// and a lord id that both decoded to 0 would otherwise resolve held. And a
// response that names no lord in any region is refused rather than scanned to
// "not held" (#234, N2): the live response carries a lord on a large share of
// regions, per `Architecture.md § Region lords`, so none at all more likely
// says the response stopped carrying them than that nobody leads. The consumer is to read any
// error here as not-lord (PR #230's safety verdict, S5), so the refusal moves
// the value nowhere; it makes the case visible instead of silent.
func RegionLord(ctx context.Context, fetch FetchRegionsFunc, userID int64) (bool, error) {
	if userID <= 0 {
		return false, fmt.Errorf("user id %d is no Turf id, so no lord can be compared with it", userID)
	}

	body, err := fetch(ctx)
	if err != nil {
		return false, fmt.Errorf("the region request: %w", err)
	}

	var regions []wireRegion
	if err := json.Unmarshal(body, &regions); err != nil {
		return false, fmt.Errorf("the region response is not a JSON array of regions: %w; body %s", err, snippet(body))
	}

	// An empty or null array would scan to "not held" from a response that
	// listed nothing to scan.
	if len(regions) == 0 {
		return false, errors.New("the region response lists no regions, so it cannot say whether the user holds any")
	}

	held, lords := false, 0

	for i, r := range regions {
		if r.RegionLord == nil {
			continue
		}

		if r.RegionLord.ID == nil {
			return false, fmt.Errorf("region record %d names a regionLord with no id", i)
		}

		if *r.RegionLord.ID <= 0 {
			return false, fmt.Errorf("region record %d names a regionLord with id %d, which is no Turf id", i, *r.RegionLord.ID)
		}

		lords++

		if *r.RegionLord.ID == userID {
			held = true
		}
	}

	if lords == 0 {
		return false, fmt.Errorf("the region response lists %d regions and a lord on none of them, so it cannot say whether the user holds any", len(regions))
	}

	return held, nil
}
