# Token/Passive local type reuse — READ-ONLY contract analysis

Aftere512bb80,user accepted read-only analysis of duplicate type queries.
ZERO source/test/candidate/controller/probe changes,target/offline-suite runs,
warmups,retries,review dispatches or frozen-slot releases. Only this analysis/
status/recovery are written under standing progress-preservation authorization.

## Finding and limit

There is a narrow SOURCE-LEVEL duplicate:Assert-SupervisionPolicyToken obtains
exact type,then calls Passive,which obtains type again. This is a candidate
attribution/contract question,NOT proven safe replacement or proven timeout
root cause. Latest partial instrumentation measured combined validation,not
this type helper. Prior H-Attribution remains not-assessed;no rule changed.

Inspection target:docs/diagnostics/isolated-fuzz/integer-candidate-validation-copy-attribution/offline-supervision-policy.ps1.
AST confirms Token(:70–74),Passive(:19–33),Bool(:75–79) texts exactly identical
to original product module. Integer(:53–66) already uses an inline local-type
pattern;that is a source precedent,not an acceptance or engine-equivalence proof.
Parameter lists:Token($Value,[string[]]$Allowed,$Metadata=$null);
Passive($Value,$Metadata=$null);Bool($Value,$Metadata=$null).
Value remains UNTYPED at each function boundary. Metadata is created internally
per evaluation and forwarded;analysis does not claim equivalence for arbitrarily
forged external Metadata objects.

## Current obligations and order

| Stage | Existing source behavior that must remain |
|---|---|
| Metadata |Token fills null metadata before type lookup;evaluation-owned context,no cross-call input acceptance cache|
| Exact type |Fixed SDK MethodInfo/GetType(BaseObject),exact string or refuse before Passive;no value coercion|
| Scalar decoration |Passive selects primitive branch,reads psextended via fixed SDK MembersProperty getter,rejects any extended member|
| Descriptors |Inspect properties;reject ScriptProperty/CodeProperty and non-Property Length/Rank shadows using ordinal-ignore-case comparisons|
| Token admission |Only AFTER passive checks,ordinal membership in the original Allowed array;wrong token refuses|
| Result / failure |No new success-pipeline output;keep refusal reason/index/order and all caller/snapshot/result-isolation behavior|

For null/non-string Token refuses before Passive. A correctly typed but decorated
string is not valid merely because its BaseObject is a string;member/descriptor
checks must still run before token admission. Invalid token with a decoration
must not skip those passive checks or trigger its executable descriptor.
Changing Value parameter to[string]/[bool],casting input,using Value.GetType(),
or ordinary input-shadowable Members access is not type-query reuse:it changes
coercion or bypass protections and is outside scope.

The narrowly interesting future question is whether a Token-local exact type
result can replace only the second query while every passive check remains
inside that invocation. Do NOT delete Passive's obligations,change shared
Passive,add caller-trusted KnownType/AlreadyChecked parameters,cache acceptance,
merge checks across distinct field reads/calls,or alter Shape/Integer/Bool.
Removing the separate Passive call would also remove a parameter-binding/frame
boundary;source equality cannot prove PSObject wrapper/extended-view identity
survives that change. This is an unresolved proof obligation,not an observed
binding defect. No target/helper execution was performed here to resolve it.

## Source-only effect model,not speedup

For the inspected clean4096 time fixture,Kind andActor each use Token once;
payload Bool calls are zero. Current model16 type queries/31 helper entries/
10 SDK Members reads per row. IF Token-local fusion were contract-equivalent:
14 type queries/27 helper entries/10 Members reads per row;8192 fewer queries
and16384 fewer helper entries across4096. Only query count falls12.5%,NOT wall
time/CPU/validation latency. Full member/descriptor work remains;body expansion/
compilation/history can offset a count reduction. Snapshot/list copies stay.
Do not widen the candidate to Bool merely because its source has a similar
duplicate:that adds semantic surface without a predicted benefit on this fixture.

## Existing fixture coverage and remaining proof needs

ownership.tests.ps1:44–65 constructs owned primitive wrappers,attaches mutable
Metadata and asserts owner-field type/value/descriptor/reference preservation.
It does not observe every internal Token/Passive parameter-binding boundary.
:195–208 decorates strings and all observation fields;this is full-API coverage,
not automatically direct helper coverage.

Phase clone checks Kind via Token BEFORE Shape;Actor via Token AFTER Shape.
A decorated Actor can therefore refuse in Shape.Passive before reaching the
Actor Token call. Full decorated-field refusal cannot alone prove a proposed
Token-local passive tail catches it. Classification/root fields can likewise
be refused earlier by Shape. Existing integer-only executable descriptor cases
also cannot substitute for string-token boundary coverage.

Before any future candidate execution,design must specify direct Token boundary
cases using owned fixtures and the actual internal metadata contract:
- valid exact strings,empty/unknown/wrong-case tokens,null/wrong primitive types,
  arrays/plain objects;preserve admission/refusal without coercion;
- decorated strings with mutable NoteProperty payload and actual identity/type/
  descriptor preservation at the helper input boundary;
- ScriptProperty/CodeProperty/AliasProperty/ScriptMethod decorations and
  Length/Rank/Members/GetType shadow cases with callback counters staying zero;
- null metadata fallback and internally created metadata forwarding;
- invalid-token plus decoration to check rejection order,not token-only refusal;
- clean→decorated→clean and input/metadata/result mutation isolation;retain full
  original API equality/refusal-index/owned-result assertions as separate gates.

These are REQUIRED COVERAGE QUESTIONS,not new test scripts,an approved candidate
or evidence any new helper/direct fixture ran. A fixture that loses its
decoration must be a fixture error,not a passing refusal test.
An instrumented or local-function candidate is not unchanged-source acceptance;
any later preparation/execution needs explicit scope and unchanged guard gates.

## Read-only evidence / next step

AST parsed original/candidate helper definitions,compared three function texts,
extracted parameter/named-call edges,read fixture/caller/phase order,and summed
the conditional model independently. No import/invocation of target/helper/test.
434 frozen pins plus latest manifest/controller/index/raw streams held;still
one prior attribution invocation. Latest index:
EE36E5EB77200E822035F19E50BC83D63AC78406E3EEE98083B500A71DBC602D.
Source module3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3
and original953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA
unchanged. This is author contract analysis,not independent safety review.

Recommend next a BOUNDED DESIGN ONLY for Token-local reuse plus explicit direct
boundary fixtures,excluding Bool/global caches/new timing instrumentation.
No source preparation/optimization/benchmark is authorized by this analysis.
Remaining engine binding/decoration/refusal-equivalence and performance are
unverified;no claim fewer queries fixes the timeout.

All prior budgets persist:attribution/ownership-marker/three-phase/split/boundary
each1/1spent,candidate-full1/2spent/PS5FROZEN;review0/1unused/ineligible.
Candidate/I1/P1/main/C12/security/runtime/release UNACCEPTED. Git restores this
analysis/status/recovery,not ignored raw/controller/cache/credentials/replay.
