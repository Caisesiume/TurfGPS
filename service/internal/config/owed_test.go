package config_test

// Tests binding `FR-091` AC1, AC2 and AC3 to config.RequireOwed's decision:
// which configurations it refuses and which it accepts.
//
// ---------------------------------------------------------------------------
// WHY THIS IS THE EXTERNAL TEST PACKAGE.
//
// `owed_corpus_test.go` is `package config` because it must reach the registry
// variable to compare it against the document. Nothing here needs that, and
// being unable to reach it is the point: the two exported functions are the
// whole of what a caller has, so a test that can only reach them cannot come to
// depend on the registry's shape and cannot hand-copy a variable name out of
// it. Every name below is derived through config.OwedEnvVars.
//
// ---------------------------------------------------------------------------
// WHAT THIS FILE CANNOT SEE, AND WHERE THAT IS COVERED.
//
// Every criterion here says WHEN THE SYSTEM STARTS. RequireOwed's return value
// is the decision, not the start: delete its call from main.go and every
// assertion in this file still passes while the service starts and serves.
// `service/cmd/turfgps/startup_refusal_test.go` binds that half, and the split
// is deliberate — the exhaustive enumeration below costs a function call per
// case, while the same enumeration over processes would cost a build and a
// spawn per case.
//
// ---------------------------------------------------------------------------
// NO FIXTURE HERE CARRIES A PLAUSIBLE VALUE, AND NONE MAY.
//
// Neither constant the cited section records as owed has a value in any
// document, and a figure invented in a test to stand in for one is exactly the
// unexplained literal `CalculationSpecification.md § Conventions` forbids,
// arriving through the one door nobody reviews it at. RequireOwed checks
// presence and never shape, deliberately, so nothing here has to be a figure —
// and the strings below are written to be unusable as one. The precedent is
// `service/cmd/turfgps/image_test.go`, which supplies presence fixtures to the
// harness for the same reason.

import (
	"strings"
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
)

// presenceFixture is what these tests supply as a configured value. See the
// note above: it is a presence fixture and it is not a figure.
const presenceFixture = "set-by-the-acceptance-test-not-a-value"

// exhaustiveSubsetLimit bounds the power-set enumeration in the AC2 test. Two
// constants are recorded as owed today, so the enumeration is two cases; the
// bound exists so that a registry which grew unexpectedly degrades to the
// boundary families rather than generating a quarter of a million subprocesses'
// worth of subtests.
const exhaustiveSubsetLimit = 8

// configuration returns a lookup presenting exactly supplied and nothing else,
// which is what makes an assertion about acceptance meaningful: a configuration
// that carries nothing beyond the constants under test cannot be accepted
// because of something in the ambient environment.
//
// It is a parameter rather than the process environment because a refusal test
// mutating os.Environ is the one thing it cannot share with its neighbours.
func configuration(supplied map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, set := supplied[name]

		return value, set
	}
}

