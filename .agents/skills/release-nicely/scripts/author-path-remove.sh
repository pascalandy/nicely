#!/usr/bin/env bash
# Sourced by author-path.sh to rehearse Remove the example of the rendered AGENTS.md.
# Markdown backticks in patterns and fixture text are literal.
# repo and scratch come from the sourcing script's caller, and fct_catalog, fct_handler,
# and fct_scenario from author-path-command.sh.
# shellcheck disable=SC2016,SC2154

# Replace the example with one complete command, following Remove the example.
fct_remove_example() {
	local step file saw_files=0 saw_manifest=0 saw_catalog=0 saw_docs=0 saw_handler=0
	awk '/^## Remove the example$/ { active=1; next } active && /^## / { exit } active && /^[0-9]+\. / { print }' \
		"${repo}/AGENTS.md" >"${scratch}/remove-steps.txt"
	while IFS= read -r step; do
		printf 'author-path: follow %s\n' "${step}"
		case "${step}" in
		'1. Delete `inspect.go`, `worker/inspect.py`, and `testdata/script/inspect.txtar`' | \
			'1. Delete `inspect.go`, `worker/inspect.py`, and `testdata/script/inspect.txtar`. Rewrite `testdata/script/refusal.txtar` for a refusal of your command, or delete it when the command has none')
			mkdir "${scratch}/removed"
			while IFS= read -r file; do
				mv "${repo}/${file}" "${scratch}/removed/${file##*/}"
			done < <(printf '%s\n' "${step}" | awk -F '`' '{ for (i=2; i<NF; i+=2) print $i }')
			saw_files=1
			;;
		'2. Replace the example declaration at `probe inspect` and its examples in `manifest.toml`. Remove `[programs.python3]` when no command needs it' | \
			'2. Replace the example declaration at `probe inspect` and its examples in `manifest.toml`. Remove the `[[errors]]` entry of `PROBE_NOT_TEXT`, and `[programs.python3]` when no command needs it')
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
