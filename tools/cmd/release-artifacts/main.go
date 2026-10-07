package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type (
	target   struct{ OS, Arch string }
	artifact struct {
		Name  string
		Files map[string]entry
	}
)

type release struct {
	Version  string
	Source   artifact
	Binaries map[target]artifact
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: release-artifacts assets|verify|formula vX.Y.Z <paths>")
	}
	version := args[1]
	if !regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`).MatchString(version) {
		return fmt.Errorf("invalid release tag %q; use vX.Y.Z", version)
	}
	switch {
	case args[0] == "assets" && len(args) == 3:
		return prepare(version, args[2])
	case args[0] == "verify" && len(args) == 4:
		r, err := inspect(version, args[2], args[3])
		if err != nil {
			return err
		}
		if err := verifySource(r, args[2]); err != nil {
			return err
		}
		fmt.Printf("Verified %s: four binary targets, source rebuilds, notices, completions, and checksums\n", version)
		return nil
	case args[0] == "formula" && len(args) == 3:
		return writeFormula(version, args[2])
	default:
		return fmt.Errorf("usage: release-artifacts assets vX.Y.Z <tree> | verify vX.Y.Z <tree> <dist> | formula vX.Y.Z <dist>")
	}
}

func writeFormula(version, dist string) error {
	name := "ncly_" + strings.TrimPrefix(version, "v") + "_source.tar.gz"
	hash, err := fileHash(filepath.Join(dist, name))
	if err != nil {
		return err
	}
	data := fmt.Sprintf(`class Ncly < Formula
  desc "Curated multilingual CLI toolbox for humans and agents"
  homepage "https://github.com/pascalandy/nicely"
  url "https://github.com/pascalandy/nicely/releases/download/%s/%s"
  sha256 "%x"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-trimpath", "-buildvcs=false", "-ldflags=-s -w -X github.com/pascalandy/nicely/internal/cli.version=%s", "-o", bin/"ncly", "./cmd/ncly"
    generate_completions_from_executable(bin/"ncly", shell_parameter_format: :cobra)
    (pkgshare/"licenses").install "LICENSE", "THIRD_PARTY_LICENSES"
  end

  test do
    assert_equal "ncly %s", shell_output("#{bin}/ncly --version").strip
  end
end
`, version, name, hash, version, version)
	path := filepath.Join(dist, "Formula", "ncly.rb")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(data), 0o644)
}
