# Evidence only: fixed SDK field metadata, no product/value/virtual getter reads.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$leaf=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../p2b-member-set-ingress'))
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
$record=[ordered]@{Schema=1;Mode='p2b-member-inventory-diagnostic-only';HostVersion=$PSVersionTable.PSVersion.ToString();Core=($PSVersionTable.PSEdition -ceq 'Core');X64=[Environment]::Is64BitProcess;Admin=$admin;Stage='Context';Captured=$false;FixtureObservable=$false;RawType=$null;MemberCount=$null;CountersAfterRaw=$null;CountersAfterInventory=$null;Members=@();Error=$null;MemberValuesRead=$false;ProductImported=$false;ProductCalled=$false;RuntimeApproved=$false}
function ID-Type($Value){
    if($null -eq $Value){return '<null>'}
    $getType=[object].GetMethod('GetType',[type[]]@());$invoke=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
    $argsLocal=[object[]]@($Value.PSObject.BaseObject,$null)
    return $invoke.Invoke($getType,$argsLocal).FullName
}
function ID-Counts {
    [pscustomobject]@{MemberReads=[P2BDescriptorFixture.CounterNote]::MemberReads;ValueReads=[P2BDescriptorFixture.CounterNote]::ValueReads;Copies=[P2BDescriptorFixture.CounterNote]::Copies;Writes=[P2BDescriptorFixture.CounterNote]::Writes;Total=[P2BDescriptorFixture.CounterNote]::Total}
}
try{
    if($admin -or $record.HostVersion -cne '7.6.5' -or -not $record.Core -or -not $record.X64){throw 'Ordinary pinned PS7 host required'}
    $record.Stage='FixtureCompile';. (Join-Path $leaf 'descriptor-fixture.ps1')
    Add-Type -TypeDefinition ([IO.File]::ReadAllText((Join-Path $leaf 'descriptor-fixture.cs'))) -ReferencedAssemblies ([Management.Automation.PSObject].Assembly.Location) -ErrorAction Stop
    $control=[P2BDescriptorFixture.CounterNote]::new('Version',[uint16]1)
    [P2BDescriptorFixture.CounterNote]::Reset();$null=$control.MemberType;$null=$control.Value;$null=$control.Copy()
    $record.FixtureObservable=([P2BDescriptorFixture.CounterNote]::MemberReads -gt 0 -and [P2BDescriptorFixture.CounterNote]::ValueReads -gt 0 -and [P2BDescriptorFixture.CounterNote]::Copies -gt 0)
    if(-not $record.FixtureObservable){throw 'Fixture control not observable'}
    $record.Stage='FixedMetadata'
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly';$baseType=[Management.Automation.PSMemberInfo]
    $nameField=$baseType.GetField('name',$flags)
    $reservedField=$baseType.GetField('<IsReservedMember>k__BackingField',$flags)
    $hiddenField=$baseType.GetField('<IsHidden>k__BackingField',$flags)
    $instanceField=$baseType.GetField('<IsInstance>k__BackingField',$flags)
    foreach($field in @($nameField,$reservedField,$hiddenField,$instanceField)){if($null -eq $field -or $field.DeclaringType -ne $baseType){throw 'Fixed SDK field unavailable'}}
    $record.Stage='Construct';$m=DG-NewMessage;DG-Install $m 'Version'
    [P2BDescriptorFixture.CounterNote]::Reset()
    $record.Stage='RawFunction';$raw=DG-RawMembers $m;$record.CountersAfterRaw=ID-Counts;$record.RawType=ID-Type $raw
    $record.Stage='Inventory';$items=[Collections.Generic.List[object]]::new();$index=0
    foreach($descriptor in $raw){
        $index++;$descriptorBase=$descriptor.PSObject.BaseObject
        $items.Add([pscustomobject]@{Index=$index;Name=$nameField.GetValue($descriptorBase);CLRType=(ID-Type $descriptor);IsReservedMember=$reservedField.GetValue($descriptorBase);IsHidden=$hiddenField.GetValue($descriptorBase);IsInstance=$instanceField.GetValue($descriptorBase)})
    }
    $record.MemberCount=$index;$record.Members=$items.ToArray();$record.CountersAfterInventory=ID-Counts
    $record.Stage='Captured';$record.Captured=$true
}catch{
    $caught=$_;$chain=[Collections.Generic.List[object]]::new();$exception=$caught.Exception
    for($i=0;$null -ne $exception -and $i -lt 8;$i++){$chain.Add([pscustomobject]@{Type=$exception.GetType().FullName;Message=$exception.Message});$exception=$exception.InnerException}
    $record.Error=[pscustomobject]@{FullyQualifiedErrorId=$caught.FullyQualifiedErrorId;Chain=$chain.ToArray();Stack=$caught.ScriptStackTrace;Position=$caught.InvocationInfo.PositionMessage}
    if($null -ne ('P2BDescriptorFixture.CounterNote' -as [type])){$record.CountersAfterInventory=ID-Counts}
    $record.Captured=$true
}
$record|ConvertTo-Json -Depth 12
exit 0
