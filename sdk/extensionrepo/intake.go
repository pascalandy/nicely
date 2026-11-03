package extensionrepo

import (
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

// specID is the form of a spec ID, as Spec IDs in the contract defines it.
var specID = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)

// intakeProblems checks the facts and the experiments of the intake record
// (T24, T30, and T31 of N2). A sourced fact gives its source, its revision,
// and the part read. An inference and an unknown fact cite no source. An
// unknown fact opens a line. Its spec ID is an open line, a line of Failures
// or Acceptance, or an accepted limit. Review still judges that the part read
// supports the fact.
func intakeProblems(r intakeRecord, spec string, limits []limit) []string {
	ids := map[string]bool{}
	for _, l := range SpecLines(spec) {
		if l.ID != "" && (l.Part == "Failures" || l.Part == "Acceptance" || l.Part == "Open lines" && l.Kind == SpecRow) {
			ids[l.ID] = true
		}
	}
	for _, l := range limits {
		ids[l.id] = ids[l.id] || l.accepted
	}
	var problems []string
	for i, f := range r.Facts {
		n := i + 1
		cited := f.Source != "" || f.Revision != "" || f.Read != ""
		if f.Statement == "" {
			problems = append(problems, fmt.Sprintf("intake.toml: the fact %d has no statement. Fix: write the fact in its key statement", n))
		}
		switch f.Kind {
		case "sourced":
			if f.Source == "" || f.Revision == "" || f.Read == "" {
				problems = append(problems, fmt.Sprintf("intake.toml: the sourced fact %d has no source, no revision, or no part read. "+
					"Fix: give its source, its revision, and the part that you read in read. A link alone is not evidence of a read", n))
			}
		case "inference", "unknown":
			if cited {
				problems = append(problems, fmt.Sprintf("intake.toml: the %s fact %d names a source, a revision, or a part read. Fix: make it a sourced fact, or remove them", f.Kind, n))
			}
		default:
			problems = append(problems, fmt.Sprintf("intake.toml: the fact %d has the kind %q. Fix: give it the kind sourced, inference, or unknown", n, f.Kind))
		}
		switch {
		case f.Kind != "unknown" && f.ID != "":
			problems = append(problems, fmt.Sprintf("intake.toml: the fact %d is no unknown fact, and has an id. Fix: give an id only to an unknown fact", n))
		case f.Kind == "unknown" && f.ID == "":
			problems = append(problems, fmt.Sprintf("intake.toml: the unknown fact %d has no id. Fix: give it the spec ID of its open line in spec.md", n))
		case f.Kind == "unknown" && (!specID.MatchString(f.ID) || !ids[f.ID]):
			problems = append(problems, fmt.Sprintf("intake.toml: the unknown fact %d names the line `%s`, which is no open line, no line of Failures or Acceptance, and no accepted limit of spec.md. Fix: write it under ## Open lines, with that ID", n, f.ID))
		}
	}
	for i, e := range r.Experiments {
		if e.Command == "" || e.Result == "" {
			problems = append(problems, fmt.Sprintf("intake.toml: the experiment %d has no command or no result. Fix: give the command that ran and what it gave", i+1))
		}
	}
	return problems
}

// intakeOrderProblems reads the unknown facts of intake.toml at the Intake
// commit. Each must be a row under ## Open lines of spec.md at the Spec
// commit. So each unknown fact opens a line before the spec (T31 of N2).
func intakeOrderProblems(intake, spec string) []string {
	data, err := exec.Command("git", "show", intake+":"+intakeFile).Output()
	if err != nil {
		return []string{fmt.Sprintf("review.toml: the Intake commit %.12s holds no intake.toml. Fix: commit the intake record in the step Intake of round 1", intake)}
	}
	var r intakeRecord
	if problems := decodeRecord(intakeFile, string(data), &r); len(problems) != 0 {
		return []string{fmt.Sprintf("review.toml: the intake.toml of the Intake commit %.12s is no valid record. Fix: commit a valid intake record in the step Intake of round 1", intake)}
	}
	text, _ := exec.Command("git", "show", spec+":spec.md").Output()
	var open []string
	for _, l := range SpecLines(string(text)) {
		if l.Part == "Open lines" && l.Kind == SpecRow && l.ID != "" {
			open = append(open, l.ID)
		}
	}
	var problems []string
	for _, f := range r.Facts {
		if f.Kind == "unknown" && !slices.Contains(open, f.ID) {
			problems = append(problems, fmt.Sprintf("review.toml: the unknown fact `%s` of the Intake commit %.12s is no row under ## Open lines of spec.md at the Spec commit %.12s. "+
				"Fix: open a line for each unknown fact in the step Spec of round 1", f.ID, intake, spec))
		}
	}
	return problems
}

// intakeAdded returns the newest commit of HEAD that added intake.toml, or
// "". Copier writes the record with the example facts of the template. A new
// extension commits it in its first commit. An extension that adopts just
// finish commits it in its update commit (#427).
func intakeAdded() string {
	out, err := exec.Command("git", "log", "--diff-filter=A", "--format=%H", "-n", "1", "HEAD", "--", intakeFile).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
