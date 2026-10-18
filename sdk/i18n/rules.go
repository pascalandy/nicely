package i18n

import (
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"text/template/parse"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

// CheckCatalogs applies the catalog rules of SDK in contract.md to the
// catalogs at the top of fsys: every entry has a description, an entry
// whose other form reads a number field has a one form, and every language
// holds exactly the IDs of en. A non-empty domain, an extension's, also
// requires every ID to start with the domain and a dot.
func CheckCatalogs(fsys fs.FS, domain string) error {
	en, err := readCatalog(fsys, "en.toml")
	if err != nil {
		return err
	}
	files, err := fs.Glob(fsys, "*.toml")
	if err != nil {
		return err
	}
	for _, file := range files {
		catalog, err := readCatalog(fsys, file)
		if err != nil {
			return err
		}
		for _, id := range slices.Sorted(maps.Keys(catalog)) {
			if description, _ := catalog[id]["description"].(string); strings.TrimSpace(description) == "" {
				return fmt.Errorf("%s: %s has no description", file, id)
			}
			if domain != "" && !strings.HasPrefix(id, domain+".") {
				return fmt.Errorf("%s: %s does not start with %q", file, id, domain+".")
			}
			if _, found := en[id]; !found {
				return fmt.Errorf("%s holds %s, which en.toml lacks", file, id)
			}
			if err := checkOneForm(catalog[id]); err != nil {
				return fmt.Errorf("%s: %s %w", file, id, err)
			}
		}
		for _, id := range slices.Sorted(maps.Keys(en)) {
			if _, found := catalog[id]; !found {
				return fmt.Errorf("%s lacks %s, which en.toml holds", file, id)
			}
		}
	}
	return nil
}

// numberFields holds the template fields that carry a number: Count, which
// Printer.N fills, and N, which Printer.Size and Printer.Duration fill.
var numberFields = []string{"Count", "N"}

// checkOneForm requires a one form of an entry whose other form reads a
// number field, such as {{.Count}} or {{ .N }}. When English writes both
// forms the same, as for a unit symbol, neither a scenario nor the
// pseudo-locale sees a missing one, and a language whose forms differ would
// show the other form for 1.
func checkOneForm(entry map[string]any) error {
	m, err := goi18n.NewMessage(entry)
	if err != nil {
		return fmt.Errorf("is not a message: %w", err)
	}
	tree := parse.New("other")
	tree.Mode = parse.SkipFuncCheck
	if _, err := tree.Parse(m.Other, m.LeftDelim, m.RightDelim, map[string]*parse.Tree{}); err != nil {
		return fmt.Errorf("has an other form that does not parse: %w", err)
	}
	if field := numberField(tree.Root); field != "" && strings.TrimSpace(m.One) == "" {
		return fmt.Errorf("reads .%s in its other form, but has no one form", field)
	}
	return nil
}

// numberField returns the first field of numberFields that node reads, or
// "" when it reads none.
func numberField(node parse.Node) string {
	var nodes []parse.Node
	switch n := node.(type) {
	case *parse.FieldNode:
		if slices.Contains(numberFields, n.Ident[0]) {
			return n.Ident[0]
		}
	case *parse.ListNode:
		if n != nil {
			nodes = n.Nodes
		}
	case *parse.ActionNode:
		nodes = []parse.Node{n.Pipe}
	case *parse.PipeNode:
		for _, c := range n.Cmds {
			nodes = append(nodes, c)
		}
	case *parse.CommandNode:
		nodes = n.Args
	case *parse.IfNode:
		nodes = []parse.Node{n.Pipe, n.List, n.ElseList}
	case *parse.RangeNode:
		nodes = []parse.Node{n.Pipe, n.List, n.ElseList}
	case *parse.WithNode:
		nodes = []parse.Node{n.Pipe, n.List, n.ElseList}
	}
	for _, n := range nodes {
		if field := numberField(n); field != "" {
			return field
		}
	}
	return ""
}

func readCatalog(fsys fs.FS, file string) (map[string]map[string]any, error) {
	data, err := fs.ReadFile(fsys, file)
	if err != nil {
		return nil, err
	}
	var catalog map[string]map[string]any
	if err := toml.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	return catalog, nil
}
