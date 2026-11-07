# Definition-only. Reuse immutable publication and path checks; no launch on import.
. (Join-Path $PSScriptRoot '../runtime-cache-wait-common.ps1')
function Get-C12PhaseProfile([string]$Profile='paired-v1') {
    if($Profile -ceq 'paired-v1'){
        return @(@{Module='golang.org/x/text';Version='v0.40.0'},@{Module='modernc.org/sqlite';Version='v1.49.1'})
    }
    if($Profile -ceq 'libc-v1'){return @{Module='modernc.org/libc';Version='v1.72.1'}}
    throw 'PROFILE'
}
function Read-C12PhaseEvents([string]$Path,[string]$Session,[string]$Profile='paired-v1') {
    $candidates=@(Get-C12PhaseProfile $Profile)
    Assert-C12WaitPath $Path
    if((Get-Item -LiteralPath $Path).Length -gt 4194304){throw 'TOTAL_LIMIT'}
    # Acknowledge the owner's existing writer while opening read-only. The writer
    # still shares Read only, so this does not allow a second writer or deletion.
    $reader=[IO.File]::Open($Path,'Open','Read','ReadWrite')
    try {
        $length=$reader.Length
        if($length -gt 4194304){throw 'TOTAL_LIMIT'}
        $bytes=[byte[]]::new([int]$length);$offset=0
        # Snapshot the current length; never chase a growing stream past the cap.
        while($offset -lt $bytes.Length){
            $n=$reader.Read($bytes,$offset,$bytes.Length-$offset)
            if($n -eq 0){throw 'STREAM_TRUNCATED'}
            $offset+=$n
        }
    }finally{$reader.Dispose()}
    $text=[Text.UTF8Encoding]::new($false,$true).GetString($bytes)
    $parts=$text.Split("`n")
    foreach($part in $parts){if([Text.Encoding]::UTF8.GetByteCount($part) -gt 65536){throw 'EVENT_LIMIT'}}
    $partial=$parts[-1].Length -gt 0
    $phases=[Collections.Generic.List[object]]::new()
    foreach($candidate in $candidates){
        foreach($phase in @('PREPARE','ZIP','DIRECTORY')){$phases.Add(@{Module=$candidate.Module;Version=$candidate.Version;Phase=$phase;Status='NotRun';ElapsedNs=$null;Digest=$null})}
    }
    $events=[Collections.Generic.List[object]]::new();$slot=0;$expect='START';$failed=$false
    for($i=0;$i -lt $parts.Length-1;$i++){
        if($slot -ge $phases.Count -or $failed){throw 'ORDER_AFTER_TERMINAL'}
        $e=$parts[$i]|ConvertFrom-Json -AsHashtable -Depth 10
        foreach($key in @('Schema','Sequence','Session','Module','Version','Phase','Kind','Status','Digest','ElapsedNs')){if(-not $e.ContainsKey($key)){throw 'EVENT_SCHEMA'}}
        if($e.Schema -ne 1 -or $e.Sequence -ne $i+1 -or $e.Session -cne $Session){throw 'BINDING'}
        $p=$phases[$slot]
        if($e.Module -cne $p.Module -or $e.Version -cne $p.Version -or $e.Phase -cne $p.Phase -or $e.Kind -cne $expect){throw 'ORDER'}
        if($e.ElapsedNs -isnot [ValueType] -or $e.ElapsedNs -lt 0 -or [decimal]$e.ElapsedNs%1 -ne 0){throw 'EVENT_TIME'}
        if($expect -ceq 'START'){
            if($e.Status -cne 'STARTED' -or $e.Digest -cne '' -or $e.ElapsedNs -ne 0){throw 'EVENT_START'}
            $p.Status='Incomplete';$expect='END'
        }else{
            if($e.Status -notin @('OK','ERROR','MISMATCH')){throw 'EVENT_END'}
            if($e.Status -ceq 'OK' -and $e.Phase -cne 'PREPARE' -and $e.Digest -cnotmatch '^h1:[A-Za-z0-9+/]{43}=$'){throw 'EVENT_DIGEST'}
            $p.Status=$(if($e.Status -ceq 'OK'){'Complete'}else{'Failed'})
            $p.ElapsedNs=$e.ElapsedNs;$p.Digest=$e.Digest;$failed=$e.Status -cne 'OK'
            $expect='START';$slot++
        }
        $events.Add($e)
    }
    return @{Complete=($slot -eq $phases.Count -and -not $partial -and -not $failed);Partial=$partial;Failed=$failed;Phases=$phases.ToArray();Events=$events.ToArray()}
}
function Invoke-C12PhaseOwned {
    param([string]$Executable,[string[]]$Arguments,[string]$WorkingDirectory,[hashtable]$Environment,[string]$Root,[int]$DeadlineMs,[string]$Session='',[string]$Profile='paired-v1')
    $null=Get-C12PhaseProfile $Profile
    if($Profile -cne 'paired-v1' -and -not $Session){throw 'PROFILE_SESSION'}
    Assert-C12WaitPath $Root
    if($DeadlineMs -le 0){throw 'INVALID_DEADLINE'}
    $p=[Diagnostics.Process]::new()
    $p.StartInfo=New-C12WaitProcessInfo $Executable $Arguments $WorkingDirectory
    foreach($key in @($p.StartInfo.Environment.Keys)){if($key -match '^(GO|GIT_)'){[void]$p.StartInfo.Environment.Remove($key)}}
    foreach($key in $Environment.Keys){$p.StartInfo.Environment[$key]=$Environment[$key]}
    $channels=[Collections.Generic.List[object]]::new();$clock=[Diagnostics.Stopwatch]::new()
    $started=$false;$confirmed=$false;$timedOut=$false;$reason='ERROR';$failure=$null;$exitCode=$null
    [long]$retained=0
    try {
        foreach($name in @('stdout','stderr')){
            $path=Join-Path $Root "$name.raw"
            $file=[IO.File]::Open($path,'CreateNew','Write','Read')
            $channels.Add(@{Name=$name;Path=$path;File=$file;Buffer=[byte[]]::new(8192);Pending=$null;Input=$null;Done=$false})
        }
        $clock.Start()
        if(-not $p.Start()){throw 'START_FAILED'}
        $started=$true
        $channels[0].Input=$p.StandardOutput.BaseStream;$channels[1].Input=$p.StandardError.BaseStream
        foreach($c in $channels){$c.Pending=$c.Input.ReadAsync($c.Buffer,0,$c.Buffer.Length)}
        $drainClock=$null
        while($true){
            if(-not $p.HasExited -and $clock.ElapsedMilliseconds -ge $DeadlineMs){$timedOut=$true;$reason='DEADLINE';$p.Kill($true);if(-not $p.WaitForExit(5000)){throw 'EXIT_UNCONFIRMED'}}
            foreach($c in $channels){
                if(-not $c.Done -and $c.Pending.IsCompleted){
                    $n=$c.Pending.GetAwaiter().GetResult()
                    if($n -eq 0){$c.Done=$true;continue}
                    # Reserve128KiB for final metadata; never retain over total cap.
                    $keep=[Math]::Min($n,[Math]::Max(0,4063232-$retained))
                    if($keep -gt 0){$c.File.Write($c.Buffer,0,$keep);$c.File.Flush();$retained+=$keep}
                    if($keep -ne $n){throw 'TOTAL_LIMIT'}
                    if($Session -and $c.Name -ceq 'stdout'){
                        $parsed=Read-C12PhaseEvents -Path $c.Path -Session $Session -Profile $Profile
                        if($parsed.Failed){throw 'PHASE_FAILED'}
                    }
                    $c.Pending=$c.Input.ReadAsync($c.Buffer,0,$c.Buffer.Length)
                }
            }
            if($p.HasExited){
                if($channels[0].Done -and $channels[1].Done){if(-not $timedOut){$reason='EXITED'};break}
                if($null -eq $drainClock){$drainClock=[Diagnostics.Stopwatch]::StartNew()}
                if($drainClock.ElapsedMilliseconds -ge 5000){throw 'DRAIN_UNCONFIRMED'}
            }
            Start-Sleep -Milliseconds 10
        }
        if($Session -and -not $timedOut){$parsed=Read-C12PhaseEvents -Path $channels[0].Path -Session $Session -Profile $Profile;if(-not $parsed.Complete){throw 'EVENTS_INCOMPLETE'}}
    }catch{$failure=$_.Exception.Message;if(-not $timedOut){$reason='ERROR'}}
    finally{
        if($started){
            try{if(-not $p.HasExited){$p.Kill($true)};$confirmed=$p.WaitForExit(5000);if($confirmed){$exitCode=$p.ExitCode}else{$failure='EXIT_UNCONFIRMED'}}catch{$failure='EXIT_UNCONFIRMED';$confirmed=$false}
        }
        foreach($c in $channels){$c.File.Dispose()}
        $p.Dispose();$clock.Stop()
    }
    $receipt=@{Started=$started;TimedOut=$timedOut;ExitConfirmed=$confirmed;ExitCode=$exitCode;ElapsedMs=$clock.Elapsed.TotalMilliseconds;EndReason=$reason;Failure=$failure;StdoutPath=(Join-Path $Root 'stdout.raw');StderrPath=(Join-Path $Root 'stderr.raw');RetainedBytes=$retained}
    $null=Write-C12WaitRecord $Root 'process-result.json' $receipt
    return $receipt
}
