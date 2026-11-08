# Native verification — PARTIAL, not accepted

## Latest addendum — PageHost correction verified on three entry paths

After e19a7e38, one authorized session PID9008 ran about215seconds and closed
normally (exact PID absent and refreshed window list empty). Settings→Connect,
Settings→Search and Settings→MFA each take one Tab; Shift+Tab returns directly.
Representative Settings explanation Escape focus return and Account lower-control
auto-scroll passed. See2026-11-08-windows-client-focus-fix.md for commands, times,
artifact hash, scope and limitations. Stale focus metadata remains untrusted.
Five observed closed sessions cumulatively, plus one historical interrupted request.
This addendum supersedes the historical extra-stop blocker below only for those
tested paths, not full keyboard/accessibility acceptance. No launch allowance remains.

## Latest keyboard/reset session, 2026-11-08 — PARTIAL

User confirmed the one-launch/20-minute keyboard/reset plan afterc4f64190.
Preconditions: AdministratorToken=False; no demo process/window; clean worktree;
App/Core unchanged against0d8a436f; executable path/hash matched the record below.
PID10372 started2026-11-08T16:58:24.3141461+04:00, Login observed within30seconds.
Ordinary Alt+F4 requested17:09:33.125+04:00 (11.15minutes). Exact PID absent in an
authorized read-only check outside the restricted context; a subsequent window
enumeration returned no Talenro window. No kill, retry, or additional launch.
This allowance1/1 consumed; cumulative5 requests,4 observed closed sessions and
one historical interrupted request. No code/test/build/restore/settings changes;
prior159/159 was not rerun. Screenshots inspected live, not saved or committed.

### PASS — observed paths, not full keyboard/accessibility acceptance

- Login: visible initial Enter focus; Tab → Reset password → Registration →
  Login scenario. Shift+Tab from scenario → Registration → Reset password →
  Enter. Lower controls automatically scrolled into view. Fixed email/password
  remained text in the tree, with no editable credential control in this path.
- Registration: keyboard Return opened the offline explanation; Escape closed
  it and the visible focus outline returned to Registration.
- Login reset: Down on scenario selected invalid input, disabled entry and showed
  the simulation explanation; Home restored normal and enabled entry. Entered,
  visited Account, keyboard-activated Exit. Login then visibly showed normal
  selector, enabled entry and blank global notice. This does not claim that a
  nonnormal scenario survived entry: normal was deliberately restored first.
- Account: from the Account sidebar button, Tab → Settings → one unidentified
  focus stop (see below) → MFA → Manage devices → Language → Delete placeholder
  → Exit. Lower actions automatically scrolled into view. Return on Manage devices
  opened the explanation; Return closed it and visible focus returned to that
  button. Delete was only traversed, not activated in this session.
- Connection: visible sidebar sequence Connection → Nodes → Traffic → Devices →
  Account → Settings; then unidentified stop → labelled primary Connect → Choose
  location → Country lock → Success → Failure → Locked-all-failed → Disconnect
  completion → Entitlement selector. Shift+Tab from selector returned to Disconnect
  completion. Lower controls auto-scrolled. No extra actionable power-glyph stop
  was identified; the unidentified stop prevents a comprehensive focus claim.
- Nodes: sidebar Settings → unidentified stop → visible Search caret → Auto row.
  Down selected jp-1, skipping JP heading; Down selected de-1, skipping disabled-1
  and DE heading; Up returned jp-1. Tab → Back → Apply → List scenario; Shift+Tab
  returned Apply. Lower actions auto-scrolled and disabled rows were not selected.
- Confirmation setup only: applied jp-1, explicitly simulated success, drafted
  de-1 with lock off. Keyboard Return on Apply opened the restart confirmation.
  Initial visible focus was Cancel. Tab cycled Cancel → Confirm → Cancel;
  Shift+Tab cycled Cancel → Confirm → Cancel. Escape closed the dialog and visible
  focus returned to Apply. No cross-country confirmation was submitted.

### Unresolved observation — do not silently mark passed

After Settings, the first Tab into each authenticated page had no identifiable
focus outline in the bounded screenshot; the next Tab reached MFA/Connect/Search
respectively. On Account a second read-only snapshot still did not identify the
intermediate focus. Metadata repeatedly reported the root pane or stale Search
even when another control visibly had focus. Therefore metadata alone cannot
identify this stop or establish whether it is an app focus defect, a container
focus presentation issue, or an observation limitation. No confirmed app defect
was established; whole-page focus visibility remains UNVERIFIED.

Next recommended step: scoped read-only source inspection of focusable content
containers and navigation (no new launch/build/settings change) before proposing
any repair or additional runtime check. Full reverse traversal of every control,
all modal variants, screen reader, contrast, enlarged text, target viewports,
RTL/English and design comparison remain unverified; broader runtime/security/
storage/atomicity acceptance unchanged. Task11/12 and44/34/6/217 NOT_ACCEPTED.

