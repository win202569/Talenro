# One evidence-only host; preserved helper/fixture, no product import or call.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$leaf=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../p2b-descriptor-ingress'))
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
$record=[ordered]@{Schema=1;Mode='p2b-rawingress-diagnostic-only';HostVersion=$PSVersionTable.PSVersion.ToString();Core=($PSVersionTable.PSEdition -ceq 'Core');X64=[Environment]::Is64BitProcess;Admin=$admin;Stage='Context';Captured=$false;FixtureObservable=$false;RootBaseType=$null;ReceiverStorageType=$null;RawResultType=$null;DescriptorCount=$null;CounterSubtypeCount=$null;CountersAfterRaw=$null;CountersAfterEnumeration=$null;Descriptors=@();OriginalGateWouldPass=$null;Error=$null;ProductImported=$false;ProductCalled=$false;RuntimeApproved=$false}
function RD-Type($Value){
    if($null -eq $Value){return '<null>'}
    $getType=[object].GetMethod('GetType',[type[]]@())
    $invoke=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
    $argsLocal=[object[]]@($Value.PSObject.BaseObject,$null)
    return $invoke.Invoke($getType,$argsLocal).FullName
}
function RD-Counts {
    [pscustomobject]@{MemberReads=[P2BDescriptorFixture.CounterNote]::MemberReads;ValueReads=[P2BDescriptorFixture.CounterNote]::ValueReads;Copies=[P2BDescriptorFixture.CounterNote]::Copies;Writes=[P2BDescriptorFixture.CounterNote]::Writes;Total=[P2BDescriptorFixture.CounterNote]::Total}
}
try{
    if($admin -or $record.HostVersion -cne '7.6.5' -or -not $record.Core -or -not $record.X64){throw 'Ordinary pinned PS7 host required'}
    $record.Stage='FixtureCompile'
    . (Join-Path $leaf 'descriptor-fixture.ps1')
    Add-Type -TypeDefinition ([IO.File]::ReadAllText((Join-Path $leaf 'descriptor-fixture.cs'))) -ReferencedAssemblies ([Management.Automation.PSObject].Assembly.Location) -ErrorAction Stop
    $control=[P2BDescriptorFixture.CounterNote]::new('Version',[uint16]1)
    [P2BDescriptorFixture.CounterNote]::Reset();$null=$control.MemberType;$null=$control.Value;$null=$control.Copy()
    $record.FixtureObservable=([P2BDescriptorFixture.CounterNote]::MemberReads -gt 0 -and [P2BDescriptorFixture.CounterNote]::ValueReads -gt 0 -and [P2BDescriptorFixture.CounterNote]::Copies -gt 0)
    if(-not $record.FixtureObservable){throw 'Fixture control not observable'}
    $record.Stage='Construct'
    $m=DG-NewMessage;DG-Install $m 'Version'
    $record.RootBaseType=RD-Type $m
    $typeMethod=[object].GetMethod('GetType',[type[]]@());$invoke=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
    $receiverArgs=[object[]]::new(2);$receiverArgs[0]=$m.PSObject;$receiverArgs[1]=$null
    $record.ReceiverStorageType=$invoke.Invoke($typeMethod,$receiverArgs).FullName
    [P2BDescriptorFixture.CounterNote]::Reset()
    $record.Stage='RawFunction'
    $raw=DG-RawMembers $m
    $record.CountersAfterRaw=RD-Counts
    $record.RawResultType=RD-Type $raw
    $record.Stage='Enumerate'
    $descriptors=[Collections.Generic.List[object]]::new();$count=0;$subtypes=0
    foreach($descriptor in $raw){
        $count++;$type=RD-Type $descriptor
        if($type -ceq 'P2BDescriptorFixture.CounterNote'){$subtypes++}
        $descriptors.Add([pscustomobject]@{Index=$count;CLRType=$type})
    }
    $record.CountersAfterEnumeration=RD-Counts
    $record.Descriptors=$descriptors.ToArray();$record.DescriptorCount=$count;$record.CounterSubtypeCount=$subtypes
    $record.OriginalGateWouldPass=($count -eq 12 -and $subtypes -eq 1 -and [P2BDescriptorFixture.CounterNote]::Total -eq 0)
    $record.Stage='Captured';$record.Captured=$true
}catch{
    $caught=$_;$chain=[Collections.Generic.List[object]]::new();$exception=$caught.Exception
    for($i=0;$null -ne $exception -and $i -lt 8;$i++){
        $chain.Add([pscustomobject]@{Type=$exception.GetType().FullName;Message=$exception.Message})
        $exception=$exception.InnerException
    }
    $record.Error=[pscustomobject]@{FullyQualifiedErrorId=$caught.FullyQualifiedErrorId;Chain=$chain.ToArray();ScriptStackTrace=$caught.ScriptStackTrace;Position=$caught.InvocationInfo.PositionMessage}
    if($null -ne ('P2BDescriptorFixture.CounterNote' -as [type])){$record.CountersAfterEnumeration=RD-Counts}
    $record.Captured=$true
}
$record|ConvertTo-Json -Depth 12
exit 0
