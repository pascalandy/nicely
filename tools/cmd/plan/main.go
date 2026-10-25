// Command plan validates the backlog, prints the next queued card, and shows item states.
// It also checks documentation links, card references, codes, and typography.
// next and status validate the plan before pointing an agent to work.
//
// plan next|status|check <repository root>
// plan limits <repository root> [--base <ref>]
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
)

const (
	milestoneDir = "docs/milestones"
	decisionDir  = "docs/north-star/decisions"
)

// codeSections are the section titles whose tables define codes, under a Code
// or a Warning column. A title is "# " in a part of a split doc, and "## " in
// a whole one.
var codeSections = []string{"Error codes", "Warning codes", "Codes"}

var sectionTitle = regexp.MustCompile(`^##? (.+)$`)

// exampleCode is the one code that a doc may name without a source: the Error
// codes section of contract.md gives it as an extension's own code, and no
// extension declares it.
const exampleCode = "FLEET_HOST_DOWN"

// plainForms maps each character that prose avoids to what to write instead.
var plainForms = map[rune]string{
	'\u2014': "a colon, a comma, or two sentences",
	'\u2018': "a straight quote",
	'\u2019': "a straight quote",
	'\u201C': "a straight quote",
	'\u201D': "a straight quote",
}

// maxAgentCards keeps an item small enough for one agent to finish and
// for Pascal to follow.
const maxAgentCards = 5

// readyRule says how many pull requests make a Planned item Ready.
const readyRule = "Use one pull request by default, and a small stack when each layer brings a distinct review. A simple fix never needs several pull requests."

var (
	cardStatuses = []string{"todo", "done"}
	owners       = []string{"agent", "Pascal"}
	cardHeader   = []string{"Card", "Title", "Owner", "Depends on", "Status"}

	decisionName = regexp.MustCompile(`^(D\d{3})-[a-z0-9-]+\.md$`)
	cardID       = regexp.MustCompile(`^(?:XT|COR)-\d{3}-T[1-9]\d*$`)
	cardRef      = regexp.MustCompile(`\b(?:XT|COR)-\d{3}-T\d+\b`)
	codeShape    = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	codeAfter    = regexp.MustCompile("\\b[Cc]ode `([^`]+)`")
	envelope     = regexp.MustCompile(`"(?:errors|warnings)"\s*:\s*\\?\[`)
	jsonCode     = regexp.MustCompile(`"code"\s*:\s*"([^"]*)"`)
	cardHeading  = regexp.MustCompile(`^### ((?:XT|COR)-\d{3}-T[1-9]\d*) (\S.*)$`)
	checkbox     = regexp.MustCompile(`^\s*[-*+] \[[ xX]\]`)
	codeSpan     = regexp.MustCompile("`([^`]+)`")
	mdLink       = regexp.MustCompile(`\]\(([^)\s]+)\)`)
)

type item struct {
	path, id, title, status      string
	state                        string
	text                         string
	hasDemo, hasCards, hasHeader bool
	demo                         []string
	open                         []string // nil when the file has no Open questions section
	cards                        []*card
	links                        []docLink
	refs                         []docLink // the card IDs that the file names
	problems                     []string
}

