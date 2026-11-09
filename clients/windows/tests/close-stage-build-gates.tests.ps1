#requires -Version 7.0
$ErrorActionPreference = 'Stop'
$project = Join-Path $PSScriptRoot '../Talenro.Demo.App/Talenro.Demo.App.csproj'
function Check([bool]$condition, [string]$message) { if (-not $condition) { throw $message } }
$rejectedModes = 0
foreach ($arguments in @(
    @('-p:CloseStageVerificationMode=layuot','-p:EnableLayoutDiagnostics=true','-p:EnableCloseStageRecording=false'),
    @('-p:CloseStageVerificationMode=ordinary','-p:EnableLayoutDiagnostics=true','-p:EnableCloseStageRecording=false'),
    @('-p:CloseStageVerificationMode=layout','-p:EnableLayoutDiagnostics=false','-p:EnableCloseStageRecording=false'),
    @('-p:CloseStageVerificationMode=layout','-p:EnableLayoutDiagnostics=true','-p:EnableCloseStageRecording=true'))) {
    $output = & dotnet msbuild $project -nologo -target:ValidateCloseStageConfiguration -p:Configuration=Debug @arguments 2>&1
    if ($LASTEXITCODE -ne 0 -and ($output -join "`n") -match 'CLOSESTAGE002') { $rejectedModes++ }
}
Check ($rejectedModes -eq 4) "Unsafe verification modes rejected: $rejectedModes/4; must fail before build/restore."
foreach ($arguments in @(
    @('-p:Configuration=Release','-p:EnableLayoutDiagnostics=true','-p:EnableCloseStageRecording=true'),
    @('-p:Configuration=Debug','-p:EnableLayoutDiagnostics=false','-p:EnableCloseStageRecording=true'),
    @('-p:Configuration=Debug','-p:EnableLayoutDiagnostics=true','-p:EnableCloseStageRecording=invalid'))) {
    $output = & dotnet msbuild $project -nologo -target:ValidateCloseStageConfiguration @arguments 2>&1
    Check ($LASTEXITCODE -ne 0 -and ($output -join "`n") -match 'CLOSESTAGE001') 'Invalid flags must fail with CLOSESTAGE001.'
}
foreach ($mode in @('ordinary','layout','trace')) {
    $layout = if ($mode -eq 'ordinary') { 'false' } else { 'true' }
    $trace = if ($mode -eq 'trace') { 'true' } else { 'false' }
    $output = & dotnet msbuild $project -nologo -getProperty:BaseIntermediateOutputPath,BaseOutputPath,MSBuildProjectExtensionsPath,DefineConstants -p:Configuration=Debug "-p:EnableLayoutDiagnostics=$layout" "-p:EnableCloseStageRecording=$trace"
    Check ($LASTEXITCODE -eq 0) 'Property evaluation failed.'
    $values = ($output -join "`n" | ConvertFrom-Json).Properties
    $expected = switch ($mode) { 'ordinary' { 'obj/' } 'layout' { 'obj/layout-diagnostics/' } 'trace' { 'obj/close-stage-recording/' } }
    Check ($values.BaseIntermediateOutputPath.Replace('\','/') -eq $expected) "Wrong intermediate isolation for $mode."
    Check ($values.BaseOutputPath.Replace('\','/') -eq $expected.Replace('obj/','bin/')) "Wrong output isolation for $mode."
    if ($mode -eq 'trace') {
        Check ($values.DefineConstants -match 'CLOSE_STAGE_RECORDING') 'Trace define missing.'
        Check ($values.MSBuildProjectExtensionsPath.Replace('\','/').EndsWith('/obj/close-stage-recording/')) 'Trace assets path not isolated.'
    } else { Check ($values.DefineConstants -notmatch 'CLOSE_STAGE_RECORDING') 'Trace define leaked.' }
}
foreach ($mode in @('ordinary','layout')) {
    $layout = if ($mode -eq 'layout') { 'true' } else { 'false' }
    $output = & dotnet msbuild $project -nologo -getProperty:BaseIntermediateOutputPath,MSBuildProjectExtensionsPath,BaseOutputPath "-p:CloseStageVerificationMode=$mode" "-p:EnableLayoutDiagnostics=$layout" -p:EnableCloseStageRecording=false
    Check ($LASTEXITCODE -eq 0) 'Verification path evaluation failed.'
    $values = ($output -join "`n" | ConvertFrom-Json).Properties
    Check ($values.BaseIntermediateOutputPath -eq "obj/close-stage-verification/$mode/") 'Verification intermediate not isolated.'
    Check ([IO.Path]::IsPathFullyQualified($values.MSBuildProjectExtensionsPath)) 'NuGet imports need an absolute extensions path.'
    Check ($values.BaseOutputPath -eq "bin/close-stage-verification/$mode/") 'Verification output not isolated.'
}
'PASS: invalid flags rejected and all output/define modes isolated (no application executed).'
