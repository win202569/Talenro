# Reader receipt assessment — readonly, no acceptance promotion

## Authorized work and finding

After d5f5d0ea, the user approved the recommended readonly assessment of the
saved administrator reader-paths receipts. No test child or diagnostic child
was launched. No original receipt, runner, test, product, fixture, permission
or security setting was changed. This document is a separate derived finding,
not a replacement for the [original packet disposition](2026-10-17-synthetic-capture-reader-admin.md).

Both saved children exited 0 without timeout and contain the exact same ordered
45 PASS labels: 43 content assertions plus API/completion sentinels. The labels
were checked against the pre-existing reader-paths inventory, not just counted.
Both administrator guards and context output match, the LIMIT marker is present,
and stdout contains no FAIL label or ASSERTION FAILED text.

PS7 stderr is empty. PS5 stderr contains 584 decoded characters, consisting of
the CLIXML header and one parsed Objs document with exactly two direct Obj
children. Both have S=progress, SourceId values 1 and 2, RefId values 0 and 1,
and progress type Completed. There are no error, warning or other stream objects
in this saved document. XML was inspected structurally with DTD processing
prohibited and external resolution disabled; no PowerShell object deserialization
or script evaluation was used. Localized activity text is already damaged by
decoding, so its exact wording and emitting module are not established.

Derived assessment: the saved stderr is serialized progress, not an observed
archive assertion error. The prior packet stopped because its parent requires
stderr.Length=0 even when the child exits 0 and all assertions complete. This
explains the recorded parent/child exit discrepancy; it does not prove the
exact origin of the progress or establish a general policy for ignoring stderr.

## Disposition and next work

- Reader assertion evidence is complete on both administrator-context engines.
- Original ExpectedResult remains true for PS7 and false for PS5; the packet
  remains stopped/unaccepted under its original empty-stderr gate.
- This is not a rerun, repair, independent code review or whole-module acceptance.
- Both new child slots remain spent. Old partition 8..10 and original 7..12
  stay frozen; their review prerequisites are unchanged.
- Remaining work includes format verification, dependent regressions and the
  pending reviews. No main merge, Stage B/I2/I3/full C12 or security clearance.

No reader rerun is recommended solely to obtain empty stderr. Before any future
verification launch, specify its stream policy explicitly: retain all raw output,
distinguish structured progress from errors, and reject malformed/unrecognized
output rather than blanket-ignore stderr. A narrowly scoped format-only packet
on PS7/PS5.1 is the next proposed behavioral verification, not executed or
authorized by this readonly assessment. No retrospective gate rewrite follows.

## Integrity and fresh-computer recovery

All five receipt files were hashed before and after inspection with identical
results. SHA256 values under the ignored local directory
`.superpowers/sdd/2026-10-17-synthetic-capture-reader-admin`:

|File|SHA256|
|---|---|
|invocations.json|080455EE56A74498B09099061C53FC0CF46801C4ED614BEA37774D72629D8211|
|run-1.stdout.txt|998E1ECF1606DEA0091C274444A02A2FFA3718AB92C408FC0E33B32A070327A7|
|run-1.stderr.txt|E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|
|run-2.stdout.txt|EC131D87DA1D475F66410F83D71EB05AB5F9752AA1DDB2A69DBB95541A7FB6DB|
|run-2.stderr.txt|2935F8F6DEA46F1BBBD7BFFB2DD5335DA5A36DB3E61E9B095B74D063775EA01A|

Unchanged product SHA256:
`508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C`.
Unchanged test SHA256:
`406C7C661A181988C7F0AE90E577D1F948148C2B19BE5CD411F92D23E7FECBFF`.

Git preserves this finding and the prior report, not ignored raw receipts or
fixture state. A new computer can recover the result and remaining boundaries
from the development branch; cloning is not permission to rerun verification.
