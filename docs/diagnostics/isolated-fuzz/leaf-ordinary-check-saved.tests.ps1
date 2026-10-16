$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'leaf-ordinary-check-fixture.ps1')
. (Join-Path $PSScriptRoot 'leaf-ordinary-check.ps1')
Check-LeafOrdinary ($null -ne (Get-Command Get-LeafOrdinarySavedEvidence -ErrorAction SilentlyContinue)) 'SavedReaderExists'
$f=Invoke-LeafOrdinaryTestCase 'success'
Check-LeafOrdinary ($null -eq $f.Error) 'SavedFixtureCompleted'
$root=$f.Root
function Invoke-MatrixNative {throw 'Saved reader launched workload'}
function Get-LeafOrdinaryCheckPlan {throw 'Saved reader called new-root planner'}
function Assert-UninstrumentedEnvelopeToolchain {throw 'Saved reader inventoried SDK'}
function Get-EnvelopeProductManifest {throw 'Saved reader inventoried product'}
$e=Get-LeafOrdinarySavedEvidence $root
Check-LeafOrdinary ($e.PackagePassed -is [bool] -and $e.PackagePassed -and $e.Acceptance -is [bool] -and -not $e.Acceptance) 'CompletedBindingPassesAndSavedReaderLaunchesZero'
$manifest=Get-Content -LiteralPath (Join-Path $root 'evidence-manifest.json') -Raw|ConvertFrom-Json
Check-LeafOrdinary ($manifest.Files.Count -eq 15) 'All15RequiredArtifacts'
$backup=@{}
foreach($name in @($manifest.Files.Path)+@('completed.json','evidence-manifest.json')){$backup[$name]=[IO.File]::ReadAllBytes((Join-Path $root $name))}
function Write-LeafTestValue([string]$Name,$Value){[IO.File]::WriteAllText((Join-Path $root $Name),($Value|ConvertTo-Json -Depth 30),[Text.UTF8Encoding]::new($false))}
function Update-LeafTestSeal {
    $bindings=Get-Content -LiteralPath (Join-Path $root 'artifact-binding.json') -Raw|ConvertFrom-Json
    foreach($binding in $bindings){foreach($row in $binding.Files){$row.SHA256=(Get-FileHash -LiteralPath $row.Path).Hash;$row.Length=(Get-Item -LiteralPath $row.Path).Length}}
    Write-LeafTestValue 'artifact-binding.json' @($bindings)
    $m=Get-Content -LiteralPath (Join-Path $root 'evidence-manifest.json') -Raw|ConvertFrom-Json
    foreach($row in $m.Files){if($row.Path -match '^[a-zA-Z0-9/.-]+$' -and -not $row.Path.Contains('..')){$file=Join-Path $root $row.Path;if(Test-Path -LiteralPath $file -PathType Leaf){$row.SHA256=(Get-FileHash -LiteralPath $file).Hash;$row.Length=(Get-Item -LiteralPath $file).Length}}}
    Write-LeafTestValue 'evidence-manifest.json' $m
    $c=Get-Content -LiteralPath (Join-Path $root 'completed.json') -Raw|ConvertFrom-Json
    $c.ManifestSHA256=(Get-FileHash -LiteralPath (Join-Path $root 'evidence-manifest.json')).Hash;Write-LeafTestValue 'completed.json' $c
}
$populated=Get-Content -LiteralPath (Join-Path $root 'build/receipt.json') -Raw|ConvertFrom-Json
$row=New-Object 'IsolatedFuzzProcess+MemberObservation';$row.Id=23;$row.Owned=$true;$row.ExitCodeRead=$true;$row.State='running';$row.ConfirmationAttempted=$true;$row.ConfirmationState='running'
$row.FinalShutdown=New-Object 'IsolatedFuzzProcess+ShutdownObservation'
foreach($entry in @{MemberId=23;SameHandleVerified=$true;Attempted=$true;State='exited';ExitCodeRead=$true;WaitBudgetMs=100;StartElapsedMs=0;EndElapsedMs=1}.GetEnumerator()){$row.FinalShutdown.($entry.Key)=$entry.Value}
$row.ProcessOpen=New-Object 'IsolatedFuzzProcess+MemberQueryObservation';$row.ProcessOpen.API='OpenProcess';$row.ProcessOpen.Succeeded=$true
$row.JobMembership=New-Object 'IsolatedFuzzProcess+MemberQueryObservation';$row.JobMembership.API='IsProcessInJob';$row.JobMembership.Succeeded=$true
$populated.MembersAtRootWait=@(23);$populated.ActiveAtRootWait=1;$populated.DescendantsRemaining=1;$populated.TotalProcesses=2
$populated.MemberEvidence=@($row);$populated.ShutdownMemberEvidence=@($row)
$populated.InitialAccountingQuery.ActiveProcesses=1;$populated.InitialAccountingQuery.TotalProcesses=2
$populated.InitialMemberQueries[0].Assigned=1;$populated.InitialMemberQueries[0].Returned=1
Write-LeafTestValue 'build/receipt.json' $populated;Update-LeafTestSeal
Check-LeafOrdinary ((Get-LeafOrdinarySavedEvidence $root).PackagePassed) 'PopulatedNestedReceiptPasses'
foreach($name in $backup.Keys.Clone()){ $backup[$name]=[IO.File]::ReadAllBytes((Join-Path $root $name)) }
foreach($case in @('ReceiptRootFlag','ReceiptCleanupFlag','ReceiptQueryFlag','ReceiptValuesFlag','LiveSampling','ScalarQueries','ScalarMembers','ScalarIDs','MemberOwnedFlag','FinalSameHandleFlag','NestedQueryFlag','ProvisionalRefused','PendingRefused','ContradictionRefused','ManifestHashMutation','ArtifactMutation','UnsafeRow','DuplicateRow','MissingRequiredRow','AlteredPlanAndReceipt','SourceDrift','InputDrift','TypedBooleanMutation','ResultBooleanMutation','QueryMutation','AttemptMutation','TestOutputMutation')){
    foreach($name in $backup.Keys){[IO.File]::WriteAllBytes((Join-Path $root $name),$backup[$name])}
    $extra=$null;$inputPath=$null;$inputBytes=$null
    switch($case){
        {$_ -cin @('ReceiptRootFlag','ReceiptCleanupFlag','ReceiptQueryFlag','ReceiptValuesFlag','LiveSampling','ScalarQueries','ScalarMembers','ScalarIDs','MemberOwnedFlag','FinalSameHandleFlag','NestedQueryFlag')} {
            $r=Get-Content -LiteralPath (Join-Path $root 'build/receipt.json') -Raw|ConvertFrom-Json
            switch($case){
                'ReceiptRootFlag' {$r.RootExited='false'}
                'ReceiptCleanupFlag' {$r.CleanupVerified='false'}
                'ReceiptQueryFlag' {$r.InitialAccountingQuery.Succeeded='false'}
                'ReceiptValuesFlag' {$r.InitialAccountingQuery.ValuesRead='false'}
                'LiveSampling' {$r.LiveSampling=$true}
                'ScalarQueries' {$r.InitialMemberQueries=$r.InitialMemberQueries[0]}
                'ScalarMembers' {$r.ShutdownMemberEvidence=$r.ShutdownMemberEvidence[0]}
                'ScalarIDs' {$r.MembersAtRootWait=23}
                'MemberOwnedFlag' {$r.ShutdownMemberEvidence[0].Owned='false'}
                'FinalSameHandleFlag' {$r.ShutdownMemberEvidence[0].FinalShutdown.SameHandleVerified='false'}
                'NestedQueryFlag' {$r.ShutdownMemberEvidence[0].ProcessOpen.Succeeded='false'}
            }
            Write-LeafTestValue 'build/receipt.json' $r;Update-LeafTestSeal
        }
        'ProvisionalRefused' {$c=Get-Content -LiteralPath (Join-Path $root 'completed.json') -Raw|ConvertFrom-Json;$c.PackagePassed=$false;Write-LeafTestValue 'completed.json' $c}
        'PendingRefused' {$extra=Join-Path $root 'completed.pending.json';[IO.File]::WriteAllText($extra,'{}')}
        'ContradictionRefused' {$extra=Join-Path $root 'stopped.json';[IO.File]::WriteAllText($extra,'{}')}
        'ManifestHashMutation' {[IO.File]::AppendAllText((Join-Path $root 'evidence-manifest.json'),' ')}
        'ArtifactMutation' {[IO.File]::AppendAllText((Join-Path $root 'package.test.exe'),'changed')}
        'UnsafeRow' {$m=Get-Content -LiteralPath (Join-Path $root 'evidence-manifest.json') -Raw|ConvertFrom-Json;$m.Files[0].Path='../escape';Write-LeafTestValue 'evidence-manifest.json' $m;Update-LeafTestSeal}
        'DuplicateRow' {$m=Get-Content -LiteralPath (Join-Path $root 'evidence-manifest.json') -Raw|ConvertFrom-Json;$m.Files[1]=$m.Files[0];Write-LeafTestValue 'evidence-manifest.json' $m;Update-LeafTestSeal}
        'MissingRequiredRow' {$m=Get-Content -LiteralPath (Join-Path $root 'evidence-manifest.json') -Raw|ConvertFrom-Json;$m.Files=@($m.Files|Select-Object -Skip 1);Write-LeafTestValue 'evidence-manifest.json' $m;Update-LeafTestSeal}
        'AlteredPlanAndReceipt' {$p=Get-Content -LiteralPath (Join-Path $root 'plan.json') -Raw|ConvertFrom-Json;$p.BuildArguments[1]='-vet=off';Write-LeafTestValue 'plan.json' $p;$r=Get-Content -LiteralPath (Join-Path $root 'build/receipt.json') -Raw|ConvertFrom-Json;$r.Arguments[1]='-vet=off';Write-LeafTestValue 'build/receipt.json' $r;Update-LeafTestSeal}
        'SourceDrift' {$p=Get-Content -LiteralPath (Join-Path $root 'plan.json') -Raw|ConvertFrom-Json;$p.SourceBinding[0].SHA256='F'*64;Write-LeafTestValue 'plan.json' $p;Write-LeafTestValue 'source-binding.json' $p.SourceBinding;Update-LeafTestSeal}
        'InputDrift' {$p=Get-Content -LiteralPath (Join-Path $root 'plan.json') -Raw|ConvertFrom-Json;$inputPath=$p.InputBinding[3].Path;$inputBytes=[IO.File]::ReadAllBytes($inputPath);[IO.File]::AppendAllText($inputPath,'changed')}
        'TypedBooleanMutation' {$c=Get-Content -LiteralPath (Join-Path $root 'completed.json') -Raw|ConvertFrom-Json;$c.Acceptance='false';Write-LeafTestValue 'completed.json' $c}
        'ResultBooleanMutation' {$r=Get-Content -LiteralPath (Join-Path $root 'result.json') -Raw|ConvertFrom-Json;$r.CompilePassed='true';Write-LeafTestValue 'result.json' $r;Update-LeafTestSeal}
        'QueryMutation' {$r=Get-Content -LiteralPath (Join-Path $root 'build/receipt.json') -Raw|ConvertFrom-Json;$r.InitialAccountingQuery.Succeeded=$false;Write-LeafTestValue 'build/receipt.json' $r;Update-LeafTestSeal}
        'AttemptMutation' {Write-LeafTestValue 'progress.json' @{Attempts=@()};Update-LeafTestSeal}
        'TestOutputMutation' {[IO.File]::WriteAllText((Join-Path $root 'execute/stdout.txt'),'PASS');Update-LeafTestSeal}
    }
    Reject-LeafOrdinary {Get-LeafOrdinarySavedEvidence $root} $case
    if($extra){[IO.File]::Delete($extra)}
    if($inputPath){[IO.File]::WriteAllBytes($inputPath,$inputBytes)}
}
foreach($name in $backup.Keys){[IO.File]::WriteAllBytes((Join-Path $root $name),$backup[$name])}
$r=Get-Content -LiteralPath (Join-Path $root 'build/receipt.json') -Raw|ConvertFrom-Json
Check-LeafOrdinary ($r.MembersAtRootWait -is [array] -and $r.InitialMemberQueries -is [array] -and $r.CleanupVerified -is [bool] -and $r.InitialMemberQueries.Count -eq 1) 'PS51Roundtrip'
Check-LeafOrdinary ((Read-LeafOrdinaryJSON (Join-Path $root 'build/receipt.json')).StartUTC -is [string]) 'NativeTimestampRemainsJSONString'
Check-LeafOrdinary ((Get-LeafOrdinarySavedEvidence $root).PackagePassed) 'RestoredFixturePasses'
