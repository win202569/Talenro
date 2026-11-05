Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
$d='.superpowers/sdd/2026-11-05-runtime-writer-phase-timing-v1'
function Need([bool]$Ok,[string]$Why){if(-not $Ok){throw "WPT_STATIC: $Why"}}
function Read([string]$Name){$p=Join-Path $PSScriptRoot $Name;Need (Test-Path $p) "missing $Name";return (Get-Content $p -Raw).Replace("`r`n","`n").TrimEnd()}
$role=Read runtime-writer-phase-timing-role.ps1
$driver=Read runtime-writer-phase-timing-driver.ps1
$capture=Read runtime-writer-phase-timing-capture.ps1
$common=Read runtime-writer-phase-timing-common.ps1
$old='.superpowers/sdd/2026-11-05-runtime-isolated-writer-business-prefix-v1'
Need ($capture -ceq (Read runtime-isolated-writer-business-capture.ps1).Replace($old,$d).Replace('runtime-isolated-writer-business-common.ps1','runtime-writer-phase-timing-common.ps1')) 'capture binding only'
Need ($common -ceq (Read runtime-isolated-writer-business-common.ps1).Replace('$intent.Computation -ne 42','$intent.Computation -ne 43').Replace('$intent.Ceiling -ne 42','$intent.Ceiling -ne 43')) 'common budget only'
Need ($driver -ceq (Read runtime-isolated-writer-business-driver.ps1).Replace($old,$d).Replace('runtime-isolated-writer-business-role.ps1','runtime-writer-phase-timing-role.ps1').Replace('BusinessPrefix42','WriterPhase43').Replace('Computation=42','Computation=43').Replace('Ceiling=42','Ceiling=43')) 'driver binding only'
Need ([regex]::Matches($driver,'Supervisor\]::RunAsync\(\$options\)').Count -eq 1) 'one supervisor call'
Need ($role.Contains("if(`$Role -ceq 'Producer')") -and $role.Contains('exit $LASTEXITCODE') -and -not $role.Contains('Producer]::RunAsync')) 'same process business'
Need ($role.Contains("(`$Role -ceq 'Writer' -and `$Name -ceq 'writer.after-map.json')")) 'after-map only'
foreach($phase in @('task.returned','pins.begin','pins.end','map.save.begin','map.save.end','map.meta.end','role.save.begin','role.save.end','exit.ready')){Need ([regex]::Matches($role,('Emit-WPT '+[regex]::Escape($phase)+'\b')).Count -eq 1) "one phase $phase"}
Need ($role.Contains('finally{') -and $role.Contains('$row.After=Map-IW writer.after-map.json') -and $role.Contains('$map=Assert-HTInputs $binding') -and $role.Contains('Save-IW $Name $map')) 'original checks retained'
Need ($role.Contains('if($null -ne $script:WPTState -and $script:WPTState.Failed){$row.ExitCode=1}') -and $role.LastIndexOf('$script:WPTState.Failed') -gt $role.IndexOf('Emit-WPT exit.ready')) 'late trace failure nonzero'
Need ($role.Contains('$script:WPTClock=[Diagnostics.Stopwatch]::StartNew()') -and $role.Contains('[Console]::OpenStandardError()')) 'clock/diagnostic source'
$verify=Get-Content (Join-Path $repo "$d/verify.ps1") -Raw
Need ($verify.Contains('Read-WPTTrace') -and $verify.Contains('$r.DiagnosticBase64') -and $verify.Contains('timing-review.json')) 'saved trace parser'
Need ($verify.Contains('TIMED_PREFIX_FAILED') -and $verify.IndexOf('TIMED_PREFIX_FAILED') -lt $verify.IndexOf('[IO.File]::ReadAllBytes($path)')) 'failure branch before data'
Need ($verify.Contains('$n=@(48,0)[$i]') -and -not $verify.Contains('$diagnostic.Length -eq 0')) 'business stderr remains zero, diagnostics separate'
Need ($verify.Contains("Need (`$timing.Status -ceq 'COMPLETE')")) 'complete timing acceptance'
$paths=@(Get-ChildItem $PSScriptRoot -Filter 'runtime-writer-phase-timing*.ps1' -File|ForEach-Object FullName)
$paths+=@("$d/prepare.ps1","$d/verify.ps1","$d/preserve.ps1"|ForEach-Object {Join-Path $repo $_})
foreach($p in $paths){
 $tokens=$null;$errors=$null;$ast=[Management.Automation.Language.Parser]::ParseFile($p,[ref]$tokens,[ref]$errors)
 Need ($errors.Count -eq 0) "parse $p"
 $bad=@($ast.FindAll({param($a)$a -is [Management.Automation.Language.CommandAst] -and $a.GetCommandName() -cin @('Add-Type','Start-Process','Start-Job','Invoke-Expression','pwsh','powershell')},$true))
 Need ($bad.Count -eq 0) "no alternate launch $p"
}
[Console]::Out.WriteLine('WRITER_PHASE_STATIC_VERIFIED')
