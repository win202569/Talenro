# Authorized evidence-only correction after19c567a1

The user explicitly approved lossless derived-summary correction and readonly
comparison only. Historical Important1 below is now resolved:8integer tokens
restored;all9embedded evidence subtrees match unchanged raw files,including
exact numeric lexemes. All26packet files/1326pins/product unchanged. Zero new
targets,compiles or reviewers. Minor2 remains accepted/deferred;all5scope
qualifications remain. This is parent verification,not a new reviewer verdict.
[Correction record](../../../roadmap/2026-10-18-p2b-inbox-evidence-correction.md).
Original parent dispositions and independent review below remain unaltered.

# Parent dispositions — 2026-10-18 (historical before authorized correction)

The text below the verbatim-review heading is the complete independent response,
not a parent rewrite. Parent reviewed every finding against unchanged sources
and saved evidence. All five runs remain exact; no product defect was identified.
Status: UNFINISHED_REVIEW_BLOCKED_PORTABLE_UINT64. No automatic fix/retest/review.

## Findings

1. Important1 ACCEPTED/BLOCKING. Readonly integer-lexeme comparison confirms the
   original raw full/final traces contain exact UInt64 boundary values, while
   both OriginalTrace.Synthetic and FinalOriginalTrace.Synthetic in the derived
   portable summary are rounded. Parent construction crossed JavaScript
   JSON.parse/Number/JSON.stringify, which cannot preserve these integers.
   Summary fields are known inaccurate and must not be used as boundary evidence.
   Raw streams/traces/hashes remain unchanged and correct. No promotion or
   automatic correction: a separately approved lossless evidence-only correction
   and direct text comparison are recommended, with zero targets/compiles/reviewers.
2. Minor2 ACCEPTED/DEFERRED. New Count.Probe additions do not directly assert
   retained or absent scalar ETS metadata, and IB-Eq ignores scalar decorations.
   Fresh scalar guards/reference assertions and original passivity coverage are
   supporting evidence, not a substitute for those missing assertions.
   No production defect demonstrated. No frozen test edit or extra host authorized.
   Cost of deferral: scalar-metadata isolation has narrower direct evidence.

## Declined-to-judge dispositions / rulings

1. Historical vanished-object identity: scope limit accepted. New Representation
   proves only its controlled object; no historical-object identity claim.
   Cost: original object's exact runtime representation remains unavailable.
2. Intermediate restoration timing: scope limit accepted. Parent observed the two
   exact-good hash checks in tool output before proceeding; five packet rows do
   not independently timestamp those two restoration events. Current/final-good
   identities are verified. Do not claim independent proof of exact timing.
   Cost: portable five-row evidence alone cannot establish both intermediate times.
3. Arbitrary/cyclic/private-tampered trees: scope limit accepted. Classifier is
   private for fixed owned trees, not an untrusted deserializer/security boundary.
   Cost: callers cannot rely on behavior under same-process reflective tampering
   or arbitrary private-helper misuse. No public input expansion is authorized.
4. Other runtimes/concurrency/live/broader branches/main: scope limit accepted.
   No such acceptance follows from these five ordinary PS7.6.5 saved runs.
   Cost: portability/integration/live properties remain unverified.
5. Independent runtime reproduction: scope limit accepted. Reviewer was expressly
   readonly, static sources/hashes/saved evidence only; no rerun or new allowance.
   Cost: review is not an independent reproduction and does not itself add tests.

## Process ruling carried forward

Use the approved finite workflow in the existing PowerShell worktree; preserve
all raw/worktree data and do not add generic suites, fix passes, replacement
reviewers, merge menus or cleanup. Standing authority permits only necessary
ordinary development-branch checkpoint commit/push. Cost: no broader verification
or acceptance is claimed. The Important finding stops completion as required.

## Verbatim independent review

## Strengths

- The production repair stays within the authorized boundary: only `Copy-IBTree` and the new private `Get-IBTreeNode` differ. The loader/registry prefix and everything from `New-IBResult` onward are unchanged.
- The classifier preserves the original property owner separately from the native value, bounds wrapper traversal to eight levels, allocates fresh `object[]` instances, and delegates scalar copying and unsupported-type rejection to the unchanged codec. No public entry point, input-validation path, token exception, or receive semantics were broadened.
- The original 32-group suite, helpers, fixtures, and descriptor fixture match their baseline hashes. The original full-output suffix and both mutant oracles are unchanged.
- Expectations are independently constructed. Object comparisons check field order and underlying types; empty-array tests establish distinct references and retained ETS mutations; byte mutations, subsequent snapshots, stopped snapshots, unsupported leaves, five exports, and original input-passivity cases are covered.
- All **1,326 frozen pins match**, with all **1,289 inherited mappings unchanged**. Frozen manifest and seal hashes match the review package. The current product and saved good product both hash to `47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC`. Both mutant snapshots contain exactly their specified single substitution.
- All five receipts’ stream and trace hashes match their saved files. Outputs match the frozen oracle:

