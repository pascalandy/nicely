package cli

import (
	"errors"
	"os"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
)

// coreChecks checks the config files: one failed core.config check when
// one cannot be read, otherwise one warning per unknown key, or one pass
// when they hold none.
func coreChecks(p *i18n.Printer, extensions []*extension) []check {
	finding := func(status string, code contract.Code, message, hint string) check {
		return check{ID: "core.config", Component: "core", Status: status, Code: code, Message: message, Hint: hint}
	}
	_, keys, err := config.Load(config.Locate(home(), os.LookupEnv))
	if err != nil {
		file, line := "", 0
		var invalid *config.InvalidError
		if errors.As(err, &invalid) {
			file, line = invalid.File, invalid.Line
		}
		data := map[string]any{"File": file, "Line": line}
		message := p.T("doctor.config_unreadable", data)
		if line > 0 {
			message = p.T("doctor.config_invalid_line", data)
		}
		return []check{finding(statusFail, contract.ConfigInvalid, message, p.T("doctor.config_invalid_hint", data))}
	}
	var checks []check
	for _, k := range unknownKeys(keys, extensions) {
		data := map[string]any{"File": k.File, "Key": k.Name}
		checks = append(checks, finding(statusWarn, contract.ConfigUnknownKey, p.T("doctor.unknown_key", data), p.T("doctor.unknown_key_hint", data)))
	}
	if len(checks) == 0 {
		checks = append(checks, finding(statusPass, "", p.T("doctor.config_valid"), ""))
	}
	return checks
}

// unknownKeys keeps the keys that doctor warns about, as Configuration in
// contract.md defines them: a top-level key that core does not know, and a
// key of an installed extension's table that its manifest does not declare.
// It skips a table that could belong to an extension not installed here,
// with its keys, and a key beneath one that it keeps, which the warning
// about that one covers.
func unknownKeys(keys []config.Key, extensions []*extension) []config.Key {
	var kept []config.Key
	for _, k := range keys {
		if slices.ContainsFunc(kept, func(o config.Key) bool { return o.File == k.File && strings.HasPrefix(k.Name, o.Name+".") }) {
			continue
		}
		slug, rest, _ := strings.Cut(k.Name, ".")
		domain, _, _ := strings.Cut(slug, "-")
		i := slices.IndexFunc(extensions, func(e *extension) bool { return e.slug == slug })
		switch {
		case i >= 0 && rest == "" && k.Table:
			continue
		case i >= 0 && rest != "":
			// A broken manifest declares nothing that doctor can read, so its
			// keys stay its own.
			if e := extensions[i]; e.broken != nil || declares(e.manifest, k.Name) {
				continue
			}
		case i < 0 && (rest != "" || k.Table) && contract.IsSlug(slug) && !slices.Contains(reservedNames, domain):
			continue
		}
		kept = append(kept, k)
	}
	return kept
}

// declares reports whether a manifest declares key, a key beneath a
// declared one, or a table above one.
func declares(m contract.Manifest, key string) bool {
	for _, d := range m.Commands {
		for _, declared := range d.Config {
			if key == declared || strings.HasPrefix(key, declared+".") || strings.HasPrefix(declared, key+".") {
				return true
			}
		}
	}
	return false
}
