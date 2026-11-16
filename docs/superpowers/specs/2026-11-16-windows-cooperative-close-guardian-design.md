# Cooperative application close with independent guardian — revised design

Independent read-only builder check2026-11-16 found no important unanswered
design-level requirements. Native identity/lifetime, bounded I/O/receipt handling,
clock and UI API semantics remain implementation-plan gates, not verified results.

Date2026-11-16. User selected this revision for DESIGN ONLY. No source changes,
build, tests, helper launch or application launch authorized. Live allowance0.
Supersedes the external window-dispatch portion of the
[earlier design](2026-11-16-windows-independent-close-guardian-design.md), not its
safety gates. The [planning blocker](../plans/2026-11-16-windows-guardian-planning-blocker.md)
remains valid for external HWND check-then-dispatch.

## Intent, decision and scope change

Guardian sends one authenticated request over an already bound local connection;
the original application validates it and closes its own retained Window object
on its UI thread. No external HWND-directed message, window-title search, button
activation, injected code or broad process operation. This removes HWND routing
from the close path; it is an architectural resolution, not runtime proof.

Alternatives: external window messages retain the unproved binding race; observer-
only timers share observer failure; user selected independent guardian plus
application cooperation. The cost is an intentional candidate-source change,
new isolated artifacts/hashes and impact-based revalidation. Old binaries/records
remain intact. Prior layout/keyboard passes do not automatically accept new hashes.

## Components and least scope

- Observer retains sole target launch and original-process exit verification.
- Guardian is a separate hidden ordinary-token process, single close-request
  authority, independent of chat/observer progress once armed.
- Application endpoint is dormant by default. It starts only for a complete,
  validated per-run bootstrap in a child-only launch environment; absent bootstrap
  means no pipe, timer or new listener. Partial/invalid bootstrap rejects this
  observation activation, never silently falls back to an unguarded run.
- Endpoint capability is NORMAL_CLOSE only; no commands to delete, logout,
  connect, execute text, read files, modify settings, or manipulate arbitrary UI.
- New candidate contains this dormant endpoint with ordinary UI/diagnostics still
  disabled as before. No instrumentation-only build may stand in for the candidate
  being accepted. Distribution policy outside this local candidate is not changed.
- No service, task scheduler entry, installation, elevated process, security-setting
  change or kill-on-parent-exit job. Original identity/ordinary-token checks remain.

## Private bootstrap and exact peer binding

Observer generates cryptographic per-run random RunId/capability material and
two distinct channels: coordinator-to-guardian control and guardian-to-app close.
The control channel remains private, invocation-bound, fixed-message only.
Guardian owns a single-instance local named-pipe server for the app connection,
created before target launch with first-instance/exclusive semantics, same-user/
same-elevation restrictions, remote-client rejection and asynchronous bounded I/O.
Collision or first invalid connection fails the run; no accept/reconnect retry.
No machine/global environment changes, command-line secrets or token files.

Child-only bootstrap supplies random endpoint locator, capability secret,
protocol version, RunId, expected guardian PID/native creation identity, policy
digest and immutable clock origin. These are private activation data, not receipt
content. Use explicit child environment rather than temporarily mutating parent
environment. Launcher support and exact resulting command line must be validated
in planning; current quoted-exe plus one U+0020 match remains mandatory unless a
separate explicit design change is approved. Never trim or use tolerant matching.

Guardian first returns READY with its verified process/source/artifact identity,
server readiness and policy digest. No target launch before READY; timeout10s,
zero retries. Capture T0 immediately before the sole candidate launch and deliver
it over the private control channel. App launch retains the original process handle.

Both ends authenticate the connected pipe instance, not a subsequently reopened
name: app queries connected server PID; guardian queries connected client PID.
App compares guardian PID plus retained native creation identity against bootstrap.
Guardian compares client PID and retained original-app creation identity against
observer's verified launch evidence; perform existing CIM path, exact command-line
and microsecond creation checks, candidate/helper hashes and policy binding.
Use original retained/duplicated process handles, not unverified PID reopening.
Any acquisition by PID must immediately prove expected creation identity before use.
Check peer still alive through handshake; a reused PID cannot become the old peer.

After OS peer checks, mutual challenge/response proves the capability without
sending its raw value. Use a standard authenticated construction (HMAC-SHA256),
fresh challenges, directional domain labels, constant-time verification, and a
canonical length-delimited transcript covering version/RunId/identities/policy/T0.
Authenticated fixed-schema messages include direction, monotonic sequence and
deadline; max frame4096 bytes, no arbitrary paths/code/serialization type names.
Application allows exactly one session and one close capability consumption.

Require BOUND acknowledgement from app and guardian by T0+10s BEFORE test input.
Binding failure rejects acceptance; no test or identity fallback. Launch-to-binding
gap remains a residual risk: if trustworthy connection was never established,
guardian cannot safely close it. Human normal closure may be needed, never force.
A successful binding remains armed if coordinator loses its acknowledgement.

Threat boundary: reject accidental/cross-run/wrong-peer/replay access and remote
clients; do not claim protection from malicious same-user memory inspection,
administrator compromise or hostile code in the application. CurrentUserOnly
alone is not sufficient peer identity authentication. Pipe identity/lifetime and
secret handling require native validation before any live approval.

## Timer, exactly-once intent and application self-close

All processes use the same validated machine-local monotonic origin; UTC only
for audit. No clock renewal on reply, reconnect, UI readiness or parent loss.

1. At290s, guardian atomically claims the single request if original app is alive
   and bound. Early completion/cancellation or coordinator loss may claim it
   earlier through the same state machine. No second automatic close owner.
