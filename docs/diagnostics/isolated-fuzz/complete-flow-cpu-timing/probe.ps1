$ErrorActionPreference='Stop'
# BEGIN COMBINED TIMING HELPERS
# Diagnostic boundary samples only; no pipeline objects or product instrumentation.
function Combined-Event([string]$Kind,[string]$Name,[string]$Event){
    $stamp=[Diagnostics.Stopwatch]::GetTimestamp()
    $cpuProcess=[Diagnostics.Process]::GetCurrentProcess()
    try{$cpuTicks=$cpuProcess.TotalProcessorTime.Ticks}finally{$cpuProcess.Dispose()}
    $record=[pscustomobject]@{Schema=2;Event=$Event;Kind=$Kind;Name=$Name;Ticks=$stamp;Frequency=[Diagnostics.Stopwatch]::Frequency;CPUTicks=$cpuTicks}
    [Console]::Out.WriteLine('CTIME: '+($record|ConvertTo-Json -Compress));[Console]::Out.Flush()
}
# END COMBINED TIMING HELPERS
$identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
Combined-Event Run CompleteFlow start
Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
Combined-Event Stage ShapeLookup start
Write-Output 'STAGE: ShapeLookup'
& (Join-Path $PSScriptRoot 'offline-supervision-policy.shape-lookup.tests.ps1')
if(-not $?){exit 1}
Combined-Event Stage ShapeLookup end
Combined-Event Stage ExistingFull start
Write-Output 'STAGE: ExistingFull'
& (Join-Path $PSScriptRoot 'existing.ps1')
if(-not $?){exit 1}
Combined-Event Stage ExistingFull end
Write-Output 'PASS: ShapeLookupRegression'
Combined-Event Run CompleteFlow end
Write-Output 'LIMIT: complete-flow CPU timing diagnostic only; no formal acceptance, retry or budget renewal.'
