// Command plan reads the milestone files and prints the next card, shows the
// progress of every milestone, or checks that the files keep the format that
// AGENTS.md describes. It takes the repository root:
//
//	plan next|status|check <root>
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
	"unicode"
)

const milestoneDir = "docs/milestones"

// maxAgentCards keeps a milestone small enough for one agent to finish and
// for Pascal to follow.
const maxAgentCards = 5

var (
	milestoneStatuses = []string{"planned", "ready", "active", "done"}
	cardStatuses      = []string{"todo", "done"}
	owners            = []string{"agent", "Pascal"}
	cardHeader        = []string{"Card", "Title", "Owner", "Depends on", "Status"}

	fileName    = regexp.MustCompile(`^(M\d\d)-[a-z0-9-]+\.md$`)
	titleLine   = regexp.MustCompile(`^# (M\d\d) (\S.*)$`)
	cardID      = regexp.MustCompile(`^M\d\d-T\d+$`)
	cardHeading = regexp.MustCompile(`^### (M\d\d-T\d+) (\S.*)$`)
	checkbox    = regexp.MustCompile(`^\s*[-*+] \[[ xX]\]`)
	codeSpan    = regexp.MustCompile("`([^`]+)`")
	mdLink      = regexp.MustCompile(`\]\(([^)\s]+)\)`)
)

type milestone struct {
	path, id, title, status, version string
	number                           int
	hasDemo, hasCards, hasHeader     bool
	open                             []string // nil when the file has no Open questions section
	cards                            []*card
	links                            []docLink
	problems                         []string
}

type card struct {
	id, title, owner, status string
	depends                  []string
	line                     int
	hasSection               bool
	body                     []string
}

type docLink struct {
	target string
	line   int
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "plan:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) != 2 || !slices.Contains([]string{"next", "status", "check"}, args[0]) {
		return errors.New("usage: plan next|status|check <repository root>")
	}
	root := args[1]
	milestones, err := load(root)
	if err != nil {
		return err
	}
	var b strings.Builder
	problems := check(root, milestones)
	switch {
	case len(problems) > 0:
		b.WriteString(strings.Join(problems, "\n") + "\n")
	case args[0] == "next":
		next(&b, milestones)
	case args[0] == "status":
		status(&b, milestones)
	default:
		fmt.Fprintf(&b, "%s keeps the format\n", milestoneDir)
	}
	if _, err := io.WriteString(out, b.String()); err != nil {
		return err
	}
	if len(problems) > 0 {
		return fmt.Errorf("%d problems in %s; fix them as AGENTS.md describes", len(problems), milestoneDir)
	}
	return nil
}

func load(root string) ([]*milestone, error) {
	entries, err := os.ReadDir(filepath.Join(root, milestoneDir))
	if err != nil {
		return nil, err
	}
	var milestones []*milestone
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, milestoneDir, e.Name()))
		if err != nil {
			return nil, err
		}
		milestones = append(milestones, parse(path.Join(milestoneDir, e.Name()), string(data)))
	}
	slices.SortFunc(milestones, func(a, b *milestone) int { return cmp.Compare(a.number, b.number) })
	return milestones, nil
}

func parse(file, text string) *milestone {
	m := &milestone{path: file, number: -1}
	if match := fileName.FindStringSubmatch(path.Base(file)); match != nil {
		m.id = match[1]
		m.number, _ = strconv.Atoi(m.id[1:])
	} else {
		m.problem(0, "the file name must look like M01-short-title.md")
	}
	section := ""
	var current *card
	fenced := false
	for i, line := range strings.Split(text, "\n") {
		n := i + 1
		if checkbox.MatchString(line) {
			m.problem(n, "a checkbox belongs in no milestone file: make the work a card")
		}
		if strings.HasPrefix(line, "```") {
			fenced = !fenced
		}
		if !fenced {
			for _, l := range mdLink.FindAllStringSubmatch(line, -1) {
				m.links = append(m.links, docLink{l[1], n})
			}
		}
		switch {
		case fenced || strings.HasPrefix(line, "```"):
		case i == 0:
			match := titleLine.FindStringSubmatch(line)
			if match == nil || match[1] != m.id {
				m.problem(n, "the first line must be \"# %s <title>\"", m.id)
			} else {
				m.title = match[2]
			}
			continue
		case strings.HasPrefix(line, "## "):
			section, current = strings.TrimPrefix(line, "## "), nil
			switch section {
			case "Demo":
				m.hasDemo = true
			case "Cards":
				m.hasCards = true
			case "Open questions":
				m.open = []string{}
			}
			continue
		case strings.HasPrefix(line, "#"):
			current = nil
			if section == "Cards" {
				current = m.section(n, line)
			}
			continue
		case section == "":
			if v, ok := strings.CutPrefix(line, "Status: "); ok {
				m.status = v
			}
			if v, ok := strings.CutPrefix(line, "Version: "); ok {
				m.version = v
			}
			continue
		case section == "Cards" && current == nil && strings.HasPrefix(line, "|"):
			m.row(n, line)
			continue
		}
		switch {
		case current != nil:
			current.body = append(current.body, line)
		case section == "Open questions" && strings.TrimSpace(line) != "":
			m.open = append(m.open, line)
		}
	}
	return m
}

