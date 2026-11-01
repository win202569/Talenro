# Preparation Recording Adjustment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after explicit execution approval. No agents or reviewers. Steps use checkbox syntax for tracking.

**Goal:** Permit exactly one additional, preparation-only seed attempt while preserving the actual stopped checkpoint and keeping all runtime grants frozen.

**Architecture:** Add one fixed, approval-bound preparation exception to the existing v3 driver. Retain stop.json, bind the five original records and unchanged27sources, and use create-once intent/entry markers in a separate recording sidecar. Wait for terminal tool completion before saving its transcript; stop after preparation regardless of outcome.

**Tech Stack:** Windows PowerShell7.6.5/Core/x64, existing pinned engine, .NET CreateNew files, SHA256, AST-only parsing. No install, Bash, Docker, runtime imports, fixtures or test hosts.

**Spec:** [Approved preparation design](../specs/2026-11-01-offline-prepare-recording-adjustment-design.md), SHA256 6F57E7DAA71C5B3E6E193B8C46BD643B524D8ACFE2A945EF166FB1F4A06298E5. The user's next “是” after commit b4b661f992ae16a9b1cffc525cc4ae18111f15a7 approved that design and drafting this plan, NOT execution. Its original draft-status text is preserved as historical provenance.

Status: PLAN_PENDING_EXECUTION_APPROVAL. Planning baseline b4b661f992ae16a9b1cffc525cc4ae18111f15a7; same linked worktree and codex/c12-b01-task9-coordinator branch.

## Global Constraints

- Preserve original five packet records, portable-stopped.json and all old sources/records byte-for-byte. No deletion, move, replacement, new runtime namespace or reset of historical failure.
- Only leaf driver.ps1 may change; other27seed sources and I1/I2 remain unchanged. No changed test expectations, process checks, publisher, collector or production writer/validator.
- NEW scope requested: one additional Prepare(seed), zero capture hosts/targets/compilations/reviews. This is an explicit exception, not a refunded preparation or resumed runtime allowance.
- All NEW7hosts/7targets/2compilations/conditional1review and older frozen/exhausted allowances remain unchanged. No RunOne even after preparation success; no main merge or production integration.
- Any actual static, provenance, marker, preparation, recording or integrity failure stops the entire adjustment. No automatic repair/retry/refreeze, renamed attempt, unknown-to-PASS conversion or new allowance.
- Existing engine/SMA identity, static AST/LF/JSON, inherited2394pins, historical byte checks, exact2482Present/28snapshots and false capability flags remain mandatory.
- Session/Ack never serialized or restored. No runtime module import, Add-Type, target, PID/name rescan, kill tree, watchdog, readiness wait or full-suite/baseline run.
- Finite approved Native workflow supersedes skill-default test/setup/Bash/reviewer reruns for this preparation-only change. No dynamic gate tests are authorized; static-only assurance is an explicit limitation requiring later scoped acceptance.

## Review Focus

1. An already-consumed intent/entry or incomplete transcript must not permit a second attempt: fixed CreateNew markers and task3 handoff checks.
2. A same-name modified record, extra file/directory or substituted link is not the retained checkpoint: exact names/hashes/bytes and reject ReparsePoint on fixed paths in tasks1–2.
3. Preparation permission must not become RunOne or later-freeze permission: fixed Action/Grant/identity checks and original runtime stop guard retained, task2 AST trace only.
4. A yielded tool result is not completion: keep all objects, wait actual session only, persist once afterward, task3.
5. A revised driver must not be misrepresented by the old static receipt or old plan: separate original/revised hashes and approval attestation, task1–3; original snapshots remain original.

## Exact files and interfaces

Repo-relative runtime packet P remains `.superpowers/sdd/2026-10-31-offline-integrated-review-repair`.
Recording-only sidecar A is `.superpowers/sdd/2026-11-01-offline-prepare-recording-adjustment`; it is NOT a replacement runtime packet or namespace.
Driver D is `docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/driver.ps1`.

Create only in A, each once: `approval.json`, `original-driver.bin`, `static.tool-result.json`, `attempt.intent.json`, `attempt.entered.json`, `prepare.tool-result.json`, `integrity.tool-result.json`, `handoff.md`. Partial/absent outputs stay partial/absent; do not fabricate missing files. Export all actual A bytes/directories to ordinary `docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/portable-prepare-adjustment.json` at handoff; do not replace portable-stopped.json.

Modify only D during implementation; create execution report `docs/roadmap/2026-11-01-offline-prepare-recording-adjustment.md`; update current-status.md and repository-recovery.md. The original spec/plan/execution reports and old packet ledger remain unchanged. Necessary later documentation updates do not authorize code or runtime continuation.

