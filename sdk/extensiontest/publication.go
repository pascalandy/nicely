package extensiontest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extensionrepo"
)

// The part ## Publication of spec.md declares each command that creates or
// replaces a user file, as Publication declaration in the contract defines.
// Its first table has a row for each mode of such a command. Its second
// table gives a reason for each obligation that no scenario proves.
var (
	publicationColumns = []string{"Command", "Mode", "Destination", "Overwrite", "Protected", "Operation", "Dry run"}
	reasonColumns      = []string{"Command", "Mode", "Obligation", "Reason"}
)

// The operations that publish a file. Each is also an ability that the file
// system of the destination needs. A create never uses a rename that
// replaces, because Generic cases forbid a check and then that rename.
var operations = []string{"link", "rename-noreplace", "rename"}

// A publication is one row of the first table: one mode of one command.
type publication struct {
	// path is the command path, such as "video convert", and mode is create
	// or replace.
	path, mode string
	// destination and overwrite are keys of the answers. The dry run and the
	// run give the path at destination, and the dry run its overwrite choice.
	destination, overwrite string
	// protected names arguments, such as file, and flags, such as --input.
	protected []string
	// operations are the operations that publish the file in this mode.
	operations []string
	// dryRun is checks or pending.
	dryRun string
	line   int
}

// obligations returns the checks that the row asks for, in order. An
// operation is the check of a destination without that ability.
func (p publication) obligations() []string {
	found := []string{"predicts"}
	if p.mode == "create" {
		found = append(found, "appears")
	}
	return append(found, p.operations...)
}

// A reason is one row of the second table. It says why no scenario proves
// an obligation of a row, or why a command publishes no file.
type reason struct {
	path, mode, obligation string
	line                   int
}

var (
	publicationCommand = regexp.MustCompile("^`ncly ([a-z0-9][a-z0-9 -]*)`$")
	publicationCode    = regexp.MustCompile("`([^`]+)`")
)

// readPublication reads ## Publication of spec.md. It returns the rows, the
// reasons, and each problem of the form.
func (k *kit) readPublication() ([]publication, []reason, []string) {
	data, err := os.ReadFile(filepath.Join(k.root, "spec.md"))
	if err != nil {
		// specLinks reports a missing spec.md.
		return nil, nil, nil
	}
	const fix = "Fix: lay out ## Publication as Publication declaration in the contract says"
	var rows []publication
	var reasons []reason
	var problems []string
	table := 0
	for _, l := range extensionrepo.SpecLines(string(data)) {
		if l.Part != "Publication" {
			table = 0
			continue
		}
		at := fmt.Sprintf("spec.md:%d: ", l.Number)
		switch {
		case l.Kind == extensionrepo.SpecHeader && slices.Equal(l.Cells, publicationColumns):
			table = 1
		case l.Kind == extensionrepo.SpecHeader && slices.Equal(l.Cells, reasonColumns):
			table = 2
		case l.Kind == extensionrepo.SpecHeader:
			table = -1
			problems = append(problems, at+"this table of ## Publication has other columns. "+fix)
		case l.Kind != extensionrepo.SpecRow || table == -1:
		case table == 1 && len(l.Cells) == len(publicationColumns):
			row, rowProblems := k.publicationRow(l.Cells, l.Number)
			problems = append(problems, rowProblems...)
			if i := slices.IndexFunc(rows, func(p publication) bool { return p.path == row.path && p.mode == row.mode }); i >= 0 {
				problems = append(problems, fmt.Sprintf("%sthe %s mode of `ncly %s` repeats the row of line %d. Fix: give each mode one row", at, row.mode, row.path, rows[i].line))
			} else if len(rowProblems) == 0 {
				rows = append(rows, row)
			}
		case table == 2 && len(l.Cells) == len(reasonColumns):
			r, problem := k.reasonRow(l.Cells, l.Number)
			if problem != "" {
				problems = append(problems, at+problem)
			} else {
				reasons = append(reasons, r)
			}
		default:
			problems = append(problems, at+"this row of ## Publication has no cell for each column of its table. "+fix)
		}
	}
	return rows, reasons, append(problems, k.publicationLinks(rows, reasons)...)
}

// reasonRow reads one row of the second table.
func (k *kit) reasonRow(cells []string, line int) (reason, string) {
	command, problem := k.publicationCommand(cells[0])
	if problem != "" {
		return reason{}, problem
	}
	r := reason{path: strings.Join(command.Path, " "), mode: cells[1], obligation: codeCell(cells[2]), line: line}
	switch {
	case r.obligation == "":
		return r, "Obligation holds no obligation in backticks. Fix: name one obligation, such as `link`"
	case strings.TrimSpace(strings.Trim(cells[3], "`*_.-")) == "":
		return r, fmt.Sprintf("the obligation `%s` has no reason. Fix: say why no scenario proves it, and what proves it instead", r.obligation)
	}
	return r, ""
}

// publicationCommand returns the running command that a cell names in
// backticks, such as `ncly video convert`.
func (k *kit) publicationCommand(cell string) (contract.Command, string) {
	m := publicationCommand.FindStringSubmatch(cell)
	if m == nil {
		return contract.Command{}, fmt.Sprintf("the cell %q names no command. Fix: write the command in backticks, such as `ncly %s`", cell, strings.Join(k.manifest.Root(), " "))
	}
	command, found := k.manifest.Command(m[1])
	if !found || !k.manifest.Runs(command) {
		return command, fmt.Sprintf("`ncly %s` is no running command of the manifest. Fix: correct the command, or delete the row", m[1])
	}
	return command, ""
}

// codeCell returns the text of a cell that holds one code span, or "".
func codeCell(cell string) string {
	if m := publicationCode.FindAllStringSubmatch(cell, -1); len(m) == 1 && strings.TrimSpace(publicationCode.ReplaceAllString(cell, "")) == "" {
		return m[0][1]
	}
	return ""
}

// codeCells returns the text of each code span of a cell.
func codeCells(cell string) []string {
	var found []string
	for _, m := range publicationCode.FindAllStringSubmatch(cell, -1) {
		found = append(found, m[1])
	}
	return found
}
