[CmdletBinding(DefaultParameterSetName = 'Focused')]
param(
  [Parameter(Mandatory = $true, ParameterSetName = 'Focused')]
  [string]$Profile,

  [Parameter(Mandatory = $true, ParameterSetName = 'Focused')]
  [string]$Packages,

  [Parameter(ParameterSetName = 'Focused')]
  [string]$Run = '',

  [Parameter(ParameterSetName = 'Focused')]
  [switch]$Race,

  [Parameter(Mandatory = $true, ParameterSetName = 'Suite')]
  [string]$Suite,

  [Parameter(Mandatory = $true, ParameterSetName = 'Focused')]
  [Parameter(Mandatory = $true, ParameterSetName = 'Suite')]
  [string]$Timeout
)

Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'

$script:c12AllowedPackages = @{
  './internal/testinfra' = $true
  './internal/store' = $true
  './internal/nodecontrol/contracts' = $true
  './internal/nodecontrol/authority' = $true
  './internal/nodecontrol/serving' = $true
  './internal/readiness' = $true
}
$script:c12NativeDeadline = [DateTime]::MaxValue
$script:c12SuiteDeadline = [DateTime]::MaxValue
if ($PSCmdlet.ParameterSetName -eq 'Suite') {
  $script:c12SuiteDeadline = [DateTime]::UtcNow.AddMinutes(120)
}

function ConvertFrom-C12Duration {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Value
  )

  if ($Value -notmatch '^([1-9][0-9]*)(s|m)$') {
    throw "timeout '$Value' is not a finite positive seconds/minutes duration"
  }
  [long]$quantity = 0
  if (-not [long]::TryParse($Matches[1], [ref]$quantity)) {
    throw "timeout '$Value' is outside the supported range"
  }
  if ($Matches[2] -eq 'm') {
    if ($quantity -gt 1440) {
      throw "timeout '$Value' is outside the supported range"
    }
    return [TimeSpan]::FromMinutes($quantity)
  }
  if ($quantity -gt 86400) {
    throw "timeout '$Value' is outside the supported range"
  }
  return [TimeSpan]::FromSeconds($quantity)
}

function New-C12RandomSuffix {
  $bytes = New-Object byte[] 16
  $random = [System.Security.Cryptography.RandomNumberGenerator]::Create()
  try {
    $random.GetBytes($bytes)
  }
  finally {
    $random.Dispose()
  }
  return (($bytes | ForEach-Object { $_.ToString('x2') }) -join '')
}

function New-C12DatabasePassword {
  param(
    [Parameter(Mandatory = $true)]
    [string]$RunSuffix
  )

  $hash = [System.Security.Cryptography.SHA256]::Create()
  try {
    $bytes = [System.Text.Encoding]::UTF8.GetBytes("TALENRO-C12-POSTGRES-PASSWORD-V1`0$RunSuffix")
    return (($hash.ComputeHash($bytes) | ForEach-Object { $_.ToString('x2') }) -join '')
  }
  finally {
    $hash.Dispose()
  }
}

Add-Type -TypeDefinition @'
using System;
using System.ComponentModel;
using System.Diagnostics;
using System.Runtime.InteropServices;

public static class C12NativeJob
{
    private const UInt32 JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE = 0x00002000;

