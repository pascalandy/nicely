package cli

import (
	"context"
	"io"
	"os"

	"github.com/pascalandy/nicely/sdk/command"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/tui"
)

// confirmation is the confirmation that a change to a stored key or a
// grant needs: CONFIRMATION_REQUIRED with required and the hint without a
// terminal, and the [y/N] prompt of sdk/tui in one, on out, where any
// answer but yes changes nothing, with declined. next holds the hint's
// words when the hint's command runs as is.
type confirmation struct {
	question, required, declined, hint string
	next                               []string
	// facts are what the answer rests on, shown under the question.
	facts []tui.Fact
}

// ask returns nil once the human said yes, and the answer of the command
// otherwise.
func (c confirmation) ask(ctx context.Context, p *i18n.Printer, g command.Globals, human bool, out io.Writer) error {
	problem := contract.Problem{Code: contract.ConfirmationRequired, Message: c.required, Hint: c.hint, Next: c.next}
	if !human {
		return command.Fail(problem)
	}
	yes, err := tui.Confirm(ctx, os.Stdin, out, g.NoColor, tui.Confirmation{Question: c.question, Facts: c.facts, Choices: p.T("confirm.choices"), Yes: p.T("confirm.yes")})
	switch {
	case ctx.Err() != nil:
		return stopped(ctx)
	case err != nil:
		return err
	case !yes:
		problem.Message = c.declined
		return command.Fail(problem)
	}
	return nil
}

// pathFacts shows the source of a grant as a fact of its confirmation. A
// bundled extension has no path to show.
func pathFacts(p *i18n.Printer, path *string) []tui.Fact {
	if path == nil {
		return nil
	}
	return []tui.Fact{{Label: p.T("auth.fact_path"), Value: *path}}
}

// replacedFacts shows the new and the old source of a grant that moves to
// another source. A bundled extension's source has no path, so it shows
// as bundled.
func replacedFacts(p *i18n.Printer, path, old *string) []tui.Fact {
	shown := func(path *string) string {
		if path == nil {
			return p.T("auth.fact_bundled")
		}
		return *path
	}
	return []tui.Fact{
		{Label: p.T("auth.fact_new_path"), Value: shown(path)},
		{Label: p.T("auth.fact_old_path"), Value: shown(old)},
	}
}
