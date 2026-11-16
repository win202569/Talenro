# Guardian phase2 controlled execution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for approved current-session execution, or superpowers:subagent-driven-development if explicitly selected. Checkboxes below are pending, not permissions.

**Goal:** Execute the approved phase2 contracts in small verifiable source/mock-test batches, without crossing into native build or live operation.
**Architecture:** Follow the six tasks and exact interfaces in the contract supplement. Batch A establishes pure control-v2/handshake contracts; subsequent batches add mocked adapters and wiring. Keep production composition unavailable to tests.
**Tech Stack:** Existing .NET10 managed test project, xUnit2.9.3, PowerShell7.6.9; existing pinned App dependencies unchanged.
**Spec:** [cooperative design](../specs/2026-11-16-windows-cooperative-close-guardian-design.md) and [phase2 contracts](2026-11-16-windows-guardian-phase2-adapter-contracts.md).

## Authority and baseline

User approved H1 (narrow guardian-recipient transfer-rights exception) and H2
(separate control-v2 envelope), then explicitly requested planning only.
Neither implementation nor any test execution is currently approved.
Planning base28b5b6c24736175d810d3c67661e3bebd772c9c8. Phase1 source85d4bdae
has historical467/467 managed tests; do not call that a fresh phase2 baseline.
Worktree: D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator.
Branch: codex/c12-b01-task9-coordinator. Preserve it; no main merge or history rewrite.

## Global Constraints

- Preserve exact native/CIM creation, connected-instance/retained-handle identity,
  ordinal command-line binding including trailing U+0020, all token/hash/source/
  launcher/path/reparse/zero-target checks. No broadened comparisons or auto-pins.
- H1 permits only the reviewed guardian-recipient transfer exception; never
  terminate rights on App, broad access, privilege adjustment or security changes.
- READY<=10000ms prelaunch; BOUND<=10000ms since actual immutable T0. Capture and
  preserve acquisition latency, never report synthetic zero to extend the budget.
- Target290000ms; complete send/accept/invoke<300000ms; exit/finalization<360000ms.
  One attempt; no reconnect/retry/force kill/external-window fallback.
- Phone confirmation60000ms, last prompt210000ms, irreversible cancellation on
  close/failure/end. No stale reply restores authority.
- No App/helper build, restore, native call, guardian/App launch, package install
  or user settings change in the proposed source/mock batches.
- Managed Core/Tests compilation is a separate explicit part of batch approval;
  existing property-only MSBuild tests may run, never App Build/Restore targets.
- Preserve all old binaries/raw receipts. No test derives a policy from current
  arbitrary binary hashes. New ordinary variant stays NOT_READY for live runs.

## Batch boundaries and owned files

All paths below relative to clients/windows; full task details remain in the
contract supplement. A batch approval does not authorize the next batch.

| Batch | Tasks/deliverable | Files permitted after approval | Stop point |
|---|---|---|---|
| A | Task1 pure control-v2, directional handshake, early/loss reason fixtures | CloseGuardian.Shared/{GuardianAdapterContracts.cs,GuardianControlProtocol.cs,GuardianHandshake.cs}; Talenro.Demo.Tests/{GuardianControlProtocolTests.cs,GuardianHandshakeTests.cs,GuardianSessionTests.cs,Talenro.Demo.Tests.csproj} | Managed evidence + checkpoint; no Windows/App/helper files |
| B | Tasks2–3 fakeable identity/ownership/clock/I/O/receipt logic | CloseGuardian.Windows/{PeerIdentity.cs,LocalPipe.cs,GuardianClock.cs,GuardianReceipts.cs}; required shared port declarations; corresponding Guardian*Tests.cs and explicit test links | Mock evidence only; native gateway not exercised |
| C | Task4 serialized runner/helper source/coordinator | CloseGuardian.Shared/GuardianEffectRunner.cs; Talenro.CloseGuardian/{Talenro.CloseGuardian.csproj,Program.cs,GuardianHost.cs}; tests/guardian-coordinator.psm1 and .tests.ps1; GuardianEffectRunnerTests.cs | No helper project build or launch |
| D | Tasks5–6 dormant App hook/isolated variant/inert entry | App Observation files/App.xaml.cs/App csproj; Directory.Build.props; observe-ordinary-cooperative-close.ps1 and .tests.ps1; endpoint/UI/isolation tests | Final independent source review, then STOP before native build |

