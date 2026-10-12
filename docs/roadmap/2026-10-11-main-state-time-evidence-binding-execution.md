# Main-state current-target time evidence binding — execution record

Date: 2026-10-11. Status: five-task local implementation and post-fix bounded
source review complete; final no-drift closeout recorded below. This is not main-state,
C12, production, whole-branch or integration acceptance.

## Approval and exact local scope

The human approved the written design and then this five-task implementation
plan on 2026-10-11. Native author execution in
D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator. No implementation
agents or per-task independent reviewers. One fresh gpt-6-astra/xhigh read-only
increment reviewer, with one corrective recheck by the same reviewer.

Branch: codex/c12-b01-task9-coordinator.
HEAD: 6752b9612a00b9d6b0bf87ae8dd95063f6b056ee.
Index SHA256: 1CBA46BEF4C1AFDC9AE7938146909F0D2D641A0D712CE18036B60FE5B8673D12.

Only twenty new Go files (ten product and ten test) in statetimebinding, and eight
approved old boundary_test.go files. No lower product/API/wire/algorithm,
dependency/lock, SQL/generated file, other existing tests or captured vector edits.
The initial 515-file baseline has exactly eight approved guard changes and 507
unchanged files. All 24 captured testdata files are unchanged; no new vectors.
Exact 28-file candidate formatting and hash manifests retained.

Implementation owns complete input, binds unchanged current root/metadata
sections and exact raw metadata in four positions, complete selected policy and
explicit time, declared common base state, request and old/signed-new checkpoints.
Summary is detached; all nested lower obligations and five added obligations stay
visible. Recheck performs full fresh Evaluate and exact normalized-input comparison,
including unselected and unprojected data and nanoseconds. Clear is bounded to
reachable owned mutable storage. No State, Ready, Commit, Authorize or consumption
capability, real provider or runtime consumer.

## Five task results and test-first evidence

Tasks 1–5 completed for the bounded local increment, with the explicit M1
historical recording limitation. Task briefs, baseline with genuine original
guard bytes, progress and actual tool outputs are retained only under
.superpowers/sdd/2026-10-11-main-state-time-evidence-binding/. Bash skill bookkeeping scripts were adapted to PowerShell
read-only checks and apply_patch writes under the approved PowerShell-only scope.
No shared scratch/configuration or Git write was used.

Every targeted command below is independently:
go test ./internal/nodecontrol/statetimebinding -run '<selector>'
-count=1 -timeout 120s -p=1. RED missing APIs are actual pre-implementation build
failures allowed by the plan, not missing dependencies or tool refusals.

| Stage | Exact selector | Actual exits |
| --- | --- | --- |
| Task 1 | ^(TestTimeBindingFixtureControls\|TestTimeBindingBudget.*\|TestTimeBindingOwnedAliases\|TestTimeBindingNormalization\|TestTimeBindingErrorMapping)$ | 1 missing APIs / 0 targeted |
| Task 2 | ^(TestTimeBindingScope\|TestTimeBindingUnsupportedSections\|TestTimeBindingMetadata\|TestTimeBindingAlternateMetadataEnvelope\|TestTimeBindingObservations\|TestTimeBindingTimeSnapshot\|TestTimeBindingCheckpoint)$ | 1 missing helpers / 0 targeted / 0 full |
| Task 3 | ^(TestTimeBindingEvaluate\|TestTimeBindingLowerRejections\|TestTimeBindingSummary\|TestTimeBindingRequestSemantics\|TestTimeBindingSummaryOwnership\|TestTimeBindingModelClear)$ | 1 missing Model/Evaluate / 0 targeted / 0 full |
| Task 4 | ^(TestTimeBindingRecheck\|TestTimeBindingObservationRecheck\|TestTimeBindingFixedTime\|TestTimeBindingCancellation\|TestTimeBindingLateCancellation\|TestTimeBindingPanicPropagation\|TestTimeBindingCleanupRegistration\|TestTimeBindingProductImports\|TestTimeBindingConsumers)$ | 1 missing Recheck/AST helper / 0 targeted / 0 full |

