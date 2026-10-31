# Regression: default exports shadow outer functions; no-members exports must not.
function Test-RIRealError($Record,[string]$Path){
    return $null -ne $Record -and $Record.Exception.GetType().FullName -ceq 'System.Management.Automation.MethodInvocationException' -and $Record.FullyQualifiedErrorId -ceq 'IOException' -and $Record.Exception.Message -ceq ('Exception calling "Open" with "4" argument(s): "The file ''{0}'' already exists."' -f $Path)
}
function Invoke-RIIsolationTests($PublisherModule,$Publishers){
    $riNames=@('New-ODRecordStream','Write-ODCaptureRecord','Write-ODProfileIntentRecord','Capture-Failure')
    $riOriginals=@{}
    foreach($riName in $riNames){$riOriginals[$riName]=(Get-Command $riName -CommandType Function).ScriptBlock}
    function Read-RICommands {
        foreach($riName in $riNames){
            $riCommand=Get-Command $riName -CommandType Function -ErrorAction Stop
            [ordered]@{Name=$riName;SameOriginal=[object]::ReferenceEquals($riOriginals[$riName],$riCommand.ScriptBlock);ModuleName=$riCommand.ModuleName;ModuleGuid=$(if($null -eq $riCommand.Module){$null}else{$riCommand.Module.Guid.ToString()})}
        }
    }
    $riBefore=@(Read-RICommands)
    # Dot sourcing retains the same active caller scope as the original test.
    . (Join-Path $PSScriptRoot case-module.ps1) $riWriter $riCollectorText
    $riCreated=@(Read-RICommands)
    $riCasePath=Join-Path $riPhaseRoot case-existing.bin
    $riCase=& $riCaseModule { # RI-CALL
        param($Path)
        $script:HostPrimary=$null
        $script:HostSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new()
        $script:HostOverflow=$false
        $script:RIIsolationSentinel='isolation-private-20261031'
        $script:RICalls=0
        $riOldFactory=(Get-Command New-ODRecordStream).ScriptBlock
        function script:New-ODRecordStream([string]$Path){
            $script:RICalls++
            & $script:OriginalFactory $Path # RI-CALL
        }
        $riFactoryChanged=-not [object]::ReferenceEquals($riOldFactory,(Get-Command New-ODRecordStream).ScriptBlock)
        $riThrown=$null
        try{Write-ODCaptureRecord $Path ([byte[]]@(65,66,0,255))}catch{$riThrown=$_;Capture-Failure $_} # RI-CALL
        [pscustomobject][ordered]@{Thrown=$riThrown;Primary=$script:HostPrimary;SameError=[object]::ReferenceEquals($riThrown,$script:HostPrimary);Factory=[ordered]@{Changed=$riFactoryChanged;Calls=$script:RICalls};State=[ordered]@{PrimaryNull=($null -eq $script:HostPrimary);SecondaryCount=$script:HostSecondary.Count;Overflow=$script:HostOverflow;SentinelPresent=($null -ne (Get-Variable RIIsolationSentinel -Scope Script -ErrorAction Ignore));SentinelValue=$script:RIIsolationSentinel}}
    } $riCasePath # RI-CALL
    $riReplaced=@(Read-RICommands)
    $riSentinel=Get-Variable RIIsolationSentinel -Scope Script -ErrorAction Ignore
    $riOuter=[ordered]@{PrimaryNull=($null -eq $script:HostPrimary);SecondaryCount=$script:HostSecondary.Count;Overflow=$script:HostOverflow;SentinelPresent=($null -ne $riSentinel);SentinelValue=$(if($null -eq $riSentinel){$null}else{$riSentinel.Value})}
    $riModuleView=[ordered]@{Name=$riCaseModule.Name;Guid=$riCaseModule.Guid.ToString();Functions=$riCaseModule.ExportedFunctions.Count;Aliases=$riCaseModule.ExportedAliases.Count;Variables=$riCaseModule.ExportedVariables.Count;Cmdlets=$riCaseModule.ExportedCmdlets.Count}
    for($riIndex=0;$riIndex -lt 4;$riIndex++){Add-RIGroup $riOracle.Names[$riIndex] $riCreated[$riIndex].SameOriginal ([ordered]@{SameOriginal=$riCreated[$riIndex].SameOriginal})}
    $riPrivate=$riCase.State.SentinelPresent -and $riCase.State.SentinelValue -ceq 'isolation-private-20261031' -and -not $riOuter.SentinelPresent -and $riOuter.PrimaryNull -and $riOuter.SecondaryCount -eq 0 -and -not $riOuter.Overflow
    Add-RIGroup InnerStatePrivate $riPrivate ([ordered]@{InnerSentinel=$riCase.State.SentinelPresent;OuterSentinel=$riOuter.SentinelPresent;OuterPrimaryNull=$riOuter.PrimaryNull;OuterSecondaryCount=$riOuter.SecondaryCount;OuterOverflow=$riOuter.Overflow})
    Add-RIGroup FactoryReplacementPrivate ($riCase.Factory.Changed -and $riCase.Factory.Calls -eq 1 -and $riReplaced[0].SameOriginal) ([ordered]@{InnerChanged=$riCase.Factory.Changed;Calls=$riCase.Factory.Calls;OuterSame=$riReplaced[0].SameOriginal})
    $riErrorCorrect=Test-RIRealError $riCase.Thrown $riCasePath
    Add-RIGroup RealCreateNewErrorRetained ($riCase.SameError -and $riCase.State.SecondaryCount -eq 0 -and -not $riCase.State.PrimaryNull -and -not $riCase.State.Overflow -and $riErrorCorrect) ([ordered]@{SameError=$riCase.SameError;SecondaryCount=$riCase.State.SecondaryCount;SignatureCorrect=$riErrorCorrect})
    $riHealthyBytes=[Text.UTF8Encoding]::new($false).GetBytes('{"Case":"PublisherHealthy","Value":1}')
    $riHealthy=& $PublisherModule {param($Publisher,$Bytes)Write-RIPublisherBytes $Publisher $Bytes} $Publishers.Healthy $riHealthyBytes # RI-CALL
    $riHealthyView=& $PublisherModule {param($Publisher)Get-RIPublisherView $Publisher} $Publishers.Healthy
    $riHealthyExact=$riHealthy.Ok -and -not $riHealthy.Rejected -and $null -eq $riHealthy.Error -and [Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes($riHealthyView.Path),$riHealthyBytes)
    $riPublisherExports=$PublisherModule.ExportedFunctions.Count+$PublisherModule.ExportedAliases.Count+$PublisherModule.ExportedVariables.Count+$PublisherModule.ExportedCmdlets.Count
    Add-RIGroup IndependentPublisherHealthy ($riHealthyExact -and $riHealthyView.Status -ceq 'Completed' -and $riHealthyView.Open -eq 1 -and $riHealthyView.Write -eq 1 -and $riHealthyView.Flush -eq 1 -and $riHealthyView.Dispose -eq 1 -and $riPublisherExports -eq 0) ([ordered]@{Ok=$riHealthy.Ok;Rejected=$riHealthy.Rejected;BytesExact=$riHealthyExact;Status=$riHealthyView.Status;Open=$riHealthyView.Open;Write=$riHealthyView.Write;Flush=$riHealthyView.Flush;Dispose=$riHealthyView.Dispose;Exports=$riPublisherExports})
    $riFirst=& $PublisherModule {param($Publisher)Write-RIPublisherBytes $Publisher ([byte[]]@(88))} $Publishers.ExistingFailure # RI-CALL
    $riFailedBefore=& $PublisherModule {param($Publisher)Get-RIPublisherView $Publisher} $Publishers.ExistingFailure
    $riSecond=& $PublisherModule {param($Publisher)Write-RIPublisherBytes $Publisher ([byte[]]@(89))} $Publishers.ExistingFailure # RI-CALL
    $riFailedView=& $PublisherModule {param($Publisher)Get-RIPublisherView $Publisher} $Publishers.ExistingFailure
    $riDelta=($riFailedView.Open+$riFailedView.Write+$riFailedView.Flush+$riFailedView.Dispose)-($riFailedBefore.Open+$riFailedBefore.Write+$riFailedBefore.Flush+$riFailedBefore.Dispose)
    $riFailedSignature=Test-RIRealError $riFirst.Error $riFailedView.Path
    $riSamePublisherError=[object]::ReferenceEquals($riFirst.Error,$riFailedView.Primary)
    Add-RIGroup FailedPublisherLatched (-not $riFirst.Ok -and -not $riFirst.Rejected -and $riFailedSignature -and $riSamePublisherError -and -not $riSecond.Ok -and $riSecond.Rejected -and $null -eq $riSecond.Error -and $riDelta -eq 0 -and $riFailedView.Status -ceq 'Failed' -and $riFailedView.Open -eq 1 -and $riFailedView.Write -eq 0 -and $riFailedView.Flush -eq 0 -and $riFailedView.Dispose -eq 0) ([ordered]@{FirstOk=$riFirst.Ok;FirstRejected=$riFirst.Rejected;SignatureCorrect=$riFailedSignature;SameError=$riSamePublisherError;SecondOk=$riSecond.Ok;SecondRejected=$riSecond.Rejected;SecondErrorNull=($null -eq $riSecond.Error);IOAfterSecond=$riDelta;Status=$riFailedView.Status;Open=$riFailedView.Open;Write=$riFailedView.Write;Flush=$riFailedView.Flush;Dispose=$riFailedView.Dispose})
    $script:RIObservations=[ordered]@{CommandsBefore=$riBefore;CommandsCreated=$riCreated;CommandsReplaced=$riReplaced;CaseModule=$riModuleView;OuterState=$riOuter;InnerState=$riCase.State;Factory=$riCase.Factory;CaseThrown=$riCase.Thrown;CasePrimary=$riCase.Primary;HealthyPublisher=[ordered]@{Result=$riHealthy;View=$riHealthyView;Bytes=[Convert]::ToBase64String([IO.File]::ReadAllBytes($riHealthyView.Path));Exports=$riPublisherExports};FailedPublisher=[ordered]@{First=$riFirst;Second=$riSecond;View=$riFailedView;IOAfterSecond=$riDelta;SameError=$riSamePublisherError}}
}
