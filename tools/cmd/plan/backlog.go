package main

import (
	"cmp"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

const backlogDir = "docs/backlog"

var (
	itemFile         = regexp.MustCompile(`^((?:XT|COR)-\d{3})-[a-z]+(?:-[a-z]+){0,2}\.md$`)
	itemID           = regexp.MustCompile(`^(?:XT|COR)-[16]\d{2}$`)
	itemState        = regexp.MustCompile(`^(Idea|Parked|Planned|Ready|Active|Shipped)(?:[. ]|$)`)
	queueEntry       = regexp.MustCompile(`^\d+\. (XT-\d{3}|COR-\d{3})$`)
	frozenMilestones = []string{
		"M00-foundation.md", "M01-extension-host.md", "M02-sdk.md", "M03-describe.md",
		"M04-skill.md", "M05-doctor.md", "M06-auth.md", "M07-forms.md", "M08-grants.md",
		"M09-agent-profiles.md", "M10-agent-run.md", "M13-transcript.md",
		"M16-transcript-finish.md", "M17-inventory.md", "M18-extension-template.md", "M23-shared-domains.md",
	}
)

type backlog struct {
	items    []*item
	queue    []string
	problems []string
}

func loadBacklog(root string) (*backlog, error) {
	entries, err := os.ReadDir(filepath.Join(root, backlogDir))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	b := &backlog{}
	for _, e := range entries {
		if e.IsDir() || e.Name() == "README.md" || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, backlogDir, e.Name()))
		if err != nil {
			return nil, err
		}
		b.items = append(b.items, parseItem(path.Join(backlogDir, e.Name()), string(data)))
	}
	slices.SortFunc(b.items, func(a, b *item) int { return cmp.Compare(a.id, b.id) })
	data, err := os.ReadFile(filepath.Join(root, backlogDir, "README.md"))
	if err != nil {
		return nil, err
	}
	inQueue, found := false, false
	var f fence
	for i, line := range strings.Split(string(data), "\n") {
		if f.in(line) {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			inQueue = line == "## Queue"
			if inQueue && found {
				b.problems = append(b.problems, backlogDir+"/README.md: use one Queue section")
			}
			found = found || inQueue
			continue
		}
		if !inQueue || strings.TrimSpace(line) == "" {
			continue
		}
		if match := queueEntry.FindStringSubmatch(line); match != nil {
			b.queue = append(b.queue, match[1])
		} else {
			b.problems = append(b.problems, fmt.Sprintf("%s/README.md:%d: a queue entry must look like 1. COR-101", backlogDir, i+1))
		}
	}
	if !found {
		b.problems = append(b.problems, backlogDir+"/README.md: add a Queue section")
	}
	return b, nil
}

func parseItem(file, text string) *item {
	m := &item{path: file, text: text, links: links(text), refs: cardRefs(text)}
	if match := itemFile.FindStringSubmatch(path.Base(file)); match != nil {
		m.id = match[1]
	} else {
		m.problem(0, "the file name needs an XT or COR ID and 1 to 3 lowercase keywords")
	}
	if !itemID.MatchString(m.id) {
		m.problem(0, "the ID must be in the 1xx or 6xx band")
	}
	m.parseBody(text)
	lines := strings.Split(text, "\n")
	if len(lines) < 5 || lines[0] != "---" || lines[4] != "---" {
		m.problem(1, "frontmatter needs exactly three keys on separate lines, between --- lines")
		return m
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(strings.Join(lines[1:4], "\n")), &doc); err != nil {
		m.problem(1, "invalid YAML frontmatter: %v", err)
		return m
	}
	kind := "core"
	if strings.HasPrefix(m.id, "XT-") {
		kind = "extension"
	}
	want := []string{"id", kind, "state"}
	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode || len(doc.Content[0].Content) != 6 {
		m.problem(1, "frontmatter keys must be id, %s, state, in that order", kind)
		return m
	}
	fields := doc.Content[0].Content
	for i, key := range want {
		k, v := fields[2*i], fields[2*i+1]
		if k.Value != key || k.Line != i+1 || v.Line != i+1 || v.Kind != yaml.ScalarNode || v.Tag != "!!str" || strings.TrimSpace(v.Value) == "" {
			m.problem(i+2, "frontmatter keys must be id, %s, state, in that order, with nonempty strings on one line", kind)
			return m
		}
		if k.HeadComment+k.LineComment+k.FootComment+v.HeadComment+v.LineComment+v.FootComment != "" {
			m.problem(i+2, "quote frontmatter values that contain #; comments can hide state text")
			return m
		}
	}
	if fields[1].Value != m.id {
		m.problem(2, "frontmatter id must match the file ID %s", m.id)
	}
	m.kind, m.title, m.state = kind, fields[3].Value, fields[5].Value
	if match := itemState.FindStringSubmatch(m.state); match != nil {
		m.status = match[1]
	} else {
		m.problem(4, "state must start with Idea, Parked, Planned, Ready, Active, or Shipped")
	}
	for i, line := range lines[5:] {
		if strings.HasPrefix(line, "Status:") || strings.HasPrefix(line, "Version:") {
			m.problem(i+6, "keep status and version in frontmatter state, without separate body lines")
		}
	}
	return m
}

