# Language focus and conditional automatic-scroll minimum

Date2026-11-15. Documentation-only proposal, awaiting review and separate launch
approval. No code/test/build/restore/launch; previous live allowance exhausted.

## Updated remaining evidence

Ordinary run e244b029f0ba42d48d85aba17446388e verified visible Settings -> MFA ->
Manage devices outlines after three acknowledged Tabs. Its layout did not scroll
and scrollbar was0. Its screenshot showed Language partly clipped at the bottom.
Neither that snapshot nor source order proves the next key's runtime behavior.

| Item | Status / evidence needed |
| --- | --- |
| Settings/MFA/Manage devices forward route | PASS for recorded ordinary run; rebuild focus in every new process |
| Language reachability | Pending; next minimum is one additional Tab with visible Language outline |
| Automatic scrolling | Pending; compare pre-key clipping and post-key viewport movement/reveal without mouse scrolling |
| Delete/Exit reachability | Pending, later focus-only steps; no activation implied |
| Reverse traversal | Pending; individual Shift+Tab transitions, not inferred from forward reachability |
| Manage devices placeholder focus return | NOT_RUN; separate keyboard-open, visible Close focus, keyboard-close and returned Devices outline evidence |
| Reset, exact geometry/DPI/RTL, screen readers, broader security/Release/C4 | Outside this minimum, unchanged |

## Smallest proposed next route

Recommend one new transition beyond the proven route: fourth Tab from Manage
devices to Language. Observe automatic scrolling only if this viewport actually
requires it. Repeating all lower controls and opening a dialog in one run would
add too many independent60s confirmations. Pointer scrolling/click-to-focus would
not establish keyboard auto-scroll and is excluded. Do not change viewport or
instrument the binary merely to force the expected observation.

One launch, zero retries, maximum360000ms; normal close begins by300s, observer
sleep cap2000ms. Unchanged exact command-line and launcher binding, six artifact
hashes, source checks, ordinary token, authorized host/CIM, clean branch,
zero-target/path/reparse checks, immutable receipts and retained-process exit.
Inherited confirmation/cancellation rules from
[phone protocol](2026-11-14-windows-ordinary-manual-keyboard-followup-design.md).
Only the future approved route/max-key count below supersedes the earlier route;
this document authorizes no launch or additional key in any previous run.

Same-conversation phone readiness, asynchronous numbered questions, one outstanding
action, up to60s per step. Record prompt/deadline before waiting; match numbered
completion then fresh read-only screenshot before next prompt. Keep desktop on
demo. No agent input during manual block; no desktop-chat reply, repeated/held or
batched keys, content clicks, resize, wheel or scrollbar use. Late/duplicate replies
cannot revive cancelled steps. No Enter/Space, button activation, dialogs, reverse
keys, reset, Nodes, connection, credentials or security-setting changes.

## Schedule and explicit partial-completion rule

| Stage | Latest prompt / response (elapsed from launch) | Required evidence |
| --- | --- | --- |
| Prepare | by60s | Ordinary Login -> Enter demo -> disconnected/lock-off -> Account, default size |
| 0 titlebar click | 60s /120s | Matching phone confirmation and reliable Account view |
| 1 first Tab | 130s /190s | Clear Settings outline |
| 2 second Tab | 200s /260s | Clear MFA outline |
| 3 third Tab, conditional | 220s /280s | Clear Manage devices outline |
| 4 fourth Tab, conditional | 220s /280s | Record actual target and viewport change; Language expected, never assumed |
| Close | promptly after last result; no later than300s | End input, normal titlebar X; retain original observer through bounded result |

Steps3 and4 deliberately share the220s latest-issuance ceiling. Step4 is issued
ONLY if step3 was already acknowledged, captured and judged before220s. Each new
prompt gets the full60s reservation plus20s observation/close reserve; never send
two steps together. If any preceding response/capture leaves insufficient time,
omit the next step, record partial coverage and close. At most FOUR Tabs total.
Five confirmations at their60s maxima cannot fit with preparation and safe close;
this plan is opportunistic within the cap, not a promise all steps will complete.
Do not hurry the user, shorten waits, extend the cap or restart to finish it.

Unexpected target, uncertain outline, loss of desktop focus or unreliable target
image ends the sequence; do not hunt for focus. One extra settled read allowed
for ambiguous post-ack image, no extra key. End immediately after step4 result
even if spare time remains. No fifth Tab or on-the-fly dialog test.

## Independent result criteria

- Language-focus PASS only for a clear Language outline after confirmed step4
  following observed Settings/MFA/Devices. A different visible target is recorded
  literally; desired Language coverage stays unverified, not automatic defect.
- Auto-scroll PASS for this transition only if a pre-step4 image shows Language
  clipped and the post-step4 image shows viewport movement revealing its full
  button/focus boundary, with no pointer scrolling/resizing/intervening action.
  Accessibility scrollbar data is supplemental, never sufficient by itself.
- If Language was already fully visible or no scrolling was needed, mark scroll
  NOT_EXERCISED; do not force it. If movement/reveal cannot be established, mark
  UNVERIFIED. Language focus can pass while scroll remains unverified.
- This does not cover all lower controls, reverse navigation or dialog return.
  Missing acknowledgement means completion UNKNOWN; unissued step NOT_RUN;
  confirmed input/unclear outline UNVERIFIED. Preserve earlier accepted results.

Future dialog proposal must separately budget reaching Devices, keyboard activation,
visible Close focus, keyboard closure and visible returned Devices outline; source
Programmatic focus intent is not proof. It is not authorized by this plan.
Preserve actual timing/labels/pre-post evidence, original finalized versus persisted
exit, post-close inventory and six hashes. Commit sanitized report/status/recovery
only, not raw ignored receipts/binaries. Current live allowance0.
