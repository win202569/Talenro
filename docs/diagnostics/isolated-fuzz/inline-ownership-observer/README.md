# Inline ownership observer — offline-only checkpoint

Approved scope: one ordinary PowerShell 7 offline batch, 30 seconds plus 2 seconds
owned-process exit and 2 seconds stream drain. No candidate imports/calls, no actual
ownership suite, no old allowance reuse. Any mismatch ends the batch without repair
or retry. The controller consumes its slot before launching and freezes all inputs.

Only V-Others differs from the stage-timing ownership test. SDK reflection and
fresh object-array argument construction remain; helper call frames are expanded
inside the scan. Original helpers and selected-leaf checks remain unchanged.
No history entries, assertions, test cases, timing boundaries or expected results
are removed. Full source outside V-Others must be identical after newline normalization.

offline.tests.ps1 parses both files as data and executes only these ten definitions
in separate dynamic modules: R-Reference, R-Type, R-Base, R-Wrapper, P1O-Fail,
P1O-Need, P1O-Leaf, V-Invariant, V-Others, P1R-Fresh. No top-level test code runs.

Fixed 45-case inventory:
- Five scalar storage forms × seven actions: clean, selected-only decoration,
  descriptor replacement, storage-type replacement, equal-value fresh base,
  expected base-type mismatch, held-view metadata contamination (35).
- Four synthetic fresh-reread-only branch probes. Their captured Members is
  deliberately replaced with an independent real clean collection; these are
  branch-sensitivity tests, NOT faithful product ownership fixtures.
- Two deliberate-sharing controls (Int32 and pre-stored PSObject String).
- Same-owner/different-field clean and dirty controls, a late dirty leaf and empty list.

Each side uses fresh fixtures, checks both phases and independent literal expected
outcomes, and requires no pipeline output. Setup guard failures are failures, not
successful contamination detection. This verifies observer equivalence/sensitivity
only, not product correctness, production behavior or a speedup. The optimized
whole ownership suite remains PREPARED ONLY and must NOT be launched by this batch.

Portable files are sufficient to inspect/reconstruct the proposal on a fresh PC.
The historical pin chain and raw receipts are intentionally ignored and machine-local;
capture.ps1 refuses to run without them. Cloning is not execution authorization.

