package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// expect is what the two validators of an answer say of a case of the
// corpus: CheckAnswer, which core runs on the reply of an extension, and
// the schema that AnswerSchema gives ncly describe.
type expect int

const (
	// bothAccept and bothReject state a rule that both validators hold.
	bothAccept expect = iota
	bothReject
	// checkRejects states a rule of the answer check alone, which the
	// schema cannot hold, and schemaRejects one of the schema alone.
	checkRejects
	schemaRejects
)

func (e expect) String() string {
	return [...]string{"both accept", "both reject", "CheckAnswer alone rejects", "AnswerSchema alone rejects"}[e]
}

// corpusData holds the data schemas of the corpus, by kind: an object that
// requires greeting, a batch whose items require url, and a report that
// lists checks.
var corpusData = map[ResultKind]map[string]any{
	ResultObject: {"type": "object", "properties": map[string]any{"greeting": map[string]any{"type": "string"}}, "required": []any{"greeting"}},
	ResultBatch: {"type": "object", "properties": map[string]any{"results": map[string]any{"type": "array", "items": map[string]any{
		"type": "object", "properties": map[string]any{"url": map[string]any{"type": "string"}}, "required": []any{"url"},
	}}}},
	ResultReport: {"type": "object", "properties": map[string]any{"checks": map[string]any{"type": "array"}}},
}

// corpusManifest declares one code of its own, HELLO_GRUMPY, with exit 1.
var corpusManifest = Manifest{Domain: "hello", Errors: []OwnCode{{"HELLO_GRUMPY", ExitRuntime}}}

