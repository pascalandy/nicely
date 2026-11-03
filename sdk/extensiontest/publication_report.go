package extensiontest

import (
	"fmt"
	"slices"
	"strings"
)

// publicationLinks reports these problems:
//   - a row that the effects of its command contradict
//   - a reason for no obligation
//   - a command that may write a user file, with no row and no reason
//
// A command may write one when its effects set user_writes or are absent,
// as Manifest says.
func (k *kit) publicationLinks(rows []publication, reasons []reason) []string {
	var problems []string
	declared := map[string]bool{}
	for _, row := range rows {
		declared[row.path] = true
		if command, _ := k.manifest.Command(row.path); command.Effects != nil && !command.Effects.UserWrites {
			problems = append(problems, fmt.Sprintf("spec.md:%d: `ncly %s` publishes a file, and its effects in manifest.toml do not set user_writes. Fix: set user_writes = true", row.line, row.path))
		}
	}
	unwritten := map[string]bool{}
	for _, r := range reasons {
		if r.obligation == "publication" {
			unwritten[r.path] = true
			if declared[r.path] || r.mode != "-" {
				problems = append(problems, fmt.Sprintf("spec.md:%d: the obligation `publication` says that `ncly %s` publishes no file. Fix: give it the mode -, and keep it only for a command with no row", r.line, r.path))
			}
			continue
		}
		if !slices.ContainsFunc(rows, func(p publication) bool {
			return p.path == r.path && p.mode == r.mode && slices.Contains(p.obligations(), r.obligation)
		}) {
			problems = append(problems, fmt.Sprintf("spec.md:%d: the reason names `%s`, which is no obligation of the %s mode of `ncly %s`. Fix: name an obligation of its row: predicts, appears for create, or an operation", r.line, r.obligation, r.mode, r.path))
		}
	}
	for _, command := range k.manifest.Commands {
		path := strings.Join(command.Path, " ")
		if k.manifest.Runs(command) && (command.Effects == nil || command.Effects.UserWrites) && !declared[path] && !unwritten[path] {
			problems = append(problems, fmt.Sprintf("spec.md: `ncly %s` may write a user file, because its effects set user_writes or are absent, and ## Publication has no row for it. "+
				"Fix: give each mode of its publication a row, or give the obligation `publication` a reason when it creates or replaces no user file", path))
		}
	}
	return problems
}

// publicationOf returns the row of a mode of a command.
func (k *kit) publicationOf(path, mode string) (publication, bool) {
	i := slices.IndexFunc(k.publications, func(p publication) bool { return p.path == path && p.mode == mode })
	if i < 0 {
		return publication{}, false
	}
	return k.publications[i], true
}

// publicationMissing reports each obligation that no scenario proved and no
// reason covers, and each reason of an obligation that a scenario proved.
// Run reports them after all scenarios, with coverage.
func (k *kit) publicationMissing() []string {
	var problems []string
	for _, row := range k.publications {
		for _, obligation := range row.obligations() {
			proved := k.proved(row.path, row.mode, obligation)
			i := slices.IndexFunc(k.reasons, func(r reason) bool {
				return r.path == row.path && r.mode == row.mode && r.obligation == obligation
			})
			switch at := fmt.Sprintf("spec.md:%d: the %s mode of `ncly %s`", row.line, row.mode, row.path); {
			case proved && i >= 0:
				problems = append(problems, fmt.Sprintf("spec.md:%d: a scenario proves `%s` of the %s mode of `ncly %s`, so its reason is stale. Fix: delete the row of the reason", k.reasons[i].line, obligation, row.mode, row.path))
			case !proved && i < 0:
				problems = append(problems, fmt.Sprintf("%s has no scenario that proves `%s`, and no reason. Fix: %s", at, obligation, obligationFix(row, obligation)))
			}
		}
	}
	return problems
}

// obligationFix says which scenario command proves an obligation.
func obligationFix(row publication, obligation string) string {
	force := ""
	if row.mode == "replace" {
		force = " --force"
	}
	const reason = "give a reason in the second table of ## Publication"
	switch obligation {
	case "predicts":
		return fmt.Sprintf("run predicts ncly %s <arg>...%s in a scenario, or %s", row.path, force, reason)
	case "appears":
		return fmt.Sprintf("run appears <marker> <destination> ncly %s <arg>... in a scenario, or %s", row.path, reason)
	}
	return fmt.Sprintf("%s, such as the source review that shows the refusal. The kit has no check of a destination without `%s` yet, as COR-104-T7 plans", reason, obligation)
}
