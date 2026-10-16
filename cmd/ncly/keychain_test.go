package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pascalandy/nicely/internal/keychain"
)

// testKeychain is the keychain of the test binary: the file
// $HOME/test-keychain.jsonl, outside Nicely's cache, so that every run of ncly in a
// scenario shares it and unchanged sees a write. Each call appends one line,
// a call record, and the entries are the replay of the set and delete calls
// that took effect, so a scenario seeds a key with a set line and counts the
// calls with grep. NCLY_TEST_KEYCHAIN, which only the test binary reads,
// changes how it answers:
//
//   - unavailable refuses every call at once
//   - probe-only answers the check of the keychain, then refuses every other
//     call, as a keychain that locks meanwhile
//   - hang never answers
//   - write-hang answers every read, then applies a write and never answers it
//   - child starts a child program at each call, as go-keyring starts
//     security on macOS, and refuses the call once that program ends
type testKeychain struct{}

// callRecord is one line of the test keychain.
type callRecord struct {
	Call    string `json:"call"`
	Account string `json:"account"`
	Key     string `json:"key,omitempty"`
	// Refused marks a call that changed nothing, under unavailable and
	// probe-only.
	Refused bool `json:"refused,omitempty"`
}

var errRefused = errors.New("test keychain: unavailable")

// useTestKeychain makes ncly use the test keychain, with limits short
// enough for a scenario to wait for one.
func useTestKeychain() {
	keys = testKeychain{}
	keychain.InteractiveLimit = 2 * time.Second
	keychain.Limit = 2 * time.Second
}

func (testKeychain) file() string { return filepath.Join(os.Getenv("HOME"), "test-keychain.jsonl") }

// record appends the call, then answers as NCLY_TEST_KEYCHAIN asks: it
// returns errRefused, blocks forever, waits for its child, or returns nil
// to go on.
func (k testKeychain) record(c callRecord, write bool) error {
	mode := os.Getenv("NCLY_TEST_KEYCHAIN")
	c.Refused = mode == "unavailable" || (mode == "probe-only" && c.Account != "keychain")
	line, err := json.Marshal(c)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(k.file(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, err = f.Write(append(line, '\n'))
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	switch {
	case err != nil:
		return err
	case c.Refused:
		return errRefused
	case mode == "hang" || (mode == "write-hang" && write):
		select {}
	case mode == "child":
		return runChild()
	}
	return nil
}

// childProgram is the child of a call under child. It writes its process
// ID to keychain-child.pid once it is ready for a signal, and the name of
// the signal that ends it to keychain-child.sig. It ends on its own after a
// minute, so a failed scenario leaves nothing running.
const childProgram = `
trap 'echo TERM > "$1.sig"; exit 143' TERM
trap 'echo INT > "$1.sig"; exit 130' INT
echo $$ > "$1.tmp"
mv "$1.tmp" "$1.pid"
n=600
while [ $n -gt 0 ]; do sleep 0.1; n=$((n - 1)); done
`

// runChild runs childProgram in the process group of ncly, without the
// program runner, as go-keyring runs security, and refuses the call once
// the program ends.
func runChild() error {
	cmd := exec.Command("/bin/sh", "-c", childProgram, "sh", filepath.Join(os.Getenv("HOME"), "keychain-child"))
	// A scenario's PATH may hold ncly alone.
	cmd.Env = append(os.Environ(), "PATH=/usr/bin:/bin")
	_ = cmd.Run()
	return errRefused
}

// entries replays the calls that took effect.
func (k testKeychain) entries() (map[string]string, error) {
	entries := map[string]string{}
	f, err := os.Open(k.file())
	if errors.Is(err, os.ErrNotExist) {
		return entries, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	lines := bufio.NewScanner(f)
	for lines.Scan() {
		var c callRecord
		if err := json.Unmarshal(lines.Bytes(), &c); err != nil {
			return nil, err
		}
		switch {
		case c.Refused:
		case c.Call == "set":
			entries[c.Account] = c.Key
		case c.Call == "delete":
			delete(entries, c.Account)
		}
	}
	return entries, lines.Err()
}

func (k testKeychain) Get(_, account string) (string, error) {
	entries, err := k.entries()
	if err != nil {
		return "", err
	}
	if err := k.record(callRecord{Call: "get", Account: account}, false); err != nil {
		return "", err
	}
	key, found := entries[account]
	if !found {
		return "", keychain.ErrNotFound
	}
	return key, nil
}

func (k testKeychain) Set(_, account, key string) error {
	return k.record(callRecord{Call: "set", Account: account, Key: key}, true)
}

func (k testKeychain) Delete(_, account string) error {
	entries, err := k.entries()
	if err != nil {
		return err
	}
	if err := k.record(callRecord{Call: "delete", Account: account}, true); err != nil {
		return err
	}
	if _, found := entries[account]; !found {
		return keychain.ErrNotFound
	}
	return nil
}
