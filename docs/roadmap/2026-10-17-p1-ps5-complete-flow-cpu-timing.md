# PS5 complete-order group timing — timeout inside unclosed boundary group

Afterb78a2a5b, user approved ONE fresh ordinary-token PS5 diagnostic process,
configured30000ms. Reused existing complete-flow-cpu-timing probe unchanged:
shape6, reflection3, fixture8, original35, ownership10 in the complete order.
No product/test/probe edit, warmup, preflight run, retry, fallback, reviewer,
old-budget use or timeout increase. New diagnostic1/1 is spent.

## Observed result

PS5 5.1.26100.9711 Desktop x64/admin=False;TimedOut=true,exit-1,wall30085ms,
cumulative process CPU31859.375ms,stderr empty,owned-host-exited.
Configured30000ms;actual30085 includes85ms termination/collection overhead.
CaptureValid=true,Complete=false,FormalAcceptance=false,Disposition=BudgetExhausted.
Controller exit0 means a valid saved diagnostic, NOT that the child tests passed.

Raw169 lines:104 timing events,65 ordinary lines,48 completed PASS groups.
50 start/end pairs closed;four scopes remain open, no malformed partial tail:
Run/CompleteFlow,Stage/ExistingFull,Stage/ExistingAll,
Original35/ArrayAndIntegerBounds. No assertion error saved before kill.

| Open scope | Start relative to first Run marker ms | Absolute cumulative process CPU sample ms |
| --- | ---: | ---: |
| Run/CompleteFlow | 0.000 | 296.875 |
| Stage/ExistingFull | 696.742 | 1156.250 |
| Stage/ExistingAll | 830.537 | 1390.625 |
| Original35/ArrayAndIntegerBounds | 9268.150 | 11656.250 |

This diagnostic positively saved the boundary group START at9268.150ms, unlike
the previous uninstrumented timeout's last-PASS-only evidence. It did not save
that group's END, so there is NO completed boundary duration or CPU delta.
There are no internal phase markers. Exact location among invalid-input checks,
4096 construction,evaluation/result comparison and4097 refusal remains unknown.
Do not use deadline minus group start as a completed evaluation duration.

The48 completed groups are shape6 + reflection3 + fixture8 + original first31
through PassiveDescriptors. The boundary and remaining3 original groups, all10
ownership regression groups and final wrapper markers are uncompleted.

## Comparison with the earlier PS7 diagnostic

Same probe bytes/current candidate, separate single observations on different
runtimes and times; NOT a controlled benchmark or unique version-causality proof.

| Region | Earlier PS7 wall ms | This PS5 wall ms | Earlier PS7 CPU ms | This PS5 CPU ms |
| --- | ---: | ---: | ---: | ---: |
| ShapeLookup stage | 510.616 | 474.290 | 656.250 | 765.625 |
| ReflectionReuse stage | 130.177 | 122.308 | 187.500 | 234.375 |
| FixtureComplete group | 775.044 | 1456.260 | 687.500 | 1515.625 |
| Original35/ProgressMonotonic | 207.813 | 963.298 | 437.500 | 937.500 |
| Original35/PhaseBudgets | 206.860 | 813.081 | 265.625 | 796.875 |
| Original35/ArrayAndIntegerBounds | 4932.728 | No END | 6406.250 | No END |

The first31 original-group span is6805.106ms here versus1915.322ms in that PS7
diagnostic, measured first group START through PassiveDescriptors END.
Boundary START is9268.150ms here versus3864.703ms in the PS7 diagnostic.
Thus reduced remaining headroom is already visible BEFORE the4096 group.
Several earlier small cases differ, not just the large-input boundary.
These observations do not establish host load, runtime version,JIT/GC,helper
checks or any other unique cause; historical load/runtime internals not traced.

CPU sums all child threads,including runtime work,not just the test's statements.
CPU can exceed wall and is not utilization or a group-exclusive application
instruction profile. Samples are coarse;zero delta is not no work. Wall then CPU
samples are non-atomic;boundary serialization/flush/sampling add unmeasured cost.
Nested parent intervals overlap children;do not sum them. An instrumented copy
cannot replace original complete-entry acceptance even if it later completes.

## Integrity and local evidence

Frozen208 pins: inherited201 plus seven previous original-full packet anchors.
Existing six diagnostic copies/product/original tests unchanged. Engine bytes
checked before/after;ordinary token checked;slot recorded before child Start;
separate UTF-8 streams,bounded waits,owned process handle only for termination.

Read-only recheck matched208 pins,controller/engine,index and both streams.
All50 closed wall AND CPU pairs independently recomputed exactly from raw events;
169/104/65/48 counts and four-scope stack matched. No rewrite or rerun.
This exercises timeout/valid-prefix/owned-exit capture;no malformed tail or
CPU-read-failure was exercised or claimed tested. Prior verdicts remain intact.

