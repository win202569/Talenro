# Offline Record Isolation Repair Implementation Plan — pending approval

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after explicit approval. No agents, independent review, default setup/suite reruns or cleanup are included. Steps use checkbox syntax for tracking.

**Goal:** Correct the confirmed test-module export boundary and prove that a new, independent capture publisher remains usable while case-local factories and collectors change.

**Architecture:** Preserve every frozen implementation/test and build a small, separate recording-harness leaf, not another full runtime tree. Its case-module builder initially retains the observed export defect; the sole RED→GREEN edit suppresses exports. Its capture publisher has a separate module/session state and never resolves the tested writer/factory/collector names.

**Tech Stack:** Windows PowerShell orchestration; pinned PowerShell 7.6.5/Core/x64; existing .NET FileStream/Process; no C# compilation or external dependencies.

**Spec:** Preserve the [approved offline design](../specs/2026-10-19-offline-diagnostic-entry-migration-design.md) and the stop/first-cause constraints of the [I1/I2 repair plan](2026-10-31-offline-diagnostic-review-repair.md). The [actual binding diagnosis](../../roadmap/2026-10-31-offline-record-chain-diagnosis.md) supplies the confirmed defect for this narrower adjustment. The local design below is a proposal, not an already approved architecture change.

Status: PLANNING_ONLY_PENDING_EXPLICIT_APPROVAL. Baseline `a003dfe128bae5f620b8e89832446da25e1004cd`, existing `codex/c12-b01-task9-coordinator` worktree. Current user consent is to write this plan only. No new runtime file, module import, Prepare, test, compilation, process or review is authorized by this planning checkpoint.

## Global Constraints

- “只涉及离线诊断，不做生产接入”; do not change product code, public entries, CI, Docker, verification scripts, security settings or installations; no main merge.
- “Session/Ack 不序列化，不从文件加载，不跨 shell/进程转移”; this narrow harness creates neither Session nor Ack and never calls RunApproved.
- “首次实际失败后，后续逻辑 Read/Parse/Save/Complete/Publish/重新打开在文件访问前拒绝”; no failed publisher retry, replacement output or permission recovered from a file.
- “不加入 PID/名称重查、杀树、Job/watchdog、无限等待、启动预热或额外 readiness 等待”.
- Preserve the original2317pins/25seed sources, all old raw/evidence/snapshots/reports/plans/portable archives, and all scope-diagnostic sources/15packet files. Do not edit or rerun either historical driver.
- Old repair remainder6hosts/7targets/1compile/conditional1review stays FROZEN. The one-host scope-diagnostic allowance is exhausted. Any allowance below is NEW and requires approval; none is refunded or transferred.
- I1/I2 remain unresolved; original writer-red's first ErrorRecord remains UNKNOWN. This plan does not reconstruct it or claim all its symptoms have a single known cause.

## Review Focus

1. Same-name exports must not alter any of the four outer function references, even while the case function is active: Task1 groups1–4 and6.
2. Module collector state must be explicit and separate, not accidentally null in an unobserved scope: Task1 groups5 and7.
3. A factory replacement must not also replace the normal evidence publisher: Task1 group8 plus actual RED/GREEN final publication.
4. Failed publication must latch before another logical operation, without successful-looking replacement evidence: Task1 group9 and parent independent inspection of the declared case artifact.
5. A complete-looking child result must not bypass unknown exit/drain/source/record states: Task1 exact parent contract and Task2 preservation/stop boundaries.

## 1. Local design, alternatives and limits

Confirmed: four commands were shadowed during diagnostic caller-function execution; module and outer script state were separate. Not confirmed: the original first exception, post-function-unwind binding lifetime, compound-error assertions or complete I1/I2 behavior.

Proposed approach A (recommended): a small harness regression for the observed export boundary, with independently bound publisher and explicit state. It exercises a real existing-file CreateNew error without a compiled fault fixture. Cost: it proves only this harness seam; later actual I1/I2 integration still needs its own approved revision.

Alternative B: immediately copy all25sources into another runtime generation and rerun seven grants. This mixes the confirmed recording defect with untested error assertions and substantially enlarges the budget; not included. Alternative C: edit frozen tests/capture in place or merely suppress parent failure. Both destroy provenance or conceal unknown evidence; excluded.

