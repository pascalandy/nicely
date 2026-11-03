package extensionrepo

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"
)

// refusing is a stream that the reader closed.
type refusing struct{}

func (refusing) Write([]byte) (int, error) { return 0, errors.New("closed") }

// TestDeliverWritesTheOneResult covers each part of a result. The notices of
// a go test that passed come after its answer, as e35bd0bb requires.
func TestDeliverWritesTheOneResult(t *testing.T) {
	for _, c := range []struct {
		name           string
		r              result
		stdout, stderr string
		exit           int
	}{
		{"answer", result{action: "contract", answer: []string{"/contract"}}, "/contract\n", "", 0},
		{"notices of a pass", result{action: "check", notices: []byte("skipped: example\n")}, "", "skipped: example\n", 0},
		{"exit of a child", result{action: "test", notices: []byte("skipped: example\n"), exit: 3}, "", "skipped: example\n", 3},
		{"refusal", result{action: "finish", err: errors.New("not done")}, "", "finish: not done\n", 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if exit := c.r.deliver(&stdout, &stderr); exit != c.exit || stdout.String() != c.stdout || stderr.String() != c.stderr {
				t.Fatalf("deliver gave exit %d, stdout %q, stderr %q; want %d, %q, %q", exit, stdout.String(), stderr.String(), c.exit, c.stdout, c.stderr)
			}
		})
	}
}

// TestDeliverFailsOnALostAnswer requires exit 1 when stdout refuses the
// answer, such as Signed off on <sha>, which the caller never reads then.
func TestDeliverFailsOnALostAnswer(t *testing.T) {
	var stderr bytes.Buffer
	r := result{action: "signoff", answer: []string{"Signed off on abc"}}
	if exit := r.deliver(refusing{}, &stderr); exit != 1 || stderr.String() != "signoff: output was not delivered: closed\n" {
		t.Fatalf("deliver gave exit %d and stderr %q", exit, stderr.String())
	}
}

// TestARunResultCarriesItsNotices covers each action that runs one child in a
// host run, such as acceptance, whose notices review round 2 found lost.
func TestARunResultCarriesItsNotices(t *testing.T) {
	run, err := NewOwnedRun(t.TempDir(), Profile{})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(noticeFile, filepath.Join(run.Dir, "notices"))
	if err := Notice("skipped: acceptance helper"); err != nil {
		t.Fatal(err)
	}
	r := run.result("acceptance", nil)
	if string(r.notices) != "skipped: acceptance helper\n" || r.err != nil || r.exit != 0 {
		t.Fatalf("the result of the run holds notices %q, err %v, exit %d", r.notices, r.err, r.exit)
	}
}
