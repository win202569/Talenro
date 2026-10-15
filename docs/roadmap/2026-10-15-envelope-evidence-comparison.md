# Envelope old/new evidence comparison — 2026-10-15

## Scope and conclusion

Approved next step: readonly comparison of the retained historical failure and
the successful004 diagnostic observation. No Go command, binary, test, build,
fuzz, new recording or old-PID query ran. Only saved artifacts, tracked source and
current fixed toolchain files were read. No product/runner/policy fix is made.

Confirmed: the701 product Path/SHA256 pairs match across the historical matrix3
failure and004 after normalizing Windows path separators, with701 unique paths
in each and zero case-sensitive differences. The original callback is unchanged.
004 directly observed the successful deadline-suppression branch. The historical
failure did not capture the corresponding operand identities, so its exact cause
cannot be retrospectively proved from004 or the controlled fixture.

## Three distinct evidence layers

| Evidence | Where it stopped/completed | What is established |
| --- | --- | --- |
| Historical matrix3-25,2026-10-13 | Original Envelope fuzz exited1 with context deadline exceeded | Actual product fuzz failure;12 executions, no operand-level shutdown trace |
| Preparation001,2026-10-15 | Origin query root exited0 but unknown member QueryError87 failed completion | No build/fuzz ran; forced owned cleanup verified, not natural success |
| Observation004,2026-10-15 | Precompiled instrumented original Envelope completed exit0 |73921 executions, valid ordered trace, natural complete shutdown/final0/empty |

Preparation001's87 is not a recorded cause of the matrix3 fuzz failure: it is a
different invocation and phase. Its saved row predates per-API acquisition fields,
so cannot distinguish OpenProcess from job-membership failure. No image/PID cause
is inferred. The later PowerShell-launch WdFilter trace was a different fixture,
not a trace of historical Envelope; it does not establish an antimalware cause
for matrix3. These failures remain separate, without claiming they can never be
related by an unobserved environmental factor.

## Comparison of the actual original-Envelope runs

| Dimension | Historical matrix3-25 | Successful004 |
| --- | --- | --- |
| Product source |701 files; callback pin318CF8DE... | Same701 paths/hashes and callback |
| Entry | go test package through fixed installed go.exe | Direct already compiled003 test binary; no build in004 |
| Core runtime source | Uninstrumented fixed Go1.26.5 | Copied Go1.26.5 with context/fuzz diagnostic hooks and observe mode |
| Fuzz contract | Envelope,5s,parallel1,test timeout30s; build-p1 | Same logical test settings; build-p1 belongs to003, not004 execution |
| Initial corpus reported |10 completed |5 completed |
| Reported executions |12, no new interesting entries |73921,6 new interesting entries,total11 |
| Elapsed measures | Fuzz6.02s,package8.844s,outer command12.1635228s | Progress6s,native binary interval7.1016999s |
| Detailed error trace | Not present |3 identity-aware records; final nil/workers0/watchdogfalse |
| Process completion evidence | Old command receipt has program,args,environment,times,exit | New owned-job receipt includes natural completion and final0/empty |

Elapsed values are different measurement scopes, not an apples-to-apples speed
comparison. Initial corpus/cache/input sequence, compilation/caching, instrumented
runtime, scheduling and launch/collection mechanisms differ. Neither throughput
ratio nor one exit0 isolates one cause. Old GOCACHE is the prior matrix scope;
004 uses its fresh fuzz-cache and003's bound binary. Historical input sequence
and scheduling are not reconstructed or copied. Recorded environment overrides
do not establish all inherited process environment values; old unrecorded GOROOT/
GODEBUG/GOMAXPROCS values must not be invented.

Readonly current file hashes confirm both go.exe files have the same pinned hash
D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F. This does
NOT mean the compiled runtimes are identical: their GOROOT source differs.
The installed context/fuzz files still match historical pins765F99F7.../D9DE1256...;
the diagnostic versions areC34B2779.../2E94C23F.... coordinator.patch adds hooks,
captures the existing child/conditional-parent reads and emits identity records;
it does not implement a new deadline-suppression rule. Hooks/serialization still
perturb scheduling, so observe is not an uninstrumented control.

