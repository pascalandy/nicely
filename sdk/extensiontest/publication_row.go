package extensiontest

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
)

// publicationRow reads one row of the first table, and checks it against
// the manifest.
func (k *kit) publicationRow(cells []string, line int) (publication, []string) {
	at := fmt.Sprintf("spec.md:%d: ", line)
	p := publication{mode: cells[1], dryRun: cells[6], line: line}
	command, problem := k.publicationCommand(cells[0])
	if problem != "" {
		return p, []string{at + problem}
	}
	p.path = strings.Join(command.Path, " ")
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, at+fmt.Sprintf(format, args...)) }
	name := "`ncly " + p.path + "`"
	if !command.Modes.DryRun {
		add("%s writes a user file and declares no dry run. Fix: give it a dry run, because every command that writes accepts --dry-run", name)
	}
	switch p.mode {
	case "create":
	case "replace":
		if !forceSwitch(command) {
			add("%s declares the mode replace and no switch force. Fix: declare the switch force, or delete the row", name)
		}
	default:
		add("the mode %q is not create or replace. Fix: write create for a run without --force, or replace for a run with it", p.mode)
	}
	p.destination, p.overwrite = codeCell(cells[2]), codeCell(cells[3])
	for _, key := range []struct {
		name, value string
		schemas     []map[string]any
	}{
		{"Destination", p.destination, []map[string]any{command.Result.DryRunSchema, command.Result.Schema}},
		{"Overwrite", p.overwrite, []map[string]any{command.Result.DryRunSchema}},
	} {
		if key.value == "" {
			add("%s holds no key in backticks. Fix: name the key of the answer, such as `output`", key.name)
			continue
		}
		for _, schema := range key.schemas {
			if properties, _ := schema["properties"].(map[string]any); properties[key.value] == nil {
				add("the key `%s` of %s is no property of a result schema of %s. Fix: declare it in the dry-run schema and, for Destination, the schema of the run", key.value, key.name, name)
				break
			}
		}
	}
	if strings.Trim(cells[4], "`*_.") != "None" {
		p.protected = codeCells(cells[4])
		if len(p.protected) == 0 {
			add("Protected names no argument or flag in backticks, and not None. Fix: name each argument or --flag whose path stays unchanged, or write None")
		}
	}
	for _, name := range p.protected {
		flag, isFlag := strings.CutPrefix(name, "--")
		if isFlag && !slices.ContainsFunc(command.Flags, func(f contract.Flag) bool { return f.Name == flag && f.Value != "" && !f.Variadic }) ||
			!isFlag && !slices.ContainsFunc(command.Args, func(a contract.Arg) bool { return a.Name == name }) {
			add("Protected names `%s`, which is no argument and no flag with one value of %s. Fix: name a declared argument, or a flag as --name", name, "`ncly "+p.path+"`")
		}
	}
	p.operations = codeCells(cells[5])
	if len(p.operations) == 0 {
		add("Operation names no operation in backticks. Fix: name each operation that publishes the file: %s", strings.Join(operations, ", "))
	}
	for _, op := range p.operations {
		switch {
		case !slices.Contains(operations, op):
			add("the operation `%s` is unknown. Fix: name %s", op, strings.Join(operations, ", "))
		case op == "rename" && p.mode == "create":
			add("the mode create publishes with `rename`, which replaces a file that appears during the run. Fix: publish with `link` or `rename-noreplace`, as Generic cases say")
		}
	}
	if p.dryRun != "checks" && p.dryRun != "pending" {
		add("the dry-run treatment %q is not checks or pending. Fix: write checks when the dry run checks each operation as the run does, or pending", p.dryRun)
	}
	return p, problems
}

// forceSwitch reports whether a command declares the switch force, which
// selects the mode replace.
func forceSwitch(command contract.Command) bool {
	return slices.ContainsFunc(command.Flags, func(f contract.Flag) bool { return f.Name == "force" && f.Value == "" })
}
