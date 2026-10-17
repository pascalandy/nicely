// Package record keeps the records of recorded operations, as Operations
// in contract.md defines them: run IDs, one durable record per run, the
// lock that one execution holds, the transitions of its steps, and the
// verified artifacts that a success answers from. The domain names its
// steps, declares their effects, and verifies their results; this package
// never decides what a step does.
package record

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/platform"
)

// FormatVersion is the record format that this package reads and writes.
const FormatVersion = 1

// LockWait bounds how long an execution waits for a run that another
// process holds, as Record format in contract.md defines.
const LockWait = 250 * time.Millisecond

// Status is the saved status of a run.
type Status string

const (
	Running     Status = "running"
	Completed   Status = "completed"
	Failed      Status = "failed"
	Interrupted Status = "interrupted"
	Unknown     Status = "unknown"
)

// StepStatus is the saved status of a step. Pending means that its effect
// has not started.
type StepStatus string

const (
	StepPending   StepStatus = "pending"
	StepCompleted StepStatus = "completed"
	StepFailed    StepStatus = "failed"
	StepUnknown   StepStatus = "unknown"
)

// Effect is the strongest effect that a step can make, as its domain
// declares it, each value more cautious than the one before it.
type Effect string

const (
	EffectNone          Effect = "none"
	EffectRepeatable    Effect = "repeatable"
	EffectNonRepeatable Effect = "non_repeatable"
	EffectPaid          Effect = "paid"
)

// Record is the record of one run, as Record format in contract.md
// defines its keys.
type Record struct {
	FormatVersion int               `json:"format_version"`
	RunID         string            `json:"run_id"`
	Path          string            `json:"path"`
	Status        Status            `json:"status"`
	StartedAt     string            `json:"started_at"`
	UpdatedAt     string            `json:"updated_at"`
	NclyVersion   string            `json:"ncly_version"`
	Attempts      int               `json:"attempts"`
	Inputs        map[string]any    `json:"inputs"`
	Resolved      map[string]any    `json:"resolved"`
	Tools         map[string]string `json:"tools"`
	Items         []Item            `json:"items"`
}

// Item is one input item of a run, with its steps in execution order.
type Item struct {
	Key string `json:"key"`
	// OutputDir is the folder selected for the item's results, or nil until
	// the domain selects one.
	OutputDir *string `json:"output_dir"`
	Steps     []Step  `json:"steps"`
}

// Step is one step of an item.
type Step struct {
	ID         string     `json:"id"`
	Effect     Effect     `json:"effect"`
	Status     StepStatus `json:"status"`
	StartedAt  *string    `json:"started_at"`
	FinishedAt *string    `json:"finished_at"`
	Artifacts  []Artifact `json:"artifacts"`
	// Error is the latest unsuccessful attempt's problem, or nil.
	Error *Problem `json:"error"`
	// Evidence holds what the domain needs to answer again, with its own
	// format_version, or nothing.
	Evidence json.RawMessage `json:"evidence,omitempty"`
}

// Problem is the error of a step, as the public problem object holds it.
type Problem struct {
	Code    contract.Code `json:"code"`
	Message string        `json:"message"`
	Hint    string        `json:"hint"`
}

// Artifact is a verified output of a step.
type Artifact struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

// idPattern is a run ID: 1 to 64 lowercase letters, digits, and hyphens,
// starting with a letter or a digit.
var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// ValidID reports whether id is a run ID, as Run IDs in contract.md
// defines it.
func ValidID(id string) bool { return idPattern.MatchString(id) }

// NewID generates a run ID from the UTC time and six random hexadecimal
// digits, such as 20261007-214501-3f9a2c. Readers never parse it.
func NewID(now time.Time) string {
	random := make([]byte, 3)
	_, _ = rand.Read(random)
	return now.UTC().Format("20060102-150405") + "-" + hex.EncodeToString(random)
}

// Folder returns the runs folder of a state folder, which holds one record
// per run and the folder of each run's artifacts.
func Folder(state string) string { return filepath.Join(state, "runs") }

// Store is a runs folder.
type Store struct {
	Dir string
}

// File returns the path of the record of a run.
func (s Store) File(id string) string { return filepath.Join(s.Dir, id+".json") }

