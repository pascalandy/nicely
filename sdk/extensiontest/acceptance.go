package extensiontest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pascalandy/nicely/sdk/extensionrepo"
	"github.com/rogpeppe/go-internal/testscript"
)

// A readCase is the arguments of one reads call.
type readCase struct {
	name, limit            string
	private                []string
	output, reference, bad string
	reader                 []string
}

// parseReads reads reads [-case <case>] [-limit <id>] [-private <text>]... <output>
// <reference> <bad> <reader> [<arg>...].
func parseReads(args []string) (readCase, bool) {
	var c readCase
	for len(args) > 1 && (args[0] == "-case" || args[0] == "-limit" || args[0] == "-private") {
		switch args[0] {
		case "-case":
			c.name = caseName(args[1])
		case "-limit":
			if c.limit != "" || args[1] == "" {
				return c, false
			}
			c.limit = args[1]
		case "-private":
			c.private = append(c.private, args[1])
		}
		args = args[2:]
	}
	if len(args) < 4 || strings.HasPrefix(args[0], "-") {
		return c, false
	}
	c.output, c.reference, c.bad, c.reader = args[0], args[1], args[2], args[3:]
	return c, true
}

// reads checks one acceptance case, as Acceptance cases in the contract
// says. It reads the output with the reader, compares the reference,
// searches each private text, and requires the bad variant to fail.
func (k *kit) reads(ts *testscript.TestScript, neg bool, args []string) {
	c, ok := parseReads(args)
	if neg || !ok {
		ts.Fatalf("usage: reads [-case <case>] [-limit <id>] [-private <text>]... <output> <reference> <bad> <reader> [<arg>...]")
	}
	direct := "ncly-" + k.manifest.Slug()
	if !k.acceptance.wasCalled(ts) {
		ts.Fatalf("reads: no call of ncly or %s ran before it in this scenario, so it reads no output of the extension. Fix: run the command of the case first", direct)
	}
	if program := filepath.Base(c.reader[0]); program == "ncly" || program == direct || slices.Contains(k.manifest.Programs, program) {
		ts.Fatalf("reads: the reader %s is the extension or a program that it wraps, so it can share the bug of the handler. "+
			"Fix: read the output with another program, such as a decoder of its format", c.reader[0])
	}
	if c.limit != "" && !slices.Contains(privateCases, c.name) {
		ts.Fatalf("reads: -limit needs a privacy Output case. Fix: use -case with a privacy case and its accepted limit ID")
	}
	if c.name != "" {
		if !slices.ContainsFunc(outputCases, func(name string) bool { return caseName(name) == c.name }) {
			ts.Fatalf("reads: -case %q names no case of the Output group of Generic cases. Fix: give its name, such as %q", c.name, caseName(outputCases[5]))
		}
		if !slices.ContainsFunc(k.outputs, func(o outputCase) bool { return o.name == c.name }) {
			ts.Fatalf("reads: ## Case matrix of spec.md names no family from the Output case %q. Fix: add the row of its family, with the Source Output: %s", c.name, c.name)
		}
		if c.limit != "" {
			if err := extensionrepo.SpecCheckOutputLimit(k.spec, c.limit, c.name); err != nil {
				ts.Fatalf("reads: %v. Fix: describe the retained data and declare its Output case and acceptance in Limits", err)
			}
		}
		if c.limit != "" && !slices.ContainsFunc(k.outputs, func(o outputCase) bool { return o.name == c.name && slices.Contains(o.limits, c.limit) }) {
			ts.Fatalf("reads: limit %q is not accepted for every matrix row of the Output case %q. Fix: declare the Output case in the accepted Limit row and name its ID in Accepted for every applicable matrix row", c.limit, c.name)
		}
		if slices.Contains(privateCases, c.name) && len(c.private) == 0 && c.limit == "" {
			ts.Fatalf("reads: the case %q needs the private text of its test file, given with -private. Fix: put private text in the block, and give it with -private", c.name)
		}
	}
	work := ts.Getenv("WORK")
	files := map[string]string{}
	for _, path := range []string{c.output, c.reference, c.bad} {
		abs := ts.MkAbs(path)
		if rel, err := filepath.Rel(work, abs); err != nil || !filepath.IsLocal(rel) {
			ts.Fatalf("reads: %s is no path in the folder of the scenario. Fix: make each file of the case in the scenario, never outside it", path)
		}
		files[path] = abs
	}
	output, err := os.ReadFile(files[c.output])
	ts.Check(err)
	reference, err := os.ReadFile(files[c.reference])
	ts.Check(err)
	got, err := k.readWith(ts, c.reader, files[c.output])
	if err != nil {
		ts.Fatalf("reads: the reader %s fails on %s: %v. Fix: correct the output of the handler, or the reader", c.reader[0], c.output, err)
	}
	for _, text := range c.private {
		if bytes.Contains(output, []byte(text)) || strings.Contains(got, text) {
			ts.Fatalf("reads: %s holds the private text %q, or its reader shows it. Fix: remove the private text in the handler, and keep what a reader needs", c.output, text)
		}
	}
	if got != string(reference) {
		ts.Fatalf("reads: %s gives %q for %s, and the reference %s holds %q. Fix: correct the handler, or the reference when it is wrong", c.reader[0], got, c.output, c.reference, reference)
	}
	bad, err := os.ReadFile(files[c.bad])
	ts.Check(err)
	badText, badErr := k.readWith(ts, c.reader, files[c.bad])
	switch {
	case len(c.private) != 0 && !slices.ContainsFunc(c.private, func(text string) bool { return bytes.Contains(bad, []byte(text)) || strings.Contains(badText, text) }):
		ts.Fatalf("reads: the bad variant %s holds no private text, so it cannot prove that the search finds it. "+
			"Fix: make it the output of a handler that keeps the raw block", c.bad)
	case len(c.private) == 0 && badErr == nil && badText == string(reference):
		ts.Fatalf("reads: the bad variant %s passes, so the reader and the reference cannot refuse a wrong output. "+
			"Fix: make it the output of a handler with the bug, and check that the reader shows the difference", c.bad)
	}
	if c.name != "" {
		k.acceptance.prove(c.name)
	}
}

