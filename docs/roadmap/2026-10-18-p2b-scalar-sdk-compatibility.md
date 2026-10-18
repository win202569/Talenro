# P2-B scalar SDK compatibility: first ownership precondition failed

## Status and approved boundary

After4f67088e user approved a NEW bounded3ordered targets+1conditional readonly
review: prove scalar/array/fresh-output metadata, reproduce baseline defect,
then repair/focused/full. Existing spent/frozen scopes,including the old final
full slot,remain unchanged. New leaf:
`docs/diagnostics/isolated-fuzz/p2b-scalar-sdk-compatibility/`.

New first target1/3 SPENT;remaining2/3 FROZEN;review0/1 INELIGIBLE. It stopped
on the FIRST literal byte ownership precondition,BEFORE any product API call.
No product repair/seal,baseline defect reproduction,GREEN/full or reviewer.
New candidate is an unchanged B468B564 copy,not a repaired implementation.
M1 stays OPEN. No main merge or acceptance. No retry/correction after stop.

## Test preparation, without product changes

New scalar test combines saved35-vector refusal behavior with structural proof
on nine literal scalar/array types and nonnull leaves of encoded/decoded/binding
results,normal/reused view histories,and182 planned malformed SDK descriptor
cases. Every partial proof and available fixed-field observations are stored
BEFORE independent test-fixture assertions. Output is fixed UTF-8 in the SAME
owned target process. The original full33tests,literal frames,definitions-only
data helper and counter-only C#/PS fixture were copied byte-for-byte.

Expected132 initial proof rows,14history API cases/additional proofs,35matrix
rows and182negative cases describe the intended test inventory,NOT executed
coverage. None of these complete inventories was reached. Count expectations
are not a substitute for structural checks. The unchanged root test helper was
used as a comparison;this run demonstrates that its common-owner precondition
does not simply transfer to this literal scalar context.

Before freezing, the new proof helper was adjusted to return an enclosing
object[] for owner references rather than sending a PSObject itself through
the pipeline. The helper's owner is extracted from that array, preserving the
actual wrapper reference for reflection argument-array comparisons. This was
test preparation before any new packet,target or product edit. All eight new
PowerShell files parsed without errors. No source changes after Prepare.

## Exact saved result

Ordinary PowerShell7.6.5 Core x64,PID29852,1381ms,exit1. No timeout or stderr;
owned exit and concurrent separate raw streams complete. Bounds60000ms
execution+2000ms owned exit+2000ms drain,only its owned target in scope.
Fixture compiled once in memory in this approved target and positive counter
control was observable. Product definitions imported,but ProductCalled=false.

Stage SDKProof:literal.Byte,History0,ValueType System.Byte. One partial proof,
zero Validated proofs,ProofsCompleted=false. OwnerType System.Management.Automation.PSObject.
Available descriptor observations relative to the saved psobject owner:

| Descriptor | Exact class | Hidden / Reserved / Instance | OwnerSame | Additional owner comparison |
| --- | --- | --- | --- | --- |
| psobject | PSInternalMemberSet | true / true / true | true | InternalOwnerSame=true |
| psextended | PSMemberSet | true / true / true | false | ConstructorOwnerSame=false |

The independent DG-ApplicationMembers assertion failed at descriptor-fixture.ps1:74,
which compares the extended-set constructor owner to the psobject shim owner.
Error is RuntimeException / Descriptor assertion;stack includes SC-ProveValue
scalar-sdk-fixture.ps1:35 and SC-InitialProof:52. The two saved owner-equality
verdicts are false;that is a real observed mismatch with the proposed common-
owner invariant,not proof that a malicious descriptor was present.

The partial proof's SDKCount,ApplicationCount and Counters are null,NOT measured
zero. No history API cases,matrix rows or negative cases occurred. No fresh-
output proof or array/type expansion occurred. The baseline HiddenDecorationRefusal
RED gate was NOT reached,even though process exit1 was captured. Controller
Outcome INCOMPLETE,ExactExpected=false is correct and remains unchanged.

## Interpretation and limits

The previous empty-raw assumption was wrong;this attempt adds a distinct,
now observed limitation: root-style equality to one PSObject wrapper owner is
not established for a normal controlled scalar. This should not be resolved by
ignoring owners,comparing only primitive values,allowing a count/name list,or
declaring the partial proof valid. Types/flags alone are not enough either.

The capture does not preserve the actual PSObject references or their base-
object/instance-collection relationships for later comparison. It records only
type/name/flags and selected equality verdicts. Both psextended comparisons
are against the common owner;they do NOT prove whether psextended's own instance
and constructor owners equal each other. Exact boxing/wrapper/metadata-sharing
mechanism is NOT established here. Do not invent a shared-base identity proof
or claim the new scalar structure has been accepted. No SDK factories/getters
were called by validation to manufacture a supposed canonical identity.

Before another repair, reconsider the scalar wrapper-identity contract using
read-only SDK source/metadata analysis. Distinguish actual PSObject-wrapper
identity from boxed-base identity and metadata collection ownership,and decide
what structural guarantees are achievable. No new probe,relaxed rule,candidate,
implementation or execution allowance is approved by this recommendation.
The remaining2slots are frozen,not reusable diagnostic slots. No reviewer
dispatch or extra fix under this stopped scope. Structural conventions cannot
be promoted to isolation against arbitrary same-process private reflection.

## Integrity and recovery

1078 mapped pins intact after target and subsequent read-only rehash:
1056 inherited+6priorpacket files+2priorreports+failedcandidate+12immutable
new inputs+baseline snapshot. Engine and SDK separately pinned as before.
Old failed18DA24E1 and historicalB468B564 remain unchanged. New copied module:
B468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039.

Ignored packet: `.superpowers/sdd/2026-10-18-p2b-scalar-sdk-compatibility/`.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 204E6E0652F8D12E97E51CAD23B6688B9392F15A69D0D5608B3BFE7A465E700E |
| invocations.json | FB1405516977527148C73A738187D1311F53A88AF3EF36B3EEFF035F63686C72 |
| red.stdout.bin | 428AA1514E945FBD73D5FA6F846FF9E7875277E848F223C1FDC5D6FE675DB790 |
| red.stderr.bin | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

New source hashes: scalar.tests.ps1 9D4C26C4136218B9395E19FE270AF42E1FBE7C7E3CC34F06B8DFD41911173CDC;
scalar-sdk-fixture.ps1 6A503446087500F5E55D704EA50FD85526D081B699CC4BC369EBD31B1E91ED71;
capture.ps1 32246ADF8378172845151AC65BF73338D9E8E2A1286B05A4B8F558A7A20E26CE.

Recover latest codex/c12-b01-task9-coordinator,not main. Portable source/tests,
this partial-observation table,status and recovery docs are preserved;clone
does not restore ignored raw,SDK/runtime,caches,mappings/sessions/conversations
or reset/replay rights. Do not adopt failed18DA24E1 or describe unchangedB468B564
as closing M1. New1/3spent+2frozen+review0/1ineligible;old2/3spent+1frozen+
review0/1ineligible and all earlier scopes unchanged. No formal P2-B/PS5/P2-C/
P3/full C12/live/runtime/security grant;separate P2-A scope unchanged.