## Exact branch supported by004

The saved parent-done record says incoming=deadline, childErr=deadline,
oldError=nil, newError=nil, parentRead=false/parentErr=unread, watchdog=false.
DiagnosticErrorIdentity labels deadline only for exact context.DeadlineExceeded
identity, not equal error text. In the inspected stop closure, equality with
the observed child error clears the incoming error; no fuzzErr is stored. The
worker record retains nil, and the final return has nil/workers0.

The separate existing controlled parent-first fixture records incoming=deadline,
childErr=nil, oldError=nil,newError=deadline, followed by return deadline/workers0.
That deliberate ordering reproduces the candidate classification mechanism in a
minimal actual-framework fixture, not the old original-Envelope event. Both
records have parentRead=false: the conditional extra parent Err read was skipped,
not observed nil. Gate/propagation flags are sampled diagnostic metadata, not a
global atomic snapshot. Do not rewrite unread values into inferred observations.

Together with the earlier true-context propagation-window probe, this supports
parent-before-child deadline classification as a concrete candidate. The missing
historical child Err/stop/worker state prevents exact attribution. It also does
not exclude a different worker/RPC/budget path or prove absence of parser defects.
The callback itself parses/compares material and does not create a context/timer;
that bounded source observation is not an exhaustive product correctness proof.

## Verification anchors

Historical raw scope relative to worktree:
`.superpowers/sdd/2026-10-13-declared-issuer-mapping-consistency/`.

| Retained evidence | SHA256 |
| --- | --- |
| matrix3-25-extra.json |58363EB6DB64A041BEEF893BC039B5AD1730A42BC4AADF271A03E66CD8DEF11E |
| matrix3-25-extra.stdout |1F7D0C659C2C2C83D171D8C2AAE3D7C9DB8B45FC33E5B8CE94C5B43E4598854E |
| matrix3-25-extra.stderr (empty) |E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| matrix3-before/stopped-source.json |9DB4FC99B644493C45EB7C0F30D59840DBA97F445EDA0C08BEAAE2A090D41B58 |
|004 before/after-product.json |5D711EE2E412812A09B4681E6287413B4F27DE842ED53C38496AABD71C18D2E2 |
| framework-001/arms/parent-first/execute/stderr.txt (diagnostic scope) |116790D2B9374277BD841E9272E7DBF2FEB1F95DC81B3DB99C1EBC67A9F99312 |

Different manifest JSON byte hashes are not source drift: old paths use backslash,
new paths forward slash; semantic comparison normalizes only those separators and
compares Path/SHA256 case-sensitively. All701 pairs match, including callback hash
318CF8DE04BA5D95E71E625ABF6AAED8DFD54B854A9A925B7A3004D4132D989E.
Detailed004 hashes/result remain in the [single-observation record](2026-10-15-envelope-single-observation.md).
No saved receipt, cache or source file was modified in this comparison.

## Next decision and recovery

Subsequent separately approved work completed one uninstrumented005 short
validation successfully; that does not rewrite this earlier readonly scope or
the historical failure. See [new scope, evidence and limits](2026-10-15-envelope-uninstrumented-validation.md).

Do not change product logic, suppress deadline errors or lengthen limits based
on this comparison. The next proposed validation is a separately designed,
explicitly authorized single uninstrumented fixed-toolchain short check with
fresh isolated evidence/cache and the original budgets, stopping on failure.
It would be new evidence, not replaying004, rewriting matrix3 or retry-until-pass.
Implementation/execution scope and how its result affects acceptance must be
approved before it runs. No such run is authorized by this readonly comparison.

Matrix3 remains24 pass/Envelope fail/Pair unrun as historical evidence.004 remains
one passing diagnostic sample. C11/C12/Task10/B01/production/main gates stay open.
Necessary comparison/status reports are saved on the existing development branch;
raw evidence/caches/binaries remain ignored. A new computer can recover this
reasoning from Git, not the old input sequence, private cache or machine timing.
Read current-status.md and docs/runbooks/repository-recovery.md before resuming.
