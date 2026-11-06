# Dot-sourced by the claimed phase harness. Real files, no workload processes.
function New-RecordFixture($content='abc'){
 $source=New-Fixture $null $content
 $invocation=Join-Path $source 'invocation';$null=[IO.Directory]::CreateDirectory($invocation)
 $session=New-C12ShadowSession $invocation -Synthetic $true
 Check ($null -eq $session.FirstReason) 'session creation'
 $identity=@{GroupID='G01';Profile='base';Package='./internal/testinfra';RunSuffix='r001';CandidateTree=$null}
 $group=Start-C12ShadowGroup $session $source $identity ([datetime]::UtcNow.AddMinutes(5))
 return @{Source=$source;Session=$session;Group=$group}
}
function Finish-Group($f){Complete-C12ShadowGroup $f.Group $false $false $true ([datetime]::UtcNow.AddMinutes(5))}
function Finish-Session($f){Complete-C12ShadowSession $f.Session 0 $false $true ([datetime]::UtcNow.AddMinutes(5))}
function Read-Record($session,$reference){Get-Content -LiteralPath (Join-Path $session.EvidenceRoot $reference.Path) -Raw|ConvertFrom-Json -AsHashtable}
# Missing final publication must not become a match; persisted references must bind bytes.
$current='records-matched';$f=New-RecordFixture
$p=Finish-Group $f
Check ($p.State -ceq 'PROVISIONAL' -and -not(Test-Path (Join-Path $f.Session.EvidenceRoot 'run.json'))) 'premature final match'
$r=Finish-Session $f;Check ($r.State -ceq 'MATCHED' -and $null -ne $r.Reference) 'final matched'
$saved=Read-Record $f.Session $r.Reference
Check ($saved.Synthetic -is [bool] -and $saved.Synthetic -and $saved.FormalExitCode -eq 0 -and $saved.State -ceq 'MATCHED') 'persisted outcome'
Record-Pass $current
# Restoring contents after a drifting before-map cannot erase drift.
$current='records-before-drift';$f=New-RecordFixture 'changed';[IO.File]::WriteAllText((Join-Path $f.Source 'a.txt'),'abc',$utf8)
$null=Finish-Group $f;$r=Finish-Session $f;Check ($r.State -ceq 'DRIFT') 'before drift erased';Record-Pass $current
# Identity mutation must not publish a match.
$current='records-binding'
foreach($mutation in @('identity','audit','reference')){
 $f=New-RecordFixture
 if($mutation -eq 'identity'){$f.Group.Identity.RunSuffix='wrong'}
 if($mutation -eq 'audit'){$f.Session.AuditId='00000000000000000000000000000000'}
 if($mutation -eq 'reference'){
  $other=Start-C12ShadowGroup $f.Session $f.Source @{GroupID='other';Profile='base';Package='p';RunSuffix='r2';CandidateTree=$null} ([datetime]::UtcNow.AddMinutes(5))
  $f.Group.Before=$other.Before
 }
 $null=Finish-Group $f;$r=Finish-Session $f
 Check ($r.State -ceq 'INCOMPLETE' -and $r.Reason -ceq 'IDENTITY') "wrong $mutation binding accepted"
}
Record-Pass $current
# Byte changes to a reference invalidate it even with valid JSON.
$current='records-tampered';$f=New-RecordFixture;$null=Finish-Group $f
[IO.File]::AppendAllText((Join-Path $f.Session.EvidenceRoot $f.Group.Before.Path),' ',$utf8)
$r=Finish-Session $f;Check ($r.State -ceq 'INCOMPLETE') 'changed reference accepted';Record-Pass $current
# Existing summary bytes cannot be overwritten, including a throwing warning sink.
$current='records-occupied-summary';$f=New-RecordFixture;$null=Finish-Group $f
$occupied=Join-Path $f.Session.EvidenceRoot 'run.json';Bytes-Test $occupied ($utf8.GetBytes('sentinel'))
function Write-Warning {param($Message) throw 'marker-password-warning'}
try{$r=Finish-Session $f}finally{Remove-Item Function:\Write-Warning}
Check ($r.State -ceq 'INCOMPLETE' -and $null -eq $r.Reference -and [IO.File]::ReadAllText($occupied) -ceq 'sentinel') 'occupied summary changed';Record-Pass $current
$current='records-occupied-group';$f=New-RecordFixture
$occupied=Join-Path $f.Session.EvidenceRoot ($f.Group.Leaf+'.group.json');Bytes-Test $occupied ($utf8.GetBytes('sentinel'))
$p=Finish-Group $f;$r=Finish-Session $f
Check ($p.State -ceq 'INCOMPLETE' -and $r.State -ceq 'INCOMPLETE' -and [IO.File]::ReadAllText($occupied) -ceq 'sentinel') 'occupied group changed';Record-Pass $current
# Locks make accidental after-scanning observable: expected reason must remain the gate.
foreach($gate in @('cleanup','ownership','deadline')){
 $current='records-'+$gate;$f=New-RecordFixture
 $lock=[IO.File]::Open((Join-Path $f.Source 'a.txt'),[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::None)
 try{
  $deadline=[datetime]::UtcNow.AddMinutes(5);if($gate -eq 'deadline'){$deadline=[datetime]::UtcNow.AddSeconds(-1)}
  $p=Complete-C12ShadowGroup $f.Group $false ($gate -eq 'cleanup') ($gate -ne 'ownership') $deadline
 }finally{$lock.Dispose()}
 $r=Finish-Session $f
 Check ($r.State -ceq 'INCOMPLETE' -and $r.Reason -ceq $gate.ToUpperInvariant() -and $null -eq $f.Group.After) "gate $gate"
 Record-Pass $current
}
$current='records-no-groups';$source=New-Fixture;$session=New-C12ShadowSession $source -Synthetic $true
$r=Complete-C12ShadowSession $session 0 $false $true ([datetime]::UtcNow.AddMinutes(5))
Check ($r.State -ceq 'INCOMPLETE' -and $r.Reason -ceq 'NO_GROUPS') 'empty session matched';Record-Pass $current
$current='records-unreadable-manifest';$source=New-Fixture;$session=New-C12ShadowSession $source -Synthetic $true
$lock=[IO.File]::Open((Join-Path $source 'testdata/c12/shadow-source-manifest.v1.json'),[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::None)
try{$g=Start-C12ShadowGroup $session $source @{GroupID='g';Profile='base';Package='p';RunSuffix='r';CandidateTree=$null} ([datetime]::UtcNow.AddMinutes(5))}finally{$lock.Dispose()}
$null=Complete-C12ShadowGroup $g $false $false $true ([datetime]::UtcNow.AddMinutes(5))
$r=Complete-C12ShadowSession $session 0 $false $true ([datetime]::UtcNow.AddMinutes(5))
Check ($r.State -ceq 'INCOMPLETE' -and $r.Reason -ceq 'IO') 'unreadable manifest';Record-Pass $current
$current='records-missing-summary';$f=New-RecordFixture
$r=Finish-Session $f;Check ($r.State -ceq 'INCOMPLETE') 'unfinished group matched';Record-Pass $current
$current='records-outer-cleanup';$f=New-RecordFixture;$null=Finish-Group $f
$r=Complete-C12ShadowSession $f.Session 1 $true $true ([datetime]::UtcNow.AddMinutes(5))
Check ($r.State -ceq 'INCOMPLETE' -and $r.Reason -ceq 'CLEANUP') 'outer cleanup ignored'
$f=New-RecordFixture;$null=Finish-Group $f
$r=Complete-C12ShadowSession $f.Session 1 $false $false ([datetime]::UtcNow.AddMinutes(5))
Check ($r.State -ceq 'INCOMPLETE' -and $r.Reason -ceq 'CLEANUP') 'restoration ignored'
$f=New-RecordFixture;$null=Finish-Group $f
$r=Complete-C12ShadowSession $f.Session 0 $false $true ([datetime]::UtcNow.AddSeconds(-1))
Check ($r.State -ceq 'INCOMPLETE' -and $r.Reason -ceq 'DEADLINE' -and $null -eq $r.Reference -and -not(Test-Path (Join-Path $f.Session.EvidenceRoot 'run.json'))) 'expired final publication'
Record-Pass $current
$current='records-privacy';$source=New-Fixture;$session=New-C12ShadowSession $source -Synthetic $true
$g=Start-C12ShadowGroup $session $source @{GroupID='g';Profile='base';Package='p';RunSuffix='r';CandidateTree=$null;Password='marker-password-secret'} ([datetime]::UtcNow.AddMinutes(5))
$null=Complete-C12ShadowGroup $g $true $false $true ([datetime]::UtcNow.AddMinutes(5))
$r=Complete-C12ShadowSession $session 1 $false $true ([datetime]::UtcNow.AddMinutes(5))
Check ($r.State -ceq 'INCOMPLETE') 'extra identity key accepted'
foreach($file in Get-ChildItem -LiteralPath $session.EvidenceRoot -File){Check (-not([IO.File]::ReadAllText($file.FullName).Contains('marker-password'))) 'secret persisted'}
Record-Pass $current
