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
// A RETURN TO THE ORIGIN NEEDS A *THERE* TO DRIVE OUT TO (`FR-012`, #5). A
// request whose destination is its origin is refused unless it names an
// intermediate destination distinct from that location: without one the base
// route is zero-length, and every minute driven would price as additional
// time. Distinct means a different location named — the coordinates differ —
// and never a location some distance away: a tolerance for an intermediate
// lying near the origin would be a constant with no home in
// `CalculationSpecification.md`. This refuses on what the request names and
// never plans: a return with a distinct intermediate is planned as any
// journey through intermediate destinations is (`FR-002`).
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
// and a leg crossing any in transit is refused. Refused means not produced:
// there is no least-bad crossing (`FR-125`'s third criterion). The routing
// provider is reached from one method only, checkedLeg, so no route reaches a
// journey by another path (#255 AC4); `checkedleg_test.go` holds that to the
// source.
//
// A LEG MAY RUN ALONG MARKED WAYS ONLY ON THE RUN AT A LOCATION THE USER
// ENTERED (#248, `FR-126`, `RD-056`, `RD-057`, `RD-061`). Every location of a
// Request is one the user entered, so both ends of every leg are such
// locations today; the stopping positions the system chooses are #256's, and
// a leg's ends carry whether each was entered (legEnds) so that #256 passes
// false where it was not. At an entered end, the leg's leading — or trailing —
// edges along marked ways are the run if they are one stretch that ends where
// the leg meets an unmarked way and carries no more marked length than the
// least that joins the location to one, both measured by the AccessMarkings
// port along the geometry the tables hold (`§ The route check`, step 4).
// Everything else on the leg is in transit (`FR-125`'s fourth criterion), and
// a run whose least is undefined refuses the leg. run.go holds the reading.
//
// A REFUSED LEG IS ASKED FOR AGAIN, AVOIDING WHAT IT CROSSED (ADR-0008 D4,
// `§ Avoiding, not only refusing`). The provider is told the places the check
// found — a point on each marked way in transit, each marked barrier in
// transit, and a point on each way of a run longer than the least that the
// least does not use — and its new answer is checked from the start. The run's
// own ways are never avoided. Avoidance is how a leg avoiding a marking is
// produced; the check is what makes it safe, so a provider that ignores what
// it is told gives a refused leg, never a crossing one. How many times a leg
// is asked for again is configured (AvoidUpTo); where that is spent, or the
// provider finds nothing keeping off the places, the leg is refused with the
// crossings of the last answer checked — a lost leg, never an unsafe one. Only
// the answer that passed reaches the journey, never one that was refused.
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
//   - Areas marked only on the feature enclosing them (`FR-128`, #253): a
//     crossing of a marked area is refused wherever it lies on the leg, and
//     is not avoided.
//   - The stopping positions the system chooses, and the gap short of one
//     (`FR-127`, #256).
//   - Refusing a request with no destination (`FR-005`, #4). Request has no way
//     to omit one, so that refusal belongs to whatever decodes a request.
//   - Zone selection, scoring, and the time budget, which later stories own.
//   - Recording on a result that it was built from a copy stale beyond a bound
//     (`FR-034`, #35). That marking reads the copy's age beside the gate in
//     Plan, never inside it: a stale copy is marked and still planned from.
package planner

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// ErrNotACoordinate is the error Plan wraps when a location of the request is
// not a coordinate, so a caller can tell a request it must refuse from a
// routing failure.
var ErrNotACoordinate = errors.New("not a coordinate")

// ErrReturnToOrigin is the error Plan returns for a request whose destination
// is its origin and which names no intermediate destination distinct from that
// location (`FR-012`): a journey with nowhere to drive out to.
var ErrReturnToOrigin = errors.New("the destination is the origin and no intermediate destination differs from it")

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
//
// A ROUTE MAY BE ASKED TO KEEP OFF PLACES (#248, ADR-0008 D4). avoid lists
// points the check found an earlier answer crossing a marking at: a point on a
// marked way, or a marked barrier's position. A provider routes around the
// roads at each where it can, and errs where it finds no route that does; it
// is never trusted to have kept off them, since the check reads every answer
// again. An empty avoid asks for the route a car is given with nothing
// avoided.
type RoutingProvider interface {
	// Route returns a route a car can drive from one location to another,
	// kept off the places avoid lists, as the edges it runs along, at least
	// one.
	Route(ctx context.Context, from, to Point, avoid []Point) ([]Edge, error)
}

