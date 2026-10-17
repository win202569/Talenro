# One-shot evaluation profile — validation dominates the diagnostic copy

After30b5cabd the user approved one fresh ordinary-token PS7 child, one4096-row
evaluation in an independent instrumented module copy, configured30000ms limit.
No original product import/evaluation, original suite/preflight, warmup, retry,
fallback, optimization or reviewer. The original four remaining runs stay frozen.

## Saved result

PS7.6.5 Core x64/admin=False; exit0,19296ms including launch/wait/drain,
TimedOut=false,owned-host-exited,stderr empty,CaptureValid=true,Complete.
All46 ordered snapshots and final PASS/summary/LIMIT present; no partial tail.
One copied API call,4096 validation rows and4096 decision fall-through rows.
Evaluation stopwatch18235.314ms; full original result equality assertion passed.

| Category | Completed helper / entered loop calls | Inclusive ms | Own ms |
| --- | ---: | ---: | ---: |
| Type | 73746 | 4232.880 | 4232.880 |
| ExtendedMembers | 40969 | 2966.034 | 2966.034 |
| Shape | 4097 | 8636.434 | 4847.386 |
| ValidationLoop | 1 | 17480.582 | 5467.848 |
| DecisionLoop | 1 | 670.873 | 670.873 |

Inclusive categories overlap: do not add them. Shape own subtracts completed
Type/ExtendedMembers child elapsed; validation own subtracts Type/ExtendedMembers/
Shape-own deltas inside that loop. Helper totals also include root setup outside
the loop. Own is residual elapsed, including uninstrumented helpers, dispatch and
instrumentation overhead, not pure CPU/self cost. Snapshots/logging also cost time.
The validation loop is about96% of this evaluation stopwatch; decision work is
much smaller. No single helper is the unique bottleneck: repeated shape, type,
extended-member and residual validation work all contribute.

This copy completed faster than the previous uninstrumented-product boundary
measurement21.253s. Separate single observations are not controlled before/after
benchmarks; instrumentation/JIT/host variability prevent claiming a speedup,
original-product timing equivalence or unique cause for the old full-suite R3.
No product optimization was implemented or accepted.

## Fidelity and safety

Five instrumentation edits reverse exactly to the unchanged current product
source. Saved copy text was checked before launch; static parsing passed for
probe/controller. Five helper functions match original35 tests exactly; the
Profile-Assert statement is the original P1-Check equality assertion. Input is
the same sequence0..4095/time/AtMs0 path with full literal prepared expectation,
LastAtMs0/LastSequence4095. Decision row counts describe this fall-through path,
not generic coverage for continue/error paths.

Frozen114 pins matched before/after (108 inherited plus prior packet/tool
anchors). Current productCF4CB4B5,ownershipC4905925 and original35 source unchanged.
The sole index slot was consumed before child launch; hidden no-shell/no-profile
ordinary-token process, separate UTF8 streams and owned-handle cleanup retained.
Timeout/error/partial-tail paths were not exercised or claimed validated.
The controller's console summary emitted a depth2 formatting warning and stringified
nested categories; the durable index is depth8 and raw snapshots retain structured
numbers. Raw/index were read directly; no replay or alteration. A post-run helper
comparison initially targeted ownership tests (which do not define those helpers);
correcting the read-only target to original35 confirmed exact equality. No host rerun.

## Recovery and limits

Staged whitespace checking reports two inherited space-only lines in the frozen
instrumented copy (170/253). They are retained for exact source reversibility;
the probe was not reformatted after measurement. This is not a test failure or
permission to relax product safety gates. Git warned about future LF/CRLF checkout
conversion; recorded SHA256 values identify measured local bytes.

Portable probe/report/status/recovery are checkpointed on the existing development
branch, not main. Ignored packet
`.superpowers/sdd/2026-10-17-p1-evaluation-profile/`
keeps scope/controller/frozen manifest/index/raw streams locally; no prior packet
was deleted. Cloning restores portable tooling, not ignored records/runtime or
authority to rerun. Hashes show retained consistency, not independent attestation.

| Artifact | SHA256 |
| --- | --- |
| Portable probe | `80324E51BA6161246BAEDCDB840B409B1E9C3E7CBA9E8663820AEFDA8E4A52A8` |
| frozen.json | `E3E79EBBE8E83C1F0F968D095E5CDC5D0B664D44F1F6DB95E74061AB77AE21D4` |
| run-once.ps1 | `3DB2403C7511597CB8C418659790B24ED7A5A44D817D68DEC7FBC0E6F510CD82` |
| invocations.json | `F1E547FF7CCF90669386955FADAAD2F5C828B53BCEFABFD95348FD2B13863063` |
| D1.stdout | `6CC91DB4CFB1E8112C1EEAAB6785710466D528D7204A9F536E41A5B95789E06E` |
| D1.stderr (empty) | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |

New diagnostic1/1 spent. Old R3=false,extended-view3/7/four frozen,review0/1
ineligible; all earlier budgets unchanged. I1/P1 unfinished, candidate unaccepted.
No main/full regression/project/C12/P2/P3/live/runtime/release acceptance.
Next propose a bounded semantic-preserving validation optimization design and
separate explicit verification scope, retaining all scalar/member/descriptor/
shape defenses, original assertions and30s gate. Do not cache away per-value
security checks, raise limits, or automatically resume R4.
