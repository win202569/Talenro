$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'run.ps1')
function Check-InitialQuery([bool]$Value,[string]$Name){if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
$flags=[Reflection.BindingFlags]'NonPublic,Static'
$read=[IsolatedFuzzProcess].GetMethod('ReadInitialJobQueries',$flags)
Check-InitialQuery ($null -ne $read) 'initial accounting and member queries expose diagnostic evidence'
Add-Type @'
using System;
using System.Runtime.InteropServices;
public static class InitialQueryFixtureJob {
 [DllImport("kernel32.dll",SetLastError=true)] public static extern IntPtr CreateJobObject(IntPtr a,string n);
 [DllImport("kernel32.dll")] public static extern bool CloseHandle(IntPtr h);
}
'@
function Check-QueryClock($Query){
    Check-InitialQuery ($Query.StartUTC -and $Query.EndUTC -and $Query.Frequency -gt 0 -and $Query.EndTimestamp -ge $Query.StartTimestamp) ('ordered clocks for '+$Query.API)
}
function Get-BaseError($ErrorRecord){$e=$ErrorRecord.Exception;while($e.InnerException){$e=$e.InnerException};return $e}
$job=[InitialQueryFixtureJob]::CreateJobObject([IntPtr]::Zero,$null)
if($job -eq [IntPtr]::Zero){throw 'Empty fixture job creation failed'}
try {
    $result=New-Object 'IsolatedFuzzProcess+Result'
    $accounting=$read.Invoke($null,[object[]]@($job,$result.PSObject.BaseObject))
    $q=$result.InitialAccountingQuery
    Check-InitialQuery ($q.API -ceq 'QueryInformationJobObject/Accounting' -and $q.Succeeded -and $q.Error -eq 0 -and $q.ValuesRead) 'real empty accounting query retains success rather than stale error'
    Check-InitialQuery ($q.TotalProcesses -eq 0 -and $q.ActiveProcesses -eq 0 -and $q.TotalTerminatedProcesses -eq 0 -and $accounting.Total -eq 0 -and $accounting.Active -eq 0) 'empty raw counts remain distinct from completion flags'
    Check-QueryClock $q
    Check-InitialQuery ($result.InitialMemberQueries.Count -eq 1 -and $result.MembersAtRootWait.Count -eq 0) 'empty member enumeration records one actual attempt and original empty output'
    $m=$result.InitialMemberQueries[0]
    Check-InitialQuery ($m.API -ceq 'QueryInformationJobObject/ProcessIdList' -and $m.Succeeded -and $m.Error -eq 0 -and $m.HeaderRead -and $m.Assigned -eq 0 -and $m.Returned -eq 0 -and $m.Attempt -eq 1 -and $m.Capacity -eq 16 -and $m.BufferBytes -eq (8+16*[IntPtr]::Size)) 'member query retains native header and buffer facts'
    Check-QueryClock $m
    Check-InitialQuery ($q.EndTimestamp -le $m.StartTimestamp) 'accounting and enumeration retain actual non-atomic order'
    Check-InitialQuery (-not $result.NaturalShutdownVerified -and -not $result.CleanupVerified -and $result.ShutdownSchema -eq 3) 'diagnostics do not manufacture shutdown or cleanup acceptance'
    $saved=ConvertTo-RecordedProcessReceipt $result 'stdout' 'stderr' 'fixture.exe' @('fixture') @{}
    $roundtrip=$saved|ConvertTo-Json -Depth 6|ConvertFrom-Json
    Check-InitialQuery ($roundtrip.InitialAccountingQuery.ValuesRead -and $roundtrip.InitialMemberQueries.Count -eq 1 -and $roundtrip.InitialMemberQueries[0].HeaderRead -and $roundtrip.InitialMemberQueries[0].Returned -eq 0) 'real receipt conversion preserves nested diagnostic values'
    $refused=$false;try{Assert-ProcessCompletion $roundtrip}catch{$refused=$true}
    Check-InitialQuery $refused 'valid diagnostics cannot upgrade incomplete completion receipt'
    $clean=New-Object 'IsolatedFuzzProcess+Result'
    foreach($entry in @{RootId=17;RootExited=$true;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;CleanupVerified=$true;MembersAtRootWait=@();MembersBeforeTermination=@();MembersAfterCleanup=@();ShutdownMemberEvidence=@()}.GetEnumerator()){$clean.($entry.Key)=$entry.Value}
    $clean.InitialAccountingQuery=$result.InitialAccountingQuery
    $clean.InitialMemberQueries=$result.InitialMemberQueries
    $receipt=ConvertTo-RecordedProcessReceipt $clean 'stdout' 'stderr' 'fixture.exe' @('fixture') @{}
    $receipt=$receipt|ConvertTo-Json -Depth 6|ConvertFrom-Json
    Assert-ProcessCompletion $receipt
    Check-InitialQuery ($receipt.Executable -ceq 'fixture.exe' -and $receipt.Arguments[0] -ceq 'fixture' -and $receipt.StdoutPath -ceq 'stdout') 'receipt conversion preserves execution context alongside new diagnostics'
    $receipt.PSObject.Properties.Remove('InitialAccountingQuery');$receipt.PSObject.Properties.Remove('InitialMemberQueries')
    Assert-ProcessCompletion $receipt
    Check-InitialQuery $true 'legacy receipt without optional diagnostics keeps the same passing gate'
    $receipt.NaturalShutdownVerified=$false;$receipt.TerminationRequested=$true
    $receipt|Add-Member InitialAccountingQuery $result.InitialAccountingQuery
    $receipt|Add-Member InitialMemberQueries $result.InitialMemberQueries
    $refused=$false;try{Assert-ProcessCompletion $receipt}catch{$refused=$true}
    Check-InitialQuery $refused 'metadata cannot forgive forced cleanup on an otherwise passing receipt'
    $failed=New-Object 'IsolatedFuzzProcess+Result';$failure=$null
    try{$null=$read.Invoke($null,[object[]]@([IntPtr](-1),$failed.PSObject.BaseObject))}catch{$failure=Get-BaseError $_}
    Check-InitialQuery ($null -ne $failure -and $failure.NativeErrorCode -eq 6 -and -not $failed.InitialAccountingQuery.Succeeded -and $failed.InitialAccountingQuery.Error -eq 6 -and -not $failed.InitialAccountingQuery.ValuesRead) 'invalid job keeps original accounting failure and no fabricated counts'
    Check-QueryClock $failed.InitialAccountingQuery
    Check-InitialQuery ($failed.InitialMemberQueries.Count -eq 0) 'failed accounting does not add an enumeration call'
    $members=[IsolatedFuzzProcess].GetMethod('MembersWithEvidence',$flags)
    $rows=New-Object 'System.Collections.Generic.List[IsolatedFuzzProcess+MemberListQueryObservation]'
    $failure=$null;try{$null=$members.Invoke($null,[object[]]@([IntPtr](-1),$rows.PSObject.BaseObject))}catch{$failure=Get-BaseError $_}
    Check-InitialQuery ($null -ne $failure -and $failure.NativeErrorCode -eq 6 -and $rows.Count -eq 1 -and -not $rows[0].Succeeded -and $rows[0].Error -eq 6 -and -not $rows[0].HeaderRead) 'failed member query retains raw API error without reading invalid header or retrying'
    Check-QueryClock $rows[0]
} finally {
    if(-not [InitialQueryFixtureJob]::CloseHandle($job)){throw 'Empty fixture job handle close failed'}
}
