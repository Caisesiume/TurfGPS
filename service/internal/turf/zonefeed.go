package turf

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// FeedClient requests the zone feed (#232): the latest-created zones, from the
// feed endpoint the Owner's zone-sync decision on #20 adds
// (`#issuecomment-6279336314`), until `Architecture.md § Retrieving zones`
// records it.
//
// The endpoint arrives from configuration as a resolved URL, as every other
// does, and must address APIVersion. Every request draws on the process's one
// budget through the same limitedTransport as Client's and SessionClient's, so
// the feed, the sync and a review session count against one limit (`NFR-001`).
// The feed is held to the API's general limit only: it is not the all-zones
// endpoint, and its requests are never metered by that endpoint's gate.
//
// IT FOLLOWS NO REDIRECT, for SessionClient's reason: the endpoint is a resolved
// URL, so a redirect is not part of an ordinary answer.
type FeedClient struct {
	endpoint *url.URL
	http     *http.Client
	maxBytes int64
}

// errFeedRedirect is returned for any redirect the feed endpoint answers with.
var errFeedRedirect = errors.New("the zone-feed endpoint is configured as a resolved URL and follows no redirect")

func refuseFeedRedirect(req *http.Request, _ []*http.Request) error {
	return fmt.Errorf("%w; refused the hop to %s", errFeedRedirect, req.URL.Redacted())
}

// afterDateParam is the feed's lower-bound parameter, and afterDateLayout the
// form the API index documents for it — `afterDate=2013-08-27T12:11:14+0000`,
// URL-encoded — read from the live index on 20 October 2026. A request
// carrying it returns only items newer than it; observed live the same day to
// exclude an item at exactly that time.
const (
	afterDateParam  = "afterDate"
	afterDateLayout = "2006-01-02T15:04:05-0700"
)

// NewFeedClient builds the zone-feed half of the TurfClient port. maxBytes caps
// one response, as it does for NewClient; only the general limit is required.
func NewFeedClient(feedURL string, maxBytes int64, limits Limits) (*FeedClient, error) {
	if feedURL == "" {
		return nil, errors.New("no zone-feed endpoint")
	}

	endpoint, err := url.Parse(feedURL)
	if err != nil {
		return nil, fmt.Errorf("the zone-feed endpoint is not a URL: %w", err)
	}

	if err := addressesCurrentVersion(endpoint); err != nil {
		return nil, fmt.Errorf("the zone-feed endpoint: %w", err)
	}

	if maxBytes <= 0 || maxBytes > maxCeiling {
		return nil, fmt.Errorf("the response ceiling %d is outside (0, %d]", maxBytes, maxCeiling)
	}

	if limits.Request <= 0 {
		return nil, fmt.Errorf("the Turf API's request interval %s is not positive", limits.Request)
	}

	return &FeedClient{
		endpoint: endpoint,
		http: &http.Client{
			CheckRedirect: refuseFeedRedirect,
			Transport:     newLimitedTransport(http.DefaultTransport, nil, limits),
		},
		maxBytes: maxBytes,
	}, nil
}

// FetchZoneFeed requests the zone feed once. A non-nil after asks only for
// items newer than it (`FR-120`); nil asks with no lower bound. It has the
// shape of zonesync.FeedFetchFunc without naming it.
//
// The parameter is added to the configured URL's own query rather than
// replacing it, and is the only thing this function adds to the address.
func (c *FeedClient) FetchZoneFeed(ctx context.Context, after *time.Time) ([]byte, error) {
	target := *c.endpoint

	if after != nil {
		query := target.Query()
		query.Set(afterDateParam, after.UTC().Format(afterDateLayout))
		target.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building the zone-feed request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "TurfGPS")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("the zone-feed request failed: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	// maxBytes is at most maxCeiling, so the plus-one cannot wrap.
	body, err := io.ReadAll(io.LimitReader(resp.Body, c.maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading the zone-feed response: %w", err)
	}

	if int64(len(body)) > c.maxBytes {
		return nil, fmt.Errorf("the zone-feed response reached the %d byte ceiling and was refused", c.maxBytes)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the zone-feed endpoint answered %s: %s", resp.Status, snippet(body))
	}

	return body, nil
}
