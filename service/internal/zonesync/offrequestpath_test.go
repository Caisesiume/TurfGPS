package zonesync

import (
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// This file is `FR-022` AC2: a planning request issues no all-zones request and
// does not wait on the refresh job.
//
// ---------------------------------------------------------------------------
// WHY IT IS AN IMPORT-GRAPH CHECK AND NOT A TEST OF A HANDLER. Read before
// editing, and especially before "improving" it into something narrower.
//
// AC2 is satisfiable by absence, and that is the whole difficulty. There is no
// plan handler in this module yet. A test that asked the plan handler whether it
// fetched would have nothing to ask, would report green, and would go on
// reporting green on the day a handler landed wired the wrong way — a criterion
// that reports success while measuring nothing, which
// `docs/Requirements/README.md § ID allocation ledger` records this corpus
// having already rejected once at the requirement level.
//
// So the check ranges over the import graph rather than over a handler:
//
//   - It covers every package in the module, so a future handler is covered
//     wherever it lands. Nobody has to remember to add it.
//   - It is transitive, so a handler that reaches the worker through two
//     intermediate packages is caught as readily as one that imports it.
//   - It FAILS rather than passes when the packages it names are missing, so
//     deleting the request surface — or this package — turns the check red
//     instead of turning it vacuous.
//
// It is not the only defence, and it is not asked to be. This package exports no
// way to run a single refresh: Scheduler.Run is a loop the process owns, and
// there is nothing here for a handler to call. Below both, the attempt is gated
// on the last attempt recorded in `sync_run` under an exclusive lock, so even a
// caller that defeated the first two could not spend the endpoint's allowance.
// ---------------------------------------------------------------------------

const (
	modulePath = "github.com/Caisesiume/TurfGPS/service"

	// syncWorker is the package no request may reach.
	syncWorker = modulePath + "/internal/zonesync"

	// requestSurface is where the service's handlers live. It is named
	// explicitly so that removing it is a failure rather than one fewer package
	// to check.
	requestSurface = modulePath + "/internal/httpapi"

	// compositionRoot wires the adapters to the ports and is the only place the
	// worker may be constructed.
	compositionRoot = modulePath + "/cmd/turfgps"

	// syncAdapter implements the ports this package declares, so it necessarily
	// names them. It is a write-path package: no request surface may reach it
	// either, which the check below asserts of requestSurface directly.
	syncAdapter = modulePath + "/internal/syncstore"
)

// mayImportTheWorker is exhaustive. Adding an entry widens the hole this test
// exists to keep shut, so an entry costs an argument in the review that adds it.
var mayImportTheWorker = []string{compositionRoot, syncAdapter}

type listedPackage struct {
	ImportPath string
	Deps       []string
}

// TestNoRequestPathCanReachTheSyncWorker is AC2.
func TestNoRequestPathCanReachTheSyncWorker(t *testing.T) {
	t.Parallel()

	packages := listPackages(t)

	paths := make([]string, 0, len(packages))
	for _, pkg := range packages {
		paths = append(paths, pkg.ImportPath)
	}

	// The four packages the invariant is stated in terms of must all exist. A
	// check whose subject is missing has not passed, it has not run.
	//
	// syncAdapter is the fourth and was once left out of this list, which made
	// the write-path leg of the request-surface check below satisfiable by
	// absence: rename or remove the adapter and slices.Contains is asked
	// whether the surface imports a package that no longer exists, which it
	// cannot, so that assertion reported green having measured nothing — the
	// defect `FR-019`'s Rationale names. The loop over every package does not
	// cover the case. It catches a renamed adapter only while the adapter still
	// imports the worker, and the refactor that most plausibly renames it,
	// extracting the shared types, is the one that stops it importing the
	// worker.
	for _, required := range []string{syncWorker, requestSurface, compositionRoot, syncAdapter} {
		if !slices.Contains(paths, required) {
			t.Fatalf("%s is not a package of this module, so this check has nothing to measure and is refusing rather than passing: the invariant is stated over the sync worker, the request surface, the composition root and the sync's write-path adapter, and all four must exist for it to mean anything",
				required)
		}
	}

	checked := 0

	for _, pkg := range packages {
		if pkg.ImportPath == syncWorker || slices.Contains(mayImportTheWorker, pkg.ImportPath) {
			continue
		}

		checked++

		if slices.Contains(pkg.Deps, syncWorker) {
			t.Errorf("%s reaches %s through its imports, so a request served by it could trigger or wait on the zone refresh: FR-022 AC2 requires the refresh to run from its schedule alone. The composition root wires the worker; nothing else may reach it.",
				pkg.ImportPath, syncWorker)
		}
	}

	if checked == 0 {
		t.Fatal("the invariant ranged over no package at all, so it measured nothing: every package in the module was either the worker itself or on the exemption list, which means the exemption list has swallowed the check")
	}

	// Named separately from the loop above because this is the one the criterion
	// is about, and because it also refuses the write-path adapter, which the
	// general rule permits to name the worker's ports.
	surface := packageByPath(t, packages, requestSurface)

	for _, forbidden := range []string{syncWorker, syncAdapter} {
		if slices.Contains(surface.Deps, forbidden) {
			t.Errorf("the request surface %s reaches %s, which is the zone sync's write path: a request may read how current the copy is and may never reach what refreshes it",
				requestSurface, forbidden)
		}
	}
}

func packageByPath(t *testing.T, packages []listedPackage, path string) listedPackage {
	t.Helper()

	for _, pkg := range packages {
		if pkg.ImportPath == path {
			return pkg
		}
	}

	t.Fatalf("%s was not listed", path)

	return listedPackage{}
}

// listPackages returns every package in this module with its transitive imports.
//
// A missing toolchain, an unresolvable module or an empty listing are all
// failures and never skips, for the reason `cmd/turfgps/build_test.go` already
// gives about the go command: `go test` is running, so a go command that cannot
// answer is a broken host rather than an absent capability — and here it would
// also be an invariant reporting green having read nothing.
func listPackages(t *testing.T) []listedPackage {
	t.Helper()

	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("cannot locate the go command on PATH: %v", err)
	}

	root := moduleRoot(t, goTool)

	out, err := runGo(t, goTool, root, "list", "-json", "./...")
	if err != nil {
		t.Fatalf("listing the module's packages: %v", err)
	}

	dec := json.NewDecoder(strings.NewReader(out))

	var packages []listedPackage

	for {
		var pkg listedPackage

		if err := dec.Decode(&pkg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			t.Fatalf("reading the package list: %v", err)
		}

		packages = append(packages, pkg)
	}

	if len(packages) == 0 {
		t.Fatal("the module listed no packages, so the off-request-path invariant read nothing")
	}

	return packages
}

// moduleRoot is the directory holding this module's go.mod.
//
// It is resolved rather than assumed because `Architecture.md § D8` puts the
// module in `service/` and not at the repository root, so every Go command's
// working directory decides which tree it measured. A test that ran `go list
// ./...` from its own package directory would list one package — itself — and
// find no violation in a module it never read.
func moduleRoot(t *testing.T, goTool string) string {
	t.Helper()

	out, err := runGo(t, goTool, ".", "env", "GOMOD")
	if err != nil {
		t.Fatalf("asking the go command where this module is: %v", err)
	}

	gomod := strings.TrimSpace(out)
	if gomod == "" || gomod == "/dev/null" || gomod == "NUL" {
		t.Fatalf("the go command reports no module file (GOMOD=%q), so this check cannot tell which tree it would be measuring", gomod)
	}

	return filepath.Dir(gomod)
}

func runGo(t *testing.T, goTool, dir string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), goTool, args...)
	cmd.Dir = dir

	out, err := cmd.Output()

	return string(out), err
}
