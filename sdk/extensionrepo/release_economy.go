package extensionrepo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/rogpeppe/go-internal/txtar"
)

// economy measures declared examples. Acceptance owns proof on the extension's real input.
func (r *releaseReceipt) economy(ctx context.Context, env []string, clone string, m contract.Manifest, fixture string) error {
	archive, err := txtar.ParseFile(filepath.Join(clone, "testdata", "examples.txtar"))
	if os.IsNotExist(err) {
		archive, err = &txtar.Archive{}, nil
	}
	if err != nil {
		return err
	}
	type sample struct {
		command []string
		budget  int
		files   []txtar.File
		input   string
	}
	samples := []sample{
		{[]string{"ncly", "skill", "view", "nicely"}, 2500, nil, ""},
		{[]string{"ncly", "skill", "view", m.Slug()}, 1600, nil, ""},
		{[]string{"ncly", "describe"}, 2500, nil, ""},
		{append([]string{"ncly", "describe"}, m.Root()...), 4500, nil, ""},
		{[]string{"ncly", "doctor", m.Slug()}, 500, nil, ""},
	}
	for _, c := range m.Commands {
		if !m.Runs(c) {
			continue
		}
		var command []string
		for _, example := range c.Examples {
			call, err := releaseExample(example)
			if err == nil && len(call) > len(c.Path) && slices.Equal(call[1:1+len(c.Path)], c.Path) {
				command = call
				break
			}
		}
		if command == nil {
			r.Reservations = append(r.Reservations, "Economy: no runnable manifest example for "+strings.Join(c.Path, " "))
			continue
		}
		input := ""
		for _, arg := range command[1+len(c.Path):] {
			if slices.ContainsFunc(archive.Files, func(file txtar.File) bool { return filepath.Clean(file.Name) == filepath.Clean(arg) }) {
				input = arg
				break
			}
		}
		if c.Modes.DryRun {
			samples = append(samples, sample{releaseFlags(command, "--dry-run"), 1000, archive.Files, input})
		}
		samples = append(samples, sample{command, 1000, archive.Files, input})
	}
	for i, sample := range samples {
		for _, jsonOutput := range []bool{false, true} {
			command, format := sample.command, "text"
			if jsonOutput {
				command, format = releaseFlags(command, "--json"), "json"
			}
			row, err := r.economyMeasure(ctx, env, fmt.Sprintf("economy-%d-%s", i, format), fixture, sample.input, sample.files, command...)
			r.Economy = append(r.Economy, row)
			if err != nil {
				return err
			}
			if row.Exit != 0 {
				r.Reservations = append(r.Reservations, fmt.Sprintf("Economy: %s exited %d", strings.Join(command, " "), row.Exit))
			}
			if row.StdoutBytes > sample.budget || row.StderrBytes > 500 {
				r.Reservations = append(r.Reservations, fmt.Sprintf("Economy: %s exceeds a budget: stdout %d/%d bytes, stderr %d/500 bytes", strings.Join(command, " "), row.StdoutBytes, sample.budget, row.StderrBytes))
			}
		}
	}
	return nil
}

// economyMeasure gives each invocation fresh inputs and output paths.
func (r *releaseReceipt) economyMeasure(ctx context.Context, env []string, name, fixture, input string, files []txtar.File, args ...string) (releaseMeasurement, error) {
	work, err := os.MkdirTemp(r.dir, "economy-work-")
	if err != nil {
		return releaseMeasurement{}, err
	}
	defer func() { _ = os.RemoveAll(work) }()
	// Extract only data files. Archive setup and fake programs never enter the command environment.
	for _, file := range files {
		if !filepath.IsLocal(file.Name) {
			return releaseMeasurement{}, fmt.Errorf("examples.txtar file is outside its working folder: %s", file.Name)
		}
		path := filepath.Join(work, file.Name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return releaseMeasurement{}, err
		}
		if err := os.WriteFile(path, file.Data, 0o600); err != nil {
			return releaseMeasurement{}, err
		}
	}
	if fixture != "" && input != "" {
		if err := releaseCopyFixture(fixture, filepath.Join(work, input)); err != nil {
			return releaseMeasurement{}, err
		}
	}
	return r.measure(ctx, env, work, name, args...)
}

// releaseCopyFixture streams media without loading the whole fixture into memory.
func releaseCopyFixture(from, to string) error {
	input, err := os.Open(from)
	if err != nil {
		return err
	}
	defer func() { _ = input.Close() }()
	output, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	_, err = io.Copy(output, input)
	return errors.Join(err, output.Close())
}
