function ConvertFrom-SEPayload([System.Collections.IDictionary]$Entry){
 $fields=@('Base64','RawBase64'|Where-Object {$Entry.Contains($_)})
 if($fields.Count -ne 1){throw 'SE payload requires exactly one encoding field'}
 $value=$Entry[$fields[0]]
 if($value -isnot [string]){throw 'SE payload requires string'}
 try{$bytes=[Convert]::FromBase64String($value)}catch{throw 'SE malformed base64'}
 if($bytes.Length -ne $Entry.Length -or [Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($bytes)) -cne $Entry.SHA256){throw 'SE archive bytes'}
 return ,$bytes
}
# Definition-only offline fixture utilities; never launch an OS child process.
function Get-SEBase{return 'D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator\.superpowers\sdd\2026-11-05-staged-evidence-fixtures-v2'}
function Assert-SEPath([string]$Path){
 if(-not [IO.Path]::IsPathFullyQualified($Path) -or $Path.Contains("'")){throw 'SE absolute safe path required'}
 $full=[IO.Path]::GetFullPath($Path);$base=Get-SEBase
 if(-not $full.StartsWith($base+'\',[StringComparison]::OrdinalIgnoreCase) -and -not $full.Equals($base,[StringComparison]::OrdinalIgnoreCase)){throw 'SE outside fixture workspace'}
 $cursor=$full
 while($cursor){
  if(Test-Path -LiteralPath $cursor){if((Get-Item -LiteralPath $cursor -Force).Attributes -band [IO.FileAttributes]::ReparsePoint){throw 'SE reparse path'}}
  $parent=[IO.Path]::GetDirectoryName($cursor);if($parent -eq $cursor){break};$cursor=$parent
 }
 return $full
}
function ConvertTo-SEFixtureVerifier([string]$Source,[string]$FixtureRoot,[string]$ReviewRelative){
 $root=Assert-SEPath $FixtureRoot
 if($ReviewRelative -cne 'fixture-review'){throw 'SE review path'}
 $digest=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.UTF8Encoding]::new($false).GetBytes($Source)))
 if($digest -cne '688CDA16DE6BF69969FACB079058E5DD7C560039C146F9A245FBDB983462077E'){throw 'SE unexpected verifier source'}
 $oldRoot="`$root='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'"
 $oldReview="`$review='.superpowers/sdd/2026-11-05-runtime-staged-integrity-review-v2'"
 $result=$Source.Replace($oldRoot,"`$root='$root'").Replace($oldReview,"`$review='$ReviewRelative'")
 if($result -ceq $Source -or $result.Contains($oldRoot) -or $result.Contains($oldReview)){throw 'SE rebind anchors'}
 if($result.Replace("`$root='$root'",$oldRoot).Replace("`$review='$ReviewRelative'",$oldReview) -cne $Source){throw 'SE rebind delta'}
 return $result
}
function Invoke-SEFixture([string]$Script,[string]$FixtureRoot,[string]$InputSHA256){
 $root=Assert-SEPath $FixtureRoot
 if(-not(Test-Path -LiteralPath $root -PathType Container)){throw 'SE missing fixture root'}
 $ps=[PowerShell]::Create();$oldDirectory=[Environment]::CurrentDirectory;$caught=$null;$output=@();$terminal=$false
 try{
  [Environment]::CurrentDirectory=$root
  $null=$ps.AddCommand('Set-Location').AddParameter('LiteralPath',$root);$null=$ps.Invoke()
  if($ps.HadErrors){throw 'SE runspace location failed'}
  $ps.Commands.Clear();$ps.Streams.Error.Clear()
  $null=$ps.AddScript($Script).AddParameter('InputSHA256',$InputSHA256)
  try{$output=@($ps.Invoke()|ForEach-Object {$_|Out-String})}catch{$caught=$_.Exception.Message}
  $state=$ps.InvocationStateInfo.State.ToString();$terminal=$state -in @('Completed','Failed','Stopped')
  $errors=@($ps.Streams.Error|ForEach-Object {$_.ToString()});if($null -ne $caught){$errors+=,$caught}
  return [ordered]@{Terminal=$terminal;HadErrors=($ps.HadErrors -or $null -ne $caught);Output=@($output|ForEach-Object {$_.TrimEnd()});Errors=$errors;State=$state}
 }finally{
  if($ps.InvocationStateInfo.State.ToString() -in @('NotStarted','Completed','Failed','Stopped')){$ps.Dispose();[Environment]::CurrentDirectory=$oldDirectory}
  else{$global:SERetainedPowerShell=$ps;throw 'SE pending runspace retained; stop all fixture work'}
 }
}
function Write-SEBytes([string]$Path,[byte[]]$Bytes){
 $full=Assert-SEPath $Path
 $null=[IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($full))
 $null=Assert-SEPath $full
 $stream=[IO.File]::Open($full,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
 try{$stream.Write($Bytes,0,$Bytes.Length);$stream.Flush()}finally{$stream.Dispose()}
}
function New-SEFixture([string]$CaseId,[string]$Workspace){
 if($CaseId -cnotmatch '^C(0[1-9]|1[0-6])$'){throw 'SE unknown case'}
 $workspaceFull=Assert-SEPath $Workspace;$root=Assert-SEPath (Join-Path $workspaceFull $CaseId)
 if(Test-Path -LiteralPath $root){throw 'SE occupied case'}
 $repo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator';$l='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2';$d='.superpowers/sdd/2026-11-05-runtime-staged-integrity-v1'
 $sets=@(@('portable-runtime-staged-integrity-observation-final.json','EEBC1498D8DD1760D36B37CB79EB7C56199AF5B62AFFBB8F8D501D17033DC905'),@('portable-runtime-staged-integrity-review-v2.json','5943F342A08B6ADF65A1605267876905F2A3E0934BCA31B9BDEB4BE3ADD1AB87'))
 $entries=@{}
 foreach($set in $sets){
  $p=Join-Path $repo "$l/$($set[0])"
  if((Get-FileHash -LiteralPath $p).Hash -cne $set[1]){throw 'SE archive identity'}
  $archive=Get-Content -LiteralPath $p -Raw|ConvertFrom-Json -AsHashtable
  foreach($entry in $archive.Files){if($entries.ContainsKey($entry.Path)){throw 'SE duplicate archive path'};$entries[$entry.Path]=$entry}
 }
 $names=@('input.json','intent.json','supervision.json','integrity-review.json','supervisor.before-map.json','supervisor.after-map.json','writer.before-map.json','writer.role-result.json','writer.claim.json','producer.claim.json','child.claim.json','child-result.json','result.json','stages.jsonl','journal.jsonl','data/stdout.bin','data/stderr.bin')
 $selected=@($names|ForEach-Object {"$d/$_"})+@("$l/runtime-staged-integrity.ps1")
 $inputBytes=ConvertFrom-SEPayload $entries["$d/input.json"]
 $binding=[Text.UTF8Encoding]::new($false,$true).GetString($inputBytes)|ConvertFrom-Json -AsHashtable
 $priorPath='.superpowers/sdd/2026-11-04-runtime-copy-transport-v1/stages.jsonl'
 $priorArchive=Get-Content (Join-Path $repo "$l/portable-runtime-copy-transport-final.json") -Raw|ConvertFrom-Json -AsHashtable
 $priorEntries=@($priorArchive.Files|Where-Object {$_.Path -ceq $priorPath})
 if($priorEntries.Count -ne 1 -or $priorEntries[0].SHA256 -cne $binding.HistoricalEvidenceSHA256[$priorPath]){throw 'SE prior stages pin'}
 $entries[$priorPath]=$priorEntries[0];$selected+=,$priorPath
 $null=[IO.Directory]::CreateDirectory($root)
 foreach($path in $selected){
  if([IO.Path]::IsPathFullyQualified($path) -or ($path -split '[/\\]') -contains '..'){throw 'SE unsafe archive entry'}
  $entry=$entries[$path];$bytes=ConvertFrom-SEPayload $entry
  if($bytes.Length -ne $entry.Length -or [Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($bytes)) -cne $entry.SHA256){throw 'SE archive bytes'}
  Write-SEBytes (Join-Path $root $path) $bytes
 }
 $null=[IO.Directory]::CreateDirectory((Join-Path $root 'fixture-review'))
 return [ordered]@{Root=$root;ReviewRelative='fixture-review';InputSHA256=$entries["$d/input.json"].SHA256}
}
