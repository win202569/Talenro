# Ordinary manual-keyboard minimal follow-up

Updated2026-11-15: user requested maximum60-second phone acknowledgement waits
and fewer checks per run. This document records that revision; no launch approval
is implied. No code/test/build/restore/launch this turn. Previous allowances were
consumed; their original records and subsequent user clarifications remain intact.

## Scope and choice

Use the existing exact-bound ordinary entry/candidate unchanged. Next proposed run
is limited to acknowledged titlebar activation plus ONE physical Tab and its
post-ack visible focus observation, then normal close. Even if time remains, no
second key, traversal sequence, dialog, reset/re-entry or optional resize is added.
Full forward/reverse traversal, auto-scroll and placeholder focus return remain
separate outstanding acceptance items, requiring separately scoped future work.

No-app timing sample on2026-11-15: clock before prompt07:43:12 UTC
(11:43:12 Asia/Tbilisi), first clock read after reply07:43:42 UTC: about30s.
This includes prompt generation/delivery, user reaction/reply and agent scheduling;
it is not network latency or an exact server message-arrival timestamp. One sample
does not establish an upper bound. User chose60s to add response margin, without
extending the six-minute process cap or weakening any identity/safety evidence.

Recommended manual-input/read-only-observer approach: user sends the physical
keys, agent observes rendered results. Automated Tab was inconclusive in the
previous run; repeating it would not isolate input delivery. Adding focus code or
diagnostic instrumentation would change the tested artifact and is out of scope.
This protocol does not promise the manual approach will establish focus.

## Unchanged binding and safety gates

- Entry clients/windows/tests/observe-ordinary-account-center.ps1; no parameter
  changes except its existing RunApproved switch after separate live approval.
- Exact one terminal U+0020, Ordinal comparison, pinned launcher versions/hash,
  six candidate hashes and build source remain per the approved
  [binding design](2026-11-14-windows-ordinary-exact-command-line-binding-design.md).
- Ordinary token, authorized host execution with CIM access, branch/clean-tree,
  product/source/loaded-command/artifact/reparse/zero-target checks unchanged.
- Same original Process/handle/native/CIM observation and immutable new receipts;
  no fallback, tolerance, replacement process, force-kill, dump or service change.
- Future approval: one launch, zero retries,360000ms outer cap, max sleep2000ms;
  begin normal close by300s, earlier on failure or lost cooperation. Waiting for
  replies counts against the same budget. No extension or second launch.
- No connections, Nodes, credentials, system settings, MFA changes or Delete
  activation. This reduced route permits one Tab only: no Enter/Space, dialog
  activation or Exit-demo/reset. Normal titlebar X remains the shutdown action.

## Single-owner input handoff

### Approaches and selected channel

User selected phone confirmation in this same conversation. Unlike elapsed-time
inference, explicit acknowledgement distinguishes waiting from completed input.
Desktop chat replies would interrupt tested focus and are not the selected route.
No app instrumentation, new hotkey listener or automation input is introduced.

Before any separately approved launch, perform a no-app channel check: ask for
"手机准备完成" from the phone and wait for receipt in this conversation. A claim
that the phone is available is not evidence that its replies reach this observer.
If delivery fails, stop before launching; do not install/configure tools or switch
to desktop replies. No screenshot or sensitive data needs sending from the phone.

During preparation the agent may perform necessary Enter demo/Account navigation.
Keep default geometry; no resizing in this reduced route.
Then agent states that automated input has ended. Separate step0: user clicks the
demo titlebar once (no double-click/resize) and replies from phone "第0步完成".
Only after this acknowledgement may step1 request a key. No content/scrollbar
clicks, wheel, held keys or desktop chat replies during the keyboard block.

### One outstanding numbered step

Agent issues exactly one numbered action, for example "第1步：只按一次Tab，松开；
然后用手机回复‘第1步完成’，不要再按键". A chord such as Shift+Tab is one action.
User presses/releases, then replies on phone; a completion message is a user
report of input, not proof of successful focus movement. Agent does not issue
stepN+1 before receiving matching acknowledgement AND inspecting stepN results.

Each prompt states a response deadline: maximum60 seconds from issuance,
within the stage deadlines below. Reserve the full60s before issuing a prompt;
if insufficient time remains, do not issue it and close normally. Record the
issuance time and deadline; do not reset the timer for clarification or tool work.
This coordination policy adds no runtime allowance. User may phone-reply "未操作",
"需要暂停" or "停止"; never ask them to hurry or batch multiple keys.

