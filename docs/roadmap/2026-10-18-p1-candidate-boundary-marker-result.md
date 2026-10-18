# Candidate outer boundary diagnostic — complete capture; classifier defect retained

After3fd072f1,user approved a NEW ONE-host ordinary fresh PS7 diagnostic,
30000ms cap,no warmup/retry,PS5 fallback or frozen-slot reuse. Sole1/1 slot
spent. No target/product/test changes,reviewer,main merge or second run occurred.

## Observed execution and independent read-only verdict

PS7 7.6.5 Core x64/admin=False,started2026-10-18T09:24:09.7423792Z.
Wall28437ms,all-thread CPU36312.500ms,exit0,no timeout,empty stderr,
owned-host-exited.91 ordered nonempty lines/64 PASS,final ShapeLookupRegression.
All eight prepared boundary markers appeared at positions66–73,monotonic Int64
timestamps. Stopwatch.Frequency=10000000 ticks/second,recorded in the receipt
outside the test stream and rechecked in the same-machine PS7 environment.

| Closed boundary | START ticks | END ticks | Wall ms |
|---|---:|---:|---:|
| InvalidInputs |3805253038629|3805255489347|245.072|
| Build4096 |3805255491855|3805261955468|646.361|
| Check4096 |3805261957207|3805430566472|16860.926|
| Check4097 |3805430568251|3805430715358|14.711|

Intervals independently recomputed from retained raw tick differences,not
controller pair fields. H-boundary is supported for THIS completed diagnostic:
Check4096 is longest of these four regions. It includes argument construction,
API evaluation and full-result equality,not API-only time. No candidate internal
profile identifies validation,copy,decision,reflection,JIT or GC shares.

All four closed spans exclude preceding tests,later groups and startup/collection.
Do not allocate the remaining wall time to one cause. CPU covers all process
threads and can exceed elapsed wall. No per-region CPU claim.

## Frozen controller classifier defect, NOT a target failure

The capture controller's Host matching regex accidentally contains two
backslashes before the dot,where the prepared inventory has one. Thus normal
`HOST: 7.6.5 Core x64=True` fails its line2 rule. It records:
CaptureValid=false,ExpectedResult=false,MismatchLine=2,MarkerCount=0,
ClosedPairs=[],OpenBoundary=null. Those fields reflect early parser termination,
NOT absence of the eight saved markers or failed test assertions.

The controller exited1 AFTER preserving full raw streams and final receipt.
Its bounded launch/wait/termination handling is distinct from this post-capture
classifier. Neither controller,manifest,index nor raw was edited or rerun.

Independent read-only audit instead used the PREEXISTING prepared
expected-output.json HostRule,not a newly relaxed host/security rule.
Checked EVERY ordered line,all literal/context/host/marker rules,64 PASS,
eight ordered/nonnegative/nondecreasing Int64 ticks,four pairs,exit0,under-cap
wall,no timeout,empty stderr,ordinary parent/child and owned exit.
It establishes complete diagnostic capture separately;it does not overwrite
the frozen false classifier verdict or promote it into formal acceptance.
Future classifiers need offline regression tests before another target launch.

## Interpretation and fidelity

Previous uninstrumented candidate PS7 timed out30090ms at65 lines/48 PASS.
This outer-marker copy completed28437ms with91 lines/64 PASS. It is not a
matched performance experiment or proof that markers improve speed. Input,
assertions,product and common helpers/loading remained unchanged,while added
statements/script-body shape and ambient runtime conditions can perturb timing.
No original-entry rerun was performed and its prior timeout remains valid.
One completed diagnostic does not demonstrate stable latency or a repaired cause.

Prepared source/README historical NOT RUN wording remains untouched. THIS report
updates observed diagnostic execution status only. Source hashing and prior
exact marker reversal are unchanged. No original replacement,cache/security
shortcut,assertion weakening or function reload occurred.

## Identity and retained records

344 pins inherit327,add five prior candidate-full packet anchors,nine portable
diagnostic files,and three full-result/read-only/preparation reports. Rechecked
all344 pins,manifest/controller,engines,one index entry and both stream hashes.
No extra diagnostic host launched for audit and no new PS5 files were created.

Local ignored packet:
`.superpowers/sdd/2026-10-18-p1-candidate-boundary-marker-run/`.
Controller/scope/manifest/index/raw retained locally,not committed or clone
prerequisites. Digests record retained consistency,not independent attestation.

| Artifact | SHA256 |
|---|---|
| frozen.json | A78E06D072F13756CFA9BB95A189CF88398257FE6549C4981FAEAC1B68637F98 |
| run.ps1 |18EF13E927D43D59ADCBB97A209DAF2873FD82F53827883928B209ECBC3CB518|
| invocations.json |8FC69424C59CD9B58856820C93AEADEC979102822248B9CD50776A3231BB8CFE|
| PS7 stdout |849E452D1D4EC53D4487F2F7CD4310DF67ABDC11A9CA7E534D121971F48AB0F6|
| PS7 empty stderr |E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|
| Candidate/module copy |3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3|
| Original product,unchanged |953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA|
| Prepared marked tests |EB4BA718BC960061F948FEB184DE18676110B264E34B1600BC01EADD11F87018|
| Prepared output inventory |6FF2D34D86C43ABFCDA28A6502ED7735D539C113665B902FC936C202E4F8FCF3|

## Gates and next recommendation

New boundary-run allowance1/1 SPENT. Prior uninstrumented candidate-full1/2
spent/PS5 FROZEN;all older spent/frozen allowances and review0/1 unused/ineligible
unchanged. FormalProjectAcceptance=false;candidate/I1/P1/main/C12/security/
runtime/release gates NOT cleared. No second target run is authorized.

Next recommend correcting a SEPARATE future offline result verifier and testing
it against saved output plus synthetic bad host/context/marker/order/tick cases.
Do not edit old frozen records/controllers or rerun product to test parsing.
This closes the demonstrated tooling defect before proposing deeper measurement;
it does not fix or independently identify the product-timeout cause.
Preserve portable reports/status/recovery on the development branch. Git restores
neither ignored raw/controllers,runtime caches,credentials nor replay authority.
