package main

// Tests binding `NFR-003` AC1: what it takes to start the service on a clean
// host. AC1's metric is the count of files required to start the service, and
// its threshold is exactly one executable.
//
// ---------------------------------------------------------------------------
// WHAT THIS FILE MUST NOT DO, AND WHY. Read before editing.
//
// `NFR-003` is deliberately silent about linkage while `Architecture.md § D6`
// is open, and #19 forbids this story to introduce any test, flag or assertion
// demanding a statically linked or cgo-free binary. So nothing here inspects
// what the executable links against, and nothing here sets GOOS, GOARCH or
// CGO_ENABLED. The build below is the host's own build and the assertion is a
// count, which is what AC1 measures.
//
// Cross-compiling would look tempting — AC1 names a target platform — and it is
// the trap. Setting GOOS on a host without a cross toolchain fails the moment
// D6 is settled toward a cgo raster sampler, and that failure would be about
// linkage rather than about the file count. The count is the same number on
// every platform, so the test asks the host for it.
// ---------------------------------------------------------------------------

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// goTool locates the go command.
//
// A missing toolchain is a failure and never a skip: `go test` is already
// running, so a go command that cannot be found is a broken host rather than an
// absent capability, and AC1 is measurable everywhere this suite runs.
func goTool(t *testing.T) string {
	t.Helper()

	path, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("cannot locate the go command on PATH: %v", err)
	}

	return path
}

// goEnv reads the named `go env` values in one call, in the order asked for.
//
// The output is split rather than trimmed as a whole because an empty value is
// a legitimate answer — GOEXE is empty everywhere except Windows — and trimming
// the trailing blank line away with it would silently return one value fewer
// than was asked for.
func goEnv(t *testing.T, names ...string) []string {
	t.Helper()

	out, err := exec.CommandContext(t.Context(), goTool(t), append([]string{"env"}, names...)...).Output()
	if err != nil {
		t.Fatalf("go env %s: %v", strings.Join(names, " "), err)
	}

	text := strings.TrimSuffix(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n")
	values := strings.Split(text, "\n")
	if len(values) != len(names) {
		t.Fatalf("go env %s returned %d values, want %d", strings.Join(names, " "), len(values), len(names))
	}

	return values
}

// moduleRoot is the directory holding the service's go.mod. It is also the
// build context of `service/Dockerfile`, which is why the image harness shares
// this helper.
func moduleRoot(t *testing.T) string {
	t.Helper()

	gomod := goEnv(t, "GOMOD")[0]
	if gomod == "" || gomod == os.DevNull {
		t.Fatalf("go env GOMOD reports no module (%q); this test must run inside the service module", gomod)
	}

	return filepath.Dir(gomod)
}

// TestBuildProducesExactlyOneExecutable binds `NFR-003` AC1's threshold: the
// files required to start the service on a clean host are exactly one
// executable.
//
// It builds the whole command tree into an empty directory and counts what
// lands there. The pattern is ./cmd/... rather than ./cmd/turfgps because AC1
// asks what it takes to start the service, not what one package compiles to: a
// second command arriving under cmd/ means a second file on the clean host, and
// this test is where that change to the deployment shape has to be argued
// rather than absorbed.
//
// The bug it catches is the one `NFR-003`'s Risk names — the build acquiring a
// tree of side files that must ship beside the executable. An empty output
// directory is what makes that observable: anything the build emits is in it.
func TestBuildProducesExactlyOneExecutable(t *testing.T) {
	root := moduleRoot(t)
	exeSuffix := goEnv(t, "GOEXE")[0]
	outDir := t.TempDir()

	// -o takes a directory, so every artefact of every matched package lands
	// here and nothing lands beside its package. -trimpath matches how the
	// Makefile and the Dockerfile build the shipped artefact. No environment is
	// added: see the note at the top of this file.
	build := exec.CommandContext(t.Context(), goTool(t),
		"build", "-trimpath", "-o", outDir+string(os.PathSeparator), "./cmd/...")
	build.Dir = root

	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/... in %s: %v\n%s", root, err, out)
	}

	var built, dirs []string
	if err := filepath.WalkDir(outDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(outDir, path)
		if err != nil {
			return err
		}

		switch {
		case rel == ".":
		case entry.IsDir():
			dirs = append(dirs, rel)
		default:
			built = append(built, rel)
		}

		return nil
	}); err != nil {
		t.Fatalf("walking the build output directory: %v", err)
	}

	if len(dirs) != 0 {
		t.Errorf("building ./cmd/... into an empty directory produced %d subdirectories [%s], want 0 — a tree of side files beside the executable is what AC1's threshold excludes",
			len(dirs), strings.Join(dirs, " "))
	}

	if len(built) != 1 {
		t.Fatalf("building ./cmd/... into an empty directory produced %d files [%s], want exactly 1 executable",
			len(built), strings.Join(built, " "))
	}

	if want := "turfgps" + exeSuffix; built[0] != want {
		t.Errorf("the built executable is named %q, want %q", built[0], want)
	}

	info, err := os.Stat(filepath.Join(outDir, built[0]))
	if err != nil {
		t.Fatalf("stat %s: %v", built[0], err)
	}

	if info.Size() == 0 {
		t.Errorf("the built executable %s is %d bytes, want a non-empty file", built[0], info.Size())
	}
}
