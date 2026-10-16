# Test-only factories. Importing this file creates no files or processes.
function Check-BuildOnly([bool]$Value,[string]$Name) {
    if(-not $Value){throw ('ASSERTION FAILED: '+$Name)}
    Write-Output ('PASS: '+$Name)
}
function New-BuildOnlyTestFiles([string]$FixtureRoot) {
    $tool=Join-Path $FixtureRoot 'uninstrumented-901/toolchain'
    $modules=Join-Path $FixtureRoot 'modules'
    foreach($dir in @((Join-Path $tool 'bin'),(Join-Path $modules 'github.com/google/uuid@v1.6.0'))){New-Item -ItemType Directory -Path $dir -Force -ErrorAction Stop|Out-Null}
    [IO.File]::WriteAllText((Join-Path $tool 'bin/go.exe'),'fake file; never executed')
    [IO.File]::WriteAllText((Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod'),'module fixture')
    $reference=Join-Path $FixtureRoot 'reference.json';[IO.File]::WriteAllText($reference,'{}')
    [pscustomobject]@{Root=(Join-Path $FixtureRoot 'leaf-stage-build-only-901');Toolchain=$tool;ReferenceReceipt=$reference;ModuleCache=$modules}
}
function New-BuildOnlyReportFixture {
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
function Quote-BuildOnlyPath([string]$Path){'"'+$Path.Replace(([string][char]92),([string][char]92)*2)+'"'}
function New-BuildOnlyPhase([long]$Start,[long]$End){
    $v=New-Object LifetimeApiEvidence;$v.Succeeded=$true;$v.Frequency=1000;$v.StartTimestamp=$Start;$v.EndTimestamp=$End
    $v.StartUTC=([DateTimeOffset]'2026-10-16T00:00:00Z').AddMilliseconds($Start).ToString('o');$v.EndUTC=([DateTimeOffset]'2026-10-16T00:00:00Z').AddMilliseconds($End).ToString('o');return $v
}
function Get-BuildOnlyBytesHash([byte[]]$Bytes){$h=[Security.Cryptography.SHA256]::Create();try{([BitConverter]::ToString($h.ComputeHash($Bytes))).Replace('-','')}finally{$h.Dispose()}}
function New-BuildOnlyTriggerReport($Native,[string]$Log){
    $raw=[Text.Encoding]::UTF8.GetBytes(($Log -split ([string][char]10),2)[0]+([string][char]10))
    $t=New-Object CompileStageTriggerReport;$t.State='Activated';$t.LineStart=0;$t.LineEnd=$raw.Length;$t.LineNumber=1;$t.BytesConsumed=$raw.Length;$t.PollCount=1
    $t.RawLineBase64=[Convert]::ToBase64String($raw);$t.LineSHA256=Get-BuildOnlyBytesHash $raw;$t.PrefixSHA256=$t.LineSHA256
    $t.CandidateObserved=New-BuildOnlyPhase 10 20;$t.Activation=New-BuildOnlyPhase 50 60;$t.RootCheck=New-Object CompileStageRootCheck
    $phase=New-BuildOnlyPhase 30 40;foreach($field in @('Succeeded','Frequency','StartTimestamp','EndTimestamp','StartUTC','EndUTC')){$t.RootCheck.$field=$phase.$field};$t.RootCheck.WaitResult=258
    $t.FirstRegistryQuery=New-Object LifetimeQueryEvidence;foreach($field in @('Succeeded','Error','Frequency','StartTimestamp','EndTimestamp','StartUTC','EndUTC','Assigned','Returned','Capacity','BufferBytes')){$t.FirstRegistryQuery.$field=$Native.LifetimeDiagnostics.FirstQuery.$field}
    return $t
}
function New-BuildOnlyNativeFixture([object]$Plan) {
    $native=New-BuildOnlyReportFixture
    $compiler=Join-Path $Plan.Toolchain 'pkg/tool/windows_amd64/compile.exe'
    $log=(Quote-BuildOnlyPath $compiler)+' -p talenro.local/platform/internal/nodecontrol/nodeleafbinding -pack '+(Quote-BuildOnlyPath (Join-Path $Plan.Trigger.SourceDirectory 'budget_test.go'))+([string][char]10)
    foreach($package in @('context','internal/fuzz')){
        $source=Join-Path $Plan.Toolchain ('src/'+$package+'/'+$(if($package -ceq 'context'){'context.go'}else{'fuzz.go'}))
        $log+=(Quote-BuildOnlyPath $compiler)+' -p '+$package+' -pack '+(Quote-BuildOnlyPath $source)+([string][char]10)
    }
    $native.CompileStageTrigger=New-BuildOnlyTriggerReport $native $log
    [pscustomobject]@{Native=$native;Log=$log}
}
function Invoke-BuildOnlyTestCase([string]$Scenario) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $unit=Join-Path $scope ('build-only unit '+[guid]::NewGuid().ToString('N'))
    $inputs=New-BuildOnlyTestFiles $unit
    $root=$inputs.Root;$tool=$inputs.Toolchain;$modules=$inputs.ModuleCache;$reference=$inputs.ReferenceReceipt
    $plan=Get-LeafStageBuildOnlyPlan $root $tool $reference $modules
    if($Scenario -ceq 'existing'){New-Item -ItemType Directory -Path $root|Out-Null}
    if($Scenario -ceq 'wrong-root'){$root=Join-Path $unit 'matrix-check-901'}
    $calls=New-Object 'System.Collections.Generic.List[object]'
    $audit=[pscustomobject]@{Tool=0;Product=0;Sources=0}
    $writer=${function:Write-RetainedChildObservationJSON}
    function Write-RetainedChildObservationJSON([string]$Path,$Value){
        if([IO.Path]::GetFileName($Path) -cin @('completed.json','completed.pending.json')){
            if($Scenario -ceq 'marker-create-failure'){$stream=[IO.File]::Open($Path,[IO.FileMode]::CreateNew);$stream.Dispose();throw 'controlled marker create failure'}
            if($Scenario -ceq 'marker-write-failure'){[IO.File]::WriteAllText($Path,'{"Finalized":true,');throw 'controlled marker write failure'}
            if($Scenario -ceq 'marker-close-failure'){& $writer $Path $Value;throw 'controlled marker close failure'}
        }
        & $writer $Path $Value
    }
    $verifier=${function:Assert-LeafStageBuildOnlySources}
    function Assert-LeafStageBuildOnlySources($Plan){
        $audit.Sources++
        if($Scenario -ceq 'multi-audit' -and $audit.Sources -eq 4){throw 'controlled final source failure'}
        & $verifier $Plan
    }
    function Get-EnvelopeProductManifest {
        $audit.Product++
        if($Scenario -ceq 'after-product-persist' -and $audit.Product -eq 2){New-Item -ItemType Directory -Path (Join-Path $root 'after-product.json')|Out-Null}
        [pscustomobject]@{Path='fixture.go';SHA256=$(if($Scenario -cin @('post-product','multi-audit') -and $audit.Product -eq 2){'C'*64}else{'A'*64})}
    }
    function Assert-UninstrumentedEnvelopeToolchain($Toolchain,$ReferenceReceipt){
        $audit.Tool++
        if($Scenario -ceq 'artifact-drift' -and $audit.Tool -eq 2){[IO.File]::WriteAllText((Join-Path $root 'build/stdout.txt'),'changed')}
        if($Scenario -ceq 'pre-tool' -or ($Scenario -cin @('post-tool','multi-audit') -and $audit.Tool -eq 2)){throw 'controlled tool failure'}
        ('A'*64)+':'+('B'*64)
    }
    function Invoke-LeafStageBuildOnlyNative($Plan,$Stdout,$Stderr){
        $calls.Add(($Plan|ConvertTo-Json -Depth 32|ConvertFrom-Json))
        $fixed=Get-LeafStageBuildOnlyFixedFields ([pscustomobject]@{RunRoot=$root;Toolchain=$tool;ModuleCache=$modules;Project=$workspace})
        if($Plan.Executable -cne (Join-Path $tool 'bin/go.exe') -or $Plan.Project -cne $workspace -or $Plan.BuildLimitSeconds -ne 600){throw 'Wrong native target/cwd/budget'}
        Assert-LeafStageBuildOnlyArguments $Plan.BuildArguments @('test','-c','-x','-p=1','-vet=off','-count=1','-timeout=120s',('-o='+(Join-Path $root 'package.test.exe')),'talenro.local/platform/internal/nodecontrol/nodeleafbinding')
        Assert-LeafStageBuildOnlyMap $Plan.Environment $fixed.Environment
        if($Scenario -cin @('native-throw','native-throw-persist')){
            if($Scenario -ceq 'native-throw-persist'){New-Item -ItemType Directory -Path (Join-Path $root 'build/launch-error.json')|Out-Null}
            throw 'controlled native exception'
        }
        $fixture=New-BuildOnlyNativeFixture $Plan;$n=$fixture.Native;$log=$fixture.Log
        switch($Scenario){
            exit{$n.ExitCode=7}
            primary-result-persist{$n.ExitCode=7}
            primary-stopped-persist{$n.ExitCode=7}
            multi-audit{$n.ExitCode=7}
            timeout{$n.TimedOut=$true}
            forced{$n.TerminationRequested=$true}
            core87{$n.ShutdownMemberEvidence[0].QueryError=87}
            missing-core{$n.ShutdownMemberEvidence=@()}
            missing-trigger{$n.CompileStageTrigger=$null}
            bad-prefix{$n.CompileStageTrigger.PrefixSHA256='C'*64}
            bad-clock{$n.CompileStageTrigger.Activation.StartTimestamp=400}
            cap32{
                $d=$n.LifetimeDiagnostics;$original=$d.Entries[0];$d.Entries.Clear();$d.Attempts=32;$d.Admitted=32;$d.ReleaseAttempts=32;$d.FirstQuery.Assigned=32;$d.FirstQuery.Returned=32
                foreach($i in 1..32){$e=New-Object LifetimeEntry;$e.EntryId=$i;$e.Pid=22+$i;$e.SourceQuery=1;foreach($field in @('SourceQueryEvidence','Open','Membership','Admission','RootWaitEnd','Final')){$e.$field=$original.$field};$d.Entries.Add($e)}
            }
            zero-entry{$n.LifetimeDiagnostics.Entries.Clear();$n.LifetimeDiagnostics.Attempts=0;$n.LifetimeDiagnostics.Admitted=0;$n.LifetimeDiagnostics.ReleaseAttempts=0}
            release-failure{$n.LifetimeDiagnostics.ReleaseFailures=1}
            scope-drift{$Plan.BuildArguments[4]='-vet=all'}
            compile-tokens{$log='missing compile tokens'}
            source-drift{[IO.File]::WriteAllText($Plan.Executable,'changed fake go')}
        }
        if($Scenario -cne 'missing-binary'){[IO.File]::WriteAllText($Plan.Binary,'fake binary; never executed')}
        [IO.File]::WriteAllText($Stdout,'');[IO.File]::WriteAllText($Stderr,$log)
        [IO.File]::WriteAllText((Join-Path $root 'go-cache/not-evidence'),'excluded')
        $obstruction=switch($Scenario){persist{'result.json'} primary-result-persist{'result.json'} primary-stopped-persist{'stopped.json'} manifest{'evidence-manifest.json'} marker{'completed.json'}}
        if($obstruction){New-Item -ItemType Directory -Path (Join-Path $root $obstruction)|Out-Null}
        return $n
    }
    $failure=$null;$output=@();try{$output=@(Invoke-LeafStageBuildOnly $root $tool $reference $modules)}catch{$failure=$_.Exception.Message}
    $saved=$null;if(Test-Path -LiteralPath (Join-Path $root 'result.json') -PathType Leaf){$saved=Get-Content -LiteralPath (Join-Path $root 'result.json') -Raw|ConvertFrom-Json}
    [pscustomobject]@{Root=$root;Plan=$plan;NativeCalls=$calls.Count;Invocation=@($calls.ToArray());ToolAuditCalls=$audit.Tool;ProductAuditCalls=$audit.Product;SourceAuditCalls=$audit.Sources;Failure=$failure;Saved=$saved;Output=$output}
}
