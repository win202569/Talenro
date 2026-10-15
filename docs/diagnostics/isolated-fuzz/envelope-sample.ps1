. (Join-Path $PSScriptRoot 'envelope-observe.ps1')
function Get-EnvelopeSampleFile([string]$Path) {
    $file=Get-IsolatedFullPath $Path
    if(-not (Test-Path -LiteralPath $file -PathType Leaf) -or [IsolatedFuzzFileInfo]::Links($file) -ne 1){throw 'Missing or linked compiled observation input'}
    return $file
}
function Read-EnvelopeSampleJson([string]$Path) {
    Get-Content -LiteralPath (Get-EnvelopeSampleFile $Path) -Raw -ErrorAction Stop|ConvertFrom-Json -ErrorAction Stop
}
function Assert-EnvelopeSampleInputs([object]$Plan,[object[]]$CurrentProduct) {
    foreach($record in $Plan.ArtifactHashes) {
        $file=Get-EnvelopeSampleFile $record.Path
        if((Get-FileHash -LiteralPath $file -ErrorAction Stop).Hash -cne $record.SHA256){throw 'Compiled observation artifact changed'}
    }
    if((Assert-EnvelopeRunProvenance $Plan.RuntimePlan) -cne $Plan.SourceIdentity){throw 'Copied observation source changed'}
    Assert-EnvelopeProductUnchanged $Plan.ProductManifest $CurrentProduct
}
function Get-EnvelopeCompiledSamplePlan([string]$RunRoot,[string]$CompileRoot,[string]$Toolchain,[string]$ModuleCache) {
    $runtime=Get-EnvelopeObservationPlan $RunRoot $Toolchain $ModuleCache
    $compiled=Get-IsolatedFullPath $CompileRoot
    $scope=Join-Path $runtime.Project '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    if(-not (Test-IsolatedWithin $compiled $scope) -or [IO.Path]::GetFileName($compiled) -cnotmatch '^envelope-observe-[0-9]{3}$' -or $compiled -ieq $runtime.RunRoot){throw 'Distinct owned compiled identity required'}
    $identity=Assert-EnvelopeRunProvenance $runtime
    $paths=@{}
    $hashes=@(foreach($name in @('compile-result.json','binary-receipt.json','build/receipt.json','build/stderr.txt','before-product.json','after-product.json','nodeissuer.test.exe')) {
        $file=Get-EnvelopeSampleFile (Join-Path $compiled $name);$paths[$name]=$file
        [pscustomobject]@{Path=$file;SHA256=(Get-FileHash -LiteralPath $file -ErrorAction Stop).Hash}
    })
    $result=Read-EnvelopeSampleJson $paths['compile-result.json']
    foreach($flag in @('CompileOnly','BuildPassed','BuildExecuted')) {if($result.$flag -isnot [bool] -or -not $result.$flag){throw 'Successful compile-only result required'}}
    foreach($flag in @('SamplingExecuted','Acceptance')) {if($result.$flag -isnot [bool] -or $result.$flag){throw 'Unexecuted unaccepted compile-only result required'}}
    if($result.RunRoot -ine $compiled -or $result.Binary -ine $paths['nodeissuer.test.exe'] -or $result.BinaryReceipt -ine $paths['binary-receipt.json'] -or $result.CompilerEvidence -ine $paths['build/stderr.txt']){throw 'Compile result path binding mismatch'}
    $binding=Read-EnvelopeSampleJson $paths['binary-receipt.json']
    Assert-DiagnosticBinarySource $result $identity;Assert-DiagnosticBinarySource $binding $identity
    $binaryHash=($hashes|Where-Object {$_.Path -ieq $paths['nodeissuer.test.exe']}).SHA256
    if($binding.BinaryHash -cne $binaryHash -or $result.BinaryHash -cne $binaryHash -or $binding.OriginalFuzzHash -cne '318CF8DE04BA5D95E71E625ABF6AAED8DFD54B854A9A925B7A3004D4132D989E' -or $binding.GOROOT -ine $runtime.Toolchain -or $binding.Project -ine $runtime.Project -or $binding.CompilerEvidence -ine $paths['build/stderr.txt'] -or $binding.ProductBefore -ine $paths['before-product.json']){throw 'Compiled binary/source binding mismatch'}
    $receipt=Read-EnvelopeSampleJson $paths['build/receipt.json']
    Assert-ProcessCompletion $receipt
    $buildArguments=@($runtime.BuildArguments|ForEach-Object {if($_ -ceq ('-o='+$runtime.Binary)){'-o='+$paths['nodeissuer.test.exe']}else{$_}})
    if($receipt.ShutdownSchema -ne 3 -or $receipt.ExitCode -ne 0 -or $receipt.Executable -ine $runtime.Executable -or ($receipt.Arguments -join '|') -cne ($buildArguments -join '|')){throw 'Successful bound compile receipt required'}
    foreach($name in @('GOROOT','GOTOOLCHAIN','GOENV','GOPROXY','GOSUMDB','GOAUTH','GOVCS','GO111MODULE','GOWORK','GOFLAGS','GOOS','GOARCH','GOMODCACHE')) {
        if($receipt.Environment.$name -cne $runtime.Environment[$name]){throw 'Compile environment binding mismatch'}
    }
    $log=Get-Content -LiteralPath $paths['build/stderr.txt'] -Raw -ErrorAction Stop
    Assert-CopiedCompileEvidence $log $runtime.Toolchain
    Assert-EnvelopeCompileEvidence $log $runtime.Toolchain $runtime.Project
    $before=@(Read-EnvelopeSampleJson $paths['before-product.json'])
    Assert-EnvelopeProductUnchanged $before @(Read-EnvelopeSampleJson $paths['after-product.json'])
    Assert-EnvelopeProductUnchanged $before @(Get-EnvelopeProductManifest)
    [pscustomobject]@{RunRoot=$runtime.RunRoot;CompileRoot=$compiled;RuntimePlan=$runtime;Binary=$paths['nodeissuer.test.exe'];BinaryHash=$binaryHash;SourceIdentity=$identity;ProductManifest=$before;ArtifactHashes=$hashes}
}
function Invoke-EnvelopeCompiledSample([string]$RunRoot,[string]$CompileRoot,[string]$Toolchain,[string]$ModuleCache) {
    $plan=Get-EnvelopeCompiledSamplePlan $RunRoot $CompileRoot $Toolchain $ModuleCache
    Assert-EnvelopeSampleInputs $plan @(Get-EnvelopeProductManifest)
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    $plan|ConvertTo-Json -Depth 7|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'plan.json') -ErrorAction Stop
    $plan.ProductManifest|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'before-product.json') -ErrorAction Stop
    foreach($name in @('go-cache','fuzz-cache','tmp')){New-Item -ItemType Directory -Path (Join-Path $plan.RunRoot $name) -ErrorAction Stop|Out-Null}
    $pending=$null;$attempted=$false;$stage='execute'
    try {
        $attempted=$true
        $receipt=Invoke-RecordedProcess $plan.Binary $plan.RuntimePlan.ExecuteArguments $plan.RuntimePlan.Environment (Join-Path $plan.RunRoot 'execute') $plan.RuntimePlan.ExecuteLimitSeconds
        $pending=[pscustomobject]@{Scope='Single original Envelope observation from validated compile';CompileRoot=$plan.CompileRoot;SourceIdentity=$plan.SourceIdentity;BinaryHash=$plan.BinaryHash;SamplingExecuted=$true;Acceptance=$false;ObservationValidationPassed=$false;ObservationValidationFailure=$null;Outcome='observed-failure-preserved';Receipt=$receipt;Trace=@()}
        $stage='validate'
        $pending.Trace=@(Get-Content -LiteralPath $receipt.StderrPath -ErrorAction Stop|Where-Object {$_.StartsWith('TALENRO_FUZZ_DIAG ')}|ForEach-Object {$_.Substring(18)|ConvertFrom-Json -ErrorAction Stop})
        Assert-ArmResult 'observe' $receipt $pending.Trace
        if($receipt.TotalProcesses -lt 2){throw 'Actual worker process evidence missing'}
        $pending.ObservationValidationPassed=$true
        $pending.Outcome='observed-success-not-acceptance'
    } catch {
        if($null -ne $pending){$pending.ObservationValidationPassed=$false;$pending.ObservationValidationFailure=$_.Exception.Message}
        [pscustomobject]@{Stage=$stage;Error=$_.Exception.Message;SamplingAttempted=$attempted;SamplingExecuted=($null -ne $pending);Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json') -ErrorAction Stop
        throw
    } finally {
        try {
            $after=@(Get-EnvelopeProductManifest)
            $after|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'after-product.json') -ErrorAction Stop
            Assert-EnvelopeProductUnchanged $plan.ProductManifest $after
            Assert-EnvelopeSampleInputs $plan $after
            if($null -ne $pending) {
                $pending|ConvertTo-Json -Depth 8|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'sample-result.json') -ErrorAction Stop
                if($pending.ObservationValidationPassed){Write-Output $pending}
            }
        } catch {
            $finalError=$_.Exception.Message
            if($null -ne $pending){$pending.ObservationValidationPassed=$false;$pending.ObservationValidationFailure=$finalError;$pending.Outcome='observed-failure-preserved';$pending|ConvertTo-Json -Depth 8|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'sample-result.json') -ErrorAction Stop}
            [pscustomobject]@{Stage='finalize';PriorStage=$stage;Error=$finalError;SamplingAttempted=$attempted;SamplingExecuted=($null -ne $pending);Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json') -ErrorAction Stop
            throw
        }
    }
}
