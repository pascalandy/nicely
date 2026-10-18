#!/usr/bin/env bash
set -Eeuo pipefail
trap 'printf "%s: line %s failed: %s\n" "${BASH_SOURCE[0]##*/}" "${LINENO}" "${BASH_COMMAND}" >&2' ERR

# Proves release-verify on a small release published here. Git reads local bare repositories,
# a gh stub serves the release files, and no case reaches the network. The fixture tag carries
# a copy of the real release-artifacts, so each case runs the real verifier.

fct_sha256() {
	if command -v sha256sum >/dev/null; then
		sha256sum "$@"
	else
		shasum -a 256 "$@"
	fi
}

# Runs the toolchain as the published fixture needs it, outside the caller's Go settings.
fct_go() {
	env -i "PATH=${PATH}" "HOME=${test_root}/home" "TMPDIR=${test_root}/tmp" "GOPATH=${go_path}" "GOCACHE=${go_cache}" "GOMODCACHE=${go_modcache}" GOENV=off GOFLAGS=-mod=readonly GOTOOLCHAIN=local CGO_ENABLED=0 "$@"
}

fct_git() {
	GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 git -c init.templateDir= -c init.defaultBranch=main -c core.hooksPath=/dev/null -c commit.gpgsign=false -c tag.gpgsign=false -c user.name=Test -c user.email=test@example.com "$@"
}

fct_write() {
	mkdir -p "${1%/*}"
	printf '%s' "${2}" >"${1}"
}

