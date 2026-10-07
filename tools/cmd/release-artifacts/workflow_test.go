package main

import (
	"encoding/json"
	"maps"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestWorkflowTag(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	var workflow struct {
		Jobs map[string]struct {
			Env   map[string]string `yaml:"env"`
			Steps []struct {
				Run string            `yaml:"run"`
				Env map[string]string `yaml:"env"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(testRead(t, filepath.Join(root, ".github/workflows/release.yml")), &workflow); err != nil {
		t.Fatal(err)
	}
	var tagValue string
	for _, job := range workflow.Jobs {
		for _, step := range job.Steps {
			if strings.Contains(step.Run, "goreleaser release") {
				env := make(map[string]string)
				maps.Copy(env, job.Env)
				maps.Copy(env, step.Env)
				tagValue = env["GORELEASER_CURRENT_TAG"]
			}
		}
	}
	const requested = "v1.0.0"
	tagValue = strings.ReplaceAll(tagValue, "${{ inputs.tag }}", requested)
	tree := t.TempDir()
	testWrite(t, filepath.Join(tree, "go.mod"), []byte("module example.com/releaseproof\n\ngo 1.27.0\n"))
	testWrite(t, filepath.Join(tree, "main.go"), []byte("package main\nimport \"fmt\"\nvar version = \"dev\"\nfunc main(){fmt.Println(version)}\n"))
	testWrite(t, filepath.Join(tree, ".gitignore"), []byte("/dist/\n"))
	testWrite(t, filepath.Join(tree, ".goreleaser.yaml"), []byte("version: 2\nproject_name: proof\nbuilds:\n  - main: .\n    binary: proof\n    ldflags:\n      - -X main.version={{ .Version }}\n"))
	env := []string{"HOME=" + t.TempDir(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1", "GITHUB_TOKEN=", "GH_TOKEN=", "GITLAB_TOKEN=", "GITEA_TOKEN=", "AUR_SSH_KEY=", "HOMEBREW_TAP_SSH_KEY="}
	testCommand(t, tree, env, "git", "-c", "init.templateDir=", "init", "-q")
	testCommand(t, tree, env, "git", "add", ".")
	testCommand(t, tree, env, "git", "-c", "core.hooksPath=/dev/null", "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "fixture")
	for _, tag := range []string{requested, "v2.0.0"} {
		testCommand(t, tree, env, "git", "-c", "tag.gpgsign=false", "tag", tag)
	}
	testCommand(t, tree, env, "git", "remote", "add", "origin", "https://github.com/pascalandy/nicely.git")
	settings := exec.CommandContext(t.Context(), "go", "env", "-json", "GOVERSION", "GOCACHE", "GOMODCACHE", "GOPATH")
	data, err := settings.Output()
	if err != nil {
		t.Fatal(err)
	}
	var paths map[string]string
	if err := json.Unmarshal(data, &paths); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"GOCACHE", "GOMODCACHE", "GOPATH"} {
		env = append(env, key+"="+paths[key])
	}
	env = append(env, "GOTOOLCHAIN="+paths["GOVERSION"], "GORELEASER_CURRENT_TAG="+tagValue)
	bin := filepath.Join(t.TempDir(), "proof")
	testCommand(t, tree, env, "go", "tool", "-modfile="+filepath.Join(root, "tools/go.mod"), "goreleaser", "build", "--clean", "--single-target", "--output", bin)
	out, err := exec.CommandContext(t.Context(), bin).CombinedOutput()
	if err != nil || string(out) != "1.0.0\n" {
		t.Fatalf("workflow requested %s but native GoReleaser built %q: %v", requested, out, err)
	}
}
