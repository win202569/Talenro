# C12 Wait Diagnostic V2 Implementation Plan

## Execution outcome — 2026-11-06 budget consumed

Approved after737d73c0. RootRedc07ff4 expected2 failures,RootGreen9a0fc3 4/4,
UTC5516be8/8,loop8d74a7 8/8. Tooling4357dbbd committed/pushed before live run.
Prelaunch493b58 verified pin/modules/sources. Preflight passed,22 bound samples
collected. Executore8905d/session1473→c2ca56 timed out120048.1025ms with owned
exit confirmed,empty output,unchanged module files. Collectorf6d739/session13532
→e0dad8 ended at DEADLINE with worker exit confirmed. No retry or new experiment.
Task1 gates complete;Task2 bounded observation performed,Go verification NOT passed;
Task3 portable120file109395byte archive decode/source-verifiedcd3428. All slots
consumed. Detailed interpretation/recovery in latest status documents. Old records
preserved; currentv2/Live occupied. Further execution needs fresh approval.
Inline author review only;44/34/6/217,NOT_ACCEPTED.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline, task by task. Preserve the existing inline author-review method. Await user approval before implementation/execution.

**Goal:** Obtain one bounded root-verifier observation with the UTC and end-reason repairs, without overwriting earlier evidence.

**Architecture:** Keep the restricted executor and separate normal-desktop read-only collector/worker. Change only the two fixed live evidence paths to a fresh v2 root; first exercise offline path and regression gates. The executor owns the unchanged Go deadline independently of collection.

**Tech Stack:** Existing PowerShell7.6.5/.NET/CIM and cached Go1.26.5; no new dependencies.

**Spec:** docs/superpowers/specs/2026-11-06-c12-lightweight-wait-diagnostic-design.md. Prior plan exhausted; repairs at c1ba31fc and659b1707. This plan requires fresh approval and does not revive old budgets.

## Global constraints and proposed budget

- Existing codex/c12-b01-task9-coordinator worktree only; inline review, no main merge.
- Fresh live root `.superpowers/sdd/2026-11-06-c12-cache-wait-v2/Live`.
  Preserve all v1/UTC/loop roots and archives. Never delete or reuse an occupied root.
- Offline: one root-path RED, one root-path GREEN (four cases each), one UTC
  V2Gate invocation (eight cases), one loop V2Gate invocation (eight cases).
  These gates create no child processes and never invoke complete live scripts.
- Only if all gates pass: one restricted executor, one collector, one persistent
  read-only worker, one parent preflight and at most one root Go child. No restart.
- Go exactly `mod verify`, pinned executable
  `C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe`,
  SHA256D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F.
- Preserve child-only offline environment: remove inherited GO*/GIT_* keys, set
  GOENV=off,GOWORK=off,GOTOOLCHAIN=local,GOPATH=C:/Users/Lenovo/go,
  GOMODCACHE=C:/Users/Lenovo/go/pkg/mod,GOPROXY=off,GOSUMDB=off,GOAUTH=off,
  GOVCS=all:off,GOFLAGS=-mod=readonly,GOMAXPROCS=2. No persistent settings.
- Go120000ms; handshake60000ms; target wait15000ms; per-query5000ms;
  at most24 samples spaced about5000ms, no catch-up bursts; owned exit wait5000ms.
  Collector failure never extends Go's cutoff. No elevation of Go/executor.
- PID/start/path/hash/session binding,64KiB messages/4MiB retained observer data,
  unavailable fields remain null. No arbitrary processes/command lines/env contents.
- No downloads, compilation,devtools,C12,Docker workload,system tracing,security
  changes, automatic retries or extended deadlines. Unexpected gate/runtime result
  stops new work; finish only owned cleanup, read-only diagnosis and evidence handoff.
-44/34/6/217 unchanged,NOT_ACCEPTED. Go completion is not whole-project acceptance.

## Review focus

1. Either role still resolves v1: root RED/GREEN must exercise both actual path expressions.
2. Occupied root: sentinel preserved and exclusive creation refused.
3. UTC roundtrip/culture drift: eight existing UTC cases must pass unchanged.
4. Early exit hides cause: eight existing loop/receipt cases must pass unchanged.
5. Real process access differs from offline doubles: one real parent preflight,
   then inspect bound sample responses; never infer success from collector exit0 alone.

## Task1 — Fresh-root preparation and offline gates

