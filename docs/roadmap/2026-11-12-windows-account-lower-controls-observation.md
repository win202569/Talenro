# Account lower-controls observation — scoped partial evidence

Checklist confirmed, then user explicitly approved one launch/zero retries/up to
six minutes. Baseline cef182f3; fixed account-center entry unchanged.
Run413d4b9827c2434a9e8ede6879f9b059, PID9232 (now exited). No build/test/source
change, system-setting change, real connection or forced termination.

## Conditions and deviations

User confirmed seeing the foreground window. User resized on Login rather than
Connection; agent then entered Connection and confirmed stable root629x620,
page409x534 DIP twice before first Account visit, rendering/text1. This differs
from the planned resize page but preserves its critical pre-Account capture gate.
Login samples remain original1136x620, not evidence for later resized Login.
One activation was blocked by detected user input; state was refreshed before
continuing. Occluding-chat screenshot was excluded; target activation restored
app-only observation. No relaunch or input based on stale coordinates.

## Observed results

| Check | Result | Evidence / limitation |
| --- | --- | --- |
| Pointer scroll and five lower labels | Scoped PASS | Account content moved; MFA, Manage devices, Language, Delete account, Exit demo/reset all fully visible at lower position; final scroll value192 |
| Forward keyboard path | Scoped PASS | Visible outline Settings→MFA→Manage devices→Language→Delete account→Exit demo/reset, one Tab per edge |
| Reverse path | Scoped PASS | Shift+Tab Exit→Delete→Language→Manage devices, visible outline at each step |
| Keyboard-caused automatic scroll | UNVERIFIED separately | Pointer-scroll animation and early Tab observations overlapped; do not attribute movement independently to keyboard. Tested focus targets were visible |
| Manage devices explanation | Scoped PASS | Enter from visibly focused button opened Offline demo / This feature is outside this demo; labelled Close readable and used once |
| Immediate visible return focus | UNVERIFIED | No outline on Manage devices after pointer Close; accessibility focus metadata remained pane2 throughout and was not reliable evidence. This is not established as a product defect |
| Next keyboard move after close | Scoped PASS | One Tab after dismissal visibly outlined Language, consistent with usable continuation from Manage devices; not proof of immediate outline |
| Logout visible recovery | Scoped PASS | Exit demo/reset returned Login, sidebar hidden, English fictional fields and RTL heading retained; previous placeholder notice cleared |
| Re-entry visible recovery | Scoped PASS | Enter demo once returned Connection disconnected/Country lock Off, English/right-side navigation retained, blank notice |

Delete account was never activated: only Tab/Shift+Tab traversed it. No other
account action was triggered. Re-entry stayed in the same process; no exhaustive
reset, all-dialog, screen-reader, ordinary-build or full C4 acceptance follows.

## Closure and integrity

Normal titlebar X before300s. Console outcome sequence118, elapsed229925.1713ms,
exited/agreement, ExitVerified=true, Finalized=true, OverrunMs0; command exit0.
Persisted outcome elapsed229904.5173ms, Finalized=false (pre-cleanup) retained.
UI target list empty, successful CIM target/prototype count0 after exit.

Manifest artifact hashes match actual fixed EXE/DLL. Process118 rows have matching
run ID and contiguous sequence. Geometry wrapper run IDs match: header, Login1/2,
Account3/4, health Samples4/Failed=false/ErrorCodes empty/Provisional=true. Account
samples root629x620/viewport409/scroll0 precede scrolling. They do NOT measure
later focus, dialog, scroll192, logout or re-entry states. Raw status not promoted.

Inventory below is relative to candidate output
`clients/windows/Talenro.Demo.App/bin/email-geometry-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/`;
RUN means413d4b9827c2434a9e8ede6879f9b059.

| Path | Bytes | SHA256 |
| --- | ---: | --- |
| email-geometry-observations/RUN/manifest.json | 730 | CBF57E7E58BE5865301EF29B7BE49C397662071A40AEDFB8A388D6C53617A8E2 |
| email-geometry-observations/RUN/samples.jsonl | 222465 | A4AD178CD5299D775B34A1AAE2759F8FC1905457AA1E134C7B27ABFDDACDCAE9 |
| email-geometry-records/RUN/geometry.jsonl | 36085 | 48E05678CE823556F540B75421392BA2A6D2151B3F1D9A9EED7167E9C77C9D77 |
| email-geometry-records/RUN.claim | 0 | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

## Next boundary

Live0. Do not repeat the whole route automatically. Before any further launch,
read-only review whether pointer-dismissal versus keyboard focus-visual behavior
explains the missing outline, and define an isolated automatic-scroll observation
only if needed. R1 is partially covered, visible logout part of R2 has current
evidence; remaining gap map still applies. C4/overall NOT_ACCEPTED. Raw receipts
stay ignored/local; Git restores this sanitized summary, not binaries or raw data.
