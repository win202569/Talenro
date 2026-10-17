# Full synthetic capture archive — execution evidence

## Authority and current status

User approved [written design](../superpowers/specs/2026-10-17-synthetic-capture-store-design.md)
afterb33e04bd,then chose1/native after
`aeb6fe363ac50b60e929f0123f762a286849a128`,approving the
[plan](../superpowers/plans/2026-10-17-synthetic-capture-store.md),12 NEW sequential
hosts120000ms each and one final readonly review,no retry/fix reserve. BASE is
aeb6fe363ac50b60e929f0123f762a286849a128. Existing dev worktree reused;initial
tree clean,new synthetic-store root absent. Old/frozen budgets unchanged.

**STOPPED / UNACCEPTED:** slot6 PS5 boundary hit its120000ms cap.6/12 consumed;
slots7..12 frozen,dependent regressions and independent review NOT started.
Only new source/tests and preservation docs changed. No fix/retry or acceptance,
production/main merge claim. This is an unfinished recoverable development checkpoint.

## Implementation boundary

New synthetic-capture-store.ps1 uses fixed TLRSYN01 payload/TLRSYC01 52-byte
marker,8MiB total cap and per-field limits from design. Full20-field records,
ordered trace,stored Lifecycle/Binding/authority facts are restored. Snapshot
metadata is reconstructed canonically from exact raw bytes;disk-derived facts
are not silently replaced by recomputation. Existing public contract verifier
must accept the whole restored record before it is exposed.

Dedicated ignored root/direct-child path checks,CreateNew,exclusive writing,
same-handle link/attribute validation,read locks,payload reread and marker-last
publication;stopped overrides a complete marker. No overwrite,repair,migration,
file deletion or executable interpretation. In-memory type/property/length
checks do not dispatch caller getters/GetType/ToString. New module imports
contract/isolation only;old source/schema/API untouched.

## Verification record

Frozen new groups:format18PASS,boundary73PASS including API/completion checks.
BothREDs intentionally stop at missing public APIs;not every case independently
demonstratedRED. Static syntax checking corrected one parenthesis before first
host/ledger slot existed;not a failed or retried test host.

|Call|Host/phase|Exit|Elapsed ms|PASS|Stderr characters|
|---|---|---|---|---|---|
|1|PS7 RED format|1|686|0|358 expected error|
|2|PS5 RED format|1|1017|0|537 expected error|
|3|PS7 GREEN format|0|5080|18|0|
|4|PS5 GREEN format|0|5264|18|0|
|5|PS7 GREEN boundary|0|54429|73|0|
|6|PS5 boundary TIMEOUT|-1|120180|28 partial|0|

Observed7.6.5Core x64/5.1.26100.9711Desktop x64. Only complete passing groups
are format18 oneachhost and boundary73 onPS7. PS5's28PASS are partial,not a
completed GREEN;the named terminal check was not reached. Firstformat PS7
ROOT_PRECONDITION existed=False dynamically verifies invalidRecord does not
create missing root;PS5 existed=True verifies preservation of existing root,
not a second missing-root branch. No root removed/moved to manufacture coverage.

The parent WaitForExit120000ms expired and terminated its testhost;120180ms is
observed total including shutdown/bookkeeping,not an increased allowed deadline
or Windows hard real-time claim. Childstderr was empty;the parent runner's
separate Unexpected/timeout exception is not a child assertion failure.
Last savedPASS is ExactFieldAndTraceBounds. The next block constructs4096tick
events,then calls New-SyntheticCaptureRecord and the real writer/final reader;
ExactEventCountBound never printed. No internal stage timing was captured,so
this cannot distinguish fixture construction,contract validation,encoding,
decoding or final verification time. Do not claim a root cause,environment-only
issue,semantic defect,or verified fix. No code/test change after anyGREEN launch.

IndependentWireLiteral and all20-field/four-lifecycle/Int64/Unicode/raw-byte/
ownership cases completed on bothhosts. PS7 additionally completed all73
boundary checks including CodeProperty nonexecution,4096events/exact byte caps,
malformed primitives with valid recomputed file hashes,stopped/partial markers,
existing-target/reparse/hardlink/locked-file refusal and cross-format refusal.
PS5 reached callback/shadow and large-byte bounds but NOT the4096event verdict
or later disk-corruption/path cases. Do not extrapolate PS7 coverage toPS5.
Old reader cross-format check is path-scope refusal only,not exhaustive parser
compatibility. The planned old-format regressions did not run.

Source SHA256 at allGREEN launches:
`508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C`.
Tests frozen beforeRED:
`185FA2569A3364CD8DC4E894490794A5B6C6799F84B323F6BB6D66B94468335F`.
Six-entry invocation index:
`ADF5012858A53A20B371BD9D187054C5DB003D4F070EF743A9171E4AE338827A`.
Ignored ledger/index/stdout/stderr under
`.superpowers/sdd/2026-10-17-synthetic-capture-store` retained. Text receipts
are synthetic host output,not real raw capture/provenance artifacts. Tracked
tests/source are portable and do not depend on ignored receipts/private caches.

## Rulings and limits

PowerShell-native ledger/brief/bookkeeping rather than Bash helper scripts;
finite approved12 hosts instead of additional baseline/task-done/full-project
runs or installs. Preserve ignored artifacts instead of deleting skill scratch.
No automatic review fix pass or extra review/host. Cost:no whole-project
certificate,no silently replenished allowance or removed failure history.

Stop ruling:no instrumentation/new timing host,source change,retry or early
review under this packet. Cost:PS5 boundary and new integration/review remain
unaccepted;timeout root cause unresolved. No deferred minor findings exist
because the independent review did not start. Older consumer/review results
remain historical/separate,not current certification of this new module.

Next proposal,not execution authority:separately bounded stage-timing diagnosis
of the4096-event path to distinguish setup/contract/codec/final read costs.
Preserve slot6 as-is,do not replay it or transfer frozen7..12 slots to diagnosis;
do not raise limits,drop the case or relax checks before cause is established.

No full live collector/watchdog/OS stop/cleanup,origin/clock/command provenance,
signature/encryption,adversarial concurrent object/path mutation,transactional
atomicity/fsync/power-loss guarantee,security/vendor clearance,Stage B/I2/I3/full
C12/production/main acceptance. Consistent synthetic incomplete/timed-out/refused
records may archive;invalid declared clock-order records remain refused.
All authority flagsfalse;archive completion is not lifecycle success.
