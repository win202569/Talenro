# PS5 boundary subphase diagnostic — construction completes; evaluation has no END

After683edfa4, user approved ONE fresh ordinary-token PS5 child, cap30000ms,
complete shape6/reflection3/fixture8/original35/ownership10 order.
New diagnostic copies only. No product/original-test edits,warmup,preflight run,
retry,fallback,reviewer,old-budget use or raised limit. New1/1 is spent.

## Saved result and narrowed location

PS5 5.1.26100.9711 Desktop x64/admin=False;TimedOut=true,exit-1,total30069ms,
cumulative CPU31296.875ms,stderr empty,owned-host-exited.
Configured30000ms;actual30069 includes69ms termination/collection overhead.
CaptureValid=true,Complete=false,FormalAcceptance=false,Disposition=BudgetExhausted.
Controller exit0 means valid saved diagnostic prefix,not test completion.

Raw175 lines:110 timing events,65 ordinary lines,48 completed PASS groups.
52 completed pairs,six open scopes,no malformed partial tail.
First48 groups are shape6/reflection3/fixture8/original first31 through
PassiveDescriptors. Boundary group and all subsequent groups lack completion.

| Boundary scope | Start ms | End ms | Completed wall ms | Completed CPU ms |
| --- | ---: | ---: | ---: | ---: |
| InvalidInputs | 8401.317 | 8659.990 | 258.672 | 234.375 |
| Build4096 | 8660.747 | 9423.061 | 762.314 | 765.625 |
| Check4096 (arguments + nested helper) | 9424.193 | Missing | Unknown | Unknown |
| Evaluate4096 | 9427.186 | Missing | Unknown | Unknown |
| Assert4096 | Not reached | Missing | Unknown | Unknown |
| Check4097 | Not reached | Missing | Unknown | Unknown |

Offsets are relative to first Run marker,not process launch. Boundary group START
was8364.368ms. Evaluate4096 is the most specific active instrumented scope.
There is no evaluation-END record,no observed result-equality stage and no4097
refusal stage. Input construction completed. This narrows the uncompleted region
to the evaluation phase rather than blaming construction or comparison.
It does NOT identify an instruction/helper inside the product,prove a permanent
hang,or supply a completed evaluation duration. Do not subtract the deadline to
invent an evaluation duration/CPU delta. A missing END is not a product exception.

Open stack:Run/CompleteFlow,Stage/ExistingFull,Stage/ExistingAll,
Original35/ArrayAndIntegerBounds,Boundary/Check4096,Boundary/Evaluate4096.
No assertion failure was saved. Earlier PS5 group-only diagnostic reached its
boundary at9268.150ms;this run reached it at8364.368ms. Different instrumentation
and runtime conditions,not a matched speedup measurement or unique cause proof.

## Fidelity and measurement limits

Portable throwaway copies:
[probe](../diagnostics/isolated-fuzz/ps5-boundary-subphase-timing/probe.ps1)
and five adjacent scripts. Five copies are byte-identical to previous group
diagnostic copies. Only original35 diagnostic copy gains boundary events and
one dedicated P1-BoundaryCheck4096 helper. Common P1-Check is unchanged:
NO new mode checks or per-call branches affect earlier tests.

Only the4096 boundary call selects the dedicated helper. Its parameters,original
API call,r local variable,equality assertion,return value and function-local
scope match P1-Check,with unconditional evaluation/assertion events inserted.
Original caller argument expressions and order remain unchanged.
Outer Check4096 encloses argument evaluation and inner evaluation/assertion.
Invalid-input/build/4097 outer events wrap unchanged original statements.
No counters/markers injected into product or per-observation loop.

All six files and controller statically parsed before launch. Removing new
Boundary events/special helper and restoring the one helper call name reverses
the new original35 copy exactly to previous group diagnostic (normalized line
endings/end blanks);stripping helper markers/name yields original P1-Check text
exactly. Other five copy hashes match previous files. Previous copies had
already been reversed to original test/wrapper sources. Assertions preserved.

Instrumentation still costs time and can alter runtime/JIT conditions;dedicated
helper is intentionally a diagnostic intervention,not an uninstrumented product
benchmark. CPU is all child threads including runtime;coarse,non-atomic wall/CPU
samples may exceed wall or show zero delta for real work. Parent regions overlap.
Old PS7 validation-loop profiling on a different product copy cannot assign
this PS5 unfinished evaluation uniquely to validation,decision,JIT,GC or load.

## Integrity and local receipts

Frozen219 pins = inherited208 + five latest packet anchors + six new copies.
Engine checked before/after;slot recorded before launch,ordinary token required,
separate UTF-8 streams,bounded waits and owned process handle only for termination.
Read-only recheck matched219 pins,controller/engine/index and both streams.
All52 closed wall/CPU durations independently recomputed from raw pairs exactly;
175/110/65/48 counts and six-scope stack matched. Raw/controller/old verdicts
not rewritten. Timeout/valid-prefix path observed;malformed-tail/CPU-read-failure
paths not exercised or claimed validated.

