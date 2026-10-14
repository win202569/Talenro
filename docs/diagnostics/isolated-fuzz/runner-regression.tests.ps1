$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'run.ps1')
function Assert-Regression([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Reject-Regression([scriptblock]$Action,[string]$Name) {$caught=$false;try{& $Action|Out-Null}catch{$caught=$true};Assert-Regression $caught $Name}
Assert-Regression ([IsolatedFuzzProcess]::CountDescendants(41,[uint64[]]@(41)) -eq 0) 'known root is not a descendant'
Assert-Regression ([IsolatedFuzzProcess]::CountDescendants(41,[uint64[]]@(41,42)) -eq 1) 'real member remains a descendant'
Assert-Regression ([IsolatedFuzzProcess]::CountDescendants(41,[uint64[]]@(42)) -eq 1) 'root absence cannot hide a descendant'
$tool='D:\owned copy\toolchain'
$log='"D:\\owned copy\\toolchain\\pkg\\tool\\windows_amd64\\compile.exe" -p context -pack "D:\\owned copy\\toolchain\\src\\context\\context.go"' + "`n" + '"D:\\owned copy\\toolchain\\pkg\\tool\\windows_amd64\\compile.exe" -p internal/fuzz -pack "D:\\owned copy\\toolchain\\src\\internal\\fuzz\\fuzz.go"'
Assert-CopiedCompileEvidence $log $tool
Assert-Regression $true 'escaped compiler and source tokens accepted'
Reject-Regression {Assert-CopiedCompileEvidence ($log.Replace('owned copy','other copy')) $tool} 'wrong compiler root refused'
Reject-Regression {Assert-CopiedCompileEvidence ($log.Replace('fuzz.go','fuzz.go.extra')) $tool} 'source suffix cannot masquerade as exact token'
Reject-Regression {Assert-CopiedCompileEvidence ($log.Replace('-p context','-p other')) $tool} 'source must occur in correct package compile command'
Reject-Regression {Assert-CopiedCompileEvidence ($log.Replace('compile.exe','fake.exe')) $tool} 'source text without copied compiler refused'
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$root=Join-Path $workspace ('.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/attempt-tests-'+[guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path (Join-Path $root 'arms/baseline') -Force | Out-Null
$paths=Get-DiagnosticAttemptPaths $root 'baseline' 'retry-001'
Assert-Regression ($paths.ArmDir -eq (Join-Path $root 'arms/baseline-retry-001') -and $paths.BuildDir -eq (Join-Path $root 'binaries/plain-retry-001')) 'explicit new baseline identity preserves old arm and build'
New-Item -ItemType Directory -Path $paths.ArmDir | Out-Null
Reject-Regression {Get-DiagnosticAttemptPaths $root 'baseline' 'retry-001'} 'duplicate retry identity refused'
Reject-Regression {Get-DiagnosticAttemptPaths $root 'observe' 'retry-002'} 'baseline approval cannot retry another arm'
Reject-Regression {Get-DiagnosticAttemptPaths $root 'baseline' '../escape'} 'attempt identity cannot escape'
New-Item -ItemType Directory -Path (Join-Path $paths.ArmDir 'execute') | Out-Null
Reject-Regression {Get-DiagnosticAttemptPaths $root 'baseline' 'retry-002'} 'completed retry cannot authorize another fuzz run'
New-Item -ItemType Directory -Path (Join-Path $root 'arms/observe') | Out-Null
$paths=Get-DiagnosticAttemptPaths $root 'observe' '' 'prep-002'
Assert-Regression ($paths.ArmDir -eq (Join-Path $root 'arms/observe-prep-002') -and $paths.BuildDir -eq (Join-Path $root 'binaries/diagnostic')) 'new preparation preserves failed observe and shares only bound diagnostic binary'
Reject-Regression {Get-DiagnosticAttemptPaths $root 'baseline' '' 'prep-002'} 'diagnostic preparation cannot repeat baseline'
Reject-Regression {Get-DiagnosticAttemptPaths $root 'observe' '' '../escape'} 'preparation identity cannot escape'
Reject-Regression {Get-DiagnosticAttemptPaths $root 'cancel' '' 'prep-002'} 'preparation requires an existing failed preparation'
New-Item -ItemType Directory -Path $paths.ArmDir | Out-Null
Reject-Regression {Get-DiagnosticAttemptPaths $root 'observe' '' 'prep-002'} 'duplicate preparation identity refused'
New-Item -ItemType Directory -Path (Join-Path $paths.ArmDir 'execute') | Out-Null
Reject-Regression {Get-DiagnosticAttemptPaths $root 'observe' '' 'prep-003'} 'executed preparation forbids further fuzz attempt'
New-Item -ItemType Directory -Path (Join-Path $root 'arms/observe/execute') | Out-Null
Reject-Regression {Get-DiagnosticAttemptPaths $root 'observe' '' 'prep-004'} 'executed original arm forbids preparation retry'
Reject-Regression {Get-DiagnosticAttemptPaths $root '../escape'} 'arm identity cannot escape'
