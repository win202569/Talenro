# Error Projection Recovery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after explicit plan/budget approval. No implementation agents, extra reviewer, suite reruns or cleanup. Steps use checkbox syntax.

**Goal:** Project unavailable ErrorRecord metadata truthfully, prove strict consumers still reject incomplete real errors, and permit one separately bound replacement writer attempt.

**Architecture:** Freeze an actual-source projection regression before RED; minimally repair capture and fixed attempt bindings; run projection GREEN and a separately versioned full scope GREEN. Preserve the old source epoch and failures as immutable history before creating one new writer epoch.

**Tech Stack:** Existing Windows PowerShell7.6.5 Core x64, AST/JSON/hash APIs; no installs.

**Spec:** [Approved error-projection design](../specs/2026-11-01-offline-error-projection-design.md), selected approach1. Written design approved by user reply "是" after commit29d5139c; this plan and execution allowance remain PENDING_APPROVAL.

## Global Constraints

- Existing worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; branch codex/c12-b01-task9-coordinator; origin https://github.com/win202569/Talenro.git. Native method retained.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2; P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner; Q=.superpowers/sdd/2026-11-01-offline-controller-recovery. Aliases below mean exactly these paths.
- Preserve all six stops,124snapshots, old tests/inputs/results/archives and frozen design/plan documents. No regrading, reconstruction, overwrite or replay of old attempts.
- Semantic repair only L/capture.ps1. Fixed rebindings/provenance may also change L/driver.ps1,record-boundary.tests.ps1,source-model.ps1,observation-contract.tests.ps1. Other26runtime sources stay unchanged in writer epoch. New auxiliaries outside31runtime inventory, never imported by capture/profile.
- Request3auxiliary computations,1replacement writer host,1fixture compile,31replacement writer snapshots. Cumulative ceilings9auxiliary/7hosts/3compiles/186snapshots; consumed6/2/2/124. Original remaining4later hosts/7targets/conditional1review unchanged. No refund or added targets/reviewers.
- Auxiliary computations use existing tool context, max2private modules each; no profile body, record writer, Add-Type, child launch or exploratory probe. Explicit controlled throws/synthetic records are permitted test data only.
- First unexpected result stops: no patch/retest/re-freeze. New P/error-projection.stop.json always blocks. Old verifier stop recoverable only by exact approval/history bindings; no public bypass/path/callback. Preserve original process ownership, timing, recorder, oracle and five false-capability gates.
- No main merge, force push, deletion, installation, security change or broader product acceptance. Mechanical archive generation/hash checks are not test computations. No default Bash/task-done suite execution.

## Review Focus

1. A synthetic record has null InvocationInfo: NullInvocationInfoProjects must expose null Position without inventing source metadata.
2. A generic null-safe converter weakens real I/O acceptance: four null/missing Position/Stack negatives must exercise the actual consumer.
3. Nested projection loses array shape, null or shared-record distinctions: NestedJsonRoundTrip and SharedExceptionIdentity cover both representation and live identity.
4. A secondary projection/assertion masks the first failure: protected setup/case reporting retains actual first ErrorRecord, partial groups and recording error; later bodies never execute.
5. New epoch borrows old successful identities or reuses spent paths: fixed distinct namespaces, historical snapshot checks, exact current receipts and newest-stop refusal.

## Task1 — freeze projection regression and observe old-source RED

**Files:** create L/error-projection.tests.ps1; P/error-projection.approval.json,
error-projection.input.json, error-projection-red.intent.json/actual records;
Q/error-projection-progress.md. Existing source/test files unchanged before RED.

**Interfaces:** Invoke-ODErrorProjectionRegression -Phase Red|Green; definitions
only on loading. Fixed phase prefixes error-projection-red/error-projection-green.
Normal ordered receipt: Schema,Phase,TestSHA256,InputSHA256,CodeSHA256,Groups,
PASS,FAIL,NotRun,ExitCode,RootError. Unexpected diagnostic: Schema,Phase,
Unexpected,ExitCode,FailedGroup,CompletedGroups,RootError,RecordingError,
TestSHA256,InputSHA256,CodeSHA256,PostCodeSHA256. Diagnostic exit2 is never GREEN.

