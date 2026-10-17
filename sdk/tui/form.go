package tui

import (
	"context"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// ErrInterrupted is the error of a form or a prompt that the human ended
// with Ctrl-C.
var ErrInterrupted = errors.New("tui: interrupted")

// Keys is the help of a form's keys, in the active language: the names of
// Enter and Escape, then what each key does.
type Keys struct {
	Enter, Esc, Submit string
	// Up, Down, and Filter are what ↑, ↓, and / do in a choice, and
	// SetFilter and ClearFilter what Escape does while the human filters
	// it and once a filter is set.
	Up, Down, Filter, SetFilter, ClearFilter string
	// Complete is what ctrl+e does in a text field that suggests values.
	Complete string
}

// Masked is the masked field of a key, with its text in the active
// language.
type Masked struct {
	Title       string
	Description string
	Keys        Keys
}

// AskMasked asks for a key in the masked field, a password field of the
// form's look, which shows one * per typed character. It reads the
// terminal in, draws on out, and returns the typed text. Ctrl-C returns
// ErrInterrupted, and the end of ctx returns its error, after the field
// has left the terminal as it found it.
func AskMasked(ctx context.Context, in *os.File, out io.Writer, noColor bool, m Masked) (string, error) {
	if dumb() {
		return askLine(ctx, in, out, m.Title+" ", true)
	}
	var typed string
	field := huh.NewInput().Title(m.Title).Description(m.Description).EchoMode(huh.EchoModePassword).Value(&typed)
	if err := ask(ctx, in, out, noColor, m.Keys, field); err != nil {
		return "", err
	}
	return typed, nil
}

// Text is a text field, with its text in the active language. It shows
// the first suggestion that starts with what the human typed, and accepts
// any other value.
type Text struct {
	Title       string
	Description string
	Suggestions []string
	Keys        Keys
}

// AskText asks for a value in a text field, as AskMasked asks for a key, and
// returns the typed text.
func AskText(ctx context.Context, in *os.File, out io.Writer, noColor bool, t Text) (string, error) {
	if dumb() {
		return askLine(ctx, in, out, t.Title+" ", false)
	}
	var typed string
	field := huh.NewInput().Title(t.Title).Description(t.Description).Suggestions(t.Suggestions).Value(&typed)
	if err := ask(ctx, in, out, noColor, t.Keys, field); err != nil {
		return "", err
	}
	return typed, nil
}

// Choice is a choice among options, with its text in the active language.
type Choice struct {
	Title       string
	Description string
	Options     []Option
	Keys        Keys
	// Number asks for the number of an option, on a terminal whose TERM is
	// dumb, which lists the options by number.
	Number string
}

// Option is one option of a choice: what it shows, and the value that it
// stands for.
type Option struct {
	Label, Value string
}

// AskChoice asks the human to pick one option of a choice, as AskMasked asks
// for a key, and returns its value. On a terminal whose TERM is dumb, it
// lists the options by number, and an answer that names none of them, the
// end of input included, returns an empty value.
func AskChoice(ctx context.Context, in *os.File, out io.Writer, noColor bool, c Choice) (string, error) {
	if dumb() {
		var b strings.Builder
		b.WriteString(c.Title + "\n")
		for i, o := range c.Options {
			b.WriteString(strconv.Itoa(i+1) + ". " + o.Label + "\n")
		}
		answer, err := askLine(ctx, in, out, b.String()+c.Number+" ", false)
		n, invalid := strconv.Atoi(strings.TrimSpace(answer))
		if err != nil || invalid != nil || n < 1 || n > len(c.Options) {
			return "", err
		}
		return c.Options[n-1].Value, nil
	}
	var chosen string
	options := make([]huh.Option[string], len(c.Options))
	for i, o := range c.Options {
		options[i] = huh.NewOption(o.Label, o.Value)
	}
	field := huh.NewSelect[string]().Title(c.Title).Description(c.Description).Options(options...).Value(&chosen)
	if err := ask(ctx, in, out, noColor, c.Keys, field); err != nil {
		return "", err
	}
	return chosen, nil
}

// ask runs a form of fields until the human submits it, presses Ctrl-C, or
// ctx ends, and leaves the terminal as it found it.
func ask(ctx context.Context, in *os.File, out io.Writer, noColor bool, keys Keys, fields ...huh.Field) error {
	fd := int(in.Fd())
	found, err := term.GetState(fd)
	if err != nil {
		return err
	}
	defer func() { _ = term.Restore(fd, found) }()
	err = draw(ctx, newForm(keys, fields...), in, out, Profile(out, noColor))
	switch {
	case ctx.Err() != nil:
		return ctx.Err()
	case errors.Is(err, tea.ErrInterrupted):
		return ErrInterrupted
	}
	return err
}

// newForm builds a form of fields in the look of Modes in contract.md.
func newForm(keys Keys, fields ...huh.Field) *huh.Form {
	keymap := huh.NewDefaultKeyMap()
	keymap.Input.Submit.SetHelp(keys.Enter, keys.Submit)
	keymap.Input.AcceptSuggestion.SetHelp("ctrl+e", keys.Complete)
	keymap.Select.Submit.SetHelp(keys.Enter, keys.Submit)
	keymap.Select.Up.SetHelp("↑", keys.Up)
	keymap.Select.Down.SetHelp("↓", keys.Down)
	keymap.Select.Filter.SetHelp("/", keys.Filter)
	keymap.Select.SetFilter.SetHelp(keys.Esc, keys.SetFilter)
	keymap.Select.ClearFilter.SetHelp(keys.Esc, keys.ClearFilter)
	return huh.NewForm(huh.NewGroup(fields...)).WithTheme(huh.ThemeFunc(theme)).WithKeyMap(keymap)
}

// theme is Huh's Base16 theme, whose key help is bright black instead of
// Bubbles' grays, which the 16 colors of the terminal do not hold, and whose
// options that the human has not chosen keep the terminal's color instead
// of white, which a light terminal cannot show.
func theme(bool) *huh.Styles {
	t := huh.ThemeBase16(true)
	help := lipgloss.NewStyle().Foreground(lipgloss.BrightBlack)
	t.Help.ShortKey, t.Help.ShortDesc, t.Help.ShortSeparator = help, help, help
	t.Help.FullKey, t.Help.FullDesc, t.Help.FullSeparator = help, help, help
	t.Help.Ellipsis = help
	for _, field := range []*huh.FieldStyles{&t.Focused, &t.Blurred} {
		field.Option = field.Option.UnsetForeground()
		field.UnselectedOption = field.UnselectedOption.UnsetForeground()
	}
	return t
}

// draw runs form as a Bubble Tea program that reads in and draws on out,
// then erases it, so the form leaves the screen once it ends. The program
// takes profile, so Bubble Tea never detects one, which runs tmux info inside
// tmux. ncly's own handler gets each signal and ends ctx, which ends the
// program.
func draw(ctx context.Context, form *huh.Form, in *os.File, out io.Writer, profile colorprofile.Profile) error {
	form.SubmitCmd, form.CancelCmd = tea.Quit, tea.Interrupt
	f := &frame{form: form}
	_, err := tea.NewProgram(f, tea.WithContext(ctx), tea.WithInput(in), tea.WithOutput(out),
		tea.WithColorProfile(profile), tea.WithoutSignalHandler()).Run()
	// Bubble Tea leaves the cursor on the last line of the frame, which it
	// erased, and the lines above it stay to erase.
	if above := strings.Count(f.last, "\n"); above > 0 {
		_, _ = io.WriteString(out, "\r"+ansi.CursorUp(above)+ansi.EraseScreenBelow)
	}
	return err
}

// frame shows a form, and keeps its last frame once the form ends, where
// Huh shows nothing. Bubble Tea v2.0.8 erases a frame that turns empty from
// the line of the cursor down, so the lines above the cursor would stay on
// screen. draw erases the kept frame instead.
type frame struct {
	form *huh.Form
	last string
}

func (f *frame) Init() tea.Cmd { return f.form.Init() }

func (f *frame) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := f.form.Update(msg)
	f.form = m.(*huh.Form)
	return f, cmd
}

