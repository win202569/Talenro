# Project status — 2026-09-08

This snapshot was checked against the tracked implementation and plans at
`2668b901`, followed by the repository-portability repair. It distinguishes
implemented components from the acceptance criteria of their containing stage.

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

Task 9 is the first confirmed semantic gap: `Coordinator` still accepts an
arbitrary resolver, calls provider Abort before the required durable claim
sequence, and does not connect activation to the existing sealed
`EffectDispatcher`, captured evidence, and persisted-outcome validator.
Integrate those existing boundaries and prove transaction/recovery behavior
before adding the Task 10 bound reader.

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
