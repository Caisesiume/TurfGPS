package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// endsBody is a journey from Stockholm to Gothenburg. straightRouting places
// both of its ends at their midpoint, some 200 km from either location.
const endsBody = `{"origin":{"latitude":59.3293,"longitude":18.0686},"destination":{"latitude":57.7089,"longitude":11.9746}}`

// endsCoordinates are the request's coordinates, and the midpoint the leg's
// ends are placed at, as a log line would spell them: none may appear in one.
var endsCoordinates = []string{"59.3293", "18.0686", "57.7089", "11.9746", "58.5191", "15.0216"}

// logInto returns a logger writing JSON lines into buf.
func logInto(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// logLines returns each JSON line in buf as its fields.
func logLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	var lines []map[string]any

	for line := range strings.Lines(buf.String()) {
		var fields map[string]any
		if err := json.Unmarshal([]byte(line), &fields); err != nil {
			t.Fatalf("a log line is not JSON: %v; line %q", err, line)
		}

		lines = append(lines, fields)
	}

	return lines
}

// TestALegRefusedForItsEndIsLoggedWithoutALocation (#250): a leg whose end
// the provider placed beyond the end tolerance is answered 502 and logged by
// which leg, which end, its distance from its location and the tolerance,
// which is what the deployment measures the tolerance from (`DEPLOYMENT.md`);
// the line carries those fields and nothing else, so no coordinate of the
// request or the leg.
func TestALegRefusedForItsEndIsLoggedWithoutALocation(t *testing.T) {
	t.Parallel()

	const tolerance = 1000.0

	p, err := planner.New(straightRouting{}, zoneCopy{held: true}, clearAccess{}, clearAccess{}, planner.EndsWithin(tolerance))
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	rec := postJourney(t, endsBody, WithPlanner(p), WithLogger(logInto(&buf)))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d; body %q", rec.Code, http.StatusBadGateway, rec.Body.String())
	}

	lines := logLines(t, &buf)
	if len(lines) != 1 {
		t.Fatalf("the refusal was logged in %d line(s), want 1: %q", len(lines), buf.String())
	}

	line := lines[0]

	wantKeys := []string{"distance_metres", "end", "leg", "level", "msg", "time", "tolerance_metres"}
	if got := slices.Sorted(maps.Keys(line)); !slices.Equal(got, wantKeys) {
		t.Errorf("the refusal's log line carries %v, want exactly %v: %q", got, wantKeys, buf.String())
	}

	if line["level"] != "WARN" || line["leg"] != 0.0 || line["end"] != "first vertex" || line["tolerance_metres"] != tolerance {
		t.Errorf("the refusal was logged as %v; want a WARN line naming leg 0, its first vertex and the tolerance of %v m", line, tolerance)
	}

	// The midpoint lies some 200 km from Stockholm; the bound overstates the
	// geodesic by at most about 1%.
	if d, ok := line["distance_metres"].(float64); !ok || d < 150e3 || d > 250e3 {
		t.Errorf("the refusal logged a distance of %v m, want the first vertex's distance from the origin, some 200 km", line["distance_metres"])
	}

	for _, coordinate := range endsCoordinates {
		if strings.Contains(buf.String(), coordinate) {
			t.Errorf("the refusal's log line holds the coordinate %s: %q", coordinate, buf.String())
		}
	}
}

// TestAJourneyIsUnavailableWhileNoEndToleranceIsSet (#250): while no end
// tolerance is configured no leg is served, and `POST /journeys` answers 503,
// as it does for the deployment's other unconfigured states, and logs nothing
// per request: the service logs it at start.
func TestAJourneyIsUnavailableWhileNoEndToleranceIsSet(t *testing.T) {
	t.Parallel()

	p, err := planner.New(straightRouting{}, zoneCopy{held: true}, clearAccess{}, clearAccess{})
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	rec := postJourney(t, endsBody, WithPlanner(p), WithLogger(logInto(&buf)))
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "no end tolerance is configured") {
		t.Errorf("status = %d, body %q; want %d saying no end tolerance is configured", rec.Code, rec.Body.String(), http.StatusServiceUnavailable)
	}

	if buf.Len() != 0 {
		t.Errorf("an unset tolerance was logged per request: %q", buf.String())
	}
}

// TestNoOtherRefusalIsLogged (#250): the end refusal is the only one the
// request surface logs; a provider's failure is answered without a word.
func TestNoOtherRefusalIsLogged(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	rec := postJourney(t, endsBody, withFakePlanner(t, failingRouting{}), WithLogger(logInto(&buf)))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d; body %q", rec.Code, http.StatusBadGateway, rec.Body.String())
	}

	if buf.Len() != 0 {
		t.Errorf("a provider's failure was logged: %q", buf.String())
	}
}
