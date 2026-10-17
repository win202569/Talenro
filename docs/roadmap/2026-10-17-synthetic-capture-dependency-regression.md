# Synthetic-store dependency regressions — six approved runs complete

## Approved execution packet

After cb5cb1ed, the user approved NEW six sequential children,120000ms each:
old output-evidence-store on PS7/PS5.1 with administrator tokens, then simulation
and contract on PS7/PS5.1 with normal-user tokens. Exact targets81/106/110,
stop on abnormal, no retry, source/test/security changes or frozen-slot reuse.
All output retained; only empty stderr or fully recognized structured progress
is allowed under the previously approved stream policy.

The new runner parsed before execution. Twelve source/test hashes and all
three ordered label inventories were frozen before the first child. Inventories
came from previously saved matching successful engine pairs, checked against
current tests; repeated labels remain repeated. No current output was used
to lower or alter expected results. Parent and child check the expected user
and exact admin/non-admin role before tests. No RunAs, ACL or security changes.
Only calls1/2 used the approved elevated tool path; calls3..6 used normal context.

## Results

|New slot|Suite|Engine|Admin|Child exit|Elapsed ms|PASS|Stderr chars|Progress objects|Result|
|---|---|---|---|---:|---:|---:|---:|---:|---|
|1|old store|7.6.5 Core|true|0|8897|81|0|0|complete|
|2|old store|5.1.26100.9711 Desktop|true|0|2836|81|584|2|complete|
|3|simulation|7.6.5 Core|false|0|2625|106|0|0|complete|
|4|simulation|5.1.26100.9711 Desktop|false|0|2089|106|584|2|complete|
|5|contract|7.6.5 Core|false|0|34628|110|0|0|complete|
|6|contract|5.1.26100.9711 Desktop|false|0|55175|110|584|2|complete|

All engines x64. No timeout, retry, cancellation or source drift. Every ordered
PASS sequence, first API label, host/context and exact suite LIMIT matched.
Store ends with WriterBoundaryCoverageComplete; simulation's final assertion
is PostReturnMutationCannotChangeSnapshots; contract's final two labels are
SplitProgressRawBytesPreserved. These last labels are not invented new terminal
assertions. Exact inventories and full output determine completion.

PS5 stderr has two Completed progress records per child; no other stream object
was present in the parsed documents. Raw decoded text remains retained, including
damaged localized activity text; no wording/emitting-module reconstruction.
XML classification prohibits DTD/external resolution and does not deserialize
objects or evaluate scripts. Malformed/unknown/error output would stop the run.
No blanket stderr ignore or progress suppression. All ExpectedResult=true.

Totals:297 PASS occurrences per engine,594 across six children. Counts include
API/completion checks and repeated labels; not594 unique requirements or a
percentage of overall coverage. Old store's hardlink/junction/locking cases
executed under administrator context; this does not establish normal-user
success or prove a universal privilege requirement.

## Disposition and remaining gates

The dependency-regression component of the [review's Important1 finding](2026-10-17-synthetic-capture-readonly-review.md)
now has fresh evidence for these unchanged suites and contexts. This report
does not rewrite the review or declare its entire acceptance finding resolved.

[Supplemental tests](2026-10-17-synthetic-capture-supplemental.md) already provide
direct evidence for the targeted Important2/3 gaps. They have not received a
second independent review. Recommend next a separately approved one-seat
readonly closing review of that added test file and final evidence boundary,
without tests, fixes or old-review/gate promotion. None dispatched here.

Original PS5 monolithic boundary timeout, normal-context reader fixture failure,
and administrator PS5 reader strict-stderr failure retain their failed historical
dispositions. Their exact causal explanations remain open. Old implementation
and partition acceptance-review prerequisites remain unmet; this new mixed-
context packet does not silently resume them. PS5 missing-root dynamic evidence,
file-level reparse, independent populated-package oracle/max-package coverage
and other recorded limitations remain. No all/boundary rerun or full-project
test occurred. No main merge, live capture, Stage B/I2/I3/full C12, provenance,
power-loss or vendor/security clearance follows.

NEW6/6 allowance spent. Original6/12 and partition7/10 remain spent with their
remaining slots frozen; reader2/2, format2/2, supplemental4/4 and previous
diagnostic/review allowances stay spent/closed. No automatic extra execution.

## Integrity and fresh-computer recovery

Readonly postrun checks matched raw stdout/index labels to frozen inventories,
exit/timeouts/context flags and stderr lengths/types; all twelve pins below
remain unchanged. New index SHA256:
`8867CB8CB41AA6450468F3754A784B663DBF2DDE1D698C6C0FB12AD31D50CFAD`.
Source-manifest SHA256:
`648BC72F9DB1722027DBBE8D935F17F91D6B4F5139A77F00AF6E994882F0290F`.
Files below are under `docs/diagnostics/isolated-fuzz`.

|File|SHA256|
|---|---|
|output-evidence-store.tests.ps1|80DCB954361F6ACE71FC9649D25A35748ACB3D578491A30C3D3F9862ACBDC147|
|output-capture-simulation.tests.ps1|B6E7E342A76FB6BD2AEE6641BAF65305846071A2562F0CD97AC4ECD1806CE7FE|
|capture-adapter-contract.tests.ps1|5A6975AAF9998D62304BA3DDB7483A6CE1A1382FBE9D05EBDA209BD7AE71566F|
|output-evidence-store.ps1|7710A1149B42CFD0C2505988CEE6748F399C3FB5E8C34D795B35FBCB8558BB81|
|output-capture-simulation.ps1|9DA8A74B984A3111F00B54AA160490DF327BE17A5A07A044B510946EB26A2A13|
|capture-adapter-contract.ps1|116FCAAA6506ADC6969D718AC61F8C79908BDED48ADA1204C2E8A60F33A35BB2|
|output-evidence-package.ps1|857543A9DDA9B9C849B2507C6EE419C1F73AA0D0C8599CD9D88738841D021198|
|output-channel-classifier.ps1|F323C9C46190DE06F25BFA55A2B7983707B1FF5877548F9703870F9BB771A0BF|
|isolation.ps1|1FD75426BBB1015664301108B1A2F5E3426874CB672EDFEE3FE7767BD1356E2D|
|synthetic-capture-store.ps1|508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C|
|synthetic-capture-store.tests.ps1|406C7C661A181988C7F0AE90E577D1F948148C2B19BE5CD411F92D23E7FECBFF|
|synthetic-capture-store.supplemental.tests.ps1|243B4673BF22614718D61FF92BB421F7EE09D993691D0803BB88EC8D11C4B87A|

Original/partition/reader/format/supplemental indexes also rechecked unchanged,
respectively ADF50128/B61DD587/080455EE/5FA944DF/33D8DFC4 (full hashes in their
linked reports). No source or test file changed this turn.

Raw runner, pins, inventories, invocation index and stdout/stderr remain
ignored/local under
`.superpowers/sdd/2026-10-17-synthetic-capture-dependency-regression`.
Old-store fixtures and all earlier artifacts remain retained. Git preserves
this portable result and recovery/status notes on the development branch,
not raw receipts or runtime state. Clone from the latest development checkpoint,
not automatically main; cloning grants no permission to rerun spent slots.
Overall acceptance remains open.
