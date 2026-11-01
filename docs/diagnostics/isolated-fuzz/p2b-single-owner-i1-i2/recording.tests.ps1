# Actual reference observations stay outside injected case-module scopes.
function Start-ODRecorderGuards {
    $refs=@{}
    foreach($name in @('New-ODRecordStream','Write-ODCaptureRecord','Write-ODProfileIntentRecord','Register-ODRecordOutcome','Import-ODProfileFailure','Capture-Failure')){$refs[$name]=(Get-Command $name -CommandType Function).ScriptBlock}
    $guard=[pscustomobject]@{Refs=$refs;Primary=$script:HostPrimary;Secondary=$script:HostSecondary;Overflow=$script:HostOverflow;Stages=[Collections.Generic.List[object]]::new()}
    Observe-ODRecorderGuard $guard Before ''
    return $guard
}
function Observe-ODRecorderGuard($Guard,[string]$Stage,[string]$Case){
    if($Guard.Stages.Count -ge 32){throw 'Recorder stage bound'}
    $rows=@(foreach($name in @('New-ODRecordStream','Write-ODCaptureRecord','Write-ODProfileIntentRecord','Register-ODRecordOutcome','Import-ODProfileFailure','Capture-Failure')){
        $command=Get-Command $name -CommandType Function -ErrorAction Stop
        [ordered]@{Name=$name;SameOriginal=[object]::ReferenceEquals($Guard.Refs[$name],$command.ScriptBlock);ModuleName=$command.ModuleName}
    })
    $collectorSame=[object]::ReferenceEquals($Guard.Primary,$script:HostPrimary) -and [object]::ReferenceEquals($Guard.Secondary,$script:HostSecondary) -and $Guard.Overflow -ceq $script:HostOverflow
    $Guard.Stages.Add([ordered]@{Stage=$Stage;Case=$Case;Commands=$rows;CollectorSame=$collectorSame})
}
function Complete-ODRecorderGuards($Guard,$Publisher,$CaseModules){
    Observe-ODRecorderGuard $Guard After ''
    $allSame=@($Guard.Stages | ForEach-Object {$_.Commands} | Where-Object {-not $_.SameOriginal -or $_.ModuleName -cne ''}).Count -eq 0 -and @($Guard.Stages|Where-Object {-not $_.CollectorSame}).Count -eq 0
    LS-Result RecorderIdentityIsolation $allSame ([ordered]@{AllSame=$allSame})
    $publisherView=& $Publisher {
        $w=Get-Command Write-ODCaptureRecord -CommandType Function;$f=Get-Command New-ODRecordStream -CommandType Function
        $absent=$true
        foreach($name in @('HostPrimary','HostSecondary','HostOverflow')){if($null -ne (Get-Variable -Name $name -Scope Script -ErrorAction Ignore)){$absent=$false}}
        [ordered]@{Name=$ExecutionContext.SessionState.Module.Name;Path=$ExecutionContext.SessionState.Module.Path;WriterFile=$w.ScriptBlock.File;FactoryFile=$f.ScriptBlock.File;WriterSHA256=(Get-FileHash -LiteralPath $w.ScriptBlock.File).Hash;DefaultFactory=[object]::ReferenceEquals($script:DefaultFactory,$f.ScriptBlock);CollectorStateAbsent=$absent}
    }
    $publisherExports=$Publisher.ExportedFunctions.Count+$Publisher.ExportedAliases.Count+$Publisher.ExportedVariables.Count+$Publisher.ExportedCmdlets.Count
    $clean=$publisherView.CollectorStateAbsent
    $cases=@(foreach($entry in $CaseModules){
        $m=$entry.Module
        $state=& $m {[ordered]@{Initialized=($null -ne (Get-Variable HostPrimary -Scope Script -ErrorAction Ignore) -and $null -ne (Get-Variable HostSecondary -Scope Script -ErrorAction Ignore) -and $null -ne (Get-Variable HostOverflow -Scope Script -ErrorAction Ignore));PrimaryNull=($null -eq $script:HostPrimary);SecondaryCount=$script:HostSecondary.Count;Overflow=$script:HostOverflow}}
        [ordered]@{Boundary=$entry.Boundary;Case=$entry.Case;ModuleName=$m.Name;Exports=[ordered]@{Functions=$m.ExportedFunctions.Count;Aliases=$m.ExportedAliases.Count;Variables=$m.ExportedVariables.Count;Cmdlets=$m.ExportedCmdlets.Count};State=$state}
    })
    $private=$publisherExports -eq 0 -and $clean -and @($cases|Where-Object {$_.Exports.Functions -ne 0 -or $_.Exports.Aliases -ne 0 -or $_.Exports.Variables -ne 0 -or $_.Exports.Cmdlets -ne 0 -or -not $_.State.Initialized -or $_.State.Overflow}).Count -eq 0
    LS-Result RecorderPrivateState $private ([ordered]@{PublisherExports=$publisherExports;CollectorStateAbsent=$clean;Cases=$cases.Count})
    $source=[IO.Path]::GetFullPath((Join-Path $repo 'scripts/offline-diagnostics-v4/private/write-boundary.ps1'))
    $sourceBound=$publisherView.WriterFile -ceq $source -and $publisherView.FactoryFile -ceq $source -and $publisherView.WriterSHA256 -ceq (Get-FileHash -LiteralPath $source).Hash
    $probePath=Join-Path $script:CaptureRoot publisher-probe.json
    $bytes=[Text.UTF8Encoding]::new($false).GetBytes('{"Recorder":"v4","Healthy":true}')
    $outcome=& $Publisher {param($Path,$Bytes)Write-ODCaptureRecord $Path $Bytes} $probePath $bytes
    $consumption=[pscustomobject]@{Consumed=$false}
    Register-ODRecordOutcome $outcome $consumption
    $outcomeSucceeded=$outcome.Succeeded -is [bool] -and $outcome.Succeeded -and $null -eq $outcome.PrimaryError -and $outcome.SecondaryErrors -is [Management.Automation.ErrorRecord[]] -and $outcome.SecondaryErrors.Count -eq 0
    $probeExact=[Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes($probePath),$bytes)
    LS-Result RecorderProductionBinding ($publisherView.DefaultFactory -and $sourceBound -and $probeExact -and $outcomeSucceeded) ([ordered]@{DefaultFactory=$publisherView.DefaultFactory;ProductionSource=$sourceBound;ProbeExact=$probeExact;OutcomeSucceeded=$outcomeSucceeded})
    LS-Details Recorder ([ordered]@{Stages=$Guard.Stages.ToArray();Publisher=$publisherView;Cases=$cases;ProbePath=$probePath;ProbeBytes=[Convert]::ToBase64String([IO.File]::ReadAllBytes($probePath));Outcome=$outcome;Consumed=$consumption.Consumed})
}
