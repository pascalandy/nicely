package i18n

import "testing"

func TestPseudoMarksTextAndKeepsTemplateActions(t *testing.T) {
	got := sdkCatalogs(t).Printer("en-XA").T("usage.unknown_command", map[string]any{"Name": "nope"})
	// The English text holds 28 characters, so a third more is 10 dots.
	want := `⟦Ûñķñöŵñ çöɱɱáñð "nope".··········⟧`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
