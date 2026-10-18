# P1 candidate single readonly review — stopped on Important R1

ReviewCompleted=true;ReviewVerdict=STOP_IMPORTANT_OWNERSHIP_EVIDENCE_GAP.
I1CandidateDisposition=InsufficientEvidence.
FormalAcceptance=false;PromotionApproved=false.

User approved the [single-review plan](../superpowers/plans/2026-10-18-p1-token-single-readonly-review.md)
afterca32f2d96f2ab551f049389d4aa22dbc13bbff16. That HEAD and
[review package](2026-10-18-p1-token-review-package.md) were frozen before dispatch.
One NEW fresh no-history reviewer:/root/p1_token_readonly_review_20261018,
gpt-6-astra/high,completed the source-and-saved-evidence review.
NEW review1/1SPENT. ZERO target executions,source/test edits or automatic repairs.
No replacement reviewer,re-review,candidate promotion or main merge.

## Independent finding and parent disposition

Critical:none identified. Important:ONE(R1). Minor:none separately identified.

### R1 — admission-time refusal does not establish ownership of accepted scalar leaves

Relevant candidate:
docs/diagnostics/isolated-fuzz/integer-token-local-candidate/offline-supervision-policy.ps1.

- Line173 shallow-copies observation property values.
- Lines191/235 directly assign Classification into sticky/normal result leaves.
- Lines216/231 directly assign other admitted scalar values.
- Ownership test225–233 decorates input BEFORE both evaluations,which then return
  refused objects with Decision=null.
- Full test176–179 replaces input fields/edits a Decision field;it does not
  attach mutable metadata to an initially plain admitted scalar AFTER return.
- P1-Equal/P1I1-Equal compare primitive type/value,not primitive output extended
  members. Fresh root/Decision objects alone do not prove scalar-leaf isolation.
- P1 contract72 requires no returned input-object/array reference;implementation
  step138 requires independently owned scalar objects.

Unresolved vector:initially undecorated caller-owned scalar in a valid scenario;
evaluate twice;then sequentially attach mutable extended metadata through the
original scalar or one returned scalar. Establish whether another prior result
can acquire/access that metadata. Normal and sticky output paths matter.
This is an ordinary sequential ownership question,not excluded concurrent
mutation or hostile-runspace/type-data modification.

**Confidence boundary:** shallow dataflow and missing assertions are established
by source inspection. Actual cross-result visibility of later-added metadata
on this exact PS7 candidate was NOT dynamically observed. R1 is an Important
ownership-EVIDENCE gap,not a reproduced runtime failure,confirmed exploit or
demonstrated authority escalation. Do not state that aliasing definitely occurs.

**Parent disposition:retain Important and STOP UNFINISHED.**
Parent read the exact assignments,contract72/138,ownership225–233,full176–179
and comparator definitions without invoking them. They support the reviewer's
coverage/dataflow observation. Existing positive tests do not answer this vector.
No sound basis to downgrade it merely because prior suites passed;no basis to
claim runtime reproduction either. No source repair,test/probe or second review.

The already-decorated-input rejection path has substantial positive evidence:
fixed SDK descriptor lookup,refusal before scalar use,field coverage,exact
refused objects and callback-zero outcomes. I1 is therefore neither globally
closed nor dismissed as wholly unaddressed. Retain InsufficientEvidence for the
current candidate,including the proposed PS7/offline disposition. The original
historical Important I1 verdict and all stopped-route outcomes remain recorded.

## Strengths retained,not overstated

Reviewer found fixed nonexecuting extended-member inspection,root preflight and
Kind checking before dispatch,cumulative Integer/Token checks/signatures retained,
and full validation before reduction. Direct fixtures retain descriptor/payload
identity,literal refusals,no helper output and callback-zero/CodeProperty checks.
Saved API/ownership suites retain full literal results,refusal indices,4096/4097,
conservative flags and decorated-input sequential isolation. None replaces R1.

## Evidence independently checked

Reviewer read the complete approved plan/package,P1 contract/original I1 review,
scalar-repair plan,joint design9/10,current candidate and historical source,
direct/full/ownership assertions and the scoped saved evidence.

Matched51 scoped and476 inherited pins before/after. CandidateF9B4E408,
integer baseline3537E9F5,current original953CC2A8,historical snapshotFE23DF3B
identities matched. Three current variants have13 functions;original→integer
changes only Integer function text,integer→Token only Token. Historical reviewed
source has12 functions,so no claim that current original was independently accepted.

Every saved output line matched its frozen oracle:direct429/429 each;
full/ownership38/22/38/22. All12 stream hashes matched;all6 stderr files empty.
Receipts consistently record exit0,no timeout,complete capture,owned-host exit,
no controller error. These are SAVED claims,not reviewer-observed executions.
Portable copy hashes/routing,fixture revision,CodeProperty source,literal/index/
bound assertions checked. Parent rechecked51+476pins at return,without reruns.

Historical Git-object provenance,HEAD/clean branch/origin/remote remain parent
provenance,not independently authenticated by reviewer. Reviewer hash-checked
the supplied historical text snapshot. No trust setting was changed.

## Every declined judgment and parent disposition

The reviewer explicitly retained all18 original considerations. Parent retains
each boundary below;none is silently converted to a pass or a source-defect excuse.