Full current-package commands in tasks 2/3/4 are independently:
go test ./internal/nodecontrol/statetimebinding -count=1 -timeout 120s -p=1.
They exit 0 with Go-reported elapsed times 2.887s, 3.035s and 4.597s.
Targeted GREEN times for tasks 1/2/3/4 are 2.325s/2.336s/2.772s/3.855s.
The early task receipts do not record original UTC start/end or per-stage hashes;
those values cannot be reconstructed and are not claimed. Tool wait duration
is not test elapsed time. This is an explicit recording deviation, not a complete
historical process audit (see author disposition M1 below).

Eight consumer commands each failed a genuine path/package assertion before the
precise exception and then passed. The first statetime RED used the unintended
name TestTimeTimeBindingBoundary; the approved TestTimeBindingBoundary replaced it
before GREEN. All eight commands used the same suffix selector; final names are
the approved ones, not an expanded capability exception.

| Existing consumer package | Initial RED exit | GREEN exit |
| --- | --- | --- |
| statepolicybinding | 1 | 0 |
| statetime | 1 | 0 |
| statebinding | 1 | 0 |
| statepolicy | 1 | 0 |
| stateverify | 1 | 0 |
| statecontract | 1 | 0 |
| statematerials | 1 | 0 |
| statecomposition | 1 | 0 |

Exact eight independent commands (also repeated post-format, all exit 0):

    go test ./internal/nodecontrol/statepolicybinding -run 'TimeBindingBoundary$' -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/statetime -run 'TimeBindingBoundary$' -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/statebinding -run 'TimeBindingBoundary$' -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/statepolicy -run 'TimeBindingBoundary$' -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/stateverify -run 'TimeBindingBoundary$' -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/statecontract -run 'TimeBindingBoundary$' -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/statematerials -run 'TimeBindingBoundary$' -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/statecomposition -run 'TimeBindingBoundary$' -count=1 -timeout 120s -p=1

Additional public/ownership/time/budget controls were included in the fresh full
matrix, not inferred from earlier component results. The supplemental targeted
receipt records exit 0 / Go time 2.373s; its exact selector was not emitted in
that retained receipt and is not reconstructed here. This is part of the early
command-recording limitation, not a substitute for the exact final full-package
verification. See format-and-supplement.json for the actual output.

## First independent review and author rulings

First source review: 0 Critical, 1 Important (I1), 1 Minor (M1).
The reviewer read the complete approved design/plan, all twenty new files, actual
baseline-to-current differences for all eight old guards, all phase receipts and
first final matrix/freeze. The five mandatory focus conclusions cover the
correct first-byte-2 signing seed; independently owned raw aliases and separately
lower-valid same-P/different-E controls; whole normalized root/metadata sections;
replacement request IDs/nonce versus fixed Recheck and legal repetition; early
cleanup registration and dynamic cancel/deadline/panic traces. No tests or external
diagnostics were run by the reviewer. The reviewer independently verified all
28 candidate and 515 existing hashes and the index. Its Git branch/HEAD read was
refused by Git ownership checks; the author performs permitted read-only Git
checks without configuration changes. An initial reviewer Hash-vs-SHA256 comparison
error was corrected, not treated as actual drift.

