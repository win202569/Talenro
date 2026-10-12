# Main-state dual-snapshot projection binding — execution evidence

2026-10-10; approved design and five-task plan, Native execution.
Worktree: D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator.
Branch: codex/c12-b01-task9-coordinator.
Status: all five tasks complete locally; sole review 0 Critical / 0 Important /
2 Minor, both closed by author; all fourteen post-fix commands completed exit 0.
No staging, commit, push, merge, reset, dependency download or cleanup.

## Delivered scope and proof limit

Isolated statebinding owns one bounded transaction snapshot: complete previous
and proposed twelve-domain L0 declarations, separately owned previous and proposed
material tables, an independent anchor and the explicit proposed desired target.
Exact Kind/P/E, node/epoch, root, metadata, established desired and the current
root transition are projected to their respective checked tables. The current
old-to-new root transition and the target's original authorizer chain stay
independent. Historical authorizers are not replaced with current metadata/root.
Legitimate same-P/different-E tables are preserved, never merged.

Evaluate calls the real complete L0, previous material-subset and proposed
single-target composition evaluators; it does not implement parsing, hashing,
signatures, clocks, I/O, policy, providers or a production consumer. The only
success is ConditionalProjectionBound. All lower responsibilities and the eight
binding obligations remain unchanged. Previous/unselected signatures remain at
their existing lower proof grade. A caller's previous declaration is not thereby
a trusted old baseline.

Summary detaches Links/Coverage; Clear reaches mutable owned buffers, not caller
or detached copies, immutable strings or unreachable runtime copies. Recheck
re-evaluates and compares the complete normalized input, including unprojected
history/platform fields, inactive expectation branches and unselected material.
Only top-level tables and nil/empty collections normalize; ordered L0 and chains
are not sorted. Inputs must not mutate during calls; owners must not be shallow
copied or concurrently used/cleared.

A coherently rewritten compatible metadata authorizer may produce a fresh valid
conditional result while selected desired raw/P/E stay unchanged. Complete old
Recheck rejects it. This is identity binding, not proof that the original
historical association was signed into the wire object.

## Workflow rulings and costs

1. PowerShell/apply_patch ledger replaces Bash skill scripts, commits and cleanup,
   under the approved Windows/no-Git scope. Cost: progress is local, not recoverable
   from Git history; the worktree and scratch are preserved.
2. Fixed offline scoped tests replace dependency downloads/full-repository setup.
   Cost: broader compatibility and production acceptance remain UNKNOWN; no DB,
   Docker, race, hardware, restart, C11/C12/PITR or whole-repository test is claimed.
3. One final fresh read-only review covers this increment and three exact parent
   test-guard changes, not existing dirty progress. Cost: unrelated branch defects
   are outside this independent review. Scoped Minor disposition follows the
   approved plan; no second independent review, and fixes require prescribed
   author verification rather than an expanded independent-review claim.

## Baseline, environment and evidence

Pinned cached Go 1.26.5, GOENV/GOWORK off, GOTOOLCHAIN local,
GOPROXY/GOSUMDB/GOAUTH off, GOVCS all:off and GOFLAGS -mod=readonly;
GOPATH/GOMODCACHE use the existing local cache. No module/lock changes.
HEAD is 6752b9612a00b9d6b0bf87ae8dd95063f6b056ee.
Index SHA256 is 1CBA46BEF4C1AFDC9AE7938146909F0D2D641A0D712CE18036B60FE5B8673D12.
Among 432 baseline Go/SQL/module files, only the three approved boundary_test.go
files in statecontract/statematerials/statecomposition differ; 429 remain exact.
Pre-format guard snapshots are SHA-verified against baseline for readable review.
Exactly 15 new Go files plus those three guards were formatted (exit 0);
independent format list exit 0 and empty. HEAD/branch/index and read-only
diff-check are unchanged/exit 0. Eighteen post-format source/test hashes are saved.

Evidence and task briefs remain at
.superpowers/sdd/2026-10-10-main-state-projection-binding/.
All started sessions are awaited; output delays are not classified as timeouts.

