# Account alignment native isolation and binding execution

## Task2 complete — corrected locked restore and isolated build (2026-11-12)

User explicitly approved preserving failed outputs, removing the global RID argument,
one corrected locked restore and one conditional build, no launch. Source remained
bed155cb6346af6e7707fd5bce167ba1e1b4f277 clean throughout restore/build/inspection.
The earlier two stops below remain historical evidence, not current blockers.

Preserved5 App and5 Core restore files in new ignored plan-local snapshot
.superpowers/sdd/2026-11-12-windows-account-alignment-native-binding/failed-restore-bed155cb/.
Every copied file's SHA256 matched its source before proceeding; no evidence deleted.
Known failed App assets hash matched the recorded baseline. Reused only the approved
owned App obj directory, while candidate bin remained absent until build. Ancestor
reparse checks and ordinary-token/process checks passed; related processes0.
Revalidated19 metadata/lock entries and16 archive/sidecar hashes. Ten old evidence
hashes matched before and after. No global RID override used: explicit evaluation
required App win-x64/candidate TargetPath and Core empty RID/default net10.0 TargetPath.

Executed from clients/windows, identical properties for restore/build:

```powershell
$variantArgs = @('-p:Configuration=Debug', '-p:EnableLayoutDiagnostics=true', '-p:EnableEmailGeometryRecording=true', '-p:EnableCloseStageRecording=false', '-p:CloseStageVerificationMode=', '-p:EmailGeometryBuildVariant=account-center')
dotnet restore .\Talenro.Demo.App\Talenro.Demo.App.csproj --locked-mode --configfile .\.nuget\account-center-offline.config @variantArgs
# Asset inspection passed before this single build:
dotnet build .\Talenro.Demo.App\Talenro.Demo.App.csproj --no-restore @variantArgs
```

Restore exit0, command approximately1.49s (Core107ms/App164ms). Both asset files
verified only the approved local cache/config/source, their correct App-isolated/
Core-default output namespaces and restoreLockedMode=true. Lockfiles unchanged.
Build exit0 in29.23s,0 errors/1 warning NU1905. Same warning in restore: local audit
source supplies no vulnerability data; offline success is not security clearance.
No download/install, dependency version change, audit suppression, App execution,
additional build retry or termination. SDK10.0.403/MSBuild18.9.15+5a2162bea,
runtime10.0.14, VS Community18.11.12328.437 as preflight inventory.

### Measured candidate for the next binding stage

Artifact directory:
clients/windows/Talenro.Demo.App/bin/email-geometry-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/

| File | Bytes | SHA256 |
| --- | ---: | --- |
| Talenro.Demo.App.exe | 312320 | CF7C577C2FF3F4A690B2EFEE054411B59006210440333C5E8EA1838A7062A48F |
| Talenro.Demo.App.dll | 329728 | 6011FA7BB63A45D5FAEB4A44529904FEC0857741A4E810B20B38555655C838C1 |

Inspected bytes with PEReader; did not load/execute App code. Both AMD64. DLL
contains EmailGeometryObserver, EmailGeometryTargets, GeometryEvaluator,
GeometryCaptureCoordinator, GeometryReceiptSession and LayoutOptions; no CloseStage
type names. Three required Layout fixture/string resources present. Parsed deps
target .NETCoreApp,Version=v10.0/win-x64; runtimeconfig net10.0/10.0.14. deps42809bytes,
runtimeconfig359bytes. All586 App generated-file-list entries reside in the selected
App bin/obj namespaces. Streams/readers disposed after inspection.

Core was rebuilt in its normal bin/Debug/net10.0 path. DLL changed from
8A4E020B477FEC9F0EC03CFC8D0235F08ACD4F6523C3969C3DEE7C9977B1C471 to
BF14C9990084D5A4EA60BB0073CE8ADF884FFCBF8ED582149032CC1077B94963;
candidate's copied Core DLL matches the rebuilt default-path output. This is an
observed dependency rebuild, not a claim that all Core outputs stayed unchanged.
Old layout/geometry binaries, App/Core lockfiles and four original run records
(10 hashes total) all unchanged. Final related-process query succeeded with0 matches.

Task2 now complete under the approved revision; Task3 fixed observer/binding and
synthetic regression/final independent review are still pending. No observer was
modified or invoked. Live allowance0, C4 NOT_ACCEPTED: native compilation/static
inspection do not prove visual clipping fixed. Necessary provenance is committed;
binaries/cache/config/failed snapshots stay local and ignored. A fresh clone must
prepare local dependencies/build anew and verify hashes, not assume these binaries.

## Task2 — one restore attempted, NU1004 stop (2026-11-12)