Fixed adjustment identity: `offline-prepare-recording-adjustment-2026-11-01-1`.
Add optional driver parameter `[string]$PreparationAdjustment`; absent retains original behavior. Any nonempty value must equal that literal AND Action must be Prepare AND Grant must be seed, before any exception. Do not expose arbitrary marker/record/packet paths or callbacks.

`approval.json` schema1 contains Identity, DesignSHA256 (above), PlanSHA256 (actual hash of this approved file), OriginalDriverSHA256 (4F75FBDD23CD6A21C3567460A1CFC62661B5B00D8D5A77F8FE4300CFE1500EE9), ExecutionBaseline (actual clean approved-plan checkpoint), ApprovalText (actual user execution approval, not design consent), AdditionalPrepare=1, Hosts=0, Targets=0, Compilations=0, Reviews=0, RuntimeRemainder="FROZEN". Record approval verbatim, do not invent tool-return or signature metadata. Hash/keys/types are checked, not proof of platform enforcement or cryptographic user identity.

`attempt.intent.json` schema1 contains Identity, ApprovalSHA256, PlanSHA256, RevisedDriverSHA256, Action="Prepare", Grant="seed", AttemptConsumed=true, RuntimeRemainder="FROZEN". Launcher creates it durably via CreateNew before the single invocation. Its presence forbids the operator from invoking again, including if launch never starts. Driver consumes its fixed contents and requires absent entered marker, then creates `attempt.entered.json` via CreateNew with Identity, IntentSHA256 and AttemptConsumed=true BEFORE its static/preparation branch. Two concurrent entries cannot both obtain this marker. Never clear either marker. Existing marker or unavailable/ambiguous terminal state means no second invocation.

## Task 1 — approval, fixed provenance and static-only preflight

**Files:** A/approval.json, A/original-driver.bin; unchanged P and portable-stopped archive.
**Consumes:** explicit future execution approval, clean checkpoint, actual original receipts.
**Produces:** fixed approval hash and retained original driver bytes; no execution permission beyond this approved plan.

- [ ] Require explicit approval of this complete plan and its one preparation attempt. Verify branch/origin/clean baseline using read-only Git and require A and portable-prepare-adjustment output absent before creating anything. Record actual approval text in A/approval.json. No historical driver invocation.
- [ ] Verify D's original hash above, all28source hashes from the preserved static receipt, all2394inherited pins and historical archive checks. Copy D bytes exactly once to A/original-driver.bin, verify equal hash; do not use a decoded/reconstructed driver as historical original.
- [ ] Verify P has exactly five regular non-linked files and no directories. Use the exact five names/hashes in the approved design, plus portable-stopped SHA256 1D10ABA536D414B06BE67422EA3662EF68AF4926858A70EBBBCFAB38E1EE6ADF and decoded original-byte comparison. Verify seed.frozen.json/snapshots/writer-red.intent.json absent. A may now contain only the approval/original-driver files just created in this run; any prior or unexpected entry stops, with no reuse/new suffix.

## Task 2 — fixed preparation-only control adjustment

**Files:** D only; A/static.tool-result.json after checks terminate.
**Consumes:** task1 actual approval/provenance and explicit identity.
**Produces:** revised driver hash and full static receipt, not dynamic acceptance.

- [ ] Add only the parameter and fixed preparation exception described above. Validate exact approval keys/values, approved design/plan hashes and original-driver bytes. Pin the approved plan hash as a literal after approval; no self-referential planned hash, arbitrary caller hash or auto-refresh on mismatch. Compare remaining27source hashes against original static receipt; separately attest revised D rather than requiring old D hash for the revised file.
- [ ] Replace only the initial unconditional stop rejection with: if no adjustment, original rejection unchanged; if adjustment, exact Prepare/seed/identity/provenance/retained-five-file checks and intent validation are required, followed by atomic entered-marker creation. Keep the later RunOne stop rejection byte-identical. Ordinary requests, invalid identity/action/grant and later Prepare grants receive no exception.
- [ ] Seed inventory in this explicit adjustment branch permits exactly the five pinned original records, not any five files or wildcard receipt names. Nonadjustment branch retains original two-name rule. Do not bypass missing-manifest, snapshot absence, engine, static, inherited/archive, manifest/pin verification or false flags. Original seed manifest BaseCommit/PlanSHA256 remain original-plan provenance; approval sidecar explicitly supplies current execution/revised-driver identity. Keep schema2/2482pins/28snapshots and original spec/plan snapshot bytes unchanged.
- [ ] AST-parse revised D without executing/dot-sourcing it. Perform LF/JSON checks; statically compare original versus revised AST/text so differences are restricted to this parameter/control region. Trace every nonempty adjustment path to exact Prepare/seed and fixed records; confirm no import/target/fixture/process policy or RunOne path changed. Compare other27source hashes and all retained bytes again. Save actual complete tool objects to A/static.tool-result.json only after terminal completion. Any actual failure stops without fix passes or invoking Prepare. No negative runtime probes are authorized.

