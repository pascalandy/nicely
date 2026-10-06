// Package cli builds the ncly command tree and runs it.
package cli

import (
	"io"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// version is set at link time by release builds:
// -ldflags "-X github.com/pascalandy/nicely/internal/cli.version=vX.Y.Z".
var version string

// Main runs ncly with args, without the program name, and returns its exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	root := newRoot()
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	if err := root.Execute(); err != nil {
		return 1
	}
	return 0
}

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "ncly",
		Version:       releaseVersion(),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	// Declared here so that Cobra does not take -v, which belongs to --verbose.
	root.Flags().Bool("version", false, "")
	root.SetVersionTemplate("ncly {{.Version}}\n")
	return root
}

// releaseVersion prefers the linked version, then the version that Go stamps
// from the module or the VCS, and finally a development marker.
func releaseVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && strings.HasPrefix(info.Main.Version, "v") {
		return info.Main.Version
	}
	return "v0.0.0-dev"
}
