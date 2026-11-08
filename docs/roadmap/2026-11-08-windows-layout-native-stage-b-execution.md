# Stage B native execution — plan approved after508475b0

Plan: docs/superpowers/plans/2026-11-08-windows-layout-native-stage-b.md.
User approved14 maximum invocations, no retries. Source checkpoint7e25a4fc;
documentation checkpoint508475b0. Sequential execution, no implementation changes.

Task1 preflight: ordinary token False; clean worktree on the expected development
branch. No Talenro.Demo.App process or window found. Fresh SHA256 reads of both
managed DLLs and both apphosts exactly match all four plan values. UI helper
initialized and list_windows succeeded. Existing system/security windows are not
task targets and will not be manipulated. Monitor/DPI/text-scale evidence not yet
obtained; exact-size claims remain UNVERIFIED.

Ruling: use the approved portable ledger and PowerShell rather than Bash-only
task helpers; no TDD/build/reviewer cycle is added to this verification-only plan.
The approved plan excludes implementation changes and extra test calls. Cost:
observations are not a fresh independent implementation review.

Pre-flight interface: Task1's variant paths/hashes bind Task2/3 process launches;
Task2/3 results feed Task4 status, no shared code changes. Task1 complete.

Final budget:9/14 consumed;5 unused, not a standing launch allowance after handoff.
Stopped at V3 for mixed-language UI. All9 processes exited; final process count0
and UI window inventory showed no demo window. No retries/builds/restores/tests,
source edits, installs, settings changes or main merge.

R1–R6: all exit2, stderr exactly 'Layout diagnostic startup rejected.' plus newline,
stdout empty. R4 marker not echoed. PID/start (host timestamps +04:00,2026-11-08):
R1 18492/19:39:05.6525798; R2 23020/19:39:20.3499166;
R3 8420/19:39:24.2174736; R4 24480/19:39:24.4165759;
R5 18592/19:39:24.6142686; R6 24504/19:39:24.8141464.
Arguments/variants exactly follow plan rows. Rejection exit/message checks PASS;
absence of transient windows UNVERIFIED: fast process exits were not continuously
sampled by UI capture. No claim that an imperceptibly brief window was excluded.

V1: ordinary/no args, PID16284 started19:39:37.2393926+04:00, normal close exit0.
Login/Connection/Nodes/Account observed in Chinese with permanent offline notice,
no diagnostic title. Tab after Account navigation visibly focused Settings.
Basic route/close PASS; exhaustive control traversal not performed.

V2: diagnostic zh-Hans/baseline, PID5068 started19:40:55.8707891+04:00, exit0.
Observed all four pages, Connecting/Connected status, de-1 selection and Chinese
confirmation/Cancel, Account, logout and re-entry to Disconnected with Chinese
labels retained. Tab brought the offscreen simulated-success button into view.
Readout root1136x620; Login page1136x534, authenticated page916x534, both render1
and text1; logout/re-entry updated page width. These are self-reported, not
independently calibrated. One corner-drag attempt produced no observed resize:
resize-update and exact-size checks UNVERIFIED, not PASS. Available UI API has no
independent client/DPI measurement; no helper or settings change introduced.
One offscreen ExitDemo click returned 'element170 has no cached bounds'; no app
error. Reobserved, scrolled Account, then clicked visible ExitDemo successfully.
Early immediate accessibility snapshots lagged screenshots; refreshed before
depending on changed controls, later used300ms settle before capture. No relaunch.

Ruling: retain the offscreen-input and resize evidence limitations without calling
them application defects or rerunning a process — existing supported scrolling
recovered the target; exact dimensions remain unverified. Cost: no resize proof.

V3: diagnostic en-US/baseline, PID10104 started19:45:20.2884220+04:00.
Login displayed English heading/buttons/account labels and offline notice. Entering
Connection displayed English sidebar, status and 'Country lock (simulated)', but
the switch's off-state text visibly remained Chinese '关'. This is an application
control, unlike OS-owned title-bar accessibility labels, and violates selected-
language presentation. V3 FAIL; immediately stopped further evidence actions and
all later launches. Normal Alt+F4 close returned exit0. V3 Account/Nodes/dialog/
logout routes NOT_RUN. Title root1136x620/page916x534, render1/text1 self-reported.

Read-only source pointer (not a confirmed complete root-cause analysis):
ConnectionPage.xaml declares CountryLock without explicit OnContent/OffContent;
LayoutStaticBindings.cs maps CountryLock.Header only. No fix attempted. Next
recommend a separately scoped diagnosis and test-first correction covering the
switch's framework-provided state content in both languages, followed by bounded
build/native revalidation; never simply suppress this finding or change OS locale.

| Rows | Outcome |
| --- | --- |
| R1–R6 | PASS exit/message; transient-window absence UNVERIFIED |
| V1 | PASS basic four-page route/ordinary title/normal close; limited focus sampling |
| V2 | PASS observed Chinese route/dialog/reset; resize/exact-size UNVERIFIED |
| V3 | FAIL mixed-language switch state; remaining route NOT_RUN; close exit0 |
| V4–V8 | NOT_RUN due to V3 stop condition |

Task2 completed exit/message checks with explicitly limited window evidence.
Task3 stopped/incomplete. Task4 portable handoff only, not acceptance. Prior222/222
automated tests and Stage A builds are historical and do not negate this native
finding. No native all-clear, no full accessibility claim, broader NOT_ACCEPTED
unchanged. Screenshots were directly observed via the UI tool, not copied into Git.
