# Adjacent automatic/manual input comparison

User approved two adjacent launches after cf0f605d: automatic jp-1 first, physical
manual jp-1 second. Each uses unchanged artifacts/recorders and the retained-handle
observer from the previous manual run. Startup30s, route120s, close criterion30s,
absolute360s per launch. No retry, kill, debugger/dump, clipboard inspection or
service/settings changes. Stop before second launch if any demo process/window
remains after the first observation. No remaining allowance beyond these two.
Baseline cf0f605dddd738ef336ba099f0b126b8751b1d48 clean; initial demo CIM/window0.
Timing, IME and service-state differences remain potential confounders; this is
not a randomized causal experiment and cannot clear broader acceptance gates.

Budget0/2 before first launch. Preflight token/hashes/module blobs are rechecked
inside each fresh observation host. No source implementation or build is planned.

## First run — automatic input (complete, shutdown unresolved)

Budget1/2 consumed, launch c4e0c74a702d44539a0d9e27fdebbc4b, PID8516,
observer23284, start2026-11-09T06:06:29.1238800Z. First navigation click reported
coordinate geometry unavailable; fresh screenshot confirmed login still present,
then one reobserved click succeeded. No launch retry. Nodes/SearchBox clicked,
caret confirmed, type_text jp-1 once. Stable exact value/group/node confirmed.
Close requested once06:07:47.746Z. Post-close native wait258/codeAPI success0/raw
exit time0; managed waitfalse/HasExitedtrue/code0/1601. No signal within30s.
User asked for optional exact PID8516 wait-chain screenshot, no end-process/dump.

User supplied screenshot confirms PID8516/thread17368 -> svchost PID9888/thread1812.
No raw screenshot is committed. All73 pairs valid:16 pre-close,57 post-close with
the same native258/code0/raw-time0 and managedfalse/true/0/1601 disagreement.
TIMEOUT06:12:29.1347743Z,360.0122052s,Killed=false; no signal or abort. At
06:12:45.4885093Z CIM found no demo or observer23284, window query also empty.
Thus second-launch precondition passed without killing/restarting anything.
PID9888 service recheck returned cbdhsvc_8b240 Running, creation before/after
enumeration2026-11-08T17:11:36.0749030Z, matching earlier host identity. Later
service membership is not an exact historical snapshot; thread lifetime unverified.

## Second run — manual input (complete, scoped PASS)

Budget2/2 consumed. Launch4f42c5bb001e4b7f9dbd8bde5dfb05c4,PID23160,
observer22500,start2026-11-09T06:12:57.9423263Z. Fresh host repeated token/artifact/
recorder checks. Startup observed within30s. Enter demo -> Nodes -> SearchBox click;
user explicitly confirmed physical jp-1 input/no paste request with “已输入”.
At06:14:12.497Z accessibility showed exact jp-1 and only JP group/jp-1 row, within
120s. Screenshot captured foreground chat rather than app, so final route evidence
is accessibility plus user confirmation, not matching screenshot pixels.
No automated type_text/set_value/paste in this run; no node/apply/connect action.

One close06:14:40.806Z, signal06:14:41.0557220Z (103.1147281s since launch,
approximately0.250s after close by separate UTC clocks, not precise monotonic latency).
All22 pairs valid/contiguous; first21 native258/code259, final native0/signaled,
successful code0, exit FILETIME high31289361/low2602630162. Managed final waittrue,
HasExitedtrue,code0,exit2026-11-09T06:14:41.0368018Z. No timeout/abort/kill.
At06:14:57.1095996Z CIM found no demo/observer22500; demo window query empty.

## Interpretation and handoff

Automatic search PASS / shutdown FAIL-UNRESOLVED; manual search and shutdown scoped
PASS. Adjacent comparison strengthens the input-path association, including repeated
wait on the same service host/thread identifiers, but does not identify an internal
tool transport, service defect or specific blocking API. Timing/IME/focus interactions,
fixed order and state carried between runs remain confounders. Two manual successes
and recurring automatic timeouts are not a general fix or statistical reliability test.
No application/recorder source, settings or services changed; no build/test rerun,
debugger/dump or clipboard-content access. All broader NOT_ACCEPTED gates retained.
Both launches consumed; no automatic further runs. Next consolidate a minimal
input-path reproduction/support report from existing evidence before proposing a
new discriminator. Never disable clipboard services or force app exit as a fix.
