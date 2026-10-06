// Package config locates Nicely's folders and reads the shared and local
// config files, as the Configuration section of cli-spec.md defines them.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/BurntSushi/toml"
)

// Paths locates Nicely's files and folders on this machine.
type Paths struct {
	// Shared is config.toml, the setup the user wants on every machine.
	Shared string
	// Local is config.local.toml, beside Shared, for this machine only.
	Local string
	Data  string
	State string
	Cache string
}

// Locate resolves the paths from home and the environment. An XDG variable
// counts only when it holds an absolute path, as the XDG specification asks.
// NCLY_CONFIG names another shared file, and the local file moves beside it.
// Without an absolute home or XDG variable, default paths stay empty instead
// of resolving from the current folder.
func Locate(home string, lookupEnv func(string) (string, bool)) Paths {
	dir := func(variable string, fallback ...string) string {
		if value, _ := lookupEnv(variable); filepath.IsAbs(value) {
			return filepath.Join(value, "nicely")
		}
		if !filepath.IsAbs(home) {
			return ""
		}
		return filepath.Join(append(append([]string{home}, fallback...), "nicely")...)
	}
	p := Paths{
		Data:  dir("XDG_DATA_HOME", ".local", "share"),
		State: dir("XDG_STATE_HOME", ".local", "state"),
		Cache: dir("XDG_CACHE_HOME", ".cache"),
	}
	if value, _ := lookupEnv("NCLY_CONFIG"); value != "" {
		p.Shared = value
	} else if config := dir("XDG_CONFIG_HOME", ".config"); config != "" {
		p.Shared = filepath.Join(config, "config.toml")
	}
	if p.Shared != "" {
		p.Local = filepath.Join(filepath.Dir(p.Shared), "config.local.toml")
	}
	return p
}

// Config holds the keys that this version of ncly knows. A zero value means
// the default applies.
type Config struct {
	Lang string `toml:"lang"`
}

// Key is a key of a config file that this version of ncly does not know.
type Key struct {
	File string
	Name string
}

// InvalidError reports a config file with a syntax error or a value of the
// wrong type. Line is 0 when the parser does not say where.
type InvalidError struct {
	File string
	Line int
	Err  error
}

func (e *InvalidError) Error() string { return fmt.Sprintf("%s: %v", e.File, e.Err) }

func (e *InvalidError) Unwrap() error { return e.Err }

// Load reads the shared file, then the local file over it. A missing file is
// not an error. A table merges key by key across the files, and any other
// value, such as a list, replaces the one below it. Unknown keys come back
// for the CONFIG_UNKNOWN_KEY warning, so an older ncly reads a newer config.
func Load(p Paths) (Config, []Key, error) {
	merged := map[string]any{}
	var unknown []Key
	for _, file := range []string{p.Shared, p.Local} {
		if file == "" {
			continue
		}
		text, err := os.ReadFile(file)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return Config{}, nil, &InvalidError{File: file, Err: err}
		}
		var layer map[string]any
		if _, err := toml.Decode(string(text), &layer); err != nil {
			return Config{}, nil, invalid(file, err)
		}
		var typed Config
		meta, err := toml.Decode(string(text), &typed)
		if err != nil {
			return Config{}, nil, invalid(file, err)
		}
		for _, key := range meta.Undecoded() {
			unknown = append(unknown, Key{File: file, Name: key.String()})
		}
		merge(merged, layer)
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(merged); err != nil {
		return Config{}, nil, err
	}
	var cfg Config
	_, err := toml.Decode(buf.String(), &cfg)
	return cfg, unknown, err
}

// merge copies src over dst. Tables merge key by key; other values replace.
func merge(dst, src map[string]any) {
	for key, value := range src {
		table, isTable := value.(map[string]any)
		below, belowIsTable := dst[key].(map[string]any)
		if isTable && belowIsTable {
			merge(below, table)
			continue
		}
		dst[key] = value
	}
}

// typeErrorLine finds the line in a type error, which the TOML parser
// reports as text only.
var typeErrorLine = regexp.MustCompile(`^toml: line (\d+) `)

func invalid(file string, err error) *InvalidError {
	e := &InvalidError{File: file, Err: err}
	var parse toml.ParseError
	if errors.As(err, &parse) {
		e.Line = parse.Position.Line
	} else if m := typeErrorLine.FindStringSubmatch(err.Error()); m != nil {
		e.Line, _ = strconv.Atoi(m[1])
	}
	return e
}
