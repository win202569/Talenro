# Single-TextBox prototype automatic/manual observation

User approved two launches after8196fce8: automatic jp-1 then physical manual jp-1.
Only TextBoxShutdownProbe, no demo launch. No rebuild/source/service/settings change,
clipboard-content read, debugger/dump, retry or kill. Each: startup30s, input route120s,
close criterion30s, absolute360s. Do not start second with leftover target process/window.
Baseline8196fce8077d91bcd4cf2f4a9ec0fcb3a04df486 clean, initial CIM/window count0.
Budget0/2. Same retained-handle observer and unchanged modules; fresh ordinary-token
host checks module blobs and prototype EXE/DLL hashes from the build ledger each run.
Session command changed only target name/path/hashes and removed demo arguments;
no Nodes/search route, just ProbeInput exact value. UI close UTC is separate from
observer monotonic time. Results remain limited by timing/IME/focus/fixed-order state.

## Automatic run

Budget1/2, launch15ec38f939d04503a5b111c01d12b3bc,PID23796,observer14272,
start2026-11-09T06:27:05.9520073Z. Startup accessibility observed within30s;
initial screenshot was foreground chat, then explicit activation produced the
prototype screenshot. Click ProbeInput, confirm caret, type_text jp-1 once.
At06:27:49.910Z screenshot/accessibility confirmed actual text jp-1 (not placeholder).
One close06:27:56.615Z; signal06:27:56.7892597Z, about0.174s by UTC, elapsed50.891s.
All12 pairs valid/contiguous. Final native wait0/codeAPI success0/exit FILETIME
high31289363,low1970073415; managed waittrue/HasExitedtrue/code0,
exit2026-11-09T06:27:56.7745863Z. No timeout/abort/kill.
At06:28:09.5545763Z independent CIM found no prototype/demo/observer14272;
window inventory empty. Second-launch prerequisite passed.

## Manual run and timing deviation

Budget2/2, launch51dc7bd2e72349faa6d6d7eec5b705cf,PID20212,observer18376,
start2026-11-09T06:28:15.8979575Z. Startup within30s, empty ProbeInput clicked;
user asked to physically type jp-1 without paste, later confirmed “已输入”.
06:29:57.694Z screenshot/accessibility still showed placeholder, not text.
First exact jp-1 accessibility confirmation06:30:17.516Z was about121.618s from
process start, after route120s deadline. Actual key-entry timestamp unknown;
do not claim the user's typing itself was late or that route passed in time.
No additional text input was issued; normal close followed at06:30:37.927Z.
The polling/response gap meant close occurred about22s after route deadline;
record this protocol deviation, not a silently extended route allowance.

Signal06:30:38.1150036Z, elapsed142.2187179s, about0.188s after close by UTC.
All30 pairs valid/contiguous; final native wait0/codeAPI success0/exit FILETIME
high31289363,low3583353911; managed waittrue/HasExitedtrue/code0,
exit2026-11-09T06:30:38.1026359Z. No timeout/abort/kill. At06:30:56.4452865Z
CIM found no prototype/demo/observer18376 and window inventory was empty.

## Interpretation / next boundary

Automatic input/close scoped PASS. Manual close has valid successful evidence,
but full manual route NOT_ACCEPTED due to verification timing deviation. Separate
UTC close/signal clocks are approximate, not precise shared monotonic latency.
The stall was not observed in this prototype in either run. This does not prove
the original demo logic is defective, rule out a tool/framework/environment role,
or establish equivalence between different artifacts, startup routes and timings.
The prototype is not yet a reproducer of the failure. No WCT needed on these runs.

No build, source/recorder, service or system changes; no demo launch or clipboard
contents read. All broader acceptance gates remain;2/2 launches exhausted. Next
read-only compare original/prototype initialization, controls and resource/layout
paths to select one possible reduction variable before further code or execution.
Do not infer a repair or automatically add another run. Raw receipts/images not
committed; this ledger preserves the evidence and deviation for fresh recovery.
