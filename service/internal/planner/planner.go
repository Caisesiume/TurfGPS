// Package planner plans a journey from an origin to a destination, through any
// intermediate destinations the user entered, in the order entered — `FR-001`,
// `FR-002`, `FR-003`.
//
// IT DECLARES THE PORT AND HOLDS NO ADAPTER. Routes come through
// RoutingProvider, the routing half of the `RoutingProvider` port in
// `Architecture.md § Ports and adapters`. The first adapter behind it is
// Valhalla, and it is #6's; nothing here may import a concrete provider, so the
// planner is tested against a fake and runs unchanged against whichever adapter
// the composition root registers.
//
// THE REQUIRED LOCATIONS ARE THE REQUEST'S, NEVER THE ROUTE'S. A journey's first
// location is the entered origin and its last is the entered destination, which
// is `FR-001`'s acceptance, and they are taken from the request rather than read
// back off the route geometry: an engine snaps a location to its own graph, so
// the first vertex of a route is generally near the origin and not at it. The
// intermediate destinations between them are the request's too, for the same
// reason.
//
// THE ENTERED ORDER IS NEVER CHANGED. The intermediate destinations are the
// journey's mandatory waypoints, and `SPECIFICATION.md § Primary use case` lets
// planning choose the roads between consecutive required locations but never
// reorder them (`FR-003`): a shorter journey in another order is a different
// journey from the one the user has to make. Each leg is routed between one
// required location and the next, in the order entered, and nothing here
// compares orders.
//
// A JOURNEY IS PLANNED FROM WHATEVER LOCAL ZONE COPY IS HELD, AND REFUSED ONLY
// WHERE NONE IS (#33). `FR-024` and `FR-033` are one gate reading one input,
// and that input is ZoneCopy.Held: whether the system holds a local synced copy
// of the zone set at all. A copy mid-refresh, a copy as old as the staleness
// bound of `Architecture.md § Retrieving zones`, and a copy older than it, or
// with no refresh on record, are all held, so all are planned from. The gate
// never reads the copy's age, because the only answer age may change is what a
// result says about itself, never whether it is produced: `FR-033`'s
// `Rationale` warns that a gate refusing more often than absence requires
// fails the half that keeps the product working.
//
// What this package does not do yet, and where each lands:
//
//   - Refusing a request with no destination (`FR-005`, #4). Request has no way
//     to omit one, so that refusal belongs to whatever decodes a request.
//   - Refusing a return to the origin that names no distinct intermediate
//     destination (#5).
//   - Zone selection, scoring, and the time budget, which later stories own.
//   - Recording on a result that it was built from a copy stale beyond a bound
//     (`FR-034`, #35). That marking reads the copy's age beside the gate in
//     Plan, never inside it: a stale copy is marked and still planned from.
package planner

import (
	"context"
	"errors"
	"fmt"
)

// ErrNotACoordinate is the error Plan wraps when a location of the request is
// not a coordinate, so a caller can tell a request it must refuse from a
// routing failure.
var ErrNotACoordinate = errors.New("not a coordinate")

// ErrNoZoneCopy is the error Plan returns when the system holds no local copy
// of the zone set: `FR-033`'s one ground for planning no journey for lack of
// zone data. How the refusal is shown to a user is not decided anywhere yet.
var ErrNoZoneCopy = errors.New("no local copy of the zone set is held")

// ErrZoneCopyUnreadable is the error Plan wraps when whether a copy is held
// could not be read. It is not ErrNoZoneCopy, and must never be answered as
// it: a store that cannot be reached has not been shown to hold nothing, and
// reporting it as empty is the misreading `internal/zonestore`'s package doc
// warns of.
var ErrZoneCopyUnreadable = errors.New("whether a local copy of the zone set is held could not be read")

// Point is a location on the Earth's surface in WGS 84 degrees.
//
// THE AXES ARE NAMED, NEVER POSITIONAL, for the reason `internal/corridor`'s
// Point gives: which axis is written first is a trap this project has shipped
// once, and a named field leaves no order to get wrong.
type Point struct {
	Latitude  float64
	Longitude float64
}

// Request is what a journey is planned from: the location it starts at, the
// locations it has to pass through on the way, and the location it has to end
// at.
type Request struct {
	Origin Point

	// Intermediates are the journey's intermediate destinations — its mandatory
	// waypoints — in the order the user entered them. There may be none.
	Intermediates []Point

	Destination Point
}

