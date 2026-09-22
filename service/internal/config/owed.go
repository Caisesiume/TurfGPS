// Package config decides whether the service's operator-facing configuration is
// one the service may start under. It reads nothing itself and imports no os:
// the caller performs the lookup and hands it in, and
// `service/cmd/turfgps/main.go` is where os.LookupEnv is reached.
//
// It does one thing so far, and deliberately only that: it refuses a
// configuration in which an enforcement constant recorded as owed carries no
// value, per `FR-091`. The wider start-time configuration shape — flags, files,
// a validated settings object — belongs to #51 and is not anticipated here.
package config

import (
	"fmt"
	"strings"
)

// owedConstant is one enforcement constant that
// `CalculationSpecification.md § Enforcement constants that do not yet exist`
// records as owed: named by the documentation set, required for the system to
// behave as specified, and carrying no value anywhere.
//
// It holds no value and no default, and none may be added. A figure authored
// here is exactly the unexplained literal
// `CalculationSpecification.md § Conventions` forbids, and the owed-constants
// section records why: the first implementer to need one of these must not be
// able to author it silently. The system being blocked until each value is
// authored, by the batch that owns it, is what `FR-091` is for rather than a
// cost of it.
//
// Nothing here exposes a value either, and #51 owns the path that will. Until
// it does, the first consumer of one of these constants cannot reach its value
// through this registry and must name the variable itself at the point of use —
// the copied name this file argues against — and a typo there leaves
// RequireOwed green while the check it feeds runs against nothing.
//
// That is the failure SHAPE `FR-091`'s Risk describes — a check that does not
// fail but silently never runs — and it is not that Risk. Its antecedent is a
// constant left unset, and on this route the value is configured and RequireOwed
// is right to pass; nothing FR-091 states is violated and no refusal it asks for
// is owed. So this is an adjacent hazard that FR-091 does not reach, and it
// stays open until #51 gives a consumer a way to read a value without naming the
// variable itself: this type binds the names a deployment must supply and binds
// no consumer that reads them.
type owedConstant struct {
	// documentedName is the constant's name exactly as the cited section
	// records it. It is what binds this registry to that section — see
	// `owed_corpus_test.go` — and it is what an operator is told is missing.
	documentedName string

	// envVar is where a deployment supplies the value.
	//
	// `DEPLOYMENT.md § Where the deployment configuration lives` records the
	// deployment-facing half of what this registry drives — that a value
	// arrives through the environment, and what the service does when one does
	// not. It records no value for either constant, because none is authored,
	// and `owed_corpus_test.go` is what holds it to this registry.
	//
	// It does not establish the environment as the only way out-of-repo
	// material reaches this process. That section requires the unit to
	// reference an out-of-repo source and records which mechanism carries it
	// as still owed; an answer that does not deliver through the environment
	// would make this variable a second path rather than the same one. Whether
	// one exists is that open question's to answer and is not decided here.
	envVar string
}

// owed is the registry — every constant the cited section records as owed.
//
// It is a declared list rather than a parse of the document at run time,
// because the document is not on the host: `NFR-003` puts exactly one file
// there, and `docs/` sits outside this module where go:embed cannot reach. What
// keeps the list honest is instead `owed_corpus_test.go`, which reads the
// section and goes red when the two disagree in either direction. So the class
// is what binds, not a hand-written pair: a third constant recorded the way that
// section records one — a bullet opening with its name in a bold span — fails
// this package until it is added here, and one recorded as that same bullet
// indented below column 0 is reported rather than dropped.
//
// The bullet is the limit of the claim and not an aside. A constant that section
// came to record in some other shape entirely is outside what that test reads,
// and this registry would not follow it; `owed_corpus_test.go` states what it
// reads and where that reading stops.
var owed = []owedConstant{
	{
		// Defined at
		// `CalculationSpecification.md § A conservative upper bound for an uncertain stop`.
		// Its strict direction, its hazard, and the feature it gates while
		// unset are stated there and under the owed-constants section; none of
		// it is restated here.
		documentedName: "The conservative upper bound for an uncertain stop",
		envVar:         "TURFGPS_UNCERTAIN_STOP_UPPER_BOUND",
	},
	{
		// Named as a member of `SPECIFICATION.md § Enforceable exclusions` and
		// reached by the last decision of the flowchart under
		// `CalculationSpecification.md § Stop time`. The owed-constants section
		// records that its shape, and not only its value, is unsettled — read
		// it there.
		documentedName: `The "implausibly steep" gradient threshold`,
		envVar:         "TURFGPS_IMPLAUSIBLE_GRADIENT_THRESHOLD",
	},
}

// RequireOwed reports whether the configuration reached through lookup carries a
// value for every constant in owed, returning an error naming all of those it
// does not.
//
// It names every missing constant rather than the first, so an operator
// configuring a deployment is not walked through them one restart at a time —
// and so a partially configured deployment reads as the partial thing it is.
//
// A value is anything non-blank: this checks presence and never shape. The cited
// section records that the gradient threshold may need to be per elevation
// provider, so a parse here would settle in code what no document has settled,
// and `FR-091` asks only whether a value is configured. A blank setting is no
// value, because an empty variable is how a deployment supplies nothing while
// appearing to configure something.
//
// lookup is os.LookupEnv in the running service. It is a parameter so that a
// test can present a configuration without mutating the process environment,
// which is the one thing a refusal test cannot share with its neighbours.
func RequireOwed(lookup func(string) (string, bool)) error {
	var missing []string

	for _, constant := range owed {
		if value, set := lookup(constant.envVar); !set || strings.TrimSpace(value) == "" {
			missing = append(missing, constant.documentedName+" ("+constant.envVar+")")
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf(
		"no configured value for %d of the %d enforcement constants recorded as owed under `CalculationSpecification.md § Enforcement constants that do not yet exist`: %s. No journey is planned under this configuration. No value may be defaulted for these; that section records what each constant is, and `FR-091` why an unset one may not be substituted for",
		len(missing), len(owed), strings.Join(missing, "; "))
}

// OwedEnvVars returns the variable each owed constant is supplied through.
//
// It exists so that nothing constructing a configuration has to write those
// names out. Every caller so far is a test that constructs one: a harness that
// must START the service in order to measure something else about it, which
// RequireOwed would otherwise refuse before that measurement could happen, and
// the refusal tests, which build partial configurations to assert which
// constants a refusal names. Deriving the names rather than writing them out is
// the same argument as the registry's own — a copied list is a second list,
// going stale the day a third constant is added, in a file nobody would think
// to look in. A count of those callers would be a third.
//
// It returns the names and never the values, because there are none.
func OwedEnvVars() []string {
	names := make([]string, 0, len(owed))
	for _, constant := range owed {
		names = append(names, constant.envVar)
	}

	return names
}
