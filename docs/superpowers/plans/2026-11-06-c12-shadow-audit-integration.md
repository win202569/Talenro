# C12 shadow source-audit implementation plan

## Execution addendum — 2026-11-06 Task2 continuation

After checkpoint8bed43f5 the user explicitly approved Task2 with reparse coverage
blocked, not passed. Records phases retain that blocked result without retrying
junction creation; other parser cases rerun in fresh Records phase roots.
Optional typed Synthetic=false on New-C12ShadowSession supplies the specified
test/real distinction without changing real defaults. RecordsRed8714cd expected
missing interface; RecordsGreen9cab1f→24eaa7 exit0,21 matched groups+1 blocked.
Total4/7 planned calls consumed plus one separately approved junction diagnosis.
Task2 core behavior implemented; author review found final summary elapsed field
missing, so full Task2 completion is NOT claimed. A bounded follow-up RED/GREEN
needs approval; do not use remaining Task3/4 calls as implicit retry authority.
No formal hooks/manifest/combined suite or live run this continuation. Historical
plan text and its original stop/budget requirements below remain applicable.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline, task by task. Await user approval before implementation.

**Goal:** Add default-off source audit recording without changing formal C12 verdict logic or resource ownership.

**Architecture:** A definition-only module validates a frozen manifest and publishes immutable group evidence. Guarded runner hooks capture maps from actual group source roots; a final run summary follows all outer cleanup/environment restoration. Audit failures stay separate from formal outcomes.

**Tech Stack:** Existing Windows PowerShell/.NET, SHA256, strict JSON, local files; no dependencies.

**Spec:** `docs/superpowers/specs/2026-11-06-c12-shadow-audit-design.md`, user-approved after2017b01d. Prior16 verifier/10 parent cases are supporting evidence, not proof of this integration.

## Global constraints

- Existing development branch only. Preserve old diagnostics. No main merge, live runner invocation, Go/Docker, OS children, compilation, installs, snapshots or capability promotion.44/34/6/217 unchanged; NOT_ACCEPTED.
- Only explicit `-ShadowAudit` enables hooks in Focused/Suite. Off means no module import, audit file reads, directories or summaries. No environment toggle/external path parameter.
- Manifest <=1048576 bytes,1..128 files; source aggregate <=33554432 bytes; summary <=1048576 bytes. No sampling when limits exceeded. Synchronous I/O has no hard cancellation guarantee; do not extend existing deadlines.
- Tasks1–3 each permit one RED and one GREEN invocation; Task4 one combined invocation: maximum7. Stop any unexpected outcome without repair/retry. Intended RED→implementation→GREEN is authorized.
- New exclusive phase/case roots under `.superpowers/sdd/2026-11-06-c12-shadow-audit-integration-v1`; never reuse/delete evidence. Combined suite has <=24 module cases plus12 hook scenarios, no child processes. Each script claims its phase root and records all outcomes.
- Inline author review only; no independent security/full-suite acceptance claim. Preserve necessary files and push at handoff. Live use requires separate approval.

## Files and initial source coverage

Create `scripts/private/c12-shadow-audit.ps1`, `scripts/tests/c12-shadow-audit.tests.ps1`, `scripts/tests/c12-shadow-hooks.tests.ps1`, `scripts/tests/c12-shadow-offline-suite.ps1`, `testdata/c12/shadow-source-manifest.v1.json`.
Modify `scripts/run-c12-integration.ps1` only for the switch, guarded hooks and additive outcome bookkeeping. Update status/recovery/gap-review documents.

Manifest exact paths: go.mod; go.sum; scripts/run-c12-integration.ps1;
scripts/private/c12-shadow-audit.ps1;
internal/testinfra/c12_dependencies_integration_test.go;
testdata/c12/authority-v7/schema-name-registry.v1.json.
These cover dependency declarations, runner/audit code, dependency-test source and
schema registry—not all Go/SQL/binaries/runtime inputs. Exclude manifest itself to
avoid self-hash recursion; bind its own digest separately. Freeze pins only after
final code changes, never regenerate during a run. Verify checkout EOL conventions
reproduce pinned bytes; keep repository EOL rules, never weaken byte checks.

## Shared contracts

Manifest exact keys: Schema(integer1), Mode(stringSHADOW), Files(array of objects
with Path(canonical slash-relative string), SHA256(64 uppercase hex string)).
Reject missing/extra keys, duplicate/case-aliased JSON keys or paths, wrong types,
absolute/drive/UNC/ADS/traversal/wildcard paths and reparse ancestors. Rooted paths
stay beneath captured SourceRoot. Valid missing source=DRIFT; unsafe/unreadable
input=INCOMPLETE. Observe missing files as Present=false, Length/SHA256=null.

