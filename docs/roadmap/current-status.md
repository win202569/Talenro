# Project status — 2026-09-08

The original snapshot was checked against the implementation and plans at
`2668b901`, followed by the repository-portability repair. This branch also
records the Task 9 core checkpoint below. Implemented components are distinct
from the acceptance criteria of their containing stage.

| Stage | Status |
| --- | --- |
| Foundation | Completed development acceptance; see the implementation sequence. |
| C1.1 account, device identity, configuration trust | Completed development acceptance. The delivered bundle is a test configuration, not a production tunnel configuration. |
| C1.2 node and POP control plane | Partially implemented within Batch 01. Contracts, schema, authority coordination, migration guards, and test infrastructure exist; full Batch 01 acceptance has not closed. |
| C1.3 entitlements, traffic ledger, quota | Planned; business implementation pending. |
| C1.4 scheduling, tunnel credentials, production distribution | Planned; implementation pending. |
| C1.5 five-POP composed fault acceptance | Pending the preceding stages. |
| Four-platform Tunnel Engine | Designed; native client implementation pending. |
| Payments, subscriptions, refunds | Planned; implementation pending. |
| Referral, affiliate, rewarded ads | Designed; implementation pending. |
| Commercial UI, localization, packaging, stores | Designed; implementation and release acceptance pending. |

The merged **C12 Secure Execution Architecture** concerns the Windows test
execution controller. Its completion does not mean C1.2 or the VPN product is
complete. The repository currently has no node-agent or node-core-supervisor
command, production node API implementation, or Xray/sing-box runtime adapter.

## Repository portability repair

Commit `91e47661944357f5dd5c6499036b0fc549629ca5` integrates the former local
repository compilation fix and removes pre-existing overlay dependencies from
tests. Authority, store, and migration package tests pass without the overlay.
The sealed factory probe also passes with race instrumentation enabled. An
independent review found no Critical or Important issues in this repair.

A clean clone exposed two further portability issues: CRLF checkout changed the
pinned embedded Go-source digest, and whole-second watchdog rounding discarded
almost one second of the native operation budget. The follow-up normalizes only
CRLF to LF before source verification and uses remaining milliseconds without
extending production deadlines. Regression coverage retains source-tamper
rejection, native descendant termination, and exact resource cleanup. This
follow-up is committed as `16c4e8c4e43ae893aad8779639bec0f80ca096cb`; its
independent review is closed with no remaining blockers.

On 2026-09-08, a new independent clone of that commit, in a path containing
spaces and with no pre-existing `.superpowers` directory, passed:

```text
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GOFLAGS="" GOENV=off GOWORK=off
go test ./... -count=1 -p=1 -timeout=60m
```

The command exited 0; `internal/testinfra` completed in 663.533 seconds. This is
the Windows untagged repository suite, not live Docker/PITR or production
acceptance. The code checked here is reproducible without the old local source
overlay; dependencies and tools must still be installed on a new computer.

## Next implementation work

1. Close the remaining [Batch 01](../superpowers/plans/2026-08-23-node-pop-control-plane-01-contracts-schema-authority.md)
   implementation and acceptance gaps in dependency order: Task 9 Coordinator
   integration, Task 10 guarded serving, Tasks 11–14 complete protocol contracts,
   then Tasks 15–18 exact-epoch recovery and final acceptance.
2. Proceed through the remaining [C1.2 batches](../superpowers/plans/2026-08-23-node-pop-control-plane.md),
   including inventory/operator state, node identity and mTLS, agent, supervisor,
   observations, core adapters, and final operational acceptance.

Clean-checkout reproducibility is now verified as described above; follow
[repository recovery](../runbooks/repository-recovery.md) on another computer.

The [implementation sequence](implementation-sequence.md) preserves historical
acceptance records. Its C1.2 `current design` label predates the current partial
implementation; it is not a declaration that C1.2 has passed final acceptance.

## First Batch 01 delivery unit

Task 9 is in progress on `codex/c12-b01-task9-coordinator`. Core checkpoints
`e0a387f3` and `1ace2bf7` replace the old arbitrary resolver constructor with the
exact sealed dispatcher, commit a durable claim before provider Abort, and connect
activation and stored-outcome recovery to the canonical evidence APIs. This is
not complete Task 9 or Batch 01 acceptance, and is not a production serving gate.

