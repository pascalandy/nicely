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
	// FlagGroups bounds how many flags of each group a call may set.
	FlagGroups []FlagGroup
	// Examples holds full command lines. Commands are never translated.
	Examples []string
	// Effects holds the effects that the command declares. A nil Effects
	// declares none, which permits no assumption: the command may have
	// every effect.
	Effects  *Effects
	Modes    Modes
	Requires Requires
	// Config holds the config keys that the command reads, as dotted paths
	// under its domain, such as agent.profiles, which covers every key
	// beneath it.
	Config []string
	Result Result
	// Doctor marks the status command of a domain, which doctor runs to
	// collect the extension's own checks, from M05.
	Doctor bool
}

// Result describes what a command answers.
type Result struct {
	Kind ResultKind
	// Schema is the JSON Schema of the data keys of a normal answer, an
	// object schema whose references are resolved, or nil when the
	// declaration gives none. AnswerSchema adds the envelope.
	Schema map[string]any
	// DryRunSchema is the same for the answer of a dry run.
	DryRunSchema map[string]any
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

// Type is the type of the value of an argument or a flag.
type Type string

const (
	TypeString  Type = "string"
	TypeBoolean Type = "boolean"
	TypeInteger Type = "integer"
	TypeNumber  Type = "number"
)

// Types lists the types of values, in the order of the spec.
var Types = []Type{TypeString, TypeBoolean, TypeInteger, TypeNumber}

// Arg is one positional argument.
type Arg struct {
	Name    string
	Summary string
	// Type is the type of the value. An empty Type means TypeString.
	Type     Type
	Required bool
	// Values lists the only accepted values, which completion also offers.
	// An empty list accepts any value.
	Values []string
	// Variadic takes every remaining word. Only the last argument may be
	// variadic.
	Variadic bool
	// ValuesFrom names the list command whose answer completion suggests
	// values from, or is nil.
	ValuesFrom *ValuesFrom
}

// ValuesFrom names where completion finds the values that only an extension
// knows, as Completion values in contract.md defines: the path of a list
// command, the key of its answer that holds an array of objects, and the
// string key of each object to suggest. It only suggests: the command still
// accepts any value.
type ValuesFrom struct {
	Command []string
	Key     string
	Field   string
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
	// Type is the type of the value. An empty Type means TypeBoolean for a
	// switch and TypeString otherwise.
	Type     Type
	Required bool
	// Variadic takes several values, one per occurrence of the flag.
	Variadic bool
	// Values lists the only accepted values. An empty list accepts any value.
	Values []string
	// Default is the typed default, or nil when none applies. A switch
	// defaults to false.
	Default any
	// ValuesFrom names the list command whose answer completion suggests
	// values from, or is nil.
	ValuesFrom *ValuesFrom
}

// FlagGroup bounds how many of its flags, by long name, a call may set.
type FlagGroup struct {
	Flags    []string
	Min, Max int
}

// Effects lists every effect a command can have.
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

// Requires names the programs and the services whose keys a command may
// need.
type Requires struct {
	// Bins holds names of programs, each a key of the manifest's programs.
	Bins []string
	// Keys holds service names, as Keys in contract.md defines them.
	Keys []string
}

// DryRunFlag and RunIDFlag are the flags that a command accepts when its
// modes set dry_run or recorded, as Global flags in contract.md defines.
var (
	DryRunFlag = Flag{Name: "dry-run", Shorthand: "n", Summary: "flag.dry_run", Type: TypeBoolean, Default: false}
	RunIDFlag  = Flag{Name: "run-id", Summary: "flag.run_id", Value: "id", Type: TypeString}
)

// ArgType returns the type of an argument's value.
func (a Arg) ArgType() Type {
	if a.Type == "" {
		return TypeString
	}
	return a.Type
}

// FlagType returns the type of a flag's value.
func (f Flag) FlagType() Type {
	switch {
	case f.Type != "":
		return f.Type
	case f.Value == "":
		return TypeBoolean
	}
	return TypeString
}

// FlagDefault returns the typed default of a flag: its declared one, false
// for a switch without one, and nil otherwise.
func (f Flag) FlagDefault() any {
	if f.Default == nil && f.Value == "" {
		return false
	}
	return f.Default
}

// ModeFlags returns the flags that the modes of a command add after its
// own: --dry-run for dry_run, then --run-id for recorded.
func (c Command) ModeFlags() []Flag {
	var flags []Flag
	if c.Modes.DryRun {
		flags = append(flags, DryRunFlag)
	}
	if c.Modes.Recorded {
		flags = append(flags, RunIDFlag)
	}
	return flags
}

// DeclaredEffects returns the effects of a command, every effect when it
// declares none, as Command descriptions in contract.md permits no
// assumption.
func (c Command) DeclaredEffects() Effects {
	if c.Effects == nil {
		return Effects{UserWrites: true, Network: true, Paid: true}
	}
	return *c.Effects
}
