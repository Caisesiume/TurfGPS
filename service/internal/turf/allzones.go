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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// errorBodySnippet is how much of an unusable response is quoted into the error.
// Enough for an API error object to be readable, short enough that a
// full-corpus body arriving under a wrong status does not become a log line.
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

// setHTTPClient substitutes the HTTP client, for a test that serves the
// endpoint itself.
//
// UNEXPORTED, AND IT RETURNS NOTHING. As an exported method returning *Client it
// read as a builder — the shape that invites NewClient(...).WithHTTPClient(...)
// — while doing the opposite of one: it mutates the receiver in place, so a
// caller that kept the value it was chaining from would find that value changed
// underneath it. It was never a construction option, and its only caller is this
// package's own test serving the endpoint from an httptest server, which is an
// argument for keeping it out of the package's API rather than in it.
func (c *Client) setHTTPClient(h *http.Client) {
	c.http = h
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

	body, err := c.read(resp.Body, resp.ContentLength)
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
//
// IT IS SIZED FROM THE RESPONSE RATHER THAN GROWN INTO. io.ReadAll starts at 512
// bytes and grows by allocating twice the current capacity and copying, holding
// both for the length of the copy — so reading a body this way peaks at about
// half as much again as the body, for a body that then stays live through the
// parse. Content-Length turns that into one allocation. The header is trusted
// for the size of that allocation and for nothing else: it is ignored unless it
// falls inside the ceiling, and the LimitReader is what enforces the ceiling, so
// a header that lies costs a wrongly-sized buffer and cannot raise the bound.
func (c *Client) read(r io.Reader, contentLength int64) ([]byte, error) {
	var buf bytes.Buffer

	if contentLength > 0 && contentLength <= c.maxBytes {
		buf.Grow(int(contentLength) + 1)
	}

	if _, err := buf.ReadFrom(io.LimitReader(r, c.maxBytes+1)); err != nil {
		return buf.Bytes(), fmt.Errorf("reading the all-zones response: %w", err)
	}

	body := buf.Bytes()

	if int64(len(body)) > c.maxBytes {
		// RETURNED RATHER THAN DISCARDED, because for this one failure the size
		// IS the cause, and returning nil left `response_bytes` NULL on exactly
		// the run whose row needed it — an operator reading that row was told a
		// fetch failed and denied the only figure that says why.
		//
		// It is the ceiling plus one, which is what the LimitReader allowed and
		// all that can be known without reading a body this function has just
		// refused to hold. On the row it reads as "at or past the ceiling". No
		// truncated body is handed to a parse by this: the error is non-nil, so
		// the caller records the size and stops.
		return body, fmt.Errorf("the all-zones response reached the %d byte ceiling and was refused unread; %d bytes were taken to establish that",
			c.maxBytes, len(body))
	}

	return body, nil
}

func snippet(body []byte) string {
	if len(body) > errorBodySnippet {
		return string(body[:errorBodySnippet]) + "…"
	}

	return string(body)
}
