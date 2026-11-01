# Offline I1/I2 Single-Owner Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after explicit execution approval. No implementation agents/per-task reviews; only the conditional single fresh readonly runtime reviewer below. Track steps with checkboxes in this plan's own ledger.

**Goal:** Implement the approved private-result/single-owner contract and complete I2 observation checks in a new offline runtime, preserving all historical sources and verdicts.

**Architecture:** Two private writers return retained live ErrorRecords instead of rethrowing handled record failures. Each boundary consumes that result once, and the real parent consumes one complete Schema2 profile contract. New versioned source/tests, frozen literal expectations, bounded unit hosts and three actual public profiles establish separate evidence.

**Tech Stack:** Existing Windows PowerShell7.6.5 Core x64, .NET streams/processes/AST, fixed CreateNew records and one host-local Add-Type fixture compilation. No installs or product/CI/Docker integration.

**Spec:** [Approved single-owner design](../specs/2026-11-01-offline-i1-i2-single-owner-design.md), committed ce63170380700b0676fda06f1504bbb9d15c2eaf; user subsequently approved “是”. Read it together with [original I1/I2 review](../../roadmap/2026-10-31-offline-diagnostic-runtime-review.md), [stopped continuation](../../roadmap/2026-11-01-offline-i1-i2-continuation.md), [original fixed contracts §2–4](2026-10-31-offline-diagnostic-review-repair.md), and [limited formal data assessment](../../roadmap/2026-11-01-offline-checker-formal-assessment.md).

Approved design SHA256 A1173B1781FD01D9FB0195FA3BF4EF07D702902C0562355CB79EE9CF08164951; preserve those exact bytes rather than editing its historical pending-status text.

Status: PLAN_PENDING_EXECUTION_AND_NEW_BUDGET_APPROVAL. This turn authorizes this document only. “Production writer” here means actual offline runtime writer, not product integration. Design approval is not execution approval.

## Global Constraints

- “旧写入源码、测试、oracle、manifest、raw、portable、stop 与审查保持不变。” New runtime R=`scripts/offline-diagnostics-v4`; new leaf L=`docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2`; new packet P=`.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner`. No old controller/Grant replay or automatic renamed attempt after failure.
- “Session/Ack 不序列化、不从磁盘恢复；五能力标记仍 false。” Disk consent/manifest/result is provenance, not runtime capability. Explicit operator approval supplies this finite workflow authority.
- “同一个 ErrorRecord 引用只留一次；不同记录不因相同 Exception/文字合并；已有 HostPrimary 不替换；次因最多8个，第9个只置 Overflow。” No global wrapper suppression or exception/message-based deduplication.
- “任何失败都不重开/重试。” First actual unexpected source/provenance/record/test/transport failure or unknown stops the remaining stages; the two exact availability REDs below are the only intended failures. No automatic repair/retest/re-freeze/review retry.
- “不动 api/cmd/internal、原 seam/codec/inbox、CI、Docker、默认验证脚本和 main。” No installs/security changes, Force reload, PID/name lookup, kill-tree, Job/watchdog, prewarming/readiness delays or enlarged time bounds.
- All old budgets remain frozen, including v3 remainder6hosts/7targets/1compile/conditional1review. Request NEW budget below, not release/refund of those units. No runtime allowance consumed this turn.
- Native finite scope replaces skill-default Bash/setup/full-repository/task-done test reruns, extra agents/reviews/fix passes and workspace deletion. Archive on success OR stop; ordinary necessary checkpoint push to existing branch is authorized, main merge is not.

## Review Focus

1. A normal function return must not convert intent failure into success: Schema2 INCOMPLETE plus zero work/targets/Ack, tested Writer/IntentFailureNoWork and I2 intent-failure fixture.
2. Wrong/duplicate outcome consumption or projection failure must not discard first cause: Writer/OutcomeConsumedOnce, MalformedOutcomeStops, SummaryFaultKeepsPrimary and ProjectionFaultKeepsPrimary.
3. A collector must not merge different records sharing one Exception: Writer/SharedExceptionDifferentRecordsKept and independent same-text real-fault control; never claim a synthetic control is a measured runtime event.
4. A complete-looking report with coercible types/unknown fields/foreign source is invalid: I2 additional18 vectors plus original20 mutations, with exact rejection signature and valid controls.
5. Private fixture factories must not contaminate normal/final publishers: real default factory checks, three recorder guards in every GREEN/profile host, complete source/call-chain review; fixture success is not OS late-I/O proof.