- [ ] Verify source baseline from29d5139c and record the actual newer documentation-checkpoint HEAD; do not require HEAD itself to equal29d5139c. Verify parent P/writer-scope-repair.frozen.json SHA7F09A142E228FF16CD7228287DEBBB5EE60DEC9A91FBF1EF691F148C0A911587, all2236pins,124snapshots and absence of new paths. Capture actual plan/design/approval hashes, six stops and old failed writer originals; establish the Q ledger. Do not run diagnostic code during preflight.
- [ ] Author and commit regression before RED. Freeze ten ordered groups: NullInvocationInfoProjects,RealCaughtMetadataExact,NullStackProjects,SharedExceptionIdentity,NestedJsonRoundTrip,NullPositionRejected,MissingPositionRejected,NullStackRejected,MissingStackRejected,UnsupportedValueRejected. Freeze old capture SHA936ED8D80A479A3C7EA6CCD0FF6E6E5AFA035552D469045F18300742E9855A5C and actual old driver hash from parent snapshot; full-source hashes, not function-text substitutes.
- [ ] Extract actual Capture-Value and actual Driver-ErrorLocation/Driver-Need definitions uniquely by AST; no driver top-level execution. A frozen named helper inside the regression file catches a literal controlled throw, preserving its actual file-backed position/stack. Create explicitly synthetic ErrorRecord instances using that exception, id ContractOnly/category NotSpecified/target null; never throw those wrappers to manufacture metadata. No module/profile/writer work required.
- [ ] First group requires actual null InvocationInfo, exact five ordered projection keys and null Position. Invoke old Capture-Value once. Keep actual PropertyNotFoundException/PropertyNotFoundStrict with PositionMessage message and real position/stack before checking signature. RED must be0PASS/1FAIL/9NotRun/exit1. Different failure or unexpected pass stops. Save entire tool input/result and actual normal receipt; do not synthesize a missing final.

## Task2 — minimal repair and one projection GREEN

**Files:** modify only the five-source window; create L/error-projection-scope.tests.ps1,
P/error-projection-scope.input.json. Old auxiliaries remain frozen.

**Interfaces:** Capture-Value($Value) unchanged; ordered ErrorRecord projection
Type,Message,ErrorId,Position,Stack. Get-ODAttemptSlot('writer-green') now returns
writer-projection-green; the other four grants remain unchanged. Get-ODWriterManifestPath
returns fixed P/writer-projection-repair.frozen.json.
New Assert-ODProjectionFacts([Collections.IDictionary]$Facts) accepts exact ordered
keys Schema,ParentSHA256,HistoricalStopSHA256,ApprovalValid,AttemptConsumed,
NewStopPresent. Schema int1; parent7F09A142…11587; historical stop4B06ED58…FD14;
flags typed bool true,false,false. Every other input throws exact RuntimeException
Message/FQID 'OD: error projection recovery mismatch', no output on acceptance.

