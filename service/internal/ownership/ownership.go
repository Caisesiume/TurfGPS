// Package ownership decides what a review session needs to know about who holds
// what: whether the user holds a zone (`FR-028`), whether the user holds a
// region lordship anywhere (`FR-029`), and a zone record's ownership fields
// carried exactly as the API sent them (`FR-030`).
//
// IT NEVER IMPORTS THE TURF ADAPTER, per `Architecture.md § Ports and adapters`.
// It declares the two reads it needs as the function types below, and
// `internal/turf` answers them without importing this package; the composition
// root joins the two. noturf_test.go holds that structurally.
//
// Each decision is made from data already in hand. The player-data call and the
// region response are each fetched once, by LoadPlayer and RegionLord, and
// nothing here fetches again: no cache, no refresh, no re-resolution, which both
// `FR-028` and `FR-029` leave to other owners.
package ownership

import (
	"context"
	"fmt"
)

// FetchPlayerFunc is the player-data call of `Architecture.md § Player data`:
// one request for the named player, returning the response body.
type FetchPlayerFunc func(ctx context.Context, name string) ([]byte, error)

// FetchRegionsFunc is the one region response of `Architecture.md § Region
// lords`: one request, returning the response body.
type FetchRegionsFunc func(ctx context.Context) ([]byte, error)

// snippetLen bounds how much of an unreadable body an error quotes.
const snippetLen = 256

func snippet(body []byte) string {
	if len(body) > snippetLen {
		return fmt.Sprintf("%q…", body[:snippetLen])
	}

	return fmt.Sprintf("%q", body)
}
