package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/platform"
)

// grantsFormat is the format_version of grants.json that this ncly reads
// and writes, as Grants in contract.md defines it.
const grantsFormat = 1

// grantsLockWait bounds the wait for grants.lock, as for the lock of a run
// in Record format of contract.md.
const grantsLockWait = 250 * time.Millisecond

// grant lets the extension of a domain, at its source, receive the key of
// a service.
type grant struct {
	Service string `json:"service"`
	Domain  string `json:"domain"`
	// Path is where discovery found the extension, links unresolved, or nil
	// for a bundled extension.
	Path      *string `json:"path"`
	GrantedAt string  `json:"granted_at"`
}

// grantsFile is the JSON of grants.json.
type grantsFile struct {
	FormatVersion int     `json:"format_version"`
	Grants        []grant `json:"grants"`
}

// grantsPath returns the path of grants.json in the state folder, or "" when
// no state folder is known, because HOME is not absolute and
// XDG_STATE_HOME is not set.
func grantsPath() string {
	state := config.Locate(home(), os.LookupEnv).State
	if state == "" {
		return ""
	}
	return filepath.Join(state, "grants.json")
}

// grantsError reports a grants file that core cannot read, which grants
// nothing, as CONFIG_INVALID.
type grantsError struct {
	file string
	// otherVersion marks a file whose format_version, version, this ncly
	// does not read.
	otherVersion bool
	version      int
	// unreadable marks a file that the system refused to read, as opposed
	// to one whose content is invalid.
	unreadable bool
}

func (e *grantsError) Error() string { return "the grants file " + e.file + " cannot be read" }

// problem is the CONFIG_INVALID of the file, whose hint names it.
func (e *grantsError) problem(p *i18n.Printer) contract.Problem {
	data := map[string]any{"File": e.file, "Version": e.version}
	message := p.T("auth.grants_invalid", data)
	switch {
	case e.file == "":
		message = p.T("auth.grants_no_state")
	case e.otherVersion:
		message = p.T("auth.grants_version", data)
	case e.unreadable:
		message = p.T("auth.grants_unreadable", data)
	}
	hint := p.T("auth.grants_hint", data)
	if e.file == "" {
		hint = p.T("auth.grants_no_state_hint")
	}
	return contract.Problem{Code: contract.ConfigInvalid, Message: message, Hint: hint}
}

// readGrants reads the grants of file. A missing file holds no grant. A
// file that cannot be read, or that holds invalid JSON, a value of the
// wrong type, a grant twice, or another format_version, returns a
// *grantsError. Readers ignore unknown keys, as Compatibility asks.
func readGrants(file string) ([]grant, error) {
	if file == "" {
		return nil, &grantsError{}
	}
	data, err := os.ReadFile(file)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, nil
	case err != nil:
		return nil, &grantsError{file: file, unreadable: true}
	}
	invalid := &grantsError{file: file}
	var top map[string]json.RawMessage
	if json.Unmarshal(data, &top) != nil || top == nil {
		return nil, invalid
	}
	var version int
	if strict(top["format_version"], &version) != nil {
		return nil, invalid
	}
	if version != grantsFormat {
		invalid.otherVersion, invalid.version = true, version
		return nil, invalid
	}
	var entries []map[string]json.RawMessage
	if strict(top["grants"], &entries) != nil || entries == nil {
		return nil, invalid
	}
	grants := make([]grant, 0, len(entries))
	for _, entry := range entries {
		var g grant
		var path string
		switch {
		case entry == nil,
			strict(entry["service"], &g.Service) != nil || !contract.IsService(g.Service),
			strict(entry["domain"], &g.Domain) != nil || !contract.IsDomain(g.Domain),
			strict(entry["granted_at"], &g.GrantedAt) != nil,
			slices.ContainsFunc(grants, func(o grant) bool { return o.Service == g.Service && o.Domain == g.Domain }):
			return nil, invalid
		}
		if _, err := time.Parse(time.RFC3339, g.GrantedAt); err != nil {
			return nil, invalid
		}
		if raw := bytes.TrimSpace(entry["path"]); !bytes.Equal(raw, []byte("null")) {
			if strict(raw, &path) != nil || !filepath.IsAbs(path) {
				return nil, invalid
			}
			g.Path = &path
		}
		grants = append(grants, g)
	}
	return grants, nil
}

// strict decodes a value that must be present and not null into v.
func strict(raw json.RawMessage, v any) error {
	if raw = bytes.TrimSpace(raw); len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return errors.New("missing")
	}
	return json.Unmarshal(raw, v)
}

// sameSource reports whether a grant's source, path, is the source of an
// extension: the same path where discovery found it, or nil for both when
// the extension is bundled.
func sameSource(path, source *string) bool {
	if path == nil || source == nil {
		return path == nil && source == nil
	}
	return *path == *source
}

// findGrant returns the grant of a service to a domain, whatever its
// source, or nil.
func findGrant(grants []grant, service, domain string) *grant {
	i := slices.IndexFunc(grants, func(g grant) bool { return g.Service == service && g.Domain == domain })
	if i < 0 {
		return nil
	}
	return &grants[i]
}

// errGrantsBusy means that another process held grants.lock past the wait,
// so nothing changed.
var errGrantsBusy = errors.New("another process changes the grants")

// grantsWriteError reports a write of the grants file that failed, whose
// effect may be unknown.
type grantsWriteError struct {
	file string
	err  error
}

func (e *grantsWriteError) Error() string { return "write " + e.file + ": " + e.err.Error() }

// changeGrants changes the grants file under grants.lock beside it, as
// Grants in contract.md defines: it reads the file again under the lock,
// gives what it holds to change, and writes the grants that change returns
// through a durable write. change returns nil to leave the file as it is,
// or an error to stop without a change. A lock that another process holds
// past grantsLockWait returns errGrantsBusy, and ctx ending first returns
// its error, both before any change.
func changeGrants(ctx context.Context, file string, change func([]grant) ([]grant, error)) error {
	if file == "" {
		return &grantsError{}
	}
	unlock, err := platform.Lock(ctx, filepath.Join(filepath.Dir(file), "grants.lock"), grantsLockWait)
	switch {
	case errors.Is(err, platform.ErrLocked):
		return errGrantsBusy
	case err != nil && ctx.Err() != nil:
		return ctx.Err()
	case err != nil:
		return &grantsWriteError{file: file, err: err}
	}
	defer unlock()
	current, err := readGrants(file)
	if err != nil {
		return err
	}
	next, err := change(current)
	if err != nil || next == nil {
		return err
	}
	// A signal that arrived while the lock was held keeps the file as it
	// was.
	if err := ctx.Err(); err != nil {
		return err
	}
	slices.SortFunc(next, func(a, b grant) int {
		return strings.Compare(a.Service+"\x00"+a.Domain, b.Service+"\x00"+b.Domain)
	})
	data, err := json.Marshal(grantsFile{FormatVersion: grantsFormat, Grants: next})
	if err != nil {
		return &grantsWriteError{file: file, err: err}
	}
	if err := platform.WriteFile(file, append(data, '\n'), 0o600); err != nil {
		return &grantsWriteError{file: file, err: err}
	}
	return nil
}

// sourceOf returns the source of an extension that a grant holds: the path
// where discovery found it, or nil for a bundled extension.
func sourceOf(e *extension) *string {
	if e.bundled {
		return nil
	}
	path := e.path
	return &path
}
