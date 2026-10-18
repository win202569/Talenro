# Candidate API internal paths — read-only phase analysis

Afterbab59d47,user approved read-only inspection of root setup,validation/copy
and decision paths. ZERO target/offline-suite launches,source/test/controller
edits,probe preparation,warmups,retries or slot releases. This is implementer
source analysis,not an independent review,profile,fix or acceptance.

## What is measured versus inferred

Latest completed PS7 split measured Evaluate4096=17567.757ms and full
comparison/assertion=22.203ms. It localizes that run's cost to the API-call
assignment region. No timestamps exist INSIDE this candidate evaluator in that
capture. Therefore root,validation,copy and decision shares are UNKNOWN.

The candidate module is3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3.
AST comparison with original953CC2A8 finds13 functions;only Integer differs.
Evaluator and12 noninteger function texts are identical to original.
That is source identity,not equal runtime latency across hosts/body histories.

## Source phase map for the successful clean4096 time fixture

Line references in integer-candidate-api-equality/offline-supervision-policy.ps1:

| Conceptual phase | Lines | Work and boundaries |
|---|---|---|
| Root/setup |115–131|Strict mode/result defaults,evaluation-owned metadata,root shape/scalars/labels,array type/decorations/4096 bound,bindings,owned list creation|
| Validation plus owned snapshot |132–163|Per-record descriptor/type,Kind,exact shape and passive values,integers/labels/actor,payload checks,then six-field snapshot/list append|
| Decision setup and progression |164–233|Decision/context construction,second loop168–232 for binding/actor/sequence/time/state checks,return result|

These are conceptual source ranges,NOT measured intervals or a ready probe
design. The catch path234–236 handles refusal/internal error;it is not the
ordinary successful fixture path. Scenario argument construction occurs in the
test caller before helper entry,not in these source ranges. Root includes
array adapter/member inspection:do NOT assume its engine cost is independent
of array size merely because the source contains no explicit per-element loop.

Validation and copying are interleaved per row,not two consecutive whole loops.
A timestamp before and after the first loop would measure BOTH. Separating
validation from copy would need per-row accounting or another intervention;
do not label a combined interval pure validation.

Fixture construction (test:184) makes4096 records,Sequence0..4095,AtMs0,Kind=time,
Actor=W,matching session/clock;caller expected result at:187 remains Prepared,
LastAtMs0,LastSequence4095 and conservative flags. Source tracing predicts:
- Both full loops process4096 rows if successful;maximum4096 array gate admits it.
- Each time row has six primitive fields,no kind-specific extra payload.
- Decision actor admission calls Get-SupervisionPolicyActors(time),then ordinal
  token membership. Binding/sequence/time pass;time case has no state transition.
  No arm/release/cleanup/persist action is requested by this synthetic fixture.
- This clean decision path does not call Shape/Passive/Integer validators.
  Its possible failure/window helpers listed by AST belong to other branches,
  not evidence they ran here. Successful source prediction is not profiling.

## Conditional call-volume model and limits

Validation only,all clean rows admitted:16 SDK type queries and10 SDK extended
Members reads per row (eight Passive calls plus two inline integer inspections).
For4096 rows:65536 type queries,40960 member reads,24576 copied field values,
4096 owned-list appends and4096 later decision rows. Metadata discovery occurs
once per evaluation;root/time allowlists are reused. Do not re-propose caching
metadata that is already reused or mistake fewer type queries for fewer guards.

These are arithmetic over inspected source,not observed dynamic counters.
Property/member enumeration lengths,adapter/runtime work,JIT/GC/dispatch costs
and per-call CPU/time are not known. The source has no scan of all preceding
observations inside each admitted-row validation;this is not a demonstrated
quadratic record search. Nor does it prove a runtime complexity/latency bound.

Keep all exact types/ordinal fields,descriptor/extended-member rejection,
integer bounds,actor/binding/state checks and owned results. Repeated validation
across helper/field boundaries is not permission to cache input acceptance.

## Comparison with retained original PS5 diagnostics

[Original region diagnostic](2026-10-18-p1-ps5-validation-regions.md) completed:
RootSetup10.384ms,ValidationCopy8572.100ms,Decision181.201ms.
Its measured disjoint per-row parts at4096 were validation8381.018ms and
copy156.168ms. The whole Shape region included nested safety work,not just
membership tests. [Earlier progress copy](2026-10-18-p1-ps5-validation-copy-progress.md)
timed out with1536 completed-row snapshot validation16490.836ms/copy258.662ms.
It demonstrated progress,not a permanent deadlock.

Those runs used original module semantics,PS5,modified bodies and a dedicated
boundary loading path. Latest candidate uses PS7 and a different integer helper.
Do NOT transfer old percentages/durations to the current candidate,subtract
them from17567.757ms or infer reflection is the unique cause. Their variability
is counterevidence to assuming instrumented timing is stable.

## Selected next hypothesis, not implementation

H-VC:in a prospective complete candidate phase diagnostic,the first whole-loop
ValidationCopy interval is larger than BOTH RootSetup and Decision intervals.
This is a localization hypothesis,not proof of one helper's cost or a fix.

A complete run with root or decision larger would falsify that ordering. A
timeout with ValidationCopy open after root closes would support locating an
unfinished diagnostic region only;no finished duration,exact completed-row
count or permanent hang could be inferred. Earlier/open later frontiers retain
their actual names. No new observations of H-VC occurred in this turn.

Next recommend a bounded DESIGN for sparse whole-phase observation without
per-row timers,guard changes or speculative optimization. Design must explicitly
choose marker placement,scope/loading policy and output/stack checks. Internal
markers necessarily alter the evaluator body/history;do not claim zero
interference or quietly introduce the old local-reload confound. It is not
possible to label a newly instrumented body as unchanged-source acceptance.
Preparation and any future launch remain separate stages needing user agreement.

## Verification and preserved state

Parsed source ASTs only,extracted both For-loop ranges/named commands,compared
function texts and independently evaluated the conditional arithmetic.367 prior
pins and latest invocation index5090D283DE61BBAE35ABF07521BCEBD3AB512007F4A704ED2CD0C4E0693DEE0D
held. No target script was imported/invoked;no new raw/controller/source files.

Split1/1 and boundary1/1 spent,candidate-full1/2 spent/PS5 FROZEN;all older
spent/frozen allowances and review0/1 unused/ineligible persist. Original
timeout,candidate/I1/P1/main/C12/security/runtime/release gates remain unresolved.
Save this analysis/status/recovery on the development branch. Git restores
neither ignored raw,cached runtimes,credentials nor execution authority.
