param([Parameter(Mandatory)][ValidateSet('GoRed','GoGreen','OwnerRed','OwnerGreen','Build','Live')][string]$Stage)
$ErrorActionPreference='Stop'
$repo='D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator'
$w=Join-Path $repo '.superpowers/sdd/2026-11-07-c12-libc-profile'
$bootstrap=Join-Path $w 'Bootstrap/cache-phase/owner.ps1'
if((Get-FileHash $bootstrap).Hash -cne '33782F9A2F723AEAAF816D9C3ADD25E3AA9D80954BBAB9F314B8309B15DD0B7E'){throw 'BOOTSTRAP_PIN'}
. $bootstrap
if($Stage -notin @('GoRed','GoGreen')){throw 'STAGE_NOT_IMPLEMENTED'}
$root=New-C12WaitRoot (Join-Path $w $Stage)
$go='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
Assert-C12WaitPath $go
if((Get-FileHash $go).Hash -cne 'D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F'){throw 'GO_PIN'}
$envMap=@{GOENV='off';GOWORK='off';GOTOOLCHAIN='local';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GOFLAGS='-mod=readonly';GOMAXPROCS='2';CGO_ENABLED='0'}
foreach($n in @('GOPATH','GOMODCACHE','GOCACHE','GOTMPDIR','TEMP','TMP')){$p=Join-Path $w "isolated/$n";Assert-C12WaitPath $p;[void][IO.Directory]::CreateDirectory($p);$envMap[$n]=$p}
$before=Get-C12WaitModuleHashes $repo
$snap=Join-Path $root 'source';[void][IO.Directory]::CreateDirectory($snap);$sources=@{}
foreach($f in Get-ChildItem -LiteralPath $PSScriptRoot -File -Recurse){Assert-C12WaitPath $f.FullName;$rel=[IO.Path]::GetRelativePath($PSScriptRoot,$f.FullName);$dest=Join-Path $snap $rel;[void][IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($dest));[IO.File]::Copy($f.FullName,$dest,$false);$sources[$rel]=(Get-FileHash $f.FullName).Hash}
$arguments=@('test','-count=1','-timeout=30s','./...')
$null=Write-C12WaitRecord $root 'claim.json' @{Stage=$Stage;Before=$before;Sources=$sources;Arguments=$arguments;Executable=$go;Environment=$envMap;DeadlineMs=180000}
$r=Invoke-C12PhaseOwned -Executable $go -Arguments $arguments -WorkingDirectory $PSScriptRoot -Environment $envMap -Root $root -DeadlineMs 180000
$after=Get-C12WaitModuleHashes $repo;$unchanged=$true;foreach($k in $before.Keys){if($before[$k] -cne $after[$k]){$unchanged=$false}}
$v=@{Stage=$Stage;Result=$r;Before=$before;After=$after;ModuleFilesUnchanged=$unchanged;Sources=$sources;Acceptance='NOT_ACCEPTED'}
$null=Write-C12WaitRecord $root 'stage-result.json' $v
$v|ConvertTo-Json -Depth 12 -Compress
