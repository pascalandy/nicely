// Command ncly-hello runs the hello fixture as an external extension.
package main

import (
	"github.com/pascalandy/nicely/sdk/extension"
	hello "github.com/pascalandy/nicely/testdata/extensions/hello"
)

func main() { extension.Main(hello.Extension) }
