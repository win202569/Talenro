# One minimal entry probe — 2026-10-15

## Approved scope and implementation

The user approved one minimal entry probe: write an exact stdout marker, immediately
flush it, then write a separate timestamp file. No explicit child launch, Go/fuzz,
old-probe replay, acceptance upgrade or change to the original two-second budget.

New `entry-startup-probe.ps1` uses the unchanged native runner, existing Windows
PowerShell5.1 runtime and a fresh owned `entry-startup-probe-NNN` identity. Its
three-statement encoded script ends naturally after the file write; it deliberately
has no Start-Process or20s sleep. Thus it is not the original descendant fixture.
Planning is readonly; existing/wrong/outside identities are rejected before launch.
All evidence, result and stopped envelopes keep Acceptance=false,
OriginalTestPassed=false and diagnostic semantics. Cleanup must verify accounting0
and an empty member list. Missing stdout/file evidence is never filled by inference.

## Test-first checks and single observation

The safe initial stub failed the two-second-budget assertion before an observed
process was launched. After implementation,12 checks passed under PowerShell7 and
Windows PowerShell5.1. The generated script is tested in the existing unit-test
host: a real buffered StreamWriter/MemoryStream is read BEFORE disposal, proving
the explicit Flush reached its stream, then its separate file is checked. Console
output is restored in finally. This is not another timed startup measurement or
child/probe launch. Other checks cover encoding roundtrip, readonly/fresh identity
guards, exact stdout-line matching versus suffix text, file-only/both/missing
evidence, and nonacceptance. No mocked process result is promoted to success.

Actual `entry-startup-probe-001` ran ONCE:

- UTC09:55:10.3034784–09:55:10.5689701 (Asia/Tbilisi13:55:10), total recorded
  runner interval265.4917ms, not an isolated startup-duration measurement.
- Stdout contains the exact `TALENRO_ENTRY_PROBE_ENTERED` line. Timestamp file
  contains `2026-10-15T09:55:10.5407494Z`; both observations are true.
- Exit0, TimedOut=false, root exited, natural shutdown verified, no termination
  request; cleanup verified, final accounting0/member list empty. stderr empty.
- Native accounting TotalProcesses2 despite no explicit script child launch.
  This aggregate does not identify the second process or demonstrate readiness
  of the original intended20s child. Do not guess its image or query archived PIDs.
- Acceptance=false and OriginalTestPassed=false. No second entry probe ran.

Readonly readback accepted the saved native process-completion receipt, rechecked
both markers/nonacceptance and refusal to reuse the consumed identity. Raw process
completion is distinct from product or original-fixture acceptance. Fractional UTC
precision was preserved when computing the265.4917ms interval.

Ignored worktree-relative evidence:
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/entry-startup-probe-001/`.
SHA256: process/receipt.json
`4C2117170074C95769684985240A3C46D0CD8D44B56FD35EF40C9C0B7ABF4834`;
result.json `C3E9E50F4162AE0F6FADFBF7D17A7207EFE4106604A2DB7BA3E1FB6BD9031B60`.

## Interpretation and independent review

This observation establishes that this minimal entry program can execute, flush
stdout and complete a file write within the current2s budget in this sample.
It does not establish why the previous longer parent program produced no marker.
Payload, instrumentation, timing and preceding unit execution differ; this is
not a controlled causal A/B comparison. Do not infer a permanent startup/file-access
failure, blame command discovery/scanning/scheduling, or declare the old test fixed.

Independent read-only review found no blocking implementation defect and supported
an unaccepted development checkpoint only. It ran no tests/processes. Its remaining
nonblocking limits are retained: stdout-only evidence and cleanup-refusal branches
are not covered by the unit suite; no apostrophe-containing path fixture was used
(escaping checked statically); FileMarkerObserved means presence/raw text, not a
validated timestamp contract. The actual saved marker text above was read directly.

No full native timing/product suite was rerun in this one-probe scope. The prior
run.tests descendant-scenario failure/16 unrun checks and original origin87 remain
open. The probes do not consume a new Envelope sampling attempt, reopen failed
preparation001, clear matrix3/Pair/C11/C12/Task10/B01/production gates or merge main.

## Next work and recovery

Subsequent separately approved full-parent probe001 is recorded in
[its own report](2026-10-15-full-parent-probe.md): ENTRY/BEFORE_START observed,
AFTER_START/child-entry absent, timeout with verified forced cleanup. This does
not change the minimal-entry observation or clear any earlier failure.

Next diagnosis should examine the full parent path and prove the intended child
actually starts, rather than treating aggregate process count alone as readiness.
Any new probe or fixture change needs its distinct approved scope; do not enlarge
the timeout, repeat until passing, or substitute minimal-entry success for the old
descendant test. Both consumed probe identities must remain intact.

Necessary scripts/tests/reports are saved on the latest
codex/c12-b01-task9-coordinator; remote SHA equality is verified at handoff. Recover
that branch and read current status/repository recovery. Unit fixtures use current
PowerShell and local temporary directories, not old Go copies/caches. Cloning does
not recover raw/live state or authorize rerunning either consumed probe.