The checkpoint includes three independently reviewed repository corrections:
`1bbe83a9` fixes the checkpoint/time material branches, `e1e9c8a1` enforces their
reason-specific matrix, and `8e8a34d2` requires domain absence for an aborted
fence while retaining the domain-outcome requirement for a committed fence.

An independent clean linked worktree at `1ace2bf7`, containing the final ordinary
Coordinator/readiness code, passed:

```text
go test ./internal/nodecontrol/authority ./internal/readiness -count=1 -timeout=10m
go test -race ./internal/nodecontrol/authority -run 'TestCoordinator|TestActivationAdmission' -count=1 -timeout=5m
```

The complete authority/readiness packages finished in 12.725s / 2.954s; the
specified race check finished in 2.549s. Both commands exited 0. Ordinary commands
used Windows/amd64, CGO=0, empty GOFLAGS and GOENV/GOWORK off; the race check used
CGO=1 and the existing MinGW GCC. Independent core protocol review is closed
with no remaining Critical, Important, or Minor issues. Recovery mutation tests
now compare independent domain inputs and terminal context, and distinguish a
failed Commit before persistence from a lost response after persistence.

One full ordinary repository run also passed without changing its timeout:

```text
go test ./... -count=1 -timeout=10m
```

It exited 0, including authority 13.695s, readiness 3.522s, store 5.162s, and
testinfra 541.302s. Subsequent fixture-only commits do not change this untagged
code. This is not execution evidence for PostgreSQL or physical PITR tests.

Claim-v1 PostgreSQL crash/atomicity and two-connection fixtures are committed
at `cdb9697d`, with a timestamp-precision correction at `cc88e916`. They are
statically reviewed, not live database accepted. An independent
integration-tag compilation check at `cc88e916` exited 0 for authority/readiness
(0.338s / 1.890s, no tests executed), using:

```text
go test -tags=integration ./internal/nodecontrol/authority ./internal/readiness -run '^$' -count=1 -timeout=5m
```

The earlier whole-branch review found no Critical issues and one Important
legacy PITR integration regression, described below. The controlled-interface
correction and final decoder fix now have independent source-review approval and
fresh ordinary verification, as recorded below. The branch remains an explicitly
unaccepted development checkpoint, not ready to merge into main: physical PITR
acceptance and the historical validation finding remain open.
Docker is unavailable on this host, and the repository has no existing GitHub
workflow to provide a replacement live run. The user approved the
[controlled PITR interface addendum](../superpowers/specs/2026-09-19-c12-controlled-pitr-test-interface-design.md)
on 2026-09-19. Its [six-task implementation plan](../superpowers/plans/2026-09-19-c12-controlled-pitr-test-interface.md)
was subsequently approved for subagent-driven execution with “是” on 2026-09-19.
Tasks1–5 are implemented through `e3b5ed93726dd0b7d02d8a609ae24963e48087fb`;
Task6 registration and portable evidence are included in the commit introducing
the execution ledger below. Cleanup remains runner-owned. The plan now records
actual commands/results, source hashes, ordered root rulings R1–22 and explicit
user exception U1; its original planned gates are not execution evidence.

The preceding review's Important legacy integration finding was that
`TestPITRBeforeRevocationFailsClosed` supplied the in-memory
`coordinatorEffectResolver` to the new sealed test dispatcher while using real
PostgreSQL transactions. That source implemented neither the registered
transactional resolver nor activator; the handler rejected this DBTX path.
The legacy schema could also fail before that point. Task5 replaced that harness
with real PostgresRepository/registered-handler/controller P/A/B transactions,
preserved the top-level name, and received independent source review approval
for a development checkpoint. Task6 closes exact PITR-only runner registration.
This is source remediation now covered by final whole-branch review and its scoped
fix rereview, not a claim that the unavailable physical gate passed or that the
checkpoint has received acceptance approval.

The existing authority-v7 runner was attempted without changing its guards or
budgets. Initial launches rejected inherited `GIT_*` variables. A separate child
process with only those inherited keys removed advanced past that guard, then
exited 1 with `Go graph go mod verify timed out after its bounded native wait`.
No database test ran. Docker's absence is a separate known environment limitation,
not the reported termination point of this attempt.

