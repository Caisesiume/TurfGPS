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
// What this package does not do yet, and where each lands:
//
//   - Refusing a request with no destination (`FR-005`, #4). Request has no way
//     to omit one, so that refusal belongs to whatever decodes a request.
//   - Refusing a return to the origin that names no distinct intermediate
//     destination (#5).
//   - Zone selection, scoring, and the time budget, which later stories own.
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

// Planner plans journeys through a RoutingProvider.
type Planner struct {
	routing RoutingProvider
}

// New returns a Planner that routes through routing.
func New(routing RoutingProvider) (*Planner, error) {
	if routing == nil {
		return nil, errors.New("no routing provider")
	}

	return &Planner{routing: routing}, nil
}

// Plan plans the journey req names. Every journey it returns begins at
// req.Origin, passes through req.Intermediates in the order they were entered,
// and ends at req.Destination.
//
// A location that is not a coordinate is refused before the provider is asked,
// with an error wrapping ErrNotACoordinate.
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

// within reports whether v lies in [lo, hi]. NaN fails both comparisons, so it
// is never within anything.
func within(v, lo, hi float64) bool {
	return v >= lo && v <= hi
}
