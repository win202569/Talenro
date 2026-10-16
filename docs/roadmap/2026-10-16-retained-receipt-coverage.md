# Populated retained-lifetime receipt coverage — 2026-10-16

## Scope and result

User approved closing the deferred nonblocking nested-receipt coverage gap after
642b599d615cbeb3c6007aa59b0c6693b3a8684f. This changes only the receipt test and
progress/recovery documentation. Collector, registry, adapters, converter and
completion guard remain unchanged. No diagnostic Run, Go, live-child probe,
historical PID query, SDK audit, matrix replay or main merge occurred.

Actual scoped checks:receipt16, registry27 and integration12 PASS on each
PowerShell7/Windows PowerShell5.1 host (55 per host,110 total; all six commands
exit0). Four integration checks remain supplementary static checks, not native
behavioral verification. The older11-suite/298-per-host run is historical and
was NOT rerun or relabelled as305 in this increment.

## What the new test exercises

A hand-populated composite diagnostic DTO contains32 entries and cap counters,
source-query headers/clocks, open/membership evidence, admission/root-wait-end/
final observations, explicit ambiguity, post-termination metadata, missing image
with error5, and failed admission87 with genuinely null unattempted fields. It is
serialization test data, not a generated valid native lifecycle or process trace.

The actual converter feeds JSON at the existing writer's depth6 with warnings
treated as errors, then deserializes into an independent object. Literal checks
pin representative nested values, counters, nulls and missing-evidence defaults.
The real completion guard still refuses this diagnostic receipt. No raw handle,
adapter or registry property is allowed in the serialized diagnostic object.

Four clones of that actual JSON are corrupted only in memory:remove source-query
evidence, clear ambiguity, change cap attempts32 to31, or inject a Handle property.
Each must trip its specific test assertion. These prove sensitivity of the new
test assertions; they are NOT new production input-validation rules or evidence
that the production guard separately diagnoses those four corruptions.

Existing correct production behavior passes the expanded test; no product RED/fix
cycle is claimed. Two initial fixture compatibility failures were diagnosed and
corrected without product changes:PS7 materializes UTC JSON values as DateTime,
so comparison normalizes the same instant; PowerShell's typed string-field null
assignment produced an empty string, so reflection supplies a genuine null fixture.
The null assertion was preserved, not weakened to accept empty strings.

## Evidence and recovery

Local ignored logs/context/summary/source binding are under
.superpowers/diagnostics/2026-10-16-retained-receipt/validation-001. Seven other
source/test hashes match the prior implementation binding; only the receipt test
hash changes. All six logs and the hash binding for the updated test/source are
preserved in that local directory. Its9-file evidence-manifest.json hashes to
`1B85674A4A2C3B2E1F73D241200131DD48EDB7979DF0AD0CEC802A408DA0F75C`;
receipt test SHA256 is
`51624535464B12E124BB352929D09DA4E0438269CB021FC5C872F4EA38312CD2`.
No raw logs, binaries, caches or SDK files are committed.

The original independent reviewer read the new test delta and confirmed M1 is
covered with no findings. Review was source/diff only, no test/native execution.
It specifically confirmed the non-tautological corruption controls, UTC instant
normalization and strict null preservation. This review covers serialization
testing only, not native lifecycle chronology or acceptance/main readiness.

Read current-status.md and docs/runbooks/repository-recovery.md after cloning the
existing development branch on a fresh machine. These three portable tests need
the repository and suitable Windows PowerShell/.NET hosts, not private Go caches
or historical receipts. Logs are local evidence, not automatically restored by
Git. M1 is covered by the new populated-roundtrip cases; this does not establish
native ownership/release/timing, error87's cause/fix or wider acceptance. Historical
matrix003 remains2PASS/1FAIL/23UNRUN and production/main gates stay open.
