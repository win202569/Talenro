# Ordinary Delete focus-only minimum

Date 2026-11-15. Proposal only, awaiting review and separate launch approval.
This update changes documentation only: no code, tests, restore, build or launch.
Live allowance 0. Existing entry and all safety gates remain unchanged.

## Current evidence and remaining coverage

[Forward observation](../../roadmap/2026-11-15-windows-language-focus-scroll-observation.md)
passed Settings -> MFA -> Devices -> Language, including automatic reveal of
Language in that viewport. [Reverse observation](../../roadmap/2026-11-15-windows-reverse-focus-observation.md)
passed Devices -> MFA only. Neither accepts complete traversal or dialog return.

| Item | Remaining evidence / boundary |
| --- | --- |
| Language -> Delete | Next minimum: one confirmed Tab from clear Language focus to clear Delete focus; never activate |
| Delete -> Exit | Separate future focus-only check; no logout or reset |
| Reverse paths | Exit -> Delete, Delete -> Language, Language -> Devices, MFA -> Settings remain unverified; Devices -> MFA already passed |
| Reverse automatic scrolling | Observe movement/reveal only when a future approved reverse transition requires it; not inferred from forward scrolling |
| Dialog return | Separately budget reaching Devices, keyboard opening, visible Close focus, keyboard closing and visible returned Devices outline |
| Reset and broader acceptance | Exact geometry/DPI/RTL, screen readers, security/Release/C4 remain outside this minimum |

Reverse-path list describes this observed button chain, not every possible focus
edge in the app. Source focus-restoration intent is not dialog evidence. Do not
transfer diagnostic-build acceptance or combine remaining items opportunistically.

## Smallest next route

Re-establish Settings -> MFA -> Devices -> Language with four individually
acknowledged Tabs, then at most ONE more Tab to inspect Delete focus. Maximum
five Tabs, no Shift+Tab, Enter, Space, activation, sixth Tab or dialog. Delete
means focus only: no deletion confirmation or destructive operation is authorized.
Do not use a pointer click/scroll shortcut to reach the lower button.

Keep [phone protocol](2026-11-14-windows-ordinary-manual-keyboard-followup-design.md):
one numbered asynchronous question outstanding, full 60s phone acknowledgement
window per issued step, fresh screenshot after matching reply before next prompt.
Desktop remains on demo; no agent input during manual block, held/repeated/batched
keys, content clicks, resize, pointer scroll or desktop chat switch. Late or
duplicate replies cannot revive cancelled steps.

One launch, zero retries, cap 360000ms, observer sleep cap 2000ms. Normal close
starts promptly after final result and no later than elapsed 300s. Retain exact
ordinal command-line/launcher binding, six artifact hashes, source/clean-branch,
ordinary token, verified host/CIM, zero-target, path/reparse checks, immutable
receipts and original retained-process exit verification. No fallback matching,
security changes, force kill, dumps or replacement observer.

| Stage | Latest prompt / acknowledgement, elapsed from launch | Fresh evidence |
| --- | --- | --- |
| Prepare | Complete by 60s | Ordinary Login -> Enter demo -> disconnected/lock-off Connection -> Account, default size |
| 0 titlebar single click | 60s / 120s | Matching phone reply, reliable Account image |
| 1 one Tab | 130s / 190s | Settings outline |
| 2 one Tab | 200s / 260s | MFA outline |
| 3 one Tab, conditional | 220s / 280s | Devices outline |
| 4 one Tab, conditional | 220s / 280s | Language outline |
| 5 one Tab, conditional | 220s / 280s | Actual target; Delete expected, never assumed |
| Close | By 300s | Normal titlebar X, original observer through bounded outcome |

Steps 3–5 share a latest-issuance ceiling, not simultaneous prompts. Each requires
the preceding confirmation and fresh clear expected focus. Step 5 is sent only
if Language is already confirmed/observed before 220s, reserving a full 60s plus
at least 20s observation/close time. Six confirmations at their maxima cannot fit
the cap: completion is conditional, not guaranteed. If insufficient time remains,
record partial results and close; never hurry the user, shorten waits, extend the
cap, batch keys or relaunch. A slow earlier step may leave later steps NOT_RUN.

## Decision and reporting

Delete-focus PASS requires clear Delete outline after acknowledged step 5 from
fresh clear Language focus. Record actual target literally if different; stop,
do not hunt for focus or automatically infer a product defect. Unreliable image,
loss of focus or uncertain outline ends sequence; at most one extra settled read,
no retry key. Pane-only accessibility metadata cannot establish visible focus.

If Delete was clipped, compare images for keyboard-only viewport movement and
full button/focus-boundary reveal. Judge this separately: scroll PASS for that
transition only, NOT_EXERCISED if no scroll needed, otherwise UNVERIFIED. Missing
acknowledgement means completion UNKNOWN; unissued step NOT_RUN; unclear focus
UNVERIFIED. Stop after step 5 even if time remains. Preserve earlier passes.

Retain actual timing, scope and visible evidence; distinguish persisted/returned
finalization, entry exit, post-close inventory and six hashes. Commit sanitized
observation/status/recovery only, not ignored receipts or binaries. Exit focus,
remaining reverse edges, dialog return and reset require separate proposals and
approvals. No new launch authorized by this document.
