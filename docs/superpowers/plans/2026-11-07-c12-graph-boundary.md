# C12 Graph Boundary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline, task by task, retaining the user-approved author-review boundary. Await approval of this plan before execution.

**Goal:** Determine whether one offline native graph command completes within60s without cached-content hashing.

**Architecture:** Reuse unchanged bounded PowerShell owner,with native graph text and no phase Session. Preserve original outputs and distinguish graph completion from root verification or historical cause attribution.

**Tech Stack:** PowerShell/.NET,pinned Go1.26.5;no build,installation or new runtime code.

**Spec:** docs/roadmap/2026-11-07-c12-root-verify-analysis.md,especially the narrower diagnostic recommendation and interpretation gates.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator;
  branch codex/c12-b01-task9-coordinator.44/34/6/217,NOT_ACCEPTED unchanged.
- New workspace .superpowers/sdd/2026-11-07-c12-graph-boundary,exclusive Live child.
  Do not overwrite any old workspace,claim or record.
- One metadata prerequisite check and at most one restricted Go process tree.
  Exact arguments @('mod','graph');no -go,-x,module argument or follow-on command.
- Go C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe;
  SHA256 D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F.
- Owner docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/cache-phase/owner.ps1;
  SHA25633782F9A2F723AEAAF816D9C3ADD25E3AA9D80954BBAB9F314B8309B15DD0B7E.
- Common docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-cache-wait-common.ps1;
  SHA25677824603B0817D32B447C50CA3818F5E34E1924D8403A800073165C65081566F.
- Existing owner clears inherited GO*/GIT_* child settings. Set GOENV=off,
  GOWORK=off,GOTOOLCHAIN=local,GOPATH=C:/Users/Lenovo/go,
  GOMODCACHE=C:/Users/Lenovo/go/pkg/mod,GOPROXY=off,GOSUMDB=off,GOAUTH=off,
  GOVCS=all:off,GOFLAGS=-mod=readonly,GOMAXPROCS=2.
- GOCACHE,GOTMPDIR,TEMP,TMP each under new workspace/isolated/<variable>.
  No persistent environment change or empty replacement for real GOMODCACHE.
- DeadlineMs60000. Existing owned termination/exit confirmation5s and bounded
  drain/cleanup apply;no extension of Go runtime. Existing4MiB evidence limit,
  including4063232-byte combined stream cap;no enlarged output quota.
- No cache inventory/repair/copy/delete,download,go version,verify,tests,build,
  collector,CIM,trace,Docker,devtools,C12,security changes or main merge.
- Tooling is unchanged;no extra RED/GREEN runtime under this invocation-only plan.
  If tooling changes become necessary,stop for separately scoped repair.
- Offline flags are not a filesystem read-only mount. Completion does not prove
  cache integrity,all dependencies available,application correctness or prior cause.

## Review Focus

1. Wrong cache/toolchain:verify coordinates and hashes before launch.
2. Native text parsed as phase JSON:omit Session,inspect raw output only after exit.
3. Wrapper exit0 confused with Go completion:use owned receipt and strict gates.
4. Graph edge count mistaken for verified modules:report edges only,not BuildList.
5. Sensitive output publication:review every bounded stream before archiving bytes.
Task1 checks pins,Task2 checks invocation/output/receipt,Task3 checks interpretation
and publication. These are direct evidence checks,not new implementation branches.

## Task1 — Freeze prerequisites and invocation

**Files:** Create ignored invocation.ps1,progress.md,preflight.json under this workspace.
**Interfaces:** Assert-C12WaitPath,New-C12WaitRoot,Get-C12WaitModuleHashes,
Write-C12WaitRecord and Invoke-C12PhaseOwned from pinned owner/common.

- [ ] Verify branch,clean worktree,Live absent and remote HEAD equals approved
  plan checkpoint. Record source commit and remote confirmation.
- [ ] Check ancestors of repo,Go,owner/common,real cache root,new evidence root
  with Assert-C12WaitPath;check required paths exist. Missing/reparse/pin mismatch
  stops without Go. Do not enumerate module-cache contents.
- [ ] Verify three pinned hashes and four module hashes against both Before/After
  in portable-c12-root-verify-completion.json's Live/stage-result.json entry.
  Decode baseline entry in memory and verify its stored length/SHA before using it.
  This avoids requiring the prior computer's ignored files.
- [ ] Prepare exact recipe/environment and syntax-parse without launch. Record
  actual hashes,source,recipe SHA,args,workdir,DeadlineMs60000 and empty Session
  in immutable preflight. Claim fresh Live and record before hashes before launch.
  Expected:all prerequisites match;no Go process yet;no modified tooling.

