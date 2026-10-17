# Remaining offline acceptance scope — proposal, not clearance

## Finding from readonly scope reconciliation

After5d7f7261, the user requested clarification of the remaining acceptance
scope. This document classifies the existing gates; it does not change the
approved design/plan, execute tests, dispatch reviews or grant acceptance.
The [closing review](2026-10-17-synthetic-capture-closing-review.md) remains the
latest independent assessment. No confirmed new code defect is inferred here.

|Item|Current classification|Proposed next treatment, not yet approved|
|---|---|---|
|Original whole-boundary/all10-success review prerequisites|Active procedural gates; later separate packets did not satisfy them literally|Draft an explicit offline-only acceptance addendum mapping immutable later evidence to each requirement; never silently mark old runs/reviews successful|
|PS5 historical timeout, normal-context link-fixture failure, strict-stderr reader failure|Failed historical records with unresolved causes|Retain results and raw records; distinguish acceptance of alternative evidence from proof that the historical failure is fixed|
|File-level reparse refusal|Explicit design section9 requirement without direct fixture evidence; directory junction is not the same case|Retain as an open requirement; design a bounded real-file test before any new execution, without changing ACL/security settings or skipping refusal|
|PS5 missing-root dynamic branch|Disclosed coverage limit, not an automatic unconditional retest requirement|Original plan step2 permits reporting an existing-root precondition to review; an acceptance addendum must explicitly decide whether that disclosed limit is acceptable or needs a bounded fresh-root test|
|Full independent populated-package oracle/max duplicated-package fixture|Additional limitations from review, not confirmed defects|Map existing required field/boundary evidence first; explicitly decide any remaining obligation instead of treating all conceivable coverage as mandatory forever|
|Current ordinary-user complete reader evidence|Context limit: later complete reader assertions used admin context|Keep the limit visible; acceptance scope must specify supported/verified contexts, not claim admin universally required or ordinary-user success|
|Specified malformed-event/writer-collision cases and named dependency regressions|Bounded evidence supplied and independently assessed|Reuse hash-matched reports; do not rerun already passing packets merely to obtain a new green record|
|Live capture, signing/provenance, power-loss durability, Stage B/I2/I3/full C12/security clearance|Outside offline-store acceptance; separate project gates remain|Do not broaden this module to meet these non-goals or claim them cleared|
|Main merge or release|Separate authority, not granted by scope clarification|No merge/release from this proposal or a future offline-only result|

## Recommended next deliverable

Prepare a proposed written **offline-only acceptance addendum**, not another
test/review run. It should specify:

1. Immutable source/test snapshots and the precise evidence-to-requirement map.
2. Which new acceptance route is proposed in place of the old stopped route,
   while old failures, frozen slots and spent budgets remain unchanged.
3. Required unresolved cases versus explicitly accepted coverage limits,
   including file-level reparse and ordinary/admin context distinctions.
4. How a final offline-only verdict could be reached, and what it cannot clear.
5. Any future test/review packet must have its own scope, context, fixed expected
   results,120s cap, fail-stop rules and fresh finite budget before execution.

The addendum would need user review before taking effect. This is a bounded
change to acceptance bookkeeping, not a redesign of the product or permission
to weaken validation. No test count, new host, reviewer, fix reserve, root
relocation/deletion, permission/security change or old-slot reuse is granted.

## Sources and recovery

- [Original plan](../superpowers/plans/2026-10-17-synthetic-capture-store.md),
  step2 explicitly permits truthful missing-root coverage disclosure when the
  root already exists; steps7/9 retain the original success/review conditions.
- [Approved design](../superpowers/specs/2026-10-17-synthetic-capture-store-design.md),
  section9 includes directory/file reparse refusal; non-goals stay excluded.
- [Coverage map](2026-10-17-synthetic-capture-coverage-summary.md),
  [supplemental results](2026-10-17-synthetic-capture-supplemental.md),
  [dependency results](2026-10-17-synthetic-capture-dependency-regression.md)
  and closing review preserve the separate contexts and original dispositions.

Only this proposal and status/recovery notes are new. No source/test or raw
receipt changed. Save on the existing development branch; a new computer can
recover the classification without conversation history. All acceptance,
security/main and spent/frozen execution boundaries remain in force.
