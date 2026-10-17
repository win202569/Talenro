param([Parameter(Mandatory=$true)][ValidateSet('A','B')][string]$Variant)
$ErrorActionPreference='Stop'
# BEGIN COMBINED TIMING HELPERS
# Throwaway diagnostic markers only; never emit objects into test result pipelines.
function Combined-Event([string]$Kind,[string]$Name,[string]$Event){
    $record=[pscustomobject]@{Schema=1;Event=$Event;Kind=$Kind;Name=$Name;Ticks=[Diagnostics.Stopwatch]::GetTimestamp();Frequency=[Diagnostics.Stopwatch]::Frequency}
    [Console]::Out.WriteLine('CTIME: '+($record|ConvertTo-Json -Compress));[Console]::Out.Flush()
}
# END COMBINED TIMING HELPERS
$identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
Combined-Event Run PrefixRegression start
Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
Combined-Event Stage ReflectionReuse start
Write-Output 'STAGE: ReflectionReuse'
& (Join-Path $PSScriptRoot 'offline-supervision-policy.reflection-reuse.tests.ps1')
if(-not $?){exit 1}
Combined-Event Stage ReflectionReuse end
Combined-Event Stage ExistingAll start
Write-Output 'STAGE: ExistingAll'
& (Join-Path $PSScriptRoot 'offline-supervision-policy.ownership.tests.ps1') -Group all -Variant $Variant
if(-not $?){exit 1}
Combined-Event Stage ExistingAll end
Write-Output 'PASS: PrefixTiming'
Combined-Event Run PrefixRegression end
Write-Output 'LIMIT: paired prefix diagnostic only; excludes boundary and ownership regression, no formal acceptance or retry.'
