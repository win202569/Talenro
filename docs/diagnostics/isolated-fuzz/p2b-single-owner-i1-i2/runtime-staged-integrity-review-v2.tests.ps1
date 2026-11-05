param([Parameter(Mandatory)][string]$VerifierPath)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$tokens=$null;$errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile($VerifierPath,[ref]$tokens,[ref]$errors)
if($errors.Count){throw 'REVIEW_TEST: parser errors'}
$loops=@($ast.FindAll({param($a)$a -is [Management.Automation.Language.ForEachStatementAst] -and $a.Extent.Text.Contains("'exact business bytes'")},$true))
if($loops.Count -ne 1){throw 'REVIEW_TEST: output loop identity'}
$statements=@($loops[0].Body.Statements|Where-Object {$_.Extent.Text -match '^\s*(\$expected=|if\(\$i -eq 0\))'})
if($statements.Count -ne 1){throw 'REVIEW_TEST: expected-value statement identity'}
$statement=[scriptblock]::Create($statements[0].Extent.Text)
$strict=[Text.UTF8Encoding]::new($false,$true)
$expected=$null;$i=1
. $statement
if($null -eq $expected -or $expected -isnot [byte[]] -or $expected.Length -ne 0){throw 'REVIEW_TEST: empty stderr must be non-null byte[]'}
if(-not [Linq.Enumerable]::SequenceEqual([byte[]]::new(0),[byte[]]$expected)){throw 'REVIEW_TEST: empty equality'}
if([Linq.Enumerable]::SequenceEqual([byte[]]@(1),[byte[]]$expected)){throw 'REVIEW_TEST: nonempty stderr accepted'}
$i=0
. $statement
if($expected -isnot [byte[]] -or $expected.Length -ne 48 -or $strict.GetString($expected) -cne "PASS: ContractAPIAvailable`r`nPASS: BaselineCore`r`n"){throw 'REVIEW_TEST: stdout contract'}
[Console]::Out.WriteLine('STAGED_REVIEW_EMPTY_ARRAY_VERIFIED: 4 assertions')