func (s Store) lock(id string) string { return filepath.Join(s.Dir, id+".lock") }

// ArtifactDir returns the folder of a run's artifacts, beside its record.
func (s Store) ArtifactDir(id string) string { return filepath.Join(s.Dir, id) }

// ErrBusy means that another process held the run for longer than
// LockWait.
var ErrBusy = errors.New("another process holds the run")

// UnreadableError is a record that cannot be read or parsed, or that holds
// a format_version that this package does not support. It never counts as
// an absent record, so it never permits a new paid operation.
type UnreadableError struct {
	File string
	Err  error
}

func (e *UnreadableError) Error() string { return fmt.Sprintf("%s: %v", e.File, e.Err) }

func (e *UnreadableError) Unwrap() error { return e.Err }

// Run is a run whose lock this process holds, with its record as read
// under that lock: nil when no record holds its ID.
type Run struct {
	ID     string
	Record *Record
	store  Store
	unlock func()
	// effect is the strongest effect of the steps whose intent this
	// process persisted.
	effect contract.Effect
}

// Open takes the lock of a run, waiting at most LockWait, then reads its
// record while it holds the lock, as Record format in contract.md asks. It
// returns ErrBusy when another process holds the run, and an
// UnreadableError for a record that it cannot use. The caller closes the
// run, which releases the lock, once it has answered.
func (s Store) Open(ctx context.Context, id string) (*Run, error) {
	unlock, err := platform.Lock(ctx, s.lock(id), LockWait)
	if errors.Is(err, platform.ErrLocked) {
		return nil, ErrBusy
	}
	if err != nil {
		return nil, err
	}
	r := &Run{ID: id, store: s, unlock: unlock}
	if r.Record, err = read(s.File(id), id); err != nil && !errors.Is(err, fs.ErrNotExist) {
		unlock()
		return nil, err
	}
	return r, nil
}

// Read reads the record of a run without its lock, for a dry run or an
// inspection: nil when no record holds the ID, and an UnreadableError for
// a record that it cannot use.
func (s Store) Read(id string) (*Record, error) {
	rec, err := read(s.File(id), id)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return rec, err
}

// Active reports whether a process holds the run now. It never creates the
// lock file and never holds the lock past its answer.
func (s Store) Active(id string) (bool, error) { return platform.Held(s.lock(id)) }

// read reads the record of a file, or returns fs.ErrNotExist when no file
// exists. Any other failure is an UnreadableError.
func read(file, id string) (*Record, error) {
	data, err := os.ReadFile(file)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if err != nil {
		return nil, &UnreadableError{File: file, Err: err}
	}
	var version struct {
		FormatVersion *int `json:"format_version"`
	}
	if err := json.Unmarshal(data, &version); err != nil {
		return nil, &UnreadableError{File: file, Err: err}
	}
	if version.FormatVersion == nil || *version.FormatVersion != FormatVersion {
		return nil, &UnreadableError{File: file, Err: errors.New("unsupported format_version")}
	}
	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, &UnreadableError{File: file, Err: err}
	}
	if !shaped(data) || !fits(&rec, id) {
		return nil, &UnreadableError{File: file, Err: errors.New("the record breaks its format")}
	}
	return &rec, nil
}

// shaped reports whether data holds every key of Record format in
// contract.md that is not optional, nested ones included, each with a value
// of its JSON type, so that a missing key never reads as null.
func shaped(data []byte) bool {
	var rec map[string]any
	if json.Unmarshal(data, &rec) != nil ||
		!holds(rec, "string", "run_id", "path", "status", "started_at", "updated_at", "ncly_version") ||
		!holds(rec, "number", "attempts") || !holds(rec, "object", "inputs", "resolved", "tools") || !holds(rec, "array", "items") {
		return false
	}
	for _, version := range rec["tools"].(map[string]any) {
		if _, isString := version.(string); !isString {
			return false
		}
	}
	for _, x := range rec["items"].([]any) {
		it, isObject := x.(map[string]any)
		if !isObject || !holds(it, "string", "key") || !holds(it, "string?", "output_dir") || !holds(it, "array", "steps") {
			return false
		}
		for _, y := range it["steps"].([]any) {
			s, isObject := y.(map[string]any)
			if !isObject || !holds(s, "string", "id", "effect", "status") || !holds(s, "string?", "started_at", "finished_at") ||
				!holds(s, "array", "artifacts") || !holds(s, "object?", "error") {
				return false
			}
			if e, _ := s["error"].(map[string]any); e != nil && !holds(e, "string", "code", "message", "hint") {
				return false
			}
			if _, present := s["evidence"]; present && !holds(s, "object", "evidence") {
				return false
			}
			for _, z := range s["artifacts"].([]any) {
				a, isObject := z.(map[string]any)
				if !isObject || !holds(a, "string", "path", "sha256") || !holds(a, "number", "size_bytes") {
					return false
				}
			}
		}
	}
	return true
}

