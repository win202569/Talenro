#requires -Version 7.0
[CmdletBinding()]
param([switch]$RunApproved)

# Approval of implementation does not authorize a live run.
if(-not $RunApproved){[pscustomobject]@{Status='NotStarted';Reason='Separate bounded approval required'};return}
$ErrorActionPreference='Stop'

function New-EGFailureState([AllowNull()]$RunId){
    if($null -ne $RunId -and ($RunId -isnot [string] -or $RunId -cnotmatch '^[0-9a-f]{32}$' -or $RunId -eq ('0'*32))){throw 'Invalid failure run ID.'}
    @{RunId=$RunId;OperationFailures=[Collections.Generic.List[string]]::new();CleanupFailures=[Collections.Generic.List[string]]::new()}
}
function Add-EGFailure($State,[string]$Category){
    $operations=@('Setup','Preflight','Import','ReceiptDirectory','ManifestOpen','ManifestConstruct','ManifestWrite','ManifestFlush','ManifestAdapter',
        'SamplesOpen','SamplesConstruct','SamplesWrite','SamplesFlush','SamplesAdapter','ProcessAcquire','ProcessAdapter','Identity','NativeRead',
        'CimRead','ManagedRead','Clock','Sleep','Publication','SessionUnclassified','Context','Result','State')
    $cleanup=@('ManifestWriterDispose','ManifestStreamDispose','SamplesWriterDispose','SamplesStreamDispose','ProcessDispose')
    if($Category -cin $operations){$list=$State.OperationFailures}
    elseif($Category -cin $cleanup){$list=$State.CleanupFailures}
    else{throw 'Unknown failure category.'}
    if(-not $list.Contains($Category)){$list.Add($Category)}
}
function Get-EGFailureSummary($State){
    [pscustomobject]@{SchemaVersion=1;RunId=$State.RunId;Outcome='Failed'
        OperationFailures=[string[]]$State.OperationFailures.ToArray();CleanupFailures=[string[]]$State.CleanupFailures.ToArray()
        GeometryEvidenceStatus='Unreviewed';Acceptance='NotAccepted'}
}

