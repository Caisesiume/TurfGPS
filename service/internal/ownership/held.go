package ownership

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Player is the user as one player-data call returned them: their id, and the
// zones they hold (`FR-028`).
//
// The player-data call returns `zones` as exactly the ids the user currently holds,
// which `Architecture.md § The user's held zones are already known` records as
// verified. That is what lets Holds be a membership test against data in hand
// rather than a lookup: one call answers the question for every zone reviewed.
type Player struct {
	// ID is the user's Turf id, which is what RegionLord scans for.
	ID int64

	held map[int32]struct{}
}

// LoadPlayer makes the session's one player-data call for name and reads it.
// It is the only request this file issues; every Holds after it is local.
func LoadPlayer(ctx context.Context, fetch FetchPlayerFunc, name string) (*Player, error) {
	if name == "" {
		return nil, errors.New("no player name to look up")
	}

	body, err := fetch(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("the player-data call for %q: %w", name, err)
	}

	return parsePlayer(body, name)
}

// Holds reports whether the user holds zoneID, by membership in the held-zone
// list the player-data call returned. It issues no request.
func (p *Player) Holds(zoneID int32) bool {
	_, ok := p.held[zoneID]

	return ok
}

type wirePlayer struct {
	ID    *int64    `json:"id"`
	Zones *[]*int32 `json:"zones"`
}

// parsePlayer reads the response to a request for one player.
//
// An unknown name is answered with an empty array, observed against the live
// API on 20 October 2026, so no record is "no such player" rather than a
// player holding nothing. An ABSENT `zones` is refused rather than read as
// none: the API's own documentation shows the array and not when it may be omitted, and
// reading its absence as empty would decide every zone unheld from a field the
// API never sent.
//
// AN ID THAT IS NOT POSITIVE IS REFUSED, NEVER STORED. A player id of 0 would
// match a lord whose id also decoded to 0 (#234, S1), and a `null` in the held
// list decodes to 0 and would make Holds(0) true for a zone the API never
// named (#234, S2). Neither is a value a Turf id takes.
func parsePlayer(body []byte, name string) (*Player, error) {
	var wire []wirePlayer
	if err := json.Unmarshal(body, &wire); err != nil {
		return nil, fmt.Errorf("the player-data response is not a JSON array of players: %w; body %s", err, snippet(body))
	}

	switch len(wire) {
	case 1:
	case 0:
		return nil, fmt.Errorf("the API knows no player named %q", name)
	default:
		return nil, fmt.Errorf("the player-data response for %q holds %d players, not one", name, len(wire))
	}

	w := wire[0]

	if w.ID == nil {
		return nil, fmt.Errorf("the player-data response for %q has no id", name)
	}

	if *w.ID <= 0 {
		return nil, fmt.Errorf("the player-data response for %q has id %d, which is no Turf id", name, *w.ID)
	}

	if w.Zones == nil {
		return nil, fmt.Errorf("the player-data response for %q has no zones array, so whether they hold any zone is unknown", name)
	}

	held := make(map[int32]struct{}, len(*w.Zones))
	for i, id := range *w.Zones {
		if id == nil {
			return nil, fmt.Errorf("the player-data response for %q has a null at held zone %d", name, i)
		}

		if *id <= 0 {
			return nil, fmt.Errorf("the player-data response for %q holds zone %d, which is no zone id", name, *id)
		}

		held[*id] = struct{}{}
	}

	return &Player{ID: *w.ID, held: held}, nil
}
