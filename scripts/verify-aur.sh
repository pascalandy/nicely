#!/usr/bin/env bash
set -Eeuo pipefail

fct_stage_package() (
	local tag="${1}" dist="${2}" aur_arch="${3}" go_arch="${4}" stage archive source_file source_url checksum file mode raw templated
	local pkgname="" pkgver="" pkgdir=""
	local -a source_x86_64=() source_aarch64=() sha256sums_x86_64=() sha256sums_aarch64=()
	stage="${dist}/checks/aur-${aur_arch}"
	archive="ncly_${tag#v}_linux_${go_arch}.tar.gz"
	raw="$(cat "${dist}/aur/ncly-bin.pkgbuild")"
	templated="https://github.com/pascalandy/nicely/releases/download/v\${pkgver}/ncly_\${pkgver}_linux_${go_arch}.tar.gz"
	if [[ "${raw}" != *"${templated}"* ]]; then
		printf 'AUR PKGBUILD URL is not tied to pkgver\n' >&2
		return 1
	fi
	mkdir -p "${stage}/source" "${stage}/package"
	cd "${stage}/source"
	# shellcheck source=/dev/null
	source "${dist}/aur/ncly-bin.pkgbuild"
	[[ "${pkgname}" == ncly-bin && "${pkgver}" == "${tag#v}" ]]
	case "${aur_arch}" in
	x86_64)
		source_file="${source_x86_64[0]}"
		checksum="${sha256sums_x86_64[0]}"
		;;
	aarch64)
		source_file="${source_aarch64[0]}"
		checksum="${sha256sums_aarch64[0]}"
		;;
	esac
	source_url="${source_file#*::}"
	[[ "${source_url}" == "https://github.com/pascalandy/nicely/releases/download/${tag}/${archive}" ]]
	printf '%s  %s\n' "${checksum}" "${dist}/${archive}" | shasum -a 256 --check >/dev/null
	[[ $(awk '$1 == "pkgver" && $2 == "=" { print $3 }' "${dist}/aur/ncly-bin.srcinfo") == "${tag#v}" ]]
	awk -v arch="${aur_arch}" '$1 == "arch" && $2 == "=" && $3 == arch { found = 1 } END { exit !found }' "${dist}/aur/ncly-bin.srcinfo"
	[[ $(awk -v key="source_${aur_arch}" '$1 == key && $2 == "=" { print $3 }' "${dist}/aur/ncly-bin.srcinfo") == "${source_file}" ]]
	[[ $(awk -v key="sha256sums_${aur_arch}" '$1 == key && $2 == "=" { print $3 }' "${dist}/aur/ncly-bin.srcinfo") == "${checksum}" ]]
	cp "${dist}/${archive}" "${source_file%%::*}"
	tar -xzf "${source_file%%::*}"
	pkgdir="${stage}/package"
	export CARCH="${aur_arch}"
	package
	cmp ncly "${pkgdir}/usr/bin/ncly"
	cmp LICENSE "${pkgdir}/usr/share/licenses/ncly-bin/LICENSE"
	diff -r THIRD_PARTY_LICENSES "${pkgdir}/usr/share/licenses/ncly-bin/THIRD_PARTY_LICENSES"
	cmp completions/ncly.bash "${pkgdir}/usr/share/bash-completion/completions/ncly"
	cmp completions/_ncly "${pkgdir}/usr/share/zsh/site-functions/_ncly"
	cmp completions/ncly.fish "${pkgdir}/usr/share/fish/vendor_completions.d/ncly.fish"
	while IFS= read -r -d '' file; do
		if [[ "$(uname -s)" == Darwin ]]; then
			mode="$(stat -f '%Lp' "${file}")"
		else
			mode="$(stat -c '%a' "${file}")"
		fi
		if [[ "${file}" == "${pkgdir}/usr/bin/ncly" ]]; then
			[[ "${mode}" == 755 ]]
		else
			[[ "${mode}" == 644 ]]
		fi
	done < <(find "${pkgdir}" -type f -print0)
)

fct_main() {
	if [[ $# -ne 2 || ! "${1}" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
		printf 'Usage: verify-aur.sh vX.Y.Z DIST\n' >&2
		exit 2
	fi
	local dist
	dist="$(cd "${2}" && pwd -P)"
	fct_stage_package "${1}" "${dist}" x86_64 amd64
	fct_stage_package "${1}" "${dist}" aarch64 arm64
	printf 'AUR package staged for x86_64 and aarch64\n'
}

fct_main "$@"
exit 0
