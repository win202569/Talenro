# Exploratory memory-only timing probe,not a test/acceptance or benchmark entry.
# Each launch needs separate bounded authorization. Does not replay the suite,
# call public file-store APIs,replace functions,write archives or change limits.
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
$events=New-Object object[] 4096
for($i=0;$i -lt 4096;$i++){$events[$i]=[pscustomobject]@{Kind='tick';ElapsedMs=[long]0}}
$packet=[pscustomobject]@{Schema=1;SourceMode='synthetic';SessionLabel='s';ExecutableDeclaration='x';ArgumentBytes=[byte[]]@();EncodingName='utf-8';OriginDeclaration='o';StartedUTC='2026-10-17T00:00:00.0000000Z';EndedUTC='2026-10-17T00:00:00.0000000Z';Events=$events}
Complete-ProbeStage
Start-ProbeStage 'construct-record'
$record=New-SyntheticCaptureRecord $packet
if(@($record.PSObject.Properties).Count -ne 20 -or $record.Events.Count -ne 4096){throw 'Probe record construction failed'}
Complete-ProbeStage
Start-ProbeStage 'verify-original'
$verification=Test-SyntheticCaptureRecord $record
if(-not $verification.Verified){throw 'Probe original record invalid'}
Complete-ProbeStage
$encodeMemory=$null;$writer=$null;$decodeMemory=$null;$reader=$null
try{
    Start-ProbeStage 'encode-memory'
    $encodeMemory=New-Object IO.MemoryStream;$writer=New-Object IO.BinaryWriter($encodeMemory)
    $writer.Write([Text.Encoding]::ASCII.GetBytes('TLRSYN01'));$writer.Write([int]1)
    # Trusted literal fixture only;using the actual private codec without wrappers
    # or replacement. This is NOT certification of helper hostile-input boundaries.
    Write-SyntheticStoreObject $writer $record 'record';$writer.Flush();$payload=$encodeMemory.ToArray()
    if($payload.Length -ne 37172){throw 'Unexpected 4096-tick fixture payload length'}
    Complete-ProbeStage
    Start-ProbeStage 'decode-memory'
    $decodeMemory=New-Object IO.MemoryStream(,$payload);$reader=New-Object IO.BinaryReader($decodeMemory)
    if([Text.Encoding]::ASCII.GetString($reader.ReadBytes(8)) -cne 'TLRSYN01' -or $reader.ReadInt32() -ne 1){throw 'Probe envelope mismatch'}
    $decoded=Read-SyntheticStoreObject $reader 'record'
    if($decodeMemory.Position -ne $decodeMemory.Length -or $decoded.Events.Count -ne 4096 -or -not [string]::Equals($decoded.BindingSHA256,$record.BindingSHA256,[StringComparison]::Ordinal)){throw 'Probe decoded record mismatch'}
    Complete-ProbeStage
    Start-ProbeStage 'verify-decoded'
    $verification=Test-SyntheticCaptureRecord $decoded
    if(-not $verification.Verified -or $decoded.Lifecycle.Status -cne 'incomplete' -or $decoded.Lifecycle.AcceptedEventCount -ne 4096 -or $null -ne $decoded.Lifecycle.ExitCode -or $null -ne $decoded.Lifecycle.Package -or $decoded.LiveActionsPerformed -or $decoded.OriginVerified -or $decoded.Acceptance -or $decoded.RuntimeApproved){throw 'Probe reconstructed record invalid'}
    Complete-ProbeStage
}finally{foreach($resource in @($reader,$decodeMemory,$writer,$encodeMemory)){if($null -ne $resource){$resource.Dispose()}}}
[Console]::Out.WriteLine('DIAGNOSTIC COMPLETE: 4096 ticks,memory only;not suite acceptance.');[Console]::Out.Flush()
