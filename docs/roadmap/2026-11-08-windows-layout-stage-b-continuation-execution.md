# Stage B continuation execution

Plan: docs/superpowers/plans/2026-11-08-windows-layout-stage-b-continuation.md.
User approved6 sequential launches after98434867. Source1ac1c19e; starting
documentation HEAD98434867225b06ddde1b8c00c249bc59273c11bc.

Task1 complete: clean expected development branch, AdministratorToken=False,
DemoProcessCount=0, no demo window; supported UI inventory succeeded.
Diagnostic directory: clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64.
Fresh Talenro.Demo.App.dll SHA256:7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB.
Fresh Talenro.Demo.App.exe SHA256:4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95.
Both match approved plan. Budget initial0/6 consumed; no builds/restores/tests.

Preflight interface: Task1 identities bind Task2 sessions; Task2 observations feed
Task3 portable results. No source interfaces change.
Ruling: retain PowerShell and portable ledger instead of Bash helpers; execute the
approved verification-only scope without new TDD/reviewer cycle. Cost: no renewed
automated-suite or independent implementation-review assurance.

## Outcome — stopped at C3 shutdown, C4 started prematurely

Consumed4/6 launches. C5/C6 NOT_RUN;2 unused slots expire at handoff. No builds,
restores, tests, code changes, system changes or force-kill. All timestamps below
are2026-11-08 +04:00 (Asia/Tbilisi). Each launch passed ordinary-token guard.

| Row | Arguments | PID / start | Result / exit |
| --- | --- | --- | --- |
| C1 | --layout-test --layout-language=en-US --layout-profile=baseline | 18836 /21:35:44.3893265 | PASS scoped route; normal close exit0 |
| C2 | --layout-test --layout-language=zh-Hans --layout-profile=long-status | 4512 /21:37:44.7204541 | PASS conditional sample and Cancel; normal close exit0 |
| C3 | --layout-test --layout-language=zh-Hans --layout-profile=long-country | 12928 /21:38:35.0789325 | UI checks below passed; FAIL normal shutdown: window closed, process remains; exit UNVERIFIED |
| C4 | --layout-test --layout-language=en-US --layout-profile=rtl-mixed | 13676 /21:40:44.0816637 | Started before C3 exit was confirmed, protocol violation; no profile acceptance checks; normal close exit0 |
| C5 | --layout-test --layout-language=zh-Hans --layout-profile=unknown-usage | not started | NOT_RUN |
| C6 | none | not started | NOT_RUN |

C1 observed English Login, Connection, Nodes, Account and permanent offline notice.
Explicit simulated connection then success, de-1 selection/Apply opened 'Confirm
demo node change', 'Switch node? The simulated connection will restart.', Confirm
and Cancel. Cancel then Back retained connected status; reopening Nodes selected
original Auto rather than de-1. Account Tab visibly outlined Settings (not clicked).
Exit demo/reset returned English Login; re-entry showed disconnected/off in text
and ordinary connect action. No broader keyboard/accessibility claim.

C2 started inactive at Login/disconnected Connection. Explicit Connect displayed
the exact committed long-status fixture, visibly wrapping in3 complete lines with
active title and Cancel action; CountryLock disabled as expected. Cancel restored
Chinese disconnected status, ordinary connection action and inactive title.

C3 showed exact long-country fixture on JP group and jp-1 row, stable ' · jp-1'
suffix, no clipping, other rows unchanged. Search 'jp-1' retained the group and
matching row; selected/applied while disconnected. Explicit connection/success,
cleared search, selected de-1/Apply: ordinary Chinese confirmation, not synthetic
country text. Cancel/Back retained connected state; reopening Nodes retained jp-1.
Only jp-1 search was sampled; JP country-code search was not separately executed.

Shutdown failure: Alt+F4 removed C3's window but its launcher did not return exit.
The executor incorrectly continued to C4 after a5s poll returned a live launcher
session without exit0. This violated the approved one-process-at-a-time gate and
must not be counted as successful serial verification. Subsequent checks stopped
all dependent work; C4 closed normally with exit0. No C4 fixture/page checks were
performed; a title identifying rtl-mixed alone is not verification.

After more than30s, C3 launcher still had no exit receipt. A broader read-only
process check with AdministratorToken=False found PID12928, matching exact path
and21:38:35 start, MainWindowHandle=0. Restricted-context Get-Process returned no
processes, so that negative result is not evidence of absence. UI inventory found
no demo windows. PID12928 remains unresolved; do not relaunch or assume cleanup.
Process wait session8604 is bounded to630s and reports timeout without killing;
its eventual status is not an exit0 claim. No forced termination authorized.

Observation limitations: the initial C1 screenshot showed the occluding app while
the text tree identified the demo; explicit target activation corrected capture
before input. Several immediate accessibility snapshots lagged screenshots;
refreshed before relying on changed text. No unrelated app was controlled.
Titles reported root1136x620 and page1136x534(Login)/916x534(authenticated), render1
and text1. These are self-reported only, not independent dimensions or scale tests.

Next: separately scope read-only diagnosis of PID12928 shutdown/lifecycle and
the launcher exit wait before any replacement native sessions. Preserve original
records. Do not patch code, force-kill, reuse unused slots or mark Stage B accepted.
Earlier switch-language closure stands only in its prior scope; Task11/12 and
44/34/6/217 remain NOT_ACCEPTED. C4–C6 and exact-size/accessibility gaps stay open.
