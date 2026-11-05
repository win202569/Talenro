Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
$d='.superpowers/sdd/2026-11-05-runtime-staged-integrity-v1'
$old='.superpowers/sdd/2026-11-05-runtime-writer-phase-timing-v1'
function Need([bool]$Ok,[string]$Why){if(-not $Ok){throw "SI_STATIC: $Why"}}
function Read([string]$Name){$p=Join-Path $PSScriptRoot $Name;Need (Test-Path $p) "missing $Name";return (Get-Content $p -Raw).Replace("`r`n","`n").TrimEnd()}
$role=Read runtime-staged-integrity-role.ps1
$driver=Read runtime-staged-integrity-driver.ps1
$capture=Read runtime-staged-integrity-capture.ps1
$common=Read runtime-staged-integrity-common.ps1
Need ($capture -ceq (Read runtime-writer-phase-timing-capture.ps1).Replace($old,$d).Replace('runtime-writer-phase-timing-common.ps1','runtime-staged-integrity-common.ps1')) 'producer capture binding only'
Need ($common -ceq (Read runtime-writer-phase-timing-common.ps1).Replace('$intent.Computation -ne 43','$intent.Computation -ne 44').Replace('$intent.Ceiling -ne 43','$intent.Ceiling -ne 44')) 'producer common budget only'
$oldRole=Read runtime-writer-phase-timing-role.ps1
$producerPattern="(?s) if\(\`$Role -ceq 'Producer'\)\{.*?exit \`$LASTEXITCODE\n }"
# Extract producer branch with fixed delimiters, avoiding parent/child output-order assumptions.
$start=$oldRole.IndexOf(" if(`$Role -ceq 'Producer'){");$end=$oldRole.IndexOf(' $row.Before=',$start)
$producer=$oldRole.Substring($start,$end-$start).Replace('runtime-writer-phase-timing-capture.ps1','runtime-staged-integrity-capture.ps1')
Need ($role.Contains($producer)) 'unchanged producer branch'
Need ($role.Contains('$row.Before=Map-IW writer.before-map.json') -and $role.Contains('$map=Assert-HTInputs $binding')) 'writer before full pass'
Need (-not $role.Contains('writer.after-map.json') -and $role.Contains("After=`$null;AfterVerificationMode='PARENT_POST_EXIT_REQUIRED'")) 'explicit writer delegation only'
foreach($p in @('task.returned','role.save.begin','role.save.end','exit.ready')){Need ([regex]::Matches($role,('Emit-SI '+[regex]::Escape($p)+'\b')).Count -eq 1) "phase $p"}
Need ($role.LastIndexOf('$script:SIState.Failed') -gt $role.IndexOf('Emit-SI exit.ready') -and $role.EndsWith('exit $row.ExitCode')) 'late failure exit'
Need ([regex]::Matches($driver,'Supervisor\]::RunAsync\(\$options\)').Count -eq 1) 'single supervisor'
Need ([regex]::Matches($driver,'Map-IW supervisor.after-map.json').Count -eq 1) 'one parent after pass'
Need ($driver.IndexOf('Save-IW supervision.json') -lt $driver.IndexOf('Need-IW $terminal') -and $driver.IndexOf('Need-IW $terminal') -lt $driver.IndexOf('Map-IW supervisor.after-map.json')) 'raw save and terminal before audit'
Need ($driver.Contains("`$audit.After=`$null;`$audit.State='FAILED'") -and $driver.Contains('Save-IW integrity-review.json $audit') -and $driver.Contains("Need-IW (`$audit.State -ceq 'VERIFIED')")) 'persisted failed audit blocks success'
Need ($driver.IndexOf('Save-IW integrity-review.json $audit') -lt $driver.IndexOf("Need-IW (`$audit.State -ceq 'VERIFIED')")) 'persist before success'
$verify=(Get-Content "$d/verify.ps1" -Raw).Replace("`r`n","`n")
Need (-not $verify.Contains('Get-FileHash $p') -and -not $verify.Contains('Assert-HTInputs')) 'no duplicate full source pass'
foreach($key in @('TransportVerified','TerminalVerified','BusinessVerified','AuditPersisted','AuditIdentityVerified','AuditMapsVerified','WriterDelegationVerified','RoleVerified','TimingVerified','OutputVerified')){Need ($verify.Contains("`$facts.$key=`$true")) "checked fact $key"}
foreach($text in @('integrity-review.json','audit identity','map path','map metadata','map count','PARENT_POST_EXIT_REQUIRED','Test-SIObservation','Read-SITrace','ObservationPassed','NOT_ACCEPTED')){Need ($verify.Contains($text)) "verifier requirement $text"}
Need ($verify.Contains('$n=@(48,0)[$i]') -and $verify.Contains('$diagnostic.Length -le 1920')) 'separate output/trace limits'
Need ($verify.IndexOf("if(`$r.SupervisionOutcome -cne 'SUCCESS')") -lt $verify.IndexOf('[IO.File]::ReadAllBytes($path)')) 'failure refuses output'
Need ($verify.Contains("Need (`$timing.Status -ceq 'COMPLETE')") -and $verify.Contains('$p.ExitCode -eq 0') -and $verify.Contains('$writer.ExitCode -eq 0')) 'trace and real exits'
Need ($verify.Contains('$audit.InputSHA256 -ceq $InputSHA256') -and $verify.Contains('$audit.IntentSHA256 -ceq $intentHash')) 'audit bound to attempt'
Need ((Get-FileHash '.superpowers/sdd/2026-11-05-runtime-isolated-writer-v3/isolated-writer.dll').Hash -ceq '2AEF2A39B5D499981378D8FD1F749A3073AD92C03641CEBBA28E9060B3FFCD7B') 'unchanged supervisor/budgets'
Need ((Get-FileHash '.superpowers/sdd/2026-11-04-runtime-bounded-disk-adapter-v1/adapter.dll').Hash -ceq '4C1470B94B3DEAA87635F1375A05161BE0E1F4C4D2CF65595214A8E39CDBCCDA') 'unchanged adapter'
$archive=Get-Content (Join-Path $PSScriptRoot portable-runtime-writer-phase-timing-final.json) -Raw|ConvertFrom-Json -AsHashtable
foreach($entry in $archive.Files){if($entry.Path.EndsWith('.ps1')){Need ((Get-FileHash -LiteralPath $entry.Path).Hash -ceq $entry.SHA256) 'old scripts immutable'}}
$paths=@(Get-ChildItem $PSScriptRoot -Filter 'runtime-staged-integrity*.ps1' -File|ForEach-Object FullName)
$paths+=@("$d/prepare.ps1","$d/verify.ps1","$d/preserve.ps1"|ForEach-Object {Join-Path $repo $_})
foreach($p in $paths){
 $tokens=$null;$errors=$null;$ast=[Management.Automation.Language.Parser]::ParseFile($p,[ref]$tokens,[ref]$errors)
 Need ($errors.Count -eq 0) "parse $p"
 $bad=@($ast.FindAll({param($a)$a -is [Management.Automation.Language.CommandAst] -and $a.GetCommandName() -cin @('Add-Type','Start-Process','Start-Job','Invoke-Expression','pwsh','powershell')},$true))
 Need ($bad.Count -eq 0) "no alternate launch $p"
}
[Console]::Out.WriteLine('STAGED_INTEGRITY_STATIC_VERIFIED')
