// Package keychain keeps keys in the OS keychain, as Keys in contract.md
// defines. Each entry uses the service nicely and the account named after
// the key's service, and each call has a time limit. Only internal/cli
// reaches it: a key reaches an extension only through grants.
package keychain

import (
	"context"
	"errors"
	"time"

	"github.com/zalando/go-keyring"
)

// Service is the keychain service of every entry that Nicely stores.
const Service = "nicely"

// probe is the account that the check of the keychain reads. keychain is
// no service name, as Keys defines, so no key lives there.
const probe = "keychain"

// The time limit of each call: InteractiveLimit in interactive mode, where a
// human can answer an unlock prompt, and Limit otherwise. The test binary
// shortens both, so a scenario proves a limit without waiting.
var (
	InteractiveLimit = 60 * time.Second
	Limit            = 10 * time.Second
)

// Backend stores the entries: the OS keychain in ncly, and a test keychain
// in the test binary alone. Get returns ErrNotFound for a missing entry.
type Backend interface {
	Get(service, account string) (string, error)
	Set(service, account, key string) error
	Delete(service, account string) error
}

// ErrNotFound is the error of a Backend whose entry does not exist.
var ErrNotFound = keyring.ErrNotFound

// ErrUnavailable means that the keychain did not answer a read in time, or
// refused it, so the command counts it unavailable for every later key.
var ErrUnavailable = errors.New("the keychain is unavailable")

// System is the OS keychain, through go-keyring. On macOS, go-keyring starts
// /usr/bin/security itself, outside the program runner, as Programs that
// ncly runs in contract.md allows.
type System struct{}

func (System) Get(service, account string) (string, error) { return keyring.Get(service, account) }

func (System) Set(service, account, key string) error { return keyring.Set(service, account, key) }

func (System) Delete(service, account string) error { return keyring.Delete(service, account) }

// Keychain asks a backend for one command. Once a call finds the keychain
// unavailable, it asks nothing more, so a command that reads several keys
// waits for one limit at most.
type Keychain struct {
	backend     Backend
	limit       time.Duration
	checked     bool
	unavailable bool
}

// New returns the keychain of one command, with the limit of its mode.
func New(b Backend, interactive bool) *Keychain {
	limit := Limit
	if interactive {
		limit = InteractiveLimit
	}
	return &Keychain{backend: b, limit: limit}
}

// Available asks the keychain once whether it answers, and reports whether
// it still counts as available. It returns the error of ctx when ctx ends
// before the keychain answers or its limit passes.
func (k *Keychain) Available(ctx context.Context) (bool, error) {
	if !k.checked && !k.unavailable {
		k.checked = true
		if _, _, err := k.Read(ctx, probe); err != nil && !errors.Is(err, ErrUnavailable) {
			return false, err
		}
	}
	return !k.unavailable, nil
}

// Read returns the key of an account and whether the keychain holds one. It
// returns ErrUnavailable without asking once the keychain is unavailable,
// and the error of ctx when ctx ends first.
func (k *Keychain) Read(ctx context.Context, account string) (string, bool, error) {
	if k.unavailable {
		return "", false, ErrUnavailable
	}
	type answer struct {
		key string
		err error
	}
	// The call may never return, so it runs apart and the command stops
	// waiting at the limit. A read writes nothing, so abandoning it is safe.
	done := make(chan answer, 1)
	go func() {
		key, err := k.backend.Get(Service, account)
		done <- answer{key, err}
	}()
	timer := time.NewTimer(k.limit)
	defer timer.Stop()
	select {
	case a := <-done:
		switch {
		case errors.Is(a.err, ErrNotFound):
			return "", false, nil
		case a.err != nil && stoppedMeanwhile(ctx):
			return "", false, ctx.Err()
		case a.err != nil:
			k.unavailable = true
			return "", false, ErrUnavailable
		}
		return a.key, a.key != "", nil
	case <-timer.C:
		k.unavailable = true
		return "", false, ErrUnavailable
	case <-ctx.Done():
		return "", false, ctx.Err()
	}
}

// signalSettle is how long a call that failed waits for a signal that ncly
// received at the same time. A Ctrl-C at the terminal reaches the program
// that go-keyring starts, such as security, and ncly together, so that
// program may fail before ncly sees its own copy, which the Go runtime
// delivers within a millisecond. The program runner waits the same way.
const signalSettle = 25 * time.Millisecond

// stoppedMeanwhile reports whether ctx ends within signalSettle of a call
// that failed, so that the call answers the signal rather than its failure.
func stoppedMeanwhile(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	case <-time.After(signalSettle):
		return false
	}
}
