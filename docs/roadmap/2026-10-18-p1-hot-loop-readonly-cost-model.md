# Read-only hot-loop call-volume model — 2026-10-18

After8833f30f, user approved the recommended read-only cost analysis. ZERO new
test/diagnostic hosts, product/test/controller edits, warmups or retries. This is
a conditional source-derived model, NOT a profiler, independent review or fix.

## Model boundary and derivation

Source: docs/diagnostics/isolated-fuzz/offline-supervision-policy.ps1,
SHA953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.
Model the validation/copy loop at:121–151 for the original boundary fixture:
4096 ordinary time records, each six primitive fields (Sequence,SessionLabel,
ClockLabel,AtMs,Actor,Kind), no payload fields, all checks succeeding. Model
excludes root setup, decision loop, fixture creation, result equality and other
tests. Counts assume the FULL validation path completes; the timed-out run did
NOT record its completed row count, so these are not its observed call totals.

Get-SupervisionPolicyType(:12) calls the cached SDK MethodInfo.Invoke for every
nonnull value. Assert-SupervisionPolicyPassive(:17) calls that helper, then for
primitive scalars obtains the SDK Members view with PropertyInfo.GetValue and
checks extension members before enumerating public properties. Integer(:51)
and Token(:59) validators each query type themselves and then call Passive,
which queries the same local value again. Shape(:32) queries record type and
calls Passive on every admitted NoteProperty value.

| Per clean time row, validation only | Type helper / SDK Invoke | Passive calls | SDK Members GetValue |
|---|---:|---:|---:|
| Initial observation type | 1 | 0 | 0 |
| Kind token validation | 2 | 1 | 1 |
| Exact shape plus six field values | 7 | 6 | 6 |
| Sequence and AtMs integer checks | 4 | 2 | 2 |
| SessionLabel and ClockLabel type checks | 2 | 0 | 0 |
| Actor token validation | 2 | 1 | 1 |
| Total per row | 18 | 10 | 10 |
| Conditional total for4096 rows | 73728 | 40960 | 40960 |

Exact allowlist Contains executes six times per admitted row,24576 for4096;
that is distinct from all the nested scalar reflection/member work in Shape.
Each row also enumerates record properties for the initial descriptor check,
the Shape property snapshot and the copy, plus iterates the six-element snapshot.
Each of ten scalar Passive calls enumerates extended members and public
properties. Those collections' element counts are NOT derived here: engine
views/type metadata/decoration matter. Getter count is not enumeration length.
There are six copied fields and one list append per admitted row.

The second decision loop performs binding/actor/sequence/time/state checks on
the copied records; this clean time path has no calls to the modeled reflection
validators there. That does not make decision work free or include its cost in
the table. Source shows no per-row scan of all previously admitted observations
in this validation path; repeated fixed-field work, not a demonstrated quadratic
record search. No runtime asymptotic/latency guarantee is claimed.

## Already reused versus repeated work

SDK metadata discovery is ALREADY reused: New-SupervisionPolicyMetadata creates
one context per evaluation, including fixed GetType MethodInfo and Members
PropertyInfo. All modeled calls receive that context; the default-null fallback
does not create metadata on every row. Fixed allowlists are also reused: root
and time yield two shape-set entries for this fixture. Do not repeat the earlier
metadata/HashSet optimization or mislabel every Invoke as metadata discovery.

There are four typed-validator calls per time row (Kind,Sequence,AtMs,Actor).
Each performs a type query followed by a second query in Passive on its local
argument. That identifies16384 repeated type-query pairs' second queries over
the full fixture. It is an investigation target, NOT a measured saving or proof
that removing them is safe. Count reductions cannot be converted to wall-time
percentages: invocation, reflection, member traversal and compilation costs differ.

Shape and later typed checks also revisit four fields' passive status. Unlike
the within-helper local argument, these involve separate field reads and contract
boundaries; treating them as reusable input acceptance needs stronger mutation/
ownership reasoning. No such cache or shortcut is proposed here.

## Security constraints and evidence limits

Keep exact types, exact ordinal fields, NoteProperty-only access before values,
all extended-member rejection, executable-descriptor rejection, integer bounds,
ordinal tokens, actor/binding checks, per-call validation and result ownership.
Do not replace fixed SDK type access with a user-shadowable GetType method or
Members access, broadly relax to coercion/type-family checks, cache acceptance
across calls, or skip decoration checks because the base value is primitive.

Existing source tests cover hostile GetType/ToString/ScriptProperty callbacks,
Members shadowing and decoration-after-clean, root/observation scalars,
sequential isolation and caller mutation. These were inspected as contract
evidence, NOT rerun. A future design must preserve their refusal behavior,
failure index/result shape and zero-callback expectations; previously passing
tests do not prove a future candidate safe.

Earlier disjoint diagnostic put51.0133% of its measured validation wall in the
whole Shape helper, which includes nested work; it did NOT time SDK Invoke or
Members GetValue individually. The latest unchanged-product local-reload run
only shows Evaluate4096 open at timeout. Neither establishes reflection as the
unique cause, a stable speedup target or a justified security tradeoff.

## Verification and next step

Parsed source ASTs without invoking them; checked call-chain sites and manually
derived this clean path, then independently evaluated arithmetic18*4096,
10*4096,6*4096 and4*4096. This arithmetic is not an executed dynamic call counter.
All283 latest frozen pins held; latest loading-context invocation count stays1
(Local timed out),Inherited slot FROZEN. No original source/test changes.

Next recommend a bounded design/safety review of whether the INTEGER validator
can reuse its own exact type result during the same local-value inspection,
while retaining the full Passive checks and safe direct-call behavior. Separate
semantic proof obligations from any performance hypothesis before implementing
or requesting fresh verification. No production fix, diagnostic launch, unfreeze
or new allowance is authorized by this report.

All old spent/frozen budgets,review0/1 unused/ineligible and PS5/I1/P1/candidate
acceptance blockers persist. No main/C12/security/runtime/release acceptance.
Recover this report, preceding reports/source,current status and runbook from
codex/c12-b01-task9-coordinator,not ignored raw/controllers/cached engines.
