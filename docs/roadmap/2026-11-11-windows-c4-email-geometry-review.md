# Email geometry independent review and disposition

Reviewed source range:e4c42a33..a28857c8. Fresh read-only reviewer, no subagents or
native build/run. Finding count:0 Critical,2 Important,0 Minor. Verdict: With fixes;
source-stage only, not permission to merge main or claim C4 acceptance.

## Important findings and single fix pass

1. ChangePage cleared the logical pending slot while physical dispatcher work still
   retained page closures. Added repeated-navigation dispatcher-double regression;
   it failed when queue length reached2. Coordinator now invalidates eligibility
   but holds the slot until callback drain or enqueue cancellation. Existing navigation
   test revised to require drain before new enqueue; both passed in the fix pass.
2. Explicit root/presenter clips were skipped after recording their larger element
   boxes. Added containing-box/overflowing-clip fixtures for both roles and bounded
   error/unsupported acquisition tests. Shared collector records separate `<role>:clip`
   boundaries for every measured node; root/presenter no longer skipped. Receipt
   validator regression failed before accepting the new allowlisted boundary identities.
   Fix-pass targeted suite49/49 passed. Overall clip ancestry remains incomplete.

The reviewer found no concrete API mismatch against installed WinUI2.6.5 references.
That is not native compilation. No second review was dispatched; regression tests
and the final full suite provide fix evidence, not a fresh reviewer verdict.

## Declined-to-judge items and executor rulings

- Native compilation/generated bindings: deferred by user scope; must be verified
  before use. Risk: source-compatible-looking adapter may still fail compilation.
- Actual output/intermediate metadata isolation and new hashes: deferred to approved
  native build; evaluated properties are not artifact evidence. Risk: build routing
  defect remains possible; old hashes must be checked again before/after that build.
- Real WinUI event order, template ancestry and event removal: deferred to later live
  observation. Portable doubles do not establish runtime behavior; risk: missing samples
  or retained handlers despite source checks.
- Clipping reproduction, glyph readability, compositor correctness/WinUI defect and
  repair: not established. Box evidence cannot prove these; risk: incorrect diagnosis
  if treated as acceptance. C4 stays NOT_ACCEPTED.
- Hostile replacement of trusted filesystem ancestors: excluded from current factory
  guarantee. Trusted parent precondition retained; metadata checks have TOCTOU limits.
  Risk: use in an adversarially mutable parent is unsafe and unsupported.
- Hard close IO deadline/crash recovery: not promised. Keep synchronous bounded-buffer
  design, external overrun failure and incomplete crash evidence; risk: close can hang
  and in-memory data can be lost. Never accept footer without original-process exit0.
- New fixed-hash entry/live readiness: deferred until approved build+verified hashes;
  risk: ad-hoc invocation could evade gates, so no new runnable entry is supplied.
- Independent rerun of tests/hashes: reviewer did not rerun. Coordinator supplied
  observed evidence; risk: review is source-only independent assurance, not independent
  execution. Preserve this distinction in handoff.
- Unchanged prior branch outside feature range: no blanket reacceptance. Existing
  six synthetic suites and portable tests are regression checks only; historical
  native acceptance blockers remain unchanged.

Deferred minors: none.
