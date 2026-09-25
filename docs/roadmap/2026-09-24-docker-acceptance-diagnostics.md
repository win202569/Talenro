# Docker acceptance diagnostics — 2026-09-24

Status: acceptance remains blocked. This record is diagnostic evidence, not a
physical-gate PASS. No production code, runner policy, deadline, security
configuration, or dependency content was changed.

## Docker prerequisite

- Docker was found at
  `C:/Users/Lenovo/AppData/Local/Programs/DockerDesktop/resources/bin/docker.exe`.
- `docker version` returned both client and server: Docker Engine 29.6.2,
  Docker Desktop 4.85.0, Linux/amd64 server, context `desktop-linux`.
- Earlier checks of PATH and common machine-wide installation locations were
  insufficient: this is a user-level installation and the existing terminal's
  PATH did not include it.

## First physical-gate attempt

From the coordinator worktree, using unchanged source at `ac012935`:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/testinfra' -Run '^TestC12AuthorityPITRProfile$' -Timeout 5m
```

The Docker binary directory was added only to the child process PATH. An initial
launch was rejected by the inherited `GIT_*` environment guard. Removing those
environment entries from the child process with the PowerShell environment
provider allowed the unchanged runner to proceed. The guard itself was not
modified or bypassed.

Session 91210 exited 1 with:

```text
Go graph go mod verify timed out after its bounded native wait
```

Observed preparation ran approximately eight minutes. Its active Go command
was the pinned Go 1.26.5 executable running `mod verify`. No selected test PASS
was produced. The other 12 physical invocations were not launched. Observed
runner/Go process IDs had exited after completion; Git remained clean before
this documentation was added.

## Bounded diagnostic observations

- With the fixed module cache and network-disabled Go settings, `go mod graph`
  exited 0 in 0.07 seconds and produced 3,423 edges. A preceding `go list -m all`
  probe reported offline module-lookup errors; that broader metadata lookup is
  not evidence that the runner's module graph could not load.
- The entire local download cache contained 801 zip files totaling 986,365,380
  bytes. This includes multiple versions and the toolchain; it is **not** the
  exact set or byte count verified for this project.
- A standalone `go mod verify` diagnostic used the same pinned executable,
  module cache, offline settings, target platform and readonly integration flags.
  It did not reproduce the runner's sealed process or private artifact layout
  and is not an acceptance substitute.
- Diagnostic samples showed continuing reads: 113,328,334 bytes / 71,285 read
  operations at 43 seconds; 305,186,228 bytes / 168,324 operations at 125 seconds.
  CPU time at the last sample was 23.17 seconds. The diagnostic was stopped at
  the first sampling boundary after its 120-second limit, with no stdout or
  stderr and no verification result. It was not a completed integrity check.
- A separate raw SHA-256 timing read a 43,682,724-byte cached SQLite zip in
  0.157 seconds, and 200 pgx source files totaling 1,303,499 bytes in 1.689 seconds.
  These are small, potentially warm-cache samples, not equivalent Go verification
  workloads and not proof of a particular filesystem or antivirus cause.
- C: and D: reported healthy NTFS volumes with approximately 85.5 GB and 535.8 GB
  available. A disk-performance sample was taken **after** the diagnostic ended;
  its idle result does not characterize load during verification.

## Conclusion and remaining work

Docker absence is no longer the current prerequisite blocker. Slow local
verification also occurs outside the runner, while graph loading was fast.
The precise source of the cost remains unproven; no attribution to antivirus,
cache corruption, network delay, or runner deadlock is justified yet.

Next useful evidence is per-module zip/directory verification timing or a
bounded file-I/O trace during the actual verification. Preserve all integrity
checks, budgets and isolation rules. Do not repeatedly rerun all physical gates
until this shared preparation issue is understood. Historical ordinary-suite
failures remain unexplained; no main merge is authorized by these diagnostics.

## Follow-up: per-module timing and unchanged formal retry

A scratch-only diagnostic used the existing `golang.org/x/mod/sumdb/dirhash`
HashZip and HashDir implementations, comparing results against cached ziphash
values. It enumerated the 369 declared go.mod requirements, **not** the exact
full module build list used internally by `go mod verify`. Both diagnostic
runs used 16 workers and a 180-second diagnostic cap. No module bytes were
modified. Missing ziphash values and hash failures were reported separately.

- Session 22130 reached its cap with 347 modules completed, zero reported hash
  failures and zero missing ziphash values. Twenty-two requirements were not
  completed (16 active and six not started). Directory timing examples:
  pgx 15.02s; Goose 19.24s; Protobuf 82.23s; regexp2 95.11s;
  kin-openapi 119.78s; golangci-lint 123.43s. Buf's directory was still active.
- A subsequent isolated regexp2 check completed in 0.141s (zip 0.034s,
  directory 0.107s), with matching hashes. That alone cannot separate cache
  effects from concurrent-workload effects.
- Session 97277 repeated the same 16-worker diagnostic. It reached the cap with
  367 modules completed, zero reported hash failures and zero missing ziphash
  values. Only doubleclick and wazero directory checks remained active. Buf's
  directory completed in 7.94s in this run. This substantial improvement without
  lowering concurrency supports an access-state/cache contribution, but does
  not identify the operating-system mechanism or establish a reliable fix.
- Both capped diagnostic programs exited 124; `go run` reported wrapper exit 1.
  Neither is a completed full integrity verification or physical acceptance.
- Directory inventories: regexp2 had 1,913 files / 534,711 bytes;
  kin-openapi 1,667 / 6,000,511; Buf 3,032 / 11,635,411. Enumeration itself took
  0.175s, 0.046s and 0.265s respectively. These are not content-hashing times.

One unchanged formal retry (session 68756) then ran the exact first physical-gate
command above with the original profile, checks and deadlines. It again exited
1 with `Go graph go mod verify timed out after its bounded native wait`.
Its Go process was observed reading approximately 1.395 GB / 1.039 million
operations; a later sample around 5m10s showed approximately 1.401 GB read.
It did not produce a selected-test PASS. No other physical gate was launched.

Stop same-condition retries. A process-scoped file-access/performance trace is
the next useful diagnostic to distinguish slow file reads, metadata operations
and system filtering. Do not infer a specific antivirus cause, disable security
software, delete module caches, or relax runner deadlines from this evidence.

## Follow-up: file-level localization

Only the scratch diagnostic was instrumented. Its reader wrapper preserves an
underlying `io.WriterTo` implementation when available and records individual
open/read/close costs over 500ms. No system-wide trace has been collected.

- Inventory: doubleclick v1.0.0 contains 127,449 files totaling 68,726,688 bytes;
  wazero v1.12.0 contains 10,877 files totaling 42,615,714 bytes. PowerShell
  enumeration took 3.28s and 0.24s respectively.
- Isolated doubleclick session 95027 reached a 180s cap in the directory check;
  zip verification took 2.29s. No >500ms individual-operation event was emitted,
  but no directory hash result was obtained either.
- Session 45841 added enumeration-complete and every-10,000-open progress events,
  with a shorter 90s diagnostic cap. The zip took 3.72s; the directory file list
  was complete at elapsed 10.54s. File-open progress reached 60,000 at 15.23s,
  but 70,000 only at 89.07s. The process then reached its diagnostic cap.
  This localizes a substantial delay to the content-hashing traversal after
  enumeration, rather than enumeration alone. The 60,000-to-70,000 interval
  took approximately 73.84s, without a reported individual >500ms operation.
- Both capped programs exited 124 (`go run` wrapper exit 1). There is no complete
  doubleclick directory verification result and no formal-gate retry in this
  follow-up. No assertion of corruption or a specific security-product cause
  is supported by these measurements.

Local help confirms the built-in Defender performance recorder is available
and supports bounded recordings. It captures antivirus scan and kernel process
events, requires administrator privileges, and offers no capture-time process
filter in its exposed parameters. A recording could therefore contain unrelated
process names and file paths. Ask for explicit approval before expanding from
the current dependency-only diagnostic to that system-level capture; keep any
raw trace local and out of Git. Never interpret its availability as proof that
Defender caused the slowdown, or as permission to add antivirus exclusions.

## Approved 30-second Defender recording

The user explicitly approved one 30-second performance recording. Session 41494
used the built-in `New-MpPerformanceRecording -Seconds 30` while the isolated
doubleclick diagnostic was running. Recording and save completed successfully.
The owned diagnostic was stopped afterwards, so this was not a complete module
integrity verification. No exclusions, protection settings, dependencies,
runner checks or deadlines were changed.

Raw trace: `.superpowers/diagnostics/module-timing/defender-30s-20260924-191852.etl`,
2,315,801 bytes. The directory is Git-ignored. Keep this trace local: it may
contain unrelated process/file information. Do not stage, upload or publish it.
The analysis output was filtered to the exact local diagnostic executable.

Target process 23640, `c12-module-timing.exe`:

| Metric | Recorded value |
| --- | ---: |
| Scan count | 3,169 |
| Aggregate scan duration | 29.1268469 seconds |
| Average scan duration | 9.1911 milliseconds |
| Median scan duration | 9.9965 milliseconds |
| Maximum scan duration | 440.3795 milliseconds |

The report's raw durations are 100-nanosecond units. The five longest target
scan records all referred to doubleclick `parser/testdata/.../explain_*.txt`
files, with `ScanType=RealTimeScan`, `Reason=OnOpen`, and
`SkipReason=Not skipped`. The diagnostic log reached 70,000 opened files at
13.50 seconds, then was still in directory hashing at the 30-second heartbeat.

This is direct evidence that Defender's on-open scanning contributes substantial
overhead to the measured many-small-file workload. Aggregate scan time must not
be treated as an exact wall-time percentage; a controlled intervention was not
performed. The capture used the diagnostic executable, not the formal runner's
Go process. It does not prove that Defender alone caused both formal timeouts,
and it does not explain the historical cancellation/CreateProcess failures.

Next non-policy-changing option: complete a separate bounded dependency
pre-verification with protection enabled, then recheck the unchanged formal
gate. Such preparation is not acceptance, cannot replace the formal check, and
may not prevent recurrence after cache/system-state changes. Do not prescribe
security exclusions or relax the approved runner budgets from this result.

## User-approved independent pre-verification attempt

The user approved proceeding with the suggested protection-enabled preparation.
Session 45694 ran the pinned Windows Go 1.26.5 executable with `mod verify` from
the coordinator worktree, using the fixed module cache, offline proxy settings,
disabled Go user configuration/workspace, readonly integration flags and the
runner's target platform. This was a separate preparation operation, **not** a
change to the formal runner's budgets or a replacement for its integrity check.
No new Defender recording was requested or collected.

The owned process (PID 39772) was bounded to 900 seconds. Read-only samples showed
592,037,594 bytes / 247,907 read operations, then 1,408,523,666 bytes / 1,171,005
operations. It did not exit successfully within the bound. The controller
terminated only that owned process at 900.69 seconds, reporting:

```text
PREVERIFY_15M_LIMIT
ExitCode=-1; Stdout=""; Stderr=""
```

No complete verification result was obtained. Empty captured output is not
proof of integrity or absence of errors in work that was interrupted. Since
pre-verification did not succeed, the conditional formal-gate retry was **not**
launched. Protection settings, module contents and formal acceptance policies
remain unchanged.

The proposed preparation approach did not resolve the blocker in this attempt.
Do not repeat or lengthen it automatically. Further progress requires evaluating
a suitable verification environment or a separately approved dependency/tooling
isolation design while preserving integrity coverage. Such a design change has
not been approved or implemented by this diagnostic work. All physical-gate
acceptance and historical ordinary-suite causation blockers remain open.
