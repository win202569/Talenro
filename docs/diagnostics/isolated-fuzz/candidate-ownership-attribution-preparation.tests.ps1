# Preparation fidelity only: read/AST parse target files,never import/invoke them.
$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$old=Join-Path $PSScriptRoot 'integer-candidate-ownership-markers'
$new=Join-Path $PSScriptRoot 'integer-candidate-validation-copy-attribution'
$entry='Get-OfflineSupervisionDecisionAttributionProbe'
$oldEntry='Get-OfflineSupervisionDecisionPhaseProbe'
$count=0
function Need([bool]$Value,[string]$CheckLabel){if(-not $Value){throw ('FIDELITY FAILED: '+$CheckLabel)}}
function Case([string]$CaseLabel,[scriptblock]$Body){& $Body;$script:count++;Write-Output ('PASS: AttributionPreparation/'+$CaseLabel)}
$names=@('offline-supervision-policy.ps1','phase-evaluator.ps1','offline-supervision-policy.tests.ps1','offline-supervision-policy.ownership.tests.ps1','offline-supervision-policy.reflection-reuse.tests.ps1','offline-supervision-policy.reflection-reuse-regression.ps1','offline-supervision-policy.shape-lookup.tests.ps1','offline-supervision-policy.shape-lookup-regression.ps1')
foreach($name in $names){Case ('Parse/'+$name){
    $tokens=$null;$errors=$null
    $null=[Management.Automation.Language.Parser]::ParseFile((Join-Path $new $name),[ref]$tokens,[ref]$errors)
    Need ($errors.Count -eq 0) ('Syntax/'+$name)
}}
foreach($name in $names){if($name -cnotin @('phase-evaluator.ps1','offline-supervision-policy.tests.ps1')){Case ('Identical/'+$name){
    Need ((Get-FileHash -LiteralPath (Join-Path $old $name)).Hash -ceq (Get-FileHash -LiteralPath (Join-Path $new $name)).Hash) ('Bytes/'+$name)
}}}
$oldText=[IO.File]::ReadAllText((Join-Path $old 'phase-evaluator.ps1'))
$text=[IO.File]::ReadAllText((Join-Path $new 'phase-evaluator.ps1'))
Case WholeCloneRestoration {
    Need ($oldText -cnotmatch '(?i)\$vcDiag') 'NoLocalNameCollision'
    Need ($oldText -cnotmatch [regex]::Escape($entry)) 'NoEntryCollision'
    $lines=@($text -split '\n');$added=@($lines|Where-Object {$_ -cmatch '\$vcDiag'})
    Need ($added.Count -eq 8) 'EightAddedLinesOnly'
    $restored=(@($lines|Where-Object {$_ -cnotmatch '\$vcDiag'}) -join "`n").Replace($entry,$oldEntry)
    Need ($restored -ceq $oldText) 'EntireCloneUnchangedAfterRemoval'
}
Case WholeCallerRestoration {
    $before=[IO.File]::ReadAllText((Join-Path $old 'offline-supervision-policy.tests.ps1'))
    $after=[IO.File]::ReadAllText((Join-Path $new 'offline-supervision-policy.tests.ps1'))
    Need ([regex]::Matches($after,[regex]::Escape($entry)).Count -eq 1) 'SingleCaller'
    Need ($after.Replace($entry,$oldEntry) -ceq $before) 'EntireCallerOnlyEntryNameChanged'
}
$tokens=$null;$errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $new 'phase-evaluator.ps1'),[ref]$tokens,[ref]$errors)
$functions=@($ast.FindAll({param($node)$node -is [Management.Automation.Language.FunctionDefinitionAst]},$true))
$loop=@($ast.FindAll({param($node)$node -is [Management.Automation.Language.ForStatementAst]},$true))[0]
Case OnlyOneUniqueDefinition {
    Need ($functions.Count -eq 1 -and $functions[0].Name -ceq $entry) 'OnlyDiagnosticEntry'
    Need ($functions[0].Parameters.Count -eq 1) 'OriginalParameterCount'
    $invocations=@($ast.FindAll({param($node)$node -is [Management.Automation.Language.CommandAst] -and $node.GetCommandName() -ceq $entry},$true))
    Need ($invocations.Count -eq 0) 'DefinitionOnlyStartupFile'
}
Case LocalTimersBracketOriginalStatements {
    $statements=$loop.Body.Statements
    Need ($statements[0].Extent.Text -ceq '$vcDiagVStart=[Diagnostics.Stopwatch]::GetTimestamp()') 'FirstRowStatement'
    $copyIndex=-1
    for($i=0;$i -lt $statements.Count;$i++){if($statements[$i].Extent.Text -ceq '$copy=[ordered]@{}'){$copyIndex=$i}}
    Need ($copyIndex -gt 0) 'OriginalCopyExists'
    Need ($statements[$copyIndex-1].Extent.Text -ceq '$vcDiagCStart=[Diagnostics.Stopwatch]::GetTimestamp()') 'BeforeOrderedSnapshot'
    Need ($statements[$copyIndex+2].Extent.Text -ceq '$observations.Add([pscustomobject]$copy)') 'OriginalListAdd'
    Need ($statements[$copyIndex+3].Extent.Text -ceq '$vcDiagCEnd=[Diagnostics.Stopwatch]::GetTimestamp()') 'AfterListAdd'
    Need ($statements[-3].Extent.Text -ceq '$vcDiagValidationTicks=[long]($vcDiagValidationTicks+($vcDiagCStart-$vcDiagVStart))') 'ValidationSumAfterEnd'
    Need ($statements[-2].Extent.Text -ceq '$vcDiagCopyTicks=[long]($vcDiagCopyTicks+($vcDiagCEnd-$vcDiagCStart))') 'CopySumAfterEnd'
    Need ($statements[-1].Extent.Text -ceq '$vcDiagCompletedRows++') 'CompletedRowCountLast'
    $readCount=[regex]::Matches($loop.Extent.Text,'\[Diagnostics.Stopwatch\]::GetTimestamp\(\)').Count
    Need ($readCount -eq 3) 'OnlyThreeTimestampReadsInRow'
    Need ($loop.Extent.Text -cnotmatch '\[Console\]|Write-Output|Write-Host') 'NoPerRowOutput'
}
Case InitializationAndSummaryOutsideLoop {
    # Position checks use parsed statement extents and whole-source restoration,
    # not executing the diagnostic function or accepting a text-only substitute.
    $initial=@($ast.FindAll({param($node)$node -is [Management.Automation.Language.AssignmentStatementAst] -and $node.Left.Extent.Text -cmatch '^\[long\]\$vcDiag'},$true))
    Need ($initial.Count -eq 6) 'SixTypedLocals'
    $rootEnd=$text.IndexOf("'PRODUCTPHASE: RootSetup END ticks='")
    $validationStart=$text.IndexOf("'PRODUCTPHASE: ValidationCopy START ticks='")
    foreach($item in $initial){Need ($item.Extent.StartOffset -gt $rootEnd -and $item.Extent.EndOffset -lt $validationStart) 'InitBetweenPhases'}
    $summary=@($ast.FindAll({param($node)$node -is [Management.Automation.Language.InvokeMemberExpressionAst] -and $node.Extent.Text.StartsWith("[Console]::Out.WriteLine('ATTRIBUTION: ",[StringComparison]::Ordinal)},$true))
    Need ($summary.Count -eq 1) 'OneSummaryOutput'
    Need ($summary[0].Extent.StartOffset -gt $loop.Extent.EndOffset -and $summary[0].Extent.StartOffset -gt $text.IndexOf("'PRODUCTPHASE: ValidationCopy END ticks='") -and $summary[0].Extent.EndOffset -lt $text.IndexOf("'PRODUCTPHASE: Decision START ticks='")) 'SummaryBetweenEndAndDecision'
}
Case InventoryRestoration {
    $prior=Get-Content -LiteralPath (Join-Path $old 'expected-output.json') -Raw|ConvertFrom-Json
    $next=Get-Content -LiteralPath (Join-Path $new 'expected-output.json') -Raw|ConvertFrom-Json
    Need ($next.Lines.Count -eq 106 -and $next.ExpectedNonemptyLines -eq 106 -and $next.ExpectedPassCount -eq 64 -and $next.ExpectedTimingMarkers -eq 22 -and $next.ExpectedClosedPairs -eq 11) 'Counts'
    $kept=@($next.Lines|Where-Object {$_.Kind -cne 'Attribution'})
    Need ($kept.Count -eq 105 -and $next.Lines[75].Kind -ceq 'Attribution' -and $next.Lines[75].Rows -eq 4096) 'OneSummaryAt76'
    for($i=0;$i -lt 105;$i++){Need (($kept[$i]|ConvertTo-Json -Depth 8 -Compress) -ceq ($prior.Lines[$i]|ConvertTo-Json -Depth 8 -Compress)) ('Record/'+$i)}
}
Case ExpectedFileScope {
    $files=@(Get-ChildItem -LiteralPath $new -File -Force|ForEach-Object {$_.Name})
    Need ($files.Count -eq 11) 'ElevenIncludingLeafAttributes'
    foreach($name in @($names+@('expected-output.json','README.md','.gitattributes'))){Need ($files -ccontains $name) ('File/'+$name)}
}
Write-Output ('TOTAL: '+$count+' static fidelity cases; target sources parsed/read only.')
