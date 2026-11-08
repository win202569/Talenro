#requires -Version 7.0
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$modulePath = Join-Path $PSScriptRoot 'process-observation.psm1'
if (-not (Test-Path -LiteralPath $modulePath)) { throw 'ASSERT: observation module is missing' }
$module = Import-Module $modulePath -Force -PassThru
$script:caseCount = 0
function Assert-Observation($Condition, [string]$Message) {
    if (-not $Condition) { throw "ASSERT: $Message" }
}
function Pass-Case([string]$Name) { $script:caseCount++; "PASS: $Name" }
function New-Readers {
    [ordered]@{
        ProcessId = { [int]123 }
        StartTime = { [datetime]::new(2026, 1, 2, 3, 4, 5, [DateTimeKind]::Utc) }
        WaitSignaled = { $true }
        HasExited = { $true }
        ExitCode = { [int]0 }
        ExitTime = { [datetime]::new(2026, 1, 2, 3, 5, 5, [DateTimeKind]::Utc) }
    }
}
$names = @('ProcessId','StartTime','WaitSignaled','HasExited','ExitCode','ExitTime')
foreach ($code in @(0, 7, 259)) {
    $readers = New-Readers
    $readers.ExitCode = { $code }.GetNewClosure()
    $record = Get-TalenroSyntheticProcessObservation -Readers $readers
    Assert-Observation ($record.Fields.ExitCode.Status -ceq 'ok' -and $record.Fields.ExitCode.Value -is [int] -and $record.Fields.ExitCode.Value -eq $code) "exit code $code was coerced/lost"
    Assert-Observation ($record.Mode -ceq 'synthetic' -and $record.SchemaVersion -eq 1) 'synthetic identity'
    Assert-Observation (($record.PSObject.Properties.Name -join ',') -ceq 'SchemaVersion,Mode,ObservationStartedUtc,ObservationEndedUtc,Fields') 'unexpected verdict or snapshot property'
    Pass-Case "preserve exit code $code"
}
$readers = New-Readers
$readers.WaitSignaled = { $false }; $readers.HasExited = { $false }
$readers.ExitCode = { throw [InvalidOperationException]::new('private marker') }
$readers.ExitTime = { throw [InvalidOperationException]::new('private marker') }
$record = Get-TalenroSyntheticProcessObservation $readers
Assert-Observation ($record.Fields.WaitSignaled.Value -is [bool] -and -not $record.Fields.WaitSignaled.Value -and -not $record.Fields.HasExited.Value) 'false must remain Boolean'
foreach ($field in @('ExitCode','ExitTime')) {
    Assert-Observation ($record.Fields.$field.Status -ceq 'error' -and $null -eq $record.Fields.$field.Value -and $record.Fields.$field.ExceptionType) 'unavailable field fabricated value'
}
Assert-Observation (($record | ConvertTo-Json -Depth 8) -notmatch 'private marker') 'error leaked message'
Pass-Case 'unavailable exits differ from zero'
$readers = New-Readers; $readers.WaitSignaled = { $false }
$record = Get-TalenroSyntheticProcessObservation $readers
Assert-Observation (-not $record.Fields.WaitSignaled.Value -and $record.Fields.HasExited.Value) 'disagreement changed'
Pass-Case 'preserve disagreement'
foreach ($failing in $names) {
    $readers = New-Readers; $calls = [Collections.Generic.List[string]]::new()
    foreach ($name in $names) {
        $original = $readers[$name]
        $readers[$name] = {
            $calls.Add($name)
            if ($name -eq $failing) { throw [InvalidOperationException]::new('sensitive text') }
            & $original
        }.GetNewClosure()
    }
    $record = Get-TalenroSyntheticProcessObservation $readers
    Assert-Observation (($calls -join ',') -ceq ($names -join ',')) 'reader ordering/count changed'
    Assert-Observation ($record.Fields.$failing.Status -ceq 'error' -and $null -eq $record.Fields.$failing.Value) "failed $failing lost error"
    foreach ($name in $names | Where-Object { $_ -ne $failing }) {
        Assert-Observation ($record.Fields.$name.Status -ceq 'ok') "error suppressed $name"
    }
    Pass-Case "isolated $failing exception"
}
$readers = New-Readers; $readers.ExitCode = { [int]0; throw 'partial-private' }
$record = Get-TalenroSyntheticProcessObservation $readers
Assert-Observation ($record.Fields.ExitCode.Status -ceq 'error' -and $null -eq $record.Fields.ExitCode.Value -and $record.Fields.ExitTime.Status -ceq 'ok') 'partial output accepted'
Pass-Case 'discard output before error'
$invalidReaders = [ordered]@{
    'no output' = {}; 'null' = { $null }; 'string zero' = { '0' }
    'Int64 zero' = { [long]0 }; 'empty array' = { ,@() }
    'singleton array' = { ,@(0) }; 'multiple array' = { ,@(0,1) }
    'multiple outputs' = { 0; 1 }
}
foreach ($label in $invalidReaders.Keys) {
    $readers = New-Readers; $readers.ExitCode = $invalidReaders[$label]
    $record = Get-TalenroSyntheticProcessObservation $readers
    Assert-Observation ($record.Fields.ExitCode.Status -ceq 'invalid' -and $null -eq $record.Fields.ExitCode.Value -and $null -eq $record.Fields.ExitCode.ExceptionType) "invalid $label accepted"
    Pass-Case "reject $label"
}
foreach ($field in $names) {
    $readers = New-Readers; $readers[$field] = { 'wrong type' }
    $record = Get-TalenroSyntheticProcessObservation $readers
    Assert-Observation ($record.Fields.$field.Status -ceq 'invalid') "type check absent for $field"
    Pass-Case "validate $field type"
}
$readers = New-Readers; $readers.ProcessId = { [int]0 }
Assert-Observation ((Get-TalenroSyntheticProcessObservation $readers).Fields.ProcessId.Status -ceq 'invalid') 'PID0 accepted'
Pass-Case 'reject zero PID'
foreach ($variant in @('missing','extra','not scriptblock')) {
    $readers = New-Readers; $calls = [Collections.Generic.List[string]]::new()
    $readers.ProcessId = { $calls.Add('unexpected'); 123 }.GetNewClosure()
    switch ($variant) {
        'missing' { $readers.Remove('ExitTime') }
        'extra' { $readers.Add('Other', { 1 }) }
        'not scriptblock' { $readers.ExitTime = 'not code' }
    }
    $threw = $false
    try { $null = Get-TalenroSyntheticProcessObservation $readers } catch { $threw = $true }
    Assert-Observation ($threw -and $calls.Count -eq 0) "invalid dictionary $variant executed reader"
    Pass-Case "reject reader dictionary $variant"
}
$readers = New-Readers; $readers.WaitSignaled = { $false }; $readers.ExitTime = { throw 'private' }
$record = Get-TalenroSyntheticProcessObservation $readers
$roundtrip = $record | ConvertTo-Json -Depth 8 | ConvertFrom-Json
Assert-Observation ($roundtrip.Mode -ceq 'synthetic' -and $roundtrip.Fields.ExitCode.Value -eq 0 -and $roundtrip.Fields.ExitCode.Value -is [long] -and $null -eq $roundtrip.Fields.ExitTime.Value -and $roundtrip.Fields.WaitSignaled.Value -is [bool] -and -not $roundtrip.Fields.WaitSignaled.Value) 'JSON changed null/false/zero/mode'
Assert-Observation ($record.Fields.StartTime.Value -ceq '2026-01-02T03:04:05.0000000Z') 'time not UTC roundtrip'
Assert-Observation ($record.ObservationStartedUtc.EndsWith('Z') -and $record.ObservationEndedUtc.EndsWith('Z') -and [datetime]$record.ObservationStartedUtc -le [datetime]$record.ObservationEndedUtc) 'invalid observation interval'
Pass-Case 'JSON and UTC preservation'
# Static declaration checks are not a runtime/safety proof of the live adapter.
$tokens = $null; $parseErrors = $null
$ast = [Management.Automation.Language.Parser]::ParseFile($modulePath, [ref]$tokens, [ref]$parseErrors)
Assert-Observation ($parseErrors.Count -eq 0) 'module parse errors'
foreach ($statement in $ast.EndBlock.Statements) {
    Assert-Observation ($statement -is [Management.Automation.Language.FunctionDefinitionAst] -or ($statement -is [Management.Automation.Language.PipelineAst] -and $statement.Extent.Text -match '^Export-ModuleMember\b')) 'import-time executable statement'
}
$exports = @($module.ExportedFunctions.Keys | Sort-Object)
Assert-Observation (($exports -join ',') -ceq 'Get-TalenroProcessObservation,Get-TalenroSyntheticProcessObservation') 'export surface changed'
$live = $ast.Find({ param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-TalenroProcessObservation' }, $true)
$invocations = @($live.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true))
foreach ($getter in @('get_Id','get_StartTime','get_HasExited','get_ExitCode','get_ExitTime')) {
    Assert-Observation (@($invocations | Where-Object { $_.Member.Value -eq $getter }).Count -eq 1) "live getter missing/duplicated: $getter"
}
$wait = @($invocations | Where-Object { $_.Member.Value -eq 'WaitForExit' })
Assert-Observation ($wait.Count -eq 1 -and $wait[0].Arguments.Count -eq 1 -and $wait[0].Arguments[0].Extent.Text -ceq '0') 'live wait must be nonblocking'
Assert-Observation ($ast.Extent.Text -notmatch '(?i)Start-Process|Get-Process|GetProcessById|GetProcesses|\.Start\(|\.Kill\(|\.Dispose\(|DllImport|Add-Type|Invoke-Expression|Out-File|Set-Content|WriteAll|Start-Sleep|WaitForExit\(-1\)') 'forbidden operation in module'
Pass-Case 'static live boundary and import surface'
"SYNTHETIC_ONLY_PASS: $script:caseCount cases; live adapter NOT_RUN"