// TestTheAnswerCheckAndAnswerSchemaAgree passes each answer of one corpus to
// both validators. Each case names its rule and the verdicts that it
// expects. The cases of a common rule come first. The schema sees no exit
// code and no stream, so each case gives the exit code and the stream that
// its answer needs. The cases after them name the rules that only one
// validator holds, so the corpus never counts them as common. A case that
// the answer check rejects also names the ID of the rule that it breaks,
// so a rejection for another reason fails the case.
func TestTheAnswerCheckAndAnswerSchemaAgree(t *testing.T) {
	cases := []struct {
		name, rule string
		kind       ResultKind
		exit       int
		stderr     bool
		answer     string
		want       expect
		checkRule  string
	}{
		// A success
		{"a success", "a success holds its data and no error", ResultObject, 0, false, `{"ok":true,"contract_version":1,"greeting":"hi"}`, bothAccept, ""},
		{"a success with an error", "a success holds no error", ResultObject, 0, false, `{"ok":true,"contract_version":1,"greeting":"hi","errors":[{"code":"RUNTIME","message":"m","hint":"h"}]}`, bothReject, "extension.answer_errors_stdout"},
		{"a success with a warning", "warnings may appear in any answer", ResultObject, 0, false, `{"ok":true,"contract_version":1,"greeting":"hi","warnings":[{"code":"CONFIG_UNKNOWN_KEY","message":"m","hint":"h"}]}`, bothAccept, ""},
		{"a success with an unknown key", "an answer may hold keys that its declaration does not name", ResultObject, 0, false, `{"ok":true,"contract_version":1,"greeting":"hi","extra":1}`, bothAccept, ""},
		{"ok as a string", "ok is a boolean", ResultObject, 0, false, `{"ok":"true","contract_version":1,"greeting":"hi"}`, bothReject, "extension.answer_ok"},
		{"no ok", "an answer holds ok", ResultObject, 0, false, `{"contract_version":1,"greeting":"hi"}`, bothReject, "extension.answer_ok"},
		{"no contract_version", "an answer holds contract_version", ResultObject, 0, false, `{"ok":true,"greeting":"hi"}`, bothReject, "extension.answer_version"},
		{"a contract_version of 2", "contract_version is 1", ResultObject, 0, false, `{"ok":true,"contract_version":2,"greeting":"hi"}`, bothReject, "extension.answer_version"},
		{"an array", "an answer is one JSON object", ResultObject, 0, false, `[{"ok":true,"contract_version":1,"greeting":"hi"}]`, bothReject, "extension.answer_json"},

		// A failure
		{"a failure", "a failure holds an error", ResultObject, 1, true, `{"ok":false,"contract_version":1,"errors":[{"code":"RUNTIME","message":"m","hint":"h"}]}`, bothAccept, ""},
		{"a failure without errors", "a failure holds at least one error", ResultObject, 1, true, `{"ok":false,"contract_version":1}`, bothReject, "extension.answer_errors_missing"},
		{"a failure with a next", "next is a non-empty array of strings", ResultObject, 78, true, `{"ok":false,"contract_version":1,"errors":[{"code":"AUTH_MISSING","message":"m","hint":"ncly auth login deepgram","next":["ncly","auth","login","deepgram"]}]}`, bothAccept, ""},
		{"a problem without a hint", "a problem holds the strings code, message, and hint", ResultObject, 1, true, `{"ok":false,"contract_version":1,"errors":[{"code":"RUNTIME","message":"m"}]}`, bothReject, "extension.answer_problem"},
		{"a problem whose hint is null", "a problem holds the strings code, message, and hint", ResultObject, 1, true, `{"ok":false,"contract_version":1,"errors":[{"code":"RUNTIME","message":"m","hint":null}]}`, bothReject, "extension.answer_problem"},
		{"a problem with an empty next", "next is a non-empty array of strings", ResultObject, 1, true, `{"ok":false,"contract_version":1,"errors":[{"code":"RUNTIME","message":"m","hint":"h","next":[]}]}`, bothReject, "extension.answer_next"},
		{"an error code in lowercase", "a code is uppercase letters, digits, and underscores", ResultObject, 1, true, `{"ok":false,"contract_version":1,"errors":[{"code":"runtime","message":"m","hint":"h"}]}`, bothReject, "extension.answer_code"},
		{"a warning code in lowercase", "a code is uppercase letters, digits, and underscores", ResultObject, 0, false, `{"ok":true,"contract_version":1,"greeting":"hi","warnings":[{"code":"config_unknown_key","message":"m","hint":"h"}]}`, bothReject, "extension.answer_code"},
		{"a warning without a message", "a problem holds the strings code, message, and hint", ResultObject, 0, false, `{"ok":true,"contract_version":1,"greeting":"hi","warnings":[{"code":"CONFIG_UNKNOWN_KEY","hint":"h"}]}`, bothReject, "extension.answer_problem"},

		// Empty arrays
		{"a success with an empty errors and warnings", "an empty errors or warnings holds no problem", ResultObject, 0, false, `{"ok":true,"contract_version":1,"greeting":"hi","errors":[],"warnings":[]}`, bothAccept, ""},
		{"a failure with an empty errors", "a failure holds at least one error", ResultObject, 1, true, `{"ok":false,"contract_version":1,"errors":[]}`, bothReject, "extension.answer_errors_missing"},
		{"errors as null", "errors, when present, is an array", ResultObject, 0, false, `{"ok":true,"contract_version":1,"greeting":"hi","errors":null}`, bothReject, "extension.answer_problem"},
		{"warnings as null", "warnings, when present, is an array", ResultObject, 0, false, `{"ok":true,"contract_version":1,"greeting":"hi","warnings":null}`, bothReject, "extension.answer_problem"},

		// A report
		{"a report that passes", "a report lists its findings under its own key", ResultReport, 0, false, `{"ok":true,"contract_version":1,"checks":[]}`, bothAccept, ""},
		{"a failed report", "a failed report lists its findings instead of errors", ResultReport, 78, false, `{"ok":false,"contract_version":1,"checks":[{"id":"core.config","status":"fail"}]}`, bothAccept, ""},
		{"the failure of a report command", "a report command that fails itself answers with errors", ResultReport, 78, true, `{"ok":false,"contract_version":1,"errors":[{"code":"CONFIG_INVALID","message":"m","hint":"h"}]}`, bothAccept, ""},

		// A batch
		{"a batch", "each item holds the boolean ok and its data", ResultBatch, 0, false, `{"ok":true,"contract_version":1,"results":[{"ok":true,"url":"u1"},{"ok":true,"url":"u2"}]}`, bothAccept, ""},
		{"a successful batch with a failed item", "a successful batch holds no failed item", ResultBatch, 0, false, `{"ok":true,"contract_version":1,"results":[{"ok":true,"url":"u1"},{"ok":false,"errors":[{"code":"RUNTIME","message":"m","hint":"h"}]}]}`, bothReject, "extension.answer_item_failed"},
		{"results as an object", "results is an array of objects", ResultBatch, 0, false, `{"ok":true,"contract_version":1,"results":{"ok":true,"url":"u1"}}`, bothReject, "extension.answer_results"},
		{"an item without ok", "each item holds the boolean ok", ResultBatch, 0, false, `{"ok":true,"contract_version":1,"results":[{"url":"u1"}]}`, bothReject, "extension.answer_results"},

		// A partial batch
		{"a partial batch", "a failed item holds an error and none of the keys that its declaration requires", ResultBatch, 1, true, `{"ok":false,"contract_version":1,"results":[{"ok":true,"url":"u1"},{"ok":false,"errors":[{"code":"RUNTIME","message":"m","hint":"h"}]}],"errors":[{"code":"RUNTIME","message":"m","hint":"h"}]}`, bothAccept, ""},
		{"a failed item without errors", "a failed item holds at least one error", ResultBatch, 1, true, `{"ok":false,"contract_version":1,"results":[{"ok":true,"url":"u1"},{"ok":false}],"errors":[{"code":"RUNTIME","message":"m","hint":"h"}]}`, bothReject, "extension.answer_item_ok"},
		{"a failed item with an empty errors", "a failed item holds at least one error", ResultBatch, 1, true, `{"ok":false,"contract_version":1,"results":[{"ok":true,"url":"u1"},{"ok":false,"errors":[]}],"errors":[{"code":"RUNTIME","message":"m","hint":"h"}]}`, bothReject, "extension.answer_item_ok"},
		{"a successful item with an error", "a successful item holds no error", ResultBatch, 1, true, `{"ok":false,"contract_version":1,"results":[{"ok":true,"url":"u1","errors":[{"code":"RUNTIME","message":"m","hint":"h"}]},{"ok":false,"errors":[{"code":"RUNTIME","message":"m","hint":"h"}]}],"errors":[{"code":"RUNTIME","message":"m","hint":"h"}]}`, bothReject, "extension.answer_item_ok"},

		// The rules of the answer check alone: the exit code, the streams,
		// the codes of the manifest, and the text of the line, which the
		// schema never sees
		{"an unknown exit code", "the exit code is one of Exit codes", ResultObject, 3, true, `{"ok":false,"contract_version":1,"errors":[{"code":"RUNTIME","message":"m","hint":"h"}]}`, checkRejects, "extension.answer_exit"},
		{"an answer on two lines", "the answer is one line", ResultObject, 0, false, "{\"ok\":true,\n\"contract_version\":1,\"greeting\":\"hi\"}", checkRejects, "extension.answer_lines"},
		{"a contract_version written 1.0", "contract_version is the integer 1 as the line writes it, while the schema compares numbers", ResultObject, 0, false, `{"ok":true,"contract_version":1.0,"greeting":"hi"}`, checkRejects, "extension.answer_version"},
		{"errors on stdout in a failed report", "an answer on stdout holds no error", ResultReport, 78, false, `{"ok":false,"contract_version":1,"checks":[],"errors":[{"code":"CONFIG_INVALID","message":"m","hint":"h"}]}`, checkRejects, "extension.answer_errors_stdout"},
		{"the failure of a report command without errors", "an answer on stderr holds an error, while the schema lets a failed report list its findings", ResultReport, 78, true, `{"ok":false,"contract_version":1}`, checkRejects, "extension.answer_errors_missing"},
		{"a code that maps to another exit code", "a code that core knows maps to the exit code", ResultObject, 2, true, `{"ok":false,"contract_version":1,"errors":[{"code":"RUNTIME","message":"m","hint":"h"}]}`, checkRejects, "extension.answer_code_exit"},
		{"a declared code of the domain", "a code of the domain is declared in the manifest", ResultObject, 1, true, `{"ok":false,"contract_version":1,"errors":[{"code":"HELLO_GRUMPY","message":"m","hint":"h"}]}`, bothAccept, ""},
		{"an undeclared code of the domain", "a code of the domain is declared in the manifest", ResultObject, 1, true, `{"ok":false,"contract_version":1,"errors":[{"code":"HELLO_SLEEPY","message":"m","hint":"h"}]}`, checkRejects, "extension.answer_undeclared"},
		{"a failed item with exit 78", "a failed item makes the exit code that of a batch", ResultBatch, 78, true, `{"ok":false,"contract_version":1,"results":[{"ok":false,"errors":[{"code":"AUTH_MISSING","message":"m","hint":"h"}]}],"errors":[{"code":"AUTH_MISSING","message":"m","hint":"h"}]}`, checkRejects, "extension.answer_batch"},
		{"results with a failed item in a successful object", "results, in an answer of any kind, holds no failed item on a success, while the schema describes results only for a batch", ResultObject, 0, false, `{"ok":true,"contract_version":1,"greeting":"hi","results":[{"ok":false,"errors":[{"code":"RUNTIME","message":"m","hint":"h"}]}]}`, checkRejects, "extension.answer_item_failed"},

		// The rules of the schema alone: the data keys that a declaration
		// describes, which the answer check never reads
		{"a success without its declared key", "a success holds the keys that its declaration requires", ResultObject, 0, false, `{"ok":true,"contract_version":1}`, schemaRejects, ""},
		{"a declared key of the wrong type", "a data key has the type that its declaration gives", ResultObject, 0, false, `{"ok":true,"contract_version":1,"greeting":1}`, schemaRejects, ""},
		{"a successful batch without results", "a successful batch holds results, its declared key", ResultBatch, 0, false, `{"ok":true,"contract_version":1}`, schemaRejects, ""},
		{"a successful item without its declared key", "a successful item holds the keys that its declaration requires", ResultBatch, 0, false, `{"ok":true,"contract_version":1,"results":[{"ok":true}]}`, schemaRejects, ""},
	}
	schemas := map[ResultKind]*jsonschema.Schema{}
	for kind, data := range corpusData {
		schemas[kind] = compileAnswerSchema(t, kind, data)
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := Reply{Exit: c.exit, Stdout: []byte(c.answer + "\n")}
			if c.stderr {
				reply = Reply{Exit: c.exit, Last: []byte(c.answer + "\n")}
			}
			_, checkErr := CheckAnswer(corpusManifest, c.kind, reply)
			var broke *RuleError
			if checkErr != nil && (!errors.As(checkErr, &broke) || broke.Rule != c.checkRule) {
				t.Errorf("%s: CheckAnswer gives %v, want the rule %q", c.name, says(checkErr), c.checkRule)
			}
			answer, err := jsonschema.UnmarshalJSON(strings.NewReader(c.answer))
			if err != nil {
				t.Fatalf("the answer of %q is no JSON: %v", c.name, err)
			}
			schemaErr := schemas[c.kind].Validate(answer)
			got := map[[2]bool]expect{{true, true}: bothAccept, {false, false}: bothReject, {false, true}: checkRejects, {true, false}: schemaRejects}[[2]bool{checkErr == nil, schemaErr == nil}]
			if got != c.want {
				t.Errorf("%s, for the rule %q: CheckAnswer gives %v and AnswerSchema %v, so %v, want %v", c.name, c.rule, says(checkErr), says(schemaErr), got, c.want)
			}
		})
	}
}

// says tells a validator's error as the verdict that it gives.
func says(err error) string {
	if err == nil {
		return "accept"
	}
	return "reject (" + err.Error() + ")"
}

// compileAnswerSchema compiles the schema of a whole answer as ncly describe
// gives it, with the loader of the package, which refuses every remote
// schema.
func compileAnswerSchema(t *testing.T, kind ResultKind, data map[string]any) *jsonschema.Schema {
	t.Helper()
	raw, err := json.Marshal(AnswerSchema(kind, data))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	c.UseLoader(noRemote{})
	if err := c.AddResource("answer.json", doc); err != nil {
		t.Fatal(err)
	}
	schema, err := c.Compile("answer.json")
	if err != nil {
		t.Fatal(err)
	}
	return schema
}
