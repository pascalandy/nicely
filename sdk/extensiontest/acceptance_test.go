package extensiontest_test

import (
	"os"
	"strings"
	"testing"
)

// acceptanceHost stands for an extension that writes <name>.txt from
// <name>.md. It drops the line author: of the header and keeps the line
// encoding:, which a reader needs. Its answer gives the line count of the
// input. BUG=prefix answers ten times that count, as the bug of 86d24e83
// hid. BUG=raw keeps the raw header, with its private text, as in #425.
const acceptanceHost = `#!/bin/sh
input=
for arg in "$@"; do
  case "$arg" in
    -*) ;;
    *) input=$arg ;;
  esac
done
if [ ! -f "$input" ]; then
  echo '{"ok":false,"contract_version":1,"retry_safe":true,"errors":[{"code":"USAGE_INVALID","message":"No input.","hint":"Fix it."}]}' >&2
  exit 2
fi
out=${input%.md}.txt
lines=$(awk 'END { print NR }' "$input")
case "$BUG" in
  prefix) lines=${lines}0 ;;
  raw) cp "$input" "$out" ;;
  *) grep -v '^author:' "$input" > "$out" ;;
esac
echo "{\"ok\":true,\"contract_version\":1,\"file\":\"$out\",\"sha256\":\"none\",\"lines\":$lines}"
`

// acceptanceSetup puts acceptanceHost in place of the installed extension.
const acceptanceSetup = "grant example\n" +
	"rm $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\ncp host.sh $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\nchmod 755 $XDG_DATA_HOME/nicely/extensions/ncly-video-convert\n"

// counted is the case of the line count. Its bad variant is an answer that
// gives 30 lines for the 3 lines of notes.md.
const counted = "ncly video convert notes.md --json\ncp stdout answer.json\nreads answer.json reference bad.json answer lines\n"

// header is the case of a block that a reader needs, with private text.
// show.sh reads the output as a reader does: it decodes the text after the
// header from the encoding that the header declares. Its bad variant keeps
// the raw header, as a handler that keeps the raw block writes it.
const header = "ncly video convert header.md --json\nreads -case 'Keeps data that a reader needs' -private 'Ana Lima' header.txt reference.txt bad.txt sh show.sh\n"

// acceptanceFiles are the made inputs, the reader, the references, and
// the bad variants of the cases.
const acceptanceFiles = "-- notes.md --\none\ntwo\nthree\n" +
	"-- reference --\nlines: 3\n" +
	"-- bad.json --\n{\"ok\":true,\"contract_version\":1,\"lines\":30}\n" +
	"-- header.md --\nencoding: iso-8859-1\nauthor: Ana Lima\n\ncaf\xe9\n" +
	"-- show.sh --\nencoding=$(sed -n 's/^encoding: //p' \"$1\")\nsed '1,/^$/d' \"$1\" | iconv -f \"${encoding:-utf-8}\" -t utf-8\n" +
	"-- reference.txt --\ncafé\n" +
	"-- bad.txt --\nencoding: iso-8859-1\nauthor: Ana Lima\n\ncaf\xe9\n" +
	"-- host.sh --\n" + acceptanceHost

// matrix is ## Case matrix of spec.md, which names the case of a block that
// a reader needs.
const matrix = "\n## Case matrix\n\n| Family | Source | Accepted | Refused |\n|---|---|---|---|\n" +
	"| The header keeps what a reader needs | Output: Keeps data that a reader needs | `header` | None |\n"

// acceptanceFixture gives standalone-named the matrix spec, a schema with a
// line count, and one scenario for each script of scripts.
func acceptanceFixture(t *testing.T, spec string, scripts map[string]string) (string, error) {
	t.Helper()
	edits := []edit{
		{"manifest.toml", "requires = { keys = [\"example\"] }", "requires = { keys = [\"example\"], bins = [\"awk\"] }"},
		{
			"manifest.toml", "properties = { file = { type = \"string\" }, sha256 = { type = \"string\" } }",
			"properties = { file = { type = \"string\" }, sha256 = { type = \"string\" }, lines = { type = \"integer\" } }\n\n[programs.awk]",
		},
		{"spec.md", "The direct executable uses that sentinel without a host grant.", "The direct executable uses that sentinel without a host grant.\n" + spec},
	}
	for name, script := range scripts {
		edits = append(edits, edit{file: "testdata/script/" + name + ".txtar", new: acceptanceSetup + script + acceptanceFiles})
	}
	return goTest(t, "standalone-named", os.Getenv("PATH"), edits...)
}

