# Windows Layout Native Stage B Verification Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to execute this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Observe the existing ordinary/diagnostic outputs under bounded native sessions without changing implementation or claiming wider acceptance.

**Architecture:** One implementer in this chat runs sequential sessions against hash-bound existing artifacts. Each process invocation consumes one slot, including rejected startup; existing views and simulated actions provide observations, not new fixtures or state injection.

**Tech Stack:** PowerShell, existing WinUI Windows demo binaries, available native UI observation tools. No installation or new dependencies.

**Spec:** docs/superpowers/specs/2026-11-08-windows-layout-test-mode-design.md.

Status: APPROVED after508475b0; executed9/14 invocations and stopped at V3 English mixed-language switch state. See docs/roadmap/2026-11-08-windows-layout-native-stage-b-execution.md. Five unused slots are not standing permission after handoff. Historical proposal/confirmation wording below is superseded. Prior sequential execution method retained. This is a verification-only plan: no implementation or TDD cycle required.

## Global Constraints

- EnableLayoutDiagnostics defaults false in every configuration, including Debug. Release with it enabled fails the build.
- Startup requires --layout-test; language zh-Hans|en-US; profiles baseline|long-status|long-country|rtl-mixed|unknown-usage.
- No OS locale/display/text-scale changes, installs, real network actions, persistence or external fixture paths.
- No Core reducer/events, account identity, offline-warning meaning, control enablement or selection changes.
- All sessions use an ordinary-user token; stop on administrator token or unclear process ownership. Never change UAC/firewall/timeouts.
- Do not rebuild, restore, run automated suites, patch code, launch Narrator or merge main. Task11/12 plus44/34/6/217 remain NOT_ACCEPTED.
- At most14 process invocations, sequentially, no retries/replacements. A failed start consumes its slot. Normal closing does not authorize relaunch.
- Per process:30seconds startup observation, at most10minutes interactive checks,30seconds normal-close observation. Rejection processes must exit within30seconds. No automatic force-kill; if normal close fails, report owned PID and stop for direction.
- Observe/wait in intervals no longer than30seconds; report meaningful results and keep the user informed during longer sessions.

## Review Focus

1. Stale/wrong variant: bind full path plus managed DLL hash and apphost hash before Task1 launch.
2. Missing activation or malformed options: Task2's rejection rows must exit2; UI absence alone is insufficient.
3. Mixed language after navigation/refresh/logout: Task3 baseline rows include those transitions and one confirmation dialog.
4. Profile text affecting semantics: Task3 checks exact conditional targets, stable JP/jp-1 identity, ordinary actions and LTR account text.
5. Rounded/stale dimensions or unavailable client measurement: Task3 records independent evidence or explicitly leaves exact-size acceptance UNVERIFIED.

## Files and artifacts

Read: Stage A execution ledger, design above, clients/windows/tests/manual-checklist.md, tests/ui-copy-cases.json under clients/windows, and existing App startup/diagnostic source as needed.
Create during execution: docs/roadmap/2026-11-08-windows-layout-native-stage-b-execution.md (portable result table, not raw session dump).
Update during execution: docs/roadmap/current-status.md, docs/runbooks/repository-recovery.md, clients/windows/tests/manual-checklist.md. No source changes.

Worktree: D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator.
Under clients/windows/Talenro.Demo.App, ordinary output O is bin/Debug/net10.0-windows10.0.19041.0/win-x64; diagnostic D is bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64.
Executable: Talenro.Demo.App.exe; managed file: Talenro.Demo.App.dll.
Expected O DLL SHA256: C39E19BC4700DA25CFA6EC90A696E8627C922E7FD8B026BD08EB2A29D5F36C8E.
Expected D DLL SHA256: C16A8479C6BD6826BDD649651645C28A49A772AE6B7BD9CC8DCBDB38C0D96C63.
Both apphosts SHA256: A32B55694C5C5BC380474C6E2F66F750A1DD9F7C9780D4E141CA05325975FF71.
These expected hashes are historical until Task1 reads them afresh. Matching apphost alone cannot identify the variant.

## Task1 — Read-only preflight (zero launches)

- [ ] Confirm approval of this finite plan, current branch, clean worktree and ordinary token. Read applicable UI tool instructions before UI use.
- [ ] Confirm no pre-existing demo process/window; do not close or kill an unknown process. Resolve exact artifact paths and compare Get-FileHash SHA256 values above. Missing/mismatched output blocks execution, not permission to rebuild.
- [ ] Confirm UI inspection/control is available; record current monitor/work-area and display/text-scale evidence if readable without settings changes. Missing independent client/DPI capability only blocks exact-size claims, not visual smoke checks.
- [ ] Initialize result ledger with remaining14 slots, source7e25a4fc plus current documentation commit, artifact paths/hashes and preflight result. No inference from earlier native sessions.

## Task2 — Rejection matrix (6 invocations)

Use exact fixed argument arrays below. Start the resolved executable directly, never through a build/run command or shell-string interpolation. Capture PID, start time and actual exit code from the process handle. For Windows helper processes use hidden windows; intentional interactive demo windows in Task3 are visible. If stderr capture is available, preserve the fixed message and verify no supplied marker is echoed; unavailable output is UNVERIFIED, not an invented success.

| ID | Variant | Exact arguments | Expected |
| --- | --- | --- | --- |
| R1 | O | --layout-test | exit2, no demo window |
| R2 | D | --layout-language=en-US | exit2, no demo window |
| R3 | D | --layout-test --layout-test | exit2, no demo window |
| R4 | D | --layout-test --layout-unknown=stage-b-marker | exit2, no demo window; marker not echoed |
| R5 | D | --layout-test --layout-language | exit2, no demo window |
| R6 | D | --layout-test --layout-profile=unsupported | exit2, no demo window |