No production writer/validator edit is part of A. Even full success leaves old runtime acceptance blocked. The new publisher is instrumentation, not a substitute implementation used to declare the reviewed production capture boundary fixed.

### File and API map (future files only)

New leaf: `docs/diagnostics/isolated-fuzz/p2b-record-isolation-repair/`.
New fixed packet: `.superpowers/sdd/2026-10-31-offline-record-isolation-repair/`.

| File | Responsibility / interface |
| --- | --- |
| `.gitattributes` | `* text eol=lf` for stable source/portable checkout bytes |
| `README.md` | Scope, fixed phase commands, no replay/acceptance claims |
| `expected-output.json` | Independent literal ordered nine-group facts, complete observation/error schemas and both exact outputs |
| `case-module.ps1` | Dot-sourced in the active test function; parameters WriterPath/CollectorText are fixed internal verified sources; assigns `$riCaseModule` using the original New-Module construction |
| `publisher.psm1` | Private functions New-RIPublisher(Grant,Case), Write-RIPublisherBytes(Publisher,Bytes); export no members, invoked only through its retained module reference |
| `isolation.tests.ps1` | Invoke-RIIsolationTests; nine fixed assertions/actual observations, no C#/fixture/Process/Session |
| `capture.ps1` | Fixed `-Grant isolation-red|isolation-green -FrozenSHA256 <actual>`; actual tests and independently bound evidence publication |
| `driver.ps1` | `-Action Prepare|RunOne -Grant seed|isolation-red|isolation-green`; Prepare green also requires actual previous manifest SHA; exactly one owned host per RunOne |

Only `case-module.ps1` may change between phases. Seed loads the frozen writer and extracted Capture-Failure definition exactly as before, without limiting exports. GREEN adds a single no-members export declaration inside that module construction; no other algorithm change. Microsoft's [Export-ModuleMember documentation](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/export-modulemember?view=powershell-7.6) documents the no-members form `Export-ModuleMember` with no arguments. Freeze all tests, expected values, capture, publisher and driver before RED.

Dot-source the builder in the test function, rather than calling a wrapper that returns after module creation: the four identities must be measured while the import's caller scope is still active, matching the actual diagnosis. Record identities immediately after creation and after inner replacement, not only after a case returns. GREEN must also have zero exported functions/aliases/variables/cmdlets; no Force/Remove-Module or global restore masking a leak.

Inside the case module initialize its HostPrimary=null, HostSecondary=empty ErrorRecord list, HostOverflow=false before operations. A unique script sentinel must be module-only. Invoke only the immutable default writer/factory and extracted collector; no runtime/profile module. Retain the original inner factory reference, replace only the module factory with a counting delegate to that reference, and call the actual capture writer against a precreated fixed file. Catch the real ErrorRecord in the case, register it once, then return raw references for in-process comparison. Never construct errors or assign fabricated expected collector results.

The publisher module uses only direct .NET File.Open(CreateNew,Write,None)/Write/Flush/Dispose, local/module-private state and distinct RI names. It exports nothing and never calls New-ODRecordStream, either Write-OD* function or Capture-Failure. New-RIPublisher accepts only the fixed Grant enum and cases HostEvidence/Healthy/ExistingFailure; paths derive internally from the fixed packet. No public arbitrary path/factory/callback interface.

Each publisher reference is one-shot: Active→Completed or Failed; mark consumed before opening. Return `{Ok,Rejected,Error}` with the actual ErrorRecord on first failure; any later operation returns false/true/null before IO. Track open/write/flush/dispose attempts; hold at most one stream, dispose once, retain original first error and bound cleanup errors to8. A healthy write has open/write/flush/dispose1; existing-file failure has open1 and others0, with zero IO on the second rejected logical call. Record/schema serialization is bounded to1MiB before opening. Unexpected serialization/write/flush/close failure freezes execution; no replacement evidence.

HostEvidence and the two test publishers are separate fixed roles declared before any case starts. A failed ExistingFailure case is a deliberate negative vector, not failure of the host evidence sink. The host sink must still publish the actual case failure and all assertions. Never choose another sink after HostEvidence fails. A separate retained direct journal stream may record actual host errors; it is declared/opened before tests, not a retry/fallback evidence file. If journal or final publication fails, preserve partial bytes and return UNKNOWN.

