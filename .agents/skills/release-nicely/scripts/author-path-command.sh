#!/usr/bin/env bash
# Sourced by author-path.sh to rehearse example replacement and command creation.
# Markdown backticks in patterns and fixture text are literal.
# repo and scratch come from the sourcing script's caller.
# shellcheck disable=SC2016,SC2154
fct_catalog() {
	cat >>"${repo}/locales/en.toml" <<'EOF'

["probe.greet.summary"]
description = "Help summary of the author path command."
other = "Give a greeting"
EOF
}

fct_handler() {
	cat >"${repo}/greet.go" <<'EOF'
package probe

import (
	"context"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
)

func greet(context.Context, extension.Call) contract.Outcome {
	return contract.Outcome{Data: map[string]any{"message": "pending"}}
}
EOF
	awk '{ print; if ($0 ~ /Handlers: map\[string\]extension.Handler\{/) print "\t\t\"probe greet\": greet," }' \
		"${repo}/extension.go" >"${repo}/extension.go.new"
	mv "${repo}/extension.go.new" "${repo}/extension.go"
}

fct_scenario() {
	cat >"${repo}/testdata/script/author_path.txtar" <<'EOF'
# The placeholder compiles, but does not give the specified greeting.
# spec: greeting
ncly probe greet --json
stdout '"message":"ready"'
EOF
}

# Replay the numbered source instructions up to the first scenario.
# Unknown steps fail explicitly, so changed documentation cannot silently pass.
fct_command_order() {
	local step saw_spec=0 saw_manifest=0 saw_catalog=0 saw_handler=0 saw_scenario=0
	while IFS= read -r step; do
		printf 'author-path: follow %s\n' "${step}"
		case "${step}" in
		*'`spec.md`'*)
			cat >>"${repo}/spec.md" <<'EOF'

## ncly probe greet

Success means that the answer gives the greeting ready.

- `greeting`: the message is ready
EOF
			saw_spec=1
			;;
		*'`manifest.toml`'*)
			cat >>"${repo}/manifest.toml" <<'EOF'

[[commands]]
path = "probe greet"
summary = "probe.greet.summary"
examples = ["ncly probe greet --json"]
effects = {}

[commands.result.schema]
type = "object"
required = ["message"]
properties = { message = { type = "string" } }
EOF
			saw_manifest=1
			;;
		*'`locales/en.toml`'*)
			fct_catalog
			saw_catalog=1
			;;
		*'placeholder handler'*)
			fct_handler
			saw_handler=1
			;;
		*'`testdata/script/`'*)
			fct_scenario
			saw_scenario=1
			break
			;;
		*)
			fct_fail "#324: unsupported Change a command step: ${step}"
			return
			;;
		esac
	done <"${scratch}/change-steps.txt"
	if [[ "${saw_spec}${saw_manifest}${saw_scenario}" != 111 ]]; then
		fct_fail '#324: Change a command omits the spec, manifest, or scenario step'
		return
	fi
	gofmt -w "${repo}"
	cat >"${repo}/author_path_test.go" <<'EOF'
package probe

import (
	"testing"

	"github.com/pascalandy/nicely/sdk/extension"
)

func TestAuthorPathCheck(t *testing.T) {
	if err := extension.Check(Extension); err != nil {
		t.Fatal(err)
	}
}
EOF
	gofmt -w "${repo}/author_path_test.go"
	if ! fct_run placeholder-check "${repo}" go test -run '^TestAuthorPathCheck$' .; then
		cat "${scratch}/placeholder-check.log" >&2
		fct_fail '#323 #324: documented test order stops at extension.Check before the scenario'
		return
	fi
	if [[ "${saw_catalog}${saw_handler}" != 11 ]]; then
		fct_fail '#324: document the catalog and registered placeholder before the scenario'
		return
	fi
	if fct_run scenario-red "${repo}" just check; then
		fct_fail '#324: the scenario accepts the placeholder without its required behavior'
		return
	fi
	if ! grep -q 'FAIL: TestExtension/author_path' "${scratch}/scenario-red.log" ||
		! grep -q 'no match for.*message.*ready' "${scratch}/scenario-red.log"; then
		cat "${scratch}/scenario-red.log" >&2
		fct_fail '#324: the scenario fails for a reason other than the missing greeting'
		return
	fi
	printf 'author-path: PASS scenario fails for the missing greeting\n'
	sed 's/"message": "pending"/"message": "ready"/' "${repo}/greet.go" >"${repo}/greet.go.new"
	mv "${repo}/greet.go.new" "${repo}/greet.go"
	if ! fct_run scenario-green "${repo}" just check; then
		cat "${scratch}/scenario-green.log" >&2
		fct_fail '#324: the completed greeting does not pass just check'
	fi
}

