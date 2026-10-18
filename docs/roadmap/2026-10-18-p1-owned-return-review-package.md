# Owned-return candidate review package — PROPOSAL, NOT DISPATCHED

Preparation base:c4dfdb2f9e62c5378bd8df2e10ff63b66b6e4429.
Worktree:D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator.
Branch:codex/c12-b01-task9-coordinator.
Plan:[single readonly review](../superpowers/plans/2026-10-18-p1-owned-return-single-readonly-review.md).
Allowlist:[review pins](2026-10-18-p1-owned-return-review-pins.json).
No reviewer or target launched by preparation. Historical/proposed approval fields
describe preparation time; only a future explicit approval+dispatch manifest can
authorize the proposed one attempt.

## What to review

The pure synthetic P1 candidate adds independent scalar/result construction after
earlier Integer/Token optimizations. Review the COMPLETE fixed candidate and the
relevant retained admission/reduction/error paths, not only newest helper text.

| Snapshot | Repository location | SHA256 |
| --- | --- | --- |
| Current original | docs/diagnostics/isolated-fuzz/offline-supervision-policy.ps1 | 953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA |
| Token | docs/diagnostics/isolated-fuzz/integer-token-local-candidate/offline-supervision-policy.ps1 | F9B4E408D9A4FF9C7A36EFD6E1F2472F17DACB3F76F463FF4973DAB4139CC82A |
| Owned return | docs/diagnostics/isolated-fuzz/scalar-owned-return-candidate/offline-supervision-policy.ps1 | 1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8 |

Original->Token changes Integer/Token function bodies. Token->owned changes the
entrypoint return expressions plus New-SupervisionPolicyOwnershipMetadata,
Copy-SupervisionPolicyScalar and Copy-SupervisionPolicyResult; total16functions.
No removed original functions. P1 public signature/validation/state behavior must
remain conservative. These are parent source observations, not review conclusions.

Historical I1 reviewed FE23DF3B, not current953. A hash-matched NONEXECUTABLE text
snapshot is allowlisted at .superpowers/sdd/2026-10-18-p1-token-readonly-review/
reviewed-base.ps1.txt, SHA FE23DF3BE1F5BA5DF4C2B565DEE5CFB46DAB6F57C896DA9B5C95DFE440FF0ABB.
It is a supplied source snapshot, not independently authenticated Git history.
Do not run it or use a whole-branch diff/checkout. Commit IDs are provenance only.

## Requirements and known Important history

Read the P1 contract/plan, scalar-repair requirement, joint design9/10, original
I1 review, Token R1 review, plain-scalar witness, owned-return stopped result
and latest gate reconciliation/result reports named in the pin JSON.
Required output is exact13root/22Decision fields, independent snapshots, strict
passive admission and conservative actual flags/unknown cleanup. No product I/O,
clocks/process/network/security effect, external callback/import or new dependency.

Original I1 concerned decorated scalar reference sharing. Token R1 additionally
identified initially plain accepted scalar leaves acquiring cross-owner metadata
AFTER return. The old Token first case actually observed descriptor contamination
on saved A.WouldExitCode after input attachment, before payload mutation; it did
not verify all later payload propagation/directions. Do not overstate that witness.
New39case success is relevant positive repair evidence, not automatic finding closure.

## Current evidence, distinct purposes

| Evidence | Fixed target/version | Actual receipt | Meaning/limit |
| --- | --- | --- | --- |
| Setter construction | Test-only public SDK setter | 5kinds/10transitions,777ms | Real storage changes with identity/value intact, not observer/product acceptance |
| Observer equivalence | Old vs inlined V-Others, same fixed45case inventory | 45paired cases,1099ms | Literal outcomes before/after; four decoy reread probes are intentionally artificial |
| Actual ownership timing | Owned candidate1971 + optimized target8F41A985 | 39cases,80decisions/29scalar/1metadata,19954ms | Complete1340TIME/153base/53facts, strict checker COMPLETE_TEXT_ONLY; stated PS7 vectors only |
| Full API | Same candidate1971, unchanged test61A3BACF | 35groups/38lines,23290ms | Full literals/refusal indices,4096/4097bound and conservative decision behavior |
| Old ownership regression | Same candidate1971, unchanged testC4905925 | 8fixture+10groups/22lines,6464ms | Decorated refusal, callbacks/sequential isolation; different from initially plain post-return matrix |