function New-EGWriterOwner($Writer,$Stream,[ValidateSet('Manifest','Samples')][string]$Kind,$State){
    $owner=[pscustomobject]@{Writer=$Writer;Stream=$Stream;Kind=$Kind;State=$State;WriterAttempted=$false;StreamAttempted=$false;Record=(Get-Command Add-EGFailure)}
    $owner|Add-Member ScriptMethod WriteLine {param([string]$line)
        try{$this.Writer.WriteLine($line)}catch{& $this.Record $this.State ($this.Kind+'Write');throw}
    }
    $owner|Add-Member ScriptMethod Flush {
        try{$this.Writer.Flush()}catch{& $this.Record $this.State ($this.Kind+'Flush');throw}
    }
    $owner|Add-Member ScriptMethod Dispose {
        $errors=[Collections.Generic.List[Exception]]::new()
        if(-not $this.WriterAttempted){
            $this.WriterAttempted=$true
            try{$this.Writer.Dispose()}catch{& $this.Record $this.State ($this.Kind+'WriterDispose');$errors.Add($_.Exception)}
        }
        if(-not $this.StreamAttempted){
            $this.StreamAttempted=$true
            try{$this.Stream.Dispose()}catch{& $this.Record $this.State ($this.Kind+'StreamDispose');$errors.Add($_.Exception)}
        }
        if($errors.Count){throw [AggregateException]::new('EG writer cleanup failed.',$errors)}
    }
    $owner
}
function New-EGProcessOwner($Process,$State){
    $owner=[pscustomobject]@{Original=$Process;State=$State;Attempted=$false;Record=(Get-Command Add-EGFailure)}
    $owner|Add-Member ScriptProperty Id {try{$this.Original.Id}catch{& $this.Record $this.State 'Identity';throw}}
    $owner|Add-Member ScriptMethod Dispose {
        if(-not $this.Attempted){
            $this.Attempted=$true
            try{$this.Original.Dispose()}catch{& $this.Record $this.State 'ProcessDispose';throw}
        }
    }
    $owner
}
function Get-EGPolicy([string]$Root,[string]$RunId){
    if($RunId -cnotmatch '^[0-9a-f]{32}$' -or $RunId -eq ('0'*32)){throw 'Invalid run ID.'}
    $exe=Join-Path $Root 'clients/windows/Talenro.Demo.App/bin/email-geometry/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe'
    $arguments='--layout-test --layout-language=en-US --layout-profile=rtl-mixed --layout-email-geometry-run='+$RunId
    @{
        Root=$Root;RunId=$RunId;Executable=$exe;Dll=[IO.Path]::ChangeExtension($exe,'.dll');Arguments=$arguments
        ExpectedCommandLine='"'+$exe+'" '+$arguments
        ExeHash='CBDDE55CBFE2928E3F179250F9B0DFC560DB45D61FD05B8B014BADB0B0CFCDB3'
        DllHash='154E7E70F9FB2D2F980D16A920CA7A3CB69C5CA001CECF7D6A7DFFD1D1F043D9'
        ReceiptRoot=Join-Path ([IO.Path]::GetDirectoryName($exe)) 'email-geometry-observations'
        GeometryRoot=Join-Path ([IO.Path]::GetDirectoryName($exe)) 'email-geometry-records'
        GeometryFile=Join-Path ([IO.Path]::GetDirectoryName($exe)) ('email-geometry-records/'+$RunId+'/geometry.jsonl')
        GeometryRelativePath='email-geometry-records/'+$RunId+'/geometry.jsonl'
        Sources=@('observe-email-geometry.ps1','c5-observation-session.psm1','observation-session-guard.psm1',
            'process-observation.psm1','native-process-observation.psm1','native-process-observation.cs')
    }
}
function Assert-EGOrdinaryToken {
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    try{
        $principal=[Security.Principal.WindowsPrincipal]::new($identity)
        if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary token required.'}
    }finally{$identity.Dispose()}
}
function Invoke-EGGit([string]$Root,[string[]]$Arguments){
    $output=& git -c "safe.directory=$Root" -C $Root @Arguments 2>$null
    if($LASTEXITCODE -ne 0){throw 'Git evidence unavailable.'}
    $output
}
function Assert-EGPreflight($Policy){
    Assert-EGOrdinaryToken
    if((Invoke-EGGit $Policy.Root @('branch','--show-current')) -cne 'codex/c12-b01-task9-coordinator'){throw 'Unexpected branch.'}
    if(@(Invoke-EGGit $Policy.Root @('status','--porcelain','--untracked-files=all')).Count -ne 0){throw 'Clean worktree required.'}
    foreach($source in $Policy.Sources){
        $relative='clients/windows/tests/'+$source
        $expected=Invoke-EGGit $Policy.Root @('rev-parse',('HEAD:'+$relative))
        $actual=Invoke-EGGit $Policy.Root @('hash-object',('--path='+$relative),(Join-Path $Policy.Root $relative))
        if([string]::IsNullOrWhiteSpace($expected) -or $expected -cne $actual){throw 'Observer source mismatch.'}
    }
    foreach($pair in @(@($Policy.Executable,$Policy.ExeHash),@($Policy.Dll,$Policy.DllHash))){
        if(-not(Test-Path -LiteralPath $pair[0] -PathType Leaf) -or
            (Get-FileHash -LiteralPath $pair[0] -Algorithm SHA256).Hash -cne $pair[1]){throw 'Preserved artifact unavailable or changed.'}
    }
    Assert-EGReceiptDestination (Join-Path $Policy.ReceiptRoot $Policy.RunId)
    Assert-EGGeometryDestination $Policy
    $targets=@(Get-CimInstance -ClassName Win32_Process -Filter "Name LIKE 'Talenro.Demo.%' OR Name LIKE '%Prototype%'" -ErrorAction Stop)
    if($targets.Count -ne 0){throw 'Existing target blocks launch.'}
}
function Assert-EGReceiptDestination([string]$Directory){
    if(Test-Path -LiteralPath $Directory -ErrorAction Stop){throw 'Receipt destination exists.'}
    # Trusted workspace only: checks are not a hostile filesystem-race guarantee.
    $ancestor=[IO.DirectoryInfo]::new($Directory).Parent
    while($null -ne $ancestor){
        if(Test-Path -LiteralPath $ancestor.FullName -ErrorAction Stop){
            $item=Get-Item -LiteralPath $ancestor.FullName -Force -ErrorAction Stop
            if(($item.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0){throw 'Reparse receipt path rejected.'}
        }
        $ancestor=$ancestor.Parent
    }
}
function New-EGReceiptDirectory([string]$Directory){
    $null=New-Item -ItemType Directory -Path $Directory -ErrorAction Stop
}
function Open-EGNewStream([string]$Path){
    [IO.FileStream]::new($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
}
function New-EGStreamWriter($Stream){
    [IO.StreamWriter]::new($Stream,[Text.UTF8Encoding]::new($false),1024,$true)
}
function New-EGTextWriter($Stream,[string]$Kind,$State){
    $writer=$null
    try{
        $writer=New-EGStreamWriter $Stream
        $writer.AutoFlush=$true
        return $writer
    }catch{
        $primary=$_.Exception
        Add-EGFailure $State ($Kind+'Construct')
        if($null -ne $writer){try{$writer.Dispose()}catch{Add-EGFailure $State ($Kind+'WriterDispose')}}
        # Caller still owns and disposes the stream if construction fails.
        throw $primary
    }
}
function Assert-EGContext($Policy,$Context){
    if($Context -isnot [Collections.IDictionary] -or $Context.Count -ne 3 -or
        $Context.RunId -isnot [string] -or $Context.RunId -cne $Policy.RunId -or
        $Context.ExpectedPath -cne $Policy.Executable -or
        $Context.ExpectedCommandLine -cne $Policy.ExpectedCommandLine){throw 'Observation context mismatch.'}
}
function Assert-EGGeometryDestination($Policy){
    Assert-EGReceiptDestination (Join-Path $Policy.GeometryRoot $Policy.RunId)
    Assert-EGReceiptDestination (Join-Path $Policy.GeometryRoot ($Policy.RunId+'.claim'))
}
function ConvertTo-EGResult($Policy,$Context,$Observation){
    Assert-EGContext $Policy $Context
    if($null -eq $Observation -or $Observation.RunId -cne $Context.RunId){throw 'Observation result mismatch.'}
    [pscustomobject]@{ProcessObservation=$Observation;GeometryEvidenceStatus='Unreviewed'
        Acceptance='NotAccepted';RunId=$Context.RunId;GeometryRelativePath=$Policy.GeometryRelativePath}
}
function New-EGReceiptWriter($Policy,$Context,$State){
    try{
        Assert-EGContext $Policy $Context
        if($State.RunId -cne $Policy.RunId){throw 'Failure state mismatch.'}
    }catch{Add-EGFailure $State 'Context';throw}
    try{
        Assert-EGGeometryDestination $Policy
        $directory=Join-Path $Policy.ReceiptRoot $Context.RunId
        Assert-EGReceiptDestination $directory
    }catch{Add-EGFailure $State 'Preflight';throw}
    try{New-EGReceiptDirectory $directory}catch{Add-EGFailure $State 'ReceiptDirectory';throw}
        $manifest=[ordered]@{
            SchemaVersion=1;Observation='EmailGeometry';RunId=$Context.RunId;Language='en-US';Profile='rtl-mixed'
            ExecutableRelativePath='clients/windows/Talenro.Demo.App/bin/email-geometry/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe'
            Arguments=$Policy.Arguments;ExeSha256=$Policy.ExeHash;DllSha256=$Policy.DllHash
            GeometryRelativePath=$Policy.GeometryRelativePath;ObservationCapMs=360000;MaxSleepMs=2000
        }

    $json=$manifest|ConvertTo-Json -Depth 4 -Compress
    foreach($kind in @('Manifest','Samples')){
        $stream=$null;$writer=$null;$owner=$null
        $stage=$kind+'Open'
        try{
            $name=if($kind -ceq 'Manifest'){'manifest.json'}else{'samples.jsonl'}
            $stream=Open-EGNewStream (Join-Path $directory $name)
            $stage=$kind+'Construct'
            $writer=New-EGTextWriter $stream $kind $State
            $stage=$kind+'Adapter'
            $owner=New-EGWriterOwner $writer $stream $kind $State
            # Ownership transfers only after the complete adapter exists.
            $writer=$null;$stream=$null;$stage=$null
            if($kind -ceq 'Samples'){return $owner}
            $owner.WriteLine($json)
            $owner.Flush()
            $owner.Dispose()
            $owner=$null
        }catch{
            $primary=$_.Exception
            if($null -ne $stage){Add-EGFailure $State $stage}
            if($null -ne $owner){try{$owner.Dispose()}catch{}}
            if($null -ne $writer){try{$writer.Dispose()}catch{Add-EGFailure $State ($kind+'WriterDispose')}}
            if($null -ne $stream){try{$stream.Dispose()}catch{Add-EGFailure $State ($kind+'StreamDispose')}}
            throw $primary
        }
    }
}
function New-EGBackend($Policy,$State){
    $validate=Get-Command Assert-EGContext
    $geometryDestination=Get-Command Assert-EGGeometryDestination
    $preflight=Get-Command Assert-EGPreflight
    $newWriter=Get-Command New-EGReceiptWriter
    $newProcessOwner=Get-Command New-EGProcessOwner
    $record=Get-Command Add-EGFailure
    $launch=Get-Command Start-Process
    $native=Get-Command Get-TalenroNativeProcessObservation
    $managed=Get-Command Get-TalenroProcessObservation
    $cim=Get-Command Get-CimInstance
    $sleep=Get-Command Start-Sleep
    $clock=[Diagnostics.Stopwatch]::new()
    $backend=@{
        Preflight={param($context)
            & $validate $Policy $context
            if($State.RunId -cne $Policy.RunId){& $record $State 'Context';throw 'Failure state mismatch.'}
            & $preflight $Policy
        }.GetNewClosure()
        AcquireWriter={param($context) & $newWriter $Policy $context $State}.GetNewClosure()
        AcquireProcess={param($context)
            & $validate $Policy $context
            if($State.RunId -cne $Policy.RunId){& $record $State 'Context';throw 'Failure state mismatch.'}
            & $geometryDestination $Policy
            $clock.Start()
            # A visible interactive window is required by the separately authorized EG route.
            $original=& $launch -FilePath $Policy.Executable -ArgumentList $Policy.Arguments -WorkingDirectory ([IO.Path]::GetDirectoryName($Policy.Executable)) -PassThru -ErrorAction Stop
            try{return (& $newProcessOwner $original $State)}
            catch{
                & $record $State 'ProcessAdapter'
                if($null -ne $original){try{$original.Dispose()}catch{& $record $State 'ProcessDispose'}}
                throw
            }
        }.GetNewClosure()
        GetIdentity={param($process) @{ProcessId=$process.Original.Id;StartUtc=$process.Original.StartTime.ToUniversalTime();Handle=$process.Original.SafeHandle}}
        ReadNative={param($handle,$runId) & $native -ProcessHandle $handle -CorrelationId $runId}.GetNewClosure()
        ReadCim={param($processId)
            $rows=@(& $cim -ClassName Win32_Process -Filter "ProcessId = $processId" -ErrorAction Stop)
            if($rows.Count -ne 1 -or $rows[0].CreationDate -isnot [datetime]){throw 'CIM identity unavailable.'}
            $row=$rows[0]
            if($row.CreationDate.Kind -eq [DateTimeKind]::Unspecified){throw 'CIM time zone unavailable.'}
            @{ProcessId=$row.ProcessId;ExecutablePath=$row.ExecutablePath;CommandLine=$row.CommandLine;CreationUtc=$row.CreationDate.ToUniversalTime()}
        }.GetNewClosure()
        ReadManaged={param($process) & $managed -ProcessObject $process.Original}.GetNewClosure()
        ElapsedMs={$clock.Elapsed.TotalMilliseconds}.GetNewClosure()
        SleepMs={param($milliseconds) & $sleep -Milliseconds $milliseconds}.GetNewClosure()
        Publish={param($record) Write-Information ($record|ConvertTo-Json -Depth 4 -Compress) -InformationAction Continue}
    }
    $categories=@{Preflight='Preflight';AcquireWriter='ReceiptDirectory';AcquireProcess='ProcessAcquire';GetIdentity='Identity';ReadNative='NativeRead';ReadCim='CimRead';ReadManaged='ManagedRead';ElapsedMs='Clock';SleepMs='Sleep';Publish='Publication'}
    foreach($key in @($backend.Keys)){
        # Factory records the exact stage; a broad label would misstate the cause.
        if($key -ceq 'AcquireWriter'){continue}
        $callback=$backend[$key];$category=$categories[$key]
        $backend[$key]={param($a,$b)
            try{& $callback $a $b}catch{& $record $State $category;throw}
        }.GetNewClosure()
    }
    $backend
}
function Stop-EGFailure($State){
    $summary=Get-EGFailureSummary $State
    try{Write-Information ($summary|ConvertTo-Json -Depth 5 -Compress) -InformationAction Continue -ErrorAction Stop}
    catch{Add-EGFailure $State 'Publication';$summary=Get-EGFailureSummary $State}
    $failure=[InvalidOperationException]::new('EG observer failed; no acceptance.')
    $failure.Data['EGFailureSummary']=$summary
    throw $failure
}
function Invoke-EGSession {
    [CmdletBinding()]
    param($Policy,$Context,$State,$Backend,[ValidateSet('Synthetic','Live')][string]$Mode)
    try{
        Assert-EGContext $Policy $Context
        if($State.RunId -cne $Policy.RunId){throw 'Failure state mismatch.'}
    }catch{Add-EGFailure $State 'Context';Stop-EGFailure $State}
    try{$result=Invoke-TalenroC5Observation -Mode $Mode -Context $Context -Backend $Backend}
    catch{Add-EGFailure $State 'SessionUnclassified';Stop-EGFailure $State}
    if($State.OperationFailures.Count -or $State.CleanupFailures.Count){Stop-EGFailure $State}
    try{ConvertTo-EGResult $Policy $Context $result}
    catch{Add-EGFailure $State 'Result';Stop-EGFailure $State}
}

$state=New-EGFailureState $null
$setupStage='Setup'
try{
    $root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
    $runId=[guid]::NewGuid().ToString('N')
    $state.RunId=$runId
    $policy=Get-EGPolicy $root $runId
    $setupStage='Preflight'
    Assert-EGPreflight $policy
    $setupStage='Import'
    Import-Module (Join-Path $PSScriptRoot 'c5-observation-session.psm1')
    Import-Module (Join-Path $PSScriptRoot 'process-observation.psm1')
    Import-Module (Join-Path $PSScriptRoot 'native-process-observation.psm1')
    $setupStage='Setup'
    $context=@{RunId=$policy.RunId;ExpectedPath=$policy.Executable;ExpectedCommandLine=$policy.ExpectedCommandLine}
    $backend=New-EGBackend $policy $state
}catch{Add-EGFailure $state $setupStage;Stop-EGFailure $state}
# Keep terminal failure publication outside the setup catch: no double finalization.
$wrapped=Invoke-EGSession $policy $context $state $backend Live
$wrapped
if(-not $wrapped.ProcessObservation.ExitVerified -or -not $wrapped.ProcessObservation.Finalized){exit 1}
