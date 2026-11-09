# Fresh approved static/rebuild comparison

After1d3e1ce4, user accepted a fresh two-launch comparison during an available
desktop interval. Previous interrupted attempt remains unchanged and NOT_ACCEPTED.
New budget2/2 consumed: static then rebuild, one automatic sky.type_text jp-1 each,
input confirmation within120s, normal close, total observation cap360s, stop on
anomaly. No retries, kill, settings/service changes, clipboard read or demo launch.

## Guards and observation

Each fresh ordinary-token host checked no existing demo/baseline/variant process,
variant EXE/DLL against build-ledger SHA256, and all three unchanged observer
module/source blobs against HEAD. CIM path, explicit mode argument and creation
time matched the original Process/retained handle. No reopening by PID for native
sampling. Both observers used PowerShell7.6.9. Existing modules were not modified.
The session polling sleep was bounded by remaining time and a pre-sample deadline
check added to avoid the prior fixed2s sleep overrun; neither run approached cap.
No exact-deadline scheduler guarantee is claimed.

Each UI route selected its unique returned window, activated it, captured both
screenshot and accessibility, clicked ProbeInput, confirmed focus, typed once,
confirmed actual value (not placeholder), then clicked the observed close button.
Both modes initially displayed jp-1/us-1/de-1 rows; rebuild tree row indexes changed
after input, consistent with recreation, not an instrumented callback count.

| Evidence (UTC2026-11-09) | Static | Rebuild |
| --- | --- | --- |
| Launch ID | 20aab6215a31480abd4976a7c5a116dd | f1f42b6df5da4c5eb6cc9112e0e5544c |
| Target / observer PID | 12096 / 14384 | 21764 / 11952 |
| Start | 16:44:09.3396865 | 16:45:25.3374369 |
| Initial UI observed | 16:44:28.704 | 16:45:47.601 |
| Exact input confirmed | 16:44:48.560 | 16:46:07.896 |
| Close action timestamp | 16:45:00.769 | 16:46:17.233 |
| Recorded process exit | 16:45:00.9892826 | 16:46:17.4651720 |
| First sampled native signal | 16:45:02.2946828 | 16:46:18.7085353 |
| Observer end elapsed seconds | 52.9497992 | 53.364124 |
| Pairs | 27 | 27 |

All27 pairs per run were parsed from retained tool outputs, contiguous1..27,
identity valid, native call statuses ok, no snapshot errors. Final native wait0,
exit code API success/code0, nonzero exit FILETIME; managed WaitSignaled/HasExited
true and exit code0. Native exit FILETIME high/low: static31289449/2845032922;
rebuild31289449/3609791816. Observer processes exited0; no timeout/abort.
Close-to-recorded-exit UTC differences approximately0.220/0.232s, NOT precise
shared-monotonic measurements. Sampling detects signals later at roughly2s cadence.

After static,16:45:18.923Z window inventory empty and next launch's CIM preflight
found no targets. Final16:46:48.372Z target window inventory empty; independent
CIM16:46:53.4966662Z found no variant/baseline/demo or either observer process.
Worktree clean before result documentation; no source or binary rebuild performed.

## Interpretation and recovery

Scoped automatic-input/normal-close routes PASS for these two runs. Stall not
reproduced in either mode. This is a reduced three-row experiment, fixed order,
one run per mode; it cannot rule out intermittent effects, different service/IME
state, original page complexity, selection/resource updates or observer interactions.
Original bug not fixed; broader NOT_ACCEPTED gates and prior deviations remain.
No additional launch budget. Recommended next step: read-only trace of original
layout observer subscriptions/disposal and window close callbacks to choose the
next single-factor boundary; no automatic source changes or new runtime runs.
Portable ledger committed; raw screenshots/tool receipts and binaries not committed.
