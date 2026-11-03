package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// This drives the shipped binary, not the test binary's fake keychain. Its
// caller bus is deliberately unusable, so the old wrapper cannot reach a real
// user's session during the failing regression.
func TestPrivateLinuxKeyring(t *testing.T) {
	for _, name := range []string{"dbus-run-session", "gnome-keyring-daemon"} {
		if _, err := exec.LookPath(name); err != nil {
			t.Skipf("%s is required for the private Linux keyring test", name)
		}
	}
	repo, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(repo, ".agents", "skills", "verify-nicely", "scripts", "verify.sh")
	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.MkdirAll(home, 0o700); err != nil {
		t.Fatal(err)
	}
	// Preserve the trusted compiler caches while all runtime state stays in
	// scratch. The wrapper must replace this invalid bus for every CLI step.
	for _, name := range []string{"GOCACHE", "GOMODCACHE"} {
		out, err := exec.Command("go", "env", name).Output()
		if err != nil {
			t.Fatal(err)
		}
		t.Setenv(name, strings.TrimSpace(string(out)))
	}
	for name, value := range map[string]string{
		"HOME": home, "XDG_CACHE_HOME": filepath.Join(home, ".cache"),
		"XDG_CONFIG_HOME": filepath.Join(home, ".config"), "XDG_DATA_HOME": filepath.Join(home, ".local", "share"),
		"XDG_STATE_HOME": filepath.Join(home, ".local", "state"), "XDG_RUNTIME_DIR": filepath.Join(root, "caller-runtime"),
		"DBUS_SESSION_BUS_ADDRESS": "unix:path=" + filepath.Join(root, "no-caller-bus"),
		"DBUS_SESSION_BUS_PID":     "", "GNOME_KEYRING_CONTROL": "", "VERIFYNICELY_API_KEY": "",
	} {
		t.Setenv(name, value)
	}
	call := func(input string, want int, args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, "bash", append([]string{script}, args...)...)
		cmd.Dir = repo
		cmd.Stdin = strings.NewReader(input)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		err := cmd.Run()
		code := 0
		if err != nil {
			code = -1
			var exit *exec.ExitError
			if errors.As(err, &exit) {
				code = exit.ExitCode()
			}
		}
		if code != want {
			t.Fatalf("verify %v exit %d, want %d\n%s\n%s", args, code, want, &stdout, &stderr)
		}
		return stdout.String()
	}
	run := strings.TrimSpace(call("", 0, "start"))
	cleaned := false
	t.Cleanup(func() {
		if !cleaned {
			call("", 0, "cleanup", run)
		}
	})
	call("", 0, "prepare", run, "auth")
	const fake = "private-linux-keyring-fixture"
	call(fake+"\n", 0, "step", run, "auth", "auth", "login", "verifynicely", "--stdin", "--force", "--json")
	status := call("", 78, "step", run, "auth", "auth", "status", "--json")
	if !strings.Contains(status, `"keychain":"available"`) || !strings.Contains(status, `"source":"keychain"`) {
		t.Fatalf("stored key is not read from the private keyring:\n%s", status)
	}
	call("", 0, "step", run, "auth", "auth", "logout", "verifynicely", "--force", "--json")
	status = call("", 78, "step", run, "auth", "auth", "status", "--json")
	if !strings.Contains(status, `"source":"none"`) || !strings.Contains(status, `"code":"AUTH_MISSING"`) {
		t.Fatalf("logout did not remove the key:\n%s", status)
	}
	// Leave a fake key for the real cleanup procedure to remove and prove gone.
	call(fake+"\n", 0, "step", run, "auth", "auth", "login", "verifynicely", "--stdin", "--force", "--json")
	pidData, err := os.ReadFile(filepath.Join(run, "scratch", "keyring", "daemon-pid"))
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
	if err != nil || pid <= 0 {
		t.Fatalf("invalid private daemon PID: %v", err)
	}
	call("", 0, "cleanup", run)
	cleaned = true
	if err := syscall.Kill(pid, 0); !errors.Is(err, syscall.ESRCH) {
		t.Errorf("private keyring daemon still exists after cleanup: %v", err)
	}
	for _, name := range []string{"scratch", "build"} {
		if _, err := os.Stat(filepath.Join(run, name)); !os.IsNotExist(err) {
			t.Errorf("cleanup kept %s: %v", name, err)
		}
	}
	for _, feature := range []string{"auth", "cleanup"} {
		quoted := call("", 0, "quote", run, feature)
		if strings.Contains(quoted, fake) {
			t.Fatal("a transcript printed the fake key")
		}
	}
}

func TestPrivateKeyringMissingProgram(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Fatal(err)
	}
	script, err := filepath.Abs(filepath.Join("..", "..", ".agents", "skills", "verify-nicely", "scripts", "verify.sh"))
	if err != nil {
		t.Fatal(err)
	}
	for _, missing := range []string{"dbus-run-session", "gnome-keyring-daemon"} {
		t.Run(missing, func(t *testing.T) {
			bin := t.TempDir()
			for _, name := range []string{"bash", "dirname", "uname"} {
				path, err := exec.LookPath(name)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(path, filepath.Join(bin, name)); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(bin, "go"), []byte("#!/bin/sh\nexit 99\n"), 0o700); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"dbus-run-session", "gnome-keyring-daemon"} {
				if name != missing {
					if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\nexit 99\n"), 0o700); err != nil {
						t.Fatal(err)
					}
				}
			}
			cmd := exec.Command(bash, script, "start")
			cmd.Env = []string{"PATH=" + bin}
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), missing+" is missing") {
				t.Fatalf("missing %s did not fail with its name: %v\n%s", missing, err, out)
			}
		})
	}
}
