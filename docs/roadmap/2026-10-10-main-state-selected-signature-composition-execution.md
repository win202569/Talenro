# Main-state selected-target signature composition — execution evidence

2026-10-10; Native implementation in existing
`D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator` on
`codex/c12-b01-task9-coordinator`. User approved the design and five-task plan.
No Git write, commit, push, merge, branch/worktree cleanup or new dependency.
Existing changes are preserved. Status: all five tasks complete locally; final
scoped verification passed and the single fresh read-only review is clean.

## Delivered scope (local, not accepted for production)

The new isolated `statecomposition` package owns a bounded complete input,
checks every material through `statematerials`, binds the independently supplied
v1 anchor and exact ordered refs, calls the sole real `stateverify` verifier,
and rechecks each returned identity and coordinate. It verifies one desired only.
Zero through 32 rotations are supported; disconnected high-version history
may remain material-only. No production consumer or policy/L0 exception added.
Only the two approved existing boundary test files may change; no lower product
code changes. Original subset responsibilities remain unchanged.

The detached summary distinguishes AnchorAssumed, ChainDerived,
SignatureChecked and MaterialOnly. Recheck reruns all stages and compares the
whole normalized snapshot, including unselected records. Clear zeros reachable
owned mutable storage, not immutable strings/inaccessible runtime copies.
The caller must not mutate input concurrently or shallow-copy/concurrently clear
the owner. No current clock or external provider is consulted.

Desired does not wire-sign metadata/root payload association; metadata does not
wire-sign root payload. A fresh, coherently rewritten valid declaration may
succeed. The previous model's Recheck rejects it; this is not authentication of
historical original authorization. Unknown-key nonzero shares are format-valid
but do not count toward signature thresholds.

## Workflow rulings

PowerShell/apply_patch ledger replaces skill Bash helpers, commits and scratch
deletion. The approved no-Git/scoped-offline plan overrides whole-repository
setup/tests and integration menus. Cost: production/whole-repository compatibility
is UNKNOWN and the existing branch remains local and dirty.
Native final Minor treatment follows the approved plan: bounded in-scope fixes
or explicit justified retention, rather than blanket deferral. No second review;
cost: scoped extra test time. No Critical/Important finding may be silently waived.

## Environment and baseline

Pinned Go 1.26.5 at the local toolchain cache. Each test uses GOENV/GOWORK off,
GOTOOLCHAIN local, GOPROXY/GOSUMDB/GOAUTH off, GOVCS all:off,
GOFLAGS -mod=readonly and existing GOPATH/GOMODCACHE. No download or lock change.
Commands below are scoped suffixes of this exact environment/pinned executable.
Every started process is awaited; output latency is not a test timeout and no
duplicate run is started while a session is active.

Scratch ledger and baseline retained at
`.superpowers/sdd/2026-10-10-main-state-selected-signature-composition/`.
Baseline 418 Go/SQL/module/lock files; HEAD
`6752b9612a00b9d6b0bf87ae8dd95063f6b056ee`; index SHA256
`1CBA46BEF4C1AFDC9AE7938146909F0D2D641A0D712CE18036B60FE5B8673D12`.
Mid-execution and final handoff checks found exactly the two approved parent
boundary tests changed; 416 other baseline files unchanged. All 16 frozen
source/test hashes match; pinned gofmt list is empty. HEAD, branch and index
retain their baseline values; final read-only Git diff-check exits 0.

## Test-driven implementation record

- Baseline full statematerials/stateverify: exit 0, 2.122s / 2.166s.
- Task 1 targeted RED exit 1: missing new API/guard helpers and exact verifier
  consumer rejection. Targeted GREEN exit 0. Independently full
  statecomposition/statematerials/stateverify exit 0: 2.448s / 2.765s / 2.861s.
- Task 2 targeted RED exit 1: missing selectBatch. First GREEN attempt exposed a
  test-fixture error: all-zero unknown share violated existing decodeFixed
  nonzero rule. Diagnosed and changed only the test fixture to a nonzero signature
  over an unrelated test transcript. Targeted GREEN 3.646s; full package 2.615s,
  both exit 0. Frozen vectors are read-only; no generator/vector was edited.
- Task 3 targeted RED exit 1: missing Evaluate/model/mapError/checkEvidence.
  Targeted GREEN 1.837s; independently full composition/materials/verifier
  2.196s / 2.077s / 2.102s, all exit 0.
- Task 4 targeted RED exit 1: missing Summary/status/obligation API. Targeted
  GREEN 4.661s, exit 0. All 99 observed Evaluate Context.Err checkpoints tested
  for both cancellation and deadline, plus final Recheck checkpoint. Caller
  unchanged under cancellation/panic. Retained owned aliases checked for zero;
  unreachable failure copies are not claimed externally observed.

