# P2-B controller causality reverification — failed GREEN, frozen unfinished

Date: 2026-10-19. Branch codex/c12-b01-task9-coordinator.
Execution base: 62f7ad23058fc13dbea1387a41cbe6fab5da00b8.
User approved Native/new2hosts+4children+conditional1readonlyreview after that
checkpoint, including first-GREEN oracle-revision exception. No product RED claim.

## Execution and evidence

New leaf: docs/diagnostics/isolated-fuzz/p2b-controller-causality-reverification/.
Five execution inputs; old owned-run/fixture/seam unchanged. The two characterized
cause functions were statically compared verbatim against the diagnostic; both
new scripts parsed without execution before Prepare. No extra baseline/targets.

Approved-plan snapshot SHA256:
5264D0FBE0A26258E4DDC1F9A54E2C06786E1CFEAD4A141B8458CDDB77BD2FAC.
Frozen manifest SHA256:
EE04A4A23F93B6CDB0011CB058DC515ED69C28C06E4A747CD0799B5B5C5671F1.
1454 mapped pins: 1434 inherited plus 7 diagnostic raw, 6 diagnostic portable,
1 diagnostic execution report, 5 new inputs and 1 approved-plan snapshot.
All pins intact before and after the unique GREEN run. No frozen input edits.

GREEN host PID40880, 1556ms, exit2; exact owned exit and complete concurrent
streams. No timeout, host termination, parent/cleanup error or outer rescue.
Parent capture correctly reported INCOMPLETE and stopped. Four exact PASS:

1. PostStartReceiptFault
2. PrimarySurvivesFinalReceiptFault
3. NormalExit
4. Timeout

Fifth group PreStartOutputFault failed: Started=false, PrimaryPresent=true,
Unassociated=false (expected true), OwnedExit=false, ExitCode=null,
TerminationRequested=false. These fields are observations, not proof of an
unrecorded Start or its absence by independent association evidence. Root cause
of the unexpected Unassociated value has not been established. No new probe,
framework/source investigation, correction or repeat was performed after stop.

Complete trace comparison differences: Groups array count;
Observations.PreStartOutputFault.Unassociated kind; Passed kind; Error kind.
The full run is not accepted merely because four groups passed.

Fault child14272/exit-1, normal child18040/exit0, timeout child9820/exit-1:
all three original owned Process objects exited without rescue. Pre-start group
created an unstarted object, with no additional child slot or Start call in the
new test. Real fault returned two Error-list records with different I/O causes;
the original result matched first/final anchors, primary-replaced/swapped/
cleanup-replaced/null/missing/deep-nine negatives all rejected. Serialized booleans
do not restore runtime references or retroactively reclassify old GREEN.

## Budget and boundaries

Spent new1host+3children=4 of6 new PS7 processes. Remaining new1host+1child+
conditional1review FROZEN. FINAL not consumed/run; no independent reviewer.
Old containment remaining1host+3children+1review still FROZEN; all old failures,
diagnostic success and other allowances unchanged. I1 remains unaccepted.

No product changes, old suite rerun, compile, install, security changes, main
merge, real handshake/capability or live action. SourceVerified,
LiveActionsPerformed, RuntimeApproved, FormalAcceptance and RealGateImplemented
remain false. This is an unfinished development checkpoint, not overall progress
acceptance. Do not use final or review slots as debugging budget.

## Preservation and next decision

Portable summary retains invocation/children/trace/details, complete stream bytes
and raw artifact hashes, not just a narrative. Ignored packet is preserved in
place; no raw is committed. The approved plan snapshot stays byte-identical even
though the tracked plan's status records this outcome.

PowerShell/manual bookkeeping replaces Bash-only skill wrappers; approved scope
overrides generic extra full-suite/task-done/fix-pass/deletion instructions. The
new packet accepts only its ledger before Prepare; prior receipts forbid reuse.
No independent review is claimed. Next recommendation: separately authorize
read-only analysis of the fifth-group observation; any assertion/code changes or
new execution require a reviewed scope and explicit finite budget. No automatic
fix, retry, old-result correction or remaining-slot unfreeze follows.

Fresh-computer recovery uses this development branch, current-status and the
recovery runbook. Clone restores tracked sources and portable observations, not
ignored raw, fixed runtime/SDK, CLR Process/ErrorRecord identities, credentials
or execution rights. Missing raw is not reconstructed from summaries.