# Replace the example with one complete command, following Remove the example.
fct_remove_example() {
	local step file saw_files=0 saw_manifest=0 saw_catalog=0 saw_docs=0 saw_handler=0
	awk '/^## Remove the example$/ { active=1; next } active && /^## / { exit } active && /^[0-9]+\. / { print }' \
		"${repo}/AGENTS.md" >"${scratch}/remove-steps.txt"
	while IFS= read -r step; do
		printf 'author-path: follow %s\n' "${step}"
		case "${step}" in
		*'Delete `inspect.go`, `worker/inspect.py`, and `testdata/script/inspect.txtar`'*)
			mkdir "${scratch}/removed"
			for file in inspect.go worker/inspect.py testdata/script/inspect.txtar; do
				mv "${repo}/${file}" "${scratch}/removed/${file##*/}"
			done
			saw_files=1
			;;
		*'Replace the example declaration at `probe inspect`'*'`manifest.toml`'*'`[programs.python3]`'*)
			cat >"${repo}/manifest.toml" <<'EOF'
format_version = 1
contract_version = 1
domain = "probe"

[[commands]]
path = "probe"
summary = "probe.summary"
examples = ["ncly probe greet --json"]

[[commands]]
path = "probe greet"
summary = "probe.greet.summary"
examples = ["ncly probe greet --json"]
effects = {}

[commands.result.schema]
type = "object"
required = ["message"]
properties = { message = { type = "string" } }
EOF
			saw_manifest=1
			;;
		*'Replace the `probe.inspect` catalog entries'*'`probe.summary`'*)
			cat >"${repo}/locales/en.toml" <<'EOF'
["probe.summary"]
description = "Help summary of the author path domain."
other = "Give greetings"
EOF
			fct_catalog
			saw_catalog=1
			;;
		*'Remove the `ncly probe inspect` section'*'`spec.md`'*'`vision.md`, `qa.md`, `skill/SKILL.md`'*'`AGENTS.md`'*'`acceptance.sh`'*)
			cat >"${repo}/spec.md" <<'EOF'
# Probe spec

## ncly probe greet

Success means that the answer gives the greeting ready.

## Acceptance

- `greeting`: the message is ready
EOF
			cat >"${repo}/skill/SKILL.md" <<'EOF'
---
name: probe
description: Use when an agent needs the author path greeting.
---

Run `ncly probe greet --json` to get the greeting.
EOF
			printf '# Probe\n\nGive a greeting through `ncly probe greet`.\n' >"${repo}/vision.md"
			if [[ -f "${repo}/qa.md" ]]; then
				printf '# Probe QA\n\nRequire the greeting ready.\n' >"${repo}/qa.md"
			fi
			if [[ -f "${repo}/acceptance.sh" ]]; then
				printf '#!/usr/bin/env bash\nset -Eeuo pipefail\ngo tool ncly probe greet --json\n' >"${repo}/acceptance.sh"
			fi
			sed 's/of `inspect`/of `greet`/' "${repo}/AGENTS.md" >"${repo}/AGENTS.md.new"
			mv "${repo}/AGENTS.md.new" "${repo}/AGENTS.md"
			saw_docs=1
			;;
		*'Remove its handler entry from `extension.go`'*'`//go:embed worker`'*'`Resources`'*'`gofmt -w .`'*)
			awk '!/^\/\/go:embed worker$/ && !/^var workers embed.FS$/ && !/Resources: workers,/ && !/"probe inspect": inspect,/' \
				"${repo}/extension.go" >"${repo}/extension.go.new"
			mv "${repo}/extension.go.new" "${repo}/extension.go"
			fct_handler
			sed 's/"message": "pending"/"message": "ready"/' "${repo}/greet.go" >"${repo}/greet.go.new"
			mv "${repo}/greet.go.new" "${repo}/greet.go"
			# The replacement must still have the shared helper that inspect.go once held.
			printf '\nvar _ = usage\n' >>"${repo}/greet.go"
			gofmt -w "${repo}"
			saw_handler=1
			;;
		*)
			fct_fail "#323 #324: unsupported Remove the example step: ${step}"
			return
			;;
		esac
	done <"${scratch}/remove-steps.txt"
	if [[ "${saw_files}${saw_manifest}${saw_catalog}${saw_docs}${saw_handler}" != 11111 ]]; then
		fct_fail '#323 #324: Remove the example omits a required step for files, manifest, catalog, documentation, or handlers'
		return
	fi
	fct_scenario
	if ! fct_run remove-example "${repo}" just check; then
		cat "${scratch}/remove-example.log" >&2
		fct_fail '#323 #324: example removal breaks the replacement command'
	fi
}
