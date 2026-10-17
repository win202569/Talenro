# PS5 validation/copy progress — completed prefix dominated by validation

Checkpoint authored2026-10-18 local time. Diagnostic started
2026-10-17T19:59:29.1561495Z. After57df3287,user approved ONE ordinary-token PS5
child,configured30000ms,same complete order. Only the dedicated4096 call loads
an instrumented product copy. Original product/tests unchanged.
No warmup,preflight test,retry,fallback,reviewer,old-budget use or raised timeout.
New progress diagnostic1/1 spent.

## Result and completed-row snapshots

PS5 5.1.26100.9711 Desktop x64/admin=False;TimedOut=true,exit-1,wall30062ms,
cumulative process CPU31828.125ms,stderr empty,owned-host-exited.
Configured30000ms;actual30062 includes62ms termination/collection overhead.
CaptureValid=true,Complete=false,FormalAcceptance=false,Disposition=BudgetExhausted.
Controller exit0 means valid diagnostic evidence,NOT successful full regression.

| Completed rows | Offset from Run ms | Validation cumulative ms | Copy cumulative ms | Validation phase elapsed ms | Process cumulative CPU ms |
| --- | ---: | ---: | ---: | ---: | ---: |
| 1 | 9353.267 | 24.155 | 3.144 | 49.751 | 11015.625 |
| 512 | 15204.132 | 5767.365 | 96.140 | 5900.616 | 17687.500 |
| 1024 | 20757.039 | 11229.985 | 178.260 | 11453.523 | 22906.250 |
| 1536 | 26106.568 | 16490.836 | 258.662 | 16803.052 | 28203.125 |

At the last COMPLETE snapshot1536 records had both validation and snapshot-copy
finished. Their measured validation16490.836ms versus copying258.662ms gives
98.4557% validation share of the TWO accounted wall regions. This percentage is
NOT of total process wall,not CPU share,not current-cutoff total,not all4096
records and not a calibrated uninstrumented-product cost fraction.

The four increasing snapshots demonstrate forward progress through1536 rows,
not an immediate permanent stall. Exact completed count at kill is unknown;
there is no2048 snapshot. Do not claim exactly1536 at cutoff,an upper bound from
missing output,or extrapolate a4096 completion time. Work after the last snapshot
and any unfinished row is not in its cumulative values.

Last phase elapsed16803.052ms minus validation+copy16749.498ms leaves53.554ms
outside the two timed row regions. This is NOT total instrumentation overhead:
timer calls/dispatch/runtime perturbation may also be charged inside measured
regions. The original script's overhead has not been calibrated by a paired run.

Raw182 lines:113 CTIME events,four VCSAMPLE snapshots,65 ordinary lines,
48 completed PASS groups,53 closed pairs,seven open scopes,no malformed tail.
Original first31 groups through PassiveDescriptors completed,after shape6,
reflection3 and fixture8. Boundary/remaining original/ownership10/final wrappers
uncompleted. No assertion error saved before kill.

RootSetup completed55.736ms/CPU62.500ms. ValidationCopy START9303.516ms,no END;
decision/assertion4097 stages not reached. These are diagnostic phase observations,
not original-entry PS5 acceptance. Earlier original PS7 PASS and PS5 timeout remain.

## What this evidence supports

For the completed instrumented prefix,validation (type/member/shape/scalar checks
and other pre-copy row work) is much more expensive than constructing the owned
record snapshot. Optimizing snapshot copying alone is therefore not supported as
the next primary investigation by THESE measurements.

This does not identify a unique helper,which checks are redundant,or whether
runtime/JIT/GC/dispatch conditions dominate the validation span. Do not remove
descriptor,exact type/field,decorated-scalar,ownership or refusal defenses from a
timing fraction. There is no product fix or speedup demonstrated here.

## Instrumentation fidelity and limits

Portable throwaway folder:
[progress probe](../diagnostics/isolated-fuzz/ps5-validation-copy-progress/probe.ps1)
and six adjacent scripts. Six files are byte-identical to previous phase copies.
Only the phase product copy gains clearly delimited VC PROBE blocks.

