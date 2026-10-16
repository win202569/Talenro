# Fake-only utilities. No workload launches at import or from fixture executables.
function Check-LeafOrdinary([bool]$Value,[string]$Name){if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Reject-LeafOrdinary([scriptblock]$Action,[string]$Name){$failed=$false;try{& $Action|Out-Null}catch{$failed=$true};Check-LeafOrdinary $failed $Name}
function New-LeafOrdinaryTestFiles([string]$FixtureRoot){
    $tool=Join-Path $FixtureRoot 'uninstrumented-901/toolchain';$modules=Join-Path $FixtureRoot 'modules'
    foreach($dir in @((Join-Path $tool 'bin'),(Join-Path $modules 'github.com/google/uuid@v1.6.0'))){New-Item -ItemType Directory -Path $dir -Force|Out-Null}
    foreach($path in @((Join-Path $tool 'bin/go.exe'),(Join-Path $FixtureRoot 'uninstrumented-901/copy-receipt.json'),(Join-Path $FixtureRoot 'reference.json'),(Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod'))){[IO.File]::WriteAllText($path,'fixture data; never executed')}
    [pscustomobject]@{Root=(Join-Path $FixtureRoot 'leaf-ordinary-check-901');Toolchain=$tool;ReferenceReceipt=(Join-Path $FixtureRoot 'reference.json');ModuleCache=$modules}
}
function New-LeafOrdinaryNativeFixture {
    $r=New-Object 'IsolatedFuzzProcess+Result'
    foreach($v in @{RootId=17;RootExited=$true;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;CleanupVerified=$true;StartUTC='2026-10-16T00:00:00Z';EndUTC='2026-10-16T00:00:01Z'}.GetEnumerator()){$r.($v.Key)=$v.Value}
    $r.MembersAtRootWait=[uint64[]]@();$r.MembersBeforeTermination=[uint64[]]@();$r.MembersAfterCleanup=[uint64[]]@();$r.MemberEvidence=@();$r.MembersDuringRun=@();$r.ShutdownMemberEvidence=@()
    $a=New-Object 'IsolatedFuzzProcess+AccountingQueryObservation'
    foreach($v in @{API='QueryInformationJobObject/Accounting';Succeeded=$true;ValuesRead=$true;StartUTC='2026-10-16T00:00:00.1Z';EndUTC='2026-10-16T00:00:00.2Z';StartTimestamp=1;EndTimestamp=2;Frequency=10}.GetEnumerator()){$a.($v.Key)=$v.Value};$r.InitialAccountingQuery=$a
    $m=New-Object 'IsolatedFuzzProcess+MemberListQueryObservation'
    foreach($v in @{API='QueryInformationJobObject/ProcessIdList';Succeeded=$true;HeaderRead=$true;StartUTC='2026-10-16T00:00:00.3Z';EndUTC='2026-10-16T00:00:00.4Z';StartTimestamp=3;EndTimestamp=4;Frequency=10;Capacity=16;BufferBytes=136}.GetEnumerator()){$m.($v.Key)=$v.Value};$r.InitialMemberQueries.Add($m)
    return $r
}
function Invoke-LeafOrdinaryTestCase([string]$Scenario){
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
    $f=New-LeafOrdinaryTestFiles (Join-Path $workspace ('.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/ordinary-unit-'+[guid]::NewGuid().ToString('N')))
    $calls=New-Object 'System.Collections.Generic.List[object]'
    $audit=@{Product=0;SDK=0;Artifact=0}
    $originalWrite=(Get-Command Write-LeafOrdinaryJSON).ScriptBlock
    $originalPlan=(Get-Command Assert-LeafOrdinaryCheckPlan).ScriptBlock
    $originalArtifacts=(Get-Command Assert-LeafOrdinaryFinalArtifacts).ScriptBlock
    function Assert-LeafOrdinaryFinalArtifacts($Plan,$Bindings){
        $audit.Artifact++
        if($Scenario -cin @('post-artifact','all-audits-fail','native-throw-all-audits-fail')){throw 'controlled artifact audit'}
        & $originalArtifacts $Plan $Bindings
    }
    function Get-EnvelopeProductManifest {
        $audit.Product++
        if($audit.Product -gt 1 -and $Scenario -cin @('post-product','all-audits-fail','native-throw-all-audits-fail')){throw 'controlled product audit'}
        [pscustomobject]@{Path='fixture.go';SHA256=('A'*64)}
    }
    function Assert-UninstrumentedEnvelopeToolchain {
        $audit.SDK++
        if($Scenario -ceq 'preflight' -or ($audit.SDK -gt 1 -and $Scenario -cin @('post-sdk','all-audits-fail','native-throw-all-audits-fail'))){throw 'controlled SDK audit'}
        'fixture SDK identity'
    }
    function Assert-LeafOrdinaryCheckPlan($Plan,$RequireNew){
        & $originalPlan $Plan $RequireNew
        if($calls.Count -eq 1 -and -not $RequireNew -and $Scenario -ceq 'binary-before-run'){[IO.File]::AppendAllText($Plan.Binary,'changed before runtime')}
        if($calls.Count -eq 1 -and -not $RequireNew -and $Scenario -ceq 'source-before-run'){$Plan.SourceBinding[0].SHA256='F'*64;& $originalPlan $Plan $RequireNew}
    }
    function Write-LeafOrdinaryJSON($Path,$Value){
        $name=[IO.Path]::GetFileName($Path)
        if(($Scenario -ceq 'native-throw-persist' -and $name -ceq 'launch-error.json') -or ($Scenario -ceq 'result-write' -and $name -ceq 'result.json') -or ($Scenario -ceq 'manifest-write' -and $name -ceq 'evidence-manifest.json') -or ($Scenario -ceq 'pending-write' -and $name -ceq 'completed.pending.json') -or ($Scenario -ceq 'stopped-write' -and $name -ceq 'stopped.json')){throw 'controlled persistence error'}
        & $originalWrite $Path $Value
        if($Scenario -ceq 'marker-publish' -and $name -ceq 'completed.pending.json'){New-Item -ItemType Directory -Path (Join-Path $f.Root 'completed.json')|Out-Null}
    }
    function Invoke-MatrixNative($Executable,$Arguments,$Environment,$WorkingDirectory,$Stdout,$Stderr,$LimitSeconds){
        $build=$calls.Count -eq 0
        $progress=Get-Content -LiteralPath (Join-Path $f.Root 'progress.json') -Raw|ConvertFrom-Json
        if($progress.Attempts.Count -ne ($calls.Count+1)){throw 'Attempt not durably recorded before call'}
        $want=if($build){@('test','-c','-x','-p=1','-count=1','-timeout=120s',('-o='+(Join-Path $f.Root 'package.test.exe')),'talenro.local/platform/internal/nodecontrol/nodeleafbinding')}else{@('-test.count=1','-test.timeout=120s','-test.v=true','-test.run=.')}
        if(($Arguments -join '|') -cne ($want -join '|') -or $LimitSeconds -ne $(if($build){600}else{150}) -or $WorkingDirectory -cne $(if($build){$workspace}else{Join-Path $workspace 'internal/nodecontrol/nodeleafbinding'}) -or $Executable -cne $(if($build){Join-Path $f.Toolchain 'bin/go.exe'}else{Join-Path $f.Root 'package.test.exe'})){throw 'Wrong native boundary fields'}
        $wantEnv=@{GOROOT=$f.Toolchain;GOMODCACHE=$f.ModuleCache;GOCACHE=(Join-Path $f.Root 'go-cache');TEMP=(Join-Path $f.Root 'tmp');TMP=(Join-Path $f.Root 'tmp');GOTMPDIR=(Join-Path $f.Root 'tmp');GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='on';GOWORK='off';GOFLAGS='-mod=readonly';GOOS='windows';GOARCH='amd64';TALENRO_FUZZ_DIAG_MODE='';TALENRO_FUZZ_FIXTURE='';GODEBUG=''}
        if($Environment.Count -ne 20){throw 'Wrong environment count'};foreach($k in $wantEnv.Keys){if($Environment[$k] -cne $wantEnv[$k]){throw 'Wrong environment value'}}
        $calls.Add([pscustomobject]@{Build=$build;Executable=$Executable;Arguments=$Arguments;Environment=$Environment;WorkingDirectory=$WorkingDirectory;Limit=$LimitSeconds})
        if($Scenario -like 'native-throw*'){throw 'controlled native exception'}
        $r=New-LeafOrdinaryNativeFixture
        $out="=== RUN   TestFixture`r`n--- PASS: TestFixture (0.00s)`r`nPASS";$err=''
        if($build){
            if($Scenario -cne 'missing-binary'){[IO.File]::WriteAllText((Join-Path $f.Root 'package.test.exe'),'fixture binary never executed')}
            $compiler=Join-Path $f.Toolchain 'pkg/tool/windows_amd64/compile.exe'
            $test=(Get-ChildItem -LiteralPath (Join-Path $workspace 'internal/nodecontrol/nodeleafbinding') -Filter '*_test.go' -File|Select-Object -First 1).FullName
            $err='"'+$compiler+'" -p talenro.local/platform/internal/nodecontrol/nodeleafbinding -pack "'+$test+'"'+"`r`n"+'"'+$compiler+'" -p context -pack "'+(Join-Path $f.Toolchain 'src/context/context.go')+'"'+"`r`n"+'"'+$compiler+'" -p internal/fuzz -pack "'+(Join-Path $f.Toolchain 'src/internal/fuzz/fuzz.go')+'"'
            if($Scenario -ceq 'missing-compile-evidence'){$err='missing'}
        }else{
            switch($Scenario){
                'no-tests' {$out='PASS'}
                'no-final-pass' {$out='--- PASS: TestFixture (0.00s)'}
                'fail-line' {$out="--- FAIL: TestBad (0.00s)`r`n"+$out}
                'diagnostic-stdout' {$out='TALENRO_FUZZ_DIAG injected'+"`r`n"+$out}
                'diagnostic-stderr' {$err='TALENRO_FUZZ_DIAG injected'}
                'skip-and-subtest' {$out="=== RUN   TestFixture`r`n=== RUN   TestFixture/child`r`n    --- PASS: TestFixture/child (0.00s)`r`n--- PASS: TestFixture (0.00s)`r`n=== RUN   TestSkipped`r`n--- SKIP: TestSkipped (0.00s)`r`nPASS"}
            }
        }
        $prefix=if($build){'build'}else{'runtime'}
        switch($Scenario){
            ($prefix+'-exit') {$r.ExitCode=7}
            ($prefix+'-timeout') {$r.TimedOut=$true}
            ($prefix+'-diagnostic') {$r.ShutdownDiagnosticOnly=$true}
            ($prefix+'-truncated') {$r.MemberEvidenceTruncated=$true}
            ($prefix+'-query') {$r.InitialAccountingQuery.Succeeded=$false}
            ($prefix+'-unknown') {$r.NaturalShutdownVerified=$false}
            'build-error87' {$r.NaturalShutdownVerified=$false;$r.ShutdownEvidenceComplete=$false}
            'build-schema' {$r.ShutdownSchema=2}
            'runtime-forced' {if(-not $build){$r.TerminationRequested=$true}}
            'stopped-write' {$r.ExitCode=7}
        }
        [IO.File]::WriteAllText($Stdout,$out);[IO.File]::WriteAllText($Stderr,$err)
        return $r
    }
    $result=$null;$errorText=$null
    try{$result=Invoke-LeafOrdinaryCheck $f.Root $f.Toolchain $f.ReferenceReceipt $f.ModuleCache}catch{$errorText=$_.Exception.Message}
    $saved=$null;if(Test-Path -LiteralPath (Join-Path $f.Root 'result.json') -PathType Leaf){$saved=Get-Content -LiteralPath (Join-Path $f.Root 'result.json') -Raw|ConvertFrom-Json}
    [pscustomobject]@{Root=$f.Root;NativeCalls=$calls.Count;Calls=@($calls.ToArray());Result=$result;Error=$errorText;Saved=$saved;AuditCalls=$audit}
}