## 1. Exact source ownership and stable interfaces

Create31sources, individually enumerated; no runtime output/portable joins the source set:

| Root | Exact files and responsibility |
| --- | --- |
| R (11) | `.gitattributes`, `diagnose-offline.ps1`, `private/catalog.json`, `private/session.psm1`, `private/sinks.ps1`, `private/owned-step.ps1`, `private/sequence.ps1`, `private/profiles.ps1`, `private/observations.ps1`, `private/write-boundary.ps1`, `private/record-owner.ps1` |
| L (20) | `.gitattributes`, `README.md`, `expected-output.json`, `expected-observations.json`, `test-helpers.ps1`, `entry.tests.ps1`, `core.tests.ps1`, `prefix.tests.ps1`, `owned.tests.ps1`, `live.tests.ps1`, `capture.ps1`, `driver.ps1`, `record-boundary.tests.ps1`, `observation-contract.tests.ps1`, `fault-stream.cs`, `observation-contract.ps1`, `source-model.ps1`, `recording.tests.ps1`, `case-module.ps1`, `publisher.psm1` |

R entry/private and L source roles derive from v3/integrated-review-repair counterparts, not executed old scripts. Add local LF attributes; R source responsibility stays narrow. `record-owner.ps1` handles private consumption; `expected-observations.json` is independent literal full-contract data; `source-model.ps1` builds readonly fixed contexts/AST extents outside the pure validator. New tests may update Schema2/path/cross-layer assertions BEFORE freezing; old tests are untouched.

Private signatures and ownership:

- `New-ODRecordStream([string]Path) -> Stream`: unchanged real CreateNew/Write/None; no production injection parameter.
- `Write-ODCaptureRecord([string]Path,[byte[]]Bytes)` and `Write-ODProfileIntentRecord([string]Path,[byte[]]Bytes)` -> exactly one PSCustomObject, ordered `Succeeded,PrimaryError,SecondaryErrors`; explicit ErrorRecord[] even empty. Success=true/null/empty; failure=false/actual first record/0..1 actual Dispose record. Catch only real resource calls, Dispose held object once; no Capture-Failure inside writers and no handled-record throw.
- `Register-ODRecordOutcome([object]Outcome,[object]Consumption) -> no output` in record-owner: Consumption is private `{Consumed=false}` created once by owner; consume it BEFORE validation/registration. Malformed/repeated consumption throws `OD: record outcome contract mismatch`, preserving any existing first cause. Valid failure registers first then cleanup exactly once using unchanged Capture-Failure; success registers none. Consumption is not an exported security/authorization object.
- `Import-ODProfileFailure([object]Report,[object]Consumption) -> no output` in record-owner: same private consume-before-use rule; import only retained live Schema2 report error references, never projected/deserialized records. Capture owner handles `Status!=COMPLETE` explicitly and stops, not by throwing the report as an error.
- `Invoke-ODProfile -Profile <fixed-id>` keeps sole export, consumes attempt before intent I/O; new reports uniformly Schema2 with the19ordered root keys in design §3.1. Intent failure: no work body, Groups/Observations empty arrays, actual Closed and actual retained errors; metadata/summary defects remain unknown, not fabricated reports.
- Publisher loads only writer definitions/default factory, no HostPrimary/HostSecondary collector; it returns the live outcome to capture. Capture owns final-record failure registration and final exit disposition, never retries that sink. Profile has its own collector and passes failure references through ProfileReturn once.
- `Assert-ODProfileObservations([object[]]Rows,[string]Profile,[Collections.IDictionary]Context) -> no output` and `Assert-ODProfileReport([Collections.IDictionary]Report,[string]Profile,[Collections.IDictionary]Context) -> no output` in L/observation-contract.ps1. Both mismatch only RuntimeException Message/FQID=`OD: observation contract mismatch`; arbitrary code errors fail verification. Report validator checks all19keys/types/metadata and calls the SAME row validator for complete reports. Valid INCOMPLETE representation is not a successful profile disposition.

