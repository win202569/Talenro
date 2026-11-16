# Cooperative guardian phase2 adapter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or superpowers:subagent-driven-development after explicit approval. Checkboxes are future work, NOT authorization.

**Goal:** Bind the tested pure decisions to narrowly scoped native adapters without weakening identity, deadlines or one-attempt semantics.
**Architecture:** Retain the selected independent guardian plus application-owned normal close. One serialized owner per endpoint drives injected ports; native calls remain behind replaceable boundaries. Observer alone launches the target and certifies its exit.
**Tech Stack:** Existing .NET10/C#, xUnit2.9.3, pinned WindowsAppSDK2.6.5, PowerShell7.6.9; no new third-party dependencies proposed.
**Spec:** [approved cooperative design](../specs/2026-11-16-windows-cooperative-close-guardian-design.md).
**Baseline:**85d4bdaef39b054d21d3d3d3ca28b1b0eb952625; [phase1 evidence](../../roadmap/2026-11-16-guardian-phase1-implementation.md),467 historical passing tests, not rerun for this document.

Status: DESIGN/IMPLEMENTATION CHECKLIST ONLY. No source/project edits, tests,
restore/build, native calls, helper or application launch authorized this turn.
User approved H1 and H2 as designs on2026-11-16, then requested the
[execution checklist](2026-11-16-windows-guardian-phase2-execution-checklist.md).
Design approval does not authorize their implementation or native invocation.
This supplements phase2 of the original plan; phase1 interfaces in source take
precedence over the older plan's byte-array digest signature.

## Global Constraints

- Exact original process/native creation/CIM microsecond identity; path comparison
  OrdinalIgnoreCase, command line Ordinal, including quoted exe plus one U+0020.
- Preserve ordinary token, launcher/source/path/reparse and zero-target gates;
  separately approve/pin helper artifacts and all SIX new App artifacts. No auto-pins.
- READY response <=10000ms before launch; actual T0-to-BOUND <=10000ms. No test input
  before both binding acknowledgements. Lost coordinator acknowledgement never disarms.
- Target290000ms; send completion, app acceptance and own-close invocation <300000ms;
  original exit and finalization <360000ms. Observer wait <=2000ms, never renew T0.
- One claim/request/enqueue/callback; no reconnect/retry, kill, external HWND close,
  Environment.Exit, process termination right, security-policy modification or install.
- Phone reserve60000ms, last new prompt210000ms; Cancel on claim, cancellation,
  binding loss or termination. Stale confirmations never revive authority.
- Preserve old binaries, raw receipts, source pins and historical acceptance.
- Phase2 source/mock tests, isolated build, native verification and live observation
  are separate approval gates. Native evidence cannot be replaced by mocks.

## Approach and approved design decisions

Continue the already chosen cooperative design. External HWND dispatch is excluded
by its identity race; observer-only timing cannot survive observer loss. This is
contract refinement, not selection of a new close architecture.

H1 — **handle transfer rights:** the old plan says every OpenProcess is query/sync
only, but DuplicateHandle needs PROCESS_DUP_HANDLE on source/recipient processes.
Approved design: observer-only, short-lived DUP_HANDLE access to the verified guardian
recipient (observer source is its own process); NEVER to the target App. The
duplicated App handle has only QUERY_LIMITED_INFORMATION|SYNCHRONIZE, non-inheritable,
no SAME_ACCESS or CLOSE_SOURCE. Verify guardian creation before transfer and keep
that recipient handle until completion. This narrow exception is now explicitly
approved at design level only; source implementation and native invocation retain
their separate gates. No wider process rights or silent fallback are authorized.

H2 — **control assignment schema:** v1 frame has no transferred-handle field or
coordinator identity. Do not overload a challenge, PID, path or command line.
Approved design: separate ControlEnvelopeV2, independent control secret/domain and exact
coordinator/guardian binding, while preserving application-frame v1 byte vectors.
Source implementation still requires execution approval (Task1 below).

H3 — **bounded execution reality:** cancel request is not completion. Keep native
overlapped buffers/handles alive until completion; at deadline consume the attempt,
invalidate all future close effects and report UNKNOWN if completion is unproved.
Never free pending buffers to pretend cleanup completed. A host that cannot settle
its own operations by360s fails acceptance; no kill or hard real-time guarantee.

