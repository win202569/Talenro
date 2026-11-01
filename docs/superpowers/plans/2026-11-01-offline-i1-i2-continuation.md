# Offline I1/I2 Continuation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after explicit approval. No implementation agents/per-task reviews. Only the conditional single fresh readonly reviewer in task5. Steps use checkbox syntax for tracking.

**Goal:** Resume the never-run v3 I1/I2 verification sequence against its actual prepared seed, repair only its two designated defects, and preserve a narrowly reviewed, recoverable result.

**Architecture:** Preserve the prepared28source baseline, existing driver and historical stop marker. Add a separately bound continuation controller beside the existing driver, mechanically retaining its full evidence/process/manifest validation while explicitly controlling the nine still-unconsumed actions. Keep the original runtime namespace and entry; no v4 migration, seed rewrite or historical rerun.

**Tech Stack:** Windows PowerShell7.6.5/Core/x64, pinned existing engine/.NET Stream/Process, two in-memory writer fixtures only, CreateNew records and SHA256. No installs, Bash, Docker or product integration.

**Spec:** [Approved offline design](../specs/2026-10-19-offline-diagnostic-entry-migration-design.md), [integration plan](2026-10-31-offline-integrated-review-repair.md), [I1/I2 exact vectors and contracts, sections2–4](2026-10-31-offline-diagnostic-review-repair.md), [independent findings](../../roadmap/2026-10-31-offline-diagnostic-runtime-review.md), and [actual successful preparation](../../roadmap/2026-11-01-offline-prepare-recording-adjustment.md). This document proposes only the explicit control/authority adjustment below; all original semantic and lifecycle requirements remain normative.

Status: DRAFT_PENDING_EXECUTION_APPROVAL. Baseline a1db8b6fdb2cfa36a55d0b51ad54263086877138. User approved drafting this plan only. No continuation source, imports, Prepare, capture hosts, targets, compilations, agents or review in this turn.

## Global Constraints

- “只涉及离线诊断，不做生产接入”; no CI/product/default verification chain, main merge or release.
- “默认为 Inspect”; no runtime import, Session, root creation or target during Inspect. Public action/profile IDs and sole module export Invoke-ODProfile unchanged.
- “Session/Ack 不序列化，不从文件加载，不跨 shell/进程转移”; records describe attempts, never restore live authority or acknowledgment references.
- “首次实际失败后，后续逻辑 Read/Parse/Save/Complete/Publish/重新打开在文件访问前拒绝”; no retry, fallback evidence, source patch after unexpected failure or success inferred from printed PASS.
- “不加入 PID/名称重查、杀树、Job/watchdog、无限等待、启动预热或额外 readiness 等待”.
- Preserve original driver, seed manifest,28seed snapshots, all original five records, adjustment markers/sidecar and both portable archives. Existing historical stop.json remains unchanged. Never run the old driver's preparation exception again.
- Only writer-green changes scripts/offline-diagnostics-v3/private/write-boundary.ps1; only repair-green changes the integration leaf observation-contract.ps1. Other26original runtime inputs never change. The new controller is frozen separately before any action.
- Request release ONLY the unused v3 remainder7capture hosts/7targets/2compilations/conditional1review, not an additional grant of the same size. Older v2 remainder6/7/1/conditional1 remains FROZEN; diagnostic/isolation/preparation allowances remain exhausted.
- Explicit approval must cover the new controller and retained-stop exception as well as release of that remainder. This proposal, disk approval data and frozen seed alone authorize nothing.
- Approved finite Native/PowerShell replaces skill-default setup/Bash/baseline/full-suite/task-done reruns, per-task agents/reviews, automatic fix passes and deletion. Only named seven hosts and conditional reviewer; no extra gate-test host.
- Any actual unexpected static/preparation/runtime/recording/integrity failure freezes every remaining action/review. No retries, renamed attempts, new runtime packet, refund or control/assertion edits after results.
- All five capability flags remain false. I1/I2 unresolved until evidence and review support a new narrow disposition; historical C1/C2 open, R2/Minor2 deferred. No full-C12/production acceptance claim.

## Review Focus

