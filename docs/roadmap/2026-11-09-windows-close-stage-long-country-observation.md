# Approved long-country close-stage observation

User approved one new launch afterff0c9bcd: change only the intended profile from
baseline to long-country, retain automatic jp-1 and titlebar close, cap360s, no
retry/kill/system changes. Budget1/1 consumed; no further launches authorized.
This is sequential evidence, not a randomized or environment-identical comparison.

## Identity and unchanged inputs

Clean development worktree atff0c9bcd. Ordinary token; no demo/prototype target
process or window before launch. Observer module/source Git blobs unchanged.
Same close-stage-recording executable/directory as the first observation:
EXE SHA256 EC9E4375948F2E7AA36B34DF32F36A5374F68F61C6A1A4FE42A8108CD0255774;
DLL 7E14F4A771DB0F7024C95E95A4193A16448D7B444C8D11273717E8B15F46611B.
Hashes rechecked immediately before launch. No rebuild or source modification.

Arguments: --layout-test --layout-language=zh-Hans --layout-profile=long-country
--layout-close-run=7a48e35683ee455880421a6e83268478. New correlation/output is
bookkeeping, not another experimental condition. Exact path/run argument/creation
time checked through CIM after launch, with retained original Process/SafeHandle.
No PID reopening. Native then managed samples approximately every2s, as baseline.

## Observed route (UTC 2026-11-09)

- Run7a48e35683ee455880421a6e83268478; target PID22688, observer PID20844;
  launch17:59:09.9669346Z.
- Selected unique returned trace window; Enter demo -> Nodes. Title reported
  zh-Hans/long-country active. Screenshot/accessibility showed the long country
  and jp-1 display labels, establishing actual profile activation.
- Clicked SearchBox, confirmed focus, used sky.type_text once for jp-1. At
  18:00:24.893Z actual field value and filtered long JP/jp-1 labels confirmed.
  No node selection, apply, connection or dialog action.
- One titlebar close requested18:00:33.533Z. Recorded exit18:00:33.7237124Z;
  approximate UTC difference191ms, not shared-monotonic precise close latency.
- Final sample43 at18:00:35.1343579Z: native wait0/signaled, exit API success/code0,
  nonzero exit FILETIME high31289460/low927736964. Managed waittrue/HasExitedtrue,
  code0 and valid exit time agree. Observer ended normally after85.208s.
- Saved43 contiguous samples all have exact correlation/PID/managed start time
  and native creation FILETIME equality. Native API statuses all ok, no snapshot
  error. Pre-exit managed exit fields remain unavailable, not fabricated values.
- Window inventory18:00:55.307Z found no demo window; subsequent independent CIM
  found0 demo/prototype/observer targets. No retry/timeout/abort/forced termination.

## Stage validation and retained evidence

Exact production reader linked into existing test assembly read the receipt:
ValidPrefix, CompletedClose=true,24 records. Includes all nine unsubscribe
before/after pairs, root-clear pair, DisposeReturn and CloseReturn; no skip paths.
Health Dropped0/InvalidProducerfalse/Failedfalse. Marker span3.6546ms, not an
overhead estimate. No WinUI launch performed for this read-only parsing step.

Ignored raw files relative to the trace output directory:
- close-stage-records/7a48e35683ee455880421a6e83268478.jsonl;
  SHA256 4BB897A28D3823E90045F8144D2B6708A41AEF1BE2310006A74810E62D8A018B.
- process-observations/7a48e35683ee455880421a6e83268478/samples.jsonl;
  SHA256 23B7E35D813C03E988343CB1FAC40EDBEB8EA1EC5199F78CE1AC6639D569DB84.
JSON verification preserved dates using -DateKind String. Raw receipts/screenshots
and binaries not committed; portable result and recovery pointers are committed.

## Interpretation and next boundary

This long-country/automatic-input/titlebar-close route completed the handler and
process exit. Both baseline and long-country now have one successful recorded
route on the same trace artifact. Long-country alone is not sufficient to cause
the stall on every run; no causal exclusion, reliability rate or original fix.
Intermittence, changed instrumentation/artifact, keyboard close versus titlebar
close and system/focus/IME timing remain unresolved relative to historical stalls.

Next candidate, only after separate approval: retain this same long-country trace
artifact/input/observer, changing only close gesture to one Alt+F4, with the same
one-attempt360s cap and no retry/kill/settings changes. Success would not clear
historical failure; full CloseReturn with continued native liveness would narrow
that run to after/outside this handler. Missing/partial records remain inconclusive.
Do not modify shutdown behavior or disable services based on these successes.

References: [first baseline trace](2026-11-09-windows-close-stage-first-observation.md),
[historical comparison](2026-11-09-windows-close-stage-history-comparison.md).
