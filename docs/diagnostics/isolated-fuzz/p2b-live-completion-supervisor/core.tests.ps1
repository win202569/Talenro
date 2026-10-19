# Test utilities, not exported controller hooks. Expectations are literal facts.
function LS-Need([bool]$Condition,[string]$Message){if(-not $Condition){throw ('ASSERT: '+$Message)}}
function LS-New([string]$Case){New-LSSession $Grant $Case $FrozenSHA256}
function LS-Path($Session,[string]$Name){Join-Path (Get-LSView $Session).Root $Name}
function LS-Details([string]$Name,$Value){$script:Details.Add([ordered]@{Name=$Name;Value=$Value})}
function LS-Result([string]$Name,[bool]$Passed,$Facts){
    $script:Groups.Add([ordered]@{Name=$Name;Passed=$Passed;Facts=$Facts})
    [Console]::Out.WriteLine($(if($Passed){'PASS: '}else{'FAIL: '})+$Name)
}
function LS-ClosedWriter {
    $memory=[IO.MemoryStream]::new()
    $writer=[IO.StreamWriter]::new($memory,[Text.UTF8Encoding]::new($false))
    $writer.Dispose()
    return $writer
}
function Invoke-LSCoreTests {
    $s=LS-New disk
    $path=LS-Path $s ledger.json
    [IO.File]::WriteAllText($path,'{"Outcome":"EXPECTED_GREEN"}')
    $before=[IO.File]::ReadAllText($path)
    $advance=Test-LSAdvance $s 1 $null
    $diskSame=[IO.File]::ReadAllText($path) -ceq $before
    $v=Get-LSView $s
    LS-Details disk $v
    $name=if($Grant -ceq 'core-red'){'DiskMustNotMintAck'}else{'DiskCannotMintAck'}
    LS-Result $name (-not $advance -and $diskSame -and $v.StartCalls -eq 0) ([ordered]@{AdvanceAccepted=$advance;FileUnchanged=$diskSame;StartCalls=$v.StartCalls})
    $null=Close-LSSession $s Stop

    $s=LS-New sink
    $path=LS-Path $s ledger.json
    $null=[IO.Directory]::CreateDirectory($path)
    $first=Write-LSLedger $s Ledger @{Rows=@()}
    $v=Get-LSView $s
    LS-Need (-not $first.Ok -and $null -ne $first.Error) 'actual first file error required'
    $count=$v.Sinks.Ledger.LogicalIOAttempts
    $r1=Read-LSLedger $s Ledger
    $r2=Complete-LSLedger $s Ledger 0 @{}
    $r3=Write-LSLedger $s Ledger @{Rows=@()}
    $v=Get-LSView $s
    $delta=$v.Sinks.Ledger.LogicalIOAttempts-$count
    $same=[object]::ReferenceEquals($v.PrimaryError,$first.Error)
    LS-Details sink $v
    $name=if($Grant -ceq 'core-red'){'FailedSinkMustNotBeReused'}else{'FailedSinkRejectsAccess'}
    LS-Result $name ($delta -eq 0 -and $same -and -not $r1.Ok -and -not $r2.Ok -and -not $r3.Ok) ([ordered]@{FirstFailed=(-not $first.Ok);PostFailureLogicalIO=$delta;PrimarySame=$same;StartCalls=$v.StartCalls})
    $null=Close-LSSession $s Stop
    if($Grant -ceq 'core-red'){return}

    $invalid=[Collections.Generic.List[object]]::new()
    foreach($case in @('missing','bad','index')){
        $s=LS-New $case
        $path=LS-Path $s ledger.json
        if($case -ceq 'bad'){[IO.File]::WriteAllText($path,'{')}
        if($case -ceq 'index'){[IO.File]::WriteAllText($path,'{"Rows":[]}')}
        $r=if($case -ceq 'index'){Complete-LSLedger $s Ledger 0 @{}}else{Read-LSLedger $s Ledger}
        $v=Get-LSView $s
        LS-Need (-not $r.Ok -and $null -ne $r.Error) ('actual error '+$case)
        $invalid.Add([ordered]@{Case=$case;Failed=$v.FailureLatched;AckIssued=$v.AckIssued;StartCalls=$v.StartCalls})
        LS-Details $case $v
        $null=Close-LSSession $s Stop
    }
    LS-Result InvalidLedgerStops ($invalid.Count -eq 3 -and @($invalid | Where-Object {-not $_.Failed -or $_.AckIssued -ne 0 -or $_.StartCalls -ne 0}).Count -eq 0) ([ordered]@{Cases=$invalid.ToArray()})

    $s=LS-New errors
    $path=LS-Path $s ledger.json
    $null=[IO.Directory]::CreateDirectory($path)
    $first=Write-LSLedger $s Ledger @{}
    Add-LSFailure $s $first.Error
    # Independent real errors: do not re-enter any failed managed sink.
    for($i=0;$i -lt 9;$i++){
        $actual=$null
        try{[IO.File]::ReadAllText((LS-Path $s ('absent-'+$i)))}catch{$actual=$_}
        LS-Need ($actual -is [Management.Automation.ErrorRecord]) 'independent actual error'
        Add-LSFailure $s $actual
    }
    $v=Get-LSView $s
    $same=[object]::ReferenceEquals($first.Error,$v.PrimaryError)
    LS-Details errors $v
    LS-Result FirstAndSecondaryBounded ($same -and $v.SecondaryErrors.Count -eq 8 -and $v.Overflow) ([ordered]@{PrimarySame=$same;SecondaryCount=$v.SecondaryErrors.Count;Overflow=$v.Overflow})
    $null=Close-LSSession $s Stop

    $s=LS-New output
    $saved=[Console]::Out
    $closed=LS-ClosedWriter
    try{[Console]::SetOut($closed);$r=Write-LSOutput $s ConsoleOut 'must fail'}finally{[Console]::SetOut($saved)}
    $v=Get-LSView $s
    LS-Details output $v
    LS-Result OutputFailureNoAck (-not $r.Ok -and $v.FailureLatched -and $v.AckIssued -eq 0) ([ordered]@{Failed=$v.FailureLatched;AckIssued=$v.AckIssued;StartCalls=$v.StartCalls})
    $null=Close-LSSession $s Stop

    $s=LS-New preconsume
    $null=[IO.Directory]::CreateDirectory((LS-Path $s start.json))
    $r=Invoke-LSStep $s 0 $null
    $v=Get-LSView $s
    $again=Invoke-LSStep $s 0 $null
    $after=Get-LSView $s
    LS-Details preconsume $after
    LS-Result PreconsumeFailureNoStart ($r.Accepted -and $v.Reserved -eq 1 -and $v.StartCalls -eq 0 -and $v.FailureLatched -and -not $again.Accepted -and $after.Sinks.Start.LogicalIOAttempts -eq 1) ([ordered]@{Reserved=$after.Reserved;StartCalls=$after.StartCalls;Failed=$after.FailureLatched;StartIO=$after.Sinks.Start.LogicalIOAttempts;SecondAccepted=$again.Accepted})
    $null=Close-LSSession $s Stop

    $s=LS-New copies
    $v=Get-LSView $s
    $fake=@($v,($v | ConvertTo-Json -Depth 20 | ConvertFrom-Json),[pscustomobject]@{Verified=$true;Grant=$Grant;Case='copies'})
    $rejected=0
    foreach($copy in $fake){if(-not (Test-LSAdvance $copy 0 $null)){$rejected++};LS-Need (-not (Invoke-LSStep $copy 0 $null).Accepted) 'copied session launch rejected'}
    LS-Result HandleCopiesRejected ($rejected -eq 3 -and (Get-LSView $s).StartCalls -eq 0) ([ordered]@{Rejected=$rejected;StartCalls=(Get-LSView $s).StartCalls})
    $null=Close-LSSession $s Stop
}
