# Cooperative guardian Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. No execution is authorized by this document.

**Goal:** Implement a per-run identity-bound cooperative normal-close path without relying on chat progress.

**Architecture:** A pure shared protocol/state layer is tested first using fake clocks, transports, receipts and UI dispatchers. Independent guardian and app adapters are added only after that gate; observer retains original-process exit authority. External HWND closure is forbidden.

**Tech Stack:** Existing .NET10 C#, xUnit2.9.3, Windows App SDK pinned by the repository, PowerShell7.6.9; no new third-party packages.

**Spec:** [approved revised design](../specs/2026-11-16-windows-cooperative-close-guardian-design.md), approved by user2026-11-16. This plan awaits approval.

## Global Constraints

- Exact original-process identity/path/command line/native creation and existing CIM precision rules; all six candidate hashes, ordinary token, launcher/source/path/reparse and zero-target gates remain.
- READY before launch within10s; BOUND by T0+10s before any test input.
- Target request290000ms; full send, app acceptance and UI close invocation strictly before300000ms; observation cap360000ms, maximum observer sleep2000ms.
- At most one claim/request/callback; partial send, error, lost ACK or cancellation never authorizes retry/reconnect. No force kill or external HWND fallback.
- Each phone step60s, latest new prompt210s; no deadline extension or batched keys.
- No helper/app/native tests in phase1; no App restore/build in phase1. Old binaries/receipts are preserved, never overwritten or silently rebound.
- Separate approvals: phase1 implementation/synthetic tests; phase2 adapters/integration; isolated build; native verification; one live no-keyboard observation. Do not chain these approvals.

## Review focus

1. Lost BOUND reply after real binding: remain armed, not safe-to-relaunch (Tasks2/4).
2. PID reuse or pipe name replacement: connected-instance/creation proof, never lookup-only acceptance (Tasks1/4).
3. Duplicate or delayed UI callback: one retained object, deadline checked at execution, not only enqueue (Tasks2/5).
4. Partial frame, blocked receipt or send crossing300s: unknown/failed, not replay (Tasks1/2/4).
5. Dormant activation and new artifacts: no accidental listener or inherited activation, no transfer of old acceptance (Tasks3/5/6).

## File ownership and authorization boundaries

Paths below are relative to `clients/windows/`. Shared files live outside App's
`Diagnostics/` exclusion patterns and are linked explicitly, not wildcarded into
all production code. Do not modify business-state Core models for this facility.

| Stage | Files to create/modify | Responsibility |
| --- | --- | --- |
| 1 | `CloseGuardian.Shared/GuardianContracts.cs`, `GuardianActivation.cs`, `GuardianProtocol.cs`, `GuardianState.cs`, `ApplicationCloseState.cs`; `Talenro.Demo.Tests/Guardian*Tests.cs`; explicit links in `Talenro.Demo.Tests.csproj` | Pure code and synthetic tests only |
| 2 | `Talenro.CloseGuardian/Talenro.CloseGuardian.csproj`, `Program.cs`, `GuardianHost.cs`; `CloseGuardian.Windows/PeerIdentity.cs`, `LocalPipe.cs`, `GuardianClock.cs`, `GuardianReceipts.cs`; `tests/guardian-coordinator.psm1`, `guardian-coordinator.tests.ps1` | Independent host, native adapters, coordinator |
| 2 | `Talenro.Demo.App/Observation/CooperativeCloseEndpoint.cs`, `CooperativeCloseUi.cs`; `App.xaml.cs`, App project, `Directory.Build.props`; `tests/observe-ordinary-cooperative-close.ps1` and `.tests.ps1` | Dormant app activation, own-window dispatch, isolated variant and new entry |
| 3+ | `tests/guardian-native-check.ps1` and dedicated receipts/manifest; roadmap/recovery docs | Explicitly approved native build/verification, not phase1 |

Keep old `observe-ordinary-account-center.ps1` and its historical pins unchanged.
New entry starts inert without `-RunApproved`, has no user-supplied executable,
hash, deadline or command-line overrides. Integration may reuse existing C5 module
only while preserving its exact identity, cap and evidence semantics.

## Phase1 — executable implementation checklist, synthetic only

### Task1: activation and authenticated message contract

