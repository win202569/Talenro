. (Join-Path $PSScriptRoot 'matrix-check.ps1')
function Get-MappingSinglePlan([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot;$tool=Get-IsolatedFullPath $Toolchain;$reference=Get-IsolatedFullPath $ReferenceReceipt;$modules=Get-IsolatedFullPath $ModuleCache
    if(-not (Test-IsolatedWithin $root $scope) -or [IO.Path]::GetFileName($root) -cnotmatch '^mapping-check-[0-9]{3}$' -or (Test-Path -LiteralPath $root)){throw 'New owned mapping identity required; no replay'}
    if(-not (Test-IsolatedWithin $tool $scope) -or -not (Test-IsolatedWithin $reference $scope)){throw 'Owned fixed toolchain/reference required'}
    $go=Get-EnvelopeSampleFile (Join-Path $tool 'bin/go.exe')
    if(-not (Test-Path -LiteralPath (Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod') -PathType Leaf)){throw 'Offline dependencies missing; no download'}
    $binary=Join-Path $root 'package.test.exe'
    $environment=@{GOROOT=$tool;GOCACHE=(Join-Path $root 'go-cache');GOMODCACHE=$modules;GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='on';GOWORK='off';GOFLAGS='-mod=readonly';GOOS='windows';GOARCH='amd64';TEMP=(Join-Path $root 'tmp');TMP=(Join-Path $root 'tmp');GOTMPDIR=(Join-Path $root 'tmp');TALENRO_FUZZ_DIAG_MODE='';TALENRO_FUZZ_FIXTURE='';GODEBUG=''}
    [pscustomobject]@{RunRoot=$root;Project=$workspace;Toolchain=$tool;ReferenceReceipt=$reference;Executable=$go;Binary=$binary;Package='internal/nodecontrol/nodeissuermapping';WorkingDirectory=(Join-Path $workspace 'internal/nodecontrol/nodeissuermapping');Environment=$environment;BuildLimitSeconds=600;ExecuteLimitSeconds=150;BuildArguments=@('test','-c','-x','-p=1','-count=1','-timeout=120s',('-o='+$binary),'talenro.local/platform/internal/nodecontrol/nodeissuermapping');Arguments=@('-test.count=1','-test.timeout=120s','-test.v=true','-test.run=.')}
}
function Get-MappingEvidenceBinding([string[]]$Files) {
    [pscustomobject]@{Files=@(foreach($file in $Files){$path=Get-EnvelopeSampleFile $file;[pscustomobject]@{Path=$path;SHA256=(Get-FileHash -LiteralPath $path).Hash}})}
}
function Assert-MappingQueryEvidence([object]$Receipt) {
    $a=$Receipt.InitialAccountingQuery;$members=@($Receipt.InitialMemberQueries)
    if($null -eq $a -or $a.API -cne 'QueryInformationJobObject/Accounting' -or -not $a.Succeeded -or $a.Error -ne 0 -or -not $a.ValuesRead -or $a.Frequency -le 0 -or $a.StartTimestamp -le 0 -or $a.EndTimestamp -lt $a.StartTimestamp -or $members.Count -lt 1 -or $members.Count -gt 8){throw 'Initial query evidence incomplete'}
    $last=$a.EndTimestamp
    foreach($q in @($a)+$members){
        if([DateTimeOffset]$q.StartUTC -lt [DateTimeOffset]$Receipt.StartUTC -or [DateTimeOffset]$q.EndUTC -lt [DateTimeOffset]$q.StartUTC -or [DateTimeOffset]$q.EndUTC -gt [DateTimeOffset]$Receipt.EndUTC){throw 'Query timestamps outside execution'}
    }
    foreach($q in $members){
        if($q.API -cne 'QueryInformationJobObject/ProcessIdList' -or -not $q.HeaderRead -or $q.Frequency -ne $a.Frequency -or $q.StartTimestamp -lt $last -or $q.EndTimestamp -lt $q.StartTimestamp -or (-not $q.Succeeded -and $q.Error -ne 234)){throw 'Initial member query evidence invalid'}
        $last=$q.EndTimestamp
    }
    if(-not $members[-1].Succeeded -or $members[-1].Error -ne 0){throw 'Final member enumeration did not succeed'}
}
function Save-MappingProgress([object]$State,[string]$Root){$State|ConvertTo-Json -Depth 9|Set-Content -LiteralPath (Join-Path $Root 'progress.json') -ErrorAction Stop}
function Invoke-MappingSingleCheck([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache) {
    $plan=Get-MappingSinglePlan $RunRoot $Toolchain $ReferenceReceipt $ModuleCache
    $state=[pscustomobject]@{Scope='One fixed nodeissuermapping package; not a matrix replay';Package=$plan.Package;PackagePassed=$false;Acceptance=$false;CompilePassed=$false;ExecutionAttempted=$false;ExecutionPassed=$false;ProductUnchanged=$false;ToolchainUnchanged=$false;FinalizationFailed=$false;FinalizationError=$null;Failure=$null;Phase='prepare';ToolchainReference=$null;ProcessAttempts=(New-Object 'System.Collections.Generic.List[object]');Bindings=(New-Object 'System.Collections.Generic.List[object]')}
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    $before=@();$identity=$null;$pending=$null
    try {
        $plan|ConvertTo-Json -Depth 5|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'plan.json') -ErrorAction Stop
        Save-MappingProgress $state $plan.RunRoot
        $identity=Assert-UninstrumentedEnvelopeToolchain $plan.Toolchain $plan.ReferenceReceipt;$state.ToolchainReference=$identity
        $before=@(Get-EnvelopeProductManifest);$before|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'before-product.json') -ErrorAction Stop
        foreach($name in @('go-cache','tmp')){New-Item -ItemType Directory -Path (Join-Path $plan.RunRoot $name) -ErrorAction Stop|Out-Null}
        $state.Phase='build';$dir=Join-Path $plan.RunRoot 'build'
        $state.ProcessAttempts.Add([pscustomobject]@{Phase=$state.Phase;Receipt=(Join-Path $dir 'receipt.json')});Save-MappingProgress $state $plan.RunRoot
        $build=Invoke-MatrixRecordedProcess $plan.Executable $plan.BuildArguments $plan.Environment $plan.Project $dir $plan.BuildLimitSeconds
        $files=@((Join-Path $dir 'receipt.json'),$build.StdoutPath,$build.StderrPath)
        if($build.ExitCode -eq 0){$files+=@(Get-EnvelopeSampleFile $plan.Binary)}
        $state.Bindings.Add((Get-MappingEvidenceBinding $files))
        Assert-ProcessCompletion $build
        if($build.ShutdownSchema -ne 3 -or $build.ExitCode -ne 0){throw 'Single mapping compile failed'}
        $log=Get-Content -LiteralPath $build.StderrPath -Raw -ErrorAction Stop
        Assert-MatrixCompileEvidence $log $plan.Toolchain $plan.Project $plan.Package
        Assert-CopiedCompileEvidence $log $plan.Toolchain
        $state.CompilePassed=$true
        Assert-MatrixBinding $state.Bindings[0]
        $state.Phase='execute';$dir=Join-Path $plan.RunRoot 'execute';$state.ExecutionAttempted=$true
        $state.ProcessAttempts.Add([pscustomobject]@{Phase=$state.Phase;Receipt=(Join-Path $dir 'receipt.json')});Save-MappingProgress $state $plan.RunRoot
        $run=Invoke-MatrixRecordedProcess $plan.Binary $plan.Arguments $plan.Environment $plan.WorkingDirectory $dir $plan.ExecuteLimitSeconds
        $state.Bindings.Add((Get-MappingEvidenceBinding @((Join-Path $dir 'receipt.json'),$run.StdoutPath,$run.StderrPath)))
        Assert-ProcessCompletion $run
        if($run.ShutdownSchema -ne 3 -or $run.ExitCode -ne 0){throw 'Single mapping test failed'}
        $stdout=@(Get-Content -LiteralPath $run.StdoutPath -ErrorAction Stop|Where-Object {$_ -ne ''})
        $stderr=Get-Content -LiteralPath $run.StderrPath -Raw -ErrorAction Stop
        if($stdout.Count -eq 0 -or $stdout[-1] -cne 'PASS' -or @($stdout|Where-Object {$_ -cmatch '^--- PASS: Test'}).Count -eq 0 -or $stderr -match '(?m)^TALENRO_FUZZ_DIAG '){throw 'Single mapping matching-test/uninstrumented PASS evidence missing'}
        Assert-MappingQueryEvidence $run
        $state.ExecutionPassed=$true
    } catch {
        $state.Failure=$_.Exception.Message;$pending=$_
        [pscustomobject]@{Stage=$state.Phase;Error=$state.Failure;NoRetry=$true;Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json') -ErrorAction Stop
    } finally {
        try {
            if($before.Count -gt 0){$after=@(Get-EnvelopeProductManifest);$after|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'after-product.json') -ErrorAction Stop;Assert-EnvelopeProductUnchanged $before $after;$state.ProductUnchanged=$true}
            if($null -ne $identity){if((Assert-UninstrumentedEnvelopeToolchain $plan.Toolchain $plan.ReferenceReceipt) -cne $identity){throw 'Single mapping toolchain identity drift'};$state.ToolchainUnchanged=$true}
            foreach($binding in $state.Bindings){Assert-MatrixBinding $binding}
            $state.PackagePassed=$state.ExecutionPassed -and $null -eq $state.Failure -and $state.ProductUnchanged -and $state.ToolchainUnchanged
            $state|ConvertTo-Json -Depth 9|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'mapping-result.json') -ErrorAction Stop
        } catch {
            $state.PackagePassed=$false;$state.FinalizationFailed=$true;$state.FinalizationError=$_.Exception.Message
            [pscustomobject]@{Stage='finalize';Error=$state.FinalizationError;PrimaryFailure=$state.Failure;NoRetry=$true;Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json') -ErrorAction Stop
            try{$state|ConvertTo-Json -Depth 9|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'mapping-result.json') -ErrorAction Stop}catch{}
            throw
        }
    }
    if($null -ne $pending){throw $pending}
    if($state.PackagePassed){return $state}
}
