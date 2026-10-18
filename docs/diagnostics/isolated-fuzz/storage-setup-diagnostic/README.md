# Synthetic storage setup diagnostic

Scope: one ordinary PS7 process, 10 seconds plus 2 seconds owned exit and 2 seconds
stream drain. No candidate import/call, old budget reuse or observer/full matrix run.
This is a new focused setup diagnosis, not a rerun of the failed 45-case batch.

Two fixed synthetic Int32 paths / 14 trace records:
1. Same Fresh-Value, Owner, P1O-Leaf and field assignment expressions as the failed
   setup. Report pre/post native stored type, captured types, descriptor/base identity.
2. Separately split and instrument the raw array box, helper return, nested return,
   owner parameter/slot, capture, constructor argument/result, wrapper helper return
   and final slot. A direct note-property construction is the raw-value control.

Only eight allowlisted definitions are extracted as AST data from the original
test into a fresh module. No test top-level code, V-Others, V-Invariant or candidate
function runs. SDK reflection/trampoline/native type observations are retained.
Trace payloads contain type names and booleans only, never private/live data.

Additional locals and calls can affect PowerShell wrapping. The split path is
diagnostic context, not an exact proof of where any invisible intermediate changes.
The final RESULT means trace inventory completed, NOT fixture validity, observer
equivalence, a speedup, product correctness or acceptance.

Old failures/source remain unchanged. One-shot controller pins the previous failed
batch, all 627 inherited inputs, five receipt anchors and this leaf before launch;
consumes its allowance first and refuses overwrite/retry. Preserve all raw records.
On failure, perform read-only review only. Candidate/fixture repair is not authorized
by this diagnostic. Fresh PC recovery uses the result report; raw receipts and the
fixed local engine are intentionally machine-local and cannot be restored by cloning.
