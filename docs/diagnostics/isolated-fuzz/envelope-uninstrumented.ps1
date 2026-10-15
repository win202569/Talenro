. (Join-Path $PSScriptRoot 'envelope-sample.ps1')
function Assert-UninstrumentedEnvelopeManifest([object[]]$Expected,[object[]]$Current) {
    if($Expected.Count -ne 11536 -or $Current.Count -ne 11536){throw 'Complete original11536-file toolchain required'}
    $expectedMap=New-Object 'System.Collections.Generic.Dictionary[string,object]' ([StringComparer]::OrdinalIgnoreCase)
    $currentMap=New-Object 'System.Collections.Generic.Dictionary[string,object]' ([StringComparer]::OrdinalIgnoreCase)
    foreach($set in @([pscustomobject]@{Rows=$Expected;Map=$expectedMap},[pscustomobject]@{Rows=$Current;Map=$currentMap})) {
        foreach($row in $set.Rows) {
            if([string]::IsNullOrWhiteSpace($row.Path) -or [IO.Path]::IsPathRooted($row.Path) -or $row.Path -match '(^|/)\.\.?(/|$)|\\|:' -or $row.Path -match '^src/(context|internal/fuzz)/diagnostic_fuzz(_test)?\.go$' -or $row.SHA256 -cnotmatch '^[0-9A-F]{64}$' -or [long]$row.Length -lt 0 -or $set.Map.ContainsKey($row.Path)){throw 'Invalid duplicate or instrumented toolchain manifest'}
            $set.Map.Add($row.Path,$row)
        }
    }
    $pins=@{'bin/go.exe'='D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F';'src/context/context.go'='765F99F7CC0710B0117FFA183E432524F429C21496FCBAD9B63FB9886FDBAAC7';'src/internal/fuzz/fuzz.go'='D9DE12569D5C18FB629C8E79A4D629EB2985B719E4FBFE30C3F3CE23FC44FEB5'}
    foreach($path in $pins.Keys){if(-not $expectedMap.ContainsKey($path) -or $expectedMap[$path].SHA256 -cne $pins[$path]){throw 'Original toolchain pin mismatch'}}
    foreach($path in $currentMap.Keys){if(-not $expectedMap.ContainsKey($path) -or $currentMap[$path].Path -cne $expectedMap[$path].Path -or $currentMap[$path].SHA256 -cne $expectedMap[$path].SHA256 -or [long]$currentMap[$path].Length -ne [long]$expectedMap[$path].Length){throw 'Uninstrumented toolchain manifest drift'}}
}
function Assert-UninstrumentedEnvelopeToolchain([string]$Toolchain,[string]$ReferenceReceipt) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $tool=Get-IsolatedFullPath $Toolchain;$parent=Split-Path -Parent $tool
    $referencePath=Get-IsolatedFullPath $ReferenceReceipt
    if(-not (Test-IsolatedWithin $tool $scope) -or [IO.Path]::GetFileName($tool) -cne 'toolchain' -or [IO.Path]::GetFileName($parent) -cnotmatch '^uninstrumented-[0-9]{3}$' -or -not (Test-IsolatedWithin $referencePath $scope)){throw 'Owned uninstrumented copy and reference required'}
    $reference=Read-EnvelopeSampleJson $referencePath
    $copyPath=Get-EnvelopeSampleFile (Join-Path $parent 'copy-receipt.json')
    $copy=Read-EnvelopeSampleJson $copyPath
    if($copy.Toolchain -ine $tool -or $copy.Source -ine $reference.Source -or $copy.Source -ieq $tool){throw 'Original source/copy binding mismatch'}
    Assert-UninstrumentedEnvelopeManifest @($reference.BeforeManifest) @($copy.BeforeManifest)
    Assert-UninstrumentedEnvelopeManifest @($reference.BeforeManifest) @($copy.CopyManifest)
    Assert-UninstrumentedEnvelopeManifest @($reference.BeforeManifest) @(Get-TreeManifest $tool)
    return ((Get-FileHash -LiteralPath $referencePath).Hash+':'+(Get-FileHash -LiteralPath $copyPath).Hash)
}
function Invoke-UninstrumentedEnvelope([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache) {
    $plan=Get-EnvelopeObservationPlan $RunRoot $Toolchain $ModuleCache
    $plan.Mode='uninstrumented'
    $plan.Environment.TALENRO_FUZZ_DIAG_MODE='';$plan.Environment.TALENRO_FUZZ_FIXTURE='';$plan.Environment.GODEBUG=''
    $identity=Assert-UninstrumentedEnvelopeToolchain $plan.Toolchain $ReferenceReceipt
    $before=@(Get-EnvelopeProductManifest)
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    $plan|ConvertTo-Json -Depth 5|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'plan.json') -ErrorAction Stop
    $before|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'before-product.json') -ErrorAction Stop
    foreach($name in @('go-cache','fuzz-cache','tmp')){New-Item -ItemType Directory -Path (Join-Path $plan.RunRoot $name) -ErrorAction Stop|Out-Null}
    $pending=$null;$binaryHash=$null;$attempted=$false;$stage='version'
    try {
        $version=Invoke-RecordedProcess $plan.Executable @('version') $plan.Environment (Join-Path $plan.RunRoot 'version') 60
        Assert-ProcessCompletion $version
        if($version.ExitCode -ne 0 -or (Get-Content -LiteralPath $version.StdoutPath -Raw).Trim() -cne 'go version go1.26.5 windows/amd64'){throw 'Uninstrumented version mismatch'}
        $stage='origin'
        $origin=Invoke-RecordedProcess $plan.Executable @('env','GOROOT','GOTOOLDIR','GOMOD') $plan.Environment (Join-Path $plan.RunRoot 'origin') 60
        Assert-ProcessCompletion $origin
        $lines=@(Get-Content -LiteralPath $origin.StdoutPath)
        if($origin.ExitCode -ne 0 -or $lines.Count -ne 3 -or $lines[0] -cne $plan.Toolchain -or $lines[1] -cne (Join-Path $plan.Toolchain 'pkg/tool/windows_amd64') -or $lines[2] -ine (Join-Path $plan.Project 'go.mod')){throw 'Uninstrumented origin mismatch'}
        $stage='build'
        $build=Invoke-RecordedProcess $plan.Executable $plan.BuildArguments $plan.Environment (Join-Path $plan.RunRoot 'build') $plan.BuildLimitSeconds
        Assert-ProcessCompletion $build
        if($build.ExitCode -ne 0){throw 'Single uninstrumented build failed'}
        $log=Get-Content -LiteralPath $build.StderrPath -Raw
        Assert-CopiedCompileEvidence $log $plan.Toolchain;Assert-EnvelopeCompileEvidence $log $plan.Toolchain $plan.Project
        Assert-EnvelopeProductUnchanged $before @(Get-EnvelopeProductManifest)
        if((Assert-UninstrumentedEnvelopeToolchain $plan.Toolchain $ReferenceReceipt) -cne $identity){throw 'Uninstrumented source binding changed during build'}
        $binaryHash=(Get-FileHash -LiteralPath (Get-EnvelopeSampleFile $plan.Binary)).Hash
        [pscustomobject]@{Uninstrumented=$true;ToolchainReference=$identity;BinaryHash=$binaryHash;CompilerEvidence=$build.StderrPath;Project=$plan.Project}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'binary-receipt.json') -ErrorAction Stop
        $stage='execute';$attempted=$true
        $receipt=Invoke-RecordedProcess $plan.Binary $plan.ExecuteArguments $plan.Environment (Join-Path $plan.RunRoot 'execute') $plan.ExecuteLimitSeconds
        $pending=[pscustomobject]@{Scope='One uninstrumented original Envelope short validation';Uninstrumented=$true;ToolchainReference=$identity;BinaryHash=$binaryHash;SamplingExecuted=$true;Acceptance=$false;ObservationValidationPassed=$false;ObservationValidationFailure=$null;RecordedExecutions=$null;Receipt=$receipt}
        $stage='validate'
        Assert-ProcessCompletion $receipt
        if($receipt.ExitCode -ne 0 -or $receipt.TotalProcesses -lt 2){throw 'Uninstrumented observation failed or worker evidence absent'}
        $stdout=@(Get-Content -LiteralPath $receipt.StdoutPath|Where-Object {$_ -ne ''})
        $stderr=Get-Content -LiteralPath $receipt.StderrPath -Raw
        $executionMatches=[regex]::Matches($stderr,'(?m)^fuzz:.*execs: ([0-9]+) ')
        $execs=[UInt64]0
        if($stdout.Count -eq 0 -or $stdout[-1] -cne 'PASS' -or $stderr -match '(?m)^TALENRO_FUZZ_DIAG ' -or $stderr -notmatch '(?m)^fuzz:.*now fuzzing with 1 workers\r?$' -or $executionMatches.Count -eq 0 -or -not [UInt64]::TryParse($executionMatches[$executionMatches.Count-1].Groups[1].Value,[ref]$execs) -or $execs -eq 0){throw 'Uninstrumented output/worker/execution proof failed'}
        $pending.RecordedExecutions=$execs;$pending.ObservationValidationPassed=$true
    } catch {
        if($null -ne $pending){$pending.ObservationValidationFailure=$_.Exception.Message}
        [pscustomobject]@{Stage=$stage;Error=$_.Exception.Message;Uninstrumented=$true;SamplingAttempted=$attempted;SamplingExecuted=($null -ne $pending);Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json') -ErrorAction Stop
        throw
    } finally {
        try {
            $after=@(Get-EnvelopeProductManifest);$after|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'after-product.json') -ErrorAction Stop
            Assert-EnvelopeProductUnchanged $before $after
            if((Assert-UninstrumentedEnvelopeToolchain $plan.Toolchain $ReferenceReceipt) -cne $identity){throw 'Uninstrumented source binding changed during finalization'}
            if($null -ne $binaryHash -and (Get-FileHash -LiteralPath (Get-EnvelopeSampleFile $plan.Binary)).Hash -cne $binaryHash){throw 'Uninstrumented binary changed'}
            if($null -ne $pending){$pending|ConvertTo-Json -Depth 8|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'sample-result.json') -ErrorAction Stop;if($pending.ObservationValidationPassed){Write-Output $pending}}
        } catch {
            $errorText=$_.Exception.Message
            if($null -ne $pending){$pending.ObservationValidationPassed=$false;$pending.ObservationValidationFailure=$errorText;$pending|ConvertTo-Json -Depth 8|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'sample-result.json') -ErrorAction Stop}
            [pscustomobject]@{Stage='finalize';PriorStage=$stage;Error=$errorText;Uninstrumented=$true;SamplingAttempted=$attempted;SamplingExecuted=($null -ne $pending);Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json') -ErrorAction Stop
            throw
        }
    }
}
