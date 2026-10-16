. (Join-Path $PSScriptRoot 'leaf-stage-retained-compile.ps1')
# Explicit diagnostic entry only; importing this file never launches a workload.
function Get-LeafStageBuildOnlySourceNames {
    @('leaf-stage-build-only.ps1','leaf-stage-retained-compile.ps1','leaf-retained-compile.ps1','retained-child-observation.ps1','mapping-single-check.ps1','matrix-check.ps1','envelope-uninstrumented.ps1','envelope-sample.ps1','envelope-observe.ps1','run.ps1','isolation.ps1','provenance.ps1','process-native.cs','retained-lifetime.cs','compile-stage-trigger.cs','compile-stage-session.cs')
}
function Get-LeafStageBuildOnlyPaths([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache,[bool]$RequireNew) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot;$tool=Get-IsolatedFullPath $Toolchain;$modules=Get-IsolatedFullPath $ModuleCache
    if([IntPtr]::Size -ne 8 -or -not(Test-IsolatedWithin $root $scope) -or [IO.Path]::GetFileName($root) -cnotmatch '^leaf-stage-build-only-[0-9]{3}$' -or ($RequireNew -and (Test-Path -LiteralPath $root))){throw 'New owned build-only identity and 64-bit host required; no replay'}
    if(-not(Test-IsolatedWithin $tool $scope) -or [IO.Path]::GetFileName($tool) -cne 'toolchain' -or [IO.Path]::GetFileName((Split-Path -Parent $tool)) -cnotmatch '^uninstrumented-[0-9]{3}$'){throw 'Owned fixed uninstrumented toolchain required'}
    foreach($inputRoot in @($tool,$modules)){if($root -ieq $inputRoot -or (Test-IsolatedWithin $root $inputRoot) -or (Test-IsolatedWithin $inputRoot $root)){throw 'Output/input overlap forbidden'}}
    $reference=Get-EnvelopeSampleFile $ReferenceReceipt
    if(-not(Test-IsolatedWithin $reference $scope)){throw 'Owned reference required'}
    $go=Get-EnvelopeSampleFile (Join-Path $tool 'bin/go.exe')
    $null=Get-EnvelopeSampleFile (Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod')
    [pscustomobject]@{RunRoot=$root;Toolchain=$tool;ReferenceReceipt=$reference;ModuleCache=$modules;Project=$workspace;Executable=$go}
}
function Get-LeafStageBuildOnlyFixedFields($Paths) {
    $root=$Paths.RunRoot;$tool=$Paths.Toolchain
    $binary=Join-Path $root 'package.test.exe'
    $environment=@{GOROOT=$tool;GOCACHE=(Join-Path $root 'go-cache');GOMODCACHE=$Paths.ModuleCache;GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='on';GOWORK='off';GOFLAGS='-mod=readonly';GOOS='windows';GOARCH='amd64';TEMP=(Join-Path $root 'tmp');TMP=(Join-Path $root 'tmp');GOTMPDIR=(Join-Path $root 'tmp');TALENRO_FUZZ_DIAG_MODE='';TALENRO_FUZZ_FIXTURE='';GODEBUG=''}
    [pscustomobject]@{Binary=$binary;Environment=$environment;BuildArguments=@('test','-c','-x','-p=1','-vet=off','-count=1','-timeout=120s',('-o='+$binary),'talenro.local/platform/internal/nodecontrol/nodeleafbinding');Trigger=[pscustomobject]@{Schema=1;Policy='leaf-test-compile-log-v1';Compiler=(Join-Path $tool 'pkg/tool/windows_amd64/compile.exe');SourceDirectory=(Join-Path $Paths.Project 'internal/nodecontrol/nodeleafbinding');Package='talenro.local/platform/internal/nodecontrol/nodeleafbinding';TotalByteLimit=16777216;IterationByteLimit=65536;LineByteLimit=65536;PollLimit=60000}}
}
function Get-LeafStageBuildOnlyPlan([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache) {
    $paths=Get-LeafStageBuildOnlyPaths $RunRoot $Toolchain $ReferenceReceipt $ModuleCache $true
    $fixed=Get-LeafStageBuildOnlyFixedFields $paths
    $binding=@(foreach($name in (Get-LeafStageBuildOnlySourceNames)){
        $path=Get-EnvelopeSampleFile (Join-Path $PSScriptRoot $name)
        [pscustomobject]@{Path=$name;SHA256=(Get-FileHash -LiteralPath $path -Algorithm SHA256 -ErrorAction Stop).Hash}
    })
    [pscustomobject]@{Schema=1;RunRoot=$paths.RunRoot;Toolchain=$paths.Toolchain;ReferenceReceipt=$paths.ReferenceReceipt;Project=$paths.Project;Package='internal/nodecontrol/nodeleafbinding';Executable=$paths.Executable;ExecutableSHA256=(Get-FileHash -LiteralPath $paths.Executable -Algorithm SHA256 -ErrorAction Stop).Hash;SourceBinding=$binding;Trigger=$fixed.Trigger;Binary=$fixed.Binary;Environment=$fixed.Environment;BuildLimitSeconds=600;BuildArguments=$fixed.BuildArguments;DiagnosticPolicy='leaf-stage-build-only-v1';VetMode='off';VetEvidenceComplete=$false;RuntimeAttempted=$false;DiagnosticOnly=$true;Acceptance=$false;PackagePassed=$false;OriginalTestPassed=$false;NoRetry=$true;Coverage='sampled'}
}
function Assert-LeafStageBuildOnlySources([object]$Plan) {
    $expected=@(Get-LeafStageBuildOnlySourceNames)
    $rows=@($Plan.SourceBinding)
    if($rows.Count -ne 16){throw 'Build-only source binding incomplete'}
    $seen=New-Object 'System.Collections.Generic.HashSet[string]' ([StringComparer]::Ordinal)
    foreach($row in $rows){
        if($row.Path -isnot [string] -or $row.Path -cnotin $expected -or -not $seen.Add($row.Path) -or $row.SHA256 -isnot [string] -or $row.SHA256 -cnotmatch '^[A-F0-9]{64}$'){throw 'Build-only source name/hash incomplete or duplicate'}
        $file=Get-EnvelopeSampleFile (Join-Path $PSScriptRoot $row.Path)
        if((Get-FileHash -LiteralPath $file -Algorithm SHA256 -ErrorAction Stop).Hash -cne $row.SHA256){throw 'Build-only helper source drift'}
    }
    if($Plan.Executable -cne (Join-Path $Plan.Toolchain 'bin/go.exe') -or $Plan.ExecutableSHA256 -isnot [string] -or $Plan.ExecutableSHA256 -cnotmatch '^[A-F0-9]{64}$' -or (Get-FileHash -LiteralPath (Get-EnvelopeSampleFile $Plan.Executable) -Algorithm SHA256 -ErrorAction Stop).Hash -cne $Plan.ExecutableSHA256){throw 'Build-only executable drift'}
}
function Assert-LeafStageBuildOnlyValue($Actual,$Expected) {
    $equal=$false
    if($Expected -is [string]){$equal=$Actual -is [string] -and $Actual -ceq $Expected}
    elseif($Expected -is [bool]){$equal=$Actual -is [bool] -and $Actual -eq $Expected}
    else{$equal=($Actual -is [int] -or $Actual -is [long]) -and $Actual -eq $Expected}
    if(-not $equal){throw 'Build-only fixed value or type mismatch'}
}
function Assert-LeafStageBuildOnlyMap($Actual,[hashtable]$Expected) {
    if($null -eq $Actual){throw 'Build-only map missing'}
    $map=New-Object 'System.Collections.Generic.Dictionary[string,object]' ([StringComparer]::Ordinal)
    if($Actual -is [System.Collections.IDictionary]){foreach($pair in $Actual.GetEnumerator()){$map.Add([string]$pair.Key,$pair.Value)}}
    else{foreach($property in $Actual.PSObject.Properties){$map.Add($property.Name,$property.Value)}}
    if($map.Count -ne $Expected.Count){throw 'Build-only map keys differ'}
    foreach($key in $Expected.Keys){if(-not $map.ContainsKey($key)){throw 'Build-only fixed key missing'};Assert-LeafStageBuildOnlyValue $map[$key] $Expected[$key]}
}
function Assert-LeafStageBuildOnlyArguments($Actual,[object[]]$Expected) {
    if($Actual -is [string] -or @($Actual).Count -ne $Expected.Count){throw 'Build-only argument count/type mismatch'}
    for($i=0;$i -lt $Expected.Count;$i++){Assert-LeafStageBuildOnlyValue $Actual[$i] $Expected[$i]}
}
function Get-LeafStageBuildOnlyEvidence([object]$Receipt,[object]$Plan,[string]$StderrPath) {
    $scope=$false;$scopeError=$null
    try {
        # Validate constants independently of both records; two altered records
        # agreeing with each other are not a fixed diagnostic plan.
        $paths=Get-LeafStageBuildOnlyPaths $Plan.RunRoot $Plan.Toolchain $Plan.ReferenceReceipt $Plan.Environment.GOMODCACHE $false
        $fixed=Get-LeafStageBuildOnlyFixedFields $paths
        $values=@{Schema=1;RunRoot=$paths.RunRoot;Toolchain=$paths.Toolchain;ReferenceReceipt=$paths.ReferenceReceipt;Project=$paths.Project;Executable=$paths.Executable;Package='internal/nodecontrol/nodeleafbinding';Binary=$fixed.Binary;BuildLimitSeconds=600;DiagnosticPolicy='leaf-stage-build-only-v1';VetMode='off';VetEvidenceComplete=$false;DiagnosticOnly=$true;Coverage='sampled';NoRetry=$true;Acceptance=$false;PackagePassed=$false;OriginalTestPassed=$false;RuntimeAttempted=$false}
        foreach($key in $values.Keys){Assert-LeafStageBuildOnlyValue $Plan.$key $values[$key]}
        Assert-LeafStageBuildOnlyArguments $Plan.BuildArguments $fixed.BuildArguments
        Assert-LeafStageBuildOnlyMap $Plan.Environment $fixed.Environment
        $trigger=@{};foreach($property in $fixed.Trigger.PSObject.Properties){$trigger[$property.Name]=$property.Value}
        Assert-LeafStageBuildOnlyMap $Plan.Trigger $trigger
        Assert-LeafStageBuildOnlySources $Plan
        Assert-LeafStageBuildOnlyMap $Receipt.BuildOnlyDiagnostic @{Schema=1;Policy='leaf-stage-build-only-v1';VetMode='off';VetEvidenceComplete=$false}
        Assert-LeafStageBuildOnlyValue $Receipt.Executable $paths.Executable
        Assert-LeafStageBuildOnlyValue $Receipt.WorkingDirectory $paths.Project
        Assert-LeafStageBuildOnlyArguments $Receipt.Arguments $fixed.BuildArguments
        Assert-LeafStageBuildOnlyMap $Receipt.Environment $fixed.Environment
        $stdout=Get-EnvelopeSampleFile (Join-Path $paths.RunRoot 'build/stdout.txt')
        $stderr=Get-EnvelopeSampleFile (Join-Path $paths.RunRoot 'build/stderr.txt')
        Assert-LeafStageBuildOnlyValue $Receipt.StdoutPath $stdout
        Assert-LeafStageBuildOnlyValue $Receipt.StderrPath $stderr
        Assert-LeafStageBuildOnlyValue $StderrPath $stderr
        $scope=$true
    }catch{$scopeError=$_.Exception.Message}
    $base=Get-LeafStageRetainedCompileEvidence $Receipt $Plan $StderrPath
    [pscustomobject]@{DiagnosticOnly=$true;Coverage='sampled';Acceptance=$false;VetEvidenceComplete=$false;ScopeEvidenceComplete=$scope;ScopeError=$scopeError;TriggerEvidenceComplete=$base.TriggerEvidenceComplete;TriggerError=$base.TriggerError;CoreNaturalComplete=$base.CoreNaturalComplete;CoreError=$base.CoreError;LifetimeEvidenceComplete=$base.LifetimeEvidenceComplete;LifetimeError=$base.LifetimeError;ObservationPassed=($scope -and $base.ObservationPassed)}
}
function Invoke-LeafStageBuildOnlyNative([object]$Plan,[string]$Stdout,[string]$Stderr) {
    [IsolatedFuzzProcess]::RunWithCompileStageRetainedDiagnostics($Plan.Executable,$Plan.BuildArguments,$Plan.Environment,$Plan.Project,$Stdout,$Stderr,600)
}
function Invoke-LeafStageBuildOnly([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache) {
    $plan=Get-LeafStageBuildOnlyPlan $RunRoot $Toolchain $ReferenceReceipt $ModuleCache
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    $state=[pscustomobject]@{Schema=1;Scope='Fixed nodeleafbinding build-only compile-stage diagnostic';Package=$plan.Package;DiagnosticPolicy='leaf-stage-build-only-v1';VetMode='off';VetEvidenceComplete=$false;DiagnosticOnly=$true;Coverage='sampled';Acceptance=$false;PackagePassed=$false;OriginalTestPassed=$false;RuntimeAttempted=$false;NoRetry=$true;NativeAttempts=0;CompileExitZero=$false;CompileEvidenceComplete=$false;ScopeEvidenceComplete=$false;TriggerEvidenceComplete=$false;Evidence=$null;ToolchainReference=$null;ProductUnchanged=$false;ToolchainUnchanged=$false;ArtifactsUnchanged=$false;ObservationPassed=$false;FinalizationRequired=$true;Failure=$null;FinalizationError=$null;AuditErrors=(New-Object 'System.Collections.Generic.List[string]')}
    $before=@();$identity=$null;$binding=$null;$pending=$null
    try {
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'plan.json') $plan
        Assert-LeafStageBuildOnlySources $plan
        $identity=Assert-UninstrumentedEnvelopeToolchain $plan.Toolchain $plan.ReferenceReceipt
        if($identity -cnotmatch '^[A-F0-9]{64}:[A-F0-9]{64}$'){throw 'Toolchain identity incomplete'}
        $state.ToolchainReference=$identity
        $before=@(Get-EnvelopeProductManifest);if($before.Count -eq 0){throw 'Product inventory empty'}
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'before-product.json') $before
        foreach($name in @('go-cache','tmp','build')){New-Item -ItemType Directory -Path (Join-Path $plan.RunRoot $name) -ErrorAction Stop|Out-Null}
        $stdout=Join-Path $plan.RunRoot 'build/stdout.txt';$stderr=Join-Path $plan.RunRoot 'build/stderr.txt'
        Assert-LeafStageBuildOnlySources $plan
        $state.NativeAttempts=1
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'attempt.json') ([pscustomobject]@{Phase='compile-only';DiagnosticPolicy='leaf-stage-build-only-v1';VetMode='off';VetEvidenceComplete=$false;NativeAttempts=1;RuntimeAttempted=$false;Acceptance=$false;NoRetry=$true})
        try{$native=Invoke-LeafStageBuildOnlyNative $plan $stdout $stderr}catch{
            $launchError=$_
            try{Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'build/launch-error.json') ([pscustomobject]@{Error=$launchError.Exception.Message;CleanupVerified=$false;Acceptance=$false})}
            catch{$state.AuditErrors.Add('launch-error-persistence:'+$_.Exception.Message)}
            throw $launchError
        }
        $receipt=ConvertTo-RecordedProcessReceipt $native $stdout $stderr $plan.Executable $plan.BuildArguments $plan.Environment
        $receipt|Add-Member WorkingDirectory $plan.Project
        $receipt|Add-Member BuildOnlyDiagnostic ([pscustomobject]@{Schema=1;Policy='leaf-stage-build-only-v1';VetMode='off';VetEvidenceComplete=$false})
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'build/receipt.json') $receipt
        $files=@((Join-Path $plan.RunRoot 'build/receipt.json'),$stdout,$stderr)
        if(Test-Path -LiteralPath $plan.Binary -PathType Leaf){$files+=@(Get-EnvelopeSampleFile $plan.Binary)}
        $binding=Get-MappingEvidenceBinding $files
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'artifact-binding.json') $binding
        $state.CompileExitZero=$receipt.ExitCode -eq 0
        $state.Evidence=Get-LeafStageBuildOnlyEvidence $receipt $plan $stderr
        $state.ScopeEvidenceComplete=$state.Evidence.ScopeEvidenceComplete
        $state.TriggerEvidenceComplete=$state.Evidence.TriggerEvidenceComplete
        if(-not $state.CompileExitZero){throw 'Diagnostic compile root exit nonzero'}
        $null=Get-EnvelopeSampleFile $plan.Binary
        $log=Get-Content -LiteralPath $stderr -Raw -ErrorAction Stop
        Assert-MatrixCompileEvidence $log $plan.Toolchain $plan.Project $plan.Package
        Assert-CopiedCompileEvidence $log $plan.Toolchain
        if($log -match '(?m)^TALENRO_FUZZ_DIAG '){throw 'Unexpected instrumentation output'}
        $state.CompileEvidenceComplete=$true
        if(-not $state.Evidence.ObservationPassed){throw 'Diagnostic compile sample incomplete; no retry'}
    }catch{$state.Failure=$_.Exception.Message;$pending=$_}finally{
        # Audit blocks are independent: preserve the primary error and every
        # later failure, including persistence failures inside an audit.
        $postChecks=@(
            @{Name='product';Check={if($before.Count -gt 0){$after=@(Get-EnvelopeProductManifest);Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'after-product.json') $after;Assert-EnvelopeProductUnchanged $before $after;$state.ProductUnchanged=$true}}},
            @{Name='toolchain';Check={if($null -ne $identity){if((Assert-UninstrumentedEnvelopeToolchain $plan.Toolchain $plan.ReferenceReceipt) -cne $identity){throw 'Toolchain drift during compile'};$state.ToolchainUnchanged=$true}}},
            @{Name='sources';Check={Assert-LeafStageBuildOnlySources $plan}},
            @{Name='artifacts';Check={if($null -ne $binding){Assert-MatrixBinding $binding;$state.ArtifactsUnchanged=$true}}}
        )
        foreach($check in $postChecks){try{$action=$check.Check;& $action}catch{$state.AuditErrors.Add($check.Name+':'+$_.Exception.Message);if($null -eq $pending){$pending=$_}}}
        if($state.AuditErrors.Count -gt 0){$state.FinalizationError=$state.AuditErrors -join '; '}
        $state.ObservationPassed=$null -eq $pending -and $state.CompileEvidenceComplete -and $state.ScopeEvidenceComplete -and $state.Evidence.ObservationPassed -and $state.ProductUnchanged -and $state.ToolchainUnchanged -and $state.ArtifactsUnchanged
        if($null -ne $pending){
            try{Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'stopped.json') ([pscustomobject]@{Error=$pending.Exception.Message;FinalizationError=$state.FinalizationError;NoRetry=$true;Acceptance=$false})}
            catch{$state.AuditErrors.Add('stopped-persistence:'+$_.Exception.Message);$state.FinalizationError=$state.AuditErrors -join '; '}
        }
        try {
            Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'result.json') $state
            # Only explicit artifacts: generated cache/tmp are never inventoried.
            $manifestFiles=@(foreach($name in @('plan.json','attempt.json','before-product.json','after-product.json','build/receipt.json','build/stdout.txt','build/stderr.txt','build/launch-error.json','package.test.exe','artifact-binding.json','result.json','stopped.json')){
                $path=Join-Path $plan.RunRoot $name
                if(Test-Path -LiteralPath $path -PathType Leaf){$file=Get-EnvelopeSampleFile $path;[pscustomobject]@{Path=$name;SHA256=(Get-FileHash -LiteralPath $file -Algorithm SHA256 -ErrorAction Stop).Hash;Length=(Get-Item -LiteralPath $file).Length}}
            })
            $manifest=Join-Path $plan.RunRoot 'evidence-manifest.json'
            Write-RetainedChildObservationJSON $manifest ([pscustomobject]@{Schema=1;DiagnosticPolicy='leaf-stage-build-only-v1';VetMode='off';VetEvidenceComplete=$false;Acceptance=$false;NoRetry=$true;Files=$manifestFiles})
            if($state.ObservationPassed){
                $completion=[pscustomobject]@{Finalized=$true;ObservationPassed=$true;DiagnosticPolicy='leaf-stage-build-only-v1';VetMode='off';VetEvidenceComplete=$false;Acceptance=$false;PackagePassed=$false;OriginalTestPassed=$false;RuntimeAttempted=$false;ManifestSHA256=(Get-FileHash -LiteralPath $manifest -Algorithm SHA256 -ErrorAction Stop).Hash}
                # Never create the final path until writing,closing and verifying
                # the staged bytes succeeded. Move is same-directory,no overwrite.
                $staged=Join-Path $plan.RunRoot 'completed.pending.json'
                Write-RetainedChildObservationJSON $staged $completion
                $safe=Get-EnvelopeSampleFile $staged
                $expected=[Text.Encoding]::UTF8.GetBytes(($completion|ConvertTo-Json -Depth 8 -WarningAction Stop))
                if([Convert]::ToBase64String([IO.File]::ReadAllBytes($safe)) -cne [Convert]::ToBase64String($expected)){throw 'Staged completion bytes differ'}
                [IO.File]::Move($safe,(Get-IsolatedFullPath (Join-Path $plan.RunRoot 'completed.json')))
            }
        }catch{
            $persistenceError=$_
            $state.ObservationPassed=$false
            $state.AuditErrors.Add('persistence:'+$persistenceError.Exception.Message)
            $state.FinalizationError=$state.AuditErrors -join '; '
            $stopped=Join-Path $plan.RunRoot 'stopped.json'
            $cause=if($null -ne $pending){$pending}else{$persistenceError}
            try{if(-not(Test-Path -LiteralPath $stopped)){Write-RetainedChildObservationJSON $stopped ([pscustomobject]@{Stage='persistence';Error=$cause.Exception.Message;FinalizationError=$state.FinalizationError;NoRetry=$true;Acceptance=$false})}}
            catch{$state.AuditErrors.Add('stopped-persistence:'+$_.Exception.Message);$state.FinalizationError=$state.AuditErrors -join '; '}
            throw [Exception]::new(($cause.Exception.Message+'; '+$state.FinalizationError),$cause.Exception)
        }
    }
    if($null -ne $pending){throw $pending}
    if(-not $state.ObservationPassed){throw 'No verified diagnostic observation'}
    $state|Add-Member Finalized $true
    return $state
}
