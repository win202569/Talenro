# In-context verifier timing — pending diagnostic proposal

Follow-up:[approved four-child observations](2026-10-17-synthetic-capture-context-profile.md).
The proposal below is historical;4/4 slots spent,all diagnostic children complete.
Instrumentation materially perturbs timing;no product repair/acceptance follows.

## Why change measurement strategy

Afterb4befe5f,user requested an in-context performance-location design. This
turn only inspected source and available command/module metadata;no new child,
trace,breakpoint,product edit or install was performed. The two isolated helper
comparisons did not establish PS5 benefit. Their opposite ordering on different
engines cannot isolate warmup/order effects;do not select either as a repair.

Local inspection found existing Stopwatch timing and PowerShell debugger/trace
commands. No ready dedicated cross-PS7/PS5 profiler was identified in the queried
module/command inventory;this is not proof none exists anywhere. The PS7 runtime
is the bundled executable,not the guessed Program Files location(which was absent).
wpr.exe was discoverable,but no system-wide trace is proposed or started.
No external profiling package,ETW session,privileged attach or installation.

## Proposed spike — question and method,NOT approved/executed

Question:within one actual public Test-SyntheticCaptureRecord call,how much
observed time/call volume falls in safe-type queries,passive property scanning,
exact shape checking and recursive comparison,relative to the complete call?
The existing manual phase reconstruction and standalone micro-tests cannot
answer this in the original public call context.

Propose NEW maximum4 fresh children in order:PS7 ordinary,PS7 instrumented,
PS5 ordinary,PS5 instrumented.120000ms per child,no retry,stop on any unexpected
result/timeout and retain partial output. These are separate slots,not reuse of
original frozen7..12 or any spent diagnostic allowance. Await user approval
before writing the diagnostic or creating source copies.

Each child uses the same trusted literal4096tick packet,fixed metadata and empty
arguments. Construct exactly once,then call the public verifier exactly once.
No codec/archive/store APIs or original-suite replay. Ordinary children load
unchanged pure contract/dependencies. Instrumented children load isolated
diagnostic source copies of that pure chain,not original definitions followed
by replacements. Preserve function names,parameters,call ordering,all original
statements,validation logic,return values and public catch/verdict behavior.
No product/workspace source is overwritten;copies are diagnostics,not fixes.

In copies only,add entry/finally timing and counts for these seven functions:
Get-OfflineCaptureValueType,Get-OfflineEvidenceValueType,
Assert-ContractPassiveProperties,Assert-OfflineEvidencePassiveValue,
Assert-ContractShape,Get-ContractPacket,Assert-ContractEquivalent. Profiling is
active only during the single public verifier call,not initial construction.
Use function-local timing state and a separate diagnostic nesting stack;do not
inspect/log caller values or change admission. Aggregate in memory with no
per-call console/file logging. Emit bounded totals/counts at completion,retain
whatever flushed markers exist if the parent terminates an incomplete child.

Report inclusive and measured-child-subtracted times separately. Recursive
inclusive totals overlap and must not be summed as percentages of wall time.
Subtracted time is relative only to instrumented children,not exact CPU/self
cost of every operation;unmeasured work and instrumentation bookkeeping remain.
Validate stack balancing/count consistency and reject inconsistent output rather
than guessing. This is source-copy instrumentation,NOT non-invasive sampling.
Entry/finally bookkeeping may distort cost or cause timeout;the cap stays120s.

Before launching,parse originals/copies and verify pinned source hashes plus
that removing ONLY designated timing additions reconstructs the original token/
statement sequence. An instrumentation-generation/check failure stops before
execution;no silent broader rewrite or alternative profiler fallback. Keep
public verifier body unchanged. On completion require Verified=true with the
same literal fixture facts and matching binding across ordinary/instrumented
children;all authority/action flags remain false. These positive fixture checks
do not prove universal semantic equivalence or hostile-input safety of copies.

Ordinary versus instrumented fresh hosts provide context for distortion,not a
precise overhead subtraction or stable performance ratio. No repeated samples,
warmup control or new benchmark guarantee. If instrumentation dominates or times
out,the result is incomplete/inconclusive;do not optimize from that artifact.

## Output and boundaries

Output is an explicitly exploratory report identifying observed cost distribution
and whether further operation-level investigation is warranted,not a fix or
permission to optimize. Original source/tests remain unchanged. Preserve portable
probe definition,copy-generation recipe/pinned hashes and report on development
branch;ignored generated copies/raw receipts stay local. Cloning is not launch
authority. No diagnostic code exists yet and no new allowance has been consumed.

Original archive6/12 spent,7..12 frozen,dependent regression/review unstarted;
coarse2/2 spent with PS5 timeout;internal2/2,lookup2/2,names2/2 spent. All prior
raw records remain. No RED/GREEN,timeout resolution,acceptance,main merge,real
capture/security/vendor clearance,Stage B/I2/I3/full C12 or extra permission.
