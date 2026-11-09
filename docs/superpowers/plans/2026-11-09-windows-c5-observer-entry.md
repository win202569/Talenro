# Fixed C5 observer entry implementation plan

> **For agentic workers:** Use superpowers:executing-plans for sequential implementation in this session, with TDD and one final independent code review. Do not launch the application.

**Goal:** Wire the existing identity/resource guards into one portable, fixed C5 observer entry with synthetic regression coverage.

**Architecture:** A no-start-by-default PowerShell script owns fixed live adapters; a small orchestration module consumes a complete trusted backend contract. Tests use that same orchestration with in-memory evidence and fake time, never live fallback.

**Tech Stack:** PowerShell7, existing process/native observation and session-guard modules; no new package or product build.

**Spec:** docs/superpowers/specs/2026-11-09-windows-c5-observer-entry-design.md (approved after f26d509e).

Status: awaiting user confirmation. Native launch budget0. No implementation yet.

## Global constraints

- Fixed preserved executable/profile and EXE/DLL hashes exactly as the spec; no CLI path, argument, timeout or callback overrides.
- Default invocation returns a no-start status before CIM/native compilation/file creation. Only -RunApproved selects live preparation; this switch never substitutes for actual user approval.
- Live scope: ordinary token, clean codex/c12-b01-task9-coordinator, matching observer source blobs/hashes, no existing demo/prototype targets. UI/window preflight and close remain operator responsibilities.
- Fixed360000ms monotonic observation cap,2000ms maximum sleep, zero retries or forced exits. Start clock immediately before the one launch attempt, not after CIM.
- Reuse original Process/SafeHandle. Native creation must match exactly, CIM at microsecond truncation; full fixed command line must match. No PID reopening or millisecond tolerance.
- Operator retains startup30s/close-by300s/exit-within30s criteria. Synchronous calls/disposal can overrun; record actual elapsed and never claim a hard real-time guarantee.
- Implementation verification: synthetic PowerShell only; no WinUI/native-probe launch, .NET build/test, restore, installs, settings changes, main merge or historical acceptance rewrite.

## Review focus

1. Default or incomplete synthetic backend accidentally reaches live code — test no side effects and no fallback.
2. Native/CIM identity failure occurs before a useful receipt — retain unverified first native observation, then stop and dispose.
3. Exception/cleanup produces partial success or loses target identity — immediate sanitized progress, failure remains failure.
4. Clock regression/API delay crosses budget — reject invalid clock and record timeout/overrun without another sample or sleep.
5. Signaled handle conflicts with managed evidence — retain both; never turn code0 alone into verified exit.

## Task1 — Shared orchestration with synthetic integration tests

**Create:** clients/windows/tests/c5-observation-session.psm1 and c5-observation-session.tests.ps1.
**Reuse unchanged:** observation-session-guard.psm1, process-observation.psm1, native-process-observation.psm1/.cs.

**Interface:** Invoke-TalenroC5Observation -Mode Synthetic|Live -Context <dictionary> -Backend <dictionary>.
Context contains RunId (nonzero lowercase Guid N), ExpectedPath, ExpectedCommandLine.
These are trusted internal inputs, not live CLI overrides. Mode is stamped on every
record; Synthetic cannot be reported as live. Require all backend keys to be named
scriptblocks, reject missing/extra keys before invoking any callback:

| Callback | Input -> result |
| --- | --- |
| Preflight | context -> no result, throw on failure |
| AcquireWriter | context -> one owned object with WriteLine/Dispose |
| AcquireProcess | context -> one original owned disposable process; start monotonic clock immediately before acquisition |
| GetIdentity | process -> ProcessId, StartUtc, Handle (original retained handle) |
| ReadNative | handle, runId -> existing native observation schema |
| ReadCim | processId -> ProcessId, ExecutablePath, CommandLine, CreationUtc |
| ReadManaged | process -> existing managed observation schema |
| ElapsedMs | no input -> finite nonnegative monotonic milliseconds |
| SleepMs | integer milliseconds -> no result |
| Publish | sanitized progress record -> no result |

The module imports the guard by its own directory. It never supplies a default
backend, starts a process, queries CIM or loads the native helper itself.

- [ ] Write behavioral tests before implementation. Use MemoryStream-backed writer fixtures, original-object identity assertions, recorded callback order and fake time. Assert literal expectations for the historical .4427247/.4427240 pair; wrong PID/path/full args/native tick/CIM microsecond and invalid native API status all reject before regular paired sampling.
- [ ] Cover writer/process acquisition failure, GetIdentity failure, native/CIM failure, sample/write/progress failure and disposal failure. Assert both acquired resources get disposal attempts, no retry, and no success output after errors. Use the existing guard for cleanup rather than duplicating it.
- [ ] Cover native-before-CIM initial evidence and native-before-managed pairs, original handle/object reuse, fresh sequence/run correlation, and receipt retention on later identity failure. No regular pair may be marked identity-verified before the guard returns.
- [ ] Cover exact cap and near-cap timing: at359500ms maximum sleep500ms; at360000ms no new sample/sleep; negative, NaN, infinity or decreasing elapsed values fail. A callback that returns after the cap retains evidence and overrun but cannot pass timely completion.
- [ ] Cover valid signaled/code0/real exit time with managed agreement, nonzero exit, missing exit time and managed disagreement. Signal ends observation without retry, but only fully valid agreeing evidence within budget sets ExitVerified=true.
- [ ] Run `./clients/windows/tests/c5-observation-session.tests.ps1`; record expected missing-behavior RED, not a live failure. Implement the module minimally and rerun for GREEN; stop to diagnose unexpected failures, never invoke live adapters as a workaround.

