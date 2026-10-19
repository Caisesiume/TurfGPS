package zonestore

import (
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/corridor"
)

// TestARouteIsWrittenLongitudeFirst pins the one place a route's axis order is
// written. Its database-side counterpart, which also catches a swap here by the
// zones a swapped corridor would return, is in corridor_postgis_test.go.
func TestARouteIsWrittenLongitudeFirst(t *testing.T) {
	t.Parallel()

	route := []corridor.Point{
		{Latitude: 59.33, Longitude: 18.02},
		{Latitude: -33.8688, Longitude: 151.2093},
		// The shortest round-tripping decimal, in plain notation: no exponent,
		// no digits lost.
		{Latitude: 0.000012, Longitude: -0.1234567890123},
	}

	const want = "SRID=4326;LINESTRING(18.02 59.33,151.2093 -33.8688,-0.1234567890123 0.000012)"

	if got := routeEWKT(route); got != want {
		t.Errorf("routeEWKT wrote %q, want %q: WKT is longitude first", got, want)
	}
}
