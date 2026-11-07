param([Parameter(Mandatory)][string]$InputPath,[ValidateSet('complete','bad','cutoff')][string]$Mode)
$ErrorActionPreference='Stop'
$payload=[IO.File]::ReadAllText($InputPath)
if($Mode -eq 'complete'){
    # Split inside the first JSON record; no newline until its remainder arrives.
    [Console]::Out.Write($payload.Substring(0,20));[Console]::Out.Flush()
    Start-Sleep -Milliseconds 100
    [Console]::Out.Write($payload.Substring(20));[Console]::Out.Flush()
    exit 0
}
[Console]::Out.Write($payload);[Console]::Out.Flush()
Start-Sleep -Seconds 30
exit 0