User authorized corrected hash checks and continuation of the unused one-restore/
one-build budget. Source fcd970e69842a780c90c9922302ce564267cf2f1 clean.
Corrected checks passed:19 framework lock entries matched metadata contentHash;
all16 unique archives' computed SHA512 matched their sidecars. Ordinary token,
relevant processes0, both candidate bin/obj absent and no reparse ancestors.
Ten preserved hashes inventoried before the command; Core ordinary DLL SHA256
8A4E020B477FEC9F0EC03CFC8D0235F08ACD4F6523C3969C3DEE7C9977B1C471.

Created ignored .nuget/account-center-offline.config with only the existing local
cache as package/audit source and globalPackagesFolder. No remote package source,
download, installation, lock change or audit suppression. Ran exactly once:

```powershell
$variantArgs = @('-p:Configuration=Debug', '-p:RuntimeIdentifier=win-x64', '-p:EnableLayoutDiagnostics=true', '-p:EnableEmailGeometryRecording=true', '-p:EnableCloseStageRecording=false', '-p:CloseStageVerificationMode=', '-p:EmailGeometryBuildVariant=account-center')
dotnet restore .\Talenro.Demo.App\Talenro.Demo.App.csproj --locked-mode --configfile .\.nuget\account-center-offline.config @variantArgs
```

Exit1 in approximately1.67s. Core NU1004: global RuntimeIdentifier=win-x64 differs
from its lockfile's empty runtime set. App restore completed353ms, Core failed95ms.
App warning NU1905: local audit source provides no vulnerability data. Overall
restore FAILED; no build, retry, App execution or process termination performed.

Root cause is the plan's global RuntimeIdentifier argument propagating into Core,
not a package/hash failure. Preflight evaluation showed the Core win-x64 TargetPath,
but the agent failed to reject that discrepancy before restore. App.csproj already
defines win-x64 locally; Core.csproj has only net10.0 and its lock has no runtime.
Read-only evaluation with the global RID argument omitted confirmed App retains
win-x64/fixed candidate output while Core has empty RID/default net10.0 output.
No corrected restore/build was executed. Keep locked mode; do not force-evaluate.

Postflight: all10 preserved binary/lock/receipt hashes identical; ordinary Core DLL
unchanged, Core bin/Debug/net10.0/win-x64 absent. Related process count0. Source tree
clean before documentation edits. Candidate bin absent; candidate obj now contains
five restore files. Assets list only the approved local config/cache/source, selected
App obj path and restoreLockedMode=true. These are partial failed-run outputs, not
an accepted restore prerequisite. Local assets SHA256:
7CAF198FEEBB2FF9007920C50879831B8037D34B309FD7C6689A1ABA3954683B;
project.nuget.cache SHA256:
6FF30BB49E97B07BC93E642E62995F8E4B862FD21B99D37C0C615ABF3C417C35.

Next proposed revised authorization (not yet granted): preserve all five failed-run
App obj files and relevant Core restore records in a new ignored evidence snapshot,
verify snapshot hashes, then allow reuse only of this known owned candidate obj
(bin must remain absent). Remove only global -p:RuntimeIdentifier=win-x64 from both
restore/build arguments; App's own RID stays fixed. Re-evaluate App/Core RID/TargetPath
and reject any Core RID suffix. Authorize one new cache-only locked restore and,
only after success/asset inspection, the unused one no-restore build. No deletion,
lock update, weakened token/process/reparse checks, dependency download or launch.
Original plan commands must not be repeated unchanged. Task2 incomplete, Task3
blocked, live0 and C4 NOT_ACCEPTED. New computer has docs/source, not these local outputs.

## Task2 — approved, stopped during preflight (2026-11-12)

User approved one cache-only locked restore followed by one conditional native
build/inspection, no launch, stop on failure. Source2d8278a9fc513d573925466cf8a362d89630cc16
was clean. Administrator=False; relevant dotnet/MSBuild/testhost/vstest/App/
XamlCompiler process count0. Both candidate directories absent, existing ancestors
checked without reparse points. Installed SDK10.0.403, MSBuild18.9.15+5a2162bea,
runtime10.0.14, Windows10.0.26200, VS Community18.11.12328.437.

