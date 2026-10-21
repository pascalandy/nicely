#!/usr/bin/env bash
# runner-rehearsal-job.sh runs in ubuntu:24.04 for runner-rehearsal.sh, with the SHA and the tag.
# As root, it adds what the runner image carries and the minimal image lacks. As the user runner,
# it clones the SHA, stands in for each action of release.yml, and runs each step's script as
# GitHub runs it, through "Rehearse without publishing". The versions come from that workflow.
set -Eeuo pipefail
sha="${1:?usage: runner-rehearsal-job.sh <sha> <tag>}"
tag="${2:?usage: runner-rehearsal-job.sh <sha> <tag>}"
fail() {
	printf 'runner-rehearsal: %s\n' "$*" >&2
	exit 1
}
if ((EUID == 0)); then
	export DEBIAN_FRONTEND=noninteractive
	apt-get update -qq
	# The runner image carries these tools, such as jq, ps, ruby, and sudo.
	apt-get install -y -qq --no-install-recommends ca-certificates curl wget git xz-utils zstd ruby gcc g++ make libc6-dev \
		python3 python3-venv python3-pip jq procps psmisc lsof unzip zip file rsync openssh-client gnupg bsdextrautils less sudo tzdata >/dev/null
	# The runner image answers /usr/bin/go with an older Go: 1.24.13 in October 2026.
	curl -fsSL https://go.dev/dl/go1.24.13.linux-amd64.tar.gz | tar -xz -C /usr/local
	ln -s /usr/local/go/bin/go /usr/local/go/bin/gofmt /usr/bin/
	useradd -m runner
	echo 'runner ALL=(ALL) NOPASSWD:ALL' >/etc/sudoers.d/runner
	exec sudo -u runner -H bash "${BASH_SOURCE[0]}" "${sha}" "${tag}"
fi

export CI=true GITHUB_ACTIONS=true RUNNER_TEMP=/home/runner/work/_temp GITHUB_WORKSPACE=/home/runner/work/nicely/nicely
export GITHUB_PATH="${RUNNER_TEMP}/github_path" PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
steps="${RUNNER_TEMP}/steps" tools=/home/runner/tools clone=/home/runner/clone
mkdir -p "${steps}" "${tools}" "${GITHUB_WORKSPACE}"
: >"${GITHUB_PATH}"
# The checkout of the workflow, staged apart: the commit, detached, with every branch and tag.
git clone -q https://github.com/pascalandy/nicely.git "${clone}"
git -C "${clone}" checkout -q --detach "${sha}"
git -C "${clone}" branch -q -D main
# The tag does not exist yet before a release, so the clone gets it. A tag on GitHub stays as it is.
git -C "${clone}" rev-parse -q --verify "refs/tags/${tag}" >/dev/null || git -C "${clone}" tag "${tag}" "${sha}"
# Each step runs in the workspace, which stays empty until the checkout step, as on GitHub.
cd "${GITHUB_WORKSPACE}"

# Each step of the release job through the rehearsal becomes files: NN.name, then NN.uses and
# NN.with.<key>, or NN.run. A key without a stand-in, such as if or defaults, stops the rehearsal
# before any step runs, because GitHub would run the steps otherwise.
ruby - "${tag}" "${steps}" "${clone}/.github/workflows/release.yml" <<'RUBY'
require "yaml"
tag, dir, path = ARGV
only = lambda do |where, keys, known|
  extra = keys.map(&:to_s) - known
  abort "runner-rehearsal: #{where} sets #{extra.join(", ")}, which has no stand-in" unless extra.empty?
end
workflow = YAML.load_file(path)
# YAML 1.1 reads the key on as true.
only.("release.yml", workflow.keys, %w[name on true permissions concurrency jobs])
job = workflow.fetch("jobs").fetch("release")
only.("the release job", job.keys, %w[runs-on env steps])
# The inputs that each stand-in reads. Any other input could change what GitHub runs.
inputs = {
  "actions/checkout" => %w[ref fetch-depth persist-credentials],
  "actions/setup-go" => %w[go-version-file cache-dependency-path],
  "extractions/setup-just" => %w[just-version],
  "astral-sh/setup-uv" => [],
}
abort "runner-rehearsal: the job runs on #{job["runs-on"]}, not ubuntu-24.04" unless job["runs-on"] == "ubuntu-24.04"
expand = lambda do |value|
  text = value.to_s.gsub("${{ inputs.tag }}", tag)
  abort "runner-rehearsal: no stand-in for the expression in #{text}" if text.include?("${{")
  text