// TestOwedEnvVarsNamesEveryVariableRequireOwedConsults checks the premise every
// other test in this file stands on.
//
// THE BUG IT CATCHES, and it is not a hypothetical one. config.OwedEnvVars is
// the only derivation available to a caller outside this package, and two of
// them depend on it: the image harness supplies a value for each name it
// returns so the container can start, and the tables below construct their
// configurations from it. If it ever reported fewer names than RequireOwed
// consults, the harness would supply an incomplete configuration and the
// container would refuse to start — and, far worse here, every table below
// would generate its cases over the wrong set and pass having left a constant
// untested. If it reported none at all, the AC1 and AC3 tables would generate
// no configuration to speak of and the AC2 table no cases whatever: three
// criteria green over nothing measured, which is the vacuous pass
// `docs/DELIVERY.md § Proof that a test can fail` exists to prevent.
//
// So this fails rather than skips on every unusable shape, and says in each
// message what went unmeasured.
func TestOwedEnvVarsNamesEveryVariableRequireOwedConsults(t *testing.T) {
	names := config.OwedEnvVars()

	if len(names) == 0 {
		t.Fatal("config.OwedEnvVars names no variable at all. Nothing in this file measured anything: AC1 and AC3 constructed empty configurations, AC2 generated no case, and all three would report green over a registry holding nothing. `CalculationSpecification.md § Enforcement constants that do not yet exist` records the constants that must be in it, and owed_corpus_test.go is what binds the two")
	}

	seen := make(map[string]bool, len(names))

	for _, name := range names {
		if strings.TrimSpace(name) == "" {
			t.Errorf("config.OwedEnvVars returned a blank name, which no deployment can supply a value for and no operator can be told to set")
		}

		if seen[name] {
			t.Errorf("config.OwedEnvVars returned %q twice; the AC2 enumeration below treats its result as a set and a duplicate makes its subsets misreport which constants a case left unconfigured", name)
		}

		seen[name] = true
	}

	// The AC2 assertions below require that a refusal naming one variable
	// cannot be mistaken for a refusal naming another. A name contained in
	// another name breaks that, silently and in the direction that passes.
	for _, outer := range names {
		for _, inner := range names {
			if outer != inner && strings.Contains(outer, inner) {
				t.Errorf("config.OwedEnvVars returns %q and %q, and the second is contained in the first. The AC2 test below asserts that a refusal does NOT name a variable the configuration supplied, and with one name inside another that assertion cannot distinguish the two — it would pass while the refusal named the wrong constant", outer, inner)
			}
		}
	}

	// A lookup that reports every variable unset, so RequireOwed has a reason
	// to consult all of them rather than stopping at the first.
	var consulted []string

	lookup := func(name string) (string, bool) {
		consulted = append(consulted, name)

		return "", false
	}

	_ = config.RequireOwed(lookup)

	asked := make(map[string]bool, len(consulted))
	for _, name := range consulted {
		asked[name] = true
	}

	for _, name := range names {
		if !asked[name] {
			t.Errorf("config.OwedEnvVars names %q and RequireOwed never asked for it. The two have drifted apart: a deployment supplying every name OwedEnvVars reports is not thereby a configuration RequireOwed accepts, and every table in this file is built over a set that is not the one being checked", name)
		}
	}

	for _, name := range consulted {
		if !seen[name] {
			t.Errorf("RequireOwed consulted %q and config.OwedEnvVars does not name it. A deployment told to set the names OwedEnvVars reports would still be refused, over a variable nothing tells an operator about — and the image harness, which supplies exactly those names, could not start the service at all", name)
		}
	}
}

// TestRequireOwedRefusesAConfigurationSupplyingNoValue binds `FR-091` AC1: a
// configuration carrying no value for a constant recorded as owed is refused.
//
// THE BUG IT CATCHES. Unset, a constant of this class is not a check that fails
// but a check that silently never runs — `FR-091`'s Risk states it, and the
// gradient threshold is the sharp case: the comparison it feeds is false for
// every candidate, so the steep limb of `FR-076` never fires and a quarry-rim
// candidate reaches park-and-walk while every test written for `FR-076` still
// passes on its other two limbs. An implementation that accepts an unconfigured
// constant — by defaulting one, by treating an absent value as satisfied, or by
// never asking — is that state, and each of them fails here.
//
// THE BLANK CASES ARE NOT PADDING. A variable present and empty is how a
// deployment supplies nothing while appearing to configure something: an
// unexpanded template, a secret that resolved to nothing, a shell that set the
// name and not the value. An implementation testing only presence in the
// environment accepts all three and this criterion is defeated by a typo.
func TestRequireOwedRefusesAConfigurationSupplyingNoValue(t *testing.T) {
	names := config.OwedEnvVars()
	if len(names) == 0 {
		t.Fatal("config.OwedEnvVars names no variable, so this test constructed no configuration and measured nothing — see TestOwedEnvVarsNamesEveryVariableRequireOwedConsults")
	}

	for _, testCase := range []struct {
		name string

		// unset omits every variable from the configuration entirely, which is
		// a different state from one present and blank and is checked
		// separately for that reason.
		unset bool

		value string
	}{
		{name: "no variable is set at all", unset: true},
		{name: "every variable is set to an empty value", value: ""},
		{name: "every variable is set to spaces", value: "   "},
		{name: "every variable is set to a tab and a newline", value: "\t\n"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			supplied := make(map[string]string, len(names))

			if !testCase.unset {
				for _, name := range names {
					supplied[name] = testCase.value
				}
			}

			err := config.RequireOwed(configuration(supplied))
			if err == nil {
				t.Fatalf("RequireOwed accepted a configuration in which %s, want it refused: no enforcement constant recorded as owed carries a value, so every check they feed would run against nothing and silently never fire", testCase.name)
			}

			for _, name := range names {
				if !strings.Contains(err.Error(), name) {
					t.Errorf("the refusal does not name %s, so an operator is not told which value to supply. Refusal was: %v", name, err)
				}
			}
		})
	}
}

