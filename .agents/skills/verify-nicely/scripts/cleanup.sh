# shellcheck shell=bash
# shellcheck disable=SC2154 # verify.sh sets marker and test_service before it sources this file
# cleanup.sh holds cleanup, quote, and prune, which verify.sh sources:
#
#   verify.sh quote <run> [<feature>]   print transcripts with placeholders for home paths
#   verify.sh cleanup <run>             remove the scratch folder, keep the evidence
#   verify.sh prune --keep <n>          remove the runs before the newest <n>, once cleaned up

# logout_fake_key logs the fake key of the auth feature out of the real
# keychain, then checks that status no longer finds it. Both calls go to the
# transcript evidence/cleanup.txt. It fails when the key may remain.
logout_fake_key() {
	local run="$1" code=0 answer outcome
	# The key must come from the keychain, never from the caller's variable.
	unset VERIFYNICELY_API_KEY
	answer="$(step "$run" cleanup auth logout "$test_service" --force --json)" || code=$?
	# 0 removed the key, and 2 with NOT_FOUND found none to remove.
	case "$code:$answer" in
	0:*) outcome="logged the fake key of $test_service out of the keychain" ;;
	2:*'"errors":[{"code":"NOT_FOUND"'*) outcome="found no fake key of $test_service in the keychain" ;;
	*)
		log "ncly auth logout $test_service exited with $code: see $run/evidence/cleanup.txt"
		return 1
		;;
	esac
	code=0
	answer="$(step "$run" cleanup auth status --json)" || code=$?
	# Only a keychain that answers without the key proves that it is gone:
	# exit 78 with AUTH_MISSING. Another code, such as the 1 of a step that
	# could not record the call, proves nothing.
	case "$code:$answer" in
	78:*'"keychain":"available"'*"{\"service\":\"$test_service\",\"configured\":false,\"source\":\"none\",\"code\":\"AUTH_MISSING\""*) ;;
	*)
		log "ncly auth status still finds a key for $test_service, or no keychain answers: see $run/evidence/cleanup.txt"
		return 1
		;;
	esac
	log "$outcome"
}

cleanup() {
	local run folder
	run="$(run_folder "${1:-}")"
	# The auth feature leaves its fake key in the real keychain, after a failed
	# run too. The scratch folder stays until the key is gone, so that a later
	# cleanup can try again.
	if [[ -f "$run/auth-prepared" && -d "$run/scratch" ]]; then
		logout_fake_key "$run" || die "kept the scratch folder of $run: run cleanup again once the keychain answers"
	fi
	# Go can put read-only module folders under the scratch home.
	for folder in "$run/scratch" "$run/build"; do
		if [[ -d "$folder" ]]; then chmod -R u+w "$folder"; fi
	done
	rm -rf "$run/scratch" "$run/build"
	[[ ! -e "$run/scratch" && ! -e "$run/build" ]] || die "could not remove the scratch folder of $run"
	log "removed the scratch folder; the evidence stays in $run/evidence"
	ls "$run/evidence"
}

quote() {
	local run file
	run="$(run_folder "${1:-}")"
	if [[ -n "${2:-}" ]]; then
		check_feature "$2"
		set -- "$run/evidence/$2.txt"
	else
		set -- "$run/evidence"/*.txt
	fi
	for file in "$@"; do
		[[ -f "$file" ]] || die "no transcript at $file"
		echo "==> evidence/${file##*/}"
		# Literal replacements: the run folder first, because it sits inside HOME.
		RUN="$run" HOME_DIR="$HOME" awk '
			function swap(s, from, to,   out, i) {
				if (from == "") return s
				out = ""
				while ((i = index(s, from)) > 0) {
					out = out substr(s, 1, i - 1) to
					s = substr(s, i + length(from))
				}
				return out s
			}
			{ print swap(swap($0, ENVIRON["RUN"], "<run>"), ENVIRON["HOME_DIR"], "/Users/me") }
		' "$file"
	done
}

# prune keeps the newest <n> runs that start created, by the time that starts their names, and
# removes the others with their evidence. It keeps a run whose scratch folder remains, because
# its cleanup has not ended, such as one that may leave a fake key in the keychain.
prune() {
	local root keep run i
	local -a runs=()
	[[ $# -eq 2 && "$1" == --keep && "$2" =~ ^[0-9]+$ ]] || usage
	keep=$((10#$2))
	root="$(runs_root)"
	for run in "$root"/*; do
		if [[ -f "$run/$marker" && ! -L "$run" ]]; then runs[${#runs[@]}]="$run"; fi
	done
	[[ ${#runs[@]} -gt 0 ]] || log "no run to prune in $root"
	for ((i = ${#runs[@]} - 1; i >= 0; i--)); do
		run="${runs[$i]}"
		if ((keep > 0)); then
			keep=$((keep - 1))
			log "kept $run, one of the newest runs"
		elif [[ -e "$run/scratch" ]]; then
			log "kept $run, whose scratch folder remains: run cleanup on it first"
		else
			rm -rf "${run:?}"
			log "removed $run"
		fi
	done
}
