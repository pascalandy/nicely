# shellcheck shell=bash disable=SC2154
# The fixture of release-verify-test.sh, which sources this file: a small release published
# here, and the helpers that run release-verify against it. fct_main of the test sets the
# variables that these functions read, such as test_root, caller, and script.

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
	# shellcheck disable=SC2034 # fct_main reads tap_commit
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
