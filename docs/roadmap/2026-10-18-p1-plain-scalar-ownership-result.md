# P1 initially plain scalar ownership — first case observed a gap; STOP

Outcome=OWNERSHIP_GAP_OBSERVED;SuitePassed=false;NewHosts=1/1SPENT.
FormalAcceptance=false;PromotionApproved=false;I1CandidateDisposition=NotAccepted.
R1Int32InputToSavedResult=DescriptorContaminationObserved.
No candidate repair, original replacement, reviewer, retry or main merge.

## Authority and actual execution

After ae21a085d233f5fef5d335dfaaa5a3b65419e4f1, the user approved preparation
and ONE bounded attempt from the [proposal](2026-10-18-p1-plain-scalar-ownership-proposal.md).
Five portable files were prepared under
docs/diagnostics/isolated-fuzz/plain-scalar-ownership: LF attributes, test,
independent41-line success inventory, capture controller and README.
No target import/helper/probe or warmup ran before the one fresh host.
Static AST parsing/JSON checks and historical pin checks were preparation only.

Subject remains the isolated Token candidate, SHA256
F9B4E408D9A4FF9C7A36EFD6E1F2472F17DACB3F76F463FF4973DAB4139CC82A.
No changes to it, the original, integer baseline or existing test routing copies.
Parent and child were ordinary/nonadministrator. Fixed Core7.6.5/x64 engine
SHA256 362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.
Command: fixed engine -NoProfile -NonInteractive -File ownership.tests.ps1
(absolute pinned file path in the local receipt). No execution-policy override.

StartedUTC=2026-10-18T11:16:30.3884451Z. Elapsed1162ms, exit1, TimedOut=false,
owned-host-exited, StreamsComplete=true, stderr0bytes, ControllerError=null.
The controller intentionally returned failure after recording the child failure;
its outer STOP exception is not child stderr or a second target execution.
One slot was saved as consumed BEFORE launch. No timeout/cleanup relaxation.

## Exact saved child stdout

```text
HOST: 7.6.5 Core x64=True admin=False
PASS: control-before
BEGIN: normal-exit-int32-input
RESULT: OWNERSHIP_GAP_OBSERVED case=normal-exit-int32-input detail=cross-owner:A.WouldExitCode:before-mutation calls=3
```

The normal control used one product call. The first matrix case made two more
calls on initially plain input, obtained evaluated/Persisted results A and B,
matched both entire13-field results/22-field Decisions to independent literals,
including exact Int32 exit code, and checked distinct result/Decision containers.
All tested scalar leaves were initially plain. These assertions occur before
the attachment path and the recorded failure; no successful-case PASS was emitted.

The fixture installed a NoteProperty named P1PlainOwnershipMetadata holding a
fresh mutable object[] through the EXISTING input progress.ExitCode scalar field,
without field replacement, wrapping, casting or reconstructing a returned value.
A separate fresh same-type/value fixture-control field retained its own descriptor
and array identity. The selected input field was retrieved again and retained
the known NoteProperty, payload ReferenceEquals, Int32 type and numeric0 value.
Then the check of other saved owners found that property on A.Decision.WouldExitCode.
It was absent before attachment. This is the first failing assertion.

## What this proves, and what it does not

On this exact PS7/source snapshot, a later-added extended property became visible
through a previously saved result after attachment through a caller input scalar.
This is now a dynamic descriptor-contamination witness for one ordinary sequential
ownership vector, not merely the prior static evidence gap. It is not an exploit,
authority-escalation, concurrent-mutation or hostile-runspace demonstration.
It is consistent with the shallow value copy at candidate173 and direct
WouldExitCode assignment at216 identified by [R1](2026-10-18-p1-token-readonly-review.md).
Fresh enclosing objects alone did not isolate the returned scalar leaf here.
No engine-internal mechanism or universal PowerShell behavior is inferred.

Fail-stop occurred BEFORE descriptor.Value[0]='changed'. The other owner's
descriptor presence was checked; its payload identity/value was NOT subsequently
read or mutated. Thus cross-owner array-content propagation, result B inspection
after attachment, fixture-control mutation retention and post-mutation checks
were not reached. Do not describe them as dynamically verified.
The initial selected/control descriptor, payload identity and unchanged scalar
checks did run; the remaining mutation-control checks did not. No fixture error
was reported, but this is not a complete fixture-validation pass or an independent
review of the newly written test/controller.

