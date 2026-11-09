# First approved close-stage observation

User approved one diagnostic launch, jp-1 input and normal window close, maximum
360 seconds, no retries/kill/settings changes. Budget consumed1/1. No second launch.
Source baseline e347d8cd; no source/build changes during this observation.

## Preparation and identity

Ordinary token confirmed. Initial restricted CIM read was unavailable; repeated
read-only preflight in ordinary external host succeeded before launch. No existing
demo or either prototype process. Three observer module/source blobs matched HEAD.
Trace EXE SHA256 EC9E4375948F2E7AA36B34DF32F36A5374F68F61C6A1A4FE42A8108CD0255774;
DLL 7E14F4A771DB0F7024C95E95A4193A16448D7B444C8D11273717E8B15F46611B.
Path: clients/windows/Talenro.Demo.App/bin/close-stage-recording/x64/Debug/
net10.0-windows10.0.19041.0/win-x64/. Both hashes checked again immediately before
launch. Arguments: --layout-test --layout-language=zh-Hans --layout-profile=baseline
--layout-close-run=5c5367fa64f0443cb0afea695703cd19.

Original Process object and retained SafeHandle used throughout; no PID reopening
for native observations. CIM exact path/run argument and creation time matched.
Run5c5367fa64f0443cb0afea695703cd19; target PID12356, observer PID17260.

## Route and results (UTC 2026-11-09)

- Launch17:49:23.1779168Z. Unique trace window selected from returned window list.
- Clicked offline EnterDemoButton, then NavNodes; did not connect VPN or use accounts.
- Clicked SearchBox, verified focus, typed jp-1 once using sky.type_text. Actual
  field value and filtered node accessibility confirmed17:50:47.138Z, not a placeholder.
- Clicked observed normal titlebar close button at17:50:58.253Z (request timestamp).
- Recorded process exit17:50:58.4871900Z, code0. Approximate UTC difference234ms,
  not a shared-monotonic precise measurement or an overhead comparison.
- Observer ended normally after97.374s, well within360s. 49 sequential pairs,
  indexes1..49, same run/PID/start time, all native calls ok/no snapshot error.
  Final native wait0/signaled, exit API successful/code0; managed WaitSignaled and
  HasExited true, ExitCode0. Native exit FILETIME high31289458/low3765306332.
  Managed exit-code/time getters before exit return recorded errors, as expected;
  those unavailable pre-exit fields are not treated as successful values.
- At17:51:19.076Z no trace window remained. Subsequent independent CIM found0
  demo/prototype/observer targets. No force-close, process termination or retries.

## Stage receipt

26 JSONL lines: ready,24 stages,healthy terminal record. Frequency10000000;
Dropped0, InvalidProducer=false, Failed=false. Exact production reader linked into
the existing net10.0 test assembly was invoked read-only on the retained file:
Status=ValidPrefix, CompletedClose=true, Records.Count=24. No WinUI assembly executed
for parsing. Stages cover close/dispose entry, all nine before/after unsubscribe
boundaries, root clear before/after, DisposeReturn and CloseReturn. Neither null
subscription skip path was used. Marker span3.889ms, not recorder overhead.

Raw files stay ignored under the trace output directory:
- close-stage-records/5c5367fa64f0443cb0afea695703cd19.jsonl;
  SHA256 F61D6C994D968E250FACACFD7B205424063BD2F59E90428A57CD974E035E71A5.
- process-observations/5c5367fa64f0443cb0afea695703cd19/samples.jsonl;
  SHA256 5FE1F1AD4A7C2E7EC416E293DD85D48C03C2E2CD19D8FCC5169387C104BBDECB.

Read-only verification initially rejected a timestamp comparison because default
ConvertFrom-Json converts ISO timestamps to DateTime. Re-reading the SAME files
with -DateKind String verified all49 identities/order. No original data altered,
no application rerun and no process guard relaxed. Future PowerShell parsing must
preserve date strings or compare explicitly normalized instants.

## Interpretation / next step

This one automatic-input/normal-close route completed both the instrumented
handler and process exit. The original stall did not reproduce; no root cause or
fix established. No inference that svchost/IME/observer can never stall, and no
general runtime acceptance. Tail loss remains possible on other runs.

Recommended next work is read-only comparison of this complete baseline against
the earlier stalled case (input method, app build, window/page and environment),
then select a single-variable experiment only if evidence justifies it. This
record grants no further launches, builds, system changes or main merge. Recovery
from Git restores this portable result, not raw local receipts or binaries.
