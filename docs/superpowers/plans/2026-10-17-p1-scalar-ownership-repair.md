# P1 Scalar Ownership Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for the proposed Native execution. Steps use checkbox (`- [ ]`) syntax. This is a proposal, NOT execution authority.

**Goal:** Close Important I1 by refusing decorated primitive scalars before evaluation, preserving ordinary P1 decisions and independently owned results.

**Architecture:** Tighten the existing passive-input boundary, not the state machine. A separate portable regression suite supplies independently expected refusals, sequential-isolation checks and plain controls; the existing35-case suite remains byte-for-byte unchanged and is invoked only in newly approved combined slots.

**Tech Stack:** Windows PowerShell5.1/PowerShell7, existing runtimes, descriptor inspection and literal plain objects; no new dependency or installation.

**Spec:** [Approved joint design §9/10](../specs/2026-10-17-raw-capture-supervision-design.md), [P1 exact contract](2026-10-17-offline-supervision-policy.md), and [Important I1/dispositions](../../roadmap/2026-10-17-offline-supervision-policy-readonly-review.md).

Status: **NATIVE APPROVED, THEN STOPPED ON SLOT1 FIXTURE FAILURE.** User yes after ad0b565250a2aacb35d47526368edbbabab976c9 approved this plan, seven hosts and one conditional scoped review. Slot1 consumed: exit1/927ms, fixture-fidelity error, not the required RED; ExpectedResult=false. Slots2–7 frozen, reviewer0/1 not eligible. No source edit or automatic repair/retest. See [stopped record](../../roadmap/2026-10-17-p1-scalar-ownership-repair-execution.md). Earlier proposal wording below grants no renewal; old8/8 and1/1 remain spent.

## Decision and alternatives

Recommend strict refusal of extended primitive metadata. The contract already admits exact primitive scalar values and promises independent snapshots; silently carrying decorations is the gap. This narrows formerly admitted decorated inputs explicitly: callers must supply plain values. Ordinary schema, reason tokens, fields and state decisions stay unchanged.

Alternative: reconstruct every admitted scalar from validated primitives, stripping metadata. This retains broader input compatibility, but introduces normalization paths across every scalar field and more cross-runtime ownership proof. Deep-copying arbitrary metadata is rejected: it expands the input language and could traverse executable/cyclic graphs. No dynamic manifestation is claimed yet; the first two proposed slots test an independent refusal expectation against the current implementation.

## Global Constraints

- Reuse D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator and codex/c12-b01-task9-coordinator; verify AGENTS, clean state and origin https://github.com/win202569/Talenro.git before execution. Planning base d2021dc6; the plan commit becomes the execution base.
- Product API remains Get-OfflineSupervisionDecision([object]$Scenario), exact13 root/22 Decision fields. Mode=policy-only, EvidenceClass=synthetic, LiveActionsPerformed/OriginVerified/Acceptance/RuntimeApproved=false, actual CleanupStatus=unknown.
- No Process/Job/pipe/Stream/real clock/timer/file/network/security operation, callback, import or new dependency in product. No P2/P3, OS/runtime/security/main/C12 approval or historical payload execution.
- Preserve120000ms absolute budget,113500ms cutoff and500/5000/1000ms phase bounds; retain equality expiry, sticky first failure and all existing capacities/reason precedence for plain inputs.
- Change only offline-supervision-policy.ps1 plus a NEW regression test and documentation. Existing offline-supervision-policy.tests.ps1 stays61A3BACF26EDE1F8B137CBF77CC4DBC35E4D6E6FB8FE6365953CDA9F41CEDEA3. Baseline source is FE23DF3BE1F5BA5DF4C2B565DEE5CFB46DAB6F57C896DA9B5C95DFE440FF0ABB.
- Protected15 = prior protected14 plus the unchanged35-case suite. Pin/check before and after each new child; do not import/rerun the prior raw coordinator or dependency suites. Existing source changes only after both precise REDs.
- Preserve every prior/new raw packet and original verdict. No retrospective correction, deletion, old-budget reuse, exploratory imports or automatic repairs/retries. Unexpected result stops; restore the specified temporary mutant even on failure.
- Windows PowerShell only. Test hosts inherit the ordinary token; no elevation/relaunch to get a pass. Git checkpoint permission is separate. No protection/trust/ownership changes, tool installs, whole-project tests or whole historical-branch audit.

