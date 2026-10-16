param([ValidateSet('flow','all')][string]$ReviewFocus='all')
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'leaf-ordinary-check-fixture.ps1')
. (Join-Path $PSScriptRoot 'leaf-ordinary-check.ps1')
Check-LeafOrdinary ($null -ne (Get-Command Invoke-LeafOrdinaryCheck -ErrorAction SilentlyContinue)) 'ControllerExists'
$s=Invoke-LeafOrdinaryTestCase 'success'
Check-LeafOrdinary ($null -eq $s.Error -and $s.NativeCalls -eq 2 -and $s.Saved.CompilePassed -and $s.Saved.OriginalTestPassed -and -not $s.Saved.PackagePassed -and $s.Saved.FinalizationRequired) 'FixedOrdinaryFlowProvisional'
if($ReviewFocus -ceq 'all'){Check-LeafOrdinary (Test-Path -LiteralPath (Join-Path $s.Root 'completed.json') -PathType Leaf) 'SuccessHasDurableCompletedMarker'}
foreach($scenario in @('preflight','build-exit','build-timeout','build-error87','build-diagnostic','build-schema','build-query','build-truncated','build-unknown','missing-binary','missing-compile-evidence','binary-before-run','source-before-run','runtime-exit','runtime-timeout','runtime-forced','runtime-query','runtime-diagnostic','runtime-truncated','runtime-unknown','no-tests','no-final-pass','fail-line','diagnostic-stdout','diagnostic-stderr','native-throw','native-throw-persist')){
    $r=Invoke-LeafOrdinaryTestCase $scenario
    $want=if($scenario -ceq 'preflight'){0}elseif($scenario -like 'runtime-*' -or $scenario -cin @('no-tests','no-final-pass','fail-line','diagnostic-stdout','diagnostic-stderr')){2}else{1}
    Check-LeafOrdinary ($null -ne $r.Error -and $r.NativeCalls -eq $want -and -not(Test-Path -LiteralPath (Join-Path $r.Root 'completed.json'))) ('RefusedWithoutRetry '+$scenario)
    if($scenario -ceq 'native-throw-persist'){Check-LeafOrdinary ($r.Error -match 'controlled native exception' -and $r.Saved.AuditErrors -match 'launch-error-persistence') 'NativePrimaryPreserved'}
}
$r=Invoke-LeafOrdinaryTestCase 'skip-and-subtest'
Check-LeafOrdinary ($null -eq $r.Error -and $r.Saved.TestOutput.TopLevelPass.Count -eq 1 -and $r.Saved.TestOutput.TopLevelSkip.Count -eq 1 -and $r.Saved.TestOutput.SubtestPass.Count -eq 1 -and $r.Saved.TestOutput.TopLevelSkip[0] -ceq 'TestSkipped') 'SkipAndSubtestDisclosed'
if($ReviewFocus -ceq 'all'){
    foreach($scenario in @('post-product','post-sdk','post-artifact','all-audits-fail','result-write','manifest-write','pending-write','marker-publish','stopped-write','native-throw-all-audits-fail')){
        $x=Invoke-LeafOrdinaryTestCase $scenario
        Check-LeafOrdinary ($null -ne $x.Error -and -not(Test-Path -LiteralPath (Join-Path $x.Root 'completed.json') -PathType Leaf)) ('FinalizationRefused '+$scenario)
        if($scenario -cin @('all-audits-fail','native-throw-all-audits-fail')){
            Check-LeafOrdinary ($x.AuditCalls.Product -eq 2 -and $x.AuditCalls.SDK -eq 2 -and $x.AuditCalls.Artifact -ge 1 -and $x.Saved.AuditErrors.Count -ge 3) ('IndependentAudits '+$scenario)
        }
        if($scenario -ceq 'native-throw-all-audits-fail'){Check-LeafOrdinary ($x.Error -match 'controlled native exception') 'PrimarySurvivesAllAudits'}
    }
}
