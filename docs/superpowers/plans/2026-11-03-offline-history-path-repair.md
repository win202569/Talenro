# Historical Path Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix historical-file resolution in the real driver without reopening the stopped runtime.
**Architecture:** Extract a single production verifier, replace the faulty gate loop with its call, and exercise its real hashing behavior in a finite RED/GREEN pair. Preserve all frozen identities and failed records.
**Tech Stack:** Windows PowerShell7.6.5, existing SHA256/AST helpers, CreateNew JSON receipts.
**Spec:** ../specs/2026-11-03-offline-history-path-repair-design.md

Status: PENDING_EXECUTION_APPROVAL. User approved design0700546b with “是”.
Original design bytes remain unchanged; approval of this plan is still required.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; existing branch codex/c12-b01-task9-coordinator.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2; P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner; Q=.superpowers/sdd/2026-11-01-offline-controller-recovery.
- Only operational edit L/driver.ps1. Other30source files unchanged.
- Only two added auxiliary computations RED20/GREEN21, ceiling21. No Prepare, host, snapshots, children, compilation, reviewer, install or cleanup.
- Actual counters start19aux/4hosts/3compiles/186snapshots and may end21/4/3/186.
- All ten old stops preserved. Old runtime routes remain blocked even after GREEN.
- First unexpected result, drift or recording error stops; no patch/retry after it.
- No old test replay, security relaxation, main merge or product acceptance.

## Review Focus

- Absolute paths on another drive must not be prefixed with repo: groups1–4.
- Drive-relative and root-relative forms are not fully qualified paths: groups9–10.
- Canonical aliases can hide conflicting hashes: groups13–14.
- External allowed identity must bind both path and hash: groups7–8.
- A correct helper unused by the real gate is insufficient: group17 plus actual extracted gate-call execution.

## Files and interface

Create L/history-path.tests.ps1 (definitions only), Q/prepare-history-path.ps1,
Q/run-history-path.ps1, Q/preserve-history-path.ps1, Q/history-path-progress.md.
Create P/history-path.approval.json, P/history-path.input.json,
P/history-path-candidate.input.json; phase intent/final/tool-record names are
history-path-red.* and history-path-green.*; stop is history-path.stop.json.
Create L/portable-history-path-{preparation,candidate,final}.json and
docs/roadmap/2026-11-03-offline-history-path-repair.md. Update status/recovery.

Production function:
Driver-VerifyHistory([string]$Repo,[Collections.IDictionary]$HistoricalFiles) -> void.
It must emit nothing on success and use unchanged Driver-CheckHash/Driver-Hash
for actual bytes. No caller-provided external allowlist. It embeds exactly:

- C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe:
  362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139
- C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe:
  60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F
- C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/System.Management.Automation.dll:
  2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B

Require Repo fully qualified and a nonempty map with string keys/hashes. Reject
blank keys, trailing spaces/dots in components, colon except absolute drive
prefix, wildcard/NUL characters and hashes not uppercase64hex. Reject UNC/device
prefixes before any file access. A rooted path must be fully qualified and resolve
to one of the three canonical external identities with its exact expected hash.
Drive-relative C:foo and rooted-relative \foo or /foo reject. Repository-relative
keys resolve with GetFullPath against explicit Repo, then must start with canonical
Repo plus directory separator (case-insensitive); root itself and escapes reject.
Internal dot segments may normalize while remaining inside Repo.

Canonical external separator/case aliases may match the same fixed identity.
Canonical duplicates use OrdinalIgnoreCase: reject different expected hashes
before hashing either alias; identical expected hashes are allowed. Resolve and
validate the entire map first, then hash each unique file once. Do not follow an
unrecognized path simply because its bytes match an allowed binary. This does not
claim protection against concurrent filesystem/reparse changes or enable relocation.

Validation errors use RuntimeException message/ErrorId
'OD: historical path binding mismatch'. Actual missing-file/hash errors retain
existing helper behavior, not converted into successful validation.

Replace exactly the faulty foreach in Driver-ExtentContractGate with:
Driver-VerifyHistory $repo $b.HistoricalFiles
No other gate, process, selector, allowance or receipt behavior changes.

## Task 1: Prepare immutable regression and obtain RED20

**Consumes:** saved extent Prepare error, original driver and2307history bindings.
**Produces:** immutable test/input, actual failing RED20, no source changes.

- [ ] Record explicit execution approval, design/plan hashes, baseline commit,
  counters and exactly2added auxiliary allowances in new approval. New files must
  not already exist. Verify original driver SHA
  A5A09F0C8D36DB0CBD56747C90E3D2DB8741A7018EF1B7FA7F21E775E94ECE81,
  extent input SHA50A088C0BFE6E931BB5F1DB3194399DEF9F1AD15CC7C949EF4091EF2A93737E2,
  stop SHAC26E23A5783CC4CC6062F9142000B2D68931D1C990AE30C0E5986800B222029F,
  and all ten stops. Keep source identities separate from immutable history;
  if history includes a live driver identity, preserve its bytes in the input
  archive rather than requiring edited live bytes to equal that historical hash.
- [ ] Freeze31baseline source hashes, all2307original history bindings, saved
  stop/tool record, test/helper hashes and exact extracted source spans in input.
  Read the three external expected hashes from the immutable history and compare
  to the literals above. Do not run any production function during preparation.
- [ ] Write Invoke-ODHistoryPathRegression -Phase Red|Green. Extraction allowlist:
  unchanged Driver-Hash and Driver-CheckHash; new Driver-VerifyHistory only in
  GREEN. Extract the unique HistoricalFiles foreach AST directly from original
  Driver-ExtentContractGate in RED; preserve original line offsets when compiling
  extracted definitions/statement. No top-level driver evaluation.
