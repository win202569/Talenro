param([switch]$QueryOverrunOnly)
$ErrorActionPreference='Stop'
$source=Join-Path $PSScriptRoot 'retained-lifetime.cs'
if(-not(Test-Path -LiteralPath $source)){throw 'ASSERTION FAILED: retained lifetime registry contract missing'}
if(-not('RetainedLifetimeRegistry' -as [type])){Add-Type -Path @($source,(Join-Path $PSScriptRoot 'retained-lifetime-fixture.cs'))}
function Check-Lifetime([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
function New-LifetimeFixture {
    $clock=New-Object FakeLifetimeClock;$api=New-Object FakeLifetimeApi;$api.Clock=$clock
    $registry=New-Object RetainedLifetimeRegistry([IntPtr]1,[uint32]1,$api,$clock)
    [pscustomobject]@{Clock=$clock;Api=$api;Registry=$registry}
}
$f=New-LifetimeFixture;$f.Registry.Poll(1000)
Check-Lifetime ($f.Registry.Report.Admitted -eq 1 -and $f.Registry.Report.Entries[0].EntryId -eq 1 -and $f.Registry.Report.Entries[0].Membership.Owned) 'AdmissionOwnsExactHandle'
$f.Registry.Poll(1000);$f.Registry.Snapshot('root-wait-end',1000,$true,$false,$false);$f.Api.Exited=$true;$f.Registry.Snapshot('post-cleanup',1000,$true,$false,$true)
Check-Lifetime ($f.Api.OpenCalls -eq 1 -and $f.Api.ReadCalls -eq 3 -and $f.Registry.Report.Entries[0].Final.State -ceq 'exited' -and $f.Registry.Report.Entries[0].Final.TerminationRequested) 'SameHandleSnapshots ForcedExitNeverNatural'
$f.Registry.Dispose();$f.Registry.Dispose()
Check-Lifetime ($f.Api.ReleaseCalls -eq 1 -and $f.Api.DoubleCloseCalls -eq 0) 'DisposeAllOnce'
$before=$f.Api.ReadCalls;$f.Registry.Snapshot('post-cleanup',1000,$true,$false,$false)
Check-Lifetime ($f.Api.ReadCalls -eq $before) 'SnapshotAfterDispose'
$f=New-LifetimeFixture;$f.Api.OpenError=87;$f.Registry.Poll(1000);$f.Registry.Poll(1000)
Check-Lifetime ($f.Registry.Report.Attempts -eq 1 -and $f.Api.MembershipCalls -eq 0 -and $f.Registry.Report.Entries[0].Open.Error -eq 87 -and $f.Api.OpenCalls -eq 1 -and $f.Registry.Report.Incomplete) 'Open87StopsBeforeMembership DiagnosticErrorStopsAdmission'
$f.Registry.Dispose()
if(-not $QueryOverrunOnly){
$f=New-LifetimeFixture;$f.Api.Exited=$true;$f.Registry.Poll(1000);$f.Api.Exited=$false;$f.Registry.Snapshot('root-wait-end',1000,$true,$false,$false);$f.Registry.Snapshot('post-cleanup',1000,$true,$false,$false)
$entry=$f.Registry.Report.Entries[0]
Check-Lifetime ($entry.Admission.State -ceq 'exited' -and $entry.RootWaitEnd.Ambiguous -and $entry.Final.Ambiguous -and $f.Registry.Report.Incomplete -and $f.Api.OpenCalls -eq 1) 'ExitedThenRunningExplicitAmbiguity'
$f.Registry.Dispose();Check-Lifetime ($f.Api.ReleaseCalls -eq 1) 'AmbiguousEntryReleasedOnce'
}
foreach($case in @('empty','root','seen')){
    $f=New-LifetimeFixture
    if($case -eq 'seen'){$f.Registry.Poll(1000)}
    $reads=$f.Api.ReadCalls;$opens=$f.Api.OpenCalls;$queries=$f.Api.QueryCalls
    switch($case){'empty'{$f.Api.Ids=[uint64[]]@()};'root'{$f.Api.Ids=[uint64[]]@(1)};'seen'{$f.Api.Ids=[uint64[]]@(2)}}
    $f.Api.AdvanceOnQuery=1000;$f.Registry.Poll(1000);$f.Registry.Poll(1000)
    Check-Lifetime ($f.Registry.Report.Incomplete -and $f.Registry.Report.Errors -contains 'list-overrun:-1' -and $f.Api.QueryCalls -eq $queries+1 -and $f.Api.OpenCalls -eq $opens -and $f.Api.ReadCalls -eq $reads) ('QueryOverrunExplicit '+$case)
    $f.Registry.Dispose()
}
$f=New-LifetimeFixture;$f.Registry.Poll(1000);$f.Clock.Now=99;$f.Registry.Poll(1000)
Check-Lifetime ($f.Registry.Report.Incomplete -and $f.Registry.Report.Errors -contains 'clock:-1' -and $f.Api.QueryCalls -eq 1) 'StaleClockStopsQueries'
$f.Registry.Dispose()
$f=New-LifetimeFixture;$f.Registry.Poll([long]::MaxValue);$f.Registry.Snapshot('root-wait-end',100,$false,$true,$false)
Check-Lifetime ($f.Registry.Report.Entries[0].RootWaitEnd.SkipReason -ceq 'budget-exhausted' -and $f.Api.ReadCalls -eq 1) 'ExtremeDeadlineNoOverflow expired snapshot explicit'
$f.Registry.Dispose()
foreach($mode in @('error','foreign','throw')){
    $f=New-LifetimeFixture;if($mode -eq 'error'){$f.Api.MembershipError=6};if($mode -eq 'foreign'){$f.Api.Owned=$false};if($mode -eq 'throw'){$f.Api.ThrowMembership=$true}
    $f.Registry.Poll(1000);$f.Registry.Dispose()
    Check-Lifetime ($f.Registry.Report.Admitted -eq 0 -and $f.Api.ReleaseCalls -eq 1 -and $f.Api.DoubleCloseCalls -eq 0 -and $f.Registry.Report.Incomplete) ('MembershipFailureReleasesTemporary '+$mode)
}
$f=New-LifetimeFixture;$f.Api.Ids=[uint64[]]@(0,1,4294967296,2,2);$f.Registry.Poll(1000);$f.Registry.Poll(1000)
Check-Lifetime ($f.Api.OpenCalls -eq 1 -and $f.Registry.Report.Entries.Count -eq 1 -and $f.Api.ReopenedHandles -eq 0 -and $f.Registry.Report.Incomplete) 'DuplicateAndRootIgnored NumericAmbiguityNeverMerged'
$f.Registry.Dispose()
$f=New-LifetimeFixture;$f.Api.Ids=[uint64[]](2..34);$f.Api.FailPid=33;$f.Api.OpenError=87;$f.Registry.Poll(1000);$f.Registry.Poll(1000)
Check-Lifetime ($f.Registry.Report.Attempts -eq 32 -and $f.Registry.Report.Admitted -eq 31 -and $f.Api.OpenCalls -eq 32 -and $f.Registry.Report.AdmissionBudgetExhausted -and $f.Api.QueryCalls -eq 1) 'Cap32No33rdOpen includes failure'
$f.Registry.Dispose();Check-Lifetime ($f.Api.ReleaseCalls -eq 31) 'Cap32 releases admitted only'
$f=New-LifetimeFixture;$f.Registry.Poll(100);$f.Registry.Snapshot('root-wait-end',100,$false,$true,$false)
Check-Lifetime ($f.Api.QueryCalls -eq 0 -and $f.Api.OpenCalls -eq 0 -and $f.Registry.Report.TimedOut) 'ExpiredBudgetNoCalls'
$f.Registry.Dispose()
$f=New-LifetimeFixture;$f.Api.AdvanceOnOpen=1000;$f.Registry.Poll(1000);$f.Registry.Dispose()
Check-Lifetime ($f.Api.MembershipCalls -eq 0 -and $f.Api.ReleaseCalls -eq 1 -and $f.Registry.Report.Incomplete) 'CallOverrunNoClockReset'
$f=New-LifetimeFixture;$f.Api.MissingExit=$true;$f.Api.MissingImage=$true;$f.Registry.Poll(1000)
Check-Lifetime (-not $f.Registry.Report.Entries[0].Admission.ExitCodeRead -and $f.Registry.Report.Entries[0].Admission.ImageError -eq 5 -and $f.Registry.Report.Incomplete) 'MissingImageOrExitUnknown'
$f.Registry.Dispose()
foreach($mode in @('return','throw')){
    $f=New-LifetimeFixture;$f.Api.Ids=[uint64[]]@(2,3);$f.Registry.Poll(1000)
    if($mode -eq 'throw'){$f.Api.ThrowClose=$true}else{$f.Api.CloseFails=$true}
    $f.Registry.Dispose();$f.Registry.Dispose()
    Check-Lifetime ($f.Api.ReleaseCalls -eq 2 -and $f.Api.DoubleCloseCalls -eq 0 -and $f.Registry.Report.ReleaseFailures -eq 2) ('CloseFailureContinues '+$mode)
}
$f=New-LifetimeFixture;$f.Registry.Poll(1000);$f.Api.ThrowRead=$true;$f.Registry.Snapshot('root-wait-end',1000,$true,$false,$false);$f.Registry.Dispose()
Check-Lifetime ($f.Registry.Report.Incomplete -and $f.Registry.Report.Entries[0].RootWaitEnd.State -ceq 'unknown' -and $f.Api.ReleaseCalls -eq 1) 'ReadExceptionStillDisposes'
$f=New-LifetimeFixture;$f.Api.Partial=$true;$f.Registry.Poll(1000);$f.Registry.Dispose()
Check-Lifetime ($f.Api.OpenCalls -eq 0 -and $f.Registry.Report.Incomplete) 'PartialListNeverAdmits'
$f=New-LifetimeFixture;$f.Registry.Poll(1000);$f.Registry.Snapshot('root-wait-end',1000,$true,$false,$false);$saved=$f.Registry.Report.Entries[0].RootWaitEnd;$f.Registry.Snapshot('root-wait-end',1000,$false,$true,$true)
Check-Lifetime ([object]::ReferenceEquals($saved,$f.Registry.Report.Entries[0].RootWaitEnd) -and $f.Api.ReadCalls -eq 2) 'RepeatedPhaseNeverOverwrites'
$json=$f.Registry.Report|ConvertTo-Json -Depth 10
Check-Lifetime ($json -notmatch '900002|"Handle"' -and $json -match 'sampled') 'NoHandleSerialization'
$f.Registry.Dispose()
