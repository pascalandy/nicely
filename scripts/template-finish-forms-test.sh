#!/usr/bin/env bash
# Sourced by template-finish-test.sh. fct_test_finish calls
# fct_test_finish_intake and fct_test_finish_forms with its repository and
# its fixture variables. They prove that finish reads the intake and the
# reviews only in their typed records, and fails on each value or key that a
# record does not define (COR-104-T4).
# Markdown backticks in the fixture text are literal.
# shellcheck disable=SC2016,SC2154

# fct_test_finish_intake proves the record intake.toml and its view, the part
# ## Intake of vision.md. ## Open lines of spec.md holds the line large-file.
fct_test_finish_intake() {
	local facts
	cp "${finish}/intake.toml" "${scratch}/intake.toml.saved"
	cp "${finish}/vision.md" "${scratch}/vision.md.saved"
	facts="$(grep -c '^\[\[facts\]\]$' "${finish}/intake.toml")"
	# A fact of each kind, each with its fields, passes.
	printf '\n[[facts]]\nkind = "inference"\nstatement = "A file over 1 GB takes more than 30 s"\n\n[[facts]]\nkind = "unknown"\nstatement = "The memory that a file over 1 GB takes"\nid = "large-file"\n' >>"${finish}/intake.toml"
	fct_finish 1
	fct_finish_said 'intake.toml:' absent
	# The view must follow the record.
	fct_finish_said 'vision.md: its part ## Intake is not the view of intake.toml. Fix: run just records, then commit both files'
	fct_records
	fct_records_said 0 'records: wrote'
	fct_finish 1
	fct_finish_said 'is not the view of' absent
	fct_has "${finish}/vision.md" '- A file over 1 GB takes more than 30 s. Inference: no source gives it'
	fct_has "${finish}/vision.md" '- The memory that a file over 1 GB takes. Unknown: see the line `large-file` of `spec.md`'

	# A settled fact needs its source, its revision, and the part read. An
	# inference and an unknown fact cite no source, and an unknown fact opens
	# a line of spec.md (T24, T30, and T31 of N2).
	printf '\n[[facts]]\nkind = "sourced"\nstatement = "The worker reads stdin"\nsource = "https://example.com/inspect"\n\n[[facts]]\nkind = "inference"\nstatement = "Large files take long"\nread = "the loop of main"\n\n[[facts]]\nkind = "guess"\nstatement = "It runs fast"\n\n[[facts]]\nkind = "unknown"\nstatement = "The time of a file over 2 GB"\n\n[[facts]]\nkind = "unknown"\nstatement = "The time of a file over 4 GB"\nid = "huge-file"\n\n[[facts]]\nkind = "sourced"\nsource = "worker/inspect.py"\nrevision = "v1"\nread = "main"\n\n[[experiments]]\ncommand = "python3 worker/inspect.py big.md"\n' >>"${finish}/intake.toml"
	fct_finish 1
	fct_finish_said "intake.toml: the sourced fact $((facts + 3)) has no source, no revision, or no part read. Fix: give its source, its revision, and the part that you read in read. A link alone is not evidence of a read"
	fct_finish_said "intake.toml: the inference fact $((facts + 4)) names a source, a revision, or a part read. Fix: make it a sourced fact, or remove them"
	fct_finish_said "intake.toml: the fact $((facts + 5)) has the kind \"guess\". Fix: give it the kind sourced, inference, or unknown"
	fct_finish_said "intake.toml: the unknown fact $((facts + 6)) has no id. Fix: give it the spec ID of its open line in spec.md"
	fct_finish_said "intake.toml: the unknown fact $((facts + 7)) names the line \`huge-file\`, which spec.md does not hold. Fix: write it under ## Open lines, with that ID"
	fct_finish_said "intake.toml: the fact $((facts + 8)) has no statement"
	fct_finish_said 'intake.toml: the experiment 2 has no command or no result'

	# A key that the record does not define fails, and just records writes no view.
	cp "${scratch}/intake.toml.saved" "${finish}/intake.toml"
	printf '\n[[facts]]\nkind = "inference"\nstatement = "Large files take long"\nurl = "https://example.com/inspect"\n' >>"${finish}/intake.toml"
	cp "${finish}/vision.md" "${scratch}/vision.md.before"
	fct_records
	fct_records_said 1 'intake.toml: the key facts.url is no field of the record'
	cmp -s "${finish}/vision.md" "${scratch}/vision.md.before" || fct_fail 'just records wrote vision.md from a record that it refused'
	fct_finish 1
	fct_finish_said 'intake.toml: the key facts.url is no field of the record'
	printf 'kind = \n' >>"${finish}/intake.toml"
	fct_finish 1
	fct_finish_said 'intake.toml: it is no valid record'
	cp "${scratch}/intake.toml.saved" "${finish}/intake.toml"
	cp "${scratch}/vision.md.saved" "${finish}/vision.md"
}

