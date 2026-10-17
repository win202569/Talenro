# Exploratory trusted-fixture probe only,not a public API test or acceptance.
# Separate finite authorization required for every launch. No product function
# replacement,archive writes,public store calls or repeated full verification.
$ErrorActionPreference='Stop'
$script:probeTotal=[Diagnostics.Stopwatch]::StartNew()
$script:probeStage=[Diagnostics.Stopwatch]::new()
$script:probeName=''
function Write-ProbeTiming([string]$Event,[string]$Name,[double]$ElapsedMs){
    $item=[ordered]@{Event=$Event;Stage=$Name;ElapsedMs=[Math]::Round($ElapsedMs,3);TotalMs=[Math]::Round($script:probeTotal.Elapsed.TotalMilliseconds,3)}
    [Console]::Out.WriteLine('TIMING '+(ConvertTo-Json -InputObject $item -Compress));[Console]::Out.Flush()
}
function Start-ProbeStage([string]$Name){$script:probeName=$Name;$script:probeStage.Restart();Write-ProbeTiming 'start' $Name 0}
function Complete-ProbeStage{Write-ProbeTiming 'end' $script:probeName $script:probeStage.Elapsed.TotalMilliseconds}
[Console]::Out.WriteLine('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess);[Console]::Out.Flush()
Start-ProbeStage 'import'
. (Join-Path $PSScriptRoot 'synthetic-capture-store.ps1')
Complete-ProbeStage
Start-ProbeStage 'packet-build'
$fixtureEvents=New-Object object[] 4096
for($i=0;$i -lt 4096;$i++){$fixtureEvents[$i]=[pscustomobject]@{Kind='tick';ElapsedMs=[long]0}}
$fixturePacket=[pscustomobject]@{Schema=1;SourceMode='synthetic';SessionLabel='s';ExecutableDeclaration='x';ArgumentBytes=[byte[]]@();EncodingName='utf-8';OriginDeclaration='o';StartedUTC='2026-10-17T00:00:00.0000000Z';EndedUTC='2026-10-17T00:00:00.0000000Z';Events=$fixtureEvents}
Complete-ProbeStage
Start-ProbeStage 'construct-record'
$record=New-SyntheticCaptureRecord $fixturePacket
if(@($record.PSObject.Properties).Count -ne 20 -or $record.Events.Count -ne 4096){throw 'Probe construction failed'}
Complete-ProbeStage
# Follow the successful verifier path exactly once,using existing helpers.
# Exceptions intentionally escape instead of becoming the public verdict.
Start-ProbeStage 'record-header'
Assert-ContractShape $record @('Schema','Policy','SourceMode','SessionLabel','ExecutableDeclaration','Arguments','EncodingName','OriginDeclaration','StartedUTC','EndedUTC','UtcOrderConsistent','Events','Lifecycle','StopIntent','CleanupStatus','LiveActionsPerformed','BindingSHA256','OriginVerified','Acceptance','RuntimeApproved')
Assert-OfflineEvidenceInteger $record.Schema 1 1
Assert-ContractString $record.Policy 29 29
if(-not (Test-ContractOrdinalEqual $record.Policy 'synthetic-capture-contract-v1')){throw 'Unknown policy'}
foreach($flag in @('LiveActionsPerformed','OriginVerified','Acceptance','RuntimeApproved')){if((Get-OfflineCaptureValueType $record.$flag) -ne [bool] -or $record.$flag){throw 'No action or authority allowed'}}
Complete-ProbeStage
Start-ProbeStage 'read-snapshot-events'
$arguments=Read-ContractSnapshot $record.Arguments 65536
$events=Read-ContractStoredEvents $record.Events
Complete-ProbeStage
Start-ProbeStage 'rebuild-packet'
$packet=[pscustomobject]@{Schema=1;SourceMode=$record.SourceMode;SessionLabel=$record.SessionLabel;ExecutableDeclaration=$record.ExecutableDeclaration;ArgumentBytes=$arguments;EncodingName=$record.EncodingName;OriginDeclaration=$record.OriginDeclaration;StartedUTC=$record.StartedUTC;EndedUTC=$record.EndedUTC;Events=$events}
$clean=Get-ContractPacket $packet
Complete-ProbeStage
Start-ProbeStage 'rebuild-expected-record'
$expected=New-ContractValidatedRecord $clean
Complete-ProbeStage
Start-ProbeStage 'binding-compare'
Assert-ContractString $record.BindingSHA256 64 64
if($record.BindingSHA256 -cnotmatch '^[A-F0-9]{64}\z' -or -not (Test-ContractOrdinalEqual $record.BindingSHA256 $expected.BindingSHA256)){throw 'Binding mismatch'}
Complete-ProbeStage
Start-ProbeStage 'recursive-equivalence'
Assert-ContractEquivalent $record $expected
Complete-ProbeStage
Start-ProbeStage 'utc-final-facts'
if(-not $expected.UtcOrderConsistent){throw 'Clock reversal'}
if($record.Events.Count -ne 4096 -or $record.Lifecycle.Status -cne 'incomplete' -or $record.Lifecycle.AcceptedEventCount -ne 4096 -or $null -ne $record.Lifecycle.ExitCode -or $null -ne $record.Lifecycle.Package -or $record.LiveActionsPerformed -or $record.OriginVerified -or $record.Acceptance -or $record.RuntimeApproved){throw 'Unexpected fixture facts'}
Complete-ProbeStage
[Console]::Out.WriteLine('DIAGNOSTIC COMPLETE: internal path,4096 ticks,memory only;not public verifier or suite acceptance.');[Console]::Out.Flush()
