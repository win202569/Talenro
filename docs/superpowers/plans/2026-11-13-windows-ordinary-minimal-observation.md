# Ordinary Minimal Observation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task in this session. Steps use checkbox syntax. Implementation requires confirmation of this checklist; live execution requires separate approval.

**Goal:** Prepare a fixed-hash, no-argument ordinary candidate observer proven only with synthetic tests before requesting a live run.

**Architecture:** Adapt the small observe-c4.ps1 wrapper into a separate ordinary entry. Reuse all five process/session dependencies unchanged; exercise the wrapper with AST-extracted functions, mocked external effects and memory writers, as the existing C4 tests do.

**Tech Stack:** Windows PowerShell 7+, Git, existing PowerShell synthetic assertion harness; no new dependencies.

**Spec:** [Approved design](../specs/2026-11-13-windows-ordinary-minimal-observation-design.md).

Status2026-11-13: checklist confirmed, implementation started. Expected missing-entry RED exit1; first GREEN attempt failed on test mock Force binding, stopped without retry. Task1 coverage incomplete; Task2 not started. See [execution record](../../roadmap/2026-11-13-windows-ordinary-minimal-observation.md). No restore/build/launch.

Continuation: separately approved Force correction and one focused run passed564 assertions; five planned regressions passed152/120/48/32 cases/78. Independent review complete:0 Critical,2 Important verification gaps,4 deferred Minors (execution record). Task1 D coverage remains open (real CreateNew behavior and combined wrapper errors). Earlier stop remains historical evidence, not current GREEN status. No further run or live allowance inferred.

## Global Constraints

- Branch codex/c12-b01-task9-coordinator; Windows PowerShell only. No product/shared-module/original-entry edits, restore, build, install, security changes or live invocation in implementation scope.
- Fixed candidate directory and all six exact SHA256 constants come from the spec table. Copy them literally into implementation AND independent test expectations; never derive expected test hashes from the policy under test.
- Build source5ee8a339b574171fc61f3cf0c0c636bd86421397; no tracked App/Core/clients/windows/Directory.Build.props difference from that commit. Missing source object is a stop, not permission to use HEAD. Observer entry and five dependencies match HEAD blobs before any live acquisition.
- Only public RunApproved switch. Default returns one NotStarted object before imports, probes, writes or acquisition. No public executable/hash/arguments/deadline/backend overrides.
- Arguments empty; Start-Process must OMIT ArgumentList. ExpectedCommandLine is quoted absolute EXE only, no trailing space. Strict identity unchanged, including on hosts with different serialization.
- Ordinary token, correct branch, clean worktree, successful zero-target/prototype query, fixed artifacts and ancestor reparse checks remain mandatory. No fallback on failure.
- Fixed receipt root ordinary-account-observations under candidate; fresh nonzero32hex RunId directory; CreateNew manifest and samples. No overwritten old receipts or invented geometry records.
- Future live budget: one separately approved visible launch, zero retries,360000ms cap, max sleep2000ms; normal close by300s. Never force-kill, dump or change services. Timeout is not verified exit.
- This plan does not authorize that live budget. No run command with RunApproved is issued during preparation or synthetic implementation.
- Keep NU1905, prior valid-build-guard coverage Minor, C4/security/Release and unobserved UI acceptance unresolved. Synthetic success proves contracts, not displayed layout or actual command-line serialization.

## Review Focus

1. Paths with spaces and serialized no-argument command lines: exact quoted identity, no empty ArgumentList workaround (Task1 B).
2. Source drift or missing history despite matching EXE: fail closed before receipt/process acquisition (Task1 C).
3. Changed Core/XBF/config or nested reparse ancestors: reject every bound file and ancestor, not only EXE/DLL (Task1 C).
4. Receipt collision or partial writer/disposal failure: no overwrite, retain primary plus cleanup failure, release owned resources (Task1 D).
5. Identity failure/deadline after acquisition: retain original handle, no retry/termination or false verified exit (Task1 E).

## Task1: Implement and prove the isolated ordinary entry

**Create:** clients/windows/tests/observe-ordinary-account-center.ps1 and clients/windows/tests/observe-ordinary-account-center.tests.ps1.

