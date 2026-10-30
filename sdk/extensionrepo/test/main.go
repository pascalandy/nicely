// Command test runs Go tests and prints their collected notices after the package results.
package main

import (
	"os"

	"github.com/pascalandy/nicely/sdk/extensionrepo"
)

func main() {
	os.Exit(extensionrepo.Test(os.Args[1:]))
}
