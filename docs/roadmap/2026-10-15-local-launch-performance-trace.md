# One local launch performance trace — 2026-10-15

## Authorization and scope

The user approved one short local performance-recording experiment around one
fresh sample, with the original2s sample budget, no retry, no Go/fuzz or security
configuration change. Raw system traces may contain process names/paths and must
stay local, unuploaded and uncommitted. No production/test implementation changed.
The temporary experiment/view scripts are ignored throwaway diagnostic artifacts,
not a supported new subsystem or authorization to repeat the experiment.

Existing WPR status reported no default recording before start. Recording used
the unique instance TalenroLaunchTrace001-c01b1bab and GeneralProfile, then the
unchanged child-readiness fixture ran once in fresh child-readiness-003. All WPR
stop commands targeted that instance only; no default/global cancel or security
setting change was used. Note that default WPR status alone does not enumerate
every named instance; unique naming and targeted cleanup avoid stopping others.

## Timing overrun and final session state

The intended recording target was10000ms, NOT a successfully enforced hard limit:

- Start requested UTC10:26:58.7771744, returned10:26:59.0180631, exit0.
- Stop requested10:27:09.0379053,10019ms after start returned.
- Stop/merge returned10:28:26.5152650, exit0:77477.3597ms after the stop request.
- ETL header reports actual coverage10:26:58.8340140–10:27:44.8367028,
  duration46.0026888s, longer than the approved approximate10s target. Do not
  describe this as a ten-second trace. Recording/stop overhead altered timing.
- While stop was pending, a readonly ETW query still listed the two uniquely
  named collectors. A later targeted stop attempt at10:28:37 found both already
  absent; both returned 'Data Collector Set was not found'. It stopped nothing.
  Original WPR stop had already succeeded. Final instance status was not recording,
  and exact-prefix ETW enumeration found no matching active collector.

The overrun was disclosed, not retried or hidden. Raw launch.etl is310378496 bytes
and remains in the ignored local diagnostic directory; no raw file was uploaded.
The trace header reports PerfCounter clock and0 lost buffers/0 lost events. Those
reported counters do not certify every decoded event or semantic interpretation.

## Single sample — still FAILED readiness, cleanup verified

child-readiness-003 ran once UTC10:26:59.0990489–10:27:01.2115014.
No child-ready.json: ChildReady=false, FixturePassed=false, Acceptance=false,
OriginalTestPassed=false. Native TimedOut=true, root not exited, TotalProcesses1,
forced owned termination requested; cleanup verified accounting0/list empty,
ShutdownElapsedMs23. No readiness success or original-test repair is claimed.
The experiment controller exited0 for successful trace start/stop, NOT for fixture
success; its saved FixturePassed=false must be read separately.

Native six-boundary receipt measured the creation bracket2082.4830ms, from budget
position5.8880ms to2088.3709ms. Resume ended at2088.8426ms, already beyond the
original2s budget. This repeats the measured kind of delay under tracing, but is
one newly authorized instrumented sample, not permission to repeat older ones.

## Offline analysis and export limits

Used installed WPAExporter with -symbols -symcacheonly: local caches only, no
symbol download or trace upload. Microsoft's [exporter documentation](https://learn.microsoft.com/en-us/windows-hardware/test/wpt/exporter)
describes profile-based CSV and time-range export; local tool help supplied the
cache-only option. Initial AppLaunch-template export returned process exit0 but
logged profile/Regions-of-Interest errors and produced partial aggregate tables.
Exit0 was NOT taken as proof of complete export. The first precise-CPU CSV also
had duplicate Ready/Waits headers, preventing direct Import-Csv parsing.

Derived a temporary local view from the installed template, retaining precise/
sampled CPU only, filtering to the recorded creator pwsh.exe (11008), expanding
rows and hiding duplicate Max columns. Re-exported the SAME ETL's first5s into
a separate directory; no second capture/sample. It produced precise-CPU10992
rows, including1153 detailed rows under the selected numeric thread/time filters.
Duplicate TraceLogging-start/ROI warnings remained. Sampled CPU stacks were mostly
empty or unresolved, so they were not used to infer a cause. Many precise stack
frames lack PDB function names; module-level and available managed frame evidence
must not be presented as fully symbol-resolved functions or complete scan telemetry.

