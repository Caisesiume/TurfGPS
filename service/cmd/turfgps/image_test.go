package main

// Tests binding `NFR-003` AC2: the service starts and serves its first request
// in the image, with no installation step between the copy and the start.
//
// ---------------------------------------------------------------------------
// WHAT OF AC1's CONDITION IS MEASURED HERE, AND WHAT IS NOT.
//
// AC1's condition names an image carrying no language runtime, no application
// server and no project-installed dependencies. Two halves of that are measured
// here, and the claim is kept to exactly those two:
//
//   no application server      — the image's entry point is asserted by
//                                identity, with an empty command beside it, so
//                                nothing may stand in front of the executable;
//                                and the container is stopped with a signal and
//                                required to exit of its own accord, which is
//                                what a wrapper cannot do.
//   nothing to install with    — a shell cannot be run in the runtime image.
//                                `service/Dockerfile` gives that as its reason
//                                no installation step can happen between the
//                                copy and the start, and it is run here rather
//                                than believed.
//
// What none of this does is enumerate the base image's contents, and the gap is
// named rather than papered over. A base carrying a language runtime or a
// package manager carries a shell in practice — the builder stage this image is
// built from does — so the shell probe catches the regression that would really
// happen, which is the runtime stage being repointed at a fatter base. It would
// not catch an image that ships a language runtime and no shell. Until an
// affirmative probe for that exists, this file claims the two halves above and
// no more.
//
// An earlier version of this comment said AC1's condition was verified here,
// on the strength of an assertion that counted the entry point's elements. A
// one-element wrapper script satisfied that count, and so did a full toolchain
// base. That is the defect these probes exist to close.
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
// ---------------------------------------------------------------------------
// A FAILURE THIS HARNESS CANNOT ATTRIBUTE IS REPORTED AS A HARNESS FAULT.
//
// A runtime CLI that does not accept a format, print a port or parse the way
// this file expects has told us nothing about the service. Reporting that as
// `NFR-003 AC2 FAILED` names the wrong defect and sends the next reader to the
// wrong file, so those failures say HARNESS FAULT and say what went unmeasured.
// They are still red: a check that did not run is never a pass.
//
// ---------------------------------------------------------------------------
// This file introduces no assertion about linkage: `NFR-003` is silent about it
// while `Architecture.md § D6` is open, and #19 forbids this story to settle it.
// What is asserted is that the image starts, serves and stops, and what the
// image contains — never what the executable links against. Those are different
// questions and the probes below stay on the first one.
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
	"strconv"
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

	// stopGrace is how long the container is given to exit after being
	// signalled. It is a whole number of seconds because that is what every
	// runtime's stop flag takes, and it sits above the service's own drain
	// budget so that a container reaching this bound has failed to react to the
	// signal rather than merely taken its time draining.
	stopGrace = 10 * time.Second

	// pollInterval separates attempts while the container is still coming up.
	pollInterval = 250 * time.Millisecond

	// containerPort is the port the entry point listens on inside the image,
	// per defaultAddr in main.go and EXPOSE in `service/Dockerfile`.
	containerPort = "8080"

	// executablePath is where `service/Dockerfile` copies the executable, and
	// so is the entry point the image must start. The literal is the assertion:
	// deriving it from the Dockerfile would only prove the image matches the
	// file it was built from, which a wrapper script satisfies just as well.
	executablePath = "/usr/local/bin/turfgps"
)

// containerRuntimes are the CLIs this harness can drive, in preference order.
// All three take the same command line for build, run, port, logs and rm.
var containerRuntimes = []string{"docker", "podman", "nerdctl"}

// shellPaths are where a shell would sit in this image's family, including the
// busybox one that a distroless `:debug` tag carries. Every one of them must
// fail to run: `service/Dockerfile` states there is no shell to install with,
// and this list is where that statement is put to the test.
//
// The list is about what the image contains. It says nothing about what the
// executable links against, and must not grow in that direction.
var shellPaths = []string{
	"/bin/sh",
	"/bin/bash",
	"/bin/dash",
	"/usr/bin/sh",
	"/usr/bin/bash",
	"/busybox/sh",
}

