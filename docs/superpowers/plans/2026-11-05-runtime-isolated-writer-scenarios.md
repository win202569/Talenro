# Isolated writer — remaining scenario continuation

User requested 继续 after verified compiler checkpoint4b935dae.
Parent design/plan remain authoritative; only numbering and script binding change.
Current32/11/6/217; S01..S09 become computations33..41, at most17new starts.
Ceiling41/28/6/217. No compilation, unit rerun, business host, installs or main merge.
Existing verified helper80896bytes/SHA256
2AEF2A39B5D499981378D8FD1F749A3073AD92C03641CEBBA28E9060B3FFCD7B
and old adapter are loaded only after identity checks. Missing DLL stops.

New scenario-only role/driver and directory:
.superpowers/sdd/2026-11-05-runtime-isolated-writer-scenarios-v1.
Do not edit consumed v1/v3 sources, scripts, claims, stops or compiled evidence.

|ID|Computation|Starts|max observation|
|---|---:|---:|---|
|S01|33|2|SUCCESS,8193stdout/48stderr pattern bytes,two exit0|
|S02|34|2|SUCCESS,0/0bytes with real EOF/close,two exit0|
|S03|35|1|ReadyTimeout,producer never starts,writer owned exit|
|S04|36|2|DrainTimeout,producer exit0,held-write writer killed|
|S05|37|2|DrainTimeout,producer exit0,held-finalization writer killed|
|S06|38|2|WriterExited or TransportFailure,writer exit23,both exits|
|S07|39|2|ProducerFailed,producer exit23,both exits|
|S08|40|2|CapacityExceeded,65537read,at most65536forwarded,both exits|
|S09|41|2|MissingReceipt,both exit0,outputs unavailable|

First unexpected result stops all later scenarios without retry or reuse of slots.
Unchanged120000execution/5000readiness/2000shared drain/2000shared exit budgets;
these bound programmed waits, not OS calls or persistence. Original owned process
references only; no PID adoption/tree kill, no disposal of pending task resources.
Failure outputs unavailable/null, not read/hash/salvaged. Only S01/S02 stable output
is compared with fixed byte(index%251) data/empty bytes after successful supervision.

1. Bind prior33-file compiled archive, source/absence pins, both binary hashes,
   script identities and this continuation. Parse four PowerShell files; record
   nine names/17starts/no compile site. Static preparation once; unexpected stop.
2. Archive/verify/push preparation before invocation. No extra RED build/test/reviewer.
3. Invoke scenario-driver once, sequential nine rows. Exclusive gate/per-case claims.
   Checkpoint every result before next row. Verify available writer maps against
   supervisor map; only terminated writer may lack after-map. Compare success bytes.
4. Independently inspect receipts, raw controls, maps, counts and all available
   results; preserve actual partial data, mark later cases NOT_RUN on stop.
5. Archive sources/evidence, exclude binaries and unavailable data, update
   status/recovery, commit/push explicit files and verify remote equality.
   End this finite gate. All five flags remain false, NOT_ACCEPTED.

No design change or C# edit. Scope exception remains finite native inline execution,
without separate broad tests, reviewer agents or cleanup.