Context carries fixed Repo/PacketRoot/ProfileRoot/Oracle/Sources plus independently parsed SourceModel, Entry/Kernel/Manifest identities. Historical data-unit context is fixed to original namespace/source bindings; actual profile context is fixed to new v4 namespace, never switched to history to accept a bad runtime report. No IO/process/code evaluation in either validator. Parent strict JSON decode/duplicate/framing checks and source reads live in driver/source-model, not inside pure predicates.

## 2. New finite budget and stage order

Propose exactly TWO availability RED computations in existing tool contexts (no launched diagnostic host), followed by NEW5capture hosts /7controlled targets /1in-memory fixture compilation /conditional1fresh readonly runtime review. No spare rerun or extra final host:

| Stage | Hosts / targets / compiles | Required result |
| --- | --- | --- |
| writer-availability-red |0/0/0|OwnerAPIAvailable FAIL;25NotRun; exit1; absent new writer/owner API only|
| contract-availability-red |0/0/0|ContractAPIAvailable FAIL;51NotRun; exit1; absent new validator API only|
| writer-green |1/0/1|26PASS,0NotRun, exit0|
| contract-green |1/0/0|52PASS,0NotRun, exit0|
| core |1/0/0|25PASS,22complete private rows, exit0|
| lifecycle |1/3/0|17PASS,8complete private rows, exit0|
| continuation |1/4/0|16PASS plus12original seam PASS,15complete private rows, exit0|

RED proves new API availability guards, NOT a reproduction/regrade of historical4FAIL RED or actual I1 acceptance. Its tests/expectations are frozen first; no Add-Type, writer fixture, validator cases, imports of runtime module or process launch behind availability. Save full tool results externally after terminal completion; neither RED depends on an unimplemented final publisher to record itself.

OwnerAPIAvailable checks the five named functions New-ODRecordStream, both Write-OD*Record functions, Register-ODRecordOutcome and Import-ODProfileFailure as one conjunctive availability group. ContractAPIAvailable checks both Assert-ODProfileObservations and Assert-ODProfileReport. Missing APIs yield only the specified FAIL/NotRun model, without invoking any later fixture or publisher definition. Define test functions Invoke-ODWriterTests/Invoke-ODContractTests to return this availability report before doing any fixture/module work; regular GREEN uses the same functions after the gate.

GREEN budgets include original3recorder groups, source/post-pin/lifecycle/transport checks, healthy real final publication and journal sequence. Do not count a complete stdout PASS list as exact if final/probe/stream/source checks fail. Conditional review only after all five host finals and both REDs exact.

All process limits remain fixed: capture120000ms except continuation150000ms; sequence140000ms/cleanup4000ms; owned exit2000ms/shared stream drain2000ms; hold45000ms; normal/receipt2000ms; timeout200ms; seam110000ms. Sessions/sinks≤32, Ack≤2, observations≤32, distinct secondaries≤8. Engine SHA362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139 and SMA2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B at original pinned paths; original fixed identity/adminFalse unchanged. Missing environment is a blocker, not permission to install or loosen checks.

## Task1 — fixed new chain/tests and TWO availability REDs

**Files:** create all31R/L sources, P consent/ledger/binding/snapshots/consumed RED intents; no old source edit.
**Consumes:** approved design/plan, original source/data identities; produces frozen literal26/52group suites and source baseline for later narrow windows.

