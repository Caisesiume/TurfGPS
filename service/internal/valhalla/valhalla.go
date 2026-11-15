// Package valhalla is the Valhalla adapter behind the routing half of the
// `RoutingProvider` port (`Architecture.md § Ports and adapters`, `§ D3`),
// which `internal/planner` declares and this package implements. The
// composition root registers it; nothing but the root imports it.
//
// EVERY ROUTE IT RETURNS IS ONE A CAR CAN DRIVE — `FR-006`, as the port's
// contract states it. Two things hold that, and neither alone would:
//
//   - The request names the `auto` costing and sets no costing option. One
//     option would undo the requirement on its own: `ignore_access` sends a car
//     down a pedestrian street, which is the red run the integration test in
//     this package was proved against.
//   - The answer is checked rather than trusted. Every maneuver of the leg must
//     report the travel mode `drive` and the travel type `car`; a leg that
//     reports any other is refused, so a leg the engine itself says was not
//     driven is never offered as a driving leg.
//
// EVERY LEG COMES WITH THE WAYS IT RUNS ALONG, so the planner can check it
// against the access markings (#255, ADR-0008, `Architecture.md § The route
// check`, step 2). The route answer names no way, so the leg's shape is sent
// back to the engine's map-matching service, `trace_attributes`, with
// `shape_match=edge_walk` — the documented match for "the shape from a prior
// Valhalla route" — asking only for each edge's `way_id`, `begin_shape_index`
// and `end_shape_index`, the matched shape, and each input point's match. A
// shape that does not match, a point left unmatched, an edge without a way id,
// or edges that do not join end to start and cover the matched shape end to
// end refuse the leg; nothing is guessed. The leg returned is the matched
// shape cut into those edges, so the geometry the planner checks is the
// geometry it hands on.
//
// THE ENGINE DOES NOT KEEP A CAR OFF PRIVATE AND ACCESS-RESTRICTED WAYS, and
// is not trusted to. At 3.9.1, as the safety review of #246 read
// `lua/graph.lua` and `src/mjolnir/pbfgraphparser.cc`, the `auto` costing
// admits a way tagged `access=private`, `destination`, `customers`,
// `delivery`, `permit` or `residents` as destination-only at a penalty, passes
// a private gate or bollard at a penalty, and treats `access=restricted`, and
// any value its tables do not list, as open (ADR-0008, § Context). The planner
// checks every leg it is given and refuses one crossing a marking in transit;
// this adapter names the ways, and keeps off what it is told to.
//
// WHAT THE PLANNER ASKS IT TO AVOID IS SENT AS `exclude_locations`, and as
// nothing else (#248, `Architecture.md § Avoiding, not only refusing`). Each
// place goes as a location the engine maps to "the closest road or roads" and
// routes around (route API reference, § Other request options), which for a
// point on a marked way is that way and for a barrier's position the roads
// through it. `exclude_polygons` and every costing option stay unsent: the
// first is limited to a hundred vertices across a request at 3.9.1's
// defaults, and the second alters what the car costing may use. An answer to
// an avoiding request is checked like any other, so an engine that ignored the
// exclusions would give a refused leg, never a crossing one.
//
// THE FIELDS ARE VALHALLA'S DOCUMENTED ONES, at the version CI runs (3.9.1),
// and none is assumed. Each is in the route API reference,
// https://github.com/valhalla/valhalla/blob/3.9.1/docs/docs/api/route/api-reference.md:
// `locations` with `lat` and `lon` (§ Locations), `exclude_locations` (§ Other
// request options), `costing` and the `auto`
// model (§ Costing models), `directions_type` (§ Directions options), a leg's
// `shape` and its maneuvers' `travel_mode` and `travel_type` (§ Trip legs and
// maneuvers), and the error body's `error_code` and `error` (§ HTTP status
// codes and conditions, § Internal error codes and conditions). The trace's
// are in the map-matching reference at the same version,
// https://github.com/valhalla/valhalla/blob/3.9.1/docs/docs/api/map-matching.md:
// `encoded_polyline`, `shape_match` and `filters` (§ Shape-matching
// parameters, § Attribute filters), the `edges` with `way_id`,
// `begin_shape_index` and `end_shape_index` (§ Edge items), the trip `shape`,
// and `matched_points` with `type` and `edge_index` (§ Matched point items).
// The shape's encoding is docs/docs/api/decoding.md at the same version.
//
// NOTHING HERE LOGS, and no error quotes a location. A route's origin is
// frequently the user's home (`Architecture.md § Personal data`), and the
// request surface answers a routing failure with a fixed message rather than
// the error, for the same reason.
package valhalla

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// The values this adapter asks for and holds a route to. They are Valhalla's
// names, documented where the package doc cites.
const (
	// costingCar is the costing model for travel by car.
	costingCar = "auto"

	// directionsManeuvers asks for maneuvers without narrative instructions:
	// the maneuvers are what carry each part of the route's travel mode, and
	// the instructions are text nothing here reads.
	directionsManeuvers = "maneuvers"

	travelModeDrive = "drive"
	travelTypeCar   = "car"
)

