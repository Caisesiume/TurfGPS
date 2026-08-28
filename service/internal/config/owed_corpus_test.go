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

	// deploymentPath is the operator-facing document that must name a variable
	// for every constant in the registry. The same hop specPath above makes,
	// for the same reason and derived from the same layout.
	deploymentPath = "../../../docs/DEPLOYMENT.md"

	// deploymentSectionHeading is the section carrying that table.
	deploymentSectionHeading = "### Where the deployment configuration lives"

	// deploymentTableHeader opens the table itself. That section carries more
	// than one table, so the heading alone does not locate it.
	deploymentTableHeader = "| Variable |"
)

// owedBullet matches a top-level bullet of that section and captures the bold
// span it opens with, which is how the section names each constant.
//
// Its anchor is a column-0 anchor ONLY because it is matched against the raw
// line. Run against a trimmed one it accepts a bullet at any depth, which is a
// materially weaker claim — see documentedOwedNames below for what that costs.
var owedBullet = regexp.MustCompile(`^[*-]\s+\*\*(.+?)\*\*`)

// deploymentTableRow matches one row of the deployment table, capturing the
// variable from the code span the first cell holds and the constant's
// documented name from the second.
var deploymentTableRow = regexp.MustCompile("^\\|\\s*`([^`]+)`\\s*\\|\\s*(.+?)\\s*\\|\\s*$")

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

// TestDeploymentTableAgreesWithTheRegistry requires the table in
// `DEPLOYMENT.md § Where the deployment configuration lives` to carry exactly
// one row per registry entry, agreeing in BOTH cells — the variable and the
// constant's documented name.
//
// THE BUG IT CATCHES. owed.go argues against a hand-written copy of these
// names, on the grounds that it goes stale the day a third constant is added,
// in a file nobody would think to look in — and that table is such a copy. Add
// a third constant to the cited section: the registry follows, because the test
// above refuses to let it not, and the deployment document still names two. An
// operator provisions a host from that document, supplies both, and the service
// refuses to start over a variable nothing told them about — the failure
// landing on the reader with the least means to diagnose it. Rename a variable
// and the same document sends them to set one the service never reads. It
// catches the other direction too: a row the registry does not hold tells an
// operator to configure something nothing consults.
//
// WHY THE COPY IS BOUND AND NOT REPLACED BY A CITATION. An operator needs the
// variable names in the deployment document, and a citation in the table's
// place would leave them without the names and leave this test nothing to
// compare. Binding is the same remedy the test above applies to the registry,
// one document further out.
func TestDeploymentTableAgreesWithTheRegistry(t *testing.T) {
	inDocument := deploymentTableRows(t)

	inRegistry := make([]string, 0, len(owed))
	for _, constant := range owed {
		inRegistry = append(inRegistry, registryRow(constant.envVar, constant.documentedName))
	}

	for _, row := range missingFrom(inDocument, inRegistry) {
		t.Errorf("the registry in owed.go holds %s and %s in %s carries no such row, so the service refuses to start over a variable that document never tells an operator to set. Add the row: the name belongs there and no value does",
			row, deploymentSectionHeading, deploymentPath)
	}

	for _, row := range missingFrom(inRegistry, inDocument) {
		t.Errorf("%s in %s carries the row %s and the registry in owed.go holds no entry agreeing with it in both cells, so an operator is told to configure a variable nothing consults, or told the wrong constant is behind one that is. A row differing in a single cell is reported here and by the loop above both",
			deploymentSectionHeading, deploymentPath, row)
	}
}

// registryRow renders one variable-and-constant pair the way a row of that
// table reads, so the comparison is over whole rows and a failure names the row
// that is wrong rather than describing it. Rendering the document's own rows
// through it too is what makes the comparison indifferent to cell padding.
func registryRow(envVar, documentedName string) string {
	return "| `" + envVar + "` | " + documentedName + " |"
}

