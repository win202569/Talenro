# P1 owned-return candidate — string fixture invalid; STOP UNVERIFIED

Outcome=FIXTURE_INVALID;SuitePassed=false;FormalAcceptance=false.
PromotionApproved=false;I1CandidateDisposition=NotAccepted.
NEW ordered allowance:1/3SPENT,2/3UNUSED_FROZEN. No retry/repair/reviewer/main.

## Approved scope and implementation

The user approved the bounded design recorded in current-status.md at
7bbaf3a683af924bcd7977f3533b2ccd49bc1bc1:NEW isolated candidate,then focused,
full API and existing ownership hosts in order,each30s+2s owned exit+2s drain.
Any failure stops later hosts;no automatic fixes/retries or promotion.

Nine portable files added under docs/diagnostics/isolated-fuzz/scalar-owned-return-candidate/:
LF attributes,README,source pins,expected output,capture controller,new focused
suite,new candidate and two byte-identical existing test routing copies.
Original TokenF9B4E408,integer/original modules,old tests and raw records untouched.

New focused assertions were prepared BEFORE the candidate. Retained old first-
case RED is historical evidence;no fresh baseline/RED/mutant execution of the
new assertions was authorized or performed. Additional assertions have not been
independently sensitivity-proven. No generic whole-project test or reviewer.

CandidateSHA256=1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8.
Added3 internal helpers for fixed ownership metadata,exact primitive reboxing/
nonempty String construction,and fresh13-root/22-Decision result construction.
Both normal and catch/refusal return sites use the owned builder. No serialization,
caller-selected method invocation,input cleansing,global cache or type-data change.
Static AST comparison:13 original functions retained plus3 new;12 original
helper bodies identical. Get-OfflineSupervisionDecision differs only at its two
return expressions;public parameters and validation/reduction text unchanged.
Both copied old test hashes match their sources. All5 new/copied scripts parse.
These are static source facts,NOT successful runtime validation of the entrypoint.

## One actual host and exact saved output

Fixed ordinary Core7.6.5/x64 engine SHA256
362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.
Command:fixed engine -NoProfile -NonInteractive -File owned-return.tests.ps1,
absolute pinned file path in local receipt. No execution-policy override/elevation.
Slot1 consumed BEFORE launch. StartedUTC=2026-10-18T11:25:02.4529676Z.
Elapsed818ms,exit1,no timeout,owned-host-exited,complete streams,empty stderr,
ControllerError=null,PostPinsIntact=true. Parent/child nonadministrator.
The outer controller STOP exception is not child stderr or an additional host.

```text
HOST: 7.6.5 Core x64=True admin=False
PASS: scalar-null
PASS: scalar-int32-zero
PASS: scalar-int32-min
PASS: scalar-int32-max
PASS: scalar-int64-zero
PASS: scalar-int64-min
PASS: scalar-int64-max
PASS: scalar-bool-true
PASS: scalar-bool-false
RESULT: FIXTURE_INVALID case=scalar-string-clean detail=attachment-not-retained:scalar-copy calls=0
```

Nine of15 scalar fixture vectors passed(null,3Int32,3Int64,2Boolean). These
non-null passing vectors checked exact type/value,clean initial leaves,distinct
backing objects,retained metadata/payload identity,mutable payload retention and
absence on input/other copy. Null checked null-return semantics only.
The first String vector reached initial type/value/plain/backing-identity checks,
then could not retrieve the just-added known NoteProperty through the saved
scalar-copy field. It failed the explicit attachment-fidelity assertion before
payload mutation/other-owner isolation checks. Five later string vectors UNRUN.

## Confidence and stopped coverage

The source operation attempted direct attachment through an existing field,
without replacing or rewrapping it. The subsequently retrieved field did not
retain that property. This demonstrates an INVALID mutation fixture for this
String path,not isolation success or a demonstrated new String product defect.
It does not establish the PowerShell engine-internal cause or whether creating
different wrappers would preserve the actual production-boundary semantics.
Do not remove the fidelity guard,replace fields after return or count an
ineffective mutation as a passing ownership test. No post-failure code changes.

