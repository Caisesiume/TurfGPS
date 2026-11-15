package valhalla

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// These tests answer as Valhalla's status service does at 3.9.1: `/status`
// with `"verbose": true` carries `osm_changeset` as an unsigned 64-bit integer
// where the tiles carry a stamp, and omits it otherwise (`status.md` and
// `serializers.cc`, cited in `DEPLOYMENT.md § The tile build`). That a real
// engine answers so is `tileidentity_integration_test.go`'s to show.

// aboveFloat is a SHA-256 whose stamp is above 2^53 and odd, so no float64
// holds it: a stamp carried through one reads as another number.
const aboveFloat = "fedcba9876543211" + "00112233445566778899aabbccddeeff0011223344556677"

// stampOf is the tile stamp of sha, as `DEPLOYMENT.md § The tile build` step 1
// derives it, written as the engine writes it.
func stampOf(t *testing.T, sha string) string {
	t.Helper()

	v, err := strconv.ParseUint(sha[:16], 16, 64)
	if err != nil {
		t.Fatalf("the stamp of %s: %v", sha, err)
	}

	return strconv.FormatUint(v, 10)
}

// statusEngine answers `/status` with the body changeset gives it, the
// changeset field written verbatim, and counts the verbose asks. An empty
// changeset is a status without the field.
type statusEngine struct {
	changeset atomic.Value
	asked     atomic.Int32
}

func newStatusEngine(t *testing.T, changeset string) (*statusEngine, *httptest.Server) {
	t.Helper()

	e := &statusEngine{}
	e.changeset.Store(changeset)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status" {
			http.Error(w, "not expected", http.StatusNotFound)

			return
		}

		raw, _ := io.ReadAll(r.Body)

		var req struct {
			Verbose bool `json:"verbose"`
		}

		_ = json.Unmarshal(raw, &req)

		w.Header().Set("Content-Type", "application/json")

		// Without verbose the engine answers before osm_changeset
		// (`status_action.cc:55`), as it does when verbose is disallowed.
		cs, _ := e.changeset.Load().(string)
		if !req.Verbose || cs == "" {
			_, _ = io.WriteString(w, `{"version":"3.9.1","tileset_last_modified":1}`)

			return
		}

		e.asked.Add(1)
		_, _ = io.WriteString(w, `{"version":"3.9.1","tileset_last_modified":1,"has_tiles":true,"osm_changeset":`+cs+`}`)
	}))
	t.Cleanup(srv.Close)

	return e, srv
}

func writeRecord(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "tiles.sha256")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing the tile record: %v", err)
	}

	return path
}

func tileIdentity(t *testing.T, engineURL, record string) *TileIdentity {
	t.Helper()

	c, err := New(engineURL, 5*time.Second)
	if err != nil {
		t.Fatalf("New(%q): %v", engineURL, err)
	}

	return c.TileIdentity(record)
}

// TestTheTileIdentityIsTheRecordWhereItsStampIsTheEngines: a record whose
// first 16 digits are the engine's osm_changeset is read as the identity, and
// the engine was asked for its verbose status to give it.
func TestTheTileIdentityIsTheRecordWhereItsStampIsTheEngines(t *testing.T) {
	t.Parallel()

	sha := strings.Repeat("0", 15) + "1" + strings.Repeat("ab", 24)

	engine, srv := newStatusEngine(t, stampOf(t, sha))

	got, err := tileIdentity(t, srv.URL, writeRecord(t, sha+"\n")).ExtractSHA256(t.Context())
	if err != nil || got != sha {
		t.Fatalf("ExtractSHA256 gave %q, %v; want the record's %s", got, err, sha)
	}

	if n := engine.asked.Load(); n != 1 {
		t.Errorf("the engine was asked for its verbose status %d time(s), want 1", n)
	}
}

// TestAStampAboveTwoToTheFiftyThirdIsReadExactly (#249 AC7): the stamp is an
// unsigned 64-bit integer, read from the digits the engine wrote; one carried
// through a float64 is another number, and reads as stale.
func TestAStampAboveTwoToTheFiftyThirdIsReadExactly(t *testing.T) {
	t.Parallel()

	stamp := stampOf(t, aboveFloat)
	if v, _ := strconv.ParseUint(stamp, 10, 64); v <= 1<<53 || uint64(float64(v)) == v {
		t.Fatalf("the stamp %s is one a float64 holds, so it tests nothing", stamp)
	}

	_, srv := newStatusEngine(t, stamp)

	got, err := tileIdentity(t, srv.URL, writeRecord(t, aboveFloat+"\n")).ExtractSHA256(t.Context())
	if err != nil || got != aboveFloat {
		t.Fatalf("with osm_changeset %s, ExtractSHA256 gave %q, %v; want the record's %s", stamp, got, err, aboveFloat)
	}
}