| Slot | Exit | Output lines | PASS lines | Elapsed |
| --- | ---: | ---: | ---: | ---: |
| RED | 1 | 3 | 0 | 1,081 ms |
| Full | 0 | 49 | 44 | 44,734 ms |
| Capacity | 1 | 4 | 1 | 1,888 ms |
| Sequence | 1 | 4 | 1 | 1,165 ms |
| Final | 0 | 49 | 44 | 47,987 ms |

All stderr files are empty. Receipts record completed owned exit, stream drains, post-pin checks, and no timeout. Full/final stdout and both corresponding traces are hash-identical. Original traces record compilation count 1, observable counters, counters 0, four synthetic cases, and eight retained decorations.

The mutant assertion locations establish incorrect acceptance, not an unrelated internal error: capacity fails at `inbox.tests.ps1:139`; sequence fails at `inbox.tests.ps1:107`.

## Issues

### Critical

None found.

### Important

**1. Portable evidence corrupts the exact UInt64 boundary values.**

File: [portable-summary.json:309](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-inbox-tree-copy-repair/portable-summary.json:309), also lines 314, 319, and 324.

The committed `OriginalTrace.Synthetic` values differ from the saved raw trace:

| Raw trace | Portable summary |
| --- | --- |
| `9223372036854775808` | `9223372036854776000` |
| `18446744073709551614` | `18446744073709552000` |

These differences are present in the file text, not merely display formatting. The latter portable value exceeds `UInt64.MaxValue`.

**Impact:** A fresh-computer reader does not receive an accurate record of the high-sequence and exhaustion-boundary cases. This undermines the handoff’s claim to preserve complete traces and can mislead subsequent analysis or reconstruction. The raw traces and their matching hashes remain intact; this is a derived-evidence defect, not evidence of a product failure.

**Fix direction:** Under explicit disposition of this finding, preserve integer lexemes exactly when producing portable evidence, or use clearly documented decimal strings for UInt64 fields. Compare the resulting portable boundary values directly with the unchanged raw traces. Do not rewrite raw evidence or rerun targets.

This is separate from the acknowledged interim two-host status, which I have not treated as an issue.

### Minor

**2. Scalar-metadata isolation is not directly observed by the new mutation assertions.**

Files: [tree.tests.ps1:86](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-inbox-tree-copy-repair/tree.tests.ps1:86), line 93; [test-helpers.ps1:14](D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/docs/diagnostics/isolated-fuzz/p2b-inbox-tree-copy-repair/test-helpers.ps1:14).

`ReturnedMutation` and `TwoSnapshots` attach `Probe` to `Count`, but do not subsequently establish that the decoration remains on the modified scalar or is absent from the other snapshot. `IB-Eq` compares scalar values without examining their ETS properties, so the later complete-literal comparisons do not fill this gap.

**Impact:** These groups demonstrate byte/value isolation more strongly than scalar-metadata isolation. The unchanged codec’s fresh-scalar guard and the focused reference checks provide supporting evidence, so this is not a demonstrated production defect.

**Fix direction:** In a separately authorized test change, retain the decorated scalar, assert that its `Probe` is observable, and explicitly check that counterpart scalars lack the decoration. Do not alter the frozen suite or run extra hosts within this review.

## Recommendations

- Preserve this review and disposition the Important finding before claiming a completed scoped handoff. No automatic repair, additional test, or replacement reviewer is implied.
- Keep the existing preflight-stop account, original failure records, and all raw receipts unchanged.
- When authorized to finish the continuation handoff, distinguish the five completed saved runs from the earlier two-host checkpoint and retain the limits on runtime authority.
- Describe scalar-metadata coverage with the qualification above; only `FirstSnapshot` had a RED run.

## Declined to judge

- **Direct runtime identity of the vanished historical failure object:** unavailable. `Representation` observes only the newly constructed test object, as correctly documented.
- **Exact timing of each intermediate restoration:** the supplied review package reports immediate exact-good checks, but the saved five-run receipts do not independently record both intermediate restore events. Current exact-good identity and its use by the final run are verified.
- **Arbitrary caller-created trees, cyclic graphs, malicious private-scope invocation, and same-process reflective tampering:** the helper is restricted to internally owned fixed trees, not a public deserialization or security boundary.
- **Other PowerShell/SDK versions, concurrent operation, live I/O, real processes/capabilities, broader historical branches, full-repository acceptance, and main integration:** outside this authorized review and unsupported by these five runs.
- **Independent runtime reproduction:** expressly prohibited. Conclusions use static inspection, Git/file comparisons, hash checks, and saved evidence only.

## Bounded verdict

**Not ready to label the scoped checkpoint complete:** the product repair and five saved runs support the bounded technical result, but portable evidence currently misstates the exact UInt64 cases. An explicitly unfinished development checkpoint may preserve this finding.

No production defect was identified in the allowed repair. This review does **not** authorize main merge, formal acceptance, live use, automatic fixes, additional hosts, or re-review.