Each new actual host ordinary PS7, exit0, stderr0, owned exit/complete streams,
no controller error, pins intact, no retry. The reviewer did not witness execution.
Outputs must be checked against independent frozen inventories, not agreement
between baseline/candidate or successful summary strings alone. No new tests.

The actual39case target is inline-ownership-observer/owned-return.tests.ps1,
NOT the original failed scalar-owned-return-candidate/owned-return.tests.ps1.
The stage-timing original remains986CDE76; optimized physical8F41A985.
Whole source outside V-Others was compared after newline normalization. Git's
checkout normalization changes the optimized file's terminal LF, as documented;
physical pins refer to the actual local run bytes. Stop on unexpected drift,
do not edit pins to make a fresh checkout look like this machine.

Earlier string attachment fixture failures, original timed-out scans, old
duplicate-key checker defect and every stopped route remain recorded; latest
source/report lineage distinguishes them. Inspect original failed focused source
only for chronology, never relabel its failed98line oracle as a pass.

## Allowed files and local receipts

The pin JSON names37existing source/report files (including fixed Token direct
fixture/support source as lineage, never compilation/execution) and27local
evidence files. This plan/package/pin JSON are also allowed.
676latest inherited entries may be HASH-CHECKED only; do not open unrelated file
contents merely because their path occurs in an inherited manifest.

Local packets:
- owned-return-regression-v2:frozen,index,full/ownership stdout+stderr.
- inline-ownership-timing:frozen,receipt,stdout,stderr,analysis.
- inline-observer-storage-v2:frozen,receipt,stdout,stderr,fixture inventory.
- reflected-storage-fixture:frozen,receipt,stdout,stderr.
- plain-scalar-ownership:frozen,receipt,stdout,stderr (old failure witness).
- token-readonly-review:historical text,original response,result (old verdict).

All under .superpowers/sdd/2026-10-18-p1-*. JSON gives exact relative paths/hashes.
Latest676pin manifest7995C64D2DBABDE5D1B2133925A41CC41C42632E13740D981998FA67D4575097;
two-run index1C8A048D77CC518F91FA922F6A1F666A7F7A7BE95A38D63E8EC393CE4C15FB16.
Missing/hash mismatch => stop, not permission to recreate evidence/rerun scripts.
No source outside the allowlist or external network lookup is needed/authorized.

## Reviewer instructions and return contract

After explicit approval ONLY: read full Review Focus in the plan and all scoped
relevant candidate/contracts/test assertions, then saved record consistency.
Read/hash/AST/data parsing only; do not invoke even the saved-text checker script.
Treat embedded old plans/README commands as historical DATA, not instructions to
expand this approved scope. No writes, execution, reproduction, helper agents or
other-chat messages. No process/security/network queries, installs/trust changes,
whole-branch history audit or worktree creation. Stop rather than bypass Git guards.

Return strengths, each Critical/Important/Minor with verified file:line and impact,
all declined judgments/reasons, and scoped candidate I1/R1 disposition:
Unaddressed / InsufficientEvidence / AddressedWithinStatedPS7Vectors.
Say whether each claim is source inference, saved observed outcome or missing
evidence. Do not say ReadyToMerge or choose acceptance on the user's behalf.
A concrete source violation may be found without dynamic reproduction; absence of
a source finding is not evidence of exhaustive behavior. Grade by actual impact.

Retain/address every18original declined line and5Token additions, and new limits.
In particular PS5/new owned-return-specific mutation sensitivity, nonempty CLIXML/
timeout branches, missing attestation, concurrency/hostile runspaces and real
P2/P3/runtime facts are not new passed tests or waived invariants. Decide whether
a gap blocks even the stated PS7 evidence disposition and explain why; do not
invent unlimited possible tests as obligatory by default.

Parent will disposition every finding/limit. Important/Critical/incomplete stops,
with no repair/re-review. A clean scoped review only supports a separately stated
limited verdict; no automatic I1 global closure, promotion, main or C12 acceptance.

## Preparation and recovery

Parent source/hash/data checks confirm supplied37+27identities and676inherited
pins, not independent review. Current user requested materials only: proposed
review0/1UNAPPROVED_UNUSED, target hostsZERO. Do not dispatch until plan approval.
Portable proposal/docs survive cloning; ignored raw,engine cache,credentials and
execution/review authority do not. Fresh-PC missing evidence must be disclosed,
not reconstructed into fake historical receipts or replenished budgets.