- [ ] Change only Capture-Value's ErrorRecord Position expression: inspect null InvocationInfo before reading PositionMessage, emit actual null or unchanged actual text. Preserve actual ScriptStackTrace including null, Type/Message/ErrorId, recursion, output shape, unsupported-value rejection, and original records. No catch-all, empty-string substitution or fabricated location.
- [ ] Rebind writer unit/capture/evidence/intent/probe/prior-final paths to writer-projection-green and manifest to writer-projection-repair.frozen.json in the existing five sources. Keep canonical grant writer-green and original oracle/groups. Add Prepare grant writer-projection-rebind; old writer-scope-rebind is spent. Later repair-green paths stay original and unconsumed.
- [ ] Add Driver-ProjectionGate and new pure facts predicate. Require fixed new approval/input/test identities and exact RED/GREEN/full-scope receipts, six historical stops and actual stopped writer streams/final/journal/tool record. Validate old successful scope/verifier source identities against snapshots/writer-scope-rebind, NOT edited current sources. For historical2236pin validation, all31historical operational sources use that snapshot epoch; non-source historical pins retain exact original paths/hashes. Do not run old Driver-Verify against mutated current sources. Keep older controller checks against snapshots/controller-rebind. Preserve parent OwnerScopeRecovery map exactly; do not regenerate historical maps from current bytes. Old current-stage gates become historical-only validation, not a replay path or a blocker on the explicitly recovered old stop.
- [ ] Preserve Driver-UnitError and Driver-ErrorLocation bodies byte-identical to baseline, including strict actual source/operation/stack checks. Old parent2hosts/2compiles remain consumed. New current chain keeps MaxHosts5/MaxCompilations1/MaxChildren7/MaxReviewers1; additive recovery provenance discloses total7hosts/3compiles/9auxiliary/186snapshots. No new runtime action is allowed by false-capability flags.
- [ ] Create full scope auxiliary from the frozen verifier version, retaining typed definition checker/protected case runner and nine ordered assertions. Static-check the checker/case-runner/error-projection helper definition texts byte-identical to the frozen verifier helpers whose earlier8group regression passed; their behavior is not being rewritten. Version only names/paths/bindings and original five-source parent window. Fixed entry Invoke-ODErrorProjectionScope -Phase Green. Exact nine-group receipt schema matches existing scope receipt; failure diagnostic retains FailedGroup,CompletedGroups,actual RootError,Comparison,RecordingError,SourceSHA256,PostSourceSHA256 and test/input identities. Setup/extraction stay protected; do not perform unguarded work after a failed group.
- [ ] Freeze/commit full auxiliary and full input BEFORE projection GREEN. Full input binds new five current source hashes, new tests, regression input, design/plan/approval, old history and parent source map. Regression input must not bind future repaired source hash; GREEN gets that identity from fixed full input. Neither input references a future phase receipt, avoiding circular hashes.
- [ ] Consume error-projection-green.intent.json as computation8of9; run frozen regression once. Require10PASS/0FAIL/0NotRun/exit0. NullInvocationInfoProjects and NullStackProjects preserve actual null; RealCaughtMetadataExact compares every field to the caught record's actual metadata. SharedExceptionIdentity checks separate record references/shared exception before and after conversion. NestedJsonRoundTrip uses a literal nested ordered dictionary containing both wrappers, the real caught error, a null slot and an array, verifies exact keys/order/counts/nulls and unchanged original references after JSON round-trip.
- [ ] Four rejection groups clone a valid actual caught-error projection and independently set/remove Position/Stack, then call actual Driver-ErrorLocation with frozen regression source allowlist. The valid original must pass. Null cases require RuntimeException Message/FQID 'Actual error source/stack required'. Omitted-member cases permit only that same rejection or PropertyNotFoundException/FQID PropertyNotFoundStrict naming the specific missing Position/Stack member; never accept arbitrary exceptions or successful evidence. Preserve raw ErrorRecord before assertions; do not substitute a fake consumer. UnsupportedValueRejected requires exact 'Unsupported evidence value' rejection from actual converter. Preserve current code/test/input hashes before/after; no test/source edit after GREEN.

## Task3 — one full scope GREEN, unique source freeze and bounded resume

**Files:** same frozen five current sources; new P/error-projection-scope-green.*,
writer-projection-repair.intent.json/frozen.json, snapshots/writer-projection-rebind;
L/portable-error-projection-freeze.json and final L/portable-error-projection.json.

- [ ] Consume full scope intent as computation9of9 and run frozen Invoke-ODErrorProjectionScope -Phase Green once. Require original nine ordered groups allPASS/0FAIL/0NotRun/exit0, null RootError, no Unexpected and unchanged five source hashes. Existing FourFunctionsPersist/DefinitionsExact/ModuleIsolation remain actual-module tests, max2modules. AttemptBindings/NewEpochBindings check new fixed namespaces; ConsumedAttempt/RecoveryGate exercise the new strict facts predicate; HistoricalEpoch/OriginalContractsUnchanged retain historical proofs and other26source equality. Never relabel the old scope GREEN as this result.
- [ ] New source freeze inherits2236pins, changes exactly the five operational source pins relative to parent7F09A142…11587 and appends exactly15unique provenance pins below; total2251. Preserve InheritedPins2482,InheritedUniquePins2172,SourceCount31 and SourcePaths. Parent OwnerScopeRecovery remains unchanged; add ErrorProjectionRecovery with Schema1,parent/new approval/design/plan/test/input hashes, all three new actual finals, historical verifier stop and failed writer originals, consumed/ceiling counters. Actual final hashes are map bindings, not circular source pins.