## File ownership

Paths relative to clients/windows; create only after implementation approval.

| Files | Responsibility |
|---|---|
| CloseGuardian.Shared/GuardianAdapterContracts.cs, GuardianHandshake.cs, GuardianControlProtocol.cs | Pure port/result types, stage/sequence progression, approved v2 control envelope |
| CloseGuardian.Windows/PeerIdentity.cs, LocalPipe.cs, GuardianClock.cs, GuardianReceipts.cs | Native evidence, connected transport, monotonic time, append-only receipts |
| CloseGuardian.Shared/GuardianEffectRunner.cs | Testable effect orchestration; no P/Invoke/UI dependency |
| Talenro.CloseGuardian/{Talenro.CloseGuardian.csproj,Program.cs,GuardianHost.cs} | Hidden helper composition only; no target launch capability |
| Talenro.Demo.App/Observation/{CooperativeCloseEndpoint.cs,CooperativeCloseUi.cs} | Dormant activation, retained own Window, one UI dispatch |
| Talenro.Demo.App/{App.xaml.cs,Talenro.Demo.App.csproj}, Directory.Build.props | Explicit links and isolated cooperative-close variant |
| tests/{guardian-coordinator.psm1,observe-ordinary-cooperative-close.ps1} | Sole launcher/exit authority; inert entry without RunApproved |
| Talenro.Demo.Tests/Guardian*Tests.cs; tests/{guardian-coordinator.tests.ps1,observe-ordinary-cooperative-close.tests.ps1} | Managed fake-port and no-start tests |

No helper/App ProjectReference in Tests. Native source compile validation, XAML
compilation and installed SDK compatibility remain later isolated-build gates.

## Typed boundary contracts (proposed signatures, not existing APIs)

Namespace Talenro.Observation.CloseGuardian. Shared adapter contracts define:
`AdapterStatus { Pending, Success, Rejected, TimedOut, Unknown, Failed }`;
`AdapterResult<T>(AdapterStatus Status, T? Value, string ErrorCode)`.
Errors are fixed codes; never raw secret/frame/environment/exception dumps.
`PeerRole { Coordinator, Guardian, Application }`.
`PipeLease` and `ProcessLease` are owned IDisposable wrappers with internal native
handles; each has an internal-constructor identity/instance token, no user-set handle.
ProcessLease owns the retained original handle; disposal never terminates a process.
`VerifiedPeer(ProcessIdentity Identity, IdentityEvidence Evidence, PeerRole Role)`
is constructible only after every native/token/CIM/policy check passes.
`BoundSession(GuardianPolicy Policy, ProcessIdentity Guardian, ProcessIdentity App,
 string GuardianChallenge, string AppChallenge, ulong NextSend, ulong NextReceive)`
is immutable; construction restricted to completed handshake, not a public boolean.
Intermediate handshake replies return Pending with no BoundSession; only the
terminal successful handshake returns Success with a value. Ports never use
Success/null as verification. Error values are absent.

- `IPeerIdentity.VerifyAsync(PipeLease connected, ProcessLease retained,
  ProcessIdentity expected, PeerRole role, long deadlineMs, CancellationToken ct)
  -> ValueTask<AdapterResult<VerifiedPeer>>`: connected client/server PID, retained
  native creation, liveness, same ordinary user/session/elevation, exact CIM fields;
  one query attempt, timeout/failure blocks. No name reopen or PID-only acceptance.
- `IHandleTransfer.DuplicateOriginal(ProcessLease app, ProcessLease guardian)
  -> AdapterResult<ulong>`: H1 only; resulting value belongs to recipient table.
  Receiver adopts once, validates handle type/PID/creation/access, owns disposal.
  Sender never CloseHandle on remote numeric value; observer retains original.
- `IGuardianClock.ElapsedMs() -> AdapterResult<long>`: immutable validated T0/frequency;
  checked/Int128 delta conversion, negative/decreasing/mismatched clocks fail.
  Actual launch evidence arriving at500ms remains500ms, never zero.
- `ILocalPipe.AcceptOnceAsync(long deadlineMs, CancellationToken ct)
  -> ValueTask<AdapterResult<PipeLease>>`: exclusive local server, one connection,
  first invalid peer fails session, never accepts another. Creator uses duplex,
  overlapped, first-instance, maxInstances1, remote-reject and explicit current-user
  ACL; peer ordinary-token validation is additional, not replaced by ACL.