// maxResponseBytes caps what one route response may put in memory. It is a
// guard against a runaway body and not a prediction of a route's size, which
// nothing has measured at the scale of `Architecture.md § D5`: the 1.8 km leg
// of this package's integration test answers in about 4 kB, and a leg too long
// for this cap is refused by an error rather than by the process running out
// of memory.
const maxResponseBytes = 16 << 20

// Client routes car legs through one Valhalla instance.
type Client struct {
	routeURL string
	traceURL string
	timeout  time.Duration
	http     *http.Client
}

var _ planner.RoutingProvider = (*Client)(nil)

// New returns a Client for the Valhalla instance at baseURL, whose route API
// is the `route` path under it and whose map-matching API is the
// `trace_attributes` path, and which allows each call timeout to answer.
//
// baseURL is taken as `config.LoadValhalla` holds it; New refuses only what it
// cannot join a path onto.
func New(baseURL string, timeout time.Duration) (*Client, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("a route timeout of %v is not positive", timeout)
	}

	routeURL, err := url.JoinPath(baseURL, "route")
	if err != nil {
		return nil, fmt.Errorf("the routing engine's URL cannot carry the route path: %w", err)
	}

	traceURL, err := url.JoinPath(baseURL, "trace_attributes")
	if err != nil {
		return nil, fmt.Errorf("the routing engine's URL cannot carry the trace path: %w", err)
	}

	return &Client{
		routeURL: routeURL,
		traceURL: traceURL,
		timeout:  timeout,
		http: &http.Client{
			// A redirect is refused rather than followed. The engine is the
			// deployment's own; one that redirects is misconfigured, and
			// following it would send the user's locations wherever it says.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}, nil
}

// ErrNotDriven is the error Route wraps when the engine answers with a leg it
// does not report as driven by car.
var ErrNotDriven = errors.New("the routing engine did not route the leg for a car")

// ErrWaysUnnamed is the error Route wraps when the engine could not name the
// ways a leg runs along, so the leg cannot be checked against the access
// markings and is not served.
var ErrWaysUnnamed = errors.New("the routing engine did not name the ways the leg runs along")

type location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type routeRequest struct {
	Locations        []location `json:"locations"`
	ExcludeLocations []location `json:"exclude_locations,omitempty"`
	Costing          string     `json:"costing"`
	DirectionsType   string     `json:"directions_type"`
}

type routeResponse struct {
	Trip *struct {
		Legs []struct {
			Shape     *string `json:"shape"`
			Maneuvers []struct {
				TravelMode string `json:"travel_mode"`
				TravelType string `json:"travel_type"`
			} `json:"maneuvers"`
		} `json:"legs"`
	} `json:"trip"`
}

// The map-matching request: the route's own encoded shape, walked edge by
// edge, with the four attributes the check reads and no others.
const shapeMatchEdgeWalk = "edge_walk"

var traceAttributes = []string{"edge.way_id", "edge.begin_shape_index", "edge.end_shape_index", "shape"}

type traceFilters struct {
	Attributes []string `json:"attributes"`
	Action     string   `json:"action"`
}

type traceRequest struct {
	EncodedPolyline string       `json:"encoded_polyline"`
	Costing         string       `json:"costing"`
	ShapeMatch      string       `json:"shape_match"`
	Filters         traceFilters `json:"filters"`
}

type traceResponse struct {
	Shape *string `json:"shape"`
	Edges []struct {
		WayID           *int64 `json:"way_id"`
		BeginShapeIndex *int   `json:"begin_shape_index"`
		EndShapeIndex   *int   `json:"end_shape_index"`
	} `json:"edges"`
}

type errorResponse struct {
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

// Route returns the route a car drives from one location to the other, as
// Valhalla's `auto` costing finds it keeping off the roads at each place avoid
// names, cut into the OpenStreetMap ways it runs along.
//
// The route's ends are where the engine placed the locations on its graph —
// for a car, on the nearest road a car may use — and so are generally near
// the locations rather than at them. A failure is an error: the engine
// unreachable or slower than the timeout, an answer it marks as an error, or
// one this adapter cannot hold to the documented shape, including a leg not
// reported as driven by car, which wraps ErrNotDriven, and a leg whose ways
// the engine could not name, which wraps ErrWaysUnnamed.
func (c *Client) Route(ctx context.Context, from, to planner.Point, avoid []planner.Point) ([]planner.Edge, error) {
	var excluded []location
	for _, p := range avoid {
		excluded = append(excluded, location{Lat: p.Latitude, Lon: p.Longitude})
	}

	raw, err := c.post(ctx, c.routeURL, routeRequest{
		Locations:        []location{{Lat: from.Latitude, Lon: from.Longitude}, {Lat: to.Latitude, Lon: to.Longitude}},
		ExcludeLocations: excluded,
		Costing:          costingCar,
		DirectionsType:   directionsManeuvers,
	})
	if err != nil {
		return nil, err
	}

	shape, err := driven(raw)
	if err != nil {
		return nil, err
	}

	raw, err = c.post(ctx, c.traceURL, traceRequest{
		EncodedPolyline: shape,
		Costing:         costingCar,
		ShapeMatch:      shapeMatchEdgeWalk,
		Filters:         traceFilters{Attributes: traceAttributes, Action: "include"},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWaysUnnamed, err)
	}

	edges, err := tracedEdges(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWaysUnnamed, err)
	}

	return edges, nil
}

// post sends one request to the engine under the call timeout and returns the
// answer's body, or an error for an answer other than 200.
func (c *Client) post(ctx context.Context, target string, request any) ([]byte, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("encoding the request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building the request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("asking the routing engine: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading the routing engine's answer: %w", err)
	}

	if len(raw) > maxResponseBytes {
		return nil, fmt.Errorf("the routing engine's answer exceeds %d bytes", maxResponseBytes)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, engineError(resp.StatusCode, raw)
	}

	return raw, nil
}

// engineError describes an answer other than 200 by the code and message the
// engine's error body carries, where it carries them.
func engineError(status int, raw []byte) error {
	var e errorResponse
	if err := json.Unmarshal(raw, &e); err != nil || e.ErrorCode == 0 {
		return fmt.Errorf("the routing engine answered %d", status)
	}

	return fmt.Errorf("the routing engine answered %d, error %d: %s", status, e.ErrorCode, e.Error)
}

// driven returns the encoded shape of the one leg a two-location route answer
// holds, once every maneuver of it is reported as driven by car and the shape
// decodes.
func driven(raw []byte) (string, error) {
	var r routeResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", fmt.Errorf("the routing engine's answer is not a route: %w", err)
	}

	if r.Trip == nil {
		return "", errors.New("the routing engine's answer holds no trip")
	}

	// Two break locations make one leg (§ Trip legs and maneuvers).
	if n := len(r.Trip.Legs); n != 1 {
		return "", fmt.Errorf("the routing engine's answer holds %d legs for two locations, want 1", n)
	}

	leg := r.Trip.Legs[0]

	if len(leg.Maneuvers) == 0 {
		return "", fmt.Errorf("%w: the leg carries no maneuvers, so nothing says how it is travelled", ErrNotDriven)
	}

	for i, m := range leg.Maneuvers {
		if m.TravelMode != travelModeDrive || m.TravelType != travelTypeCar {
			return "", fmt.Errorf("%w: maneuver %d is travelled as %q/%q, want %q/%q",
				ErrNotDriven, i, m.TravelMode, m.TravelType, travelModeDrive, travelTypeCar)
		}
	}

	if leg.Shape == nil {
		return "", errors.New("the routing engine's leg carries no shape")
	}

	shape, err := decodeShape(*leg.Shape)
	if err != nil {
		return "", fmt.Errorf("the routing engine's leg shape: %w", err)
	}

	if len(shape) == 0 {
		return "", errors.New("the routing engine's leg shape has no vertices")
	}

	return *leg.Shape, nil
}

