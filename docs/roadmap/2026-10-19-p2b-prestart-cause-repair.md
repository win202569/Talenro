# Prestart / first-cause execution — 2026-10-19

Status: DYNAMIC_EXACT_REVIEW_PENDING. No full-C12 or runtime acceptance.

Human approved the short in-chat design and NEW4hosts+7children, conditional
1readonly review. [Approved execution brief](2026-10-19-p2b-prestart-cause-budget.md)
SHA256 011CCA498D3745DD4DFB12E86045783015636395353939C0F51E53C1D4C5457A.
Base c1203e71f61f59fc75580b6392347750aef32c9f on existing development branch.
Old remainder1host/4children/review remains frozen; no old source/raw/report
or portable bytes were changed. The old actual prestart-order failure is
preserved, not relabelled or rerun in its old stage.

15LF inputs/static parse0errors before first freeze.1930mapping freeze:
inherited1795+old103raw+old15leaf+old report1+approved snapshot1+new15inputs.
Two snapshots of all15sources; GREEN changes capture.ps1 only,1929other
mappings unchanged. Tests/oracles/driver/session/gate/executor immutable.
Owned/sequence reuse GREEN input identity; never recover Session/Ack from disk.

| Grant | Host PID | Exit | Exact outcome | Children |
| --- | ---: | ---: | --- | ---: |
| classification-red |41924|2|FAIL identity; fixed assertion primary; actual gate error retained in step/session|0|
| classification-green |14396|2|PASS same identity; actual gate error primary, assertion sole secondary|0|
| owned-green |34096|0|all6groups PASS, prestart FIRST; actual3owned targets exited/drained|3|
| sequence-green |7804|0|all4sequence groups and original12seam PASS, fixed whole output/trace|4|

Both hosts exited/drained with no outer termination/rescue/transport error,
all1930pins intact pre/post. Whole stdout/stderr, ordered groups, schema,
error type/id/message, actual root and lifecycle/IO counters checked against
fixed literal oracle. Only these complete exit2 signatures are expected.
One actual reserved step,StartIO1,OutIO0,StartCalls0,ack0,Observation=null;
unchanged host child cap0 throws before Process creation. No fabricated errors.

RED manifest DC2EEC699B31DCCEDF6F2DEB3DE0A94A59CA8443F6DE08A27CFBDFA69A553C9C.
GREEN manifest 01B99203F29B6073843756E347054F8489F5825C325B6F4084DF9C8018A07123.
RED evidence 32788AED5DB8A6395B7060ECD7B0033681117E4D48653F138CD54775CD430BDC.
GREEN evidence 7FB58FCCCA74FA0A1E5779C0CB77EE252E45FA78D88E04FD912184F586309F14.

The test consumer classifies fixed receipt/open/output negatives using the
actual failed sink, actual ErrorRecord type/id/reference, private session
snapshot and exact lifecycle tuple. Unknown step errors are promoted before
caller assertions. No classifier IO, grant/ack mutation or authority setter.
The common probe dynamically covers unexpected gate-error promotion, not all
possible classifier/late IO faults. Broader expected-negative paths depend on
the separately authorized owned/sequence runs, now both exact.

## Owned result

Prestart first actually reached File.Create/UnauthorizedAccessException at
sinks.ps1:62:Out0Failed/IO1,StartCalls0,actual Unassociated/PIDnull/Kill0/Wait0.
No host-primary promotion for this fixed expected negative. Then receipt child
20680 exited-1/drained in44ms, one failed receipt IO, no secondary, same actual
per-session cause, exact owned kill/wait. Normal10696 exit0/drained in804ms,
19bytes each stream; timeout40456 exited-1/drained in215ms, bounded termination,
valid literal prefixes. No extra target, outer termination, transport error or
rescue; whole6group order/evidence exact and all1930pins intact pre/post.
Owned evidence6BEE4E38165E9AA08D5AF3D0341B351A845DFFCB00F95931B60A5BC51DF7181E.
## Sequence result and review boundary

A0 child23540 exit0/drained459ms, B0 child8536 exit0/drained350ms, C0 child45024
exit0/drained335ms before actual ConsoleOut ObjectDisposedException, A1 original
seam child41696 exit0/drained4780ms/SeamExacttrue. All4associated and distinct.
Six invalid live requests rejected without consuming genuine A0 ack; live A1
consumes it once; replay rejected. B budget prevents second target. C gets no
ack after actual output failure and cannot continue. All closed Session handles
and fresh disk-positive view cannot resume. Four final sequence groups exact;
original12seam PASS and full raw stdout/trace hashes unchanged. No mocked ack,
disk success continuation or extra target. Host7804 exited/drained with empty
stderr, no outer termination/rescue/error. Sequence evidence
0A1D363FF423253FBA1207AF994977C641A62F37B798494AF014D28E8C3104C0.
All1930pins verified intact before/after all grants and again at final readonly
check. NEW4hosts/7children exhausted; conditional1readonly review now eligible.
No more runtime tests, targets, fixes or review rounds in this scope.

Runtime verification is scoped: classification RED-to-GREEN identity proof,
owned6groups, sequence4groups plus original12seam groups. Prefix/core tests
were preserved but NOT rerun in this four-host budget; prior14PASS remains
historical evidence, not a new broad suite. No full C12 acceptance inferred.
Source checkpoint before review; portable export and sole readonly review
pending. Review findings requiring code change will block, not authorize fixes.

Manual finite PowerShell ledger replaces default Bash/setup/full-suite/task-done
reruns, deletion and extra fix/re-review rounds — approved scope/preservation
rules — cost: narrower coverage. Existing isolated worktree reused, no install
or broad baseline — cost: no broad suite. Routing/consumer seam and reorder
assembled before freeze; old failure plus common classifier probe cover scoped
repair — cost: bootstrap paths not all RED-tested. These are process rulings.

No PID/name lookup/tree kill/watchdog/Job or widened timing. Parent observes
only exact capture hosts; child lifecycle comes from frozen executor, not an
independent OS inventory. FileIO is not hard realtime; outer host termination
cannot prove all descendants exited. All originals remain ignored/preserved.
Portable handoff pending, data only, never raw reconstruction or replay rights.
C1/C2 unclosed; R2/Minor2 deferred. All five capability flagsfalse. No main merge,
compile/install/production/live/formal acceptance. Overall project unfinished.
