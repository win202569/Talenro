# PS5 disjoint validation-region diagnostic — completes once; timing variability remains

After8800cbd7,user approved ONE ordinary-token PS5 child,configured30000ms,
same complete test order. Only the4096 boundary call loads an instrumented copy.
Original product/tests unchanged. No warmup,preflight test,retry,fallback,reviewer,
old-budget use or timeout increase. New region diagnostic1/1 spent.
Started2026-10-17T20:04:29.3923577Z (checkpoint local date2026-10-18).

## Complete diagnostic result

PS5 5.1.26100.9711 Desktop x64/admin=False;exit0,wall18993ms,
cumulative process CPU21015.625ms,TimedOut=false,stderr empty,owned-host-exited.
CaptureValid=true,Complete=true,FormalAcceptance=false,Disposition=Complete.

Raw245 lines:152 CTIME events,nine VCSAMPLE snapshots,84 ordinary lines,
64 PASS markers (62 test groups plus2 wrapper markers). All76 pairs closed,
no active scopes or partial tail. Original35 boundary and all later ownership
groups/assertions completed in this diagnostic path.
This is NOT unchanged-original-entry PS5 acceptance and does not erase its timeout.

| Completed rows | Record/kind ms | Shape ms | Field-value ms | Validation total ms | Copy ms |
| --- | ---: | ---: | ---: | ---: | ---: |
| 1 | 1.848 | 1.868 | 2.720 | 6.436 | 0.223 |
| 512 | 183.404 | 497.437 | 324.078 | 1004.919 | 22.317 |
| 1024 | 358.565 | 970.751 | 627.350 | 1956.667 | 39.154 |
| 1536 | 546.893 | 1535.698 | 975.585 | 3058.176 | 64.736 |
| 2048 | 818.716 | 2329.298 | 1462.414 | 4610.429 | 90.560 |
| 2560 | 975.020 | 2785.676 | 1748.397 | 5509.094 | 105.309 |
| 3072 | 1135.252 | 3284.631 | 2035.550 | 6455.433 | 120.512 |
| 3584 | 1283.488 | 3765.588 | 2315.386 | 7364.461 | 137.855 |
| 4096 | 1457.368 | 4275.431 | 2648.219 | 8381.018 | 156.168 |

At4096 completed rows,exact integer tick sums match:
Record1457.368ms + Shape4275.431ms + Field2648.219ms = Validation8381.018ms
(after displayed rounding). Copy156.168ms is separate.
Validation shares from exact ticks:Record17.3889%,Shape51.0133%,Field31.5978%.
These are disjoint instrumented elapsed-wall shares,NOT CPU shares,not a
calibrated untouched-product profile and not unique-cause attribution.

| Completed boundary phase | Elapsed wall ms | All-thread process CPU delta ms |
| --- | ---: | ---: |
| Boundary/InvalidInputs | 263.783 | 265.625 |
| Boundary/Build4096 | 436.115 | 437.500 |
| ProductPhase/RootSetup | 10.384 | 15.625 |
| ProductPhase/ValidationCopy | 8572.100 | 8781.250 |
| ProductPhase/Decision | 181.201 | 468.750 |
| Boundary/Evaluate4096 | 8783.485 | 9281.250 |
| Boundary/Assert4096 | 7.773 | 15.625 |
| Boundary/Check4096 | 8836.794 | 9296.875 |
| Boundary/Check4097 | 5.482 | 0.000 |

Validation/copy whole phase8572.100ms;decision181.201ms;API evaluation8783.485ms.
Assertion4096 and4097 refusal finished. These completed intervals are actual
recorded start/end differences,not deadline extrapolations. Nested rows overlap;
do not sum inclusive phases.

## Meaning of the three regions

Record/kind starts at original per-row result/index setup and includes plain
record type/member checks,Kind token validation and fixed schema-extra selection.
Shape is the complete Assert-SupervisionPolicyShape call,including nested passive
scalar/type/extended-member checks. Its51% MUST NOT be called hash-set membership
cost alone;the region does much more than field-name lookup.
Field-value spans original integer,declared-label,actor and kind-specific value
checks after shape through the pre-copy boundary. Input is the literal time kind,
so no extra per-kind payload branch executes there.

Regions partition the existing validation window algebraically and retain every
original statement and check. They narrow where THIS run spent time;they do not
say which helper is redundant or justify deleting guards/caching input acceptance.

## Important timing variability, not a product repair

Previous progress diagnostic timed out and its1536-row snapshot measured
validation16490.836ms/copy258.662ms. This region diagnostic at the SAME completed
count measured validation3058.176ms/copy64.736ms,then completed4096 and the suite.
Same product semantics but added instrumentation/function-body changes and
uncontrolled runtime/host conditions. Both validation AND copy differed strongly.
There is no matched/calibrated experiment proving why,that extra timestamps make
code faster,or that any product optimization occurred. NONE occurred.

