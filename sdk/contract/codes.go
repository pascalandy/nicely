// Package contract holds the agent contract that every command shares: exit
// codes, the error registry, answers and their verdict, modes, and command
// declarations. docs/north-star/contract.md is the source of truth.
package contract

import (
	"slices"
	"strings"
)

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
	ExtensionInvalid     Code = "EXTENSION_INVALID"
	TerminalRequired     Code = "TERMINAL_REQUIRED"
	Runtime              Code = "RUNTIME"
	Temporary            Code = "TEMPORARY"
	Interrupted          Code = "INTERRUPTED"
	Terminated           Code = "TERMINATED"

	ConfigUnknownKey Code = "CONFIG_UNKNOWN_KEY"
)

type codeExit struct {
	code Code
	exit ExitCode
}

// errorCodes lists the error codes from the most cautious to the least, which
// is the order in which they compete to lead an answer. Each maps to one exit.
var errorCodes = []codeExit{
	{Interrupted, ExitInterrupted},
	{Terminated, ExitTerminated},
	{Runtime, ExitRuntime},
	{ConfigInvalid, ExitHuman},
	{ExtensionInvalid, ExitHuman},
	{TerminalRequired, ExitHuman},
	{UsageInvalid, ExitUsage},
	{ConfirmationRequired, ExitUsage},
	{Temporary, ExitTemporary},
}

var warningCodes = []Code{ConfigUnknownKey}

// Known reports whether e is one of the exit codes of the contract.
func (e ExitCode) Known() bool {
	return e == ExitOK || slices.ContainsFunc(errorCodes, func(c codeExit) bool { return c.exit == e })
}

// Exit returns the exit code that one of core's error codes maps to, and
// false for any other code.
func (c Code) Exit() (ExitCode, bool) { return Core.exit(c) }

// IsWarning reports whether c is one of core's warning codes.
func (c Code) IsWarning() bool { return Core.isWarning(c) }

// Registry holds the codes that an answer may use, as Error codes in
// contract.md defines: core's, and those of the extension that answers.
// Its error codes compete to lead in its order.
type Registry struct {
	errors   []codeExit
	warnings []Code
	// prefix admits every warning of an extension, such as HELLO_.
	prefix string
}

// Core is the registry of core's own commands.
var Core = Registry{errors: errorCodes, warnings: warningCodes}

// Registry returns the registry of the extension that the manifest
// declares, ordered as Output in contract.md defines: each own code comes
// after core's codes of its exit, except for exit 1, where the own codes
// come before RUNTIME. Every warning that starts with the domain and an
// underscore belongs to the extension too.
func (m Manifest) Registry() Registry {
	own := func(exit ExitCode) []codeExit {
		var codes []codeExit
		for _, e := range m.Errors {
			if e.Exit == exit {
				codes = append(codes, codeExit{e.Code, e.Exit})
			}
		}
		return codes
	}
	r := Registry{warnings: warningCodes, prefix: m.codePrefix()}
	for i, e := range errorCodes {
		if e.code == Runtime {
			r.errors = append(r.errors, own(ExitRuntime)...)
		}
		r.errors = append(r.errors, e)
		if last := i+1 == len(errorCodes) || errorCodes[i+1].exit != e.exit; last && e.exit != ExitRuntime {
			r.errors = append(r.errors, own(e.exit)...)
		}
	}
	return r
}

func (r Registry) exit(c Code) (ExitCode, bool) {
	for _, e := range r.errors {
		if e.code == c {
			return e.exit, true
		}
	}
	return 0, false
}

func (r Registry) isWarning(c Code) bool {
	return slices.Contains(r.warnings, c) || r.prefix != "" && strings.HasPrefix(string(c), r.prefix)
}

// rank orders error codes for the lead; a lower rank is more cautious.
func (r Registry) rank(c Code) int {
	for i, e := range r.errors {
		if e.code == c {
			return i
		}
	}
	return len(r.errors)
}
