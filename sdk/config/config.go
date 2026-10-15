// Package config locates Nicely's folders and reads the shared and local
// config files, as the Configuration section of contract.md defines them.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

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

// layer is one config file that was read, with its keys.
type layer struct {
	file   string
	values map[string]any
}

// read reads the shared file, then the local file, and checks each against
// the keys that this version knows. A missing file is skipped. Unknown keys
// come back for the CONFIG_UNKNOWN_KEY warning.
func read(p Paths) ([]layer, []Key, error) {
	var layers []layer
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
			return nil, nil, &InvalidError{File: file, Err: err}
		}
		var values map[string]any
		if _, err := toml.Decode(string(text), &values); err != nil {
			return nil, nil, invalid(file, err)
		}
		if err := finite("", values); err != nil {
			return nil, nil, &InvalidError{File: file, Err: err}
		}
		var typed Config
		meta, err := toml.Decode(string(text), &typed)
		if err != nil {
			return nil, nil, invalid(file, err)
		}
		for _, key := range meta.Undecoded() {
			unknown = append(unknown, Key{File: file, Name: key.String()})
		}
		layers = append(layers, layer{file: file, values: values})
	}
	return layers, unknown, nil
}

// Load reads the shared file, then the local file over it. A missing file is
// not an error. A table merges key by key across the files, and any other
// value, such as a list, replaces the one below it. Unknown keys come back
// for the CONFIG_UNKNOWN_KEY warning, so an older ncly reads a newer config.
func Load(p Paths) (Config, []Key, error) {
	layers, unknown, err := read(p)
	if err != nil {
		return Config{}, nil, err
	}
	merged := map[string]any{}
	for _, l := range layers {
		merge(merged, l.values)
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(merged); err != nil {
		return Config{}, nil, err
	}
	var cfg Config
	_, err = toml.Decode(buf.String(), &cfg)
	return cfg, unknown, err
}

// Default is the source of a value that no flag, variable, or file sets.
const Default = "default"

// Value is one resolved value with its source: a flag such as --lang, a
// variable such as NCLY_LANG, the path of the file that set it, or Default.
type Value struct {
	Key    string `json:"key"`
	Value  any    `json:"value"`
	Source string `json:"source"`
}

// Values holds every value of the config files by dotted key, such as
// skill.paths, each with the file that set it. A table with keys is no
// value of its own: each of its keys keeps its source. An empty table with
// no key beneath it is the value {} of its file.
type Values map[string]Value

// LoadValues reads the files as Load does, and keeps the source of each
// value: a table merges key by key, and any other value, an empty list
// included, keeps the file that replaced the one below it.
func LoadValues(p Paths) (Values, error) {
	layers, _, err := read(p)
	if err != nil {
		return nil, err
	}
	v := Values{}
	for _, l := range layers {
		v.add("", l.values, l.file)
	}
	return v, nil
}

// add sets each value of a table under prefix, replacing what a lower
// file set at its key, beneath it, or above it. A table merges key by key,
// so an empty one adds nothing over the keys that a lower file set beneath
// it, and otherwise is the value of its key, unlike an absent key.
func (v Values) add(prefix string, table map[string]any, file string) {
	for key, value := range table {
		key = prefix + key
		if sub, isTable := value.(map[string]any); isTable && (len(sub) > 0 || v.beneath(key)) {
			delete(v, key)
			v.add(key+".", sub, file)
			continue
		}
		for other := range v {
			if strings.HasPrefix(other, key+".") || strings.HasPrefix(key, other+".") {
				delete(v, other)
			}
		}
		v[key] = Value{Key: key, Value: value, Source: file}
	}
}

// beneath reports whether a value lies beneath key.
func (v Values) beneath(key string) bool {
	for other := range v {
		if strings.HasPrefix(other, key+".") {
			return true
		}
	}
	return false
}

// Under returns the values at key and beneath it, in the order of their
// keys, or one value of key from Default when no file sets any: a declared
// key such as agent.profiles covers every key beneath it.
func (v Values) Under(key string) []Value {
	var found []Value
	for _, k := range slices.Sorted(maps.Keys(v)) {
		if k == key || strings.HasPrefix(k, key+".") {
			found = append(found, v[k])
		}
	}
	if len(found) == 0 {
		return []Value{{Key: key, Source: Default}}
	}
	return found
}

// Decode decodes the value at key, such as skill.paths, into target, a
// pointer to the type that the key holds, as Load decodes the keys of core.
// It returns the value with the file that set it, or one from Default that
// leaves target unchanged when no file sets key. A value of another type is
// an InvalidError about the file that set it, and so is a table at key, or
// a value where a table above key belongs.
func (v Values) Decode(key string, target any) (Value, error) {
	words := strings.Split(key, ".")
	for i := 1; i < len(words); i++ {
		if above, found := v[strings.Join(words[:i], ".")]; found {
			return above, &InvalidError{File: above.Source, Err: fmt.Errorf("%q is not a table", above.Key)}
		}
	}
	set, found := v[key]
	if !found {
		if beneath := v.Under(key); beneath[0].Source != Default {
			return beneath[0], &InvalidError{File: beneath[0].Source, Err: fmt.Errorf("%q is a table", key)}
		}
		return Value{Key: key, Source: Default}, nil
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(map[string]any{"value": set.Value}); err != nil {
		return set, &InvalidError{File: set.Source, Err: err}
	}
	var wrapper struct {
		Value toml.Primitive `toml:"value"`
	}
	meta, err := toml.Decode(buf.String(), &wrapper)
	if err == nil {
		err = meta.PrimitiveDecode(wrapper.Value, target)
	}
	if err != nil {
		return set, &InvalidError{File: set.Source, Err: fmt.Errorf("%q: %w", key, err)}
	}
	return set, nil
}

// Path resolves a path that a config file holds: an absolute path as is,
// and one that starts with ~/ from home. Any other path is an error,
// because its meaning would depend on the folder where ncly runs.
func Path(home, path string) (string, error) {
	if rest, found := strings.CutPrefix(path, "~/"); found && filepath.IsAbs(home) {
		return filepath.Join(home, rest), nil
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return "", fmt.Errorf("%q is neither absolute nor starts with ~/", path)
}

// Lang resolves the language with its source, in the order of
// Configuration in contract.md: the flag, NCLY_LANG, lang in the config
// files, LC_ALL, LC_MESSAGES, LANG, then en from Default. files reads the
// config only when no flag or NCLY_LANG sets the language, and returns nil
// when the files cannot be read.
func Lang(flag string, getenv func(string) string, files func() Values) Value {
	if flag != "" {
		return Value{Key: "lang", Value: flag, Source: "--lang"}
	}
	if lang := getenv("NCLY_LANG"); lang != "" {
		return Value{Key: "lang", Value: lang, Source: "NCLY_LANG"}
	}
	if set, found := files()["lang"]; found {
		if lang, isString := set.Value.(string); isString && lang != "" {
			return set
		}
	}
	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if lang := getenv(name); lang != "" {
			return Value{Key: "lang", Value: lang, Source: name}
		}
	}
	return Value{Key: "lang", Value: "en", Source: Default}
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

// finite reports the first number that JSON cannot hold, nan or inf, at
// or beneath key. A config value reaches JSON in the answers of describe and
// of extensions, so such a number is a value of the wrong type.
func finite(key string, value any) error {
	var items []any
	switch v := value.(type) {
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("%q holds a number that is not finite", key)
		}
	case map[string]any:
		for _, name := range slices.Sorted(maps.Keys(v)) {
			if err := finite(strings.TrimPrefix(key+"."+name, "."), v[name]); err != nil {
				return err
			}
		}
	case []map[string]any:
		for _, table := range v {
			items = append(items, table)
		}
	case []any:
		items = v
	}
	for _, item := range items {
		if err := finite(key, item); err != nil {
			return err
		}
	}
	return nil
}
