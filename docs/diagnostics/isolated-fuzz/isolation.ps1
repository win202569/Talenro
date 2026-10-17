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
    public static Info Snapshot(SafeFileHandle handle) {
        Info info;
        if (!GetFileInformationByHandle(handle, out info))
            throw new System.ComponentModel.Win32Exception(Marshal.GetLastWin32Error());
        return info;
    }
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
    $receipt=[pscustomobject]@{ Source=$src; Toolchain=$toolchain; Cache=$cache; BeforeManifest=$before; CopyManifest=$copied }
    $receiptPath=Get-IsolatedFullPath (Join-Path $root 'copy-receipt.json')
    $bytes=[Text.Encoding]::UTF8.GetBytes(($receipt|ConvertTo-Json -Depth 6))
    $stream=[IO.File]::Open($receiptPath,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    try{$stream.Write($bytes,0,$bytes.Length)}finally{$stream.Dispose()}
    return $receipt
}
function Install-DiagnosticPatch([string]$Toolchain) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
    $allowed=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $tool=Get-IsolatedFullPath $Toolchain
    if (-not (Test-IsolatedWithin $tool $allowed)) { throw 'Patch target outside dedicated copy scope' }
    $pins=@{ 'src/context/context.go'='765F99F7CC0710B0117FFA183E432524F429C21496FCBAD9B63FB9886FDBAAC7'; 'src/internal/fuzz/fuzz.go'='D9DE12569D5C18FB629C8E79A4D629EB2985B719E4FBFE30C3F3CE23FC44FEB5' }
    foreach($path in $pins.Keys) {
        $file=Get-IsolatedFullPath (Join-Path $tool $path)
        if ([IsolatedFuzzFileInfo]::Links($file) -ne 1 -or (Get-FileHash -LiteralPath $file).Hash -ne $pins[$path]) { throw "Source pin or independent file check failed: $path" }
    }
    $templates=@{ 'context-diag.go.txt'='src/context/diagnostic_fuzz.go'; 'context-diag_test.go.txt'='src/context/diagnostic_fuzz_test.go'; 'fuzz-diag.go.txt'='src/internal/fuzz/diagnostic_fuzz.go'; 'fuzz-diag_test.go.txt'='src/internal/fuzz/diagnostic_fuzz_test.go' }
    foreach($name in $templates.Keys) {
        $destination=Get-IsolatedFullPath (Join-Path $tool $templates[$name])
        if (Test-Path -LiteralPath $destination) { throw 'Diagnostic template target already exists' }
        if (-not (Test-Path -LiteralPath (Join-Path $PSScriptRoot $name) -PathType Leaf)) { throw 'Required portable template missing' }
    }
    $relative=$tool.Substring($workspace.Length+1).Replace('\','/')
    $patch=Join-Path $PSScriptRoot 'coordinator.patch'
    git -c core.fsmonitor=false -C $workspace apply --check --unidiff-zero --ignore-space-change --directory=$relative $patch
    if ($LASTEXITCODE -ne 0) { throw 'Patch precheck failed' }
    git -c core.fsmonitor=false -C $workspace apply --unidiff-zero --ignore-space-change --directory=$relative $patch
    if ($LASTEXITCODE -ne 0) { throw 'Patch application failed; preserve incomplete copy' }
    foreach($name in $templates.Keys) { Copy-Item -LiteralPath (Join-Path $PSScriptRoot $name) -Destination (Join-Path $tool $templates[$name]) -ErrorAction Stop }
    $installed=@{}
    foreach($path in @($pins.Keys)+@($templates.Values)) {$installed[$path]=(Get-FileHash -LiteralPath (Join-Path $tool $path)).Hash}
    [pscustomobject]@{Toolchain=$tool;SourcePins=$pins;InstalledHashes=$installed;PatchHash=(Get-FileHash -LiteralPath $patch).Hash;UTC=[DateTime]::UtcNow.ToString('o')} | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath (Join-Path (Split-Path -Parent $tool) 'patch-receipt.json')
}
