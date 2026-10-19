package transcript

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/record"
)

// maxCandidates bounds the result folders that one intent tries: name,
// then name-2 up to name-1000.
const maxCandidates = 1000

// execute runs the prepared job under the run ID of the call, or a new one,
// as Records and replay in spec/records-and-replay.md defines. It reads the
// granted key before any record, then takes the run lock. It replays a
// completed run, refuses an uncertain one, and otherwise starts an attempt.
func execute(ctx context.Context, call extension.Call, j job, id string) contract.Outcome {
	p := call.Printer
	key, problem := call.Key("deepgram")
	if problem != nil {
		return refused(problem)
	}
	runs := store()
	if id == "" {
		id = record.NewID(time.Now())
	}
	run, err := runs.Open(ctx, id)
	var bad *record.UnreadableError
	switch {
	case errors.Is(err, record.ErrBusy):
		return refusal(inspect(p, contract.Temporary, id, "transcript.run.busy", map[string]any{"ID": id}), id)
	case errors.As(err, &bad):
		return refusal(unreadable(p, bad.File, id), id)
	case ctx.Err() != nil:
		return failed(stopCause(ctx, p, j))
	case err != nil:
		return refusal(unsaved(p, runs.Dir), id)
	}
	defer run.Close()
	if run.Record != nil {
		completed, problem := decide(p, call, j, run.Record, run.File(), id)
		if problem != nil {
			return refusal(problem, id)
		}
		if completed {
			return replay(call, run.Record, run.File(), id)
		}
	}
	a := &attempt{call: call, p: p, j: j, run: run, saved: run.Record != nil}
	return a.execute(ctx, key)
}

// attempt is one attempt of a run through the protocol of spec/protocol.md.
// Its record exists once the plan is valid. Its step turns unknown once the
// record holds the intent, and its result folder exists once reserved.
type attempt struct {
	call  extension.Call
	p     *i18n.Printer
	j     job
	run   *record.Run
	s     *session
	saved bool
	begun bool
	dir   string
	// codes holds the program warnings of the terminal step.
	codes []string
}

func (a *attempt) execute(ctx context.Context, key string) contract.Outcome {
	s, problem := start(ctx, a.call, a.j, false, map[string]string{"DEEPGRAM_API_KEY": key})
	if problem != nil {
		return a.refuse(problem)
	}
	a.s = s
	defer s.close()
	if problem := s.plan(ctx, a.p, a.j); problem != nil || ctx.Err() != nil {
		return a.refuse(problem)
	}
	if err := a.create(); err != nil {
		return a.refuse(unsaved(a.p, filepath.Dir(a.run.File())))
	}
	m, problem := a.next(ctx)
	switch {
	case ctx.Err() != nil:
		return a.stop(ctx)
	case problem != nil:
		return a.end(record.StepPending, problem)
	case m.Type == "step":
		return a.terminal(ctx, m)
	case m.Type != "intent":
		return a.end(record.StepPending, s.broke(a.p, "transcript.protocol.order"))
	case !validName(m.Name) || m.Tools["transcript"] == "" || m.Tools["yt-dlp"] == "":
		return a.end(record.StepPending, s.broke(a.p, "transcript.protocol.intent"))
	}
	a.run.Record.Tools = map[string]string{"transcript": m.Tools["transcript"], "yt-dlp": m.Tools["yt-dlp"]}
	if problem := a.reserve(m.Name); problem != nil {
		return a.end(record.StepFailed, problem)
	}
	s.send(ack{ProtocolVersion: protocolVersion, Type: "ack", Item: a.j.url, Step: stepID, OutputDir: a.dir})
	m, problem = a.next(ctx)
	switch {
	case ctx.Err() != nil:
		return a.stop(ctx)
	case problem != nil:
		return a.end(record.StepUnknown, problem)
	case m.Type != "step":
		return a.end(record.StepUnknown, s.broke(a.p, "transcript.protocol.order"))
	}
	return a.terminal(ctx, m)
}

// create durably records the validated plan with its pending step. It
// writes a new record, or a new attempt of a run whose step is still
// pending, with the values that this preparation selected.
func (a *attempt) create() error {
	if a.saved {
		a.run.Record.Resolved = resolved(a.j)
		return a.run.Again(nclyVersion(), map[string]string{})
	}
	return a.run.Create(record.Record{
		Path:        a.call.Path,
		NclyVersion: nclyVersion(),
		Inputs:      inputs(a.j),
		Resolved:    resolved(a.j),
		Tools:       map[string]string{},
		Items:       []record.Item{{Key: a.j.url, Steps: []record.Step{{ID: stepID, Effect: record.EffectPaid}}}},
	})
}

// next reads the next message for the item, and the problem of an engine
// that broke the protocol before it.
func (a *attempt) next(ctx context.Context) (message, *contract.Problem) {
	m, open, broken := a.s.next(ctx)
	switch {
	case ctx.Err() != nil:
		return m, nil
	case broken:
		return m, a.s.broke(a.p, "transcript.protocol.malformed")
	case !open:
		if result := a.s.wait(); result.Exit != 0 {
			return m, a.s.broke(a.p, "transcript.protocol.exit")
		}
		return m, a.s.broke(a.p, "transcript.protocol.step")
	case m.ProtocolVersion != protocolVersion:
		return m, a.s.broke(a.p, "transcript.protocol.version")
	case m.Item != a.j.url || m.Step != stepID:
		return m, a.s.broke(a.p, "transcript.protocol.item")
	}
	return m, nil
}

// validName reports whether the engine's proposed name is one folder name.
func validName(name string) bool {
	return name != "" && name != "." && name != ".." && filepath.Base(name) == name && len(name) <= 200
}

// reserve records each candidate result folder before it creates it with
// an exclusive mkdir, as Output and cookies in spec/output-and-cookies.md
// asks. The candidates are name, then name-2, name-3, and so on. Begin records the candidate
// and the unknown step. The parent is synced, so the folder survives a
// crash before the ack. The ack has not left, so a failure here sends no
// audio, and the step fails.
func (a *attempt) reserve(name string) *contract.Problem {
	for n := 1; n <= maxCandidates; n++ {
		dir := filepath.Join(a.j.output, name)
		if n > 1 {
			dir += "-" + strconv.Itoa(n)
		}
		a.run.Record.Items[0].OutputDir = &dir
		if err := a.run.Begin(0, 0); err != nil {
			return unsaved(a.p, filepath.Dir(a.run.File()))
		}
		a.begun = true
		err := os.Mkdir(dir, 0o755)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		if err == nil {
			err = syncPath(a.j.output)
		}
		if err != nil {
			break
		}
		a.dir = dir
		return nil
	}
	data := map[string]any{"Dir": a.j.output}
	return &contract.Problem{Code: contract.Runtime, Message: a.p.T("transcript.run.reserve_failed", data), Hint: a.p.T("transcript.run.reserve_failed_hint", data)}
}
