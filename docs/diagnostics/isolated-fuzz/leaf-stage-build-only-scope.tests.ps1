param([ValidateSet('all','scope','guard')][string]$Focus='all')
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'leaf-stage-build-only-fixture.ps1')
. (Join-Path $PSScriptRoot 'leaf-stage-build-only.ps1')
if($Focus -cne 'scope'){
    $n=New-Object 'IsolatedFuzzProcess+Result'
    foreach($v in @{CleanupVerified=$true;RootExited=$true;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;RootId=17;MembersBeforeTermination=[uint64[]]@();MembersAfterCleanup=[uint64[]]@();MembersAtRootWait=[uint64[]]@();ShutdownMemberEvidence=@()}.GetEnumerator()){$n.($v.Key)=$v.Value}
    $ordinary=ConvertTo-RecordedProcessReceipt $n 'out' 'err' 'exe' @() @{}
    Assert-ProcessCompletion $ordinary
    Check-BuildOnly ($null -eq $ordinary.PSObject.Properties['BuildOnlyDiagnostic']) 'OrdinaryMarkerAbsent'
    foreach($value in @($null,$false,'bad',[pscustomobject]@{Schema=1;Policy='leaf-stage-build-only-v1';VetMode='off';VetEvidenceComplete=$false})){
        $ordinary|Add-Member BuildOnlyDiagnostic $value -Force
        $message=$null;try{Assert-ProcessCompletion $ordinary}catch{$message=$_.Exception.Message}
        Check-BuildOnly ($message -ceq 'Build-only diagnostics are not acceptance') 'MarkerOnlyRefusal'
    }
}
if($Focus -ceq 'guard'){return}
Check-BuildOnly ($null -ne (Get-Command Get-LeafStageBuildOnlyEvidence -ErrorAction SilentlyContinue)) 'BuildOnlyScopeExists'
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
$unit=Join-Path $workspace ('.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/build-only-scope-unit-'+[guid]::NewGuid().ToString('N'))
$f=New-BuildOnlyTestFiles $unit;$plan=Get-LeafStageBuildOnlyPlan $f.Root $f.Toolchain $f.ReferenceReceipt $f.ModuleCache
$data=New-BuildOnlyNativeFixture $plan
New-Item -ItemType Directory -Path (Join-Path $f.Root 'build')|Out-Null
$stdout=Join-Path $f.Root 'build/stdout.txt';$stderr=Join-Path $f.Root 'build/stderr.txt'
[IO.File]::WriteAllText($stdout,'');[IO.File]::WriteAllText($stderr,$data.Log)
$receipt=ConvertTo-RecordedProcessReceipt $data.Native $stdout $stderr $plan.Executable $plan.BuildArguments $plan.Environment
$receipt|Add-Member WorkingDirectory $plan.Project
$receipt|Add-Member BuildOnlyDiagnostic ([pscustomobject]@{Schema=1;Policy='leaf-stage-build-only-v1';VetMode='off';VetEvidenceComplete=$false})
$before=Get-BuildOnlyBytesHash ([Text.Encoding]::UTF8.GetBytes(($receipt|ConvertTo-Json -Depth 9 -WarningAction Stop)))
$logBefore=(Get-FileHash $stderr).Hash
function Get-LeafStageBuildOnlyPlan {throw 'Saved validation must not call the unused-root planner'}
$e=Get-LeafStageBuildOnlyEvidence $receipt $plan $stderr
Check-BuildOnly ($e.ScopeEvidenceComplete -and $e.TriggerEvidenceComplete -and $e.CoreNaturalComplete -and $e.LifetimeEvidenceComplete -and $e.ObservationPassed -and -not $e.VetEvidenceComplete -and -not $e.Acceptance) 'TypedMarkerFacts'
$round=$receipt|ConvertTo-Json -Depth 9 -WarningAction Stop|ConvertFrom-Json
$rp=$plan|ConvertTo-Json -Depth 9 -WarningAction Stop|ConvertFrom-Json
$e=Get-LeafStageBuildOnlyEvidence $round $rp $stderr
Check-BuildOnly ($e.ObservationPassed -and $round.BuildOnlyDiagnostic.VetEvidenceComplete -is [bool] -and -not $round.BuildOnlyDiagnostic.VetEvidenceComplete) 'TypedMarkerRoundtrip'
Check-BuildOnly ($before -ceq (Get-BuildOnlyBytesHash ([Text.Encoding]::UTF8.GetBytes(($receipt|ConvertTo-Json -Depth 9 -WarningAction Stop)))) -and $logBefore -ceq (Get-FileHash $stderr).Hash) 'ExistingRootReadOnly'
foreach($case in @('missing-marker','null-marker','bad-schema','schema-string','bad-policy','bad-mode','true-vet','string-vet','extra-marker','missing-vet','duplicate-vet','arg-order','environment','extra-env','wrong-root','wrong-package','wrong-exe','wrong-stream','missing-pin','duplicate-pin','wrong-pin','plan-acceptance','plan-timeout','joint-vet','joint-env','joint-trigger')){
    $r=$receipt|ConvertTo-Json -Depth 9|ConvertFrom-Json;$p=$plan|ConvertTo-Json -Depth 9|ConvertFrom-Json
    switch($case){
        missing-marker{$r.PSObject.Properties.Remove('BuildOnlyDiagnostic')}
        null-marker{$r.BuildOnlyDiagnostic=$null}
        bad-schema{$r.BuildOnlyDiagnostic.Schema=2}
        schema-string{$r.BuildOnlyDiagnostic.Schema='1'}
        bad-policy{$r.BuildOnlyDiagnostic.Policy='other'}
        bad-mode{$r.BuildOnlyDiagnostic.VetMode='all'}
        true-vet{$r.BuildOnlyDiagnostic.VetEvidenceComplete=$true}
        string-vet{$r.BuildOnlyDiagnostic.VetEvidenceComplete='false'}
        extra-marker{$r.BuildOnlyDiagnostic|Add-Member Acceptance $true}
        missing-vet{$r.Arguments=@($r.Arguments|Where-Object {$_ -cne '-vet=off'})}
        duplicate-vet{$r.Arguments+=@('-vet=off')}
        arg-order{$r.Arguments[0]='-c';$r.Arguments[1]='test'}
        environment{$r.Environment.GOPROXY='direct'}
        extra-env{$r.Environment|Add-Member UNAPPROVED '1'}
        wrong-root{$p.RunRoot=Join-Path $unit 'leaf-stage-retained-compile-901'}
        wrong-package{$p.Package='internal/nodecontrol/nodeissuer'}
        wrong-exe{$r.Executable=Join-Path $unit 'other.exe'}
        wrong-stream{$r.StderrPath=$stdout}
        missing-pin{$p.SourceBinding=@($p.SourceBinding|Select-Object -Skip 1)}
        duplicate-pin{$p.SourceBinding[0].Path=$p.SourceBinding[1].Path}
        wrong-pin{$p.SourceBinding[0].SHA256='0'*64}
        plan-acceptance{$p.Acceptance=$true}
        plan-timeout{$p.BuildLimitSeconds=601}
        joint-vet{$p.BuildArguments[4]='-vet=all';$r.Arguments[4]='-vet=all'}
        joint-env{$p.Environment.GOPROXY='direct';$r.Environment.GOPROXY='direct'}
        joint-trigger{$p.Trigger.TotalByteLimit=33554432;$r.CompileStageTrigger.TotalByteLimit=33554432}
    }
    $e=Get-LeafStageBuildOnlyEvidence $r $p $stderr
    Check-BuildOnly (-not $e.ScopeEvidenceComplete -and -not $e.ObservationPassed -and -not [string]::IsNullOrEmpty($e.ScopeError)) ('ScopeRefusal '+$case)
}
foreach($case in @('trigger','cap32','zero','core87','missing-core','entry-clock')){
    $r=$receipt|ConvertTo-Json -Depth 9|ConvertFrom-Json
    switch($case){trigger{$r.CompileStageTrigger=$null} cap32{$r.LifetimeDiagnostics.Attempts=32;$r.LifetimeDiagnostics.AdmissionBudgetExhausted=$false} zero{$r.LifetimeDiagnostics.Attempts=0;$r.LifetimeDiagnostics.Admitted=0;$r.LifetimeDiagnostics.Entries=@()} core87{$r.ShutdownMemberEvidence[0].QueryError=87} missing-core{$r.ShutdownMemberEvidence=@()} entry-clock{$r.LifetimeDiagnostics.Entries[0].Open.StartTimestamp=1}}
    $e=Get-LeafStageBuildOnlyEvidence $r $plan $stderr
    Check-BuildOnly ($e.ScopeEvidenceComplete -and -not $e.ObservationPassed) ('IndependentFacts '+$case)
}
Write-Output 'LIMIT: synthetic receipts and saved byte streams only; no Run or PID query.'
