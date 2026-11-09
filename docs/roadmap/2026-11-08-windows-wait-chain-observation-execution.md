# User-assisted wait-chain observation

User confirmed Analyze wait chain menu exists and explicitly approved one new
demo attempt plus manual wait-chain inspection afterc2253b14. This is a new
budget1 attempt, not reuse of the exhausted paired-observation allowance.
Reuse unchanged artifact/recorders and paired-observation safety limits:30s close,
360s absolute cap, no retry/kill/dump/debugger/settings changes. User inspects only
the exact demo PID supplied after close failure; no unrelated process diagnosis.
Preflight ordinary token, demo count0, no demo window, clean worktree and EXE/DLL
hashes match prior plan. Task Manager already open. Budget0/1 before launch.
Close timing is recorded in UI UTC; no exact shared monotonic-close claim.

## Current attempt

- Baseline c2253b146a883cbfe605c2751bb1ed59ae5e4ef1; unchanged artifact and
  recorder identities match the paired-native-observation plan.
- Attempt 1/1 consumed. Launch ID dfe89109bba445ff90d2a8eec82864b7;
  demo PID 1772, observer PID 12200; start 2026-11-08T19:10:03.5392494Z.
- Enter demo -> Nodes -> search jp-1 completed; stable filtered group/node
  observed. No selection, apply, connection or dialog action.
- One Alt+F4 requested at 19:12:21.575Z. Later window inventory found no demo
  window. Native wait remained 258, successful exit-code API returned 0 and
  successful process-times API returned raw exit FILETIME high=0/low=0.
- At 19:13:14.2318740Z independent CIM query still found matching demo PID 1772
  (creation 19:10:03 UTC), one thread, and observer PID 12200. This is after
  the 30-second close criterion; shutdown remains FAIL/UNRESOLVED.
- User asked to inspect Analyze wait chain for exactly PID 1772 and return its
  screenshot/text, without ending a process or creating a dump. Result pending.
- Observation ended at the original 360-second cap; no additional launch budget.

## Final observation and handoff

73 pairs had valid creation identity and native API reads: 28 before the close
request and 45 after it. Before close, native wait=258 and code=259; after close,
all 45 pairs had native wait=258, successful code=0, successful raw exit time=0.
Managed post-close fields were wait=false, HasExited=true, code=0 and exit time
1601-01-01T00:00:00Z. Initial pairs 1-2 remain in the launch tool receipt;
subsequent poll receipts cover pairs 3-73. No raw receipt/screenshot is committed.

Final pair started 19:16:03.5546480Z, elapsed 360.0183617s. TIMEOUT followed at
19:16:03.5563769Z, elapsed 360.0187623s, Killed=false. No signal or abort occurred.
Observer exited 0 after disposing its original Process; that host exit code is
not application shutdown acceptance. At 19:16:16.8714064Z a separate CIM query
found neither recorded PID, and demo name count was 0. No causal claim follows
from disappearance after observation ends.

Search/filter PASS; shutdown FAIL/UNRESOLVED. Wait-chain evidence remains pending:
the menu was user-confirmed, but no actual chain screenshot/text has been received.
Do not label this a completed WCT capture or infer no waits/deadlock from missing
evidence. Next review the user's saved result; if unavailable, record unavailable
without another launch. Any new runtime attempt requires separately scoped consent.
No source/settings changes, rebuild, test rerun, debugger, dump or forced termination.
Task11/12 and 44/34/6/217 NOT_ACCEPTED boundaries remain unchanged.

## Follow-up 2026-11-09 — screenshot and service identity

Supersedes the pending-evidence handoff above. User supplied a Task Manager
Analyze wait chain screenshot: Talenro.Demo.App.exe PID1772/thread10816 waits on
svchost.exe PID9888/thread1812. User confirmed clicking Cancel, not End process.
The screenshot filename is dated 2026-11-08 231523; this is not an independently
verified capture timestamp. No raw screenshot is added to Git.

At 2026-11-09T05:22:34.5965105Z a read-only CIM query found PID9888 hosting exactly
one returned service, cbdhsvc_8b240, Running. Process name/path were svchost.exe /
C:\WINDOWS\system32\svchost.exe. Creation UTC before and after service enumeration
was identical: 2026-11-08T17:11:36.0749030Z, earlier than the demo observation.
This supports process continuity rather than a post-observation PID reuse; service
membership was queried later and is not a historical service-membership snapshot.
The local cbdhsvc service template ImagePath uses ClipboardSvcGroup. No clipboard
contents were read. An app-source search for Clipboard/cbdhsvc returned no matches;
that does not exclude indirect framework/automation involvement. The documented
type_text contract does not establish its internal clipboard behavior.

The evidence narrows investigation to a clipboard-service-related wait, not a
proven service defect, deadlock or application root cause. No stack/operation is
identified and thread1812 lifetime was not verified. Next perform read-only review
of the input/shutdown path and diagnostic options before proposing a discriminating
test; do not stop/restart the service or change clipboard/security settings.
No demo launch, debugger, dump, build or tests in this follow-up. Runtime allowance
remains exhausted; shutdown and all broader acceptance blockers are unchanged.

## Read-only input/shutdown review, 2026-11-09

Re-read App.xaml.cs, MainWindow.xaml.cs, Views/NodesPage.xaml(.cs),
Diagnostics/LayoutWindowObserver.cs and the Nodes diagnostic bindings.
SearchBox is a standard TextBox; TextChanged calls Refresh synchronously.
Refresh trims text, filters fixture nodes, clears/repopulates list items and uses
the rendering guard against selection dispatch. Nodes diagnostic bindings change
Header/PlaceholderText, not Text. No direct clipboard call or app-owned wait was
found on this path. App.Exit is used for rejected startup arguments, not normal
close. The diagnostics Closed handler unsubscribes events through Dispose; source
alone does not prove that callback completed in the observed process.

The bundled computer-use API/guidance documents type_text as literal-text input
but does not document its internal clipboard transport. Do not infer paste use,
service fault, framework fault or a fix from this review. No implementation edits.

Suggested next discriminator (NOT EXECUTED / approval required): one new launch
with unchanged artifact, arguments, original-handle paired observer and safety
checks; same Enter demo -> Nodes route, but user physically types jp-1 without
paste instead of type_text. Confirm exact value/filter, then one normal close.
Keep startup30s, route120s and absolute360s cap; if user is not ready or route
deadline is missed, do not extend/retry. Close criterion remains30s; no signal
continues read-only to cap. Any WCT handoff targets only that run's verified PID.
No clipboard reads/clears, service restart, settings changes, debugger/dump or kill.
Single success would only motivate the input-path hypothesis, not prove it or
clear acceptance; repeated failure would show automatic text entry is not necessary
for that failure under the tested conditions. Changed timing and intervening system
state limit comparison with the historical run. No new launch allowance yet.
