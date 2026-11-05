# Parse only the predicate definition: never execute the full static suite here.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$tokens=$null;$errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot runtime-staged-integrity-static.tests.ps1),[ref]$tokens,[ref]$errors)
if($errors.Count){throw 'SI_PREDICATE: invalid checker syntax'}
$definitions=@($ast.FindAll({param($a)$a -is [Management.Automation.Language.FunctionDefinitionAst] -and $a.Name -ceq 'Test-SINoSourceRescan'},$true))
if($definitions.Count -ne 1){throw 'SI_PREDICATE: missing exact predicate'}
. ([scriptblock]::Create($definitions[0].Extent.Text))
# Literal fixtures catch prefix matching, ignored comments, named arguments,
# command-name casing, nested arguments, and fail-open parse errors.
$fixtures=@(
 @{Text='Get-FileHash $path';Allowed=$true},
 @{Text='Get-FileHash $p';Allowed=$false},
 @{Text='Get-FileHash -LiteralPath ${P}';Allowed=$false},
 @{Text='get-filehash -Path ($p)';Allowed=$false},
 @{Text='Get-FileHash $ref.Path';Allowed=$true},
 @{Text='# Get-FileHash $p';Allowed=$true},
 @{Text='"Get-FileHash $p"';Allowed=$true},
 @{Text='Assert-HTInputs $binding';Allowed=$false},
 @{Text='Get-FileHash $script:p';Allowed=$false}
)
foreach($fixture in $fixtures){if((Test-SINoSourceRescan $fixture.Text) -ne $fixture.Allowed){throw "SI_PREDICATE: wrong verdict for $($fixture.Text)"}}
$rejected=$false
try{$null=Test-SINoSourceRescan 'Get-FileHash ('}catch{$rejected=$true}
if(-not $rejected){throw 'SI_PREDICATE: parse error accepted'}
[Console]::Out.WriteLine('STAGED_STATIC_PREDICATE_VERIFIED: 10 cases')