Existing-fence/lost-claim recovery is in scope. Reconstructing an entirely lost
claim-v1 fence remains fail-closed because the current contracts do not provide
its exact authoritative ProtocolActivationID. The activation/runtime/lease-aware
v7 readiness gate is owned by Task 15 and its staging-state acceptance by Task 18;
ordinary Task 9 readiness must not be used as a substitute.

Task 10 has no `WithConsistentReadyRead`, authorized-certificate projection, or
active desired-state projection yet. The migrated test-local PITR read checks readiness
and reads a fixed test snapshot; it does not provide the required same-connection
readiness/query/readiness contract for those real projections.

Later tasks should reuse the existing private canonical/envelope/evidence
helpers, while completing the public schema/role registry and exact-epoch
projection. The current ordinary checkpoint query still selects
`MAX(authority_epoch)`; that is not the planned activation/terminal-chain
projection. Finally, `-Suite batch01` requires the tracked
`testdata/c12/integration-contracts-schema-authority.v1.json`, which is absent.
The runner currently rejects that missing manifest before starting test groups;
the manifest must be derived from the completed required integration cases.

## Controlled PITR implementation checkpoint — 2026-09-19

Branch: `codex/c12-b01-task9-coordinator`, isolated worktree
`D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`. Plan implementation base:
`d8b8e01483ae5e7fffa5dfb76b4f847ab95740de`. Task source checkpoints are Task1
`29e03bc8`, Task2 `816fa7fa`, Task3 `374a8098`, Task4 `0cb6c720`, Task5
`e3b5ed93726dd0b7d02d8a609ae24963e48087fb`, and the Task6 commit containing this
ledger (`test(c12): close controlled PITR gates and record evidence`). The Task6
handoff originally awaited independent review; the subsequent whole-branch
findings and bounded fix verification are recorded below. Scoped final rereview
is complete; authorized development-branch push is pending at this documentation
commit. No PR, main merge or worktree removal is claimed.