The first cache preflight command stopped with `Cache hash mismatch:
microsoft.windowsappsdk`. This was an agent-authored comparison error: it compared
the .nupkg.sha512 archive hash to the lock's contentHash. These are different
hash domains for signed packages, not evidence of cache corruption.
[NuGet's extractor](https://github.com/NuGet/NuGet.Client/blob/dev/src/NuGet.Core/NuGet.Packaging/PackageExtractor.cs)
writes the archive hash to the sidecar and content hash to .nupkg.metadata;
[metadata specification](https://github.com/NuGet/Home/wiki/Nupkg-Metadata-File)
describes the content hash excluding the signature.

Read-only follow-up: all19 lock entries (including repeated framework entries)
have .nupkg.metadata contentHash matching their lock entry. For WindowsAppSDK2.6.5,
computed SHA512 of the72436-byte archive equals its sidecar:
0CBOgQYP6BOwAfsJap3KnpMVRGBtY633GrsV0tE1bGSabvCal6eJ14bZgAHjPVS7MbiduvWR3EfbOb1AMQaLKQ==.
Metadata/lock contentHash is:
XmxRiyoq82OeitV2FpFKxahDEnnw1aL4j0GGEiF8QQYcVO3/IhSMsjbcoC02AWXnRqIT5JbQ2IbK7YH0LJZARQ==.
This does not establish full archive integrity/signature validation for all packages.

No restore/build command ran; no local config was created. Later property evaluation,
full old-evidence inventory and Core snapshot were not reached. Candidate directories
remain absent; no package/cache/source/lock/binding changes, no launch or termination.
Task2 is incomplete; Task3 remains blocked. Preserve this failed preflight record.
Next proposed action: corrected read-only cache checks (metadata contentHash against
lock; actual archive SHA512 against sidecar), then remaining preflight and the still
unused one-restore/one-build budget, only after user confirms continuation. No
weakened checks, download, lock change or automatic restore/build retry.

## Task1 — fixed App-only variant, offline checkpoint

User confirmed the three-stage plan and Task1 inline execution. Source baseline:
433b935a16b31eea555b7bfafbdb795d10f5464e, clean linked worktree on
codex/c12-b01-task9-coordinator. Task2 still requires separate explicit approval.

Added EmailGeometryBuildVariant=account-center paths to App only, after existing
geometry paths. Empty variant retains existing behavior. Initial validation
ValidateEmailGeometryBuildVariant precedes existing guards and rejects unsupported
values or combinations with EMAILGEOMETRY002. Existing guards remain unchanged.
Core/Tests path comparison is evaluated against each project's empty-variant result.
No App XAML, runtime source, original observer or shared observation module changed.

## Verification on 2026-11-12

All three test commands used the approved external execution context with an
ordinary Windows token (administrator rejected) and no existing test process.
No restore requested. Managed Core/Tests compiled; App was evaluated, not built.
Each command finished within the120s observation budget; no retry or termination.

From clients/windows:

```powershell
dotnet test .\Talenro.Demo.Tests\Talenro.Demo.Tests.csproj --no-restore --filter 'FullyQualifiedName~EmailGeometryIsolationTests' --logger 'console;verbosity=normal'
dotnet test .\Talenro.Demo.Tests\Talenro.Demo.Tests.csproj --no-restore --filter 'FullyQualifiedName~EmailGeometry' --logger 'console;verbosity=normal'
```

- RED before implementation:18 total,10 passed,8 failed, exit1, test runtime12.539s.
  AccountCenterPathsStayInApp still returned bin/email-geometry/. Seven
  InvalidAccountCenterVariantFailsEarly cases lacked EMAILGEOMETRY002; an absent
  target/older guard was not treated as successful new validation. Original8 and
  two Core/Tests comparison cases passed. No compile/infrastructure failure.
- GREEN after implementation:18/18, exit0,13.499s. All seven invalid cases now
  require EMAILGEOMETRY002. Selected paths/TargetPath/defines/Compile items pass.
- Geometry regression:107/107, exit0,17.915s, no failures or skips.
- Post-run related test/dotnet/application process query succeeded:0 matches.
- git diff --check passed (Git only reported expected LF-to-CRLF notices).

Verification is the approved targeted offline scope, not a full-project suite or
rendered-layout acceptance. Final independent whole-plan review remains Task3.

## Preserved evidence

Before/after comparison verified all10 hashes identical: old geometry EXE/DLL,
old layout EXE/DLL, App/Core lockfiles, and all four records of
run55550244cd6449eab7d870733015eb23 (manifest, samples, geometry, claim).
Binary/lock baselines are recorded in
[prior build](2026-11-11-windows-c4-email-geometry-build.md).
Receipt baselines:

| Record | SHA256 |
| --- | --- |
| manifest.json | BE56A8778568679C2A2545147A44F62C9DE51691476F55F31B28C2186212008C |
| samples.jsonl | A60840C9A21A469F9E41A3AC29504652586E684C1939BF0BBA7132C4B81118FE |
| geometry.jsonl | AE5ABE46357E8E5AB2DF100A09843CB6C62BB4E9E8B865D226FE5619AB97DB1A |
| run.claim | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

New App bin/email-geometry-account-center and obj/email-geometry-account-center
were absent before and after testing. No candidate native artifact/hash exists yet.
No cache restore, native build, binding edit, launch, installation or security
change occurred. Raw evidence, caches and generated outputs stay ignored/local.

## Next gate

Task2 requires separate approval for one cache-only locked restore followed,
conditionally, by one no-restore native build and metadata inspection. Follow
the approved plan's absence/reparse/cache/process/token checks and300s limits.
No remote package download, copied assets, automatic retry or launch. Missing
cache or unexpected result stops downstream work. Task3 awaits measured hashes.
Live allowance0; C4 NOT_ACCEPTED. Fresh clone restores code/tests/docs, not local
artifacts, receipts or cache; resume from this development branch, not main.
