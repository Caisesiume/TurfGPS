package accessstore

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// TestNoToleranceChecksNoLeg: with the barrier tolerance unconfigured, the
// store answers every leg with ErrNoTolerance and never asks the database —
// the pool here points at no server, so a query would fail differently.
func TestNoToleranceChecksNoLeg(t *testing.T) {
	t.Parallel()

	pool, err := pgxpool.New(t.Context(), "postgres://nobody@127.0.0.1:1/none?connect_timeout=1")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(pool.Close)

	s, err := New(pool, nil)
	if err != nil {
		t.Fatal(err)
	}

	edges := []planner.Edge{{WayID: 1, Shape: []planner.Point{{Latitude: 57.7, Longitude: 11.96}, {Latitude: 57.71, Longitude: 11.97}}}}

	if got, err := s.Crossings(t.Context(), edges); !errors.Is(err, ErrNoTolerance) || got != nil {
		t.Errorf("Crossings with no tolerance gave %v, %v; want ErrNoTolerance", got, err)
	}
}

// TestTheToleranceIsCopied: the store keeps its own copy, so the caller's
// variable changing afterwards does not change what is checked.
func TestTheToleranceIsCopied(t *testing.T) {
	t.Parallel()

	pool, err := pgxpool.New(t.Context(), "postgres://nobody@127.0.0.1:1/none")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(pool.Close)

	tolerance := 2.0

	s, err := New(pool, &tolerance)
	if err != nil {
		t.Fatal(err)
	}

	tolerance = 0

	if s.tolerance == nil || *s.tolerance != 2.0 {
		t.Errorf("the store checks with %v, want the 2.0 it was given", s.tolerance)
	}

	if _, err := New(nil, &tolerance); err == nil {
		t.Error("New with no pool returned no error")
	}
}
