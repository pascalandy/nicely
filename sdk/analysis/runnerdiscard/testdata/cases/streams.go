package cases

import (
	stdio "io"
	"os"

	process "github.com/pascalandy/nicely/sdk/runner"
)

// Program is an alias of the SDK type, which the rule follows.
type Program = process.Program

var (
	_ = process.Program{Stdout: stdio.Discard}     // want `runner.Program.Stdout: use nil or omit the field instead of io.Discard`
	_ = Program{Stderr: stdio.Discard}             // want `runner.Program.Stderr: use nil or omit the field instead of io.Discard`
	_ = &process.Program{Stderr: stdio.Discard}    // want `runner.Program.Stderr: use nil or omit the field instead of io.Discard`
	_ = []process.Program{{Stdout: stdio.Discard}} // want `runner.Program.Stdout: use nil or omit the field instead of io.Discard`
	_ = process.Program{Stdout: nil, Stderr: os.Stderr}
	_ = process.Program{}
	_ = struct{ Stdout stdio.Writer }{Stdout: stdio.Discard}
)

func assign(p *process.Program, q process.Program) {
	p.Stdout = stdio.Discard // want `runner.Program.Stdout: use nil or omit the field instead of io.Discard`
	q.Stderr = stdio.Discard // want `runner.Program.Stderr: use nil or omit the field instead of io.Discard`
	p.Stdout, q.Stderr = nil, os.Stderr
	var other struct{ Stdout stdio.Writer }
	other.Stdout = stdio.Discard
	_, _ = q, other
}

func shadow() process.Program {
	stdio := struct{ Discard stdio.Writer }{Discard: os.Stdout}
	return process.Program{Stdout: stdio.Discard}
}
