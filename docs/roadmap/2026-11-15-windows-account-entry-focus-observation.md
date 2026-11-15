# Ordinary Account-entry focus observation — 2026-11-15

User approved the Account-entry design and one launch/zero retries/360s maximum,
60s phone confirmation per step, no activation. Starting checkpoint
aec5cabe1b874a2fc4d55b132a97ef3bb12919b5. No source/test/build changes.
Existing ordinary entry retained all binding/preflight/safety checks.

Run813ce5d94c374895822a7bbb6abb2987, PID25612, retained session14841.
Agent observed Chinese Login, entered offline demo, observed disconnected and
lock-off Connection, then Account. Default1422x815 screenshot; no resize.
Automated input ended for the manual block. Numbered asynchronous user questions
had one outstanding action each,60s maximum, matching phone reply before capture.

| Step | Prompt clock UTC / first post-reply clock UTC | Evidence |
| --- | --- | --- |
| 0 | 07:56:26 /07:56:38 | User confirmed titlebar click; fresh Account screenshot, pointer in titlebar |
| 1 | 07:56:50 /07:57:02 | User confirmed one Tab; clear Settings outline |
| 2 | 07:57:12 /07:57:28 | User confirmed one more Tab; clear MFA outline in Account content |

Clock reads bracket processing, not exact server message-arrival times. All
confirmations were within their60s windows and original stage ceilings. No agent
activation/click/key during the block. No third Tab, Enter/Space, content click,
scroll, button activation, dialog, reset or connection. Accessibility focused
element still said pane2; visible outlines, not metadata alone, establish results.
Agent ended input and clicked normal titlebar X after the result.

## Result

PASS for the narrowly scoped Settings -> MFA Account-entry keyboard route.
Not full traversal: Devices/Language/Delete/Exit reachability, reverse navigation,
auto-scroll, placeholder keyboard close/focus return and reset remain unverified
or not run. No broader dimensions/DPI/RTL/screen-reader/security/Release/C4 claim.

Persisted sequence54:101566.1399ms, exited/agreement, ExitVerified=true,
Finalized=false, OverrunMs0. Returned:101579.1769ms, ExitVerified/Finalized=true,
OverrunMs0. Entry exit0. Post-close host CIM matching count0; six artifact hashes
match manifest. Exact-bound launch/exit PASS; no retry. Remaining allowance0.

Ignored raw receipts beneath ordinary candidate:
ordinary-account-observations/813ce5d94c374895822a7bbb6abb2987/.
Manifest SHA256 FFDA57F4009A6E66AED096AC0606460387C05414C45156C465A4F42F9C3B2F15.
Samples SHA256 FAE93A6443B3C3676243FAA8C0156558825F037637A6D87925EDF7A2F5491548.

Next: read-only plan remaining traversal/scroll/dialog coverage within the same
time and confirmation limits. Do not assume every remaining60s step fits one run,
automatically launch again, or change focus code based on previously inconclusive
observations. Clone restores sanitized records/source, not ignored raw artifacts.
