. (Join-Path $PSScriptRoot 'mapping-single-check.ps1')
. (Join-Path $PSScriptRoot 'retained-child-observation.ps1')
# Dedicated diagnostic entry. Importing/planning never executes Go or a binary.
function Get-LeafRetainedCompilePlan([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot;$tool=Get-IsolatedFullPath $Toolchain;$modules=Get-IsolatedFullPath $ModuleCache
    if([IntPtr]::Size -ne 8 -or -not(Test-IsolatedWithin $root $scope) -or [IO.Path]::GetFileName($root) -cnotmatch '^leaf-retained-compile-[0-9]{3}$' -or (Test-Path -LiteralPath $root)){throw 'New owned leaf compile identity and 64-bit host required; no replay'}
    if(-not(Test-IsolatedWithin $tool $scope) -or [IO.Path]::GetFileName($tool) -cne 'toolchain' -or [IO.Path]::GetFileName((Split-Path -Parent $tool)) -cnotmatch '^uninstrumented-[0-9]{3}$'){throw 'Owned fixed uninstrumented toolchain required'}
    foreach($inputRoot in @($tool,$modules)){if($root -ieq $inputRoot -or (Test-IsolatedWithin $root $inputRoot) -or (Test-IsolatedWithin $inputRoot $root)){throw 'Output/input overlap forbidden'}}
    $reference=Get-EnvelopeSampleFile $ReferenceReceipt
    if(-not(Test-IsolatedWithin $reference $scope)){throw 'Owned reference required'}
    $go=Get-EnvelopeSampleFile (Join-Path $tool 'bin/go.exe')
    $null=Get-EnvelopeSampleFile (Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod')
    $binding=@(foreach($name in @('leaf-retained-compile.ps1','retained-child-observation.ps1','mapping-single-check.ps1','matrix-check.ps1','envelope-uninstrumented.ps1','envelope-sample.ps1','envelope-observe.ps1','run.ps1','isolation.ps1','provenance.ps1','process-native.cs','retained-lifetime.cs','compile-stage-trigger.cs','compile-stage-session.cs','ordinary-ledger.cs','ordinary-ledger-session.cs','ordinary-ledger-receipt.ps1')){
        $path=Get-EnvelopeSampleFile (Join-Path $PSScriptRoot $name)
        [pscustomobject]@{Path=$name;SHA256=(Get-FileHash -LiteralPath $path -Algorithm SHA256 -ErrorAction Stop).Hash}
    })
    $binary=Join-Path $root 'package.test.exe'
    $environment=@{GOROOT=$tool;GOCACHE=(Join-Path $root 'go-cache');GOMODCACHE=$modules;GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='on';GOWORK='off';GOFLAGS='-mod=readonly';GOOS='windows';GOARCH='amd64';TEMP=(Join-Path $root 'tmp');TMP=(Join-Path $root 'tmp');GOTMPDIR=(Join-Path $root 'tmp');TALENRO_FUZZ_DIAG_MODE='';TALENRO_FUZZ_FIXTURE='';GODEBUG=''}
    [pscustomobject]@{Schema=1;RunRoot=$root;Toolchain=$tool;ReferenceReceipt=$reference;Project=$workspace;Package='internal/nodecontrol/nodeleafbinding';Executable=$go;ExecutableSHA256=(Get-FileHash -LiteralPath $go -Algorithm SHA256 -ErrorAction Stop).Hash;SourceBinding=$binding;Binary=$binary;Environment=$environment;BuildLimitSeconds=600;BuildArguments=@('test','-c','-x','-p=1','-count=1','-timeout=120s',('-o='+$binary),'talenro.local/platform/internal/nodecontrol/nodeleafbinding');RuntimeAttempted=$false;DiagnosticOnly=$true;Acceptance=$false;NoRetry=$true;Coverage='sampled'}
}
function Assert-LeafRetainedCompileSources($Plan) {
    Assert-OrdinaryLedgerSources $Plan.SourceBinding
    foreach($row in $Plan.SourceBinding){if((Get-FileHash -LiteralPath (Get-EnvelopeSampleFile (Join-Path $PSScriptRoot $row.Path)) -Algorithm SHA256 -ErrorAction Stop).Hash -cne $row.SHA256){throw 'Leaf compile helper source drift'}}
    if((Get-FileHash -LiteralPath (Get-EnvelopeSampleFile $Plan.Executable) -Algorithm SHA256 -ErrorAction Stop).Hash -cne $Plan.ExecutableSHA256){throw 'Leaf compile executable drift'}
}
function Get-LeafRetainedCompileEvidence($Receipt) {
    $core=$false;$coreError=$null;$lifetime=$false;$lifetimeError=$null
    try {
        if($Receipt.LifetimeDiagnosticOnly -isnot [bool] -or -not $Receipt.LifetimeDiagnosticOnly -or $Receipt.ShutdownDiagnosticOnly -isnot [bool] -or -not $Receipt.ShutdownDiagnosticOnly -or $Receipt.LiveSampling){throw 'Retained-only diagnostic flags required'}
        if($Receipt.ShutdownSchema -ne 3 -or $Receipt.ShutdownStatePolicy -cne 'owned-final-signaled-exit-image-optional-v2' -or $Receipt.RootId -le 0 -or $Receipt.ExitCode -ne 0 -or -not $Receipt.RootExited -or $Receipt.TimedOut -or $Receipt.TerminationRequested -or -not $Receipt.NaturalShutdownVerified -or -not $Receipt.ShutdownEvidenceComplete -or -not $Receipt.CleanupVerified -or $Receipt.MemberEvidenceTruncated){throw 'Core natural compile completion incomplete'}
        # DescendantsRemaining describes root-wait, not final residuals.
        if($Receipt.ActiveBeforeTermination -ne 0 -or $Receipt.ActiveAfterCleanup -ne 0 -or $Receipt.MembersBeforeTermination.Count -ne 0 -or $Receipt.MembersAfterCleanup.Count -ne 0 -or $Receipt.NaturalObservationMs -lt 0 -or $Receipt.NaturalObservationMs -gt 500 -or $Receipt.ShutdownElapsedMs -lt 0 -or $Receipt.ShutdownElapsedMs -gt 5000){throw 'Core cleanup bounds incomplete'}
        Assert-MappingQueryEvidence $Receipt
        $ids=New-Object 'System.Collections.Generic.HashSet[UInt64]';$rows=@($Receipt.ShutdownMemberEvidence)
        if($rows.Count -gt 32){throw 'Core member bound exceeded'}
        foreach($row in $rows){
            if($null -eq $row -or $row.Id -le 0 -or $row.Id -gt [uint32]::MaxValue -or -not $ids.Add([uint64]$row.Id) -or $null -ne $row.DiagnosticShutdown){throw 'Core member identity incomplete'}
            if($row.Id -eq $Receipt.RootId){continue}
            $member=New-Object 'IsolatedFuzzProcess+MemberObservation'
            foreach($field in @('Id','Owned','QueryError','ExitCodeRead','ExitCode','State','ConfirmationAttempted','ConfirmationState')){$member.$field=$row.$field}
            if($null -ne $row.FinalShutdown){
                $final=New-Object 'IsolatedFuzzProcess+ShutdownObservation'
                foreach($field in @('MemberId','SameHandleVerified','Attempted','State','QueryError','ExitCodeRead','ExitCode','WaitBudgetMs','StartElapsedMs','EndElapsedMs')){$final.$field=$row.FinalShutdown.$field};$member.FinalShutdown=$final
            }
            if(-not [IsolatedFuzzProcess]::IsFinalShutdownStateVerified($member)){throw 'Core final same-handle evidence missing'}
        }
        foreach($id in $Receipt.MembersAtRootWait){if(-not $ids.Contains([uint64]$id)){throw 'Core root-wait member missing final evidence'}}
        $core=$true
    }catch{$coreError=$_.Exception.Message}
    try {
        $d=$Receipt.LifetimeDiagnostics
        if($d.Schema -ne 1 -or $d.Coverage -cne 'sampled' -or $d.Incomplete -or $d.AdmissionBudgetExhausted -or $d.TimedOut -or -not $d.RootExited -or -not $d.TeardownAttempted -or $d.ReleaseFailures -ne 0 -or $d.Errors.Count -ne 0){throw 'Lifetime sample incomplete or capped'}
        # The producer marks attempt32 capped even if no33rd candidate arrives.
        if($d.Attempts -lt 1 -or $d.Attempts -ge 32 -or $d.Admitted -ne $d.Attempts -or $d.Entries.Count -ne $d.Attempts -or $d.ReleaseAttempts -ne $d.Admitted -or $d.QueryCount -lt 1){throw 'Lifetime entry/release counters incomplete or capped'}
        if(($d|ConvertTo-Json -Depth 6 -WarningAction Stop) -match '"[^"\r\n]*(Handle|Registry|Api)[^"\r\n]*"\s*:'){throw 'Internal handle property forbidden'}
        $frequency=$d.FirstQuery.Frequency;$ids=New-Object 'System.Collections.Generic.HashSet[UInt64]';$entries=New-Object 'System.Collections.Generic.HashSet[int]'
        if($d.LastQuery.StartTimestamp -lt $d.FirstQuery.StartTimestamp){throw 'Lifetime query ordering invalid'}
        foreach($e in $d.Entries){
            if($e.EntryId -lt 1 -or $e.EntryId -gt $d.Attempts -or -not $entries.Add([int]$e.EntryId) -or $e.Pid -le 0 -or $e.Pid -gt [uint32]::MaxValue -or $e.Pid -eq $Receipt.RootId -or -not $ids.Add([uint64]$e.Pid) -or $e.SourceQuery -lt 1 -or $e.SourceQuery -gt $d.QueryCount){throw 'Lifetime identity ambiguous'}
            $last=-1L
            foreach($row in @($e.SourceQueryEvidence,$e.Open,$e.Membership,$e.Admission,$e.RootWaitEnd,$e.Final)){
                if($null -eq $row -or $frequency -le 0 -or $row.Frequency -ne $frequency -or $row.StartTimestamp -lt $last -or $row.EndTimestamp -lt $row.StartTimestamp){throw 'Lifetime phase clocks missing or unordered'}
                $start=[DateTimeOffset]$row.StartUTC;$end=[DateTimeOffset]$row.EndUTC
                if($start -lt [DateTimeOffset]$Receipt.StartUTC -or $end -lt $start -or $end -gt [DateTimeOffset]$Receipt.EndUTC){throw 'Lifetime phase outside receipt interval'};$last=$row.EndTimestamp
            }
            foreach($q in @($e.SourceQueryEvidence,$d.FirstQuery,$d.LastQuery)){
                if(-not $q.Succeeded -or $q.Error -ne 0 -or $q.Frequency -ne $frequency -or $q.StartTimestamp -lt 0 -or $q.EndTimestamp -lt $q.StartTimestamp -or $q.Assigned -ne $q.Returned -or $q.Returned -gt 32 -or $q.Capacity -ne 32 -or $q.BufferBytes -ne 264){throw 'Lifetime source query incomplete'}
            }
            if(-not $e.Open.Succeeded -or $e.Open.Error -ne 0 -or -not $e.Membership.Succeeded -or $e.Membership.Error -ne 0 -or -not $e.Membership.Owned){throw 'Lifetime ownership unknown'}
            $exited=$false;$code=0
            foreach($r in @($e.Admission,$e.RootWaitEnd,$e.Final)){
                if(-not $r.Attempted -or -not $r.SameEntry -or $r.Ambiguous -or $r.TerminationRequested -or $r.Error -ne 0 -or -not $r.ExitCodeRead -or $r.State -cnotin @('running','exited') -or -not [string]::IsNullOrEmpty($r.SkipReason)){throw 'Lifetime state unknown or ambiguous'}
                if($exited -and ($r.State -ceq 'running' -or $r.ExitCode -ne $code)){throw 'Contradictory lifetime observations'}
                if($r.State -ceq 'exited'){$exited=$true;$code=$r.ExitCode}
            }
            if($e.Final.State -cne 'exited' -or $e.Final.ExitCode -ne 0){throw 'Lifetime final exit incomplete'}
        }
        $lifetime=$true
    }catch{$lifetimeError=$_.Exception.Message}
    [pscustomobject]@{DiagnosticOnly=$true;Coverage='sampled';Acceptance=$false;CoreNaturalComplete=$core;CoreError=$coreError;LifetimeEvidenceComplete=$lifetime;LifetimeError=$lifetimeError;ObservationPassed=($core -and $lifetime)}
}
function Invoke-LeafRetainedCompileNative($Plan,[string]$Stdout,[string]$Stderr) {
    [IsolatedFuzzProcess]::RunWithRetainedHandleDiagnostics($Plan.Executable,$Plan.BuildArguments,$Plan.Environment,$Plan.Project,$Stdout,$Stderr,600)
}
function Invoke-LeafRetainedCompile([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache) {
    $plan=Get-LeafRetainedCompilePlan $RunRoot $Toolchain $ReferenceReceipt $ModuleCache
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    $state=[pscustomobject]@{Schema=1;Scope='Fixed nodeleafbinding retained-only diagnostic compile';Package=$plan.Package;DiagnosticOnly=$true;Coverage='sampled';Acceptance=$false;PackagePassed=$false;OriginalTestPassed=$false;RuntimeAttempted=$false;NoRetry=$true;NativeAttempts=0;CompileExitZero=$false;CompileEvidenceComplete=$false;Evidence=$null;ToolchainReference=$null;ProductUnchanged=$false;ToolchainUnchanged=$false;ArtifactsUnchanged=$false;ObservationPassed=$false;FinalizationRequired=$true;Failure=$null;FinalizationError=$null;AuditErrors=(New-Object 'System.Collections.Generic.List[string]')}
    $before=@();$identity=$null;$binding=$null;$pending=$null
    try {
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'plan.json') $plan
        Assert-LeafRetainedCompileSources $plan
        $identity=Assert-UninstrumentedEnvelopeToolchain $plan.Toolchain $plan.ReferenceReceipt
        if($identity -cnotmatch '^[A-F0-9]{64}:[A-F0-9]{64}$'){throw 'Toolchain identity incomplete'}
        $state.ToolchainReference=$identity
        $before=@(Get-EnvelopeProductManifest);if($before.Count -eq 0){throw 'Product inventory empty'}
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'before-product.json') $before
        foreach($name in @('go-cache','tmp','build')){New-Item -ItemType Directory -Path (Join-Path $plan.RunRoot $name) -ErrorAction Stop|Out-Null}
        $stdout=Join-Path $plan.RunRoot 'build/stdout.txt';$stderr=Join-Path $plan.RunRoot 'build/stderr.txt'
        Assert-LeafRetainedCompileSources $plan
        $state.NativeAttempts=1
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'attempt.json') ([pscustomobject]@{Phase='compile-only';NativeAttempts=1;RuntimeAttempted=$false;Acceptance=$false;NoRetry=$true})
        try{$native=Invoke-LeafRetainedCompileNative $plan $stdout $stderr}catch{
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
        $state.Evidence=Get-LeafRetainedCompileEvidence $receipt
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
            @{Name='sources';Check={Assert-LeafRetainedCompileSources $plan}},
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
