# Extent contract rebind — actual bounded result

Status: STOPPED / NOT_ACCEPTED. Design2c58fb47, plan881bdd0a and naming
correction193be974 were approved. Native execution; no additional reviewer,
suite, compilation, install or cleanup. Original spec/plan/addendum unchanged.

## Actual sequence

1. Preparation:31baseline sources,2276inherited pins verified using preserved
   snapshots,2307historical bindings,30new provenance paths; no overlap.
   Four new test/helper files parsed before invocation. Preparation checkpoint
   18c98249b60c6c9b6c796d076d81ae0f02185b9b pushed and remote verified.
2. RED18:0PASS/1FAIL/21NotRun/exit1,Expected=true. Exact RuntimeException
   “Extent contract regression: missing Get-ODEpochManifestName”. SetupError
   and RecordingError null;31source and2307history hashes unchanged pre/post.
3. Changed only driver.ps1,capture.ps1,observation-contract.tests.ps1,
   source-model.ps1 among31operational files. Candidate checkpoint
   4f14deaca80ced283a248cde725d338eb5f7928b pushed/remote verified.
4. GREEN19:22PASS/0FAIL/0NotRun/exit0,Expected=true. All ten nested command
   extent invariants PASS; RootError/SetupError/RecordingError null. Actual
   inputs unchanged. This auxiliary result is not full contract acceptance.
5. One Prepare invocation exited1 in Driver-ExtentContractGate at driver line896.
   It wrote extent-contract.stop.json. Prepare intent, manifest, snapshot tree
   and host intent are all absent. No host was launched. No retry or code repair
   after this unexpected result.

## Root cause from saved error and source

The immutable HistoricalFiles map contains three absolute external identities:
the pinned bundled pwsh.exe and System.Management.Automation.dll, and Windows
PowerShell powershell.exe. Preparation and regression use these paths directly.
The new actual driver gate instead applies Join-Path $repo $rel to every map key.
For a different-drive absolute key this produced a D:\project\C:\Users\... path;
Get-FileHash rejected the invalid filename. The first observed failure is the
pwsh.exe path. This is a newly introduced path-resolution defect, not evidence
that Docker, runtime installation or permissions need changing.

The auxiliary test exercised fact predicates and bindings but did not invoke the
actual gate's historical-file path resolution. Therefore22PASS did not cover
this integration edge. Do not reclassify GREEN19 or the failed Prepare as a
successful full contract run. A future repair needs a real path-resolution
regression covering pinned absolute and repository-relative paths, while retaining
all original identities and checks; no such repair/test was run here.

## Immutable evidence

- Input50A088C0BFE6E931BB5F1DB3194399DEF9F1AD15CC7C949EF4091EF2A93737E2.
- Test9382D7EEADE427FE006A86E43F55316E4C98999B5D62674167E0E662C1595599.
- RED finalFA1601603A3779E780BA7533D65B937F1992307DD91CAE5486B4F0A9BCF83C2D.
- GREEN final9DD04B19DEE76A7056B58D9D0DA71D4F428E42FC89E41C76DE8F326C0D2A3604.
- StopC26E23A5783CC4CC6062F9142000B2D68931D1C990AE30C0E5986800B222029F.
- All three actual tool inputs/results retained under P, including
  extent-contract-prepare.tool-record.json. Auxiliary slots use binding-red/green;
  reserved host slot extent-contract-green remains absent.
- Preparation archive547files/135directories/31zero-byte, SHA
  226681694613AEE6996C2973134A7E8B48FFD75BBA834B4A12CED98211ED12AD.
- Candidate archive551files/135directories/31zero-byte, SHA
  DE8230554CE0BD554C0F955444B1A6F79ECBEAD4CA4CD82889089D5A03743C68.
- Final archive portable-extent-contract-final.json contains all actual P/Q
  originals plus the two required H inputs. See latest status for its verified
  inventory and hash. No freeze archive exists because no freeze completed.

## Budgets and recovery boundary

Cumulative19auxiliary/4hosts/3compiles/186snapshots/0targets/0reviews. Both added
auxiliary runs consumed. The single Prepare call was attempted and failed before
its internal intent/snapshot writes; absence of that intent does not authorize
replay. Added1host/31snapshots remain unconsumed but FROZEN by the new stop. The
legacy downstream3hosts/7targets/conditional1review remain frozen. Nine historical
stops plus the new stop must all be retained. Product NOT_ACCEPTED; no main merge.

Resume from the latest development branch and final archive, never from main
or a consumed phase. Further code changes, tests and a new Prepare route require
separately approved recovery scope; do not delete/bypass the stop or edit old pins.

## Recorded execution decisions

- Used approved finite PowerShell sequence in existing linked worktree instead
  of generic skill extra suites/reviewer/cleanup. Broader acceptance is unproven.
- Candidate references immutable input history instead of duplicating2307hashes;
  the same input hashes were checked pre/post and remain bound by SHA256.
- Passed existing PreviousFrozenSHA256 argument explicitly using approved parent
  identity; no parent change or additional allowance.
- Preserved approved plan/addendum bytes and recorded separate execution status.