**Interfaces:** namespace `Talenro.Observation.CloseGuardian`; records
`ProcessIdentity(int Pid, long CreationFileTimeUtc, string Path, string CommandLine)`;
`GuardianPolicy(Guid RunId, byte[] Digest, long T0Ticks, long Frequency)`;
`ActivationResult(ActivationStatus Status, GuardianBootstrap? Value, string? ErrorCode)`.
`GuardianActivation.Parse(IReadOnlyDictionary<string,string?> environment)` returns
Disabled/Valid/Rejected; never opens resources. `GuardianProtocol.Encode(Frame,
ReadOnlySpan<byte> key)` and `Decode(ReadOnlySpan<byte>, ReadOnlySpan<byte> key)`
return a bounded frame or explicit rejection, no reflection-based polymorphism.

Bootstrap uses only `TALENRO_GUARDIAN_V1` prefix keys: VERSION,RUN_ID,PIPE,KEY,
GUARDIAN_PID,GUARDIAN_CREATION,POLICY,T0,FREQUENCY. Missing all means Disabled;
partial, unknown prefixed key, invalid size/value or version means Rejected.
No bootstrap secret in ToString/errors/receipts. Secret bytes have explicit disposal.

Protocol decisions: version1, little-endian uint32 body length, max4096 bytes for
whole frame including prefix and32-byte MAC. Fixed field order, length-prefixed
UTF8 strings with strict decoding, no duplicate/optional unknown fields. Message
types are HELLO,CHALLENGE,PROOF,BOUND,CLOSE_NORMAL,ACK,REJECT; control channel has
its own separate domain and types READY,ASSIGN_TARGET,CLOSE_EARLY,ABORT.
MAC uses HMAC-SHA256 over domain/version/direction/RunId/sequence/policy digest/
peer identities/T0/frequency/challenges/payload.32-byte random key/challenges in
production; deterministic test vectors. Constant-time tag verification. Close
frame carries only fixed reason enum, no code/path/UI target supplied by sender.

- [ ] Write `GuardianActivationTests` for Disabled with empty map; every missing
  required key and unknown key rejected; wrong version/negative PID/zero frequency
  rejected; malformed secret rejected; no secret emitted by error formatting.
- [ ] Write `GuardianProtocolTests`: roundtrip fixed vector; tamper every covered
  field => reject; wrong direction/domain/key/RunId => reject; max4096 allowed,
 4097 rejected; incomplete prefix/body rejected without acceptance; bad UTF8,
  sequence replay and changed policy/T0 rejected. Fixed vector must be reviewed
  independently from Encode output rather than generated by the method under test.
- [ ] Run filtered tests RED, implement minimal pure files, rerun GREEN. No real
  pipes, process queries, P/Invoke, OS clocks, external files or UI objects in fixture.
- [ ] Review diff and save explicit task checkpoint; do not progress on unexpected failures.

### Task2: guardian and app state machines with injected effects

**Interfaces:** `GuardianState.Advance(GuardianEvent input, long elapsedMs)` returns
`GuardianDecision(State, Effect?, Reason)`; `ApplicationCloseState.Advance(AppEvent,
long elapsedMs)` returns the same form. Events carry explicit verified-identity,
receipt-completion, full-send, endpoint-bound and original-window-generation
evidence; never infer verification from non-null PID or ACK. Effects are enums
ClaimReceipt,SendClose,AcceptReceipt,EnqueueClose,InvokeOwnClose,Finish.
Executors acknowledge each effect by a subsequent event; pure machine never invokes
an OS action. Terminal states ignore duplicates without producing a new effect.

- [ ] `GuardianStateTests` asserts timer289999 => no claim;290000 => one claim;
  persist success => one send only if now<300000;300000 => no send;360000 =>
  terminal. Early/timer/channel-loss races in every ordering yield<=1 send.
- [ ] Assert missing READY prevents LaunchAllowed; BOUND after10000 rejected;
  BOUND at10000 allowed only if other evidence complete (deadline inclusive for
  handshake only); already bound/lost acknowledgement never resets to unbound.
- [ ] Assert original exited => no request; unknown identity/PID reuse => blocked;
  receipt failure => zero sends; partial/failed/unknown send => no second send;
  complete send at300000 => deadline failure. Decreasing clock => fail closed.
