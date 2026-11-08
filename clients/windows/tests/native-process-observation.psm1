#requires -Version 7.0

function New-NativeCallRecord {
    param([string]$StartedUtc, [string]$EndedUtc)
    [pscustomobject][ordered]@{
        Status='invalid'; StartedUtc=$StartedUtc; EndedUtc=$EndedUtc
        ApiSucceeded=$null; RawResult=$null; WaitState=$null; ExitCode=$null
        Times=$null; LastError=$null; ExceptionType=$null
    }
}

function Get-RawMember {
    param($Value, [string]$Name)
    if ($Value -is [Collections.IDictionary]) { return $Value[$Name] }
    if ($null -ne $Value) {
        $property=$Value.PSObject.Properties[$Name]
        if ($null -ne $property) { return $property.Value }
    }
}

function Convert-NativeCall {
    param([string]$Name, $Payload, [string]$StartedUtc, [string]$EndedUtc)
    $record=New-NativeCallRecord $StartedUtc $EndedUtc
    if ($null -eq $Payload -or $Payload -is [array] -or
        ($Payload -isnot [Collections.IDictionary] -and $Payload -isnot [pscustomobject])) { return $record }
    $lastError=Get-RawMember $Payload 'LastError'
    if ($Name -ceq 'Wait') {
        $raw=Get-RawMember $Payload 'RawResult'
        if ($raw -isnot [uint32]) { return $record }
        if ($raw -eq [uint32]::MaxValue -and $lastError -isnot [int]) { return $record }
        $record.RawResult=$raw
        switch ($raw) {
            0 { $record.Status='ok'; $record.ApiSucceeded=$true; $record.WaitState='signaled' }
            258 { $record.Status='ok'; $record.ApiSucceeded=$true; $record.WaitState='timeout' }
            ([uint32]::MaxValue) { $record.Status='api-error'; $record.ApiSucceeded=$false; $record.WaitState='failed'; $record.LastError=$lastError }
            default { $record.WaitState='unexpected' }
        }
        return $record
    }
    $success=Get-RawMember $Payload 'ApiSucceeded'
    if ($success -isnot [bool]) { return $record }
    if (-not $success) {
        if ($lastError -isnot [int]) { return $record }
        $record.Status='api-error'; $record.ApiSucceeded=$false; $record.LastError=$lastError
        return $record
    }
    if ($Name -ceq 'ExitCode') {
        $code=Get-RawMember $Payload 'ExitCode'
        if ($code -isnot [uint32]) { return $record }
        $record.ExitCode=$code
    } else {
        $times=Get-RawMember $Payload 'Times'; $copy=[ordered]@{}
        foreach ($key in @('Creation','Exit','Kernel','User')) {
            $pair=Get-RawMember $times $key
            $high=Get-RawMember $pair 'High'; $low=Get-RawMember $pair 'Low'
            if ($high -isnot [uint32] -or $low -isnot [uint32]) { return $record }
            $copy[$key]=[pscustomobject]@{High=$high;Low=$low}
        }
        $record.Times=[pscustomobject]$copy
    }
    $record.Status='ok'; $record.ApiSucceeded=$true
    return $record
}

function New-NativeSnapshot {
    param([string]$Mode, [string]$CorrelationId, [string]$StartedUtc, [string]$EndedUtc, $SnapshotError, $Calls)
    [pscustomobject][ordered]@{
        SchemaVersion=1; Mode=$Mode; CorrelationId=$CorrelationId
        ObservationStartedUtc=$StartedUtc; ObservationEndedUtc=$EndedUtc
        SnapshotError=$SnapshotError; Calls=[pscustomobject]$Calls
    }
}

function Get-TalenroSyntheticNativeProcessObservation {
    <# Trusted in-memory readers only. Never loads or executes native helper. #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][Collections.IDictionary]$Readers,
        [Parameter(Mandatory)][ValidateLength(1,128)][string]$CorrelationId
    )
    $ErrorActionPreference='Stop'
    $names=@('Wait','ExitCode','Times')
    if ($null -eq $Readers -or $Readers.Count -ne 3) { throw [ArgumentException]::new('Expected three readers.') }
    foreach ($name in $names) {
        if ($name -cnotin $Readers.Keys -or $Readers[$name] -isnot [scriptblock]) { throw [ArgumentException]::new('Expected three named scriptblocks.') }
    }
    $started=[datetime]::UtcNow.ToString('o'); $calls=[ordered]@{}
    foreach ($name in $names) {
        $begin=[datetime]::UtcNow.ToString('o')
        try {
            $items=@(& $Readers[$name])
            $end=[datetime]::UtcNow.ToString('o')
            if ($items.Count -eq 1) { $calls[$name]=Convert-NativeCall $name $items[0] $begin $end }
            else { $calls[$name]=New-NativeCallRecord $begin $end }
        } catch {
            $record=New-NativeCallRecord $begin ([datetime]::UtcNow.ToString('o'))
            $record.Status='exception'; $record.ExceptionType=$_.Exception.GetType().FullName
            $calls[$name]=$record
        }
    }
    New-NativeSnapshot 'synthetic' $CorrelationId $started ([datetime]::UtcNow.ToString('o')) $null $calls
}

Export-ModuleMember -Function Get-TalenroSyntheticNativeProcessObservation
