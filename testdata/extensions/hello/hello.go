// Package hello is the Go fixture extension that proves the SDK. It mirrors
// the shell fixtures of M01: an own code, an own warning, a report, a text
// result, and a command that runs a program until a signal arrives. It
// embeds its skill, as every extension ships one, and the workers that hello
// count and hello scan run through sdk/wrap.
package hello

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/runner"
)

//go:embed manifest.toml
var manifest []byte

//go:embed locales/*.toml
var locales embed.FS

//go:embed skill
var skill embed.FS

//go:embed worker
var resources embed.FS

// Extension is the hello fixture, for both transports.
var Extension = extension.Extension{
	Manifest:  manifest,
	Locales:   locales,
	Skill:     skill,
	Resources: resources,
	Handlers: map[string]extension.Handler{
		"hello count": count,
		"hello run":   run,
		"hello check": check,
		"hello say":   say,
		"hello wait":  wait,
		"hello pick":  pick,
		"hello scan":  scan,
	},
}

// greeting greets the name of the call with its --greeting, and reports
// whether the call named someone.
func greeting(call extension.Call) (string, bool) {
	name, named := call.Printer.T("hello.world"), len(call.Args) > 0
	if named {
		name = call.Args[0]
	}
	word, given := call.Flags["greeting"]
	if !given {
		word = call.Printer.T("hello.default_greeting")
	}
	return call.Printer.T("hello.greeting", map[string]any{"Greeting": word, "Name": name}), named
}

func run(_ context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	if word, given := call.Flags["greeting"]; given && word == "" {
		return contract.Outcome{Errors: []contract.Problem{{Code: contract.UsageInvalid, Message: p.T("hello.empty_greeting"), Hint: "ncly hello run --greeting Hi Ada"}}}
	}
	if call.Flags["grumpy"] == "true" {
		return contract.Outcome{Errors: []contract.Problem{{Code: "HELLO_GRUMPY", Message: p.T("hello.grumpy"), Hint: "ncly hello run Ada"}}}
	}
	text, named := greeting(call)
	if os.Getenv("NCLY_VERBOSE") == "1" {
		_, _ = fmt.Fprintln(call.Stderr, p.T("hello.verbose", map[string]any{"Greeting": text}))
	}
	_, _ = fmt.Fprintln(call.Stdout, text)
	o := contract.Outcome{Data: map[string]string{"greeting": text}}
	if !named {
		o.Warnings = []contract.Problem{{Code: "HELLO_UNNAMED", Message: p.T("hello.unnamed"), Hint: "ncly hello run Ada"}}
	}
	return o
}

func check(_ context.Context, call extension.Call) contract.Outcome {
	p := call.Printer
	if call.Flags["fail"] == "true" {
		return contract.Outcome{Errors: []contract.Problem{{Code: contract.ConfigInvalid, Message: p.T("hello.config"), Hint: "ncly hello check"}}}
	}
	_, _ = fmt.Fprintln(call.Stdout, p.T("hello.mood"))
	return contract.Outcome{
		Data:        map[string]any{"checks": []map[string]string{{"id": "hello.mood", "status": "fail"}}},
		HumanNeeded: true,
	}
}

func say(_ context.Context, call extension.Call) contract.Outcome {
	text, named := greeting(call)
	_, _ = fmt.Fprintln(call.Stdout, text)
	if !named {
		return contract.Outcome{Warnings: []contract.Problem{{Code: "HELLO_UNNAMED", Message: call.Printer.T("hello.unnamed"), Hint: "ncly hello say Ada"}}}
	}
	return contract.Outcome{}
}

// pick writes the file picked in the current folder, its execution marker,
// and answers with the call it received. A dry run writes nothing.
func pick(_ context.Context, call extension.Call) contract.Outcome {
	data := map[string]any{"args": call.Args, "flags": call.Flags, "lists": call.Lists}
	if call.Flags["dry-run"] == "true" {
		data["planned"] = "picked"
		return contract.Outcome{Data: data}
	}
	data["picked"] = "picked"
	if err := os.WriteFile("picked", []byte(call.Args[0]+"\n"), 0o644); err != nil {
		return contract.Outcome{Errors: []contract.Problem{{Code: contract.Runtime, Message: call.Printer.T("hello.program"), Hint: "ncly hello pick"}}}
	}
	return contract.Outcome{Data: data, Effect: contract.NonRepeatableEffect}
}

// program writes finished.txt, starts a grandchild in its own process group
// through the pgroup program of the scenarios, and runs until a signal ends
// it. It writes program.pid after the grandchild writes its ID, so a scenario
// that waits for program.pid finds both. Under STUBBORN_CHILD, it ignores
// SIGTERM, so only SIGKILL ends it.
const program = `
echo done > finished.txt
pgroup sh -c 'echo $$ > grandchild.tmp; mv grandchild.tmp grandchild.pid; exec sleep 60' >/dev/null 2>&1 &
while [ ! -f grandchild.pid ]; do sleep 0.01; done
if [ -n "$STUBBORN_CHILD" ]; then trap '' TERM; fi
echo $$ > program.tmp
mv program.tmp program.pid
while :; do sleep 0.1; done
`

// wait runs program in the current folder until its context ends, then
// answers with that folder. Under --stubborn, it never returns.
func wait(ctx context.Context, call extension.Call) contract.Outcome {
	stubbornChild := ""
	if call.Flags["stubborn-child"] == "true" {
		stubbornChild = "1"
	}
	_, err := runner.Run(ctx, runner.Program{
		Path: "/bin/sh",
		Args: []string{"-c", program},
		Env:  runner.Environ(os.Environ(), map[string]string{"STUBBORN_CHILD": stubbornChild}),
	})
	if err != nil {
		return contract.Outcome{Errors: []contract.Problem{{Code: contract.Runtime, Message: call.Printer.T("hello.program"), Hint: "ncly hello wait"}}}
	}
	if call.Flags["stubborn"] == "true" {
		select {}
	}
	dir, _ := os.Getwd()
	return contract.Outcome{Data: map[string]string{"output_dir": dir}}
}
