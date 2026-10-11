package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"slices"
	"strings"
)

// Reply is what an extension wrote under --json, as core holds it once the
// extension has exited.
type Reply struct {
	Exit   int
	Stdout []byte
	// Last is the last line of stderr, the answer of a failure.
	Last []byte
}

// OnStdout reports whether the answer of a command of this kind belongs on
// stdout: a success, and a report whose findings need a human.
func (r Reply) OnStdout(kind ResultKind) bool {
	return r.Exit == int(ExitOK) || (kind == ResultReport && r.Exit == int(ExitHuman) && len(r.Stdout) > 0)
}

// ExtensionAnswer is an answer that passed the answer check.
type ExtensionAnswer struct {
	// Line is the answer as the extension wrote it, on the stream that
	// OnStdout names.
	Line []byte
	keys map[string]json.RawMessage
}

// Failed renders the answer as a failure that lead decides: ok false, an
// error with lead's code first, and every other key kept, so the caller
// keeps the data of the work that happened, as Output in contract.md asks.
// lead goes first when the answer holds no error with its code.
func (a ExtensionAnswer) Failed(lead Problem) ([]byte, error) {
	var errs []json.RawMessage
	if raw, present := a.keys["errors"]; present {
		if err := json.Unmarshal(raw, &errs); err != nil {
			return nil, err
		}
	}
	i := slices.IndexFunc(errs, func(raw json.RawMessage) bool {
		var p Problem
		return json.Unmarshal(raw, &p) == nil && p.Code == lead.Code
	})
	if i < 0 {
		first, err := json.Marshal(lead)
		if err != nil {
			return nil, err
		}
		errs = append([]json.RawMessage{first}, errs...)
	} else {
		first := errs[i]
		errs = append([]json.RawMessage{first}, slices.Delete(errs, i, i+1)...)
	}
	o := object{}
	o.add("ok", false)
	o.add("contract_version", Version)
	for _, key := range slices.Sorted(maps.Keys(a.keys)) {
		if !slices.Contains(reserved, key) {
			o.addRaw(key, a.keys[key], nil)
		}
	}
	if raw, present := a.keys["results"]; present {
		o.addRaw("results", raw, nil)
	}
	o.add("errors", errs)
	if raw, present := a.keys["warnings"]; present {
		o.addRaw("warnings", raw, nil)
	}
	return o.bytes()
}

// CheckAnswer applies the rules of Answer check in contract.md to the reply
// of a command of the extension that the manifest declares. A broken rule
// returns a *RuleError.
func CheckAnswer(m Manifest, kind ResultKind, r Reply) (ExtensionAnswer, error) {
	exit := ExitCode(r.Exit)
	fail := func(rule string, data ...map[string]any) (ExtensionAnswer, error) {
		all := map[string]any{"Domain": m.Domain, "Exit": r.Exit}
		for _, more := range data {
			maps.Copy(all, more)
		}
		return ExtensionAnswer{}, broken(rule, all)
	}
	if !exit.Known() {
		return fail("extension.answer_exit")
	}
	onStdout := r.OnStdout(kind)
	line, stream := r.Stdout, "stdout"
	if !onStdout {
		if len(r.Stdout) > 0 {
			return fail("extension.answer_two_places")
		}
		line, stream = r.Last, "stderr"
	}
	if len(line) == 0 {
		return fail("extension.answer_missing", map[string]any{"Stream": stream})
	}
	if bytes.Contains(bytes.TrimSuffix(line, []byte("\n")), []byte("\n")) {
		return fail("extension.answer_lines", map[string]any{"Stream": stream})
	}
	keys, err := decodeObject(line)
	if err != nil {
		return fail("extension.answer_json")
	}
	if ok, isBool := jsonBool(keys["ok"]); !isBool || ok != (exit == ExitOK) {
		return fail("extension.answer_ok")
	}
	if string(keys["contract_version"]) != "1" {
		return fail("extension.answer_version")
	}
	a := answerParts{}
	if err := a.decode(keys); err != nil {
		return fail(err.Error())
	}
	switch {
	case onStdout && len(a.errors) > 0:
		return fail("extension.answer_errors_stdout")
	case !onStdout && len(a.errors) == 0:
		return fail("extension.answer_errors_missing")
	}
	if code, invalid := a.invalidCode(); invalid {
		return fail("extension.answer_code", map[string]any{"Code": code})
	}
	for _, it := range a.results {
		if it.ok == (len(it.errors) > 0) {
			return fail("extension.answer_item_ok")
		}
		if !it.ok && exit == ExitOK {
			return fail("extension.answer_item_failed")
		}
	}
	own := func(c Code) (ExitCode, bool) {
		i := slices.IndexFunc(m.Errors, func(o OwnCode) bool { return o.Code == c })
		if i < 0 {
			return c.Exit()
		}
		return m.Errors[i].Exit, true
	}
	for _, p := range a.allErrors() {
		if _, known := own(p.Code); !known && strings.HasPrefix(string(p.Code), m.codePrefix()) {
			return fail("extension.answer_undeclared", map[string]any{"Code": string(p.Code)})
		}
	}
	for i, p := range a.errors {
		want, known := own(p.Code)
		if known && (i == 0 && want != exit || cautious(want) < cautious(exit)) {
			return fail("extension.answer_code_exit", map[string]any{"Code": string(p.Code), "Want": int(want)})
		}
	}
	if a.itemFailed() {
		var lead Code
		if len(a.errors) > 0 {
			lead = a.errors[0].Code
		}
		batch := map[ExitCode]Code{ExitRuntime: Runtime, ExitTemporary: Temporary, ExitInterrupted: Interrupted, ExitTerminated: Terminated}
		if want, allowed := batch[exit]; !allowed || lead != want {
			return fail("extension.answer_batch", map[string]any{"Code": string(lead)})
		}
	}
	if exit == ExitTemporary {
		for _, p := range a.allErrors() {
			if p.Code != Temporary {
				return fail("extension.answer_temporary", map[string]any{"Code": string(p.Code)})
			}
		}
	}
	return ExtensionAnswer{Line: line, keys: keys}, nil
}

