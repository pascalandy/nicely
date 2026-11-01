package extensiontest

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// validate checks an answer against a schema of draft 2020-12, with a
// loader that refuses every remote schema.
func validate(schema map[string]any, answer string) error {
	raw, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	doc, err := jsonschema.UnmarshalJSON(strings.NewReader(string(raw)))
	if err != nil {
		return err
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.UseLoader(noRemote{})
	if err := compiler.AddResource("answer.json", doc); err != nil {
		return err
	}
	compiled, err := compiler.Compile("answer.json")
	if err != nil {
		return err
	}
	value, err := jsonschema.UnmarshalJSON(strings.NewReader(answer))
	if err != nil {
		return fmt.Errorf("not JSON: %w", err)
	}
	return compiled.Validate(value)
}

// noRemote refuses to load any schema, so a reference outside the schema
// fails instead of reaching the network.
type noRemote struct{}

func (noRemote) Load(url string) (any, error) { return nil, fmt.Errorf("no remote schema: %s", url) }
