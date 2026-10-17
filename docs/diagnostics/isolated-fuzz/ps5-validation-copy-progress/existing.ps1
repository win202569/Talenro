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
Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
Combined-Event Stage ReflectionReuse start
Write-Output 'STAGE: ReflectionReuse'
& (Join-Path $PSScriptRoot 'offline-supervision-policy.reflection-reuse.tests.ps1')
if(-not $?){exit 1}
Combined-Event Stage ReflectionReuse end
Combined-Event Stage ExistingAll start
Write-Output 'STAGE: ExistingAll'
& (Join-Path $PSScriptRoot 'offline-supervision-policy.ownership.tests.ps1') -Group all
if(-not $?){exit 1}
Combined-Event Stage ExistingAll end
Write-Output 'PASS: ReflectionReuseRegression'