## Test-driven steps and restored probes

All five task selectors are the exact approved plan selectors; expected missing
API/helper failures are RED exit 1, not environment failures.

- Task 1: budget equality/+1, nineteen collections, inactive expectations,
  repeated raw aliases, independent owning and exact consumer guards.
  Targeted GREEN exit 0; four full packages exit 0.
- Task 2: independently generated fixtures, 0/1/32 chains, all two-sided field
  projections, ordered keys/revoked/slots, RootSet E=0, same-P/different-E,
  old desired NeverEstablished and distinct current/original authorizer chains.
  Targeted 3.779s and full binding 5.260s, both exit 0.
- Task 3: integrated lower evaluators, fixed errors, proof-strength limits.
  Targeted 2.619s and four full packages exit 0.
- Task 4: detached ownership, all nineteen nonempty lower-valid collections,
  every L0 leaf mutation, fresh-valid unprojected replacements, complete recheck,
  all observed cancellation/deadline and owned-stage panic checkpoints.
  Targeted 3.455s, full binding 3.847s, restored full 3.960s; exit 0.
- Task 5: expected missing AST helper RED exit 1. Conservative approved-shape AST
  rejects 36 unsafe variants; actual import/consumer scans and exact parent
  guards GREEN exit 0 (binding 3.665s, contract 2.791s, materials 2.818s,
  composition 3.541s). This AST is a structural guard, not a general control-flow
  proof or an observation of every unreachable lower-layer allocation.

The three omission probes each produced the intended assertion RED:
omitted previous projection, omitted Evaluate current-transition binding,
and omitted whole-input Recheck comparison. Each was immediately restored and
hash-verified before full binding GREEN. Pre-format restored hashes:
projection.go 7CD20DB6067143F4AFCCE8EF73C89B46E6154FF54129AFB3BDA47ACDD5740468;
evaluate.go 75BDFFB9DB298900581E2A9727D66DE8532F54365D921A1650C7AA468D5A062C;
model.go E15D2692554E05997632E4903EC99710F5F1DB2346CDE4B6C39DFD7884DB0D77.
No probe remains. Fixture namespace/empty-process/wrong-root-index diagnostics
and the unused-import probe compile failure were documented separately and
never counted as behavioral RED.

## Pre-review fourteen approved commands — all completed

Each command is independent, uses the pinned offline environment and -count=1.
Full packages: statebinding, statecontract, statematerials, statecomposition,
stateformat, stateverify, statepolicy, staterecovery, statetime, serving, contracts
under ./internal/nodecontrol/, and ./internal/readiness.
Authority only: test ./internal/nodecontrol/authority
-run '^(TestBoundRead|TestCoordinator)' -count=1.
Store compile only: test ./internal/store -run '^$' -count=1.
Store compilation is not store test acceptance.

| Package / scope | Exit | Reported package time |
| --- | --- | --- |
| statebinding | 0 | 8.610s |
| statecontract | 0 | 6.549s |
| statematerials | 0 | 7.069s |
| statecomposition | 0 | 9.617s |
| stateformat | 0 | 5.505s |
| stateverify | 0 | 5.735s |
| statepolicy | 0 | 5.183s |
| staterecovery | 0 | 4.388s |
| statetime | 0 | 16.093s |
| serving | 0 | 7.528s |
| contracts | 0 | 7.356s |
| readiness | 0 | 6.737s |
| authority (selected) | 0 | 6.169s |
| store (compile only) | 0 | 3.609s; no tests to run |

All fourteen sessions completed with exit 0; each result is separately recorded
in final-regression-all.log and final-regression-results.json. After completion,
all eighteen frozen source/test hashes still match and the index is unchanged.

## Single incremental review — complete; two Minor findings closed by author

Exactly one clean-context gpt-6-astra xhigh reviewer read all fifteen new Go files,
the three exact guard diffs, approved documents, relevant lower integration paths,
and all evidence. Independent hashes confirmed all eighteen original freeze files,
three approved differences among 432 baseline files and the baseline snapshots.
Reviewer did not run tests, mutate state, use network or delegate.
Original verdict: With fixes, 0 Critical / 0 Important / 2 Minor.

