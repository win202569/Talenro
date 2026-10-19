Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:Repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$script:Packet=Join-Path $script:Repo '.superpowers/sdd/2026-10-19-p2b-prestart-cause-repair'
$script:Oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw | ConvertFrom-Json -AsHashtable
$script:Sessions=[Collections.Generic.List[object]]::new()
$script:Clock=[Diagnostics.Stopwatch]::StartNew()
$script:HostStarts=0
$script:Engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
. (Join-Path $PSScriptRoot sinks.ps1)
. (Join-Path $PSScriptRoot owned-step.ps1)
. (Join-Path $PSScriptRoot sequence.ps1)
. (Join-Path $PSScriptRoot '../p2b-process-association-guard-repair/owned-process-observation.ps1')

function Find-LSSession([object]$Handle){
    foreach($entry in $script:Sessions){if([object]::ReferenceEquals($entry.Handle,$Handle)){return $entry}}
    return $null
}
function Assert-LSPins($Entry){
    if((Get-FileHash -LiteralPath $Entry.ManifestPath -Algorithm SHA256).Hash -cne $Entry.Frozen){throw 'Frozen manifest identity changed'}
    foreach($pin in $Entry.Pins){
        if((Get-FileHash -LiteralPath $pin.VerificationPath -Algorithm SHA256).Hash -cne $pin.SHA256){throw ('Frozen input changed: '+$pin.VerificationPath)}
    }
}
function New-LSSession([string]$Grant,[string]$Case,[string]$FrozenSHA256){
    if($Grant -cnotin $script:Oracle.Grants.Keys -or $Case -cnotin $script:Oracle.Cases.Keys){throw 'Unknown grant/case'}
    $profile=$script:Oracle.Cases[$Case]
    if($Grant -cnotin $profile.Grants -or $FrozenSHA256 -cnotmatch '^[A-F0-9]{64}$'){throw 'Invalid grant/case/epoch'}
    if($script:Sessions.Count -ge 32){throw 'Session limit'}
    foreach($prior in $script:Sessions){if($prior.Grant -ceq $Grant -and $prior.Case -ceq $Case){throw 'Case already reserved in this host'}}
    $phase=if($Grant -cin @('owned-green','sequence-green')){'classification-green'}else{$Grant}
    $manifestPath=Join-Path $script:Packet ($phase+'.frozen.json')
    if((Get-FileHash -LiteralPath $manifestPath -Algorithm SHA256).Hash -cne $FrozenSHA256){throw 'Unbound frozen input'}
    $manifest=Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json
    $root=Join-Path $script:Packet ($Grant+'/'+$Case)
    if([IO.Directory]::Exists($root) -or [IO.File]::Exists($root)){throw 'Case path already exists: no resume'}
    $entry=[pscustomobject]@{
        Handle=[object]::new();Grant=$Grant;Case=$Case;Frozen=$FrozenSHA256;ManifestPath=$manifestPath;Pins=$manifest.Pins
        Profile=$profile;TableSHA256=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.Encoding]::UTF8.GetBytes(($profile | ConvertTo-Json -Depth 10 -Compress))))
        Root=$root;Status='New';FailureLatched=$false;NextOrdinal=0;Reserved=0;StartCalls=0
        AckIssued=0;AckConsumed=0;Acks=[Collections.Generic.List[object]]::new();Sinks=[ordered]@{}
        PrimaryError=$null;SecondaryErrors=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new();Overflow=$false
    }
    Assert-LSPins $entry
    $null=[IO.Directory]::CreateDirectory($root)
    foreach($pair in @(@('Ledger','ledger.json'),@('Start','start.json'),@('Receipt0','receipt-0.json'),@('Receipt1','receipt-1.json'),@('Out0','stdout-0.bin'),@('Err0','stderr-0.bin'),@('Out1','stdout-1.bin'),@('Err1','stderr-1.bin'),@('Trace1','seam-trace.json'))){Register-LSSink $entry $pair[0] (Join-Path $root $pair[1]) File}
    Register-LSSink $entry ConsoleOut ConsoleOut Output
    Register-LSSink $entry ConsoleError ConsoleError Output
    $script:Sessions.Add($entry)
    return $entry.Handle
}
function Add-LSFailure([object]$Session,[Management.Automation.ErrorRecord]$Record){
    $entry=Find-LSSession $Session
    if($null -eq $entry -or $null -eq $Record){return}
    $entry.FailureLatched=$true
    if($entry.Status -cne 'Closed'){$entry.Status='Failed'}
    foreach($ack in $entry.Acks){$ack.Revoked=$true}
    if($null -eq $entry.PrimaryError){$entry.PrimaryError=$Record;return}
    if([object]::ReferenceEquals($entry.PrimaryError,$Record)){return}
    foreach($existing in $entry.SecondaryErrors){if([object]::ReferenceEquals($existing,$Record)){return}}
    if($entry.SecondaryErrors.Count -lt 8){$entry.SecondaryErrors.Add($Record)}else{$entry.Overflow=$true}
}
function Test-LSAdvance([object]$Session,[int]$Ordinal,[object]$PreviousAck){
    $entry=Find-LSSession $Session
    if($null -eq $entry){return $false}
    # Only private live registration can grant continuation. No disk read,
    # deserialization, supplied fields or Outcome participate in this gate.
    if($entry.FailureLatched -or $entry.Status -cnotin @('New','Active') -or $Ordinal -ne $entry.NextOrdinal -or $Ordinal -lt 0 -or $Ordinal -ge $entry.Profile.Steps.Count -or $entry.Reserved -ge $entry.Profile.Budget){return $false}
    if($Ordinal -eq 0){return ($entry.Status -ceq 'New' -and $null -eq $PreviousAck)}
    foreach($ack in $entry.Acks){
        if([object]::ReferenceEquals($ack.Handle,$PreviousAck) -and -not $ack.Consumed -and -not $ack.Revoked -and [object]::ReferenceEquals($ack.Owner,$entry) -and $ack.Ordinal -eq ($Ordinal-1) -and $ack.Frozen -ceq $entry.Frozen -and $ack.Table -ceq $entry.TableSHA256){return $true}
    }
    return $false
}
function Get-LSView([object]$Session){
    $e=Find-LSSession $Session
    if($null -eq $e){return $null}
    $sinks=[ordered]@{}
    foreach($key in $e.Sinks.Keys){$s=$e.Sinks[$key];$sinks[$key]=[pscustomobject][ordered]@{Path=$s.Path;State=$s.State;LogicalIOAttempts=$s.LogicalIOAttempts;RejectedWithoutIO=$s.RejectedWithoutIO}}
    return [pscustomobject][ordered]@{Status=$e.Status;FailureLatched=$e.FailureLatched;NextOrdinal=$e.NextOrdinal;Reserved=$e.Reserved;StartCalls=$e.StartCalls;AckIssued=$e.AckIssued;AckConsumed=$e.AckConsumed;Root=$e.Root;Sinks=[pscustomobject]$sinks;PrimaryError=$e.PrimaryError;SecondaryErrors=$e.SecondaryErrors.ToArray();Overflow=$e.Overflow}
}
function Close-LSSession([object]$Session,[ValidateSet('Finish','Stop')][string]$Mode){
    $e=Find-LSSession $Session
    if($null -eq $e){return $null}
    if($Mode -ceq 'Finish' -and ($e.FailureLatched -or $e.NextOrdinal -ne $e.Profile.Steps.Count)){throw 'Cannot finish incomplete session'}
    foreach($ack in $e.Acks){$ack.Revoked=$true}
    $e.Status='Closed'
    return Get-LSView $Session
}
Export-ModuleMember -Function New-LSSession,Test-LSAdvance,Invoke-LSStep,Write-LSLedger,Read-LSLedger,Complete-LSLedger,Write-LSOutput,Get-LSView,Add-LSFailure,Close-LSSession
