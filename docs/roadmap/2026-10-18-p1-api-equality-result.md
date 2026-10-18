# API/equality diagnostic — complete once; API-call region dominates measured split

After8d38ab9b,user approved a NEW ONE-host ordinary fresh PS7 diagnostic,
30000ms cap,no warmup/retry/PS5 fallback or old-slot reuse. Sole1/1 slot SPENT.
No target-source/product/assertion changes,reviewer or main merge.

## Observed result

PS7 7.6.5 Core x64;parent and child admin=False.
Started2026-10-18T09:33:47.8730125Z,wall28208ms,all-thread CPU36953.125ms.
Exit0,no timeout,empty stderr,owned-host-exited,streams saved.
95 ordered nonempty lines/64 PASS;final ShapeLookupRegression.
12 timing markers,six closed pairs,zero open scopes.
CaptureValid=true,ExpectedResult=true,Classification=complete-diagnostic,
FormalProjectAcceptance=false. No saved assertion/fixture/classifier error.

Stopwatch.Frequency=10000000 ticks/sec,recorded outside the target output.
Raw START/END differences,not guessed deadline durations:

| Closed region | START ticks | END ticks | Wall ms |
|---|---:|---:|---:|
| InvalidInputs |3811028120812|3811030358781|223.797|
| Build4096 |3811030359921|3811035998734|563.881|
| Evaluate4096 |3811036033191|3811211710761|17567.757|
| Assert4096 |3811211713500|3811211935529|22.203|
| Check4096 parent |3811036000389|3811211938104|17593.772|
| Check4097 |3811211938968|3811212068598|12.963|

H-API is supported for THIS completed diagnostic:Evaluate4096 is far larger
than Assert4096. The observed outer Check4096 cost is predominantly in the
API-call assignment region,not full result comparison/assertion. No inner
evaluator helper/statement attribution follows.

Evaluate includes command lookup/binding,API execution and return assignment.
Assert includes full P1-Equal/P1-Need and dispatch. These are elapsed wall regions,
not pure internal API/equality CPU. Parent contains the two children;DO NOT sum
all three as independent work. Exact parent-minus-children residual3.812ms
(after rounding) is unassigned arguments/dispatch/return/logging/gaps,etc.,
not a calibrated overhead estimate or proof of one component's cost.
Do not assign the rest of process wall or all-thread CPU to these regions.

## Fidelity, constraints and what remains unresolved

All prepared copies,original helpers,full assertions,input expressions and
product bytes remained fixed. Only the already-prepared dedicated4096 helper
and four inner markers differ from the outer-marker path. No local product
reload or per-row instrumentation. Distinct helper identity/body/history can
change runtime dispatch/compilation;equivalent source scope does not prove
identical cached engine behavior.

Earlier uninstrumented candidate PS7 timed out30090ms;outer-marker diagnostic
completed28437ms;this split diagnostic completed28208ms. These are NOT matched
performance arms. Do not claim stable improvement,original timeout repair or
cause from their differences. This run does not accept the unchanged original
entry or promote the isolated integer candidate. Historical PS5/candidate
timeouts and original verdicts remain intact.

Prepared source/README NOT RUN wording is historical and left unchanged.
THIS report updates observed execution status for this one split diagnostic;
performance,root cause,full acceptance and promotion remain unverified.

## Controller and independent read-only audit

Before any manifest/index/target existed,controller preparation hit an
absolute-path join error in its own static parse check. Fixed that path handling,
then prepared/froze the controller. No target slot consumed at that stage.
The ONLY target invocation is the one recorded above;no target retry.

Controller delegates classification to the prepared/tested nested offline
verifier instead of duplicating its host regex or stack logic.
367 pins inherit344,add five prior boundary-run packet anchors,nine source/
inventory/README files,five offline utility files and four intervening reports.
Checked source/engine/controller pins before and after target execution.

Independent post-run read-only audit checked all367 pins,manifest/controller,
engines,index count1,stream hashes,and every ordered raw line/context/host/
marker. A separate parser using integer timestamp arithmetic recomputed all
six pairs,12-marker nesting/monotonicity,95/64 counts and zero open scopes,
matching the saved receipt. One combined read-only shell audit attempt exited1
without output;split read-only checks and separate parsing then completed.
No raw/index/controller edits or new target launches were used to audit.

Ignored local packet:
.superpowers/sdd/2026-10-18-p1-api-equality-run/.
Scope/controller/frozen/index/raw are retained locally,not committed or required
to clone portable artifacts. Digests describe local retained consistency,
not independent attestation or authenticated process/security proof.

| Artifact | SHA256 |
|---|---|
| frozen.json |C06ABF60A3F4983348001B77DD7AB309452A5C4B954A8108B3228490382DC92D|
| run.ps1 |452A696A7E8263AE02248930917576AA7ACA9A779ED0520955D67D09B4804459|
| invocations.json |5090D283DE61BBAE35ABF07521BCEBD3AB512007F4A704ED2CD0C4E0693DEE0D|
| PS7 stdout |CFD7AA5FB3F452413DFFAF6ED6EB1C0C7625A36D1CC268FC55DA532615B4C113|
| PS7 empty stderr |E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|
| Candidate/module copy |3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3|
| Original module,unchanged |953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA|
| Prepared marked tests |FC85C3851FCE229C6326658D3480407A0B68C2D7CF14F37C3DA291D2F8EDD820|
|95-line inventory |224A640A30F0B4F0AC891DB81AC545F3B2D621AA1865733EF88530942C807940|
| Nested verifier |803F55C79F40D481C4AC07E5FDEF4C7D024CC34A7C4C79CFBD57A5E29D39B1AA|

## Gates and next recommendation

New split allowance1/1 SPENT. Boundary-run1/1 spent,candidate-full1/2 spent/
PS5 FROZEN;all older spent/frozen allowances and review0/1 unused/ineligible
persist. No independent reviewer,start of main integration,I1/P1/C12/security/
runtime/release acceptance or automatic additional execution.

Next recommend READ-ONLY inspection of candidate API root setup,validation/
owned-snapshot loop and decision loop,and comparison with existing original
diagnostic evidence to choose one further phase-localization hypothesis.
Do not optimize the equality helper on this evidence,prepare another probe,
remove validation guards or rerun product without separately agreed scope.
This result narrows a measured region,not the original root cause.
Recover report/status/runbook and prepared sources/tools from development branch;
Git restores neither ignored records,cached engines,credentials nor replay
authority.
