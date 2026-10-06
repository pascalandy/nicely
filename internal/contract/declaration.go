package contract

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
}

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
