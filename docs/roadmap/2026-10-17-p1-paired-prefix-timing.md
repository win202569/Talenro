# Paired prefix diagnostic — added helper branches did not reproduce major slowdown

User approved two fresh ordinary-token PS7 children aftercf9590b0, fixed order
A then B, configured30000ms each, no warmup/retry/fallback. Same prefix workload,
product/assertions unchanged. Any failed/invalid/timed-out slot would freeze the
next. Both allowed slots completed; no third child or old-budget use.

## Result

| Variant | Helper | Process wall ms | Process cumulative CPU ms | Original31 span ms | Exit |
| --- | --- | ---: | ---: | ---: | ---: |
| A first | Original P1-Check | 6615 | 9656.250 | 4395.405 | 0 |
| B second | Three added mode conditions | 4407 | 5468.750 | 1781.390 | 0 |

Both PS7.6.5 Core x64/admin=False,hidden/no-profile/no-shell,stderr empty,
owned-host-exited,TimedOut=false,CaptureValid=true,Complete=true,
FormalAcceptance=false. Each raw146 lines:90 timing events,56 ordinary lines,
44 PASS markers,45 closed pairs,no active scopes or partial tail. PASS count is
3 reflection checks +8 fixture groups +31 original groups +2 diagnostic completion
markers, not44 distinct cases. Prefix ending is explicit, not a fake full-suite
Complete. No ArrayAndIntegerBounds/4096/remaining original4/ownership10 ran.

CPU is the child process's TotalProcessorTime after exit: cumulative across its
threads, including startup/runtime/JIT/logging, not per-test CPU or percentage.
It may exceed wall elapsed when threads run concurrently. Wall includes process
start/wait/stream drain; source/hash checks happen outside that watch. No child
CPU samples within individual groups, contention trace,GC/JIT/thermal or global
load timeline were collected. These totals cannot identify where CPU was spent.

The extra-branch variant was faster in this single sequential pair, including
lower cumulative CPU. This does NOT establish that branches speed up validation
or have zero overhead. It DOES mean this experiment did not reproduce the proposed
large branch-associated slowdown. First-vs-second order, file/runtime/OS caches,
scheduling and other conditions remain uncontrolled. Fresh processes are not a
cold-system reset. No warmup was deliberately run; no order reversal/repetition
was authorized or attempted. Earlier timed-out records remain unchanged.

Conclusion: do not treat three diagnostic conditions as a demonstrated primary
cause of the previous pre-boundary slowdown. A unique cause is still unknown.
This prefix result cannot accept product performance or explain unfinished4096
evaluation in the previous run.

## Workload and fidelity

Both run the identical reflection3 and fixture8 before first31 original35 groups.
Original35 diagnostic copy emits an explicit prefix-stop marker/return just
before ArrayAndIntegerBounds. The common ownership dispatcher returns immediately
after that prefix; it never imports/runs ownership regression on this path.
The common wrapper labels PrefixTiming, not full acceptance.

A and B initialize the same script-local mode=None and differ ONLY in P1-Check
function body. B uses the previously introduced three conditions;None remains
unchanged for the entire prefix, so no subphase markers execute. A uses the
original direct helper. Original product call, equality assertion, helper scope,
test inputs/bodies/order retained. The diagnostic helper's runtime shape is the
deliberate intervention, not a product change. The saved A copy reverses its
common mode initialization and early prefix return exactly to the prior group
diagnostic source. Reflection copy is byte-identical; common dispatcher/wrapper
changes only select A/B and terminate/label the same subset.

Static parsing passed for five probe scripts and controller before launch. Saved
A/B files were checked to differ only by helper replacement. Every158 immutable
pin matched before/after:147 inherited +five previous packet anchors +five new
portable scripts +engine executable. Product6CF0D354 and original35/ownership
suites unchanged. Frozen A/B order, same engine bytes/arguments/working directory,
ordinary-token check,90-event/56-line inventories and deadline used both times.
Each of90 receipt durations across A/B independently recomputed from raw event
pairs; counts and closed stacks matched without replay or modification.

Engine SHA256 pinned at preparation and checked before/after both:
`362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139`.
Both headers PS7.6.5; PE file version7.6.5.500 from preceding read-only audit.
A start2026-10-17T19:00:26.9986115Z;B start19:00:40.4874931Z.
Success paths exercised; timeout/CPU-read-failure/partial-tail paths not exercised
or claimed validated.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | `4D3A44E6CD5A4E7801E4FAB6046B55020231469695C93C6E92D84674EE14C1F5` |
| run.ps1 | `1C3013DFDA33F9A85814A73EAC331B6947D27D0282B116DA4B48C7A5FF6F5D4D` |
| invocations.json | `B60C4850B5017E2D2BCD474A262147F4D37C1EC20735AD83DF91DD09092B1687` |
| A.stdout.txt | `30662ECF29494C424B32E718006A08BFF234F44DB3A2A7772D20BF952AA4F00E` |
| A.stderr.txt | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |
| B.stdout.txt | `C368D12121FE743D38F74FE7A52CA2FD1F81F5C8F7CCBD723055AC3013917A01` |
| B.stderr.txt | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |
| offline-supervision-policy.ownership.tests.ps1 (portable) | `85EC32A986A5839BF69F2CF2DE28EFA37CB8ED8741870E0FB8A2CB3767DA5EC9` |
| offline-supervision-policy.reflection-reuse.tests.ps1 (portable) | `BED41B65AB4B785F808559FE1487DE9F57D55EE4E440F931A027F4514ECEA626` |
| offline-supervision-policy.tests.A.ps1 (portable) | `4D46B0C5E6D17A135B6C4BDEAF9FA3B3606B3943C8620BAB076F051059680CBD` |
| offline-supervision-policy.tests.B.ps1 (portable) | `C437F3A7FDE520B49523C3CFD0F8F2EF89053455B07286A631F3128290B4C84B` |
| probe.ps1 (portable) | `BCBF6FB3F689B1DB4306DBBABA2CEC99FFD01CE49CFB42BFAC1E56E2D1017EFD` |

