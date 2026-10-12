# SDD ledger — plan: docs/superpowers/plans/2026-10-09-task10-desired-signature-evidence.md

用户已批准计划及新增纯单元测试。Native，无提交推送。
BASE: 6752b9612a00b9d6b0bf87ae8dd95063f6b056ee

## Follow-up: active-root error classification

User requested next step after the review. Scope: fix the one deferred classifier
Minor with TDD, preserving rejection, ordinary signatures, limits and all other
gates. Diagnosis: the unequal active-root branch returns BindingMismatch before
calling the sole format parser, unlike other envelope paths. New tests cover
0/1/32 rotations, exact matches, malformed/noncanonical/bad-digest/wrong-Kind
inputs, valid-but-different identities, cancellation and owned artifact slot 35.
Fresh baseline: 500 existing visible project files. No commit/push or broad tests.
RED: targeted tests exit1, 3.802s; all 12 malformed/noncanonical/bad-digest/
wrong-Kind cases across 0/1/32 rotations returned the wrong class. Exact match,
valid-different, cancellation and owner-slot controls remained covered.
Fix: only unequal active bytes go through existing owner.parse with fixed
InitialRoot (no rotation) or RootRotation Kind, registered in free slot 35.
Malformed inputs now return InvalidInput; valid different artifacts still return
BindingMismatch; identical-byte fast path and all acceptance behavior remain.
Two older wrong-Kind test expectations updated from BindingMismatch to
InvalidInput, with separate valid-but-different cases retaining the former.
GREEN: full stateverify exit0, 3.088s. Allowed regression all exit0: stateformat
4.060s, serving6.214s, contracts6.668s, readiness4.486s; selected authority
6.618s; store compile-only6.336s [no tests to run]. This Minor is closed.
Formatting listing empty; branch and HEAD unchanged, git diff --check exit0.
Only verify.go, two tests and progress documents changed in this follow-up;
old fixture bytes, parser/signature semantics, dependency locks and other sources
are preserved. The old review/deferred entries below are historical evidence.
This is a follow-up bugfix, not a second independent review or a trust expansion.

Final protection: 469 baseline entries checked after documentation updates; exactly seven allowed changes (five stateformat sources, current plan, current-status). Remaining 462 baseline files, including every old format fixture and serving/authority/store/locks/specs, match. New files are scoped additions only.

Final review: fresh /root/signature_evidence_final_review (gpt-6-astra/high), read-only; 0 Critical, 0 Important, 1 Minor. All five focus areas checked; no signature acceptance or owned-lifetime defect found. Review did not rerun tests.
Final: minor (deferred): verify.go:94 classifies malformed bounded ActiveRootEnvelope as ErrBindingMismatch instead of ErrInvalidInput. Still rejects, no bypass. Main agent verified branch and accepts Minor classification; no fix pass needed. Two previous canonical Minors addressed; two historical test Minors remain open.
Final: Ruling: exclude deployment trust/time/freshness/rollback/root installation/production authorization and historical branch from verdict — approved conditional-only increment — cost: cannot use this result to serve or install roots.
Final: Ruling: test runs and protected hashes are author evidence, not independently rerun by reviewer — reviewer restricted to read-only source inspection — cost: no independent execution corroboration.
Final: Ruling: direct owner panic tests plus scoped source inspection do not prove library/runtime copies, allocation failure or concurrent-owner cleanup — documented ownership boundary — cost: no general memory-erasure or injected-verifier-panic guarantee.
Final: Ruling: reviewer HEAD check was blocked by Git ownership; rely on author's scoped read-only escalated HEAD verification — no security settings changed — cost: no independent Git confirmation.
Final: Ruling: whole-project/security gates and historical Minor closure remain unaccepted — no corresponding tests/review in scope — cost: later acceptance work remains required.

Current: Tasks 1–5 complete locally; original review had no Critical/Important, and its one Minor is now fixed by the TDD follow-up above. No second independent review. In-progress/pending lines below are historical checkpoints, not current state. No commit/push; worktree and scratch evidence preserved.

