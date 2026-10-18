# Four-case retention diagnostic — instrumentation error; INCOMPLETE

Outcome=INCOMPLETE;DiagnosticComplete=false;SchemaValidated=false.
NewDiagnosticHosts=1/1SPENT;FormalAcceptance=false;PromotionApproved=false.
No candidate/product invocation or edit,retry,new reviewer or main merge.

## Approved scope and preparation

The user approved preparation+ONE attempt from the four-case design recorded in
current-status.md at4b88ae0ab7721521188ca0aab1b7c08a52de7c58. Scope was held
String-view control,raw Int32 control,raw String field and pre-stored explicitly
wrapped String field. This was fixture-only capability characterization,not an
ownership regression or candidate repair. The previous two frozen slots were
not used or renewed.

Five portable files prepared in docs/diagnostics/isolated-fuzz/string-retention-characterization/:
probe,capture,output schema,README and LF attributes. The schema fixed six-line
complete envelope,four ordered19-field records,Boolean/null/enum constraints and
positive-control invariants BEFORE execution. Unknown observations were not
filled in as expected values. Static AST parsing passed;probe command inventory
had no dot-source,dynamic command,product helper/import or compiler invocation.
This static check did not establish runtime correctness of the new instrumentation.

Probe retained views/collections/field descriptors/base references and used fixed
public reflection for identity/type measurement. First-control calibration used
two wrappers around ONE base to distinguish wrapper identity from base identity.
It then attempted an incorrect direct BaseObject property read. This authoring
mistake is in the diagnostic code,not evidence of a candidate defect.

## Actual receipt and output

Fixed ordinary Core7.6.5/x64 engine SHA256
362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.
Command:fixed engine -NoProfile -NonInteractive -File probe.ps1,absolute pinned
path in local receipt. No elevation/execution-policy change. Slot preconsumed
BEFORE launch. StartedUTC=2026-10-18T11:34:21.5015162Z.
Elapsed12462ms,exit1,TimedOut=false,owned-host-exited,StreamsComplete=true,
PostPinsIntact=true. stdout1line;stderr320bytes. ControllerError=null means
capture/integrity checks did not report a controller error,NOT that the child
succeeded. The controller correctly refused completion due to child failure.
Elapsed is a receipt fact,not a performance or delay-cause conclusion.

Exact stdout:

```text
HOST: 7.6.5 Core x64=True admin=False productCalls=0
```

Saved stderr reports the diagnostic's catch Write-Error site(probe110) and:

```text
The property 'BaseObject' cannot be found on this object. Verify that the property exists.
```

No FACT records or diagnostic-complete terminal were emitted. First control
was not completed;remaining three cases were not reached. No retention capability,
positive-control success or String wrapper hypothesis was dynamically established.
No candidate imported/called;old candidate1971F808 and its unrun39 behavioral
cases remain unchanged. The earlier String FIXTURE_INVALID and earlier Int32
descriptor-contamination witness are not rewritten by this failed diagnostic.

## Readonly diagnosis, not a repaired result

Probe54 contains direct $left.BaseObject/$right.BaseObject in the shared-base
identity calibration. This is the first such property access on the current
control path. The other explicit BaseObject accesses occur later through the
PSObject view,after that calibration. Source/error comparison identifies the
direct calibration property access as the concrete instrumentation problem.
The new code incorrectly treated adapted member access on a wrapped primitive
as if it exposed the wrapper's SDK properties directly.

The saved catch output does not include the original invocation stack or exact
original failing expression. Thus source-order diagnosis must not be presented
as a separately captured runtime stack. Preliminary checks before this expression
do not constitute a validated identity observer or a completed positive control.
No post-failure modification,rerun,stream suppression,oracle relaxation or
conversion of the error into a successful capability measurement occurred.

## Frozen identities and reconciliation

Local ignored packet:.superpowers/sdd/2026-10-18-p1-string-retention-characterization/.
Parent reconciled source/receipt/streams as DATA only. All557 frozen entries
matched before/after:548 inherited+4 prior packet anchors+5 portable files,
including repeated paths rather than557 distinct files. Exact one-line stdout,
stderr error content/byte count and both hashes matched the saved receipt.
No target/helper/constructor experiment ran in the parent during reconciliation.

|Artifact|SHA256|
|---|---|
|frozen.json|B9C0DDB343E54B89A152C3CCF65214C9513050D0E10F321651F9C0E49FD8D08C|
|invocation.json|2301A273E8727F67B9AB362A8E60E4E318D49AFA927B27A58BB47D940B437082|
|stdout.txt|A2D61CB8B2644E0CEE9CAD025496FB4C10184E238F09530333666CCF4F79610A|
|stderr.txt|AFEE33A515998AD4ED0B268BCC487FBBB774E9ADA0544191D4E6BA7B1CCE34DE|
|probe.ps1|9D0962DCEB451C1A3ACB44BA4887D19DF34F6E620E589D931CE74F7E420C4AB7|
|capture.ps1|FAB7DEBAAC19EC6675E12319AF0B092540B22D9E64C75DA83E657857CC6C82A9|
|output-schema.json|238C1781FBC893BE16A554018FE94D4555EED8720FE9130DF6472BA655F974DF|
|README.md|FCEE3ECCE1990B90FA5E2AF9AF975BF5FDE1697B4F5BA0CCC03DFF79C38527CB|
|.gitattributes|A79691A93B46E49CE460C26EF22AFCC03D6ECA1E63BF2EDBC20E96159510F6C9|

## Rulings and bounded next design — NOT APPROVED/IMPLEMENTED

Ruling:explicit single-host/no-retry scope overrides generic full-suite/RED-GREEN
repair/final-review defaults. Cost:unfinished diagnostic,not a repaired candidate
or independently validated regression. Ruling:public-view identities are proxy
observations,not a full binding trace;invalid instrumentation stops. Cost:no
mechanism conclusion from this host. Ruling:PowerShell-native existing isolation/
capture,not Bash/install/new worktree/packet deletion. Cost:no generic skill-runner
claim;all historical records retained. No independently graded Minor findings.

Recommend a BOUNDED diagnostic-only correction in a NEW sibling copy:replace
the direct calibration BaseObject reads with the fixed PUBLIC PSObject.BaseObject
getter,invoked through the existing public MethodBase.Invoke trampoline to retain
the actual wrapper target. Add only that fixed metadata/helper and call-site
correction;keep all four cases,identity calibration,positive controls,field
semantics,mutation checks and output schema unchanged. Do not merely skip the
failing check or change the candidate to accommodate the observer.

Proposed NEW max1 ordinary fresh PS7 host,same30s+2s exit+2s drain,ZERO product
imports/helpers/decisions,no preparatory probe/warmup/second host/retry. Freeze
new source/controller identities plus all557 prior entries and raw anchors;
persist slot consumption first. Any error stops and is retained. This proposed
correction/preparation/one attempt needs fresh explicit approval;NEW0/1UNAPPROVED_UNUSED,
not a reuse of the spent diagnostic1/1 or candidate repair's2 frozen slots.
No corrected source/controller has been created this turn. No extra spec/plan
artifact is needed for this bounded existing-code correction.

Candidate repair allowance remains1/3SPENT+2/3UNUSED_FROZEN;all older budgets,
I1/review/runtime/security/main/C12/release boundaries remain. Git saves necessary
diagnostic source/schema and this failed result/status/recovery,not ignored raw,
runtimes,credentials or authority to rerun. Restore latest remote-confirmed dev
branch,not main;missing local state cannot replenish spent/frozen allowances.
