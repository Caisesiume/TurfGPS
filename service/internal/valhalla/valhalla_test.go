package valhalla

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// These tests answer as Valhalla through httptest servers, in the shapes the
// route API reference documents (cited on the package doc). That the engine
// itself answers this way, and keeps a car off a pedestrian street, is the
// integration test's to show, against a real instance: `valhalla_integration_test.go`.

// encodeShape is decodeShape's inverse, for building answers.
func encodeShape(points []planner.Point) string {
	var (
		b          strings.Builder
		pLat, pLon int64
	)

	value := func(v int64) {
		u := v << 1
		if v < 0 {
			u = ^u
		}

		for u >= 0x20 {
			b.WriteByte(byte((0x20 | (u & 0x1f)) + 63))
			u >>= 5
		}

		b.WriteByte(byte(u + 63))
	}

	for _, p := range points {
		lat := int64(math.Round(p.Latitude * shapePrecision))
		lon := int64(math.Round(p.Longitude * shapePrecision))
		value(lat - pLat)
		value(lon - pLon)
		pLat, pLon = lat, lon
	}

	return b.String()
}

type maneuver struct {
	TravelMode string `json:"travel_mode"`
	TravelType string `json:"travel_type"`
}

type leg struct {
	Shape     string     `json:"shape"`
	Maneuvers []maneuver `json:"maneuvers"`
}

var byCar = maneuver{TravelMode: "drive", TravelType: "car"}

// answer is a route answer holding legs, in the documented shape.
func answer(legs ...leg) string {
	b, err := json.Marshal(map[string]any{"trip": map[string]any{"status": 0, "legs": legs}})
	if err != nil {
		panic(err)
	}

	return string(b)
}

// recordedRequest is what the fake engine was asked, decoded generically so
// that a field the adapter should not send is seen if it is sent.
type recordedRequest struct {
	Method string
	Path   string
	Body   map[string]any
}

// fakeEngine answers every route request with respond's answer and records
// the request.
type fakeEngine struct {
	mu      sync.Mutex
	asked   []recordedRequest
	respond func(req map[string]any) (int, string)
}

func (f *fakeEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)

	var body map[string]any
	_ = json.Unmarshal(raw, &body)

	f.mu.Lock()
	f.asked = append(f.asked, recordedRequest{Method: r.Method, Path: r.URL.Path, Body: body})
	f.mu.Unlock()

	status, out := f.respond(body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, out)
}

