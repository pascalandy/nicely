package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pascalandy/nicely/extensions/transcript"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/record"
	"github.com/rogpeppe/go-internal/testscript"
)

// transcriptExternal runs the transcript extension with the test transport
// of extensions/transcript/spec/protocol.md. It runs uv offline from the uv
// cache of the test binary, which NCLY_TEST_UV_CACHE names, with the test
// launcher, which NCLY_TEST_LAUNCHER names. Without both, it refuses to run, so a
// scenario never reaches YouTube or Deepgram. It installs the boundary hook
// that NCLY_TEST_BOUNDARY asks for, as agent-external does.
func transcriptExternal() {
	shortCleanup()
	record.Boundary = boundaryHook(os.Getenv("NCLY_TEST_BOUNDARY"))
	transcript.Test.UVCache = os.Getenv("NCLY_TEST_UV_CACHE")
	transcript.Test.Launcher = os.Getenv("NCLY_TEST_LAUNCHER")
	if transcript.Test.UVCache == "" || transcript.Test.Launcher == "" {
		fmt.Fprintln(os.Stderr, "transcript-external: the test transport needs NCLY_TEST_UV_CACHE and NCLY_TEST_LAUNCHER")
		os.Exit(1)
	}
	transcript.Test.Offline = true
	extension.Main(transcript.Extension)
}

// setupTranscript gives the scenarios of transcript the real uv and Python
// of the host, as UV and PYTHON3. It also gives them the uv cache of the
// test binary, which it warms once with the network, as Go warms its module
// cache.
func setupTranscript(env *testscript.Env) error {
	tools, err := uvTools()
	if err != nil {
		return err
	}
	env.Setenv("UV", tools.uv)
	env.Setenv("PYTHON3", tools.python)
	env.Setenv("NCLY_TEST_UV_CACHE", tools.cache)
	env.Setenv("NCLY_TEST_LAUNCHER", tools.launcher)
	return nil
}

type hostTools struct{ uv, python, cache, launcher string }

// uvTools finds the real uv and Python 3.12 or later, then warms the uv
// cache, once per test process. It runs before setupHome moves HOME.
var uvTools = sync.OnceValues(func() (hostTools, error) {
	var t hostTools
	scratch, err := os.MkdirTemp("", "ncly-uv-warm-")
	if err != nil {
		return t, err
	}
	defer func() { _ = os.RemoveAll(scratch) }()
	if t.uv, err = hostProgram("uv", scratch, "--version"); err != nil {
		return t, err
	}
	check := "import sys; sys.exit(sys.version_info < (3, 12))"
	if t.python, err = hostProgram("python3", scratch, "-c", check); err != nil {
		return t, err
	}
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return t, err
	}
	t.cache = filepath.Join(cacheRoot, "nicely-test", "uv")
	if t.launcher, err = filepath.Abs(filepath.Join(engineDir, "testing", "launcher.py")); err != nil {
		return t, err
	}
	return t, warm(t, scratch)
})

// hostProgram returns the first program of PATH named name that answers
// args with exit 0 under a scratch HOME, as in a scenario. A shim that needs
// the real HOME, such as one of mise, never counts.
func hostProgram(name, home string, args ...string) (string, error) {
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			continue
		}
		cmd := exec.Command(path, args...)
		cmd.Env = []string{"HOME=" + home, "PATH=" + filepath.Dir(path)}
		if cmd.Run() == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("the transcript scenarios need %s in PATH", name)
}

// engineDir is the engine's folder, from the folder of this package.
var engineDir = filepath.Join("..", "..", "extensions", "transcript", "python")

// warm runs the engine's dry run through the line of uv that the scenarios
// run, with the test launcher. It runs offline first, then with the network
// when the cache lacks a package.
func warm(t hostTools, scratch string) error {
	bin := filepath.Join(scratch, "bin")
	if err := os.Mkdir(bin, 0o755); err != nil {
		return err
	}
	for name, target := range map[string]string{"uv": t.uv, "python3": t.python} {
		if err := os.Symlink(target, filepath.Join(bin, name)); err != nil {
			return err
		}
	}
	engine, err := filepath.Abs(engineDir)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(10 * time.Minute).UTC().Format(time.RFC3339)
	request := fmt.Sprintf(`{"protocol_version":1,"type":"request","source":"youtube","urls":["https://youtu.be/warm"],"zoom":null,"output":%q,"temp_dir":%q,"browser":null,"dry_run":true,"deadline":%q}`+"\n", scratch, scratch, deadline)
	var failure []byte
	for _, offline := range []bool{true, false} {
		args := []string{"run", "--script", "--upgrade-package", "yt-dlp", "--no-python-downloads", "--cache-dir", t.cache}
		if offline {
			args = append(args, "--offline")
		}
		args = append(args, "--with-requirements", filepath.Join(engine, "transcript.py"), t.launcher, engine)
		cmd := exec.Command(filepath.Join(bin, "uv"), args...)
		cmd.Env = []string{"HOME=" + scratch, "PATH=" + bin}
		cmd.Stdin = strings.NewReader(request)
		var out []byte
		if out, err = cmd.CombinedOutput(); err == nil && strings.Contains(string(out), `"type": "plan"`) {
			return nil
		}
		failure = out
	}
	return errors.New("cannot warm the uv cache of the transcript scenarios: " + string(failure))
}