// Crossing is one access marking a leg crosses.
type Crossing struct {
	// Edge is the index, in the leg's Edges, of the edge that crosses it,
	// which is where on the leg it lies: the run at a location the user
	// entered is told from transit by it.
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
// every marking that refused it and where on the leg each lies (#255 AC2): each
// one in transit, and each on a run the leg was not admitted along. Where the
// leg was asked for again avoiding them, it is the last answer checked.
type CrossingError struct {
	// Leg is the index of the refused leg in the journey.
	Leg int

	// Edges are the refused leg's edges, which Crossings index.
	Edges []Edge

	Crossings []Crossing

	// Avoided is how many times the leg was asked for again, avoiding what
	// the answer before it crossed, before it was refused.
	Avoided int
}

func (e *CrossingError) Error() string {
	names := make([]string, len(e.Crossings))
	for i, c := range e.Crossings {
		names[i] = c.String()
	}

	return fmt.Sprintf("leg %d crosses %d access marking(s), asked for again %d time(s): %s", e.Leg, len(e.Crossings), e.Avoided, strings.Join(names, "; "))
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

	// Run measures a leg's run at a location the user entered (`FR-126`,
	// `Architecture.md § What the tables hold`). run is the leg's edges from
	// that location in the order travelled away from it — at a leg's far end,
	// its last edges reversed, each stretch reversed with them — every one
	// along a marked way; its first point is where the leg's shape meets the
	// location's way, and its last is where the leg meets an unmarked way.
	//
	// It returns the marked length the run travels and the least marked
	// length that joins its first point to an unmarked way, both measured the
	// same way along the geometry the tables hold, so that the run is
	// admitted exactly where the first is no more than the second. An error
	// wrapping ErrNoLeastRun means the least is undefined, or the run could
	// not be placed on the tables' geometry, and refuses the leg; any other
	// error means the run was not measured.
	Run(ctx context.Context, run []Edge) (RunMeasure, error)
}

// RunMeasure is what AccessMarkings.Run measures of a run.
type RunMeasure struct {
	// Marked is the marked length, in metres, the run travels.
	Marked float64

	// Least is the least marked length, in metres, that joins the run's first
	// point to an unmarked way.
	Least float64

	// LeastWays are the ways a least run travels along. A run longer than the
	// least is asked for again avoiding its other ways, and these never.
	LeastWays []int64

	// Exit is the node the run ends at, where it meets the unmarked way: an
	// exit of the run's last way. A marked barrier the first edge after the
	// run passes is the run's only where it is this node. Zero, which no
	// OpenStreetMap node is, admits no barrier.
	Exit int64
}

// ErrNoLeastRun is the error AccessMarkings.Run wraps where the least run is
// undefined — the location's way is not a marked way the tables hold, or no
// unmarked way is reachable from it over marked ways — or where the run cannot
// be placed on the tables' geometry. An undefined least run is no bound: the
// leg it would bound is refused, never read as unbounded
// (`Architecture.md § What the tables hold`).
var ErrNoLeastRun = errors.New("the run at the location has no least run to be held to")

// RoutingData reports which extract the routing engine's data was built from.
//
// How the service reads the tile set's identity is
// `DEPLOYMENT.md § The tile build` ("How the service reads it"): the engine's
// verbose status and the tile record beside its tiles, which
// `internal/valhalla`'s TileIdentity reads and the composition root registers
// (#249). An identity it cannot read is an error, and nothing is planned.
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

	// avoidUpTo is how many times a refused leg is asked for again; zero
	// unless AvoidUpTo sets it.
	avoidUpTo int
}

// Option configures a Planner beyond what New requires.
type Option func(*Planner) error