I1 accepted and fixed within the new package. A valid selected-policy fixture
with Time.Metadata replaced by the zero observation but unchanged raw envelopes
previously returned BindingMismatch before availability could return
AssuranceMissing. It remained fail-closed; no erroneous success was identified.
The author first added public TestTimeBindingMissingObservationPrecedence and
observed the intended assertion RED exit 1 / Go time 2.835s. The test covers all
four zero observations and Unknown/Unavailable/Rollback/out-of-range statuses,
including contradictory metadata coordinates, with selected-policy success
controls. It also checks selected-policy MissingHistory still wins and callers
are unchanged. Only then were four explicit Available switches added before
bindMetadata interprets coordinates. Scope and complete selected-policy ordering,
existing bindObservations checks and lower algorithms are preserved. Targeted
GREEN exit 0 / Go time 1.893s. Only binding.go and evaluate_test.go changed since
the first review; exact 28-file reformat/check passed, then all sixteen commands
below were freshly rerun. The same reviewer explicitly reread both changed files,
the new regression, RED/GREEN and all revised final receipts, independently
recomputed 28/515 hashes and verified that only these two files changed. Verdict:
I1 resolved, no new Critical/Important/Minor. First full review plus this explicit
two-file correction recheck cover the revised freeze, not any later source edits.
Final documents were not reviewed as completed; author closeout covers them.

M1 accepted as an audit limitation, not silently fixed. Original early UTC
receipts/per-step source freezes were not emitted, so no invented reconstruction,
complete historical audit claim, or claim of full original recording compliance.
Actual early exits/output plus final frozen UTC receipts are preserved. Cost:
the early process sequence cannot be independently timestamped/historical hashes
cannot be proved; fresh final pass is the current verification, not retroactive
evidence. No unresolved product defect is inferred from that limitation.

## Actual mistakes and non-RED attempts

All attempts remain disclosed; these do not count as behavioral RED or passing tests:

- Early baseline code used Hash for a prior SHA256 manifest, falsely reporting
  drift; corrected before editing. Missing-source/script-path reads were resolved
  from actual files, not treated as tests.
- Two guard apply_patch constructions were rejected (unprefixed newline and
  non-forward hunk order), with no file change. A Bash-style brace expansion in
  a PowerShell read command was a parser error, not a source result.
- Task 4 targeted command initially exited 1 from test-helper redeclaration
  (boundary_test.go:415); corrected helper variable. A subsequent exit 1 came
  from a guard confusing explicit observation.Now fields with time.Now; narrowed
  the test guard to the approved import/member policy, retaining clock denial.
- Supplemental targeted command exited 1 because reflect import was removed
  before its leftover tautology; removing that tautology produced exit 0.
- The initial unexecuted I1 regression patch used an incorrect TimeObservation
  type name; the actual input interface was read and corrected before running.
- The post-fix freeze precheck compared slash/backslash paths literally and
  falsely listed eight approved guards as unexpected. Normalized comparison
  proves no other drift; it is not a test or product failure.
- Two post-fix command-label quoting attempts (authority/store) exited 1 with
  PowerShell parser errors before tests started. The label quoting was corrected;
  selectors, environment and 120s budgets were unchanged. Their actual test
  sessions then independently exited 0. No pending test was re-launched to mask
  output delay. Original failure outputs are retained in the post-fix matrix.

## Fixed offline environment and actual sixteen-command matrices

Every test process used PowerShell/login=false and:
GOENV=off; GOWORK=off; GOTOOLCHAIN=local; GOPROXY=off; GOSUMDB=off;
GOAUTH=off; GOVCS=all:off; GOFLAGS=-mod=readonly.
GOPATH=C:/Users/Lenovo/go; GOMODCACHE=C:/Users/Lenovo/go/pkg/mod.
GOCACHE=D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/.superpowers/sdd/2026-10-11-main-state-time-evidence-binding/go-cache.
Go/gofmt binaries:
C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/.
No downloads or permanent environment/configuration changes.

The following sixteen exact independent commands were run both before the first
review and freshly after I1's fix. All actual sessions ended with exit 0.
Fourteen full packages, one selected authority test set, one store compile-only
(no store tests run). No full-repository/race/DB/Docker/C11/C12/PITR command.

    go test ./internal/nodecontrol/statetimebinding -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/statepolicybinding -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/statebinding -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/statecontract -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/statematerials -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/statecomposition -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/stateformat -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/stateverify -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/statepolicy -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/staterecovery -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/statetime -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/serving -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/contracts -count=1 -timeout 120s -p=1
    go test ./internal/readiness -count=1 -timeout 120s -p=1
    go test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -timeout 120s -p=1
    go test ./internal/store -run '^$' -count=1 -timeout 120s -p=1

