# Native launch timing diagnostic — 2026-10-15

## Approved scope and implementation

The user approved UTC/monotonic observations before and after process creation,
owned-job assignment and thread resume; helper checks followed by one fresh actual
sample. No API reordering, timeout/security/acceptance change, retry or Go/fuzz.

process-native.cs now records six LaunchBoundaries: Before/AfterCreateProcess,
Before/AfterAssignProcessToJobObject and Before/AfterResumeThread. Each record has
Name, UTC, absolute Stopwatch Timestamp, original budget-watch ElapsedTicks and
Frequency. run.ps1 exports this additive diagnostic field into the saved receipt.
The existing stopwatch is not reset or extended; native arguments, rights, handle
lifecycle, assignment and cleanup predicates remain unchanged. Original Check calls
still run immediately after each API, before after-boundary recording; assignment
still sets assigned=true immediately upon success. No acceptance code consumes
the timing field. Old receipts are not rewritten or supplied invented phases.

This is timing instrumentation, not a hard-deadline fix. The existing synchronous
CreateProcess call cannot be interrupted by the subsequent stopwatch wait loop;
that unchanged limitation can overrun the nominal2s before the loop checks time.
No watchdog, detached launcher, timer reset or broader process rights were added.

## Test-first verification and independent review

New launch-timing.tests.ps1 first failed at the missing real boundary-recorder
assertion. After implementation,8 assertions passed on PowerShell7 and Windows
PowerShell5.1. They exercise the real recorder with a running stopwatch and real
UTC/monotonic reads, verify retained call order, elapsed-tick continuity, frequency,
JSON roundtrip and absence of invented process/shutdown proof. No process starts.

Existing helper suites also passed on both hosts: child-readiness28,
full-parent12, tree-stage9, minimal-entry12, runner-regression21 and shutdown-state9.
Total99 assertions per host including the new8. This is scoped helper coverage,
not a full native runner, Go/product or fuzz suite. The historical runner failure
and16 unrun checks remain open; no old actual probe or failed preparation replayed.

Independent read-only review found no Critical/Important issue and supported the
one diagnostic sample and an unfinished checkpoint only. It ran no tests/native
calls or PID queries. Retained minor limits: recorder units do not exercise all
six placements/export integration; that sequence/export was checked from the one
actual saved result below. Failed launches throw without returning Result, so
their accumulated boundaries are not preserved by the unchanged launch-error
envelope. Boundary intervals include instrumentation/successful Check overhead,
not pure native API latency. The reviewer excluded runtime prediction, strict
hard deadline during blocking startup, failure-envelope redesign and acceptance;
none of these exclusions is silently treated as verified.

## One actual sample — budget consumed before root thread resumed

Used the unchanged child-readiness fixture in fresh child-readiness-002 ONCE,
not a replay of001. UTC10:21:20.0235178–10:21:22.1333839 (Asia/Tbilisi14:21).
The wrapper returned exit1: no child-ready.json, ChildReady=false,
FixturePassed=false, Acceptance=false and OriginalTestPassed=false.

All six boundaries survived native receipt/result serialization, in the expected
order, with monotonic nondecreasing timestamps/elapsed ticks and Frequency10000000.
The following interval deltas use absolute monotonic Timestamp/Frequency; budget
positions use the original watch's ElapsedTicks/Frequency:

| Observed launch interval | Start budget position | End budget position | Boundary-to-boundary duration |
| --- | ---: | ---: | ---: |
| CreateProcess | 7.2225ms | 2104.9304ms | 2097.7060ms |
| AssignProcessToJobObject | 2104.9338ms | 2105.3155ms | 0.3811ms |
| ResumeThread | 2105.3167ms | 2105.4985ms | 0.1816ms |

Creation-before/after UTC10:21:20.0307626Z /10:21:22.1284643Z;
assignment-before/after10:21:22.1284817Z /10:21:22.1288588Z;
resume-before/after10:21:22.1288647Z /10:21:22.1290449Z.
Fields are sampled successively, not an atomic synchronized UTC/tick snapshot.

The original2s budget was already exhausted when CreateProcess returned and the
root thread was resumed. The existing wait loop could then report timeout without
a remaining planned execution window for descendant readiness. This localizes the
dominant measured interval for THIS sample to the creation bracket, rather than
establishing a parent-script Start-Process deadlock or child-file-write defect.
It does not isolate the OS/security/runtime/scheduling reason for that bracket,
nor explain older receipts that lack these records. Instrumented samples are not
a controlled causal A/B experiment; no speculative exclusion or security change.

Native result: TimedOut=true, root not exited, TotalProcesses1. The one saved
member was owned/running powershell.exe, QueryError0/read exit259. Forced owned
termination requested; cleanup verified with accounting0/member list empty,
ShutdownElapsedMs4. NaturalShutdownVerified=false, unchanged schema3/v2.
Stdout/stderr both empty. Total UTC receipt interval2109.8661ms includes shutdown;
it is not an enforced exact2s wall-clock limit or a pure creation-time measurement.
No second timing sample ran; the actual readiness test remains failed.

Ignored worktree-relative evidence:
.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/child-readiness-002/.
SHA256 process/receipt.json:
BFBEC70085CC85EA1CC58246CEEC7DAA95C3A22D2B71EF4651B1D71A65418E96;
result.json:
AD73295A7CBBEC9D2900F1749C98E97DFD1F4D62EA26E1F17AE2CECFEE34C6B5.

## Recovery and next boundary

### Subsequent readonly checks, no new sample

Queried Defender/Operational, CodeIntegrity/Operational, AppLocker/EXE and DLL,
and PowerShell/Operational from one second before002 StartUTC to one second after
EndUTC. Each query returned no matching event, not an access error. This is not
proof that scanning, policy checks or scheduling played no part: these logs do
not provide a complete creation-latency trace. Receipt/result hashes above were
rechecked unchanged. No security configuration, recording or process sample ran.

Readonly application discovery found the built-in WPR and installed Windows
Performance Toolkit WPR/WPAExporter, plus tracerpt. WPR reported version10.0.26100
and no recording active at this check; its profile inventory included GeneralProfile,
CPU, DiskIO, FileIO and Minifilter. Only status/profile queries ran, not recording.
This is a prerequisite check, not assurance that no session will be active later.
A proposed single short local performance-trace experiment needs its own approval;
it may include system process names/paths and must not upload/commit raw traces
or interfere with another recording. It must not repeat old identities or relax
the2s sample budget. No cause-specific fix is established by the available logs.

Recover the latest codex/c12-b01-task9-coordinator checkpoint, not automatically
main; read current-status.md and repository-recovery.md. Necessary diagnostic
source/tests and this report are tracked; raw receipts, tool caches and live state
are not. No dependency on old ignored copies is needed for helper tests. Neither
consumed child-readiness001 nor002 may be replayed after cloning, even if their
raw directories are absent. Existing run.tests and child-readiness payload remain
unchanged; there is no claim that the original descendant regression is fixed.

Next investigation should examine the creation interval before changing scripts
or budgets. No cause-specific fix is justified yet. Further actual samples,
failure-envelope redesign, watchdogs or test-timing changes need separately agreed
scope. Original origin87, Envelope/Pair and C11/C12/Task10/B01/production/main gates
remain unresolved; this diagnostic does not authorize their continuation/acceptance.
