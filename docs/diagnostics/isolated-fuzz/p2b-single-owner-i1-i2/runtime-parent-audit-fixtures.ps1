# Definition-only utilities. Never execute the original operational driver entry.
function Get-PABase{return 'D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator\.superpowers\sdd\2026-11-05-parent-audit-persistence-v1'}
function Assert-PAPath([string]$Path){
 if(-not [IO.Path]::IsPathFullyQualified($Path) -or $Path.Contains("'")){throw 'PA absolute safe path required'}
 $full=[IO.Path]::GetFullPath($Path);$base=Get-PABase
 if(-not $full.Equals($base,[StringComparison]::OrdinalIgnoreCase) -and -not $full.StartsWith($base+'\',[StringComparison]::OrdinalIgnoreCase)){throw 'PA outside workspace'}
 $cursor=$full
 while($cursor){
  if(Test-Path -LiteralPath $cursor){if((Get-Item -LiteralPath $cursor -Force).Attributes -band [IO.FileAttributes]::ReparsePoint){throw 'PA reparse path'}}
  $parent=[IO.Path]::GetDirectoryName($cursor);if($parent -eq $cursor){break};$cursor=$parent
 }
 return $full
}
function Write-PABytes([string]$Path,[byte[]]$Bytes){
 $full=Assert-PAPath $Path;$null=[IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($full));$null=Assert-PAPath $full
 $s=[IO.File]::Open($full,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
 try{$s.Write($Bytes,0,$Bytes.Length);$s.Flush()}finally{$s.Dispose()}
}
function Save-PA([string]$Path,$Value){Write-PABytes $Path ([Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 80)))}
function Get-ParentAuditSlice([string]$DriverSource,[string]$CommonSource){
 $defs=[Collections.Generic.List[string]]::new();$driverAst=$null
 foreach($pair in @(@($DriverSource,'5031ED56375B5DB1194567E6D7D8B95405DD36595324D9BC6B4AED7A89BFECDC'),@($CommonSource,'EF5517999303FC3903F75EFDD7A952F890CA3A4FF7137A1BADBCFC45B2550B76'))){
  if([Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.UTF8Encoding]::new($false).GetBytes($pair[0]))) -cne $pair[1]){throw 'PA source identity'}
  $tokens=$null;$errors=$null;$ast=[Management.Automation.Language.Parser]::ParseInput($pair[0],[ref]$tokens,[ref]$errors)
  if($errors.Count){throw 'PA source syntax'}
  $names=if($null -eq $driverAst){$driverAst=$ast;@('Need-IW','Save-IW','Map-IW','Journal-IW')}else{@('Get-HTHash','Assert-HTInputs')}
  foreach($name in $names){
   $nodes=@($ast.EndBlock.Statements|Where-Object {$_ -is [Management.Automation.Language.FunctionDefinitionAst] -and $_.Name -ceq $name})
   if($nodes.Count -ne 1){throw 'PA definition ambiguity'}
   $defs.Add($nodes[0].Extent.Text)
  }
 }
 $outer=@($driverAst.EndBlock.Statements|Where-Object {$_ -is [Management.Automation.Language.TryStatementAst]})
 if($outer.Count -ne 1 -or $outer[0].CatchClauses.Count -ne 1){throw 'PA outer structure'}
 $first=@($outer[0].Body.Statements|Where-Object {$_.Extent.Text -ceq 'Save-IW supervision.json $retained'})
 $last=$outer[0].Body.Statements[-1]
 if($first.Count -ne 1 -or $last.Extent.Text -cne "[Console]::Out.WriteLine('BUSINESS_PREFIX_TRANSPORT_COMPLETED_REVIEW_REQUIRED')"){throw 'PA tail anchors'}
 $tail=$DriverSource.Substring($first[0].Extent.StartOffset,$last.Extent.EndOffset-$first[0].Extent.StartOffset)
 $slice=($defs -join "`n")+"`ntry{`n"+$tail+"`n}"+$outer[0].CatchClauses[0].Extent.Text
 $tokens=$null;$errors=$null;$check=[Management.Automation.Language.Parser]::ParseInput($slice,[ref]$tokens,[ref]$errors)
 if($errors.Count){throw 'PA slice syntax'}
 $allowed=@('ConvertTo-Json','Need-IW','Join-Path','Assert-HTInputs','Save-IW','Get-Item','Get-FileHash','Test-Path','Get-HTHash','Journal-IW','Map-IW')
 foreach($node in $check.FindAll({param($n) $n -is [Management.Automation.Language.CommandAst]},$true)){if($node.GetCommandName() -cnotin $allowed){throw 'PA unexpected command'}}
 return $slice
}
function Invoke-ParentAuditCase([string]$CaseId,[string]$Workspace,[string]$Slice){
 if($CaseId -cnotmatch '^P(0[1-9]|10)$'){throw 'PA unknown case'}
 $ws=Assert-PAPath $Workspace;$root=Assert-PAPath (Join-Path $ws $CaseId)
 if(Test-Path -LiteralPath $root){throw 'PA occupied case'}
 $null=[IO.Directory]::CreateDirectory($root);$utf8=[Text.UTF8Encoding]::new($false)
 $baseline=$utf8.GetBytes('fixture');$hash=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($baseline))
 Write-PABytes (Join-Path $root source.baseline) $baseline
 $actual=if($CaseId -in @('P02','P08','P09')){$utf8.GetBytes('drift')}else{$baseline}
 Write-PABytes (Join-Path $root source.txt) $actual
 if($CaseId -ceq 'P03'){Write-PABytes (Join-Path $root absent.txt) ($utf8.GetBytes('present'))}
 $sentinels=@{P04='supervisor.after-map.json';P05='integrity-review.json';P06='result.json';P07='supervision.json'}
 if($sentinels.ContainsKey($CaseId)){Write-PABytes (Join-Path $root $sentinels[$CaseId]) ($utf8.GetBytes('SENTINEL-NOT-EVIDENCE'))}
 $retained=[ordered]@{NextAllowed=$true;PendingResources=($CaseId -ceq 'P08');ExitUnconfirmed=($CaseId -ceq 'P09');SupervisionOutcome=$(if($CaseId -ceq 'P10'){'FAILED'}else{'SUCCESS'});Writer=@{OwnedExit=$true;ExitCode=0};Producer=@{OwnedExit=$true;ExitCode=0}}
 $binding=@{SourceSHA256=@{'source.txt'=$hash};PreservedSHA256=@{};HistoricalEvidenceSHA256=@{};HelperSHA256=@{};AbsentPaths=@('absent.txt')}
 $context=@{Synthetic=$true;Case=$CaseId;Root=$root;InputSHA256=('A'*64);Options=@{IntentSHA256=('B'*64)};Before=@{Synthetic=$true};Retained=$retained;Binding=$binding}
 Save-PA (Join-Path $root context.json) $context
 Write-PABytes (Join-Path $root slice.ps1) ($utf8.GetBytes($Slice))
 Write-PABytes (Join-Path $root journal.jsonl) ($utf8.GetBytes('{"Seq":1,"Kind":"begin","Case":"StagedIntegrity44","Synthetic":true}'+"`n"))
 $setup='param($Context) Set-StrictMode -Version Latest; $ErrorActionPreference="Stop"; $script:HTRepo=$Context.Root; $base=$Context.Root; $folder=$Context.Root; $relative="."; $part=""; $InputSHA256=$Context.InputSHA256; $options=$Context.Options; $before=$Context.Before; $retained=$Context.Retained; $binding=$Context.Binding;'
 Write-PABytes (Join-Path $root setup.ps1) ($utf8.GetBytes($setup))
 $ps=[PowerShell]::Create();$old=[Environment]::CurrentDirectory;$output=@();$caught=$null
 try{
  [Environment]::CurrentDirectory=$root
  $null=$ps.AddCommand('Set-Location').AddParameter('LiteralPath',$root);$null=$ps.Invoke()
  if($ps.HadErrors){throw 'PA location failed'}
  $ps.Commands.Clear();$ps.Streams.Error.Clear()
  $null=$ps.AddScript($setup+"`n"+$Slice).AddParameter('Context',$context)
  try{$output=@($ps.Invoke()|ForEach-Object {($_|Out-String).TrimEnd()})}catch{$caught=$_.Exception.Message}
  $state=$ps.InvocationStateInfo.State.ToString();$terminal=$state -in @('Completed','Failed','Stopped')
  $errors=@($ps.Streams.Error|ForEach-Object {$_.ToString()});if($null -ne $caught){$errors+=,$caught}
  $run=[ordered]@{Case=$CaseId;Root=$root;Terminal=$terminal;State=$state;HadErrors=($ps.HadErrors -or $null -ne $caught);Output=$output;Errors=$errors;FileHashes=[ordered]@{}}
  if(-not $terminal){throw 'PA pending runspace'}
  foreach($file in Get-ChildItem -LiteralPath $root -File){$null=Assert-PAPath $file.FullName;$run.FileHashes[$file.Name]=(Get-FileHash -LiteralPath $file.FullName).Hash}
  Save-PA (Join-Path $root invocation.json) $run
  return $run
 }finally{
  if($ps.InvocationStateInfo.State.ToString() -in @('NotStarted','Completed','Failed','Stopped')){$ps.Dispose();[Environment]::CurrentDirectory=$old}
  else{$global:PARetainedPowerShell=$ps;throw 'PA active runspace retained; stop'}
 }
}
