#requires -Version 7.0
$ErrorActionPreference='Stop'
$entry=Join-Path $PSScriptRoot 'observe-email-geometry-account-center.ps1'
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
Check (Test-Path -LiteralPath $entry) 'RED: EG entry missing'

# Only the four reviewed binding tokens may differ; all other safety behavior is inherited.
$originalText=[IO.File]::ReadAllText((Join-Path $PSScriptRoot 'observe-email-geometry.ps1'))
$candidateText=[IO.File]::ReadAllText($entry)
$bindingReplacements=@(
    @('clients/windows/Talenro.Demo.App/bin/email-geometry/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe','clients/windows/Talenro.Demo.App/bin/email-geometry-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe',2),
    @('CBDDE55CBFE2928E3F179250F9B0DFC560DB45D61FD05B8B014BADB0B0CFCDB3','CF7C577C2FF3F4A690B2EFEE054411B59006210440333C5E8EA1838A7062A48F',1),
    @('154E7E70F9FB2D2F980D16A920CA7A3CB69C5CA001CECF7D6A7DFFD1D1F043D9','6011FA7BB63A45D5FAEB4A44529904FEC0857741A4E810B20B38555655C838C1',1),
    @("'observe-email-geometry.ps1'","'observe-email-geometry-account-center.ps1'",1)
)
$expectedText=$originalText
foreach($replacement in $bindingReplacements){
    Check ([regex]::Matches($originalText,[regex]::Escape($replacement[0])).Count -eq $replacement[2]) 'original binding token count'
    Check (-not $originalText.Contains($replacement[1])) 'candidate token absent from baseline'
    $expectedText=$expectedText.Replace($replacement[0],$replacement[1])
}
function Assert-CandidateParity([string]$text){
    if($text.Replace("`r`n","`n").TrimEnd("`n") -cne $expectedText.Replace("`r`n","`n").TrimEnd("`n")){
        throw 'Candidate differs beyond fixed binding tokens.'
    }
}
Assert-CandidateParity $candidateText
Check $true 'exact candidate binding-only parity'
Reject {Assert-CandidateParity ($candidateText.Replace('ObservationCapMs=360000','ObservationCapMs=360001'))} 'parity rejects changed observation cap'
Reject {Assert-CandidateParity ($candidateText.Replace('    Assert-EGPreflight $policy','    # preflight removed'))} 'parity rejects removed preflight'

