param([Parameter(Mandatory)][ValidateSet('RED','GREEN')][string]$Phase,
[Parameter(Mandatory)][string]$TracePath)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-19-p2b-controller-exception-containment'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
if((Get-FileHash -LiteralPath $engine).Hash -cne '362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'){throw 'Engine identity'}
$id=[Security.Principal.WindowsIdentity]::GetCurrent()
if(([Security.Principal.WindowsPrincipal]::new($id)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary test host required'}
. (Join-Path $PSScriptRoot 'owned-run.ps1')
$groups=[Collections.Generic.List[string]]::new()
$observations=[ordered]@{}
$rescues=0
function C-Need([bool]$Condition,[string]$Name){if(-not $Condition){throw ('ASSERTION FAILED: '+$Name)}}
function C-Save($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 20),[Text.UTF8Encoding]::new($false))}
function C-Pass([string]$Name){$groups.Add($Name);[Console]::Out.WriteLine('PASS: '+$Name)}
function C-Process([string]$Mode){
    $p=[Diagnostics.Process]::new()
    $p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
    $p.StartInfo.FileName=$engine
    $p.StartInfo.UseShellExecute=$false
    $p.StartInfo.CreateNoWindow=$true
    $p.StartInfo.RedirectStandardOutput=$true
    $p.StartInfo.RedirectStandardError=$true
    foreach($a in @('-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot 'fixture-child.ps1'),'-Mode',$Mode)){$p.StartInfo.ArgumentList.Add($a)}
    return $p
}
function C-Consume([string]$Case){
    $path=Join-Path $packet 'children.json'
    $rows=@();if(Test-Path -LiteralPath $path){$rows=@(Get-Content -LiteralPath $path -Raw|ConvertFrom-Json)}
    $order=@('red-fault','green-fault','green-normal','green-timeout','final-seam')
    if($rows.Count -ge 4 -or $Case -cne $order[$rows.Count]){throw 'Ordered child budget'}
    $row=[pscustomobject][ordered]@{Number=($rows.Count+1);Case=$Case;AttemptConsumed=$true;Started=$false;PID=$null;OwnedExit=$false;Rescue=$false;Outcome='INCOMPLETE'}
    $rows+=,$row;C-Save $path $rows
    return $row
}
function C-ChildFinish($Row,$Process,[bool]$Exit,[bool]$Rescue,[string]$Outcome){
    $path=Join-Path $packet 'children.json'
    $rows=@(Get-Content -LiteralPath $path -Raw|ConvertFrom-Json)
    $Row.Started=$true;$Row.PID=$Process.Id;$Row.OwnedExit=$Exit;$Row.Rescue=$Rescue;$Row.Outcome=$Outcome
    $rows[$Row.Number-1]=$Row;C-Save $path $rows
}
function C-Observe($Result,$Process){
    [pscustomobject][ordered]@{
        Started=$Result.Started;PrimaryPresent=$null -ne $Result.PrimaryError
        CleanupCount=$Result.CleanupErrors.Count;TimedOut=$Result.TimedOut
        TerminationRequested=$Result.TerminationRequested;OwnedExit=$Result.OwnedExit
        StreamsComplete=$Result.StreamsComplete;ExitCode=$Result.ExitCode
        ActualHasExited=$Process.HasExited;CleanupOverflow=$Result.CleanupOverflow
    }
}
$passed=$false;$errorText=$null
try{
    $caseDir=Join-Path $packet ($Phase.ToLowerInvariant()+'-cases')
    if(Test-Path -LiteralPath $caseDir){throw 'Fresh case directory required'}
    $null=New-Item -ItemType Directory -Path $caseDir
    $badReceipt=Join-Path $caseDir 'receipt-is-directory'
    $null=New-Item -ItemType Directory -Path $badReceipt
    $p=C-Process 'Hold';$row=C-Consume ($Phase.ToLowerInvariant()+'-fault');$r=$null
    try{
        $Error.Clear()
        $r=Invoke-P2BOwnedRun -Process $p -StdoutPath (Join-Path $caseDir 'fault.stdout.bin') -StderrPath (Join-Path $caseDir 'fault.stderr.bin') -ReceiptPath $badReceipt -MaxMs 2000
        $firstRecord=if($Error.Count -gt 0){$Error[$Error.Count-1]}else{$null}
        $observations.PostStartReceiptFault=C-Observe $r $p
        if($Phase -ceq 'RED'){
            C-Need ($r.Started -and $null -ne $r.PrimaryError -and -not $r.OwnedExit -and -not $p.HasExited) 'RED root cause prerequisites'
            $observations.PostStartReceiptFault|Add-Member -NotePropertyName FailedAssertion -NotePropertyValue 'PostStartReceiptFault.OwnedExit'
            C-Save (Join-Path $packet 'red.before-rescue.json') $observations.PostStartReceiptFault
            [Console]::Out.WriteLine('FAIL: PostStartReceiptFault.OwnedExit')
            $p.Kill();$rescues++
            $exit=$p.WaitForExit(2000)
            C-ChildFinish $row $p $exit $true 'EXPECTED_RED_RESCUED'
            C-Need $exit 'RED exact-owned rescue exit'
            [Console]::Out.WriteLine('RESCUE: exact-owned exit confirmed')
        }else{
            C-Need ($r.Started -and $null -ne $r.PrimaryError -and $r.OwnedExit -and $p.HasExited -and $r.TerminationRequested -and -not $r.TimedOut -and $r.StreamsComplete -and -not $r.CleanupOverflow) 'PostStartReceiptFault'
            C-ChildFinish $row $p $true $false 'FAULT_CONTAINED'
            C-Pass 'PostStartReceiptFault'
            $same=[object]::ReferenceEquals($r.PrimaryError,$firstRecord)
            $distinct=$r.CleanupErrors.Count -eq 1 -and -not [object]::ReferenceEquals($r.PrimaryError,$r.CleanupErrors[0])
            $observations.PrimarySurvivesFinalReceiptFault=[pscustomobject][ordered]@{FirstRecordIdentity=$same;DistinctFinalError=$distinct;CleanupCount=$r.CleanupErrors.Count}
            C-Need ($same -and $distinct) 'PrimarySurvivesFinalReceiptFault'
            C-Pass 'PrimarySurvivesFinalReceiptFault'
        }
    }finally{
        if($row.Started -or ($null -ne $r -and $r.Started)){
            if(-not $p.HasExited){
                $p.Kill();$rescues++
                $exit=$p.WaitForExit(2000)
                C-ChildFinish $row $p $exit $true 'UNEXPECTED_RESCUE'
                if(-not $exit){throw 'Exact-owned rescue failed'}
            }
        }
        $p.Dispose()
    }
    if($Phase -ceq 'GREEN'){
        $p=C-Process 'EmitAndExit';$row=C-Consume 'green-normal';$r=$null
        try{
            $r=Invoke-P2BOwnedRun -Process $p -StdoutPath (Join-Path $caseDir 'normal.stdout.bin') -StderrPath (Join-Path $caseDir 'normal.stderr.bin') -ReceiptPath (Join-Path $caseDir 'normal.receipt.json') -MaxMs 2000
            $observations.NormalExit=C-Observe $r $p
            C-Need ($r.Started -and $null -eq $r.PrimaryError -and $r.CleanupErrors.Count -eq 0 -and -not $r.TimedOut -and -not $r.TerminationRequested -and $r.OwnedExit -and $r.StreamsComplete -and $p.HasExited -and $r.ExitCode -eq 0 -and [IO.File]::ReadAllText((Join-Path $caseDir 'normal.stdout.bin')) -ceq 'P2B-CONTAINMENT-OUT' -and [IO.File]::ReadAllText((Join-Path $caseDir 'normal.stderr.bin')) -ceq 'P2B-CONTAINMENT-ERR') 'NormalExit'
            C-ChildFinish $row $p $true $false 'NORMAL_EXIT'
            C-Pass 'NormalExit'
        }finally{
            if($null -ne $r -and $r.Started -and -not $p.HasExited){$p.Kill();$rescues++;$exit=$p.WaitForExit(2000);C-ChildFinish $row $p $exit $true 'UNEXPECTED_RESCUE';if(-not $exit){throw 'Normal rescue failed'}}
            $p.Dispose()
        }
        $p=C-Process 'Hold';$row=C-Consume 'green-timeout';$r=$null
        try{
            $r=Invoke-P2BOwnedRun -Process $p -StdoutPath (Join-Path $caseDir 'timeout.stdout.bin') -StderrPath (Join-Path $caseDir 'timeout.stderr.bin') -ReceiptPath (Join-Path $caseDir 'timeout.receipt.json') -MaxMs 200
            $observations.Timeout=C-Observe $r $p
            C-Need ($r.Started -and $r.TimedOut -and $r.TerminationRequested -and $r.OwnedExit -and $r.StreamsComplete -and $p.HasExited -and $null -ne $r.ExitCode -and $r.CleanupErrors.Count -eq 0 -and $null -eq $r.PrimaryError) 'Timeout'
            C-ChildFinish $row $p $true $false 'TIMEOUT_CONTAINED'
            C-Pass 'Timeout'
        }finally{
            if($null -ne $r -and $r.Started -and -not $p.HasExited){$p.Kill();$rescues++;$exit=$p.WaitForExit(2000);C-ChildFinish $row $p $exit $true 'UNEXPECTED_RESCUE';if(-not $exit){throw 'Timeout rescue failed'}}
            $p.Dispose()
        }
        $p=C-Process 'Hold'
        try{
            $r=Invoke-P2BOwnedRun -Process $p -StdoutPath $badReceipt -StderrPath (Join-Path $caseDir 'prestart.stderr.bin') -ReceiptPath (Join-Path $caseDir 'prestart.receipt.json') -MaxMs 2000
            $unassociated=$false
            try{$null=$p.Id}catch{$unassociated=$true}
            $observations.PreStartOutputFault=[pscustomobject][ordered]@{Started=$r.Started;PrimaryPresent=$null -ne $r.PrimaryError;Unassociated=$unassociated;OwnedExit=$r.OwnedExit;ExitCode=$r.ExitCode}
            C-Need (-not $r.Started -and $null -ne $r.PrimaryError -and $unassociated -and -not $r.OwnedExit -and $null -eq $r.ExitCode) 'PreStartOutputFault'
            C-Pass 'PreStartOutputFault'
        }finally{$p.Dispose()}
        C-Need ($groups.Count -eq 5 -and $rescues -eq 0) 'GREEN completion'
        $passed=$true
    }
}catch{$errorText=$_.ToString()}
$trace=[pscustomobject][ordered]@{Schema=1;Phase=$Phase;Groups=$groups.ToArray();Observations=[pscustomobject]$observations;Passed=$passed;Error=$errorText;Rescues=$rescues;CompileCount=0;ProductChanged=$false;RealGateImplemented=$false;RuntimeApproved=$false;FormalAcceptance=$false}
C-Save $TracePath $trace
if($null -ne $errorText){[Console]::Error.WriteLine($errorText);exit 2}
if($Phase -ceq 'RED'){exit 1}
exit 0
