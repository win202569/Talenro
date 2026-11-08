# Windows Layout Stage B Continuation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to execute this verification-only plan sequentially. No implementation or new TDD/reviewer cycle is included.

**Goal:** Observe the remaining English dialog route and four synthetic profiles after the switch fix, followed by a no-activation session.

**Architecture:** Reuse existing hash-bound diagnostic output, existing views and simulated controls. One process at a time; preserve earlier evidence rather than rerunning completed switch checks or rejection rows.

**Tech Stack:** PowerShell, existing .NET10/WinUI application and supported native UI tools; no new dependencies.

**Spec:** docs/superpowers/specs/2026-11-08-windows-layout-test-mode-design.md.

Status: approved after98434867; execution stopped after4/6 starts at C3 shutdown failure. C4 was prematurely started before C3 exit confirmation and then normally closed without profile validation; C5/C6 NOT_RUN. See docs/roadmap/2026-11-08-windows-layout-stage-b-continuation-execution.md. Unused slots do not carry over; historical approval wording below is superseded.

## Global constraints

- Maximum6 app invocations, including failed starts; no retry or replacement. Zero builds, restores, suite runs, installs or source changes.
- Ordinary-user token before every launch; stop on administrator token or uncertain ownership. No UAC/firewall/locale/display/text-scale changes, resize, real account/network actions or main merge.
- Startup30s, interaction10minutes per process, normal close30s. Poll at most30s per wait; retain process handle/PID/start/actual exit. On unexpected errors, wrong semantics, mixed ordinary UI language or clipped required content, stop dependent sessions and close normally. No force-kill.
- Fixture text is intentionally synthetic, sometimes mixed-language; compare it to the committed sample rather than treating its designed language mix as a localization defect.
- Task11/12 and44/34/6/217 remain NOT_ACCEPTED.226/226 is historical and is not repeated.

## Review focus

1. Wrong artifact: Task1 binds managed DLL and apphost, not apphost alone.
2. Dialog language/state: C1 checks connected cross-country confirmation and Cancel.
3. Conditional samples leaking into normal states: C2/C5 check entry and exit conditions.
4. Presentation changing identity/search/bidi semantics: C3/C4 inspect stable identifiers and original search behavior.
5. Profile persistence or exaggerated measurement claims: C6 checks fresh defaults; all rows keep independent size and accessibility limitations explicit.

## Task1 — Read-only preflight

Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; branch codex/c12-b01-task9-coordinator. Read AGENTS.md, spec, prior switch/native ledgers, UI instructions and clients/windows/tests/ui-copy-cases.json.

- [ ] Verify approval, clean expected branch, ordinary token, no existing demo process/window.
- [ ] Resolve clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe. Compare sibling DLL SHA256 to7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB and EXE to4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95. Mismatch/missing output blocks; it does not permit rebuilding.
- [ ] Initialize docs/roadmap/2026-11-08-windows-layout-stage-b-continuation-execution.md with source1ac1c19e, current documentation HEAD, identities and6-slot budget. Supported UI facility must be usable before launch.

## Task2 — Six sequential sessions

All use Task1's diagnostic executable directly, never dotnet run. Each starts at Login. Use only synthetic Enter demo and existing scenario controls; normal close and exit0 before the next row. Prefix C1–C5 arguments with --layout-test.

| Row | Remaining arguments | Required observations |
| --- | --- | --- |
| C1 | --layout-language=en-US --layout-profile=baseline | English Login/Connection/Nodes/Account, permanent offline notice. Explicit simulated Connect then simulated success; select de-1 while connected, Apply opens English cross-country confirmation; Cancel preserves existing connected selection. Account logout/re-entry resets disconnected state with English retained. Sample relevant Tab focus without activating scenario buttons accidentally. |
| C2 | --layout-language=zh-Hans --layout-profile=long-status | Initial ordinary disconnected status and inactive profile; explicit Connect enters Connecting with exact long-status fixture and active title. Cancel leaves Connecting, restores ordinary status and inactive title; action meaning/enablement unchanged. |
| C3 | --layout-language=zh-Hans --layout-profile=long-country | JP group and jp-1 row use exact long-country fixture (row retains jp-1 suffix); other rows unchanged. Search original jp-1/JP identity using existing search, select jp-1/apply while disconnected, then simulate connected. Choose de-1 and Apply: cross-country dialog retains real country semantics; Cancel preserves selection. |
| C4 | --layout-language=en-US --layout-profile=rtl-mixed | Login heading uses exact rtl-mixed fixture and root mirrored. Actual demo email on Login/Account and technical identifiers retain LTR presentation. Navigate Login/Connection/Nodes/Account; logout/re-entry retains chosen profile and resets model. Synthetic bidi stress is not natural-language shaping certification. |
| C5 | --layout-language=zh-Hans --layout-profile=unknown-usage | Via existing Connection entitlement selector, choose Unknown then Account: exact unknown-usage fixture. Repeat Stale: same fixture. Return known Active: ordinary usage value restored. Title activity follows Account target condition; no model state injection. |
| C6 | none | After C1–C5, fresh Chinese ordinary Login/state without diagnostic marker or profile samples; navigate Connection/Nodes/Account and close normally. No repeat of full switch regression needed. |

- [ ] For each row, record exact arguments/PID/start/exit, pages/state transitions, visible text and any clipping, selected language/profile marker and conditional activity. Compare fixture text to JSON including bidi controls through available text evidence; if tooling cannot establish exact control characters or direction, mark that subcheck UNVERIFIED rather than inferring it from title.
- [ ] Scroll existing panes before interacting with offscreen controls. Refresh after each input, inspect before the next input. Unsupported input targeting is an evidence limitation, not permission for process retries or new helpers.
- [ ] Require offline notice and unchanged fictional identity throughout. Preserve screenshots in the session only; do not put private screenshots/raw dumps in Git.
- [ ] Record title values as self-reported only. No independent client/DPI tool currently exists in this workflow: exact sizes, resize updates, enlarged text, screen-reader speech, full keyboard/accessibility and contrast remain UNVERIFIED. Do not stop visual checks merely for these planned limitations.

## Task3 — Preserve results

- [ ] Record each C row PASS/FAIL/UNVERIFIED/NOT_RUN, including partial results; do not convert a partial row into full acceptance. Check no owned demo process remains or report its PID.
- [ ] Update execution ledger, docs/roadmap/current-status.md, docs/runbooks/repository-recovery.md and clients/windows/tests/manual-checklist.md only for observations actually obtained. Old R1–R6 evidence stays historical; transient-window absence is not newly proved.
- [ ] Review explicit documentation diff, run git diff --check, commit/push existing development branch and verify remote SHA equals HEAD plus clean status. Exclude binaries/caches/raw receipts. Unused slots expire at handoff.

## Self-review / handoff

Task1 identities feed Task2; Task2 observed results feed Task3. This is a bounded continuation of existing design, not new functionality. C1 completes the interrupted English route, C2–C5 cover each remaining profile once and C6 checks defaults after those profiles. Full language/profile cross-product, independent measurement, resize and accessibility certification are deliberately excluded and must stay pending. Retain sequential execution by the current implementer. Please approve the6-launch ceiling before runtime work.
