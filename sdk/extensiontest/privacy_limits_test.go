package extensiontest_test

import (
	"strings"
	"testing"
)

const privacyLimitSpec = `
## Case matrix

| Family | Source | Accepted | Refused |
|---|---|---|---|
| Retained SVG metadata | Output: Keeps private data of the input; Output: Keeps data that a reader needs | ` + "`private-metadata`" + ` | None |

**Limits.**

| ID | State | Limit | Accepted |
|---|---|---|---|
| ` + "`private-metadata`" + ` | limit | Retain author, capture date, and location | Pascal, 2026-11-03, issue #462 |
| ` + "`unrelated`" + ` | limit | A different case keeps metadata | Pascal, 2026-11-03, issue #462 |
`

const retainedSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 40 20">
<metadata><author>Ana Lima</author><date>2026-11-03</date><gps>45.5,-73.6</gps></metadata>
<g transform="rotate(90)"><rect width="20" height="10"/></g>
</svg>
`

const privacyReader = `package main

import (
 "encoding/xml"
 "fmt"
 "os"
)

func main() {
 data, err := os.ReadFile(os.Args[1])
 if err != nil { panic(err) }
 var svg struct {
  ViewBox string ` + "`xml:\"viewBox,attr\"`" + `
  Metadata struct {
   Author string ` + "`xml:\"author\"`" + `
   Date string ` + "`xml:\"date\"`" + `
   GPS string ` + "`xml:\"gps\"`" + `
  } ` + "`xml:\"metadata\"`" + `
  Group struct { Transform string ` + "`xml:\"transform,attr\"`" + ` } ` + "`xml:\"g\"`" + `
 }
 if err := xml.Unmarshal(data, &svg); err != nil { panic(err) }
 fmt.Printf("author: %s\ndate: %s\ngps: %s\nviewBox: %s\ntransform: %s\n", svg.Metadata.Author, svg.Metadata.Date, svg.Metadata.GPS, svg.ViewBox, svg.Group.Transform)
}
`

const privacyReference = "author: Ana Lima\ndate: 2026-11-03\ngps: 45.5,-73.6\nviewBox: 0 0 40 20\ntransform: rotate(90)\n"

func privacyScript(reads string) string {
	return "env BUG=raw\nncly video convert privacy.md --json\n" + reads +
		"\n-- privacy.md --\n" + retainedSVG +
		"-- privacy-reference.txt --\n" + privacyReference +
		"-- privacy-bad.txt --\n" + strings.ReplaceAll(retainedSVG, "rotate(90)", "rotate(0)") +
		"-- read-privacy.go --\n" + privacyReader
}

func TestAcceptedPrivacyLimitKeepsIndependentProof(t *testing.T) {
	reads := "reads -case 'Keeps private data of the input' -limit private-metadata privacy.txt privacy-reference.txt privacy-bad.txt go run read-privacy.go\n" +
		"reads -case 'Keeps data that a reader needs' -limit private-metadata privacy.txt privacy-reference.txt privacy-bad.txt go run read-privacy.go\n"
	if out, err := acceptanceFixture(t, privacyLimitSpec, map[string]string{"privacy": privacyScript(reads)}); err != nil {
		t.Fatalf("accepted privacy limit failed: %v\n%s", err, out)
	}
}

func TestPrivacyLimitsRejectMissingOrUnrelatedApproval(t *testing.T) {
	invalid := strings.Replace(privacyLimitSpec, "`private-metadata` | None |", "`private-metadata`, `pending`, `wrong-state`, `duplicate`, `empty`, `placeholder` | None |", 1) +
		"| `pending` | limit | Retain author | |\n" +
		"| `wrong-state` | open | Retain author | Pascal, 2026-11-03, issue #462 |\n" +
		"| `duplicate` | limit | Retain author | Pascal, 2026-11-03, issue #462 |\n" +
		"| `duplicate` | limit | Another meaning | Pascal, 2026-11-03, issue #462 |\n" +
		"| `empty` | limit | | Pascal, 2026-11-03, issue #462 |\n" +
		"| `placeholder` | limit | Retain author | None. |\n"
	const tail = " privacy.txt privacy-reference.txt privacy-bad.txt go run read-privacy.go\n"
	scripts := map[string]string{}
	for _, id := range []string{"pending", "wrong-state", "duplicate", "empty", "placeholder", "missing", "unrelated"} {
		scripts[id] = privacyScript("reads -case 'Keeps private data of the input' -limit " + id + tail)
	}
	scripts["strict"] = privacyScript("reads -case 'Keeps private data of the input' -limit private-metadata -private 'Ana Lima'" + tail)
	scripts["opt-in"] = privacyScript("reads -case 'Keeps private data of the input'" + tail)
	scripts["bad"] = privacyScript("reads -case 'Keeps private data of the input' -limit private-metadata privacy.txt privacy-reference.txt privacy.txt go run read-privacy.go\n")
	scripts["nonprivacy"] = privacyScript("reads -case 'Loses data that changes the meaning' -limit private-metadata" + tail)
	scripts["repeated"] = privacyScript("reads -case 'Keeps private data of the input' -limit missing -limit private-metadata" + tail)
	out, err := acceptanceFixture(t, invalid, scripts)
	if err == nil {
		t.Fatal("invalid privacy bindings passed")
	}
	for _, id := range []string{"pending", "wrong-state", "duplicate", "empty", "placeholder", "missing", "unrelated"} {
		if want := `reads: limit "` + id + `" is not accepted`; !strings.Contains(out, want) {
			t.Errorf("missing rejection %q:\n%s", want, out)
		}
	}
	for _, want := range []string{
		`privacy.txt holds the private text "Ana Lima"`,
		`needs the private text of its test file`,
		`reads: the bad variant privacy.txt passes`,
		`reads: -limit needs a privacy Output case`,
		`usage: reads [-case <case>] [-limit <id>]`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing rejection %q:\n%s", want, out)
		}
	}
}

func TestPrivacyLimitsRequireEveryExactCaseBinding(t *testing.T) {
	for name, spec := range map[string]string{
		"another family": strings.Replace(privacyLimitSpec, "\n**Limits.**", "\n| Another family | Output: Keeps private data of the input | `unrelated` | None |\n\n**Limits.**", 1),
		"case in prose":  strings.Replace(privacyLimitSpec, "Output: Keeps private data of the input; Output: Keeps data that a reader needs", "Output: Loses data that changes the meaning; note: Keeps private data of the input; note: Keeps data that a reader needs", 1),
	} {
		t.Run(name, func(t *testing.T) {
			reads := "reads -case 'Keeps private data of the input' -limit private-metadata privacy.txt privacy-reference.txt privacy-bad.txt go run read-privacy.go\n"
			out, err := acceptanceFixture(t, spec, map[string]string{"privacy": privacyScript(reads)})
			want := `reads: limit "private-metadata" is not accepted for every matrix row`
			if err == nil || !strings.Contains(out, want) {
				t.Fatalf("ambiguous case binding passed or failed elsewhere: %v\n%s", err, out)
			}
		})
	}
}
