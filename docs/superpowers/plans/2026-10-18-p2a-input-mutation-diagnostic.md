# P2-A InputMutation Diagnostic Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans after approval. Native only, no subagents. Checkboxes track the proposed diagnostic, not authorization.

**Goal:** Recover the exact failing operation/subassertion without changing product behavior or claiming a repair.

**Architecture:** Preserve the failed frozen delivery. A separate diagnostic leaf will hold an exact product copy and a mechanically instrumented test prefix; one new ordinary host records the failure site and exits. No fixes, acceptance tests or reviewer follow automatically.

**Tech Stack:** Existing Windows x64 PS7.6.5, native PowerShell and existing fixed SDK metadata; no installation or dependencies.

**Spec:** Original [P2 boundaries](../specs/2026-10-18-p2-adapter-boundaries-design.md), [approved finite implementation route](2026-10-18-p2a-timebase-plan.md), and [actual stopped evidence](../../roadmap/2026-10-18-p2a-timebase-execution.md). This proposal addresses observability only.

Status: PROPOSED_NOT_APPROVED. This turn performed readonly diagnosis and wrote
documentation only. No diagnostic source/controller/packet, host or reviewer created.

## Readonly findings

Frozen tests SHA256=11BA52E114FC8F14D2D2A020FC6E6C70C52A3F2105B0DADE548416A483B71EB9.
The only failure information is ASSERTION FAILED: P2AInputMutation. Test line195
catches every exception and drops the exception record, stack and subassertion.
This is a confirmed observability limitation, NOT a confirmed product root cause.

| Original test location | Candidate failure boundary | Existing evidence |
| --- | --- | --- |
| 159 | input and two outputs constructed | Group entered; no intermediate result saved |
| 160, helpers50–62 | enumerate input/output leaves, require initially plain metadata | No leaf/stored-type record saved |
| 161, helper66 | reference independence versus outputs | DecoratedScalars passed with empty Others; cross-owner branch not proven there |
| helper68–81 | attach, retain, mutate descriptor/payload and keep base/type stable | Five earlier input-only fixtures passed, not proof of this execution |
| helper82 | other held views and fresh rereads remain clean | Earlier empty-Others cases cannot exercise it |
| 162–164 | root payload retention and input ObservedTick replacement | No operation marker saved |
| 165 | compare both saved outputs with complete literal | Earlier NormalStop passed before these mutations only |
| 166 | new legal call remains normal after mutation | No record proves this line reached |

No row is identified as the actual failing subassertion. Neither product copying
nor fixture handling can yet be cleared or blamed. Saved run lasted1599ms, no
timeout,stderr0,owned host exited,streams complete. Rechecked775common pins,
five packet anchors and both raw stream pairs intact. Targets remain2/5, unused3
frozen, conditional review0/1ineligible. No extra execution in this audit.

## Global constraints

- Do not edit original seven portable files, old plan semantics, frozen manifests,
  index or raw streams. Original product remains SHA256
  350CAB1A36CECEA57D8335581ADC780CE1F17BDC9379D63F03130C6EEF4FC35B.
- Proposed allowance is ONE NEW ordinary diagnostic host, not the old unused slots.
  No retries, warmups, helper probes, product-only imports, mutants, fixes or review.
- No real product clocks, process supervision, handles/streams, callbacks, safety
  queries, P2-B/C/P3/main/live authority. Controller alone uses owned process/stream
  APIs and a monotonic elapsed bound as in the previous approved capture pattern.
- Preserve current worktree/branch/origin. Only necessary documentation/source for
  the diagnostic is saved; raw packets stay ignored. No new worktree or installs.
- This explicit finite diagnostic overrides generic TDD RED/GREEN, full-project,
  fix/retest/review loops. A diagnostic is not regression or acceptance evidence.

## Review Focus

