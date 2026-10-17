# P1 Fixture Construction Revision Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for Native execution after approval. Steps use checkbox (`- [ ]`) syntax. This plan proposes fixture edits and two fixture-only hosts, NOT product repair.

**Goal:** Construct decorated primitive fixtures without changing their type/value, preserving every existing fidelity check and product expectation.

**Architecture:** Explicitly select PSObject's object-parameter constructor in the ownership test's two wrapper-construction sites. Add a fixture-only group that returns before importing product or running product assertions. Validate it once per engine; keep product repair and its revised sequence separately gated.

**Tech Stack:** Existing Windows PowerShell5.1/PowerShell7, reflection over the installed managed PSObject type, owned fixtures and literal expectations; no Add-Type, installation or new dependency.

**Spec:** [Recorded construction-boundary observations](../../roadmap/2026-10-17-p1-ownership-fixture-diagnosis.md), [stopped repair requirements](2026-10-17-p1-scalar-ownership-repair.md), and [Important I1](../../roadmap/2026-10-17-offline-supervision-policy-readonly-review.md).

Status: **APPROVED FIXTURE REVISION; STOPPED ON F1 ISOLATION ASSERTION.** User yes afterb98251f4bcb3cf5a3069dcc3704705706eae9e3d approved fixture edits/two hosts only. F1 consumed: exit1/973ms, three fixture cases passed then FixtureIsolation failed, ExpectedResult=false. F2 unrun/frozen, no automatic retry/fix/product import/test/reviewer. See [stopped report](../../roadmap/2026-10-17-p1-fixture-construction-revision.md). Original repair false verdict/six frozen slots/review ineligible and diagnostic2/2 spent unchanged. Historical proposal wording below grants no renewal.

## Evidence, candidate and alternatives

Saved D1 PS7 first diverged at wrapper construction: Int32 became PSCustomObject, type/value predicates false, metadata/reference checks true. D2 PS5.1 retained Int32 and all predicates. Both copied original/instrumented paths agreed. The frozen diagnostic conclusion remains Inconclusive; it did not inspect overload binding or prove a replacement.

