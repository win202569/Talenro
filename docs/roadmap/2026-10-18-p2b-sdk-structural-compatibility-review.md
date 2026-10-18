# Independent readonly review — P2-B SDK structural compatibility

2026-10-18. Reviewer: /root/p2b_sdk_structural_readonly_review.
One fresh seat, dispatched after all three exact hosts; read-only, no nested
reviewer or target execution. Base98f21970a41288ca2acbd008f46cde8ae2657135.
ModuleB468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039.
The reviewer response below is preserved verbatim; executor dispositions are
in 2026-10-18-p2b-sdk-structural-compatibility.md,not substituted for this text.

## Strengths

Read-only review completed for all ten current candidate files, the controlling text spec/plan, SDK metadata evidence, and saved receipts. No product/helper/test/controller execution, compilation, network access, edits, or nested reviewer occurred.

- Product SHA256 matches `B468B564…D039`; spec, frozen manifest, product seal, pinned engine and SDK hashes match. All 998 mapped pins rehashed with zero drift.
- `control-frame.ps1:76–142` preserves Message12/Expected13 business shapes, uses raw member enumeration, checks exact descriptor classes/name/flags before business values, and rejects hidden/extra/duplicate/subclass/masquerading root entries.
- Owner checks preserve actual PSObject references through reflection argument arrays (`:27–31`, `:91–94`). Optional metadata requires the declared owner references; `pstypenames` additionally requires the fixed SDK getter reference and null setter (`:109–119`). No metadata getter/factory invocation appears in this validation path.
- Product contains function definitions only. Wire layout, declared integer validation, reason ordering, independent result construction, and three false authority flags remain intact.
- Tests use independent literal frames and real observable descriptor mutations, not product-generated expected results. Saved full output includes all 33 PASS groups, including `SequenceExact`.
- Recomputed raw-output comparisons and stream/trace hashes match all three receipts: RED 6 lines/exit1, focused 7/exit0, full 36/exit0. Recorded ordinary context, owned exit, complete streams, no timeout and empty stderr are consistent with their saved evidence.
- Controller source enforces preconsumption/order/no retries, pinned inputs, separate bounded stream capture, and removal of at most one final newline.

## Findings

Critical: none verified.  
Important: none verified.

Minor — scalar-decoration coverage is narrower than root-descriptor coverage.

Current `control-frame.ps1:33–36` checks scalar/byte-array decorations through the public `PSMemberSet.Members` view, whereas root descriptors use the raw collection. `control-frame.tests.ps1:171–188` exercises retained visible scalar decorations, but does not cover hidden or subclass descriptors attached to scalar/byte-array inputs.

Impact: the saved counter results do not establish exhaustive scalar-decoration passivity/refusal. This is a verified coverage gap, **not a verified accepted-input bypass or callback activation**.

Suggested disposition: preserve this limitation explicitly. Any targeted investigation or additional test requires fresh authorization; do not reopen these consumed hosts or modify frozen tests.

## Recommendations

Preserve the candidate and receipts unchanged. Describe the result as narrowly demonstrated SDK structural compatibility—not factory provenance, exhaustive passivity, or security isolation. Record the Minor coverage limitation alongside the next decision.

## Declined to judge

Every considered boundary not established here:

- Hidden/subclass scalar-decoration behavior beyond existing vectors: no matching saved vector or sufficient permitted SDK evidence; no new probe authorized.
- Universal metadata-getter nonexecution: counters measure CounterNote only. The reviewed product path avoids metadata getter calls, but this is not universal dynamic instrumentation.
- Arbitrary same-process private reflection, unsafe mutation, concurrent mutation, or hostile type-data: outside the approved structural contract.
- Other PowerShell/SDK versions, PS5, architectures or platforms: only pinned PS7.6.5 Core x64 is supported by this packet.
- Catastrophic recovery, unsupported-SDK setup failures and exhaustive internal-error fault injection: not demonstrated by saved runs.
- Actual factory provenance, endpoint identity, authentication, capabilities, clocks, processes/Jobs, release, queues, capture completeness, cleanup, P2-C/P3, security or live operation: outside this declared-only codec review.
- Historical executable correctness: inherited sources and baseline snapshot were hash-only.
- Independent recreation of runtime outcomes: prohibited; runtime conclusions remain limited to saved receipts.

## Assessment

Ready for **LIMITED OFFLINE NEXT DECISION**, with the Minor coverage limitation retained. No verified Critical/Important finding blocks that narrow handoff. This is not main/production acceptance, formal P2-B acceptance, blanket passivity assurance, or authorization for further execution.
