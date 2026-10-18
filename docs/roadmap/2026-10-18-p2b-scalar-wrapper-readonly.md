# P2-B scalar wrapper ownership: read-only SDK explanation

Base d65637131a65373f9f54c2cafd6806c16de2e639. User approved read-only SDK
source/metadata comparison after the first scalar owner precondition failed.
No product/test/controller/fixture edit,import,execution or compile;no new
target host,reviewer,packet,install or security change. Only existing saved
data/local source and official version-tagged source were inspected. This
report/status/recovery are the only changes,under progress-save authorization.

## SDK model, supported by source

PowerShell7.6.5 can create a new PSObject around an unwrapped object. Its lazy
instance-member collection uses a weak-table key. Actual key selection returns
the innermost wrapper for PSCustomObject,string,or local-storage mode;otherwise
it returns the wrapped base object. Hence distinct wrappers around the SAME
boxed number or array can share metadata. This is reference identity,not equal
numeric/string contents. Follow the actual branch,not the older comment that
also mentions Type. [Wrapping/key selection](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshObject.cs#L933-L986),
[collection lookup](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshObject.cs#L606-L623).
[Microsoft weak-table identity semantics](https://learn.microsoft.com/en-us/dotnet/api/system.runtime.compilerservices.conditionalweaktable-2?view=net-10.0#remarks)
specifies reference-based keys rather than overridable value equality.

Reserved factories retain an existing same-name descriptor. Newly created
descriptors record their creating owner;an extended set uses that owner's
instance collection. Consequently a shared collection can retain reserved
descriptors created by different wrappers. [Reserved creation/reuse](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshMemberInfo.cs#L4058-L4105),
[extended-set collection](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/MshMemberInfo.cs#L2939-L2954).

Illustrative source-consistent history,NOT a replay or recovered past trace:

| Step | Assumed relationship | Consequence |
| --- | --- | --- |
| 1 | Wrapper A wraps base B and accesses metadata table C | One reserved view retains A |
| 2 | Different wrapper D wraps the same B and reaches C | Another view can retain D |
| 3 | A and D differ,while B and C are shared | A single common-wrapper assertion can fail |

This model is a reason to investigate key/collection relations,not permission
to accept arbitrary foreign owners. No SDK generator/getter was executed here.

## What the saved byte observation actually establishes

The stopped target's one partial literal.Byte proof records:
psobject exact PSInternalMemberSet,all three flags true,OwnerSame=true and
InternalOwnerSame=true; psextended exact PSMemberSet,all three flags true,
OwnerSame=false and ConstructorOwnerSame=false. The independent comparison
failed at descriptor-fixture.ps1:74 before any product API. These facts remain
unchanged;the partial proof is NOT promoted to Validated.

Inference: the SDK sharing model explains how different wrapper owners could
arise without a hostile descriptor. It is compatible with the observed failure,
but does NOT prove that the actual historical owners shared a base/key/table.
That run saved equality verdicts,not retained CLR references or a complete
owner/base/key/collection relationship matrix. It also did not compare the
extended descriptor's own instance owner with its own constructor owner.

The test setup crosses several function/value boundaries while constructing
and reading independently boxed scalars (data-helpers.ps1:14-23 and
scalar-sdk-fixture.ps1:23-35). Source reading alone does not identify which
boundary created/reused each historical wrapper. No exact call-history claim.
No proof for string/array/other numeric types or fresh outputs was reached.
The previous copied-value inner error remains masked in its historical capture.

## Contract implications, not an implemented relaxation

Wrapper identity,boxed/reference-base identity,and metadata collection identity
must be treated as separate questions. A count,name list,Reserved flag,matching
primitive value,or generic replacement of every owner check with base equality
would discard important distinctions. Root/custom-object and string behavior
must not be inferred from the byte case. Preserve exact descriptor structure,
receiver checks,fixed-field/code-reference checks and no metadata-getter
execution. Structural compatibility is not unforgeable factory provenance or
protection against arbitrary same-process private reflection.

Next decision should first define the required scalar ownership contract and
what evidence would distinguish legitimate shared metadata from a foreign
owner,same-valued different box,or another metadata table. Only a separately
approved bounded evidence capture could establish those missing reference
relations for controlled values. Do not start another product patch or another
three-stage repair based only on this static explanation. Do not reuse frozen
slots,promote old receipts,or silently replace the common-owner assertion.
No new design/candidate/probe or execution allowance is granted by this report.

## Integrity and recovery

All1078 inherited mapped pins rehashed intact. Latest raw stdout remains
428AA1514E945FBD73D5FA6F846FF9E7875277E848F223C1FDC5D6FE675DB790;
receipt FB1405516977527148C73A738187D1311F53A88AF3EF36B3EEFF035F63686C72,
Outcome INCOMPLETE. New candidate and historical structural module remain
B468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039;
earlier failed18DA24E1 is unchanged and must not be adopted. No local SDK method
or IL was executed during research;version-tagged source supports the model,
not a new dynamic verification of the local binary or old instance state.

Latest scope1/3spent+2FROZEN,review0/1INELIGIBLE;previous scope2/3spent+1FROZEN
and all older budgets unchanged. No new target/reviewer this turn. M1 OPEN,
product NOT fixed;historical33PASS is not new acceptance. Recover latest
codex/c12-b01-task9-coordinator,not main. Portable findings/source are restored
by clone,not ignored raw/SDK/runtime/cache/mappings/sessions or reset/replay
rights. No formal P2-B/PS5/P2-C/P3/full C12/live/runtime/security grant.
