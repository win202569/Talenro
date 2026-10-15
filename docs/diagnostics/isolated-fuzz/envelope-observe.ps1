. (Join-Path $PSScriptRoot 'run.ps1')
function Get-EnvelopeObservationPlan([string]$RunRoot,[string]$Toolchain,[string]$ModuleCache) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot;$tool=Get-IsolatedFullPath $Toolchain;$modules=Get-IsolatedFullPath $ModuleCache
    if(-not (Test-IsolatedWithin $root $scope) -or [IO.Path]::GetFileName($root) -cnotmatch '^envelope-observe-[0-9]{3}$' -or (Test-Path -LiteralPath $root)) {throw 'Envelope sampling requires a new explicitly approved owned identity'}
    if(-not (Test-IsolatedWithin $tool $scope) -or -not (Test-Path -LiteralPath (Join-Path $tool 'bin/go.exe') -PathType Leaf)) {throw 'Verified copied toolchain required, never official source'}
    if(-not (Test-Path -LiteralPath $modules -PathType Container) -or -not (Test-Path -LiteralPath (Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod') -PathType Leaf)) {throw 'Existing dependency cache missing; no download'}
    if((Get-FileHash -LiteralPath (Join-Path $workspace 'internal/nodecontrol/nodeissuer/fuzz_test.go')).Hash -cne '318CF8DE04BA5D95E71E625ABF6AAED8DFD54B854A9A925B7A3004D4132D989E') {throw 'Original Envelope callback drift'}
    $binary=Join-Path $root 'nodeissuer.test.exe'
    $environment=@{GOROOT=$tool;GOCACHE=(Join-Path $root 'go-cache');GOMODCACHE=$modules;GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='on';GOWORK='off';GOFLAGS='-mod=readonly';GOOS='windows';GOARCH='amd64';TEMP=(Join-Path $root 'tmp');TMP=(Join-Path $root 'tmp');GOTMPDIR=(Join-Path $root 'tmp');TALENRO_FUZZ_DIAG_MODE='observe';TALENRO_FUZZ_FIXTURE='normal'}
    [pscustomobject]@{RunRoot=$root;Toolchain=$tool;Project=$workspace;Package='talenro.local/platform/internal/nodecontrol/nodeissuer';Mode='observe';Executable=(Join-Path $tool 'bin/go.exe');Binary=$binary;Environment=$environment;BuildLimitSeconds=600;ExecuteLimitSeconds=60;BuildArguments=@('test','-c','-x','-p=1','-count=1','-run=^$','-fuzz=^FuzzIssuerEnvelope$','-fuzztime=5s','-parallel=1','-timeout=30s',('-o='+$binary),'talenro.local/platform/internal/nodecontrol/nodeissuer');ExecuteArguments=@('-test.run=^$','-test.fuzz=^FuzzIssuerEnvelope$','-test.fuzztime=5s','-test.parallel=1','-test.timeout=30s','-test.count=1',('-test.fuzzcachedir='+(Join-Path $root 'fuzz-cache')))}
}
function Get-EnvelopeProductManifest {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
    $files=@(& rg --files -g '*.go' -g '*.sql' -g 'go.mod' -g 'go.sum' -g '!.superpowers/**' -g '!.worktrees/**' $workspace)
    if($LASTEXITCODE -ne 0){throw 'Product inventory enumeration failed'}
    foreach($path in ($files|Sort-Object)) {
        $file=Get-IsolatedFullPath $path
        if(-not (Test-IsolatedWithin $file $workspace) -or [IsolatedFuzzFileInfo]::Links($file) -ne 1){throw 'Product source path/link invalid'}
        [pscustomobject]@{Path=$file.Substring($workspace.Length+1).Replace('\','/');SHA256=(Get-FileHash -LiteralPath $file).Hash}
    }
}
function Assert-EnvelopeProductUnchanged([object[]]$Before,[object[]]$After) {
    if($Before.Count -eq 0 -or $After.Count -eq 0 -or @(Compare-Object $Before $After -Property Path,SHA256).Count) {throw 'Product inventory changed; stop without retry'}
}
function Assert-EnvelopeCompileEvidence([string]$Log,[string]$Toolchain,[string]$Project) {
    $compiler=Join-Path $Toolchain 'pkg/tool/windows_amd64/compile.exe'
    $sources=@('fuzz_test.go','der.go','signature.go')|ForEach-Object {Join-Path $Project ('internal/nodecontrol/nodeissuer/'+$_)}
    foreach($line in ($Log -split '\r?\n')) {
        $tokens=@([regex]::Matches($line,'"(?:\\.|[^"\\])*"|[^\s"]+')|ForEach-Object {$v=$_.Value;if($v.StartsWith('"')){$v=$v.Substring(1,$v.Length-2).Replace('\\','\').Replace('\"','"')};$v})
        if($tokens.Count -lt 4 -or $tokens[0] -cne $compiler){continue}
        $package=$false
        for($i=1;$i -lt $tokens.Count-1;$i++){if($tokens[$i] -ceq '-p' -and $tokens[$i+1] -ceq 'talenro.local/platform/internal/nodecontrol/nodeissuer'){$package=$true}}
        if($package -and @($sources|Where-Object {$tokens -cnotcontains $_}).Count -eq 0){return}
    }
    throw 'Original Envelope test/parser compile evidence missing'
}
function Invoke-EnvelopeObservation([string]$RunRoot,[string]$Toolchain,[string]$ModuleCache,[switch]$PreparationOnly,[switch]$CompileOnly) {
    if($PreparationOnly -and $CompileOnly){throw 'Preparation-only and compile-only modes are mutually exclusive'}
    $plan=Get-EnvelopeObservationPlan $RunRoot $Toolchain $ModuleCache
    $identity=Assert-EnvelopeRunProvenance $plan
    $before=@(Get-EnvelopeProductManifest)
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    $before|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'before-product.json')
    foreach($dir in @('go-cache','fuzz-cache','tmp')){New-Item -ItemType Directory -Path (Join-Path $plan.RunRoot $dir)|Out-Null}
    $stage='version'
    $prepared=$null
    $compiled=$null
    try {
        $version=Invoke-RecordedProcess $plan.Executable @('version') $plan.Environment (Join-Path $plan.RunRoot 'version') 60
        Assert-ProcessCompletion $version
        if($version.ExitCode -ne 0 -or (Get-Content -Raw $version.StdoutPath).Trim() -cne 'go version go1.26.5 windows/amd64'){throw 'Version verification failed'}
        $stage='origin'
        $origin=Invoke-RecordedProcess $plan.Executable @('env','GOROOT','GOTOOLDIR','GOMOD') $plan.Environment (Join-Path $plan.RunRoot 'origin') 60
        Assert-ProcessCompletion $origin
        $lines=@(Get-Content $origin.StdoutPath)
        if($origin.ExitCode -ne 0 -or $lines.Count -ne 3 -or $lines[0] -cne $plan.Toolchain -or -not $lines[1].StartsWith($plan.Toolchain+'\',[StringComparison]::OrdinalIgnoreCase) -or $lines[2] -ine (Join-Path $plan.Project 'go.mod')){throw 'Original project/copied compiler origin not verified'}
        if($PreparationOnly) {
            Assert-EnvelopeProductUnchanged $before @(Get-EnvelopeProductManifest)
            if((Assert-EnvelopeRunProvenance $plan) -cne $identity){throw 'Copied diagnostic source changed during preparation'}
            $prepared=[pscustomobject]@{Scope='Envelope version and origin preparation only';PreparationOnly=$true;PreparationPassed=$true;BuildExecuted=$false;SamplingExecuted=$false;Acceptance=$false;SourceIdentity=$identity;RunRoot=$plan.RunRoot;VersionReceipt=(Join-Path $plan.RunRoot 'version/receipt.json');OriginReceipt=(Join-Path $plan.RunRoot 'origin/receipt.json')}
            return
        }
        $stage='build'
        $build=Invoke-RecordedProcess $plan.Executable $plan.BuildArguments $plan.Environment (Join-Path $plan.RunRoot 'build') $plan.BuildLimitSeconds
        Assert-ProcessCompletion $build
        if($build.ExitCode -ne 0){throw 'Single build failed; dependencies are not downloaded or installed'}
        $log=Get-Content -Raw $build.StderrPath
        Assert-CopiedCompileEvidence $log $plan.Toolchain
        Assert-EnvelopeCompileEvidence $log $plan.Toolchain $plan.Project
        Assert-EnvelopeProductUnchanged $before @(Get-EnvelopeProductManifest)
        if((Assert-EnvelopeRunProvenance $plan) -cne $identity){throw 'Copied diagnostic source changed during build'}
        $binaryHash=(Get-FileHash -LiteralPath $plan.Binary).Hash
        [pscustomobject]@{SourceIdentity=$identity;BinaryHash=$binaryHash;OriginalFuzzHash='318CF8DE04BA5D95E71E625ABF6AAED8DFD54B854A9A925B7A3004D4132D989E';ProductBefore=(Join-Path $plan.RunRoot 'before-product.json');CompilerEvidence=$build.StderrPath;GOROOT=$plan.Toolchain;Project=$plan.Project}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'binary-receipt.json')
        if($CompileOnly) {
            $compiled=[pscustomobject]@{Scope='Envelope compilation only';CompileOnly=$true;BuildPassed=$true;BuildExecuted=$true;SamplingExecuted=$false;Acceptance=$false;SourceIdentity=$identity;BinaryHash=$binaryHash;RunRoot=$plan.RunRoot;Binary=$plan.Binary;BinaryReceipt=(Join-Path $plan.RunRoot 'binary-receipt.json');CompilerEvidence=$build.StderrPath}
            return
        }
        $stage='execute'
        $receipt=Invoke-RecordedProcess $plan.Binary $plan.ExecuteArguments $plan.Environment (Join-Path $plan.RunRoot 'execute') $plan.ExecuteLimitSeconds
        $trace=@(Get-Content $receipt.StderrPath|Where-Object {$_.StartsWith('TALENRO_FUZZ_DIAG ')}|ForEach-Object {$_.Substring(18)|ConvertFrom-Json})
        $failure=$null
        try {Assert-ArmResult 'observe' $receipt $trace;if($receipt.TotalProcesses -lt 2){throw 'Actual worker process evidence missing'}}catch {$failure=$_.Exception.Message}
        $result=[pscustomobject]@{Scope='Original FuzzIssuerEnvelope observation only';Acceptance=$false;SamplingExecuted=$true;ObservationValidationPassed=($null -eq $failure);ObservationValidationFailure=$failure;Outcome=$(if($receipt.ExitCode -eq 0){'observed-success-not-acceptance'}else{'observed-failure-preserved'});SourceIdentity=$identity;OriginalFuzzHash='318CF8DE04BA5D95E71E625ABF6AAED8DFD54B854A9A925B7A3004D4132D989E';BinaryHash=$binaryHash;Receipt=$receipt;Trace=$trace}
        $result|ConvertTo-Json -Depth 8|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'sample-result.json')
        Assert-ProcessCompletion $receipt
        if(@($trace|Where-Object {$_.watchdog}).Count){throw 'Watchdog gate failed; preserve sample, never retry'}
        if((Get-FileHash -LiteralPath $plan.Binary).Hash -cne $binaryHash -or (Assert-EnvelopeRunProvenance $plan) -cne $identity){throw 'Source/binary drift after observation'}
        return [pscustomobject]@{Outcome=$result.Outcome;ObservationValidationPassed=$result.ObservationValidationPassed;Failure=$failure;Acceptance=$false;ExitCode=$receipt.ExitCode;TraceCount=$trace.Count;ResultPath=(Join-Path $plan.RunRoot 'sample-result.json')}
    } catch {
        [pscustomobject]@{Stage=$stage;UTC=[DateTime]::UtcNow.ToString('o');Error=$_.Exception.Message;PreparationOnly=[bool]$PreparationOnly;CompileOnly=[bool]$CompileOnly;BuildExecuted=(Test-Path -LiteralPath (Join-Path $plan.RunRoot 'build'));SamplingExecuted=(Test-Path -LiteralPath (Join-Path $plan.RunRoot 'execute'));Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json')
        throw
    } finally {
        try {
            $after=@(Get-EnvelopeProductManifest)
            $after|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'after-product.json')
            Assert-EnvelopeProductUnchanged $before $after
            if($CompileOnly -and $null -ne $compiled) {
                if((Get-FileHash -LiteralPath $plan.Binary).Hash -cne $binaryHash -or (Assert-EnvelopeRunProvenance $plan) -cne $identity){throw 'Source/binary drift during compile finalization'}
                $compiled|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'compile-result.json') -ErrorAction Stop
                Write-Output $compiled
            }
            if($PreparationOnly -and $null -ne $prepared){$prepared|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'preparation-result.json');Write-Output $prepared}
        } catch {
            if($PreparationOnly -or $CompileOnly){[pscustomobject]@{Stage='finalize';PriorStage=$stage;UTC=[DateTime]::UtcNow.ToString('o');Error=$_.Exception.Message;PreparationOnly=[bool]$PreparationOnly;CompileOnly=[bool]$CompileOnly;BuildExecuted=(Test-Path -LiteralPath (Join-Path $plan.RunRoot 'build'));SamplingExecuted=$false;Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json')}
            throw
        }
    }
}
function Assert-EnvelopeRunProvenance([object]$Plan) {
    if((Get-FileHash -LiteralPath $Plan.Executable).Hash -cne 'D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F'){throw 'Pinned copied Go executable missing or drifted'}
    $parent=Split-Path -Parent $Plan.Toolchain
    $copy=Get-Content -Raw -LiteralPath (Join-Path $parent 'copy-receipt.json')|ConvertFrom-Json
    if($copy.Toolchain -cne $Plan.Toolchain -or $copy.BeforeManifest.Count -ne 11536){throw 'Full verified copy receipt missing'}
    $patch=Get-Content -Raw -LiteralPath (Join-Path $parent 'patch-receipt.json')|ConvertFrom-Json
    Get-DiagnosticSourceIdentity $Plan.Toolchain $patch
}
