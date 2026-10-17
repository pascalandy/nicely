package command

import (
	"errors"
	"testing"

	"github.com/pascalandy/nicely/sdk/i18n"
)

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
