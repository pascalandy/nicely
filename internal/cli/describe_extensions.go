package cli

import "os"

// skill returns the folder ncly-<domain>.skill/ beside the manifest of an
// external extension, or a nil *string when it ships none or is bundled.
func (e *extension) skill() *string {
	if e.bundled {
		return nil
	}
	folder := e.path + ".skill"
	if info, err := os.Stat(folder); err != nil || !info.IsDir() {
		return nil
	}
	return &folder
}
