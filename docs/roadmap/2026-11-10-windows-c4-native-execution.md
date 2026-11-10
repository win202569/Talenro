# C4 English/RTL native observation — display defect found

Date2026-11-10, Asia/Tbilisi (UTC+04:00). Baseline818f1604.
User renewed one launch, zero retries,360s observer cap, startup30s, normal close
requested by300s and exit assessment30s. Ran unchanged fixed observe-c4.ps1 with
RunApproved once in the approved ordinary host context; all existing gates retained.
No settings changes, build, test rerun, product edits or force-close.

## Identity and bounded execution

- RunId `f6071741fab546ef8f856467def00375`, target PID25136.
- Manifest: C4, en-US, rtl-mixed, exact fixed command arguments and both approved hashes.
- Process start09:14:46.2693107 UTC (13:14:46.2693107 local).
- Login screenshot/accessibility observed by09:15:02.858 UTC, about16.6s after start.
- Normal titlebar close requested09:17:14.031 UTC, about147.76s after process start.
- Native and managed exact exit09:17:14.2823079 UTC: approximately251.31ms after
  the pre-click timestamp; the input call itself completed before this exit time.
- Original-handle native wait signaled, native exit0, managed WaitSignaled/HasExited
  true and exit0, exact agreeing exit timestamps.
- Terminal observer result: Outcome=exited, ExitVerified=true, Finalized=true,
  ElapsedMs149851.9463, OverrunMs0; PowerShell exit0.
-78 raw rows, sequence1..78 contiguous;75 complete identity-verified native/managed
  pairs, zero incomplete pairs; all run IDs match. Raw outcome Finalized=false as
  designed; terminal post-cleanup result separately finalized.
- Follow-up supported window inventory empty for demo/prototypes; authoritative CIM
  target inventory count0. EXE/DLL SHA256 unchanged from manifest/approved policy.

The shell invocation reference09:14:38 UTC includes approval/tool/preflight time;
it is not the process start. UI deadline checks used that conservative earlier
reference. Actual observer clock is the retained scope's independent duration above.

## UI evidence and separate verdicts

Evidence is direct supported-tool screenshots and accessibility trees in this session,
not inferred solely from source, titles or an observer exit code. Immediate snapshots
sometimes preceded navigation/scroll completion; refreshed stable state before
judging. No search text, node selection/apply or connection action was performed.

| Check | Scoped verdict | Evidence |
| --- | --- | --- |
| English labels and persistent offline notice | Observed pass | Login, Connection, Nodes, Account displayed English and no-connection/submission notice |
| Visible mirrored layout | Observed pass | Right-side navigation, right-aligned ordinary labels, scrollbar on left |
| Login actual-email LTR | Observed pass | Separate DemoEmail showed complete demo@example.com left-to-right, not merely the heading sample |
| Node technical identifiers LTR | Observed pass | jp-1, disabled-1 and de-1 visible left-to-right; no node selection |
| Exact synthetic heading codepoints | Observed pass for exposed text | Extracted heading equaled fixed fixture including U+202E/U+202C/U+2066/U+2069 |
| Account actual-email readability/clipping | FAIL | At stable scroll-top Account view, email starts beyond left window boundary; only its trailing portion was visible |
| Profile retention after logout/re-entry | Observed pass | Return login retained same mixed heading and English; re-entry retained right-side navigation |
| Selected visible reset state | Observed pass, limited | Login ordinary fixed-account notice; re-entered Connection disconnected, country lock Off |
| Overall clipping/layout acceptance | NOT_ACCEPTED | Account email defect; full viewport/layout matrix not performed |

Exact exposed heading matched escaped fixture:
`\u202eRTL synthetic label\u202c \u2066demo@example.com\u2069`.
The accessibility tree preserved these controls in this run (not stripped); this
supports the exposed-text codepoint claim, not natural-language shaping correctness.

## Account email defect

On the Account page at scroll0, screenshot showed only the suffix of the actual
email at the extreme left; its beginning was outside the window. Accessibility
still exposed the entire `demo@example.com`. A second stable observation about9s
later reproduced clipping before any scrolling. Thus full accessibility text does
not establish visual readability. The later scroll preserved the same horizontal
clipping until the email left the viewport vertically.

Observed title dimensions: root1136x620 DIP, page916x534 DIP on main pages,
render1.000/text1.000. These are observed current dimensions, not a claim that the
required target viewport/scale matrix passed. Login page width1136 DIP.
Only vertical scrolling to reach Exit demo and reset was used; no resize or settings
change. Returned login and re-entered demo within the same process, then closed.
No root cause or fix is claimed; next proposal is source-only diagnosis of the
Account email RTL layout/clipping, without another launch or artifact rebuild.

## Scope and recovery

C4 is PARTIAL / NOT_ACCEPTED because of the visible defect. This single successful
normal exit does not resolve the historical shutdown cause. Overall C5, C6,
Task11/12 and wider44/34/6/217 gates remain unchanged. Exhaustive reset, shaping,
screen-reader, contrast, full keyboard, fallback-language and viewport matrix are
not accepted by this run. Launch allowance consumed1/1; retries0, remaining0.

Raw manifest/samples remain locally under the fixed artifact's ignored
`process-observations/f6071741fab546ef8f856467def00375/`. Screenshots/raw receipts,
private desktop inventory and binaries are not committed. This portable report,
status and recovery notes are preserved on the existing development branch; no
main merge. A new computer restores source/report, not ignored raw evidence.
