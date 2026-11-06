# Definition-only, opt-in shadow evidence helpers. Never launch or own processes.
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
