# One readable-script readiness diagnostic — 2026-10-15

## Approved bounded change

The user approved adding an independent script-file mode to the existing helper,
then one fresh actual sample after helper checks. Preserve default encoded mode,
child behavior/readiness proof, native runner,2s budget, ownership/cleanup gates
and all earlier failures. No execution-policy override, security exclusion,
disabled protection, new performance recording or Go/fuzz. Any policy refusal
would remain failure, without a retry or bypass.

child-readiness.ps1 adds optional Boolean UseScriptFiles (default false) to the
planner/root/invoker flow. Explicit true requires a fresh file-child-readiness-NNN
identity; default mode still requires child-readiness-NNN. The common child script
is unchanged. File-mode parent launches it hidden with normal -NoProfile -File
arguments; the native parent arguments are likewise only -NoProfile -File path.
No policy flags are added to either command. Existing encoded parent generation
and its argument encoding are retained for default mode, not silently migrated.

Write-ChildReadinessScripts materializes fixed parent.ps1/child.ps1 paths under the
validated owned root. It prechecks both destinations, uses CreateNew/no overwrite,
UTF8 BOM for Windows PowerShell5.1 Unicode paths, and disposes its writers/streams.
Materialization is inside the invoker's stopped-envelope try block, before the
unchanged native-runner timer starts. Partial preparation consumes the root rather
than relaunching. The evaluator accepts the explicit mode for root validation
but otherwise keeps the same strict nonce/non-root PID/owned-running proof,
timestamp checks, timeout/empty cleanup conjunction and false acceptance labels.
Old serialized encoded plans lacking the optional flag retain default semantics.

## Test-first verification and review

New file-readiness.tests.ps1 first failed at explicit mode/fresh identity/budget
support. After implementation,19 assertions passed on PowerShell7 and Windows
PowerShell5.1. They exercise readonly planning, distinct identities, unchanged
encoded roundtrip, normal command arguments, apostrophe/Unicode/spaced paths,
parent/child AST parsing, actual exclusive file materialization, child BOM,
same-host child-prefix nonce/self-ID write, no overwrite, substituted destination
refusal and the existing diagnostic evaluator's ownership requirements.

No valid observed process is launched by these helpers. The child prefix runs
only inside the unit host, excluding its sleep; a synthetic receipt exercises
evaluation, not a native-ownership acceptance claim. Unicode test characters are
constructed from code points so the test source itself has no PS5.1 ANSI ambiguity.

Existing helper checks passed on both hosts: child-readiness28, launch-timing8,
full-parent12, tree-stage9, minimal-entry12, runner-regression21, shutdown-state9.
Total118 per host including new19. No full runner/product/Go/fuzz suite was run;
the original descendant regression and16 unrun checks remain unresolved.

Independent read-only review found no Critical/Important defect and supported
one actual diagnostic and an unfinished checkpoint only. It ran no tests, native
calls, PID queries or trace access. Its minor substituted-path coverage finding
was fixed: the check now uses a fresh root with neither script present, so generic
duplicate refusal cannot satisfy it;19 checks reran PASS on both hosts. Remaining
minor limit: BOM is asserted on the child only, although both use the same writer.
Its excluded actual policy compatibility was subsequently observed in the one
sample below; strict durable-ready timing, causal attribution, original acceptance,
main merge and unauthorized extra samples remain explicitly unestablished.

## Single actual observation — diagnostic fixture PASS only

file-child-readiness-001 ran ONCE, UTC10:47:08.8904028–10:47:10.9112642
(Asia/Tbilisi14:47). Saved plan, scripts, readiness marker and native result were
read back without reopening historical process handles.

- ReadinessRecordObserved=true, ChildReady=true, FixturePassed=true; wrapper exit0.
- Marker UTC10:47:09.5061901,615.7873ms after native StartUTC. Its nonce exactly
  matches the saved plan, and its own PID33784 matches exactly one non-root
  owned/running powershell.exe member, QueryError0, exit code successfully read259,
  core observation UTC10:47:10.9063560. No aggregate-count substitution was used.
- TimedOut=true, root not exited, TotalProcesses4. Forced owned termination was
  expected for the20s parent/child fixture: CleanupVerified=true, accounting0,
  member list empty, ShutdownElapsedMs12. Exit57345 is a forced-termination code,
  NOT natural product success; NaturalShutdownVerified=false.
- Native launch brackets: creation21.2584ms (budget6.0109–27.2694ms), assignment
  0.2421ms (27.2705–27.5128ms), resume0.1166ms (27.5131–27.6297ms).
- Stdout/stderr empty. Acceptance=false and OriginalTestPassed=false throughout.
  No second file sample, policy override or performance recording ran.

The original UTC/flush limitation remains: child UTC is captured before file IO,
and owned-running evidence is sampled at shutdown. The successful diagnostic
correlation is not an exact monotonic proof of durable file completion at615ms.

Compared with the prior separately recorded encoded sample's2082.4830ms creation
bracket, this readable-file sample was much faster and supplied intended-child
evidence under the same nominal budget. This supports further investigation of
script delivery as a relevant factor, NOT proof that encoding alone caused the
delay or that Defender scanning was eliminated. File reads, preparation before
timing, effective policy, caches, system load and tracing conditions differ; there
is no randomized causal A/B trial or antimalware-path trace for this sample.

## Local evidence and portable recovery

Ignored worktree-relative root:
.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/file-child-readiness-001/.
SHA256:

| Local evidence | SHA256 |
| --- | --- |
| process/receipt.json | 1F415E0225670D958B59CBC46BBD759460722C076C36289E44FCCB07786DE9D3 |
| result.json | 8EF2B2558542621A2A64D0F610939A9272F92AEB7D14066E36EB13E999A3E13A |
| parent.ps1 | D0231DF4E3859FF5865B74A8E367ED770AFA0164D6442D5589E5B6367E17B674 |
| child.ps1 | D3AEECE2D64BEE3174205BB9147C438CF5D1559B30B567E2FCB271DDE34DE901 |

Necessary helper/tests and this public summary are tracked on the latest
codex/c12-b01-task9-coordinator, not automatically main. Read current-status.md
and repository-recovery.md on a new computer. Helper fixtures regenerate without
old Go caches or raw receipts. Cloning does not recover ignored scripts/receipts
or the sensitive earlier ETL. Consumed file-child-readiness001 may not be replayed
even if absent after cloning; all previous sample identities stay consumed too.

No original test was replaced, full suite accepted or production/main gate cleared.
Original descendant failure/16 unrun checks, origin87, Envelope/Pair and
C11/C12/Task10/B01/production gates remain open. Adopting readable scripts into
the original regression suite requires a separately approved test change and its
verification; this successful independent sample is evidence, not that approval.
