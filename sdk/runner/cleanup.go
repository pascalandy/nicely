package runner

import (
	"syscall"
	"time"
)

// waitCleanup reserves a tail of the existing deadline for the direct program
// after forcing its descendants to stop. The final deadline never moves.
func waitCleanup(exited <-chan struct{}, program int, ends time.Time, reserve time.Duration) {
	reserve = min(max(reserve, 0), Cleanup)
	if reserve > 0 {
		select {
		case <-exited:
			return
		case <-time.After(time.Until(ends) - reserve):
			signalTree(syscall.SIGKILL, program)
		}
	}
	select {
	case <-exited:
	case <-time.After(time.Until(ends)):
		signalTree(syscall.SIGKILL, 0)
		<-exited
	}
}
