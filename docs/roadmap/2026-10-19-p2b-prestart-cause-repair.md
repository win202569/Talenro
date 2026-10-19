# Prestart / first-cause execution — 2026-10-19

Status: SCOPED_DEVELOPMENT_CHECKPOINT_ACCEPTABLE. Overall acceptance unfinished.

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
Source checkpoint94ca52f48b8e5f430f9c933a314e804dd7883892 pushed; author
verified remote equality and clean tree before review. Sole fresh readonly
review completed with0Critical/0Important/0actionableMinor. Independent review
confirmed1930GREEN mappings,30source snapshots,12host artifact hashes,
capture-only phase delta, complete literal outputs/groups and original seam
identities. Scoped dev checkpoint acceptable; no production/main acceptance.
NEW4hosts/7children/1review exhausted. No additional runtime/fix/review granted.

## Review disposition and author-only handoff

All seven review exclusions retained explicitly, not silently accepted:

1. Late-write/flush/close/concurrent failure paths remain unverified; no such
   guarantee inferred from successful saved publication.
2. Host death/hard-fileIO/descendant cleanup after termination remain unproven;
   fixed finite waits cannot promise those behaviors.
3. Classifier-exception/observation-limit paths inspected statically only;
   actual dynamic edge coverage remains incomplete.
4. Prefix/core and independent bootstrap RED coverage not reopened; no new
   broad suite claim from historical observations.
5. C1/C2 stay unclosed; R2/Minor2 stay deferred, so overall acceptance unfinished.
6. Later export/report/status/recovery author-checked only, not independently
   reviewed; cost: handoff export lacks a fresh independent reviewer.
7. No production/live/runtime/formal/main/full-C12 signoff; fixtures lack that
   authority and all five flags remain false.

No new minor finding. Existing R2/Minor2 remain deferred. These dispositions
and the three process rulings above are complete; no further fix is implied.
Raw review response and all original records preserved in ignored packet.
Final portable export/byte verification is author-only mechanical handoff,
not another target run, review, raw reconstruction or executable permission.

## Portable byte-preserving handoff

[Portable data](../diagnostics/isolated-fuzz/p2b-prestart-cause-repair/portable-summary.json)
contains all92packet originals:38JSON artifacts as raw UTF8 text and54byte
artifacts as base64, including all30source snapshots and sole review response;
19directory entries and7empty files preserved. Raw JSON numeric lexemes are
never converted through JS numbers. Embedded line endings/bytes unchanged.
All92decoded payloads independently compared byte-for-byte, length and SHA256
against originals,0differences. Wrapper physicalLF,2123937bytes,SHA256
0D38601783035306FE5165301A1A8364D21676BA0AE6AC77520D052DBD81E6E8.

One export serialization/CreateNew/write/flush/close succeeded. The first
author-only verification command then failed when its PowerShell if-output
enumerated an empty byte array into null, making HashData overload ambiguous.
The already-written wrapper and all originals were left untouched. A corrected
READONLY verifier assigns decoded arrays directly within branches; all92,
including7empty files, then matched. No re-export, source/oracle change,
additional host/child/review, allowance reset or evidence reconstruction.
This checker correction is not runtime acceptance and is not independently
reviewed. Original ledger snapshot intentionally ends before export; later
wrapper/commit/remote checks live here and in Git/tool evidence.

All original records remain ignored/preserved. Fresh clone obtains source,
tests, approved brief, complete portable records, review/disposition and these
recovery notes from the development branch; not dependencies, credentials,
live sessions/acks, original private process objects or executable permission.

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
Portable handoff is data only, never raw reconstruction or replay rights.
C1/C2 unclosed; R2/Minor2 deferred. All five capability flagsfalse. No main merge,
compile/install/production/live/formal acceptance. Overall project unfinished.
