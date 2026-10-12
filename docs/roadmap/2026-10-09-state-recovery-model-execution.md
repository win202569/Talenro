# Task 10 state recovery model — execution record

Date: 2026-10-09. User approved design and implementation plan; Native execution.
Status: implemented locally; the single Important review issue was reproduced
and fixed by the author; post-fix targeted/scoped verification passed. One Minor
coverage item was deferred at completion and closed by the 2026-10-10 test-only
follow-up (see 2026-10-10-state-recovery-test-followup.md). No second independent review.
Worktree: `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`.
Branch: `codex/c12-b01-task9-coordinator`; BASE/HEAD `6752b9612a00b9d6b0bf87ae8dd95063f6b056ee`.
No staging, commits, pushes, merge, branch/worktree cleanup or dependency changes.

## Scope and ownership

Added `internal/nodecontrol/staterecovery` only: 9 product and 8 test files.
RecoverModel validates both slots and selects exactly one counter-matching blob;
PrepareModel distinguishes no-change from capacity-consuming changes;
AdvanceModel binds ordered model events; ConfirmModel independently rescans after
failure, uncertain response or acknowledged increment. Results are conditional
value-owned models, with explicit cancellation and invalidation semantics.

Available/Intact and durable events are caller assumptions, NOT actual disk,
TPM, seal, authorization or persistence evidence. No business high-water or
complete sealed-state schema is inferred from ContentDigest. No pointer action,
automatic retry, backup fallback, deletion, quarantine or LKG grant occurs.
No production consumer and no statepolicy/statetime mutation. AST tests reject
forbidden imports/member access through all aliases, including function values.

Protected baseline: 370 pre-existing Go/SQL/module files; initial post-format
comparison: all unchanged. Historical dirty/untracked changes preserved.

## Verification evidence

All commands use pinned Go 1.26.5, PowerShell, offline GOENV/GOWORK/GOTOOLCHAIN/
GOPROXY/GOSUMDB/GOAUTH/GOVCS settings and readonly modules from the approved plan.
`go` below abbreviates the pinned executable, not PATH resolution.
Every RED was the expected compile failure for missing new interfaces/helpers,
not an intentionally broken existing implementation. These are interface-first
REDs, not claims that every table row separately failed behaviorally.

| Run | Command suffix after `go test` | Result |
| --- | --- | --- |
| Baseline | statetime/stateformat/stateverify/statepolicy package paths, `-count=1` | exit 0: 2.657/2.686/2.812/2.998s |
| Task 1 RED | `./internal/nodecontrol/staterecovery -run '^TestRecovery(Select\|Observations\|DuplicateEpoch\|FutureAndPointer)$' -count=1` | exit 1, RecoveryInput/Scope absent |
| Task 1 GREEN | same | exit 0, 1.730s |
| Task 2 RED | `./internal/nodecontrol/staterecovery -run '^TestRecoveryPrepare' -count=1` | exit 1, ProposalObservation/PrepareModel absent |
| Task 2 GREEN/full | targeted then `./internal/nodecontrol/staterecovery -count=1` | exit 0: 1.491/1.491s |
| Task 3 RED | `./internal/nodecontrol/staterecovery -run '^TestRecovery(Advance\|Recheck)' -count=1` | exit 1, Event/AdvanceModel absent |
| Task 3 GREEN/full | targeted then full package | exit 0: 1.676/1.519s |
| Task 4 RED | `./internal/nodecontrol/staterecovery -run '^TestRecovery(Confirm\|Crash\|Ownership\|Cancellation)' -count=1` | exit 1, ConfirmationInput/ConfirmModel absent |
| Task 4 GREEN/full | targeted then full package | exit 0: 1.839/1.835s |
| Task 5 RED | `./internal/nodecontrol/staterecovery -run '^TestRecoveryBoundary' -count=1` | exit 1, AST helpers absent |
| Task 5 GREEN | same | exit 0, 1.464s |

Pre-review regression session 79019 completed, all commands exit 0:

```powershell
& $recoveryGo test ./internal/nodecontrol/staterecovery ./internal/nodecontrol/statetime ./internal/nodecontrol/stateformat ./internal/nodecontrol/stateverify ./internal/nodecontrol/statepolicy ./internal/nodecontrol/serving ./internal/nodecontrol/contracts ./internal/readiness -count=1
& $recoveryGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1
& $recoveryGo test ./internal/store -run '^$' -count=1
```

