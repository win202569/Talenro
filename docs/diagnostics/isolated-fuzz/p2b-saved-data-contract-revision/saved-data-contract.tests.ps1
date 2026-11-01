# Definitions only. Caller consumes one explicit RED or GREEN intent.
function New-ODDataContext {
    $testRoot='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
    $probePath=Join-Path $testRoot 'docs/diagnostics/isolated-fuzz/p2b-error-reference-diagnostic/reference-probe.ps1'
    $probeHash=(Get-FileHash -LiteralPath $probePath -Algorithm SHA256).Hash
    if($probeHash -cne '494A5945B278A0F8DC299EECA3786839BEF722E65EDC884D4871D881AE0E1496'){throw 'TEST: frozen source drift'}
    $sourceText=[IO.File]::ReadAllText($probePath)
    $parseTokens=$null;$parseIssues=$null
    $sourceAst=[Management.Automation.Language.Parser]::ParseInput($sourceText,[ref]$parseTokens,[ref]$parseIssues)
    $faultFunctions=@($sourceAst.FindAll({param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -ceq 'Invoke-ODReferenceFault'},$true))
    if($parseIssues.Count -ne 0 -or $faultFunctions.Count -ne 1){throw 'TEST: source AST'}
    $faultStatements=@($faultFunctions[0].Body.FindAll({param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] -and $node.Member.Value -ceq 'ReadAllText'},$true))
    $caseCalls=@($sourceAst.FindAll({param($node) $node -is [Management.Automation.Language.CommandAst] -and $node.GetCommandName() -ceq 'Invoke-ODReferenceFault'},$true))
    $entryCalls=@($sourceAst.FindAll({param($node) $node -is [Management.Automation.Language.CommandAst] -and $node.GetCommandName() -ceq 'Ref-Case'},$true))
    if($faultStatements.Count -ne 1 -or $caseCalls.Count -ne 6 -or $entryCalls.Count -ne 1){throw 'TEST: source call model'}
    $caseLines=@($caseCalls|ForEach-Object{$_.Extent.StartLineNumber})
    if(($caseLines -join ',') -cne '54,56,60,65,71,77' -or $faultStatements[0].Extent.StartLineNumber -ne 16 -or $faultStatements[0].Extent.StartColumnNumber -ne 5 -or $entryCalls[0].Extent.StartLineNumber -ne 140){throw 'TEST: frozen source extents'}
    [ordered]@{PacketPath=(Join-Path $testRoot '.superpowers/sdd/2026-11-01-offline-error-reference-diagnostic');ProbePath=$probePath;ProbeHash=$probeHash;SourceText=$sourceText;SourceModel=[ordered]@{FaultLine=16;FaultColumn=5;CaseLines=$caseLines;EntryLine=140}}
}
function ConvertTo-ODTestBytes([object[]]$Rows){
    $rowTexts=@(foreach($rowValue in $Rows){ConvertTo-Json -InputObject $rowValue -Depth 30 -Compress})
    ,([Text.UTF8Encoding]::new($false).GetBytes(($rowTexts -join "`n")+"`n"))
}
function Get-ODTestRows([byte[]]$Bytes){
    $rowStrings=[Text.UTF8Encoding]::new($false,$true).GetString($Bytes).TrimEnd([char]10).Split([char]10)
    ,@(foreach($rowText in $rowStrings){ConvertFrom-Json -InputObject $rowText -AsHashtable})
}
function Set-ODTestAlternate([object[]]$Rows){
    $Rows[1].RecordSame=@(@($true,$true),@($true,$true))
    $Rows[1].ExceptionSame=@(@($true,$true),@($true,$true))
    $Rows[1].Collector.SecondaryIndices=@()
    $Rows[1].Events[1].AfterCount=1
}
function Assert-ODTestReport($Report,$Expected){
    if($Report -isnot [Collections.IDictionary] -or ($Report.Keys -join ',') -cne 'Schema,CasesChecked,FaultCalls,RecordedMeasurements' -or $Report.Schema -ne 1 -or $Report.CasesChecked -ne 6 -or $Report.FaultCalls -ne 7 -or $Report.RecordedMeasurements.Count -ne 6){throw 'TEST: report contract'}
    for($summaryIndex=0;$summaryIndex -lt 6;$summaryIndex++){
        $summary=$Report.RecordedMeasurements[$summaryIndex];$want=$Expected[$summaryIndex]
        if(($summary.Keys -join ',') -cne 'Case,RecordSame,ExceptionSame,CollectorTotal' -or $summary.Case -cne $want.Case -or $summary.RecordSame -isnot [bool] -or $summary.ExceptionSame -isnot [bool] -or $summary.RecordSame -ne $want.RecordSame -or $summary.ExceptionSame -ne $want.ExceptionSame -or $summary.CollectorTotal -ne $want.CollectorTotal){throw 'TEST: measurement mismatch'}
    }
}
function Assert-ODTestStatic([string]$Group,[string]$Leaf){
    $sourceNames=@('saved-data-contract.ps1','saved-data-contract.tests.ps1','receipt.ps1')
    $asts=@{}
    foreach($sourceName in $sourceNames){
        $staticTokens=$null;$staticIssues=$null
        $asts[$sourceName]=[Management.Automation.Language.Parser]::ParseFile((Join-Path $Leaf $sourceName),[ref]$staticTokens,[ref]$staticIssues)
        if($staticIssues.Count -ne 0){throw 'TEST: new source syntax'}
    }
    if($Group -ceq 'NoReservedVariableWrites'){
        $reserved=@('error','host','pid','home','args','input','psversiontable','pscommandpath','psscriptroot')
        foreach($tree in $asts.Values){
            $writes=@($tree.FindAll({param($node) $node -is [Management.Automation.Language.AssignmentStatementAst] -or $node -is [Management.Automation.Language.ParameterAst] -or $node -is [Management.Automation.Language.ForEachStatementAst]},$true))
            foreach($writeNode in $writes){
                $variables=@()
                if($writeNode -is [Management.Automation.Language.AssignmentStatementAst]){$variables=@($writeNode.Left.FindAll({param($node) $node -is [Management.Automation.Language.VariableExpressionAst]},$true))}
                elseif($writeNode -is [Management.Automation.Language.ParameterAst]){$variables=@($writeNode.Name)}
                else{$variables=@($writeNode.Variable)}
                foreach($variable in $variables){if(($variable.VariablePath.UserPath -split ':')[-1].ToLowerInvariant() -in $reserved){throw 'TEST: reserved variable write'}}
            }
        }
    }elseif($Group -ceq 'PureValidatorNoIOOrProcess'){
        $pureTree=$asts['saved-data-contract.ps1']
        foreach($command in @($pureTree.FindAll({param($node) $node -is [Management.Automation.Language.CommandAst]},$true))){
            if($command.GetCommandName() -notin @('ODR-Need','ODR-Keys','ODR-JSONNode','ODR-Integer','ConvertFrom-Json')){throw 'TEST: non-pure command'}
        }
        foreach($member in @($pureTree.FindAll({param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst]},$true))){
            if($member.Member.Value -notin @('new','Parse','Dispose','GetString','Split','Substring','StartsWith','EndsWith','Contains','GetRawText','EnumerateObject','EnumerateArray','Add','TryParse','ToArray','Replace','Escape','Match','GetHashCode')){throw ('TEST: non-pure method '+$member.Member.Value)}
        }
        if($pureTree.Extent.Text -match '(?i)\[(?:IO\.|Diagnostics\.|Net\.)|ScriptBlock\]::Create|Invoke-Expression|Add-Type|Start-Process'){throw 'TEST: validator side effects'}
    }else{
        $receiptTree=$asts['receipt.ps1']
        $receiptFunctions=@($receiptTree.FindAll({param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -ceq 'Write-ODSavedDataReceipt'},$true))
        if($receiptFunctions.Count -ne 1 -or $receiptFunctions[0].Body.ParamBlock.Parameters.Count -ne 1){throw 'TEST: fixed receipt interface'}
        $tryBlocks=@($receiptTree.FindAll({param($node) $node -is [Management.Automation.Language.TryStatementAst]},$true))
        foreach($tryBlock in $tryBlocks){if($null -ne $tryBlock.Finally -and @($tryBlock.Finally.FindAll({param($node) $node -is [Management.Automation.Language.ThrowStatementAst]},$true)).Count -ne 0){throw 'TEST: throw from finally'}}
        $opens=@($receiptTree.FindAll({param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] -and $node.Member.Value -ceq 'Open'},$true))
        $disposes=@($receiptTree.FindAll({param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] -and $node.Member.Value -ceq 'Dispose'},$true))
        if($opens.Count -ne 1 -or $disposes.Count -ne 1 -or $receiptTree.Extent.Text -notmatch 'FileMode\]::CreateNew' -or $receiptTree.Extent.Text -notmatch 'throw \$firstCause' -or $receiptTree.Extent.Text -notmatch 'if\(\$script:ReceiptAttempted\)' -or $receiptTree.Extent.Text -notmatch '\$script:ReceiptAttempted=\$true' -or $receiptTree.Extent.Text -notmatch '\$script:ReceiptFailed=\$true' -or $receiptTree.Extent.Text -notmatch 'ReceiptSecondary.Count -lt 8') {throw 'TEST: first-cause receipt structure'}
        if(@($receiptTree.FindAll({param($node) $node -is [Management.Automation.Language.LoopStatementAst]},$true)).Count -ne 0){throw 'TEST: receipt retry loop'}
    }
}
function Invoke-ODSavedDataTests([string]$Leaf,[byte[]]$OriginalBytes,[Collections.IDictionary]$Context){
    $expected=Get-Content -LiteralPath (Join-Path $Leaf 'expected-data-tests.json') -Raw|ConvertFrom-Json -AsHashtable
    $results=[Collections.Generic.List[object]]::new()
    $available=$null -ne (Get-Command Assert-ODSavedReferenceData -CommandType Function -ErrorAction SilentlyContinue)
    if(-not $available){
        $results.Add([ordered]@{Name='FunctionAvailable';Status='FAIL';Detail='New API absent'})
        foreach($name in $expected.Groups[1..41]){$results.Add([ordered]@{Name=$name;Status='NotRun';Detail='Availability gate'})}
        return [ordered]@{Schema=1;PASS=0;FAIL=1;NotRun=41;ExitCode=1;Groups=$results.ToArray()}
    }
    $results.Add([ordered]@{Name='FunctionAvailable';Status='PASS';Detail=''})
    foreach($groupIndex in 1..41){
        $name=$expected.Groups[$groupIndex];$status='PASS';$detail=''
        try{
            if($groupIndex -eq 1){
                $report=Assert-ODSavedReferenceData -Bytes $OriginalBytes -Context $Context
                Assert-ODTestReport $report $expected.OriginalMeasurements
            }elseif($groupIndex -eq 2){
                $alternateRows=Get-ODTestRows $OriginalBytes
                Set-ODTestAlternate $alternateRows
                $alternateExpected=@($expected.OriginalMeasurements)
                $alternateExpected[1]=$expected.AlternateMeasurement
                $report=Assert-ODSavedReferenceData -Bytes (ConvertTo-ODTestBytes $alternateRows) -Context $Context
                Assert-ODTestReport $report $alternateExpected
            }elseif($groupIndex -ge 39){Assert-ODTestStatic $name $Leaf}
            else{
                $variantNumber=$groupIndex-2
                $rows=Get-ODTestRows $OriginalBytes
                switch($variantNumber){
                    1 {$rows=@($rows[1..6])}
                    2 {$rows[1]=$rows[0]}
                    3 {$hold=$rows[0];$rows[0]=$rows[1];$rows[1]=$hold}
                    4 {$rows[0].Case='Unknown'}
                    5 {$rows[0].Extra=$true}
                    6 {$swapped=[ordered]@{Case=$rows[0].Case;Schema=$rows[0].Schema};foreach($key in @('FaultCalls','Events','RecordSame','ExceptionSame','Collector','ErrorProjections')){$swapped[$key]=$rows[0][$key]};$rows[0]=$swapped}
                    7 {$rows[0].Events=$rows[0].Events[0]}
                    8 {$rows[0].ErrorProjections=$rows[0].ErrorProjections[0]}
                    9 {$rows[0].RecordSame=$true}
                    10 {$rows[0].ExceptionSame=@($true)}
                    11 {$rows[1].Collector.SecondaryIndices=1}
                    12 {$rows[0].Events[0].ObservationIndex='0'}
                    13 {$rows[0].Collector.PrimaryIndex='0'}
                    14 {$rows[1].Collector.SecondaryIndices=@('1')}
                    15 {$rows[6].Schema='1'}
                    16 {$rows[6].Cases='6'}
                    17 {$rows[6].FaultCalls='7'}
                    18 {$rows[0].FaultCalls='1'}
                    19 {$rows[0].RecordSame[0][0]='true'}
                    20 {$rows[0].RecordSame[0][0]=$false}
                    21 {$rows[1].RecordSame[0][1]=$true}
                    22 {Set-ODTestAlternate $rows;$rows[1].ExceptionSame=@(@($true,$false),@($false,$true))}
                    23 {$rows[0].ErrorProjections[0].Position=$rows[0].ErrorProjections[0].Position.Replace(':16 char:5',':x char:5')}
                    24 {$rows[0].ErrorProjections[0].Position=$rows[0].ErrorProjections[0].Position.Replace(':16 char:5',':1 char:5')}
                    25 {$rows[0].ErrorProjections[0].Position=$rows[0].ErrorProjections[0].Position.Replace('reference-probe.ps1','other-probe.ps1')}
                    26 {$rows[0].ErrorProjections[0].ErrorId='Wrong'}
                    27 {$rows[0].ErrorProjections[0].Message='Wrong'}
                    36 {$rows[0].ErrorProjections[0].Stack=$rows[0].ErrorProjections[0].Stack.Replace(': line 16',': line 17')}
                }
                [byte[]]$variant=ConvertTo-ODTestBytes $rows
                $originalText=[Text.UTF8Encoding]::new($false).GetString($OriginalBytes)
                switch($variantNumber){
                    28 {$variant=[Text.UTF8Encoding]::new($false).GetBytes($originalText+"`n")}
                    29 {$variant=[Text.UTF8Encoding]::new($false).GetBytes($originalText.Substring(0,$originalText.Length-1))}
                    30 {$variant=[Text.UTF8Encoding]::new($false).GetBytes($originalText.Replace("`n","`r`n"))}
                    31 {$variant=[byte[]](@(239,187,191)+@($OriginalBytes))}
                    32 {$variant=[Text.UTF8Encoding]::new($false).GetBytes($originalText.Replace('"Schema":1,','"Schema":1,"Schema":1,'))}
                    33 {$variant=[byte[]]@(255,10)}
                    34 {$variant=[byte[]]::new(1048577)}
                    35 {$variant=[Text.UTF8Encoding]::new($false).GetBytes($originalText.Replace("`n","`n`n"))}
                }
                $rejected=$false
                try{$unexpectedReport=Assert-ODSavedReferenceData -Bytes $variant -Context $Context}
                catch{
                    if($_.Exception -isnot [Management.Automation.RuntimeException] -or $_.Exception.Message -cne $expected.Mismatch -or $_.FullyQualifiedErrorId -cne $expected.Mismatch){throw}
                    $rejected=$true
                }
                if(-not $rejected){throw 'TEST: invalid variant accepted'}
            }
        }catch{$status='FAIL';$detail=$_.Exception.GetType().FullName+': '+$_.Exception.Message+' | '+$_.ScriptStackTrace}
        $results.Add([ordered]@{Name=$name;Status=$status;Detail=$detail})
    }
    $failCount=@($results|Where-Object{$_.Status -ceq 'FAIL'}).Count
    [ordered]@{Schema=1;PASS=42-$failCount;FAIL=$failCount;NotRun=0;ExitCode=$(if($failCount -eq 0){0}else{1});Groups=$results.ToArray()}
}
