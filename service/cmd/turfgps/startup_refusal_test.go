package main

// Tests binding `FR-091` AC1, AC2 and AC3 at the level all three are written
// at: WHEN THE SYSTEM STARTS.
//
// ---------------------------------------------------------------------------
// WHY THIS FILE EXISTS BESIDE THE UNIT TESTS. Read before deleting it as
// duplication.
//
// `service/internal/config/owed_test.go` binds config.RequireOwed's decision,
// exhaustively and cheaply. The decision is not the criterion. Delete the
// RequireOwed call from main below, or move it under net.Listen, and every one
// of those assertions stays green while the built service starts, listens, and
// plans journeys with an implausible-gradient threshold nobody configured —
// which is precisely the state `FR-091`'s Risk names: not a check that fails
// but a check that silently never runs. Nothing that stops at the package
// boundary can see it, so this file starts the executable.
//
// Each criterion's second clause is what forces that. AC1 and AC2 say the
// configuration is refused AND NO JOURNEY IS PLANNED UNDER IT; AC3 says a
// complete one is accepted. A process that exited before reaching its listener
// plans nothing by construction, which is the strongest available reading of
// that clause and the one asserted below.
//
// ---------------------------------------------------------------------------
// WHAT "IT NEVER SERVED" IS OBSERVED BY, WHAT KEEPS THAT HONEST, AND THE ONE
// THING IT DOES NOT REACH. The limit is stated because it was MEASURED, not
// guessed at.
//
// The child's start-up record on stderr is the only thing observable between
// the process starting and it serving. A refused start prints the refusal and
// nothing else; a start that got as far as serving prints `listening`, and one
// whose bind the host refused prints `cannot listen`. So AC1 and AC2 require
// the refusal to be present and both of those to be absent.
//
// An absence is only worth as much as the guarantee that the thing would have
// been printed. What supplies that guarantee is AC3, which requires one of
// those lines to appear on a complete configuration: if either is renamed, or
// the start-up record stops carrying it, AC3 goes red rather than AC1 and AC2
// quietly weakening into assertions about a string nothing prints. That
// interlock is why all three live in one file.
//
// WHAT IT DOES NOT REACH. Moving the RequireOwed call BELOW net.Listen but
// above serve is not caught here, and that was demonstrated rather than
// reasoned: with the gate moved down, every test in this file still passed,
// because a successful bind prints nothing and the refusal exits before serve
// is entered. The variant left uncaught is a process that holds the port for
// the microseconds before it refuses and exits — it still serves nothing and
// still plans no journey, so both criteria hold on it. Closing it would mean
// occupying port 8080 to make the bind observable, which makes every case here
// fail on a host where something already holds that port. The gap is named
// instead, so nobody reads these assertions as proving an ordering they do not
// measure. What they do prove is that the process never began serving.
//
// ---------------------------------------------------------------------------
// NO FIXTURE HERE CARRIES A PLAUSIBLE VALUE, AND NONE MAY. Neither constant
// recorded as owed has a value in any document, and a figure invented in a test
// to unblock a start-up is the unexplained literal
// `CalculationSpecification.md § Conventions` forbids, arriving through the one
// door nobody reviews it at. The precedent is owedConstantFixture in
// `image_test.go`; startupFixture below is the same thing for this file, named
// after its own supplier so that neither can be read as a value.
//
// The variable names are derived through config.OwedEnvVars and never written
// out. A hand-copied list here would go stale the day a third constant is
// recorded, in a file nobody would think to look in — the staleness this item
// exists to refuse.

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
)