- [ ] `ApplicationCloseStateTests`: verified frame and receipt -> one enqueue;
  callback at299999 with same original generation -> one invocation; at300000,
  closed/replaced window, expired state or duplicate callback -> zero invocation.
  Failed enqueue, lost ACK, invalid MAC, duplicate request, receipt fault and
  guardian death never authorize another invocation or a pipe-disconnect close.
- [ ] Separate verdict test matrix: timely request but no exit != pass; ACK !=
  exit; manual/early original exit marks autonomous close NOT_EXERCISED; overdue
  cleanup/missing finalization != pass. Fake clock reaches360000 without chat.
- [ ] Run RED/GREEN as Task1. Commit task after review, no native adapter instantiation.

### Task3: synthetic end-to-end orchestration and phase1 gate

**Files:** `Talenro.Demo.Tests/GuardianSessionTests.cs`, `GuardianIsolationTests.cs`.
Test harness implements only queued in-memory transport/receipt/dispatcher effects;
links pure shared files explicitly in existing test project. No helper project
reference and no App project build dependency.

- [ ] Simulate READY->launch evidence->BOUND->silence->290s->send->UI->verified
  exit. Assert one launch intent/claim/send/enqueue/invoke and all timestamps.
- [ ] Simulate observer loss, app-binding gap, pipe replacement, truncated frame,
  delayed UI, hung effect (no completion event), blocked receipt and suspend-like
  clock jump past300/360. Assert bounded verdict failure, no replay/late dispatch.
- [ ] Assert absent bootstrap produces zero listener/timer intents; partial
  bootstrap fails observation; phone command at210000 allowed if full reserve
  remains,210001 rejected; stale acknowledgement adds no action.
- [ ] Run phase1 commands below once per deliberate RED/GREEN change; test failures
  stop implementation for diagnosis, never launch App to investigate.
- [ ] Review all diff, update status/recovery and push explicit source/tests/docs
  checkpoint. Stop and obtain separate phase2 authorization.

Commands (from repository worktree, existing dependencies only):

```powershell
dotnet test ./clients/windows/Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore --filter 'FullyQualifiedName~Guardian'
dotnet test ./clients/windows/Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore
git diff --check
```

Expected: first RED is the intended missing contract/assertion failure; final
commands exit0 with zero failed tests. Missing assets/dependencies is a blocker,
not permission to restore/install. These commands compile managed test/Core code,
not App; approval must explicitly allow managed test compilation. No tests run
while preparing this document.

## Phase2 — separately approved adapters and integration checklist

Do not implement this stage under phase1 approval. Before starting, expand exact
adapter method signatures and fixtures using the now-tested shared contracts;
review that supplement before native execution. These gates are not placeholders
that permit an executor to improvise unsafe APIs.

- [ ] Native identity adapter: connected-pipe GetNamedPipeClientProcessId /
  GetNamedPipeServerProcessId plus retained creation identity; DuplicateHandle of
  observer-owned app handle into verified guardian. Any OpenProcess use is query/
  synchronize-only and immediately creation-checked; no terminate rights. Reuse
  existing native/CIM precision guards. Unit mocks cover reuse/wrong connected
  peer, closed handle, failed native query and lost BOUND response.
- [ ] Local transport: CreateNamedPipe with duplex, overlapped, first-instance,
  maxInstances1, PIPE_REJECT_REMOTE_CLIENTS; explicit security descriptor limiting
  this newly created pipe to current ordinary user, plus peer token/elevation
  verification. This is not changing system/user security policy. No broad ACL,
  remote endpoint, second accept or reconnect. Bounded asynchronous frame assembly
  with cancellation no later than remaining deadline; never assume cancel equals
  proof of nondelivery. Test all transport effects through mocks first.
- [ ] Shared clock: Stopwatch/QPC timestamps/frequency, same host, immutable T0;
  negative/decreasing/mismatched frequency fails. Keep timing math overflow-safe.
  Guardian receipt writer cannot block timer thread indefinitely: bounded receipt
  result is required before send; missing completion makes dispatch fail, not late.
- [ ] Hidden net10.0 Windows guardian host: no target launch capability, one
  client/session, ordinary token, exit its own loop at360s. Create private control
  channel before target launch; cancellation and parent loss feed same single
  close machine. No auto-restart or heartbeat-dependent timer renewal.