Source changes to existing pure files beyond these listed needs must be justified
by a failing contract test and documented; architectural/permission changes return
for approval. No blanket permission to repair unrelated legacy tests.
Every batch may update current-status/recovery/evidence docs and commit/push only
its reviewed necessary files under standing repository instructions.

## Review Focus

1. A: Control-v2 accidentally accepted by app-v1, or replay after stage change.
2. B: Fake lease disposal reaches a native finalizer or closes reused numeric handle.
3. B/C: cancellation completes the waiter but leaves a late action armed.
4. C/D: phone confirmation or lost BOUND reply reactivates/relaunches a session.
5. D: project/import evaluation triggers build/native initialization or inherited activation.

## Common preflight — after the selected batch is approved

- [ ] Record approved batch, scope, baseline SHA and expected tests in plan ledger.
- [ ] Read worktree AGENTS.md and applicable nested instructions; verify branch,
  current diff and upstream. Preserve unrelated changes; never reset/clean them.
- [ ] Read-only confirm ordinary host token and required existing test assets.
  If host is elevated, context unknown or another relevant test session active,
  stop and reconcile; do not alter UAC, kill processes or start a concurrent suite.
- [ ] Inspect selected tests/project graph. Confirm no App/helper ProjectReference,
  production static initializer, native gateway construction or launch target.
- [ ] Run managed baseline once using command M below, unchanged timeout and
  --no-restore. Exit0 with explicit counts is required before source work.
  Missing dependencies/testhost infrastructure failure stops, no automatic rerun.

## Batch A execution steps (recommended first approval)

Consumes existing ProcessIdentity/GuardianPolicy/v1 codec and phase1 state types.
Produces ControlEnvelopeV2/ControlProtocol and GuardianHandshake signatures from
Task1 of the supplement. Leases at this stage are declarations/in-memory tokens,
not owning real SafeHandles or calling native finalizers.

- [ ] Add pure tests and explicit shared Compile links in Tests csproj only.
  `ReadyBeforeTargetExists`: app unassigned,T0=0,handle0 valid only at READY.
  `AssignHandleOnce`: real identity/T0/nonzero value, one adoption intent, duplicate0.
  `AbortBeforeAssignment`: seq1 terminates; no later assignment accepted.
  `AbortAfterAssignment`: seq2, real tuple, no counter reset.
  `WrongCoordinatorOrMacRejects`: every identity/MAC/domain/version change rejects.
  `SequenceResetRejects`: directional sequence continuous across different kinds.
  `LostBoundAckRemainsArmed`: no relaunch intent and no prompt authority.
- [ ] Add literal independent control-v2 vector, not generated by Encode under
  test; assert total4096 fits,4097/truncated/invalid UTF8/tail/replay reject. Verify
  v1 byte vector unchanged and app never accepts control-v2 as a close frame.
- [ ] Add EarlyReasonSurvivesWire at10000ms and ParentLossSurvivesWire at10001ms;
  decoded reason equals state-selected reason, not fixed Deadline fixture; mark
  scheduled autonomous coverage NOT_EXERCISED for these paths.
- [ ] Run G below; observe intended RED assertions or missing contract only.
  Infrastructure/fixture setup errors are not accepted RED evidence.
- [ ] Implement the smallest pure contract/state progression; no OS resources,
  environment reads/writes, RNG invocation, timers or sleeps in test fixtures.
- [ ] Run G then M; require exit0, zero failures/skips attributable to new tests,
  expected test discovery and no changes to default timeouts/security checks.
- [ ] Review diff, record actual counts and source SHA, run git diff --check,
  explicitly stage source/tests/docs, commit/push existing dev branch and verify
  remote SHA equality. Stop for batch B authorization; no native compatibility claim.

## Subsequent batch execution tests

Each row is a separate RED->implementation->GREEN->commit task using the file and
interface ownership already defined in Tasks2–6; do not collapse all tests at end.

