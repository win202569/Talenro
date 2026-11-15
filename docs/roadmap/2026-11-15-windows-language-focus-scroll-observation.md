# Ordinary Language focus and automatic scrolling — 2026-11-15

User approved one launch/zero retries/360s maximum, four Tabs maximum, no
activation,60s phone acknowledgements. Starting checkpoint
8f17df407112570b57f3e9d9ac8cadf3372f2638. No code/test/build changes.
Exact binding and all ordinary entry/preflight/safety checks retained.

Run16c1222c16db44bbb50e0f8c0e7d567e, PID18408, retained session37687.
Agent observed Chinese Login, entered offline demo, observed disconnected/lock-off
Connection, then Account at default1422x815 screenshot bounds. No resizing.
Agent input stopped for manual block. All five numbered phone acknowledgements
arrived within60s and were followed by fresh read-only screenshots.

| Step | Clock reads UTC | Result |
| --- | --- | --- |
| 0 titlebar | Prompt15:31:05, after reply15:31:20 | Account view, titlebar pointer |
| 1 Tab | After reply15:31:47 | Clear Settings outline |
| 2 Tab | After reply15:32:11 | Clear MFA outline |
| 3 Tab | Prompt15:32:30, after reply15:32:49 | Clear Manage devices outline; Language bottom clipped |
| 4 Tab | After reply15:33:17 | Clear full Language outline; Account content shifted upward |

Some replies arrived before a separate prompt-clock call; exact issuance timestamps
for those prompts are not invented. Reads bracket processing, not server delivery.
All stage ceilings met; fourth prompt issued before220s. No fifth Tab, Enter/Space,
button activation, pointer scrolling, content click or interleaved agent activation.

Before step4, Language's lower boundary was clipped at viewport bottom. After
step4, heading/email/usage and buttons shifted upward together while sidebar and
window geometry remained stable; Language's full button and focus outline became
visible. Final accessibility scrollbar value31 supports but does not replace this
visual comparison. Focus metadata still pane2, despite clear Language outline.

PASS for Settings -> MFA -> Manage devices -> Language visible keyboard route.
PASS for automatic keyboard scrolling revealing Language on this transition only.
Delete/Exit, reverse traversal, dialog keyboard closure/focus return and reset
remain outside this observation; no all-controls/all-viewports/UI/C4/security claim.

Agent announced end then clicked normal titlebar X. Persisted sequence102:
198145.2887ms, exited/agreement, ExitVerified=true, Finalized=false, OverrunMs0.
Returned198158.5821ms, ExitVerified/Finalized=true, OverrunMs0. Entry0. Post-close
matching CIM count0; six artifact hashes match manifest. Launch/exit PASS, no retry,
allowance0. No source/test/build changes. Next only separately scoped planning.

Ignored receipts beneath ordinary candidate:
ordinary-account-observations/16c1222c16db44bbb50e0f8c0e7d567e/.
Manifest SHA256 687078B8661DF89A257148526915429E4373512467A5A1359C2DF19BDF0096E9.
Samples SHA256 F7BC9714CAE2B5E4A9C4C0AC07092B61C4BE635AEAC3A5C79E5FAF9C099F091B.
