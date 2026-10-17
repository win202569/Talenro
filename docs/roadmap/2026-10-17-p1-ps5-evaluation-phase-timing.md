# PS5 evaluation phase diagnostic — root setup completes; validation/copy remains open

Afterd7ea8aed, user approved ONE ordinary-token PS5 child,configured30000ms,
same complete test order. Only the4096 boundary call uses a locally loaded,
phase-marked product copy. Product original and original tests remain untouched.
No warmup,preflight run,retry,fallback,reviewer,old-budget use or timeout increase.
New phase diagnostic1/1 spent.

## Saved evidence and hypothesis result

PS5 5.1.26100.9711 Desktop x64/admin=False;TimedOut=true,exit-1,wall30063ms,
cumulative CPU33234.375ms,stderr empty,owned-host-exited.
Configured30000ms;actual30063 includes63ms termination/collection overhead.
CaptureValid=true,Complete=false,FormalAcceptance=false,Disposition=BudgetExhausted.
Controller exit0 means a valid retained diagnostic prefix,NOT successful tests.

Raw178 lines:113 timing events,65 ordinary lines,48 completed PASS groups.
53 closed pairs,seven open scopes,no malformed partial tail.
The48 groups are shape6/reflection3/fixture8/original first31 through
PassiveDescriptors;boundary and all subsequent groups lack completion.

| Relevant region | Start ms | End ms | Completed wall ms | Completed CPU ms |
| --- | ---: | ---: | ---: | ---: |
| InvalidInputs | 10247.041 | 10526.071 | 279.030 | 250.000 |
| Build4096 | 10526.570 | 11258.269 | 731.699 | 734.375 |
| Evaluate4096 | 11347.704 | Missing | Unknown | Unknown |
| ProductPhase/RootSetup | 11368.395 | 11400.802 | 32.407 | 46.875 |
| ProductPhase/ValidationCopy | 11401.517 | Missing | Unknown | Unknown |
| ProductPhase/Decision | Not started | Missing | Unknown | Unknown |
| Assert4096 / Check4097 | Not started | Missing | Unknown | Unknown |

Times relative to first Run marker,not process launch. RootSetup excludes the
helper's copy-file loading and Set-StrictMode preceding its first marker.
ValidationCopy includes original per-observation validation AND snapshot copying.
No per-row/helper/copy-subphase markers are present.

H1 from the read-only trace (validation/copy not finished and decision not yet
reached at cutoff) is SUPPORTED for this diagnostic run:root completed,validation
START recorded without END,and decision START absent. This is a phase-localizing
observation,not proof of the unique mechanism behind the original timeout.
We still do not know which record/helper/allocation is active,how much of the
phase is pure validation versus copying,or whether JIT/GC/runtime conditions
dominate. No completed validation duration/CPU delta is available. Never derive
one by subtracting the deadline or treating total process CPU as phase CPU.

Seven open scopes:Run/CompleteFlow,Stage/ExistingFull,Stage/ExistingAll,
Original35/ArrayAndIntegerBounds,Boundary/Check4096,Boundary/Evaluate4096,
ProductPhase/ValidationCopy. No assertion error saved before kill.

## Fidelity and isolation

Portable throwaway folder:
[phase probe](../diagnostics/isolated-fuzz/ps5-evaluation-phase-timing/probe.ps1)
and six adjacent scripts,including offline-supervision-policy.phase-copy.ps1.
Complete order stays shape6/reflection3/fixture8/original35/ownership10.

Five wrapper/test copies are byte-identical to previous subphase copies.
Only the original35 diagnostic copy adds ONE dot-source statement inside
P1-BoundaryCheck4096,loading phase-copy in that helper's local function scope,
before its Evaluate4096 START. Earlier tests/ordinary calls import the unchanged
original parent product. No global replacement,extra mode branches or per-row
events are introduced. Local return/restoration/later-test runtime path was NOT
exercised because the boundary call did not finish;scope isolation is source-
reviewed here,not claimed dynamically proven by this timeout.

The phase product copy differs only by a diagnostic header and six unconditional
start/end events at root,validation/copy and decision boundaries. Original helper
bodies,loops,guards,copying,decisions and catch remain otherwise unchanged.
All seven files/controller statically parsed before launch. Removing header/
phase events restores product source exactly (normalized line endings/end blanks).
Removing the local dot-source restores previous original35 diagnostic exactly;
other five hashes match. Existing assertion expressions/argument order remain.

Instrumentation and separate function definitions/import can affect dispatch/JIT
and timing;they are explicit diagnostic interventions. This is not an untouched
original-entry benchmark,proof of stable throughput,or permission to waive PS5
acceptance. Old PS7 profiling is not substituted as current PS5 measurement.
CPU includes all child threads,coarse sampling and runtime work;can exceed wall,
not utilization. Wall/CPU samples are non-atomic;logging adds unmeasured overhead.
Nested intervals overlap;do not sum rows or infer unique helper attribution.

## Integrity

Frozen231 pins:inherited219 + five prior packet anchors + seven new copies.
Engine checked before/after;ordinary token required;slot consumed before Start;
separate UTF-8 streams,bounded wait/drain,owned process handle only for termination.
Read-only recheck matched231 pins/controller/engine/index/streams. All53 completed
wall AND CPU pairs independently recomputed exactly from raw records;
178/113/65/48 counts and seven-scope stack matched. No historical rewrite/rerun.
Valid-timeout prefix exercised;malformed-tail/CPU-read-failure paths not exercised.

