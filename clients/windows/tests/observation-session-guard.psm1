#requires -Version 7.0

function Assert-TalenroObservationIdentity {
    <# Pure evidence check. Caller must obtain NativeCreationFileTime from the
       original retained handle with a successful native Times call. This function
       neither opens a PID nor proves that caller-supplied evidence is authentic. #>
    [CmdletBinding()]
    param(
        $ProcessId, $CimProcessId, $ExpectedPath, $CimPath,
        $ExpectedCommandLine, $CimCommandLine,
        $ProcessStartUtc, $CimCreationUtc, $NativeCreationFileTime
    )
    $ErrorActionPreference='Stop'
    foreach ($id in @($ProcessId,$CimProcessId)) {
        if (($id -isnot [int] -and $id -isnot [uint32]) -or $id -le 0) {
            throw [ArgumentException]::new('Invalid process identity.')
        }
    }
    if ($ProcessId -ne $CimProcessId) { throw [InvalidOperationException]::new('Process ID mismatch.') }
    foreach ($value in @($ExpectedPath,$CimPath,$ExpectedCommandLine,$CimCommandLine)) {
        if ($value -isnot [string] -or [string]::IsNullOrWhiteSpace($value)) {
            throw [ArgumentException]::new('Missing path or command-line evidence.')
        }
    }
    if (-not [string]::Equals($ExpectedPath,$CimPath,[StringComparison]::OrdinalIgnoreCase) -or
        -not [string]::Equals($ExpectedCommandLine,$CimCommandLine,[StringComparison]::Ordinal)) {
        throw [InvalidOperationException]::new('Path or command-line mismatch.')
    }
    foreach ($time in @($ProcessStartUtc,$CimCreationUtc)) {
        if ($time -isnot [datetime] -or $time.Kind -ne [DateTimeKind]::Utc) {
            throw [ArgumentException]::new('Explicit UTC creation times required.')
        }
    }
    if ($NativeCreationFileTime -isnot [long] -or $NativeCreationFileTime -le 0 -or
        $NativeCreationFileTime -ne $ProcessStartUtc.ToFileTimeUtc()) {
        throw [InvalidOperationException]::new('Exact native creation identity mismatch.')
    }
    # WMI date-time uses six fractional digits. Integer remainder avoids float
    # rounding and admits only truncation within this microsecond, not a tolerance.
    if ($CimCreationUtc.Ticks % 10 -ne 0 -or
        $CimCreationUtc.Ticks -ne ($ProcessStartUtc.Ticks - ($ProcessStartUtc.Ticks % 10))) {
        throw [InvalidOperationException]::new('CIM creation microsecond mismatch.')
    }
}

function Invoke-TalenroObservationScope {
    <# Trusted local callbacks only, never source strings. Factories transfer one
       owned disposable resource on success; partial factory failure is the
       factory's cleanup responsibility. Observe includes handle acquisition,
       identity checks and sampling. Dispose is NOT process termination.
       This is resource protection, not a launch policy or observation timeout. #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][scriptblock]$AcquireWriter,
        [Parameter(Mandatory)][scriptblock]$AcquireProcess,
        [Parameter(Mandatory)][scriptblock]$Observe
    )
    $ErrorActionPreference='Stop'
    $writer=$null; $processResource=$null; $result=$null
    $errors=[Collections.Generic.List[Exception]]::new()
    try {
        $writer=& $AcquireWriter
        if ($null -eq $writer -or $writer -is [array] -or $null -eq $writer.PSObject.Methods['Dispose']) {
            throw [InvalidOperationException]::new('Writer factory must return one disposable resource.')
        }
        $processResource=& $AcquireProcess
        if ($null -eq $processResource -or $processResource -is [array] -or $null -eq $processResource.PSObject.Methods['Dispose']) {
            throw [InvalidOperationException]::new('Process factory must return one disposable resource.')
        }
        $result=@(& $Observe $processResource $writer)
    } catch { $errors.Add($_.Exception) }
    finally {
        foreach ($resource in @($writer,$processResource)) {
            if ($null -ne $resource) {
                try { $resource.Dispose() } catch { $errors.Add($_.Exception) }
            }
        }
    }
    if ($errors.Count -eq 1) { throw $errors[0] }
    if ($errors.Count -gt 1) { throw [AggregateException]::new('Observation and/or resource cleanup failed.',$errors.ToArray()) }
    $result
}

Export-ModuleMember -Function Assert-TalenroObservationIdentity, Invoke-TalenroObservationScope
