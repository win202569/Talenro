# ValidationCopy — read-only field/call-path cost model

After0874e244,user accepted read-only analysis of the measured ValidationCopy
region. ZERO target/offline-suite launches,probe/controller/source/test edits,
warmups,retries,review dispatches or frozen-slot releases. Only source/AST and
saved records read;this report/status/recovery are the only new tracked writes.

## Evidence boundary

Latest complete ownership-marker PS7 diagnostic:
ValidationCopy17515.183ms,RootSetup11.598ms,Decision572.380ms,
Evaluate4096(parent)18129.450ms. ValidationCopy is about96.6% of the measured API
parent interval,NOT96.6% of whole wall time or CPU. Whole wall28983ms left1017ms
under the cap. These are observations,not a stable performance guarantee.

The marker clone first loop is phase-evaluator.ps1:24–55,between START:23 and
END:56. It interleaves validation then owned copying per row;copy is not a
separate measured whole-loop phase. The unmarked candidate loop:132–163 has
exactly identical AST extent text;both second loops also identical. No source
was dot-sourced or invoked to establish that identity.

Fixture source:offline-supervision-policy.tests.ps1:185 constructs4096 calls to
P1-Observation(index,0,time),whose default Actor isW (:14–16). Each observation
has exactly Sequence,SessionLabel,ClockLabel,AtMs,Actor,Kind and no payload.
P1-SplitCheck4096:55–62 calls the phase clone before the separate full assertion.
All arithmetic below is CONDITIONAL on the clean time-row success path,
ordinary helper resolution and admission of all4096 rows. It is not a dynamic
counter,profiler sample or count of lower-level engine operations.

Paths below use docs/diagnostics/isolated-fuzz/integer-candidate-ownership-markers/.

## Per-row helper paths

Clone:26 performs exact observation type lookup;:27 rejects non-NoteProperty
descriptors;:28 requires Kind;:29 validates its exact string/passivity/token.
Shape:37 calls Assert-SupervisionPolicyShape(module:34–52),which checks object
type,materializes/counts properties,checks ordinal allowlist and NoteProperty,
then calls Passive on every field. Integer checks:38 retain exact type,extended
member/descriptor/bounds checks. Labels:39 add exact-string and length checks;
Actor:40 adds exact-string/passivity/ordinal admission.
For Kind=time,the payload switch:41–53 has no matching payload branch:zero Bool
helper calls on THIS path,not a claim about other fixtures or input kinds.
Copy:54 builds a fresh ordered map,reads each property value,casts an owned
PSCustomObject and appends to the evaluation-owned list.

Module type query:14–18 uses fixed SDK MethodInfo on BaseObject.
Passive:19–33 checks scalar type,uses fixed SDK MembersProperty getter,then
rejects decorated scalars/executable descriptors/shadow bounds.
Integer:53–66 already reuses its own exact integer type locally instead of
calling Passive again,while retaining the corresponding safety checks.
Token:70–74 performs a direct exact-string query,then Passive (which queries
type again),then ordinal membership via TestToken:67–69.
Existing metadata is evaluation-owned;root/time shape allowlists already reuse
fixed schema sets. Do not re-propose global/input-trust caching or claim
metadata discovery occurs anew for every row.

## Conditional per-field type/member ledger

Members reads mean the fixed SDK extended-members getter,not ordinary property
lookups or individual member enumeration iterations.

| Value | Type queries per row | Members reads per row | Where |
|---|---:|---:|---|
| Sequence |2|2|Shape Passive + Integer inline check|
| AtMs |2|2|Shape Passive + Integer inline check|
| SessionLabel |2|1|Shape Passive + label exact-string check|
| ClockLabel |2|1|Shape Passive + label exact-string check|
| Kind |3|2|Token direct type + Token Passive + Shape Passive|
| Actor |3|2|Shape Passive + Token direct type + Token Passive|
| Observation object |2|0|Initial object check + Shape object check|
| Total |16|10|65536 queries/40960 Members reads over4096 rows|

There are eight Passive helper entries per time row (six Shape fields plus
Kind/Actor),two Integer,one Shape,two AssertToken,two TestToken,and16 Type helper
entries:31 helper entries per row,126976 over4096. This excludes evaluator entry,
root setup,fixtures,second decision loop and engine-internal method calls.

Three SOURCE-LEVEL observation Properties collection traversal/materialization
sites exist per row:initial descriptor scan,Shape materialization,and copy.
Shape separately iterates its already-materialized six-descriptor array.
The initial/Shape validation loops check six descriptors EACH;copy assigns six
values:12 checks +6 assignments per row. Do not mistake18 such visits for all
engine/property materialization work. The ten scalar Properties enumeration
sites per row come from eight Passive plus two inline Integer inspections;
their iterator lengths/adapter work are unknown,not all one item.

Further conditional source arithmetic:
- Six Shape allowlist Contains checks per row:24576 total.
- Kind=time is fifth in its ordinal list;Actor=W second:seven successful-path
  TestToken string comparisons per row,28672 total.
- One new ordered map,one PSCustomObject snapshot and one list Add per row;
  24576 copied field values and4096 list appends.
- The time shape allowlist is built once in this evaluation's first time row,
  not4096 times. Root's set was built before this region. Shape retains its
  ContainsKey/name/type/value checks for every admitted object.

## Interpretation and next decision

Established locality:the measured loop covers a large repeat of type/passivity/
descriptor/shape/token operations PLUS snapshot allocation/copy and dispatch.
Repeated paths identify attribution candidates,not redundant permission checks
that may be deleted. Exact type and passive/descriptor refusal are different
contracts;field count/name checks and snapshot ownership also serve separate
purposes. Exception/refusal index/order and mutated input/results must stay.

Not established:individual helper latency,copy share,allocation/GC share,
parameter binding/adapter/reflection shares,exact dynamic invocation counts or
a unique root cause of the earlier timeout. No obvious source-level scan over
all prior observations appears inside the first loop;that is not an engine
complexity/latency proof. Repeating Type/Token work does not prove it dominates,
nor that reducing a query count produces the same percentage speedup.

Next recommend DESIGN ONLY for a minimal attribution experiment distinguishing
validation helper work from owned snapshot/copy,with explicit instrumentation/
body/history cost and falsifiable outcomes. Do not silently move validation or
copy into a second traversal,reorder guards,cache input acceptance,drop copies,
raise the cap,add per-row output or run a new target. No design/probe prepared
in this turn;preparation/execution need separate scoped agreement.
No speculative optimization or claim that the prior timeout is fixed.

## Read-only verification / recovery

AST-parsed candidate and marker clone,extracted named helper edges and loop
ranges,checked both loop texts exact,read fixture/caller definitions,and
independently summed the field/call model. No utility imports product sources.
412 frozen pins and latest single-invocation index held before and after work.
Latest index:
7F8D8368B2F2263A85FF352ECB8756952DD4F151A13D5F344EF66993ECB9A292.
Candidate3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3;
clone6610F6F4396B9C0B1DF6AA202A63572F45EA037C175F30DF85973146F1881AE6.
The [prior result](2026-10-18-p1-ownership-markers-result.md) preserves full raw/
controller/manifest/stream identities. No receipts or saved outcomes rewritten.

Ownership-marker/three-phase/split/boundary each1/1spent,candidate-full1/2spent
with PS5FROZEN;all old budgets unchanged. Review0/1unused/ineligible;this is
author source analysis,not independent review. Candidate/I1/P1/main/C12/security/
runtime/release UNACCEPTED. Git restores this analysis and portable preparation,
not ignored raw/controllers,cached runtimes,credentials or execution authority.
