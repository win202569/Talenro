# Storage-aware observer fixture v2 — PREPARED ONLY

This is an independent copy of the failed inline observer offline harness.
Do not run it or the old controllers under the completed setter-check allowance.
There is deliberately no launch controller/new execution allowance in this leaf.

All45 case definitions, literal expected outcomes and original checks are retained.
The original/optimized observer test sources are referenced read-only as AST data.
The same ten allowlisted functions are used, plus the single validated S-SetStored
definition parsed from ../reflected-storage-fixture/set-stored.ps1.

Changes are restricted to real fixture storage construction:
- Public SDK reflected setter retains raw versus PSObject storage on assignment.
- WrappedString is explicitly initialized as stored PSObject, and both captured
  primary leaves must have the independently specified initial native storage type.
- Type mutation uses reflected set with existing wrapped/raw replacement logic.
- Equal-value fresh-base and deliberate-sharing setup use the same setter so the
  WrappedString representation is not silently erased by PowerShell assignment.
- Same-owner extra field is initialized consistently if WrappedString is used.

No observer implementation, matrix size, expected outcome, failure guard or timeout
has been weakened. Whole original/optimized sources still must match outside V-Others.
Setter construction passed5kinds/10transitions; this NEW full harness is NOT runtime
validated, so held/reread/shared/base fixtures and all45paired outcomes remain open.
No candidate import/call, product execution, performance or acceptance claim.
See latest roadmap report for actual results and proposed next bounded allowance.
