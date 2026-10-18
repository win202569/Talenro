# P2-B inbox tree-copy repair — unfinished checkpoint

Date: 2026-10-18. Existing development branch: `codex/c12-b01-task9-coordinator`.
Base commit: `131fee8584a1b308d28152abff3151da3ab41554`.
User approved the displayed Native implementation and NEW5+1 allowance.
Status: **UNFINISHED — preflight identity stop after one successful 44-group full**.
No main merge or formal/runtime acceptance.

## Implemented scope

New portable leaf:
[tree-copy repair](../diagnostics/isolated-fuzz/p2b-inbox-tree-copy-repair/README.md).
Only `Copy-IBTree` and the new private `Get-IBTreeNode([object]$Value)`
differ from baseline. Loader/registry prefix and every byte from
`function New-IBResult` onward match the original. Classifier keeps the
property owner separate from the native value, follows the fixed wrapper field
at most eight levels, dispatches exact PSCustomObject/object[] nodes, allocates
new arrays, and leaves scalar copying/rejection with the unchanged codec.
No new public inputs, exports, setter or state restoration interface.

Good product SHA256:
`47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC`.
Original product baseline:
`43067AD0939CB56426568DDF435695383E6C21ABFD223B5A369BC6AE6F2451C8`.
Original32 test SHA remains
`904F20D4D60540D7A312F5ACF14B0E1E01D263C7A905B18F578FB0846C7267CE`;
helpers, fixtures and descriptor source also match their original bytes.
Codec `9A532E17...` is unchanged.

## Measured results

| New slot | Actual result |
| --- | --- |
| red | PID45540,1081ms,exit1; exact3lines/P2BTreeFirstSnapshot assertion |
| full | PID35788,44734ms,exit0; exact49lines/44PASS(new12+unchanged32) |
| capacity | Preflight rejected source SHA; no target/preconsumption |
| sequence | Not run; frozen |
| final | Not run; frozen |

Both launched hosts were ordinary PS7.6.5 Core x64, no timeout, empty stderr,
owned exit and concurrent drains completed, all1326pins intact.
RED independently observed its controlled test object as raw PSObject,
base PSCustomObject with Label/Values properties. The single public snapshot
returned internal-error, Accepted=false, null payloads and false authority flags.
This is not a direct observation of the vanished historical diagnostic object.

The repaired full passed all twelve focused groups, including first snapshot,
object/array ownership, empty arrays, nested fields, supported and unsupported
leaves, returned mutations, separate snapshots/inboxes and stopped snapshots.
The unchanged original32 all passed in the same host. Its trace records one
managed fixture compilation, Observable=true, Counters=0, four synthetic
sequence cases and eight retained cold descriptor decorations, all counters0.
Only FirstSnapshot had a RED run; the other eleven focused groups first ran
in this successful full. No claim that each separately had RED coverage.

## Why execution stopped

The agent's capacity mutation patch mistakenly changed the site's indentation
from twelve spaces to four in addition to Count -ge4 -> -ge5.
The controller rejected this source because it did not match the already sealed
capacity-mutant SHA. The rejection happened before the invocation row was
preconsumed and before a target process was launched. It is a mutation
application error, **not a product test failure**, and not evidence that the
capacity mutation was detected by its test.

An initial comparison restoration left the indentation mismatch (SHA7DE8460B...).
The immediate required restoration then reinstated the original twelve spaces;
the product is now byte-for-byte the sealed good SHA47BABFED... again.
Only restoration was completed; no retry or further product fix was made.
The successful full's verdict and all original receipts remain unchanged.

NEW allowance:2/5 target hosts spent +3FROZEN. Conditional reviewer:0/1,
ineligible because all five exact outcomes were not obtained.
Old original2/5SPENT+3FROZEN and independent diagnostic1/1SPENT are unchanged.
No fresh reviewer was dispatched. No additional probe, import, compilation,
repository/codec/Go/Docker test, security query or settings change was made.

## Evidence and handoff

[Portable summary](../diagnostics/isolated-fuzz/p2b-inbox-tree-copy-repair/portable-summary.json)
preserves the two receipts, complete three traces, seals, eleven input hashes and
preflight-stop account. Raw packet remains ignored and untouched:
`.superpowers/sdd/2026-10-18-p2b-inbox-tree-copy-repair/`.
Frozen manifest SHA:
`DB3FFAFC5D8A9E9501B3C12FFBFBF3735CB5B52D8414CCB020EC71ED52E58418`.
Product seal SHA:
`DB47535AF5D567C23D94DAC03E491E9E6D3A9C3BC4C92708AC0B44416AA305C8`.
1326 mappings =1289 inherited unchanged +37 additions
(4diagnostic packet +5diagnostic portable +11tick-correction portable
+3reports/spec +10new immutable inputs +2engine/SDK +2snapshots).
Full mapping inventory was displayed before freeze; final readonly hash check
confirmed every mapping. New summary/report/status files are handoff outputs,
not retroactively frozen target inputs.

Next recommendation, requiring explicit authorization: retain all current
frozen tests/oracles/controller, old raw records and exact good product;
apply each already sealed mutant with indentation preserved and verify its SHA
before calling the controller. Release only the three unlaunched slots in order
capacity/sequence/final and the existing conditional readonly reviewer after
all remaining outcomes are exact. Do not rerun RED or the completed full.
This stop does not automatically grant that continuation or a new budget.
A fresh computer needs the original ignored packet/dependencies reconciled;
cloning portable files does not reconstruct historical raw identity or rights.

The execution/verification skills guided independent RED, restricted source
changes and evidence-based claims. The approved finite stop rule overrides
generic retry/fix-pass or additional-suite instructions.
