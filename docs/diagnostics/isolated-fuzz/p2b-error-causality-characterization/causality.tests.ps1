param([Parameter(Mandatory)][string]$TracePath,[Parameter(Mandatory)][string]$DetailsPath)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-19-p2b-error-causality-characterization'
if([IO.Path]::GetFullPath($TracePath) -cne (Join-Path $packet 'diagnostic.trace.json') -or [IO.Path]::GetFullPath($DetailsPath) -cne (Join-Path $packet 'diagnostic.details.json')){throw 'Fixed diagnostic output paths required'}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
if(([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary diagnostic required'}
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
if((Get-FileHash -LiteralPath $engine).Hash -cne '362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'){throw 'Engine identity'}
$groups=[Collections.Generic.List[string]]::new()
$observations=[ordered]@{}
$events=[Collections.Generic.List[object]]::new()
function C-Need([bool]$Condition,[string]$Name){if(-not $Condition){throw ('ASSERTION FAILED: '+$Name)}}
function C-Pass([string]$Name){$groups.Add($Name);[Console]::Out.WriteLine('PASS: '+$Name)}
function C-SameIOCause([object]$Left,[object]$Right){
    if($Left -isnot [Management.Automation.ErrorRecord] -or $Right -isnot [Management.Automation.ErrorRecord]){return $false}
    [Exception[]]$leaves=@($null,$null)
    $records=[object[]]@($Left,$Right)
    for($side=0;$side -lt 2;$side++){
        [Management.Automation.ErrorRecord]$record=$records[$side]
        [Exception]$node=$record.Exception
        $seen=[Collections.Generic.List[Exception]]::new()
        while($null -ne $node){
            if($seen.Count -ge 8){return $false}
            foreach($prior in $seen){if([object]::ReferenceEquals($prior,$node)){return $false}}
            $seen.Add($node)
            if($node -is [UnauthorizedAccessException] -or $node -is [IO.IOException]){$leaves[$side]=$node}
            $node=$node.InnerException
        }
        if($null -eq $leaves[$side]){return $false}
    }
    return [object]::ReferenceEquals($leaves[0],$leaves[1])
}
function C-PrimaryPreserved([object]$Result,[object]$FirstAnchor,[object]$FinalAnchor){
    if($null -eq $Result -or $Result -isnot [pscustomobject]){return $false}
    $primaryProperty=$Result.PSObject.Properties['PrimaryError']
    $cleanupProperty=$Result.PSObject.Properties['CleanupErrors']
    if($primaryProperty -isnot [Management.Automation.PSNoteProperty] -or $cleanupProperty -isnot [Management.Automation.PSNoteProperty]){return $false}
    if($Result.CleanupErrors -isnot [Management.Automation.ErrorRecord[]] -or $Result.CleanupErrors.Count -ne 1){return $false}
    return ((C-SameIOCause $Result.PrimaryError $FirstAnchor) -and
        -not (C-SameIOCause $Result.PrimaryError $FinalAnchor) -and
        (C-SameIOCause $Result.CleanupErrors[0] $FinalAnchor) -and
        -not (C-SameIOCause $Result.CleanupErrors[0] $FirstAnchor))
}
function Write-Receipt([string]$Path){
    [IO.File]::WriteAllText($Path,'fixed-receipt',[Text.UTF8Encoding]::new($false))
}
function C-Describe([Management.Automation.ErrorRecord]$Record){
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
$passed=$false;$errorText=$null
try{
    $badPath=Join-Path $packet 'receipt-is-directory'
    if(Test-Path -LiteralPath $badPath){throw 'Fresh fault directory required'}
    $null=New-Item -ItemType Directory -Path $badPath
    $firstCatch=$null;$secondCatch=$null
    $Error.Clear()
    try{Write-Receipt $badPath}catch{$firstCatch=$_}
    C-Need ($null -ne $firstCatch -and $Error.Count -gt 0) 'First real write failed'
    $firstAnchor=$Error[0]
    $firstOldest=$Error[$Error.Count-1]
    $firstErrorCount=$Error.Count
    try{Write-Receipt $badPath}catch{$secondCatch=$_}
    C-Need ($null -ne $secondCatch -and $Error.Count -gt 0) 'Second real write failed'
    $finalAnchor=$Error[0]
    $finalOldest=$Error[$Error.Count-1]
    $finalErrorCount=$Error.Count
    $events.Add([pscustomobject][ordered]@{Stage='first-write';Catch=(C-Describe $firstCatch);Anchor=(C-Describe $firstAnchor);ErrorCount=$firstErrorCount;OldestRecordMatches=[object]::ReferenceEquals($firstCatch,$firstOldest);OldestCauseMatches=(C-SameIOCause $firstCatch $firstOldest)})
    $events.Add([pscustomobject][ordered]@{Stage='final-write';Catch=(C-Describe $secondCatch);Anchor=(C-Describe $finalAnchor);ErrorCount=$finalErrorCount;OldestRecordMatches=[object]::ReferenceEquals($secondCatch,$finalOldest);OldestCauseMatches=(C-SameIOCause $secondCatch $finalOldest)})
    [Exception]$firstLeaf=$firstCatch.Exception
    [Exception]$secondLeaf=$secondCatch.Exception
    $firstDepth=1;$secondDepth=1
    while($null -ne $firstLeaf.InnerException -and $firstDepth -le 8){$firstLeaf=$firstLeaf.InnerException;$firstDepth++}
    while($null -ne $secondLeaf.InnerException -and $secondDepth -le 8){$secondLeaf=$secondLeaf.InnerException;$secondDepth++}
    C-Need ($firstDepth -le 8 -and $secondDepth -le 8 -and ($firstLeaf -is [UnauthorizedAccessException] -or $firstLeaf -is [IO.IOException]) -and ($secondLeaf -is [UnauthorizedAccessException] -or $secondLeaf -is [IO.IOException])) 'Bounded allowed real IO leaves'

    $o=[pscustomobject][ordered]@{RecordReferencesDiffer=-not [object]::ReferenceEquals($firstCatch,$firstAnchor);SameCause=(C-SameIOCause $firstCatch $firstAnchor)}
    $observations.CatchCopyNotIdentity=$o
    C-Need ($o.RecordReferencesDiffer -and $o.SameCause) 'CatchCopyNotIdentity'
    C-Pass 'CatchCopyNotIdentity'

    $o=[pscustomobject][ordered]@{FirstMatches=(C-SameIOCause $firstCatch $firstAnchor);FinalMatches=(C-SameIOCause $secondCatch $finalAnchor);FirstVsFinal=(C-SameIOCause $firstCatch $finalAnchor)}
    $observations.OriginalCauseRetained=$o
    C-Need ($o.FirstMatches -and $o.FinalMatches -and -not $o.FirstVsFinal) 'OriginalCauseRetained'
    C-Pass 'OriginalCauseRetained'

    $o=[pscustomobject][ordered]@{MessagesNonemptyEqual=$firstLeaf.Message.Length -gt 0 -and $firstLeaf.Message -ceq $secondLeaf.Message;LeafReferencesDiffer=-not [object]::ReferenceEquals($firstLeaf,$secondLeaf);SameCause=(C-SameIOCause $firstCatch $secondCatch)}
    $observations.DistinctEqualText=$o
    C-Need ($o.MessagesNonemptyEqual -and $o.LeafReferencesDiffer -and -not $o.SameCause) 'DistinctEqualText'
    C-Pass 'DistinctEqualText'

    $original=[pscustomobject]@{PrimaryError=$firstCatch;CleanupErrors=[Management.Automation.ErrorRecord[]]@($secondCatch)}
    $replaced=[pscustomobject]@{PrimaryError=$secondCatch;CleanupErrors=[Management.Automation.ErrorRecord[]]@($secondCatch)}
    $swapped=[pscustomobject]@{PrimaryError=$secondCatch;CleanupErrors=[Management.Automation.ErrorRecord[]]@($firstCatch)}
    $cleanupReplaced=[pscustomobject]@{PrimaryError=$firstCatch;CleanupErrors=[Management.Automation.ErrorRecord[]]@($firstCatch)}
    $o=[pscustomobject][ordered]@{Original=(C-PrimaryPreserved $original $firstAnchor $finalAnchor);PrimaryReplaced=(C-PrimaryPreserved $replaced $firstAnchor $finalAnchor);Swapped=(C-PrimaryPreserved $swapped $firstAnchor $finalAnchor);CleanupReplaced=(C-PrimaryPreserved $cleanupReplaced $firstAnchor $finalAnchor)}
    $observations.ValidatorRejectsReplacement=$o
    C-Need ($o.Original -and -not $o.PrimaryReplaced -and -not $o.Swapped -and -not $o.CleanupReplaced) 'ValidatorRejectsReplacement'
    C-Pass 'ValidatorRejectsReplacement'

    [Exception]$deep=$firstLeaf
    for($i=0;$i -lt 9;$i++){$deep=[Management.Automation.RuntimeException]::new('diagnostic wrapper',$deep)}
    $deepRecord=[Management.Automation.ErrorRecord]::new($deep,'CausalityDeep',[Management.Automation.ErrorCategory]::NotSpecified,$null)
    $missing=[pscustomobject]@{CleanupErrors=[Management.Automation.ErrorRecord[]]@($secondCatch)}
    $o=[pscustomobject][ordered]@{NullAnchor=(C-SameIOCause $firstCatch $null);MissingPrimary=(C-PrimaryPreserved $missing $firstAnchor $finalAnchor);DeepNine=(C-SameIOCause $deepRecord $firstCatch)}
    $observations.MissingOrUnboundedCauseRejected=$o
    C-Need (-not $o.NullAnchor -and -not $o.MissingPrimary -and -not $o.DeepNine) 'MissingOrUnboundedCauseRejected'
    C-Pass 'MissingOrUnboundedCauseRejected'
    C-Need ($groups.Count -eq 5 -and $events.Count -eq 2) 'Completion'
    $passed=$true
}catch{$errorText=$_.ToString()}
$details=[pscustomobject][ordered]@{Schema=1;Mode='p2b-error-causality-details';Events=$events.ToArray()}
[IO.File]::WriteAllText($DetailsPath,($details|ConvertTo-Json -Depth 20),[Text.UTF8Encoding]::new($false))
$trace=[pscustomobject][ordered]@{Schema=1;Mode='p2b-error-causality-characterization';Groups=$groups.ToArray();Observations=[pscustomobject]$observations;Passed=$passed;Error=$errorText;CompileCount=0;ProductChanged=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
[IO.File]::WriteAllText($TracePath,($trace|ConvertTo-Json -Depth 20),[Text.UTF8Encoding]::new($false))
if(-not $passed){[Console]::Error.WriteLine($errorText);exit 1}
exit 0