- `IFrameTransport.ReadOneAsync(PipeLease pipe, long deadlineMs, CancellationToken ct)
  -> ValueTask<AdapterResult<byte[]>>`: bounded4096 total, read4 length first, strict
  remaining count, one assembly, EOF/extra invalid stage fail; no resynchronization.
- `IFrameTransport.SendOnceAsync(PipeLease pipe, ReadOnlyMemory<byte> frame,
  long deadlineMs, CancellationToken ct) -> ValueTask<AdapterResult<SendReceipt>>`;
  `SendReceipt(long StartedMs,long CompletedMs,int BytesWritten,bool FullFrame)`.
  Partial write is terminal UNKNOWN/FAILED; do not resend remainder as a new request.
  Full completion at300000 fails regardless of earlier start or ACK.
- `IReceiptSink.AppendAsync(ReceiptRow row,long deadlineMs,CancellationToken ct)
  -> ValueTask<AdapterResult<ReceiptStamp>>`; `ReceiptStamp(long CompletedMs)`.
  `ReceiptRow(Guid RunId,string PolicyDigest,PeerRole Producer,ulong Sequence,
  string Event,long ElapsedMs,string Outcome,ReceiptEvidence Evidence)`.
  `ReceiptEvidence(ProcessIdentity? Subject,IdentityEvidence? Identity,long? T0Ticks,
  long? Frequency,CloseReason? RequestReason,int? ExitCode,bool? OriginalExitVerified,
  string? ArtifactManifestDigest,long? WindowGeneration,bool Finalized)`.
  Event/Outcome fixed allowlists; per-event required/forbidden fields validated:
  context includes verified subject,T0/frequency and manifest digest; claim/send/
  accepted/invoked include request reason; identity events include native/CIM
  evidence; exit includes original identity/exit code/verification. Final summary
  includes finalization and matching manifest digest. Unknown/missing/contradictory
  fields fail evidence correlation; no arbitrary payload or exception text. Receipt
  data contain no keys/challenges/authenticated raw frames/remote handle values.
  One append-only producer file; write+flush acknowledgement before dependent
  effect; queue overflow, blocked/failing flush prohibit send/enqueue. No shared writer.
- `IOwnWindow.TryEnqueue(Action callback) -> bool`, `IsOriginalAlive() -> bool`,
  `HasUnsupportedModal() -> bool`, `CloseOriginal() -> void`. Concrete wrapper
  retains actual Window reference and lifecycle flag, not HWND/title lookup.
  All three checks and normal CloseOriginal occur on original UI thread.
- `GuardianHost.RunAsync(CancellationToken ct) -> Task<AdapterStatus>` consumes
  these ports; Program has no target-start method. One loop owns state/sequence;
  clock progress independent of control reader and receipt I/O waits.

Cancellation is NOT an authorization token or evidence of nondelivery. Late port
completion is observed for cleanup/audit but cannot issue a new effect. Phase2
must wrap public phase1 bool events; UI/coordinator input cannot set Verified=true.

## Control and application handshake

Observer creates the private control server before starting its single hidden
guardian child; guardian creates the application server before READY. Both are
one-instance, local and ordinary-token checked. Guardian pins coordinator from
child-only bootstrap; observer pins guardian from its retained launch handle.
Expected paths/command lines/hashes come from reviewed fixed policy, never from
the peer's message. Policy digest is SHA256 of the reviewed canonical manifest
(schema/order, source commit/pins, launcher pins, helper pins, six App pins, exact
paths/command profiles, deadline constants); per-run T0/identities are separately
MAC-covered. Native build must freeze the manifest before live approval.

