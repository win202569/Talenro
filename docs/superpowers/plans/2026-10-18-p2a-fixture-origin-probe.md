# P2-A Fixture Origin Probe Plan

> **For agentic workers:** REQUIRED SUB-SKILL: superpowers:executing-plans after explicit approval. Native only, no subagents. Checkboxes describe a proposed diagnostic, not authority to run it.

**Goal:** Determine whether the unmodified fixture constructor can propagate Schema metadata without any product call.

**Architecture:** One isolated fixture-only characterization host compares shared default values with fresh typed-array Int32 controls. Original product/tests/failed packets remain untouched; no automatic repair follows.

**Tech Stack:** Existing ordinary Windows x64 PS7.6.5, built-in SDK reflection and typed arrays, PowerShell only; no installs or compilation.

**Spec:** [P2-A boundaries](../specs/2026-10-18-p2-adapter-boundaries-design.md), [captured failure site](../../roadmap/2026-10-18-p2a-input-mutation-diagnostic.md). Scope is fixture provenance, not product implementation or acceptance.

Status: PROPOSED_NOT_APPROVED. After c3e56d5b user requested further diagnosis;
this turn did readonly source/evidence inspection and documentation only.
No new probe source/packet/host, repair, reviewer or product execution.

## Source investigation and evidence boundary

Local New-TBRequest assigns its object-typed parameters directly to fields;
Schema defaults to Int32=1. DecoratedScalars intentionally attaches metadata to
numeric leaves before InputMutation. Last diagnostic establishes only that input
Schema already had a NoteProperty at TB-Leaf55,called from input collection160.
It did not record metadata identity before the two preceding product calls.

Official version-tagged source supports a concrete mechanism, not a historical
runtime identity claim:

