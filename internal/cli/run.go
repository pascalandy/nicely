package cli

import (
	"cmp"
	_ "embed"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/record"
	"github.com/spf13/cobra"
)

//go:embed run_list_schema.json
var runListSchema []byte

//go:embed run_view_schema.json
var runViewSchema []byte

var runCommand = contract.Command{
	Path:    []string{"run"},
	Summary: "run.summary",
	Effects: &contract.Effects{},
	Examples: []string{
		`ncly run list`,
		`ncly run list --path 'agent run' --json`,
		`ncly run view review-1`,
	},
}

var runListCommand = contract.Command{
	Path:    []string{"run", "list"},
	Result:  contract.Result{Schema: mustSchema(runListSchema)},
	Summary: "run.list.summary",
	Flags: []contract.Flag{
		{Name: "path", Value: "command-path", Summary: "run.list.path"},
		{Name: "key", Value: "item-key", Summary: "run.list.key"},
		{Name: "since", Value: "time", Summary: "run.list.since"},
		{Name: "limit", Value: "count", Type: contract.TypeInteger, Summary: "run.list.limit", Default: int64(defaultRunLimit)},
	},
	Effects: &contract.Effects{},
	Examples: []string{
		`ncly run list`,
		`ncly run list --path 'transcript run youtube' --key <url> --json`,
		`ncly run list --since 2026-10-07T00:00:00Z --limit 50`,
	},
}

var runViewCommand = contract.Command{
	Path:    []string{"run", "view"},
	Result:  contract.Result{Schema: mustSchema(runViewSchema)},
	Summary: "run.view.summary",
	Args:    []contract.Arg{{Name: "run-id", Summary: "run.view.run_id", Required: true}},
	Effects: &contract.Effects{},
	Examples: []string{
		`ncly run view review-1`,
		`ncly run view 20261007-214501-3f9a2c --json`,
	},
}

// defaultRunLimit bounds a list without --limit, as ncly run in
// core-spec.md defines.
const defaultRunLimit = 20

// keyPreview is how many item keys a list entry shows.
const keyPreview = 3

// runList is the answer of ncly run list.
type runList struct {
	Runs       []runEntry `json:"runs"`
	More       bool       `json:"more"`
	Unreadable []string   `json:"unreadable"`
}

// runEntry is one run of a list, as ncly run in core-spec.md defines it.
type runEntry struct {
	RunID     string   `json:"run_id"`
	Path      string   `json:"path"`
	Status    string   `json:"status"`
	StartedAt string   `json:"started_at"`
	UpdatedAt string   `json:"updated_at"`
	Active    bool     `json:"active"`
	ItemCount int      `json:"item_count"`
	Keys      []string `json:"keys"`
	// started is StartedAt as a time, which orders the list.
	started time.Time
}

// runFilter holds the filters of a list, which all apply before the limit.
type runFilter struct {
	path, key string
	since     *time.Time
}

// keep reports whether a run passes every filter.
func (f runFilter) keep(rec *record.Record, started time.Time) bool {
	if f.path != "" && rec.Path != f.path {
		return false
	}
	if f.key != "" && !slices.ContainsFunc(rec.Items, func(it record.Item) bool { return it.Key == f.key }) {
		return false
	}
	return f.since == nil || !started.Before(*f.since)
}

// runs returns the runs folder of the state folder.
func runs() record.Store {
	return record.Store{Dir: record.Folder(config.Locate(home(), os.LookupEnv).State)}
}

