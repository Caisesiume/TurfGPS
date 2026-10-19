package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// EnvCORSAllowedOrigins names the browser origins the service grants
// cross-origin access to, comma-separated (#224).
//
// The client is deployed on Vercel and calls the API on the API's own
// subdomain, never through a Vercel rewrite (`docs/adr/ADR-0006-deployment-topology.md`
// § Decision, D4), so every call it makes is cross-origin, and this is the one
// place the origins it is served from are named.
//
// IT HAS NO DEFAULT, for the reason the package's other required values have
// none: an origin compiled in here would be a second home for the deployment's
// address, and one that goes on granting access after the client has moved.
// Unset or empty, the service grants cross-origin access to nothing, which is
// how it starts and serves with nothing else configured (`NFR-003`).
const EnvCORSAllowedOrigins = "TURFGPS_CORS_ALLOWED_ORIGINS"

// LoadAllowedOrigins reads the origins the service grants cross-origin access
// to, and returns them in the exact form a browser sends in `Origin`.
//
// It returns (nil, nil) when the variable is unset or empty. Every entry it
// cannot hold to that form is a refusal naming the variable, never a value
// quietly corrected: an origin rewritten here is an origin the operator
// believes they set, and one that is merely mistyped is not refused by the
// browser either — it simply never matches, so the client fails with nothing
// on the service's side saying why.
//
// A wildcard is refused in any position. `*` would send
// `Access-Control-Allow-Origin: *`, granting every site on the web a read of
// every response, and a pattern such as `https://*.vercel.app` grants every
// other tenant of the same platform. Neither is an origin, so neither is
// accepted as one.
func LoadAllowedOrigins(lookup LookupFunc) ([]string, error) {
	raw, ok := lookup(EnvCORSAllowedOrigins)
	if !ok || strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	var origins []string

	for entry := range strings.SplitSeq(raw, ",") {
		origin := strings.TrimSpace(entry)

		if err := validateOrigin(origin); err != nil {
			return nil, fmt.Errorf("%s: %w", EnvCORSAllowedOrigins, err)
		}

		origins = append(origins, origin)
	}

	return origins, nil
}

// validateOrigin holds one entry to the serialization of an https origin —
// scheme, host and an optional non-default port, lowercase, and nothing else —
// because the browser compares nothing: the grant is an exact string match on
// the `Origin` header, and that header is always in this form.
//
// HTTPS IS REQUIRED. The client is served from Vercel over https, and an http
// origin admitted here grants a read of the API's responses to any page an
// observer on the network path can inject into that origin.
func validateOrigin(origin string) error {
	if origin == "" {
		return fmt.Errorf("an entry is empty: list each origin once, separated by commas")
	}

	if strings.Contains(origin, "*") {
		return fmt.Errorf("%q is a wildcard: name each origin exactly, because a wildcard grants every site it matches", origin)
	}

	u, err := url.Parse(origin)
	if err != nil {
		return fmt.Errorf("%q is not an origin: %w", origin, err)
	}

	if u.Scheme != "https" {
		return fmt.Errorf("%q has scheme %q, want https", origin, u.Scheme)
	}

	if u.Hostname() == "" {
		return fmt.Errorf("%q names no host", origin)
	}

	canonical := "https://" + strings.ToLower(u.Hostname())
	if strings.Contains(u.Hostname(), ":") {
		canonical = "https://[" + strings.ToLower(u.Hostname()) + "]"
	}

	if port := u.Port(); port != "" && port != "443" {
		canonical = "https://" + net.JoinHostPort(strings.ToLower(u.Hostname()), port)
	}

	if origin != canonical {
		return fmt.Errorf("%q is not in the form a browser sends: write it as %q — scheme, host and a non-default port, lowercase, with no path, query or trailing slash", origin, canonical)
	}

	return nil
}
