# PS5 evaluation read-only trace — validation/copy is a hypothesis, not a proven cause

Afterffe3735e, user requested continuation. ZERO product/test executions,new
diagnostic children,code/tool edits or budget spend. This turn reads current
source and retained receipts,then saves portable findings/status/recovery only.
Coordinator source audit,not independent review or acceptance.

## Backward trace for the exact boundary input

Caller builds4096 plain six-field time observations:Sequence=i,AtMs=0,
SessionLabel=s,ClockLabel=synthetic-ms,Actor=W,Kind=time;StartMs=0.
P1-BoundaryCheck4096 retains original function-local call/result/assertion scope.
Saved Evaluate4096 START9427.186ms has no END;assertion has no START.

Source has three relevant phases,none separately timed on current PS5:

| Phase | Source | Fixed-input work |
| --- | --- | --- |
| Root setup | Get-OfflineSupervisionDecision lines104–120 | Root schema/scalars/array checks;new per-evaluation metadata/list |
| Validation and snapshot copy | for loop line121,copy line151 | Validate every record and scalar;copy all six values into owned records |
| Decision loop | for loop line157 through result return | Check bindings/actor/sequence/time;time branch leaves state Prepared |

This ordering means decision work begins only after ALL4096 records are
validated and copied. For this literal input,no arm/release/cleanup/persist
transition occurs:time actor W is allowed,all bindings match,Sequence=i and
AtMs=0 remain valid,synthetic deadlines are not crossed. Expected terminal
LastSequence=4095,LastAtMs=0,State=Prepared matches the original literal test.

Synthetic time is INPUT data,not the real stopwatch killing the test host.
Changing synthetic DeadlineMs or StopAtMs cannot extend the external30s budget
and would change product semantics rather than resolve measured evaluation cost.

## Static termination/work findings, not runtime proof

AST parse finds13 function definitions and exactly two for loops (lines121/157).
Both iterate the bounded record count. Validation adds only to a distinct list;
decision does not append to that list. Fixed-schema/token/member enumerations
are additional bounded work for this known plain input. No explicit sleep,poll,
file/network call or external-process wait occurs in the current source path;
managed reflection/member enumeration and runtime work still have costs.
No direct recursive function-name call appears in the AST inventory.
This is NOT proof against runtime stalls,indirect resolution issues,JIT/GC,
contention or an incorrectly measured active scope.

ShapeSets is per evaluation,with only root/time keys for this input;it does not
grow with the4096 distinct record instances. Cache hits reuse allowed NAMES,
not acceptance. Records and values are still checked and copied every time.
Prior source accounting retains18 type-helper calls and10 scalar extended-member
checks per valid time row,plus root setup. Allocation/copy/member enumeration
and repeated dispatch remain candidate costs,not measured current PS5 attribution.

Do not remove the initial full validation pass,return caller-owned inputs,skip
later scalar checks,stream decisions before all validation,or cache accepted
input objects based on this audit. Such changes could weaken descriptor/ownership/
refusal guarantees;current evidence has not justified them.

## Single next hypothesis and falsifier

H1:the unfinished PS5 evaluation is primarily in validation/snapshot-copy work,
with the decision loop not yet reached at cutoff.
Support is qualitative current-path work density plus an OLDER PS7 diagnostic
on a different product copy (validation17480.582ms,decision670.873ms).
That old profile does NOT measure current candidate953CC2A8 on PS5 and must not
be imported as its timings or asserted root cause.

A useful future measurement must distinguish root setup,validation/copy,and
decision boundaries on CURRENT candidate/current PS5 in the same complete order.
H1 is falsified for that run if validation ends and decision starts before the
cutoff. If validation starts without END,only that phase is localized;individual
helper or allocation cost still needs evidence. If evaluation completes,report
actual phase timings without claiming a stable fix or unique cause.
No new marker code,design approval or execution allowance is created by this report.

## Integrity, budgets and recovery

Read-only checks matched219 retained pins/current candidate and raw stdout/stderr
hashes. Latest invocation index stays:
AA151B62F1C1EAA28B8831368F5F84B238FD8691F57E5DCD66ADD7312BCE98AA.
Product stays:
953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA.

PS5 subphase1/1 and group1/1 spent;original-full2/2 spent (PS7 PASS,PS5 timeout);
earlier PS7 CPU1/1,targeted2/2 spent;shape2/3/one frozen;prior independent full1/2/
one frozen;reflection3/4/one frozen;extended-view3/7/four frozen.
Review0/1 unused/ineligible;all other budgets unchanged.
I1/P1 unresolved,candidate UNACCEPTED;no main/C12/project/live/runtime/release
acceptance. ZERO retries,warmups,old-slot use or product fixes this turn.

Save findings/status/recovery on existing development branch. Clone restores
portable source/reports,not ignored raw/private runtime/credentials or authority
to replay. Next scope a discriminating measurement only after the user approves
its question/probe and independent bounded allowance;do not rerun the unchanged
suite,raise30s or optimize from an unproven phase attribution.