- Constant defaults are stored in parameter metadata: [Compiler.cs around1684](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/parser/Compiler.cs#L1684).
- Default retrieval reads that value; binding uses a copying policy rather than an unconditional fresh box: [scriptparameterbinder.cs](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/scriptparameterbinder.cs#L130).
- Primitive assignment may retain the original object: [Binders.cs around1749–1842](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/runtime/Binding/Binders.cs#L1749).
- Numeric instance-member lookup can use the base-object identity; typed-array storage is also used internally to copy value types: [MshObject.cs around968 and1500](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshObject.cs#L968).

Inference: fixture default/backing reuse is a plausible source of cross-case
metadata. This is not proof that the exact prior Schema shared that reference,
that the product is innocent, or that cloning a fixture fixes the full suite.
Version-tag source was read online,not executed or installed; local binary build
identity was not independently proven from that public tag. No new local host
was launched.790common pins and4latest packet anchors were readonly intact.

## Global constraints

- Old allowance remains2/5consumed,3frozen;last diagnostic1/1spent;old reviewineligible.
- Proposed ONE NEW fixture-only ordinary host; no product file copied/imported/
  called, no original suite, extra helper probes, retries, fixes, mutants or review.
- Keep original seven-file delivery, six-file diagnostic, manifests,indexes/raw
  streams unchanged. Preserve all790mapped/current pins and4latest evidence anchors.
- Same worktree/dev branch/origin; no main/live/B/C/P3/PS5/security queries, imports,
  dependencies, type-data changes or cleanup of raw records.
- Explicit diagnostic scope overrides generic extra TDD/whole-project/repair/review
  cycles; native bounded PowerShell controller/ledger replaces Bash helpers.

## Review Focus

1. Shared backing is measured by SDK ReferenceEquals,not numeric equality.
2. Attached descriptor/payload must really persist and change; no no-op counted.
3. Existing B and newly constructed C distinguish retained views from new fixture reads.
4. Fresh D/E retain exact Int32/value1 while remaining independent and undecorated.
5. No-sharing/mixed/incomplete outcomes remain honest nonconfirmation,not a repair.

## Task 1 — one fixture-only characterization, after approval

**Proposed files, not created:** docs/diagnostics/isolated-fuzz/p2a-fixture-origin-probe/
fixture-origin.ps1,capture.ps1,source-pins.json,expected-output.json,README.md,
.gitattributes(LF). New ignored packet .superpowers/sdd/2026-10-18-p2a-fixture-origin-probe/.
Report docs/roadmap/2026-10-18-p2a-fixture-origin-probe.md;status/recovery/plan ledger.

**Consumes:** exact New-TBRequest function text from frozen tests
11BA52E114FC8F14D2D2A020FC6E6C70C52A3F2105B0DADE548416A483B71EB9,
not the rest of that script. **Produces:** bounded identity/metadata/control facts,
not a changed fixture constructor or product API result.

- [ ] **Step 1 — prepare without target execution.** Verify clean branch/origin,
 790pins+4latest raw anchors;freeze original constructor extent and exact diff.
 Copy ONLY that function definition verbatim into probe;do not dot-source or
 evaluate original script. Add test-only New-OriginFreshInt32([int]$Value) using
 one-element typed-array storage and fixed-SDK identity/descriptor observers.
 Use exactly FIVE constructor calls, each with Origin0/Frequency1000/Observed4/
 Stop2 Int64: A/B omit Schema, C omits Schema after mutation, D/E each receive a
 separately allocated fresh Int32=1. No loops creating additional fixtures.
- [ ] **Step 2 — freeze the exact characterization sequence.** Hold A/B Schema
 views before mutation,confirm exact Int32/value1 and initially plain views;
 record backing identity AB. Attach ONE NoteProperty P2AFixtureOriginProbe to A's
 held Schema view with one-element payload ['initial'];change payload once to
 'changed'. Prove descriptor identity,payload identity and retained mutation.
 Observe B held and fresh views;construct C and observe C held/fresh views and AC
 backing identity. Construct D/E with fresh integer controls;record independent
 backing versus A/B/C and each other,exact type/value,absence of probe metadata;
 recheck A's descriptor/payload still retained. Never clear metadata or coerce a
 wrong-type fixture. No product function exists in the probe source.
- [ ] **Step 3 — run ONE bounded ordinary host and stop.** Existing engine
 C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe,
 SHA256=362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139;
 -NoProfile -NonInteractive -File fixture-origin.ps1 only. Persist consumption
 before launch,max30000ms+2000ms owned exit+2000ms separate concurrent raw-byte
 drain;ordinary CONTEXT/HOST required,stderr0,intact pins,confirmed owned exit.
 Output exactly CONTEXT,HOST,one ORIGIN JSON(max64KiB),fixed LIMIT line. JSON freezes
 types/values,plain-before flags,AB/AC identity,B/C held/fresh descriptor+payload
 visibility,A-retention,D/E independence/plainness,and ProductImported=false,
 ProductCalled=false,RepairPerformed=false. All fixture/control validity checks
 must hold;AB+AC sharing and matching B/C descriptor/payload gives
 DEFAULT_REUSE_CONFIRMED(exit0). If controls hold but that full chain does not,
 DEFAULT_REUSE_NOT_CONFIRMED(exit2),including mixed sharing. Invalid fixture,
 missing evidence,drift,timeout/unknown cleanup gives INCOMPLETE(exit1 or controller
 failure). No outcome grants another host or an automatic source change.
- [ ] **Step 4 — readonly disposition and preservation.** If confirmed,describe a
 reproducible fixture-only pollution pathway,not exact historical attribution or
 global product clearance. Only then propose fresh numeric-fixture construction
 while preserving strict wrong-type/decorated-input/ownership assertions;any
 repair and regression/mutant validation needs separate scope. If not confirmed,
 retain the competing hypothesis and missing observations without reruns. Save
 necessary files/report/status/recovery by explicit list to original dev branch;
 confirm remote SHA equality and clean. Keep all ignored raw records.

Self-review: one task,one host,zero product calls,zero reviewer. Five focus items
map to identity,retention,held/fresh comparisons,fresh controls and exact outcomes.
No broad acceptance claim. Native method retained. Please approve/reject this
fixture-only diagnostic;this proposal does not reopen any frozen allowance.