The completed path is useful evidence,not stable throughput or regression-gate
clearance. Preserve prior original-entry timeout and all earlier diagnostic
failures. Do not promote this copy's PASS into original PS5 verification or use
one region percentage as a safe basis for another speculative product change.

## Fidelity and instrumentation limits

Portable throwaway folder:
[region probe](../diagnostics/isolated-fuzz/ps5-validation-regions/probe.ps1)
and six adjacent scripts. Six files are byte-identical to prior progress copies.
Only the copied product adds REGION PROBE blocks and expands snapshot schema/
parameters/call. Two new per-row timestamps bracket the original shape call:
five total timestamps per row versus three previously. Region accumulators update
after copying. Snapshots remain first row/every512 completed rows,max9.

No new common-helper branches,product guard removals,changed order,input values
or assertion expressions. Remove REGION PROBE blocks,restore snapshot schema/
signature/call,and source reverses exactly to prior progress copy (normalized line
endings/end blanks). Previous layers reverse to original product. All seven files
and controller statically parsed before launch. Original product hash unchanged.

Local copied-module loading is still restricted to dedicated4096 helper.
Later groups now completed with expected assertions after it returned. Function-
definition provenance/restoration was not separately logged,so no direct dynamic
identity attestation is claimed. Original scripts were not edited.

Timer calls,counters,checkpoint conditions,logging,new body size and dispatch/JIT
effects can change performance;overhead not calibrated. Inherited copy header
describing phase markers only is incomplete after progress/region additions;
explicit probe blocks/report define the measured changes,frozen bytes untouched.
CPU snapshots include ALL child threads and runtime work,may exceed wall,are
coarse and non-atomic with wall. No per-region CPU attribution exists.
Percentages cover timed validation regions,not total process/runtime-exclusive cost.

## Integrity and retained receipts

Frozen255 pins = inherited243 + five prior packet anchors + seven new copies.
Engine checked before/after;ordinary token enforced;sole slot recorded before
Start;separate UTF-8 streams,bounded waits and owned process handle termination.
Read-only recheck matched255 pins/controller/engine/index/streams. All76 pair
wall/CPU values and nine snapshots independently recomputed. Each snapshot's
three exact tick totals equals ValidationTicks;245/152/9/84/64 counts and closed
stack verified. No raw/index/controller repair or historical verdict rewrite.
This exercises complete/snapshot success paths,not new malformed/error paths.

Ignored local packet:
.superpowers/sdd/2026-10-18-p1-ps5-validation-regions/
retains scope/controller/frozen inventory/index/raw,NOT committed.
Digests show retained local consistency,not independent attestation.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 20BE43D9B2F193F9582ACCCFD9E552BD07D2F934103AC61814B7EDA2C37FDBFA |
| run-once.ps1 | DC4548B028A75D17109795590C3015E11B03D7BB6993F2BB4540355FE8648D23 |
| invocations.json | 393C74BD4E96F8607E241A8F50FF9B9362516F1A76D5FC5C54F772CB09125ACF |
| D1.stdout.txt | 9ED077F782849DD9B6388A44F27F254ECBE26731A5E4777F3E719CB54B4967FC |
| D1.stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| Region product copy | D3D063440039A31FEF45B3C696CF3EB6D587E759775D37215FA93AF15E18AD6D |

Original product:
953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.
PS5 executable:
60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F.

## Handoff

New regions1/1 spent;progress1/1,PS5 phase1/1,subphase1/1,group1/1 spent;
original-full2/2 spent (PS7 PASS,PS5 timeout);earlier PS7 CPU1/1,targeted2/2 spent.
Shape2/3/one frozen;prior independent full1/2/one frozen;reflection3/4/one frozen;
extended-view3/7/four frozen;review0/1 unused/ineligible;other budgets unchanged.
Candidate UNACCEPTED,I1/P1 unresolved;no main/C12/project/live/runtime/release
acceptance or automatic reviewer start.

Save portable scripts/report/status/recovery to existing development branch.
Clone restores artifacts,not ignored raw,cached engines,credentials,replay
authority or renewed budget. Next read-only compare instrumentation differences
and variable measurement conditions before proposing a controlled discrimination
or fix. Shape includes nested safety work;do not assume membership is the cause.
No automatic original-entry retry,old-slot use,timeout increase or guard removal.

## Completed pairs

Milliseconds;inclusive parents overlap children;CPU sums all child threads.