## Review Focus

1. Int32/Int64 nested NoteProperty aliases must refuse without evaluating metadata: DecoratedInteger.
2. Boolean/string paths, including NaturalCompletion/Classification, must not bypass the boundary: DecoratedBoolean/DecoratedString.
3. Root labels and every observation scalar must be checked before conversion, branching or reduction: RootScalars/ObservationScalars.
4. Getter/method/alias metadata must not execute during refusal; plain CLR string Length must remain admissible: PassiveDescriptors/PlainControls.
5. Two calls with the same decorated input and later nested mutation must not affect prior results or a fresh normal call: SequentialIsolation.

## Proposed new allowance — seven sequential hosts, one conditional reviewer

No allowance is consumed until user approval of this written plan and execution method. Each child has a120000ms harness ceiling including its test work. No parallel hosts, retries or unlisted imports. Ordinary process scheduling is not a hard wall-clock guarantee. Harness waits/drains must remain bounded; timeout or unproved termination is a stopped incomplete result, never permission for arbitrary PID cleanup.

New test: docs/diagnostics/isolated-fuzz/offline-supervision-policy.ownership.tests.ps1, parameter `[ValidateSet('regression','all')][string]$Group='regression'`.
Regression ordered labels (10): P1I1ApiRequired, P1I1DecoratedInteger, P1I1DecoratedBoolean, P1I1DecoratedString, P1I1RootScalars, P1I1ObservationScalars, P1I1PassiveDescriptors, P1I1SequentialIsolation, P1I1PlainControls, P1I1Complete.
Group all first invokes the unchanged old test with `& <old-test> -Group all` in the SAME child, checks success, then runs these10 cases. No nested test host. Thus a combined success contains the original ordered35 labels followed by the ordered10, exactly45; context/host/limit lines have their own frozen inventory. Sentinels are not coverage percentages.

|Slot|Engine/group/version|Precise expected outcome|
|---|---|---|
|1|PS7 regression, original source|exit1; only API PASS, exact ASSERTION FAILED: P1I1DecoratedInteger|
|2|PS5.1 regression, same original source|same precise RED|
|3|PS7 all, fixed source|exit0, exact45 ordered PASS labels|
|4|PS5.1 all, same fixed source|same GREEN|
|5|PS7 regression, specified mutant|exit1, only API PASS, exact P1I1DecoratedInteger assertion|
|6|PS7 all, restored fixed source|exit0, exact45 ordered labels|
|7|PS5.1 all, same restored source|same final GREEN|

Mutation: remove ONLY the newly added primitive-metadata refusal guard, leaving existing passive checks, test bytes and all other production behavior unchanged. Save exact diff, original/fixed/mutant/restored hashes and retained mutant-source bytes in the NEW ignored packet; restore exact fixed bytes immediately afterward. The mutation must expose the same refusal gap, not alter expected output or disable a test.

Reuse the old controller's bounded ownership/stream principles, but create a separate packet `.superpowers/sdd/2026-10-17-p1-scalar-ownership-repair/` with a new manifest/index/controller; no sibling packet edit. Fixed engines: C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe and C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe. Use -NoProfile -NonInteractive -File <new-test> -Group <group>, PS5 inbox PSModulePath. Missing runtime/permission/fixture fidelity stops, no fallback/install/elevation.

Freeze the complete new suite, expected inventories and protected15 before slot1; fixed source before slot3. Preconsume slots, save separate stdout/stderr, host/context/admin/version, exit/elapsed/timeout, every source/test/output hash. RED is the exact assertion plus exit1, not just nonzero. Empty stderr or exclusively structurally recognized CLIXML Completed-progress is allowed under the existing strict DTD/external-resolution-disabled/no-deserialization policy; error/unknown/trailing payload stops. No suppression or retrospective relaxation. Receipt checks/static parsing/hashing do not execute product and are allowed; dynamic probes all belong to the seven slots.

