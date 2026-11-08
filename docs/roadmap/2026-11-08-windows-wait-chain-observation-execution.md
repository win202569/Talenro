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
