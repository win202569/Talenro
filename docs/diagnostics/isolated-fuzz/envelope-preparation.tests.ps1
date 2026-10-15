$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'envelope-observe.ps1')
function Check-Preparation([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$tool=Join-Path $scope 'framework-001/toolchain'
$modules='C:/Users/Lenovo/go/pkg/mod'
function Test-PreparationCase([string]$Scenario) {
    $target=Join-Path $scope ('envelope-preparation-tests-'+[guid]::NewGuid().ToString('N')+'/envelope-observe-901')
    $plan=Get-EnvelopeObservationPlan $target $tool $modules
    $calls=New-Object 'System.Collections.Generic.List[string]'
    $inventoryFile=Join-Path (Split-Path -Parent $target) 'product-fixture.go'
    New-Item -ItemType Directory -Path (Split-Path -Parent $target) -Force|Out-Null
    [IO.File]::WriteAllText($inventoryFile,'package fixture')
    # Controlled one-file inventory avoids repeated full-tree scans in branch
    # tests. Real planner, provenance, inventory comparison, completion gates,
    # driver branches and output file writes run; no observed Go process runs.
    $inventoryState=[pscustomobject]@{Calls=0}
    function Get-EnvelopeProductManifest {
        $inventoryState.Calls++
        if($inventoryState.Calls -eq 3 -and $Scenario -ceq 'final-drift'){[IO.File]::WriteAllText($inventoryFile,'package changed')}
        if($inventoryState.Calls -eq 3 -and $Scenario -ceq 'final-read'){throw 'Controlled final inventory read failure'}
        [pscustomobject]@{Path='product-fixture.go';SHA256=(Get-FileHash -LiteralPath $inventoryFile).Hash}
    }
    function Invoke-RecordedProcess([string]$Executable,[string[]]$Arguments,[hashtable]$Environment,[string]$ReceiptDir,[int]$LimitSeconds) {
        $stage=Split-Path -Leaf $ReceiptDir
        $calls.Add($stage)
        if($Executable -cne $plan.Executable -or $LimitSeconds -ne 60 -or $Environment.GOPROXY -cne 'off' -or $Environment.GOFLAGS -cne '-mod=readonly'){throw 'Wrong preparation process contract'}
        if($stage -ceq 'version' -and ($Arguments -join '|') -ceq 'version') {$text='go version go1.26.5 windows/amd64'}
        elseif($stage -ceq 'origin' -and ($Arguments -join '|') -ceq 'env|GOROOT|GOTOOLDIR|GOMOD') {$text=$plan.Toolchain+"`r`n"+(Join-Path $plan.Toolchain 'pkg/tool/windows_amd64')+"`r`n"+(Join-Path $plan.Project 'go.mod')}
        else {throw 'Forbidden build or execute reached test boundary'}
        if($Scenario -ceq 'origin-output' -and $stage -ceq 'origin') {$text='wrong-origin'}
        New-Item -ItemType Directory -Path $ReceiptDir -Force|Out-Null
        $stdout=Join-Path $ReceiptDir 'stdout.txt';$stderr=Join-Path $ReceiptDir 'stderr.txt'
        [IO.File]::WriteAllText($stdout,$text);[IO.File]::WriteAllText($stderr,'')
        $receipt=[pscustomobject]@{ExitCode=0;TimedOut=$false;RootExited=$true;CleanupVerified=$true;ActiveAfterCleanup=0;DescendantsRemaining=0;TotalProcesses=1;RootId=17;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;TerminationRequested=$false;ActiveBeforeTermination=0;MembersBeforeTermination=@();MembersAfterCleanup=@();MembersAtRootWait=@();NaturalObservationMs=0;ShutdownElapsedMs=0;ShutdownSchema=3;ShutdownStatePolicy='owned-final-signaled-exit-image-optional-v2';MemberEvidenceTruncated=$false;ShutdownMemberEvidence=@();StdoutPath=$stdout;StderrPath=$stderr}
        if(($Scenario -ceq 'version-gate' -and $stage -ceq 'version') -or ($Scenario -ceq 'origin-gate' -and $stage -ceq 'origin')) {$receipt.NaturalShutdownVerified=$false;$receipt.TerminationRequested=$true}
        return $receipt
    }
    $result=$null;$failure=$null
    $observed=New-Object 'System.Collections.Generic.List[object]'
    try {
        if($Scenario -ceq 'default') {Invoke-EnvelopeObservation $target $tool $modules|ForEach-Object {$observed.Add($_)}}
        else {Invoke-EnvelopeObservation $target $tool $modules -PreparationOnly|ForEach-Object {$observed.Add($_)}}
        if($observed.Count -eq 1){$result=$observed[0]}
    } catch {$failure=$_.Exception.Message}
    [pscustomobject]@{Root=$target;Result=$result;Failure=$failure;Calls=@($calls.ToArray());Observed=@($observed.ToArray())}
}
$success=Test-PreparationCase 'success'
Check-Preparation ($null -eq $success.Failure -and $null -ne $success.Result -and $success.Result.PreparationPassed) 'preparation-only driver returns success before any build boundary'
Check-Preparation (($success.Calls -join '|') -ceq 'version|origin') 'preparation-only permits exactly version and origin commands'
Check-Preparation ($success.Result.PreparationOnly -and -not $success.Result.BuildExecuted -and -not $success.Result.SamplingExecuted -and -not $success.Result.Acceptance) 'preparation success cannot claim build sampling or acceptance'
$saved=Get-Content -LiteralPath (Join-Path $success.Root 'preparation-result.json') -Raw|ConvertFrom-Json
Check-Preparation ($saved.PreparationPassed -and $saved.PreparationOnly -and -not $saved.SamplingExecuted -and -not $saved.BuildExecuted -and -not $saved.Acceptance) 'saved preparation result preserves bounded labels'
Check-Preparation (@(@('build','execute','nodeissuer.test.exe','binary-receipt.json','sample-result.json','stopped.json')|Where-Object {Test-Path -LiteralPath (Join-Path $success.Root $_)}).Count -eq 0) 'success creates no build binary sampling or stopped artifacts'
$before=@(Get-Content -LiteralPath (Join-Path $success.Root 'before-product.json') -Raw|ConvertFrom-Json)
$after=@(Get-Content -LiteralPath (Join-Path $success.Root 'after-product.json') -Raw|ConvertFrom-Json)
Check-Preparation ($before.Count -gt 0 -and @(Compare-Object $before $after -Property Path,SHA256).Count -eq 0) 'early return still executes final product inventory check'
foreach($scenario in @('version-gate','origin-gate','origin-output')) {
    $failed=Test-PreparationCase $scenario
    $stopped=Get-Content -LiteralPath (Join-Path $failed.Root 'stopped.json') -Raw|ConvertFrom-Json
    $expectedStage=if($scenario -ceq 'version-gate'){'version'}else{'origin'}
    $expectedCalls=if($scenario -ceq 'version-gate'){'version'}else{'version|origin'}
    Check-Preparation ($null -ne $failed.Failure -and $stopped.Stage -ceq $expectedStage -and $stopped.PreparationOnly -and -not $stopped.SamplingExecuted -and -not $stopped.Acceptance -and ($failed.Calls -join '|') -ceq $expectedCalls) ('failed '+$scenario+' stops without retry or later stage')
    Check-Preparation ((Test-Path -LiteralPath (Join-Path $failed.Root 'after-product.json')) -and -not (Test-Path -LiteralPath (Join-Path $failed.Root 'preparation-result.json'))) ('failed '+$scenario+' preserves final inventory without success result')
}
$default=Test-PreparationCase 'default'
Check-Preparation ($default.Failure -ceq 'Wrong preparation process contract' -and ($default.Calls -join '|') -ceq 'version|origin|build' -and -not (Test-Path -LiteralPath (Join-Path $default.Root 'preparation-result.json'))) 'default still reaches its existing build boundary rather than silently becoming preparation-only'
foreach($scenario in @('final-drift','final-read')) {
    $failed=Test-PreparationCase $scenario
    Check-Preparation ($failed.Observed.Count -eq 0) ('failed '+$scenario+' cannot emit premature streaming success')
    Check-Preparation ($null -ne $failed.Failure -and -not (Test-Path -LiteralPath (Join-Path $failed.Root 'preparation-result.json'))) ('failed '+$scenario+' cannot leave a successful preparation result')
    $stopped=Get-Content -LiteralPath (Join-Path $failed.Root 'stopped.json') -Raw|ConvertFrom-Json
    Check-Preparation ($stopped.Stage -ceq 'finalize' -and $stopped.PreparationOnly -and -not $stopped.SamplingExecuted -and -not $stopped.Acceptance -and ($failed.Calls -join '|') -ceq 'version|origin') ('failed '+$scenario+' preserves finalization stop without more processes')
}
