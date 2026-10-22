package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// THE WIRE SHAPE BELOW IS PROVISIONAL (#1). No document defines the journey
// API's route or payload: `Architecture.md § The call budget` draws the request
// as origin, destination and budget and fixes no form. This is the smallest
// shape `FR-001` and `FR-002` need — an origin, a destination and an optional
// ordered list of intermediate destinations in (#2), the journey out — and the
// story that designs the API replaces it rather than extends it.
//
// Locations are coordinates. Search, pins and zone names (`DESIGN.md
// § Locations`) resolve to one before planning, through the `Geocoder` port,
// which is not built.

// maxJourneyRequestBytes bounds a planning request's body. The payload is a
// coordinate per required location, a few hundred bytes for the journeys
// `SPECIFICATION.md § Primary use case` describes; the bound only stops a client
// holding a handler on an arbitrarily long body, so it is set well above the
// payload — it admits hundreds of intermediate destinations — and
// carries no meaning beyond that. No document caps how many intermediate
// destinations a journey may name, and this is not that cap.
const maxJourneyRequestBytes = 16 << 10

type wirePoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type journeyRequest struct {
	// Pointers, so that an absent location is told apart from the coordinate
	// (0, 0), which is a real place in the Gulf of Guinea.
	Origin *wireLocation `json:"origin"`

	// Intermediates are the intermediate destinations in the order the user
	// entered them, which is the order they are visited in (`FR-003`). Absent,
	// null and empty all name none. A null element decodes with neither axis,
	// so it is refused like any location missing one.
	Intermediates []wireLocation `json:"intermediates"`

	Destination *wireLocation `json:"destination"`
}

// wireLocation is a location as a request names it. Its axes are pointers for
// the same reason: an absent or null latitude or longitude would otherwise
// decode as 0 and be planned from, a coordinate nobody entered.
type wireLocation struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

// point returns the location as a planner point, and false when it lacks an
// axis.
func (l *wireLocation) point() (planner.Point, bool) {
	if l.Latitude == nil || l.Longitude == nil {
		return planner.Point{}, false
	}

	return planner.Point{Latitude: *l.Latitude, Longitude: *l.Longitude}, true
}

type journeyResponse struct {
	Locations []wirePoint `json:"locations"`
	Legs      []wireLeg   `json:"legs"`
}

type wireLeg struct {
	Shape []wirePoint `json:"shape"`
}

// planJourney answers `POST /journeys`: it plans the journey the request names
// and returns it.
//
// NOTHING HERE LOGS, for the reason `NewHandler` gives and one more: the
// request carries the user's origin, which is frequently their home
// (`Architecture.md § Personal data`), and a routing error may quote it. Each
// failure is answered with a status and a fixed message, never the error.
func planJourney(plans *planner.Planner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if plans == nil {
			http.Error(w, "journey planning is unavailable: no routing provider is configured", http.StatusServiceUnavailable)

			return
		}

		req, ok := decodeJourneyRequest(w, r)
		if !ok {
			return
		}

		journey, err := plans.Plan(r.Context(), req)

		switch {
		case errors.Is(err, planner.ErrNotACoordinate):
			http.Error(w, "a location is not a coordinate", http.StatusBadRequest)

			return
		case err != nil:
			http.Error(w, "the journey could not be routed", http.StatusBadGateway)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(encodeJourney(journey))
	}
}

// decodeJourneyRequest reads the request body, answering the request itself and
// returning false when the body does not name an origin and a destination, or
// names a location without both of its axes.
//
// An unknown field is refused rather than ignored. A client that sends one is
// asking for something this shape cannot express — or naming intermediate
// destinations under another name — and planning the journey without it would
// answer a question it did not ask.
func decodeJourneyRequest(w http.ResponseWriter, r *http.Request) (planner.Request, bool) {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJourneyRequestBytes))
	dec.DisallowUnknownFields()

	var body journeyRequest

	if err := dec.Decode(&body); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "the request body is too large", http.StatusRequestEntityTooLarge)

			return planner.Request{}, false
		}

		http.Error(w, "the request body is not a journey request", http.StatusBadRequest)

		return planner.Request{}, false
	}

	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		http.Error(w, "the request body holds more than one journey request", http.StatusBadRequest)

		return planner.Request{}, false
	}

	if body.Origin == nil || body.Destination == nil {
		http.Error(w, "a journey request names an origin and a destination", http.StatusBadRequest)

		return planner.Request{}, false
	}

	origin, hasOrigin := body.Origin.point()
	destination, hasDestination := body.Destination.point()
	intermediates := make([]planner.Point, len(body.Intermediates))
	hasIntermediates := true

	for i := range body.Intermediates {
		var ok bool

		intermediates[i], ok = body.Intermediates[i].point()
		hasIntermediates = hasIntermediates && ok
	}

	if !hasOrigin || !hasDestination || !hasIntermediates {
		http.Error(w, "a location names a latitude and a longitude", http.StatusBadRequest)

		return planner.Request{}, false
	}

	return planner.Request{Origin: origin, Intermediates: intermediates, Destination: destination}, true
}

func encodeJourney(j planner.Journey) journeyResponse {
	out := journeyResponse{Locations: encodePoints(j.Locations), Legs: make([]wireLeg, len(j.Legs))}

	for i, leg := range j.Legs {
		out.Legs[i] = wireLeg{Shape: encodePoints(leg.Shape)}
	}

	return out
}

func encodePoints(points []planner.Point) []wirePoint {
	out := make([]wirePoint, len(points))

	for i, p := range points {
		out[i] = wirePoint(p)
	}

	return out
}
