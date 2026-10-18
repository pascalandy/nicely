#!/usr/bin/env bash
# scenario.sh runs testscript scenarios by exact name, for just scenario. It
# runs on macOS and Linux: bash 3.2, and only the flags that BSD and GNU tools share.
#
#   just scenario <name>|<group>/<name>... [--count N]
#
# A name is a .txtar file without its suffix. A name alone must match one
# scenario; <group>/<name> picks it when several groups hold one of that name.
# The groups are those of TestScripts in cmd/ncly: core, then each extension
# and each fixture extension. Each scenario must pass exactly N times, 1 by
# default, in the events of go test -json, or the run fails. See #95.
set -Eeuo pipefail

die() {
	echo "scenario: $*" >&2
	exit 2
}

usage() {
	sed -n '5p' "$0" | sed 's/^# *//'
}

cd "$(dirname "$0")/.."
# A hook or a caller can export GIT_DIR. Fixture commits must see a normal repository.
while IFS= read -r git_variable; do unset "${git_variable}"; done < <(git rev-parse --local-env-vars)

count=1
names=()
while (($# > 0)); do
	case "$1" in
	-h | --help)
		usage
		exit 0
		;;
	--count)
		(($# > 1)) || die "--count needs a value"
		count="$2"
		shift 2
		;;
	--count=*)
		count="${1#--count=}"
		shift
		;;
	-*) die "unknown flag $1" ;;
	*)
		names+=("$1")
		shift
		;;
	esac
done
[[ "$count" =~ ^[1-9][0-9]*$ ]] || die "--count takes a whole number above 0, not \"$count\""
((${#names[@]} > 0)) || die "name at least one scenario: $(usage)"

# The groups and their folders, in the order of TestScripts.
groups=(core)
dirs=(testdata/script)
for dir in extensions/*/testdata/script testdata/extensions/*/testdata/script; do
	[[ -d "$dir" ]] || continue
	group="${dir%/testdata/script}"
	groups+=("${group##*/}")
	dirs+=("$dir")
done

# selected holds each scenario once, as <group>/<name>.
selected=()
for arg in "${names[@]}"; do
	found=()
	if [[ "$arg" == */* ]]; then
		group="${arg%%/*}"
		name="${arg#*/}"
		for i in "${!groups[@]}"; do
			if [[ "${groups[$i]}" == "$group" && "$name" != */* && -f "${dirs[$i]}/$name.txtar" ]]; then
				found+=("$arg")
			fi
		done
	else
		for i in "${!groups[@]}"; do
			if [[ -f "${dirs[$i]}/$arg.txtar" ]]; then found+=("${groups[$i]}/$arg"); fi
		done
	fi
	((${#found[@]} > 0)) || die "no scenario named \"$arg\"; the groups are ${groups[*]}"
	((${#found[@]} == 1)) || die "\"$arg\" names ${#found[@]} scenarios, so name one of them: ${found[*]}"
	case " ${selected[*]-} " in
	*" ${found[0]} "*) ;;
	*) selected+=("${found[0]}") ;;
	esac
done

# regex prints its argument with the regular expression characters escaped.
regex() {
	printf '%s' "$1" | sed 's/[][\.*^$+?(){}|]/\\&/g'
}

# show prints, as go test runs, the result of each scenario run, the whole
# output of any test that fails, and the lines of the package.
# shellcheck disable=SC2016
show='foreach inputs as $e ({buf: {}};
	.print = null
	| if $e.Test == null then .print = $e.Output
	elif $e.Action == "run" then .buf[$e.Test] = ""
	elif $e.Action == "output" then .buf[$e.Test] += $e.Output
	elif $e.Action == "fail" then .print = .buf[$e.Test]
	elif ($e.Action == "pass" or $e.Action == "skip") and ($e.Test | test("^TestScripts/[^/]+/"))
	then .print = "--- \($e.Action | ascii_upcase): \($e.Test) (\($e.Elapsed)s)\n"
	else . end;
	.print // empty)'

work="$(mktemp -d "${TMPDIR:-/tmp}/ncly-scenario.XXXXXX")"
trap 'rm -rf "$work"' EXIT
events="$work/events.jsonl"
: >"$events"
status=0
# One go test per group: a pattern of alternatives per level would also run
# the other groups' scenarios of the same names.
for group in "${groups[@]}"; do
	pattern=""
	for scenario in "${selected[@]}"; do
		[[ "${scenario%%/*}" == "$group" ]] || continue
		pattern="${pattern:+$pattern|}$(regex "${scenario#*/}")"
	done
	[[ -n "$pattern" ]] || continue
	# The timeout grows with the count, so it stays 10 minutes per round of scenarios.
	go test ./cmd/ncly -json -count "$count" -timeout "$((count * 10))m" \
		-run "^TestScripts\$/^$(regex "$group")\$/^($pattern)\$" |
		tee -a "$events" | jq --unbuffered -nrj "$show" || status=1
done

# The verdict counts the scenario runs that the events show, never the files.
# shellcheck disable=SC2016
results="$(jq -rs '[.[] | select((.Test // "" | test("^TestScripts/[^/]+/")) and (.Action == "pass" or .Action == "fail" or .Action == "skip"))]
	| group_by(.Test)[] | "\(.[0].Test) \(map(select(.Action == "pass")) | length) \(map(select(.Action == "fail")) | length) \(map(select(.Action == "skip")) | length)"' "$events")"
for scenario in "${selected[@]}"; do
	line="$(awk -v test="TestScripts/$scenario" '$1 == test' <<<"$results")"
	read -r _ passed failed skipped <<<"${line:-none 0 0 0}"
	if [[ "$passed" == "$count" && "$failed" == 0 && "$skipped" == 0 ]]; then
		echo "scenario: $scenario passed $passed of $count runs"
	else
		echo "scenario: $scenario passed $passed of $count runs, failed $failed, skipped $skipped" >&2
		status=1
	fi
done
while read -r test _; do
	[[ -n "$test" ]] || continue
	case " ${selected[*]} " in
	*" ${test#TestScripts/} "*) ;;
	*)
		echo "scenario: ${test#TestScripts/} ran, but no name selected it" >&2
		status=1
		;;
	esac
done <<<"$results"
exit "$status"
