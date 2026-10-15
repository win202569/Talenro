# Full diagnostic runner regression — 2026-10-15

## Approved scope and result

Following the selected-case checkpoint940be2bc8bc9f13c1d99112bafe590fe06ef3baf,
the user approved the proposed next step: one full diagnostic runner regression,
stopping on the first failure without retry. Ran the default run.tests.ps1 entry
once through Windows PowerShell5.1, without -DescendantTimeoutOnly. No source,
test, budget, native runner or security setting was changed for this execution.

Result: exit0, 31 PASS assertions, zero assertion failures. All16 assertions after
the previously blocked descendant case executed and passed. This is the full
diagnostic runner script, NOT the full repository/product suite or Go/fuzz.
The historical encoded-case failure and earlier unknown-ownership results are
not retroactively changed. The revised file-mode case's global Acceptance and
OriginalTestPassed labels remain false; RegressionPassed=true is case-specific.

## Saved evidence checks

Native receipt interval across the seven actual fixtures:
2026-10-15T11:02:08.2612666Z–11:02:27.9136568Z (Asia/Tbilisi15:02).
All seven receipts have CleanupVerified=true, ActiveAfterCleanup0 and empty final
member lists. No launch-error files were found under this run's root. These are
saved owned-job observations, not a global process scan or reopening old PIDs.

| Fixture | Key observed result |
| --- | --- |
| nonzero | Exit7 preserved; stdout/stderr markers; natural empty completion |
| timeout | TimedOut=true; forced cleanup; final0/empty |
| file-child-readiness-001 | RegressionPassed/ChildReady=true; TotalProcesses4; timeout and final0/empty |
| env | Exit0; child-only environment override; natural empty completion |
| root-exit-child-live | RootExited=true, initial remainder2; forced cleanup; natural acceptance correctly refused |
| live-sampling | Owned-running non-root child sampled; timeout; forced cleanup to0/empty |
| natural-short-child | Exit0, initial remainder2; natural empty completion, no termination requested, shutdown306ms |

Other assertions covered duplicate receipt/arm refusal, invalid source provenance,
missing/wrong/watchdog trace refusal, complete forced trace, mandatory worker
evidence and natural-shutdown rejection controls. Forced cleanup does not become
natural success: those negative gates and timed-out receipt refusal passed.

The nominal process budgets and shared cleanup bounds remain unchanged. Native
synchronous startup/observation can exceed nominal timing before a loop checks
it; this run is not proof of a hard wall-clock deadline. Child marker UTC is
captured before file IO and correlated to shutdown-time owned-running evidence,
not an exact monotonic durable-write deadline measurement.

## Local evidence identity and hashes

Ignored worktree-relative root beneath
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/`:
`runner-tests-2df767361a6342af80870c157d7e8352/`.
Output log in the same scope:
`full-runner-output-58cc7e59244b43049cc191a050a73602.log`.
This root is consumed. No retry or second-host actual run was performed.

| Local evidence relative to run root | SHA256 |
| --- | --- |
| nonzero/receipt.json | C5EA83BECB2F67CD247A42DBF9C6D1546DFA6E158FB53DC95DD20AE01976FFB6 |
| timeout/receipt.json | 207D214BBA7A465E82905AB27F7872BDA5B103AA37A712B7493EB9AC8F78BB50 |
| file-child-readiness-001/process/receipt.json | 1CDD800A831A4AC2BF2A37ABBAA029987EF73F7E8576EC5300AA19FD3DCAE2FA |
| env/receipt.json | 8017705C7BFC8F395B294F8DE629643EAD1E4AA18D5779D4BD2935327194F909 |
| root-exit-child-live/receipt.json | B431F950DE21DF2D87518F367B961804D7C076491300A1F607361E9A1114FC58 |
| live-sampling/receipt.json | 59B244E7CC981BD2CC6D52233F5A71BF5858A17278E8FD8074545E41358892DD |
| natural-short-child/receipt.json | 235A63E13AB6943FDDB8D13AC46E3949848D088C5FE994B6ACA04BA4F02C0C1D |
| Output log (scope-relative name above) | FD0A663E23C8932CC34D31DC4BE867E1D1635C30DC426C648AF4DCF123BEBBE5 |

## Recovery and remaining gates

This documentation-only checkpoint records the passing runner regression on the
existing codex/c12-b01-task9-coordinator branch. Recover the latest remotely
verified branch checkpoint, not automatically main. Read current-status.md and
docs/runbooks/repository-recovery.md. The prior tracked helper/tests and these
reports restore source/progress without ignored raw receipts/scripts/caches.
Cloning does not recreate the local receipts or authorize replay of consumed runs.

The prior137 helper checks per PowerShell5.1/7 host are results of the previous
checkpoint, not newly rerun here. This actual full-script run used5.1 only.

The diagnostic runner regression blocker and its16-unrun tail are now cleared for
this revised script on this machine. Original origin87, failed Envelope
preparation, Envelope/Pair fuzz and C11/C12/Task10/B01/production acceptance remain
open. No Go/build/fuzz, new preparation, WPR, install, security change or main
merge occurred. Next proposed work is a bounded plan for resuming the stopped
Envelope preparation using the verified runner; a new actual preparation needs
its own explicit authorization and fresh identity, not a replay of the old one.
