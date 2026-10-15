# Matrix003 acquisition boundary — read-only analysis

## Scope and finding

User approved continuing the proposed read-only comparison after checkpoint
22c1a0fc960b666a7de1a481dd90408561d8d147. Read saved receipts, collector/tests and
Microsoft primary documentation only. No Go/native invocation, test, historical
PID query, SDK audit, new observation, implementation or gate change occurred.

The established failure is an initial member-handle acquisition failure, not a
failed final wait. There is no newly established OS-level cause or verified code
defect. Matrix003 remains2PASS/1FAIL/23UNRUN; Acceptance=false. Further matrix
repetition is not the recommended next diagnostic step.

## Saved samples are different coverage classes

| Sample | Initial active/list | Acquisition | What it establishes |
| --- | --- | --- | --- |
| Matrix003 first four native attempts |0/empty each |No member open attempted |Empty-job natural completion only |
| mapping-check-001 runtime |0/empty |No member open attempted |Single-package success, not acquisition success |
| short-child-observe-001 |2/two |Both open and membership succeed; retained handles signal exit0 |Owned live-after-root members can finish within the existing shared window |
| Matrix002 check01 runtime |1/one |OpenProcess87; no membership |Initial acquisition failed; enumeration timing unavailable |
| Matrix003 check03 build |1/one |OpenProcess87; no membership |Initial acquisition failed after a successfully timed list query |

The short-child rows identify that sample's PowerShell/conhost images only; they
do not identify either historical failed member. All listed failures are non-root
rows. Matrix003's failed ID is positive, fits UInt32, matches the returned list,
and was not truncated. One list query returned assigned1/returned1/capacity16;
there was neither a resize retry nor an earlier member wait in this one-row batch.
The four preceding successes do not isolate why another package/phase failed.

## Backward trace through unchanged collector

Current process-native.cs references:

1. Lines306–314 wait for the root, then separately query accounting/list. These
   calls are not one atomic process-lifetime snapshot.
2. Lines67–94 parse the list as two32-bit counts followed by pointer-sized IDs.
   The observed complete one-entry list fits the buffer; no evidence indicates
   count truncation, UInt32 narrowing or root-ID substitution in this sample.
3. Lines194–206 validate the numeric ID, call OpenProcess with0x101000
   (limited-query plus synchronize), and capture last-error immediately before
   timing. Saved ProcessOpen.Error87 is not the later IsProcessInJob error.
4. Lines207–225 only perform membership/state/exit/image reads after acquisition.
   All those unknown/default fields in the failed row remain missing evidence.
5. Lines226–239 acquire the bounded batch before final waits and keep handles
   through that inspection call. With just one failed initial open, waiting on an
   earlier member cannot explain this sample.
6. Lines323–348 reject unverified non-root evidence before natural waiting, then
   request job termination. The saved2ms shutdown is not a500ms wait timeout;
   final0/empty cleanup does not retroactively establish natural completion.

List end to open start is0.1961ms; the open call itself spans0.3210ms. Neither
interval records a process OS exit or an acquired handle. A disappearing process
is a plausible hypothesis, not a proven cause. Error87 alone is not proof of
access denial, and does not justify broader privileges or an error exemption.

## Why existing alternatives do not close this gap

The live-sampling path (lines299–303) stores observations, not a retained handle
registry. InspectMembers closes its handles at line239; final inspection opens
the current IDs again. Merely enabling live sampling therefore does not guarantee
that a later failed open has an earlier same-handle identity/exit proof. It also
changes sampling behavior and is not performed here.

Static test inspection: member-query.tests.ps1 retains a live child creation
handle and tests successful open followed by invalid-job membership error6,
successful ownership, and nonmembership. It does not reproduce this initial
OpenProcess87. The short-child sample validates successful acquisition/final
waiting, not acquisition failure. No test was rerun or silently expanded.

Microsoft documents process IDs and handles as different lifetime mechanisms;
acquired handles remain valid until closed after termination. The job-list
structure returns identifiers, not retained handles. See
[process handles and identifiers](https://learn.microsoft.com/en-us/windows/win32/procthread/process-handles-and-identifiers),
[job process-ID list](https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_basic_process_id_list)
and [OpenProcess](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-openprocess).
These contracts motivate investigating the lifetime-evidence boundary; they do
not diagnose the saved nonzero member's error87.

Completion-port messages alone are not an authoritative replacement: Microsoft
warns ordinary job notifications are not guaranteed, and message PIDs can already
have exited/recycled without a maintained handle. See
[job completion-port association](https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_associate_completion_port).
This is a rejected shortcut, not an implemented notification design.

## Evidence binding and next decision

Rehashed all25 files in matrix-validation-003/final-evidence-manifest.json without
drift. Failed receipt SHA256:
`244646AAE8F4D7C27755152AB207150C389BBD22070A91ACEE060D7EBBD0B844`.
Short-child receipt SHA256:
`786A64E1B9C3412DBF1DBF8696BC20D27F2FA9D9EC8980C7B1A22A10CBF581A4`.
Original matrix002 receipt remains
`5D6542FF46D808E32E5C89CBD360C593DFE517C0977DCD42152644BDD25AA2BA`.
Collector source remains
`FC8A9914BD41F539A18864426DAF3E5B5C3DE4CC6DB7B6E68D691E96CE3B364D`.

Independent read-only review found no Critical/Important/Minor issue and confirmed
sample coverage, source trace, handle-lifetime conclusion, clocks, cited Microsoft
contracts and stated hashes. Checkpoint-ready only. Reviewer did not verify Git
base/branch/diff, rerun tests, query processes/PIDs or audit the SDK; neither review
nor saved receipts establish an OS cause, a code defect or absence of all
unrecorded external activity. Controller verified Git scope separately.

Next recommended scope is a bounded design review of lifecycle/handle evidence:
what can be retained before root exit, how ownership and identity stay verified,
and how missing/recycled/duplicate/truncated observations fail closed. Compare
options and specify process-free tests before any implementation or live run.
Preserve existing rights, budgets and completion gates. A notification-only or
error87-as-success fallback is not sufficient. No such design is approved or
implemented by this analysis; no new experiment is authorized implicitly.

Recover from the existing development branch, read current-status.md and
docs/runbooks/repository-recovery.md. This report is portable; raw receipts,
SDK/cache/binaries remain ignored and are not restored by cloning. Main and
broader C11/C12/Task10/B01/production acceptance remain blocked.
