package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestReleaseArtifacts(t *testing.T) {
	tool := filepath.Join(t.TempDir(), "release-artifacts")
	testCommand(t, ".", nil, "go", "build", "-o", tool, ".")
	t.Run("invalid tag", func(t *testing.T) {
		out, err := exec.CommandContext(t.Context(), tool, "verify", "v01.2.3", "missing", "missing").CombinedOutput()
		if err == nil || !strings.Contains(string(out), "invalid release tag") {
			t.Fatalf("invalid tag must fail before reading artifacts: %s (%v)", out, err)
		}
	})
	source := map[string][]byte{
		"go.mod":                   []byte("module github.com/pascalandy/nicely\n\ngo 1.27.0\n\nrequire example.com/dependency v0.0.0\nreplace example.com/dependency => ./dependency\n"),
		"go.sum":                   {},
		"LICENSE":                  []byte("MIT license fixture\n"),
		"cmd/ncly/main.go":         []byte("package main\nimport (\"fmt\"; \"github.com/pascalandy/nicely/internal/cli\"; _ \"example.com/dependency\")\nfunc main() { fmt.Println(\"ncly \"+cli.Version()) }\n"),
		"internal/cli/version.go":  []byte("package cli\nvar version = \"dev\"\nfunc Version() string { return version }\n"),
		"dependency/go.mod":        []byte("module example.com/dependency\n\ngo 1.27.0\n"),
		"dependency/dependency.go": []byte("package dependency\n"),
		"dependency/LICENSE":       []byte("dependency license\n"),
		"dependency/NOTICE":        []byte("dependency attribution\n"),
	}
	buildTree := t.TempDir()
	for name, data := range source {
		testWrite(t, filepath.Join(buildTree, name), data)
	}
	binaries := make(map[string][]byte)
	for _, target := range []string{"darwin_amd64", "darwin_arm64", "linux_amd64", "linux_arm64"} {
		parts := strings.Split(target, "_")
		bin := filepath.Join(t.TempDir(), "ncly")
		testCommand(t, buildTree, []string{"GOOS=" + parts[0], "GOARCH=" + parts[1], "CGO_ENABLED=0"}, "go", "build", "-trimpath", "-buildvcs=false", "-ldflags=-s -w -X github.com/pascalandy/nicely/internal/cli.version=v0.0.1", "-o", bin, "./cmd/ncly")
		binaries[target] = testRead(t, bin)
	}
	cases := []struct {
		name, reason string
		mutate       func(*testing.T, string, map[string]map[string][]byte)
	}{
		{name: "valid"},
		{name: "missing target", reason: "binary targets", mutate: func(t *testing.T, tree string, archives map[string]map[string][]byte) {
			delete(archives, "ncly_0.0.1_linux_arm64.tar.gz")
		}},
		{name: "duplicate target", reason: "duplicate target", mutate: func(t *testing.T, tree string, archives map[string]map[string][]byte) {
			archives["duplicate.tar.gz"] = archives["ncly_0.0.1_linux_arm64.tar.gz"]
		}},
		{name: "missing NOTICE", reason: "NOTICE", mutate: func(t *testing.T, tree string, archives map[string]map[string][]byte) {
			delete(archives["ncly_0.0.1_linux_arm64.tar.gz"], "THIRD_PARTY_LICENSES/example.com/dependency/NOTICE")
		}},
		{name: "collector omits NOTICE", reason: "runtime dependency license or NOTICE missing", mutate: func(t *testing.T, tree string, archives map[string]map[string][]byte) {
			if err := os.Remove(filepath.Join(tree, ".release/notices/example.com/dependency/NOTICE")); err != nil {
				t.Fatal(err)
			}
			for _, files := range archives {
				delete(files, "THIRD_PARTY_LICENSES/example.com/dependency/NOTICE")
				delete(files, "ncly-0.0.1/THIRD_PARTY_LICENSES/example.com/dependency/NOTICE")
			}
		}},
		{name: "unsafe archive", reason: "unsafe archive path", mutate: func(t *testing.T, tree string, archives map[string]map[string][]byte) {
			archives["ncly_0.0.1_linux_arm64.tar.gz"]["../escape"] = []byte("escape")
		}},
		{name: "stale source", reason: "source differs", mutate: func(t *testing.T, tree string, archives map[string]map[string][]byte) {
			testWrite(t, filepath.Join(tree, "internal/cli/version.go"), []byte("package cli\nvar version = \"changed\"\nfunc Version() string { return version }\n"))
		}},
		{name: "extra source file", reason: "unexpected source archive member", mutate: func(t *testing.T, tree string, archives map[string]map[string][]byte) {
			archives["ncly_0.0.1_source.tar.gz"]["ncly-0.0.1/untracked.txt"] = []byte("untracked release content\n")
		}},
		{name: "packaged version differs", reason: "packaged binary reports wrong release version", mutate: func(t *testing.T, tree string, archives map[string]map[string][]byte) {
			bad := t.TempDir()
			testWrite(t, filepath.Join(bad, "go.mod"), []byte("module example.com/wrong\n\ngo 1.27.0\n"))
			testWrite(t, filepath.Join(bad, "main.go"), []byte("package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"ncly wrong\")}\n"))
			bin := filepath.Join(t.TempDir(), "ncly")
			testCommand(t, bad, []string{"CGO_ENABLED=0"}, "go", "build", "-trimpath", "-buildvcs=false", "-ldflags=-s -w -X github.com/pascalandy/nicely/internal/cli.version=v0.0.1", "-o", bin, ".")
			archives["ncly_0.0.1_"+runtime.GOOS+"_"+runtime.GOARCH+".tar.gz"]["ncly"] = testRead(t, bin)
		}},
		{name: "foreign binary differs", reason: "binary differs from source rebuild", mutate: func(t *testing.T, tree string, archives map[string]map[string][]byte) {
			osName, arch := "darwin", "arm64"
			if runtime.GOOS == osName && runtime.GOARCH == arch {
				osName, arch = "linux", "amd64"
			}
			bin := filepath.Join(t.TempDir(), "ncly")
			testCommand(t, tree, []string{"GOOS=" + osName, "GOARCH=" + arch, "CGO_ENABLED=0"}, "go", "build", "-trimpath", "-buildvcs=false", "-ldflags=-s -w -X github.com/pascalandy/nicely/internal/cli.version=v0.0.2", "-o", bin, "./cmd/ncly")
			archives["ncly_0.0.1_"+osName+"_"+arch+".tar.gz"]["ncly"] = testRead(t, bin)
		}},
		{name: "wrong hash", reason: "checksum mismatch"},
		{name: "corrupt gzip trailer", reason: "gzip: invalid checksum"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree, dist, archives := testFixture(t, source, binaries)
			if tc.mutate != nil {
				tc.mutate(t, tree, archives)
			}
			testArchives(t, dist, archives)
			if tc.name == "wrong hash" {
				testWrite(t, filepath.Join(dist, "ncly_0.0.1_checksums.txt"), []byte(strings.Repeat("0", 64)+"  ncly_0.0.1_source.tar.gz\n"))
			}
			if tc.name == "corrupt gzip trailer" {
				path := filepath.Join(dist, "ncly_0.0.1_source.tar.gz")
				data := testRead(t, path)
				original := sha256.Sum256(data)
				data[len(data)-8] ^= 0xff
				testWrite(t, path, data)
				corrupted := sha256.Sum256(data)
				manifest := filepath.Join(dist, "ncly_0.0.1_checksums.txt")
				testWrite(t, manifest, []byte(strings.ReplaceAll(string(testRead(t, manifest)), fmt.Sprintf("%x", original), fmt.Sprintf("%x", corrupted))))
			}
			cmd := exec.CommandContext(t.Context(), tool, "verify", "v0.0.1", tree, dist)
			if tc.name == "valid" {
				gitPath, err := exec.LookPath("git")
				if err != nil {
					t.Fatal(err)
				}
				guards := t.TempDir()
				testWrite(t, filepath.Join(guards, "git"), []byte("#!/bin/sh\nif [ \"$1\" = ls-files ]; then exec \"$NCLY_TEST_GIT\" \"$@\"; fi\nprintf 'git command blocked: %s\\n' \"$*\" >&2\nexit 1\n"))
				if err := os.Chmod(filepath.Join(guards, "git"), 0o755); err != nil {
					t.Fatal(err)
				}
				cmd.Env = append(os.Environ(), "PATH="+guards+string(os.PathListSeparator)+os.Getenv("PATH"), "NCLY_TEST_GIT="+gitPath)
			}
			out, err := cmd.CombinedOutput()
			if tc.reason != "" {
				if err == nil || !strings.Contains(string(out), tc.reason) {
					t.Fatalf("want rejection %q; got %s (%v)", tc.reason, out, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("valid artifacts rejected: %s (%v)", out, err)
			}
			testCommand(t, tree, nil, tool, "formula", "v0.0.1", dist)
			formula := string(testRead(t, filepath.Join(dist, "Formula/ncly.rb")))
			hash := sha256.Sum256(testRead(t, filepath.Join(dist, "ncly_0.0.1_source.tar.gz")))
			for _, want := range []string{fmt.Sprintf("sha256 \"%x\"", hash), "ncly_0.0.1_source.tar.gz", "internal/cli.version=v0.0.1", "generate_completions_from_executable", "THIRD_PARTY_LICENSES"} {
				if !strings.Contains(formula, want) {
					t.Fatalf("formula missing %q: %s", want, formula)
				}
			}
			testCommand(t, dist, nil, "ruby", "-c", "Formula/ncly.rb")
		})
	}
}

func testFixture(t *testing.T, source, binaries map[string][]byte) (string, string, map[string]map[string][]byte) {
	t.Helper()
	tree, dist := t.TempDir(), t.TempDir()
	for name, data := range source {
		testWrite(t, filepath.Join(tree, name), data)
	}
	testCommand(t, tree, nil, "git", "init", "-q")
	testCommand(t, tree, nil, "git", "add", ".")
	completions := map[string][]byte{"ncly.bash": []byte("bash completion\n"), "_ncly": []byte("zsh completion\n"), "ncly.fish": []byte("fish completion\n")}
	notices := map[string][]byte{"example.com/dependency/LICENSE": []byte("dependency license\n"), "example.com/dependency/NOTICE": []byte("dependency attribution\n")}
	for name, data := range completions {
		testWrite(t, filepath.Join(tree, ".release/completions", name), data)
	}
	for name, data := range notices {
		testWrite(t, filepath.Join(tree, ".release/notices", name), data)
	}
	archives := make(map[string]map[string][]byte)
	for target, data := range binaries {
		files := map[string][]byte{"ncly": data, "LICENSE": source["LICENSE"]}
		for name, data := range completions {
			files["completions/"+name] = data
		}
		for name, data := range notices {
			files["THIRD_PARTY_LICENSES/"+name] = data
		}
		archives["ncly_0.0.1_"+target+".tar.gz"] = files
	}
	sourceFiles := make(map[string][]byte)
	for name, data := range source {
		sourceFiles["ncly-0.0.1/"+name] = data
	}
	for name, data := range notices {
		sourceFiles["ncly-0.0.1/THIRD_PARTY_LICENSES/"+name] = data
	}
	archives["ncly_0.0.1_source.tar.gz"] = sourceFiles
	return tree, dist, archives
}

func testArchives(t *testing.T, dist string, archives map[string]map[string][]byte) {
	t.Helper()
	names := make([]string, 0, len(archives))
	for name := range archives {
		names = append(names, name)
	}
	slices.Sort(names)
	var sums strings.Builder
	for _, name := range names {
		f, err := os.Create(filepath.Join(dist, name))
		if err != nil {
			t.Fatal(err)
		}
		gz := gzip.NewWriter(f)
		tw := tar.NewWriter(gz)
		if strings.HasSuffix(name, "_source.tar.gz") {
			if err := tw.WriteHeader(&tar.Header{Name: "pax_global_header", Typeflag: tar.TypeXGlobalHeader, PAXRecords: map[string]string{"comment": "git source archive fixture"}}); err != nil {
				t.Fatal(err)
			}
		}
		files := make([]string, 0, len(archives[name]))
		for p := range archives[name] {
			files = append(files, p)
		}
		slices.Sort(files)
		for _, p := range files {
			data := archives[name][p]
			mode := int64(0o644)
			if p == "ncly" {
				mode = 0o755
			}
			if err := tw.WriteHeader(&tar.Header{Name: p, Mode: mode, Size: int64(len(data)), Typeflag: tar.TypeReg}); err != nil {
				t.Fatal(err)
			}
			if _, err := tw.Write(data); err != nil {
				t.Fatal(err)
			}
		}
		for _, close := range []func() error{tw.Close, gz.Close, f.Close} {
			if err := close(); err != nil {
				t.Fatal(err)
			}
		}
		hash := sha256.Sum256(testRead(t, filepath.Join(dist, name)))
		fmt.Fprintf(&sums, "%x  %s\n", hash, name)
	}
	testWrite(t, filepath.Join(dist, "ncly_0.0.1_checksums.txt"), []byte(sums.String()))
}

func testCommand(t *testing.T, dir string, env []string, name string, args ...string) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %s (%v)", name, args, out, err)
	}
}

func testWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func testRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
