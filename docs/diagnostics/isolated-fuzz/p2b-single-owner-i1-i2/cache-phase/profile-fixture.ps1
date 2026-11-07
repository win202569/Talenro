param([Parameter(Mandatory)][string]$InputPath,[ValidateSet('complete','wait')][string]$Mode)
$ErrorActionPreference='Stop'
$b=[IO.File]::ReadAllBytes($InputPath);$s=[Console]::OpenStandardOutput();$s.Write($b,0,$b.Length);$s.Flush()
if($Mode -ceq 'wait'){Start-Sleep -Seconds 30}
