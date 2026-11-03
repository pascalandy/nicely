// Command guard runs one call of a scenario for the test kit, which builds
// it from this source, as it builds the fake child. It runs only on Linux,
// where it starts the call as its child and supervises it:
//
//   - It adopts each orphan of the call as a subreaper. After the call
//     exits, it waits up to two seconds for each descendant to exit. Then it
//     records each descendant that still runs, and kills it
//   - Its seccomp filter sends each link and rename of the call and of its
//     descendants to the guard, which reads the target path. The guard
//     itself and the kit stay outside the filter
//   - It records each such operation whose target is in the scope folder,
//     outside the skipped folders. It makes that operation itself, so it
//     knows whether the operation succeeded. With -refuse, it refuses that
//     operation there, as a file system without that ability does
//
// The usage is guard -report <file> -scope <folder> [-skip <folder>]...
// [-refuse <operation>] -- <program> <arg>... It exits with the exit code of
// the call, or 128 plus the number of the signal that ended it. It writes
// the report as one JSON object, when the call and its descendants have
// ended. It exits with 125 when it cannot supervise the call, and the report
// gives the reason.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Report is what the guard writes for the kit. extensiontest reads the same
// keys.
type Report struct {
	Exit       int         `json:"exit"`
	Error      string      `json:"error,omitempty"`
	Operations []Operation `json:"operations"`
	Children   []string    `json:"children"`
}

// Operation is one link or rename of the call, whose target is in the scope.
type Operation struct {
	// Name is link, rename-noreplace, or rename, as Publication declaration
	// in the contract names them.
	Name   string `json:"operation"`
	Target string `json:"target"`
	// Refused says that the guard refused it, and Done that it succeeded.
	Refused bool `json:"refused"`
	Done    bool `json:"done"`
}

// options are the flags of the guard.
type options struct {
	report, scope, refuse string
	skip                  []string
	stage                 bool
	call                  []string
}

func main() {
	o, err := parse(os.Args[1:])
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "guard:", err)
		os.Exit(125)
	}
	if o.stage {
		os.Exit(stage(o.call))
	}
	r := supervise(o)
	if r.Error != "" {
		r.Exit = 125
	}
	data, err := json.Marshal(r)
	if err == nil {
		err = os.WriteFile(o.report, append(data, '\n'), 0o644)
	}
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "guard:", err)
		os.Exit(125)
	}
	os.Exit(r.Exit)
}

// parse reads the flags, and the call after --.
func parse(args []string) (options, error) {
	var o options
	flags := flag.NewFlagSet("guard", flag.ContinueOnError)
	flags.StringVar(&o.report, "report", "", "the file of the report")
	flags.StringVar(&o.scope, "scope", "", "the folder whose operations the guard records")
	flags.StringVar(&o.refuse, "refuse", "", "the operation to refuse in the scope")
	flags.Func("skip", "a folder of the scope whose operations the guard ignores", func(dir string) error {
		o.skip = append(o.skip, dir)
		return nil
	})
	flags.BoolVar(&o.stage, "stage", false, "install the filter, then run the call; the guard runs this stage itself")
	if err := flags.Parse(args); err != nil {
		return o, err
	}
	o.call = flags.Args()
	switch {
	case len(o.call) == 0:
		return o, errors.New("give the call after --")
	case o.stage:
	case o.report == "" || o.scope == "":
		return o, errors.New("give -report and -scope")
	case o.refuse != "" && !strings.Contains(" link rename-noreplace rename ", " "+o.refuse+" "):
		return o, fmt.Errorf("unknown operation %q: give link, rename-noreplace, or rename", o.refuse)
	}
	return o, nil
}