## Task1: Refusal boundary and portable regression evidence

**Files:** Modify docs/diagnostics/isolated-fuzz/offline-supervision-policy.ps1 (passive validation only); create docs/diagnostics/isolated-fuzz/offline-supervision-policy.ownership.tests.ps1 and docs/roadmap/2026-10-17-p1-scalar-ownership-repair-execution.md. Update current-status, repository-recovery and this plan. Existing tests/dependencies remain unchanged.

**Interfaces:** Consumes the unchanged exact Scenario/observation/result contract. Produces conservative structural refusals before reduction; no new public API. New test helpers use P1I1- prefix, own literal13-field refusal and complete normal13+22 expectations, fixture constructors and recursive exact comparator. Never derive expected decisions from product helpers. Keep all new test helpers separate from old P1 helpers.

- [ ] **Step1 — write/freeze the entire new test first.** Construct fresh PSObject wrappers around each intended primitive and attach a NoteProperty named Metadata containing a fresh mutable object[]; assign the wrapper directly into a plain fixture field without typed helper coercion. Before each product call, independently assert the field still has the exact BaseObject type/value and its descriptor/value reference matches the intended metadata. Fixture loss/coercion is an unexpected infrastructure failure, not the expected RED. Do not mutate shared literal/type-data state. New case helper emits one PASS after all assertions, catch emits exact named assertion/exit1; unrelated errors never become that assertion.
- [ ] **Step2 — pin complete independent expectations.** Every decorated root scalar expects the literal13-field Refused result with Evaluated=false, Reason=invalid-scenario, index=null, Decision=null and actual flags false/cleanup unknown. Every decorated observation scalar expects the same descriptor with Reason=invalid-observation and its exact observation index. No partial Decision or retained input wrapper. Integer first vector: normal five-observation fixture with only progress ExitCode at index2 decorated Int32(0); full literal invalid-observation/index2 expectation. Also cover Int64, true/false booleans, clean/refused/unknown strings, roots Schema/Mode/SessionLabel/ClockLabel/StartMs, all common observation fields and every scalar payload field in arm/release/progress/control/fault/cleanup/persist. Use independent valid fixture per field, decorating exactly one scalar; data-driven rows remain within their named case, not extra hosts.
- [ ] **Step3 — pin nonexecution, full-preflight and ownership cases.** PassiveDescriptors decorates primitive values with ScriptProperty, AliasProperty and ScriptMethod separately; counters for getters/methods must remain0. Inspect descriptor names/types only, never invoke/read decorated values. RootScalars covers each root scalar; ObservationScalars includes invalid scalar in a later structurally valid observation following a modeled timeout, requiring full-preflight refusal with that later index. SequentialIsolation calls twice on the SAME decorated fixture, checks both full refusals, mutates original nested Metadata contents, verifies both earlier refusals unchanged and distinct root objects, then mutates one returned result and confirms the other and a fresh plain normal decision unchanged. Repeat for integer/bool/string. PlainControls covers undecorated exact integer/long/bool/string/null-exit, string Length, normal complete, empty Prepared and sticky failure with complete literal expectations. No claim of concurrency or hostile-runspace coverage.
- [ ] **Step4 — run slots1/2 before any source change.** Use regression group; both must fail at the exact decorated-integer refusal expectation, not fixture sanity, API absence or unrelated failure. This reproduces missing structural refusal, not necessarily the full historical alias manifestation on both runtimes. Any mismatch stops; do not revise expectations after seeing output.
- [ ] **Step5 — add the minimal guard to Assert-SupervisionPolicyPassive($Value).** For exact BaseObject types System.Int32/System.Int64/System.Boolean/System.String, inspect PSObject.Members descriptors without reading member values; allow only intrinsic CLR Property/Method/ParameterizedProperty member kinds, reject extended NoteProperty/AliasProperty/ScriptProperty/CodeProperty/ScriptMethod/CodeMethod/MemberSet/PropertySet and any other kind. Preserve existing executable-property/shadow-bound checks for nonscalars and arrays. Guard must run before scalar casts, token comparisons, Kind dispatch and observation copy, via existing root/shape/scalar checks. No recursive metadata traversal, stripping, deep copying or state-machine/output change. Trace root and every observation scalar entry point; plain object container NoteProperties remain valid. Any inability to distinguish intrinsic/extended descriptors conservatively on the two runtimes stops rather than widening the allowlist ad hoc.
- [ ] **Step6 — run fixed slots3/4.** Exact45 labels, protected15/inventories/streams all match. An unexpected result ends this allowance without a repair loop.
- [ ] **Step7 — run slot5's one mutation and restore.** Capture mutant bytes/diff/hash, require exact failure, restore frozen fixed source even if failed unexpectedly. Do not change tests or old evidence.
- [ ] **Step8 — run final slots6/7 and reconcile.** Hash every new stream/index/manifest, verify fixed restoration/protected15 and all four combined inventories. Document actual results and limits. Review explicit necessary-file diff, parse/whitespace/local links and ordinary checkpoint on existing origin/branch; remote SHA must equal local with clean tracked state. No completion/acceptance claim before review disposition.

