# Read-only timing audit — instrumentation and host effects remain confounded

User approved read-only inspection after247a64bb. No product/test/probe/controller
edits, diagnostic execution, replay, warmup, reviewer, install, security change,
old-slot use or new test allowance. Only source/receipts/Git/current environment
were inspected; this report and recovery notes preserve the audit.

## Findings supported by saved evidence

1. Product and original35/ownership suites did not change between813a8c76 and
   247a64bb. Current hashes still6CF0D354/61A3BACF/C4905925. Both historical packets
   passed all138/147 immutable pins and their controller/raw hash checks.
2. Both runs report PS7.6.5 Core x64/admin=False, same engine path, hidden
   no-profile/no-shell process and configured30000ms bound. Starts are286.931s
   apart. Launch/wait/kill/drain/raw-save supervision block is text-identical
   (974 characters). Parser changes occur after watch.Stop; they do not explain
   elapsed time inside the child from a direct code-path perspective.
3. The original35 subphase copy reverses exactly to its prior copy with six
   targeted edits reversed. Ownership/reflection copies unchanged; wrapper only
   changes diagnostic LIMIT wording. All original assertions/call expressions
   remain. There is no added product call in the new boundary helper.
4. Instrumentation DOES change before-boundary work: script-local mode is
   initialized to None, and P1-Check evaluates THREE new conditional branches
   on every call, even outside the boundary group. With mode None they emit no
   subphase markers. The function's runtime shape also differs. This is a real
   uncontrolled variable; its overhead has not been quantified.
5. No new Subphase event is emitted before the boundary group. Existing
   per-group markers/order remain the same for the first31 original groups.
   Their measured slowdowns range1.184x–28.394x, not a uniform additive or
   multiplicative effect. This rules out claiming that emitted NEW boundary
   marker serialization alone accounts for earlier delays; it does NOT rule
   out helper instrumentation, host scheduling, runtime/JIT or interactions.
6. First31 original groups span4234.894ms in the prior run and15911.312ms in the
   next, difference11676.418ms (span includes gaps). Boundary group begins at
   5823.606 versus18181.547ms from the Run marker. The4096 evaluation is not yet
   executing during those differences. Its later missing end marker cannot
   explain the already-observed pre-boundary slowdown.

No unique cause has been established. An input-validation defect, product
performance regression, stable metadata-cache speedup, CPU contention, power
throttling or instrumentation-dominant overhead is NOT proved by this audit.
The broad slower pattern and changed helper mean the two runs are not a
controlled before/after performance comparison.

## Measurement conditions: what is and is not recorded

| Item | Evidence and limit |
| --- | --- |
| Historical engine label | Both child stdout headers say PS7.6.5 Core x64; executable bytes were not hashed at each launch |
| Historical privilege | Both controller parents and child headers report admin=False |
| Historical deadline | Same configured30000ms; successful29783ms vs timeout30080ms total are separate outcomes |
| Historical CPU time/load | Not captured; wall-clock timers cannot recover it |
| Historical memory/GC/JIT/contention/thermal state | Not captured; cannot reconstruct from these receipts |
| Historical power scheme | Not captured |
| Current snapshot time | 2026-10-17T18:55:13.2556976Z, AFTER both runs |
| Current OS | Microsoft Windows NT10.0.26200.0, from .NET Environment |
| Current visible logical processors | 16; not a utilization or available-CPU measure |
| Current engine file version | 7.6.5.500; not proof bytes were identical at historical launch |
| Current engine SHA256 | `362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139` |
| Current active power scheme | Balanced, read with powercfg; not historical state or evidence of throttling |
| Current CIM CPU/memory query | Blocked by environment; no privileged retry or security change |

The initial CIM query failed and a dependent expression had no OS object; no
CPU/memory/boot data from that attempt was treated as evidence. A separate
successful .NET/file read supplied only the basic current values shown above.
Current snapshots cannot substitute for measurements taken during the child run.
No process command lines, full environment dump, credentials or unrelated private
files were inspected.

## Outcome and proposed next direction

