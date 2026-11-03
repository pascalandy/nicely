------------------------------- MODULE Finish -------------------------------
\* The finish gate of an extension repository, as finish-gate.md defines it.
\* README.md maps each action to the code of sdk/extensionrepo that it models.
EXTENDS Naturals

CONSTANTS Findings, MaxRounds, PreFix
VARIABLES round, states, severity, born, proof, accepted, reasons, verdict, gate
vars == <<round, states, severity, born, proof, accepted, reasons, verdict, gate>>

\* "unset" is a finding of review.toml with no key state.
States == {"absent", "unset", "open", "fixed", "limit", "declined"}
Severities == {"advise", "warn", "block"}
Rank == [s \in Severities |-> CASE s = "advise" -> 1 [] s = "warn" -> 2 [] s = "block" -> 3]
Verdicts == {"Ready", "Ready with stated limits", "Needs work"}
Present == {f \in Findings : states[f] # "absent"}

Init == /\ round = 1
        /\ states = [f \in Findings |-> "absent"]
        /\ severity = [f \in Findings |-> "advise"]
        /\ born = [f \in Findings |-> 0]
        /\ proof = {} /\ accepted = {} /\ reasons = {}
        /\ verdict = "Needs work" /\ gate = "unchecked"

\* Each change of the records needs a new run of finish. A pass ends the build.
Change == gate # "pass" /\ gate' = "unchecked"

\* The review of this round reports a new finding, with a state or none.
Report(f, s, g) == /\ Change /\ states[f] = "absent"
                   /\ states' = [states EXCEPT ![f] = s]
                   /\ severity' = [severity EXCEPT ![f] = g]
                   /\ born' = [born EXCEPT ![f] = round]
                   /\ UNCHANGED <<round, proof, accepted, reasons, verdict>>

\* A later review gives a finding another state. Its ID and round stay.
Restate(f, s) == /\ Change /\ f \in Present /\ states[f] # s
                 /\ states' = [states EXCEPT ![f] = s]
                 /\ UNCHANGED <<round, severity, born, proof, accepted, reasons, verdict>>

\* A severity never goes down.
Regrade(f, g) == /\ Change /\ f \in Present /\ Rank[g] > Rank[severity[f]]
                 /\ severity' = [severity EXCEPT ![f] = g]
                 /\ UNCHANGED <<round, states, born, proof, accepted, reasons, verdict>>

\* The evidence of a state: a fixing commit with a pair of tests of its
\* family, the acceptance of the user in spec.md, or the reason of the author.
Prove(f) == /\ Change /\ f \notin proof /\ proof' = proof \cup {f}
            /\ UNCHANGED <<round, states, severity, born, accepted, reasons, verdict>>
Accept(f) == /\ Change /\ f \notin accepted /\ accepted' = accepted \cup {f}
             /\ UNCHANGED <<round, states, severity, born, proof, reasons, verdict>>
Reason(f) == /\ Change /\ f \notin reasons /\ reasons' = reasons \cup {f}
             /\ UNCHANGED <<round, states, severity, born, proof, accepted, verdict>>

Review(v) == /\ Change /\ verdict # v /\ verdict' = v
             /\ UNCHANGED <<round, states, severity, born, proof, accepted, reasons>>

Closed(f) == \/ states[f] = "fixed" /\ f \in proof
             \/ states[f] = "limit" /\ f \in accepted
             \/ states[f] = "declined" /\ severity[f] = "advise" /\ f \in reasons

\* PreFix is the state check at the parent of 59d33f38: a finding of this
\* round with no state added no problem.
StateOK(f) == Closed(f) \/ (PreFix /\ states[f] = "unset" /\ born[f] = round)

VerdictOK ==
    \/ /\ verdict = "Ready" /\ accepted = {}
       /\ \A f \in Present : severity[f] \in {"warn", "block"} => states[f] = "fixed"
    \/ /\ verdict = "Ready with stated limits"
       /\ \A f \in Present : severity[f] = "block" => states[f] = "fixed"
       /\ \A f \in Present : severity[f] = "warn" =>
             states[f] = "fixed" \/ (states[f] = "limit" /\ f \in accepted)

Finish == /\ gate = "unchecked"
          /\ gate' = IF VerdictOK /\ \A f \in Present : StateOK(f) THEN "pass" ELSE "fail"
          /\ UNCHANGED <<round, states, severity, born, proof, accepted, reasons, verdict>>

\* A failed finish starts the next round, up to the bound of the model.
NextRound == /\ gate = "fail" /\ round < MaxRounds
             /\ round' = round + 1 /\ gate' = "unchecked"
             /\ UNCHANGED <<states, severity, born, proof, accepted, reasons, verdict>>

Next == \/ \E f \in Findings, s \in States \ {"absent"}, g \in Severities : Report(f, s, g)
        \/ \E f \in Findings, s \in States \ {"absent"} : Restate(f, s)
        \/ \E f \in Findings, g \in Severities : Regrade(f, g)
        \/ \E f \in Findings : Prove(f) \/ Accept(f) \/ Reason(f)
        \/ \E v \in Verdicts : Review(v)
        \/ Finish \/ NextRound

Spec == Init /\ [][Next]_vars

TypeOK == /\ round \in 1..MaxRounds /\ states \in [Findings -> States]
          /\ severity \in [Findings -> Severities] /\ born \in [Findings -> 0..MaxRounds]
          /\ proof \subseteq Findings /\ accepted \subseteq Findings /\ reasons \subseteq Findings
          /\ verdict \in Verdicts /\ gate \in {"unchecked", "pass", "fail"}

\* A pass leaves no finding open: each one holds a state and its evidence.
NoOpenPass == gate = "pass" => \A f \in Present : Closed(f)
\* A pass leaves no block unfixed, even one that the user accepted as a limit.
NoUnfixedBlockPass == gate = "pass" => \A f \in Present : severity[f] = "block" => states[f] = "fixed"
\* Needs work never passes.
NoWorkPass == gate = "pass" => verdict # "Needs work"
=============================================================================