// AvoidUpTo sets how many times a leg refused for crossing a marking is asked
// for again, avoiding what it crossed (ADR-0008 D4). Without it a refused leg
// is never asked for again, so every leg crossing a marking in transit is
// refused: the stricter state, which loses legs rather than serving any. The
// bound is the deployment's, taken from configuration with its origin
// (`config.LoadAccessAvoidance`), because no document gives it
// (`Architecture.md § What is not decided`). It trades lost legs against the
// engine calls a leg may cost, and no value of it lets a crossing leg through.
func AvoidUpTo(attempts int) Option {
	return func(p *Planner) error {
		if attempts < 0 {
			return fmt.Errorf("an avoidance bound of %d attempts is negative", attempts)
		}

		p.avoidUpTo = attempts

		return nil
	}
}

// New returns a Planner that routes through routing, plans only while zones
// reports a copy held, and produces only legs that markings finds crossing
// nothing, while markings and data report one extract. Every argument is
// required: a planner that could not ask would have to guess, and every guess
// fails one way or the other.
func New(routing RoutingProvider, zones ZoneCopy, markings AccessMarkings, data RoutingData, opts ...Option) (*Planner, error) {
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

	p := &Planner{routing: routing, zones: zones, markings: markings, data: data}

	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}

	return p, nil
}

// Plan plans the journey req names. Every journey it returns begins at
// req.Origin, passes through req.Intermediates in the order they were entered,
// and ends at req.Destination, and no leg of it crosses an access marking but
// along the run at a location the user entered.
//
// A location that is not a coordinate is refused before the provider is asked,
// with an error wrapping ErrNotACoordinate, and a return to the origin naming
// no distinct intermediate destination with ErrReturnToOrigin; see
// returnsWithNowhereBetween. Where no zone copy is held, no
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

	if returnsWithNowhereBetween(req) {
		return Journey{}, ErrReturnToOrigin
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
		// Every location of a request is one the user entered (`RD-056`).
		leg, err := p.checkedLeg(ctx, i, locations[i], locations[i+1], legEnds{fromEntered: true, toEntered: true})
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
// is checked before it is used (#255 AC4), each answer to avoidance included.
//
// An answer that crosses a marking in transit, or runs a run it is not
// admitted along, is asked for again up to avoidUpTo times, each time avoiding
// every place found so far, and the leg is refused where no answer passes. A
// provider that errs on a later ask — finding no route keeping off the places,
// or refusing as many as were sent — leaves the leg refused for the crossings
// of the last answer checked, unless the plan itself has ended.
func (p *Planner) checkedLeg(ctx context.Context, i int, from, to Point, ends legEnds) (Leg, error) {
	var (
		avoid   []Point
		refusal *CrossingError
	)

	for attempt := 0; ; attempt++ {
		edges, err := p.routing.Route(ctx, from, to, avoid)
		if err != nil {
			if refusal == nil || ctx.Err() != nil {
				return Leg{}, fmt.Errorf("routing leg %d: %w", i, err)
			}

			refusal.Avoided = attempt

			return Leg{}, refusal
		}

		shape, err := joinEdges(edges)
		if err != nil {
			return Leg{}, fmt.Errorf("leg %d is %w: %w", i, ErrAccessUnchecked, err)
		}

		crossings, err := p.markings.Crossings(ctx, edges)
		if err != nil {
			return Leg{}, fmt.Errorf("leg %d is %w: %w", i, ErrAccessUnchecked, err)
		}

		refused, places, err := p.refusing(ctx, edges, crossings, ends)
		if err != nil {
			return Leg{}, fmt.Errorf("leg %d is %w: %w", i, ErrAccessUnchecked, err)
		}

		if len(refused) == 0 {
			return Leg{Shape: shape, Edges: edges}, nil
		}

		refusal = &CrossingError{Leg: i, Edges: edges, Crossings: refused, Avoided: attempt}

		more := 0

		for _, pl := range places {
			if !slices.Contains(avoid, pl) {
				avoid = append(avoid, pl)
				more++
			}
		}

		if attempt >= p.avoidUpTo || more == 0 {
			return Leg{}, refusal
		}
	}
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

// returnsWithNowhereBetween reports whether req names one location as both its
// origin and its destination and no intermediate destination distinct from it
// (`FR-012`): none at all, or only that location again. Two points are one
// location exactly where their coordinates are equal, as the package doc
// gives.
func returnsWithNowhereBetween(req Request) bool {
	if req.Origin != req.Destination {
		return false
	}

	for _, via := range req.Intermediates {
		if via != req.Origin {
			return false
		}
	}

	return true
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
