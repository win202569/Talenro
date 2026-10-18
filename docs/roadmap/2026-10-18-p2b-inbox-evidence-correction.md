# P2-B inbox portable evidence — authorized lossless correction

Date:2026-10-18. Approval followed development checkpoint
`19c567a18dfbe6490bc77cdd46ad1d48cdadf3a7`.
Scope:only derived-summary integer correction and readonly comparison;
zero product targets,managed compilations or reviewers.
Status: **BOUNDED_VERIFIED_EVIDENCE_CORRECTED**.
This completes the specified evidence correction,not formal/runtime acceptance.

## Cause and exact changes

The parent had passed raw JSON through JavaScript JSON.parse/Number/stringify.
Numbers above the safe-integer range were rounded before reaching the portable
file. Original raw trace text/hashes were always exact. Eight JSON number tokens
in portable-summary.json were restored directly with apply_patch; no numeric
field became a string and no raw record was rewritten.

| Path in each OriginalTrace and FinalOriginalTrace | Incorrect token | Restored original token |
| --- | --- | --- |
| Synthetic[0..2].After (three each) |9223372036854776000|9223372036854775808|
| Synthetic[3].After (one each) |18446744073709552000|18446744073709551614|

Do not roundtrip these JSON integers through JavaScript Number in future.
Use exact token/BigInteger-aware handling,or explicitly documented decimal
strings in a separately chosen schema. Plain JSON numeric text can be exact
even though a consumer's binary64 representation is not.

## Readonly verification

A direct System.Text.Json.JsonDocument traversal compared all nine embedded
evidence subtrees with the unchanged packet files:Invocations,ProductSeal,
RedTrace,FullTrace,OriginalTrace,CapacityTrace,SequenceTrace,FinalTrace and
FinalOriginalTrace. It compared JSON kinds,ordered property names,array order/
length,decoded strings,and exact raw numeric/Boolean/null tokens. It never
converted number tokens into JavaScript Number or floating point.
Before correction:exactly8 numeric differences;after correction and metadata
updates:0differences across all9subtrees. No other evidence field differed.

All26packet files were hashed before and after and the entire sorted name/hash
inventory matched,including streams,traces,manifest,seals,source snapshots and
review package. All1326frozen mappings and eleven saved portable input hashes
were also rechecked. Product remains exact47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC.
No product import,probe,test,controller invocation or compilation was used.

Summary SHA256 before correction:
`F4FBD0B0B154063721AA8100D5F3027CF2E98E42C6A5BE110A9A27ADE0C0A790`.
Corrected summary including explicit disposition/correction metadata:
`7E64C8EED19BEFC8C75775E8077A79568ABBEA06A6DD3FD035E1CBFE12D2B185`.
The old inaccurate checkpoint remains recoverable in Git;no history rewrite.
The original raw packet remains ignored and is not included in the commit.

## Review disposition and recovery

[Original review and every disposition](../diagnostics/isolated-fuzz/p2b-inbox-tree-copy-repair/review.md)
remain preserved verbatim. Historical Important1 is now resolved by the user's
specifically authorized correction and parent's readonly verification. There was
no re-review,replacement reviewer or change to the original review verdict.
Minor2 remains accepted/deferred:Count.Probe scalar ETS isolation lacks direct
retained/absent assertions. Existing scalar guards/reference checks support
narrower coverage;no frozen test modification or broader coverage claim.

All five saved runs remain EXPECTED_WITHIN_SCOPE;two combined runs each44PASS,
specified mutants detected. New5/5hosts and1/1reviewer are exhausted;oldoriginal
2spent+3FROZEN anddiagnostic1spent unchanged. No main merge,live use,other runtime,
concurrency,full repository/fullC12 or formal acceptance follows.

Retain the historical-object/restore-timing/private-tree/environment/readonly-
reproduction qualifications in review.md. Preserve worktree and all raw records;
no generic extra suites,fix pass,re-review or cleanup. The cost is that broader
properties remain unverified;this approval did not widen implementation rights.
Fresh-computer recovery starts from the existing development branch,reads this
record/current-status/repository-recovery,and does not automatically replay any
host or reconstruct missing raw evidence. Dependencies,credentials and genuine
environment acceptance remain separate preparation.
