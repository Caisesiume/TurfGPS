package config

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// The routing engine's configuration (#6), a block of its own: journeys are
// routed through Valhalla (`Architecture.md § D3`), and a service with no
// engine configured starts anyway and answers `POST /journeys` with 503, as
// `NFR-003` permits.
//
// THE URL HAS NO DEFAULT. Valhalla runs beside the service and binds loopback
// (`DEPLOYMENT.md § What is deployed`), but the port it listens on is the
// deployment's arrangement, which `DEPLOYMENT.md § Where the deployment
// configuration lives` keeps outside the repository; an address compiled in
// here would be a second home for it.
//
// Nothing here chooses how a route is computed. The travel mode is the
// routing port's contract (`internal/planner`), and the request that carries
// it is the adapter's (`internal/valhalla`); this block only says where the
// engine is and how long one route may take.
const (
	EnvValhallaURL     = "TURFGPS_VALHALLA_URL"
	EnvValhallaTimeout = "TURFGPS_VALHALLA_TIMEOUT"
)

// defaultValhallaTimeout bounds one route request end to end. It is this
// service's operational choice, not a measurement: it exists to make a call to
// an engine that has stopped answering finite, well inside the request
// surface's own write budget, and it restates no figure from any document.
const defaultValhallaTimeout = 10 * time.Second

// Valhalla is the configuration the routing adapter needs.
type Valhalla struct {
	// URL is the engine's base URL. The route API is the `route` path under
	// it.
	URL string

	// Timeout bounds one route request.
	Timeout time.Duration
}

// LoadValhalla reads the routing engine's configuration.
//
// It returns (nil, nil) when neither variable is set, which is a service
// running without journey planning. A timeout set without a URL is refused as
// a partial configuration rather than ignored, so that an operator who meant
// to configure the engine does not start a service that answers 503 with
// nothing at start saying why.
//
// A URL that is not an http or https URL naming a host, or that carries a
// query or a fragment, is refused: the route API's path is joined onto it, and
// a query or fragment would be silently dropped or misplaced by that join.
func LoadValhalla(lookup LookupFunc) (*Valhalla, error) {
	raw, ok := lookup(EnvValhallaURL)
	if !ok || raw == "" {
		if v, set := lookup(EnvValhallaTimeout); set && v != "" {
			return nil, fmt.Errorf("the routing engine is partially configured: %s is set and %s is not — set %s to route journeys, or neither to run without",
				EnvValhallaTimeout, EnvValhallaURL, EnvValhallaURL)
		}

		return nil, nil
	}

	if err := validateEngineURL(raw); err != nil {
		return nil, fmt.Errorf("%s: %w", EnvValhallaURL, err)
	}

	timeout, err := positiveDuration(lookup, EnvValhallaTimeout, defaultValhallaTimeout)
	if err != nil {
		return nil, err
	}

	return &Valhalla{URL: raw, Timeout: timeout}, nil
}

// validateEngineURL holds a routing engine's base URL to the form described on
// LoadValhalla. Plain http is accepted, unlike validateEndpoint's: the engine
// is the deployment's own, bound to loopback beside the service, and not a
// third party reached across a network.
//
// No refusal quotes the URL as given. main logs the refusal, and a URL may
// carry a password, so each names the URL in its redacted form — and a parse
// failure, whose error quotes the input whole, by its cause alone.
func validateEngineURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		var parseErr *url.Error
		if errors.As(err, &parseErr) {
			err = parseErr.Err
		}

		return fmt.Errorf("not a URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%q has the scheme %q, want http or https", u.Redacted(), u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("%q names no host", u.Redacted())
	}

	if u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return fmt.Errorf("%q carries a query or a fragment, want a base URL the route path can be joined onto", u.Redacted())
	}

	return nil
}
