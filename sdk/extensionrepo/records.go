package extensionrepo

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

// The build records of COR-104-T4, which build-records.md in the contract
// defines. intake.toml holds the facts of the intake, and review.toml the
// readiness review and the ledger of the builds. Finish reads only these
// records. just records writes their views: ## Intake of vision.md, and
// readiness.md. Each state and each grade has one key, so prose never holds
// one (#422).
const (
	intakeFile = "intake.toml"
	reviewFile = "review.toml"
)

// An intake record holds the facts of the intake and the experiments apart.
type intakeRecord struct {
	Facts       []fact       `toml:"facts"`
	Experiments []experiment `toml:"experiments"`
}

// A fact is sourced, an inference, or unknown. Only a sourced fact names a
// source, its revision, and the part read. An unknown fact names the spec ID
// of the line that it opens in spec.md.
type fact struct {
	Kind      string `toml:"kind"`
	Statement string `toml:"statement"`
	Source    string `toml:"source"`
	Revision  string `toml:"revision"`
	Read      string `toml:"read"`
	ID        string `toml:"id"`
}

type experiment struct {
	Command string `toml:"command"`
	Result  string `toml:"result"`
}

// A review record holds the last readiness review and the ledger of the
// builds (#417).
type reviewRecord struct {
	Commit   string    `toml:"commit"`
	Reviewer string    `toml:"reviewer"`
	Verdict  string    `toml:"verdict"`
	Findings []finding `toml:"findings"`
	Rounds   []round   `toml:"rounds"`
	Order    order     `toml:"order"`
	FirstRun *firstRun `toml:"first_run"`
}

// A finding has one ID, severity, state, spec ID, and family. Its ID, such as
// R2-F3, names the round that first reported it. The keys after Reason hold
// prose, which no check reads, so a word there never grades (#419).
type finding struct {
	ID          string `toml:"id"`
	Title       string `toml:"title"`
	Severity    string `toml:"severity"`
	State       string `toml:"state"`
	Spec        string `toml:"spec"`
	Family      string `toml:"family"`
	Commit      string `toml:"commit"`
	Reason      string `toml:"reason"`
	Evidence    string `toml:"evidence"`
	Command     string `toml:"command"`
	Consequence string `toml:"consequence"`
	Owner       string `toml:"owner"`
	Fix         string `toml:"fix"`
}

// A round is one row of the ledger. Its place in the record gives its number.
type round struct {
	Head            string `toml:"head"`
	ChangeReview    string `toml:"change_review"`
	ChangeCommits   string `toml:"change_commits"`
	ReadinessReview string `toml:"readiness_review"`
	Finish          string `toml:"finish"`
}

// An order names the commit of each step of round 1 of the current build.
type order struct {
	Intake   string `toml:"intake"`
	Spec     string `toml:"spec"`
	Scenario string `toml:"scenario"`
	Handler  string `toml:"handler"`
}

// A first run records the run of the first scenario at its commit, before the
// handler (#420). Signoff runs the scenario of that name again (#422).
type firstRun struct {
	Commit    string `toml:"commit"`
	Scenario  string `toml:"scenario"`
	Exit      int    `toml:"exit"`
	Assertion string `toml:"assertion"`
}

// decodeRecord reads the record file from text into v. It reports a text that
// is no record. It also reports each key that the record does not define. So
// no state or grade hides in a key that finish never reads.
func decodeRecord(file, text string, v any) []string {
	md, err := toml.Decode(text, v)
	if err != nil {
		return []string{fmt.Sprintf("%s: it is no valid record: %v. Fix: write it as TOML with the keys that build-records.md in the contract lists", file, err)}
	}
	fix := "Fix: use only the keys that build-records.md in the contract lists"
	if file == reviewFile {
		fix = "Fix: grade a finding only in its key severity, give its state only in its key state, and use only the keys that build-records.md in the contract lists"
	}
	var problems, keys []string
	for _, key := range md.Undecoded() {
		keys = append(keys, key.String())
	}
	for _, key := range keys {
		// A table that the record does not define is one problem, not one for each of its keys.
		if parent, _, found := cutLast(key); found && slices.Contains(keys, parent) {
			continue
		}
		problems = append(problems, fmt.Sprintf("%s: the key %s is no field of the record, so no check reads it. %s", file, key, fix))
	}
	return problems
}

