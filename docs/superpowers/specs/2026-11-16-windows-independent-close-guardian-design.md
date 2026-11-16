# Independent normal-close guardian — formal design

Date 2026-11-16. Selected approach: independent guardian, explicitly chosen by
the user. DESIGN ONLY; approval of this document, implementation/testing approval,
and a new single-launch approval are distinct gates. No application or helper is
launched by this design work. Existing live allowance remains0.

Independent read-only builder check completed2026-11-16: no important unanswered
design requirements. Concrete APIs, clock/handle transfer, bounded execution and
window-lifecycle validation remain implementation-plan obligations, not verified
capabilities. This review is not implementation, native validation or user approval.

## Intent and limits

The interrupted run d2a50105f2ff463d856c166e0bc69068 reached its observation
deadline without a normal-close request. Current observation cleanup disposes
resources, not the application. Remove dependence on continued chat execution
for requesting normal close; retain original-process exit evidence and safety.

Guarantee sought under a functioning host: target-bound, at-most-once normal
close request scheduled at elapsed290s, issued before300s, observation ends by
360s. Request delivery is NOT proof of application exit. OS suspension, crashes,
scheduler stalls, identity uncertainty and API failures cannot be turned into
successful guarantees. Record lateness/unknown outcomes; never silently extend
deadlines or force termination. No claim of real-time scheduling guarantees.

## Approaches and decision

- Add closure inside observer: smaller change, survives chat loss but shares
  observer stalls/termination. Rejected for this requirement.
- Independent local guardian: selected. Own timer, retained target identity,
  bounded close attempt and independent records; additional handshake and
  lifecycle complexity. Does not change the candidate application binary.
- Human timer: optional practical backup, not the selected automatic guarantee.

## Components and ownership

1. Coordinator/entry keeps existing preflight, candidate launch and original
   process observation. It starts ONE hidden ordinary-token guardian before
   target launch. No service, scheduled task, elevated helper, security-setting
   modification, or new dependency installation.
2. Guardian is sole owner of programmatic normal-close dispatch. It runs in a
   separate process and does not need chat polling, heartbeat renewal or observer
   progress to reach its timer. The target must not belong to a kill-on-parent-
   exit arrangement. Process/window APIs and process creation behavior must be
   checked against primary platform documentation in implementation planning.
3. Coordinator and guardian use a private local authenticated channel bound to
   this invocation (prefer narrowly inherited pipe handles, not a public named
   endpoint or arbitrary command-file reader). No executable/script text over
   the channel. Fixed typed messages, length/schema/version validation, one RunId,
   fresh session nonce, expected sender identity and monotonic message sequence.
   These protect accidental/cross-run confusion, not malicious same-user control.
4. Observer owns original retained process evidence and final acceptance.
   Guardian records request evidence, not a replacement exit verdict. Each
   process writes its own append-only receipt, avoiding concurrent writer races.

## Startup and exact binding

- Preserve clean authorized branch, unchanged approved candidate source and all
  six hashes, ordinary token, launcher version/hash, exact command line including
  terminal U+0020, path/reparse, zero matching target/prototype and CIM gates.
- Extend checks to guardian source/build provenance, hash, ordinary token and
  exact instance identity. It must not launch the candidate or discover arbitrary
  targets. Revised observer/guardian source pins must be explicit and reviewed;
  never bypass checks to accommodate the change. Existing candidate hashes stay.
- Guardian emits READY with policy digest and timing capability BEFORE app launch.
  Failure/timeout means no launch, no retry. Prelaunch handshake budget10s;
  orphan guardian without target assignment exits after10s, without app actions.
- T0 is captured immediately before the sole candidate launch attempt. Share an
  immutable machine-local monotonic deadline origin with guardian; validate clock
  comparability in planning/tests. UTC is audit display only. No timer reset on
  reconnect, message delay, window discovery or parent loss.
- Transfer a retained handle to the original process plus expected PID, creation
  time, path, exact command line, RunId and candidate policy digest. Guardian
  independently validates original native creation identity and CIM evidence
  with existing precision rules; PID/title alone is never sufficient.
- Require BOUND acknowledgement by T0+10s before any interactive test. Window
  association requires exactly one intended top-level application window owned
  by that verified process; no unrelated/modal fallback. Track its owner and
  lifecycle, not merely its numeric handle. Before dispatch recheck original
  process alive/creation identity, window ownership and valid association.
- If target launch succeeds but binding fails, no test may begin. Record failure;
  request human normal closure if automation lacks a trustworthy target. Never
  manufacture binding to close it. The launch-to-binding race is an explicit
  residual risk, not a claim of total coverage. A valid guardian binding remains
  armed even if the parent loses the acknowledgement.

## Lifecycle and single close authority

