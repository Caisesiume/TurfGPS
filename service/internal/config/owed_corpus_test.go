package config

// The one thing that binds `owed` in owed.go to
// `CalculationSpecification.md § Enforcement constants that do not yet exist`.
//
// Nothing else connects them: the registry is a hand-written list in this
// module, the section is prose in a sibling directory, and neither moves when
// the other does. `FR-091`'s antecedent is the CLASS that section records
// rather than the two constants it happens to record today, and a Go slice
// cannot hold a class — so what holds it is this file going red the moment the
// two disagree.

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const (
	// specPath is the document this registry is bound to, relative to this
	// package's own directory — which is where `go test` runs a package's
	// tests from. The three levels are `service/internal/config` back to the
	// repository root: `Architecture.md § D8` puts the Go module in `service/`
	// as a peer of `docs/`, so this path is that layout and moves with it.
	specPath = "../../../docs/CalculationSpecification.md"

	// owedSectionHeading is the section whose bullets are the class `FR-091`
	// binds. Renaming it in the document without renaming it here turns this
	// test red, which is the intended warning and the only one there is: a
	// heading rename leaves no diff in this module at all.
	owedSectionHeading = "### Enforcement constants that do not yet exist"
)

// owedBullet matches a top-level bullet of that section and captures the bold
// span it opens with, which is how the section names each constant.
var owedBullet = regexp.MustCompile(`^[*-]\s+\*\*(.+?)\*\*`)

// TestOwedRegistryAgreesWithTheSection requires `owed` to hold exactly the
// constants the cited section records as owed — no more and no fewer.
//
// THE BUG IT CATCHES. Add a third constant to that section and, without this
// test, the registry still covers two: RequireOwed accepts a deployment that
// configured neither of the original pair's successor, the third check silently
// never runs, and every test written against the first two stays green. That is
// the staleness `FR-091` was widened from one constant to a class to refuse,
// and a list of two carries it by construction. It catches the other direction
// too — a constant removed from the section, or misspelled in the registry,
// leaves this module refusing to start over something no document owes.
//
// WHY EVERY UNREADABLE STATE IS A FAILURE AND NOT A SKIP. A corpus test that
// cannot find its document, its heading, or any bullet under it has compared
// nothing, and passing on that reading is worse than having no test here: this
// is the only thing standing between the registry and the section, so its green
// is read as agreement by everyone downstream. Each fatal below therefore says
// what went unmeasured in its own words.
func TestOwedRegistryAgreesWithTheSection(t *testing.T) {
	inSection := documentedOwedNames(t)

	inRegistry := make([]string, 0, len(owed))
	for _, constant := range owed {
		inRegistry = append(inRegistry, constant.documentedName)
	}

	for _, name := range missingFrom(inRegistry, inSection) {
		t.Errorf("%s records %q as owed and the registry in owed.go does not hold it, so RequireOwed accepts a configuration that supplies no value for it and the check it feeds never runs. FR-091 binds the class that section records, not the constants the registry happened to be written against",
			owedSectionHeading, name)
	}

	for _, name := range missingFrom(inSection, inRegistry) {
		t.Errorf("the registry in owed.go holds %q and %s no longer records it as owed, so the service refuses to start over a value no document owes. Read that section: either the constant was renamed there, or it now has a value and its entry belongs elsewhere",
			name, owedSectionHeading)
	}
}

// documentedOwedNames returns the constants the cited section records as owed,
// named exactly as it names them.
func documentedOwedNames(t *testing.T) []string {
	t.Helper()

	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("cannot read %s: %v\nNothing compared the registry in owed.go against the section it is bound to, so this run is evidence of neither agreement nor disagreement. If the path is wrong, `Architecture.md § D8` is the layout it is derived from.",
			specPath, err)
	}

	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")

	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == owedSectionHeading {
			start = i + 1

			break
		}
	}

	if start == -1 {
		t.Fatalf("%s carries no heading %q.\nThe section the registry in owed.go is bound to has been renamed, re-split or removed, and nothing in this module moved with it. Nothing was compared. Repair the citation here and in owed.go against whatever that section is now called, and check that FR-091's own citation still resolves.",
			specPath, owedSectionHeading)
	}

	var names []string

	for _, line := range lines[start:] {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			break
		}

		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "* ") && !strings.HasPrefix(trimmed, "- ") {
			continue
		}

		match := owedBullet.FindStringSubmatch(trimmed)
		if match == nil {
			// Reported rather than skipped: a bullet this cannot name is a
			// constant that silently leaves the comparison, which is the same
			// staleness arriving through the parser instead of the registry.
			t.Errorf("%s carries the bullet %q and this test cannot read a constant's name from it, so that bullet was left out of the comparison entirely. The section names each owed constant in a bold span opening its bullet; either that bullet is not a constant, or the convention changed and this test must change with it.",
				owedSectionHeading, trimmed)

			continue
		}

		names = append(names, match[1])
	}

	if len(names) == 0 {
		t.Fatalf("%s in %s yielded no constants.\nEither the section no longer lists them as top-level bullets or this test's reading of it is wrong. Nothing was compared, and an empty expectation would agree with any registry at all — including an empty one, which is the state FR-091 exists to refuse.",
			owedSectionHeading, specPath)
	}

	return names
}

// missingFrom returns every element of want that have is missing, sorted so a
// failure reads the same on every run.
func missingFrom(have, want []string) []string {
	held := make(map[string]bool, len(have))
	for _, name := range have {
		held[name] = true
	}

	var absent []string

	for _, name := range want {
		if !held[name] {
			absent = append(absent, name)
		}
	}

	sort.Strings(absent)

	return absent
}
