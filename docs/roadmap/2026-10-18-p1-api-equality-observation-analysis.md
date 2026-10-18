# API versus equality observation — read-only design analysis, NOT PREPARED

Afterd5d21b17,user approved designing how to separate API evaluation from full
result comparison. This turn inspected source and existing records only.
ZERO target/offline-suite launches,new probe/controller/verifier implementations,
product/test edits,retries,warmups or slot releases. This report preserves
analysis and a PROPOSED bounded design;preparation still needs user confirmation.

## Source evidence and design question

integer-candidate-boundary-markers/offline-supervision-policy.tests.ps1:54:
P1-Check accepts Scenario and Expected,calls Get-OfflineSupervisionDecision into
local r,then P1-Need(P1-Equal r Expected),and returns r. At:178 caller arguments
construct the scenario/expected values before the helper body executes.
Check4096 outer markers include those arguments,helper dispatch,API,result
comparison and return. Existing16860.926ms therefore cannot be labeled API-only.

P1-Equal(:43–53) recursively checks literal full expected result fields with
a depth bound. For this fixture the expected result is the ordinary fixed-size
result/Decision object,not the4096 input array. This source fact does not prove
equality latency or exclude engine/property-access overhead. It gives no license
to replace full equality with a small subset of fields.

Original ps5-boundary-subphase-timing already used a dedicated helper at:68
without loading product inside it;its PS5 evaluation interval remained open.
Later diagnostic copies added local product reload. Those historical
original-module PS5 observations do NOT measure the current integer candidate
PS7 split;do not reuse old code blindly or promote them to current attribution.

## Options considered

1. Instrument common P1-Check,conditionally selecting this call:changes common
   helper body/earlier dispatch or adds branches to all previous invocations.
   That perturbs the preceding execution history;not recommended.
2. Inline API/equality into the case body:removes original helper function
   scope/parameter binding and changes where local r exists;not recommended.
3. Dedicated single-call helper with identical parameter/local/result operations,
   plus four phase markers:no product reload,no changes to existing helpers or
   earlier calls. RECOMMENDED,with explicit fresh-helper/history caveat.

No zero-interference observation is claimed. The question is whether a future
diagnostic spends more time in its API-call region or comparison/assertion region,
not whether this design reproduces unchanged-entry latency.

## Proposed bounded preparation design

Create a NEW labeled isolated folder derived from the existing outer-marker
candidate folder,not edits to any measured source. Product and five other
entry/wrapper/test copies remain byte-identical. In its original35 copy only:

- Add one separately named function P1-SplitCheck4096 at test-script level,
  alongside P1-Check. Keep the same Scenario/Expected parameter shape,API
  command and arguments,local r assignment,full P1-Equal/P1-Need expression,
  return value and ordinary function call style. No extra scriptblock wrapper.
- Insert SUBPHASE Evaluate4096 START/END around the original API assignment,
  and SUBPHASE Assert4096 START/END around the original full comparison/assertion.
  Use direct Console.Out monotonic timestamps as existing outer markers do;
  no marker helper or product/internal-loop instrumentation.
- Change ONLY the4096 boundary caller's helper command name to the new function.
  Keep all argument expressions,outer markers and statement order intact.
- Keep original P1-Check/P1-Case/P1-Equal and every other original function text
  byte-identical.4097 and other calls still use original P1-Check.
- No product dot-source/reload inside the new helper,no accepted-input cache,
  changed exact types,descriptor checks,bounds,actor checks or assertions.

Unavoidable intervention:the helper has a different name/FunctionInfo/ScriptBlock,
new body size and fresh invocation history rather than P1-Check's earlier calls.
Dynamic command resolution and compilation behavior may differ despite equivalent
source-level function scope. No runtime binding/provenance guarantee is claimed.
Product loading DECLARATIONS remain the same;that does not attest engine caches.

## Proposed output and offline validation obligations

Complete expected inventory:95 ordered nonempty lines,64 PASS,12 timing markers,
six complete pairs (four outer,two inner).Four additional records appear between
outer Check4096 START and END in strict order Evaluate START/END,Assert START/END.
Outer Check4097 remains later. No new PASS/test-case counts.

The existing91-line verifier intentionally rejects this different inventory
and nested timing scopes. Do not weaken it or silently repurpose its hash.
Preparation would include a DISTINCT pinned95-line inventory and a separate
split-capture classifier with a strict stack:Check4096 contains Evaluate and
Assert siblings;parents may remain open on timeout. Preserve the old verifier
and its47-case suite unchanged. New offline tests must precede implementation,
include positive nested capture and invalid context/host/marker/order/tick/
containment/prefix/end/receipt cases,and never invoke the product.

Static fidelity gates:AST parse only,exact six-file equality,unchanged original
function definitions and one extra helper;remove extra helper/four inner markers
and restore the single caller name to recover the measured outer-marker test
exactly. Compare the new helper stripped of phase lines/name with P1-Check,
preserving parameters,operations and order. No controller or target launch in
preparation. Verify complete95-line inventory and12-marker nesting independently.

Only closed pairs get durations. Require nonnegative invariant Int64 ticks,
global monotonicity,matching name/kind/edge,fixed positions,proper containment,
and a positive pinned capture frequency. Do not add parent and child durations.
Outer-minus-children is unassigned residual (arguments,dispatch,return,logging,
gaps,etc.),not proof of one component's cost or calibrated instrumentation cost.
Do not subtract runtime-deadline from an unfinished phase.

## Prospective interpretation; no measured split result yet

H-API:for the future completed split capture,Evaluate4096 exceeds Assert4096.
If evaluation starts but does not finish before timeout,that localizes the
unfinished diagnostic region to the API-call assignment. It does not locate
validation/copy/decision/SDK/JIT/GC internals or prove a permanent hang.
If assertion is larger or open instead,H-API is unsupported for that run.
If neither starts,retain the actual earlier frontier. A complete fast run does
not reproduce the historical timeout;no stable speedup claim or guard removal.

The Evaluate interval includes command dispatch/parameter binding/API execution/
return assignment,not pure evaluator CPU. The Assert interval includes full
P1-Equal and P1-Need plus dispatch,not equality alone. Captured all-thread process
CPU cannot be divided into these wall-time regions.

## Scope, verification and next decision

Read-only recheck held all344 prior pins and old invocation-index hash.
No script/product/controller implementation was created. Prior original and
candidate hashes,the frozen host-regex failure,and separate offline assessment
remain unchanged. No diagnosis of the original timeout has been completed.

If user accepts this design,the NEXT step is preparation ONLY:new isolated
source/output copies,strict offline split-classifier tests and static fidelity
checks. No autorunning controller,product execution or replay permission.
Any later launch needs a separately approved NEW ordinary PS7 single-host/
30000ms/no warmup/no retry scope,with frozen identities and owned-process limits.
PS5 stays FROZEN;current boundary1/1 spent,candidate-full1/2 spent,and all older
allowances/review0/1 unused/ineligible persist. I1/P1/main/C12/security/runtime/
release gates are not cleared. Preserve this proposal/status/recovery on the
development branch;Git restores neither ignored raw/caches nor replay authority.
