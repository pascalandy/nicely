package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"debug/buildinfo"
	"debug/elf"
	"debug/macho"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

type entry struct {
	Data []byte
	Mode fs.FileMode
}

func fileHash(filename string) ([32]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(data), nil
}

func readArchive(filename string) (map[string]entry, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	files := make(map[string]entry)
	seen := make(map[string]bool)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(h.Name, "/")
		if name == "" || path.Clean(name) != name || path.IsAbs(name) || name == ".." || strings.HasPrefix(name, "../") || strings.Contains(name, "\\") {
			return nil, fmt.Errorf("unsafe archive path %q", h.Name)
		}
		if seen[name] {
			return nil, fmt.Errorf("duplicate archive path %q", name)
		}
		seen[name] = true
		if h.Typeflag == tar.TypeDir || h.Typeflag == tar.TypeXGlobalHeader {
			continue
		}
		if h.Typeflag != tar.TypeReg {
			return nil, fmt.Errorf("unsafe archive entry %q", name)
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		files[name] = entry{data, fs.FileMode(h.Mode)}
	}
	// Reading past the last tar entry checks the gzip trailer checksum.
	if _, err := io.Copy(io.Discard, gz); err != nil {
		return nil, err
	}
	return files, nil
}

// releaseTargets is the closed set a release ships. .goreleaser.yaml builds
// that set, and inspection rejects any other count.
var releaseTargets = []target{
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"linux", "amd64"},
	{"linux", "arm64"},
}

func inspect(version, tree, dist string) (release, error) {
	r := release{Version: version, Binaries: make(map[target]artifact)}
	names, err := filepath.Glob(filepath.Join(dist, "*.tar.gz"))
	if err != nil {
		return r, err
	}
	sums, err := readChecksums(dist)
	if err != nil {
		return r, err
	}
	numeric := strings.TrimPrefix(version, "v")
	for _, filename := range names {
		hash, err := fileHash(filename)
		if err != nil {
			return r, err
		}
		name := filepath.Base(filename)
		if sums[name] != fmt.Sprintf("%x", hash) {
			return r, fmt.Errorf("checksum mismatch for %s", name)
		}
		delete(sums, name)
		files, err := readArchive(filename)
		if err != nil {
			return r, fmt.Errorf("%s: %w", name, err)
		}
		a := artifact{name, files}
		if name == "ncly_"+numeric+"_source.tar.gz" {
			r.Source = a
			continue
		}
		binary, ok := files["ncly"]
		if !ok || binary.Mode&0o111 != 0o111 {
			return r, fmt.Errorf("%s missing executable ncly", name)
		}
		t, err := binaryTarget(binary.Data)
		if err != nil {
			return r, fmt.Errorf("%s: %w", name, err)
		}
		if _, exists := r.Binaries[t]; exists {
			return r, fmt.Errorf("duplicate target %s/%s", t.OS, t.Arch)
		}
		r.Binaries[t] = a
		if err := compareFile(files, "LICENSE", filepath.Join(tree, "LICENSE")); err != nil {
			return r, err
		}
		if err := compareTree(files, "completions", filepath.Join(tree, ".release/completions")); err != nil {
			return r, err
		}
		if err := compareTree(files, "THIRD_PARTY_LICENSES", filepath.Join(tree, ".release/notices")); err != nil {
			return r, err
		}
	}
	if len(r.Binaries) != len(releaseTargets) {
		return r, fmt.Errorf("release needs exactly %d binary targets; got %d", len(releaseTargets), len(r.Binaries))
	}
	for t, a := range r.Binaries {
		if a.Name != "ncly_"+numeric+"_"+t.OS+"_"+t.Arch+".tar.gz" {
			return r, fmt.Errorf("archive name differs from binary target: %s", a.Name)
		}
	}
	if r.Source.Name == "" {
		return r, fmt.Errorf("missing source archive")
	}
	if len(sums) != 0 {
		return r, fmt.Errorf("checksum references absent archive")
	}
	if err := compareTree(r.Source.Files, "ncly-"+numeric+"/THIRD_PARTY_LICENSES", filepath.Join(tree, ".release/notices")); err != nil {
		return r, err
	}
	return r, checkSourceNotices(tree, filepath.Join(tree, ".release/notices"), releaseTargets)
}

func readChecksums(dist string) (map[string]string, error) {
	paths, err := filepath.Glob(filepath.Join(dist, "*checksums.txt"))
	if err != nil {
		return nil, err
	}
	if len(paths) != 1 {
		return nil, fmt.Errorf("expected one checksum file")
	}
	data, err := os.ReadFile(paths[0])
	if err != nil {
		return nil, err
	}
	sums := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || len(fields[0]) != 64 || fields[1] != filepath.Base(fields[1]) || sums[fields[1]] != "" {
			return nil, fmt.Errorf("invalid checksum record %q", line)
		}
		sums[fields[1]] = fields[0]
	}
	return sums, nil
}

func binaryTarget(data []byte) (target, error) {
	info, err := buildinfo.Read(bytes.NewReader(data))
	if err != nil {
		return target{}, fmt.Errorf("binary has no Go build info: %w", err)
	}
	settings := make(map[string]string)
	for _, s := range info.Settings {
		settings[s.Key] = s.Value
	}
	t := target{settings["GOOS"], settings["GOARCH"]}
	if !slices.Contains(releaseTargets, t) {
		return t, fmt.Errorf("unsupported binary target %s/%s", t.OS, t.Arch)
	}
	if settings["CGO_ENABLED"] != "0" {
		return t, fmt.Errorf("binary uses CGO")
	}
	if t.OS == "linux" {
		f, err := elf.NewFile(bytes.NewReader(data))
		if err != nil {
			return t, err
		}
		defer func() { _ = f.Close() }()
		expected := elf.EM_X86_64
		if t.Arch == "arm64" {
			expected = elf.EM_AARCH64
		}
		if f.Machine != expected {
			return t, fmt.Errorf("ELF architecture differs from Go target")
		}
	} else {
		f, err := macho.NewFile(bytes.NewReader(data))
		if err != nil {
			return t, err
		}
		defer func() { _ = f.Close() }()
		expected := macho.CpuAmd64
		if t.Arch == "arm64" {
			expected = macho.CpuArm64
		}
		if f.Cpu != expected {
			return t, fmt.Errorf("Mach-O architecture differs from Go target")
		}
	}
	return t, nil
}

func compareFile(files map[string]entry, name, source string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	actual, ok := files[name]
	if !ok || !bytes.Equal(actual.Data, data) {
		return fmt.Errorf("archive missing or changed %s", name)
	}
	return nil
}

func compareTree(files map[string]entry, prefix, source string) error {
	expected, err := readTree(source)
	if err != nil {
		return err
	}
	if len(expected) == 0 {
		return fmt.Errorf("empty release asset tree %s", prefix)
	}
	for name, data := range expected {
		a, ok := files[prefix+"/"+name]
		if !ok || !bytes.Equal(a.Data, data) {
			return fmt.Errorf("archive missing or changed %s/%s", prefix, name)
		}
	}
	for name := range files {
		if strings.HasPrefix(name, prefix+"/") {
			if _, ok := expected[strings.TrimPrefix(name, prefix+"/")]; !ok {
				return fmt.Errorf("unexpected release asset %s", name)
			}
		}
	}
	return nil
}

func readTree(root string) (map[string][]byte, error) {
	files := make(map[string][]byte)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("release asset is not a regular file: %s", p)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		name, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(name)] = data
		return nil
	})
	return files, err
}