Ignored packet:
.superpowers/sdd/2026-10-17-p1-ps5-evaluation-phase-timing/
retains scope/controller/frozen inventory/index/raw,NOT committed.
Digests show retained local consistency,not independent attestation.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 63C21E3CDC564D89EDE2930BD5703A223C119D19AF27C8BE4976AB4F273821FE |
| run-once.ps1 | BC5ADD71DE0954380A5620D17B75CDD56BD7ECDC79B914050DF176827D9FAD32 |
| invocations.json | E6D934770C7CAD48070486CFDDC392E760000893FB4BB52A25B11CBD74ED7025 |
| D1.stdout.txt | A42DF88DC3986718AFC4C08E4D7EF1538FD0A47CF21ED46D7041A419F25D64A8 |
| D1.stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| Phase product copy | F8782165C34A39A91ED8A7009FE5D372D605DA6C678CC18C6CE01B316053F8DB |
| Original35 diagnostic copy | 6110E654D7BD695ABB06A474B5A3503FE2C30461D4D019A6EC0D70930C45A1C4 |

Original product remains:
953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.
PS5 executable:
60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F.

## Handoff

New phase1/1 spent;PS5 subphase1/1 and group1/1 spent;original-full2/2 spent
(PS7 PASS,PS5 timeout);earlier PS7 CPU1/1,targeted2/2 spent.
Shape2/3/one frozen;prior independent full1/2/one frozen;reflection3/4/one frozen;
extended-view3/7/four frozen;review0/1 unused/ineligible;other budgets unchanged.
Candidate UNACCEPTED,I1/P1 unresolved;no main/C12/project/live/runtime/release
acceptance or automatic independent reviewer start.

Save portable diagnostic copies/report/status/recovery on the development branch.
Clone restores inspectable artifacts,not ignored raw,cached engines,credentials,
replay permission or budget renewal. Previous failed verdicts remain intact.

Next inspect the validation/copy body and design evidence that distinguishes its
type/member checks,shape/scalar validation and copying costs. Do not equate the
phase localization with a proven reason to remove defenses or change semantics.
Any new child/product change needs separate scope;no automatic full rerun,
old-slot use,raised timeout or speculative optimization follows this result.

## Completed pairs

Milliseconds;inclusive parents overlap children;CPU sums all process threads.

| Kind/name | Elapsed wall ms | CPU delta ms |
| --- | ---: | ---: |
| Shape/RootOrdinal | 168.962 | 203.125 |
| Shape/FieldBounds | 14.913 | 46.875 |
| Shape/ObservationOrdinal | 97.600 | 156.250 |
| Shape/KindIsolation | 81.215 | 171.875 |
| Shape/RepeatedCalls | 81.731 | 281.250 |
| Shape/DescriptorSafety | 19.822 | 46.875 |
| Stage/ShapeLookup | 522.412 | 921.875 |
| Reflection/InputMutation | 51.727 | 140.625 |
| Reflection/DecorationAfterClean | 39.463 | 62.500 |
| Reflection/DescriptorAfterClean | 13.394 | 31.250 |
| Stage/ReflectionReuse | 123.451 | 265.625 |
| Ownership/FixtureConstructor | 11.460 | 15.625 |
| Ownership/FixturePrimitiveMatrix | 66.131 | 46.875 |
| Ownership/FixtureDescriptors | 10.622 | 15.625 |
| Ownership/FixtureInputIdentity | 13.755 | 0.000 |
| Ownership/FixtureIsolation | 9.165 | 0.000 |
| Ownership/FixtureWrappedInputIdentity | 3.735 | 0.000 |
| Ownership/FixtureExtendedView | 11.992 | 15.625 |
| Ownership/FixtureComplete | 1596.444 | 1484.375 |
| Original35/ApiRequired | 7.997 | 0.000 |
| Original35/EmptyPrepared | 78.473 | 109.375 |
| Original35/NormalStages | 336.718 | 468.750 |
| Original35/AbsoluteDeadline | 159.158 | 343.750 |
| Original35/StopBoundary | 347.614 | 796.875 |
| Original35/TotalDeadline | 200.451 | 437.500 |
| Original35/StartupConsumesBudget | 108.093 | 218.750 |
| Original35/EqualTimeSequence | 141.800 | 328.125 |
| Original35/BackwardTime | 110.907 | 187.500 |
| Original35/SequenceRefusal | 128.319 | 250.000 |
| Original35/SessionClockRefusal | 243.858 | 515.625 |
| Original35/ActorDirection | 114.429 | 218.750 |
| Original35/ArmPrerequisites | 410.078 | 718.750 |
| Original35/ReleaseOnce | 332.168 | 328.125 |
| Original35/ProgressMonotonic | 750.471 | 734.375 |
| Original35/ExitBeforeEOF | 136.916 | 140.625 |
| Original35/UnknownExit | 115.962 | 312.500 |
| Original35/NonzeroExit | 119.574 | 125.000 |
| Original35/DataBounds | 386.711 | 437.500 |
| Original35/ControlBounds | 507.414 | 500.000 |
| Original35/ReadFault | 124.394 | 203.125 |
| Original35/SupervisorLost | 108.277 | 125.000 |
| Original35/LauncherLost | 98.520 | 93.750 |
| Original35/FailureSticky | 127.733 | 125.000 |
| Original35/PhaseBudgets | 799.314 | 812.500 |
| Original35/CleanupUnknown | 393.840 | 375.000 |
| Original35/PersistRefused | 360.683 | 343.750 |
| Original35/TerminalTrailing | 144.860 | 171.875 |
| Original35/InvalidScenario | 208.713 | 218.750 |
| Original35/InvalidObservation | 187.365 | 218.750 |
| Original35/PassiveDescriptors | 129.996 | 125.000 |
| Boundary/InvalidInputs | 279.030 | 250.000 |
| Boundary/Build4096 | 731.699 | 734.375 |
| ProductPhase/RootSetup | 32.407 | 46.875 |
