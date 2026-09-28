# Private to standalone devtools entries. Loading only defines functions.
function Restore-DevtoolsExecutionEnvironment {
  param([hashtable]$Saved)
  foreach ($item in @(Get-ChildItem Env: | Where-Object { $_.Name -match '^GO' })) {
    Remove-Item -LiteralPath ('Env:' + $item.Name)
  }
  foreach ($key in $Saved.Go.Keys) {
    [Environment]::SetEnvironmentVariable($key, $Saved.Go[$key], 'Process')
  }
  $env:PATH = $Saved.Path
}

function Initialize-DevtoolsExecutionEnvironment {
  # The verifier's child restores its own environment. Bind its consumers too,
  # including Go resolved by Buf's nested local plugin command.
  if ($env:GOFLAGS -match '(^|\s)-(modfile|overlay)(=|\s|$)') { throw 'module override' }
  $cache = Join-Path $env:USERPROFILE 'go/pkg/mod'
  $bin = Join-Path $cache 'golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin'
  if (-not [IO.File]::Exists((Join-Path $bin 'go.exe'))) { throw 'missing fixed toolchain' }
  $saved = @{ Go = @{}; Path = $env:PATH }
  foreach ($item in @(Get-ChildItem Env: | Where-Object { $_.Name -match '^GO' })) {
    $saved.Go[$item.Name] = $item.Value
  }
  try {
    foreach ($key in $saved.Go.Keys) { Remove-Item -LiteralPath ('Env:' + $key) }
    $fixed = @{
      GOWORK='off'; GOENV='off'; GOTOOLCHAIN='local'; GOPROXY='off'; GOSUMDB='off';
      GOAUTH='off'; GOVCS='all:off'; GOFLAGS='-mod=readonly';
      GOPATH=(Join-Path $env:USERPROFILE 'go'); GOMODCACHE=$cache
    }
    foreach ($key in $fixed.Keys) { [Environment]::SetEnvironmentVariable($key, $fixed[$key], 'Process') }
    $env:PATH = $bin + [IO.Path]::PathSeparator + $saved.Path
    return $saved
  } catch {
    Restore-DevtoolsExecutionEnvironment -Saved $saved
    throw
  }
}

function Initialize-DevtoolsProcessOwnership {
  # Keep this unnamed, non-inherited handle alive until the owning host exits.
  # PowerShell 7.6.5 compiles this definition in-process; public entries reject
  # other runtimes before loading this file or starting any work.
  Add-Type -TypeDefinition @'
using System;
using System.ComponentModel;
using System.Runtime.InteropServices;
public static class DevtoolsProcessOwnership {
  private static IntPtr job;
  [StructLayout(LayoutKind.Sequential)] private struct Basic {
    public long processTime, jobTime;
    public uint flags;
    public UIntPtr minimum, maximum;
    public uint active;
    public UIntPtr affinity;
    public uint priority, scheduling;
  }
  [StructLayout(LayoutKind.Sequential)] private struct Counters {
    public ulong readOperations, writeOperations, otherOperations, readBytes, writeBytes, otherBytes;
  }
  [StructLayout(LayoutKind.Sequential)] private struct Extended {
    public Basic basic; public Counters io;
    public UIntPtr processMemory, jobMemory, peakProcess, peakJob;
  }
  [DllImport("kernel32.dll", SetLastError=true)] private static extern IntPtr CreateJobObject(IntPtr attributes, string name);
  [DllImport("kernel32.dll", SetLastError=true)] private static extern bool SetInformationJobObject(IntPtr job, int kind, ref Extended data, uint size);
  [DllImport("kernel32.dll", SetLastError=true)] private static extern bool AssignProcessToJobObject(IntPtr job, IntPtr process);
  [DllImport("kernel32.dll")] private static extern IntPtr GetCurrentProcess();
  [DllImport("kernel32.dll")] private static extern bool CloseHandle(IntPtr handle);
  public static void Initialize() {
    if (job != IntPtr.Zero) throw new InvalidOperationException("already initialized");
    IntPtr created = CreateJobObject(IntPtr.Zero, null);
    if (created == IntPtr.Zero) throw new Win32Exception();
    bool assigned = false;
    try {
      var limits = new Extended(); limits.basic.flags = 0x2000;
      if (!SetInformationJobObject(created, 9, ref limits, (uint)Marshal.SizeOf(typeof(Extended)))) throw new Win32Exception();
      if (!AssignProcessToJobObject(created, GetCurrentProcess())) throw new Win32Exception();
      job = created; assigned = true;
    } finally { if (!assigned) CloseHandle(created); }
  }
}
'@ -ErrorAction Stop
  [DevtoolsProcessOwnership]::Initialize()
}
