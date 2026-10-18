# Fixed-getter retention diagnostic — four cases complete;not candidate acceptance

Outcome=DIAGNOSTIC_COMPLETE;SchemaValidated=true;NewDiagnosticHosts=1/1SPENT.
FormalAcceptance=false;PromotionApproved=false;I1CandidateDisposition=NotAccepted.
ZERO candidate/product imports,helper/decision calls or source edits.

## Authority and exact correction

After1749770ddcff82c0edbdc6ad2fb2893ce8613f0e,the user approved the bounded
diagnostic-only getter correction and ONE same-scope attempt. New sibling leaf:
docs/diagnostics/isolated-fuzz/string-retention-fixed-getter/.
Five portable files:probe,capture,byte-identical output schema,README,LF attributes.
Old failed diagnostic,old String fixture stop and all candidate files untouched.

Probe diff is limited to fixed public PSObject.BaseObject getter metadata,
R-Base through the existing public MethodBase.Invoke trampoline,and the two
calibration property reads at one call site. Mechanically reversing those three
edits exactly reconstructs the old probe. No case,identity calibration,positive
control,field semantics,mutation guard or output-schema relaxation.
Controller changes only new packet identity/prior anchors/count;capture and
schema-validation rules retained. Both scripts parse;schema SHA matches old.
Static command inventory excludes product/import/dynamic/compiler invocation.
No preparatory execution,baseline,warmup,second child or reviewer.

## Actual single execution

Fixed ordinary Core7.6.5/x64 engine SHA256
362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.
Command:fixed engine -NoProfile -NonInteractive -File probe.ps1,absolute pinned
new-file path in local receipt. No elevation/execution-policy/security changes.
Slot consumed BEFORE launch. StartedUTC=2026-10-18T12:38:21.6453050Z.
Elapsed776ms,exit0,TimedOut=false,owned-host-exited,StreamsComplete=true,
stderr0bytes,ControllerError=null,PostPinsIntact=true. Six stdout lines:fixed HOST,
four ordered FACT records,exact DIAGNOSTIC_COMPLETE terminal. All frozen schema
field/type/enum/null/control/fidelity requirements matched;not a production PASS.
No timing/performance comparison with old failed run is inferred.

## Four measured observations

|Case|Actual stored CLR type before/after|Same held/reread view and member collection|Property/payload visible through reread|Capability|
|---|---|---|---|---|
|Held String-view control|System.String|true / true|yes,mutation retained|RETAINED|
|Raw Int32 field control|System.Int32|true / true|yes,mutation retained|RETAINED|
|Raw String field|System.String|false / false|no,null absence markers|NOT_RETAINED|
|Pre-stored wrapped String field|System.Management.Automation.PSObject|true / true|yes,mutation retained|RETAINED|

For ALL four:BaseIdentitySame=true,FieldDescriptorSame=true;unchanged exact base
type/value;held descriptor/payload identity and payload mutation retained. Public
view type was System.Management.Automation.PSInternalMemberSet. The raw String
case also had PostViewSameAsReread=false;the other three had true. No field was
reassigned after initialization or attachment. Base values were owned nonempty
clean Strings or owned boxed Int32(0),not shared string/primitive literals.

Raw String reread had RereadDescriptorRetained=false,RereadPayloadSame=null,
RereadMutationRetained=null. These are absence observations,NOT false payload
equality or an ownership-isolation pass. The held view still retained its own
descriptor,payload identity and changed contents. Both positive controls completed.
The pre-stored wrapper case retained the same field descriptor/base and actual
PSObject storage type across reads while preserving the mutation. Identity
calibration on two wrappers of one base completed before any FACT emission.

Exact terminal:RESULT: DIAGNOSTIC_COMPLETE cases=4 productCalls=0 acceptance=False.
No copied/reconstructed subject objects or candidate-generated expectations were
used;only primitive fact records were serialized. Measurements were accepted
against a pre-frozen observation schema,not answers chosen after seeing output.

## Confidence and effect on prior findings

