package main

import (
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

// A private session stops cleanly on request and ends when its owning state
// or runtime marker disappears.
func TestPrivateKeyringLifecycle(t *testing.T) {
	for _, name := range []string{"dbus-run-session", "gnome-keyring-daemon"} {
		if _, err := exec.LookPath(name); err != nil {
			t.Skipf("%s is required for the private Linux keyring test", name)
		}
	}
	helper, err := filepath.Abs(filepath.Join("..", "..", ".agents", "skills", "verify-nicely", "scripts", "keyring.sh"))
	if err != nil {
		t.Fatal(err)
	}
	for _, ending := range []string{"normal stop", "state", "runtime marker"} {
		t.Run(ending, func(t *testing.T) {
			root := t.TempDir()
			// Secret Service control sockets must fit Unix's short path limit.
			runtimeRoot, err := os.MkdirTemp(os.TempDir(), "ka-")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.RemoveAll(runtimeRoot); err != nil {
					t.Error(err)
				}
			})
			state, home := filepath.Join(root, "state"), filepath.Join(root, "home")
			if err := os.Mkdir(home, 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("TMPDIR", runtimeRoot)
			call := func(verb string, args ...string) error {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, "bash", append([]string{helper, verb}, args...)...)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Logf("private keyring %s: %v\n%s", verb, err, out)
				}
				return err
			}
			if err := call("start", state, home); err != nil {
				_ = call("stop", state)
				t.Fatal(err)
			}
			read := func(name string) string {
				t.Helper()
				data, err := os.ReadFile(filepath.Join(state, name))
				if err != nil {
					t.Fatal(err)
				}
				return strings.TrimSpace(string(data))
			}
			pid, err := strconv.Atoi(read("daemon-pid"))
			if err != nil || pid <= 1 {
				t.Fatalf("invalid private daemon PID: %v", err)
			}
			runtime := read("runtime")
			if !strings.HasPrefix(runtime, runtimeRoot+string(os.PathSeparator)) {
				t.Fatalf("runtime %q escaped the private test folder", runtime)
			}
			marker := filepath.Join(runtime, ".verify-nicely-keyring")
			t.Cleanup(func() {
				// Restore only our own moved state/marker so the old helper can
				// stop after the failing probe. Never leave the regression daemon.
				if _, err := os.Stat(state + ".moved"); err == nil {
					if err := os.Rename(state+".moved", state); err != nil {
						t.Error(err)
					}
				}
				if errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
					if _, err := os.Stat(runtime); os.IsNotExist(err) {
						return
					}
				}
				if err := os.WriteFile(marker, nil, 0o600); err != nil {
					t.Error(err)
				}
				if err := call("stop", state); err != nil {
					t.Error(err)
				}
			})
			if ending == "normal stop" {
				if err := call("stop", state); err != nil {
					t.Fatal(err)
				}
				if code := read("stopped"); code != "0" {
					t.Errorf("normal private session exited %s, want 0", code)
				}
				if log := read("session.log"); strings.Contains(log, "unbound variable") {
					t.Errorf("normal private stop has a shell failure: %s", log)
				}
				if !errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
					t.Error("private daemon survived normal stop")
				}
				if _, err := os.Stat(runtime); !os.IsNotExist(err) {
					t.Errorf("private runtime remains after normal stop: %v", err)
				}
				if err := call("stop", state); err != nil {
					t.Fatal(err)
				}
				if code := read("stopped"); code != "0" {
					t.Errorf("repeated stop changed the clean session status: %s", code)
				}
				return
			}
			if ending == "state" {
				err = os.Rename(state, state+".moved")
			} else {
				err = os.Remove(marker)
			}
			if err != nil {
				t.Fatal(err)
			}
			deadline := time.Now().Add(5 * time.Second)
			for {
				_, ended := os.Stat(filepath.Join(state, "stopped"))
				finished := ended == nil
				if ending == "state" {
					_, err := os.Stat(runtime)
					finished = os.IsNotExist(err)
				}
				if errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) && finished {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("private daemon survived the loss of its owning " + ending)
				}
				time.Sleep(20 * time.Millisecond)
			}
		})
	}
}