- [ ] Freeze18ordered groups:
  SavedExternalReadable,AllOriginalHistory,OtherExternalIdentities,
  RepositoryRelative,WrongHashRejected,MissingFileRejected,
  ExternalHashRejected,UnknownAbsoluteRejected,DriveRelativeRejected,
  RootRelativeRejected,NetworkDeviceRejected,TraversalRejected,
  ConflictingAliasesRejected,IdenticalAliasesAccepted,MalformedInputsRejected,
  SourceWindow,ActualGateWiring,OriginalGuardsUnchanged.
- [ ] RED group1 runs the actual extracted foreach with repo fixed and
  b.HistoricalFiles containing only the saved pwsh key/hash. It expects valid
  history to be readable; current production loop fails. Do not copy/reimplement
  Join-Path in the test or substitute a missing-function failure.
- [ ] Expected RED is0PASS/1FAIL/17NotRun/exit1; null SetupError/RecordingError.
  Require ErrorId exactly FileReadError,Microsoft.PowerShell.Commands.GetFileHashCommand,
  exception message equal to saved stop.Message, and stack including actual
  Driver-Hash and Driver-CheckHash at preserved original lines17/18. Record actual
  exception type, Position and TargetObject without inventing an unsaved type.
  Comparison to saved diagnostic text is valid only in the unchanged environment;
  localization differences are an unexpected result, not a relaxed pass.
- [ ] GREEN uses real production verifier and these fixed fixtures, all in memory:
  group1 same pwsh key; group2 complete2307entry map; group3 other two absolute
  identities; group4 relative L/expected-output.json with its original hash.
  group5 changes that relative hash to64zeros, requires existing hash mismatch.
  group6 P/history-path-absent-fixture.bin with64zeros: preparation must confirm
  absent, never create it; require actual missing-file error.
  group7 allowed pwsh with64zeros; group8 C:/history-path-unapproved.exe with
  valid-format hash: require binding rejection before lookup.
  groups9–12 respectively C:relative.txt, \relative.txt and /relative.txt,
  \\server\share\file plus \\?\C:\file and \\.\device, and ../outside-history.txt.
  Require binding error, not incidental I/O error.
  group13 relative expected-output path and its ./ alias with differing hashes;
  group14 same aliases/same hash: rejection versus silent success.
  group15 null/empty map, blank key, trailing-dot/space component, wildcard,
  invalid colon, invalid-length/lowercase/nonhex hash; all reject.
- [ ] Group16 compares31sources to candidate/original: only driver differs.
  Group17 extracts the actual new gate call and invokes it with the original
  full history through loaded production helper definitions; assert silent
  success and AST exact argument binding repo,b.HistoricalFiles. Never invoke
  the gate's approval/stop logic. Group18 reverse-removes new function and restores
  the original foreach, requiring original driver bytes/hash and other30files.
- [ ] Implement phase runner with mandatory expected input/test SHA, exclusive
  intent before invocation, immutable pre/post source/history comparisons,
  full actual receipts and first-unexpected stop. Each phase is callable once.
  Mechanical preparation/archive helpers do not invoke tests or start processes.
- [ ] Preserve preparation archive, verify decoded bytes/hashes/lengths/directories/
  zero-byte files and push explicit tests/archive/docs before RED.
- [ ] Run Q/run-history-path.ps1 -Phase Red with frozen SHA arguments once;
  compare actual result to exact RED contract. Any other result ends scope.

## Task 2: Implement verifier, run GREEN21, preserve outcome

**Consumes:** immutable Task1 inputs/test and exact RED20.
**Produces:** one-file repair and actual local result; runtime stays blocked.

- [ ] Implement Driver-VerifyHistory with the fixed interface/algorithm above;
  insert before Driver-ExtentContractGate and replace only its historical loop.
  Retain existing Driver-Hash/Driver-CheckHash bytes, all old stops, all other
  selectors/loops and process tail. Parse candidate statically.
- [ ] Create candidate input binding original input hash, actual RED final hash,
  all31candidate hashes and original-history input hash. Do not rewrite original
  input/test or archived baseline. Preserve ledger and candidate archive.
- [ ] Verify explicit candidate diff; commit/push source and candidate archive,
  confirm remote equality before GREEN. No test calls hidden in this checkpoint.
- [ ] Run Q/run-history-path.ps1 -Phase Green with frozen input/test SHA once.
  Require18PASS/0FAIL/0NotRun/exit0, no root/setup/recording error, exact pre/post
  input identities and ten original stops unchanged. Both complete-history
  checks must execute the production verifier, not fact-flag surrogates.
- [ ] On any unexpected outcome, record new stop and partial records, no retry.
  On success, stop after preservation; do not call Prepare or remove extent stop.
- [ ] Write actual result, consumed20or21auxiliary counters, unchanged4hosts/
  3compiles/186snapshots and frozen allowances. Archive all P/Q records plus
  required historical inputs, verify decoded originals. Keep original archives.
- [ ] Update status/recovery, explicitly stage source/tests/portable archives/docs
  only, check diff and raw staged hashes, commit/push existing branch and verify
  remote SHA equality. No main merge; NOT_ACCEPTED at product level.

## Self-review and handoff

Spec coverage: Task1 covers real failing call and all path/hash classes; Task2
covers minimal production wiring, finite verification and preservation. Review
Focus items map to named behavioral groups. Namespaces do not overlap previous
attempts or host slot. Extraction never invokes stopped driver top level.
This is native execution with no added reviewer under the approved design.
No execution occurred while writing this plan; await plan/execution approval.
