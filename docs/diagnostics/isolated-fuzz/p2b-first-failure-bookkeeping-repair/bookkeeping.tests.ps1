param([Parameter(Mandatory)][ValidateSet('RED','GREEN')][string]$Phase,[Parameter(Mandatory)][string]$TracePath,[Parameter(Mandatory)][string]$DetailsPath)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-19-p2b-first-failure-bookkeeping-repair'
$prefix=$Phase.ToLowerInvariant()
if([IO.Path]::GetFullPath($TracePath) -cne (Join-Path $packet ($prefix+'.trace.json')) -or [IO.Path]::GetFullPath($DetailsPath) -cne (Join-Path $packet ($prefix+'.details.json'))){throw 'Fixed bookkeeping outputs required'}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
if(([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary test required'}
. (Join-Path $PSScriptRoot 'bookkeeping.ps1')
$control=New-P2BFailureState
$groups=[Collections.Generic.List[string]]::new();$observations=[ordered]@{};$events=[Collections.Generic.List[object]]::new()
function B-Need([bool]$Condition,[string]$Name){if(-not $Condition){throw ('ASSERTION FAILED: '+$Name)}}
function B-Pass([string]$Name){$groups.Add($Name);[Console]::Out.WriteLine('PASS: '+$Name)}
function B-Describe([Management.Automation.ErrorRecord]$Record){
    $items=[Collections.Generic.List[object]]::new()
    [Exception]$node=$Record.Exception
    while($null -ne $node -and $items.Count -lt 8){
        $message=$node.Message
        $truncated=$message.Length -gt 1024
        if($truncated){$message=$message.Substring(0,1024)}
        $items.Add([pscustomobject][ordered]@{Type=$node.GetType().FullName;Message=$message;MessageTruncated=$truncated})
        $node=$node.InnerException
    }
    return [pscustomobject][ordered]@{RecordType=$Record.GetType().FullName;FullyQualifiedErrorId=$Record.FullyQualifiedErrorId;Chain=$items.ToArray();ChainTruncated=$null -ne $node}
}
function B-Record([string]$Name,$State){
    $p=if($null -ne $State.PrimaryError){B-Describe $State.PrimaryError}else{$null}
    $s=@($State.SecondaryErrors|ForEach-Object{B-Describe $_})
    $events.Add([pscustomobject][ordered]@{Name=$Name;Primary=$p;Secondary=$s;Overflow=$State.SecondaryOverflow})
}
$passed=$false;$errorText=$null;$redTrace=$null
try{
    $caseDir=Join-Path $packet ($prefix+'-bookkeeping-cases')
    if(Test-Path -LiteralPath $caseDir){throw 'Fresh bookkeeping cases required'}
    $null=New-Item -ItemType Directory -Path $caseDir
    $badSink=Join-Path $caseDir 'sink-is-directory';$null=New-Item -ItemType Directory -Path $badSink
    if($Phase -ceq 'RED'){
        $first=$null;try{$null=[IO.File]::ReadAllText((Join-Path $caseDir 'missing-first'))}catch{$first=$_}
        B-Need ($null -ne $first) 'Real first failure'
        $s=New-P2BFailureState;Add-P2BFailure $s $first
        $escaped=$null;$returnedNormally=$false
        try{$null=Save-P2BLedger $s $badSink @([pscustomobject]@{Number=1});$returnedNormally=$true}catch{$escaped=$_}
        $redTrace=[pscustomobject][ordered]@{Schema=1;Mode='p2b-bookkeeping-red';Escaped=$null -ne $escaped;ReturnedNormally=$returnedNormally;PrimarySame=[object]::ReferenceEquals($s.PrimaryError,$first);EscapedDistinct=$null -ne $escaped -and -not [object]::ReferenceEquals($escaped,$first);StartCalls=0;CompileCount=0;ProductChanged=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
        B-Need ($redTrace.Escaped -and -not $redTrace.ReturnedNormally -and $redTrace.PrimarySame -and $redTrace.EscapedDistinct) 'Exact known RED'
        B-Record 'red-first' $s
        $events.Add([pscustomobject]@{Name='red-escaped';Error=(B-Describe $escaped)})
        $passed=$true
    }else{
        $names=@('ChildFinalWritePreservesFirst','IntegrationFinalWritePreservesFirst','ParentFinalWritePreservesFirst')
        foreach($name in $names){
            $first=$null;try{$null=[IO.File]::ReadAllText((Join-Path $caseDir ($name+'.missing')))}catch{$first=$_}
            B-Need ($null -ne $first) 'Real first write-case cause'
            $s=New-P2BFailureState;Add-P2BFailure $s $first
            $escaped=$false;$written=$null
            try{$written=Save-P2BLedger $s $badSink @([pscustomobject]@{Number=1})}catch{$escaped=$true;throw}
            $o=[pscustomobject][ordered]@{Escaped=$escaped;ReturnedFalse=$written -is [bool] -and -not $written;PrimarySame=[object]::ReferenceEquals($s.PrimaryError,$first);SecondaryCount=$s.SecondaryErrors.Count;SecondaryDistinct=$s.SecondaryErrors.Count -eq 1 -and -not [object]::ReferenceEquals($s.SecondaryErrors[0],$first);Overflow=$s.SecondaryOverflow;Outcome=$s.Outcome;FinalLedgerWritten=$s.FinalLedgerWritten}
            $observations[$name]=$o;B-Record $name $s
            B-Need (-not $o.Escaped -and $o.ReturnedFalse -and $o.PrimarySame -and $o.SecondaryCount -eq 1 -and $o.SecondaryDistinct -and -not $o.Overflow -and $o.Outcome -ceq 'INCOMPLETE' -and -not $o.FinalLedgerWritten) $name
            B-Pass $name
        }
        $first=$null;try{$null=[IO.File]::ReadAllText((Join-Path $caseDir 'parse-first.missing'))}catch{$first=$_}
        $s=New-P2BFailureState;Add-P2BFailure $s $first
        $missing=Join-Path $caseDir 'final.missing';$badJson=Join-Path $caseDir 'bad.json'
        [IO.File]::WriteAllText($badJson,'{',[Text.UTF8Encoding]::new($false))
        $m=Complete-P2BLedger $s $missing 0 ([pscustomobject]@{Number=1}) 'child'
        $j=Complete-P2BLedger $s $badJson 0 ([pscustomobject]@{Number=1}) 'parent'
        $o=[pscustomobject][ordered]@{Escaped=$false;MissingReturnedFalse=$m -is [bool] -and -not $m;ParseReturnedFalse=$j -is [bool] -and -not $j;PrimarySame=[object]::ReferenceEquals($s.PrimaryError,$first);SecondaryCount=$s.SecondaryErrors.Count;SecondaryDistinct=$s.SecondaryErrors.Count -eq 2 -and -not [object]::ReferenceEquals($s.SecondaryErrors[0],$s.SecondaryErrors[1]);MissingNotRebuilt=-not(Test-Path -LiteralPath $missing);BadJsonUnchanged=[IO.File]::ReadAllText($badJson) -ceq '{';Overflow=$s.SecondaryOverflow;Outcome=$s.Outcome;FinalLedgerWritten=$s.FinalLedgerWritten}
        $observations.FinalReadParseFailurePreservesFirst=$o;B-Record 'FinalReadParseFailurePreservesFirst' $s
        B-Need ($o.MissingReturnedFalse -and $o.ParseReturnedFalse -and $o.PrimarySame -and $o.SecondaryCount -eq 2 -and $o.SecondaryDistinct -and $o.MissingNotRebuilt -and $o.BadJsonUnchanged -and -not $o.Overflow -and -not $o.FinalLedgerWritten) 'FinalReadParseFailurePreservesFirst'
        B-Pass 'FinalReadParseFailurePreservesFirst'
        $s=New-P2BFailureState;$written=Save-P2BLedger $s $badSink @([pscustomobject]@{Number=1})
        $leaf=if($null -ne $s.PrimaryError){$s.PrimaryError.Exception.GetBaseException()}else{$null}
        $o=[pscustomobject][ordered]@{ReturnedFalse=$written -is [bool] -and -not $written;PrimaryPresent=$null -ne $s.PrimaryError;PrimaryIsIO=$leaf -is [UnauthorizedAccessException] -or $leaf -is [IO.IOException];SecondaryCount=$s.SecondaryErrors.Count;Overflow=$s.SecondaryOverflow;Outcome=$s.Outcome;FinalLedgerWritten=$s.FinalLedgerWritten}
        $observations.FirstBookkeepingFailureLatches=$o;B-Record 'FirstBookkeepingFailureLatches' $s
        B-Need ($o.ReturnedFalse -and $o.PrimaryPresent -and $o.PrimaryIsIO -and $o.SecondaryCount -eq 0 -and -not $o.Overflow -and -not $o.FinalLedgerWritten) 'FirstBookkeepingFailureLatches'
        B-Pass 'FirstBookkeepingFailureLatches'
        $first=$null;try{$null=[IO.File]::ReadAllText((Join-Path $caseDir 'overflow-first.missing'))}catch{$first=$_}
        $s=New-P2BFailureState;Add-P2BFailure $s $first;$falseReturns=0
        for($i=0;$i -lt 9;$i++){if(-not(Save-P2BLedger $s $badSink @([pscustomobject]@{Number=1}))){$falseReturns++}}
        $o=[pscustomobject][ordered]@{FalseReturns=$falseReturns;PrimarySame=[object]::ReferenceEquals($s.PrimaryError,$first);SecondaryCount=$s.SecondaryErrors.Count;Overflow=$s.SecondaryOverflow;Outcome=$s.Outcome;FinalLedgerWritten=$s.FinalLedgerWritten}
        $observations.SecondaryOverflowStops=$o;B-Record 'SecondaryOverflowStops' $s
        B-Need ($o.FalseReturns -eq 9 -and $o.PrimarySame -and $o.SecondaryCount -eq 8 -and $o.Overflow -and -not $o.FinalLedgerWritten) 'SecondaryOverflowStops'
        B-Pass 'SecondaryOverflowStops'
        $healthy=Join-Path $caseDir 'healthy.json'
        [IO.File]::WriteAllText($healthy,'[{"Number":1,"Value":"before","Accepted":false}]',[Text.UTF8Encoding]::new($false))
        $s=New-P2BFailureState
        $written=Complete-P2BLedger $s $healthy 0 ([pscustomobject][ordered]@{Number=1;Value='after';Accepted=$false}) 'integration'
        $actual=@([IO.File]::ReadAllText($healthy)|ConvertFrom-Json)
        $exact=$actual.Count -eq 1 -and @($actual[0].PSObject.Properties.Name).Count -eq 3 -and (@($actual[0].PSObject.Properties.Name) -join ',') -ceq 'Number,Value,Accepted' -and $actual[0].Number -eq 1 -and $actual[0].Value -ceq 'after' -and $actual[0].Accepted -is [bool] -and -not $actual[0].Accepted
        $o=[pscustomobject][ordered]@{Written=$written;ExactContent=$exact;PrimaryAbsent=$null -eq $s.PrimaryError;SecondaryCount=$s.SecondaryErrors.Count;Overflow=$s.SecondaryOverflow;FinalLedgerWritten=$s.FinalLedgerWritten}
        $observations.HealthyFinalLedger=$o;B-Record 'HealthyFinalLedger' $s
        B-Need ($o.Written -and $o.ExactContent -and $o.PrimaryAbsent -and $o.SecondaryCount -eq 0 -and -not $o.Overflow -and $o.FinalLedgerWritten) 'HealthyFinalLedger'
        B-Pass 'HealthyFinalLedger'
        B-Need ($groups.Count -eq 7) 'All bookkeeping groups'
        $passed=$true
    }
}catch{Add-P2BFailure $control $_;$errorText=$_.ToString()}
try{
    $details=[pscustomobject][ordered]@{Schema=1;Mode='p2b-bookkeeping-details';Phase=$Phase;Events=$events.ToArray()}
    $null=Save-P2BLedger $control $DetailsPath $details
    $trace=if($Phase -ceq 'RED'){$redTrace}else{[pscustomobject][ordered]@{Schema=1;Mode='p2b-bookkeeping-green';Groups=$groups.ToArray();Observations=[pscustomobject]$observations;Passed=$passed;Error=$errorText;ControlPrimaryAbsent=$null -eq $control.PrimaryError;ControlSecondaryCount=$control.SecondaryErrors.Count;ControlOverflow=$control.SecondaryOverflow;CompileCount=0;ProductChanged=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}}
    $null=Save-P2BLedger $control $TracePath $trace
}catch{Add-P2BFailure $control $_}
if($null -ne $control.PrimaryError -or -not $passed){Write-P2BFailureFallback $control;exit 2}
if($Phase -ceq 'RED'){[Console]::Out.WriteLine('FAIL: FinalLedgerMustNotEscape');exit 1}
exit 0