// row reads one line of the cards table.
func (m *milestone) row(n int, line string) {
	var cells []string
	for cell := range strings.SplitSeq(strings.Trim(strings.TrimSpace(line), "|"), "|") {
		cells = append(cells, strings.TrimSpace(cell))
	}
	switch {
	case strings.HasPrefix(strings.TrimSpace(line), "|---"):
		return
	case slices.Equal(cells, cardHeader):
		m.hasHeader = true
		return
	case !m.hasHeader:
		m.problem(n, "the cards table must start with the columns %s", strings.Join(cardHeader, ", "))
		return
	case len(cells) != len(cardHeader):
		m.problem(n, "a card row has %d cells, not %d", len(cells), len(cardHeader))
		return
	}
	c := &card{id: cells[0], title: cells[1], owner: cells[2], status: cells[4], line: n}
	if deps := cells[3]; deps != "—" && deps != "-" && deps != "" {
		for d := range strings.SplitSeq(deps, ",") {
			c.depends = append(c.depends, strings.TrimSpace(d))
		}
	}
	m.cards = append(m.cards, c)
}

// section starts the section of the card that a "### " heading names.
func (m *milestone) section(n int, line string) *card {
	match := cardHeading.FindStringSubmatch(line)
	if match == nil {
		m.problem(n, "a heading in Cards must look like \"### %s-T1 <title>\"", m.id)
		return nil
	}
	c := m.card(match[1])
	switch {
	case c == nil:
		m.problem(n, "%s has a section but no row in the cards table", match[1])
		return nil
	case c.hasSection:
		m.problem(n, "%s has two sections", c.id)
		return nil
	case match[2] != c.title:
		m.problem(n, "the section of %s is titled %q, and its row %q", c.id, match[2], c.title)
	}
	c.hasSection = true
	return c
}

func (m *milestone) card(id string) *card {
	for _, c := range m.cards {
		if c.id == id {
			return c
		}
	}
	return nil
}

func (m *milestone) problem(line int, format string, args ...any) {
	where := m.path
	if line > 0 {
		where += ":" + strconv.Itoa(line)
	}
	m.problems = append(m.problems, where+": "+fmt.Sprintf(format, args...))
}

func (m *milestone) done() int {
	n := 0
	for _, c := range m.cards {
		if c.status == "done" {
			n++
		}
	}
	return n
}

// field returns the text after a "- **Name:**" line of the card's section.
func (c *card) field(name string) (string, bool) {
	for _, line := range c.body {
		if v, ok := strings.CutPrefix(line, "- **"+name+":**"); ok {
			return strings.TrimSpace(v), true
		}
	}
	return "", false
}