## Task2 — One graph attempt and classification

**Files:** Live/exclusive.claim,invocation-claim.json,process-result.json,
stdout.raw,stderr.raw,stage-result.json (immutable evidence).
**Interface:** Invoke-C12PhaseOwned -Executable <pinnedGo> -Arguments @('mod','graph')
-WorkingDirectory <repo> -Environment <allowlist> -Root <exclusiveLive>
-DeadlineMs60000. Omit Session entirely.

- [ ] Invoke once restricted;retain terminal session and wait for result/cutoff.
  Update user within60s;no second workload. Empty output is pending,not stalled.
- [ ] Confirm owned ExitConfirmed;capture after module hashes and raw stream
  byte lengths/SHA256. Preserve bytes including trailing newline and partial output.
- [ ] Classify GRAPH_COMMAND_COMPLETED only if Started=true,ExitConfirmed=true,
  ExitCode=0,TimedOut=false,Failure=null,EndReason=EXITED,four hashes unchanged,
  stderr length0,and graph text is nonempty valid UTF-8 with at least one edge.
  For this repository,each nonempty LF-separated line (optional trailing CR)
  must contain exactly two nonempty whitespace-separated tokens;no interior blank
  lines. An optional final newline is allowed. Count edges,not selected modules.
  This is structural validation,not a new semantic module-version parser.
- [ ] Otherwise classify STOP/NOT_PASSED and list the exact failed gates;retain
  output and process receipt. Empty exit0 is inconclusive,not success. Decode or
  shape failures must not prevent preserving hashes/bytes and process outcome.
  Expected:finite observed result,not guaranteed completion or root-cause proof.
- [ ] Stop runtime work regardless of outcome. Do not retry,fetch missing metadata,
  run root verify or increase cap. GRAPH_COMMAND_COMPLETED only describes this
  invocation;startup/traversal/output time and cache warmth prevent exact attribution.

## Task3 — Review,portable evidence and checkpoint

**Files:** Create docs/roadmap/2026-11-07-c12-graph-boundary-result.md and
docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/portable-c12-graph-boundary.json;
update current-status.md,repository-recovery.md and this plan outcome.

- [ ] Review raw output and error text for secrets/private identifiers or unrelated
  local paths before publication. If safe,archive original raw bytes. If unsafe or
  uncertain,keep originals local and archive only length/SHA plus explicit omission
  reason;mark portable evidence incomplete and seek publication direction. Never
  silently redact originals or claim a redacted file is original evidence.
- [ ] Archive exact recipe,owner/common source,preflight/claim/receipts and safe
  raw output;exclude binaries,caches,module content,credentials. Publication review
  also applies to receipt error fields. Decode each included entry and compare
  length/SHA against its original;LF outer JSON. Record archive hash/count.
- [ ] Report elapsed time,receipt,gates and edge count if complete. State that graph
  is not content integrity or BuildList. A fast graph only narrows a future question;
  timeout cannot distinguish graph loading from startup/traversal/output. Preserve
  NOT_ACCEPTED/counters and old failures. No independent review claim.
- [ ] Review explicit file list and diff --check;require empty pre-existing index,
  stage only list,verify staged set,commit/push existing origin/branch. Confirm remote
  SHA==HEAD and clean worktree. Do not delete local evidence. Runtime budget consumed.

## Self-review and handoff

Task1 produces exact immutable inputs for Task2;Task2 original evidence feeds
Task3 with privacy gate. No producer/consumer mismatch found. Runtime budget,
failure gates and ambiguous-output behavior are explicit. No custom runtime source
or tests needed. Inline author review retained. PLAN_READY_FOR_REVIEW;no Go run
while writing this plan. Approval covers at most one60s graph attempt,not root verify.

## Execution outcome

User approved after1197cbd2. Task1 remote83c14e and syntax6e25e2 passed;
preflight checked fixed pins/paths and decoded portable baseline.
Task2 single invocation7f9b8f completed306.9959ms,exit0,exit confirmed,no timeout,
1059 edges,empty stderr,unchanged module hashes;all classification gates passed.
Task3 output review9064b5 found no sensitive/unusual tokens or error text;original
outputs retained and portable archive checked before checkpoint. See result report
docs/roadmap/2026-11-07-c12-graph-boundary-result.md. Inline author review retained.
No additional runtime or source changes;budget consumed. NOT_ACCEPTED remains.
This outcome supersedes PLAN_READY_FOR_REVIEW and historical unchecked steps above.
