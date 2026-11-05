# Silent trace and identity helpers; no operational entry at import.
function Write-HTOnce([string]$Path,$Value){
 $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 100))
 $s=[IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
 try{$s.Write($bytes,0,$bytes.Length);$s.Flush()}finally{$s.Dispose()}
}
function Get-HTHash([string]$Path){return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash}
function Get-HTError($Record){
 if($null -eq $Record){return $null}
 return [ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Position=$Record.InvocationInfo.PositionMessage;Stack=$Record.ScriptStackTrace}
}
function Assert-HTInputs($Binding){
 $observed=[ordered]@{}
 foreach($p in $Binding.AbsentPaths){if(Test-Path -LiteralPath $p){throw ('HT: absent pin '+$p)}}
 foreach($map in @($Binding.SourceSHA256,$Binding.PreservedSHA256,$Binding.HistoricalEvidenceSHA256,$Binding.HelperSHA256)){
  foreach($p in $map.Keys){
   $path=if([IO.Path]::IsPathFullyQualified($p)){$p}else{Join-Path $script:HTRepo $p}
   $hash=Get-HTHash $path
   if($hash -cne $map[$p]){throw ('HT: input drift '+$p)}
   $observed[$p]=$hash
  }
 }
 return $observed
}
function Write-HTEvent([string]$Stage,[string]$Kind,[string]$Case){
 if($script:HTJournalFailed){throw 'HT: journal latched'}
 try{
  if($Kind -cnotin @('begin','end','error') -or $Case -cnotin @('','ContractAPIAvailable','BaselineCore') -or $Stage -cnotmatch '^[a-z][a-z.-]{0,63}$'){throw 'HT: event fields'}
  if($Kind -ceq 'begin'){
   $script:HTEventId++
   $frame=[ordered]@{EventId=$script:HTEventId;ParentId=$(if($script:HTStack.Count){$script:HTStack.Peek().EventId}else{$null});Stage=$Stage;Case=$Case}
  }else{
   if(-not $script:HTStack.Count){throw 'HT: empty event stack'}
   $frame=$script:HTStack.Peek()
   if($frame.Stage -cne $Stage -or $frame.Case -cne $Case){throw 'HT: event stack mismatch'}
  }
  $event=[ordered]@{Schema=1;Seq=([long]$script:HTSeq+1);EventId=[long]$frame.EventId;ParentId=$frame.ParentId;Stage=$Stage;Kind=$Kind;Case=$Case;ElapsedTicks=[long]$script:HTClock.ElapsedTicks;Frequency=[long][Diagnostics.Stopwatch]::Frequency}
  $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json $event -Depth 8 -Compress)+[char]10)
  if($script:HTSeq -ge 4096 -or $script:HTBytes+$bytes.Length -gt 1048576){throw 'HT: journal bound'}
  $script:HTJournal.Write($bytes,0,$bytes.Length);$script:HTJournal.Flush()
  $script:HTSeq++;$script:HTBytes+=$bytes.Length;$script:HTLastEvent=$event
  if($Kind -ceq 'begin'){$script:HTStack.Push($frame)}else{$null=$script:HTStack.Pop()}
 }catch{
  $script:HTJournalFailed=$true
  if($null -eq $script:HTRecordingError){$script:HTRecordingError=Get-HTError $_}
  throw
 }
}
function Write-HTFailureEvent([string]$Stage,$Original){
 # Nonthrowing instrumentation on original error/cleanup paths.
 if($script:HTJournalFailed){return}
 try{
  if($Stage -ceq 'error.entry'){
   if($script:HTStack.Count){$active=$script:HTStack.Peek();Write-HTEvent $active.Stage error $active.Case}
  }
  if($Stage -ceq 'terminal.flush.begin'){Write-HTEvent terminal.flush begin '';return}
  if($Stage -ceq 'terminal.flush.end'){Write-HTEvent terminal.flush end '';return}
  Write-HTEvent $Stage begin '';Write-HTEvent $Stage end ''
 }catch{
  if($null -eq $script:HTRecordingError){$script:HTRecordingError=Get-HTError $_}
 }
}
function Initialize-HT([string]$InputSHA256,[string]$IntentSHA256){
 $script:HTClock=[Diagnostics.Stopwatch]::StartNew()
 $intent=Get-Content (Join-Path $script:HTD intent.json) -Raw|ConvertFrom-Json -AsHashtable
 if($intent.Schema -ne 1 -or $intent.Computation -ne 44 -or $intent.Ceiling -ne 44 -or $intent.MaxProcesses -ne 2 -or $intent.AttemptConsumed -isnot [bool] -or -not $intent.AttemptConsumed -or $intent.InputSHA256 -cne $InputSHA256 -or $intent.Token -cnotmatch '^[0-9a-f-]{36}$'){throw 'HT: intent'}
 Write-HTOnce (Join-Path $script:HTD child.claim.json) ([ordered]@{InputSHA256=$InputSHA256;IntentSHA256=$IntentSHA256;Token=$intent.Token})
 $script:HTStack=[Collections.Generic.Stack[object]]::new();$script:HTSeq=0L;$script:HTEventId=0L;$script:HTBytes=0L;$script:HTJournalFailed=$false;$script:HTRecordingError=$null;$script:HTLastEvent=$null;$script:HTCutoff=$false
 $script:HTBefore=$null;$script:HTAfter=$null
 $script:HTJournal=[IO.File]::Open((Join-Path $script:HTD stages.jsonl),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
 Write-HTEvent input.before begin ''
 $script:HTBefore=Assert-HTInputs $script:HTBinding
 Write-HTEvent input.before end ''
}
function Complete-HT([int]$Code,[string]$InputSHA256,[string]$IntentSHA256){
 $status='INCOMPLETE';$identityError=$null
 try{
  Write-HTEvent input.after begin ''
  $script:HTAfter=Assert-HTInputs $script:HTBinding
  Write-HTEvent input.after end ''
 }catch{$identityError=Get-HTError $_;$Code=2}
 $groups=$script:Groups.ToArray()
 $schema=@($script:Details|Where-Object {$_.Name -ceq 'SchemaSuite'})
 $schemaRows=if($schema.Count -eq 1){@($schema[0].Value)}else{@()}
 $exact=$groups.Count -eq 2 -and $groups[0].Name -ceq 'ContractAPIAvailable' -and $groups[1].Name -ceq 'BaselineCore' -and $groups[0].Passed -eq $true -and $groups[1].Passed -eq $true -and $schemaRows.Count -eq 2
 if($exact){$exact=$schemaRows[0].Name -ceq 'ContractAPIAvailable' -and $schemaRows[1].Name -ceq 'BaselineCore' -and -not $schemaRows[1].ActualRejected -and $null -eq $schemaRows[1].Error}
 if($Code -eq 0 -and $exact -and $script:HTCutoff -and -not $script:Aborted -and $null -eq $script:HostPrimary -and $script:HostSecondary.Count -eq 0 -and -not $script:HostOverflow -and -not $journalState.Failed -and -not $script:HTJournalFailed -and $script:HTStack.Count -eq 0 -and $null -eq $identityError){$status='HOST_PREFIX_COMPLETED'}else{$Code=2}
 if($null -ne $script:HostPrimary){$status='ASSERTION_OR_WRAPPER_FAILED'}
 if($null -ne $identityError){$status='INPUT_DRIFT'}
 if($null -ne $script:HTRecordingError){$status='RECORDING_FAILED'}
 try{$script:HTJournal.Dispose()}catch{$script:HTRecordingError=Get-HTError $_;$status='RECORDING_FAILED';$Code=2}
 Write-HTOnce (Join-Path $script:HTD child-result.json) ([ordered]@{
  Schema=1;Status=$status;ExitCode=$Code;InputSHA256=$InputSHA256;IntentSHA256=$IntentSHA256
  Cutoff=$script:HTCutoff;Groups=(Capture-Value $groups);SchemaSuite=(Capture-Value $schemaRows)
  Error=(Get-HTError $script:HostPrimary);SecondaryErrors=@($script:HostSecondary|ForEach-Object {Get-HTError $_})
  RecordingError=$script:HTRecordingError;IdentityError=$identityError;Aborted=$script:Aborted;Overflow=$script:HostOverflow;CaptureJournalFailed=$journalState.Failed
  Before=$script:HTBefore;After=$script:HTAfter;LastEvent=$script:HTLastEvent;UnclosedIntervals=$script:HTStack.Count
  SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false
 })
 return $Code
}