Ignored packet `.superpowers/sdd/2026-10-17-p1-paired-prefix-timing/` retains
scope/controller/manifest/index/raw. All earlier evidence retained. Portable
scripts/report/status/recovery are development-branch checkpoints, not main.
Cloning restores inspectable work, not ignored receipts/runtime or rerun authority.
Digests identify measured local bytes, not independent attestation; checkout
line conversion can change them.

## Handoff

New2/2 spent. Product/original suites unchanged. Prior reflection-reuse3/4/one
frozen,extended-view3/7/four frozen,old failed/incomplete results and review0/1
ineligible remain. Candidate6CF0D354 unaccepted,I1/P1 unresolved. No formal full
regression,PS5,main,C12,project,live/runtime/release acceptance.

Do not continue product changes or branch-removal fixes based on a causality
claim this pair did not support. Next return to planning explicit full PS7/PS5
verification under the retained30s gate in controlled conditions, keeping
diagnostics outside the original check helper and recording engine/CPU/wall
receipts. Such execution needs separate authority; no automatic remaining-slot
resume,timeout increase or promotion of prefix PASS to suite acceptance.

## Paired completed durations

Milliseconds. Stage/Run are inclusive and overlap child groups; do not sum rows.

| Kind/name | A elapsed ms | B elapsed ms |
| --- | ---: | ---: |
| Reflection/InputMutation | 170.554 | 190.090 |
| Reflection/DecorationAfterClean | 38.094 | 48.298 |
| Reflection/DescriptorAfterClean | 33.149 | 30.705 |
| Stage/ReflectionReuse | 281.251 | 310.840 |
| Ownership/FixtureConstructor | 10.642 | 12.888 |
| Ownership/FixturePrimitiveMatrix | 63.135 | 71.870 |
| Ownership/FixtureDescriptors | 10.666 | 14.156 |
| Ownership/FixtureInputIdentity | 35.644 | 12.268 |
| Ownership/FixtureIsolation | 16.644 | 7.814 |
| Ownership/FixtureWrappedInputIdentity | 3.417 | 2.997 |
| Ownership/FixtureExtendedView | 11.183 | 10.472 |
| Ownership/FixtureComplete | 839.680 | 1310.985 |
| Original35/ApiRequired | 2.539 | 3.197 |
| Original35/EmptyPrepared | 17.213 | 24.445 |
| Original35/NormalStages | 76.464 | 104.850 |
| Original35/AbsoluteDeadline | 34.836 | 62.684 |
| Original35/StopBoundary | 77.035 | 104.176 |
| Original35/TotalDeadline | 34.599 | 46.072 |
| Original35/StartupConsumesBudget | 17.764 | 26.610 |
| Original35/EqualTimeSequence | 25.204 | 37.084 |
| Original35/BackwardTime | 19.704 | 26.764 |
| Original35/SequenceRefusal | 16.731 | 23.789 |
| Original35/SessionClockRefusal | 40.696 | 48.987 |
| Original35/ActorDirection | 20.330 | 26.378 |
| Original35/ArmPrerequisites | 81.134 | 104.327 |
| Original35/ReleaseOnce | 68.924 | 89.072 |
| Original35/ProgressMonotonic | 198.318 | 185.435 |
| Original35/ExitBeforeEOF | 29.528 | 27.317 |
| Original35/UnknownExit | 25.518 | 23.373 |
| Original35/NonzeroExit | 30.325 | 25.612 |
| Original35/DataBounds | 102.408 | 83.716 |
| Original35/ControlBounds | 384.153 | 105.153 |
| Original35/ReadFault | 163.014 | 22.642 |
| Original35/SupervisorLost | 127.510 | 18.555 |
| Original35/LauncherLost | 145.422 | 19.567 |
| Original35/FailureSticky | 181.998 | 23.852 |
| Original35/PhaseBudgets | 1081.711 | 175.853 |
| Original35/CleanupUnknown | 337.912 | 90.090 |
| Original35/PersistRefused | 367.489 | 80.587 |
| Original35/TerminalTrailing | 126.847 | 32.386 |
| Original35/InvalidScenario | 215.981 | 53.338 |
| Original35/InvalidObservation | 184.400 | 40.898 |
| Original35/PassiveDescriptors | 142.346 | 35.349 |
| Stage/ExistingAll | 5554.289 | 3349.182 |
| Run/PrefixRegression | 6059.417 | 3922.804 |
