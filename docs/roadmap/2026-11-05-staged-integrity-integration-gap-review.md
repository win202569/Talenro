# Staged integrity integration readiness review

## Latest shadow integration follow-up: Task1 incomplete

Plan9830af55 approved; parser RED552c9d expected missing module, GREENd95195
stopped at junction fixture creation after seven groups matched. Reparse rejection
not evaluated; underlying creation exception was discarded, cause unknown.
No retries or post-stop fixes.2/7 calls consumed; Task2 records, Task3 formal hooks
and Task4 combined checks not started. Formal runner and production manifest are
unchanged/absent respectively. These partial helpers do not establish integration
or safety acceptance. Portable stop archive53files/73directories/81186bytes verified
14596b by decoding hashes and comparing original files. See recovery runbook.
Next proposed scope: retain original junction error in new-root diagnosis, then
seek bounded continuation if needed. No security changes, business/Docker, full
suite, independent review or main merge;44/34/6/217 unchanged;NOT_ACCEPTED.

## Latest parent persistence follow-up: ten cases matched

Approved plan19946250 executed with exact pinned driver helpers/tail/catch and
real tiny-file hashing/CreateNew writes, not mocked persistence. No driver entry
or supervisor launched. Bootstrap2b289d→d461d6 passed; checkpoint6229b867.

|Cases|Result|
|---|---|
|P01|PASS: normal audit VERIFIED, result remains NOT_ACCEPTED|
|P02-P03|PASS: real hash drift/absent-pin violation yields FAILED audit and stops|
|P04-P07|PASS: preoccupied map/audit/result/supervision bytes preserved; no false success|
|P08-P09|PASS: pending/unknown exit blocks audit before deliberate input drift|
|P10|PASS: audit success does not upgrade failed transport|

Sole matrix1ef38d exit0,10 matches;2 harmless bootstrap plus10 terminal parent
runspaces. Capture of original Console.Out success marker is synchronous and
separate from PowerShell output; source decisions unchanged. Archive147files
184491bytes verifiedcc18fc, containing setup/context/slice/input/output/verdicts.
This covers real exclusive-open collisions, not general disk resilience. Mid-write,
flush, disk-full, ACL, hung filesystem and failure-marker save failures remain
untested. Formal lifecycle, clean-machine dependencies and full-suite/independent
acceptance remain open. Counters44/34/6/217 unchanged;NOT_ACCEPTED.
Recommend separate formal integration design next; no direct fixed-root driver
wiring, runtime budget, security changes or main merge follows from these tests.

## Next gate planned: parent persistence, not formal wiring

Read-only source review confirms diagnostic driver Save-IW uses exclusive
CreateNew writes; supervision is persisted before terminal gating, audit before
result, and outer failure recording can itself fail. Formal Invoke-C12Group has
separate package/profile gates, resource ownership and deadlines. A fixed-root
diagnostic run is not a substitute for that lifecycle.
Plan: docs/superpowers/plans/2026-11-05-parent-audit-persistence-validation.md.
Ten proposed cases use exact existing parent-tail code and tiny real files; cover
hash drift, absent-pin violation, exclusive-open collisions, pending/unknown exits
and transport failure. No live driver entry, runtime binaries or full scan.
Write/flush/disk-full/ACL/hang and failure-marker persistence remain separate gaps.
Formal integration design is deferred until this finite gate is reviewed; passing
it cannot authorize real runtime work or capability promotion. Plan not executed;
44/34/6/217 unchanged,NOT_ACCEPTED. This is an author review, not independent audit.

## Latest v4 follow-up: complete offline matrix matched

Strict Boolean repair approved after986d5a48: Test-SIBool requires actual Boolean
type and expected value;29 expression mappings replace truthiness/false coercion.
Inverse-delta check preserves all other v3 logic, including duplicate-key parsing,
numeric identity/map/output checks. Actual role predicate REDaae9d4/GREEN00decd
passed65 assertions across typed values and string/number/null/array/object errors.
Preflight054211 passed; implementationbf3aecb8.

|Cases|v4 result|
|---|---|
|C01|PASS: unchanged valid baseline accepted|
|C02-C06|PASS: missing/failed/malformed/wrong-identity audit refused|
|C07-C08|PASS: duplicate keys and string Boolean refused|
|C09-C11|PASS: missing/drifted map and actual exit1 refused|
|C12-C13|PASS: pending/unknown terminal state refused before missing outputs|
|C14-C15|PASS: incomplete trace and wrong stderr length refused|
|C16|PASS: existing verdict sentinel preserved; no passing verdict|

