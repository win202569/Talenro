# Definition-only, opt-in shadow evidence helpers. Never launch or own processes.
function Set-C12ShadowIssue($Context,[string]$Reason){
 if($null -eq $Context.FirstReason){$Context.FirstReason=$Reason}
 if($Context.Contains('Session') -and $null -ne $Context.Session){
  if($null -eq $Context.Session.FirstReason){$Context.Session.FirstReason=$Reason}
 }
}
function Write-C12ShadowWarning($Session,[string]$Reason){
 try{Write-Warning ('C12 shadow audit '+$Session.AuditId+' '+$Reason) -ErrorAction Stop}catch{}
}
function Get-C12ShadowDigest([byte[]]$Bytes){[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($Bytes))}
function Write-C12ShadowRecord($Session,[string]$Leaf,$Value){
 $path=Resolve-C12ShadowPath $Session.EvidenceRoot $Leaf
 $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 30 -Compress))
 if($bytes.Length -gt 1048576){throw 'LIMIT'}
 $stream=[IO.File]::Open($path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
 try{$stream.Write($bytes,0,$bytes.Length);$stream.Flush()}finally{$stream.Dispose()}
 $digest=Get-C12ShadowDigest $bytes
 $saved=Read-C12ShadowBytes $path 1048576
 if($saved.Length -ne $bytes.Length -or (Get-C12ShadowDigest $saved) -cne $digest){throw 'IO'}
 return [ordered]@{Path=$Leaf;Length=$bytes.Length;SHA256=$digest}
}
function Read-C12ShadowReference($Session,$Reference){
 Assert-C12ShadowKeys $Reference @('Path','Length','SHA256')
 if($Reference.Path -isnot [string] -or $Reference.Path -cnotmatch '^[a-z0-9.-]+\.json$' -or $Reference.SHA256 -isnot [string] -or $Reference.SHA256 -cnotmatch '^[A-F0-9]{64}$' -or ($Reference.Length -isnot [int] -and $Reference.Length -isnot [long]) -or $Reference.Length -lt 1 -or $Reference.Length -gt 1048576){throw 'IDENTITY'}
 $bytes=Read-C12ShadowBytes (Resolve-C12ShadowPath $Session.EvidenceRoot $Reference.Path) 1048576
 if($bytes.Length -ne $Reference.Length -or (Get-C12ShadowDigest $bytes) -cne $Reference.SHA256){throw 'IDENTITY'}
 return (ConvertFrom-C12ShadowJSON ([Text.UTF8Encoding]::new($false,$true).GetString($bytes)))
}
function New-C12ShadowSession([string]$InvocationRoot,[bool]$Synthetic=$false){
 $session=[ordered]@{Clock=[Diagnostics.Stopwatch]::StartNew();AuditId=[guid]::NewGuid().ToString('N');EvidenceRoot=$null;Groups=[Collections.Generic.List[object]]::new();FirstReason=$null;ModuleSHA256=$null;Synthetic=$Synthetic;Finalized=$false}
 try{
  $root=Assert-C12ShadowRoot $InvocationRoot
  $parent=Resolve-C12ShadowPath $root '.superpowers/c12-shadow'
  $null=[IO.Directory]::CreateDirectory($parent)
  $null=Assert-C12ShadowRoot $parent
  $destination=Resolve-C12ShadowPath $root ('.superpowers/c12-shadow/'+$session.AuditId)
  # New-Item without Force refuses occupied directories; no automatic retry.
  $null=New-Item -ItemType Directory -Path $destination -ErrorAction Stop
  $session.EvidenceRoot=Assert-C12ShadowRoot $destination
  $module=$MyInvocation.MyCommand.ScriptBlock.File
  $session.ModuleSHA256=Get-C12ShadowDigest (Read-C12ShadowBytes $module 1048576)
  $null=Write-C12ShadowRecord $session 'session.json' ([ordered]@{Schema=1;Mode='SHADOW';Synthetic=$Synthetic;AuditId=$session.AuditId;ModuleSHA256=$session.ModuleSHA256})
 }catch{Set-C12ShadowIssue $session (Get-C12ShadowReason $_);Write-C12ShadowWarning $session $session.FirstReason}
 return $session
}
function Copy-C12ShadowIdentity($Identity){
 Assert-C12ShadowKeys $Identity @('GroupID','Profile','Package','RunSuffix','CandidateTree')
 $copy=[ordered]@{}
 foreach($key in @('GroupID','Profile','Package','RunSuffix','CandidateTree')){
  $value=$Identity[$key]
  if($key -eq 'CandidateTree' -and $null -eq $value){$copy[$key]=$null;continue}
  if($value -isnot [string] -or $value.Length -gt 256 -or $value -cnotmatch '^[A-Za-z0-9_./-]+$'){throw 'IDENTITY'}
  $copy[$key]=$value
 }
 return $copy
}
function Get-C12ShadowBinding($Group){
 $value=[ordered]@{AuditId=$Group.Session.AuditId;Leaf=$Group.Leaf;Synthetic=$Group.Session.Synthetic;Identity=(Copy-C12ShadowIdentity $Group.Identity);SourceRootSHA256=(Get-C12ShadowDigest ([Text.UTF8Encoding]::new($false).GetBytes($Group.SourceRoot)));ManifestSHA256=$Group.Manifest.SHA256;ModuleSHA256=$Group.Session.ModuleSHA256}
 return (ConvertTo-Json -InputObject $value -Depth 10 -Compress)
}
function New-C12ShadowEnvelope($Group,[string]$Kind,$Payload){
 return [ordered]@{Schema=1;Mode='SHADOW';Synthetic=$Group.Session.Synthetic;AuditId=$Group.Session.AuditId;Kind=$Kind;Binding=$Group.Binding;Payload=$Payload}
}
function Start-C12ShadowGroup($Session,[string]$SourceRoot,[System.Collections.IDictionary]$Identity,[datetime]$Deadline){
 $group=[ordered]@{Session=$Session;Leaf=$null;Identity=$null;SourceRoot=$null;Manifest=$null;Binding=$null;Before=$null;After=$null;Record=$null;FirstReason=$null;MapState='INCOMPLETE';Finished=$false;TestFailed=$false;CleanupFailed=$false;OwnershipConfirmed=$false;Clock=[Diagnostics.Stopwatch]::StartNew()}
 try{
  if($Session.Finalized){throw 'IDENTITY'}
  $group.Leaf='g{0:D4}' -f ($Session.Groups.Count+1)
  $Session.Groups.Add($group)
  if($null -ne $Session.FirstReason){throw ([string]$Session.FirstReason)}
  if([datetime]::UtcNow -ge $Deadline){throw 'DEADLINE'}
  $group.Identity=Copy-C12ShadowIdentity $Identity
  $group.SourceRoot=Assert-C12ShadowRoot $SourceRoot
  $group.Manifest=Read-C12ShadowManifest $group.SourceRoot
  if($group.Manifest.State -cne 'READY'){throw ([string]$group.Manifest.Reason)}
  $group.Binding=Get-C12ShadowBinding $group
  $map=Get-C12ShadowMap $group.SourceRoot $group.Manifest $Deadline
  $group.MapState=$map.State
  if($map.State -ceq 'INCOMPLETE'){Set-C12ShadowIssue $group $map.Reason}
  $group.Before=Write-C12ShadowRecord $Session ($group.Leaf+'.before.json') (New-C12ShadowEnvelope $group 'before' $map)
 }catch{Set-C12ShadowIssue $group (Get-C12ShadowReason $_);Write-C12ShadowWarning $Session $group.FirstReason}
 return $group
}
function Assert-C12ShadowEnvelope($Group,$Reference,[string]$Kind){
 $value=Read-C12ShadowReference $Group.Session $Reference
 Assert-C12ShadowKeys $value @('Schema','Mode','Synthetic','AuditId','Kind','Binding','Payload')
 if($value.Schema -isnot [long] -and $value.Schema -isnot [int]){throw 'IDENTITY'}
 if($value.Schema -ne 1 -or $value.Mode -cne 'SHADOW' -or $value.Synthetic -isnot [bool] -or $value.Synthetic -ne $Group.Session.Synthetic -or $value.AuditId -cne $Group.Session.AuditId -or $value.Kind -cne $Kind -or $value.Binding -cne $Group.Binding -or (Get-C12ShadowBinding $Group) -cne $Group.Binding){throw 'IDENTITY'}
 return $value.Payload
}
function Complete-C12ShadowGroup($Group,[bool]$TestFailed,[bool]$CleanupFailed,[bool]$OwnershipConfirmed,[datetime]$Deadline){
 try{
  if($Group.Finished -or $Group.Session.Finalized){throw 'IDENTITY'}
  $Group.Finished=$true;$Group.TestFailed=$TestFailed;$Group.CleanupFailed=$CleanupFailed;$Group.OwnershipConfirmed=$OwnershipConfirmed
  if($null -ne $Group.FirstReason){throw ([string]$Group.FirstReason)}
  if($CleanupFailed){throw 'CLEANUP'}
  if(-not $OwnershipConfirmed){throw 'OWNERSHIP'}
  if([datetime]::UtcNow -ge $Deadline){throw 'DEADLINE'}
  $before=Assert-C12ShadowEnvelope $Group $Group.Before 'before'
  $after=Get-C12ShadowMap $Group.SourceRoot $Group.Manifest $Deadline
  if($after.State -ceq 'INCOMPLETE'){throw ([string]$after.Reason)}
  $Group.After=Write-C12ShadowRecord $Group.Session ($Group.Leaf+'.after.json') (New-C12ShadowEnvelope $Group 'after' $after)
  $Group.MapState='MATCHED'
  if($before.State -ceq 'DRIFT' -or $after.State -ceq 'DRIFT'){$Group.MapState='DRIFT'}
  if($before.State -ceq 'INCOMPLETE'){throw ([string]$before.Reason)}
  $payload=[ordered]@{State='PROVISIONAL';MapState=$Group.MapState;Before=$Group.Before;After=$Group.After;TestFailed=$TestFailed;CleanupFailed=$CleanupFailed;OwnershipConfirmed=$OwnershipConfirmed;ElapsedMilliseconds=$Group.Clock.Elapsed.TotalMilliseconds}
  $Group.Record=Write-C12ShadowRecord $Group.Session ($Group.Leaf+'.group.json') (New-C12ShadowEnvelope $Group 'group' $payload)
  return [ordered]@{State='PROVISIONAL';Reason=$null;Reference=$Group.Record}
 }catch{
  Set-C12ShadowIssue $Group (Get-C12ShadowReason $_);$Group.MapState='INCOMPLETE'
  Write-C12ShadowWarning $Group.Session $Group.FirstReason
  return [ordered]@{State='INCOMPLETE';Reason=$Group.FirstReason;Reference=$null}
 }finally{$Group.Clock.Stop()}
}
function Complete-C12ShadowSession($Session,[int]$FormalExitCode,[bool]$OuterCleanupFailed,[bool]$EnvironmentRestored,[datetime]$Deadline){
 $state='MATCHED';$refs=[Collections.Generic.List[object]]::new();$reference=$null
 try{
  if($Session.Finalized){throw 'IDENTITY'}
  $Session.Finalized=$true
  if([datetime]::UtcNow -ge $Deadline){throw 'DEADLINE'}
  if($OuterCleanupFailed -or -not $EnvironmentRestored){Set-C12ShadowIssue $Session 'CLEANUP'}
  if($Session.Groups.Count -eq 0){Set-C12ShadowIssue $Session 'NO_GROUPS'}
  foreach($group in $Session.Groups){
   if([datetime]::UtcNow -ge $Deadline){throw 'DEADLINE'}
   try{
    if($null -ne $group.FirstReason){throw ([string]$group.FirstReason)}
    if(-not $group.Finished -or $null -eq $group.Record){throw 'IDENTITY'}
    $record=Assert-C12ShadowEnvelope $group $group.Record 'group'
    $before=Assert-C12ShadowEnvelope $group $record.Before 'before'
    $after=Assert-C12ShadowEnvelope $group $record.After 'after'
    if($record.State -cne 'PROVISIONAL' -or $record.TestFailed -isnot [bool] -or $record.CleanupFailed -isnot [bool] -or $record.OwnershipConfirmed -isnot [bool]){throw 'IDENTITY'}
    if($record.TestFailed -ne $group.TestFailed -or $record.CleanupFailed -ne $group.CleanupFailed -or $record.OwnershipConfirmed -ne $group.OwnershipConfirmed){throw 'IDENTITY'}
    if($record.CleanupFailed){throw 'CLEANUP'}
    if(-not $record.OwnershipConfirmed){throw 'OWNERSHIP'}
    foreach($map in @($before,$after)){
     if($map.State -cnotin @('MATCHED','DRIFT')){throw 'IDENTITY'}
     if($map.State -ceq 'DRIFT'){$state='DRIFT'}
    }
    $refs.Add($group.Record)
   }catch{Set-C12ShadowIssue $Session (Get-C12ShadowReason $_)}
  }
  if($null -ne $Session.FirstReason){$state='INCOMPLETE'}
  # Elapsed wall time through summary preparation, excluding its own publication.
  $summary=[ordered]@{Schema=1;Mode='SHADOW';Synthetic=$Session.Synthetic;AuditId=$Session.AuditId;ModuleSHA256=$Session.ModuleSHA256;State=$state;Reason=$Session.FirstReason;Groups=$refs.ToArray();FormalExitCode=$FormalExitCode;OuterCleanupFailed=$OuterCleanupFailed;EnvironmentRestored=$EnvironmentRestored;ElapsedMilliseconds=$Session.Clock.Elapsed.TotalMilliseconds}
  if([datetime]::UtcNow -ge $Deadline){throw 'DEADLINE'}
  $reference=Write-C12ShadowRecord $Session 'run.json' $summary
 }catch{Set-C12ShadowIssue $Session (Get-C12ShadowReason $_);$state='INCOMPLETE';$reference=$null}
 if($state -ceq 'INCOMPLETE'){Write-C12ShadowWarning $Session $Session.FirstReason}
 return [ordered]@{State=$state;Reason=$Session.FirstReason;Reference=$reference}
}
function Get-C12ShadowReason($ErrorRecord){
 $message=$ErrorRecord.Exception.Message
 if($message -cin @('INVALID_MANIFEST','UNSAFE_PATH','LIMIT','IO','IDENTITY','DEADLINE','OWNERSHIP','CLEANUP','NO_GROUPS')){return $message}
 return 'IO'
}
function Assert-C12ShadowRoot([string]$Root){
 if(-not [IO.Path]::IsPathFullyQualified($Root)){throw 'UNSAFE_PATH'}
 $full=[IO.Path]::GetFullPath($Root).TrimEnd([IO.Path]::DirectorySeparatorChar)
 $cursor=$full
 while($cursor){
  if(Test-Path -LiteralPath $cursor){if((Get-Item -LiteralPath $cursor -Force -ErrorAction Stop).Attributes -band [IO.FileAttributes]::ReparsePoint){throw 'UNSAFE_PATH'}}
  $parent=[IO.Path]::GetDirectoryName($cursor);if($parent -eq $cursor){break};$cursor=$parent
 }
 if(-not(Test-Path -LiteralPath $full -PathType Container)){throw 'IO'}
 return $full
}
function Resolve-C12ShadowPath([string]$Root,[string]$Relative){
 $base=Assert-C12ShadowRoot $Root
 if([string]::IsNullOrEmpty($Relative) -or $Relative -cnotmatch '^[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*$'){throw 'UNSAFE_PATH'}
 foreach($part in $Relative.Split('/')){if($part -in @('.','..') -or $part.EndsWith('.') -or $part -match '^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])(?:\.|$)'){throw 'UNSAFE_PATH'}}
 $full=[IO.Path]::GetFullPath((Join-Path $base $Relative))
 if(-not $full.StartsWith($base+[IO.Path]::DirectorySeparatorChar,[StringComparison]::OrdinalIgnoreCase)){throw 'UNSAFE_PATH'}
 $cursor=$full
 while($cursor.Length -gt $base.Length){
  if(Test-Path -LiteralPath $cursor){if((Get-Item -LiteralPath $cursor -Force -ErrorAction Stop).Attributes -band [IO.FileAttributes]::ReparsePoint){throw 'UNSAFE_PATH'}}
  $cursor=[IO.Path]::GetDirectoryName($cursor)
 }
 return $full
}
function Read-C12ShadowBytes([string]$Path,[long]$Limit){
 $stream=[IO.File]::Open($Path,[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::Read)
 $memory=[IO.MemoryStream]::new()
 try{
  if($stream.Length -gt $Limit){throw 'LIMIT'}
  $buffer=[byte[]]::new(65536);$total=0L
  while(($n=$stream.Read($buffer,0,$buffer.Length)) -gt 0){$total+=$n;if($total -gt $Limit){throw 'LIMIT'};$memory.Write($buffer,0,$n)}
  return ,$memory.ToArray()
 }finally{$stream.Dispose();$memory.Dispose()}
}
function ConvertFrom-C12ShadowJSON([string]$Text){
 $document=$null
 try{
  $document=[System.Text.Json.JsonDocument]::Parse($Text)
  $pending=[Collections.Generic.Stack[System.Text.Json.JsonElement]]::new();$pending.Push($document.RootElement)
  while($pending.Count){
   $element=$pending.Pop()
   if($element.ValueKind -eq [System.Text.Json.JsonValueKind]::Object){
    $keys=[Collections.Generic.HashSet[string]]::new([StringComparer]::OrdinalIgnoreCase)
    foreach($property in $element.EnumerateObject()){if(-not $keys.Add($property.Name)){throw 'INVALID_MANIFEST'};$pending.Push($property.Value)}
   }elseif($element.ValueKind -eq [System.Text.Json.JsonValueKind]::Array){foreach($item in $element.EnumerateArray()){$pending.Push($item)}}
  }
  return ($Text|ConvertFrom-Json -AsHashtable -ErrorAction Stop)
 }catch{throw 'INVALID_MANIFEST'}finally{if($null -ne $document){$document.Dispose()}}
}
function Assert-C12ShadowKeys($Value,[string[]]$Keys){
 if($Value -isnot [Collections.IDictionary] -or $Value.Count -ne $Keys.Count){throw 'INVALID_MANIFEST'}
 foreach($key in $Keys){if($key -cnotin @($Value.Keys)){throw 'INVALID_MANIFEST'}}
}
function Read-C12ShadowManifest([string]$SourceRoot){
 try{
  $root=Assert-C12ShadowRoot $SourceRoot
  $path=Resolve-C12ShadowPath $root 'testdata/c12/shadow-source-manifest.v1.json'
  $bytes=Read-C12ShadowBytes $path 1048576
  $text=[Text.UTF8Encoding]::new($false,$true).GetString($bytes)
  $data=ConvertFrom-C12ShadowJSON $text
  Assert-C12ShadowKeys $data @('Schema','Mode','Files')
  if(($data.Schema -isnot [long] -and $data.Schema -isnot [int]) -or $data.Schema -ne 1 -or $data.Mode -isnot [string] -or $data.Mode -cne 'SHADOW' -or $data.Files -isnot [array]){throw 'INVALID_MANIFEST'}
  if($data.Files.Count -lt 1 -or $data.Files.Count -gt 128){throw 'LIMIT'}
  $seen=[Collections.Generic.HashSet[string]]::new([StringComparer]::OrdinalIgnoreCase);$files=[Collections.Generic.List[object]]::new()
  foreach($entry in $data.Files){
   Assert-C12ShadowKeys $entry @('Path','SHA256')
   if($entry.Path -isnot [string] -or $entry.SHA256 -isnot [string] -or $entry.SHA256 -cnotmatch '^[A-F0-9]{64}$'){throw 'INVALID_MANIFEST'}
   if(-not $seen.Add($entry.Path)){throw 'INVALID_MANIFEST'}
   $null=Resolve-C12ShadowPath $root $entry.Path
   $files.Add([ordered]@{Path=$entry.Path;SHA256=$entry.SHA256})
  }
  return [ordered]@{State='READY';Reason=$null;SourceRoot=$root;SHA256=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($bytes));Files=$files.ToArray()}
 }catch{return [ordered]@{State='INCOMPLETE';Reason=(Get-C12ShadowReason $_);SourceRoot=$null;SHA256=$null;Files=@()}}
}
function Get-C12ShadowMap([string]$SourceRoot,$Manifest,[datetime]$Deadline){
 $clock=[Diagnostics.Stopwatch]::StartNew();$observed=[Collections.Generic.List[object]]::new()
 try{
  if([datetime]::UtcNow -ge $Deadline){throw 'DEADLINE'}
  if($Manifest.State -cne 'READY'){throw ([string]$Manifest.Reason)}
  $root=Assert-C12ShadowRoot $SourceRoot
  if($root -cne $Manifest.SourceRoot){throw 'IDENTITY'}
  $total=0L;$state='MATCHED'
  foreach($entry in $Manifest.Files){
   if([datetime]::UtcNow -ge $Deadline){throw 'DEADLINE'}
   $path=Resolve-C12ShadowPath $root $entry.Path
   if(-not(Test-Path -LiteralPath $path)){$state='DRIFT';$observed.Add([ordered]@{Path=$entry.Path;Present=$false;Length=$null;SHA256=$null});continue}
   $stream=[IO.File]::Open($path,[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::Read)
   $sha=[Security.Cryptography.IncrementalHash]::CreateHash([Security.Cryptography.HashAlgorithmName]::SHA256)
   try{
    if($stream.Length -gt 33554432-$total){throw 'LIMIT'}
    $buffer=[byte[]]::new(65536);$length=0L
    while(($n=$stream.Read($buffer,0,$buffer.Length)) -gt 0){$total+=$n;$length+=$n;if($total -gt 33554432){throw 'LIMIT'};$sha.AppendData($buffer,0,$n)}
    $hash=[Convert]::ToHexString($sha.GetHashAndReset())
   }finally{$sha.Dispose();$stream.Dispose()}
   if($hash -cne $entry.SHA256){$state='DRIFT'}
   $observed.Add([ordered]@{Path=$entry.Path;Present=$true;Length=$length;SHA256=$hash})
  }
  return [ordered]@{State=$state;Reason=$null;Files=$observed.ToArray();ElapsedMilliseconds=$clock.Elapsed.TotalMilliseconds}
 }catch{return [ordered]@{State='INCOMPLETE';Reason=(Get-C12ShadowReason $_);Files=$observed.ToArray();ElapsedMilliseconds=$clock.Elapsed.TotalMilliseconds}}
 finally{$clock.Stop()}
}
