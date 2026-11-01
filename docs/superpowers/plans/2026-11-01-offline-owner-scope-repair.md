# Offline Owner Scope Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after explicit design/plan/budget approval. Track checkbox steps; no implementation agents or extra review.

**Goal:** Repair private fixture function lifetime and permit exactly one additional, separately recorded writer attempt.

**Architecture:** Retain the four complete actual function definitions in private module script scope. Bind five narrowly changed fixture/controller consumers to a new immutable source epoch and fixed append-only writer slot; preserve every old attempt and stop.

**Tech Stack:** Existing Windows PowerShell7.6.5 Core x64, .NET AST/hash/JSON APIs, current bounded process controller. No installation.

**Spec:** [Proposed scope design](../specs/2026-11-01-offline-owner-scope-repair-design.md), [original implementation plan](2026-11-01-offline-i1-i2-single-owner.md), [controller amendment](2026-11-01-offline-controller-recovery.md). This amendment supersedes only the writer attempt binding, source window and explicitly added budget; original assertions/limits remain.

Status: DESIGN_PLAN_AND_BUDGET_PENDING_APPROVAL. No execution authorized yet.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; branch codex/c12-b01-task9-coordinator; verified origin https://github.com/win202569/Talenro.git. PowerShell only; apply_patch edits; explicit necessary-file checkpoints, no main merge/deletion/history rewrite.
- L/P/Q and exact five-source window are defined in the spec. Other26runtime sources unchanged. Auxiliary owner-scope.tests.ps1 remains outside the31runtime inventory and is never imported by capture/profile.
- Keep2215existing pins,93snapshots, all three stops, old manifests, REDs, auxiliary results, writer partial outputs/zero-byte files and portable-controller-writer-stopped.json unchanged. New epoch may replace only the five approved source hashes; new immutable provenance pins are separately inventoried, not silently counted as old pins.
- Parent freeze SHA D8D31B8BE9F38A35F12CF5951DC28B3723D759B96AF3BC3B3D8B3D4D95FAC57C; active historical stop SHA ECC2C74E5544CA0EC9A7D2C5A4377DA35F1A352F26CFEADBFEF35124C2A41607. Old3freeze/3stop identities and actual stopped artifacts must all validate.
- Additional budget only:2scope computations/1writer host/1fixture compile/31snapshots. At most2private modules per scope computation; no profile body, record I/O, child, Add-Type or target in them. Total snapshot ceiling155; total runtime ceiling6hosts/7targets/2compiles/conditional1review, including the consumed failed writer1host/1compile. Four later hosts/7targets are not spare retries.
- Fixed existing engine and SMA hashes, fivefalse capability flags, reference identity, source-error model, original oracle/output bytes,32recorder stages/12case inventory and all resource/time/process limits remain unchanged. No Force/IgnoreStop, arbitrary roots/callbacks, PID lookup, kill-tree, Job/watchdog, install or security change.
- First unexpected result stops; no automatic source correction, second GREEN/RED, repeat Prepare, retry host or extra review. Preserve actual chunks, not inferred terminal success. Review remains conditional on all actual stages passing; old failed host remains in history.

## Review Focus

1. All four functions, including inline parameters, persist across module invocations without body substitutions (FourFunctionsPersist/DefinitionsExact).
2. Script scope remains private and isolated; no export/global leakage or cross-module state (ModuleIsolation).
3. A stopped writer result is never accepted as the new writer result (AttemptBindings/ConsumedAttempt).
4. Historical GREEN describes old bytes, not edited current sources (HistoricalEpoch).
5. Recovery cannot bypass stops or missing approval, and every consumer agrees on the new epoch (RecoveryGate/NewEpochBindings).

## Task1 — freeze and reproduce the actual scope defect once

**Files:** create L/owner-scope.tests.ps1; append-only P/owner-scope.approval.json,
P/owner-scope.input.json, P/owner-scope-red.intent.json/records; Q/owner-scope-progress.md.

**Interfaces:** auxiliary function Invoke-ODOwnerScopeRegression -Phase Red|Green.
Fixed paths only. Returns a single phase receipt with Schema,Phase,TestSHA256,
InputSHA256,SourceSHA256,Groups,PASS,FAIL,NotRun,ExitCode,RootError. RootError stores
actual type/ErrorId/message/position/stack, not a fabricated exception. SourceSHA256
maps the five exact source paths; values must be equal before/after each phase.

