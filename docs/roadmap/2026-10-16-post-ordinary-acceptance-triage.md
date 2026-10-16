# Post-ordinary acceptance triage — source/report inspection only

Baseline:c35f42c549e719ca24a3393d974502730987c68e on
codex/c12-b01-task9-coordinator. User requested the recommended saved-only
remaining-gate triage after ordinary check001. This increment reads tracked
reports and current source/path inventories,not new native evidence. No Go,
test,binary,PID query,SDK inventory,Docker/DB operation or diagnostic replay.
Only this report/status/recovery documentation is changed;no implementation or
acceptance predicate is changed. No completion percentage is inferred.

## What is established, and what must not be repeated

| Item | Existing evidence | Current interpretation |
| --- | --- | --- |
| Ordinary nodeleafbinding | [check001](2026-10-16-leaf-ordinary-check-001.md):default-vet build and original runtime exit0;33 top-level/135 subtest PASS,0 SKIP;valid15-file completion binding | Scoped PackagePassed=true,Acceptance=false;001 consumed;no rebuild/replay/manual binary run |
| Statecontract scan defect | [bounded repair](2026-10-15-statecontract-scan-scope.md):exact diagnostic-subtree exclusion,RED controls then33 top-level/96 subtest package PASS | Specific scanner defect has scoped repair evidence;not a current unfixed scanner blocker;historical matrix00112PASS/1FAIL/13UNRUN unchanged |
| Diagnostic runner | [full runner](2026-10-15-full-runner-regression.md):31 assertions PASS,7 actual fixtures | Earlier descendant/full-runner-pending statements are historical;not whole-product validation |
| Latest limited matrix | [matrix003](2026-10-15-limited-matrix-003.md):2PASS/1FAIL/23UNRUN,check03 build root exit0 but member acquisition87 refused | Still failed historical matrix;new leaf package pass does not amend its counts or prove87 fixed |
| Original Envelope | [uninstrumented005](2026-10-15-envelope-uninstrumented-validation.md) short sample PASS | Does not explain historical deadline or supply Pair/complete fuzz acceptance |
| Task10 private bound read | [execution history](2026-10-08-task10-read-boundary.md) and current bound_read.go | Mechanism exists;not production v7/lease/exact-epoch authorization;do not restart from historical “no WithConsistentReadyRead” prose |

These are carried-forward execution records,not tests rerun by this triage.

## Remaining gates: validation and implementation are separate tracks

| Track | Concrete remaining boundary | Safe next action |
| --- | --- | --- |
| Historical process evidence | Matrix003 failed before a verifiable member handle was acquired;[analysis](2026-10-15-matrix003-acquisition-analysis.md) establishes no OS cause | Preserve error87 and fail-closed guard;no repeated matrix or error exemption. Any further diagnostic needs its own defined question/scope. Successful leaf run is not a root-cause experiment. |
| Historical ordinary/fuzz failures | Older timeout/cancellation findings and [Envelope deadline](2026-10-13-fuzz-deadline-diagnosis.md) are not causally closed by later successes | Keep named historical records;only investigate a specifically chosen event,not “retry everything until green”. This triage does not establish a new cause or enumerate every old finding. |
| Protocol foundation | Task11 sole v7 schema registry/canonical-envelope/evidence APIs are not found under the specified public names;planned registry/golden files absent | Next recommended development boundary:Task11 literal registry/role/ownership mapping and independent vectors,after scoped design approval;no live workload required for the design itself. |
| Exact-epoch serving | Task15 named Head/readiness files/tests absent;ordinary fence query still uses MAX(authority_epoch) | Do not relabel Task9/private Task10 as v7 readiness. Task15 consumes Task8,Tasks11–14 and Task10;cannot jump directly to production wiring. |
| Physical acceptance | PostgreSQL/PITR behavior and real connection/failover evidence remain separately gated | Check environment/ownership in a separately approved preparation;installed Docker alone is not daemon/image/physical acceptance. No Docker readiness was tested here. |
| Batch01 exit | Task18 integration manifest absent;required integration/golden cases incomplete by path inventory | Build manifest only from completed exact required cases;no placeholder to bypass runner,then separately run physical/ordinary/race/generated-artifact gates. |
| Integration/release | Broader C11/C12/Task9/10/15/18/security/production/main acceptance remains open | Development checkpoint only;no PR/main merge or deployment authority inferred. |

## Current source checks

