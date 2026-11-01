# One private holder per boundary; latch before validation or registration.
function Use-ODRecordConsumption($Consumption){
    if($Consumption -isnot [Management.Automation.PSCustomObject] -or
       (@($Consumption.PSObject.Properties.Name) -join '|') -cne 'Consumed'){
        throw 'OD: record outcome contract mismatch'
    }
    if($Consumption.Consumed -isnot [bool] -or $Consumption.Consumed){throw 'OD: record outcome contract mismatch'}
    $Consumption.Consumed=$true
}

function Assert-ODLiveErrorSequence($Primary,$Secondary,[int]$Maximum){
    if($Secondary -isnot [Management.Automation.ErrorRecord[]] -or $Secondary.Count -gt $Maximum){
        throw 'OD: record outcome contract mismatch'
    }
    foreach($record in $Secondary){if($null -eq $record){throw 'OD: record outcome contract mismatch'}}
    if($null -eq $Primary){
        if($Secondary.Count -ne 0){throw 'OD: record outcome contract mismatch'}
    }elseif($Primary -isnot [Management.Automation.ErrorRecord]){throw 'OD: record outcome contract mismatch'}
}

function Register-ODRecordOutcome($Outcome,$Consumption){
    Use-ODRecordConsumption $Consumption
    if($Outcome -isnot [Management.Automation.PSCustomObject] -or
       (@($Outcome.PSObject.Properties.Name) -join '|') -cne 'Succeeded|PrimaryError|SecondaryErrors'){
        throw 'OD: record outcome contract mismatch'
    }
    if($Outcome.Succeeded -isnot [bool]){throw 'OD: record outcome contract mismatch'}
    Assert-ODLiveErrorSequence $Outcome.PrimaryError $Outcome.SecondaryErrors 1
    if($Outcome.Succeeded -ne ($null -eq $Outcome.PrimaryError)){throw 'OD: record outcome contract mismatch'}
    if($null -ne $Outcome.PrimaryError){Capture-Failure $Outcome.PrimaryError}
    foreach($record in $Outcome.SecondaryErrors){Capture-Failure $record}
}

function Import-ODProfileFailure($Report,$Consumption){
    Use-ODRecordConsumption $Consumption
    $keys='Schema|Profile|EntryPath|EntrySHA256|KernelPath|KernelSHA256|FrozenSHA256|Status|Groups|Observations|PrimaryError|SecondaryErrors|Overflow|Closed|SourceVerified|LiveActionsPerformed|RuntimeApproved|FormalAcceptance|RealGateImplemented'
    if($Report -isnot [Management.Automation.PSCustomObject] -or
       (@($Report.PSObject.Properties.Name) -join '|') -cne $keys){throw 'OD: record outcome contract mismatch'}
    if($Report.Schema -isnot [int] -or $Report.Schema -ne 2 -or
       $Report.Profile -isnot [string] -or $Report.Profile -cnotin @('entry-migration-core','entry-migration-owned','entry-migration-live') -or
       $Report.Status -isnot [string] -or $Report.Status -cnotin @('COMPLETE','INCOMPLETE') -or
       $Report.Groups -isnot [array] -or $Report.Observations -isnot [array]){throw 'OD: record outcome contract mismatch'}
    foreach($name in @('EntryPath','EntrySHA256','KernelPath','KernelSHA256','FrozenSHA256')){
        if($Report.$name -isnot [string] -or [string]::IsNullOrEmpty($Report.$name)){throw 'OD: record outcome contract mismatch'}
    }
    foreach($name in @('Overflow','Closed','SourceVerified','LiveActionsPerformed','RuntimeApproved','FormalAcceptance','RealGateImplemented')){
        if($Report.$name -isnot [bool]){throw 'OD: record outcome contract mismatch'}
    }
    foreach($name in @('SourceVerified','LiveActionsPerformed','RuntimeApproved','FormalAcceptance','RealGateImplemented')){
        if($Report.$name){throw 'OD: record outcome contract mismatch'}
    }
    Assert-ODLiveErrorSequence $Report.PrimaryError $Report.SecondaryErrors 8
    if($Report.Status -ceq 'COMPLETE' -and ($null -ne $Report.PrimaryError -or $Report.Overflow -or -not $Report.Closed)){
        throw 'OD: record outcome contract mismatch'
    }
    if($Report.Overflow -and ($null -eq $Report.PrimaryError -or $Report.SecondaryErrors.Count -ne 8)){
        throw 'OD: record outcome contract mismatch'
    }
    if($null -ne $Report.PrimaryError){Capture-Failure $Report.PrimaryError}
    foreach($record in $Report.SecondaryErrors){Capture-Failure $record}
    if($Report.Overflow){$script:HostOverflow=$true}
}
