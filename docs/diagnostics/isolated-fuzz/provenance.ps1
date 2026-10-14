Set-StrictMode -Version Latest
function Get-DiagnosticSourceIdentity([string]$Toolchain,[object]$PatchRecord) {
    $tool=Get-IsolatedFullPath $Toolchain
    $map=@{'context-diag.go.txt'='src/context/diagnostic_fuzz.go';'context-diag_test.go.txt'='src/context/diagnostic_fuzz_test.go';'fuzz-diag.go.txt'='src/internal/fuzz/diagnostic_fuzz.go';'fuzz-diag_test.go.txt'='src/internal/fuzz/diagnostic_fuzz_test.go'}
    $expected=@($map.Values)+@('src/context/context.go','src/internal/fuzz/fuzz.go')
    $properties=@($PatchRecord.InstalledHashes.PSObject.Properties)
    $patchHash=(Get-FileHash -LiteralPath (Join-Path $PSScriptRoot 'coordinator.patch')).Hash
    if($PatchRecord.Toolchain -ne $tool -or $PatchRecord.PatchHash -ne $patchHash -or $properties.Count -ne 6){throw 'Diagnostic source receipt incomplete or drifted'}
    foreach($property in $properties){if($property.Name -notin $expected){throw 'Unexpected diagnostic source entry'}}
    $hashes=@{}
    foreach($path in $expected) {
        $file=Get-IsolatedFullPath (Join-Path $tool $path)
        if([IsolatedFuzzFileInfo]::Links($file) -ne 1){throw 'Installed diagnostic source is linked'}
        $hashes[$path]=(Get-FileHash -LiteralPath $file).Hash
        if($hashes[$path] -ne $PatchRecord.InstalledHashes.$path){throw 'Installed diagnostic source drift'}
    }
    foreach($name in $map.Keys) {
        $portable=Get-IsolatedFullPath (Join-Path $PSScriptRoot $name)
        if([IsolatedFuzzFileInfo]::Links($portable) -ne 1 -or (Get-FileHash -LiteralPath $portable).Hash -ne $hashes[$map[$name]]){throw 'Portable template differs from installed helper'}
    }
    # Preserve the recorded identity's order; runtime/culture collation is not a source binding.
    $canonicalPaths=@('src/context/context.go','src/context/diagnostic_fuzz_test.go','src/context/diagnostic_fuzz.go','src/internal/fuzz/diagnostic_fuzz_test.go','src/internal/fuzz/diagnostic_fuzz.go','src/internal/fuzz/fuzz.go')
    $parts=@('patch='+$patchHash)+@(foreach($path in $canonicalPaths){ $path+'='+$hashes[$path] })
    $sha=[Security.Cryptography.SHA256]::Create()
    try{return [BitConverter]::ToString($sha.ComputeHash([Text.Encoding]::UTF8.GetBytes(($parts -join "`n")))).Replace('-','')}finally{$sha.Dispose()}
}
function Assert-DiagnosticBinarySource([object]$Record,[string]$Identity) {
    $property=$Record.PSObject.Properties['SourceIdentity']
    if($null -eq $property -or [string]::IsNullOrWhiteSpace($Identity) -or $property.Value -cne $Identity){throw 'Binary source identity missing or mismatched; no automatic rebuild'}
}
