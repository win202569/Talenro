Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$helper=Join-Path $PSScriptRoot runtime-staged-evidence-fixtures.ps1
if(-not(Test-Path $helper)){throw 'SE_BOOTSTRAP: missing helper'}
. $helper
$repo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
$workspace=Join-Path $repo '.superpowers/sdd/2026-11-05-staged-evidence-fixtures-v1'
$source=[IO.File]::ReadAllText((Join-Path $PSScriptRoot runtime-staged-integrity-review-v2.ps1))
function Check([bool]$Ok,[string]$Why){if(-not $Ok){throw "SE_BOOTSTRAP: $Why"}}
function Reject([scriptblock]$Body){$caught=$false;try{& $Body|Out-Null}catch{$caught=$true};Check $caught 'unsafe input accepted'}
$root=Join-Path $workspace 'bootstrap-root';$review='fixture-review'
$rebound=ConvertTo-SEFixtureVerifier $source $root $review
$inverse=$rebound.Replace("`$root='$root'","`$root='$repo'").Replace("`$review='$review'","`$review='.superpowers/sdd/2026-11-05-runtime-staged-integrity-review-v2'")
Check ($inverse -ceq $source) 'decision code changed'
Check (-not $rebound.Contains("`$root='$repo'")) 'old root binding retained'
[Console]::Out.WriteLine('PASS: B01')
Reject {ConvertTo-SEFixtureVerifier $source 'D:\outside-fixture' $review}
Reject {ConvertTo-SEFixtureVerifier $source $root '../escape'}
Reject {ConvertTo-SEFixtureVerifier ($source.Replace("`$root=","`$otherRoot=")) $root $review}
$occupied=Join-Path $workspace 'bootstrap-occupied'
$null=New-Item -ItemType Directory -Path (Join-Path $occupied C01)
Reject {New-SEFixture C01 $occupied}
Reject {New-SEFixture '../escape' $workspace}
[Console]::Out.WriteLine('PASS: B02')
$null=New-Item -ItemType Directory -Path $root
$one=Invoke-SEFixture 'param($InputSHA256) "fixture"; exit 1' $root ('A'*64)
Check ($one.Terminal -and 'fixture' -in $one.Output) 'exit isolation'
$two=Invoke-SEFixture 'param($InputSHA256) "success"' $root ('A'*64)
Check ($two.Terminal -and -not $two.HadErrors -and 'success' -in $two.Output) 'fresh success runspace'
[Console]::Out.WriteLine('PASS: B03')
[Console]::Out.WriteLine('STAGED_EVIDENCE_BOOTSTRAP_VERIFIED')