The15new pin paths are fixed:
P/writer-scope-repair.frozen.json; P/owner-scope-verifier.stop.json;
docs/superpowers/specs/2026-11-01-offline-error-projection-design.md;
docs/superpowers/plans/2026-11-01-offline-error-projection.md;
P/error-projection.approval.json; L/error-projection.tests.ps1;
P/error-projection.input.json; L/error-projection-scope.tests.ps1;
P/error-projection-scope.input.json; P/writer-scope-green.final.json;
P/capture/writer-scope-green/journal.jsonl; P/writer-scope-green.intent.json;
P/writer-scope-green.stdout.bin; P/writer-scope-green.stderr.bin;
P/owner-scope-verifier-writer.tool-record.json.
Readonly planning confirmed all15absent from parent's pin paths; recheck before freeze.

- [ ] Run driver -Action Prepare -Grant writer-projection-rebind -PreviousFrozenSHA256 7F09A142E228FF16CD7228287DEBBB5EE60DEC9A91FBF1EF691F148C0A911587 once. Require2251pins/31snapshots,total155/186, exact five changed sources. CreateNew source intent precedes snapshots/manifest; no retry on partial output. Gate refuses new writer/future outputs, but permits historical consumed writer-scope-green artifacts. Later contract start must not falsely reject its already-approved repair freeze.
- [ ] Archive every actual P/Q file/directory/zero-byte file into NEW portable-error-projection-freeze.json (LF container, original embedded bytes, decoded hash/byte equality). Update execution/current-status/recovery; explicitly stage necessary files only, verify raw staged bytes==disk and diff checks, commit/push existing branch, remoteSHA==HEAD, clean. Freeze source/test bytes before runtime.
- [ ] Run driver -Action RunOne -Grant writer-green -FrozenSHA256 <actual new manifest hash> once. New fixed attempt writer-projection-green consumes1host/1fixture compile. Require original26PASS, exit0, exact stdout/stderr, complete schema/evidence and recorder journal, actual parent acceptance and intact pins. Missing evidence/partial/unknown is NOT_ACCEPTED; preserve actual stop and freeze all remaining stages. No additional compilation or writer retry within this plan.
- [ ] Only after actual writer acceptance: original observation-contract.ps1 implementation and unique Prepare repair-green/31snapshots (total186), then original contract52PASS/core25PASS22rows/lifecycle17PASS8rows3targets/continuation16PASS+12seamPASS15rows4targets chain. Same original120000ms host cap except continuation150000, unchanged target/process/stream bounds and actual-source/false-capability gates. Conditional single runtime reviewer only after every actual runtime stage succeeds; no earlier review or new reviewer allowance.
- [ ] Success or stop: preserve actual errors/tool chunks without missing-field reconstruction, create NEW portable-error-projection.json with decoded equality; old archives unchanged. Update execution report/current-status/recovery and push explicit necessary-file checkpoint with verified remoteSHA/clean worktree. Passing auxiliary tests or writer tests alone is not product/I1/I2 acceptance.

## Self-review and execution approval

Spec coverage: Task1observed failure; Task2truthful projection/strict negatives;
Task3new-source integration, historical binding, freeze/runtime/portable handoff.
Review focus maps to ten regression groups and nine integration assertions.
Signatures/key order/phase paths are fixed, inputs precede outputs and avoid cycles.
Approval records must preserve the user's actual reply and exact approved plan/design
hashes; no source/test/computation occurred while planning. Plan/budget approval
must explicitly cover the3auxiliary/1host/1compile/31snapshot addition and bounded
conditional resumption of the original remaining chain. No automatic execution
from design approval; native method preserved. Stop rather than invent extra work.
