$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot 'run.ps1')
function Assert-True([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Assert-Throws([scriptblock]$Action,[string]$Name) {$caught=$false;try{& $Action|Out-Null}catch{$caught=$true};Assert-True $caught $Name}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$root=Join-Path $workspace ('.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/runner-tests-'+[guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $root | Out-Null
$shell=Join-Path ([Environment]::GetFolderPath('System')) 'WindowsPowerShell/v1.0/powershell.exe'
function Encoded([string]$Code) {return [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($Code))}
$nonzero=Invoke-RecordedProcess $shell @('-NoProfile','-EncodedCommand',(Encoded '[Console]::Out.WriteLine("out-marker");[Console]::Error.WriteLine("err-marker");exit 7')) @{} (Join-Path $root 'nonzero') 10
Assert-True ($nonzero.ExitCode -eq 7) 'nonzero exit preserved'
Assert-True ((Get-Content -Raw -LiteralPath $nonzero.StdoutPath).Contains('out-marker')) 'stdout preserved'
Assert-True ((Get-Content -Raw -LiteralPath $nonzero.StderrPath).Contains('err-marker')) 'stderr preserved'
Assert-True (-not $nonzero.TimedOut -and $nonzero.DescendantsRemaining -eq 0) 'finished child is not timeout'
Assert-Throws {Invoke-RecordedProcess $shell @('-NoProfile') @{} (Join-Path $root 'nonzero') 10} 'duplicate receipt refused'
$timeout=Invoke-RecordedProcess $shell @('-NoProfile','-EncodedCommand',(Encoded 'Start-Sleep -Seconds 20')) @{} (Join-Path $root 'timeout') 1
Assert-True ($timeout.TimedOut -and $timeout.CleanupVerified -and $timeout.ActiveAfterCleanup -eq 0) 'timeout bounded to owned job and cleaned'
Assert-Throws {Invoke-RecordedProcess $shell @('-NoProfile') @{GOROOT='C:/invalid-source'} (Join-Path $root 'bad-root') 10} 'invalid source provenance rejected'
$receipt=[pscustomobject]@{ExitCode=1;TimedOut=$false;DescendantsRemaining=0;CleanupVerified=$true;ActiveAfterCleanup=0}
Assert-Throws {Assert-ArmResult 'parent-first' $receipt @()} 'nonzero without trace rejected'
$badTrace=@([pscustomobject]@{event='return';watchdog=$true;activeWorkers=0;seq=1;schema=1;mode='parent-first';source='return';newError='deadline'})
Assert-Throws {Assert-ArmResult 'parent-first' $receipt $badTrace} 'watchdog trace rejected'
$receipt.DescendantsRemaining=1
Assert-Throws {Assert-ArmResult 'crash' $receipt @()} 'unfinished descendants rejected'
$armRoot=Join-Path $root 'duplicate-arm'
New-Item -ItemType Directory -Path (Join-Path $armRoot 'arms/observe') -Force | Out-Null
Assert-Throws {Invoke-DiagnosticArm $armRoot 'observe'} 'duplicate arm refused before execution'
$receipt.DescendantsRemaining=0
$valid=@(
 [pscustomobject]@{schema=1;seq=1;mode='parent-first';event='stop';source='parent-done';incoming='deadline';childErr='nil';parentRead=$false;parentErr='unread';oldError='nil';newError='deadline';stopping=$false;stoppingRead=$true;gateEntered=$true;propagated=$false;watchdog=$false;activeWorkers=1},
 [pscustomobject]@{schema=1;seq=2;mode='parent-first';event='stop';source='worker';incoming='nil';childErr='deadline';parentRead=$false;parentErr='unread';oldError='deadline';newError='deadline';stopping=$true;stoppingRead=$true;gateEntered=$true;propagated=$true;watchdog=$false;activeWorkers=1},
 [pscustomobject]@{schema=1;seq=3;mode='parent-first';event='return';source='return';incoming='deadline';childErr='unread';parentRead=$false;parentErr='unread';oldError='unread';newError='deadline';stopping=$false;stoppingRead=$false;gateEntered=$true;propagated=$true;watchdog=$false;activeWorkers=0}
)
Assert-ArmResult 'parent-first' $receipt $valid
Assert-True $true 'complete forced trace accepted'
foreach($row in $valid){$row.mode='observe'}
Assert-Throws {Assert-ArmResult 'parent-first' $receipt $valid} 'wrong mode cannot masquerade as forced evidence'
foreach($row in $valid){$row.mode='parent-first'}
$missingWorker=@($valid[0],$valid[2]);$missingWorker[1].seq=2
Assert-Throws {Assert-ArmResult 'parent-first' $receipt $missingWorker} 'worker shutdown evidence required'
$childCode=Encoded 'Start-Sleep -Seconds 20'
$parentCode='Start-Process -FilePath "'+$shell+'" -ArgumentList "-NoProfile -EncodedCommand '+$childCode+'" -WindowStyle Hidden; Start-Sleep -Seconds 20'
$tree=Invoke-RecordedProcess $shell @('-NoProfile','-EncodedCommand',(Encoded $parentCode)) @{} (Join-Path $root 'tree-timeout') 2
Assert-True ($tree.TotalProcesses -ge 2 -and $tree.TimedOut -and $tree.CleanupVerified -and $tree.ActiveAfterCleanup -eq 0) 'real owned descendant cleaned on timeout'
$envName='TALENRO_DIAG_ENV_TEST'
$beforeEnv=[Environment]::GetEnvironmentVariable($envName)
$wrappedEnv=[System.Management.Automation.PSObject]::AsPSObject('child-only')
$envRun=Invoke-RecordedProcess $shell @('-NoProfile','-EncodedCommand',(Encoded '[Console]::Out.WriteLine($env:TALENRO_DIAG_ENV_TEST)')) @{TALENRO_DIAG_ENV_TEST=$wrappedEnv} (Join-Path $root 'env') 10
Assert-True ((Get-Content -Raw -LiteralPath $envRun.StdoutPath).Trim() -eq 'child-only' -and [Environment]::GetEnvironmentVariable($envName) -eq $beforeEnv) 'environment override remains child-only'
Write-Output "Fixtures retained: $root"
$orphanCode='Start-Process -FilePath "'+$shell+'" -ArgumentList "-NoProfile -EncodedCommand '+$childCode+'" -WindowStyle Hidden; Start-Sleep -Milliseconds 300; exit 0'
$orphan=Invoke-RecordedProcess $shell @('-NoProfile','-EncodedCommand',(Encoded $orphanCode)) @{} (Join-Path $root 'root-exit-child-live') 10
Assert-True ($orphan.RootExited -and -not $orphan.TimedOut -and $orphan.DescendantsRemaining -ge 1 -and $orphan.CleanupVerified -and $orphan.MembersAfterCleanup.Count -eq 0) 'root exit cannot conceal live owned child'
Assert-Throws {Assert-ArmResult 'baseline' $orphan @()} 'cleaned live child still fails baseline gate'
$evidenceProperty=$orphan.PSObject.Properties['MemberEvidence']
$evidence=@(if($null -ne $evidenceProperty){$evidenceProperty.Value})
Assert-True (@($evidence | Where-Object {$null -ne $_ -and $_.Owned -and $_.State -eq 'running' -and [IO.Path]::GetFileName($_.ImagePath) -eq 'powershell.exe'}).Count -ge 1) 'live owned child identity and state captured before cleanup'
$sampled=Invoke-RecordedProcess $shell @('-NoProfile','-EncodedCommand',(Encoded $parentCode)) @{} (Join-Path $root 'live-sampling') 2 $true
$liveProperty=$sampled.PSObject.Properties['MembersDuringRun']
$live=@(if($null -ne $liveProperty){$liveProperty.Value})
Assert-True (@($live | Where-Object {$null -ne $_ -and $_.Id -ne $sampled.RootId -and $_.Owned -and $_.State -eq 'running' -and [IO.Path]::GetFileName($_.ImagePath) -eq 'powershell.exe'}).Count -ge 1) 'opt-in sampling captures child while root is still running'
$eventName='Local\TalenroNatural-'+[guid]::NewGuid().ToString('N')
$shortChild=Encoded ('$ready=[Threading.EventWaitHandle]::OpenExisting("'+$eventName+'");$ready.Set()|Out-Null;[Threading.Thread]::Sleep(300);exit 0')
$shortParent='$ready=New-Object Threading.EventWaitHandle($false,[Threading.EventResetMode]::ManualReset,"'+$eventName+'");Start-Process -FilePath "'+$shell+'" -ArgumentList "-NoProfile -EncodedCommand '+$shortChild+'" -WindowStyle Hidden;if(-not $ready.WaitOne(5000)){exit 9};exit 0'
$natural=Invoke-RecordedProcess $shell @('-NoProfile','-EncodedCommand',(Encoded $shortParent)) @{} (Join-Path $root 'natural-short-child') 10
$naturalFlag=$natural.PSObject.Properties['NaturalShutdownVerified']
Assert-True ($null -ne $naturalFlag -and $naturalFlag.Value -and $natural.DescendantsRemaining -ge 1 -and -not $natural.TerminationRequested -and $natural.CleanupVerified) 'short-lived real child exits naturally without erasing initial remainder'
Assert-ArmResult 'baseline' $natural @()
Assert-True $true 'natural receipt accepted only after verified empty group'
foreach($case in @(
    @('TerminationRequested',$true,'forced cleanup cannot masquerade as natural success'),
    @('ShutdownEvidenceComplete',$false,'unknown member evidence remains failure'),
    @('NaturalShutdownVerified',$false,'empty after cleanup does not prove natural shutdown'),
    @('NaturalObservationMs',501,'natural observation cannot exceed approved window'),
    @('ShutdownElapsedMs',5001,'natural and forced cleanup share the original bound'),
    @('ActiveBeforeTermination',1,'natural decision cannot retain active accounting'),
    @('MembersBeforeTermination',@(42),'natural decision cannot retain job members'),
    @('ShutdownSchema',99,'unknown shutdown schema refused')
)) {
    $changed=$natural | ConvertTo-Json -Depth 6 | ConvertFrom-Json
    $changed.($case[0])=$case[1]
    Assert-Throws {Assert-ProcessCompletion $changed} $case[2]
}
Assert-Throws {Assert-ProcessCompletion $timeout} 'timed-out real process remains failed after successful cleanup'
