package command

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/pascalandy/nicely/sdk/i18n"
)

// The prompt of a numbered choice takes the plural form of the number of
// its last option, formatted. English writes both forms the same, so the
// catalog of this test makes them differ.
func TestNumberPromptPluralForms(t *testing.T) {
	catalogs, err := i18n.Load(fstest.MapFS{"en.toml": {Data: []byte(`["form.number"]
description = "d"
one = "Type the number {{.Count}}:"
other = "Type a number from 1 to {{.Count}}:"
`)}})
	if err != nil {
		t.Fatal(err)
	}
	f := &Form{p: catalogs.Printer("en")}
	for count, want := range map[int]string{
		1:    "Type the number 1:",
		2:    "Type a number from 1 to 2:",
		1500: "Type a number from 1 to 1,500:",
	} {
		if got := f.number(count); got != want {
			t.Errorf("%d options: got %q, want %q", count, got, want)
		}
	}
}

// refusing is a stream whose reader is gone.
type refusing struct{}

func (refusing) Write([]byte) (int, error) { return 0, errors.New("refused") }

// A Next time: line that stderr refuses fails the run, as Output in
// contract.md asks of lost output, and a run without the form writes none.
func TestPrintNextReportsALostLine(t *testing.T) {
	catalogs, err := i18n.Load(i18n.Locales())
	if err != nil {
		t.Fatal(err)
	}
	f := &Form{p: catalogs.Printer("en"), out: refusing{}, noColor: true, next: "ncly hello pick Ada"}
	if f.PrintNext(0) == nil {
		t.Error("PrintNext reported no error for a line that stderr refused")
	}
	if f.PrintNext(130) != nil {
		t.Error("PrintNext wrote after Ctrl-C")
	}
	var none *Form
	if none.PrintNext(0) != nil {
		t.Error("PrintNext wrote without a form")
	}
}