// TestAnIdentityTheEngineDoesNotGiveIsMissing (#249 AC6, missing): an engine
// that does not answer, answers an error, or gives a status without
// osm_changeset — how a zero stamp reaches the service — leaves the identity
// missing, whatever the record holds.
func TestAnIdentityTheEngineDoesNotGiveIsMissing(t *testing.T) {
	t.Parallel()

	record := writeRecord(t, aboveFloat+"\n")

	gone := httptest.NewServer(http.NotFoundHandler())
	goneURL := gone.URL
	gone.Close()

	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error_code":199,"error":"Unknown"}`, http.StatusInternalServerError)
	}))
	t.Cleanup(failing.Close)

	_, noStamp := newStatusEngine(t, "")
	_, zero := newStatusEngine(t, "0")
	_, null := newStatusEngine(t, "null")
	_, quoted := newStatusEngine(t, `"`+stampOf(t, aboveFloat)+`"`)
	_, negative := newStatusEngine(t, "-1")
	_, fraction := newStatusEngine(t, stampOf(t, aboveFloat)+".0")
	_, exponent := newStatusEngine(t, "1.8e19")

	for name, url := range map[string]string{
		"an engine that does not answer":        goneURL,
		"an engine that answers an error":       failing.URL,
		"a status without osm_changeset":        noStamp.URL,
		"a status whose osm_changeset is zero":  zero.URL,
		"a status whose osm_changeset is null":  null.URL,
		"an osm_changeset written as a string":  quoted.URL,
		"a negative osm_changeset":              negative.URL,
		"an osm_changeset written with a point": fraction.URL,
		"an osm_changeset written as exponent":  exponent.URL,
	} {
		got, err := tileIdentity(t, url, record).ExtractSHA256(t.Context())
		if !errors.Is(err, ErrTileIdentityMissing) || got != "" {
			t.Errorf("%s: ExtractSHA256 gave %q, %v; want no identity and an error wrapping ErrTileIdentityMissing", name, got, err)
		}
	}
}

// TestARecordThatIsNotOneSHA256IsUnreadable (#249 AC6, unreadable): a record
// not named, absent, or holding anything but 64 lower-case hexadecimal digits
// and a newline leaves the identity unreadable, though the engine answers.
func TestARecordThatIsNotOneSHA256IsUnreadable(t *testing.T) {
	t.Parallel()

	_, srv := newStatusEngine(t, stampOf(t, aboveFloat))

	for name, record := range map[string]string{
		"no record named":                 "",
		"an absent record":                filepath.Join(t.TempDir(), "absent.sha256"),
		"a directory":                     t.TempDir(),
		"an empty record":                 writeRecord(t, ""),
		"no newline":                      writeRecord(t, aboveFloat),
		"a CRLF":                          writeRecord(t, aboveFloat+"\r\n"),
		"two newlines":                    writeRecord(t, aboveFloat+"\n\n"),
		"a second line":                   writeRecord(t, aboveFloat+"\n"+aboveFloat+"\n"),
		"upper-case digits":               writeRecord(t, strings.ToUpper(aboveFloat)+"\n"),
		"63 digits":                       writeRecord(t, aboveFloat[:63]+"\n"),
		"65 digits":                       writeRecord(t, aboveFloat+"0\n"),
		"a leading space":                 writeRecord(t, " "+aboveFloat[:63]+"\n"),
		"a sha256sum line":                writeRecord(t, aboveFloat+"  extract.osm.pbf\n"),
		"a character that is not a digit": writeRecord(t, aboveFloat[:63]+"g\n"),
	} {
		got, err := tileIdentity(t, srv.URL, record).ExtractSHA256(t.Context())
		if !errors.Is(err, ErrTileIdentityUnreadable) || got != "" {
			t.Errorf("%s: ExtractSHA256 gave %q, %v; want no identity and an error wrapping ErrTileIdentityUnreadable", name, got, err)
		}
	}
}

// TestARecordWhoseStampIsNotTheEnginesIsStale (#249 AC6, stale): a record
// whose first 16 digits are not osm_changeset — a record left from another
// build, one whose stamp is zero, one differing in a single digit — is not
// the identity of the tiles the engine serves.
func TestARecordWhoseStampIsNotTheEnginesIsStale(t *testing.T) {
	t.Parallel()

	_, srv := newStatusEngine(t, stampOf(t, aboveFloat))

	for name, sha := range map[string]string{
		"another build's record":       strings.Repeat("1", 64),
		"a record whose stamp is zero": strings.Repeat("0", 16) + aboveFloat[16:],
		"one digit of the stamp off":   "fedcba9876543210" + aboveFloat[16:],
	} {
		got, err := tileIdentity(t, srv.URL, writeRecord(t, sha+"\n")).ExtractSHA256(t.Context())
		if !errors.Is(err, ErrTileIdentityStale) || got != "" {
			t.Errorf("%s: ExtractSHA256 gave %q, %v; want no identity and an error wrapping ErrTileIdentityStale", name, got, err)
		}
	}
}

// TestTheTileIdentityIsReadAgainOnEveryCall: the reader holds no copy of the
// stamp or the record between reads, so a rewritten record or a restarted
// engine is what the next read sees.
func TestTheTileIdentityIsReadAgainOnEveryCall(t *testing.T) {
	t.Parallel()

	engine, srv := newStatusEngine(t, stampOf(t, aboveFloat))
	record := writeRecord(t, aboveFloat+"\n")
	identity := tileIdentity(t, srv.URL, record)

	if got, err := identity.ExtractSHA256(t.Context()); err != nil || got != aboveFloat {
		t.Fatalf("the first read gave %q, %v; want %s", got, err, aboveFloat)
	}

	other := strings.Repeat("2", 64)
	if err := os.WriteFile(record, []byte(other+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if got, err := identity.ExtractSHA256(t.Context()); !errors.Is(err, ErrTileIdentityStale) {
		t.Errorf("after the record was rewritten for another build the read gave %q, %v; want stale", got, err)
	}

	engine.changeset.Store(stampOf(t, other))

	if got, err := identity.ExtractSHA256(t.Context()); err != nil || got != other {
		t.Errorf("after the engine came to report the new record's stamp the read gave %q, %v; want %s", got, err, other)
	}

	if n := engine.asked.Load(); n != 3 {
		t.Errorf("three reads asked the engine %d time(s), want 3", n)
	}
}
