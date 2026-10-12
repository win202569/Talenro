# Main-state materials subset execution — 2026-10-10

Approved spec and five-task Native plan implemented locally; all five tasks complete.
Approved regressions and one fresh source review finished. No stage/commit/push.

Scope: standalone RootSet raw parser; exact embedded root identity and detached
desired slot facts in stateformat; isolated statematerials checks bounded raw
materials, explicit expected fields and subset relationships. Full composition,
signatures/anchor, history/other domains, physical commit and runtime authority
remain outstanding. No production consumer or L0 import exception added.

Baseline: HEAD 6752b9612a00b9d6b0bf87ae8dd95063f6b056ee, 403 existing Go/SQL/module
files. Only artifact.go/desired.go/envelope.go/root.go in stateformat changed;
399 existing files unchanged. New files are statematerials and two subset
stateformat files. Existing dirty and untracked work remains preserved.

Evidence: stateformat baseline exit0 2.397s; Task1 expected absent API RED exit1,
targeted GREEN exit0 2.419s, full stateformat exit0 4.801s. Task2 expected absent
input/API RED exit1, targeted GREEN exit0 2.451s (test syntax fixed before RED).
Task3 relationship assertion RED exit1 2.316s, GREEN exit0 2.462s, full package
exit0 4.977s. Task4 absent result/recheck API RED exit1, GREEN exit0 2.619s;
Task5 absent guard helper RED exit1, full new package exit0 2.605s. Gofmt clean.
Ownership enumerates all 12 populated mutable collections with independent
fixtures. Valid alternate envelopes and metadata bindings fail model recheck.

Workflow rulings: PowerShell/direct tests and apply_patch ledger replace Bash
helpers; retain scratch because there is no Git commit record. Tests stay within
the approved scope, so unrelated suite status remains unknown. Evidence lives
in this record and .superpowers/sdd/2026-10-10-main-state-materials-subset/.

Approved regression output: statematerials 7.994s, stateformat 7.453s,
statecontract 7.303s, staterecovery 7.226s, statetime 7.284s, stateverify 9.004s,
statepolicy 9.262s, serving 8.528s, contracts 5.734s, readiness 3.664s;
selected authority exit0 8.802s, store compile-only exit0 9.268s (no tests).
All test processes and the one source review have finished.
Git read-only HEAD check unchanged; diff --check exit0.

Fresh read-only gpt-6-astra high review: 0 Critical / 0 Important / 3 Minor.
All five Review Focus requirements confirmed. Minor test gaps closed by author:
same-coordinate root key replacement reaches the actual embedded identity guard;
two disconnected historical rotation fragments with explicit 2020 windows pass;
valid expected slot-b versus actual slot-a fails. TestSubsetReviewCoverage first
passes the reviewed source exit0 2.544s; deliberate mutations fail all three
subcases exit1 2.502s; all mutations restored, 20/20 reviewed product hashes
identical. Final statematerials exit0 2.620s and stateformat exit0 2.568s.
The total-byte limit test also covers exactly 16MiB+1. No deferred minor remains,
no product fix was needed, no second independent review. Unchanged surrounding
packages retain the earlier successful approved regression evidence.

Ruling: the approved plan permits Minor resolution; missing required acceptance
assertions were completed immediately within test scope. Cost: targeted extra
validation. The one-review method remains preserved.
Reviewer exclusions accepted against spec: signatures/anchor; current key/time/
resource authority; full history/material closure, cross-batch rollback and L0
composition; physical commit/recovery/runtime/integration; concurrent Clear and
runtime/string erasure; database/Docker/race/hardware/global acceptance. Misusing
this subset result for those claims would be unsafe; unsupported concurrency may
race and detached copies remain. They need separate later gates, not inference
from this increment's green tests.

Worktree and scratch evidence retained; no integration/cleanup was requested.
This is completion evidence for this subset only; overall project gates remain open.