Final verification (after independent review, no source fix): all exit0 — stateformat3.932s, stateverify3.870s, serving6.737s, contracts6.422s, readiness3.343s; selected authority6.977s; store compile-only4.737s [no tests to run]. HEAD again 6752b9612a00b9d6b0bf87ae8dd95063f6b056ee; git diff --check exit0. No running tests. Worktree/scratch retained because there are no commits; no destructive cleanup or integration.

Pre-flight: Task1→3/4 材料类型一致；Task2→4 Batch/evidence 一致；Task3→4 verifyThreshold 借读材料不持有；Task4→5 owner.Clear 一致。
Ruling: Bash 专用 sdd/task 脚本用 PowerShell 读取 brief、运行限定测试并 apply_patch 记录等效步骤 — 用户限定 PowerShell 且无提交 — 代价是记录需人工维护。
Ruling: 只运行计划允许测试，不采用技能默认全仓/下载/提交/清理 — 已批准范围优先 — 整体验收仍未知。
Ruling: 以开始时 469 个文件的 hash 和可修改源快照划定本次增量，不用历史 merge-base 全量差异 — 原分支有未提交旧成果 — 代价是审查需同时核对快照。

Task 1: in_progress；2–5 pending。

Task 1: complete — BASE unchanged; baseline stateformat PASS 2.987s; selector RED exit1 (kind gate/owned clear; canonical limit+1 mutation rejected test); restored guard; full stateformat GREEN exit0 3.345s. Fixed test-only copied digest typo using independent .NET SHA256 before accepting RED. No wire fixtures changed.
Task 2: in_progress.

Task 5: complete — final combined allowed regression exit0: stateformat 4.842s, stateverify 4.916s, serving 8.486s, contracts 7.707s, readiness 3.884s. Selected authority exit0 9.322s; store compile-only exit0 8.597s [no tests to run]. All formatting listings empty. No test sessions pending. Final independent read-only review dispatched next; not yet accepted.

Task 4: complete — full stateverify GREEN exit0 1.669s for chain/bindings/windows/conditional evidence plus owner/cancel tests. Additional distinct-root rotation (new keys 4/5/6 threshold2, metadata3 votes, swapped groups, too few votes) added; stateverify GREEN exit0 2.886s.
Task 5: in_progress — boundary checker RED exit1 forbidden net/http 2.317s then implemented AST allowlist/capability checks. Nonempty-share alias mutation RED 6.367s; restored clone. Final combined allowed regression pending.
Ruling: correct plan's nonexistent internal/nodecontrol/readiness and internal/nodecontrol/store to actual internal/readiness and internal/store — same approved components, no added tests — cost if wrong: accidental scope expansion, prevented by prior ledger/path inspection and store -run '^$'. Wrong-readiness-path run exited1 despite other package passes, not accepted as a green suite.
Protection check: of 469 baseline files, only approved plan and five existing stateformat sources changed; other 463 including all old fixture bytes unchanged. New files limited to stateverify, stateformat/material pair and execution ledger. HEAD unchanged; git diff --check exit0.

Task 3: complete — threshold selector RED 6.506s; full pure package GREEN 2.729s; early-quorum mutation RED TestThresholdKnownBadAfterQuorum 2.615s, restored full-check; Evidence/Frozen/Threshold GREEN exit0 5.659s.
Task 4: in_progress — TestVerify RED exit1 2.591s (0/1/32, binding/window/conditional positives rejected by closed stub); implementing fixed-array owner and ordered verification.
Task 5 ownership test written early because Task4 consumes its cleanup — TestOwnerExitCleanup RED exit1 1.588s for error/cancel/panic; no task completion yet.

Task 2: complete — TestEvidenceOwnership RED exit1 lost evidence; evidence/frozen GREEN exit0 4.987s. Frozen digest test first caught apply_patch trailing LF; explicit fixture-only end-of-line normalization restored independently generated bytes, not changed oracle. Bad domain mutation rejected. Task 3: in_progress; threshold RED exit1 q=1/votes=1 and separate-group positive rejected by stub.
Ruling: frozen transcript bytes stored as hex text (.transcript.hex) and decoded by tests — apply_patch textual artifacts cannot portably represent NUL — exact decoded domain+NUL+P remains oracle, at cost of one decoding step.