# fct_test_finish_forms proves that a finding has one ID, severity, state,
# spec ID, and family, in their keys only. It runs at a commit that passes
# finish, whose review declines the advise R1-F2.
fct_test_finish_forms() {
	local kept
	kept="$(fct_finding R1-F2 advise declined lines "${declined}")"
	# At the parent of 59d33f38, a finding with no state beside None. under
	# ## Open lines passed finish (#423). A record has no finding without a state.
	fct_readiness "${sha}" Ready "$(fct_finding R1-F2 advise '' lines | grep -v '^state = ')"
	fct_finish 1
	fct_finish_said 'review.toml: the finding R1-F2 has no state. Fix: give it the state open, fixed, limit, or declined'

	# The free-form grades of #422 passed finish beside the severity advise.
	# A record has no key for them, so finish fails, and just records writes no view.
	cp "${finish}/readiness.md" "${scratch}/readiness.md.before"
	fct_readiness "${sha}" Ready "${kept}"$'\ngrade = "block"\n\n[notes]\nR1-F1 = "really a warn"'
	fct_records_said 1 'review.toml: the key findings.grade is no field of the record'
	cmp -s "${finish}/readiness.md" "${scratch}/readiness.md.before" || fct_fail 'just records wrote readiness.md from a record that it refused'
	fct_finish 1
	fct_finish_said 'review.toml: the key findings.grade is no field of the record, so no check reads it. Fix: grade a finding only in its key severity'
	fct_finish_said 'review.toml: the key notes is no field of the record'
	fct_finish_said 'review.toml: the key notes.R1-F1' absent
	fct_readiness "${sha}" Ready "$(fct_finding R1-F2 'really a warn' declined lines "${declined}")"
	fct_finish 1
	fct_finish_said 'review.toml: the finding R1-F2 has the severity "really a warn", which is not block, warn, or advise'
	fct_readiness "${sha}" Ready "$(fct_finding R1-F2 '' declined lines "${declined}" 'evidence = "Grade: block. The GPS leak is a block for me. This issue is warn-level."' | grep -v '^severity = ')"
	fct_finish 1
	fct_finish_said 'review.toml: the finding R1-F2 has no severity'

	# Each finding has its own ID, of a round of the ledger, and each key of its state.
	fct_readiness "${sha}" Ready "${kept}"$'\n\n'"${kept}"$'\n\n'"$(fct_finding F1 advise closed lines)"$'\n\n'"$(fct_finding R9-F1 advise declined lines | grep -v '^family = \|^title = ')"
	fct_finish 1
	fct_finish_said 'review.toml: two findings have the ID R1-F2'
	fct_finish_said 'review.toml: the finding "F1" has no ID of the form R<round>-F<number>'
	fct_finish_said 'review.toml: the finding F1 has the state "closed", which is not open, fixed, limit, or declined'
	fct_finish_said 'review.toml: the finding R9-F1 is of round 9, after the last round 2 of the ledger'
	fct_finish_said 'review.toml: the finding R9-F1 has no title'
	fct_finish_said 'review.toml: the finding R9-F1 has no family'
	fct_finish_said 'review.toml: the declined finding R9-F1 gives no reason'

	# A fixed finding needs a passing and a refused test of its family (T43 of N2).
	fct_readiness "${sha}" Ready "$(fct_finding R1-F2 advise fixed size "commit = \"${sha}\"" | sed 's/^family = .*/family = "The input holds text"/')"
	fct_finish 1
	fct_finish_said 'review.toml: the fixed finding R1-F2 names the line `size`, which is not in its family "The input holds text"'
	fct_readiness "${sha}" Ready "$(fct_finding R1-F2 advise fixed lines "commit = \"${sha}\"" | sed 's/^family = .*/family = "Long lines"/')"
	fct_finish 1
	fct_finish_said 'review.toml: the fixed finding R1-F2 names the family "Long lines", which is no row of ## Case matrix'
	fct_edit "${finish}/spec.md" '| `python3` runs the worker | Other: Missing program, and the intake | `size` | `no-python`, `worker-fails` |' '| `python3` runs the worker | Other: Missing program, and the intake | `size` | None |'
	fct_readiness "${sha}" Ready "$(fct_finding R1-F2 advise fixed size "commit = \"${sha}\"" | sed 's/^family = .*/family = "`python3` runs the worker"/')"
	fct_finish 1
	fct_finish_said 'review.toml: the fixed finding R1-F2 has no refused test in its family "`python3` runs the worker"'
	git -C "${finish}" checkout -q spec.md
	fct_readiness "${sha}" Ready "$(fct_finding R1-F2 advise fixed lines "commit = \"${sha}\"")"
	fct_finish 0
	git -C "${finish}" checkout -q review.toml readiness.md
}
