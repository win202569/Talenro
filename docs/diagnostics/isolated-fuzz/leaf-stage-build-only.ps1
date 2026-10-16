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
