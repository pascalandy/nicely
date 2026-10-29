package keychain_test

import (
	"context"
	"testing"

	"github.com/pascalandy/nicely/internal/keychain"
)

// A HOME without a default keychain makes security answer that no entry
// exists, and a write then opens a window. See #359.
func TestNoDefaultKeychainIsUnavailable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	available, err := keychain.New(keychain.System{}, false).Available(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if available {
		t.Fatal("the keychain counts as available under a HOME without a default keychain")
	}
}
