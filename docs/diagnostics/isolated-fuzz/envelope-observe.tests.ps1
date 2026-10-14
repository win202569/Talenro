$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'envelope-observe.ps1')
function Check-EnvelopePlan([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Reject-EnvelopePlan([scriptblock]$Action,[string]$Name) {$refused=$false;try {& $Action|Out-Null}catch {$refused=$true};Check-EnvelopePlan $refused $Name}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$testRoot=Join-Path $scope ('envelope-plan-tests-'+[guid]::NewGuid().ToString('N'))
$target=Join-Path $testRoot 'envelope-observe-901'
$tool=Join-Path $scope 'framework-001/toolchain'
$moduleCache='C:/Users/Lenovo/go/pkg/mod'
$plan=Get-EnvelopeObservationPlan $target $tool $moduleCache
Check-EnvelopePlan ($plan.Package -ceq 'talenro.local/platform/internal/nodecontrol/nodeissuer' -and $plan.Mode -ceq 'observe') 'plan selects original Envelope package and observation only'
Check-EnvelopePlan (($plan.ExecuteArguments -join '|') -ceq ('-test.run=^$|-test.fuzz=^FuzzIssuerEnvelope$|-test.fuzztime=5s|-test.parallel=1|-test.timeout=30s|-test.count=1|-test.fuzzcachedir='+(Join-Path $target 'fuzz-cache'))) 'actual entry point retains approved fuzz flags and fresh cache'
Check-EnvelopePlan ($plan.BuildLimitSeconds -eq 600 -and $plan.ExecuteLimitSeconds -eq 60 -and $plan.Environment.GOFLAGS -ceq '-mod=readonly' -and $plan.Environment.GOPROXY -ceq 'off' -and $plan.Environment.GOTOOLCHAIN -ceq 'local' -and $plan.Environment.TALENRO_FUZZ_DIAG_MODE -ceq 'observe') 'build and execution budgets plus offline readonly module mode fixed'
Check-EnvelopePlan (-not (Test-Path -LiteralPath $testRoot)) 'planning creates no cache or receipt and launches no process'
$identity=Assert-EnvelopeRunProvenance $plan
Check-EnvelopePlan ($identity -ceq 'E3949C0C3EC7F5D7BB8D239CEF2BBD2B753EC162070C8FB289DC8A60799ED306') 'verified copied executable and full helper binding accepted before launch'
Reject-EnvelopePlan {Get-EnvelopeObservationPlan (Join-Path $workspace 'outside') $tool $moduleCache} 'outside owned receipt root refused'
Reject-EnvelopePlan {Get-EnvelopeObservationPlan (Join-Path $testRoot 'pair-observe-901') $tool $moduleCache} 'Pair cannot masquerade as Envelope sampling'
Reject-EnvelopePlan {Get-EnvelopeObservationPlan $target 'C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64' $moduleCache} 'official toolchain target refused'
Reject-EnvelopePlan {Get-EnvelopeObservationPlan $target $tool (Join-Path $testRoot 'missing-cache')} 'missing dependency cache stops without download'
New-Item -ItemType Directory -Path $target -Force|Out-Null
Reject-EnvelopePlan {Get-EnvelopeObservationPlan $target $tool $moduleCache} 'existing identity refused without automatic retry'
Reject-EnvelopePlan {Invoke-EnvelopeObservation $target $tool $moduleCache} 'execution refuses existing identity before invoking Go'
$badTool=Join-Path $testRoot 'bad-toolchain'
New-Item -ItemType Directory -Path (Join-Path $badTool 'bin') -Force|Out-Null
[IO.File]::WriteAllText((Join-Path $badTool 'bin/go.exe'),'not-a-toolchain')
$badTarget=Join-Path $testRoot 'envelope-observe-902'
Reject-EnvelopePlan {Invoke-EnvelopeObservation $badTarget $badTool $moduleCache} 'unverified executable cannot launch Go'
Check-EnvelopePlan (-not (Test-Path -LiteralPath $badTarget)) 'failed provenance creates no preparation or receipt'
$fakeTool='D:\owned copy\toolchain';$fakeProject='D:\owned project'
$log='"D:\\owned copy\\toolchain\\pkg\\tool\\windows_amd64\\compile.exe" -p talenro.local/platform/internal/nodecontrol/nodeissuer -pack "D:\\owned project\\internal\\nodecontrol\\nodeissuer\\fuzz_test.go" "D:\\owned project\\internal\\nodecontrol\\nodeissuer\\der.go" "D:\\owned project\\internal\\nodecontrol\\nodeissuer\\signature.go"'
Assert-EnvelopeCompileEvidence $log $fakeTool $fakeProject
Check-EnvelopePlan $true 'exact copied compiler and original test plus parsers occur in one package command'
Reject-EnvelopePlan {Assert-EnvelopeCompileEvidence ($log.Replace('compile.exe','fake.exe')) $fakeTool $fakeProject} 'text without real copied compiler refused'
Reject-EnvelopePlan {Assert-EnvelopeCompileEvidence ($log.Replace('fuzz_test.go','fuzz_test.go.extra')) $fakeTool $fakeProject} 'source suffix cannot substitute for original test'
Reject-EnvelopePlan {Assert-EnvelopeCompileEvidence ($log.Replace('nodeissuer -pack','other -pack')) $fakeTool $fakeProject} 'source in wrong compile package refused'
$before=@([pscustomobject]@{Path='a.go';SHA256='ABC'})
Assert-EnvelopeProductUnchanged $before @([pscustomobject]@{Path='a.go';SHA256='ABC'})
Check-EnvelopePlan $true 'unchanged literal source inventory accepted'
Reject-EnvelopePlan {Assert-EnvelopeProductUnchanged $before @([pscustomobject]@{Path='a.go';SHA256='DEF'})} 'changed product hash refused'
Reject-EnvelopePlan {Assert-EnvelopeProductUnchanged $before @([pscustomobject]@{Path='b.go';SHA256='ABC'})} 'changed source membership refused'
$manifest=@(Get-EnvelopeProductManifest)
Check-EnvelopePlan (@($manifest|Where-Object {$_.Path -ceq 'internal/nodecontrol/nodeissuer/fuzz_test.go' -and $_.SHA256 -ceq '318CF8DE04BA5D95E71E625ABF6AAED8DFD54B854A9A925B7A3004D4132D989E'}).Count -eq 1 -and @($manifest|Where-Object {$_.Path -ceq 'go.mod'}).Count -eq 1 -and @($manifest|Where-Object {$_.Path -ceq 'go.sum'}).Count -eq 1) 'actual product manifest covers pinned original callback and module files'
