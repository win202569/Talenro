# Synthetic capture store — cross-packet evidence and remaining gates

Later dependency regressions: [six-run results](2026-10-17-synthetic-capture-dependency-regression.md)
now provide81/106/110 PASS per engine in the separately approved contexts.
The historical unrun-dependency statement below describes the earlier snapshot;
remaining acceptance/context/coverage limits and old failed receipts persist.
No source/test change or automatic old-review promotion follows.

Later coverage extension: [supplemental tests](2026-10-17-synthetic-capture-supplemental.md)
add direct malformed-event and writer-failure evidence (new9/6 PASS groups on
each engine, four normal-user children). Original table below remains the
five-group historical map; new tests do not rewrite failed receipts or review
gates. Dependency regressions and remaining limitations stay open.

Subsequent disposition: user approved the NEW one-review scope below after
562ec07e; that review is now complete. See [review findings](2026-10-17-synthetic-capture-readonly-review.md).
Three Important validation gaps remain, no confirmed product defect; no tests,
repairs or acceptance promotion. Proposal text below records the historical
pre-dispatch scope, not another unused review allowance. Old reviews stay gated.

## Scope of this consolidation

After a841f997, the user requested the recommended coverage/gap consolidation
and proposed independent-review scope. This was readonly inspection of saved
receipts, current source and existing design/plan. No test/diagnostic child,
independent reviewer, code change or acceptance promotion occurred.

Source snapshot: development branch at
`a841f997654c5426b3e145f8471a796660b699a1`, before this documentation checkpoint.
Reports below are distinct execution packets, not one resumed successful suite.

## Completed assertion evidence at the current test hash

|Group|Content assertion occurrences per engine|PASS including sentinels|PS7 evidence|PS5.1 evidence|Context / disposition|
|---|---:|---:|---|---|---|
|admission|26|28|10843ms|2517ms|Original partition slots1/2 complete; ordinary context|
|byte-bound|1|3|18480ms|4367ms|Partition slots3/4 complete; ordinary context|
|event-bound|1|3|47442ms|80106ms|Partition slots5/6 complete; ordinary context|
|reader-paths|43|45|2372ms|2083ms|Separate administrator packet; PS7 receipt passes, PS5 original strict-stderr receipt fails|
|format|16|18|8323ms|2939ms|Separate normal-user packet; both pass prior-approved structured-progress policy|

Sources: [partitions](2026-10-17-synthetic-capture-partitions.md),
[administrator readers](2026-10-17-synthetic-capture-reader-admin.md),
[readonly reader disposition](2026-10-17-synthetic-capture-reader-receipt-assessment.md),
[normal-user format](2026-10-17-synthetic-capture-format.md).

The ten selected children provide 97 PASS occurrences per engine, comprising
87 content assertions and ten repeated API/completion sentinels. These are
occurrences in the selected tests, not 87 independent requirements or a measured
percentage of all possible inputs. Repeated unsafe-path labels retain their
order/multiplicity. Do not add the old partial reader's 38 PASS as new coverage.
The four boundary partitions contain 71 content assertions in total.

Readonly checks matched all ten raw stdout sequences and stderr lengths to
their saved indexes, then matched exact labels to the existing group inventories.
All ten index entries pin test hash406C7C66. No product/test drift was observed.
PS5 reader's 584-character stderr contains two Completed progress objects, as
separately assessed; its original ExpectedResult=false remains untouched.

## Remaining gaps and historical evidence

1. **Dependent regressions not run in this implementation packet.** Original
   plan calls for output-evidence-store (81 PASS), output-capture-simulation
   (106 PASS), capture-adapter-contract (110 PASS), each on PS7/PS5.1. Counts
   are plan targets, not new observed results. No spare slots are authorized.
2. **Independent review not started.** Neither original implementation review
   nor all10-success partition review prerequisite is satisfied. A new readonly
   review can assess source and evidence without pretending those gates passed.
