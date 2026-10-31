function OC-Need([bool]$Condition,[string]$Message){if(-not $Condition){throw 'OD: observation contract mismatch'}}
function OC-EqualJSON($Actual,$Expected){
    return ((ConvertTo-Json -InputObject $Actual -Depth 50 -Compress) -ceq (ConvertTo-Json -InputObject $Expected -Depth 50 -Compress))
}

function OC-Keys($Value,[string[]]$Names,[string]$Label){
    OC-Need ($Value -is [Collections.IDictionary] -and (OC-EqualJSON @($Value.Keys) $Names)) ('Ordered complete root mismatch '+$Label)
}
function OC-ErrorLocation($Record,[string[]]$Sources){
    OC-Need ($Record.Position -is [string] -and $Record.Stack -is [string] -and $Record.Stack.Length -gt 0) 'Actual error source/stack required'
    $matched=$false
    foreach($relative in $Sources){
        $source=[IO.Path]::GetFullPath($relative)
        $match=[regex]::Match($Record.Position,([regex]::Escape($source)+':([0-9]+)'),[Text.RegularExpressions.RegexOptions]::IgnoreCase)
        if($match.Success){$line=[int]$match.Groups[1].Value;$matched=$line -gt 0 -and $line -le @([IO.File]::ReadAllLines($source)).Count;if($matched){break}}
    }
    OC-Need $matched 'Actual error must bind a frozen source and valid line'
}
function OC-Step($Step,$Oracle,[string]$ProfileId,$Context){
    $repo=$Context.Repo;$packet=$Context.PacketRoot
    OC-Keys $Step @('View','Result') StepObservation
    $r=$Step.Result;$v=$Step.View
    OC-Keys $r @('Accepted','Rejection','AckPresent','Observation','Error') StepResult
    OC-Need ($r.Accepted -is [bool] -and $r.AckPresent -is [bool]) 'Actual root booleans'
    if($null -ne $v){
        OC-Keys $v @('Status','FailureLatched','NextOrdinal','Reserved','StartCalls','AckIssued','AckConsumed','Root','Sinks','PrimaryError','SecondaryErrors','Overflow') SessionView
        $case=[IO.Path]::GetFileName($v.Root)
        OC-Need ($Oracle.Cases.Contains($case) -and $ProfileId -cin $Oracle.Cases[$case].Grants) 'Fixed profile/case binding'
        OC-Need ($v.Root -ceq [IO.Path]::GetFullPath((Join-Path $packet ('profiles/'+$ProfileId+'/'+$case)))) 'Fixed managed case root'
        OC-Need ($v.SecondaryErrors.Count -le 8 -and $v.StartCalls -le $Oracle.Cases[$case].Budget) 'Session evidence bounds'
    }else{OC-Need (-not $r.Accepted -and $null -eq $r.Observation -and $null -eq $r.Error) 'Unknown handle cannot create a step';return $null}
    if($null -ne $r.Error){
        OC-ErrorLocation $r.Error @($Context.Sources.Sinks)
        OC-Need ($case -cin @('preconsume','prestart','receipt-fault','sequence-output-fault')) 'Unexpected actual step error'
        OC-Need ($r.Error.Type -ceq 'System.Management.Automation.MethodInvocationException' -and $v.FailureLatched -and $v.AckIssued -eq 0 -and -not $r.AckPresent -and (OC-EqualJSON $r.Error $v.PrimaryError)) 'Actual negative cause/binding mismatch'
        if($case -ceq 'sequence-output-fault'){
            $expected='Exception calling "Write" with "1" argument(s): "Cannot write to a closed TextWriter.'+[Environment]::NewLine+'Object name: ''StreamWriter''."'
            OC-Need ($r.Error.ErrorId -ceq 'ObjectDisposedException' -and $r.Error.Message -ceq $expected -and $v.Sinks.ConsoleOut.State -ceq 'Failed' -and $v.Sinks.ConsoleOut.LogicalIOAttempts -eq 1) 'Actual output-fault signature'
        }else{
            $filename=switch -CaseSensitive ($case){'preconsume'{'start.json'}'prestart'{'stdout-0.bin'}'receipt-fault'{'receipt-0.json'}}
            $method=if($case -ceq 'prestart'){'Create'}else{'WriteAllText'}
            $argc=if($case -ceq 'prestart'){1}else{3}
            $expected=('Exception calling "{0}" with "{1}" argument(s): "Access to the path ''{2}'' is denied."' -f $method,$argc,(Join-Path $v.Root $filename))
            OC-Need ($r.Error.ErrorId -ceq 'UnauthorizedAccessException' -and $r.Error.Message -ceq $expected) 'Actual managed file-fault signature'
        }
    }
    $o=$r.Observation
    if($null -eq $o){return $null}
    OC-Keys $o @('Started','Association','PID','OwnedExit','ExitCode','StreamsComplete','TimedOut','TerminationRequested','ElapsedMs','KillCalls','WaitCalls','SeamExact','Stdout') OwnedObservation
    foreach($key in @('Started','OwnedExit','StreamsComplete','TimedOut','TerminationRequested','SeamExact')){OC-Need ($o[$key] -is [bool]) 'Lifecycle boolean type'}
    OC-Need ($o.ElapsedMs -ge 0) 'Negative actual duration'
    if(-not $o.Started){
        OC-Need ($case -ceq 'prestart' -and $o.Association -ceq 'Unassociated' -and $null -eq $o.PID -and $o.KillCalls -eq 0 -and $o.WaitCalls -eq 0 -and $v.StartCalls -eq 0 -and $null -ne $r.Error) 'Prestart exact nonassociation'
        return $null
    }
    OC-Need ($o.Association -ceq 'Associated' -and $o.PID -gt 0 -and $o.OwnedExit -and $o.StreamsComplete) 'Actual direct lifecycle incomplete'
    $max=if($o.SeamExact){114000}elseif($case -ceq 'timeout'){4200}else{6000}
    OC-Need ($o.ElapsedMs -le $max) 'Actual duration beyond unchanged bound'
    if($case -ceq 'timeout'){
        OC-Need ($o.TimedOut -and $o.TerminationRequested -and $o.ExitCode -eq -1 -and $null -eq $r.Error) 'Timeout tuple mismatch'
    }elseif($case -ceq 'receipt-fault'){
        OC-Need (-not $o.TimedOut -and $o.TerminationRequested -and $o.ExitCode -eq -1 -and $v.Sinks.Receipt0.LogicalIOAttempts -eq 1 -and $v.SecondaryErrors.Count -eq 0) 'Receipt no-retry tuple mismatch'
    }else{OC-Need (-not $o.TimedOut -and -not $o.TerminationRequested -and $o.ExitCode -eq 0) 'Normal direct lifecycle mismatch'}
    return $o
}

# Seed: same reviewed Step-only observation loop, intentionally incomplete.
function Assert-ODProfileObservations([object[]]$Rows,[string]$Profile,[Collections.IDictionary]$Context){
    foreach($detail in $Rows){
        OC-Keys $detail @('Name','Value') ProfileObservation
        if($detail.Name -ceq 'Step'){$null=OC-Step $detail.Value $Context.Oracle $Profile $Context}
    }
}