Sole matrix51d8a8 ended6f360b exit0;16 terminal evaluations, none NOT_RUN;
MATRIX_VERIFIED. Read-only405868 confirms persisted outcomes. All45 original
files unchanged per case. v4 archive425files56994469bytes verified7f3916.
This closes the saved-evidence matrix gap, not real parent audit-write/disk-hang
fault coverage, formal-entry lifecycle integration, clean-machine dependencies,
independent review or full-suite acceptance. No business/OS children/Docker;
44/34/6/217 unchanged. NOT_ACCEPTED/all five flagsfalse; no main merge.
Next: read-only scoped plan for remaining parent persistence and integration gaps,
then separate execution approval. No new runtime budget implied by this result.

## Latest v3 follow-up: duplicate keys refused, Boolean coercion reproduced

Approved duplicate-key repair is isolated in review-v3, with raw JSON member
inspection before dictionary conversion for file JSON and control/event/journal
lines. Nested/escaped/case-alias duplicate tests pass; valid siblings and literal
key text remain accepted (9 cases, REDfe578f/GREENa83450). No downstream predicates
changed; preflight02d611 checks permitted deltas and45 unchanged originals.

|Cases|v3 result|
|---|---|
|C01-C06|PASS: valid baseline and earlier refusal checks|
|C07|PASS: duplicate State rejected with duplicate/ambiguous JSON reason|
|C08|FAIL: Writer.Associated JSON string "false" accepted|
|C09-C16|NOT_RUN: stopped at first unexpected outcome|

Sole matrixee8f8d endedb3ecba exit1;8 terminal in-process evaluations.
Read-only4a06cd confirms C08 raw string and wrong passing verdict. Existing
role predicate uses $p.Associated as truthiness without requiring Boolean type.
This previously suspected gap is now reproduced. Matching result.Result was
mutated too, so a copy mismatch did not mask the actual type-validation defect.
Do not weaken C08 expectation or count its verdict as acceptance. No repair or
retry this turn. Proposed next scope: strict Boolean evidence validation in a
new verifier version, targeted type regressions and fresh finite matrix.
Portable v3 package225files28215523bytes verified273761. No business/OS children,
Docker, full suite, independent review or main merge.44/34/6/217 unchanged;
NOT_READY_FOR_FORMAL_INTEGRATION/NOT_ACCEPTED.

## Latest follow-up: duplicate audit key acceptance reproduced

After user approval, versioned payload compatibility passed13 checks (REDa50432,
GREEN6127a8), including both real archive formats. Existing byte length/hash
checks remain. Matrix decisions and original verifier source unchanged (003768).
The sole v2 matrix8af4a6 ended6e1914 exit1 after7 terminal evaluations:

|Cases|Result|
|---|---|
|C01|PASS: baseline accepted|
|C02-C06|PASS: missing/failed/malformed/wrong-identity audits rejected|
|C07|FAIL: duplicate State FAILED then VERIFIED accepted|
|C08-C16|NOT_RUN: stopped at first unexpected outcome|

Read-only d9f9a3 confirms the raw duplicate keys and persisted passing verdict.
ReadJson currently uses ConvertFrom-Json -AsHashtable and checks only resulting
audit.State; the C07 artifact demonstrates that this loses the conflicting key.
This is now an observed validation defect, not merely a source-review concern.
C08 string Boolean coercion remains untested; no conclusion about that case.
Do not use C07's passing verdict as acceptance or weaken the expected rejection.
No verifier repair/retry this turn. Propose separately approved strict duplicate
key rejection in a new verifier version, targeted regression and bounded matrix.
188-file portable-runtime-staged-evidence-matrix-v2.json preserves all v2 fixtures,
snapshots, mutated/baseline inputs, verdicts and stop result (ad6d6b verified).
45 originals unchanged,44/34/6/217 unchanged; zero business/OS children/Docker.
NOT_READY_FOR_FORMAL_INTEGRATION/NOT_ACCEPTED; no full suite or independent review.

## Execution follow-up: fixture preparation stopped

Plan072f86ac was approved. Bootstrap RED735a03/GREEN2b2b86 passed B01–B03
(two harmless in-process runspaces). Matrix1b894b stopped restoring C01 before
any verifier invocation; C01–C16 all NOT_RUN. Read-only262952 located archive
schema mismatch: historical stages uses Base64, new helper expects RawBase64.
No fix/retry;18restored files and45unchanged original files verified/preserved.
This exposes a test-harness compatibility gap; it neither confirms nor refutes
the JSON-coercion/duplicate-key concerns below. All integration blockers remain.
Stopped archive29files3515023bytes SHA256
DD1CC3CA249928653F8C8068F76C9404C900541DCA0FF2E0A068620930EC36F9.
Next proposed approval: versioned dual-schema decoder with missing/ambiguous
field tests, then one matrix invocation in fresh roots, no business processes.

