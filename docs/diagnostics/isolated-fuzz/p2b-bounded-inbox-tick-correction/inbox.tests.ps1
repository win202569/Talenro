param([ValidateSet('all','capacity','sequence')][string]$Group='all',[string]$TracePath)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
[Console]::Out.WriteLine('CONTEXT: '+$identity.Name+' admin='+$admin)
[Console]::Out.WriteLine('HOST: PS'+$PSVersionTable.PSVersion.ToString()+' Core x64')
. (Join-Path $PSScriptRoot 'test-helpers.ps1')
$script:ibTrace=[pscustomobject]@{Schema=1;Mode='p2b-inbox-tests';Group=$Group;Groups=@();CompileCount=0;Observable=$false;Counters=0;Synthetic=@();Decorations=@();Error=$null;Passed=$false;RuntimeApproved=$false}
function IB-Case([string]$Name,[scriptblock]$Body){
    $wanted=$Group -ceq 'all' -or $Name -ceq 'ApiRequired' -or ($Group -ceq 'capacity' -and $Name -ceq 'EndpointCapacity') -or ($Group -ceq 'sequence' -and $Name -ceq 'SequenceExact')
    if(-not $wanted){return};$script:ibGroup=$Name;& $Body
    $script:ibTrace.Groups+=,$Name;[Console]::Out.WriteLine('PASS: P2BInbox'+$Name)
}
try{
    $modulePath=Join-Path $PSScriptRoot 'inbox.psm1'
    if(Test-Path -LiteralPath $modulePath){$script:ibModule=Import-Module $modulePath -PassThru -DisableNameChecking}
    IB-Case ApiRequired {
        foreach($name in $script:ibNames){IB-Need ($null -ne (Get-Command -Name $name -ErrorAction SilentlyContinue))}
        $exports=[string[]]@($script:ibModule.ExportedFunctions.Keys|Sort-Object)
        IB-Need (IB-Eq $exports ([string[]]@($script:ibNames|Sort-Object)))
    }
    $script:ibLiterals=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'fixtures.json') -Raw|ConvertFrom-Json
    IB-Case BindingShape {
        foreach($which in @('null','hashtable','extra','missing','case','type','digest','session','time')){
            $b=IB-Binding
            switch($which){
                null {$b=$null} hashtable {$b=@{Schema=1}} extra {$b|Add-Member NoteProperty Extra 1}
                missing {$b.PSObject.Properties.Remove('Schema')} case {$b.PSObject.Properties.Remove('Schema');$b|Add-Member NoteProperty schema ([int]1)}
                type {$b.Schema=[long]1} digest {$b.BindingDigest=[byte[]]::new(31)} session {$b.SessionId=[byte[]]::new(17)} time {$b.DeadlineTick=[long]120001}
            }
            $null=IB-Result (New-P2BInbox $b) 'invalid-input'
        }
    }
    IB-Case BindingIsolation {
        $b=IB-Binding;$r=IB-Result (New-P2BInbox $b) created;$t=$r.Inbox
        $b.BindingDigest[0]=99;$b.SessionId[0]=99;$b.Frequency=[long]1001
        $s=IB-Snap $t;$s.Binding.BindingDigest[1]=98;$s.Binding.Schema=[int]7
        IB-Check $t (IB-Initial)
    }
    IB-Case TokenIdentity {
        $t=IB-New
        foreach($fake in @([object]::new(),[pscustomobject]@{},[Management.Automation.PSSerializer]::Deserialize([Management.Automation.PSSerializer]::Serialize($t)))){
            $null=IB-Result (Get-P2BInboxSnapshot $fake) invalid-input
        }
        $t|Add-Member NoteProperty FakeState ([int]7);IB-Check $t (IB-Initial)
    }
    IB-Case EmptyInbox {
        $t=IB-New;IB-Check $t (IB-Initial)
        foreach($e in 1..4){$null=IB-Result (Receive-P2BInboxFrame $t ([byte]$e)) empty ([byte]$e)}
        IB-Check $t (IB-Initial)
    }
    IB-Case WholeFrame {
        foreach($e in 1..4){
            $t=IB-New;$f=IB-Frame $e;$null=IB-Add $t $f ([byte]$e)
            $ep=IB-Initial;$ep[$e-1]=IB-Endpoint ([byte]$e) 2 1 ([object[]]@(,$f));IB-Check $t $ep
            $r=IB-Result (Receive-P2BInboxFrame $t ([byte]$e)) dequeued ([byte]$e);IB-Need (IB-Eq $r.Frame $f)
            $ep[$e-1]=IB-Endpoint ([byte]$e) 2 1;IB-Check $t $ep
        }
    }
    IB-Case SplitEdges {
        foreach($n in @(1,255)){
            $t=IB-New;$f=IB-Frame;$a=[byte[]]::new($n);$b=[byte[]]::new(256-$n)
            [Array]::Copy($f,0,$a,0,$n);[Array]::Copy($f,$n,$b,0,$b.Length)
            $null=IB-Add $t $a 1 partial;$ep=IB-Initial;$ep[0]=IB-Endpoint 1 1 0 -Partial $a;IB-Check $t $ep
            $null=IB-Add $t $b;$ep[0]=IB-Endpoint 1 2 1 ([object[]]@(,$f));IB-Check $t $ep
        }
    }
    IB-Case SingleByteChunks {
        $t=IB-New;$f=IB-Frame;$total=0
        for($i=0;$i -lt 256;$i++){
            $part=[byte[]]::new(1);$part[0]=$f[$i];$reason=if($i -eq 255){'queued'}else{'partial'}
            $r=IB-Add $t $part 1 $reason;$total+=$r.AcceptedBytes
        }
        IB-Need ($total -eq 256);$ep=IB-Initial;$ep[0]=IB-Endpoint 1 2 1 ([object[]]@(,$f));IB-Check $t $ep
    }
    IB-Case ChunkBounds {
        foreach($n in @(0,257,512)){
            $t=IB-New;$null=IB-Result (Add-P2BInboxFragment $t ([byte]1) ([byte[]]::new($n))) invalid-input ([byte]1)
            IB-Check $t (IB-Initial) (IB-Failure invalid-input ([byte]1) input)
        }
        $t=IB-New;$null=IB-Add $t ([byte[]]@(80)) 1 partial
        $null=IB-Result (Add-P2BInboxFragment $t ([byte]1) (IB-Frame)) invalid-input ([byte]1)
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 -Partial ([byte[]]@(80));IB-Check $t $ep (IB-Failure invalid-input ([byte]1) input)
    }
    IB-Case LiteralDirections {
        foreach($e in 1..4){
            $t=IB-New;$wrong=($e%4)+1;$f=IB-Frame $wrong;$null=IB-Add $t $f ([byte]$e) binding-mismatch
            $ep=IB-Initial;$ep[$e-1]=IB-Endpoint ([byte]$e) -Partial $f;IB-Check $t $ep (IB-Failure binding-mismatch ([byte]$e) binding ([uint64]1))
        }
        foreach($bad in @([int]1,[byte]0,[byte]5,'1',$null)){
            $t=IB-New;$null=IB-Result (Receive-P2BInboxFrame $t $bad) invalid-input;IB-Check $t (IB-Initial) (IB-Failure invalid-input $null input)
        }
    }
    IB-Case BindingMismatch {
        foreach($which in @('digest','session','origin','frequency','deadline')){
            $t=IB-New;$f=IB-Frame
            switch($which){digest{$f[16]=99}session{$f[48]=99}origin{IB-Put $f 72 8 10;IB-Put $f 88 8 120010;IB-Put $f 96 8 11}frequency{IB-Put $f 80 8 1001;IB-Put $f 88 8 120120}deadline{IB-Put $f 72 8 1;IB-Put $f 88 8 120001}}
            $null=IB-Add $t $f 1 binding-mismatch
            $ep=IB-Initial;$ep[0]=IB-Endpoint 1 -Partial $f;IB-Check $t $ep (IB-Failure binding-mismatch ([byte]1) binding ([uint64]1))
        }
    }
    IB-Case SequenceExact {
        $t=IB-New;$f=IB-Frame 1 2;$r=Add-P2BInboxFragment $t ([byte]1) $f
        if(-not $r.Accepted -and $r.Reason -cne 'sequence-mismatch'){throw 'Unexpected sequence result'}
        IB-Need (-not $r.Accepted)
        $null=IB-Result $r sequence-mismatch ([byte]1) 256
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 -Partial $f;IB-Check $t $ep (IB-Failure sequence-mismatch ([byte]1) sequence ([uint64]2))
        $t=IB-New;$null=IB-Add $t (IB-Frame 1 0) 1 invalid-frame
        $t=IB-New;$null=IB-Add $t (IB-Frame)
    }
    IB-Case SequenceAfterTake {
        foreach($take in @($false,$true)){
            $t=IB-New;$f=IB-Frame;$null=IB-Add $t $f
            if($take){$null=IB-Result (Receive-P2BInboxFrame $t ([byte]1)) dequeued ([byte]1)}
            $null=IB-Add $t $f 1 sequence-mismatch
            $q=[object[]]::new(0);if(-not $take){$q=[object[]]@(,$f)}
            $ep=IB-Initial;$ep[0]=IB-Endpoint 1 2 1 $q -Partial $f;IB-Check $t $ep (IB-Failure sequence-mismatch ([byte]1) sequence ([uint64]1))
        }
    }
    IB-Case SequenceHigh {
        $high=[uint64]::Parse('9223372036854775808')
        $t=IB-New;IB-PrivateSequence $t $high;$f=IB-Frame 1 $high;$null=IB-Add $t $f
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 ([uint64]::Parse('9223372036854775809')) $high ([object[]]@(,$f));IB-Check $t $ep
        foreach($seq in @([uint64]::Parse('9223372036854775807'),[uint64]::Parse('9223372036854775809'))){$t=IB-New;IB-PrivateSequence $t $high;$null=IB-Add $t (IB-Frame 1 $seq) 1 sequence-mismatch}
    }
    IB-Case SequenceExhaustion {
        $near=[uint64]::Parse('18446744073709551614');$max=[uint64]::MaxValue
        $t=IB-New;IB-PrivateSequence $t $near;$f=IB-Frame 1 $near;$null=IB-Add $t $f
        $bad=IB-Frame 1 $max;$null=IB-Add $t $bad 1 sequence-exhausted
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 $max $near ([object[]]@(,$f)) -Partial $bad
        IB-Check $t $ep (IB-Failure sequence-exhausted ([byte]1) sequence $max)
    }
    IB-Case EndpointCapacity {
        $t=IB-New;$q=[object[]]::new(4);for($i=1;$i -le 4;$i++){$q[$i-1]=IB-Frame 1 ([uint64]$i);$null=IB-Add $t $q[$i-1]}
        $f=IB-Frame 1 5;$r=Add-P2BInboxFragment $t ([byte]1) $f
        if(-not $r.Accepted -and $r.Reason -cne 'capacity'){throw 'Unexpected capacity result'}
        IB-Need (-not $r.Accepted);$null=IB-Result $r capacity ([byte]1) 256
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 5 4 $q -Partial $f;IB-Check $t $ep (IB-Failure capacity ([byte]1) capacity ([uint64]5))
    }
    IB-Case GlobalCapacity {
        $t=IB-New;$ep=IB-Initial
        foreach($e in 1..4){
            $q=[object[]]::new(4);for($i=1;$i -le 4;$i++){$q[$i-1]=IB-Frame $e ([uint64]$i);$null=IB-Add $t $q[$i-1] ([byte]$e)}
            $hb=IB-Frame $e 5 8 4;$null=IB-Add $t $hb ([byte]$e) heartbeat-updated;$part=[byte[]]@(80);$null=IB-Add $t $part ([byte]$e) partial
            $ep[$e-1]=IB-Endpoint ([byte]$e) 6 4 $q $hb $part
        };IB-Check $t $ep
    }
    IB-Case HeartbeatMixed {
        $t=IB-New;$f1=IB-Frame 1 1;$f4=IB-Frame 1 4;$hb=IB-Frame 1 3 8 1
        $null=IB-Add $t $f1;$null=IB-Add $t (IB-Frame 1 2 8 1) 1 heartbeat-updated;$null=IB-Add $t $hb 1 heartbeat-updated;$null=IB-Add $t $f4
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 5 4 ([object[]]@($f1,$f4)) $hb;IB-Check $t $ep
        $r=IB-Result (Receive-P2BInboxFrame $t ([byte]1)) dequeued ([byte]1);IB-Need (IB-Eq $r.Frame $f1)
    }
    IB-Case HeartbeatReference {
        foreach($variant in @('reference','repeat','gap')){
            $t=IB-New;$old=IB-Frame 1 1 8 0;$null=IB-Add $t $old 1 heartbeat-updated
            $seq=if($variant -ceq 'repeat'){1}elseif($variant -ceq 'gap'){3}else{2}
            $last=if($variant -ceq 'reference'){[uint64]1}else{[uint64]0}
            $why=if($variant -ceq 'reference'){'reference-mismatch'}else{'sequence-mismatch'}
            $f=IB-Frame 1 ([uint64]$seq) 8 $last;$null=IB-Add $t $f 1 $why
            $ep=IB-Initial;$ep[0]=IB-Endpoint 1 2 0 -HB $old -Partial $f
            $stage=if($variant -ceq 'reference'){'reference'}else{'sequence'}
            IB-Check $t $ep (IB-Failure $why ([byte]1) $stage ([uint64]$seq))
        }
    }
    IB-Case HeartbeatFullQueue {
        $t=IB-New;$q=[object[]]::new(4);for($i=1;$i -le 4;$i++){$q[$i-1]=IB-Frame 1 ([uint64]$i);$null=IB-Add $t $q[$i-1]}
        $hb=IB-Frame 1 5 8 4;$null=IB-Add $t $hb 1 heartbeat-updated;$f=IB-Frame 1 6;$null=IB-Add $t $f 1 capacity
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 6 4 $q $hb $f;IB-Check $t $ep (IB-Failure capacity ([byte]1) capacity ([uint64]6))
    }
    IB-Case EofEmpty {
        foreach($prior in @($false,$true)){
            $t=IB-New;$ep=IB-Initial
            if($prior){$f=IB-Frame;$null=IB-Add $t $f;$ep[0]=IB-Endpoint 1 2 1 ([object[]]@(,$f))}
            $null=IB-Result (Close-P2BInboxInput $t ([byte]1)) peer-closed ([byte]1);IB-Check $t $ep (IB-Failure peer-closed ([byte]1) eof)
        }
    }
    IB-Case EofPartial {
        foreach($n in @(1,255)){
            $t=IB-New;$f=IB-Frame;$part=[byte[]]::new($n);[Array]::Copy($f,$part,$n);$null=IB-Add $t $part 1 partial
            $null=IB-Result (Close-P2BInboxInput $t ([byte]1)) truncated-frame ([byte]1)
            $ep=IB-Initial;$ep[0]=IB-Endpoint 1 -Partial $part;IB-Check $t $ep (IB-Failure truncated-frame ([byte]1) eof)
        }
    }
    IB-Case InvalidFrame {
        foreach($offset in @(0,9,128,224)){
            $t=IB-New;$f=IB-Frame;$f[$offset]=255;$null=IB-Add $t $f 1 invalid-frame
            $ep=IB-Initial;$ep[0]=IB-Endpoint 1 -Partial $f;IB-Check $t $ep (IB-Failure invalid-frame ([byte]1) frame)
        }
    }
    IB-Case FirstFailure {
        $t=IB-New;for($i=1;$i -le 4;$i++){$null=IB-Add $t (IB-Frame 1 ([uint64]$i))}
        $null=IB-Add $t (IB-Frame 1 5) 1 capacity;$before=IB-Snap $t
        $null=IB-Result (Add-P2BInboxFragment $t ([byte]2) ([byte[]]::new(256))) stopped
        $null=IB-Result (Close-P2BInboxInput $t ([byte]4)) stopped;IB-Need (IB-Eq (IB-Snap $t) $before)
        $t=IB-New;$null=IB-Result (Add-P2BInboxFragment $t ([int]1) ([byte[]]::new(0))) invalid-input
        IB-Check $t (IB-Initial) (IB-Failure invalid-input $null input)
    }
    IB-Case StoppedDelivery {
        $t=IB-New;$null=IB-Add $t (IB-Frame);$null=IB-Result (Close-P2BInboxInput $t ([byte]2)) peer-closed ([byte]2)
        $before=IB-Snap $t;$script:poisonReads=0;$poison=[pscustomobject]@{}
        $poison|Add-Member ScriptProperty Value {$script:poisonReads++;throw 'Poison'}
        foreach($e in 1..4){$null=IB-Result (Receive-P2BInboxFrame $t ([byte]$e)) stopped;$null=IB-Result (Close-P2BInboxInput $t $poison) stopped;$null=IB-Result (Add-P2BInboxFragment $t $poison $poison) stopped}
        IB-Need ($script:poisonReads -eq 0 -and (IB-Eq (IB-Snap $t) $before))
    }
    IB-Case InputPassivity {
        IB-InitDescriptors
        foreach($sub in @($false,$true)){foreach($hidden in @($false,$true)){
            foreach($what in @('bytes','endpoint')){
                $t=IB-New
                $base=if($what -ceq 'bytes'){IB-Frame}else{$a=[byte[]]@(1);$a.GetValue(0)}
                $proof=IB-Cold $base $sub $hidden
                if($what -ceq 'bytes'){$r=Add-P2BInboxFragment $t ([byte]1) $proof[0];$null=IB-Result $r invalid-input ([byte]1)}
                else{$r=Add-P2BInboxFragment $t $proof[0] (IB-Frame);$null=IB-Result $r invalid-input}
                IB-ProveTag $proof
            }
        }}
        foreach($which in @('subclass','hidden','executable')){
            $b=IB-Binding;$b.PSObject.Properties.Remove('Schema')
            if($which -ceq 'executable'){$b|Add-Member ScriptProperty Schema {$script:poisonReads++;throw 'Poison'}}
            else{
                $extended=$b.psextended
                $table=$script:ibRawField.GetValue($extended.PSObject.BaseObject)
                $tag=if($which -ceq 'subclass'){[P2BDescriptorFixture.CounterNote]::new('Schema',[int]1)}else{[Management.Automation.PSNoteProperty]::new('Schema',[int]1)}
                if($which -ceq 'hidden'){$script:ibHidden.SetValue($tag,$true)}
                $add=$script:ibRawField.FieldType.GetMethod('Add',[type[]]@([Management.Automation.PSMemberInfo]));$null=$add.Invoke($table,[object[]]@($tag))
            }
            $null=IB-Result (New-P2BInbox $b) invalid-input;IB-Need ([P2BDescriptorFixture.CounterNote]::Total -eq 0 -and $script:poisonReads -eq 0)
        }
        foreach($name in @('Length','Rank')){
            $t=IB-New;$f=IB-Frame;$f.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($name,[int]1))
            $null=IB-Result (Add-P2BInboxFragment $t ([byte]1) $f) invalid-input ([byte]1)
        }
    }
    IB-Case InputOwnership {
        $t=IB-New;$f=IB-Frame;$part=[byte[]]::new(128);[Array]::Copy($f,$part,128);$null=IB-Add $t $part 1 partial
        $part[0]=0;$part.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Probe',[int]1))
        $rest=[byte[]]::new(128);[Array]::Copy($f,128,$rest,0,128);$null=IB-Add $t $rest
        $rest[0]=99;$r=IB-Result (Receive-P2BInboxFrame $t ([byte]1)) dequeued ([byte]1);IB-Need (IB-Eq $r.Frame $f)
    }
    IB-Case OutputOwnership {
        $t=IB-New;$s=IB-Snap $t;$q=$s.Endpoints[0].Queue;$p=$s.Endpoints[0].PartialBytes
        $q.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Length',[int]99));$p.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Rank',[int]99))
        IB-Need ($q.Length -eq 99 -and $p.Rank -eq 99);IB-Check $t (IB-Initial)
        $f=IB-Frame;$null=IB-Add $t $f;$a=IB-Snap $t;$a.Endpoints[0].Queue[0][0]=0;$a.Binding.SessionId[0]=0
        $a.Endpoints[0].Count.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Probe',[int]1))
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 2 1 ([object[]]@(,$f));IB-Check $t $ep
        $r=IB-Result (Receive-P2BInboxFrame $t ([byte]1)) dequeued ([byte]1);$r.Frame[0]=0
        $ep[0]=IB-Endpoint 1 2 1;IB-Check $t $ep
    }
    IB-Case SnapshotBounds {
        $t=IB-New
        for($i=1;$i -le 12;$i++){$null=IB-Add $t (IB-Frame 1 ([uint64]$i));$null=IB-Result (Receive-P2BInboxFrame $t ([byte]1)) dequeued ([byte]1)}
        for($i=13;$i -le 20;$i++){$null=IB-Add $t (IB-Frame 1 ([uint64]$i) 8 12) 1 heartbeat-updated}
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 21 12 -HB (IB-Frame 1 20 8 12);IB-Check $t $ep
        $s=IB-Snap $t;IB-Need ($s.Endpoints.Count -eq 4 -and @($s.Endpoints[0].PSObject.Properties).Count -eq 8)
    }
    IB-Case MultiInstance {
        $a=IB-New;$b=IB-New;IB-Need (-not [object]::ReferenceEquals($a,$b))
        $null=IB-Add $a (IB-Frame);$null=IB-Result (Close-P2BInboxInput $a ([byte]1)) peer-closed ([byte]1)
        $s=IB-Snap $a;$s.Endpoints[1].Queue.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Length',[int]99))
        IB-Check $b (IB-Initial);$null=IB-Add $b (IB-Frame)
    }
    IB-Case AuthorityBoundary {
        $t=IB-New;$arm=IB-Frame 2 1 4;$null=IB-Add $t $arm 2
        $t=IB-New;$ack=IB-Frame 1 1 10;$null=IB-Add $t $ack
        $r=IB-Result (Receive-P2BInboxFrame $t ([byte]1)) dequeued ([byte]1);IB-Need (IB-Eq $r.Frame $ack)
    }
    IB-Case Complete {
        IB-Need ($script:ibTrace.Groups.Count -eq 31 -and $script:ibTrace.CompileCount -eq 1 -and $script:ibTrace.Observable)
        $script:ibTrace.Counters=[P2BDescriptorFixture.CounterNote]::Total;IB-Need ($script:ibTrace.Counters -eq 0)
    }
    $script:ibTrace.Passed=$true
    if($Group -ceq 'all'){[Console]::Out.WriteLine('LIMIT: declarations/inbox only; no runtime authority')}
}catch{
    $script:ibTrace.Error=[pscustomobject]@{Message=$_.Exception.Message;Type=$_.Exception.GetType().FullName;Position=$_.InvocationInfo.PositionMessage;Stack=$_.ScriptStackTrace}
    [Console]::Out.WriteLine($_.Exception.Message)
}finally{
    if(-not [string]::IsNullOrEmpty($TracePath)){[IO.File]::WriteAllText($TracePath,($script:ibTrace|ConvertTo-Json -Depth 24),[Text.UTF8Encoding]::new($false))}
}
if(-not $script:ibTrace.Passed){exit 1}