1. Same original ordered prefix, not InputMutation in isolation: preserve preceding mutation/pollution context.
2. All original assertions and literals remain unchanged; instrumentation cannot weaken checks.
3. Failure context must distinguish leaf setup, selected mutation and other-owner checks.
4. Product bytes identical; observed failure remains distinct from diagnostic observer effects.
5. Unexpected earlier failure, no reproduction or incomplete capture stops with honest classification.

## Task 1 — one diagnostic capture, only after explicit approval

**Files proposed, not yet created:**
docs/diagnostics/isolated-fuzz/p2a-input-mutation-diagnostic/ with exact product
copy offline-timebase-plan.ps1, diagnostic.tests.ps1, capture.ps1, source-pins.json,
README.md and LF .gitattributes. New ignored packet
.superpowers/sdd/2026-10-18-p2a-input-mutation-diagnostic/ must not exist.
Report: docs/roadmap/2026-10-18-p2a-input-mutation-diagnostic.md.

**Consumes:** exact frozen product/test, existing ordered prefix through InputMutation,
and preserved prior evidence identities. **Produces:** diagnostic-site/exception
record and complete stream receipts, not a new product or success decision.

- [ ] **Step 1 — prepare without execution.** Verify clean expected branch/origin,
  original source/test and775common pins plus current frozen/index/seal/four streams.
  Copy exact product bytes. Derive test prefix mechanically from the original:
  preserve all helpers/assertions/literals and ordered groups ApiRequired through
  InputMutation; remove later group execution only. Keep -Group all. Add diagnostic
  scalar context markers before each candidate operation, selected field and
  other-owner check; no extra calls to product or repeated value observations.
  In TB-Need retain identical Boolean argument/throw condition, adding only an
  assertion ordinal. Capture exception type, error id, invocation line and stack
  in the outer catch. Emit structured DIAG record with case/stage/field/other field/
  assertion ordinal plus stack and diagnostic line; preserve original failure label.
  Freeze a static original-to-diagnostic line map and allowlisted exact diff.
- [ ] **Step 2 — freeze bounded controller.** No product import during preparation.
  Persist preconsumption before the single launch; engine SHA256
  362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.
  Existing engine path is recorded in the original source-pins.json. Ordinary
  -NoProfile -NonInteractive -File diagnostic.tests.ps1 -Group all only.
  Max30000ms execution+2000ms owned exit+2000ms concurrent separate raw-byte drain;
  no PID scanning, unrelated termination or elevation. Static parse/hash/diff
  checks only before launch; no warmup. Preserve all original mappings and pins.
- [ ] **Step 3 — run once and stop.** A reproduced InputMutation exception with
  exactly the previous18PASS prefix, a complete bounded DIAG record,exit1,stderr0,
  intact pins,ordinary context and confirmed owned exit/drain is REPRODUCED_SITE_CAPTURED.
  If InputMutation completes, emit DIAG outcome NOT_REPRODUCED and exit2 before
  later groups; it is not green/repair evidence. Earlier failure is EARLIER_FAILURE;
  missing/oversized diagnostic record (max64KiB), drift, timeout or unknown cleanup
  is INCOMPLETE. No outcome grants a second host or fixes. Observer-induced changes
  remain a limitation even if the same site is recovered.
- [ ] **Step 4 — readonly reconcile and preserve.** Map captured line/stack/ordinal
  to the frozen source, distinguish observed fact from hypothesis, record a focused
  proposed repair only if supported. If no site recovered, state that limitation.
  Save diagnostic source and report/status/recovery as necessary development
  checkpoint; explicit stage list,ordinary commit/push existing branch,remote SHA
  equality. Keep raw ignored records. Further repair/validation needs a new decision.

Self-review: no feature changes or product fixes; one task/one host/no reviewer.
Five focus risks map to ordered prefix, exact diff, markers/stack, byte pin and
explicit outcomes. Native method retained. This document requests review and
approval of ONE diagnostic run only; it does not resume the original five-slot route.
