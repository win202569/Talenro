param([Parameter(Mandatory)][ValidateSet('GoRed','GoGreen','Build','Live')][string]$Stage)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'owner.ps1')
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../../..'))
$workspace=Join-Path $repo '.superpowers/sdd/2026-11-06-c12-cache-phase-v1'
$root=New-C12WaitRoot (Join-Path $workspace $Stage)
$go='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
if((Get-FileHash -LiteralPath $go).Hash -cne 'D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F'){throw 'PIN_MISMATCH'}
$environment=@{GOENV='off';GOWORK='off';GOTOOLCHAIN='local';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GOFLAGS='-mod=readonly';GOMAXPROCS='2';CGO_ENABLED='0'}
foreach($name in @('GOPATH','GOMODCACHE','GOCACHE','GOTMPDIR','TEMP','TMP')){
    $directory=Join-Path $workspace "isolated/$name";Assert-C12WaitPath $directory;[void][IO.Directory]::CreateDirectory($directory);$environment[$name]=$directory
}
$before=Get-C12WaitModuleHashes $repo
$sources=@{};foreach($file in Get-ChildItem -LiteralPath $PSScriptRoot -File -Recurse){if($file.Extension -in @('.go','.mod','.ps1','.json') -or $file.Name -in @('LICENSE','PATENTS')){$sources[[IO.Path]::GetRelativePath($PSScriptRoot,$file.FullName)]=(Get-FileHash -LiteralPath $file.FullName).Hash}}
$session='';$executable=$go;$deadline=180000
switch($Stage){
    GoRed {$arguments=@('test','-count=1','-timeout=30s','./...')}
    GoGreen {$arguments=@('test','-count=1','-timeout=30s','./...')}
    Build {$arguments=@('build','-trimpath','-o',(Join-Path $root 'helper.exe'),'.')}
    Live {
        $build=Read-C12WaitRecord (Join-Path $workspace 'Build') 'stage-result.json'
        if($null -eq $build -or -not $build.Passed){throw 'BUILD_NOT_VERIFIED'}
        $executable=Join-Path $workspace 'Build/helper.exe'
        if((Get-FileHash -LiteralPath $executable).Hash -cne $build.BinarySHA256){throw 'HELPER_HASH_MISMATCH'}
        foreach($key in $build.Sources.Keys){if(-not $sources.ContainsKey($key) -or $sources[$key] -cne $build.Sources[$key]){throw 'SOURCE_CHANGED'}}
        $deadline=120000;$session=[guid]::NewGuid().ToString('N');$arguments=@('-session',$session,'-repo',$repo)
    }
}
$null=Write-C12WaitRecord $root 'claim.json' @{Stage=$Stage;Session=$session;Sources=$sources;Before=$before;Executable=$executable;Arguments=$arguments;DeadlineMs=$deadline;Environment=$environment}
$result=Invoke-C12PhaseOwned -Executable $executable -Arguments $arguments -WorkingDirectory $PSScriptRoot -Environment $environment -Root $root -DeadlineMs $deadline -Session $session
$after=Get-C12WaitModuleHashes $repo;$unchanged=$true;foreach($key in $before.Keys){if($before[$key] -cne $after[$key]){$unchanged=$false}}
$passed=$result.Started -and $result.ExitConfirmed -and -not $result.TimedOut -and $null -eq $result.Failure -and $result.ExitCode -eq 0 -and $unchanged
$binaryHash=$null;if($Stage -ceq 'Build' -and $passed){$binaryHash=(Get-FileHash -LiteralPath (Join-Path $root 'helper.exe')).Hash}
$events=$null;if($Stage -ceq 'Live'){$events=Read-C12PhaseEvents -Path $result.StdoutPath -Session $session}
$receipt=@{Stage=$Stage;Passed=$passed;Result=$result;Sources=$sources;BinarySHA256=$binaryHash;Before=$before;After=$after;ModuleFilesUnchanged=$unchanged;Events=$events;Acceptance='NOT_ACCEPTED'}
$null=Write-C12WaitRecord $root 'stage-result.json' $receipt
$receipt|ConvertTo-Json -Depth 15
if(-not $passed){exit 1}
