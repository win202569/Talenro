# Fixed C4 English/RTL observation entry

Implementation and synthetic checks do not authorize a live run. Current live
allowance is zero. C4 UI remains UNVERIFIED and existing acceptance gates remain.
Use PowerShell7 in the existing development checkout. Default invocation:

```powershell
pwsh -NoProfile -File clients/windows/tests/observe-c4.ps1
```

Returns one NotStarted object before imports, environment queries or writes.
Only after separate bounded approval and operator readiness, the same command with
`-RunApproved` enables the fixed entry. This switch is not proof of authorization.
There are no path, profile, language, duration or callback overrides.

## Preparation and fixed identity

Ordinary non-administrator token; clean `codex/c12-b01-task9-coordinator`; committed
entry and consumed shared source blobs must match working files. The conservative
demo/prototype process inventory must be empty; inaccessible inventory blocks.
Operator must independently ensure no existing target windows and usable supported
UI observation. Never change security settings to pass a gate.

Artifact relative path:
`clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe`.
EXE SHA256: `4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95`.
DLL SHA256: `7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB`.
Arguments: `--layout-test --layout-language=en-US --layout-profile=rtl-mixed`.
No close-run argument is supported by this preserved artifact.

Cloning restores source/tests/docs, not ignored binaries or receipts. A missing or
changed artifact blocks; do not rebuild, install or substitute automatically.
Restore the separately retained approved artifact and verify both hashes before
requesting a live run. Check the latest status and execution ledger, not old budgets.

## Receipt meaning and failures

Below the executable directory, `process-observations/<lowercase-runId>/` contains
create-only `manifest.json` then `samples.jsonl`. Existing destinations and reparse
ancestors reject. The manifest contains only schema1, C4 label, runId, en-US,
rtl-mixed, relative executable path, exact arguments, both hashes,360000ms cap and
2000ms maximum sleep. Its runId correlates all samples and terminal progress.
It has no private environment inventory or arbitrary exception payload.

Manifest preparation is not proof of launch or UI success. Manifest write, flush,
disposal or samples-writer construction failure prevents process acquisition.
Partially created directories/files remain evidence; never overwrite/delete/retry.
The factory cleans partial acquisitions, then transfers the samples writer to the
unchanged shared resource scope. That scope disposes writer and original Process;
the native adapter borrows its original SafeHandle. Never reopen by PID or kill.

Shared orchestration retains its historical `Invoke-TalenroC5Observation` name and
C5-prefixed error messages. Those indicate shared observer failures, not a C5 run.
The C4 manifest and independently fixed full command line establish the profile.
Native creation identity is exact; CIM time is microsecond-truncated with explicit
UTC. Native evidence precedes managed observations. Failure evidence is retained.
Raw outcomes are provisional (`Finalized=false`). Only terminal post-cleanup output
can finalize; a cleanup failure or overrun cannot become timely success. Nonverified
exit or failure is unsuccessful. UI acceptance is separate even after verified exit0.

Synchronous OS/write/cleanup calls can exceed the cap. This is neither a hard
real-time deadline nor protection against hostile filesystem races. Trusted local
workspace only. Errors do not trigger a retry or force-close.

## Future bounded UI route — requires separate approval

Proposed one launch, zero retries,360s cap, at most2s between samples. Startup30s;
request normal titlebar close by300s; assess normal exit within30s. Check remaining
budget before every action; unfinished checks are NOT_RUN, never extend the run.

1. Login: ordinary English labels and permanent offline notice; visibly mirrored
   root and synthetic rtl-mixed heading. Inspect actual read-only email separately.
2. Enter demo; inspect Connection, Nodes and Account. Technical identifiers and
   actual email should remain readable LTR. No search typing, connection request
   or node selection. Inaccessible identifiers remain UNVERIFIED.
3. Only if time remains, logout/re-enter and inspect retained profile and visible
   ordinary reset state. This is not exhaustive state-reset proof.
4. Normal titlebar close; retained-handle signal/code/time plus managed agreement.

Report separate verdicts for mirroring, actual-email LTR, technical-ID LTR, exact
fixture codepoints, clipping and profile retention. Screenshots cannot prove
invisible bidi controls; text extraction that strips them also cannot. Exact
codepoint claims remain UNVERIFIED without preserving those characters.
No natural-language shaping, screen-reader, contrast, target viewport, full keyboard,
fallback-language, C6, overall C5 or historical shutdown-root-cause acceptance.

## Synthetic verification

`pwsh -NoProfile -File clients/windows/tests/observe-c4.tests.ps1` extracts private
definitions and replaces OS boundaries with recording/in-memory doubles. It never
invokes the live script body, compiles native code, calls real CIM, sleeps or launches.
Run the unchanged five observation suites separately as documented in the execution
ledger. Passing synthetic checks does not establish Windows live or RTL behavior.
