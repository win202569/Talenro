# Complete-order CPU timing diagnostic — completed once; original acceptance still blocked

User approved one fresh ordinary-token PS7 child after148c3d51, configured30000ms.
No warmup, preflight test run, retry, fallback, PS5, reviewer, product/original-test
edits, old-budget use or timeout increase. New diagnostic budget1/1 is spent.
This is an instrumented diagnostic, NOT formal acceptance or independent review.

## Saved result

PS7.6.5 Core x64/admin=False, exit0, total10226ms, cumulative CPU12625.000ms;
TimedOut=false, owned-host-exited, stderr empty, CaptureValid=true, Complete=true,
FormalAcceptance=false. Started2026-10-17T19:29:34.5490617Z.

Raw218 lines comprise134 timing events and84 ordinary lines. All67 event pairs
closed, no active scopes or partial tail. Ordinary output contains64 PASS markers:
shape6 + reflection3 + fixture8 + original35 + ownership10 =62 test groups,
plus two original wrapper completion markers. Exact original Context/ApiRequired/
HOST ordering is preserved. All assertions in the complete copied path finished.

| Region (nested, not additive) | Elapsed wall ms | Cumulative process CPU delta ms |
| --- | ---: | ---: |
| Stage/ShapeLookup | 510.616 | 656.250 |
| Stage/ReflectionReuse | 130.177 | 187.500 |
| Ownership/FixtureComplete | 775.044 | 687.500 |
| Original35/ArrayAndIntegerBounds | 4932.728 | 6406.250 |
| Ownership/ObservationScalars | 542.579 | 687.500 |
| Ownership/SequentialIsolation | 132.995 | 140.625 |
| Stage/ExistingAll | 8928.622 | 11171.875 |
| Stage/ExistingFull | 9079.004 | 11390.625 |
| Run/CompleteFlow | 9822.441 | 12187.500 |

ArrayAndIntegerBounds is about48.2% of process wall and
50.7% of process CPU in THIS run; it is the largest group.
It still includes invalid-input checks,4096 construction, API evaluation,
result comparison and4097 refusal;4.933s must NOT be called evaluation alone.
No product internals or per-observation/per-helper timing were added.

The original35 span from first group START through Complete END is6938.261ms.
This run completed rather than reproducing the earlier30s timeout.
That is evidence of a completed diagnostic path, not a controlled speedup,
a unique root cause, a stable latency guarantee or clearance of the old failure.

## Meaning and measurement limits

Candidate is unchanged953CC2A8. Historical same-candidate original full flow
timed out30071ms. The diagnostic differs by boundary instrumentation, copied
script paths and extra diagnostic limit output; runtime/host/order/cache/JIT/GC
conditions were not controlled as a matched repeated experiment.
Do NOT infer that the product optimization caused the difference, that markers
speed up the program, or that the previous timeout was a false failure.

CPU is cumulative across ALL child process threads, not just the executing
PowerShell test. Background runtime/JIT/GC work can fall within a group interval.
CPU deltas may exceed wall, are NOT CPU utilization percentages, and cannot
isolate application instruction cost from runtime work or establish contention.
Observed sample resolution is coarse (15.625ms increments here): a zero group
CPU delta does not mean no CPU work.

Each event samples monotonic wall ticks first, then process CPU ticks, then
serializes/flushes one console line. These samples are not atomic. Start-line
sampling/serialization/flush and end sampling affect the interval; some end-line
serialization lies outside its own pair. Boundary sampling overhead is unmeasured.
All parent regions include child regions; DO NOT sum nested rows.
Run marker scope9822.441ms differs from process start/wait/drain total10226ms.

Original acceptance remains failed/incomplete until separately authorized
original-entry verification meets the unchanged gate. No original PS7/PS5
formal verification occurred this turn; do not relabel this diagnostic as one.

## Fidelity and integrity

Portable throwaway copies:
[probe](../diagnostics/isolated-fuzz/complete-flow-cpu-timing/probe.ps1)
and five adjacent copied scripts. Complete sequence remains shape6, reflection3,
fixture8, original35, ownership10 in ONE process.
All six copied scripts statically parsed before launch. Removing boundary
helpers/events, restoring product/wrapper relative paths and removing the one
diagnostic LIMIT restores each of the four original test scripts and both
original wrappers exactly after line-ending normalization. P1-Check and all
test bodies/assertions are unchanged; no hot-helper mode branches introduced.
All product imports point to the unchanged parent product, not an instrumented
module. Prior diagnostic folders and raw records were not edited.

