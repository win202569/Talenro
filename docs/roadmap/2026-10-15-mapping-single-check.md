# One nodeissuermapping check — 2026-10-15

## Approved scope and fixed execution

After diagnostic checkpoint `5177956d8319a4174920dbfc3f5a0627c9e4fd03`, the user
approved no-process entry tests followed by exactly one new compile and one
ordinary execution of internal/nodecontrol/nodeissuermapping. New mapping-check-001,
fresh offline cache/tmp, existing module cache and independently verified retained
uninstrumented Go1.26.5 copy. No dependency download, toolchain installation/patch,
version/list/origin/fuzz invocation, matrix replay/expansion, Docker/database work,
security/exit-predicate change, retry or main merge.

mapping-single-check.ps1 reuses the existing process boundary and package-cwd
adapter. Compile is `test -c -x -p=1 -count=1 -timeout=120s` for the one fixed
package,600s native bound. Its executable runs with count1, timeout120s, verbose
and run selector dot at that package directory,150s native bound. No additional
Go command or second package can be selected through this entry.

The frozen11536-file toolchain and product inventories are audited before/after.
Compilation requires copied compiler/context/fuzz-source and original package-test
evidence, plus a real single-link binary before CompilePassed. Four build artifacts
(binary/receipt/stdout/stderr) and three runtime artifacts (receipt/stdout/stderr)
are hashed; runtime artifacts are bound before completion-gate validation, even
if that validation refuses the run. Final audits and result persistence precede
any streamed PackagePassed. Acceptance always remains false. Metadata is required
for this diagnostic check but does not relax the existing native completion gate.

## Tests and review

Initial RED showed the fixed-package entry was missing. A second new negative
case showed that absent binary output could incorrectly leave CompilePassed and
ExecutionAttempted true, although the existing wrapper would refuse the missing
file before actual launch. Immediate binary validation fixed that through
RED/GREEN without changing the native wrapper or compiler/test limits.

New36 scoped assertions PASS on PS7.6.5/5.1, plus19 existing initial-query checks
on each host (55 each). They replace native
execution and large toolchain/product scans with controlled fixtures, leaving
real cwd/argument/environment/path checks, receipts/files/hashes, compile-output
parsers, schema3 gates and finalization. Fixtures include original query metadata;
no unit launches Go or a child process. Cases cover fresh/wrong/existing identity,
source refusal, build failure/missing binary/gate/evidence, runtime exit/timeout/
gate/missing tests/PASS/hooks/metadata and final product/tool/build/runtime drift.
Independent read-only review found no actionable Critical/Important/Minor issue
and supported one bounded execution only. Build cwd is the workspace root with
explicit import path, runtime cwd the package directory, matching old matrix
semantics. Review did not run tests/Go/PID queries, verify real cleanup/query
behavior, authenticate the full retained SDK/reference artifacts or judge wider
matrix/production/main readiness. No acceptance was inferred from that review.

The host-side source-binding inventory initially named two nonexistent legacy
helper paths and stopped before any actual Go invocation; it was corrected to the
ten real imported/source/test files and written once after all four unit logs were
count-checked. No implementation, test limit or native guard changed for that
preparation correction.

## One actual mapping-check-001 — PASS, not old failure resolution

Exactly one build and one package runtime occurred. Final PackagePassed,
CompilePassed, ExecutionPassed, ProductUnchanged and ToolchainUnchanged=true;
Acceptance=false, FinalizationFailed=false, Failure=null. No stopped record,
automatic retry, second package or further Go command. Final mapping-result.json,
not progress.json's last pre-runtime snapshot, determines the outcome.

| Native phase (UTC, 2026-10-15) | Start | End | Duration | Exit |
| --- | --- | --- | --- | --- |
| Build |14:33:39.3717644 |14:35:11.3940538 |92022.2894ms |0 |
| Package tests |14:35:13.1622844 |14:35:21.6733113 |8511.0269ms |0 |

