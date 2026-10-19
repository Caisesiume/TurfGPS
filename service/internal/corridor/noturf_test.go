package corridor

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"slices"
	"testing"
)

// This file is the "no Turf API request is issued" half of `FR-023` AC1, held
// structurally: a Turf request is an HTTP request, so a resolution path that
// cannot reach net/http, or the Turf adapter, cannot issue one — on any host,
// with or without a database.
//
// The other half, that the zones come back from the local copy, needs PostGIS
// and is `internal/zonestore/corridor_postgis_test.go`, which also watches the
// process's default HTTP transport while the corridor resolves.
//
// IT FAILS CLOSED. Each package it names must be listed, and the Turf adapter
// must itself reach net/http, so a renamed package or a Turf client that stopped
// using net/http turns this red instead of leaving a leg that measures nothing.

const (
	modulePath = "github.com/Caisesiume/TurfGPS/service"

	// thisPackage and zoneAdapter are every package corridor resolution runs
	// through: this one, and the adapter that answers its port.
	thisPackage = modulePath + "/internal/corridor"
	zoneAdapter = modulePath + "/internal/zonestore"

	// turfAdapter is the TurfClient adapter of `Architecture.md § Ports and
	// adapters`.
	turfAdapter = modulePath + "/internal/turf"

	httpClient = "net/http"
)

type listedPackage struct {
	ImportPath string
	Deps       []string
}

// TestCorridorResolutionCannotReachTurf is FR-023 AC1's "no Turf API request".
func TestCorridorResolutionCannotReachTurf(t *testing.T) {
	t.Parallel()

	listed := listPackages(t, thisPackage, zoneAdapter, turfAdapter)

	turf, ok := listed[turfAdapter]
	if !ok {
		t.Fatalf("%s was not listed, so this check has nothing to measure and is refusing rather than passing", turfAdapter)
	}

	if !slices.Contains(turf.Deps, httpClient) {
		t.Fatalf("the Turf adapter %s no longer reaches %s, so forbidding %s below no longer forbids a Turf request; this check is refusing rather than passing until it names what a Turf request is made with",
			turfAdapter, httpClient, httpClient)
	}

	for _, path := range []string{thisPackage, zoneAdapter} {
		pkg, ok := listed[path]
		if !ok {
			t.Fatalf("%s was not listed, so this check has nothing to measure and is refusing rather than passing", path)
		}

		for _, forbidden := range []string{turfAdapter, httpClient} {
			if pkg.ImportPath == forbidden || slices.Contains(pkg.Deps, forbidden) {
				t.Errorf("%s reaches %s, so resolving a corridor could issue a Turf API request: FR-023 AC1 requires the corridor to resolve from the local copy alone",
					path, forbidden)
			}
		}
	}
}

// listPackages lists the named packages with their transitive imports.
//
// A missing go command or an unlistable package is a failure and never a skip:
// `go test` is running, so a go command that cannot answer is a broken host, and
// an invariant over the import graph that read no graph has not passed.
func listPackages(t *testing.T, paths ...string) map[string]listedPackage {
	t.Helper()

	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("cannot locate the go command on PATH: %v", err)
	}

	cmd := exec.CommandContext(t.Context(), goTool, append([]string{"list", "-json"}, paths...)...)

	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			t.Fatalf("listing %v: %v\n%s", paths, err, exitErr.Stderr)
		}

		t.Fatalf("listing %v: %v", paths, err)
	}

	listed := make(map[string]listedPackage, len(paths))
	dec := json.NewDecoder(bytes.NewReader(out))

	for {
		var pkg listedPackage

		if err := dec.Decode(&pkg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			t.Fatalf("reading the package list: %v", err)
		}

		listed[pkg.ImportPath] = pkg
	}

	return listed
}
