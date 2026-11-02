# Contract call-model scope amendment

Status: PROPOSED_DESIGN; user approved drafting this amendment, not its written
design, implementation plan, execution or added auxiliary allowance. Small
readonly-model/binding amendment; native execution retained, no extra agents/reviews.

## Evidence and purpose

Baseline1fe01267301bd0cb13f577a3f00198d2cf8e467a; see
[actual projection execution](../../roadmap/2026-11-01-offline-error-projection.md).
The replacement writer is genuinely accepted:26PASS, parent/childexit0,
Exact/EvidenceValid true, original streams and recorder/source checks complete.
EvidenceSHA73DBE760267684A9CFE675C5435921C836CAF36CEB2F95C506DE2D8906BBA106.
Do not rerun, relabel or refund that attempt. Whole product/I1/I2 is not accepted.

Only an untested observation-contract.ps1 candidate followed. Readonly inspection
found an actual saved IntentFailureNoWork stack frame whose caller is
Invoke-ODOwnerControls at record-boundary.tests.ps1:162. Its CommandAst invokes
variable controls with a scriptblock; GetCommandName is null. Read-ODSourceModel
does not describe that variable-call boundary. This is a static model gap, not
a newly observed failed contract run. Contract Prepare/host has not happened.

Goal: describe that fixed call from independently parsed, pinned source and
validate its exact boundary without evaluating the variable or guessing a line.
Non-goals: run the control body, import a live module to learn its identity,
resolve arbitrary dynamic calls, accept any function-range line, relax error
source/operation/stack checks, or convert the unit context into runtime authority.

## Approaches

Use the narrowly scoped readonly descriptor direction discussed with the user.
Only the fixed owner-control case call receives a descriptor, exposed only in
the explicit IntentFailureNoWork unit context. Existing named Calls and fixed
entry/record-factory handling remain unchanged.

A generic variable-call resolver would widen the model and require dynamic
execution/identity policy beyond this repair. Ignoring the frame or accepting
any line in its enclosing function violates the existing strict contract.
Those alternatives do not fit the agreed constraints; no bypass is proposed.

## Descriptor and acceptance boundary

source-model.ps1 reads pinned source/AST only. Add a separate DynamicCalls data
collection; do not change the ordered shape of existing named Calls. A descriptor
records source identity, enclosing owner function, exact module-variable name,
invocation line/column, and the literal scriptblock's start/end extent.

Recognition is restricted to the pinned record-boundary.tests.ps1,
Invoke-ODOwnerControls, variable controls, and the case scriptblock whose ordered
parameter list is exactly Name. Uniquely distinguish this call from the initializer
(Definitions/ProfileText/etc.) and cleanup call. Zero/multiple matches reject;
unknown variable calls receive no permission. Derive all positions from AST,
never from the supplied report/stack or hardcoded line162.

New-ODIntentUnitContext binds the unique descriptor as explicit unit-only data.
The pure contract may accept the module bridge only when the caller's source,
function and exact invocation line match, the previous frame is the literal
scriptblock in the same source and within its AST extent, and the context carries
that exact binding. Keep precise first-error operation AST and all other named
call-chain checks. Missing/foreign/ambiguous/out-of-extent bindings reject with
the existing observation-contract mismatch signature. No blanket catch converts
implementation errors into valid rejections. New-ODRuntimeContext does not gain
unit bindings or an anonymous-source fallback from report contents.

## Source and provenance scope

Replace the remaining contract-only window with exactly three files:
source-model.ps1 (readonly descriptor), observation-contract.ps1 (existing
untested candidate plus strict descriptor consumer), driver.ps1 (current epoch,
historical-source validation, approval/stop/provenance gates). Other28of31runtime
sources, tests and oracles remain frozen at the accepted writer epoch.

Keep writer-projection-repair.frozen.json and its accepted final/streams/evidence/
probe/journal immutable. Historical scope/projection receipts must validate their
old source hashes against snapshots/writer-projection-rebind, not edited current
source-model/driver bytes. Preserve existing historical recovery maps unchanged.
The new contract manifest separately binds the changed three sources and added
immutable amendment provenance; never regrade the accepted writer against it.

Keep the still-absent fixed repair-green.frozen.json and unspent contract-green/
core/lifecycle/continuation output slots. This avoids changing entry, session,
test or profile manifest selectors. Permit one new fixed preparation grant for
the three-source repair; old writer preparations remain spent. Exact schemas,
pin count and preparation identity belong in the later implementation plan.
Do not embed a current driver's hash and its containing input hash circularly.

Preserve all seven stops, including the static error-projection stop, all actual
records/archives and155snapshots. Recover that historical stop only through exact
written approval/hash binding; a newest contract-call-model stop always blocks.
No public IgnoreStop, arbitrary path/callback or old attempt replay.

## Proposed verification and allowance

Request only TWO additional finite auxiliary computations: old-model RED once,
new-model GREEN once (cumulative auxiliary ceiling11,9already consumed). Use
the actual extracted source-model functions and independently pinned AST plus
saved writer input; readonly file/AST access is explicit. No profile/control
body, record writer, module probe, child host, compilation or target execution.
Freeze new regression/input identities before RED; preserve actual first error,
completed groups and full tool chunks, no patch/retest after unexpected results.

GREEN must prove one exact descriptor, initializer/cleanup exclusion, unchanged
named-call handling, no unknown-variable authorization, and acceptance/rejection
of the actual saved unit-error stack through the pure consumer with independent
wrong-source/function/line/body-extent/missing-binding mutations. Runtime contexts
without unit metadata must not accept the special bridge. Test data acceptance
is not runtime or product acceptance. Exact group/receipt definitions will be
frozen by the implementation plan before either computation.

No additional runtime host/compile/target/reviewer/snapshot requested. After
auxiliary acceptance, consume the existing remaining31snapshot contract epoch
(total186), then original contract52PASS/core25PASS/lifecycle17PASS/continuation
16PASS+12seamPASS chain:4hosts/7targets/0compiles/conditional1review. Cumulative
runtime ceilings stay7hosts/3compiles/186snapshots; actual consumption remains
3hosts/3compiles/155snapshots/0targets/0reviews during drafting. Preserve all
original process ownership, time limits, source/oracle/recorder predicates and
five false-capability flags. First unexpected result stops, without extra review.

Success or stop requires a NEW actual-record portable archive and explicit-file
development-branch checkpoint/push with remote equality. No main merge, history
rewrite, deletion, install, security change or broader acceptance. Written design
approval precedes a separately reviewed implementation plan/allowance approval;
this document itself authorizes no execution.
