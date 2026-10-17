# P1 fixture isolation — single-host localization record

Status: **BOUNDED DIAGNOSTIC COMPLETE / OBSERVED FAILURE LOCALIZED.** User approved
the [one-host plan](../superpowers/plans/2026-10-17-p1-fixture-isolation-diagnosis.md)
after81871cf9932861e6dd92272bf11d810906d782bb. The sole PS7 child met transport,
schema and integrity expectations: exit0 after690ms, no timeout, stderr empty.
Its diagnostic Disposition is ObservedFailure, NOT an isolation-test pass, fixture
acceptance, product repair or I1 reproduction. New allowance1/1 spent, no retry.

## Exact observed failure

Context DESKTOP-H4SVCLU\Lenovo admin=False, PS7.6.5 Core x64, ordinary inherited
token. No product/original test import or execution, old F1/warmup replay, fixture
edit, alternate constructor, second host or reviewer. The
[standalone probe](../diagnostics/isolated-fuzz/offline-supervision-policy.isolation-probe.ps1)
copies currentCF81 Type/WrapPrimitive/Decorate; static AST text matched exactly
after helper-prefix rename. Existing source/test/probe files stayed unchanged.

|Category|Input observation|Reached original isolation checks|Outcome|
|---|---|---|---|
|Int32|A/B both System.Int32; BaseObjectSame=true|ArraysABDistinct=true; ADescriptorNote=true; AFieldMatchesAArray=false|first failure at third check|
|Int64|all input measurements null|all11 NotReached/null|not evaluated|
|Boolean|all input measurements null|all11 NotReached/null|not evaluated|
|String|all input measurements null|all11 NotReached/null|not evaluated|

For Int32, the two returned metadata arrays are distinct. After both helpers
returned and am[0] was changed, A's Metadata descriptor is a NoteProperty, but
its Value is NOT reference-identical to the array returned as am. The remaining
eight Int32 checks were NotReached, including every B descriptor/reference/
content check and all C checks. Across four rows, only3 of44 declared original
checks were evaluated: two true, one false; the other41 are NotReached/null.
NotReached is not a pass or an independently observed failure.

Summary: FirstFailedCategory=Int32, FirstFailedCheck=AFieldMatchesAArray,
Disposition=ObservedFailure. WarmupReplayed/ProductImported/OriginalTestsRun/
FixtureChanged are all false. Saved stdout has exactly one context/host pair,
four ordered category rows, one summary and the diagnostic-only LIMIT marker.

## What this establishes, and what it does not

This localizes the new diagnostic's failure to a specific category/reference
predicate, unlike the earlier F1 group-only message. It is consistent with that
group failure, but the old receipt did not record the subcondition; do not
retroactively add this tuple to its immutable evidence or promote its false verdict.

Shared underlying A/B integer identity is observed. It does not by itself prove
shared metadata or a runtime mechanism. The failed predicate establishes only
that A.Metadata.Value and am differ at that point. No immediately-after-A-return
snapshot was taken, so the record does not prove when mismatch arose, whether
A.Metadata.Value equals bm, whether helper-return transport matters, or that
mutation/second decoration specifically caused it. B's checks were deliberately
not evaluated. Avoid saying the pointer was observed to switch to B's array.

The clean host omitted the preceding constructor/matrix/descriptor test cases.
The isolated path still exhibited failure, so those warmups were not necessary
for THIS observation; that does not establish they never influence behavior.
Instrumentation adds input-identity/type and predicate observations and may affect
wrapper transport. No PS5.1 result, language-wide boxing/cache explanation,
unique runtime root cause, independent review or authenticated execution history
is claimed. No technical source beyond local saved evidence was needed.

Potential next decision, NOT execution authority: a narrow independent-versus-
shared input-identity comparison could distinguish fixture-factory reuse from
return-transport/metadata effects before selecting a correction. It must preserve
the isolation assertions and account for the exact returned reference, not replace
ReferenceEquals with value equality. No such comparison, fresh boxing strategy,
fixture change or further probe was performed under this spent allowance.

## Frozen local-byte hashes and reconciliation

|Item|SHA256|
|---|---|
|New isolation probe|B6D8C6DD910369682C0636F5DCCA2768E95773EB63DE981323A58D0549B6E985|
|New controller|C008CA77DDE688DF5BD67E3517EAFBE21C41A935334F5C68E6829E117BEF6B44|
|Manifest/schema/pins|69108F17DA0E94FD255CBE1AB56C435CAF53108F9421C70866000870411D6CEA|
|Sole-run index|E1F08B7426628CE30797347E862F05B8F1E93078764CDFBE263D1924FD724EBB|
|Diagnostic stdout|4558EBA6BC75350F371B02A1147B0EFFC426D461C0C64B5DFD69CD043E8281B6|
|Empty stderr|E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|

Manifest protects35 entries: the fixture-revision packet's29 pins plus current
ownership suiteCF81 and its prior controller/manifest/index/F1 stream pair. All
matched before/after launch and parent final readonly reconciliation. Every
category/check row, halt/null behavior, summary and stream hash reconciled.
CurrentCF81 suite, productFE23DF3B, original35 suite61A3BACF, pre-edit9AA7 snapshot,
older diagnostic probe/raw and original false verdicts are unchanged. Integrity
consistency is not authenticated execution; Git line endings are separate.

## Rulings, budgets and recovery

- Approved one-host fixture-only diagnosis overrides generic extra TDD/baseline/
  task-done/whole-project runs, reviewer, automatic repairs and deletion. Cost:
  finite observations only, not independent review or repair confidence.
- PowerShell-native ledger/controller reused the verified linked worktree, no
  environment/security/trust/ownership change. Cost: narrower bookkeeping.
- Old packets read solely for integrity anchors and all retained. Cost: raw is
  local, not authenticated or restored by a clone.
- Only the original prefix was observed, without warmups or forcing past failure.
  Cost:41 checks/other engines and observer/context effects remain unproved; do
  not generalize or silently run a follow-up.

Retain NEW `.superpowers/sdd/2026-10-17-p1-fixture-isolation-diagnosis/` with ledger,
controller/manifest/index and original stdout/stderr; no packet deleted. Recover
portable probe/approved plan/report/status from latest development branch
codex/c12-b01-task9-coordinator, not main. Cloning does not restore local raw,
conversation or runtime state and grants no new run authority.

This diagnosis1/1 is spent. Fixture revisionF1 remains false and F2 frozen; original
product-repair slot1 remains false, six slots frozen/review ineligible. Earlier
two-host diagnosis remains spent2/2. Future R1–R7/source repair/conditional review
remain unapproved. Product I1, fixture acceptance, security/Stage B/P2/P3/main/
C12/release gates remain open. Next work requires a separately bounded decision.
