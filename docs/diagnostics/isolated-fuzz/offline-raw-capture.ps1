# Synthetic owned-memory transfer only. No external streams, processes, timers or automatic persistence.
. (Join-Path $PSScriptRoot 'capture-adapter-contract.ps1')
function Get-RawCaptureType($Value){if($null -eq $Value){return $null};return [object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)}
function Assert-RawCapturePassive($Value){
    if($null -eq $Value){return}
    foreach($p in $Value.PSObject.Properties){
        if($p.MemberType -in @([Management.Automation.PSMemberTypes]::ScriptProperty,[Management.Automation.PSMemberTypes]::CodeProperty)){throw 'Executable property'}
        if(([StringComparer]::OrdinalIgnoreCase.Equals($p.Name,'Length') -or [StringComparer]::OrdinalIgnoreCase.Equals($p.Name,'Rank')) -and $p.MemberType -ne [Management.Automation.PSMemberTypes]::Property){throw 'Shadow bound'}
    }
}
function Assert-RawCaptureShape($Value,[string[]]$Names){
    $type=Get-RawCaptureType $Value
    if($null -eq $type -or $type.FullName -cne 'System.Management.Automation.PSCustomObject'){throw 'Plain object required'}
    $props=@($Value.PSObject.Properties)
    if($props.Count -ne $Names.Count){throw 'Exact fields required'}
    foreach($p in $props){
        $found=$false;foreach($name in $Names){if([string]::Equals($p.Name,$name,[StringComparison]::Ordinal)){$found=$true;break}}
        if($p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or -not $found){throw 'Exact passive fields required'}
    }
}
function Assert-RawCaptureInteger($Value,[long]$Minimum,[long]$Maximum){
    $type=Get-RawCaptureType $Value
    if($type -ne [int] -and $type -ne [long]){throw 'Exact integer required'}
    Assert-RawCapturePassive $Value
    if($Value -lt $Minimum -or $Value -gt $Maximum){throw 'Integer bound'}
}
function Assert-RawCaptureToken($Value,[string[]]$Allowed){
    if((Get-RawCaptureType $Value) -ne [string]){throw 'Exact string required'}
    Assert-RawCapturePassive $Value
    $found=$false;foreach($item in $Allowed){if([string]::Equals($Value,$item,[StringComparison]::Ordinal)){$found=$true;break}}
    if(-not $found){throw 'Exact token required'}
}
function Assert-RawCaptureBytes($Value,[int]$Maximum){
    if((Get-RawCaptureType $Value) -ne [byte[]]){throw 'Byte array required'}
    Assert-RawCapturePassive $Value
    if($Value.Length -gt $Maximum){throw 'Byte bound'}
}
function Copy-RawCaptureBytes([byte[]]$Value){$copy=New-Object byte[] $Value.Length;[Array]::Copy($Value,$copy,$Value.Length);return ,$copy}
function New-OfflineRawCaptureRecord([object]$Scenario){
    Set-StrictMode -Version Latest
    $result=[pscustomobject]@{Schema=1;SourceMode='synthetic';Created=$false;Reason='invalid-scenario';Record=$null;Actions=$null;FailureActionIndex=$null;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false}
    $stdout=$null;$stderr=$null;$outCopy=$null;$errCopy=$null;$buffer=$null;$argumentCopy=$null
    try{
        Assert-RawCaptureShape $Scenario @('Schema','SourceMode','Header','StdoutBytes','StderrBytes','Actions')
        Assert-RawCaptureInteger $Scenario.Schema 1 1;Assert-RawCaptureToken $Scenario.SourceMode @('synthetic')
        Assert-RawCaptureBytes $Scenario.StdoutBytes 1048576;Assert-RawCaptureBytes $Scenario.StderrBytes 1048576
        $result.Reason='invalid-header'
        $header=$Scenario.Header
        Assert-RawCaptureShape $header @('SessionLabel','ExecutableDeclaration','ArgumentBytes','EncodingName','OriginDeclaration','StartedUTC','EndedUTC')
        foreach($p in $header.PSObject.Properties){Assert-RawCapturePassive $p.Value}
        Assert-RawCaptureBytes $header.ArgumentBytes 65536
        $argumentCopy=Copy-RawCaptureBytes $header.ArgumentBytes
        $packet=[pscustomobject]@{Schema=1;SourceMode='synthetic';SessionLabel=$header.SessionLabel;ExecutableDeclaration=$header.ExecutableDeclaration;ArgumentBytes=$argumentCopy;EncodingName=$header.EncodingName;OriginDeclaration=$header.OriginDeclaration;StartedUTC=$header.StartedUTC;EndedUTC=$header.EndedUTC;Events=[object[]]@()}
        $headerRecord=New-SyntheticCaptureRecord $packet
        if(-not (Test-SyntheticCaptureRecord $headerRecord).Verified){return $result}
        $result.Reason='invalid-action'
        $inputActions=$Scenario.Actions
        if((Get-RawCaptureType $inputActions) -ne [object[]]){throw 'Exact action array required'}
        Assert-RawCapturePassive $inputActions
        if($inputActions.Rank -ne 1 -or $inputActions.Length -gt 4096){throw 'Action bound'}
        $actions=New-Object 'System.Collections.Generic.List[object]'
        for($i=0;$i -lt $inputActions.Length;$i++){
            $result.FailureActionIndex=$i;$a=$inputActions[$i]
            # Inspect all descriptors before even reading Kind.
            $type=Get-RawCaptureType $a
            if($null -eq $type -or $type.FullName -cne 'System.Management.Automation.PSCustomObject'){throw 'Plain action required'}
            foreach($p in $a.PSObject.Properties){if($p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){throw 'Passive action required'}}
            if($null -eq $a.PSObject.Properties['Kind']){throw 'Missing kind'}
            Assert-RawCaptureToken $a.Kind @('read','exit','tick')
            $fields=switch -CaseSensitive ($a.Kind){read{@('Kind','ElapsedMs','Channel','Count')} exit{@('Kind','ElapsedMs','ExitCode')} tick{@('Kind','ElapsedMs')}}
            Assert-RawCaptureShape $a $fields
            Assert-RawCaptureInteger $a.ElapsedMs 0 ([long]::MaxValue)
            switch -CaseSensitive ($a.Kind){
                read{
                    Assert-RawCaptureToken $a.Channel @('stdout','stderr');Assert-RawCaptureInteger $a.Count 1 65536
                    $actions.Add([pscustomobject]@{Kind='read';ElapsedMs=[long]$a.ElapsedMs;Channel=[string]$a.Channel;Count=[int]$a.Count})
                }
                exit{Assert-RawCaptureInteger $a.ExitCode ([int]::MinValue) ([int]::MaxValue);$actions.Add([pscustomobject]@{Kind='exit';ElapsedMs=[long]$a.ElapsedMs;ExitCode=[int]$a.ExitCode})}
                tick{$actions.Add([pscustomobject]@{Kind='tick';ElapsedMs=[long]$a.ElapsedMs})}
            }
        }
        $result.FailureActionIndex=$null
        $result.Actions=[object[]]$actions.ToArray()
        $result.Reason='adapter-error'
        $outCopy=Copy-RawCaptureBytes $Scenario.StdoutBytes;$errCopy=Copy-RawCaptureBytes $Scenario.StderrBytes
        $stdout=[IO.MemoryStream]::new($outCopy,$false);$stderr=[IO.MemoryStream]::new($errCopy,$false)
        $buffer=New-Object byte[] 65536;$events=New-Object 'System.Collections.Generic.List[object]'
        $outEOF=$false;$errEOF=$false;$exitObserved=$false;$complete=$false;$last=[long]0
        $expectedStatus='incomplete';$expectedReason='incomplete';$failureIndex=$null
        for($i=0;$i -lt $actions.Count;$i++){
            $a=$actions[$i];$time=[long]$a.ElapsedMs;$stop=$false;$event=$null
            if($complete){
                $expectedStatus='refused';$expectedReason='trailing-after-complete';$stop=$true
                if($a.Kind -ceq 'read'){$event=[pscustomobject]@{Kind='tick';ElapsedMs=$time}}
            }elseif($time -lt $last){
                $expectedStatus='refused';$expectedReason='time-went-backward';$stop=$true
                if($a.Kind -ceq 'read'){$event=[pscustomobject]@{Kind='tick';ElapsedMs=$time}}
            }elseif($time -ge 120000){
                $expectedStatus='timed-out';$expectedReason='deadline-reached';$stop=$true
                if($a.Kind -ceq 'read'){$event=[pscustomobject]@{Kind='tick';ElapsedMs=$time}}
            }elseif($a.Kind -ceq 'read' -and (($a.Channel -ceq 'stdout' -and $outEOF) -or ($a.Channel -ceq 'stderr' -and $errEOF))){
                $expectedStatus='refused';$expectedReason='chunk-after-eof';$stop=$true
                $event=[pscustomobject]@{Kind='chunk';ElapsedMs=$time;Channel=$a.Channel;Bytes=[byte[]]@()}
            }elseif($a.Kind -ceq 'exit' -and $exitObserved){$expectedStatus='refused';$expectedReason='duplicate-exit';$stop=$true}
            if($null -eq $event){
                switch -CaseSensitive ($a.Kind){
                    read{
                        $stream=if($a.Channel -ceq 'stdout'){$stdout}else{$stderr}
                        $n=$stream.Read($buffer,0,$a.Count)
                        if($n -eq 0){
                            $event=[pscustomobject]@{Kind='eof';ElapsedMs=$time;Channel=$a.Channel}
                            if($a.Channel -ceq 'stdout'){$outEOF=$true}else{$errEOF=$true}
                        }else{
                            $bytes=New-Object byte[] $n;[Array]::Copy($buffer,$bytes,$n)
                            $event=[pscustomobject]@{Kind='chunk';ElapsedMs=$time;Channel=$a.Channel;Bytes=$bytes}
                        }
                    }
                    exit{$event=[pscustomobject]@{Kind='exit';ElapsedMs=$time;ExitCode=$a.ExitCode};if(-not $stop){$exitObserved=$true}}
                    tick{$event=[pscustomobject]@{Kind='tick';ElapsedMs=$time}}
                }
            }
            $events.Add($event)
            if($stop){$failureIndex=$i;break}
            $last=$time;$complete=$outEOF -and $errEOF -and $exitObserved
            if($complete){$expectedStatus='complete';$expectedReason='complete'}
        }
        $packet.Events=[object[]]$events.ToArray()
        $record=New-SyntheticCaptureRecord $packet
        $result.Reason='adapter-inconsistent'
        if(-not (Test-SyntheticCaptureRecord $record).Verified){return $result}
        if($record.Lifecycle.Status -cne $expectedStatus -or $record.Lifecycle.Reason -cne $expectedReason -or $record.Lifecycle.FailureEventIndex -ne $failureIndex){return $result}
        $result.Created=$true;$result.Reason='record-created';$result.Record=$record;$result.FailureActionIndex=$failureIndex
        return $result
    }catch{return $result}
    finally{
        if($null -ne $stdout){$stdout.Dispose()};if($null -ne $stderr){$stderr.Dispose()}
        foreach($owned in @($outCopy,$errCopy,$buffer,$argumentCopy)){if($null -ne $owned){[Array]::Clear($owned,0,$owned.Length)}}
    }
}