const (
	// refusalBudget bounds how long a refused start is given to exit. A
	// refusal happens before any I/O the process does not control, so this is
	// generous by orders of magnitude rather than tuned; reaching it means the
	// process did not exit, which is itself the failure.
	refusalBudget = 30 * time.Second

	// acceptanceBudget bounds how long an accepted start is given to reach its
	// listener. It covers a cold start of a freshly linked executable.
	acceptanceBudget = 30 * time.Second

	// recordPoll separates reads of the child's start-up record while it is
	// still running.
	recordPoll = 25 * time.Millisecond

	// refusalMarker is what main logs when config.RequireOwed refuses. The
	// error itself follows it on the same line and names every unconfigured
	// constant, which is what the assertions below read.
	refusalMarker = "refusing to start"

	// startupFixture is what this file supplies as a configured value. It is a
	// PRESENCE FIXTURE and it is not a figure — see the note at the top of this
	// file.
	startupFixture = "set-by-the-startup-refusal-test-not-a-value"
)

// listenerMarkers are the lines main prints once execution has reached
// net.Listen: one for a bind that succeeded and one for a bind the host
// refused. Either proves the start got past the refusal, which is what AC3
// needs; neither may appear on a refused start, which is what AC1 and AC2 need.
//
// Both are kept because AC3 must not depend on port 8080 being free on the host
// running it. A host that refuses the bind has still answered the only question
// this file asks of that step — whether execution reached it.
var listenerMarkers = []string{"listening", "cannot listen"}

// TestServiceRefusesToStartWithNoOwedValueSupplied binds `FR-091` AC1 at the
// start: a configuration carrying no value for a constant recorded as owed is
// refused, and no journey is planned under it.
//
// THE BUG IT CATCHES. A RequireOwed that is correct and never called. The unit
// tests cannot see it; this asserts that the built executable, given nothing,
// stops before it can serve anything at all.
func TestServiceRefusesToStartWithNoOwedValueSupplied(t *testing.T) {
	executable := buildServiceExecutable(t)
	names := owedEnvVarNames(t)

	record := runRefusedStart(t, executable, nil)

	assertRefused(t, record, names, "a configuration supplying no value for any constant recorded as owed")
}

// TestServiceRefusesToStartOnAPartialConfiguration binds `FR-091` AC2 at the
// start: a configuration carrying a value for some of the constants and none
// for the rest is refused likewise.
//
// It runs one start per single omission rather than per subset. The exhaustive
// enumeration of partial configurations belongs to
// `service/internal/config/owed_test.go`, where a case costs a function call;
// here a case costs a process, and the single omissions are the family that
// discriminates — each one is a deployment that configured everything except
// one constant, which is both the likeliest real misconfiguration and the case
// a check that stops at its first configured entry lets through.
func TestServiceRefusesToStartOnAPartialConfiguration(t *testing.T) {
	executable := buildServiceExecutable(t)
	names := owedEnvVarNames(t)

	if len(names) < 2 {
		t.Fatalf("config.OwedEnvVars names %d variable(s). AC2's antecedent — a value for SOME of the constants recorded as owed and none for the rest — cannot be constructed with fewer than two, so this test started no process and measured nothing rather than passing", len(names))
	}

	for _, omitted := range names {
		t.Run("without "+omitted, func(t *testing.T) {
			supplied := make(map[string]string, len(names)-1)

			for _, name := range names {
				if name != omitted {
					supplied[name] = startupFixture
				}
			}

			record := runRefusedStart(t, executable, supplied)

			assertRefused(t, record, []string{omitted}, "a configuration supplying a value for every constant recorded as owed except "+omitted)

			for _, name := range names {
				if name == omitted {
					continue
				}

				if strings.Contains(record, name) {
					t.Errorf("the refusal names %s, which this configuration supplied a value for. An operator reading it re-supplies a value that is already set while the one actually missing is lost in the list.\n%s", name, record)
				}
			}
		})
	}
}

