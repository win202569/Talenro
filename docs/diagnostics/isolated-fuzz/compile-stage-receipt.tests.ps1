$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'run.ps1')
function Check-StageReceipt([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
Check-StageReceipt ($null -ne [IsolatedFuzzProcess+Result].GetField('CompileStageTrigger')) 'optional compile-stage report exists'
Check-StageReceipt ($null -ne [IsolatedFuzzProcess].GetMethod('RunWithCompileStageRetainedDiagnostics')) 'explicit native compile-stage entry exists'
$n=New-Object 'IsolatedFuzzProcess+Result';$n.CleanupVerified=$true;$n.RootExited=$true;$n.NaturalShutdownVerified=$true;$n.ShutdownEvidenceComplete=$true;$n.RootId=17;$n.MembersBeforeTermination=[uint64[]]@();$n.MembersAfterCleanup=[uint64[]]@();$n.MembersAtRootWait=[uint64[]]@();$n.ShutdownMemberEvidence=@()
$legacy=ConvertTo-RecordedProcessReceipt $n 'out' 'err' 'exe' @() @{}
Check-StageReceipt ($null -eq $legacy.PSObject.Properties['CompileStageTrigger']) 'LegacyDefaultNullOmitted'
Assert-ProcessCompletion $legacy;Check-StageReceipt $true 'legacy ordinary guard unchanged'
$report=New-Object CompileStageTriggerReport;$report.State='Activated';$report.RawLineBase64='YQo=';$report.LineStart=0;$report.LineEnd=2;$report.LineNumber=1;$report.LineSHA256=('A'*64);$report.PrefixSHA256=('B'*64);$report.CandidateObserved=New-Object LifetimeApiEvidence;$report.CandidateObserved.Frequency=1000;$report.Activation=New-Object LifetimeApiEvidence;$report.RootCheck=New-Object CompileStageRootCheck;$report.FirstRegistryQuery=New-Object LifetimeQueryEvidence
$n.CompileStageTrigger=$report;$n.LifetimeDiagnosticOnly=$true;$n.ShutdownDiagnosticOnly=$true;$r=ConvertTo-RecordedProcessReceipt $n 'out' 'err' 'exe' @() @{};$round=$r|ConvertTo-Json -Depth 8 -WarningAction Stop|ConvertFrom-Json
Check-StageReceipt ($round.CompileStageTrigger.RawLineBase64 -ceq 'YQo=' -and $round.CompileStageTrigger.LineEnd -eq 2 -and $round.CompileStageTrigger.CandidateObserved.Frequency -eq 1000 -and $null -eq $round.CompileStageTrigger.CandidateObserved.StartUTC -and $round.CompileStageTrigger.LineSHA256 -ceq ('A'*64)) 'TriggerRoundtrip populated report and nulls'
foreach($value in @($null,$false,'malformed',$report)){$receipt=[pscustomobject]@{CompileStageTrigger=$value;LifetimeDiagnosticOnly=$false;ShutdownDiagnosticOnly=$false};$refusal=$null;try{Assert-ProcessCompletion $receipt}catch{$refusal=$_.Exception.Message};Check-StageReceipt ($refusal -ceq 'Compile-stage diagnostics are not acceptance') 'present trigger refused with cleared flags'}
$refusal=$null;try{Assert-ProcessCompletion $round}catch{$refusal=$_.Exception.Message};Check-StageReceipt ($refusal -ceq 'Compile-stage diagnostics are not acceptance') 'both flags true refused'
# Explicit static integration checks, never native execution.
$code=Get-Content (Join-Path $PSScriptRoot 'process-native.cs') -Raw
Check-StageReceipt ($code -match 'stageSession.Tick\(runtimeDeadline\)' -and $code -match 'new NativeCompileStageRootProbe\(pi.Process\)') 'STATIC exact-root stage wiring'
foreach($file in @('run.ps1','preflight.ps1','retained-child-observation.ps1','leaf-retained-compile.ps1','retained-lifetime-integration.tests.ps1')){$text=Get-Content (Join-Path $PSScriptRoot $file) -Raw;Check-StageReceipt ($text.Contains('compile-stage-trigger.cs') -and $text.Contains('compile-stage-session.cs')) ('STATIC LoaderInventoryCompatibility '+$file)}
