package hello

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/wrap"
)

// count counts the lines of a file with the worker count.sh, which bash
// runs through sdk/wrap. It answers with the number that the worker prints
// as JSON, and checks the file before the worker starts.
func count(ctx context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	file := call.Args[0]
	data := map[string]any{"File": file}
	timeout := 10 * time.Second
	if value, given := call.Flags["timeout"]; given {
		d, err := time.ParseDuration(value)
		if err != nil || d <= 0 {
			data["Timeout"] = value
			return usage(p.T("hello.count.bad_timeout", data), "ncly hello count "+file+" --timeout 30s")
		}
		timeout = d
	}
	if info, err := os.Stat(file); err != nil || !info.Mode().IsRegular() {
		return usage(p.T("hello.count.no_file", data), "ncly hello count notes.md")
	}
	worker, problem := call.Worker("count.sh")
	if problem != nil {
		return contract.Outcome{Errors: []contract.Problem{*problem}}
	}
	program := wrap.Program{Name: "bash", Args: []string{worker, file}, Timeout: timeout}
	return wrap.Run(ctx, call, program, func(end wrap.Ended) contract.Outcome {
		var answer struct {
			Lines *int `json:"lines"`
		}
		switch {
		case end.Exit != 0:
			return end.Failed()
		case json.Unmarshal(end.Stdout, &answer) != nil || answer.Lines == nil || *answer.Lines < 0:
			return end.Malformed()
		}
		data["Lines"] = p.Number(float64(*answer.Lines), 0)
		_, _ = fmt.Fprintln(call.Stdout, p.T("hello.count.lines", data))
		return contract.Outcome{Data: map[string]int{"lines": *answer.Lines}}
	})
}

// usage returns the USAGE_INVALID outcome of a message, with the command
// that fixes it as its hint.
func usage(message, hint string) contract.Outcome {
	return contract.Outcome{Errors: []contract.Problem{{Code: contract.UsageInvalid, Message: message, Hint: hint}}}
}