func (b *backlog) check(root string) []string {
	problems := slices.Clone(b.problems)
	known, byID := cardIDs(b.items), map[string]*item{}
	for _, m := range b.items {
		if other := byID[m.id]; other != nil {
			m.problem(0, "%s uses the same ID", other.path)
		}
		byID[m.id] = m
		m.checkItem(root)
		m.checkLinks(root)
		for _, r := range m.refs {
			if !known[r.target] {
				m.problem(r.line, "no item defines the card %s", r.target)
			}
		}
		problems = append(problems, m.problems...)
	}
	seen := map[string]bool{}
	for _, id := range b.queue {
		m := byID[id]
		switch {
		case seen[id]:
			problems = append(problems, backlogDir+"/README.md: duplicate queue entry "+id)
		case m == nil:
			problems = append(problems, backlogDir+"/README.md: queue names no item "+id)
		case m.status != "" && !queued(m.status):
			problems = append(problems, backlogDir+"/README.md: queue item "+id+" must be Planned, Ready, or Active")
		}
		seen[id] = true
	}
	for _, m := range b.items {
		if queued(m.status) && !seen[m.id] {
			problems = append(problems, m.path+": a Planned, Ready, or Active item must be in the queue")
		}
	}
	entries, err := os.ReadDir(filepath.Join(root, milestoneDir))
	if err != nil && !os.IsNotExist(err) {
		problems = append(problems, err.Error())
	}
	for _, e := range entries {
		if !e.IsDir() && !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if e.Name() != "README.md" && !slices.Contains(frozenMilestones, e.Name()) {
			problems = append(problems, milestoneDir+"/"+e.Name()+": only the 16 frozen milestone files may remain")
		}
	}
	return problems
}

func queued(status string) bool {
	return slices.Contains([]string{"Planned", "Ready", "Active"}, status)
}

func (m *item) checkItem(root string) {
	if m.status == "" {
		return
	}
	if !m.hasDemo {
		m.problem(0, "an item needs a Demo section")
	}
	if strings.Contains(m.id, "-6") && m.status != "Idea" {
		m.problem(4, "a 6xx item must have status Idea")
	}
	if strings.Contains(m.id, "-1") && m.status == "Idea" {
		m.problem(4, "Idea needs a 6xx ID")
	}
	agents, ids := 0, map[string]bool{}
	for _, c := range m.cards {
		m.checkCard(root, c, ids)
		if c.owner == "agent" {
			agents++
		}
	}
	m.checkCycles()
	if len(m.cards) > 0 && len(extractSentences(m.path, []byte(strings.Join(m.demo, "\n")))) != 1 {
		m.problem(0, "an item with cards needs a one-sentence Demo")
	}
	if agents > maxAgentCards {
		m.problem(0, "%d agent cards is more than %d: split the item", agents, maxAgentCards)
	}
	done, total := m.done(), len(m.cards)
	switch m.status {
	case "Idea":
		if total > 0 || m.hasCards {
			m.problem(0, "an Idea item has no cards")
		}
	case "Parked":
		if done > 0 {
			m.problem(0, "a Parked item has no started card")
		}
	case "Planned":
		if done > 0 || total == 0 && len(m.open) == 0 {
			m.problem(0, "a Planned item needs open questions or todo cards, and no done card")
		}
	case "Ready":
		if m.open != nil || total == 0 || done > 0 {
			m.problem(0, "a Ready item needs todo cards, no done card, and no open questions")
		}
	case "Active":
		if m.open != nil || done == 0 || done == total {
			m.problem(0, "an Active item needs a done card and a todo card, and no open questions")
		}
	case "Shipped":
		if m.open != nil || done < total {
			m.problem(0, "a Shipped item has no todo card or open questions")
		}
	}
}
