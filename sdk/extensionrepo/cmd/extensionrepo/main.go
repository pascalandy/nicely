// Command extensionrepo runs repository tasks through the shared SDK, including from a downloaded Nicely module.
package main

import (
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/extensionrepo"
)

func main() {
	extensionrepo.Main(extension.Extension{})
}
