$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'run.ps1')
function Check-Confirmation([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
Add-Type @'
using System;
using System.Runtime.InteropServices;
public static class ConfirmationTestJob {
 [DllImport("kernel32.dll",SetLastError=true)] public static extern IntPtr CreateJobObject(IntPtr a,string n);
 [DllImport("kernel32.dll",SetLastError=true)] public static extern bool AssignProcessToJobObject(IntPtr j,IntPtr p);
 [DllImport("kernel32.dll",SetLastError=true)] public static extern bool IsProcessInJob(IntPtr p,IntPtr j,out bool owned);
 [DllImport("kernel32.dll")] public static extern bool TerminateJobObject(IntPtr j,uint code);
 [DllImport("kernel32.dll")] public static extern bool TerminateProcess(IntPtr p,uint code);
 [DllImport("kernel32.dll")] public static extern bool CloseHandle(IntPtr h);
}
'@
function Test-Confirmation([bool]$Exited) {
    $job=[ConfirmationTestJob]::CreateJobObject([IntPtr]::Zero,$null)
    if($job -eq [IntPtr]::Zero){throw 'Test job creation failed'}
    $process=New-Object Diagnostics.Process
    $handle=[IntPtr]::Zero;$assigned=$false
    try {
        $process.StartInfo.FileName=(Get-Process -Id $PID).Path
        $process.StartInfo.Arguments='-NoProfile -NonInteractive -Command "[Threading.Thread]::Sleep(30000)"'
        if($Exited){$process.StartInfo.Arguments='-NoProfile -NonInteractive -Command "[Threading.Thread]::Sleep(1000); exit 7"'}
        $process.StartInfo.UseShellExecute=$false;$process.StartInfo.CreateNoWindow=$true
        if(-not $process.Start()){throw 'Test child creation failed'}
        $handle=$process.Handle
        if(-not [ConfirmationTestJob]::AssignProcessToJobObject($job,$handle)){throw 'Test child ownership assignment failed'}
        $assigned=$true
        $owned=$false
        if(-not [ConfirmationTestJob]::IsProcessInJob($handle,$job,[ref]$owned) -or -not $owned){throw 'Test child ownership unverified'}
        $row=New-Object 'IsolatedFuzzProcess+MemberObservation'
        $row.Id=$process.Id;$row.Owned=$owned;$row.State='running';$row.ExitCodeRead=$true;$row.ExitCode=259;$row.ImageError=5
        if($Exited -and -not $process.WaitForExit(10000)){throw 'Test child did not finish within test bound'}
        [IsolatedFuzzProcess]::ConfirmMemberExit($handle,$row)
        Check-Confirmation ($row.InitialState -eq 'running' -and $row.InitialExitCode -eq 259 -and $row.InitialExitCodeRead) 'first wait and exit-code read preserved'
        Check-Confirmation ($row.ConfirmationAttempted -and $row.ConfirmationUTC -and $row.ImageError -eq 5) 'one confirmation recorded without erasing image failure'
        if($Exited) {
            Check-Confirmation ($row.State -eq 'exited' -and $row.ConfirmationState -eq 'exited' -and $row.ExitCode -eq 7 -and $row.QueryError -eq 0) 'same signaled handle refreshes stale exit code'
            Check-Confirmation ([IsolatedFuzzProcess]::IsShutdownStateVerified($row)) 'confirmed exit satisfies state policy despite missing image'
        } else {
            Check-Confirmation ($row.State -eq 'running' -and $row.ConfirmationState -eq 'running') 'unsignaled handle remains running'
            Check-Confirmation (-not [IsolatedFuzzProcess]::IsShutdownStateVerified($row)) 'still-running missing image remains rejected'
        }
        $refused=$false;try {[IsolatedFuzzProcess]::ConfirmMemberExit($handle,$row)}catch {$refused=$true}
        Check-Confirmation $refused 'second confirmation cannot become a retry loop'
    } finally {
        if($handle -ne [IntPtr]::Zero -and -not $assigned){[ConfirmationTestJob]::TerminateProcess($handle,0xE003) | Out-Null}
        [ConfirmationTestJob]::TerminateJobObject($job,0xE003) | Out-Null
        [ConfirmationTestJob]::CloseHandle($job) | Out-Null
        $process.Dispose()
    }
}
Test-Confirmation $true
Test-Confirmation $false
$row=New-Object 'IsolatedFuzzProcess+MemberObservation';$row.State='running';$row.ExitCodeRead=$true
$refused=$false;try {[IsolatedFuzzProcess]::ConfirmMemberExit([IntPtr]::Zero,$row)}catch {$refused=$true}
Check-Confirmation ($refused -and -not $row.ConfirmationAttempted) 'unowned observation cannot query a process handle'
