#requires -Version 7.0
# Compile/reflect only. Never invoke Capture or native functions, even with bad handles.
$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$script:count=0
function Check($ok,[string]$name) { if (-not $ok) {throw "ASSERT: $name"}; $script:count++; "PASS: $name" }
$sourcePath=Join-Path $PSScriptRoot 'native-process-observation.cs'
Check (Test-Path $sourcePath) 'helper source exists'
Check ($null -eq ('Talenro.Diagnostics.NativeObservationV1.NativeProbe' -as [type])) 'fresh compilation host'
$types=Add-Type -Path $sourcePath -PassThru
$probe=$types | Where-Object FullName -CEQ 'Talenro.Diagnostics.NativeObservationV1.NativeProbe'
Check ($null -ne $probe -and $null -eq $probe.TypeInitializer) 'helper has no static initializer'
$capture=$probe.GetMethod('Capture')
Check ($capture.GetParameters().Count -eq 1 -and $capture.GetParameters()[0].ParameterType -eq [Microsoft.Win32.SafeHandles.SafeProcessHandle] -and $capture.ReturnType.FullName -ceq 'Talenro.Diagnostics.NativeObservationV1.ProbeSnapshot') 'Capture signature only'
$flags=[Reflection.BindingFlags]'NonPublic,Static'
$imports=@($probe.GetMethods($flags) | Where-Object { $_.GetCustomAttributes([Runtime.InteropServices.DllImportAttribute],$false).Count -gt 0 })
Check ($imports.Count -eq 3) 'three private native imports'
foreach ($method in $imports) {
    $attr=$method.GetCustomAttributes([Runtime.InteropServices.DllImportAttribute],$false)[0]
    Check ($attr.Value -ceq 'kernel32.dll' -and $attr.SetLastError -and $method.GetParameters()[0].ParameterType -eq [IntPtr]) "import flags $($method.Name)"
    if ($method.Name -ceq 'WaitForSingleObject') {
        Check ($method.ReturnType -eq [uint32] -and $method.GetParameters()[1].ParameterType -eq [uint32]) 'wait UInt32 ABI'
    } else {
        $marshal=$method.ReturnParameter.GetCustomAttributes([Runtime.InteropServices.MarshalAsAttribute],$false)
        Check ($method.ReturnType -eq [bool] -and $marshal.Count -eq 1 -and $marshal[0].Value -eq [Runtime.InteropServices.UnmanagedType]::Bool) "BOOL ABI $($method.Name)"
    }
}
$filetime=$types | Where-Object Name -CEQ 'NativeFileTime'
Check ([Runtime.InteropServices.Marshal]::SizeOf([type]$filetime) -eq 8 -and [Runtime.InteropServices.Marshal]::OffsetOf([type]$filetime,'Low').ToInt32() -eq 0 -and [Runtime.InteropServices.Marshal]::OffsetOf([type]$filetime,'High').ToInt32() -eq 4) 'FILETIME size and offsets'
$exit=$probe.GetMethod('GetExitCodeProcess',$flags)
Check ($exit.GetParameters()[1].IsOut -and $exit.GetParameters()[1].ParameterType.GetElementType() -eq [uint32]) 'DWORD output ABI'
$times=$probe.GetMethod('GetProcessTimes',$flags)
Check ($times.GetParameters().Count -eq 5) 'four FILETIME outputs'
foreach ($p in $times.GetParameters() | Select-Object -Skip 1) { Check ($p.IsOut -and $p.ParameterType.GetElementType() -eq $filetime) "FILETIME ABI $($p.Name)" }
$call=$types | Where-Object Name -CEQ 'ProbeCall'
Check ($call.GetProperty('RawResult').PropertyType -eq [Nullable[uint32]] -and $call.GetProperty('ExitCode').PropertyType -eq [Nullable[uint32]] -and $call.GetProperty('ApiSucceeded').PropertyType -eq [Nullable[bool]] -and $call.GetProperty('LastError').PropertyType -eq [Nullable[int]]) 'nullable raw DTO types'
$module=Import-Module (Join-Path $PSScriptRoot 'native-process-observation.psm1') -Force -PassThru
# Invoke only private loader rejection branch; helper already exists and no Capture is called.
$rejected=$false
try { & $module { Get-NativeProbeType } | Out-Null } catch { $rejected=$_.Exception.Message -match 'Preloaded native helper' }
Check $rejected 'untrusted preloaded helper rejected before use'
"COMPILE_REFLECTION_PASS: $script:count assertions; native NOT_RUN"
