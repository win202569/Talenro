# Windows offline demo gap-only verification plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans
> after approval. Preserve inline execution; no delegated implementation.

**Goal:** Observe remaining functional native checks without repeating accepted paths.
**Architecture:** One existing offline-demo process, ordinary user, explicit fixture
events only. Record each result independently; no code changes or automated tests.
**Tech Stack:** Existing WinUI3 executable, native UI observation, PowerShell.
**Spec:** ../specs/2026-11-07-windows-client-four-screen-design.md.

Status: proposed after user requested continuation following c76c9b0a; exact new
allowance awaits approval. Previous launch allowances remain consumed.

## Global Constraints

- At most ONE new launch, up to20minutes including startup and close observation;
  startup limit30seconds, ordinary-close observation15seconds. No restart/retry.
- Use only the same worktree executable and fixed fictional data. Check ordinary
  token, unchanged source/artifact and absence of existing demo processes first.
  Restricted process invisibility is not proof of absence; use authorized read-only
  ordinary-context checks. Stop if process identity cannot be established.
- No build/test/restore, source edits, install, external links, real accounts,
  services, network integration, or system/security/display/accessibility changes.
- On confirmed application defect, crash, hang, or elevation request: stop scenarios,
  attempt one ordinary close if targetable, verify exact process exit; no kill.
- At minute18 stop new scenarios and close within20minutes. Preserve unfinished
  items as UNVERIFIED. No extra launch to finish the list.
- No main merge; Task11/12 and44/34/6/217 NOT_ACCEPTED unchanged.

## Review Focus

1. Nonactive permission must actually refuse Connect, not merely display a notice.
2. Loading/unavailable/empty lists must not allow an invalid Apply.
3. Cross-country Cancel/default Enter must preserve preference and lock.
4. Recovery retry/stay must not silently cross countries or reconnect.
5. Logout must reset deliberately changed page-local selectors and search.

## Task 1: Remaining functional observations

Files: update docs/roadmap/2026-11-08-windows-client-native-verification.md and
clients/windows/tests/manual-checklist.md with PASS/FAIL/UNVERIFIED evidence.

- [ ] Check preconditions and launch exact executable once; record PID/path/time.
- [ ] With Auto/no known country, enable lock: refusal and lock remains off.
- [ ] For expired, exhausted, unknown, stale and device-limit: select scenario,
  actually press Connect, verify remains disconnected with correct reason; inspect
  Account values, including unavailable rather than zero and activated count.
- [ ] Restore active. In Nodes exercise loading/unavailable/explicit empty and
  disabled-row input; verify appropriate unavailable Apply and no invalid selection.
- [ ] Select JP, connect explicitly, lock JP. Draft DE and Apply: confirmation
  names JP/DE and lock change. Default Enter cancels; inspect JP preference/lock.
  Repeat with Cancel button and inspect preserved state. Confirm separately:
  Connecting, saved DE preference/lock; do not infer hidden atomicity from UI alone.
- [ ] Explicit locked-country failure: Retry this country enters Connecting with
  same lock; fail again, Stay disconnected leaves disconnected and clears offer.
- [ ] Exercise remaining registration/sidebar/account placeholder explanations
  without real account operations. Inspect focus return and selected keyboard paths.
- [ ] Change search/list scenario and then logout; reenter to inspect default
  list scenario and empty search. Change login scenario before entry, return normal;
  logout and inspect normal selector. Do not repeat process-reopen verification.
- [ ] Check pointer scrolling within observed content. Keyboard traversal is
  limited to controls exercised here, not full screen-reader acceptance.
- [ ] Close normally, verify exact process exited and record elapsed time.

## Task 2: Preserve bounded result

- [ ] Reconcile every step against observed evidence, leaving all gaps explicit.
- [ ] Update current-status.md, offline-demo-execution.md and repository-recovery.md;
  review explicit file diff, commit/push existing development branch and compare SHA.

## Deferred, not hidden acceptance

Screen reader, measured contrast, exact target logical dimensions, enlarged text,
RTL/long strings, English fallback, final design comparison and security/network/
storage auditing remain separate scope. Do not change settings to obtain coverage.
Self-review: remaining functional gaps map to Task1; inaccessible environmental
checks remain explicitly deferred. This is verification-only, not a TDD/code plan.
