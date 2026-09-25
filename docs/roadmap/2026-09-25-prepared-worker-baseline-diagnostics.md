# Prepared worker baseline failure — 2026-09-25

Historical investigation record: the HEAD and publication statements below refer
to the investigation time. The user subsequently requested publication of this
sanitized report to the existing project repository for recovery. This is not a
Microsoft/vendor submission or security clearance. Vendor submission was deferred
because the user has no Microsoft account; the blocker remains open. See the
[current recovery handoff](2026-09-25-progress-recovery-handoff.md).

Status: investigation evidence, not a fix or acceptance result. Branch
`codex/c12-b01-task9-coordinator`, HEAD `bea8b5348d1f6fb2e2fb80b331eeeda1d6e6780f`.
The three uncommitted development-tool files remain incomplete and unintegrated.

## Observed failure

The offline Windows ordinary suite (`go test ./... -count=1 -timeout=15m`)
returned exit 1. `internal/testinfra` took 683.825 seconds and reported:

`TestC12PreparedProcessIsSuspendedUntilVerifiedJobMembership/prepared-exit-pending`

The error was `CreateProcessW failed (Win32 5)` from
`C12SuspendedProcessController.Create`. The other reported ordinary packages passed.
The original output remains in the ignored plan workspace as `ordinary-baseline.log`.

Read-only source tracing places the failed boundary at creation of the Windows
PowerShell bootstrap process. This precedes Job assignment, membership validation,
and the test's deliberate restriction of the process-termination handle. Therefore
the observed creation failure is not the intentional termination-denial assertion
that this test is supposed to exercise.

## Minimal reproduction

The unchanged exact subtest was run once, offline, with the same fixed Go 1.26.5
toolchain and reviewed elevation required for toolchain access:

```text
go test -v ./internal/testinfra -run '^TestC12PreparedProcessIsSuspendedUntilVerifiedJobMembership$/^prepared-exit-pending$' -count=1 -timeout=5m
```

It passed: subtest 6.50 seconds, parent 8.12 seconds, package 9.841 seconds.
No production change preceded this reproduction. This proves only that the failure
is not reproduced by this isolated run, not that the baseline or historical I3
failure is resolved. The original harness deadlines and cleanup were unchanged.

## Relevant host evidence

Read-only event queries covered 14:10:29–14:26:29, local UTC+04:00, September 25.
Defender Operational recorded the following linked events:

| Local time | Event ID / record | Observation |
| --- | --- | --- |
| 14:23:44.398 | 1116 / 17273 | Detection: `Trojan:Win32/Commando.A!ml`, command-line resource pointing to Windows PowerShell with `-NoProfile -NonInteractive -ExecutionPolicy Bypass -EncodedCommand` |
| 14:25:01.889 | 1117 / 17280 | Same detection ID, action reported as deletion, error `0x00000000` |

Both events share detection ID `{C4A251B5-F254-4E7F-8B2C-0D58DB840E56}`.
Security intelligence: 1.459.388.0; engine: 1.1.26080.3.
The stored command-line resource is only 1023 characters long. Decoding its prefix
without executing it yields the MemoryStream / FromBase64String loader prefix used
by `New-C12PreparedNativeWorker` in the unchanged runner.
Code Integrity and AppLocker EXE/DLL queries returned no matching-time events;
this absence is not proof that all other controls were inactive.

No complete encoded command, bootstrap key, or raw event dump is included here.
The action label applies to the recorded command-line resource; it does not establish
that the system PowerShell executable was deleted.

## Interpretation and limits

There is strong evidence that endpoint protection rejected a C12-style bootstrap
command during the failed suite. This is a substantially better-supported hypothesis
for the creation denial than the test's later intentional handle restriction.
However, the suite log has no per-subtest timestamp/PID correlation, and the event
command is truncated. These records do not conclusively attribute that detection to
this exact subtest or prove that it was a false positive. They also do not explain
all historical I3 failures or establish that tool-module isolation would fix it.

