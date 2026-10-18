# Token Direct Bounded Validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking. No subagent/reviewer launch is authorized by this proposal.

**Goal:** Test the prepared Token helper against independent literal outcomes, without claiming full acceptance or performance.
**Architecture:** One new bounded controller, separate fresh baseline/candidate PS7 children, baseline-first fail-stop gating. Existing candidate and fixture sources remain frozen.
**Tech Stack:** Windows PowerShell orchestration, ordinary-token PS7 7.6.5, .NET process APIs; no installs/Docker/PS5.
**Spec:** docs/roadmap/2026-10-18-p1-token-local-preparation.md and the bounded design recorded in current-status.md.
**Status:** PROPOSED ONLY. User approved drafting this plan after95360155, not these new target executions.

## Global constraints

- New allowance proposed: at most TWO target-host launch attempts total, ONE baseline then ONE candidate. Not an allowance drawn from historical frozen slots.
- Both slots start UNAPPROVED/UNUSED. Any launch attempt consumes its slot, including startup/compilation/fixture errors.
- No candidate launch unless baseline meets every success criterion. Baseline failure freezes candidate slot; no retries, replacement hosts, warmup, reload, alternate engine or timeout increase.
- No changes to frozen candidate, test, support or source-pins; no oracle relaxation, fixture repair or automatic follow-up run under this proposal.
- Candidate/helper full API, performance, independent review, I1/P1/main/C12/security/runtime/release gates remain unaccepted.
- Controller preparation/static checks and author self-review only are included in the proposed future implementation, not performed by writing this document.
- No elevated target processes, policy changes, installations, main merge or unrelated process termination.

## Files and immutable identities

Read:
- docs/diagnostics/isolated-fuzz/integer-token-local-candidate/token-direct.tests.ps1
  SHA256 04B72510A4FA5D915A3A2E5AF024FA387E92C0FD83A25811AAAB0573FE1A3692
- Same folder source-pins.json: F72A570701A448394FEC5F4C427788C102E9DB712D640CE5EBF8E174D943698B
- Candidate: F9B4E408D9A4FF9C7A36EFD6E1F2472F17DACB3F76F463FF4973DAB4139CC82A
- Baseline: 3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3
- Support: BF82CE30B9CE82449C90ADAA96C0B03F3892609FB903EF87B07C7FD68489409A
- Engine: C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe
  SHA256 362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139

Future create only after execution approval:
- .superpowers/sdd/2026-10-18-p1-token-direct-run/run.ps1 (ignored, new controller; never overwrite existing directory)
- Same folder frozen.json, invocations.json, baseline.stdout.txt, baseline.stderr.txt, candidate.stdout.txt, candidate.stderr.txt (as applicable).
- docs/roadmap/2026-10-18-p1-token-direct-result.md (portable evidence/result summary).
Modify current-status.md and repository-recovery.md at handoff. No test/product source edits.

## Review focus

1. Null success oracle must remain null, not typed-string coerced to empty.
2. PSObject decoration may disappear across field/helper binding: baseline discrepancy is failure, not accepted candidate equivalence.
3. Empty strings/allowlists and seven wrong-type inputs must not be omitted by enumeration/parameter binding.
4. Members/GetType/Length/Rank shadowing and executable descriptors must never invoke callbacks.
5. Timeout, truncated/extra/duplicate output, uncertain exit/drain or altered pins must prevent candidate launch.

### Task 1: Bounded direct check and evidence handoff

**Interfaces:** Consume frozen token-direct.tests.ps1 -Variant baseline|candidate -ExecutionSeparatelyApproved. Produce append-only per-attempt receipt and a portable report with DirectSuitePassed, FormalAcceptance=false. Never learn expected results from target output.

- [ ] Step 1: After approval, construct expected ordered stdout inventory as data from the fixture's fixed loops, not by loading it. Per metadata mode: seven named scalar/allowlist cases, seven wrong-type lines, then five descriptor kinds x five names x two values x four Check-Call labels. 214 PASS lines per mode;428 case lines total plus terminal PASS direct-suite-<variant>,429 nonempty lines. Duplicate labels retain their ordered multiplicities. Do not compare only counts or terminal marker.
- [ ] Step 2: Prepare controller as source. Statically verify PS AST, fixed inventory, launch order, durable slot-before-launch recording, timeout/error cleanup and no broad kill/retry path. Read-only inventory mutations (missing/extra/reordered lines, nonzero exit/stderr, timeout, unknown cleanup) must classify as failure by inspection; no synthetic child-host runs. Self-review only; independent review allowance stays unchanged.
- [ ] Step 3: Freeze controller, inventory, test/support/module/engine hashes, historical434 pins and prior manifest/controller/index/stdout/stderr anchors before first launch. Assert ordinary parent token; fail without launching if elevated, pins changed, engine missing, destination exists, receipts unreadable or static safety checks unresolved.
- [ ] Step 4: Launch baseline ONCE in fresh no-profile noninteractive PS7, ordinary inherited token, working directory this worktree. Command arguments: -NoProfile -NonInteractive -File <absolute token-direct.tests.ps1> -Variant baseline -ExecutionSeparatelyApproved. The existing script alone compiles test-only C# and loads only selected module. Never pre-import, precompile, profile, warm up or intercept production metadata.
- [ ] Step 5: Enforce 30000ms child wall deadline including startup/compile/import. Drain stdout/stderr concurrently. On timeout/error terminate ONLY the exact owned Process object, bounded exit wait2000ms;bounded stream drain2000ms after exit, at most34000ms supervision per host. Failure to prove exit/drain is incomplete and freezes all later launches;no PID/name-wide or arbitrary descendant kill. No existing process touched. Capture CPU as receipt only, not speed evidence.
- [ ] Step 6: Retain original streams and receipt even on failure. Record launch-attempt UTC/PID/engine/source hashes, timeout/exit/elapsed/CPU/cleanup/capture/stream hashes. Baseline success requires exit0, no timeout/controller error, verified owned exit and complete drains, EMPTY stderr, exact429-line ordered literal inventory including terminal baseline label, all pre/post pins intact. Any discrepancy halts here, candidate unused/frozen. No repair/rerun authorized.
- [ ] Step 7: Only after full baseline success, repeat Steps4–6 ONCE in another fresh host with -Variant candidate and terminal candidate label. No baseline/candidate co-loading. Candidate must independently meet the same oracle;agreement with baseline alone is insufficient.
- [ ] Step 8: Save result/progress/recovery, before/after pins and attempt consumption. Missing raw session files on a new computer do not replenish slots. Report direct-only success or precise failure/incomplete cause;do not claim RED/GREEN, full API equivalence, timeout fix or speedup. Stage necessary portable report/docs only, commit/push existing dev branch, verify remote SHA and clean tree. Do not commit ignored raw/controller/cache.

## Approval requested

Approve this complete bounded proposal to authorize controller preparation and up to the two ordered target attempts above, including test-only support compilation inside each approved child. Approval does NOT authorize source repairs, extra launches, full API regressions, performance probes, independent review, PS5 or main integration. If the controller cannot meet safety checks, stop before launch and report the unresolved condition.
