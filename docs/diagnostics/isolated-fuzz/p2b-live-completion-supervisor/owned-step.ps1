# Baseline deliberately preserves the old final receipt retry for owned-red.
# Do not call the frozen legacy helper. Only this file changes after owned-red.
function Invoke-LSOwnedStep($SessionEntry,$StepDescriptor){
    $e=$SessionEntry;$d=$StepDescriptor
    $r=[pscustomobject][ordered]@{Started=$false;Association='Unknown';PID=$null;OwnedExit=$false;ExitCode=$null;StreamsComplete=$false;TimedOut=$false;TerminationRequested=$false;ElapsedMs=[long]0;KillCalls=0;WaitCalls=0;SeamExact=$false;Stdout=$null}
    $process=[Diagnostics.Process]::new()
    $process.StartInfo=[Diagnostics.ProcessStartInfo]::new()
    $process.StartInfo.FileName=$script:Engine
    $process.StartInfo.UseShellExecute=$false
    $process.StartInfo.CreateNoWindow=$true
    $process.StartInfo.RedirectStandardOutput=$true
    $process.StartInfo.RedirectStandardError=$true
    $process.StartInfo.WorkingDirectory=$script:Repo
    foreach($arg in @('-NoLogo','-NoProfile','-NonInteractive','-File',$d.Script)){$process.StartInfo.ArgumentList.Add($arg)}
    foreach($arg in $d.Arguments){$process.StartInfo.ArgumentList.Add($arg)}
    $timer=[Diagnostics.Stopwatch]::StartNew()
    $stdout=$null;$stderr=$null;$ot=$null;$et=$null
    try{
        $stdout=Open-LSStream $e $d.Out
        if($null -eq $stdout){return $r}
        $stderr=Open-LSStream $e $d.Err
        if($null -eq $stderr){return $r}
        $remaining=[Math]::Min($d.MaxMs,([int]140000-[int]$script:Clock.ElapsedMilliseconds-4000))
        if($remaining -le 0){throw 'No remaining whole-sequence launch budget'}
        $e.StartCalls++;$script:HostStarts++
        if(-not $process.Start()){throw 'Owned target did not start'}
        $r.Started=$true
        $a=Read-P2BProcessAssociation $process
        $r.Association=$a.State;$r.PID=$a.ProcessId
        if($a.State -cne 'Associated'){throw 'Owned process association unknown'}
        $e.Sinks[$d.Out].LogicalIOAttempts++
        $ot=$process.StandardOutput.BaseStream.CopyToAsync($stdout)
        $e.Sinks[$d.Err].LogicalIOAttempts++
        $et=$process.StandardError.BaseStream.CopyToAsync($stderr)
        $receipt=Write-LSLedger $e.Handle $d.Receipt ([ordered]@{Started=$true;CandidateOnly=$true;PID=$r.PID})
        if($receipt.Ok){
            $wait=[Math]::Max(0,[Math]::Min(($d.MaxMs-[int]$timer.ElapsedMilliseconds),(140000-[int]$script:Clock.ElapsedMilliseconds-4000)))
            $r.WaitCalls++
            if(-not $process.WaitForExit([int]$wait)){$r.TimedOut=$true}
        }
    }catch{Add-LSFailure $e.Handle $_}
    finally{
        # Cleanup depends only on the still-held exact Process, never ledger IO.
        $association=Read-P2BProcessAssociation $process
        if($association.State -ceq 'Associated'){
            $r.Association=$association.State;$r.PID=$association.ProcessId
            $exited=$false
            try{$exited=$process.HasExited}catch{Add-LSFailure $e.Handle $_}
            if(-not $exited){
                $r.TerminationRequested=$true;$r.KillCalls++
                try{$process.Kill()}catch{Add-LSFailure $e.Handle $_}
            }
            try{$r.WaitCalls++;$r.OwnedExit=$process.WaitForExit(2000);if($r.OwnedExit){$r.ExitCode=$process.ExitCode}}
            catch{Add-LSFailure $e.Handle $_}
        }elseif($association.State -ceq 'Unassociated'){$r.Association='Unassociated';$r.PID=$null}
        else{if($null -ne $association.Error){Add-LSFailure $e.Handle $association.Error}}
        if($null -ne $ot -and $null -ne $et){
            try{$r.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),2000)}
            catch{$e.Sinks[$d.Out].State='Failed';$e.Sinks[$d.Err].State='Failed';Add-LSFailure $e.Handle $_}
        }
        Close-LSStream $e $d.Out
        Close-LSStream $e $d.Err
        try{$process.Dispose()}catch{Add-LSFailure $e.Handle $_}
        $timer.Stop();$r.ElapsedMs=$timer.ElapsedMilliseconds
        # RED extraction: actual second attempt bypasses the repaired core gate.
        # Its actual second ErrorRecord is retained, not manufactured.
        if($r.Started){
            $receiptSink=$e.Sinks[$d.Receipt]
            $receiptSink.LogicalIOAttempts++
            try{[IO.File]::WriteAllText($receiptSink.Path,(ConvertTo-Json -InputObject $r -Depth 10),[Text.UTF8Encoding]::new($false))}
            catch{Fail-LSIO $e $receiptSink $_}
        }
    }
    return $r
}
