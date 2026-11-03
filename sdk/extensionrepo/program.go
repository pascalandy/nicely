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

// Main runs a repository task, then exits.
// The export subcommand writes embedded sidecars for the staged build.
func Main(e extension.Extension) {
	os.Exit(run(e, os.Args[1:]))
}

// run gives the one result of the action that args names to its caller.
func run(e extension.Extension, args []string) int {
	r, ok := action(e, args)
	if !ok {
		fmt.Fprintln(os.Stderr, "usage: export install | uninstall | signoff | release <tag> [fixture] | check | finish | scenario <name> | acceptance [fixture] | contract | export <folder> | test [go test arguments] | update vX.Y.Z /absolute/repository [--continue]")
		return 2
	}
	return r.deliver(os.Stdout, os.Stderr)
}

// action runs the action that args names, or reports false for no action.
func action(e extension.Extension, args []string) (result, bool) {
	if len(args) == 0 {
		return result{}, false
	}
	switch name, rest := args[0], args[1:]; {
	case name == "release":
		return release(e, rest), true
	case name == "test":
		return test("test", rest, nil), true
	case name == "update":
		// The updater writes its own lines, and gives only its exit code.
		return result{action: "update", exit: update(rest)}, true
	case name == "acceptance" && len(rest) <= 1:
		return acceptance(rest), true
	case name == "signoff" && len(rest) == 0:
		recipes := []string{"check"}
		if len(e.Manifest) != 0 {
			// An extension repository signs no commit before the loop of its
			// skill is done (#413). Core has no just finish.
			recipes = []string{"finish", "check"}
		}
		return signoff(recipes), true
	case name == "check" && len(rest) == 0:
		return checkRepository(), true
	case name == "finish" && len(rest) == 0:
		return finishRepository(), true
	case name == "scenario" && len(rest) == 1:
		return runScenario(rest[0]), true
	case name == "contract" && len(rest) == 0:
		return contractDirectory(), true
	case name == "export" && len(rest) == 1:
		return result{action: "export", err: extension.Export(e, rest[0])}, true
	case (name == "install" || name == "uninstall") && len(rest) == 0:
		return installAction(e, name), true
	}
	return result{}, false
}

// installAction runs just install or just uninstall in the extensions folder
// of the caller's data folder.
func installAction(e extension.Extension, action string) result {
	if !filepath.IsAbs(os.Getenv("HOME")) {
		return failed(action, errors.New("HOME must be an absolute path"))
	}
	m, err := contract.ParseManifest(e.Manifest, "")
	if err != nil {
		return failed(action, err)
	}
	dir := filepath.Join(config.Locate(os.Getenv("HOME"), os.LookupEnv).Data, "extensions")
	name := "ncly-" + m.Slug()
	r := result{action: action}
	switch action {
	case "install":
		r.err = install(dir, name, func(stage string) error {
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
		if r.err == nil {
			r.answer = []string{"installed " + filepath.Join(dir, name)}
		}
	case "uninstall":
		var removed []string
		removed, r.err = Uninstall(e, dir)
		for _, path := range removed {
			r.answer = append(r.answer, "removed "+path)
		}
		if r.err == nil && len(removed) == 0 {
			r.answer = []string{fmt.Sprintf("uninstall: nothing to remove, since %s holds no %s", dir, name)}
		}
	}
	return r
}

func report(action string, err error) int {
	return failed(action, err).deliver(os.Stdout, os.Stderr)
}
