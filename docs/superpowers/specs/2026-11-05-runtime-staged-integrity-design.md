# Staged writer integrity review design

Status: DESIGN_PENDING_APPROVAL. User selected staged verification; this document
requires review before implementation planning. Current43/32/6/217 unchanged.
This is a narrow scheduling/acceptance change, not proof of a timeout fix.

## Intent and approaches

Observation43 recorded task.returned then pins.begin before DrainTimeout. Its full
scan involves2703comparisons and227013378logical file bytes. A2second writer exit
window currently includes this historical audit as well as transport finalization.

Selected: preserve pre-run verification; make the parent responsible for a complete
post-exit audit, outside active child supervision. Final observation acceptance
still requires that audit. The parent already has a post-supervision full-map check;
reuse that single pass rather than add a duplicate audit.

Alternative: optimize scanning while retaining writer pre-exit verification. This
changes the timing semantics less but leaves a growing full-history scan inside
the2second window. Throughput sufficient for that limit is not demonstrated; simple
deduplication affects only about0.54percent of current logical bytes. Not selected.

## Exact scheduling change

|Phase|Required work and owner|
|---|---|
|Preparation|Pin full manifest, source, runtime and DLL identities; checkpoint before launch|
|Before processing|Keep existing parent full before-map, writer full before-map and original producer before checks|
|Business prefix|Keep original guards, two groups, cutoff,62events and producer after checks|
|Writer finalization|Keep actual stream EOF, bounded payloads, sink close/hash and authenticated COMPLETE; publish role result and exit|
|After confirmed terminal supervision|Parent performs full post-map verification of every original input/absence pin|
|Final verdict|Require transport, business, full parent audit and eligible output/evidence checks; otherwise reject|

Remove ONLY the writer's post-task full Map-IW call, its after-map publication and
associated timing markers in a new version. No reduction in the parent's checked
set, no cached hashes, no deduplication, no selective archive skipping. Do not remove
producer checks or change the original capture body/guards. Writer exceptions and
role-result failures still fail its exit; no failed operation becomes a success.

Writer role evidence explicitly reports After=null and
AfterVerificationMode="PARENT_POST_EXIT_REQUIRED". It must never contain an invented
writer after-map, a copied before-map labeled after, or a flag implying it performed
that audit. The new final verifier accepts this declared ownership change only for
the newly approved scenario; old scenarios retain their old contracts unchanged.

Reuse the existing compiled supervisor/adapter and their process associations.
The parent begins its audit only after supervision is terminal with no pending
resources or unconfirmed exits. On transport failure it may retain the existing
post-map diagnostic check, but cannot upgrade FAILED to success or inspect
unavailable output files. Unknown exits/pending resources stop that postprocessing.

## Separate statuses, timing and failure behavior

Preserve the raw supervisor outcome before audit. Add a separate exclusive audit
record with InputSHA256, IntentSHA256, parent-before map reference, parent-after map
reference (null on failure), state VERIFIED/FAILED, and measured duration when
available. Final status remains NOT_ACCEPTED if the audit is missing, fails, or
cannot be persisted; transport SUCCESS alone does not authorize use/publication.
Use a separate final observation verdict; do not edit the supervisor receipt.

Keep120000ms execution,5000ms readiness,2000ms shared normal drain/finalization and
2000ms shared failure exit-confirmation budgets. The parent audit is synchronous
filesystem work after child exit, as the current post-map already is. These wait
budgets do NOT bound audit time or total wall time. No process/async wrapper is
introduced to pretend a hard audit deadline exists. A hung filesystem can still
prevent final evidence publication; report pending/incomplete, never accepted.

Assistant choice: preserve small writer phase diagnostics through a NEW contract
with only task.returned, role.save.begin, role.save.end, exit.ready (plus at most
one terminal error), maximum5records/384bytes each/1920total. Use new identity role
"writer-staged-timing". Preserve the same no-retry/latching and raw-tail rules as
the prior timing design. Do not silently feed this four-phase trace into the old
nine-phase parser or change the old helper. No new compiled code. Timer records
still cannot prove actual process exit or exclude instrumentation overhead.

## Guarantees retained and intentionally not equivalent

Retained: all pinned files are checked before the run and by the parent after
confirmed exits; persisted drift, missing files and malformed/absent evidence fail
closed. Business results, real EOF, exact48/0byte output, hashes, valid COMPLETE,
both normal owned exits and completed supervision tasks remain required.

Changed: there is no independent writer-process full check immediately before its
exit. The parent audits later and in a different process. A transient modification
visible during the old writer audit but restored before the parent reads could be
missed; the later parent audit could instead observe a later modification. Neither
sequential-hash scheme is an atomic snapshot or proof that files never changed.
Do not claim equal temporal coverage, race freedom, loaded-module attestation,
protection against a malicious parent/local administrator, or complete equivalence.

This change trades that writer-local checkpoint for separation of process cleanup
from historical evidence auditing. Approval must explicitly accept this narrower
temporal guarantee; if that is unacceptable, retain the current arrangement and
design a genuinely different integrity mechanism separately. No locking, snapshot,
service, security-setting change or stronger trust guarantee is implied here.

## Validation and authority boundary

Future plan must test final-verdict refusal for transport-success/audit-failure,
missing/malformed audit, input drift, wrong identity, absent map, late role failure
and incomplete timing. Static allowed-delta checks must prove producer guards and
complete parent input sets remain unchanged; old evidence stays byte-identical.
Test version-specific contracts so old cases cannot bypass writer-after requirements.

Proposed later observation: one new computation44, at most writer+producer2starts,
zero compilation/snapshots, proposed ceiling44/34/6/217. Not authorized by saving
this document. Exact scripts, fixtures, finite tests and stop conditions belong in
the separately approved plan; no automatic retry after a failed gate.

All five capability flags remain false, no full-suite/main-merge/production acceptance
or original-timeout-root-cause claim. Preserve necessary source and portable evidence
on the existing development branch, excluding binaries/caches/unavailable data.
This design turn changes documentation only; no runtime is launched.
