Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
. (Join-Path $repo 'scripts/offline-diagnostics-v4/private/write-boundary.ps1')
$tokens=$null;$parseErrors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot capture.ps1),[ref]$tokens,[ref]$parseErrors)
$collector=@($ast.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'Capture-Failure'},$false))
if($parseErrors.Count -ne 0 -or $collector.Count -ne 1){throw 'Publisher fixed collector source'}
. ([ScriptBlock]::Create($collector[0].Extent.Text))
$script:HostPrimary=$null
$script:HostSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new()
$script:HostOverflow=$false
$script:DefaultFactory=(Get-Command New-ODRecordStream).ScriptBlock
Export-ModuleMember
