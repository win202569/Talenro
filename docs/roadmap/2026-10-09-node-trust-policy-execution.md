# Node trust policy execution — 2026-10-09

Approved spec: ../superpowers/specs/2026-10-09-task10-node-trust-policy-design.md.
Approved plan: ../superpowers/plans/2026-10-09-task10-node-trust-policy.md.
Native execution in codex/c12-b01-task9-coordinator; BASE/HEAD at start
6752b9612a00b9d6b0bf87ae8dd95063f6b056ee. All work uncommitted/unpushed.

## Current evidence

New internal/nodecontrol/statepolicy composes existing stateverify and
stateformat without modifying their product code. Positive results are conditional
models only: no platform authentication, trusted clock implementation, storage,
hardware commit, production caller or production authorization.

All five task test groups were written before their implementation. New-interface
RED failures reported missing loadFacts/checkTime/checkTransition/checkHistory/
EvaluateModel/Candidate/RecheckModel; exit 1. No environment failure counted as RED.
Task 1's first GREEN attempt exposed an invalid test assumption: all-zero signatures
are rejected by the format layer. Replaced the test input with a nonzero signature
with one flipped bit; existing parser semantics were not relaxed.

Observed GREEN exits 0:

| Scope | Reported test time |
| --- | --- |
| Baseline stateverify | 1.955s |
| Baseline stateformat | 1.821s |
| Task 1 facts/ownership | 3.253s |
| Task 2 explicit time | 3.226s |
| Task 3 transition | 1.840s |
| Task 4 history | 2.590s |
| Full statepolicy after Task 5 | 2.838s |

Post-implementation regressions returned stateformat4.813s, serving14.934s,
contracts9.510s, readiness10.601s, selected authority20.654s, store compile-only
6.976s (no tests), all exit0. stateverify initially failed TestStateverifyBoundary
because the previous guard forbade all non-test consumers. The approved isolated
composition is now allowed only for the exact statepolicy directory/package;
the same scan forbids non-test consumers of statepolicy. New guard tests RED
undefined isolatedImportAllowed, then full stateverify GREEN2.141s exit0.

Independent gpt-6-astra/high read-only review found 0Critical/2Important/0Minor:
revoked-deadline reconciliation and allocated-empty slice false staleness. Author
reproduced both through public EvaluateModel/RecheckModel; expanded TestReview*
RED5.627s exit1 includes present/absent and known/unknown revoked-history cases.
Both findings have targeted GREEN evidence: revoked deadlines3.512s and empty-list
snapshot/recheck3.677s, exit0. The final scoped suite completed, all exit0:
statepolicy4.390s, stateformat6.428s, stateverify3.846s, serving4.915s,
contracts4.368s, readiness3.641s, selected authority6.755s, and store compile-only
3.141s (no tests). This isolated model increment is complete within the approved
scope. No second independent review; reviewer did not execute tests.
Formatting list empty and trailing-whitespace scan found no
matches. Initial protection check: 251 scoped pre-existing files hashed, only the
current plan's execution status had changed. New files are outside that baseline.
Final code protection check additionally identifies the intentional stateverify
boundary_test.go change; no pre-existing product source was changed. The current
status, approved spec status and plan completion records are intentional document
updates. All19 new Go files have no whitespace diagnostics; tracked diff check and
gofmt listing are clean. HEAD remains 6752b9612a00b9d6b0bf87ae8dd95063f6b056ee.

## Recorded rulings

1. PowerShell/apply_patch equivalents replace Bash-only skill scripts, preserving
   plan identity, BASE, task briefs and evidence. Risk: manual bookkeeping omission.
2. No commits, cleanup, downloads or full-suite expansion, per approved constraints.
   Consequence: uncommitted progress and broader verification remain outstanding.
3. A new revocation requires fresh observations scoped to old state and target
   metadata; an old DeadlinesKnown flag cannot prove no later issuance. Consequence:
   conservative rejection when fresh historical evidence is absent.
4. Candidate.State returns Status Unknown while retaining old snapshot identifiers;
   a proposal cannot be reused as a committed snapshot. Test-only simulated commits
   must explicitly supply Available status/new identity. No production authority.
5. Permit only the exact approved statepolicy package to import stateverify, while
   forbidding runtime consumers of statepolicy in the same boundary scan. This
   modifies a test guard, not verifier product code. Risk: incorrect dependency
   classification; positive and negative exact-path/package cases cover it.
6. Candidate tests live in evaluate_test.go alongside the shared public-entry
   fixture tests rather than a separate candidate_test.go. No behavior change;
   the only tradeoff is test discoverability.

Reviewer declined-to-judge rulings are retained, not silently accepted:
platform provenance/reboot guarantees require later real-provider validation;
atomic commits, crash recovery, durability, quarantine, TLS and production
authorization require later integration; concurrent mutation/Clear remain an
unsupported caller contract; chains beyond32 remain unsupported; unrelated
history and overall Task10/B01/C12/main remain unaccepted; runtime evidence is
author-run only, not independently rerun. Each is a limitation, not a production
security claim.

## Acceptance boundaries

No real DB/Docker, full repository, C11/C12, bootstrap, PITR or race tests.
Real platform provenance, reboot anti-rollback, durability/crash recovery, TLS and
production authorization remain unverified. Deployment compatibility UNKNOWN;
not full Task 10/B01/C12, main or merge acceptance. Existing unrelated test Minors
remain outside scope. Private rule tests sometimes mutate verified fact copies to
isolate a branch; public entry tests use signed raw samples and real verification.

Scratch evidence retained in .superpowers/sdd/2026-10-09-task10-node-trust-policy/;
it is not removed because no commit history preserves this work yet.
