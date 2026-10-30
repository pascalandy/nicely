package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/pascalandy/nicely/extensions/agent"
	"github.com/pascalandy/nicely/extensions/transcript"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/record"
	"github.com/pascalandy/nicely/sdk/runner"
	hello "github.com/pascalandy/nicely/testdata/extensions/hello"
	"github.com/pascalandy/nicely/testdata/extensions/keyed"
	convert "github.com/pascalandy/nicely/testdata/extensions/video-convert"
	subtitle "github.com/pascalandy/nicely/testdata/extensions/video-subtitle"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	// Every command of the test binary refuses a plural message sent through Printer.T.
	i18n.PluralGuard = true
	testscript.Main(m, map[string]func(){
		"ncly":       testNcly,
		"closedpipe": closedPipe,
		"exitcode":   exitCode,
		"pgroup":     pgroup,
		"pty":        ptyRun,
		"ptystart":   ptyStart,
		"ttyin":      ttyinRun,
		// codes lists the registry of sdk/contract, one code per line,
		// because no command of ncly lists its codes.
		"codes":                   printCodes,
		"holdlock":                holdLock,
		"hello-external":          func() { shortCleanup(); extension.Main(hello.Extension) },
		"keyed-external":          func() { shortCleanup(); extension.Main(keyed.Extension) },
		"video-convert-external":  func() { extension.Main(convert.Extension) },
		"video-subtitle-external": func() { extension.Main(subtitle.Extension) },
		"agent-external": func() {
			shortCleanup()
			record.Boundary = boundaryHook(os.Getenv("NCLY_TEST_BOUNDARY"))
			extension.Main(agent.Extension)
		},
		"transcript-external": transcriptExternal,
		// ncly-bundled is ncly with the hello fixture bundled. It starts the
		// fixture as itself under the name ncly-hello, which reaches the
		// fixture through the routing of run, as in a release.
		"ncly-bundled": testBundled,
		"ncly-hello":   testBundled,
		// ncly-skill is ncly started under the name of its bundled skill
		// extension, as core starts it.
		"ncly-skill": testNcly,
	})
}

func testBundled() {
	shortCleanup()
	useTestKeychain()
	os.Exit(run([]extension.Extension{hello.Extension}))
}

// fixtures holds the Go extensions that scenarios install as external ones,
// by domain. They are the fixtures of testdata/extensions/, and agent and
// transcript of extensions/, which just install installs the same way. The
// setup of their own scenarios installs each one in the data folder.
var fixtures = map[string]extension.Extension{
	"hello": hello.Extension, "keyed": keyed.Extension, "agent": agent.Extension, "transcript": transcript.Extension,
	"video-convert": convert.Extension, "video-subtitle": subtitle.Extension,
}

// boundaryHook returns the boundary hook of record support that
// NCLY_TEST_BOUNDARY asks for, which only the test binary reads: none when
// it is empty. A boundary name, such as after-intent, kills the process
// with SIGKILL at that boundary, as a crash would. <boundary>@<n> kills at
// its n-th time, such as after-intent@2. pause:<boundary> writes $WORK/paused
// at that boundary, then waits 2 seconds, so a scenario can send a signal
// there.
func boundaryHook(name string) func(string) {
	if name == "" {
		return nil
	}
	pause, paused := strings.CutPrefix(name, "pause:")
	kill, count := name, 1
	if before, n, found := strings.Cut(name, "@"); found {
		kill = before
		count, _ = strconv.Atoi(n)
	}
	seen := 0
	return func(at string) {
		switch {
		case paused && at == pause:
			_ = os.WriteFile(filepath.Join(os.Getenv("WORK"), "paused"), nil, 0o644)
			time.Sleep(2 * time.Second)
		case at == kill:
			if seen++; seen == count {
				_ = syscall.Kill(os.Getpid(), syscall.SIGKILL)
				select {}
			}
		}
	}
}

// testNcly runs ncly with a short cleanup and the test keychain.
func testNcly() {
	shortCleanup()
	useTestKeychain()
	main()
}

// shortCleanup gives cleanup 4 seconds instead of 15, so that a scenario
// reaches the SIGKILL of cleanup quickly. The drain limit keeps its 2
// seconds, so the scenarios still tell the two clocks apart.
func shortCleanup() {
	runner.Cleanup = 4 * time.Second
}

// TestScripts runs core's scenarios, then those of each extension and of
// each fixture extension, which keep theirs in their own folder.
func TestScripts(t *testing.T) {
	core, others := scriptDirs(filepath.Join("..", ".."))
	t.Run("core", func(t *testing.T) { runScripts(t, core, setupHome) })
	for _, dir := range others {
		domain := filepath.Base(filepath.Dir(filepath.Dir(dir)))
		t.Run(domain, func(t *testing.T) {
			runScripts(t, dir, func(env *testscript.Env) error {
				if domain == "transcript" {
					if err := setupTranscript(env); err != nil {
						return err
					}
				}
				if err := setupHome(env); err != nil {
					return err
				}
				if e, isFixture := fixtures[domain]; isFixture {
					return install(env.WorkDir, env.Setenv, domain, e)
				}
				return nil
			})
		})
	}
}

