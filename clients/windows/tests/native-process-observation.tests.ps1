#requires -Version 7.0
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$script:count = 0
function Check($ok, [string]$name) {
    if (-not $ok) { throw "ASSERT: $name" }
    $script:count++; "PASS: $name"
}
function Fixtures {
    [ordered]@{
        Wait = { @{ RawResult=[uint32]258; LastError=5 } }
        ExitCode = { @{ ApiSucceeded=$true; ExitCode=[uint32]0; LastError=5 } }
        Times = { @{ ApiSucceeded=$true; LastError=5; Times=@{
            Creation=@{High=[uint32]0;Low=[uint32]0}; Exit=@{High=[uint32]0;Low=[uint32]0}
            Kernel=@{High=[uint32]0;Low=[uint32]0}; User=@{High=[uint32]0;Low=[uint32]0}
        } } }
    }
}
function Observe($r) { Get-TalenroSyntheticNativeProcessObservation -Readers $r -CorrelationId 'fixture' }
Check ($null -eq ('Talenro.Diagnostics.NativeObservationV1.NativeProbe' -as [type])) 'fresh host has no helper'
$path = Join-Path $PSScriptRoot 'native-process-observation.psm1'
if (-not (Test-Path $path)) { throw 'ASSERT: native recorder module missing' }
$module = Import-Module $path -Force -PassThru
foreach ($row in @(@([uint32]0,'ok','signaled'),@([uint32]258,'ok','timeout'),@([uint32]::MaxValue,'api-error','failed'),@([uint32]128,'invalid','unexpected'))) {
    $raw=$row[0]; $r=Fixtures; $r.Wait={ @{RawResult=$raw;LastError=5} }.GetNewClosure()
    $c=(Observe $r).Calls.Wait
    Check ($c.RawResult -is [uint32] -and $c.RawResult -eq $raw -and $c.Status -ceq $row[1] -and $c.WaitState -ceq $row[2]) "wait $raw"
    Check (($raw -eq [uint32]::MaxValue -and $c.LastError -eq 5) -or ($raw -ne [uint32]::MaxValue -and $null -eq $c.LastError)) "wait error $raw"
}
foreach ($code in @([uint32]0,[uint32]7,[uint32]259,[uint32]2147483648,[uint32]::MaxValue)) {
    $r=Fixtures; $r.ExitCode={ @{ApiSucceeded=$true;ExitCode=$code;LastError=5} }.GetNewClosure()
    $c=(Observe $r).Calls.ExitCode
    Check ($c.Status -ceq 'ok' -and $c.ExitCode -is [uint32] -and $c.ExitCode -eq $code -and $null -eq $c.LastError) "exit $code"
}
foreach ($name in @('ExitCode','Times')) {
    $r=Fixtures; $r[$name]={ @{ApiSucceeded=$false;ExitCode=0;Times=0;LastError=5} }
    $c=(Observe $r).Calls.$name
    Check ($c.Status -ceq 'api-error' -and $c.ApiSucceeded -eq $false -and $null -eq $c.ExitCode -and $null -eq $c.Times -and $c.LastError -eq 5) "failed $name discards buffers"
}
$s=Observe (Fixtures)
Check ($s.Calls.Times.Times.Exit.High -eq 0 -and $s.Calls.Times.Times.Exit.Low -eq 0 -and $null -eq $s.Calls.Times.LastError) 'zero time is raw data'
$r=Fixtures; $r.Times={ @{ApiSucceeded=$true;LastError=$null;Times=@{
    Creation=@{High=[uint32]::MaxValue;Low=[uint32]::MaxValue}; Exit=@{High=[uint32]::MaxValue;Low=[uint32]::MaxValue}
    Kernel=@{High=[uint32]::MaxValue;Low=[uint32]::MaxValue}; User=@{High=[uint32]::MaxValue;Low=[uint32]::MaxValue}
}} }
$s=Observe $r; $json=$s | ConvertTo-Json -Depth 12; $round=$json | ConvertFrom-Json
Check ($s.Calls.Times.Status -ceq 'ok' -and $round.Calls.Times.Times.Exit.High -eq 4294967295 -and $round.Calls.Times.Times.Exit.Low -eq 4294967295 -and $null -eq $round.Calls.Times.LastError) 'FILETIME JSON precision'
foreach ($target in @('Wait','ExitCode','Times')) {
    $r=Fixtures; $log=[Collections.Generic.List[string]]::new()
    foreach ($name in @('Wait','ExitCode','Times')) {
        $original=$r[$name]; $r[$name]={ $log.Add($name); if ($name -ceq $target) { 'partial'; throw 'private marker' }; & $original }.GetNewClosure()
    }
    $s=Observe $r
    Check (($log -join ',') -ceq 'Wait,ExitCode,Times' -and $s.Calls.$target.Status -ceq 'exception' -and $s.Calls.$target.ExceptionType -and $null -eq $s.Calls.$target.RawResult) "isolated $target exception"
    Check (($s | ConvertTo-Json -Depth 12) -notmatch 'private marker|partial') "no $target leaked payload"
}
$badReaders=@(
    { }, { $null }, { '0' }, { @{RawResult=0;LastError=5} },
    { @{RawResult=[uint32]0}; @{RawResult=[uint32]0} },
    { ,@(@{RawResult=[uint32]0;LastError=$null}) },
    { @{RawResult=[uint32]::MaxValue;LastError='5'} }
)
foreach ($reader in $badReaders) {
    $r=Fixtures; $r.Wait=$reader; $s=Observe $r
    Check ($s.Calls.Wait.Status -ceq 'invalid' -and $null -eq $s.Calls.Wait.RawResult -and $s.Calls.Times.Status -ceq 'ok') "invalid wait reader $reader"
}
foreach ($payload in @(@{ApiSucceeded=1;ExitCode=[uint32]0;LastError=$null},@{ApiSucceeded=$true;ExitCode=0;LastError=$null},@{ApiSucceeded=$false;LastError=$null})) {
    $r=Fixtures; $r.ExitCode={ $payload }.GetNewClosure()
    Check ((Observe $r).Calls.ExitCode.Status -ceq 'invalid') 'invalid exit payload'
}
foreach ($payload in @(@{ApiSucceeded=$true;Times=@{};LastError=$null},@{ApiSucceeded=$true;Times=@{Creation=@{High=0;Low=0}};LastError=$null})) {
    $r=Fixtures; $r.Times={ $payload }.GetNewClosure()
    Check ((Observe $r).Calls.Times.Status -ceq 'invalid') 'invalid time payload'
}
foreach ($kind in @('missing','extra','not-scriptblock')) {
    $r=Fixtures; $log=[Collections.Generic.List[string]]::new(); $r.Wait={ $log.Add('called') }.GetNewClosure()
    switch ($kind) { missing {$r.Remove('Times')} extra {$r.Other={}} not-scriptblock {$r.Times=0} }
    $threw=$false; try { Observe $r | Out-Null } catch {$threw=$true}
    Check ($threw -and $log.Count -eq 0) "reject dictionary $kind before calls"
}
foreach ($id in @('',('x'*129))) {
    $r=Fixtures; $log=[Collections.Generic.List[string]]::new(); $r.Wait={ $log.Add('called') }.GetNewClosure()
    $threw=$false; try { Get-TalenroSyntheticNativeProcessObservation -Readers $r -CorrelationId $id | Out-Null } catch {$threw=$true}
    Check ($threw -and $log.Count -eq 0) 'correlation validation before reads'
}
$s=Observe (Fixtures)
Check ($s.Mode -ceq 'synthetic' -and $s.SchemaVersion -eq 1 -and $s.CorrelationId -ceq 'fixture' -and $null -eq $s.SnapshotError) 'snapshot identity'
Check (($s.PSObject.Properties.Name -join ',') -ceq 'SchemaVersion,Mode,CorrelationId,ObservationStartedUtc,ObservationEndedUtc,SnapshotError,Calls') 'no verdict'
$previous=[datetime]::Parse($s.ObservationStartedUtc)
foreach ($name in @('Wait','ExitCode','Times')) {
    $c=$s.Calls.$name
    Check ($c.StartedUtc.EndsWith('Z') -and $c.EndedUtc.EndsWith('Z') -and [datetime]::Parse($c.StartedUtc) -ge $previous -and [datetime]::Parse($c.EndedUtc) -ge [datetime]::Parse($c.StartedUtc)) "ordered interval $name"
    $previous=[datetime]::Parse($c.EndedUtc)
}
Check ([datetime]::Parse($s.ObservationEndedUtc) -ge $previous) 'snapshot contains calls'
Check ($null -eq ('Talenro.Diagnostics.NativeObservationV1.NativeProbe' -as [type])) 'synthetic path never loads helper'
$tokens=$null; $errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile($path,[ref]$tokens,[ref]$errors)
Check ($errors.Count -eq 0) 'module parses'
Check (($module.ExportedFunctions.Keys | Sort-Object) -join ',' -ceq 'Get-TalenroNativeProcessObservation,Get-TalenroSyntheticNativeProcessObservation') 'only two public adapters'
foreach ($statement in $ast.EndBlock.Statements) {
    Check ($statement -is [Management.Automation.Language.FunctionDefinitionAst] -or $statement.Extent.Text.StartsWith('Export-ModuleMember ')) 'import declarations only'
}
$source=Get-Content (Join-Path $PSScriptRoot 'native-process-observation.cs') -Raw
Check (([regex]::Matches($source,'\[DllImport\(')).Count -eq 3) 'only three native imports'
Check ($source -notmatch 'OpenProcess|CloseHandle|DuplicateHandle|TerminateProcess|Process\.Start|\.Dispose\(') 'helper has no process lifetime operations'
Check ($source -match 'WaitForSingleObject\(pointer, 0\)' -and $source -notmatch 'Thread\.Sleep|while\s*\(') 'zero wait without polling'
Check ($source -match 'DangerousAddRef\(ref held\)' -and $source -match 'finally[\s\S]*if \(held\) handle.DangerousRelease\(\)') 'balanced borrowed reference structure'
Check (([regex]::Matches($source,'Marshal.GetLastPInvokeError\(\)')).Count -eq 3) 'three immediate error capture sites'
$loader=$ast.Find({param($n) $n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'Get-NativeProbeType'},$true)
Check ($null -ne $loader -and $loader.Extent.Text -match 'Preloaded native helper' -and $loader.Extent.Text -match 'Add-Type') 'private loader fails closed for stale helper'
Check ($ast.Extent.Text -notmatch 'Start-Process|Stop-Process|Get-Process|Invoke-Expression|Set-Content|Out-File') 'adapter no lifecycle or write commands'
"SYNTHETIC_STATIC_PASS: $script:count assertions; native NOT_RUN"
