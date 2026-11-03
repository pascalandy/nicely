package extensionrepo

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

const verdictFix = "Fix: write the verdict that the findings and the limits give, as the readiness review of the nicely-extension skill says"

// verdictProblems reports a verdict that does not follow from the findings,
// their states, and the limits of spec.md (#409, #417). Finish also runs it
// on an earlier review that would end a build.
func verdictProblems(verdict string, findings []finding, limits []limit) []string {
	var problems []string
	switch verdict {
	case "Ready":
		for _, f := range findings {
			if (f.Severity == "block" || f.Severity == "warn") && f.State != "fixed" {
				problems = append(problems, fmt.Sprintf("review.toml: the verdict is Ready, but the finding %s has the severity %s and is not fixed. %s", f.ID, f.Severity, verdictFix))
			}
		}
		for _, l := range limits {
			problems = append(problems, fmt.Sprintf("spec.md:%d: the verdict is Ready, but spec.md states the limit `%s`. %s", l.line, l.id, verdictFix))
		}
	case "Ready with stated limits":
		for _, f := range findings {
			switch {
			case f.State == "fixed":
			case f.Severity == "block":
				problems = append(problems, fmt.Sprintf("review.toml: the verdict is Ready with stated limits, but the finding %s has the severity block and is not fixed. %s", f.ID, verdictFix))
			case f.Severity == "warn" && !acceptedLimit(f, limits):
				problems = append(problems, fmt.Sprintf("review.toml: the verdict is Ready with stated limits, but the warn %s is neither fixed nor an accepted limit. "+
					"Fix: give it the state limit and the spec ID of its accepted limit, or write the verdict Needs work", f.ID))
			}
		}
	case "Needs work":
		problems = append(problems, "review.toml: the verdict is Needs work. Fix: start the next round from its findings")
	default:
		problems = append(problems, fmt.Sprintf("review.toml: the verdict %q is not Ready, Ready with stated limits, or Needs work. %s", verdict, readinessFix))
	}
	return problems
}

// acceptedLimit reports whether the finding has the state limit, and names a
// limit of spec.md that the user accepted.
func acceptedLimit(f finding, limits []limit) bool {
	return f.State == "limit" && slices.ContainsFunc(limits, func(l limit) bool { return l.id == f.Spec && l.accepted })
}

// limitsOnly reports whether spec.md differs from its version in commit only
// under **Limits.**. So the user accepts a limit after a review, and no new
// round starts for it (#417).
func limitsOnly(commit string) bool {
	old, err := exec.Command("git", "show", commit+":spec.md").Output()
	if err != nil {
		return false
	}
	current, err := os.ReadFile("spec.md")
	return err == nil && sameOutsideLimits(string(old), string(current))
}

// sameOutsideLimits reports whether two texts of spec.md differ only under
// **Limits.**, and under ## Open lines by the rows of the limits of b. So
// an open finding that the user accepts moves its row from ## Open lines to
// **Limits.** with no new round. Other lines of ## Open lines, such as its
// header or None., hold no open work.
func sameOutsideLimits(a, b string) bool {
	accepted := map[string]bool{}
	for _, l := range SpecLines(b) {
		if l.Part == "Limits" && l.Kind == SpecRow && l.ID != "" {
			accepted[l.ID] = true
		}
	}
	outside := func(spec string) []string {
		lines := strings.Split(spec, "\n")
		var kept []string
		for _, l := range SpecLines(spec) {
			open := l.Kind == SpecRow || l.Kind == SpecBullet
			if l.Part != "Limits" && (l.Part != "Open lines" || open && !accepted[l.ID]) {
				kept = append(kept, lines[l.Number-1])
			}
		}
		return kept
	}
	return slices.Equal(outside(a), outside(b))
}
