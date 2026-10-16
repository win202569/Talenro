$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'leaf-stage-build-only-fixture.ps1')
$helper=Join-Path $PSScriptRoot 'leaf-stage-build-only.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
Check-BuildOnly ($null -ne (Get-Command Get-LeafStageBuildOnlyPlan -ErrorAction SilentlyContinue)) 'BuildOnlyPlannerExists'
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
$unit=Join-Path $workspace ('.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/build-only plan unit '+[guid]::NewGuid().ToString('N'))
$f=New-BuildOnlyTestFiles $unit
function Assert-UninstrumentedEnvelopeToolchain {throw 'Unexpected SDK audit during planning'}
function Get-EnvelopeProductManifest {throw 'Unexpected product inventory during planning'}
function Invoke-LeafStageRetainedCompileNative {throw 'Unexpected legacy launch during planning'}
$plan=Get-LeafStageBuildOnlyPlan $f.Root $f.Toolchain $f.ReferenceReceipt $f.ModuleCache
$expected=@('test','-c','-x','-p=1','-vet=off','-count=1','-timeout=120s',('-o='+(Join-Path $f.Root 'package.test.exe')),'talenro.local/platform/internal/nodecontrol/nodeleafbinding')
Check-BuildOnly (($plan.BuildArguments -join "`n") -ceq ($expected -join "`n")) 'FixedArguments'
$envExpected=@{GOROOT=$f.Toolchain;GOCACHE=(Join-Path $f.Root 'go-cache');GOMODCACHE=$f.ModuleCache;TEMP=(Join-Path $f.Root 'tmp');TMP=(Join-Path $f.Root 'tmp');GOTMPDIR=(Join-Path $f.Root 'tmp');GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='on';GOWORK='off';GOFLAGS='-mod=readonly';GOOS='windows';GOARCH='amd64';TALENRO_FUZZ_DIAG_MODE='';TALENRO_FUZZ_FIXTURE='';GODEBUG=''}
Check-BuildOnly ($plan.Environment.Count -eq 20) 'FixedEnvironmentKeyCount'
foreach($pair in $envExpected.GetEnumerator()){Check-BuildOnly ($plan.Environment[$pair.Key] -ceq $pair.Value) ('FixedEnvironment '+$pair.Key)}
Check-BuildOnly ($plan.BuildLimitSeconds -eq 600 -and $plan.DiagnosticPolicy -ceq 'leaf-stage-build-only-v1' -and $plan.VetMode -ceq 'off' -and $plan.NoRetry -and $plan.DiagnosticOnly -and $plan.Coverage -ceq 'sampled') 'FixedScopePolicy'
foreach($name in @('VetEvidenceComplete','Acceptance','PackagePassed','OriginalTestPassed','RuntimeAttempted')){Check-BuildOnly ($plan.$name -is [bool] -and -not $plan.$name) ('TypedFalse '+$name)}
$names=@('leaf-stage-build-only.ps1','leaf-stage-retained-compile.ps1','leaf-retained-compile.ps1','retained-child-observation.ps1','mapping-single-check.ps1','matrix-check.ps1','envelope-uninstrumented.ps1','envelope-sample.ps1','envelope-observe.ps1','run.ps1','isolation.ps1','provenance.ps1','process-native.cs','retained-lifetime.cs','compile-stage-trigger.cs','compile-stage-session.cs')
Check-BuildOnly ($plan.SourceBinding.Count -eq 16 -and @(Compare-Object $names @($plan.SourceBinding.Path)).Count -eq 0) 'CompleteNamedPins'
foreach($row in $plan.SourceBinding){Check-BuildOnly ($row.SHA256 -ceq (Get-FileHash -LiteralPath (Join-Path $PSScriptRoot $row.Path)).Hash) ('SourceHash '+$row.Path)}
Check-BuildOnly ($plan.Trigger.Schema -eq 1 -and $plan.Trigger.Policy -ceq 'leaf-test-compile-log-v1' -and $plan.Trigger.TotalByteLimit -eq 16777216 -and $plan.Trigger.IterationByteLimit -eq 65536 -and $plan.Trigger.LineByteLimit -eq 65536 -and $plan.Trigger.PollLimit -eq 60000 -and $plan.Trigger.Compiler -ceq (Join-Path $f.Toolchain 'pkg/tool/windows_amd64/compile.exe')) 'FixedTriggerUnchanged'
Assert-LeafStageBuildOnlySources $plan
Check-BuildOnly (-not(Test-Path -LiteralPath $f.Root)) 'PlanningCreatesNoRunRoot'
$round=$plan|ConvertTo-Json -Depth 8 -WarningAction Stop|ConvertFrom-Json
Check-BuildOnly (@($round.BuildArguments).Count -eq 9 -and @($round.SourceBinding).Count -eq 16 -and $round.Environment.GODEBUG -ceq '' -and $round.VetEvidenceComplete -is [bool]) 'FlatPlanRoundtrip'
Assert-LeafStageBuildOnlySources $round
foreach($case in @('duplicate','missing','bad-hash')){
    $bad=$plan|ConvertTo-Json -Depth 8|ConvertFrom-Json
    switch($case){duplicate{$bad.SourceBinding[0].Path=$bad.SourceBinding[1].Path} missing{$bad.SourceBinding=@($bad.SourceBinding|Select-Object -Skip 1)} bad-hash{$bad.SourceBinding[0].SHA256='0'*64}}
    $failed=$false;try{Assert-LeafStageBuildOnlySources $bad}catch{$failed=$true};Check-BuildOnly $failed ('PlanSourceDrift '+$case)
}
foreach($case in @('wrong-root','overlap','missing-reference','missing-module','existing')){
    $root=$f.Root;$reference=$f.ReferenceReceipt;$modules=$f.ModuleCache
    switch($case){wrong-root{$root=Join-Path $unit 'leaf-stage-retained-compile-901'} overlap{$root=Join-Path $f.Toolchain 'leaf-stage-build-only-901'} missing-reference{$reference=Join-Path $unit 'absent.json'} missing-module{$modules=Join-Path $unit 'absent-modules'} existing{New-Item -ItemType Directory -Path $root|Out-Null}}
    $failed=$false;try{$null=Get-LeafStageBuildOnlyPlan $root $f.Toolchain $reference $modules}catch{$failed=$true};Check-BuildOnly $failed ('PlannerRejects '+$case)
}
foreach($kind in @('eager','stage')){
    $root=Join-Path $unit $(if($kind -ceq 'eager'){'leaf-retained-compile-901'}else{'leaf-stage-retained-compile-901'})
    $old=if($kind -ceq 'eager'){Get-LeafRetainedCompilePlan $root $f.Toolchain $f.ReferenceReceipt $f.ModuleCache}else{Get-LeafStageRetainedCompilePlan $root $f.Toolchain $f.ReferenceReceipt $f.ModuleCache}
    $oldExpected=@('test','-c','-x','-p=1','-count=1','-timeout=120s',('-o='+(Join-Path $root 'package.test.exe')),'talenro.local/platform/internal/nodecontrol/nodeleafbinding')
    Check-BuildOnly (($old.BuildArguments -join "`n") -ceq ($oldExpected -join "`n")) ('LegacyVetEnabled '+$kind)
}
[IO.File]::AppendAllText((Join-Path $f.Toolchain 'bin/go.exe'),'drift')
$failed=$false;try{Assert-LeafStageBuildOnlySources $plan}catch{$failed=$true};Check-BuildOnly $failed 'ExecutableDrift'
Write-Output 'LIMIT: tiny files only; no Go/native runner/PID query/full SDK audit.'
