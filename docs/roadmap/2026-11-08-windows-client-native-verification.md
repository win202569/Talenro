# Native verification — interrupted before startup observation

User approved the native acceptance plan after735383cb. Preconditions: ordinary
command context AdministratorToken=False; no existing demo process/window found;
App/Core source unchanged from0d8a436f; executable hash matched the plan record.

One launch request was issued through the native control tool. The tool reported
that the user stopped Computer Use with physical Escape. No window snapshot or
successful startup was observed; no further UI calls were made that turn.

User subsequently requested continuation. Read-only window/process checks found
no Talenro window or process at2026-11-08T16:09:18+04:00. A bounded Application
event-log query for recent30minutes,IDs1000/1026,matching Talenro.Demo.App returned
no matching output (exit1). This does not establish successful launch, clean exit
or absence of a crash. No unrelated log contents were retained.

Status: startup UNVERIFIED; all page/visual/keyboard/accessibility/reset checks
UNVERIFIED. No screenshot evidence. No app process remains visible in the checks;
no close/kill was attempted. No source changes/build/test/restore/settings changes.

Accounting:1 launch request issued with unknown outcome;0 observed sessions.
Second planned launch is reserved for reset after a successful first session,
not a retry. Need explicit approval for one replacement first-session launch
before continuing; do not reinterpret general continuation as extra retry scope.
Proposed replacement retains20-minute first-session/30-second startup bounds,
then at most1 five-minute reset session only if the first succeeds and exits.
Any fault stops work; no automatic retry. Broader NOT_ACCEPTED unchanged.
