//go:build valhalla

package valhalla

// `FR-006` against a real Valhalla, built from the OpenStreetMap extract in
// testdata/ — the Owner's decision on #6 (`#issuecomment-6334699894`): a wrong
// assumption about the engine surfaces here rather than on the server.
//
// IT IS BUILT ONLY UNDER THE `valhalla` TAG, AND THAT IS NOT A SKIP. The default
// suite does not compile this file, so it cannot print `ok` having measured
// nothing; CI's `valhalla` job builds the tiles, starts the engine, builds this
// file on purpose and fails if it fails. Built under the tag with no engine
// named, it fails rather than skips.
//
// THE STREET. Kungsgatan, inside Göteborg's old moat, is a pedestrian street
// (`highway=pedestrian`, OSM ways 222674401, 304998012 and 304998017 in the
// extract) for some 600 m between two streets a car may use. A car driving
// from one end to the other must go round; on foot it is the straight line.
// The test proves its own premise first — a walking route between the same
// two ends runs along the street — so that a car route that keeps off it is
// evidence about the costing and not about an extract the street has left.

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// testValhallaEnv names the engine under test. It is a test's input and never
// the service's configuration.
const testValhallaEnv = "TURFGPS_TEST_VALHALLA_URL"

// The two ends of the pedestrian section, each on the street a car may use
// beyond it: the residential part of Kungsgatan to the west, Östra Larmgatan
// to the east.
var (
	westEnd = planner.Point{Latitude: 57.70360, Longitude: 11.96110}
	eastEnd = planner.Point{Latitude: 57.70520, Longitude: 11.97140}
)

// alongTheStreet are nodes of way 304998012, the middle block of the
// pedestrian section, each away from any crossing street — so a route passing
// near one is travelling along the pedestrian street, not across it.
var alongTheStreet = []planner.Point{
	{Latitude: 57.7041922, Longitude: 11.9648422},
	{Latitude: 57.7042743, Longitude: 11.9653725},
	{Latitude: 57.7044401, Longitude: 11.9664338},
	{Latitude: 57.7045711, Longitude: 11.9672787},
}

const (
	// onTheStreet is how near a route along the street passes each point. The
	// walking route passes through them; this allows for the engine placing
	// its shape a little off the mapped line.
	onTheStreet = 5.0

	// offTheStreet is how far a car route must keep from every point. The car
	// route Valhalla 3.9.1 finds on this extract keeps 137 m or more from
	// them, and one travelling along the pedestrian street passes within
	// onTheStreet of them all, so the margin is wide on both sides.
	offTheStreet = 25.0
)

func engineUnderTest(t *testing.T) string {
	t.Helper()

	base := os.Getenv(testValhallaEnv)
	if base == "" {
		t.Fatalf("%s is not set: this file is built only under the valhalla tag, for an engine built from testdata/, and it fails rather than skips without one", testValhallaEnv)
	}

	return base
}

// metresFrom returns the least distance in metres from p to the polyline
// shape, on a local flat projection — exact enough over a few kilometres.
func metresFrom(p planner.Point, shape []planner.Point) float64 {
	const metresPerDegreeLat = 111_320.0

	kx := metresPerDegreeLat * math.Cos(p.Latitude*math.Pi/180)
	xy := func(q planner.Point) (float64, float64) {
		return (q.Longitude - p.Longitude) * kx, (q.Latitude - p.Latitude) * metresPerDegreeLat
	}

	best := math.Inf(1)

	for i := range shape {
		ax, ay := xy(shape[i])
		bx, by := ax, ay

		if i+1 < len(shape) {
			bx, by = xy(shape[i+1])
		}

		dx, dy := bx-ax, by-ay
		tt := 0.0

		if l := dx*dx + dy*dy; l > 0 {
			tt = math.Max(0, math.Min(1, -(ax*dx+ay*dy)/l))
		}

		best = math.Min(best, math.Hypot(ax+tt*dx, ay+tt*dy))
	}

	return best
}

// walkingShape asks the engine directly for the walking route between the
// two ends. It is this test's instrument, not the adapter's.
func walkingShape(t *testing.T, base string) []planner.Point {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"locations": []location{{Lat: westEnd.Latitude, Lon: westEnd.Longitude}, {Lat: eastEnd.Latitude, Lon: eastEnd.Longitude}},
		"costing":   "pedestrian",
	})
	if err != nil {
		t.Fatal(err)
	}

	routeURL, err := url.JoinPath(base, "route")
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, routeURL, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("asking the engine for the walking route: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("the walking route answered %d: %s", resp.StatusCode, raw)
	}

	var r routeResponse
	if err := json.Unmarshal(raw, &r); err != nil || r.Trip == nil || len(r.Trip.Legs) != 1 || r.Trip.Legs[0].Shape == nil {
		t.Fatalf("the walking route is not one leg with a shape: %v: %s", err, raw)
	}

	shape, err := decodeShape(*r.Trip.Legs[0].Shape)
	if err != nil {
		t.Fatalf("the walking route's shape: %v", err)
	}

	return shape
}

