package command

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/charmbracelet/x/ansi"
	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
	"github.com/pascalandy/nicely/sdk/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Form is the form for missing values of Modes in contract.md: in
// interactive mode, it asks for the values that a call lacks before the
// command runs, then keeps the command that the call became, which
// PrintNext shows once the command has run. Core and every Go extension
// share it through Build.
type Form struct {
	p       *i18n.Printer
	in      *os.File
	out     io.Writer
	noColor bool
	// next is the equivalent command, once the form completed the call.
	next string
}

// NewForm returns the form of a command line that reads stdin and draws on
// stderr, or nil outside interactive mode, where a missing value fails.
func NewForm(p *i18n.Printer, g Globals, stdin, stdout *os.File, stderr io.Writer) *Form {
	mode := contract.DetectMode(term.IsTerminal(int(stdin.Fd())), term.IsTerminal(int(stdout.Fd())), g.NoInput, g.JSON, os.LookupEnv)
	if mode != contract.Interactive {
		return nil
	}
	return &Form{p: p, in: stdin, out: stderr, noColor: g.NoColor}
}

type formKey struct{}

// WithForm returns ctx with f, which each command that Build makes asks its
// missing values in. A nil f makes a missing value fail.
func WithForm(ctx context.Context, f *Form) context.Context {
	return context.WithValue(ctx, formKey{}, f)
}

// PrintNext prints, after a run that the form completed, the line Next
// time: and the equivalent command, unless a signal stopped the run, which
// code tells.
func (f *Form) PrintNext(code int) {
	if f == nil || f.next == "" || code == int(contract.ExitInterrupted) || code == int(contract.ExitTerminated) {
		return
	}
	_, _ = io.WriteString(tui.Output(f.out, f.noColor), tui.Muted(f.p.T("form.next_time"))+"\n  "+f.next+"\n")
}

// FormKeys returns the help of a form's keys in the language of p.
func FormKeys(p *i18n.Printer) tui.Keys {
	return tui.Keys{
		Enter:    p.T("form.key_enter"),
		Submit:   p.T("form.submit"),
		Up:       p.T("form.up"),
		Down:     p.T("form.down"),
		Filter:   p.T("form.filter"),
		Complete: p.T("form.complete"),
	}
}

// complete returns the arguments of a call that lacks no value. The form of
// the command line asks for each value that the call lacks, and the call
// is then checked as if it had held them. Without a form, or after an empty
// answer, a missing value fails. text renders the declaration's catalog IDs.
func complete(p, text *i18n.Printer, c *cobra.Command, d contract.Command, args []string) ([]string, error) {
	lacks := lacking(c, d, args)
	if len(lacks) == 0 {
		return args, nil
	}
	f, _ := c.Context().Value(formKey{}).(*Form)
	if f == nil {
		return nil, Fail(missing(p, c, d, args, lacks))
	}
	args, err := f.ask(text, c, d, args, lacks)
	if err != nil {
		return nil, err
	}
	if problem := invalid(p, c, d, args); problem != nil {
		return nil, Fail(*problem)
	}
	if lacks := lacking(c, d, args); len(lacks) > 0 {
		return nil, Fail(missing(p, c, d, args, lacks))
	}
	f.next = commandLine(c, d, args, nil)
	return args, nil
}

// ask asks for each value that a call lacks, one at a time, sets each flag
// that the human gives, and returns the arguments with those that the human
// gives. An empty answer stops the form, and leaves its value missing.
// Ctrl-C, or a signal that ends the form, returns a failure that the signal
// decides.
func (f *Form) ask(text *i18n.Printer, c *cobra.Command, d contract.Command, args []string, lacks []lack) ([]string, error) {
	ctx, done := onSignal()
	defer done()
	stopped := func(err error) error {
		if ctx.Err() == nil && !errors.Is(err, tui.ErrInterrupted) {
			return err
		}
		stop := contract.Interrupt
		if ctx.Err() != nil {
			stop = contract.StopSignal(runner.SignalOf(ctx))
		}
		return &Failure{contract.Outcome{Signal: stop}}
	}
	for _, l := range lacks {
		switch {
		case l.arg != nil:
			suggest := func() []cobra.Completion {
				if c.ValidArgsFunction == nil {
					return nil
				}
				values, _ := c.ValidArgsFunction(c, args, "")
				return values
			}
			value, err := f.value(ctx, text, l.arg.Name, l.arg.Summary, l.arg.Values, suggest)
			if err != nil || value == "" {
				return args, stopped(err)
			}
			args = append(args, value)
		case l.flag != nil:
			value, err := f.flagValue(ctx, text, c, d, *l.flag, args)
			if err != nil || value == "" {
				return args, stopped(err)
			}
			_ = c.Flags().Set(l.flag.Name, value)
		default:
			for setFlags(c, *l.group) < l.group.Min {
				name, err := f.pick(ctx, text, c, d, *l.group)
				if err != nil {
					return args, stopped(err)
				}
				i := slices.IndexFunc(d.Flags, func(flag contract.Flag) bool { return flag.Name == name })
				value := "true"
				if d.Flags[i].Value != "" {
					if value, err = f.flagValue(ctx, text, c, d, d.Flags[i], args); err != nil || value == "" {
						return args, stopped(err)
					}
				}
				_ = c.Flags().Set(name, value)
			}
		}
	}
	return args, nil
}

