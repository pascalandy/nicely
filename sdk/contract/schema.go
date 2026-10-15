package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// SchemaDialect is the JSON Schema draft of every schema that ncly
// describe returns.
const SchemaDialect = "https://json-schema.org/draft/2020-12/schema"

// envelopeKeys are the keys of every answer that the envelope owns, which
// keys leaves out, as ncly describe in core-spec.md defines.
var envelopeKeys = []string{"ok", "contract_version", "errors", "warnings"}

// ResolveSchema returns a copy of a schema whose references are resolved
// within it: each $ref, which must be local, such as #/$defs/item, gives
// way to the schema that it points to, and $defs goes. Keywords beside a
// $ref still apply, through allOf. Only the keywords that hold schemas are
// walked, so a literal value, such as a const that holds $defs, stays as
// it is. A $schema other than draft 2020-12 at the root, or any $schema
// below it, a schema that is not valid draft 2020-12, a remote reference,
// $id, $anchor, $dynamicRef, a reference that points nowhere, a cycle, and
// a root that resolves to no object are errors.
func ResolveSchema(schema map[string]any) (map[string]any, error) {
	if dialect, declared := schema["$schema"]; declared && dialect != SchemaDialect {
		return nil, errors.New("schema: only draft 2020-12 is supported")
	}
	if err := compileSchema(schema); err != nil {
		return nil, err
	}
	body := maps.Clone(schema)
	delete(body, "$schema")
	resolved, err := inline(body, schema, nil)
	if err != nil {
		return nil, err
	}
	root, isObject := resolved.(map[string]any)
	if !isObject {
		return nil, errors.New("schema: the root resolves to no object schema")
	}
	return root, nil
}

// compileSchema checks a schema against draft 2020-12, with a loader that
// refuses every remote schema.
func compileSchema(schema map[string]any) error {
	data, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return err
	}
	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	c.UseLoader(noRemote{})
	if err := c.AddResource("schema.json", doc); err != nil {
		return err
	}
	_, err = c.Compile("schema.json")
	return err
}

// noRemote refuses to load a schema, so that a reference outside the
// schema fails instead of reaching the network.
type noRemote struct{}

func (noRemote) Load(url string) (any, error) {
	return nil, errors.New("schema: no remote schema: " + url)
}

// Keywords of draft 2020-12 that hold schemas: one schema, a map of
// schemas, or a list of schemas. Every other keyword holds a literal value.
var (
	schemaKeywords = []string{
		"additionalProperties", "items", "contains", "propertyNames", "if", "then", "else",
		"not", "unevaluatedItems", "unevaluatedProperties", "contentSchema",
	}
	schemaMapKeywords  = []string{"properties", "patternProperties", "dependentSchemas"}
	schemaListKeywords = []string{"prefixItems", "allOf", "anyOf", "oneOf"}
)

// inline resolves the references of one schema, an object or a boolean.
func inline(node any, root map[string]any, seen []string) (any, error) {
	n, isObject := node.(map[string]any)
	if !isObject {
		return node, nil
	}
	for _, key := range []string{"$schema", "$id", "$anchor", "$dynamicRef", "$dynamicAnchor"} {
		if _, found := n[key]; found {
			return nil, errors.New("schema: " + key + " is not supported")
		}
	}
	out := map[string]any{}
	for key, value := range n {
		var err error
		switch {
		case key == "$ref" || key == "$defs":
			continue
		case slices.Contains(schemaKeywords, key):
			value, err = inline(value, root, seen)
		case slices.Contains(schemaMapKeywords, key):
			value, err = inlineEach(table(value), root, seen)
		case slices.Contains(schemaListKeywords, key):
			value, err = inlineList(list(value), root, seen)
		}
		if err != nil {
			return nil, err
		}
		out[key] = value
	}
	raw, isRef := n["$ref"]
	if !isRef {
		return out, nil
	}
	ref, isString := raw.(string)
	if !isString || !strings.HasPrefix(ref, "#") {
		return nil, errors.New("schema: only a local $ref is supported")
	}
	if slices.Contains(seen, ref) {
		return nil, errors.New("schema: $ref " + ref + " refers to itself")
	}
	target, found := pointer(root, ref)
	if !found {
		return nil, errors.New("schema: $ref " + ref + " points nowhere")
	}
	resolved, err := inline(target, root, append(slices.Clone(seen), ref))
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return resolved, nil
	}
	out["allOf"] = append([]any{resolved}, list(out["allOf"])...)
	return out, nil
}

func inlineEach(schemas map[string]any, root map[string]any, seen []string) (any, error) {
	out := map[string]any{}
	for name, schema := range schemas {
		resolved, err := inline(schema, root, seen)
		if err != nil {
			return nil, err
		}
		out[name] = resolved
	}
	return out, nil
}

func inlineList(schemas []any, root map[string]any, seen []string) (any, error) {
	out := make([]any, len(schemas))
	for i, schema := range schemas {
		resolved, err := inline(schema, root, seen)
		if err != nil {
			return nil, err
		}
		out[i] = resolved
	}
	return out, nil
}

func list(value any) []any {
	items, _ := value.([]any)
	return items
}

// pointer follows a JSON pointer fragment, such as #/$defs/item, from root.
// The fragment is percent-decoded first, as a URI fragment is.
func pointer(root map[string]any, ref string) (any, bool) {
	var node any = root
	path, err := url.PathUnescape(strings.TrimPrefix(ref, "#"))
	if err != nil {
		return nil, false
	}
	if path == "" {
		return node, true
	}
	if !strings.HasPrefix(path, "/") {
		return nil, false
	}
	for _, token := range strings.Split(path[1:], "/") {
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		switch n := node.(type) {
		case map[string]any:
			next, found := n[token]
			if !found {
				return nil, false
			}
			node = next
		case []any:
			i, err := strconv.Atoi(token)
			if err != nil || i < 0 || i >= len(n) {
				return nil, false
			}
			node = n[i]
		default:
			return nil, false
		}
	}
	return node, true
}

