# String metadata retention — readonly source analysis, not runtime confirmation

Status=SOURCE_SUPPORTED_WRAPPER_IDENTITY_HYPOTHESIS.
NewTargetHosts=0;NewProductOrHelperInvocations=0;CandidateEdits=0.
FormalAcceptance=false;PromotionApproved=false;I1CandidateDisposition=NotAccepted.

After e5e1a95654dea7c60ea622f11037b69469941c7b,the user requested the recommended
readonly analysis of the [String fixture stop](2026-10-18-p1-owned-return-repair-result.md).
Read local source/saved evidence and upstream PowerShell7.6.5 source only.
No import,helper call,probe,test,compiler,process experiment or candidate change.

## Local dataflow and established facts

Candidate Copy-SupervisionPolicyScalar134-158 constructs a nonempty CLR String
using fixed String constructor metadata,checks exact type/value and a distinct
backing object,and returns it. It does not deliberately persist a particular
explicit PSObject wrapper as part of a caller field's storage contract.
The focused fixture134-135 stores two helper returns as NoteProperty values.
These initial String type/value/plain/backing checks completed before the failure.

P1O-Attach151-154 reads Owner.Field,accesses its PSObject.Properties and adds a
known NoteProperty. It neither assigns that view back to the field nor retains
the exact wrapping PSObject used during addition. P1O-CheckAttachment155-160
reads Owner.Field AGAIN into a variable,then retrieves its PSObject.Properties.
That second traversal checks base type/value and descriptor/payload retention,
but the saved failure contains no identity comparison of the two wrappers.

Saved terminal line:FIXTURE_INVALID case=scalar-string-clean
detail=attachment-not-retained:scalar-copy calls=0. The descriptor was unavailable
on the second traversal. No metadata-payload mutation or cross-owner String
assertion was reached. The9 preceding null/integer/Boolean vectors passed.
The decision entrypoint,owned-result builder and39 behavioral cases never ran.
These outcomes remain exactly as recorded;this analysis does not upgrade them.

## Relevant upstream source, and its limit

PowerShell7.6.5 MshObject.cs stores instance members through a resurrection table.
GetKeyForResurrectionTables968-986 chooses the innermost PSObject wrapper for
String;ordinary boxed primitive bases use their base object unless local-storage
mode applies. AsPSObject933-958 retains an existing PSObject but creates a wrapper
for a raw object. Thus sharing a String base alone does not guarantee shared
extended-member identity across independently created wrappers.
[MshObject.cs](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshObject.cs#L933).

MshMemberInfo.cs GenerateMemberSet4058-4087 obtains a PSObject using AsPSObject
for reserved views. Its PSNoteProperty constructor1189-1197 and Value1227-1239
store/return the supplied object;they do not promise to create a persistent
String wrapper. Copy1205-1209 also retains the same stored value,not a deep copy
of mutable metadata. This is not a justification to replace ownership with
NoteProperty.Copy.
[MshMemberInfo.cs](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshMemberInfo.cs#L1189),
[reserved view construction](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshMemberInfo.cs#L4058).

Binders.cs4446-4475 routes the reserved PSObject view through that generator.
The reserved member view is an access operation,not a field assignment that
persists the newly obtained wrapper.
[Binders.cs](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/runtime/Binding/Binders.cs#L4446).

These are upstream tagged-source observations,not proof that the installed binary
is byte-for-byte built from that tag or a recorded runtime trace of this failure.
Source reading used network text only;no upstream source was installed or executed.

## Best-supported explanation and unresolved details

Inference:the fixture assumes a stable instance-member identity while repeatedly
obtaining views over a stored String. Wrapper-specific keys make that assumption
unsafe. A property may be attached to a temporary String wrapper and disappear
from a later independently obtained wrapper. This explains why distinct String
backing storage can coexist with failed attachment retention,and is consistent
with the primitive/String difference in the saved run.

NOT established:the actual stored value's raw CLR wrapper identity at each
boundary,which exact binding operation changed the wrapper,or whether a proposed
explicit-wrapper storage strategy would survive construction,helper return,
field storage,retrieval and parameter binding. Old traces did not record these.
No dynamic proof of this mechanism or a successful replacement fixture exists.
The actual wrapper/debasing path remains a hypothesis,not a fully reproduced root
cause. Do not label the candidate's String copy broken,or declare it isolated.

The older P1I1-WrapPrimitive/Decorate helpers explicitly construct a PSObject,
decorate it,then assign the decorated wrapper to Owner.Field BEFORE evaluation.
They check a different storage/mutation path. Their saved successes cannot validate
post-return attachment to an already stored String. Reusing that replacement
operation after return could conceal the very ownership relation being tested.
No old helper was invoked for this comparison.

## Next bounded question, not an implementation instruction

First establish which public field-access path retains the same wrapper and a
known NoteProperty/payload,comparing raw String storage with explicitly retained
wrapper storage established BEFORE mutation. Record both value and wrapper
identity across construction,storage,attachment and reread. Do not reassign or
reconstruct the owner field after attachment,or remove the fidelity guard.
This should be a separately approved fixture-only characterization before any
candidate revision,not a full regression retry or an automatic wrapping change
merely to make a test pass. Exact probe design/finite new allowance remain pending.

If raw String storage cannot support persistent attachment,report that capability
as such;do not silently count it as cross-result isolation or a product failure.
Any later regression design must distinguish that case from caller-supplied
retained wrappers with actually mutable extended metadata. Preserve old Int32
descriptor-contamination evidence independently;this String analysis does not
explain it away or close I1.

## Integrity and recovery

All548 frozen entries matched before/after;frozen manifestEA27F7D7 and unchanged
invocations5C682004/stdout259CC202/empty-stderrE3B0C442 hashes are detailed in the
prior result. Static source/evidence reads only;no failed oracle or receipt edits.
NEW repair allowance stays1/3SPENT,2/3UNUSED_FROZEN;all older seats/budgets held.
No source/test/controller changes,new reviewer,promotion,main/security/runtime/
C12/release acceptance. Candidate1971F808 remains UNVERIFIED/UNACCEPTED.
Only this analysis/status/recovery documentation is saved to the existing dev
branch. Restore its latest remote-confirmed checkpoint,not main;raw/runtimes and
permission to run further experiments do not clone or replenish frozen slots.
