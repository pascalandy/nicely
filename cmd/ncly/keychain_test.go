package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
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
//   - hang never answers
//   - write-hang answers every read, then applies a write and never answers it
type testKeychain struct{}

// callRecord is one line of the test keychain.
type callRecord struct {
	Call    string `json:"call"`
	Account string `json:"account"`
	Key     string `json:"key,omitempty"`
	// Refused marks a call that changed nothing, under unavailable.
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
// returns errRefused, blocks forever, or returns nil to go on.
func (k testKeychain) record(c callRecord, write bool) error {
	mode := os.Getenv("NCLY_TEST_KEYCHAIN")
	c.Refused = mode == "unavailable"
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
	case mode == "unavailable":
		return errRefused
	case mode == "hang" || (mode == "write-hang" && write):
		select {}
	}
	return nil
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
