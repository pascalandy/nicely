package platform

import "os"

// flush writes a file's data and metadata to the device, with fsync.
func flush(f *os.File) error { return f.Sync() }