Definition-only public APIs (return dictionaries, no raw exceptions/workload output):

- `New-C12ShadowSession([string]$InvocationRoot)` → AuditId(32 lowercase hex), EvidenceRoot, Groups(list), FirstReason, ModuleSHA256. EvidenceRoot is InvocationRoot/.superpowers/c12-shadow/<AuditId>; exclusive creation, reject occupied/reparse roots, no retry or external destinations.
- `Start-C12ShadowGroup($Session,[string]$SourceRoot,[System.Collections.IDictionary]$Identity,[datetime]$Deadline)` → frozen group context and persisted before-map. Identity exact keys GroupID/Profile/Package/RunSuffix/CandidateTree; last may be null when unavailable. Read manifest from SourceRoot; freeze list/pins and actual manifest/module identities.
- `Complete-C12ShadowGroup($Group,[bool]$TestFailed,[bool]$CleanupFailed,[bool]$OwnershipConfirmed,[datetime]$Deadline)` → provisional group record reference. Skip after-scan on cleanup failure, uncertain ownership or expired deadline. Never final MATCHED here.
- `Complete-C12ShadowSession($Session,[int]$FormalExitCode,[bool]$OuterCleanupFailed,[bool]$EnvironmentRestored,[datetime]$Deadline)` → final record reference or in-memory INCOMPLETE. Verify referenced bytes and attempt bindings before publishing. Caller never assigns its exit code from this return value.

Records schema1 include ModeSHADOW, Synthetic(bool), audit/group bindings,
covered relative paths, map references(Path/Length/SHA256), elapsed milliseconds,
original test/cleanup Boolean outcomes and closed state/reason. Real hooks use
Synthetic=false; tests true. Serialize only allowlisted fields, never arbitrary
caller objects, source contents, environment values, secrets or raw logs.
No ObservationPassed field. State precedence INCOMPLETE > DRIFT > MATCHED;
zero groups=INCOMPLETE. Before drift remains drift even if later restored.
Reason codes: INVALID_MANIFEST,UNSAFE_PATH,LIMIT,IO,IDENTITY,DEADLINE,OWNERSHIP,
CLEANUP,NO_GROUPS. First error retained; missing final summary is INCOMPLETE.
Public boundaries catch failures; warning sink itself is best-effort, emitting
only audit ID/reason with prefix `C12 shadow audit`. No overwrite/retry.

## Review focus

1. Wrong Suite root/snapshot writes → Task3 distinct-root fixtures and snapshot immutability.
2. Outer cleanup fails after maps match → Task3 final INCOMPLETE, unchanged formal error.
3. Missing/wrong-attempt evidence treated as matched → Task2 identity/publication cases.
4. Off-path side effects or warnings mask primary error → Task3 off equivalence/throwing warning sink.
5. Timing/privacy/EOL errors → deadline/marker-secret cases and Task4 checkout-byte check; no hard-timeout claim.

## Task1: Strict manifest and safe source scanning

**Files:** module and module tests. **Interfaces:** private `Read-C12ShadowManifest(SourceRoot)` → frozen validated manifest/digest; `Get-C12ShadowMap(SourceRoot,Manifest,Deadline)` → observations/state/elapsed time, consumed by Task2.

- [ ] Write tests: valid tiny file→MATCHED; changed/missing→DRIFT; duplicate keys/paths, Boolean Schema, malformed JSON, absolute/traversal/ADS/reparse paths→INCOMPLETE. Expected file digests come from independent fixture bytes.
- [ ] Run `& ./scripts/tests/c12-shadow-audit.tests.ps1 -Phase ParserRed` once; expected RED missing module/interface. Save receipt.
- [ ] Implement strict parsing/path/size checks without importing old diagnostic entry points. Hash bounded streams; enforce aggregate limit during reads (files can grow). Reparse checks are defensive, not hostile-race proof. Tests may use a local fixture junction only if supported without privilege changes; if unavailable record that coverage blocked, not passed.
- [ ] Run same script `-Phase ParserGreen` once; require matching cases including129 files, manifest oversize, aggregate oversize and expired deadline. Use bounded fixtures, never fill a disk. Commit module/tests; stop unexpected outcome.

## Task2: Immutable publication

**Files:** same module/tests. **Consumes:** Task1 scanner. **Produces:** four public APIs above.

