# Observation identity and resource scope

The helpers in ../tests/observation-session-guard.psm1 are preparation for a future
bounded observer invocation. Importing them does not launch, open or terminate any
process. They do not replace approval, preflight, artifact hashes or the observer's
absolute deadline. Do not turn synthetic test success into native acceptance.

## Identity evidence

Assert-TalenroObservationIdentity requires positive integer ProcessId/CimProcessId,
nonempty ExpectedPath/CimPath and ExpectedCommandLine/CimCommandLine, explicit UTC
DateTime ProcessStartUtc/CimCreationUtc, and Int64 NativeCreationFileTime.
Path comparison is ordinal case-insensitive; the entire command line is ordinal
exact, not a substring/profile match. No argument parser or path resolver is used.
The expected path/command line must come from the fixed approved launch, not be
copied from observed CIM values merely to pass this check.

The native creation FILETIME must come from a successful Times observation on the
ORIGINAL retained SafeProcessHandle. It must exactly equal ProcessStartUtc's
FILETIME. Combine native high/low words with integer arithmetic, not floating-point
conversion. Unknown/error/native mismatch must stop before accepting CIM evidence.
The pure helper cannot authenticate caller-supplied records or protect against a
caller deliberately supplying evidence from the wrong handle.

CIM creation time must lie on a microsecond boundary and equal ProcessStartUtc
with only its sub-microsecond remainder removed. No ±time tolerance and no rounding
across a microsecond boundary. Native evidence still distinguishes a one-tick
difference within that microsecond. WMI's format has six fractional microsecond
digits ([Microsoft documentation](https://learn.microsoft.com/en-us/windows/win32/wmisdk/cim-datetime));
the historical observed pair was .4427247Z versus .4427240Z. If a different provider
rounds or reports another precision, fail closed and investigate, do not widen it.

## Resource lifetime

Invoke-TalenroObservationScope takes trusted local scriptblock callbacks:

1. AcquireWriter returns exactly one owned disposable writer.
2. AcquireProcess returns exactly one owned Process, with no validation after its
   launch inside that factory. Move SafeHandle acquisition, CIM/native checks and
   sampling into Observe, which receives process then writer as its parameters.
3. All checks and the sampling loop execute inside the scope's try/finally.
4. The scope attempts writer disposal and process disposal even if a check, sample
   or the other disposal throws. Primary and cleanup errors are retained; success
   output is withheld on failure. Process.Dispose is not Kill/TerminateProcess.

Factories own any partially created resources until successful return; they must
clean those up if they themselves throw. Do not return multiple success-stream
objects or a resource shared by both factories. If wrapping a newly opened stream
in a writer can throw, that factory must dispose the stream before rethrowing.
Callbacks are trusted executable code, not a sandbox or a way to evaluate text
from receipts. Do not pass arbitrary Dispose implementations in live use.

The helper imposes no timer and cannot guarantee completion of a blocking Dispose;
callers must keep existing bounded observation policy. On abort, record target
identity and request normal UI close separately if safe. Never infer target exit
from observer-resource disposal. Do not reopen a PID to manufacture retained-handle
evidence. Preserve incomplete/empty raw receipts unchanged.

## Verification boundary

Run from the worktree with PowerShell7:

```powershell
./clients/windows/tests/observation-session-guard.tests.ps1
./clients/windows/tests/process-observation.tests.ps1
./clients/windows/tests/native-process-observation.tests.ps1
```

These are synthetic/in-memory checks only. New coverage uses actual MemoryStream
disposal and an explicit throwing-disposal double; no WinUI process or native
probe is run. Live wiring and the C5 route require separate bounded approval.
