# Actual reference observations stay outside injected case-module scopes.
function Start-ODRecorderGuards {
    $refs=@{}
    foreach($name in @('New-ODRecordStream','Write-ODCaptureRecord','Write-ODProfileIntentRecord','Capture-Failure')){$refs[$name]=(Get-Command $name -CommandType Function).ScriptBlock}
    $guard=[pscustomobject]@{Refs=$refs;Stages=[Collections.Generic.List[object]]::new()}
    Observe-ODRecorderGuard $guard Before ''
    return $guard
}
function Observe-ODRecorderGuard($Guard,[string]$Stage,[string]$Case){
    if($Guard.Stages.Count -ge 32){throw 'Recorder stage bound'}
    $rows=@(foreach($name in @('New-ODRecordStream','Write-ODCaptureRecord','Write-ODProfileIntentRecord','Capture-Failure')){
        $command=Get-Command $name -CommandType Function -ErrorAction Stop
        [ordered]@{Name=$name;SameOriginal=[object]::ReferenceEquals($Guard.Refs[$name],$command.ScriptBlock);ModuleName=$command.ModuleName}
    })
    $Guard.Stages.Add([ordered]@{Stage=$Stage;Case=$Case;Commands=$rows})
}
function Complete-ODRecorderGuards($Guard,$Publisher,$CaseModules){
    Observe-ODRecorderGuard $Guard After ''
    $allSame=@($Guard.Stages | ForEach-Object {$_.Commands} | Where-Object {-not $_.SameOriginal -or $_.ModuleName -cne ''}).Count -eq 0
    LS-Result RecorderIdentityIsolation $allSame ([ordered]@{AllSame=$allSame})
    $publisherView=& $Publisher {
        $w=Get-Command Write-ODCaptureRecord -CommandType Function;$f=Get-Command New-ODRecordStream -CommandType Function
        [ordered]@{Name=$ExecutionContext.SessionState.Module.Name;Path=$ExecutionContext.SessionState.Module.Path;WriterFile=$w.ScriptBlock.File;FactoryFile=$f.ScriptBlock.File;WriterSHA256=(Get-FileHash -LiteralPath $w.ScriptBlock.File).Hash;DefaultFactory=[object]::ReferenceEquals($script:DefaultFactory,$f.ScriptBlock);PrimaryNull=($null -eq $script:HostPrimary);SecondaryCount=$script:HostSecondary.Count;Overflow=$script:HostOverflow}
    }
    $publisherExports=$Publisher.ExportedFunctions.Count+$Publisher.ExportedAliases.Count+$Publisher.ExportedVariables.Count+$Publisher.ExportedCmdlets.Count
    $clean=$publisherView.PrimaryNull -and $publisherView.SecondaryCount -eq 0 -and -not $publisherView.Overflow
    $cases=@(foreach($entry in $CaseModules){
        $m=$entry.Module
        $state=& $m {[ordered]@{Initialized=($null -ne (Get-Variable HostPrimary -Scope Script -ErrorAction Ignore) -and $null -ne (Get-Variable HostSecondary -Scope Script -ErrorAction Ignore) -and $null -ne (Get-Variable HostOverflow -Scope Script -ErrorAction Ignore));PrimaryNull=($null -eq $script:HostPrimary);SecondaryCount=$script:HostSecondary.Count;Overflow=$script:HostOverflow}}
        [ordered]@{Boundary=$entry.Boundary;Case=$entry.Case;ModuleName=$m.Name;Exports=[ordered]@{Functions=$m.ExportedFunctions.Count;Aliases=$m.ExportedAliases.Count;Variables=$m.ExportedVariables.Count;Cmdlets=$m.ExportedCmdlets.Count};State=$state}
    })
    $private=$publisherExports -eq 0 -and $clean -and @($cases|Where-Object {$_.Exports.Functions -ne 0 -or $_.Exports.Aliases -ne 0 -or $_.Exports.Variables -ne 0 -or $_.Exports.Cmdlets -ne 0 -or -not $_.State.Initialized -or $_.State.Overflow}).Count -eq 0
    LS-Result RecorderPrivateState $private ([ordered]@{PublisherExports=$publisherExports;PublisherClean=$clean;Cases=$cases.Count})
    $source=[IO.Path]::GetFullPath((Join-Path $repo 'scripts/offline-diagnostics-v3/private/write-boundary.ps1'))
    $sourceBound=$publisherView.WriterFile -ceq $source -and $publisherView.FactoryFile -ceq $source -and $publisherView.WriterSHA256 -ceq (Get-FileHash -LiteralPath $source).Hash
    $probePath=Join-Path $script:CaptureRoot publisher-probe.json
    $bytes=[Text.UTF8Encoding]::new($false).GetBytes('{"Recorder":"v3","Healthy":true}')
    & $Publisher {param($Path,$Bytes)Write-ODCaptureRecord $Path $Bytes} $probePath $bytes
    $probeExact=[Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes($probePath),$bytes)
    LS-Result RecorderProductionBinding ($publisherView.DefaultFactory -and $sourceBound -and $probeExact) ([ordered]@{DefaultFactory=$publisherView.DefaultFactory;ProductionSource=$sourceBound;ProbeExact=$probeExact})
    LS-Details Recorder ([ordered]@{Stages=$Guard.Stages.ToArray();Publisher=$publisherView;Cases=$cases;ProbePath=$probePath;ProbeBytes=[Convert]::ToBase64String([IO.File]::ReadAllBytes($probePath))})
}
