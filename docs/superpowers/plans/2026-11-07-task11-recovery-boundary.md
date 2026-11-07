# Task11 Metadata Registry Recovery Plan

> **For agentic workers:** Use superpowers:executing-plans for an approved continuation. Preserve Native/inline; this document does not authorize runtime or product implementation.

**Goal:** Resume useful Task11 preparation without restarting environment diagnostics or bypassing the stopped test gate.

**Architecture:** Keep the approved metadata design and implementation plan unchanged. Separate source-only preparation from native verification and subsequent implementation; a documentation checkpoint is not a test result.

**Tech Stack:** PowerShell read-only file/Git inspection and Markdown; no Go, Docker, exporter, registry access or new dependencies.

**Spec:** [Existing approved plan](2026-10-17-authority-v7-metadata-registry.md) and [actual stopped result](../../roadmap/2026-10-17-task11-metadata-registry-implementation.md). This is a recovery addendum, not a replacement protocol design.

## Current evidence — 2026-11-07

- Current checkout contains the initial `internal/nodecontrol/contracts/schema_registry_test.go` inventory test.
- Direct path checks confirm `internal/nodecontrol/contracts/schema_registry.go`, `testdata/c12/authority-v7/schema-registry.v1.json` and `testdata/c12/authority-v7/schema-registry-golden.v1.json` are absent.
- Historical fixture-red build succeeded; natural runtime shutdown was not verified. Missing-fixture output is not accepted RED. No fresh test was run for this recovery note.
- The earlier 146-row reconciliation is design evidence, not an implementation or passing-test claim. Do not create another registry design or duplicate that table merely to report progress.
- User reported STOP_ADMINISTRATOR_CONTEXT twice during manual environment handoff. These are user-reported stops, not independently observed ordinary-user comparisons. The environment/capture route is shelved; do not automatically ask for another attempt.

## Global constraints

- Worktree: `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`; branch: `codex/c12-b01-task9-coordinator`; origin: `https://github.com/win202569/Talenro.git`.
- Keep consumed attempts and original records unchanged. Remaining historical four-build/three-runtime slots are not continuation permission.
- No product/test/fixture edits under this recovery note. In particular, do not create GREEN fixtures or the metadata API before the existing RED prerequisite is satisfied or an explicit revised workflow is approved.
- No process-security exceptions, administrator retry, account/token/UAC/registry changes, capture, dependency installation or main merge.
- Overall counters remain 44/34/6/217; NOT_ACCEPTED. Task11 and wider C12 acceptance remain open.

## Review focus

1. Old progress entries must not override newer source evidence.
2. A successful forced cleanup must not be relabelled natural completion.
3. Static source review must not be called RED/GREEN or independent certification.
4. Existing 146-row reconciliation must be reused, not regenerated from implementation.
5. Clone recovery must not assume ignored binaries, caches or consumed control roots exist.

## Task 1: One bounded source-drift review — completed, source-only

**Files:** Read the existing plan and all specifications it explicitly lists, `docs/roadmap/2026-10-17-task11-final-metadata-reconciliation.md`, the inventory test, and four frozen name-only files. Record findings in this document; update current status/recovery only if findings change the next action.

**Interface:** Input is committed design/source; output is a short discrepancy list with exact paths and references, not generated fixtures or an executable controller.

- [x] Read the existing authoritative design chain completely, using its documented precedence. Do not re-open already settled design choices without a concrete contradiction.
- [x] Compare the four name-only file hashes against the original plan's frozen values and check inventory-test expectations against the existing 146-row reconciliation.
- [x] Report each mismatch, or explicitly report no mismatch in this limited comparison. Keep missing metadata implementation separate from design drift.
- [x] Stop at the report. No helper repair, native test, broad cache scan, generated fixture or performance investigation follows automatically.

### Source-only result at baseline 68b13226

User approved this bounded review. Read the six clarification/design documents linked by the original plan and the final reconciliation; retain their recorded local precedence, not older pending-count wording. This is not a fresh audit of every original v7 consumer or canonical registration.

