# Combined regression timing — boundary group dominates a near-limit diagnostic

Approved one fresh ordinary-token PS7 child afterf7b0a7e3. Diagnostic copies add
only phase/case markers and adjust relative product-import paths; current product,
original assertions and test order retained. Configured30000ms, no preflight
outside the existing combined path, warmup, retry, fallback, repair, reviewer,
old-slot use or timeout increase. This is a throwaway diagnostic, not acceptance.

## Result and interpretation

PS7.6.5 Core x64/admin=False; exit0,total29783ms,TimedOut=false,owned-host-exited,
stderr empty,CaptureValid=true,Complete=true,FormalAcceptance=false.
Raw188 lines:118 timing events,70 ordinary lines including57 PASS markers and
diagnostic LIMIT. All59 start/end pairs closed; no active group or partial tail.
The57 PASS markers are3 new reflection checks +8 fixture groups +35 original
groups +10 ownership groups +1 final wrapper marker, not57 distinct test cases.

| Measured region | Start–end relative to first Run marker (ms) | Elapsed ms |
| --- | --- | ---: |
| ReflectionReuse stage | 223.609–527.084 | 303.475 |
| Fixture group span | 616.440–1505.138 | 888.698 |
| Original35 group span | 1587.987–25658.285 | 24070.298 |
| Original35/ArrayAndIntegerBounds | 5823.606–25392.446 | 19568.840 |
| Ownership regression span | 25677.054–29351.061 | 3674.007 |
| Ownership/ObservationScalars | 26117.818–28224.831 | 2107.012 |
| ExistingAll stage | 527.352–29351.882 | 28824.530 |
| CombinedRegression Run | 0.000–29352.623 | 29352.623 |

Regions nest: do not sum the table. Group spans include intervening markers/
headers; individual durations include test body, original PASS output and start
marker serialization/flush overhead. End-marker serialization is outside that
duration. Run scope differs from process launch/wait/drain total. Ticks come
from the same monotonic Stopwatch clock; times are elapsed wall time, not CPU.

ArrayAndIntegerBounds is about65.7% of process total and66.7% of Run elapsed,
the largest measured group here. The group includes invalid numeric/array/
shadow-Length checks,4096 construction, full4096 comparison and4097 refusal.
These markers do NOT subdivide that group, so19.569s must not be called the
4096 API evaluation alone. Earlier isolated timings suggest an evaluation
candidate for finer tracing, not precise current subphase attribution.

Ownership ObservationScalars started at26.118s and completed in2.107s. The full
Run ended29.353s; process total left only217ms below30000. This demonstrates
cumulative work near the gate in this diagnostic; it does not show a permanent
hang, a stable throughput guarantee, the exact old timeout statement or a unique
cause for old S3. Previous runs varied materially. One completed instrumented
copy cannot overwrite the original uninstrumented S3 timeout or accept candidate.

## Fidelity, parser and integrity

All four saved diagnostic copies were reversed in memory to exact original
source before launch: remove added timing helpers/markers, restore three relative
product-import paths. Original group bodies/assertions/order are untouched.
Product is imported from parent directory with fixed candidate6CF0D354, not a
modified diagnostic product. No original test file was edited or directly run.

Static parsing passed for four copied scripts/controller. New parser freezes
actual original35 Context/ApiRequired/HOST order, correcting the OLD inventory
only in this new parser; prior controller/index/raw remain byte-identical.
It checks ordered timing records, monotonic ticks/frequency, nested matching
pairs, expected ordinary lines, complete/prefix counts, host identity/role and
stderr policy. The completed run exercised its success path; timeout/partial-tail/
error cleanup remain unexercised, not claimed tested.

