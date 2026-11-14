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
// EVERY CAR LEG IS CHECKED AGAINST THE ACCESS MARKINGS BEFORE IT IS USED
// (#255, `Architecture.md § The route check`, ADR-0008). A route comes back from
// the RoutingProvider as the OpenStreetMap ways it runs along, each with its
// stretch of shape; the AccessMarkings port says which markings those cross;
// and a leg crossing any is refused. Refused means not produced: there is no
// least-bad crossing (`FR-125`'s third criterion). The routing provider is
// reached from one method only, checkedLeg, so no route reaches a journey by
// another path (AC4); `checkedleg_test.go` holds that to the source.
//
// UNTIL #248 LANDS, EVERY CROSSING IS REFUSED. `FR-126`'s run at a location the
// user entered is #248's, so today a home on a private road gets no leg. That
// is the stricter state, and no driver is served it: every story that lets a
// client call `POST /journeys` is blocked through #248. Areas marked only on
// the feature enclosing them are #253's (`FR-128`), and the stopping-position
// checks of `FR-127` are #256's.
//
// THE MARKINGS AND THE ROUTING DATA MUST COME FROM ONE EXTRACT, or nothing is
// planned. A way absent from the markings reads as unmarked, which is true only
// of the extract the tables were loaded from, so the planner reads the SHA-256
// each side reports when it starts a journey and again before it returns one,
// and refuses the journey where either cannot be read, where they differ, or
// where either changed between the reads (`§ Loading from the extract the
// tiles are built from`). An identity that cannot be read is unknown, and
// unknown refuses as a difference does.
//
// What this package does not do yet, and where each lands:
//
//   - Producing the leg that avoids a marking where one exists, and admitting
//     `FR-126`'s run at a location the user entered (#248).
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
	"strings"
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

// ErrAccessUnchecked is the error Plan wraps when a leg, or the journey, could
// not be checked against the access markings: the markings or the routing
// data's identity could not be read, the markings hold nothing, the provider
// answered a leg the check cannot read, or the check itself failed. A leg that
// was not checked is not produced (`FR-125`).
var ErrAccessUnchecked = errors.New("not checked against the access markings")

// ErrExtractMismatch is the error Plan wraps when the access markings and the
// routing data were not built from one extract, or either changed while the
// journey was planned.
var ErrExtractMismatch = errors.New("the access markings and the routing data are not from one extract")

// ErrCrossesMarking is the error a CrossingError matches: a leg ran along a
// way, passed a barrier or crossed an area the map data marks private or
// access-restricted (`FR-125`).
var ErrCrossesMarking = errors.New("the leg crosses an access marking")

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

// Leg is the route between two consecutive locations of a journey, as it was
// checked against the access markings.
type Leg struct {
	// Shape is the leg's geometry: its edges' stretches joined, each sharing
	// its first point with the last point of the one before, so it is exactly
	// the geometry the check read. Its first and last vertices are where the
	// engine placed the leg's ends, which need not be the locations
	// themselves.
	Shape []Point

	// Edges are the OpenStreetMap ways the leg runs along, in order.
	Edges []Edge
}

// Edge is one stretch of a car leg along one OpenStreetMap way.
type Edge struct {
	// WayID is the OpenStreetMap id of the way the stretch runs along.
	WayID int64

	// Shape is the stretch's geometry, at least two points long; a stretch
	// over a single point repeats it.
	Shape []Point
}

// RoutingProvider is what planning needs of a routing engine.
//
// It is declared here, by its consumer, as `internal/corridor` declares its
// half of `ZoneRepository`, so the adapter is substitutable and this package
// depends on no provider.
//
// THE TRAVEL MODE IS PART OF THE CONTRACT, NOT THE ADAPTER'S CHOICE. Every leg
// a journey has today is driven (`FR-006`), so Route answers for a car: an
// adapter that cannot obtain a route a car can drive returns an error, never a
// route by another mode. Stating it here is what binds every adapter to it,
// the openrouteservice one of `Architecture.md § D3` included. The
// park-and-walk leg `FR-006` excepts is a later story's, and it arrives as a
// method of its own rather than as a mode this one may switch to.
//
// A CAR LEG COMES AS THE WAYS IT RUNS ALONG (ADR-0008, § Consequences). The
// planner checks every leg against the access markings by its OpenStreetMap
// way ids and its shape, so Route returns the leg as its edges: every way it
// runs along, in order, each with its stretch of shape, the stretches joining
// end to start and covering the route end to end. A provider that cannot name
// the ways of a route cannot serve a car leg, and returns an error rather than
// a route without them; the openrouteservice adapter inherits the obligation.
// The provider is not trusted to keep off marked ways — it need not, and the
// Valhalla one does not — because the check is what keeps every leg off them.
type RoutingProvider interface {
	// Route returns a route a car can drive from one location to another, as
	// the edges it runs along, at least one.
	Route(ctx context.Context, from, to Point) ([]Edge, error)
}

