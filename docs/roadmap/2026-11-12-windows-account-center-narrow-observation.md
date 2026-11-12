# Narrow-window observation: condition not obtained

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
