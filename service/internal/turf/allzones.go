// Package turf is the adapter for the Turf API — the TurfClient port of
// `Architecture.md § Ports and adapters`.
//
// It carries the zone sync's half of that port and nothing else yet.
//
// NO API PATH IS WRITTEN IN THIS PACKAGE. The all-zones endpoint is named by
// role in `Architecture.md § Retrieving zones` and arrives here as a resolved
// URL from configuration, so this module is not a second home for the API
// version — binding the version by citation is `FR-019`'s work, and a versioned
// path compiled in here would quietly pre-empt it.
package turf

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// errorBodySnippet is how much of an unusable response is quoted into the error.
// Enough for an API error object to be readable, short enough that a 43 MB body
// arriving under a wrong status does not become a log line.
const errorBodySnippet = 512

// Client fetches from the Turf API.
type Client struct {
	allZonesURL string
	http        *http.Client
	maxBytes    int64
}

// NewClient builds the adapter.
//
// The HTTP client deliberately carries NO timeout of its own. Every call takes a
// context whose deadline the caller sets, and a second bound here would be a
// budget nothing in the caller could see — the sync's fetch budget is
// configuration, and this is where it is obeyed rather than where it is decided.
func NewClient(allZonesURL string, maxBytes int64) (*Client, error) {
	if allZonesURL == "" {
		return nil, errors.New("no all-zones endpoint")
	}

	if maxBytes <= 0 {
		return nil, fmt.Errorf("the response ceiling %d is not positive", maxBytes)
	}

	return &Client{
		allZonesURL: allZonesURL,
		http:        &http.Client{},
		maxBytes:    maxBytes,
	}, nil
}

// WithHTTPClient substitutes the HTTP client, for a test that serves the
// endpoint itself.
func (c *Client) WithHTTPClient(h *http.Client) *Client {
	c.http = h

	return c
}

// FetchAllZones fetches the complete zone set.
//
// It returns the body, the HTTP status, and an error. The status is zero when no
// response arrived at all, which is what lets a refused request and a body that
// would not parse be told apart once both are recorded on the run — the
// distinction `Architecture.md § The sync write path` asks http_status and
// response_bytes to carry.
//
// It has the shape of zonesync.FetchFunc without naming it. The port is declared
// by its consumer, so this adapter does not import the package it serves.
func (c *Client) FetchAllZones(ctx context.Context) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.allZonesURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("building the all-zones request: %w", err)
	}

	// Safe on a data endpoint. `Architecture.md § API version` records that the
	// self-documenting index answers 406 to a JSON Accept header while every
	// data endpoint under it succeeds, which is a reason not to set this
	// globally across the base URL rather than a reason not to set it here.
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "TurfGPS")

	// Compression is deliberately not requested by hand: net/http negotiates
	// gzip and decompresses transparently exactly when the caller has not set
	// Accept-Encoding itself, and setting it here would hand back a compressed
	// body this function would then have to decode.
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("the all-zones request failed: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := c.read(resp.Body)
	if err != nil {
		return body, resp.StatusCode, err
	}

	if resp.StatusCode != http.StatusOK {
		return body, resp.StatusCode, fmt.Errorf("the all-zones endpoint answered %s: %s", resp.Status, snippet(body))
	}

	return body, resp.StatusCode, nil
}

// read takes the body up to the ceiling, and refuses rather than truncates when
// it is exceeded.
//
// TRUNCATING WOULD BE THE WORST AVAILABLE ANSWER. A body cut at the ceiling is
// either unparseable — the ordinary case, and merely noisy — or, for a shape
// that happens to close cleanly, a short response that looks complete. That is
// exactly the truncated response the staging assertions of `Architecture.md §
// The sync write path` exist to catch, and manufacturing one here would be this
// service producing the failure it is defending against.
func (c *Client) read(r io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r, c.maxBytes+1))
	if err != nil {
		return body, fmt.Errorf("reading the all-zones response: %w", err)
	}

	if int64(len(body)) > c.maxBytes {
		return nil, fmt.Errorf("the all-zones response exceeds the %d byte ceiling and was not read", c.maxBytes)
	}

	return body, nil
}

func snippet(body []byte) string {
	if len(body) > errorBodySnippet {
		return string(body[:errorBodySnippet]) + "…"
	}

	return string(body)
}