func runScripts(t *testing.T, dir string, setup func(*testscript.Env) error) {
	setup = setupCoverage(t, dir, setup)
	testscript.Run(t, testscript.Params{
		Dir:                 dir,
		RequireExplicitExec: true,
		RequireUniqueNames:  true,
		Setup:               setup,
		Condition:           condition,
		Cmds: map[string]func(*testscript.TestScript, bool, []string){
			"exits":     cmdExits,
			"answer":    cmdAnswer,
			"snapshot":  cmdSnapshot,
			"unchanged": cmdUnchanged,
			"pseudo":    cmdPseudo,
			"plain":     cmdPlain,
			"waitfile":  cmdWaitfile,
			"gone":      cmdGone,
			"term":      cmdTerm,
			"install":   cmdInstall,
			"onlyncly":  cmdOnlyncly,
			"agree":     cmdAgree,
			"schema":    cmdSchema,
			"mentions":  cmdMentions,
		},
	})
}

// cmdInstall installs a Go fixture in the data folder, as the setup of the
// fixture's own scenarios does: install <domain>
func cmdInstall(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) != 1 {
		ts.Fatalf("usage: install <domain>")
	}
	e, isFixture := fixtures[args[0]]
	if !isFixture {
		ts.Fatalf("install: no fixture %q", args[0])
	}
	ts.Check(install(ts.Getenv("WORK"), ts.Setenv, args[0], e))
}

// cmdOnlyncly keeps in PATH only ncly, the listed commands of the test
// binary, and the fixtures of $WORK/bin when the scenario has that folder, so
// the other commands of the test binary, such as ncly-hello, are no
// extensions: onlyncly [command...], such as onlyncly exitcode. It links each
// command into $WORK/nbin, which a scenario may name in a PATH of its own,
// after $WORK/bin. A program of the system is then reachable only by its
// absolute path, such as /usr/bin/sed, as the fixtures name it.
func cmdOnlyncly(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("usage: onlyncly [command...]")
	}
	var commands string
	for _, dir := range filepath.SplitList(ts.Getenv("PATH")) {
		if info, err := os.Stat(filepath.Join(dir, "ncly")); err == nil && !info.IsDir() {
			commands = dir
			break
		}
	}
	if commands == "" {
		ts.Fatalf("onlyncly: no ncly in PATH")
	}
	work := ts.Getenv("WORK")
	nbin := filepath.Join(work, "nbin")
	ts.Check(os.Mkdir(nbin, 0o755))
	for _, name := range append([]string{"ncly"}, args...) {
		program := filepath.Join(commands, name)
		if _, err := os.Stat(program); err != nil || strings.HasPrefix(name, "ncly-") {
			ts.Fatalf("onlyncly: %s is no command of the test binary that onlyncly keeps", name)
		}
		ts.Check(os.Symlink(program, filepath.Join(nbin, name)))
	}
	path := nbin
	if info, err := os.Stat(filepath.Join(work, "bin")); err == nil && info.IsDir() {
		path = filepath.Join(work, "bin") + string(filepath.ListSeparator) + nbin
	}
	ts.Setenv("PATH", path)
}

// install puts a Go extension in the data folder as an external one:
// the script ncly-<domain> that runs it, with its manifest, catalogs, and
// skill beside it. The script appends its domain to $WORK/external.log, then runs
// the test binary under the name <domain>-external, so the fixture runs as
// an external extension. DATA names the folder.
func install(work string, setenv func(name, value string), domain string, e extension.Extension) error {
	dir := filepath.Join(work, ".local", "share", "nicely", "extensions")
	executable := filepath.Join(dir, "ncly-"+domain)
	err := os.MkdirAll(dir, 0o755)
	if err == nil {
		err = extension.Export(e, dir)
	}
	if err == nil {
		script := "#!/bin/sh\necho " + domain + " >> \"$WORK/external.log\"\nexec " + domain + "-external \"$@\"\n"
		err = os.WriteFile(executable, []byte(script), 0o755)
	}
	setenv("DATA", dir)
	return err
}

// setupHome points HOME and the XDG folders at $WORK, so ncly never touches
// the host's own config, data, state, or cache.
func setupHome(env *testscript.Env) error {
	home := env.WorkDir
	env.Setenv("HOME", home)
	env.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	env.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
	env.Setenv("XDG_STATE_HOME", filepath.Join(home, ".local", "state"))
	env.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	return nil
}

// condition answers [bash-completion], which holds when the bash-completion
// package is installed, as the generated bash script needs it.
func condition(cond string) (bool, error) {
	if cond != "bash-completion" {
		return false, fmt.Errorf("unknown condition %q", cond)
	}
	for _, path := range []string{
		"/usr/share/bash-completion/bash_completion",
		"/opt/homebrew/share/bash-completion/bash_completion",
	} {
		if _, err := os.Stat(path); err == nil {
			return true, nil
		}
	}
	return false, nil
}