Microsoft documents distinct PSObject(Int32) member-capacity and PSObject(Object) wrapping constructors. This supports selecting the object signature as a candidate, not claiming it was the uniquely established cause of D1. [Official constructor reference](https://learn.microsoft.com/en-us/dotnet/api/system.management.automation.psobject.-ctor?view=powershellsdk-7.6.0).

Recommended candidate: resolve GetConstructor with exactly one System.Object parameter, verify that signature, then invoke that ConstructorInfo with a one-element object argument array holding the original primitive. Do not use numeric-overload inference. Alternative AsPSObject may reuse an existing wrapper; this proposal does not evaluate or fall back to it. Removing type/value/identity guards is not an alternative. Any candidate failure stops, rather than trying another construction in the same allowance.

## Global Constraints

- Reuse D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator, branch codex/c12-b01-task9-coordinator, origin https://github.com/win202569/Talenro.git. Planning base535c54da; plan commit becomes execution base. Verify clean state/AGENTS first. Windows PowerShell only, ordinary inherited-token hosts.
- Modify ONLY docs/diagnostics/isolated-fuzz/offline-supervision-policy.ownership.tests.ps1 for fixture construction/new fixture-only group, plus required documents. No product source, original35-case suite, accepted raw coordinator/dependencies or diagnostic probe edit/import/run.
- Product source remains FE23DF3BE1F5BA5DF4C2B565DEE5CFB46DAB6F57C896DA9B5C95DFE440FF0ABB; original suite61A3BACF26EDE1F8B137CBF77CC4DBC35E4D6E6FB8FE6365953CDA9F41CEDEA3 and prior protected15 remain unchanged. No new source guard or model/state-machine/output change.
- Ownership test current baseline9AA7C29CCB640270EBCFD562F75AA3A47C95324C5A4CE786E0E354BDBC0D694F is intentionally superseded ONLY if execution is approved. Retain its pre-edit bytes in the NEW packet and Git history. Old manifests retain their historical9AA7 pin; do not pretend the revised current file still matches that old pin. Never rewrite old manifests/indexes/verdicts.
- Existing P1I1-Decorate combined guard, separate expected13/22-field product decisions, ten existing case labels, callback counters and fixture sanity assertions remain intact. Only its wrapper construction and PassiveDescriptors' wrapper construction switch to the selected object signature. No weakening, output suppression or retrospective expected-result adjustment.
- No product/test import in fixture group; return before the existing module dot-source/old-suite branch/API/product cases. No original-slot replay, mutation, product regression group or combined45-case run under this allowance.
- Old failed slot1 stays ExpectedResult=false, six repair slots frozen, reviewer0/1 not eligible. Prior diagnostic2/2 stays spent. Proposed fixture hosts are separately budgeted, not transferred from either packet.
- Preserve all raw/session packets. No installs, native code, real workload, arbitrary PID action, protection/trust/ownership change, review/delegation, whole-project tests, main merge or P2/P3/security/C12 acceptance.

## Review Focus

1. Constructor resolution must select System.Object, not integer capacity: FixtureConstructor.
2. Int32/Int64/bool/string values and decorated metadata must survive without coercion: FixturePrimitiveMatrix.
3. Script/Alias/Method fixture descriptors must survive construction without invocation: FixtureDescriptors.
4. Independent fixtures and metadata must not contaminate each other through wrapper reuse: FixtureIsolation.
5. Fixture success must not become product/I1 success or silently launch old tests: early-return fixture group and explicit fixture-only LIMIT/FixtureComplete.

## Proposed fixture allowance — two sequential hosts only

F1 PS7, thenF2 PS5.1 ifF1 meets all exact expectations; each<=30000ms including output collection. No retry/third host/alternate construction/reviewer. Freeze revised suite, controller, expected inventory and integrity pins beforeF1. Any unexpected assertion, drift, stderr, timeout or environment failure stops; preserve unfinished state, no automatic fixes.

Fixed engines: C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe and C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe. Invocation -NoProfile -NonInteractive -File <ownership-test> -Group fixture; PS5 inbox PSModulePath. Missing permission/runtime stops; no fallback/elevation/install. Ordinary inherited token, exact owned-child handle, bounded wait/drain; timeout/unproved cleanup is incomplete, not arbitrary PID authority.

Expected each: exit0, exact five ordered labels P1I1FixtureConstructor, P1I1FixturePrimitiveMatrix, P1I1FixtureDescriptors, P1I1FixtureIsolation, P1I1FixtureComplete, one context/host pair and one fixture-only LIMIT marker. No old P1/P1I1 product-case labels. Empty stderr or exclusively structurally recognized CLIXML Completed-progress under the existing strict no-DTD/no-external-resolution/no-deserialization policy; unknown/error/trailing data stops. Freeze exact metadata/inventory beforeF1, retain every original stream and false verdict.

New ignored packet `.superpowers/sdd/2026-10-17-p1-fixture-construction-revision/` owns its ledger, controller, manifest/index, pre-edit suite snapshot and F1/F2 streams. Pin unchanged product/protected15 and all original stopped/diagnostic raw anchors read-only. Pin revised suite as a NEW version, not an old9AA7 match. No sibling packet edits.

## Task1: Fixture-only constructor revision and cross-engine checks

**Files:** Modify docs/diagnostics/isolated-fuzz/offline-supervision-policy.ownership.tests.ps1; create docs/roadmap/2026-10-17-p1-fixture-construction-revision.md at handoff; update this plan/current-status/repository-recovery. No product or diagnostic probe edit.

**Interfaces:** Add private P1I1-WrapPrimitive([object]$Value), used only by owned fixtures, returning the result of the selected object constructor. New Group=fixture executes self-contained fixture assertions and returns before any existing import/product/old-suite path. Existing regression/all groups retain product expectations unchanged; they are not run here.

- [x] **Step1 — preserve baseline and write all fixture checks before changing construction.** Retain original9AA7 suite bytes and constructor-site diff; no execution of the old suite. Extend ValidateSet with fixture and add five independent checks in an early branch. Move only identity/host emission needed for that branch, not product/helper semantics. The branch's required data/guards are defined before use. Static call-graph inspection must show no module/test invocation reachable from fixture mode.
- [x] **Step2 — pin constructor and primitive expectations.** FixtureConstructor requires exactly one constructor parameter of type System.Object and a nonnull public ConstructorInfo; fail if absent, do not select another signature. FixturePrimitiveMatrix checks exact input BaseObject type/value, Metadata NoteProperty and same fresh nested object-array reference after construction, assignment, retrieval and helper return for Int32 0/1/-1/MinValue/MaxValue, Int64 0/MinValue/MaxValue, bool true/false, and strings empty/clean/refused/unknown. Assign wrappers directly into fresh own fixture fields without typed coercion. Keep the original combined fidelity guard. Fixture sanity/constructor failure must stay a fixture failure, never the planned P1I1DecoratedInteger product RED.
- [x] **Step3 — pin descriptor and ownership expectations.** FixtureDescriptors uses the same selected constructor for separately created ScriptProperty, AliasProperty and ScriptMethod fixtures, verifies exact primitive type and expected descriptor kind/name, never reads/invokes executable members, and requires callback count0. FixtureIsolation creates two fresh owned fixtures of each primitive category, attaches distinct Metadata arrays and checks reference inequality; mutate first array and require the second and a third fresh fixture unchanged. Do not mutate global type data or shared test literals. These checks validate fixture fidelity, not product output independence.
- [x] **Step4 — implement the sole candidate.** P1I1-WrapPrimitive resolves `[Management.Automation.PSObject].GetConstructor([type[]]@([object]))`, verifies the exact public one-object signature, invokes it with `[object[]]@($Value)` and preserves the returned wrapper through the fixture path. Reject null (not a wrapped primitive). Replace only the two previously ambiguous construction expressions with calls to this helper; all type/value/descriptor/reference guards remain. If return transport loses wrapper fidelity, stop at validation, do not substitute another API or weaken checks.
- [ ] **Step5 — static review, freeze and runF1.** Confirm only authorized fixture/group edits, no product expectation changes; statically parse, pin updated suite/protected/raw/controller/inventory, preconsumeF1, run once in PS7. Compare exact five labels/streams/context and all integrity hashes. Unexpected result freezesF2; no code repair loop.
- [ ] **Step6 — runF2 only afterF1's exact expectations hold.** Same frozen suite and schema in PS5.1, same five literal checks; preserve outputs and recheck all pins. No product call, old-suite execution or third host. This plan does not include a dynamic original-RED rerun; the prior original construction failure is retained as historical evidence only.

## Task2: Save fixture result and explicitly separate the product sequence

**Interfaces:** Consumes the two fixture receipts/versioned test diff; produces a bounded fixture disposition and a separately gated product verification proposal, not I1 closure.

- [x] **Step1 — reconcile saved results and limits.** State actual exact fixture outcomes on each engine, immutable old failure/diagnostic records and authorized ownership-test version change. If both match, call only the tested fixture vectors verified; no universal constructor/root-cause, product alias reproduction or repair claim. No independent review occurred. Any failure remains unfinished and stops further work.
- [x] **Step2 — record the proposed product sequence below without executing it.** Six frozen slots cannot cover two valid REDs, two initial GREENs, one mutation and two restored final GREENs (seven runs). Do not count failed old slot1 as a valid RED, drop a required check or automatically spend a seventh host. Original product-repair allowance remains stopped. A subsequent approval must explicitly authorize a replacement seven-run allowance/sequence and its conditional scoped review; whether to retire the unused old allowance is a human decision, not silent budget transfer.
- [ ] **Step3 — save/push necessary portable files even on stop.** Explicit stage list, relevant static parsing/whitespace/document links and integrity checks; ordinary existing origin/branch push, confirm remote=local and tracked worktree clean. Exclude raw/controllers/cache/binaries/secrets; preserve all original and new packets locally. Recovery starts from latest development branch, not main; cloning does not restore ignored evidence or runtime state.

## Proposed future product verification order — NOT authorized here

|Future run|Scope/required result|
|---|---|
|R1|PS7 revised regression suite against unchanged product: precise P1I1DecoratedInteger RED, not fixture error|
|R2|PS5.1 same unchanged product: same precise RED|
|R3|After separately approved I1 source repair, PS7 combined45 exact GREEN|
|R4|PS5.1 same fixed product, combined45 GREEN|
|R5|PS7 guard-removed mutant, exact integer RED; restore exact fixed bytes immediately|
|R6|Restored PS7 combined45 GREEN|
|R7|Restored PS5.1 combined45 GREEN|

Those would each require the original120000ms ceiling/ordinary-token/integrity/stream/fail-stop gates, a fresh packet and explicit approval before any run or product edit. One fresh scoped readonly reviewer would be conditional on all seven outcomes matching, not an automatic repair/re-review cycle. They are not F1/F2 and are not launched by fixture success. Product I1/security/Stage B/P2/P3/main/C12/release remain unaccepted.

## Self-review and approval handoff

All five Review Focus items map to named fixture checks; no product assertion was changed to produce a pass. The candidate is grounded in saved boundary observations and documented distinct constructor signatures, but not dynamically verified. Explicitly selecting a signature does not establish historical overload selection. User requested this written plan, overriding the generic small-change preference for no plan document; cost is extra documentation, not extra execution authority.

Native fixture-only execution was approved, then stopped after unexpectedF1 isolation failure. Current ownership suite is an unaccepted revised version; old snapshot/raw preserved. No future R1–R7/source repair/reviewer approval follows. Generic extra runs/automatic fixes/reviews/packet deletion remain excluded. New direction is required for any further isolation diagnostic, fix or host.
