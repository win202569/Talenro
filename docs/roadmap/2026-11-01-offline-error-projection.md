# Error projection — actual execution

Approved [design](../superpowers/specs/2026-11-01-offline-error-projection-design.md)
and [native plan/allowance](../superpowers/plans/2026-11-01-offline-error-projection.md).
Baselinecf1e5ce0; this is not product/I1/I2 acceptance. Six historical stops,
old actual failures and all archives remain unchanged.

## Latest — writer accepted; contract source-model scope blocker

Source-freeze checkpoint28d4cf73b7be512f3178692e0963f03f0ecb212e was pushed
and remoteSHA verified before the replacement host. Writer-projection-green ran
once (host29400): parentexit0/childexit0, Exact=true,EvidenceValid=true,
RequiredExact, owned exit/complete streams/no timeout/termination/rescue,
post-pins intact,26PASS and exact original stdout/empty stderr. Actual evidence
SHA73DBE760267684A9CFE675C5435921C836CAF36CEB2F95C506DE2D8906BBA106.
The error-projection repair has passed this writer boundary; wider I1/I2/product
acceptance remains NOT_ACCEPTED and historical failures are unchanged.

Original contract implementation was then released. Only observation-contract.ps1
was authored as an UNTESTED candidate, parsed successfully, and not source-frozen.
Static review found an actual retained IntentFailureNoWork stack frame:
Invoke-ODOwnerControls at record-boundary.tests.ps1:162 invokes variable controls.
Independent AST extraction shows CommandAst.GetCommandName=null, first element
VariableExpressionAst/controls. Current Read-ODSourceModel records named calls
and only the fixed entry/record-factory variable aliases, omitting this boundary.
Strict full stack validation cannot match that omitted caller descriptor. Adding
it to the mapper is outside the approved contract-only source window; neither
guessing a location nor loosening the stack check is authorized.

Stopped BEFORE contract Prepare/host. New seventh error-projection.stop.json
records a static scope blocker, not an invented thrown runtime exception.
No contract/core/lifecycle/continuation host, target or review ran; no186snapshot
epoch exists. Cumulative9auxiliary/3hosts/3compiles/155snapshots/0targets/0reviews;
remaining4hosts/7targets/0compiles/conditional1review frozen. Do not rerun the
successful writer or treat the untested contract candidate as accepted.

Post-candidate bulk readonly pin reconciliation produced no output during bounded
waits and was interrupted; it did NOT complete, so no fresh all-pin integrity
claim is made. Writer post-pin acceptance refers to its actual completed run.
Next requires a written scope amendment for independent readonly dynamic-call
descriptors in source-model.ps1 and necessary new-epoch controller bindings,
before any new edits or execution. This is not approval of another attempt.

Final portable-error-projection.json preserves449actual P/Q files,125directories,
31zero-byte files; LF container and every decoded byte/hash/directory verified.
SHA FD9B1CD5D6F08F419B816F0C788561CA5AFBE1AE1305AC43EDD2A9F343F40310.
Earlier freeze/archive files are unchanged. The original full tool results,
writer evidence and static-blocker record are included, not reconstructed.

## Source-freeze checkpoint

Regression frozen at0e636f61bba4c6f4e6c888077fe08f50e7b352a8 beforeRED;
testSHAC27D74F4AC49DCE668747D5E212C986034F594C29047CC6E9BF3AD7E3AA8E995.
RED0PASS/1FAIL/9NotRun/exit1: actual PropertyNotFoundStrict/PositionMessage
failure from old extracted Capture-Value; complete original/tool record retained.
This characterizes the source defect, not a reconstruction of the old failed object.

Candidate source/integration frozen ataec023cb245071f5a3684e04671e12245b1fff1a.
Projection GREEN10PASS/0FAIL/0NotRun/exit0: truthful absent location/stack,
real caught metadata preserved, live wrapper/exception identity, nested JSON,
four actual strict-consumer null/missing negatives and unsupported-value rejection.
Four verifier helpers and two strict consumer definitions unchanged from baseline.
Full scope GREEN9PASS/0FAIL/0NotRun/exit0 on actual new five source hashes.
No source edits afterGREEN. All9cumulative auxiliary computations consumed.

Unique Prepare writer-projection-rebind succeeded:2251pins/31new snapshots,
total155/186. Parent7F09A142E228FF16CD7228287DEBBB5EE60DEC9A91FBF1EF691F148C0A911587.
New manifest182B85886304BFE3CBFB248DDA70F2EB99DF3E2158F574DBE299519DE7CE0461.
Exactly the original five operational sources changed, other26unchanged;
parent OwnerScopeRecovery preserved,15new provenance pins/actual phase bindings added.

No new writer host/compile/target/review yet at this checkpoint. Historical
2hosts/2compiles consumed; remaining5hosts/7targets/1compile/conditional1review
under cumulative7host/3compile/186snapshot ceilings. First unexpected result stops.

Portable-error-projection-freeze.json in the same diagnostic directory contains
417actual P/Q files,108directories,20zero-byte files; LF container/decoded-byte
equality verified. SHA E13D03435C988D3EF5F69B2777F6480577170CC2B4482F8C24F5975CD887294E.
Restore progress only from the existing development branch, not main. No replay,
overwrite or conversion of historical absolute manifests into new acceptance.

Rulings: approved finite PowerShell workflow replaces default suite/install/
cleanup helpers; full input contains current driver hash, so driver validates
immutable test/approval/input identities through actual phase intents/receipts,
then pins the input in the new manifest rather than introducing a circular hash.
Cost if wrong: manual ledger reconciliation or stop on any identity mismatch;
neither ruling authorizes relaxed source checks or extra computation.

Next: push this frozen checkpoint, then one original writer-projection-green
host/compile; only actual complete parent acceptance releases later stages.