// TestRequireOwedRefusesEveryPartialConfiguration binds `FR-091` AC2: a
// configuration carrying a value for some of the constants and none for the
// rest is refused likewise.
//
// THE BUG IT CATCHES, and why it enumerates rather than sampling. A check that
// returns as soon as it finds one constant configured accepts a deployment that
// configured the other, and the criterion is defeated by whichever constant the
// registry happens to list first. Sampling one partial configuration cannot see
// that: it passes if the check happens to look at the omitted one. Every proper
// non-empty subset therefore gets a case, generated from config.OwedEnvVars so
// a third constant added to the registry is covered without this file changing.
//
// THE SECOND CASE PER SUBSET IS THE ADVERSARIAL ONE. It supplies the subset
// blank rather than filled, so a configuration that LOOKS partial is in fact
// empty, and the refusal must name every constant rather than the omitted ones
// alone. An implementation checking presence in the environment rather than a
// value reports a partial deployment as one constant short when it is two, and
// walks the operator toward a start that will refuse again.
func TestRequireOwedRefusesEveryPartialConfiguration(t *testing.T) {
	names := config.OwedEnvVars()
	if len(names) < 2 {
		t.Fatalf("config.OwedEnvVars names %d variable(s). AC2's antecedent — a value for SOME of the constants recorded as owed and none for the rest — cannot be constructed with fewer than two, so this test generated no case and measured nothing rather than passing. `CalculationSpecification.md § Enforcement constants that do not yet exist` records two; if it now records one, this criterion has lost its subject and belongs back with `@requirements-engineer` rather than being quietly green here", len(names))
	}

	for _, supplied := range partialSubsets(names) {
		for _, filling := range []struct {
			name string

			// value is what each variable IN the subset carries. A blank one
			// is supplied and yet unconfigured, so it must be refused and
			// named alongside the variables the subset omitted.
			value string

			// alsoMissing says whether the supplied subset counts as missing
			// too, which is what a blank value makes it.
			alsoMissing bool
		}{
			{name: "carries a value", value: presenceFixture},
			{name: "carries only whitespace", value: "   ", alsoMissing: true},
		} {
			t.Run(strings.Join(supplied, "+")+" "+filling.name, func(t *testing.T) {
				present := make(map[string]string, len(supplied))
				for _, name := range supplied {
					present[name] = filling.value
				}

				err := config.RequireOwed(configuration(present))
				if err == nil {
					t.Fatalf("RequireOwed accepted a configuration in which %s %s and the rest of %s carries none, want it refused: a partial configuration leaves at least one check running against nothing, and a check that never runs reports no fault",
						strings.Join(supplied, ", "), filling.name, strings.Join(names, ", "))
				}

				for _, name := range names {
					_, inSubset := present[name]
					wantNamed := !inSubset || filling.alsoMissing

					named := strings.Contains(err.Error(), name)

					switch {
					case wantNamed && !named:
						t.Errorf("%s carries no value in this configuration and the refusal does not name it, so an operator is told the deployment is closer to complete than it is. Refusal was: %v", name, err)
					case !wantNamed && named:
						t.Errorf("%s carries a value in this configuration and the refusal names it as missing, so an operator is sent to re-supply a value that is already there while the one actually missing is lost in the list. Refusal was: %v", name, err)
					}
				}
			})
		}
	}
}

