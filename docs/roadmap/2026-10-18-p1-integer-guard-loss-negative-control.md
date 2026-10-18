# Isolated integer guard-loss negative control — detected on both engines

Aftera02dfaa6, user approved the recommended NEW two-host negative control:
ordinary fresh PS7 then PS5 once each,30000ms per host,no retry/warmup/original
product edit/test weakening/reviewer/old-slot reuse. Exact intended assertion,
not an arbitrary exception, was frozen before launch. Both2/2 slots spent.

## Isolated mutation, original source unchanged

Added `docs/diagnostics/isolated-fuzz/integer-guard-loss-negative-control/` with
three explicitly diagnostic files:deliberately defective product copy,byte-identical
direct test copy,and warning README. This fixture is NOT a product candidate;
never import it in normal/full acceptance paths. It has no serving/live effect.

Only semantic mutation:remove the integer helper's
Assert-SupervisionPolicyPassive call,retaining exact int/long and bounds checks.
Two leading comments warn of the intentional defect. Preflight restored the
removed call/removed warning header and matched the full original product after
newline/end normalization. AST audit also matched all12 other function bodies.
No type-result reuse candidate or original product modification occurred.

Test copy SHA256 equals the original direct test exactly:
E442D1DDF20C901A511097D210FA66A23AAB48F838CEB262382599BFEDC21C98.
The unchanged original product remains
953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.

## Expected and observed behavior

Both hosts emitted exactly nine ordered lines:CONTEXT,HOST,six PASS markers
through P1IntegerWrongTypeNoCallbacks,then:

`INTEGER-ERROR: ASSERTION FAILED: P1IntegerNoteDecorations missing refusal`

| Engine | Started UTC,2026-10-18 | Wall ms | Process CPU ms | Child exit | Observed result |
|---|---|---:|---:|---:|---|
| PS7 7.6.5 Core x64 | 09:06:03.6661540Z | 828 | 734.375 | 1 | Exact intended assertion |
| PS5 5.1.26100.9711 Desktop x64 | 09:06:09.1745536Z | 573 | 515.625 | 1 | Exact intended assertion |

Both:no timeout,empty stderr,parent/child admin=False,owned-host-exited.
NegativeControl=true,ControlExpectedFailureObserved=true,CaptureValid=true,
ExpectedResult=true,FormalProjectAcceptance=false. Child exit1 is the required
negative-control outcome,NOT a normal test PASS. The controller completed normally
only because the exact anticipated failure matched. Any different error,exit0,
timeout or integrity drift would have stopped/frozen the next slot.

The [unmodified baseline](2026-10-18-p1-integer-direct-baseline.md) passed14 groups
on both engines;the identical suite now detects the omitted direct helper guard.
This establishes sensitivity to THIS mutation at the first NoteProperty vector.
Later Alias/Executable/error-order/transition groups were not reached in these
negative runs. Do not infer negative sensitivity for every later guard,CodeProperty
coverage,unique timeout cause,performance savings or safety of a future candidate.
No timings here are a full-suite benchmark;each run intentionally stops early.

## Integrity and remaining gates

Independent read-only checks matched301 pins,manifest/controller,engines,receipts,
four streams and every expected line. Rechecked exit1/under-cap/zero stderr/owned
exit,unchanged original product/test,byte-identical copied test and12 unchanged
function bodies.301 pins inherit290,add seven baseline packet anchors,three
fixture files and baseline report. Old records and verdicts remain intact.

Ignored local packet `.superpowers/sdd/2026-10-18-p1-integer-guard-loss-control/`
retains scope/controller/frozen/index/raw,not committed or portable prerequisites.

| Artifact | SHA256 |
|---|---|
| frozen.json | 904B290BB65DA9F5511CC0AD92DFAA9E7FAD82B55DC6AA68B6099F5B95E0CDFD |
| run.ps1 | 0EF62DA1400ED15471D0B7A2C0E1F9DEEA1A9828E9261A9050869D5FA33BE0A0 |
| invocations.json | C3CFF474F7DC8A2D9A1AF3F20A53EB25DDAE064AF3E29EBF7D489F536E14B3AC |
| Deliberately defective product copy | C187D6B1764ED561047B40F78E4567EE97A175E660AB0C1BF39E44DB4EDC8485 |
| PS7 stdout | 285FE2CBD0796DEF7D2C92BEA7E2783BF8C8A15C8E6BDE46F83B6477EC124569 |
| PS5 stdout | 69F2EF7B8B3EE81A611CA61FF300994B2CD25818F9213CB963A0950D1A371EB6 |
| Both empty stderr | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

All older spent/frozen allowances remain,including loading1/2 spent/Inherited
FROZEN and review0/1 unused/ineligible. Original PS5 full-entry timeout and
I1/P1/candidate/main/C12/security/runtime/release gates remain unresolved.
This controlled red result plus baseline green is test-sensitivity evidence,
not a completed optimization implementation/verification cycle.

Next recommend preparing a SEPARATE minimal candidate copy following the reviewed
integer-local specialization:query exact type once,retain all equivalent SDK
extended-member/descriptor checks before bounds,keep common Passive/Token/Bool,
interfaces and all tests unchanged. Do not replace the original product or use
this defective control as the candidate. Preparation alone would grant no new
execution budget;candidate verification needs its own explicitly approved scope.
No candidate implementation or additional launch was performed in this turn.

Recover fixture/source/reports,current status and runbook on the existing
development branch;cloning restores neither ignored raw/controllers,runtimes nor
replay authority. Keep deliberate-defect labeling intact during recovery.