|#|Reviewer declined / reason|Parent disposition and consequence|
|---|---|---|
|1|Current execution;imports/runs forbidden|Saved evidence only,no new runtime claim|
|2|Exhaustive vectors;finite source/data|Bounded coverage;R1 separately blocking,not dismissed|
|3|QPC conversion/cross-process clock agreement;synthetic ms|Unproved,separate P2|
|4|OS containment/startup races/handle inheritance;no live adapter|Unproved,no OS approval|
|5|Actor/session authenticity;declared labels|No identity-authentication claim|
|6|Concurrent pipes/backpressure/byte-event provenance;aggregate metadata|No real capture proof|
|7|Cancellation/termination/physical cleanup;Would* only|No live effects;actual cleanup unknown|
|8|Hard wall-clock/scheduler;modeled/saved successful hosts|No physical-time guarantee|
|9|Durability/crash atomicity/physical readback;modeled booleans|Unproved,separate persistence gate|
|10|Harness timeout/kill/drain;recent6hosts did not time out|No broader dynamic timeout verdict;older records not amended|
|11|Nonempty CLIXML;recent stderr empty|Parser branch not demonstrated|
|12|Catastrophic recovery;ordinary catch insufficient|No catastrophe guarantee|
|13|Concurrent caller mutation/hostile runspaces;excluded/untested|Remain excluded;R1 ordinary sequential case is NOT excluded|
|14|Historical authenticity/mutant chronology;hashes not attestation|Consistency only,no authentication of past executions|
|15|Protected dependency behavior;hash checks only|Integrity matched,no fresh behavior clearance|
|16|Git base/HEAD/branch/origin/remote;not independently authenticated|Parent Git provenance only,no security/trust bypass|
|17|Historical whole-branch/project correctness;outside exact scope|No whole-project clearance|
|18|Security-owner/P2/P3/runtime/main/C12/release;separate gates|Remain open/unapproved|

Additional five declined limits and parent dispositions:

|Limit / reason|Parent disposition|
|---|---|
|Token PS5 compatibility;new records PS7 only|Unverified/frozen;Integer evidence does not transfer|
|Reflection/shape extra-suite execution on Token;not observed|Not passed by static inspection,no new run authority|
|Token-specific negative controls;absent from recent packets|Older Integer omission control supports only that mutation|
|Performance/earlier combined timeouts;receipts not matched benchmarks|No speedup/root-cause/timeout-fix claim|
|PowerShell binding/descriptor cases beyond covered calls|No universal engine-equivalence claim|

## Immutable review artifacts and preservation

Ignored packet:.superpowers/sdd/2026-10-18-p1-token-readonly-review/.
The saved reviewer-response.md preserves returned response text with LF file
serialization,including original links;this portable report records all findings,
confidence limits and declined-judgment dispositions for fresh-computer recovery.
It does not rewrite the reviewer's outcome or imply original transport-byte
attestation. No raw/session artifacts are committed.

|Artifact|SHA256|
|---|---|
|Approved plan|FA4BDEA649AE4203C4B3C90DFE9B85FA2FAD2244D6420954B898A2706582ABE3|
|Portable review package|BF285E360D2D2A09A169F2B387DF4816A071F460F37906F393179DBFF858A132|
|manifest.json(51scoped pins+inherited476 reference)|3AEC22EAA09003FC2BE6B03C32AD4C9728DFC4E660AC0340A6D124D48CBCC0B2|
|dispatch.json(consumed BEFORE dispatch)|8F67EC59E31013E00CAF429C7651B8A101E2CE5AC3370A93C4A4389F2B308599|
|reviewed-base.ps1.txt(NONEXECUTABLE)|FE23DF3BE1F5BA5DF4C2B565DEE5CFB46DAB6F57C896DA9B5C95DFE440FF0ABB|
|reviewer-response.md|736C72EA10BCCAE9174BAC5BF8AF88869730D609B9559480F12B2DD327815D85|
|result.json|865BC62FB94B2C8CD2EFD24D4B207200BB129C88C84FEB72653D68C0CBDB7D50|

No historical packet/controller/index/stream/source changed. All prior full4/4,
direct2/2 and old failed/frozen slots remain. Old conditional review0/1UNUSED/
INELIGIBLE and original spent1/1 are distinct from NEW review1/1SPENT.

Ruling:explicit single-seat/zero-host/readonly terms override generic skill defaults
for reproduction,auto-fixes,re-review,whole-branch audit,new worktrees/Bash helpers
and scratch deletion. Cost:the ownership question remains unresolved;formal
acceptance stays withheld. Ruling:R1's sequential case is not covered by the
concurrency exclusion;retaining it is a source/evidence disposition,not a dynamic
aliasing assertion. No Minor findings were deferred.

## Stopped handoff and next recommendation

This authorized review/handoff is complete;the candidate increment remains
UNFINISHED/UNACCEPTED. No automatic promotion,original replacement,merge or release.

Next recommend a NEW narrowly bounded ownership-characterization PROPOSAL for
initially plain scalar leaves:two accepted results,post-return metadata attachment,
normal/sticky output paths,field identity/descriptor/payload checks and callback
zero. First design the independent literal oracle and finite fail-stop budget.
Only if separately approved may a future test observe the question;only if aliases
are established should a separately authorized correction be considered.
Do not reflexively reconstruct/cast values or weaken current scalar refusal.

Git carries package/report/progress/recovery on latest dev branch,not main.
Ignored raw/controller/session/runtimes/credentials and test/review authority do
not clone;missing packets cannot replenish consumed or frozen allowances.