1. Retained historical stop must not hide a new failure: separate create-once continuation stop/attempt records and complete parent errors, tasks1–5.
2. Controller not covered by2482runtime pins must not drift or silently relax checking: separate retained control binding, original-to-new region comparison and whole-controller review, tasks1/5.
3. Write/flush plus cleanup failures must preserve original references across BOTH actual production writers: unchanged ten vectors/real publisher binding, task2.
4. Arbitrary validator errors must not masquerade as intended schema rejection: exact23vectors/signature, full22/8/15observations through SAME validator, tasks3–4.
5. Actual capture publication/journal/owned exit/drain and valid-budget missing-Ack checks cannot be replaced by summary PASS: unchanged exact parent predicates/recorder groups, all seven hosts and task5.

## 1. Choice, exact ownership and seed

Recommended A: a new fixed continuation controller, no change to frozen driver.
Alternative B, patching the frozen driver, would require an explicit runtime-input
identity transition/refreeze before RED; this wastes the verified seed and expands
the change window. Alternative C, deleting/moving stop.json or copying another
runtime tree, obscures the actual history. This plan recommends A; it is not yet
user-approved. Approval of this complete plan selects A, not automatic execution
from the earlier preparation consent.

Runtime packet P remains `.superpowers/sdd/2026-10-31-offline-integrated-review-repair`.
Leaf L remains `docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair`.
Create only one new control source: `L/continuation-driver.ps1`. It is NOT a
public entry, alternate capture or generalized replay tool. It does not invoke
the existing stopped driver. Public runtime remains scripts/offline-diagnostics-v3/diagnose-offline.ps1.
New recording/control sidecar C is `.superpowers/sdd/2026-11-01-offline-i1-i2-continuation`, not a replacement runtime namespace.

Actual seed manifest SHA256:
9411278BDB6636F38C1D643C647C754194B0D67249958342BF26FC49D6DFDDC1.
Existing driver SHA256:
2C488ABB4D5565984D4459E3D7025A2BE5854618105A26E8187B71DEDFE6E647.
Preparation portable SHA256:
D6606BCD7DD03157205C67B061D22C03A80F46BF279B0C32DC41E3A15164030B.
Original stopped portable SHA256:
1D10ABA536D414B06BE67422EA3662EF68AF4926858A70EBBBCFAB38E1EE6ADF.
Original stop SHA256:
0D4FE34A03FAD3FD1D45B8C4B850C5B6D8C8BF9F3C9283304503B7119FD7EDE2.

Before execution, independently compare actual2482seed pins/28snapshots and
all44archived files/12directories. Do not rebuild missing local originals from
archives and call them historical originals. On a new computer, mismatching
pinned absolute paths/engine/identity is a stop requiring separately scoped
environment preparation, not permission to rewrite this plan or the manifest.

## 2. Fixed control interface and one-shot action order

Controller parameters: original Action, Grant, FrozenSHA256 and
PreviousFrozenSHA256 plus required64uppercase-hex ControlBindingSHA256.
No PreparationAdjustment, arbitrary paths, callbacks, Resume/Retry/Force/skip
switches, user-selected runtime namespace or dynamic grant registration.

Create C/approval.json with actual explicit execution approval, baseline,
this approved plan's actual SHA, identity `offline-i1-i2-continuation-2026-11-01-1`,
ReleasedHosts=7, Targets=7, Compilations=2, ConditionalReviews=1 and exact
action order below. Snapshot approved plan and complete control source in C.
Create C/control.binding.json once, binding controller SHA, actual approval SHA,
plan SHA, seed SHA, fixed source/packet paths and original stop SHA. The executor
retains its actual hash and supplies that exact ControlBindingSHA256 on every
action, never auto-refreshing on drift. Files/hash data attest explicit consent
and source identity, not cryptographic user identity or a security sandbox.

Effective verification has2482runtime pins PLUS four separately bound control
artifacts (controller, plan, approval and binding). Do not claim2482covers all
inputs. Keep runtime schema2/2482 and original dynamic phase contracts unchanged;
no new manifests overwrite seed. Control source/plan/approval snapshots are
separate provenance, not additional runtime snapshots counted toward84.

