# Read-only comparison: historical stall versus complete close trace

No new launch, build, test run, service/settings change or application edit.
Compared saved records after76f9e3f8. Existing runtime budget remains exhausted.

## Evidence matrix

| Dimension | Historical automatic stall | First close-stage success |
| --- | --- | --- |
| Primary run | 2026-11-09 06:06Z, PID8516; earlier PID1772 also stalled | 2026-11-09 17:49Z, PID12356 |
| Application | Preserved layout-diagnostics build | New isolated close-stage-recording build |
| Profile/language | long-country / zh-Hans | baseline / zh-Hans |
| Route | Enter demo, Nodes, SearchBox, jp-1 | Same visible route and literal text |
| Input | sky.type_text once | sky.type_text once |
| Close | Alt+F4 | Click titlebar close button |
| Stage instrumentation | None; handler position unknown | 24 valid stages, CloseReturn and healthy footer |
| Native observation | Original retained handle, about5s samples plus bounded outer wait | Original retained handle, about2s samples, no outer WaitForExit wait |
| Native outcome | Wait258, successful code0, zero exit FILETIME until360s cap | Signaled, successful code0, nonzero exit time |
| Managed outcome | Waitfalse/HasExitedtrue/code0/1601 time disagreement | Waittrue/HasExitedtrue/code0/valid time agreement |
| Wait-chain evidence | User screenshot -> svchost9888/thread1812 | Not captured; run exited normally |
| Environment identity | Ordinary token; exact OS/IME state not captured | Ordinary token; no matched OS/IME/service snapshot |

This is NOT an equal-conditions reproduction of the historical failure. Input
method is shared, but profile, artifact/instrumentation, close gesture, observation
cadence and elapsed system state differ. Automatic input is not sufficient to
produce the failure on every observed configuration. No probability/reliability
estimate follows from these few non-randomized runs.

## Source-backed differences

Git comparison cf0f605d..e347d8cd has no changes in NodesPage.xaml.cs,
LayoutPresentation.cs, Core, App packages.lock.json or the three native/managed
observer module/source files. Application/config changes are the recorder,
conditional hooks/startup and isolation gates. Unchanged declarations/lockfile do
not prove identical loaded native dependencies or a matching host environment.

LayoutPresentation.Text changes CountryJP and NodeJp1 display text for long-country;
NodesPage.Refresh/AddNode consume those labels. The search predicate remains based
on original node id/country/localized country text; long-country does not replace
the predicate with the long display label. The profile can alter rendered content
and layout, but source inspection does not establish a shutdown causal mechanism.

The recent baseline success therefore does not exclude a profile-specific
interaction, nor prove that recording instrumentation fixed or masked a defect.
The titlebar versus keyboard close difference is likewise an untested candidate,
not evidence that either action causes the stall.

## Historical evidence retained

Adjacent automatic long-country input stalled while physical manual input with
the preserved artifact/profile exited normally. This supports an association only:
fixed ordering, focus/IME/timing and intervening state were not controlled.
Single-TextBox and list-refresh prototypes also had successful automatic runs,
but are different applications and do not clear the original failure.

Screenshots show waits on svchost PID9888/thread1812. Later service enumeration
mapped that host to cbdhsvc_8b240 with stable host creation time. The mapping is
later evidence, not an exact capture-time service/thread identity. It does not
identify a blocking API, establish type_text clipboard transport, or prove a
service defect. Do not disable or restart that service as a speculative repair.

Additional read-only verification of the retained recent49 samples reconstructed
each native creation FILETIME as UInt64(high)*4294967296+low. All49 exactly match
the retained managed launch time17:49:23.1779168Z. This strengthens identity checks
without adding a launch or changing any receipt. Current evidence still does not
retroactively locate the historical stalled close handler.

## Recommended next discriminator (not executed / new approval required)

If another bounded observation is approved, use the SAME close-stage EXE/DLL and
observer procedure as the successful baseline, changing only the layout profile
from baseline to long-country (a fresh correlation id/output file is bookkeeping).
Keep automatic jp-1, titlebar close, ordinary token, same retained-handle sampling,
no node selection/apply/connect action, one attempt and360s absolute cap. Verify
the long-country title and actual field value before normal close. Record exact
command/timings and stop on anomalies; no retry, kill or settings changes.

Do not simultaneously switch to Alt+F4: that would change two intended factors.
Even this later comparison is not randomized or environment-identical. Success
would show that one long-country route can exit on the instrumented build, not
clear intermittent failure. Failure with full valid CloseReturn narrows the stall
to after/outside that handler for that run. A partial/missing trace is inconclusive
about the last executed boundary and does not prove a particular removal blocked.
Any wait-chain handoff must concern only that run's verified PID; never end tasks.

Do not rewrite shutdown behavior, remove subscriptions, add forced exit, or issue
an external defect report as if root cause were known. All broader acceptance
blockers remain. This document itself grants no additional execution budget.

## Source records

- [Historical wait chain](2026-11-08-windows-wait-chain-observation-execution.md)
- [Adjacent automatic/manual runs](2026-11-09-windows-adjacent-input-comparison.md)
- [Historical reproduction report](2026-11-09-windows-input-shutdown-reproduction-report.md)
- [First complete trace](2026-11-09-windows-close-stage-first-observation.md)
- [Single TextBox](2026-11-09-windows-textbox-prototype-observation.md)
- [List-refresh variants](2026-11-09-windows-list-refresh-comparison.md)
