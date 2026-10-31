[CmdletBinding(PositionalBinding=$false)]
param()
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'

function Convert-ODInheritedPin($Pin){
    if($null -eq $Pin){throw 'OD: required pin identity invalid'}
    $path=$Pin.PSObject.Properties['VerificationPath']
    $hash=$Pin.PSObject.Properties['SHA256']
    if($null -eq $path -or $null -eq $hash -or
       $path.Value -isnot [string] -or [string]::IsNullOrWhiteSpace($path.Value) -or
       $hash.Value -isnot [string] -or $hash.Value -cnotmatch '^[A-F0-9]{64}$'){
        throw 'OD: required pin identity invalid'
    }
    $role=$Pin.PSObject.Properties['Role']
    [pscustomobject][ordered]@{
        VerificationPath=[IO.Path]::GetFullPath($path.Value)
        SHA256=$hash.Value
        Role=$(if($null -eq $role){$null}else{$role.Value})
        ExpectedState='Present'
    }
}

function Assert-ODNeed([bool]$Condition,[string]$Message){
    if(-not $Condition){throw $Message}
}
function Get-ODHash([string]$Path){
    (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash
}
function Assert-ODHash([string]$Path,[string]$SHA256){
    Assert-ODNeed ((Get-ODHash $Path) -ceq $SHA256) ('OD: identity differs '+$Path)
}
function Save-ODOnce([string]$Path,[byte[]]$Bytes){
    $stream=$null;$primary=$null
    try{
        $stream=[IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
        $stream.Write($Bytes,0,$Bytes.Length);$stream.Flush()
    }catch{$primary=$_}
    finally{
        if($null -ne $stream){
            try{$stream.Dispose()}catch{if($null -eq $primary){$primary=$_}}
        }
    }
    if($null -ne $primary){throw $primary}
}
function Assert-ODPins($Pins){
    $seen=[Collections.Generic.Dictionary[string,string]]::new([StringComparer]::OrdinalIgnoreCase)
    foreach($pin in $Pins){
        $path=[IO.Path]::GetFullPath($pin.VerificationPath)
        if($pin.ExpectedState -ceq 'Absent'){
            Assert-ODNeed (-not (Test-Path -LiteralPath $path)) ('OD: expected absent '+$path)
            Assert-ODNeed ($null -eq $pin.SHA256) 'OD: absent hash must be null'
        }elseif($pin.ExpectedState -ceq 'Present'){
            Assert-ODHash $path $pin.SHA256
        }else{throw 'OD: unknown pin state'}
        $identity=$pin.ExpectedState+':'+$pin.SHA256
        if($seen.ContainsKey($path)){
            Assert-ODNeed ($seen[$path] -ceq $identity) ('OD: conflicting mapping '+$path)
        }else{$seen.Add($path,$identity)}
    }
}

$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-10-19-offline-diagnostic-entry-preparation-v2'
$revisionRelative='docs/superpowers/plans/2026-10-19-offline-diagnostic-entry-preparation-v2.md'
$revision=Join-Path $repo $revisionRelative
$spec=Join-Path $repo 'docs/superpowers/specs/2026-10-19-offline-diagnostic-entry-migration-design.md'
$plan=Join-Path $repo 'docs/superpowers/plans/2026-10-19-offline-diagnostic-entry-migration.md'
$specHash='DA27DA1C6BD3578512FB60A7F38B6635EAC9ACEF6570534F8B868BD07D088D9A'
$planHash='062AAC17317B5F8FBE5E29A88A67ABF6E53D5BD4BC6CA8151AC04DE783F81205'
$revisionHash='CB511F97B2FB40FE4B67B739EDEC2E36C11B2990B06C7359040F3EA7A1B63E9F'
$audit=[ordered]@{
    'collection-status.json'='99653A4F9BE56DCEAF73A331A095ED9C1BC7D26E13AAEB6B1E34BA54B44EB5D8'
    'recovered-regression-event.json'='8A9F21354290DE3A527E8D8CBA664BB752749F86BE992A56131FF3CFA266D787'
    'recovered-regression-output.json'='8954D1E6978E42B3F446CA425FBD1359A812452C410C37F6C9CB1DE7E77CEF3A'
}
$existing=@(Get-ChildItem -LiteralPath $packet -Force)
$allowed=@('progress.md')+@($audit.Keys)
Assert-ODNeed ($existing.Count -eq 4 -and @($existing | Where-Object {$_.PSIsContainer -or $_.Name -cnotin $allowed}).Count -eq 0) 'OD: preparation attempted or unknown packet content'
Assert-ODNeed ((Get-Content -LiteralPath (Join-Path $packet progress.md) -TotalCount 1) -ceq ('# SDD ledger — plan: '+$revisionRelative)) 'OD: ledger identity'
foreach($name in $audit.Keys){Assert-ODHash (Join-Path $packet $name) $audit[$name]}
Assert-ODHash $spec $specHash
Assert-ODHash $plan $planHash
Assert-ODHash $revision $revisionHash
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
Assert-ODHash $engine '362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
Assert-ODHash (Join-Path ([IO.Path]::GetDirectoryName($engine)) System.Management.Automation.dll) '2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'

$inputs=@(
    'scripts/.gitattributes',
    'docs/diagnostics/isolated-fuzz/p2b-entry-migration/.gitattributes',
    'docs/diagnostics/isolated-fuzz/p2b-entry-migration/README.md',
    'docs/diagnostics/isolated-fuzz/p2b-entry-migration/expected-output.json',
    'docs/diagnostics/isolated-fuzz/p2b-entry-migration/test-helpers.ps1',
    'docs/diagnostics/isolated-fuzz/p2b-entry-migration/entry.tests.ps1',
    'docs/diagnostics/isolated-fuzz/p2b-entry-migration/core.tests.ps1',
    'docs/diagnostics/isolated-fuzz/p2b-entry-migration/prefix.tests.ps1',
    'docs/diagnostics/isolated-fuzz/p2b-entry-migration/owned.tests.ps1',
    'docs/diagnostics/isolated-fuzz/p2b-entry-migration/live.tests.ps1',
    'docs/diagnostics/isolated-fuzz/p2b-entry-migration/capture.ps1',
    'docs/diagnostics/isolated-fuzz/p2b-entry-migration/driver.ps1',
    'scripts/diagnose-offline.ps1',
    'scripts/private/offline-diagnostics/catalog.json',
    'scripts/private/offline-diagnostics/session.psm1',
    'scripts/private/offline-diagnostics/sinks.ps1',
    'scripts/private/offline-diagnostics/owned-step.ps1',
    'scripts/private/offline-diagnostics/sequence.ps1',
    'scripts/private/offline-diagnostics/profiles.ps1',
    'scripts/private/offline-diagnostics/observations.ps1',
    'docs/diagnostics/isolated-fuzz/p2b-entry-preparation-v2/.gitattributes',
    'docs/diagnostics/isolated-fuzz/p2b-entry-preparation-v2/README.md',
    'docs/diagnostics/isolated-fuzz/p2b-entry-preparation-v2/prepare.ps1',
    'docs/diagnostics/isolated-fuzz/p2b-entry-preparation-v2/prepare.tests.ps1',
    $revisionRelative
)
$absent=@($inputs | Select-Object -Skip 12 -First 8)
Assert-ODNeed ($inputs.Count -eq 25 -and $absent.Count -eq 8) 'OD: explicit input inventory'
foreach($relative in $inputs){
    $path=Join-Path $repo $relative
    if($relative -cin $absent){
        Assert-ODNeed (-not (Test-Path -LiteralPath $path)) ('OD: planned runtime still absent '+$relative)
        continue
    }
    Assert-ODNeed ([IO.File]::Exists($path)) ('OD: missing explicit source '+$relative)
    Assert-ODNeed (-not ([IO.File]::ReadAllBytes($path) -contains 13)) ('OD: source not LF '+$relative)
    if($path.EndsWith('.ps1') -or $path.EndsWith('.psm1')){
        $tokens=$null;$errors=$null
        $null=[Management.Automation.Language.Parser]::ParseFile($path,[ref]$tokens,[ref]$errors)
        Assert-ODNeed ($errors.Count -eq 0) ('OD: parse error '+$relative)
    }
    if($path.EndsWith('.json')){$null=Get-Content -LiteralPath $path -Raw | ConvertFrom-Json}
}
$oldPacket=Join-Path $repo '.superpowers/sdd/2026-10-19-p2b-prestart-cause-repair'
$oldManifest=Join-Path $oldPacket classification-green.frozen.json
Assert-ODHash $oldManifest '01B99203F29B6073843756E347054F8489F5825C325B6F4084DF9C8018A07123'
$old=Get-Content -LiteralPath $oldManifest -Raw | ConvertFrom-Json
Assert-ODNeed ($old.Pins.Count -eq 1930) 'OD: inherited1930'
$pins=[Collections.Generic.List[object]]::new()
foreach($pin in $old.Pins){$pins.Add((Convert-ODInheritedPin $pin))}
Assert-ODPins $pins
$raw=@(Get-ChildItem -LiteralPath $oldPacket -Recurse -File -Force | Sort-Object FullName)
$oldLeaf=Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-prestart-cause-repair'
$oldLeafFiles=@(Get-ChildItem -LiteralPath $oldLeaf -File -Force | Sort-Object Name)
Assert-ODNeed ($raw.Count -eq 92 -and $oldLeafFiles.Count -eq 16) 'OD: historical92raw16leaf'
$failedPacket=Join-Path $repo '.superpowers/sdd/2026-10-19-offline-diagnostic-entry-migration'
$failedPortablePath=Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-entry-migration/portable-summary.json'
Assert-ODHash $failedPortablePath '71C76165D18395E240EEE241482DA4E70D88B5D42B9B3BC56B584A9E96417A18'
$failed=Get-Content -LiteralPath $failedPortablePath -Raw | ConvertFrom-Json
$failedNames=@('approved-plan.snapshot.md','approved-spec.snapshot.md','progress.md')
$failedFiles=@(Get-ChildItem -LiteralPath $failedPacket -Force)
Assert-ODNeed ($failed.Artifacts.Count -eq 3 -and $failedFiles.Count -eq 3 -and @($failedFiles | Where-Object {$_.PSIsContainer -or $_.Name -cnotin $failedNames}).Count -eq 0) 'OD: failed originals exact3'
foreach($artifact in $failed.Artifacts){
    Assert-ODNeed ($artifact.Path -cin $failedNames -and $artifact.Encoding -ceq 'base64') 'OD: failed record name/encoding'
    $actual=[IO.File]::ReadAllBytes((Join-Path $failedPacket $artifact.Path))
    Assert-ODNeed ($actual.Length -eq $artifact.Length -and [Convert]::ToBase64String($actual) -ceq $artifact.Data) 'OD: failed bytes differ'
    Assert-ODHash (Join-Path $failedPacket $artifact.Path) $artifact.SHA256
}
$failedReport=Join-Path $repo 'docs/roadmap/2026-10-19-offline-diagnostic-entry-migration.md'
Assert-ODHash $failedReport 'FBBB9DDA9E1309C1110CAFB5F60DE51636265E783D5CEDE81D068384F142CD60'

# First writes: immutable approved-document snapshots. Any partial write blocks retry.
$specSnapshot=Join-Path $packet approved-spec.snapshot.md
$planSnapshot=Join-Path $packet approved-plan.snapshot.md
$revisionSnapshot=Join-Path $packet approved-revision.snapshot.md
Save-ODOnce $specSnapshot ([IO.File]::ReadAllBytes($spec))
Save-ODOnce $planSnapshot ([IO.File]::ReadAllBytes($plan))
Save-ODOnce $revisionSnapshot ([IO.File]::ReadAllBytes($revision))
$dependencies=@(
    (Join-Path $repo 'docs/roadmap/2026-10-19-p2b-prestart-cause-repair.md'),
    (Join-Path $repo 'docs/roadmap/2026-10-19-p2b-c1-c2-closure-map.md'),
    $spec,$plan,$specSnapshot,$planSnapshot,$revisionSnapshot
)
$failedPaths=@($failedFiles.FullName)+@($failedPortablePath,$failedReport)
$extra=@($raw.FullName)+@($oldLeafFiles.FullName)+$dependencies+$failedPaths
foreach($path in $extra){
    $pins.Add([pscustomobject][ordered]@{VerificationPath=[IO.Path]::GetFullPath($path);SHA256=(Get-ODHash $path);Role='preparation-frozen-dependency';ExpectedState='Present'})
}
foreach($relative in $inputs){
    $path=[IO.Path]::GetFullPath((Join-Path $repo $relative))
    $isAbsent=$relative -cin $absent
    $pins.Add([pscustomobject][ordered]@{VerificationPath=$path;SHA256=$(if($isAbsent){$null}else{Get-ODHash $path});Role='preparation-planned-input';ExpectedState=$(if($isAbsent){'Absent'}else{'Present'})})
}
foreach($name in $audit.Keys){
    $pins.Add([pscustomobject][ordered]@{VerificationPath=[IO.Path]::GetFullPath((Join-Path $packet $name));SHA256=$audit[$name];Role='preserved-preparation-audit';ExpectedState='Present'})
}
Assert-ODNeed ($dependencies.Count -eq 7 -and $failedPaths.Count -eq 5 -and $pins.Count -eq 2078) 'OD: exact2078mapping count'
Assert-ODNeed (@($pins | Where-Object ExpectedState -CEQ Absent).Count -eq 8) 'OD: exact8Absent'
Assert-ODPins $pins
$snapshotRoot=Join-Path $packet snapshots
Assert-ODNeed (-not (Test-Path -LiteralPath $snapshotRoot)) 'OD: snapshot attempted'
$null=[IO.Directory]::CreateDirectory($snapshotRoot)
$snapshotCount=0
foreach($relative in $inputs){
    if($relative -cin $absent){continue}
    $target=[IO.Path]::GetFullPath((Join-Path $snapshotRoot $relative))
    Assert-ODNeed ($target.StartsWith(([IO.Path]::GetFullPath($snapshotRoot)+[IO.Path]::DirectorySeparatorChar),[StringComparison]::OrdinalIgnoreCase)) 'OD: snapshot path boundary'
    $null=[IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($target))
    Save-ODOnce $target ([IO.File]::ReadAllBytes((Join-Path $repo $relative)))
    Assert-ODHash $target (Get-ODHash (Join-Path $repo $relative))
    $snapshotCount++
}
Assert-ODNeed ($snapshotCount -eq 17) 'OD: exact17source snapshots'
Assert-ODPins $pins
$manifest=[pscustomobject][ordered]@{
    Schema=2;Stage='offline-diagnostic-entry-preparation-v2';Status='PREPARED_NOT_RUNTIME_AUTHORITY'
    BaseCommit='1ef57163845119ca3bce143d82042f92d1095cc4'
    OriginalPlanSHA256=$planHash;SpecSHA256=$specHash;ApprovedRevisionSHA256=$revisionHash
    AdjustmentCheckpoint='1ef57163845119ca3bce143d82042f92d1095cc4'
    InheritedPins=1930;InputCount=25;SourceSnapshots=17;Present=2070;Absent=8
    Pins=$pins.ToArray();CaptureHostAllowance=0;ChildAllowance=0;ReviewerAllowance=0
    OldRuntimeAllowancesFrozen=$true;Authority='None'
    SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false
    FormalAcceptance=$false;RealGateImplemented=$false
}
$manifestPath=Join-Path $packet preparation.frozen.json
Save-ODOnce $manifestPath ([Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $manifest -Depth 12)))
[pscustomobject][ordered]@{
    Prepared='preparation-only';Status=$manifest.Status;FrozenSHA256=(Get-ODHash $manifestPath)
    Pins=$pins.Count;Present=2070;Absent=8;SourceSnapshots=$snapshotCount
    CaptureHosts=0;Children=0;Reviewers=0;Authority='None'
}