The raw String boundary now has a dynamic public-view witness:unchanged stored
String base/field descriptor,but different public views/collections;metadata stays
on the retained attachment view and is absent through new field reads. Pre-stored
explicit wrapping changes that fixture capability on this snapshot. This supports
the [source-supported explanation](2026-10-18-p1-string-retention-readonly-analysis.md)
for why the earlier String attachment-fidelity assertion was not a valid isolation
measurement. It is NOT a complete raw CLR binder trace,universal engine statement,
or replay of the old candidate/helper call path.

The earlier instrument error's missing-BaseObject site is no longer encountered
on this corrected path. The old failure remains preserved;do not retroactively
mark its1/1 run successful. This is finite author verification,not independent
review or a proven reusable regression gate.

Candidate1971F808 remains UNVERIFIED/UNACCEPTED. Its decision entrypoint/owned
result builder and39 behavioral cases were NOT run by this diagnostic. The old
Int32 descriptor-contamination witness remains unchanged and is not explained
away by String-specific behavior. No I1 closure,candidate fix/promotion,full API
rerun,PS5,benchmark,security/runtime/main/C12/release acceptance follows.

## Readonly reconciliation and identities

Parent read six saved lines/four fact records/receipt as DATA only,checked actual
table observations and fixed schema/fidelity,matched both stream hashes and all
566 frozen entries before/after:557 inherited+4 prior raw/manifest anchors+5 new
files,including repeated paths. No parent product/helper/constructor experiment.
No source/schema/controller edits or extra target after execution.
Ignored packet:.superpowers/sdd/2026-10-18-p1-string-retention-fixed-getter/.

|Artifact|SHA256|
|---|---|
|frozen.json|7EA1FBB567215A231E1D49649D8631AF3029BF551C6DDF04EAD0FA4717E47C15|
|invocation.json|443D954016ECFEA58322CA290F0F1C1C68D5CADE53EACA8A9FDD91C980393B02|
|stdout.txt|AAD79341B92C1B4779DB469871CA2198F32AE52F11194C8C727C75FFD003BCB8|
|stderr.txt(empty)|E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|
|probe.ps1|99FC2EE4E365018370EFF6B475625360ED205EBB165A6AE2B0CBE795BE6722DD|
|capture.ps1|8F4D9E9D97BC91E99920DB58D198218ECBCFF5F04BF0DDB5C5CED4321A7C2342|
|output-schema.json|238C1781FBC893BE16A554018FE94D4555EED8720FE9130DF6472BA655F974DF|
|README.md|5F9E725F48486CDE5A7A982452CF868BDDC51FD28F87C1A23F932C4F9537E847|
|.gitattributes|A79691A93B46E49CE460C26EF22AFCC03D6ECA1E63BF2EDBC20E96159510F6C9|

## Rulings and next recommendation

Ruling:retained previous failed diagnostic+unchanged assertions replaces another
baseline/RED;no full-project/reviewer or repeat outside this single allowance.
Cost:author-only finite diagnosis,not independent regression acceptance.
Ruling:public-view/member references are proxy evidence,not a full binding trace.
Cost:conclusions restricted to these fixture boundaries. Ruling:PowerShell-native
existing isolation/capture,no Bash/install/new worktree or packet deletion.
Cost:no generic skill-runner claim;all historical records retained.
No independently graded/deferred Minor findings introduced.

Next recommend revising the OWNERSHIP REGRESSION DESIGN using the measured
distinction:initialize mutable wrapper inputs BEFORE evaluating;never replace
their field after mutation. For actual raw String outputs,separately inspect
retained views and field rereads rather than demanding unsupported persistence
or calling ineffective attachment a pass. Do not wrap/change candidate outputs
merely to satisfy the observer. Agree explicit semantic expectations and NEW
finite validation scope before any test/candidate revision or regression run.

This diagnostic allowance1/1SPENT;old diagnostic1/1 failure and candidate repair
1/3SPENT+2/3UNUSED_FROZEN remain separate. All older seats/budgets/gates persist.
Git carries necessary diagnostic files/report/status/recovery on the existing
development branch,not ignored raw/runtimes/credentials or renewed authority.
Restore latest remote-confirmed dev checkpoint,not main.