## 2. Proposed finite verification contract

NEW request: **2 capture hosts / 0 target children / 0 compilations / 0 reviews**. Native author execution only. No setup/baseline/full-suite/task-done reruns, supplementary probes, diagnosis rerun, extra hosts, agents or review. This explicitly limits verification: the new independent publisher receives positive/negative instrumentation checks in both phases, not its own extra missing-function RED cycle or independent security review.

Nine groups in this exact order:

| # / group | Required GREEN fact | Seed expected |
| --- | --- | --- |
| 1 OuterFactorySame | Actual outer New-ODRecordStream ScriptBlock identical to pre-module reference | FAIL |
| 2 OuterCaptureWriterSame | Same for Write-ODCaptureRecord | FAIL |
| 3 OuterIntentWriterSame | Same for Write-ODProfileIntentRecord | FAIL |
| 4 OuterCollectorSame | Same for Capture-Failure | FAIL |
| 5 InnerStatePrivate | Module sentinel present, outer absent; outer primary=null/secondary0/overflowfalse unchanged | PASS |
| 6 FactoryReplacementPrivate | Inner factory reference changes and counter1; outer factory still original after replacement | FAIL |
| 7 RealCreateNewErrorRetained | Actual thrown capture-writer error and module HostPrimary are same reference, secondary0; existing file preserved | PASS |
| 8 IndependentPublisherHealthy | Healthy publisher emits exact UTF8 `{"Case":"PublisherHealthy","Value":1}` bytes once despite case factory replacement; publisher module exports0 | PASS |
| 9 FailedPublisherLatched | ExistingFailure first call actual CreateNew error; second call rejected with IO0, no fallback; original file preserved | PASS |

Each phase precreates two distinct fixed case files containing ASCII `PRIOR`: one for the immutable case writer, one for ExistingFailure. Paths are under that phase's fixed root; no user-supplied locations. For groups7/9 bind actual Type=`System.Management.Automation.MethodInvocationException`, ErrorId=`IOException`, and message `Exception calling "Open" with "4" argument(s): "The file '{path}' already exists."` with only the exact corresponding fixed path expanded. Position/Stack bind to the pinned source and valid deterministic call sites frozen before RED. Unknown signatures/errors do not count as correct rejection. Parent may independently compare original artifact bytes as evidence inspection; the failed publisher itself must not reread/reopen them.

Both phases must complete all9groups, empty stderr, no host PrimaryError/secondary/overflow, `NotRun=[]`, complete ordered observations and healthy final evidence. RED stdout is exactly five FAIL/four PASS lines in table order, exit1; GREEN is exactly nine PASS lines, exit0. No extra exception is an expected RED. Missing/unknown data is never accepted as failure reproduction.

Parent parses exact schemas/keys/types and all actual facts, not just group names. It must verify zero exports in GREEN, command identity booleans/module associations, module/outer states, counter tuples, exact bytes, actual error projections and complete ordered groups. Actual reference equality is computed inside the same host before serialization, not inferred from equal printed errors. Dynamic module names/GUIDs are reported as actual identities with cross-record consistency, never copied from a prior run into the oracle.

Each host120000ms; exact-owned exit2000ms; shared stream drain2000ms. Use the original pinned engine/SMA hashes and non-admin identity from section4 of the I1/I2 plan. Preserve exact-object association and bounded cleanup logic from the diagnostic driver as text, never invoke that historical driver. Start attempts consume their allowance; no PID/name lookup or kill tree.

### Frozen inventory

Inherit the scope-diagnostic pins.json2365 entries unchanged. Add exactly29 entries: its15actual packet files; its portable-final1; its execution report1; stopped I1/I2 repair execution report1; this new plan1; new packet's approved plan/spec snapshots2; and8new leaf sources. Total2394Present entries, with consistent duplicate identities only. Validate the original seed SHA and diagnostic portable SHA recorded in the actual report; no reconstruction of missing raw originals.

Prepare seed once snapshots8files; Prepare isolation-green once requires actual seed SHA and exactly1changed source (case-module.ps1), snapshots8files. Total16source snapshots. Do not add the eventual new portable export to runtime inputs. All inherited files and the other7new sources remain byte-identical. RunOne always takes the actual matching manifest SHA; there is no auto-selected generation or new run namespace on failure.

