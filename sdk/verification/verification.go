// Package verification runs the TLA+ models of a repository with a pinned
// TLC, as Formal models in the contract defines (COR-104-T8).
//
// This file is the stub of the first commit of the card: it checks no
// digest, runs no model, and reads no record.
package verification

// SuiteFile and RecordFile are the suite and the record of a repository.
const (
	SuiteFile  = "verification/tlc.json"
	RecordFile = "verification/tlc-record.json"
)

// Suite pins Java and TLC, and lists the checks of a repository.
type Suite struct {
	Java   string  `json:"java"`
	URL    string  `json:"tlc_url"`
	SHA256 string  `json:"tlc_sha256"`
	Checks []Check `json:"checks"`
}

// Check is one model and configuration, with its expected violation.
type Check struct {
	Name      string `json:"name"`
	Model     string `json:"model"`
	Config    string `json:"config"`
	Violation string `json:"violation,omitempty"`
}

// Result is the outcome of one check.
type Result struct {
	Name      string `json:"name"`
	Exit      int    `json:"exit"`
	Violation string `json:"violation,omitempty"`
	Distinct  int    `json:"distinct,omitempty"`
	Depth     int    `json:"depth,omitempty"`
}

// Record is what just check reads.
type Record struct {
	Version int               `json:"version"`
	Java    string            `json:"java"`
	SHA256  string            `json:"tlc_sha256"`
	Inputs  map[string]string `json:"inputs"`
	Results []Result          `json:"results"`
}

// Options are the folders and the environment of one run.
type Options struct {
	Root, Dir, Jar string
	Env            []string
}

// Run is a stub.
func Run(Options) ([]Result, error) { return nil, nil }

// CheckRecord is a stub.
func CheckRecord(string) []string { return nil }

func parseResult(check Check, code int, _ string) (Result, error) {
	return Result{Name: check.Name, Exit: code}, nil
}
