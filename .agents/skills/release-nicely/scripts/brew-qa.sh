#!/usr/bin/env bash
# brew-qa.sh installs ncly from the Homebrew tap on a Mac, as a user would, or upgrades it.
# It checks the installed version, then types "ncly <Tab>" in a clean zsh.
# Usage: bash brew-qa.sh vX.Y.Z. ncly stays installed, so the next release tests the upgrade.
set -Eeuo pipefail
want="${1:?usage: brew-qa.sh vX.Y.Z}"
export PATH="/opt/homebrew/bin:/opt/homebrew/sbin:/usr/bin:/bin:/usr/sbin:/sbin" HOMEBREW_NO_ENV_HINTS=1
brew update --quiet
# brew install upgrades an older ncly, and fails when the formula cannot write its completions.
brew install pascalandy/tap/ncly
got="$(ncly --version)"
[[ "${got}" == "ncly ${want}" ]] || {
	printf 'brew-qa: ncly --version printed "%s", want "ncly %s"\n' "${got}" "${want}" >&2
	exit 1
}
expect "${BASH_SOURCE[0]%/*}/brew-qa-tab.exp" "$(brew --prefix)"
printf 'brew-qa: %s installs on macOS %s, and zsh completes it\n' "${got}" "$(sw_vers -productVersion)"