- [ ] New inert observer entry with `-RunApproved` only. Start-Process's explicit
  child `-Environment` is selected for activation, no ArgumentList/parent env
  mutation. Pin exact launcher as before; actual command line MUST remain quoted
  exe + single U+0020. Native verification of that result is a gate, not assumed
  from docs. Add fixtures forbidding launch when helper/identity/policy gates fail.
- [ ] Application activation: retain original MainWindow object, connect only for
  complete validated bootstrap, erase process bootstrap secret env after parsing,
  keep secret lifetime bounded. Pipe work off UI thread; DispatcherQueue.TryEnqueue
  one callback, recheck<300s and original object then Window.Close(). Preserve all
  normal close handlers; reject active unsupported modality rather than dismiss
  it. No Environment.Exit/kill/WM_CLOSE/alternate window fallback. Test enqueue
  rejection/delay/reentrancy/object replacement with fake adapter, not real window.
- [ ] New `OrdinaryBuildVariant=cooperative-close` in isolated App-only
  `obj/ordinary-cooperative-close/`, `bin/ordinary-cooperative-close/`; enforce
  Debug, diagnostics false and no variant collisions. Existing account-center
  variant/output remains unchanged. Shared protocol explicitly included in App
  and helper. Pure tests must not pull either native host into execution.
- [ ] Preserve existing observer tests. Add new entry no-start tests mirroring
  existing forbidden launch/CIM/compile/write mocks. Proposed commands after
  authorization: `pwsh -NoProfile -File ./clients/windows/tests/guardian-coordinator.tests.ps1`
  and `pwsh -NoProfile -File ./clients/windows/tests/observe-ordinary-cooperative-close.tests.ps1`.
  Use verified absolute pwsh if PATH missing; exit0 and explicit check counts.

## Phases3–5 — hold points, NOT currently executable approvals

3. Isolated build/static check: inventory installed locked SDK/dependencies; freeze
   approved source commit; prepare exact local locked restore/build command sheet
   for new helper and new App variant before approval. One authorized attempt,
   failure stops. Preserve old outputs; review new source provenance and helper
   hashes plus six App hashes before writing policy. Never auto-pin arbitrary output.
4. Native verification: separately approve fixtures that exercise real pipe peer
   APIs, handle transfer, ordinary token/local-only ACL, command-line/env behavior,
   clock and cancellation. No production App launch or live close implicit. Check
   installed WindowsAppSDK2.6.5 signatures and cancellation/lifecycle semantics;
   online generic docs do not prove local compatibility. Stop on unsupported API.
5. First live approval: one startup, zero retries, initial page only, no keyboard
   coverage. Guardian closes at290s; send/app/UI timestamps<300s, original exit
   verified<360s, correlated final receipts, residual inventory and six hashes.
   If source or UI changes require revised scope, obtain approval before launching.
   Then separately plan dormant-mode and impact-based UI/keyboard/scroll/dialog
   regression on new hashes. Old passes remain historical, not transferred.

## Platform references and interpretation

- [Start-Process](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.management/start-process?view=powershell-7.6)
  documents child environment overrides; does not prove exact observed command-line bytes.
- [CreateNamedPipe](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-createnamedpipea)
  documents first-instance, overlapped and remote-rejection flags; NOT PIPE_NOWAIT as an async substitute.
- [TryEnqueue](https://learn.microsoft.com/en-us/windows/windows-app-sdk/api/winrt/microsoft.ui.dispatching.dispatcherqueue.tryenqueue?view=windows-app-sdk-1.8)
  true means queued, not callback executed.
- [Window.Close](https://learn.microsoft.com/en-us/windows/windows-app-sdk/api/winrt/microsoft.ui.xaml.window.close?view=windows-app-sdk-1.8)
  closes the app window; [Closed](https://learn.microsoft.com/en-us/windows/windows-app-sdk/api/winrt/microsoft.ui.xaml.window.closed?view=windows-app-sdk-1.8)
  is a lifecycle notification, not a substitute for original-process exit evidence.

Self-review: first-phase contracts/tests are actionable without native adapters;
all later live/native operations explicitly gated. No claim that this checklist
implements or verifies the guardian. Recommend in-session phase1 execution with
independent review at its completion; user may choose per-task independent agents.
