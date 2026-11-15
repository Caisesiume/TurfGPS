package valhalla

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/Caisesiume/TurfGPS/service/internal/planner"
)

// The three ways the tile set's identity is not read. They are the cases
// `DEPLOYMENT.md § The tile build` ("How the service reads it") names, and
// every one is an unknown identity, which plans nothing
// (`Architecture.md § Loading from the extract the tiles are built from`).
var (
	// ErrTileIdentityMissing is the error ExtractSHA256 wraps where the
	// engine does not answer its status, or its verbose status carries no
	// tile stamp.
	ErrTileIdentityMissing = errors.New("the tile identity is missing: the routing engine reported no tile stamp")

	// ErrTileIdentityUnreadable is the error ExtractSHA256 wraps where the
	// tile record is not named, is absent, or holds anything but its SHA-256.
	ErrTileIdentityUnreadable = errors.New("the tile identity is unreadable: the tile record could not be read as one SHA-256")

	// ErrTileIdentityStale is the error ExtractSHA256 wraps where the tile
	// record's stamp is not the one the running engine reports.
	ErrTileIdentityStale = errors.New("the tile identity is stale: the tile record's stamp is not the running engine's")
)

// tileRecordLength is the tile record's whole length: 64 lower-case
// hexadecimal digits and a newline (`DEPLOYMENT.md § The tile build`).
const tileRecordLength = 64 + 1

// tileStampDigits is how many of the record's leading digits are the tile
// stamp: the 64 bits a tile header carries (`DEPLOYMENT.md § The tile build`).
const tileStampDigits = 16

// TileIdentity is the planner's RoutingData read from the engine a Client
// routes through and the tile record beside its tiles, as
// `DEPLOYMENT.md § The tile build` ("How the service reads it") defines.
//
// IT HOLDS NO COPY OF EITHER BETWEEN READS. Every call asks the engine and
// reads the record again, so a swap of the tiles or a rewritten record is seen
// by the next read and never masked by an earlier one.
type TileIdentity struct {
	client *Client
	record string
}

var _ planner.RoutingData = (*TileIdentity)(nil)

// TileIdentity returns the identity of the tiles this Client's engine serves,
// read against the tile record at recordPath. An empty recordPath is a record
// not named, which every read reports unreadable.
func (c *Client) TileIdentity(recordPath string) *TileIdentity {
	return &TileIdentity{client: c, record: recordPath}
}

type statusRequest struct {
	Verbose bool `json:"verbose"`
}

// statusResponse holds `osm_changeset` as the bytes the engine wrote. Decoded
// into a number type it would pass through float64 or be refused on a value
// above 2^63, and the stamp is an unsigned 64-bit integer
// (`DEPLOYMENT.md § The tile build`).
type statusResponse struct {
	OSMChangeset json.RawMessage `json:"osm_changeset"`
}

// ExtractSHA256 returns the SHA-256 the tile record holds, where its first 16
// digits are the stamp the running engine reports, and an error naming the
// case otherwise: ErrTileIdentityMissing, ErrTileIdentityUnreadable or
// ErrTileIdentityStale.
func (t *TileIdentity) ExtractSHA256(ctx context.Context) (string, error) {
	stamp, err := t.engineStamp(ctx)
	if err != nil {
		return "", err
	}

	sha, err := t.recordSHA256()
	if err != nil {
		return "", err
	}

	recorded, err := strconv.ParseUint(sha[:tileStampDigits], 16, 64)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrTileIdentityUnreadable, err)
	}

	// A record whose stamp is zero is stale here too: engineStamp refuses a
	// zero stamp, which the engine omits from its status, so none matches it.
	if recorded != stamp {
		return "", fmt.Errorf("%w: the record's stamp is %d and the engine's %d", ErrTileIdentityStale, recorded, stamp)
	}

	return sha, nil
}

// engineStamp asks the engine for its verbose status and returns its
// `osm_changeset`, read as an unsigned 64-bit integer from the digits the
// engine wrote.
func (t *TileIdentity) engineStamp(ctx context.Context) (uint64, error) {
	raw, err := t.client.post(ctx, t.client.statusURL, statusRequest{Verbose: true})
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrTileIdentityMissing, err)
	}

	var status statusResponse
	if err := json.Unmarshal(raw, &status); err != nil {
		return 0, fmt.Errorf("%w: the status answer is not one: %w", ErrTileIdentityMissing, err)
	}

	digits := string(bytes.TrimSpace(status.OSMChangeset))
	if digits == "" || digits == "null" {
		return 0, fmt.Errorf("%w: the status carries no osm_changeset", ErrTileIdentityMissing)
	}

	stamp, err := strconv.ParseUint(digits, 10, 64)
	if err != nil || stamp == 0 {
		return 0, fmt.Errorf("%w: the status's osm_changeset %.32q is not a positive unsigned 64-bit integer", ErrTileIdentityMissing, digits)
	}

	return stamp, nil
}

// recordSHA256 reads the tile record and returns the SHA-256 it holds, where
// it holds exactly 64 lower-case hexadecimal digits and a newline.
func (t *TileIdentity) recordSHA256() (string, error) {
	if t.record == "" {
		return "", fmt.Errorf("%w: no tile record is named", ErrTileIdentityUnreadable)
	}

	f, err := os.Open(t.record)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrTileIdentityUnreadable, err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrTileIdentityUnreadable, err)
	}

	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%w: the tile record is not a regular file", ErrTileIdentityUnreadable)
	}

	// One byte past the record's length, so a longer file is seen as longer.
	raw, err := io.ReadAll(io.LimitReader(f, tileRecordLength+1))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrTileIdentityUnreadable, err)
	}

	if len(raw) != tileRecordLength || raw[tileRecordLength-1] != '\n' {
		return "", fmt.Errorf("%w: the tile record is not 64 digits and a newline", ErrTileIdentityUnreadable)
	}

	for _, b := range raw[:tileRecordLength-1] {
		if (b < '0' || b > '9') && (b < 'a' || b > 'f') {
			return "", fmt.Errorf("%w: the tile record holds a character that is not a lower-case hexadecimal digit", ErrTileIdentityUnreadable)
		}
	}

	return string(raw[:tileRecordLength-1]), nil
}