Ignored local packet:
.superpowers/sdd/2026-10-17-p1-ps5-complete-flow-cpu-timing/
retains scope/controller/frozen inventory/index/raw,NOT committed.
Digests show retained local consistency,not independent attestation.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 4EAF8C66FA8282FE67E3404296D3AA7CE78114B0841A886F22A57795AAE9D78E |
| run-once.ps1 | C49A84E08A47C5F9D8632B0D4578949F7F96080001E4048FBB6807B6E5161B6A |
| invocations.json | 631A4C215CCB1A65892142ECCD38E127537AB30A883542F7890ACBECC75956A8 |
| D1.stdout.txt | 5F39778CA4147ED64AE0C102155A88A7E5451A2A280026521D4275936329D4CC |
| D1.stderr.txt (empty) | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Candidate unchanged:
953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.
Probe unchanged:
04526277000A8D362458B9B7BFC6F84D4BE14CA2C0BF6F63011F6EE175DB6A88.
PS5 executable:
60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F.

## Handoff

New PS5 diagnostic1/1 spent. Original-full2/2 spent (PS7 PASS,PS5 timeout).
Earlier PS7 CPU diagnostic1/1,targeted2/2 spent. Shape2/3/one frozen;
prior independent full1/2/one frozen;reflection3/4/one frozen;
extended-view3/7/four frozen;review0/1 unused/ineligible;other budgets unchanged.
Candidate UNACCEPTED,I1/P1 unresolved;no main/C12/project/live/runtime/release
acceptance or automatic reviewer start.

Save report/status/recovery only to the development branch;existing portable
probe already committed. Clone restores source/tooling/reports,not ignored raw,
cached engines,credentials or replay authority.

Next design narrowly separated boundary subphase markers in diagnostic copies,
retaining the complete order,original assertions and30s cap. Keep markers out of
hot common helpers;separate input construction from existing check/equality work
before proposing product changes. This report does not authorize another child.
No unchanged full rerun,limit increase,frozen-slot use or speculative optimization.

## All completed pairs

Milliseconds;CPU includes all process threads. Inclusive stages overlap children.

| Kind/name | Elapsed wall ms | CPU delta ms |
| --- | ---: | ---: |
| Shape/RootOrdinal | 140.701 | 156.250 |
| Shape/FieldBounds | 15.231 | 31.250 |
| Shape/ObservationOrdinal | 91.796 | 125.000 |
| Shape/KindIsolation | 83.661 | 109.375 |
| Shape/RepeatedCalls | 73.783 | 203.125 |
| Shape/DescriptorSafety | 21.769 | 93.750 |
| Stage/ShapeLookup | 474.290 | 765.625 |
| Reflection/InputMutation | 51.126 | 140.625 |
| Reflection/DecorationAfterClean | 36.906 | 46.875 |
| Reflection/DescriptorAfterClean | 13.850 | 15.625 |
| Stage/ReflectionReuse | 122.308 | 234.375 |
| Ownership/FixtureConstructor | 12.251 | 15.625 |
| Ownership/FixturePrimitiveMatrix | 65.298 | 93.750 |
| Ownership/FixtureDescriptors | 11.975 | 15.625 |
| Ownership/FixtureInputIdentity | 17.030 | 15.625 |
| Ownership/FixtureIsolation | 10.005 | 31.250 |
| Ownership/FixtureWrappedInputIdentity | 3.580 | 0.000 |
| Ownership/FixtureExtendedView | 14.431 | 15.625 |
| Ownership/FixtureComplete | 1456.260 | 1515.625 |
| Original35/ApiRequired | 3.548 | 0.000 |
| Original35/EmptyPrepared | 22.115 | 15.625 |
| Original35/NormalStages | 83.066 | 140.625 |
| Original35/AbsoluteDeadline | 44.608 | 93.750 |
| Original35/StopBoundary | 74.189 | 156.250 |
| Original35/TotalDeadline | 33.619 | 78.125 |
| Original35/StartupConsumesBudget | 24.045 | 31.250 |
| Original35/EqualTimeSequence | 156.922 | 328.125 |
| Original35/BackwardTime | 105.159 | 187.500 |
| Original35/SequenceRefusal | 117.183 | 250.000 |
| Original35/SessionClockRefusal | 251.533 | 484.375 |
| Original35/ActorDirection | 125.746 | 281.250 |
| Original35/ArmPrerequisites | 455.599 | 859.375 |
| Original35/ReleaseOnce | 369.893 | 375.000 |
| Original35/ProgressMonotonic | 963.298 | 937.500 |
| Original35/ExitBeforeEOF | 132.715 | 109.375 |
| Original35/UnknownExit | 109.514 | 281.250 |
| Original35/NonzeroExit | 136.590 | 156.250 |
| Original35/DataBounds | 420.934 | 468.750 |
| Original35/ControlBounds | 483.684 | 515.625 |
| Original35/ReadFault | 94.302 | 171.875 |
| Original35/SupervisorLost | 94.837 | 125.000 |
| Original35/LauncherLost | 88.063 | 93.750 |
| Original35/FailureSticky | 124.031 | 125.000 |
| Original35/PhaseBudgets | 813.081 | 796.875 |
| Original35/CleanupUnknown | 354.245 | 359.375 |
| Original35/PersistRefused | 379.220 | 375.000 |
| Original35/TerminalTrailing | 130.631 | 156.250 |
| Original35/InvalidScenario | 204.838 | 203.125 |
| Original35/InvalidObservation | 227.177 | 250.000 |
| Original35/PassiveDescriptors | 152.929 | 140.625 |
