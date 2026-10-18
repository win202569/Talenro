# Integer type-result reuse — read-only safety audit

Aftere185bbc1, user approved the recommended bounded design/safety review only.
ZERO target hosts, no product/test/controller edits or implementation. This
records source findings and an UNAPPROVED candidate option, not a formal spec,
independent review, proven-safe patch, performance result or acceptance.

## Existing contract and coverage gap

Assert-SupervisionPolicyInteger in offline-supervision-policy.ps1:51 has four
parameters, unchanged proposed interface:
Value; Min(long,default0); Max(long,default long.MaxValue); Metadata(default null).
Order today: create owned default metadata if needed; query exact real type;
reject anything other than int/long; perform full Passive inspection; check bounds.
Passive repeats the type query on the helper's local argument before inspecting
the fixed SDK Members view and public property descriptors.

Original product hash remains
953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.
AST inspection of original35,ownership,shape and reflection test files found ZERO
direct Assert-SupervisionPolicyInteger calls. Their scenario-level decoration
tests exercise earlier Shape/Passive rejection before reaching integer checks
for those fields. They remain necessary end-to-end contracts but cannot alone
prove that a changed integer helper retains its own member-safety defenses.
This is a static coverage finding, not a claim that those tests failed.

## Options assessed

1. Preferred bounded candidate: specialize the integer helper's own passive
   inspection after its int/long type gate, retaining all of Passive's applicable
   SDK extended-member and descriptor checks in the same order. Reuse only that
   helper's local type result; keep standalone Passive,Token,Bool and Shape
   unchanged. No new parameters or trusted-caller mode. Tradeoff: a small copied
   safety block must remain contract-equivalent to Passive; future changes need
   tests covering both paths.
2. Not selected: add a caller-supplied KnownType/AlreadyValidated parameter to
   common Passive. It broadens the interface and creates a new proof obligation
   against bypassing type checks; validating the hint again loses the intended
   saving. No need to introduce that trust surface for this bounded candidate.
3. Rejected: remove integer passive checks because Shape already ran, or cache
   input acceptance. This weakens direct helper behavior and introduces mutation/
   cross-call assumptions instead of eliminating only a local repeated query.

The preferred option is suitable for a proposed minimal implementation, subject
to approval and direct regressions. It is NOT yet established behaviorally safe.

## Invariants the candidate must preserve

| Boundary | Required behavior |
|---|---|
| Input typing | Value stays untyped; exact int/long only, no coercion to long or acceptance of bool,string,float,unsigned/enum values |
| Null/wrong type | Same Exact integer required failure before passive/bounds work |
| Extended members | Fixed SDK MembersProperty.GetValue, reject every extension member without reading executable values; never user-shadowable Members/GetType access |
| Descriptors | Retain ScriptProperty/CodeProperty rejection and ordinal-ignore-case Length/Rank shadow checks; retain public-property traversal |
| Error order | Type rejection, then decorated/descriptor rejection, then Integer bound; no moved range comparison before safety inspection |
| Bounds/interface | Same four parameters/defaults, long range binding and comparisons, no new output or wrapping/catching exceptions |
| Context/ownership | Preserve default metadata creation and owned per-evaluation context; no global cache or retained input/result references |
| Other paths | No Token/Bool/Shape/evaluator/fixture edits or acceptance flags; no internal product clock/logging |

The reuse claim is limited to one synchronous helper invocation on its local
Value, with the ordinary owned SDK metadata context. It does not establish
hardening against hostile replacement of functions/metadata by arbitrary host
code, concurrent input mutation or live/runtime execution. Do not advertise any
of those as newly supported or silently weaken existing per-call checks.

## Direct regressions needed BEFORE changing the helper

- Direct clean int/long calls:default and explicit owned metadata, zero and
  endpoints/custom signed range; no success pipeline output.
- Direct null,bool,numeric string,float,unsigned integer and enum refusal;
  exact existing error categories, not merely an arbitrary thrown exception.
- Fresh owned wrapped integers with extension NoteProperty,AliasProperty,
  ScriptProperty or ScriptMethod (including Members/Length/Rank shadow names);
  fixture must prove decoration survived, refusal must not execute callbacks.
  Retain the CodeProperty guard even where fixture construction is constrained.
- Combined wrong-type/decoration and decoration/out-of-range cases to pin error
  precedence. Add clean-after-refusal and decoration-after-clean direct calls;
  never decorate cached shared literal values as fixtures.
- Retain complete scenario-level mutation/ownership/reflection/shape tests and
  original full-entry checks. A direct test PASS alone is not full acceptance.

These are proposed requirements, not written/run tests. A mutant that removes
integer member inspection should be rejected by direct tests even when the
scenario path still refuses the decorated object earlier. Reduced reflection
work needs independent characterization; no source-text count is runtime proof.

## Performance ceiling and next boundary

For the clean4096-time validation model, Integer is called twice per row, not
four times (the other two typed calls are Token and outside this candidate).
The intended static change avoids8192 second type queries,18 to16 per row,
73728 to65536 for a COMPLETE path. It retains40960 SDK Members reads and all
equivalent passive inspections. Fewer calls to the named Passive helper after
specialization must not be confused with deleting those inspections.
No wall-time improvement, stable latency or timeout cure is inferred. The prior
hot-loop model and incomplete Local run do not identify the unique root cause.

Verified the baseline signature/call chain, direct-call coverage inventory and
model arithmetic without invoking product/tests. All283 frozen pins held;
loading-context count stays1/2 spent,Inherited unrun/FROZEN. Old budgets and
review0/1 unused/ineligible unchanged. PS5/I1/P1/candidate blockers persist;
no main/C12/security/runtime/release acceptance.

Next present this small candidate in chat for approval. No implementation or
fresh verification allowance is granted by this audit; never reuse frozen slots.
Save/recover this audit with current status and runbook on the existing development
branch,not ignored raw/controllers/runtime caches.
