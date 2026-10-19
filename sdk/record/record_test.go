package record

import (
	"context"
	"testing"

	"github.com/pascalandy/nicely/sdk/contract"
)

// TestEndStampsFinishedAtOnlyOnAnEndedStep ends a step that never started
// with each status, as Record format in contract.md defines finished_at:
// null until the step starts or ends. A pending step never started, and an
// unknown one has not ended, so only a failed step gets the time.
func TestEndStampsFinishedAtOnlyOnAnEndedStep(t *testing.T) {
	for _, c := range []struct {
		status   StepStatus
		finished bool
	}{{StepPending, false}, {StepUnknown, false}, {StepFailed, true}} {
		s := Store{Dir: t.TempDir()}
		run, err := s.Open(context.Background(), "r1")
		if err != nil {
			t.Fatal(err)
		}
		err = run.Create(Record{Path: "hello run", NclyVersion: "v1", Inputs: map[string]any{}, Resolved: map[string]any{}, Tools: map[string]string{}, Items: []Item{{Key: "k", Steps: []Step{{ID: "s", Effect: EffectPaid}}}}})
		if err == nil {
			err = run.End(0, 0, c.status, contract.Problem{Code: contract.Temporary, Message: "m", Hint: "h"}, Failed)
		}
		run.Close()
		if err != nil {
			t.Fatal(err)
		}
		saved, err := s.Read("r1")
		if err != nil {
			t.Fatal(err)
		}
		step := saved.Items[0].Steps[0]
		if step.Status != c.status || step.Error == nil || (step.FinishedAt != nil) != c.finished {
			t.Errorf("End(%s) saved status %s, error %v, finished_at %v; want finished_at set %v", c.status, step.Status, step.Error, step.FinishedAt, c.finished)
		}
	}
}
