# Test integration only: no product imports or public APIs.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$OutputEncoding=[Text.UTF8Encoding]::new($false)
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
$record=[ordered]@{Schema=1;Mode='p2b-receiver-retention-tests';HostVersion=$PSVersionTable.PSVersion.ToString();Core=($PSVersionTable.PSEdition -ceq 'Core');X64=[Environment]::Is64BitProcess;Admin=$admin;Stage='Context';Captured=$false;Passed=$false;Rows=@();References=@();Error=$null;ProductImported=$false;ProductCalled=$false;FixtureCompiled=$false;FixtureObservable=$false;RuntimeApproved=$false}
$script:typeMethod=[object].GetMethod('GetType',[type[]]@())
$script:invokeMethod=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
$script:referenceMethod=[object].GetMethod('ReferenceEquals',[type[]]@([object],[object]))
$script:fieldReadMethod=[Reflection.FieldInfo].GetMethod('GetValue',[type[]]@([object]))
$script:referenceIds=[Collections.Generic.Dictionary[object,int]]::new([Collections.Generic.ReferenceEqualityComparer]::Instance)
$script:idLookup=$script:referenceIds.GetType().GetMethod('TryGetValue',[type[]]@([object],[int].MakeByRefType()))
$script:idAdd=$script:referenceIds.GetType().GetMethod('Add',[type[]]@([object],[int]))
. (Join-Path $PSScriptRoot 'receiver-fixture.ps1')
try{
    Need (-not $admin -and $record.HostVersion -ceq '7.6.5' -and $record.Core -and $record.X64)
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $sdk=[Management.Automation.PSObject].Assembly
    $script:shimType=$sdk.GetType('System.Management.Automation.PSInternalMemberSet',$true)
    $reserved=$sdk.GetType('System.Management.Automation.ReservedNameMembers',$true)
    $script:objectCtor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
    $script:emptyCtor=[Management.Automation.PSObject].GetConstructor([type[]]@())
    $script:asObject=[Management.Automation.PSObject].GetMethod('AsPSObject',[type[]]@([object]))
    $script:baseField=[Management.Automation.PSObject].GetField('_immediateBaseObject',$flags)
    $script:tableField=[Management.Automation.PSObject].GetField('_instanceMembers',$flags)
    $script:flagsField=[Management.Automation.PSObject].GetField('_flags',$flags)
    $script:keyMethod=[Management.Automation.PSObject].GetMethod('GetKeyForResurrectionTables',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
    $script:psobjectFactory=$reserved.GetMethod('GeneratePSObjectMemberSet',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
    $script:extendedFactory=$reserved.GetMethod('GeneratePSExtendedMemberSet',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
    $script:rawField=[Management.Automation.PSMemberSet].GetField('internalMembers',$flags)
    $script:ownerField=[Management.Automation.PSMemberInfo].GetField('instance',$flags)
    $script:nameField=[Management.Automation.PSMemberInfo].GetField('name',$flags)
    $script:hiddenField=[Management.Automation.PSMemberInfo].GetField('<IsHidden>k__BackingField',$flags)
    $script:reservedField=[Management.Automation.PSMemberInfo].GetField('<IsReservedMember>k__BackingField',$flags)
    $script:instanceField=[Management.Automation.PSMemberInfo].GetField('<IsInstance>k__BackingField',$flags)
    $script:shimOwnerField=$script:shimType.GetField('_psObject',$flags)
    $script:constructorOwnerField=[Management.Automation.PSMemberSet].GetField('_constructorPSObject',$flags)

    $record.Stage='FixtureCompile'
    Need ($null -eq ('P2BDescriptorFixture.CounterNote' -as [type]))
    Add-Type -TypeDefinition ([IO.File]::ReadAllText((Join-Path $PSScriptRoot 'descriptor-fixture.cs'))) -ReferencedAssemblies ([Management.Automation.PSObject].Assembly.Location) -ErrorAction Stop
    $record.FixtureCompiled=$true
    $control=[P2BDescriptorFixture.CounterNote]::new('Observable',[int]1)
    [P2BDescriptorFixture.CounterNote]::Reset()
    $null=$control.MemberType;$null=$control.Value;$null=$control.Copy();$control.Value=[int]2
    $record.FixtureObservable=[P2BDescriptorFixture.CounterNote]::MemberReads -gt 0 -and [P2BDescriptorFixture.CounterNote]::ValueReads -gt 0 -and [P2BDescriptorFixture.CounterNote]::Copies -gt 0 -and [P2BDescriptorFixture.CounterNote]::Writes -gt 0
    Need $record.FixtureObservable
    $rawAdd=$script:rawField.FieldType.GetMethod('Add',[type[]]@([Management.Automation.PSMemberInfo]))
    Need ($null -ne $rawAdd -and $rawAdd.DeclaringType -eq $script:rawField.FieldType)
    foreach($kind in @('Byte','UInt16','UInt32','UInt64','Int32','Int64','Boolean','String','Bytes')){
        foreach($variant in @('clean','visible-note','hidden-note','visible-subclass','hidden-subclass')){
            $record.Stage=$kind+':'+$variant
            $row=[pscustomobject][ordered]@{Number=($record.Rows.Count+1);Kind=$kind;Variant=$variant;Phase='Construction';SourceWrapperID=$null;SourceBaseID=$null;TagID=0;Before=$null;AfterCapture=$null;AfterRelay=$null;InputWasWrapper=$null;CaptureRetained=$null;HandleRetained=$null;CaptureVerdict=$null;RelayVerdict=$null;CounterBefore=$null;CounterAfter=$null;Negative=$null;Passed=$false}
            $record.Rows+=,$row
            [P2BDescriptorFixture.CounterNote]::Reset()
            $source=New-TestSource $kind
            $null=Generate 'psobject' $source[0];$null=Generate 'psextended' $source[0]
            $row.SourceWrapperID=Ref-ID $source[0]
            $row.SourceBaseID=Ref-ID ((Read-Field $script:baseField $source[0])[0])
            $row.InputWasWrapper=(Raw-Type $source[0]) -eq [Management.Automation.PSObject]
            if($variant -cne 'clean'){
                $tag=$null
                if($variant -clike '*subclass'){$tag=[P2BDescriptorFixture.CounterNote]::new('AttachedProbe',[int]42)}
                else{$tag=[Management.Automation.PSNoteProperty]::new('AttachedProbe',[int]42)}
                if($variant -clike 'hidden-*'){$script:hiddenField.SetValue($tag,$true)}
                $argsLocal=[object[]]::new(1);$argsLocal[0]=$tag
                $raw=(Read-Field $script:tableField $source[0])[0]
                $null=$rawAdd.Invoke($raw,$argsLocal)
                $row.TagID=Ref-ID $tag
            }
            # Observe the actual already-decorated source BEFORE capture.
            $row.Before=Snapshot $source[0]
            $row.CounterBefore=[P2BDescriptorFixture.CounterNote]::Total
            $row.Phase='Capture'
            $handle=Capture-P2BTestReceiver $source
            $row.CaptureRetained=Same $source[0] $handle[0]
            $row.AfterCapture=Observe-P2BTestReceiver $handle
            $row.CaptureVerdict=Get-P2BRetentionVerdict $row.Before $row.AfterCapture $row.TagID ($variant -clike 'hidden-*') ($variant -clike '*subclass')
            $row.Phase='Relay'
            $again=Relay-P2BTestReceiver $handle
            $row.HandleRetained=Same $handle $again
            $row.AfterRelay=Observe-P2BTestReceiver $again
            $row.RelayVerdict=Get-P2BRetentionVerdict $row.Before $row.AfterRelay $row.TagID ($variant -clike 'hidden-*') ($variant -clike '*subclass')
            $row.CounterAfter=[P2BDescriptorFixture.CounterNote]::Total
            $row.Phase='ConditionsSaved'
            Need ($row.InputWasWrapper -and $row.CaptureRetained -and $row.HandleRetained -and $row.CounterBefore -eq 0 -and $row.CounterAfter -eq 0 -and (All-P2BRetentionConditions $row.CaptureVerdict) -and (All-P2BRetentionConditions $row.RelayVerdict))
            # Wrong-receiver control: equal contents in a freshly allocated base.
            $wrongSource=New-TestSource $kind
            $null=Generate 'psobject' $wrongSource[0];$null=Generate 'psextended' $wrongSource[0]
            $wrong=Snapshot $wrongSource[0]
            $verdict=Get-P2BRetentionVerdict $row.Before $wrong $row.TagID ($variant -clike 'hidden-*') ($variant -clike '*subclass')
            $row.Negative=[pscustomobject]@{Snapshot=$wrong;Verdict=$verdict;ReplacementDetected=(-not (All-P2BRetentionConditions $verdict));CounterTotal=[P2BDescriptorFixture.CounterNote]::Total}
            Need ($row.Negative.ReplacementDetected -and $row.Negative.CounterTotal -eq 0)
            $row.Phase='Completed';$row.Passed=$true
        }
    }
    Need ($record.Rows.Count -eq 45)
    $record.Stage='Passed';$record.Passed=$true;$record.Captured=$true
}catch{
    $caught=$_;$chain=[Collections.Generic.List[object]]::new();$errorItem=$caught.Exception
    for($n=0;$null -ne $errorItem -and $n -lt 8;$n++){$chain.Add([pscustomobject]@{Type=$errorItem.GetType().FullName;Message=$errorItem.Message});$errorItem=$errorItem.InnerException}
    $record.Error=[pscustomobject]@{Chain=$chain.ToArray();Stack=$caught.ScriptStackTrace;Position=$caught.InvocationInfo.PositionMessage}
    $record.Captured=$true
}
$record|ConvertTo-Json -Depth 24
if($record.Passed){exit 0}else{exit 1}
