# Lightweight Windows process observation records

Status: proposed design, pending document approval. User selected the lightweight
approach after8ce36286. This document does not authorize a demo launch or debugger.

## Intent and approaches

Collect explicit same-object observations so wait failure, unavailable properties
and genuine exit0 cannot be confused. Existing C3 evidence and all NOT_ACCEPTED
gates remain unchanged; this is not a fix for the application or runtime.

Selected: a small PowerShell recording module plus synthetic tests, separate from
the application. It supplies one-shot observations, not process lifetime control.
Alternative: replace the entire launcher and its timing/ownership logic. Deferred
because it changes more variables and is unnecessary to validate receipt semantics.
Caller integration and real-process verification are later, separately scoped work.

## Scope and files

Proposed builder choices, not pre-existing user requirements:
- `clients/windows/tests/process-observation.psm1`: reusable observation functions.
- `clients/windows/tests/process-observation.tests.ps1`: self-contained synthetic
  assertions using existing PowerShell7, no package installs or Pester dependency.
- Execution/current-status/recovery docs: record actual test evidence and limits.

Import must have no process or filesystem side effects. The module never starts,
enumerates, terminates, disposes or attaches to a process; never prompts for higher
privilege, calls native interop, changes settings or writes files. No polling loop,
sleep, infinite wait or automatic retry. Caller retains ownership of the Process.

## Observation contract

Provide a live adapter accepting one existing `System.Diagnostics.Process` object,
with no PID-based reopening. Read identity from that object and capture each of
`WaitForExit(0)`, `get_HasExited()`, `get_ExitCode()` and `get_ExitTime()` once,
in that order, with a separate try/catch per field. Exceptions in one field must
not suppress later fields. Property reads must be explicit method calls, not
formatted property expansion that can hide getter failures.

Each field record has `Status` (`ok`, `error`, `invalid`), `Value` and
`ExceptionType`; include observation start/end UTC timestamps. Failed/invalid
reads have null Value. Error records carry the caught exception type only, not
exception messages, paths, stack traces, command lines or environment values.
Record PID and process start time with the same success/error convention. Times
serialize unambiguously in ISO8601 UTC; a caller must not infer ownership from a
record whose identity fields are unavailable. The module cannot retrospectively
prove that the caller supplied the originally launched object.

Validate successful values before labeling `ok`: Boolean for wait/HasExited,
Int32 for ExitCode, positive Int32 PID, DateTime for time getters. Null, strings,
arrays/multiple output values and implicit numeric coercion must not become zero
or false. Preserve actual ExitCode0, nonzero and259 as numbers, not truthy/falsy
conditions. Do not label any snapshot PASS or normal shutdown; return observations
without choosing a winner when wait and HasExited disagree.

Snapshots are sequential, not atomic. Timestamps expose the observation interval;
an app may transition during it. Do not claim all fields reflect one instant.

## Synthetic seam and verification

Use a separate clearly named synthetic entrypoint accepting fixed scriptblock
readers for these six fields, feeding the same field-validation/record assembly
logic. Label every snapshot `Mode=synthetic` or `Mode=live` so simulation cannot
be mistaken for native evidence. Reader code is trusted test input, not a security
boundary; no reader expressions from external files/arguments are evaluated.

Required synthetic cases:
1. Exit0 and nonzero retain type/value;259 is not automatically interpreted.
2. Wait false/HasExited false with unavailable exit fields stays distinct from0.
3. Wait false/HasExited true remains an unchanged disagreement, with no verdict.
4. Every field may throw; remaining readers are still called once in order.
5. Null, wrong type and multi-value output are invalid, not coerced.
6. Identity failures are preserved without manufacturing a PID/start timestamp.
7. JSON round-trip retains nulls, Boolean values, integer0 and mode labels.
8. Module import and synthetic tests do not execute live adapter/process actions.

Use test-first development with in-memory synthetic readers; no demonstration app,
child-process probe, native stack capture or external service calls. Synthetic
success proves only record logic, not Windows process semantics or root cause.
No build/restore/full application-suite run is included in this small component.

## Handoff

After design approval, prepare the implementation plan and test commands. Keep
the live adapter unexecuted until new runtime scope is approved. Commit necessary
module/tests/docs to the existing development branch once implemented and verified;
exclude raw runtime receipts. No main merge or widened acceptance.
