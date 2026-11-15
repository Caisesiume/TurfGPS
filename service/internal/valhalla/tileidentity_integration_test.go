//go:build valhalla

package valhalla

// #249 AC8: the tile identity read through TileIdentity from a real Valhalla
// 3.9.1, whose tiles CI's `valhalla` job built from the act's file with
// `mjolnir.dataset_id` set to that file's tile stamp and serves with
// `service_limits.status.allow_verbose` true, against the tile record the job
// wrote beside them (`DEPLOYMENT.md § The tile build`). It holds the engine
// properties that section cites — the stamp carried into every tile, and
// reported whole as `osm_changeset` — to the engine CI pins, not to a fake.
//
// Built only under the `valhalla` tag, as this package's other integration
// test is, and failing rather than skipping without its inputs.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// The test's other inputs, which CI's `valhalla` job sets. Neither is the
// service's configuration.
const (
	// testTileRecordEnv names the tile record the job wrote beside the tiles.
	testTileRecordEnv = "TURFGPS_TEST_TILE_RECORD"

	// testDatabaseEnv names the database the job loaded the act's file into.
	testDatabaseEnv = "TURFGPS_TEST_DATABASE_URL"
)

func requiredEnv(t *testing.T, name string) string {
	t.Helper()

	v := os.Getenv(name)
	if v == "" {
		t.Fatalf("%s is not set: this file is built only under the valhalla tag, for the engine and the load CI's valhalla job makes, and it fails rather than skips without one", name)
	}

	return v
}

// currentLoad is the SHA-256 the loader recorded as current.
func currentLoad(t *testing.T) string {
	t.Helper()

	conn, err := pgx.Connect(t.Context(), requiredEnv(t, testDatabaseEnv))
	if err != nil {
		t.Fatalf("connecting to the loaded database: %v", err)
	}
	defer func() { _ = conn.Close(t.Context()) }()

	var sha string
	if err := conn.QueryRow(t.Context(), `SELECT extract_sha256 FROM public.osm_extract_load WHERE current`).Scan(&sha); err != nil {
		t.Fatalf("reading the current load: %v", err)
	}

	return sha
}

// TestTheTileIdentityIsTheLoadOfTheFileTheTilesWereBuiltFrom: through the
// reader, the running engine's stamp and the record beside its tiles give the
// SHA-256 the loader recorded as current for the same file; a record written
// for another SHA-256 reads as stale.
func TestTheTileIdentityIsTheLoadOfTheFileTheTilesWereBuiltFrom(t *testing.T) {
	client, err := New(engineUnderTest(t), 30*time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	want := currentLoad(t)

	got, err := client.TileIdentity(requiredEnv(t, testTileRecordEnv)).ExtractSHA256(t.Context())
	if err != nil || got != want {
		t.Fatalf("the tile identity read through the engine is %q, %v; want the current load's %s", got, err, want)
	}

	// Another SHA-256: the same digits with the stamp's first one changed.
	other := []byte(want)
	if other[0] == '0' {
		other[0] = '1'
	} else {
		other[0] = '0'
	}

	record := filepath.Join(t.TempDir(), "other.sha256")
	if err := os.WriteFile(record, append(other, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}

	if got, err := client.TileIdentity(record).ExtractSHA256(t.Context()); !errors.Is(err, ErrTileIdentityStale) {
		t.Errorf("a record written for %s read as %q, %v; want stale", other, got, err)
	}
}
