package platform

import (
	"os"

	"golang.org/x/sys/unix"
)

// flush writes a file's data and metadata to the device: fsync, then
// F_FULLFSYNC, because on macOS fsync alone leaves the data in the drive's
// cache, as Apple's fsync documentation says, and Record format in
// contract.md asks for both.
func flush(f *os.File) error {
	if err := unix.Fsync(int(f.Fd())); err != nil {
		return err
	}
	_, err := unix.FcntlInt(f.Fd(), unix.F_FULLFSYNC, 0)
	return err
}
