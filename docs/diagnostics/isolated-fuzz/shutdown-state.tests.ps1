$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'run.ps1')
function Check-ShutdownState([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function New-StateRow { $row=New-Object 'IsolatedFuzzProcess+MemberObservation';$row.Id=42;$row.Owned=$true;$row.QueryError=0;$row.ExitCodeRead=$true;$row.ExitCode=0;$row.State='exited';$row.ImageError=31;$row.ImagePath=$null;return $row }
$row=New-StateRow
Check-ShutdownState ([IsolatedFuzzProcess]::IsShutdownStateVerified($row)) 'owned verified exited member does not need readable image metadata'
foreach($case in @(
    @('Owned',$false,'unknown ownership still rejected'),
    @('QueryError',5,'failed state query still rejected'),
    @('ExitCodeRead',$false,'unread exit status still rejected'),
    @('State','unknown','unknown wait state still rejected'),
    @('State','running','running member with missing image still rejected')
)) {$row=New-StateRow;$row.($case[0])=$case[1];Check-ShutdownState (-not [IsolatedFuzzProcess]::IsShutdownStateVerified($row)) $case[2]}
$row=New-StateRow;$row.State='running';$row.ExitCode=0;$row.ImageError=0;$row.ImagePath=$null
Check-ShutdownState (-not [IsolatedFuzzProcess]::IsShutdownStateVerified($row)) 'zero exit-code read cannot substitute for signaled state or image'
$row.ImagePath='C:\Windows\System32\powershell.exe'
Check-ShutdownState ([IsolatedFuzzProcess]::IsShutdownStateVerified($row)) 'known running member state remains observable not completed'
Check-ShutdownState (-not [IsolatedFuzzProcess]::IsShutdownStateVerified($null)) 'absent member observation cannot verify state'
