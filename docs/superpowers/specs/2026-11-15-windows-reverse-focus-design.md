# Ordinary reverse-focus minimum and remaining coverage

Date 2026-11-15. Documentation-only proposal; review and a separate single-launch
approval required. No implementation, tests, restore, build or application launch.
Previous allowance consumed; current allowance 0.

## Evidence and remaining scope

[Recorded ordinary observation](../../roadmap/2026-11-15-windows-language-focus-scroll-observation.md)
verified Settings -> MFA -> Manage devices -> Language and automatic scrolling
revealing clipped Language in that default viewport. This is not complete keyboard
acceptance; diagnostic-binary results cannot substitute for ordinary evidence.

| Remaining item | Minimum evidence still needed | Boundary |
| --- | --- | --- |
| Delete button | Acknowledged forward key from visibly focused Language, visible Delete outline, reveal if clipped | Focus only; never activate deletion |
| Exit button | Acknowledged forward key from visibly focused Delete, visible Exit outline, reveal if clipped | Focus only; no logout/reset/re-entry |
| Reverse traversal | Individually acknowledged Shift+Tab transitions with fresh before/after outlines | Next minimum covers Devices -> MFA only, not all reverse edges or reverse scrolling |
| Dialog focus return | Reach Devices by keyboard; separately acknowledge activation, observe dialog/Close focus, keyboard close, returned Devices outline | Separate future approval; no assumption from source Programmatic focus intent |
| Broader coverage | Other viewport/DPI/RTL, screen reader, reset, security/Release/C4 | Unchanged, not accepted by these checks |

Delete/Exit need a future bounded plan; do not append extra keys to this run or
batch them to fit. Dialog work needs its own time budget for reaching Devices,
opening, identifying Close, closing and observing return. If that cannot fit,
reduce the declared evidence scope rather than waive acknowledgements or timing.
No deletion, logout, connection or credential operation is proposed here.

## Next minimum: three Tabs, then one Shift+Tab

Re-establish Settings -> MFA -> Manage devices in a fresh ordinary process, then
observe exactly one reverse transition back to MFA. Maximum four key actions:
three individual Tabs followed by one Shift+Tab chord (press once, release both).
No Enter/Space, button activation, fourth forward Tab, extra reverse key or dialog.

Keep the [60-second phone protocol](2026-11-14-windows-ordinary-manual-keyboard-followup-design.md)
and all existing preflight/exit gates: exact ordinal command-line/launcher binding,
six artifact hashes, unchanged source, clean authorized branch, ordinary token,
verified host/CIM access, zero matching targets, path/reparse checks, immutable
receipts and original retained-process exit verification. No security changes,
force termination, dumps, replacement observer or fallback identity matching.
Use the existing observe-ordinary-account-center.ps1 entry only after new approval.

One launch, zero retries, total maximum 360000ms; normal close starts no later
than elapsed 300s. Observer sleep cap 2000ms. Preparation is ordinary Login ->
Enter demo -> disconnected/lock-off Connection -> Account at default size.

| Stage | Latest prompt / acknowledgement, elapsed from launch | Required observation |
| --- | --- | --- |
| Preparation | Complete by 60s | Reliable Account view |
| 0 titlebar single click | 60s / 120s | Matching phone confirmation |
| 1 one Tab | 130s / 190s | Settings outline |
| 2 one Tab | 200s / 260s | MFA outline |
| 3 one Tab, conditional | 220s / 280s | Manage devices outline |
| 4 one Shift+Tab, conditional | 220s / 280s | Actual returned focus; MFA expected, not assumed |
| Normal close | Promptly after final result, by 300s | Original observer retained through bounded outcome |

Steps 3 and 4 share the latest issuance ceiling, not a simultaneous prompt.
Step 4 requires step 3 confirmation AND fresh clear Devices evidence before 220s.
Every issued step reserves a full 60s response window plus at least 20s for
observation/closing before 300s. Five confirmations at their maxima cannot fit;
partial completion is expected if replies are slow. Never shorten waits, extend
the cap, batch actions or relaunch to complete coverage.

Use numbered asynchronous phone questions, one outstanding at a time; record
prompt/deadline, match the numbered reply, then take a fresh read-only screenshot
before the next prompt. Desktop stays on the demo; replies are phone-only.
No agent input during the manual block, content clicks, resize, wheel/scrollbar
use, held/repeated keys or desktop chat switching. Late/duplicate replies do not
revive cancelled steps. An unexpected target, uncertain outline or unreliable
target capture stops the sequence; allow at most one extra settled read, no key
retry or focus hunting. Normal titlebar X is reserved for ending the run.

## Verdict and preservation

Reverse-pair PASS requires confirmed step 4 after visible Devices focus and a
fresh clear MFA outline, with no intervening action. Pane/window focus metadata
alone is insufficient. Record a different actual target literally; desired
coverage stays unverified, without automatically diagnosing a product defect.
No acknowledgement means completion UNKNOWN; unissued step NOT_RUN; unclear
post-key focus UNVERIFIED. Preserve previous accepted forward/scroll results.
Stop after this result even if time remains. Delete/Exit, other reverse edges,
reverse auto-scroll and dialog return remain pending.

Preserve actual timings, evidence descriptions, original persisted versus returned
finalization, entry exit code, post-close inventory and six hashes. Save sanitized
report/status/recovery only; no ignored raw receipts, binaries or caches in Git.
