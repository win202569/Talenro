#requires -Version 7.0
[CmdletBinding()]
param([switch]$RunApproved)

# This guard precedes imports, environment probes and receipt creation.
if(-not $RunApproved){[pscustomobject]@{Status='NotStarted';Reason='Separate bounded approval required'};return}
$ErrorActionPreference='Stop'

function Get-C5Policy([string]$Root){
    $exe=Join-Path $Root 'clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe'
    $arguments='--layout-test --layout-language=zh-Hans --layout-profile=unknown-usage'
    @{
        Root=$Root;Executable=$exe;Dll=[IO.Path]::ChangeExtension($exe,'.dll');Arguments=$arguments
        ExpectedCommandLine='"'+$exe+'" '+$arguments
        ExeHash='4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95'
        DllHash='7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB'
        ReceiptRoot=Join-Path ([IO.Path]::GetDirectoryName($exe)) 'process-observations'
        Sources=@('observe-c5.ps1','c5-observation-session.psm1','observation-session-guard.psm1',
            'process-observation.psm1','native-process-observation.psm1','native-process-observation.cs')
    }
}
function Assert-C5OrdinaryToken {
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    try{
        $principal=[Security.Principal.WindowsPrincipal]::new($identity)
        if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary token required.'}
    }finally{$identity.Dispose()}
}
function Invoke-C5Git([string]$Root,[string[]]$Arguments){
    $output=& git -c "safe.directory=$Root" -C $Root @Arguments 2>$null
    if($LASTEXITCODE -ne 0){throw 'Git evidence unavailable.'}
    $output
}
function Assert-C5Preflight($Policy){
    Assert-C5OrdinaryToken
    if((Invoke-C5Git $Policy.Root @('branch','--show-current')) -cne 'codex/c12-b01-task9-coordinator'){throw 'Unexpected branch.'}
    if(@(Invoke-C5Git $Policy.Root @('status','--porcelain','--untracked-files=all')).Count -ne 0){throw 'Clean worktree required.'}
    foreach($source in $Policy.Sources){
        $relative='clients/windows/tests/'+$source
        $expected=Invoke-C5Git $Policy.Root @('rev-parse',('HEAD:'+$relative))
        $actual=Invoke-C5Git $Policy.Root @('hash-object',('--path='+$relative),(Join-Path $Policy.Root $relative))
        if([string]::IsNullOrWhiteSpace($expected) -or $expected -cne $actual){throw 'Observer source mismatch.'}
    }
    foreach($pair in @(@($Policy.Executable,$Policy.ExeHash),@($Policy.Dll,$Policy.DllHash))){
        if(-not(Test-Path -LiteralPath $pair[0] -PathType Leaf) -or
            (Get-FileHash -LiteralPath $pair[0] -Algorithm SHA256).Hash -cne $pair[1]){throw 'Preserved artifact unavailable or changed.'}
    }
    $targets=@(Get-CimInstance -ClassName Win32_Process -Filter "Name LIKE 'Talenro.Demo.%' OR Name LIKE '%Prototype%'" -ErrorAction Stop)
    if($targets.Count -ne 0){throw 'Existing target blocks launch.'}
}
function New-C5Writer($Policy,$Context){
    if($Context.RunId -cnotmatch '^[0-9a-f]{32}$' -or $Context.RunId -eq ('0'*32)){throw 'Invalid receipt ID.'}
    # Refuse reparse points along every existing ancestor; not an adversarial race guarantee.
    $directory=Join-Path $Policy.ReceiptRoot $Context.RunId
    $ancestor=[IO.DirectoryInfo]::new($directory)
    while($null -ne $ancestor){
        if(([IO.Directory]::Exists($ancestor.FullName) -or [IO.File]::Exists($ancestor.FullName)) -and
            (([IO.File]::GetAttributes($ancestor.FullName) -band [IO.FileAttributes]::ReparsePoint) -ne 0)){throw 'Reparse receipt path rejected.'}
        $ancestor=$ancestor.Parent
    }
    if([IO.Directory]::Exists($directory) -or [IO.File]::Exists($directory)){throw 'Receipt destination exists.'}
    [void][IO.Directory]::CreateDirectory($directory)
    $stream=$null
    try{
        $stream=[IO.FileStream]::new((Join-Path $directory 'samples.jsonl'),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
        $writer=[IO.StreamWriter]::new($stream,[Text.UTF8Encoding]::new($false))
        $writer.AutoFlush=$true
        return $writer
    }catch{if($null -ne $stream){$stream.Dispose()};throw}
}
function New-C5Backend($Policy){
    # Capture commands before closures: private functions must not depend on caller scope.
    $preflight=Get-Command Assert-C5Preflight
    $newWriter=Get-Command New-C5Writer
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
            # A visible interactive window is required for the separately authorized C5 operator route.
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
    $policy=Get-C5Policy $root
    # Validate source and environment before importing any observation implementation.
    Assert-C5Preflight $policy
    Import-Module (Join-Path $PSScriptRoot 'c5-observation-session.psm1')
    Import-Module (Join-Path $PSScriptRoot 'process-observation.psm1')
    Import-Module (Join-Path $PSScriptRoot 'native-process-observation.psm1')
    $context=@{RunId=[guid]::NewGuid().ToString('N');ExpectedPath=$policy.Executable;ExpectedCommandLine=$policy.ExpectedCommandLine}
    $result=Invoke-TalenroC5Observation -Mode Live -Context $context -Backend (New-C5Backend $policy)
    $result
    if(-not $result.ExitVerified){exit 1}
}catch{
    Write-Information 'C5 blocked or aborted; no acceptance. No retry. If a target started, arrange normal close.' -InformationAction Continue
    throw [InvalidOperationException]::new('C5 observer failed; no acceptance.')
}
