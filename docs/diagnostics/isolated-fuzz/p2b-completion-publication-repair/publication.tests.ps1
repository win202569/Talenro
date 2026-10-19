param([Parameter(Mandatory)][ValidateSet('RED','GREEN')][string]$Phase,[Parameter(Mandatory)][string]$TracePath,[Parameter(Mandatory)][string]$DetailsPath)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-19-p2b-completion-publication-repair'
$prefix=$Phase.ToLowerInvariant()
if([IO.Path]::GetFullPath($TracePath) -cne (Join-Path $packet ($prefix+'.trace.json')) -or [IO.Path]::GetFullPath($DetailsPath) -cne (Join-Path $packet ($prefix+'.details.json'))){throw 'Fixed outputs only'}
$id=[Security.Principal.WindowsIdentity]::GetCurrent()
if(([Security.Principal.WindowsPrincipal]::new($id)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary host required'}
. (Join-Path $PSScriptRoot 'bookkeeping.ps1')
. (Join-Path $PSScriptRoot 'publication.ps1')
$control=New-P2BFailureState
$observations=[ordered]@{};$groups=[Collections.Generic.List[string]]::new();$details=[Collections.Generic.List[object]]::new()
$passed=$false;$errorText=$null
function P-Need([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)}}
function P-Record([string]$Name,$State){
    $primary=if($null -ne $State.PrimaryError){[pscustomobject]@{Type=$State.PrimaryError.GetType().FullName;ExceptionType=$State.PrimaryError.Exception.GetType().FullName;RootType=$State.PrimaryError.Exception.GetBaseException().GetType().FullName;Message=$State.PrimaryError.Exception.Message.Substring(0,[Math]::Min(1024,$State.PrimaryError.Exception.Message.Length))}}else{$null}
    $details.Add([pscustomobject]@{Name=$Name;Primary=$primary;SecondaryCount=$State.SecondaryErrors.Count;Overflow=$State.SecondaryOverflow})
}
function P-Group([string]$Name){$groups.Add($Name);if($Phase -ceq 'GREEN'){[Console]::Out.WriteLine('PASS: '+$Name)}}
try{
    $dir=Join-Path $packet ($prefix+'-publication-cases')
    if(Test-Path -LiteralPath $dir){throw 'Fresh unit directory required'}
    $null=New-Item -ItemType Directory -Path $dir
    foreach($kind in @('Missing','Malformed')){
        $path=Join-Path $dir ($kind+'.json')
        if($kind -ceq 'Malformed'){[IO.File]::WriteAllText($path,'{',[Text.UTF8Encoding]::new($false))}
        $s=New-P2BFailureState
        $row=[pscustomobject][ordered]@{Number=1;Outcome='INCOMPLETE'}
        $completed=Complete-P2BLedger $s $path 0 $row 'integration'
        $first=$s.PrimaryError
        $row.Outcome='FINAL_EXACT'
        $published=Save-P2BIntegrationPublication $s $path 0 $row @($row)
        $unchanged=if($kind -ceq 'Missing'){-not(Test-Path -LiteralPath $path)}else{[IO.File]::ReadAllText($path) -ceq '{'}
        $o=[pscustomobject][ordered]@{CompletionReturnedFalse=$completed -is [bool] -and -not $completed;PublicationReturnedFalse=$published -is [bool] -and -not $published;OriginalPreserved=$unchanged;PrimarySame=$null -ne $first -and [object]::ReferenceEquals($s.PrimaryError,$first);SecondaryCount=$s.SecondaryErrors.Count;Overflow=$s.SecondaryOverflow;Outcome=$s.Outcome}
        $name=$kind+'LedgerPreserved';$observations[$name]=$o;P-Record $name $s
        P-Need ($o.CompletionReturnedFalse -and $o.PrimarySame -and $o.SecondaryCount -eq 0 -and -not $o.Overflow -and $o.Outcome -ceq 'INCOMPLETE') $name
        if($Phase -ceq 'RED'){P-Need (-not $o.PublicationReturnedFalse -and -not $o.OriginalPreserved) ('Known RED '+$name)}else{P-Need ($o.PublicationReturnedFalse -and $o.OriginalPreserved) $name}
        P-Group $name
    }
    $path=Join-Path $dir 'output-failure.json'
    [IO.File]::WriteAllText($path,'[{"Number":1,"Outcome":"INCOMPLETE"}]',[Text.UTF8Encoding]::new($false))
    $before=[IO.File]::ReadAllText($path)
    $s=New-P2BFailureState;$row=[pscustomobject][ordered]@{Number=1;Outcome='EXPECTED_GREEN'}
    $originalOut=[Console]::Out
    $memory=[IO.MemoryStream]::new();$closed=[IO.StreamWriter]::new($memory);$closed.Dispose()
    $published=$null
    try{[Console]::SetOut($closed);$published=Publish-P2BHostOutcome $s $path @($row) $row 'FIXED-OUTPUT'}finally{[Console]::SetOut($originalOut)}
    $disk=@([IO.File]::ReadAllText($path)|ConvertFrom-Json)
    $o=[pscustomobject][ordered]@{PublicationReturnedFalse=$published -is [bool] -and -not $published;LedgerUnchanged=[IO.File]::ReadAllText($path) -ceq $before;NextSlotEligible=$disk.Count -eq 1 -and $disk[0].Outcome -ceq 'EXPECTED_GREEN';PrimaryPresent=$null -ne $s.PrimaryError;DisposedOutputCause=$null -ne $s.PrimaryError -and $s.PrimaryError.Exception.GetBaseException() -is [ObjectDisposedException];SecondaryCount=$s.SecondaryErrors.Count;Overflow=$s.SecondaryOverflow;Outcome=$s.Outcome}
    $observations.OutputFailureStopsNextSlot=$o;P-Record 'OutputFailureStopsNextSlot' $s
    P-Need ($o.PublicationReturnedFalse -and $o.PrimaryPresent -and $o.DisposedOutputCause -and $o.SecondaryCount -eq 0 -and -not $o.Overflow -and $o.Outcome -ceq 'INCOMPLETE') 'Output failure real'
    if($Phase -ceq 'RED'){P-Need (-not $o.LedgerUnchanged -and $o.NextSlotEligible) 'Known RED output gate'}else{P-Need ($o.LedgerUnchanged -and -not $o.NextSlotEligible) 'OutputFailureStopsNextSlot'}
    P-Group 'OutputFailureStopsNextSlot'
    if($Phase -ceq 'GREEN'){
        $path=Join-Path $dir 'already-failed.json';$literal='[{"Number":1,"Outcome":"INCOMPLETE"}]'
        [IO.File]::WriteAllText($path,$literal,[Text.UTF8Encoding]::new($false))
        $s=New-P2BFailureState;$anchor=$null;try{$null=[IO.File]::ReadAllText((Join-Path $dir 'first.missing'))}catch{$anchor=$_}
        Add-P2BFailure $s $anchor;$row=[pscustomobject][ordered]@{Number=1;Outcome='FINAL_EXACT'}
        $published=Save-P2BIntegrationPublication $s $path 0 $row @($row)
        $o=[pscustomobject][ordered]@{PublicationReturnedFalse=$published -is [bool] -and -not $published;LedgerUnchanged=[IO.File]::ReadAllText($path) -ceq $literal;PrimarySame=[object]::ReferenceEquals($s.PrimaryError,$anchor);SecondaryCount=$s.SecondaryErrors.Count}
        $observations.LatchedFailureDoesNotPublish=$o;P-Record 'LatchedFailureDoesNotPublish' $s
        P-Need ($o.PublicationReturnedFalse -and $o.LedgerUnchanged -and $o.PrimarySame -and $o.SecondaryCount -eq 0) 'LatchedFailureDoesNotPublish';P-Group 'LatchedFailureDoesNotPublish'
        $path=Join-Path $dir 'healthy-integration.json'
        [IO.File]::WriteAllText($path,$literal,[Text.UTF8Encoding]::new($false))
        $s=New-P2BFailureState;$row=[pscustomobject][ordered]@{Number=1;Outcome='FINAL_EXACT'}
        $published=Save-P2BIntegrationPublication $s $path 0 $row @($row)
        $disk=@([IO.File]::ReadAllText($path)|ConvertFrom-Json)
        $o=[pscustomobject][ordered]@{Published=$published -is [bool] -and $published;ExactRow=$disk.Count -eq 1 -and (@($disk[0].PSObject.Properties.Name) -join ',') -ceq 'Number,Outcome' -and $disk[0].Number -eq 1 -and $disk[0].Outcome -ceq 'FINAL_EXACT';PrimaryAbsent=$null -eq $s.PrimaryError;SecondaryCount=$s.SecondaryErrors.Count;FinalLedgerWritten=$s.FinalLedgerWritten}
        $observations.HealthyIntegrationPublication=$o;P-Record 'HealthyIntegrationPublication' $s
        P-Need ($o.Published -and $o.ExactRow -and $o.PrimaryAbsent -and $o.SecondaryCount -eq 0 -and $o.FinalLedgerWritten) 'HealthyIntegrationPublication';P-Group 'HealthyIntegrationPublication'
        $path=Join-Path $dir 'healthy-host.json';[IO.File]::WriteAllText($path,$literal,[Text.UTF8Encoding]::new($false))
        $s=New-P2BFailureState;$row=[pscustomobject][ordered]@{Number=1;Outcome='EXPECTED_GREEN'}
        $writer=[IO.StringWriter]::new([Globalization.CultureInfo]::InvariantCulture);$originalOut=[Console]::Out
        try{[Console]::SetOut($writer);$published=Publish-P2BHostOutcome $s $path @($row) $row 'FIXED-OUTPUT'}finally{[Console]::SetOut($originalOut)}
        $captured=$writer.ToString();$writer.Dispose()
        $disk=@([IO.File]::ReadAllText($path)|ConvertFrom-Json)
        $lf=[string][char]10;$wanted='{'+$lf+'  "Number": 1,'+$lf+'  "Outcome": "EXPECTED_GREEN"'+$lf+'}'+$lf+'FIXED-OUTPUT'
        $o=[pscustomobject][ordered]@{Published=$published -is [bool] -and $published;ExactOutput=$captured.Replace(([string][char]13+[char]10),$lf) -ceq $wanted;NextSlotEligible=$disk.Count -eq 1 -and $disk[0].Number -eq 1 -and $disk[0].Outcome -ceq 'EXPECTED_GREEN';PrimaryAbsent=$null -eq $s.PrimaryError;SecondaryCount=$s.SecondaryErrors.Count}
        $observations.HealthyHostPublication=$o;P-Record 'HealthyHostPublication' $s
        P-Need ($o.Published -and $o.ExactOutput -and $o.NextSlotEligible -and $o.PrimaryAbsent -and $o.SecondaryCount -eq 0) 'HealthyHostPublication';P-Group 'HealthyHostPublication'
    }
    $passed=$true
}catch{Add-P2BFailure $control $_;try{$errorText=$_.ToString()}catch{Add-P2BFailure $control $_;$errorText='summary unavailable'}}
$trace=[pscustomobject][ordered]@{Schema=1;Mode=('p2b-publication-'+$prefix);Groups=$groups.ToArray();Observations=[pscustomobject]$observations;Passed=$passed;Error=$errorText;ControlPrimaryAbsent=$null -eq $control.PrimaryError;StartCalls=0;CompileCount=0;ProductChanged=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
$null=Save-P2BLedger $control $DetailsPath ([pscustomobject][ordered]@{Schema=1;Mode='p2b-publication-details';Phase=$Phase;Events=$details.ToArray()})
$null=Save-P2BLedger $control $TracePath $trace
if($null -ne $control.PrimaryError -or -not $passed){Write-P2BFailureFallback $control;exit 2}
if($Phase -ceq 'RED'){[Console]::Out.WriteLine('FAIL: MissingLedgerMustRemainMissing');[Console]::Out.WriteLine('FAIL: MalformedLedgerMustRemainUnchanged');[Console]::Out.WriteLine('FAIL: OutputFailureMustNotOpenNextSlot');exit 1}
exit 0
