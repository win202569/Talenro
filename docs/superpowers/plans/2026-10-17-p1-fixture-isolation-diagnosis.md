# P1 Fixture Isolation Diagnosis Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for Native execution only after approval. Steps use checkbox (`- [ ]`) syntax. One diagnostic host proposed; no repair or original test replay.

**Goal:** Record the earliest observed primitive category and isolation subcondition that fails, or explicitly report non-reproduction/inconclusive evidence.

**Architecture:** A new standalone probe copies only the current owned wrapper/type/decoration/isolation fixture path, splitting original assertions into named observations without changing their meaning. One fresh PS7 host processes the original category order and stops the modeled path at its first failure. It does not import/run the ownership test, product, earlier fixture cases or other probes.

**Tech Stack:** Existing PS7 runtime, owned primitive/metadata fixtures and fixed JSON rows; no dependency, installation, native code or external callback.

**Spec:** [Stopped F1 evidence](../../roadmap/2026-10-17-p1-fixture-construction-revision.md), [approved/stopped fixture revision](2026-10-17-p1-fixture-construction-revision.md) and [open product I1](../../roadmap/2026-10-17-offline-supervision-policy-readonly-review.md).

Status: **PLAN WRITING ONLY; DIAGNOSTIC HOST NOT APPROVED.** User requested continuation after c6a94949d472d88582dfca1b97be414f4645c164. The prior handoff required separately defining scope/allowance. This document proposes exactly one new PS7 host<=30000ms; it does not reuse frozen F2 or the six product slots. No probe/code/test change or run in this planning turn.

## Evidence and uncertainty

Saved F1 passed constructor/primitive-matrix/descriptor cases, then emitted only ASSERTION FAILED: P1I1FixtureIsolation. No failing category/subcondition was captured. Static inspection identifies checks of returned-array inequality, each field's Metadata descriptor/reference, unaffected second/third contents, and third-array distinctness. Boxed-input reuse, field-metadata aliasing and helper-return transport are hypotheses, not causes established by that message.

The new clean host deliberately omits F1's preceding fixture cases. Their earlier allocations/decorations and any observer effects may matter. If the isolated path does not fail, report NotReproduced and retain the original false verdict. Do not expand the probe to replay warmups, all F1 or a different engine after seeing results.

## Global Constraints

- Reuse D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator, branch codex/c12-b01-task9-coordinator, origin https://github.com/win202569/Talenro.git. Planning base c6a94949; plan commit becomes execution base. Verify AGENTS/clean state first; PowerShell only, ordinary inherited token.
- No changes/imports/runs of product, either original P1 suite, the current ownership suite, existing probes/controllers or sibling raw packets. Current ownership suite stays CF81BD28AE69E7A49897FACFBFEA68022A384353BD155F9EC521B9B451FE6256; product FE23DF3B, original35 suite61A3BACF, pre-edit9AA7 and prior protected pins unchanged.
- Preserve F1=false and all older verdicts. Prior fixtureF2 and six repair slots remain frozen/review ineligible; earlier diagnostic2/2 spent. R1–R7/product repair/conditional review remain unapproved. No budget transfer, reviewer/delegation or automatic correction.
- No alternative wrapper, changed primitive factories, weakened assertion, adaptive warmup, retry, second host/engine, whole-project test, security query/change, arbitrary process action, main merge or P2/P3/C12 acceptance.
- Copy managed owned fixture operations as diagnostic tooling only, without dot-sourcing tests. Do not mutate global type data or external state; fixtures remain local to the one child. Preserve all raw and verify fixed file/receipt hashes before/after.

## Review Focus

1. Failure must be tied to a category and exact predicate, not only a group name: fixed category/check identifiers.
2. Shared base-object identity is not automatically shared metadata: record identity independently from field/reference checks.
3. Missing/wrong descriptor must not be treated as reference inequality: distinguish kind/presence from ReferenceEquals; do not read executable descriptors.
4. Probe must preserve original failure order: later checks/categories NotReached, never force past an original failure.
5. A clean-host pass cannot overrule F1 or prove a fix: explicit NotReproduced/observer-context limits and no warmup replay.

## Proposed allowance and frozen output

ONE new PS7 diagnostic child, <=30000ms including output collection, no repeat. Engine C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe, -NoProfile -NonInteractive -File <new-probe>. Missing runtime/permission stops; no fallback/elevation/install. Owned exact child handle only, bounded wait/drain; timeout or unproved exit is stopped/incomplete, not arbitrary PID authority.

Before launch freeze probe/controller/manifest, source/test/old packet anchors and exact row inventory in NEW ignored `.superpowers/sdd/2026-10-17-p1-fixture-isolation-diagnosis/`. Preconsume the sole slot. Save separate stdout/stderr, context/host/version/admin/exit/elapsed/timeout and hashes. Required transport: exit0, one context/host pair, four category rows in order Int32/Int64/Boolean/String, one summary row and one diagnostic-only LIMIT (eight lines total). A recorded false predicate is permitted diagnostic data, not an expected product RED. Unknown/malformed/missing/extra rows, drift, unexpected exception or timeout stops without repairs.

