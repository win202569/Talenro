# Bounded statecontract panic-cleanup follow-up — 2026-10-10

## Approved scope and status

User approved the short in-chat design. This is an existing-flow bounded fix,
not a new architectural increment: one existing evaluator plus one new test.
No new spec or implementation-plan document. Existing worktree/dirty progress
and all prior records retained. No staging, commit, push, merge or cleanup.

Complete locally. Targeted RED/GREEN and one fresh source-only review completed
with 0 Critical / 0 Important / 0 Minor. All fourteen approved offline scoped
commands completed exit 0. This is bounded local completion, not project/main/
production acceptance.

## Cause and minimal correction

Previously EvaluateModel allocated a fully cloned ContractModel, then called
ctx.Err again. Ordinary cancellation explicitly cleared that owner, but a panic
from that callback bypassed Clear; a caller cannot clear an owner never returned.
The earlier projection-binding increment correctly retained this as an exclusion,
since lower product changes were not authorized within that increment.

This separately approved follow-up preserves the function's public type and
all existing checks, error precedence and proof grades. Named result values
allow immediate deferred Clear iff result remains nil. Only success after the
last callback transfers ownership. There is no recover or panic conversion.
The now-redundant cancellation-only Clear is removed. RecheckModel is unchanged.

## Tests and evidence boundaries

New cleanup_test.go provides a conservative AST-tail registration check with
unsafe-shape negative examples; actual success/canceled/deadline/exact-panic
outcomes; both Recheck panic boundaries preserving original owner/caller input;
and actual Clear zeroing all nineteen independently populated, retained mutable
backing slices without changing caller collections.

Pinned offline Go 1.26.5, GOENV/GOWORK off, local toolchain, proxies/checksum/auth/
VCS network disabled, modules read-only. No dependency or module/lock changes.
Selector before and after the correction:

`^TestContract(EvaluateCleanupRegistration|EvaluateCleanupOutcomes|RecheckCleanupPanicPreservesOwner|CleanupOwnedBackingStores)$`

RED completed exit 1, package 3.207s: sole expected failure was the missing
immediate deferred-cleanup assertion, not a compile/environment error.
Product was unchanged until that RED. Same selector GREEN exit 0, 2.963s.
Runtime outcome tests already passed before correction; they are not claimed
as a direct dynamic erasure reproduction. The structural assertion is the RED.

The unreturned private owner is not dynamically observable without production
hooks or unsafe access, neither introduced. The guard plus real Clear backing-
store tests is bounded evidence, not formal arbitrary-Go control-flow proof.
Partial clone/OOM panic, immutable strings, unreachable/caller/runtime copies
and concurrent misuse remain outside guarantees. No whole-process erasure.

## Preservation checkpoint

Baseline: 447 Go/SQL/module/lock-scope files and 87 existing Git status entries.
Only internal/nodecontrol/statecontract/evaluate.go differs; 446 unchanged.
Only new source is internal/nodecontrol/statecontract/cleanup_test.go, census 448.
All eighteen prior binding freeze hashes still match. Pinned format-list clean.
HEAD 6752b9612a00b9d6b0bf87ae8dd95063f6b056ee and branch
codex/c12-b01-task9-coordinator remain unchanged. Index SHA256 remains
1CBA46BEF4C1AFDC9AE7938146909F0D2D641A0D712CE18036B60FE5B8673D12.
Final verification repeats and confirms this preservation checkpoint. All 87
previous status entries retained (now 88, adding this execution record); source
scan, format-list and read-only diff-check each exit 0. Both candidate hashes
remain exact after review/tests. Only documents changed after final source checks.

Candidate SHA256:

- evaluate.go: 0832CFAEC72F629EF47EE054FB3ECBF7686AC3C7CA5267E54F3FC3EC463D7CA8
- cleanup_test.go: 90256EF36CF9C3C8967A61E3FC780CD43EDB11D5294023C006310761ECAB4164

Scratch: .superpowers/sdd/2026-10-10-statecontract-panic-cleanup/ retains baseline,
exact pre-fix evaluator, RED/GREEN, review brief and verification commands.

## Final scoped verification — all fourteen commands completed

Each command independently uses pinned offline Go 1.26.5 with -count=1.
Full packages: ./internal/nodecontrol/{statebinding,statecontract,statematerials,
statecomposition,stateformat,stateverify,statepolicy,staterecovery,statetime,
serving,contracts}, and ./internal/readiness (one package per command).
Authority only: ./internal/nodecontrol/authority with selector
`^(TestBoundRead|TestCoordinator)`. Store compile only: ./internal/store with
selector `^$`; no store tests executed. Not full authority/store test acceptance.

| Package / scope | Exit | Reported package time |
| --- | --- | --- |
| statebinding | 0 | 19.085s |
| statecontract | 0 | 13.519s |
| statematerials | 0 | 15.816s |
| statecomposition | 0 | 24.982s |
| stateformat | 0 | 9.955s |
| stateverify | 0 | 7.457s |
| statepolicy | 0 | 10.574s |
| staterecovery | 0 | 6.998s |
| statetime | 0 | 8.798s |
| serving | 0 | 6.305s |
| contracts | 0 | 6.889s |
| readiness | 0 | 5.500s |
| authority selected | 0 | 6.875s |
| store compile only | 0 | 2.750s; no tests to run |

All fourteen sessions ended, no failure/cancellation/session left pending.
Full results: scratch verification-results.json and verification-all.log.
Final source/format/Git evidence: scratch final-handoff-checks.json.
No source changes after review or these tests; no second independent review.

## Remaining acceptance limits

The sole reviewer independently checked both candidate hashes and exact delta,
read both complete files plus ownership/fixtures/guards, and found no verified
blocker. Reviewer did not run tests, mutate state, use network or delegate.
No source changes afterward. Full report and every author disposition/cost remain
in scratch review-report.md and review-dispositions.md. No second review.
Retain all nine exclusions: partial clone/OOM; fatal termination/prevented defer;
immutable/unreachable/caller/runtime copies; concurrency; arbitrary-Go AST/
control-flow proof; direct private-owner observation; independent tests/race/
benchmarks/broader environment; earlier whole dirty branch/project/domain/
physical/runtime acceptance; Git integration and cleanup. These limits do not
invalidate the bounded closure but prevent broader guarantees.

The approved matrix is twelve full packages (binding, contract, materials,
composition, format, verification, policy, recovery, time, serving, contracts,
readiness), selected authority tests and store compilation only. This explicit
scope overrides the generic TDD skill's whole-repository command. Cost: broad
compatibility, store test/DB/Docker/race/hardware/restart/C11/C12/PITR acceptance
remains UNKNOWN. No whole dirty branch audit, merge or main acceptance.

Trusted old state, anchor/platform provenance, unsigned historical association
declarations, previous/unselected signatures, history/set truth, other domains,
current time, physical commit and runtime authority remain unproved. This cleanup
does not strengthen lower or binding evidence and does not close Task10/B01/C12/
main production acceptance gates.
