package main

import (
	"fmt"
	"strings"
)

func nextItem(out *strings.Builder, b *backlog) {
	if len(b.queue) == 0 {
		fmt.Fprintln(out, "The queue is empty. Plan the next item, as \"Make an item ready\" in docs/backlog/README.md says.")
		return
	}
	var m *item
	for _, candidate := range b.items {
		if candidate.id == b.queue[0] {
			m = candidate
			break
		}
	}
	if m.status == "Planned" && (m.open != nil || len(m.cards) == 0) {
		fmt.Fprintf(out, "Next: make %s %s ready\nFile: %s\n\n", m.id, m.title, m.path)
		fmt.Fprintf(out, "The item is Planned. Settle its open questions in the spec, finish its cards, and set state to Ready. See \"Make an item ready\" in docs/backlog/README.md. %s\n", readyRule)
		if len(m.open) > 0 {
			fmt.Fprintf(out, "\nOpen questions:\n%s\n", strings.Join(m.open, "\n"))
		}
		return
	}
	agent, pascal := startable(m)
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
	fmt.Fprintf(out, "Next card: %s %s\nItem: %s %s\nFile: %s\n\n%s\n\n", c.id, c.title, m.id, m.title, m.path, body(c))
	fmt.Fprintln(out, "Follow \"Work on a card\" in AGENTS.md. The pull request sets this card to done.")
}

func itemStatus(out *strings.Builder, items []*item) {
	for _, kind := range []struct{ prefix, column string }{{"XT-", "Extension"}, {"COR-", "Core"}} {
		fmt.Fprintln(out, strings.TrimSuffix(kind.prefix, "-"))
		rows := [][]string{{"ID", kind.column, "State"}}
		for _, m := range items {
			if strings.HasPrefix(m.id, kind.prefix) {
				rows = append(rows, []string{m.id, m.title, m.state})
			}
		}
		out.WriteString(table("", rows))
		if kind.prefix == "XT-" {
			fmt.Fprintln(out)
		}
	}
}
