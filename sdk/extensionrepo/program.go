package extensionrepo

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
)

// Main runs install or uninstall from the repository, then exits.
// The export subcommand writes embedded sidecars for the staged build.
func Main(e extension.Extension) {
	os.Exit(run(e, os.Args[1:]))
}

func run(e extension.Extension, args []string) int {
	if len(args) == 2 && args[0] == "export" {
		return report("export", extension.Export(e, args[1]))
	}
	if len(args) != 1 || (args[0] != "install" && args[0] != "uninstall") {
		fmt.Fprintln(os.Stderr, "usage: export install | uninstall | export <folder>")
		return 2
	}
	action := args[0]
	if !filepath.IsAbs(os.Getenv("HOME")) {
		return report(action, errors.New("HOME must be an absolute path"))
	}
	m, err := contract.ParseManifest(e.Manifest, "")
	if err != nil {
		return report(action, err)
	}
	dir := filepath.Join(config.Locate(os.Getenv("HOME"), os.LookupEnv).Data, "extensions")
	name := "ncly-" + m.Slug()
	switch action {
	case "install":
		err = install(dir, name, func(stage string) error {
			// Both programs come from one build. The staged exporter supplies the sidecars.
			build := exec.Command("go", "build", "-o", stage+string(filepath.Separator), "./cmd/"+name, "./internal/export")
			build.Stdout, build.Stderr = os.Stdout, os.Stderr
			if err := build.Run(); err != nil {
				return err
			}
			export := exec.Command(filepath.Join(stage, "export"), "export", stage)
			export.Stdout, export.Stderr = os.Stdout, os.Stderr
			return export.Run()
		})
		if err == nil {
			fmt.Println("installed", filepath.Join(dir, name))
		}
	case "uninstall":
		var removed []string
		removed, err = Uninstall(e, dir)
		for _, path := range removed {
			fmt.Println("removed", path)
		}
		if err == nil && len(removed) == 0 {
			fmt.Printf("uninstall: nothing to remove, since %s holds no %s\n", dir, name)
		}
	}
	return report(action, err)
}

func report(action string, err error) int {
	if err == nil {
		return 0
	}
	fmt.Fprintln(os.Stderr, action+":", err)
	return 1
}
