package extensionrepo

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

// finishRepository runs the finish gate of the loop of the nicely-extension
// skill, which "Done when" of the template AGENTS.md runs (#405). just check
// stays green for an incomplete round. This gate fails until the loop ends.
// Then each form is filled, no line is open, the user accepted each limit,
// and a passing readiness verdict names the commit under review.
func finishRepository() int {
	problems, spec := formProblems()
	specs, limits := specProblems(spec)
	problems = append(append(problems, specs...), readinessProblems(limits)...)
	for _, problem := range problems {
		fmt.Fprintln(os.Stderr, problem)
	}
	if len(problems) != 0 {
		return report("finish", errors.New("the loop of the nicely-extension skill is not done"))
	}
	fmt.Println("finish: no line is open, the user accepted each limit, and readiness.md holds a passing verdict for this commit")
	return 0
}

// formProblems reports each missing form of the template, and each line of
// a form that starts with "Fill:" after its indent. Round 1 of the skill
// replaces each such line (#408, #412). It also returns spec.md, or "" when
// spec.md is missing, so finish goes on to its other checks.
func formProblems() ([]string, string) {
	var problems []string
	var spec string
	for _, form := range []string{"vision.md", "spec.md", "qa.md"} {
		data, err := os.ReadFile(form)
		if errors.Is(err, os.ErrNotExist) {
			problems = append(problems, form+": the form is missing. Fix: restore it from the Git history of the repository, then fill it")
			continue
		} else if err != nil {
			problems = append(problems, form+": "+err.Error())
			continue
		}
		if form == "spec.md" {
			spec = string(data)
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "Fill:") {
				problems = append(problems, fmt.Sprintf("%s:%d: this line of the form is not filled. Fix: write what it asks for the commands of this extension, then delete it", form, i+1))
			}
		}
	}
	return problems, spec
}

// A limit is a row with a spec ID under **Limits.** of spec.md. It is
// accepted when it has the state limit and the record of the user's
// acceptance.
type limit struct {
	id       string
	line     int
	accepted bool
}

// specProblems reports each open line under ## Open lines, and each line
// under **Limits.** that is no table row with a spec ID, the state limit,
// and the record of the user's acceptance. It also returns the limits, for
// the verdict of readiness.md.
func specProblems(spec string) ([]string, []limit) {
	const freeLimit = "this limit has no spec ID, no state limit, and no acceptance. Fix: write each limit as a row of a table with the columns ID, State, Limit, and Accepted"
	var problems []string
	var limits []limit
	state, accepted := -1, -1
	for _, l := range SpecLines(spec) {
		switch l.Part {
		case "Open lines":
			if l.Kind == SpecRow || l.Kind == SpecBullet {
				problems = append(problems, fmt.Sprintf("spec.md:%d: this line is open. Fix: prove it with a scenario in a round, or ask the user to accept it as a limit", l.Number))
			}
		case "Limits":
			switch l.Kind {
			case SpecLabel:
				state, accepted = -1, -1
				if l.Rest != "" && l.Rest != "None." {
					problems = append(problems, fmt.Sprintf("spec.md:%d: %s", l.Number, freeLimit))
				}
			case SpecHeader:
				state, accepted = slices.Index(l.Cells, "State"), slices.Index(l.Cells, "Accepted")
				if state < 0 || accepted < 0 {
					problems = append(problems, fmt.Sprintf("spec.md:%d: the table of Limits has no State or Accepted column. Fix: give it the columns ID, State, Limit, and Accepted", l.Number))
				}
			case SpecRow:
				if l.ID != "" {
					limits = append(limits, limit{id: l.ID, line: l.Number})
				}
				switch {
				case l.ID == "":
					problems = append(problems, fmt.Sprintf("spec.md:%d: this limit has no spec ID. Fix: start its row with an ID in backticks", l.Number))
				case state < 0 || accepted < 0 || state >= len(l.Cells) || accepted >= len(l.Cells):
					problems = append(problems, fmt.Sprintf("spec.md:%d: the limit `%s` has no State or Accepted cell. Fix: give its table the columns ID, State, Limit, and Accepted", l.Number, l.ID))
				case strings.Trim(l.Cells[state], "`") != "limit":
					problems = append(problems, fmt.Sprintf("spec.md:%d: the limit `%s` has the state %q. Fix: give it the state `limit`", l.Number, l.ID, strings.Trim(l.Cells[state], "`")))
				case l.Cells[accepted] == "":
					problems = append(problems, fmt.Sprintf("spec.md:%d: no record says that the user accepted the limit `%s`. Fix: ask the user, then record who accepted it, when, and where", l.Number, l.ID))
				default:
					limits[len(limits)-1].accepted = true
				}
			case SpecBullet, SpecText:
				problems = append(problems, fmt.Sprintf("spec.md:%d: %s", l.Number, freeLimit))
			}
		}
	}
	return problems, limits
}
