# Keyboard-only follow-up — scoped observations, 2026-11-13

New explicit authority: one launch, zero retries, maximum six minutes, continuous
user cooperation. Baseline 56d84130c48454f2dc29af2c9842efdb1931e682, existing
account-center entry and candidate unchanged. No source edits/build/test rerun.
Run `5b14bb63977d4ceb98cf0ea0a1cd2afb`, PID24888. Earlier timed-out run remains
preserved; this result does not rewrite its outcome.

## Observed keyboard checks

User confirmed visibility and resizing. Agent observed Login at650x620 DIP,
navigated Enter demo to Connection, verified root650x620/page430x534, then Account.
Language en-US, profile rtl-mixed, render/text scales1. No pointer scrolling,
scrollbar operation or programmatic scroll was used. Initial Account scroll0.

1. Tab visibly outlined Settings, MFA, Manage devices, Language, Delete account,
   then Exit demo and reset. Content moved automatically to expose lower buttons;
   settled final scrollbar value172, with full Exit label and outline visible.
   Intermediate accessibility values can lag animations (observed68/120).
   Shift+Tab visibly returned through Delete and Language to Manage devices.
   Delete and Exit were focused only, never activated in this run.
2. Return on Manage devices opened the Offline demo explanation. Close had a
   visible keyboard focus outline. Return closed it; after animation settled,
   no dialog remained and Manage devices visibly regained its outline.

Both narrowly specified keyboard observations pass on this artifact/configuration.
Accessibility focused-element metadata remained pane2 and was not treated as
reliable focus identity; conclusions use rendered outlines and visible behavior.
This does not establish immediate outline behavior after *mouse* dialog closure,
all dialogs, screen-reader behavior, other sizes/scales, or ordinary builds.

## Closure and receipt review

Titlebar X used once after checks. Console outcome sequence114 at222319.048ms:
exited/agreement, ExitVerified=true, Finalized=true, OverrunMs0; command exit0.
Persisted pre-cleanup outcome is222292.7498ms, Finalized=false; preserved unchanged.
Successful subsequent CIM query found0 matching demo/prototype processes.
All114 process rows have this RunId and continuous sequence1..114.
Geometry contains Login2/Account2; health Samples4, Provisional=true, Failed=false,
ErrorCodes[]. Initial geometry samples are not evidence of later keyboard actions.
Original GeometryEvidenceStatus=Unreviewed and Acceptance=NotAccepted remain intact.

Manifest EXE SHA256: CF7C577C2FF3F4A690B2EFEE054411B59006210440333C5E8EA1838A7062A48F

Manifest DLL SHA256: 6011FA7BB63A45D5FAEB4A44529904FEC0857741A4E810B20B38555655C838C1

Raw files remain local/ignored beneath the fixed account-center win-x64 directory:

| Relative file | Bytes | SHA256 |
| --- | ---: | --- |
| email-geometry-observations/5b14bb63977d4ceb98cf0ea0a1cd2afb/manifest.json | 730 | 4B6CD560BC9CA2662B5D043021596EC09E9A5FFA479D09319F18DD790B41E76E |
| email-geometry-observations/5b14bb63977d4ceb98cf0ea0a1cd2afb/samples.jsonl | 214879 | 9885F67FFFE9650F0FEB08F578550EB2DF2D66FC0E27C73FF3F234EBBA693BE9 |
| email-geometry-records/5b14bb63977d4ceb98cf0ea0a1cd2afb/geometry.jsonl | 36073 | 066E7ED7996A2F8BBAF37CA16801A9334D46B845761E81986B9021C2F82E786D |
| email-geometry-records/5b14bb63977d4ceb98cf0ea0a1cd2afb.claim | 0 | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Live allowance0. No retry, force-kill or additional scenario. C4/overall remain
NOT_ACCEPTED. Next: read-only consolidate remaining ordinary-build and layout
coverage; no automatic launch/build. Fresh clone recovers this report, not raw
receipts, screenshots or binaries. Historical timeout and shutdown gaps remain.
