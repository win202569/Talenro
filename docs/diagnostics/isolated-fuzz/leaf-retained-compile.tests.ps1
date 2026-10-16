param([ValidateSet('all','cap','audits','identity')][string]$ReviewFocus='all')
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'mapping-single-check.ps1')
$helper=Join-Path $PSScriptRoot 'leaf-retained-compile.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
function Check-Leaf([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
Check-Leaf ($null -ne (Get-Command Invoke-LeafRetainedCompile -ErrorAction SilentlyContinue)) 'fixed leaf retained compile entry exists'
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
function New-LeafNativeFixture {
    $n=New-Object 'IsolatedFuzzProcess+Result'
    foreach($v in @{RootId=17;RootExited=$true;LifetimeDiagnosticOnly=$true;ShutdownDiagnosticOnly=$true;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;CleanupVerified=$true;TotalProcesses=2;ActiveAtRootWait=1;DescendantsRemaining=1;NaturalObservationMs=200;ShutdownElapsedMs=210;StartUTC='2026-10-16T00:00:00Z';EndUTC='2026-10-16T00:00:01Z';MembersAtRootWait=[uint64[]]@(23);MembersBeforeTermination=[uint64[]]@();MembersAfterCleanup=[uint64[]]@()}.GetEnumerator()){$n.($v.Key)=$v.Value}
    $row=New-Object 'IsolatedFuzzProcess+MemberObservation';$row.Id=23;$row.Owned=$true;$row.ExitCodeRead=$true;$row.State='exited'
    $n.MemberEvidence=@($row);$n.ShutdownMemberEvidence=@($row)
    $a=New-Object 'IsolatedFuzzProcess+AccountingQueryObservation'
    foreach($v in @{API='QueryInformationJobObject/Accounting';Succeeded=$true;ValuesRead=$true;StartTimestamp=320;EndTimestamp=330;Frequency=1000;StartUTC='2026-10-16T00:00:00.320Z';EndUTC='2026-10-16T00:00:00.330Z'}.GetEnumerator()){$a.($v.Key)=$v.Value};$n.InitialAccountingQuery=$a
    $m=New-Object 'IsolatedFuzzProcess+MemberListQueryObservation'
    foreach($v in @{API='QueryInformationJobObject/ProcessIdList';Succeeded=$true;HeaderRead=$true;Attempt=1;Assigned=1;Returned=1;Capacity=16;BufferBytes=136;StartTimestamp=340;EndTimestamp=350;Frequency=1000;StartUTC='2026-10-16T00:00:00.340Z';EndUTC='2026-10-16T00:00:00.350Z'}.GetEnumerator()){$m.($v.Key)=$v.Value};$n.InitialMemberQueries.Add($m)
    $d=New-Object LifetimeReport;$n.LifetimeDiagnostics=$d
    $d.Attempts=1;$d.Admitted=1;$d.QueryCount=1;$d.RootExited=$true;$d.TeardownAttempted=$true;$d.ReleaseAttempts=1
    $e=New-Object LifetimeEntry;$e.EntryId=1;$e.Pid=23;$e.SourceQuery=1
    $q=New-Object LifetimeQueryEvidence
    foreach($v in @{Succeeded=$true;Assigned=2;Returned=2;Capacity=32;BufferBytes=264;StartTimestamp=100;EndTimestamp=110;Frequency=1000;StartUTC='2026-10-16T00:00:00.100Z';EndUTC='2026-10-16T00:00:00.110Z'}.GetEnumerator()){$q.($v.Key)=$v.Value}
    $e.SourceQueryEvidence=$q;$d.FirstQuery=$q;$d.LastQuery=$q
    $e.Open=New-Object LifetimeApiEvidence;$e.Open.Succeeded=$true;$e.Open.StartTimestamp=120;$e.Open.EndTimestamp=130;$e.Open.Frequency=1000;$e.Open.StartUTC='2026-10-16T00:00:00.120Z';$e.Open.EndUTC='2026-10-16T00:00:00.130Z'
    $e.Membership=New-Object LifetimeMembership;$e.Membership.Succeeded=$true;$e.Membership.Owned=$true;$e.Membership.StartTimestamp=140;$e.Membership.EndTimestamp=150;$e.Membership.Frequency=1000;$e.Membership.StartUTC='2026-10-16T00:00:00.140Z';$e.Membership.EndUTC='2026-10-16T00:00:00.150Z'
    foreach($phase in @('Admission','RootWaitEnd','Final')){
        $r=New-Object LifetimeRead;$r.Attempted=$true;$r.SameEntry=$true;$r.ExitCodeRead=$true;$r.Frequency=1000
        $r.StartTimestamp=switch($phase){Admission{160} RootWaitEnd{300} Final{500}};$r.EndTimestamp=$r.StartTimestamp+10
        $r.State=if($phase -ceq 'Final'){'exited'}else{'running'};$r.ExitCode=if($phase -ceq 'Final'){0}else{259}
        $r.StartUTC=([DateTimeOffset]'2026-10-16T00:00:00Z').AddMilliseconds($r.StartTimestamp).ToString('o');$r.EndUTC=([DateTimeOffset]'2026-10-16T00:00:00Z').AddMilliseconds($r.EndTimestamp).ToString('o');$e.$phase=$r
    }
    $d.Entries.Add($e);return $n
}
function Test-LeafCase([string]$Scenario){
    $unit=Join-Path $scope ('leaf-retained-unit-'+[guid]::NewGuid().ToString('N'))
    $root=Join-Path $unit 'leaf-retained-compile-901';$tool=Join-Path $unit 'uninstrumented-901/toolchain';$modules=Join-Path $unit 'modules';$reference=Join-Path $unit 'reference.json'
    foreach($dir in @((Join-Path $tool 'bin'),(Join-Path $modules 'github.com/google/uuid@v1.6.0'))){New-Item -ItemType Directory -Path $dir -Force|Out-Null}
    [IO.File]::WriteAllText((Join-Path $tool 'bin/go.exe'),'never executed');[IO.File]::WriteAllText((Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod'),'module fixture');[IO.File]::WriteAllText($reference,'{}')
    $plan=Get-LeafRetainedCompilePlan $root $tool $reference $modules
    if($Scenario -ceq 'plan'){
        Check-Leaf (-not(Test-Path $root) -and $plan.Package -ceq 'internal/nodecontrol/nodeleafbinding' -and $plan.BuildLimitSeconds -eq 600 -and -not $plan.RuntimeAttempted -and -not $plan.Acceptance) 'planning fixes compile only target and creates nothing'
        return
    }
    if($Scenario -ceq 'existing'){New-Item -ItemType Directory -Path $root|Out-Null}
    if($Scenario -ceq 'wrong-root'){$root=Join-Path $unit 'matrix-check-901'}
    $calls=New-Object 'System.Collections.Generic.List[object]';$audit=[pscustomobject]@{Tool=0;Product=0;Sources=0}
    $sourceVerifier=${function:Assert-LeafRetainedCompileSources}
    function Assert-LeafRetainedCompileSources($Plan){$audit.Sources++;& $sourceVerifier $Plan}
    # Replace external full SDK inventory and full product walk with tiny data;
    # real planner, source pins, manifest comparison, parsers and filesystem run.
    function Get-EnvelopeProductManifest {
        $audit.Product++
        if($Scenario -ceq 'after-product-persist' -and $audit.Product -gt 1){New-Item -ItemType Directory -Path (Join-Path $root 'after-product.json')|Out-Null}
        [pscustomobject]@{Path='fixture.go';SHA256=$(if($Scenario -cin @('post-product','multi-audit') -and $audit.Product -gt 1){'C'*64}else{'A'*64})}
    }
    function Assert-UninstrumentedEnvelopeToolchain($Toolchain,$ReferenceReceipt){
        $audit.Tool++
        if($audit.Tool -gt 1 -and $Scenario -cin @('artifact-drift','multi-audit')){[IO.File]::WriteAllText((Join-Path $root 'build/stdout.txt'),'changed after binding')}
        if($Scenario -ceq 'pre-tool' -or ($Scenario -cin @('post-tool','multi-audit') -and $audit.Tool -gt 1)){throw 'controlled tool drift'}
        return (('A'*64)+':'+('B'*64))
    }
    function Invoke-LeafRetainedCompileNative($Plan,$Stdout,$Stderr){
        $calls.Add($Plan)
        $expected=@('test','-c','-x','-p=1','-count=1','-timeout=120s',('-o='+(Join-Path $root 'package.test.exe')),'talenro.local/platform/internal/nodecontrol/nodeleafbinding')
        if($Plan.Executable -cne (Join-Path $tool 'bin/go.exe') -or $Plan.Project -cne $workspace -or $Plan.BuildLimitSeconds -ne 600 -or ($Plan.BuildArguments -join '|') -cne ($expected -join '|')){throw 'Unauthorized native target/cwd/arguments/budget'}
        foreach($v in @{GOROOT=$tool;GOMODCACHE=$modules;GOCACHE=(Join-Path $root 'go-cache');TEMP=(Join-Path $root 'tmp');TMP=(Join-Path $root 'tmp');GOTMPDIR=(Join-Path $root 'tmp');GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='on';GOWORK='off';GOFLAGS='-mod=readonly';GOOS='windows';GOARCH='amd64';TALENRO_FUZZ_DIAG_MODE='';TALENRO_FUZZ_FIXTURE='';GODEBUG=''}.GetEnumerator()){if($Plan.Environment[$v.Key] -cne $v.Value){throw ('Wrong offline environment '+$v.Key)}}
        if($Scenario -ceq 'native-throw'){throw 'controlled native exception'}
        $n=New-LeafNativeFixture
        switch($Scenario){
            exit{$n.ExitCode=7}
            timeout{$n.TimedOut=$true}
            forced{$n.TerminationRequested=$true}
            core87{$n.ShutdownMemberEvidence[0].QueryError=87}
            core-missing{$n.ShutdownMemberEvidence=@()}
            final-active{$n.ActiveAfterCleanup=1}
            final-list{$n.MembersAfterCleanup=[uint64[]]@(23)}
            flags{$n.LifetimeDiagnosticOnly=$false}
            cap{$n.LifetimeDiagnostics.AdmissionBudgetExhausted=$true}
            counter-cap{
                $d=$n.LifetimeDiagnostics;$original=$d.Entries[0];$d.Entries.Clear();$d.Attempts=32;$d.Admitted=32;$d.ReleaseAttempts=32
                $d.FirstQuery.Assigned=32;$d.FirstQuery.Returned=32
                foreach($i in 1..32){$e=New-Object LifetimeEntry;$e.EntryId=$i;$e.Pid=22+$i;$e.SourceQuery=1;foreach($field in @('SourceQueryEvidence','Open','Membership','Admission','RootWaitEnd','Final')){$e.$field=$original.$field};$d.Entries.Add($e)}
            }
            incomplete{$n.LifetimeDiagnostics.Incomplete=$true}
            empty{$n.LifetimeDiagnostics.Entries.Clear();$n.LifetimeDiagnostics.Attempts=0;$n.LifetimeDiagnostics.Admitted=0;$n.LifetimeDiagnostics.ReleaseAttempts=0}
            release{$n.LifetimeDiagnostics.ReleaseFailures=1}
            phase{$n.LifetimeDiagnostics.Entries[0].RootWaitEnd=$null}
            clock{$n.LifetimeDiagnostics.Entries[0].Final.StartTimestamp=10}
            ambiguous{$n.LifetimeDiagnostics.Entries[0].Final.Ambiguous=$true}
            contradictory{$n.LifetimeDiagnostics.Entries[0].Admission.State='exited';$n.LifetimeDiagnostics.Entries[0].Admission.ExitCode=0}
            missing-query{$n.InitialMemberQueries.Clear()}
            multi-audit{$n.ExitCode=7}
        }
        if($Scenario -cne 'missing-binary'){[IO.File]::WriteAllText($Plan.Binary,'never executed')}
        $compiler=Join-Path $tool 'pkg/tool/windows_amd64/compile.exe'
        $test=Join-Path $workspace 'internal/nodecontrol/nodeleafbinding/boundary_test.go'
        $log='"'+$compiler+'" -p talenro.local/platform/internal/nodecontrol/nodeleafbinding -pack "'+$test+'"'+"`r`n"+'"'+$compiler+'" -p context -pack "'+(Join-Path $tool 'src/context/context.go')+'"'+"`r`n"+'"'+$compiler+'" -p internal/fuzz -pack "'+(Join-Path $tool 'src/internal/fuzz/fuzz.go')+'"'
        if($Scenario -ceq 'compile-log'){$log='missing copied compile evidence'}
        [IO.File]::WriteAllText($Stdout,'');[IO.File]::WriteAllText($Stderr,$log)
        [IO.File]::WriteAllText((Join-Path $root 'go-cache/not-evidence'),'cache fixture')
        if($Scenario -ceq 'persist'){New-Item -ItemType Directory -Path (Join-Path $root 'result.json')|Out-Null}
        if($Scenario -ceq 'manifest'){New-Item -ItemType Directory -Path (Join-Path $root 'evidence-manifest.json')|Out-Null}
        return $n
    }
    $failure=$null;$seen=@();try{$seen=@(Invoke-LeafRetainedCompile $root $tool $reference $modules)}catch{$failure=$_.Exception.Message}
    $saved=$null;if(Test-Path (Join-Path $root 'result.json') -PathType Leaf){$saved=Get-Content (Join-Path $root 'result.json') -Raw|ConvertFrom-Json}
    [pscustomobject]@{Root=$root;Calls=$calls.Count;ToolCalls=$audit.Tool;SourceCalls=$audit.Sources;Failure=$failure;Seen=$seen;Saved=$saved}
}
if($ReviewFocus -cne 'all'){
    switch($ReviewFocus){
        cap{$bad=Test-LeafCase 'counter-cap';Check-Leaf ($null -ne $bad.Failure -and -not $bad.Saved.ObservationPassed) 'counter32 with cleared cap flag is still incomplete'}
        audits{$bad=Test-LeafCase 'post-product';Check-Leaf ($bad.ToolCalls -eq 2 -and $bad.SourceCalls -eq 3 -and $bad.Saved.ArtifactsUnchanged) 'product failure does not skip remaining final audits'}
        identity{$good=Test-LeafCase 'success';$property=$good.Saved.PSObject.Properties['ToolchainReference'];Check-Leaf ($null -ne $property -and $property.Value -ceq (('A'*64)+':'+('B'*64))) 'verified SDK receipt identity persists in bound result'}
    }
    return
}
Test-LeafCase 'plan'
$good=Test-LeafCase 'success'
Check-Leaf ($null -eq $good.Failure -and $good.Calls -eq 1 -and $good.Seen.Count -eq 1 -and $good.Saved.ObservationPassed -and $good.Saved.CompileEvidenceComplete -and -not $good.Saved.PackagePassed -and -not $good.Saved.Acceptance -and -not $good.Saved.RuntimeAttempted) 'one compile only yields sampled diagnostic completion not package acceptance'
$property=$good.Saved.PSObject.Properties['ToolchainReference'];Check-Leaf ($null -ne $property -and $property.Value -ceq (('A'*64)+':'+('B'*64))) 'verified SDK receipt identity persists in bound result'
$receipt=Get-Content (Join-Path $good.Root 'build/receipt.json') -Raw|ConvertFrom-Json
Check-Leaf ($receipt.LifetimeDiagnosticOnly -and $receipt.ShutdownDiagnosticOnly -and $receipt.DescendantsRemaining -eq 1 -and $receipt.LifetimeDiagnostics.Entries.Count -eq 1) 'receipt preserves both diagnostic flags nested metadata and root wait count'
$rejected=$false;try{Assert-ProcessCompletion $receipt}catch{$rejected=$true};Check-Leaf $rejected 'ordinary acceptance still refuses diagnostic compile'
$manifestPath=Join-Path $good.Root 'evidence-manifest.json';$manifest=Get-Content $manifestPath -Raw|ConvertFrom-Json
$hashes=$true;foreach($row in $manifest.Files){$file=Join-Path $good.Root $row.Path;if((Get-FileHash $file).Hash -cne $row.SHA256 -or (Get-Item $file).Length -ne $row.Length){$hashes=$false}}
$completed=Get-Content (Join-Path $good.Root 'completed.json') -Raw|ConvertFrom-Json
Check-Leaf ($hashes -and $completed.ManifestSHA256 -ceq (Get-FileHash $manifestPath).Hash -and $completed.Finalized -and -not $completed.Acceptance -and @($manifest.Files|Where-Object {$_.Path -match '^(go-cache|tmp)/'}).Count -eq 0) 'explicit artifact manifest and completion marker exclude cache and temporary trees'
foreach($case in @(@('existing',0),@('wrong-root',0),@('pre-tool',0),@('native-throw',1),@('exit',1),@('timeout',1),@('forced',1),@('core87',1),@('core-missing',1),@('final-active',1),@('final-list',1),@('flags',1),@('cap',1),@('counter-cap',1),@('incomplete',1),@('empty',1),@('release',1),@('phase',1),@('clock',1),@('ambiguous',1),@('contradictory',1),@('missing-query',1),@('missing-binary',1),@('compile-log',1),@('post-product',1),@('after-product-persist',1),@('post-tool',1),@('multi-audit',1),@('artifact-drift',1),@('persist',1),@('manifest',1))){
    $bad=Test-LeafCase $case[0]
    Check-Leaf ($null -ne $bad.Failure -and $bad.Calls -eq $case[1] -and $bad.Seen.Count -eq 0) ($case[0]+' stops at first failure without retry or success output')
    if($case[0] -cnotin @('existing','wrong-root')){
        Check-Leaf ((Test-Path (Join-Path $bad.Root 'stopped.json') -PathType Leaf) -and -not (Test-Path (Join-Path $bad.Root 'completed.json'))) ($case[0]+' has stopped marker and never durable completion')
        if($null -ne $bad.Saved){
            if($case[0] -ceq 'manifest'){Check-Leaf ($bad.Saved.FinalizationRequired -and -not $bad.Saved.Acceptance) 'provisional result cannot override failed finalization'}
            else{Check-Leaf (-not $bad.Saved.ObservationPassed -and -not $bad.Saved.Acceptance -and -not $bad.Saved.RuntimeAttempted) ($case[0]+' preserves unaccepted incomplete result')}
        }
        if($case[0] -cin @('post-product','after-product-persist','post-tool')){Check-Leaf ($bad.ToolCalls -eq 2 -and $bad.SourceCalls -eq 3 -and $bad.Saved.ArtifactsUnchanged) ($case[0]+' attempts every applicable final audit')}
        if($case[0] -ceq 'multi-audit'){Check-Leaf ($bad.ToolCalls -eq 2 -and $bad.SourceCalls -eq 3 -and $bad.Saved.AuditErrors.Count -eq 3 -and $bad.Saved.Failure -match 'exit nonzero') 'multiple audit failures preserve original compile failure and all audit errors'}
    }
}
Write-Output 'LIMIT: fake SDK/product inventories and native boundary; no Go, child, live PID query or matrix execution.'
