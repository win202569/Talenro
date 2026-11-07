param([Parameter(Mandatory)][ValidateSet('Red','Green')][string]$Phase)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'owner.ps1')
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../../..'))
$root=New-C12WaitRoot (Join-Path $repo ".superpowers/sdd/2026-11-07-c12-phase-stream-fix/$Phase")
$results=[Collections.Generic.List[object]]::new()
$session='a'*32
$digest='h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU='
$events=[Collections.Generic.List[object]]::new()
foreach($id in @(@('golang.org/x/text','v0.40.0'),@('modernc.org/sqlite','v1.49.1'))){
    foreach($p in @('PREPARE','ZIP','DIRECTORY')){
        foreach($kind in @('START','END')){
            $events.Add(@{Schema=1;Sequence=$events.Count+1;Session=$session;Module=$id[0];Version=$id[1];Phase=$p;Kind=$kind;Status=$(if($kind -eq 'START'){'STARTED'}else{'OK'});Digest=$(if($kind -eq 'END' -and $p -ne 'PREPARE'){$digest}else{''});ElapsedNs=0})
        }
    }
}
$lines=@($events|ForEach-Object{($_|ConvertTo-Json -Compress)+"`n"})
function Check($Condition,$Message){if(-not $Condition){throw $Message}}
function Case($Name,$Body){& $Body;$results.Add(@{Name=$Name;Status='PASS'})}
function Reject($Body,$Code){$message=$null;try{& $Body}catch{$message=$_.Exception.Message};Check ($null -ne $message -and $message.Contains($Code)) "Expected $Code; got $message"}
function File($Name,$Text){$p=Join-Path $root $Name;[IO.File]::WriteAllText($p,$Text,[Text.UTF8Encoding]::new($false));return $p}
function Append($Writer,$Text){$b=[Text.Encoding]::UTF8.GetBytes($Text);$Writer.Write($b,0,$b.Length);$Writer.Flush()}
function Owned($Name,$Mode,$InputPath,$Deadline){
    $pwsh=[Diagnostics.Process]::GetCurrentProcess().MainModule.FileName
    return Invoke-C12PhaseOwned -Executable $pwsh -Arguments @('-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot 'stream-fixture.ps1'),'-InputPath',$InputPath,'-Mode',$Mode) -WorkingDirectory $repo -Environment @{} -Root (New-C12WaitRoot (Join-Path $root $Name)) -DeadlineMs $Deadline -Session $session
}
# Retain the exact implementation exercised at each stage,including RED.
[IO.File]::WriteAllBytes((Join-Path $root 'owner-source.ps1'),[IO.File]::ReadAllBytes((Join-Path $PSScriptRoot 'owner.ps1')))
try{
    Case 'open-writer-readable' {
        $p=Join-Path $root 'open.jsonl';$w=[IO.File]::Open($p,'CreateNew','Write','Read')
        try{
            Append $w $lines[0]
            $a=Read-C12PhaseEvents -Path $p -Session $session
            Check ($a.Events.Count -eq 1 -and $a.Phases[0].Status -eq 'Incomplete' -and -not $a.Complete) 'Open START lost'
            # Reader must not close/change writer or permit a second writer.
            Reject { $other=[IO.File]::Open($p,'Open','Write','ReadWrite');$other.Dispose() } 'used by another process'
            Append $w $lines[1]
            $a=Read-C12PhaseEvents -Path $p -Session $session
            Check ($a.Events.Count -eq 2 -and $a.Phases[0].Status -eq 'Complete' -and $a.Phases[1].Status -eq 'NotRun') 'Append not visible'
        }finally{$w.Dispose()}
        $a=Read-C12PhaseEvents -Path $p -Session $session
        Check ($a.Events.Count -eq 2) 'Closed regression'
    }
    Case 'partial-then-complete' {
        $p=Join-Path $root 'partial.jsonl';$w=[IO.File]::Open($p,'CreateNew','Write','Read')
        try{
            Append $w ($lines[0]+$lines[1].Substring(0,15))
            $a=Read-C12PhaseEvents -Path $p -Session $session
            Check ($a.Partial -and $a.Events.Count -eq 1 -and $a.Phases[0].Status -eq 'Incomplete') 'Partial accepted'
            Append $w $lines[1].Substring(15)
            $a=Read-C12PhaseEvents -Path $p -Session $session
            Check (-not $a.Partial -and $a.Events.Count -eq 2) 'Partial not completed'
        }finally{$w.Dispose()}
    }
    Case 'binding-rejected' {
        $bad=$events[0].Clone();$bad.Session='b'*32
        Reject {Read-C12PhaseEvents -Path (File 'binding.jsonl' (($bad|ConvertTo-Json -Compress)+"`n")) -Session $session} 'BINDING'
    }
    Case 'order-rejected' {
        $bad=$events[0].Clone();$bad.Phase='ZIP'
        Reject {Read-C12PhaseEvents -Path (File 'order.jsonl' (($bad|ConvertTo-Json -Compress)+"`n")) -Session $session} 'ORDER'
    }
    Case 'missing-digest-rejected' {
        $bad=$events[3].Clone();$bad.Digest=''
        Reject {Read-C12PhaseEvents -Path (File 'digest.jsonl' (($lines[0..2]-join '')+($bad|ConvertTo-Json -Compress)+"`n")) -Session $session} 'EVENT_DIGEST'
    }
    Case 'event-cap' {Reject {Read-C12PhaseEvents -Path (File 'event-cap.jsonl' ('x'*65537)) -Session $session} 'EVENT_LIMIT'}
    Case 'total-cap' {
        $p=File 'total-cap.jsonl' '';$w=[IO.File]::Open($p,'Open','Write','Read')
        try{$w.SetLength(4194305);$w.Flush();Reject {Read-C12PhaseEvents -Path $p -Session $session} 'TOTAL_LIMIT'}finally{$w.Dispose()}
    }
    if($Phase -eq 'Red'){throw 'UNEXPECTED_RED_PASS'}
    Case 'owned-stream-complete' {
        $p=File 'complete-input.jsonl' ($lines-join '')
        $a=Owned 'complete' 'complete' $p 5000
        Check ($a.Started -and $a.ExitConfirmed -and $a.ExitCode -eq 0 -and -not $a.TimedOut -and $null -eq $a.Failure) "Completion failed: $($a.Failure)"
        $parsed=Read-C12PhaseEvents -Path $a.StdoutPath -Session $session
        Check ($parsed.Complete -and $parsed.Events.Count -eq 12 -and -not $parsed.Partial) 'Stream not complete'
        Check ((Get-FileHash $p).Hash -ceq (Get-FileHash $a.StdoutPath).Hash) 'Stream bytes changed'
    }
    Case 'owned-protocol-stop' {
        $bad=$events[0].Clone();$bad.Session='b'*32
        $p=File 'bad-input.jsonl' (($bad|ConvertTo-Json -Compress)+"`n")
        $a=Owned 'bad' 'bad' $p 5000
        Check ($a.Started -and $a.ExitConfirmed -and -not $a.TimedOut -and $a.EndReason -eq 'ERROR' -and $a.Failure -match 'BINDING') "Protocol stop failed: $($a.Failure)"
    }
    Case 'owned-stream-cutoff' {
        $p=File 'cutoff-input.jsonl' $lines[0]
        $a=Owned 'cutoff' 'cutoff' $p 3000
        Check ($a.Started -and $a.ExitConfirmed -and $a.TimedOut -and $a.EndReason -eq 'DEADLINE' -and $null -eq $a.Failure -and $a.ElapsedMs -lt 9000) "Cutoff failed: $($a.Failure)"
        $parsed=Read-C12PhaseEvents -Path $a.StdoutPath -Session $session
        Check ($parsed.Events.Count -eq 1 -and -not $parsed.Complete -and $parsed.Phases[0].Status -eq 'Incomplete' -and $parsed.Phases[1].Status -eq 'NotRun') 'Cutoff evidence lost'
    }
    $receipt=@{Phase=$Phase;Status='PASS';Count=$results.Count;Cases=$results.ToArray()}
    [IO.File]::WriteAllText((Join-Path $root 'result.json'),($receipt|ConvertTo-Json -Depth 10))
    $receipt|ConvertTo-Json -Depth 10
}catch{
    [IO.File]::WriteAllText((Join-Path $root 'result.json'),(@{Phase=$Phase;Status='STOP';Error=$_.Exception.Message;Cases=$results.ToArray()}|ConvertTo-Json -Depth 10))
    throw
}