Mechanically derive the controller from the actual existing driver's complete
functions and Prepare/RunOne bodies at this baseline. Preserve all ordered
evidence/recorder/observation/error/exit/drain/pin checks, engine/SMA checks,
CreateNew outputs and exact process-object ownership. Only change parameter and
control-gate regions, add fixed continuation control functions, and redirect its
outer unexpected-error marker to C/stop.json. In particular do not copy only the
launch loop or drop validation helpers. Same PSScriptRoot keeps existing fixed
test/validator/capture paths; launch the existing frozen capture.ps1, never a new copy.

Control functions (internal, no exports):

- Assert-ODContinuationControl(Action,Grant,FrozenSHA256,PreviousFrozenSHA256,ControlBindingSHA256): no-output precondition check. Requires fixed identity/approval/binding/source hashes, exact old stop identity, absent C/stop.json, next action ordinal, absent relevant outputs and original phase predicates. Reject reparse/substituted control/packet paths. Reject seed Prepare unconditionally. Recheck control bindings before/after every action. RunOne checks its complete phase pins before/after; Prepare checks unchanged pins plus exactly its designated changed source before, then complete new pins afterward. Do not reject the expressly permitted source edit by applying an unconditional previous-phase live hash check.
- Start-ODContinuationAttempt(Ordinal,Action,Grant): create C/actions/NN/intent.json once with binding/phase hashes and consumed=true, flush/close before running the body. Existing intent consumes that action even if launch never starts. No resume after missing/unknown completion. Restrict NN to01..09 and fixed action table, not a caller path.
- Complete-ODContinuationAttempt(Ordinal,Exact): create unique terminal control result only after the existing complete action result is checked. Next ordinal requires all earlier terminal results exact and their archived evidence/manifest/snapshot identities intact. Check current live source against the appropriate phase/change window, not a superseded phase's live writer hash. This is workflow accounting, not serialized Session/Ack or recovery of runtime permission after crash.

Replace both original blanket historical-stop rejections only in the NEW
controller with the fixed current continuation checks. The historical marker
must always match the exact known old error; a different/missing marker fails.
Do not accept any new error because that old marker already exists. Catch the
actual new ErrorRecord; create C/stop.json once with action/ordinal/type/message/
id/position/stack and remaining=FROZEN, preserve/rethrow the original. If stop
recording also fails, preserve tool/raw error and remain UNKNOWN; no fallback
success. Existing P grant intents/finals/raw paths and absence checks unchanged.

| Ordinal | Action / Grant | Capture hosts / targets | Exact result |
| --- | --- | --- | --- |
|01|RunOne writer-red|1 / 0|4FAIL/9PASS, exit1; seed SHA|
|02|RunOne contract-red|1 / 0|16FAIL/10PASS, exit1; seed SHA|
|03|Prepare writer-green|0 / 0|only writer changed;28snapshots; seed previous SHA|
|04|RunOne writer-green|1 / 0|13PASS, exit0; writer-green SHA|
|05|Prepare repair-green|0 / 0|only validator changed;28snapshots; writer-green previous SHA|
|06|RunOne contract-green|1 / 0|26PASS, exit0; repair-green SHA|
|07|RunOne core|1 / 0|25PASS, exit0;22private observations|
|08|RunOne lifecycle|1 / 3|17PASS, exit0;8private observations|
|09|RunOne continuation|1 / 4|16PASS plus12seamPASS, exit0;15private observations|

Both REDs require intended complete fixed failures, healthy publisher/evidence/
journal,3recorderPASS, empty stderr, no unexpected error and exact owned exit/drain.
An arbitrary crash, compilation or recording failure is not an expected RED.
Two writer hosts compile fixture once each; no other compile or process probe.
Runtime snapshots: existing28 + new28 + new28 =84, not a new seed preparation.
Use actual SHA returned by each successful Prepare, never predicted hashes.

## Task1 — explicit release, controller and immutable preflight

**Files:** create L/continuation-driver.ps1; C/progress.md, approval.json,
control.binding.json, approved-plan.snapshot.md, control.source.snapshot.ps1,
static.tool-result.json. Existing28sources untouched.
**Interfaces:** consumes actual approved plan/seed; produces retained control
binding SHA used by every later action, never a general execution capability.