Go-reported times are test output, not command wall time. UTC start/end below are
the actual post-fix process receipts, not reconstructed historical timestamps.
First matrix UTC envelope: 2026-10-11T12:06:01.9903171Z to
2026-10-11T12:07:35.3509538Z. Detailed actual receipts for both matrices are in
final-matrix.json and post-review-fix-final-matrix.json.

| # | Scope | First Go time / exit | Post-fix Go time / exit | Post-fix start UTC | Post-fix finish UTC |
| --- | --- | --- | --- | --- | --- |
| 1 | internal/nodecontrol/statetimebinding | 19.843s / 0 | 3.305s / 0 | 2026-10-11T12:15:55.0387349Z | 2026-10-11T12:16:21.3115291Z |
| 2 | internal/nodecontrol/statepolicybinding | 22.336s / 0 | 3.519s / 0 | 2026-10-11T12:15:55.4809108Z | 2026-10-11T12:16:21.9538992Z |
| 3 | internal/nodecontrol/statebinding | 17.496s / 0 | 2.521s / 0 | 2026-10-11T12:15:55.4856351Z | 2026-10-11T12:16:18.7886005Z |
| 4 | internal/nodecontrol/statecontract | 7.095s / 0 | 0.881s / 0 | 2026-10-11T12:15:55.5219006Z | 2026-10-11T12:16:30.1003220Z |
| 5 | internal/nodecontrol/statematerials | 1.474s / 0 | 0.988s / 0 | 2026-10-11T12:15:55.5228427Z | 2026-10-11T12:16:17.3285884Z |
| 6 | internal/nodecontrol/statecomposition | 12.659s / 0 | 1.124s / 0 | 2026-10-11T12:15:55.5791371Z | 2026-10-11T12:16:18.3484909Z |
| 7 | internal/nodecontrol/stateformat | 6.919s / 0 | 0.933s / 0 | 2026-10-11T12:15:55.6279226Z | 2026-10-11T12:16:29.6999561Z |
| 8 | internal/nodecontrol/stateverify | 11.390s / 0 | 1.497s / 0 | 2026-10-11T12:15:56.3979585Z | 2026-10-11T12:16:31.5236158Z |
| 9 | internal/nodecontrol/statepolicy | 6.276s / 0 | 1.506s / 0 | 2026-10-11T12:15:56.3151131Z | 2026-10-11T12:16:32.7857300Z |
| 10 | internal/nodecontrol/staterecovery | 7.104s / 0 | 0.919s / 0 | 2026-10-11T12:15:56.3234113Z | 2026-10-11T12:16:29.6840200Z |
| 11 | internal/nodecontrol/statetime | 7.204s / 0 | 1.345s / 0 | 2026-10-11T12:15:56.4286859Z | 2026-10-11T12:16:31.5199071Z |
| 12 | internal/nodecontrol/serving | 10.593s / 0 | 1.853s / 0 | 2026-10-11T12:15:56.5253916Z | 2026-10-11T12:16:38.4323928Z |
| 13 | internal/nodecontrol/contracts | 4.281s / 0 | 2.334s / 0 | 2026-10-11T12:15:56.6322966Z | 2026-10-11T12:16:39.1625167Z |
| 14 | internal/readiness | 6.864s / 0 | 2.504s / 0 | 2026-10-11T12:15:56.7375899Z | 2026-10-11T12:16:39.2882705Z |
| 15 | internal/nodecontrol/authority | 10.288s / 0 | 2.092s / 0 | 2026-10-11T12:16:14.2489448Z | 2026-10-11T12:16:40.8252336Z |
| 16 | internal/store | 4.689s / 0 | 1.159s / 0 | 2026-10-11T12:16:24.7070419Z | 2026-10-11T12:16:39.9733035Z |

## Reviewer exclusions — explicit author dispositions

