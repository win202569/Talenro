# Isolated writer business-prefix implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Observe the original two-group business prefix once through the verified isolated writer, without changing its assertions or claiming a timeout fix.

**Architecture:** The existing supervisor owns sibling writer and business-producer processes. A new versioned role entry replaces only the synthetic producer with the original prefix capture in that same process; no grandchildren. Reuse both pinned DLLs without compilation.

**Tech Stack:** Windows PowerShell 7.6.5 Core x64, existing C# assemblies, exclusive JSON/JSONL evidence, PowerShell static checks.

**Spec:** `docs/superpowers/specs/2026-11-04-runtime-isolated-writer-design.md`, especially its later business-prefix boundary. The user approved preparing this extension and plan on 2026-11-05; this document awaits execution approval. Native inline execution remains the finite-scope method; no separate agents or broad test runs.

## Global Constraints

- Current counters: computations/process starts/compilations/snapshots = 41/28/6/217. Proposed single computation 42, at most two new starts, zero compilation and zero snapshots; ceiling 42/30/6/217. Unused starts never become retries.
- Planning changes only now. Do not create an attempt claim, run preparation scripts, load assemblies, or launch a business process before plan approval.
- Preserve original guards/wrapper/cutoff and 62-event structure; actual48/0-byte output and all original success conditions are required.
- Execution120000ms, readiness5000ms from the same origin, one shared2000ms normal drain/finalization and one shared2000ms failure exit-confirmation window. These are programmed wait budgets, not OS wall-clock guarantees.
- Original owned Process references only; no PID adoption/tree kill, no new Job/PInvoke/service/security setting, no disposal of pending resources.
- Failure output unavailable with null length/hash; do not inspect or archive failed data files. Retain original sources, records and claims byte-for-byte.
- Five capability flags remain false; NOT_ACCEPTED. No full suite, beyond-prefix execution, original-timeout-fix claim or main merge.
- Missing/mismatched binaries stop. New-machine compilation requires separate approval; Git contains no DLL/runtime/cache.

## File map

All new source names below are under `docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/` (L). Evidence directory D is `.superpowers/sdd/2026-11-05-runtime-isolated-writer-business-prefix-v1`.

- Create L/runtime-isolated-writer-business-capture.ps1: versioned original capture; binding changes only.
- Create L/runtime-isolated-writer-business-common.ps1: versioned initialization/result helpers; computation42 and two-start intent binding.
- Create L/runtime-isolated-writer-business-role.ps1: Writer/Producer dispatch, identity checks and exclusive role claims.
- Create L/runtime-isolated-writer-business-driver.ps1: one invocation of existing Supervisor.RunAsync and terminal evidence.
- Create L/runtime-isolated-writer-business-static.tests.ps1: definition/AST-only checks; never invokes role, driver or capture.
- Create D/prepare.ps1, D/verify.ps1, D/preserve.ps1 and progress.md: preparation, read-only post-run verdict and portable evidence. Portable preparation/final JSON files live under L with prefix portable-runtime-isolated-writer-business-prefix-.
- Update docs/roadmap/current-status.md, docs/roadmap/2026-11-04-runtime-isolated-writer-result.md and docs/runbooks/repository-recovery.md at execution checkpoints.

## Review Focus

1. Old computation29 paths/intent accidentally reused: static exact substitution checks and fresh-directory/claim checks must reject them.
2. Synthetic PASS bytes masquerade as business success: require original group/schema results, exact stage structure and cutoff, not just stdout.
3. Capture's exit bypasses wrapper after-checks: capture performs original Complete-HT before exit; never rely on code after its invocation for producer evidence.
4. Valid transport with failed/missing business receipt: overall observation fails even if transport succeeded.
5. Missing DLL or a consumed attempt on another computer: stop before any start; never rebuild or replay implicitly.

## Task 1: Versioned binding and static contract

**Interfaces:** Capture retains `Grant`, `FrozenSHA256`, `Epoch`, `InputSHA256`, `IntentSHA256`. Common retains Initialize-HT and Complete-HT signatures. Role accepts the existing supervisor's `Role`, `Scenario`, `InputSHA256`, `IntentSHA256` arguments; fixed Scenario is `BusinessPrefix42`.

- [ ] Write static tests first. Assert capture differs from runtime-copy-transport-capture.ps1 only in HTD and the common-helper import; all business test/helper imports, guards, try/finally, Complete-HT and exit remain identical. Assert common differs only in the computation/ceiling/start allowance comparison. Compare normalized source with an explicit replacement allowlist, not broad regular expressions.
- [ ] Run that definition-only check in the current shell: `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-isolated-writer-business-static.tests.ps1`. Initial expected result is missing new files; no execution ledger slot or child process is consumed. If tests require runtime behavior, stop for a revised budget rather than treating it as static.
- [ ] Add the two versioned files. Bind D/intent.json to computation42, ceiling42, MaxProcesses2, InputSHA256 and token; preserve all original Complete-HT acceptance checks.
- [ ] Implement the role with the existing scenario role's identity/map/claim checks. Writer loads verified assemblies and calls Writer.RunAsync with normal nonfault Scenario. Producer invokes the new capture in the same process with Grant contract-green, Epoch history-contract-runtime, frozen hash186FE2C2B44A31A98EA5406109AC888A625FD1B0061C5338D0382268015D9BBA. Do not call synthetic Producer.RunAsync or start another process. Producer post-input map comes from Complete-HT.
- [ ] Extend static tests to assert exact role dispatch, no synthetic producer call, no new process/compile site, pinned scenario, and no producer success logic after capture exit. Re-run expecting `BUSINESS_PREFIX_STATIC_VERIFIED`; preserve both RED/GREEN outcomes separately.

