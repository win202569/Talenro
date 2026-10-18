# Static preparation fidelity only: never imports or invokes target sources.
$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$old=Join-Path $PSScriptRoot 'integer-candidate-three-phase'
$new=Join-Path $PSScriptRoot 'integer-candidate-ownership-markers'
$count=0
function Need([bool]$Value,[string]$Name){if(-not $Value){throw ('FIDELITY FAILED: '+$Name)}}
function Case([string]$CaseLabel,[scriptblock]$Body){& $Body;$script:count++;Write-Output ('PASS: OwnershipPreparation/'+$CaseLabel)}
$sourceNames=@(
    'offline-supervision-policy.ps1','phase-evaluator.ps1',
    'offline-supervision-policy.tests.ps1','offline-supervision-policy.ownership.tests.ps1',
    'offline-supervision-policy.reflection-reuse.tests.ps1','offline-supervision-policy.reflection-reuse-regression.ps1',
    'offline-supervision-policy.shape-lookup.tests.ps1','offline-supervision-policy.shape-lookup-regression.ps1'
)
foreach($name in $sourceNames){Case ('Parse/'+$name) {
    $tokens=$null;$errors=$null
    $null=[Management.Automation.Language.Parser]::ParseFile((Join-Path $new $name),[ref]$tokens,[ref]$errors)
    Need ($errors.Count -eq 0) ('AST/'+$name)
}}
foreach($name in $sourceNames){if($name -cne 'offline-supervision-policy.ownership.tests.ps1'){Case ('Identical/'+$name) {
    Need ((Get-FileHash -LiteralPath (Join-Path $old $name)).Hash -ceq (Get-FileHash -LiteralPath (Join-Path $new $name)).Hash) ('Bytes/'+$name)
}}}
$ownershipPath=Join-Path $new 'offline-supervision-policy.ownership.tests.ps1'
$text=[IO.File]::ReadAllText($ownershipPath)
Case FourMarkersRestoreEntireOriginal {
    $lines=@($text -split '\n')
    $marks=@($lines|Where-Object {$_ -cmatch '^        \[Console\]::Out\.WriteLine\(''OWNERSHIP: '})
    Need ($marks.Count -eq 4) 'ExactlyFour'
    $restored=(@($lines|Where-Object {$_ -cnotmatch '^        \[Console\]::Out\.WriteLine\(''OWNERSHIP: '}) -join "`n")
    Need ($restored -ceq [IO.File]::ReadAllText((Join-Path $old 'offline-supervision-policy.ownership.tests.ps1'))) 'WholeSourceRestored'
    Need ((Get-FileHash -LiteralPath (Join-Path $old 'offline-supervision-policy.ownership.tests.ps1')).Hash -ceq 'C4905925B309B00E94834B2FDA3EE8034265B23CDF274276C9ECA5511250C6E7') 'OriginalOwnershipIdentity'
}
$tokens=$null;$errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile($ownershipPath,[ref]$tokens,[ref]$errors)
foreach($name in @('SequentialIsolation','PlainControls')){Case ('BodyEdges/'+$name) {
    $calls=@($ast.FindAll({param($node)
        $node -is [Management.Automation.Language.CommandAst] -and $node.GetCommandName() -ceq 'P1I1-Case' -and $node.CommandElements.Count -eq 3 -and $node.CommandElements[1].Extent.Text -ceq $name
    },$true))
    Need ($calls.Count -eq 1) ('SingleBody/'+$name)
    $statements=$calls[0].CommandElements[2].ScriptBlock.EndBlock.Statements
    $prefix="[Console]::Out.WriteLine('OWNERSHIP: "+$name
    $suffix=" ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))"
    Need ($statements[0].Extent.Text -ceq ($prefix+' START'+$suffix)) ('FirstStatement/'+$name)
    Need ($statements[-1].Extent.Text -ceq ($prefix+' END'+$suffix)) ('LastStatement/'+$name)
}}
Case InventoryRestoresPreviousOrder {
    $prior=Get-Content -LiteralPath (Join-Path $old 'expected-output.json') -Raw|ConvertFrom-Json
    $next=Get-Content -LiteralPath (Join-Path $new 'expected-output.json') -Raw|ConvertFrom-Json
    $kept=@($next.Lines|Where-Object {$_.Kind -cne 'OwnershipCase'})
    Need ($next.Lines.Count -eq 105 -and $next.ExpectedNonemptyLines -eq 105 -and $next.ExpectedPassCount -eq 64 -and $next.ExpectedTimingMarkers -eq 22 -and $next.ExpectedClosedPairs -eq 11) 'InventoryCounts'
    Need ($kept.Count -eq 101) 'PriorCount'
    for($i=0;$i -lt 101;$i++){
        Need (($kept[$i]|ConvertTo-Json -Depth 8 -Compress) -ceq ($prior.Lines[$i]|ConvertTo-Json -Depth 8 -Compress)) ('PriorRecord/'+$i)
    }
    foreach($entry in @(@{Index=95;Name='SequentialIsolation';Edge='START'},@{Index=96;Name='SequentialIsolation';Edge='END'},@{Index=98;Name='PlainControls';Edge='START'},@{Index=99;Name='PlainControls';Edge='END'})){
        $rule=$next.Lines[$entry.Index]
        Need ($rule.Kind -ceq 'OwnershipCase' -and $rule.Name -ceq $entry.Name -and $rule.Edge -ceq $entry.Edge -and $rule.Parent -ceq '') ('OwnershipRule/'+$entry.Index)
    }
}
Case OnlyExpectedPreparationFiles {
    $names=@(Get-ChildItem -LiteralPath $new -File|ForEach-Object {$_.Name})
    Need ($names.Count -eq 10) 'TenFiles'
    foreach($name in @($sourceNames+@('expected-output.json','README.md'))){Need ($names -ccontains $name) ('ExpectedFile/'+$name)}
}
Write-Output ('TOTAL: '+$count+' static fidelity cases; target sources parsed/read only.')