Ruling: this approved observation-only plan uses the portable evidence ledger,
not implementation/TDD/build or delegated review machinery; it authorizes no
code repair. Some Account and Login focus checks were performed while completing
the reset sequence to avoid repeating navigation. Cost/limit: no implementation
or independent accessibility acceptance is inferred. All prior records below
remain historical, superseded only by the explicit observations in this entry.

## Latest gap-only session, 2026-11-08 — PARTIAL

User approved the one-launch/20-minute gap plan after14d50ff8. The allowance is
consumed; no further launch or retry is authorized by this record. Preconditions
confirmed ordinary token (AdministratorToken=False), no existing demo process or
window, clean worktree and unchanged App/Core source against0d8a436f. Executable
path/hash are the same as recorded below; this does not attest adjacent files.

- PID17108, start2026-11-08T16:33:04.8659079+04:00. Startup observed within30s.
- Ordinary Alt+F4 close at about16:50:38+04:00, elapsed17.6minutes, before the
  minute18 scenario cutoff and20-minute session limit. Read-only process check
  outside the restricted context confirmed PID17108 absent; repeated window
  enumeration confirmed no Talenro window. No kill, restart or retry launch.
- Cumulative accounting: four launch requests, three observed closed sessions,
  including the historical interrupted request. This allowance used1/1.
- No source edit, automated test/build/restore, installation or settings change.
  Earlier159/159 and compilation evidence were not rerun. No confirmed app defect
  was observed in the exercised paths; this is not full acceptance.

### Additional PASS — native observations only

1. Auto/no known country refuses country lock and leaves it off, with a reason.
2. Expired, exhausted, unknown, stale and device-limit each actually refuse the
   Connect button: remain disconnected and display the corresponding reason.
   Account: expired12.4/100GB and2/5; exhausted100/100GB and2/5; unknown/stale
   usage and devices unavailable; device-limit12.4/100GB and5/5 activated devices,
   explicitly not concurrent. Returning through logout restores active entry.
3. Disabled node input leaves Auto selected. Loading, unavailable and explicit
   empty scenarios each show an empty list, matching explanation and disabled
   Apply; normal restores the list and enables Apply.
4. With jp-1 connected and JP locked, draft de-1 Apply explicitly names JP→DE,
   saved lock change and simulated restart. Default Return cancels; inspection
   shows connected/JP lock and reopening Nodes shows jp-1 selected. Independently
   repeating with the Cancel button preserves those same values. Confirm enters
   Connecting with DE lock; later Nodes inspection shows de-1 selected. These
   observations do not establish hidden atomicity or internal dispatch counts.
5. Explicit locked-country failure followed by Retry locked country enters
   Connecting with DE lock. Failing again then Stay disconnected leaves it
   disconnected, preserves DE lock, clears the recovery offer and notice.
   Together with the earlier one-time recovery this covers all three choices.
6. Registration, sidebar Traffic/Devices/Settings, and account Manage devices,
   Language/Delete account all show the offline unimplemented explanation;
   Escape dismisses. These are placeholders, not real account/settings actions.
   Earlier reset-password/MFA observations remain applicable. No full focus-return
   or assistive-technology claim is made for these additional dialogs.
7. Deliberately set Nodes search JP and list Loading, then logout. Login returns
   normal with blank notice. Reenter: disconnected, lock off; Nodes screenshot
   confirms empty search, normal full list and Auto selected (no saved DE choice).
8. Pointer wheel scrolling visibly moves Connection, Account and Nodes content;
   selected keyboard dropdown/navigation paths work. No viewport-size assertion.

One native dropdown extended outside the main window. Attempting to click its
item returned an outside-window geometry error. A refreshed observation followed
by keyboard selection reached Loading successfully. This is recorded as a tool
targeting limitation, not an established application defect. No launch retry.

### Remaining UNVERIFIED / next scope

- Change Login scenario before entry, return it to normal, then logout and inspect
  its reset: this particular sequence was not performed in the gap session.
- Full keyboard order and focus return, screen-reader behavior, measured contrast,
  enlarged text, target1024x768/1440x1024 logical viewports, RTL/long content,
  English fallback and final design comparison. Do not change settings implicitly.
- Hidden atomicity, broader runtime/network/storage/security acceptance remain
  outside this native observation scope. Task11/12 and44/34/6/217 NOT_ACCEPTED;
  no main merge. Next agree a bounded remaining-check scope, not repeat core flows.

The following session records are historical. Their then-unverified gaps are
superseded only by the explicit additional evidence above.

## Earlier observed sessions, 2026-11-08

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
