package extensionrepo

import (
	"fmt"
	"strings"
)

// intakeView returns the part ## Intake of vision.md that just records writes
// from the intake record.
func intakeView(r intakeRecord) string {
	var b strings.Builder
	b.WriteString("## Intake\n\n`just records` writes this part from `intake.toml`. Change that file, then run `just records` again.\n\n")
	for _, f := range r.Facts {
		statement := strings.TrimSuffix(oneLine(f.Statement), ".")
		switch f.Kind {
		case "sourced":
			fmt.Fprintf(&b, "- %s. Source: %s, at %s. Read: %s\n", statement, oneLine(f.Source), oneLine(f.Revision), oneLine(f.Read))
		case "inference":
			fmt.Fprintf(&b, "- %s. Inference: no source gives it\n", statement)
		case "unknown":
			fmt.Fprintf(&b, "- %s. Unknown: see the line `%s` of `spec.md`\n", statement, oneLine(f.ID))
		default:
			fmt.Fprintf(&b, "- %s. Kind: %s\n", statement, oneLine(f.Kind))
		}
	}
	if len(r.Facts) == 0 {
		b.WriteString("None.\n")
	}
	b.WriteString("\n**Experiments.**")
	if len(r.Experiments) == 0 {
		b.WriteString(" None.\n")
		return b.String()
	}
	b.WriteString("\n\n")
	for _, e := range r.Experiments {
		fmt.Fprintf(&b, "- `%s`: %s\n", oneLine(e.Command), oneLine(e.Result))
	}
	return b.String()
}

// withIntake returns vision with its part ## Intake replaced by view. The
// part ends before the next heading of level 1 or 2, outside fenced code. It
// reports false when vision holds no such part.
func withIntake(vision, view string) (string, bool) {
	lines := strings.SplitAfter(vision, "\n")
	start, end, fence := -1, len(lines), ""
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case fence != "":
			if strings.HasPrefix(trimmed, fence) && strings.Trim(trimmed, fence[:1]) == "" {
				fence = ""
			}
			continue
		case strings.HasPrefix(trimmed, "```"), strings.HasPrefix(trimmed, "~~~"):
			fence = trimmed[:len(trimmed)-len(strings.TrimLeft(trimmed, trimmed[:1]))]
			continue
		}
		if heading := readinessHeading.FindStringSubmatch(strings.TrimRight(line, "\n")); heading != nil && len(heading[1]) <= 2 {
			if start >= 0 {
				end = i
				break
			}
			if len(heading[1]) == 2 && heading[2] == "Intake" {
				start = i
			}
		}
	}
	if start < 0 {
		return vision, false
	}
	if end < len(lines) {
		view += "\n"
	}
	return strings.Join(lines[:start], "") + view + strings.Join(lines[end:], ""), true
}

// readinessView returns readiness.md, which just records writes from the
// review record. Its first three lines name the commit, the reviewer, and the
// verdict, as the readiness review of the nicely-extension skill shows.
func readinessView(r reviewRecord) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Commit: %s\nReviewer: %s\nVerdict: %s\n\n# Readiness review\n\n", oneLine(r.Commit), oneLine(r.Reviewer), oneLine(r.Verdict))
	b.WriteString("`just records` writes this file from `review.toml`. Change that file, then run `just records` again.\n\n## Findings\n\n")
	if len(r.Findings) == 0 {
		b.WriteString("None.\n")
	}
	for i, f := range r.Findings {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "### %s %s\n\n", oneLine(f.ID), oneLine(f.Title))
		for _, part := range [][2]string{
			{"Severity", code(f.Severity)},
			{"State", code(f.State)},
			{"Spec", code(f.Spec)},
			{"Family", f.Family},
			{"Commit", f.Commit},
			{"Reason", f.Reason},
			{"Evidence", f.Evidence},
			{"Command", f.Command},
			{"Consequence", f.Consequence},
			{"Owner", f.Owner},
			{"Fix", f.Fix},
		} {
			if part[1] != "" {
				fmt.Fprintf(&b, "- **%s**: %s\n", part[0], oneLine(part[1]))
			}
		}
	}
	b.WriteString("\n## Rounds\n\n| Round | Head | Change review | Change commits | Readiness review | Finish |\n|---|---|---|---|---|---|\n")
	for i, round := range r.Rounds {
		fmt.Fprintf(&b, "| %d | %s | %s | %s | %s | %s |\n", i+1, cell(round.Head), cell(round.ChangeReview), cell(round.ChangeCommits), cell(round.ReadinessReview), cell(code(round.Finish)))
	}
	b.WriteString("\n## Order of round 1\n\n| Step | Commit |\n|---|---|\n")
	for _, step := range [][2]string{{"Intake", r.Order.Intake}, {"Spec", r.Order.Spec}, {"Scenario", r.Order.Scenario}, {"Handler", r.Order.Handler}} {
		fmt.Fprintf(&b, "| %s | %s |\n", step[0], cell(step[1]))
	}
	if run := r.FirstRun; run != nil {
		fmt.Fprintf(&b, "\n## First failing run\n\n- **Commit**: %s\n- **Command**: `just scenario %s`\n- **Exit**: %d\n- **Assertion**: %s\n",
			oneLine(run.Commit), oneLine(run.Scenario), run.Exit, oneLine(run.Assertion))
	}
	return b.String()
}

// oneLine joins the lines of a value with spaces, so a value of a record
// stays on its line of the view.
func oneLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func code(value string) string {
	if value == "" {
		return ""
	}
	return "`" + value + "`"
}

func cell(value string) string {
	return strings.ReplaceAll(oneLine(value), "|", `\|`)
}
