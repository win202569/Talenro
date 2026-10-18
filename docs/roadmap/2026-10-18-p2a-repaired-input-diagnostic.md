# Repaired P2-A InputMutation diagnostic — Plan container misclassified

## Approved scope and actual outcome

User approved ONE new repaired-version diagnostic after the readonly review of
6f984cab31e569a93b26211b5c912872d14fc307. Investigative probe only: no product
changes, fixes, retries, mutants, reviewers, wider acceptance or main merge.

**REPRODUCED_SITE_CAPTURED; new allowance1/1spent; STOP.**
Original repair remains3/6spent,unused3frozen,review0/1ineligible.
No automatic replay on this or another computer.

One ordinary PS7.6.5 Core x64 target,admin=False,exit1,1389ms,22stdout lines,
stderr0,no timeout,owned-host-exited,complete concurrent separate raw-byte drain.
ControllerError=null;829mapped checks intact after capture and readonly audit.
Max30s execution+2s owned exit+2s drain;durable consumption before launch.
No extra target/import/probe after the single diagnostic.

## Exact finding and its limits

The earlier18PASS prefix through RangeRelationships reproduced. The current
failure is NOT the old input-Schema failure. Diagnostic field=Plan,
stage=leaf-initially-plain,assertion ordinal137.

The current repaired full.tests.ps1 operation is line53:
`foreach($p in $members){TB-Need ($p.MemberType -ne ...::NoteProperty)}`.
It was reached via TB-Leaves line59 from the output-leaf enumeration at158.
The diagnostic stack is TB-Need64 -> TB-Leaf110 -> TB-Leaves117 ->
InputMutation body242 -> TB-Case69 -> top-level232.

Input leaves had finished enumeration. The first output traversal attempted
TB-Leaf on the Plan container instead of recursively enumerating its fields.
Plan correctly has the eight NoteProperties OriginTick,Frequency,DeadlineTick,
StopAtTick,NaturalUntilTick,CleanupUntilTick,PublishUntilTick,
ObservedDeadlineExpired. Applying an initially-plain SCALAR precondition to this
container triggers the recorded assertion. Input mutation loop/root mutation/
saved-result comparison and later groups were NOT reached.

Failure-only state capture succeeded (CaptureError=null). The observed output
root reports Schema Int32=1,Kind=timebase-plan-only,Created Boolean=true,
Reason=plan-created,Plan underlying PSCustomObject,and three authority flags
Boolean=false. This records one output at the failing traversal,not both results
or all numerical Plan values. The failure-only inspection does not create a
before/after ownership comparison or global product clearance.

Static contrast: TB-Type passes `$Value` as the fixed reflection receiver;
the focused helper F-Type explicitly passes `$Value.PSObject.BaseObject`.
TB-Leaves relies on TB-Type matching PSCustomObject to choose recursion. The
captured path proves that branch did not recognize this Plan container; the
receiver/ETS wrapping difference is a concrete candidate explanation. The
actual return from TB-Type was not separately captured. No isolated corrective
experiment was authorized,so do not claim this candidate correction verified.

Next recommendation: a bounded TEST-HELPER-only correction of TB-Type receiver
normalization in a new isolated copy,with unchanged product,fixture helper,
assertion expressions,literals and ordering. Obtain a separate finite
verification allowance before implementing/testing it;do not weaken container/
ownership requirements or reuse the frozen3slots. This diagnostic is not RED/
GREEN regression or full acceptance.

## Transformation and verification

Seven portable files under
docs/diagnostics/isolated-fuzz/p2a-repaired-input-diagnostic/.
Exact repaired helper C720E247 and canonical product350CAB1A copied unchanged.
Reused the prior diagnostic's stage/counter instrumentation against the new
helper,with operation lines mapped to repaired full.tests.ps1 (old>=38 minus2).
All original assertion expressions,literal inputs,ordered prefix and product
calls through InputMutation preserved. Later groups omitted;all outcomes stop.

Before launch,independently reconstructed the allowed transformation from the
repaired full suite using the previous allowlisted changes and verified exact
text equality after removing the additional failure-only inspection helper.
PowerShell parser checks on all scripts,JSON parse,physical LF and exact helper/
product hashes passed. No target was run for these static checks.

New failure-only inspection reads known passive NoteProperties,scalar types/
values and attached NoteProperty names;it skips ScriptProperties. It runs only
after a boolean assertion fails,before the assertion throw. State capture
errors remain explicit. Instrumentation can affect observation,so the finding
does not retroactively prove the exact failing subassertion of the prior run.

Brainstorming spike path used the approved in-chat probe,no new spec/plan.
Systematic debugging prioritizes the observed current site over the old trace.
Verification skill requires reporting diagnostic reproduction,not a fix.
PowerShell-native capture replaces Bash helpers under the Windows-only scope.
No generic extra test/fix/review loop was added.

## Portable diagnostic record

This transcription preserves the captured fields and stack;it is not another
run. Ordered stdout consisted of the same context/host and18PASS prefix as the
[previous report](2026-10-18-p2a-fixture-isolation-repair.md),then this DIAG record,
then `ASSERTION FAILED: P2AInputMutation`.

