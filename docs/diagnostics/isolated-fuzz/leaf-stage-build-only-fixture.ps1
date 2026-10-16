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