While awaiting acknowledgement, monitor the retained session/time in bounded
calls, not a final handoff that abandons the live observer. Screenshot absence of
change is NOT input failure and must not cause early termination by itself.
Safety failures, user stop, original process exit and deadline still take priority.
No acknowledgement by deadline: announce the step cancelled, no further keys,
then normal close; no reminder that repeats the key instruction or retry.

Accept only the current step number. Duplicate, old, future, ambiguous "完成",
or late replies never advance the sequence or cause a repeated key. Clarify the
number via phone within the original deadline; do not extend it. Any cancelled
step stays cancelled. If input was duplicated, misdirected or desktop focus was
lost, stop the keyboard block and record contamination; no silent restart.

After matching acknowledgement, allow a short render settle then take a fresh
read-only screenshot (at most one further settled capture if ambiguous). Only
these post-ack captures support that step's focus judgment. No agent activation,
click or key is interleaved. Phone replies alone do not establish button focus.

Use screenshots/visible outlines and scrolling as primary evidence. Accessibility
pane/window focused-element text alone neither passes nor fails focus behavior.
Do not use a snapshot acquired before resize/input to identify the next target.

## Bounded route

| Elapsed ceiling | Action / stop condition |
| --- | --- |
| 0–60s | Confirm visible ordinary window; necessary Enter demo, disconnected/lock-off Connection and Account. No repeated polling if window unavailable. |
| Issue step0 by60s; complete by120s | One titlebar click then matching phone acknowledgement, up to60s from prompt. No key yet. Timeout cancels route; no inference of input failure. |
| Issue step1 by150s; complete by210s | Only after step0 acknowledgement and a fresh target observation: request ONE Tab, release, then matching phone acknowledgement. Wait up to60s. If target evidence is unreliable, stop rather than send a key. |
| By240s | After matching step1 acknowledgement, settle and read fresh target screenshot; at most one extra capture if unclear. Record visible focused control or UNVERIFIED. No further key, even if focus is absent. End input and start normal close promptly. |
| By300s | Agent announces end of user input and normally closes the observed window by titlebar X; if agent cannot act safely, user performs X. Keep original observer running to its bounded result; no retry. |

The stage ceilings remain upper bounds, not a promise all checks fit with phone
round trips. Do fewer steps rather than shorten acknowledgement waiting, batch
keys, extend the run or relax evidence. Missed deadline/uncertain post-ack outline
ends the keyboard block with normal close. No arbitrary focus fixes.

## Result criteria

- Reduced-route success means only one acknowledged Tab followed by a clearly
  visible actionable focus outline. It is NOT complete keyboard traversal,
  auto-scroll or dialog acceptance. Those deferred checks remain NOT_RUN in this
  route even if their controls happen to be visible. No-focus alone is UNVERIFIED,
  not evidence of a reproduced product defect.

- User explicitly confirms no input occurred: NOT_RUN. Missing acknowledgement alone means input
  completion UNKNOWN, UI result UNVERIFIED (never assume it was or was not pressed).
  Confirmed input with ambiguous focus: UNVERIFIED, not automatically FAIL.
  Only observed behavior supports PASS/FAIL; preserve later user clarifications.

- Keyboard traversal PASS only for the visibly recorded forward/reverse sequence;
  record start point and each label. Auto-scroll PASS only if lower controls were
  revealed by key navigation alone; do not infer it from scrollbar metadata.
- Dialog PASS only if keyboard activation, visible dialog Close focus, keyboard
  closure and visible returned Manage outline are all observed. Mouse closure
  cannot substitute; one placeholder is not all dialogs.
- Reset PASS only for observed visible Login/sidebar and re-entry initial state;
  not all persisted state. Otherwise NOT_RUN.
- Native exit judged by original retained session; preserve persisted pre-cleanup
  versus returned finalized results. Later inventory0 is supplemental, never a
  substitute for ExitVerified. Failures stay failures regardless of UI results.
- Exact root/client/DPI/1024x768/1440x1024/RTL, screen readers, security/Release/C4
  remain outside this follow-up. Screenshot bounds are not root DIP.

## Handoff and evidence

Preserve prior ordinary/diagnostic successes and failures without rewriting them.
Record each prompt number/key, issuance/deadline, matching phone receipt or timeout,
post-ack screenshot order, input ownership changes, observed labels/outlines,
per-item PASS/FAIL/UNVERIFIED/NOT_RUN, original session result and post-close
read-only inventory/hashes. No fabricated geometry receipts. Commit sanitized
record/status/recovery only; raw images/receipts/candidate remain ignored.

Approval of this protocol is not a launch approval. Next ask for one explicitly
bounded run only after the user confirms the uninterrupted-input arrangement.
No implementation or synthetic rerun is necessary for an observation-only
protocol unless a separate code change is later requested. Current live budget0.