$result=@(& $entry)
Check ($result.Count -eq 1 -and $result[0].Status -ceq 'NotStarted') 'one no-start result'
foreach($parameter in @('Executable','Profile','Language','RunId','Deadline','Backend','ExeHash','DllHash')){
    Reject {& $entry @{$parameter='forbidden'}} "reject public $parameter override"
}
Check ($null -eq ('Talenro.Diagnostics.NativeObservationV1.NativeProbe' -as [type])) 'native helper absent'
$tokens=$null;$errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile($entry,[ref]$tokens,[ref]$errors)
Check ($errors.Count -eq 0) 'entry parses'
foreach($fn in $ast.FindAll({param($a)$a -is [Management.Automation.Language.FunctionDefinitionAst]},$false)){
    . ([scriptblock]::Create($fn.Extent.Text))
}
# Missing state validation/deduplication would leak payload or grow unboundedly.
$state=New-EGFailureState $null
Check ($null -eq $state.RunId) 'pre-ID state remains null'
foreach($bad in @('', 'abc', ('0'*32), ('A'*32))){Reject {New-EGFailureState $bad} 'invalid failure run id'}
$state=New-EGFailureState '11111111111141118111111111111111'
foreach($i in 1..1000){Add-EGFailure $state 'ManagedRead';Add-EGFailure $state 'ProcessDispose'}
Add-EGFailure $state 'NativeRead'
Reject {Add-EGFailure $state 'private-secret'} 'unknown failure token rejected'
$snapshot=Get-EGFailureSummary $state
Check (($snapshot.OperationFailures -join ',') -ceq 'ManagedRead,NativeRead') 'bounded ordered operation categories'
Check (($snapshot.CleanupFailures -join ',') -ceq 'ProcessDispose') 'bounded cleanup categories'
Check (@($snapshot.PSObject.Properties).Count -eq 7 -and $snapshot.Outcome -ceq 'Failed' -and $snapshot.Acceptance -ceq 'NotAccepted') 'exact failure summary'
$snapshot.OperationFailures[0]='changed'
Check ((Get-EGFailureSummary $state).OperationFailures[0] -ceq 'ManagedRead') 'detached summary'
Check ((New-EGFailureState ('b'*32)).OperationFailures.Count -eq 0) 'per-run isolation'
$allState=New-EGFailureState ('c'*32)
$allOperations='Setup,Preflight,Import,ReceiptDirectory,ManifestOpen,ManifestConstruct,ManifestWrite,ManifestFlush,ManifestAdapter,SamplesOpen,SamplesConstruct,SamplesWrite,SamplesFlush,SamplesAdapter,ProcessAcquire,ProcessAdapter,Identity,NativeRead,CimRead,ManagedRead,Clock,Sleep,Publication,SessionUnclassified,Context,Result,State'.Split(',')
$allCleanup='ManifestWriterDispose,ManifestStreamDispose,SamplesWriterDispose,SamplesStreamDispose,ProcessDispose'.Split(',')
foreach($token in ($allOperations+$allCleanup)){Add-EGFailure $allState $token;Add-EGFailure $allState $token}
Check ($allState.OperationFailures.Count -eq 27 -and $allState.CleanupFailures.Count -eq 5) 'finite allowlist covers all stages'
# A cleanup throw must not skip another resource or retry the same release.
function New-FaultResource($fail=$false){
    $r=[pscustomobject]@{Attempts=0;Fail=$fail;Writes=0;Flushes=0;Id=123}
    $r|Add-Member ScriptMethod Dispose {$this.Attempts++;if($this.Fail){throw 'private-dispose'}}
    $r|Add-Member ScriptMethod WriteLine {param($line)$this.Writes++;if($this.Fail){throw 'private-write'}}
    $r|Add-Member ScriptMethod Flush {$this.Flushes++;if($this.Fail){throw 'private-flush'}}
    $r
}
foreach($fail in @($false,$true)){
    $s=New-EGFailureState ('a'*32);$w=New-FaultResource $fail;$stream=New-FaultResource $fail
    $owner=New-EGWriterOwner $w $stream 'Samples' $s
    if($fail){Reject {$owner.WriteLine('test')} 'sample write recorded';Reject {$owner.Flush()} 'sample flush recorded';Reject {$owner.Dispose()} 'combined disposal fails'}
    else{$owner.WriteLine('test');$owner.Flush();$owner.Dispose()}
    $owner.Dispose()
    Check ($w.Attempts -eq 1 -and $stream.Attempts -eq 1) 'writer and stream attempted once'
    if($fail){Check (($s.CleanupFailures -join ',') -ceq 'SamplesWriterDispose,SamplesStreamDispose') 'both cleanup categories';Check (($s.OperationFailures -join ',') -ceq 'SamplesWrite,SamplesFlush') 'write flush categories'}
    $p=New-FaultResource $fail;$processOwner=New-EGProcessOwner $p $s
    Check ([object]::ReferenceEquals($p,$processOwner.Original) -and $processOwner.Id -eq 123) 'process adapter original identity'
    if($fail){Reject {$processOwner.Dispose()} 'process cleanup fails'}else{$processOwner.Dispose()}
    $processOwner.Dispose();Check ($p.Attempts -eq 1) 'process dispose once'
    if($fail){Check ($s.CleanupFailures -contains 'ProcessDispose') 'process cleanup retained'}
}
$policy=Get-EGPolicy 'D:\fixture' '11111111111141118111111111111111'

$expectedCandidate=Join-Path 'D:\fixture' 'clients/windows/Talenro.Demo.App/bin/email-geometry-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe'
Check ($policy.Executable -ceq $expectedCandidate) 'candidate exact executable path'
Check ($policy.Dll -ceq [IO.Path]::ChangeExtension($expectedCandidate,'.dll')) 'candidate DLL co-located'
$expectedCandidateDirectory=[IO.Path]::GetDirectoryName($expectedCandidate)
Check ($policy.ReceiptRoot -ceq (Join-Path $expectedCandidateDirectory 'email-geometry-observations')) 'candidate receipt root'
Check ($policy.GeometryRoot -ceq (Join-Path $expectedCandidateDirectory 'email-geometry-records')) 'candidate geometry root'
Check ($policy.GeometryFile -ceq (Join-Path $expectedCandidateDirectory 'email-geometry-records/11111111111141118111111111111111/geometry.jsonl')) 'candidate geometry file'

