# C12 Lightweight Wait Diagnostic Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline, task by task. Await plan approval. Preserve the design's inline author review boundary.

**Goal:** Observe one root-only offline verification attempt without changing its execution context or120s limit.

**Architecture:** A restricted executor owns Go and its watchdog. A separately launched normal-desktop collector owns one read-only query worker, consumes identity-bound JSON requests, and writes bounded samples. Collector/worker failure cannot extend Go's deadline.

**Tech Stack:** Existing PowerShell7.6.5,.NET Process/Stopwatch,CIM Win32_Process,exclusive JSON files. No dependencies, compilation or system-wide tracing.

**Spec:** docs/superpowers/specs/2026-11-06-c12-lightweight-wait-diagnostic-design.md, approved after2bec1f67.

## Global constraints and budget

- Existing codex/c12-b01-task9-coordinator worktree only. No main merge, security
  setting changes, downloads, Docker/C12/devtools execution or old-root reuse.
- Fixed Go1.26.5 path and SHA256D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F;
  root command exactly `mod verify`. Copy previous child-only offline environment
  from archived sampled launcher, including GOMAXPROCS=2 and GOFLAGS=-mod=readonly.
- One RED and one GREEN test-script invocation, followed only if green by one
  cross-process preflight and at most one120s Go attempt. Unexpected result stops
  the stage, no repair/retry or extra test invocation without new approval.
- GREEN permits at most two harmless owned PowerShell lifecycle fixture children:
  one ordinary exit, one sleeping child stopped after1s. No Go in unit tests.
- Real attempt: one executor, one collector, one persistent read-only worker, one
  Go child. The worker is an implementation choice to enforce a5s per-query bound
  without assuming CIM's OperationTimeoutSec alone is a hard wall-clock bound.
  No worker restart, concurrent requests or extra query processes.
- Up to24 scheduled samples at5s intervals. No catch-up bursts. End at Go exit or
  original120s deadline; observer overhead may affect performance and is not zero.
- Initial handshake at most60s, target-arrival wait at most15s after preflight
  acknowledgement. These precede/coordinate observation and do not extend the
  independent120s verification timer. Missing acknowledgement prevents Go launch.
- Timeout cleanup: kill only retained owned process handles/trees, wait at most5s
  for exit confirmation; unconfirmed exit is a blocker, never success or retry.
- Only own parent/Go target counters, IDs/times and expected executable metadata;
  no command-line/environment enumeration, arbitrary process queries or file contents.
-44/34/6/217 unchanged;NOT_ACCEPTED. Keep all original evidence and publish portable
  records. No independent security/full-suite acceptance claim.

## Files and shared contracts

Create under docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/:
`runtime-cache-wait-common.ps1` (definition-only validation/publication),
`runtime-cache-wait-executor.ps1` (restricted owner/watchdog),
`runtime-cache-wait-collector.ps1` (normal-desktop coordinator),
`runtime-cache-wait-worker.ps1` (read-only query loop),
`runtime-cache-wait.tests.ps1` (tests).
Modify only status/recovery/readiness documents; no production runner/module/pins.

Session root: `.superpowers/sdd/2026-11-06-c12-cache-wait-v1/<phase>`; phases
Red,Green,Live are exclusive, claimed once. Reject reparse ancestors/occupied roots.
Publish JSON via CreateNew temporary file then non-overwriting same-directory move;
consumers never read incomplete files. Preserve abandoned temporary records.

Common APIs:
- `Assert-C12WaitIdentity($Expected,$Observed)` throws closed IDENTITY on mismatch.
- `Write-C12WaitRecord([string]$Root,[string]$Leaf,$Value)` returns Path/Length/SHA256.
- `Get-C12WaitDecision($NowMs,$DeadlineMs,$Exited,$CollectorReady)` returns one of
  WAIT,START,STOP,EXITED; clock and observations supplied, no process launch.

Identity schema1: SessionId32hex,Role(parent/go),PID positive integer,
CreationUtcTicks integer,ExecutablePath canonical absolute path,ExecutableSHA25664hex.
Use executor-retained Process identity; worker independently checks path/start time
before and after querying. Compare CIM creation time at its representable precision
as an additional check, never PID alone. Any unobservable identity fails closed.
Request schema: SessionId,Sequence(integer),Phase(preflight/sample),Identity.
Worker responses: matching bindings,Status(OBSERVED/UNAVAILABLE/EXITED/IDENTITY),
TimestampUtc,CPUms,WorkingSetBytes,PrivateBytes,ReadOperations,ReadBytes,
WriteOperations,WriteBytes,ThreadStates,WaitReasons,UnavailableFields. Missing values
are null plus explicit unavailable names, not zero. No arbitrary script commands.
Cap each message/record at64KiB, total retained observer data at4MiB; exceeding
limits stops observation. Samples retain raw counters and validated deltas;
decreasing counters invalidate that delta rather than becoming negative throughput.