## Task 3 — single preparation, terminal recording and immediate stop

**Files:** A/attempt.intent.json, attempt.entered.json, prepare.tool-result.json; actual P preparation outputs only.
**Consumes:** successful task2 static/provenance checks and actual revised-driver hash.
**Produces:** one actual preparation result, possibly partial/unknown; runtime stays frozen.

- [ ] Recheck fixed original P inventory, absent seed manifest/snapshots and absent markers. Launcher writes A/attempt.intent.json once via CreateNew, flushes/closes it before launch. No launch retry if the next call fails or returns no associated session. Driver validates intent/approval/revised-D hashes and writes its entered marker once before the preparation branch. No initial tool receipt is written to P or A during the invocation.
- [ ] Invoke exactly once with the pinned existing PS7 engine, from this worktree:

  `& ./docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/driver.ps1 -Action Prepare -Grant seed -PreparationAdjustment offline-prepare-recording-adjustment-2026-11-01-1`

  This is the single additional preparation allowance, not a capture host. No RunOne afterward. Preserve initial tool object, then every actual wait object in order, waiting only its returned session ID. A tool-call yield is not permission to run another command. Keep each wait bounded to allow progress updates; do not terminate/restart just because it yields.
- [ ] After terminal completion only, persist the entire initial/wait object sequence once to A/prepare.tool-result.json. For unknown/lost completion preserve available objects as UNKNOWN and stop; do not invent a terminal result or persist concurrent partial/final variants. A recording error itself stops. Do not modify the original stop.json; it describes the first failure, not this attempt's error.
- [ ] If terminal exit0: require exact Prepared=seed, Pins=2482, Snapshots=28, ChangedInputs empty; verify actual manifest SHA, all2482Present entries, exact28snapshot paths/bytes, original spec/plan snapshots and all five original records. These are post-preparation read-only checks, not runtime imports/tests. Save complete actual integrity tool objects after completion. If nonzero or partial, preserve every actual output, error and snapshot without cleanup; no fix/retry. In every case stop after this task with runtime allowance FROZEN and I1/I2 unresolved.

## Task 4 — portable stopped handoff

**Files:** actual report, A/handoff.md, portable-prepare-adjustment.json, current-status/repository-recovery.
**Consumes:** actual partial/success/unknown result, never inferred test success.
**Produces:** recoverable ordinary development checkpoint.

- [ ] Record actual additional Prepare consumption and zero capture hosts/targets/compilations/reviews, original and revised driver hashes, approval/plan/design identities, exact outcome and missing/unproved facts. Keep original historical first cause UNKNOWN and known recording-order failure separate. No finding disposition, runtime acceptance or main integration claim.
- [ ] Export actual A files/directories, zero-byte/partial objects and original-driver.bin as portable base64; verify decoded byte lengths/SHA256/equality. Preserve original portable-stopped unchanged. Snapshot the actual new P outputs, if any, in this new portable archive under an explicit separate P mapping; include all five retained records unchanged. Do not imply the five-file stopped archive already contains later outputs. No invented snapshots or result files. No secrets/caches/binaries beyond the original PowerShell source bytes stored as a provenance artifact.
- [ ] Update current-status/recovery with outcome and a clear prohibition on RunOne or later freezes. Explicitly stage only D, new actual report/archive and necessary updated docs; verify staged byte identities, commit/push existing branch, confirm remote SHA equals HEAD. No main merge, force push, deletes or runtime continuation. New computer restores records, not permissions, live sessions, dependencies or review eligibility.

## Self-review and execution approval

Spec coverage: preservation/provenance→tasks1–2; one fixed exception without general override→task2; durable one-attempt and terminal-only recording→task3; no runtime resumption and recoverability→tasks3–4. ReviewFocus1–5 have explicit checks above. Static-only path review cannot prove dynamic gate behavior; this plan deliberately requests no extra test host and makes no such claim.

Please approve this Native plan explicitly: change only the specified driver preparation controls and perform at most ONE additional Prepare(seed), with 0capture hosts/0targets/0compilations/0reviews, preserve all original records, then stop even on success. The existing7/7/2/conditional1 runtime remainder stays FROZEN. Approval of the design or of plan drafting alone does not authorize this execution.