Approved ControlEnvelopeV2 canonical order: uint32 body length, length-prefixed
UTF8 label `Talenro.CloseGuardian.Control.v2`, byte version2, byte direction,
byte kind, RunId32 ASCII hex, uint64 sequence, digest32, int64 T0/frequency,
coordinator identity, guardian identity, app identity (same field order as v1),
uint64 recipientAppHandle, int64 deadline300000, byte reason, MAC32. All scalar
integers little-endian; text uint32 byte length; total <=4096; strict UTF8, no tail.
Direction/kind reuse named control enum values, no arbitrary payload. Key32 is
separate from app capability; connected coordinator identity verified out of band
before MAC exchange. No challenge placeholders in control-v2. Unknown version fails.
READY: app unassigned, handle0,T0=0. ASSIGN_TARGET: real app/T0, handle nonzero,
strictly one adoption. Other control messages: handle0; no transfer duplication.
READY seq1; ASSIGN_TARGET coordinator seq1; guardian ACK seq2 means assignment
adopted, NOT bound; guardian ACK seq3 means app BOUND confirmed. CLOSE_EARLY or
ABORT coordinator seq2 after assignment, at most one; no late mutation of policy or target.
Before assignment ABORT is coordinator seq1 with unassigned app/T0=0; it is mutually
exclusive with ASSIGN_TARGET and terminates control state. After assignment use
seq2 and real tuple. No skip in directional sequence and no assignment after abort.
If transfer succeeds but assignment delivery fails, record unknown adoption;
never send it again or close an unverified remote numeric handle. Guardian's normal
own-process cleanup eventually releases unadopted handles; missing cleanup proof
blocks acceptance, not permission to kill or extend deadlines.

App v1 remains HELLO A->G seq1 (app fresh nonzero challenge, guardian zeros),
CHALLENGE G->A seq1 (both), PROOF A->G seq2, BOUND G->A seq2, BOUND A->G seq3,
CLOSE_NORMAL G->A seq3, ACK A->G seq4. Reject terminates at the expected sequence.
Handshake tracks kind progression AND directional counters; do not instantiate
fresh untracked sessions to bypass replay checks. All post-assignment fields pinned.
Guardian arms after authenticated app BOUND; failure delivering control ACK seq3
does not disarm/relaunch. Coordinator prohibits prompts until that ACK is verified.
App endpoint authenticates its own actual identity before HELLO, verifies connected
guardian and all pins before BOUND. App capability absent: no pipe/timer/listener.
Partial bootstrap: explicit rejected observation, no unguarded fallback.

## Effect execution and evidence

Capture T0 directly before the sole approved target launch, construct child-only
environment and retain original launch handle. Never mutate parent environment;
erase child bootstrap after parsing and clear owned key bytes on endpoint completion.
Control bootstrap additionally pins coordinator PID/creation and separate key;
never put keys on command line or disk. Guardian child gets no launch authority.
Receipts live in a new validated cooperative-close-observations/<RunId> directory,
with observer.jsonl/guardian.jsonl/app.jsonl distinct exclusive-create owners.
Root derived from reviewed candidate policy; RunId exact lowercase32 hex, no
user path override, no reparse ancestors, no reuse/overwrite of an old run.
Unrelated child processes must not inherit bootstrap: remove keys before any
ordinary application child-start path; prove this in later native dormant tests.

Before ClaimWritten->SendClose and again directly before physical send, verify
same retained target, connected instance and clock<300000. Receipt time consumes
the original budget. Timer claims at290000 (or records scheduling overrun); it
does not block behind control reads. Parent loss/early use the SAME one-use claim.
At300000 invalidate unsent/queued close; at360000 finish own bounded observation.
At UI callback execution recheck live original object, no unsupported modal and
clock<300000 immediately before normal close. No deferred continuation after timeout.

Observer verifies original handle signaled plus exit code and matching identity;
process enumeration absence/ACK/UI receipt is insufficient. Build CloseEvidence
only from correlated finalized producer rows and post-run hash verification.
Preserve request reason: Early/ParentLost/manual paths cannot be labeled scheduled
autonomous290s coverage. App invocation, exit, deadline and UI coverage are separate.

## Review Focus

1. Remote handle numeric reuse/failed adoption never closes another object (Task2).
2. I/O completion racing cancellation never frees pending memory or retries (Task3).
3. Lost final BOUND control acknowledgement preserves guardian arming (Task1/4).
4. Immediate UI callback and concurrently lost guardian never double-close (Task5).
5. Default-off import/build environment cannot inherit activation or launch (Task6).

## Implementation tasks — all pending approval

Each task uses RED->minimal implementation->GREEN->explicit-file commit; an
unexpected infrastructure failure stops for diagnosis, not timeout changes.

### Task1 — pure handshake/control-v2 and early-reason closure

