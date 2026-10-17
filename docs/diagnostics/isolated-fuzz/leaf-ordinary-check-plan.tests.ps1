$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'leaf-ordinary-check-fixture.ps1')
if(Test-Path -LiteralPath (Join-Path $PSScriptRoot 'leaf-ordinary-check.ps1')){. (Join-Path $PSScriptRoot 'leaf-ordinary-check.ps1')}
Check-LeafOrdinary ($null -ne (Get-Command Get-LeafOrdinaryCheckPlan -ErrorAction SilentlyContinue)) 'PlannerExists'
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
$f=New-LeafOrdinaryTestFiles (Join-Path $workspace ('.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/ordinary-unit-'+[guid]::NewGuid().ToString('N')))
function Invoke-MatrixNative {throw 'Forbidden planning launch'}
function Assert-UninstrumentedEnvelopeToolchain {throw 'Forbidden planner SDK inventory'}
function Get-EnvelopeProductManifest {throw 'Forbidden planner product inventory'}
$p=Get-LeafOrdinaryCheckPlan $f.Root $f.Toolchain $f.ReferenceReceipt $f.ModuleCache
Check-LeafOrdinary (-not(Test-Path -LiteralPath $f.Root)) 'PlanningLaunchesZeroAndDoesNotCreateRoot'
Check-LeafOrdinary ($p.Policy -ceq 'leaf-ordinary-check-v1' -and $p.VetMode -ceq 'default' -and $p.Package -ceq 'internal/nodecontrol/nodeleafbinding') 'FixedScope'
Check-LeafOrdinary ($p.BuildLimitSeconds -eq 600 -and $p.ExecuteLimitSeconds -eq 150) 'FixedBudgets'
Check-LeafOrdinary ($p.Acceptance -is [bool] -and -not $p.Acceptance -and $p.NoRetry -is [bool] -and $p.NoRetry) 'TypedFlags'
$build=@('test','-c','-x','-p=1','-count=1','-timeout=120s',('-o='+(Join-Path $f.Root 'package.test.exe')),'talenro.local/platform/internal/nodecontrol/nodeleafbinding')
Check-LeafOrdinary (($p.BuildArguments -join '|') -ceq ($build -join '|') -and ($p.Arguments -join '|') -ceq '-test.count=1|-test.timeout=120s|-test.v=true|-test.run=.') 'FixedArguments'
$envWant=@{GOROOT=$f.Toolchain;GOCACHE=(Join-Path $f.Root 'go-cache');GOMODCACHE=$f.ModuleCache;TEMP=(Join-Path $f.Root 'tmp');TMP=(Join-Path $f.Root 'tmp');GOTMPDIR=(Join-Path $f.Root 'tmp');GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='on';GOWORK='off';GOFLAGS='-mod=readonly';GOOS='windows';GOARCH='amd64';TALENRO_FUZZ_DIAG_MODE='';TALENRO_FUZZ_FIXTURE='';GODEBUG=''}
Check-LeafOrdinary ($p.Environment.Count -eq 20) 'FixedEnvironment20'
foreach($key in $envWant.Keys){Check-LeafOrdinary ($p.Environment[$key] -ceq $envWant[$key]) ('Environment '+$key)}
Check-LeafOrdinary ($p.SourceBinding.Count -eq 16 -and $p.InputBinding.Count -eq 4) 'BindingsPresent'
foreach($field in @('Policy','Package','VetMode','BuildLimitSeconds','Arguments','BuildArguments','Acceptance','Environment','SourceBinding','InputBinding')){
    $q=$p|ConvertTo-Json -Depth 20|ConvertFrom-Json
    switch($field){
        'Environment' {$q.Environment.GOFLAGS='-mod=mod'}
        'SourceBinding' {$q.SourceBinding[0].SHA256='A'*64}
        'InputBinding' {$q.InputBinding[0].SHA256='A'*64}
        'Acceptance' {$q.Acceptance='false'}
        'Arguments' {$q.Arguments=@('-test.run=^$')}
        'BuildArguments' {$q.BuildArguments[1]='-vet=off'}
        'BuildLimitSeconds' {$q.BuildLimitSeconds=601}
        default {$q.$field='altered'}
    }
    Reject-LeafOrdinary {Assert-LeafOrdinaryCheckPlan $q $true} ('TamperedPlanRejected '+$field)
}
Reject-LeafOrdinary {Get-LeafOrdinaryCheckPlan (Join-Path (Split-Path $f.Root) 'matrix-check-901') $f.Toolchain $f.ReferenceReceipt $f.ModuleCache} 'WrongRootRejected'
Reject-LeafOrdinary {Get-LeafOrdinaryCheckPlan $f.Root $f.Toolchain $f.ReferenceReceipt $f.Root} 'OverlapRejected'
Reject-LeafOrdinary {Get-LeafOrdinaryCheckPlan $f.Root $f.Toolchain $f.Toolchain $f.ModuleCache} 'UnsafeInputRejected'
New-Item -ItemType Directory -Path $f.Root|Out-Null
Reject-LeafOrdinary {Get-LeafOrdinaryCheckPlan $f.Root $f.Toolchain $f.ReferenceReceipt $f.ModuleCache} 'ExistingRootRejected'
Assert-LeafOrdinaryCheckPlan $p $false
Check-LeafOrdinary $true 'ExistingRootValidationWithoutPlanner'
[IO.File]::AppendAllText((Join-Path $f.Toolchain 'bin/go.exe'),'changed')
Reject-LeafOrdinary {Assert-LeafOrdinaryCheckPlan $p $false} 'InputDriftRejected'
