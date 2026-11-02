# Pure complete observation/report predicates. Context supplies frozen source data.
function OC-Need([bool]$Condition){if(-not $Condition){throw 'OD: observation contract mismatch'}}
function OC-Keys($Value,[string[]]$Names){
    OC-Need ($Value -is [Collections.IDictionary])
    $keys=@($Value.Keys);OC-Need ($keys.Count -eq $Names.Count)
    for($i=0;$i -lt $Names.Count;$i++){OC-Need ($keys[$i] -is [string] -and $keys[$i] -ceq $Names[$i])}
}
function OC-Integer($Value){OC-Need ($Value -is [int] -or $Value -is [long])}
function OC-Expand([string]$Text,$Context){
    if($Text.Contains('@root')){$Text=$Text.Replace('@root',$Context.ProfileRoot).Replace('/','\')}
    return $Text.Replace('@NL',[Environment]::NewLine)
}
function OC-SourceError($Actual,$Expected,$Context){
    OC-Keys $Actual @('Type','Message','ErrorId','Position','Stack')
    foreach($k in @('Type','Message','ErrorId','Position','Stack')){OC-Need ($Actual[$k] -is [string])}
    OC-Need ($Actual.Type -ceq $Expected.Type -and $Actual.ErrorId -ceq $Expected.ErrorId -and $Actual.Message -ceq (OC-Expand $Expected.Message $Context))
    $source=$Expected.Source
    if($Context.Sources.Contains($source)){$source=$Context.Sources[$source]}
    OC-Need ($Context.SourceModel.Contains($source))
    $model=$Context.SourceModel[$source]
    $match=[regex]::Match($Actual.Position,('^At '+[regex]::Escape($source)+':([0-9]+) char:([0-9]+)\r?\n'))
    OC-Need $match.Success
    $line=[int]$match.Groups[1].Value;$column=[int]$match.Groups[2].Value
    $operation=$null
    if($Expected.Contains('Operation')){$operation=$Expected.Operation}
    elseif($Expected.Type -ceq 'System.Management.Automation.MethodInvocationException'){
        $method=[regex]::Match($Expected.Message,'^Exception calling "([^"]+)"')
        OC-Need $method.Success;$operation=$method.Groups[1].Value
    }elseif($Expected.ErrorId -ceq 'Invalid ledger row index'){$operation='throw'}
    elseif($Expected.ErrorId -ceq 'System.ArgumentException,Microsoft.PowerShell.Commands.ConvertFromJsonCommand'){$operation='ConvertFrom-Json'}
    else{OC-Need $false}
    $faults=@($model.Faults|Where-Object {$_.Function -ceq $Expected.Function -and $_.Operation -ceq $operation -and $_.Line -eq $line -and $_.Column -eq $column})
    OC-Need ($faults.Count -ge 1)
    $stack=$Actual.Stack.Replace(([string][char]13+[char]10),[string][char]10).Split([char]10)
    OC-Need ($stack.Count -ge 2 -and $stack.Count -le 32)
    $previousFunction=$null;$previousSource=$null;$previousLine=0
    for($i=0;$i -lt $stack.Count;$i++){
        $frame=[regex]::Match($stack[$i],'^at (.+), (.+): line ([0-9]+)$')
        OC-Need $frame.Success
        $function=$frame.Groups[1].Value -creplace '^(script|local|global|private):',''
        $path=$frame.Groups[2].Value;$frameLine=[int]$frame.Groups[3].Value
        OC-Need ($Context.SourceModel.Contains($path))
        $frameModel=$Context.SourceModel[$path]
        if($i -eq 0){OC-Need ($path -ceq $source -and $function -ceq $Expected.Function -and $frameLine -eq $line)}
        else{
            $calls=@($frameModel.Calls|Where-Object {$_.Caller -ceq $function -and $_.Line -eq $frameLine -and ($_.Callee -ceq $previousFunction -or ($_.Callee -ceq '@fixed-entry' -and $previousFunction -ceq '<ScriptBlock>'))})
            if($calls.Count -eq 0 -and $Context.Contains('UnitBindings') -and $previousSource -ceq $Context.UnitBindings.FactorySource -and $previousFunction -ceq $Context.UnitBindings.FactoryFunction){
                $calls=@($frameModel.Calls|Where-Object {$_.Caller -ceq $function -and $_.Line -eq $frameLine -and $_.Callee -ceq '@fixed-record-factory'})
            }
            if($calls.Count -eq 0 -and $Context.Contains('UnitBindings') -and $Context.UnitBindings.Contains('OwnerControlCall')){
                $binding=$Context.UnitBindings.OwnerControlCall
                $keys=@('Source','SHA256','Owner','Variable','Line','Column','BodyStartLine','BodyStartColumn','BodyEndLine','BodyEndColumn')
                OC-Keys $binding $keys
                OC-Need ($frameModel.Contains('DynamicCalls') -and $frameModel.DynamicCalls -is [array] -and $frameModel.DynamicCalls.Count -eq 1)
                $descriptor=$frameModel.DynamicCalls[0];OC-Keys $descriptor $keys
                foreach($key in $keys){
                    if($key -cin @('Source','SHA256','Owner','Variable')){OC-Need ($binding[$key] -is [string] -and $descriptor[$key] -is [string])}
                    else{OC-Integer $binding[$key];OC-Integer $descriptor[$key];OC-Need ($binding[$key] -gt 0)}
                    OC-Need ($binding[$key] -ceq $descriptor[$key])
                }
                OC-Need ($binding.Source -ceq $path -and $binding.SHA256 -ceq $frameModel.SHA256 -and $binding.Owner -ceq 'Invoke-ODOwnerControls' -and $binding.Variable -ceq 'controls' -and $function -ceq $binding.Owner -and $frameLine -eq $binding.Line)
                OC-Need ($previousFunction -ceq '<ScriptBlock>' -and $previousSource -ceq $path -and $binding.BodyEndLine -ge $binding.BodyStartLine -and $previousLine -ge $binding.BodyStartLine -and $previousLine -le $binding.BodyEndLine)
                $calls=@($descriptor)
            }
            OC-Need ($calls.Count -ge 1)
        }
        $previousFunction=$function;$previousSource=$path;$previousLine=$frameLine
    }
    OC-Need ($previousSource -ceq $Context.Sources.Capture -and $previousFunction -ceq '<ScriptBlock>')
}
function OC-Match($Actual,$Expected,$Context,$State){
    if($null -eq $Expected){OC-Need ($null -eq $Actual);return}
    if($Expected -is [Collections.IDictionary]){
        if($Expected.Contains('$error')){OC-SourceError $Actual $Expected['$error'] $Context;return}
        if($Expected.Contains('$dynamic')){
            OC-Integer $Actual
            $kind=$Expected['$dynamic'];$key=$Expected.Key
            if($kind -ceq 'pid'){
                OC-Need ($Actual -gt 0)
                if($State.Pids.Contains($key)){OC-Need ($State.Pids[$key] -eq $Actual)}
                else{OC-Need ($Actual -cnotin @($State.Pids.Values));$State.Pids[$key]=$Actual}
            }elseif($kind -ceq 'elapsed'){
                OC-Need ($Actual -ge 0 -and $Actual -le $Expected.Max)
                if($State.Elapsed.Contains($key)){OC-Need ($State.Elapsed[$key] -eq $Actual)}else{$State.Elapsed[$key]=$Actual}
            }else{OC-Need $false}
            return
        }
        OC-Keys $Actual @($Expected.Keys)
        foreach($k in $Expected.Keys){OC-Match $Actual[$k] $Expected[$k] $Context $State}
        return
    }
    if($Expected -is [array]){
        OC-Need ($Actual -is [array] -and $Actual.Count -eq $Expected.Count)
        for($i=0;$i -lt $Expected.Count;$i++){OC-Match $Actual[$i] $Expected[$i] $Context $State}
        return
    }
    if($Expected -is [bool]){OC-Need ($Actual -is [bool] -and $Actual -ceq $Expected);return}
    if($Expected -is [string]){OC-Need ($Actual -is [string] -and $Actual -ceq (OC-Expand $Expected $Context));return}
    OC-Integer $Expected;OC-Integer $Actual;OC-Need ($Actual -eq $Expected)
}
function Assert-ODProfileObservations([object[]]$Rows,[string]$Profile,[Collections.IDictionary]$Context){
    OC-Need ($Profile -cin @('entry-migration-core','entry-migration-owned','entry-migration-live') -and $Context.Rows.Contains($Profile))
    $count=switch -CaseSensitive ($Profile){'entry-migration-core'{22}'entry-migration-owned'{8}'entry-migration-live'{15}}
    OC-Need ($Rows.Count -eq $count -and $Context.Rows[$Profile] -is [array] -and $Context.Rows[$Profile].Count -eq $count)
    $state=[pscustomobject]@{Pids=[ordered]@{};Elapsed=[ordered]@{}}
    OC-Match $Rows $Context.Rows[$Profile] $Context $state
}
function Assert-ODProfileReport([Collections.IDictionary]$Report,[string]$Profile,[Collections.IDictionary]$Context){
    OC-Keys $Report @('Schema','Profile','EntryPath','EntrySHA256','KernelPath','KernelSHA256','FrozenSHA256','Status','Groups','Observations','PrimaryError','SecondaryErrors','Overflow','Closed','SourceVerified','LiveActionsPerformed','RuntimeApproved','FormalAcceptance','RealGateImplemented')
    OC-Integer $Report.Schema
    OC-Need ($Report.Schema -eq 2 -and $Profile -cin @('entry-migration-core','entry-migration-owned','entry-migration-live') -and $Report.Profile -is [string] -and $Report.Profile -ceq $Profile)
    foreach($k in @('EntryPath','EntrySHA256','KernelPath','KernelSHA256','FrozenSHA256','Status')){OC-Need ($Report[$k] -is [string])}
    OC-Need ($Report.EntryPath -ceq $Context.Entry.Path -and $Report.EntrySHA256 -ceq $Context.Entry.SHA256 -and $Report.KernelPath -ceq $Context.Kernel.Path -and $Report.KernelSHA256 -ceq $Context.Kernel.SHA256 -and $Report.FrozenSHA256 -ceq $Context.Manifest.SHA256)
    OC-Need ($Report.Groups -is [array] -and $Report.Observations -is [array] -and $Report.SecondaryErrors -is [array] -and $Report.SecondaryErrors.Count -le 8 -and $Report.Overflow -is [bool] -and -not $Report.Overflow -and $Report.Closed -is [bool])
    foreach($k in @('SourceVerified','LiveActionsPerformed','RuntimeApproved','FormalAcceptance','RealGateImplemented')){OC-Need ($Report[$k] -is [bool] -and -not $Report[$k])}
    if($Report.Status -ceq 'COMPLETE'){
        OC-Need ($Report.Closed -and $null -eq $Report.PrimaryError -and $Report.SecondaryErrors.Count -eq 0)
        $grant=switch -CaseSensitive ($Profile){'entry-migration-core'{'core'}'entry-migration-owned'{'lifecycle'}'entry-migration-live'{'continuation'}}
        $all=$Context.Oracle.Grants[$grant].Groups
        $groups=@($all|Select-Object -Skip 6 -First ($all.Count-11))
        OC-Match $Report.Groups $groups $Context ([pscustomobject]@{Pids=[ordered]@{};Elapsed=[ordered]@{}})
        Assert-ODProfileObservations $Report.Observations $Profile $Context
    }elseif($Report.Status -ceq 'INCOMPLETE'){
        OC-Need ($Context.Contains('IntentFailure') -and $Report.Groups.Count -eq 0 -and $Report.Observations.Count -eq 0 -and $null -ne $Report.PrimaryError -and $Report.SecondaryErrors.Count -eq 0)
        OC-SourceError $Report.PrimaryError $Context.IntentFailure $Context
    }else{OC-Need $false}
}
