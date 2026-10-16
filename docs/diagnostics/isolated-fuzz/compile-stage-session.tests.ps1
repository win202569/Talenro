$ErrorActionPreference='Stop'
function Check-Session([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
$file=Join-Path $PSScriptRoot 'compile-stage-session.cs'
Check-Session (Test-Path $file) 'CompileStageRetainedSession exists'
Add-Type -Path @((Join-Path $PSScriptRoot 'retained-lifetime.cs'),(Join-Path $PSScriptRoot 'compile-stage-trigger.cs'),$file,(Join-Path $PSScriptRoot 'retained-lifetime-fixture.cs'),(Join-Path $PSScriptRoot 'compile-stage-fixture.cs'))
$line='"C:/SDK/pkg/tool/windows_amd64/compile.exe" -p talenro.local/platform/internal/nodecontrol/nodeleafbinding -pack "C:/Repo/internal/nodecontrol/nodeleafbinding/a_test.go"'+"`n"
$f=New-Object FakeCompileStageHarness;$f.Bytes.Append([Text.Encoding]::UTF8.GetBytes('partial'));$f.Session.Tick(100000)
Check-Session ($f.Factories -eq 0 -and $f.Api.OpenCalls -eq 0 -and $f.Api.QueryCalls -eq 0 -and $null -eq $f.Session.LifetimeReport) 'WaitingCreatesNothing'
$f.Session.Dispose();Check-Session ($f.Bytes.DisposeCalls -eq 1) 'waiting reader closed'
$f=New-Object FakeCompileStageHarness;$f.Bytes.Append([Text.Encoding]::UTF8.GetBytes($line));$f.Root.WaitResult=0;$f.Session.Tick(100000)
Check-Session ($f.Factories -eq 0 -and $f.Api.OpenCalls -eq 0 -and $f.Session.TriggerReport.State -ceq 'RootExitedWithoutActivation') 'RootExitBetweenReadAndCheck'
$f.Session.Dispose()
$f=New-Object FakeCompileStageHarness;$f.Bytes.Append([Text.Encoding]::UTF8.GetBytes($line));$f.Session.Tick(100000);$reads=$f.Bytes.ReadCalls;$f.Session.Tick(100000)
Check-Session ($f.Factories -eq 1 -and $f.Api.OpenCalls -eq 1 -and $f.Api.QueryCalls -eq 2 -and $f.Bytes.ReadCalls -eq $reads) 'ActivateOnce'
Check-Session ($f.Session.TriggerReport.Activation.EndTimestamp -le $f.Session.TriggerReport.FirstRegistryQuery.StartTimestamp) 'activation before first query'
$f.Api.Exited=$true;$f.Session.RootExited();$f.Session.Snapshot('root-wait-end',100000,$true,$false,$false);$f.Session.Snapshot('post-cleanup',100000,$true,$false,$false)
Check-Session ($f.Session.LifetimeReport.Entries[0].Final.State -ceq 'exited' -and $f.Session.TriggerReport.State -ceq 'Activated') 'root exit after activation preserves snapshots'
$f.Bytes.ThrowDispose=$true;[LifetimeCoreFixture]::Expected=New-Object InvalidOperationException('core stage failure');$caught=$null
try{[CompileStageRetainedTeardown]::Run([Action][Delegate]::CreateDelegate([Action],[LifetimeCoreFixture].GetMethod('ThrowCore')),$f.Session)}catch{$caught=$_.Exception;while($caught.InnerException){$caught=$caught.InnerException}}
Check-Session ([object]::ReferenceEquals($caught,[LifetimeCoreFixture]::Expected) -and $f.Api.ReleaseCalls -eq 1 -and $f.Bytes.DisposeCalls -eq 1 -and $f.Session.TriggerReport.Errors.Count -eq 1) 'TeardownPreservesCoreFailure'
$f.Session.Dispose();Check-Session ($f.Api.ReleaseCalls -eq 1) 'session teardown once'
$f=New-Object FakeCompileStageHarness;$f.Bytes.Append([Text.Encoding]::UTF8.GetBytes($line));$f.Api.Ids=[uint64[]](2..33);$f.Session.Tick(100000);$f.Session.Tick(100000)
Check-Session ($f.Factories -eq 1 -and $f.Api.OpenCalls -eq 32 -and $f.Api.QueryCalls -eq 1 -and $f.Session.LifetimeReport.AdmissionBudgetExhausted) 'no extra cohort after cap'
$f.Api.Exited=$true;$f.Session.Snapshot('post-cleanup',100000,$true,$false,$false);$f.Session.Dispose();Check-Session ($f.Api.ReleaseCalls -eq 32) 'cap still releases all'
$f=New-Object FakeCompileStageHarness;$f.Bytes.Append([Text.Encoding]::UTF8.GetBytes($line));$f.Root.Error=87;$f.Session.Tick(100000)
Check-Session ($f.Factories -eq 0 -and $f.Session.TriggerReport.State -ceq 'Failed' -and $f.Session.TriggerReport.RootCheck.Error -eq 87) 'root check failure retains original error'
$f.Session.Dispose()
$f=New-Object FakeCompileStageHarness;$f.Bytes.Append([Text.Encoding]::UTF8.GetBytes($line));$f.Bytes.ThrowRead=$true;$f.Session.Tick(100000);$f.Session.RootExited();$f.Session.Dispose()
Check-Session ($f.Factories -eq 0 -and $f.Bytes.DisposeCalls -eq 1 -and $f.Session.TriggerReport.State -ceq 'Failed') 'reader failure still disposes no fallback'
$f=New-Object FakeCompileStageHarness;$f.Bytes.Append([Text.Encoding]::UTF8.GetBytes($line));$f.Session.RootExited();$f.Session.Tick(100000);$f.Session.Dispose();Check-Session ($f.Factories -eq 0) 'session no late activation'
$f=New-Object FakeCompileStageHarness;$f.Api.Ids=[uint64[]]@();$f.Bytes.Append([Text.Encoding]::UTF8.GetBytes($line));$f.Session.Tick(100000);$f.Session.Dispose();Check-Session ($f.Session.LifetimeReport.Admitted -eq 0 -and $f.Factories -eq 1) 'zero samples not fabricated'
