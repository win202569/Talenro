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

## Next implementation work

1. Establish clean-checkout test reproducibility without the former local source
   overlay; use [repository recovery](../runbooks/repository-recovery.md).
2. Close the remaining [Batch 01](../superpowers/plans/2026-08-23-node-pop-control-plane-01-contracts-schema-authority.md)
   implementation and acceptance gaps. These include guarded serving, complete
   canonical protocol/verifier coverage, exact-epoch readiness, and the tracked
   integration manifests/final Batch 01 gate. Do not infer completion from table
   creation or isolated oracle results.
3. Proceed through the remaining [C1.2 batches](../superpowers/plans/2026-08-23-node-pop-control-plane.md),
   including inventory/operator state, node identity and mTLS, agent, supervisor,
   observations, core adapters, and final operational acceptance.

The [implementation sequence](implementation-sequence.md) preserves historical
acceptance records. Its C1.2 `current design` label predates the current partial
implementation; it is not a declaration that C1.2 has passed final acceptance.