138 immutable pins matched before/after:123 inherited +10 previous packet anchors
+current product +four diagnostic copies. All59 receipt durations were independently
recomputed from saved raw timestamp pairs and matched;188/118/70/57 counts checked.
No rerun or historical verdict revision during read-only verification. Hashes
show consistency of retained files, not independent attestation.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | `C7E0A18A68AA05C0CD915DF4393775991FF78F1CFEB99DE915382338F44F82BB` |
| run-once.ps1 | `F65D4384396B5133C96FD8EF85EDBFAC2ECEAF6A67E6D595D47E6C4E58CB0049` |
| invocations.json | `D914EDE73FAEF84F05668AFAAE772CE7F1278E45D8B77DA4AFBED9D47C8CC0B4` |
| D1.stdout.txt | `7B5DEA6DE0A9A9923706B6F7CF3E187ABAE317B046A7C5A29CCC54E19BAC227B` |
| D1.stderr.txt | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |
| probe.ps1 (portable copy) | `7DAA9BD47D98374F8114DE5771C1D8D4989476258C015EFA6DC89978AC228B40` |
| offline-supervision-policy.tests.ps1 (portable copy) | `1C2E9FCA26B4B3C64D23271CCBBC5381C23C5CE60EB63C8322075F1C0EFFAF57` |
| offline-supervision-policy.ownership.tests.ps1 (portable copy) | `03A8B7A39F15BA2AF234DC3B0D2DF638DBDD16C7287BE7854357E437D1C519DE` |
| offline-supervision-policy.reflection-reuse.tests.ps1 (portable copy) | `BED41B65AB4B785F808559FE1487DE9F57D55EE4E440F931A027F4514ECEA626` |

Ignored packet `.superpowers/sdd/2026-10-17-p1-combined-timing/` retains scope,
controller/manifest/index and raw streams; all earlier packets retained. Portable
copies/report/status/recovery are development-branch checkpoints, not main.
Cloning restores inspectable diagnostic tooling/measurements, not ignored raw,
private runtimes or authorization to rerun. Checkout line conversion can change
byte hashes; these identify measured local bytes.

## Handoff and next scope

New diagnostic1/1 spent. Product6CF0D354/original35/ownership suites unchanged.
Previous reflection-reuse3/4/one frozen (PS5 never run), extended-view3/7/four
frozen, old R3 and S3 timeouts, review0/1 ineligible and all other budgets remain.
I1/P1 unresolved; no formal regression/PS5/main/C12/project/live/runtime/release
acceptance and no product fix this turn.

Next recommendation: narrowly subdivide ArrayAndIntegerBounds in the SAME
combined order (invalid checks, construction,4096 evaluation/assertion,4097 refusal)
with unchanged candidate/assertions/30s cap. This can distinguish expensive API
work from other work inside the dominant group. It needs its own explicit bounded
allowance, not automatic reuse of frozen slots, a retry of this completed run,
or a raised limit. Do not tune away defensive validation based on aggregate timing.

## Full saved pair inventory

Offsets relative to first Run marker; rounded milliseconds. Stage/Run entries
are inclusive and appear when their closing marker is reached.