- [ ] Verify clean linked existing branch/origin, absent R/L/P/output, explicit execution approval including NEW budget and changed cross-layer/Schema2 contracts. Read all actual source/test counterparts before copying. No historical command execution.
- [ ] New preflight must compare the original seed manifest SHA9411278BDB6636F38C1D643C647C754194B0D67249958342BF26FC49D6DFDDC1 and every2482Present pin/28snapshot to actual bytes, plus explicitly referenced historical report/data archives. This is newly authorized full checking, not a claim that the earlier failed audit completed. Missing raw originals stop; never reconstruct them from portable.
- [ ] Mechanically adapt paths/source bindings and wire caller-side Schema2/outcome handling before first freeze. Do not implement writer/owner APIs in write-boundary/record-owner yet; keep those files definition-free apart from source comments. Do not implement the two validators in observation-contract yet. This intentional availability seed is NOT a preserved buggy algorithm. All other source/helpers, complete tests and literal oracle must exist and be frozen before RED.
- [ ] Freeze independent19key ProfileReport schema; complete core22/owned8/live15 inventory, all View/Sinks/IORoot/Result/Error/summary/sequence fields from normative tests/contracts, not from future GREEN. Existing closed-file/missing-file/invalid-JSON/invalid-index/output-error exact templates remain as in original contract §2–3. Use literal matrices/counts/error strings; paths/AST/PID/time vary only by explicit fixed rules.
- [ ] Readonly static source checks: no reserved automatic-variable writes, LF/JSON/AST validity, bounded source lists, caller→writer/owner and driver→both validator binding, unchanged collector algorithm, live outcome not serialized before import, no failure→work/Ack path, no module exports except Invoke-ODProfile. Save results, do not call RunApproved/Prepare/tests as a static check.
- [ ] Create initial seed manifest from the explicit inherited2482mapping and named new31sources plus approved design/plan/consent. Canonical deduplicate only identical path/state/hash; fixed existing Present values do not become Absent. Exact combined count and path map recorded before freeze; no repository-glob or mutable progress/tool receipt as runtime pin. Freeze all31source snapshots once with CreateNew; plan/spec/consent snapshots separate. Scope outputs cannot appear while inventory-sensitive freeze is still running.
- [ ] Consume unique writer RED intent, load only new test definitions, run availability branch once. Require26groups:1FAIL/25NotRun/exit1 for missing API only. No unrelated exception accepted. Then consume contract RED intent similarly:52groups,1FAIL/51NotRun/exit1. Preserve complete exact tool chunks and frozen source/test hashes. Either unexpected outcome stops all later stages.

After each separate CreateNew intent, execute the corresponding body in its own approved tool context (do not run these while planning):

```powershell
# Writer RED body; definition loading must not invoke factory/module/fixture.
$oracle=Get-Content 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/expected-output.json' -Raw|ConvertFrom-Json -AsHashtable
. 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/record-boundary.tests.ps1'
$report=Invoke-ODWriterTests
$report|ConvertTo-Json -Depth 20
exit $report.ExitCode # expected1,26ordered groups:1FAIL/25NotRun
```

```powershell
# Contract RED body; same availability-first definition-only rule.
$oracle=Get-Content 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/expected-output.json' -Raw|ConvertFrom-Json -AsHashtable
. 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/observation-contract.tests.ps1'
$report=Invoke-ODContractTests
$report|ConvertTo-Json -Depth 20
exit $report.ExitCode # expected1,52ordered groups:1FAIL/51NotRun
```

Each test function's absent-API branch returns exactly one report with Schema/PASS/FAIL/NotRun/ExitCode/Groups; group names/order come from the frozen literal oracle. Its available-API branch uses actual LS-Result/LS-Details and the host's complete evidence, not a fabricated availability report. No test utility code executes simply by dot-sourcing these files.

## Task2 — I1 actual private-result writers/owners, one GREEN host

**Files:** change ONLY R/private/write-boundary.ps1 and record-owner.ps1; all29other sources/tests/oracles remain frozen. P writer-green manifest/snapshots/records.
**Consumes:** Task1 frozen writer26group model/caller wiring; produces actual producer outcomes and single-owner registration used by profiles/publisher/capture.

