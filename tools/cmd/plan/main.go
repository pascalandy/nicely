// Command plan reads the milestone files and prints the next card, shows the
// progress of every milestone and the order of the current one's cards, or
// checks that the files keep the format that docs/milestones/README.md
// describes, that every
// link in the docs resolves, that every card and code the docs name exists,
// and that the prose holds no em dash or curly quote. It takes the
// repository root:
//
//	plan next|status|check <root>
//	plan limits <root> [--base <ref>]
//
// next and status check the format first, so a broken plan never yields a
// pointer that an agent would follow.
package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
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

// maxAgentCards keeps a milestone small enough for one agent to finish and
// for Pascal to follow.
const maxAgentCards = 5

// readyRule says how many pull requests make a planned milestone ready.
// dev-preferences.md and D042 state it in these words.
const readyRule = "Use one pull request by default, and a small stack when each layer brings a distinct review. A simple fix never needs several pull requests."

var (
	milestoneStatuses = []string{"planned", "ready", "active", "done"}
	cardStatuses      = []string{"todo", "done"}
	owners            = []string{"agent", "Pascal"}
	cardHeader        = []string{"Card", "Title", "Owner", "Depends on", "Status"}

	fileName     = regexp.MustCompile(`^(M\d\d)-[a-z0-9-]+\.md$`)
	decisionName = regexp.MustCompile(`^(D\d{3})-[a-z0-9-]+\.md$`)
	titleLine    = regexp.MustCompile(`^# (M\d\d) (\S.*)$`)
	cardID       = regexp.MustCompile(`^(?:M\d\d|(?:XT|COR)-\d{3})-T[1-9]\d*$`)
	cardRef      = regexp.MustCompile(`\b(?:M\d\d|(?:XT|COR)-\d{3})-T\d+\b`)
	codeShape    = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	codeAfter    = regexp.MustCompile("\\b[Cc]ode `([^`]+)`")
	envelope     = regexp.MustCompile(`"(?:errors|warnings)"\s*:\s*\\?\[`)
	jsonCode     = regexp.MustCompile(`"code"\s*:\s*"([^"]*)"`)
	cardHeading  = regexp.MustCompile(`^### ((?:M\d\d|(?:XT|COR)-\d{3})-T[1-9]\d*) (\S.*)$`)
	checkbox     = regexp.MustCompile(`^\s*[-*+] \[[ xX]\]`)
	codeSpan     = regexp.MustCompile("`([^`]+)`")
	mdLink       = regexp.MustCompile(`\]\(([^)\s]+)\)`)
)

type item struct {
	path, id, title, status, version string
	kind, state                      string
	text                             string
	number                           int
	hasDemo, hasCards, hasHeader     bool
	demo                             []string
	open                             []string // nil when the file has no Open questions section
	cards                            []*card
	links                            []docLink
	refs                             []docLink // the card IDs that the file names
	problems                         []string
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
	var milestones []*item
	var problems []string
	if b != nil {
		milestones = b.items
		problems = b.check(root)
	} else {
		milestones, err = load(root)
		if err != nil {
			return err
		}
		problems = check(root, milestones)
	}
	var output strings.Builder
	if args[0] == "check" {
		codes, err := knownCodes(root)
		if err != nil {
			return err
		}
		for _, m := range milestones {
			problems = append(problems, checkProse(m.path, m.text, codes)...)
		}
		broken, err := checkDocs(root, cardIDs(milestones), codes)
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
		next(&output, milestones)
	case args[0] == "status":
		status(&output, milestones)
	default:
		dir := milestoneDir
		if b != nil {
			dir = backlogDir
		}
		fmt.Fprintf(&output, "%s keeps the format, and the docs pass every check\n", dir)
	}
	if _, err := io.WriteString(out, output.String()); err != nil {
		return err
	}
	if len(problems) > 0 {
		return fmt.Errorf("%d problems in the docs; fix them as each line says", len(problems))
	}
	return nil
}

func load(root string) ([]*item, error) {
	entries, err := os.ReadDir(filepath.Join(root, milestoneDir))
	if err != nil {
		return nil, err
	}
	var milestones []*item
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == "README.md" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, milestoneDir, e.Name()))
		if err != nil {
			return nil, err
		}
		milestones = append(milestones, parse(path.Join(milestoneDir, e.Name()), string(data)))
	}
	slices.SortFunc(milestones, func(a, b *item) int { return cmp.Compare(a.number, b.number) })
	return milestones, nil
}

