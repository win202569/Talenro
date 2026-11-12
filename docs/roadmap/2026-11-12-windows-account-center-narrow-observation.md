# Narrow-window observations

## Latest: separately approved visible-window run obtained the condition

After the first run below, user reported not seeing its window. A new explicit
one-launch/zero-retry/up-to-six-minute approval was obtained, with visibility
confirmation before proceeding. Baseline faf8476b. New run
`f86a7c0d8cad4a02a6fd307c1652c790`, PID26792 (now exited).

The first inspected page was already Connection; the agent did not navigate to
it in this run and does not attribute that transition to a specific action.
After activation the user confirmed seeing the window, then manually resized it
and confirmed completion. Two title observations more than one second apart
showed root640x620/page420x534 DIP, rendering/text1. Account was first visited
after this confirmation; no subsequent resize/scroll/navigation. The unobscured
Account screenshot showed the complete demo@example.com. An occluded screenshot
while the user replied was not treated as layout evidence; activation restored
an unobscured target capture before Account navigation.

Normal title-bar close: console sequence100, elapsed193911.7273ms,
Outcome=exited, Reason=agreement, ExitVerified=true, Finalized=true, OverrunMs=0,
command exit0. Persisted outcome elapsed193890.3365ms retains Finalized=false
(pre-cleanup). UI target window absent and successful CIM query matched0 targets/
prototypes afterward. No relaunch, forced termination, source/build/settings change.

### Read-only evidence check

Manifest language/profile and actual artifact hashes matched the approved fixed
binding. All wrapper and snapshot run IDs correlate;100 process sequences are
contiguous. Six geometry records: header, Login1/2, Account3/4, health. Samples4,
Failed=false, ErrorCodes empty, Provisional=true. Login is NOT narrow evidence.

Both Account samples have root640x620, viewport420x534, scales1, offsets0. Panel
Center width360, RootDIP horizontal250–610 (30 DIP margins within presenter220–640).
Email RootDIP383.095703125–570, vertical141–167.6015625, unchanged size
186.904296875x26.6015625. All three recorded boundaries have usable matching-space
bounds and Complete=true; insets in BOTH samples:

| Boundary | Left | Right | Top | Bottom (DIP) |
| --- | ---: | ---: | ---: | ---: |
| root | 383.095703125 | 70 | 141 | 452.3984375 |
| presenter | 163.095703125 | 70 | 98 | 409.3984375 |
| presenter:clip | 163.095703125 | 70 | 98 | 409.3984375 |

Limited result: narrow condition obtained; email visible and contained in known
boxes. ClipChainComplete=false remains, and samples are only17.0264ms apart: no
full compositor-clipping or long-duration claim. RootDIP right is not desktop right.
Panel height726 exceeds the viewport, so whole-page fit and lower controls are
not accepted by this email-only observation. Raw Unreviewed/NotAccepted remain;
C4, LTR/other scales/full regression/security acceptance remain outstanding.

### New run immutable local inventory

Relative to the same candidate output directory specified below; RUN here means
f86a7c0d8cad4a02a6fd307c1652c790. Old run records remain untouched.

| Path | Bytes | SHA256 |
| --- | ---: | --- |
| email-geometry-observations/RUN/manifest.json | 730 | 30B94F0CCAE9E7477F6DB4B03248D37B3E80578FC72CDF56D8AF261DFCC82702 |
| email-geometry-observations/RUN/samples.jsonl | 187957 | 70C2545FCF26AABFCFA6A5BA1A9804C8474CE5D0B42C449C018C9424C80407B7 |
| email-geometry-records/RUN/geometry.jsonl | 36074 | 20DC34827A610A6775CE746027B1E2A7B22635AA830DF7D52A17E05B1C447347 |
| email-geometry-records/RUN.claim | 0 | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

The separately granted live allowance is consumed (0). Next: consolidate remaining
acceptance gaps without another launch; no further source change is established
as necessary by this result. Clone restores this summary, not local evidence.

## Earlier run: condition not obtained (preserved history)

User approved the checklist, then explicitly approved one launch, zero retries,
up to six minutes with manual resize assistance. Source baseline433adcb1.
Run `ef33d39285174f1095ac69feae7e6da5`, target PID9832 (now exited).
No relaunch, build, code change, forced termination or settings change occurred.

## Result

Narrow condition INVALID / NOT_OBTAINED; layout result INCONCLUSIVE. The window
remained root1136x620 DIP, rendering/text1. Connection page was confirmed at
page916x534 DIP, simulated disconnected. No Account navigation was performed.
The user was asked to resize to root620–660 wide with height unchanged, but no
completion reply or changed dimensions were observed before normal closure.
This does not establish that resizing is impossible or that narrow layout fails.

Initial accessibility click reported coordinate input geometry unavailable; a
fresh observation and activation recovered the target window. A later Enter demo
click was followed by Connection being observed. A screenshot initially showed
an occluding window; it was not used as layout evidence. No new process was started
to recover UI access. Title-bar close was performed once near the end of the budget.

Console outcome sequence156: elapsed306259.0091ms, exited/agreement,
ExitVerified=true, Finalized=true, OverrunMs=0; command exit0. The saved outcome
has elapsed306245.0755ms and Finalized=false (pre-cleanup), kept unchanged.
Post-close UI window list empty; successful CIM query matched0 target/prototype
processes. Ordinary token, exact candidate hashes and six source blobs matching
HEAD were checked before authorization; the fixed entry repeated its preflight.

Geometry: four records (header, Login snapshots1/2, health), Samples2,
Provisional=true, Failed=false, ErrorCodes empty. Account samples0, so the required
two narrow Account snapshots do not exist. No Account inset comparison or visual
acceptance is possible. Raw Unreviewed/NotAccepted values are retained; C4 remains
NOT_ACCEPTED. Normal process exit is not layout acceptance.

## Immutable local inventory

Paths below are relative to the candidate output
`clients/windows/Talenro.Demo.App/bin/email-geometry-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/`.
In the table, RUN means ef33d39285174f1095ac69feae7e6da5.

| Path | Bytes | SHA256 |
| --- | ---: | --- |
| email-geometry-observations/RUN/manifest.json | 730 | 59FAE0C79A55E3DB2DF7DC18105C54DE1ABC34FD8433BC90A57341C0D2F5C43B |
| email-geometry-observations/RUN/samples.jsonl | 295429 | 796CDA69F4CDC1CF7A4A29C23CFDADF164E6871D4BDB097D4006E74D1B92E6E5 |
| email-geometry-records/RUN/geometry.jsonl | 18150 | 1CEF8A930A2F9E98E8A5E056B51ED1456339A6CA50FE3A4996B5896DCD6A3CEF |
| email-geometry-records/RUN.claim | 0 | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Manifest EXE/DLL bindings match the approved design's candidate hashes. Raw files
remain local/ignored and unmodified. Clone restores this sanitized account, not
receipts or binaries. No product tests rerun for this observation/documentation.

## Next step

Live allowance consumed (0 remaining). Before requesting any new run, establish
whether manual resize assistance is available or prepare a separately reviewed
normal-window-control resize procedure. Do not repeat the failed coordination by
automatically relaunching. Preserve this run and all previous evidence.

See [approved design](../superpowers/specs/2026-11-12-windows-account-center-narrow-observation-design.md)
and [checklist](../superpowers/plans/2026-11-12-windows-account-center-narrow-observation.md).
