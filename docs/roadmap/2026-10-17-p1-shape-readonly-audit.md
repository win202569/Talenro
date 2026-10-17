# Read-only shape lookup audit — no confirmed public-path semantic regression, acceptance blocked

User requested read-only review afterb3e662b1. ZERO test processes, no product/
test/probe/controller edits,replay,install,security change or budget spend.
This is the implementer's self-audit, NOT the separate independent review gate.

## Static review outcome

No confirmed runtime correctness/security regression was identified in the current
public Get-OfflineSupervisionDecision call path. This is a limited static finding,
not proof of absence and not acceptance. No product change is recommended solely
from a speculative timeout cause.

| Invariant | Source evidence | Limit |
| --- | --- | --- |
| Cache lifetime | Entry creates fresh Metadata each call;ShapeSets factory allocates a new dictionary | No concurrency claim |
| Trusted schema | Root key is literal;observation Kind validated as exact passive ordinal token before key use;Names built from fixed literals/switch | Assumes existing single-threaded synthetic contract |
| Bounded cache | Public path has root plus8 admitted Kind values | Internal helper accepts arbitrary caller-supplied keys |
| Exact names | Dictionary and HashSet use StringComparer.Ordinal | PS5 not executed on candidate |
| Shape/count | Exact object type and original Names.Count checks occur before field membership | Original schema literals must remain consistent |
| Member safety | Name membership AND NoteProperty kind checked before reading p.Value;passive value check still called | Does not replace full malicious-input suite |
| No acceptance cache | Hit reuses allowed names only;each object/value still checked | No per-input result/type memoization |
| No result exposure | Metadata not returned or stored globally;result still has original fields | Internal factory itself remains callable like other helpers |
| Decisions unchanged | Entry differs only in two ShapeKey arguments;state machine and other helpers exact | Static comparison is not runtime proof |

AST/source comparison covered13 functions. Only New-SupervisionPolicyMetadata,
Assert-SupervisionPolicyShape and the two entry call arguments differ from sealed
baseline6CF0D354. All other helper function texts match exactly;removing the two
ShapeKey arguments restores the complete original entry function text. No product
timing hooks were introduced.

### Maintenance caveats, not confirmed current API defects

The cache helper assumes one stable Names list per nonempty ShapeKey within a
Metadata instance. It does not independently bind/compare Names on a cache hit.
Current two callers satisfy that invariant:fixed root schema and fixed per-Kind
schema after token validation. Reusing a key for different fields in a FUTURE
internal caller would be unsafe. That caller would need review; this audit did
not invent a current external injection path,run a reproducer or change helper
interfaces. Do not allow new callers to treat arbitrary input as schema keys.

Header line2 still says only fixed SDK metadata is reused;it predates schema
allowlist reuse. The detailed inline comments describe the new behavior. This is
a minor documentation follow-up,not evidence that input acceptance is cached.
Frozen measured candidate was not edited to correct wording during this review.

## Test coverage and outstanding evidence

Saved baseline/candidate receipts are unchanged:baseline new6 PASS1045ms;
candidate full PS7 timeout30071ms,CPU38796.875ms,77 lines/59 PASS. New tests are
real literal-output checks,no mocks or source-text assertions:
root ordinal casing,missing/extra/same-count fields,observation base-name casing,
mixed time/control schemas with later malformed instances,repeated calls and
result ownership,zero-callback script descriptors.

The new mixed-schema group covers time/control explicitly,not all eight kinds
on its own. Completed original35 exercise other kinds. This does not remove the
missing overall completion or PS5 requirement. New repeated-call tests do not
replace the specific decorated-scalar SequentialIsolation matrix.

Still absent on candidate:
- ownership SequentialIsolation:decorated ExitCode/NaturalCompletion/
  Classification,repeated refusals,result independence and clean recovery;
- ownership PlainControls:plain/long/empty/unknown-exit/sticky-failure controls;
- ownership Complete and both wrapper completion markers;
- PS5 execution,independent review and remaining frozen acceptance requirements.

No saved assertion failure preceded the kill. Last completed PassiveDescriptors
does not locate the exact executing instruction at timeout. Those missing groups
are UNVERIFIED,not known incorrect and not assumed passing.

## Performance interpretation

Field-name lookup replaces a short repeated list scan with a hash set and adds
dictionary/set allocation/lookups. Fixed schemas have6–13 fields;this reduces
some repeated comparison work but still checks every4096 observation and value.
Both old/new public evaluation remain linear in observation count for fixed
schemas. Type reflection,scalar member checks,descriptor validation and copying
remain. No set-hit/miss/allocation/per-function CPU timings were measured on this
candidate;there is no demonstrated net speedup.

Previous full tests and new full wrapper are not matched benchmark workloads:
new wrapper includes six added contract groups. More completed PASS groups before
the deadline and cumulative CPU differences do not establish an improvement.
Historical engine/load/JIT/instrumentation uncertainty remains. Do not append
another speculative optimization from an unproven bottleneck attribution.

## Integrity and next step

173 immutable pins,controller,two invocation rows and saved streams rechecked.
Candidate953CC2A8,test9DAA38EF,sealFB089A89,index498FA51C unchanged. No historical
verdict changed and no PS5 child was launched. Only this audit/report and progress/
recovery notes are updated and checkpointed. Ignored packets/raw/runtime remain
local;clone restores portable code/reports,not runtime or replay authority.

Current shape budget2/3/one frozen;prior independent full1/2/one frozen,
reflection3/4/one frozen,extended-view3/7/four frozen and all diagnostic counts
unchanged. Review0/1 remains unused/ineligible. Candidate NOT accepted,I1/P1
unresolved;no main/C12/project/live/runtime/release acceptance.

Prioritize obtaining the missing decorated-isolation/plain-control and PS5
semantic evidence with a separately approved targeted original-suite scope,
without editing candidate or weakening checks. Existing ownership script
supports regression-only invocation,avoiding the large original35 boundary case.
Such targeted completion would ONLY close evidence gaps for that subset:the
unchanged full30s gate would remain blocked and must not be relabelled PASS by
splitting tests. No targeted process,helper fix,new budget or old-slot resume
was performed or authorized by this report.