- [ ] Write tests: before drift then restored; valid maps finalized only at session completion; wrong audit/group binding; changed reference bytes; occupied outputs; missing summary; cleanup/ownership/deadline refusal; no groups. Use actual filesystem writes/sentinels, not mocked saving.
- [ ] Run module tests `-Phase RecordsRed` once; expected missing lifecycle interface RED.
- [ ] Implement exclusive before/after/group/run files with closed leaf names from audit ID/group ordinal, never untrusted GroupID paths. Verify published bytes; keep failed references null and first reason. Never infer persistence from an in-memory success. All summaries obey byte limit.
- [ ] Run module tests `-Phase RecordsGreen` once including Task1 cases, <=24 total. Include unreadable manifest, occupied summary and a throwing warning sink; marker passwords must not appear in evidence/console. Commit after GREEN; no runner call.

## Task3: Guarded formal runner integration

**Files:** runner, hook tests, manifest. **Consumes:** module APIs. **Produces:** switch and bound shadow evidence, unchanged formal verdict logic.

- [ ] Write tests extracting actual hook blocks and original outcome/cleanup tails by AST; never execute/dot-source complete runner. Supply synthetic process outcomes, but use real module I/O and real extracted formal decision code.
- [ ] Run `& ./scripts/tests/c12-shadow-hooks.tests.ps1 -Phase Red` once; expected missing switch/hooks RED, never live work.
- [ ] Add switch to both parameter sets. Capture InvocationRoot before Suite switches roots. Guard import/session creation and contain import exceptions. Load audit module from invocation checkout and record its digest; measure Suite source-module pin separately rather than assuming equality.
- [ ] Start group audit just after runSuffix validation, before resource/database setup, with actual source root and existing groupDeadline. End after existing group finally but before existing pass/throw. Use primaryFailure/cleanupFailures without rewriting them. OwnershipConfirmed=false on cleanup failure or native/process failure lacking positive ownership confirmation; skip rather than guess. SourceRoot is Focused checkout or actual Suite groupSnapshot.Root; evidence always outside snapshots.
- [ ] Add shadow-only flags for caller snapshot validation/removal and suite candidate cleanup failures, preserving original rethrows/finally. Final hook follows outer prepared-artifact cleanup and environment restoration immediately before existing exit; retain original scriptExitCode on audit errors. If restoration itself throws, no summary is valid.
- [ ] Group hooks retain original deadlines. Final hook uses last stored group deadline capped by suite deadline, never new time allowance; no groups→NO_GROUPS. Expiry may leave no final summary. Original cleanup keeps its independent existing budget; audit never runs inside cleanup loops.
- [ ] Freeze manifest pins from completed files. Run hook tests `-Phase Green` once:12 scenarios—off valid, off missing module, matched, drift, audit I/O failure, formal test failure, group cleanup failure, unknown ownership, Suite distinct roots, Suite outer cleanup failure, restoration failure, final deadline expired. In each assert original formal output/error/exit selection. Off uses throwing audit-call sentinels proving unreachable; on uses real persistence. Fold warning-sink failure into I/O case.
- [ ] Diff review removes only named additive switch/hooks/bookkeeping: original allowed sets, resource operations, cleanup order, deadline expressions and pass/throw/exit conditions remain identical. Commit only after GREEN.

## Task4: Combined offline handoff

- [ ] Create suite orchestrator for <=24 module cases+12 hook scenarios in fresh claimed roots. Run once `& ./scripts/tests/c12-shadow-offline-suite.ps1`; expected `C12_SHADOW_OFFLINE_VERIFIED`, no failed/unrun required cases. No full runner, Go/Docker/process API or consumed diagnostic rerun.
- [ ] Verify manifest pins against Git checkout byte conventions; module import definition-only; unrelated original artifacts unchanged. Archive bounded receipts/fixtures with decode/hash check, excluding binaries/caches/credentials and redacting marker secrets from published artifacts.
- [ ] Update status/recovery/gap review: OFF_BY_DEFAULT, coverage, counts, source-pin maintenance, timing caveat. Suite candidates without feature/manifest give INCOMPLETE when opted in, not accidental checkout fallback.
- [ ] Commit explicit files, push existing branch, verify remote SHA/clean status. Stop. Any live Focused run needs separately proposed budget, dependency/environment checks and user approval.

## Self-review and approval

Tasks map to spec:1 types/paths/limits,2 records/states/privacy,3 roots/lifecycle/
verdict/deadlines,4 recovery. Interfaces and field names align. Reparse coverage
cannot silently pass if environment prevents its fixture. Full disk, mid-write/
flush, hung I/O, clean-machine operation and independent security acceptance are
not covered. PLAN_READY_FOR_REVIEW; no implementation or tests performed yet.
