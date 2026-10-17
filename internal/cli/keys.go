package cli

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/pascalandy/nicely/internal/keychain"
	"github.com/pascalandy/nicely/sdk/contract"
)

// missingKeysVariable names each declared service whose key core does not
// pass to an extension, with its code, as Grants in contract.md defines.
const missingKeysVariable = "NCLY_MISSING_KEYS"

// lookupKey reads the key of a service as Keys in contract.md reads it: from
// its variable first, then from the keychain when the variable is empty. It
// returns the key and its source, or the code of its absence:
// KEYRING_UNAVAILABLE when the keychain does not answer, and AUTH_MISSING
// when it answers without the key. It returns the error of ctx when ctx
// ends before the keychain answers.
func lookupKey(ctx context.Context, k *keychain.Keychain, service string) (key, source string, code contract.Code, err error) {
	if key := os.Getenv(keyVariable(service)); key != "" {
		return key, sourceEnv, "", nil
	}
	key, found, err := k.Read(ctx, service)
	switch {
	case errors.Is(err, keychain.ErrUnavailable):
		return "", sourceNone, contract.KeyringUnavailable, nil
	case err != nil:
		return "", sourceNone, "", err
	case !found:
		return "", sourceNone, contract.AuthMissing, nil
	}
	return key, sourceKeychain, "", nil
}

// keyVars returns the variables that carry the keys of one command of an
// extension, as Grants in contract.md defines: the variable of each service
// that the command declares and that a human granted to the extension at
// its source, and NCLY_MISSING_KEYS, which names each declared service
// whose key it does not pass, in the order of the declaration, or is empty
// when it passes every one, which removes a value that ncly inherited. With
// read false, as for a dry run, a status command without --live=true, and a
// list command that completion runs, it reads no key, and names only the
// services that no human granted. A grants file that core cannot read
// grants nothing. It returns the error of ctx when ctx ends during a
// keychain read.
func keyVars(ctx context.Context, k *keychain.Keychain, e *extension, services []string, read bool) (map[string]string, error) {
	vars := map[string]string{missingKeysVariable: ""}
	if len(services) == 0 {
		return vars, nil
	}
	grants, _ := readGrants(grantsPath())
	var missing []string
	for _, s := range services {
		var code contract.Code
		switch g := findGrant(grants, s, e.domain); {
		case g == nil || !sameSource(g.Path, sourceOf(e)):
			// A grant binds the key to the source of the extension, so
			// another extension of the same domain inherits nothing.
			code = contract.KeyNotGranted
		case read:
			key, _, absent, err := lookupKey(ctx, k, s)
			if err != nil {
				return nil, err
			}
			code = absent
			if code == "" {
				vars[keyVariable(s)] = key
			}
		}
		if code != "" {
			missing = append(missing, s+"="+string(code))
		}
	}
	vars[missingKeysVariable] = strings.Join(missing, ",")
	return vars, nil
}
