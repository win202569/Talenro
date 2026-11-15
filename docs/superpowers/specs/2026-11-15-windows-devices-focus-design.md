# Minimal Manage-devices focus observation and remaining scope

Date2026-11-15. Proposed observation only; awaiting review/separate launch approval.
No code/test/build/restore/launch. Prior run allowances exhausted.

## Coverage and evidence needed

| Item | Existing ordinary evidence | Still required |
| --- | --- | --- |
| Settings -> MFA | PASS, run813ce5d94c374895822a7bbb6abb2987, visible outlines after acknowledged Tabs | Re-establish starting route in any new process; never reuse old focus |
| MFA -> Manage devices | Not observed | One further acknowledged Tab with visible Devices outline; next minimum |
| Language/Delete/Exit forward reachability | Not observed | Separate visible key transitions, Delete focus only, no activation |
| Reverse traversal | Not observed | Individually confirmed Shift+Tab transitions with actual labels |
| Automatic scrolling | UNVERIFIED | Before/after view showing keyboard-only reveal of a previously clipped control; metadata alone insufficient |
| Manage devices placeholder close/focus return | NOT_RUN | Visible Devices focus, keyboard activation, visible Close focus, keyboard closure, visible returned Devices outline |
| Reset/re-entry and exact geometry/accessibility/broad acceptance | Separate outstanding scope | Not silently included in these checks |

Source Placeholder_Click awaits ContentDialog then requests Programmatic focus on
the sender. That is intent, not proof of a keyboard-visible return outline.
Diagnostics-binary evidence must not replace ordinary-candidate observation.

## Approach and tradeoff

Recommend only one additional transition beyond the proven route: third Tab to
Manage devices, no activation. Full traversal/scroll/dialog together cannot be
guaranteed within300s close deadline when each confirmation may take60s. Clicking
a target to shorten the route would not prove keyboard reachability and may
activate it, so is not part of this plan. Future dialog or lower-control checks
need their own bounded starting-state/step plan; this document does not authorize
a chain of launches or promise complete acceptance within a fixed number of runs.

## Gates and maximum scope

Inherit all exact launcher/command-line/source/six-artifact binding, ordinary
token, host/CIM, clean worktree, zero-target, path/reparse, immutable receipt and
retained-process exit checks from the existing ordinary observer and
[phone protocol](2026-11-14-windows-ordinary-manual-keyboard-followup-design.md).
One launch, zero retries,360000ms total, normal close by300s; max observer sleep2000ms.
No code changes, alternative launcher, fallback matching or security changes.

Asynchronous numbered phone questions, one outstanding step,60s maximum each;
same-conversation readiness before launch. Matching acknowledgement then fresh
read-only image before next step. No agent input during keyboard block. No desktop
chat replies, batched/held keys, content clicks, pointer scrolling or resizing.
No Enter/Space, activation of ANY button, dialog, reset, Nodes or connection.
Maximum THREE physical Tab presses, even if time remains after success.

## Time-bounded route

| Step | Latest prompt / acknowledgement (elapsed) | Gate |
| --- | --- | --- |
| Prepare | by60s | Login -> Enter demo -> disconnected/lock-off -> Account at default size |
| 0 | 60s /120s | Titlebar middle single click, phone 第0步完成 |
| 1 | 130s /190s | One Tab, phone 第1步完成; fresh visible Settings outline |
| 2 | 200s /260s | One Tab, phone 第2步完成; fresh visible MFA outline |
| 3, conditional | 220s /280s | ONLY if step2 already confirmed/observed by220s: one Tab, phone 第3步完成; inspect actual target |
| Close | immediately after final result, no later than300s | End input, normal X, retain observer to bounded result |

The overlapping deadlines are deliberate: if step2 completes late (after220s),
step3 is NOT issued, despite the step2 deadline being260s. Full60s acknowledgement
plus20s observation/close reserve is mandatory for step3. Four replies at their
maximum can exhaust available time; partial completion is expected and acceptable.
Never shorten a user's window, hurry them, batch keys or extend the process cap.
If preparation or any earlier gate fails, stop rather than search for focus.
Post-ack image may have one additional settled capture if ambiguous; never an
additional key. Missing/duplicate/late acknowledgements follow existing cancellation
rules and cannot silently advance or restart a step.

## Result boundary

PASS only for observed Settings -> MFA -> Manage devices sequence with distinct
matching acknowledgements and clear visible outlines. A different actual target
is recorded, not converted into the expected result. Unclear outline is UNVERIFIED,
not automatically product failure; unissued step is NOT_RUN; no acknowledgement
means input completion UNKNOWN. Do not infer automatic scrolling if none occurred;
an incidental scroll can be described but does not establish all lower controls.
No full forward/reverse traversal, dialog return or reset acceptance from this run.

Preserve prompt/ack times, actual labels and original exit/finalization records;
post-close CIM and six hashes supplement, never replace retained-process exit.
Commit sanitized report/status/recovery only, not ignored artifacts/raw receipts.
Next is review and separate live authorization; current allowance0.
