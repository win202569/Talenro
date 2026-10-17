$ErrorActionPreference='Stop'
function Check-Contract([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
$helper=Join-Path $PSScriptRoot 'capture-adapter-contract.ps1';if(Test-Path -LiteralPath $helper){. $helper}
Check-Contract ($null -ne (Get-Command New-SyntheticCaptureRecord -CommandType Function -ErrorAction SilentlyContinue) -and $null -ne (Get-Command Test-SyntheticCaptureRecord -CommandType Function -ErrorAction SilentlyContinue)) 'SyntheticCaptureContractRequired'
Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
function Contract-Packet{
    [pscustomobject]@{Schema=1;SourceMode='synthetic';SessionLabel='s';ExecutableDeclaration='x';ArgumentBytes=[byte[]]@();EncodingName='utf-8';OriginDeclaration='o';StartedUTC='2026-10-17T00:00:00.0000000Z';EndedUTC='2026-10-17T00:00:00.0000000Z';Events=[object[]]@()}
}
function Contract-Chunk($Channel,$Bytes,$Time=0){[pscustomobject]@{Kind='chunk';ElapsedMs=$Time;Channel=$Channel;Bytes=$Bytes}}
function Contract-EOF($Channel,$Time=0){[pscustomobject]@{Kind='eof';ElapsedMs=$Time;Channel=$Channel}}
function Contract-Exit($Code=0,$Time=0){[pscustomobject]@{Kind='exit';ElapsedMs=$Time;ExitCode=$Code}}
function Contract-Tick($Time=0){[pscustomobject]@{Kind='tick';ElapsedMs=$Time}}
function Contract-Verify($Record){
    $v=Test-SyntheticCaptureRecord $Record
    if(@($v.PSObject.Properties).Count -ne 5 -or $v.OriginVerified -or $v.Acceptance -or $v.RuntimeApproved -or $v.Verified -isnot [bool]){throw 'Bad verifier shape/flags'}
    return $v
}
function Contract-Good($Packet){
    $r=New-SyntheticCaptureRecord $Packet
    $names=@('Schema','Policy','SourceMode','SessionLabel','ExecutableDeclaration','Arguments','EncodingName','OriginDeclaration','StartedUTC','EndedUTC','UtcOrderConsistent','Events','Lifecycle','StopIntent','CleanupStatus','LiveActionsPerformed','BindingSHA256','OriginVerified','Acceptance','RuntimeApproved')
    if(@($r.PSObject.Properties).Count -ne 20 -or @($r.PSObject.Properties|Where-Object{$_.Name -cnotin $names}).Count -ne 0 -or $r.Schema -ne 1 -or $r.Policy -cne 'synthetic-capture-contract-v1' -or $r.SourceMode -cne 'synthetic'){throw 'Expected exact synthetic record'}
    foreach($flag in @('LiveActionsPerformed','OriginVerified','Acceptance','RuntimeApproved')){if($r.$flag -isnot [bool] -or $r.$flag){throw 'Authority/action claim'}}
    return $r
}
function Contract-Refusal($Packet,[string]$Name){
    $r=New-SyntheticCaptureRecord $Packet
    Check-Contract (@($r.PSObject.Properties).Count -eq 7 -and $r.Created -is [bool] -and -not $r.Created -and $r.SourceMode -ceq 'synthetic' -and $r.Reason -ceq 'invalid-packet' -and -not $r.OriginVerified -and -not $r.Acceptance -and -not $r.RuntimeApproved -and $null -eq $r.PSObject.Properties['BindingSHA256']) $Name
}
# Break caught:linguistically equivalent protocol text is not byte-exact text.
$script:ordinalFailures=0
function Check-ContractOrdinal([bool]$Value,[string]$Name){
    if($Value){Write-Output ('PASS: '+$Name)}else{$script:ordinalFailures++;Write-Output ('FAIL: '+$Name)}
}
function Contract-OrdinalRefusal($Packet,[string]$Name){
    $value=New-SyntheticCaptureRecord $Packet
    Check-ContractOrdinal (@($value.PSObject.Properties).Count -eq 7 -and $value.Created -is [bool] -and -not $value.Created -and [string]::Equals($value.Reason,'invalid-packet',[StringComparison]::Ordinal) -and $null -eq $value.PSObject.Properties['BindingSHA256'] -and -not $value.OriginVerified -and -not $value.Acceptance -and -not $value.RuntimeApproved) $Name
}
$nul=[string][char]0;$softHyphen=[string][char]0x00AD
$p=Contract-Packet;$p.EncodingName='utf'+$nul+'-8'
Contract-OrdinalRefusal $p 'OrdinalEncodingTokenRejectsEmbeddedNull'
$p=Contract-Packet;$p.EncodingName='utf'+$softHyphen+'-8'
Contract-OrdinalRefusal $p 'OrdinalEncodingTokenRejectsIgnorableCharacter'
$p=Contract-Packet;$p.EncodingName='UTF-8';Contract-OrdinalRefusal $p 'OrdinalEncodingTokenRejectsCaseChange'
$p=Contract-Packet;$p.SourceMode='synthe'+$nul+'tic';Contract-OrdinalRefusal $p 'OrdinalSourceModeRejectsEmbeddedNull'
$p=Contract-Packet;$p.Events=[object[]]@([pscustomobject]@{Kind=('ti'+$nul+'ck');ElapsedMs=0})
Contract-OrdinalRefusal $p 'OrdinalRawEventKindRejectsEmbeddedNull'
$p=Contract-Packet;$p.Events=[object[]]@([pscustomobject]@{Kind='Tick';ElapsedMs=0})
Contract-OrdinalRefusal $p 'OrdinalRawEventKindRejectsCaseChange'
$p=Contract-Packet;$p.Events=[object[]]@((Contract-Chunk ('std'+$nul+'out') ([byte[]]@(97))))
Contract-OrdinalRefusal $p 'OrdinalRawChannelRejectsEmbeddedNull'
$p=Contract-Packet;$p.PSObject.Properties.Remove('Schema');Add-Member -InputObject $p -NotePropertyName ('Schema'+$nul) -NotePropertyValue 1
Contract-OrdinalRefusal $p 'OrdinalPacketFieldRejectsEmbeddedNull'
$p=Contract-Packet;$p.Events=[object[]]@([pscustomobject]@{Kind='tick';('ElapsedM'+$nul+'s')=0})
Contract-OrdinalRefusal $p 'OrdinalRawEventFieldRejectsEmbeddedNull'
$p=Contract-Packet;$p.Events=[object[]]@((Contract-Tick));$r=Contract-Good $p;$r.Events[0].Kind='ti'+$nul+'ck'
Check-ContractOrdinal (-not (Contract-Verify $r).Verified) 'OrdinalStoredEventKindRejectsEmbeddedNull'
$p=Contract-Packet;$p.Events=[object[]]@((Contract-Chunk 'stdout' ([byte[]]@(97))));$r=Contract-Good $p;$r.Events[0].Channel='std'+$nul+'out'
Check-ContractOrdinal (-not (Contract-Verify $r).Verified) 'OrdinalStoredChannelRejectsEmbeddedNull'
$r=Contract-Good (Contract-Packet);$r.PSObject.Properties.Remove('Policy');Add-Member -InputObject $r -NotePropertyName ('Policy'+$nul) -NotePropertyValue 'synthetic-capture-contract-v1'
Check-ContractOrdinal (-not (Contract-Verify $r).Verified) 'OrdinalRecordFieldRejectsEmbeddedNull'
$pre=([string][char]0x00E9)+'e'+[char]0x0301;$post='e'+[char]0x0301+[char]0x00E9
$p=Contract-Packet;$p.OriginDeclaration=$pre;$p.Events=[object[]]@((Contract-Exit),(Contract-EOF 'stdout'),(Contract-EOF 'stderr'))
$r=Contract-Good $p;$r.Lifecycle.Package.OriginDeclaration=$post
Check-ContractOrdinal (-not (Contract-Verify $r).Verified) 'OrdinalDerivedTextRejectsEqualLengthCanonicalEquivalent'
$p1=Contract-Packet;$p1.OriginDeclaration=$pre;$p2=Contract-Packet;$p2.OriginDeclaration=$post
$r1=Contract-Good $p1;$r2=Contract-Good $p2
Check-ContractOrdinal ((Contract-Verify $r1).Verified -and (Contract-Verify $r2).Verified -and [string]::Equals($r1.OriginDeclaration,$pre,[StringComparison]::Ordinal) -and [string]::Equals($r2.OriginDeclaration,$post,[StringComparison]::Ordinal) -and -not [string]::Equals($r1.BindingSHA256,$r2.BindingSHA256,[StringComparison]::Ordinal)) 'OrdinalUnicodeDeclarationsRemainDistinctWithoutNormalization'
$script:ordinalCallbacks=0;$invalidKind=[pscustomobject]@{}
Add-Member -InputObject $invalidKind -MemberType ScriptMethod -Name ToString -Force -Value {$script:ordinalCallbacks++;'tick'}
$p=Contract-Packet;$p.Events=[object[]]@([pscustomobject]@{Kind=$invalidKind;ElapsedMs=0})
Contract-OrdinalRefusal $p 'OrdinalInvalidTokenNeverCoerced'
Check-ContractOrdinal ($script:ordinalCallbacks -eq 0) 'OrdinalTokenToStringCallbackNotInvoked'
Check-Contract ($script:ordinalFailures -eq 0) 'OrdinalContractMustRejectNonExactText'
# Break caught: ETS length/rank shadows must never clip bytes or bypass bounds.
# Fixtures call the public APIs;literal refusal is independent of implementation.
function Contract-Shadow($Value,[string]$Name,$Shadow){
    # Retain the returned ETS wrapper;immutable strings do not keep an in-place addition.
    $Value=Add-Member -InputObject $Value -MemberType NoteProperty -Name $Name -Value $Shadow -Force -PassThru
    if($Value.PSObject.Properties[$Name].MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or $Value.$Name -ne $Shadow){throw 'Shadow fixture not attached'}
    return ,$Value
}
$script:shadowFailures=0
function Check-ContractShadow([bool]$Value,[string]$Name){
    if($Value){Write-Output ('PASS: '+$Name)}else{$script:shadowFailures++;Write-Output ('FAIL: '+$Name)}
}
function Contract-ShadowRefusal($Packet,[string]$Name){
    $value=New-SyntheticCaptureRecord $Packet
    Check-ContractShadow (@($value.PSObject.Properties).Count -eq 7 -and $value.Created -is [bool] -and -not $value.Created -and $value.Reason -ceq 'invalid-packet' -and $null -eq $value.PSObject.Properties['BindingSHA256'] -and -not $value.OriginVerified -and -not $value.Acceptance -and -not $value.RuntimeApproved) $Name
}
$p=Contract-Packet;$p.ArgumentBytes=Contract-Shadow ([byte[]]@(65,0)) 'Length' 0
Contract-ShadowRefusal $p 'ShadowArgumentBytesNotClippedToEmpty'
$p=Contract-Packet;$p.ArgumentBytes=Contract-Shadow (New-Object byte[] 65538) 'Length' 0
Contract-ShadowRefusal $p 'ShadowOversizedArgumentsRefused'
$p=Contract-Packet;$p.Events=[object[]]@((Contract-Chunk 'stdout' (Contract-Shadow ([byte[]]@(97,98,99)) 'Length' 0)))
Contract-ShadowRefusal $p 'ShadowChunkBytesNotClippedToEmpty'
$p=Contract-Packet;$p.Events=[object[]]@((Contract-Chunk 'stdout' (Contract-Shadow (New-Object byte[] 65537) 'Length' 0)))
Contract-ShadowRefusal $p 'ShadowOversizedChunkRefused'
$p=Contract-Packet;$many=New-Object object[] 4097;for($i=0;$i -lt 4097;$i++){$many[$i]=Contract-Tick}
$p.Events=Contract-Shadow $many 'Length' 0
Contract-ShadowRefusal $p 'ShadowEventCountCannotBypassPreflight'
$p=Contract-Packet;$p.Events=Contract-Shadow ([object[]]@((Contract-Tick))) 'Rank' 1
Contract-ShadowRefusal $p 'ShadowRankRefusedEvenWhenValueMatches'
$p=Contract-Packet;$p.ExecutableDeclaration=Contract-Shadow ('x'*1025) 'Length' 1
Contract-ShadowRefusal $p 'ShadowExecutableCharacterBoundRefused'
$p=Contract-Packet;$p.OriginDeclaration=Contract-Shadow ('o'*1025) 'Length' 1
Contract-ShadowRefusal $p 'ShadowOriginCharacterBoundRefused'
$p=Contract-Packet;$p.Events=[object[]]@((Contract-Tick 120000),(Contract-Chunk 'stdout' (Contract-Shadow ([byte[]]@(97)) 'Length' 0) 120000))
Contract-ShadowRefusal $p 'ShadowBytesAfterTimeoutStillRefused'
$r=Contract-Good (Contract-Packet);$r.Events=Contract-Shadow $r.Events 'Length' 0
Check-ContractShadow (-not (Contract-Verify $r).Verified) 'ShadowStoredEventArrayRefused'
$independent=Contract-Good (Contract-Packet)
Check-ContractShadow ((Contract-Verify $independent).Verified -and $independent.Lifecycle.Status -ceq 'incomplete') 'ReturnedEmptyEventsDoNotPoisonNextRecord'
$r.Events.PSObject.Properties.Remove('Length')
$r=Contract-Good (Contract-Packet);$r.Arguments.Base64=Contract-Shadow ([string]::Empty) 'Length' 0
Check-ContractShadow (-not (Contract-Verify $r).Verified) 'ShadowArgumentBase64RefusedBeforeDecode'
$r.Arguments.Base64.PSObject.Properties.Remove('Length')
$r=Contract-Good (Contract-Packet);$r.Lifecycle.Stdout.Base64=Contract-Shadow ([string]::Empty) 'Length' 0
Check-ContractShadow (-not (Contract-Verify $r).Verified) 'ShadowDerivedSnapshotStringRefused'
$r.Lifecycle.Stdout.Base64.PSObject.Properties.Remove('Length')
$r=Contract-Good (Contract-Packet);$r.Lifecycle.Status=Contract-Shadow 'incomplete' 'Length' 10
Check-ContractShadow (-not (Contract-Verify $r).Verified) 'ShadowDerivedStatusStringRefused'
$r.Lifecycle.Status.PSObject.Properties.Remove('Length')
Check-Contract ($script:shadowFailures -eq 0) 'PassiveShadowInputsMustBeRefused'
$p=Contract-Packet;$r=Contract-Good $p
Check-Contract ((Contract-Verify $r).Verified -and $r.Lifecycle.Status -ceq 'incomplete' -and $r.Arguments.Base64 -ceq '' -and $r.CleanupStatus -ceq 'unknown' -and $r.StopIntent -ceq 'none') 'ConsistentIncompleteIsNotRuntimeSuccess'
$hex='544C524341503031010000000100000001000000730100000078010000006F1C000000323032362D31302D31375430303A30303A30302E303030303030305A1C000000323032362D31302D31375430303A30303A30302E303030303030305A0000000000000000'
$literal=New-Object byte[] ($hex.Length/2);for($i=0;$i -lt $literal.Length;$i++){$literal[$i]=[Convert]::ToByte($hex.Substring($i*2,2),16)}
$sha=[Security.Cryptography.SHA256]::Create();try{$expectedHash=[BitConverter]::ToString($sha.ComputeHash($literal)).Replace('-','')}finally{$sha.Dispose()}
Check-Contract ($r.BindingSHA256 -ceq $expectedHash) 'BindingHashMatchesIndependentLiteralBytes'
$p=Contract-Packet;$p.ArgumentBytes=[byte[]]@(65,0,45,78)
$p.Events=[object[]]@((Contract-Chunk 'stdout' ([byte[]]@(97,98,99))),(Contract-Chunk 'stderr' ([byte[]]@(0,255,13,10))),(Contract-Exit),(Contract-EOF 'stdout'),(Contract-EOF 'stderr'))
$r=Contract-Good $p
Check-Contract ((Contract-Verify $r).Verified -and $r.Arguments.Base64 -ceq 'QQAtTg==' -and $r.Lifecycle.Stdout.Base64 -ceq 'YWJj' -and $r.Lifecycle.Stderr.Base64 -ceq 'AP8NCg==' -and $r.Lifecycle.Status -ceq 'complete' -and $r.Lifecycle.Package.Classification -ceq 'refused') 'ExactArgumentsAndFailedOutputStillConsistent'
foreach($fixture in @(@{Events=[object[]]@();Status='incomplete';Stop='none'},@{Events=[object[]]@((Contract-Exit),(Contract-EOF 'stdout'),(Contract-EOF 'stderr'));Status='complete';Stop='none'},@{Events=[object[]]@((Contract-Tick 120000));Status='timed-out';Stop='would-request-owned-stop'},@{Events=[object[]]@((Contract-EOF 'stdout'),(Contract-EOF 'stdout'));Status='refused';Stop='would-refuse-and-stop-if-owned'},@{Events=[object[]]@((Contract-Tick 2),(Contract-Tick 1));Status='refused';Stop='would-refuse-and-stop-if-owned'},@{Events=[object[]]@((Contract-Exit),(Contract-EOF 'stdout'),(Contract-EOF 'stderr'),(Contract-Tick));Status='refused';Stop='would-refuse-and-stop-if-owned'})){
    $p=Contract-Packet;$p.Events=$fixture.Events;$r=Contract-Good $p
    Check-Contract ((Contract-Verify $r).Verified -and $r.Lifecycle.Status -ceq $fixture.Status -and $r.StopIntent -ceq $fixture.Stop -and $r.CleanupStatus -ceq 'unknown') ('LifecycleAndIntent-'+$fixture.Status)
}
$p=Contract-Packet;$p.Events=[object[]]@((Contract-Exit 0 1),(Contract-Chunk 'stdout' ([byte[]]@(97,98,99)) 2),(Contract-EOF 'stdout' 3),(Contract-EOF 'stderr' 3));$r=Contract-Good $p
Check-Contract ((Contract-Verify $r).Verified -and $r.Lifecycle.Status -ceq 'complete' -and $r.Lifecycle.Stdout.Base64 -ceq 'YWJj') 'BufferedBytesAfterExitRetained'
$p1=Contract-Packet;$p2=Contract-Packet
$a=Contract-Chunk 'stdout' ([byte[]]@(97));$b=Contract-Chunk 'stderr' ([byte[]]@(98))
$p1.Events=[object[]]@($a,$b);$p2.Events=[object[]]@($b,$a);$r1=Contract-Good $p1;$r2=Contract-Good $p2
Check-Contract ($r1.BindingSHA256 -cne $r2.BindingSHA256 -and $r1.Lifecycle.Stdout.Base64 -ceq $r2.Lifecycle.Stdout.Base64 -and $r1.Lifecycle.Stderr.Base64 -ceq $r2.Lifecycle.Stderr.Base64) 'HashBindsInterleavingNotOnlyAggregates'
$r1.Events=[object[]]@($r1.Events[1],$r1.Events[0]);Check-Contract (-not (Contract-Verify $r1).Verified) 'TraceReorderWithoutBindingUpdateRefused'
$p=Contract-Packet;$p.EndedUTC='2026-10-16T23:59:59.9999999Z';$p.Events=[object[]]@((Contract-Chunk 'stdout' ([byte[]]@(97))))
$r=Contract-Good $p;Check-Contract (-not $r.UtcOrderConsistent -and $r.Lifecycle.Stdout.Base64 -ceq 'YQ==' -and -not (Contract-Verify $r).Verified) 'ClockReversalRetainsBytesButCannotVerify'
foreach($bad in @('2026-10-17T00:00:00Z','2026-10-17T00:00:00.0000000+00:00','2026-02-30T00:00:00.0000000Z',1)){$p=Contract-Packet;$p.StartedUTC=$bad;Contract-Refusal $p 'BadUtcRefused'}
foreach($kind in @('schema','live','missing','extra','case','argument-odd','argument-bom','argument-surrogate','argument-oversize','argument-type','session','session-over','executable','encoding','origin','origin-surrogate','event-count','event-kind','chunk-bound','chunk-type')){
    $p=Contract-Packet
    switch($kind){
        schema{$p.Schema='1'} live{$p.SourceMode='live'} missing{$p.PSObject.Properties.Remove('SessionLabel')} extra{$p|Add-Member Extra 1}
        case{$p.PSObject.Properties.Remove('SessionLabel');$p|Add-Member sessionlabel 's'}
        argument-odd{$p.ArgumentBytes=[byte[]]@(1)} argument-bom{$p.ArgumentBytes=[byte[]]@(255,254)} argument-surrogate{$p.ArgumentBytes=[byte[]]@(0,216)}
        argument-oversize{$p.ArgumentBytes=New-Object byte[] 65538} argument-type{$p.ArgumentBytes=''} session{$p.SessionLabel='bad label'} session-over{$p.SessionLabel='s'*81}
        executable{$p.ExecutableDeclaration='x'*1025} encoding{$p.EncodingName='guess'} origin{$p.OriginDeclaration=''} origin-surrogate{$p.OriginDeclaration=[string][char]0xD800}
        event-count{$p.Events=New-Object object[] 4097} event-kind{$p.Events=[object[]]@([pscustomobject]@{Kind='other';ElapsedMs=0})}
        chunk-bound{$p.Events=[object[]]@((Contract-Chunk 'stdout' (New-Object byte[] 65537)))} chunk-type{$p.Events=[object[]]@((Contract-Chunk 'stdout' 'a'))}
    }
    Contract-Refusal $p ('InvalidPacket-'+$kind)
}
foreach($bad in @($null,@{},'packet')){Contract-Refusal $bad 'NonPlainPacketRefused'}
$p=Contract-Packet;$p.ArgumentBytes=New-Object byte[] 65536;$p.SessionLabel='s'*80;$p.ExecutableDeclaration='x'*1024;$p.OriginDeclaration='o'*1024;$p.Events=New-Object object[] 4096
for($i=0;$i -lt 4096;$i++){$p.Events[$i]=Contract-Tick}
Check-Contract ((Contract-Verify (Contract-Good $p)).Verified) 'ExactMetadataArgumentAndEventBounds'
$both=New-Object 'System.Collections.Generic.List[object]'
for($i=0;$i -lt 16;$i++){foreach($channel in @('stdout','stderr')){$both.Add((Contract-Chunk $channel (New-Object byte[] 65536)))}}
$p=Contract-Packet;$p.Events=[object[]]$both.ToArray();$r=Contract-Good $p
Check-Contract ((Contract-Verify $r).Verified -and $r.Lifecycle.Stdout.ByteLength -eq 1048576 -and $r.Lifecycle.Stderr.ByteLength -eq 1048576) 'ExactTwoMiBWholeTraceBound'
$p.Events=[object[]](@((Contract-Tick 120000))+$both.ToArray()+@((Contract-Chunk 'stdout' ([byte[]]@(1)) 120000)))
Contract-Refusal $p 'PostTimeoutBytesStillCountTowardPacketCap'
$one=New-Object 'System.Collections.Generic.List[object]';for($i=0;$i -lt 16;$i++){$one.Add((Contract-Chunk 'stdout' (New-Object byte[] 65536)))};$one.Add((Contract-Chunk 'stdout' ([byte[]]@(1))))
$p=Contract-Packet;$p.Events=[object[]]$one.ToArray();$r=Contract-Good $p
Check-Contract ((Contract-Verify $r).Verified -and $r.Lifecycle.Reason -ceq 'stream-limit' -and $r.Lifecycle.Stdout.ByteLength -eq 1048576 -and $r.Events.Count -eq 17 -and $r.Events[16].Data.Base64 -ceq 'AQ==') 'AdmittedPrefixAndRejectedTraceBothRetained'
$owned=[byte[]]@(97,98,99);$p=Contract-Packet;$p.ArgumentBytes=[byte[]]@(65,0);$p.Events=[object[]]@((Contract-Chunk 'stdout' $owned));$r=Contract-Good $p;$oldHash=$r.BindingSHA256
$owned[0]=0;$p.ArgumentBytes[0]=66;$p.Events[0].ElapsedMs=1
Check-Contract ((Contract-Verify $r).Verified -and $r.BindingSHA256 -ceq $oldHash -and $r.Arguments.Base64 -ceq 'QQA=' -and $r.Events[0].Data.Base64 -ceq 'YWJj') 'OwnedSnapshotsSurviveCallerMutation'
foreach($kind in @('schema','policy','mode','metadata','utc','trace','arguments-data','arguments-length','arguments-hash','arguments-whitespace','arguments-padbits','arguments-over','binding','status','package','snapshot','intent','cleanup','clock-flag','action','origin-verified','acceptance','runtime','extra','missing')){
    $p=Contract-Packet;$p.ArgumentBytes=[byte[]]@(65,0);$p.Events=[object[]]@((Contract-Chunk 'stdout' ([byte[]]@(97))))
    $bad=Contract-Good $p
    switch($kind){
        schema{$bad.Schema=2} policy{$bad.Policy='other'} mode{$bad.SourceMode='live'} metadata{$bad.SessionLabel='other'} utc{$bad.EndedUTC='2026-10-17T00:00:01.0000000Z'}
        trace{$bad.Events[0].ElapsedMs=1} arguments-data{$bad.Arguments.Base64='QgA='} arguments-length{$bad.Arguments.ByteLength=1} arguments-hash{$bad.Arguments.SHA256='A'*64}
        arguments-whitespace{$bad.Arguments.Base64+="`n"} arguments-padbits{$bad.Arguments.Base64='QQB='} arguments-over{$bad.Arguments.Base64='A'*87388}
        binding{$bad.BindingSHA256='A'*64} status{$bad.Lifecycle.Status='complete'} package{$bad.Lifecycle.Package='invented'} snapshot{$bad.Lifecycle.Stdout.Base64='Yg=='}
        intent{$bad.StopIntent='would-request-owned-stop'} cleanup{$bad.CleanupStatus='known-clean'} clock-flag{$bad.UtcOrderConsistent=$false}
        action{$bad.LiveActionsPerformed=$true} origin-verified{$bad.OriginVerified=$true} acceptance{$bad.Acceptance=$true} runtime{$bad.RuntimeApproved=$true}
        extra{$bad|Add-Member Extra 1} missing{$bad.PSObject.Properties.Remove('Arguments')}
    }
    Check-Contract (-not (Contract-Verify $bad).Verified) ('TamperedRecordRefused-'+$kind)
}
$script:contractCallbacks=0
$p=Contract-Packet;$p|Add-Member -MemberType ScriptMethod -Name GetType -Force -Value {$script:contractCallbacks++;[System.Management.Automation.PSCustomObject]}
$e=Contract-Tick;$e|Add-Member -MemberType ScriptMethod -Name GetType -Force -Value {$script:contractCallbacks++;[System.Management.Automation.PSCustomObject]};$p.Events=[object[]]@($e)
$r=Contract-Good $p;$r.Arguments|Add-Member -MemberType ScriptMethod -Name GetType -Force -Value {$script:contractCallbacks++;[System.Management.Automation.PSCustomObject]}
$r.Lifecycle|Add-Member -MemberType ScriptMethod -Name GetType -Force -Value {$script:contractCallbacks++;[System.Management.Automation.PSCustomObject]}
Check-Contract ((Contract-Verify $r).Verified -and $script:contractCallbacks -eq 0) 'NestedGetTypeMethodsNeverDispatched'
$p=Contract-Packet;$p.PSObject.Properties.Remove('SourceMode');$p|Add-Member ScriptProperty SourceMode {$script:contractCallbacks++;'synthetic'}
Contract-Refusal $p 'PacketScriptPropertyRefusedWithoutEvaluation'
foreach($target in @('arguments','lifecycle')){
    $r=Contract-Good (Contract-Packet)
    if($target -eq 'arguments'){$r.Arguments.PSObject.Properties.Remove('Base64');$r.Arguments|Add-Member ScriptProperty Base64 {$script:contractCallbacks++;''}}
    else{$r.Lifecycle.PSObject.Properties.Remove('Status');$r.Lifecycle|Add-Member ScriptProperty Status {$script:contractCallbacks++;'incomplete'}}
    Check-Contract (-not (Contract-Verify $r).Verified -and $script:contractCallbacks -eq 0) 'NestedExecutablePropertyRefused'
}
$descriptor=New-SyntheticCaptureRecord $null;Check-Contract (-not (Contract-Verify $descriptor).Verified) 'RefusalDescriptorIsNotRecord'
$xml='#< CLIXML'+"`n"+'<Objs Version="1.1.0.1" xmlns="http://schemas.microsoft.com/powershell/2004/04"><Obj S="progress" RefId="0"><TN RefId="0"><T>System.Management.Automation.PSCustomObject</T><T>System.Object</T></TN><MS><I64 N="SourceId">1</I64><PR N="Record"><AV>fixture</AV><AI>0</AI><Nil /><PI>-1</PI><PC>-1</PC><T>Completed</T><SR>-1</SR><SD> </SD></PR></MS></Obj></Objs>'
foreach($enc in @('utf-8','utf-16le')){
    $codec=if($enc -eq 'utf-8'){[Text.Encoding]::UTF8}else{[Text.Encoding]::Unicode};$bytes=$codec.GetBytes($xml)
    $p=Contract-Packet;$p.EncodingName=$enc;$p.Events=[object[]]@((Contract-Chunk 'stderr' ([byte[]]$bytes[0..2])),(Contract-Chunk 'stderr' ([byte[]]$bytes[3..($bytes.Length-1)])),(Contract-Exit),(Contract-EOF 'stdout'),(Contract-EOF 'stderr'))
    $r=Contract-Good $p;Check-Contract ((Contract-Verify $r).Verified -and $r.Lifecycle.Package.Classification -ceq 'progress-only' -and $r.Lifecycle.Package.ProgressCount -eq 1 -and $r.Lifecycle.Stderr.Base64 -ceq [Convert]::ToBase64String($bytes)) 'SplitProgressRawBytesPreserved'
}
Write-Output 'LIMIT: synthetic declarations/trace consistency only;no live capture,clock,cleanup,persistence or acceptance.'