- [ ] Implement signatures in §1: real operation catches only, first cause retained locally, held stream disposed once, writer outcome returned without collectors/rethrow. Owner consumes once, validates exact shape/types, registers first→cleanup and preserves preexisting HostPrimary. Outcome arrays never scalar-unroll. Do not change global collector deduplication.
- [ ] Ten literal cases (Capture and Intent ×5), using original real-fault fixture behavior in isolated private modules: WriteDisposeFirstCause requires Open1/Write1/Flush0/Dispose1, false success, actual Write primary and1actual Dispose secondary; FlushDisposeFirstCause=1/1/1/1 with Flush primary/1Dispose secondary; DisposeOnly=1/1/1/1 with Dispose primary/0secondary; SuccessOnce=1/1/1/1, true/null/empty and exact bytes65,66,0,255; OpenFailureOnce=1/0/0/0 with actual IOException, existing bytes80,82,73,79,82 unchanged. Operation messages/Type/ErrorId/source must match original §2fixed templates/new frozen AST. Compare returned/captured live references, not outer-catch wrapper identity.
- [ ] Writer suite core23groups: OwnerAPIAvailable + ten cases above + twelve named controls below. Then three recorder groups: RecorderIdentityIsolation,RecorderPrivateState,RecorderProductionBinding. Freeze their NEW expectations before RED: outer recorder command/collector/factory references unchanged; publisher exports0 and has NO HostPrimary/HostSecondary/HostOverflow collector state; cases cannot leak private replacements; actual publisher uses default source-bound factory and returns Succeeded=true/null/empty with exact probe bytes. Consume that healthy probe outcome once through actual capture owner. Do not retain the old publisher-state assertion that expected a private empty collector.

| Twelve owner controls | Exact assertion |
| --- | --- |
|SameRecordOnce|re-register same live ErrorRecord does not add secondary|
|DistinctSameTextKept|two separate real missing-file faults with same text remain two records|
|SharedExceptionDifferentRecordsKept|two explicitly constructed wrapper records sharing one real Exception remain distinct; labelled contract-only, not real I/O measurements|
|ExistingPrimaryPreserved|preexisting real HostPrimary reference unchanged; operation first/cleanup appended in that order|
|OverflowBound8|8distinct real secondary records retained; ninth only sets Overflow; no count growth|
|OutcomeConsumedOnce|second use of exact private consumption holder rejects before another registration or IO|
|MalformedOutcomeStops|wrong keys/types or true+nonnull error, false+null error, scalar/>1cleanup rejected; retained first cause not cleared|
|IntentFailureNoWork|execute actual new Invoke-ODProfile body in isolated module with fixed fault factory; Schema2 INCOMPLETE, actual error refs, Groups/Observations empty, body/starts/Ack0|
|PublisherSingleOwner|actual publisher returns live outcome without private collector; actual capture owner consumes once; original refs preserved|
|SummaryFaultKeepsPrimary|real closed TextWriter summary fault adds bounded secondary and cannot replace original; no second summary attempt|
|ProjectionFaultKeepsPrimary|actual unsupported evidence value causes projection failure, original retained, unknown/INCOMPLETE not success; no fabricated ErrorRecord|
|FinalWriteFailureNoRetry|actual final-writer fault consumes evidence attempt before IO; no reopen/second publication/default-factory replacement|

- [ ] FaultStream compiled once INSIDE writer-green host. No compile/target/fixture run in RED or separate helper host. Test factory replacement stays in private per-case module; invoke real default factory for healthy/existing cases. Intent/publisher/capture behavioral checks use actual new function definitions, not mock implementations or product module injection. Expected fault wrapper signatures must be literal before seed; unexpected methods/errors stop.
- [ ] IntentFailureNoWork must not depend on the not-yet-created repair-green manifest. Its isolated case-module uses the actual profile FunctionDefinitionAst and actual writer/owner/collector definitions, with a fixed unit metadata binding to the CURRENT writer-green manifest and its own unused case root. Supply private phase/path constants only in that isolated module; normal session.psm1 fixes them to P/repair-green and exposes no test switch or caller context. Use core profile only (zero targets) and the private one-shot failing factory; no public RunApproved/full session module import. Missing binding or reaching work body is a test failure, not permission to manufacture a repair-green success file. Freeze this separation and metadata schema before both REDs.
- [ ] Prepare writer-green once against actual seed SHA: only these2source pins change, other inherited/new pins equal;31new snapshots. Host launch consumes unique intent first. Run `L/driver.ps1 -Action RunOne -Grant writer-green -FrozenSHA256 <actual-writer-freeze>` once. Require26PASS/exit0, exact transport/recorder/probe and original pins. Record complete original tool chunks only after actual completion. No fixture result substitutes for real final publisher success.
- [ ] Preserve necessary source/docs/portable checkpoint, staged-byte identity and push equality; do not call Task2done/default suite. Only next planned stage may proceed if exact.

