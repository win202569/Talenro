#requires -Version 7.0
$ErrorActionPreference='Stop'
$entry=Join-Path $PSScriptRoot 'observe-c4.ps1'
$script:checks=0
function Check($condition,$name){if(-not $condition){throw "FAIL: $name"};$script:checks++}
function Reject([scriptblock]$Action,$name){$failed=$false;try{& $Action|Out-Null}catch{$failed=$true};Check $failed $name}
# External effects are replaced before any entry code executes.
function Start-Process {throw 'Forbidden launch'}
function Get-CimInstance {throw 'Forbidden CIM'}
function Add-Type {throw 'Forbidden compilation'}
function Start-Sleep {throw 'Forbidden sleep'}
function New-Item {throw 'Forbidden write'}
function Import-Module {throw 'Forbidden default import'}
Check (Test-Path -LiteralPath $entry) 'RED: C4 entry missing'
$result=@(& $entry)
Check ($result.Count -eq 1 -and $result[0].Status -ceq 'NotStarted') 'one no-start result'
foreach($parameter in @('Executable','Profile','Language','Deadline','Backend')){
    Reject {& $entry @{$parameter='forbidden'}} "reject public $parameter override"
}
Check ($null -eq ('Talenro.Diagnostics.NativeObservationV1.NativeProbe' -as [type])) 'native helper absent'
$tokens=$null;$errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile($entry,[ref]$tokens,[ref]$errors)
Check ($errors.Count -eq 0) 'entry parses'
foreach($fn in $ast.FindAll({param($a)$a -is [Management.Automation.Language.FunctionDefinitionAst]},$false)){
    . ([scriptblock]::Create($fn.Extent.Text))
}
$policy=Get-C4Policy 'D:\fixture'
Check ($policy.Arguments -ceq '--layout-test --layout-language=en-US --layout-profile=rtl-mixed') 'fixed C4 arguments'
Check ($policy.ExpectedCommandLine -ceq ('"'+$policy.Executable+'" --layout-test --layout-language=en-US --layout-profile=rtl-mixed')) 'independent command line'
Check ($policy.ExeHash -ceq '4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95') 'EXE identity'
Check ($policy.DllHash -ceq '7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB') 'DLL identity'
Check (($policy.Sources -join ',') -ceq 'observe-c4.ps1,c5-observation-session.psm1,observation-session-guard.psm1,process-observation.psm1,native-process-observation.psm1,native-process-observation.cs') 'all consumed sources checked'
Remove-Item Function:Import-Module
Import-Module (Join-Path $PSScriptRoot 'c5-observation-session.psm1')
$script:fault='';$script:acquisitions=0;$script:inventory=0
function Assert-C4OrdinaryToken {if($script:fault -eq 'token'){throw 'blocked'}}
function Invoke-C4Git {param($Root,$Arguments)
    switch($Arguments[0]){
        'branch' {if($script:fault -eq 'branch'){'main'}else{'codex/c12-b01-task9-coordinator'}}
        'status' {if($script:fault -eq 'dirty'){' M changed'}}
        'rev-parse' {if($script:fault -eq 'source'){'different'}else{'blob'}}
        'hash-object' {'blob'}
        default {throw 'Unexpected git command'}
    }
}
function Test-Path {param($LiteralPath,$PathType,$ErrorAction) $script:fault -ne 'missing'}
function Get-FileHash {param($LiteralPath,$Algorithm)
    if($script:fault -eq 'hash'){return @{Hash='wrong'}}
    @{Hash=$(if($LiteralPath.EndsWith('.exe')){$policy.ExeHash}else{$policy.DllHash})}
}
function Get-CimInstance {param($ClassName,$Filter,$ErrorAction)
    $script:inventory++
    Check ($Filter -ceq "Name LIKE 'Talenro.Demo.%' OR Name LIKE '%Prototype%'") 'conservative inventory filter'
    if($script:fault -eq 'cim'){throw 'inaccessible'}
    if($script:fault -eq 'target'){@{ProcessId=12}}
}
$context=@{RunId='a'*32;ExpectedPath=$policy.Executable;ExpectedCommandLine=$policy.ExpectedCommandLine}
$gate=@{
    Preflight={param($c) Assert-C4Preflight $policy}
    AcquireWriter={$script:acquisitions++;throw 'unexpected writer'}
    AcquireProcess={$script:acquisitions++;throw 'unexpected process'}
    GetIdentity={};ReadNative={};ReadManaged={};ReadCim={};ElapsedMs={0};SleepMs={};Publish={}
}
Assert-C4Preflight $policy
Check ($script:inventory -eq 1) 'complete preflight'
foreach($fault in @('token','branch','dirty','source','missing','hash','target','cim')){
    $script:fault=$fault
    Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $gate} "preflight rejects $fault"
    Check ($script:acquisitions -eq 0) "no acquisitions after $fault"
}
$script:fault=''
Check ($null -ne (Get-Command New-C4ReceiptWriter -ErrorAction SilentlyContinue)) 'RED: receipt factory missing'
# A removed path check must expose collision/reparse destinations before any write.
function Test-Path {param($LiteralPath,$PathType,$ErrorAction) $LiteralPath -in $script:existing}
function Get-Item {param($LiteralPath,$Force,$ErrorAction)
    @{Attributes=$(if($LiteralPath -eq $script:reparse){[IO.FileAttributes]::ReparsePoint}else{[IO.FileAttributes]::Directory})}
}
$destination=Join-Path $policy.ReceiptRoot $context.RunId
$script:existing=@();$script:reparse=''
Assert-C4ReceiptDestination $destination
Check $true 'new destination accepted'
foreach($kind in @('directory','file')){
    $script:existing=@($destination)
    Reject {Assert-C4ReceiptDestination $destination} "existing $kind rejected"
}
$script:existing=@($policy.ReceiptRoot);$script:reparse=$policy.ReceiptRoot
Reject {Assert-C4ReceiptDestination $destination} 'reparse ancestor rejected'
$script:existing=@();$script:reparse=''
# Capture real factory and encoder; wrappers below isolate all disk access.
$script:factory=Get-Command New-C4ReceiptWriter
$encoder=Get-Command New-C4TextWriter
$memory=[IO.MemoryStream]::new();$encoded=& $encoder $memory
$encoded.WriteLine('C4')
Check ($memory.Length -eq 4 -and $memory.ToArray()[0] -eq 67) 'UTF8 no BOM and immediate flush'
$encoded.Dispose()
function Reset-Receipt($fault=''){
    $script:receipt=@{Fault=$fault;Events=[Collections.Generic.List[string]]::new();Resources=[Collections.Generic.List[object]]::new();Rows=[Collections.Generic.List[string]]::new();Launches=0}
}
function Assert-C4ReceiptDestination {param($Directory)
    $script:receipt.Events.Add('check');if($script:receipt.Fault -eq 'collision'){throw 'private-collision'}
}
function New-C4ReceiptDirectory {param($Directory)
    $script:receipt.Events.Add('directory');if($script:receipt.Fault -eq 'directory'){throw 'private-directory'}
}
function Open-C4NewStream {param($Path)
    $kind=if($Path.EndsWith('manifest.json')){'manifest'}elseif($Path.EndsWith('samples.jsonl')){'samples'}else{throw 'wrong receipt path'}
    $script:receipt.Events.Add("$kind-open")
    if($script:receipt.Fault -eq "$kind-open"){throw 'private-open'}
    $s=[pscustomobject]@{Kind=$kind;State=$script:receipt;Disposed=$false;Memory=[IO.MemoryStream]::new()}
    $s|Add-Member ScriptMethod Dispose {$this.Disposed=$true;$this.State.Events.Add($this.Kind+'-stream-dispose');$this.Memory.Dispose()}
    $script:receipt.Resources.Add($s);$s
}
function New-C4TextWriter {param($Stream)
    if($script:receipt.Fault -eq ($Stream.Kind+'-writer')){throw 'private-writer'}
    $w=[pscustomobject]@{Kind=$Stream.Kind;State=$script:receipt;Stream=$Stream;Disposed=$false}
    $w|Add-Member ScriptMethod WriteLine {param($line)
        $this.State.Events.Add($this.Kind+'-write')
        if($this.State.Fault -eq ($this.Kind+'-write')){throw 'private-write'}
        $this.State.Rows.Add($line);$this.Stream.Memory.WriteByte(1)
    }
    $w|Add-Member ScriptMethod Flush {
        $this.State.Events.Add($this.Kind+'-flush');if($this.State.Fault -eq ($this.Kind+'-flush')){throw 'private-flush'}
    }
    $w|Add-Member ScriptMethod Dispose {
        $this.Disposed=$true;$this.State.Events.Add($this.Kind+'-writer-dispose')
        if($this.State.Fault -eq ($this.Kind+'-dispose')){throw 'private-dispose'}
        $this.Stream.Dispose()
    }
    $script:receipt.Resources.Add($w);$w
}
Reset-Receipt
$owned=@(& $script:factory $policy $context)
Check ($owned.Count -eq 1 -and $owned[0].Kind -eq 'samples') 'exactly one transferred writer'
Check (($script:receipt.Events -join ',') -ceq 'check,directory,manifest-open,manifest-write,manifest-flush,manifest-writer-dispose,manifest-stream-dispose,manifest-stream-dispose,samples-open') 'manifest closed before sample acquisition'
$manifest=$script:receipt.Rows[0]|ConvertFrom-Json
$expectedManifest=@{SchemaVersion=1;Observation='C4';RunId='a'*32;Language='en-US';Profile='rtl-mixed';ExecutableRelativePath='clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe';Arguments='--layout-test --layout-language=en-US --layout-profile=rtl-mixed';ExeSha256='4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95';DllSha256='7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB';ObservationCapMs=360000;MaxSleepMs=2000}
Check (@($manifest.PSObject.Properties).Count -eq 11) 'manifest no private extras'
foreach($key in $expectedManifest.Keys){Check ($manifest.$key -ceq $expectedManifest[$key]) "manifest $key"}
Check (-not $owned[0].Disposed) 'sample ownership transferred alive'
$owned[0].Dispose()
$gate.Preflight={}
$gate.AcquireWriter={param($c) & $script:factory $policy $c}
$gate.AcquireProcess={$script:receipt.Launches++;throw 'unexpected launch'}
foreach($fault in @('collision','directory','manifest-open','manifest-writer','manifest-write','manifest-flush','manifest-dispose','samples-open','samples-writer')){
    Reset-Receipt $fault
    Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $gate} "receipt failure blocks: $fault"
    Check ($script:receipt.Launches -eq 0) "no launch: $fault"
    Check (@($script:receipt.Resources|Where-Object {-not $_.Disposed}).Count -eq 0) "all owned resources cleaned: $fault"
    Check (($script:receipt.Rows -join '') -notmatch 'private-') "no private receipt payload: $fault"
}
foreach($bad in @(('A'*32),('0'*32),'../escape','', 'abc')){
    Reset-Receipt;$invalid=$context.Clone();$invalid.RunId=$bad
    Reject {& $script:factory $policy $invalid} 'invalid receipt id'
    Check ($script:receipt.Events.Count -eq 0) 'invalid id no filesystem effects'
}
foreach($key in @('ExpectedPath','ExpectedCommandLine')){
    Reset-Receipt;$invalid=$context.Clone();$invalid[$key]='wrong'
    Reject {& $script:factory $policy $invalid} "mismatched $key"
    Check ($script:receipt.Events.Count -eq 0) 'context mismatch no filesystem effects'
}
Check ($null -ne (Get-Command New-C4Backend -ErrorAction SilentlyContinue)) 'RED: backend missing'
Import-Module (Join-Path $PSScriptRoot 'native-process-observation.psm1')
Import-Module (Join-Path $PSScriptRoot 'process-observation.psm1')
$script:events=[Collections.Generic.List[string]]::new()
$script:start=[datetime]::Parse('2026-11-09T18:39:47.4427247Z').ToUniversalTime()
$script:original=[pscustomobject]@{Id=123;StartTime=$script:start;SafeHandle=[object]::new();Disposed=$false}
$script:original|Add-Member ScriptMethod Dispose {$this.Disposed=$true}
$script:cimFault='';$script:managedFault=$false
function Assert-C4Preflight {param($Policy) $script:events.Add('preflight')}
function Start-Process {param($FilePath,$ArgumentList,$WorkingDirectory,[switch]$PassThru,$ErrorAction)
    Check ($FilePath -ceq $policy.Executable -and $ArgumentList -ceq '--layout-test --layout-language=en-US --layout-profile=rtl-mixed' -and $PassThru) 'fixed launch arguments'
    Check ($WorkingDirectory -ceq [IO.Path]::GetDirectoryName($policy.Executable)) 'fixed launch directory'
    Check ((& $script:backend.ElapsedMs) -gt 0) 'clock running before launch'
    $script:events.Add('launch');$script:original
}
function Get-TalenroNativeProcessObservation {param($ProcessHandle,$CorrelationId)
    Check ([object]::ReferenceEquals($ProcessHandle,$script:original.SafeHandle)) 'original handle'
    $script:events.Add('native')
    $ft=$script:start.ToFileTimeUtc();$exit=$ft+10000
    $times=@{Creation=@{High=[uint32]($ft -shr 32);Low=[uint32]($ft -band 4294967295L)};Exit=@{High=[uint32]($exit -shr 32);Low=[uint32]($exit -band 4294967295L)};Kernel=@{High=[uint32]0;Low=[uint32]0};User=@{High=[uint32]0;Low=[uint32]0}}
    Get-TalenroSyntheticNativeProcessObservation -CorrelationId $CorrelationId -Readers @{
        Wait={@{RawResult=[uint32]0}};ExitCode={@{ApiSucceeded=$true;ExitCode=[uint32]0}}
        Times={@{ApiSucceeded=$true;Times=$times}}.GetNewClosure()
    }
}
function Get-TalenroProcessObservation {param($ProcessObject)
    Check ([object]::ReferenceEquals($ProcessObject,$script:original)) 'original managed object'
    $script:events.Add('managed');if($script:managedFault){throw 'private-managed'}
    Get-TalenroSyntheticProcessObservation -Readers @{
        ProcessId={[int]123};StartTime={$script:start};WaitSignaled={$true};HasExited={$true};ExitCode={[int]0};ExitTime={$script:start.AddTicks(10000)}
    }
}
function Get-CimInstance {param($ClassName,$Filter,$ErrorAction)
    Check ($ClassName -ceq 'Win32_Process' -and $Filter -ceq 'ProcessId = 123') 'exact CIM PID filter'
    $date=$script:start.AddTicks(-7)
    if($script:cimFault -eq 'zero'){return}
    if($script:cimFault -eq 'missing'){$date=$null}
    if($script:cimFault -eq 'unspecified'){$date=[datetime]::SpecifyKind($date,[DateTimeKind]::Unspecified)}
    if($script:cimFault -eq 'local'){$date=$date.ToLocalTime()}
    $row=@{ProcessId=123;ExecutablePath=$policy.Executable;CommandLine=$policy.ExpectedCommandLine;CreationDate=$date}
    $row;if($script:cimFault -eq 'multiple'){$row}
}
function Start-Sleep {param($Milliseconds) Check ($Milliseconds -eq 7) 'bounded sleep forwarding only'}
$script:backend=New-C4Backend $policy
Check ((($script:backend.Keys|Sort-Object) -join ',') -ceq 'AcquireProcess,AcquireWriter,ElapsedMs,GetIdentity,Preflight,Publish,ReadCim,ReadManaged,ReadNative,SleepMs') 'exact ten callbacks'
Check ((& $script:backend.ElapsedMs) -eq 0) 'clock not running before acquisition'
$process=& $script:backend.AcquireProcess $context
Check ([object]::ReferenceEquals($process,$script:original)) 'original process transferred'
$identity=& $script:backend.GetIdentity $process
Check ([object]::ReferenceEquals($identity.Handle,$script:original.SafeHandle) -and $identity.StartUtc.Kind -eq 'Utc') 'identity original handle and UTC'
foreach($bad in @('zero','multiple','missing','unspecified')){
    $script:cimFault=$bad;Reject {& $script:backend.ReadCim 123} "CIM rejects $bad"
}
$script:cimFault='local';$row=& $script:backend.ReadCim 123
Check ($row.CreationUtc.Kind -eq 'Utc' -and $row.CreationUtc -eq $script:start.AddTicks(-7)) 'local CIM converted UTC'
& $script:backend.SleepMs 7
$script:cimFault=''
# Real C4 factory + real ten-callback wiring + shared orchestration; only OS effects replaced.
foreach($failure in @($false,$true)){
    Reset-Receipt;$script:events.Clear();$script:original.Disposed=$false;$script:managedFault=$failure
    $script:backend=New-C4Backend $policy
    $progress=[Collections.Generic.List[object]]::new()
    $script:backend.Publish={param($r)$progress.Add($r)}.GetNewClosure()
    if($failure){Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $script:backend} 'managed failure retained'}
    else{$result=Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $script:backend;Check ($result.ExitVerified -and $result.Finalized) 'C4 finalized synthetic exit'}
    Check (($script:events -join ',') -ceq 'preflight,launch,native,native,managed') 'native first single launch'
    Check ($script:original.Disposed -and @($script:receipt.Resources|Where-Object {-not $_.Disposed}).Count -eq 0) 'all originals disposed'
    $rows=@($script:receipt.Rows|ForEach-Object {$_|ConvertFrom-Json})
    Check (@($rows|Where-Object RunId -CNE ('a'*32)).Count -eq 0 -and $rows[0].Observation -ceq 'C4') 'manifest and samples correlate'
    Check (($script:receipt.Rows -join '') -notmatch 'private-') 'sanitized shared failure receipts'
    if($failure){Check ($rows[-1].Kind -ceq 'outcome' -and $rows[-1].Outcome -ceq 'observation-failed' -and @($rows|Where-Object {$_.Kind -eq 'pair' -and $_.Incomplete}).Count -eq 1) 'failure preserves incomplete native pair'}
}
Reset-Receipt;$script:events.Clear();$script:backend=New-C4Backend $policy;$script:backend.Remove('ReadNative')
Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $script:backend} 'missing callback rejected'
Check ($script:events.Count -eq 0 -and $script:receipt.Events.Count -eq 0) 'missing callback no live fallback'
Check ($null -eq ('Talenro.Diagnostics.NativeObservationV1.NativeProbe' -as [type])) 'native helper still absent'
Write-Output "PASS: $script:checks C4 synthetic assertions; no real launch/CIM/native/sleep."
