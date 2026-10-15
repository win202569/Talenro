# Revised descendant timeout case — 2026-10-15

## Approved implementation and limits

The user approved migrating only the original `real owned descendant cleaned on
timeout` case to the existing readable-file helper, adding a single-case switch,
helper checks/review, and exactly one actual selected execution. No full runner
suite, Go/fuzz, retry, WPR, security change, budget relaxation or main merge.

`descendant-timeout.ps1` shares the actual evaluator between tests and execution.
It requires explicit file mode, the existing strict nonce/non-root child PID,
owned-running member/timestamp evidence, timeout, verified zero accounting and
empty final membership, plus a parsed unsigned TotalProcesses >= 2. These retain
the original numeric/timeout/cleanup conditions and strengthen intended-child
proof. Invocation uses one fresh helper run, then reloads fixed scoped saved
plan/receipt paths with single-link checks and saved-root binding. The result is
saved as regression-result.json. Global Acceptance/OriginalTestPassed stay false;
RegressionPassed describes only the revised case.

`run.tests.ps1 -DescendantTimeoutOnly` selects this case before all other actual
fixtures and returns. The default suite replaces only this case, not the other
fixtures; its encoded childCode/parentCode remain for later orphan/live tests.
Native runner and its nominal 2s budget are unchanged. Synchronous calls and
cleanup can exceed that nominal duration; there is no hard wall-clock guarantee.

## Test-first checks and review

The new test first failed because its shared evaluator was missing. After
implementation, 19 new assertions passed on both PowerShell 7 and Windows
PowerShell 5.1. They exercise actual marker files with synthetic receipt inputs:
positive conjunction, false global acceptance, invalid/too-small process counts,
missing timeout/cleanup, nonempty cleanup, truncation, wrong nonce, parent PID,
late marker, unowned member, absent marker, encoded-mode refusal, consumed-root
refusal before launch and parsed single-case parameter. No valid native fixture
is launched by these unit checks, and synthetic receipts are not actual evidence.

Existing file-readiness19, child-readiness28, launch-timing8, full-parent12,
tree-stage9, entry-startup12, runner-regression21 and shutdown-state9 assertions
also passed on each host: 137 total per host including the new19.

Independent read-only review found no Critical/Important defects and supported
the one selected execution/unfinished checkpoint, not main merge. Minor remaining
coverage gaps: the AST unit test checks the parameter rather than structural
early return; saved-root mismatch, malformed saved JSON and linked saved-input
refusal lack direct unit tests. Static review verified selector isolation; the
actual saved-directory inventory below also confirms this execution's isolation.
Full-suite success, exact durable-ready timing, original encoded repair, causal
attribution and broader acceptance were deliberately not judged or claimed.

## One selected execution — PASS

Ran ONCE with Windows PowerShell 5.1, using the selected entry above. Wrapper exit0
and exactly one PASS: `real owned descendant cleaned on timeout`.
Native UTC interval10:57:45.7001067–10:57:47.8233197 (Asia/Tbilisi14:57).

- RegressionPassed=true, ChildReady=true, FixturePassed=true. Native TotalProcesses4.
- Marker nonce989af10f2aaa46798f8d907fd9514824 matches the saved plan. Child PID25300
  is distinct from root24184 and matches exactly one owned/running PowerShell
  member, QueryError0, successfully read exit259, core observation10:57:47.7717052Z.
  Marker UTC10:57:46.8402261Z is1140.1194ms after saved native StartUTC.
- TimedOut=true, root not exited, forced termination requested; CleanupVerified=true,
  ActiveAfterCleanup0, final members empty, ShutdownElapsedMs107. Native exit57345
  is expected forced termination for the20s fixture, NOT natural product success.
  NaturalShutdownVerified=false. Native interval2123.213ms includes cleanup.
- Creation bracket146.4417ms (budget39.8612–186.3029ms); assignment0.3594ms;
  resume0.2102ms ending at budget186.8742ms. These are instrumented brackets,
  not pure API CPU time or an antimalware causality measurement.
- Stdout/stderr empty; Acceptance=false and OriginalTestPassed=false throughout.
- The fresh runner-tests root contains only file-child-readiness-001; no nonzero,
  timeout, env, orphan, live-sampling or natural-case directories were generated.
  No second selected execution, full suite, Go/fuzz, trace or live old-PID query.

Readiness UTC is captured before file IO; native owned-running evidence is sampled
at shutdown. This is correlated script-entry/ownership evidence, not exact
monotonic proof of durable file completion by1140ms or the deadline. Earlier
encoded failures remain historical failures; a changed fixture is not their
retroactive repair or a controlled causal comparison.

## Evidence and fresh-computer handoff

Ignored worktree-relative root:
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/runner-tests-10fb394189874b1680d7feff8cfe7b9c/file-child-readiness-001/`.
This complete identity is consumed; it is distinct from the earlier independent
root-level file-child-readiness-001. Do not replay either after cloning.

| Local evidence | SHA256 |
| --- | --- |
| regression-result.json | 8409376187653847A816F5AE86D25728098F113C8656F7F87BDD818317F9D2F3 |
| result.json | A004C62885D56CAA5D533991E271E2540E3D8865CFED47C4DC1A07900D88497E |
| process/receipt.json | 9250874E81823F00EA62D8D799BA20F0BDF32F0E83C7842A630B19D9A365F308 |

Necessary helper, tests, selector and this summary are tracked on
codex/c12-b01-task9-coordinator. Recover from that branch's latest remotely verified
checkpoint, then read current-status.md and docs/runbooks/repository-recovery.md.
No ignored raw receipts, generated scripts, trace, caches or binaries are needed
to restore source/progress; cloning does not restore those machine artifacts.

The revised single-case blocker is locally addressed by this run, but the full
runner suite has NOT been rerun; the16 checks after the historical failure remain
UNRUN. Earlier failure remains recorded. Original origin87, Envelope/Pair,
C11/C12/Task10/B01/production and main gates stay open. Next proposed step is a
separately approved full runner regression, stopping on its first failure without
retry; do not start Go/fuzz or infer permission from this single-case approval.