## Disposition

- No runner, production, module, security configuration, deadline, or cleanup change.
- No exclusion, allow-list entry, protection disablement, quarantine restoration,
  encoded-command rewrite, or automatic retry used to work around the detection.
- No upload to a vendor or external service; no commit, push, or merge.
- Tool isolation Task 1 remains incomplete; Task 2/3 remain unstarted. I2/I3 remain open.

Next decision: review the bootstrap source and this detection with the security
owner/vendor through the normal false-positive assessment process before choosing
any remediation. A new test pass alone cannot close this blocker. Any modification
to the runner would exceed the approved tool-isolation scope and needs a separately
approved design and regression evidence.

## Follow-up: scoped static security assessment

The user approved continuing the assessment. This pass read source only; it did
not execute, reconstruct for execution, or submit the detected bootstrap.

Provenance checks:

- Working-tree runner Git blob and HEAD blob both equal
  `23a6752b05a18cb89546c78c1f091f0e769d75ea`; no uncommitted runner change.
- Runner SHA-256:
  `047B07EEC9615FC43AA112E867AF861493309D4C91083A7A2755F0665D42B359`.
- The local system PowerShell executable's Authenticode status is Valid, signed
  by Microsoft Windows. This verifies the executable's signature, not the safety
  of commands supplied to it. Git equality similarly is not a safety verdict.

Inspected boundaries: `New-C12PreparedNativeWorker`, the complete embedded worker
script, prepared receipt validation, protocol serialization/authentication helpers,
and the local named-pipe implementation (runner approximately lines 3150–3453 and
3791–3857; native process creation was traced in the preceding investigation).

Observed intended behavior:

1. The compressed command source is assembled from functions and embedded code in
   the checked-in runner, not fetched from a network location in this path.
2. The worker connects to the local machine's named pipe, checks the controller
   PID, and uses framed messages bound to a run, nonce, sequence and HMAC.
   The server explicitly rejects remote clients.
3. Before invoking the selected executable, the worker verifies the supplied file
   hash and identity/permissions projection, checks the role/capability allowlist,
   and waits for the controller's release gate. Results return over that pipe.
4. The inspected bootstrap path contains no observed network download-and-execute,
   persistence installation, or external data-exfiltration operation. This is a
   limited static observation, not an exhaustive runner or whole-system audit;
   the downstream executable and every native sealing implementation were not
   independently audited in this follow-up.

The compression + Base64 + EncodedCommand + dynamic ScriptBlock construction is
visible in source and consistent with the truncated detected command. Its presence
does not prove maliciousness; equally, a legitimate intended purpose does not
prove the Defender classification is wrong. The precise detection rule is unknown.
No rewrite to change detection characteristics was attempted.

Assessment: a suspected false-positive review is reasonable, but the alert remains
unresolved. No clearance, blanket safety certification, or acceptance claim is made.

## Proposed vendor review, not yet authorized or submitted

Microsoft's [submission guidance](https://learn.microsoft.com/en-us/defender-xdr/submission-guide)
directs software developers disputing a detection to submit for analysis and await
a final determination. The [Security Intelligence submission portal](https://www.microsoft.com/en-us/wdsi/filesubmission)
provides a Software developer category and requires sign-in for submission.

Proposed disclosure, requiring explicit user approval before upload:

- The exact checked-in `scripts/run-c12-integration.ps1` as source context, clearly
  identified as NOT an exact captured instance of the truncated command-line alert.
- This sanitized diagnostic document, including product/signature versions, event
  IDs, timestamp offset, source hash, failed test and isolated-pass evidence.

Do not send the whole repository, environment files, credentials, bootstrap keys,
ETL recordings, raw event dumps, or quarantine contents. No exact original runtime
sample is currently established; do not silently substitute the source file or a
new execution as an exact sample. If the portal requires unavailable runtime
evidence, request vendor guidance instead of regenerating the detected command.
The owner must review the disclosure scope and use an authorized account.
