package contract

// GlobalFlags declares the flags that every command accepts.
var GlobalFlags = []Flag{
	{Name: "help", Shorthand: "h", Summary: "flag.help"},
	{Name: "json", Summary: "flag.json", Env: "NCLY_JSON"},
	{Name: "no-input", Summary: "flag.no_input", Env: "NCLY_NO_INPUT"},
	{Name: "no-color", Summary: "flag.no_color", Env: "NO_COLOR"},
	{Name: "lang", Summary: "flag.lang", Value: "tag", Env: "NCLY_LANG"},
	{Name: "verbose", Shorthand: "v", Summary: "flag.verbose", Env: "NCLY_VERBOSE"},
	{Name: "version", Summary: "flag.version"},
}

// Command declares one command once, for help, completion, validation, and
// discovery. Text fields hold catalog IDs, never text. The declaration does
// not describe steps: the domain owns its behavior.
type Command struct {
	// Path holds the words after ncly, such as {"completion"}.
	Path    []string
	Summary string
	Args    []Arg
	Flags   []Flag
	// Examples holds full command lines. Commands are never translated.
	Examples []string
	Effects  Effects
	Modes    Modes
	Result   Result
}

// Result describes what a command answers.
type Result struct {
	Kind ResultKind
}

// ResultKind tells how a command answers, and so how core checks the answer
// of an extension's command. A manifest that names no kind gets ResultObject.
type ResultKind string

const (
	ResultObject ResultKind = "object"
	ResultBatch  ResultKind = "batch"
	ResultReport ResultKind = "report"
	ResultText   ResultKind = "text"
)

// Arg is one positional argument.
type Arg struct {
	Name    string
	Summary string
	// Values lists the only accepted values, which completion also offers.
	// An empty list accepts any value.
	Values []string
}

// Flag is one flag. Env names the variable with the same effect, for a
// global flag.
type Flag struct {
	Name      string
	Shorthand string
	Summary   string
	// Value names the value, such as tag. A flag without a value is a switch.
	Value string
	Env   string
}

// Effects lists every effect a command can have. A command that declares
// none changes nothing and calls no network.
type Effects struct {
	UserWrites bool
	Network    bool
	Paid       bool
}

// Modes lists the execution modes a command supports. A mode it does not
// declare is unsupported.
type Modes struct {
	DryRun   bool
	Recorded bool
	Resume   bool
}
