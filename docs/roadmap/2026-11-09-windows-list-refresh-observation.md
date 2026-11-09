# List refresh comparison: interrupted before automatic input

Baseline c1fde9c0971c1b4639db116af4b686c398f475d4. User approved one static
and one rebuild launch, automatic jp-1 in each, input confirmation within120s,
normal close observation at most360s, stop on anomaly, no kill/settings changes.
Before launch the operator stated the stricter360s total observation cap per run.

## Preflight / first launch

Restricted CIM query failed; a separately approved read-only check outside that
restriction succeeded, confirmed ordinary non-administrator token, zero existing
demo/baseline/variant processes, and variant EXE/DLL hashes matching the build
ledger. Window inventory had no diagnostic target. No system changes were made.
The launch host repeated token/process/hash guards and verified all three observer
module/source Git blobs against HEAD before importing the unchanged adapters.

Static launch06b605c19d444ed19fab8973e81d8183,PID13124,observer19144,
start2026-11-09T16:21:35.3813083Z,explicit --mode=static. CIM executable path,
command line and creation time checked against original Process and retained
handle. Observer PowerShell7.6.9. Native/managed pairs sampled approximately2s.
At16:22:02.128Z the accessibility tree showed empty ProbeInput and all three rows.

## Interrupted UI route

Clicking ProbeInput returned `user input was detected in this window; call
get_window_state before continuing`. No type_text call was issued. A fresh read
at16:22:25.819Z still showed no TextBox value. The agent stopped the comparison,
notified the user, and did not launch rebuild mode.

Normal-close click attempted at16:22:42.327Z returned `coordinate input geometry
is unavailable`; this is NOT evidence that a close reached the app. A fresh
screenshot captured unrelated foreground content despite target accessibility;
no coordinates from that image were used. Explicit target activation followed
by a fresh screenshot showed the diagnostic window with placeholder and rows.
Close attempt16:23:22.599Z again returned the user-input guard. No confirmed close
delivery; do not interpret continued liveness as a shutdown stall. User was asked
to manually click the window X, not terminate it in Task Manager.

These are tool-reported conditions; they do not establish what physical input
the user performed. Rebuild remains unstarted. Budget1/2 used, second withheld
under stop-on-anomaly; no replacement static run authorized. No source/build,
clipboard read, service/security change, debugger/dump or forced termination.

## Outcome

Observer completed with exit0,179 pairs, final native timeout. Last sample at
16:27:34.6751096Z (elapsed359.268s) had identity valid, native wait258/code259,
raw exit FILETIME0; managed WaitSignaled=false/HasExited=false and unavailable
exit code/time. End elapsed361.2595s: the2s polling sleep overshot the stated360s
total cap by1.26s. Record this timing deviation; do not call it exact-budget
compliance. No further native samples were taken after the final359.268s sample.
Some early displayed output was truncated; do not claim independently reviewed
complete raw receipts for all179 pairs. The loop stops on identity/wait errors,
and the final pair and End were retained and reviewed.

Final window inventory16:27:55.467Z still had the single variant window2231144.
Independent CIM16:28:00.6784044Z found PID13124 with original creation identity,
path and --mode=static; no baseline/demo or observer19144 remained. The target was
left running, not killed. No manual-close confirmation arrived during observation.
User must close its window normally; no additional UI inputs or launches planned.

Automatic-input comparison NOT_ACCEPTED; no root-cause/fix or acceptance claim.
Existing project gates remain. Do not automatically resume or replenish launches.