| Kind/name | Start ms | End ms | Elapsed ms |
| --- | ---: | ---: | ---: |
| Reflection/InputMutation | 273.715 | 447.578 | 173.864 |
| Reflection/DecorationAfterClean | 447.792 | 486.516 | 38.724 |
| Reflection/DescriptorAfterClean | 486.700 | 526.538 | 39.838 |
| Stage/ReflectionReuse | 223.609 | 527.084 | 303.475 |
| Ownership/FixtureConstructor | 616.440 | 628.264 | 11.825 |
| Ownership/FixturePrimitiveMatrix | 628.504 | 700.276 | 71.772 |
| Ownership/FixtureDescriptors | 700.516 | 713.163 | 12.647 |
| Ownership/FixtureInputIdentity | 713.397 | 725.756 | 12.359 |
| Ownership/FixtureIsolation | 725.991 | 734.324 | 8.333 |
| Ownership/FixtureWrappedInputIdentity | 734.552 | 738.289 | 3.736 |
| Ownership/FixtureExtendedView | 738.565 | 750.013 | 11.448 |
| Ownership/FixtureComplete | 750.243 | 1505.138 | 754.895 |
| Original35/ApiRequired | 1587.987 | 1590.308 | 2.321 |
| Original35/EmptyPrepared | 1590.658 | 1607.519 | 16.861 |
| Original35/NormalStages | 1607.740 | 1680.359 | 72.620 |
| Original35/AbsoluteDeadline | 1680.611 | 1717.767 | 37.157 |
| Original35/StopBoundary | 1718.100 | 1789.248 | 71.148 |
| Original35/TotalDeadline | 1789.506 | 1826.722 | 37.216 |
| Original35/StartupConsumesBudget | 1826.961 | 1846.070 | 19.109 |
| Original35/EqualTimeSequence | 1846.325 | 1871.805 | 25.480 |
| Original35/BackwardTime | 1872.042 | 1891.708 | 19.666 |
| Original35/SequenceRefusal | 1892.170 | 1909.081 | 16.911 |
| Original35/SessionClockRefusal | 1909.312 | 1947.377 | 38.065 |
| Original35/ActorDirection | 1947.622 | 1967.486 | 19.864 |
| Original35/ArmPrerequisites | 1967.737 | 2052.056 | 84.318 |
| Original35/ReleaseOnce | 2052.287 | 2126.366 | 74.079 |
| Original35/ProgressMonotonic | 2126.606 | 2328.409 | 201.804 |
| Original35/ExitBeforeEOF | 2328.719 | 2363.584 | 34.866 |
| Original35/UnknownExit | 2363.901 | 2391.149 | 27.248 |
| Original35/NonzeroExit | 2391.481 | 2421.946 | 30.465 |
| Original35/DataBounds | 2422.262 | 2528.205 | 105.943 |
| Original35/ControlBounds | 2528.454 | 2634.694 | 106.240 |
| Original35/ReadFault | 2634.959 | 2684.494 | 49.536 |
| Original35/SupervisorLost | 2685.268 | 2831.787 | 146.518 |
| Original35/LauncherLost | 2832.448 | 2982.145 | 149.696 |
| Original35/FailureSticky | 2983.281 | 3171.911 | 188.630 |
| Original35/PhaseBudgets | 3173.027 | 4337.820 | 1164.793 |
| Original35/CleanupUnknown | 4338.776 | 4724.106 | 385.331 |
| Original35/PersistRefused | 4725.215 | 5093.446 | 368.231 |
| Original35/TerminalTrailing | 5093.862 | 5237.708 | 143.846 |
| Original35/InvalidScenario | 5238.705 | 5462.211 | 223.506 |
| Original35/InvalidObservation | 5463.094 | 5666.470 | 203.377 |
| Original35/PassiveDescriptors | 5667.207 | 5822.881 | 155.674 |
| Original35/ArrayAndIntegerBounds | 5823.606 | 25392.446 | 19568.840 |
| Original35/CallerMutation | 25393.479 | 25568.079 | 174.600 |
| Original35/ConservativeFlags | 25568.534 | 25655.209 | 86.675 |
| Original35/Complete | 25655.615 | 25658.285 | 2.670 |
| Ownership/ApiRequired | 25677.054 | 25679.374 | 2.319 |
| Ownership/DecoratedInteger | 25679.624 | 25784.602 | 104.978 |
| Ownership/DecoratedBoolean | 25785.292 | 25882.586 | 97.294 |
| Ownership/DecoratedString | 25883.203 | 26027.991 | 144.788 |
| Ownership/RootScalars | 26028.453 | 26117.575 | 89.122 |
| Ownership/ObservationScalars | 26117.818 | 28224.831 | 2107.012 |
| Ownership/PassiveDescriptors | 28225.359 | 28363.105 | 137.745 |
| Ownership/SequentialIsolation | 28363.663 | 28978.231 | 614.568 |
| Ownership/PlainControls | 28978.704 | 29318.560 | 339.857 |
| Ownership/Complete | 29319.123 | 29351.061 | 31.938 |
| Stage/ExistingAll | 527.352 | 29351.882 | 28824.530 |
| Run/CombinedRegression | 0.000 | 29352.623 | 29352.623 |
