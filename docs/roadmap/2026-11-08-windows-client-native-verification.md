# Native verification — PARTIAL, not accepted

## Latest observed sessions, 2026-11-08

User explicitly approved one replacement first session after the interrupted
request below. Two observed sessions are now closed; no further launch allowance
remains. No confirmed application defect was observed, but the full checklist was
not completed. Do not infer acceptance from these selected successful paths.

- Replacement session: PID10208, start16:10:59.844+04:00; ordinary close requested
  at16:25:36+04:00, approximately14.6minutes (limit20). Startup observed within30s.
- Reset-only session: PID14520, start16:25:56.704+04:00; ordinary close requested
  at16:27:22+04:00, approximately1.43minutes (limit5). Login observed within30s.
- Exact executable: worktree-relative
  `clients/windows/Talenro.Demo.App/bin/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe`.
  SHA256 `026D8642F47107CE84465BCF0E3133EC02242AD25B59D41068A58BC11CA67C95`.
  Source baseline0d8a436f; prior test/build checkpointfe9c9f14, not rerun here.
- Both windows disappeared; ordinary-context process checks outside the restricted
  command context confirmed both exact PIDs absent. No kill was used.
- Restricted process queries could not see the second running process, and CIM /
  event-log reads were denied. Their empty results are NOT exit/crash evidence.
  A permitted read-only check identified PID14520 and its exact path/start time.
- Reset session's first input returned `coordinate input geometry is unavailable`.
  Rebinding the still-present returned window and refreshing its screenshot restored
  control; one input retry succeeded. No additional launch or app fault was inferred.
- Total accounting: one interrupted request + one explicitly approved replacement
  + one original reset-only request = three requests, two observed sessions.

### PASS — directly observed, narrowly scoped

1. Login: fixed read-only email/password mask, permanent offline notice, Chinese
   copy; invalid/submitting/rejected/unavailable/MFA states disable entry and show
   appropriate simulation explanations. Normal entry works. Reset-password opens
   an unimplemented explanation; Escape dismisses and focus visibly returns.
2. Connection: connect/cancel; success after cancellation ignored; explicit success
   enters simulated connected; disconnect waits for explicit completion then becomes
   disconnected. Ordinary failure shows retry/choose-location explanation. One
   labelled primary action and textual status; decorative power glyph not in the
   interactive accessibility tree (not a screen-reader audit).
3. Nodes: Auto initially selected; JP/DE groups; disabled fixture/group rows exposed
   disabled. Search for a nonmatching string shows no matching nodes; JP filter
   shows Japan only. Applying jp-1 while disconnected saves preference. Connected
   de-1 Apply opens a modal explaining simulated restart. Escape returns visible
   focus to Apply. Back then reopen retains jp-1 and connected DE lock. Cancel
   button also dismisses the modal; Confirm enters Connecting. No hidden dispatch
   count or atomicity claim is made from UI observations alone.
4. Country recovery: jp-1 + JP lock; explicit locked-country failure offers all
   three choices. One-time recovery then explicit success preserves saved JP lock.
   Turning lock off/on displays DE, confirming recovered-country relock behavior.
   Retry-in-country and stay-disconnected buttons were visible but not exercised.
5. Account: active example12.4/100GB and2/5 fictional activated devices, explicitly
   not concurrent. Unknown and stale both show unavailable usage/device counts,
   with refusal notices on Connection. MFA placeholder opens scope explanation;
   Enter dismisses. Admission by actually pressing Connect in these states was
   not exercised and is not marked passed.
6. Logout: returns normal Login and clears notice; reentry is disconnected with
   lock off, active account and Auto selection. Before close, jp-1 was applied and
   JP locked again. Reopen starts Login; entering then shows disconnected/lock off,
   and Nodes visibly selects Auto with empty search. Both ordinary closes verified.
7. Keyboard: selected Tab/Shift+Tab traversals, visible focus, Enter/Escape and
   automatic scrolling to lower controls worked. Screenshots of all four pages
   and dialogs were inspected in-session; no screenshot files were committed.

### UNVERIFIED — next bounded session must cover these gaps

- Registration and remaining sidebar/account placeholders; full keyboard order,
  modal default-Enter cancellation and screen-reader behavior.
- Nodes loading/unavailable/explicit-empty scenarios, actual disabled-row input,
  cross-country confirmation copy/lock change, independent Cancel state preservation.
- Unknown-country lock refusal; retry-in-country and stay-disconnected actions.
- Expired/exhausted/device-limit presentation and actual admission refusal in all
  nonactive states; logout resetting deliberately changed list/login selectors.
- Exact current desktop scaling/client dimensions were not measured. Bounded
  screenshots were1422x815, not proof of target logical viewport coverage.
  Lower controls required scrolling; keyboard auto-scroll worked. Pointer wheel
  attempts did not visibly move content, so pointer scroll remains unverified.
- Target1024x768/1440x1024 layouts, design comparison, contrast measurement,
  enlarged text, RTL/long content, English fallback, full accessibility and
  runtime/network/storage/security behavior remain unverified.

First-session work prioritized core flows and reset; it ended before the maximum
time, not because all checks passed or the time budget was exhausted. No source,
tests, build, restore, installation or system/display/security changes were made.
Further native launches require a new finite allowance; no automatic retry.
Task11/12 and44/34/6/217 NOT_ACCEPTED unchanged; no main merge.

## Historical interrupted request (superseded status)

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