- [ ] Obtain explicit approval of this complete plan, new controller and release of unused7/7/2/conditional1. Verify existing linked worktree/clean branch/origin and exact seed/history before creating anything. C/controller/portable-continuation outputs must not already exist. Record that this is NOT release of old v2 allowance.
- [ ] Before implementing control, enumerate static assertions for exact allowed derivation regions, nine-row order, only two Prepare grants, zero wildcard path/override flags, retained old-stop hash, separate new stop, CreateNew before body, immutable binding and unchanged process/evidence predicates. These assertions inspect source/AST only; no extra dynamic host is authorized. Mark lack of dynamic control-negative proof for final review.
- [ ] Implement only the fixed controller above; AST/LF/JSON and exact full-body comparison against frozen driver must pass. All28original hashes remain seed-exact. Check approval/binding schema/types/field set; preserve actual static tool chunks only after terminal completion, outside P. Freeze controller/approval/plan snapshots and binding once. Any actual preflight/static error stops with no implementation fix pass or launch.
- [ ] After an action invocation yields a session, wait only that session to terminal completion and retain every original tool object in C/actions/NN/tool-result.json afterward. Never write a pending initial result into P. Missing terminal/record errors freeze all later actions. Apply this rule to every later task.

## Task2 — intended REDs, then I1 both production writers

**Files:** only scripts/offline-diagnostics-v3/private/write-boundary.ps1 after both REDs; actual P outputs/C action records.
**Interfaces:** Write-ODCaptureRecord(Path,Bytes), Write-ODProfileIntentRecord(Path,Bytes), internal default New-ODRecordStream unchanged signatures. Same source used by trusted final publisher and profile intent.

- [ ] Run ordinal01 and02 once each with actual seed SHA and retained control binding. Require exact table output and ten writer/23schema observations, all recorder guards and journals. Inspect original error reference/order/count facts and exact rejection signatures, not only totals. Neither real defect is repaired before these intended REDs.
- [ ] Repair only the writer file: preserve first actual ErrorRecord through Write/Flush/Dispose, defer cleanup secondaries until first registration, preserve existing HostPrimary and reference-deduplicate bounded8secondaries, then rethrow original first. No finally replacement/reopen/retry/second Dispose or injectable production factory.
- [ ] Ordinal03 Prepare requires exactly this one runtime source changed, other2481pins identical;28new snapshots and actual previous seed SHA. Ordinal04 must return13PASS/exit0 with healthy real default publisher/probe/final and all ten cases. Preserve checkpoint; do not infer I2 acceptance.

## Task3 — I2 complete validator and actual parent consumption

**Files:** only L/observation-contract.ps1; existing oracle/tests/controller/capture/publisher stay fixed.
**Interfaces:** Assert-ODProfileObservations([object[]]Rows,[string]Profile,[Collections.IDictionary]Context), success no output; mismatch exactly RuntimeException with Message/FQID `OD: observation contract mismatch`.

- [ ] Read saved ordinal02 once, no rerun. Preserve3historical positive frames +20single mutations, including4Step controls; expect16unrejected non-Step mutations rather than unrelated errors. Use original literal oracle/models, never learn expected values from actual GREEN output.
- [ ] Implement complete ordered recursive contracts: core22/owned8/live15 observations; all View/Sinks/IORoot/Result/errors/summary keys, values, types and unknown/missing/duplicate/order rejection. Preserve explicit permitted path/PID/time/source-location variability only. All20single mutations require exact mismatch signature, not arbitrary exceptions. Same function remains called by parent for every actual profile.
- [ ] Ordinal05 requires exactly1validator change/28snapshots, other2481pins identical, previous writer-green SHA. Ordinal06 must produce26PASS/exit0, all23complete cases, healthy records and3recorder groups. No schema-unit result substitutes for actual profiles.

## Task4 — three actual public profiles, unchanged lifecycle limits

**Files:** no source changes; actual P/C records only.
**Interfaces:** consumes repair-green SHA and fixed v3 entry; actual Session/Ack only in each live host.

