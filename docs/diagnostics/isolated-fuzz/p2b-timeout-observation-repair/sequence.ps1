# Pure byte predicate used by actual completion, never an authority setter.
function Test-LSStreamContract([object]$Bytes,[string]$Channel,[bool]$AllowPrefix){
    if($Bytes -isnot [byte[]] -or $Channel -cnotin @('Out','Err')){return $false}
    [byte[]]$expected=if($Channel -ceq 'Out'){[Text.Encoding]::ASCII.GetBytes('P2B-CONTAINMENT-OUT')}else{[Text.Encoding]::ASCII.GetBytes('P2B-CONTAINMENT-ERR')}
    # Only the fixed timeout profile supplies AllowPrefix=true. Zero bytes
    # remain a byte[]; null, bad types, non-prefix and suffix bytes reject.
    if($AllowPrefix){if($Bytes.Length -gt $expected.Length){return $false}}
    elseif($Bytes.Length -ne $expected.Length){return $false}
    for($i=0;$i -lt $Bytes.Length;$i++){if($Bytes[$i] -ne $expected[$i]){return $false}}
    return $true
}
function Get-LSStepDescriptor($Entry,[int]$Ordinal){
    $kind=$Entry.Profile.Steps[$Ordinal]
    if($kind -ceq 'Seam'){
        $scriptPath=Join-Path $PSScriptRoot '../p2b-handshake-seam-verification/handshake-seam.tests.ps1'
        $arguments=@('-TracePath',$Entry.Sinks.Trace1.Path)
        $max=110000
    }else{
        $scriptPath=Join-Path $PSScriptRoot '../p2b-controller-exception-containment/fixture-child.ps1'
        $arguments=@('-Mode',$kind)
        $max=if($Entry.Case -ceq 'timeout'){200}else{2000}
    }
    [pscustomobject]@{Kind=$kind;Script=[IO.Path]::GetFullPath($scriptPath);Arguments=$arguments;MaxMs=$max;Out=('Out'+$Ordinal);Err=('Err'+$Ordinal);Receipt=('Receipt'+$Ordinal);Ordinal=$Ordinal}
}
function LS-StepRoot([bool]$Accepted,$Rejection,$Ack,$Observation,$ErrorRecord){
    [pscustomobject][ordered]@{Accepted=$Accepted;Rejection=$Rejection;Ack=$Ack;Observation=$Observation;Error=$ErrorRecord}
}
function Complete-LSOwnedStep($Entry,$Descriptor,$Observation){
    $e=$Entry;$d=$Descriptor;$o=$Observation
    if($e.FailureLatched){return $false}
    if(-not $o.Started -or $o.Association -cne 'Associated' -or -not $o.OwnedExit -or -not $o.StreamsComplete){throw 'Incomplete owned lifecycle'}
    if($e.Case -ceq 'timeout'){
        if(-not $o.TimedOut -or -not $o.TerminationRequested -or $o.ExitCode -ne -1){throw 'Timeout lifecycle mismatch'}
    }elseif($o.TimedOut -or $o.TerminationRequested -or $o.ExitCode -ne 0){throw 'Unexpected target result'}
    $out=Read-LSBytes $e $d.Out
    if(-not $out.Ok){return $false}
    $err=Read-LSBytes $e $d.Err
    if(-not $err.Ok){return $false}
    $text=[Text.UTF8Encoding]::new($false,$true).GetString($out.Value)
    $errorText=[Text.UTF8Encoding]::new($false,$true).GetString($err.Value)
    if($d.Kind -ceq 'Seam'){
        if($errorText.Length -ne 0){throw 'Seam stderr not empty'}
        $sha=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($out.Value))
        if($sha -cne $script:Oracle.SeamStdoutSHA256){throw 'Seam stdout hash mismatch'}
        $trace=Read-LSBytes $e Trace1
        if(-not $trace.Ok){return $false}
        $traceText=[Text.UTF8Encoding]::new($false,$true).GetString($trace.Value)
        $traceHash=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($trace.Value))
        if($traceHash -cne $script:Oracle.SeamTraceSHA256){throw 'Seam trace hash mismatch'}
        # Compare ordered raw numeric tokens too; do not round UInt64 via JS.
        $expected=Get-Content -LiteralPath (Join-Path $PSScriptRoot '../p2b-handshake-seam-verification/expected-output.json') -Raw | ConvertFrom-Json -AsHashtable
        $actual=$traceText | ConvertFrom-Json -AsHashtable
        if((ConvertTo-Json -InputObject $actual -Depth 50 -Compress) -cne (ConvertTo-Json -InputObject $expected.TraceExpected -Depth 50 -Compress)){throw 'Seam complete ordered trace mismatch'}
        $o.SeamExact=$true
    }else{
        $allowPrefix=$e.Case -ceq 'timeout'
        if(-not (Test-LSStreamContract $out.Value Out $allowPrefix) -or -not (Test-LSStreamContract $err.Value Err $allowPrefix)){throw 'Fixture raw streams mismatch'}
    }
    Assert-LSPins $e
    $result=Write-LSLedger $e.Handle Ledger ([ordered]@{Outcome='CANDIDATE_COMPLETED';Ordinal=$d.Ordinal;OwnedExit=$o.OwnedExit;StreamsComplete=$o.StreamsComplete})
    if(-not $result.Ok){return $false}
    if($e.Grant -ceq 'sequence-green'){
        $publish=Write-LSOutput $e.Handle ConsoleOut ('STEP: '+$e.Case+'/'+$d.Ordinal+"`r`n")
        if(-not $publish.Ok){return $false}
        if($d.Kind -ceq 'Seam'){
            $publish=Write-LSOutput $e.Handle ConsoleOut $text
            if(-not $publish.Ok){return $false}
        }
    }
    # All fallible owned completion IO is above. This function returns before
    # the supervisor registers an acknowledgement; there is no success finally.
    return (-not $e.FailureLatched)
}
function Invoke-LSStep([object]$Session,[int]$Ordinal,[object]$PreviousAck){
    $e=Find-LSSession $Session
    if(-not (Test-LSAdvance $Session $Ordinal $PreviousAck)){return LS-StepRoot $false 'gate-rejected' $null $null $null}
    # Baseline disk-positive core fixture never calls this invalid step path.
    if($Ordinal -ge $e.Profile.Steps.Count){return LS-StepRoot $false 'no-fixed-step' $null $null $null}
    $observation=$null
    try{
        Assert-LSPins $e
        if($Ordinal -gt 0){foreach($ack in $e.Acks){if([object]::ReferenceEquals($ack.Handle,$PreviousAck)){$ack.Consumed=$true;$e.AckConsumed++;break}}}
        $e.Reserved++;$e.Status='Active'
        $pre=Write-LSLedger $Session Start ([ordered]@{Ordinal=$Ordinal;Reserved=$e.Reserved;CandidateOnly=$true})
        if(-not $pre.Ok){return LS-StepRoot $true $null $null $null $e.PrimaryError}
        if($script:HostStarts -ge $script:Oracle.Grants[$e.Grant].Children){throw 'Host target budget exhausted before Process creation'}
        $d=Get-LSStepDescriptor $e $Ordinal
        $observation=Invoke-LSOwnedStep $e $d
        if(-not (Complete-LSOwnedStep $e $d $observation)){return LS-StepRoot $true $null $null $observation $e.PrimaryError}
        $e.NextOrdinal++
        if($e.NextOrdinal -eq $e.Profile.Steps.Count){
            $null=Close-LSSession $Session Finish
            return LS-StepRoot $true $null $null $observation $null
        }
        if($e.Acks.Count -ge 2){throw 'Acknowledgement registration limit'}
        $handle=[object]::new()
        $e.Acks.Add([pscustomobject]@{Handle=$handle;Owner=$e;Ordinal=$Ordinal;Frozen=$e.Frozen;Table=$e.TableSHA256;Consumed=$false;Revoked=$false})
        $e.AckIssued++
        return LS-StepRoot $true $null $handle $observation $null
    }catch{Add-LSFailure $Session $_;return LS-StepRoot $true $null $null $observation $e.PrimaryError}
}