// holds reports whether o holds each named key with a value of kind:
// string, number, object, or array, or null too when kind ends with ?.
func holds(o map[string]any, kind string, names ...string) bool {
	kind, nullable := strings.CutSuffix(kind, "?")
	for _, name := range names {
		value, present := o[name]
		ok := present && value == nil && nullable
		switch value.(type) {
		case string:
			ok = ok || kind == "string"
		case float64:
			ok = ok || kind == "number"
		case map[string]any:
			ok = ok || kind == "object"
		case []any:
			ok = ok || kind == "array"
		}
		if !ok {
			return false
		}
	}
	return true
}

// fits reports whether rec holds a value in the range of each key of Record
// format in contract.md, and the run ID of its file, so that no reader uses
// a record that it would misread.
func fits(rec *Record, id string) bool {
	if rec.RunID != id || rec.Path == "" || rec.NclyVersion == "" || rec.Attempts < 1 ||
		!isTime(rec.StartedAt) || !isTime(rec.UpdatedAt) ||
		!slices.Contains([]Status{Running, Completed, Failed, Interrupted, Unknown}, rec.Status) ||
		rec.Inputs == nil || rec.Resolved == nil || rec.Tools == nil || rec.Items == nil {
		return false
	}
	for _, it := range rec.Items {
		if it.Key == "" || it.Steps == nil {
			return false
		}
		for _, s := range it.Steps {
			if s.ID == "" || s.Artifacts == nil ||
				!slices.Contains([]StepStatus{StepPending, StepCompleted, StepFailed, StepUnknown}, s.Status) ||
				!slices.Contains([]Effect{EffectNone, EffectRepeatable, EffectNonRepeatable, EffectPaid}, s.Effect) ||
				s.StartedAt != nil && !isTime(*s.StartedAt) || s.FinishedAt != nil && !isTime(*s.FinishedAt) {
				return false
			}
		}
	}
	return true
}

// isTime reports whether text is a time in RFC 3339.
func isTime(text string) bool {
	_, err := time.Parse(time.RFC3339, text)
	return err == nil
}

// Close releases the lock of the run.
func (r *Run) Close() { r.unlock() }

// Effect returns the strongest effect of the steps whose intent this
// process persisted, which the outcome carries, so that exit 75 holds only
// when Retry safety in contract.md allows it.
func (r *Run) Effect() contract.Effect { return r.effect }

// File returns the path of the run's record.
func (r *Run) File() string { return r.store.File(r.ID) }

// ArtifactDir returns the folder of the run's artifacts.
func (r *Run) ArtifactDir() string { return r.store.ArtifactDir(r.ID) }

// now returns the time of a transition, in UTC and RFC 3339.
func now() string { return time.Now().UTC().Format(time.RFC3339) }

// Create writes the first record of a new run, with the status running and
// every step pending, before any effect. The domain fills the path, the
// version of ncly, the inputs, the resolved values, the tools, and the
// items with their steps and effects.
func (r *Run) Create(rec Record) error {
	t := now()
	rec.FormatVersion, rec.RunID, rec.Status, rec.StartedAt, rec.Attempts = FormatVersion, r.ID, Running, t, 1
	for i := range rec.Items {
		for j := range rec.Items[i].Steps {
			s := &rec.Items[i].Steps[j]
			s.Status, s.StartedAt, s.FinishedAt, s.Artifacts, s.Error = StepPending, nil, nil, []Artifact{}, nil
		}
	}
	r.Record = &rec
	return r.save()
}