**Records:** JSONL schema1 with Mode, RunId, Kind, Sequence and elapsed time.
Kinds: start, native-identity (explicitly unverified), pair, outcome. Preserve raw
native/managed receipts in their respective records; integer FILETIME arithmetic.
Outcome values: identity-failed, observation-failed, timed-out, exited; include
ExitVerified, reason code and overrun where applicable. No exception messages/stacks.
Keep start/abort/completion progress on the immediate information stream via Publish,
not the success stream buffered by Invoke-TalenroObservationScope. Writer errors
must propagate even when a failure outcome cannot be saved. Final return occurs
only after successful resource cleanup; it never means C5 UI acceptance.

- [ ] Run the existing48-case session-guard,32-case managed and78-assertion native synthetic/static suites in separate hosts. Review explicit diff and checkpoint Task1 with result counts; no independent review until Task2 is complete.

## Task2 — Fixed live entry, no-start tests and recovery

**Create:** clients/windows/tests/observe-c5.ps1, observe-c5.tests.ps1,
clients/windows/diagnostics/c5-observation.md and docs/roadmap/2026-11-09-windows-c5-observer-entry-execution.md.
**Update:** docs/roadmap/current-status.md, docs/runbooks/repository-recovery.md.
**Consume:** Task1 callback/receipt contracts without changing their names or semantics.

- [ ] Write no-start tests first: invoke the entry without switches; expect exactly a NotStarted status, no launch/native helper/CIM/file-write calls. Unknown override parameters must reject. Use recording command doubles for this no-start path, not real process operations.
- [ ] Implement the default guard before all live setup. Under -RunApproved only, resolve worktree from script location, use fixed spec path/args/hashes, validate ordinary token and clean expected Git branch/sources, reject existing demo/prototype processes. Missing artifact/cache/tool or inaccessible evidence blocks; no install/rebuild/retry.
- [ ] Generate a fresh runId and create a new receipt directory/file under ignored process-observations. Reject existing destination/reparse ancestors and use CreateNew; trusted workspace only, not an adversarial race guarantee. Clean partially acquired stream if writer construction fails.
- [ ] Build all ten backend callbacks explicitly. AcquireProcess performs only the single launch/transfer; GetIdentity obtains its original SafeHandle inside protected Observe. ReadNative and ReadManaged call existing adapters. ReadCim converts time to explicit UTC. Expected full command line derives from the fixed launch contract, never observed CIM text.
- [ ] Native API success/type validation precedes integer creation FILETIME conversion and guard invocation. Handle retention covers validation and all pairs. Dispose writer/process through the existing scope; never independently dispose a borrowed handle or terminate the target.
- [ ] Make tests exercise preflight rejection and adapter wiring through private in-memory command doubles without exposing callback CLI parameters. Fail tests on attempts to invoke actual Start-Process/Get-CimInstance/Add-Type/sleep. Assert missing evidence, dirty/wrong branch, hash/source drift and existing target all prevent acquisition.
- [ ] Run `./clients/windows/tests/observe-c5.tests.ps1` RED then GREEN; run the Task1 and existing three synthetic suites. Verify PowerShell parsing, no product/binary changes and `git diff --check`. No call with -RunApproved against real adapters in this implementation scope.
- [ ] Document default versus future authorized invocation, operator UI/deadline duties, receipt interpretation, no-start/failure statuses and recovery from missing preserved artifacts. The historical empty C5 receipt and its verdict stay unchanged.
- [ ] Obtain one fresh read-only review of these new files/contracts, especially the five Review Focus items. Resolve Critical/Important findings with regression tests; record limitations and remaining minor issues. Do not substitute live runs for missing synthetic proof.
- [ ] Save actual commands/counts, review dispositions and remaining unverified live wiring in the execution ledger. Stage explicit necessary files, commit/push existing branch, verify remote SHA and clean worktree. No main merge or new application allowance.

## Self-review and handoff

Task1's exact callbacks/receipt contract feed Task2; both reuse the existing guard.
The five risk classes map to Task1 ordering/lifetime/time/outcome tests and Task2
default/preflight tests. Source/runtime separation prevents synthetic fallback.
No unsupported provider rounding, generic executable runner, UI automation or
application fix is added. Real CIM conversion, process races, OS blocking calls,
C5 display behavior and shutdown root cause remain separate native evidence gaps.

Please confirm this plan for sequential implementation; approval covers code and
synthetic verification only. No additional per-task approvals are needed within
that scope, but a real application launch remains a separate decision.
