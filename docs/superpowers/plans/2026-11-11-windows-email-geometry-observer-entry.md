# Fixed Email Geometry Observer Entry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (recommended sequential inline) or superpowers:subagent-driven-development after plan confirmation. Steps use checkbox tracking.

**Goal:** Prepare a default-inert observer entry bound to the newly verified diagnostic artifacts, without launching an application.

**Architecture:** A new PowerShell entry supplies fixed policy and backend callbacks to the unchanged Invoke-TalenroC5Observation orchestrator. One internally generated run ID correlates command line, process manifest/samples and app geometry path. Synthetic tests replace every external effect; no observation loop is copied.

**Tech Stack:** Existing PowerShell7, Git, SHA256, native/managed observation modules; no packages or build changes.

**Spec:** docs/superpowers/specs/2026-11-10-windows-c4-email-geometry-design.md, especially Isolation and activation / Verification and later runtime boundary. Artifact evidence: docs/roadmap/2026-11-11-windows-c4-email-geometry-build.md.

Status: confirmed and implemented2026-11-11; Task1 committed d6da5ec9. Task2 review
found two Important issues: reparse test false-positive fixed RED→GREEN; failure-
category preservation remains BLOCKED pending entry resource/bookkeeping design.
Final synthetic642 pass, no launch. See tracked execution ledger for disposition.
The checklist below retains original requirements; review completion is not acceptance.

## Global Constraints

- Existing coordinator worktree/branch, PowerShell only; ordinary checkpoint push allowed.
- New EXE SHA256:CBDDE55CBFE2928E3F179250F9B0DFC560DB45D61FD05B8B014BADB0B0CFCDB3.
- New DLL SHA256:154E7E70F9FB2D2F980D16A920CA7A3CB69C5CA001CECF7D6A7DFFD1D1F043D9.
- Fixed relative EXE:clients/windows/Talenro.Demo.App/bin/email-geometry/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe.
- Fixed args:--layout-test --layout-language=en-US --layout-profile=rtl-mixed --layout-email-geometry-run=<same run ID>.
- No modifications to existing C4/C5 entries/tests, shared observer/native/identity modules, App source, old outputs or fixed hashes. No rebuild/restore/install.
- No live invocation, real process acquisition, CIM/native probe calls or app launch in verification. Default invocation returns NotStarted before imports/probes/writes.
- Future runtime route stays1launch/0retry,360000ms cap, operator startup30s/normal-close by300s/exit assessment30s. These operator timings are not newly implemented automated UI timers.
- No raw receipts/binaries committed. No PR/main merge or acceptance promotion. Live allowance0.

## Review Focus

1. Run ID mismatch can associate unrelated process/geometry records: reject mismatched context before resources; test exact args/manifest/paths.
2. Hash/source/branch/token failures must block acquisition: fault-inject every gate and assert zero writer/process effects.
3. Stale receipt/claim or reparse parent must not overwrite data: reject both observer and app destinations before launch, test no write/launch on collision.
4. Writer/open/flush/dispose failure must not leak resources or become success: fault-inject acquisition and shared-session failures, preserve primary plus cleanup categories.
5. Verified process exit is not valid geometry or readability: output GeometryEvidenceStatus=Unreviewed and Acceptance=NotAccepted even on synthetic exit0; no geometry parser/verdict in this entry.

## File map and interfaces

- Create clients/windows/tests/observe-email-geometry.ps1: default guard; policy, preflight, create-only manifest/writer and ten backend callbacks; invokes existing shared loop once.
- Create clients/windows/tests/observe-email-geometry.tests.ps1: synthetic fixtures, no external effects; follows established C4 tests' AST function extraction without running entry main.
- Update clients/windows/diagnostics/email-geometry.md: entry readiness versus live authority, operator route, two receipt locations, unreviewed geometry.
- Create docs/roadmap/2026-11-11-windows-email-geometry-observer-entry.md: verification/review/rulings ledger; update status/recovery and this plan.

Only public parameter: `[switch]$RunApproved`. Without it return exactly one object
Status=NotStarted; no imports, filesystem/identity checks or run ID generation. The
switch is not approval by itself; only a future separately approved run may use it.
No public path/hash/profile/language/run-ID/deadline/backend overrides.

Private prefix EG:
- Get-EGPolicy([string]$Root,[string]$RunId) -> hashtable with Root,RunId,Executable,Dll,
  Arguments,ExpectedCommandLine,ExeHash,DllHash,ReceiptRoot,GeometryRoot,GeometryFile,Sources.
- Assert-EGPreflight($Policy) -> no output; throws on any gate failure.
- Assert-EGReceiptDestination([string]$Directory) -> reject existing destination/reparse ancestors.
- New-EGReceiptWriter($Policy,$Context) -> one owned samples writer; manifest closed first.
- New-EGBackend($Policy) -> exact ten callbacks required by Invoke-TalenroC5Observation.
- ConvertTo-EGResult($Policy,$Context,$Observation) -> one wrapped object with
  ProcessObservation,GeometryEvidenceStatus,Acceptance,RunId,GeometryRelativePath;
  reject policy/context mismatches rather than wrapping unrelated observations.