end
File.write("#{dir}/env", job.fetch("env", {}).map { |key, value| "#{key}=#{expand.(value)}\n" }.join)
job.fetch("steps").each_with_index do |step, index|
  file = format("%s/%02d", dir, index)
  File.write("#{file}.name", step.fetch("name"))
  if step["uses"]
    only.(step["name"], step.keys, %w[name uses with])
    action = step["uses"].split("@").first
    abort "runner-rehearsal: no stand-in for the action #{action}" unless inputs.key?(action)
    with = step["with"] || {}
    only.("#{step["name"]} (#{action})", with.keys, inputs[action])
    # The clone gives the tag's commit with every branch and tag, as these inputs ask.
    if action == "actions/checkout" && (expand.(with["ref"]) != "refs/tags/#{tag}" || with["fetch-depth"].to_s != "0")
      abort "runner-rehearsal: #{step["name"]} must check out refs/tags/#{tag} with fetch-depth 0"
    end
    File.write("#{file}.uses", action)
    with.each { |key, value| File.write("#{file}.with.#{key}", expand.(value)) }
  else
    only.(step["name"], step.keys, %w[name run shell])
    abort "runner-rehearsal: #{step["name"]} runs in #{step["shell"]}, not bash" unless step["shell"] == "bash"
    File.write("#{file}.run", expand.(step.fetch("run")))
  end
  break if step["name"] == "Rehearse without publishing"
end
RUBY
while IFS= read -r variable; do export "${variable?}"; done <"${steps}/env"

# fct_action stands in for the action of a step, in the order of the workflow.
fct_action() {
	local step="${1}" version file
	case "$(cat "${step}.uses")" in
	actions/checkout)
		# The action clears what earlier steps left in the workspace.
		find "${GITHUB_WORKSPACE}" -mindepth 1 -delete
		rmdir "${GITHUB_WORKSPACE}"
		mv "${clone}" "${GITHUB_WORKSPACE}"
		cd "${GITHUB_WORKSPACE}"
		;;
	actions/setup-go)
		file="$(cat "${step}.with.go-version-file")"
		version="$(sed -n 's/^toolchain go//p' "${file}")"
		[[ -n "${version}" ]] || version="$(sed -n 's/^go //p' "${file}")"
		mkdir -p "${tools}/go${version}"
		curl -fsSL "https://go.dev/dl/go${version}.linux-amd64.tar.gz" | tar -xz -C "${tools}/go${version}" --strip-components=1
		PATH="${tools}/go${version}/bin:${HOME}/go/bin:${PATH}"
		export GOTOOLCHAIN=local
		;;
	extractions/setup-just)
		version="$(cat "${step}.with.just-version")"
		mkdir -p "${tools}/just"
		curl -fsSL "https://github.com/casey/just/releases/download/${version}/just-${version}-x86_64-unknown-linux-musl.tar.gz" | tar -xz -C "${tools}/just" just
		PATH="${tools}/just:${PATH}"
		;;
	astral-sh/setup-uv)
		mkdir -p "${tools}/uv"
		curl -fsSL https://github.com/astral-sh/uv/releases/latest/download/uv-x86_64-unknown-linux-gnu.tar.gz | tar -xz -C "${tools}/uv" --strip-components=1
		PATH="${tools}/uv:${PATH}"
		;;
	*) fail "no stand-in for the action $(cat "${step}.uses")" ;;
	esac
}

for step in "${steps}"/*.name; do
	step="${step%.name}"
	name="$(cat "${step}.name")"
	printf '== %s\n' "${name}"
	if [[ -e "${step}.uses" ]]; then
		fct_action "${step}"
	else
		bash --noprofile --norc -eo pipefail "${step}.run" </dev/null
		# Each folder that a step adds to GITHUB_PATH comes first in the next steps.
		while IFS= read -r folder; do PATH="${folder}:${PATH}"; done <"${GITHUB_PATH}"
		: >"${GITHUB_PATH}"
	fi
	if [[ "${name}" == "Rehearse without publishing" ]]; then
		printf 'runner-rehearsal: the steps of release.yml through "%s" passed at %s for %s\n' "${name}" "${sha}" "${tag}"
		exit 0
	fi
done
fail 'release.yml has no step named "Rehearse without publishing"'