```json
{
  "Schema": 1,
  "Outcome": "INPUT_MUTATION_EXCEPTION",
  "Case": "InputMutation",
  "Stage": "leaf-initially-plain",
  "Field": "Plan",
  "OtherField": "",
  "OtherIndex": 0,
  "AssertionOrdinal": 137,
  "OriginalOperationLine": 53,
  "ExceptionType": "System.Management.Automation.RuntimeException",
  "ErrorId": "Assertion",
  "Message": "Assertion",
  "DiagnosticLine": 64,
  "Stack": "at TB-Need, D:\\Projects\\Talenro\\.worktrees\\c12-b01-task9-coordinator\\docs\\diagnostics\\isolated-fuzz\\p2a-repaired-input-diagnostic\\diagnostic.tests.ps1: line 64\r\nat TB-Leaf, D:\\Projects\\Talenro\\.worktrees\\c12-b01-task9-coordinator\\docs\\diagnostics\\isolated-fuzz\\p2a-repaired-input-diagnostic\\diagnostic.tests.ps1: line 110\r\nat TB-Leaves, D:\\Projects\\Talenro\\.worktrees\\c12-b01-task9-coordinator\\docs\\diagnostics\\isolated-fuzz\\p2a-repaired-input-diagnostic\\diagnostic.tests.ps1: line 117\r\nat <ScriptBlock>, D:\\Projects\\Talenro\\.worktrees\\c12-b01-task9-coordinator\\docs\\diagnostics\\isolated-fuzz\\p2a-repaired-input-diagnostic\\diagnostic.tests.ps1: line 242\r\nat TB-Case, D:\\Projects\\Talenro\\.worktrees\\c12-b01-task9-coordinator\\docs\\diagnostics\\isolated-fuzz\\p2a-repaired-input-diagnostic\\diagnostic.tests.ps1: line 69\r\nat <ScriptBlock>, D:\\Projects\\Talenro\\.worktrees\\c12-b01-task9-coordinator\\docs\\diagnostics\\isolated-fuzz\\p2a-repaired-input-diagnostic\\diagnostic.tests.ps1: line 232",
  "State": {
    "Frames": [
      {
        "Scope": 2,
        "Variable": "Owner",
        "Properties": [
          {
            "Field": "Schema",
            "Type": "System.Int32",
            "Scalar": 1,
            "NoteProperties": []
          },
          {
            "Field": "Kind",
            "Type": "System.String",
            "Scalar": "timebase-plan-only",
            "NoteProperties": []
          },
          {
            "Field": "Created",
            "Type": "System.Boolean",
            "Scalar": true,
            "NoteProperties": []
          },
          {
            "Field": "Reason",
            "Type": "System.String",
            "Scalar": "plan-created",
            "NoteProperties": []
          },
          {
            "Field": "Plan",
            "Type": "System.Management.Automation.PSCustomObject",
            "Scalar": null,
            "NoteProperties": [
              "OriginTick",
              "Frequency",
              "DeadlineTick",
              "StopAtTick",
              "NaturalUntilTick",
              "CleanupUntilTick",
              "PublishUntilTick",
              "ObservedDeadlineExpired"
            ]
          },
          {
            "Field": "SourceVerified",
            "Type": "System.Boolean",
            "Scalar": false,
            "NoteProperties": []
          },
          {
            "Field": "LiveActionsPerformed",
            "Type": "System.Boolean",
            "Scalar": false,
            "NoteProperties": []
          },
          {
            "Field": "RuntimeApproved",
            "Type": "System.Boolean",
            "Scalar": false,
            "NoteProperties": []
          }
        ]
      },
      {
        "Scope": 2,
        "Variable": "Field",
        "Text": "Plan"
      }
    ],
    "CaptureError": null
  },
  "ProductChanged": false,
  "RepairPerformed": false,
  "Acceptance": false
}
```

## Identities and recovery

| Portable file | SHA256 |
| --- | --- |
| .gitattributes | E3083818059D2956756D32B1D4A7098E28F57DFB3B971F429A73E51672505B9D |
| capture.ps1 | CB0C1619DFCB78E9649301E42D809E1B69B0D8C6E51479075F79501E7E2DA786 |
| diagnostic.tests.ps1 | E9588CAFB5FAAC13DFB98EA02DD8E2EFC832E4C537EC7C5972D47CFE16889B12 |
| fixture-builder.ps1 | C720E247298FA84B8BD348A4A3FBDB3759B7F1F2A510EA0D67D9479E00045A58 |
| offline-timebase-plan.ps1 | 350CAB1A36CECEA57D8335581ADC780CE1F17BDC9379D63F03130C6EEF4FC35B |
| README.md | EDF70B1AC14BB76011C9406654B303DD5E2F038404D4C77DECC5B403C89E4A07 |
| source-pins.json | 625D083459DFA4B1C76B7E63C642C7EF87EF2A830A6C1055E98FC6C05F02F10F |

Ignored packet:.superpowers/sdd/2026-10-18-p2a-repaired-input-diagnostic/.
811 inherited mapped pins+10 prior-packet anchors+original product+7 files=829.

| Local receipt | Bytes | SHA256 |
| --- | ---: | --- |
| frozen.json | 360771 | DFD29E83D47E632376BB794793E475ADF7061587C76A0F895E6234C821262A71 |
| invocation.json | 4819 | D56482B283A69ACBA16B9BDC636F485E13A60330AA1F5E9311796BC2B67B0F04 |
| stdout.bin | 2975 | BDE3183AFA316ECCC9072306B9B42A3E6094F4FF2EDA06C79A5A4E5A15B15FDE |
| stderr.bin | 0 | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

EngineSHA256=362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.
All prior files/raw packets retained unchanged. Necessary source/report/status/
recovery saved on the existing development branch under standing preservation
authorization. Fresh clones recover source/diagnostic record/stop state,NOT raw
ignored receipts,historical local mappings,engine caches or sessions. No secrets,
cache,binary or raw packet added to Git. No merge/acceptance/live-runtime claim.