Baseline: a79c83767e44037b8894767b69f0a0f66a5cccee. Read-only source/document
review; no workload, fixture execution, full suite, Docker or compilation.
Counters44/34/6/217 unchanged. Conclusion: NOT_READY_FOR_FORMAL_INTEGRATION.
This is a scoped author review, not an independent security audit of C12.

## Evidence and limits

|Area|Evidence established|Remaining gap|
|---|---|---|
|Normal staged run|Observation44 owned exits, persisted parent audit, complete trace and exact48/0bytes; v2 saved-result review passed|One fixed two-group prefix, not full host or C12 integration suite|
|Trace/gate memory logic|11cases102assertions; strict trace parser and ten-Boolean conjunction|G03/G04 set verified facts false; they do not feed corrupted records through the real verifier|
|Static checks|Corrected AST predicate and allowed source checks passed|Source assertions are not disk-failure or malformed-evidence execution|
|Empty output|Four focused assertions and one complete saved-result review passed|No end-to-end negative output/receipt matrix for the staged version|
|Old supervisor failures|Earlier nine synthetic process scenarios recorded|They predate removal of writer-local after-check; do not prove new parent audit/publication failure handling|
|Recovery|Portable archives contain source, receipts and verified48/0bytes|Absolute worktree/account/runtime paths and excluded DLLs prevent an arbitrary fresh machine from simply running the harness|

Primary inspected files (repository-relative):
- docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-staged-integrity.tests.ps1: G02–G04 operate on facts, not serialized artifacts.
- runtime-staged-integrity-review-v2.ps1 in the same directory: ReadJson, Read-SIMap, role/control checks and final fact assignments.
- runtime-staged-integrity-driver.ps1: one post-exit Map-IW, audit Save-IW and state gate.
- scripts/run-c12-integration.ps1: Focused/Suite entry, package/profile allowlists, prepared authority receipts, Docker and Go integration execution.
- docs/superpowers/specs/2026-11-05-runtime-staged-integrity-design.md: explicitly changed temporal guarantee and no hard parent-audit deadline.

## Prioritized blockers

1. **Real evidence rejection coverage.** Missing/FAILED audit, wrong input/intent,
   missing/drifted map and late role failure need complete-verifier fixture tests.
   Passing Boolean gate tests is insufficient evidence about upstream validation.
2. **JSON type/schema consistency.** Trace parser explicitly rejects duplicate keys
   and wrong numeric types; general ReadJson uses ConvertFrom-Json and many checks
   use Boolean truthiness or non-strict equality. Example: audit state and role
   checks are not a uniform strict schema. Duplicate keys/string booleans must be
   tested. This is a source-based risk, not a reproduced exploit or a claim that
   the genuine observation44 evidence is invalid.
3. **Parent audit and publication failure behavior.** Real scan/save failure and
   missing persisted verdict have not been exercised in the new scheduling model.
   Offline verdict tests cannot prove that the driver actually catches disk faults.
   A hung parent filesystem remains outside the process wait deadlines by design.
4. **Formal entry ownership and acceptance.** Inspected C12 entry has a separate
   lifecycle (profiles, resource ownership, receipts, Go/Docker deadlines). No
   direct staged-integrity reference was found in searched non-diagnostic .ps1
   sources. This is not proof about all dynamic invocations, but there is no
   demonstrated formal integration. Diagnostic SUCCESS must not replace its gates.
5. **Portability/release evidence.** Current scripts bind a machine, root and cached
   runtime; one archival decode check is not a clean-machine operational rehearsal.
   No authority to install dependencies, rebuild DLLs, alter trust guarantees or
   promote any of the five capability flags follows from observation44.

## What can be carried forward

Reuse as design/test evidence: separate transport/audit/verdict statuses, explicit
writer delegation, versioned trace contract, exact-byte checks and immutable
failure records. Reuse the typed empty-array fix in a future version, not by
overwriting consumed evidence. The existing binary identities are tested inputs,
not portable release dependencies ready for automatic redistribution.

Do not directly wire the fixed-root diagnostic driver into scripts/run-c12-integration.ps1.
First close verifier evidence-rejection gaps; then design the formal ownership,
configuration and deadline boundary. Parent scan latency and changed temporal
coverage must remain explicit. Full original timeout root cause remains unknown.

## Recommended next gate

Approve only the offline evidence matrix in
docs/superpowers/plans/2026-11-05-staged-integrity-offline-evidence-validation.md.
It proposes no child processes, Go/Docker work or new business computation.
Unexpected acceptance is a finding: preserve and stop; do not patch the verifier
until findings and a separate repair scope are reviewed. Even all fixtures passing
does not close blockers3–5 or authorize a main merge.
