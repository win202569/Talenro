#requires -Version 7.0
$ErrorActionPreference='Stop'
$entry=Join-Path $PSScriptRoot 'observe-ordinary-account-center.ps1'
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
function git {throw 'Forbidden default git'}
function Get-FileHash {throw 'Forbidden default hash'}
function Stop-Process {throw 'Forbidden termination'}
Check (Test-Path -LiteralPath $entry) 'RED: Ordinary entry missing'
$result=@(& $entry)
Check ($result.Count -eq 1 -and $result[0].Status -ceq 'NotStarted') 'one no-start result'
foreach($parameter in @('Executable','Hash','Arguments','Profile','Language','Deadline','Backend')){
    Reject {& $entry @{$parameter='forbidden'}} "reject public $parameter override"
}
Check ($null -eq ('Talenro.Diagnostics.NativeObservationV1.NativeProbe' -as [type])) 'native helper absent'
$tokens=$null;$errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile($entry,[ref]$tokens,[ref]$errors)
Check ($errors.Count -eq 0) 'entry parses'
Check ($ast.ParamBlock.Parameters.Count -eq 1 -and $ast.ParamBlock.Parameters[0].Name.VariablePath.UserPath -ceq 'RunApproved') 'only approval parameter'
Check (@($ast.FindAll({param($a) $a -is [Management.Automation.Language.CommandParameterAst] -and $a.ParameterName -eq 'ArgumentList'},$true)).Count -eq 0) 'no ArgumentList launch parameter'
foreach($fn in $ast.FindAll({param($a)$a -is [Management.Automation.Language.FunctionDefinitionAst]},$false)){
    . ([scriptblock]::Create($fn.Extent.Text))
}
$policy=Get-OrdinaryPolicy 'D:\fixture with spaces'
Check ($policy.Arguments -ceq '') 'fixed Ordinary arguments'
Check ($policy.ExpectedCommandLine -ceq ('"'+$policy.Executable+'"')) 'independent command line'
$expectedHashes=@{
    'Talenro.Demo.App.exe'='FEDA8E835E8390ED7F166FA764413FBEC9BCE591C7C7CF7E8E43E028B03C0375'
    'Talenro.Demo.App.dll'='773AE00EC4E5552D65ECDFA527FEEE6688AE380D8ED4DB4D3282DD10C1B76A0B'
    'Talenro.Demo.Core.dll'='67FDB3F42E69350400B8A9FC11BCFB54670233B4571F2E27BA91F30E8D19DE18'
    'Views/AccountPage.xbf'='A50DDC403AD9E41EA977AE868551AE25B6079674094EBE0BE38CD12C2B7EFE6F'
    'Talenro.Demo.App.deps.json'='99DD4D70751D66997DF9DA858C09DDA2BB68A8FDFAC7F671E4FAB5B35D910ED6'
    'Talenro.Demo.App.runtimeconfig.json'='3FA241FAED5CC59579FA3B630E1A873E5539428B9CED9A4CA48F5828488E9666'
}
Check ($policy.ArtifactHashes.Count -eq 6) 'six artifacts'
foreach($name in $expectedHashes.Keys){Check ($policy.ArtifactHashes[$name] -ceq $expectedHashes[$name]) "binding $name"}
Check ($policy.BuildSourceCommit -ceq '5ee8a339b574171fc61f3cf0c0c636bd86421397') 'build source'
Check (($policy.Sources -join ',') -ceq 'observe-ordinary-account-center.ps1,c5-observation-session.psm1,observation-session-guard.psm1,process-observation.psm1,native-process-observation.psm1,native-process-observation.cs') 'all consumed sources checked'
Remove-Item Function:Import-Module
Import-Module (Join-Path $PSScriptRoot 'c5-observation-session.psm1')
$script:fault='';$script:acquisitions=0;$script:inventory=0
function Assert-OrdinaryOrdinaryToken {if($script:fault -eq 'token'){throw 'blocked'}}
function Invoke-OrdinaryGit {param($Root,$Arguments)
    switch($Arguments[0]){
        'branch' {if($script:fault -eq 'branch'){'main'}else{'codex/c12-b01-task9-coordinator'}}
        'status' {if($script:fault -eq 'dirty'){' M changed'};if($script:fault -eq 'untracked'){'?? new'}}
        'rev-parse' {if($script:fault -eq 'source' -or $Arguments[1] -ceq $script:badSource){'different'}else{'blob'}}
        'hash-object' {'blob'}
        'cat-file' {Check ($Arguments[2] -ceq '5ee8a339b574171fc61f3cf0c0c636bd86421397^{commit}') 'fixed source object';if($script:fault -eq 'history'){throw 'missing commit'}}
        'diff' {
            Check (($Arguments -join '|') -ceq 'diff|--exit-code|5ee8a339b574171fc61f3cf0c0c636bd86421397|--|clients/windows/Talenro.Demo.App|clients/windows/Talenro.Demo.Core|clients/windows/Directory.Build.props') 'all product paths bound'
            if($script:fault -eq 'product' -or ($script:changedProduct -and $script:changedProduct -in $Arguments)){throw 'changed product'}
        }
        default {throw 'Unexpected git command'}
    }
}
function Test-Path {param($LiteralPath,$PathType,$ErrorAction) $script:fault -ne 'missing'}
function Get-Item {param($LiteralPath,[switch]$Force,$ErrorAction) @{Attributes=[IO.FileAttributes]::Normal}}
function Get-FileHash {param($LiteralPath,$Algorithm,$ErrorAction)
    if($script:fault -eq 'hash'){return @{Hash='wrong'}}
    foreach($name in $expectedHashes.Keys){if($LiteralPath -eq (Join-Path ([IO.Path]::GetDirectoryName($policy.Executable)) $name)){return @{Hash=$expectedHashes[$name]}}}
    throw 'unexpected hash path'
}
function Get-CimInstance {param($ClassName,$Filter,$ErrorAction)
    $script:inventory++
    Check ($Filter -ceq "Name LIKE 'Talenro.Demo.%' OR Name LIKE '%Prototype%'") 'conservative inventory filter'
    if($script:fault -eq 'cim'){throw 'inaccessible'}
    if($script:fault -eq 'target'){@{ProcessId=12}}
}
$context=@{RunId='a'*32;ExpectedPath=$policy.Executable;ExpectedCommandLine=$policy.ExpectedCommandLine}
$gate=@{
    Preflight={param($c) Assert-OrdinaryPreflight $policy}
    AcquireWriter={$script:acquisitions++;throw 'unexpected writer'}
    AcquireProcess={$script:acquisitions++;throw 'unexpected process'}
    GetIdentity={};ReadNative={};ReadManaged={};ReadCim={};ElapsedMs={0};SleepMs={};Publish={}
}
Assert-OrdinaryPreflight $policy
Check ($script:inventory -eq 1) 'complete preflight'
foreach($fault in @('token','branch','dirty','untracked','source','history','product','missing','hash','target','cim')){
    $script:fault=$fault
    Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $gate} "preflight rejects $fault"
    Check ($script:acquisitions -eq 0) "no acquisitions after $fault"
}
$script:fault=''
foreach($source in $policy.Sources){
    $script:badSource='HEAD:clients/windows/tests/'+$source
    Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $gate} "changed source $source"
    Check ($script:acquisitions -eq 0) 'source failure before acquisition'
}
$script:badSource=''
foreach($path in @('clients/windows/Talenro.Demo.App','clients/windows/Talenro.Demo.Core','clients/windows/Directory.Build.props')){
    $script:changedProduct=$path
    Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $gate} "changed product $path"
    Check ($script:acquisitions -eq 0) 'product failure before acquisition'
}
$script:changedProduct=''
function Test-Path {param($LiteralPath,$PathType,$ErrorAction) $LiteralPath -ne $script:missingPath}
function Get-Item {param($LiteralPath,[switch]$Force,$ErrorAction)
    @{Attributes=$(if($LiteralPath -eq $script:reparsePath){[IO.FileAttributes]::ReparsePoint}else{[IO.FileAttributes]::Normal})}
}
function Get-FileHash {param($LiteralPath,$Algorithm,$ErrorAction)
    if($LiteralPath -eq $script:badHashPath){return @{Hash='wrong'}}
    foreach($name in $expectedHashes.Keys){if($LiteralPath -eq (Join-Path ([IO.Path]::GetDirectoryName($policy.Executable)) $name)){return @{Hash=$expectedHashes[$name]}}}
    throw 'unexpected hash path'
}
$script:missingPath='';$script:badHashPath='';$script:reparsePath=''
foreach($name in $expectedHashes.Keys){
    $path=Join-Path ([IO.Path]::GetDirectoryName($policy.Executable)) $name
    foreach($kind in @('missingPath','badHashPath')){
        Set-Variable -Scope Script -Name $kind -Value $path
        Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $gate} "$kind $name rejected"
        Check ($script:acquisitions -eq 0) 'artifact rejection before acquisition'
        Set-Variable -Scope Script -Name $kind -Value ''
    }
    while($path){
        $script:reparsePath=$path
        Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $gate} "reparse $path rejected"
        Check ($script:acquisitions -eq 0) 'reparse before acquisition'
        $path=[IO.Path]::GetDirectoryName($path)
    }
    $script:reparsePath=''
}
Assert-OrdinaryPreflight $policy
Check ($null -ne (Get-Command New-OrdinaryReceiptWriter -ErrorAction SilentlyContinue)) 'RED: receipt factory missing'
# A removed path check must expose collision/reparse destinations before any write.
function Test-Path {param($LiteralPath,$PathType,$ErrorAction) $LiteralPath -in $script:existing}
function Get-Item {param($LiteralPath,[switch]$Force,$ErrorAction)
    @{Attributes=$(if($LiteralPath -eq $script:reparse){[IO.FileAttributes]::ReparsePoint}else{[IO.FileAttributes]::Directory})}
}
$destination=Join-Path $policy.ReceiptRoot $context.RunId
$script:existing=@();$script:reparse=''
Assert-OrdinaryReceiptDestination $destination
Check $true 'new destination accepted'
foreach($kind in @('directory','file')){
    $script:existing=@($destination)
    Reject {Assert-OrdinaryReceiptDestination $destination} "existing $kind rejected"
}
$script:existing=@($policy.ReceiptRoot);$script:reparse=$policy.ReceiptRoot
Reject {Assert-OrdinaryReceiptDestination $destination} 'reparse ancestor rejected'
$script:existing=@();$script:reparse=''
# Capture real factory and encoder; wrappers below isolate all disk access.
$script:factory=Get-Command New-OrdinaryReceiptWriter
$encoder=Get-Command New-OrdinaryTextWriter
# Exercise the actual opener, not its memory double. Only these fresh ignored
# fixture files touch disk; never use the candidate or its observation directory.
$realOpener=Get-Command Open-OrdinaryNewStream
$fixtureRoot=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../.superpowers/sdd/2026-11-13-windows-ordinary-minimal-observation/fixtures'))
$fixtureDirectory=Join-Path $fixtureRoot ([guid]::NewGuid().ToString('N'))
$null=[IO.Directory]::CreateDirectory($fixtureDirectory)
foreach($filename in @('manifest.json','samples.jsonl')){
    $fixturePath=Join-Path $fixtureDirectory $filename
    $stream=$null
    try{
        $stream=& $realOpener $fixturePath
        Check ($stream -is [IO.FileStream] -and $stream.CanWrite) "real opener creates $filename"
        $stream.WriteByte(42);$stream.Flush()
    }finally{if($null -ne $stream){$stream.Dispose()}}
    $second=$null;$collision=$null
    try{$second=& $realOpener $fixturePath}catch{$collision=$_.Exception}
    finally{if($null -ne $second){$second.Dispose()}}
    Check ($null -ne $collision -and $null -eq $second) "real opener refuses existing $filename"
    $bytes=[IO.File]::ReadAllBytes($fixturePath)
    Check ($bytes.Length -eq 1 -and $bytes[0] -eq 42) "existing $filename bytes preserved"
}
$memory=[IO.MemoryStream]::new();$encoded=& $encoder $memory
$encoded.WriteLine('C4')
Check ($memory.Length -eq 4 -and $memory.ToArray()[0] -eq 67) 'UTF8 no BOM and immediate flush'
$encoded.Dispose()
function Reset-Receipt($fault=''){
    $script:receipt=@{Fault=$fault;Events=[Collections.Generic.List[string]]::new();Resources=[Collections.Generic.List[object]]::new();Rows=[Collections.Generic.List[string]]::new();Launches=0}
}
function Assert-OrdinaryReceiptDestination {param($Directory)
    $script:receipt.Events.Add('check');if($script:receipt.Fault -eq 'collision'){throw 'private-collision'}
}
function New-OrdinaryReceiptDirectory {param($Directory)
    $script:receipt.Events.Add('directory');if($script:receipt.Fault -eq 'directory'){throw 'private-directory'}
}
function Open-OrdinaryNewStream {param($Path)
    $kind=if($Path.EndsWith('manifest.json')){'manifest'}elseif($Path.EndsWith('samples.jsonl')){'samples'}else{throw 'wrong receipt path'}
    $script:receipt.Events.Add("$kind-open")
    if($script:receipt.Fault -eq "$kind-open"){throw 'private-open'}
    $s=[pscustomobject]@{Kind=$kind;State=$script:receipt;Disposed=$false;Memory=[IO.MemoryStream]::new()}
    $s|Add-Member ScriptMethod Dispose {
        $this.Disposed=$true;$this.State.Events.Add($this.Kind+'-stream-dispose');$this.Memory.Dispose()
        if($this.State.Fault -contains ($this.Kind+'-stream-dispose')){throw 'private-stream-dispose'}
    }
    $script:receipt.Resources.Add($s);$s
}
function New-OrdinaryTextWriter {param($Stream)
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
$expectedManifest=@{SchemaVersion=1;Observation='OrdinaryAccountCenter';RunId='a'*32;Language='default-environment';Profile='default-environment';ExecutableRelativePath='clients/windows/Talenro.Demo.App/bin/ordinary-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe';Arguments='';BuildSourceCommit='5ee8a339b574171fc61f3cf0c0c636bd86421397';ObservationCapMs=360000;MaxSleepMs=2000}
Check (@($manifest.PSObject.Properties).Count -eq 11) 'manifest no private extras'
foreach($name in $expectedHashes.Keys){Check ($manifest.ArtifactHashes.$name -ceq $expectedHashes[$name]) "manifest hash $name"}
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
# Direct factory assertions: the primary error must survive alongside every
# subsequent cleanup error, and later cleanup must run after an earlier throw.
foreach($case in @(
    @{Faults=@('manifest-write','manifest-dispose','manifest-stream-dispose');Primary='private-write';Count=3;Writer=$true;Kind='manifest'},
    @{Faults=@('manifest-flush','manifest-dispose','manifest-stream-dispose');Primary='private-flush';Count=3;Writer=$true;Kind='manifest'},
    @{Faults=@('manifest-writer','manifest-stream-dispose');Primary='private-writer';Count=2;Writer=$false;Kind='manifest'},
    @{Faults=@('samples-writer','samples-stream-dispose');Primary='private-writer';Count=2;Writer=$false;Kind='samples'}
)){
    Reset-Receipt $case.Faults
    $failure=$null;$output=@()
    try{$output=@(& $script:factory $policy $context)}catch{$failure=$_.Exception}
    Check ($failure -is [AggregateException] -and $output.Count -eq 0) 'combined failure has no successful writer'
    Check ($failure.InnerExceptions.Count -eq $case.Count) 'all primary and cleanup errors retained'
    Check ($failure.InnerExceptions[0].ToString().Contains($case.Primary)) 'primary error retained first'
    if($case.Writer){
        Check ($failure.InnerExceptions[1].ToString().Contains('private-dispose')) 'writer disposal error retained second'
        Check (($script:receipt.Events | Select-Object -Last 2) -join ',' -ceq ($case.Kind+'-writer-dispose,'+$case.Kind+'-stream-dispose')) 'stream cleanup continues after writer failure'
    }
    Check ($failure.InnerExceptions[-1].ToString().Contains('private-stream-dispose')) 'stream cleanup error retained last'
    Check (@($script:receipt.Resources|Where-Object {-not $_.Disposed}).Count -eq 0) 'all acquired combined-failure resources attempted cleanup'
    Check ($script:receipt.Launches -eq 0) 'combined factory failure never acquires process'
}
foreach($key in @('ExpectedPath','ExpectedCommandLine')){
    Reset-Receipt;$invalid=$context.Clone();$invalid[$key]='wrong'
    Reject {& $script:factory $policy $invalid} "mismatched $key"
    Check ($script:receipt.Events.Count -eq 0) 'context mismatch no filesystem effects'
}
Check ($null -ne (Get-Command New-OrdinaryBackend -ErrorAction SilentlyContinue)) 'RED: backend missing'
Import-Module (Join-Path $PSScriptRoot 'native-process-observation.psm1')
Import-Module (Join-Path $PSScriptRoot 'process-observation.psm1')
$script:events=[Collections.Generic.List[string]]::new()
$script:start=[datetime]::Parse('2026-11-09T18:39:47.4427247Z').ToUniversalTime()
$script:original=[pscustomobject]@{Id=123;StartTime=$script:start;SafeHandle=[object]::new();Disposed=$false}
$script:original|Add-Member ScriptMethod Dispose {$this.Disposed=$true}
$script:cimFault='';$script:managedFault=$false
function Assert-OrdinaryPreflight {param($Policy) $script:events.Add('preflight')}
function Start-Process {param($FilePath,$ArgumentList,$WorkingDirectory,[switch]$PassThru,$ErrorAction)
    Check ($FilePath -ceq $policy.Executable -and (-not $PSBoundParameters.ContainsKey('ArgumentList')) -and $PassThru) 'fixed launch arguments'
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
    if($script:cimFault -eq 'trailing'){$row.CommandLine+=' '}
    if($script:cimFault -eq 'extra'){$row.CommandLine+=' --layout-test'}
    $row;if($script:cimFault -eq 'multiple'){$row}
}
function Start-Sleep {param($Milliseconds) Check ($Milliseconds -eq 7) 'bounded sleep forwarding only'}
$script:backend=New-OrdinaryBackend $policy
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
# Real Ordinary factory + real ten-callback wiring + shared orchestration; only OS effects replaced.
foreach($failure in @($false,$true)){
    Reset-Receipt;$script:events.Clear();$script:original.Disposed=$false;$script:managedFault=$failure
    $script:backend=New-OrdinaryBackend $policy
    $progress=[Collections.Generic.List[object]]::new()
    $script:backend.Publish={param($r)$progress.Add($r)}.GetNewClosure()
    if($failure){Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $script:backend} 'managed failure retained'}
    else{$result=Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $script:backend;Check ($result.ExitVerified -and $result.Finalized) 'Ordinary finalized synthetic exit'}
    Check (($script:events -join ',') -ceq 'preflight,launch,native,native,managed') 'native first single launch'
    Check ($script:original.Disposed -and @($script:receipt.Resources|Where-Object {-not $_.Disposed}).Count -eq 0) 'all originals disposed'
    $rows=@($script:receipt.Rows|ForEach-Object {$_|ConvertFrom-Json})
    Check (-not $rows[-1].Finalized) 'persisted outcome precedes cleanup finalization'
    Check (@($rows|Where-Object RunId -CNE ('a'*32)).Count -eq 0 -and $rows[0].Observation -ceq 'OrdinaryAccountCenter') 'manifest and samples correlate'
    Check (($script:receipt.Rows -join '') -notmatch 'private-') 'sanitized shared failure receipts'
    if($failure){Check ($rows[-1].Kind -ceq 'outcome' -and $rows[-1].Outcome -ceq 'observation-failed' -and @($rows|Where-Object {$_.Kind -eq 'pair' -and $_.Incomplete}).Count -eq 1) 'failure preserves incomplete native pair'}
}
foreach($bad in @('trailing','extra')){
    Reset-Receipt;$script:events.Clear();$script:original.Disposed=$false;$script:managedFault=$false;$script:cimFault=$bad
    $script:backend=New-OrdinaryBackend $policy
    $script:backend.Publish={param($record)}
    Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $script:backend} "strict command line rejects $bad"
    Check (@($script:events|Where-Object {$_ -eq 'launch'}).Count -eq 1 -and $script:original.Disposed) 'identity failure single acquisition and cleanup'
}
$script:cimFault=''
Reset-Receipt;$script:events.Clear();$script:original.Disposed=$false
$script:backend=New-OrdinaryBackend $policy
$script:backend.ElapsedMs={[double]360000}
$script:backend.Publish={param($record)}
$result=Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $script:backend
Check ($result.Outcome -ceq 'timed-out' -and -not $result.ExitVerified -and $result.Finalized) 'exact deadline not verified exit'
Check (@($script:events|Where-Object {$_ -eq 'launch'}).Count -eq 1 -and -not $script:events.Contains('native') -and $script:original.Disposed) 'deadline no native sample or retry and cleanup'
Reset-Receipt;$script:events.Clear();$script:backend=New-OrdinaryBackend $policy;$script:backend.Remove('ReadNative')
Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $script:backend} 'missing callback rejected'
Check ($script:events.Count -eq 0 -and $script:receipt.Events.Count -eq 0) 'missing callback no live fallback'
Check ($null -eq ('Talenro.Diagnostics.NativeObservationV1.NativeProbe' -as [type])) 'native helper still absent'
Write-Output "PASS: $script:checks Ordinary synthetic assertions; no real launch/CIM/native/sleep."
