#requires -Version 7.0
[CmdletBinding()]
param([switch]$RunApproved)

# Approval of implementation does not authorize a live run.
if(-not $RunApproved){[pscustomobject]@{Status='NotStarted';Reason='Separate bounded approval required'};return}
$ErrorActionPreference='Stop'

function Get-OrdinaryPolicy([string]$Root){
    $exe=Join-Path $Root 'clients/windows/Talenro.Demo.App/bin/ordinary-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe'
    $arguments=''
    @{
        Root=$Root;Executable=$exe;Dll=[IO.Path]::ChangeExtension($exe,'.dll');Arguments=$arguments
        ExpectedCommandLine='"'+$exe+'"'+[char]32
        CommandLineProfile='shell-quoted-exe-one-u0020-v1'
        LauncherExpected=@{
            CommandName='Microsoft.PowerShell.Management\Start-Process';CommandType='Cmdlet'
            ImplementingType='Microsoft.PowerShell.Commands.StartProcessCommand'
            PowerShellVersion='7.6.9';Edition='Core';AssemblyName='Microsoft.PowerShell.Commands.Management.dll'
            AssemblyVersion='7.6.0.500';FileVersion='7.6.9.500'
            Sha256='2181516B3A55866A2263DDAB22A7547FC815240F474199088B8ABA030B92264A'
        }
        BuildSourceCommit='5ee8a339b574171fc61f3cf0c0c636bd86421397'
        ArtifactHashes=@{
            'Talenro.Demo.App.exe'='FEDA8E835E8390ED7F166FA764413FBEC9BCE591C7C7CF7E8E43E028B03C0375'
            'Talenro.Demo.App.dll'='773AE00EC4E5552D65ECDFA527FEEE6688AE380D8ED4DB4D3282DD10C1B76A0B'
            'Talenro.Demo.Core.dll'='67FDB3F42E69350400B8A9FC11BCFB54670233B4571F2E27BA91F30E8D19DE18'
            'Views/AccountPage.xbf'='A50DDC403AD9E41EA977AE868551AE25B6079674094EBE0BE38CD12C2B7EFE6F'
            'Talenro.Demo.App.deps.json'='99DD4D70751D66997DF9DA858C09DDA2BB68A8FDFAC7F671E4FAB5B35D910ED6'
            'Talenro.Demo.App.runtimeconfig.json'='3FA241FAED5CC59579FA3B630E1A873E5539428B9CED9A4CA48F5828488E9666'
        }
        ReceiptRoot=Join-Path ([IO.Path]::GetDirectoryName($exe)) 'ordinary-account-observations'
        Sources=@('observe-ordinary-account-center.ps1','c5-observation-session.psm1','observation-session-guard.psm1',
            'process-observation.psm1','native-process-observation.psm1','native-process-observation.cs')
    }
}
function Get-OrdinaryLauncherEvidence {
    $commands=@(Get-Command 'Microsoft.PowerShell.Management\Start-Process' -ErrorAction Stop)
    if($commands.Count -ne 1 -or $commands[0] -isnot [Management.Automation.CmdletInfo]){throw 'Launcher discovery invalid.'}
    $command=$commands[0]
    $assembly=$command.ImplementingType.Assembly
    $location=$assembly.Location
    if([string]::IsNullOrWhiteSpace($location) -or -not [IO.Path]::IsPathFullyQualified($location) -or
        -not(Test-Path -LiteralPath $location -PathType Leaf -ErrorAction Stop)){throw 'Launcher file unavailable.'}
    $path=$location
    while(-not [string]::IsNullOrEmpty($path)){
        $item=Get-Item -LiteralPath $path -Force -ErrorAction Stop
        if(($item.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0){throw 'Reparse launcher path rejected.'}
        $path=[IO.Path]::GetDirectoryName($path)
    }
    @{
        Command=$command;CommandName=$command.ModuleName+'\'+$command.Name;CommandType=$command.CommandType.ToString()
        ImplementingType=$command.ImplementingType.FullName;PowerShellVersion=$PSVersionTable.PSVersion.ToString();Edition=$PSVersionTable.PSEdition
        AssemblyPath=$location;AssemblyName=[IO.Path]::GetFileName($location);AssemblyVersion=$assembly.GetName().Version.ToString()
        FileVersion=(Get-Item -LiteralPath $location -Force -ErrorAction Stop).VersionInfo.FileVersion
        Sha256=(Get-FileHash -LiteralPath $location -Algorithm SHA256 -ErrorAction Stop).Hash
    }
}
function Assert-OrdinaryLauncherBinding($Policy,$Evidence){
    if($null -eq $Evidence){throw 'Launcher evidence absent.'}
    $snapshot=@{}
    foreach($key in $Policy.LauncherExpected.Keys){
        if($Evidence[$key] -isnot [string] -or $Evidence[$key] -cne $Policy.LauncherExpected[$key]){throw ('Launcher binding mismatch: '+$key)}
        $snapshot[$key]=$Evidence[$key]
    }
    if($Evidence.Command -isnot [Management.Automation.CmdletInfo] -or
        ($Evidence.Command.ModuleName+'\'+$Evidence.Command.Name) -cne $snapshot.CommandName -or
        $Evidence.Command.ImplementingType.FullName -cne $snapshot.ImplementingType -or
        $Evidence.Command.ImplementingType.Assembly.Location -cne $Evidence.AssemblyPath -or
        [string]::IsNullOrWhiteSpace($Evidence.AssemblyPath) -or -not [IO.Path]::IsPathFullyQualified($Evidence.AssemblyPath) -or
        [IO.Path]::GetFileName($Evidence.AssemblyPath) -cne $snapshot.AssemblyName){throw 'Launcher command evidence mismatch.'}
    @{Command=$Evidence.Command;Evidence=$snapshot}
}
function Assert-OrdinaryOrdinaryToken {
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    try{
        $principal=[Security.Principal.WindowsPrincipal]::new($identity)
        if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary token required.'}
    }finally{$identity.Dispose()}
}
function Invoke-OrdinaryGit([string]$Root,[string[]]$Arguments){
    $output=& git -c "safe.directory=$Root" -C $Root @Arguments 2>$null
    if($LASTEXITCODE -ne 0){throw 'Git evidence unavailable.'}
    $output
}
function Assert-OrdinaryPreflight($Policy){
    Assert-OrdinaryOrdinaryToken
    if((Invoke-OrdinaryGit $Policy.Root @('branch','--show-current')) -cne 'codex/c12-b01-task9-coordinator'){throw 'Unexpected branch.'}
    if(@(Invoke-OrdinaryGit $Policy.Root @('status','--porcelain','--untracked-files=all')).Count -ne 0){throw 'Clean worktree required.'}
    $null=Invoke-OrdinaryGit $Policy.Root @('cat-file','-e',($Policy.BuildSourceCommit+'^{commit}'))
    $null=Invoke-OrdinaryGit $Policy.Root @('diff','--exit-code',$Policy.BuildSourceCommit,'--','clients/windows/Talenro.Demo.App','clients/windows/Talenro.Demo.Core','clients/windows/Directory.Build.props')
    foreach($source in $Policy.Sources){
        $relative='clients/windows/tests/'+$source
        $expected=Invoke-OrdinaryGit $Policy.Root @('rev-parse',('HEAD:'+$relative))
        $actual=Invoke-OrdinaryGit $Policy.Root @('hash-object',('--path='+$relative),(Join-Path $Policy.Root $relative))
        if([string]::IsNullOrWhiteSpace($expected) -or $expected -cne $actual){throw 'Observer source mismatch.'}
    }
    Assert-OrdinaryArtifactPaths $Policy
    foreach($name in $Policy.ArtifactHashes.Keys){
        $path=Join-Path ([IO.Path]::GetDirectoryName($Policy.Executable)) $name
        if(-not(Test-Path -LiteralPath $path -PathType Leaf -ErrorAction Stop) -or
            (Get-FileHash -LiteralPath $path -Algorithm SHA256 -ErrorAction Stop).Hash -cne $Policy.ArtifactHashes[$name]){throw 'Preserved artifact unavailable or changed.'}
    }
    $targets=@(Get-CimInstance -ClassName Win32_Process -Filter "Name LIKE 'Talenro.Demo.%' OR Name LIKE '%Prototype%'" -ErrorAction Stop)
    if($targets.Count -ne 0){throw 'Existing target blocks launch.'}
    Assert-OrdinaryLauncherBinding $Policy (Get-OrdinaryLauncherEvidence)
}
function Assert-OrdinaryArtifactPaths($Policy){
    foreach($name in $Policy.ArtifactHashes.Keys){
        $path=Join-Path ([IO.Path]::GetDirectoryName($Policy.Executable)) $name
        if(-not(Test-Path -LiteralPath $path -PathType Leaf -ErrorAction Stop)){throw 'Artifact missing.'}
        while(-not [string]::IsNullOrEmpty($path)){
            $item=Get-Item -LiteralPath $path -Force -ErrorAction Stop
            if(($item.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0){throw 'Reparse artifact path rejected.'}
            $path=[IO.Path]::GetDirectoryName($path)
        }
    }
}
function Assert-OrdinaryReceiptDestination([string]$Directory){
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
function New-OrdinaryReceiptDirectory([string]$Directory){
    $null=New-Item -ItemType Directory -Path $Directory -ErrorAction Stop
}
function Open-OrdinaryNewStream([string]$Path){
    [IO.FileStream]::new($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
}
function New-OrdinaryTextWriter($Stream){
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
function New-OrdinaryReceiptWriter($Policy,$Context,$Binding){
    if($null -eq $Binding -or $null -eq $Binding.Command -or $null -eq $Binding.Evidence){throw 'Verified launcher binding required.'}
    $launcher=@{}
    foreach($key in $Policy.LauncherExpected.Keys){
        if($Binding.Evidence[$key] -isnot [string] -or $Binding.Evidence[$key] -cne $Policy.LauncherExpected[$key]){throw 'Receipt launcher mismatch.'}
        $launcher[$key]=$Binding.Evidence[$key]
    }
    if($Context.RunId -isnot [string] -or $Context.RunId -cnotmatch '^[0-9a-f]{32}$' -or
        $Context.RunId -eq ('0'*32)){throw 'Invalid receipt ID.'}
    if($Context.ExpectedPath -cne $Policy.Executable -or
        $Context.ExpectedCommandLine -cne $Policy.ExpectedCommandLine){throw 'Receipt context mismatch.'}
    $directory=Join-Path $Policy.ReceiptRoot $Context.RunId
    Assert-OrdinaryReceiptDestination $directory
    New-OrdinaryReceiptDirectory $directory
    $stream=$null;$writer=$null
    try{
        $manifest=[ordered]@{
            SchemaVersion=1;Observation='OrdinaryAccountCenter';RunId=$Context.RunId;Language='default-environment';Profile='default-environment'
            ExecutableRelativePath='clients/windows/Talenro.Demo.App/bin/ordinary-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe'
            Arguments=$Policy.Arguments;ArtifactHashes=$Policy.ArtifactHashes;BuildSourceCommit=$Policy.BuildSourceCommit
            ObservationCapMs=360000;MaxSleepMs=2000
            CommandLineProfile=$Policy.CommandLineProfile;ExpectedCommandLine=$Policy.ExpectedCommandLine;Launcher=$launcher
        }
        $json=$manifest|ConvertTo-Json -Depth 4 -Compress
        $stream=Open-OrdinaryNewStream (Join-Path $directory 'manifest.json')
        $writer=New-OrdinaryTextWriter $stream
        $writer.WriteLine($json)
        $writer.Flush()
        # Dispose separately: even a writer disposal failure must not leak its stream.
        $closing=$writer;$writer=$null
        $closing.Dispose()
        $closing=$stream;$stream=$null
        $closing.Dispose()
        $stream=Open-OrdinaryNewStream (Join-Path $directory 'samples.jsonl')
        $writer=New-OrdinaryTextWriter $stream
        $result=$writer;$writer=$null;$stream=$null
        return $result
    }catch{
        $failures=[Collections.Generic.List[Exception]]::new()
        $failures.Add($_.Exception)
        if($null -ne $writer){try{$writer.Dispose()}catch{$failures.Add($_.Exception)}}
        if($null -ne $stream){try{$stream.Dispose()}catch{$failures.Add($_.Exception)}}
        throw [AggregateException]::new('Ordinary receipt acquisition failed.',$failures)
    }
}
function New-OrdinaryBackend($Policy){
    $preflight=Get-Command Assert-OrdinaryPreflight
    $newWriter=Get-Command New-OrdinaryReceiptWriter
    $state=@{Binding=$null}
    $native=Get-Command Get-TalenroNativeProcessObservation
    $managed=Get-Command Get-TalenroProcessObservation
    $cim=Get-Command Get-CimInstance
    $sleep=Get-Command Start-Sleep
    $clock=[Diagnostics.Stopwatch]::new()
    @{
        Preflight={param($context) $state.Binding=$null;$state.Binding=& $preflight $Policy}.GetNewClosure()
        AcquireWriter={param($context)
            if($null -eq $state.Binding){throw 'Launcher not validated.'}
            & $newWriter $Policy $context $state.Binding
        }.GetNewClosure()
        AcquireProcess={param($context)
            if($null -eq $state.Binding){throw 'Launcher not validated.'}
            $launch=$state.Binding.Command
            $clock.Start()
            # A visible interactive window is required by the separately authorized Ordinary route.
            & $launch -FilePath $Policy.Executable -WorkingDirectory ([IO.Path]::GetDirectoryName($Policy.Executable)) -PassThru -ErrorAction Stop
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
    $policy=Get-OrdinaryPolicy $root
    $null=Assert-OrdinaryPreflight $policy
    Import-Module (Join-Path $PSScriptRoot 'c5-observation-session.psm1')
    Import-Module (Join-Path $PSScriptRoot 'process-observation.psm1')
    Import-Module (Join-Path $PSScriptRoot 'native-process-observation.psm1')
    $context=@{RunId=[guid]::NewGuid().ToString('N');ExpectedPath=$policy.Executable;ExpectedCommandLine=$policy.ExpectedCommandLine}
    $result=Invoke-TalenroC5Observation -Mode Live -Context $context -Backend (New-OrdinaryBackend $policy)
    $result
    if(-not $result.ExitVerified){exit 1}
}catch{
    Write-Information 'Ordinary blocked or aborted; no acceptance. No retry. If a target started, arrange normal close.' -InformationAction Continue
    throw [InvalidOperationException]::new('Ordinary observer failed; no acceptance.')
}