// save writes the record durably, as Record format in contract.md asks:
// a temporary file of its own, flushed, renamed over the record, then the
// folder flushed. A record written this way survives a crash.
func (r *Run) save() error {
	r.Record.UpdatedAt = now()
	data, err := json.Marshal(r.Record)
	if err != nil {
		return err
	}
	return platform.WriteFile(r.store.File(r.ID), append(data, '\n'), 0o600)
}

// Again starts a new attempt of a run whose saved intents allow it, as
// Run IDs in contract.md defines: attempts plus one and the status running,
// with the release of ncly and the versions of the tools that this attempt
// uses, saved before any intent.
func (r *Run) Again(nclyVersion string, tools map[string]string) error {
	r.Record.Attempts++
	r.Record.Status, r.Record.NclyVersion, r.Record.Tools = Running, nclyVersion, tools
	return r.save()
}

// Begin persists the intent of a step before its effect: the step turns
// unknown, with its start time. The effect may start only once Begin
// returns nil. A step that may bill stays unknown until its domain records
// a verified result.
func (r *Run) Begin(item, step int) error {
	s := &r.Record.Items[item].Steps[step]
	t := now()
	s.Status, s.StartedAt, s.FinishedAt, s.Error = StepUnknown, &t, nil, nil
	if err := r.save(); err != nil {
		return err
	}
	r.effect = max(r.effect, effectOf(s.Effect))
	return nil
}

// effectOf maps a declared effect to the effect of an invocation that
// started it.
func effectOf(e Effect) contract.Effect {
	switch e {
	case EffectNone:
		return contract.NoEffect
	case EffectRepeatable:
		return contract.RepeatableEffect
	case EffectNonRepeatable:
		return contract.NonRepeatableEffect
	}
	return contract.PaidEffect
}

// Complete records the verified result of a step: its artifacts, which the
// domain verified, and its evidence. The run completes with its last step.
func (r *Run) Complete(item, step int, artifacts []Artifact, evidence any) error {
	s := &r.Record.Items[item].Steps[step]
	data, err := json.Marshal(evidence)
	if err != nil {
		return err
	}
	t := now()
	s.Status, s.FinishedAt, s.Artifacts, s.Evidence, s.Error = StepCompleted, &t, artifacts, data, nil
	r.Record.Status = Completed
	for _, it := range r.Record.Items {
		for _, other := range it.Steps {
			if other.Status != StepCompleted {
				r.Record.Status = Running
			}
		}
	}
	return r.save()
}

// End records how an execution ended without completion: the problem of a
// step, the step's status, and the run's status, which the domain derives
// from its evidence, so a missing answer never reads as no effect.
func (r *Run) End(item, step int, status StepStatus, p contract.Problem, run Status) error {
	s := &r.Record.Items[item].Steps[step]
	t := now()
	s.Status, s.Error = status, &Problem{Code: p.Code, Message: p.Message, Hint: p.Hint}
	if status != StepUnknown {
		s.FinishedAt = &t
	}
	r.Record.Status = run
	return r.save()
}

// Publish writes data as the artifact name in the folder of the run,
// durably, then reads it back, and returns it with its size and SHA-256.
// The domain records completion only after Publish returns.
func (r *Run) Publish(name string, data []byte) (Artifact, error) {
	path := filepath.Join(r.ArtifactDir(), name)
	if err := platform.WriteFile(path, data, 0o600); err != nil {
		return Artifact{}, err
	}
	sum := sha256.Sum256(data)
	a := Artifact{Path: path, SizeBytes: int64(len(data)), SHA256: hex.EncodeToString(sum[:])}
	if _, err := a.Read(); err != nil {
		return Artifact{}, err
	}
	return a, nil
}

// ErrChanged means that an artifact no longer holds what its record says.
var ErrChanged = errors.New("the artifact changed since its run recorded it")

// Read returns the content of an artifact once it matches its recorded
// size and SHA-256, and ErrChanged otherwise.
func (a Artifact) Read() ([]byte, error) {
	data, err := os.ReadFile(a.Path)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	if int64(len(data)) != a.SizeBytes || hex.EncodeToString(sum[:]) != a.SHA256 {
		return nil, ErrChanged
	}
	return data, nil
}