The [portable execution record](../superpowers/plans/2026-09-19-c12-controlled-pitr-test-interface.md#execution-record--2026-09-19-development-checkpoint-not-acceptance)
contains the actual approval/execution checklist, exact commands/results, source
hashes, complete ordered root decisions R1–22 with reasons/costs, explicit user U1,
and all 13 separate physical gate commands. It does not require local scratch to
identify what was implemented, tested or left unaccepted.

Task6 actual PowerShell registration RED failed before implementation at the
first resource boundary under the wrong profile; the final exact registration/API
pair passes (2.473s). Exact no-Docker behavior passes (testinfra2.799s,
authority bridge0.318s); integration compile-only passes (0.324s/0.367s, no tests
executed); direct native race passes (6.337s/1.384s, no race report); ordinary
authority/readiness passes (10.336s/2.822s). `-Race` in the physical runner is still
not forwarded and provides no race evidence.

All source/tests were frozen before the one final ordinary command:

```powershell
$env:GOOS='windows'; go test ./... -count=1 -timeout=15m
```

Session81642 exited0; all packages passed, testinfra587.277s. Three changed
source/test hashes matched after completion. Documentation was appended afterward,
not included retroactively in that source freeze. Task5's preceding frozen full
run26567 also passed (testinfra563.456s). Neither success erases prior failures or
establishes their cause: Task4 frozen80961 failed its transient-validator150s
helper and testinfra15m limit (903.139s); unchanged R21 run26986 failed platform
cancellation1.2828222s>1s and prepared oversized-frame CreateProcessW Win325
(testinfra583.287s). The unchanged exact transient-validator rerun passed130.82s,
which is not a full-suite pass or cause diagnosis. The earlier baseline/Task1/Task3
failure and focused-rerun history remains in the linked record.

User U1 explicitly chose “继续第 5–6 项，保留验收阻塞”: continue development and
push only an unaccepted branch checkpoint, without merging main. Task4's Important
validation finding remains OPEN after final review; this is not a general waiver of
new source findings, physical gates or final reporting.

Physical precheck again found no Docker executable; its daemon, endpoint/image
identity, restored candidate behavior and live cleanup could not be checked.
Go/PowerShell/Git are available; optional Go-tool goose lookup was interrupted
without a result, so dependency availability is not claimed. Inherited GIT_* keys
also require the existing legal clean-child launch. No physical runner was invoked,
guard removed, Docker installed or deadline expanded. The seven PITR calls
(two public consumers and five private seam values) plus six preserved authority-v7
crash calls are **unavailable/not executed**, not failed/skipped/passed tests. Each
retains a separate5m budget. Required explicit go-json PASS, foreign-canary/exact
cleanup, real P Ready/A/B terminals, A-present/B-absent restore, timeline advance,
precise B-provider denial and same-candidate A-provider Ready remain unverified.

The legacy authority-owned Docker/raw-pool/migration/cleanup harness was replaced,
not retained as fallback. Its unrelated fresh-cluster physical branch was removed;
the retained `TestAuthorityReadiness` identity-mismatch unit test is **not equivalent
physical coverage**. R17 transparently corrects the original plan's inclusive-on
record-end defect to off/false with PostgreSQL18.4 source citations. R20 retains a
private recovery-safe identity query, without new consumer SQL/grants. Task9/B01
is still unaccepted; Task10 certificate/desired-state serving and reconcile
recovery remain undelivered.

### Whole-branch review and bounded final source fixes

The independent whole-branch review of `e1571f9f..3998d5c9` identified Important
I1 (materialized results retained a shared mutable driver decoder), Important I2
(all13 physical gates unexecuted), Important I3 (the historical Task4 ordinary
validation finding), and Minor M1/M2 (an ineffective uncertain-Commit read
assertion and a raw-driver guard that missed function-value aliases). No Critical
finding or further production Coordinator/repository/readiness defect was found.

The single final fix wave based on `3998d5c9` addresses I1 and both Minors only:
each result keeps its independent fixed pgx decoder and reserves4KiB of the
existing shared64MiB budget before allocation; no Scan acquires backend ownership.
The stable-map regression exercises two legal materialized GetAuthorityFenceHead
rows separately through Access and Transaction, then barrier-starts their first
pg_lsn scans. Pre-fix assertions proved shared retained state and missing decoder
charging; an actual pre-fix race was **not** reproduced. The corrected concurrent
tests pass under the direct native race detector. M1 now requires a non-error
materialized read and exactly one driver-query delta after uncertain Commit.
M2 checks imported selector references against the existing small symbol inventory,
with real-guard alias-negative and approved-symbol-positive controls.

Focused final-fix verification: new regression RED0.382s, guard-alias RED1.693s;
GREEN new regressions0.400s, ordinary API/selector guards1.421s, approved controller
behavior/registry2.793s, direct new-regression race3.830s, broader scoped race4.385s,
pure bridge0.339s/race1.381s, and integration compile-only authority0.343s /
testinfra0.371s. All GREEN commands exited0. The existing codec, NULL/empty-byte,
expiry, eager-close and Commit-boundary tests are included in the covering gates.
Exact commands and decoder-cost rationale are in the linked portable record.

The fix is committed as `b1e1d7c209bd63acd01f3ee00ba06eb765044990`. Independent
scoped rereview verified I1/M1/M2 **ADDRESSED**, with no new Critical, Important,
Minor or out-of-scope source finding. It independently checked the pinned decoder
allocation bound, both Access/Transaction regressions and unchanged Commit ordering.
Source review is closed; acceptance is not approved.

Root then completed the fresh frozen ordinary command shown above on that exact
clean commit: session5472 exited0, all packages passed, including testinfra469.316s,
authority17.181s, readiness4.484s, trust4.228s and trustclient1.304s. The run began
at2026-09-19T18:20:11+04; post-run verification at18:28:56+04 found all seven source/
test/document SHA256s unchanged and the same clean HEAD. No source/test/document
edits or other test runs occurred during it. Subsequent commits update only the
portable evidence. Ordinary81642 remains prior-tree evidence, while5472 covers
the final source fix. Neither run diagnoses the historical failures.

I2 and I3 remain OPEN acceptance
blockers under U1. All13 physical calls remain unavailable/not executed, and the
historical failure causes remain unproved. The review's excluded Task10/15/18,
B03 provider/decoder, lost-fence identity, cross-process restart, hostile same-process
sandboxing, removed fresh-cluster coverage, runner race forwarding and wholesale
runner audit remain outside this delivery; none is claimed implemented or accepted.
