# Matrix002 member-open failure — read-only diagnosis

## Scope and established result

The user approved continuing the proposed read-only investigation after the
matrix002 failure checkpoint `fe0a60115f3adfdf361fdafd19839a2a2e1e5bf1`.
Only saved receipts, current source/tests and primary Windows documentation were
read. No Go/native test, live capture, matrix retry, archived-PID lookup, SDK write,
collector/validator change or main merge occurred. This is not a fix or acceptance.

The failed step is specifically OpenProcess, before job-membership validation.
Receipt `matrix-check-002/01-nodeissuermapping/execute/receipt.json` still hashes
to `5D6542FF46D808E32E5C89CBD360C593DFE517C0977DCD42152644BDD25AA2BA`.
Its single listed member is positive and differs from RootId. ProcessOpen records
Succeeded=false/Error87; JobMembership=null, CoreObservedUTC=null, and neither
same-handle final-exit observation nor exit-code read exists. The monotonic open
call interval is0.3106ms. This is not the duration between enumeration and open.

## Backward trace and missing evidence

In `docs/diagnostics/isolated-fuzz/process-native.cs`:

1. Run observes root signaled, then queries job accounting and Members separately
   (lines295–301). Members (line66) queries the variable-length process-ID buffer,
   validates buffer capacity/count, and returns numeric IDs, not process handles.
2. InspectMembers (line169) opens each current-list member with unchanged
   QUERY_LIMITED_INFORMATION|SYNCHRONIZE rights. Failure captures last-error
   immediately, before timing bookkeeping (lines181–185). Error87 therefore
   belongs to the failed open in this saved row, not the later membership call.
3. Ownership, state, exit and image reads occur only after successful handle open
   and membership verification. Their default false/zero/null values are not
   independent observations. In particular, Owned=false does not prove foreign
   ownership; ImageError=0 does not prove an image query succeeded.
4. The unknown non-root row makes ShutdownEvidenceComplete=false. The bounded
   natural-wait loop is conditional on complete evidence (line308), so it is not
   entered. The runner requests owned-job termination and verifies empty cleanup;
   the PowerShell completion guard refuses the receipt. Its6ms shutdown does not
   demonstrate expiry of the500ms natural-observation budget.

The receipt does not record start/end timestamps or raw headers for the Members
query, nor timestamps for the accounting snapshots/root wait completion. There
is no successful handle acquisition or live member history (LiveSampling=false).
Consequently it cannot locate a member exit relative to enumeration, identify its
image/creator, or prove whether accounting/list samples describe one instant.
The source uses distinct API calls; treating them as an atomic snapshot would be
an unsupported assumption. Current collector SHA256:
`7521EAF32B63939F91E0F0450471861083CCDE471381B293BA58C3759ACDEE2E`.

## Comparison without causal overreach

| Saved observation | Evidence available | What it does not prove |
| --- | --- | --- |
| Matrix001 check01 | TotalProcesses2, initial active0/list empty, natural completion; no member rows | That OpenProcess succeeded on the corresponding transient member |
| Matrix002 check01 | TotalProcesses2, initial active1/list length1, OpenProcess87, no ownership/state proof | A surviving product child, ownership loss or member identity |
| Earlier preparation001 origin | QueryError87 with no separate API phase metadata | That its error arose in OpenProcess rather than membership validation |

Matrix001 check01 runtime receipt SHA256:
`87BE6998E03BA64437B066978E8D2232AC63F2C41F78A3588128C8CEA88C1FFD`.
Both matrix source inventories have25 nodeissuermapping files with zero path/hash
differences and the same frozen toolchain reference/copy identities. Across their
701-entry source inventories, only the separately approved statecontract boundary
test differs. They have different fresh caches, binaries and execution timing;
source equality is not a controlled proof of an OS or timing cause.

The prior batch-acquisition repair already opens/verifies all listed members
before any final wait (lines175–217). Matrix002 has only one listed member and its
open fails before any such wait. Waiting on an earlier member in the same batch
therefore cannot explain this observation. Previously passing final-exit fixtures
start with retained, owned handles; they do not exercise a lost initial acquisition.
The member-query fixture covers successful open followed by membership failure,
not this failed-open case. These are static coverage observations; no test reran.

## Primary documentation and hypothesis boundary

Windows documents process identifiers as lifetime-bound, while acquired handles
remain valid until closed even after termination. This makes lifetime-sensitive
handle acquisition a plausible investigation direction, not proof that this member
had exited. See [Process Handles and Identifiers](https://learn.microsoft.com/en-us/windows/win32/procthread/process-handles-and-identifiers).
The job-list API returns identifiers; it does not supply retained process handles.
See [JOBOBJECT_BASIC_PROCESS_ID_LIST](https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_basic_process_id_list).

Error87 means ERROR_INVALID_PARAMETER, not a dedicated process-exited result.
OpenProcess documents a zero-PID case for that error; this saved member is nonzero,
so that example does not identify this cause. Do not infer access denial, request
broader privileges or ignore the error from its number alone. See
[system error codes](https://learn.microsoft.com/en-us/windows/win32/debug/system-error-codes--0-499-)
and [OpenProcess](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-openprocess).
Job accounting also distinguishes lifetime TotalProcesses from currently associated
ActiveProcesses; neither count identifies an image. See
[job accounting](https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_basic_accounting_information).

## Next decision and recovery

The failed acquisition is established; its OS/process cause remains unresolved.
Further historical rereading cannot manufacture the missing timing/handle evidence.
The next useful bounded increment would specify enumeration/accounting API timing
and raw-result diagnostics, preserving the original error and every existing
ownership/exit/cleanup predicate. Test such metadata separately before considering
one controlled short-lived-process observation. That implementation/live capture
requires its own approved scope; none is performed or implicitly authorized here.
Do not replay002, raise timeouts, add error/image exemptions or repeat until PASS.

Recover the development branch and read current-status.md and repository-recovery.md.
This report preserves the findings without requiring private logs to understand
the next decision. Raw receipts/toolchains/caches remain ignored and machine-local.
Matrix002 stays0PASS/1FAIL/25UNRUN; wider acceptance and main remain blocked.