    [StructLayout(LayoutKind.Sequential)]
    private struct JOBOBJECT_BASIC_LIMIT_INFORMATION
    {
        public Int64 PerProcessUserTimeLimit;
        public Int64 PerJobUserTimeLimit;
        public UInt32 LimitFlags;
        public UIntPtr MinimumWorkingSetSize;
        public UIntPtr MaximumWorkingSetSize;
        public UInt32 ActiveProcessLimit;
        public IntPtr Affinity;
        public UInt32 PriorityClass;
        public UInt32 SchedulingClass;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct IO_COUNTERS
    {
        public UInt64 ReadOperationCount;
        public UInt64 WriteOperationCount;
        public UInt64 OtherOperationCount;
        public UInt64 ReadTransferCount;
        public UInt64 WriteTransferCount;
        public UInt64 OtherTransferCount;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct JOBOBJECT_EXTENDED_LIMIT_INFORMATION
    {
        public JOBOBJECT_BASIC_LIMIT_INFORMATION BasicLimitInformation;
        public IO_COUNTERS IoInfo;
        public UIntPtr ProcessMemoryLimit;
        public UIntPtr JobMemoryLimit;
        public UIntPtr PeakProcessMemoryUsed;
        public UIntPtr PeakJobMemoryUsed;
    }

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern IntPtr CreateJobObject(IntPtr attributes, string name);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool SetInformationJobObject(IntPtr job, int informationClass, IntPtr information, UInt32 length);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool AssignProcessToJobObject(IntPtr job, IntPtr process);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool CloseHandle(IntPtr handle);

    public static IntPtr CreateKillOnClose()
    {
        IntPtr job = CreateJobObject(IntPtr.Zero, null);
        if (job == IntPtr.Zero) throw new Win32Exception(Marshal.GetLastWin32Error());
        var information = new JOBOBJECT_EXTENDED_LIMIT_INFORMATION();
        information.BasicLimitInformation.LimitFlags = JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE;
        int length = Marshal.SizeOf(typeof(JOBOBJECT_EXTENDED_LIMIT_INFORMATION));
        IntPtr pointer = Marshal.AllocHGlobal(length);
        try
        {
            Marshal.StructureToPtr(information, pointer, false);
            if (!SetInformationJobObject(job, 9, pointer, (UInt32)length))
            {
                int error = Marshal.GetLastWin32Error();
                CloseHandle(job);
                throw new Win32Exception(error);
            }
        }
        finally { Marshal.FreeHGlobal(pointer); }
        return job;
    }

    public static void Assign(IntPtr job, int processId)
    {
        using (Process process = Process.GetProcessById(processId))
        {
            if (!AssignProcessToJobObject(job, process.Handle))
                throw new Win32Exception(Marshal.GetLastWin32Error());
        }
    }

    public static void Close(IntPtr job)
    {
        if (job != IntPtr.Zero && !CloseHandle(job))
            throw new Win32Exception(Marshal.GetLastWin32Error());
    }
}
'@

function Invoke-C12Native {
  param(
    [Parameter(Mandatory = $true)]
    [ValidateSet('docker', 'go', 'git')]
    [string]$Executable,

    [Parameter(Mandatory = $true)]
    [string[]]$Arguments,

    [Parameter(Mandatory = $true)]
    [string]$Stage,

    [Parameter(Mandatory = $true)]
    [TimeSpan]$Timeout,

    [Parameter(Mandatory = $true)]
    [string]$WorkingDirectory,

    [DateTime]$Deadline = [DateTime]::MaxValue,

    [switch]$AllowFailure
  )

  if ($Deadline -eq [DateTime]::MaxValue -and $script:c12NativeDeadline -ne [DateTime]::MaxValue) {
    $Deadline = $script:c12NativeDeadline
  }
  if ($Timeout -le [TimeSpan]::Zero) {
    throw "$Stage has a non-positive watchdog timeout"
  }
  $remaining = $Deadline - [DateTime]::UtcNow
  if ($remaining -le [TimeSpan]::Zero) {
    throw "$Stage exceeded its absolute deadline"
  }
  if ($remaining -lt $Timeout) {
    $Timeout = $remaining
  }
  $watchdogSuffix = New-C12RandomSuffix
  $watchdogRoot = Join-Path ([IO.Path]::GetTempPath()) "talenro-c12-watchdog-$watchdogSuffix"
  $pidPath = Join-Path $watchdogRoot 'host.pid'
  $pidTemporaryPath = Join-Path $watchdogRoot 'host.pid.tmp'
  $releasePath = Join-Path $watchdogRoot 'release'
  [void][IO.Directory]::CreateDirectory($watchdogRoot)
  $invocation = [pscustomobject]@{
    Executable = $Executable
    Arguments = [string[]]$Arguments
    WorkingDirectory = $WorkingDirectory
    PIDPath = $pidPath
    PIDTemporaryPath = $pidTemporaryPath
    ReleasePath = $releasePath
  }
  $job = $null
  $nativeJobHandle = [IntPtr]::Zero
  $watch = [System.Diagnostics.Stopwatch]::StartNew()
  try {
    $job = Start-Job -ArgumentList $invocation -ScriptBlock {
      param($Invocation)
      [IO.File]::WriteAllText([string]$Invocation.PIDTemporaryPath, [string]$PID)
      [IO.File]::Move([string]$Invocation.PIDTemporaryPath, [string]$Invocation.PIDPath)
      while (-not [IO.File]::Exists([string]$Invocation.ReleasePath)) {
        Start-Sleep -Milliseconds 10
      }
      Set-Location -LiteralPath ([string]$Invocation.WorkingDirectory)
      $nativeArgs = @($Invocation.Arguments | ForEach-Object { [string]$_ })
      $boundedOutput = New-Object 'System.Collections.Generic.List[string]'
      $capture = {
        process {
          if ($boundedOutput.Count -lt 4096) {
            $line = ([string]$_) -replace '[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]', '?'
            if ($line.Length -gt 512) {
              $line = $line.Substring(0, 512)
            }
            $boundedOutput.Add($line)
          }
        }
      }
      $priorNativeErrorActionPreference = $ErrorActionPreference
      try {
        $ErrorActionPreference = 'Continue'
        switch ([string]$Invocation.Executable) {
          'docker' {
            $dockerArgs = $nativeArgs
            & docker @dockerArgs 2>&1 | & $capture
            $nativeExitCode = $LASTEXITCODE
          }
          'go' {
            $goArgs = $nativeArgs
            & go @goArgs 2>&1 | & $capture
            $nativeExitCode = $LASTEXITCODE
          }
          'git' {
            $gitArgs = $nativeArgs
            & git @gitArgs 2>&1 | & $capture
            $nativeExitCode = $LASTEXITCODE
          }
          default {
            $nativeExitCode = 127
          }
        }
      }
      catch {
        $nativeExitCode = 127
      }
      finally {
        $ErrorActionPreference = $priorNativeErrorActionPreference
      }
      [pscustomobject]@{ ExitCode = [int]$nativeExitCode; Output = [string[]]$boundedOutput.ToArray() }
    }
    while (-not [IO.File]::Exists($pidPath) -and $watch.Elapsed -lt $Timeout) {
      Start-Sleep -Milliseconds 10
    }
    if (-not [IO.File]::Exists($pidPath)) {
      throw "$Stage timed out before native containment"
    }
    [int]$jobHostID = 0
    if (-not [int]::TryParse([IO.File]::ReadAllText($pidPath), [ref]$jobHostID) -or $jobHostID -le 0 -or $jobHostID -eq $PID) {
      throw "$Stage returned a malformed watchdog host PID"
    }
    $nativeJobHandle = [C12NativeJob]::CreateKillOnClose()
    [C12NativeJob]::Assign($nativeJobHandle, $jobHostID)
    [IO.File]::WriteAllText($releasePath, 'contained')
    $nativeRemaining = $Timeout - $watch.Elapsed
    if ($nativeRemaining -le [TimeSpan]::Zero) {
      [C12NativeJob]::Close($nativeJobHandle)
      $nativeJobHandle = [IntPtr]::Zero
      throw "$Stage timed out before native execution"
    }
    $waitSeconds = [int][Math]::Floor($nativeRemaining.TotalSeconds)
    if ($waitSeconds -lt 1) {
      [C12NativeJob]::Close($nativeJobHandle)
      $nativeJobHandle = [IntPtr]::Zero
      throw "$Stage has less than one bounded second remaining"
    }
    $completed = $null -ne (Wait-Job -Job $job -Timeout $waitSeconds)
    if (-not $completed) {
      [C12NativeJob]::Close($nativeJobHandle)
      $nativeJobHandle = [IntPtr]::Zero
      $null = Wait-Job -Job $job -Timeout 2
      throw "$Stage timed out after $waitSeconds seconds"
    }
    $received = @(Receive-Job -Job $job -ErrorAction SilentlyContinue)
    $result = @($received | Where-Object { $_.PSObject.Properties.Name -contains 'ExitCode' } | Select-Object -Last 1)
    if ($result.Count -ne 1) {
      throw "$Stage did not return a bounded native result"
    }
    if (-not $AllowFailure -and [int]$result[0].ExitCode -ne 0) {
      throw "$Stage failed with exit code $([int]$result[0].ExitCode)"
    }
    return [pscustomobject]@{
      ExitCode = [int]$result[0].ExitCode
      Output = @($result[0].Output | ForEach-Object { [string]$_ })
    }
  }
  finally {
    if ($nativeJobHandle -ne [IntPtr]::Zero) {
      [C12NativeJob]::Close($nativeJobHandle)
      $nativeJobHandle = [IntPtr]::Zero
    }
    if ($null -ne $job) {
      Stop-Job -Job $job -ErrorAction SilentlyContinue
      Remove-Job -Job $job -Force -ErrorAction SilentlyContinue
    }
    if ([IO.Directory]::Exists($watchdogRoot)) {
      [IO.Directory]::Delete($watchdogRoot, $true)
    }
  }
}

function Invoke-C12Git {
  param(
    [Parameter(Mandatory = $true)]
    [string[]]$Arguments,

    [Parameter(Mandatory = $true)]
    [string]$Stage,

    [Parameter(Mandatory = $true)]
    [string]$WorkingDirectory,

    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  return Invoke-C12Native -Executable 'git' -Arguments $Arguments -Stage $Stage -Timeout ([TimeSpan]::FromSeconds(30)) -WorkingDirectory $WorkingDirectory -Deadline $Deadline
}

function Remove-C12CandidateSnapshot {
  param(
    [Parameter(Mandatory = $true)]
    [string]$OwnerRoot
  )

  $resolvedParent = [IO.Path]::GetFullPath([IO.Path]::GetTempPath()).TrimEnd('\')
  $resolvedOwner = [IO.Path]::GetFullPath($OwnerRoot).TrimEnd('\')
  if ([IO.Path]::GetDirectoryName($resolvedOwner).TrimEnd('\') -cne $resolvedParent -or
      [IO.Path]::GetFileName($resolvedOwner) -notmatch '^talenro-c12-candidate-[0-9a-f]{32}$') {
    throw 'refusing cleanup of malformed C12 candidate snapshot path'
  }
  if ([IO.Directory]::Exists($resolvedOwner)) {
    $cleanupJob = Start-Job -ArgumentList $resolvedOwner -ScriptBlock {
      param($ExactOwnerRoot)
      [IO.Directory]::Delete([string]$ExactOwnerRoot, $true)
    }
    try {
      if ($null -eq (Wait-Job -Job $cleanupJob -Timeout 5)) {
        Stop-Job -Job $cleanupJob -ErrorAction SilentlyContinue
        throw 'staged candidate snapshot cleanup timed out after 5 seconds'
      }
      $null = Receive-Job -Job $cleanupJob -ErrorAction Stop
    }
    finally {
      Stop-Job -Job $cleanupJob -ErrorAction SilentlyContinue
      Remove-Job -Job $cleanupJob -Force -ErrorAction SilentlyContinue
    }
  }
}

function Copy-C12CandidateIndex {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Source,

    [Parameter(Mandatory = $true)]
    [string]$Destination
  )

  $sourceLength = (Get-Item -LiteralPath $Source).Length
  if ($sourceLength -lt 1 -or $sourceLength -gt 134217728) {
    throw 'staged candidate index exceeds the bounded 128 MiB copy limit'
  }
  $copyJob = Start-Job -ArgumentList @($Source, $Destination) -ScriptBlock {
    param($ExactSource, $ExactDestination)
    [IO.File]::Copy([string]$ExactSource, [string]$ExactDestination, $false)
  }
  try {
    if ($null -eq (Wait-Job -Job $copyJob -Timeout 5)) {
      Stop-Job -Job $copyJob -ErrorAction SilentlyContinue
      throw 'staged candidate index copy timed out after 5 seconds'
    }
    $null = Receive-Job -Job $copyJob -ErrorAction Stop
  }
  finally {
    Stop-Job -Job $copyJob -ErrorAction SilentlyContinue
    Remove-Job -Job $copyJob -Force -ErrorAction SilentlyContinue
  }
  if (-not [IO.File]::Exists($Destination) -or (Get-Item -LiteralPath $Destination).Length -ne $sourceLength) {
    throw 'staged candidate index copy did not produce one exact bounded file'
  }
}

function New-C12CandidateSnapshot {
  $suffix = New-C12RandomSuffix
  $ownerRoot = Join-Path ([IO.Path]::GetTempPath()) "talenro-c12-candidate-$suffix"
  $candidateRoot = Join-Path $ownerRoot 'tree'
  $candidateIndex = Join-Path $ownerRoot 'candidate.index'
  $candidateObjects = Join-Path $ownerRoot 'objects'
  [void][IO.Directory]::CreateDirectory($candidateRoot)
  [void][IO.Directory]::CreateDirectory($candidateObjects)
  try {
    $indexResult = Invoke-C12Git -Arguments @('rev-parse', '--path-format=absolute', '--git-path', 'index') -Stage 'locate staged candidate index' -WorkingDirectory $script:c12RepositoryRoot
    $indexLines = @($indexResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    if ($indexLines.Count -ne 1) {
      throw 'locate staged candidate index did not return one exact path'
    }
    $sourceIndex = [IO.Path]::GetFullPath([string]$indexLines[0])
    if (-not [IO.File]::Exists($sourceIndex)) {
      throw 'staged candidate index is absent'
    }
    $objectsResult = Invoke-C12Git -Arguments @('rev-parse', '--path-format=absolute', '--git-path', 'objects') -Stage 'locate common Git object directory' -WorkingDirectory $script:c12RepositoryRoot
    $objectsLines = @($objectsResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    if ($objectsLines.Count -ne 1) {
      throw 'locate common Git object directory did not return one exact path'
    }
    $commonObjects = [IO.Path]::GetFullPath([string]$objectsLines[0])
    if (-not [IO.Directory]::Exists($commonObjects) -or $commonObjects.Contains([string][IO.Path]::PathSeparator)) {
      throw 'common Git object directory is absent or is not one exact alternate path'
    }
    Copy-C12CandidateIndex -Source $sourceIndex -Destination $candidateIndex
    $priorIndex = [System.Environment]::GetEnvironmentVariable('GIT_INDEX_FILE', 'Process')
    $priorObjectDirectory = [System.Environment]::GetEnvironmentVariable('GIT_OBJECT_DIRECTORY', 'Process')
    $priorAlternateObjects = [System.Environment]::GetEnvironmentVariable('GIT_ALTERNATE_OBJECT_DIRECTORIES', 'Process')
    try {
      [System.Environment]::SetEnvironmentVariable('GIT_INDEX_FILE', $candidateIndex, 'Process')
      [System.Environment]::SetEnvironmentVariable('GIT_OBJECT_DIRECTORY', $candidateObjects, 'Process')
      [System.Environment]::SetEnvironmentVariable('GIT_ALTERNATE_OBJECT_DIRECTORIES', $commonObjects, 'Process')
      $treeResult = Invoke-C12Git -Arguments @('write-tree') -Stage 'capture staged candidate tree' -WorkingDirectory $script:c12RepositoryRoot
      $treeLines = @($treeResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
      if ($treeLines.Count -ne 1 -or [string]$treeLines[0] -notmatch '^[0-9a-f]{40}$') {
        throw 'capture staged candidate tree did not return one exact tree identity'
      }
      $tree = [string]$treeLines[0]
      $prefix = $candidateRoot.TrimEnd('\') + '\'
      $null = Invoke-C12Git -Arguments @('checkout-index', '--all', "--prefix=$prefix") -Stage 'materialize staged candidate tree' -WorkingDirectory $script:c12RepositoryRoot
    }
    finally {
      [System.Environment]::SetEnvironmentVariable('GIT_INDEX_FILE', $priorIndex, 'Process')
      [System.Environment]::SetEnvironmentVariable('GIT_OBJECT_DIRECTORY', $priorObjectDirectory, 'Process')
      [System.Environment]::SetEnvironmentVariable('GIT_ALTERNATE_OBJECT_DIRECTORIES', $priorAlternateObjects, 'Process')
    }
    return [pscustomobject]@{
      OwnerRoot = $ownerRoot
      Root = $candidateRoot
      Tree = $tree
      Index = $candidateIndex
      ObjectDirectory = $candidateObjects
      AlternateObjectDirectory = $commonObjects
    }
  }
  catch {
    Remove-C12CandidateSnapshot -OwnerRoot $ownerRoot
    throw
  }
}

function Invoke-C12Docker {
  param(
    [Parameter(Mandatory = $true)]
    [string[]]$Arguments,

    [Parameter(Mandatory = $true)]
    [string]$Stage,

    [TimeSpan]$Timeout = [TimeSpan]::FromSeconds(15),

    [DateTime]$Deadline = [DateTime]::MaxValue,

    [switch]$AllowFailure
  )

  return Invoke-C12Native -Executable 'docker' -Arguments $Arguments -Stage $Stage -Timeout $Timeout -WorkingDirectory $script:c12RepositoryRoot -Deadline $Deadline -AllowFailure:$AllowFailure
}

function Invoke-C12Go {
  param(
    [Parameter(Mandatory = $true)]
    [string[]]$Arguments,

    [Parameter(Mandatory = $true)]
    [string]$Stage,

    [TimeSpan]$Timeout = [TimeSpan]::FromMinutes(2),

    [DateTime]$Deadline = [DateTime]::MaxValue,

    [switch]$AllowFailure
  )

  return Invoke-C12Native -Executable 'go' -Arguments $Arguments -Stage $Stage -Timeout $Timeout -WorkingDirectory $script:c12RepositoryRoot -Deadline $Deadline -AllowFailure:$AllowFailure
}

function Get-C12SingleOutputLine {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Result,

    [Parameter(Mandatory = $true)]
    [string]$Stage
  )

  $lines = @($Result.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
  if ($Result.ExitCode -ne 0 -or $lines.Count -ne 1) {
    throw "$Stage did not return one bounded line"
  }
  return [string]$lines[0]
}

function Assert-C12ContainerIdentity {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix
  )

  if ([string]::IsNullOrEmpty([string]$Resource.ID)) {
    throw "container $($Resource.Kind) has no captured ID"
  }
  $dockerArgs = @(
    'container', 'inspect', '--format',
    '{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.run` }}|{{ index .Config.Labels `talenro.c12.role` }}|{{.Config.Image}}|{{.Image}}',
    [string]$Resource.ID
  )
  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "inspect $($Resource.Kind) identity"
  $identity = Get-C12SingleOutputLine -Result $result -Stage "inspect $($Resource.Kind) identity"
  $parts = $identity -split '\|', 6
  if ($parts.Count -ne 6 -or $parts[0] -cne [string]$Resource.ID -or $parts[1] -cne "/$($Resource.Name)" -or $parts[2] -cne $RunSuffix) {
    throw "container $($Resource.Kind) captured ID/name/run-label mismatch"
  }
  if ($parts[3] -cne [string]$Resource.Kind) {
    throw "container $($Resource.Kind) role mismatch"
  }
  if ($parts[4] -cne [string]$Resource.ImageRef) {
    throw "container $($Resource.Kind) image reference mismatch"
  }
  if ($parts[5] -cne [string]$Resource.ImageID) {
    throw "container $($Resource.Kind) immutable image ID mismatch"
  }
}

function Resolve-C12ImageIdentity {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource
  )

  $dockerArgs = @('image', 'inspect', '--format', '{{.Id}}', [string]$Resource.ImageRef)
  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "resolve $($Resource.Kind) immutable image"
  $imageID = Get-C12SingleOutputLine -Result $result -Stage "resolve $($Resource.Kind) immutable image"
  if ($imageID -notmatch '^sha256:[0-9a-f]{64}$') {
    throw "resolve $($Resource.Kind) immutable image returned a malformed image ID"
  }
  $Resource.ImageID = $imageID
}

function Set-C12MappedPort {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource
  )

  $dockerArgs = @('container', 'port', [string]$Resource.ID, "$($Resource.ContainerPort)/tcp")
  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "inspect $($Resource.Kind) mapped port"
  $mapping = Get-C12SingleOutputLine -Result $result -Stage "inspect $($Resource.Kind) mapped port"
  if ($mapping -notmatch '^127\.0\.0\.1:([1-9][0-9]{0,4})$') {
    throw "container $($Resource.Kind) has a non-loopback or malformed mapped port"
  }
  $port = [int]$Matches[1]
  if ($port -gt 65535) {
    throw "container $($Resource.Kind) has an invalid mapped port"
  }
  $Resource.Port = $port
}

function Start-C12Container {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix,

    [Parameter(Mandatory = $true)]
    [string]$DatabaseName,

    [Parameter(Mandatory = $true)]
    [string]$DatabasePassword
  )

  switch ($Resource.Kind) {
    'postgres' {
      $dockerArgs = @(
        'run', '--detach', '--name', [string]$Resource.Name,
        '--label', "talenro.c12.run=$RunSuffix",
        '--label', 'talenro.c12.role=postgres',
        '--publish', '127.0.0.1::5432',
        '--env', 'POSTGRES_USER=talenro',
        '--env', "POSTGRES_PASSWORD=$DatabasePassword",
        '--env', "POSTGRES_DB=$DatabaseName",
        '--health-cmd', "pg_isready -U talenro -d $DatabaseName",
        '--health-interval', '1s', '--health-timeout', '2s', '--health-retries', '60',
        [string]$Resource.ImageRef
      )
    }
    'redis' {
      $dockerArgs = @(
        'run', '--detach', '--name', [string]$Resource.Name,
        '--label', "talenro.c12.run=$RunSuffix",
        '--label', 'talenro.c12.role=redis',
        '--publish', '127.0.0.1::6379',
        [string]$Resource.ImageRef
      )
    }
    'nats' {
      $dockerArgs = @(
        'run', '--detach', '--name', [string]$Resource.Name,
        '--label', "talenro.c12.run=$RunSuffix",
        '--label', 'talenro.c12.role=nats',
        '--publish', '127.0.0.1::4222',
        [string]$Resource.ImageRef, '-js', '--name', [string]$Resource.Name
      )
    }
    default {
      throw "unknown C12 dependency kind $($Resource.Kind)"
    }
  }

  try {
    $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "start $($Resource.Kind) container"
  }
  catch {
    throw "$($_.Exception.Message); possible orphan name $($Resource.Name) was not adopted because create returned no captured ID"
  }
  $containerID = Get-C12SingleOutputLine -Result $result -Stage "start $($Resource.Kind) container"
  if ($containerID -notmatch '^[0-9a-f]{64}$') {
    throw "start $($Resource.Kind) did not return an exact container ID"
  }
  $Resource.ID = $containerID
  Assert-C12ContainerIdentity -Resource $Resource -RunSuffix $RunSuffix
  Set-C12MappedPort -Resource $Resource
}

function Test-C12TCPProtocol {
  param(
    [Parameter(Mandatory = $true)]
    [int]$Port,

    [Parameter(Mandatory = $true)]
    [string]$Request,

    [Parameter(Mandatory = $true)]
    [string]$Expected
  )

  $client = New-Object System.Net.Sockets.TcpClient
  try {
    $connection = $client.BeginConnect('127.0.0.1', $Port, $null, $null)
    if (-not $connection.AsyncWaitHandle.WaitOne(1000)) {
      return $false
    }
    $client.EndConnect($connection)
    $stream = $client.GetStream()
    $stream.ReadTimeout = 1000
    $stream.WriteTimeout = 1000
    $requestBytes = [System.Text.Encoding]::ASCII.GetBytes($Request)
    $stream.Write($requestBytes, 0, $requestBytes.Length)
    $stream.Flush()
    $buffer = New-Object byte[] 4096
    $response = New-Object System.Text.StringBuilder
    for ($attempt = 0; $attempt -lt 3; $attempt++) {
      $read = $stream.Read($buffer, 0, $buffer.Length)
      if ($read -le 0) {
        break
      }
      [void]$response.Append([System.Text.Encoding]::ASCII.GetString($buffer, 0, $read))
      if ($response.ToString().Contains($Expected)) {
        return $true
      }
    }
    return $false
  }
  catch {
    return $false
  }
  finally {
    $client.Close()
  }
}

function Wait-C12Dependencies {
  param(
    [Parameter(Mandatory = $true)]
    [object[]]$Resources
  )

  $postgres = @($Resources | Where-Object { $_.Kind -eq 'postgres' })[0]
  $redis = @($Resources | Where-Object { $_.Kind -eq 'redis' })[0]
  $nats = @($Resources | Where-Object { $_.Kind -eq 'nats' })[0]
  $deadline = [DateTime]::UtcNow.AddSeconds(60)
  while ([DateTime]::UtcNow -lt $deadline) {
    $dockerArgs = @('container', 'inspect', '--format', '{{.State.Health.Status}}', [string]$postgres.ID)
    $healthResult = Invoke-C12Docker -Arguments $dockerArgs -Stage 'inspect PostgreSQL health' -AllowFailure
    $postgresReady = $false
    if ($healthResult.ExitCode -eq 0) {
      $healthLines = @($healthResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
      $postgresReady = $healthLines.Count -eq 1 -and [string]$healthLines[0] -ceq 'healthy'
    }
    $redisReady = Test-C12TCPProtocol -Port $redis.Port -Request "*1`r`n`$4`r`nPING`r`n" -Expected "+PONG`r`n"
    $natsReady = Test-C12TCPProtocol -Port $nats.Port -Request "CONNECT {`"verbose`":false,`"pedantic`":false}`r`nPING`r`n" -Expected "PONG`r`n"
    if ($postgresReady -and $redisReady -and $natsReady) {
      return
    }
    Start-Sleep -Milliseconds 250
  }
  throw 'C12 dependencies did not pass health and protocol probes within 60 seconds'
}

function Resolve-C12CleanupIdentity {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix
  )

  if ([string]::IsNullOrEmpty([string]$Resource.ID)) {
    return $null
  }
  $dockerArgs = @(
    'container', 'inspect', '--format',
    '{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.run` }}|{{ index .Config.Labels `talenro.c12.role` }}|{{.Config.Image}}|{{.Image}}',
    [string]$Resource.ID
  )
  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "re-inspect $($Resource.Kind) before cleanup" -Timeout ([TimeSpan]::FromSeconds(5)) -AllowFailure
  if ($result.ExitCode -ne 0) {
    throw "captured $($Resource.Kind) container is absent before cleanup"
  }
  $identity = Get-C12SingleOutputLine -Result $result -Stage "re-inspect $($Resource.Kind) before cleanup"
  $parts = $identity -split '\|', 6
  if ($parts.Count -ne 6 -or
      $parts[0] -cne [string]$Resource.ID -or
      $parts[1] -cne "/$($Resource.Name)" -or
      $parts[2] -cne $RunSuffix -or
      $parts[3] -cne [string]$Resource.Kind -or
      $parts[4] -cne [string]$Resource.ImageRef -or
      $parts[5] -cne [string]$Resource.ImageID) {
    throw "refusing cleanup for $($Resource.Kind): captured ID/name/run/role/image identity mismatch"
  }
  return [string]$Resource.ID
}

function Remove-C12Container {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix
  )

  $containerID = Resolve-C12CleanupIdentity -Resource $Resource -RunSuffix $RunSuffix
  if ([string]::IsNullOrEmpty([string]$containerID)) {
    return
  }
  $dockerArgs = @('container', 'stop', '--time', '2', [string]$containerID)
  $null = Invoke-C12Docker -Arguments $dockerArgs -Stage "stop exact $($Resource.Kind) container" -Timeout ([TimeSpan]::FromSeconds(5))
  $dockerArgs = @('container', 'rm', [string]$containerID)
  $null = Invoke-C12Docker -Arguments $dockerArgs -Stage "remove exact $($Resource.Kind) container" -Timeout ([TimeSpan]::FromSeconds(5))
  $dockerArgs = @('container', 'inspect', '--format', '{{.Id}}', [string]$containerID)
  $absence = Invoke-C12Docker -Arguments $dockerArgs -Stage "verify $($Resource.Kind) cleanup" -Timeout ([TimeSpan]::FromSeconds(5)) -AllowFailure
  if ($absence.ExitCode -eq 0) {
    throw "$($Resource.Kind) container remains after exact cleanup"
  }
}

function Convert-C12RunToTests {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Pattern
  )

  if ($Pattern -match '^\^Test[A-Za-z0-9_]+\$$') {
    return @($Pattern.Substring(1, $Pattern.Length - 2))
  }
  if ($Pattern -match '^\^\(Test[A-Za-z0-9_]+(?:\|Test[A-Za-z0-9_]+)+\)\$$') {
    return @(($Pattern.Substring(2, $Pattern.Length - 4)) -split '\|')
  }
  throw '-Run must be an anchored literal Test name or anchored alternation of literal Test names'
}

function ConvertTo-C12RunPattern {
  param(
    [Parameter(Mandatory = $true)]
    [string[]]$Tests
  )

  if ($Tests.Count -eq 1) {
    return '^' + $Tests[0] + '$'
  }
  return '^(' + ($Tests -join '|') + ')$'
}

function Resolve-C12FocusedTestMap {
  param(
    [Parameter(Mandatory = $true)]
    [string[]]$Packages,

    [Parameter(Mandatory = $true)]
    [string[]]$RequestedTests
  )

  $goArgs = @(
    'test', '-run', '^TestC12ResolveFocusedIntegrationTests$', '-count=1', '-v', './internal/testinfra',
    '-args', '-c12-focused-packages', ($Packages -join '|'), '-c12-focused-tests', ($RequestedTests -join '|')
  )
  $result = Invoke-C12Go -Arguments $goArgs -Stage 'resolve focused package-local integration tests' -Timeout ([TimeSpan]::FromMinutes(2))
  $prefix = 'C12_FOCUSED_MAP:'
  $mappingLines = @($result.Output | Where-Object { ([string]$_).StartsWith($prefix, [StringComparison]::Ordinal) })
  if ($mappingLines.Count -ne 1) {
    throw 'focused test resolver did not return one bounded package map'
  }
  try {
    $mapping = ([string]$mappingLines[0]).Substring($prefix.Length) | ConvertFrom-Json
  }
  catch {
    throw 'focused test resolver returned malformed JSON'
  }
  $resolved = @{}
  $covered = @{}
  foreach ($package in $Packages) {
    $property = @($mapping.PSObject.Properties | Where-Object { $_.Name -ceq $package })
    if ($property.Count -ne 1) {
      throw "focused test resolver omitted exact package $package"
    }
    $localTests = @($property[0].Value | ForEach-Object { [string]$_ })
    if ($localTests.Count -lt 1 -or $localTests.Count -gt 512) {
      throw "focused test resolver returned an invalid bounded test set for $package"
    }
    foreach ($testName in $localTests) {
      if ($testName -notmatch '^Test[A-Za-z0-9_]+$' -or $covered.ContainsKey($testName)) {
        throw "focused test resolver returned a malformed or ambiguous test $testName"
      }
      $covered[$testName] = $package
    }
    $resolved[$package] = $localTests
  }
  foreach ($testName in $RequestedTests) {
    if (-not $covered.ContainsKey($testName)) {
      throw "focused test resolver did not cover requested test $testName"
    }
  }
  if ($covered.Count -ne $RequestedTests.Count) {
    throw 'focused test resolver returned extra tests'
  }
  return $resolved
}

function Assert-C12GoJSONResult {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Result,

    [Parameter(Mandatory = $true)]
    [string]$Package,

    [string[]]$ExpectedTests = @()
  )