**One raw-byte mismatch:** `testdata/c12/authority-v7/schema-name-registry.v1.json` has SHA256 `C0946C5813A1A93DFFC7B990C1CA45F6546D3AF4903C925B83EF5414245B12DB` in this worktree, not frozen `F505A358044D4F4292B97BE7003C54FC21E573276C3308B6E2055A755AD981E1`. `git ls-files --eol` reports `i/lf w/crlf attr/text=auto`; the current file contains 590 CRLF pairs. Replacing CRLF with LF in memory only produces the exact frozen SHA256. The discrepancy is accounted for by line endings, not changed JSON contents. No bytes, Git settings or expected pins were changed. The raw-byte prerequisite remains unmet; semantic equivalence is not a waiver.

The other three frozen files match their original SHA256 values exactly: `internal/nodecontrol/contracts/schema_names.go`, `internal/nodecontrol/contracts/schema_names_test.go`, and `testdata/c12/authority-v7/schema-name-registry-golden.v1.json`.

Read-only PowerShell parsing/comparison found:

- All 146 exact schema/object pairs in the final reconciliation match the production name table and each name fixture, with zero pair differences in each comparison.
- Final-table forms are 71 EXT, 27 DB and 48 empty; empty splits into 32 non-null and 16 null roles; 37 distinct non-null roles. These match the initial inventory test's constants at `schema_registry_test.go:55,58,106` (three framing vectors are required by the plan, not present metadata fixtures).
- All four metadata fields match the approved N9 table (9 rows), D16 table (16 rows), and R27 profile mapping (27 rows, 11 profiles). C1 #7 retains its dedicated request role/empty evidence; #16 retains null/empty. No design rewrite follows from this limited comparison.
- The inventory test at lines 45/48 still uses ordinary `json.Unmarshal`; strict required-key/duplicate/type rejection, framing validation and API/ownership tests belong to unimplemented later steps. Do not treat this partial test as complete metadata validation or a newly discovered production regression.

Task 1 is complete as a source-only review, not RED/GREEN, package acceptance or independent certification. Task 2 remains blocked. Before any future runtime proposal, the exact one-file LF restoration can be separately scoped and re-hashed against the unchanged original pin; do not normalize all files, rewrite fixtures or silently update expected hashes. Fixing this preflight mismatch would not resolve the historical natural-shutdown failure or authorize a test run.

Ruling: apply the approved documentation-only scope instead of generic implementation-skill test/reviewer/cleanup commands. Keep the result here as the durable ledger and use inline author review; no Bash bookkeeping, runtime, subagent or scratch deletion. Cost/limit: no independent review or executable correctness claim is obtained.

## Task 2: Runtime resumption decision — blocked, not executable

The metadata work is independent of WPR/exporter access, but its accepted native test path is not yet established. Before implementation, a separate scoped decision must identify a verification route preserving ownership, identity, natural-exit, cleanup and audit requirements, with fresh source bindings and explicit budget. This document neither selects an unverified alternate runner nor spends the old remaining allowance.

Once that decision exists, resume the original plan at its unsatisfied fixture-red prerequisite, not fixture-green. Preserve the original failed attempt and do not reuse its identity. Follow the original independent fixtures, malformed-input, framing, API ownership and complete-package test requirements. No default authorization to replay the failed phase is granted here.

If no safe verification route is available, report Task11 implementation as blocked. Do not replace it with another diagnostic loop or advance to unrelated product phases without direction.

## Documentation checkpoint and self-review

This turn only prepares the recovery boundary and corrects the latest handoff. Review an explicit three-file diff, run `git diff --check`, commit and push the necessary documentation under standing repository instructions, then compare remote/local commit IDs. No test execution is required or claimed for this documentation-only checkpoint.

Self-review: no change to the four-field API, 146 identities, metadata trust limits or original safety contract; all five review risks have explicit boundaries above. Task 1 was approved and is now complete as recorded above; Task 2 remains blocked, and product implementation has not resumed. Retain Native/inline and the explicit no-runtime boundary.
