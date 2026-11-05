# Staged integrity integration readiness review

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
