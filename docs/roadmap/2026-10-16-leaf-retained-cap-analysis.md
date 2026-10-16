# Leaf retained compile001 — saved-only cap/chronology analysis

## Scope and evidence

User approved continuation with saved-only analysis after ba1b29ae8c880a180423e3606cf13ff1f987a203.
No native Run, Go compile, generated binary execution, SDK audit, PID lookup,
cache reuse, cap change or code/test change occurred. The11 original manifest
artifact hashes/lengths and12 plan-bound helper/native source hashes matched.
Original manifest/receipt bindings remain respectively
`6572BB08617612DD5065109345926715C3288589D274B635D07A41FE7821DC0D` and
`6EDD316680499D16D29D4FD6DFDF3A62B66CC206887D7DD6456DB904BAE47702`.

This analysis follows the producer/validator chain, not a modified replay with
the cap flag cleared. Compile001 remains STOPPED; LifetimeEvidenceComplete,
ObservationPassed and Acceptance remain false. No historical result is replaced.

## Recorded chronology

Offsets below are differences between recorded UTC timestamps and receipt start,
not exact OS creation/exit times or an independently monotonic whole-run clock.

| Recorded event | Offset from native receipt start |
| --- | ---: |
| First registry query | 1276.9870ms |
| First admission:conhost.exe metadata | 1287.0019ms |
| Second admission:compile.exe metadata | 21387.5656ms |
| Last query:number1362 | 23419.4592ms (end) |
| Admission32 completed | 23419.8823ms |
| Native receipt ended | 74944.8801ms |

First-query-start to last-query-end spans22142.2453ms on the recorded monotonic
clock. Last query to receipt end is51525.4209ms on recorded UTC. That latter
interval includes root completion/cleanup; it is not a measured compiler-only
tail. The report stores first/last and admission-source query metadata, not all
1362 query records, so no full cadence, missed-member count or continuous
concurrency curve can be reconstructed.

Admission image basenames:17 asm.exe,14 compile.exe,1 conhost.exe. These are
metadata for the32 retained entries only, not identities inferred for historical
unknown members. All32 raw admission states are running, root-wait/final states
exited, final exit codes0. These are phase observations, not exact exit times;
they do not prove every per-entry validator invariant or uncaptured lifecycle.

The maximum Assigned value among the saved admission-source queries is3, whereas
root-wait accounting reports545 total processes. Neither is a count of captured
identities or a global maximum concurrency claim. In particular, the32 cap is
not evidence that32 processes were simultaneously active.

## Producer-to-refusal explanation

In retained-lifetime.cs, Poll uses a lifetime-wide seen set and increments
Attempts for each new eligible identity before opening. There is no eviction,
released-slot reuse or new sampling cohort. It marks AdmissionBudgetExhausted
at attempt32 unconditionally, even if that admission succeeds and no33rd arrives.
Subsequent Poll calls return before QueryMembers. Thus QueryCount1362 and the
last query near23.42s are consistent with the preserved source and admission32.
The source establishes why queries stop; no live reproduction is needed here.

Root-wait/post-cleanup Snapshot does not use that Poll cap early-return. It still
reads already-owned handles, followed by Dispose releases. Accordingly, late
snapshots of early entries exist even though late new-member sampling does not.
Calling the entire last51.53s unobserved would be inaccurate.

The leaf validator rejects AdmissionBudgetExhausted before per-entry validation,
and separately rejects Attempts>=32 even with an inconsistent cleared flag.
The controller then saves the stopped result. Incomplete=false means no separate
producer Error path was recorded; it does not override the cap. This is the
specified fail-closed diagnostic policy, not evidence of compiler failure or a
newly discovered cap-implementation defect.

## Diagnostic usefulness and next scope

The sample establishes compile exit0 and early retained samples with final
exited0 observations and successful release counters. It cannot explain the
historical root-wait OpenProcess87:the original core list/rows here are empty,
and new-member sampling stops well before this compilation ends. Increasing the
cap alone gives no guarantee of capturing a timing-dependent end-phase member;
it would also change the agreed resource/evidence policy. A warm-cache retry
would change the workload and would not repair this recorded observation.

Recommended next scope is a design discussion for a bounded diagnostic sample
targeted at the end-phase question, with explicit start trigger, coverage limits,
identity/ownership rules and unchanged ordinary acceptance refusal. Compare its
usefulness with the existing controlled-child fixture before implementing.
Do not select a larger cap, delayed trigger, eviction scheme or live rerun by
implication. A concrete design and fake-only tests require their own agreed
scope; real execution requires a separate fresh identity and approval.

No broader suite was rerun and no production/ordinary acceptance gate changed.
Historical matrix003 remains2PASS/1FAIL/23UNRUN; error87 remains unexplained.

## Preservation and recovery

Ignored derived analysis:.superpowers/diagnostics/2026-10-16-leaf-retained/
cap-analysis-001/saved-chronology.json; SHA256
`694D3CFF0040AE8779D74003BDBAE325BB47D646F5F18D0EBDCF411F6C4A0AAE`.
It binds original manifest/receipt hashes,12 checked source pins,32 derived entry
rows and arithmetic/grouping facts; LiveCalls=0. It is not a new process receipt
or acceptance artifact. Raw/derived local evidence remains ignored, not pushed.

Recover the existing development branch and [compile001 account](2026-10-16-leaf-retained-compile-001.md),
[entry/portable tests](2026-10-16-leaf-retained-compile-entry.md), current-status.md
and [repository recovery](../runbooks/repository-recovery.md). This portable
account preserves facts/limits without requiring the previous machine's cache;
Git does not restore raw evidence, SDK, binaries or live process state.
