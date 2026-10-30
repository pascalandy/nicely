package extensiontest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pascalandy/nicely/sdk/extensionrepo"
)

// A promise is a line of spec.md that a test must cover. It is a row of a
// table, or a bullet, in a part named Failures or Acceptance. Its id is
// empty when the line starts with no ID.
type promise struct {
	id, part string
	line     int
}

// A claim is the ID that a line # spec: <id>... of a test names.
type claim struct {
	id, file string
	line     int
}

var claimLine = regexp.MustCompile(`^\s*#\s*spec:(.*)$`)

// promises returns each promise of a spec.md: each row and each bullet of
// the parts Failures and Acceptance, which extensionrepo.SpecLines reads.
func promises(spec string) []promise {
	var found []promise
	for _, l := range extensionrepo.SpecLines(spec) {
		if (l.Part == "Failures" || l.Part == "Acceptance") && (l.Kind == extensionrepo.SpecRow || l.Kind == extensionrepo.SpecBullet) {
			found = append(found, promise{id: l.ID, part: l.Part, line: l.Number})
		}
	}
	return found
}

// claims returns the IDs that the lines # spec: <id>... name, in the script
// of each scenario and in acceptance.sh.
func claims(file, text string) []claim {
	var found []claim
	for i, line := range strings.Split(text, "\n") {
		if m := claimLine.FindStringSubmatch(line); m != nil {
			for _, id := range strings.Fields(m[1]) {
				found = append(found, claim{id: id, file: file, line: i + 1})
			}
		}
	}
	return found
}

// specLinks reports these problems of spec.md and of the claims:
//   - a promise without an ID
//   - an ID that two lines share
//   - an ID that no test claims
//   - a claim of an ID that spec.md lacks
func (k *kit) specLinks(scripts []script) []string {
	data, err := os.ReadFile(filepath.Join(k.root, "spec.md"))
	if os.IsNotExist(err) {
		return []string{"spec.md: the repository has no spec. Fix: add spec.md, laid out as Extension repository in the contract says"}
	}
	if err != nil {
		return []string{err.Error()}
	}
	var found []claim
	for _, s := range scripts {
		found = append(found, claims(s.name, s.text)...)
	}
	switch acceptance, err := os.ReadFile(filepath.Join(k.root, "acceptance.sh")); {
	case err == nil:
		found = append(found, claims("acceptance.sh", string(acceptance))...)
	case !os.IsNotExist(err):
		return []string{err.Error()}
	}
	claimed := map[string]bool{}
	for _, c := range found {
		claimed[c.id] = true
	}
	var problems []string
	declared := map[string]int{}
	for _, p := range promises(string(data)) {
		switch first, seen := declared[p.id]; {
		case p.id == "":
			problems = append(problems, fmt.Sprintf("spec.md:%d: this line of %s has no ID. Fix: give it an ID in backticks, such as `still-image`, as the first cell of a row, or at the start of a bullet before a colon", p.line, p.part))
		case seen:
			problems = append(problems, fmt.Sprintf("spec.md:%d: the ID `%s` repeats the one of line %d. Fix: give each line its own ID", p.line, p.id, first))
		default:
			declared[p.id] = p.line
			if !claimed[p.id] {
				problems = append(problems, fmt.Sprintf("spec.md:%d: no test claims `%s`. Fix: write the line `# spec: %s` above the commands that prove it, in a scenario of testdata/script/ or in acceptance.sh", p.line, p.id, p.id))
			}
		}
	}
	for _, c := range found {
		if _, ok := declared[c.id]; !ok {
			problems = append(problems, fmt.Sprintf("%s:%d: `# spec: %s` names no line of Failures or Acceptance in spec.md. Fix: correct the ID, or remove it from the claim", c.file, c.line, c.id))
		}
	}
	return problems
}