Files: Shared GuardianControlProtocol.cs/GuardianHandshake.cs/GuardianAdapterContracts.cs;
Tests GuardianControlProtocolTests.cs/GuardianHandshakeTests.cs/GuardianSessionTests.cs.
Interfaces: `ControlProtocol.Encode(ControlEnvelopeV2,ReadOnlySpan<byte>) -> byte[]`,
`Decode(ReadOnlySpan<byte>,ReadOnlySpan<byte>) -> ControlEnvelopeV2`; envelope fields
exactly as above. `GuardianHandshake.Receive(byte[] frame,long elapsedMs) ->
AdapterResult<BoundSession>` owns pinned context/key/counters, no native calls.
- [ ] Write fixtures: `ReadyBeforeTargetExists`, `AssignHandleOnce`,
  `WrongCoordinatorOrMacRejects`, `LostBoundAckRemainsArmed`, `SequenceResetRejects`;
  assert absent marker only before assignment, <=4096, tamper/truncate/replay reject.
- [ ] Add `EarlyReasonSurvivesWire` at10000ms and `ParentLossSurvivesWire` at10001ms;
  assert RequestReason equals decoded reason and scheduled coverage NOT_EXERCISED.
- [ ] Run filtered managed command below; expected intended missing-contract RED.
- [ ] Implement only approved H2 contract; keep v1 independent vector unchanged.
- [ ] Run filtered and full managed suites; expected exit0,0 failures; commit files.

### Task2 — identity and transfer ownership

Files: Windows PeerIdentity.cs; Tests GuardianPeerIdentityTests.cs/GuardianHandleTests.cs.
Consumes Task1 leases/results; produces IPeerIdentity/IHandleTransfer implementations.
- [ ] Tests `WrongConnectedPid`, `NativeCreationDiffersByOne`, `CimMicrosecondMismatch`,
  `TrailingSpaceRemoved`, `ElevatedPeer`, `RecipientReuse`, `AdoptTwice` all reject;
  `TransferFailsKeepsOriginal`, `LateAdoptionNeverClosesForeignHandle` assert ownership.
- [ ] RED using fake native gateway, never actual P/Invoke.
- [ ] Implement declared methods after H1 approval; access masks/handle cleanup
  explicit; query-token access fails closed, never privilege adjustment.
- [ ] GREEN mock suite, static no-terminate/no-broad-access audit; commit.

### Task3 — bounded clock, transport and receipts

Files: Windows LocalPipe.cs/GuardianClock.cs/GuardianReceipts.cs;
Tests GuardianTransportTests.cs/GuardianClockTests.cs/GuardianReceiptTests.cs.
Consumes Task1 result types; produces ILocalPipe/IFrameTransport/IGuardianClock/IReceiptSink.
- [ ] RED cases:4096/4097, partial prefix/body, second accept, malformed ACL config,
  wrong frequency/backward/overflow; send completion299999 passes,300000 fails.
  `CancelRaceKeepsBufferOwned`, `BlockedFlushNeverSends`, `CompletionAfterCapNoEffect`.
- [ ] Implement behind fake gateway; absolute deadlines and single outstanding
  request, no synchronous waits on the timer owner. No default production gateway in tests.
- [ ] GREEN including never-completing fake I/O and late completion; commit.

### Task4 — helper/effect runner and coordinator

Files: Shared GuardianEffectRunner.cs; helper project/Program/GuardianHost.cs;
tests guardian-coordinator.psm1/.tests.ps1; Tests GuardianEffectRunnerTests.cs.
Consumes Tasks1–3; `GuardianEffectRunner.RunAsync(CancellationToken) -> Task<AdapterStatus>`
uses injected ports. `Invoke-GuardianCoordinator -Ports <hashtable> -Policy <hashtable>`
returns structured provisional/final evidence, exactly one launch intent.
- [ ] RED: observer loss, blocked receipt while timer advances, delayed launch500ms,
  binding10001ms, missing READY, ACK loss, once-only launch, phone Cancel on every
  close/fault/end path; no action from stale phone reply.
- [ ] Implement state/effect mapping; helper cannot launch target; Finish disposes
  own resources only. Mock host entry; do not build or execute helper.
- [ ] GREEN managed/no-start PowerShell suites; commit explicit files.