| Kind/name | Elapsed wall ms | CPU delta ms |
| --- | ---: | ---: |
| Shape/RootOrdinal | 140.592 | 171.875 |
| Shape/FieldBounds | 13.799 | 31.250 |
| Shape/ObservationOrdinal | 92.589 | 109.375 |
| Shape/KindIsolation | 73.310 | 125.000 |
| Shape/RepeatedCalls | 78.724 | 265.625 |
| Shape/DescriptorSafety | 18.235 | 46.875 |
| Stage/ShapeLookup | 464.706 | 796.875 |
| Reflection/InputMutation | 43.127 | 140.625 |
| Reflection/DecorationAfterClean | 34.032 | 31.250 |
| Reflection/DescriptorAfterClean | 14.831 | 15.625 |
| Stage/ReflectionReuse | 112.614 | 218.750 |
| Ownership/FixtureConstructor | 11.287 | 15.625 |
| Ownership/FixturePrimitiveMatrix | 64.595 | 46.875 |
| Ownership/FixtureDescriptors | 12.465 | 15.625 |
| Ownership/FixtureInputIdentity | 35.377 | 31.250 |
| Ownership/FixtureIsolation | 12.187 | 31.250 |
| Ownership/FixtureWrappedInputIdentity | 3.280 | 0.000 |
| Ownership/FixtureExtendedView | 12.155 | 15.625 |
| Ownership/FixtureComplete | 1421.614 | 1296.875 |
| Original35/ApiRequired | 2.650 | 15.625 |
| Original35/EmptyPrepared | 17.310 | 31.250 |
| Original35/NormalStages | 82.633 | 93.750 |
| Original35/AbsoluteDeadline | 28.713 | 62.500 |
| Original35/StopBoundary | 70.060 | 156.250 |
| Original35/TotalDeadline | 33.505 | 78.125 |
| Original35/StartupConsumesBudget | 17.781 | 46.875 |
| Original35/EqualTimeSequence | 22.623 | 31.250 |
| Original35/BackwardTime | 18.010 | 31.250 |
| Original35/SequenceRefusal | 16.264 | 15.625 |
| Original35/SessionClockRefusal | 34.311 | 93.750 |
| Original35/ActorDirection | 17.600 | 31.250 |
| Original35/ArmPrerequisites | 152.084 | 281.250 |
| Original35/ReleaseOnce | 391.889 | 796.875 |
| Original35/ProgressMonotonic | 758.851 | 750.000 |
| Original35/ExitBeforeEOF | 120.353 | 109.375 |
| Original35/UnknownExit | 110.740 | 250.000 |
| Original35/NonzeroExit | 136.058 | 156.250 |
| Original35/DataBounds | 373.706 | 453.125 |
| Original35/ControlBounds | 476.816 | 500.000 |
| Original35/ReadFault | 112.184 | 187.500 |
| Original35/SupervisorLost | 93.953 | 93.750 |
| Original35/LauncherLost | 88.343 | 93.750 |
| Original35/FailureSticky | 112.765 | 109.375 |
| Original35/PhaseBudgets | 806.342 | 843.750 |
| Original35/CleanupUnknown | 328.639 | 328.125 |
| Original35/PersistRefused | 311.054 | 312.500 |
| Original35/TerminalTrailing | 129.636 | 140.625 |
| Original35/InvalidScenario | 197.003 | 203.125 |
| Original35/InvalidObservation | 199.173 | 218.750 |
| Original35/PassiveDescriptors | 116.723 | 109.375 |
| Boundary/InvalidInputs | 263.783 | 265.625 |
| Boundary/Build4096 | 436.115 | 437.500 |
| ProductPhase/RootSetup | 10.384 | 15.625 |
| ProductPhase/ValidationCopy | 8572.100 | 8781.250 |
| ProductPhase/Decision | 181.201 | 468.750 |
| Boundary/Evaluate4096 | 8783.485 | 9281.250 |
| Boundary/Assert4096 | 7.773 | 15.625 |
| Boundary/Check4096 | 8836.794 | 9296.875 |
| Boundary/Check4097 | 5.482 | 0.000 |
| Original35/ArrayAndIntegerBounds | 9550.499 | 10000.000 |
| Original35/CallerMutation | 57.332 | 78.125 |
| Original35/ConservativeFlags | 41.920 | 46.875 |
| Original35/Complete | 1.307 | 0.000 |
| Ownership/ApiRequired | 1.554 | 0.000 |
| Ownership/DecoratedInteger | 37.478 | 78.125 |
| Ownership/DecoratedBoolean | 36.887 | 31.250 |
| Ownership/DecoratedString | 55.056 | 125.000 |
| Ownership/RootScalars | 29.023 | 15.625 |
| Ownership/ObservationScalars | 756.166 | 781.250 |
| Ownership/PassiveDescriptors | 56.715 | 46.875 |
| Ownership/SequentialIsolation | 185.906 | 171.875 |
| Ownership/PlainControls | 120.817 | 125.000 |
| Ownership/Complete | 9.245 | 0.000 |
| Stage/ExistingAll | 17962.738 | 19625.000 |
| Stage/ExistingFull | 18089.052 | 19875.000 |
| Run/CompleteFlow | 18733.648 | 20750.000 |
