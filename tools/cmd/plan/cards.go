package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func (m *item) parseBody(text string) {
	section := ""
	var current *card
	fenced := false
	for i, line := range strings.Split(text, "\n") {
		n := i + 1
		if checkbox.MatchString(line) {
			m.problem(n, "a checkbox belongs in no item file: make the work a card")
		}
		if strings.HasPrefix(line, "```") {
			fenced = !fenced
		}
		switch {
		case fenced || strings.HasPrefix(line, "```"):
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
		case section == "Cards" && current == nil && strings.HasPrefix(line, "|"):
			m.row(n, line)
			continue
		}
		switch {
		case current != nil:
			current.body = append(current.body, line)
		case section == "Demo":
			m.demo = append(m.demo, line)
		case section == "Open questions" && strings.TrimSpace(line) != "":
			m.open = append(m.open, line)
		}
	}
}

// row reads one line of the cards table.
func (m *item) row(n int, line string) {
	cells := tableCells(line)
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
func (m *item) section(n int, line string) *card {
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

func (m *item) card(id string) *card {
	for _, c := range m.cards {
		if c.id == id {
			return c
		}
	}
	return nil
}

func (m *item) problem(line int, format string, args ...any) {
	where := m.path
	if line > 0 {
		where += ":" + strconv.Itoa(line)
	}
	m.problems = append(m.problems, where+": "+fmt.Sprintf(format, args...))
}

func (m *item) done() int {
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

func (m *item) checkCard(root string, c *card, ids map[string]bool) {
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

func (m *item) checkCycles() {
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
func (m *item) checkLinks(root string) {
	for _, l := range m.links {
		if why := brokenLink(root, m.path, l.target); why != "" {
			m.problem(l.line, "%s", why)
		}
	}
}
