package skill

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/pascalandy/nicely/sdk/extension"
)

func TestTheExtensionMatchesItsManifest(t *testing.T) {
	if err := extension.Check(Extension); err != nil {
		t.Fatal(err)
	}
}

// The nicely skill quotes the install rule of "Release and install" in the
// contract, which stays its one source. A change of the rule fails this test
// until the quote follows it.
func TestTheNicelySkillQuotesTheInstallRule(t *testing.T) {
	contract, err := os.ReadFile("../../docs/north-star/contract/extensions.md")
	if err != nil {
		t.Fatal(err)
	}
	_, section, ok := strings.Cut(string(contract), "\n### Release and install\n")
	if !ok {
		t.Fatal("extensions.md has no section Release and install")
	}
	section, _, _ = strings.Cut(section, "\n#")
	rule := strings.Split(section, "\n")
	var quotes []string
	for line := range strings.SplitSeq(string(nicely), "\n") {
		if quote, ok := strings.CutPrefix(line, "> "); ok {
			quotes = append(quotes, quote)
			if !slices.Contains(rule, quote) {
				t.Errorf("the nicely skill quotes %q, which Release and install in extensions.md lacks", quote)
			}
		}
	}
	if !slices.ContainsFunc(quotes, func(quote string) bool { return strings.HasPrefix(quote, "A user installs a fixed version with ") }) {
		t.Error("the nicely skill does not quote the install rule of Release and install in extensions.md")
	}
}
