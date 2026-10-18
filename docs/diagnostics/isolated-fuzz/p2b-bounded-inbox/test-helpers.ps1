# Independent typed literals and assertions. No product-generated expected frames.
$script:ibNames=@('New-P2BInbox','Add-P2BInboxFragment','Close-P2BInboxInput','Receive-P2BInboxFrame','Get-P2BInboxSnapshot')
$script:ibGroup=''
function IB-Need([bool]$Condition){if(-not $Condition){throw ('ASSERTION FAILED: P2BInbox'+$script:ibGroup)}}
function IB-Type($v){if($null -eq $v){return $null};return [object].GetMethod('GetType',[type[]]@()).Invoke($v.PSObject.BaseObject,$null)}
function IB-Eq($a,$b){
    if($null -eq $a -or $null -eq $b){return $null -eq $a -and $null -eq $b}
    $at=IB-Type $a;$bt=IB-Type $b;if($at -ne $bt){return $false}
    if($a -is [Array]){if($a.GetLength(0) -ne $b.GetLength(0)){return $false};for($i=0;$i -lt $a.GetLength(0);$i++){if(-not (IB-Eq $a[$i] $b[$i])){return $false}};return $true}
    if($at -eq [Management.Automation.PSCustomObject]){
        $ap=@($a.PSObject.Properties);$bp=@($b.PSObject.Properties);if($ap.Count -ne $bp.Count){return $false}
        for($i=0;$i -lt $ap.Count;$i++){if($ap[$i].Name -cne $bp[$i].Name -or -not (IB-Eq $ap[$i].Value $bp[$i].Value)){return $false}};return $true
    }
    return $a -ceq $b
}
function IB-Hex([string]$s){return ,[Convert]::FromHexString($s)}
function IB-Binding {
    [pscustomobject][ordered]@{Schema=[int]1;BindingDigest=(IB-Hex '000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F');SessionId=(IB-Hex 'A0A1A2A3A4A5A6A7A8A9AAABACADAEAF');OriginTick=[long]0;Frequency=[long]1000;DeadlineTick=[long]120000}
}
function IB-Put([byte[]]$a,[int]$offset,[int]$width,[uint64]$number){
    # Test-only independent unsigned little endian conversion; no product helpers.
    $bytes=[BitConverter]::GetBytes($number);if(-not [BitConverter]::IsLittleEndian){[Array]::Reverse($bytes)}
    [Array]::Copy($bytes,0,$a,$offset,$width)
}
function IB-Frame([int]$Endpoint=1,[uint64]$Seq=1,[int]$Kind=1,[uint64]$Last=0){
    $f=$script:ibLiterals|Where-Object{$_.Kind -eq $Kind};if(@($f).Count -ne 1){throw 'Fixture kind'}
    $a=IB-Hex $f.FrameHex
    $pairs=@(@(1,2),@(2,1),@(2,3),@(3,2));$a[7]=[byte]$pairs[$Endpoint-1][0];$a[8]=[byte]$pairs[$Endpoint-1][1]
    IB-Put $a 64 8 $Seq;IB-Put $a 72 8 0;IB-Put $a 80 8 1000;IB-Put $a 88 8 120000;IB-Put $a 96 8 4
    if($Kind -eq 8){IB-Put $a 120 8 $Last}
    if($Kind -eq 4){IB-Put $a 120 8 $Seq;IB-Put $a 128 8 113500}
    return ,$a
}
function IB-Result($r,[string]$Reason,[object]$Endpoint=$null,[int]$Bytes=0){
    $success=$Reason -cin @('created','partial','queued','heartbeat-updated','empty','dequeued','inspected')
    $names=@('Schema','Mode','Accepted','Reason','Inbox','EndpointId','AcceptedBytes','Frame','Snapshot','SourceVerified','LiveActionsPerformed','RuntimeApproved')
    IB-Need ((IB-Type $r) -eq [Management.Automation.PSCustomObject] -and (IB-Eq ([string[]]@($r.PSObject.Properties.Name)) ([string[]]$names)))
    IB-Need ((IB-Type $r.Schema) -eq [int] -and $r.Schema -eq 1 -and (IB-Type $r.Mode) -eq [string] -and $r.Mode -ceq 'p2b-inbox-only')
    IB-Need ((IB-Type $r.Accepted) -eq [bool] -and $r.Accepted -eq $success -and (IB-Type $r.Reason) -eq [string] -and $r.Reason -ceq $Reason)
    IB-Need ((IB-Type $r.AcceptedBytes) -eq [int] -and $r.AcceptedBytes -eq $Bytes -and (IB-Eq $r.EndpointId $Endpoint))
    foreach($n in @('SourceVerified','LiveActionsPerformed','RuntimeApproved')){IB-Need ((IB-Type $r.$n) -eq [bool] -and -not $r.$n)}
    if($Reason -ceq 'created'){IB-Need ((IB-Type $r.Inbox) -eq [object])}else{IB-Need ($null -eq $r.Inbox)}
    if($Reason -ceq 'dequeued'){IB-Need ((IB-Type $r.Frame) -eq [byte[]] -and $r.Frame.GetLength(0) -eq 256)}else{IB-Need ($null -eq $r.Frame)}
    if($Reason -ceq 'inspected'){IB-Need ((IB-Type $r.Snapshot) -eq [Management.Automation.PSCustomObject])}else{IB-Need ($null -eq $r.Snapshot)}
    return $r
}
function IB-New { $r=IB-Result (New-P2BInbox (IB-Binding)) 'created';return $r.Inbox }
function IB-Snap($token){$r=IB-Result (Get-P2BInboxSnapshot $token) 'inspected';return $r.Snapshot}
function IB-Add($token,[byte[]]$bytes,[byte]$e=1,[string]$reason='queued'){
    $r=Add-P2BInboxFragment $token $e $bytes
    return IB-Result $r $reason $e $bytes.GetLength(0)
}
function IB-Endpoint([byte]$Id,[uint64]$Next=1,[uint64]$Last=0,[object[]]$Queue=([object[]]::new(0)),[byte[]]$HB=$null,[byte[]]$Partial=([byte[]]::new(0))){
    [pscustomobject][ordered]@{EndpointId=$Id;Count=[int]$Queue.GetLength(0);NextSequence=$Next;LastNonHeartbeatSequence=$Last;PartialBytes=$Partial;Queue=$Queue;Heartbeat=$HB;PartialLength=[int]$Partial.GetLength(0)}
}
function IB-Expected([object[]]$Endpoints,[object]$Failure=$null){
    [pscustomobject][ordered]@{Stopped=[bool]($null -ne $Failure);FirstFailure=$Failure;Binding=(IB-Binding);Endpoints=$Endpoints}
}
function IB-Initial { $a=[object[]]::new(4);for($i=0;$i -lt 4;$i++){$a[$i]=IB-Endpoint ([byte]($i+1))};return ,$a }
function IB-Failure([string]$Reason,[object]$Endpoint,[string]$Stage,[object]$Sequence=$null){[pscustomobject][ordered]@{Reason=$Reason;EndpointId=$Endpoint;Stage=$Stage;Sequence=$Sequence}}
function IB-Check($token,[object[]]$Endpoints,[object]$Failure=$null){IB-Need (IB-Eq (IB-Snap $token) (IB-Expected $Endpoints $Failure))}
function IB-PrivateSequence($token,[uint64]$Next,[uint64]$Last=0){
    # Explicit synthetic boundary, never a production setter or restored history.
    $state=& $script:ibModule {param($key) $a=[object]$null;if(-not $script:registry.TryGetValue($key,[ref]$a)){throw 'Missing test token'};return $a} $token
    IB-Need ($state.Endpoints[0].NextSequence -eq 1 -and $state.Endpoints[0].Count -eq 0)
    $script:ibTrace.Synthetic+=,[pscustomobject]@{Before=[uint64]$state.Endpoints[0].NextSequence;After=$Next;Last=$Last}
    $state.Endpoints[0].NextSequence=$Next;$state.Endpoints[0].LastNonHeartbeatSequence=$Last
}
function IB-InitDescriptors {
    Add-Type -TypeDefinition ([IO.File]::ReadAllText((Join-Path $PSScriptRoot 'descriptor-fixture.cs')))
    $script:ibTrace.CompileCount++
    [P2BDescriptorFixture.CounterNote]::Reset()
    $probe=[P2BDescriptorFixture.CounterNote]::new('Positive',[int]1)
    $null=$probe.Value;$null=$probe.MemberType;$null=$probe.Copy()
    IB-Need ([P2BDescriptorFixture.CounterNote]::Total -gt 0)
    $script:ibTrace.Observable=$true;[P2BDescriptorFixture.CounterNote]::Reset()
    $script:ibReflectionFlags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $script:ibRawField=[Management.Automation.PSMemberSet].GetField('internalMembers',$script:ibReflectionFlags)
    $script:ibCacheField=[Management.Automation.PSObject].GetField('_instanceMembers',$script:ibReflectionFlags)
    $script:ibBaseField=[Management.Automation.PSObject].GetField('_immediateBaseObject',$script:ibReflectionFlags)
    $script:ibHidden=[Management.Automation.PSMemberInfo].GetField('<IsHidden>k__BackingField',[Reflection.BindingFlags]'NonPublic,Instance')
}
function IB-Decorate([object[]]$Source,[bool]$Subclass,[bool]$Hidden){
    # Actual wrapper carried in object[]; only the PEER is materialized.
    $wrapper=$Source[0]
    $psType=[Management.Automation.PSObject].Assembly.GetType('System.Management.Automation.ReservedNameMembers',$true)
    $factory=$psType.GetMethod('GeneratePSObjectMemberSet',[Reflection.BindingFlags]'NonPublic,Static')
    $a=[object[]]::new(1);$a[0]=$wrapper;$null=$factory.Invoke($null,$a)
    $factory=$psType.GetMethod('GeneratePSExtendedMemberSet',[Reflection.BindingFlags]'NonPublic,Static');$null=$factory.Invoke($null,$a)
    $read=[Reflection.FieldInfo].GetMethod('GetValue',[type[]]@([object]))
    $table=$read.Invoke($script:ibCacheField,$a)
    if($Subclass){$tag=[P2BDescriptorFixture.CounterNote]::new('InboxProbe',[int]42)}else{$tag=[Management.Automation.PSNoteProperty]::new('InboxProbe',[int]42)}
    if($Hidden){$script:ibHidden.SetValue($tag,$true)}
    $add=$script:ibRawField.FieldType.GetMethod('Add',[type[]]@([Management.Automation.PSMemberInfo]))
    $argsLocal=[object[]]::new(1);$argsLocal[0]=$tag;$null=$add.Invoke($table,$argsLocal)
    if([P2BDescriptorFixture.CounterNote]::Total -ne 0){throw 'Descriptor construction callback'}
    return [pscustomobject]@{Table=$table;Tag=$tag}
}
function IB-Wrapper($base){$a=[object[]]::new(1);$a[0]=$base;$out=[object[]]::new(1);$out[0]=[Management.Automation.PSObject].GetConstructor([type[]]@([object])).Invoke($a);return ,$out}
function IB-Cold($base,[bool]$Sub,[bool]$Hidden){
    $peer=IB-Wrapper $base;$proof=IB-Decorate $peer $Sub $Hidden;$cold=IB-Wrapper $base
    $a=[object[]]::new(1);$a[0]=$cold[0]
    $read=[Reflection.FieldInfo].GetMethod('GetValue',[type[]]@([object]))
    if($null -ne $read.Invoke($script:ibCacheField,$a)){throw 'Cold precondition'}
    $script:ibTrace.Decorations+=,[pscustomobject]@{Subclass=$Sub;Hidden=$Hidden;Cold=$true;TagClass=([object].GetMethod('GetType').Invoke($proof.Tag,$null)).FullName;Retained=[bool]$false;Counters=0}
    $coldProof=[object[]]::new(3);$coldProof[0]=$cold[0];$coldProof[1]=$proof;$coldProof[2]=$script:ibTrace.Decorations[-1]
    return ,$coldProof
}
function IB-ProveTag([object[]]$Proof){
    $found=$false
    foreach($entry in $Proof[1].Table){if([object]::ReferenceEquals($entry,$Proof[1].Tag)){$found=$true}}
    $Proof[2].Retained=$found;$Proof[2].Counters=[P2BDescriptorFixture.CounterNote]::Total
    IB-Need ($found -and $Proof[2].Counters -eq 0)
}
