package extensiontest_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/pascalandy/nicely/sdk/runner"
)

// TestSignalRunnerCleanupBudget uses the real runner without shortening its cleanup.
func TestSignalRunnerCleanupBudget(t *testing.T) {
	out, err := goTest(t, "standalone-named", os.Getenv("PATH"),
		edit{"extension.go", `"fmt"`, "\"fmt\"\n\"github.com/pascalandy/nicely/sdk/runner\""},
		edit{"extension.go", "func convert(_ context.Context", "func convert(ctx context.Context"},
		edit{"extension.go", `key, problem := call.Key("example")`, `
		_, _ = runner.Run(ctx, runner.Program{Path: "/bin/sh", Args: []string{"-c", "trap '' INT TERM; echo ready > started; while :; do /bin/sleep 1; done"}})
		key, problem := call.Key("example")`},
		edit{file: "testdata/script/convert.txtar", new: "grant example\nsignal TERM started ncly-video-convert clip.mov\nstderr '\"code\":\"TERMINATED\"'\n"})
	if err != nil {
		t.Fatalf("full SDK cleanup failed: %v\n%s", err, out)
	}
}

// TestSignalTimeoutBoundsStreams requires timeout to stop the host and its stream-holding child.
func TestSignalTimeoutBoundsStreams(t *testing.T) {
	checkSignalStreams(t, "", "host did not exit after terminated", runner.Cleanup+runner.DrainLimit+8*time.Second)
}

// TestSignalDrainBoundsStreams bounds reading when the host exits but its child keeps the streams open.
func TestSignalDrainBoundsStreams(t *testing.T) {
	checkSignalStreams(t, "trap 'exit 0' TERM\n", "WaitDelay expired before I/O complete", runner.DrainLimit+8*time.Second)
}

func checkSignalStreams(t *testing.T, afterStart, problem string, limit time.Duration) {
	t.Helper()
	pidfile := filepath.Join(t.TempDir(), "child.pid")
	// The finite child bounds a broken kit's red run. It must still be alive at the kit's timeout.
	host := "#!/bin/sh\ntrap '' INT TERM\n/bin/sleep 35 &\necho $! > '" + pidfile + "'\n" + afterStart + "echo ready > started\nwait\n"
	t.Cleanup(func() {
		if data, err := os.ReadFile(pidfile); err == nil {
			if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && pid > 0 {
				_ = syscall.Kill(pid, syscall.SIGKILL)
			}
		}
	})
	start := time.Now()
	out, err := pairedFixture(t, host, "signal TERM started ncly-video-convert clip.mov\n")
	took := time.Since(start)
	t.Logf("timeout returned after %v: %v\n%s", took, err, out)
	if err == nil || !strings.Contains(out, problem) {
		t.Errorf("want %q, got %v\n%s", problem, err, out)
	}
	if took > limit {
		t.Errorf("stream-holding child delayed timeout to %v, want less than %v", took, limit)
	}
	data, err := os.ReadFile(pidfile)
	if err != nil {
		t.Fatal(err)
	}
	pid := strings.TrimSpace(string(data))
	state, err := exec.Command("/bin/ps", "-p", pid, "-o", "stat=").Output()
	if err == nil && !strings.HasPrefix(strings.TrimSpace(string(state)), "Z") {
		t.Errorf("stream-holding child %s still runs: %s", pid, state)
	}
}
