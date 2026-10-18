# Integer-local type-reuse candidate prepared — NOT RUN or promoted

Afterd41feabd, user approved preparing the reviewed minimal candidate only,
without replacing the original product or running new tests. Added a separate
`docs/diagnostics/isolated-fuzz/integer-type-reuse-candidate/` folder containing
candidate product copy,byte-identical direct suite,and UNVERIFIED warning README.
No target host,product dot-source,test body or fixture was executed this turn.
This is an unfinished candidate checkpoint,not semantic/performance acceptance.

## The one bounded change

Only Assert-SupervisionPolicyInteger differs semantically from the original:
after its existing exact int/long check,it inspects the SDK extended Members
view and public descriptors directly,then checks range. It no longer calls
Passive,which would query the same local value's type a second time. Since its
own type gate already excludes null/bool/string/noninteger values,the scalar
inspection branch is specialized for admitted int/long values only.

Retained unchanged:

- Four parameters/defaults and metadata fallback;untyped Value,exact type gate.
- Fixed SDK MembersProperty.GetValue and rejection of EVERY extended member.
- Public-property traversal,ScriptProperty/CodeProperty rejection and ordinal-
  ignore-case Length/Rank non-Property rejection.
- Order:type first,extended members/descriptors next,range last;existing error
  strings and no added success output,exception wrapper or logging.
- All12 other function bodies,including common Passive,Token,Bool,Shape,
  metadata/type helpers and evaluator;no known-type flag,acceptance cache or
  global state added.

The small specialized block is intentionally duplicated and commented to remain
contract-equivalent to Passive. Future changes to those safety rules require
both paths to remain covered. The original product and deliberate guard-loss
negative-control fixture were NOT changed. Never confuse the defective control
with this safety-preserving INTENT;the candidate is not yet behaviorally verified.

## Static verification actually performed

PowerShell7.6.5 AST parser reported no syntax errors for candidate and test.
Compared all13 function definitions:12 body texts identical;integer signature,
first three statements (metadata/type/gate),three scalar-inspection statements,
descriptor loop and final bounds statement match their original counterparts.
The candidate integer has one direct type-helper source call and no delegated
Passive call. Replacing its one helper definition with the original and removing
the two warning comments restored the entire original product text after newline/
end normalization. This checks source correspondence,NOT runtime equivalence or
dynamic type-query counts.

The direct suite remains exactly
E442D1DDF20C901A511097D210FA66A23AAB48F838CEB262382599BFEDC21C98.
Its relative module load would select the adjacent candidate during a separately
authorized run;the suite itself is not edited or weakened. The earlier baseline
passed14 groups on both engines;the isolated guard-loss control failed at the
exact intended assertion on both. Those are prior test-sensitivity evidence,
not a fresh candidate PASS or a completed candidate red/green cycle.

All301 previous frozen pins held. Invocation counts remain baseline2,negative
control2,loading1 (Inherited FROZEN). No controller/allowance was created for the
candidate and no old slot was reused. CodeProperty runtime fixture/coverage gap
persists. Whole-project/full original regression and independent review were
not performed. No performance or timeout-fix claim.

## Candidate identity and limits

| Artifact | SHA256 |
|---|---|
| Original product,UNCHANGED | 953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA |
| Isolated candidate product | 3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3 |
| Copied direct tests | E442D1DDF20C901A511097D210FA66A23AAB48F838CEB262382599BFEDC21C98 |
| Candidate README | C16AEF8C0F085F938F40E6E167E2F5F3AD368FB482D60678194280F89E9F44D9 |

The prior clean4096-time source model predicts two fewer type-helper invocations
per validation row if the new integer path behaves as intended:18 to16,8192 fewer
over a COMPLETE4096 path. Corresponding member inspections remain,including the
same modeled40960 SDK Members reads. These conditional counts are not actual
cutoff totals,measured savings,stable latency or unique-cause attribution.

The reuse reasoning is limited to a synchronous invocation on the local value
with normal owned SDK metadata. It adds no concurrency/live/hostile-host metadata
support. Direct-helper semantics and source-level copying do not establish
original complete-entry acceptance. Original PS5 timeout/I1/P1 remain blocked;
review0/1 unused/ineligible and all spent/frozen budgets/main/C12/security/runtime/
release gates remain unchanged.

## Next boundary and recovery

Next recommend a separately approved NEW semantic-validation allowance:ordinary
fresh PS7 then PS5,adjacent byte-identical14-group suite once each,30000ms per
host,exact frozen17-line output/exit0/empty stderr/source and engine checks,
stop/freeze later slot on any failure or timeout,no retry/warmup/product edits/
prior-budget reuse. It would test the candidate copy only,not promote it or
constitute full original regression. Do not launch without approval.

Recover candidate/test/README,this report,current status and recovery runbook
from codex/c12-b01-task9-coordinator. Original product stays in place. Git restores
neither ignored controller/raw packets,runtime caches nor execution authority.
