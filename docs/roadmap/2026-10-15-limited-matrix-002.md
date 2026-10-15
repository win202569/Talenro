# Limited matrix002 — 2026-10-15

## Approved execution

After the bounded scanner repair at `3e90ec7ce539cc75f163a126d14326e69b7ee90b`,
the user separately approved one new `matrix-check-002` attempt. Use the existing
unchanged matrix-check.ps1 and native process boundary, independent offline cache
and the retained uninstrumented Go1.26.5 copy. First failure stops; no automatic
retry, historical-identity replay, dependency download, SDK patch/deletion,
Docker/database integration operation, main merge or broader acceptance.

The26 checks mean22 complete package tests, selected authority tests matching
`^(TestBoundRead|TestCoordinator)`, store compilation only, and original
Envelope/Pair fuzz samples. Maximum24 builds/25 runtimes. Build budget600s;
ordinary tests120s/native150s; fuzz5s/parallel1/test30s/native60s. Relative test
fixtures and source checks run at package cwd. MatrixPassed is this scope only;
Acceptance remains false even if all26 pass.

## Result

The attempt stopped at check01/nodeissuermapping's execution completion gate.
No check02 directory was created; no retry or later runtime/fuzz invocation.
Final result:0PASS/1FAIL/25UNRUN, MatrixPassed=false, Acceptance=false,
FinalizationFailed=false. The earlier matrix-check-001 remains
12PASS/1FAIL/13UNRUN and is not rewritten by this new attempt.

Check01 compiled successfully, UTC13:36:56.6165856Z–13:39:22.9056806Z,
146289.095ms, exit0, no timeout, verified natural shutdown/cleanup. Its test runtime
UTC13:39:30.4992951Z–13:40:13.6924834Z,43193.1883ms, exit0 and no timeout.
Output has35 top-level PASS records,65 subtest PASS records, no FAIL and terminal
PASS; stderr is empty. This is test-output success, not an accepted matrix check.

The runtime receipt has RootExited=true, ActiveAtRootWait=1, one listed member,
TotalProcesses=2 (aggregate only). Its OpenProcess call returned error87;
ownership/state/exit/image were not established: Owned=false, State=unknown,
ExitCodeRead=false, ImagePath=null, and no membership or final-handle observation.
Owned=false here is the unverified default, not proof the member was outside the
job. The existing collector marks evidence incomplete and does not enter its
natural-observation loop when a non-root member is unverified. It requests job
termination instead. Recorded NaturalShutdownVerified/ShutdownEvidenceComplete=false,
TerminationRequested=true; final CleanupVerified=true, ActiveAfterCleanup=0 and
MembersAfterCleanup empty. ShutdownElapsedMs/NaturalObservationMs both6ms, not an
exhausted500ms natural-wait allowance. Completion guard correctly refuses this
receipt with `Natural shutdown not verified; forced cleanup is not success`.

This identifies the failed evidence step, not its underlying OS/process cause.
An exited/disappearing member or observation timing is not proven; neither is an
actual surviving product child, ownership loss, business-logic failure or relation
to historical origin001/error87. No historical PID was queried and no guard was
relaxed. No source/test/framework fix is made in this execution-only checkpoint.

| Local evidence under matrix-check-002 | SHA256 |
| --- | --- |
| plan.json |2EAE87F1111741F61E12A1A2110C6D1EED7660DF1C07508B1916404DCE541605 |
| stopped.json |DC2947616AA3AA85F652B309E5DC41DC412D215CA66EB7E826D2BF659B988271 |
|01-nodeissuermapping/build/receipt.json |D942992168FC3752AA4EA32F6870F20421928D2D453410F683FF3CF6A76E5BAD |
|01-nodeissuermapping/execute/receipt.json |5D6542FF46D808E32E5C89CBD360C593DFE517C0977DCD42152644BDD25AA2BA |
|01-nodeissuermapping/execute/stdout.txt |E0CE3F0929D2229E87DE5A44A447C367224A73094F1AAFE496B435CB6FD1F9A9 |
|01-nodeissuermapping/execute/stderr.txt (empty) |E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| before/after-product.json |7E328EB5D4E1A43D372F8C86DD2FE938556F226323587143BAEF3E25BD10374A |
| matrix-result.json |35AFC684857F922723B347B644C5F4F23E747EE0F4A0E35656F6E4776633091F |
| original-source-postcheck.json |7D130C9B10FAE95717F7240870669A30EF4D8C31411B9D161C556227905DF398 |

Finalization passed:701/701 source inventory entries unchanged; the complete
11536-file copied-toolchain audit matched its frozen reference; four compile
artifacts remained bound. Reference/copy receipt hashes are
`AA86DADA2C7EF1BB591C78504FB8B5C168D35D836493A574ABC37DDA84880B55` /
`DD62FDC0704E8ADEE49F77AD7E2A91D3A1CDFC6E430ABD9AADBFC74D74F04D10`.
There were exactly two native attempts, successful compilation and the refused
runtime. The failed runtime is not in the driver's success-only RuntimeEvidence
array; its raw receipt/streams are preserved and hashed above. Aggregate counts
do not establish individual process identity. The final result, not progress.json
which precedes failure handling, is authoritative. Final original-source postcheck
also passed:11536 files/218649457 bytes matched the frozen original manifest,
161837ms, completed UTC13:45:00.0531864Z. No original SDK write or invocation.

## Next bounded investigation

Use saved receipts and collector source to examine the member-open failure and
its evidence limitations. No current PID lookup can recover historical identity.
If additional controlled fixtures or new live capture are necessary, specify and
approve their scope separately; do not replay002 or silently weaken the unknown-
member refusal. This checkpoint implements no diagnostic collector redesign.

## Recovery and remaining gates

Necessary progress records are preserved on `codex/c12-b01-task9-coordinator`,
not main. Read current-status.md and docs/runbooks/repository-recovery.md.
Raw receipts, caches, binaries and retained SDK copies are ignored and not restored
by Git. Another computer needs independently prepared fixed dependencies and a
separately approved fresh validation identity; do not replay002.

Short fuzz observations cannot establish adequate coverage or absence of defects.
This matrix does not resolve historical matrix3/origin001 causality, execute store
database tests, establish production readiness, or close C11/C12/Task10/B01/main
acceptance gates.
