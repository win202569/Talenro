# Approved long-country / Alt+F4 close-stage observation

User accepted the proposed next step aftera9e04843: one launch, at most360s,
same long-country trace artifact/input/observer, change only intended close action
from titlebar button to one Alt+F4. No retry/kill/settings changes. Budget1/1 used.
No code edits, rebuilds, services, clipboard-content reads, debugger/dump or merge.

## Preparation and identity

Clean a9e04843 worktree; ordinary token; no demo/prototype process/window. Three
observer module/source blobs matched HEAD. Same trace artifact hashes checked
inside launch host:
EXE EC9E4375948F2E7AA36B34DF32F36A5374F68F61C6A1A4FE42A8108CD0255774;
DLL 7E14F4A771DB0F7024C95E95A4193A16448D7B444C8D11273717E8B15F46611B.
Output directory remains clients/windows/Talenro.Demo.App/bin/close-stage-recording/
x64/Debug/net10.0-windows10.0.19041.0/win-x64/.

Arguments: --layout-test --layout-language=zh-Hans --layout-profile=long-country
--layout-close-run=893c39a377184099ae584b63c1ff9048. Original Process/SafeHandle
retained; CIM path/run argument/start time validated, no reopening by PID.
Same native-first/managed-second approximately2s observation procedure as prior
trace runs. New correlation/output file only bookkeeping. No matched-environment
or randomized-experiment claim; focus/timing and intervening state can differ.

## Route and outcome (UTC 2026-11-09)

- Launch18:08:36.7548486Z; target20112, observer23620;
  run893c39a377184099ae584b63c1ff9048.
- Unique returned trace window activated. Enter demo -> Nodes; long-country
  active title and long fixture labels verified via screenshot/accessibility.
- Clicked SearchBox, verified focus, sky.type_text jp-1 once. Immediate snapshot
  still showed empty input; no retyping. Subsequent read-only snapshot at
  18:10:03.667Z showed exact jp-1 and filtered long country/jp-1 labels, within120s.
- Sent exactly one sky.press_key Alt_L+F4 at18:10:14.851Z, while SearchBox focused.
  No titlebar close or second close. Recorded exit18:10:15.0356781Z, about185ms
  after request by separate UTC readings, not precise shared-monotonic latency.
- Final sample50 at18:10:15.9941576Z: native wait0/signaled, successful exit code0,
  nonzero exit FILETIME high31289461/low2445889325. Managed waittrue/HasExitedtrue,
  exit0/valid exit time agree. Observer ended normally at99.2827494s.
- All50 saved pairs contiguous and native calls ok/no snapshot error. Correlation,
  PID, managed start time and reconstructed native creation FILETIME exactly match.
  Date strings preserved with ConvertFrom-Json -DateKind String during verification.
- At18:10:35.405Z no demo window; subsequent independent CIM found0 demo/prototype/
  observer targets. No timeout, forced exit or replacement run.

## Stage receipt

Exact production reader linked in existing test assembly returned ValidPrefix,
CompletedClose=true,24 stages: close/dispose entry, all nine before/after removal
pairs, root-clear before/after, DisposeReturn, CloseReturn. No skipped conditional
subscription path. Health Dropped0/InvalidProducerfalse/Failedfalse. Marker span
3.7512ms is not measurement of instrumentation overhead. No app launch for parsing.

Ignored raw receipts beneath trace output directory:
- close-stage-records/893c39a377184099ae584b63c1ff9048.jsonl:
  SHA256 BE683A6DFDE589BCA8A2CA63A21DE33818B99F2169CAE6406AA20A91CE66E060.
- process-observations/893c39a377184099ae584b63c1ff9048/samples.jsonl:
  SHA256 19A8C1D00FF1381F11AC815673E1125E69ACFE19BD860B9EE14EC83BF8A5A955.
Raw screenshots/receipts and binaries not committed. This ledger restores results,
not original local artifacts, on a fresh computer.

## Combined interpretation / stop point

| Trace-artifact route | Complete healthy stages | Process result |
| --- | --- | --- |
| baseline + titlebar close | 24 | exit0 |
| long-country + titlebar close | 24 | exit0 |
| long-country + Alt+F4 | 24 | exit0 |

Long-country and Alt+F4, individually and together, did not force a stall in these
observed trace runs. This does not exclude an intermittent interaction, nor prove
the original code fixed, the recorder harmless, a service healthy under all states,
or any success rate. Original preserved binary differs from the instrumented build;
historical service/IME/focus/time and observation procedure were not controlled.
The historical native-unsignaled/wait-chain evidence remains valid and unresolved.

Stop automatic repetitions here. Next useful preparation is a read-only inventory
of remaining original-versus-trace artifact/build/dependency/observer differences,
then decide whether a specifically justified comparison merits a new budget.
Do not rebuild originals, alter shutdown semantics, remove subscriptions or disable
services based on three successes. No further runtime budget or broader acceptance.

References: [baseline](2026-11-09-windows-close-stage-first-observation.md),
[long-country/titlebar](2026-11-09-windows-close-stage-long-country-observation.md),
[historical comparison](2026-11-09-windows-close-stage-history-comparison.md).
