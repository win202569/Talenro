# Root native verification — bounded attempt result

Status: STOP/NOT_PASSED; global NOT_ACCEPTED; counters44/34/6/217 unchanged.

User approved the paired root-verify-completion plan after source checkpoint
f51f018196f573909a2087af49d6a7ac33eef50a. Remote matched before execution
(a55c42). Syntax check e66e29 passed; fixed paths,three tool/source pins and
four module hashes matched the v3 baseline before the single Go invocation.

## Observed result

Invocation7571a5/session46364,final95c553: native pinned Go1.26.5,
arguments `mod verify`,root worktree,real existing module cache,offline settings,
GOMAXPROCS2,unchanged bounded owner,no phase Session or collector.

- Elapsed300059.2263ms; Started=true,TimedOut=true,EndReason=DEADLINE.
- ExitConfirmed=true,ExitCode=-1,Failure=null. This is deadline termination,
  not a successful native result. The wrapping PowerShell exit0 is not Go success.
- stdout0 bytes,stderr0 bytes; both SHA256
  E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855.
- All four root/devtools go.mod/go.sum hashes unchanged before/after.
- One attempt consumed; no retry,build,test,Docker,C12 or collector launched.

The command did not finish within the approved300s budget. Empty output and
timeout establish neither cache corruption nor its cause. Prior two-module
phase timing cannot predict all-module completion. No within-run attribution
was collected. Offline settings are not an atomic read-only cache snapshot.
Even a future native success would skip fully uncached modules and would not
prove dependency availability,buildability,tests or C12 acceptance.

## Portable evidence and next boundary

portable-c12-root-verify-completion.json contains exact invocation/packaging
recipes,preflight,exclusive claim,process and stage receipts,raw outputs,
unchanged owner/common source and this report. Each decoded entry is checked
against original length/SHA256. No binaries,module contents,caches or secrets.
All old evidence is retained. Author inline review only; no independent review
or wider acceptance is claimed.

Resume from the latest development-branch checkpoint and repository-recovery.md.
The live budget is exhausted. Recommended next step is read-only review of
existing evidence and pinned verification flow to propose a narrower diagnostic
question and finite budget; do not automatically rerun or increase the timeout.
New runtime actions or source repairs require a separate approved scope.
