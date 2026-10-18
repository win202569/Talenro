# P1 Candidate Single Readonly Review Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for parent preparation/disposition and superpowers:requesting-code-review for the ONE explicitly approved fresh reviewer. No implementation delegation or nested reviewer.

**Goal:** Obtain one independent source-and-saved-evidence assessment of the current P1 candidate's scalar-refusal/I1 paths and cumulative Integer/Token changes.
**Architecture:** Freeze a narrowly scoped review package,dispatch one fresh no-history reviewer,then record every finding and declined judgment without code changes. This is a NEW review route,not fulfillment/reuse of any stopped historical route.
**Tech Stack:** PowerShell file/hash/AST reads and path-scoped Git object reads;no product/test execution.
**Spec:** [Gate audit](../../roadmap/2026-10-18-p1-token-promotion-gate-audit.md),[P1 contract](2026-10-17-offline-supervision-policy.md),[original I1 review](../../roadmap/2026-10-17-offline-supervision-policy-readonly-review.md).
**Status:** PROPOSAL ONLY afteref8eb82b. User approved drafting only;no reviewer dispatched. Proposed NEW review allowance0/1UNAPPROVED/UNUSED,target-host allowanceZERO.

## Global constraints

- After explicit approval of THIS full plan,maximum ONE fresh independent review dispatch attempt. Record consumption BEFORE dispatch;an error,incomplete review or unavailable reviewer stops,no replacement or automatic retry.
- Use a fresh reviewer with no inherited conversation turns;do not reuse prior implementer/reviewer agents. Reviewer must not delegate,spawn helpers or message another chat.
- Source/data inspection ONLY:no dot-source/import,script/helper/test execution,Add-Type,compilation,synthetic probe,mutant,warmup,benchmark,target host,process inventory,OS/security/network query,install or settings/trust/ownership change.
- No source/test/oracle changes,no new runtime controller and no acceptance guard relaxation. Parent writes only package/receipt/report/status/recovery documents;reviewer writes no files/index/HEAD.
- All476 historical pins and frozen packet identities remain immutable. Full API4/4,direct2/2 and older spent/frozen slots persist. Old conditional review0/1UNUSED/INELIGIBLE and original spent review1/1 stay distinct from this proposed NEW seat.
- No main/PR merge,original replacement,candidate promotion,PS5/P2/P3/C12/security/runtime/release or whole-project clearance. ReviewCompleted is not FormalAcceptance.
- Explicit plan terms override generic skill defaults for running tests,further reviewers/auto-fixes,whole-branch Git diffs,temporary worktrees and scratch deletion. No test mandate is satisfied by silently running a forbidden reproduction.

## NEW review route and evidence limits

Approval would authorize review of later PS7 direct+separate full API/ownership
evidence as a NEW limited evidence set. It would NOT assert that old seven-slot
PS7/PS5 RED/GREEN/mutant/restoration expectations occurred or that stopped
combined30s flows now pass. Old verdicts/budgets stay unchanged.

Required BEFORE dispatch:current source identities,exact-copy fidelity,complete
PS7 direct/API/ownership receipts with their independent oracles,raw hashes,
and the original I1 contract/finding. Reviewer must inspect relevant current
P1 source and test assertions,not merely PASS totals or newest Token diff.

For this READONLY REVIEW ONLY,PS5 compatibility,reflection/shape extra-suite
execution on the Token snapshot,new Token-specific negative controls and matched
performance experiments are disclosed unresolved limits,NOT new runtime jobs
or asserted passes. This proposal does not waive them for a future broader
acceptance profile. Reviewer may judge a limit Important/insufficient for even
a PS7-scoped I1 disposition and must explain the concrete effect/evidence gap.
No input/security invariant can be accepted as a mere limitation.

Concurrency/hostile-runspace,authenticity/attestation,QPC/real clocks,live capture,
containment/cancellation/cleanup,power-loss durability and project/release gates
remain outside the synthetic module's claimed proof. Retain all18 original
declined judgments;new evidence changes only specifically justified dispositions.

## Snapshot and allowed source inventory