// TestAcceptanceRejectsPastMistakes proves the past mistakes of the card
// COR-104-T5. At the parent of 86d24e83, acceptance.sh matched "lines":3 as
// a prefix, so an answer of 30 lines passed. In #425, no test file held
// private text in a block that a reader needs, and two builds kept it. The
// check of that parent still passes on both outputs, and reads refuses them.
func TestAcceptanceRejectsPastMistakes(t *testing.T) {
	out, err := acceptanceFixture(t, matrix, map[string]string{
		"prefix":  "env BUG=prefix\n" + "ncly video convert notes.md --json\nstdout '\"lines\":3'\ncp stdout answer.json\nreads answer.json reference bad.json answer lines\n",
		"private": "env BUG=raw\n" + header,
	})
	t.Logf("go test: %v\n%s", err, out)
	for _, want := range []string{
		`reads: answer gives "lines: 30\n" for answer.json, and the reference reference holds "lines: 3\n"`,
		`reads: header.txt holds the private text "Ana Lima"`,
	} {
		if err == nil || !strings.Contains(out, want) {
			t.Errorf("want rejection with %q, got %v", want, err)
		}
	}
}

// TestAcceptanceRefusesWeakCases proves that each part of a case can fail.
// It refuses these cases:
//   - a bad variant that passes
//   - a bad variant of a private case without the private text
//   - a reader that is the extension or the wrapped program
//   - a read before any call
//   - a path outside the scenario folder
//   - a private case without private text
//   - an Output case of the matrix that no case reads
func TestAcceptanceRefusesWeakCases(t *testing.T) {
	spec := matrix + "| The output keeps the encoding | Output: Loses data that changes the meaning | `header` | None |\n"
	out, err := acceptanceFixture(t, spec, map[string]string{
		"weak":      "ncly video convert notes.md --json\ncp stdout answer.json\nreads answer.json reference reference answer lines\n",
		"clean":     "ncly video convert header.md --json\nreads -case 'Keeps data that a reader needs' -private 'Ana Lima' header.txt reference.txt reference.txt sh show.sh\n",
		"extension": "ncly video convert notes.md --json\ncp stdout answer.json\nreads answer.json reference bad.json ncly-video-convert answer.json\n",
		"wrapped":   "ncly video convert notes.md --json\ncp stdout answer.json\nreads answer.json reference bad.json awk '{ print }'\n",
		"first":     "reads notes.md reference bad.json answer lines\n",
		"outside":   "ncly video convert notes.md --json\nreads $WORK/../notes.md reference bad.json answer lines\n",
		"public":    "ncly video convert header.md --json\nreads -case 'Keeps data that a reader needs' header.txt reference.txt bad.txt sh show.sh\n",
	})
	t.Logf("go test: %v\n%s", err, out)
	for _, want := range []string{
		"reads: the bad variant reference passes, so the reader and the reference cannot refuse a wrong output",
		"reads: the bad variant reference.txt holds no private text, so it cannot prove that the search finds it",
		"reads: the reader ncly-video-convert is the extension or a program that it wraps",
		"reads: the reader awk is the extension or a program that it wraps",
		"reads: no call of ncly or ncly-video-convert ran before it in this scenario",
		"is no path in the folder of the scenario",
		"reads: the case \"Keeps data that a reader needs\" needs the private text of its test file, given with -private",
		"spec.md:18: ## Case matrix names the Output case \"Loses data that changes the meaning\", and no scenario proves it with a passing reads -case",
	} {
		if err == nil || !strings.Contains(out, want) {
			t.Errorf("want rejection with %q, got %v", want, err)
		}
	}
}

// TestAcceptanceRepaired proves that a handler that keeps the count and
// drops the private text passes both cases.
func TestAcceptanceRepaired(t *testing.T) {
	if out, err := acceptanceFixture(t, matrix, map[string]string{"counted": counted, "header": header}); err != nil {
		t.Fatalf("repaired fixture failed: %v\n%s", err, out)
	}
}