// addRun adds ncly run, whose commands read the records of the state folder
// directly, with no index, and start no work.
func addRun(p *i18n.Printer, root *cobra.Command) {
	group := command.Build(p, p, runCommand, nil)
	group.AddCommand(command.Build(p, p, runListCommand, func(c *cobra.Command, _ []string) error {
		f := runFilter{}
		f.path, _ = c.Flags().GetString("path")
		f.key, _ = c.Flags().GetString("key")
		if value, _ := c.Flags().GetString("since"); c.Flags().Changed("since") {
			since, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return command.Fail(command.HelpUsage(c, p.T("run.list.since_invalid", map[string]any{"Value": i18n.Clip(value)})))
			}
			f.since = &since
		}
		limit := defaultRunLimit
		if value, _ := c.Flags().GetString("limit"); c.Flags().Changed("limit") {
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				return command.Fail(command.HelpUsage(c, p.T("run.list.limit_invalid", map[string]any{"Value": i18n.Clip(value)})))
			}
			limit = n
		}
		store := runs()
		list, err := listRuns(store, f, limit)
		if err != nil {
			return command.Fail(contract.Problem{Code: contract.Runtime, Message: p.T("run.list.folder_unreadable", map[string]any{"Dir": store.Dir}), Hint: p.T("run.list.folder_unreadable_hint", map[string]any{"Dir": store.Dir})})
		}
		return &answered{
			outcome: contract.Outcome{Data: list},
			text:    func(w io.Writer) error { return writeRuns(w, p, list) },
		}
	}))
	group.AddCommand(command.Build(p, p, runViewCommand, func(c *cobra.Command, args []string) error {
		id := args[0]
		if !record.ValidID(id) {
			return command.Fail(command.HelpUsage(c, p.T("run.view.id_invalid", map[string]any{"ID": i18n.Clip(id)})))
		}
		store := runs()
		rec, keys, err := store.View(id)
		if err != nil {
			return command.Fail(contract.Problem{Code: contract.ResumeUnsafe, Message: p.T("run.view.unreadable", map[string]any{"File": store.File(id)}), Hint: p.T("run.view.unreadable_hint")})
		}
		if rec == nil {
			list := []string{"ncly", "run", "list"}
			return command.Fail(contract.Problem{Code: contract.NotFound, Message: p.T("run.view.not_found", map[string]any{"ID": id}), Hint: strings.Join(list, " "), Next: list})
		}
		active, err := store.Active(id)
		if err != nil {
			// A lock that cannot be read leaves active unknown, which never
			// reads as false.
			data := map[string]any{"File": store.Lock(id), "ID": id}
			return command.Fail(contract.Problem{Code: contract.ResumeUnsafe, Message: p.T("run.view.lock_unreadable", data), Hint: p.T("run.view.unreadable_hint")})
		}
		keys["active"] = active
		return &answered{
			outcome: contract.Outcome{Data: map[string]any{"run": keys}},
			text:    func(w io.Writer) error { return writeRun(w, p, id, rec, keys, active) },
		}
	}))
	root.AddCommand(group)
}

// listRuns reads every record of the folder, keeps the runs that pass the
// filters, newest first, then by run ID, and bounds them by the limit. A
// record that it cannot use goes to unreadable, whatever the filters.
func listRuns(store record.Store, f runFilter, limit int) (runList, error) {
	ids, err := store.IDs()
	if err != nil {
		return runList{}, err
	}
	list := runList{Runs: []runEntry{}, Unreadable: []string{}}
	for _, id := range ids {
		rec, err := store.Read(id)
		if rec == nil || err != nil {
			// The record vanished, or cannot be used: never an absent run.
			if err != nil {
				list.Unreadable = append(list.Unreadable, id)
			}
			continue
		}
		started, err := time.Parse(time.RFC3339, rec.StartedAt)
		// A lock that cannot be read leaves active unknown, which never
		// reads as false.
		active, errActive := store.Active(id)
		if err != nil || errActive != nil {
			list.Unreadable = append(list.Unreadable, id)
			continue
		}
		if !f.keep(rec, started) {
			continue
		}
		e := runEntry{RunID: id, Path: rec.Path, Status: string(rec.Status), StartedAt: rec.StartedAt, UpdatedAt: rec.UpdatedAt, Active: active, ItemCount: len(rec.Items), Keys: []string{}, started: started}
		for _, it := range rec.Items[:min(keyPreview, len(rec.Items))] {
			e.Keys = append(e.Keys, it.Key)
		}
		list.Runs = append(list.Runs, e)
	}
	slices.SortFunc(list.Runs, func(a, b runEntry) int {
		return cmp.Or(b.started.Compare(a.started), strings.Compare(a.RunID, b.RunID))
	})
	if len(list.Runs) > limit {
		list.Runs, list.More = list.Runs[:limit], true
	}
	return list, nil
}
