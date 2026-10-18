# P2-B PayloadLength — expression defect diagnosed, NOT repaired

2026-10-18; base139f1d9ce3e1dc631d65c2c4d971aff58cfaed45.
User answered “按你的建议继续” to the proposal for one minimal diagnostic.
Scope: one standalone ordinary pinned PS7.6.5 expression-only host; no product
or test import, API call, full validation, mutation, repair or reviewer.
This separate diagnostic consumed1/1. Original route remains2/5 consumed,
3 FROZEN; conditional review0/1 INELIGIBLE. No slot reuse or acceptance.

## Hypothesis and observation

Frozen test lines195–201 obtain Hello payload length16, then evaluate
`@($len-1,$len+1,113)` before entering the loop that calls BC-UInt/decoder.
Hypothesis: comma/operator precedence creates array operands rather than
three independent integers. The diagnostic copies that expression exactly,
uses len16, records its AST and exception, and evaluates a parenthesized
control. It imports neither product nor tests and does not learn an oracle.

```text
HOST: 7.6.5 admin=False
AST: Plus left=[$len-1,$len] right=[1,113]
AST: Minus left=[$len] right=[1,$len]
ORIGINAL: MethodNotFound
MESSAGE: Method invocation failed because [System.Object[]] does not contain a method named 'op_Subtraction'.
LOCATION: 17
BODY_ENTRIES: 0
PARENTHESIZED: 15,17,113
LIMIT: expression only; no product import or acceptance
```

This confirms a test expression defect: the subtraction receives Object[],
and the loop body is never entered. A candidate minimal test-only correction
is `@(($len-1),($len+1),113)`. That expression was evaluated only as a
diagnostic control; the frozen test has NOT been changed. The original full
group-only failure remains unaltered; its missing exception is not retroactively
filled in. This result establishes a reproducible blocker in that test path,
not that the codec is defect-free or that groups14–32 pass.

## Evidence

Portable scripts: docs/diagnostics/isolated-fuzz/p2b-payloadlength-diagnostic/.
Controller reserves a new ignored packet before launch, refuses replay,
checks ordinary parent and pinned engine/module/868mapped inputs/original
index, captures separate raw streams concurrently and owns only its child.
Bounds:10000ms execution+2000ms owned exit+2000ms concurrent drain.
Actual PID28200,exit0,no timeout,owned exit=true,streams complete,stderr empty.
The exit0 describes completed diagnosis, not validation of the product.

| Artifact | SHA256 |
| --- | --- |
| expression.ps1 | 3E640DC9B0454A9F5931074B1DDE68D27510B173DC16BCB306F0B6045AD375DE |
| capture.ps1 | 29E7F87A28720CBE87363C22991ECF664EE0323EEB645F540235E13A61FA4550 |
| stdout.bin | 7897C2DABE5A26308CEC4333865812C175F30D8B3D5A5A70E20EE4AFEC874FA5 |
| stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Raw packet retained ignored at
.superpowers/sdd/2026-10-18-p2b-payloadlength-diagnostic/.
Original source/oracle/packet retained; no replay, rewrite or deletion.
868mapped pins, sealed module and original raw/index identities rechecked
after the diagnostic. No new module import, test host, mutation or reviewer.

## Handoff and limits

Systematic debugging required a single hypothesis and smallest isolated
experiment before proposing a repair. The skill's referenced Codex tool-map
file was absent; existing PowerShell tools provided the equivalent workflow,
without changing scope or adding agents. No implementation skill/fix was used.
Necessary scripts/report/status/recovery are preserved on the existing
development branch under standing checkpoint authorization, not main.
Fresh clone recovers these portable files, not ignored raw/cache/mappings;
it does not grant replay rights or reset any consumed/frozen budget.

Recommended next scope: separate test-only one-expression correction in a new
candidate, with fresh explicitly bounded verification while preserving original
tests/oracle/records. Do not spend old frozen slots or automatically import/run.
No product fix, further diagnosis, runtime/security grant, acceptance or merge
is authorized by this report. P2-A limited acceptance remains unchanged.