Matrix: first case reached and failed, zero complete matrix-case passes;remaining
17 cases UNRUN/FROZEN. Later Int32 directions, Int64, Boolean, normal/sticky
classification, cross-case checks and final normal control were NOT reached.
Actual calls3/max38. No evidence supports completing all18 cases or generalizing
to all scalar types. The41-line full-success inventory did not match, as required.

## Readonly reconciliation and immutable evidence identities

Parent reread frozen source/assertion order, the four saved stdout lines and the
receipt as DATA, without importing test/product functions or starting another host.
Matched all535 frozen pin entries before/after, including51 scoped+476 inherited
entries (not a claim of535 unique paths), anchors, new portable files and proposal.
The exact four-line failure inventory and both stream hashes matched separately.
No parser correction, receipt rewriting, oracle edits or silent classification
upgrade was needed. This reconciliation is author verification, not a new review.

Ignored local packet: .superpowers/sdd/2026-10-18-p1-plain-scalar-ownership/.

|Artifact|SHA256|
|---|---|
|frozen.json|9CD8A8459C5566EE9B845784EDA56C1D9CB7666BE1CDEDD8189C09DF479D5BAD|
|invocation.json|4E0D5F420CD988A7B183539ACF4EB7EAD6E21CCB486602E50BDA2F709D50918C|
|stdout.txt|C543A1F1FB404EC4C4EC99346FB34C36F15FECBBEE3146F3FDEDB654F4ED04C5|
|stderr.txt(empty)|E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|
|ownership.tests.ps1|7A3A842D63AC6B7D18BF6E5C49DC47E543703CF568FCC9B8AE99CAD504AB8544|
|capture.ps1|CB6508101CB776EB597BD5566D666F84599E556FD691770BE0F9057CBC9CC7A7|
|expected-output.json|6461C1F04C749DAF19F8E1FE0DA7C8FF160B2B57F46E121FBCD8A8D6C6D40A74|
|README.md|1996500F183463F583DBE820BE63C1D3A423E754B5D7DF5B7006669D40E73038|
|.gitattributes|A79691A93B46E49CE460C26EF22AFCC03D6ECA1E63BF2EDBC20E96159510F6C9|

Previous independent review remains historical0Critical/1Important/0Minor and
review1/1SPENT; its InsufficientEvidence verdict is not rewritten as a reviewer
runtime observation. This new author-run Int32 observation is additional evidence,
leaving I1 unclosed and the candidate unaccepted. Previous direct2/2/full4/4
successes do not answer this later-attachment failure; their hashes/outcomes and
all older failed/frozen/ineligible seats remain unchanged. All18+5 declined limits
remain as reported, with this one sequential vector now observed in the stated way.

## Rulings, stopped handoff and next recommendation

Ruling: the explicit one-host/no-retry/no-repair/no-review scope overrides generic
skill baseline/full-suite/repair-RED-GREEN/final-review/scratch-deletion defaults.
Cost: a failing unfinished checkpoint, not a fixed or independently accepted
implementation. Ruling: PowerShell-native existing capture pattern and retained
packets replace generic Bash skill runners; cost: no skill-runner completion claim.
No deferred Minor findings were introduced or independently reviewed this turn.

Next recommend a narrowly scoped ownership-repair design and finite regression
budget, using this retained failing vector. Specify independent scalar ownership
for accepted outputs without weakening admission/refusal guards; choose the
mechanism only after design/source reasoning. No correction or validation retry
is authorized by this completed attempt. Do not simply cast/reconstruct values
and assert that PowerShell wrappers are isolated.

Necessary test/controller/oracle/report/status/recovery files belong on the
existing codex/c12-b01-task9-coordinator development branch, not main. Git carries
portable sources and this exact result summary, not raw packets/engines/credentials
or authority to rerun. Restore the latest remote-confirmed checkpoint on another
computer; missing packets cannot replenish the spent1/1 allowance.
