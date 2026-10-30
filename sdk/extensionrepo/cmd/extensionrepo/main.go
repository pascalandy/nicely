// Command extensionrepo runs repository tasks for core through the shared SDK.
package main

import (
	"github.com/pascalandy/nicely/sdk/extension"
	"github.com/pascalandy/nicely/sdk/extensionrepo"
)

func main() {
	extensionrepo.Main(extension.Extension{})
}