- [ ] Execute R1–R6 in order, recording each before starting the next. Fixed rejection text is 'Layout diagnostic startup rejected.'. No window observed during bounded observation is not proof against an undetectably brief transient; retain that limitation.
- [ ] Stop all further launches on wrong exit, unexpected visible window, hanging process, ownership ambiguity or other unexpected error. Preserve original evidence; no silent retry or repair. This samples rejection categories, not every parser permutation.

## Task3 — Visible session matrix (8 invocations)

Every activated row explicitly includes --layout-test. Each starts from Login, no state seeding. Use existing simulated controls; do not submit real credentials or invoke outside services. Close normally and confirm recorded process exit before proceeding. App close exit0 expected.

| ID | Variant | Exact arguments | Required observations |
| --- | --- | --- | --- |
| V1 | O | none | Ordinary Login, offline notice, login/navigation through all four pages, no diagnostic title/extra controls; normal close |
| V2 | D | --layout-test --layout-language=zh-Hans --layout-profile=baseline | Chinese shell/four pages, dynamic notices and cross-country confirmation buttons; cancel dialog, logout/login, text retained; title resize check |
| V3 | D | --layout-test --layout-language=en-US --layout-profile=baseline | Same baseline route in English; refreshed labels/dialogs consistent; no mixed-language claim beyond observed controls |
| V4 | D | --layout-test --layout-language=zh-Hans --layout-profile=long-status | Explicit Connect enters Connecting: exact fixture at ConnectionStatus, title active; Cancel leaves Connecting: ordinary status, inactive title, action meanings intact |
| V5 | D | --layout-test --layout-language=zh-Hans --layout-profile=long-country | Nodes JP group/jp-1 row show fixture with stable row suffix; select/search using original labels, open and cancel cross-country confirmation, IDs/selection semantics unchanged |
| V6 | D | --layout-test --layout-language=en-US --layout-profile=rtl-mixed | Login fixture exact, root mirrored; Login/Account email and technical IDs LTR; navigate four pages and logout, heading survives reset |
| V7 | D | --layout-test --layout-language=zh-Hans --layout-profile=unknown-usage | Existing Unknown then Stale scenarios show exact UsageValue fixture; known state returns ordinary value; title activity follows condition |
| V8 | D | none | After active sessions, fresh ordinary Login/state, no diagnostic title or persisted language/profile; ordinary navigation/close |

- [ ] Execute V1–V8 in order. Record visible text/clipping/focus anomalies and available screenshots per page/state; no automatic resize or edits to hide failures. Traverse relevant controls with Tab in baselines; record focus visibility rather than claiming screen-reader correctness.
- [ ] In V2 only, manually resize within the current work area once, then navigate page/state and observe title updating without relaunch. Record before/after root/page dimensions, render/text scales, timestamp and visible full-client availability. No screen-resolution, text-scale, monitor or locale changes. Other rows record dimensions at their natural size.
- [ ] Independently read native client physical bounds/effective DPI only through an already supported read-only observation facility. Do not invent/inject a new measurement helper in this plan. Check chrome/root-client correspondence before comparing root to client. PageHost has no independent bound proof unless separately observed.
- [ ] If independent measurement exists, compare physical bounds converted by DPI/96 to corresponding logical dimensions with tolerance ±0.5 DIP per axis. Include the title's ±0.0005 DIP rounding interval entirely inside tolerance. Record actual dimensions, not a requested size. Exact1440x1024 acceptance requires full client available plus independently corroborated target; otherwise UNVERIFIED. Screenshot pixels/outer bounds/title alone cannot establish it.
- [ ] Stop remaining launches for unexpected startup/UI errors, wrong semantics, clipped required content, missing fixture or failed normal close. A planned evidence limitation (no independent size tool, inaccessible target or unavailable speech/contrast evidence) is UNVERIFIED, not a reason for retries or settings changes.

## Task4 — Portable handoff

- [ ] Record each slot as PASS/FAIL/UNVERIFIED/NOT_RUN with arguments, variant hashes, PID/start/exit, pages/states, expected/observed result and evidence limitations. Never collapse partial compound checks into PASS.
- [ ] Report consumed/remaining slots and no owned demo process remaining (or the exact unresolved process). Unused slots do not become a standing allowance after handoff.
- [ ] Update current status, recovery and manual checklist only for observations actually obtained. Reader speech, contrast, enlarged text, Windows resource fallback, unobserved language/profile combinations and exact targets without independent evidence remain UNVERIFIED. No broader acceptance or main merge.
- [ ] Review explicit documentation file list, commit/push existing development branch under AGENTS.md; verify remote SHA equals local and clean status. Keep raw receipts, screenshots containing private data, binaries and caches out of Git.

## Self-review and handoff

Coverage: activation/defaults→R1/R2/V1/V8; malformed/duplicate/unknown→R3–R6; languages/refresh→V2/V3; four exact profile conditions→V4–V7; measurement/lifecycle→V2 plus normal close on each row. Five review-focus risks have explicit checks. Invalid-fixture startup and all option permutations remain source-test evidence, not native claims. Four profiles use one language each, not a full cross-product.

Recommended execution: sequentially by the current implementer, retaining the earlier method. Please confirm this14-invocation ceiling and scope before execution. No independent implementation review is added because this plan changes no implementation; any defect requires a separate repair scope and verification allowance.
