# Read-only comparison of PS5 progress/region timing — instrumentation/conditions unresolved

After2e565651,user requested continuation. ZERO product/test executions,new
diagnostic children,source/tool edits or budget spend. Only findings/status/
recovery documentation is updated. This is coordinator audit,not independent review.

## What is held constant and what is not

Six scripts (probe,existing wrapper,original35 diagnostic,ownership,reflection,
shape tests) are byte-identical between progress and region folders.
Same original candidate953CC2A8,same PS5 engine path and hash:
60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F.
Launch setup code matches:no-shell,hidden,no-profile,noninteractive,working
directory,UTF-8 separate-stream settings. Actual target folder/path differs by
variant;do not claim identical command-line file arguments.

Only copied product differs in test-executed source:
- snapshot helper accepts three additional region counters and emits schema2;
- three new counters initialized;
- two additional timestamps per row around the existing shape call;
- three region accumulations and expanded snapshot call.
Original validation/copy statements and guards are unchanged.
Removing these additions/schema changes restores previous copy exactly.
No performance optimization was implemented.

New controller differs in pin anchors/target and post-run snapshot parsing.
Hash/source checks occur outside the child stopwatch;output parsing follows
child exit. They are not inside the recorded per-row validation interval.
They can affect surrounding system/cache state;neither run resets the machine
or controls historical global load,runtime/JIT/GC,power/thermal or filesystem state.
Fresh processes alone do not establish a cold or equal environment.

## Matched completed-count comparison

Earlier A started2026-10-17T19:59:29.1561495Z;later B20:04:29.3923577Z.
A timed out30062ms with last1536 snapshot;B completed18993ms/4096 rows.
Compare common1536 count,NOT censored process totals as a speedup:

| Retained metric | Earlier progress A | Later regions B |
| --- | ---: | ---: |
| Validation cumulative wall ms at1536 | 16490.836 | 3058.176 |
| Copy cumulative wall ms at1536 | 258.662 | 64.736 |
| Phase elapsed wall ms at1536 | 16803.052 | 3143.601 |
| Process CPU delta:phase START to1536 snapshot ms | 17250.000 | 3453.125 |
| First31 original-group span before boundary ms | 5534.513 | 5399.445 |

CPU window includes ALL process threads and checkpoint/timer/runtime work,
not region-exclusive CPU. It is derived from saved phase-start and snapshot
cumulative samples. Both wall and CPU differ strongly;this is not merely
comparing30s censoring with a completed run. It still does not exclude scheduling
effects or identify instruction count,CPU frequency,JIT,GC or host contention.

Unchanged copying also differed,while earlier original-group span was close.
Thus the pattern is not evidence that the added shape timers optimized a specific
safety check. Same product semantics with different instrumentation and execution
conditions remain confounded. No single historical run can resolve that.

## Safe next question, not execution authority

Test whether the difference recurs with unchanged existing probes in a reverse-
order pair:B (region variant) then A (progress variant),same frozen PS5 settings/
candidate/full order and separate ordinary processes,without warmup or retries.
Record matched row-count wall/CPU outcomes,including any censored prefix.

One pair would still be exploratory,not a stable benchmark or unique-cause proof.
If B fails/times out,the second process should not launch under the usual
stop-on-failure scope. Any new pair requires its own explicitly approved
two-host allowance,each30000ms;this report neither creates nor spends it.
Do not invoke old consumed controllers or use frozen slots.
No new probe/controller code was created this turn.

## Integrity and recovery

All255 pins and A/B raw stream digests rechecked,engine bytes match.
Current index remains:
393C74BD4E96F8607E241A8F50FF9B9362516F1A76D5FC5C54F772CB09125ACF.
Original product:
953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.
The read-only no-index source diff returned1 because files differ as expected;
it is not a new test failure. No old record rewritten or product evaluated.

Regions1/1,progress1/1,PS5 phase1/1,subphase1/1,group1/1 spent;
original-full2/2 (PS7 PASS,PS5 timeout),earlier PS7 CPU1/1,targeted2/2 spent.
Shape2/3/one frozen;prior independent full1/2/one frozen;reflection3/4/one frozen;
extended-view3/7/four frozen;review0/1 unused/ineligible;other budgets unchanged.
Candidate UNACCEPTED,I1/P1 unresolved;no main/project/C12/live/runtime/release
acceptance. Save this audit/status/recovery on existing development branch.
Clone restores artifacts,not ignored raw/private runtimes/credentials or replay
permission. No guard deletion,timeout increase or automatic original retry.
