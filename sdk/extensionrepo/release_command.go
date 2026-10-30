package extensionrepo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// releaseProgram resolves PATH from the child's environment, before exec.Command reads the caller's PATH.
func releaseProgram(env []string, name string) (string, error) {
	if strings.ContainsRune(name, filepath.Separator) {
		return name, nil
	}
	for _, entry := range env {
		search, ok := strings.CutPrefix(entry, "PATH=")
		if !ok {
			continue
		}
		for _, folder := range filepath.SplitList(search) {
			candidate := filepath.Join(folder, name)
			if !filepath.IsAbs(candidate) {
				continue
			}
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
				return candidate, nil
			}
		}
	}
	return "", fmt.Errorf("%s is not on the command's PATH", name)
}

func releaseOutput(ctx context.Context, env []string, dir, name string, args ...string) (string, error) {
	program, err := releaseProgram(env, name)
	if err != nil {
		return "", err
	}
	return signoffOutput(ctx, env, dir, program, args...)
}
