# Independent native process observation — short design

Status: approach 1 selected by user; written design approved after3bf6a2d5.
This is a small diagnostic component, not an application shutdown fix.

## Intent and approaches

Separate OS-returned wait/code/time data from managed Process properties for the
unresolved false-wait/true-HasExited/exit0/1601 tuple. Prior evidence and all
Task11/12 and44/34/6/217 NOT_ACCEPTED gates remain unchanged.

Selected: independent native recorder beside the unchanged managed recorder.
It preserves the existing module's no-native-call contract and permits paired
observations. Alternative: extend that module; fewer files, but changes its
verified boundary and mixes contracts. User selected independence, not a live run.

## Proposed implementation choices

Use existing PowerShell7 on Windows with a small, repository-owned C# interop
helper, loaded explicitly by the live entrypoint only. No packages or installs.
Suggested files under clients/windows/tests:
- native-process-observation.psm1: one-shot live and separate synthetic entrypoints.
- native-process-observation.cs: typed interop and immediate error capture.
- native-process-observation.tests.ps1: dependency-free synthetic/static tests.

These filenames and internal structure are builder proposals. Public schema and
test cases will be pinned in the implementation plan after design approval.
Import and synthetic tests must not load the helper or call Windows APIs.

## Handle ownership and safety

The live entrypoint accepts a caller-retained SafeProcessHandle obtained from the
original launched Process while it was alive, not a PID or bare numeric handle.
The caller retains the same Process and handle for later paired observations;
it must not reacquire, refresh, replace or dispose them during collection.
The helper holds a balanced SafeHandle reference across the complete native
snapshot, releasing only its own reference in finally. It never closes or owns
the caller's handle. Reject closed/invalid handles; no fallback OpenProcess,
handle duplication, enumeration or privilege adjustment. Caller-provided identity
and correlation labels are metadata, not independently verified ownership proof.
Compare returned creation time with the established launch identity at integration.

Only WaitForSingleObject(handle,0), GetExitCodeProcess and GetProcessTimes are
allowed, once each in that order. Required handle rights are SYNCHRONIZE and
PROCESS_QUERY_LIMITED_INFORMATION (or QUERY_INFORMATION). Failure is evidence,
not permission to reopen with more rights. No launch/kill/attach/dump, polling,
sleep, retries, infinite waits, file writes or settings changes inside the module.

## Receipt contract

Include schema version, native-live versus synthetic mode, correlation metadata,
snapshot and per-call start/end UTC timestamps. Reads are sequential, not atomic.
Preserve each call independently; exceptions do not manufacture default outputs.

- Wait: preserve UInt32 raw result, decoded signaled/timeout/failed/unexpected
  label, and last-error only for WAIT_FAILED. Timeout is not API failure.
- Exit code: preserve BOOL success separately from UInt32 output. On API failure,
  output is null and last-error is captured immediately in the interop helper.
- Times: preserve BOOL success; on success retain creation/exit/kernel/user raw
  FILETIME high/low UInt32 pairs, avoiding JSON precision loss. On failure all
  time outputs are null and last-error is captured immediately.
- Successful calls have null last-error even if an old thread error exists.
  Interop/validation exceptions carry type only, no messages/stacks/private paths.
  A handle-acquisition failure yields an explicit snapshot error with calls marked
  not-run, never three fabricated native failures. Type-invalid synthetic receipts
  remain distinct from native API failure.

Do not interpret raw zero exit time as actual1601 shutdown, or successful time
retrieval as proof of exit: exit time is undefined before exit. Do not convert
code0, code259 or a signaled wait alone into application acceptance. Keep raw
unsigned exit codes; comparison with signed managed values must preserve bits.
The module produces no PASS verdict and never alters the old managed receipt.

## Verification and limits

Test-first synthetic cases: signaled/timeout/failed/unexpected wait; exit0/nonzero/
259/high-bit codes; failed APIs with null outputs; stale errors suppressed on
success; zero and maximum FILETIME halves; thrown calls with later calls retained;
wrong types/null/multiple outputs rejected; JSON round-trip; mode/correlation and
sequential timing. Static checks cover API allowlist, zero wait, balanced handle
reference, immediate error capture and unchanged existing recorder. Synthetic
tests cannot prove ABI correctness, handle lifetime at runtime or native cause.

Implementation plan should include helper compilation/signature verification
without executing native calls; it is not an application build or real probe.
Any actual native invocation, including against the current process or a child
fixture, requires the later explicit runtime scope. No app launch budget is
granted by this design. Later paired live collection must preserve the existing
30s shutdown criterion/360s overall ceiling and stop on invalid ownership or
identity mismatch; it must not automatically relaunch or kill on timeout.

## Handoff and primary references

After design approval, write the implementation plan. Save portable design/test/
progress records on the existing development branch; exclude raw private receipts.
Root cause and acceptance remain unresolved even if synthetic tests pass.

- https://learn.microsoft.com/en-us/windows/win32/api/synchapi/nf-synchapi-waitforsingleobject
- https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-getexitcodeprocess
- https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-getprocesstimes