### Task5 — dormant application endpoint and own-window bridge

Files: App Observation files/App.xaml.cs; Tests GuardianEndpointTests.cs/GuardianUiTests.cs.
`CooperativeCloseEndpoint.TryActivate(IReadOnlyDictionary<string,string?> env,
 IOwnWindow window) -> ActivationResult`; dependencies injected via constructor.
`CooperativeCloseUi` implements IOwnWindow over retained main Window only.
- [ ] RED: no bootstrap yields0 resource intents; partial rejects; child secret
  erasure, callback-before-enqueue-return, duplicate callback, replacement/closed
  object/modal rejects, guardian lost, receipt fault, callback300000 never closes.
- [ ] Implement inert default hook and UI bridge; no dismissal/modal workaround;
  extract fakeable endpoint logic for managed tests, not WinUI instantiation.
- [ ] GREEN fake tests and source audit; WinUI compilation remains unverified
  until separate isolated build approval. Commit, do not claim native compatibility.

### Task6 — isolated variant and inert observation entry

Files: Directory.Build.props/App csproj; tests observe-ordinary-cooperative-close.ps1/.tests.ps1;
Tests OrdinaryBuildIsolationTests.cs; roadmap/recovery documents.
- [ ] RED property/static tests: cooperative-close Debug, diagnostics all false,
  dedicated obj/bin; no collisions with account-center/geometry/close-stage modes.
  Import/no RunApproved => NotStarted with0 launch/CIM/receipt/native/build calls.
- [ ] Implement explicit links/new fixed policy entry, no CLI hash/path/time overrides;
  preserve old entry/pins. Policy remains NOT_READY until separately approved hashes.
- [ ] GREEN managed and mock PowerShell commands; no App target Build/Restore.
- [ ] Independent final source review; regression-fix Important findings; run all
  scoped managed/mock suites, update recovery, commit/push ordinary dev checkpoint.
- [ ] STOP: request isolated native build plan/approval, not a live run.

## Future verification commands (NOT run for this document)

From the existing worktree, verified ordinary-token PowerShell host:

```powershell
dotnet test ./clients/windows/Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore --filter 'FullyQualifiedName~Guardian'
dotnet test ./clients/windows/Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore
pwsh -NoProfile -File ./clients/windows/tests/guardian-coordinator.tests.ps1
pwsh -NoProfile -File ./clients/windows/tests/observe-ordinary-cooperative-close.tests.ps1
git diff --check
```

PowerShell suites must inject fake ports before importing any implementation;
all production native/launch/build commands forbidden by mocks. Inspect tests
before execution. Missing dependencies or pwsh path is a blocker, not installation
permission. Expected GREEN exit0 and explicit pass counts, never inferred from silence.
Compile-only native validation and real pipe/handle/UI tests require later approval.

## Self-review and authority

Mapped native identity/ownership->Task2, protocol/handshake->Task1, I/O/clock->Task3,
independent progress/phone->Task4, dormant own-window->Task5, pins/isolation->Task6.
H1/H2 design approval is recorded above; this is not an executable authorization.
H3 is an acceptance limitation, never a promise to forcibly end a stalled process.
Recommend current-session implementation plus one independent final review after
approval, preserving the user's phase1 workflow. No implementation begins now.
Bounded independent builder check identified two gaps: receipt evidence fields
and preassignment ABORT sequence. Both are now specified above and self-checked
for field/sequence consistency; not a native audit or an implementation review.
No H1/H2 design decision remains pending. Follow the execution checklist for the
next approval boundary; do not infer source/test/native execution permission.

## Primary-source checks, not local compatibility proof

- [DuplicateHandle](https://learn.microsoft.com/en-us/windows/win32/api/handleapi/nf-handleapi-duplicatehandle):
  source/recipient require DUP_HANDLE; remote value belongs to recipient process.
- [CancelIoEx](https://learn.microsoft.com/en-us/windows/win32/api/ioapiset/nf-ioapiset-cancelioex):
  request is asynchronous, completion can win; pending structures must remain valid.
- [CreateNamedPipe](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-createnamedpipea):
  first-instance and overlapped controls; nonblocking mode is not async completion.
Checked2026-11-16. Installed WinUI/native signature compatibility still needs
separately approved local compilation/native verification.
