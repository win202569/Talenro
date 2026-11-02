# Definitions only. Single saved-data diagnosis; never a runtime acceptance test.
function Convert-ODDiagnosisError($Record){
    if($null -eq $Record){return $null}
    return [ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Position=$Record.InvocationInfo.PositionMessage;Stack=$Record.ScriptStackTrace}
}
function Write-ODDiagnosisJSON([string]$Path,$Value){
    $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 100))
    $stream=[IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
    try{$stream.Write($bytes,0,$bytes.Length);$stream.Flush()}finally{$stream.Dispose()}
}
function Trace-ODSavedValue($Actual,$Expected,$Context,$State,[string]$Path){
    $kind=if($Expected -is [Collections.IDictionary] -and -not $Expected.Contains('$error') -and -not $Expected.Contains('$dynamic')){'Dictionary'}elseif($Expected -is [array]){'Array'}else{'Atomic'}
    $visit=[ordered]@{Path=$Path;Kind=$kind;Matched=$false;Error=$null}
    $diagnosisTrace.Visits.Add($visit)
    try{
        if($kind -ceq 'Dictionary'){OC-Keys $Actual @($Expected.Keys)}
        elseif($kind -ceq 'Array'){OC-Need ($Actual -is [array] -and $Actual.Count -eq $Expected.Count)}
        else{OC-Match $Actual $Expected $Context $State}
        $visit.Matched=$true
    }catch{
        $visit.Error=Convert-ODDiagnosisError $_
        $diagnosisTrace.Failed=[ordered]@{Path=$Path;Actual=$Actual;Expected=$Expected;Error=$visit.Error}
        throw
    }
    if($kind -ceq 'Dictionary'){
        foreach($key in $Expected.Keys){Trace-ODSavedValue $Actual[$key] $Expected[$key] $Context $State ($Path+"['"+$key.Replace("'","''")+"']")}
    }elseif($kind -ceq 'Array'){
        for($index=0;$index -lt $Expected.Count;$index++){Trace-ODSavedValue $Actual[$index] $Expected[$index] $Context $State ($Path+'['+$index+']')}
    }
}
function Invoke-ODFaultLocationDiagnosis([string]$Repo,[string]$InputPath){
    Set-StrictMode -Version Latest
    $ErrorActionPreference='Stop'
    $fixedRepo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
    $leaf='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2'
    $packet=Join-Path $fixedRepo '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
    $finalPath=Join-Path $packet 'fault-location.final.json'
    $stopPath=Join-Path $packet 'fault-location.stop.json'
    if((Test-Path -LiteralPath $finalPath) -or (Test-Path -LiteralPath $stopPath)){throw 'OD diagnosis: output already exists; no replay'}
    $result=[ordered]@{Schema=1;Computation=15;Status='Inconclusive';InputSHA256=$null;DiagnosticSHA256=$null;PreHashes=[ordered]@{};PostHashes=[ordered]@{};Visits=@();FirstMismatch=$null;DiagnosticError=$null;RecordingError=$null;OriginalAcceptance='NOT_ACCEPTED';ExitCode=2}
    $diagnosisTrace=[pscustomobject]@{Visits=[Collections.Generic.List[object]]::new();Failed=$null}
    $binding=$null
    try{
        if([IO.Path]::GetFullPath($Repo) -cne $fixedRepo){throw 'OD diagnosis: original root required'}
        $inputFull=[IO.Path]::GetFullPath((Join-Path $Repo $InputPath))
        if($inputFull -cne (Join-Path $packet 'fault-location.input.json')){throw 'OD diagnosis: fixed input required'}
        $result.InputSHA256=(Get-FileHash -LiteralPath $inputFull).Hash
        $binding=Get-Content -LiteralPath $inputFull -Raw|ConvertFrom-Json -AsHashtable
        $intent=Get-Content -LiteralPath (Join-Path $packet 'fault-location.intent.json') -Raw|ConvertFrom-Json -AsHashtable
        $scriptPath=Join-Path $Repo ($leaf+'/fault-location.diagnostic.ps1')
        $result.DiagnosticSHA256=(Get-FileHash -LiteralPath $scriptPath).Hash
        if($intent.Schema -ne 1 -or $intent.AttemptConsumed -isnot [bool] -or -not $intent.AttemptConsumed -or $intent.Computation -ne 15 -or $intent.Ceiling -ne 15 -or $intent.InputSHA256 -cne $result.InputSHA256 -or $intent.DiagnosticSHA256 -cne $result.DiagnosticSHA256 -or $binding.DiagnosticSHA256 -cne $result.DiagnosticSHA256 -or $binding.Schema -ne 1 -or $binding.Repo -cne $fixedRepo -or $binding.Profile -cne 'entry-migration-core'){throw 'OD diagnosis: intent/input mismatch'}
        foreach($relative in $binding.Files.Keys){
            $path=[IO.Path]::GetFullPath((Join-Path $Repo $relative))
            if(-not $path.StartsWith($fixedRepo+'\',[StringComparison]::Ordinal)){throw 'OD diagnosis: input outside root'}
            $hash=(Get-FileHash -LiteralPath $path).Hash
            $result.PreHashes[$relative]=$hash
            if($hash -cne $binding.Files[$relative]){throw ('OD diagnosis: input changed '+$relative)}
        }
        $manifest=Get-Content -LiteralPath (Join-Path $packet 'repair-green.frozen.json') -Raw|ConvertFrom-Json -AsHashtable
        if($manifest.SourcePaths.Count -ne 31 -or $binding.OperationalSources.Count -ne 31){throw 'OD diagnosis: source inventory count'}
        foreach($relative in $manifest.SourcePaths){
            if(-not $binding.OperationalSources.Contains($relative) -or $result.PreHashes[$relative] -cne $binding.OperationalSources[$relative]){throw 'OD diagnosis: source inventory mismatch'}
        }
        $allowed=[ordered]@{
            'source-model.ps1'=@('Get-ODPinnedSHA256','Get-ODSourceOwner','Get-ODFailureExtent','Read-ODSourceModel','New-ODFixedContext','New-ODHistoricalContext')
            'observation-contract.ps1'=@('OC-Need','OC-Keys','OC-Integer','OC-Expand','OC-SourceError','OC-Match')
        }
        if((ConvertTo-Json $allowed -Compress) -cne (ConvertTo-Json $binding.HelperNames -Compress)){throw 'OD diagnosis: helper allowlist mismatch'}
        foreach($name in $allowed.Keys){
            $tokens=$null;$issues=$null
            $tree=[Management.Automation.Language.Parser]::ParseFile((Join-Path $Repo ($leaf+'/'+$name)),[ref]$tokens,[ref]$issues)
            if($issues.Count -ne 0){throw 'OD diagnosis: source parse failure'}
            foreach($helper in $allowed[$name]){
                $defs=@($tree.EndBlock.Statements|Where-Object {$_ -is [Management.Automation.Language.FunctionDefinitionAst] -and $_.Name -ceq $helper})
                if($defs.Count -ne 1){throw 'OD diagnosis: ambiguous helper'}
                # Prefix newlines preserve original predicate call-site line numbers.
                . ([ScriptBlock]::Create(([string][char]10 * ($defs[0].Extent.StartLineNumber-1))+$defs[0].Extent.Text))
            }
        }
        $historical=Join-Path $Repo '.superpowers/sdd/2026-10-31-offline-diagnostic-runtime-migration/runtime-green.evidence.json'
        $saved=Get-Content -LiteralPath $historical -Raw|ConvertFrom-Json -AsHashtable
        $returns=@($saved.Observations|Where-Object {$_.Name -ceq 'ProfileReturn'})
        if($returns.Count -ne 1 -or $returns[0].Value.Profile -cne 'entry-migration-core'){throw 'OD diagnosis: canonical baseline unavailable'}
        $rows=ConvertFrom-Json -AsHashtable -InputObject (ConvertTo-Json -InputObject $returns[0].Value.Observations -Depth 60)
        $context=New-ODHistoricalContext $Repo 'entry-migration-core' $manifest
        $state=[pscustomobject]@{Pids=[ordered]@{};Elapsed=[ordered]@{}}
        OC-Need ($context.Rows.Contains('entry-migration-core') -and $rows.Count -eq 22 -and $context.Rows['entry-migration-core'] -is [array] -and $context.Rows['entry-migration-core'].Count -eq 22)
        try{Trace-ODSavedValue $rows $context.Rows['entry-migration-core'] $context $state '$'}catch{
            if($null -eq $diagnosisTrace.Failed){throw}
            if($_.Exception.GetType().FullName -cne 'System.Management.Automation.RuntimeException' -or $_.Exception.Message -cne 'OD: observation contract mismatch' -or $_.FullyQualifiedErrorId -cne 'OD: observation contract mismatch'){throw}
        }
        $result.ExitCode=1
        if($null -ne $diagnosisTrace.Failed){
            $failed=$diagnosisTrace.Failed
            $rowMatch=[regex]::Match($failed.Path,'^\$\[([0-9]+)\]')
            $detail=[ordered]@{Path=$failed.Path;RowIndex=$(if($rowMatch.Success){[int]$rowMatch.Groups[1].Value}else{$null});ActualError=$null;ExpectedError=$null;Source=$null;Function=$null;Operation=$null;Line=$null;Column=$null;Candidates=@();ASTExtents=@();PredicateSite=$failed.Error}
            $result.FirstMismatch=$detail
            $result.Status='HypothesisContradicted'
            if($failed.Expected -is [Collections.IDictionary] -and $failed.Expected.Contains('$error')){
                $detail.ActualError=$failed.Actual;$detail.ExpectedError=$failed.Expected['$error']
                $expectedError=$detail.ExpectedError
                $source=$expectedError.Source
                if($context.Sources.Contains($source)){$source=$context.Sources[$source]}
                $detail.Source=$source;$detail.Function=$expectedError.Function
                $operation=$null
                if($expectedError.Contains('Operation')){$operation=$expectedError.Operation}
                elseif($expectedError.Type -ceq 'System.Management.Automation.MethodInvocationException'){
                    $methodMatch=[regex]::Match($expectedError.Message,'^Exception calling "([^"]+)"')
                    if($methodMatch.Success){$operation=$methodMatch.Groups[1].Value}
                }elseif($expectedError.ErrorId -ceq 'Invalid ledger row index'){$operation='throw'}
                elseif($expectedError.ErrorId -ceq 'System.ArgumentException,Microsoft.PowerShell.Commands.ConvertFromJsonCommand'){$operation='ConvertFrom-Json'}
                $detail.Operation=$operation
                $positionMatch=[regex]::Match($failed.Actual.Position,('^At '+[regex]::Escape($source)+':([0-9]+) char:([0-9]+)\r?\n'))
                if($positionMatch.Success -and $context.SourceModel.Contains($source) -and $null -ne $operation){
                    $detail.Line=[int]$positionMatch.Groups[1].Value;$detail.Column=[int]$positionMatch.Groups[2].Value
                    $detail.Candidates=@($context.SourceModel[$source].Faults|Where-Object {$_.Function -ceq $expectedError.Function -and $_.Operation -ceq $operation})
                    $exact=@($detail.Candidates|Where-Object {$_.Line -eq $detail.Line -and $_.Column -eq $detail.Column})
                    $tokens=$null;$issues=$null
                    $tree=[Management.Automation.Language.Parser]::ParseFile($source,[ref]$tokens,[ref]$issues)
                    if($issues.Count -ne 0){throw 'OD diagnosis: detail AST parse failure'}
                    $nodes=@($tree.FindAll({param($node)$node.Extent.StartLineNumber -eq $detail.Line -and $node.Extent.StartColumnNumber -eq $detail.Column},$true))
                    $extents=[Collections.Generic.List[object]]::new()
                    foreach($node in $nodes){
                        $parent=$node;$depth=0
                        while($null -ne $parent -and $depth -lt 16){
                            $extents.Add([ordered]@{Depth=$depth;Kind=$parent.GetType().Name;Line=$parent.Extent.StartLineNumber;Column=$parent.Extent.StartColumnNumber;EndLine=$parent.Extent.EndLineNumber;EndColumn=$parent.Extent.EndColumnNumber;Text=$parent.Extent.Text})
                            if($parent -is [Management.Automation.Language.StatementBlockAst]){break}
                            $parent=$parent.Parent;$depth++
                        }
                    }
                    $detail.ASTExtents=$extents.ToArray()
                    if($failed.Error.Stack -match '(?m)^at OC-SourceError, .*: line 33\r?$' -and $exact.Count -eq 0 -and $detail.Candidates.Count -gt 0 -and $nodes.Count -gt 0){$result.Status='LocatedMismatch';$result.ExitCode=0}
                }else{$result.Status='Inconclusive'}
            }
        }
    }catch{$result.DiagnosticError=Convert-ODDiagnosisError $_;$result.Status='Inconclusive';$result.ExitCode=2}
    $result.Visits=$diagnosisTrace.Visits.ToArray()
    try{
        if($null -ne $binding){
            foreach($relative in $binding.Files.Keys){
                $hash=(Get-FileHash -LiteralPath (Join-Path $fixedRepo $relative)).Hash
                $result.PostHashes[$relative]=$hash
                if($hash -cne $binding.Files[$relative]){throw ('OD diagnosis: post-input changed '+$relative)}
            }
        }
    }catch{$result.DiagnosticError=Convert-ODDiagnosisError $_;$result.Status='Inconclusive';$result.ExitCode=2}
    try{Write-ODDiagnosisJSON $finalPath $result}catch{
        $result.RecordingError=Convert-ODDiagnosisError $_;$result.Status='Inconclusive';$result.ExitCode=2
        $result|ConvertTo-Json -Depth 100
        throw
    }
    if($result.Status -cne 'LocatedMismatch'){
        Write-ODDiagnosisJSON $stopPath ([ordered]@{Schema=1;Computation=15;Status=$result.Status;DiagnosticError=$result.DiagnosticError;FirstMismatch=$result.FirstMismatch;Remainder='FROZEN'})
    }
    [ordered]@{Status=$result.Status;Visits=$result.Visits.Count;FirstPath=$(if($null -ne $result.FirstMismatch){$result.FirstMismatch.Path}else{$null});DiagnosticError=$result.DiagnosticError;ExitCode=$result.ExitCode;OriginalAcceptance=$result.OriginalAcceptance}|ConvertTo-Json -Depth 8
    $global:LASTEXITCODE=$result.ExitCode
}