// TestRequireOwedAcceptsAConfigurationSupplyingEveryValue binds `FR-091` AC3: a
// configuration carrying a value for every constant recorded as owed is
// accepted, and no journey is refused on this ground.
//
// THE BUG IT CATCHES, and why the criterion is not merely the negation of the
// other two. A refusal that cannot be satisfied is not a gate but a wall: the
// service never starts, and the pressure that lands on whoever meets it is to
// delete the check or to default a value, which is the outcome `FR-091` exists
// to prevent. It also fixes the meaning of the other two — a RequireOwed that
// refused everything would satisfy AC1 and AC2 while measuring nothing about
// the configuration it was given.
//
// The configuration presented here carries the owed constants AND NOTHING ELSE,
// so acceptance cannot come from something in the ambient environment; and the
// values are presence fixtures rather than figures, which is both what this
// lane is permitted to supply and what RequireOwed's presence-only contract
// asks for. That the shape of a value is not examined is deliberate: the cited
// section records that the gradient threshold may need to be per elevation
// provider, so a parse here would settle in code what no document has settled.
func TestRequireOwedAcceptsAConfigurationSupplyingEveryValue(t *testing.T) {
	names := config.OwedEnvVars()
	if len(names) == 0 {
		t.Fatal("config.OwedEnvVars names no variable, so the configuration this test called complete is the empty one and its acceptance says nothing — see TestOwedEnvVarsNamesEveryVariableRequireOwedConsults")
	}

	for _, testCase := range []struct {
		name  string
		value func(index int) string
	}{
		{
			name:  "every variable carries the same value",
			value: func(int) string { return presenceFixture },
		},
		{
			// Distinct per variable, so an implementation that compared the
			// values to one another rather than checking each for presence
			// cannot pass by accident.
			name:  "every variable carries a distinct value",
			value: func(index int) string { return presenceFixture + "-" + string(rune('a'+index%26)) },
		},
		{
			// Non-blank with surrounding whitespace: trimmed it is still a
			// value, and a check that trimmed the value itself into emptiness
			// would refuse a configured deployment.
			name:  "every variable carries a value padded with whitespace",
			value: func(int) string { return "  " + presenceFixture + "  " },
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			supplied := make(map[string]string, len(names))
			for index, name := range names {
				supplied[name] = testCase.value(index)
			}

			if err := config.RequireOwed(configuration(supplied)); err != nil {
				t.Fatalf("RequireOwed refused a configuration in which %s for all of %s, want it accepted: every constant recorded as owed carries a value, so nothing is refused on this ground. Refusal was: %v",
					testCase.name, strings.Join(names, ", "), err)
			}
		})
	}
}

// partialSubsets returns every configuration that supplies some of names and
// not all of them — AC2's antecedent, enumerated.
//
// It is exhaustive over the power set while that is small, which for the two
// constants recorded as owed today is two subsets. Above
// exhaustiveSubsetLimit it returns the two boundary families instead — each
// single omission and each single inclusion — because those are where a check
// that stops early or that examines only one entry actually fails, and because
// an exhaustive enumeration of a registry that large would generate more
// subtests than a run could report. The degradation is deliberate rather than
// silent: it is stated here, and the families it keeps are the ones that
// discriminate.
func partialSubsets(names []string) [][]string {
	if len(names) > exhaustiveSubsetLimit {
		var subsets [][]string

		for omitted := range names {
			supplied := make([]string, 0, len(names)-1)

			for index, name := range names {
				if index != omitted {
					supplied = append(supplied, name)
				}
			}

			subsets = append(subsets, supplied, []string{names[omitted]})
		}

		return subsets
	}

	var subsets [][]string

	// 0 supplies nothing, which is AC1's antecedent rather than AC2's, and the
	// full mask supplies everything, which is AC3's. Both bounds are excluded
	// so that every case here is genuinely partial.
	for mask := 1; mask < (1<<len(names))-1; mask++ {
		var supplied []string

		for index, name := range names {
			if mask&(1<<index) != 0 {
				supplied = append(supplied, name)
			}
		}

		subsets = append(subsets, supplied)
	}

	return subsets
}
