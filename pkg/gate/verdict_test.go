package gate

import (
	"bytes"
	"strings"
	"testing"
)

func TestNoVerdictMeansNoGate(t *testing.T) {
	// A backend that predates the gate returns an empty body. That is not a
	// failure, and an upload against it must still succeed.
	var v *Verdict
	if got := v.ExitCode(true); got != ExitPassed {
		t.Fatalf("a missing verdict must not fail the build, got %d", got)
	}
	var out bytes.Buffer
	v.Render(&out)
	if out.Len() != 0 {
		t.Fatalf("a missing verdict must print nothing, got %q", out.String())
	}
}

func TestWithoutTheFlagNothingIsGated(t *testing.T) {
	// Watching what the gate would do has to be possible before it is allowed to
	// block anything, or teams meet the gate for the first time as an outage.
	v := &Verdict{Action: ActionFail}
	if got := v.ExitCode(false); got != ExitPassed {
		t.Fatalf("without --gate a failing verdict must not fail the command, got %d", got)
	}
	if got := v.ExitCode(true); got != ExitPolicyFailed {
		t.Fatalf("with --gate a failing verdict must exit %d, got %d", ExitPolicyFailed, got)
	}
}

func TestOurOwnFaultIsNeverAPolicyFailure(t *testing.T) {
	// The distinction the exit-code contract exists for. A backend that could not
	// evaluate must not red a build under a code that says the code failed.
	server := &Verdict{Action: ActionUnevaluated, ReasonClass: ClassServer, ReasonCode: "EVALUATION_ERROR"}
	if got := server.ExitCode(true); got != ExitError {
		t.Fatalf("a server-side gap must exit %d, got %d", ExitError, got)
	}
	artifact := &Verdict{Action: ActionUnevaluated, ReasonClass: "artifact", ReasonCode: "NOT_STORED"}
	if got := artifact.ExitCode(true); got != ExitUnevaluated {
		t.Fatalf("an artifact-side gap must exit %d, got %d", ExitUnevaluated, got)
	}
}

func TestRenderLeadsWithWhatToDo(t *testing.T) {
	v := &Verdict{
		Action:        ActionFail,
		PolicyEnabled: true,
		ArtifactKey:   "https://github.com/org/repo",
		Message:       "59 finding(s) at HIGH or above with a fix available; the budget is 0",
		Rules: []RuleResult{
			{Rule: "severity_budget", Found: 59, Violated: true, Detail: "59 over a budget of 0"},
			{Rule: "kev", Found: 0, Detail: "0 on CISA KEV"},
			{Rule: "secret", NotApplicable: true, Detail: "not applicable to an image"},
		},
		RunID: "run-1",
	}
	var out bytes.Buffer
	v.Render(&out)
	got := out.String()
	for _, want := range []string{"Policy: FAILED", "https://github.com/org/repo", "budget is 0", "FAIL", "n/a", "run-1"} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered verdict is missing %q:\n%s", want, got)
		}
	}
}

func TestRenderSaysWhenNothingWasStored(t *testing.T) {
	// The shape of a truncated or edited report, seen from the pipeline.
	v := &Verdict{
		Action: ActionUnevaluated, ReasonClass: "artifact", ReasonCode: "NOT_STORED",
		DeclaredPackageCount: 1311, StoredPackageCount: 0,
	}
	var out bytes.Buffer
	v.Render(&out)
	if !strings.Contains(out.String(), "declared 1311 packages but none were stored") {
		t.Fatalf("the divergence must be stated plainly:\n%s", out.String())
	}
}

func TestAPassOnAnUnenforcedTargetSaysSo(t *testing.T) {
	v := &Verdict{Action: ActionPass, PolicyEnabled: false}
	var out bytes.Buffer
	v.Render(&out)
	if !strings.Contains(out.String(), "not enforced") {
		t.Fatalf("a pass with no policy must not look like a pass against one:\n%s", out.String())
	}
}
