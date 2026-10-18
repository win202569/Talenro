# Three-phase diagnostic — phase measurements closed; FULL RUN TIMEOUT

After9c867a1b,user approved a NEW ONE-host ordinary fresh PS7 diagnostic,
30000ms cap,no warmup/retry/PS5 fallback or spent-slot reuse. Sole1/1 SPENT.
No target source/product/assertion edits,reviewer or main merge. No second run.

## Overall outcome: failed complete-regression verdict

PS7 7.6.5 Core x64,ordinary parent/child admin=False.
Started2026-10-18T09:43:08.0311704Z,TimedOut=true,exit-1 following owned kill.
Wall30081ms (configured30000ms plus81ms termination/collection overhead),
all-thread CPU40562.500ms,empty stderr,owned-host-exited,streams saved.
95 nonempty lines/59 PASS,not required101/64.
Last PASS: P1I1PassiveDescriptors.
Next expected literal: PASS: P1I1SequentialIsolation.
CaptureValid=true,ExpectedResult=false,Classification=valid-prefix,
FormalProjectAcceptance=false. This is NOT a complete diagnostic/suite PASS.

No assertion/fixture/classifier error was saved. All measured4096 scopes ended
and4097 refusal completed;later ownership sequential-isolation/plain-controls/
complete and final wrappers lack completion output. Last PASS/next expected
record do NOT identify the exact active instruction or prove the next case
started. No later test duration,exact cutoff position or permanent hang inferred.

## Closed phase measurements, not full acceptance

Stopwatch.Frequency=10000000 ticks/sec,recorded outside the target stream.

| Closed scope | START ticks | END ticks | Wall ms |
|---|---:|---:|---:|
| InvalidInputs |3816652648292|3816655087065|243.877|
| Build4096 |3816655089363|3816661490115|640.075|
| RootSetup |3816661765164|3816661862797|9.763|
| ValidationCopy |3816661865034|3816830676520|16881.149|
| Decision |3816830677539|3816837219159|654.162|
| Evaluate4096 parent |3816661524615|3816837220982|17569.637|
| Assert4096 |3816837221850|3816837565906|34.406|
| Check4096 parent |3816661491815|3816837568220|17607.640|
| Check4097 |3816837569593|3816837747016|17.742|

18 markers/nine closed pairs,no open INSTRUMENTED scopes. Empty OpenScopes
does not mean the full process/test sequence finished:markers cover only the
selected boundary work,not every later test.

H-VC ordering is supported for these completed phases:ValidationCopy16881.149ms
is larger than RootSetup9.763ms and Decision654.162ms. The original formulation
referred to a prospective complete diagnostic;this run supplies closed phase
evidence but does NOT satisfy the complete-run condition. Keep those claims
separate. Combined validation/copy is the largest measured API internal phase.

ValidationCopy includes BOTH per-row guards and owned snapshots/list appends.
It does NOT measure pure validation,copy alone,a specific SDK helper or cache
behavior. Root includes root/member/array checks,metadata and setup;decision
includes result/context setup and the second loop. These are elapsed-wall,
not CPU-attributed intervals. Parents overlap children;do not sum all table
rows as disjoint work or derive an unfinished later scope from the deadline.

The prior two diagnostics completed28437/28208ms;this differently instrumented
entry timed out30081ms later in the suite. Not matched performance arms,not
stable slowdown/speedup proof or original-timeout repair. All earlier outcomes
remain valid at their original scopes.

## Source identity and measurement limits

Original candidate13-function module remains3537E9F5;only the previously
prepared unique evaluator clone is used by4096. The clone was defined once at
test-script startup,no local module reload/per-row instrumentation/new guards.
Original product953CC2A8 and all input/assertion expressions remain unchanged.

Fresh clone name/source/body/history,additional startup definition load and
six inner output statements can affect runtime/compilation. Source reversal
establishes static operations/check identity,not runtime binding/cache identity.
Prepared source/README NOT RUN wording is left unchanged;THIS report updates
only this observed execution,not promotion/performance/full acceptance.

## Integrity and independent read-only verification

388 pins inherit367,add five previous split packet anchors,ten new preparation
files,two new offline verifier/test files and four intervening reports.
Pinned target/controller/verifier/engine identities were checked pre/post run.

Post-run audit checked all388 pins,frozen manifest/controller,engines,index count1
and both stream hashes. Separate read-only parsing matched every saved line to
the frozen101-line PREFIX and independently verified18-marker three-level
nesting,all integer timestamps,59 PASS,nine closed pairs,zero open scopes.
Every exact START/END and recomputed rounded interval matches the receipt.
The first independent parser used a different half-rounding convention for
Check4096 (exact17607.6405ms). Corrected ONLY that read-only audit arithmetic
to nearest/ties-to-even,then matched17607.640ms;no raw/index/controller change
and no target rerun. Original receipt was already correct.

Local ignored packet:
.superpowers/sdd/2026-10-18-p1-three-phase-run/.
Scope/controller/frozen/index/raw retained locally,not committed or clone
prerequisites. Digests document local consistency,not independent attestation.

| Artifact | SHA256 |
|---|---|
| frozen.json |2311AF3348CF7292ADB020958049C22289AF453A40EA372963BF877F4106291E|
| run.ps1 |5471DE401D2F99A79E2D8BB17E81D93B73229E5936F5E50CE8DEA487F90585BB|
| invocations.json |02A8907BE3E1D12B3013EEF77ECABE4B6BAE314D0EBF034633D4EF07526A56D9|
| PS7 stdout |C691AE63BCA4BCDEE56ECD3265E4059EAD2A53FEFD406774B6BFA7E7FEE33CCD|
| PS7 empty stderr |E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|
| Candidate/module copy |3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3|
| Phase clone |6610F6F4396B9C0B1DF6AA202A63572F45EA037C175F30DF85973146F1881AE6|
|101-line inventory |7AC1F1B990AB0E1A63838ED7690CA857A11ED50DC0B8C696151649A4DB8A9944|

## Gates and next recommendation

New three-phase1/1 SPENT. Split1/1 and boundary1/1 spent,candidate-full1/2 spent/
PS5 FROZEN;all older spent/frozen allowances and review0/1 unused/ineligible
persist. Candidate/I1/P1/main/C12/security/runtime/release unaccepted.
No authority to retry,raise timeout,release frozen PS5 or dispatch review.

Next recommend READ-ONLY inspection of the later sequential-isolation source/
workload and existing complete versus timeout evidence before designing another
diagnostic. Address this new full-suite frontier explicitly rather than declaring
success from closed API phases or blindly continuing finer timing. No speculative
guard removal,product optimization or replay. Root cause remains unresolved.
Recover portable sources/tools,this report,status and runbook from development
branch,not ignored raw,cached engines,credentials or renewed execution authority.