// problemSchema is the schema of one entry of errors or warnings.
func problemSchema() map[string]any {
	str := map[string]any{"type": "string"}
	return map[string]any{
		"type":     "object",
		"required": []any{"code", "message", "hint"},
		"properties": map[string]any{
			"code":    map[string]any{"type": "string", "pattern": "^[A-Z0-9_]+$"},
			"message": str,
			"hint":    str,
			"next":    map[string]any{"type": "array", "minItems": 1, "items": str},
		},
	}
}

// dataKeywords are the keywords of the top level of a data schema, which
// AnswerSchema carries into the schema of the answer.
var dataKeywords = []string{"$schema", "title", "description", "type", "properties", "required"}

// IsDataSchema reports whether a resolved schema can describe the data keys
// of an answer: an object schema whose top level holds only dataKeywords,
// so that AnswerSchema drops no constraint. The item of a batch's results,
// when declared, follows the same rule.
func IsDataSchema(schema map[string]any) bool {
	for key := range schema {
		if !slices.Contains(dataKeywords, key) {
			return false
		}
	}
	results, declared := table(schema["properties"])["results"]
	if declared {
		r, isObject := results.(map[string]any)
		item, hasItem := r["items"]
		if !isObject || hasItem && (table(item) == nil || !IsDataSchema(table(item))) {
			return false
		}
	}
	return schema["type"] == "object"
}

// AnswerSchema returns the JSON Schema of a whole answer of a result whose
// data keys data describes, or nil for a text result. The envelope adds ok,
// contract_version, errors, warnings, and, for a batch, results, whose items
// add ok and errors to the item that data's results describes. A success
// holds the keys that data requires and no error, nor a failed item for a
// batch, and a failure other than a report's holds at least one error.
// Unknown keys stay allowed, as Compatibility asks. A nil data describes no
// key.
func AnswerSchema(kind ResultKind, data map[string]any) map[string]any {
	if kind == ResultText {
		return nil
	}
	problems := map[string]any{"type": "array", "items": problemSchema()}
	properties := map[string]any{}
	maps.Copy(properties, table(data["properties"]))
	properties["ok"] = map[string]any{"type": "boolean"}
	properties["contract_version"] = map[string]any{"const": Version}
	properties["errors"] = problems
	properties["warnings"] = problems
	required := slices.Clone(list(data["required"]))
	if kind == ResultBatch {
		properties["results"] = resultsSchema(table(properties["results"]), problems)
		if !slices.Contains(required, any("results")) {
			required = append(required, "results")
		}
	}
	schema := map[string]any{
		"$schema":    SchemaDialect,
		"type":       "object",
		"required":   []any{"ok", "contract_version"},
		"properties": properties,
	}
	success := verdict(schema, required, kind != ResultReport)
	if kind == ResultBatch {
		success["results"] = map[string]any{"items": map[string]any{"properties": map[string]any{"ok": map[string]any{"const": true}}}}
	}
	return schema
}

// verdict ties a schema to ok, as Answer check does: a success holds the
// required keys and no error, and, when failureErrors is set, a failure
// holds at least one error. A report's failure may list its findings
// instead. It returns the properties that a success fits, so that a batch
// can add that none of its items failed.
func verdict(schema map[string]any, required []any, failureErrors bool) map[string]any {
	properties := map[string]any{"errors": map[string]any{"maxItems": 0}}
	success := map[string]any{"properties": properties}
	if len(required) > 0 {
		success["required"] = required
	}
	schema["if"] = map[string]any{"properties": map[string]any{"ok": map[string]any{"const": true}}}
	schema["then"] = success
	if failureErrors {
		schema["else"] = map[string]any{"required": []any{"errors"}, "properties": map[string]any{"errors": map[string]any{"minItems": 1}}}
	}
	return properties
}

// resultsSchema wraps the schema of a batch's results, keeping its own
// constraints, such as minItems. Each item adds ok and errors to the item
// that the declaration describes, and holds its required keys when it
// succeeded and at least one error when it failed.
func resultsSchema(declared map[string]any, problems map[string]any) map[string]any {
	item := table(declared["items"])
	properties := map[string]any{}
	maps.Copy(properties, table(item["properties"]))
	properties["ok"] = map[string]any{"type": "boolean"}
	properties["errors"] = problems
	itemSchema := map[string]any{"type": "object", "required": []any{"ok"}, "properties": properties}
	verdict(itemSchema, list(item["required"]), true)
	results := map[string]any{}
	maps.Copy(results, declared)
	results["type"] = "array"
	results["items"] = itemSchema
	return results
}

// Keys returns the top-level data keys of a successful answer that data
// describes, sorted: its properties without the keys of the envelope, and
// results for a batch.
func Keys(kind ResultKind, data map[string]any) []string {
	keys := []string{}
	if kind == ResultText {
		return keys
	}
	for key := range table(data["properties"]) {
		if !slices.Contains(envelopeKeys, key) {
			keys = append(keys, key)
		}
	}
	if kind == ResultBatch && !slices.Contains(keys, "results") {
		keys = append(keys, "results")
	}
	slices.Sort(keys)
	return keys
}

func table(value any) map[string]any {
	m, _ := value.(map[string]any)
	return m
}
