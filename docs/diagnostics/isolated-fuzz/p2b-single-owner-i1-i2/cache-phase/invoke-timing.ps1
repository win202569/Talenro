param([Parameter(Mandatory)][ValidateSet('TimingRed','TimingGreen')][string]$Stage)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'owner.ps1')
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../../..'))
$workspace=Join-Path $repo '.superpowers/sdd/2026-11-07-c12-phase-timing-fix'
$root=New-C12WaitRoot (Join-Path $workspace $Stage)
$go='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
if((Get-FileHash -LiteralPath $go).Hash -cne 'D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F'){throw 'PIN_MISMATCH'}
$environment=@{GOENV='off';GOWORK='off';GOTOOLCHAIN='local';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GOFLAGS='-mod=readonly';GOMAXPROCS='2';CGO_ENABLED='0'}
foreach($name in @('GOPATH','GOMODCACHE','GOCACHE','GOTMPDIR','TEMP','TMP')){
    $directory=Join-Path $workspace "isolated/$name";Assert-C12WaitPath $directory;[void][IO.Directory]::CreateDirectory($directory);$environment[$name]=$directory
}
$before=Get-C12WaitModuleHashes $repo
$sources=@{}
foreach($file in Get-ChildItem -LiteralPath $PSScriptRoot -File -Recurse){$sources[[IO.Path]::GetRelativePath($PSScriptRoot,$file.FullName)]=(Get-FileHash -LiteralPath $file.FullName).Hash}
$arguments=@('test','-v','-count=1','-timeout=30s','./...')
$null=Write-C12WaitRecord $root 'claim.json' @{Stage=$Stage;Sources=$sources;Before=$before;Executable=$go;Arguments=$arguments;DeadlineMs=180000;Environment=$environment}
$result=Invoke-C12PhaseOwned -Executable $go -Arguments $arguments -WorkingDirectory $PSScriptRoot -Environment $environment -Root $root -DeadlineMs 180000
$after=Get-C12WaitModuleHashes $repo;$unchanged=$true;foreach($key in $before.Keys){if($before[$key] -cne $after[$key]){$unchanged=$false}}
$passed=$result.Started -and $result.ExitConfirmed -and -not $result.TimedOut -and $null -eq $result.Failure -and $result.ExitCode -eq 0 -and $unchanged
$receipt=@{Stage=$Stage;Passed=$passed;Result=$result;Sources=$sources;Before=$before;After=$after;ModuleFilesUnchanged=$unchanged;Acceptance='NOT_ACCEPTED'}
$null=Write-C12WaitRecord $root 'stage-result.json' $receipt
$receipt|ConvertTo-Json -Depth 15
if(-not $passed){exit 1}