## Task2: One scoped review and stopped-or-ready handoff

**Files:** New execution report and necessary status/recovery/plan/spec links only. No automatic product repair.

**Interfaces:** Consumes fixed source, unchanged old tests, new frozen regression suite and seven-slot evidence; produces finding/limit dispositions, not runtime permission.

- [ ] **Step1 — dispatch ONE fresh independent readonly reviewer only if all seven precise outcomes matched.** Inspect ONLY this repair diff plus relevant P1 contract/source/tests/receipts/protected pins and I1. Fresh context, available most-capable reviewer configuration; no import/test/reproduction/edit/security/network/trust change/delegation or historical whole-branch review. Review whether rejection actually covers each scalar before use, preserves plain values and avoids callbacks or references. Review allowance is one seat, not an automatic fix/re-review cycle.
- [ ] **Step2 — disposition every finding and every declined judgment.** Important/Critical stops unfinished, no automatic repair. Minor explicitly retained/deferred. Parent static/Git checks are not independent review. Retain all18 prior limits: saved evidence not authenticated execution; synthetic timing/actor metadata not real clocks/authentication/containment/capture/cancellation/cleanup/persistence/scheduler guarantees; timeout/CLIXML branches may be unexercised; no catastrophic/concurrent/hostile-runspace/whole-project/Stage B/security/P2/P3/main/C12/release approval.
- [ ] **Step3 — save/push necessary final files even on stop.** Preserve old/new raw locally, exclude raw/controller/cache/binaries/secrets from Git; update exact verified/unverified facts, remaining decision and new-budget accounting. Verify branch/origin, explicit stage list, remote=local and clean state. A clean review may support a separately stated limited P1 acceptance proposal; it does not automatically accept, merge or launch P2/P3.

## Self-review and approval handoff

Scope is only I1, not a new subsystem: existing written P1 ownership requirement supplies the design; this user-requested plan records the proposed stricter admission rule for explicit review. No separate design-writing or implementation is inferred. All five Review Focus classes map to named new tests; original35 expectations stay unchanged. New10 labels include sentinels, not ten independent guarantees. All commands/imports belong to a frozen slot; no hidden baseline/task-done/whole-repo rerun. Native is recommended for this coupled boundary/suite; Subagent-driven implementation would need a revised allowance before dispatch.

Selected after ad0b5652: **Native parent implementation, seven synthetic hosts and one conditional scoped readonly review**. Slot1 failed unexpectedly in fixture preparation; stopped without source repair/retest/review. New bounded direction is required before any continuation. Generic extra tests, whole-branch review, automatic fixes or scratch deletion remain excluded.
