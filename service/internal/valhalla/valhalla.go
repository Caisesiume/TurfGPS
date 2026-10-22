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
// PRIVATE AND ACCESS-RESTRICTED WAYS ARE NOT YET EXCLUDED, and a route this
// returns may run along one. `SPECIFICATION.md § Enforceable exclusions`
// forbids constructing a route through an area the map data marks as private
// or access-restricted, and the `auto` costing does not enforce it. At 3.9.1,
// as the safety review of #246 read `lua/graph.lua` and
// `src/mjolnir/pbfgraphparser.cc`, it admits a way tagged `access=private`,
// `destination`, `customers`, `delivery`, `permit` or `residents` as
// destination-only at a penalty, passes a private gate or bollard at a
// penalty, and treats `access=restricted` as open. A route that saves more
// than the penalty takes such a way, and a location snapped onto one starts or
// ends the leg on it. The exclusion is a story of its own, owed against this
// adapter ("exclude private and access-restricted ways from driving legs"),
// and no client may call `POST /journeys` before it lands.
//
// THE FIELDS ARE VALHALLA'S DOCUMENTED ONES, at the version CI runs (3.9.1),
// and none is assumed. Each is in the route API reference,
// https://github.com/valhalla/valhalla/blob/3.9.1/docs/docs/api/route/api-reference.md:
// `locations` with `lat` and `lon` (§ Locations), `costing` and the `auto`
// model (§ Costing models), `directions_type` (§ Directions options), a leg's
// `shape` and its maneuvers' `travel_mode` and `travel_type` (§ Trip legs and
// maneuvers), and the error body's `error_code` and `error` (§ HTTP status
// codes and conditions, § Internal error codes and conditions). The shape's
// encoding is docs/docs/api/decoding.md at the same version.
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
	timeout  time.Duration
	http     *http.Client
}

var _ planner.RoutingProvider = (*Client)(nil)

// New returns a Client for the Valhalla instance at baseURL, whose route API
// is the `route` path under it, and which allows each route timeout to answer.
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

	return &Client{
		routeURL: routeURL,
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

type location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type routeRequest struct {
	Locations      []location `json:"locations"`
	Costing        string     `json:"costing"`
	DirectionsType string     `json:"directions_type"`
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

type errorResponse struct {
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

// Route returns the geometry of the route a car drives from one location to
// the other, as Valhalla's `auto` costing finds it.
//
// The route's ends are where the engine placed the locations on its graph —
// for a car, on the nearest road a car may use — and so are generally near
// the locations rather than at them. A failure is an error: the engine
// unreachable or slower than the timeout, an answer it marks as an error, or
// one this adapter cannot hold to the documented shape, including a leg not
// reported as driven by car, which wraps ErrNotDriven.
func (c *Client) Route(ctx context.Context, from, to planner.Point) ([]planner.Point, error) {
	body, err := json.Marshal(routeRequest{
		Locations:      []location{{Lat: from.Latitude, Lon: from.Longitude}, {Lat: to.Latitude, Lon: to.Longitude}},
		Costing:        costingCar,
		DirectionsType: directionsManeuvers,
	})
	if err != nil {
		return nil, fmt.Errorf("encoding the route request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.routeURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building the route request: %w", err)
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

	return driven(raw)
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

// driven returns the shape of the one leg a two-location route answer holds,
// once every maneuver of it is reported as driven by car.
func driven(raw []byte) ([]planner.Point, error) {
	var r routeResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("the routing engine's answer is not a route: %w", err)
	}

	if r.Trip == nil {
		return nil, errors.New("the routing engine's answer holds no trip")
	}

	// Two break locations make one leg (§ Trip legs and maneuvers).
	if n := len(r.Trip.Legs); n != 1 {
		return nil, fmt.Errorf("the routing engine's answer holds %d legs for two locations, want 1", n)
	}

	leg := r.Trip.Legs[0]

	if len(leg.Maneuvers) == 0 {
		return nil, fmt.Errorf("%w: the leg carries no maneuvers, so nothing says how it is travelled", ErrNotDriven)
	}

	for i, m := range leg.Maneuvers {
		if m.TravelMode != travelModeDrive || m.TravelType != travelTypeCar {
			return nil, fmt.Errorf("%w: maneuver %d is travelled as %q/%q, want %q/%q",
				ErrNotDriven, i, m.TravelMode, m.TravelType, travelModeDrive, travelTypeCar)
		}
	}

	if leg.Shape == nil {
		return nil, errors.New("the routing engine's leg carries no shape")
	}

	shape, err := decodeShape(*leg.Shape)
	if err != nil {
		return nil, fmt.Errorf("the routing engine's leg shape: %w", err)
	}

	if len(shape) == 0 {
		return nil, errors.New("the routing engine's leg shape has no vertices")
	}

	return shape, nil
}
