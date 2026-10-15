$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'run.ps1')
function Check-LaunchTiming([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
Check-LaunchTiming ($null -ne [IsolatedFuzzProcess].GetMethod('RecordLaunchBoundary')) 'native runner exposes the real launch-boundary recorder'
$result=New-Object IsolatedFuzzProcess+Result
$watch=[Diagnostics.Stopwatch]::StartNew()
[Threading.Thread]::SpinWait(20000)
$elapsedBefore=$watch.ElapsedTicks;$stampBefore=[Diagnostics.Stopwatch]::GetTimestamp();$utcBefore=[DateTimeOffset]::UtcNow
[IsolatedFuzzProcess]::RecordLaunchBoundary($result,'BeforeCreateProcess',$watch)
[Threading.Thread]::SpinWait(20000)
[IsolatedFuzzProcess]::RecordLaunchBoundary($result,'AfterCreateProcess',$watch)
$utcAfter=[DateTimeOffset]::UtcNow;$stampAfter=[Diagnostics.Stopwatch]::GetTimestamp()
$rows=@($result.LaunchBoundaries)
Check-LaunchTiming ($rows.Count -eq 2 -and $rows[0].Name -ceq 'BeforeCreateProcess' -and $rows[1].Name -ceq 'AfterCreateProcess') 'recorder appends real observations in call order'
Check-LaunchTiming ($rows[0].ElapsedTicks -ge $elapsedBefore -and $rows[1].ElapsedTicks -ge $rows[0].ElapsedTicks) 'elapsed boundaries use existing budget clock without resetting it'
Check-LaunchTiming ($rows[0].Timestamp -ge $stampBefore -and $rows[1].Timestamp -ge $rows[0].Timestamp -and $rows[1].Timestamp -le $stampAfter) 'monotonic timestamps bracket the real calls'
Check-LaunchTiming ($rows[0].Frequency -eq [Diagnostics.Stopwatch]::Frequency -and $rows[1].Frequency -eq $rows[0].Frequency -and $rows[0].Frequency -gt 0) 'frequency makes recorded ticks interpretable'
Check-LaunchTiming ([DateTimeOffset]::Parse($rows[0].UTC) -ge $utcBefore -and [DateTimeOffset]::Parse($rows[1].UTC) -le $utcAfter -and $watch.IsRunning) 'UTC boundaries are captured while original clock continues'
$roundtrip=$result|ConvertTo-Json -Depth 5|ConvertFrom-Json
Check-LaunchTiming ($roundtrip.LaunchBoundaries.Count -eq 2 -and $roundtrip.LaunchBoundaries[1].Name -ceq 'AfterCreateProcess' -and $roundtrip.LaunchBoundaries[1].Timestamp -eq $rows[1].Timestamp -and $roundtrip.LaunchBoundaries[1].ElapsedTicks -eq $rows[1].ElapsedTicks) 'native diagnostic boundaries survive receipt-style JSON serialization'
Check-LaunchTiming ($result.RootId -eq 0 -and -not $result.CleanupVerified -and -not $result.NaturalShutdownVerified) 'timing records alone cannot invent process or shutdown proof'
