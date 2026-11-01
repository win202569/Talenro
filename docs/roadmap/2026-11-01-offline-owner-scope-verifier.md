# Owner scope verifier — actual execution

Approved native execution of the [plan](../superpowers/plans/2026-11-01-offline-owner-scope-verifier.md)
and [design](../superpowers/specs/2026-11-01-offline-owner-scope-verifier-design.md).
This is a development checkpoint, not I1/I2 or product acceptance. Historical
failed attempts, five historical stops and archives remain unchanged.

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
No runtime stage after this source freeze has executed yet.

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
blocks execution. Next: save/push this checkpoint, then the one original
writer-scope-green host/fixture compile with unchanged safety and timing gates.