Runtime output:35 top-level tests and65 subtests PASS, zero FAIL, terminal PASS;
stderr empty. Both native calls have no timeout, RootExited/NaturalShutdownVerified/
ShutdownEvidenceComplete/CleanupVerified=true, TerminationRequested=false, final
active0 and member list empty. Build natural/shutdown3ms; runtime0ms.
Runtime initial accounting total2/active0/terminated0 and one enumeration with
assigned0/returned0; both calls succeed/error0 and retain UTC/monotonic boundaries.
Build also records successful total549/active0 and one empty enumeration. Those
process counts are lifetime aggregates, not concurrent counts or image identities.

In particular, the runtime's initial member list is empty and MemberEvidence is
empty. This passing execution did not exercise OpenProcess on a residual member,
so it neither reproduces the failed initial acquisition nor proves error87 fixed.
The earlier controlled short-child sample covered a different owned-live-member
path; it cannot identify this run's already-absent transient process or the old
matrix002 unknown member. No causal/performance inference is drawn from duration
or aggregate differences between separate runs.

Final audits confirm701/701 product inventory entries unchanged, the complete
copied11536-file SDK matching its frozen manifest and seven bound binary/receipt/
stream files unchanged. Ten reviewed tooling/test/source files also matched their
pre-launch hashes after execution. Frozen reference/copy receipt identities:
`AA86DADA2C7EF1BB591C78504FB8B5C168D35D836493A574ABC37DDA84880B55` /
`DD62FDC0704E8ADEE49F77AD7E2A91D3A1CDFC6E430ABD9AADBFC74D74F04D10`.
The entry invokes only the retained copy with explicit offline/local environment,
not the original installed SDK. No new full original-installed-tree audit is claimed.

| Ignored local evidence under mapping-check-001 | SHA256 |
| --- | --- |
| plan.json |0C52B5DE20EE7E0D7AFA6285425129BA4DB742D0FBF93728E6215C9D5AB210F4 |
| mapping-result.json |C50F992CE016419510A7E4F8E72E7532D9FE3948DEDC02407061ED33BC48B56A |
| build/receipt.json |2AF0903F9EE587BC4F784D14E012A1BB5642A031F46A2844A0A655A9BA294447 |
| execute/receipt.json |7E8B92BAAC1D56462F255704E095A1B8C0372FC7FDEB8FB66495FB861341984F |
| execute/stdout.txt |C8C5BBCBF4F2E74B0186F975E78FA442BB4CB7A35B5D8985B7012DCD2D1CB602 |
| execute/stderr.txt (empty) |E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| before/after-product.json |7E328EB5D4E1A43D372F8C86DD2FE938556F226323587143BAEF3E25BD10374A |

Ignored mapping-validation-001 retains four unit logs and the ten-file source
binding. Its final-evidence-manifest.json binds16 plan/result/product/compiled/
runtime/unit/source-binding artifacts; SHA256
`7EFEE97ED49CB65CBF3A1E6BA6C74D36EE5DD4F9329564FA60E12B83BDDF7641`.
Saved native completion/query predicates, artifact bindings and product inventories
were revalidated read-only without another native invocation.
The same independent reviewer then compared the final report/status/README with
saved receipts/output/result: no factual overreach or actionable finding. It also
read/hashed the seven bound artifacts,701-entry inventories, ten source pins and
16 manifest entries, and counted36+19 PASS records per labelled host. It did not
independently certify host provenance from native receipts, rerun tests/Go/PID
queries, or repeat the full SDK audit. These limits do not grant broader acceptance.

## Recovery and remaining boundaries

Save helper/tests, this report and current-status.md on the existing development
branch. Read docs/runbooks/repository-recovery.md. Portable unit fixtures need the
current Windows PowerShell/.NET host, not private SDK/cache/raw records. Actual Go
execution on a new machine needs independently prepared verified fixed dependencies
and a separately approved fresh identity. Do not replay mapping-check-001.
Raw SDKs, binaries, caches, logs and receipts remain ignored and are not cloned.

Passing this single check would not establish why historical OpenProcess87 failed,
prove it fixed, run the remaining matrix packages/fuzz/database tests, or close
C11/C12/Task10/B01/production/main acceptance. Matrix002 remains0PASS/1FAIL/25UNRUN.