func cutLast(key string) (string, string, bool) {
	i := strings.LastIndex(key, ".")
	if i < 0 {
		return "", key, false
	}
	return key[:i], key[i+1:], true
}

// readIntake reads intake.toml. It returns the record, whether the file
// exists, and its problems.
func readIntake() (intakeRecord, bool, []string) {
	var record intakeRecord
	data, err := os.ReadFile(intakeFile)
	if errors.Is(err, os.ErrNotExist) {
		return record, false, nil
	} else if err != nil {
		return record, false, []string{intakeFile + ": " + err.Error()}
	}
	return record, true, decodeRecord(intakeFile, string(data), &record)
}

// readReview reads review.toml, as readIntake reads intake.toml.
func readReview() (reviewRecord, bool, []string) {
	var record reviewRecord
	data, err := os.ReadFile(reviewFile)
	if errors.Is(err, os.ErrNotExist) {
		return record, false, nil
	} else if err != nil {
		return record, false, []string{reviewFile + ": " + err.Error()}
	}
	return record, true, decodeRecord(reviewFile, string(data), &record)
}

// writeRecords runs just records. It writes ## Intake of vision.md from
// intake.toml, and readiness.md from review.toml when it exists. It writes
// nothing when a record has a problem of decodeRecord.
func writeRecords() result {
	intake, found, problems := readIntake()
	if !found && len(problems) == 0 {
		problems = append(problems, intakeFile+": the record is missing. Fix: restore it from the Git history of the repository, then fill it")
	}
	review, reviewed, more := readReview()
	problems = append(problems, more...)
	var vision []byte
	var err error
	if len(problems) == 0 {
		vision, err = os.ReadFile("vision.md")
		if err != nil {
			problems = append(problems, "vision.md: "+err.Error())
		}
	}
	view, ok := "", false
	if len(problems) == 0 {
		if view, ok = withIntake(string(vision), intakeView(intake)); !ok {
			problems = append(problems, "vision.md: it holds no part ## Intake. Fix: add the heading ## Intake, then run just records")
		}
	}
	if len(problems) != 0 {
		for _, problem := range problems {
			fmt.Fprintln(os.Stderr, problem)
		}
		return failed("records", errors.New("no view is written"))
	}
	if err := os.WriteFile("vision.md", []byte(view), 0o644); err != nil {
		return failed("records", err)
	}
	if !reviewed {
		return result{action: "records", answer: []string{"records: wrote the part ## Intake of vision.md from intake.toml. review.toml is missing, so readiness.md stays as it is"}}
	}
	if err := os.WriteFile("readiness.md", []byte(readinessView(review)), 0o644); err != nil {
		return failed("records", err)
	}
	return result{action: "records", answer: []string{"records: wrote the part ## Intake of vision.md from intake.toml, and readiness.md from review.toml"}}
}

const viewFix = "Fix: run just records, then commit both files"

// intakeViewProblems reports a part ## Intake of vision.md that differs from
// what just records writes from the intake record.
func intakeViewProblems(intake intakeRecord) []string {
	vision, err := os.ReadFile("vision.md")
	if err != nil {
		return nil
	}
	switch view, ok := withIntake(string(vision), intakeView(intake)); {
	case !ok:
		return []string{"vision.md: it holds no part ## Intake, so it shows no view of intake.toml. Fix: add the heading ## Intake, then run just records"}
	case view != string(vision):
		return []string{"vision.md: its part ## Intake is not the view of intake.toml. " + viewFix}
	}
	return nil
}

// reviewViewProblems reports a readiness.md that differs from what just
// records writes from the review record.
func reviewViewProblems(review reviewRecord) []string {
	if data, err := os.ReadFile("readiness.md"); err != nil || string(data) != readinessView(review) {
		return []string{"readiness.md: it is not the view of review.toml. " + viewFix}
	}
	return nil
}
