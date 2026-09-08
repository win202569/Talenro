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

Final whole-branch review found no Critical issues and one unresolved Important
legacy PITR integration regression, described below. The branch is an explicitly
unfinished checkpoint, not review-approved or ready to merge into main.
Remaining work includes this correction, physical PITR integration, and live
execution of the required gates. Docker is unavailable on this host, and the
repository has no existing GitHub workflow to provide a replacement live run. The runner-owned
PITR controller's external-package consumer interface is incomplete: authority
tests cannot observe their actual transaction commits or inspect candidate
domain rows through the current public contract. A controller-bound interface
extension requires approval outside the six Task 9 files; cleanup must remain
runner-owned. The authoritative `authority-v7-pitr` gate is listed in Task 10.
The older self-owned Docker harness is not equivalent acceptance and has not
been expanded to work around this boundary.

There is also a known legacy integration regression, established by static
review: `TestPITRBeforeRevocationFailsClosed` still supplies the in-memory
`coordinatorEffectResolver` to the new sealed test dispatcher while using real
PostgreSQL transactions. That source implements neither the registered
transactional resolver nor activator; the handler rejects this DBTX path.
Its legacy schema may fail before that point. The current wiring therefore does
not preserve runnable physical-restore/readiness coverage, even though the test
source is retained and compiles. Adapting/replacing it through the approved
runner-owned controller is unfinished. This Important finding prevents treating
the checkpoint as merge-ready; it is not merely an unavailable Docker run.

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
active desired-state projection yet. The existing PITR helper checks readiness
and reads a fence count; it does not provide the required same-connection
readiness/query/readiness contract for those real projections.

Later tasks should reuse the existing private canonical/envelope/evidence
helpers, while completing the public schema/role registry and exact-epoch
projection. The current ordinary checkpoint query still selects
`MAX(authority_epoch)`; that is not the planned activation/terminal-chain
projection. Finally, `-Suite batch01` requires the tracked
`testdata/c12/integration-contracts-schema-authority.v1.json`, which is absent.
The runner currently rejects that missing manifest before starting test groups;
the manifest must be derived from the completed required integration cases.