States: READY -> BOUND/ARMED -> CLAIMED -> REQUEST_RECORDED -> FINISHED;
failure and already-exited outcomes are terminal. Guardian serializes all timer,
early-stop and channel-loss events through one decision point.

- At290s, claim the sole close attempt if the bound original is still alive.
  All identity work needed for that decision must be bounded; no waiting for chat.
- Early normal completion or acknowledged cancellation requests CLOSE_EARLY
  through the private channel. Parent/channel loss after binding also requests
  early normal close. In either case the guardian remains sole dispatcher.
- Persist the claim before attempting the normal window-close operation. If
  persistence fails, do not dispatch and report safety failure. If claim exists
  but delivery evidence is lost, record UNKNOWN; never replay after restart.
- Use one bounded, non-forcing normal-close request addressed only to the bound
  window, not button activation, keyboard injection, title search, broadcast,
  process kill or process-tree kill. Select the exact Windows API and race-safe
  ownership validation in the implementation plan; verify supported semantics
  before coding. No automatic fallback API or second window/request.
- Distinguish CLAIMED, REQUEST_ISSUED, API_RETURNED and EXIT_VERIFIED. API success
  means only that API's documented result, not that the app processed or obeyed
  the request. API failure/uncertain completion consumes the attempt too.
- If already exited before claim, record ALREADY_EXITED and send no request.
  If identity/window association is uncertain, send nothing: SAFETY_BLOCKED.
- Do not dispatch at or after300s. Late scheduling records REQUEST_DEADLINE_MISSED,
  continues only bounded read-only observation to360s, and never makes a late
  request look compliant. App may remain alive; report this without killing it.
- Guardian exits itself by360s, releases its resources and never kills the app.
  Observer retains its existing360s limit and no-retry behavior. Blocking platform
  calls cannot be assumed preemptible; actual overruns must be recorded, not
  reported as a hard cap achieved. Host failure may prevent any final record.

## Interaction protocol changes needed for future approval

Keep one numbered phone step outstanding, full60s acknowledgement reservation,
fresh read-only focus capture and no agent input during the keyboard block.
Because guardian may close at290s, new prompt cutoff becomes210s (60s response
plus20s observation reserve). No new key instruction after210s; no pending input
may be revived by late replies. Existing220s-cutoff live plans are NOT reusable
unchanged with this design. Guardian closure always wins over an unfinished test.

At successful early completion, coordinator asks guardian to close instead of
clicking X itself; no dual automatic close owners. Inform user at startup of the
automatic deadline and to stop typing if window closes. Unexpected target/focus,
unreliable image or missing acknowledgement ends coverage; no extra key or retry.
Manual emergency closure is reported as manual, never credited to guardian.

## Results, evidence and acceptance

Record RunId/policy digests, READY/BOUND, original process creation identity,
window association, immutable T0/deadlines, each state timestamp, trigger reason,
claim count, attempted dispatch count, API outcome, lateness, channel loss and
helper finalization. No credentials, window content dumps or arbitrary paths in
portable reports. Preserve raw receipts locally; sanitized records in Git.

Separate verdicts: focus coverage, guardian close-request compliance, and original
process exit verification. Full run acceptance requires all relevant checks,
at-most-one request issued before300s (or independently verified earlier exit),
verified original-process exit within360s, receipt/finalization integrity and
unchanged six candidate hashes. Missing guardian evidence cannot be replaced by
current process absence; a late exit never repairs a missed request deadline.

## Required validation before any new live run

Synthetic tests first, with no target app/native close side effects:

- Missing/forged/late READY prevents launch; malformed/replayed/wrong-run messages.
- Launch-to-binding failure, lost BOUND reply, parent loss before/after binding.
- PID reuse, creation/path/command-line/hash mismatch, invalid/ambiguous/replaced
  window, original process already exited; no unsafe dispatch.
- Timer at289/290/299/300/360s; delayed scheduler, backward/invalid clock,
  sleep/resume discontinuity, no deadline reset or late request.
- Early close and timer race, duplicate cancellation, channel loss, receipt-write
  failure, API failure/hang/unknown outcome; at most one claim/dispatch, no retry.
- Chat silent while guardian progresses; observer stalled/exited while guardian
  remains armed; guardian death causes missing compliance evidence, not success.
- Exit agreement/disagreement, late exit, helper cleanup failures, raw provisional
  versus final results; existing observer safety/identity tests remain intact.
- Phone cutoff210s and late acknowledgements cannot authorize extra keys.

Then separately approve isolated helper build/static inspection and any native
verification. Do not replace the candidate or reuse past live approval. First
future live test should exercise guardian closure with no keyboard coverage,
one launch/zero retries, and explicit user approval. Implementation planning must
specify concrete APIs, artifacts and test commands; this document grants none.
