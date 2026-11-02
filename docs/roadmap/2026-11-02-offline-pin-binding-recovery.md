# Pin-binding recovery — actual execution

Approved plan after d3e129ac; native, no extra agents/review or writer replay.
Whole product/I1/I2 NOT_ACCEPTED pending actual original runtime chain.

## Auxiliary results

-11 pin RED:0PASS/1FAIL/6NotRun/exit1, exact SHA256 PropertyNotFoundStrict.
-12 pin GREEN:7PASS/0FAIL/0NotRun/exit0, exact rejection checks retained;
 rename-only source change verified before execution, pin-array conflict fixed.
-13 descriptor v2 RED:0PASS/1FAIL/19NotRun/exit1, exact missing-descriptor error;
 no setup failure. Original failed computation10 remains unchanged.
-14 descriptor v2 GREEN:20PASS/0FAIL/0NotRun/exit0, exact fixed AST descriptor,
 actual saved unit stack, rejection controls, historical epoch and unchanged
 process guards all pass. Original frozen20group test retained; v2 only rebinds
 namespaces/epoch/facts/stop selectors. Pre/post hashes equal, no recording error.

Candidate81473dfb was committed/pushed before GREEN14. Actual current source SHA:
source-model1C7897E486D43601D4D0E963A566737C97E33FC8C88BF7E4ECF83348A5A61418;
observation-contractB68A9D8C9674692C8B75264CDB687E92063911254F64F5D69124E03265E3CB12;
driverCB3E0B17EAF39BE39F4E2403ED9F8C48AE028A3B372F519681BB464C12FADCC9.
No source/test edits after final auxiliary GREEN.

## Source freeze

Unique Prepare pin-binding-rebind exit0:2276pins,31new snapshots,total186.
repair-green.frozen.json SHA
A9949DD46C7F928CFE7DB33F092AF0497E6F89556E8F359A8A55CB4484BB1556.
Exactly source-model/observation-contract/driver changed; other28sources frozen.
Original writer2251pin epoch/maps and eight historical stops preserved.

portable-pin-binding-freeze.json:510actual P/Q files/133directories/31zero-byte,
decoded original bytes/lengths/hashes/directories verified; SHA
0FF966E80721AFBAE838E3D67E7A62ACE3F817C2B2F271AAA12DCCC17FF78CA2.
An additional descriptor-preparation archive preserves v2 inputs before RED13;
it adds no computation, host, compilation or source-snapshot allowance.

At this checkpoint consumed14auxiliary/3hosts/3compiles/186snapshots/0targets/
0reviews. Remaining4hosts/7targets/0compiles/conditional1review, gated in sequence.
Next original contract-green host; no replay of successful writer or auxiliaries.

## Actual contract host — stopped, not accepted

After pushed freeze0d1cb71b, contract-green ran once, PID5560. Childexit2/parentexit1;
OwnedExit/StreamsComplete/PostPinsIntact true; no timeout, forced termination,
overflow or rescue. Parent Exactfalse/EvidenceValidfalse/EvidenceStateUnknown;
all five capability flagsfalse. The parent final remains authoritative and its
EvidenceSHA256null is not backfilled from a later file read.

Actual stdout: PASS ContractAPIAvailable, FAIL BaselineCore. Stderr: ASSERT exact
full-contract result BaselineCore. Saved incomplete evidence has1PASS/1FAIL/
50NotRun and retains the underlying mismatch in SchemaSuite. Its independently
read SHA5993F5D9A5C93BDE307C4EDD14E365D04774C1EF718895134C0636B95E7C4F40
establishes file identity only, not successful parent evidence validation.
Full original tool output, final, streams, recorder journal and ninth stop preserved.
StopSHA D5D3CE0845438682B770E2FA747607B5052A38B39A1A861A0021CEAB707066AD.

Readonly diagnosis: actual inner stack reaches OC-SourceError line33, which checks
fault function/operation/line/column. Historical bad-JSON record points to
scripts/private/offline-diagnostics/sinks.ps1:36 char50. Raw AST parent chain is
CommandAst36:50 → PipelineAst36:16 → AssignmentStatementAst36:9; existing
Get-ODFailureExtent promotes to the assignment. This is a concrete suspected
command-versus-assignment extent mismatch, not an extra executed reproduction
or verified fix. The saved mismatch does not identify a row directly, so do not
claim this static comparison alone proves the complete cause of BaselineCore.

No source/test changes or retry after stop. Core/lifecycle/continuation intents
remain absent; no conditional reviewer ran. Consumed14auxiliary/4hosts/3compiles/
186snapshots/0targets/0reviews. Remaining3hosts/7targets/0compiles/0snapshots/
conditional1review frozen. Auxiliary budget exhausted. A recovery revision must
explicitly account for any new source freeze/host/test; do not reuse spent slots,
weaken location checks or alter historical evidence. Product remains NOT_ACCEPTED.

## Final portable preservation

portable-pin-binding.json contains518actual P/Q files/134directories/31zero-byte
files. Decoded bytes, lengths, hashes and directories verified; SHA
8FF8BA63EFA96FFA476CFB2AA165312459DF09979610EB2829F64FC75B638029.
All earlier archives remain unchanged. Ledger rulings were limited to honoring
the approved PowerShell-only finite-run workflow instead of generic helpers,
and adding a separate descriptor-preparation archive to avoid overwriting earlier
ones. No additional runtime/test allowance was consumed by either decision.