- [ ] Validate baseline Git identity and all2215pins/93snapshots, old3freezes/3stops and actual stopped writer files. Confirm new slot/intent/stop/manifest/snapshot root absent. Record explicit approval with actual design/plan hashes, exact incremental budget and historical identities, CreateNew only.
- [ ] Author and commit auxiliary before RED; freeze test hash and five baseline source hashes in input. Freeze nine ordered groups: FourFunctionsPersist, DefinitionsExact, ModuleIsolation, AttemptBindings, ConsumedAttempt, HistoricalEpoch, RecoveryGate, NewEpochBindings, OriginalContractsUnchanged.
- [ ] Extract the actual owner initializer CommandAst uniquely: command variable controls, scriptblock with ordered parameters Definitions,ProfileText,Root,Repo,Packet,Leaf,Publisher,FailurePublisher. Reject0/multiple matches; do not confuse later control-loop/clear invocations. Extract four full function AST texts from actual capture/profiles sources, not helper copies. Use actual New-ODCaseModule initialization with actual collector/writer/owner definitions; no fault-stream compile or writers. Pass fixed scope-only roots and null publishers which must never be invoked.
- [ ] Consume RED intent before computation. Run `. L/owner-scope.tests.ps1; Invoke-ODOwnerScopeRegression -Phase Red` once in existing tool context. Execute initializer, let it return, then in a separate invocation of the same module resolve Invoke-ODProfile with Get-Command -CommandType Function -ErrorAction Stop. Require actual CommandNotFoundException/ErrorId CommandNotFoundException:1FAIL/8NotRun/exit1. Any other failure or unexpected pass stops. No profile body call.
- [ ] Save complete actual tool command/chunks and phase final separately; retain old availability/decoder REDs and all failed GREENs without rerunning. Static parsing/inspection does not grant exploratory module executions.

## Task2 — minimal scope and complete fixed-slot provenance repair

**Files:** modify only five existing sources in spec; frozen auxiliary unchanged.
Create P/owner-scope-green.intent.json/records; later P/writer-scope-repair.intent.json,
P/writer-scope-repair.frozen.json and P/snapshots/writer-scope-rebind/<31sources>.
New failures use P/writer-scope.stop.json, never overwrite the three old stops.

**Interfaces:** retain canonical RunOne grant writer-green. Private helper
Get-ODWriterManifestPath() returns fixed writer-scope-repair.frozen.json.
Add private Get-ODAttemptSlot([string]$Grant): writer-green -> writer-scope-green;
the four other existing grants -> unchanged name; all other values reject.
No public attempt/path choice. Receipts keep canonical Grant and add AttemptSlot;
require both identities, actual freeze and the corresponding fixed file paths.
Prepare adds only fixed grant writer-scope-rebind; old preparation paths stay spent.

