param([switch]$Probe,[string]$TracePath)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop';$script:descriptorCase='Context';$savedError=$null
. (Join-Path $PSScriptRoot 'descriptor-fixture.ps1')
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    $admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    [Console]::Out.WriteLine('CONTEXT: '+$identity.Name+' admin='+$admin)
    [Console]::Out.WriteLine('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    if($admin -or $PSVersionTable.PSVersion.ToString() -cne '7.6.5' -or $PSVersionTable.PSEdition -cne 'Core' -or -not [Environment]::Is64BitProcess){throw 'Ordinary pinned host required'}
    . (Join-Path $PSScriptRoot 'control-frame.ps1')
    Invoke-P2BDescriptorRegression -Probe $Probe.IsPresent -Trace $true
    [Console]::Out.WriteLine('LIMIT: descriptor passivity only; no runtime approval')
    exit 0
}catch{$savedError=$_;[Console]::Out.WriteLine('ASSERTION FAILED: P2BDescriptor'+$script:descriptorCase);exit 1}
finally{DG-SaveEvidence $TracePath $savedError}