All paths repository-relative;use existing isolated development worktree.
Do NOT run a whole-branch diff from the old commit.

- Historical reviewed source:Git commit e8bb42e70b13a97e52d47d56cfebdc972ef85137,
  path docs/diagnostics/isolated-fuzz/offline-supervision-policy.ps1,
  blob28b4f37d8141954723076f1517b19b3697db3b8a,
  SHA256 FE23DF3BE1F5BA5DF4C2B565DEE5CFB46DAB6F57C896DA9B5C95DFE440FF0ABB.
  Read ONLY that blob as source text;no checkout/import/execution.
- Current original:same workspace module path,SHA256953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.
- Integer baseline:docs/diagnostics/isolated-fuzz/integer-candidate-validation-copy-attribution/offline-supervision-policy.ps1,
  SHA2563537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3.
- Candidate:docs/diagnostics/isolated-fuzz/integer-token-local-candidate/offline-supervision-policy.ps1,
  SHA256F9B4E408D9A4FF9C7A36EFD6E1F2472F17DACB3F76F463FF4973DAB4139CC82A.
- Direct target:same folder token-direct.fixed-members.tests.ps1,
  SHA2561EA8CDB1114EA6A7C68EF29E03F0029E277EFB64FF039EBCA8EB07B706E007A9.
  Same folder TokenDescriptorSupport.cs,source-pins.json,original failed test and
  revision report may be read to check fixture lineage;never compile/run.
- Full API tree:docs/diagnostics/isolated-fuzz/integer-token-api-regression/,
  six selected-module/test copies plus .gitattributes,source-pins.json,
  expected-output.json,README.md. Full test SHA25661A3BACF26EDE1F8B137CBF77CC4DBC35E4D6E6FB8FE6365953CDA9F41CEDEA3;
  ownership test SHA256C4905925B309B00E94834B2FDA3EE8034265B23CDF274276C9ECA5511250C6E7.
- Required evidence reports:2026-10-18-p1-token-direct-fixed-members-result.md
  (SHA256A8ACA345FC632E0D5D614DADC93E4BF9CB60B24C28EC70951E9671918BA8C686)
  and2026-10-18-p1-token-full-api-result.md
  (SHA256556CB16E748207D834A3D98C31DCF73BE2478C12CB9709CBA1724EF00EF4D37E)
  under docs/roadmap/. Gate audit and original I1 review required context.
- Direct ignored packet:.superpowers/sdd/2026-10-18-p1-token-direct-fixed-members-run/
  manifest,controller,expected-output,index,two stdout/two stderr as data.
- Full API ignored packet:.superpowers/sdd/2026-10-18-p1-token-full-api-run/
  manifest,controller,index,four stdout/four stderr as data.
  Hashes are in the required reports;all476 inherited pins may be hash-checked
  without opening unrelated protected source contents.
- Original P1 plan and joint design sections9/10,scalar-repair Task2 and relevant
  gate-audit-linked failure/negative-control reports may be read for requirements/
  chronology only. No unrelated C12/history audit or old raw replay.

Current-original→candidate changes Integer+Token function texts;review full
current P1 module and retained refusal/ownership paths too. The I1-reviewed
FE23 snapshot predates other retained repairs;do not treat953 as independently
reviewed/accepted merely because it is called current original.
Package freezes the actual checkout HEAD after plan approval and the file hashes;
Git commit range is provenance,NOT permission to inspect every changed file.

## Review focus

1. Mutable scalar metadata:fixed SDK type/member lookups and refusal before
   every scalar use/cast/branch/copy;never admit decorations through a fused path.
2. Integer/Token local reuse and removed helper-binding boundary:all original
   checks,ordinal membership,type/default/positional parameters and refusal order
   retained;no caller trust,coercion,cache or executable descriptor evaluation.
3. Result ownership:valid/refused/sticky/plain outputs and repeated calls must
   not retain mutable input references or alias results;scope I1 explicitly.
4. Fixture/oracle honesty:owner-field identity/payload checks,zero callbacks,
   CodeProperty support,late invalid index and4096/4097 tests really assert the
   contract;review saved literal outputs and completeness,not count-only evidence.