func (f *frame) View() tea.View {
	if v := f.form.View(); v != "" {
		f.last = v
	}
	return tea.NewView(f.last)
}

// dumb reports whether TERM is dumb: the terminal cannot redraw a line,
// so a form asks on lines instead of drawing.
func dumb() bool { return os.Getenv("TERM") == "dumb" }

// askLine asks one field on a line, for a terminal whose TERM is dumb, as
// Huh's accessible mode lays it out: the prompt, then the answer on the
// same line. A masked field turns echo off before its prompt shows, so no
// typed character echoes, and drops what the human typed before the prompt.
// askLine returns the answer, an empty one at the end of input, or the
// error of ctx once ctx ends, and leaves the terminal as it found it.
func askLine(ctx context.Context, in *os.File, out io.Writer, prompt string, masked bool) (string, error) {
	fd := int(in.Fd())
	found, err := term.GetState(fd)
	if err != nil {
		return "", err
	}
	defer func() { _ = term.Restore(fd, found) }()
	if masked {
		if err := echoOff(fd); err != nil {
			return "", err
		}
	}
	if _, err := io.WriteString(out, prompt); err != nil {
		return "", err
	}
	line, err := readLine(ctx, in)
	if masked || err != nil {
		// The Enter of a masked answer does not echo, and no Enter ended
		// the line otherwise, so what follows starts its own line.
		_, _ = io.WriteString(out, "\n")
	}
	if errors.Is(err, io.EOF) {
		return "", nil
	}
	return line, err
}

// echoOff turns off the echo of the terminal fd, and drops the input that it
// holds, which the terminal echoed as the human typed it.
func echoOff(fd int) error {
	t, err := unix.IoctlGetTermios(fd, ioctlGetTermios)
	if err != nil {
		return err
	}
	t.Lflag &^= unix.ECHO
	return unix.IoctlSetTermios(fd, ioctlSetTermiosFlush, t)
}