**Read only:** observe-c4.ps1/.tests.ps1, c5-observation-session.psm1, observation-session-guard.psm1, process-observation.psm1, native-process-observation.psm1, native-process-observation.cs (all under clients/windows/tests).

**Interfaces:** Follow C4 signatures with prefix Ordinary: Get-OrdinaryPolicy([string]$Root) returns fixed policy; Invoke-OrdinaryGit($Root,$Arguments) returns checked Git output; Assert-OrdinaryOrdinaryToken() and Assert-OrdinaryPreflight($Policy) throw on failure; Assert-OrdinaryReceiptDestination([string]$Destination) checks collision/ancestors; New-OrdinaryReceiptDirectory($Destination), Open-OrdinaryNewStream($Path), New-OrdinaryTextWriter($Stream), New-OrdinaryReceiptWriter($Policy,$Context) preserve C4 ownership/cleanup; New-OrdinaryBackend($Policy) returns the exact ten callback keys consumed by Invoke-TalenroC5Observation. Add Assert-OrdinaryArtifactPaths($Policy) for leaf/ancestor reparse checks. No exported/general-purpose module.

Policy contains Root, Executable, Arguments as empty string, ExpectedCommandLine, BuildSourceCommit, ReceiptRoot, Sources (new entry plus unchanged five dependencies), and ArtifactHashes (six relative filenames to literal SHA256). Manifest SchemaVersion1, Observation=OrdinaryAccountCenter, Arguments='', ArtifactHashes, BuildSourceCommit, Language='default-environment', Profile='default-environment', ObservationCapMs=360000 and MaxSleepMs=2000, plus baseline correlation/path fields. Context remains exactly RunId,ExpectedPath,ExpectedCommandLine; never add manifest properties to strict Context.

- [ ] A. Write the focused harness first: shadow external effects before any entry execution; invoke entry WITHOUT RunApproved only; AST-extract function declarations for tests. Assertions: result.Count=1, Status='NotStarted', zero import/probe/write/acquisition counts, native helper type absent. Reject Executable,Hash,Arguments,Deadline,Backend parameter overrides. Missing entry must yield an explicit RED assertion, not a tool/environment failure.
- [ ] Run the new test script in a fresh existing PowerShell7 host (command below); record expected RED, exit1. Do not invoke the approved branch of the script or load native helpers in RED/GREEN.
- [ ] B. Add policy/backend assertions: independent exact six constants and fixed base, source commit and six observer source names; Arguments.Length=0; ExpectedCommandLine equals ('"'+policy.Executable+'"') including a fixture root with spaces. Mock Start-Process and assert NOT $PSBoundParameters.ContainsKey('ArgumentList'), exact FilePath/WorkingDirectory/PassThru, clock started before call, one acquisition, original object/handle passed onward. AST additionally rejects a real ArgumentList parameter at the launch site.
- [ ] C. Add table-driven preflight tests: ordinary-token failure, wrong branch, tracked/untracked dirty tree, each observer blob mismatch, failed Git command, missing build commit, product diff in each of three bound paths, missing/mismatched EACH of six artifacts, query exception/nonzero inventory, reparse leaf and every existing ancestor. Each failure asserts writerAcquisitions=0 and processAcquisitions=0. Mock a clean documentation/new-entry HEAD change with unchanged build-source paths and assert preflight succeeds. Stub Git/path/hash/CIM effects; no real candidate required.
- [ ] D. Add writer tests using memory streams: nonzero32hex RunId accepted; malformed/zero RunId rejected; existing destination file/directory and reparse ancestor rejected before writes. Assert CreateNew for manifest/samples, UTF8 no BOM, six manifest hashes/source/empty arguments/default-environment markers, matching RunId, no geometry path. Inject each directory/open/write/flush/dispose failure; assert no process acquired when writer acquisition fails, all obtained resources disposed, primary and cleanup errors retained rather than hidden.
- [ ] E. Add integrated synthetic tests using the real new factory/backend and unchanged shared orchestrator with external effects replaced: exact ten callbacks; missing callback rejected without fallback; exact quoted command line accepted and trailing-space/extra-argument forms rejected; native/managed/CIM identity disagreement fails; zero/multiple/missing/unspecified-date CIM rejected. Fake time exercises360000ms deadline and sleep<=2000 without wall-clock waiting. Assert acquisitions<=1, no kill/dump/service calls, original handle only, timeout ExitVerified=false, success returned Finalized=true versus persisted pre-cleanup Finalized=false, sanitized failure receipts and disposed owned resources. Do not replace finalization distinctions with a new acceptance flag.
- [ ] Implement only the new entry to satisfy A–E; reuse baseline code locally without changing shared files. Add checked Git object existence and path-scoped diff before acquisition; retain source/blob and strict identity checks. Insert default guard before all effects, and omit ArgumentList at launch. Never call live code to debug these tests.
- [ ] Run focused GREEN once; expected exit0 and PASS assertion count, native helper absent, no real target/CIM/native/sleep. On unexpected failure stop and report; do not silently repeat tests or widen scope. Any correction and rerun requires an explicitly recorded bounded continuation.

