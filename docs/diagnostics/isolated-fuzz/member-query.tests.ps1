$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'run.ps1')
function Check-MemberQuery([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
Add-Type @'
using System;
using System.Runtime.InteropServices;
public static class MemberQueryTestJob {
 [DllImport("kernel32.dll",SetLastError=true)] public static extern IntPtr CreateJobObject(IntPtr a,string n);
 [DllImport("kernel32.dll",SetLastError=true)] public static extern bool AssignProcessToJobObject(IntPtr j,IntPtr p);
 [DllImport("kernel32.dll")] public static extern bool TerminateJobObject(IntPtr j,uint code);
 [DllImport("kernel32.dll")] public static extern bool TerminateProcess(IntPtr p,uint code);
 [DllImport("kernel32.dll")] public static extern bool CloseHandle(IntPtr h);
}
'@
$inspect=[IsolatedFuzzProcess].GetMethod('InspectMembers',[Reflection.BindingFlags]'NonPublic,Static')
$job=[MemberQueryTestJob]::CreateJobObject([IntPtr]::Zero,$null)
$other=[MemberQueryTestJob]::CreateJobObject([IntPtr]::Zero,$null)
$process=New-Object Diagnostics.Process;$handle=[IntPtr]::Zero;$assigned=$false
try {
    if($job -eq [IntPtr]::Zero -or $other -eq [IntPtr]::Zero){throw 'Fixture job creation failed'}
    $process.StartInfo.FileName=(Get-Process -Id $PID).Path
    $process.StartInfo.Arguments='-NoProfile -NonInteractive -Command "[Threading.Thread]::Sleep(30000)"'
    $process.StartInfo.UseShellExecute=$false;$process.StartInfo.CreateNoWindow=$true
    if(-not $process.Start()){throw 'Fixture child creation failed'}
    $handle=$process.Handle
    if(-not [MemberQueryTestJob]::AssignProcessToJobObject($job,$handle)){throw 'Fixture assignment failed'}
    $assigned=$true
    # The live child and creation handle stay retained throughout; no archived PID.
    $rows=$inspect.Invoke($null,[object[]]@([IntPtr](-1),[UInt64[]]@($process.Id),$null,$null))
    $failed=$rows[0]
    Check-MemberQuery ($failed.QueryError -eq 6 -and -not $failed.Owned -and $failed.State -eq 'unknown') 'real invalid-job query preserves native error and unverified state'
    Check-MemberQuery ($null -ne $failed.PSObject.Properties['ProcessOpen'] -and $null -ne $failed.PSObject.Properties['JobMembership']) 'failed acquisition exposes both API phase records'
    Check-MemberQuery ($failed.ProcessOpen.API -ceq 'OpenProcess' -and $failed.ProcessOpen.Succeeded -and $failed.ProcessOpen.Error -eq 0) 'successful handle open stays separate from ownership failure'
    Check-MemberQuery ($failed.JobMembership.API -ceq 'IsProcessInJob' -and -not $failed.JobMembership.Succeeded -and $failed.JobMembership.Error -eq 6) 'failing ownership API retains original error before diagnostic bookkeeping'
    foreach($query in @($failed.ProcessOpen,$failed.JobMembership)) {
        Check-MemberQuery ($query.StartUTC -and $query.EndUTC -and $query.Frequency -gt 0 -and $query.EndTimestamp -ge $query.StartTimestamp) ('before after and monotonic timing retained for '+$query.API)
    }
    Check-MemberQuery ($failed.ProcessOpen.EndTimestamp -le $failed.JobMembership.StartTimestamp) 'acquisition phase order survives serialization timing overhead'
    Check-MemberQuery (-not $failed.ExitCodeRead -and $null -eq $failed.FinalShutdown -and -not [IsolatedFuzzProcess]::IsFinalShutdownStateVerified($failed)) 'phase diagnostics never upgrade failed membership into completion'
    $saved=$failed|ConvertTo-Json -Depth 6|ConvertFrom-Json
    Check-MemberQuery ($saved.ProcessOpen.API -ceq 'OpenProcess' -and $saved.JobMembership.Error -eq 6 -and $saved.QueryError -eq 6) 'nested API evidence survives receipt-style JSON roundtrip'
    $rows=$inspect.Invoke($null,[object[]]@($job,[UInt64[]]@($process.Id),$null,$null));$owned=$rows[0]
    Check-MemberQuery ($owned.ProcessOpen.Succeeded -and $owned.JobMembership.Succeeded -and $owned.Owned -and $owned.QueryError -eq 0 -and $owned.State -eq 'running') 'real assigned child keeps successful ownership and core behavior'
    $rows=$inspect.Invoke($null,[object[]]@($other,[UInt64[]]@($process.Id),$null,$null));$foreign=$rows[0]
    Check-MemberQuery ($foreign.JobMembership.Succeeded -and $foreign.JobMembership.Error -eq 0 -and -not $foreign.Owned -and $foreign.State -eq 'ownership-lost') 'API success cannot masquerade as membership in another job'
    Check-MemberQuery (-not [IsolatedFuzzProcess]::IsFinalShutdownStateVerified($foreign)) 'nonmember remains rejected despite successful diagnostic API records'
} finally {
    if($handle -ne [IntPtr]::Zero -and -not $assigned){[MemberQueryTestJob]::TerminateProcess($handle,0xE003)|Out-Null}
    if($job -ne [IntPtr]::Zero){[MemberQueryTestJob]::TerminateJobObject($job,0xE003)|Out-Null;[MemberQueryTestJob]::CloseHandle($job)|Out-Null}
    if($other -ne [IntPtr]::Zero){[MemberQueryTestJob]::CloseHandle($other)|Out-Null}
    $process.Dispose()
}