Empty stderr or only structurally recognized CLIXML Completed-progress under the existing no-DTD/no-external-resolution/no-deserialization policy is permitted; unknown/error/trailing payload fails. No suppression or retrospective relaxation. Parent static parsing/hashing/reading saved receipts does not execute the probe and is allowed; all dynamic fixture evaluations belong to this one child.

## Task1: Fixed isolation observations and one capture

**Files:** Create docs/diagnostics/isolated-fuzz/offline-supervision-policy.isolation-probe.ps1 and, after execution, docs/roadmap/2026-10-17-p1-fixture-isolation-diagnosis.md; update this plan/current-status/repository-recovery only. Current ownership test and all existing product/probe files stay unchanged.

**Interfaces:** Private IsolationProbe- helper names copy CF81's Type/WrapPrimitive/Decorate and local factories, never import their file. Emits schema1 category/summary JSON records, no PASS labels/product API calls. Every check has Reached(bool) and Result(bool/null); null means not evaluated, not false.

Disposition rule: an original isolation check returning false yields ObservedFailure; a known helper-fidelity exception yields Inconclusive/HelperFidelity, not a reproduced isolation assertion. If every original check completes true, use NotReproduced. NotReached categories have null input measurements and all11 checks Reached=false/Result=null. Only the exact known fidelity message is captured; constructor/unknown exceptions remain infrastructure failures.

- [ ] **Step1 — write/freeze the whole probe before execution.** Use the exact existing factory expressions in original order: Int32(0), Int64(0), true, and a newly constructed clean string. Copy selected object constructor, combined fidelity guard and metadata construction unchanged except helper prefixes. Create a/b then decorate a/b in the original order. Capture input BaseTypeNameA/B and BaseObjectSame as observations only. No alternate construction/factory or earlier constructor/matrix/descriptor-case execution.
- [ ] **Step2 — name every original check and preserve evaluation order.** Fixed check IDs: ArraysABDistinct; ADescriptorNote; AFieldMatchesAArray; BDescriptorNote; BFieldMatchesBArray; BContentsInitial; CContentsInitial; CArrayDistinctA; CArrayDistinctB; CDescriptorNote; CFieldMatchesCArray. Mutate am[0]='changed' only after ArraysABDistinct succeeds, exactly where the old path does. Create c only after the preceding checks succeed. Split the original third combined assertion left-to-right; stop on its first false component. For field identity inspect descriptor kind first, read Value only for the known owned Metadata NoteProperty. Use ReferenceEquals, not value-equality, for array identities. Never invoke executable members. Do not convert unavailable checks to false.
- [ ] **Step3 — freeze the exact five-row schema.** Each category row: Schema=1, Kind=category, Category, Status=Observed/NotReached/HelperFailure, BaseTypeNameA/B(string/null), BaseObjectSame(bool/null), Checks(exact11 ordered check objects with Id/Reached/Result), FirstFailedCheck(string/null). Known copied-helper fixture-fidelity exceptions become HelperFailure with FirstFailedCheck=HelperFidelity and remaining checks NotReached; unknown exceptions remain infrastructure failures. On any false check/helper failure, later categories are NotReached. Summary fields: Schema=1, Kind=summary, Disposition=ObservedFailure/NotReproduced/Inconclusive, FirstFailedCategory(string/null), FirstFailedCheck(string/null), WarmupReplayed=false, ProductImported=false, OriginalTestsRun=false, FixtureChanged=false. Freeze metadata types/field sets/order before launch, not after observing output.
- [ ] **Step4 — static review/pin and launch once.** Verify copied semantics against CF81, original candidate/test/product/raw unchanged, static parsing clean and controller row checks frozen. Launch the one ordinary-token PS7 child. Preserve all output even on transport failure; no post-failure code change, retry or another category pass. Read the saved rows, not an interactive probe invocation.

## Task2: Evidence-based localization and handoff

- [ ] **Step1 — reconcile every row/check/hash.** Report the first observed failing category/predicate, not a universal runtime cause. BaseObjectSame alone never proves metadata contamination. If none fails, explicitly state NotReproduced because prior F1 context/warmups were omitted; contradictory/ambiguous observations remain Inconclusive. Instrumentation may affect wrapper transport; no independent review or authenticated historical replay is claimed.
- [ ] **Step2 — record the decision boundary.** Product I1 and fixture acceptance remain unresolved. No post-diagnostic repair or F2/R1–R7/reviewer dispatch follows automatically. Any future change requires a narrowly reasoned proposal/new allowance; do not drop isolation requirements to manufacture acceptance. Preserve original F1=false and every consumed/frozen budget.
- [ ] **Step3 — ordinary commit/push necessary files.** Explicit stage list, document links/static parsing/whitespace and integrity checks; existing origin/branch, remote SHA=local and clean tracked status. Never commit raw/controller/cache/binaries/secrets; preserve ignored packets locally. Recover latest development branch, not main; cloning cannot restore raw/conversation/runtime state.

## Self-review and approval handoff

This is a single diagnostic question, not a fourth construction attempt. All five Review Focus items are pinned to row/check rules; no existing assertion or test is changed. One new host is proposed, with no fallback or independent reviewer. Native method is preserved, but the new single-host allowance requires explicit written-plan approval before probe creation/execution. Generic extra TDD/baseline/task-done/whole-project/review/deletion defaults do not override these limits. Cost: omitted warmup/observer context may leave diagnosis inconclusive. Current turn is documents only.
