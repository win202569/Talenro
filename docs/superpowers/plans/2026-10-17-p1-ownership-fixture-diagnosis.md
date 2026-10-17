# P1 Ownership Fixture Diagnosis Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for Native diagnostic execution after approval. Steps use checkbox (`- [ ]`) syntax. This plan does not authorize product/test repairs.

**Goal:** Identify which fixture-fidelity condition failed, or record the result as inconclusive, without changing the frozen tests or reclassifying the original failure.

**Architecture:** One new standalone diagnostic probe copies only the owned fixture-construction path and emits fixed structured observations at each boundary. It never imports the product, original tests or their helpers; two fresh bounded hosts examine PowerShell7 and5.1 separately. A portable report distinguishes observed facts, hypotheses and limitations; no repair follows automatically.

**Tech Stack:** Existing Windows PowerShell5.1/PowerShell7, owned primitive/PSObject fixtures, descriptor inspection and bounded text receipts; no installation or new dependency.

**Spec:** [Stopped repair record](../../roadmap/2026-10-17-p1-scalar-ownership-repair-execution.md), [approved but stopped repair plan](2026-10-17-p1-scalar-ownership-repair.md), and [P1 I1 ownership requirement](../../roadmap/2026-10-17-offline-supervision-policy-readonly-review.md).

Status: **PLAN WRITING ONLY; TWO DIAGNOSTIC HOSTS NOT APPROVED.** User yes after stopped checkpoint25bc3b3a80311cf6989abdd5543c6d1d8e219adc requested this plan only. The original repair slot1 remains ExpectedResult=false; slots2–7 remain frozen, review0/1 remains ineligible. This proposal requests a separate allowance, not transfer/renewal of those six slots.

## Known facts and hypotheses

The saved PS7 child reached API PASS, then failed P1I1-Decorate's combined guard before the first decorated-scenario call. The guard checks primitive type, primitive value, Metadata descriptor existence/kind, and nested-array reference identity. The saved exception does not name the failing condition. It is not evidence that the product refused or accepted the fixture.

Read-only trace: caller assigns Int32(0) to normal progress ExitCode; helper saves original/type, constructs PSObject(original), allocates object[]('initial'), attaches Metadata NoteProperty, assigns wrapper into the caller field, retrieves it, checks the combined guard, then would return metadata. Potential boundaries are construction, property attachment, field assignment/retrieval and helper argument/return transport. None is yet established as the cause. The proposed probe observes the original construction first; it does not try alternative wrappers or implement a correction.

## Global Constraints

- Reuse D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator, branch codex/c12-b01-task9-coordinator, verified origin https://github.com/win202569/Talenro.git. Planning base25bc3b3a; plan commit becomes execution base. Verify clean state/AGENTS first.
- Do not modify/import/run offline-supervision-policy.ps1, either P1 test file, protected dependencies, old controllers or sibling session packets. No exploratory shell evaluation of their helper functions. No product guard, fixture repair, counter/expectation change or original-slot replay.
- Preserve original source FE23DF3BE1F5BA5DF4C2B565DEE5CFB46DAB6F57C896DA9B5C95DFE440FF0ABB, original suite61A3BACF26EDE1F8B137CBF77CC4DBC35E4D6E6FB8FE6365953CDA9F41CEDEA3, new frozen suite9AA7C29CCB640270EBCFD562F75AA3A47C95324C5A4CE786E0E354BDBC0D694F and all15 prior protected pins.
- Read old manifest/index/streams solely to hash/reconcile their preserved contents; never update them. Retain false verdict, stopped manifest086E26F5, index410D7449, stdoutB2E1D823 and stderr6A1A52B8. Full hashes remain in the stopped record.
- New probe uses only internally constructed fixtures; no external callbacks, hostile runspaces, type-data mutation, arbitrary processes, security queries, native APIs or protected workload. It is test tooling, not product functionality. No installs, elevated test hosts, security/trust/ownership changes, main merge or P2/P3/C12 acceptance.
- No diagnostic rerun, adaptive alternative construction, repair/retest, automatic reviewer or extra host. Any infrastructure failure freezes the rest; a valid diagnostic observation reporting a false predicate is data, not a demand to make it pass.
- Preserve all raw. Check hashes before/after each diagnostic. A changed source/test/pin/earlier receipt is infrastructure failure. Git preservation authority remains separate from test-host authority.

## Review Focus

1. Type/value changes must not be confused with descriptor loss: emit separate named predicates.
2. Metadata presence must not be confused with array identity: emit descriptor kind and ReferenceEquals independently.
3. Instrumentation may change wrapper transport: retain the original combined guard and mark instrumented-versus-original disagreement inconclusive.
4. PS7 results must not be generalized to PS5.1: report each engine independently, including differing results.
5. Successful diagnostic execution must not be called a product test pass or I1 repair: output diagnostic-only terminology and retain all frozen budgets.

## Proposed new allowance

Exactly TWO sequential fresh synthetic diagnostic hosts, at most30000ms per host including probe execution/output collection. SlotD1 PS7 then slotD2 PS5.1. No use of the frozen six repair slots; no reviewer allowance. Diagnostics stop afterD2 even if cause is unknown. Parent authoring/static parsing/hash/receipt checks do not execute probes and are allowed after plan approval.

Fixed existing engines: C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe and C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe. Invocation: -NoProfile -NonInteractive -File <probe>; PS5 inbox PSModulePath. Ordinary inherited token only; missing engine/permission is a stop, not fallback or escalation.