  $expected = @{}
  foreach ($testName in $ExpectedTests) {
    $expected[$testName] = 0
  }
  $observedParents = @{}
  $topLevelTerminals = @{}
  $descendants = @{}
  $topLevelPasses = 0
  $packagePasses = 0
  foreach ($line in $Result.Output) {
    if ([string]::IsNullOrWhiteSpace([string]$line)) {
      continue
    }
    try {
      $event = [string]$line | ConvertFrom-Json
    }
    catch {
      throw "go test for $Package emitted a non-JSON line"
    }
    $hasTest = $event.PSObject.Properties.Name -contains 'Test'
    $testName = if ($hasTest) { [string]$event.Test } else { '' }
    $eventPackage = if ($event.PSObject.Properties.Name -contains 'Package') { [string]$event.Package } else { '' }
    if (-not [string]::IsNullOrEmpty($eventPackage) -and $eventPackage -cne $Package) {
      throw "go test emitted an event for unexpected package $eventPackage"
    }
    if (-not $hasTest) {
      if ($event.Action -eq 'skip') {
        throw "go test for $Package skipped the package"
      }
      if ($event.Action -eq 'fail') {
        throw "go test for $Package failed the package"
      }
      if ($event.Action -eq 'pass') {
        $packagePasses++
      }
      continue
    }

    if ($testName -match '/') {
      $parts = @($testName -split '/')
      $parent = [string]$parts[0]
      if (($expected.Count -gt 0 -and -not $expected.ContainsKey($parent)) -or
          ($expected.Count -eq 0 -and -not $observedParents.ContainsKey($parent))) {
        throw "go test for $Package emitted orphan descendant $testName"
      }
      $depth = $parts.Count - 1
      if ($depth -gt 4) {
        throw "go test for $Package descendant $testName exceeds depth 4"
      }
      $suffixLength = $testName.Length - $parent.Length - 1
      if ($suffixLength -lt 1 -or $suffixLength -gt 256) {
        throw "go test for $Package descendant $testName has malformed bounded suffix length"
      }
      for ($index = 1; $index -lt $parts.Count; $index++) {
        if ([string]$parts[$index] -notmatch '^[A-Za-z0-9_.-]{1,64}$') {
          throw "go test for $Package descendant $testName has malformed suffix"
        }
      }
      if (-not $descendants.ContainsKey($testName)) {
        if ($descendants.Count -ge 256) {
          throw "go test for $Package exceeds the bounded descendant count 256"
        }
        $descendants[$testName] = [pscustomobject]@{ Passes = 0; Skips = 0; Fails = 0 }
      }
      $state = $descendants[$testName]
      switch ([string]$event.Action) {
        'pass' { $state.Passes = [int]$state.Passes + 1 }
        'skip' { $state.Skips = [int]$state.Skips + 1 }
        'fail' { $state.Fails = [int]$state.Fails + 1 }
      }
      continue
    }

    $observedParents[$testName] = $true
    if ($expected.Count -gt 0 -and -not $expected.ContainsKey($testName)) {
      throw "go test for $Package emitted extra top-level test $testName"
    }
    if ($event.Action -eq 'skip') {
      throw "go test for $Package skipped $testName"
    }
    if ($event.Action -eq 'fail') {
      throw "go test for $Package failed $testName"
    }
    if ($event.Action -eq 'pass') {
      if (-not $topLevelTerminals.ContainsKey($testName)) {
        $topLevelTerminals[$testName] = 0
      }
      $topLevelTerminals[$testName] = [int]$topLevelTerminals[$testName] + 1
      if ($expected.ContainsKey($testName)) {
        $expected[$testName] = [int]$expected[$testName] + 1
      }
      $topLevelPasses++
    }
  }
  foreach ($testName in $expected.Keys) {
    if ([int]$expected[$testName] -ne 1) {
      throw "go test for $Package has missing or duplicate terminal pass for $testName"
    }
  }
  foreach ($testName in $topLevelTerminals.Keys) {
    if ([int]$topLevelTerminals[$testName] -ne 1) {
      throw "go test for $Package has missing or duplicate terminal pass for $testName"
    }
  }
  foreach ($testName in $descendants.Keys) {
    $state = $descendants[$testName]
    if ([int]$state.Skips -ne 0) {
      throw "go test for $Package skipped descendant $testName"
    }
    if ([int]$state.Fails -ne 0) {
      throw "go test for $Package failed descendant $testName"
    }
    if ([int]$state.Passes -eq 0) {
      throw "go test for $Package has missing terminal pass for descendant $testName"
    }
    if ([int]$state.Passes -ne 1) {
      throw "go test for $Package has duplicate terminal pass for descendant $testName"
    }
  }
  if ($topLevelPasses -eq 0 -or $packagePasses -ne 1 -or $Result.ExitCode -ne 0) {
    throw "go test for $Package did not produce an exact passing JSON result"
  }
}

function Assert-C12ExecutionPlan {
  param(
    [Parameter(Mandatory = $true)]
    [object[]]$Groups
  )

  if ($Groups.Count -lt 1 -or $Groups.Count -gt 256) {
    throw 'suite execution plan has an invalid bounded group count'
  }
  $priorID = ''
  foreach ($group in $Groups) {
    $groupID = [string]$group.id
    $package = [string]$group.package
    $profile = [string]$group.profile
    $timeout = [string]$group.timeout
    $tests = @($group.tests | ForEach-Object { [string]$_ })
    if ($groupID -notmatch '^[a-z][a-z0-9-]{0,63}$' -or
        (-not [string]::IsNullOrEmpty($priorID) -and [string]::CompareOrdinal($priorID, $groupID) -ge 0)) {
      throw "invalid suite execution-plan group ID $groupID"
    }
    $priorID = $groupID
    if (-not $script:c12AllowedPackages.ContainsKey($package)) {
      throw "suite execution-plan package $package is outside the Task 4 allowed set"
    }
    if ($profile -cne 'base') {
      throw "suite execution-plan group $groupID has unsupported Task 4 profile"
    }
    $groupDuration = ConvertFrom-C12Duration -Value $timeout
    if ($groupDuration -gt [TimeSpan]::FromMinutes(30)) {
      throw "suite execution-plan group $groupID exceeds the 30m group timeout"
    }
    if ($tests.Count -lt 1 -or $tests.Count -gt 512) {
      throw "suite execution-plan group $groupID has an invalid bounded test count"
    }
    $priorTest = ''
    foreach ($testName in $tests) {
      if ($testName -notmatch '^Test[A-Za-z0-9_]+$' -or
          (-not [string]::IsNullOrEmpty($priorTest) -and [string]::CompareOrdinal($priorTest, $testName) -ge 0)) {
        throw "suite execution-plan group $groupID contains an invalid, duplicate, or unsorted literal test"
      }
      $priorTest = $testName
    }
  }
}

function Invoke-C12Group {
  param(
    [Parameter(Mandatory = $true)]
    [string]$GroupID,

    [Parameter(Mandatory = $true)]
    [string]$GroupProfile,

    [Parameter(Mandatory = $true)]
    [string]$Package,

    [string]$RunPattern = '',

    [Parameter(Mandatory = $true)]
    [string]$GroupTimeout,

    [string[]]$ExpectedTests = @()
  )

  if ($GroupProfile -cne 'base') {
    throw 'Task 4 runner accepts only the base profile'
  }
  $groupDuration = ConvertFrom-C12Duration -Value $GroupTimeout
  $groupDeadline = [DateTime]::UtcNow.Add($groupDuration).AddMinutes(3)
  if ($script:c12SuiteDeadline -lt $groupDeadline) {
    $groupDeadline = $script:c12SuiteDeadline
  }
  $priorNativeDeadline = $script:c12NativeDeadline
  $script:c12NativeDeadline = $groupDeadline
  $runSuffix = New-C12RandomSuffix
  if ($runSuffix -notmatch '^[0-9a-f]{32}$') {
    throw 'cryptographic C12 run suffix is malformed'
  }
  $databaseName = "talenro_c12_$runSuffix"
  $databasePassword = New-C12DatabasePassword -RunSuffix $runSuffix
  $resources = @(
    [pscustomobject]@{ Kind = 'postgres'; Name = "talenro-c12-$runSuffix-postgres"; ID = ''; ImageRef = 'postgres:18.4-alpine3.23'; ImageID = ''; Port = 0; ContainerPort = 5432 },
    [pscustomobject]@{ Kind = 'redis'; Name = "talenro-c12-$runSuffix-redis"; ID = ''; ImageRef = 'redis:8.8.1-alpine3.23'; ImageID = ''; Port = 0; ContainerPort = 6379 },
    [pscustomobject]@{ Kind = 'nats'; Name = "talenro-c12-$runSuffix-nats"; ID = ''; ImageRef = 'nats:2.14.3-alpine3.22'; ImageID = ''; Port = 0; ContainerPort = 4222 }
  )
  $primaryFailure = $null
  $cleanupFailures = @()
  try {
    foreach ($resource in $resources) {
      Resolve-C12ImageIdentity -Resource $resource
    }
    foreach ($resource in $resources) {
      Start-C12Container -Resource $resource -RunSuffix $runSuffix -DatabaseName $databaseName -DatabasePassword $databasePassword
    }
    Wait-C12Dependencies -Resources $resources
    $postgres = @($resources | Where-Object { $_.Kind -eq 'postgres' })[0]
    $redis = @($resources | Where-Object { $_.Kind -eq 'redis' })[0]
    $nats = @($resources | Where-Object { $_.Kind -eq 'nats' })[0]
    $databaseURL = "postgres://talenro:$databasePassword@127.0.0.1:$($postgres.Port)/$databaseName`?sslmode=disable&application_name=talenro-c12-$runSuffix"
    [System.Environment]::SetEnvironmentVariable('TALENRO_DATABASE_URL', $databaseURL, 'Process')
    [System.Environment]::SetEnvironmentVariable('TALENRO_REDIS_ADDRESS', "127.0.0.1:$($redis.Port)", 'Process')
    [System.Environment]::SetEnvironmentVariable('TALENRO_NATS_URL', "nats://127.0.0.1:$($nats.Port)", 'Process')

    $migrationDirectory = Join-Path $script:c12RepositoryRoot 'db\migrations'
    $goArgs = @('tool', 'goose', '-dir', $migrationDirectory, 'postgres', $databaseURL, 'up-to', '6')
    $null = Invoke-C12Go -Arguments $goArgs -Stage "ordinary Goose base migration for $GroupID" -Deadline $groupDeadline

    $goArgs = @('test', '-json', '-tags=integration', '-p=1', '-count=1', '-timeout', $GroupTimeout)
    if (-not [string]::IsNullOrEmpty($RunPattern)) {
      $goArgs += @('-run', $RunPattern)
    }
    $goArgs += $Package
    $testResult = Invoke-C12Go -Arguments $goArgs -Stage "tagged Go test group $GroupID" -Timeout $groupDuration -Deadline $groupDeadline -AllowFailure
    Assert-C12GoJSONResult -Result $testResult -Package $Package -ExpectedTests $ExpectedTests
  }
  catch {
    $primaryFailure = $_.Exception.Message
  }
  finally {
    [System.Environment]::SetEnvironmentVariable('TALENRO_DATABASE_URL', $null, 'Process')
    [System.Environment]::SetEnvironmentVariable('TALENRO_REDIS_ADDRESS', $null, 'Process')
    [System.Environment]::SetEnvironmentVariable('TALENRO_NATS_URL', $null, 'Process')
    for ($index = $resources.Count - 1; $index -ge 0; $index--) {
      try {
        Remove-C12Container -Resource $resources[$index] -RunSuffix $runSuffix
      }
      catch {
        $cleanupFailures += "$($resources[$index].Kind): $($_.Exception.Message)"
      }
    }
  }
  if ($null -ne $primaryFailure) {
    [Console]::Error.WriteLine("C12 test failure [$GroupID]: $primaryFailure")
  }
  foreach ($cleanupFailure in $cleanupFailures) {
    [Console]::Error.WriteLine("C12 cleanup failure [$GroupID]: $cleanupFailure")
  }
  $script:c12NativeDeadline = $priorNativeDeadline
  if ($null -ne $primaryFailure -or $cleanupFailures.Count -ne 0) {
    throw "C12 group $GroupID failed"
  }
  Write-Output "C12 group $GroupID passed"
}

function Assert-C12NoInheritedDependencies {
  foreach ($name in @('TALENRO_DATABASE_URL', 'TALENRO_REDIS_ADDRESS', 'TALENRO_NATS_URL')) {
    if (-not [string]::IsNullOrEmpty([System.Environment]::GetEnvironmentVariable($name, 'Process'))) {
      throw 'inherited dependency endpoints are forbidden'
    }
  }
  foreach ($name in @('TALENRO_INSTALLATION_KIND', 'TALENRO_ENVIRONMENT')) {
    if (-not [string]::IsNullOrEmpty([System.Environment]::GetEnvironmentVariable($name, 'Process'))) {
      throw 'inherited production markers are forbidden'
    }
  }
  foreach ($name in @('GIT_INDEX_FILE', 'GIT_OBJECT_DIRECTORY', 'GIT_ALTERNATE_OBJECT_DIRECTORIES')) {
    if (-not [string]::IsNullOrEmpty([System.Environment]::GetEnvironmentVariable($name, 'Process'))) {
      throw 'inherited Git candidate overrides are forbidden'
    }
  }
}

function Invoke-C12FocusedMode {
  if ($Profile -cne 'base') {
    throw 'Task 4 runner accepts only the base profile'
  }
  if ($Race) {
    throw 'Task 4 base profile does not support -Race while CGO is disabled'
  }
  $duration = ConvertFrom-C12Duration -Value $Timeout
  if ($duration -gt [TimeSpan]::FromMinutes(30)) {
    throw 'focused timeout must not exceed 30m'
  }
  if ($Packages -notmatch '^\./[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*(?:\|\./[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*)*$' -or $Packages.Contains('./...')) {
    throw 'invalid -Packages: expected pipe-delimited explicit ./package tokens'
  }
  $packageList = @($Packages -split '\|')
  $seen = @{}
  foreach ($package in $packageList) {
    if ($seen.ContainsKey($package)) {
      throw "duplicate package $package"
    }
    $seen[$package] = $true
    if (-not $script:c12AllowedPackages.ContainsKey($package)) {
      throw "unknown package $package"
    }
  }
  $expectedTests = @()
  $testMap = @{}
  if (-not [string]::IsNullOrEmpty($Run)) {
    $expectedTests = @(Convert-C12RunToTests -Pattern $Run)
    $testMap = Resolve-C12FocusedTestMap -Packages $packageList -RequestedTests $expectedTests
  }
  foreach ($package in $packageList) {
    $groupID = 'focused-' + ($package.TrimStart('.').TrimStart('/').Replace('/', '-'))
    $localTests = @()
    if ($testMap.ContainsKey($package)) {
      $localTests = @($testMap[$package])
    }
    $localRun = if ($localTests.Count -gt 0) { ConvertTo-C12RunPattern -Tests $localTests } else { '' }
    Invoke-C12Group -GroupID $groupID -GroupProfile 'base' -Package $package -RunPattern $localRun -GroupTimeout $Timeout -ExpectedTests $localTests
  }
}

function Invoke-C12SuiteMode {
  if ($Suite -cne 'batch01') {
    throw 'Task 4 runner accepts only the batch01 suite'
  }
  $duration = ConvertFrom-C12Duration -Value $Timeout
  if ($duration -ne [TimeSpan]::FromMinutes(120)) {
    throw 'batch01 suite timeout must be exactly 120m'
  }
  $priorNativeDeadline = $script:c12NativeDeadline
  $script:c12NativeDeadline = $script:c12SuiteDeadline
  $originalRoot = $script:c12RepositoryRoot
  $priorCandidateGitEnvironment = @{
    GIT_INDEX_FILE = [System.Environment]::GetEnvironmentVariable('GIT_INDEX_FILE', 'Process')
    GIT_OBJECT_DIRECTORY = [System.Environment]::GetEnvironmentVariable('GIT_OBJECT_DIRECTORY', 'Process')
    GIT_ALTERNATE_OBJECT_DIRECTORIES = [System.Environment]::GetEnvironmentVariable('GIT_ALTERNATE_OBJECT_DIRECTORIES', 'Process')
  }
  $candidate = $null
  try {
    $candidate = New-C12CandidateSnapshot
    $script:c12RepositoryRoot = [string]$candidate.Root
    [System.Environment]::SetEnvironmentVariable('GIT_INDEX_FILE', [string]$candidate.Index, 'Process')
    [System.Environment]::SetEnvironmentVariable('GIT_OBJECT_DIRECTORY', [string]$candidate.ObjectDirectory, 'Process')
    [System.Environment]::SetEnvironmentVariable('GIT_ALTERNATE_OBJECT_DIRECTORIES', [string]$candidate.AlternateObjectDirectory, 'Process')
    $manifestRelativePath = 'testdata/c12/integration-contracts-schema-authority.v1.json'
    $manifestPath = Join-Path $script:c12RepositoryRoot ($manifestRelativePath.Replace('/', '\'))
    if (-not (Test-Path -LiteralPath $manifestPath -PathType Leaf)) {
      throw 'canonical Batch 01 manifest is absent from the staged candidate'
    }

    $goArgs = @(
      'test', '-run', '^TestC12SelectedIntegrationManifest$', '-count=1', './internal/testinfra',
      '-args', '-c12-suite', 'batch01', '-c12-manifests', $manifestRelativePath,
      '-c12-suite-timeout', $Timeout, '-c12-candidate-tree', [string]$candidate.Tree
    )
    $null = Invoke-C12Go -Arguments $goArgs -Stage 'validate canonical Batch 01 integration manifest'
    $manifest = Get-Content -Raw -LiteralPath $manifestPath | ConvertFrom-Json
    $groups = @($manifest.groups)
    Assert-C12ExecutionPlan -Groups $groups
    foreach ($group in $groups) {
      $tests = @($group.tests | ForEach-Object { [string]$_ })
      $runPattern = '^(' + ($tests -join '|') + ')$'
      Invoke-C12Group -GroupID ([string]$group.id) -GroupProfile ([string]$group.profile) -Package ([string]$group.package) -RunPattern $runPattern -GroupTimeout ([string]$group.timeout) -ExpectedTests $tests
    }
  }
  finally {
    $script:c12RepositoryRoot = $originalRoot
    foreach ($name in $priorCandidateGitEnvironment.Keys) {
      [System.Environment]::SetEnvironmentVariable($name, $priorCandidateGitEnvironment[$name], 'Process')
    }
    if ($null -ne $candidate) {
      Remove-C12CandidateSnapshot -OwnerRoot ([string]$candidate.OwnerRoot)
    }
    $script:c12NativeDeadline = $priorNativeDeadline
  }
}

$script:c12RepositoryRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$selectedMode = $PSCmdlet.ParameterSetName
$priorGoEnvironment = @{
  GOOS = [System.Environment]::GetEnvironmentVariable('GOOS', 'Process')
  GOARCH = [System.Environment]::GetEnvironmentVariable('GOARCH', 'Process')
  CGO_ENABLED = [System.Environment]::GetEnvironmentVariable('CGO_ENABLED', 'Process')
}
$scriptExitCode = 0
try {
  Assert-C12NoInheritedDependencies
  [System.Environment]::SetEnvironmentVariable('GOOS', 'windows', 'Process')
  [System.Environment]::SetEnvironmentVariable('GOARCH', 'amd64', 'Process')
  [System.Environment]::SetEnvironmentVariable('CGO_ENABLED', '0', 'Process')
  if ($selectedMode -eq 'Focused') {
    Invoke-C12FocusedMode
  }
  else {
    Invoke-C12SuiteMode
  }
}
catch {
  [Console]::Error.WriteLine($_.Exception.Message)
  $scriptExitCode = 1
}
finally {
  foreach ($name in $priorGoEnvironment.Keys) {
    [System.Environment]::SetEnvironmentVariable($name, $priorGoEnvironment[$name], 'Process')
  }
}
exit $scriptExitCode