| Task | Required checks/assertions | Verification before checkpoint |
|---|---|---|
| 2 | Native creation+1/CIM precision mismatch/wrong connected peer/token/trimmed command reject; original handle retained on failure; wrong-recipient/adopt-twice/late numeric reuse produces no close; fake lease never native-finalizes | G+M; native gateway invocation count0 |
| 3 |4096/4097, EOF/partial frames, first invalid accept terminal; backward/overflow/frequency mismatch fail;299999 completion timely/300000 not; pending buffers retained across cancel race; blocked flush never sends | G+M; fake clock jump360000 leaves no pending authorized effect |
| 4 | READY missing prevents launch intent; delayed assignment500ms preserved; BOUND10001 rejects; control ACK lost remains armed; timer independent of hanging I/O; all close/failure/end paths Cancel phone, stale reply adds0 actions | G+M+C; fake launcher count<=1, real launch0 |
| 5 | Absent bootstrap0 listeners/timers; partial rejects; secret erasure contract; immediate/delayed/duplicate callback; dead/replaced/modal window, guardian loss; callback300000 invokes0; no external HWND | G+M; original fake close count<=1, real UI/native0 |
| 6 | isolated cooperative-close Debug/diagnostics false; incompatible variants reject; old paths/pins unchanged; no RunApproved/import yields NotStarted and0 CIM/receipt/native/build/launch; no inherited activation in modeled child env | G+M+C+O; property evaluation only, no App/helper Build/Restore |

Native gateway signatures/marshalling/ownership must be concretized in batch B
tests before their implementation; contradictory installed API requirements stop
for a scoped contract revision, never permission escalation by assumption.

## Command sheet — future, not run in this planning turn

Run one command at a time in the stated worktree and ordinary-token host. Read
its exit/result before the next. Capture each session ID and preserve raw failures.
Use the verified absolute pwsh executable if PATH lacks it; do not install one.

M — existing managed project, compiles Core/Tests only:
```powershell
dotnet test ./clients/windows/Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore
```
G — selected pure/adapter fake tests, same compilation boundary:
```powershell
dotnet test ./clients/windows/Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore --filter 'FullyQualifiedName~Guardian'
```
C — only after batch C test source exists and forbidden-call mocks inspected:
```powershell
pwsh -NoProfile -File ./clients/windows/tests/guardian-coordinator.tests.ps1
```
O — only after batch D test source exists and forbidden-call mocks inspected:
```powershell
pwsh -NoProfile -File ./clients/windows/tests/observe-ordinary-cooperative-close.tests.ps1
```
Expected GREEN: exit0, explicit pass counts,0 failures, no unexpected skips, and
zero real native/launch calls. A missing test file is not permission to substitute
a production entry. Do not pass -RunApproved to any observation script.

## Stops, final review and handoff

- [ ] On unexpected timeout/session cancellation, preserve output and stop.
  Only read-only reconciliation; no restart/kill/time-limit extension without approval.
- [ ] On test failure caused by changed code, diagnose using mocks, add regression
  and repeat deliberate RED/GREEN; never launch App to investigate.
- [ ] On architecture/rights/schema changes beyond H1/H2, report the exact mismatch
  and seek revision approval; do not weaken existing safety tests.
- [ ] After D, request one independent whole-phase source review, scoped from
  pre-A SHA; reviewer read-only, no own native tests. Fix Important findings with
  regressions, then final M/G/C/O; record deferred minors explicitly.
- [ ] Preserve verified versus NOT_VERIFIED fields: source/mock gates can pass;
  helper/native/XAML compilation, real pipe/handle cancellation, own Window
  lifetime, new hashes and live acceptance remain NOT_VERIFIED.
- [ ] Commit/push necessary files and recovery state; no merge, new runtime hashes,
  native build or new live allowance. Next proposal is isolated-build planning only.

Self-review: batch boundaries match Tasks1–6, source ownership/commands/test outcomes
are explicit, H1/H2 no longer pending design decisions. No duplicate approval
requested for those designs. Execution approval is still required, starting with A.