func startEngine(t *testing.T, respond func(map[string]any) (int, string)) (*fakeEngine, *Client) {
	t.Helper()

	engine := &fakeEngine{respond: respond}
	srv := httptest.NewServer(engine)
	t.Cleanup(srv.Close)

	c, err := New(srv.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return engine, c
}

// straightByCar answers every request with a two-vertex car leg between the
// locations it was asked for.
func straightByCar(req map[string]any) (int, string) {
	locs, _ := req["locations"].([]any)
	if len(locs) != 2 {
		return http.StatusBadRequest, `{"error_code":120,"error":"Insufficient number of locations provided","status_code":400}`
	}

	var shape []planner.Point

	for _, l := range locs {
		m, _ := l.(map[string]any)
		lat, _ := m["lat"].(float64)
		lon, _ := m["lon"].(float64)
		shape = append(shape, planner.Point{Latitude: lat, Longitude: lon})
	}

	return http.StatusOK, answer(leg{Shape: encodeShape(shape), Maneuvers: []maneuver{byCar, byCar}})
}

// TestEveryLegIsRequestedForACar is `FR-006` AC1 at the request: every leg of
// every journey planned through this adapter, intermediate destinations
// included, is asked of the engine once, between the leg's two locations in
// order, with the `auto` costing and no costing option.
func TestEveryLegIsRequestedForACar(t *testing.T) {
	t.Parallel()

	engine, c := startEngine(t, straightByCar)

	p, err := planner.New(c)
	if err != nil {
		t.Fatalf("planner.New: %v", err)
	}

	r := rand.New(rand.NewPCG(6, 6))
	point := func() planner.Point {
		return planner.Point{Latitude: r.Float64()*180 - 90, Longitude: r.Float64()*360 - 180}
	}

	// Journeys of one to four legs: intermediate destinations (#2) are what
	// give a journey more than one, and every one of them is a driving leg.
	var legs [][2]planner.Point

	for range 20 {
		req := planner.Request{Origin: point(), Destination: point()}
		for range r.IntN(4) {
			req.Intermediates = append(req.Intermediates, point())
		}

		journey, err := p.Plan(t.Context(), req)
		if err != nil {
			t.Fatalf("Plan(%+v): %v", req, err)
		}

		if len(journey.Legs) != len(req.Intermediates)+1 {
			t.Fatalf("Plan(%+v) returned %d legs, want %d", req, len(journey.Legs), len(req.Intermediates)+1)
		}

		for i := range journey.Legs {
			legs = append(legs, [2]planner.Point{journey.Locations[i], journey.Locations[i+1]})
		}
	}

	engine.mu.Lock()
	defer engine.mu.Unlock()

	if len(engine.asked) != len(legs) {
		t.Fatalf("the engine was asked %d times for %d legs, want once a leg", len(engine.asked), len(legs))
	}

	for i, got := range engine.asked {
		locs, _ := got.Body["locations"].([]any)
		if len(locs) != 2 {
			t.Fatalf("request %d named %d locations, want the leg's 2", i, len(locs))
		}

		for j, want := range legs[i] {
			m, _ := locs[j].(map[string]any)
			if m["lat"] != want.Latitude || m["lon"] != want.Longitude {
				t.Errorf("request %d location %d was %v, want leg %d's %v", i, j, m, i, want)
			}
		}
	}

	for i, got := range engine.asked {
		if got.Method != http.MethodPost || got.Path != "/route" {
			t.Errorf("request %d was %s %s, want POST /route", i, got.Method, got.Path)
		}

		if got.Body["costing"] != "auto" {
			t.Errorf("request %d asked for the costing %v, want auto: the leg was not routed for a car", i, got.Body["costing"])
		}

		for _, key := range []string{"costing_options", "exclude_locations", "exclude_polygons", "linear_cost_factors"} {
			if v, ok := got.Body[key]; ok {
				t.Errorf("request %d carries %s = %v, want none: it alters what the car costing may use", i, key, v)
			}
		}
	}
}

// TestTheLegIsAskedBetweenItsLocationsInOrder: the two locations reach the
// engine as `lat` and `lon`, origin first — the axis order this project has
// shipped wrong once, held here by name.
func TestTheLegIsAskedBetweenItsLocationsInOrder(t *testing.T) {
	t.Parallel()

	engine, c := startEngine(t, straightByCar)

	from := planner.Point{Latitude: 57.7036, Longitude: 11.9611}
	to := planner.Point{Latitude: 57.7052, Longitude: 11.9714}

	shape, err := c.Route(t.Context(), from, to)
	if err != nil {
		t.Fatalf("Route: %v", err)
	}

	locs, _ := engine.asked[0].Body["locations"].([]any)
	if len(locs) != 2 {
		t.Fatalf("the request named %d locations, want 2", len(locs))
	}

	for i, want := range []planner.Point{from, to} {
		m, _ := locs[i].(map[string]any)
		if m["lat"] != want.Latitude || m["lon"] != want.Longitude {
			t.Errorf("location %d was sent as %v, want lat %v and lon %v", i, m, want.Latitude, want.Longitude)
		}
	}

	if len(shape) != 2 || shape[0] != from || shape[1] != to {
		t.Errorf("the route's shape is %v, want the leg the engine answered, %v to %v", shape, from, to)
	}
}

// TestALegNotDrivenByCarIsRefused is `FR-006` AC2 at the answer: a leg the
// engine reports as travelled any way but by car — on foot, by bicycle, by
// another vehicle, in part or with nothing saying how — is refused with
// ErrNotDriven and never returned as a route.
func TestALegNotDrivenByCarIsRefused(t *testing.T) {
	t.Parallel()

	shape := encodeShape([]planner.Point{{Latitude: 57.70, Longitude: 11.96}, {Latitude: 57.71, Longitude: 11.97}})

	for name, maneuvers := range map[string][]maneuver{
		"on foot":            {{TravelMode: "pedestrian", TravelType: "foot"}},
		"by bicycle":         {{TravelMode: "bicycle", TravelType: "hybrid"}},
		"driven then walked": {byCar, {TravelMode: "pedestrian", TravelType: "foot"}, byCar},
		"driven by truck":    {{TravelMode: "drive", TravelType: "truck"}},
		"no travel mode":     {{}},
		"no maneuvers":       {},
	} {
		_, c := startEngine(t, func(map[string]any) (int, string) {
			return http.StatusOK, answer(leg{Shape: shape, Maneuvers: maneuvers})
		})

		got, err := c.Route(t.Context(), planner.Point{Latitude: 57.70, Longitude: 11.96}, planner.Point{Latitude: 57.71, Longitude: 11.97})
		if !errors.Is(err, ErrNotDriven) {
			t.Errorf("a leg travelled %s gave %v, %v; want an error wrapping ErrNotDriven", name, got, err)
		}

		if got != nil {
			t.Errorf("a leg travelled %s was returned as a route, %v", name, got)
		}
	}
}

// TestAnAnswerThatIsNotOneLegIsRefused: an answer that does not hold exactly
// the one leg two locations make, or whose leg has no usable shape, is an
// error and never an empty or partial route.
func TestAnAnswerThatIsNotOneLegIsRefused(t *testing.T) {
	t.Parallel()

	good := encodeShape([]planner.Point{{Latitude: 57.70, Longitude: 11.96}, {Latitude: 57.71, Longitude: 11.97}})

	for name, body := range map[string]string{
		"not JSON":    `<html>`,
		"no trip":     `{}`,
		"no legs":     answer(),
		"two legs":    answer(leg{Shape: good, Maneuvers: []maneuver{byCar}}, leg{Shape: good, Maneuvers: []maneuver{byCar}}),
		"no shape":    `{"trip":{"legs":[{"maneuvers":[{"travel_mode":"drive","travel_type":"car"}]}]}}`,
		"empty shape": answer(leg{Shape: "", Maneuvers: []maneuver{byCar}}),
		"bad shape":   answer(leg{Shape: "_p~iF~ps|U_", Maneuvers: []maneuver{byCar}}),
	} {
		_, c := startEngine(t, func(map[string]any) (int, string) { return http.StatusOK, body })

		got, err := c.Route(t.Context(), planner.Point{Latitude: 57.70, Longitude: 11.96}, planner.Point{Latitude: 57.71, Longitude: 11.97})
		if err == nil || got != nil {
			t.Errorf("an answer with %s gave %v, %v; want an error and no route", name, got, err)
		}
	}
}

// TestAnEngineErrorIsAnError: an answer other than 200 is an error carrying
// the engine's own code and message where its body has them, and a status
// alone where it does not.
func TestAnEngineErrorIsAnError(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		status int
		body   string
		want   []string
	}{
		{http.StatusBadRequest, `{"error_code":442,"status":"Bad Request","status_code":400,"error":"No path could be found for input"}`, []string{"400", "442", "No path could be found for input"}},
		{http.StatusBadRequest, `{"error_code":171,"status":"Bad Request","status_code":400,"error":"No suitable edges near location"}`, []string{"171", "No suitable edges near location"}},
		{http.StatusInternalServerError, `oops`, []string{"500"}},
	} {
		_, c := startEngine(t, func(map[string]any) (int, string) { return tc.status, tc.body })

		got, err := c.Route(t.Context(), planner.Point{Latitude: 57.70, Longitude: 11.96}, planner.Point{Latitude: 57.71, Longitude: 11.97})
		if err == nil || got != nil {
			t.Errorf("an answer of %d gave %v, %v; want an error and no route", tc.status, got, err)

			continue
		}

		for _, w := range tc.want {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("the error %q for an answer of %d does not carry %q", err, tc.status, w)
			}
		}
	}
}

