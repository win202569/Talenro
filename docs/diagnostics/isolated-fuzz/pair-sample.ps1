. (Join-Path $PSScriptRoot 'envelope-uninstrumented.ps1')
function Get-PairCompiledPlan([string]$RunRoot,[string]$CompileRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache){
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot;$compiled=Get-IsolatedFullPath $CompileRoot;$tool=Get-IsolatedFullPath $Toolchain;$modules=Get-IsolatedFullPath $ModuleCache
    if(-not (Test-IsolatedWithin $root $scope) -or [IO.Path]::GetFileName($root) -cnotmatch '^pair-observe-[0-9]{3}$' -or (Test-Path -LiteralPath $root)){throw 'Fresh owned Pair identity required'}
    if(-not (Test-IsolatedWithin $compiled $scope) -or [IO.Path]::GetFileName($compiled) -cnotmatch '^envelope-observe-[0-9]{3}$' -or (Test-Path -LiteralPath (Join-Path $compiled 'stopped.json'))){throw 'Successful owned uninstrumented compile required'}
    if(-not (Test-Path -LiteralPath (Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod') -PathType Leaf)){throw 'Existing offline dependencies required'}
    if((Get-FileHash -LiteralPath (Join-Path $workspace 'internal/nodecontrol/nodeissuer/fuzz_test.go')).Hash -cne '318CF8DE04BA5D95E71E625ABF6AAED8DFD54B854A9A925B7A3004D4132D989E'){throw 'Original Pair/Envelope callback drift'}
    $identity=Assert-UninstrumentedEnvelopeToolchain $tool $ReferenceReceipt
    $paths=@{};$hashes=@(foreach($name in @('plan.json','sample-result.json','binary-receipt.json','build/receipt.json','build/stderr.txt','before-product.json','after-product.json','nodeissuer.test.exe')){
        $file=Get-EnvelopeSampleFile (Join-Path $compiled $name);$paths[$name]=$file;[pscustomobject]@{Path=$file;SHA256=(Get-FileHash -LiteralPath $file).Hash}
    })
    $prior=Read-EnvelopeSampleJson $paths['sample-result.json'];$binding=Read-EnvelopeSampleJson $paths['binary-receipt.json'];$old=Read-EnvelopeSampleJson $paths['plan.json']
    foreach($field in @('Uninstrumented','SamplingExecuted','ObservationValidationPassed')){if($prior.$field -isnot [bool] -or -not $prior.$field){throw 'Prior uninstrumented validation not successful'}}
    if($prior.Acceptance -isnot [bool] -or $prior.Acceptance -or $binding.Uninstrumented -isnot [bool] -or -not $binding.Uninstrumented){throw 'Prior acceptance/source label invalid'}
    $binary=$paths['nodeissuer.test.exe'];$binaryHash=(Get-FileHash -LiteralPath $binary).Hash
    if($prior.ToolchainReference -cne $identity -or $binding.ToolchainReference -cne $identity -or $prior.BinaryHash -cne $binaryHash -or $binding.BinaryHash -cne $binaryHash -or $binding.Project -ine $workspace -or $binding.CompilerEvidence -ine $paths['build/stderr.txt']){throw 'Prior binary/source binding mismatch'}
    if($old.RunRoot -ine $compiled -or $old.Binary -ine $binary -or $old.Toolchain -ine $tool -or $old.Project -ine $workspace -or $old.Mode -cne 'uninstrumented'){throw 'Prior compile plan mismatch'}
    $build=Read-EnvelopeSampleJson $paths['build/receipt.json'];Assert-ProcessCompletion $build
    $expectedBuild=@('test','-c','-x','-p=1','-count=1','-run=^$','-fuzz=^FuzzIssuerEnvelope$','-fuzztime=5s','-parallel=1','-timeout=30s',('-o='+$binary),'talenro.local/platform/internal/nodecontrol/nodeissuer')
    if($build.ShutdownSchema -ne 3 -or $build.ExitCode -ne 0 -or $build.Executable -ine (Join-Path $tool 'bin/go.exe') -or ($build.Arguments -join '|') -cne ($expectedBuild -join '|')){throw 'Bound successful original compile required'}
    $environment=@{GOROOT=$tool;GOCACHE=(Join-Path $compiled 'go-cache');GOMODCACHE=$modules;GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='on';GOWORK='off';GOFLAGS='-mod=readonly';GOOS='windows';GOARCH='amd64';TEMP=(Join-Path $compiled 'tmp');TMP=(Join-Path $compiled 'tmp');GOTMPDIR=(Join-Path $compiled 'tmp');TALENRO_FUZZ_DIAG_MODE='';TALENRO_FUZZ_FIXTURE='';GODEBUG=''}
    foreach($name in $environment.Keys){if($build.Environment.$name -cne $environment[$name]){throw 'Prior compile environment mismatch'}}
    $log=Get-Content -LiteralPath $paths['build/stderr.txt'] -Raw -ErrorAction Stop
    Assert-CopiedCompileEvidence $log $tool;Assert-EnvelopeCompileEvidence $log $tool $workspace
    $before=@(Read-EnvelopeSampleJson $paths['before-product.json'])
    Assert-EnvelopeProductUnchanged $before @(Read-EnvelopeSampleJson $paths['after-product.json'])
    Assert-EnvelopeProductUnchanged $before @(Get-EnvelopeProductManifest)
    foreach($name in @('GOCACHE','TEMP','TMP','GOTMPDIR')){$environment[$name]=Join-Path $root $(if($name -eq 'GOCACHE'){'go-cache'}else{'tmp'})}
    [pscustomobject]@{RunRoot=$root;CompileRoot=$compiled;Toolchain=$tool;ReferenceReceipt=$ReferenceReceipt;ToolchainReference=$identity;Binary=$binary;BinaryHash=$binaryHash;ProductManifest=$before;ArtifactHashes=$hashes;Environment=$environment;ExecuteLimitSeconds=60;Arguments=@('-test.run=^$','-test.fuzz=^FuzzIssuerPair$','-test.fuzztime=5s','-test.parallel=1','-test.timeout=30s','-test.count=1',('-test.fuzzcachedir='+(Join-Path $root 'fuzz-cache')))}
}
function Assert-PairSampleInputs([object]$Plan,[object[]]$Product){
    foreach($row in $Plan.ArtifactHashes){if((Get-FileHash -LiteralPath (Get-EnvelopeSampleFile $row.Path)).Hash -cne $row.SHA256){throw 'Prior compiled artifact drift'}}
    if((Assert-UninstrumentedEnvelopeToolchain $Plan.Toolchain $Plan.ReferenceReceipt) -cne $Plan.ToolchainReference){throw 'Pair copied source binding drift'}
    Assert-EnvelopeProductUnchanged $Plan.ProductManifest $Product
}
function Invoke-PairCompiledSample([string]$RunRoot,[string]$CompileRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache){
    $plan=Get-PairCompiledPlan $RunRoot $CompileRoot $Toolchain $ReferenceReceipt $ModuleCache
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    $plan|ConvertTo-Json -Depth 7|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'plan.json') -ErrorAction Stop
    $plan.ProductManifest|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'before-product.json') -ErrorAction Stop
    foreach($name in @('go-cache','fuzz-cache','tmp')){New-Item -ItemType Directory -Path (Join-Path $plan.RunRoot $name) -ErrorAction Stop|Out-Null}
    $pending=$null;$attempted=$false;$stage='execute'
    try{
        $attempted=$true
        $receipt=Invoke-RecordedProcess $plan.Binary $plan.Arguments $plan.Environment (Join-Path $plan.RunRoot 'execute') $plan.ExecuteLimitSeconds
        $pending=[pscustomobject]@{Scope='One Pair validation from bound uninstrumented compile';Target='FuzzIssuerPair';Uninstrumented=$true;SamplingExecuted=$true;Acceptance=$false;ObservationValidationPassed=$false;ObservationValidationFailure=$null;RecordedExecutions=$null;BinaryHash=$plan.BinaryHash;ToolchainReference=$plan.ToolchainReference;CompileRoot=$plan.CompileRoot;Receipt=$receipt}
        $stage='validate';Assert-ProcessCompletion $receipt
        if($receipt.ShutdownSchema -ne 3 -or $receipt.ExitCode -ne 0 -or $receipt.TotalProcesses -lt 2){throw 'Pair exit/worker evidence invalid'}
        $stdout=@(Get-Content -LiteralPath $receipt.StdoutPath -ErrorAction Stop|Where-Object {$_ -ne ''})
        $stderr=Get-Content -LiteralPath $receipt.StderrPath -Raw -ErrorAction Stop
        $executionMatches=[regex]::Matches($stderr,'(?m)^fuzz:.*execs: ([0-9]+) ');$execs=[UInt64]0
        if($stdout.Count -eq 0 -or $stdout[-1] -cne 'PASS' -or $stderr -match '(?m)^TALENRO_FUZZ_DIAG ' -or $stderr -notmatch '(?m)^fuzz:.*now fuzzing with 1 workers\r?$' -or $executionMatches.Count -eq 0 -or -not [UInt64]::TryParse($executionMatches[$executionMatches.Count-1].Groups[1].Value,[ref]$execs) -or $execs -eq 0){throw 'Pair PASS/worker/execution proof invalid'}
        $pending.RecordedExecutions=$execs;$pending.ObservationValidationPassed=$true
    }catch{
        if($null -ne $pending){$pending.ObservationValidationFailure=$_.Exception.Message}
        [pscustomobject]@{Stage=$stage;Error=$_.Exception.Message;SamplingAttempted=$attempted;SamplingExecuted=($null -ne $pending);Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json') -ErrorAction Stop
        throw
    }finally{
        try{
            $after=@(Get-EnvelopeProductManifest);$after|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'after-product.json') -ErrorAction Stop
            Assert-PairSampleInputs $plan $after
            if($null -ne $pending){$pending|ConvertTo-Json -Depth 8|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'sample-result.json') -ErrorAction Stop;if($pending.ObservationValidationPassed){Write-Output $pending}}
        }catch{
            $errorText=$_.Exception.Message
            if($null -ne $pending){$pending.ObservationValidationPassed=$false;$pending.ObservationValidationFailure=$errorText;$pending|ConvertTo-Json -Depth 8|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'sample-result.json') -ErrorAction Stop}
            [pscustomobject]@{Stage='finalize';PriorStage=$stage;Error=$errorText;SamplingAttempted=$attempted;SamplingExecuted=($null -ne $pending);Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json') -ErrorAction Stop
            throw
        }
    }
}
