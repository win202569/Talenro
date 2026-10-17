# Metadata runtime retained-handle probe001 — actual diagnostic result

Date:2026-10-17. Baseline:72022bfe94147feabbb89875a97d0d7650ff543c.
The user explicitly approved the one-shot probe after that checkpoint.
Status:DIAGNOSTIC EVIDENCE RECORDED;NOT ORDINARY/TDD/METADATA ACCEPTANCE.
The approved allowance is consumed:0 Go builds,1 runtime,no retry.

## Question answered and limits

This run captured one non-root own-job entry before root exit. Its admission
reported image path C:/Windows/System32/conhost.exe. The same retained handle
then yielded an exited state and exit code0 at root-wait end and after cleanup.
This answers the bounded question of whether earlier acquisition can capture
such an entry in this runtime context.

It does NOT identify the old unknown PID16616,prove that old Error87 was caused
by this executable,or repair the old phase. No cross-run/PID-only identity join
was made. The retained handle changes lifetime/timing and the polling cadence
differs from ordinary mode;this is sampled observation,not a full event trace,
an A/B experiment or a causal explanation. No automatic follow-up run.

## Frozen scope and preparation

W=D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator.
Branch:codex/c12-b01-task9-coordinator;origin https://github.com/win202569/Talenro.git.
New run root:.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/metadata-runtime-retained-probe-001.
Control:the same path with -control suffix. Both were absent before preparation
and are now consumed. They remain ignored;do not delete/reuse them to retry.

The fixed target was the unchanged metadata001/fixture-red/contracts.test.exe,
SHA256 18A77B83455F4429B254627401D16D20D4004549F989728FA93E632DBE5FC93C,
length20043264,matching its original successful default-vet build binding.
No recompilation,copy/replacement or new dependency acquisition.
The disposable probe.ps1 had current-host/PS5.1 parsing and a static sole-boundary
check before invocation;it is operational scratch,not a new product controller.

A fresh64-bit Windows PowerShell5.1 host invoked only the existing
RunWithRetainedHandleDiagnostics entry once. Working directory was
W/internal/nodecontrol/contracts,arguments exactly:

```text
-test.count=1 -test.timeout=120s -test.v=true -test.run=^TestAuthorityV7MetadataFixtureInventory$
```

Native runtime limit150s;existing32-admission cap,10ms-class polling slices,
500ms natural window and shared5000ms cleanup rule unchanged.
Twenty frozen environment overrides retained offline/local settings and used
new probe-owned tmp/cache paths;the unchanged native helper also inherits other
host environment entries. No blanket clean-environment claim is made.
The original four input pins/71-helper aggregate,full11536-file SDK/reference,
product manifest,original metadata001 records/binary/fixtures and wrapper bytes
were checked before/after as applicable. All eligible final audits ran independently.

## Observed result

| Fact | Saved result |
| --- | --- |
| Wrapper interval |10:10:59.4916869–10:12:47.4477199 UTC;107.956033s including audits. |
| Native interval |10:11:41.3401733–10:11:41.4297708 UTC;0.0895975s. |
| Test output | One top-level inventory RUN/FAIL,exit1 for the expected missing schema-registry.v1.json;no PASS/subtest/SKIP. |
| Mode | LifetimeDiagnosticOnly=true,ShutdownDiagnosticOnly=true,LiveSampling=false,ShutdownSchema=3,Coverage=sampled. |
| Core cleanup facts | RootExited=true,NaturalShutdownVerified=true,TerminationRequested=false,CleanupVerified=true,TimedOut=false;root-wait/final member lists empty. |
| Timing | NaturalObservationMs=21,ShutdownElapsedMs=26. |
| Job/sample counts | TotalProcesses=2,QueryCount=2,Attempts=Admitted=1;cap not reached,Incomplete=false,Errors=[]. |
| Teardown | TeardownAttempted=true,ReleaseAttempts=1,ReleaseFailures=0. |
| Finalization | Finalized=true,Status=diagnostic-recorded-not-acceptance,Failure=null,AuditErrors=[],Acceptance=false. |

Observed entry is EntryId1/PID40500 from query2. Open succeeded,error0;
IsProcessInJob succeeded,Owned=true. Admission wait state was running,exit-code
read available (259),image read succeeded with the path above.
RootWaitEnd and Final use SameEntry=true,Ambiguous=false,Error0,
ExitCodeRead=true,State=exited,ExitCode0,TerminationRequested=false.
Image reads after exit returned31/null;the earlier image observation and later
same-handle wait/exit facts are separate. Image-read failure is not promoted to
positive ownership/exit evidence and does not erase those previous observations.

The monotonic source-list-end→open-start interval was3.2282ms. That does not
measure the old member's lifetime or explain its0.2196ms acquisition gap.
No historical PID lookup was performed.

The probe question has positive evidence for this sampled entry only.
Both diagnostic mode flags remain set;the ordinary acceptance guard is unchanged.
This test output is not an accepted metadata RED,even though cleanup was natural.
The old failed phase and its refusal remain byte-identical.

## Saved-only verification and portable fingerprints

After the worker returned,a separate readonly check rehashed all entries in
native-binding.json and final-binding.json,compared before/after product and
extra-file snapshots,and checked diagnostic flags,scope,admission/membership,
three same-entry reads,cap/error/teardown counters and cleanup facts.
Those saved-only checks succeeded;no additional executable was launched.

| Artifact | SHA256 |
| --- | --- |
| receipt.json |4BC52875C2137E89D66861722019B9DAE6C6D439CF9BAC41D91F02C6893332A0 |
| result.json | A49650AD951FE81CC0EF461D6B17A061CEA62E878CDAB7D7F346BE01030A655E |
| stdout.txt | C8DAEBE01AAF4681EA3ACE98CF0230BEFBA935BFC453C45A00E535FC7324C36B |
| native-binding.json |3A198983CF6C793F32DD159BACE245B89522EAA4E0EB601411C29DDCD21C247D |
| final-binding.json |79F33488415283379A6ADE89EDC2A0CE46A7E2BAB6C63C17CB01028956B7EAC0 |
| before-product.json and after-product.json | E6652D141D90A1BE480485148C50063F4036C1AAB8621448CA02EC355EB1DEC2 |
| before-extra.json and after-extra.json | C99F80CA3CEE097EAB5BE3CDA11706DAF69D1E2552D88ED5246F64323CBCECAC |
| control/probe.ps1 |56E0501A1168F0BB862616A1EBA657B35142F44484867E5930BA2A83D01B811D |

Raw receipts/control/SDK/cache/binary stay ignored,not committed.
Original metadata001 result/receipt/control hashes were also preserved by the
extra-file snapshots. Git preserves this portable result,not the binary or logs.

## Handoff and safe next step

No product/test/diagnostic-helper/SDK/security-policy change was made.
Metadata Task1/2 remain stopped:product registry and both metadata fixtures absent.
Metadata plan usage stays1 build/1 runtime;this independent diagnostic adds
0 builds/1 runtime. Neither identity may replay;no remaining diagnostic run.

Next recommendation is a readonly comparison of the saved ordinary and sampled
timelines before deciding whether any ordinary observer redesign is justified.
Do not promote retained diagnostic evidence into an acceptance path,whitelist
conhost.exe,ignore87,increase rights/waits or rewrite the old receipt.
Any change to ordinary identity acquisition or its acceptance policy requires a
separate explicit design/approval and covering tests;none is authorized here.
Source/codecs/storage/canonical registration/DB/PITR/main/production gates stay open.