Within its validation/copy loop,three monotonic timestamps per row delimit:
row start,validation end before original copy statement,and copy end after the
original list append. Accumulators update only after both regions complete.
Snapshot output occurs after row1 and every512 completed rows (max9 for4096),
not per row or per member. CPU sampled only at these sparse snapshot boundaries,
as cumulative ALL-thread process CPU,not separately attributed to validation/copy.
A per-row checkpoint condition/counter updates are diagnostic additions;there is
no claim of zero overhead. Logging,integer arithmetic,extra variables,new function
body size and runtime/JIT effects can perturb performance. No warmup or baseline
calibration was authorized or run.

All original validation/copy statements,order,inputs/assertions remain.
Removing all VC PROBE blocks restores the previous phase copy exactly after
line-ending/end-blank normalization. Seven sources/controller statically parsed
before launch. Previous copy already reverses to original product by removal of
phase markers/header;original product bytes are unchanged.
Inherited copy header saying phase markers only is not an exhaustive description
after these additions;VC PROBE blocks/report describe the measured instrumentation.
Frozen measured bytes were not edited to revise that wording.

Only4096 helper locally loads the copied module;earlier/other calls use original.
Return/restoration/later-test path again not exercised by this timeout;scope is
source-reviewed,not dynamically proven on the non-returning path.
CPU can exceed wall,coarse samples are non-atomic with wall;inclusive parents
overlap children. Snapshots do not prove stable latency or original performance.

## Integrity and retained receipts

Frozen243 pins = inherited231 + five latest packet anchors + seven new copies.
Engine checked before/after,ordinary token required,slot recorded before launch,
separate UTF-8 streams,bounded waits and owned handle only for termination.
After the run243 pins/controller/engine/index/streams matched.
All53 completed wall/CPU pairs and all four row-count/tick/time snapshots were
independently recomputed from raw records and matched.182/113/4/65/48 counts
and seven-scope stack checked. No raw/index/controller repair or historical rewrite.
Valid timeout/partial progress exercised;malformed-tail/CPU-read-failure paths not.

Ignored local packet (created before local date rollover):
.superpowers/sdd/2026-10-17-p1-ps5-validation-copy-progress/
retains scope/controller/frozen inventory/index/raw,NOT committed.
Digests show retained local consistency,not independent attestation.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 207B3A613337AA4681E9DC0696159326118481B1ABCDDA9529622A7C3CEAED19 |
| run-once.ps1 | 1274A523B70A348DC2452009921D1A2DD584D9B348D95B6F5088A8BB0A2C9848 |
| invocations.json | 55D6711349111EC5783C41AEBCA8B860A8184E768AD64B823B2DD997B50C1641 |
| D1.stdout.txt | 811D61A78E1B030288B096C8FA795D29EB99A65DA766C5DE7C8C7DDD8CE8FCC7 |
| D1.stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| Progress product copy | 37935F0E66B7ADC44E4E591FECE4B040A50E4E2586A17DE2760E3A594663A07E |

Original product:
953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.
PS5 executable:
60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F.

## Handoff

New progress1/1 spent;PS5 phase1/1,subphase1/1,group1/1 spent;
original-full2/2 spent (PS7 PASS,PS5 timeout);earlier PS7 CPU1/1,targeted2/2 spent.
Shape2/3/one frozen;prior independent full1/2/one frozen;reflection3/4/one frozen;
extended-view3/7/four frozen;review0/1 unused/ineligible;all other budgets unchanged.
Candidate UNACCEPTED,I1/P1 unresolved;no main/C12/project/live/runtime/release
acceptance or automatic reviewer start.

Save portable tooling/report/status/recovery on existing development branch.
Clone restores these artifacts,not ignored raw,cached engines,credentials,
replay authority or renewed allowance. Preserve earlier failures and scoped PASSes.
Next read-only inspect validation's type/member/shape/scalar call costs before
proposing any new measurement or semantic-preserving change. No further process,
full retry,limit increase,old-slot use or defense removal authorized here.