// Journey is a planned journey.
type Journey struct {
	// Locations are the journey's required locations in the order they are
	// visited: the request's origin, its intermediate destinations in the order
	// entered, and its destination.
	Locations []Point

	// Legs holds one leg per consecutive pair of Locations, in the same order.
	Legs []Leg
}

// Leg is the route between two consecutive locations of a journey.
type Leg struct {
	// Shape is the route's geometry as the routing provider returned it. Its
	// first and last vertices are where the engine placed the leg's ends, which
	// need not be the locations themselves.
	Shape []Point
}

// RoutingProvider is what planning needs of a routing engine.
//
// It is declared here, by its consumer, as `internal/corridor` declares its
// half of `ZoneRepository`, so the adapter is substitutable and this package
// depends on no provider. Which travel mode a route is computed for is
// `FR-006`'s and is not decided here.
type RoutingProvider interface {
	// Route returns the geometry of a route from one location to another, at
	// least one vertex long.
	Route(ctx context.Context, from, to Point) ([]Point, error)
}

// ZoneCopy is what planning reads of the local synced copy of the zone set
// before it plans: whether one is held at all (#33).
//
// It is declared here, by its consumer, as RoutingProvider is, so the gate is
// tested against a fake and the store behind it is substitutable.
type ZoneCopy interface {
	// Held reports whether the system holds a local copy of the zone set: at
	// least one zone, whatever its age, and whether or not a refresh of it is
	// running. A copy whose first fill has not committed is not held yet.
	Held(ctx context.Context) (bool, error)
}

// Planner plans journeys through a RoutingProvider, from the zone copy a
// ZoneCopy reports.
type Planner struct {
	routing RoutingProvider
	zones   ZoneCopy
}

// New returns a Planner that routes through routing and plans only while zones
// reports a copy held. Both are required: a planner that could not ask whether
// a copy is held would have to guess, and either guess fails one branch of
// `FR-033`.
func New(routing RoutingProvider, zones ZoneCopy) (*Planner, error) {
	if routing == nil {
		return nil, errors.New("no routing provider")
	}

	if zones == nil {
		return nil, errors.New("no zone copy to plan from")
	}

	return &Planner{routing: routing, zones: zones}, nil
}

// Plan plans the journey req names. Every journey it returns begins at
// req.Origin, passes through req.Intermediates in the order they were entered,
// and ends at req.Destination.
//
// A location that is not a coordinate is refused before the provider is asked,
// with an error wrapping ErrNotACoordinate. Where no zone copy is held, no
// journey is planned and the error is ErrNoZoneCopy; see requireZoneCopy.
func (p *Planner) Plan(ctx context.Context, req Request) (Journey, error) {
	locations := make([]Point, 0, len(req.Intermediates)+2)
	locations = append(locations, req.Origin)
	locations = append(locations, req.Intermediates...)
	locations = append(locations, req.Destination)

	for i, loc := range locations {
		if !within(loc.Latitude, -90, 90) || !within(loc.Longitude, -180, 180) {
			return Journey{}, fmt.Errorf("location %d of the request: %w", i, ErrNotACoordinate)
		}
	}

	if err := p.requireZoneCopy(ctx); err != nil {
		return Journey{}, err
	}

	legs := make([]Leg, 0, len(locations)-1)

	for i := range len(locations) - 1 {
		shape, err := p.routing.Route(ctx, locations[i], locations[i+1])
		if err != nil {
			return Journey{}, fmt.Errorf("routing leg %d: %w", i, err)
		}

		if len(shape) == 0 {
			return Journey{}, fmt.Errorf("routing leg %d: the routing provider returned a route with no geometry", i)
		}

		legs = append(legs, Leg{Shape: shape})
	}

	return Journey{Locations: locations, Legs: legs}, nil
}

// requireZoneCopy is the gate in front of planning, both of its branches in
// one guard (#33): it returns nil wherever a copy is held, and ErrNoZoneCopy
// only where none is.
//
// IT READS PRESENCE AND NOTHING ELSE. A refresh in progress, the age of the
// last refresh, and a refresh never recorded are not its input, so none of
// them can become a refusal here (`FR-024`, and `FR-033`'s first criterion).
// Whatever is later done with the copy's age — marking a result (#35) — is
// done beside this call and never as a third branch of it.
func (p *Planner) requireZoneCopy(ctx context.Context) error {
	held, err := p.zones.Held(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrZoneCopyUnreadable, err)
	}

	if !held {
		return ErrNoZoneCopy
	}

	return nil
}

// within reports whether v lies in [lo, hi]. NaN fails both comparisons, so it
// is never within anything.
func within(v, lo, hi float64) bool {
	return v >= lo && v <= hi
}
