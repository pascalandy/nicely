#!/usr/bin/env bash
# Sourced by template-test.sh. fct_test_acceptance_cases proves that the
# acceptance case of a rendered repository runs through the kit, as
# COR-104-T5 asks. At the parent of 86d24e83, acceptance.sh matched
# "lines":3 as a prefix of the answer, so a count of 30 passed. The case reads
# the answer with the reader answer of the kit, and refuses that count.

# fct_test_acceptance_cases makes the worker of the rendered repository $1
# count ten times the lines, then requires the refusal of reads in the
# scenario acceptance. The scratch folder is $2.
fct_test_acceptance_cases() {
	local repo="${1}" scratch="${2}" status=0
	local count='    print(json.dumps({"size_bytes": size, "lines": lines, "text": text}))'
	fct_edit "${repo}/worker/inspect.py" "${count}" '    print(json.dumps({"size_bytes": size, "lines": lines * 10, "text": text}))'
	(cd "${repo}" && just scenario acceptance) >"${scratch}/acceptance-cases.out" 2>&1 || status=$?
	fct_edit "${repo}/worker/inspect.py" '    print(json.dumps({"size_bytes": size, "lines": lines * 10, "text": text}))' "${count}"
	[[ "${status}" != 0 ]] || fct_fail 'just scenario acceptance accepts a count of 30 lines for an input of 3 lines'
	grep -qF 'reads: answer gives "size_bytes: 29\nlines: 30\n" for answer.json, and the reference reference holds "size_bytes: 29\nlines: 3\n"' "${scratch}/acceptance-cases.out" ||
		fct_fail "just scenario acceptance fails without the refusal of reads: $(cat "${scratch}/acceptance-cases.out")"
	git -C "${repo}" diff --quiet -- worker/inspect.py || fct_fail 'the test of the acceptance case leaves worker/inspect.py changed'
}