// Crossing is one access marking a leg crosses.
type Crossing struct {
	// Edge is the index, in the leg's Edges, of the edge that crosses it,
	// which is where on the leg it lies: #248 tells `FR-126`'s run from
	// transit by it.
	Edge int

	// Kind is "way", "barrier" or "area".
	Kind string

	// OSMType and OSMID name the feature: "w" for a way, "n" for a barrier's
	// node, "w" or "r" for an area.
	OSMType string
	OSMID   int64

	// At is a barrier's position. It is the zero Point for a way or an area,
	// whose position is the edge's stretch.
	At Point
}

func (c Crossing) String() string {
	if c.Kind == "barrier" {
		return fmt.Sprintf("%s %s%d at edge %d (%.7f, %.7f)", c.Kind, c.OSMType, c.OSMID, c.Edge, c.At.Latitude, c.At.Longitude)
	}

	return fmt.Sprintf("%s %s%d at edge %d", c.Kind, c.OSMType, c.OSMID, c.Edge)
}

// CrossingError is the refusal of a leg that crosses access markings. It names
// every marking found and where on the leg each lies (#255 AC2).
type CrossingError struct {
	// Leg is the index of the refused leg in the journey.
	Leg int

	// Edges are the refused leg's edges, which Crossings index.
	Edges []Edge

	Crossings []Crossing
}

func (e *CrossingError) Error() string {
	names := make([]string, len(e.Crossings))
	for i, c := range e.Crossings {
		names[i] = c.String()
	}

	return fmt.Sprintf("leg %d crosses %d access marking(s): %s", e.Leg, len(e.Crossings), strings.Join(names, "; "))
}

// Is reports a CrossingError as ErrCrossesMarking.
func (e *CrossingError) Is(target error) bool { return target == ErrCrossesMarking }

// AccessMarkings is what planning reads of the access-marking tables
// (`Architecture.md § The access-marking tables`). It is declared here, by its
// consumer; the PostGIS implementation is `internal/accessstore`.
type AccessMarkings interface {
	// ExtractSHA256 returns the SHA-256 of the extract the markings were
	// loaded from, in lower-case hex. It is an error where no load is current,
	// or where the tables hold no feature, so that an empty table is never
	// read as a map without markings.
	ExtractSHA256(ctx context.Context) (string, error)

	// Crossings returns every marking the edges cross: each marked way an edge
	// runs along, each marked barrier on an edge's way within the barrier
	// tolerance of its stretch, and each marked area a stretch intersects. An
	// error means the edges were not checked, never that they cross nothing.
	Crossings(ctx context.Context, edges []Edge) ([]Crossing, error)
}

