package skill

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/pascalandy/nicely/sdk/runner"
)

// installedSkill is the skill folder of an installed extension, with its
// domain.
type installedSkill struct{ domain, folder string }

// described returns the skill of each installed extension, in the order of
// ncly describe, whose entry of each extension's domain gives its folder.
// skill never searches for extensions itself, so it reads the skill of the
// extension that core runs, never one that core hides. NCLY_BIN, which
// core sets, names ncly, which starts under its own name.
func described(ctx context.Context) ([]installedSkill, error) {
	bin := os.Getenv("NCLY_BIN")
	if bin == "" {
		return nil, errors.New("NCLY_BIN is not set")
	}
	var stdout, stderr bytes.Buffer
	result, err := runner.Run(ctx, runner.Program{
		Path:   bin,
		Name:   "ncly",
		Args:   []string{"describe", "--json"},
		Env:    runner.Environ(os.Environ(), nil),
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return nil, err
	}
	if result.Signal != nil || result.Exit != 0 {
		return nil, fmt.Errorf("ncly describe exited with %d", result.Exit)
	}
	var answer struct {
		OK       bool `json:"ok"`
		Commands []struct {
			Path  string  `json:"path"`
			Skill *string `json:"skill"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &answer); err != nil || !answer.OK {
		return nil, fmt.Errorf("ncly describe gave an answer that skill cannot read: %w", err)
	}
	var skills []installedSkill
	for _, c := range answer.Commands {
		if c.Skill != nil {
			skills = append(skills, installedSkill{c.Path, *c.Skill})
		}
	}
	return skills, nil
}