Package durations respectively: staterecovery 4.589s, statetime 4.905s,
stateformat 4.923s, stateverify 4.946s, statepolicy 5.034s, serving 5.147s,
contracts 3.938s, readiness 2.708s, selected authority 3.468s;
store compile-only 1.877s [no tests to run]. gofmt -l empty; tracked diff
whitespace check exit 0; HEAD unchanged. Eighteen named model tests plus their
table/subcases ran in the package suite. No whole-repo, real DB/Docker, C11/C12,
bootstrap, PITR, race or hardware/reboot tests run.

## Rulings made

1. PowerShell reads and apply_patch ledger/brief equivalents replace Bash-only
   skill helpers because the user requires PowerShell. Cost if wrong:
   bookkeeping divergence; exact plan briefs and command evidence are preserved.
2. Approved no-commit/no-cleanup/offline/scoped-test constraints override generic
   skill templates for commits, scratch cleanup, downloads and whole-repo tests.
   Cost: no broader compatibility or durable Git checkpoint claim. Scratch ledger
   retained because work remains uncommitted.

## Review and remaining gates

One final fresh independent read-only source review completed using
`gpt-6-astra` high: 0 Critical / 1 Important / 1 Minor. Reviewer read all 17 files,
approved spec/plan, ledger and execution evidence; no tests or baseline checks
were independently rerun. Verdict at review time: isolated model requires fixes.

Important: equal BlobDigest attributed to two different epochs was accepted by
RecoverModel. Root cause was comparing identities only inside the equal-epoch
branch. BlobDigest covers epoch, so this is a locally visible contradictory
observation, not a request to implement a seal. CandidateWritten already rejected
base identity reuse. `TestRecoveryBlobIdentityEpoch` covers both counters,
same/different content, and both slot orders. Real behavioral RED: exit1, 1.558s,
unexpected nonnil model with nil error. The author added a cross-epoch identity
guard; post-fix verification passed. Existing same-epoch
DuplicateEpoch/Fork precedence is preserved. No second independent review.

Minor deferred: TestRecoverySelect asserts FutureSlot differs from active rather
than the exact other slot; an unset diagnostic could escape that assertion.
Current product sets the correct diagnostic. Exact other-slot/no-future assertions
are a follow-up, not silently included in the one Important-fix pass.

Post-fix session 39878, all commands exit0: targeted
`./internal/nodecontrol/staterecovery -run '^TestRecoveryBlobIdentityEpoch$' -count=1`
1.556s; the same complete approved regression commands above returned
staterecovery4.141s, statetime6.122s, stateformat6.165s, stateverify6.044s,
statepolicy5.055s, serving3.406s, contracts4.020s, readiness2.674s,
selected authority3.383s, store compile-only1.988s [no tests to run].
The fixed suite includes 19 named staterecovery tests and their table/subcases.
The reviewer did not independently review this fix; author RED→GREEN plus scoped
regression is the approved verification. Final 370-file protected hash comparison
unchanged; gofmt -l empty. Branch/worktree and scratch evidence preserved.

### Final review rulings (including all declined judgments)

3. Cross-epoch same BlobDigest maps to SecurityFault/CorruptSlot, retaining
   same-epoch classifications. Reason: complete identity includes epoch.
   Cost if wrong: conservative refusal, never fallback.
4. Observation authenticity remains a caller assumption. Cost: forged observation
   labels or identities are not authenticated here.
5. Digest/seal/key provenance, full-state coverage, authorization and business
   high-waters remain unverified because bytes/validators are absent. Cost:
   successful model evaluation is not content or business rollback proof.
6. Writes/fsync/hardware increments/power loss/ACK provenance remain outside pure
   scope. Cost: no real durability proof.
7. Scan freshness/atomicity remains an explicit-ID premise. Cost: a new identifier
   does not establish a real fresh atomic scan.
8. Locks/concurrent writers/old-model replay/exactly-once/concurrent Clear remain
   excluded or prohibited. Cost: callers must enforce execution/lifetime rules.
9. Provisioning/reenrollment/migration/reset stay deferred. Cost: this module
   cannot initialize or migrate nodes.
10. Pointer persistence/retry/deletion/repair/quarantine remain absent. Cost:
    runtime must orchestrate safe actions; diagnostics authorize no actions.
11. Runtime consumers/three guards/latch/supervisor/LKG/TLS/compatibility remain
    separate integration gates. Cost: no production startup or overall acceptance.
12. Whole-repo/DB/Docker/race/hardware/reboot verification remains unrun under the
    authorized scope. Cost: behavior outside scoped verification is UNKNOWN.

Production compatibility UNKNOWN; complete state schema, business monotonicity,
trusted origins, seal/key separation, exclusive writer locking, actual fsync and
atomic hardware increments, three independent guards, LocalSecurityLatch-first
startup, controlled restart/replay rejection and TLS gates remain unproven.
Not overall Task10/B01/C12/main acceptance.
