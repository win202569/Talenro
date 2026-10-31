# Test utilities, not exported controller hooks. Expectations are literal facts.
function LS-Need([bool]$Condition,[string]$Message){if(-not $Condition){throw ('ASSERT: '+$Message)}}
function LS-New([string]$Case){
    $script:CurrentGroup=$Case
    $handle=New-LSSession $Grant $Case $FrozenSHA256
    $script:ActiveSessions.Add($handle)
    return $handle
}
function LS-Path($Session,[string]$Name){Join-Path (Get-LSView $Session).Root $Name}
function LS-Details([string]$Name,$Value){
    if($script:Details.Count -ge 32){throw 'Observation limit exceeded'}
    $script:Details.Add([ordered]@{Name=$Name;Value=$Value})
}
function LS-Result([string]$Name,[bool]$Passed,$Facts){
    $script:CurrentGroup=$Name
    $script:Groups.Add([ordered]@{Name=$Name;Passed=$Passed;Facts=$Facts})
    [Console]::Out.WriteLine($(if($Passed){'PASS: '}else{'FAIL: '})+$Name)
    $script:CurrentGroup=$null
}
function LS-ClosedWriter {
    $memory=[IO.MemoryStream]::new()
    $writer=[IO.StreamWriter]::new($memory,[Text.UTF8Encoding]::new($false))
    $writer.Dispose()
    return $writer
}
# Test-only observation consumers. They cannot create lifecycle observations or
# grant acknowledgements; capture retains no opaque Ack in portable evidence.
function LS-Step($Session,[int]$Ordinal,$PreviousAck){
    $result=Invoke-LSStep $Session $Ordinal $PreviousAck
    LS-Details Step ([ordered]@{View=(Get-LSView $Session);Result=[ordered]@{Accepted=$result.Accepted;Rejection=$result.Rejection;AckPresent=($null -ne $result.Ack);Observation=$result.Observation;Error=$result.Error}})
    Capture-StepError $Session $result
    return $result
}
function LS-Write($Session,[string]$SinkId,$Value){
    $result=Write-LSLedger $Session $SinkId $Value
    LS-Details Write ([ordered]@{View=(Get-LSView $Session);Result=$result})
    return $result
}
function LS-Read($Session,[string]$SinkId){
    $result=Read-LSLedger $Session $SinkId
    LS-Details Read ([ordered]@{View=(Get-LSView $Session);Result=$result})
    return $result
}
function LS-Complete($Session,[string]$SinkId,[int]$RowIndex,$Row){
    $result=Complete-LSLedger $Session $SinkId $RowIndex $Row
    LS-Details Complete ([ordered]@{View=(Get-LSView $Session);Result=$result})
    return $result
}
function LS-Output($Session,[string]$SinkId,[string]$Text){
    $result=Write-LSOutput $Session $SinkId $Text
    LS-Details Output ([ordered]@{View=(Get-LSView $Session);Result=$result})
    return $result
}