- [ ] Ordinal07 core once25PASS/0targets, default+explicit Inspect import/IO-free,14request rejections, complete22observations through repaired validator and all recorder/source bindings.
- [ ] Ordinal08 lifecycle once17PASS/3targets, nonassociation/prestart0, receipt-failure one-attempt, exact owned resource exit/drain and normal/timeout tuples; complete8observations and3recorder groups.
- [ ] Ordinal09 continuation once16PASS+12seamPASS/4targets, Active valid-step nonzero-budget missing/wrong Ack rejection, genuine Ack advances once, six invalid/replay requests, real output fault without contaminating B, complete15observations and actual A1seam output/trace. No extra authority/FINAL host.
- [ ] Preserve original limits: noncontinuation capture120000ms, continuation150000ms; sequence140000ms/cleanup4000ms; exact owned exit2000ms and shared drain2000ms; hold45000ms, normal/receipt2000ms, timeout200ms, seam110000ms. Engine SHA362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139; SMA2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B; original pinned path/identity DESKTOP-H4SVCLU\\Lenovo/adminFalse. Session/sink≤32, Ack≤2, observations≤32, distinct secondaries≤8 unchanged. No timing expansion on failure.

## Task5 — sole conditional review and portable handoff

**Files:** C/review.md and all actual records; create L/portable-continuation.json;
new docs/roadmap/2026-11-01-offline-i1-i2-continuation.md; update current-status,
repository-recovery and offline-diagnostics-v3 safety/operator notice as justified.
**Interfaces:** consumes seven exact actual hosts/7targets/2compilations,
2482runtime pins+control binding/84runtime snapshots and complete original records.

- [ ] Only if all nine action terminals and seven actual grants are exact, preserve source/results checkpoint then use the one released fresh-context readonly reviewer. Select most capable available approved model with explicit metadata. Review complete entry→profile→writers→validator→publisher/case-module→capture→NEW controller, original I1/I2 and actual records/control exceptions, not only two diffs. No execution/import/network/write/Git/subagents by reviewer. Capture complete findings/declined-to-judge list. No second reviewer, automatic fixes or retests.
- [ ] Important/Critical/incomplete review freezes handoff as unaccepted; even clean review supports only proposed narrow new-v3 disposition. Dynamic OS late write/real FileStream.Dispose, host loss/concurrent outlet failure/classifier exceptions and global crash-safe authorization remain unproved unless actually evidenced; controlled fixture success is not those events.
- [ ] On success OR any earlier stop, preserve every actual P/C file/directory/partial/zero-byte stream, source/control snapshots, consent and all original tool chunks in new portable-continuation.json with explicit roots/base64/length/hash. Verify decoded bytes against originals. Never replace earlier archives or fabricate missing outcomes. A stopping failure goes directly to this mechanical handoff, not task repair/retry.
- [ ] Record actual consumed versus frozen remainder and precise narrow claims. Explicit necessary source/controller/doc/archive list only; verify staged bytes, commit/push existing dev branch and remote SHA=HEAD. No main merge/force/deletion/cache/credential/raw-directory commit. Fresh computer restores progress, not execution permission/live authority.

## Self-review and approval gate

Source ownership/control exception→tasks1/5; first-cause ten vectors and real
publisher→task2; full schema23vectors and actual22/8/15consumption→tasks3/4;
owned lifecycle/Ack/exits→task4; failure/unknown/portable boundaries→all tasks.
Counts:7hosts,3+4=7targets,2fixture compilations,conditional1review; only2remaining
Prepare phases/56new runtime snapshots,84total. Frozen driver/seed untouched;
control identities separately bound and explicitly not counted as2482runtime pins.
No dynamic control-negative host is included; this limitation must reach reviewer.

Please review and explicitly approve Native execution, the new fixed continuation
controller/retained-stop exception, and release ONLY unused v3 7/7/2/conditional1
remainder under this exact sequence. Planning consent does not release anything.
Any unexpected actual failure stops all remaining work and preserves an unfinished
checkpoint; no follow-on workaround or rerun is authorized.