// TestImageStartsAndServesFirstRequest binds `NFR-003` AC2, and the two halves
// of AC1's condition named at the top of this file.
//
// It builds the image from `service/Dockerfile`, probes the image itself,
// starts a container from it, makes the first request and stops it with a
// signal. Nothing is copied in, mounted or executed inside the container
// between the build and the start — no volume, no `exec`, no wrapper — which is
// how "no installation step is performed between the copy and the start" is
// held: by construction, and readable here as the absence of any such command.
func TestImageStartsAndServesFirstRequest(t *testing.T) {
	runtime := findContainerRuntime()
	if runtime == "" {
		t.Fatalf("NFR-003 AC2 IS NOT VERIFIED BY THIS RUN.\n\n"+
			"No container runtime is on PATH (looked for: %s), so the image was never built,\n"+
			"the service was never started in it, and nothing measured either AC2's metric —\n"+
			"start-up and the first served request in that image — or the two halves of AC1's\n"+
			"condition this file probes: that nothing wraps the executable, and that there is\n"+
			"no shell in the image to install anything with.\n\n"+
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
		t.Fatalf("NFR-003 AC1/AC2 FAILED, OR A HARNESS FAULT — this run cannot tell which,\n"+
			"and the runtime's own output below is what separates the two readings.\n\n"+
			"The image did not build (%s build %s): %v\n%s\n"+
			"A failure to fetch the base images `service/Dockerfile` pins by digest — a rate-limited\n"+
			"registry, no network, a proxy, an expired credential — is a missing capability of THIS\n"+
			"HOST, not a fault in the service. The digest pins make that fetch a mandatory network\n"+
			"step of every build, so it is the likeliest reading on a fresh runner. On it the image\n"+
			"was never built, so nothing measured AC2's metric — start-up and the first served\n"+
			"request in that image — nor either half of AC1's condition this file probes.\n"+
			"A failure the runtime attributes to the build itself — an instruction exiting non-zero,\n"+
			"a file missing from the context — is the service's defect, and `service/Dockerfile` is\n"+
			"where to read it.\n\n"+
			"Red on either reading: a check that did not run is never a pass.",
			runtime, root, err, out)
	}
	t.Cleanup(func() { runCleanup(runtime, "image", "rm", "--force", tag) })

	assertNothingWrapsTheExecutable(t, runtime, tag)
	assertNoShellInTheImage(t, runtime, tag)

	out, err := exec.CommandContext(t.Context(), runtime,
		"run", "--detach", "--publish", "127.0.0.1:0:"+containerPort, tag).Output()
	if err != nil {
		t.Fatalf("NFR-003 AC2 FAILED, OR A HARNESS FAULT — this run cannot tell which, and the\n"+
			"runtime's own refusal below is what separates the two readings.\n\n"+
			"The container did not start from the built image (%s run --detach): %v\n%s\n"+
			"A refusal naming the host rather than the image — no permission to publish a port, the\n"+
			"address already in use, no network namespace, a runtime daemon that is not running — is\n"+
			"a missing capability of THIS HOST, not a fault in the service. On that reading no\n"+
			"container ever ran, so nothing measured AC2's metric: start-up and the first served\n"+
			"request in that image. A refusal naming the image's entry point is the service's defect.\n\n"+
			"Red on either reading: a check that did not run is never a pass.",
			runtime, err, commandStderr(err))
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

	assertStopsCleanlyOnSignal(t, runtime, container)
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
// itself, and starts it with nothing appended.
//
// The assertion is the entry point's IDENTITY, not its length. A single-element
// entry point was the previous check and it is satisfied by a wrapper script,
// which is precisely the thing "no application server" excludes: one element
// named /docker-entrypoint.sh is one element. The command must be empty for the
// same reason — a default command is appended to an exec-form entry point, so
// "the executable is started directly" is only true while there is nothing
// beside it.
func assertNothingWrapsTheExecutable(t *testing.T, runtime, tag string) {
	t.Helper()

	entrypoint, ok := imageConfigList(t, runtime, tag, "Entrypoint")
	if !ok {
		return
	}

	if len(entrypoint) != 1 || entrypoint[0] != executablePath {
		t.Errorf("NFR-003 AC1 — the image entry point is %q, want exactly [%s]: the executable must be what the image starts, with nothing wrapping it",
			entrypoint, executablePath)
	}

	cmd, ok := imageConfigList(t, runtime, tag, "Cmd")
	if !ok {
		return
	}

	if len(cmd) != 0 {
		t.Errorf("NFR-003 AC1 — the image carries the default command %q, want none: a command is appended to the entry point, so the executable is only started directly while there is none",
			cmd)
	}
}

// assertNoShellInTheImage runs each shell path as the image's entry point and
// requires every one of them to fail.
//
// This is the affirmative half of AC1's condition, and the reason it is here:
// reading the entry point proves what the image starts, and proves nothing
// about what is sitting in the image beside it. A shell that runs is how a
// project-installed dependency arrives, and is what `service/Dockerfile` claims
// this base does not have.
//
// A run that fails for a reason this harness cannot attribute to the path being
// absent is reported as a harness fault, not as a pass. Otherwise the probe
// would read as satisfied by any broken invocation, which is the same vacuous
// pass in a new place.
func assertNoShellInTheImage(t *testing.T, runtime, tag string) {
	t.Helper()

	for _, path := range shellPaths {
		out, err := exec.CommandContext(t.Context(), runtime,
			"run", "--rm", "--entrypoint", path, tag, "-c", "exit 0").CombinedOutput()
		if err == nil {
			t.Errorf("NFR-003 AC1 — %s ran to completion in the image, so the runtime image carries a shell. AC1's condition is an image with no project-installed dependencies, and a shell is what installs them.",
				path)

			continue
		}

		if !reportsMissingExecutable(string(out), path) {
			t.Errorf("HARNESS FAULT — running %s as the image entry point failed for a reason this harness cannot attribute: %v\n%s\nA failure the runtime does not report as a missing executable is not evidence that the image has no shell, so this probe measured nothing.",
				path, err, out)
		}
	}
}

// reportsMissingExecutable says whether the runtime's own output attributes a
// failed run to the entry point not being there. Every runtime this harness
// drives names the path and says it could not be found; anything else is a
// failure of a different kind and is treated as one.
func reportsMissingExecutable(output, path string) bool {
	lowered := strings.ToLower(output)
	if !strings.Contains(lowered, strings.ToLower(path)) {
		return false
	}

	for _, phrase := range []string{"no such file", "not found", "executable file"} {
		if strings.Contains(lowered, phrase) {
			return true
		}
	}

	return false
}

// imageConfigList reads one string-list field of the image's configuration.
//
// The second result is false when the harness could not read the field at all.
// That is reported as a harness fault rather than as a criterion failure: a
// runtime whose CLI does not accept this format has measured the criterion
// neither way, and saying otherwise sends the reader to the wrong file.
func imageConfigList(t *testing.T, runtime, tag, field string) ([]string, bool) {
	t.Helper()

	out, err := exec.CommandContext(t.Context(), runtime,
		"image", "inspect", "--format", "{{json .Config."+field+"}}", tag).Output()
	if err != nil {
		t.Errorf("HARNESS FAULT — %s image inspect could not read .Config.%s: %v\n%s\nThe runtime may not accept this format. AC1's no-application-server condition was not checked by this run.",
			runtime, field, err, commandStderr(err))

		return nil, false
	}

	var list []string
	if err := json.Unmarshal(out, &list); err != nil {
		t.Errorf("HARNESS FAULT — could not parse .Config.%s from %s image inspect: %q: %v. AC1's no-application-server condition was not checked by this run.",
			field, runtime, out, err)

		return nil, false
	}

	return list, true
}

// assertStopsCleanlyOnSignal stops the container the way an orchestrator does —
// a signal, then a grace period — and requires the process to have exited of
// its own accord inside it.
//
// This is where the exec form of the entry point stops being a claim.
// `service/Dockerfile` uses exec form so the executable is PID 1 and receives
// SIGTERM directly, and nothing measured that. A wrapper in front of the
// executable is what swallows the signal, and it shows up here as a container
// the runtime has to kill at the end of the grace period, exiting with a
// signal's status instead of 0.
//
// It runs before the forced removal registered as this test's cleanup. That
// removal is teardown for the paths that fail earlier and is evidence of
// nothing; a container killed outright can be removed just as cleanly as one
// that stopped by itself, which is why the graceful stop has to be its own
// assertion here rather than left to the cleanup.
func assertStopsCleanlyOnSignal(t *testing.T, runtime, container string) {
	t.Helper()

	start := time.Now()

	if out, err := exec.CommandContext(t.Context(), runtime,
		"stop", "--time", strconv.Itoa(int(stopGrace/time.Second)), container).CombinedOutput(); err != nil {
		t.Errorf("HARNESS FAULT — %s stop did not complete for the running container: %v\n%s\nWhether the entry point exits on a signal was not measured by this run.",
			runtime, err, out)

		return
	}

	elapsed := time.Since(start)

	code, ok := containerExitCode(t, runtime, container)
	if !ok {
		return
	}

	if code != 0 {
		t.Errorf("NFR-003 AC1 — the container exited %d after being signalled, want 0: the executable is what the image starts, so the signal reaches it directly, and a non-zero status is what a wrapper swallowing the signal looks like at stop time\ncontainer logs:\n%s",
			code, containerLogs(runtime, container))
	}

	if elapsed >= stopGrace {
		t.Errorf("NFR-003 AC1 — the container took %s to stop within a grace period of %s, so the runtime had to force it rather than the entry point exiting on the signal",
			elapsed.Round(time.Millisecond), stopGrace)
	}
}

// containerExitCode reads the status a stopped container exited with.
func containerExitCode(t *testing.T, runtime, container string) (int, bool) {
	t.Helper()

	out, err := exec.CommandContext(t.Context(), runtime,
		"inspect", "--format", "{{.State.ExitCode}}", container).Output()
	if err != nil {
		t.Errorf("HARNESS FAULT — %s inspect could not read the stopped container's exit code: %v\n%s\nWhether the entry point exits on a signal was not measured by this run.",
			runtime, err, commandStderr(err))

		return 0, false
	}

	code, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		t.Errorf("HARNESS FAULT — could not parse the exit code %q reported by %s inspect: %v. Whether the entry point exits on a signal was not measured by this run.",
			out, runtime, err)

		return 0, false
	}

	return code, true
}

// publishedAddr returns the host address the container's port is published on.
//
// Every failure below is the harness failing to learn where to send a request,
// which is a different thing from the service failing to serve one. The
// container's own output is included because a container that had already
// exited looks identical from here, and that is the one reading of a harness
// fault that would be a real AC2 failure.
func publishedAddr(t *testing.T, runtime, container string) string {
	t.Helper()

	out, err := exec.CommandContext(t.Context(), runtime, "port", container, containerPort+"/tcp").Output()
	if err != nil {
		t.Fatalf("HARNESS FAULT — %s port could not report where the container is published: %v\n%s\nAC2 was not measured by this run.\ncontainer logs:\n%s",
			runtime, err, commandStderr(err), containerLogs(runtime, container))
	}

	lines := strings.Split(strings.ReplaceAll(strings.TrimSpace(string(out)), "\r\n", "\n"), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("HARNESS FAULT — %s port printed no address for port %s. AC2 was not measured by this run.\ncontainer logs:\n%s",
			runtime, containerPort, containerLogs(runtime, container))
	}

	_, port, err := net.SplitHostPort(strings.TrimSpace(lines[0]))
	if err != nil {
		t.Fatalf("HARNESS FAULT — the address %q printed by %s port is not one this harness can parse: %v. AC2 was not measured by this run.",
			lines[0], runtime, err)
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
			t.Fatalf("HARNESS FAULT — this harness could not build a request for %q: %v\nNo request was ever sent, so AC2's metric — the first request served in the image — was not measured by this run.",
				addr, err)
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
			t.Fatalf("NFR-003 AC2 FAILED, OR A HARNESS FAULT — this run cannot tell which, and the\n"+
				"container's own logs below are what separate the two readings.\n\n"+
				"The service in the image served no request within %s of the container starting.\n"+
				"last error: %v\n"+
				"Logs showing a service that started and is listening mean the request never reached it —\n"+
				"a host that cannot dial the port it published, a firewall, a rootless network namespace —\n"+
				"which is a missing capability of THIS HOST, and on that reading AC2's metric went\n"+
				"unmeasured rather than unmet. Logs showing a service that never started, or that exited,\n"+
				"are the service's defect and are AC2 unmet.\n\n"+
				"Red on either reading: a check that did not run is never a pass.\n"+
				"container logs:\n%s",
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
