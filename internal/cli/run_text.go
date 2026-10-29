package cli

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/record"
	"github.com/pascalandy/nicely/sdk/tui"
)

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
		line := func(marker string, cells []string, styles []func(string) string) string {
			parts := make([]string, len(cells))
			for i, cell := range cells {
				parts[i] = styles[i](cell)
				if i < len(cells)-1 {
					parts[i] += strings.Repeat(" ", widths[i]-ansi.StringWidth(cell))
				}
			}
			return marker + strings.Join(parts, "  ") + "\n"
		}
		b.WriteString(line("  ", header, slices.Repeat([]func(string) string{tui.Bold}, len(header))))
		active := false
		for i, e := range list.Runs {
			marker := "  "
			if e.Active {
				marker, active = activeMark+" ", true
			}
			b.WriteString(line(marker, rows[i], []func(string) string{unstyled, styleOf(listStatus, record.Status(e.Status)), unstyled, unstyled, unstyled}))
			for _, key := range e.Keys {
				b.WriteString("    " + tui.Muted(key) + "\n")
			}
			if more := e.ItemCount - len(e.Keys); more > 0 {
				b.WriteString("    " + tui.Muted(p.N("run.list.more_keys", more)) + "\n")
			}
		}
		if active {
			b.WriteString("\n" + activeMark + " " + tui.Muted(p.T("run.list.active_legend")) + "\n")
		}
	}
	if len(list.Unreadable) > 0 {
		b.WriteString(p.N("run.list.unreadable", len(list.Unreadable), map[string]any{"IDs": strings.Join(list.Unreadable, ", ")}) + "\n")
	}
	if list.More {
		b.WriteString(tui.Muted(p.T("run.list.more")) + "\n")
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// activeMark starts the row of a run that a process holds, in green.
var activeMark = colored(ansi.Green)("●")

// listStatus colors the status of a row of run list, and stepStatus the
// status of a step of run view, as ncly run in core-spec.md defines. A status
// that a map leaves out stays plain, as does the status of a run in view
// other than completed, because the picked looks give it no style.
var (
	listStatus = map[record.Status]func(string) string{
		record.Completed:   colored(ansi.Green),
		record.Running:     colored(ansi.Cyan),
		record.Failed:      colored(ansi.Red),
		record.Unknown:     colored(ansi.Yellow),
		record.Interrupted: colored(ansi.Yellow),
	}
	viewStatus = map[record.Status]func(string) string{
		record.Completed: colored(ansi.Green),
	}
	stepStatus = map[record.StepStatus]func(string) string{
		record.StepCompleted: colored(ansi.Green),
		record.StepFailed:    colored(ansi.Red),
		record.StepPending:   tui.Muted,
	}
)

// colored returns the style of a text in one of the 16 ANSI colors, which
// tui.Output removes when color is off.
func colored(color ansi.BasicColor) func(string) string {
	style := ansi.Style{}.ForegroundColor(color)
	return func(s string) string { return style.Styled(s) }
}

// unstyled leaves a text plain.
func unstyled(s string) string { return s }

// styleOf returns the style that styles gives a status, or none.
func styleOf[S comparable](styles map[S]func(string) string, status S) func(string) string {
	if style, found := styles[status]; found {
		return style
	}
	return unstyled
}

// writeRun prints a record as ncly run in core-spec.md shows it: the run
// ID and the command path, its facts as labeled lines, then its inputs, its
// resolved values, the versions of its tools, and its items under bold
// headings. keys are the keys of the record file, whose values keep their
// numbers as the file holds them.
func writeRun(w io.Writer, p *i18n.Printer, id string, rec *record.Record, keys map[string]any, active bool) error {
	var b strings.Builder
	b.WriteString(tui.Bold(id) + "  " + rec.Path + "\n")
	held := p.T("run.view.no")
	if active {
		held = p.T("run.view.yes")
	}
	writeValues(&b, "", tui.Muted, [][2]any{
		{p.T("run.view.status"), styleOf(viewStatus, rec.Status)(string(rec.Status))},
		{p.T("run.view.active"), held},
		{p.T("run.view.started"), rec.StartedAt},
		{p.T("run.view.updated"), rec.UpdatedAt},
		{p.T("run.view.attempts"), p.Number(float64(rec.Attempts), 0)},
		{p.T("run.view.ncly"), rec.NclyVersion},
	})
	// The labels of the inputs and the resolved values are bright black, as
	// those of the facts are, and each tool stays plain beside its version.
	for _, section := range []struct {
		heading, key string
		label        func(string) string
	}{
		{"run.view.inputs", "inputs", tui.Muted},
		{"run.view.resolved", "resolved", tui.Muted},
		{"run.view.tools", "tools", unstyled},
	} {
		if values, _ := keys[section.key].(map[string]any); len(values) > 0 {
			b.WriteString("\n" + tui.Bold(p.T(section.heading)) + "\n")
			writeValues(&b, "  ", section.label, sortedValues(values))
		}
	}
	if len(rec.Items) > 0 {
		b.WriteString("\n" + tui.Bold(p.T("run.view.items")) + "\n")
	}
	widths := [3]int{}
	for _, it := range rec.Items {
		for _, s := range it.Steps {
			for i, cell := range []string{string(s.Status), s.ID, string(s.Effect)} {
				widths[i] = max(widths[i], ansi.StringWidth(cell))
			}
		}
	}
	for _, it := range rec.Items {
		b.WriteString("  " + it.Key)
		if it.OutputDir != nil {
			b.WriteString(" " + tui.Muted("→") + " " + *it.OutputDir)
		}
		b.WriteString("\n")
		for _, s := range it.Steps {
			cells := []string{string(s.Status), s.ID, string(s.Effect)}
			styles := []func(string) string{styleOf(stepStatus, s.Status), unstyled, tui.Muted}
			for i, cell := range cells {
				cells[i] = styles[i](cell) + strings.Repeat(" ", widths[i]-ansi.StringWidth(cell))
			}
			line := strings.Join(cells, "  ")
			if s.StartedAt != nil {
				line += "  " + stepTimes(p, *s.StartedAt, s.FinishedAt)
			}
			b.WriteString("    " + strings.TrimRight(line, " ") + "\n")
			for _, a := range s.Artifacts {
				b.WriteString("      " + a.Path + "  " + tui.Muted(p.Size(a.SizeBytes)) + "\n")
			}
			if s.Error != nil {
				b.WriteString("      " + tui.ErrorLabel(p.T("error.label", map[string]any{"Code": s.Error.Code})) + " " + s.Error.Message + "\n")
				if s.Error.Hint != "" {
					b.WriteString("      " + tui.HintLabel(p.T("error.hint")) + " " + s.Error.Hint + "\n")
				}
			}
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// stepTimes shows when a step started, then, once it finished, when, and
// how long it took.
func stepTimes(p *i18n.Printer, started string, finished *string) string {
	if finished == nil {
		return started
	}
	text := started + " → " + *finished
	from, errFrom := time.Parse(time.RFC3339, started)
	to, errTo := time.Parse(time.RFC3339, *finished)
	if errFrom == nil && errTo == nil && !to.Before(from) {
		text += ", " + p.Duration(to.Sub(from))
	}
	return text
}

// sortedValues returns the keys of an object, in name order, with their
// values.
func sortedValues(values map[string]any) [][2]any {
	var pairs [][2]any
	for _, k := range slices.Sorted(maps.Keys(values)) {
		pairs = append(pairs, [2]any{k, values[k]})
	}
	return pairs
}

// writeValues prints labeled values in aligned columns, each label at
// indent, in the style of label. An object takes its own labeled lines under
// its label, and an array one line per element, both indented once more.
func writeValues(b *strings.Builder, indent string, label func(string) string, pairs [][2]any) {
	width := 0
	for _, pair := range pairs {
		width = max(width, ansi.StringWidth(pair[0].(string)))
	}
	for _, pair := range pairs {
		name := pair[0].(string)
		switch v := pair[1].(type) {
		case map[string]any:
			b.WriteString(indent + label(name) + "\n")
			writeValues(b, indent+"  ", label, sortedValues(v))
		case []any:
			b.WriteString(indent + label(name) + "\n")
			for _, element := range v {
				if object, isObject := element.(map[string]any); isObject {
					writeValues(b, indent+"  ", label, sortedValues(object))
					continue
				}
				b.WriteString(indent + "  " + scalar(element) + "\n")
			}
		default:
			b.WriteString(indent + label(name) + strings.Repeat(" ", width-ansi.StringWidth(name)+2) + scalar(v) + "\n")
		}
	}
}

// scalar shows a JSON value that is no object and no array.
func scalar(v any) string {
	switch v := v.(type) {
	case nil:
		return "null"
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	}
	return fmt.Sprint(v)
}
