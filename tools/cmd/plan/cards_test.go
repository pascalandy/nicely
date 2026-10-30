package main

import (
	"strconv"
	"strings"
	"testing"
)

// TestCheckAcceptsManyCards keeps an item whole when its feature needs more
// than five cards, as COR-103 does.
func TestCheckAcceptsManyCards(t *testing.T) {
	root := repository(t, map[string]string{"XT-101-host.md": manyCards(15)})
	var out strings.Builder
	if err := run([]string{"check", root}, &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
}

func manyCards(n int) string {
	text := "---\nid: XT-101\nextension: Host\nstate: Ready\n---\n\n# XT-101 Host\n\n## Demo\n\nIt grows.\n\n## Cards\n\n| Card | Title | Owner | Depends on | Status |\n|---|---|---|---|---|\n"
	var sections string
	for i := 1; i <= n; i++ {
		id := "XT-101-T" + strconv.Itoa(i)
		text += "| " + id + " | Card | agent | — | todo |\n"
		sections += "\n### " + id + " Card\n\n- **Read:** [Demo](#demo)\n- **Proves:** `x.txtar`\n"
	}
	return text + sections
}
