package hello

import (
	"context"
	"fmt"
	"os"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/wrap"
)

// scan reads a file with the worker scan.sh, which bash runs through
// sdk/wrap. The worker prints its progress on stdout, as an encoder does,
// and answers with its exit code alone. The decoder reads no stdout, so
// DiscardStdout drops it, and a long run never passes the output limit.
func scan(ctx context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	file := call.Args[0]
	data := map[string]any{"File": file}
	if info, err := os.Stat(file); err != nil || !info.Mode().IsRegular() {
		return usage(p.T("hello.scan.no_file", data), "ncly hello scan notes.md")
	}
	worker, problem := call.Worker("scan.sh")
	if problem != nil {
		return contract.Outcome{Errors: []contract.Problem{*problem}}
	}
	program := wrap.Program{Name: "bash", Args: []string{worker, file}, DiscardStdout: true}
	return wrap.Run(ctx, call, program, func(end wrap.Ended) contract.Outcome {
		if end.Exit != 0 {
			return end.Failed()
		}
		_, _ = fmt.Fprintln(call.Stdout, p.T("hello.scan.done", data))
		return contract.Outcome{Data: map[string]string{"file": file}}
	})
}
