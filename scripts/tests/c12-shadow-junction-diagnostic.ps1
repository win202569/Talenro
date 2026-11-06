# One authorized diagnostic attempt; not the parser suite or acceptance gate.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=Split-Path (Split-Path $PSScriptRoot)
$root=Join-Path $repo '.superpowers/sdd/2026-11-06-c12-shadow-junction-diagnostic-v1'
if(Test-Path -LiteralPath $root){throw 'Diagnostic root occupied; no retry'}
$null=[IO.Directory]::CreateDirectory($root)
function Save-Diagnostic([string]$Path,$Value){
 $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 12))
 $stream=[IO.File]::Open($Path,[IO.FileMode]::CreateNew)
 try{$stream.Write($bytes,0,$bytes.Length)}finally{$stream.Dispose()}
}
Save-Diagnostic (Join-Path $root 'claim.json') @{Attempt=1;MaximumAttempts=1;Scope='junction creation and conditional module rejection';OriginalSuiteRerun=$false}
$target=Join-Path $root 'target'
$junction=Join-Path $root 'link'
$result=[ordered]@{Schema=1;Overall='NOT_ACCEPTED';PowerShell=$PSVersionTable.PSVersion.ToString();Stage='prepare';CreationSucceeded=$false;ModuleEvaluated=$false;RejectionMatched=$false;OriginalError=$null;LinkMetadata=$null;ModuleSHA256=(Get-FileHash -LiteralPath (Join-Path $repo 'scripts/private/c12-shadow-audit.ps1')).Hash}
try{
 $null=[IO.Directory]::CreateDirectory($target)
 $result.Stage='create'
 $null=New-Item -ItemType Junction -Path $junction -Target $target
 $result.CreationSucceeded=$true
 $item=Get-Item -LiteralPath $junction -Force
 $result.LinkMetadata=@{Attributes=$item.Attributes.ToString();LinkType=$item.LinkType;Target=@($item.Target)}
 if(-not ($item.Attributes -band [IO.FileAttributes]::ReparsePoint)){throw 'Created entry lacks reparse attribute'}
 $result.Stage='module'
 . (Join-Path $repo 'scripts/private/c12-shadow-audit.ps1')
 $manifest=Read-C12ShadowManifest $junction
 $result.ModuleEvaluated=$true
 $result.RejectionMatched=($manifest.State -ceq 'INCOMPLETE' -and $manifest.Reason -ceq 'UNSAFE_PATH')
 $result.ModuleResult=@{State=$manifest.State;Reason=$manifest.Reason}
 if(-not $result.RejectionMatched){throw 'Unexpected module rejection result'}
 $result.Stage='diagnostic-complete'
}catch{
 $result.OriginalError=[ordered]@{Type=$_.Exception.GetType().FullName;Message=$_.Exception.Message;HResult=$_.Exception.HResult;FullyQualifiedErrorId=$_.FullyQualifiedErrorId;Category=$_.CategoryInfo.ToString();Stack=$_.ScriptStackTrace;InnerMessage=$(if($null -ne $_.Exception.InnerException){$_.Exception.InnerException.Message}else{$null})}
}
Save-Diagnostic (Join-Path $root 'result.json') $result
$result|ConvertTo-Json -Depth 12
if($null -ne $result.OriginalError){exit 1}
