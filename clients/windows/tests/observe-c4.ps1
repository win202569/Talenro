#requires -Version 7.0
[CmdletBinding()]
param([switch]$RunApproved)

# Approval of implementation does not authorize a live run.
if(-not $RunApproved){[pscustomobject]@{Status='NotStarted';Reason='Separate bounded approval required'};return}
$ErrorActionPreference='Stop'

function Get-C4Policy([string]$Root){
    $exe=Join-Path $Root 'clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe'
    $arguments='--layout-test --layout-language=en-US --layout-profile=rtl-mixed'
    @{
        Root=$Root;Executable=$exe;Dll=[IO.Path]::ChangeExtension($exe,'.dll');Arguments=$arguments
        ExpectedCommandLine='"'+$exe+'" '+$arguments
        ExeHash='4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95'
        DllHash='7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB'
        ReceiptRoot=Join-Path ([IO.Path]::GetDirectoryName($exe)) 'process-observations'
        Sources=@('observe-c4.ps1','c5-observation-session.psm1','observation-session-guard.psm1',
            'process-observation.psm1','native-process-observation.psm1','native-process-observation.cs')
    }
}
function Assert-C4OrdinaryToken {
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    try{
        $principal=[Security.Principal.WindowsPrincipal]::new($identity)
        if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary token required.'}
    }finally{$identity.Dispose()}
}
function Invoke-C4Git([string]$Root,[string[]]$Arguments){
    $output=& git -c "safe.directory=$Root" -C $Root @Arguments 2>$null
    if($LASTEXITCODE -ne 0){throw 'Git evidence unavailable.'}
    $output
}
function Assert-C4Preflight($Policy){
    Assert-C4OrdinaryToken
    if((Invoke-C4Git $Policy.Root @('branch','--show-current')) -cne 'codex/c12-b01-task9-coordinator'){throw 'Unexpected branch.'}
    if(@(Invoke-C4Git $Policy.Root @('status','--porcelain','--untracked-files=all')).Count -ne 0){throw 'Clean worktree required.'}
    foreach($source in $Policy.Sources){
        $relative='clients/windows/tests/'+$source
        $expected=Invoke-C4Git $Policy.Root @('rev-parse',('HEAD:'+$relative))
        $actual=Invoke-C4Git $Policy.Root @('hash-object',('--path='+$relative),(Join-Path $Policy.Root $relative))
        if([string]::IsNullOrWhiteSpace($expected) -or $expected -cne $actual){throw 'Observer source mismatch.'}
    }
    foreach($pair in @(@($Policy.Executable,$Policy.ExeHash),@($Policy.Dll,$Policy.DllHash))){
        if(-not(Test-Path -LiteralPath $pair[0] -PathType Leaf) -or
            (Get-FileHash -LiteralPath $pair[0] -Algorithm SHA256).Hash -cne $pair[1]){throw 'Preserved artifact unavailable or changed.'}
    }
    $targets=@(Get-CimInstance -ClassName Win32_Process -Filter "Name LIKE 'Talenro.Demo.%' OR Name LIKE '%Prototype%'" -ErrorAction Stop)
    if($targets.Count -ne 0){throw 'Existing target blocks launch.'}
}
function Assert-C4ReceiptDestination([string]$Directory){
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
function New-C4ReceiptDirectory([string]$Directory){
    $null=New-Item -ItemType Directory -Path $Directory -ErrorAction Stop
}
function Open-C4NewStream([string]$Path){
    [IO.FileStream]::new($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
}
function New-C4TextWriter($Stream){
    $writer=$null
    try{
        $writer=[IO.StreamWriter]::new($Stream,[Text.UTF8Encoding]::new($false))
        $writer.AutoFlush=$true
        return $writer
    }catch{
        $primary=$_.Exception
        if($null -ne $writer){try{$writer.Dispose()}catch{}}
        # Caller still owns and disposes the stream if construction fails.
        throw $primary
    }
}
function New-C4ReceiptWriter($Policy,$Context){
    if($Context.RunId -isnot [string] -or $Context.RunId -cnotmatch '^[0-9a-f]{32}$' -or
        $Context.RunId -eq ('0'*32)){throw 'Invalid receipt ID.'}
    if($Context.ExpectedPath -cne $Policy.Executable -or
        $Context.ExpectedCommandLine -cne $Policy.ExpectedCommandLine){throw 'Receipt context mismatch.'}
    $directory=Join-Path $Policy.ReceiptRoot $Context.RunId
    Assert-C4ReceiptDestination $directory
    New-C4ReceiptDirectory $directory
    $stream=$null;$writer=$null
    try{
        $manifest=[ordered]@{
            SchemaVersion=1;Observation='C4';RunId=$Context.RunId;Language='en-US';Profile='rtl-mixed'
            ExecutableRelativePath='clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe'
            Arguments=$Policy.Arguments;ExeSha256=$Policy.ExeHash;DllSha256=$Policy.DllHash
            ObservationCapMs=360000;MaxSleepMs=2000
        }
        $json=$manifest|ConvertTo-Json -Depth 4 -Compress
        $stream=Open-C4NewStream (Join-Path $directory 'manifest.json')
        $writer=New-C4TextWriter $stream
        $writer.WriteLine($json)
        $writer.Flush()
        # Dispose separately: even a writer disposal failure must not leak its stream.
        $closing=$writer;$writer=$null
        $closing.Dispose()
        $closing=$stream;$stream=$null
        $closing.Dispose()
        $stream=Open-C4NewStream (Join-Path $directory 'samples.jsonl')
        $writer=New-C4TextWriter $stream
        $result=$writer;$writer=$null;$stream=$null
        return $result
    }catch{
        $failures=[Collections.Generic.List[Exception]]::new()
        $failures.Add($_.Exception)
        if($null -ne $writer){try{$writer.Dispose()}catch{$failures.Add($_.Exception)}}
        if($null -ne $stream){try{$stream.Dispose()}catch{$failures.Add($_.Exception)}}
        throw [AggregateException]::new('C4 receipt acquisition failed.',$failures)
    }
}
function New-C4Backend($Policy){
    $preflight=Get-Command Assert-C4Preflight
    $newWriter=Get-Command New-C4ReceiptWriter
    $launch=Get-Command Start-Process
    $native=Get-Command Get-TalenroNativeProcessObservation
    $managed=Get-Command Get-TalenroProcessObservation
    $cim=Get-Command Get-CimInstance
    $sleep=Get-Command Start-Sleep
    $clock=[Diagnostics.Stopwatch]::new()
    @{
        Preflight={param($context) & $preflight $Policy}.GetNewClosure()
        AcquireWriter={param($context) & $newWriter $Policy $context}.GetNewClosure()
        AcquireProcess={param($context)
            $clock.Start()
            # A visible interactive window is required by the separately authorized C4 route.
            & $launch -FilePath $Policy.Executable -ArgumentList $Policy.Arguments -WorkingDirectory ([IO.Path]::GetDirectoryName($Policy.Executable)) -PassThru -ErrorAction Stop
        }.GetNewClosure()
        GetIdentity={param($process) @{ProcessId=$process.Id;StartUtc=$process.StartTime.ToUniversalTime();Handle=$process.SafeHandle}}
        ReadNative={param($handle,$runId) & $native -ProcessHandle $handle -CorrelationId $runId}.GetNewClosure()
        ReadCim={param($processId)
            $rows=@(& $cim -ClassName Win32_Process -Filter "ProcessId = $processId" -ErrorAction Stop)
            if($rows.Count -ne 1 -or $rows[0].CreationDate -isnot [datetime]){throw 'CIM identity unavailable.'}
            $row=$rows[0]
            if($row.CreationDate.Kind -eq [DateTimeKind]::Unspecified){throw 'CIM time zone unavailable.'}
            @{ProcessId=$row.ProcessId;ExecutablePath=$row.ExecutablePath;CommandLine=$row.CommandLine;CreationUtc=$row.CreationDate.ToUniversalTime()}
        }.GetNewClosure()
        ReadManaged={param($process) & $managed -ProcessObject $process}.GetNewClosure()
        ElapsedMs={$clock.Elapsed.TotalMilliseconds}.GetNewClosure()
        SleepMs={param($milliseconds) & $sleep -Milliseconds $milliseconds}.GetNewClosure()
        Publish={param($record) Write-Information ($record|ConvertTo-Json -Depth 4 -Compress) -InformationAction Continue}
    }
}

try{
    $root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
    $policy=Get-C4Policy $root
    Assert-C4Preflight $policy
    Import-Module (Join-Path $PSScriptRoot 'c5-observation-session.psm1')
    Import-Module (Join-Path $PSScriptRoot 'process-observation.psm1')
    Import-Module (Join-Path $PSScriptRoot 'native-process-observation.psm1')
    $context=@{RunId=[guid]::NewGuid().ToString('N');ExpectedPath=$policy.Executable;ExpectedCommandLine=$policy.ExpectedCommandLine}
    $result=Invoke-TalenroC5Observation -Mode Live -Context $context -Backend (New-C4Backend $policy)
    $result
    if(-not $result.ExitVerified){exit 1}
}catch{
    Write-Information 'C4 blocked or aborted; no acceptance. No retry. If a target started, arrange normal close.' -InformationAction Continue
    throw [InvalidOperationException]::new('C4 observer failed; no acceptance.')
}
