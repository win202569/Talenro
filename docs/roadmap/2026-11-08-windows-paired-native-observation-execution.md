# Paired native observation execution

Plan approved after0fbb76e0. Preflight baseline clean, same development branch.
Ordinary token confirmed. Restricted CIM query failed; approved broader read-only
query succeeded with ordinary token and demo count0; UI inventory had no demo.
Ruling: initial restricted query is not absence evidence; resolved using broader
read-only check before launch, without changing account/security settings.
Expected EXE/DLL hashes match plan. Managed module blob92ee0f0c23fd1521fbe103a781ab1ae2c9a58a7c;
native module c1fccc8485a111eab1814fb21cf43807ac78cb67; helper7c7d64dc0aaafffed6a6efb3459bece8ab949bfc.
Budget0/1 before session; helper will be loaded in the observation host before launch.
Session command uses unchanged modules, original Process/SafeHandle, native-first
pairs about every5s, lossless creation FILETIME comparison, bounded outer waits
and360s cap. No stream redirection/reopen/kill. Raw JSON remains session-local.

## Live sequence (UTC)

Budget1/1 consumed. Host20004, PowerShell7.6.9/runtime10.0.14, OS10.0.26200.
Helper preloaded18:55:13.2160720Z; app PID21236/start18:55:13.2458312Z.
Original Process/SafeHandle retained; native creation FILETIME134386377132458312
exactly matched original StartTime. Initial pair1 at18:55:13.2601311Z valid:
native wait258(timeout), exit-code API success/output259, times success/exit halves0;
managed waitfalse/HasExitedfalse and expected exit-property errors/nulls.

Entered offline demo -> Nodes -> screenshot-backed explicit SearchBox click ->
visible caret -> typed jp-1 once -> stable value and only JP group/jp-1 row.
No select/apply/connect/dialog; route completed within2minutes. Normal close via
one Alt+F4 requested18:56:15.465Z. Close has UTC receipt only, not a synchronized
launcher monotonic close event; elapsed close is approximately62.22s from launch.
Do not claim exact monotonic close timing. No retries or further UI input.

Pair13 at18:56:13.8202313Z was still running-state259/false. Pair14 at
18:56:18.8381853Z (elapsed65.595s) retained identity and reported:
- Native Wait: ok, ApiSucceeded=true, RawResult258, timeout, LastError=null.
- Native ExitCode: ok, ApiSucceeded=true, output0, LastError=null.
- Native Times: ok, ApiSucceeded=true, exit High0/Low0, LastError=null.
- Managed: waitfalse, HasExitedtrue, code0, ExitTime1601-01-01 UTC.

Fresh UI inventory showed no demo window. At18:56:32.1258354Z broader read-only
CIM still listed app21236/one thread and launcher20004; name count1. Evidence
places the disagreement at the native-return boundary on this retained handle,
not solely in managed conversion/caching; it does not establish OS/app/environment
root cause. Continue bounded observation only; no forced termination.

## Final result

All73 pairs retained creation identity. Pairs1-13 were running-state259; pairs14-73
(60 post-close pairs) showed native258/exit0/raw exit time0 and managed false wait/
true HasExited/code0/1601. No OUTER_WAIT_SIGNAL, native call failure or ABORT.
Final pair73 spans19:01:13.2584854Z..13.2598885Z, elapsed360.0158833s.
TIMEOUT at19:01:13.2603903Z, elapsed360.0164045s, Killed=false. Observation shell
exit0 reports command completion, not successful application shutdown. Original
Process disposed only after the last read. At19:01:31.3053532Z independent CIM
app/launcher count0 and demo-name count0; no kill or new launch.

Input/filter PASS; shutdown FAIL/UNRESOLVED under30s criterion. Native live
running/conflicting-return paths now observed; native signaled-exit and API-failure
paths remain unobserved. Raw successful API output0 is not a fallback for failure:
API success was separately true and LastError null. Native zero time is confirmed,
but its interpretation as a real exit date is invalid. One thread in CIM and
later absence do not establish teardown mechanism or causality of observer exit.

Budget1/1 consumed. No rebuild, source/settings change, debugger, dump or automated
test run. Only helper compilation and approved live calls occurred. Raw receipts
and screenshots remain session-local; portable summary and exact identities saved.
Next: read-only analysis of teardown and observer lifetime, guided by native-boundary
evidence; no further identical launch. Stack/dump, changed launch context or other
runtime experiment requires a new discriminator and separately approved scope.
All Task11/12 and44/34/6/217 NOT_ACCEPTED remain.
