# Fixed C5 observation entry — short design

Status: user approved the written design after f26d509e. Implementation plan awaits
confirmation; implementation and native execution have not started. Existing launch
allowance remains zero.

## Intent and approaches

Replace the untracked, repeatedly assembled C5 observation command with a portable,
testable entry using the existing precision-safe identity guard and resource scope.

- Selected: fixed C5 entry with preserved artifact, arguments and budgets. Small
  scope; another profile later needs an explicitly reviewed change.
- Alternative: generic configurable runner. More reusable, but introduces arbitrary
  executable/arguments, timing policies and extra validation; unnecessary here.
- Continuing ad-hoc commands would leave the previous integration errors without
  regression coverage and is not selected.

## Fixed boundaries

PowerShell7; no new dependency or product-code change. Target the preserved
layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64 output under
clients/windows/Talenro.Demo.App/bin, binding both previously recorded hashes:
EXE4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95;
DLL7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB.
Fixed arguments --layout-test --layout-language=zh-Hans --layout-profile=unknown-usage.
No executable/profile/deadline overrides, rebuilding or replacing missing artifacts.

Normal invocation must not start a process. A clearly named explicit live switch
is necessary but is NOT user authorization by itself; the agent still needs a new
bounded one-launch approval. Require ordinary token, clean expected development
branch, unchanged observer sources, matching artifact hashes and no existing demo/
prototype targets. UI availability/window inventory is the operator's preflight.

## Composition and lifecycle

Use one fixed entry script and a small testable orchestration module beside the
current observation tests. Reuse observation-session-guard.psm1 and the existing
native/managed observation adapters, not copies of their implementations.

Create a fresh correlation and CreateNew raw receipt under ignored artifact output;
never overwrite prior receipts. No logs/binaries/screenshots enter Git. Acquire
writer before the single process acquisition. Put handle acquisition, identity
checks and every subsequent observation inside the existing protected scope.
Factories clean partial acquisition failures. Disposal never kills the target.

Obtain native creation evidence from the original retained SafeHandle first;
require valid API status, exact native/Process creation identity, and the existing
microsecond CIM rule plus exact PID/path/full command-line checks. Do not derive
expected identity fields from observed values or reopen a PID. Unknown evidence
fails closed. Capture the first native identity observation in the receipt even
if later CIM validation fails, without labelling it a verified pair.

Preserve approximately2s native-then-managed sampling. Monotonic observation budget
360s from launch attempt; check remaining budget before actions and bound sleep to
the remainder. No retries or outer process waits. Record actual elapsed time and
any overrun: synchronous OS calls/disposal are not guaranteed hard real-time.
Native signal alone does not imply successful exit: require valid exit time/code0
and record managed disagreements without erasing raw values.

Emit immediate limited start/abort/completion progress separately from buffered
success output so the operator can identify the target if validation aborts.
Include no arbitrary exception text, secrets or unrelated process inventory.
Persist distinct identity-failed, observation-failed, timed-out and exited outcomes
where the writer permits; write failure must remain a failure, never invent a receipt.

The script does NOT automate UI or close the app. The existing C5 operator route
retains startup30s, close request by300s, normal-exit criterion30s and no force-kill.
On observer abort the operator stops functional actions and handles normal close;
the entry must not start a replacement. No real exit verdict from synthetic evidence.

## Synthetic verification and acceptance

Tests exercise the same orchestration with explicitly labelled in-memory fixtures,
fake monotonic time and recording callbacks. Synthetic mode cannot silently fall
back to live adapters, CIM, native compilation, process launch or real sleeps.
Live entry must not expose user-supplied callback/code-string parameters.

Cover: no-start default, preflight rejection before acquisition, one acquisition,
original resource/handle reuse, native-before-CIM validation, precision boundaries,
wrong identities and invalid API evidence, native-before-managed order, bounded
sleep/timeout, failed receipt writes, disposal on every post-acquisition failure,
no success after cleanup failure, valid versus contradictory exit evidence, and
no retry/forced termination. Preserve existing48/32/78 helper checks.

Implementation acceptance is synthetic tests plus code review and portable docs;
it does not grant live acceptance. C5, RTL, shutdown cause and all wider NOT_ACCEPTED
gates remain unchanged. Follow implementation with a separately approved native
session only after the fixed entry has passed these checks.
