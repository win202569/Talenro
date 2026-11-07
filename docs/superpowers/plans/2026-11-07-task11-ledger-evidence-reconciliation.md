# Task11 Ledger Evidence Reconciliation Implementation Plan

> **For agentic workers:** Use superpowers:executing-plans for the approved Native/inline method. This is a documentation-only evidence task; do not substitute generic test, helper-import or runtime steps.

**Goal:** Produce one source-bound gap table and a concrete next-action disposition for the existing ledger path.

**Architecture:** Compare committed implementation/test definitions with saved reports and their source bindings. Keep source existence, historical fake checks, native observations and rollout authority separate. Produce no new runner or fixture.

**Tech Stack:** PowerShell read-only text/hash/Git inspection and Markdown.

**Spec:** [Approved staged design](../specs/2026-11-07-task11-ledger-validation-integration-design.md), approved after bd297480.

**Status:** Approved and executed documentation review. Stage1 complete; Stage2/3 BLOCKED. See ../../roadmap/2026-11-07-task11-ledger-evidence-gap-review.md. Conditional follow-up proposal is not applicable; no runtime authority granted. Commit/push preservation is checked at handoff.

## Global constraints

- Existing worktree `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`, branch `codex/c12-b01-task9-coordinator`, origin `https://github.com/win202569/Talenro.git` only.
- Zero Go builds/runs, test hosts, helper imports, native probes, Docker, registry/security queries, installs or permission changes. Read test definitions without executing or dot-sourcing them.
- No source/test/fixture/attribute edits. Preserve public schema4 refusal, historical failed RED, consumed identities and all original records.
- Do not recheck the completed 146-row metadata reconciliation or LF correction as new work. No WPR/exporter/cache investigation.
- Existing 44/34/6/217 and NOT_ACCEPTED remain unchanged. This plan grants no new observation or retry allowance.
- Only necessary documentation is committed/pushed. No raw secrets, binaries, ignored receipts or caches enter Git.

## Review focus

1. Old findings already repaired must not be reported as current defects: trace superseding evidence.
2. A fake check or evaluator returning EvidenceComplete must not be relabelled native evidence or workload acceptance.
3. Mismatched/missing source binding makes evidence inapplicable or uncertain, not a pass.
4. External safety disposition cannot be invented from user continuation, a clean Git tree or absent alerts in a capped query.
5. A blocked row must produce a finite stop/decision, not another self-authorized diagnostic loop.

## Task 1: Build the single evidence table

**Files:** Create `docs/roadmap/2026-11-07-task11-ledger-evidence-gap-review.md`; update this plan's checkboxes. No code changes.

**Interface:** Output exactly eight requirement rows, each with ID, requirement, current source reference/hash, saved report and recorded binding, evidence class, applicability, remaining gap and responsible role. Do not assign a real external owner without evidence.

- [x] Record current HEAD/branch, read worktree instructions, and inspect status. Preserve unrelated changes. Use `git -c safe.directory=D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator` for read-only Git commands.
- [x] Read `ordinary-ledger.cs`, `ordinary-ledger-session.cs`, `ordinary-ledger-receipt.ps1`, `ordinary-ledger-native-probe.ps1`, `process-native.cs`, `run.ps1`, and `matrix-check.ps1` under `docs/diagnostics/isolated-fuzz`; hash these seven files. Read the matching ledger test definitions for the requirements below without loading them.
- [x] Read the ledger implementation, final review, review-repair, initial-snapshot-fix, probe preparation/review/repair, inbox-policy, timing-coverage and stageb-readiness reports under `docs/roadmap/2026-10-17-ordinary-ledger-*.md`. Follow explicit superseding references only for an unresolved row. Use `current-status.md` as an index, not acceptance evidence. Missing portable evidence is a recorded gap, not permission to reconstruct it from private caches.
- [x] Fill G1 acquisition/ownership/retained identity; G2 sticky failure/cap/sampled count; G3 time/release/cleanup; G4 schema/diagnostic/legacy isolation; G5 serialization and pure evidence validation; G6 source/executable/run bindings; G7 raw capture/outer containment/durable finalization; G8 security disposition and exact runtime/rollout authority.
- [x] Give each row an evidence class (`source-only`, `historical-fake`, `historical-native`, `external-disposition`, or `missing`) and applicability (`bound`, `changed`, `unknown`, or `not-applicable`). If multiple classes apply, list each separately inside that row. Reported success without a usable binding remains historical, not a new pass. Where source changed, identify the relevant delta before proposing regression.
- [x] Check the eight rows against the approved design and the five review-focus risks. G1-G5 cannot close G6-G8. The table must explicitly distinguish a known missing test, a stale binding, an external dependency and a rollout prohibition.

## Task 2: Issue one finite disposition and preserve it

**Files:** Append disposition to the same report; update `docs/roadmap/current-status.md`, `docs/runbooks/repository-recovery.md` and this plan. Create a follow-up plan only in the eligible branch below.

**Interface:** Consumes G1-G8; produces either a concrete blocked handoff or one bounded follow-up proposal, never a launch instruction authorized by this plan.

- [x] If safety disposition or a safe outer runner remains unavailable, mark native/rollout work BLOCKED with the exact missing evidence and responsible role. Stop after the report. Do not draft an executable experiment packet, issue nonces or automatically start repair/environment work. Avoid repeated requests for generic approval as a substitute for missing evidence.
- [x] Otherwise, create `docs/superpowers/plans/2026-11-07-task11-ledger-targeted-verification.md` listing only uncovered or source-changed cases, exact suite/case names, source/host prerequisites, fixed maximum calls and stop rules. Historical passing suites are not blindly replayed. No speculative case count may be reported as approved. **Not applicable: blocked branch selected.**
- [x] If the follow-up is ready to propose native observations, cap it at the existing two questions (root then conditional child), with 5s/20s native and 120s outer ceilings, shared 500ms natural/5000ms cleanup limits. Require a later separately approved exact-input packet; never invoke a probe here. If exact inputs cannot yet be fixed, record that prerequisite instead of fabricating a runnable plan. **Not applicable: blocked branch selected.**
- [x] Review the report for unsupported claims; verify all local reference paths exist, every G1-G8 row is present, and no product/test file changed. Use `git diff --check` (expected exit0), `git diff --name-only` and `git status --short`; explicitly include newly created documentation when reviewing.
- [ ] Stage only the reviewed documentation paths, commit and push normally per AGENTS.md, then compare `git ls-remote origin refs/heads/codex/c12-b01-task9-coordinator` with local HEAD and require a clean task worktree (or disclose preserved unrelated changes). Report the disposition without claiming runtime acceptance.

## Completion and self-review

One complete gap report is the deliverable even if its conclusion is BLOCKED.
The conditional follow-up plan is not required when an external/safety prerequisite
is missing. Do not manufacture implementation work merely to avoid that result.
Execute both tasks in one approved pass without asking between read-only steps;
stop for genuinely new authority or inaccessible essential evidence.

Self-review: all Stage1 design requirements map to G1-G8 and the disposition.
Stage2/3 are expressly excluded. The five review risks have explicit source/report
checks; no runtime test is needed or claimed for a document-only review. No new
algorithm/API or generic framework is introduced. Review this plan before execution.
