package main

// Tests binding `NFR-003` AC2: the service starts and serves its first request
// in the image, with no installation step between the copy and the start. AC1's
// *condition* — an image carrying no language runtime, no application server
// and no project-installed dependencies — is verified here too, because the
// image is the only place the absence of a runtime can be observed at all.
//
// ---------------------------------------------------------------------------
// WHY THIS FAILS INSTEAD OF SKIPPING WHEN NO CONTAINER RUNTIME IS PRESENT.
// Read before "fixing" it.
//
// `go test` prints nothing whatever for a package that passes: not a skip's
// message, not t.Log, not even a direct write to the test binary's stderr. All
// of it is buffered and discarded unless the package fails or -v is given.
// Measured on this repository's toolchain, go1.26.2, on 15 August 2026.
//
// The gate in `local-gates` runs `go test -race -count=1 ./...`, which is not
// verbose. So a t.Skip here would reduce the entire evidence for AC2 to the
// word `ok` — character-for-character what a run that built the image, started
// the container and served a request prints. A criterion whose absence of
// verification is indistinguishable from its verification is asserted, not
// verified, which is the vacuous pass `docs/DELIVERY.md § Proof that a test can
// fail` exists to prevent.
//
// So a missing container runtime is reported the way `Makefile`'s test target
// already reports a missing C compiler: as a red naming a missing HOST
// capability, not as a quiet pass. The message says so in its own words, so the
// red cannot be mistaken for a fault in the service. Do not convert this to a
// skip, and do not add an environment variable that silences it — either one
// restores the invisible pass this shape exists to remove.
//
// This file introduces no assertion about linkage: `NFR-003` is silent about it
// while `Architecture.md § D6` is open, and #19 forbids this story to settle it.
// What is asserted is that the image starts and serves, never what it links to.
// ---------------------------------------------------------------------------

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	// startupTimeout bounds image start-up and the first served request. It is
	// generous because it covers a cold container start, not a solve.
	startupTimeout = 60 * time.Second

	// requestTimeout bounds a single attempt at the first request.
	requestTimeout = 5 * time.Second

	// cleanupTimeout bounds each teardown command.
	cleanupTimeout = 60 * time.Second

	// pollInterval separates attempts while the container is still coming up.
	pollInterval = 250 * time.Millisecond

	// containerPort is the port the entry point listens on inside the image,
	// per defaultAddr in main.go and EXPOSE in `service/Dockerfile`.
	containerPort = "8080"
)

// containerRuntimes are the CLIs this harness can drive, in preference order.
// All three take the same command line for build, run, port, logs and rm.
var containerRuntimes = []string{"docker", "podman", "nerdctl"}

// TestImageStartsAndServesFirstRequest binds `NFR-003` AC2, and AC1's condition
// along with it.
//
// It builds the image from `service/Dockerfile`, starts a container from it and
// makes the first request. Nothing is copied in, mounted or executed inside the
// container between the build and the start — no volume, no `exec`, no wrapper
// — which is how "no installation step is performed between the copy and the
// start" is held: by construction, and readable here as the absence of any such
// command.
func TestImageStartsAndServesFirstRequest(t *testing.T) {
	runtime := findContainerRuntime()
	if runtime == "" {
		t.Fatalf("NFR-003 AC2 IS NOT VERIFIED BY THIS RUN.\n\n"+
			"No container runtime is on PATH (looked for: %s), so the image was never built,\n"+
			"the service was never started in it, and neither AC2's metric — start-up and the\n"+
			"first served request in that image — nor AC1's condition that the image carries\n"+
			"no language runtime was measured by anything.\n\n"+
			"This is a missing capability of THIS HOST, not a fault in the service. Nothing\n"+
			"in this run is evidence for or against AC2.\n\n"+
			"Remedy: install a container runtime and re-run. Do not make this test skip: a\n"+
			"skipped test prints nothing at all under the gate's non-verbose `go test`, which\n"+
			"would leave this criterion looking verified when nothing checked it.",
			strings.Join(containerRuntimes, ", "))
	}

	root := moduleRoot(t)
	tag := fmt.Sprintf("turfgps-nfr003-test:%d", time.Now().UnixNano())

	// The build context is the module root, which is `service/`, exactly as
	// `make image` builds it.
	if out, err := exec.CommandContext(t.Context(), runtime, "build", "--tag", tag, root).CombinedOutput(); err != nil {
		t.Fatalf("NFR-003 AC1/AC2 FAILED — the image did not build (%s build %s): %v\n%s", runtime, root, err, out)
	}
	t.Cleanup(func() { runCleanup(runtime, "image", "rm", "--force", tag) })

	assertNothingWrapsTheExecutable(t, runtime, tag)

	out, err := exec.CommandContext(t.Context(), runtime,
		"run", "--detach", "--publish", "127.0.0.1:0:"+containerPort, tag).Output()
	if err != nil {
		t.Fatalf("NFR-003 AC2 FAILED — the container did not start from the built image: %v\n%s",
			err, commandStderr(err))
	}

	container := strings.TrimSpace(string(out))
	t.Cleanup(func() { runCleanup(runtime, "rm", "--force", container) })

	addr := publishedAddr(t, runtime, container)
	body, status := firstServedRequest(t, runtime, container, addr)

	if status != http.StatusOK {
		t.Errorf("NFR-003 AC2 — the first request served in the image returned status %d, want %d",
			status, http.StatusOK)
	}

	if len(body) == 0 {
		t.Errorf("NFR-003 AC2 — the first request served in the image returned an empty body, want a non-empty response")
	}
}

