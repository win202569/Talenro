# Windows unknown-usage native verification plan

> **For agentic workers:** Use superpowers:executing-plans sequentially in this session after explicit approval. Verification only; no implementation or agent delegation.

**Goal:** Observe the previously unrun C5 usage route through real existing controls.

**Architecture:** One preserved layout-diagnostic process, fresh ordinary state,
fixed unknown-usage profile. Observe Active -> Unknown -> Stale -> Active using
the existing entitlement selector and Account page; retain native exit evidence.
This verifies display refresh, not a hypothesis about the historical shutdown cause.

**Tech Stack:** PowerShell, existing WinUI output, supported computer-use interface,
existing native/managed process-observation modules. No new dependencies.

**Spec:** docs/superpowers/specs/2026-11-08-windows-layout-test-mode-design.md;
historical C5 in 2026-11-08-windows-layout-stage-b-continuation.md.

Status: user approved after cc565557; one launch consumed. Execution stopped before
the first sample at creation-time string mismatch; normal close sent and no targets
remained at subsequent checks. C5 NOT_RUN/NOT_ACCEPTED. See
../../roadmap/2026-11-09-windows-unknown-usage-native-execution.md. No retry allowance.

## Global constraints

- Proposed budget: one app invocation including failed startup, zero retries,
  builds, restores or test-suite runs. No source/artifact changes.
- Ordinary-user token; exact target identity; no pre-existing demo/prototype target.
- Startup within30s; request normal close by elapsed300s at the latest, even if
  incomplete. Normal exit criterion30s after close; absolute observation cap360s
  from launch. Check deadlines before each action; no waiting tool call over30s.
- No automatic text entry, clipboard access, connection action, node selection,
  account submission, network workflow, debugger/dump, force-kill or service change.
- No resize, locale/display/security-setting change, main merge or acceptance waiver.
  Task11/12 and44/34/6/217 remain NOT_ACCEPTED. RTL/C4 and default/C6 are excluded.

## Review focus

1. Wrong artifact: verify both apphost and managed DLL, not filename alone.
2. Fixture leaking into Active: record initial and final ordinary numeric demo values.
3. Unknown versus Stale: exercise both via controls, never inject model state.
4. Off-page activity: title must become inactive on Connection despite unavailable usage.
5. False exit: native retained-handle signal and valid exit evidence, not window disappearance.

## Task1 — Preflight without launching

Files consumed: clients/windows/tests/process-observation.psm1,
native-process-observation.psm1 and native-process-observation.cs; unchanged.

- [ ] Read AGENTS.md, the spec, the shutdown convergence report, this plan, and
  current computer-use instructions. Confirm explicit one-run approval and clean branch.
- [ ] Confirm ordinary token and zero existing demo/prototype processes/windows.
  Initialize supported UI observation before starting; unavailable tools block launch.
- [ ] Bind clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/
  net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe:
  EXE SHA2564180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95;
  sibling DLL7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB.
  Missing/drifted artifacts block, not authorize rebuilding. Verify observer source blobs.

## Task2 — One bounded route

- [ ] Launch once with --layout-test --layout-language=zh-Hans
  --layout-profile=unknown-usage. No --layout-close-run on this old artifact.
  Save fresh observer-only correlation, PID/path/args/creation time and original
  Process/SafeHandle. Approximately2s native-first/managed-second samples, as the
  recent preserved-artifact run; no historical5s outer-wait procedure substitution.
- [ ] Observe Login/offline notice, Enter demo, then Account: ordinary usage
  `12.4 / 100 GB（虚构示例）`, devices `2 / 5（虚构已激活设备）`, profile inactive.
- [ ] Navigate Connection, scroll to the existing entitlement selector and choose
  Unknown. Confirm selection; activity remains inactive on Connection. On Account,
  usage must be `不可用 / Unavailable — no numeric usage is known`, devices `不可用`,
  status Unknown/localized equivalent, title active. Record actual visible text.
- [ ] Repeat through Connection with Stale: inactive off-page, same usage/device
  unavailable values and active title on Account. No synthetic state injection.
- [ ] Return through Connection to Active, then Account: restore the exact initial
  numeric demo values, active status and inactive profile. No stale sample remains.
- [ ] Refresh observation after every action. Screenshot/text evidence must match
  the selected target; inability to verify a value yields UNVERIFIED, not inference
  from the title alone. Preserve clipping/mixed-language failures; do not repair.
- [ ] Request one normal titlebar close on completion/error or by300s. At30s without
  verified exit, mark shutdown failed/unresolved; observe only until360s, no more
  interaction/retry. Require native signal, successful code0 and valid exit time,
  correlate managed evidence and original creation identity. No stage claims.
  At cap, stop observer and report any live target; do not forcibly end it.

## Task3 — Evidence and handoff

- [ ] Create docs/roadmap/2026-11-09-windows-unknown-usage-native-execution.md with
  budget, identity, timings, each step's PASS/FAIL/UNVERIFIED/NOT_RUN and exit result.
  Preserve raw records/screenshots locally ignored; record raw hash, no private data.
- [ ] Check post-run process/window inventory and unchanged artifact hashes.
  A successful C5 route cannot clear historical shutdown or certify reliability.
- [ ] Update current-status.md, repository-recovery.md and the manual checklist
  only for facts observed. Commit explicit necessary docs, push existing development
  branch and verify remote SHA. Unused allowance expires at handoff.

## Self-review and approval

One already designed profile and existing controls; no repeated search/close
experiment. Negative and positive state conditions distinguish whether the actual
Account refresh removes/restores the sample. Exact viewport, RTL, accessibility,
hidden state and general shutdown acceptance remain excluded. Execute sequentially
only after the user approves this new one-launch scope.
