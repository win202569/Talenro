$ErrorActionPreference='Stop'
Add-Type -Path @((Join-Path $PSScriptRoot 'process-native.cs'),(Join-Path $PSScriptRoot 'retained-lifetime.cs'),(Join-Path $PSScriptRoot 'compile-stage-trigger.cs'),(Join-Path $PSScriptRoot 'compile-stage-session.cs'),(Join-Path $PSScriptRoot 'retained-lifetime-fixture.cs'))
function Check-Integration([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
Check-Integration ($null -ne [IsolatedFuzzProcess].GetMethod('RunWithRetainedHandleDiagnostics')) 'NewEntryExists'
Check-Integration (@([IsolatedFuzzProcess].GetMethods()|Where-Object Name -eq 'Run').Count -eq 2) 'LegacyOverloadsRemain'
$close=[IsolatedFuzzProcess].GetMethod('CloseHandle',[Reflection.BindingFlags]'NonPublic,Static')
$import=@($close.GetCustomAttributes([Runtime.InteropServices.DllImportAttribute],$false))[0]
Check-Integration $import.SetLastError 'CloseFailureErrorNotStale'
$r=New-Object 'IsolatedFuzzProcess+Result'
Check-Integration (-not $r.LifetimeDiagnosticOnly -and $null -eq $r.LifetimeDiagnostics -and $r.ShutdownSchema -eq 3) 'LegacyDefaultNoRegistry'
$clock=New-Object FakeLifetimeClock;$api=New-Object FakeLifetimeApi;$api.Clock=$clock;$api.Ids=[uint64[]]@(2,3)
$registry=New-Object RetainedLifetimeRegistry([IntPtr]1,[uint32]1,$api,$clock);$registry.Poll(1000);$api.ThrowClose=$true
$coreError=New-Object InvalidOperationException('expected core failure')
[LifetimeCoreFixture]::Expected=$coreError
$action=[Action][Delegate]::CreateDelegate([Action],[LifetimeCoreFixture].GetMethod('ThrowCore'))
try{[RetainedLifetimeTeardown]::Run($action,$registry);throw 'core error swallowed'}catch{$caught=$_.Exception;while($caught.InnerException){$caught=$caught.InnerException};Check-Integration ([object]::ReferenceEquals($caught,$coreError)) 'CoreFailurePreservesIdentity'}
Check-Integration ($api.ReleaseCalls -eq 2 -and $registry.Report.ReleaseFailures -eq 2 -and $registry.Report.TeardownAttempted) 'CoreFailureStillDisposesAll'
$clock=New-Object FakeLifetimeClock;$api=New-Object FakeLifetimeApi;$registry=New-Object RetainedLifetimeRegistry([IntPtr]1,[uint32]1,$api,$clock);$registry.Poll(1000)
[LifetimeCoreFixture]::Api=$api
$cleanup=[Action][Delegate]::CreateDelegate([Action],[LifetimeCoreFixture].GetMethod('MarkCleanup'))
[RetainedLifetimeTeardown]::Run($cleanup,$registry)
Check-Integration ($api.Calls.IndexOf('cleanup') -lt $api.Calls.IndexOf('close:900002')) 'CleanupPrecedesRelease'
$deadline=[IsolatedFuzzProcess].GetMethod('LifetimeDeadline',[Reflection.BindingFlags]'NonPublic,Static')
try{$deadline.Invoke($null,@([long]::MaxValue,[long]1000,[long]1000));throw 'overflow not rejected'}catch{Check-Integration ($_.Exception.ToString() -match 'OverflowException') 'DeadlineOverflowRejected'}
# Supplementary static wiring inspection, NOT native behavioral validation.
$source=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'process-native.cs') -Raw
Check-Integration ($source -match 'RunCore\(executable,args,overrides,cwd,stdout,stderr,limitSeconds,captureLiveMembers,captureShutdownDiagnostic,false\)' -and $source -match 'RunCore\(executable,args,overrides,cwd,stdout,stderr,limitSeconds,false,false,true\)') 'STATIC LegacyModeSelection'
Check-Integration ($source.IndexOf('registry.Snapshot("root-wait-end"') -lt $source.IndexOf('accounting=ReadInitialJobQueries(job,result)') -and $source.IndexOf('registry.Snapshot("post-cleanup"') -gt $source.IndexOf('result.MembersAfterCleanup=Members(job)')) 'STATIC HookOrder'
Check-Integration ($source -match 'RetainedLifetimeTeardown.Run' -and $source -match 'LifetimeDeadline\(cleanupStart,5000,Stopwatch.Frequency\)') 'STATIC SharedTeardownAndBudget'
$loader=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'run.ps1') -Raw
$inventory=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'preflight.ps1') -Raw
Check-Integration ($loader -match 'retained-lifetime.cs' -and $inventory -match 'retained-lifetime.cs') 'STATIC LoaderAndInventory'
Write-Output 'LIMIT: no native adapter or Run entry invoked; static/fake evidence only.'