func parse(file, text string) *item {
	m := &item{path: file, text: text, number: -1, links: links(text), refs: cardRefs(text)}
	if match := fileName.FindStringSubmatch(path.Base(file)); match != nil {
		m.id = match[1]
		m.number, _ = strconv.Atoi(m.id[1:])
	} else {
		m.problem(0, "the file name must look like M01-short-title.md")
	}
	first, _, _ := strings.Cut(text, "\n")
	match := titleLine.FindStringSubmatch(first)
	if match == nil || match[1] != m.id {
		m.problem(1, "the first line must be \"# %s <title>\"", m.id)
	} else {
		m.title = match[2]
	}
	for line := range strings.Lines(text) {
		if v, ok := strings.CutPrefix(line, "Status: "); ok {
			m.status = strings.TrimSpace(v)
		}
		if v, ok := strings.CutPrefix(line, "Version: "); ok {
			m.version = strings.TrimSpace(v)
		}
	}
	m.parseBody(text)
	return m
}

func check(root string, milestones []*item) []string {
	var problems []string
	numbers := map[int]string{}
	known := cardIDs(milestones)
	current := -1
	for i, m := range milestones {
		if other, taken := numbers[m.number]; taken && m.number >= 0 {
			m.problem(0, "%s uses the same number", other)
		}
		numbers[m.number] = m.path
		m.checkLinks(root)
		for _, r := range m.refs {
			if !known[r.target] {
				m.problem(r.line, "no milestone defines the card %s", r.target)
			}
		}
		if m.status == "open" {
			if m.id != "M99" || len(m.cards) > 0 {
				m.problem(0, "only M99, the parking lot, is open, and it holds no cards")
			}
		} else {
			m.checkMilestone(root)
			if current < 0 && m.status != "done" {
				current = i
			}
			if current >= 0 && i > current && (m.status == "active" || m.status == "done" || m.done() > 0) {
				m.problem(0, "milestones go in order, so %s cannot start before %s is done", m.id, milestones[current].id)
			}
		}
		problems = append(problems, m.problems...)
	}
	return problems
}

func (m *item) checkMilestone(root string) {
	if !slices.Contains(milestoneStatuses, m.status) {
		m.problem(0, "Status must be one of %s", strings.Join(milestoneStatuses, ", "))
	}
	// M<n> ships in v0.<n>.0, and M00 in v0.0.1, as decision D043 says.
	want := fmt.Sprintf("v0.%d.0", m.number)
	if m.number == 0 {
		want = "v0.0.1"
	}
	if m.version != want {
		m.problem(0, "Version must be %s, the release that ships %s", want, m.id)
	}
	if !m.hasDemo {
		m.problem(0, "a milestone needs a \"## Demo\" section")
	}
	if !m.hasCards {
		m.problem(0, "a milestone needs a \"## Cards\" section")
	}
	agents, ids := 0, map[string]bool{}
	for _, c := range m.cards {
		m.checkCard(root, c, ids)
		if c.owner == "agent" {
			agents++
		}
	}
	m.checkCycles()
	if agents > maxAgentCards && m.status != "done" {
		m.problem(0, "%d agent cards is more than %d: split the milestone", agents, maxAgentCards)
	}
	done, total := m.done(), len(m.cards)
	switch {
	case m.status != "planned" && m.open != nil:
		m.problem(0, "a %s milestone has no open questions: settle them in the spec, then delete the section", m.status)
	case m.status != "planned" && total == 0:
		m.problem(0, "a %s milestone needs cards", m.status)
	case (m.status == "planned" || m.status == "ready") && done > 0:
		m.problem(0, "a milestone with a done card is active, or done when every card is")
	case m.status == "active" && done == total:
		m.problem(0, "every card is done, so the milestone is done")
	case m.status == "active" && done == 0:
		m.problem(0, "an active milestone has a done card; until then it is ready")
	case m.status == "done" && done < total:
		m.problem(0, "a done milestone has only done cards")
	}
}