3. **No unified full-suite pass.** Original PS5 boundary timeout remains a
   failure; old normal-context partition reader slot7 remains failed at fixture
   creation. Separate later runs do not rewrite either. Exact timeout/denial
   causes remain unresolved; context dependence is observed, not attributed.
4. **Context coverage is mixed.** Latest full reader assertions were observed
   under administrator context, not a completed current normal-user reader
   run. Do not infer normal-user success or a universal admin requirement.
5. **Missing-root evidence is engine-specific.** Original PS7 format run3 has
   PreflightPreservesRootState PASS and ROOT_PRECONDITION existed=False (old
   pre-partition test hash185FA256). Original PS5 run4 and both latest format
   runs have existed=True. Thus PS7 has historical dynamic evidence, while
   PS5.1 missing-root behavior lacks a corresponding dynamic result. Latest
   tests retained the original body; no root deletion/move or extra run occurred.
6. **Isolation and durability limits stay explicit.** No hostile directory-race,
   power-loss, live-process, source-authentication or full-project acceptance
   claim. Old-reader cross-format check is path rejection, not comprehensive
   new-payload parsing coverage. Hash consistency is not provenance or signing.

## Proposed NEW independent readonly review — not dispatched

Recommend one fresh reviewer for the offline synthetic-store implementation
and evidence boundary. This would be a NEW diagnostic/readiness review, not
automatic resumption or completion of either old conditional review.

- Primary files: `docs/diagnostics/isolated-fuzz/synthetic-capture-store.ps1`
  and `synthetic-capture-store.tests.ps1` in the same directory.
- Binding requirements: approved store design and implementation plan under
  `docs/superpowers/specs` and `docs/superpowers/plans`, dated2026-10-17.
- Readonly dependencies: capture-adapter-contract, isolation, output-capture-
  simulation, output-evidence-package, output-channel-classifier and old store,
  only as needed to trace calls and compatibility; no dependency edits.
- Evidence: this map and linked reports, exact source hashes and original
  receipt dispositions; raw receipts may be read locally, not required to clone.
- Focus: passive getter/method avoidance and bounds before allocation; exact
  fixed wire layout and independent oracle; preservation/revalidation of stored
  derived facts; safe handle/path/link/stopped/error behavior; honest treatment
  of mixed contexts, repeated sentinels and unrun regressions.
- Output: findings with severity, precise file/line evidence and rationale;
  uncovered requirements and explicit limitations even if no defect is found.
- Prohibitions: no dot-sourcing, tests, child processes, network, Git writes,
  file edits, fixture cleanup, safety changes, fixes, retests or further
  delegation. Parent may preserve the final report under standing Git authority.
- Stop after one review. Critical/Important findings remain unresolved, with no
  automatic repair or second review. A clean report does not clear regressions,
  original gate failures, security restrictions, Stage B/I2/I3/full C12 or main.

User approval of this proposed reviewer scope is still required. This document
does not launch a reviewer, grant more test slots or broaden current authority.

## Integrity and recovery

Current product SHA256:
`508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C`.
Current test SHA256:
`406C7C661A181988C7F0AE90E577D1F948148C2B19BE5CD411F92D23E7FECBFF`.
Rechecked index SHA256 values:

|Packet|SHA256|
|---|---|
|Partitions|B61DD587834978007736566C0951A8C91FAA9B5FADE1760B903455F932A19AB1|
|Admin readers|080455EE56A74498B09099061C53FC0CF46801C4ED614BEA37774D72629D8211|
|Format|5FA944DF33467C084ED50260CAA2201C77EF4341F482CFBD3E12163D71D52600|

Original6/12 spent,7..12 frozen; partitions7/10 spent,8..10 frozen; later reader
and format allowances each2/2 spent. All earlier diagnostic allowances stay
spent/closed. Raw receipts, runners and fixtures remain ignored/local. Git
preserves this portable evidence map and proposal on the existing development
branch, not main. Cloning restores neither ignored state nor execution authority.
Overall implementation remains an unfinished, unaccepted checkpoint.