**Files:** Under docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/ modify
runtime-cache-wait-executor.ps1 and runtime-cache-wait-collector.ps1 (fixed root only),
runtime-cache-wait-utc.tests.ps1 and runtime-cache-wait-loop.tests.ps1 (add V2Gate
phase only); create runtime-cache-wait-root.tests.ps1.

**Interfaces:** root test takes `-Phase RootRed|RootGreen`; UTC/loop tests accept
`-Phase V2Gate` with existing assertions unchanged. Live script parameters unchanged.

- [ ] Create four offline root cases: actual executor path equals absolute v2/Live;
  actual collector path equals the same literal expected path; occupied disposable
  root rejects with sentinel bytes unchanged; fresh disposable root is exclusively
  claimed and second claim rejected. Use AST to isolate path expressions, omit
  executor's New-C12WaitRoot wrapper when resolving its literal Join-Path expression.
  Root cases must not create Live or execute launch code. Test data roots:
  `.superpowers/sdd/2026-11-06-c12-cache-wait-v2/RootRed` and `/RootGreen`.
- [ ] Run root test `-Phase RootRed` once; expect the two old-path assertions to
  fail, with no real processes. Preserve receipt. Other failures stop implementation.
- [ ] Replace v1/Live with v2/Live in the two role scripts only; add V2Gate to both
  existing test ValidateSet lists. Their phase-relative directories remain exclusive.
- [ ] Run root test `-Phase RootGreen`, UTC test `-Phase V2Gate`, loop test
  `-Phase V2Gate`, each once. Expect4/4,8/8,8/8 respectively, exit0 and no children.
  Any unexpected result stops; no repair/re-run under this budget.
- [ ] Inline diff check: no pin,deadline,identity,worker,environment or production
  changes. Commit/push explicit tooling checkpoint before real observation.

## Task2 — One preflight and at most one120s observation

**Interfaces:** executor `-SessionId <new32hex>`; collector same SessionId plus
`-ParentPID` and `-ParentCreationUtcTicks` from this executor's output. Root is fixedv2.

- [ ] Confirm v2/Live absent, pinned executable hash and four module-file hashes;
  retain exact five diagnostic source hashes. Record session and budget ledger.
- [ ] Start executor in restricted execution, keep session active; immediately
  launch collector separately with normal-desktop permission and exact observed
  parent coordinates. No process-name searches, extra probes or alternate launches.
- [ ] Require identity/I/O parent preflight before Go launch. Failure/missing/late
  acknowledgement means NOT_STARTED; no retry, permission widening or fallback.
- [ ] If started, observe at most120s; keep user updates under60s. On collector
  failure do not restart it; executor still completes its existing cutoff/cleanup.
- [ ] Confirm completed executor/collector tool sessions and owned Go/worker exit
  receipts. Uncertain exit blocks completion. Compare four go.mod/go.sum hashes.
- [ ] Inspect SessionId/identity/sequence/raw counters/deltas/unavailable fields
  and EndReason. Report actual sample count; zero samples means no Go activity
  evidence even with COLLECTION_ENDED. A deadline reason is not root-cause proof.
  Separate Go completion/timeout from observation quality; no new experiment.

## Task3 — Portable recovery checkpoint

**Files:** create portable-c12-cache-wait-v2.json alongside diagnostics; update
docs/roadmap/current-status.md,docs/runbooks/repository-recovery.md,
docs/roadmap/2026-11-06-c12-shadow-runtime-readiness.md and this execution ledger.

- [ ] Archive exact new claims,source hashes/source bytes,offline/live receipts and
  samples, with synthetic data clearly labelled. No cache binaries or credentials.
- [ ] Decode archive and compare every length/SHA256 with original bytes; normalize
  outer JSON to repository LF before reporting any envelope hash. Preserve originals.
- [ ] Record consumed slots, failure/exit/sample evidence and interpretation limits.
  Stage explicit reviewed files, commit/push existing branch, confirm remote SHA
  equals HEAD and clean worktree. Stop without devtools/C12/main merge.

## Self-review / handoff

Spec separation,bounds,privacy/identity and evidence requirements retained. Task1
covers new path isolation and both repaired contracts; Task2 handles actual access
and independent cleanup; Task3 covers portable evidence. Offline gates do not
prove process/CIM integration. No scope expansion beyond a newly budgeted attempt.
PLAN_READY_FOR_REVIEW. This planning turn launches nothing and changes no code.
