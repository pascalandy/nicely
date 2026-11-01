package contract

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

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