## Review focus

1. PID reuse/identity precision: refuse wrong process, never infer from name.
2. Collector blocked or delayed: Go watchdog remains independent; no extended run.
3. Missing/null metrics: report unavailable, never fabricated zero activity.
4. Partial/occupied receipts: fail closed, preserve original bytes.
5. Collector elevated while Go remains restricted: no cross-boundary execution
   instructions or source-path substitution through request files.

## Task1 — Contracts and owned-process lifecycle

- [ ] Write tests before implementation: wrong PID/start/path/hash or session rejected;
  occupied output sentinel unchanged; partial temp not accepted; missing metrics stay
  null; decreasing counters unavailable;64KiB/4MiB limit; no acknowledgement means no
  launch; collector loss never changes120000ms deadline. Add clock-boundary assertions:
  `Get-C12WaitDecision 120000 120000 $false $true` must be STOP, and no START without readiness.
- [ ] Run `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-cache-wait.tests.ps1 -Phase Red`
  once. Expected missing contract implementation; save receipt.
- [ ] Implement common APIs and the three role scripts. Executor first publishes its
  parent identity and waits; collector launches only its fixed worker and queries
  that parent before exclusive preflight-ready publication. On failure publish stop.
- [ ] Executor launches exact Go only after valid ready receipt, starts Stopwatch
  immediately before launch, publishes target identity, and waits independently.
  Reuse previous120s ownership cleanup, asynchronous stdout/stderr draining and
  lockfile hashes. Preserve restricted launch; elevated collector never starts Go.
- [ ] Collector submits one JSON request at a time to the persistent worker over
  redirected stdin/stdout; bound response wait5s. Worker projects only listed CIM
  fields and thread state aggregates. Query timeout/mismatch/parse failure stops
  collection and kills only its worker. Record stop reason; never change executor timer.
- [ ] Add lifecycle tests using two harmless children: normal exit confirmed; sleep
  fixture terminated at1s with confirmed exit. Simulate missing/late collector and
  wrong target identity without launching Go. At most16 contract/lifecycle cases.
- [ ] Run same test script `-Phase Green` once. Expected WAIT_DIAGNOSTIC_TESTS_VERIFIED,
  all required cases matched, no leftover owned fixture process. Stop unexpected.
- [ ] Inline diff/privacy/ownership review; commit explicit tooling checkpoint. Do not
  claim a normal-process fixture proves restricted/elevated cross-process access.

## Task2 — One real preflight and one bounded observation

- [ ] Start executor in original restricted context in exclusive Live root; keep its
  tool session active. Then start collector separately via approved normal-desktop
  permission with exact same fixed root/SessionId. Do not elevate executor.
- [ ] Require parent-identity and cross-process I/O preflight before Go launch. If
  unavailable/mismatched/late, stop with NOT_STARTED and do not retry.
- [ ] If preflight passes, observe at most one root verification for120s, at most24
  scheduled samples. Poll tool sessions without blocking user updates beyond60s.
  A collector failure is recorded; executor still owns the unchanged cutoff.
- [ ] Confirm executor/collector/worker termination, four module files unchanged,
  and receipt identities. Lack of positive exit evidence blocks completion.
- [ ] Report observed completion/timeout separately from diagnostic interpretation.
  Increasing I/O/CPU gives activity evidence, not disk bottleneck/lock-owner proof.
  Stable counters plus generic wait reasons can remain INCONCLUSIVE.

## Task3 — Recovery handoff

- [ ] Package exact tooling/claims/results/samples into a portable archive; decode and
  compare length/SHA256 with originals. No private cache, binaries or credentials.
- [ ] Update status/recovery/readiness with consumed budget, unavailable fields,
  timing/collector overhead and inference limits. Preserve old attempts unchanged.
- [ ] Stage explicit files, commit/push existing branch, verify remote HEAD equality
  and clean worktree. Stop; no devtools/live C12 or automatic next experiment.

## Self-review

Spec coverage maps to Task1 contracts/watchdogs,Task2 separation/preflight/bounds,
Task3 evidence/privacy/recovery. Single persistent worker is explicit, bounded
and owned; no system trace or new dependencies. Exact environment and verification
context remain unchanged. No counter combination guarantees root-cause attribution.
PLAN_READY_FOR_REVIEW; no tooling implemented or new verification attempted.
