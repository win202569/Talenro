param([Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$InputSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
$d='.superpowers/sdd/2026-11-05-runtime-staged-integrity-v1'
$review='.superpowers/sdd/2026-11-05-runtime-staged-integrity-review-v2'
function Need([bool]$Ok,[string]$Why){if(-not $Ok){throw "BUSINESS_VERIFY: $Why"}}
function ReadJson([string]$Path){Get-Content -LiteralPath $Path -Raw|ConvertFrom-Json -AsHashtable}
function Save-SIReview([string]$Name,$Value){
 $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 30))
 if($bytes.Length -gt 65536){throw 'SI review bound'}
 $stream=[IO.File]::Open((Join-Path $review $Name),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
 try{$stream.Write($bytes,0,$bytes.Length);$stream.Flush()}finally{$stream.Dispose()}
}
try{
Need ((Get-Location).Path -ceq $root) 'root'
Need ((Get-FileHash "$d/input.json").Hash -ceq $InputSHA256) 'input identity'
$b=ReadJson "$d/input.json";$all=@{};$checks=0
foreach($map in @($b.SourceSHA256,$b.PreservedSHA256,$b.HistoricalEvidenceSHA256,$b.HelperSHA256)){
 foreach($p in $map.Keys){if($all.ContainsKey($p)){Need ($all[$p] -ceq $map[$p]) "conflicting pin $p"};$all[$p]=$map[$p];$checks++}
}
$facts=@{TransportVerified=$false;TerminalVerified=$false;BusinessVerified=$false;AuditPersisted=$false;AuditIdentityVerified=$false;AuditMapsVerified=$false;WriterDelegationVerified=$false;RoleVerified=$false;TimingVerified=$false;OutputVerified=$false}
$r=ReadJson "$d/supervision.json"
. (Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-staged-integrity.ps1')
Need ($r.NextAllowed -and -not $r.PendingResources -and -not $r.ExitUnconfirmed) 'terminal supervision required before trace interpretation'
foreach($role in @('Writer','Producer')){Need (-not $r[$role].StartAttempted -or $r[$role].OwnedExit) 'owned exit before review'}
$facts.TerminalVerified=$true
$intent=ReadJson "$d/intent.json";$intentHash=(Get-FileHash "$d/intent.json").Hash
Need ($r.Token -ceq $intent.Token -and $r.InputSHA256 -ceq $InputSHA256 -and $intent.InputSHA256 -ceq $InputSHA256) 'trace attempt identity'
$timing=Read-SITrace ([Convert]::FromBase64String($r.DiagnosticBase64)) $intent.Token $InputSHA256
Save-SIReview timing-review.json ([ordered]@{RawBase64=$r.DiagnosticBase64;Truncated=$r.DiagnosticTruncated;Timing=$timing;Transport=$r.SupervisionOutcome;Reason=$r.Reason;Caveat='Local intervals include instrumentation and scheduling; not COMPLETE or exit timestamps.'})
$audit=ReadJson "$d/integrity-review.json"
$facts.AuditPersisted=$true
Need ($audit.Schema -is [long] -and $audit.Schema -eq 1 -and $audit.Scenario -ceq 'StagedIntegrity44' -and $audit.InputSHA256 -ceq $InputSHA256 -and $audit.IntentSHA256 -ceq $intentHash) 'audit identity'
$facts.AuditIdentityVerified=$true
Need ($audit.State -ceq 'VERIFIED' -and $null -eq $audit.Error -and $null -ne $audit.After -and $null -ne $audit.ElapsedMilliseconds -and $audit.ElapsedMilliseconds -is [ValueType] -and $audit.ElapsedMilliseconds -isnot [bool] -and $audit.ElapsedMilliseconds -ge 0) 'persisted full audit required'
Need (($audit.AbsentPaths|ConvertTo-Json -Compress -AsArray) -ceq ($b.AbsentPaths|ConvertTo-Json -Compress -AsArray)) 'absence pins attested by completed audit'
function Read-SIMap($ref,[string]$Name){
 Need ($null -ne $ref -and $ref.Path -is [string] -and $ref.Path.Replace('//','/') -ceq "$d/$Name") 'map path'
 Need ($ref.Count -eq $all.Count -and (Get-Item -LiteralPath $ref.Path).Length -eq $ref.Length -and (Get-FileHash -LiteralPath $ref.Path).Hash -ceq $ref.SHA256) 'map metadata'
 $map=ReadJson $ref.Path
 Need ($map.Count -eq $all.Count) 'map count'
 foreach($p in $all.Keys){Need ($map.Contains($p) -and $map[$p] -ceq $all[$p]) "map $p"}
 return $map
}
$null=Read-SIMap $audit.Before supervisor.before-map.json
$null=Read-SIMap $audit.After supervisor.after-map.json
$facts.AuditMapsVerified=$true
if($r.SupervisionOutcome -cne 'SUCCESS'){
 $verdict=[ordered]@{Status='STAGED_PREFIX_FAILED';ObservationPassed=$false;Reason=$r.Reason;TimingStatus=$timing.Status;TraceRecords=$timing.Records.Count;HashChecks=$checks;UniquePins=$all.Count;FailedOutputInspected=$false;Overall='NOT_ACCEPTED'}
 Save-SIReview verification.json $verdict
 $verdict|ConvertTo-Json -Depth 12
 exit 1
}
Need ($timing.Status -ceq 'COMPLETE') 'complete four-phase timing required'
$facts.TimingVerified=$true
# Refuse any business/file inspection without confirmed terminal supervision.
Need ($r.SupervisionOutcome -ceq 'SUCCESS' -and $r.NextAllowed -and -not $r.PendingResources -and -not $r.ExitUnconfirmed -and $null -eq $r.Reason) 'transport terminal success'
$intent=ReadJson "$d/intent.json";$intentHash=(Get-FileHash "$d/intent.json").Hash
Need ($intent.Schema -eq 1 -and $intent.Computation -eq 44 -and $intent.Ceiling -eq 44 -and $intent.MaxStarts -eq 2 -and $intent.MaxProcesses -eq 2 -and $intent.AttemptConsumed -and $intent.Scenario -ceq 'StagedIntegrity44') 'single attempt'
Need ($intent.InputSHA256 -ceq $InputSHA256 -and $r.InputSHA256 -ceq $InputSHA256 -and $r.Token -ceq $intent.Token -and $r.Scenario -ceq $intent.Scenario) 'attempt identity'
foreach($role in @('Writer','Producer')){
 $p=$r[$role];Need ($p.StartAttempted -and $p.Associated -and $p.OwnedExit -and $p.ExitCode -eq 0 -and -not $p.KillRequested -and $null -eq $p.Error) "$role owned exit0"
 $claim=ReadJson ("$d/"+$role.ToLowerInvariant()+'.claim.json')
 Need ($claim.Role -ceq $role -and $claim.Computation -eq 44 -and $claim.Token -ceq $intent.Token -and $claim.InputSHA256 -ceq $InputSHA256 -and $claim.IntentSHA256 -ceq $intentHash) "$role claim"
}
foreach($name in @('RelayTask','ControlTask','DiagnosticTask')){Need ($r[$name] -ceq 'RanToCompletion') "terminal $name"}
Need ($r.Relay.StdoutEOF -and $r.Relay.StderrEOF -and $null -eq $r.Relay.Error) 'real EOF'
$strict=[Text.UTF8Encoding]::new($false,$true)
$raw=[Convert]::FromBase64String($r.ControlBase64);$diagnostic=[Convert]::FromBase64String($r.DiagnosticBase64)
Need ($raw.Length -le 8192 -and $diagnostic.Length -le 1920 -and -not $r.ControlTruncated -and -not $r.DiagnosticTruncated) 'control bounds'
$text=$strict.GetString($raw)
Need ($text.EndsWith([string][char]10) -and -not $text.Contains([char]13)) 'control framing'
$lines=@($text.Split([char]10,[StringSplitOptions]::RemoveEmptyEntries));Need ($lines.Count -eq 2) 'two receipts'
$receipts=@()
foreach($line in $lines){
 Need ($strict.GetByteCount($line)+1 -le 4096) 'receipt bound'
 $receipt=$line|ConvertFrom-Json -AsHashtable
 Need ($receipt.Schema -eq 1 -and $receipt.Role -ceq 'writer' -and $receipt.Token -ceq $intent.Token -and $receipt.InputSHA256 -ceq $InputSHA256) 'receipt identity'
 $receipts+=,$receipt
}
Need ($receipts[0].State -ceq 'READY' -and $receipts[0].Count -eq 5 -and $receipts[1].State -ceq 'COMPLETE' -and $receipts[1].Count -eq 12) 'receipt schema/order'
Need ($receipts[1].InputEOF -and $receipts[1].CloseSucceeded -and -not $receipts[1].Faulted) 'finalized receipt'
$child=ReadJson "$d/child-result.json"
Need ($child.Schema -eq 1 -and $child.Status -ceq 'HOST_PREFIX_COMPLETED' -and $child.ExitCode -eq 0 -and $child.Cutoff -and $child.InputSHA256 -ceq $InputSHA256 -and $child.IntentSHA256 -ceq $intentHash) 'business completion'
foreach($name in @('Error','RecordingError','IdentityError')){Need ($null -eq $child[$name]) "business $name"}
Need ($child.SecondaryErrors.Count -eq 0 -and -not $child.Aborted -and -not $child.Overflow -and -not $child.CaptureJournalFailed -and $child.UnclosedIntervals -eq 0) 'business clean terminal'
Need ($child.Groups.Count -eq 2 -and $child.SchemaSuite.Count -eq 2) 'two business groups'
foreach($i in 0..1){
 $name=@('ContractAPIAvailable','BaselineCore')[$i]
 Need ($child.Groups[$i].Name -ceq $name -and $child.Groups[$i].Passed -and $child.Groups[$i].Facts.Verified -and $child.SchemaSuite[$i].Name -ceq $name -and -not $child.SchemaSuite[$i].ActualRejected -and -not $child.SchemaSuite[$i].ExpectedRejected -and $null -eq $child.SchemaSuite[$i].Error) "business $name"
}
foreach($flag in @('SourceVerified','LiveActionsPerformed','RuntimeApproved','FormalAcceptance','RealGateImplemented')){Need ($r[$flag] -eq $false -and $child[$flag] -eq $false) "false $flag"}
$facts.TransportVerified=$true
$facts.BusinessVerified=$true
$writer=ReadJson "$d/writer.role-result.json"
Need ($writer.ExitCode -eq 0 -and $null -eq $writer.Error -and $writer.InputSHA256 -ceq $InputSHA256 -and $writer.IntentSHA256 -ceq $intentHash) 'writer result'
Need ($writer.Role -ceq 'Writer' -and $writer.Scenario -ceq 'StagedIntegrity44' -and $null -eq $writer.After -and $writer.AfterVerificationMode -ceq 'PARENT_POST_EXIT_REQUIRED') 'writer delegation'
$facts.WriterDelegationVerified=$true
$facts.RoleVerified=$true
$row=ReadJson "$d/result.json"
Need ($row.Computation -eq 44 -and -not $row.BusinessVerified -and ($row.Result|ConvertTo-Json -Depth 80 -Compress) -ceq ($r|ConvertTo-Json -Depth 80 -Compress)) 'terminal result unchanged'
Need (($row.Before|ConvertTo-Json -Compress) -ceq ($audit.Before|ConvertTo-Json -Compress) -and ($row.After|ConvertTo-Json -Compress) -ceq ($audit.After|ConvertTo-Json -Compress)) 'result audit references'
$writerMap=Read-SIMap $writer.Before writer.before-map.json
$maps=@($child.Before,$child.After,$writerMap)
foreach($map in $maps){Need ($map.Count -eq $all.Count) 'map count';foreach($p in $all.Keys){Need ($map[$p] -ceq $all[$p]) "map $p"}}
$events=@(Get-Content "$d/stages.jsonl"|ForEach-Object {$_|ConvertFrom-Json -AsHashtable})
$prior=@(Get-Content '.superpowers/sdd/2026-11-04-runtime-copy-transport-v1/stages.jsonl'|ForEach-Object {$_|ConvertFrom-Json -AsHashtable})
Need ($events.Count -eq 62 -and $prior.Count -eq 62 -and (Get-Item "$d/stages.jsonl").Length -le 1048576) '62 events'
$stack=[Collections.Generic.Stack[object]]::new();$lastTicks=-1L
foreach($i in 0..61){
 $event=$events[$i]
 foreach($field in @('Schema','Seq','EventId','ParentId','Stage','Kind','Case')){Need ($event[$field] -ceq $prior[$i][$field]) "stage $i $field"}
 Need ($event.ElapsedTicks -ge $lastTicks -and $event.Frequency -gt 0) 'event time';$lastTicks=$event.ElapsedTicks
 if($event.Kind -ceq 'begin'){$stack.Push($event)}else{Need ($event.Kind -ceq 'end' -and $stack.Count -gt 0) 'paired end';$start=$stack.Pop();Need ($start.EventId -eq $event.EventId -and $start.Stage -ceq $event.Stage -and $start.Case -ceq $event.Case) 'interval identity'}
}
Need ($stack.Count -eq 0 -and $child.LastEvent.Seq -eq 62 -and $child.LastEvent.Stage -ceq 'input.after') 'closed structure'
$journal=@(Get-Content "$d/journal.jsonl"|ForEach-Object {$_|ConvertFrom-Json -AsHashtable})
Need ($journal.Count -eq 2 -and $journal[0].Seq -eq 1 -and $journal[1].Seq -eq 2 -and $journal[0].Kind -ceq 'begin' -and $journal[1].Kind -ceq 'end' -and $journal[0].Case -ceq 'StagedIntegrity44' -and $journal[1].Case -ceq 'StagedIntegrity44') 'attempt journal'
$childClaim=ReadJson "$d/child.claim.json"
Need ($childClaim.InputSHA256 -ceq $InputSHA256 -and $childClaim.IntentSHA256 -ceq $intentHash -and $childClaim.Token -ceq $intent.Token) 'child claim'
# Only now inspect successful, stable writer data. No failed-output salvage.
foreach($i in 0..1){
 $label=@('Stdout','Stderr')[$i];$path="$d/data/"+$label.ToLowerInvariant()+'.bin'
 if($i -eq 0){$expected=$strict.GetBytes("PASS: ContractAPIAvailable`r`nPASS: BaselineCore`r`n")}else{$expected=[byte[]]::new(0)}
 $n=@(48,0)[$i];$e=$r[$label]
 Need ($e.Available -and $e.Length -eq $n -and (Get-Item $path).Length -eq $n) 'stable length'
 $bytes=[IO.File]::ReadAllBytes($path)
 Need ([Linq.Enumerable]::SequenceEqual([byte[]]$bytes,[byte[]]$expected)) 'exact business bytes'
 $hash=(Get-FileHash $path).Hash
 Need ($e.SHA256 -ceq $hash -and $receipts[1][$label+'Length'] -eq $n -and $receipts[1][$label+'SHA256'] -ceq $hash -and $r.Relay[$label+'SHA256'] -ceq $hash -and $r.Relay.Read[$i] -eq $n -and $r.Relay.Admitted[$i] -eq $n -and $r.Relay.Sent[$i] -eq $n) 'output receipt/hash/count'
}
Need (-not(Test-Path "$d/execution.stop.json")) 'no execution stop'
$facts.OutputVerified=$true
$decision=Test-SIObservation $facts
Need $decision.Passed ('conjunctive gate '+($decision.Reasons -join ','))
$verdict=[ordered]@{Status='STAGED_PREFIX_VERIFIED';ObservationPassed=$decision.Passed;Facts=$facts;ParentAuditMilliseconds=$audit.ElapsedMilliseconds;InputSHA256=$InputSHA256;HashChecks=$checks;UniquePins=$all.Count;Maps=$maps.Count;Events=62;Intervals=31;StdoutBytes=48;StderrBytes=0;Starts=2;Compilations=0;ElapsedMs=$r.ElapsedMs;PendingResources=0;UnknownExits=0;Counters=@(44,34,6,217);Overall='NOT_ACCEPTED'}
Save-SIReview verification.json $verdict
$verdict|ConvertTo-Json -Depth 12
}catch{
 $failure=[ordered]@{Status='STAGED_PREFIX_FAILED';ObservationPassed=$false;Reason=$_.Exception.Message;Stack=$_.ScriptStackTrace;Overall='NOT_ACCEPTED'}
 if(-not(Test-Path "$review/verification.json")){Save-SIReview verification.json $failure}
 $failure|ConvertTo-Json -Depth 12
 exit 1
}