// checkDecisions reports a decision file whose name, title, or number breaks
// the rules of decisions/README.md, or that has no row in that index.
func checkDecisions(root string) []string {
	entries, err := os.ReadDir(filepath.Join(root, decisionDir))
	if err != nil {
		return nil
	}
	index, _ := os.ReadFile(filepath.Join(root, decisionDir, "README.md"))
	var problems []string
	seen := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || e.Name() == "README.md" {
			continue
		}
		where := decisionDir + "/" + e.Name()
		match := decisionName.FindStringSubmatch(e.Name())
		if match == nil {
			problems = append(problems, where+": the file name must look like D001-short-title.md")
			continue
		}
		id := match[1]
		if other, taken := seen[id]; taken {
			problems = append(problems, fmt.Sprintf("%s: %s uses the same number", where, other))
		}
		seen[id] = e.Name()
		data, err := os.ReadFile(filepath.Join(root, decisionDir, e.Name()))
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", where, err))
			continue
		}
		if !strings.HasPrefix(string(data), "# "+id+" ") {
			problems = append(problems, fmt.Sprintf("%s:1: the first line must be \"# %s <title>\"", where, id))
		}
		if !strings.Contains(string(index), "]("+e.Name()+")") {
			problems = append(problems, fmt.Sprintf("%s: the decision needs a row in %s/README.md", where, decisionDir))
		}
	}
	return problems
}

// knownCodes returns the codes that a doc may name. They come from the tables
// under codeSections in the parts of contract.md and in each extension's spec,
// whole or split into spec/.
// They also come from the errors that the fixture manifests declare, and from
// exampleCode.
func knownCodes(root string) (map[string]bool, error) {
	codes := map[string]bool{exampleCode: true}
	parts, err := filepath.Glob(filepath.Join(root, "docs", "north-star", "contract", "*.md"))
	if err != nil {
		return nil, err
	}
	specs, err := filepath.Glob(filepath.Join(root, "extensions", "*", "spec.md"))
	if err != nil {
		return nil, err
	}
	specParts, err := filepath.Glob(filepath.Join(root, "extensions", "*", "spec", "*.md"))
	if err != nil {
		return nil, err
	}
	for _, name := range slices.Concat(parts, specs, specParts) {
		data, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}
		section := ""
		var header []string
		var f fence
		for line := range strings.Lines(string(data)) {
			// A table in an example, such as a fenced Markdown block, defines no code.
			if f.in(line) {
				header = nil
				continue
			}
			if m := sectionTitle.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
				section = m[1]
			}
			cells, isHeader := tableRow(line, &header)
			if cells == nil || isHeader || !slices.Contains(codeSections, section) {
				continue
			}
			for i, h := range header {
				if (h == "Code" || h == "Warning") && i < len(cells) {
					for _, c := range cellCodes(cells[i]) {
						codes[c] = true
					}
				}
			}
		}
	}
	manifests, err := filepath.Glob(filepath.Join(root, "testdata", "extensions", "*", "manifest.toml"))
	if err != nil {
		return nil, err
	}
	for _, name := range manifests {
		var manifest struct {
			Errors []struct {
				Code string `toml:"code"`
			} `toml:"errors"`
		}
		if _, err := toml.DecodeFile(name, &manifest); err != nil {
			rel, _ := filepath.Rel(root, name)
			return nil, fmt.Errorf("%s: %w", filepath.ToSlash(rel), err)
		}
		for _, e := range manifest.Errors {
			codes[e.Code] = true
		}
	}
	return codes, nil
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

// cardIDs returns the cards that the tables of the milestones define.
func cardIDs(milestones []*item) map[string]bool {
	known := map[string]bool{}
	for _, m := range milestones {
		for _, c := range m.cards {
			known[c.id] = true
		}
	}
	return known
}

// currentMilestone returns the first milestone that is not done, or nil.
func currentMilestone(milestones []*item) *item {
	for _, m := range milestones {
		if m.status != "done" && m.status != "open" {
			return m
		}
	}
	return nil
}

