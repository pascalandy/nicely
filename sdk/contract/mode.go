package contract

// Mode says whether ncly may ask a human questions.
type Mode int

const (
	NonInteractive Mode = iota
	Interactive
)

// DetectMode applies the Modes section of contract.md. noInput and json are
// the resolved values of --no-input and --json, each flag over its
// variable: machine output never waits for a human. A set CI variable, even
// empty, means a machine runs ncly.
func DetectMode(stdinTerminal, stdoutTerminal, noInput, json bool, lookupEnv func(string) (string, bool)) Mode {
	if _, ci := lookupEnv("CI"); ci || noInput || json || !stdinTerminal || !stdoutTerminal {
		return NonInteractive
	}
	return Interactive
}
