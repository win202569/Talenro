# Original complete-entry verification — PS7 PASS; PS5 timeout

After48db896f, user approved NEW independent two-host budget: ordinary-token
PS7 then PS5, configured30000ms each, stop on any failure/timeout. Both2/2 spent.
No retries, warmups, preflight tests, fallback, diagnostic copies/markers,
product/original-test edits, reviewer, old-budget use or timeout increase.

## Results from unchanged original entry

Entry: offline-supervision-policy.shape-lookup-regression.ps1.
Order: shape6, reflection3, fixture8, original35, ownership10 in each fresh host.
Expected83 ordered lines/64 PASS markers (62 test groups plus2 wrapper markers).

| Host | Wall ms | Cumulative CPU ms | Exit | Timeout | Lines / PASS | Result |
| --- | ---: | ---: | ---: | --- | --- | --- |
| PS7 7.6.5 Core x64 | 10528 | 13968.750 | 0 | No | 83 / 64 | Full regression PASS this run |
| PS5 5.1.26100.9711 Desktop x64 | 30079 | 30281.250 | -1 | Yes | 65 / 48 | Incomplete, NOT PASS |

PS7 completed all62 groups and both original wrapper markers with exact ordered
output and ExpectedResult=true. This is fresh original complete-regression
evidence, unlike the previous instrumented diagnostic. It does not prove stable
latency or repair of the earlier timeouts; their failed verdicts remain unchanged.

PS5 completed48 groups:shape6 + reflection3 + fixture8 + original first31 through
P1PassiveDescriptors. No completed P1ArrayAndIntegerBounds, CallerMutation,
ConservativeFlags or Complete. Original35 LIMIT, ownership10 regression and
final wrapper markers absent. Last completed PASS is not the exact instruction
active at termination; no internal markers were added. Do not attribute30s to
the4096 API alone or claim a unique cause.

PS5 configured cap30000ms, actual30079ms includes79ms termination/collection
overhead. Both stderr streams empty, parent/child admin=False, owned-host-exited.
CaptureValid=true both means exact full/prefix evidence, NOT PS5 PASS.
PS5 ExpectedResult=false. Controller's final STOP exception intentionally reports
the failed allowance, not a saved product assertion failure. No assertion error
was saved before kill. FormalProjectAcceptance=false on both.

CPU totals include all child threads, may exceed wall, are not CPU percentage or
per-group cost. These differing runtimes and runtime/host conditions are not a
matched causal experiment. No unique version/load/JIT/GC/validation cause is known.

## Integrity

Frozen201 pins: inherited196 plus five last diagnostic anchors. Engines checked
before/after. Slot recorded before launch, ordinary token enforced, separate
UTF-8 streams and bounded waits, owned handle only for termination.
After both runs,201 pins/controller/engines/two receipts/four streams rechecked.
Counts and every ordered output line independently matched frozen inventory
(full PS7, prefix PS5). No prior controller/index/raw/verdict rewrite.

Ignored local packet:
.superpowers/sdd/2026-10-17-p1-original-full-verification/
retains scope/controller/frozen inventory/index/raw, NOT committed.
Hashes identify retained local bytes, not independent attestation.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | AB7C72214BEE7136CAC0223163147B9E136E97FFC810E813BEA4C330DDFE30B7 |
| run.ps1 | AF2A0159C9B8AFFA3C5B1E21F875B95CA6A3585486874D74B38D809AC123A657 |
| invocations.json | 7934BD411A44BCB9356E5123A189951C685A54EF77818F49369FAB201B162081 |
| PS7 stdout | 3BCD1E5C506D248A2417652F3E75F3E82126A843A923B6EA4F1C5AF25660BAC4 |
| PS5 stdout | CC6DA49921E2CA5A07B5A2DC1752EBAC3C058950834FD85F5990B23B71E44DB6 |
| Both empty stderr | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Unchanged candidate:
953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.
Unchanged full entry:
12D6F5460F97B9E09453B0387C30E3D20CE19B2EDF1CDF4DBF4E77DA31561422.
PS7 executable:
362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.
PS5 executable:
60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F.

## Handoff

New original-full2/2 spent. Previous CPU diagnostic1/1 and targeted2/2 spent;
shape2/3/one frozen; prior independent full1/2/one frozen;
reflection3/4/one frozen; extended-view3/7/four frozen; other budgets unchanged.
Review0/1 unused/ineligible. Candidate UNACCEPTED, I1/P1 unresolved.
No main/C12/project/live/runtime/release acceptance or automatic reviewer start.

Save this partial-success report/status/recovery to the existing development
branch. Original entry/source/tests already committed. Clone restores portable
artifacts, NOT ignored raw/cached engines/credentials or authority to replay.

Next design bounded PS5 timing attribution around the not-completed original35
boundary in the SAME complete order, preserving assertions. This is not permission
to launch it. No unchanged full retry, speculative product optimization, raised
limit or automatic frozen-slot use. Targeted PS5 PASS cannot replace full PS5.