fct_source() {
	local source="${1}" file
	fct_write "${source}/go.mod" $'module github.com/pascalandy/nicely\n\ngo 1.27.0\n\nrequire example.com/dependency v0.0.0\n\nreplace example.com/dependency => ./dependency\n'
	fct_write "${source}/LICENSE" $'MIT license fixture\n'
	fct_write "${source}/cmd/ncly/main.go" $'package main\n\nimport (\n\t"fmt"\n\t"os"\n\n\t_ "example.com/dependency"\n\t"github.com/pascalandy/nicely/internal/cli"\n)\n\nfunc main() {\n\tif len(os.Args) == 3 && os.Args[1] == "completion" {\n\t\tfmt.Printf("%s completion\\n", os.Args[2])\n\t\treturn\n\t}\n\tfmt.Println("ncly " + cli.Version())\n}\n'
	fct_write "${source}/internal/cli/version.go" $'package cli\n\nvar version = "dev"\n\nfunc Version() string { return version }\n'
	fct_write "${source}/dependency/go.mod" $'module example.com/dependency\n\ngo 1.27.0\n'
	fct_write "${source}/dependency/dependency.go" $'package dependency\n'
	fct_write "${source}/dependency/LICENSE" $'dependency license\n'
	fct_write "${source}/dependency/NOTICE" $'dependency attribution\n'
	# The workflow takes the toolchain line before the go line, as setup-go does.
	fct_write "${source}/tools/go.mod" $'module github.com/pascalandy/nicely/tools\n\ngo 1.27.0\n\ntoolchain go1.27.1\n\nrequire github.com/google/go-licenses/v2 v2.0.0\n\nreplace github.com/google/go-licenses/v2 => ./licenses\n'
	# A stand-in for go-licenses: save copies the license and the notice of the dependency.
	fct_write "${source}/tools/licenses/go.mod" $'module github.com/google/go-licenses/v2\n\ngo 1.27.0\n'
	fct_write "${source}/tools/licenses/main.go" $'package main\n\nimport (\n\t"os"\n\t"path/filepath"\n\t"strings"\n)\n\nfunc main() {\n\tdir := ""\n\tfor _, arg := range os.Args[1:] {\n\t\tif value, ok := strings.CutPrefix(arg, "--save_path="); ok {\n\t\t\tdir = filepath.Join(value, "example.com/dependency")\n\t\t}\n\t}\n\tif len(os.Args) < 2 || os.Args[1] != "save" || dir == "" {\n\t\tos.Exit(2)\n\t}\n\tif err := os.MkdirAll(dir, 0o755); err != nil {\n\t\tpanic(err)\n\t}\n\tfor _, name := range []string{"LICENSE", "NOTICE"} {\n\t\tdata, err := os.ReadFile(filepath.Join("dependency", name))\n\t\tif err != nil {\n\t\t\tpanic(err)\n\t\t}\n\t\tif err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {\n\t\t\tpanic(err)\n\t\t}\n\t}\n}\n'
	mkdir -p "${source}/tools/cmd/release-artifacts"
	for file in "${repo_root}"/tools/cmd/release-artifacts/*.go; do
		if [[ "${file}" != *_test.go ]]; then
			cp "${file}" "${source}/tools/cmd/release-artifacts/"
		fi
	done
}

# Packs one binary archive from its staging folder, as GoReleaser lays it out.
fct_pack() {
	(cd "${1}" && COPYFILE_DISABLE=1 tar -czf "${2}" ncly LICENSE completions THIRD_PARTY_LICENSES)
}

fct_checksums() {
	(cd "${1}" && fct_sha256 ncly_0.0.1_*.tar.gz >ncly_0.0.1_checksums.txt)
}

fct_publish() {
	local source="${1}" published="${2}" target stage
	mkdir -p "${published}"
	for target in darwin_amd64 darwin_arm64 linux_amd64 linux_arm64; do
		stage="${test_root}/stage/${target}"
		mkdir -p "${stage}/completions" "${stage}/THIRD_PARTY_LICENSES/example.com/dependency"
		(cd "${source}" && fct_go env "GOOS=${target%_*}" "GOARCH=${target#*_}" go build -trimpath -buildvcs=false -ldflags='-s -w -X github.com/pascalandy/nicely/internal/cli.version=v0.0.1' -o "${stage}/ncly" ./cmd/ncly)
		cp "${source}/LICENSE" "${stage}/LICENSE"
		printf 'bash completion\n' >"${stage}/completions/ncly.bash"
		printf 'zsh completion\n' >"${stage}/completions/_ncly"
		printf 'fish completion\n' >"${stage}/completions/ncly.fish"
		cp "${source}/dependency/LICENSE" "${source}/dependency/NOTICE" "${stage}/THIRD_PARTY_LICENSES/example.com/dependency/"
		fct_pack "${stage}" "${published}/ncly_0.0.1_${target}.tar.gz"
	done
	mkdir -p "${test_root}/stage/source"
	fct_git -C "${source}" archive --format=tar --prefix=ncly-0.0.1/ v0.0.1 | tar -xf - -C "${test_root}/stage/source"
	mkdir -p "${test_root}/stage/source/ncly-0.0.1/THIRD_PARTY_LICENSES/example.com/dependency"
	cp "${source}/dependency/LICENSE" "${source}/dependency/NOTICE" "${test_root}/stage/source/ncly-0.0.1/THIRD_PARTY_LICENSES/example.com/dependency/"
	(cd "${test_root}/stage/source" && COPYFILE_DISABLE=1 tar -czf "${published}/ncly_0.0.1_source.tar.gz" ncly-0.0.1)
	fct_checksums "${published}"
}

# Gives a case its own copy of the published files.
fct_release() {
	local dist="${test_root}/cases/${1}"
	mkdir -p "${dist}"
	cp "${test_root}/published"/* "${dist}/"
	printf '%s\n' "${dist}"
}

# Commits a formula, or none, to the main branch of the bare Homebrew tap.
fct_tap() {
	if [[ -n "${1}" ]]; then
		mkdir -p "${test_root}/tap-work/Formula"
		cp "${1}" "${test_root}/tap-work/Formula/ncly.rb"
		fct_git -C "${test_root}/tap-work" add Formula/ncly.rb
	else
		fct_git -C "${test_root}/tap-work" rm --quiet Formula/ncly.rb
	fi
	fct_git -C "${test_root}/tap-work" commit --quiet --allow-empty -m "formula ${2}"
	fct_git -C "${test_root}/tap-work" push --quiet origin HEAD:main
	tap_commit="$(git -C "${test_root}/tap.git" rev-parse main)"
}

fct_state() {
	git -C "${caller}" rev-parse HEAD
	git -C "${caller}" for-each-ref
	git -C "${caller}" status --porcelain=v1 --ignored --untracked-files=all
	git -C "${test_root}/origin.git" for-each-ref
	git -C "${test_root}/tap.git" for-each-ref
}

# Runs release-verify with the fixture in place of GitHub, and checks its exit code.
fct_verify() {
	local expected="${1}" dist="${2}" status=0
	shift 2
	: >"${test_root}/calls"
	: >"${test_root}/toolchains"
	fct_state >"${test_root}/state.before"
	(cd "${caller}" && env "PATH=${test_root}/bin:${PATH}" "TMPDIR=${test_root}/tmp" "NCLY_TEST_DIST=${dist}" "NCLY_TEST_CALLS=${test_root}/calls" \
		GIT_ALLOW_PROTOCOL=file ${hook_env[@]+"${hook_env[@]}"} bash "${script}" "$@") >"${test_root}/stdout" 2>"${test_root}/stderr" || status=$?
	fct_state >"${test_root}/state.after"
	if [[ "${status}" != "${expected}" ]]; then
		cat "${test_root}/stdout" "${test_root}/stderr" >&2
		printf 'Expected exit %s, got %s\n' "${expected}" "${status}" >&2
		exit 1
	fi
	if ! cmp -s "${test_root}/state.before" "${test_root}/state.after"; then
		diff "${test_root}/state.before" "${test_root}/state.after" >&2 || true
		printf 'release-verify changed the checkout or a repository it read\n' >&2
		exit 1
	fi
}

# Fails unless the output of the last run holds the line, or the text with --text.
fct_says() {
	local mode=-x
	if [[ "${1}" == --text ]]; then
		mode=""
		shift
	fi
	if ! grep -q -F ${mode} -e "${1}" "${test_root}/stdout" "${test_root}/stderr"; then
		cat "${test_root}/stdout" "${test_root}/stderr" >&2
		printf 'Missing from the output: %s\n' "${1}" >&2
		exit 1
	fi
}

fct_never_says() {
	if grep -q -i -F -e "${1}" "${test_root}/stdout" "${test_root}/stderr"; then
		cat "${test_root}/stdout" "${test_root}/stderr" >&2
		printf 'Unexpected in the output: %s\n' "${1}" >&2
		exit 1
	fi
}

fct_main() {
	local script test_root repo_root caller hook_env=() real_git real_go go_path go_cache go_modcache go_version tag_commit tap_commit dist host other formula
	repo_root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
	script="${repo_root}/scripts/release-verify.sh"
	test_root="$(mktemp -d "${TMPDIR:-/tmp}/ncly-release-verify-test.XXXXXXXX")"
	readonly NCLY_RELEASE_VERIFY_TEST_ROOT="${test_root}"
	trap 'rm -rf "${NCLY_RELEASE_VERIFY_TEST_ROOT}"' EXIT
	test_root="$(cd "${test_root}" && pwd -P)"
	mkdir -p "${test_root}/bin" "${test_root}/home" "${test_root}/tmp"
	real_git="$(command -v git)"
	real_go="$(command -v go)"
	go_path="$(go env GOPATH)"
	go_cache="$(go env GOCACHE)"
	go_modcache="$(go env GOMODCACHE)"
	go_version="$(fct_go go env GOVERSION)"
	go_version="${go_version%%[ -]*}"
	# The release keeps whatever toolchain the checks run, so the fixture builds and the
	# verifier rebuilds with one compiler. The log keeps the toolchain that each call asked for.
	cat >"${test_root}/bin/go" <<SCRIPT
#!/usr/bin/env bash
printf '%s\n' "\${GOTOOLCHAIN:-}" >>'${test_root}/toolchains'
GOTOOLCHAIN=local exec '${real_go}' "\$@"
SCRIPT
	# Git reads the bare repositories in place of GitHub, and GIT_ALLOW_PROTOCOL blocks any other URL.
	cat >"${test_root}/bin/git" <<SCRIPT
#!/usr/bin/env bash
args=()
for arg in "\$@"; do
	case "\${arg}" in
	https://github.com/pascalandy/nicely.git) args+=('file://${test_root}/origin.git') ;;
	https://github.com/pascalandy/homebrew-tap.git) args+=('file://${test_root}/tap.git') ;;
	*) args+=("\${arg}") ;;
	esac
done
exec '${real_git}' "\${args[@]}"
SCRIPT
	cat >"${test_root}/bin/gh" <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail
printf '%s\n' "$*" >>"${NCLY_TEST_CALLS}"
if [[ $# -ne 7 || "${1} ${2}" != "release download" || "${4} ${5} ${6}" != "--repo pascalandy/nicely --dir" ]]; then
	printf 'unexpected gh call: %s\n' "$*" >&2
	exit 1
fi
cp "${NCLY_TEST_DIST}"/* "${7}/"
SCRIPT
	chmod 755 "${test_root}/bin/git" "${test_root}/bin/go" "${test_root}/bin/gh"

	fct_source "${test_root}/source"
	fct_git -C "${test_root}/source" init --quiet
	fct_git -C "${test_root}/source" add .
	fct_git -C "${test_root}/source" commit --quiet -m source
	fct_git -C "${test_root}/source" tag v0.0.1
	tag_commit="$(git -C "${test_root}/source" rev-parse HEAD)"
	fct_git clone --quiet --bare "${test_root}/source" "${test_root}/origin.git"
	# A branch that bears the name of the tag must not stand in for the tag.
	fct_git -C "${test_root}/source" push --quiet "${test_root}/origin.git" "$(fct_git -C "${test_root}/source" commit-tree -p HEAD -m decoy 'HEAD^{tree}'):refs/heads/v0.0.1"
	fct_publish "${test_root}/source" "${test_root}/published"
	mkdir -p "${test_root}/rendered"
	cp "${test_root}/published"/* "${test_root}/rendered/"
	(cd "${test_root}/source/tools" && fct_go go run ./cmd/release-artifacts formula v0.0.1 "${test_root}/rendered")
	formula="${test_root}/rendered/Formula/ncly.rb"
	fct_git init --quiet --bare "${test_root}/tap.git"
	fct_git clone --quiet "${test_root}/tap.git" "${test_root}/tap-work" 2>/dev/null
	caller="${test_root}/caller"
	fct_git init --quiet "${caller}"
	printf 'caller data\n' >"${caller}/sentinel.txt"
	fct_git -C "${caller}" add sentinel.txt
	fct_git -C "${caller}" commit --quiet -m caller

	fct_verify 2 "${test_root}/published" v01.2.3
	fct_says 'Usage: just release-verify vX.Y.Z'
	[[ ! -s "${test_root}/calls" ]]

	fct_tap "${formula}" match
	# A git hook can export GIT_DIR, and release-verify must still leave the caller alone.
	hook_env=("GIT_DIR=${caller}/.git" "GIT_WORK_TREE=${caller}")
	fct_verify 0 "${test_root}/published" v0.0.1
	hook_env=()
	fct_says 'Tag: v0.0.1'
	fct_says "Commit: ${tag_commit}"
	fct_says 'Go: go1.27.1, from tools/go.mod at v0.0.1'
	fct_says 'Assets: verified'
	fct_says "Homebrew tap commit: ${tap_commit}"
	fct_says 'Homebrew tap: Formula/ncly.rb matches v0.0.1'
	fct_says 'Result: assets verified, Homebrew tap matches'
	fct_says --text "Release verification files retained at ${test_root}/tmp/ncly-release-verify."
	[[ "$(cat "${test_root}/calls")" == "release download v0.0.1 --repo pascalandy/nicely --dir ${test_root}/tmp/ncly-release-verify."*/dist ]]
	grep -q -x go1.27.1 "${test_root}/toolchains"
	if grep -v -x -e '' -e go1.27.1 "${test_root}/toolchains"; then
		printf 'release-verify asked for another toolchain than tools/go.mod names\n' >&2
		exit 1
	fi

	# The Homebrew tap gets its own result: a newer release there leaves the assets verified.
	sed 's/0\.0\.1/0.0.2/g' "${formula}" >"${test_root}/next.rb"
	fct_tap "${test_root}/next.rb" next
	fct_verify 1 "${test_root}/published" v0.0.1
	fct_says 'Assets: verified'
	fct_says "Homebrew tap commit: ${tap_commit}"
	fct_says 'Homebrew tap: Formula/ncly.rb announces v0.0.2'
	fct_says 'Result: assets verified, Homebrew tap on another version'
	fct_never_says 'corrupt'
	fct_never_says 'failed'

	sed 's/sha256 "[0-9a-f]*"/sha256 "'"$(printf '0%.0s' {1..64})"'"/' "${formula}" >"${test_root}/stale.rb"
	fct_tap "${test_root}/stale.rb" stale
	fct_verify 1 "${test_root}/published" v0.0.1
	fct_says 'Assets: verified'
	fct_says "Homebrew tap commit: ${tap_commit}"
	fct_says 'Homebrew tap: Formula/ncly.rb differs from the formula rendered from the v0.0.1 assets'
	fct_says 'Result: assets verified, Homebrew tap formula wrong'

	fct_tap "" removed
	fct_verify 1 "${test_root}/published" v0.0.1
	fct_says "Homebrew tap commit: ${tap_commit}"
	fct_says 'Homebrew tap: Formula/ncly.rb is missing'
	fct_says 'Result: assets verified, Homebrew tap formula wrong'

	fct_tap "${formula}" restored
	# An archive changed after publication keeps the checksum that the release published.
	dist="$(fct_release altered)"
	printf 'changed license\n' >"${test_root}/stage/linux_amd64/LICENSE"
	fct_pack "${test_root}/stage/linux_amd64" "${dist}/ncly_0.0.1_linux_amd64.tar.gz"
	cp "${test_root}/source/LICENSE" "${test_root}/stage/linux_amd64/LICENSE"
	fct_verify 1 "${dist}" v0.0.1
	fct_says 'Assets: failed: checksum mismatch for ncly_0.0.1_linux_amd64.tar.gz'
	fct_says 'Homebrew tap: not compared, because the assets failed'
	fct_says 'Result: assets failed'

	dist="$(fct_release missing-checksum)"
	grep -v -F ncly_0.0.1_darwin_arm64.tar.gz "${test_root}/published/ncly_0.0.1_checksums.txt" >"${dist}/ncly_0.0.1_checksums.txt"
	fct_verify 1 "${dist}" v0.0.1
	fct_says 'Assets: failed: checksum mismatch for ncly_0.0.1_darwin_arm64.tar.gz'
	fct_says 'Result: assets failed'

	dist="$(fct_release wrong-checksum)"
	other="$(sed -n 's/^\([0-9a-f]*\)  ncly_0.0.1_linux_arm64.tar.gz$/\1/p' "${dist}/ncly_0.0.1_checksums.txt")"
	sed 's/^[0-9a-f]*  ncly_0.0.1_source.tar.gz$/'"${other}"'  ncly_0.0.1_source.tar.gz/' "${test_root}/published/ncly_0.0.1_checksums.txt" >"${dist}/ncly_0.0.1_checksums.txt"
	fct_verify 1 "${dist}" v0.0.1
	fct_says 'Assets: failed: checksum mismatch for ncly_0.0.1_source.tar.gz'
	fct_says 'Result: assets failed'

	# A binary that another compiler built differs from its rebuild without being corrupt.
	host="$(fct_go go env GOOS)_$(fct_go go env GOARCH)"
	other=darwin_arm64
	if [[ "${host}" == darwin_arm64 ]]; then
		other=linux_amd64
	fi
	dist="$(fct_release compiler)"
	perl -0777 -pi -e "s/\\Q${go_version}\\E/go0${go_version#go1}/g" "${test_root}/stage/${other}/ncly"
	fct_pack "${test_root}/stage/${other}" "${dist}/ncly_0.0.1_${other}.tar.gz"
	fct_checksums "${dist}"
	fct_verify 1 "${dist}" v0.0.1
	fct_says "Assets: failed: binary differs from source rebuild: ${other%_*}/${other#*_}"
	fct_says --text "Compiler: ncly_0.0.1_${other}.tar.gz was built with go0${go_version#go1}"
	fct_says 'Result: assets not verified, compiler gap'
	fct_never_says 'corrupt'
	printf 'Release-verify assets, checksums, compiler, and Homebrew tap formula checks passed\n'
}

fct_main
exit 0