Check ($policy.Arguments -ceq '--layout-test --layout-language=en-US --layout-profile=rtl-mixed --layout-email-geometry-run=11111111111141118111111111111111') 'fixed EG arguments'
Check ($policy.ExpectedCommandLine -ceq ('"'+$policy.Executable+'" --layout-test --layout-language=en-US --layout-profile=rtl-mixed --layout-email-geometry-run=11111111111141118111111111111111')) 'independent command line'
Check ($policy.ExeHash -ceq 'CF7C577C2FF3F4A690B2EFEE054411B59006210440333C5E8EA1838A7062A48F') 'EXE identity'
Check ($policy.DllHash -ceq '6011FA7BB63A45D5FAEB4A44529904FEC0857741A4E810B20B38555655C838C1') 'DLL identity'
Check (($policy.Sources -join ',') -ceq 'observe-email-geometry-account-center.ps1,c5-observation-session.psm1,observation-session-guard.psm1,process-observation.psm1,native-process-observation.psm1,native-process-observation.cs') 'all consumed sources checked'
Remove-Item Function:Import-Module
Import-Module (Join-Path $PSScriptRoot 'c5-observation-session.psm1')
$script:fault='';$script:acquisitions=0;$script:inventory=0
function Assert-EGOrdinaryToken {if($script:fault -eq 'token'){throw 'blocked'}}
function Invoke-EGGit {param($Root,$Arguments)
    switch($Arguments[0]){
        'branch' {if($script:fault -eq 'branch'){'main'}else{'codex/c12-b01-task9-coordinator'}}
        'status' {if($script:fault -eq 'dirty'){' M changed'}}
        'rev-parse' {if($script:fault -eq 'source'){'different'}else{'blob'}}
        'hash-object' {'blob'}
        default {throw 'Unexpected git command'}
    }
}
function Test-Path {param($LiteralPath,$PathType,$ErrorAction)
    if($PathType -eq 'Leaf'){return $script:fault -ne 'missing'}
    if($script:fault -eq 'observer-collision'){return $LiteralPath -ceq (Join-Path $policy.ReceiptRoot $policy.RunId)}
    if($script:fault -eq 'geometry-collision'){return $LiteralPath -ceq (Join-Path $policy.GeometryRoot $policy.RunId)}
    if($script:fault -eq 'claim-collision'){return $LiteralPath -ceq (Join-Path $policy.GeometryRoot ($policy.RunId+'.claim'))}
    if($script:fault -eq 'geometry-reparse'){return $LiteralPath -ceq $policy.GeometryRoot}
    return $false
}
function Get-Item {param($LiteralPath,[switch]$Force,$ErrorAction)
    @{Attributes=[IO.FileAttributes]::ReparsePoint}
}
function Get-FileHash {param($LiteralPath,$Algorithm)
    if($script:fault -eq 'hash'){return @{Hash='wrong'}}
    if($script:fault -eq 'old-exe' -and $LiteralPath.EndsWith('.exe')){return @{Hash='CBDDE55CBFE2928E3F179250F9B0DFC560DB45D61FD05B8B014BADB0B0CFCDB3'}}
    if($script:fault -eq 'old-dll' -and $LiteralPath.EndsWith('.dll')){return @{Hash='154E7E70F9FB2D2F980D16A920CA7A3CB69C5CA001CECF7D6A7DFFD1D1F043D9'}}
    @{Hash=$(if($LiteralPath.EndsWith('.exe')){$policy.ExeHash}else{$policy.DllHash})}
}
function Get-CimInstance {param($ClassName,$Filter,$ErrorAction)
    $script:inventory++
    Check ($Filter -ceq "Name LIKE 'Talenro.Demo.%' OR Name LIKE '%Prototype%'") 'conservative inventory filter'
    if($script:fault -eq 'cim'){throw 'inaccessible'}
    if($script:fault -eq 'target'){@{ProcessId=12}}
}
$context=@{RunId='11111111111141118111111111111111';ExpectedPath=$policy.Executable;ExpectedCommandLine=$policy.ExpectedCommandLine}
$gate=@{
    Preflight={param($c) Assert-EGPreflight $policy}
    AcquireWriter={$script:acquisitions++;throw 'unexpected writer'}
    AcquireProcess={$script:acquisitions++;throw 'unexpected process'}
    GetIdentity={};ReadNative={};ReadManaged={};ReadCim={};ElapsedMs={0};SleepMs={};Publish={}
}
Assert-EGPreflight $policy
Check ($script:inventory -eq 1) 'complete preflight'
foreach($fault in @('token','branch','dirty','source','missing','hash','old-exe','old-dll','target','cim','observer-collision','geometry-collision','claim-collision','geometry-reparse')){
    $script:fault=$fault
    if($fault -eq 'geometry-reparse'){
        $geometryError=$null
        try{Assert-EGPreflight $policy}catch{$geometryError=$_.Exception.Message}
        Check ($geometryError -ceq 'Reparse receipt path rejected.') 'geometry preflight inspects reparse attributes'
    }
    Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $gate} "preflight rejects $fault"
    Check ($script:acquisitions -eq 0) "no acquisitions after $fault"
}
$script:fault=''
Check ($null -ne (Get-Command New-EGReceiptWriter -ErrorAction SilentlyContinue)) 'RED: receipt factory missing'
# A removed path check must expose collision/reparse destinations before any write.
function Test-Path {param($LiteralPath,$PathType,$ErrorAction) $LiteralPath -in $script:existing}
function Get-Item {param($LiteralPath,[switch]$Force,$ErrorAction)
    @{Attributes=$(if($LiteralPath -eq $script:reparse){[IO.FileAttributes]::ReparsePoint}else{[IO.FileAttributes]::Directory})}
}
$destination=Join-Path $policy.ReceiptRoot $context.RunId
$script:existing=@();$script:reparse=''
Assert-EGReceiptDestination $destination
Check $true 'new destination accepted'
foreach($kind in @('directory','file')){
    $script:existing=@($destination)
    Reject {Assert-EGReceiptDestination $destination} "existing $kind rejected"
}
$script:existing=@($policy.ReceiptRoot);$script:reparse=$policy.ReceiptRoot
$reparseError=$null
try{Assert-EGReceiptDestination $destination}catch{$reparseError=$_.Exception.Message}
Check ($reparseError -ceq 'Reparse receipt path rejected.') 'reparse rejected by attributes, not binding failure'
$script:reparse=''
Assert-EGReceiptDestination $destination
Check $true 'ordinary existing parent accepted'
$script:existing=@();$script:reparse=''
# Capture real factory and encoder; wrappers below isolate all disk access.
$script:factory=Get-Command New-EGReceiptWriter
$encoder=Get-Command New-EGTextWriter
$memory=[IO.MemoryStream]::new();$encoded=& $encoder $memory 'Samples' (New-EGFailureState $context.RunId)
$encoded.WriteLine('EG')
Check ($memory.Length -eq 4 -and $memory.ToArray()[0] -eq 69) 'UTF8 no BOM and immediate flush'
$encoded.Dispose()
Check $memory.CanWrite 'writer leaves stream for its explicit owner'
$memory.Dispose()
# Real constructor wrapper must retain configuration and its own cleanup failure.
$script:constructorWriter=New-FaultResource $true
$script:constructorWriter|Add-Member ScriptProperty AutoFlush {$false} {param($value)throw 'private-autoflush'}
function New-EGStreamWriter {param($Stream)$script:constructorWriter}
$constructionState=New-EGFailureState $context.RunId
$memory=[IO.MemoryStream]::new()
Reject {& $encoder $memory 'Manifest' $constructionState} 'autoflush failure rejected'
Check (($constructionState.OperationFailures -join ',') -ceq 'ManifestConstruct' -and ($constructionState.CleanupFailures -join ',') -ceq 'ManifestWriterDispose') 'constructor primary and cleanup retained'
Check ($script:constructorWriter.Attempts -eq 1 -and $memory.CanWrite) 'constructor leaves stream caller owned'
$memory.Dispose()
Remove-Item Function:New-EGStreamWriter
function Reset-Receipt($fault=''){
    $script:failureState=New-EGFailureState $context.RunId
    $script:receipt=@{Fault=$fault;Events=[Collections.Generic.List[string]]::new();Resources=[Collections.Generic.List[object]]::new();Rows=[Collections.Generic.List[string]]::new();Launches=0}
}
function Assert-EGReceiptDestination {param($Directory)
    $script:receipt.Events.Add('check');if($script:receipt.Fault -eq 'collision'){throw 'private-collision'}
}
function New-EGReceiptDirectory {param($Directory)
    $script:receipt.Events.Add('directory');if($script:receipt.Fault -eq 'directory'){throw 'private-directory'}
}
function Open-EGNewStream {param($Path)
    $kind=if($Path.EndsWith('manifest.json')){'manifest'}elseif($Path.EndsWith('samples.jsonl')){'samples'}else{throw 'wrong receipt path'}
    $script:receipt.Events.Add("$kind-open")
    if($script:receipt.Fault -eq "$kind-open"){throw 'private-open'}
    $s=[pscustomobject]@{Kind=$kind;State=$script:receipt;Disposed=$false;Attempts=0;Memory=[IO.MemoryStream]::new()}
    $s|Add-Member ScriptMethod Dispose {$this.Attempts++;$this.Disposed=$true;$this.State.Events.Add($this.Kind+'-stream-dispose');$this.Memory.Dispose();if($this.State.Fault -contains ($this.Kind+'-stream-dispose')){throw 'private-stream-dispose'}}
    $script:receipt.Resources.Add($s);$s
}
function New-EGTextWriter {param($Stream)
    if($script:receipt.Fault -eq ($Stream.Kind+'-writer')){throw 'private-writer'}
    $w=[pscustomobject]@{Kind=$Stream.Kind;State=$script:receipt;Stream=$Stream;Disposed=$false;Attempts=0}
    $w|Add-Member ScriptMethod WriteLine {param($line)
        $this.State.Events.Add($this.Kind+'-write')
        if($this.State.Fault -eq ($this.Kind+'-write')){throw 'private-write'}
        $this.State.Rows.Add($line);$this.Stream.Memory.WriteByte(1)
    }
    $w|Add-Member ScriptMethod Flush {
        $this.State.Events.Add($this.Kind+'-flush');if($this.State.Fault -eq ($this.Kind+'-flush')){throw 'private-flush'}
    }
    $w|Add-Member ScriptMethod Dispose {
        $this.Attempts++;$this.Disposed=$true;$this.State.Events.Add($this.Kind+'-writer-dispose')
        if($this.State.Fault -eq ($this.Kind+'-dispose')){throw 'private-dispose'}
        # leaveOpen=true: production owner disposes the stream separately.
    }
    $script:receipt.Resources.Add($w);$w
}
Reset-Receipt
$owned=@(& $script:factory $policy $context $script:failureState)
Check ($owned.Count -eq 1 -and $owned[0].Kind -eq 'samples') 'exactly one transferred writer'
Check (($script:receipt.Events -join ',') -ceq 'check,check,check,directory,manifest-open,manifest-write,manifest-flush,manifest-writer-dispose,manifest-stream-dispose,samples-open') 'manifest closed before sample acquisition'
$manifest=$script:receipt.Rows[0]|ConvertFrom-Json
$expectedManifest=@{SchemaVersion=1;Observation='EmailGeometry';RunId='11111111111141118111111111111111';Language='en-US';Profile='rtl-mixed';ExecutableRelativePath='clients/windows/Talenro.Demo.App/bin/email-geometry-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe';Arguments='--layout-test --layout-language=en-US --layout-profile=rtl-mixed --layout-email-geometry-run=11111111111141118111111111111111';ExeSha256='CF7C577C2FF3F4A690B2EFEE054411B59006210440333C5E8EA1838A7062A48F';DllSha256='6011FA7BB63A45D5FAEB4A44529904FEC0857741A4E810B20B38555655C838C1';GeometryRelativePath='email-geometry-records/11111111111141118111111111111111/geometry.jsonl';ObservationCapMs=360000;MaxSleepMs=2000}
Check (@($manifest.PSObject.Properties).Count -eq 12) 'manifest no private extras'
foreach($key in $expectedManifest.Keys){Check ($manifest.$key -ceq $expectedManifest[$key]) "manifest $key"}
Check (-not $owned[0].WriterAttempted) 'sample ownership transferred alive'
$owned[0].Dispose()
$gate.Preflight={}
$gate.AcquireWriter={param($c) & $script:factory $policy $c $script:failureState}
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
    Reject {& $script:factory $policy $invalid $script:failureState} 'invalid receipt id'
    Check ($script:receipt.Events.Count -eq 0) 'invalid id no filesystem effects'
}
foreach($key in @('ExpectedPath','ExpectedCommandLine')){
    Reset-Receipt;$invalid=$context.Clone();$invalid[$key]='wrong'
    Reject {& $script:factory $policy $invalid $script:failureState} "mismatched $key"
    Check ($script:receipt.Events.Count -eq 0) 'context mismatch no filesystem effects'
}
# Factory transfer fails after both raw resources exist: neither may be orphaned.
$realWriterOwner=(Get-Command New-EGWriterOwner).ScriptBlock
function New-EGWriterOwner {param($Writer,$Stream,$Kind,$State)
    if($Kind -ceq $script:adapterFault){throw 'private-adapter'}
    & $realWriterOwner $Writer $Stream $Kind $State
}
foreach($kind in @('Manifest','Samples')){
    $script:adapterFault=$kind
    Reset-Receipt @(($kind.ToLower()+'-dispose'),($kind.ToLower()+'-stream-dispose'))
    Reject {& $script:factory $policy $context $script:failureState} 'partial factory adapter failure'
    Check ($script:failureState.OperationFailures -contains ($kind+'Adapter')) 'adapter acquisition category'
    Check ($script:failureState.CleanupFailures -contains ($kind+'WriterDispose') -and $script:failureState.CleanupFailures -contains ($kind+'StreamDispose')) 'partial acquisition cleanup categories'
    Check (@($script:receipt.Resources|Where-Object Attempts -NE 1).Count -eq 0) 'factory cleans each acquired resource once'
}
$script:adapterFault=''
foreach($case in @(@('manifest-write','manifest-dispose','manifest-stream-dispose'),@('manifest-flush','manifest-dispose'),@('samples-writer','samples-stream-dispose'))){
    Reset-Receipt $case
    Reject {& $script:factory $policy $context $script:failureState} 'combined factory failure'
    Check ($script:failureState.OperationFailures.Count -gt 0 -and $script:failureState.CleanupFailures.Count -gt 0) 'factory preserves operation plus cleanup'
    Check (@($script:receipt.Resources|Where-Object Attempts -NE 1).Count -eq 0) 'combined cleanup attempted once'
}
Check ($null -ne (Get-Command New-EGBackend -ErrorAction SilentlyContinue)) 'RED: backend missing'
Import-Module (Join-Path $PSScriptRoot 'native-process-observation.psm1')
Import-Module (Join-Path $PSScriptRoot 'process-observation.psm1')
$script:events=[Collections.Generic.List[string]]::new()
$script:start=[datetime]::Parse('2026-11-09T18:39:47.4427247Z').ToUniversalTime()
$script:original=[pscustomobject]@{Id=123;StartTime=$script:start;SafeHandle=[object]::new();Disposed=$false;DisposeFault=$false;Attempts=0}
$script:original|Add-Member ScriptMethod Dispose {$this.Attempts++;$this.Disposed=$true;if($this.DisposeFault){throw 'private-process-dispose'}}
$script:cimFault='';$script:managedFault=$false
function Assert-EGPreflight {param($Policy) $script:events.Add('preflight')}
function Start-Process {param($FilePath,$ArgumentList,$WorkingDirectory,[switch]$PassThru,$ErrorAction)
    Check ($FilePath -ceq $policy.Executable -and $ArgumentList -ceq '--layout-test --layout-language=en-US --layout-profile=rtl-mixed --layout-email-geometry-run=11111111111141118111111111111111' -and $PassThru) 'fixed launch arguments'
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
$script:backend=New-EGBackend $policy $script:failureState
Check ((($script:backend.Keys|Sort-Object) -join ',') -ceq 'AcquireProcess,AcquireWriter,ElapsedMs,GetIdentity,Preflight,Publish,ReadCim,ReadManaged,ReadNative,SleepMs') 'exact ten callbacks'
Check ((& $script:backend.ElapsedMs) -eq 0) 'clock not running before acquisition'
$process=& $script:backend.AcquireProcess $context
Check ([object]::ReferenceEquals($process.Original,$script:original)) 'adapter retains original process'
$identity=& $script:backend.GetIdentity $process
Check ([object]::ReferenceEquals($identity.Handle,$script:original.SafeHandle) -and $identity.StartUtc.Kind -eq 'Utc') 'identity original handle and UTC'
foreach($bad in @('zero','multiple','missing','unspecified')){
    $script:cimFault=$bad;Reject {& $script:backend.ReadCim 123} "CIM rejects $bad"
}
$script:cimFault='local';$row=& $script:backend.ReadCim 123
Check ($row.CreationUtc.Kind -eq 'Utc' -and $row.CreationUtc -eq $script:start.AddTicks(-7)) 'local CIM converted UTC'
& $script:backend.SleepMs 7
$script:cimFault=''
# Real EG factory + real ten-callback wiring + shared orchestration; only OS effects replaced.
foreach($failure in @($false,$true)){
    Reset-Receipt;$script:events.Clear();$script:original.Disposed=$false;$script:managedFault=$failure
    $script:backend=New-EGBackend $policy $script:failureState
    $progress=[Collections.Generic.List[object]]::new()
    $script:backend.Publish={param($r)$progress.Add($r)}.GetNewClosure()
    if($failure){Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $script:backend} 'managed failure retained'}
    else{$result=Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $script:backend;Check ($result.ExitVerified -and $result.Finalized) 'EG finalized synthetic exit'}
    Check (($script:events -join ',') -ceq 'preflight,launch,native,native,managed') 'native first single launch'
    Check ($script:original.Disposed -and @($script:receipt.Resources|Where-Object {-not $_.Disposed}).Count -eq 0) 'all originals disposed'
    $rows=@($script:receipt.Rows|ForEach-Object {$_|ConvertFrom-Json})
    Check (@($rows|Where-Object RunId -CNE '11111111111141118111111111111111').Count -eq 0 -and $rows[0].Observation -ceq 'EmailGeometry') 'manifest and samples correlate'
    Check (($script:receipt.Rows -join '') -notmatch 'private-') 'sanitized shared failure receipts'
    if($failure){Check ($rows[-1].Kind -ceq 'outcome' -and $rows[-1].Outcome -ceq 'observation-failed' -and @($rows|Where-Object {$_.Kind -eq 'pair' -and $_.Incomplete}).Count -eq 1) 'failure preserves incomplete native pair'}
}
# End-to-end boundary must preserve concurrent operation/cleanup failures, not just throw.
Reset-Receipt 'manifest-open';$script:events.Clear()
$script:backend=New-EGBackend $policy $script:failureState;$script:backend.Publish={param($r)}
$caught=$null
try{Invoke-EGSession $policy $context $script:failureState $script:backend Synthetic 6>$null|Out-Null}catch{$caught=$_.Exception}
Check (($caught.Data['EGFailureSummary'].OperationFailures -join ',') -ceq 'ManifestOpen,SessionUnclassified') 'factory failure not mislabelled as directory failure'
Check ($script:events -notcontains 'launch') 'writer failure prevents launch'
Reset-Receipt @('samples-dispose','samples-stream-dispose')
$script:managedFault=$true;$script:original.DisposeFault=$true;$script:original.Attempts=0
$script:backend=New-EGBackend $policy $script:failureState
$script:backend.Publish={param($r)}
$caught=$null;$info=@()
try{Invoke-EGSession $policy $context $script:failureState $script:backend Synthetic -InformationVariable info|Out-Null}catch{$caught=$_.Exception}
Check ($null -ne $caught.Data['EGFailureSummary']) 'safe final summary survives shared catch'
$summary=$caught.Data['EGFailureSummary']
Check ($summary.OperationFailures -contains 'ManagedRead' -and $summary.OperationFailures -contains 'SessionUnclassified') 'managed and opaque session categories'
Check (($summary.CleanupFailures -join ',') -ceq 'SamplesWriterDispose,SamplesStreamDispose,ProcessDispose') 'all final cleanup categories'
Check ($caught.InnerException -eq $null -and ($summary|ConvertTo-Json -Depth 5) -notmatch 'private-') 'no raw final payload'
Check ($script:original.Attempts -eq 1) 'final process cleanup once'
Check ($summary.Acceptance -ceq 'NotAccepted') 'failed summary never accepts'
Check (@($info).Count -eq 1) 'one final information summary'
$script:original.DisposeFault=$false;$script:managedFault=$false
Reset-Receipt;$script:backend=New-EGBackend $policy $script:failureState;$script:backend.Publish={param($r)}
$success=Invoke-EGSession $policy $context $script:failureState $script:backend Synthetic
Check ($success.ProcessObservation.ExitVerified -and $success.ProcessObservation.Finalized -and $success.Acceptance -ceq 'NotAccepted' -and $success.GeometryEvidenceStatus -ceq 'Unreviewed') 'clean wrapper result remains not accepted'
# Shared internal validation cannot be inferred from callback exception strings.
Reset-Receipt;$script:backend=New-EGBackend $policy $script:failureState
$script:backend.ReadCim={param($pidValue)@{ProcessId=999;ExecutablePath='wrong';CommandLine='wrong';CreationUtc=[datetime]::UtcNow}}
$script:backend.Publish={param($r)}
$caught=$null
try{Invoke-EGSession $policy $context $script:failureState $script:backend Synthetic 6>$null|Out-Null}catch{$caught=$_.Exception}
Check (($caught.Data['EGFailureSummary'].OperationFailures -join ',') -ceq 'SessionUnclassified') 'internal validation explicitly unclassified'
# Failure information publication may itself fail, but must not retry or leak.
$identityMisses=[Collections.Generic.List[string]]::new()
foreach($propertyName in @('Id','StartTime','SafeHandle')){
    $savedValue=$script:original.PSObject.Properties[$propertyName].get_Value()
    $script:original|Add-Member -MemberType ScriptProperty -Name $propertyName -Value {throw 'private-identity-getter'} -Force
    Reset-Receipt;$script:original.Attempts=0
    $script:backend=New-EGBackend $policy $script:failureState;$script:backend.Publish={param($r)}
    $caught=$null
    try{Invoke-EGSession $policy $context $script:failureState $script:backend Synthetic 6>$null|Out-Null}catch{$caught=$_.Exception}
    finally{$script:original|Add-Member -MemberType NoteProperty -Name $propertyName -Value $savedValue -Force}
    if($null -eq $caught -or $caught.Data['EGFailureSummary'].OperationFailures -notcontains 'Identity'){$identityMisses.Add($propertyName)}
    Check ($script:original.Attempts -eq 1) "identity getter cleanup $propertyName"
}
Check ($identityMisses.Count -eq 0) ('identity getter categories retained: '+($identityMisses -join ','))
$script:publications=0
function Write-Information {param($MessageData,$InformationAction,$ErrorAction)$script:publications++;throw 'private-publication'}
$caught=$null
try{Stop-EGFailure $script:failureState}catch{$caught=$_.Exception}
Check ($script:publications -eq 1 -and $caught.Data['EGFailureSummary'].OperationFailures -contains 'Publication') 'publication failure terminal without retry'
Remove-Item Function:Write-Information
# A nominal successful shared result cannot override an already recorded failure.
Reset-Receipt;Add-EGFailure $script:failureState 'State';$script:backend=New-EGBackend $policy $script:failureState
$script:backend.Publish={param($r)}
$caught=$null
try{Invoke-EGSession $policy $context $script:failureState $script:backend Synthetic 6>$null|Out-Null}catch{$caught=$_.Exception}
Check ($caught.Data['EGFailureSummary'].OperationFailures -contains 'State') 'failure state forbids success'
$realProcessOwner=(Get-Command New-EGProcessOwner).ScriptBlock
function New-EGProcessOwner {param($Process,$State)throw 'private-process-adapter'}
Reset-Receipt;$script:original.Attempts=0;$script:original.DisposeFault=$true
$script:backend=New-EGBackend $policy $script:failureState;$script:backend.Publish={param($r)}
$caught=$null
try{Invoke-EGSession $policy $context $script:failureState $script:backend Synthetic 6>$null|Out-Null}catch{$caught=$_.Exception}
Check ($script:original.Attempts -eq 1 -and $caught.Data['EGFailureSummary'].OperationFailures -contains 'ProcessAdapter' -and $caught.Data['EGFailureSummary'].CleanupFailures -contains 'ProcessDispose') 'process adapter failure retains cleanup'
function New-EGProcessOwner {param($Process,$State)& $realProcessOwner $Process $State}
$script:original.DisposeFault=$false
Reset-Receipt;$script:events.Clear();$script:failureState.RunId=('b'*32)
$script:backend=New-EGBackend $policy $script:failureState
$caught=$null
try{Invoke-EGSession $policy $context $script:failureState $script:backend Synthetic 6>$null|Out-Null}catch{$caught=$_.Exception}
Check ($script:events.Count -eq 0 -and $script:receipt.Resources.Count -eq 0 -and $caught.Data['EGFailureSummary'].OperationFailures -contains 'Context') 'mismatched state no acquisition'
Reset-Receipt;$script:events.Clear();$script:backend=New-EGBackend $policy $script:failureState;$script:backend.Remove('ReadNative')
Reject {Invoke-TalenroC5Observation -Mode Synthetic -Context $context -Backend $script:backend} 'missing callback rejected'
Check ($script:events.Count -eq 0 -and $script:receipt.Events.Count -eq 0) 'missing callback no live fallback'
Check ($null -eq ('Talenro.Diagnostics.NativeObservationV1.NativeProbe' -as [type])) 'native helper still absent'

