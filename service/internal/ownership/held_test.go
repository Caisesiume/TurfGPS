package ownership

import (
	"context"
	"errors"
	"testing"
)

// fetchPlayerOnce stands in for the player-data call and counts its calls. The
// end-to-end request counts for #23 are in internal/turf/session_test.go,
// against the real adapter; this covers the parse.
func fetchPlayerOnce(body string, calls *int) FetchPlayerFunc {
	return func(context.Context, string) ([]byte, error) {
		*calls++

		return []byte(body), nil
	}
}

func TestLoadPlayerReadsTheIDAndTheHeldZones(t *testing.T) {
	t.Parallel()

	calls := 0

	p, err := LoadPlayer(t.Context(), fetchPlayerOnce(`[{"id":81469,"name":"x","zones":[2305,2306]}]`, &calls), "x")
	if err != nil {
		t.Fatalf("LoadPlayer: %v", err)
	}

	if p.ID != 81469 || !p.Holds(2305) || !p.Holds(2306) || p.Holds(2307) {
		t.Errorf("read id %d, holds 2305=%v 2306=%v 2307=%v; want 81469, true, true, false", p.ID, p.Holds(2305), p.Holds(2306), p.Holds(2307))
	}

	if calls != 1 {
		t.Errorf("LoadPlayer made %d player-data calls, want 1", calls)
	}
}

func TestAPlayerHoldingNothingHoldsNoZone(t *testing.T) {
	t.Parallel()

	calls := 0

	p, err := LoadPlayer(t.Context(), fetchPlayerOnce(`[{"id":5,"zones":[]}]`, &calls), "x")
	if err != nil {
		t.Fatalf("LoadPlayer: %v", err)
	}

	if p.Holds(1) {
		t.Error("a player whose held-zone list is empty was decided to hold zone 1")
	}
}

// TestAPlayerResponseItCannotDecideFromIsRefused: an unknown name (the live
// API answers it with []), a missing id, and a missing zones array are each
// an error, and never a player holding nothing.
func TestAPlayerResponseItCannotDecideFromIsRefused(t *testing.T) {
	t.Parallel()

	for name, body := range map[string]string{
		"an unknown name":   `[]`,
		"two players":       `[{"id":1,"zones":[]},{"id":2,"zones":[]}]`,
		"no id":             `[{"zones":[1]}]`,
		"no zones array":    `[{"id":1}]`,
		"not a player list": `{"id":1,"zones":[]}`,
	} {
		calls := 0
		if _, err := LoadPlayer(t.Context(), fetchPlayerOnce(body, &calls), "x"); err == nil {
			t.Errorf("%s: LoadPlayer succeeded, want it refused", name)
		}
	}

	failing := func(context.Context, string) ([]byte, error) { return nil, errors.New("down") }
	if _, err := LoadPlayer(t.Context(), failing, "x"); err == nil {
		t.Error("a failed player-data call produced a player")
	}
}