// readWith runs the reader on the file at path, and returns what it
// prints. The word {} stands for the path. Without it, the path is the
// last argument. The reader answer of the kit reads a JSON object.
func (k *kit) readWith(ts *testscript.TestScript, reader []string, path string) (string, error) {
	if reader[0] == "answer" {
		return readAnswer(path, reader[1:])
	}
	args, given := []string{}, false
	for _, arg := range reader[1:] {
		if arg == "{}" {
			arg, given = path, true
		}
		args = append(args, arg)
	}
	if !given {
		args = append(args, path)
	}
	if err := ts.Exec(reader[0], args...); err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(ts.ReadFile("stderr")))
	}
	return ts.ReadFile("stdout"), nil
}

// readAnswer reads the JSON object in the file at path. It prints the line
// <key>: <value> for each key, with the value in compact JSON. A key that
// the object lacks gives <key>: missing.
func readAnswer(path string, keys []string) (string, error) {
	if len(keys) == 0 {
		return "", errors.New("answer reads no key. Give each key to read, as in answer lines")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var object map[string]json.RawMessage
	if err := decoder.Decode(&object); err != nil {
		return "", fmt.Errorf("no JSON object: %w", err)
	}
	var b strings.Builder
	for _, key := range keys {
		value, found := object[key]
		if !found {
			fmt.Fprintf(&b, "%s: missing\n", key)
			continue
		}
		var compact bytes.Buffer
		if err := json.Compact(&compact, value); err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "%s: %s\n", key, compact.String())
	}
	return b.String(), nil
}
