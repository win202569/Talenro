# Validation/copy attribution — TIMEOUT, UNACCEPTED

After7bb53651,user said continue to the proposed ONE ordinary fresh PS7 diagnostic,
30000ms,no warmup/retry/fallback/PS5. NEW attribution allowance1/1 now SPENT.
All old spent/frozen allowances and review gates remain unchanged.

## Actual outcome

StartedUTC2026-10-18T10:11:59.4361887Z. Parent ordinary;all saved child context
records ordinary,PS7.6.5 Core x64. One child launched. TimedOut=true,exit-1;
wall30085ms (30000 target cap plus85ms elapsed through owned-host exit/cleanup),
CPU40359.375ms,owned-host-exited,streams saved,empty stderr.
Controller exits1 at the planned STOP after saving an incomplete result;no retry
or later host. No saved ControllerError/assertion/fixture failure record.
This does NOT prove no unobserved fault or completed test beyond saved output.

Saved95/106 ordered nonempty records,58/64 PASS,18/22 markers,nine closed pairs,
OpenScopes=[]. Last PASS:P1I1ObservationScalars,next EXPECTED output
PASS:P1I1PassiveDescriptors. No entry marker for that latter body exists in this
copy,so its entry/active instruction/duration are not established.
SequentialIsolation andPlainControls markers/PASS unobserved;do not blame those
groups or invent their entry/end. Closed earlier pairs are not whole-suite PASS.

Classifier CaptureValid=true,ExpectedResult=false,valid-prefix,Termination=timeout.
AttributionAvailable=true because its earlier phase closed and summary was saved;
FormalProjectAcceptance=false,Hypothesis=not-assessed. Preset complete-run gate
was NOT met;do not relax it after observing numeric ordering.

## Retained closed-region attribution

Frequency10000000ticks/sec. One summary at nonempty line76 reports4096 rows:

| Item | Exact ticks | Rounded ms |
|---|---:|---:|
| Validation subinterval sum |179933225|17993.322|
| Snapshot/copy subinterval sum |4732362|473.236|
| Unattributed parent remainder |1332674|133.267|
| Whole ValidationCopy parent |185998261|18599.826|

Exact arithmetic:179933225+4732362=184665587 <=185998261;
remainder1332674. Per-component rounded millisecond values need not sum exactly.
Validation17993.3225ms rounds nearest ties-to-even to17993.322,not17993.323.

These are INSTRUMENTED completed-row subintervals from a later incomplete run.
Validation is numerically larger than copy here,but the predeclared hypothesis
requires a complete106-record diagnostic:remain NOT ASSESSED,not accepted or
falsified under that rule. Summary can guide further analysis only as partial
run evidence. It does not time a particular helper,type query or pure CPU share.

Three per-row timestamp samples,assignments/accounting,parameter/adapter/GC/
history can perturb timing. The133.267ms remainder includes loop/accounting/
unmeasured gaps;it is NOT calibrated timestamp overhead. No empty-loop or prior
run subtraction. Do not compare this snapshot share as an original-run share
or attribute the whole timeout solely to instrumentation.

## All closed spans

Parent/child intervals overlap;do not sum them as independent work.

| Scope | Start ticks | End ticks | Rounded ms |
|---|---:|---:|---:|
| InvalidInputs | 3833956564158 | 3833958786309 | 222.215 |
| Build4096 | 3833958787536 | 3833964256205 | 546.867 |
| RootSetup | 3833964725296 | 3833965135841 | 41.054 |
| ValidationCopy | 3833965176717 | 3834151174978 | 18599.826 |
| Decision | 3834151208297 | 3834156840841 | 563.254 |
| Evaluate4096 | 3833964273801 | 3834156842620 | 19256.882 |
| Assert4096 | 3834156843501 | 3834157097032 | 25.353 |
| Check4096 | 3833964257113 | 3834157099267 | 19284.215 |
| Check4097 | 3834157101204 | 3834157241086 | 13.988 |

## Independent read-only verification

No target rerun or classifier edits. Saved raw prefix independently compared to
the first95 pinned inventory rules,all contexts/hosts/literals/marker order and
nesting checked;nonnegative Int64 ticks globally monotonic. Nine pairs rebuilt
with BigInt differences and ties-to-even rounding. Summary fields independently
parsed,4096 row count/exact sum/parent remainder verified. Every pair and all
attribution values matched the saved receipt. Next expected line checked from
inventory,not guessed as an active statement.
434 frozen pins,manifest/controller/single-invocation index and saved stream
hashes independently held. Raw/controller files retained exactly;no repair,
tail completion or output reinterpretation into full PASS.

Prelaunch only:160 offline classifier and21 static preparation cases passed
again. No target sources imported by those checks. No independent reviewer;
review0/1 remains unused/ineligible.

## Identities / fresh-computer recovery

Ignored local packet:
`.superpowers/sdd/2026-10-18-p1-validation-copy-attribution-run/`.
Retain frozen.json,run.ps1,scope.md,invocations.json,PS7.stdout.txt,PS7.stderr.txt.
These raw/controller files are NOT committed. Git carries the already-pushed
portable source/tests plus this failed diagnostic report/status/recovery,not
ignored raw/runtime cache/credentials or permission to replay.

| Artifact | SHA256 |
|---|---|
| Frozen434-pin manifest | `B20E92F32A122D558DA3F69D5AB40AFDCAFA4B1A2841D0735AAF38ECDA7ABD8B` |
| Controller | `88CE53D8ED77CC5144FF228CCDC6F38BF72D7656696832E24ED7E7C28007A409` |
| Final single-invocation index | `EE36E5EB77200E822035F19E50BC83D63AC78406E3EEE98083B500A71DBC602D` |
| PS7 stdout | `9762AB4F5E1664F4729A7786C62F506CD01F189D579F701B1A7C61E905B45D68` |
| PS7 stderr(empty) | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |
| PS7 engine | `362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139` |
| Target entry | `12D6F5460F97B9E09453B0387C30E3D20CE19B2EDF1CDF4DBF4E77DA31561422` |
| Original product module | `953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA` |
| Isolated candidate module | `3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3` |

Controller derives from frozen ownership-marker controller;only new preparation/
inventory/verifier/434-pin routing and attribution receipt fields changed.
Process start/wait/drain/owned-host kill/stream saving unchanged;one slot saved
before one launch. No source/test/guard/timeout changes this turn.

## Recommendation / remaining gates

STOP target execution;no auto retry,timeout increase,PS5/frozen-slot reuse,
additional per-row markers or acceptance promotion.
Next recommend READ-ONLY contract analysis of duplicate type queries in the
Token/Passive path:can a later LOCAL reuse preserve every exact-type/passivity/
descriptor/bounds/refusal-order check? This is an analysis question,not a fix,
candidate preparation or claimed root cause. Keep snapshot/owned-result checks.
Do not infer that fewer calls guarantees the needed latency improvement.

Attribution/ownership-marker/three-phase/split/boundary each1/1spent;
candidate-full1/2spent/PS5FROZEN,all other historical budgets unchanged.
Review0/1unused/ineligible. Candidate/I1/P1/main/C12/security/runtime/release
UNACCEPTED;original timeout unresolved. Save only necessary failed-checkpoint
documents on the existing development branch,no main merge or release.
