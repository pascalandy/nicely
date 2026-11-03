package hello

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/runner"
)

// staged models a handler whose worker writes into private staging. Cleanup
// takes a little time after Run returns, as removing a real staging tree does.
func staged(ctx context.Context, call extension.Call) contract.Outcome {
	stage, err := os.MkdirTemp(".", ".ncly-stage-")
	if err != nil {
		return stageFailure(call)
	}
	defer func() {
		started := time.Now()
		time.Sleep(100 * time.Millisecond)
		removing := time.Now()
		_ = os.RemoveAll(stage)
		stats, _ := json.Marshal(map[string]int64{
			"delete_ms": time.Since(removing).Milliseconds(), "cleanup_ms": time.Since(started).Milliseconds(),
		})
		_ = os.WriteFile("cleanup.json", stats, 0o600)
		_ = os.WriteFile("cleanup.done", []byte("done\n"), 0o600)
	}()
	if value := os.Getenv("CLEANUP_FILES"); value != "" {
		count, err := strconv.Atoi(value)
		if err != nil || count < 0 {
			return stageFailure(call)
		}
		for i := range count {
			if err := os.WriteFile(filepath.Join(stage, fmt.Sprintf("f%d", i)), nil, 0o600); err != nil {
				return stageFailure(call)
			}
		}
	}
	_, err = runner.Run(ctx, runner.Program{
		Path: "/bin/sh", Args: []string{call.Args[0], stage}, Env: os.Environ(),
	})
	if err != nil {
		return stageFailure(call)
	}
	return contract.Outcome{Data: map[string]string{"stage": stage}}
}

func stageFailure(call extension.Call) contract.Outcome {
	return contract.Outcome{Errors: []contract.Problem{{
		Code:    contract.Runtime,
		Message: call.Printer.T("hello.program"), Hint: "ncly hello staged worker.sh",
	}}}
}
