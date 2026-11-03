package cases

import "os"

// A test can read the environment of its caller on purpose.
var _ = os.Environ()
