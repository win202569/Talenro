#requires -Version 7.0
$ErrorActionPreference='Stop'
$entry=Join-Path $PSScriptRoot 'observe-c5.ps1'
if(-not(Test-Path $entry)){throw 'RED: fixed entry missing.'}
$script:checks=0
function Check($condition,$name){if(-not $condition){throw "FAIL: $name"};$script:checks++}
function Start-Process {throw 'Forbidden real launch'}
function Get-CimInstance {throw 'Forbidden real CIM'}
function Add-Type {throw 'Forbidden native compilation'}
function Start-Sleep {throw 'Forbidden real sleep'}
function New-Item {throw 'Forbidden default file creation'}
$result=@(& $entry)
Check ($result.Count -eq 1 -and $result[0].Status -ceq 'NotStarted') 'default returns only NotStarted'
foreach($parameter in @('Executable','Profile','Deadline','Backend')){
    $failed=$false;try{& $entry @{$parameter='forbidden'}|Out-Null}catch{$failed=$true}
    Check $failed "reject $parameter override"
}
Check ($null -eq ('Talenro.Diagnostics.NativeProcessProbe' -as [type])) 'no native helper'
# Execute private definitions only, never the live script body. No CLI test seam.
$tokens=$null;$errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile($entry,[ref]$tokens,[ref]$errors)
Check ($errors.Count -eq 0) 'entry parses'
foreach($fn in $ast.FindAll({param($a)$a -is [Management.Automation.Language.FunctionDefinitionAst]},$false)){
    . ([scriptblock]::Create($fn.Extent.Text))
}
$policy=Get-C5Policy 'D:\fixture'
Check ($policy.Arguments -ceq '--layout-test --layout-language=zh-Hans --layout-profile=unknown-usage') 'fixed arguments'
Check ($policy.ExpectedCommandLine -ceq ('"'+$policy.Executable+'" '+$policy.Arguments)) 'independent command line'
Check ($policy.ExeHash -ceq '4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95') 'fixed EXE hash'
Check ($policy.DllHash -ceq '7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB') 'fixed DLL hash'
$script:fault='';$script:calls=[Collections.Generic.List[string]]::new()
function Assert-C5OrdinaryToken {$script:calls.Add('token');if($script:fault -eq 'token'){throw 'blocked'}}
function Invoke-C5Git {param($Root,$Arguments)
    $script:calls.Add('git')
    switch($Arguments[0]){
        'branch' {if($script:fault -eq 'branch'){'main'}else{'codex/c12-b01-task9-coordinator'}}
        'status' {if($script:fault -eq 'dirty'){' M changed'}}
        'rev-parse' {if($script:fault -eq 'source'){'different'}else{'blob'}}
        'hash-object' {'blob'}
        default {throw 'Unexpected git command'}
    }
}
function Test-Path {param($LiteralPath,$PathType) $script:fault -ne 'missing'}
function Get-FileHash {param($LiteralPath,$Algorithm)
    if($script:fault -eq 'hash'){return @{Hash='wrong'}}
    @{Hash=$(if($LiteralPath.EndsWith('.exe')){$policy.ExeHash}else{$policy.DllHash})}
}
function Get-CimInstance {param($ClassName,$Filter,$ErrorAction)
    $script:calls.Add('cim')
    if($script:fault -eq 'cim'){throw 'inaccessible'}
    if($script:fault -eq 'target'){[pscustomobject]@{ProcessId=12}}
}
Assert-C5Preflight $policy
Check ($script:calls.Contains('cim')) 'complete preflight includes inventory'
foreach($fault in @('token','branch','dirty','source','missing','hash','target','cim')){
    $script:fault=$fault;$failed=$false
    try{Assert-C5Preflight $policy}catch{$failed=$true}
    Check $failed "preflight rejects $fault before acquisition"
}
$script:fault=''
# Private adapter factory: every effect replaced before calling a callback.
$script:events=[Collections.Generic.List[string]]::new()
$script:original=[pscustomobject]@{Id=123;StartTime=[datetime]::UtcNow;SafeHandle=[object]::new()}
function Start-Process {param($FilePath,$ArgumentList,$WorkingDirectory,[switch]$PassThru,$ErrorAction)
    Check ($FilePath -ceq $policy.Executable -and $ArgumentList -ceq $policy.Arguments -and $PassThru) 'fixed one launch adapter'
    $script:events.Add('launch');$script:original
}
function Get-TalenroNativeProcessObservation {param($ProcessHandle,$CorrelationId)
    Check ([object]::ReferenceEquals($ProcessHandle,$script:original.SafeHandle)) 'original native handle'
    $script:events.Add('native');@{CorrelationId=$CorrelationId}
}
function Get-TalenroProcessObservation {param($ProcessObject)
    Check ([object]::ReferenceEquals($ProcessObject,$script:original)) 'original managed object'
    $script:events.Add('managed');@{Mode='fake'}
}
function Get-CimInstance {param($ClassName,$Filter,$ErrorAction)
    Check ($Filter -ceq 'ProcessId = 123') 'CIM exact PID selector'
    [pscustomobject]@{ProcessId=123;ExecutablePath=$policy.Executable;CommandLine=$policy.ExpectedCommandLine;CreationDate=[datetime]::UtcNow}
}
function Start-Sleep {param($Milliseconds) Check ($Milliseconds -eq 7) 'sleep adapter forwards bounded value'}
$backend=New-C5Backend $policy
Check ($backend.Count -eq 10) 'complete backend'
Check ((& $backend.ElapsedMs) -eq 0) 'clock not started in factory'
$process=& $backend.AcquireProcess @{}
Check ([object]::ReferenceEquals($process,$script:original)) 'launch transfers original'
$identity=& $backend.GetIdentity $process
Check ([object]::ReferenceEquals($identity.Handle,$script:original.SafeHandle)) 'identity borrows handle'
Check ($identity.StartUtc.Kind -eq [DateTimeKind]::Utc) 'process explicit UTC'
& $backend.ReadNative $identity.Handle ('a'*32)|Out-Null
& $backend.ReadManaged $process|Out-Null
$cim=& $backend.ReadCim 123
Check ($cim.CreationUtc.Kind -eq [DateTimeKind]::Utc) 'CIM explicit UTC'
& $backend.SleepMs 7
Check (($script:events -join ',') -ceq 'launch,native,managed') 'adapter order and single launch'
Write-Output "PASS: $script:checks fixed-entry synthetic assertions; no real launch/CIM/native/sleep."
