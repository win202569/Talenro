$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'envelope-observe.ps1')
function Check-Compile([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$tool=Join-Path $scope 'framework-001/toolchain'
$modules='C:/Users/Lenovo/go/pkg/mod'
function Test-CompileCase([string]$Scenario) {
    $target=Join-Path $scope ('envelope-compile-tests-'+[guid]::NewGuid().ToString('N')+'/envelope-observe-901')
    $plan=Get-EnvelopeObservationPlan $target $tool $modules
    New-Item -ItemType Directory -Path (Split-Path -Parent $target) -Force|Out-Null
    $inventoryFile=Join-Path (Split-Path -Parent $target) 'product-fixture.go'
    [IO.File]::WriteAllText($inventoryFile,'package fixture')
    $inventoryState=[pscustomobject]@{Calls=0}
    # Substitute only the slow full-tree inventory and external Go boundary.
    # Real planner, source binding, evidence parsers, completion gates, driver,
    # result persistence and streaming consumer run. This binary is never run.
    function Get-EnvelopeProductManifest {
        $inventoryState.Calls++
        if($inventoryState.Calls -eq 3) {
            if($Scenario -ceq 'final-binary'){[IO.File]::WriteAllText($plan.Binary,'changed-binary')}
            if($Scenario -ceq 'final-inventory'){[IO.File]::WriteAllText($inventoryFile,'package changed')}
            if($Scenario -ceq 'final-read'){throw 'Controlled final inventory failure'}
        }
        [pscustomobject]@{Path='product-fixture.go';SHA256=(Get-FileHash -LiteralPath $inventoryFile).Hash}
    }
    $calls=New-Object 'System.Collections.Generic.List[string]'
    function Invoke-RecordedProcess([string]$Executable,[string[]]$Arguments,[hashtable]$Environment,[string]$ReceiptDir,[int]$LimitSeconds) {
        $stage=Split-Path -Leaf $ReceiptDir;$calls.Add($stage)
        if($Executable -cne $plan.Executable -or $Environment.GOPROXY -cne 'off' -or $Environment.GOFLAGS -cne '-mod=readonly'){throw 'Wrong copied executable or offline contract'}
        $stderrText=''
        if($stage -ceq 'version' -and $LimitSeconds -eq 60 -and ($Arguments -join '|') -ceq 'version') {$text='go version go1.26.5 windows/amd64'}
        elseif($stage -ceq 'origin' -and $LimitSeconds -eq 60 -and ($Arguments -join '|') -ceq 'env|GOROOT|GOTOOLDIR|GOMOD') {$text=$plan.Toolchain+"`r`n"+(Join-Path $plan.Toolchain 'pkg/tool/windows_amd64')+"`r`n"+(Join-Path $plan.Project 'go.mod')}
        elseif($stage -ceq 'build' -and $LimitSeconds -eq 600 -and ($Arguments -join '|') -ceq ('test|-c|-x|-p=1|-count=1|-run=^$|-fuzz=^FuzzIssuerEnvelope$|-fuzztime=5s|-parallel=1|-timeout=30s|-o='+$plan.Binary+'|talenro.local/platform/internal/nodecontrol/nodeissuer')) {
            $text='';[IO.File]::WriteAllText($plan.Binary,'controlled-not-executable')
            $compiler=Join-Path $tool 'pkg/tool/windows_amd64/compile.exe'
            $stderrText='"'+$compiler+'" -p context -pack "'+(Join-Path $tool 'src/context/context.go')+'"'+"`r`n"+'"'+$compiler+'" -p internal/fuzz -pack "'+(Join-Path $tool 'src/internal/fuzz/fuzz.go')+'"'+"`r`n"+'"'+$compiler+'" -p talenro.local/platform/internal/nodecontrol/nodeissuer -pack "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/fuzz_test.go')+'" "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/der.go')+'" "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/signature.go')+'"'
            if($Scenario -ceq 'build-log'){$stderrText='missing compile evidence'}
        } else {throw 'Forbidden sample execution or wrong process arguments'}
        New-Item -ItemType Directory -Path $ReceiptDir -Force|Out-Null
        $stdout=Join-Path $ReceiptDir 'stdout.txt';$stderr=Join-Path $ReceiptDir 'stderr.txt'
        [IO.File]::WriteAllText($stdout,$text);[IO.File]::WriteAllText($stderr,$stderrText)
        $receipt=[pscustomobject]@{ExitCode=0;TimedOut=$false;RootExited=$true;CleanupVerified=$true;ActiveAfterCleanup=0;DescendantsRemaining=0;TotalProcesses=1;RootId=17;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;TerminationRequested=$false;ActiveBeforeTermination=0;MembersBeforeTermination=@();MembersAfterCleanup=@();MembersAtRootWait=@();NaturalObservationMs=0;ShutdownElapsedMs=0;ShutdownSchema=3;ShutdownStatePolicy='owned-final-signaled-exit-image-optional-v2';MemberEvidenceTruncated=$false;ShutdownMemberEvidence=@();StdoutPath=$stdout;StderrPath=$stderr}
        if($stage -ceq 'build' -and $Scenario -ceq 'build-exit'){$receipt.ExitCode=7}
        if($stage -ceq 'build' -and $Scenario -ceq 'build-gate'){$receipt.NaturalShutdownVerified=$false;$receipt.TerminationRequested=$true}
        return $receipt
    }
    $observed=New-Object 'System.Collections.Generic.List[object]';$failure=$null
    try {
        if($Scenario -ceq 'default'){Invoke-EnvelopeObservation $target $tool $modules|ForEach-Object {$observed.Add($_)}}
        else {Invoke-EnvelopeObservation $target $tool $modules -CompileOnly|ForEach-Object {$observed.Add($_)}}
    } catch {$failure=$_.Exception.Message}
    [pscustomobject]@{Root=$target;Failure=$failure;Calls=@($calls.ToArray());Observed=@($observed.ToArray())}
}
$success=Test-CompileCase 'success'
Check-Compile ($null -eq $success.Failure -and $success.Observed.Count -eq 1 -and $success.Observed[0].BuildPassed) 'compile-only succeeds after build without reaching sampling boundary'
Check-Compile (($success.Calls -join '|') -ceq 'version|origin|build') 'compile-only executes exactly the three approved stages'
$saved=Get-Content -LiteralPath (Join-Path $success.Root 'compile-result.json') -Raw|ConvertFrom-Json
$binary=Get-Content -LiteralPath (Join-Path $success.Root 'binary-receipt.json') -Raw|ConvertFrom-Json
Check-Compile ($saved.CompileOnly -and $saved.BuildPassed -and $saved.BuildExecuted -and -not $saved.SamplingExecuted -and -not $saved.Acceptance) 'compile result cannot imply sampling or acceptance'
Check-Compile ($saved.BinaryHash -ceq (Get-FileHash -LiteralPath (Join-Path $success.Root 'nodeissuer.test.exe')).Hash -and $saved.BinaryHash -ceq $binary.BinaryHash -and $saved.SourceIdentity -ceq $binary.SourceIdentity) 'compile result binds saved binary and copied-source receipt'
Check-Compile (-not (Test-Path -LiteralPath (Join-Path $success.Root 'execute')) -and -not (Test-Path -LiteralPath (Join-Path $success.Root 'sample-result.json')) -and (Test-Path -LiteralPath (Join-Path $success.Root 'after-product.json'))) 'compile-only retains final inventory without execution artifacts'
foreach($scenario in @('build-exit','build-gate','build-log','final-binary','final-inventory','final-read')) {
    $failed=Test-CompileCase $scenario
    Check-Compile ($null -ne $failed.Failure -and $failed.Observed.Count -eq 0 -and -not (Test-Path -LiteralPath (Join-Path $failed.Root 'compile-result.json'))) ('failed '+$scenario+' emits no premature success and saves no success result')
    $stopped=Get-Content -LiteralPath (Join-Path $failed.Root 'stopped.json') -Raw|ConvertFrom-Json
    $expectedStage=if($scenario.StartsWith('final-')){'finalize'}else{'build'}
    Check-Compile ($stopped.Stage -ceq $expectedStage -and $stopped.CompileOnly -and $stopped.BuildExecuted -and -not $stopped.SamplingExecuted -and -not $stopped.Acceptance -and ($failed.Calls -join '|') -ceq 'version|origin|build') ('failed '+$scenario+' stops without retry or sampling')
}
$default=Test-CompileCase 'default'
Check-Compile (($default.Calls -join '|') -ceq 'version|origin|build|execute' -and $default.Failure -ceq 'Wrong copied executable or offline contract' -and -not (Test-Path -LiteralPath (Join-Path $default.Root 'compile-result.json'))) 'default still reaches its sampling boundary rather than silently becoming compile-only'
$modeError=$null
try {Invoke-EnvelopeObservation '' '' '' -PreparationOnly -CompileOnly|Out-Null}catch{$modeError=$_.Exception.Message}
Check-Compile ($modeError -ceq 'Preparation-only and compile-only modes are mutually exclusive') 'conflicting modes fail before planning paths or launching commands'
