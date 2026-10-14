Set-StrictMode -Version Latest
if (-not ('IsolatedFuzzFileInfo' -as [type])) {
    Add-Type -TypeDefinition @'
using System;
using System.IO;
using System.Runtime.InteropServices;
using Microsoft.Win32.SafeHandles;
public static class IsolatedFuzzFileInfo {
    [StructLayout(LayoutKind.Sequential)]
    public struct Info {
        public uint Attributes, CreationLow, CreationHigh, AccessLow, AccessHigh,
                    WriteLow, WriteHigh, Volume, SizeHigh, SizeLow, Links, IndexHigh, IndexLow;
    }
    [DllImport("kernel32.dll", SetLastError=true)]
    static extern bool GetFileInformationByHandle(SafeFileHandle handle, out Info info);
    public static uint Links(string path) {
        using (var stream = new FileStream(path, FileMode.Open, FileAccess.Read, FileShare.Read)) {
            Info info;
            if (!GetFileInformationByHandle(stream.SafeFileHandle, out info))
                throw new System.ComponentModel.Win32Exception(Marshal.GetLastWin32Error());
            return info.Links;
        }
    }
}
'@
}
function Get-IsolatedFullPath([string]$Path) {
    if ([string]::IsNullOrWhiteSpace($Path) -or -not [IO.Path]::IsPathRooted($Path)) { throw 'Absolute path required' }
    $full=[IO.Path]::GetFullPath($Path).TrimEnd('\','/')
    if ($full.Length -lt 4 -or $full.Substring(2).Contains(':') -or $full.StartsWith('\\')) { throw 'Drive-root, UNC or stream path rejected' }
    $cursor=$full
    while ($cursor -and $cursor.Length -gt 3) {
        if (Test-Path -LiteralPath $cursor) {
            $item=Get-Item -LiteralPath $cursor -Force -ErrorAction Stop
            if (($item.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw "Reparse path rejected: $cursor" }
        }
        $cursor=Split-Path -Parent $cursor
    }
    return $full
}
function Test-IsolatedWithin([string]$Path,[string]$Parent) {
    return $Path.StartsWith($Parent+'\',[StringComparison]::OrdinalIgnoreCase)
}
function Assert-IsolatedPaths([string]$Source,[string]$Destination,[string]$AllowedRoot) {
    $src=Get-IsolatedFullPath $Source
    $dst=Get-IsolatedFullPath $Destination
    $allowed=Get-IsolatedFullPath $AllowedRoot
    if (-not (Test-Path -LiteralPath $src -PathType Container)) { throw 'Source directory missing' }
    if (-not (Test-Path -LiteralPath $allowed -PathType Container)) { throw 'Allowed root missing' }
    if (-not (Test-IsolatedWithin $dst $allowed)) { throw 'Destination outside dedicated root' }
    if ($src.Equals($dst,[StringComparison]::OrdinalIgnoreCase) -or (Test-IsolatedWithin $src $dst) -or (Test-IsolatedWithin $dst $src)) { throw 'Source/destination overlap' }
    if (Test-Path -LiteralPath $dst) { throw 'Destination already exists' }
}
function Get-IsolatedFiles([string]$Root) {
    $resolved=Get-IsolatedFullPath $Root
    if (-not (Test-Path -LiteralPath $resolved -PathType Container)) { throw 'Manifest directory missing' }
    Get-IsolatedChildren $resolved
}
function Get-IsolatedChildren([string]$Root) {
    # Ancestors were validated at the root; descent checks every component once.
    # Never recurse through a reparse entry, even if its target stays in scope.
    $resolved=$Root
    foreach ($entry in Get-ChildItem -LiteralPath $resolved -Force -ErrorAction Stop) {
        if (-not (Test-IsolatedWithin $entry.FullName $resolved) -or ($entry.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw "Reparse or escaped entry rejected: $($entry.FullName)" }
        if ($entry.PSIsContainer) { Get-IsolatedChildren $entry.FullName }
        else {
            if ([IsolatedFuzzFileInfo]::Links($entry.FullName) -ne 1) { throw "Multiple hardlink rejected: $($entry.FullName)" }
            $entry
        }
    }
}
function Get-TreeManifest([string]$Root) {
    $resolved=Get-IsolatedFullPath $Root
    foreach ($entry in @(Get-IsolatedFiles $resolved | Sort-Object FullName)) {
        [pscustomobject]@{ Path=$entry.FullName.Substring($resolved.Length+1).Replace('\','/'); SHA256=(Get-FileHash -LiteralPath $entry.FullName -Algorithm SHA256).Hash; Length=$entry.Length }
    }
}
function New-IsolatedToolchain([string]$Source,[string]$RunRoot) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
    $allowed=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    Assert-IsolatedPaths $Source $RunRoot $allowed
    $src=Get-IsolatedFullPath $Source
    $root=Get-IsolatedFullPath $RunRoot
    $before=@(Get-TreeManifest $src)
    $toolchain=Join-Path $root 'toolchain'
    $cache=Join-Path $root 'cache'
    New-Item -ItemType Directory -Path $toolchain -ErrorAction Stop | Out-Null
    New-Item -ItemType Directory -Path $cache -ErrorAction Stop | Out-Null
    $parents=New-Object 'System.Collections.Generic.HashSet[string]' ([StringComparer]::OrdinalIgnoreCase)
    foreach ($row in $before) {
        $original=Join-Path $src $row.Path
        $target=Join-Path $toolchain $row.Path
        $parent=Split-Path -Parent $original
        if ($parents.Add($parent)) { $null=Get-IsolatedFullPath $parent }
        if (((Get-Item -LiteralPath $original -Force).Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'Source became reparse entry' }
        if ([IsolatedFuzzFileInfo]::Links($original) -ne 1) { throw 'Source link changed' }
        New-Item -ItemType Directory -Path (Split-Path -Parent $target) -Force -ErrorAction Stop | Out-Null
        Copy-Item -LiteralPath $original -Destination $target -ErrorAction Stop
        [IO.File]::SetAttributes($target,([IO.File]::GetAttributes($target) -band (-bnot [IO.FileAttributes]::ReadOnly)))
    }
    $copied=@(Get-TreeManifest $toolchain)
    $after=@(Get-TreeManifest $src)
    $expected=ConvertTo-Json -InputObject $before -Compress
    if ($expected -cne (ConvertTo-Json -InputObject $copied -Compress) -or $expected -cne (ConvertTo-Json -InputObject $after -Compress)) { throw 'Source or copy manifest changed' }
    [pscustomobject]@{ Source=$src; Toolchain=$toolchain; Cache=$cache; BeforeManifest=$before; CopyManifest=$copied }
}
