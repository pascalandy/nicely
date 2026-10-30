package extensionrepo

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func checkRepository() error {
	if err := updateRun("go", "mod", "tidy", "-diff"); err != nil {
		return err
	}
	files, err := exec.Command("gofmt", "-l", ".").Output()
	if err != nil {
		return err
	}
	if len(files) != 0 {
		fmt.Fprint(os.Stderr, string(files))
		return errors.New("run gofmt -w . to format these files")
	}
	if err := updateRun("go", "vet", "./..."); err != nil {
		return err
	}
	return updateRun("go", "test", "-count=1", "./...")
}

func printContractDirectory() error {
	if err := updateRun("go", "mod", "download", nicelyModule); err != nil {
		return err
	}
	dir, err := updateOutput("go", "list", "-m", "-f", "{{.Dir}}", nicelyModule)
	if err != nil {
		return err
	}
	path := filepath.Join(strings.TrimSpace(string(dir)), "docs", "north-star", "contract")
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	fmt.Println(path)
	return nil
}
