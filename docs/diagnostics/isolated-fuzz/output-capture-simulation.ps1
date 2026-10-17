# Pure supplied-event model. No OS clock,external streams,files or process API.
. (Join-Path $PSScriptRoot 'output-evidence-package.ps1')
function Get-OfflineCaptureValueType([object]$Value){
    if($null -eq $Value){return $null}
    # Unwrap the PSObject layer;invoke Object.GetType without caller ETS dispatch.
    return [object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)
}
function Assert-OfflineCaptureEvent($Event){
    $type=Get-OfflineCaptureValueType $Event
    if($null -eq $type -or $type.FullName -cne 'System.Management.Automation.PSCustomObject'){throw 'invalid-event'}
    # Inspect all descriptors before reading any value;never invoke ScriptProperty.
    $properties=@($Event.PSObject.Properties)
    foreach($property in $properties){if($property.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){throw 'invalid-event'}}
    if($null -eq $Event.PSObject.Properties['Kind'] -or $Event.Kind -isnot [string] -or $Event.Kind -cnotin @('chunk','eof','exit','tick')){throw 'invalid-event'}
    $names=switch($Event.Kind){
        chunk{@('Kind','ElapsedMs','Channel','Bytes')}
        eof{@('Kind','ElapsedMs','Channel')}
        exit{@('Kind','ElapsedMs','ExitCode')}
        tick{@('Kind','ElapsedMs')}
    }
    # Do not forward caller objects to the package helper's ETS GetType check.
    if($properties.Count -ne $names.Count){throw 'invalid-event'}
    foreach($property in $properties){if($property.Name -cnotin $names){throw 'invalid-event'}}
    Assert-OfflineEvidenceInteger $Event.ElapsedMs 0 ([long]::MaxValue)
    if($Event.Kind -cin @('chunk','eof')){
        if($Event.Channel -isnot [string] -or $Event.Channel -cnotin @('stdout','stderr')){throw 'invalid-event'}
    }
    if($Event.Kind -ceq 'chunk'){
        if($Event.Bytes -isnot [byte[]] -or $Event.Bytes.Length -gt 65536){throw 'invalid-event'}
    }
    if($Event.Kind -ceq 'exit'){Assert-OfflineEvidenceInteger $Event.ExitCode ([int]::MinValue) ([int]::MaxValue)}
}
function Get-OfflineCaptureSimulation([object]$Events,[object]$EncodingName,[object]$OriginDeclaration){
    Set-StrictMode -Version Latest
    $result=[pscustomobject]@{
        Schema=1;Policy='offline-capture-simulation-v1';Status='incomplete';Reason='incomplete'
        AcceptedEventCount=0;FailureEventIndex=$null;LastObservedElapsedMs=[long]0
        StdoutEOF=$false;StderrEOF=$false;ExitObserved=$false;ExitCode=$null
        Stdout=$null;Stderr=$null;Package=$null;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false
    }
    $stdout=New-Object IO.MemoryStream;$stderr=New-Object IO.MemoryStream
    try{
        $headerValid=$true
        try{
            if((Get-OfflineCaptureValueType $Events) -ne [object[]] -or $Events.Rank -ne 1 -or $Events.Length -gt 4096){throw 'invalid-input'}
            Assert-OfflineEvidenceMetadata 0 $EncodingName $OriginDeclaration
            $utf8=New-Object Text.UTF8Encoding($false,$true)
            if($utf8.GetByteCount($OriginDeclaration) -gt 4096){throw 'invalid-input'}
        }catch{$headerValid=$false;$result.Status='refused';$result.Reason='invalid-input'}
        if($headerValid){
            for($index=0;$index -lt $Events.Length;$index++){
                $event=$Events[$index]
                if($result.Status -ceq 'complete'){
                    # Trailing input always refuses,even malformed or beyond deadline.
                    try{
                        Assert-OfflineCaptureEvent $event
                        if($event.ElapsedMs -ge $result.LastObservedElapsedMs){$result.LastObservedElapsedMs=[long]$event.ElapsedMs}
                    }catch{}
                    $result.Status='refused';$result.Reason='trailing-after-complete';$result.FailureEventIndex=$index;break
                }
                try{Assert-OfflineCaptureEvent $event}catch{
                    $result.Status='refused';$result.Reason='invalid-event';$result.FailureEventIndex=$index;break
                }
                if($event.ElapsedMs -lt $result.LastObservedElapsedMs){
                    $result.Status='refused';$result.Reason='time-went-backward';$result.FailureEventIndex=$index;break
                }
                $result.LastObservedElapsedMs=[long]$event.ElapsedMs
                if($event.ElapsedMs -ge 120000){
                    $result.Status='timed-out';$result.Reason='deadline-reached';$result.FailureEventIndex=$index;break
                }
                $failure=$null
                switch($event.Kind){
                    chunk{
                        $isStdout=$event.Channel -ceq 'stdout'
                        $ended=if($isStdout){$result.StdoutEOF}else{$result.StderrEOF}
                        $target=if($isStdout){$stdout}else{$stderr}
                        if($ended){$failure='chunk-after-eof'}
                        elseif(($target.Length+$event.Bytes.Length) -gt 1048576){$failure='stream-limit'}
                        else{$target.Write($event.Bytes,0,$event.Bytes.Length)}
                    }
                    eof{
                        if($event.Channel -ceq 'stdout'){
                            if($result.StdoutEOF){$failure='duplicate-eof'}else{$result.StdoutEOF=$true}
                        }else{
                            if($result.StderrEOF){$failure='duplicate-eof'}else{$result.StderrEOF=$true}
                        }
                    }
                    exit{
                        if($result.ExitObserved){$failure='duplicate-exit'}else{$result.ExitObserved=$true;$result.ExitCode=[int]$event.ExitCode}
                    }
                    tick{}
                }
                if($null -ne $failure){$result.Status='refused';$result.Reason=$failure;$result.FailureEventIndex=$index;break}
                $result.AcceptedEventCount++
                if($result.StdoutEOF -and $result.StderrEOF -and $result.ExitObserved){$result.Status='complete';$result.Reason='complete'}
            }
        }
        $outBytes=$stdout.ToArray();$errBytes=$stderr.ToArray()
        $result.Stdout=New-OfflineEvidenceStream $outBytes;$result.Stderr=New-OfflineEvidenceStream $errBytes
        if($result.Status -ceq 'complete'){
            try{
                $package=New-OfflineOutputEvidencePackage -StdoutBytes $outBytes -StderrBytes $errBytes -ExitCode $result.ExitCode -EncodingName $EncodingName -OriginDeclaration $OriginDeclaration
                if(-not (Test-OfflineOutputEvidencePackage $package).Verified){throw 'package-integrity-failed'}
                $result.Package=$package
            }catch{$result.Status='refused';$result.Reason='package-integrity-failed';$result.Package=$null;$result.FailureEventIndex=$null}
        }
    }finally{$stdout.Dispose();$stderr.Dispose()}
    return $result
}
