param([string]$TracePath)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$hsIdentity=[Security.Principal.WindowsIdentity]::GetCurrent()
$hsAdmin=([Security.Principal.WindowsPrincipal]::new($hsIdentity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
[Console]::Out.WriteLine('CONTEXT: '+$hsIdentity.Name+' admin='+$hsAdmin)
[Console]::Out.WriteLine('HOST: PS'+$PSVersionTable.PSVersion.ToString()+' Core x64')
$script:hsLegacy=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../p2b-inbox-tree-copy-repair'))
. (Join-Path $script:hsLegacy 'test-helpers.ps1')
$script:hsGroup=''
$script:hsRecord=[pscustomobject][ordered]@{Schema=1;Mode='p2b-handshake-seam-tests';Groups=[string[]]@();Observations=[pscustomobject][ordered]@{};Passed=$false;Error=$null;CompileCount=[int]0;ProductChanged=$false;RealGateImplemented=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false}
function HS-Need([bool]$Condition){if(-not $Condition){throw ('ASSERTION FAILED: P2BHSeam'+$script:hsGroup)}}
function HS-Case([string]$Name,[scriptblock]$Body){
    $script:hsGroup=$Name;$script:ibGroup='Seam'+$Name
    & $Body
    $script:hsRecord.Groups=[string[]]@($script:hsRecord.Groups + $Name)
    [Console]::Out.WriteLine('PASS: P2BHSeam'+$Name)
}
function HS-Expected([byte]$Sender,[byte]$Receiver,[uint64]$Next,[uint64]$Arm=0,[uint64]$Stop=0,[uint64]$Last=0){
    [pscustomobject][ordered]@{Schema=[int]1;Sender=$Sender;Receiver=$Receiver;BindingDigest=(IB-Hex '000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F');SessionId=(IB-Hex 'A0A1A2A3A4A5A6A7A8A9AAABACADAEAF');NextSequence=$Next;OriginTick=[long]0;Frequency=[long]1000;DeadlineTick=[long]120000;StopAtTick=[long]113500;ArmSequence=$Arm;StopSequence=$Stop;LastNonHeartbeatSequence=$Last}
}
function HS-AssertCodec([object]$Result,[string]$Reason,[object]$Next=$null){
    $names=[string[]]@('Schema','Mode','Accepted','Reason','Bytes','Message','NextSequence','SourceVerified','LiveActionsPerformed','RuntimeApproved')
    HS-Need ((IB-Type $Result) -eq [Management.Automation.PSCustomObject] -and (IB-Eq ([string[]]@($Result.PSObject.Properties.Name)) $names))
    HS-Need ((IB-Type $Result.Schema) -eq [int] -and $Result.Schema -eq 1 -and (IB-Type $Result.Mode) -eq [string] -and $Result.Mode -ceq 'p2b-codec-only')
    HS-Need ((IB-Type $Result.Accepted) -eq [bool] -and $Result.Accepted -eq ($Reason -cin @('decoded','binding-matched')) -and (IB-Type $Result.Reason) -eq [string] -and $Result.Reason -ceq $Reason)
    HS-Need ($null -eq $Result.Bytes -and (IB-Eq $Result.NextSequence $Next))
    if($Reason -ceq 'decoded'){HS-Need ((IB-Type $Result.Message) -eq [Management.Automation.PSCustomObject])}else{HS-Need ($null -eq $Result.Message)}
    foreach($name in @('SourceVerified','LiveActionsPerformed','RuntimeApproved')){HS-Need ((IB-Type $Result.$name) -eq [bool] -and -not $Result.$name)}
}
function HS-Codec([ValidateSet('Decode','Check')][string]$Action,[object]$Value,[object]$Expected=$null){
    $r=& $script:hsModule {param($op,$item,$expect)if($op -ceq 'Decode'){ConvertFrom-P2BControlFrame $item}else{Test-P2BControlBinding $item $expect}} $Action $Value $Expected
    if($Action -ceq 'Decode'){
        HS-AssertCodec $r decoded
        # Independent wire reads from the controlled literal input, not codec helpers.
        HS-Need ([BitConverter]::IsLittleEndian)
        $length=[int][BitConverter]::ToUInt16($Value,104)
        $payload=[byte[]]::new($length);[Array]::Copy($Value,112,$payload,0,$length)
        $want=[pscustomobject][ordered]@{Version=[uint16]1;Kind=[byte]$Value[6];Sender=[byte]$Value[7];Receiver=[byte]$Value[8];BindingDigest=(IB-Hex '000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F');SessionId=(IB-Hex 'A0A1A2A3A4A5A6A7A8A9AAABACADAEAF');Sequence=[BitConverter]::ToUInt64($Value,64);OriginTick=[long]0;Frequency=[long]1000;DeadlineTick=[long]120000;ObservedTick=[BitConverter]::ToInt64($Value,96);Payload=$payload}
        HS-Need (IB-Eq $r.Message $want)
    }
    return $r
}
try{
    $script:hsModule=Import-Module (Join-Path $script:hsLegacy 'inbox.psm1') -PassThru -DisableNameChecking
    $script:ibModule=$script:hsModule
    $script:ibLiterals=Get-Content -LiteralPath (Join-Path $script:hsLegacy 'fixtures.json') -Raw|ConvertFrom-Json
    HS-Case FixedBoundary {
        HS-Need ($PSVersionTable.PSVersion.ToString() -ceq '7.6.5' -and $PSVersionTable.PSEdition -ceq 'Core' -and [IntPtr]::Size -eq 8 -and -not $hsAdmin)
        $exports=[string[]]@($script:hsModule.ExportedFunctions.Keys|Sort-Object)
        HS-Need (IB-Eq $exports ([string[]]@($script:ibNames|Sort-Object)))
        $pins=[pscustomobject][ordered]@{Module=(Get-FileHash -LiteralPath (Join-Path $script:hsLegacy 'inbox.psm1')).Hash;Helper=(Get-FileHash -LiteralPath (Join-Path $script:hsLegacy 'test-helpers.ps1')).Hash;Fixture=(Get-FileHash -LiteralPath (Join-Path $script:hsLegacy 'fixtures.json')).Hash}
        HS-Need ($pins.Module -ceq '47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC' -and $pins.Helper -ceq '4F418D2BE07ACE9597CE7A90438E7325F692A3C578010580548299016A2A5E8D' -and $pins.Fixture -ceq '5045D321677C5FC3ABFA8674CF23CA9062C86A92D64E0B596001564E303D2A8D')
        $script:hsRecord.Observations|Add-Member NoteProperty FixedBoundary ([pscustomobject][ordered]@{Exports=$exports;Pins=$pins})
    }
    HS-Case OrdinaryGap {
        $r=IB-Result (New-P2BInbox (IB-Binding)) created;$t=$r.Inbox
        $f1=IB-Frame 1 1 1;$f2=IB-Frame 1 2 8 1;$f3=IB-Frame 1 3 8 1;$f4=IB-Frame 1 4 2
        $script:hsVector=[object[]]@($f1,$f2,$f3,$f4)
        $wantReasons=[string[]]@('queued','heartbeat-updated','heartbeat-updated','queued');$reasons=[string[]]::new(4)
        for($i=0;$i -lt 4;$i++){$a=IB-Result (Add-P2BInboxFragment $t ([byte]1) $script:hsVector[$i]) $wantReasons[$i] ([byte]1) 256;$reasons[$i]=$a.Reason}
        $before=(IB-Result (Get-P2BInboxSnapshot $t) inspected).Snapshot
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 5 4 ([object[]]@($f1,$f4)) $f3
        HS-Need (IB-Eq $before (IB-Expected $ep))
        $hb=HS-Codec Decode $before.Endpoints[0].Heartbeat
        $d1=IB-Result (Receive-P2BInboxFrame $t ([byte]1)) dequeued ([byte]1)
        $d4=IB-Result (Receive-P2BInboxFrame $t ([byte]1)) dequeued ([byte]1)
        HS-Need ((IB-Eq $d1.Frame $f1) -and (IB-Eq $d4.Frame $f4))
        $script:hsTaken4=$d4.Frame
        $null=IB-Result (Receive-P2BInboxFrame $t ([byte]1)) empty ([byte]1)
        $after=(IB-Result (Get-P2BInboxSnapshot $t) inspected).Snapshot
        $ep[0]=IB-Endpoint 1 5 4 -HB $f3
        HS-Need (IB-Eq $after (IB-Expected $ep))
        $script:hsRecord.Observations|Add-Member NoteProperty OrdinaryGap ([pscustomobject][ordered]@{EnqueuedReasons=$reasons;BeforeCount=$before.Endpoints[0].Count;BeforeNext=$before.Endpoints[0].NextSequence;BeforeLast=$before.Endpoints[0].LastNonHeartbeatSequence;HeartbeatSequence=$hb.Message.Sequence;DeliveredSequences=[uint64[]]@([BitConverter]::ToUInt64($d1.Frame,64),[BitConverter]::ToUInt64($d4.Frame,64));AfterCount=$after.Endpoints[0].Count;AfterNext=$after.Endpoints[0].NextSequence;AfterLast=$after.Endpoints[0].LastNonHeartbeatSequence})
    }
    HS-Case TakeDoesNotResetAdmission {
        $r=IB-Result (New-P2BInbox (IB-Binding)) created;$t=$r.Inbox
        $reasons=@('queued','heartbeat-updated','heartbeat-updated','queued')
        for($i=0;$i -lt 4;$i++){$null=IB-Result (Add-P2BInboxFragment $t ([byte]1) $script:hsVector[$i]) $reasons[$i] ([byte]1) 256}
        foreach($i in @(0,3)){$d=IB-Result (Receive-P2BInboxFrame $t ([byte]1)) dequeued ([byte]1);HS-Need (IB-Eq $d.Frame $script:hsVector[$i])}
        $bad=IB-Result (Add-P2BInboxFragment $t ([byte]1) $script:hsVector[3]) sequence-mismatch ([byte]1) 256
        $s=(IB-Result (Get-P2BInboxSnapshot $t) inspected).Snapshot
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 5 4 -HB $script:hsVector[2] -Partial $script:hsVector[3]
        HS-Need (IB-Eq $s (IB-Expected $ep (IB-Failure sequence-mismatch ([byte]1) sequence ([uint64]4))))
        $script:hsRecord.Observations|Add-Member NoteProperty TakeDoesNotResetAdmission ([pscustomobject][ordered]@{Reason=$bad.Reason;AcceptedBytes=$bad.AcceptedBytes;FirstFailureSequence=$s.FirstFailure.Sequence;Next=$s.Endpoints[0].NextSequence;Last=$s.Endpoints[0].LastNonHeartbeatSequence;PartialLength=$s.Endpoints[0].PartialLength})
    }
    HS-Case StaleExpectedRejected {
        $script:hsDecoded4=(HS-Codec Decode $script:hsTaken4).Message
        $e=HS-Expected 1 2 2;$r=HS-Codec Check $script:hsDecoded4 $e
        HS-AssertCodec $r sequence-mismatch
        $script:hsRecord.Observations|Add-Member NoteProperty StaleExpectedRejected ([pscustomobject][ordered]@{ConsumedSequence=$script:hsDecoded4.Sequence;ExpectedNext=$e.NextSequence;Reason=$r.Reason;ReturnedNext=$r.NextSequence})
    }
    HS-Case DeclaredExpectedIsNotAuthority {
        # Independent fixture value 4; never a production admission cursor.
        $e=HS-Expected 1 2 4;$r=HS-Codec Check $script:hsDecoded4 $e
        HS-AssertCodec $r binding-matched ([uint64]5)
        $script:hsRecord.Observations|Add-Member NoteProperty DeclaredExpectedIsNotAuthority ([pscustomobject][ordered]@{ExpectedNext=$e.NextSequence;Reason=$r.Reason;ReturnedNext=$r.NextSequence;Flags=[bool[]]@($r.SourceVerified,$r.LiveActionsPerformed,$r.RuntimeApproved)})
    }
    HS-Case ArmReference {
        $m=(HS-Codec Decode (IB-Frame 1 7 5)).Message
        $values=[uint64[]]@(3,0,2);$reasons=[string[]]::new(3);$next=[object[]]::new(3)
        for($i=0;$i -lt 3;$i++){
            $r=HS-Codec Check $m (HS-Expected 1 2 7 -Arm $values[$i])
            if($i -eq 0){HS-AssertCodec $r binding-matched ([uint64]8)}else{HS-AssertCodec $r reference-mismatch}
            $reasons[$i]=$r.Reason;$next[$i]=$r.NextSequence
        }
        $reference=[BitConverter]::ToUInt64($m.Payload,8);HS-Need ($reference -eq [uint64]3)
        $script:hsRecord.Observations|Add-Member NoteProperty ArmReference ([pscustomobject][ordered]@{Sequence=$m.Sequence;PayloadReference=$reference;ExpectedReferences=$values;Reasons=$reasons;ReturnedNext=$next})
    }
    HS-Case StopReference {
        $m=(HS-Codec Decode (IB-Frame 1 7 10)).Message
        $values=[uint64[]]@(2,0,1);$reasons=[string[]]::new(3);$next=[object[]]::new(3)
        for($i=0;$i -lt 3;$i++){
            $r=HS-Codec Check $m (HS-Expected 1 2 7 -Stop $values[$i])
            if($i -eq 0){HS-AssertCodec $r binding-matched ([uint64]8)}else{HS-AssertCodec $r reference-mismatch}
            $reasons[$i]=$r.Reason;$next[$i]=$r.NextSequence
        }
        $reference=[BitConverter]::ToUInt64($m.Payload,8);HS-Need ($reference -eq [uint64]2)
        $script:hsRecord.Observations|Add-Member NoteProperty StopReference ([pscustomobject][ordered]@{Sequence=$m.Sequence;PayloadReference=$reference;ExpectedReferences=$values;Reasons=$reasons;ReturnedNext=$next})
    }
    HS-Case HeartbeatReference {
        $m=(HS-Codec Decode (IB-Frame 1 2 8 1)).Message
        $values=[uint64[]]@(1,0);$reasons=[string[]]::new(2);$next=[object[]]::new(2)
        for($i=0;$i -lt 2;$i++){
            $r=HS-Codec Check $m (HS-Expected 1 2 2 -Last $values[$i])
            if($i -eq 0){HS-AssertCodec $r binding-matched ([uint64]3)}else{HS-AssertCodec $r reference-mismatch}
            $reasons[$i]=$r.Reason;$next[$i]=$r.NextSequence
        }
        $reference=[BitConverter]::ToUInt64($m.Payload,8);HS-Need ($reference -eq [uint64]1)
        $script:hsRecord.Observations|Add-Member NoteProperty HeartbeatReference ([pscustomobject][ordered]@{Sequence=$m.Sequence;PayloadReference=$reference;ExpectedReferences=$values;Reasons=$reasons;ReturnedNext=$next})
    }
    HS-Case ArmDeclaredExpiry {
        $values=[long[]]@(113499,113500);$observed=[long[]]::new(2);$reasons=[string[]]::new(2);$next=[object[]]::new(2);$arm=$null
        for($i=0;$i -lt 2;$i++){
            $f=IB-Frame 2 1 4;IB-Put $f 96 8 ([uint64]$values[$i]);$m=(HS-Codec Decode $f).Message;$arm=$m
            $r=HS-Codec Check $m (HS-Expected 2 1 1)
            if($i -eq 0){HS-AssertCodec $r binding-matched ([uint64]2)}else{HS-AssertCodec $r expired-declaration}
            $observed[$i]=$m.ObservedTick;$reasons[$i]=$r.Reason;$next[$i]=$r.NextSequence
        }
        $script:hsRecord.Observations|Add-Member NoteProperty ArmDeclaredExpiry ([pscustomobject][ordered]@{ObservedTicks=$observed;ArmReference=[BitConverter]::ToUInt64($arm.Payload,8);ReleaseBefore=[BitConverter]::ToUInt64($arm.Payload,16);Reasons=$reasons;ReturnedNext=$next})
    }
    HS-Case LookalikeInboxRejected {
        $r=IB-Result (New-P2BInbox (IB-Binding)) created;$t=$r.Inbox
        $fakes=[object[]]@([object]::new(),[pscustomobject]@{Inbox=$t;Accepted=$true},[Management.Automation.PSSerializer]::Deserialize([Management.Automation.PSSerializer]::Serialize($t)))
        $reasons=[string[]]::new(3)
        for($i=0;$i -lt 3;$i++){$bad=IB-Result (Get-P2BInboxSnapshot $fakes[$i]) invalid-input;$reasons[$i]=$bad.Reason}
        $s=(IB-Result (Get-P2BInboxSnapshot $t) inspected).Snapshot;$same=IB-Eq $s (IB-Expected (IB-Initial));HS-Need $same
        $script:hsRecord.Observations|Add-Member NoteProperty LookalikeInboxRejected ([pscustomobject][ordered]@{Reasons=$reasons;OriginalSnapshotUnchanged=$same;FakeCount=$fakes.GetLength(0)})
    }
    HS-Case StoppedInboxNoRecovery {
        $createdCount=0;$r=IB-Result (New-P2BInbox (IB-Binding)) created;$createdCount++;$t=$r.Inbox
        $f=IB-Frame 1 1 1;$null=IB-Result (Add-P2BInboxFragment $t ([byte]1) $f) queued ([byte]1) 256
        $closed=IB-Result (Close-P2BInboxInput $t ([byte]1)) peer-closed ([byte]1)
        $postFeed=IB-Result (Add-P2BInboxFragment $t ([byte]1) (IB-Frame 1 2 10)) stopped
        $postTake=IB-Result (Receive-P2BInboxFrame $t ([byte]1)) stopped
        $s=(IB-Result (Get-P2BInboxSnapshot $t) inspected).Snapshot
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 2 1 ([object[]]@(,$f))
        HS-Need ($createdCount -eq 1 -and (IB-Eq $s (IB-Expected $ep (IB-Failure peer-closed ([byte]1) eof))))
        $script:hsRecord.Observations|Add-Member NoteProperty StoppedInboxNoRecovery ([pscustomobject][ordered]@{CreatedCount=$createdCount;CloseReason=$closed.Reason;PostFeed=$postFeed.Reason;PostTake=$postTake.Reason;FirstReason=$s.FirstFailure.Reason;FirstSequence=$s.FirstFailure.Sequence;Count=$s.Endpoints[0].Count;Next=$s.Endpoints[0].NextSequence;Last=$s.Endpoints[0].LastNonHeartbeatSequence})
    }
    HS-Case Complete {
        $prior=$script:hsRecord.Groups.Count
        $script:hsRecord.ProductChanged=(Get-FileHash -LiteralPath (Join-Path $script:hsLegacy 'inbox.psm1')).Hash -cne '47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC'
        $exports=[string[]]@($script:hsModule.ExportedFunctions.Keys|Sort-Object)
        HS-Need ($prior -eq 11 -and $script:hsRecord.CompileCount -eq 0 -and -not $script:hsRecord.ProductChanged -and -not $script:hsRecord.RealGateImplemented -and (IB-Eq $exports ([string[]]@($script:ibNames|Sort-Object))))
        foreach($name in @('SourceVerified','LiveActionsPerformed','RuntimeApproved')){HS-Need ((IB-Type $script:hsRecord.$name) -eq [bool] -and -not $script:hsRecord.$name)}
        $script:hsRecord.Observations|Add-Member NoteProperty Complete ([pscustomobject][ordered]@{PriorGroups=$prior;Exports=$exports})
    }
    $script:hsRecord.Passed=$true
    [Console]::Out.WriteLine('LIMIT: declarations/seam only; no handshake state or runtime authority')
}catch{
    $script:hsRecord.Error=[pscustomobject][ordered]@{Message=$_.Exception.Message;Type=$_.Exception.GetType().FullName;Position=$_.InvocationInfo.PositionMessage;Stack=$_.ScriptStackTrace}
    [Console]::Out.WriteLine($_.Exception.Message)
}finally{
    if(-not [string]::IsNullOrEmpty($TracePath)){[IO.File]::WriteAllText($TracePath,($script:hsRecord|ConvertTo-Json -Depth 30),[Text.UTF8Encoding]::new($false))}
}
if(-not $script:hsRecord.Passed){exit 1}
