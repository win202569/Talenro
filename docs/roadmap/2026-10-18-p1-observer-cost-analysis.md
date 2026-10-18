# P1 ownership observer cost / coverage — read-only analysis

Date: 2026-10-18. Base checkpoint: 753960b7dceb9efaed72c719394f7cbdff1bf484.
Scope authorized by "continue as recommended": source and saved-result analysis
only. No candidate/test/controller edits or imports, no target host, helper call,
warmup, benchmark, retry, reviewer or frozen-slot execution.

Using systematic debugging: separate the observed timeout, statically provable
work growth, and unmeasured runtime attribution. This report does not fix or
accept the test or candidate.

## Evidence rechecked

582 frozen entries and saved stdout/stderr hashes still match the original
manifest/receipt. One consumed focused slot remains EXECUTION_INCOMPLETE.
Full and ownership capture files remain absent. Last saved line remains
BEGIN: root-mode-A, following root-schema-B PASS. Candidate SHA remains
1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8.
The [prior result](2026-10-18-p1-view-aware-ownership-result.md) retains raw
anchors, 23/39 completed cases, 37 facts and exact confidence limits.

The controller starts its wall timer immediately before process launch, waits
for the remaining 30000ms, and stops it after owned exit/stream drain.
Post-run pin checks and output-schema validation happen AFTER this timer.
Therefore the later fact parser/post-pin pass cannot explain the child's
30-second timeout. Launch, child execution, stream transport and the in-window
PID receipt save are within the wall window. No split timings exist for them.

BEGIN is emitted BEFORE fixture creation, two decisions, full-value checks,
leaf capture, selection and mutation. It cannot identify which of those stages
was running when the host was killed. There are no per-case elapsed times,
stage counters or CPU/profile measurements. Exact runtime attribution remains
UNDETERMINED; no claim that candidate calls themselves are slow or deadlocked.

## Static workload derivation (not a replay or benchmark)

Source: view-aware-ownership/owned-return.tests.ps1 at F06D199B.
Line 138 V-Others runs V-Invariant for every non-selected leaf. Lines 155/162
invoke V-Others for CURRENT leaves and ALL accumulated priorViews, both before
and after payload mutation. Line 170 appends all current non-selected contexts;
nothing clears priorViews. Each V-Invariant uses two R-Reference and two R-Type
observer helper calls in addition to field access/comparison. R-Type uses the
public reflection trampoline; these are observer-helper counts, NOT measured
reflection timings or a count of all CLR calls.

Let L(j) be current non-null scalar leaf count and P(j) retained prior count.
Starting P(1)=0 over the 14 non-null helper exercises followed by 39 cases:

- P(j+1) = P(j) + L(j) - 1.
- V-Exercise calls V-Invariant 2*(L(j)-1+P(j))+2 times on a success path.
- The extra 2 are selected-leaf checks. Two fixture sharing controls, value
  oracle comparisons, selection scans and leaf creation are excluded here.
- Short-circuit failure or interruption can reduce actual work; the complete
  totals below assume every preceding exercise succeeds.

Independent scalar helper exercises have L=3 and retain 2 each: 28 contexts
before the first behavioral case. Behavioral counts derived from literal shapes:

| Scenario kind | Input leaves | Leaves per returned result | L | Cases |
| --- | ---: | ---: | ---: | ---: |
| Normal | 53 | 33 | 119 | 29 |
| Sticky | 47 | 33 | 113 | 6 |
| Prepared | 5 | 28 | 61 | 2 |
| Refused | 53 | 12 | 77 | 2 |

Normal input: 5 root scalars + observations 11+7+13+9+8 = 53.
Normal result: 11 non-null root scalars + 22 Decision scalars = 33.
Sticky replaces the 13-leaf progress observation with a 7-leaf fault;
its non-null FailureObservationIndex adds one root leaf while null exit code
removes one Decision leaf, keeping 33. Prepared has no observations and five
null Decision values (three phase boundaries, sequence, exit code), so 28.
Refused has 12 non-null root scalars and no Decision container.

| Point in fixed success path | Prior contexts before exercise | Invariant calls in exercise |
| --- | ---: | ---: |
| First behavioral case | 28 | 294 |
| Last completed: root-schema-B | 2588 | 5414 |
| Interrupted next case: root-mode-A, if fully executed | 2706 | 5650 |
| Last planned case: refused-index-B, if reached | 4318 | 8790 |

The complete success path retains 4394 prior contexts, invokes V-Invariant
184178 times inside V-Exercise, and thus makes 736712 R-Reference/R-Type helper
invocations from those checks alone. The confirmed completed 37 exercises
(14 helper + 23 behavioral) imply 66522 such invariant calls. This is a static
count inferred from the saved success prefix, NOT instrumentation. It excludes
the two sharing controls and other observer/oracle work. Roughly 36% of modeled
invariant calls lie in the completed prefix, despite 23/39 behavioral cases.
Do not extrapolate elapsed time linearly from cases or these counts.

This establishes quadratic accumulated-history scan work as case count grows
with similar leaf counts. It does NOT establish its percentage of the 30 seconds.
Other source-visible costs include P1O-Equal's per-property Where-Object searches
over full schemas, repeated V-Selected scans, leaf/view allocation, candidate
evaluation and capture/host overhead. None has measured attribution here.

## Coverage constraints any later proposal must preserve

The accumulated scan is not merely redundant bookkeeping. It checks that a new
selected mutation does not appear on OLD unselected contexts' retained member
collections OR their freshly reread fields, with stable descriptors/storage/
base types and references. The current test deliberately excludes each prior
selected context from priorViews; do not claim it proves all historical owners.

| Proposed shortcut | Why it cannot silently replace current coverage |
| --- | --- |
| Clear priorViews or check only current objects | Drops cross-case unselected-owner evidence |
| Check only last/same-field/same-type historical owner | Narrows the all-retained-context contract |
| Cache prior invariants once | Does not check post-mutation stability each time |
| Compare bases instead of retained views/rereads | Misses the actual metadata visibility observation |
| Rewrap/reassign output String | Changes the subject after evaluation |
| Split cases across fresh hosts without an explicit scope change | Removes same-host accumulated history across partitions |
| Increase timeout or run remaining slots directly | Changes frozen execution rules; not a diagnosis |

Keep independent full 13+22 literals and exact base types, 15 scalar vectors,
39 cases, fresh bases/containers, both mutation phases, held and reread branches,
two fixture sharing controls and their limited meaning. No skips or relaxed
failure classification. No fresh product-negative-control proof has been added.

## Next decision, not authority to execute

Recommend a separate, bounded instrumentation DESIGN before any repair:
low-volume monotonic cumulative stage durations/counts for fixture/oracle work,
candidate decisions, view capture/selection, current-owner scans and prior-owner
scans. It should record stage entry as well as completion so an interruption is
locatable, account for instrumentation overhead, and distinguish completed counts
from attempted counts. This is an evidence request, not a promised speedup.

Do not remove assertions first and assume the growing-history hypothesis is the
cause. Any implementation, new fact/output schema, target-host run, observer
optimization or changed budget needs separate approval. Preserve the current
frozen leaf and raw records. No new allowance is implied by this report.
Existing NEW remaining 2/3 and OLD frozen slots stay frozen.

Analysis only, no I1/main/security/runtime/C12/release/PS5/performance acceptance.
Portable report/status/recovery documentation is preserved under the existing
standing checkpoint authorization; raw/session artifacts remain ignored.