## Completed phase/group pairs

Milliseconds;inclusive parents overlap children;CPU sums all process threads.

| Kind/name | Elapsed wall ms | CPU delta ms |
| --- | ---: | ---: |
| Shape/RootOrdinal | 134.043 | 171.875 |
| Shape/FieldBounds | 14.987 | 31.250 |
| Shape/ObservationOrdinal | 96.148 | 109.375 |
| Shape/KindIsolation | 78.383 | 156.250 |
| Shape/RepeatedCalls | 79.615 | 328.125 |
| Shape/DescriptorSafety | 25.954 | 46.875 |
| Stage/ShapeLookup | 472.638 | 890.625 |
| Reflection/InputMutation | 44.722 | 125.000 |
| Reflection/DecorationAfterClean | 32.847 | 31.250 |
| Reflection/DescriptorAfterClean | 14.414 | 15.625 |
| Stage/ReflectionReuse | 111.627 | 203.125 |
| Ownership/FixtureConstructor | 10.774 | 15.625 |
| Ownership/FixturePrimitiveMatrix | 65.234 | 62.500 |
| Ownership/FixtureDescriptors | 10.888 | 15.625 |
| Ownership/FixtureInputIdentity | 14.966 | 15.625 |
| Ownership/FixtureIsolation | 11.589 | 0.000 |
| Ownership/FixtureWrappedInputIdentity | 3.485 | 15.625 |
| Ownership/FixtureExtendedView | 12.378 | 15.625 |
| Ownership/FixtureComplete | 1393.151 | 1484.375 |
| Original35/ApiRequired | 2.550 | 0.000 |
| Original35/EmptyPrepared | 23.107 | 15.625 |
| Original35/NormalStages | 77.786 | 93.750 |
| Original35/AbsoluteDeadline | 29.233 | 46.875 |
| Original35/StopBoundary | 67.180 | 140.625 |
| Original35/TotalDeadline | 40.189 | 62.500 |
| Original35/StartupConsumesBudget | 23.419 | 31.250 |
| Original35/EqualTimeSequence | 22.163 | 78.125 |
| Original35/BackwardTime | 17.122 | 31.250 |
| Original35/SequenceRefusal | 16.143 | 31.250 |
| Original35/SessionClockRefusal | 34.948 | 62.500 |
| Original35/ActorDirection | 18.346 | 31.250 |
| Original35/ArmPrerequisites | 70.316 | 156.250 |
| Original35/ReleaseOnce | 342.307 | 671.875 |
| Original35/ProgressMonotonic | 798.221 | 875.000 |
| Original35/ExitBeforeEOF | 130.666 | 125.000 |
| Original35/UnknownExit | 119.040 | 312.500 |
| Original35/NonzeroExit | 121.218 | 140.625 |
| Original35/DataBounds | 399.800 | 437.500 |
| Original35/ControlBounds | 454.358 | 468.750 |
| Original35/ReadFault | 100.017 | 171.875 |
| Original35/SupervisorLost | 93.322 | 109.375 |
| Original35/LauncherLost | 90.466 | 93.750 |
| Original35/FailureSticky | 108.807 | 109.375 |
| Original35/PhaseBudgets | 848.078 | 828.125 |
| Original35/CleanupUnknown | 358.506 | 343.750 |
| Original35/PersistRefused | 430.871 | 375.000 |
| Original35/TerminalTrailing | 144.022 | 171.875 |
| Original35/InvalidScenario | 215.880 | 203.125 |
| Original35/InvalidObservation | 204.093 | 218.750 |
| Original35/PassiveDescriptors | 109.756 | 109.375 |
| Boundary/InvalidInputs | 280.524 | 281.250 |
| Boundary/Build4096 | 918.338 | 875.000 |
| ProductPhase/RootSetup | 55.736 | 62.500 |