2. Persist claim before dispatch; write failure blocks dispatch. Send exactly one
   bounded authenticated CLOSE_NORMAL frame on the existing pipe, never reconnect.
   Partial write, error or missing acknowledgement consumes the attempt and is
   UNKNOWN/FAILED, not permission to retry. Incremental transport reads are frame
   assembly, not repeated requests; enforce fixed size and time bounds.
3. Recheck clock immediately before send; no dispatch at/after300s. Record send
   start and completion separately. Require full frame completion before300s to
   claim request deadline compliance; an in-flight overrun fails that criterion.
4. App receives and authenticates the complete frame before300s, confirms RunId,
   original peer/session/policy, sequence and its own instance, then atomically
   consumes its one-use capability. Invalid/duplicate/late message never closes.
5. Persist APP_ACCEPTED before scheduling UI work. Receipt failure blocks self-
   close and fails the run. Post exactly one callback to the original Window's
   dispatcher; no HWND lookup, no replacement Window, no fallback callback.
6. UI callback checks the original object/lifecycle is still valid, receipt and
   capability state intact, and elapsed<300s immediately before normal Close().
   At/after300s or with unavailable dispatcher/object, record UI_CLOSE_NOT_IN_TIME
   or SAFETY_BLOCKED and do not close later. No repeated callback or force exit.
7. App's normal close handlers may cancel/hang. Keep ordinary lifecycle behavior;
   no Environment.Exit, Process.Kill or bypass of cancellation. Original observer
   alone proves exit and exit code by360s. ACK, APP_ACCEPTED and UI_CLOSE_INVOKED
   are distinct from EXIT_VERIFIED. ACK loss never triggers retry.

No absolute delivery/execution guarantees under OS suspension, scheduling stalls,
deadlocks or crashes. Target290s gives10s request/UI margin but does not prove it.
Every overrun/missing finalization is failure/unknown, never retroactive success.
At360s guardian/observer end their bounded work, close their own resources and
invalidate app endpoint state; no process termination or deferred close execution.
App endpoint uses the same local deadline to reject queued/late callbacks even
if guardian dies. Closing a pipe by itself is NOT authority to self-close: app
requires the authenticated request. Guardian death leaves acceptance blocked.

## UI and authority boundaries

Phone step reservation remains60s; latest prompt210s,20s capture/close reserve
before290s. One outstanding numbered step, fresh evidence, late replies cannot
revive cancelled input. Old220s plans require revision before use. During keyboard
block no agent clicks/keys; guardian's scheduled close preempts the test.
Early finish goes through coordinator->guardian CLOSE_EARLY, not agent titlebar X.
Human emergency X is allowed as a separately recorded action, not credited to
guardian. First new live validation should stay on initial page, no keyboard task
or dialogs, to isolate autonomous close behavior. Dialog/focus testing later needs
explicit modality/cancellation handling coverage, never dismiss dialogs silently.

## Evidence and acceptance

Separate append-only observer/guardian/app receipts, common RunId and policy,
no secrets/challenges/raw authenticated frames in committed reports. Each producer
is responsible for provisional versus finalized state. Record READY/BOUND, peer
identities, T0, claim, frame-send start/completion, app acceptance, UI enqueue/invoke,
reject reason, exit agreement, deadlines and overruns. No shared concurrent writer.
If a write blocks, mark overrun when possible; host failure can leave no final row.

Guardian compliance, app close handling, exit verification and UI coverage are
separate verdicts. Full guarded-close pass requires timely full-frame request and
app acceptance/UI invocation before300s, verified original exit within360s,
complete correlated evidence and unchanged new hashes. Already-exited original
process requires no request, but is NOT_EXERCISED for autonomous close coverage.
Current absence, API success, app ACK or old-candidate passes cannot substitute.

## Implementation and verification gates

Preserve existing gates and tests, old artifacts/raw receipts. New explicit source
and helper pins plus all six new app artifact hashes must be reviewed; no automatic
policy regeneration from whatever binary happens to exist. Isolated output only.
No widening of current command-line binding or weakening of token/CIM/path checks.

First implementation phase, separately authorized: pure activation/protocol/time/
state-machine tests and simulated transports/dispatchers, no native close/app run.
Cover dormant/partial bootstrap, wrong peer/PID reuse, cross-run/sequence/replay,
malformed/oversized/partial frames, invalid MAC, secret nonlogging, connection loss,
READY/BOUND10s limits,290/300/360 boundaries, early/timer races, receipt failures,
missing ACK, duplicate callback, UI queue delay/cancellation, original-window
replacement, guardian/observer death and backward clock/suspend/resume overruns.

Then separately approve isolated native build/static inspection and native peer/
transport/clock verification. Actual tool-supported test commands and concrete
app/UI shutdown API semantics belong in the implementation plan. Any native API
that cannot meet the required identity/lifetime checks blocks progress; no HWND
fallback. One new live approval only after all gates, first no-keyboard test.
Revalidate activation-off behavior, ordinary UI regression and impact-based
keyboard/scroll/dialog evidence on new hashes before expanded acceptance.

## Platform basis, not a safety proof

- [GetNamedPipeClientProcessId](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-getnamedpipeclientprocessid)
  exposes a connected client's PID; PID alone is insufficient, hence retained
  creation identity and capability/session checks above.
- [GetNamedPipeServerProcessId](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-getnamedpipeserverprocessid)
  exposes server PID; actual .NET/native handle compatibility needs verification.
- [PipeOptions](https://learn.microsoft.com/en-us/dotnet/api/system.io.pipes.pipeoptions?view=net-10.0)
  documents same-user/elevation checking and asynchronous mode. Additional local-
  only/first-instance enforcement must be selected explicitly in planning.

This revised design awaits user approval; it authorizes no implementation or run.