// findContainerRuntime returns the first usable runtime CLI, or "".
func findContainerRuntime() string {
	for _, candidate := range containerRuntimes {
		if path, err := exec.LookPath(candidate); err == nil {
			return path
		}
	}

	return ""
}

// assertNothingWrapsTheExecutable checks that the image starts the executable
// directly. A single-element entry point is the observable form of "no
// application server" and of "no installation step between the copy and the
// start": there is no shell, no supervisor and no start-up script in front of
// the binary that could install anything.
func assertNothingWrapsTheExecutable(t *testing.T, runtime, tag string) {
	t.Helper()

	out, err := exec.CommandContext(t.Context(), runtime,
		"image", "inspect", "--format", "{{json .Config.Entrypoint}}", tag).Output()
	if err != nil {
		// The criterion is untouched by this: it is the harness that could not
		// read the image. Reported red rather than passed over, because a check
		// that silently did not run is the failure this file exists to avoid.
		t.Errorf("harness could not read the image entry point (%s image inspect): %v\n%s — the runtime may not accept this format; AC1's no-application-server condition was not checked",
			runtime, err, commandStderr(err))

		return
	}

	var entrypoint []string
	if err := json.Unmarshal(out, &entrypoint); err != nil {
		t.Errorf("harness could not parse the image entry point %q: %v — AC1's no-application-server condition was not checked", out, err)

		return
	}

	if len(entrypoint) != 1 {
		t.Errorf("NFR-003 AC1 — the image entry point is %v (%d elements), want exactly 1: the executable must be started directly, with nothing wrapping it",
			entrypoint, len(entrypoint))
	}
}

// publishedAddr returns the host address the container's port is published on.
func publishedAddr(t *testing.T, runtime, container string) string {
	t.Helper()

	out, err := exec.CommandContext(t.Context(), runtime, "port", container, containerPort+"/tcp").Output()
	if err != nil {
		t.Fatalf("NFR-003 AC2 FAILED — could not read the published port of the running container: %v\n%s",
			err, commandStderr(err))
	}

	lines := strings.Split(strings.ReplaceAll(strings.TrimSpace(string(out)), "\r\n", "\n"), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("NFR-003 AC2 FAILED — the container published no address for port %s", containerPort)
	}

	_, port, err := net.SplitHostPort(strings.TrimSpace(lines[0]))
	if err != nil {
		t.Fatalf("NFR-003 AC2 FAILED — could not parse the published address %q: %v", lines[0], err)
	}

	// The published address is dialled on the loopback the container was bound
	// to, rather than on whatever the runtime prints, which may be 0.0.0.0.
	return net.JoinHostPort("127.0.0.1", port)
}

// firstServedRequest makes requests until one is served or startupTimeout
// expires, and returns the first served response. A refused connection is a
// container that has not finished starting, not a served request.
func firstServedRequest(t *testing.T, runtime, container, addr string) ([]byte, int) {
	t.Helper()

	client := &http.Client{Timeout: requestTimeout}
	deadline := time.Now().Add(startupTimeout)

	var lastErr error
	for {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+addr+"/", nil)
		if err != nil {
			t.Fatalf("building the request: %v", err)
		}

		resp, err := client.Do(req)
		if err == nil {
			body, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()

			if readErr != nil {
				t.Fatalf("NFR-003 AC2 FAILED — reading the first served response: %v", readErr)
			}

			return body, resp.StatusCode
		}

		lastErr = err

		if time.Now().After(deadline) {
			t.Fatalf("NFR-003 AC2 FAILED — the service in the image served no request within %s of the container starting.\nlast error: %v\ncontainer logs:\n%s",
				startupTimeout, lastErr, containerLogs(runtime, container))
		}

		time.Sleep(pollInterval)
	}
}

// containerLogs returns the container's output, for a failure message. Its own
// failure is returned as text: it runs only on a path that is already failing,
// and losing the primary message to a secondary error helps nobody.
func containerLogs(runtime, container string) string {
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, runtime, "logs", container).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("(could not read container logs: %v)", err)
	}

	return string(out)
}

// runCleanup runs a teardown command on its own context.
//
// Not t.Context(): that context is cancelled just before cleanups run, so a
// teardown built on it would be killed before it could remove anything.
func runCleanup(name string, args ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	_ = exec.CommandContext(ctx, name, args...).Run()
}

// commandStderr returns the stderr captured by exec when a command fails, which
// is where every runtime CLI puts the reason.
func commandStderr(err error) string {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return string(exitErr.Stderr)
	}

	return ""
}
