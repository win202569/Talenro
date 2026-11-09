#requires -Version 7.0
$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$script:count=0
function Check($condition, $label) {
    if (-not $condition) { throw "ASSERT: $label" }
    $script:count++; "PASS: $label"
}
$modulePath=Join-Path $PSScriptRoot 'observation-session-guard.psm1'
if (-not (Test-Path $modulePath)) { throw 'ASSERT: observation session guard missing' }
Import-Module $modulePath -Force
$start=[datetime]::Parse('2026-11-09T18:39:47.4427247Z').ToUniversalTime()
$cim=$start.AddTicks(-7)
function IdentityArgs {
    @{
        ProcessId=23364; CimProcessId=23364
        ExpectedPath='D:\fixture\demo.exe'; CimPath='D:\fixture\demo.exe'
        ExpectedCommandLine='"D:\fixture\demo.exe" --layout-test --layout-profile=unknown-usage'
        CimCommandLine='"D:\fixture\demo.exe" --layout-test --layout-profile=unknown-usage'
        ProcessStartUtc=$start; CimCreationUtc=$cim; NativeCreationFileTime=$start.ToFileTimeUtc()
    }
}
function Reject($argsMap, $label) {
    $rejected=$false
    try { Assert-TalenroObservationIdentity @argsMap } catch { $rejected=$true }
    Check $rejected $label
}
Check ($start.ToString('o') -cne $cim.ToString('o')) 'historical exact-string mismatch reproduced'
foreach ($tail in 0..9) {
    $a=IdentityArgs; $a.ProcessStartUtc=$cim.AddTicks($tail)
    $a.NativeCreationFileTime=$a.ProcessStartUtc.ToFileTimeUtc()
    Assert-TalenroObservationIdentity @a
    Check $true "same CIM microsecond with exact native identity: $tail"
}
foreach ($delta in @(-10,10)) {
    $a=IdentityArgs; $a.CimCreationUtc=$cim.AddTicks($delta)
    Reject $a "adjacent CIM microsecond $delta rejected"
}
$a=IdentityArgs; $a.CimCreationUtc=$start; Reject $a 'non-microsecond CIM evidence rejected'
$a=IdentityArgs; $a.NativeCreationFileTime++; Reject $a 'native one-tick mismatch rejected'
foreach ($entry in @(
    @('CimProcessId',23365), @('ProcessId',0), @('CimPath','D:\other\demo.exe'),
    @('CimCommandLine','"D:\fixture\demo.exe" --layout-test --layout-profile=unknown-usage-extra'),
    @('CimCommandLine',''), @('CimCreationUtc',$null),
    @('ProcessStartUtc',[datetime]::SpecifyKind($start,[DateTimeKind]::Unspecified)),
    @('NativeCreationFileTime','123'), @('CimProcessId','23364')
)) {
    $a=IdentityArgs; $a[$entry[0]]=$entry[1]; Reject $a "invalid identity $($entry[0])"
}

# MemoryStream provides real IDisposable behavior without processes or disk files.
foreach ($phase in @('success','identity','sampling','acquire-process')) {
    $writer=[IO.MemoryStream]::new(); $processResource=[IO.MemoryStream]::new()
    $factory={ if($phase -eq 'acquire-process'){throw 'acquire-marker'}; return ,$processResource }.GetNewClosure()
    $body={ param($p,$w)
        if($phase -eq 'identity'){ $a=IdentityArgs; $a.CimProcessId=1; Assert-TalenroObservationIdentity @a }
        if($phase -eq 'sampling'){throw 'sampling-marker'}
        $w.WriteByte(42); 'observed'
    }.GetNewClosure()
    $failed=$false; $result=$null
    try { $result=Invoke-TalenroObservationScope -AcquireWriter {return ,$writer} -AcquireProcess $factory -Observe $body }
    catch { $failed=$true }
    Check ($failed -eq ($phase -ne 'success')) "failure preserved: $phase"
    Check (-not $writer.CanWrite) "writer disposed: $phase"
    Check ($processResource.CanWrite -eq ($phase -eq 'acquire-process')) "only acquired process resource disposed: $phase"
    if($phase -eq 'success'){Check ($result -ceq 'observed') 'success result preserved'}
    $processResource.Dispose()
}
$called=[Collections.Generic.List[string]]::new()
$failure=$null
try {
    Invoke-TalenroObservationScope -AcquireWriter {throw 'writer-marker'} -AcquireProcess {$called.Add('process')} -Observe {$called.Add('observe')}
} catch {$failure=$_}
Check ($null -ne $failure -and $called.Count -eq 0) 'writer acquisition failure prevents later callbacks'

$writer=[pscustomobject]@{Calls=0}
$writer | Add-Member ScriptMethod Dispose { $this.Calls++; throw 'dispose-marker' }
$processResource=[IO.MemoryStream]::new()
$failure=$null
try {
    Invoke-TalenroObservationScope -AcquireWriter {return $writer} -AcquireProcess {return ,$processResource} -Observe {throw 'identity-marker'}
} catch {$failure=$_}
Check ($writer.Calls -eq 1 -and -not $processResource.CanWrite) 'cleanup continues after dispose failure'
Check ($failure.Exception -is [AggregateException] -and $failure.Exception.InnerExceptions.Count -eq 2) 'primary and cleanup errors both retained'
Check ($failure.Exception.InnerExceptions[0].Message -match 'identity-marker') 'primary error stays first'
foreach ($fault in @('writer','process','both')) {
    $writer=[pscustomobject]@{Calls=0;Fail=($fault -ne 'process')}
    $processResource=[pscustomobject]@{Calls=0;Fail=($fault -ne 'writer')}
    foreach($resource in @($writer,$processResource)) {
        $resource | Add-Member ScriptMethod Dispose { $this.Calls++; if($this.Fail){throw 'cleanup-marker'} }
    }
    $output=[Collections.Generic.List[object]]::new(); $failure=$null
    try {
        Invoke-TalenroObservationScope -AcquireWriter {return $writer} -AcquireProcess {return $processResource} -Observe {'must-not-publish-success'} |
            ForEach-Object {$output.Add($_)}
    } catch {$failure=$_}
    Check ($writer.Calls -eq 1 -and $processResource.Calls -eq 1) "both disposals attempted after success: $fault"
    Check ($null -ne $failure -and $output.Count -eq 0) "cleanup failure prevents success publication: $fault"
    if($fault -eq 'both') {
        Check ($failure.Exception -is [AggregateException] -and $failure.Exception.InnerExceptions.Count -eq 2) 'both cleanup failures retained'
    }
}
"SYNTHETIC_ONLY_PASS: $script:count assertions; no process launched"