// RoutingData reports which extract the routing engine's data was built from.
//
// How the service reads the tile set's identity is `DEPLOYMENT.md`'s to fix
// with #25's tile build (`Architecture.md § What is not decided`); until it
// is, the composition root registers one that reports it unknown, and nothing
// is planned.
type RoutingData interface {
	// ExtractSHA256 returns the SHA-256 of the extract the routing data the
	// engine serves was built from, in lower-case hex, or an error where it
	// cannot be read.
	ExtractSHA256(ctx context.Context) (string, error)
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
// ZoneCopy reports, and checks every leg against AccessMarkings.
type Planner struct {
	routing  RoutingProvider
	zones    ZoneCopy
	markings AccessMarkings
	data     RoutingData
}

// New returns a Planner that routes through routing, plans only while zones
// reports a copy held, and produces only legs that markings finds crossing
// nothing, while markings and data report one extract. Every argument is
// required: a planner that could not ask would have to guess, and every guess
// fails one way or the other.
func New(routing RoutingProvider, zones ZoneCopy, markings AccessMarkings, data RoutingData) (*Planner, error) {
	switch {
	case routing == nil:
		return nil, errors.New("no routing provider")
	case zones == nil:
		return nil, errors.New("no zone copy to plan from")
	case markings == nil:
		return nil, errors.New("no access markings to check legs against")
	case data == nil:
		return nil, errors.New("no identity of the routing data to check the access markings against")
	}

	return &Planner{routing: routing, zones: zones, markings: markings, data: data}, nil
}

// Plan plans the journey req names. Every journey it returns begins at
// req.Origin, passes through req.Intermediates in the order they were entered,
// and ends at req.Destination, and no leg of it crosses an access marking.
//
// A location that is not a coordinate is refused before the provider is asked,
// with an error wrapping ErrNotACoordinate. Where no zone copy is held, no
// journey is planned and the error is ErrNoZoneCopy; see requireZoneCopy. A
// journey the access markings cannot vouch for is refused with an error
// wrapping ErrAccessUnchecked or ErrExtractMismatch, and one with a leg
// crossing a marking with a *CrossingError.
func (p *Planner) Plan(ctx context.Context, req Request) (Journey, error) {
	locations := make([]Point, 0, len(req.Intermediates)+2)
	locations = append(locations, req.Origin)
	locations = append(locations, req.Intermediates...)
	locations = append(locations, req.Destination)

	for i, loc := range locations {
		if !validPoint(loc) {
			return Journey{}, fmt.Errorf("location %d of the request: %w", i, ErrNotACoordinate)
		}
	}

	if err := p.requireZoneCopy(ctx); err != nil {
		return Journey{}, err
	}

	extract, err := p.oneExtract(ctx)
	if err != nil {
		return Journey{}, err
	}

	legs := make([]Leg, 0, len(locations)-1)

	for i := range len(locations) - 1 {
		leg, err := p.checkedLeg(ctx, i, locations[i], locations[i+1])
		if err != nil {
			return Journey{}, err
		}

		legs = append(legs, leg)
	}

	// Read again before returning: a load or a tile build between the two
	// reads means the legs were checked against tables that may no longer
	// match the routing data.
	again, err := p.oneExtract(ctx)
	if err != nil {
		return Journey{}, err
	}

	if again != extract {
		return Journey{}, fmt.Errorf("%w: the extract changed while the journey was planned", ErrExtractMismatch)
	}

	return Journey{Locations: locations, Legs: legs}, nil
}

// checkedLeg routes one leg and checks it against the access markings. It is
// the only caller of the routing provider, so every route the planner obtains
// is checked before it is used (#255 AC4).
func (p *Planner) checkedLeg(ctx context.Context, i int, from, to Point) (Leg, error) {
	edges, err := p.routing.Route(ctx, from, to)
	if err != nil {
		return Leg{}, fmt.Errorf("routing leg %d: %w", i, err)
	}

	shape, err := joinEdges(edges)
	if err != nil {
		return Leg{}, fmt.Errorf("leg %d is %w: %w", i, ErrAccessUnchecked, err)
	}

	crossings, err := p.markings.Crossings(ctx, edges)
	if err != nil {
		return Leg{}, fmt.Errorf("leg %d is %w: %w", i, ErrAccessUnchecked, err)
	}

	if len(crossings) > 0 {
		return Leg{}, &CrossingError{Leg: i, Edges: edges, Crossings: crossings}
	}

	return Leg{Shape: shape, Edges: edges}, nil
}

// joinEdges holds a provider's edges to the contract RoutingProvider states and
// returns the leg's shape. Equality at a join is exact: both sides are the same
// point of one shape, so any difference is a stretch the check would not have
// read.
func joinEdges(edges []Edge) ([]Point, error) {
	if len(edges) == 0 {
		return nil, errors.New("the routing provider returned a route with no edges")
	}

	shape := make([]Point, 0, len(edges)+1)

	for i, e := range edges {
		if e.WayID <= 0 {
			return nil, fmt.Errorf("edge %d names the way id %d", i, e.WayID)
		}

		if len(e.Shape) < 2 {
			return nil, fmt.Errorf("edge %d has %d points, want at least two", i, len(e.Shape))
		}

		for _, pt := range e.Shape {
			if !validPoint(pt) {
				return nil, fmt.Errorf("edge %d holds a point that is not a coordinate", i)
			}
		}

		if i == 0 {
			shape = append(shape, e.Shape...)

			continue
		}

		if shape[len(shape)-1] != e.Shape[0] {
			return nil, fmt.Errorf("edge %d does not start where edge %d ends", i, i-1)
		}

		shape = append(shape, e.Shape[1:]...)
	}

	return shape, nil
}

// oneExtract returns the extract both the access markings and the routing data
// report, or an error where either cannot be read or they differ.
func (p *Planner) oneExtract(ctx context.Context) (string, error) {
	data, err := p.data.ExtractSHA256(ctx)
	if err != nil {
		return "", fmt.Errorf("%w: the routing data's extract is unknown: %w", ErrAccessUnchecked, err)
	}

	markings, err := p.markings.ExtractSHA256(ctx)
	if err != nil {
		return "", fmt.Errorf("%w: the access markings' extract is unknown: %w", ErrAccessUnchecked, err)
	}

	if data == "" || markings == "" {
		return "", fmt.Errorf("%w: an extract's identity is empty", ErrAccessUnchecked)
	}

	if data != markings {
		return "", fmt.Errorf("%w: the routing data is from %s and the access markings from %s", ErrExtractMismatch, data, markings)
	}

	return data, nil
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

// validPoint reports whether p is a coordinate.
func validPoint(p Point) bool {
	return within(p.Latitude, -90, 90) && within(p.Longitude, -180, 180)
}

// within reports whether v lies in [lo, hi]. NaN fails both comparisons, so it
// is never within anything.
func within(v, lo, hi float64) bool {
	return v >= lo && v <= hi
}
