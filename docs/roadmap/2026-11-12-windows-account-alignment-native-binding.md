# Account alignment native isolation and binding execution

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
