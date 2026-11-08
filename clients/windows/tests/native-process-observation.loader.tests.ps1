#requires -Version 7.0
# Fresh host, compilation/loading only. No handles, Capture or native calls.
$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$script:count=0
function Check($ok,[string]$name) {
    if (-not $ok) { throw "ASSERT: $name" }
    $script:count++; "PASS: $name"
}
$path=Join-Path $PSScriptRoot 'native-process-observation.psm1'
$typeName='Talenro.Diagnostics.NativeObservationV1.NativeProbe'
Check ($null -eq ($typeName -as [type])) 'fresh host required'
$module=Import-Module $path -Force -PassThru
Check ($null -eq ($typeName -as [type])) 'import does not compile helper'
$first=@(& $module { Get-NativeProbeType })
Check ($first.Count -eq 1 -and $first[0] -is [type] -and $first[0].FullName -ceq $typeName) 'first load returns only helper type'
Check ($null -eq $first[0].TypeInitializer) 'loaded helper has no static initializer'
$second=@(& $module { Get-NativeProbeType })
Check ($second.Count -eq 1 -and [object]::ReferenceEquals($first[0],$second[0])) 'same module reuses exact cached type'
$cache=& $module { Get-Variable NativeProbeType -Scope Script -ValueOnly }
Check ([object]::ReferenceEquals($first[0],$cache)) 'cache belongs to module instance'
$reloaded=Import-Module $path -Force -PassThru
Check ($null -eq (& $reloaded { Get-Variable NativeProbeType -Scope Script -ErrorAction SilentlyContinue })) 'forced reimport has no trusted cache'
foreach ($attempt in 1..2) {
    $rejected=$false
    try { & $reloaded { Get-NativeProbeType } | Out-Null }
    catch { $rejected=$_.Exception.Message -match 'Preloaded native helper' }
    Check $rejected "forced reimport rejects existing type attempt $attempt"
}
Check ([object]::ReferenceEquals($first[0],($typeName -as [type]))) 'rejection does not replace loaded type'
Check ($null -eq (& $reloaded { Get-Variable NativeProbeType -Scope Script -ErrorAction SilentlyContinue })) 'rejection does not adopt foreign cache'
"LOADER_ONLY_PASS: $script:count assertions; Capture/native NOT_RUN"
