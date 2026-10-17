. (Join-Path $PSScriptRoot 'leaf-retained-compile.ps1')
function Get-LeafStageRetainedCompilePlan([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot;$tool=Get-IsolatedFullPath $Toolchain;$modules=Get-IsolatedFullPath $ModuleCache
    if([IntPtr]::Size -ne 8 -or -not(Test-IsolatedWithin $root $scope) -or [IO.Path]::GetFileName($root) -cnotmatch '^leaf-stage-retained-compile-[0-9]{3}$' -or (Test-Path -LiteralPath $root)){throw 'New owned leaf compile identity and 64-bit host required; no replay'}
    if(-not(Test-IsolatedWithin $tool $scope) -or [IO.Path]::GetFileName($tool) -cne 'toolchain' -or [IO.Path]::GetFileName((Split-Path -Parent $tool)) -cnotmatch '^uninstrumented-[0-9]{3}$'){throw 'Owned fixed uninstrumented toolchain required'}
    foreach($inputRoot in @($tool,$modules)){if($root -ieq $inputRoot -or (Test-IsolatedWithin $root $inputRoot) -or (Test-IsolatedWithin $inputRoot $root)){throw 'Output/input overlap forbidden'}}
    $reference=Get-EnvelopeSampleFile $ReferenceReceipt
    if(-not(Test-IsolatedWithin $reference $scope)){throw 'Owned reference required'}
    $go=Get-EnvelopeSampleFile (Join-Path $tool 'bin/go.exe')
    $null=Get-EnvelopeSampleFile (Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod')
    $binding=@(foreach($name in @('leaf-stage-retained-compile.ps1','leaf-retained-compile.ps1','retained-child-observation.ps1','mapping-single-check.ps1','matrix-check.ps1','envelope-uninstrumented.ps1','envelope-sample.ps1','envelope-observe.ps1','run.ps1','isolation.ps1','provenance.ps1','process-native.cs','retained-lifetime.cs','compile-stage-trigger.cs','compile-stage-session.cs','ordinary-ledger.cs','ordinary-ledger-session.cs','ordinary-ledger-receipt.ps1')){
        $path=Get-EnvelopeSampleFile (Join-Path $PSScriptRoot $name)
        [pscustomobject]@{Path=$name;SHA256=(Get-FileHash -LiteralPath $path -Algorithm SHA256 -ErrorAction Stop).Hash}
    })
    $binary=Join-Path $root 'package.test.exe'
    $environment=@{GOROOT=$tool;GOCACHE=(Join-Path $root 'go-cache');GOMODCACHE=$modules;GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='on';GOWORK='off';GOFLAGS='-mod=readonly';GOOS='windows';GOARCH='amd64';TEMP=(Join-Path $root 'tmp');TMP=(Join-Path $root 'tmp');GOTMPDIR=(Join-Path $root 'tmp');TALENRO_FUZZ_DIAG_MODE='';TALENRO_FUZZ_FIXTURE='';GODEBUG=''}
    [pscustomobject]@{Schema=1;RunRoot=$root;Toolchain=$tool;ReferenceReceipt=$reference;Project=$workspace;Package='internal/nodecontrol/nodeleafbinding';Executable=$go;ExecutableSHA256=(Get-FileHash -LiteralPath $go -Algorithm SHA256 -ErrorAction Stop).Hash;SourceBinding=$binding;Trigger=[pscustomobject]@{Schema=1;Policy='leaf-test-compile-log-v1';Compiler=(Join-Path $tool 'pkg/tool/windows_amd64/compile.exe');SourceDirectory=(Join-Path $workspace 'internal/nodecontrol/nodeleafbinding');Package='talenro.local/platform/internal/nodecontrol/nodeleafbinding';TotalByteLimit=16777216;IterationByteLimit=65536;LineByteLimit=65536;PollLimit=60000};Binary=$binary;Environment=$environment;BuildLimitSeconds=600;BuildArguments=@('test','-c','-x','-p=1','-count=1','-timeout=120s',('-o='+$binary),'talenro.local/platform/internal/nodecontrol/nodeleafbinding');RuntimeAttempted=$false;DiagnosticOnly=$true;Acceptance=$false;NoRetry=$true;Coverage='sampled'}
}
function Assert-LeafStageRetainedCompileSources($Plan) {
    Assert-OrdinaryLedgerSources $Plan.SourceBinding
    foreach($row in $Plan.SourceBinding){if((Get-FileHash -LiteralPath (Get-EnvelopeSampleFile (Join-Path $PSScriptRoot $row.Path)) -Algorithm SHA256 -ErrorAction Stop).Hash -cne $row.SHA256){throw 'Leaf compile helper source drift'}}
    if((Get-FileHash -LiteralPath (Get-EnvelopeSampleFile $Plan.Executable) -Algorithm SHA256 -ErrorAction Stop).Hash -cne $Plan.ExecutableSHA256){throw 'Leaf compile executable drift'}
}
function Get-LeafStageRetainedCompileEvidence($Receipt,$Plan,[string]$StderrPath) {
    $base=Get-LeafRetainedCompileEvidence $Receipt
    $trigger=$false;$errorText=$null
    try {
        $t=$Receipt.CompileStageTrigger
        if($null -eq $t -or $t.Schema -ne 1 -or $t.Policy -cne 'leaf-test-compile-log-v1' -or $t.State -cne 'Activated' -or $t.Errors.Count -ne 0){throw 'Trigger missing, failed or unrecognized'}
        foreach($name in @('TotalByteLimit','IterationByteLimit','LineByteLimit','PollLimit')){if($t.$name -ne $Plan.Trigger.$name){throw 'Trigger limit differs from fixed plan'}}
        if($t.PollCount -lt 1 -or $t.PollCount -gt 60000 -or $null -eq $t.LineStart -or $null -eq $t.LineEnd -or $null -eq $t.LineNumber -or $t.LineStart -lt 0 -or $t.LineEnd -le $t.LineStart -or $t.LineEnd -gt 16777216 -or $t.LineNumber -lt 1 -or $t.BytesConsumed -ne $t.LineEnd){throw 'Trigger counters or offsets incomplete'}
        foreach($value in @($t.LineStart,$t.LineEnd,$t.LineNumber,$t.PollCount,$t.BytesConsumed)){if($value -ne [long]$value){throw 'Trigger counters must be integral'}}
        $raw=[Convert]::FromBase64String($t.RawLineBase64)
        if($raw.Length -ne ($t.LineEnd-$t.LineStart) -or $raw.Length -gt 65536 -or -not [CompileStageTrigger]::IsEligibleCommand($raw,$Plan.Trigger.Compiler,$Plan.Trigger.SourceDirectory)){throw 'Trigger line is not the fixed eligible command'}
        $hash=[Security.Cryptography.SHA256]::Create()
        try{$lineHash=([BitConverter]::ToString($hash.ComputeHash($raw))).Replace('-','')}finally{$hash.Dispose()}
        if($lineHash -cne $t.LineSHA256 -or $t.PrefixSHA256 -cnotmatch '^[A-F0-9]{64}$'){throw 'Trigger line hash differs'}
        $file=Get-EnvelopeSampleFile $StderrPath
        $stream=[IO.File]::Open($file,[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::Read)
        $hash=[Security.Cryptography.SHA256]::Create();$buffer=New-Object byte[] 65536;$lineBuffer=New-Object byte[] 65536;$lineLength=0;$position=0L;$lineCount=0L;$lastLineStart=0L
        try {
            if($stream.Length -lt $t.LineEnd){throw 'Trigger prefix missing from saved stream'}
            while($position -lt $t.LineEnd){
                $count=$stream.Read($buffer,0,[int][Math]::Min(65536,($t.LineEnd-$position)))
                if($count -le 0){throw 'Trigger prefix read incomplete'}
                for($i=0;$i -lt $count;$i++){
                    $at=$position+$i
                    if($lineLength -ge 65536){throw 'Trigger prefix line exceeds bound'}
                    $lineBuffer[$lineLength++]=$buffer[$i]
                    if($at -ge $t.LineStart -and $buffer[$i] -ne $raw[[int]($at-$t.LineStart)]){throw 'Trigger raw line changed'}
                    if($buffer[$i] -eq 10){
                        $lineCount++
                        if($at+1 -lt $t.LineEnd){
                            $prior=New-Object byte[] $lineLength;[Array]::Copy($lineBuffer,$prior,$lineLength)
                            if([CompileStageTrigger]::GetCommandDisposition($prior,$Plan.Trigger.Compiler,$Plan.Trigger.SourceDirectory) -ne 0){throw 'A prior eligible or malformed compiler candidate cannot be replaced'}
                            $lastLineStart=$at+1
                        }
                        $lineLength=0
                    }
                }
                $null=$hash.TransformBlock($buffer,0,$count,$buffer,0);$position+=$count
            }
            $null=$hash.TransformFinalBlock([byte[]]@(),0,0)
            if(([BitConverter]::ToString($hash.Hash)).Replace('-','') -cne $t.PrefixSHA256 -or $lineCount -ne $t.LineNumber -or $lastLineStart -ne $t.LineStart -or $lineLength -ne 0){throw 'Trigger prefix or line boundary differs'}
        }finally{$hash.Dispose();$stream.Dispose()}
        $frequency=$Receipt.LifetimeDiagnostics.FirstQuery.Frequency;$last=-1L;$lastUTC=[DateTimeOffset]$Receipt.StartUTC
        foreach($row in @($t.CandidateObserved,$t.RootCheck,$t.Activation,$t.FirstRegistryQuery)){
            if($null -eq $row -or $row.Succeeded -isnot [bool] -or -not $row.Succeeded -or $row.Error -ne 0 -or $frequency -le 0 -or $row.Frequency -ne $frequency -or $row.StartTimestamp -lt $last -or $row.EndTimestamp -lt $row.StartTimestamp){throw 'Trigger clock evidence missing or unordered'}
            $start=[DateTimeOffset]$row.StartUTC;$end=[DateTimeOffset]$row.EndUTC
            if($start -lt $lastUTC -or $end -lt $start -or $end -gt [DateTimeOffset]$Receipt.EndUTC){throw 'Trigger UTC evidence outside receipt or unordered'}
            $last=$row.EndTimestamp;$lastUTC=$end
        }
        if($t.RootCheck.WaitResult -ne 258){throw 'Trigger root was not confirmed unsignaled'}
        foreach($field in @('Succeeded','Error','Frequency','StartTimestamp','EndTimestamp','Assigned','Returned','Capacity','BufferBytes')){if($t.FirstRegistryQuery.$field -ne $Receipt.LifetimeDiagnostics.FirstQuery.$field){throw 'Trigger first query differs from registry'}}
        foreach($field in @('StartUTC','EndUTC')){if([DateTimeOffset]$t.FirstRegistryQuery.$field -ne [DateTimeOffset]$Receipt.LifetimeDiagnostics.FirstQuery.$field){throw 'Trigger first query UTC differs from registry'}}
        $first=$Receipt.LifetimeDiagnostics.FirstQuery;$final=$Receipt.LifetimeDiagnostics.LastQuery
        foreach($entry in $Receipt.LifetimeDiagnostics.Entries){
            $q=$entry.SourceQueryEvidence
            if($q.StartTimestamp -lt $first.StartTimestamp -or $q.EndTimestamp -gt $final.EndTimestamp -or $entry.Open.StartTimestamp -lt $t.Activation.EndTimestamp -or $entry.Admission.StartTimestamp -lt $first.EndTimestamp){throw 'Entry evidence precedes activation or exceeds registry query bounds'}
            if([DateTimeOffset]$q.StartUTC -lt [DateTimeOffset]$first.StartUTC -or [DateTimeOffset]$q.EndUTC -gt [DateTimeOffset]$final.EndUTC -or [DateTimeOffset]$entry.Open.StartUTC -lt [DateTimeOffset]$t.Activation.EndUTC -or [DateTimeOffset]$entry.Admission.StartUTC -lt [DateTimeOffset]$first.EndUTC){throw 'Entry UTC evidence is outside deferred query window'}
            if($entry.SourceQuery -gt 1 -and ($q.StartTimestamp -lt $first.EndTimestamp -or [DateTimeOffset]$q.StartUTC -lt [DateTimeOffset]$first.EndUTC)){throw 'Later source query precedes first query completion'}
            $expected=@();if($entry.SourceQuery -eq 1){$expected+=@($first)};if($entry.SourceQuery -eq $Receipt.LifetimeDiagnostics.QueryCount){$expected+=@($final)}
            foreach($expectedQuery in $expected){
                foreach($field in @('Succeeded','Error','Frequency','StartTimestamp','EndTimestamp','Assigned','Returned','Capacity','BufferBytes')){if($q.$field -ne $expectedQuery.$field){throw 'Entry source query contradicts report query metadata'}}
                foreach($field in @('StartUTC','EndUTC')){if([DateTimeOffset]$q.$field -ne [DateTimeOffset]$expectedQuery.$field){throw 'Entry source query contradicts report query UTC'}}
            }
        }
        $trigger=$true
    }catch{$errorText=$_.Exception.Message}
    [pscustomobject]@{DiagnosticOnly=$true;Coverage='sampled';Acceptance=$false;TriggerEvidenceComplete=$trigger;TriggerError=$errorText;CoreNaturalComplete=$base.CoreNaturalComplete;CoreError=$base.CoreError;LifetimeEvidenceComplete=$base.LifetimeEvidenceComplete;LifetimeError=$base.LifetimeError;ObservationPassed=($trigger -and $base.ObservationPassed)}
}
function Invoke-LeafStageRetainedCompileNative($Plan,[string]$Stdout,[string]$Stderr) {
    [IsolatedFuzzProcess]::RunWithCompileStageRetainedDiagnostics($Plan.Executable,$Plan.BuildArguments,$Plan.Environment,$Plan.Project,$Stdout,$Stderr,600)
}
function Invoke-LeafStageRetainedCompile([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache) {
    $plan=Get-LeafStageRetainedCompilePlan $RunRoot $Toolchain $ReferenceReceipt $ModuleCache
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    $state=[pscustomobject]@{Schema=1;Scope='Fixed nodeleafbinding compile-stage retained-only diagnostic compile';Package=$plan.Package;DiagnosticOnly=$true;Coverage='sampled';Acceptance=$false;PackagePassed=$false;OriginalTestPassed=$false;RuntimeAttempted=$false;NoRetry=$true;NativeAttempts=0;CompileExitZero=$false;CompileEvidenceComplete=$false;TriggerEvidenceComplete=$false;Evidence=$null;ToolchainReference=$null;ProductUnchanged=$false;ToolchainUnchanged=$false;ArtifactsUnchanged=$false;ObservationPassed=$false;FinalizationRequired=$true;Failure=$null;FinalizationError=$null;AuditErrors=(New-Object 'System.Collections.Generic.List[string]')}
    $before=@();$identity=$null;$binding=$null;$pending=$null
    try {
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'plan.json') $plan
        Assert-LeafStageRetainedCompileSources $plan
        $identity=Assert-UninstrumentedEnvelopeToolchain $plan.Toolchain $plan.ReferenceReceipt
        if($identity -cnotmatch '^[A-F0-9]{64}:[A-F0-9]{64}$'){throw 'Toolchain identity incomplete'}
        $state.ToolchainReference=$identity
        $before=@(Get-EnvelopeProductManifest);if($before.Count -eq 0){throw 'Product inventory empty'}
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'before-product.json') $before
        foreach($name in @('go-cache','tmp','build')){New-Item -ItemType Directory -Path (Join-Path $plan.RunRoot $name) -ErrorAction Stop|Out-Null}
        $stdout=Join-Path $plan.RunRoot 'build/stdout.txt';$stderr=Join-Path $plan.RunRoot 'build/stderr.txt'
        Assert-LeafStageRetainedCompileSources $plan
        $state.NativeAttempts=1
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'attempt.json') ([pscustomobject]@{Phase='compile-only';NativeAttempts=1;RuntimeAttempted=$false;Acceptance=$false;NoRetry=$true})
        try{$native=Invoke-LeafStageRetainedCompileNative $plan $stdout $stderr}catch{
            Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'build/launch-error.json') ([pscustomobject]@{Error=$_.Exception.Message;CleanupVerified=$false;Acceptance=$false});throw
        }
        $receipt=ConvertTo-RecordedProcessReceipt $native $stdout $stderr $plan.Executable $plan.BuildArguments $plan.Environment
        $receipt|Add-Member WorkingDirectory $plan.Project
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'build/receipt.json') $receipt
        $files=@((Join-Path $plan.RunRoot 'build/receipt.json'),$stdout,$stderr)
        if(Test-Path -LiteralPath $plan.Binary -PathType Leaf){$files+=@(Get-EnvelopeSampleFile $plan.Binary)}
        $binding=Get-MappingEvidenceBinding $files
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'artifact-binding.json') $binding
        $state.CompileExitZero=$receipt.ExitCode -eq 0
        $state.Evidence=Get-LeafStageRetainedCompileEvidence $receipt $plan $stderr
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
        # Each applicable audit is attempted even if another failed. Preserve
        # the primary compile error as well as every subsequent audit failure.
        $postChecks=@(
            @{Name='product';Check={if($before.Count -gt 0){$after=@(Get-EnvelopeProductManifest);Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'after-product.json') $after;Assert-EnvelopeProductUnchanged $before $after;$state.ProductUnchanged=$true}}},
            @{Name='toolchain';Check={if($null -ne $identity){if((Assert-UninstrumentedEnvelopeToolchain $plan.Toolchain $plan.ReferenceReceipt) -cne $identity){throw 'Toolchain drift during compile'};$state.ToolchainUnchanged=$true}}},
            @{Name='sources';Check={Assert-LeafStageRetainedCompileSources $plan}},
            @{Name='artifacts';Check={if($null -ne $binding){Assert-MatrixBinding $binding;$state.ArtifactsUnchanged=$true}}}
        )
        foreach($check in $postChecks){try{$action=$check.Check;& $action}catch{$state.AuditErrors.Add($check.Name+':'+$_.Exception.Message);if($null -eq $pending){$pending=$_}}}
        if($state.AuditErrors.Count -gt 0){$state.FinalizationError=$state.AuditErrors -join '; '}
        $state.ObservationPassed=$null -eq $pending -and $state.CompileEvidenceComplete -and $state.Evidence.ObservationPassed -and $state.ProductUnchanged -and $state.ToolchainUnchanged -and $state.ArtifactsUnchanged
        try {
            if($null -ne $pending){Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'stopped.json') ([pscustomobject]@{Error=$state.Failure;FinalizationError=$state.FinalizationError;NoRetry=$true;Acceptance=$false})}
            Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'result.json') $state
            # Explicit bounded artifacts only: never inventory the generated cache/tmp.
            $manifestFiles=@(foreach($name in @('plan.json','attempt.json','before-product.json','after-product.json','build/receipt.json','build/stdout.txt','build/stderr.txt','build/launch-error.json','package.test.exe','artifact-binding.json','result.json','stopped.json')){
                $path=Join-Path $plan.RunRoot $name
                if(Test-Path -LiteralPath $path -PathType Leaf){$file=Get-EnvelopeSampleFile $path;[pscustomobject]@{Path=$name;SHA256=(Get-FileHash -LiteralPath $file -Algorithm SHA256 -ErrorAction Stop).Hash;Length=(Get-Item -LiteralPath $file).Length}}
            })
            $manifest=Join-Path $plan.RunRoot 'evidence-manifest.json'
            Write-RetainedChildObservationJSON $manifest ([pscustomobject]@{Schema=1;Acceptance=$false;NoRetry=$true;Files=$manifestFiles})
            if($state.ObservationPassed){Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'completed.json') ([pscustomobject]@{Finalized=$true;ObservationPassed=$true;Acceptance=$false;PackagePassed=$false;RuntimeAttempted=$false;ManifestSHA256=(Get-FileHash -LiteralPath $manifest -Algorithm SHA256 -ErrorAction Stop).Hash})}
        }catch{
            $stopped=Join-Path $plan.RunRoot 'stopped.json'
            if(-not(Test-Path -LiteralPath $stopped)){Write-RetainedChildObservationJSON $stopped ([pscustomobject]@{Stage='persistence';Error=$_.Exception.Message;NoRetry=$true;Acceptance=$false})};throw
        }
    }
    if($null -ne $pending){throw $pending}
    if(-not $state.ObservationPassed){throw 'No verified diagnostic observation'}
    $state|Add-Member Finalized $true
    return $state
}
