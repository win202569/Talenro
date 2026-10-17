# PS5 reverse-order diagnostic pair — 2026-10-18

Diagnostic completion only; candidate remains UNACCEPTED. Original PS5 entry,
I1/P1 and wider project/runtime/security gates are not cleared.

## Approved scope and execution

Following the [read-only comparison](2026-10-18-p1-ps5-instrumentation-readonly-comparison.md),
the user approved a new independent two-host allowance: unchanged region probe B
first, unchanged progress probe A second. Each ordinary, non-administrator PS5
process had a 30000ms cap, no warmup, retry, product edit or old-budget reuse.
Both used the same pinned Windows PowerShell executable, no profile and
noninteractive launch, separate saved output/error streams and owned-process
cleanup. No additional tests were launched for the subsequent read-only audit.

Before launch, static inspection found a controller guard that would prevent
the second slot. The unlaunched v1 controller/manifest were retained unchanged;
v2 corrected the guards and pinned those two draft artifacts. V1 has no invocation
index and consumed no host allowance. Only v2 launched B then A; its 2/2 allowance
is now spent. Neither measured probe was changed.

| Measurement (ms unless noted) | B: regions, first | A: progress, second |
|---|---:|---:|
| Started UTC, 2026-10-17 | 20:22:55.0023034Z | 20:23:16.9077627Z |
| Host elapsed | 13280 | 13425 |
| Host process CPU | 14828.125 | 14890.625 |
| Exit / timeout / stderr characters | 0 / false / 0 | 0 / false / 0 |
| ValidationCopy phase wall | 7471.303 | 7701.661 |
| ValidationCopy phase CPU | 7609.375 | 7968.750 |
| At 1536 rows: validation | 2635.680 | 2779.913 |
| At 1536 rows: copy | 42.696 | 47.950 |
| Phase start to 1536 snapshot: process CPU | 2890.625 | 3140.625 |
| At 4096 rows: validation | 7321.486 | 7552.246 |
| At 4096 rows: copy | 124.314 | 129.613 |
| First 31 original-group span | 1516.195 | 1336.003 |

Each output has 245 nonempty lines: 152 CTIME events, nine snapshots and 84
ordinary lines, including 64 PASS markers (62 groups plus two wrappers).
All 76 timer pairs closed; no partial tail or active scope remained. CaptureValid
and Complete are true; FormalAcceptance is false. Both owned hosts exited.
Snapshot row inventory is 1,512,1024,1536,2048,2560,3072,3584,4096.

## Independent read-only verification

Recomputed all 76 pair wall/CPU differences per output from raw ticks and
verified equality with receipts; checked event stack and monotonic event/sample
clocks, nine snapshot row counts, validation/copy tick conversions, ordinary/PASS
counts, output hashes, controller/engine hashes and all 262 frozen file pins.
For B, all nine raw disjoint region sums equal ValidationTicks exactly. At4096:
record1279.021ms, shape3692.671ms, field2349.795ms (display rounding only).
The shape region includes nested type/member/passive checks, not merely name
membership. A has no region counters; absence must not be treated as zero cost.
CPU is cumulative all-thread process time, not per-region CPU attribution.

## Interpretation and remaining gates

Both unchanged diagnostic copies completed this reverse-order pair with similar
wall times. This does not reproduce historical A's timeout/16490.836ms validation
at1536 rows. It weakens a deterministic claim that A's instrumentation alone
necessarily produces the historical slowdown, but does not isolate host/runtime
state, order effects or instrumentation overhead as a unique cause. One pair is
not a stable benchmark or proof of a product speedup. Parent/child inclusive
timings must not be added together; historical censored totals are not speedup
baselines. Historical results remain intact.

Original product SHA remains
`953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA`.
The uninstrumented original PS5 full-entry timeout remains an acceptance blocker;
diagnostic completion cannot replace that gate. All older spent/frozen allowances
and review ineligibility remain unchanged (review0/1 unused, not eligible).
No main merge or C12/project/live/runtime/release acceptance is implied.

Next recommended scope: separately approve one fresh ordinary PS5 run of the
unchanged original full entry with its existing 30000ms cap and exact output/
process safety checks. It would be a new explicitly bounded allowance, not a
retry from any frozen budget. Do not launch automatically, increase limits,
remove defenses or call this issue fixed; stop and preserve evidence if it fails.

## Provenance and recovery

Ignored local packet: `.superpowers/sdd/2026-10-18-p1-ps5-reverse-pair-v2/`.
Its controller/raw/index are retained locally, not portable Git recovery inputs.

- Frozen manifest: `D29351E7C5DBBA15DC8C1E0BEE2A4344CB9C25A37BA0CE809A823E651AB4909A`
- Controller: `16102464E80400C61C61F4F16177347FC52A3381213A58627A250180395BEA56`
- Final index: `F43B8E8B68E6EC1FA74315D9806E0D5E3D84CCCEA3BF49F828E6BBFBACB885AE`
- B stdout: `9BF2200189B216213F27B384097F8B155A2B9170EDC96CB4DB0C13F5E008BC07`
- A stdout: `8DB58AEB34D289288754F881CBA71F2AB20EDF3BF5D95A9840A6CEA4B19FD06D`
- Both empty stderr: `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855`
- PS5 engine: `60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F`
- B product copy: `D3D063440039A31FEF45B3C696CF3EB6D587E759775D37215FA93AF15E18AD6D`
- A product copy: `37935F0E66B7ADC44E4E591FECE4B040A50E4E2586A17DE2760E3A594663A07E`

Both probe entry hashes are
`04526277000A8D362458B9B7BFC6F84D4BE14CA2C0BF6F63011F6EE175DB6A88`;
the six shared scripts remain identical. Fresh-computer recovery uses this report,
the committed diagnostic folders, current status and recovery runbook on
`codex/c12-b01-task9-coordinator`, not ignored packets, private caches or replay
authority. Cloning restores neither raw records nor permission for new runs.