## Focused evidence and interpretation

In the precise CPU data, leaf rows were selected with numeric CPU/old-thread IDs,
numeric switch-in time and a New Thread Stack containing IsolatedFuzzProcess::Run
and WdFilter.sys. Two such rows remained. The dominant one, Count1:

- Creator pwsh.exe (11008), pipeline thread29528.
- Wait stack includes the managed runner, native/PInvoke frames, KernelBase,
  ntdll, ntoskrnl, WdFilter.sys and FLTMGR.SYS (function PDBs generally missing).
- Readying process MsMpEng.exe (5236), thread9052; its stack includes MpRtp.dll,
  fltLib.dll and FLTMGR.SYS.
- Trace-relative last-switch-out0.2745637s, switch-in2.3360195s;
  Waits2061217.7us (2061.2177ms), Ready238.1us. This is scheduler waiting time,
  NOT2061ms of Defender CPU execution or a measured single scan's duration.
- Header-aligned approximate UTC10:26:59.1085777–10:27:01.1700334 lies within
  native creation-before/after UTC10:26:59.1049474–10:27:01.1874294. Relative
  decimal-to-UTC conversion is approximate, not an atomic clock sample.
- The wait is98.97885% of the native2082.4830ms creation bracket. The other
  matching leaf had only25.8us waiting; aggregate/parent rows were not summed.

Local WdFilter.sys metadata read afterward describes it as a Microsoft antimalware
file-system filter driver. Microsoft lists MsMpEng.exe as its antivirus executable
in the [component overview](https://learn.microsoft.com/en-us/defender-endpoint/microsoft-defender-antivirus-windows).
Together, the captured waiter/readier stacks and timing provide strong evidence
that the antimalware filtering path participates in the dominant observed creation
wait in THIS sample. This is an inference from correlated scheduling/stack data,
not proof of the exact scan target, policy, cloud behavior, defect or misconfiguration.
It does not retrospectively explain origin87, every old failure or the Go fuzz issue.
Profiling overhead and the long stop/rundown remain experimental limitations.

No Defender exclusion, disabled protection, registry edit, driver operation,
tool installation, timer extension or product fix was performed. A cause-specific
security change is neither justified nor authorized by this experiment.

## Evidence, recovery and remaining work

Ignored worktree-relative trace directory:
.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/launch-trace-001/.
Related sample directory: child-readiness-003 under the same diagnostic scope.
SHA256:

| Local evidence | SHA256 |
| --- | --- |
| launch.etl | 785DFC3BB97ACA6873714B33D515C3E29DDCAF5792292DCA0A2BCF48D1A61DE6 |
| child-readiness-003/process/receipt.json | 014228F9C2C4A01AF86C92085276A83B9A32A8731683F712D4B3A20F17FDAA0F |
| child-readiness-003/result.json | 5ACAFB9BAAB2DB8CF04A7A6AA899D75278B0417B58FD8EE25689A64CBBE29E50 |
| export-002/CPU_Usage_(Precise)_By_Process.csv | 9A7A9F47341A1CDB747891A9FAFD38C00EDCCFF9D67C3C3A447444F32B051E10 |
| focused-analysis.json | 736B7D50F55CD13B0163134351AA5C7038DCD69169E9E6EC89B9B8F98C4A3FF8 |

Only this public summary and recovery/status links are tracked. No raw trace,
CSV, system logs, temporary profiles/scripts, caches or machine activity are
committed. Cloning the latest codex/c12-b01-task9-coordinator restores the summary,
not the sensitive local ETL; read current-status.md and repository-recovery.md.
Trace001 and child-readiness003 authorizations/identities are consumed even if
their directories are absent on a new computer. Preserve the original failures.

This turn did not rerun helper/full product suites or Go/fuzz. The previous99
helper assertions per host remain historical results, not fresh full acceptance.
Original descendant regression/16 unrun checks, origin87, Envelope/Pair and
C11/C12/Task10/B01/production/main gates remain open. Further capture, exact scan
attribution, symbol acquisition, fixture redesign or mitigation needs separately
agreed scope; do not repeat recordings or weaken security to obtain a passing test.