// deploymentTableRows returns that table's rows, rendered through registryRow.
//
// Every unreadable state below is a failure and not a skip, for the reason
// TestOwedRegistryAgreesWithTheSection gives: this is the only thing standing
// between the registry and the document an operator provisions from.
func deploymentTableRows(t *testing.T) []string {
	t.Helper()

	raw, err := os.ReadFile(deploymentPath)
	if err != nil {
		t.Fatalf("cannot read %s: %v\nNothing compared the registry in owed.go against the table an operator is given, so this run is evidence of neither agreement nor disagreement. If the path is wrong, `Architecture.md § D8` is the layout it is derived from.",
			deploymentPath, err)
	}

	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")

	start := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == deploymentSectionHeading {
			start = i + 1

			break
		}
	}

	if start == -1 {
		t.Fatalf("%s carries no heading %q.\nThe section holding the variable table has been renamed, re-split or removed, and nothing in this module moved with it. Nothing was compared. Repair the citation here and in owed.go against whatever that section is now called.",
			deploymentPath, deploymentSectionHeading)
	}

	header := -1

	for i, line := range lines[start:] {
		if strings.HasPrefix(line, "#") {
			break
		}

		// Column 0, for the reason owedBullet gives: a row of some nested block
		// is not this table.
		if strings.HasPrefix(line, deploymentTableHeader) {
			header = start + i

			break
		}
	}

	if header == -1 {
		t.Fatalf("%s in %s opens no table %q.\nNothing was compared and the registry is bound to nothing in the document an operator provisions from, which is the state this test exists to refuse. Either the table was removed — in which case an operator has lost the variable names — or its header was reworded and this constant must be reworded with it.",
			deploymentSectionHeading, deploymentPath, deploymentTableHeader)
	}

	var rows []string

	for i := header + 1; i < len(lines) && strings.HasPrefix(lines[i], "|"); i++ {
		// The delimiter row under the header is not a row of the table. Found by
		// shape rather than by offset, so nothing here assumes where it sits.
		if strings.Trim(lines[i], "|-: ") == "" {
			continue
		}

		match := deploymentTableRow.FindStringSubmatch(lines[i])
		if match == nil {
			// Reported rather than skipped, for the reason the bullet parser
			// reports: a row this cannot read leaves the comparison silently,
			// and silence here is read as agreement.
			t.Errorf("%s in %s carries the row %q and this test cannot read a variable and a constant name from it, so that row was left out of the comparison entirely. The table names the variable in a code span in its first cell and the constant in its second; either that row is not a variable, or the convention changed and this test must change with it.",
				deploymentSectionHeading, deploymentPath, lines[i])

			continue
		}

		rows = append(rows, registryRow(match[1], match[2]))
	}

	if len(rows) == 0 {
		t.Fatalf("%s in %s opens the table %q and it carries no readable row.\nNothing was compared, and an empty expectation would agree with any registry at all — including an empty one, which is the state FR-091 exists to refuse.",
			deploymentSectionHeading, deploymentPath, deploymentTableHeader)
	}

	return rows
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

		// Column 0, and never the trimmed line. A sub-bullet under an entry —
		// which is how that section would qualify one — is a note ABOUT a
		// constant and not a second constant. Read at any indentation it enters
		// the comparison as a name the registry does not hold, and the repair
		// the resulting failure asks for is to register it, after which the
		// service refuses to start over something no document owes. That is the
		// exact state the second loop of the test above calls a defect,
		// arriving here through the parser instead.
		if !strings.HasPrefix(line, "* ") && !strings.HasPrefix(line, "- ") {
			continue
		}

		match := owedBullet.FindStringSubmatch(line)
		if match == nil {
			// Reported rather than skipped: a bullet this cannot name is a
			// constant that silently leaves the comparison, which is the same
			// staleness arriving through the parser instead of the registry.
			t.Errorf("%s carries the bullet %q and this test cannot read a constant's name from it, so that bullet was left out of the comparison entirely. The section names each owed constant in a bold span opening its bullet; either that bullet is not a constant, or the convention changed and this test must change with it.",
				owedSectionHeading, line)

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

// missingFrom returns every element of want that have does not hold, sorted so
// a failure reads the same on every run.
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