The focused child imported the module and exercised ownership helpers only.
Get-OfflineSupervisionDecision calls=0. Both normal controls and ALL39 behavioral
ownership cases were UNRUN. New owned-result builder/reduction/catch/refusal paths
have no runtime evidence from this host. Do not claim historical Int32 aliasing
fixed merely because direct primitive helper checks passed. The old descriptor-
contamination witness remains unchanged in the [prior result](2026-10-18-p1-plain-scalar-ownership-result.md).

Full API and previous ownership hosts were NOT launched;their new candidate
allowances2/3 remain unused/frozen,not available for alternate probes.98-line
focused success inventory did not match(actual11lines);38/22 later inventories
remain unobserved. New candidate is implemented but UNVERIFIED/UNACCEPTED.

## Readonly reconciliation and frozen identities

Parent reread source/assertion order,receipt and original output as DATA only.
Matched all548 frozen pin entries before/after:535 inherited+4 old packet
anchors+9 new files. Counts include repeated paths,not548 distinct files.
The exact11-line failure and both stream hashes were independently reconciled
against the saved receipt. No extra product/helper/import/host execution in the
parent;no reclassification,oracle edits or parser repair after the failure.
This is author reconciliation,not independent review.

Ignored packet:.superpowers/sdd/2026-10-18-p1-owned-return-repair/.

|Artifact|SHA256|
|---|---|
|frozen.json|EA27F7D75EFCAF291A4B59F15B531122B41ED19AACAF6EA1850031010FC613F8|
|invocations.json|5C6820047BBB033AA43AC5F92766CE39E0F7CE5B421850E5F164DEFBF1032783|
|focused.stdout.txt|259CC202BD1004450F0EECCD4FE99EFB3535278D267CB9A135A16A8DFE45C192|
|focused.stderr.txt(empty)|E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|
|offline-supervision-policy.ps1|1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8|
|owned-return.tests.ps1|7DE3A737E177F1B191393AACC9D67479B0FE62A9C82521AD5E602522E1D92D44|
|capture.ps1|99ACE622727A6BB963FCC7F55D47691573B42E66398EF6207EDECA66FAE87288|
|expected-output.json|4DFE985FABAE080D21A3C254D209DF5BAD5AA0A4E89E1D271786CBD149242D58|
|source-pins.json|4D7DC625DA698C6ECD64E77A53368DF2C289BF99F3D79E5D827FBBAD7DE6EB14|
|README.md|3AE81B30B64D27F6749262667EA0F765F64DD208D5B10301EE5A2BE19C3526EC|
|.gitattributes|A79691A93B46E49CE460C26EF22AFCC03D6ECA1E63BF2EDBC20E96159510F6C9|
|full API routing copy|61A3BACF26EDE1F8B137CBF77CC4DBC35E4D6E6FB8FE6365953CDA9F41CEDEA3|
|previous ownership routing copy|C4905925B309B00E94834B2FDA3EE8034265B23CDF274276C9ECA5511250C6E7|

## Rulings and next recommendation

Ruling:explicit approved retained RED plus tests-before-change replaces another
baseline/RED run. Cost:no fresh negative-control proof for additional assertions.
Ruling:approved max3/fail-stop scope overrides generic full-project reruns,repair
cycles or final reviewer. Cost:unfinished failing checkpoint,no acceptance.
Ruling:reuse PowerShell-native isolated workspace/capture,no Bash skill scripts,
installs,new worktree or scratch deletion. Cost:no generic skill-runner claim;
all historical records retained. No independently graded/deferred Minor findings.

Next recommend readonly analysis of String extended-member retention at the
stored-field boundary,then a narrowly justified fixture/candidate adjustment
design with NEW finite authority. Do not lower the attachment/identity oracle,
reuse frozen later slots or change the candidate under this spent attempt.
Any revised design must distinguish ineffective attachment from actual isolation.

Prior ownership characterization1/1,review1/1,direct2/2,full4/4 and all older
failed/frozen/ineligible seats persist. All18+5 declined limits remain,PS5 and
performance unverified,no wider security/runtime/main/C12/release clearance.
Git preserves these necessary portable files/report/status/recovery on the
existing development branch,not ignored raw/controllers' local state,runtimes
or credentials. Restore latest remote-confirmed development commit,not main;
missing local state cannot replenish consumed or frozen validation slots.
