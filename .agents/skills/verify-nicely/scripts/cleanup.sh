# shellcheck shell=bash
# shellcheck disable=SC2154 # verify.sh sets marker and test_service before it sources this file
# cleanup.sh holds the steps of verify.sh that end a run, which verify.sh sources:
#
#   verify.sh quote <run> [<feature>]   print transcripts with placeholders for home paths
#   verify.sh cleanup <run>             remove the scratch folder, keep the evidence

# logout_fake_key logs the fake key of the auth feature out of the real
# keychain, then checks that status no longer finds it. Both calls go to the
# transcript evidence/cleanup.txt. It fails when the key may remain.
logout_fake_key() {
	local run="$1" code=0 answer
	# The key must come from the keychain, never from the caller's variable.
	unset VERIFYNICELY_API_KEY
	answer="$(step "$run" cleanup auth logout "$test_service" --force --json)" || code=$?
	# 0 removed the key, and 2 with NOT_FOUND found none to remove.
	case "$code:$answer" in
	0:* | 2:*'"errors":[{"code":"NOT_FOUND"'*) ;;
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
	log "logged the fake key of $test_service out of the keychain"
}

cleanup() {
	local run
	run="$(run_folder "${1:-}")"
	# The auth feature leaves its fake key in the real keychain, after a failed
	# run too. The scratch folder stays until the key is gone, so that a later
	# cleanup can try again.
	if [[ -f "$run/auth-prepared" && -d "$run/scratch" ]]; then
		logout_fake_key "$run" || die "kept the scratch folder of $run: run cleanup again once the keychain answers"
	fi
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
