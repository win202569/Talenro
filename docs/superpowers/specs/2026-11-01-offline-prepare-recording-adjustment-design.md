# Preparation recording adjustment — proposed written appendix

Status: DRAFT_FOR_APPROVAL; documentation only. No execution authority.
Baseline: b0d7dbb0480bd07d7cfc643d19951814514e873e.

## Intent and boundary

Preserve the [actual preparation failure](../../roadmap/2026-10-31-offline-integrated-review-repair.md)
and propose the smallest safe way to correct its recording order. This appendix
does not silently amend the [original integration plan](../plans/2026-10-31-offline-integrated-review-repair.md).
User consent in this turn covers writing this document, not implementation,
another Prepare, tests, imports, compilations, targets, review or budget release.

The known error was writing a receipt while preparation was still checking
packet inventory. Waiting first prevents that race, but alone cannot permit
another Prepare: the retained stop.json deliberately blocks every action and
the original seed inventory rejects three of the five retained files. These
are working safety gates, not false positives to bypass implicitly.

## Approaches and recommendation

1. Timing-only operator correction: always wait before persisting a receipt.
   This needs no code change, but correctly leaves the existing packet blocked.
   It cannot deliver a new seed freeze.
2. Recommended for a later, explicitly approved execution plan: one narrowly
   bounded preparation-only exception in the existing driver, binding all five
   actual retained records and the unchanged seed. Preserve the stop marker,
   all historical bytes and the existing namespace. Exactly one additional
   seed preparation may be requested; stop immediately afterward even on success.
3. Not recommended: delete/move records or stop.json, loosen inventory to ignore
   unknown files, copy to a fresh namespace, or refund/restart the runtime grant.
   These hide the actual stop or expand authority and are outside this proposal.

Option2 is an author recommendation, not a user-selected or approved exception.
This is a small design appendix, not an executable implementation plan.

## Proposed constraints for the subsequent plan

Only docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/driver.ps1 may
receive a preparation-control change. Its current seed hash is
4F75FBDD23CD6A21C3567460A1CFC62661B5B00D8D5A77F8FE4300CFE1500EE9.
The other27seed sources, both unresolved I1/I2 implementations, all test
vectors/oracles, publisher, recorder and process checks must remain identical.
Retain the original driver bytes from the committed checkpoint as provenance;
do not represent the revised driver as the original28source static result.

Before any proposed exception, exact inventory must be these five regular
files, no directories, extras or substituted paths. Hashes below bind the
stopped checkpoint; a mismatch is a stop, not grounds for refreshing values:

| File in retained packet | SHA256 |
| --- | --- |
| progress.md | 55B652081DA45F3F891D14C1DE95D634A3BA5872B53F9C6D60ECF59C84061E63 |
| seed.prepare.tool-result.json | C6EDB04CFD2C1E78BE02E497A48DBA4A8E4F74CB4E95EB83CCDAD700F8903719 |
| static.tool-result.json | 69620AFC7571CB0D83F10812A405D7044A5F0135FEF628B9F2D40512A68441EA |
| stop.json | 0D4FE34A03FAD3FD1D45B8C4B850C5B6D8C8BF9F3C9283304503B7119FD7EDE2 |
| author-integrity.tool-result.json | 8F091601FAD7F9603D603F786458B7EB4EFAFD0952A65FF38A82FE2237994E02 |

Portable-stopped.json remains unchanged, SHA256
1D10ABA536D414B06BE67422EA3662EF68AF4926858A70EBBBCFAB38E1EE6ADF.
The exception must be specific to Action=Prepare, Grant=seed and an explicit
approved adjustment identity. It must never authorize RunOne, writer-green or
repair-green while the stop remains. Keep the original stop.json unchanged.
Do not introduce arbitrary paths, callbacks, permissive file patterns or a
general override switch. Preserve engine/identity, AST/LF/JSON, inherited2394,
historical portable byte checks, exact2482pins and28source snapshot checks.
The original plan/spec snapshots remain those original documents; any later
adjustment approval/provenance is a separate attestation, not a relabelled
original approval or an unannounced new manifest schema.

The later implementation plan must specify a durable, create-once attempt
marker outside the inventory-sensitive packet, before invocation. Its exact
path, contents, approval binding and failure behavior must be fixed there.
This is needed because a failed attempt before creating seed.frozen.json must
not accidentally make another attempt eligible. No automatic retry on crash,
unknown result, partial snapshot or failed receipt persistence.

For the proposed single attempt, keep initial and every subsequent actual tool
object in memory; wait only its returned session ID until terminal completion.
Persist the complete transcript once under a new, predeclared non-overwriting
receipt name only after that command has terminated. Never write its initial
yield into the packet. A terminal nonzero result, missing chunks, lost process
association or unknown completion is a real stopped result. The original catch
does not replace an existing stop.json; therefore the new transcript must retain
the actual new exception without relying on that old marker as new-error proof.

No capture host, controlled target, fixture compilation, runtime import or
review is part of this proposed preparation-only scope. All NEW7/7/2/conditional1
runtime allowance stays FROZEN, even if a new manifest is produced. All older
frozen/exhausted grants remain unchanged. Preparation success would prove only
the explicitly checked preparation facts, not I1/I2 or runtime acceptance.

## Approval and handoff

Current turn: document and preserve this proposal only. No source was changed
and no preparation/test was invoked to develop it. Next ask the user to approve
option2 as the design for a separate, finite preparation-only implementation
plan. That plan must pin the adjustment identity, exact marker/receipt paths,
static-only preflight, one-attempt ceiling, retained-record verification and
handoff steps before any execution approval is requested. Do not combine design
approval with automatic runtime resumption. Any disagreement or inability to
retain the exact boundaries means stop and report, not a new generation.