## Task3 — I2 full pure contract, one GREEN host

**Files:** change ONLY L/observation-contract.ps1; all30other sources/tests/oracles/driver frozen. P repair-green manifest/snapshots/records.
**Consumes:** frozen52group model, source contexts and Task2 new live report contract; produces both actual parent predicates from §1.

- [ ] Implement exact ordered19key Schema2 root and complete rows recursively. View/Sinks/IORoot/errors/summary and sequence are mandatory, unknown/missing/duplicate/order/type changes reject. Derive source locations from frozen SourceModel, not an arbitrary source substring/valid-line range. No IO/process/import in pure functions. Do not blanket-catch implementation bugs as contract mismatch.
- [ ] Contract core49groups: ContractAPIAvailable1 + historical rows BaselineCore/BaselineOwned/BaselineLive3 + original Reject01..20 + Schema2CompleteCore/Owned/Live3 + Schema2IntentFailure1 + AdditionalReject01..18 + StaticReservedNames/StaticPureContract/StaticAllConsumersBound3. Add the same3recorder groups for total52.
- [ ] Original20mutations retain original §3definitions and exact rejection signature; the3actual saved ProfileReturn frames are readonly data with original paths/source context. They do not certify new v4 runtime. Schema2 data fixtures are independent literal wrappers around valid rows with fixed historical test context; actual parent context only accepts new source/path identities.

| AdditionalReject01..18 | One independent mutation each |
| --- | --- |
|01–06|Schema1 instead of2; missing root key; extra root key; swapped root key order; a capability flagtrue; Groups scalar|
|07–12|Observations scalar; SecondaryErrors scalar; PrimaryError string; intent-failure Status COMPLETE with empty rows; Closed string; failure report missing actual cause|
|13–18|success report carries unexpected primary; foreign ProfileRoot; wrong AST source line; integer counter string; boolean field string; fabricated Schema2 UNKNOWN/partial object treated as complete report|

- [ ] Schema2IntentFailure is a positive representation check only: Status INCOMPLETE, empty Groups/Observations, actual projected primary and bounded cleanup, Closed from actual outcome. Parent disposition MUST remain unsuccessful. Truly unknown exceptions/partial reports cannot be manufactured into this known-failure branch; AdditionalReject18 and actual capture behavior guard it.
- [ ] Prepare repair-green once against actual writer-green SHA: only1contract pin changes,31snapshots; other pins unchanged. Run `L/driver.ps1 -Action RunOne -Grant contract-green -FrozenSHA256 <actual-repair-freeze>` once;52PASS/exit0 required, all valid controls accepted, negatives exact mismatch, complete records/recorder/source checks. No actual targets and no Add-Type.
- [ ] Preserve source checkpoint and actual binding; freeze ALL runtime/test/oracle files thereafter. Total new source snapshots31×3=93. No second GREEN even if failure seems cosmetic.

## Task4 — three actual v4 entry profiles, same complete validator

**Files:** source frozen; only exact P runtime records and handoff docs/archives.
**Consumes:** repair-green identity and real new entry/profile/owner/publisher/validators; produces real consumer evidence, not restored historical authorization.

