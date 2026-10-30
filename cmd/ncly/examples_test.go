package main

import (
	"path/filepath"
	"testing"

	"github.com/pascalandy/nicely/sdk/extensiontest"
	"github.com/rogpeppe/go-internal/testscript"
)

// TestManifestExamples applies the kit's example checks to first-party extensions and hello.
func TestManifestExamples(t *testing.T) {
	for _, slug := range []string{"agent", "transcript", "hello"} {
		t.Run(slug, func(t *testing.T) {
			e := fixtures[slug]
			checks, err := extensiontest.NewCoverage(e)
			if err != nil {
				t.Fatal(err)
			}
			folder := filepath.Join("..", "..", "extensions", slug)
			if slug == "hello" {
				folder = filepath.Join("..", "..", "testdata", "extensions", slug)
			}
			checks.RunExamples(t, filepath.Join(folder, "testdata", "examples.txtar"), testscript.Params{
				RequireUniqueNames: true,
				Setup: func(env *testscript.Env) error {
					if slug == "transcript" {
						if err := setupTranscript(env); err != nil {
							return err
						}
					}
					if err := setupHome(env); err != nil {
						return err
					}
					return install(env.WorkDir, env.Setenv, slug, e)
				},
				Cmds: map[string]func(*testscript.TestScript, bool, []string){"onlyncly": cmdOnlyncly},
			})
		})
	}
}
