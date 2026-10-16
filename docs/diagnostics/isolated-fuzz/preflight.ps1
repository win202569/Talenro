Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot 'isolation.ps1')
function Test-DiagnosticPrerequisites([string]$Source) {
    if([Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT){throw 'Windows required; no automatic installation'}
    $root=Get-IsolatedFullPath $Source
    $pins=@{
        'bin/go.exe'='D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F'
        'src/context/context.go'='765F99F7CC0710B0117FFA183E432524F429C21496FCBAD9B63FB9886FDBAAC7'
        'src/internal/fuzz/fuzz.go'='D9DE12569D5C18FB629C8E79A4D629EB2985B719E4FBFE30C3F3CE23FC44FEB5'
    }
    foreach($path in @($pins.Keys)+@('VERSION')) {
        $file=Get-IsolatedFullPath (Join-Path $root $path)
        if(-not (Test-Path -LiteralPath $file -PathType Leaf) -or [IsolatedFuzzFileInfo]::Links($file) -ne 1){throw 'Missing or linked prerequisite file'}
        if($pins.ContainsKey($path) -and (Get-FileHash -LiteralPath $file).Hash -ne $pins[$path]){throw 'Pinned prerequisite drift'}
    }
    if(@(Get-Content -LiteralPath (Join-Path $root 'VERSION'))[0] -cne 'go1.26.5'){throw 'Static version mismatch'}
    foreach($name in @('isolation.ps1','run.ps1','provenance.ps1','process-native.cs','retained-lifetime.cs','compile-stage-trigger.cs','compile-stage-session.cs','coordinator.patch','context-diag.go.txt','context-diag_test.go.txt','fuzz-diag.go.txt','fuzz-diag_test.go.txt','fixture_test.go.txt')) {
        if(-not (Test-Path -LiteralPath (Join-Path $PSScriptRoot $name) -PathType Leaf)){throw 'Portable diagnostic material missing'}
    }
    [pscustomobject]@{Ready=$true;Source=$root;StaticVersion='go1.26.5';GoHash=$pins['bin/go.exe'];Scope='Prerequisites only; full copy manifest and actual version/origin checks still required'}
}
