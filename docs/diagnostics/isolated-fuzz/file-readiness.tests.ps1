$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'child-readiness.ps1')
function Check-FileReadiness([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Reject-FileReadiness([scriptblock]$Action,[string]$Name) {$refused=$false;try{& $Action|Out-Null}catch{$refused=$true};Check-FileReadiness $refused $Name}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$fixture=Join-Path $scope ('file-readiness-tests-'+[guid]::NewGuid().ToString('N'))
$unicodeFolder="O'Brien "+[char]0x4e2d+[char]0x6587+' fixtures'
$target=Join-Path (Join-Path $fixture $unicodeFolder) 'file-child-readiness-901'
$plan=$null;try{$plan=Get-ChildReadinessPlan $target $true}catch{}
Check-FileReadiness ($null -ne $plan -and $plan.UseScriptFiles -and $plan.LimitSeconds -eq 2) 'explicit file mode requires its fresh identity and preserves whole-process budget'
Check-FileReadiness (-not (Test-Path -LiteralPath $fixture)) 'file planning performs no materialization or process launch'
Reject-FileReadiness {Get-ChildReadinessPlan $target} 'default encoded mode refuses file-mode identity'
Reject-FileReadiness {Get-ChildReadinessPlan (Join-Path $fixture 'child-readiness-901') $true} 'file mode refuses consumed encoded identity class'
$legacy=Get-ChildReadinessPlan (Join-Path $fixture 'child-readiness-902')
Check-FileReadiness (-not $legacy.UseScriptFiles -and $legacy.Arguments[1] -ceq '-EncodedCommand' -and [Text.Encoding]::Unicode.GetString([Convert]::FromBase64String($legacy.Arguments[2])) -ceq $legacy.ParentScript) 'default encoded command behavior remains intact'
Check-FileReadiness (($plan.Arguments -join '|') -ceq ('-NoProfile|-File|'+$plan.ParentScriptPath)) 'native file invocation contains only normal profile/file arguments'
$tokens=$null;$errors=$null;$parentAst=[Management.Automation.Language.Parser]::ParseInput($plan.ParentScript,[ref]$tokens,[ref]$errors)
Check-FileReadiness (@($errors).Count -eq 0) 'file parent parses with apostrophe Unicode and spaced path'
$launch=@($parentAst.FindAll({param($a) $a -is [Management.Automation.Language.CommandAst] -and $a.GetCommandName() -eq 'Start-Process'},$true))[0]
Check-FileReadiness ($launch.CommandElements[4].Value -ceq ('-NoProfile -File "'+$plan.ChildScriptPath+'"') -and $launch.CommandElements[6].Value -ceq 'Hidden') 'hidden child receives quoted file path without policy override'
New-Item -ItemType Directory -Path $target -Force|Out-Null
Write-ChildReadinessScripts $plan
Check-FileReadiness ([IO.File]::ReadAllText($plan.ChildScriptPath) -ceq $plan.ChildScript -and [IO.File]::ReadAllText($plan.ParentScriptPath) -ceq $plan.ParentScript) 'exclusive materialization preserves complete generated programs'
$bytes=[IO.File]::ReadAllBytes($plan.ChildScriptPath)
Check-FileReadiness ($bytes.Length -gt 3 -and $bytes[0] -eq 239 -and $bytes[1] -eq 187 -and $bytes[2] -eq 191) 'script BOM preserves Unicode paths for Windows PowerShell5.1'
$childAst=[Management.Automation.Language.Parser]::ParseInput([IO.File]::ReadAllText($plan.ChildScriptPath),[ref]$tokens,[ref]$errors)
Check-FileReadiness (@($errors).Count -eq 0) 'materialized child parses without executing sleep or launching process'
$prefix=($childAst.EndBlock.Statements|Select-Object -SkipLast 1|ForEach-Object {$_.Extent.Text}) -join "`n"
& ([scriptblock]::Create($prefix))
$marker=Get-Content -LiteralPath $plan.ReadinessPath -Raw|ConvertFrom-Json
Check-FileReadiness ($marker.ProcessId -eq $PID -and $marker.Nonce -ceq $plan.Nonce) 'same-host child prefix writes unchanged self-ID nonce readiness contract'
$beforeHash=(Get-FileHash -LiteralPath $plan.ChildScriptPath).Hash
Reject-FileReadiness {Write-ChildReadinessScripts $plan} 'existing scripts cannot be overwritten'
Check-FileReadiness ((Get-FileHash -LiteralPath $plan.ChildScriptPath).Hash -ceq $beforeHash) 'duplicate materialization leaves existing script bytes intact'
$bad=Get-ChildReadinessPlan (Join-Path $fixture 'file-child-readiness-902') $true
New-Item -ItemType Directory -Path $bad.RunRoot|Out-Null
$bad.ChildScriptPath=Join-Path $fixture 'outside-child.ps1'
Reject-FileReadiness {Write-ChildReadinessScripts $bad} 'materializer refuses an escaped or substituted script destination'
Check-FileReadiness (-not (Test-Path -LiteralPath $bad.ChildScriptPath)) 'rejected destination is not written'
Reject-FileReadiness {Invoke-ChildReadinessFixture $target $true} 'execution refuses existing file identity before any process launch'
$time=[DateTimeOffset]$marker.UTC
$receipt=[pscustomobject]@{RootId=1;StartUTC=$time.AddMilliseconds(-100).ToString('o');TimedOut=$true;CleanupVerified=$true;ActiveAfterCleanup=0;MembersAfterCleanup=@();MemberEvidenceTruncated=$false;MemberEvidence=@([pscustomobject]@{Id=$PID;Owned=$true;State='running';QueryError=0;ExitCodeRead=$true;ExitCode=259;CoreObservedUTC=$time.AddMilliseconds(100).ToString('o')})}
$e=Get-ChildReadinessEvidence $plan $receipt
Check-FileReadiness ($e.ChildReady -and $e.FixturePassed -and -not $e.Acceptance -and -not $e.OriginalTestPassed) 'file-mode evidence reuses the same strict diagnostic evaluator'
$receipt.MemberEvidence[0].Owned=$false
Check-FileReadiness (-not (Get-ChildReadinessEvidence $plan $receipt).FixturePassed) 'file-mode record cannot replace missing ownership proof'