// TestAnOversizedAnswerIsRefused: an answer past maxResponseBytes is refused
// by an error rather than held in memory whole.
func TestAnOversizedAnswerIsRefused(t *testing.T) {
	t.Parallel()

	_, c := startEngine(t, func(map[string]any) (int, string) {
		return http.StatusOK, strings.Repeat(" ", maxResponseBytes+1)
	})

	_, err := c.Route(t.Context(), planner.Point{Latitude: 57.70, Longitude: 11.96}, planner.Point{Latitude: 57.71, Longitude: 11.97})
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("an answer of %d bytes gave %v, want a refusal for its size", maxResponseBytes+1, err)
	}
}

// TestARedirectIsNotFollowed: an engine that redirects is misconfigured, and
// the locations are not sent on to wherever it points.
func TestARedirectIsNotFollowed(t *testing.T) {
	t.Parallel()

	elsewhere := &fakeEngine{respond: straightByCar}

	target := httptest.NewServer(elsewhere)
	t.Cleanup(target.Close)

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/route", http.StatusTemporaryRedirect)
	}))
	t.Cleanup(redirector.Close)

	c, err := New(redirector.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if _, err := c.Route(t.Context(), planner.Point{Latitude: 57.70, Longitude: 11.96}, planner.Point{Latitude: 57.71, Longitude: 11.97}); err == nil {
		t.Error("a redirect was answered as a route, want an error")
	}

	elsewhere.mu.Lock()
	defer elsewhere.mu.Unlock()

	if len(elsewhere.asked) != 0 {
		t.Errorf("the redirect was followed: its target was asked %d times", len(elsewhere.asked))
	}
}

