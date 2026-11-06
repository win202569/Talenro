param([ValidateSet('ParserRed','ParserGreen','RecordsRed','RecordsGreen','Combined')][string]$Phase)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=Split-Path (Split-Path $PSScriptRoot)
$workspace=Join-Path $repo '.superpowers/sdd/2026-11-06-c12-shadow-audit-integration-v1'
$phaseRoot=Join-Path $workspace $Phase
if(Test-Path $phaseRoot){throw 'occupied test phase'}
$null=[IO.Directory]::CreateDirectory($phaseRoot)
$utf8=[Text.UTF8Encoding]::new($false)
function Save-Test($path,$value){$b=$utf8.GetBytes((ConvertTo-Json -InputObject $value -Depth 80));$s=[IO.File]::Open($path,[IO.FileMode]::CreateNew);try{$s.Write($b,0,$b.Length)}finally{$s.Dispose()}}
function Bytes-Test($path,[byte[]]$bytes){$null=[IO.Directory]::CreateDirectory((Split-Path $path));$s=[IO.File]::Open($path,[IO.FileMode]::CreateNew);try{$s.Write($bytes,0,$bytes.Length)}finally{$s.Dispose()}}
Save-Test (Join-Path $phaseRoot claim.json) @{Phase=$Phase;Consumed=$true}
$results=[Collections.Generic.List[object]]::new();$current='load';$serial=0
$hash='BA7816BF8F01CFEA414140DE5DAE2223B00361A396177A9CB410FF61F20015AD'
function Check($ok,$why){if(-not $ok){throw "SHADOW_TEST: $why"}}
function New-Fixture($raw=$null,$content='abc'){
 $script:serial++;$root=Join-Path $phaseRoot ('fixture{0:D3}' -f $script:serial);$null=[IO.Directory]::CreateDirectory($root)
 if($null -eq $raw){$raw='{"Schema":1,"Mode":"SHADOW","Files":[{"Path":"a.txt","SHA256":"'+$hash+'"}]}'}
 Bytes-Test (Join-Path $root testdata/c12/shadow-source-manifest.v1.json) ($utf8.GetBytes($raw))
 if($null -ne $content){Bytes-Test (Join-Path $root a.txt) ($utf8.GetBytes($content))}
 return $root
}
function Scan-Fixture($root,$deadline=[datetime]::UtcNow.AddMinutes(5)){$m=Read-C12ShadowManifest $root;Get-C12ShadowMap $root $m $deadline}
function Record-Pass($name){$script:results.Add(@{Case=$name;Matched=$true});Write-Output "PASS: $name"}
try{
 $module=Join-Path $repo scripts/private/c12-shadow-audit.ps1
 if(-not(Test-Path $module)){throw 'SHADOW_RED: missing module'}
 . $module
 if($Phase -eq 'RecordsRed' -and -not(Get-Command New-C12ShadowSession -ErrorAction SilentlyContinue)){throw 'SHADOW_RED: missing lifecycle interface'}
 $current='valid';$r=Scan-Fixture (New-Fixture);Check ($r.State -ceq 'MATCHED' -and $r.Files.Count -eq 1 -and $r.Files[0].SHA256 -ceq $hash -and $r.Files[0].Present -is [bool] -and $r.Files[0].Present) 'valid map';Record-Pass $current
 $current='drift';$r=Scan-Fixture (New-Fixture $null 'changed');Check ($r.State -ceq 'DRIFT') 'changed source';Record-Pass $current
 $current='missing';$r=Scan-Fixture (New-Fixture $null $null);Check ($r.State -ceq 'DRIFT' -and -not $r.Files[0].Present -and $null -eq $r.Files[0].SHA256) 'missing source';Record-Pass $current
 $current='schema';foreach($raw in @('{bad','{"Schema":true,"Mode":"SHADOW","Files":[]}','{"Schema":1,"Schema":1,"Mode":"SHADOW","Files":[]}','{"Schema":1,"schema":1,"Mode":"SHADOW","Files":[]}','{"Schema":1,"Mode":"SHADOW","Files":[],"Extra":0}','{"Schema":1,"Mode":"SHADOW","Files":[{"Path":"a.txt","SHA256":true}]}')){$r=Scan-Fixture (New-Fixture $raw);Check ($r.State -ceq 'INCOMPLETE') 'invalid schema accepted'};Record-Pass $current
 $current='paths';foreach($path in @('../a.txt','D:/a.txt','//server/a','a.txt:stream','*.txt','x/./a.txt','x//a.txt','a.txt ')){$raw=@{Schema=1;Mode='SHADOW';Files=@(@{Path=$path;SHA256=$hash})}|ConvertTo-Json -Depth 5;$r=Scan-Fixture (New-Fixture $raw);Check ($r.State -ceq 'INCOMPLETE') "unsafe path $path"}
 $raw=@{Schema=1;Mode='SHADOW';Files=@(@{Path='a.txt';SHA256=$hash},@{Path='A.txt';SHA256=$hash})}|ConvertTo-Json -Depth 5;$r=Scan-Fixture (New-Fixture $raw);Check ($r.State -ceq 'INCOMPLETE') 'duplicate path';Record-Pass $current
 $current='limits';$raw=@{Schema=1;Mode='SHADOW';Files=@(1..129|ForEach-Object {@{Path="f$_.txt";SHA256=$hash}})}|ConvertTo-Json -Depth 5;$r=Scan-Fixture (New-Fixture $raw);Check ($r.Reason -ceq 'LIMIT') 'count limit'
 $r=Scan-Fixture (New-Fixture (' '*1048577));Check ($r.Reason -ceq 'LIMIT') 'manifest limit'
 $root=New-Fixture $null $null;$s=[IO.File]::Open((Join-Path $root a.txt),[IO.FileMode]::CreateNew);try{$s.SetLength(33554433)}finally{$s.Dispose()};$r=Scan-Fixture $root;Check ($r.Reason -ceq 'LIMIT') 'aggregate limit';Record-Pass $current
 $current='deadline';$r=Scan-Fixture (New-Fixture) ([datetime]::UtcNow.AddSeconds(-1));Check ($r.State -ceq 'INCOMPLETE' -and $r.Reason -ceq 'DEADLINE') 'deadline';Record-Pass $current
 $current='reparse';$root=New-Fixture;$target=Join-Path $root target;$null=[IO.Directory]::CreateDirectory($target)
 $junction=Join-Path $root link
 try{$null=New-Item -ItemType Junction -Path $junction -Target $target}catch{throw 'SHADOW_COVERAGE_BLOCKED: local junction unavailable'}
 $r=Scan-Fixture $junction;Check ($r.State -ceq 'INCOMPLETE' -and $r.Reason -ceq 'UNSAFE_PATH') 'reparse root accepted';Record-Pass $current
 Save-Test (Join-Path $phaseRoot result.json) @{Status='SHADOW_PARSER_VERIFIED';Phase=$Phase;Results=$results.ToArray();Overall='NOT_ACCEPTED'}
 'SHADOW_PARSER_VERIFIED'
}catch{
 Save-Test (Join-Path $phaseRoot result.json) @{Status='STOPPED';Phase=$Phase;Case=$current;Reason=$_.Exception.Message;Stack=$_.ScriptStackTrace;Results=$results.ToArray();NoRetry=$true;Overall='NOT_ACCEPTED'}
 throw
}