// cautious ranks an exit code by caution, as Output in contract.md orders
// the codes that compete to lead: a lower rank is more cautious. 130 and 143
// share a rank, because the first signal leads.
func cautious(e ExitCode) int {
	if e == ExitTerminated {
		e = ExitInterrupted
	}
	return slices.Index([]ExitCode{ExitInterrupted, ExitRuntime, ExitHuman, ExitUsage, ExitTemporary, ExitOK}, e)
}

// decodeObject decodes one JSON object and nothing after it.
func decodeObject(line []byte) (map[string]json.RawMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(line))
	var keys map[string]json.RawMessage
	if err := dec.Decode(&keys); err != nil || keys == nil {
		return nil, errors.New("not an object")
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("more than one value")
	}
	return keys, nil
}

// answerParts holds the parts of an answer that the rules read.
type answerParts struct {
	errors, warnings []Problem
	results          []itemParts
}

type itemParts struct {
	ok     bool
	errors []Problem
}

// decode reads errors, warnings, and results. A part of the wrong shape
// returns the catalog ID of the rule that it breaks.
func (a *answerParts) decode(keys map[string]json.RawMessage) error {
	var err error
	if a.errors, err = problemList(keys["errors"]); err != nil {
		return err
	}
	if a.warnings, err = problemList(keys["warnings"]); err != nil {
		return err
	}
	raw, present := keys["results"]
	if !present {
		return nil
	}
	var items []map[string]json.RawMessage
	if json.Unmarshal(raw, &items) != nil || items == nil {
		return errors.New("extension.answer_results")
	}
	for _, item := range items {
		var it itemParts
		var isBool bool
		if it.ok, isBool = jsonBool(item["ok"]); !isBool {
			return errors.New("extension.answer_results")
		}
		if it.errors, err = problemList(item["errors"]); err != nil {
			return err
		}
		a.results = append(a.results, it)
	}
	return nil
}

// problemList reads an array of problems that each hold the strings code,
// message, and hint. An absent key is an empty list.
func problemList(raw json.RawMessage) ([]Problem, error) {
	if raw == nil {
		return nil, nil
	}
	var entries []map[string]json.RawMessage
	if json.Unmarshal(raw, &entries) != nil || entries == nil {
		return nil, errors.New("extension.answer_problem")
	}
	var list []Problem
	for _, entry := range entries {
		var p Problem
		if !jsonString(entry["code"], (*string)(&p.Code)) || !jsonString(entry["message"], &p.Message) || !jsonString(entry["hint"], &p.Hint) {
			return nil, errors.New("extension.answer_problem")
		}
		list = append(list, p)
	}
	return list, nil
}

// jsonBool reads a raw value that must be true or false. Unmarshal would
// read null as false.
func jsonBool(raw json.RawMessage) (value, isBool bool) {
	switch string(raw) {
	case "true":
		return true, true
	case "false":
		return false, true
	}
	return false, false
}

// jsonString reads a raw value that must be a string, which may be empty.
// Unmarshal would read null as "".
func jsonString(raw json.RawMessage, s *string) bool {
	return len(raw) > 0 && raw[0] == '"' && json.Unmarshal(raw, s) == nil
}

// invalidCode returns the first code that is not uppercase letters, digits,
// and underscores, and whether it found one.
func (a answerParts) invalidCode() (Code, bool) {
	for _, p := range append(a.allErrors(), a.warnings...) {
		if !codeForm.MatchString(string(p.Code)) {
			return p.Code, true
		}
	}
	return "", false
}

func (a answerParts) allErrors() []Problem {
	all := slices.Clone(a.errors)
	for _, it := range a.results {
		all = append(all, it.errors...)
	}
	return all
}

func (a answerParts) itemFailed() bool {
	return slices.ContainsFunc(a.results, func(it itemParts) bool { return !it.ok })
}
