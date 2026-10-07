package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func verifySource(r release, tree string) error {
	tracked, err := command(tree, nil, "git", "ls-files", "-z")
	if err != nil {
		return err
	}
	prefix := "ncly-" + strings.TrimPrefix(r.Version, "v") + "/"
	expected := make(map[string]bool)
	for _, name := range strings.Split(strings.TrimSuffix(string(tracked), "\x00"), "\x00") {
		data, err := os.ReadFile(filepath.Join(tree, name))
		if err != nil {
			return err
		}
		if actual, ok := r.Source.Files[prefix+name]; !ok || !bytes.Equal(actual.Data, data) {
			return fmt.Errorf("source differs from tracked tree: %s", name)
		}
		info, err := os.Stat(filepath.Join(tree, name))
		if err != nil {
			return err
		}
		if info.Mode().Perm()&0o111 != r.Source.Files[prefix+name].Mode.Perm()&0o111 {
			return fmt.Errorf("source mode differs from tracked tree: %s", name)
		}
		expected[prefix+name] = true
	}
	notices, err := readTree(filepath.Join(tree, ".release/notices"))
	if err != nil {
		return err
	}
	for name := range notices {
		expected[prefix+"THIRD_PARTY_LICENSES/"+name] = true
	}
	for name := range r.Source.Files {
		if !expected[name] {
			return fmt.Errorf("unexpected source archive member: %s", name)
		}
	}
	scratch, err := os.MkdirTemp("", "ncly-source-check-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(scratch) }()
	host, ok := r.Binaries[target{runtime.GOOS, runtime.GOARCH}]
	if !ok {
		return fmt.Errorf("no archive for host platform")
	}
	hostBin := filepath.Join(scratch, "ncly-packaged")
	if err := os.WriteFile(hostBin, host.Files["ncly"].Data, 0o755); err != nil {
		return err
	}
	packagedVersion, err := command(scratch, nil, hostBin, "--version")
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(packagedVersion)) != "ncly "+r.Version {
		return fmt.Errorf("packaged binary reports wrong release version: %s", packagedVersion)
	}
	for name, e := range r.Source.Files {
		if !strings.HasPrefix(name, prefix) {
			return fmt.Errorf("source archive has wrong root: %s", name)
		}
		p := filepath.Join(scratch, strings.TrimPrefix(name, prefix))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, e.Data, e.Mode.Perm()); err != nil {
			return err
		}
	}
	for t, a := range r.Binaries {
		bin := filepath.Join(scratch, "ncly-source-"+t.OS+"-"+t.Arch)
		if _, err := command(scratch, targetEnv(t), "go", "build", "-mod=readonly", "-trimpath", "-buildvcs=false", "-ldflags=-s -w -X github.com/pascalandy/nicely/internal/cli.version="+r.Version, "-o", bin, "./cmd/ncly"); err != nil {
			return fmt.Errorf("source build %s/%s: %w", t.OS, t.Arch, err)
		}
		built, err := os.ReadFile(bin)
		if err != nil {
			return err
		}
		if !bytes.Equal(built, a.Files["ncly"].Data) {
			return fmt.Errorf("binary differs from source rebuild: %s/%s", t.OS, t.Arch)
		}
	}
	bin := filepath.Join(scratch, "ncly-source-"+runtime.GOOS+"-"+runtime.GOARCH)
	out, err := command(scratch, nil, bin, "--version")
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(out)) != "ncly "+r.Version {
		return fmt.Errorf("source build reports wrong release version: %s", out)
	}
	return nil
}

func verifyAUR(r release, dist string) error {
	pkg, err := os.ReadFile(filepath.Join(dist, "aur/ncly-bin.pkgbuild"))
	if err != nil {
		return err
	}
	info, err := os.ReadFile(filepath.Join(dist, "aur/ncly-bin.srcinfo"))
	if err != nil {
		return err
	}
	v := strings.TrimPrefix(r.Version, "v")
	if !strings.Contains(string(pkg), "pkgname='ncly-bin'\n") || !strings.Contains(string(pkg), "pkgver="+v+"\n") || !strings.Contains(string(info), "pkgver = "+v+"\n") {
		return fmt.Errorf("AUR package has wrong name or version")
	}
	for _, pair := range []struct{ goarch, aur string }{{"amd64", "x86_64"}, {"arm64", "aarch64"}} {
		a := r.Binaries[target{"linux", pair.goarch}]
		source := "https://github.com/pascalandy/nicely/releases/download/" + r.Version + "/" + a.Name
		for _, want := range []string{"arch = " + pair.aur + "\n", "source_" + pair.aur + " = ncly-bin_" + v + "_" + pair.aur + ".tar.gz::" + source + "\n", fmt.Sprintf("sha256sums_%s = %x\n", pair.aur, a.Hash)} {
			if !strings.Contains(string(info), want) {
				return fmt.Errorf("AUR metadata missing or wrong %s", strings.TrimSpace(want))
			}
		}
		for _, want := range []string{"source_" + pair.aur + "=", fmt.Sprintf("sha256sums_%s=('%x')", pair.aur, a.Hash), strings.ReplaceAll(strings.ReplaceAll(source, r.Version, "v${pkgver}"), v, "${pkgver}")} {
			if !strings.Contains(string(pkg), want) {
				return fmt.Errorf("AUR PKGBUILD missing or wrong %s", want)
			}
		}
	}
	return nil
}
