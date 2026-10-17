package extension

import (
	"os"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
)

// Key returns the key of a service that the command declares, from the
// variable that core sets for a key that a human granted, or the problem
// that NCLY_MISSING_KEYS names for the service, with the hint and the next
// of Grants in contract.md, so a handler returns that problem and never
// builds one. A service whose variable is empty and that NCLY_MISSING_KEYS
// does not name, as when ncly-<domain> runs directly, gets AUTH_MISSING. A
// service that the command does not declare gets RUNTIME, because core
// never passes its key.
func (c Call) Key(service string) (string, *contract.Problem) {
	p := c.Printer
	variable := strings.ToUpper(service) + "_API_KEY"
	data := map[string]any{"Service": service, "Domain": c.domain, "Variable": variable, "Path": c.Path}
	if !slices.Contains(c.services, service) {
		return "", &contract.Problem{Code: contract.Runtime, Message: p.T("key.undeclared", data), Hint: p.T("outcome.check_hint")}
	}
	code := missingCode(os.Getenv("NCLY_MISSING_KEYS"), service)
	if key := os.Getenv(variable); code == "" && key != "" {
		return key, nil
	}
	switch code {
	case contract.KeyNotGranted:
		grant := []string{"ncly", "auth", "grant", service, c.domain}
		return "", &contract.Problem{Code: code, Message: p.T("key.not_granted", data), Hint: strings.Join(grant, " "), Next: grant}
	case contract.KeyringUnavailable:
		return "", &contract.Problem{Code: code, Message: p.T("key.unavailable", data), Hint: p.T("key.unavailable_hint", data)}
	}
	login := []string{"ncly", "auth", "login", service}
	return "", &contract.Problem{Code: contract.AuthMissing, Message: p.T("key.missing", data), Hint: strings.Join(login, " "), Next: login}
}

// missingCode returns the code that NCLY_MISSING_KEYS, missing, names for a
// service, as <service>=<code> separated by commas, or "" when it names
// none.
func missingCode(missing, service string) contract.Code {
	for entry := range strings.SplitSeq(missing, ",") {
		if name, code, found := strings.Cut(entry, "="); found && name == service {
			return contract.Code(code)
		}
	}
	return ""
}
