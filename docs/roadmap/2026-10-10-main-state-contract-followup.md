# Main-state contract minor follow-up — 2026-10-10

Scope: the user asked to continue after the two explicitly deferred Minor findings. This follow-up closes only ownership-test coverage and nested coverage error classification; no new model capabilities, integrations or authorization.

Status: both deferred Minors closed; targeted RED→GREEN and full approved regression completed. Uncommitted/unpushed.
Worktree: `codex/c12-b01-task9-coordinator`, HEAD `6752b9612a00b9d6b0bf87ae8dd95063f6b056ee`. No Git writes, commits, push, merge or cleanup.

## Changes

- New `followup_test.go` builds two independent expected/input fixtures and populates every collection in both snapshots plus RootTransition (19 collections). Each subtest mutates exactly one caller collection, compares the entire privately owned snapshot, rechecks against independently constructed original values, and confirms Clear does not modify caller storage. Expectations do not use product cloning helpers. Existing ownership logic was correct and unchanged; tests passed before the classification fix, with no artificial RED claimed.
- Table-driven public-entry tests cover previous/proposed verified/applied/transition records: Unknown, invalid tags and NeverEstablished with residual fields must return IncompleteCoverage; Present with malformed data must still return InvalidInput. Valid NeverEstablished records remain accepted.
- `shape.go` now classifies missing/illegal nested coverage before validating Present data. Rejection remains fail-closed; no invalid input gains acceptance.

## Evidence

Fixed offline Go1.26.5, approved package scope only. Targeted command:

`go test ./internal/nodecontrol/statecontract -run '^TestContract(OwnershipEveryCollection|NestedCoverageErrors)$' -count=1`

Before fix: exit1, 1.713s. Sixteen nested coverage subcases reproduced the InvalidInput/IncompleteCoverage mismatch; ownership and malformed-Present controls did not fail.

After the minimal classification fix: same targeted command exit0, 1.722s.

Final full approved regression, all exit0: statecontract6.426s, staterecovery6.392s, statetime6.785s, stateformat7.188s, stateverify5.620s, statepolicy5.849s, serving8.325s, contracts6.326s, readiness4.134s. Selected authority TestBoundRead/TestCoordinator8.143s; store compile-only13.398s (no tests). Fixed gofmt clean.

SHA256 baseline: 402 existing Go/SQL/module files; only shape.go changed, 401 unchanged; followup_test.go is new. No dependency changes. No whole-repository, DB/Docker, race or physical-platform tests.

This bounded follow-up is author-verified, not a new independent review. Prior review results/history remain preserved in the initial execution record. All L1–L5 and overall Task10/B01/C12/main production gates remain open.
