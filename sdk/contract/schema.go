package contract

import (
	"errors"
	"maps"
	"slices"
	"strconv"
	"strings"
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
// $ref still apply, through allOf. A remote reference, $id, $anchor, or
// $dynamicRef, a reference that points nowhere, and a cycle are errors.
func ResolveSchema(schema map[string]any) (map[string]any, error) {
	resolved, err := inline(schema, schema, nil)
	if err != nil {
		return nil, err
	}
	return resolved.(map[string]any), nil
}

func inline(node any, root map[string]any, seen []string) (any, error) {
	switch n := node.(type) {
	case map[string]any:
		out := map[string]any{}
		for _, key := range []string{"$id", "$anchor", "$dynamicRef", "$dynamicAnchor"} {
			if _, found := n[key]; found {
				return nil, errors.New("schema: " + key + " is not supported")
			}
		}
		for key, value := range n {
			if key == "$ref" || key == "$defs" {
				continue
			}
			v, err := inline(value, root, seen)
			if err != nil {
				return nil, err
			}
			out[key] = v
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
	case []any:
		out := make([]any, len(n))
		for i, value := range n {
			v, err := inline(value, root, seen)
			if err != nil {
				return nil, err
			}
			out[i] = v
		}
		return out, nil
	}
	return node, nil
}

func list(value any) []any {
	items, _ := value.([]any)
	return items
}

// pointer follows a JSON pointer fragment, such as #/$defs/item, from root.
func pointer(root map[string]any, ref string) (any, bool) {
	var node any = root
	path := strings.TrimPrefix(ref, "#")
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
// so that AnswerSchema drops no constraint.
func IsDataSchema(schema map[string]any) bool {
	for key := range schema {
		if !slices.Contains(dataKeywords, key) {
			return false
		}
	}
	return schema["type"] == "object"
}

// AnswerSchema returns the JSON Schema of a whole answer of a result whose
// data keys data describes, or nil for a text result. The envelope adds ok,
// contract_version, errors, warnings, and, for a batch, results, whose items
// add ok and errors to the item that data's results describes. A success
// holds the keys that data requires. Unknown keys stay allowed, as
// Compatibility asks. A nil data describes no key.
func AnswerSchema(kind ResultKind, data map[string]any) map[string]any {
	if kind == ResultText {
		return nil
	}
	problems := map[string]any{"type": "array", "minItems": 1, "items": problemSchema()}
	properties := map[string]any{}
	maps.Copy(properties, table(data["properties"]))
	properties["ok"] = map[string]any{"type": "boolean"}
	properties["contract_version"] = map[string]any{"const": Version}
	properties["errors"] = problems
	properties["warnings"] = problems
	success := slices.Clone(list(data["required"]))
	if kind == ResultBatch {
		results := table(properties["results"])
		item := table(results["items"])
		itemProperties := map[string]any{}
		maps.Copy(itemProperties, table(item["properties"]))
		itemProperties["ok"] = map[string]any{"type": "boolean"}
		itemProperties["errors"] = problems
		itemSchema := map[string]any{}
		maps.Copy(itemSchema, item)
		itemSchema["type"] = "object"
		itemSchema["properties"] = itemProperties
		itemSchema["required"] = append([]any{"ok"}, list(item["required"])...)
		properties["results"] = map[string]any{"type": "array", "items": itemSchema}
		if !slices.Contains(success, any("results")) {
			success = append(success, "results")
		}
	}
	schema := map[string]any{
		"$schema":    SchemaDialect,
		"type":       "object",
		"required":   []any{"ok", "contract_version"},
		"properties": properties,
	}
	if len(success) > 0 {
		schema["if"] = map[string]any{"properties": map[string]any{"ok": map[string]any{"const": true}}}
		schema["then"] = map[string]any{"required": success}
	}
	return schema
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
