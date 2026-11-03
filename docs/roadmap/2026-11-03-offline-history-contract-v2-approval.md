# Versioned history-contract test correction: approved scope

The user replied “是” to preserving originals, correcting a versioned test and
adding one auxiliary verification after checkpoint625f5e91.

- Preserve original RED22/GREEN23, all13stops, frozen inputs/tests and candidate.
- Correct HC-Calls pipeline array shape and every affected consumer in a new
  history-contract-v2.tests.ps1. Keep all20groups and acceptance requirements.
- One additional GREEN24 only, ceiling24. The actual GREEN23 ComparisonFailure
  is the saved failing evidence; do not replay it or RED22.
- No production edits, Prepare, host, snapshots, compilation, reviewers, installs,
  main merge, deletion, stop bypass or acceptance expansion.
- Authenticate old files, source hashes, new amendment/test/helper hashes;
  archive/push preparation before the one run, then preserve actual outcome.
- First unexpected preparation/execution/recording result stops without retry.
- Even20PASS ends this scope. The old runtime gate remains stopped and still
  references original receipts; a corrected auxiliary is not runtime approval.

Implementation sequence: inspect saved failure and all call consumers; author
versioned test and bounded helpers; parse and fix identities; checkpoint; invoke
once; verify/preserve exact results and synchronize recovery documentation.

Skill application: systematic debugging uses saved failing evidence; TDD's next
verification is the single authorized correction run, not extra diagnostics.
No fresh subagent/reviewer or broad suite is authorized by this bounded approval.
