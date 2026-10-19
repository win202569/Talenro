function Invoke-P2BOwnedRun {
    param([Parameter(Mandatory)][Diagnostics.Process]$Process,
    [Parameter(Mandatory)][string]$StdoutPath,[Parameter(Mandatory)][string]$StderrPath,
    [Parameter(Mandatory)][string]$ReceiptPath,[ValidateRange(1,120000)][int]$MaxMs=120000,
    [ValidateRange(1,2000)][int]$ExitWaitMs=2000,[ValidateRange(1,2000)][int]$DrainMs=2000)
    $r=[pscustomobject][ordered]@{Started=$false;PrimaryError=$null;CleanupErrors=[Management.Automation.ErrorRecord[]]@();TimedOut=$false;TerminationRequested=$false;OwnedExit=$false;StreamsComplete=$false;ExitCode=$null;ElapsedMs=[long]0;CleanupOverflow=$false}
    $cleanup=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new()
    $stdout=$null;$stderr=$null;$ot=$null;$et=$null
    $timer=[Diagnostics.Stopwatch]::StartNew()
    function Add-Cleanup([Management.Automation.ErrorRecord]$Record){
        if($cleanup.Count -lt 8){$cleanup.Add($Record)}else{$r.CleanupOverflow=$true}
    }
    function Save-Receipt{
        $value=[ordered]@{Started=$r.Started;TimedOut=$r.TimedOut;TerminationRequested=$r.TerminationRequested;OwnedExit=$r.OwnedExit;StreamsComplete=$r.StreamsComplete;ExitCode=$r.ExitCode;ElapsedMs=$timer.ElapsedMilliseconds;PrimaryError=if($null -ne $r.PrimaryError){$r.PrimaryError.ToString()}else{$null};CleanupCount=$cleanup.Count;CleanupOverflow=$r.CleanupOverflow}
        [IO.File]::WriteAllText($ReceiptPath,($value|ConvertTo-Json -Depth 5),[Text.UTF8Encoding]::new($false))
    }
    # Attach both drains before persistence. All post-Start exits converge on
    # the same bounded cleanup while the caller still owns this exact Process.
    try{
        $stdout=[IO.File]::Create($StdoutPath);$stderr=[IO.File]::Create($StderrPath)
        if(-not $Process.Start()){throw 'Owned target did not start'}
        $r.Started=$true
        $ot=$Process.StandardOutput.BaseStream.CopyToAsync($stdout)
        $et=$Process.StandardError.BaseStream.CopyToAsync($stderr)
        Save-Receipt
        $remaining=[Math]::Max(0,($MaxMs-[int]$timer.ElapsedMilliseconds))
        if(-not $Process.WaitForExit($remaining)){$r.TimedOut=$true}
    }catch{$r.PrimaryError=$_}
    finally{
        if($r.Started){
            $alreadyExited=$false
            try{$alreadyExited=$Process.HasExited}catch{Add-Cleanup $_}
            if(-not $alreadyExited){
                $r.TerminationRequested=$true
                try{$Process.Kill()}catch{Add-Cleanup $_}
            }
            try{
                $r.OwnedExit=$Process.WaitForExit($ExitWaitMs)
                if($r.OwnedExit){$r.ExitCode=$Process.ExitCode}
            }catch{Add-Cleanup $_}
        }
        if($null -ne $ot -and $null -ne $et){
            try{$r.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),$DrainMs)}catch{Add-Cleanup $_}
        }
        if($null -ne $stdout){try{$stdout.Dispose()}catch{Add-Cleanup $_}}
        if($null -ne $stderr){try{$stderr.Dispose()}catch{Add-Cleanup $_}}
        $timer.Stop();$r.ElapsedMs=$timer.ElapsedMilliseconds
        # Secondary receipt errors never replace the original ErrorRecord and
        # cannot skip the owned-exit and stream cleanup already performed.
        try{Save-Receipt}catch{Add-Cleanup $_}
    }
    $r.CleanupErrors=$cleanup.ToArray()
    return $r
}
