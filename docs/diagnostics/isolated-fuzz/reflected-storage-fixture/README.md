# Reflected storage fixture validation

Test-only public PSNoteProperty.Value setter through the existing MethodBase.Invoke
trampoline and fresh object-array arguments. No private members, settings changes,
candidate imports/calls or observer/full ownership suite.

One ordinary PS7 host, 10s execution + 2s owned exit + 2s stream drain. Fixed
five storage kinds / ten transitions: Int32, Int64, Boolean, raw String start raw,
wrap then unwrap; pre-stored PSObject String starts wrapped, unwrap then wrap.
Each transition requires the literal expected native type, a real type change,
same descriptor/base identity, same base type/value, no metadata and no output.
Changing the setter to ordinary assignment must fail the existing real-type guard.
The earlier failed batch and completed 14-record trace are the preserved negative
witness, not rerun here. This is a targeted construction check, not a full suite.

setter.tests.ps1 was written before set-stored.ps1. It extracts only the same
eight known SDK helper definitions from original test AST data and the single
S-SetStored definition from its separate test utility. No whole-test code runs.
The prior 637 frozen inputs and four raw receipt anchors remain pinned unchanged.
Any failure stops without repair/retry. A positive result permits preparing a
separate corrected fixture copy; it is not 45-case equivalence or speed evidence.
Old test/diagnostic inputs and budgets stay frozen. Portable source survives
cloning; historical raw receipts/engine cache remain machine-local.