// TestServiceStartsWithEveryOwedValueSupplied binds `FR-091` AC3 at the start: a
// configuration carrying a value for every constant recorded as owed is
// accepted, and no journey is refused on this ground.
//
// THE BUG IT CATCHES, and the two things it holds up on its own.
//
// A refusal that cannot be satisfied is a wall rather than a gate — the service
// never starts however it is configured, and the pressure landing on whoever
// meets that is to delete the check or to default a value, which is the outcome
// `FR-091` exists to prevent. A start-up wired to the wrong lookup fails exactly
// this way and passes every unit test in `internal/config`.
//
// It is also what makes the two tests above mean anything. An executable that
// exited 1 on every input would satisfy both of them completely; this is the
// control that separates a refusal attributable to the configuration from a
// process that simply cannot start. And it is what keeps their absence
// assertions honest: it requires a listener line to actually be printed, so a
// rename that stopped those lines matching goes red here rather than silently
// weakening them.
func TestServiceStartsWithEveryOwedValueSupplied(t *testing.T) {
	executable := buildServiceExecutable(t)
	names := owedEnvVarNames(t)

	supplied := make(map[string]string, len(names))
	for _, name := range names {
		supplied[name] = startupFixture
	}

	process := startService(t, executable, supplied)

	record, reached := process.awaitRecord(acceptanceBudget, func(record string) bool {
		return containsAny(record, listenerMarkers) || strings.Contains(record, refusalMarker)
	})

	if strings.Contains(record, refusalMarker) {
		t.Fatalf("the service refused to start under a configuration supplying a value for every constant recorded as owed (%s), want it accepted. A refusal that a complete configuration cannot satisfy is a wall rather than a gate, and no deployment can get past it.\n%s",
			strings.Join(names, ", "), record)
	}

	if !reached {
		t.Fatalf("within %s the service neither refused to start nor reached its listener under a configuration supplying a value for every constant recorded as owed (%s).\nWanted one of %q, which is what main prints once execution has passed the refusal and reached net.Listen. Nothing here observed the configuration being accepted, and the absence assertions in the two tests above rest on one of those lines being printed on this path.\n%s",
			acceptanceBudget, strings.Join(names, ", "), listenerMarkers, record)
	}
}

// assertRefused holds the two halves of AC1's and AC2's consequent against one
// start: the configuration was refused, naming each constant it left
// unconfigured, and no journey was planned under it.
//
// The second half is read as the process having exited without ever beginning
// to serve. A process that has exited serves nothing, and one whose start-up
// record carries neither listener line never got as far as a request being
// possible. What that reading does not reach is named at the top of this file.
func assertRefused(t *testing.T, record string, unconfigured []string, configuration string) {
	t.Helper()

	if !strings.Contains(record, refusalMarker) {
		t.Errorf("the service did not refuse %s. Every enforcement check fed by an unconfigured constant would run against nothing and silently never fire, which is the state `FR-091` refuses.\n%s",
			configuration, record)
	}

	for _, name := range unconfigured {
		if !strings.Contains(record, name) {
			t.Errorf("the refusal of %s does not name %s, so an operator is not told which value to supply.\n%s", configuration, name, record)
		}
	}

	for _, marker := range listenerMarkers {
		if strings.Contains(record, marker) {
			t.Errorf("the service got past the refusal under %s — its start-up record carries %q. A deployment missing an enforcement constant reached a state in which a journey can be asked for, which is what `FR-091` refuses.\n%s",
				configuration, marker, record)
		}
	}
}

