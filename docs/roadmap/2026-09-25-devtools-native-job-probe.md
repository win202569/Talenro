# Development tools: Windows containment feasibility probe

Historical probe record. The later approved direction is PowerShell Core 7.6.5,
not a native outer supervisor. Publication and current status are recorded in
[the recovery handoff](2026-09-25-progress-recovery-handoff.md). Statements below
about approvals or no commit/push describe the time of this probe only.

Date: 2026-09-25. Result: feasible in the tested topology, NOT a production fix.
The user approved a throwaway probe after both Bash cancellation boundaries failed.

## Scope and topology

A standalone native Windows supervisor joins its own unnamed kill-on-close Job
before starting Bash. Its Job handle is not inherited. Bash starts a benign native
fixture which starts one sleeping child. The driver retains native process handles
before cancellation; it does not discover or terminate unrelated processes.
A separate sleeping canary is launched outside the Job.

No C12 runner, encoded PowerShell loader, Docker, network dependency, security
setting, production script or acceptance deadline is changed or executed by this
probe. The Go program uses the standard library and Windows APIs only.
Sources/binary/results are ignored scratch under the current plan's
`native-job-spike/` directory. They are throwaway, not approved production code.

## Fresh final results

| Entry | Normal exit | Launcher killed | Interpreter killed | Supervisor killed |
| --- | --- | --- | --- | --- |
| Git `bin/bash.exe` wrapper | PASS | PASS | PASS | PASS |
| Git `usr/bin/bash.exe` interpreter | PASS | PASS | PASS | PASS |

In all eight cases, retained-handle waits confirmed the launched Bash process(es)
and both native fixture processes exited. The external canary remained alive and
was subsequently cleaned up by the driver. Observed cancellation-to-confirmation
times were 1–6 ms in this final run; these are observations, not performance promises.

The negative control deliberately omitted Job containment in the disposable
supervisor, then killed the interpreter. The probe detected a surviving owned
process within its two-second check and used retained handles for fixture cleanup.
Thus the positive results are not based only on the shell returning a nonzero code.
The final probe process returned exit 0 after eight positive cases and the expected
negative detection. Full output is in ignored `native-job-spike/result.log`.

## Limits and recommendation

This proves feasibility when the native supervisor is already outside and above
Bash, with containment established before any child starts. It does NOT show that
adding a helper inside an already-running Bash script covers its launcher ancestor.
The tested outer supervisor is a new entry topology, not a drop-in change to the
existing public `bash scripts/verify-devtools.sh` contract.

Production integration therefore needs a scoped design amendment specifying the
Windows public entry, C11 dispatch, pre-launch containment, exact-handle ownership,
deadline/output/cleanup contracts, and behavior if Job setup fails. Arbitrary shell
ancestors cannot silently be adopted or killed. Keep the current Bash cancellation
tests, and add the supported supervisor entry tests rather than declaring those
failures waived. Native Unix Bash behavior remains untested by this Windows probe.

Task 1 remains incomplete. The Defender/C12 investigation, ordinary-suite failure,
I2/I3 and physical acceptance blocks remain open. No commit, push or merge occurred.
