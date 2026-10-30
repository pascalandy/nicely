package extensionrepo

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/config"
)

type signoffCheck struct {
	Command []string `json:"command"`
	Exit    int      `json:"exit"`
}

type signoffReceipt struct {
	SHA     string            `json:"sha"`
	OS      string            `json:"os"`
	Tools   map[string]string `json:"tools"`
	Checks  []signoffCheck    `json:"checks"`
	Signoff *signoffCheck     `json:"signoff"`
}

func (r signoffReceipt) write(run string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(run, "receipt.json"), append(data, '\n'), 0o600)
}

func newSignoffRun(root, sha string) (string, error) {
	cache := config.Locate(os.Getenv("HOME"), os.LookupEnv).Cache
	if !filepath.IsAbs(cache) {
		return "", errors.New("signoff needs an absolute cache folder outside the repository")
	}
	// Resolve symlinks after creation, so a cache link cannot place receipts inside tracked files.
	base := filepath.Join(cache, "signoff")
	if err := os.MkdirAll(base, 0o700); err != nil {
		return "", err
	}
	base, err := filepath.EvalSymlinks(base)
	if err != nil {
		return "", err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, base)
	if err != nil {
		return "", err
	}
	if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("cache folder is inside the repository; set XDG_CACHE_HOME outside it and repeat just signoff")
	}
	return os.MkdirTemp(base, time.Now().UTC().Format("20060102T150405Z")+"-"+sha+"-")
}
