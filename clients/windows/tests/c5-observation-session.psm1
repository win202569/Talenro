#requires -Version 7.0
Import-Module (Join-Path $PSScriptRoot 'observation-session-guard.psm1')

function Invoke-TalenroC5Observation {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][ValidateSet('Synthetic','Live')][string]$Mode,
        [Parameter(Mandatory)][Collections.IDictionary]$Context,
        [Parameter(Mandatory)][Collections.IDictionary]$Backend
    )
    $ErrorActionPreference='Stop'
    $names=@('Preflight','AcquireWriter','AcquireProcess','GetIdentity','ReadNative','ReadCim','ReadManaged','ElapsedMs','SleepMs','Publish')
    if($Backend.Count -ne $names.Count){throw 'Invalid C5 backend.'}
    foreach($name in $names){if($name -cnotin $Backend.Keys -or $Backend[$name] -isnot [scriptblock]){throw 'Invalid C5 backend.'}}
    if($Context.Count -ne 3 -or $Context.RunId -isnot [string] -or
        $Context.RunId -cnotmatch '^[0-9a-f]{32}$' -or $Context.RunId -eq ('0'*32) -or
        $Context.ExpectedPath -isnot [string] -or [string]::IsNullOrWhiteSpace($Context.ExpectedPath) -or
        $Context.ExpectedCommandLine -isnot [string] -or [string]::IsNullOrWhiteSpace($Context.ExpectedCommandLine)){throw 'Invalid C5 context.'}
    # Complete trusted callbacks only: no live fallback is present in this module.
    & $Backend.Preflight $Context | Out-Null
    $state=@{LastMs=[double]0;Sequence=0;TargetPid=$null;Phase='identity-failed'}
    $readClock={
        $value=& $Backend.ElapsedMs
        if(($value -isnot [double] -and $value -isnot [long] -and $value -isnot [int]) -or
            -not[double]::IsFinite([double]$value) -or $value -lt 0 -or $value -lt $state.LastMs){throw 'Invalid monotonic clock.'}
        $state.LastMs=[double]$value
        return [double]$value
    }.GetNewClosure()
    $identityGuard=Get-Command Assert-TalenroObservationIdentity -CommandType Function
    $body={param($processResource,$writer)
        function Clock {
            & $readClock
        }
        function Save([string]$kind,[Collections.IDictionary]$data) {
            $state.Sequence++
            $record=[ordered]@{SchemaVersion=1;Mode=$Mode;RunId=$Context.RunId;Kind=$kind;Sequence=$state.Sequence;ElapsedMs=$state.LastMs}
            foreach($key in $data.Keys){$record[$key]=$data[$key]}
            $writer.WriteLine(($record|ConvertTo-Json -Depth 20 -Compress))
            return [pscustomobject]$record
        }
        function FileTime($pair) {
            if($null -eq $pair -or $pair.High -isnot [uint32] -or $pair.Low -isnot [uint32] -or $pair.High -gt [int]::MaxValue){throw 'Invalid FILETIME.'}
            return [long](([long]$pair.High * 4294967296L) + [long]$pair.Low)
        }
        function ValidateNative($n) {
            $expectedMode=if($Mode -eq 'Live'){'native-live'}else{'synthetic'}
            if($null -eq $n -or $n -is [array] -or $n.SchemaVersion -ne 1 -or $n.Mode -cne $expectedMode -or
                $n.CorrelationId -cne $Context.RunId -or $null -ne $n.SnapshotError){throw 'Invalid native snapshot.'}
            foreach($key in @('Wait','ExitCode','Times')){
                if($n.Calls.$key.Status -cne 'ok' -or $n.Calls.$key.ApiSucceeded -isnot [bool] -or -not $n.Calls.$key.ApiSucceeded){throw 'Invalid native API evidence.'}
            }
            if($n.Calls.Wait.RawResult -isnot [uint32] -or $n.Calls.Wait.RawResult -notin @(0,258) -or
                $n.Calls.Wait.WaitState -cne $(if($n.Calls.Wait.RawResult -eq 0){'signaled'}else{'timeout'}) -or
                $n.Calls.ExitCode.ExitCode -isnot [uint32]){throw 'Invalid native values.'}
            $creation=FileTime $n.Calls.Times.Times.Creation
            [void](FileTime $n.Calls.Times.Times.Exit)
            return $creation
        }
        function ValidateManagedIdentity($m,$identity) {
            $expectedMode=if($Mode -eq 'Live'){'live'}else{'synthetic'}
            if($null -eq $m -or $m -is [array] -or $m.SchemaVersion -ne 1 -or $m.Mode -cne $expectedMode -or
                $m.Fields.ProcessId.Status -cne 'ok' -or $m.Fields.ProcessId.Value -isnot [int] -or $m.Fields.ProcessId.Value -ne $identity.ProcessId -or
                $m.Fields.StartTime.Status -cne 'ok' -or $m.Fields.StartTime.Value -cne $identity.StartUtc.ToString('o')){throw 'Managed identity mismatch.'}
        }
        function Finish([string]$outcome,[bool]$verified,[string]$reason){
            Save 'outcome' @{Outcome=$outcome;ExitVerified=$verified;Finalized=$false;Reason=$reason;TargetPid=$state.TargetPid;OverrunMs=[Math]::Max(0,$state.LastMs-360000)}
        }
        try {
            if($null -eq $writer.PSObject.Methods['WriteLine']){throw 'Missing receipt writer.'}
            # PID is available even if later handle acquisition fails on a real Process.
            if($processResource.PSObject.Properties['Id']){$state.TargetPid=$processResource.Id}
            [void](Clock)
            $start=Save 'start' @{TargetPid=$state.TargetPid}
            & $Backend.Publish $start | Out-Null
            if((Clock) -ge 360000){return (Finish 'timed-out' $false 'deadline')}
            $identity=& $Backend.GetIdentity $processResource
            $state.TargetPid=$identity.ProcessId
            if($null -eq $identity.Handle){throw 'Missing retained handle.'}
            if((Clock) -ge 360000){return (Finish 'timed-out' $false 'deadline')}
            $native=& $Backend.ReadNative $identity.Handle $Context.RunId
            [void](Clock)
            [void](Save 'native-identity' @{Native=$native;IdentityVerified=$false;TargetPid=$state.TargetPid})
            $creation=ValidateNative $native
            if((Clock) -ge 360000){return (Finish 'timed-out' $false 'deadline')}
            $cim=& $Backend.ReadCim $identity.ProcessId
            [void](Clock)
            & $identityGuard -ProcessId $identity.ProcessId -CimProcessId $cim.ProcessId `
                -ExpectedPath $Context.ExpectedPath -CimPath $cim.ExecutablePath `
                -ExpectedCommandLine $Context.ExpectedCommandLine -CimCommandLine $cim.CommandLine `
                -ProcessStartUtc $identity.StartUtc -CimCreationUtc $cim.CreationUtc -NativeCreationFileTime $creation
            $state.Phase='observation-failed'
            while((Clock) -lt 360000){
                $native=& $Backend.ReadNative $identity.Handle $Context.RunId
                $managed=$null
                try{
                    [void](Clock)
                    if($state.LastMs -lt 360000){
                        $managed=& $Backend.ReadManaged $processResource
                        [void](Clock)
                    }
                }catch{
                    # Keep the already acquired native evidence even if managed/time acquisition fails.
                    [void](Save 'pair' @{Native=$native;Managed=$managed;IdentityVerified=$false;Incomplete=$true})
                    throw
                }
                if($null -eq $managed -and $state.LastMs -ge 360000){
                    [void](Save 'pair' @{Native=$native;Managed=$null;IdentityVerified=$false;Incomplete=$true})
                    return (Finish 'timed-out' $false 'deadline')
                }
                # Save raw receipts before checks, including disagreement or invalid evidence.
                [void](Save 'pair' @{Native=$native;Managed=$managed;IdentityVerified=$true;Incomplete=$false})
                if($state.LastMs -ge 360000){return (Finish 'timed-out' $false 'deadline')}
                if((ValidateNative $native) -ne $creation){throw 'Native identity changed.'}
                ValidateManagedIdentity $managed $identity
                if($native.Calls.Wait.WaitState -ceq 'signaled'){
                    $exit=FileTime $native.Calls.Times.Times.Exit
                    $valid=$native.Calls.ExitCode.ExitCode -eq 0 -and $exit -ge $creation -and $exit -gt 0
                    foreach($field in @('WaitSignaled','HasExited','ExitCode','ExitTime')){if($managed.Fields.$field.Status -cne 'ok'){$valid=$false}}
                    $valid=$valid -and $managed.Fields.WaitSignaled.Value -is [bool] -and $managed.Fields.WaitSignaled.Value -and
                        $managed.Fields.HasExited.Value -is [bool] -and $managed.Fields.HasExited.Value -and
                        $managed.Fields.ExitCode.Value -is [int] -and $managed.Fields.ExitCode.Value -eq 0
                    if($exit -gt 0){$valid=$valid -and $managed.Fields.ExitTime.Value -ceq [datetime]::FromFileTimeUtc($exit).ToString('o')}
                    return (Finish 'exited' ([bool]$valid) $(if($valid){'agreement'}else{'unverified-exit'}))
                }
                $left=360000-(Clock)
                if($left -le 0){break}
                $sleep=[int][Math]::Floor([Math]::Min(2000,$left))
                if($sleep -le 0){return (Finish 'timed-out' $false 'submillisecond-remainder')}
                & $Backend.SleepMs $sleep | Out-Null
            }
            return (Finish 'timed-out' $false 'deadline')
        } catch {
            # Best-effort failure receipt. Writer failure still throws; no fabricated receipt.
            $timingVerified=$false
            try{[void](Clock);$timingVerified=$true}catch{}
            try {[void](Save 'outcome' @{Outcome=$state.Phase;ExitVerified=$false;Finalized=$false;TimingVerified=$timingVerified;Reason='callback-or-evidence-error';TargetPid=$state.TargetPid;OverrunMs=[Math]::Max(0,$state.LastMs-360000)})} catch {}
            throw [InvalidOperationException]::new("C5 $($state.Phase).")
        }
    }.GetNewClosure()
    try {
        $result=Invoke-TalenroObservationScope -AcquireWriter { & $Backend.AcquireWriter $Context } `
            -AcquireProcess { & $Backend.AcquireProcess $Context } -Observe $body
        # Persisted outcome is provisional: cleanup can fail or take us past the cap.
        $elapsed=& $readClock
        $result.ElapsedMs=[double]$elapsed
        $result.OverrunMs=[Math]::Max(0,$elapsed-360000)
        if($elapsed -ge 360000){$result.Outcome='timed-out';$result.ExitVerified=$false;$result.Reason='deadline'}
        $result.Finalized=$true
        & $Backend.Publish $result | Out-Null
        return $result
    } catch {
        $timingVerified=$false
        try{[void](& $readClock);$timingVerified=$true}catch{}
        try { & $Backend.Publish ([pscustomobject]@{SchemaVersion=1;Mode=$Mode;RunId=$Context.RunId;Kind='abort';TargetPid=$state.TargetPid;ExitVerified=$false;Reason='session-failed';ElapsedMs=$state.LastMs;OverrunMs=[Math]::Max(0,$state.LastMs-360000);TimingVerified=$timingVerified}) | Out-Null } catch {}
        throw [InvalidOperationException]::new('C5 session failed; no acceptance.')
    }
}

Export-ModuleMember -Function Invoke-TalenroC5Observation