These exclusions are not approvals; each remains bounded as follows:

1. Physical metadata/base-state provenance: caller-declared only; common StateID
   proves model correspondence, not a co-capture or physical commit.
2. Trusted clock/TLS/HTTP Date/nonce randomness: external responsibilities;
   signature does not supply Now/floor, seed fixtures do not prove provider truth.
3. Atomic nonce consumption/replay prevention: intentionally not implemented;
   zero nonce and repeated successful Evaluate remain legal model behavior.
4. Durable checkpoint writes/guard seals/crash or reboot recovery: no actual I/O
   or commit authority; binding declarations is not a write or durability proof.
5. Root/metadata installation/fresh install/historical target/floor changes:
   rejected unsupported shapes; require a future separately approved design.
6. Complete historical/unselected authentication and other-domain/runtime
   authorization: all lower obligations retained; selected evidence is partial.
7. Concurrent mutation/shallow owner copies/concurrent Clear: unsupported caller
   use; no concurrency or copy-safety guarantee and no added synchronization.
8. String/unreachable/runtime-copy erasure/OOM/fatal termination: excluded; only
   reachable mutable owner storage has direct Clear observations. Private failed
   owner erasure is early-registration source evidence, not directly observed.
   Fixed AST guards are not a formal/general program proof.
9. Full-repository/race/DB/Docker/hardware/C11/C12/PITR acceptance: unexecuted and
   prohibited here, existing gates unchanged; production compatibility UNKNOWN.
10. Prior unrelated dirty work/Git integration/production readiness: preserved,
    not independently reviewed or accepted; no staging/commit/push/merge/reset/
    deletion/worktree cleanup, no new cross-computer recovery checkpoint.
11. Missing early execution receipts: M1 limitation explicitly retained; fresh
    final UTC/freeze evidence cannot reconstruct the early historical process.

## Reviewed post-fix candidate hashes

The first freeze is preserved separately; the corrected candidate is
post-review-fix-source-test-freeze.json. At 2026-10-11T12:17:03.3706504Z all
28 candidate and 515 existing hashes still matched it, index unchanged.
The corrective reviewer independently confirmed this revised freeze. The final
author closeout rereads source/guards/vectors, documents and Git/index after
documentation; the actual receipt is final-closeout.json in the same dedicated
evidence directory. No source edits are allowed after this reviewed/tested freeze.

