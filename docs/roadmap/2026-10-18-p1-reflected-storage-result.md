# Reflected storage construction — 10/10; revised harness prepared only

Date: 2026-10-18. Base f29c7c72d6d011ca9e53dc3c91ecbf9842f5e5b9.
Branch: codex/c12-b01-task9-coordinator.

## Verified outcome

User continued: verify wrapper-retaining field assignment, then prepare a corrected
fixture. One ordinary-user PS7 synthetic-only host completed all5 storage kinds /
10 bidirectional transitions in777ms, exit0, no timeout, stderr0. Owned exit and
streams complete, all647 frozen file pins intact, ControllerError null.
CandidateImports/ProductCalls0; FormalAcceptance false. No observer scan, actual
ownership matrix, failed-batch rerun or old-slot reuse. No further child launched.

Fixed scope was10,000ms execution +2,000ms owned exit +2,000ms stream drain.
The allowance was consumed before launch. The prepared tests were written before
the test-only helper implementation. Earlier failed storage fixture plus the
14-record trace are preserved negative witnesses; neither was rerun.
No fresh full RED/GREEN matrix, product mutation, independent review, project-wide
suite or performance comparison was run under this narrow construction allowance.

S-SetStored invokes the public PSNoteProperty.Value setter through the same
MethodBase.Invoke trampoline, with fresh object[] argument arrays at both layers.
No private member, security setting or product operation is involved.

Each transition independently requires the literal expected native stored type
and a real change from the previous type, while retaining descriptor identity,
base identity/type/value, no metadata and zero pipeline output.

| Storage kind | Transition1 | Transition2 |
| --- | --- | --- |
| Int32 | Int32 to PSObject | PSObject to Int32 |
| Int64 | Int64 to PSObject | PSObject to Int64 |
| Boolean | Boolean to PSObject | PSObject to Boolean |
| String | String to PSObject | PSObject to String |
| WrappedString | PSObject to String | String to PSObject |

All10 checks passed. PSObject means System.Management.Automation.PSObject.
This validates this test-utility construction on the stated engine, not complete
observer equivalence or a claim about all PowerShell binding/runtime contexts.

## Independent prepared harness copy

New leaf: docs/diagnostics/isolated-fuzz/inline-observer-storage-v2.
Its offline.tests.ps1 is PREPARED ONLY, NOT executed. No launch controller or
new full-batch execution allowance was created.

It preserves all45 case definitions and literal expected outcomes byte-for-byte
after newline normalization (static inventory comparison passed). Parser errors0.
Original/optimized observers remain unchanged and are referenced as AST data.
Only the previously allowlisted ten functions plus the one verified setter
definition can enter a fresh observer module.

Fixture changes:
- Use reflected set for storage mutations, fresh-base assignments and intentional
  sharing, avoiding silent PSObject erasure through ordinary property assignment.
- Explicitly initialize WrappedString owner slots as actually stored PSObject.
- Independently assert both captured primary leaves have their specified initial
  native storage type before scanning.
- Keep initial storage consistent for an optional same-owner additional field.
- Parse the setter utility as data, extracting exactly its one function definition.

No existing guard, type/reference/value comparison, outcome literal, scan, case
count or negative metadata check was removed or relaxed. The same source-mask
comparison outside V-Others remains. Original failed harness/controllers, candidate,
optimized observer source, historical reports and receipts stay unchanged.

The new harness has NOT established any of its45paired runtime outcomes. In
particular, sharing, held/reread views, base replacement and other fixture branches
still require complete validation. Prior historical matrix remains only3/45paired
cases completed. Do not infer equivalence or a speedup from10setter transitions.

## Evidence and input hashes

Ignored local packet: .superpowers/sdd/2026-10-18-p1-reflected-storage-fixture.
StartUTC2026-10-18T13:28:45.8977972Z, owned childPID20532.
637 prior pins +4 prior receipt anchors +6 new fixed files =647.
Nine fixed controller SHA literals were independently checked before launch.
All3new validation scripts parsed with zero errors.

| File | SHA256 |
| --- | --- |
| frozen.json | 301B63F26DD68DDA5AFD3266EC0DBE05FEF8DA9BF4F49DEF34A3694ECDFA5323 |
| invocation.json | B269EA0963B925996D15350E4099777F2A616975ABEBB5CDB508F966D879BAF7 |
| stdout.txt | B6DC1ED6ED076C40CB1FFB54CFB8B22E0327E7AA8E53BD6C8A1C497C8F76FD1B |
| stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| capture.ps1 | 7889F335A2D596031DA487BBCBBD9833FE27D7C7FAED830FBC6C9D39FBE73CC7 |
| setter.tests.ps1 | 7D7093AA44A01B2AD32190972FE00F51189AEECC41231E7B5681767EC876F55C |
| set-stored.ps1 | 91C107BF6D5A08D265F0837CB161D75C5C08FC28DB962745EC13DABEBE1A5B01 |
| README.md | E1E661596EA941BDCCF46BA9C37680720B049A7B877FAC54D8ADD4C31E8A62A6 |
| source-pins.json | 73B66717A58CEB3E5B8BEC0A7E3DC13ABC14B27E8AC862AE2E5538127292C05D |
| .gitattributes | 0CD209979B0AD7DB7DDC9CB658156B0A9C18859999F71E541861F3AB58E7AB11 |

Prepared v2 harness SHA256:
D6531D884313F22DBBC3A03F5D6909A7D8E58C51C0B216FF011663EE68DCFB33.
V2 README931DDCAAC7C508DE466523BD32C3624F9E95429F7754F52D146035B09F96503E;
v2 source-pins83670485E1BE0EA1EA03B2C80E6C5212CE9C3450BF6D72D303FF455F79E01059.
Both leaves' attributes are identical. Raw evidence remains ignored and untouched;
portable scripts/table/result allow recovery without claiming raw files were cloned.

## Next bounded proposal

Prepare a new one-shot controller for v2's unchanged45case inventory: one ordinary
PS7 offline child,30s execution +2s owned exit +2s stream drain. Inherit/pin all
old records, validation utility and prepared v2 sources; consume before launch;
stop on any setup/equivalence mismatch, with no post-launch repair/retry.
No candidate import/call, actual ownership test, timeout increase, old-slot reuse
or removal of storage-type coverage. This is the next proposal, not executed here.

Only after full offline equivalence succeeds may a separately agreed candidate
timing diagnostic be considered. No main merge, I1 closure, runtime/security
acceptance or C12 completion. Fresh-PC recovery starts on this development branch;
local receipts/engine cache and execution authority do not follow from cloning.