Freeze probe/controller/schema/expected row inventory and integrity pins beforeD1. Preconsume each slot. New ignored packet `.superpowers/sdd/2026-10-17-p1-ownership-fixture-diagnosis/` holds manifest, controller, index and separate stdout/stderr. Controller owns only the exact child handle; bounded wait/drain, timeout stopped/cleanup unknown if unproved, no arbitrary PID cleanup. Empty stderr or exclusively structurally recognized CLIXML Completed-progress under the existing strict no-DTD/no-external-resolution/no-deserialization policy; other stderr stops. Context/version/identity/exit/elapsed/timeout and every output hash are retained. No retrospective verdict relaxation.

## Task1: Instrument the fixture path and capture bounded observations

**Files:** Create docs/diagnostics/isolated-fuzz/offline-supervision-policy.fixture-probe.ps1; create docs/roadmap/2026-10-17-p1-ownership-fixture-diagnosis.md at handoff; update this plan/current-status/repository-recovery. No existing source/test edit.

**Interfaces:** Probe owns separate FixtureProbe- helpers. Consumes no product API/input files. Emits context, schema1 diagnostic rows and a final diagnostic-only marker, not PASS labels or product assertions. Parent consumes those rows to localize observations, not to generate expected product decisions.

- [ ] **Step1 — write the entire probe and frozen observation contract before execution.** Copy the current fixture constructors (Scenario/Observation/Normal/Type) and the exact scalar decoration operations into isolated probe helper names without dot-sourcing tests. Use the same complete normal scenario and Int32(0) ExitCode, same mutable single-element object array containing 'initial', same Metadata name and NoteProperty. Keep the original combined fidelity expression in an owned path, catch its fixture failure as a diagnostic row rather than treating it as a product result. No alternative construction or modification to original files.
- [ ] **Step2 — pin rows and independently defined predicate meanings.** For stages original, wrapper-created, metadata-added, field-assigned, field-retrieved, helper-return, report exact primitive BaseTypeName, PrimitiveValueMatches(bool), MetadataPresent(bool), MetadataKind(string or null), MetadataReferenceSame(bool or null). Missing descriptor means null identity, not false primitive equality. Report helper guard predicates individually as TypeMatches, ValueMatches, DescriptorPresent, DescriptorIsNoteProperty, NestedReferenceSame; report OriginalCombinedGuard(bool) from the unaltered combined expression separately. Do not read arbitrary descriptor values: inspect kind first and read Value only for the known owned Metadata NoteProperty. Emit first-false predicate names in fixed order, or empty array. Freeze scalar/null types, exact row count/order and field set; unknown/missing/duplicate rows are infrastructure failure.
- [ ] **Step3 — protect against observer effects.** Execute the unaltered owned helper path once on a fresh fixture and the instrumented path once on a separate fresh fixture inside EACH approved diagnostic host; neither calls product. Record whether their combined guards agree, including disagreement or exception. The original helper's thrown fidelity error is a named captured diagnostic outcome, not swallowed infrastructure stderr. Instrumented helper-return row may report NotReached when guard throws; do not force past the guard. Any original/instrumented disagreement, multiple indistinguishable boundaries or all predicates true despite the prior saved failure yields Inconclusive, not a fabricated cause. Do not run alternative probes to resolve it under this allowance.
- [ ] **Step4 — statically parse/freeze and runD1 only.** Verify source/test/protected pins and untouched old packet hashes; launch fixed PS7 once. Required transport result: valid complete declared diagnostic inventory, exit0 and allowed stderr within30000ms. Predicate false and a captured fixture failure are permitted data. Uncaught error, malformed output, drift or timeout freezesD2. Preserve receipts regardless.
- [ ] **Step5 — runD2 ifD1 transport/integrity held.** Fixed PS5.1, exactly the same frozen probe and schema. Same transport requirements; do not require matching PS7 predicate values. No third host, retry, product call or reviewer.

## Task2: Readonly diagnosis and durable handoff

**Interfaces:** Consumes the two saved diagnostic stream pairs and frozen row schema, produces a reasoned report or explicit inconclusive result. No automatic test-fixture correction follows.

- [ ] **Step1 — reconcile every row/hash against saved streams.** Explain earliest observed divergence per engine and which guard predicate(s) account for that run's failure. Distinguish local observation from language-wide claims, observer effects and the original stopped execution. Do not assert a root cause beyond the captured evidence; the initial message alone cannot identify one. No external technical claims are needed; if they become necessary, use official primary documentation and label inference.
- [ ] **Step2 — preserve current blockers and next decision.** Record actual D1/D2 outcomes and unproved matters, product I1 still open, original slot1 false, six repair slots still frozen/review ineligible. Propose at most a separately approved fixture correction/revised repair sequence if evidence supports it; otherwise state exactly what remains unknown. Do not edit the frozen test/guard or create a wider repair budget in this task.
- [ ] **Step3 — ordinary commit/push of necessary portable files.** Explicit file list, documentation links/static parsing/whitespace, unchanged source/tests/pins, no raw/controller/cache/binaries/secrets. Verify origin/existing branch, remote SHA=local and clean tracked worktree. Preserve all ignored raw locally, including failures. Fresh computer reads current-status and this plan/report, not main or an assumed restored session.

## Self-review and approval handoff

All five Review Focus items map to fixed observation definitions and explicit inconclusive rulings. This is a diagnostic question/probe, not a product repair: which predicate/boundary differs during the owned decoration path? The existing error is the evidence gap; writing the probe is not permission to fix the fixture. Literal diagnostic predicates replace product-derived expected output; no original test expectation changes. Native execution method is preserved, but this NEW written two-host diagnostic allowance still requires explicit approval. No reviewer/delegation or historical branch audit is proposed.

This user-requested plan is written under systematic-debugging/writing-plans; generic reproduction/extra-test/automatic-fix defaults do not override its finite scope or the stopped repair's gates. Cost: the diagnosis may remain inconclusive and another decision may be required. Current turn is documents only; zero diagnostics, test imports, product changes or budget use.
