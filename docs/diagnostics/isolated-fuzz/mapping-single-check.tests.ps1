$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'matrix-check.ps1')
$helper=Join-Path $PSScriptRoot 'mapping-single-check.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
function Check-Mapping([bool]$Value,[string]$Name){if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
Check-Mapping ($null -ne (Get-Command Invoke-MappingSingleCheck -ErrorAction SilentlyContinue)) 'single fixed-package entry exists'
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
function Test-MappingCase([string]$Scenario) {
    $unit=Join-Path $scope ('mapping-unit-'+[guid]::NewGuid().ToString('N'))
    $root=Join-Path $unit 'mapping-check-901';$tool=Join-Path $unit 'uninstrumented-901/toolchain';$modules=Join-Path $unit 'modules'
    foreach($dir in @((Join-Path $tool 'bin'),(Join-Path $modules 'github.com/google/uuid@v1.6.0'))){New-Item -ItemType Directory -Path $dir -Force|Out-Null}
    [IO.File]::WriteAllText((Join-Path $tool 'bin/go.exe'),'never executed')
    [IO.File]::WriteAllText((Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod'),'module fixture')
    if($Scenario -eq 'existing'){New-Item -ItemType Directory -Path $root|Out-Null}
    if($Scenario -eq 'wrong-root'){$root=Join-Path $unit 'matrix-check-901'}
    $nativeCalls=New-Object 'System.Collections.Generic.List[object]';$source=[pscustomobject]@{Calls=0};$product=[pscustomobject]@{Calls=0}
    function Get-EnvelopeProductManifest {
        $product.Calls++
        [pscustomobject]@{Path='fixture.go';SHA256=$(if($Scenario -eq 'post-product' -and $product.Calls -gt 1){'C'*64}else{'A'*64})}
    }
    function Assert-UninstrumentedEnvelopeToolchain($Toolchain,$ReferenceReceipt) {
        $source.Calls++
        if($Scenario -eq 'pre-tool' -or ($Scenario -eq 'post-tool' -and $source.Calls -gt 1)){throw 'controlled tool drift'}
        if($source.Calls -gt 1 -and $Scenario -in @('post-build-binding','post-runtime-binding')){
            $phase=if($Scenario -eq 'post-build-binding'){'build'}else{'execute'}
            [IO.File]::WriteAllText((Join-Path $root ($phase+'/stdout.txt')),'changed evidence')
        }
        return (('A'*64)+':'+('B'*64))
    }
    function Invoke-MatrixNative($Executable,$Arguments,$Environment,$WorkingDirectory,$Stdout,$Stderr,$LimitSeconds) {
        $build=$Executable -ceq (Join-Path $tool 'bin/go.exe')
        $nativeCalls.Add([pscustomobject]@{Build=$build;Arguments=$Arguments;WorkingDirectory=$WorkingDirectory;Limit=$LimitSeconds})
        $binary=Join-Path $root 'package.test.exe'
        $expected=if($build){@('test','-c','-x','-p=1','-count=1','-timeout=120s',('-o='+$binary),'talenro.local/platform/internal/nodecontrol/nodeissuermapping')}else{@('-test.count=1','-test.timeout=120s','-test.v=true','-test.run=.')}
        if(($Arguments -join '|') -cne ($expected -join '|') -or $LimitSeconds -ne $(if($build){600}else{150}) -or $WorkingDirectory -ine $(if($build){$workspace}else{Join-Path $workspace 'internal/nodecontrol/nodeissuermapping'})){throw 'Unexpected target/args/cwd/budget'}
        foreach($key in @('GOTOOLCHAIN','GOENV','GOPROXY','GOSUMDB','GOAUTH','GOVCS','GO111MODULE','GOWORK','GOFLAGS','GOOS','GOARCH','TALENRO_FUZZ_DIAG_MODE','TALENRO_FUZZ_FIXTURE','GODEBUG')){
            $value=@{GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='on';GOWORK='off';GOFLAGS='-mod=readonly';GOOS='windows';GOARCH='amd64';TALENRO_FUZZ_DIAG_MODE='';TALENRO_FUZZ_FIXTURE='';GODEBUG=''}[$key]
            if($Environment[$key] -cne $value){throw ('Offline environment changed: '+$key)}
        }
        if($Environment.GOROOT -cne $tool -or $Environment.GOMODCACHE -cne $modules -or $Environment.GOCACHE -cne (Join-Path $root 'go-cache') -or $Environment.GOTMPDIR -cne (Join-Path $root 'tmp')){throw 'Fresh cache/tool environment missing'}
        $r=New-Object 'IsolatedFuzzProcess+Result'
        foreach($entry in @{RootId=17;ExitCode=0;RootExited=$true;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;CleanupVerified=$true;MembersAtRootWait=@();MembersBeforeTermination=@();MembersAfterCleanup=@();ShutdownMemberEvidence=@();StartUTC='2026-01-01T00:00:00Z';EndUTC='2026-01-01T00:00:01Z'}.GetEnumerator()){$r.($entry.Key)=$entry.Value}
        $q=New-Object 'IsolatedFuzzProcess+AccountingQueryObservation'
        foreach($entry in @{API='QueryInformationJobObject/Accounting';Succeeded=$true;Error=0;ValuesRead=$true;StartUTC='2026-01-01T00:00:00.1Z';EndUTC='2026-01-01T00:00:00.2Z';StartTimestamp=1;EndTimestamp=2;Frequency=10}.GetEnumerator()){$q.($entry.Key)=$entry.Value}
        $r.InitialAccountingQuery=$q
        $m=New-Object 'IsolatedFuzzProcess+MemberListQueryObservation'
        foreach($entry in @{API='QueryInformationJobObject/ProcessIdList';Succeeded=$true;Error=0;HeaderRead=$true;Attempt=1;Capacity=16;BufferBytes=(8+16*[IntPtr]::Size);StartUTC='2026-01-01T00:00:00.3Z';EndUTC='2026-01-01T00:00:00.4Z';StartTimestamp=3;EndTimestamp=4;Frequency=10}.GetEnumerator()){$m.($entry.Key)=$entry.Value}
        $r.InitialMemberQueries.Add($m)
        $out="=== RUN   TestFixture`r`n--- PASS: TestFixture (0.00s)`r`nPASS";$err=''
        if($build){
            if($Scenario -ne 'missing-binary'){[IO.File]::WriteAllText($binary,'never executed')}
            $test=@(Get-ChildItem (Join-Path $workspace 'internal/nodecontrol/nodeissuermapping') -Filter '*_test.go' -File)[0].FullName
            $compiler=Join-Path $tool 'pkg/tool/windows_amd64/compile.exe'
            $err='"'+$compiler+'" -p talenro.local/platform/internal/nodecontrol/nodeissuermapping -pack "'+$test+'"'+"`r`n"+'"'+$compiler+'" -p context -pack "'+(Join-Path $tool 'src/context/context.go')+'"'+"`r`n"+'"'+$compiler+'" -p internal/fuzz -pack "'+(Join-Path $tool 'src/internal/fuzz/fuzz.go')+'"'
            if($Scenario -eq 'build-exit'){$r.ExitCode=7}
            if($Scenario -eq 'compile-evidence'){$err='missing evidence'}
            if($Scenario -eq 'build-gate'){$r.NaturalShutdownVerified=$false}
        }else{
            switch($Scenario){
                'run-exit'{$r.ExitCode=7}
                'run-timeout'{$r.TimedOut=$true}
                'run-gate'{$r.NaturalShutdownVerified=$false;$r.TerminationRequested=$true}
                'no-tests'{$out='PASS'}
                'no-pass'{$out='--- PASS: TestFixture (0.00s)'}
                'diagnostic-hook'{$err='TALENRO_FUZZ_DIAG invalid'}
                'missing-query'{$r.InitialMemberQueries.Clear()}
            }
        }
        [IO.File]::WriteAllText($Stdout,$out);[IO.File]::WriteAllText($Stderr,$err)
        return $r
    }
    $failure=$null;$seen=@()
    try{$seen=@(Invoke-MappingSingleCheck $root $tool (Join-Path $unit 'reference.json') $modules)}catch{$failure=$_}
    $resultPath=Join-Path $root 'mapping-result.json'
    $saved=if(Test-Path -LiteralPath $resultPath){Get-Content $resultPath -Raw|ConvertFrom-Json}else{$null}
    [pscustomobject]@{Root=$root;Calls=@($nativeCalls.ToArray());Failure=$failure;Seen=$seen;Saved=$saved}
}
$good=Test-MappingCase 'success'
Check-Mapping ($null -eq $good.Failure -and $good.Calls.Count -eq 2 -and $good.Calls[0].Build -and -not $good.Calls[1].Build -and $good.Seen.Count -eq 1 -and $good.Saved.PackagePassed -and -not $good.Saved.Acceptance) 'only one build and one fixed-package runtime publish final scoped success'
Check-Mapping ($good.Saved.Bindings.Count -eq 2 -and $good.Saved.Bindings[0].Files.Count -eq 4 -and $good.Saved.Bindings[1].Files.Count -eq 3) 'compile binary and both receipt stream sets are integrity bound'
$missing=Test-MappingCase 'missing-binary'
Check-Mapping ($null -ne $missing.Failure -and $missing.Calls.Count -eq 1 -and -not $missing.Saved.CompilePassed -and -not $missing.Saved.ExecutionAttempted) 'missing binary cannot mark compile passed or attempt runtime'
foreach($case in @(@('existing',0),@('wrong-root',0),@('pre-tool',0),@('build-exit',1),@('build-gate',1),@('compile-evidence',1),@('run-exit',2),@('run-timeout',2),@('run-gate',2),@('no-tests',2),@('no-pass',2),@('diagnostic-hook',2),@('missing-query',2),@('post-product',2),@('post-tool',2),@('post-build-binding',2),@('post-runtime-binding',2))){
    $bad=Test-MappingCase $case[0]
    Check-Mapping ($null -ne $bad.Failure -and $bad.Calls.Count -eq $case[1] -and $bad.Seen.Count -eq 0) ($case[0]+' stops without retry or success output')
    if($null -ne $bad.Saved){Check-Mapping (-not $bad.Saved.PackagePassed -and -not $bad.Saved.Acceptance -and (Test-Path -LiteralPath (Join-Path $bad.Root 'stopped.json'))) ($case[0]+' retains unaccepted failed checkpoint')}
}
