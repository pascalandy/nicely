package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

func command(dir string, env []string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = slices.DeleteFunc(os.Environ(), func(s string) bool {
		for _, e := range env {
			if strings.HasPrefix(s, strings.SplitN(e, "=", 2)[0]+"=") {
				return true
			}
		}
		return false
	})
	cmd.Env = append(cmd.Env, env...)
	out, err := cmd.Output()
	if err != nil {
		var failure *exec.ExitError
		if errors.As(err, &failure) {
			return nil, fmt.Errorf("%s %v: %w\n%s", name, args, err, failure.Stderr)
		}
		return nil, fmt.Errorf("%s %v: %w", name, args, err)
	}
	return out, nil
}

func hostEnv() []string {
	return []string{"GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH, "CGO_ENABLED=0"}
}

func targetEnv(t target) []string {
	return []string{"GOOS=" + t.OS, "GOARCH=" + t.Arch, "CGO_ENABLED=0"}
}

func configuredTargets(tree string) ([]target, error) {
	data, err := os.ReadFile(filepath.Join(tree, ".goreleaser.yaml"))
	if err != nil {
		return nil, err
	}
	var config struct {
		Builds []struct {
			GOOS   []string `yaml:"goos"`
			GOARCH []string `yaml:"goarch"`
		} `yaml:"builds"`
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	var targets []target
	for _, b := range config.Builds {
		for _, os := range b.GOOS {
			for _, arch := range b.GOARCH {
				t := target{os, arch}
				if slices.Contains(targets, t) {
					return nil, fmt.Errorf("duplicate configured target %s/%s", os, arch)
				}
				targets = append(targets, t)
			}
		}
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("release has no configured targets")
	}
	return targets, nil
}

func prepare(version, tree string) error {
	tree, err := filepath.Abs(tree)
	if err != nil {
		return err
	}
	targets, err := configuredTargets(tree)
	if err != nil {
		return err
	}
	root := filepath.Join(tree, ".release")
	if err := os.Mkdir(root, 0o755); err != nil {
		return fmt.Errorf("release assets require a fresh tree: %w", err)
	}
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0o755); err != nil {
		return err
	}
	ncly := filepath.Join(bin, "ncly")
	if _, err := command(tree, hostEnv(), "go", "build", "-mod=readonly", "-buildvcs=false", "-ldflags=-X github.com/pascalandy/nicely/internal/cli.version="+version, "-o", ncly, "./cmd/ncly"); err != nil {
		return err
	}
	licenses := filepath.Join(bin, "go-licenses")
	if _, err := command(filepath.Join(tree, "tools"), hostEnv(), "go", "build", "-mod=readonly", "-buildvcs=false", "-o", licenses, "github.com/google/go-licenses/v2"); err != nil {
		return err
	}
	completionDir := filepath.Join(root, "completions")
	if err := os.Mkdir(completionDir, 0o755); err != nil {
		return err
	}
	for _, c := range []struct{ shell, file string }{{"bash", "ncly.bash"}, {"zsh", "_ncly"}, {"fish", "ncly.fish"}} {
		data, err := command(tree, nil, ncly, "completion", c.shell)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return fmt.Errorf("empty %s completion", c.shell)
		}
		if err := os.WriteFile(filepath.Join(completionDir, c.file), data, 0o644); err != nil {
			return err
		}
	}
	union := make(map[string][]byte)
	for _, t := range targets {
		dir := filepath.Join(root, "notices-"+t.OS+"-"+t.Arch)
		if _, err := command(tree, targetEnv(t), licenses, "save", "./cmd/ncly", "--ignore=github.com/pascalandy/nicely", "--save_path="+dir); err != nil {
			return err
		}
		files, err := readTree(dir)
		if err != nil {
			return err
		}
		for name, data := range files {
			if previous, ok := union[name]; ok && !bytes.Equal(previous, data) {
				return fmt.Errorf("target notice content differs: %s", name)
			}
			union[name] = data
		}
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	for name, data := range union {
		p := filepath.Join(root, "notices", name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, data, 0o644); err != nil {
			return err
		}
	}
	return checkSourceNotices(tree, filepath.Join(root, "notices"), targets)
}

func checkSourceNotices(tree, notices string, targets []target) error {
	collected, err := readTree(notices)
	if err != nil {
		return err
	}
	checked := make(map[string]bool)
	for _, t := range targets {
		data, err := command(tree, targetEnv(t), "go", "list", "-mod=readonly", "-buildvcs=false", "-deps", "-json", "./cmd/ncly")
		if err != nil {
			return err
		}
		decoder := json.NewDecoder(bytes.NewReader(data))
		for {
			var pkg struct {
				Dir      string
				Standard bool
				Module   *struct {
					Dir  string
					Main bool
					Path string
				}
			}
			if err := decoder.Decode(&pkg); err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			if pkg.Standard || pkg.Module == nil || pkg.Module.Main {
				continue
			}
			foundLicense := false
			for dir := pkg.Dir; strings.HasPrefix(dir+string(filepath.Separator), pkg.Module.Dir+string(filepath.Separator)); dir = filepath.Dir(dir) {
				entries, err := os.ReadDir(dir)
				if err != nil {
					return err
				}
				var licenseNames []string
				for _, e := range entries {
					if !e.IsDir() && licenseName(e.Name()) {
						licenseNames = append(licenseNames, e.Name())
					}
				}
				if len(licenseNames) == 0 {
					if dir == pkg.Module.Dir {
						break
					}
					continue
				}
				foundLicense = true
				for _, e := range entries {
					name := e.Name()
					if e.IsDir() || (!licenseName(name) && !noticeName(name)) {
						continue
					}
					p := filepath.Join(dir, name)
					if checked[p] {
						continue
					}
					checked[p] = true
					original, err := os.ReadFile(p)
					if err != nil {
						return err
					}
					found := false
					for saved, content := range collected {
						if strings.HasPrefix(saved, pkg.Module.Path+"/") && filepath.Base(saved) == name && bytes.Equal(original, content) {
							found = true
							break
						}
					}
					if !found {
						return fmt.Errorf("runtime dependency license or NOTICE missing: %s", p)
					}
				}
				break
			}
			if !foundLicense {
				return fmt.Errorf("runtime dependency has no recognized license: %s", pkg.Dir)
			}
		}
	}
	return nil
}

func licenseName(name string) bool {
	n := strings.ToLower(name)
	return n == "license" || n == "license.md" || n == "license.txt" || n == "copying" || n == "copying.md" || n == "copying.txt"
}

func noticeName(name string) bool {
	n := strings.ToLower(name)
	return n == "notice" || n == "notice.md" || n == "notice.txt"
}