- The Batch01 plan's [Task11](../superpowers/plans/2026-08-23-node-pop-control-plane-01-contracts-schema-authority.md)
 produces CanonicalBodyDigest,ParseCanonicalBody,VerifySignatureEnvelope and
 VerifyCanonicalEvidenceBundle through one schema registry. Definition search
 under internal found none of those exact exported functions or specified
 SchemaRegistry/SchemaID/TrustedSignerPolicy/VerifiedEnvelope/VerifiedEvidenceBundle
 type declarations. Planned testdata/c12/authority-v7/schema-registry.v1.json and
 canonical-envelope-golden.v1.json are absent;testdata/c12 itself is absent.
- Existing contracts/authority_v7_boundary.go,authority/v7_migration_grant.go,
 authority/v7_down_authorization.go,authority/v7_staging_import.go and stateformat/
 stateverify implementations contain v7 boundary/private canonical/envelope
 mechanisms. Missing planned files/public names therefore does NOT mean “no
 v7 or canonical implementation exists”. Reuse/reconciliation and preserving
 historical schemas must precede designing a sole public layer;do not duplicate
 canonicalizers or expose private fixture assumptions as production policy.
- authority/bound_read.go defines private Coordinator.newBoundReadSource and
 boundAuthorityReadSource.WithConsistentReadyRead. Their presence is not proof
 of Task15's complete Head/proof/incarnation/lease contract or physical DB behavior.
- db/queries/nodecontrol_authority.sql:GetAuthorityFenceHead still contains
 COALESCE(MAX(authority_epoch),0). That inspected ordinary query is not the
 required exact effective/terminal chain projector;no live SQL was executed.
- scripts/run-c12-integration.ps1:Invoke-C12SuiteMode checks exactly
 testdata/c12/integration-contracts-schema-authority.v1.json in a staged candidate
 and throws if absent. The literal path is absent in this checkout. This is a
 source/path observation,not a new attempted runner outcome or Docker diagnosis.
- matrix-check.ps1 remains a fixed26-check controller with no public resume or
 single-check selector. Do not edit its plan/check list or invoke low-level native
 helpers to skip history. New leaf evidence remains its separate fixed scope.

### Planned-path inventory (not task completion scores)

For each Task11–18 Files block,read Create/Modify/Test literal paths and Test-Path
each leaf. Existing Modify targets may predate the task. Absent paths might have
partial equivalents elsewhere;these counts are evidence of named deliveries,
not a functional audit,percent-complete metric or claim of zero capabilities.

| Task | Listed paths | Existing literal paths |
| --- | ---: | ---: |
|11|10|0|
|12|9|0|
|13|9|0|
|14|6|0|
|15|15|9|
|16|4|0|
|17|4|0|
|18|5|0|

Task15's existing9 are ordinary repository/query/generated-store targets;the
new database_head.go,exact_epoch_readiness.go,serving/postgres_repository.go and
three planned verification files are absent. No generated code is regenerated.

## Recommended next bounded scope

Return to Task11's foundation rather than another broad diagnostic/matrix run.
Prepare a written design for the **literal v7 schema registry and independent
positive/mutation vectors** as the first reviewable slice. It should:

1. Reconcile every §4.1 schema/Go type/signer role/algorithm/ownership row with
 existing code and historical schemas;do not infer that the old Task11 plan's
 unchecked boxes are current implementation evidence.
2. Define how private canonical/signature helpers can be reused or remain private
 without duplicate policy or falsely promoting fixture signers/ordinary readiness.
3. Bound the first implementation to registry/vector deliverables and their
 offline tests;leave public envelope/bundle APIs and later Task12–18 work explicit
 unless included by an approved follow-up design/plan.
4. Preserve unknown-schema/historical-reregistration refusal,strict types and
 finite errors. Do not import new cryptographic dependencies,relax security or
 run the old plan's fuzz/integration commands by implication.

This recommendation is not that design,implementation approval or permission to
start tests. Task11 contract work can be scoped without resolving87 first;
validation/physical gates nevertheless remain binding before broader acceptance.
If the user instead prioritizes test infrastructure,choose a distinct saved-only
87 evidence question;do not silently replace product work with another matrix.

## Recovery

Preserve this report,latest status and recovery pointer on the existing
development branch. The earlier post-build-only triage's first priority is now
superseded ONLY by scoped leaf ordinary evidence;all old raw outcomes remain.
Another computer recovers source/plans/reports from Git,not ignored SDK/cache/
binaries or live Docker/database state. Main is not the recovery default.