// runRefusedStart starts the executable with exactly supplied and returns its
// whole start-up record, requiring it to have exited non-zero.
//
// A process that did not exit is a failure and never a wait: `FR-091` asks that
// no journey be planned under a refused configuration, and a process still
// running is one that may yet serve.
//
// IT REPORTS THAT AND KEEPS GOING RATHER THAN STOPPING THE TEST, which is not a
// stylistic preference. A service that ignores the refusal does not exit AND
// serves, and those are two different findings; a fatal here would report the
// first and take the second out of the run, leaving the caller's assertion that
// nothing was served green in every case where it matters most. The record is
// returned either way so every assertion is evaluated against what the process
// actually did.
func runRefusedStart(t *testing.T, executable string, supplied map[string]string) string {
	t.Helper()

	process := startService(t, executable, supplied)

	code, exited := process.awaitExit(refusalBudget)
	record := process.record()

	switch {
	case !exited:
		t.Errorf("the service had not exited %s after being started with a configuration that supplies no value for at least one constant recorded as owed. It is still running, so it may yet serve, and `FR-091` asks that no journey be planned under such a configuration.\n%s",
			refusalBudget, record)
	case code == 0:
		t.Errorf("the service exited 0 after being started with a configuration that supplies no value for at least one constant recorded as owed, want a non-zero status: a supervisor reading 0 restarts nothing and reports nothing, so a deployment that never ran a safety check looks like one that shut down cleanly.\n%s",
			record)
	}

	return record
}

// serviceProcess is one started executable and the start-up record it is
// writing.
type serviceProcess struct {
	stderr *lockedBuffer

	// wait memoises cmd.Wait. It is a sync.OnceValue rather than a bare method
	// so the test goroutine and the cleanup cannot call Wait concurrently,
	// which is a data race in os/exec regardless of what it prints.
	wait func() error

	// done is closed once wait has returned.
	done chan struct{}
}

// startService starts the executable with an environment supplying exactly
// supplied, and arranges for it to be stopped however the test ends.
func startService(t *testing.T, executable string, supplied map[string]string) *serviceProcess {
	t.Helper()

	// Deliberately not exec.CommandContext: cancellation would kill the child
	// with no record of why, and every wait below is already bounded by a
	// budget whose expiry is a failure with its own message.
	command := exec.Command(executable)
	command.Env = environmentSupplying(t, supplied)

	stderr := &lockedBuffer{}
	command.Stderr = stderr

	if err := command.Start(); err != nil {
		t.Fatalf("starting the built service (%s): %v", executable, err)
	}

	process := &serviceProcess{
		stderr: stderr,
		wait:   sync.OnceValue(command.Wait),
		done:   make(chan struct{}),
	}

	go func() {
		_ = process.wait()

		close(process.done)
	}()

	t.Cleanup(func() {
		// Kill rather than a signal: this runs on every path including a test
		// that has already failed, and a service that is meant to keep running
		// would otherwise outlive the run. Where the process has already
		// exited, both calls are no-ops.
		_ = command.Process.Kill()
		<-process.done
	})

	return process
}

// record returns the start-up record written so far.
func (p *serviceProcess) record() string { return p.stderr.String() }

// awaitExit waits for the process to exit, returning its status and whether it
// exited within the budget.
func (p *serviceProcess) awaitExit(budget time.Duration) (int, bool) {
	select {
	case <-p.done:
	case <-time.After(budget):
		return 0, false
	}

	err := p.wait()
	if err == nil {
		return 0, true
	}

	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode(), true
	}

	// Not an exit status at all — the process could not be waited for. Reported
	// as not-exited so the caller's own message describes what was not
	// observed, rather than a zero status being read as a clean exit.
	return 0, false
}

// awaitRecord polls the start-up record until decided reports it sufficient,
// the process exits, or the budget expires. It returns the record and whether
// decided ever accepted it.
//
// Polling rather than scanning the pipe, because what the caller asks about is
// the record as a whole and the answer must be available for a process that is
// still running and has stopped writing.
func (p *serviceProcess) awaitRecord(budget time.Duration, decided func(string) bool) (string, bool) {
	deadline := time.Now().Add(budget)

	for {
		record := p.record()
		if decided(record) {
			return record, true
		}

		select {
		case <-p.done:
			// One last read: everything the process wrote is flushed by the
			// time Wait returns, so a record that only became sufficient at
			// exit is not missed.
			record = p.record()

			return record, decided(record)
		case <-time.After(recordPoll):
		}

		if time.Now().After(deadline) {
			record = p.record()

			return record, decided(record)
		}
	}
}