Frozen196 pins = inherited183 + seven latest-targeted packet anchors +
six new diagnostic copies. Engine bytes checked before/after. Slot consumed
before child Start; ordinary token enforced; only the owned process handle is
used for termination; no elevated test run.
After completion, all196 pins, controller, saved index and both streams were
read-only verified. All67 wall AND CPU durations independently recomputed from
saved event pairs and matched;218/134/84/64 counts and closed stack checked.
Successful path was exercised; timeout, malformed-tail and CPU-read-failure paths
were NOT exercised or claimed tested.

Ignored local packet:
.superpowers/sdd/2026-10-17-p1-complete-flow-cpu-timing/
retains scope/controller/frozen inventory/index/raw, deliberately not committed.
Digests identify measured local bytes, not independent attestation; checkout line
conversion may change byte hashes.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 55E22CDC9AF28AFF83468406855D5A038774616A6B64E11B35CEEF6A12429898 |
| run-once.ps1 | CD63EF665FE7CB5DE1463CBE858F51A020E6986FB5E10C5F045D7965013CAE59 |
| invocations.json | F76944DB16F778E84BAEBCEEF774A57BE69EE474BF47036283F4845E15B3E8FB |
| D1.stdout.txt | 78ADEE5B6710382ACB355D31DAA168207DFEB68AE3B20B56622D3F6DBC8A8B70 |
| D1.stderr.txt (empty) | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| probe.ps1 | 04526277000A8D362458B9B7BFC6F84D4BE14CA2C0BF6F63011F6EE175DB6A88 |
| existing.ps1 | 1C50939CF715FA2E3E6D39B3ECBF0279580022032662013A58E9E44F56DEDE9F |
| shape tests copy | 1BC8EBAB350E1CDC9B9A18ABC143468FDF0FBD087BD99910E5F85E037CF8CC3F |
| original35 copy | 69DD632FDFCE1A8F4E1FFF440E9F508C7DC166325AE0B8514CCFB1CC32FB0639 |
| ownership copy | C6D43775268D1192C4E693A9A9ADC1E42CBE250E33BDEFA05E01C82E696901EB |
| reflection copy | AE8EA5EED04C830FCF4996C1687CAF5B47C59680E008E63D34776EFB15A06760 |

PS7 executable SHA256:
362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.
Product SHA256:
953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.

## Handoff

New diagnostic1/1 spent; targeted2/2 spent; shape2/3/one frozen;
prior independent full1/2/one frozen; reflection3/4/one frozen;
extended-view3/7/four frozen; review0/1 unused/ineligible; other budgets unchanged.
I1/P1 unresolved, candidate unaccepted. No main/C12/project/live/runtime/release
acceptance, no independent review or automatic reviewer eligibility.

Portable scripts/report/status/recovery are saved on the existing development
branch. Clone restores inspectable work, NOT ignored raw, cached engines,
credentials, budget renewal or replay permission. Do not rerun from a fresh
computer solely because the diagnostic entry is present.

Next keep product fixed and design separately scoped original-entry verification
or a controlled instrumentation/condition comparison. Existing data do not justify
another product optimization, unique-cause claim, automatic unchanged retry,
use of frozen PS5 slots, higher timeout or acceptance promotion.

## All completed boundary pairs

Elapsed wall and cumulative all-thread CPU delta in milliseconds. Inclusive
parents overlap children. Coarse zero CPU samples are not zero work.

