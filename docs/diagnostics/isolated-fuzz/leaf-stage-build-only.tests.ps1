param([ValidateSet('all','native-persistence','final-persistence','marker-persistence')][string]$ReviewFocus='all')
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'leaf-stage-build-only-fixture.ps1')
. (Join-Path $PSScriptRoot 'leaf-stage-build-only.ps1')
Check-BuildOnly ($null -ne (Get-Command Invoke-LeafStageBuildOnly -ErrorAction SilentlyContinue)) 'BuildOnlyControllerExists'
# Inspect routing before any scenario: no real native boundary is invoked here.
$native=(Get-Command Invoke-LeafStageBuildOnlyNative).Definition
$controller=(Get-Command Invoke-LeafStageBuildOnly).Definition
Check-BuildOnly ($native -match '\[IsolatedFuzzProcess\]::RunWithCompileStageRetainedDiagnostics\(' -and $native -match ',600\)' -and $controller -notmatch '\[IsolatedFuzzProcess\]::' -and ([regex]::Matches($controller,'Invoke-LeafStageBuildOnlyNative ')).Count -eq 1) 'STATIC SoleReplaceableNativeRoute'
if($ReviewFocus -cin @('all','native-persistence')){
$launchFailure=Invoke-BuildOnlyTestCase 'native-throw-persist'
Check-BuildOnly ($launchFailure.Failure -ceq 'controlled native exception' -and $launchFailure.Saved.Failure -ceq 'controlled native exception' -and $launchFailure.Saved.AuditErrors.Count -eq 1 -and $launchFailure.Saved.AuditErrors[0] -match '^launch-error-persistence:' -and $launchFailure.NativeCalls -eq 1 -and $launchFailure.ToolAuditCalls -eq 2 -and $launchFailure.SourceAuditCalls -eq 3 -and $launchFailure.Output.Count -eq 0 -and -not(Test-Path (Join-Path $launchFailure.Root 'completed.json') -PathType Leaf)) 'PrimaryNativeErrorSurvivesPersistenceFailure'
if($ReviewFocus -ceq 'native-persistence'){return}
}
if($ReviewFocus -cin @('all','final-persistence')){
    $resultFailure=Invoke-BuildOnlyTestCase 'primary-result-persist'
    $stopped=Get-Content (Join-Path $resultFailure.Root 'stopped.json') -Raw|ConvertFrom-Json
    Check-BuildOnly ($resultFailure.Failure -match 'Diagnostic compile root exit nonzero' -and $resultFailure.Failure -match 'persistence:' -and $stopped.Error -ceq 'Diagnostic compile root exit nonzero' -and $resultFailure.NativeCalls -eq 1 -and $resultFailure.Output.Count -eq 0) 'PrimaryErrorSurvivesFinalPersistenceFailure'
    $stoppedFailure=Invoke-BuildOnlyTestCase 'primary-stopped-persist'
    Check-BuildOnly ($stoppedFailure.Failure -ceq 'Diagnostic compile root exit nonzero' -and $stoppedFailure.Saved.Failure -ceq 'Diagnostic compile root exit nonzero' -and $stoppedFailure.Saved.AuditErrors.Count -eq 1 -and $stoppedFailure.Saved.AuditErrors[0] -match '^stopped-persistence:' -and -not $stoppedFailure.Saved.ObservationPassed) 'StoppedWriteFailureStillAttemptsResult'
    if($ReviewFocus -ceq 'final-persistence'){return}
}
if($ReviewFocus -cin @('all','marker-persistence')){
    foreach($scenario in @('marker-create-failure','marker-write-failure','marker-close-failure')){
        $markerFailure=Invoke-BuildOnlyTestCase $scenario
        Check-BuildOnly ($null -ne $markerFailure.Failure -and $markerFailure.NativeCalls -eq 1 -and $markerFailure.Output.Count -eq 0 -and (Test-Path (Join-Path $markerFailure.Root 'stopped.json') -PathType Leaf) -and -not(Test-Path (Join-Path $markerFailure.Root 'completed.json') -PathType Leaf)) ('NoPublishedMarkerAfterFault '+$scenario)
    }
    if($ReviewFocus -ceq 'marker-persistence'){return}
}
$good=Invoke-BuildOnlyTestCase 'success'
Check-BuildOnly ($null -eq $good.Failure -and $good.NativeCalls -eq 1 -and $good.ToolAuditCalls -eq 2 -and $good.ProductAuditCalls -eq 2 -and $good.SourceAuditCalls -eq 4 -and $good.Output.Count -eq 1) 'SingleCallAuditedCompletion'
Check-BuildOnly ($good.Output[0].Finalized -and $good.Saved.FinalizationRequired -and $good.Saved.ObservationPassed -and $good.Saved.CompileExitZero -and $good.Saved.CompileEvidenceComplete -and $good.Saved.ScopeEvidenceComplete -and $good.Saved.TriggerEvidenceComplete -and $good.Saved.Evidence.CoreNaturalComplete -and $good.Saved.Evidence.LifetimeEvidenceComplete) 'IndependentCompletionFacts'
foreach($record in @($good.Plan,$good.Saved,(Get-Content (Join-Path $good.Root 'completed.json') -Raw|ConvertFrom-Json))){
    Check-BuildOnly ($record.DiagnosticPolicy -ceq 'leaf-stage-build-only-v1' -and $record.VetMode -ceq 'off' -and $record.VetEvidenceComplete -is [bool] -and -not $record.VetEvidenceComplete -and -not $record.Acceptance -and -not $record.PackagePassed -and -not $record.OriginalTestPassed -and -not $record.RuntimeAttempted) 'ScopeAndFalseAcceptanceFlags'
}
$manifestPath=Join-Path $good.Root 'evidence-manifest.json';$manifest=Get-Content $manifestPath -Raw|ConvertFrom-Json
$marker=Get-Content (Join-Path $good.Root 'completed.json') -Raw|ConvertFrom-Json
Check-BuildOnly ($marker.Finalized -and $marker.ManifestSHA256 -ceq (Get-FileHash $manifestPath -Algorithm SHA256).Hash -and -not(Test-Path (Join-Path $good.Root 'stopped.json'))) 'BoundCompletionMarker'
Check-BuildOnly ($manifest.Files.Count -eq 10 -and @($manifest.Files.Path -match 'go-cache|tmp').Count -eq 0) 'BoundedManifestExcludesCaches'
foreach($row in $manifest.Files){$file=Join-Path $good.Root $row.Path;Check-BuildOnly ($row.SHA256 -ceq (Get-FileHash $file -Algorithm SHA256).Hash -and $row.Length -eq (Get-Item $file).Length) ('ManifestBinding '+$row.Path)}
$receipt=Get-Content (Join-Path $good.Root 'build/receipt.json') -Raw|ConvertFrom-Json
Check-BuildOnly ($receipt.LifetimeDiagnosticOnly -and $receipt.ShutdownDiagnosticOnly -and $receipt.BuildOnlyDiagnostic.Policy -ceq 'leaf-stage-build-only-v1' -and -not $receipt.BuildOnlyDiagnostic.VetEvidenceComplete -and $null -ne $receipt.CompileStageTrigger) 'ReceiptKeepsDiagnosticEvidence'
foreach($scenario in @('existing','wrong-root','pre-tool','native-throw','exit','timeout','forced','core87','missing-core','missing-trigger','bad-prefix','bad-clock','cap32','zero-entry','release-failure','scope-drift','missing-binary','compile-tokens','post-product','post-tool','source-drift','artifact-drift','multi-audit','persist','manifest','marker','after-product-persist')){
    $bad=Invoke-BuildOnlyTestCase $scenario
    $expected=if($scenario -cin @('existing','wrong-root','pre-tool')){0}else{1}
    Check-BuildOnly ($null -ne $bad.Failure -and $bad.NativeCalls -eq $expected -and $bad.Output.Count -eq 0 -and -not(Test-Path -LiteralPath (Join-Path $bad.Root 'completed.json') -PathType Leaf)) ('StoppedNoRetry '+$scenario)
    if($scenario -cnotin @('existing','wrong-root')){Check-BuildOnly (Test-Path -LiteralPath (Join-Path $bad.Root 'stopped.json') -PathType Leaf) ('DurableStopped '+$scenario)}
    if($scenario -ceq 'native-throw'){$launch=Get-Content (Join-Path $bad.Root 'build/launch-error.json') -Raw|ConvertFrom-Json;Check-BuildOnly (-not $launch.CleanupVerified -and -not(Test-Path (Join-Path $bad.Root 'build/receipt.json'))) 'LaunchCleanupUnverified'}
    if($scenario -ceq 'scope-drift'){Check-BuildOnly (-not $bad.Saved.ScopeEvidenceComplete -and $bad.Invocation[0].BuildArguments[4] -ceq '-vet=off') 'FixedScopeRefusesPostLaunchDrift'}
    if($scenario -ceq 'multi-audit'){Check-BuildOnly ($bad.Saved.Failure -ceq 'Diagnostic compile root exit nonzero' -and $bad.Saved.AuditErrors.Count -eq 3 -and $bad.ToolAuditCalls -eq 2 -and $bad.SourceAuditCalls -eq 4 -and $bad.Saved.ArtifactsUnchanged) 'IndependentFinalAuditsKeepPrimary'}
    if($scenario -ceq 'after-product-persist'){Check-BuildOnly ($bad.ToolAuditCalls -eq 2 -and $bad.SourceAuditCalls -eq 4 -and $bad.Saved.ArtifactsUnchanged) 'PersistenceFaultDoesNotSkipAudits'}
    if($scenario -cin @('persist','manifest','marker')){$name=switch($scenario){persist{'result.json'} manifest{'evidence-manifest.json'} marker{'completed.json'}};Check-BuildOnly (Test-Path -LiteralPath (Join-Path $bad.Root $name) -PathType Container) ('PreserveObstruction '+$scenario)}
}
Write-Output 'LIMIT: fake native/SDK/product boundaries only; no Go, binary execution, PID query or real inventory.'
