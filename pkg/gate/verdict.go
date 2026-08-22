// Package gate renders the policy verdict the backend returns with an upload,
// and maps it to the exit code a pipeline reads.
package gate

import (
	"fmt"
	"io"
	"strings"
)

// Exit codes. They are a contract: a pipeline distinguishes "your code did not
// meet the policy" from "the tool could not do its job", and the two must never
// be conflated. A build that reds because our backend was busy, under a message
// saying the code failed a security policy, is how a gate gets deleted.
const (
	ExitPassed = 0
	// ExitError covers everything that is not a verdict about the code: a bad
	// config, an unreachable backend, an upload that failed, and our own
	// inability to evaluate.
	ExitError = 1
	// ExitPolicyFailed is the only code that means "the artifact did not meet
	// the policy". Gate on this one.
	ExitPolicyFailed = 2
	// ExitUnevaluated means the scan could not be judged for a reason on this
	// side — usually a report missing the git metadata the backend keys on — and
	// the policy says not to let that through.
	ExitUnevaluated = 3
)

const (
	ActionPass        = "PASS"
	ActionFail        = "FAIL"
	ActionUnevaluated = "UNEVALUATED"

	ClassServer = "server"
)

// RuleResult is one rule's outcome, as the backend reported it.
type RuleResult struct {
	Rule          string `json:"rule"`
	Limit         *int   `json:"limit"`
	Found         int    `json:"found"`
	Violated      bool   `json:"violated"`
	NotApplicable bool   `json:"notApplicable"`
	Detail        string `json:"detail"`
}

// Verdict is the "gate" object in the ingest response. Every field is optional:
// a backend that predates the gate returns no object at all, and rscli must
// treat that as "no gate configured" rather than as a failure.
type Verdict struct {
	Action               string       `json:"action"`
	ReasonClass          string       `json:"reasonClass"`
	ReasonCode           string       `json:"reasonCode"`
	Message              string       `json:"message"`
	ArtifactKind         string       `json:"artifactKind"`
	ArtifactKey          string       `json:"artifactKey"`
	PolicyEnabled        bool         `json:"policyEnabled"`
	StoredPackageCount   int          `json:"storedPackageCount"`
	StoredFindingCount   int          `json:"storedFindingCount"`
	DeclaredPackageCount int          `json:"declaredPackageCount"`
	DeclaredFindingCount int          `json:"declaredFindingCount"`
	Rules                []RuleResult `json:"rules"`
	RunID                string       `json:"runId"`
}

// Response is the ingest response body.
type Response struct {
	Gate *Verdict `json:"gate"`
}

// Render writes the verdict for a human staring at a red pipeline. It leads with
// what to do, because that is the only thing they need in the first ten seconds.
func (v *Verdict) Render(out io.Writer) {
	if v == nil {
		return
	}
	var b strings.Builder

	switch v.Action {
	case ActionFail:
		b.WriteString("\nPolicy: FAILED\n")
	case ActionUnevaluated:
		b.WriteString("\nPolicy: NOT EVALUATED\n")
	default:
		if !v.PolicyEnabled {
			b.WriteString("\nPolicy: not enforced for this target\n")
		} else {
			b.WriteString("\nPolicy: passed\n")
		}
	}

	if v.ArtifactKey != "" {
		b.WriteString(fmt.Sprintf("  target   %s\n", v.ArtifactKey))
	}
	if v.Message != "" {
		b.WriteString(fmt.Sprintf("  %s\n", v.Message))
	}

	if len(v.Rules) > 0 {
		b.WriteString("\n")
		for _, r := range v.Rules {
			mark := "ok  "
			switch {
			case r.Violated:
				mark = "FAIL"
			case r.NotApplicable:
				mark = "n/a "
			}
			b.WriteString(fmt.Sprintf("  %s  %-16s %s\n", mark, r.Rule, r.Detail))
		}
	}

	// A report that declared far more than we stored is worth saying out loud:
	// it is what a truncated or edited report looks like from here.
	if v.DeclaredPackageCount > 0 && v.StoredPackageCount == 0 {
		b.WriteString(fmt.Sprintf(
			"\n  The report declared %d packages but none were stored.\n",
			v.DeclaredPackageCount))
	}

	if v.RunID != "" {
		b.WriteString(fmt.Sprintf("\n  run %s\n", v.RunID))
	}
	fmt.Fprint(out, b.String())
}

// ExitCode maps a verdict to a process exit code.
//
// gate is whether the caller asked to be gated at all. Without it the verdict is
// printed and the command still succeeds, so a team can watch what the gate
// would do before letting it block anything — the difference between adopting a
// gate and having one imposed.
func (v *Verdict) ExitCode(gate bool) int {
	if v == nil || !gate {
		return ExitPassed
	}
	switch v.Action {
	case ActionFail:
		return ExitPolicyFailed
	case ActionUnevaluated:
		// Our fault, never theirs. A server-side gap allows the build; it is
		// reported as a tool error so it is visibly not a policy failure.
		if v.ReasonClass == ClassServer {
			return ExitError
		}
		return ExitUnevaluated
	default:
		return ExitPassed
	}
}
