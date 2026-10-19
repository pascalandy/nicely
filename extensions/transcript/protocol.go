package transcript

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/i18n"
	"github.com/pascalandy/nicely/sdk/runner"
)

// protocolVersion is the version of the JSON Lines protocol that this
// extension and its engine speak, as spec/protocol.md defines it.
const protocolVersion = 1

// maxLine bounds one message of the engine.
const maxLine = 1 << 20

// request is the first message, which Go sends, with every key of the
// table of spec/protocol.md. zoom stays null.
type request struct {
	ProtocolVersion int      `json:"protocol_version"`
	Type            string   `json:"type"`
	Source          string   `json:"source"`
	URLs            []string `json:"urls"`
	Zoom            *string  `json:"zoom"`
	Output          string   `json:"output"`
	TempDir         string   `json:"temp_dir"`
	Browser         *string  `json:"browser"`
	DryRun          bool     `json:"dry_run"`
	// Deadline is the UTC end of execution, in RFC 3339.
	Deadline string `json:"deadline"`
}

// message is one JSON line that the engine sends. Each type uses some of
// its fields, as the table of spec/protocol.md lists them.
type message struct {
	ProtocolVersion int    `json:"protocol_version"`
	Type            string `json:"type"`
	// plan
	Items []struct {
		Item   string `json:"item"`
		Output string `json:"output"`
	} `json:"items"`
	Code   *string `json:"code"`
	Detail string  `json:"detail"`
}

// session is one run of the engine through uv: the request on its stdin,
// and its messages, read from its stdout, in order.
type session struct {
	// out holds the messages for the engine's stdin, which one goroutine
	// writes in order. Closing it closes that stdin.
	out      chan []byte
	once     sync.Once
	messages chan message
	// broken holds the first line that is no message, which ends messages.
	broken chan struct{}
	exited chan runner.Result
	result runner.Result
	tmp    string
}

// start extracts the engine, starts it through uv with the environment env
// adds to that of ncly, and sends the request. A dry run sends no key.
func start(ctx context.Context, call extension.Call, j job, dryRun bool, env map[string]string) (*session, *contract.Problem) {
	p := call.Printer
	dir, err := extract(j.cache)
	if err != nil {
		return nil, &contract.Problem{Code: contract.Runtime, Message: p.T("transcript.run.extract_failed", map[string]any{"Dir": j.cache}), Hint: p.T("transcript.run.extract_failed_hint")}
	}
	tmp, err := os.MkdirTemp("", "ncly-transcript-")
	if err != nil {
		return nil, &contract.Problem{Code: contract.Runtime, Message: p.T("transcript.run.start_failed"), Hint: p.T("transcript.run.verbose_hint")}
	}
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	var stderr io.Writer
	if os.Getenv("NCLY_VERBOSE") == "1" || os.Getenv("NCLY_DEBUG") != "" {
		stderr = call.Stderr
	}
	s := &session{out: make(chan []byte, 2), messages: make(chan message), broken: make(chan struct{}), exited: make(chan runner.Result, 1), tmp: tmp}
	// Python writes no bytecode, so the folder of a hash keeps the files of
	// that hash alone.
	set := map[string]string{"PYTHONDONTWRITEBYTECODE": "1"}
	maps.Copy(set, env)
	program := runner.Program{Path: j.uv, Args: uvArgs(dir, j.cache), Env: runner.Environ(os.Environ(), set), Stdin: stdinR, Stdout: stdoutW, Stderr: stderr}
	go func() {
		result, err := runner.Run(ctx, program)
		if err != nil {
			result.Exit = -1
		}
		_ = stdoutW.Close()
		_ = stdinR.Close()
		s.exited <- result
	}()
	go s.read(stdoutR)
	go func() {
		for data := range s.out {
			if _, err := stdinW.Write(data); err != nil {
				break
			}
		}
		_ = stdinW.Close()
		for range s.out {
		}
	}()
	var browser *string
	if j.browser != "" {
		browser = &j.browser
	}
	s.send(request{
		ProtocolVersion: protocolVersion, Type: "request", Source: "youtube", URLs: []string{j.url},
		Output: j.output, TempDir: tmp, Browser: browser, DryRun: dryRun,
		Deadline: time.Now().Add(j.timeout).UTC().Format(time.RFC3339),
	})
	return s, nil
}

// read decodes each line of the engine's stdout until it ends, or until a
// line is no message of this protocol version.
func (s *session) read(r io.Reader) {
	defer close(s.messages)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), maxLine)
	for scanner.Scan() {
		var m message
		if json.Unmarshal(scanner.Bytes(), &m) != nil {
			close(s.broken)
			break
		}
		s.messages <- m
	}
	_, _ = io.Copy(io.Discard, r)
}

