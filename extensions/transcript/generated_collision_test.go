package transcript

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/record"
	"github.com/rogpeppe/go-internal/txtar"
)

// The subprocess serves the actual extension with only its generated ID fixed.
// The ordinary binary exposes no flag or environment setting for this seam.
func TestMain(m *testing.M) {
	if len(os.Args) > 2 && os.Args[1] == "collision-cli" {
		id := os.Args[2]
		newRunID = func(time.Time) string { return id }
		os.Args = append(os.Args[:1], os.Args[3:]...)
		extension.Main(Extension)
	}
	os.Exit(m.Run())
}

// Named calls are the standing control for an ID that already exists. A fresh
// call must reach the same answers when its generated ID collides after preflight.
func TestGeneratedIDCollisionsKeepExistingAnswers(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("..", "..", "testdata", "script", "transcript_named_replay.txtar"))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	for _, file := range txtar.Parse(fixture).Files {
		files[file.Name] = file.Data
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	pending := func(r *record.Record) {
		r.Status = record.Interrupted
		step := &r.Items[0].Steps[0]
		step.Status, step.StartedAt, step.FinishedAt = record.StepPending, nil, nil
		step.Artifacts, step.Evidence = []record.Artifact{}, nil
	}
	cases := []struct {
		name   string
		exit   int
		code   contract.Code
		busy   bool
		change func(*record.Record)
	}{
		{name: "busy", exit: 75, code: contract.Temporary, busy: true},
		{name: "completed replay", exit: 0},
		{name: "other URL", exit: 2, code: contract.UsageInvalid, change: func(r *record.Record) { r.Items[0].Key = "https://youtu.be/other" }},
		{name: "other command", exit: 2, code: contract.UsageInvalid, change: func(r *record.Record) { r.Path = "agent run" }},
		{name: "other explicit output", exit: 2, code: contract.UsageInvalid, change: func(r *record.Record) { r.Inputs["output"] = "other-output" }},
		{name: "unknown paid effect", exit: 1, code: contract.ResumeUnsafe, change: func(r *record.Record) {
			r.Status, r.Items[0].Steps[0].Status = record.Unknown, record.StepUnknown
		}},
		{name: "failed paid effect", exit: 1, code: contract.ResumeUnsafe, change: func(r *record.Record) {
			r.Status, r.Items[0].Steps[0].Status = record.Failed, record.StepFailed
		}},
		{name: "wrong step", exit: 1, code: contract.ResumeUnsafe, change: func(r *record.Record) { r.Items[0].Steps[0].ID = "other" }},
		{name: "missing item", exit: 1, code: contract.ResumeUnsafe, change: func(r *record.Record) { r.Items = nil }},
		{name: "changed completed artifact", exit: 1, code: contract.ResumeUnsafe, change: func(r *record.Record) {
			r.Items[0].Steps[0].Artifacts[0].SHA256 = strings.Repeat("0", 64)
		}},
		{name: "pending invalid resolution", exit: 1, code: contract.ResumeUnsafe, change: func(r *record.Record) {
			pending(r)
			r.Resolved["output"] = "relative"
		}},
		{name: "pending missing output", exit: 1, code: contract.ResumeUnsafe, change: func(r *record.Record) {
			pending(r)
			r.Resolved["output"] = filepath.Join(r.Resolved["output"].(string), "missing")
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			config := filepath.Join(root, ".config", "nicely")
			runs := record.Store{Dir: filepath.Join(root, ".local", "state", "nicely", "runs")}
			out, bin := filepath.Join(root, "out"), filepath.Join(root, "bin")
			for _, dir := range []string{config, runs.Dir, filepath.Join(out, "saved"), bin} {
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			write := func(path string, data []byte, mode os.FileMode) {
				t.Helper()
				if err := os.WriteFile(path, data, mode); err != nil {
					t.Fatal(err)
				}
			}
			write(filepath.Join(config, "config.toml"), fmt.Appendf(nil, "[transcript.youtube]\noutput = %q\n", out), 0o600)
			// Version probes pass, but an accidental engine call cannot reach
			// uv, Python, YouTube or Deepgram through the caller's PATH.
			for name, version := range map[string]string{"uv": "uv 0.12.0", "python3": "Python 3.12.0"} {
				write(filepath.Join(bin, name), fmt.Appendf(nil, "#!/bin/sh\n[ \"$*\" = --version ] || exit 97\nprintf '%%s\\n' '%s'\n", version), 0o700)
			}
			var saved record.Record
			if err := json.Unmarshal([]byte(strings.ReplaceAll(string(files["saved.json"]), "ROOT", root)), &saved); err != nil {
				t.Fatal(err)
			}
			artifacts := map[string][]byte{}
			for _, artifact := range saved.Items[0].Steps[0].Artifacts {
				data := files["words.txt"]
				if strings.HasSuffix(artifact.Path, ".json") {
					data = files["paragraphs.json"]
				}
				write(artifact.Path, data, 0o600)
				artifacts[artifact.Path] = data
			}
			if test.change != nil {
				test.change(&saved)
			}
			before, err := json.Marshal(saved)
			if err != nil {
				t.Fatal(err)
			}
			write(runs.File("saved"), before, 0o600)
			if test.busy {
				run, err := runs.Open(context.Background(), "saved")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(run.Close)
			}
			call := func(named bool) (string, string) {
				t.Helper()
				args := []string{"collision-cli", "saved", "run", "youtube", "--url", "https://youtu.be/abc", "--json"}
				if named {
					args = append(args, "--run-id", "saved")
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, binary, args...)
				cmd.Dir = root
				cmd.Env = []string{
					"HOME=" + root, "PATH=" + bin, "NO_COLOR=1", "NCLY_LANG=en", "NCLY_NO_INPUT=1",
					"XDG_CONFIG_HOME=" + filepath.Join(root, ".config"), "XDG_DATA_HOME=" + filepath.Join(root, ".local", "share"),
					"XDG_STATE_HOME=" + filepath.Join(root, ".local", "state"), "XDG_CACHE_HOME=" + filepath.Join(root, ".cache"),
					"DEEPGRAM_API_KEY=generated-collision-fixture", "DBUS_SESSION_BUS_ADDRESS=disabled:",
				}
				var stdout, stderr bytes.Buffer
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				err := cmd.Run()
				if cmd.ProcessState == nil {
					t.Fatalf("named=%t did not start: %v", named, err)
				}
				if code := cmd.ProcessState.ExitCode(); code != test.exit {
					t.Fatalf("named=%t exit %d, want %d: %v\n%s\n%s", named, code, test.exit, err, &stdout, &stderr)
				}
				answer := stdout.Bytes()
				if test.exit != 0 {
					answer = stderr.Bytes()
				}
				var reply struct {
					RunID  string                         `json:"run_id"`
					Errors []struct{ Code contract.Code } `json:"errors"`
				}
				if err := json.Unmarshal(answer, &reply); err != nil || reply.RunID != "saved" {
					t.Fatalf("named=%t has no saved-run answer: %v\n%s", named, err, answer)
				}
				if test.code != "" && (len(reply.Errors) == 0 || reply.Errors[0].Code != test.code) {
					t.Fatalf("named=%t has wrong code, want %s:\n%s", named, test.code, answer)
				}
				if data, err := os.ReadFile(runs.File("saved")); err != nil || !bytes.Equal(data, before) {
					t.Fatalf("named=%t changed the existing record: %v", named, err)
				}
				for path, want := range artifacts {
					if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, want) {
						t.Fatalf("named=%t changed saved output: %v", named, err)
					}
				}
				return stdout.String(), stderr.String()
			}
			namedOut, namedErr := call(true)
			generatedOut, generatedErr := call(false)
			if namedOut != generatedOut || namedErr != generatedErr {
				t.Fatalf("generated collision differs from named control\nnamed:\n%s\n%s\ngenerated:\n%s\n%s", namedOut, namedErr, generatedOut, generatedErr)
			}
		})
	}
}
