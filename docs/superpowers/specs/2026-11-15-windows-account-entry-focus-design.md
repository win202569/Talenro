# Minimal Account-content focus entry observation

Date2026-11-15. Proposed next observation, not launch authorization. Documentation
only; no code/test/build/restore/launch. Existing one-run allowances exhausted.

## Evidence and alternatives

Ordinary run b3aac484bee7448d8db59c6bb3b29de8 demonstrated a user-confirmed single
Tab to visible Settings outline, despite pane2 accessibility metadata. It did not
test Account-content traversal. Source declares MFA then Devices/Language/Delete/
Exit buttons; declaration order is not runtime proof of the next focus target.

Recommended minimum: reproduce the same starting route, then add ONE Tab to
inspect entry into Account content. Full forward/reverse traversal plus dialog
would require many60s acknowledgement windows and is not this proposal. Clicking
MFA to establish focus would change the keyboard route and could activate a
placeholder; that is not selected. No code/instrumentation change is needed.

## Inherited safety and confirmation rules

Inherit unchanged exact command-line/launcher/source/six-artifact binding, ordinary
token, authorized host/CIM, clean branch, zero-target, path/reparse checks and
original retained-process exit validation from the
[phone protocol](2026-11-14-windows-ordinary-manual-keyboard-followup-design.md).
This proposed route replaces only that document's single-Tab scope/stage table
for a separately approved future run; it does not authorize extra keys in old runs.

One launch, zero retries,360000ms total cap, normal close no later than300s,
observer max sleep2000ms. Same-conversation phone readiness required before launch.
Use asynchronous numbered user questions, as in the successful run, with one
outstanding action at a time. Each waits at most60s; matching phone completion
must precede fresh screenshot judgment and next prompt. No agent input during
manual block, no desktop chat replies, no batched/held keys or pointer scrolling.
No confirmation is not proof of failed input. Duplicate/late messages do not
advance or cause repeated keys; cancellation is final. No pause extends the cap.

No Enter/Space, MFA or Delete activation, dialogs, connections, Nodes, credentials,
settings changes, resize, reset/re-entry or additional keys. Default geometry only.

## Reduced route and time budget

Elapsed time is measured from the original launch. Each prompt needs its FULL60s
reservation within its stage deadline; if insufficient time remains, omit the step
and close normally. User completion can be earlier; no need to wait the full60s.

| Step | Latest issuance / completion | Action and gate |
| --- | --- | --- |
| Preparation | by60s | Agent Enter demo, verify disconnected/lock-off Connection, navigate Account. End automated input. |
| 0 | by60s /120s | One titlebar-middle click, phone 第0步完成. Fresh read-only target view. |
| 1 | by130s /190s | ONE Tab, phone 第1步完成. Fresh view must clearly show Settings outline to continue. |
| 2 | by200s /260s | ONE more Tab, phone 第2步完成. Fresh view records actual focus target, with at most one extra settled capture if ambiguous. |
| Close | promptly after result; aim by280s, hard latest300s | End user input and normal titlebar X; retain original observer to bounded result. |

Maximum TWO physical Tab presses total. If step1 differs from Settings, do not
hunt for it. If step2 focuses a pane, another control, or has no clear outline,
record that result and close; no third Tab or corrective click. If tool latency
consumes the10s inter-stage margin, stop with partial coverage rather than shorten
the user's60s allowance. No guarantee all stages will fit.

## Acceptance and remaining scope

- PASS for the narrow Account-entry goal only if a fresh step1 image shows
  Settings and the post-ack step2 image clearly shows MFA button focus without
  intervening agent/user content input. Never activate MFA.
- If another target is visibly focused, record its label as observed; the desired
  Account-entry goal remains UNVERIFIED, not automatically a product failure.
- Confirmed input but unclear outline: UNVERIFIED. Missing acknowledgement: input
  completion UNKNOWN. Explicit user confirmation of no input: NOT_RUN.
- Even success does not prove Devices/Language/Delete/Exit reachability, reverse
  navigation, automatic scrolling, placeholder close/focus return or reset.
  No broader dimensions/DPI/RTL/screen-reader/security/Release/C4 acceptance.
- Preserve prompt/ack timing, actual labels, screenshots' observation order,
  original exit result, post-close inventory and hashes. Commit sanitized docs
  only; retain ignored raw receipts unchanged. New computer restores source/docs,
  not local artifacts or raw session evidence.

Next action is user review and separate bounded launch approval, not execution.
