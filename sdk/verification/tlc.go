package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// maxJar bounds the jar that a run reads. TLC v1.7.4 has 2,274,532 bytes.
	maxJar = 64 << 20
	// download bounds the request for the jar.
	download = 2 * time.Minute
	// checkTime bounds one program of a run.
	checkTime = 10 * time.Minute
)

var (
	// The line of the end of a run. A progress line before it gives a count
	// of the states found so far.
	distinctStates = regexp.MustCompile(`(?m)^[\d,]+ states generated, ([\d,]+) distinct states found, ([\d,]+) states left on queue\.$`)
	graphDepth     = regexp.MustCompile(`depth of the complete state graph search is (\d+)`)
)

// acquire reads the local jar, or downloads the pinned URL, and checks its
// SHA-256 before anything runs it. It writes the jar into the folder of the
// run, and returns its path.
func acquire(o Options, suite Suite) (string, error) {
	var jar []byte
	var err error
	from := o.Jar
	if o.Jar != "" {
		jar, err = readJar(o.Jar)
	} else {
		from = suite.URL
		jar, err = fetchJar(suite.URL)
	}
	if err != nil {
		return "", fmt.Errorf("get TLC from %s: %w", from, err)
	}
	sum := sha256.Sum256(jar)
	if got := hex.EncodeToString(sum[:]); got != suite.SHA256 {
		return "", fmt.Errorf("the TLC jar of %s has the SHA-256 %s, and %s pins %s. Java did not start. "+
			"Fix: give the jar with the pinned digest. Changed bytes never prove a model, and a new pin needs a review", from, got, SuiteFile, suite.SHA256)
	}
	path := filepath.Join(o.Dir, "tla2tools.jar")
	if err := os.WriteFile(path, jar, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// readJar reads a local jar of at most maxJar bytes.
func readJar(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return limited(f)
}

// fetchJar downloads a jar of at most maxJar bytes.
func fetchJar(url string) ([]byte, error) {
	client := http.Client{Timeout: download}
	response, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", response.StatusCode)
	}
	return limited(response.Body)
}

func limited(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxJar+1))
	if err == nil && len(data) > maxJar {
		err = fmt.Errorf("the jar has more than %d bytes", maxJar)
	}
	return data, err
}

// startJava runs java -version with the pinned Java of mise, which installs
// it when it is missing.
func startJava(o Options, suite Suite) error {
	code, _, err := command(o, "java", "mise", "exec", suite.Java, "--", "java", "-version")
	if err == nil && code != 0 {
		err = fmt.Errorf("exit %d", code)
	}
	if err != nil {
		return fmt.Errorf("start the pinned Java %s with mise: %w. Inspect %s", suite.Java, err, filepath.Join(o.Dir, "java.log"))
	}
	return nil
}

// runCheck runs TLC on one check, and accepts only its expected outcome.
func runCheck(o Options, suite Suite, jar string, c Check) (Result, error) {
	code, output, err := command(o, "tlc-"+c.Name,
		"mise", "exec", suite.Java, "--", "java", "-XX:+UseParallelGC", "-Djava.io.tmpdir="+filepath.Join(o.Dir, "java"),
		"-jar", jar, "-workers", "1", "-metadir", filepath.Join(o.Dir, "states", c.Name),
		"-config", filepath.Join(o.Root, filepath.FromSlash(c.Config)), filepath.Join(o.Root, filepath.FromSlash(c.Model)))
	if err != nil {
		return Result{}, err
	}
	result, err := parseResult(c, code, output)
	if err != nil {
		return Result{}, fmt.Errorf("%w. Inspect %s", err, filepath.Join(o.Dir, "tlc-"+c.Name+".log"))
	}
	return result, nil
}

// command runs one program in the root, with the environment of the run. It
// keeps the arguments, the output, and the exit of the program in
// <label>.log of the folder of the run. It returns the exit code, or an error
// when the program did not run to its end.
func command(o Options, label string, args ...string) (int, string, error) {
	if err := os.MkdirAll(filepath.Join(o.Dir, "java"), 0o700); err != nil {
		return 0, "", err
	}
	program, err := lookPath(o.Env, args[0])
	if err != nil {
		return 0, "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), checkTime)
	defer cancel()
	cmd := exec.CommandContext(ctx, program, args[1:]...)
	cmd.Dir = o.Root
	cmd.Env = o.Env
	cmd.WaitDelay = 5 * time.Second
	output, err := cmd.CombinedOutput()
	code := 0
	var exit *exec.ExitError
	if errors.As(err, &exit) && ctx.Err() == nil && exit.ExitCode() >= 0 {
		code, err = exit.ExitCode(), nil
	} else if err != nil && ctx.Err() != nil {
		err = fmt.Errorf("%s ran longer than %s", label, checkTime)
	}
	header, _ := json.Marshal(args)
	log := fmt.Sprintf("%s\n%s\nexit=%d\n", header, output, code)
	path := filepath.Join(o.Dir, label+".log")
	if writeErr := os.WriteFile(path, []byte(log), 0o600); writeErr != nil {
		return 0, "", errors.Join(err, writeErr)
	}
	if err != nil {
		return 0, "", fmt.Errorf("%s: %w. Inspect %s", label, err, path)
	}
	return code, string(output), nil
}

// lookPath finds name in the PATH of env, as the program of the run would,
// since exec.Command searches the PATH of the caller.
func lookPath(env []string, name string) (string, error) {
	var path string
	for _, entry := range env {
		if value, ok := strings.CutPrefix(entry, "PATH="); ok {
			path = value
		}
	}
	for _, dir := range filepath.SplitList(path) {
		file := filepath.Join(dir, name)
		if info, err := os.Stat(file); filepath.IsAbs(dir) && err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0 {
			return file, nil
		}
	}
	return "", fmt.Errorf("%s is in no folder of the PATH of the run. Fix: install mise, which installs the pinned Java", name)
}

// parseResult accepts the output of TLC only with the exit, the message,
// and the statistics of the outcome that c expects.
func parseResult(c Check, code int, output string) (Result, error) {
	want, message := expected(c), "Model checking completed. No error has been found."
	if c.Violation != "" {
		message = "Invariant " + c.Violation + " is violated."
	}
	count, depth := distinctStates.FindStringSubmatch(output), graphDepth.FindStringSubmatch(output)
	if code != want || !strings.Contains(output, message) || count == nil || depth == nil {
		return Result{}, fmt.Errorf("TLC %s: exit %d, want %d with %q and the statistics of the graph", c.Name, code, want, message)
	}
	result := Result{Name: c.Name, Exit: code, Violation: c.Violation}
	if c.Violation != "" {
		return result, nil
	}
	if count[2] != "0" {
		return Result{}, fmt.Errorf("TLC %s: states are left on the queue, so the exploration is not complete", c.Name)
	}
	var err error
	if result.Distinct, err = strconv.Atoi(strings.ReplaceAll(count[1], ",", "")); err == nil {
		result.Depth, err = strconv.Atoi(depth[1])
	}
	if err != nil || result.Distinct < 1 || result.Depth < 1 {
		return Result{}, fmt.Errorf("TLC %s: invalid statistics of the graph", c.Name)
	}
	return result, nil
}
