// Package config reads the service's operator-facing configuration.
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
type owedConstant struct {
	// documentedName is the constant's name exactly as the cited section
	// records it. It is what binds this registry to that section — see
	// `owed_corpus_test.go` — and it is what an operator is told is missing.
	documentedName string

	// envVar is where a deployment supplies the value.
	//
	// The environment is the mechanism the service's unit already reaches
	// out-of-repo material through, per
	// `DEPLOYMENT.md § Where the deployment configuration lives`, so reading
	// configuration from it opens no second delivery path. That section also
	// records that the service offers no environment variable, flag or argument
	// today: these are the first, and no document yet describes them.
	envVar string
}

// owed is the registry — every constant the cited section records as owed.
//
// It is a declared list rather than a parse of the document at run time,
// because the document is not on the host: `NFR-003` puts exactly one file
// there, and `docs/` sits outside this module where go:embed cannot reach. What
// keeps the list honest is instead `owed_corpus_test.go`, which reads the
// section and goes red when the two disagree in either direction. So the class
// is what binds, not a hand-written pair: a third constant added to that section
// fails this package until it is added here.
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