- Small private Git/token/new-directory/create-new-stream/text-writer helpers mirror
  proven C4 resource ownership, without importing/executing the old entry or duplicating its loop.

Run ID canonical nonzero lowercase32 GUID, generated once internally before policy.
Context remains exactly RunId/ExpectedPath/ExpectedCommandLine as required by shared
module. Every writer/acquisition validates context against policy; run ID is never
recomputed during backend callbacks. Capture command functions in closures.

Preflight: ordinary token, exact existing dev branch, clean tree including untracked,
HEAD blob equality for new entry and all five consumed shared source files, exact
EXE/DLL hashes, conservative existing-target inventory (same C4 filter). Source gate
runs before importing shared modules; orchestrator preflight repeats before resources.
Check observer destination and app geometry run directory + `<run>.claim` absent;
check existing ancestors for reparse points. Trusted workspace assumption unchanged.

Observer path:<artifact directory>/email-geometry-observations/<run>/manifest.json and
samples.jsonl. App path:<artifact directory>/email-geometry-records/<run>/geometry.jsonl.
Observer never creates/deletes the app run directory/claim or rewrites app records.
Manifest exact fields:SchemaVersion=1,Observation=EmailGeometry,RunId,Language=en-US,
Profile=rtl-mixed,ExecutableRelativePath,Arguments,ExeSha256,DllSha256,
GeometryRelativePath (relative to artifact directory),ObservationCapMs=360000,
MaxSleepMs=2000. No exception messages/user text or arbitrary metadata.

On eventual approved run, keep original Process/SafeHandle and exact CIM command line,
shared monotonic deadline/paired evidence/cleanup semantics. No forced termination,
retry or UI interaction. Entry result wraps shared result as ProcessObservation plus
GeometryEvidenceStatus=Unreviewed,Acceptance=NotAccepted,RunId,GeometryRelativePath.
Failure emits sanitized category and unsuccessful exit; never announces diagnosis.

### Task1 — Fixed entry and synthetic contract tests

- [ ] Write default guard/policy tests; run `pwsh -NoProfile -File clients/windows/tests/observe-email-geometry.tests.ps1`. Expected RED missing entry/function, not external-effect execution.
- [ ] Implement fixed guard/policy. Test literal run11111111111141118111111111111111; exact hashes/args/path/source list; reject uppercase/zero/traversal/malformed run ID and public overrides. Expected GREEN, no imports/launch without switch.
- [ ] Add RED gate/writer/backend tests before implementation: each token/branch/dirty/source/missing/hash/CIM/existing-target failure blocks all resources; context mismatches and both destination/claim/reparse collisions block launch. Writer construction/open/write/flush/dispose injections preserve ownership. Then implement minimal callbacks and rerun GREEN.
- [ ] Test actual new backend with unchanged shared orchestrator in Synthetic mode using fake original Process/handle and synthetic native/managed readers. Assert one launch callback, native-first ordering, correlation, normal-exit0 finalization, failure cleanup and no fallback; all actual OS effects remain forbidden. Add regression proving finalized exit still yields Unreviewed/NotAccepted.
- [ ] Run new synthetic suite and six existing scripts separately:observe-c4,observe-c5,c5-observation-session,observation-session-guard,process-observation,native-process-observation. Expected old152/32/120/48/32/78 counts, all exit0; record actual new count. No App compilation/live checks.
- [ ] `git diff --check`; verify old/shared files unchanged and new/old artifact hashes still match. Commit explicit new files with `feat: bind isolated email geometry observation entry`.

### Task2 — Review and durable handoff

- [ ] Document readiness/limitations/operator instructions and both receipt locations. No ready-to-run claim or live approval implied; geometry parsing/visual correlation remains a separately scoped post-run review.
- [ ] One fresh independent read-only review of Task1, using all five Review Focus items. Fix Important/Critical in one RED→GREEN pass then rerun new + existing synthetic suites. Record deferred Minor and all declined-to-judge rulings.
- [ ] Update ledger/status/recovery/plan with exact evidence, no-live boundary and remaining steps. Commit/push necessary files, verify remote SHA=local and clean tree. No PR/merge.

## Self-review and handoff

Coverage: approved fixed-entry continuation, shared-loop reuse and verified hashes
are explicit; C4/C5 and app contract remain unchanged. Correlation and geometry
acceptance are deliberately separate. Source tests have no live fallback. Main
uncertainty is later real execution/receipt quality, not permission to launch.
Execution used sequential implementation and one independent review. Do not launch:
resolve the remaining Important failure-category issue before claiming entry readiness.