type card struct {
	id, title, owner, status string
	depends                  []string
	line                     int
	hasSection               bool
	body                     []string
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "plan:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) > 0 && args[0] == "limits" {
		return runLimits(args[1:])
	}
	if len(args) != 2 || !slices.Contains([]string{"next", "status", "check"}, args[0]) {
		return errors.New("usage: plan next|status|check <repository root>")
	}
	root := args[1]
	b, err := loadBacklog(root)
	if err != nil {
		return err
	}
	if b == nil {
		return errors.New("docs/backlog/README.md is required")
	}
	items := b.items
	problems := b.check(root)
	var output strings.Builder
	if args[0] == "check" {
		codes, err := knownCodes(root)
		if err != nil {
			return err
		}
		for _, m := range items {
			text := m.text
			if m.state != "" {
				lines := strings.Split(text, "\n")
				lines[3] = "state: " + m.state
				text = strings.Join(lines, "\n")
			}
			problems = append(problems, checkProse(m.path, text, codes)...)
		}
		broken, err := checkDocs(root, cardIDs(items), codes, backlogDir)
		if err != nil {
			return err
		}
		problems = append(problems, broken...)
		problems = append(problems, checkDecisions(root)...)
	}
	switch {
	case len(problems) > 0:
		output.WriteString(strings.Join(problems, "\n") + "\n")
	case args[0] == "next":
		nextItem(&output, b)
	case args[0] == "status":
		itemStatus(&output, b.items)
	default:
		fmt.Fprintf(&output, "%s keeps the format, and the docs pass every check\n", backlogDir)
	}
	if _, err := io.WriteString(out, output.String()); err != nil {
		return err
	}
	if len(problems) > 0 {
		return fmt.Errorf("%d problems in the docs; fix them as each line says", len(problems))
	}
	return nil
}

// checkProse reports, in text, each code outside codes in a context whose
// role is explicit, and each character of plainForms in prose. The contexts
// are a cell under a Code column, a code span right after "code", such as
// "error code `X`", and the "code" of an object in an "errors" or
// "warnings" array. Prose is the text outside code, where a table cell that
// holds only an em dash, an empty Depends on, is no prose. Code is a fenced
// block or a code span on one line, as for links. A value that is no code,
// such as the exit code `78`, is no reference.
func checkProse(file, text string, codes map[string]bool) []string {
	var problems []string
	code := func(n int, value string) {
		if codeShape.MatchString(value) && !codes[value] {
			problems = append(problems, fmt.Sprintf("%s:%d: no code table or fixture manifest defines the code %s", file, n, value))
		}
	}
	var header, block []string
	var f fence
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		n := i + 1
		if opened := f.open != ""; f.in(line) {
			switch {
			case !opened:
				block, header = nil, nil
			case f.open == "":
				envelopeCodes(strings.Join(block, "\n"), n-len(block), code)
			default:
				block = append(block, line)
			}
			continue
		}
		envelopeCodes(line, n, code)
		for _, m := range codeAfter.FindAllStringSubmatch(line, -1) {
			code(n, m[1])
		}
		prose := []string{line}
		if cells, isHeader := tableRow(line, &header); cells != nil {
			prose = nil
			for i, cell := range cells {
				if !isHeader && i < len(header) && header[i] == "Code" {
					for _, c := range cellCodes(cell) {
						code(n, c)
					}
				}
				if cell != "\u2014" {
					prose = append(prose, cell)
				}
			}
		}
		for _, p := range prose {
			for _, r := range codeSpan.ReplaceAllString(p, "") {
				if plain, ok := plainForms[r]; ok {
					problems = append(problems, fmt.Sprintf("%s:%d: the prose holds %c (U+%04X): write %s", file, n, r, r, plain))
				}
			}
		}
	}
	// A block that the file never closes runs to its end, as CommonMark says.
	if f.open != "" {
		envelopeCodes(strings.Join(block, "\n"), len(lines)-len(block)+1, code)
	}
	return problems
}

// envelopeCodes calls code with the line and the value of each "code" inside
// an "errors" or "warnings" array of chunk, a line or a code block whose first
// line is first. The array ends at its matching bracket outside a string, or
// with the chunk.
func envelopeCodes(chunk string, first int, code func(int, string)) {
	for _, loc := range envelope.FindAllStringIndex(chunk, -1) {
		end, depth, quoted := len(chunk), 0, false
	scan:
		for i := loc[1] - 1; i < len(chunk); i++ {
			switch c := chunk[i]; {
			case quoted && c == '\\':
				i++
			case c == '"':
				quoted = !quoted
			case quoted:
			case c == '[':
				depth++
			case c == ']':
				if depth--; depth == 0 {
					end = i
					break scan
				}
			}
		}
		for _, m := range jsonCode.FindAllStringSubmatchIndex(chunk[loc[1]:end], -1) {
			at := loc[1] + m[2]
			code(first+strings.Count(chunk[:at], "\n"), chunk[at:loc[1]+m[3]])
		}
	}
}