Minor 1 accepted: missing current transition start was converted to BindingMismatch,
contrary to MissingMaterial classification. New historical-v1/separate-current-v2-v3
fixture independently passes all lower layers and both projections; the integrated
case then assertion-REDs with nil model/wrong sentinel. Targeted RED exit 1, 2.704s;
preserve lookup error, check coordinates separately, then GREEN exit 0, 2.603s.
The private old-payload endpoint expectation is corrected as well. No acceptance
or proof-grade expansion. This product correction triggers all fourteen reruns.

Minor 2 accepted: add exact per-table 16 MiB + 1 byte cases, other budgets in
range. Already correct budget.go remains unchanged; new tests pass. Coarse cases
retained. Combined cap is implied by table/anchor limits; no isolated overflow
claim. The initial grouped slash selector selected only the private assertion,
not the integrated case; not counted as integrated RED. Correct selector then
reproduced the real bug before any product edit.

Exactly target.go and four new test files changed after review; thirteen other
reviewed files stay exact. Original reviewed freeze retained, separate post-review
freeze saved. No second independent review; correction verified by author, with
all fourteen post-fix commands completed exit 0. No deferred Minor finding.
Full report and every author disposition are retained in final-review-report.md
and review-dispositions.md in this plan's scratch directory.

All reviewer exclusions explicitly retained: prior dirty progress, trusted old
state/durable anti-rollback, atomic external acquisition, anchor/platform provenance
and opaque ID equivalence, unsigned original historical associations,
previous/unselected signatures, history/set truth, all other domain material
validators, clock/runtime permission, physical persistence/hardware/recovery,
independent lower algorithm audit, unsupported concurrency/shallow copies,
immutable/unreachable/caller copies, arbitrary-Go formal sandbox/control-flow
proof, and excluded broad test suites. Costs: none may be inferred from this
conditional projection result. In particular, the unchanged lower contract owner
can be allocated before a panicking context callback without cleanup; binding
cannot certify erasure of an owner never returned by that layer. No lower product
change is allowed here and no complete runtime-memory erasure is claimed.

## Post-review final regression — all completed

The corrected candidate reran all approved commands in the same pinned offline
environment. Every independent session ended with exit 0:

| Package / scope | Exit | Reported package time |
| --- | --- | --- |
| statebinding | 0 | 5.165s |
| statecontract | 0 | 3.166s |
| statematerials | 0 | 3.178s |
| statecomposition | 0 | 3.473s |
| stateformat | 0 | 2.825s |
| stateverify | 0 | 2.895s |
| statepolicy | 0 | 3.353s |
| staterecovery | 0 | 2.790s |
| statetime | 0 | 6.003s |
| serving | 0 | 5.476s |
| contracts | 0 | 6.517s |
| readiness | 0 | 5.154s |
| authority (selected) | 0 | 8.073s |
| store (compile only) | 0 | 3.147s; no tests to run |

Outputs: post-review-regression-all.log / post-review-regression-results.json.
Final eighteen post-review freeze hashes match. Source census is 447:
432 baseline plus fifteen added files, all in statebinding. Only three approved
baseline test guards differ; 429 unchanged. All 85 previous dirty status entries
preserved (current 87); HEAD/branch/index unchanged, read-only diff-check exit 0.
Pinned final format-list exit 0/empty. No source changes after these tests.
Both Minor findings closed by author verification; no second independent review,
whole-branch approval or main/production readiness. Every ruling/cost retained.

## Remaining gates

Trusted old baseline, external anchor/platform provenance, signed original
historical associations, previous/unselected signatures, history/set digests,
domain/current-time validity, full material closure and full composition,
physical commit and runtime authorization remain unproved.
Task10/B01/C12/main acceptance gates are unchanged; production compatibility UNKNOWN.
All five local tasks and final document checklists complete. Final source hashes,
HEAD/branch/index and prior dirty progress verified unchanged within approved
exceptions. Worktree/scratch retained; no Git writes or cleanup.
