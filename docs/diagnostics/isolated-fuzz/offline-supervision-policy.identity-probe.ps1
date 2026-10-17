# Diagnostic spike only: owned Int32 fixtures; no product/test imports or repairs.
param([Parameter(Mandatory=$true)][ValidateSet('independent','shared')][string]$Mode)
$ErrorActionPreference='Stop'
function IdentityProbe-Type($Value){[object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)}
function IdentityProbe-WrapPrimitive([object]$Value){
    if($null -eq $Value){throw 'FIXTURE ERROR: null is not a wrapped primitive'}
    $ctor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
    if($null -eq $ctor -or -not $ctor.IsPublic){throw 'FIXTURE ERROR: object constructor unavailable'}
    $parameters=@($ctor.GetParameters())
    if($parameters.Count -ne 1 -or $parameters[0].ParameterType -ne [object]){throw 'FIXTURE ERROR: exact object signature required'}
    return ,($ctor.Invoke([object[]]@($Value)))
}
function IdentityProbe-Decorate($Owner,[string]$Name){
    $original=$Owner.$Name;$type=IdentityProbe-Type $original
    $wrapped=IdentityProbe-WrapPrimitive $original
    $metadata=[object[]]@('initial')
    $wrapped.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Metadata',$metadata))
    $Owner.$Name=$wrapped;$field=$Owner.$Name;$property=$field.PSObject.Properties['Metadata']
    if((IdentityProbe-Type $field) -ne $type -or $field.PSObject.BaseObject -cne $original.PSObject.BaseObject -or $null -eq $property -or $property.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or -not [object]::ReferenceEquals($property.Value,$metadata)){throw 'FIXTURE ERROR: decorated scalar was not preserved'}
    return ,$metadata
}
function IdentityProbe-Snapshot([string]$Stage,$A,$B,[object]$AM,[object]$BM){
    $ap=$A.Value.PSObject.Properties['Metadata'];$bp=$B.Value.PSObject.Properties['Metadata']
    $an=$null -ne $ap -and $ap.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty
    $bn=$null -ne $bp -and $bp.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty
    [pscustomobject]@{
        Stage=$Stage;Reached=$true;BaseObjectSame=[object]::ReferenceEquals($A.Value.PSObject.BaseObject,$B.Value.PSObject.BaseObject)
        AInt32Zero=((IdentityProbe-Type $A.Value) -eq [int] -and $A.Value.PSObject.BaseObject -ceq 0)
        BInt32Zero=((IdentityProbe-Type $B.Value) -eq [int] -and $B.Value.PSObject.BaseObject -ceq 0)
        ADescriptorPresent=($null -ne $ap);ADescriptorNote=$an;BDescriptorPresent=($null -ne $bp);BDescriptorNote=$bn
        ARefA=$(if($an){[object]::ReferenceEquals($ap.Value,$AM)}else{$null})
        ARefB=$(if($an -and $null -ne $BM){[object]::ReferenceEquals($ap.Value,$BM)}else{$null})
        BRefA=$(if($bn){[object]::ReferenceEquals($bp.Value,$AM)}else{$null})
        BRefB=$(if($bn -and $null -ne $BM){[object]::ReferenceEquals($bp.Value,$BM)}else{$null})
        ArraysDistinct=$(if($null -ne $BM){-not [object]::ReferenceEquals($AM,$BM)}else{$null})
        AMContent=[string]$AM[0];BMContent=$(if($null -ne $BM){[string]$BM[0]}else{$null})
    }
}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
    Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
    Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    # One fixed strategy: one retrieval per separate typed array, or one shared retrieval.
    $sourceA=[Array]::CreateInstance([int],1);$valueA=$sourceA.GetValue(0)
    if($Mode -ceq 'independent'){$sourceB=[Array]::CreateInstance([int],1);$valueB=$sourceB.GetValue(0)}else{$valueB=$valueA}
    $a=[pscustomobject]@{Value=$valueA};$b=[pscustomobject]@{Value=$valueB}
    $same=[object]::ReferenceEquals($a.Value.PSObject.BaseObject,$b.Value.PSObject.BaseObject)
    $pre=[pscustomobject]@{AType=(IdentityProbe-Type $a.Value).FullName;BType=(IdentityProbe-Type $b.Value).FullName;AZero=($a.Value.PSObject.BaseObject -ceq 0);BZero=($b.Value.PSObject.BaseObject -ceq 0);BaseObjectSame=$same;Held=$false}
    $pre.Held=$pre.AType -ceq 'System.Int32' -and $pre.BType -ceq 'System.Int32' -and $pre.AZero -and $pre.BZero -and ($same -eq ($Mode -ceq 'shared'))
    $stages=@(foreach($name in @('AfterA','AfterB','AfterAMutation')){
        [pscustomobject]@{Stage=$name;Reached=$false;BaseObjectSame=$null;AInt32Zero=$null;BInt32Zero=$null;ADescriptorPresent=$null;ADescriptorNote=$null;BDescriptorPresent=$null;BDescriptorNote=$null;ARefA=$null;ARefB=$null;BRefA=$null;BRefB=$null;ArraysDistinct=$null;AMContent=$null;BMContent=$null}
    })
    $failure=if($pre.Held){$null}else{'IdentityPrecondition'}
    if($pre.Held){
        try{
            $am=IdentityProbe-Decorate $a Value
            $stages[0]=IdentityProbe-Snapshot AfterA $a $b $am $null
            $bm=IdentityProbe-Decorate $b Value
            $stages[1]=IdentityProbe-Snapshot AfterB $a $b $am $bm
            $am[0]='changed'
            $stages[2]=IdentityProbe-Snapshot AfterAMutation $a $b $am $bm
        }catch{if($_.Exception.Message -cne 'FIXTURE ERROR: decorated scalar was not preserved'){throw};$failure='HelperFidelity'}
    }
    $result=[pscustomobject]@{Schema=1;Mode=$Mode;Precondition=$pre;Stages=$stages;Failure=$failure;ProductImported=$false;OriginalTestsRun=$false;WarmupReplayed=$false;FixtureChanged=$false}
    Write-Output ('DIAGNOSTIC: '+($result|ConvertTo-Json -Depth 7 -Compress))
    Write-Output 'LIMIT: identity comparison only; no original tests, warmups, product calls, repairs or budget renewal.'
}catch{Write-Error $_;exit 1}