// TestARouteEndsWithItsTimeoutOrItsCaller: an engine that does not answer
// holds a route no longer than the timeout, and no longer than the caller's
// own context either.
func TestARouteEndsWithItsTimeoutOrItsCaller(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))

	// Cleanups run last first: release any handler still waiting, then close.
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(release) })

	from, to := planner.Point{Latitude: 57.70, Longitude: 11.96}, planner.Point{Latitude: 57.71, Longitude: 11.97}

	c, err := New(srv.URL, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	start := time.Now()
	if _, err := c.Route(t.Context(), from, to); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("a silent engine gave %v, want the timeout's deadline", err)
	}

	if took := time.Since(start); took > 5*time.Second {
		t.Errorf("a route with a 100ms timeout took %v", took)
	}

	patient, err := New(srv.URL, time.Hour)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	time.AfterFunc(100*time.Millisecond, cancel)

	if _, err := patient.Route(ctx, from, to); !errors.Is(err, context.Canceled) {
		t.Errorf("a cancelled caller gave %v, want its cancellation", err)
	}
}

// TestTheRoutePathIsJoinedOntoTheBaseURL: the route API is the `route` path
// under the configured base, with or without a trailing slash and under a
// prefix.
func TestTheRoutePathIsJoinedOntoTheBaseURL(t *testing.T) {
	t.Parallel()

	for base, want := range map[string]string{
		"http://127.0.0.1:8002":            "http://127.0.0.1:8002/route",
		"http://127.0.0.1:8002/":           "http://127.0.0.1:8002/route",
		"https://routing.example/valhalla": "https://routing.example/valhalla/route",
	} {
		c, err := New(base, time.Second)
		if err != nil {
			t.Errorf("New(%q): %v", base, err)

			continue
		}

		if c.routeURL != want {
			t.Errorf("New(%q) routes at %q, want %q", base, c.routeURL, want)
		}
	}

	for _, timeout := range []time.Duration{0, -time.Second} {
		if _, err := New("http://127.0.0.1:8002", timeout); err == nil {
			t.Errorf("New accepted a timeout of %v, want a refusal", timeout)
		}
	}

	if _, err := New("http://127.0.0.1:%zz", time.Second); err == nil {
		t.Error("New accepted a URL that does not parse, want a refusal")
	}
}