// TestACarLegIsNeverRoutedAlongAPedestrianStreet is `FR-006` AC2: given a leg
// whose direct way is traversable only on foot, the leg this adapter offers
// as a driving leg does not travel along it.
//
// Its red run is the adapter asking for the `auto` costing with
// `ignore_access` set — the car then drives the length of Kungsgatan, every
// maneuver still reported as `drive`/`car`, and this test fails on the
// distance.
func TestACarLegIsNeverRoutedAlongAPedestrianStreet(t *testing.T) {
	base := engineUnderTest(t)

	walk := walkingShape(t, base)
	for _, p := range alongTheStreet {
		if d := metresFrom(p, walk); d > onTheStreet {
			t.Fatalf("the premise does not hold: the walking route passes %.1f m from %v, want within %.0f m — the extract no longer routes walkers along Kungsgatan, so this test would measure nothing", d, p, onTheStreet)
		}
	}

	c, err := New(base, 30*time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	drive, err := c.Route(t.Context(), westEnd, eastEnd)
	if err != nil {
		t.Fatalf("Route: %v", err)
	}

	for _, p := range alongTheStreet {
		if d := metresFrom(p, drive); d < offTheStreet {
			t.Errorf("the driving leg passes %.1f m from %v, on the pedestrian street, want at least %.0f m: a leg a car cannot use was offered as a driving leg", d, p, offTheStreet)
		}
	}
}

// TestEveryLegOfAJourneyIsRoutedForACar is `FR-006` AC1 through the planner:
// every leg of a journey planned through this adapter against the engine,
// intermediate destinations included, is the engine's car route — every
// request named the `auto` costing, and every maneuver of every answer
// reported `drive`/`car`, which the adapter refuses a leg without.
func TestEveryLegOfAJourneyIsRoutedForACar(t *testing.T) {
	base := engineUnderTest(t)

	c, err := New(base, 30*time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	rec := &recordingTransport{next: http.DefaultTransport}
	c.http.Transport = rec

	p, err := planner.New(c)
	if err != nil {
		t.Fatalf("planner.New: %v", err)
	}

	// Three legs in all: one journey of one leg, and one of two through an
	// intermediate destination (#2).
	for _, req := range []planner.Request{
		{Origin: westEnd, Destination: eastEnd},
		{Origin: eastEnd, Intermediates: []planner.Point{westEnd}, Destination: eastEnd},
	} {
		journey, err := p.Plan(t.Context(), req)
		if err != nil {
			t.Fatalf("Plan(%+v): %v", req, err)
		}

		if len(journey.Legs) != len(journey.Locations)-1 || len(journey.Legs) == 0 {
			t.Fatalf("Plan(%+v) gave %d legs for %d locations", req, len(journey.Legs), len(journey.Locations))
		}

		for i, leg := range journey.Legs {
			if len(leg.Shape) < 2 {
				t.Errorf("leg %d of %+v has %d vertices, want a route", i, req, len(leg.Shape))
			}
		}
	}

	if len(rec.answers) != 3 {
		t.Fatalf("the engine was asked %d times for three legs, want once a leg", len(rec.answers))
	}

	for i, a := range rec.answers {
		if a.costing != "auto" {
			t.Errorf("request %d named the costing %q, want auto", i, a.costing)
		}

		if len(a.modes) == 0 {
			t.Errorf("answer %d carried no maneuvers", i)
		}

		for j, m := range a.modes {
			if m != [2]string{"drive", "car"} {
				t.Errorf("answer %d maneuver %d was travelled as %v, want drive/car", i, j, m)
			}
		}
	}
}

// recordingTransport records, for each route request, the costing it named
// and the travel modes its answer reported.
type recordingTransport struct {
	next    http.RoundTripper
	answers []recordedAnswer
}

type recordedAnswer struct {
	costing string
	modes   [][2]string
}

func (r *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	req.Body = io.NopCloser(bytes.NewReader(reqBody))

	var asked routeRequest
	_ = json.Unmarshal(reqBody, &asked)

	resp, err := r.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if err != nil {
		return nil, err
	}

	resp.Body = io.NopCloser(bytes.NewReader(respBody))

	rec := recordedAnswer{costing: asked.Costing}

	var answered routeResponse
	if json.Unmarshal(respBody, &answered) == nil && answered.Trip != nil {
		for _, leg := range answered.Trip.Legs {
			for _, m := range leg.Maneuvers {
				rec.modes = append(rec.modes, [2]string{m.TravelMode, m.TravelType})
			}
		}
	}

	r.answers = append(r.answers, rec)

	return resp, nil
}