Keep the saved partial verdict: new subphase diagnostic was1/1 spent, timed out
during Evaluate4096 with no completed duration/assertion; prior group diagnostic
was1/1 spent and completed once. Prior reflection reuse3/4/one frozen,
extended-view3/7/four frozen, old failures and review0/1 ineligible unchanged.
Candidate6CF0D354 remains unaccepted; I1/P1 unresolved. This audit consumes ZERO
test processes and grants no acceptance/main/C12/project/runtime authority.

Before further product optimization, isolate the instrumentation variable in a
separately approved short paired experiment. Prefer boundary-only outer timing
or a wrapper that does not add conditions to P1-Check for all preceding tests;
retain identical call/assertion scope, inputs and check order. Record child CPU
time alongside wall time and precisely identify executable bytes/configuration.
Counterbalance/repeat only if explicitly budgeted; one pair cannot establish
stable causality. Do not automatically alter tooling, run that experiment, resume
frozen slots, lengthen the cap or treat split subsets as full-suite acceptance.

All prior packets/controllers/indexes/raw preserved. Portable audit/status/recovery
on the current development branch enable a new computer to understand why no
further optimization was attempted; ignored receipts/runtime do not clone, and
cloning grants no replay authority. Documents changed only to record findings.

## All31 matched original groups

Elapsed wall milliseconds. Ratio is current/prior for these two observations,
not a stable benchmark. Delta excludes uninstrumented gaps between groups.

| Group | Prior ms | Current ms | Delta ms | Ratio |
| --- | ---: | ---: | ---: | ---: |
| ApiRequired | 2.321 | 4.029 | 1.708 | 1.736 |
| EmptyPrepared | 16.861 | 33.602 | 16.741 | 1.993 |
| NormalStages | 72.620 | 122.784 | 50.164 | 1.691 |
| AbsoluteDeadline | 37.157 | 55.648 | 18.491 | 1.498 |
| StopBoundary | 71.148 | 454.644 | 383.496 | 6.390 |
| TotalDeadline | 37.216 | 1056.706 | 1019.490 | 28.394 |
| StartupConsumesBudget | 19.109 | 383.394 | 364.285 | 20.064 |
| EqualTimeSequence | 25.480 | 347.793 | 322.313 | 13.650 |
| BackwardTime | 19.666 | 214.211 | 194.545 | 10.892 |
| SequenceRefusal | 16.911 | 214.456 | 197.545 | 12.681 |
| SessionClockRefusal | 38.065 | 442.752 | 404.687 | 11.631 |
| ActorDirection | 19.864 | 171.621 | 151.757 | 8.640 |
| ArmPrerequisites | 84.318 | 854.367 | 770.049 | 10.133 |
| ReleaseOnce | 74.079 | 835.918 | 761.839 | 11.284 |
| ProgressMonotonic | 201.804 | 2709.436 | 2507.632 | 13.426 |
| ExitBeforeEOF | 34.866 | 321.189 | 286.323 | 9.212 |
| UnknownExit | 27.248 | 274.194 | 246.946 | 10.063 |
| NonzeroExit | 30.465 | 314.128 | 283.663 | 10.311 |
| DataBounds | 105.943 | 1055.983 | 950.040 | 9.967 |
| ControlBounds | 106.240 | 1355.658 | 1249.418 | 12.760 |
| ReadFault | 49.536 | 260.921 | 211.385 | 5.267 |
| SupervisorLost | 146.518 | 188.907 | 42.389 | 1.289 |
| LauncherLost | 149.696 | 200.373 | 50.677 | 1.339 |
| FailureSticky | 188.630 | 254.322 | 65.692 | 1.348 |
| PhaseBudgets | 1164.793 | 1459.321 | 294.528 | 1.253 |
| CleanupUnknown | 385.331 | 586.644 | 201.313 | 1.522 |
| PersistRefused | 368.231 | 567.379 | 199.148 | 1.541 |
| TerminalTrailing | 143.846 | 243.687 | 99.841 | 1.694 |
| InvalidScenario | 223.506 | 366.096 | 142.590 | 1.638 |
| InvalidObservation | 203.377 | 339.144 | 135.767 | 1.668 |
| PassiveDescriptors | 155.674 | 184.288 | 28.614 | 1.184 |