# Policy IDs and result correlation have no file or process effects.
foreach($bad in @('', 'abc', '../escape', ('0'*32), ('A'*32))){
    Reject {Get-EGPolicy 'D:\fixture' $bad} 'invalid policy run ID'
}
$wrapped=ConvertTo-EGResult $policy $context @{RunId=$context.RunId;ExitVerified=$true;Finalized=$true}
Check ($wrapped.GeometryEvidenceStatus -ceq 'Unreviewed' -and $wrapped.Acceptance -ceq 'NotAccepted') 'exit0 never accepts geometry'
Check ($wrapped.RunId -ceq '11111111111141118111111111111111' -and $wrapped.GeometryRelativePath -ceq 'email-geometry-records/11111111111141118111111111111111/geometry.jsonl') 'wrapper fixed app correlation'
Reject {ConvertTo-EGResult $policy $context @{RunId=('b'*32)}} 'unrelated observation rejected'
foreach($key in @('RunId','ExpectedPath','ExpectedCommandLine')){
    $invalid=$context.Clone();$invalid[$key]='wrong'
    Reject {ConvertTo-EGResult $policy $invalid @{RunId=$context.RunId}} "result context $key"
    Reject {& $script:backend.AcquireProcess $invalid} "launch context $key"
}
Write-Output "PASS: $script:checks account-center EG synthetic assertions; no real launch/CIM/native/sleep."
