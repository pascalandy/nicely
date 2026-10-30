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
	problems, err := formProblems()
	if err != nil {
		return report("finish", err)
	}
	spec, err := os.ReadFile("spec.md")
	if err != nil {
		return report("finish", err)
	}
	problems = append(append(problems, specProblems(string(spec))...), readinessProblems()...)
	for _, problem := range problems {
		fmt.Fprintln(os.Stderr, problem)
	}
	if len(problems) != 0 {
		return report("finish", errors.New("the loop of the nicely-extension skill is not done"))
	}
	fmt.Println("finish: no line is open, the user accepted each limit, and readiness.md holds a passing verdict for this commit")
	return 0
}

// formProblems reports each line that starts with "Fill:" in the forms that
// the template writes. Round 1 of the skill replaces each one (#408).
func formProblems() ([]string, error) {
	var problems []string
	for _, form := range []string{"vision.md", "spec.md", "qa.md"} {
		data, err := os.ReadFile(form)
		if err != nil {
			return nil, err
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "Fill:") {
				problems = append(problems, fmt.Sprintf("%s:%d: this line of the form is not filled. Fix: write what it asks for the commands of this extension, then delete it", form, i+1))
			}
		}
	}
	return problems, nil
}

// specProblems reports each open line under ## Open lines, and each line
// under **Limits.** that is no table row with a spec ID, the state limit,
// and the record of the user's acceptance.
func specProblems(spec string) []string {
	const freeLimit = "this limit has no spec ID, no state limit, and no acceptance. Fix: write each limit as a row of a table with the columns ID, State, Limit, and Accepted"
	var problems []string
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
				switch {
				case l.ID == "":
					problems = append(problems, fmt.Sprintf("spec.md:%d: this limit has no spec ID. Fix: start its row with an ID in backticks", l.Number))
				case state < 0 || accepted < 0 || state >= len(l.Cells) || accepted >= len(l.Cells):
					problems = append(problems, fmt.Sprintf("spec.md:%d: the limit `%s` has no State or Accepted cell. Fix: give its table the columns ID, State, Limit, and Accepted", l.Number, l.ID))
				case strings.Trim(l.Cells[state], "`") != "limit":
					problems = append(problems, fmt.Sprintf("spec.md:%d: the limit `%s` has the state %q. Fix: give it the state `limit`", l.Number, l.ID, strings.Trim(l.Cells[state], "`")))
				case l.Cells[accepted] == "":
					problems = append(problems, fmt.Sprintf("spec.md:%d: no record says that the user accepted the limit `%s`. Fix: ask the user, then record who accepted it, when, and where", l.Number, l.ID))
				}
			case SpecBullet, SpecText:
				problems = append(problems, fmt.Sprintf("spec.md:%d: %s", l.Number, freeLimit))
			}
		}
	}
	return problems
}
