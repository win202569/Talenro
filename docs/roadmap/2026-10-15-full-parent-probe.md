# One full-parent stdout-stage probe — 2026-10-15

## Approved scope

The user approved one fresh complete-parent probe with synchronously flushed
stdout stages at entry, before Start-Process and after it returns. The original
hidden child launch, parent/child20s sleep and native whole-process2s budget remain
unchanged. No security, runner, production or acceptance policy was changed.
No Go/fuzz, preparation replay or repeat-until-pass was authorized or performed.

The existing tree-timeout-probe helper now has an explicit SynchronousStdout mode.
It accepts only a new owned full-parent-probe-NNN identity; default mode still
accepts only tree-timeout-probe-NNN. Planning launches nothing. The parent captures
Start-Process's success flag before emitting the return marker. Evidence retains
only exact observed stdout-stage lines and the original four possible stage files.
Acceptance and OriginalTestPassed remain false, even with every marker present.

## Verification and one actual observation

Test-first: the new suite failed at the explicit-mode identity/budget assertion
before implementation. After implementation,12 full-parent checks,9 existing
tree-stage checks and12 minimal-entry checks passed separately under PowerShell7
and Windows PowerShell5.1 (33 each). These tests do not launch a valid timed probe.
The new buffered-stream test executes only the first two parent AST statements
in its existing host and reads bytes before disposal to verify the entry Flush.
Other checks cover parsing/encoding, fresh separate identities, duplicate refusal,
missing/exact/suffix stdout evidence and nonacceptance. This is not a rerun of the
failed full runner suite or evidence of original-fixture acceptance.

Actual full-parent-probe-001 ran ONCE:

- UTC10:05:09.0923318–10:05:11.1320874 (Asia/Tbilisi14:05).
- Exact stdout stages: TALENRO_FULL_PARENT_ENTRY, then
  TALENRO_FULL_PARENT_BEFORE_START. AFTER_START is absent.
- parent-entry.utc: 2026-10-15T10:05:10.5795654Z;
  parent-before-start.utc: 2026-10-15T10:05:10.5845634Z.
  parent-after-start.json and child-entry.utc are absent.
- TimedOut=true, root not exited, TotalProcesses3. Saved native observations
  show three owned/running members, with two powershell.exe images and one
  conhost.exe image. Their initial OpenProcess/IsProcessInJob calls succeeded.
  These are saved observations, not new queries of historical PIDs.
- Forced owned-job termination requested; cleanup verified, final accounting0
  and member list empty. NaturalShutdownVerified=false. Returned exit57345 is
  not a natural success exit. Stdout61 bytes; stderr11 bytes, only '#< CLIXML'
  header, not an explanatory exception or proof of no failure.
- Acceptance=false, OriginalTestPassed=false. No second full-parent probe ran.

Ignored worktree-relative evidence:
.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/full-parent-probe-001/.
SHA256 process/receipt.json:
502ABD8F3E580EC8843EE395FB0C77FE528E976AC18C25D5256836470450780C;
result.json:
84EF057EF541ADD8FAE62AAC27B6BF3DAEDB106C14923E3C01160531C23F0114.

## Interpretation and remaining work

Independent read-only review found no Critical/Important defect and supported
only an unfinished development checkpoint. It ran no tests or probes. Minor
coverage limits remain: the units exercise ENTRY's real Flush but not structural
assertions for BEFORE/AFTER Flush, immediate success-flag capture or preserved
child invocation; these were inspected statically. Cleanup-refusal is not tested.
The reviewer declined causal attribution, process identity/readiness from count
alone, original-test success, unauthorized replay/policy changes and the unchanged
stopped-artifact schema. Those limits are retained. Image names above come from
the saved member image records, not aggregate counts or later PID queries; they
still do not establish intended-child readiness or explain the old unknown member.

This parent reached and flushed the pre-launch marker, unlike the earlier
no-marker sample. There is no observed completion of Start-Process or child-script
entry. Native process presence does not prove the intended child's script is ready.
The gap remains between the pre-launch marker and the missing return marker;
it includes launch execution and any intervening scheduling. Do not conclude a
deadlock, permanent command failure, startup latency cause or repaired old test.
Different payload/instrumentation/timing is not a controlled causal A/B comparison.

The original origin87, descendant-timeout regression/16 unrun checks, Envelope
sampling and Pair remain unresolved. C11/C12/Task10/B01/production/main gates stay
open. This diagnostic does not authorize new Go sampling or relaxation of gates.
Any next actual probe, fixture change or timing experiment needs separate scope;
do not enlarge the budget or reuse any consumed diagnostic identity.

## Recovery

Recover the latest codex/c12-b01-task9-coordinator development checkpoint, not
automatically main; read current-status.md and repository-recovery.md. Necessary
source/tests and this public summary are tracked. Raw receipts, caches and live
state are not committed or restored by cloning. Unit tests generate fresh owned
fixtures from tracked scripts and require no previous Go cache or raw receipt.
The consumed full-parent-probe-001 identity must not be replayed after recovery.