// tableRow returns the cells of a table row, or nil for another line, which
// ends the table. The first row of a table is its header, which it stores in
// header; the delimiter row returns no cells.
func tableRow(line string, header *[]string) (cells []string, isHeader bool) {
	if !strings.HasPrefix(strings.TrimSpace(line), "|") {
		*header = nil
		return nil, false
	}
	cells = tableCells(line)
	switch {
	case *header == nil:
		*header = cells
		return cells, true
	case strings.HasPrefix(strings.TrimSpace(line), "|---"):
		return nil, false
	}
	return cells, false
}

// tableCells splits a table row into its cells, without their spaces.
func tableCells(line string) []string {
	var cells []string
	for cell := range strings.SplitSeq(strings.Trim(strings.TrimSpace(line), "|"), "|") {
		cells = append(cells, strings.TrimSpace(cell))
	}
	return cells
}

// cellCodes returns the code spans of a cell, or the cell when it has none.
func cellCodes(cell string) []string {
	spans := codeSpan.FindAllStringSubmatch(cell, -1)
	if len(spans) == 0 {
		return []string{cell}
	}
	var values []string
	for _, s := range spans {
		values = append(values, s[1])
	}
	return values
}

// fence follows the fenced code blocks of a Markdown file, indented or not. A
// block opens with three or more backticks or tildes, and only a line of at
// least as many of the same character, alone, closes it.
type fence struct {
	open string // the marker of the current block, or "" outside one
}

// in reports whether line opens, closes, or lies inside a fenced block.
func (f *fence) in(line string) bool {
	text := strings.TrimLeft(line, " ")
	marker := ""
	if strings.HasPrefix(text, "```") || strings.HasPrefix(text, "~~~") {
		marker = text[:len(text)-len(strings.TrimLeft(text, text[:1]))]
	}
	switch {
	case f.open == "":
		f.open = marker
		return marker != ""
	case marker != "" && marker[0] == f.open[0] && len(marker) >= len(f.open) && strings.TrimSpace(text[len(marker):]) == "":
		f.open = ""
	}
	return true
}

// cardIDs returns the cards that the item tables define.
func cardIDs(items []*item) map[string]bool {
	known := map[string]bool{}
	for _, m := range items {
		for _, c := range m.cards {
			known[c.id] = true
		}
	}
	return known
}

// startable returns the todo cards of m whose dependencies are done, by owner
// and in table order.
func startable(m *item) (agent, pascal []*card) {
	done := func(c *card) bool { return c.status == "done" }
	for _, c := range m.cards {
		waiting := slices.ContainsFunc(c.depends, func(d string) bool { return !done(m.card(d)) })
		if done(c) || waiting {
			continue
		}
		if c.owner == "agent" {
			agent = append(agent, c)
		} else {
			pascal = append(pascal, c)
		}
	}
	return agent, pascal
}

func body(c *card) string {
	return strings.TrimSpace(strings.Join(c.body, "\n"))
}

// table aligns rows in columns two spaces apart, after indent.
func table(indent string, rows [][]string) string {
	var widths []int
	for _, row := range rows {
		for i, cell := range row {
			if i == len(widths) {
				widths = append(widths, 0)
			}
			widths[i] = max(widths[i], len(cell))
		}
	}
	var b strings.Builder
	for _, row := range rows {
		b.WriteString(indent)
		for i, cell := range row {
			if i == len(row)-1 {
				b.WriteString(cell + "\n")
			} else {
				b.WriteString(cell + strings.Repeat(" ", widths[i]-len(cell)+2))
			}
		}
	}
	return b.String()
}