func next(out *strings.Builder, milestones []*item) {
	m := currentMilestone(milestones)
	if m == nil {
		fmt.Fprintln(out, "Every milestone is done. Plan the next one, as \"Make a milestone ready\" in docs/milestones/README.md says.")
		return
	}
	if m.status == "planned" && (m.open != nil || len(m.cards) == 0) {
		fmt.Fprintf(out, "Next: make %s %s ready\nFile: %s\n\n", m.id, m.title, m.path)
		fmt.Fprintf(out, "The milestone is planned. Settle its open questions in the spec, finish its cards, and set its status to ready, as \"Make a milestone ready\" in docs/milestones/README.md says. %s\n", readyRule)
		if len(m.open) > 0 {
			fmt.Fprintf(out, "\nOpen questions:\n%s\n", strings.Join(m.open, "\n"))
		}
		return
	}
	agent, pascal := startable(m, nil)
	for _, c := range pascal {
		fmt.Fprintf(out, "Waiting on Pascal: %s %s\n", c.id, c.title)
	}
	if len(pascal) > 0 {
		fmt.Fprintln(out, "Prepare what each card for Pascal asks, if nobody has yet, then tell Pascal.")
		fmt.Fprintln(out)
	}
	if len(agent) == 0 {
		fmt.Fprintf(out, "No agent card of %s can start until Pascal finishes the cards above.\n", m.id)
		for _, c := range pascal {
			fmt.Fprintf(out, "\n### %s %s\n%s\n", c.id, c.title, body(c))
		}
		return
	}
	c := agent[0]
	fmt.Fprintf(out, "Next card: %s %s\nMilestone: %s %s, which ships in %s\nFile: %s\n\n%s\n\n", c.id, c.title, m.id, m.title, m.version, m.path, body(c))
	fmt.Fprintln(out, "Follow \"Work on a card\" in AGENTS.md. The pull request sets this card to done.")
	if m.status == "planned" {
		fmt.Fprintf(out, "%s has no open questions, so this pull request also sets it to active.\n", m.id)
	}
}

// startable returns the todo cards of m whose dependencies are done, by owner
// and in the order of the table. next serves the first agent card. A card in
// assumed counts as done, so status can plan ahead without changing a card.
func startable(m *item, assumed map[string]bool) (agent, pascal []*card) {
	done := func(c *card) bool { return c.status == "done" || assumed[c.id] }
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

// order returns the agent cards of m in the order that next serves them, if
// each one succeeds. Pascal's cards keep their real status, so the order
// stops when every agent card left waits on one of them.
func order(m *item) []*card {
	assumed := map[string]bool{}
	var cards []*card
	for {
		agent, _ := startable(m, assumed)
		if len(agent) == 0 {
			return cards
		}
		assumed[agent[0].id] = true
		cards = append(cards, agent[0])
	}
}

func body(c *card) string {
	return strings.TrimSpace(strings.Join(c.body, "\n"))
}

func status(out *strings.Builder, milestones []*item) {
	current := currentMilestone(milestones)
	var rows [][]string
	for _, m := range milestones {
		if m.status == "open" {
			continue
		}
		title := m.title
		if m == current {
			title += "  <- current"
		}
		rows = append(rows, []string{m.id, m.status, fmt.Sprintf("%d/%d", m.done(), len(m.cards)), title})
	}
	out.WriteString(table("", rows))
	if current == nil {
		return
	}
	fmt.Fprintf(out, "\n%s %s (%s), ships in %s\n", current.id, current.title, current.status, current.version)
	rows = nil
	for _, c := range current.cards {
		title := c.title
		if len(c.depends) > 0 {
			title += "  after " + strings.Join(c.depends, ", ")
		}
		rows = append(rows, []string{c.status, c.owner, c.id, title})
	}
	out.WriteString(table("  ", rows))
	served := order(current)
	var ids, blocked, notes []string
	for _, c := range served {
		ids = append(ids, c.id)
	}
	for _, c := range current.cards {
		if c.owner == "agent" && c.status == "todo" && !slices.Contains(served, c) {
			blocked = append(blocked, c.id)
		}
	}
	if len(ids) > 0 {
		notes = append(notes, "If each agent card succeeds, just next serves them in this order: "+strings.Join(ids, ", "))
	}
	if len(blocked) > 0 {
		notes = append(notes, "These agent cards wait on a card of Pascal: "+strings.Join(blocked, ", "))
	}
	if len(notes) > 0 {
		fmt.Fprintf(out, "\n%s\n", strings.Join(notes, "\n"))
	}
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