func check(root string, milestones []*milestone) []string {
	var problems []string
	numbers := map[int]string{}
	current := -1
	for i, m := range milestones {
		if other, taken := numbers[m.number]; taken && m.number >= 0 {
			m.problem(0, "%s uses the same number", other)
		}
		numbers[m.number] = m.path
		m.checkLinks(root)
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

func (m *milestone) checkMilestone(root string) {
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

func (m *milestone) checkCard(root string, c *card, ids map[string]bool) {
	if !strings.HasPrefix(c.id, m.id+"-T") || !cardID.MatchString(c.id) {
		m.problem(c.line, "the card %q must be named %s-T<number>", c.id, m.id)
	}
	if ids[c.id] {
		m.problem(c.line, "%s appears twice in the cards table", c.id)
	}
	ids[c.id] = true
	if !slices.Contains(owners, c.owner) {
		m.problem(c.line, "the owner of %s must be one of %s", c.id, strings.Join(owners, ", "))
	}
	if !slices.Contains(cardStatuses, c.status) {
		m.problem(c.line, "the status of %s must be one of %s", c.id, strings.Join(cardStatuses, ", "))
	}
	if !c.hasSection {
		m.problem(c.line, "%s needs a section \"### %s %s\"", c.id, c.id, c.title)
	}
	for _, d := range c.depends {
		dep := m.card(d)
		switch {
		case dep == nil || d == c.id:
			m.problem(c.line, "%s depends on %q, which is no other card of %s", c.id, d, m.id)
		case c.status == "done" && dep.status != "done":
			m.problem(c.line, "%s is done before %s, which it depends on", c.id, d)
		}
	}
	if c.owner != "agent" {
		return
	}
	if read, _ := c.field("Read"); !mdLink.MatchString(read) {
		m.problem(c.line, "the agent card %s needs a \"- **Read:**\" line that links what to read", c.id)
	}
	proves, _ := c.field("Proves")
	if proves == "" {
		m.problem(c.line, "the agent card %s needs a \"- **Proves:**\" line that names its proof", c.id)
	}
	if c.status != "done" {
		return
	}
	// A done card's proof exists: every code span of its Proves line is a
	// path from the repository root.
	for _, span := range codeSpan.FindAllStringSubmatch(proves, -1) {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(span[1]))); err != nil {
			m.problem(c.line, "%s is done, but its proof %s does not exist", c.id, span[1])
		}
	}
}

func (m *milestone) checkCycles() {
	state := map[string]int{} // 1 while visiting, 2 once finished
	var visit func(c *card) bool
	visit = func(c *card) bool {
		switch state[c.id] {
		case 1:
			return false
		case 2:
			return true
		}
		state[c.id] = 1
		for _, d := range c.depends {
			if dep := m.card(d); dep != nil && !visit(dep) {
				return false
			}
		}
		state[c.id] = 2
		return true
	}
	for _, c := range m.cards {
		if !visit(c) {
			m.problem(c.line, "the dependencies of %s form a cycle", c.id)
			return
		}
	}
}

// checkLinks reports a relative link whose file or heading is missing.
func (m *milestone) checkLinks(root string) {
	for _, l := range m.links {
		if strings.Contains(l.target, "://") || strings.HasPrefix(l.target, "mailto:") {
			continue
		}
		file, anchor, _ := strings.Cut(l.target, "#")
		target := filepath.Join(root, filepath.FromSlash(m.path))
		if file != "" {
			target = filepath.Join(root, filepath.FromSlash(path.Join(path.Dir(m.path), file)))
		}
		info, err := os.Stat(target)
		if err != nil {
			m.problem(l.line, "the link %s points to a missing file", l.target)
			continue
		}
		if anchor == "" || info.IsDir() || !strings.HasSuffix(target, ".md") {
			continue
		}
		if !slices.Contains(anchors(target), anchor) {
			m.problem(l.line, "the link %s points to a missing heading", l.target)
		}
	}
}

var anchorCache = map[string][]string{}

// anchors returns the anchors that GitHub gives the headings of a Markdown file.
func anchors(file string) []string {
	if a, ok := anchorCache[file]; ok {
		return a
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil
	}
	var list []string
	seen := map[string]int{}
	fenced := false
	for line := range strings.Lines(string(data)) {
		if strings.HasPrefix(line, "```") {
			fenced = !fenced
		}
		level := len(line) - len(strings.TrimLeft(line, "#"))
		if fenced || level == 0 || level > 6 || !strings.HasPrefix(line[level:], " ") {
			continue
		}
		s := slug(strings.TrimSpace(line[level:]))
		if n := seen[s]; n > 0 {
			list = append(list, s+"-"+strconv.Itoa(n))
		} else {
			list = append(list, s)
		}
		seen[s]++
	}
	anchorCache[file] = list
	return list
}

func slug(heading string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(heading) {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r), unicode.IsMark(r), r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	return b.String()
}

// currentMilestone returns the first milestone that is not done, or nil.
func currentMilestone(milestones []*milestone) *milestone {
	for _, m := range milestones {
		if m.status != "done" && m.status != "open" {
			return m
		}
	}
	return nil
}

func next(out *strings.Builder, milestones []*milestone) {
	m := currentMilestone(milestones)
	if m == nil {
		fmt.Fprintln(out, "Every milestone is done. Plan the next one, as \"Make a milestone ready\" in AGENTS.md says.")
		return
	}
	if m.status == "planned" && (m.open != nil || len(m.cards) == 0) {
		fmt.Fprintf(out, "Next: make %s %s ready\nFile: %s\n\n", m.id, m.title, m.path)
		fmt.Fprintln(out, "The milestone is planned. Settle its open questions in the spec, finish its cards, and set its status to ready, in one pull request, as \"Make a milestone ready\" in AGENTS.md says.")
		if len(m.open) > 0 {
			fmt.Fprintf(out, "\nOpen questions:\n%s\n", strings.Join(m.open, "\n"))
		}
		return
	}
	var agent, pascal []*card
	for _, c := range m.cards {
		waiting := slices.ContainsFunc(c.depends, func(d string) bool { return m.card(d).status != "done" })
		if c.status != "todo" || waiting {
			continue
		}
		if c.owner == "agent" {
			agent = append(agent, c)
		} else {
			pascal = append(pascal, c)
		}
	}
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

func body(c *card) string {
	return strings.TrimSpace(strings.Join(c.body, "\n"))
}

func status(out *strings.Builder, milestones []*milestone) {
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