- [ ] `L/driver.ps1 -Action RunOne -Grant core -FrozenSHA256 <actual-repair-freeze>` once:25PASS (original core22groups plus3recorder),0targets; default/explicit Inspect without runtime import/creation,14request refusals before IO, prefix98 conditions and complete22private observations. Actual ProfileReturn Schema2 checked by both parent predicates, no Step-only fallback.
- [ ] Same command `-Grant lifecycle` once:17PASS,3targets,8private rows. Prestart Start0/unassociated/noKillWait; real receipt fault one attempt/first cause/exact owned exit+drain; normal and timeout original tuples. Full context/source-bound contract mandatory.
- [ ] Same command `-Grant continuation` once:16PASS plus12unchanged seamPASS,4targets,15private rows. Active valid-step nonzero remaining budget missing/wrong Ack controls, genuine Ack consumed once, six invalid/replay requests, actual output fault without B contamination, actual A1seam original output/trace. No extra final target/host.
- [ ] Every host uses one retained capture Process, finite waits, one consumer per original stdout/stderr stream. Tool session yield is not completion; wait only returned session ID, preserve all chunks and terminal result before parsing/writing full receipt. Lost/unknown/nonzero recorder/transport/final result freezes remaining grants/reviewer immediately.
- [ ] Parent final success requires exact schema/disposition, expected raw stdout, empty stderr, capture exit0, owned exit/drain, no rescue/timeout/termination and intact post-pins. No success inferred from journal/summary/PASS output alone. Healthy final publication uses actual new private-result writer and capture owner.

## Task5 — one conditional readonly runtime review and handoff

**Files:** no source changes; one actual review package/result, execution report, new v4 operator guide, current-status/recovery and L/portable-final.json.
**Consumes:** both exact REDs, five exact host finals, all93snapshots and actual new manifests; produces narrow proposed disposition, not automatic closure.

- [ ] Only after all exact checks, use NEWconditional1fresh-context readonly runtime reviewer (not the earlier document feasibility check). Review complete public entry→session/profile→actual writers/owners→publisher/capture→parent/I2 chain plus source-bound evidence, spec/plan/rulings and all5Review Focus. No code execution/import/network/writes/Git/subagents by reviewer; no second reviewer/fix/retest authorized.
- [ ] Record all findings and declined-to-judge items; Critical/Important/incomplete review stops as unaccepted checkpoint. Even if clean, label only new v4 offline boundaries and exactly exercised conditions. Old I1/I2/C1/C2 records are not rewritten; R2/Minor2 and uninduced OS late-I/O/host-death/hard synchronous-block guarantees remain deferred/unverified.
- [ ] Success OR earlier stop: preserve every actual intent/source snapshot/manifest/original stream/partial/zero-byte/error/tool chunk/review and exact directories as byte/hash-labelled base64. Archive actual new originals, never create missing “originals” from old portable. Verify decoded byte equality and explicit staged blobs. Preserve input integer text without JS-number reserialization of raw evidence.
- [ ] Update `docs/roadmap/2026-11-01-offline-i1-i2-single-owner.md`, new `docs/runbooks/offline-diagnostics-v4.md`, current-status/repository-recovery with real consumption/remainder/stop, authoritative original versus new verdicts and machine-recovery limits. Explicit necessary sources/tests/docs/archive commit/push existing development branch; verify remote SHA=HEAD. No main merge, branch deletion, secrets/raw/cache/binary commit or workspace cleanup.

## Self-review / approval gate

Spec coverage: single-owner outcome and cross-layer change→Tasks1–2; unchanged deduplication/first-cause bounds/real resources→Task2; Schema2 complete and known-failure contracts→Task3; actual default factories/parent consumers and lifecycle→Task4; conditional whole-chain review/limits/recovery→Task5. Five review risks each have named checks. 31sources,93snapshots,26writer groups,52contract groups, two0hostREDs then5hosts/7targets/1compile/conditional1review. Availability RED does not claim historical defect reproduction; successful data schema is not runtime integration acceptance.

Please approve this complete plan, Native method, changed private-result/Schema2 contracts and NEW finite budget explicitly before implementation. Planning consent/design approval alone authorizes none of these runs. Old frozen budgets are not released, and any new unexpected failure stops without automatic retry or fix pass.