- [ ] In record-boundary.tests.ps1 qualify only each unique imported declaration with script: for the four allowlisted functions. Preserve full parameter/body text and line breaks. Bind its ten-case unit root and owner root to fixed unit/writer-scope-green and its profile manifest to the new epoch. No case/body/oracle edit.
- [ ] In capture.ps1 map only writer capture root and evidence filename to writer-scope-green; retain switch/catch/collector/publisher logic. In source-model.ps1 update only writer manifest and the two fixed intent unit/error-path bindings. In observation-contract.tests.ps1 update only writer evidence/final input filenames. No weakened observations/source locations.
- [ ] In driver.ps1 use the same fixed slot for intent/final, streams, evidence, capture/probe/journal/unit roots and prior-final reads. Validate canonicalGrant+AttemptSlot+source epoch. Existing old writer outputs are required historical inputs with exact hashes, not deleted/treated as the new result; all new slot paths must be absent before the new attempt. Preserve every per-grant consumed-intent guard. Other four slots and original output oracles unchanged.
- [ ] Bind old auxiliary GREEN's three source hashes to P/snapshots/controller-rebind, not newly edited disk. Keep its test/input/final exact. Validate new scope RED against parent snapshots and new GREEN against all five current sources. No regrading historical receipts. Add strict fixed recovery checks for original3stops, consumed writer final/streams/evidence/tool record, explicit new approval, new RED/GREEN and parent freeze. A new scope stop always blocks. Missing/mutated identity rejects; no fallback to old writer manifest.
- [ ] Consume sole GREEN intent; run same frozen auxiliary -Phase Green once. Require9PASS/0FAIL/0NotRun/exit0, null RootError and exact current five-source hashes. FourFunctionsPersist resolves all four after initializer and again after another invocation. DefinitionsExact compares complete parameters/bodies, allowing only declaration scope qualifier; ModuleIsolation checks two private modules, zero exports and no outer function changes without invoking profile/record I/O.
- [ ] Remaining GREEN groups statically/through extracted pure gates check: every attempt path consumer and prior final; old-attempt rejection; historical snapshots versus current-source hashes; exact/missing approval, changed old stop, wrong parent and present new stop; manifest consumer agreement; original31inventory/oracles/cases/bounds unchanged. These cases must be literal independent expectations fixed before RED, not expected values read from changed implementation. Any needed test correction stops, not refreeze/retry.
- [ ] Create recovery intent before snapshots; Prepare writer-scope-rebind once with actual parent SHA. New manifest Phase writer-green, same31SourcePaths and inherited original Plan/Design/limits; exactly five permitted changed source pins. Add exactly seven immutable provenance pins: parent freeze, active historical stop, new design, this plan, new approval, auxiliary test and frozen auxiliary input. Existing2215+7=2222unique pins; earlier provenance pins stay exact. No mutable ledger/current success output as a source pin. New OwnerScopeRecovery map binds old identities/stopped artifact hashes, approval/design/plan/test/input and RED/GREEN final hashes; no circular self-hash. Verify against actual records at Prepare and RunOne.
- [ ] Create31snapshots/new manifest once with CreateNew, verify exact five changed inputs/new2222pins and all old bytes. Snapshot total124 now; reserve remaining31 for original contract. Save actual Prepare tool record. Commit/push explicit source/auxiliary/docs and new portable records before launching host; no post-GREEN source edits.

## Task3 — one new writer, original remaining stages, truthful handoff

**Consumes:** new epoch, exact9-group GREEN and complete recovery chain.
**Produces:** one separately recorded writer result, then original stages or stop.

- [ ] Run driver RunOne writer-green once with actual new frozen SHA. Slot is writer-scope-green. Consume new intent before host. Allow1host/1compile/0targets. Require original26PASS/exit0 AND exact streams, OwnedExit/StreamsComplete/PostPinsIntact, complete32recorder stages/12cases, source models/live references/healthy publisher probe and parent EvidenceValid/Exact. Not18partial PASS or INCOMPLETE publication. Save all actual tool chunks. Do not rerun on any mismatch.
- [ ] Only if exact, resume original Task3: modify only observation-contract.ps1, other30sources fixed at scope epoch; Prepare repair-green once with new writer parent,31new snapshots, then contract-green once52PASS/exit0/0targets/0compile. Contract reads new writer slot. Snapshot total155; old writer stays preserved.
- [ ] Only if preceding stage exact, core once25PASS/22rows/0targets, lifecycle once17PASS/8rows/3targets, continuation once16PASS+12seamPASS/15rows/4targets. Keep original120000/150000capture,140000sequence,4000cleanup,2000owned-exit/shared-drain,45000hold,200timeout,110000seam and all original bounds. These four hosts cannot fund writer retries.
- [ ] Only after all five successful current-chain actual finals, use original single conditional readonly runtime review. Include the failed historical host, both scope phases, five-source diff and complete old/new provenance. No additional review or automatic fix/retest on Important/Critical/incomplete findings.
- [ ] Success or stop: create NEW L/portable-owner-scope-repair.json containing only actual P/Q files/directories, base64/length/SHA, including partial/empty artifacts; decode-verify against actual bytes. Never synthesize missing records or overwrite older portable archives. Update current-status/execution/recovery, actual consumed/remaining budget and NOT_ACCEPTED unless warranted. Explicit-file commit/push with staged raw bytes==disk, remote SHA==HEAD and clean worktree. No main merge or historical I1/I2 closure inferred.

## Planning self-review

Scope/lifetime, parameter retention, privacy, all writer consumers, historical
source epochs and finite resource additions map to the tasks above. The original
assertions and later pure validator task are preserved. This proposal deliberately
requires approval before auxiliary authoring, source changes, module execution,
Prepare, host, compilation or later validation. Execution method remains Native.