// tracedEdges cuts the matched shape of a trace answer into its edges.
//
// IT TRUSTS THE WALK AND CHECKS THE ANSWER. An `edge_walk` answer is the
// route's own path: at 3.9.1 the walk starts on an edge the shape's first
// point lies on, follows the shape point by point, and succeeds only on
// reaching an edge its last point lies on (`src/thor/route_matcher.cc`,
// RouteMatcher::FormPath), so a shape it cannot walk whole is an error, which
// Route refuses. What is checked here is that the answer says so whole: at
// least one edge, every edge naming a way, and the edges' stretches joining
// end to start and spanning the matched shape from its first point to its
// last.
func tracedEdges(raw []byte) ([]planner.Edge, error) {
	var r traceResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("the trace answer is not one: %w", err)
	}

	if r.Shape == nil {
		return nil, errors.New("the trace answer carries no shape")
	}

	shape, err := decodeShape(*r.Shape)
	if err != nil {
		return nil, fmt.Errorf("the trace answer's shape: %w", err)
	}

	if len(shape) < 2 {
		return nil, fmt.Errorf("the trace answer's shape has %d points", len(shape))
	}

	if len(r.Edges) == 0 {
		return nil, errors.New("the trace answer holds no edges")
	}

	edges := make([]planner.Edge, 0, len(r.Edges))
	next := 0

	for i, e := range r.Edges {
		if e.WayID == nil || *e.WayID <= 0 {
			return nil, fmt.Errorf("edge %d names no way", i)
		}

		if e.BeginShapeIndex == nil || e.EndShapeIndex == nil {
			return nil, fmt.Errorf("edge %d carries no stretch of the shape", i)
		}

		begin, end := *e.BeginShapeIndex, *e.EndShapeIndex

		if begin != next || end < begin || end >= len(shape) {
			return nil, fmt.Errorf("edge %d spans points %d to %d of %d, want it to start at %d", i, begin, end, len(shape), next)
		}

		stretch := shape[begin : end+1]
		if len(stretch) == 1 {
			stretch = []planner.Point{stretch[0], stretch[0]}
		}

		edges = append(edges, planner.Edge{WayID: *e.WayID, Shape: append([]planner.Point(nil), stretch...)})
		next = end
	}

	if next != len(shape)-1 {
		return nil, fmt.Errorf("the edges end at point %d of %d, want the shape covered to its last", next, len(shape))
	}

	return edges, nil
}
