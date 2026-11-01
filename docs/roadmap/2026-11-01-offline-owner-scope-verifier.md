# Owner scope verifier — actual execution

Approved native execution of the [plan](../superpowers/plans/2026-11-01-offline-owner-scope-verifier.md)
and [design](../superpowers/specs/2026-11-01-offline-owner-scope-verifier-design.md).
This is a development checkpoint, not I1/I2 or product acceptance. Historical
failed attempts, five historical stops and archives remain unchanged.

## Latest — runtime stopped; NOT_ACCEPTED

After checkpoint8c28bbf80680eb60fdb2e817d9be4fd7884b2203 was pushed and verified,
the unique writer-scope-green host ran once. Its stdout has exactly26PASS lines
and the journal records tests-complete. It then failed while projecting evidence,
before creating writer-scope-green.evidence.json. No valid writer acceptance.

Actual childexit2/parentexit1; process17124 owned exit and complete streams,
no timeout/termination/rescue;2236post-pins intact. Stdout hash equals the original
expected stdout hash006DBE68586DA79FF098382619B4C943777C539E5E26E95EDAB37771D6047600;
stderr contains `The property 'PositionMessage' cannot be found on this object.`
The parent refuses exact stdout/stderr and creates the newest verifier stop.

The actual journal retains PropertyNotFoundStrict/PropertyNotFoundException at
capture.ps1:32, through recursive Capture-Value into Capture-Evidence:79 and the
publication boundary:127. Source-only tracing shows synthetic ContractOnly
ErrorRecords in SharedExceptionDifferentRecordsKept are retained in OwnerControls
details and later traversed. Their missing invocation metadata is a plausible
origin; the precise failed record was not retained, so do not claim an actual
object-level reconstruction. No instrumentation, reproduction, source/test edit,
retry, re-freeze or runtime reviewer was performed after this failure. The
systematic-debugging skill guided readonly tracing only under the approved stop.

New sixth stopSHA4B06ED585A7F63DC38F084D0718E653F17266E2065175E58CAD67EC5AEF6FD14.
WriterFinalSHA384E124FC861529BF1B07AC18793FFBE2F177A196F56507E4E7DFF753EF081F3.
JournalSHA7264F64D7CFFF99E0A44A63EB566C2193FC111B6D9D06D8EB6F929CDC7789362.
All27historical bound files, five current source identities,2236pins and124snapshots
verified unchanged; no later-attempt artifacts. Cumulative2hosts/2compiles/
0targets/0reviews,6auxiliary computations consumed. Remaining4hosts/7targets/
0compiles/conditional1review FROZEN, not released or refunded. Contract source
change/freeze and all later stages have not happened.

Final portable-owner-scope-verifier.json preserves369actual P/Q files,
100directories,20zero-byte files; LF container/decoded originals verified equal.
ArchiveSHA16839616C3342A3C84D4CF24E4FC63B5038E5A2DC5D344299E022EE729AD6E34.
Old archives and the earlier source-freeze checkpoint archive are unchanged.
Next recommendation: readonly review of error-record projection and the whole
evidence path, then a newly approved bounded design/attempt budget before edits
or execution. Do not append ad-hoc retries or merge main.

## Actual results

- Regression frozen before RED at572cfff2749bb4f7a8d373d7ee5d085bfc1868cd;
  SHA BD3DF80E75C62B7924BBC4AB399493AE9AD1F102117B3061D4B7DE0775AFD926.
- Old checker RED:0PASS/1FAIL/7NotRun/exit1, exact original RuntimeException
  `Owner scope regression: actual body unchanged Capture-Value`. Both AST kinds
  and actual comparison operands retained. No old failure was regraded.
- New auxiliary frozen before GREEN at3e6a3259ee8de1bd04a40d2de2e25def93563fc9;
  SHA52B237AF4BF7DB0AD0D45919B41860C8F39831E2B26E69BC840BE341E096FF4E.
- Verifier GREEN:8PASS/0FAIL/0NotRun/exit0. Actual four imported definitions;
  literal body/inline-parameter/node/name negatives; retained actual first error,
  diagnostic projection fault, prior completed groups and later-body suppression.
- Full scope GREEN:9PASS/0FAIL/0NotRun/exit0. Five actual source hashes unchanged
  across the computation. DriverSHA B93A7B03C2EEB5DA55DA720FA4D2C94AD5F222B025B9ABB7B16D21BC71F03527.
- Unique Prepare writer-scope-rebind:exit0;2236pins/31new snapshots,124/155 total.
  ManifestSHA7F09A142E228FF16CD7228287DEBBB5EE60DEC9A91FBF1EF691F148C0A911587.
  Original five-source window, other26runtime sources and all prior snapshots retained.

All six cumulative auxiliary computations are consumed; no repeats or new probes.
These three computations invoked no profile body, record writer, target, child
host, compilation or reviewer. Runtime totals remain1host/1compile/0targets/0reviews.
Remaining approved chain:5hosts/7targets/1compile/conditional1runtime review;
contract source epoch may use the final31snapshots only after actual writer GREEN.
At the earlier source-freeze checkpoint, no new runtime stage had executed.
The later stopped writer result above supersedes that pending status.

## Portable source-freeze checkpoint

`docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/portable-owner-scope-verifier-freeze.json`
contains338actual P/Q files,83directories,10zero-byte files. Decoded bytes, lengths,
hashes and directories verified. ArchiveSHA3735B585AF79FD22EE9B81DF2BF07865AB09C65BDCF0054CCF1FEF91B3BD08DA.
The LF container embeds original bytes without normalization. Sources/tests are
ordinary Git files; raw ignored session files are not staged separately.

Recover only from codex/c12-b01-task9-coordinator, not main. Read the recovery
runbook before restoring into a clean session location; old absolute manifests
are historical evidence, not portable permission to replay an attempt. Keep all
five old stops; the newest owner-scope-verifier.stop.json, if created, always
blocks execution. That checkpoint was saved/pushed before the single original
writer-scope-green host/fixture compile; its later failure is recorded above.
