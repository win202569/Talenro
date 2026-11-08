#requires -Version 7.0

function Read-ObservationField {
    param([scriptblock]$Reader, [string]$Name)
    $ErrorActionPreference = 'Stop'
    try {
        # Capture all success output; neither false nor zero is a missing value.
        $items = @(& $Reader)
        $valid = $false
        $value = $null
        if ($items.Count -eq 1) {
            $candidate = $items[0]
            $valid = switch ($Name) {
                'ProcessId' { $candidate -is [int] -and $candidate -gt 0 }
                'StartTime' { $candidate -is [datetime] }
                'WaitSignaled' { $candidate -is [bool] }
                'HasExited' { $candidate -is [bool] }
                'ExitCode' { $candidate -is [int] }
                'ExitTime' { $candidate -is [datetime] }
            }
            if ($valid) {
                $value = if ($candidate -is [datetime]) {
                    $candidate.ToUniversalTime().ToString('o', [Globalization.CultureInfo]::InvariantCulture)
                } else { $candidate }
            }
        }
        [pscustomobject][ordered]@{
            Status = $(if ($valid) { 'ok' } else { 'invalid' })
            Value = $value
            ExceptionType = $null
        }
    }
    catch {
        # Do not copy messages, stacks or partial success output into a receipt.
        [pscustomobject][ordered]@{
            Status = 'error'
            Value = $null
            ExceptionType = $_.Exception.GetType().FullName
        }
    }
}

function New-ObservationSnapshot {
    param([System.Collections.IDictionary]$Readers, [string]$Mode)
    $names = @('ProcessId','StartTime','WaitSignaled','HasExited','ExitCode','ExitTime')
    if ($null -eq $Readers -or $Readers.Count -ne $names.Count) {
        throw [ArgumentException]::new('Expected exactly six named readers.')
    }
    foreach ($name in $names) {
        if ($name -cnotin $Readers.Keys -or $Readers[$name] -isnot [scriptblock]) {
            throw [ArgumentException]::new('Expected exactly six named scriptblock readers.')
        }
    }
    $started = [datetime]::UtcNow.ToString('o', [Globalization.CultureInfo]::InvariantCulture)
    $fields = [ordered]@{}
    foreach ($name in $names) {
        $fields[$name] = Read-ObservationField -Reader $Readers[$name] -Name $name
    }
    [pscustomobject][ordered]@{
        SchemaVersion = 1
        Mode = $Mode
        ObservationStartedUtc = $started
        ObservationEndedUtc = [datetime]::UtcNow.ToString('o', [Globalization.CultureInfo]::InvariantCulture)
        Fields = [pscustomobject]$fields
    }
}

function Get-TalenroSyntheticProcessObservation {
    <# Trusted in-memory test readers only. No reader text is evaluated.
       A synthetic receipt is never native process evidence. #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][System.Collections.IDictionary]$Readers)
    New-ObservationSnapshot -Readers $Readers -Mode 'synthetic'
}

function Get-TalenroProcessObservation {
    <# One sequential, non-atomic snapshot of the caller's existing Process.
       Caller owns its identity and lifetime. No reopening, disposal or verdict.
       This live adapter requires separately scoped runtime verification. #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][ValidateNotNull()][System.Diagnostics.Process]$ProcessObject)
    $readers = [ordered]@{
        ProcessId = { $ProcessObject.get_Id() }.GetNewClosure()
        StartTime = { $ProcessObject.get_StartTime() }.GetNewClosure()
        WaitSignaled = { $ProcessObject.WaitForExit(0) }.GetNewClosure()
        HasExited = { $ProcessObject.get_HasExited() }.GetNewClosure()
        ExitCode = { $ProcessObject.get_ExitCode() }.GetNewClosure()
        ExitTime = { $ProcessObject.get_ExitTime() }.GetNewClosure()
    }
    New-ObservationSnapshot -Readers $readers -Mode 'live'
}

Export-ModuleMember -Function Get-TalenroProcessObservation, Get-TalenroSyntheticProcessObservation