// flagValue asks for the value of a flag, which suggests what its
// completion suggests.
func (f *Form) flagValue(ctx context.Context, text *i18n.Printer, c *cobra.Command, d contract.Command, flag contract.Flag, args []string) (string, error) {
	suggest := func() []cobra.Completion {
		complete, found := c.GetFlagCompletionFunc(flag.Name)
		if !found {
			return nil
		}
		values, _ := complete(c, args, "")
		return values
	}
	return f.value(ctx, text, flagUsage(d, flag.Name), flag.Summary, flag.Values, suggest)
}

// value asks for one value, whose title is the argument's name or the flag
// and its value, and whose description is its summary: a choice among the
// values that its declaration lists, or else a text field that suggests
// what completion suggests.
func (f *Form) value(ctx context.Context, text *i18n.Printer, title, summary string, values []string, suggest func() []cobra.Completion) (string, error) {
	if len(values) > 0 {
		options := make([]tui.Option, len(values))
		for i, v := range values {
			options[i] = tui.Option{Label: v, Value: v}
		}
		return tui.AskChoice(ctx, f.in, f.out, f.noColor, tui.Choice{Title: title, Description: text.T(summary), Options: options, Keys: FormKeys(f.p)})
	}
	var suggestions []string
	for _, s := range suggest() {
		// A completion may add its description after a tab.
		value, _, _ := strings.Cut(s, "\t")
		suggestions = append(suggestions, value)
	}
	typed, err := tui.AskText(ctx, f.in, f.out, f.noColor, tui.Text{Title: title, Description: text.T(summary), Suggestions: suggestions, Keys: FormKeys(f.p)})
	return strings.TrimSpace(typed), err
}

// pick asks which flag of a group to set, among those that the call did not
// set, each with its value and its summary, aligned.
func (f *Form) pick(ctx context.Context, text *i18n.Printer, c *cobra.Command, d contract.Command, g contract.FlagGroup) (string, error) {
	var names []string
	width := 0
	for _, name := range g.Flags {
		if !c.Flags().Lookup(name).Changed {
			names = append(names, name)
			width = max(width, ansi.StringWidth(flagUsage(d, name)))
		}
	}
	options := make([]tui.Option, len(names))
	for i, name := range names {
		usage := flagUsage(d, name)
		j := slices.IndexFunc(d.Flags, func(flag contract.Flag) bool { return flag.Name == name })
		options[i] = tui.Option{Label: usage + strings.Repeat(" ", width-ansi.StringWidth(usage)+2) + text.T(d.Flags[j].Summary), Value: name}
	}
	return tui.AskChoice(ctx, f.in, f.out, f.noColor, tui.Choice{
		Title:       f.p.T("form.group_title"),
		Description: groupMessage(f.p, g, g.Min),
		Options:     options,
		Keys:        FormKeys(f.p),
	})
}

// onSignal returns a context that Ctrl-C or SIGTERM ends while a form asks,
// with the signal as its cause, as runner.OnSignal does, and a function that
// stops listening, so a later signal reaches the command as before.
func onSignal() (context.Context, func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() {
		select {
		case s := <-signals:
			cancel(runner.SignalCause{Signal: s})
		case <-ctx.Done():
		}
	}()
	return ctx, func() {
		signal.Stop(signals)
		cancel(nil)
	}
}

// commandLine writes a call as the Next time: line of Modes in contract.md
// shows it: ncly, the command path, each argument, then each flag that the
// call set, in the order of the declaration, each word quoted for a POSIX
// shell. Each value in lacks shows in the notation of the usage line, which
// makes the full command of a hint. --force never shows.
func commandLine(c *cobra.Command, d contract.Command, args []string, lacks []lack) string {
	lacks = slices.Clone(lacks)
	words := append([]string{"ncly"}, d.Path...)
	arguments := make([]string, 0, len(args))
	dashed := false
	for _, a := range args {
		arguments = append(arguments, quote(a))
		dashed = dashed || strings.HasPrefix(a, "-")
	}
	var flags []string
	for _, l := range lacks {
		if l.arg != nil {
			name := l.arg.Name
			if l.arg.Variadic {
				name += "..."
			}
			arguments = append(arguments, "<"+name+">")
		}
	}
	for _, f := range append(slices.Clone(d.Flags), d.ModeFlags()...) {
		flag := c.Flags().Lookup(f.Name)
		switch {
		case f.Name == "force":
		case flag.Changed && f.Value == "":
			if on, _ := strconv.ParseBool(flag.Value.String()); on {
				flags = append(flags, "--"+f.Name)
			} else {
				flags = append(flags, "--"+f.Name+"=false")
			}
		case flag.Changed:
			for _, v := range FlagValues(flag) {
				flags = append(flags, "--"+f.Name, quote(v))
			}
		case slices.ContainsFunc(lacks, func(l lack) bool { return l.flag != nil && l.flag.Name == f.Name }):
			flags = append(flags, flagUsage(d, f.Name))
		}
	}
	for _, l := range lacks {
		if l.group != nil {
			flags = append(flags, groupUsage(d, *l.group))
		}
	}
	if dashed {
		return strings.Join(slices.Concat(words, flags, []string{"--"}, arguments), " ")
	}
	return strings.Join(slices.Concat(words, arguments, flags), " ")
}

// plainWord is a word that a POSIX shell reads as is.
var plainWord = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./-]+$`)

// quote returns a word as a POSIX shell reads it back: as is when it needs
// no quotes, and between single quotes otherwise.
func quote(word string) string {
	if plainWord.MatchString(word) {
		return word
	}
	return "'" + strings.ReplaceAll(word, "'", `'\''`) + "'"
}
