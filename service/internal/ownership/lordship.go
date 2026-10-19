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
func RegionLord(ctx context.Context, fetch FetchRegionsFunc, userID int64) (bool, error) {
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

	held := false

	for i, r := range regions {
		if r.RegionLord == nil {
			continue
		}

		if r.RegionLord.ID == nil {
			return false, fmt.Errorf("region record %d names a regionLord with no id", i)
		}

		if *r.RegionLord.ID == userID {
			held = true
		}
	}

	return held, nil
}
