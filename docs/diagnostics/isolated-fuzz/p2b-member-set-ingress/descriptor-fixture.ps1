# Definitions only. Test fixture compilation occurs solely in approved hosts.
function DG-Need([bool]$Condition){if(-not $Condition){throw 'Descriptor assertion'}}
function DG-NewMessage {
    $f=(Get-Content -LiteralPath (Join-Path $PSScriptRoot 'fixtures.json') -Raw|ConvertFrom-Json)[0]
    [pscustomobject][ordered]@{Version=[uint16]1;Kind=[byte]1;Sender=[byte]1;Receiver=[byte]2
        BindingDigest=[Convert]::FromHexString($f.BindingHex);SessionId=[Convert]::FromHexString($f.SessionHex)
        Sequence=[uint64]::Parse('72623859790382856');OriginTick=[long]10;Frequency=[long]1000
        DeadlineTick=[long]120010;ObservedTick=[long]11;Payload=[Convert]::FromHexString($f.PayloadHex)}
}
function DG-NewExpected($Message){
    [pscustomobject][ordered]@{Schema=[int]1;Sender=[byte]1;Receiver=[byte]2
        BindingDigest=$Message.BindingDigest.Clone();SessionId=$Message.SessionId.Clone()
        NextSequence=[uint64]::Parse('72623859790382856');OriginTick=[long]10;Frequency=[long]1000
        DeadlineTick=[long]120010;StopAtTick=[long]113510;ArmSequence=[uint64]0
        StopSequence=[uint64]0;LastNonHeartbeatSequence=[uint64]0}
}
function DG-RawMembers($Value){
    $script:memberSetEvidence=[ordered]@{Stage='MemberSet';SetType=$null;RawType=$null;RetainedCount=$null;SubtypeCount=$null;RawCallbacks=$null;OriginalCallbacks=$null}
    $typeMethod=[object].GetMethod('GetType',[type[]]@())
    $set=$Value.psextended
    $setBase=$set.PSObject.BaseObject
    $setType=$typeMethod.Invoke($setBase,$null)
    $script:memberSetEvidence.SetType=$setType.FullName
    DG-Need ($setType -eq [Management.Automation.PSMemberSet])
    $script:memberSetEvidence.Stage='RawField'
    $field=[Management.Automation.PSMemberSet].GetField('internalMembers',[Reflection.BindingFlags]'NonPublic,Instance')
    DG-Need ($null -ne $field -and $field.DeclaringType -eq [Management.Automation.PSMemberSet])
    $raw=$field.GetValue($setBase)
    $rawType=$typeMethod.Invoke($raw.PSObject.BaseObject,$null)
    $script:memberSetEvidence.RawType=$rawType.FullName
    DG-Need ($rawType -eq $field.FieldType)
    $script:memberSetEvidence.Stage='RawReturned'
    return ,$raw
}
function DG-Install($Owner,[string]$Name){
    $value=$Owner.PSObject.Properties[$Name].Value
    $Owner.PSObject.Properties.Remove($Name)
    $Owner.PSObject.Properties.Add([P2BDescriptorFixture.CounterNote]::new($Name,$value))
}
function DG-Rejected($Result){
    DG-Need (-not $Result.Accepted -and $Result.Reason -ceq 'invalid-input' -and
        $null -eq $Result.Bytes -and $null -eq $Result.Message -and $null -eq $Result.NextSequence -and
        -not $Result.SourceVerified -and -not $Result.LiveActionsPerformed -and -not $Result.RuntimeApproved)
}
function Invoke-P2BDescriptorRegression([bool]$Probe=$false,[bool]$Trace=$false){
    $script:descriptorCase='FixtureObservable'
    if($null -ne ('P2BDescriptorFixture.CounterNote' -as [type])){throw 'Fresh fixture type required'}
    $source=[IO.File]::ReadAllText((Join-Path $PSScriptRoot 'descriptor-fixture.cs'))
    Add-Type -TypeDefinition $source -ReferencedAssemblies ([Management.Automation.PSObject].Assembly.Location) -ErrorAction Stop
    $control=[P2BDescriptorFixture.CounterNote]::new('Version',[uint16]1)
    [P2BDescriptorFixture.CounterNote]::Reset()
    $null=$control.MemberType;$null=$control.Value;$null=$control.Copy()
    DG-Need ([P2BDescriptorFixture.CounterNote]::MemberReads -gt 0 -and
        [P2BDescriptorFixture.CounterNote]::ValueReads -gt 0 -and [P2BDescriptorFixture.CounterNote]::Copies -gt 0)
    if($Trace){[Console]::Out.WriteLine('PASS: P2BDescriptorFixtureObservable')}
    $script:descriptorCase='RawIngress'
    $m=DG-NewMessage;DG-Install $m 'Version'
    [P2BDescriptorFixture.CounterNote]::Reset()
    $raw=DG-RawMembers $m;$count=0;$subtypes=0;$typeMethod=[object].GetMethod('GetType',[type[]]@())
    foreach($descriptor in $raw){$count++;if($typeMethod.Invoke($descriptor.PSObject.BaseObject,$null) -eq [P2BDescriptorFixture.CounterNote]){$subtypes++}}
    $script:memberSetEvidence.RetainedCount=$count;$script:memberSetEvidence.SubtypeCount=$subtypes
    $script:memberSetEvidence.RawCallbacks=[P2BDescriptorFixture.CounterNote]::Total
    DG-Need ($count -eq 12 -and $subtypes -eq 1 -and [P2BDescriptorFixture.CounterNote]::Total -eq 0)
    if($Trace){[Console]::Out.WriteLine('PASS: P2BDescriptorRawIngress')}
    if($Probe){
        $script:descriptorCase='OriginalCallbackObserved'
        [P2BDescriptorFixture.CounterNote]::Reset();$result=ConvertTo-P2BControlFrame $m
        $script:memberSetEvidence.OriginalCallbacks=[P2BDescriptorFixture.CounterNote]::Total
        DG-Need ([P2BDescriptorFixture.CounterNote]::Total -gt 0)
        if($Trace){[Console]::Out.WriteLine('PASS: P2BDescriptorOriginalCallbackObserved')}
        $script:descriptorCase='Passivity';DG-Need ([P2BDescriptorFixture.CounterNote]::Total -eq 0)
        return
    }
    $script:descriptorCase='Passivity'
    foreach($mode in @('encode','binding-message','binding-expected')){
        $sample=DG-NewMessage;$expected=DG-NewExpected $sample
        $owner=if($mode -ceq 'binding-expected'){$expected}else{$sample}
        $names=@($owner.PSObject.Properties.Name)
        foreach($name in $names){
            $m=DG-NewMessage;$e=DG-NewExpected $m
            if($mode -ceq 'binding-expected'){DG-Install $e $name}else{DG-Install $m $name}
            [P2BDescriptorFixture.CounterNote]::Reset()
            $result=if($mode -ceq 'encode'){ConvertTo-P2BControlFrame $m}else{Test-P2BControlBinding $m $e}
            $callbacks=[P2BDescriptorFixture.CounterNote]::Total
            DG-Need ($callbacks -eq 0);DG-Rejected $result
        }
    }
    # The SDK's own PSNoteProperty subtype must also be refused without coercion.
    $m=DG-NewMessage;$m.PSObject.Properties.Remove('Version')
    $m.PSObject.Properties.Add([Management.Automation.PSVariableProperty]::new([Management.Automation.PSVariable]::new('Version',[uint16]1)))
    DG-Rejected (ConvertTo-P2BControlFrame $m)
    $m=DG-NewMessage;$e=DG-NewExpected $m
    DG-Need ((ConvertTo-P2BControlFrame $m).Accepted -and (Test-P2BControlBinding $m $e).Accepted)
    if($Trace){[Console]::Out.WriteLine('PASS: P2BDescriptorPassivity')}
}
function DG-SaveEvidence([string]$Path,$Failure){
    if([string]::IsNullOrEmpty($Path)){return}
    $evidence=$null;$variable=Get-Variable -Name memberSetEvidence -Scope Script -ErrorAction SilentlyContinue
    if($null -ne $variable){$evidence=$variable.Value}
    $counts=$null
    if($null -ne ('P2BDescriptorFixture.CounterNote' -as [type])){
        $counts=[pscustomobject]@{MemberReads=[P2BDescriptorFixture.CounterNote]::MemberReads;ValueReads=[P2BDescriptorFixture.CounterNote]::ValueReads;Copies=[P2BDescriptorFixture.CounterNote]::Copies;Writes=[P2BDescriptorFixture.CounterNote]::Writes;Total=[P2BDescriptorFixture.CounterNote]::Total}
    }
    $errorData=$null
    if($null -ne $Failure){
        $chain=[Collections.Generic.List[object]]::new();$exception=$Failure.Exception
        for($i=0;$null -ne $exception -and $i -lt 8;$i++){$chain.Add([pscustomobject]@{Type=$exception.GetType().FullName;Message=$exception.Message});$exception=$exception.InnerException}
        $errorData=[pscustomobject]@{FullyQualifiedErrorId=$Failure.FullyQualifiedErrorId;Chain=$chain.ToArray();Stack=$Failure.ScriptStackTrace;Position=$Failure.InvocationInfo.PositionMessage}
    }
    $record=[pscustomobject]@{Schema=1;Evidence=$evidence;Counts=$counts;Error=$errorData;RuntimeApproved=$false}
    [IO.File]::WriteAllText($Path,($record|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($false))
}
