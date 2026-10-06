// Package contract holds the agent contract that every command shares: exit
// codes, the error registry, answers and their verdict, modes, and command
// declarations. docs/north-star/cli-spec.md is the source of truth.
package contract

// ExitCode is the process exit status of an invocation.
type ExitCode int

const (
	ExitOK          ExitCode = 0
	ExitRuntime     ExitCode = 1
	ExitUsage       ExitCode = 2
	ExitTemporary   ExitCode = 75
	ExitHuman       ExitCode = 78
	ExitInterrupted ExitCode = 130
	ExitTerminated  ExitCode = 143
)

// Code is a stable machine identifier of an error or a warning.
type Code string

const (
	UsageInvalid         Code = "USAGE_INVALID"
	ConfirmationRequired Code = "CONFIRMATION_REQUIRED"
	ConfigInvalid        Code = "CONFIG_INVALID"
	TerminalRequired     Code = "TERMINAL_REQUIRED"
	Runtime              Code = "RUNTIME"
	Temporary            Code = "TEMPORARY"
	Interrupted          Code = "INTERRUPTED"
	Terminated           Code = "TERMINATED"

	ConfigUnknownKey Code = "CONFIG_UNKNOWN_KEY"
)

// errorCodes lists the error codes from the most cautious to the least, which
// is the order in which they compete to lead an answer. Each maps to one exit.
var errorCodes = []struct {
	code Code
	exit ExitCode
}{
	{Interrupted, ExitInterrupted},
	{Terminated, ExitTerminated},
	{Runtime, ExitRuntime},
	{ConfigInvalid, ExitHuman},
	{TerminalRequired, ExitHuman},
	{UsageInvalid, ExitUsage},
	{ConfirmationRequired, ExitUsage},
	{Temporary, ExitTemporary},
}

var warningCodes = []Code{ConfigUnknownKey}

// Exit returns the exit code that an error code maps to, and false for a
// code that is not a registered error.
func (c Code) Exit() (ExitCode, bool) {
	for _, e := range errorCodes {
		if e.code == c {
			return e.exit, true
		}
	}
	return 0, false
}

// IsWarning reports whether c is a registered warning code.
func (c Code) IsWarning() bool {
	for _, w := range warningCodes {
		if w == c {
			return true
		}
	}
	return false
}

// rank orders error codes for the lead; a lower rank is more cautious.
func (c Code) rank() int {
	for i, e := range errorCodes {
		if e.code == c {
			return i
		}
	}
	return len(errorCodes)
}