5. Evidence boundaries:ordinary PS7 scope,saved not authenticated execution,
   missing cross-engine/extra/mutation/performance evidence,old freezes and all
   declined judgments must be reported,not converted into new authority.

### Task1: One independent readonly review and stopped-or-recorded handoff

**Future files:**create docs/roadmap/2026-10-18-p1-token-review-package.md and
docs/roadmap/2026-10-18-p1-token-readonly-review.md;modify current-status/recovery.
Ignored new packet:.superpowers/sdd/2026-10-18-p1-token-readonly-review/
with manifest.json,dispatch.json,progress.md and preserved reviewer response.
No executable controller/source/test created;do not overwrite an existing packet.

**Interfaces:** consume exactly the pinned files/contracts above. Reviewer returns
Strengths,Critical/Important/Minor findings with file:line/effect,all declined
judgments,an explicit scoped I1 candidate disposition,and evidence limits.
Parent records each disposition,ReviewCompleted,ReviewVerdict,FormalAcceptance=false
and PromotionApproved=false. ReadyToMerge is NOT a meaningful authorized verdict.

- [ ] Step1:After approval,check clean branch/AGENTS,pin identities,476 immutable pins and both recent packet indices/streams as data. Freeze checkout HEAD and plan hash. If missing/mismatched evidence or safe readonly access cannot be established,stop BEFORE dispatch,no synthetic reconstruction/rerun. Git ownership errors must not trigger trust/security changes.
- [ ] Step2:Prepare concise package with description,exact source snapshots/diff scope,requirements,reports/raw paths and all five Review Focus items. Explicitly flag inherited Integer/Token and scalar-refusal lineage,PS7-only limits and original I1. Historical blob may be supplied as a hash-verified NONEXECUTABLE text snapshot if direct readonly Git access is unavailable;record provenance limit rather than claiming independent Git authentication.
- [ ] Step3:Persist NEW review attempt1/1consumed BEFORE requesting ONE fresh no-history reviewer via collaboration. No already-running agent reuse,no other chat message. Provide this plan/package;instruct source/data-only inspection,no writes/execution/delegation. Dispatch failure or incomplete result ends allowance,no automatic replacement.
- [ ] Step4:Reviewer independently reads scoped candidate/contracts/test assertions and checks recent saved evidence/hash/oracle consistency. Source-only reasoning may find defects without running them;label inferred/runtime-unobserved claims. Missing behavioral proof is not permission to reproduce. Return findings and every declined judgment,even when no blocking source finding is detected.
- [ ] Step5:Parent reconciles each finding and each declined judgment with concrete evidence;no blind acceptance or dismissal. Critical/Important or insufficient evidence =>STOP UNFINISHED,no fixes/new tests/review cycle. Minor explicitly retained/deferred. I1 candidate disposition may be Unaddressed,InsufficientEvidence or AddressedWithinPS7OfflineProfile,with rationale;original historical I1 verdict remains recorded and broader closure is not inferred.
- [ ] Step6:If no blocking findings,record only completed scoped review and its limits. Any limited acceptance,original replacement,main merge,new performance/PS5/extra-suite run requires a separate explicit decision. No AutomaticAcceptance=true or ReadyToMerge=true. Do not issue a project/security clearance.
- [ ] Step7:Save original reviewer response and all finding/limit dispositions,source/raw hashes,attempt use and remaining gates;verify no source/packet drift. Explicitly stage portable package/report/status/recovery only,commit/push existing dev branch and verify remote SHA/clean tree. Raw/session/controller/cache/credentials not committed;review record does not renew test or review slots.

## Approval requested

Approve this NEW single P1 source-and-saved-evidence review route,including
package preparation and ONE fresh independent reviewer dispatch,with ZERO new
target executions and no automatic fixes/re-review/acceptance/promotion/merge.
This is not approval to revive the old ineligible seat or waive cross-engine/
input-security requirements. Until approved,this document authorizes nothing
beyond the current planning/documentation checkpoint.