Ignored local packet:
.superpowers/sdd/2026-10-17-p1-ps5-boundary-subphase-timing/
retains scope/controller/frozen inventory/index/raw,NOT committed.
Digests show retained local consistency,not independent attestation.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 1DCB87ABE8FED552D5FA923B013E2E1020E095667D93FDF383EAB3D29A303581 |
| run-once.ps1 | 0B3F388370DE58E6F35C4CA6E60023DF928D455E8D9B15E6AD2B1697FB677D86 |
| invocations.json | AA151B62F1C1EAA28B8831368F5F84B238FD8691F57E5DCD66ADD7312BCE98AA |
| D1.stdout.txt | 0609630662E28B3AE51389EE0B56ACC6F5C2769A120D5B54324C60B6DA91CD7F |
| D1.stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| New original35 copy | 592513B237B5CA898BF994627863B1EBAF1AF54893E03A0E7BA340EE2FDC8F58 |

Unchanged probe04526277,unchanged candidate full SHA256:
953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.
PS5 executable:
60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F.

## Handoff

New subphase1/1 spent;PS5 group diagnostic1/1 spent;original-full2/2 spent
(PS7 PASS,PS5 timeout);earlier PS7 CPU1/1,targeted2/2 spent.
Shape2/3/one frozen;prior independent full1/2/one frozen;reflection3/4/one frozen;
extended-view3/7/four frozen;review0/1 unused/ineligible;other budgets unchanged.
Candidate UNACCEPTED,I1/P1 unresolved. No main/C12/project/live/runtime/release
acceptance or automatic reviewer start.

Save portable copies/report/status/recovery to existing development branch.
Clone restores inspectable source/tooling/reports,not ignored raw,cached engines,
credentials,replay permission or renewed budget. Historical failures retained.

Next read-only trace current evaluation internals and identify a specific
testable hypothesis before any product edit or further process. Construction and
result comparison are not demonstrated fixes. No automatic full rerun,old-slot
use,timeout increase or speculative optimization follows this diagnostic.

## Completed pairs

Milliseconds;inclusive parents overlap children. CPU is cumulative all-thread.

| Kind/name | Elapsed wall ms | CPU delta ms |
| --- | ---: | ---: |
| Shape/RootOrdinal | 139.862 | 203.125 |
| Shape/FieldBounds | 13.375 | 0.000 |
| Shape/ObservationOrdinal | 92.263 | 125.000 |
| Shape/KindIsolation | 81.134 | 125.000 |
| Shape/RepeatedCalls | 86.444 | 296.875 |
| Shape/DescriptorSafety | 19.957 | 46.875 |
| Stage/ShapeLookup | 477.355 | 843.750 |
| Reflection/InputMutation | 44.152 | 187.500 |
| Reflection/DecorationAfterClean | 36.072 | 46.875 |
| Reflection/DescriptorAfterClean | 14.714 | 15.625 |
| Stage/ReflectionReuse | 117.486 | 265.625 |
| Ownership/FixtureConstructor | 12.322 | 15.625 |
| Ownership/FixturePrimitiveMatrix | 111.079 | 78.125 |
| Ownership/FixtureDescriptors | 12.224 | 15.625 |
| Ownership/FixtureInputIdentity | 13.279 | 15.625 |
| Ownership/FixtureIsolation | 9.530 | 15.625 |
| Ownership/FixtureWrappedInputIdentity | 3.787 | 15.625 |
| Ownership/FixtureExtendedView | 11.976 | 31.250 |
| Ownership/FixtureComplete | 1389.871 | 1312.500 |
| Original35/ApiRequired | 2.985 | 0.000 |
| Original35/EmptyPrepared | 16.335 | 15.625 |
| Original35/NormalStages | 80.724 | 140.625 |
| Original35/AbsoluteDeadline | 29.842 | 46.875 |
| Original35/StopBoundary | 69.041 | 140.625 |
| Original35/TotalDeadline | 32.315 | 46.875 |
| Original35/StartupConsumesBudget | 16.492 | 31.250 |
| Original35/EqualTimeSequence | 22.133 | 78.125 |
| Original35/BackwardTime | 17.287 | 31.250 |
| Original35/SequenceRefusal | 16.518 | 31.250 |
| Original35/SessionClockRefusal | 34.609 | 62.500 |
| Original35/ActorDirection | 115.130 | 218.750 |
| Original35/ArmPrerequisites | 460.889 | 875.000 |
| Original35/ReleaseOnce | 348.976 | 500.000 |
| Original35/ProgressMonotonic | 871.263 | 828.125 |
| Original35/ExitBeforeEOF | 127.191 | 140.625 |
| Original35/UnknownExit | 105.183 | 281.250 |
| Original35/NonzeroExit | 118.942 | 156.250 |
| Original35/DataBounds | 380.032 | 375.000 |
| Original35/ControlBounds | 447.898 | 468.750 |
| Original35/ReadFault | 109.658 | 218.750 |
| Original35/SupervisorLost | 106.817 | 93.750 |
| Original35/LauncherLost | 91.162 | 93.750 |
| Original35/FailureSticky | 120.688 | 125.000 |
| Original35/PhaseBudgets | 782.983 | 781.250 |
| Original35/CleanupUnknown | 347.807 | 343.750 |
| Original35/PersistRefused | 343.000 | 359.375 |
| Original35/TerminalTrailing | 149.451 | 171.875 |
| Original35/InvalidScenario | 200.910 | 203.125 |
| Original35/InvalidObservation | 190.659 | 203.125 |
| Original35/PassiveDescriptors | 128.753 | 125.000 |
| Boundary/InvalidInputs | 258.672 | 234.375 |
| Boundary/Build4096 | 762.314 | 765.625 |
