package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
)

// Version is the integer version of the public protocol, carried by every
// answer as contract_version.
const Version = 1

// Problem is one entry of errors or warnings.
type Problem struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint"`
}

// Effect is the strongest change an invocation has made so far, from
// harmless to unsafe to repeat. It covers the whole invocation, including
// the items of a batch that succeeded.
type Effect int

const (
	// NoEffect means nothing changed outside Nicely's cache and the run's
	// own temporary files.
	NoEffect Effect = iota
	// RepeatableEffect means every change is covered by a demonstrated
	// idempotent operation.
	RepeatableEffect
	// NonRepeatableEffect means a second invocation would repeat or conflict
	// with a change.
	NonRepeatableEffect
	// UnknownEffect means a change may have happened, with no known result.
	UnknownEffect
	// PaidEffect means a paid request started, even if its response was lost.
	PaidEffect
)

func (e Effect) safeToRepeat() bool { return e <= RepeatableEffect }

// Signal records the signal that stopped an invocation, if any.
type Signal int

const (
	NoSignal Signal = iota
	Interrupt
	Terminate
)

// Item is the result of one input of a batch.
type Item struct {
	// Data holds the keys of the item, such as url. It marshals to an object.
	Data   any
	Errors []Problem
}

// Outcome is everything a command reports once its work stops.
type Outcome struct {
	// Data holds the command's keys, kept on failure when they still apply,
	// such as output_dir. It marshals to an object.
	Data any
	// Results holds one item per attempted input, in input order. It stays
	// nil for a command that does not process several items.
	Results  []Item
	Errors   []Problem
	Warnings []Problem
	Effect   Effect
	Signal   Signal
}

// Verdict is the exit code of an outcome and the code of its leading error.
type Verdict struct {
	Exit ExitCode
	Lead Code
}

// Decide computes the verdict of an outcome. The order in which failures
// arrived never changes it.
func Decide(o Outcome) Verdict {
	switch o.Signal {
	case Interrupt:
		return Verdict{ExitInterrupted, Interrupted}
	case Terminate:
		return Verdict{ExitTerminated, Terminated}
	}
	itemFailed, allTemporary := false, true
	for _, it := range o.Results {
		for _, p := range it.Errors {
			itemFailed = true
			allTemporary = allTemporary && p.Code == Temporary
		}
	}
	if len(o.Errors) == 0 && !itemFailed {
		return Verdict{ExitOK, ""}
	}
	lead := Temporary
	for _, p := range o.Errors {
		if p.Code.rank() < lead.rank() {
			lead = p.Code
		}
	}
	if lead != Temporary {
		exit, _ := lead.Exit()
		return Verdict{exit, lead}
	}
	if allTemporary && o.Effect.safeToRepeat() {
		return Verdict{ExitTemporary, Temporary}
	}
	return Verdict{ExitRuntime, Runtime}
}

// Answer is a decided outcome, ready to print.
type Answer struct {
	Verdict
	data     any
	results  []Item
	errors   []Problem
	warnings []Problem
}

func (a Answer) OK() bool { return a.Exit == ExitOK }

// Errors returns the errors, the leading one first.
func (a Answer) Errors() []Problem { return a.errors }

func (a Answer) Warnings() []Problem { return a.warnings }

// Finish decides an outcome and puts an error with the verdict's code first.
// When no error carries that code, as for an incomplete batch or a signal,
// summarize supplies one in the active language.
func Finish(o Outcome, summarize func(Code) Problem) Answer {
	mustRegister(o)
	v := Decide(o)
	errs := slices.Clone(o.Errors)
	if v.Lead != "" {
		if i := slices.IndexFunc(errs, func(p Problem) bool { return p.Code == v.Lead }); i >= 0 {
			lead := errs[i]
			errs = append([]Problem{lead}, slices.Delete(errs, i, i+1)...)
		} else {
			errs = append([]Problem{summarize(v.Lead)}, errs...)
		}
	}
	return Answer{Verdict: v, data: o.Data, results: o.Results, errors: errs, warnings: o.Warnings}
}

// mustRegister panics on a code missing from the registry, because core uses
// only the codes that cli-spec.md lists.
func mustRegister(o Outcome) {
	check := func(p Problem, warning bool) {
		_, isError := p.Code.Exit()
		if (warning && !p.Code.IsWarning()) || (!warning && !isError) {
			panic(fmt.Sprintf("contract: unregistered code %q", p.Code))
		}
	}
	for _, p := range o.Errors {
		check(p, false)
	}
	for _, p := range o.Warnings {
		check(p, true)
	}
	for _, it := range o.Results {
		for _, p := range it.Errors {
			check(p, false)
		}
	}
}

// WriteJSON prints the answer as one line: on stdout for a success, and at
// the end of stderr for a failure, after any diagnostics.
func (a Answer) WriteJSON(stdout, stderr io.Writer) error {
	line, err := json.Marshal(a)
	if err != nil {
		return err
	}
	w := stdout
	if !a.OK() {
		w = stderr
	}
	_, err = w.Write(append(line, '\n'))
	return err
}

// MarshalJSON renders ok and contract_version, the command's keys, then
// results, errors, and warnings when they hold entries.
func (a Answer) MarshalJSON() ([]byte, error) {
	o := object{}
	o.add("ok", a.OK())
	o.add("contract_version", Version)
	if err := o.merge(a.data); err != nil {
		return nil, err
	}
	if a.results != nil {
		o.add("results", a.results)
	}
	if len(a.errors) > 0 {
		o.add("errors", a.errors)
	}
	if len(a.warnings) > 0 {
		o.add("warnings", a.warnings)
	}
	return o.bytes()
}

// MarshalJSON renders the item's own ok, its keys, then its errors.
func (it Item) MarshalJSON() ([]byte, error) {
	o := object{}
	o.add("ok", len(it.Errors) == 0)
	if err := o.merge(it.Data); err != nil {
		return nil, err
	}
	if len(it.Errors) > 0 {
		o.add("errors", it.Errors)
	}
	return o.bytes()
}

// reserved holds the keys that the envelope owns, which data cannot use.
var reserved = []string{"ok", "contract_version", "results", "errors", "warnings"}

// object builds a JSON object whose keys keep the order they were added in.
type object struct {
	buf  bytes.Buffer
	keys []string
	err  error
}

func (o *object) add(key string, value any) {
	raw, err := json.Marshal(value)
	o.addRaw(key, raw, err)
}

func (o *object) addRaw(key string, raw []byte, err error) {
	if o.err != nil {
		return
	}
	if err != nil {
		o.err = err
		return
	}
	if len(o.keys) > 0 {
		o.buf.WriteByte(',')
	}
	k, _ := json.Marshal(key)
	o.buf.Write(k)
	o.buf.WriteByte(':')
	o.buf.Write(raw)
	o.keys = append(o.keys, key)
}

// merge adds the keys of data, which must marshal to an object or null.
func (o *object) merge(data any) error {
	if data == nil {
		return nil
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return fmt.Errorf("contract: data must be an object: %w", err)
	}
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		if slices.Contains(reserved, key) {
			return fmt.Errorf("contract: data uses the reserved key %q", key)
		}
		o.addRaw(key, fields[key], nil)
	}
	return nil
}

func (o *object) bytes() ([]byte, error) {
	if o.err != nil {
		return nil, o.err
	}
	return append(append([]byte{'{'}, o.buf.Bytes()...), '}'), nil
}