Targeted selectors are the exact Task 1–4 selectors in the approved plan; Task 4
also used -v to record the 99 checkpoints. Full package commands use -count=1.
Task 4 full tests before probes: exit 0, 2.123s. Three independent probes each
failed as intended: omitted metadata selection guard (1.811s), upgraded
MaterialOnly rows (1.754s), omitted fresh whole-input comparison (3.394s; all four
fresh-valid replacement tests failed). Each source restored to exact SHA256:
selection.go `72DE2BFEDD510A394BDEC421D70C72FF6CE8162FB34C61989D016B4EC2315738`;
model.go `D8DE3C23610A2BAD371F2084DA3A6F2DB941C198973BB456891904AD91C20A7B`.
Restored full package exit 0, 3.091s. Formatting afterward only changes layout;
post-format source/test freeze contains 16 files.

Task 5 cleanup guard RED exit 1 (missing AST helper), then boundary/cleanup/
consumer/product-import tests GREEN exit 0: composition 2.715s, materials 2.787s,
verifier 2.804s. Pinned gofmt/list exit 0, list empty. Read-only HEAD/branch and
diff-check exit 0; baseline HEAD/branch/index unchanged; only two approved
baseline tests differ (416 unchanged). The first final-regression batch passed:
composition 6.387s, materials 6.428s,
format 4.444s, verifier 4.576s, each independent exit 0. All final commands are
now complete; the single fresh read-only independent review is complete.

Final command suffixes: `test ./internal/nodecontrol/<package> -count=1` for the
first ten full packages, `test ./internal/readiness -count=1`,
`test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1`,
and `test ./internal/store -run '^$' -count=1`.

| Package / scope | Exit | Reported package time |
| --- | --- | --- |
| statecomposition | 0 | 6.387s |
| statematerials | 0 | 6.428s |
| stateformat | 0 | 4.444s |
| stateverify | 0 | 4.576s |
| statepolicy | 0 | 3.452s |
| statecontract | 0 | 3.385s |
| staterecovery | 0 | 3.628s |
| statetime | 0 | 2.997s |
| serving | 0 | 7.824s |
| contracts | 0 | 8.791s |
| authority (selected) | 0 | 8.384s |
| readiness | 0 | 6.926s |
| store (compile only) | 0 | 5.547s; no tests |

The initial author invocation used nonexistent internal/nodecontrol/readiness
and exited 1 at setup, with no tests run. The directory/package was verified as
internal/readiness from repository files and the prior scoped evidence. Only
the path was corrected, with no source or scope expansion; the actual approved
package then passed. This failure remains recorded, not relabeled as GREEN.

Final validation preserves the post-format 16-file source/test freeze. Baseline
418-file check and HEAD/index are unchanged except the two approved parent
tests. The independent review is confined to this local increment, not the
many pre-existing changes or the full branch's acceptance blockers.

## Single independent incremental review and root disposition

Fresh gpt-6-astra/high reviewer read all 16 scoped files and relevant lower
contracts/guards. Verdict: ready for scoped incremental handoff; 0 Critical /
0 Important / 0 Minor, all five Review Focus cases satisfied. Independently
verified all 16 frozen hashes and baseline two-changed/416-unchanged split.
Reviewer executed no tests and made no writes/network calls/delegations.
Root accepts the verdict; no source/test fix or second review, no deferred minors.
Only documents changed after frozen tests/review. This is not main merge approval.

Final post-review handoff checks exit 0: all 16 freeze hashes matched, exactly
two approved parent test exceptions among 418 baseline files, unchanged index
SHA256/HEAD/branch, empty pinned gofmt list and read-only Git diff-check. The
13 final command records were reread, each exit 0. No tests were rerun merely
for documentation changes; the tested/reviewed source and tests remain identical.

Root explicitly rules on every declined behavior: anchor provenance/acquisition
independence remains external (risk: untrusted anchor mistaken for trust);
historical original authorization after coherent rewrite remains unproved
(risk: fresh success misread as historical authorization); unselected history
signatures/current authorization remains MaterialOnly (risk: unchecked history
promoted); current/trusted time and business policy remains external (risk:
expired/domain-invalid evidence used). These are not within-scope defects.

One-target/v1/max32 policy stays fixed (cost: future unsupported inputs rejected).
Full material closure/composition, physical commit and runtime authorization
remain open (risk: partial proof mistaken for full state). Concurrent mutation,
shallow-copy and concurrent Clear/access remain unsupported (cost: misuse can
race/invalidate state). Strings/unreachable runtime/detached caller copies are
not erased (risk: stronger erasure assumed). Prior dirty changes and lower-layer
redesign/security hardening are not reviewed (risk: unrelated existing defects
missed). Whole-project compatibility and all real DB/Docker/race/hardware/reboot/
C11/C12/PITR/Task10/B01/C12/main gates remain unverified (risk: unaccepted project
mistaken for production-ready). Full per-line rulings are retained in scratch
review-result.md and progress.md. No source action required for any declined line.

## Remaining obligations

External anchor provenance, unselected signatures and complete history, domain
semantics/trusted current time, full main-state material closure/composition,
physical commit and runtime authorization remain open even on success. No real
DB/Docker/race/hardware/reboot/C11/C12/PITR or whole-repository tests run here.
Store compilation is not store test acceptance. This incremental source review
will not be whole-branch/project security or main acceptance. Overall
Task 10/B01/C12/main gates and historical blockers remain unchanged.
