function Copy-ODContractData($Value){
    return ConvertFrom-Json -AsHashtable -InputObject (ConvertTo-Json -InputObject $Value -Depth 80)
}
function Set-ODWrongFaultLocation($Value,$Cause,[int]$Line){
    if($Value -is [Collections.IDictionary]){
        if($Value.Contains('Position') -and $Value.Contains('Stack') -and $Value.Contains('Message') -and $Value.Message -ceq $Cause.Message){
            $Value.Position=[regex]::Replace($Value.Position,':([0-9]+) char:',(':'+ $Line+' char:'))
            $Value.Stack=[regex]::Replace($Value.Stack,'^(at [^\r\n]+: line )[0-9]+',('${1}'+$Line))
        }
        foreach($key in @($Value.Keys)){Set-ODWrongFaultLocation $Value[$key] $Cause $Line}
    }elseif($Value -is [array]){foreach($item in $Value){Set-ODWrongFaultLocation $item $Cause $Line}}
}
function Invoke-ODContractCase([string]$Name,[string]$Profile,[string]$Kind,[bool]$ShouldReject,[ScriptBlock]$Body,$Suite){
    $actualError=$null;$rejected=$false;$emitted=[Collections.Generic.List[object]]::new()
    try{& $Body|ForEach-Object {$emitted.Add($_)}}catch{
        $actualError=$_
        if($_.Exception.GetType().FullName -cne 'System.Management.Automation.RuntimeException' -or $_.Exception.Message -cne 'OD: observation contract mismatch' -or $_.FullyQualifiedErrorId -cne 'OD: observation contract mismatch'){throw}
        $rejected=$true
    }
    $passed=$rejected -eq $ShouldReject -and $emitted.Count -eq 0
    $Suite.Add([ordered]@{Name=$Name;Profile=$Profile;Kind=$Kind;ExpectedRejected=$ShouldReject;ActualRejected=$rejected;Error=$actualError})
    $script:ContractSuiteRow.Value=$Suite.ToArray()
    LS-Result $Name $passed ([ordered]@{Verified=$passed})
    LS-Need $passed ('exact full-contract result '+$Name)
}
function New-ODSchema2Fixture([string]$Profile,[object[]]$Rows,$Context){
    $grant=switch -CaseSensitive ($Profile){'entry-migration-core'{'core'}'entry-migration-owned'{'lifecycle'}'entry-migration-live'{'continuation'}default{throw 'ASSERT: unknown fixture profile'}}
    $all=$oracle.Grants[$grant].Groups
    # Frozen literal profile groups exclude six entry groups, two runtime
    # binding groups and three recorder groups. No predicate derives this data.
    $groups=@($all|Select-Object -Skip 6 -First ($all.Count-11))
    return [ordered]@{
        Schema=2;Profile=$Profile;EntryPath=$Context.Entry.Path;EntrySHA256=$Context.Entry.SHA256
        KernelPath=$Context.Kernel.Path;KernelSHA256=$Context.Kernel.SHA256;FrozenSHA256=$Context.Manifest.SHA256
        Status='COMPLETE';Groups=$groups;Observations=$Rows;PrimaryError=$null;SecondaryErrors=@();Overflow=$false;Closed=$true
        SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false
    }
}
function Read-ODContractTestAST([string]$Path,$Manifest){
    $null=Get-ODPinnedSHA256 $Path $Manifest
    $tokens=$null;$issues=$null
    $ast=[Management.Automation.Language.Parser]::ParseFile($Path,[ref]$tokens,[ref]$issues)
    LS-Need ($issues.Count -eq 0) 'frozen static source parses'
    $null=Get-ODPinnedSHA256 $Path $Manifest
    return $ast
}
function Invoke-ODContractStaticChecks($Manifest,$Suite){
    $paths=@(foreach($name in @('diagnose-offline.ps1','private/session.psm1','private/sinks.ps1','private/owned-step.ps1','private/sequence.ps1','private/profiles.ps1','private/observations.ps1','private/write-boundary.ps1','private/record-owner.ps1')){Join-Path $repo ('scripts/offline-diagnostics-v4/'+$name)})
    foreach($name in @('test-helpers.ps1','entry.tests.ps1','core.tests.ps1','prefix.tests.ps1','owned.tests.ps1','live.tests.ps1','capture.ps1','driver.ps1','record-boundary.tests.ps1','observation-contract.tests.ps1','observation-contract.ps1','source-model.ps1','recording.tests.ps1','case-module.ps1','publisher.psm1')){$paths+=Join-Path $PSScriptRoot $name}
    $trees=@{}
    foreach($path in $paths){$trees[$path]=Read-ODContractTestAST $path $Manifest}
    Invoke-ODContractCase StaticReservedNames '' Static $false {
        foreach($tree in $trees.Values){
            $bad=@($tree.FindAll({param($node)
                ($node -is [Management.Automation.Language.AssignmentStatementAst] -and $node.Left -is [Management.Automation.Language.VariableExpressionAst] -and $node.Left.VariablePath.UserPath -imatch '^(script:|local:|global:)?(Error|PID|Host|HOME|Args|Input|PSItem|PSScriptRoot|PSCommandPath)$') -or
                ($node -is [Management.Automation.Language.ParameterAst] -and $node.Name.VariablePath.UserPath -imatch '^(Error|PID|Host|HOME|Args|Input|PSItem)$') -or
                ($node -is [Management.Automation.Language.UnaryExpressionAst] -and $node.Child -is [Management.Automation.Language.VariableExpressionAst] -and $node.Child.VariablePath.UserPath -imatch '^(script:|local:|global:)?(Error|PID|Host|HOME|Args|Input|PSItem)$')
            },$true))
            LS-Need ($bad.Count -eq 0) 'no reserved automatic-variable writes'
        }
    } $Suite
    $validator=$trees[(Join-Path $PSScriptRoot observation-contract.ps1)]
    Invoke-ODContractCase StaticPureContract '' Static $false {
        $bad=@($validator.FindAll({param($node)
            ($node -is [Management.Automation.Language.CommandAst] -and ($null -eq $node.GetCommandName() -or $node.GetCommandName() -cmatch '^(Get-Content|Set-Content|Add-Content|Get-FileHash|Test-Path|Import-Module|New-Module|Invoke-Expression|Start-Process|Add-Type|Get-Command|Get-Item|Get-ChildItem|Remove-Item|New-Item|Out-File|Invoke-WebRequest|Invoke-RestMethod)$')) -or
            ($node -is [Management.Automation.Language.TypeExpressionAst] -and $node.TypeName.FullName -imatch '^(System\.)?(IO\.(File|Directory|FileStream|StreamReader|StreamWriter)|Diagnostics\.|Reflection\.|Management\.Automation\.ScriptBlock)')
        },$true))
        LS-Need ($bad.Count -eq 0) 'pure validator cannot access resources/evaluate code'
    } $Suite
    Invoke-ODContractCase StaticAllConsumersBound '' Static $false {
        $driver=$trees[(Join-Path $PSScriptRoot driver.ps1)]
        $consumers=@($driver.FindAll({param($node)$node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -ceq 'Driver-Evidence'},$false))
        LS-Need ($consumers.Count -eq 1) 'one actual final evidence consumer'
        $calls=@($consumers[0].FindAll({param($node)$node -is [Management.Automation.Language.CommandAst]},$true)|ForEach-Object {$_.GetCommandName()})
        LS-Need ('New-ODRuntimeContext' -cin $calls -and 'Assert-ODProfileReport' -cin $calls -and 'Assert-ODProfileObservations' -cin $calls -and 'New-ODHistoricalContext' -cnotin $calls -and 'New-ODIntentUnitContext' -cnotin $calls) 'actual parent fixed to runtime context and both predicates'
        $reports=@($validator.FindAll({param($node)$node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -ceq 'Assert-ODProfileReport'},$false))
        LS-Need ($reports.Count -eq 1) 'one actual report predicate'
        $rowCalls=@($reports[0].FindAll({param($node)$node -is [Management.Automation.Language.CommandAst] -and $node.GetCommandName() -ceq 'Assert-ODProfileObservations'},$true))
        LS-Need ($rowCalls.Count -ge 1) 'report predicate uses same complete row predicate'
    } $Suite
}
function Invoke-ODContractTests {
    param([string]$Epoch='legacy')
    if($Epoch -cnotin @('legacy','extent-contract','history-contract','history-contract-runtime')){throw 'OD: invalid explicit epoch'}
    . (Join-Path $PSScriptRoot observation-contract.ps1)
    $available=$true
    foreach($name in @('Assert-ODProfileObservations','Assert-ODProfileReport')){
        if($null -eq (Get-Command $name -CommandType Function -ErrorAction Ignore)){$available=$false}
    }
    if(-not $available){
        $names=@($oracle.ContractGroupNames)+@($oracle.Recorder.Names)
        $groups=@(foreach($name in $names){[ordered]@{Name=$name;Status=$(if($name -ceq 'ContractAPIAvailable'){'FAIL'}else{'NotRun'})}})
        return [pscustomobject][ordered]@{Schema=1;PASS=0;FAIL=1;NotRun=51;ExitCode=1;Groups=$groups}
    }
    LS-Result ContractAPIAvailable $true ([ordered]@{Verified=$true})
    $suite=[Collections.Generic.List[object]]::new()
    $suite.Add([ordered]@{Name='ContractAPIAvailable';Profile='';Kind='Availability';ExpectedRejected=$false;ActualRejected=$false;Error=$null})
    LS-Details SchemaSuite ($suite.ToArray())
    $script:ContractSuiteRow=$script:Details[$script:Details.Count-1]
    . (Join-Path $PSScriptRoot source-model.ps1)
    $manifestPath=Join-Path $packet (Get-ODEpochManifestName $Epoch)
    LS-Need ((Get-FileHash -LiteralPath $manifestPath -Algorithm SHA256).Hash -ceq $FrozenSHA256) 'contract host bound to repair freeze'
    $manifest=Get-Content -LiteralPath $manifestPath -Raw|ConvertFrom-Json -AsHashtable
    $historical=Join-Path $repo '.superpowers/sdd/2026-10-31-offline-diagnostic-runtime-migration'
    $refs=@(
        @('runtime-green','entry-migration-core','58B2D81047526EFB2BEE3FD8155DFAF493FC3FD6ECAAE32B0E7ED19C7DB66EA0'),
        @('lifecycle','entry-migration-owned','3F7BDEE2B22F3C91EB5957259F18E1DF96E0CFD18A3F5E692D54C26460A51911'),
        @('continuation','entry-migration-live','D6729F1EBB54435B426013704B1AFD754E82F2EE7F6C725F520C02A56C532726')
    )
    $frames=[Collections.Generic.List[object]]::new()
    $contexts=[Collections.Generic.List[object]]::new()
    foreach($ref in $refs){
        $path=Join-Path $historical ($ref[0]+'.evidence.json')
        LS-Need ((Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash -ceq $ref[2]) 'historical baseline original hash'
        $e=Get-Content -LiteralPath $path -Raw | ConvertFrom-Json -AsHashtable
        $returns=@($e.Observations|Where-Object {$_.Name -ceq 'ProfileReturn'})
        LS-Need ($returns.Count -eq 1 -and $returns[0].Value.Profile -ceq $ref[1]) 'baseline canonical profile'
        $frames.Add($returns[0].Value.Observations)
        $contexts.Add((New-ODHistoricalContext $repo $ref[1] $manifest))
    }
    for($caseIndex=0;$caseIndex -lt 23;$caseIndex++){
        $vector=if($caseIndex -lt 3){$caseIndex+1}else{$caseIndex-2}
        $profileIndex=if($caseIndex -lt 3){$caseIndex}elseif($vector -eq 16){2}elseif($vector -cin @(17,18)){1}else{0}
        $profile=$refs[$profileIndex][1]
        $rows=ConvertFrom-Json -AsHashtable -InputObject (ConvertTo-Json -InputObject $frames[$profileIndex] -Depth 60)
        if($caseIndex -ge 3){
            switch($vector){
                1 {$rows=@($rows[0..1]) + @($rows[3..21])}
                2 {$rows=@($rows[0..2]) + @($rows[2]) + @($rows[3..21])}
                3 {$tmp=$rows[2];$rows[2]=$rows[3];$rows[3]=$tmp}
                4 {$rows[2].Name='Unknown'}
                5 {$rows[6].Value.Result['Extra']=0}
                6 {$null=$rows[6].Value.Result.Remove('Error')}
                7 {$rows[6].Value.Result.Error.Type='System.Exception'}
                8 {$rows[6].Value.Result.Error.ErrorId='UnrelatedError'}
                9 {$rows[6].Value.Result.Error.Message='unrelated'}
                10 {$rows[6].Value.View.Status='Active'}
                11 {$rows[6].Value.Result.Ok=$true}
                12 {$rows[6].Value.Result.Value=[ordered]@{}}
                13 {$null=$rows[1].Value.Result.Remove('Value')}
                14 {$rows[14].Value.View.Sinks.ConsoleOut.State='Healthy'}
                15 {$rows[13].Value.SecondaryErrors=@($rows[13].Value.SecondaryErrors|Select-Object -First 7)}
                16 {$rows[14].Value.A.AckConsumed=0}
                17 {$rows[4].Value.Result.Observation.OwnedExit=$false}
                18 {$rows[4].Value.Result.Observation.PID=0}
                19 {$rows[16].Value.Result.Error.ErrorId='UnrelatedError'}
                20 {$rows[16].Value.Result.Accepted='true'}
            }
        }
        $kind=if($caseIndex -lt 3){'Baseline'}else{'Mutation'}
        $name=if($caseIndex -lt 3){@('BaselineCore','BaselineOwned','BaselineLive')[$caseIndex]}else{'Reject'+$vector.ToString('00')}
        $context=$contexts[$profileIndex]
        Invoke-ODContractCase $name $profile $kind ($caseIndex -ge 3) {Assert-ODProfileObservations $rows $profile $context} $suite
    }
    $reports=[Collections.Generic.List[object]]::new()
    for($i=0;$i -lt 3;$i++){
        $profile=$refs[$i][1];$context=$contexts[$i]
        $report=New-ODSchema2Fixture $profile $frames[$i] $context
        $reports.Add($report)
        $name=@('Schema2CompleteCore','Schema2CompleteOwned','Schema2CompleteLive')[$i]
        Invoke-ODContractCase $name $profile Schema2 $false {Assert-ODProfileReport $report $profile $context} $suite
    }
    # Input is the actual saved earlier host's report, not an expectation
    # learned from GREEN. Its expected error signature/context are literal.
    $writerEvidencePath=Join-Path $packet 'writer-projection-green.evidence.json'
    $writerFinal=Get-Content (Join-Path $packet 'writer-projection-green.final.json') -Raw|ConvertFrom-Json -AsHashtable
    LS-Need ($writerFinal.Exact -is [bool] -and $writerFinal.Exact) 'only exact writer host may feed data fixture'
    $writerEvidence=Get-Content -LiteralPath $writerEvidencePath -Raw|ConvertFrom-Json -AsHashtable
    $owners=@($writerEvidence.Observations|Where-Object Name -CEQ OwnerControls)
    LS-Need ($owners.Count -eq 1) 'actual owner controls observation'
    $intentCases=@($owners[0].Value|Where-Object Name -CEQ IntentFailureNoWork)
    LS-Need ($intentCases.Count -eq 1) 'actual intent-failure input'
    $intentReport=$intentCases[0].IntentReport
    $intentContext=New-ODIntentUnitContext $repo $manifest $FrozenSHA256 $Epoch
    Invoke-ODContractCase Schema2IntentFailure entry-migration-core IntentFailure $false {Assert-ODProfileReport $intentReport entry-migration-core $intentContext;LS-Need ($intentReport.Status -ceq 'INCOMPLETE') 'valid representation is not successful disposition'} $suite
    for($vector=1;$vector -le 18;$vector++){
        $report=Copy-ODContractData $(if($vector -cin @(10,12)){$intentReport}else{$reports[0]})
        $context=if($vector -cin @(10,12)){$intentContext}else{$contexts[0]}
        switch($vector){
            1 {$report.Schema=1}
            2 {$null=$report.Remove('Closed')}
            3 {$report.Extra=0}
            4 {$swapped=[ordered]@{Profile=$report.Profile;Schema=$report.Schema};foreach($key in $report.Keys){if($key -cnotin @('Schema','Profile')){$swapped[$key]=$report[$key]}};$report=$swapped}
            5 {$report.RuntimeApproved=$true}
            6 {$report.Groups=$report.Groups[0]}
            7 {$report.Observations=$report.Observations[0]}
            8 {$report.SecondaryErrors='scalar'}
            9 {$report.PrimaryError='scalar'}
            10 {$report.Status='COMPLETE'}
            11 {$report.Closed='true'}
            12 {$report.PrimaryError=$null}
            13 {$report.PrimaryError=Copy-ODContractData $report.Observations[6].Value.Result.Error}
            14 {$report.Observations[0].Value.Root=$report.Observations[0].Value.Root+'-foreign'}
            15 {
                $cause=Copy-ODContractData $report.Observations[6].Value.Result.Error
                $functions=$context.SourceModel[$context.Sources.Sinks].Functions
                $function=@($functions|Where-Object Name -CEQ Read-LSLedger)
                LS-Need ($function.Count -eq 1) 'independently parsed fault function'
                # A real source line, but a function declaration, NOT a fault
                # statement. Change all duplicate projections coherently so
                # reference-projection equality cannot substitute for AST checks.
                Set-ODWrongFaultLocation $report $cause $function[0].StartLine
            }
            16 {$report.Observations[0].Value.StartCalls='0'}
            17 {$report.Observations[0].Value.FailureLatched='false'}
            18 {$report=[ordered]@{Schema=2;Profile='entry-migration-core';Status='UNKNOWN';Groups=@();Observations=@()}}
        }
        Invoke-ODContractCase ('AdditionalReject{0:00}' -f $vector) entry-migration-core Schema2Mutation $true {Assert-ODProfileReport $report entry-migration-core $context} $suite
    }
    Invoke-ODContractStaticChecks $manifest $suite
    LS-Need ($suite.Count -eq 49 -and ((@($suite|ForEach-Object {$_.Name}) -join '|') -ceq ($oracle.ContractGroupNames -join '|'))) 'complete ordered contract suite'
    # One bounded aggregate, not49top-level details exceeding the32row limit.
    $script:ContractSuiteRow.Value=$suite.ToArray()
}