| Candidate path | SHA256 |
| --- | --- |
| internal/nodecontrol/statetimebinding/doc.go | DF8D11E8A1D99602BA52831F5A9F9BA7A0CF132BD6752B8D24EB81E64C88CFC9 |
| internal/nodecontrol/statetimebinding/input.go | 028F8BFC40DE37A9420BAB7E9BBB9C76C9D0CE672AEE2D3B6186AC15D77647FB |
| internal/nodecontrol/statetimebinding/errors.go | 88D71E48ADD88957DB27CBB9C2748206AF1785A0C905DD5A73888AAE00DF8BD6 |
| internal/nodecontrol/statetimebinding/budget.go | 50791D8FD8441085A71798B1F0E78AD54F351882EA35564EC9D2D1C40F91C861 |
| internal/nodecontrol/statetimebinding/owned.go | 5E5D0C59346B33AD0937B28D2306AE10127A126792034C6558AD215C6AEDF07F |
| internal/nodecontrol/statetimebinding/binding.go | B7168814B5B1E241B61DD31A522D62B04C36CD960FF99E3AE69A86F3257B13F4 |
| internal/nodecontrol/statetimebinding/observations.go | 9AFD801EEFC2DD339411EA5B73C9891C2CFB62616CE111793BE4F5D595B21378 |
| internal/nodecontrol/statetimebinding/checkpoint.go | CE8D7F3AD1B8DBFD58415AC99E8C7CE3F91202E59AB2C22694630D30159A8B18 |
| internal/nodecontrol/statetimebinding/evaluate.go | C296F226E4D767476B6F6264E34D1BD1E1ADA4980D9CC780901881CD619C3A2F |
| internal/nodecontrol/statetimebinding/model.go | D1281BD603EDF699BE7AEF9BEBE22A1A61406FF519F1E80E70228B0747CBB717 |
| internal/nodecontrol/statetimebinding/fixtures_test.go | EA9150CB83FE4BB993547FC1B168FA46A607458D585B450512C0EF51A956126F |
| internal/nodecontrol/statetimebinding/budget_test.go | E426474BEACD3A868763733A1D3DBFE342C259B14EBB10AC88B16BB4A2F8C392 |
| internal/nodecontrol/statetimebinding/owned_test.go | F2CC2C1AB997AE1609671D60125313CFE579D965E02757FF870FF7C5B4795F5E |
| internal/nodecontrol/statetimebinding/errors_test.go | 541F30DA229F61B8C403AD67E658C847356672974C1260E922689293512E37B1 |
| internal/nodecontrol/statetimebinding/binding_test.go | 1DDAE10C4ECE757574E30A04789F35A9BB9E45C1DB8DF2714E5246E4B69E4581 |
| internal/nodecontrol/statetimebinding/observations_test.go | 65760EC0E83C172F9747C31ED13D8EFB1C8D37F5230F65155A459930832E18AB |
| internal/nodecontrol/statetimebinding/checkpoint_test.go | 0A3897AF845471710680C5DEEA84AEAB4A5FEDB674F2DE0C026624DD3CE9EA14 |
| internal/nodecontrol/statetimebinding/evaluate_test.go | 79E426EFA59D52DA3CAB5C049FB99ACF43C2566BCAE631B401A3ECD52CA48BF4 |
| internal/nodecontrol/statetimebinding/model_test.go | ABC77F786453E8EBF6E7466D5F78D48BF288765358E74F1972CB524B1BF75030 |
| internal/nodecontrol/statetimebinding/boundary_test.go | 5BB29A8B2825E3CE51ECC67C82D8D4581DD5195E4AEE9FB363E12D0579B33906 |
| internal/nodecontrol/statepolicybinding/boundary_test.go | D17FFDB9515CE8F85BC37050EB62F4FD12F54626B1DE7E04C9F918AAFAE9C724 |
| internal/nodecontrol/statetime/boundary_test.go | EFDBFACC8CB6EAF6B7360A547DF46F05B66620EC19E057F3C1A5C0472726FF8B |
| internal/nodecontrol/statebinding/boundary_test.go | 6464E5185E11C48CF38BC3E15209F42920453AAC7A6E9FE337D5ADEB3ADBE94A |
| internal/nodecontrol/statepolicy/boundary_test.go | 1521D823F0E2C104B48F367A3FA3043847B271DCB43E65EC27C50B1E3FEBF2FF |
| internal/nodecontrol/stateverify/boundary_test.go | 533D53AC2FDDC437C2CE335361134D4F044F776BFB1428B69536E3862DD5FA02 |
| internal/nodecontrol/statecontract/boundary_test.go | 64E839E666470DB1CCBA6FCBCFE776880DAE6393400506D25EC99081871C1E4D |
| internal/nodecontrol/statematerials/boundary_test.go | 130AAEC7CAC38A72EEEE8C6C77D4A85CCB8DB3778FBE5FB9D129CC4D89A01FE0 |
| internal/nodecontrol/statecomposition/boundary_test.go | 8BDC290F2DA5D7A1098D63A785EC74CEFB1898292DF728FE348BA4B4F31BB1DF |

## Handoff limits

Only local implementation/evidence is delivered. It is not committed or pushed;
another computer cannot obtain these new local changes from the current remote.
Existing whole-project/C12/main acceptance blockers remain open. No side-effect
implementation, actual checkpoint commit, atomic request consumption or runtime
authority was added. Implementation skills determined test-first ownership,
exact bounded verification and one fresh read-only incremental review; they do
not grant production or Git integration approval.

