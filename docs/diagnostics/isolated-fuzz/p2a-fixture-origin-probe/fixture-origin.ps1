# Fixture-only characterization. No product source, imports or product calls.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:originReference=[object].GetMethod('ReferenceEquals',[type[]]@([object],[object]))
$script:originType=[object].GetMethod('GetType',[type[]]@())
$script:originName='P2AFixtureOriginProbe'
function New-TBRequest([object]$OriginTick=[long]0,[object]$Frequency=[long]1000,[object]$ObservedTick=[long]0,[object]$StopTick=$null,[object]$Schema=[int]1){
    [pscustomobject]@{Schema=$Schema;OriginTick=$OriginTick;Frequency=$Frequency;ObservedTick=$ObservedTick;StopTick=$StopTick}
}
function New-OriginFreshInt32([int]$Value){
    $storage=[Array]::CreateInstance([int],1)
    $storage.SetValue($Value,0)
    return ,($storage.GetValue(0))
}
function Origin-Reference($Left,$Right){
    $parameters=[object[]]@($Left,$Right)
    return [bool]$script:originReference.Invoke($null,$parameters)
}
function Origin-Need([bool]$Value){if(-not $Value){throw 'fixture-invalid'}}
function Origin-Plain($Members){
    foreach($p in $Members){if($p.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty){return $false}}
    return $true
}
function Get-OriginSchemaView([object]$Owner){
    $descriptor=$Owner.PSObject.Properties['Schema']
    Origin-Need ($null -ne $descriptor -and $descriptor.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty)
    $base=$Owner.Schema.PSObject.BaseObject
    $type=$script:originType.Invoke($base,$null)
    Origin-Need ($type -eq [int] -and $base -eq 1)
    $view=$Owner.Schema.PSObject
    [pscustomobject]@{Owner=$Owner;Descriptor=$descriptor;Base=$base;Type=$type.FullName;Value=$base;View=$view;Members=$view.Properties}
}
function Origin-Matches($Members,$Descriptor,$Payload,[string]$Value){
    $found=$Members[$script:originName]
    if($null -eq $found){return $false}
    return ($found.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty -and
        (Origin-Reference $found $Descriptor) -and (Origin-Reference $found.Value $Payload) -and
        $found.Value.Count -eq 1 -and $found.Value[0] -ceq $Value)
}
$d=[ordered]@{
    Schema=1;Outcome='INCOMPLETE';Stage='startup';ConstructorCalls=0
    Types=[ordered]@{A=$null;B=$null;C=$null;D=$null;E=$null}
    Values=[ordered]@{A=$null;B=$null;C=$null;D=$null;E=$null}
    PlainBeforeA=$null;PlainBeforeB=$null;SharedAB=$null;SharedAC=$null
    AInitialAttachmentRetained=$null;AMutationRetained=$null
    BHeldMatch=$null;BFreshMatch=$null;CHeldMatch=$null;CFreshMatch=$null
    FreshDPlain=$null;FreshEPlain=$null;FreshDIndependent=$null;FreshEIndependent=$null;AFinalRetained=$null
    ProductImported=$false;ProductCalled=$false;RepairPerformed=$false;ErrorCode=''
}
$exitCode=1
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    $principal=[Security.Principal.WindowsPrincipal]::new($identity)
    $admin=$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    [Console]::Out.WriteLine('CONTEXT: '+$identity.Name+' admin='+$admin)
    [Console]::Out.WriteLine('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    Origin-Need (-not $admin -and $PSVersionTable.PSVersion.ToString() -ceq '7.6.5' -and $PSVersionTable.PSEdition -ceq 'Core' -and [Environment]::Is64BitProcess)
    $d.Stage='construct-before'
    $a=New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2);$d.ConstructorCalls++
    $b=New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2);$d.ConstructorCalls++
    $av=Get-OriginSchemaView $a;$bv=Get-OriginSchemaView $b
    $d.Types.A=$av.Type;$d.Types.B=$bv.Type;$d.Values.A=$av.Value;$d.Values.B=$bv.Value
    $d.PlainBeforeA=Origin-Plain $av.Members;$d.PlainBeforeB=Origin-Plain $bv.Members
    $d.SharedAB=Origin-Reference $av.Base $bv.Base
    Origin-Need ($d.PlainBeforeA -and $d.PlainBeforeB)
    $d.Stage='attach-and-mutate'
    $payload=[object[]]@('initial')
    $av.Members.Add([Management.Automation.PSNoteProperty]::new($script:originName,$payload))
    $held=$av.Members[$script:originName]
    $d.AInitialAttachmentRetained=Origin-Matches $av.Members $held $payload 'initial'
    Origin-Need $d.AInitialAttachmentRetained
    $held.Value[0]='changed'
    $d.AMutationRetained=(Origin-Matches $av.Members $held $payload 'changed') -and
        (Origin-Matches $a.Schema.PSObject.Properties $held $payload 'changed') -and $payload[0] -ceq 'changed'
    Origin-Need $d.AMutationRetained
    $d.BHeldMatch=Origin-Matches $bv.Members $held $payload 'changed'
    $d.BFreshMatch=Origin-Matches $b.Schema.PSObject.Properties $held $payload 'changed'
    $d.Stage='construct-after'
    $c=New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2);$d.ConstructorCalls++
    $cv=Get-OriginSchemaView $c
    $d.Types.C=$cv.Type;$d.Values.C=$cv.Value
    $d.SharedAC=Origin-Reference $av.Base $cv.Base
    $d.CHeldMatch=Origin-Matches $cv.Members $held $payload 'changed'
    $d.CFreshMatch=Origin-Matches $c.Schema.PSObject.Properties $held $payload 'changed'
    $d.Stage='fresh-controls'
    $freshD=New-OriginFreshInt32 1
    $dOwner=New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2) $freshD;$d.ConstructorCalls++
    $freshE=New-OriginFreshInt32 1
    $eOwner=New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2) $freshE;$d.ConstructorCalls++
    $dv=Get-OriginSchemaView $dOwner;$ev=Get-OriginSchemaView $eOwner
    $d.Types.D=$dv.Type;$d.Types.E=$ev.Type;$d.Values.D=$dv.Value;$d.Values.E=$ev.Value
    $d.FreshDPlain=(Origin-Plain $dv.Members) -and (Origin-Plain $dOwner.Schema.PSObject.Properties)
    $d.FreshEPlain=(Origin-Plain $ev.Members) -and (Origin-Plain $eOwner.Schema.PSObject.Properties)
    $d.FreshDIndependent=(-not (Origin-Reference $dv.Base $av.Base)) -and
        (-not (Origin-Reference $dv.Base $bv.Base)) -and (-not (Origin-Reference $dv.Base $cv.Base))
    $d.FreshEIndependent=(-not (Origin-Reference $ev.Base $av.Base)) -and
        (-not (Origin-Reference $ev.Base $bv.Base)) -and (-not (Origin-Reference $ev.Base $cv.Base)) -and
        (-not (Origin-Reference $ev.Base $dv.Base))
    $d.AFinalRetained=(Origin-Matches $av.Members $held $payload 'changed') -and
        (Origin-Matches $a.Schema.PSObject.Properties $held $payload 'changed')
    Origin-Need ($d.ConstructorCalls -eq 5 -and $d.FreshDPlain -and $d.FreshEPlain -and
        $d.FreshDIndependent -and $d.FreshEIndependent -and $d.AFinalRetained)
    $d.Stage='complete'
    if($d.SharedAB -and $d.SharedAC -and $d.BHeldMatch -and $d.BFreshMatch -and $d.CHeldMatch -and $d.CFreshMatch){
        $d.Outcome='DEFAULT_REUSE_CONFIRMED';$exitCode=0
    }else{$d.Outcome='DEFAULT_REUSE_NOT_CONFIRMED';$exitCode=2}
}catch{
    $d.Outcome='INCOMPLETE'
    $d.ErrorCode=if($_.Exception.Message -ceq 'fixture-invalid'){'fixture-invalid'}else{'exception'}
    $exitCode=1
}
$json=$d|ConvertTo-Json -Depth 5 -Compress
if([Text.UTF8Encoding]::new($false).GetByteCount($json) -gt 65536){
    [Console]::Out.WriteLine('ORIGIN: {"Schema":1,"Outcome":"INCOMPLETE","ErrorCode":"record-too-large"}')
    [Console]::Out.WriteLine('LIMIT: fixture only; no product import, call or repair')
    exit 1
}
[Console]::Out.WriteLine('ORIGIN: '+$json)
[Console]::Out.WriteLine('LIMIT: fixture only; no product import, call or repair')
exit $exitCode