// TestTheShapeIsDecodedAtSixDigits holds decodeShape to two published
// vectors. The first is the Google polyline algorithm page's own example
// (developers.google.com/maps/documentation/utilities/polylinealgorithm),
// read at Valhalla's six digits rather than Google's five, so each coordinate
// is a tenth of the page's. The second is the start of a shape a Valhalla
// 3.9.1 built from this package's extract returned for the integration test's
// car leg.
func TestTheShapeIsDecodedAtSixDigits(t *testing.T) {
	t.Parallel()

	for encoded, want := range map[string][]planner.Point{
		"_p~iF~ps|U_ulLnnqC_mqNvxq`@": {
			{Latitude: 3.85, Longitude: -12.02}, {Latitude: 4.07, Longitude: -12.095}, {Latitude: 4.3252, Longitude: -12.6453},
		},
		"_e}`mBqs`yUOeA]gD": {
			{Latitude: 57.70352, Longitude: 11.961161}, {Latitude: 57.703528, Longitude: 11.961196}, {Latitude: 57.703543, Longitude: 11.96128},
		},
	} {
		got, err := decodeShape(encoded)
		if err != nil {
			t.Errorf("decodeShape(%q): %v", encoded, err)

			continue
		}

		if len(got) != len(want) {
			t.Errorf("decodeShape(%q) = %v, want %v", encoded, got, want)

			continue
		}

		for i := range want {
			if math.Abs(got[i].Latitude-want[i].Latitude) > 1e-9 || math.Abs(got[i].Longitude-want[i].Longitude) > 1e-9 {
				t.Errorf("decodeShape(%q) vertex %d = %v, want %v", encoded, i, got[i], want[i])
			}
		}
	}
}

// TestEveryShapeRoundTrips: whatever coordinates a shape encodes, at six
// digits, decodeShape returns them.
func TestEveryShapeRoundTrips(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(6, 1))

	for c := range 500 {
		n := 1 + r.IntN(20)
		points := make([]planner.Point, n)

		for i := range points {
			points[i] = planner.Point{
				Latitude:  math.Round((r.Float64()*180-90)*shapePrecision) / shapePrecision,
				Longitude: math.Round((r.Float64()*360-180)*shapePrecision) / shapePrecision,
			}
		}

		points[0] = planner.Point{Latitude: []float64{-90, 90, 0}[c%3], Longitude: []float64{-180, 180, 0}[c%3]}

		got, err := decodeShape(encodeShape(points))
		if err != nil {
			t.Fatalf("case %d: decodeShape: %v", c, err)
		}

		if len(got) != n {
			t.Fatalf("case %d: %d vertices became %d", c, n, len(got))
		}

		for i := range points {
			if math.Abs(got[i].Latitude-points[i].Latitude) > 1e-9 || math.Abs(got[i].Longitude-points[i].Longitude) > 1e-9 {
				t.Fatalf("case %d vertex %d: %v became %v", c, i, points[i], got[i])
			}
		}
	}
}

// TestAMalformedShapeIsRefused: a shape that ends inside a value, leaves the
// encoding's alphabet, runs a value past any coordinate's length, or decodes
// off the Earth is an error, never a vertex.
func TestAMalformedShapeIsRefused(t *testing.T) {
	t.Parallel()

	for name, encoded := range map[string]string{
		"a latitude with no longitude": "_p~iF",
		"a value cut short":            "_p~iF~ps|U_",
		"below the alphabet":           "_p~iF ~ps|U",
		"above the alphabet":           "_p~iF\x7f~ps|U",
		"a value past seven chunks":    strings.Repeat("~", 8) + "?",
		"off the Earth":                encodeShape([]planner.Point{{Latitude: 89, Longitude: 0}}) + encodeShape([]planner.Point{{Latitude: 2, Longitude: 0}}),
	} {
		if got, err := decodeShape(encoded); err == nil {
			t.Errorf("a shape with %s decoded to %v, want an error", name, got)
		}
	}
}

// TestTheRequestBodyIsTheDocumentedOne holds the whole request body, so that
// a field added to it is a visible diff in a test and not only in the adapter.
func TestTheRequestBodyIsTheDocumentedOne(t *testing.T) {
	t.Parallel()

	b, err := json.Marshal(routeRequest{
		Locations:      []location{{Lat: 57.7036, Lon: 11.9611}, {Lat: 57.7052, Lon: 11.9714}},
		Costing:        costingCar,
		DirectionsType: directionsManeuvers,
	})
	if err != nil {
		t.Fatal(err)
	}

	const want = `{"locations":[{"lat":57.7036,"lon":11.9611},{"lat":57.7052,"lon":11.9714}],"costing":"auto","directions_type":"maneuvers"}`
	if string(b) != want {
		t.Errorf("the route request is\n%s\nwant\n%s", b, want)
	}
}