// send queues one message for the engine's stdin. A write to an engine that
// has exited fails, and the exit then decides the outcome.
func (s *session) send(m any) {
	data, _ := json.Marshal(m)
	s.out <- append(data, '\n')
}

// closeStdin ends the engine's stdin once the queued messages are written.
func (s *session) closeStdin() { s.once.Do(func() { close(s.out) }) }

// next returns the next message, or false once the engine's stdout ended.
// A line that is no JSON object breaks the protocol.
func (s *session) next(ctx context.Context) (message, bool, bool) {
	select {
	case m, open := <-s.messages:
		if !open {
			select {
			case <-s.broken:
				return message{}, false, true
			default:
			}
		}
		return m, open, false
	case <-ctx.Done():
		return message{}, false, false
	}
}

// wait waits for the engine to exit, and returns how it ended.
func (s *session) wait() runner.Result {
	if s.exited != nil {
		s.result = <-s.exited
		s.exited = nil
	}
	return s.result
}

// close ends the engine's stdin, waits for it to exit, and removes its
// temporary folder.
func (s *session) close() {
	s.closeStdin()
	s.wait()
	_ = os.RemoveAll(s.tmp)
}

// plan reads the engine's plan and checks it against the job, as steps 1
// and 2 of spec/protocol.md ask. The plan holds one item, with the URL as
// supplied and the output parent of the request.
func (s *session) plan(ctx context.Context, p *i18n.Printer, j job) *contract.Problem {
	m, open, broken := s.next(ctx)
	switch {
	case ctx.Err() != nil:
		return s.stopped(ctx, p, j)
	case broken:
		return s.broke(p, "transcript.protocol.malformed")
	case !open:
		return s.missing(ctx, p, j)
	case m.ProtocolVersion != protocolVersion:
		return s.broke(p, "transcript.protocol.version")
	case m.Type != "plan":
		return s.broke(p, "transcript.protocol.order")
	case len(m.Items) == 0 && m.Code != nil:
		return s.broke(p, "transcript.protocol.preparation")
	case len(m.Items) != 1:
		return s.broke(p, "transcript.protocol.items")
	case m.Items[0].Item != j.url || m.Items[0].Output != j.output:
		return s.broke(p, "transcript.protocol.item")
	}
	return nil
}

// finish closes the engine's stdin once it has nothing more to send, and
// checks that it sends no other message and exits 0.
func (s *session) finish(ctx context.Context, p *i18n.Printer, j job) *contract.Problem {
	s.closeStdin()
	_, open, broken := s.next(ctx)
	switch {
	case ctx.Err() != nil:
		return s.stopped(ctx, p, j)
	case broken:
		return s.broke(p, "transcript.protocol.malformed")
	case open:
		return s.broke(p, "transcript.protocol.order")
	}
	if result := s.wait(); result.Exit != 0 {
		return s.broke(p, "transcript.protocol.exit")
	}
	return nil
}

// missing is the problem of an engine whose stdout ended before its plan.
// uv or Python stopped first, such as when uv could not prepare the
// dependencies of the engine.
func (s *session) missing(ctx context.Context, p *i18n.Printer, j job) *contract.Problem {
	result := s.wait()
	if ctx.Err() != nil || result.Signal != nil {
		return s.stopped(ctx, p, j)
	}
	if result.Exit != 0 {
		return &contract.Problem{Code: contract.Runtime, Message: p.T("transcript.run.not_started", map[string]any{"Exit": result.Exit}), Hint: p.T("transcript.run.not_started_hint")}
	}
	return s.broke(p, "transcript.protocol.plan")
}

// broke is the RUNTIME problem of an engine that broke the protocol, with
// the message that names how.
func (s *session) broke(p *i18n.Printer, id string) *contract.Problem {
	return &contract.Problem{Code: contract.Runtime, Message: p.T(id), Hint: p.T("transcript.run.verbose_hint")}
}

// stopped is the problem of a call that its timeout stopped: RUNTIME, whose
// hint asks for a longer timeout. It is nil for a signal, which the SDK
// answers.
func (s *session) stopped(ctx context.Context, p *i18n.Printer, j job) *contract.Problem {
	s.close()
	if !errors.Is(context.Cause(ctx), errTimedOut) {
		return nil
	}
	data := map[string]any{"Limit": limitText(j.timeout)}
	return &contract.Problem{Code: contract.Runtime, Message: p.T("transcript.run.timed_out", data), Hint: p.T("transcript.run.timed_out_hint")}
}

// limitText writes a timeout as a caller writes it, such as 30s, 5m, or
// 9m30s, without the zero units that time.Duration adds.
func limitText(d time.Duration) string {
	text := d.String()
	if strings.HasSuffix(text, "m0s") {
		text = strings.TrimSuffix(text, "0s")
	}
	if strings.HasSuffix(text, "h0m") {
		text = strings.TrimSuffix(text, "0m")
	}
	return text
}
