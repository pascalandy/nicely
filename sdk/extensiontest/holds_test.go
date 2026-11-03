package extensiontest_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// heldHost writes part of its output in its stage, writes the marker of HOLD, and waits until the kit removes it.
var heldHost = strings.Replace(publicationHost, `cat "$input" > "$stage"`, `printf part > "$stage"
if [ -n "$HOLD" ]; then
  # PUBLIC
  echo held > "$HOLD"
  while [ -e "$HOLD" ]; do sleep 0.01; done
fi
cat "$input" >> "$stage"`, 1)

// TestHolds checks holds through the fixture of TestPredicts: the public path while the worker waits, then the commit.
func TestHolds(t *testing.T) {
	cases := []struct{ name, host, script, problem string }{
		{"create", heldHost, "env HOLD=held\nholds held out/clip.txt ncly video convert clip.mov\nstdout '\"output\":\"out/clip.txt\"'\ngrep '^partfake media$' out/clip.txt\n", ""},
		{"replace", heldHost, "cp clip.mov out/clip.txt\nenv HOLD=held\nholds held out/clip.txt ncly-video-convert clip.mov --force\ngrep '^partfake media$' out/clip.txt\n", ""},
		{"partial output at the public path", strings.Replace(heldHost, "# PUBLIC", `printf part > "$out"`, 1), "env HOLD=held\nholds held out/clip.txt ncly video convert clip.mov\n", "holds: while the worker waits, out/clip.txt holds work of the run"},
		{"partial output over the old file", strings.Replace(heldHost, "# PUBLIC", `printf part > "$out"`, 1), "cp clip.mov out/clip.txt\nenv HOLD=held\nholds held out/clip.txt ncly video convert clip.mov --force\n", "holds: while the worker waits, out/clip.txt holds work of the run"},
		{"failed run", strings.Replace(heldHost, `cat "$input" >> "$stage"`, `fail RUNTIME "The worker failed." 1`, 1), "env HOLD=held\nholds held out/clip.txt ncly video convert clip.mov\n", "holds: ncly exited with 1 after the kit released the worker, want 0"},
		{"other path", heldHost, "env HOLD=held\nholds held out/other.txt ncly video convert clip.mov\n", `holds: the run published at "out/clip.txt", want out/other.txt`},
		{"protected destination", heldHost, "cp clip.mov out/clip.txt\nprotect out/clip.txt\nenv HOLD=held\nholds held out/clip.txt ncly video convert clip.mov --force\n", "holds: the destination out/clip.txt is a protected path of the call"},
		{"stale marker", heldHost, "cp clip.mov held\nenv HOLD=held\nholds held out/clip.txt ncly video convert clip.mov\n", "holds: marker held must not exist before the call"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := publicationFixture(t, c.host, declared, c.script)
			switch {
			case c.problem == "" && strings.Contains(out, "holds:"):
				t.Fatalf("valid scenario failed: %v\n%s", err, out)
			case c.problem != "" && (err == nil || !strings.Contains(out, c.problem)):
				t.Fatalf("want %q, got %v\n%s", c.problem, err, out)
			}
		})
	}
}

// TestHoldsStopsDetachedWorkers proves that a failed holds stops a worker in its own session, as runner starts one.
func TestHoldsStopsDetachedWorkers(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("only the guard of Linux adopts a worker in its own session")
	}
	pidFile := filepath.Join(t.TempDir(), "worker.pid")
	worker := `setsid sh -c 'echo $$ > "` + pidFile + `"; exec sleep 30' </dev/null >/dev/null 2>&1 &
  until [ -s "` + pidFile + `" ]; do sleep 0.01; done
  printf part > "$out"`
	out, err := publicationFixture(t, strings.Replace(heldHost, "# PUBLIC", worker, 1), declared, "env HOLD=held\nholds held out/clip.txt ncly video convert clip.mov\n")
	if err == nil || !strings.Contains(out, "holds: while the worker waits") {
		t.Fatalf("want the partial output refused, got %v\n%s", err, out)
	}
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	for deadline := time.Now().Add(5 * time.Second); syscall.Kill(pid, 0) == nil; time.Sleep(10 * time.Millisecond) {
		if time.Now().After(deadline) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
			t.Fatalf("the worker %d in its own session still runs after holds failed", pid)
		}
	}
}
