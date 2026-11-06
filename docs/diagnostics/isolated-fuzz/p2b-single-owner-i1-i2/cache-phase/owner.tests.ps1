param([Parameter(Mandatory)][ValidateSet('OwnerRed','OwnerGreen')][string]$Phase)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot '../runtime-cache-wait-common.ps1')
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../../..'))
$root=New-C12WaitRoot (Join-Path $repo ".superpowers/sdd/2026-11-06-c12-cache-phase-v1/$Phase")
$results=[Collections.Generic.List[object]]::new()
function Check($Condition,$Message) {if(-not $Condition){throw $Message}}
function Reject($Body,$Pattern) { $errorText=$null;try{& $Body}catch{$errorText=$_.Exception.Message};Check ($null -ne $errorText -and $errorText -match $Pattern) "Expected $Pattern;actual $errorText" }
function Case($Name,$Body) {& $Body;$results.Add(@{Name=$Name;Status='PASS'})}
function Fixture($Name,$Text) { $path=Join-Path $root "$Name.jsonl";[IO.File]::WriteAllText($path,$Text,[Text.UTF8Encoding]::new($false));return $path }
$session='a'*32
$start=@{Schema=1;Sequence=1;Session=$session;Module='golang.org/x/text';Version='v0.40.0';Phase='PREPARE';Kind='START';Status='STARTED';Digest='';ElapsedNs=0}
try {
    $owner=Join-Path $PSScriptRoot 'owner.ps1'
    if(-not(Test-Path -LiteralPath $owner)){throw 'MISSING_OWNER_IMPLEMENTATION'}
    . $owner
    Case 'sequence-session-binding' {
        foreach($field in @('Sequence','Session')) { $bad=$start.Clone();$bad[$field]=$(if($field -eq 'Sequence'){2}else{'b'*32});$path=Fixture $field (($bad|ConvertTo-Json -Compress)+"`n");Reject {Read-C12PhaseEvents -Path $path -Session $session} 'BINDING' }
    }
    Case 'partial-line-not-complete' {
        $path=Fixture 'partial' (($start|ConvertTo-Json -Compress)+"`n"+'{')
        $r=Read-C12PhaseEvents -Path $path -Session $session
        Check ($r.Partial -and -not $r.Complete -and $r.Phases[0].Status -ceq 'Incomplete') 'Partial accepted as complete'
    }
    Case 'start-without-end' {
        $r=Read-C12PhaseEvents -Path (Fixture 'start' (($start|ConvertTo-Json -Compress)+"`n")) -Session $session
        Check (-not $r.Complete -and $r.Phases[0].Status -ceq 'Incomplete' -and $r.Phases[1].Status -ceq 'NotRun') 'Lost incomplete phase'
    }
    Case 'illegal-phase-order' {
        $bad=$start.Clone();$bad.Phase='ZIP';Reject {Read-C12PhaseEvents -Path (Fixture 'order' (($bad|ConvertTo-Json -Compress)+"`n")) -Session $session} 'ORDER'
    }
    Case 'event-size-bound' {Reject {Read-C12PhaseEvents -Path (Fixture 'large' ('x'*65537)) -Session $session} 'EVENT_LIMIT'}
    Case 'total-stream-bound' {
        $path=Fixture 'total' ''; $f=[IO.File]::OpenWrite($path);try{$f.SetLength(4194305)}finally{$f.Dispose()}
        Reject {Read-C12PhaseEvents -Path $path -Session $session} 'TOTAL_LIMIT'
    }
    Case 'occupied-output-preserved' {
        $sentinelRoot=New-C12WaitRoot (Join-Path $root 'sentinel')
        $record=Write-C12WaitRecord $sentinelRoot 'sentinel.json' @{Value='original'}
        Reject {Write-C12WaitRecord $sentinelRoot 'sentinel.json' @{Value='changed'}} 'OCCUPIED'
        Check ((Get-FileHash -LiteralPath $record.Path).Hash -ceq $record.SHA256) 'Sentinel changed'
    }
    Case 'deadline-boundary' {
        Check ((Get-C12WaitDecision 1000 1000 $false $true) -ceq 'STOP') 'Deadline extended'
        Check ((Get-C12WaitDecision 999 1000 $false $false) -ceq 'WAIT') 'Premature start'
    }
    $pwsh=[Diagnostics.Process]::GetCurrentProcess().MainModule.FileName
    Case 'owned-normal-exit' {
        $r=Invoke-C12PhaseOwned -Executable $pwsh -Arguments @('-NoProfile','-NonInteractive','-Command','Write-Output ''fixture''; exit 0') -WorkingDirectory $repo -Environment @{} -Root (New-C12WaitRoot (Join-Path $root 'normal')) -DeadlineMs 5000
        Check ($r.Started -and $r.ExitConfirmed -and -not $r.TimedOut -and $r.ExitCode -eq 0 -and $r.EndReason -ceq 'EXITED') 'Normal lifecycle failed'
        Check ([IO.File]::ReadAllText($r.StdoutPath).Contains('fixture')) 'Output lost'
    }
    Case 'owned-1000ms-cutoff' {
        $r=Invoke-C12PhaseOwned -Executable $pwsh -Arguments @('-NoProfile','-NonInteractive','-Command','Start-Sleep -Seconds 30') -WorkingDirectory $repo -Environment @{} -Root (New-C12WaitRoot (Join-Path $root 'cutoff')) -DeadlineMs 1000
        Check ($r.Started -and $r.TimedOut -and $r.ExitConfirmed -and $r.EndReason -ceq 'DEADLINE' -and $r.ElapsedMs -lt 7000) 'Cutoff failed'
    }
    if($Phase -eq 'OwnerRed'){throw 'UNEXPECTED_RED_PASS'}
    $receipt=@{Phase=$Phase;Status='OWNER_TESTS_VERIFIED';Count=$results.Count;Cases=$results.ToArray()}
    # Quota fixtures intentionally occupy >4MiB; test summary is separate fixture output.
    [IO.File]::WriteAllText((Join-Path $root 'result.json'),($receipt|ConvertTo-Json -Depth 10))
    $receipt|ConvertTo-Json -Depth 10
} catch {
    [IO.File]::WriteAllText((Join-Path $root 'result.json'),(@{Phase=$Phase;Status='STOP';Error=$_.Exception.Message;Cases=$results.ToArray()}|ConvertTo-Json -Depth 10))
    throw
}
