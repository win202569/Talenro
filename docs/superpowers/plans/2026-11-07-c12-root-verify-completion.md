# C12 Root Verify Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline, task by task, with the existing approved author-review boundary. Await this plan's approval before execution.

**Goal:** Obtain one bounded native root cache-integrity verification result.

**Architecture:** Reuse unchanged tested PowerShell owner with raw native Go output.
No phase JSON parser session or separate collector. Archive the exact invocation,
source/pin/module hashes and original outputs whether the attempt passes or fails.

**Tech Stack:** Existing PowerShell/.NET,pinned Go1.26.5;no build/install/dependencies.

**Spec:** docs/superpowers/specs/2026-11-07-c12-root-verify-completion-design.md.

## Global constraints

- Existing codex/c12-b01-task9-coordinator worktree only;44/34/6/217,NOT_ACCEPTED.
- New root .superpowers/sdd/2026-11-07-c12-root-verify-completion/Live;never overwrite.
- One prerequisite check,one native Go attempt at most;no new tests/build/collector.
  No runtime/source implementation change is planned,so no extra RED/GREEN cycle.
  If source changes become necessary,stop and obtain a separate repair budget.
- Command exactly pinned Go with arguments @('mod','verify'),working directory
  D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator. No module argument.
- Go path C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe
  SHA256D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F.
- Owner docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/cache-phase/owner.ps1
  SHA25633782F9A2F723AEAAF816D9C3ADD25E3AA9D80954BBAB9F314B8309B15DD0B7E.
- Common runtime-cache-wait-common.ps1 in that parent directory,
  SHA25677824603B0817D32B447C50CA3818F5E34E1924D8403A800073165C65081566F.
- Child-only environment clears inherited GO*/GIT_* through existing owner,then
  GOENV=off,GOWORK=off,GOTOOLCHAIN=local,GOPATH=C:/Users/Lenovo/go,
  GOMODCACHE=C:/Users/Lenovo/go/pkg/mod,GOPROXY=off,GOSUMDB=off,GOAUTH=off,
  GOVCS=all:off,GOFLAGS=-mod=readonly,GOMAXPROCS=2. No persistent settings.
- GOCACHE,GOTMPDIR,TEMP,TMP under new workspace/isolated/<variable>;do not replace
  real GOMODCACHE with isolated test cache. No download/repair/copy of module data.
- DeadlineMs300000,existing5s owned exit wait and bounded drain;no retries.
  Empty output during running means pending,not stalled or successful.
- No C12,devtools,Docker,trace,CIM,security changes,main merge or wider acceptance.

## Review focus

1. Wrong cache/root gives misleading success:preflight asserts exact coordinates.
2. Native text treated as JSON:invoke owner with Session omitted/empty.
3. Empty exit0 or timeout mistaken for pass:read all output and apply strict contract.
4. Old output reused:exclusive new root and before/after hashes.
5. Native success overstated:report skipped uncached dependencies and NOT_ACCEPTED.
These are invocation/interpretation checks,not new code branches requiring tests.

## Task1 — Freeze invocation and prerequisites

**Files:** No production/source edits. Create invocation.ps1 and preflight.json
inside this plan's ignored workspace;include their exact bytes in final archive.

**Interfaces:** Existing Assert-C12WaitPath,New-C12WaitRoot,Get-C12WaitModuleHashes,
Write-C12WaitRecord and Invoke-C12PhaseOwned,all imported from pinned owner/common.

- [ ] Confirm branch and clean state;read this spec/plan;ensure Live absent.
- [ ] Check path ancestors for reparse points on repo,Go executable,module cache
  root and evidence root using existing Assert-C12WaitPath. No recursive real-cache
  inventory or hashing outside the command. Missing/unsafe prerequisite stops.
- [ ] Verify three pinned hashes above and four project module hashes against
  latest v3 Before/After receipt;record actual hashes,not merely a boolean.
- [ ] Prepare the exact single-run invocation recipe with fixed arguments/env,
  empty Session and300000ms bound;syntax-parse only (no launch). Record recipe hash,
  owner/common hashes,source commit,Go hash,working directory and environment
  allowlist in an immutable preflight/claim. Confirm plan/source checkpoint pushed.
  Expected:all coordinates/hashes match,no Go/collector launched.

## Task2 — One authoritative root attempt

**Interface:** Invoke-C12PhaseOwned -Executable <pinnedGo> -Arguments @('mod','verify')
-WorkingDirectory <repo> -Environment <allowlist> -Root <exclusiveLive>
-DeadlineMs300000. Do not supply helper phase Session.

- [ ] Claim new Live exactly once;record before hashes and invoke once restricted.
  Keep the returned terminal session;poll without launching any second workload.
  Update user within60s while waiting;never infer progress percentage.
- [ ] Wait for completion/cutoff;confirm owner receipt ExitConfirmed=true.
  On error only owned cleanup,read-only inspection and preservation are permitted.
- [ ] Capture after hashes and inspect raw stdout/stderr. Success requires all:
  Started=true,ExitConfirmed=true,ExitCode=0,TimedOut=false,Failure=null,
  EndReason=EXITED,four hashes unchanged,stderr length0,and stdout.Trim()
  case-sensitively equal to 'all modules verified'. Preserve whitespace/raw bytes.
- [ ] Write immutable result with ROOT_CACHE_VERIFY_PASSED only if all conditions
  hold;otherwise STOP/NOT_PASSED and exact reason. Record elapsed time,output
  hashes,lengths and unchanged/failing prerequisites. No devtools follow-on.
  Expected:evidence-backed native result,not guaranteed success within300s.

## Task3 — Evidence, interpretation and cross-machine checkpoint

**Files:** create docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/portable-c12-root-verify-completion.json;
update current-status.md,repository-recovery.md,2026-11-06-c12-shadow-runtime-readiness.md,
this plan outcome and the paired evidence/spec if necessary.

- [ ] Archive exact invocation/owner/common source,claims/receipts,bounded raw
  outputs,pin/module hashes. No executable,caches,credentials or module contents.
- [ ] Decode each entry and verify length/SHA256 against original;LF outer JSON.
- [ ] Report native cache-integrity result separately from dependency availability,
  build/test/C12 acceptance. On timeout report lack of completed result,not cause.
  Do not extrapolate two-module timing to all modules or alter caches/security.
- [ ] Stage only explicit reviewed files,commit/push existing branch,verify remote
  SHA equals HEAD and clean worktree. Preserve all old evidence. Budget consumed;
  stop without new runtime actions regardless of outcome.

## Self-review / handoff

Evidence/resource rationale is explicit;300s is proposed,not derived completion
estimate. Existing owner supplies tested bounded drains/exits;no new launch library.
No need for extra fixture children or full-suite reruns for unchanged owner.
Preflight covers coordinates/pins,Task2 covers exact success and cleanup,Task3
covers archive integrity and non-overclaiming. Inline method retained.
PLAN_READY_FOR_REVIEW. This turn is documents/read-only checks only;await approval.
