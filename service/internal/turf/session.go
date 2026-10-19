package turf

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// SessionEndpoints are the two endpoints a review session reads, each as a
// resolved URL from configuration, as the all-zones endpoint is: the
// player-data call of `Architecture.md § Player data` (a POST to `users`) and
// the region response of `Architecture.md § Region lords` (a GET of `regions`).
// No path is written in this package; NewSessionClient refuses either one
// unless it addresses APIVersion.
type SessionEndpoints struct {
	Users   string
	Regions string
}

// SessionClient makes a review session's Turf requests: one player-data call
// and one region request, per `Architecture.md § The call budget`.
//
// Every request it sends draws on the process's one budget, through the same
// limitedTransport as Client's, so the session and the sync count against one
// limit (`NFR-001`).
//
// IT FOLLOWS NO REDIRECT. Both endpoints are configured as resolved URLs, so a
// redirect is not part of an ordinary answer; following one would let a POST
// turn into a GET, and would let a hop reach an endpoint this client meters
// with the general gate alone. A redirect is refused with the endpoint named.
type SessionClient struct {
	users    string
	regions  string
	http     *http.Client
	maxBytes int64
}

// errRedirect is returned for any redirect a SessionClient is answered with.
var errRedirect = errors.New("a review session's Turf endpoints are configured as resolved URLs and follow no redirect")

func refuseRedirect(req *http.Request, _ []*http.Request) error {
	return fmt.Errorf("%w; refused the hop to %s", errRedirect, req.URL.Redacted())
}

// NewSessionClient builds the session half of the TurfClient port.
//
// maxBytes caps one response, as it does for NewClient. Only the general limit
// is required: this client addresses no all-zones endpoint.
func NewSessionClient(endpoints SessionEndpoints, maxBytes int64, limits Limits) (*SessionClient, error) {
	for _, e := range []struct{ role, raw string }{
		{"player-data", endpoints.Users},
		{"region", endpoints.Regions},
	} {
		if e.raw == "" {
			return nil, fmt.Errorf("no %s endpoint", e.role)
		}

		u, err := url.Parse(e.raw)
		if err != nil {
			return nil, fmt.Errorf("the %s endpoint is not a URL: %w", e.role, err)
		}

		if err := addressesCurrentVersion(u); err != nil {
			return nil, fmt.Errorf("the %s endpoint: %w", e.role, err)
		}
	}

	if maxBytes <= 0 || maxBytes > maxCeiling {
		return nil, fmt.Errorf("the response ceiling %d is outside (0, %d]", maxBytes, maxCeiling)
	}

	if limits.Request <= 0 {
		return nil, fmt.Errorf("the Turf API's request interval %s is not positive", limits.Request)
	}

	return &SessionClient{
		users:   endpoints.Users,
		regions: endpoints.Regions,
		http: &http.Client{
			CheckRedirect: refuseRedirect,
			Transport:     newLimitedTransport(http.DefaultTransport, nil, limits),
		},
		maxBytes: maxBytes,
	}, nil
}

// FetchPlayer makes the player-data call for name: one POST to the users endpoint whose
// body is `[{"name": name}]`, the form the API's own documentation gives, with the
// `Content-Type: application/json` it requires. It has the shape of
// ownership.FetchPlayerFunc without naming it.
func (c *SessionClient) FetchPlayer(ctx context.Context, name string) ([]byte, error) {
	payload, err := json.Marshal([]map[string]string{{"name": name}})
	if err != nil {
		return nil, fmt.Errorf("encoding the player-data request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.users, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("building the player-data request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return c.do(req, "player-data")
}

// do sends req once and returns its body, refusing a status other than 200 and
// a body past the ceiling rather than truncating it.
func (c *SessionClient) do(req *http.Request, role string) ([]byte, error) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "TurfGPS")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("the %s request failed: %w", role, err)
	}

	defer func() { _ = resp.Body.Close() }()

	// maxBytes is at most maxCeiling, so the plus-one cannot wrap.
	body, err := io.ReadAll(io.LimitReader(resp.Body, c.maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading the %s response: %w", role, err)
	}

	if int64(len(body)) > c.maxBytes {
		return nil, fmt.Errorf("the %s response reached the %d byte ceiling and was refused", role, c.maxBytes)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the %s endpoint answered %s: %s", role, resp.Status, snippet(body))
	}

	return body, nil
}
