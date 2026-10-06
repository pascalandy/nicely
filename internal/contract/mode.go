package contract

// Mode says whether ncly may ask a human questions.
type Mode int

const (
	NonInteractive Mode = iota
	Interactive
)

// DetectMode applies the Modes section of cli-spec.md. noInput is the
// resolved --no-input flag or NCLY_NO_INPUT. A set CI variable, even empty,
// means a machine runs ncly.
func DetectMode(stdinTerminal, stdoutTerminal, noInput bool, lookupEnv func(string) (string, bool)) Mode {
	if _, ci := lookupEnv("CI"); ci || noInput || !stdinTerminal || !stdoutTerminal {
		return NonInteractive
	}
	return Interactive
}
