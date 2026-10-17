package main

// ncly loads a Bubbles spinner and Glamour, which the screens of M07 build
// on, beside the Huh of sdk/tui, so terminal.txtar proves that loading them
// asks the terminal nothing. The Lip Gloss under them is the fork that
// go.mod names, which detects the terminal on its first print instead of at
// load. They stay out of sdk/ until a screen needs them, because a module
// that imports the SDK gets the Lip Gloss of its own go.mod, without this
// replace line.
import (
	_ "charm.land/bubbles/v2/spinner"
	_ "charm.land/glamour/v2"
)
