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

	"github.com/charmbracelet/x/ansi"
	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/config"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/record"
	"github.com/pascalandy/nicely/sdk/tui"
	"github.com/spf13/cobra"
)

//go:embed run_list_schema.json
var runListSchema []byte

var runCommand = contract.Command{
	Path:    []string{"run"},
	Summary: "run.summary",
	Effects: &contract.Effects{},
	Examples: []string{
		`ncly run list`,
		`ncly run list --path 'agent run' --json`,
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
				return command.Fail(command.HelpUsage(c, p.T("run.list.since_invalid", map[string]any{"Value": value})))
			}
			f.since = &since
		}
		limit := defaultRunLimit
		if value, _ := c.Flags().GetString("limit"); c.Flags().Changed("limit") {
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				return command.Fail(command.HelpUsage(c, p.T("run.list.limit_invalid", map[string]any{"Value": value})))
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

// writeRuns prints a list as ncly run in core-spec.md shows it: a bold
// header, one row per run, its previewed keys under it, then the legend of
// an active run, the records that it could not read, and the hint of more.
func writeRuns(w io.Writer, p *i18n.Printer, list runList) error {
	var b strings.Builder
	if len(list.Runs) == 0 {
		b.WriteString(p.T("run.list.none") + "\n")
	} else {
		header := []string{p.T("run.list.column_run"), p.T("run.list.column_status"), p.T("run.list.column_started"), p.T("run.list.column_command"), p.T("run.list.column_items")}
		rows := [][]string{}
		for _, e := range list.Runs {
			rows = append(rows, []string{e.RunID, e.Status, p.Date(e.started.UTC()), e.Path, p.Number(float64(e.ItemCount), 0)})
		}
		widths := make([]int, len(header))
		for _, row := range append([][]string{header}, rows...) {
			for i, cell := range row {
				widths[i] = max(widths[i], ansi.StringWidth(cell))
			}
		}
		line := func(marker string, cells []string, style func(string) string) string {
			parts := make([]string, len(cells))
			for i, cell := range cells {
				parts[i] = style(cell)
				if i < len(cells)-1 {
					parts[i] += strings.Repeat(" ", widths[i]-ansi.StringWidth(cell))
				}
			}
			return marker + strings.Join(parts, "  ") + "\n"
		}
		plain := func(s string) string { return s }
		b.WriteString(line("  ", header, tui.Bold))
		active := false
		for i, e := range list.Runs {
			marker := "  "
			if e.Active {
				marker, active = activeMark+" ", true
			}
			b.WriteString(line(marker, rows[i], plain))
			for _, key := range e.Keys {
				b.WriteString("    " + key + "\n")
			}
			if more := e.ItemCount - len(e.Keys); more > 0 {
				b.WriteString("    " + p.T("run.list.more_keys", map[string]any{"Count": p.Number(float64(more), 0)}) + "\n")
			}
		}
		if active {
			b.WriteString("\n" + activeMark + " " + p.T("run.list.active_legend") + "\n")
		}
	}
	if len(list.Unreadable) > 0 {
		b.WriteString(p.N("run.list.unreadable", len(list.Unreadable), map[string]any{"IDs": strings.Join(list.Unreadable, ", ")}) + "\n")
	}
	if list.More {
		b.WriteString(p.T("run.list.more") + "\n")
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// activeMark starts the row of a run that a process holds, in green.
var activeMark = ansi.Style{}.ForegroundColor(ansi.Green).Styled("●")
