param([Parameter(Mandatory)][ValidateSet('GoRed','GoGreen','OwnerRed','OwnerGreen','Build','Live')][string]$Stage)
$ErrorActionPreference='Stop'
$repo='D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator'
$w=Join-Path $repo '.superpowers/sdd/2026-11-07-c12-libc-profile'
$bootstrap=Join-Path $w 'Bootstrap/cache-phase/owner.ps1'
if((Get-FileHash $bootstrap).Hash -cne '33782F9A2F723AEAAF816D9C3ADD25E3AA9D80954BBAB9F314B8309B15DD0B7E'){throw 'BOOTSTRAP_PIN'}
. $bootstrap
. (Join-Path $PSScriptRoot 'manifest.ps1')
foreach($p in @((Join-Path $PSScriptRoot '../runtime-cache-wait-common.ps1'),(Join-Path $w 'Bootstrap/runtime-cache-wait-common.ps1'))){if((Get-FileHash $p).Hash -cne '77824603B0817D32B447C50CA3818F5E34E1924D8403A800073165C65081566F'){throw 'COMMON_PIN'}}
Assert-C12WaitPath $repo;Assert-C12WaitPath $PSScriptRoot
$root=New-C12WaitRoot (Join-Path $w $Stage)
$go='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
Assert-C12WaitPath $go
if((Get-FileHash $go).Hash -cne 'D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F'){throw 'GO_PIN'}
$envMap=@{GOENV='off';GOWORK='off';GOTOOLCHAIN='local';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GOFLAGS='-mod=readonly';GOMAXPROCS='2';CGO_ENABLED='0'}
foreach($n in @('GOPATH','GOMODCACHE','GOCACHE','GOTMPDIR','TEMP','TMP')){$p=Join-Path $w "isolated/$n";Assert-C12WaitPath $p;[void][IO.Directory]::CreateDirectory($p);$envMap[$n]=$p}
$before=Get-C12WaitModuleHashes $repo
$setup=Read-C12WaitRecord $w 'setup.json';Assert-C12ProfileManifest $setup.Before $before
$snap=Join-Path $root 'source';[void][IO.Directory]::CreateDirectory($snap);$sources=@{}
foreach($f in Get-ChildItem -LiteralPath $PSScriptRoot -File -Recurse){Assert-C12WaitPath $f.FullName;$rel=[IO.Path]::GetRelativePath($PSScriptRoot,$f.FullName);$dest=Join-Path $snap $rel;[void][IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($dest));[IO.File]::Copy($f.FullName,$dest,$false);$sources[$rel]=(Get-FileHash $f.FullName).Hash}
$arguments=@('test','-count=1','-timeout=30s','./...')
$manifest=$sources.Clone()
foreach($rel in @('docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-cache-wait-common.ps1','docs/superpowers/specs/2026-11-07-c12-libc-profile-design.md','docs/superpowers/plans/2026-11-07-c12-libc-profile.md')){$manifest["repo:$rel"]=(Get-FileHash (Join-Path $repo $rel)).Hash}
$executable=$go;$deadline=180000;$session='';$profile='paired-v1';$expected=$null
if($Stage -in @('OwnerRed','OwnerGreen')){$executable=[Diagnostics.Process]::GetCurrentProcess().MainModule.FileName;$deadline=60000;$arguments=@('-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot 'profile.tests.ps1'),'-Phase',$Stage,'-Root',(Join-Path $w "$Stage-fixtures"))}
if($Stage -in @('Build','Live')){
 foreach($green in @('GoGreen','OwnerGreen')){
  $v=Read-C12WaitRecord (Join-Path $w $green) 'stage-result.json'
  if($null -eq $v -or -not $v.ModuleFilesUnchanged -or -not $v.Result.Started -or -not $v.Result.ExitConfirmed -or $v.Result.ExitCode -ne 0 -or $v.Result.TimedOut -or $null -ne $v.Result.Failure){throw 'GREEN_REQUIRED'}
  $keys=if($green -ceq 'GoGreen'){@($sources.Keys|Where-Object {$_ -like '*.go' -or $_ -ceq 'go.mod'})}else{@('owner.ps1','profile.tests.ps1','profile-fixture.ps1')}
  $old=@{};$current=@{}
  foreach($k in $v.Sources.Keys){if(($green -ceq 'GoGreen' -and ($k -like '*.go' -or $k -ceq 'go.mod')) -or ($green -ceq 'OwnerGreen' -and $k -in $keys)){$old[$k]=$v.Sources[$k]}}
  foreach($k in $keys){$current[$k]=$sources[$k]};Assert-C12ProfileManifest $old $current
 }
 $test=Read-C12WaitRecord (Join-Path $w 'OwnerGreen-fixtures') 'test-result.json'
 if($test.Status -cne 'PASS' -or $test.Count -ne 18){throw 'OWNER_TESTS_REQUIRED'}
 $status=git -c safe.directory=$repo -C $repo status --porcelain
 if($LASTEXITCODE -ne 0 -or $status){throw 'COMMITTED_SOURCE_REQUIRED'}
}
if($Stage -ceq 'Build'){$arguments=@('build','-buildvcs=false','-trimpath','-o',(Join-Path $root 'helper.exe'),'.')}
if($Stage -ceq 'Live'){
 $build=Read-C12WaitRecord (Join-Path $w 'Build') 'stage-result.json'
 if($null -eq $build -or -not $build.Passed){throw 'BUILD_REQUIRED'}
 Assert-C12ProfileManifest $build.Manifest $manifest
 $executable=Join-Path $w 'Build/helper.exe';Assert-C12WaitPath $executable
 if((Get-FileHash $executable).Hash -cne $build.BinarySHA256){throw 'BINARY_CHANGED'}
 $cache='C:/Users/Lenovo/go/pkg/mod';$envMap.GOPATH='C:/Users/Lenovo/go';$envMap.GOMODCACHE=$cache
 foreach($p in @($cache,(Join-Path $cache 'modernc.org/libc@v1.72.1'),(Join-Path $cache 'cache/download/modernc.org/libc/@v/v1.72.1.zip'),(Join-Path $cache 'cache/download/modernc.org/libc/@v/v1.72.1.ziphash'))){Assert-C12WaitPath $p;if(-not(Test-Path -LiteralPath $p)){throw 'LIVE_INPUT_MISSING'}}
 $sumPath=Join-Path $repo 'go.sum';$zhPath=Join-Path $cache 'cache/download/modernc.org/libc/@v/v1.72.1.ziphash';Assert-C12WaitPath $sumPath
 if((Get-Item $sumPath).Length -gt 4194304 -or (Get-Item $zhPath).Length -gt 1024){throw 'EXPECTED_LIMIT'}
 $expected='h1:db1xwJ6u1kE3KHTFTTbe2GCrczHPKzlURP0aDC4NGD0='
 $matches=@(Get-Content $sumPath|Where-Object {$_ -match '^modernc\.org/libc v1\.72\.1\s'})
 if($matches.Count -ne 1 -or ($matches[0] -split '\s+')[2] -cne $expected -or [IO.File]::ReadAllText($zhPath).Trim() -cne $expected){throw 'EXPECTED_MISMATCH'}
 . (Join-Path $PSScriptRoot 'owner.ps1')
 $session=[guid]::NewGuid().ToString('N');$profile='libc-v1';$deadline=120000;$arguments=@('-profile',$profile,'-session',$session,'-repo',$repo)
}
$head=git -c safe.directory=$repo -C $repo rev-parse HEAD;if($LASTEXITCODE){throw 'HEAD'}
$null=Write-C12WaitRecord $root 'claim.json' @{Stage=$Stage;SourceCommit=$head;Profile=$profile;Session=$session;Expected=$expected;Before=$before;Sources=$sources;Manifest=$manifest;Arguments=$arguments;Executable=$executable;ExecutableSHA256=(Get-FileHash $executable).Hash;Environment=$envMap;DeadlineMs=$deadline}
$params=@{Executable=$executable;Arguments=$arguments;WorkingDirectory=$PSScriptRoot;Environment=$envMap;Root=$root;DeadlineMs=$deadline}
if($Stage -ceq 'Live'){$params.Session=$session;$params.Profile=$profile}
$r=Invoke-C12PhaseOwned @params
$after=Get-C12WaitModuleHashes $repo;$unchanged=$true;foreach($k in $before.Keys){if($before[$k] -cne $after[$k]){$unchanged=$false}}
$passed=$r.Started -and $r.ExitConfirmed -and $r.ExitCode -eq 0 -and -not $r.TimedOut -and $null -eq $r.Failure -and $r.EndReason -ceq 'EXITED' -and $unchanged
$events=$null;$parseFailure=$null;$binaryHash=$null
if($Stage -ceq 'Live'){
 try{$events=Read-C12PhaseEvents $r.StdoutPath $session $profile}catch{$parseFailure=$_.Exception.Message}
 $passed=$passed -and $null -eq $parseFailure -and $null -ne $events -and $events.Complete -and $events.Events.Count -eq 6 -and $events.Phases.Count -eq 3 -and (Get-Item $r.StderrPath).Length -eq 0
 if($passed){foreach($e in $events.Events){if($e.Kind -ceq 'END' -and ($e.Status -cne 'OK' -or $e.Expected -cne $expected -or ($e.Phase -cne 'PREPARE' -and $e.Digest -cne $expected))){$passed=$false}}}
}
if($Stage -ceq 'Build' -and $passed){$binaryHash=(Get-FileHash (Join-Path $root 'helper.exe')).Hash}
$v=@{Stage=$Stage;Profile=$profile;Session=$session;Passed=$passed;Result=$r;Before=$before;After=$after;ModuleFilesUnchanged=$unchanged;Sources=$sources;Manifest=$manifest;BinarySHA256=$binaryHash;Events=$events;ParseFailure=$parseFailure;Acceptance='NOT_ACCEPTED'}
$null=Write-C12WaitRecord $root 'stage-result.json' $v
$v|ConvertTo-Json -Depth 12 -Compress
