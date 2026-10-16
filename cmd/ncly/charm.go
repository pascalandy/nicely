package main

// ncly loads Huh, a Bubbles spinner, and Glamour, which the forms and screens
// of M07 build on, so terminal.txtar proves that loading them asks the
// terminal nothing. The Lip Gloss under them is the fork that go.mod names,
// which detects the terminal on its first print instead of at load. They stay
// out of sdk/ until a form needs them, because a module that imports the SDK
// gets the Lip Gloss of its own go.mod, without this replace line.
import (
	_ "charm.land/bubbles/v2/spinner"
	_ "charm.land/glamour/v2"
	_ "charm.land/huh/v2"
)
