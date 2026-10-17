package tui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// The styles of a confirmation, as Modes in contract.md describes it.
var (
	questionMark = ansi.Style{}.ForegroundColor(ansi.Cyan)
	question     = ansi.Style{}.Bold()
	choices      = ansi.Style{}.ForegroundColor(ansi.BrightBlack)
)

// Confirmation is a question that a human answers by typing a line, never
// by picking a button, with its text in the active language.
type Confirmation struct {
	Question string
	// Facts are what the answer rests on, each on a labeled line under the
	// question, such as the path of an extension that a grant names.
	Facts []Fact
	// Choices shows the answers and the default, such as [y/N].
	Choices string
	// Yes holds the answers that confirm, separated by commas, such as
	// "y, yes".
	Yes string
}

// Fact is one fact of a confirmation: a label, such as Path, and its value.
type Fact struct {
	Label, Value string
}

// Confirm asks c on out, then reads the answer from in, a terminal whose
// echo leaves the question and the answer on screen. It reports whether the
// answer is one of the words of c.Yes. Any other answer, an empty one, and
// the end of input mean no. The end of ctx, such as by Ctrl-C, returns its
// error.
func Confirm(ctx context.Context, in io.Reader, out io.Writer, noColor bool, c Confirmation) (bool, error) {
	yes := yesWords(c.Yes)
	if _, err := io.WriteString(Output(out, noColor), prompt(c)); err != nil {
		return false, err
	}
	line, err := readLine(ctx, in)
	if err != nil {
		// No Enter ended the line, so what follows starts its own.
		_, _ = io.WriteString(out, "\n")
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		return false, err
	}
	return slices.Contains(yes, strings.ToLower(strings.TrimSpace(line))), nil
}

// prompt lays out c as Modes in contract.md describes: ? and the question,
// then [y/N] at the end of its line. Facts go under the question, indented
// by two spaces, with their values aligned, and [y/N] then takes a line of
// its own, indented as they are.
func prompt(c Confirmation) string {
	head := questionMark.Styled("?") + " " + question.Styled(c.Question)
	if len(c.Facts) == 0 {
		return head + " " + choices.Styled(c.Choices) + " "
	}
	width := 0
	for _, f := range c.Facts {
		width = max(width, ansi.StringWidth(f.Label))
	}
	var b strings.Builder
	b.WriteString(head + "\n")
	for _, f := range c.Facts {
		b.WriteString("  " + muted.Styled(f.Label) + strings.Repeat(" ", width-ansi.StringWidth(f.Label)+2) + f.Value + "\n")
	}
	b.WriteString("  " + choices.Styled(c.Choices) + " ")
	return b.String()
}

// yesWords splits the comma-separated words that confirm.
func yesWords(list string) []string {
	var words []string
	for word := range strings.SplitSeq(list, ",") {
		if word = strings.ToLower(strings.TrimSpace(word)); word != "" {
			words = append(words, word)
		}
	}
	if len(words) == 0 {
		panic(fmt.Sprintf("tui: no word confirms in %q", list))
	}
	return words
}

// readLine reads one line from in without its line ending, or returns the
// error of ctx when ctx ends first. A signal ends the wait at once: the read
// that it abandons writes nothing, and the process ends right after.
func readLine(ctx context.Context, in io.Reader) (string, error) {
	type answer struct {
		line string
		err  error
	}
	done := make(chan answer, 1)
	go func() {
		line, err := bufio.NewReader(in).ReadString('\n')
		if errors.Is(err, io.EOF) && line != "" {
			err = nil
		}
		done <- answer{strings.TrimRight(line, "\r\n"), err}
	}()
	select {
	case a := <-done:
		return a.line, a.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
