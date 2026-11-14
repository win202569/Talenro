# Ordinary manual-keyboard minimal follow-up

Date2026-11-14. Proposed observation protocol, awaiting approval. No code/test
changes, build/restore, application launch or new live allowance this turn.

## Scope and choice

Use the existing exact-bound ordinary entry/candidate unchanged. Primary missing
evidence: visible keyboard reachability/auto-scroll and one Manage devices
placeholder's keyboard close/focus return. Reset/re-entry is optional only when
time remains. Do not repeat display work except necessary navigation.

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
  activation. Focus-only passage over those account controls is permitted; never
  press Enter/Space unless the identified target is the approved placeholder,
  its Close, or separately announced Exit demo/reset.

## Single-owner input handoff

Before launch, confirm user can keep continuous attention and see prompts without
switching away from the demo during the keyboard block. If that is impractical,
do not promise this run can prove uninterrupted keyboard focus; stop planning
and agree on a usable observation arrangement rather than launching speculatively.

During preparation the agent may perform necessary Enter demo/Account navigation.
User alone resizes horizontal edges if convenient, keeping height unchanged.
One preparation reply is allowed. Then agent states that automated input has
ended; user reactivates the demo by clicking its titlebar once, without a double
click or resizing, and subsequently uses keyboard only. No content/scrollbar
clicks, wheel, repeated/held keys or chat replies during the keyboard block.

Agent issues one short prompt for one physical key at a time. User presses it
once, releases, and pauses; agent obtains read-only fresh screenshots and text.
No agent activation/click/press_key is interleaved during this block. Returning
to chat would contaminate the focus sequence: report that and stop the block,
do not infer where keys landed or silently restart it.

Use screenshots/visible outlines and scrolling as primary evidence. Accessibility
pane/window focused-element text alone neither passes nor fails focus behavior.
Do not use a snapshot acquired before resize/input to identify the next target.

## Bounded route

| Elapsed ceiling | Action / stop condition |
| --- | --- |
| 0–60s | Confirm visible ordinary window; necessary Enter demo, disconnected/lock-off Connection and Account. No repeated polling if window unavailable. |
| By90s | Optional horizontal resize and preparation handoff complete. If resize cannot be completed promptly, default-size route only; narrower geometry remains UNVERIFIED. |
| By150s | Establish a visible actionable-button focus with manual Tab prompts. Maximum12 individual Tabs to find a clear start; stop at an observed wrap/no progress, uncertainty, or time ceiling, whichever comes first. Do not count window/pane metadata as an established start. |
| By230s | From visible MFA (or earlier sidebar button), observe individual Tab steps through MFA, Manage devices, Language, Delete(focus only), Exit(focus only), then individual Shift+Tab steps back through Delete/Language to Manage. No pointer scrolling. Record automatic scrolling only if it actually occurs; at a roomy viewport it may be NOT_EXERCISED. |
| By260s | Only with Manage devices visibly focused, user presses Enter once. Observe one offline placeholder. Only with its Close visibly focused, Enter once to close; wait for settled rendering and inspect returned Manage outline. If Close focus unclear, do not send speculative Enter. |
| Before280s, optional | If primary checks succeeded and ample time remains, announce Exit-demo/reset action separately, observe Login/sidebar hidden and re-enter to disconnected/lock-off. Use existing safe navigation; end keyboard-only block first. Otherwise skip. |
| By300s | Agent announces end of user input and normally closes the observed window by titlebar X; if agent cannot act safely, user performs X. Keep original observer running to its bounded result; no retry. |

After each key allow a short settle interval and fresh read-only capture; at most
one additional settled capture if the first image is ambiguous. These are
observation reads, not additional key attempts. Any missed stage deadline or
uncertain input/outline ends the keyboard block with UNVERIFIED and normal close;
do not consume the rest of the budget trying arbitrary focus fixes.

## Result criteria

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
Record timestamps/key sequence, input ownership changes, observed labels/outlines,
per-item PASS/FAIL/UNVERIFIED/NOT_RUN, original session result and post-close
read-only inventory/hashes. No fabricated geometry receipts. Commit sanitized
record/status/recovery only; raw images/receipts/candidate remain ignored.

Approval of this protocol is not a launch approval. Next ask for one explicitly
bounded run only after the user confirms the uninterrupted-input arrangement.
No implementation or synthetic rerun is necessary for an observation-only
protocol unless a separate code change is later requested. Current live budget0.