// environmentSupplying builds the child's environment: this process's own, with
// every variable the registry names stripped out, plus exactly supplied.
//
// The strip is what makes a case mean what it says. A developer or a CI runner
// with one of these variables already exported would otherwise hand the child a
// configuration no test asked for, and the AC1 and AC2 cases would silently
// become AC3's — passing or failing on the host's environment rather than on
// the configuration under test.
func environmentSupplying(t *testing.T, supplied map[string]string) []string {
	t.Helper()

	names := owedEnvVarNames(t)

	registry := make(map[string]bool, len(names))
	for _, name := range names {
		registry[strings.ToUpper(name)] = true
	}

	for name := range supplied {
		if !registry[strings.ToUpper(name)] {
			t.Fatalf("this test supplies %q, which config.OwedEnvVars does not name. Either the registry changed and this file is constructing configurations over a stale set, or a variable name was written out here rather than derived", name)
		}
	}

	inherited := os.Environ()
	env := make([]string, 0, len(inherited)+len(supplied))

	for _, entry := range inherited {
		name, _, _ := strings.Cut(entry, "=")

		// Compared upper-cased because environment variable names are
		// case-insensitive on Windows, where this suite runs today: a variable
		// exported in another case is the same variable to the child and would
		// survive a case-sensitive strip.
		if registry[strings.ToUpper(name)] {
			continue
		}

		env = append(env, entry)
	}

	for name, value := range supplied {
		env = append(env, name+"="+value)
	}

	return env
}

// owedEnvVarNames returns the variables the registry names, refusing an empty
// result.
//
// An empty registry would make every configuration below the same configuration
// and every assertion about which constants a refusal named vacuous: AC1 and
// AC2 would start a process expecting a refusal over nothing, and AC3 would
// call the empty configuration complete. That is the vacuous pass
// `docs/DELIVERY.md § Proof that a test can fail` exists to prevent, so it is a
// failure here and not a skip.
func owedEnvVarNames(t *testing.T) []string {
	t.Helper()

	names := config.OwedEnvVars()
	if len(names) == 0 {
		t.Fatal("config.OwedEnvVars names no variable at all, so no test in this file constructed a configuration that differs from any other and none of AC1, AC2 or AC3 measured anything. `CalculationSpecification.md § Enforcement constants that do not yet exist` records what must be in the registry, and `service/internal/config/owed_corpus_test.go` is what binds the two")
	}

	return names
}

// buildServiceExecutable builds the service into a directory of the test's own
// and returns the path to it.
//
// It builds rather than calling main in-process because main exits the process
// on a refusal, which a test cannot survive, and because what AC1 and AC2 assert
// about a refused start — that it exits, and that it never reached its listener
// — are properties of a process rather than of a function.
func buildServiceExecutable(t *testing.T) string {
	t.Helper()

	executable := filepath.Join(t.TempDir(), "turfgps"+goEnv(t, "GOEXE")[0])

	// -trimpath matches how the Makefile and the Dockerfile build the shipped
	// artefact. No GOOS, GOARCH or CGO_ENABLED is set here, for the reason
	// stated at the top of `build_test.go`: this story may take no position on
	// linkage.
	build := exec.CommandContext(t.Context(), goTool(t),
		"build", "-trimpath", "-o", executable, "./cmd/turfgps")
	build.Dir = moduleRoot(t)

	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/turfgps in %s: %v\n%s", build.Dir, err, out)
	}

	return executable
}

// containsAny reports whether record holds any of the markers.
func containsAny(record string, markers []string) bool {
	for _, marker := range markers {
		if strings.Contains(record, marker) {
			return true
		}
	}

	return false
}

// lockedBuffer collects a child's stderr so it can be read while the child is
// still writing to it.
//
// The gate runs this suite under the race detector, and a bytes.Buffer read
// from the test goroutine while os/exec's copier writes to it is a data race
// whatever it happens to print.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}