## Task 1 — independent recording harness and export RED→GREEN

**Files:** create the eight listed sources only; after seed change case-module.ps1 only.
**Interfaces:** consumes immutable writer/collector definitions and historical pins as data/source, not historical entry execution; produces the new builder, actual nine-group observations and independent final evidence.

- [ ] Read this plan, original design/repair constraints and actual diagnosis; record explicit plan/2host budget approval in a new ledger. Verify linked branch/clean state, all inherited pins and byte-exact archived sources. No install or historical module import.
- [ ] Write nine assertions and independent expected-output.json first, including exact RED/GREEN facts/schemas above. Implement fixed phase capture/independent publisher and seed builder preserving default exports. Statically check source bindings, parameter sets, LF/AST/JSON and exact inventory before any host; unexpected static failure stops for disposition rather than consuming extra execution.
- [ ] Prepare seed once. Expected2394Present/8snapshots/no changes/no process. Save actual manifest SHA, snapshots, approval and full tool results before interpretation.
- [ ] RunOne isolation-red once with seed SHA. Expected table's five FAIL/four PASS, exit1, healthy independent records/empty stderr/complete exit/drain/all pins. Otherwise stop; do not relabel another error as RED.
- [ ] Add only no-members Export-ModuleMember inside the new builder's module body. No repair to frozen writer, collector, tests, error assertions, publisher, parent validator or oracle.
- [ ] Prepare isolation-green once with previous actual seed SHA. Expected exactly builder hash changed, other2393pins unchanged,8more snapshots. RunOne isolation-green once with its actual SHA. Expected9PASS/exit0/healthy complete evidence/all pins. No third host or second fix attempt.

## Task 2 — immutable preservation and explicit integration boundary

**Files:** new `docs/roadmap/2026-10-31-offline-record-isolation-repair.md`, leaf `portable-final.json`, current-status/repository-recovery updates; no old reports edited.
**Interfaces:** consumes actual manifests/raw/exit/evidence/final/ledger/tool objects, including any failure; produces a resumable unfinished development checkpoint, never a new permission token.

- [ ] Retain complete original tool objects/chunks and actual session IDs before parsing; wait only the returned session, never restart on an early tool yield. Record actual consumed host counts, unknowns and both phases' outcomes. First genuine unexpected failure freezes all remaining stages, without auto-fix/refreeze/retry/new namespace/budget refund.
- [ ] Export all actual new packet bytes/directories, including zero-byte/partial records and16snapshots if both phases occurred; compare bytes/hash/count against originals. Normalize only new portable JSON layout to LF if needed, never its decoded bytes. Do not commit ignored raw/cache/secrets or delete anything.
- [ ] Update status/recovery with narrow results and limitations; explicit files only, verify staged bytes, commit/push existing development branch and verify remote HEAD equality. No main merge or acceptance review. Author static/byte checks are not an independent review.
- [ ] Stop after this harness stage. I1 compound-write/dispose ordering, I2 complete observation contract, actual profile integration and any later independent review require a separately approved revision with a new exact source list and budget. Do not apply this builder/publisher to frozen v2 or resume its6/7/1/review allowance.

## Self-review and approval gate

The confirmed command-shadowing defect maps to groups1–4/6 and the sole builder edit. Module state and actual error retention map to5/7. Independent healthy/failed publication maps to8/9 plus each host's real final evidence; the parent stop contract covers unknowns.29new entries=15+1+1+1+1+2+8;2365+29=2394;8×2=16snapshots; exactly2hosts/0children/0compiles/0reviews. No full runtime tree or production source is created.

Planning-only checks re-read all2365existing diagnostic dependency entries and matched them; diagnostic portable remains `F47946FC985FE6F80B98D20057E38944FC8AE46DEA0F03EFC4BB1E56CD0459DB`. No runtime check or module was executed during drafting. The proposed no-export fix is backed by the observed mechanism and official documentation, but is not yet tested here.

Please review and explicitly approve this narrower Native plan and its NEW2/0/0/0 ceiling before implementation. Approval means only recording-harness RED→GREEN, not I1/I2 completion, old-budget resume, product integration or main merge.
