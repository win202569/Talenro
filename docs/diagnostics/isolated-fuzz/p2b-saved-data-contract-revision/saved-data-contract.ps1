# Pure saved bytes only. No file/process authority or live reference claims.
function ODR-Need([bool]$Condition){
    if(-not $Condition){throw 'ODR: saved data contract mismatch'}
}
function ODR-Keys($Value,[string[]]$Names){
    ODR-Need ($Value -is [Collections.IDictionary])
    $actualKeys=@($Value.Keys)
    ODR-Need ($actualKeys.Count -eq $Names.Count)
    for($keyIndex=0;$keyIndex -lt $Names.Count;$keyIndex++){
        ODR-Need ($actualKeys[$keyIndex] -is [string] -and $actualKeys[$keyIndex] -ceq $Names[$keyIndex])
    }
}
function ODR-Integer($Value){
    ODR-Need ($Value -is [long])
}
function ODR-JSONNode([Text.Json.JsonElement]$Node,[int]$Depth){
    ODR-Need ($Depth -le 32)
    switch([string]$Node.ValueKind){
        'Object' {
            $seenNames=[Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
            foreach($property in $Node.EnumerateObject()){
                ODR-Need ($seenNames.Add($property.Name))
                ODR-JSONNode $property.Value ($Depth+1)
            }
        }
        'Array' {foreach($element in $Node.EnumerateArray()){ODR-JSONNode $element ($Depth+1)}}
        'Number' {
            $rawNumber=$Node.GetRawText();[long]$integerValue=0
            ODR-Need ($rawNumber -cmatch '^-?(0|[1-9][0-9]*)$')
            ODR-Need ([long]::TryParse($rawNumber,[Globalization.NumberStyles]::AllowLeadingSign,[Globalization.CultureInfo]::InvariantCulture,[ref]$integerValue))
        }
        'String' {}
        'True' {}
        'False' {}
        'Null' {}
        default {ODR-Need $false}
    }
}
function Assert-ODSavedReferenceData([byte[]]$Bytes,[Collections.IDictionary]$Context){
    ODR-Need ($null -ne $Bytes -and $Bytes.Length -ge 1 -and $Bytes.Length -le 1048576)
    $text=$null
    try{$text=[Text.UTF8Encoding]::new($false,$true).GetString($Bytes)}
    catch{ODR-Need $false}
    ODR-Need (-not $text.StartsWith([string][char]0xFEFF) -and -not $text.Contains("`r") -and $text.EndsWith("`n") -and -not $text.EndsWith("`n`n"))
    $rowTexts=$text.Substring(0,$text.Length-1).Split([char]10)
    ODR-Need ($rowTexts.Count -eq 7)
    $rows=[Collections.Generic.List[object]]::new()
    foreach($rowText in $rowTexts){
        ODR-Need ($rowText.Length -gt 0)
        $document=$null
        try{
            $options=[Text.Json.JsonDocumentOptions]::new();$options.MaxDepth=32
            try{$document=[Text.Json.JsonDocument]::Parse($rowText,$options)}catch{ODR-Need $false}
            ODR-JSONNode $document.RootElement 0
            try{$parsedRow=ConvertFrom-Json -InputObject $rowText -AsHashtable -Depth 32 -ErrorAction Stop}catch{ODR-Need $false}
            $rows.Add($parsedRow)
        }finally{if($null -ne $document){$document.Dispose()}}
    }
    $names=@('DirectOuter','CatchBare','CatchRetained','FinallyBare','FinallyRetained','IndependentSameText')
    $measurements=[Collections.Generic.List[object]]::new()
    for($caseIndex=0;$caseIndex -lt 6;$caseIndex++){
        $caseRow=$rows[$caseIndex]
        ODR-Keys $caseRow @('Schema','Case','FaultCalls','Events','RecordSame','ExceptionSame','Collector','ErrorProjections')
        ODR-Integer $caseRow.Schema;ODR-Integer $caseRow.FaultCalls
        $observationCount=2;$faultCount=1
        [string[]]$labels=@('Inner','Outer')
        if($caseIndex -eq 0){$observationCount=1;$labels=@('Outer')}
        elseif($caseIndex -eq 5){$faultCount=2;$labels=@('First','Second')}
        ODR-Need ($caseRow.Schema -eq 1 -and $caseRow.Case -is [string] -and $caseRow.Case -ceq $names[$caseIndex] -and $caseRow.FaultCalls -eq $faultCount)
        ODR-Need ($caseRow.Events -is [array] -and $caseRow.Events.Count -eq $observationCount -and $caseRow.ErrorProjections -is [array] -and $caseRow.ErrorProjections.Count -eq $observationCount)
        ODR-Keys $caseRow.Collector @('PrimaryIndex','SecondaryIndices','Overflow')
        ODR-Integer $caseRow.Collector.PrimaryIndex
        ODR-Need ($caseRow.Collector.PrimaryIndex -eq 0 -and $caseRow.Collector.SecondaryIndices -is [array] -and $caseRow.Collector.SecondaryIndices.Count -le 1 -and $caseRow.Collector.Overflow -is [bool] -and -not $caseRow.Collector.Overflow)
        if($caseRow.Collector.SecondaryIndices.Count -eq 1){ODR-Integer $caseRow.Collector.SecondaryIndices[0];ODR-Need ($caseRow.Collector.SecondaryIndices[0] -eq 1 -and $observationCount -eq 2)}
        $total=1+$caseRow.Collector.SecondaryIndices.Count
        for($observationIndex=0;$observationIndex -lt $observationCount;$observationIndex++){
            $eventRow=$caseRow.Events[$observationIndex]
            ODR-Keys $eventRow @('Label','BeforeCount','AfterCount','ObservationIndex')
            foreach($integerKey in @('BeforeCount','AfterCount','ObservationIndex')){ODR-Integer $eventRow[$integerKey]}
            ODR-Need ($eventRow.Label -is [string] -and $eventRow.Label -ceq $labels[$observationIndex] -and $eventRow.ObservationIndex -eq $observationIndex -and $eventRow.BeforeCount -ge 0 -and $eventRow.BeforeCount -le 2 -and $eventRow.AfterCount -ge 1 -and $eventRow.AfterCount -le 2 -and ($eventRow.AfterCount-$eventRow.BeforeCount) -in @(0,1))
            if($observationIndex -eq 0){ODR-Need ($eventRow.BeforeCount -eq 0 -and $eventRow.AfterCount -eq 1)}
            else{ODR-Need ($eventRow.BeforeCount -eq $caseRow.Events[$observationIndex-1].AfterCount)}
            $recordProjection=$caseRow.ErrorProjections[$observationIndex]
            ODR-Keys $recordProjection @('Type','Message','ErrorId','Position','Stack')
            foreach($projectionKey in @('Type','Message','ErrorId','Position','Stack')){ODR-Need ($recordProjection[$projectionKey] -is [string])}
            $missing=$Context.PacketPath+'\cases\'+$names[$caseIndex]+'\absent.txt'
            $expectedMessage='Exception calling "ReadAllText" with "1" argument(s): "Could not find file '''+$missing+'''."'
            ODR-Need ($recordProjection.Type -ceq 'System.Management.Automation.MethodInvocationException' -and $recordProjection.ErrorId -ceq 'FileNotFoundException' -and $recordProjection.Message -ceq $expectedMessage)
            $positionLines=$recordProjection.Position.Replace("`r`n","`n").Split([char]10)
            ODR-Need ($positionLines.Count -eq 3)
            $positionMatch=[regex]::Match($positionLines[0],'^At '+[regex]::Escape($Context.ProbePath)+':(?<line>[0-9]+) char:(?<column>[0-9]+)$')
            ODR-Need ($positionMatch.Success)
            [long]$sourceLine=0;[long]$sourceColumn=0
            ODR-Need ([long]::TryParse($positionMatch.Groups['line'].Value,[ref]$sourceLine) -and [long]::TryParse($positionMatch.Groups['column'].Value,[ref]$sourceColumn))
            ODR-Need ($sourceLine -eq $Context.SourceModel.FaultLine -and $sourceColumn -eq $Context.SourceModel.FaultColumn)
            $sourceLines=$Context.SourceText.Replace("`r`n","`n").Split([char]10)
            ODR-Need ($positionLines[1] -ceq ('+ '+$sourceLines[$Context.SourceModel.FaultLine-1]) -and $positionLines[2] -ceq '+     ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~')
            $stackLines=$recordProjection.Stack.Replace("`r`n","`n").Split([char]10)
            ODR-Need ($stackLines.Count -eq 3)
            $frameNames=@('Invoke-ODReferenceFault','Ref-Case','<ScriptBlock>')
            $frameLines=@($Context.SourceModel.FaultLine,$Context.SourceModel.CaseLines[$caseIndex],$Context.SourceModel.EntryLine)
            for($frameIndex=0;$frameIndex -lt 3;$frameIndex++){
                $stackMatch=[regex]::Match($stackLines[$frameIndex],'^at '+[regex]::Escape($frameNames[$frameIndex])+', '+[regex]::Escape($Context.ProbePath)+': line (?<line>[0-9]+)$')
                ODR-Need ($stackMatch.Success)
                [long]$stackLine=0
                ODR-Need ([long]::TryParse($stackMatch.Groups['line'].Value,[ref]$stackLine) -and $stackLine -eq $frameLines[$frameIndex])
            }
        }
        foreach($matrixKey in @('RecordSame','ExceptionSame')){
            ODR-Need ($caseRow[$matrixKey] -is [array] -and $caseRow[$matrixKey].Count -eq $observationCount)
            for($rowIndex=0;$rowIndex -lt $observationCount;$rowIndex++){
                ODR-Need ($caseRow[$matrixKey][$rowIndex] -is [array] -and $caseRow[$matrixKey][$rowIndex].Count -eq $observationCount)
            }
            for($rowIndex=0;$rowIndex -lt $observationCount;$rowIndex++){
                for($columnIndex=0;$columnIndex -lt $observationCount;$columnIndex++){
                    ODR-Need ($caseRow[$matrixKey][$rowIndex][$columnIndex] -is [bool] -and $caseRow[$matrixKey][$rowIndex][$columnIndex] -eq $caseRow[$matrixKey][$columnIndex][$rowIndex])
                }
                ODR-Need ($caseRow[$matrixKey][$rowIndex][$rowIndex])
            }
        }
        for($rowIndex=0;$rowIndex -lt $observationCount;$rowIndex++){
            for($columnIndex=0;$columnIndex -lt $observationCount;$columnIndex++){
                ODR-Need (-not $caseRow.RecordSame[$rowIndex][$columnIndex] -or $caseRow.ExceptionSame[$rowIndex][$columnIndex])
            }
        }
        ODR-Need ($total -eq $caseRow.Events[$observationCount-1].AfterCount)
        $recordSame=$true;$exceptionSame=$true
        if($observationCount -eq 2){
            $recordSame=$caseRow.RecordSame[0][1];$exceptionSame=$caseRow.ExceptionSame[0][1]
            ODR-Need (($recordSame -and $total -eq 1) -or (-not $recordSame -and $total -eq 2))
        }
        if($caseIndex -eq 0){ODR-Need ($total -eq 1)}
        if($caseIndex -eq 5){ODR-Need ($total -eq 2 -and -not $recordSame)}
        $measurements.Add([ordered]@{Case=$caseRow.Case;RecordSame=$recordSame;ExceptionSame=$exceptionSame;CollectorTotal=$total})
    }
    $complete=$rows[6]
    ODR-Keys $complete @('Schema','Status','Cases','FaultCalls','SourceVerified','LiveActionsPerformed','RuntimeApproved','FormalAcceptance','RealGateImplemented')
    foreach($integerKey in @('Schema','Cases','FaultCalls')){ODR-Integer $complete[$integerKey]}
    ODR-Need ($complete.Schema -eq 1 -and $complete.Status -is [string] -and $complete.Status -ceq 'COMPLETE' -and $complete.Cases -eq 6 -and $complete.FaultCalls -eq 7)
    foreach($flag in @('SourceVerified','LiveActionsPerformed','RuntimeApproved','FormalAcceptance','RealGateImplemented')){ODR-Need ($complete[$flag] -is [bool] -and -not $complete[$flag])}
    [ordered]@{Schema=1;CasesChecked=6;FaultCalls=7;RecordedMeasurements=$measurements.ToArray()}
}
