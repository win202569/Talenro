# Pure contract validation. No native calls, receipt enrichment or acceptance.
function Test-OrdinaryLedgerEvidence([object]$Receipt) {
    function Need($o,[string]$name,[string]$kind) {
        if($null -eq $o -or $null -eq $o.PSObject.Properties[$name]){throw ('Missing '+$name)}
        $v=$o.$name
        switch($kind){
            'bool' {if($v -isnot [bool]){throw ('Boolean type '+$name)}}
            'int' {if($null -eq $v -or $v.GetType().FullName -notin @('System.Byte','System.SByte','System.Int16','System.UInt16','System.Int32','System.UInt32','System.Int64','System.UInt64')){throw ('Integer type '+$name)}}
            'text' {if($v -isnot [string] -or [string]::IsNullOrWhiteSpace($v)){throw ('String type '+$name)}}
            'array' {if($v -isnot [System.Collections.IList]){throw ('Array type '+$name)}}
            'object' {if($null -eq $v -or $v -is [string] -or $v -is [ValueType] -or $v -is [System.Collections.IList]){throw ('Object type '+$name)}}
        }
    }
    function Num($o,[string]$n,[decimal]$min=0,[decimal]$max=9223372036854775807){Need $o $n 'int';if([decimal]$o.$n -lt $min -or [decimal]$o.$n -gt $max){throw ('Integer range '+$n)}}
    function Flag($o,[string]$n,[bool]$want){Need $o $n 'bool';if($o.$n -ne $want){throw ('Flag '+$n)}}
    function Text($o,[string]$n,[string]$want){Need $o $n 'text';if($o.$n -cne $want){throw ('Value '+$n)}}
    function Obj($o,[string]$n){Need $o $n 'object';foreach($raw in @('Handle','ProcessHandle','JobHandle')){if($null -ne $o.$n.PSObject.Properties[$raw]){throw 'Raw handle field'}}}
    function Empty($o,[string]$n){Need $o $n 'array';if($o.$n.Count -ne 0){throw ('Nonempty '+$n)}}
    function Utc-Metadata($o,[string]$n){
        if($null -eq $o.PSObject.Properties[$n]){throw ('Missing '+$n)}
        $v=$o.$n
        # ConvertFrom-Json on newer hosts materializes ISO dates as DateTime.
        # Only UTC metadata permits this representation; identity/flags stay strict.
        if($v -is [DateTime]){if($v.Kind -ne [DateTimeKind]::Utc){throw ('Non-UTC '+$n)};return}
        if($v -is [DateTimeOffset]){if($v.Offset -ne [TimeSpan]::Zero){throw ('Non-UTC '+$n)};return}
        Need $o $n 'text';$parsed=[DateTimeOffset]::MinValue
        if(-not [DateTimeOffset]::TryParse($v,[Globalization.CultureInfo]::InvariantCulture,[Globalization.DateTimeStyles]::None,[ref]$parsed) -or $parsed.Offset -ne [TimeSpan]::Zero){throw ('UTC metadata '+$n)}
    }
    function Bounds($o){
        Num $o 'StartTimestamp';Num $o 'EndTimestamp';Num $o 'Frequency' 1
        Utc-Metadata $o 'StartUTC';Utc-Metadata $o 'EndUTC'
        if($o.Frequency -ne $frame.Frequency -or $o.EndTimestamp -lt $o.StartTimestamp -or $o.EndTimestamp -gt $frame.EndTimestamp){throw 'API timing'}
    }
    function Api($o){Flag $o 'Succeeded' $true;Num $o 'Error' 0 0;Bounds $o}
    function Query($o){Api $o;Num $o 'Assigned' 0 4294967295;Num $o 'Returned' 0 4294967295;Num $o 'Capacity' 1 1048576;Num $o 'BufferBytes' 1 16777216;if($o.Assigned -ne $o.Returned -or $o.Returned -gt $o.Capacity -or $o.BufferBytes -lt 8+8*[decimal]$o.Capacity){throw 'List counts'}}
    function Read-State($read,$entry,[bool]$final){
        Text $read 'RunIdentity' $Receipt.RunIdentity;Num $read 'EntryId' 1 32
        if($read.EntryId -ne $entry.EntryId){throw 'Read entry reference'}
        Flag $read 'SameHandleVerified' $true;Flag $read 'Attempted' $true;Flag $read 'ExitCodeRead' $true;Flag $read 'TerminationRequested' $false
        Need $read 'Signaled' 'bool';Num $read 'ExitCode' 0 4294967295;Num $read 'Error' 0 0;Num $read 'ImageError' 0 2147483647;Need $read 'State' 'text';Bounds $read
        if($read.State -cnotin @('running','exited') -or $read.Signaled -ne ($read.State -ceq 'exited')){throw 'Wait/state contradiction'}
        if(-not [string]::IsNullOrEmpty($read.SkipReason)){throw 'Skipped read'}
        if($final -and (-not $read.Signaled -or $read.StartTimestamp -lt $frame.ShutdownStartTimestamp -or $read.StartTimestamp -lt $entry.Initial.EndTimestamp)){throw 'Final state/timing'}
    }
    $errors=@();$complete=$false
    try{
        if($null -eq $Receipt){throw 'Missing envelope'}
        foreach($name in @('LifetimeDiagnostics','CompileStageTrigger','BuildOnlyDiagnostic','Handle')){if($null -ne $Receipt.PSObject.Properties[$name]){throw 'Diagnostic/raw contamination'}}
        foreach($name in @('MemberEvidence','ShutdownMemberEvidence')){if($null -ne $Receipt.PSObject.Properties[$name]){Empty $Receipt $name}}
        Num $Receipt 'ShutdownSchema' 4 4;Text $Receipt 'ShutdownStatePolicy' 'ordinary-owned-ledger-final-exit-v1';Need $Receipt 'RunIdentity' 'text';Num $Receipt 'RootId' 1 4294967295
        foreach($n in @('RootAssigned','RootExited','CleanupVerified','NaturalShutdownVerified','ShutdownEvidenceComplete')){Flag $Receipt $n $true}
        foreach($n in @('TimedOut','TerminationRequested','LifetimeDiagnosticOnly','ShutdownDiagnosticOnly','LiveSampling','MemberEvidenceTruncated')){Flag $Receipt $n $false}
        Num $Receipt 'ActiveAfterCleanup' 0 0;Num $Receipt 'ActiveBeforeTermination' 0 0;Empty $Receipt 'MembersAfterCleanup';Empty $Receipt 'MembersBeforeTermination'
        Num $Receipt 'NaturalObservationMs' 0 499;Num $Receipt 'ShutdownElapsedMs' 0 5000
        Obj $Receipt 'LedgerFinalOutcome';$frame=$Receipt.LedgerFinalOutcome
        Flag $frame 'EvidenceComplete' $true;Flag $frame 'NaturalShutdownVerified' $true;Flag $frame 'CountReconciled' $true;Empty $frame 'Errors';Empty $frame 'FinalMembers'
        Num $frame 'ShutdownStartTimestamp';Num $frame 'EndTimestamp';Num $frame 'Frequency' 1;Num $frame 'NaturalObservationMs' 0 499
        if($frame.EndTimestamp -lt $frame.ShutdownStartTimestamp){throw 'Shutdown clock order'}
        $ms=([decimal]$frame.EndTimestamp-[decimal]$frame.ShutdownStartTimestamp)*1000/[decimal]$frame.Frequency
        if($ms -ge 500 -or [math]::Floor($ms) -ne $frame.NaturalObservationMs -or $Receipt.NaturalObservationMs -ne $frame.NaturalObservationMs -or $Receipt.ShutdownElapsedMs -lt $Receipt.NaturalObservationMs){throw 'Shared time window'}
        Obj $Receipt 'MemberIdentityLedger';$ledger=$Receipt.MemberIdentityLedger
        Num $ledger 'Schema' 1 1;Text $ledger 'RunIdentity' $Receipt.RunIdentity;Text $ledger 'Mode' 'ordinary-identity-ledger-v1';Text $ledger 'Coverage' 'sampled';Num $ledger 'RootId' 1 4294967295
        if($ledger.RootId -ne $Receipt.RootId){throw 'Root reference'}
        foreach($n in @('Incomplete','Ambiguous','CapExhausted')){Flag $ledger $n $false}
        Flag $ledger 'TeardownAttempted' $true;Flag $ledger 'Sealed' $true;Empty $ledger 'Errors'
        Num $ledger 'Attempts' 0 31;Num $ledger 'Admitted' 0 31;Num $ledger 'QueryCount' 1 2147483647;Num $ledger 'ReleaseAttempts' 0 31;Num $ledger 'ReleaseFailures' 0 0;Need $ledger 'Entries' 'array'
        if($ledger.Attempts -ne $ledger.Admitted -or $ledger.Admitted -ne $ledger.Entries.Count -or $ledger.ReleaseAttempts -ne $ledger.Admitted){throw 'Ledger counts'}
        Obj $ledger 'FirstQuery';Query $ledger.FirstQuery;Obj $ledger 'LastQuery';Query $ledger.LastQuery
        if($ledger.LastQuery.StartTimestamp -lt $ledger.FirstQuery.StartTimestamp){throw 'Query order'}
        $entries=New-Object 'System.Collections.Generic.HashSet[int]';$pids=New-Object 'System.Collections.Generic.HashSet[UInt64]'
        foreach($entry in $ledger.Entries){
            foreach($raw in @('Handle','ProcessHandle','JobHandle','DiagnosticShutdown')){if($null -ne $entry.PSObject.Properties[$raw]){throw 'Raw/diagnostic entry'}}
            Text $entry 'RunIdentity' $Receipt.RunIdentity;Num $entry 'EntryId' 1 32;Num $entry 'Pid' 1 4294967295;Num $entry 'SourceQuery' 1 $ledger.QueryCount
            if($entry.Pid -eq $Receipt.RootId -or -not $entries.Add([int]$entry.EntryId) -or -not $pids.Add([UInt64]$entry.Pid)){throw 'Duplicate/invalid identity'}
            Obj $entry 'SourceQueryEvidence';Query $entry.SourceQueryEvidence
            Obj $entry 'Open';Api $entry.Open;Obj $entry 'Membership';Api $entry.Membership;Flag $entry.Membership 'Owned' $true
            Obj $entry 'Initial';Read-State $entry.Initial $entry $false;Obj $entry 'Final';Read-State $entry.Final $entry $true
            if($entry.Open.StartTimestamp -lt $entry.SourceQueryEvidence.EndTimestamp -or $entry.Membership.StartTimestamp -lt $entry.Open.EndTimestamp -or $entry.Initial.StartTimestamp -lt $entry.Membership.EndTimestamp){throw 'Acquisition call order'}
            if($entry.Initial.State -ceq 'exited' -and $entry.Initial.ExitCode -ne $entry.Final.ExitCode){throw 'Exited code changed'}
            Flag $entry 'ReleaseAttempted' $true;Flag $entry 'ReleaseSucceeded' $true;Num $entry 'ReleaseError' 0 0
        }
        Obj $frame 'FinalAccounting';$a=$frame.FinalAccounting;Api $a
        if($a.StartTimestamp -lt $frame.ShutdownStartTimestamp){throw 'Final accounting before shutdown'}
        foreach($n in @('TotalProcesses','ActiveProcesses','TotalTerminatedProcesses')){Num $a $n 0 4294967295}
        if($a.TotalProcesses -ne 1+$ledger.Admitted -or $a.ActiveProcesses -ne 0 -or $a.TotalTerminatedProcesses -ne 0){throw 'Accounting mismatch'}
        Obj $frame 'FinalMemberQuery';Query $frame.FinalMemberQuery
        if($frame.FinalMemberQuery.Returned -ne 0 -or $frame.FinalMemberQuery.StartTimestamp -lt $frame.ShutdownStartTimestamp -or $frame.FinalMemberQuery.StartTimestamp -lt $a.EndTimestamp){throw 'Final list mismatch'}
        foreach($entry in $ledger.Entries){if($a.StartTimestamp -lt $entry.Final.EndTimestamp){throw 'Final accounting order'}}
        $complete=$true
    }catch{$errors=@($_.Exception.Message)}
    [pscustomobject]@{EvidenceComplete=[bool]$complete;Acceptance=$false;Errors=[string[]]$errors}
}
