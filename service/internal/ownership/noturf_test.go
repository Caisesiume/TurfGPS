package ownership

import (
	"os/exec"
	"slices"
	"strings"
	"testing"
)

// TestTheDecisionsNeverReachTheTurfAdapter holds `Architecture.md § Ports and
// adapters` structurally: the held-zone and lordship decisions reach neither
// the Turf adapter nor net/http, so they can issue no request of their own and
// every request they cause goes through the port they were handed.
//
// It fails closed: the adapter must itself reach net/http, or forbidding
// net/http here would forbid nothing.
func TestTheDecisionsNeverReachTheTurfAdapter(t *testing.T) {
	t.Parallel()

	const (
		module  = "github.com/Caisesiume/TurfGPS/service"
		self    = module + "/internal/ownership"
		adapter = module + "/internal/turf"
		client  = "net/http"
	)

	if !slices.Contains(deps(t, adapter), client) {
		t.Fatalf("%s no longer reaches %s, so this check would measure nothing; it refuses rather than passes", adapter, client)
	}

	own := deps(t, self)
	if len(own) == 0 {
		t.Fatalf("%s listed no dependencies, so this check would measure nothing", self)
	}

	for _, forbidden := range []string{adapter, client} {
		if slices.Contains(own, forbidden) {
			t.Errorf("%s reaches %s: the ownership decisions must consume the port, never the Turf client", self, forbidden)
		}
	}
}

// deps lists pkg's transitive imports. A go command that cannot answer is a
// failure, never a skip.
func deps(t *testing.T, pkg string) []string {
	t.Helper()

	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("cannot locate the go command: %v", err)
	}

	out, err := exec.CommandContext(t.Context(), goTool, "list", "-f", `{{join .Deps "\n"}}`, pkg).Output()
	if err != nil {
		t.Fatalf("listing %s: %v", pkg, err)
	}

	return strings.Fields(string(out))
}
