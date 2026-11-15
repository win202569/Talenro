# Ordinary Exit focus-only minimum

Date 2026-11-15. Documentation-only proposal, review and separate launch approval
required. No code/test/restore/build/launch. Current live allowance 0.

## Updated coverage

[Delete observation](../../roadmap/2026-11-15-windows-delete-focus-observation.md)
passed Settings -> MFA -> Devices -> Language -> Delete and keyboard reveal of
Delete, without activation. [Reverse observation](../../roadmap/2026-11-15-windows-reverse-focus-observation.md)
passed Devices -> MFA. These are bounded ordinary-candidate results only.

| Remaining item | Minimum evidence and exclusions |
| --- | --- |
| Delete -> Exit | Next minimum: one acknowledged Tab from visible Delete focus, clear Exit outline; no activation/logout/reset |
| Reverse button-chain edges | Exit -> Delete, Delete -> Language, Language -> Devices, MFA -> Settings remain unverified; Devices -> MFA already passed |
| Reverse scrolling | Separately observe keyboard-only reveal when needed; not inferred from forward results |
| Devices dialog focus return | Reach Devices, keyboard-open harmless placeholder, observe Close focus, keyboard-close, observe returned Devices outline; future separately budgeted approval |
| Reset and wider acceptance | Reset/re-entry, exact viewport/DPI/RTL, screen readers, security/Release/C4 remain outside this minimum |

This edge list is not exhaustive app focus coverage. Dialog source restoration
intent and diagnostic-build results do not prove ordinary runtime acceptance.
Do not append reverse/dialog actions to a successful Exit-focus run.

## Proposed next run

Inherit the [Delete focus plan](2026-11-15-windows-delete-focus-design.md) and
[phone protocol](2026-11-14-windows-ordinary-manual-keyboard-followup-design.md),
changing ONLY the future approved maximum to six individually acknowledged Tabs:
Settings -> MFA -> Devices -> Language -> Delete -> Exit. No seventh Tab,
Shift+Tab, Enter/Space, button activation, deletion, logout or dialog.
Normal titlebar X at the end is distinct from activating the Exit button.

One launch, zero retries, maximum360000ms, observer sleep cap2000ms. Normal close
starts promptly after last result and no later than300s. Keep exact ordinal
command-line/launcher binding, six hashes, unchanged source/clean branch,
ordinary token, verified host/CIM, zero-target/path/reparse checks, immutable
receipts and retained-original-process exit verification. No security changes,
fallback matching, force kill, dump, replacement observer or automatic retry.

| Stage | Latest prompt / acknowledgement from launch | Required fresh evidence |
| --- | --- | --- |
| Prepare | Complete by60s | Ordinary Login -> Enter demo -> disconnected/lock-off Connection -> default Account |
| 0 titlebar click | 60s /120s | Matching phone reply and reliable view |
| 1 Tab | 130s /190s | Settings outline |
| 2 Tab | 200s /260s | MFA outline |
| 3 Tab, conditional | 220s /280s | Devices outline |
| 4 Tab, conditional | 220s /280s | Language outline |
| 5 Tab, conditional | 220s /280s | Delete outline |
| 6 Tab, conditional | 220s /280s | Actual target; Exit expected, never assumed |
| Normal close | By300s | Titlebar X; original observer through bounded outcome |

Each conditional step requires the preceding matching acknowledgement AND clear
fresh expected outline. All share the220s issuance ceiling, not simultaneous
prompts. Step6 is omitted unless Delete has already been confirmed and captured
before220s. Reserve full60s response plus at least20s observation/close time.

The previous five-Tab run's Delete capture occurred after220s; that pace would
NOT permit a sixth prompt. Seven confirmations at their60s maxima also cannot
fit. This is a conditional maximum, not a completion promise: preserve partial
results and close if time is insufficient. Never hurry the user, shorten waits,
batch keys, skip confirmation, extend the cap or restart to finish.

Numbered asynchronous phone questions, one outstanding; record prompt/deadline,
match completion, then fresh read-only screenshot before next prompt. Desktop
stays on demo; no agent input during manual block, pointer scroll, resize, content
clicks, held/repeated keys or desktop chat switching. Late/duplicate replies cannot
revive cancelled steps. Unexpected target, lost focus or unreliable image stops
the sequence; at most one extra settled read, never an extra key or focus hunt.

## Verdict and handoff

Exit-focus PASS requires a clear Exit outline after confirmed sixth Tab from
clear Delete focus. Record actual target literally; pane-only metadata is not
focus proof. If initially clipped, judge viewport movement/full button and outline
reveal separately: PASS for this transition only, NOT_EXERCISED if not needed,
otherwise UNVERIFIED. No activation is necessary or authorized.

Missing acknowledgement: completion UNKNOWN; unissued step: NOT_RUN; unclear
outline: UNVERIFIED. Preserve earlier passes; inconclusive evidence is not itself
a reproduced bug. Stop after final result even with spare time. Record actual
timings, scope, persisted/returned finalization, exit code, post-close inventory
and six hashes; save sanitized report/status/recovery, not raw ignored receipts
or binaries. All remaining reverse/dialog/reset checks require separate scope.