| Kind/name | Elapsed wall ms | CPU delta ms |
| --- | ---: | ---: |
| Shape/RootOrdinal | 137.565 | 125.000 |
| Shape/FieldBounds | 13.306 | 15.625 |
| Shape/ObservationOrdinal | 79.919 | 62.500 |
| Shape/KindIsolation | 80.282 | 156.250 |
| Shape/RepeatedCalls | 115.809 | 218.750 |
| Shape/DescriptorSafety | 17.442 | 31.250 |
| Stage/ShapeLookup | 510.616 | 656.250 |
| Reflection/InputMutation | 49.998 | 93.750 |
| Reflection/DecorationAfterClean | 31.443 | 31.250 |
| Reflection/DescriptorAfterClean | 17.908 | 15.625 |
| Stage/ReflectionReuse | 130.177 | 187.500 |
| Ownership/FixtureConstructor | 11.173 | 15.625 |
| Ownership/FixturePrimitiveMatrix | 56.951 | 62.500 |
| Ownership/FixtureDescriptors | 10.886 | 0.000 |
| Ownership/FixtureInputIdentity | 11.491 | 15.625 |
| Ownership/FixtureIsolation | 7.099 | 0.000 |
| Ownership/FixtureWrappedInputIdentity | 2.837 | 0.000 |
| Ownership/FixtureExtendedView | 9.311 | 0.000 |
| Ownership/FixtureComplete | 775.044 | 687.500 |
| Original35/ApiRequired | 2.513 | 15.625 |
| Original35/EmptyPrepared | 17.173 | 31.250 |
| Original35/NormalStages | 67.522 | 62.500 |
| Original35/AbsoluteDeadline | 34.199 | 46.875 |
| Original35/StopBoundary | 73.611 | 140.625 |
| Original35/TotalDeadline | 44.055 | 78.125 |
| Original35/StartupConsumesBudget | 25.905 | 31.250 |
| Original35/EqualTimeSequence | 31.774 | 78.125 |
| Original35/BackwardTime | 36.339 | 93.750 |
| Original35/SequenceRefusal | 37.775 | 62.500 |
| Original35/SessionClockRefusal | 51.245 | 125.000 |
| Original35/ActorDirection | 25.290 | 31.250 |
| Original35/ArmPrerequisites | 108.211 | 203.125 |
| Original35/ReleaseOnce | 98.404 | 203.125 |
| Original35/ProgressMonotonic | 207.813 | 437.500 |
| Original35/ExitBeforeEOF | 34.652 | 62.500 |
| Original35/UnknownExit | 28.293 | 78.125 |
| Original35/NonzeroExit | 31.436 | 31.250 |
| Original35/DataBounds | 111.345 | 93.750 |
| Original35/ControlBounds | 120.362 | 125.000 |
| Original35/ReadFault | 26.004 | 31.250 |
| Original35/SupervisorLost | 25.032 | 15.625 |
| Original35/LauncherLost | 32.311 | 15.625 |
| Original35/FailureSticky | 35.635 | 31.250 |
| Original35/PhaseBudgets | 206.860 | 265.625 |
| Original35/CleanupUnknown | 105.777 | 109.375 |
| Original35/PersistRefused | 97.812 | 93.750 |
| Original35/TerminalTrailing | 39.616 | 62.500 |
| Original35/InvalidScenario | 59.002 | 62.500 |
| Original35/InvalidObservation | 50.757 | 46.875 |
| Original35/PassiveDescriptors | 38.223 | 46.875 |
| Original35/ArrayAndIntegerBounds | 4932.728 | 6406.250 |
| Original35/CallerMutation | 54.660 | 46.875 |
| Original35/ConservativeFlags | 32.870 | 15.625 |
| Original35/Complete | 1.749 | 0.000 |
| Ownership/ApiRequired | 1.306 | 0.000 |
| Ownership/DecoratedInteger | 28.997 | 31.250 |
| Ownership/DecoratedBoolean | 24.028 | 15.625 |
| Ownership/DecoratedString | 47.681 | 46.875 |
| Ownership/RootScalars | 29.948 | 31.250 |
| Ownership/ObservationScalars | 542.579 | 687.500 |
| Ownership/PassiveDescriptors | 30.994 | 31.250 |
| Ownership/SequentialIsolation | 132.995 | 140.625 |
| Ownership/PlainControls | 87.474 | 78.125 |
| Ownership/Complete | 0.916 | 0.000 |
| Stage/ExistingAll | 8928.622 | 11171.875 |
| Stage/ExistingFull | 9079.004 | 11390.625 |
| Run/CompleteFlow | 9822.441 | 12187.500 |
