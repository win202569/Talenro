param([Parameter(Mandatory)][string]$TracePath,[Parameter(Mandatory)][string]$DetailsPath)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-19-p2b-completion-publication-repair'
if([IO.Path]::GetFullPath($TracePath) -cne (Join-Path $packet 'containment.trace.json') -or [IO.Path]::GetFullPath($DetailsPath) -cne (Join-Path $packet 'containment.details.json')){throw 'Fixed outputs required'}
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
if((Get-FileHash -LiteralPath $engine).Hash -cne '362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'){throw 'Engine identity'}
$id=[Security.Principal.WindowsIdentity]::GetCurrent()
if(([Security.Principal.WindowsPrincipal]::new($id)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary test host required'}
. (Join-Path $PSScriptRoot 'bookkeeping.ps1')
$failure=New-P2BFailureState
. (Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-process-association-guard-repair/owned-process-observation.ps1')
$oldLeaf=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-controller-exception-containment'
. (Join-Path $oldLeaf 'owned-run.ps1')
$groups=[Collections.Generic.List[string]]::new();$observations=[ordered]@{}
$rescues=0;$faultErrors=@();$results=[ordered]@{};$guards=[ordered]@{};$associationDetails=[ordered]@{}
function C-Need([bool]$Condition,[string]$Name){if(-not $Condition){throw ('ASSERTION FAILED: '+$Name)}}
function C-Save($Path,$Value){$null=Save-P2BLedger $failure $Path $Value}
function C-Pass([string]$Name){$groups.Add($Name);[Console]::Out.WriteLine('PASS: '+$Name)}
function C-SameIOCause([object]$Left,[object]$Right){
    if($Left -isnot [Management.Automation.ErrorRecord] -or $Right -isnot [Management.Automation.ErrorRecord]){return $false}
    [Exception[]]$leaves=@($null,$null)
    $records=[object[]]@($Left,$Right)
    for($side=0;$side -lt 2;$side++){
        [Management.Automation.ErrorRecord]$record=$records[$side]
        [Exception]$node=$record.Exception
        $seen=[Collections.Generic.List[Exception]]::new()
        while($null -ne $node){
            if($seen.Count -ge 8){return $false}
            foreach($prior in $seen){if([object]::ReferenceEquals($prior,$node)){return $false}}
            $seen.Add($node)
            if($node -is [UnauthorizedAccessException] -or $node -is [IO.IOException]){$leaves[$side]=$node}
            $node=$node.InnerException
        }
        if($null -eq $leaves[$side]){return $false}
    }
    return [object]::ReferenceEquals($leaves[0],$leaves[1])
}
function C-PrimaryPreserved([object]$Result,[object]$FirstAnchor,[object]$FinalAnchor){
    if($null -eq $Result -or $Result -isnot [pscustomobject]){return $false}
    $primaryProperty=$Result.PSObject.Properties['PrimaryError']
    $cleanupProperty=$Result.PSObject.Properties['CleanupErrors']
    if($primaryProperty -isnot [Management.Automation.PSNoteProperty] -or $cleanupProperty -isnot [Management.Automation.PSNoteProperty]){return $false}
    if($Result.CleanupErrors -isnot [Management.Automation.ErrorRecord[]] -or $Result.CleanupErrors.Count -ne 1){return $false}
    return ((C-SameIOCause $Result.PrimaryError $FirstAnchor) -and
        -not (C-SameIOCause $Result.PrimaryError $FinalAnchor) -and
        (C-SameIOCause $Result.CleanupErrors[0] $FinalAnchor) -and
        -not (C-SameIOCause $Result.CleanupErrors[0] $FirstAnchor))
}
function C-Describe([Management.Automation.ErrorRecord]$Record){
    $items=[Collections.Generic.List[object]]::new()
    [Exception]$node=$Record.Exception
    while($null -ne $node -and $items.Count -lt 8){
        $message=$node.Message
        $truncated=$message.Length -gt 1024
        if($truncated){$message=$message.Substring(0,1024)}
        $items.Add([pscustomobject][ordered]@{Type=$node.GetType().FullName;Message=$message;MessageTruncated=$truncated})
        $node=$node.InnerException
    }
    return [pscustomobject][ordered]@{RecordType=$Record.GetType().FullName;FullyQualifiedErrorId=$Record.FullyQualifiedErrorId;Chain=$items.ToArray();ChainTruncated=$null -ne $node}
}
function C-Process([string]$Mode){
    $p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
    $p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true
    $p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
    foreach($a in @('-NoProfile','-NonInteractive','-File',(Join-Path $oldLeaf 'fixture-child.ps1'),'-Mode',$Mode)){$p.StartInfo.ArgumentList.Add($a)}
    return $p
}
function C-Consume([string]$Case){
    $hosts=@(Get-Content -LiteralPath (Join-Path $packet 'hosts.json') -Raw|ConvertFrom-Json)
    if($hosts.Count -ne 3 -or $hosts[0].Outcome -cne 'EXPECTED_RED' -or $hosts[1].Outcome -cne 'EXPECTED_GREEN' -or $hosts[2].Slot -cne 'containment' -or -not $hosts[2].AttemptConsumed -or $hosts[2].Outcome -cne 'INCOMPLETE'){throw 'GREEN host guard'}
    $path=Join-Path $packet 'children.json';$rows=@()
    if(Test-Path -LiteralPath $path){$rows=@(Get-Content -LiteralPath $path -Raw|ConvertFrom-Json)}
    $order=@('fault','normal','timeout')
    if($rows.Count -ge 3 -or $Case -cne $order[$rows.Count]){throw 'Ordered child budget'}
    if(@($rows|Where-Object{$_.Outcome -cne 'EXPECTED' -or -not $_.OwnedExit -or $_.Rescue}).Count -ne 0){throw 'Previous child incomplete'}
    $row=[pscustomobject][ordered]@{Number=($rows.Count+1);Case=$Case;AttemptConsumed=$true;Started=$false;PID=$null;OwnedExit=$false;ExitCode=$null;Rescue=$false;Outcome='INCOMPLETE'}
    $rows+=,$row;C-Save $path $rows
    if($null -ne $failure.PrimaryError){throw $failure.PrimaryError}
    return $row
}
function C-Finish($Row,$Process,[bool]$KnownStarted){
    $a=$null;$kill=$false;$wait=$false;$pidMatches=$false
    $errors=[Collections.Generic.List[string]]::new()
    function G-Error($e){if($e -is [Management.Automation.ErrorRecord]){Add-P2BFailure $failure $e};if($errors.Count -lt 8){$v=$e.ToString();$errors.Add($v.Substring(0,[Math]::Min(1024,$v.Length)))}}
    try{
        $a=Read-P2BProcessAssociation $Process
        if($a.State -ceq 'Associated'){
            $directId=$Process.Id
            $pidMatches=$directId -is [int] -and $directId -gt 0 -and $a.ProcessId -eq $directId
            $Row.PID=$a.ProcessId;$Row.Started=$true
        }
        if($a.State -ceq 'Associated' -or $a.State -ceq 'Unknown' -or $KnownStarted){
            if($a.State -cne 'Associated'){if($null -ne $a.Error){Add-P2BFailure $failure $a.Error};G-Error 'Unknown or contradictory association'}
            try{if(-not $Process.HasExited){$kill=$true;$Row.Rescue=$true;$script:rescues++;$Process.Kill()}}catch{G-Error $_}
            try{$wait=$true;$Row.OwnedExit=$Process.WaitForExit(2000);if($Row.OwnedExit){$Row.ExitCode=$Process.ExitCode}}catch{G-Error $_}
        }
    }catch{G-Error $_}
    finally{
        try{$Process.Dispose()}catch{G-Error $_}
        $state=if($null -eq $a){'Unknown'}else{$a.State}
        $script:guards[$Row.Case]=[pscustomobject][ordered]@{State=$state;Started=$Row.Started;OwnedExit=$Row.OwnedExit;KillAttempted=$kill;WaitAttempted=$wait;PidMatches=$pidMatches;CleanupCount=$errors.Count}
        $script:associationDetails[$Row.Case]=[pscustomobject]@{ProcessId=$Row.PID;Errors=$errors.ToArray()}
        if($Row.Number -gt 0){
            if($state -cne 'Associated' -or -not $pidMatches -or -not $Row.OwnedExit -or $Row.Rescue -or $errors.Count -ne 0){$Row.Outcome='INCOMPLETE'}
            $path=Join-Path $packet 'children.json'
            $null=Complete-P2BLedger $failure $path ($Row.Number-1) $Row 'child'
        }
    }
}
function C-Observe($Result,$Process){
    $a=Read-P2BProcessAssociation $Process;$directId=$Process.Id
    $matches=$a.State -ceq 'Associated' -and $directId -is [int] -and $directId -gt 0 -and $a.ProcessId -eq $directId
    [pscustomobject][ordered]@{Started=$Result.Started;PrimaryPresent=$null -ne $Result.PrimaryError;CleanupCount=$Result.CleanupErrors.Count;TimedOut=$Result.TimedOut;TerminationRequested=$Result.TerminationRequested;OwnedExit=$Result.OwnedExit;StreamsComplete=$Result.StreamsComplete;ExitCodeKnown=$null -ne $Result.ExitCode;ActualHasExited=$Process.HasExited;CleanupOverflow=$Result.CleanupOverflow;AssociationMatches=$matches}
}
$passed=$false;$errorText=$null
try{
    $p=[Diagnostics.Process]::new()
    try{
        $a=Read-P2BProcessAssociation $p
        $observations.UnstartedAssociation=[pscustomobject][ordered]@{State=$a.State;ProcessId=$a.ProcessId;ErrorPresent=$null -ne $a.Error}
        if($null -ne $a.Error){$associationDetails.Unstarted=C-Describe $a.Error}
        C-Need ($a.State -ceq 'Unassociated' -and $null -eq $a.ProcessId -and $null -ne $a.Error) 'UnstartedAssociation'
        C-Pass 'UnstartedAssociation'
    }catch{Add-P2BFailure $failure $_;throw}finally{try{$p.Dispose()}catch{Add-P2BFailure $failure $_}}
    $nullObservation=Read-P2BProcessAssociation $null
    $similarObservation=Read-P2BProcessAssociation ([pscustomobject]@{Id=123;State='Associated'})
    $observations.UnknownAssociationRejected=[pscustomobject][ordered]@{NullUnknown=$nullObservation.State -ceq 'Unknown' -and $null -eq $nullObservation.ProcessId -and $null -eq $nullObservation.Error;SimilarUnknown=$similarObservation.State -ceq 'Unknown' -and $null -eq $similarObservation.ProcessId -and $null -eq $similarObservation.Error}
    C-Need ($observations.UnknownAssociationRejected.NullUnknown -and $observations.UnknownAssociationRejected.SimilarUnknown) 'UnknownAssociationRejected'
    C-Pass 'UnknownAssociationRejected'
    $caseDir=Join-Path $packet 'green-cases'
    if(Test-Path -LiteralPath $caseDir){throw 'Fresh case directory required'}
    $null=New-Item -ItemType Directory -Path $caseDir
    $badReceipt=Join-Path $caseDir 'receipt-is-directory';$null=New-Item -ItemType Directory -Path $badReceipt
    $row=C-Consume 'fault';$p=C-Process 'Hold';$r=$null
    try{
        $Error.Clear()
        $r=Invoke-P2BOwnedRun -Process $p -StdoutPath (Join-Path $caseDir 'fault.stdout.bin') -StderrPath (Join-Path $caseDir 'fault.stderr.bin') -ReceiptPath $badReceipt -MaxMs 2000
        $faultErrors=@($Error)
        $results.Fault=[pscustomobject]@{ExitCode=$r.ExitCode;ElapsedMs=$r.ElapsedMs}
        $observations.PostStartReceiptFault=C-Observe $r $p
        C-Need ($r.Started -and $null -ne $r.PrimaryError -and $r.OwnedExit -and $p.HasExited -and $r.TerminationRequested -and -not $r.TimedOut -and $r.StreamsComplete -and -not $r.CleanupOverflow) 'PostStartReceiptFault'
        C-Pass 'PostStartReceiptFault'
        $first=$null;$final=$null
        if($faultErrors.Count -eq 2){$first=$faultErrors[1];$final=$faultErrors[0]}
        $original=(C-PrimaryPreserved $r $first $final)
        $replaced=[pscustomobject]@{PrimaryError=$r.CleanupErrors[0];CleanupErrors=[Management.Automation.ErrorRecord[]]@($r.CleanupErrors[0])}
        $swapped=[pscustomobject]@{PrimaryError=$r.CleanupErrors[0];CleanupErrors=[Management.Automation.ErrorRecord[]]@($r.PrimaryError)}
        $cleanupReplaced=[pscustomobject]@{PrimaryError=$r.PrimaryError;CleanupErrors=[Management.Automation.ErrorRecord[]]@($r.PrimaryError)}
        $missing=[pscustomobject]@{CleanupErrors=[Management.Automation.ErrorRecord[]]@($r.CleanupErrors[0])}
        [Exception]$deep=$r.PrimaryError.Exception
        for($i=0;$i -lt 9;$i++){$deep=[Management.Automation.RuntimeException]::new('test wrapper',$deep)}
        $deepRecord=[Management.Automation.ErrorRecord]::new($deep,'CausalityDeep',[Management.Automation.ErrorCategory]::NotSpecified,$null)
        $o=[pscustomobject][ordered]@{ErrorCount=$faultErrors.Count;AnchorsDistinct=$null -ne $first -and $null -ne $final -and -not (C-SameIOCause $first $final);Original=$original;PrimaryReplaced=(C-PrimaryPreserved $replaced $first $final);Swapped=(C-PrimaryPreserved $swapped $first $final);CleanupReplaced=(C-PrimaryPreserved $cleanupReplaced $first $final);NullAnchor=(C-SameIOCause $r.PrimaryError $null);MissingPrimary=(C-PrimaryPreserved $missing $first $final);DeepNine=(C-SameIOCause $deepRecord $first);CleanupCount=$r.CleanupErrors.Count}
        $observations.PrimarySurvivesFinalReceiptFault=$o
        C-Need ($o.ErrorCount -eq 2 -and $o.AnchorsDistinct -and $o.Original -and -not $o.PrimaryReplaced -and -not $o.Swapped -and -not $o.CleanupReplaced -and -not $o.NullAnchor -and -not $o.MissingPrimary -and -not $o.DeepNine -and $o.CleanupCount -eq 1) 'PrimarySurvivesFinalReceiptFault'
        C-Pass 'PrimarySurvivesFinalReceiptFault';$row.Outcome='EXPECTED'
    }catch{Add-P2BFailure $failure $_;throw}finally{C-Finish $row $p ($null -ne $r -and $r.Started)}
    $row=C-Consume 'normal';$p=C-Process 'EmitAndExit';$r=$null
    try{
        $r=Invoke-P2BOwnedRun -Process $p -StdoutPath (Join-Path $caseDir 'normal.stdout.bin') -StderrPath (Join-Path $caseDir 'normal.stderr.bin') -ReceiptPath (Join-Path $caseDir 'normal.receipt.json') -MaxMs 2000
        $observations.NormalExit=C-Observe $r $p
        $results.Normal=[pscustomobject]@{ExitCode=$r.ExitCode;ElapsedMs=$r.ElapsedMs}
        C-Need ($r.Started -and $null -eq $r.PrimaryError -and $r.CleanupErrors.Count -eq 0 -and -not $r.CleanupOverflow -and -not $r.TimedOut -and -not $r.TerminationRequested -and $r.OwnedExit -and $r.StreamsComplete -and $p.HasExited -and $r.ExitCode -eq 0 -and [IO.File]::ReadAllText((Join-Path $caseDir 'normal.stdout.bin')) -ceq 'P2B-CONTAINMENT-OUT' -and [IO.File]::ReadAllText((Join-Path $caseDir 'normal.stderr.bin')) -ceq 'P2B-CONTAINMENT-ERR') 'NormalExit'
        C-Pass 'NormalExit';$row.Outcome='EXPECTED'
    }catch{Add-P2BFailure $failure $_;throw}finally{C-Finish $row $p ($null -ne $r -and $r.Started)}
    $row=C-Consume 'timeout';$p=C-Process 'Hold';$r=$null
    try{
        $r=Invoke-P2BOwnedRun -Process $p -StdoutPath (Join-Path $caseDir 'timeout.stdout.bin') -StderrPath (Join-Path $caseDir 'timeout.stderr.bin') -ReceiptPath (Join-Path $caseDir 'timeout.receipt.json') -MaxMs 200
        $observations.Timeout=C-Observe $r $p
        $results.Timeout=[pscustomobject]@{ExitCode=$r.ExitCode;ElapsedMs=$r.ElapsedMs}
        C-Need ($r.Started -and $r.TimedOut -and $r.TerminationRequested -and $r.OwnedExit -and $r.StreamsComplete -and $p.HasExited -and $null -ne $r.ExitCode -and $r.CleanupErrors.Count -eq 0 -and $null -eq $r.PrimaryError -and -not $r.CleanupOverflow) 'Timeout'
        C-Pass 'Timeout';$row.Outcome='EXPECTED'
    }catch{Add-P2BFailure $failure $_;throw}finally{C-Finish $row $p ($null -ne $r -and $r.Started)}
    $p=C-Process 'Hold'
    $row=[pscustomobject][ordered]@{Number=0;Case='prestart';Started=$false;PID=$null;OwnedExit=$false;ExitCode=$null;Rescue=$false}
    $r=$null
    try{
        $r=Invoke-P2BOwnedRun -Process $p -StdoutPath $badReceipt -StderrPath (Join-Path $caseDir 'prestart.stderr.bin') -ReceiptPath (Join-Path $caseDir 'prestart.receipt.json') -MaxMs 2000
        $a=Read-P2BProcessAssociation $p
        $observations.PreStartOutputFault=[pscustomobject][ordered]@{Started=$r.Started;PrimaryPresent=$null -ne $r.PrimaryError;Unassociated=$a.State -ceq 'Unassociated';ProcessId=$a.ProcessId;AssociationErrorPresent=$null -ne $a.Error;OwnedExit=$r.OwnedExit;ExitCode=$r.ExitCode;TerminationRequested=$r.TerminationRequested}
        if($null -ne $a.Error){$associationDetails.PreStart=C-Describe $a.Error}
        C-Need (-not $r.Started -and $null -ne $r.PrimaryError -and $a.State -ceq 'Unassociated' -and $null -eq $a.ProcessId -and $null -ne $a.Error -and -not $r.OwnedExit -and $null -eq $r.ExitCode -and -not $r.TerminationRequested) 'PreStartOutputFault'
    }catch{Add-P2BFailure $failure $_;throw}finally{C-Finish $row $p ($null -ne $r -and $r.Started)}
    $g=$guards.prestart
    C-Need ($g.State -ceq 'Unassociated' -and -not $g.Started -and -not $g.OwnedExit -and -not $g.KillAttempted -and -not $g.WaitAttempted -and $g.CleanupCount -eq 0) 'PreStart guard'
    C-Pass 'PreStartOutputFault'
    C-Need ($groups.Count -eq 7 -and $rescues -eq 0) 'GREEN completion'
    $passed=$true
}catch{Add-P2BFailure $failure $_;try{$errorText=$failure.PrimaryError.ToString()}catch{Add-P2BFailure $failure $_;$errorText='INCOMPLETE: error summary unavailable'}}
try{
$descriptions=@($faultErrors|Select-Object -First 8|ForEach-Object{C-Describe $_})
C-Save $DetailsPath ([pscustomobject][ordered]@{Schema=1;Mode='p2b-controller-causality-details';ErrorCount=$faultErrors.Count;RecordsNewestFirst=$descriptions;RecordsTruncated=$faultErrors.Count -gt 8;Results=[pscustomobject]$results;AssociationDetails=[pscustomobject]$associationDetails})
C-Save $TracePath ([pscustomobject][ordered]@{Schema=1;Mode='p2b-publication-containment';Groups=$groups.ToArray();Observations=[pscustomobject]$observations;Guards=[pscustomobject]$guards;Passed=$passed;Error=$errorText;Rescues=$rescues;CompileCount=0;ProductChanged=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false;Bookkeeping=[pscustomobject][ordered]@{PrimaryAbsent=$null -eq $failure.PrimaryError;SecondaryCount=$failure.SecondaryErrors.Count;Overflow=$failure.SecondaryOverflow;FinalLedgerWritten=$failure.FinalLedgerWritten}})
}catch{Add-P2BFailure $failure $_}
if($null -ne $failure.PrimaryError -or -not $passed){Write-P2BFailureFallback $failure;exit 2}
exit 0