Commands for future authorized execution only, from worktree root; $PSHOME is the current confirmed PowerShell7 installation (verify version first; missing host is a stop, no install). Each test invocation is a fresh process:

```powershell
$PSVersionTable.PSVersion
& (Join-Path $PSHOME 'pwsh.exe') -NoProfile -File '.\clients\windows\tests\observe-ordinary-account-center.tests.ps1'
"Synthetic exit code: $LASTEXITCODE"
```

- [ ] Review explicit diff, git diff --check; update status/recovery and add a sanitized execution record at docs/roadmap/2026-11-13-windows-ordinary-minimal-observation.md with RED/GREEN counts and limitations. Stage only these new source/test/docs plus checklist progress; checkpoint without claiming live readiness if integration review remains outstanding. No raw receipts/binaries/cache.

## Task2: Synthetic regression, scoped review and no-launch handoff

**Files:** no additional implementation files; update this checklist, the Task1 execution record, docs/roadmap/current-status.md and docs/runbooks/repository-recovery.md.

**Consumes:** Task1 new entry/backend and test; fixed spec identity. **Produces:** sanitized command/exit/assertion results, scoped review disposition, clean checkpoint and separate future-run approval boundary.

- [ ] Inspect each suite for synthetic-only behavior before execution. Run each ONCE in a fresh PowerShell7 process using the Task1 command pattern: observe-c4.tests.ps1, c5-observation-session.tests.ps1, observation-session-guard.tests.ps1, process-observation.tests.ps1, native-process-observation.tests.ps1. Expected each exit0/PASS; stop at first unexpected failure. Exclude loader/compile suites and all App/dotnet/build/live tests.
- [ ] Independently review the new entry/test diff against C4 and the approved spec, concentrating on Review Focus1–5, complete six-artifact/source coverage, no-argument launch, cleanup ownership and no hidden live path in tests. Preserve inherited native/session guards unchanged. Review is static; do not authorize a run or broaden into full project acceptance. Resolve Critical/Important findings before any live proposal; report deferred findings explicitly.
- [ ] Check git diff --check and confirm product/shared-module/old-entry diffs absent. Record suite counts, mocked-versus-real evidence and exact reviewed commit. Explicitly stage necessary files, commit/push the existing development branch, and verify remote SHA equals local and worktree clean. Do not merge main.
- [ ] Hand off a future one-run request only: fixed six hashes, default-language qualitative route from spec, continuous user availability<=6min, close by300s, no retry. Mark exact root/DPI viewports, reserved-option rejection, broader functional/C4/security/Release acceptance UNVERIFIED. Do not issue or execute a launch command until separate approval.

## Recovery and self-review

Fresh clone restores this plan and future entry/tests, not ignored artifacts/receipts or cached tooling. Synthetic fixtures must work without candidate binaries. Missing fixed candidate or build commit blocks future live execution; never rebuild/substitute to satisfy the entry without new authority.

Preparation self-review: approved design mapped to Task1 A–E and Task2 gates; fixed constants remain in linked spec; Context/manifest separation and exact ten callbacks preserved; all five Review Focus cases assigned tests. No live action is included in either task. Checklist awaits confirmation; all implementation boxes intentionally unchecked.