## Task 2: Single-run driver and preparation checkpoint

**Interfaces:** Driver takes `-InputSHA256 <64 uppercase hex>` only; constructs existing ScenarioOptions with Scenario,Token,InputSHA256,IntentSHA256,Engine,RoleScript,Root, then calls Supervisor.RunAsync once. Preparation produces input.json, preparation report and portable archive; it does not consume intent/attempt claims.

- [ ] Add static assertions for exactly one supervisor invocation, no loops/retries, no Add-Type compilation, fixed computation42 and MaxStarts2; failed or incomplete supervision cannot reach stable-output checks.
- [ ] Implement driver and preparation using scenario-v1 patterns with explicit absolute worktree and D paths. Reject existing claim/intent/stop artifacts. Write exclusive attempt intent before starting; record actual starts even when producer never starts.
- [ ] Pin all historical inputs/absence paths and new helpers, including prior final scenario archive SHA25671B8CFA5EBFF2519FEA1EE29E9EF9777671E844D75120871BBCDB950DBF98FB5. Reuse v3 isolated-writer.dll SHA2562AEF2A39B5D499981378D8FD1F749A3073AD92C03641CEBBA28E9060B3FFCD7B (80896 bytes) and adapter.dll SHA2564C1470B94B3DEAA87635F1375A05161BE0E1F4C4D2CF65595214A8E39CDBCCDA (50176 bytes). Verify fixed engine/account using existing guards.
- [ ] Parse new scripts and run static checks in the current shell. Prepare once; record exact pins, file counts and hashes. Stop on any unexpected failure, preserve it, and do not silently repair/re-run preparation.
- [ ] Decode/compare the preparation archive to its sources. Commit/push the explicit necessary source/docs/portable archive files to the existing branch and verify remote SHA equality before launch.

## Task 3: One observation and independent read-only verdict

**Interfaces:** D contains intent/claims, supervisor result, bounded control records, input maps and the original child-result/stages/capture journal. Verify consumes these only after terminal supervision; no assembly loading or workload invocation.

- [ ] Implement verify.ps1 before launch. Require transport SUCCESS, both owned exits0, no pending tasks/resources, authenticated COMPLETE, matching hashes/EOF, and stable writer files containing exactly UTF8 `PASS: ContractAPIAvailable\r\nPASS: BaselineCore\r\n` (48 bytes), stderr0. Never use synthetic pattern assertions here.
- [ ] Require child-result HOST_PREFIX_COMPLETED, cutoff true, both exact groups and schema rows passed, no primary/secondary/recording/identity errors, no aborted/overflow/journal failure, zero open intervals, all five flags false. Compare the 62 events/31 paired intervals against computation29's structure, including sequence/event/parent/stage/kind/case, excluding timestamps/frequency; do not compare total byte length because timings vary. Require unchanged source maps, identities and balanced attempt records. Missing producer receipt is failure even if transport succeeded.
- [ ] Invoke driver exactly once with the newly recorded input hash. Preserve actual tool outputs; no retry, synthetic rerun, compile, extra probe or automatic salvage. A failed start consumes the attempt. Unknown exit or pending resource stops all further runtime work.
- [ ] Run read-only verification once against actual evidence, recording actual outcome separately from transport result. On failure retain all records, mark unavailable data excluded and explain which conditions failed. Do not rewrite receipts to manufacture success.
- [ ] Package only stable eligible evidence/source; decode-check path/length/hash against originals, exclude DLLs and failed output data. Update result/status/recovery with counters and exact verified versus unverified scope. Commit/push explicit files and verify remote equality/clean worktree. End the finite gate regardless of outcome.

## Plan self-review and handoff

Reviewed against ownership, framing, timing, failure-output and later business-prefix boundaries of the parent design. No protocol or compiled-code changes are planned. Review Focus1/3/5 are covered by Task1/2 static/preparation checks; Focus2/4 by Task3's conjunctive verdict. All runtime failure paths retain the already tested supervisor behavior, not new injected process tests.

Status: PLAN_READY_FOR_REVIEW; implementation/preparation/runtime not performed. Approval of execution must cover this finite ledger; subsequent unexpected changes or any required extra compile/start need a revised scope. The original timeout cause and full acceptance remain unresolved even if this observation passes.
