param([ValidateSet('all','phase-rounding','deadline-equality')][string]$Group='all')
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:case='ApiRequired';$script:callback=0;$script:metadataSequence=0
$script:typeMethod=[object].GetMethod('GetType',[type[]]@())
$script:invokeMethod=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
$script:referenceMethod=[object].GetMethod('ReferenceEquals',[type[]]@([object],[object]))
function TB-Type($Value){if($null -eq $Value){return $null};$arguments=[object[]]@($Value,$null);$script:invokeMethod.Invoke($script:typeMethod,$arguments)}
function TB-Ref($Left,$Right){$arguments=[object[]]@($Left,$Right);[bool]$script:referenceMethod.Invoke($null,$arguments)}
function TB-Need([bool]$Value){if(-not $Value){throw 'Assertion'}}
function TB-Case([string]$Name,[scriptblock]$Body){
    if($Group -cne 'all' -and $Name -cne 'ApiRequired' -and
       -not ($Group -ceq 'phase-rounding' -and $Name -ceq 'FractionalPhaseRounding') -and
       -not ($Group -ceq 'deadline-equality' -and $Name -ceq 'DeadlineEquality')){return}
    $script:case=$Name;& $Body
    [Console]::Out.WriteLine('PASS: P2A'+$Name)
}
function TB-Equal([object]$Actual,[object]$Expected,[int]$Depth=0){
    if($Depth -gt 3){return $false}
    if($null -eq $Expected){return ($null -eq $Actual)}
    if((TB-Type $Actual) -ne (TB-Type $Expected)){return $false}
    if((TB-Type $Expected) -eq [Management.Automation.PSCustomObject]){
        $props=@($Actual.PSObject.Properties);$want=@($Expected.PSObject.Properties)
        if($props.Count -ne $want.Count){return $false}
        foreach($p in $want){
            $found=@($props|Where-Object{[StringComparer]::Ordinal.Equals($_.Name,$p.Name)})
            if($found.Count -ne 1 -or $found[0].MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){return $false}
            if(-not (TB-Equal $found[0].Value $p.Value ($Depth+1))){return $false}
        };return $true
    }
    if((TB-Type $Expected) -eq [string]){return [StringComparer]::Ordinal.Equals($Actual.PSObject.BaseObject,$Expected.PSObject.BaseObject)}
    return ($Actual.PSObject.BaseObject -ceq $Expected.PSObject.BaseObject)
}
function New-TBRequest([object]$OriginTick=[long]0,[object]$Frequency=[long]1000,[object]$ObservedTick=[long]0,[object]$StopTick=$null,[object]$Schema=[int]1){
    [pscustomobject]@{Schema=$Schema;OriginTick=$OriginTick;Frequency=$Frequency;ObservedTick=$ObservedTick;StopTick=$StopTick}
}
# Test-only literal builder: every numeric argument below is independently frozen.
function TB-Want([long]$Origin,[long]$Frequency,[long]$Deadline,[long]$Stop,$Natural,$Cleanup,$Publish,[bool]$Expired){
    [pscustomobject]@{Schema=[int]1;Kind='timebase-plan-only';Created=$true;Reason='plan-created';Plan=[pscustomobject]@{
        OriginTick=$Origin;Frequency=$Frequency;DeadlineTick=$Deadline;StopAtTick=$Stop
        NaturalUntilTick=$(if($null -eq $Natural){$null}else{[long]$Natural})
        CleanupUntilTick=$(if($null -eq $Cleanup){$null}else{[long]$Cleanup})
        PublishUntilTick=$(if($null -eq $Publish){$null}else{[long]$Publish})
        ObservedDeadlineExpired=$Expired
    };SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false}
}
function TB-Normal {TB-Want 0 1000 120000 113500 502 5502 6502 $false}
function TB-Refusal([string]$Reason){[pscustomobject]@{Schema=[int]1;Kind='timebase-plan-only';Created=$false;Reason=$Reason;Plan=$null;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false}}
function TB-Reject($Request,[string]$Reason='invalid-input'){TB-Need (TB-Equal (New-OfflineTimebasePlan $Request) (TB-Refusal $Reason))}
function TB-Leaf($Owner,[string]$Field){
    $descriptor=$Owner.PSObject.Properties[$Field]
    TB-Need ($null -ne $descriptor -and $null -ne $descriptor.Value)
    $stored=TB-Type $descriptor.Value;$base=$Owner.$Field.PSObject.BaseObject
    $view=$Owner.$Field.PSObject;$members=$view.Properties
    foreach($p in $members){TB-Need ($p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty)}
    [pscustomobject]@{Owner=$Owner;Field=$Field;Descriptor=$descriptor;Stored=$stored;Base=$base;Type=(TB-Type $base);View=$view;Members=$members}
}
function TB-Leaves($Value){
    foreach($p in $Value.PSObject.Properties){
        if($null -eq $p.Value){continue}
        if((TB-Type $p.Value) -eq [Management.Automation.PSCustomObject]){TB-Leaves $p.Value}else{TB-Leaf $Value $p.Name}
    }
}
function TB-Mutate($Selected,$Others){
    $script:metadataSequence++;$name='P2AMetadata'+$script:metadataSequence
    foreach($other in $Others){TB-Need (-not (TB-Ref $Selected.Base $other.Base))}
    $payload=[object[]]@('initial')
    $Selected.Members.Add([Management.Automation.PSNoteProperty]::new($name,$payload))
    $held=$Selected.Members[$name]
    TB-Need ($null -ne $held -and $held.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty -and (TB-Ref $held.Value $payload))
    $read=$Selected.Owner.($Selected.Field).PSObject.Properties[$name];$retained=$null -ne $read
    if($Selected.Stored -ne [string]){TB-Need $retained}
    $held.Value[0]='changed'
    TB-Need ((TB-Ref $Selected.View.Properties[$name] $held) -and $held.Value[0] -ceq 'changed' -and $payload[0] -ceq 'changed')
    $after=$Selected.Owner.($Selected.Field).PSObject.Properties[$name]
    TB-Need (($null -ne $after) -eq $retained)
    if($retained){TB-Need ((TB-Ref $after.Value $payload) -and $after.Value[0] -ceq 'changed')}
    TB-Need ((TB-Ref $Selected.Owner.PSObject.Properties[$Selected.Field] $Selected.Descriptor) -and
        (TB-Type $Selected.Descriptor.Value) -eq $Selected.Stored -and
        (TB-Type $Selected.Owner.($Selected.Field).PSObject.BaseObject) -eq $Selected.Type -and
        (TB-Ref $Selected.Owner.($Selected.Field).PSObject.BaseObject $Selected.Base))
    foreach($other in $Others){TB-Need ($null -eq $other.Members[$name] -and $null -eq $other.Owner.($other.Field).PSObject.Properties[$name])}
}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    $principal=[Security.Principal.WindowsPrincipal]::new($identity)
    $admin=$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    [Console]::Out.WriteLine('CONTEXT: '+$identity.Name+' admin='+$admin)
    [Console]::Out.WriteLine('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    TB-Need (-not $admin -and $PSVersionTable.PSVersion.ToString() -ceq '7.6.5' -and $PSVersionTable.PSEdition -ceq 'Core' -and [Environment]::Is64BitProcess)
    $module=Join-Path $PSScriptRoot 'offline-timebase-plan.ps1'
    if(Test-Path -LiteralPath $module){. $module}
    TB-Case ApiRequired {TB-Need ($null -ne (Get-Command New-OfflineTimebasePlan -CommandType Function -ErrorAction SilentlyContinue))}
    TB-Case NullStop {TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest)) (TB-Want 0 1000 120000 113500 $null $null $null $false))}
    TB-Case NormalStop {TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2))) (TB-Normal))}
    TB-Case NonzeroOrigin {TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]1000) ([long]1000) ([long]1004) ([long]1002))) (TB-Want 1000 1000 121000 114500 1502 6502 7502 $false))}
    TB-Case FractionalFrequency {
        TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]10) ([long]1001) ([long]120130))) (TB-Want 10 1001 120130 113623 $null $null $null $true))
        TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]0) ([long]9007199254740993))) (TB-Want 0 9007199254740993 1080863910568919160 1022317115413102705 $null $null $null $false))
    }
    TB-Case FractionalPhaseRounding {TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]10) ([long]1001) ([long]14) ([long]12))) (TB-Want 10 1001 120130 113623 512 5517 6518 $false))}
    TB-Case DeadlineEquality {
        foreach($v in @(@([long]119999,$false),@([long]120000,$true),@([long]120001,$true))){TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]0) ([long]1000) $v[0])) (TB-Want 0 1000 120000 113500 $null $null $null $v[1]))}
    }
    TB-Case StopCutoff {
        foreach($n in @([long]113499,[long]113500,[long]113501)){TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]0) ([long]1000) $n)) (TB-Want 0 1000 120000 113500 $null $null $null $false))}
        foreach($n in @([long]113612,[long]113613,[long]113614)){TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]0) ([long]1001) $n)) (TB-Want 0 1001 120120 113613 $null $null $null $false))}
    }
    TB-Case StageWindows {foreach($n in @([long]501,[long]502,[long]5501,[long]5502,[long]6501,[long]6502)){TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]0) ([long]1000) $n ([long]2))) (TB-Normal))}}
    TB-Case LateStop {foreach($n in @([long]120000,[long]120001)){TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]0) ([long]1000) $n $n)) (TB-Want 0 1000 120000 113500 114000 119000 120000 $true))}}
    TB-Case OriginLimit {
        TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]9223372036854655807) ([long]1000) ([long]::MaxValue))) (TB-Want 9223372036854655807 1000 9223372036854775807 9223372036854769307 $null $null $null $true))
        TB-Reject (New-TBRequest ([long]9223372036854655808) ([long]1000) ([long]::MaxValue)) range-refused
    }
    TB-Case FrequencyLimit {
        TB-Reject (New-TBRequest ([long]0) ([long]999)) range-refused
        TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest)) (TB-Want 0 1000 120000 113500 $null $null $null $false))
        TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]0) ([long]76861433640456465))) (TB-Want 0 76861433640456465 9223372036854775800 8723772718191808777 $null $null $null $false))
        TB-Reject (New-TBRequest ([long]0) ([long]76861433640456466)) range-refused
        TB-Reject (New-TBRequest ([long]0) ([long]::MaxValue)) range-refused
    }
    TB-Case IntermediateOverflow {TB-Reject (New-TBRequest ([long]9223372036854655807) ([long]1000) ([long]::MaxValue) ([long]::MaxValue)) range-refused}
    TB-Case RequestShape {
        TB-Reject $null;TB-Reject @{};TB-Reject ([object]::new())
        $r=New-TBRequest;$r.PSObject.Properties.Remove('StopTick');TB-Reject $r
        $r=New-TBRequest;$r.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Extra',1));TB-Reject $r
        $r=New-TBRequest;$r.PSObject.Properties.Remove('Schema');$r.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('schema',[int]1));TB-Reject $r
    }
    TB-Case ExactTypes {
        foreach($v in @([long]1,[int]2,'1',$null,$true,[double]1)){TB-Reject (New-TBRequest -Schema $v)}
        foreach($field in @('OriginTick','Frequency','ObservedTick','StopTick')){
            foreach($v in @([int]0,[double]0,'0',$true,$null)){
                if($field -ceq 'StopTick' -and $null -eq $v){continue}
                $r=New-TBRequest;$r.PSObject.Properties[$field].Value=$v;TB-Reject $r
            }
        }
    }
    TB-Case PassiveDescriptors {
        $r=New-TBRequest;$r.PSObject.Properties.Remove('ObservedTick')
        $r.PSObject.Properties.Add([Management.Automation.PSScriptProperty]::new('ObservedTick',{$script:callback++;[long]0}))
        TB-Need ($r.ObservedTick -eq 0 -and $script:callback -eq 1);$script:callback=0;TB-Reject $r;TB-Need ($script:callback -eq 0)
        $r=New-TBRequest;$r.PSObject.Properties.Remove('ObservedTick');$r.PSObject.Properties.Add([Management.Automation.PSAliasProperty]::new('ObservedTick','OriginTick'))
        TB-Need ($r.ObservedTick -eq 0 -and $r.PSObject.Properties['ObservedTick'].MemberType -eq [Management.Automation.PSMemberTypes]::AliasProperty);TB-Reject $r
        foreach($name in @('Probe','GetType','ToString')){
            $r=New-TBRequest;$r.PSObject.Methods.Add([Management.Automation.PSScriptMethod]::new($name,{$script:callback++;'observable'}))
            TB-Need ($r.$name() -ceq 'observable' -and $script:callback -eq 1);$script:callback=0;TB-Reject $r;TB-Need ($script:callback -eq 0)
        }
    }
    TB-Case DecoratedScalars {
        foreach($field in @('Schema','OriginTick','Frequency','ObservedTick','StopTick')){
            $r=New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2)
            $leaf=TB-Leaf $r $field;TB-Mutate $leaf @();TB-Reject $r
        }
    }
    TB-Case RangeRelationships {
        foreach($r in @((New-TBRequest ([long]-1)),(New-TBRequest ([long]0) ([long]1000) ([long]-1)),(New-TBRequest ([long]0) ([long]1000) ([long]0) ([long]-1)),(New-TBRequest ([long]10) ([long]1000) ([long]9)),(New-TBRequest ([long]10) ([long]1000) ([long]12) ([long]9)),(New-TBRequest ([long]0) ([long]1000) ([long]1) ([long]2)))){TB-Reject $r range-refused}
    }
    TB-Case InputMutation {
        $r=New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2);$a=New-OfflineTimebasePlan $r;$b=New-OfflineTimebasePlan $r
        $inputs=@(TB-Leaves $r);$outputs=@((TB-Leaves $a);(TB-Leaves $b))
        foreach($leaf in $inputs){TB-Mutate $leaf $outputs}
        $payload=[object[]]@('initial');$r.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('RootMutation',$payload));$payload[0]='changed'
        TB-Need ((TB-Ref $r.RootMutation $payload) -and $r.RootMutation[0] -ceq 'changed')
        $r.PSObject.Properties['ObservedTick'].Value=[long]5
        TB-Need ($r.ObservedTick -eq 5 -and (TB-Equal $a (TB-Normal)) -and (TB-Equal $b (TB-Normal)))
        TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2))) (TB-Normal))
    }
    TB-Case OutputOwnership {
        $r=New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2);$a=New-OfflineTimebasePlan $r;$b=New-OfflineTimebasePlan $r
        $input=@(TB-Leaves $r);$la=@(TB-Leaves $a);$lb=@(TB-Leaves $b)
        TB-Need (-not (TB-Ref $a $b) -and -not (TB-Ref $a.Plan $b.Plan) -and $la.Count -eq 15 -and $lb.Count -eq 15)
        foreach($leaf in $la){$others=@($input+$lb+@($la|Where-Object{-not (TB-Ref $_ $leaf)}));TB-Mutate $leaf $others}
        TB-Need ((TB-Equal $a (TB-Normal)) -and (TB-Equal $b (TB-Normal)))
    }
    TB-Case SequentialIsolation {
        $r=New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2);$a=New-OfflineTimebasePlan $r;$bad=New-OfflineTimebasePlan $null;$b=New-OfflineTimebasePlan $r
        TB-Need ((TB-Equal $a (TB-Normal)) -and (TB-Equal $bad (TB-Refusal invalid-input)) -and (TB-Equal $b (TB-Normal)))
        $la=@(TB-Leaves $a);$lb=@(TB-Leaves $b);$lf=@(TB-Leaves $bad);$li=@(TB-Leaves $r)
        foreach($set in @(@{Selected=$lb;Other=@($la+$lf+$li)},@{Selected=$la;Other=@($lb+$lf+$li)},@{Selected=$lf;Other=@($la+$lb+$li)})){
            foreach($leaf in $set.Selected){TB-Mutate $leaf @($set.Other+@($set.Selected|Where-Object{-not (TB-Ref $_ $leaf)}))}
        }
        TB-Need ((TB-Equal $a (TB-Normal)) -and (TB-Equal $bad (TB-Refusal invalid-input)) -and (TB-Equal $b (TB-Normal)))
        TB-Need (TB-Equal (New-OfflineTimebasePlan $r) (TB-Normal))
    }
    TB-Case RefusalShape {TB-Reject $null invalid-input;TB-Reject (New-TBRequest ([long]0) ([long]999)) range-refused}
    TB-Case ConservativeFlags {
        foreach($r in @((New-OfflineTimebasePlan $null),(New-OfflineTimebasePlan (New-TBRequest)),(New-OfflineTimebasePlan (New-TBRequest ([long]0) ([long]999))))){
            TB-Need ($r.Kind -ceq 'timebase-plan-only' -and @($r.PSObject.Properties).Count -eq 8)
            foreach($f in @('SourceVerified','LiveActionsPerformed','RuntimeApproved')){TB-Need ((TB-Type $r.$f) -eq [bool] -and $r.$f -eq $false)}
        }
    }
    TB-Case Complete {TB-Need ($script:callback -eq 0)}
    [Console]::Out.WriteLine('LIMIT: declared integers only; no real clock or runtime evidence')
    exit 0
}catch{[Console]::Out.WriteLine('ASSERTION FAILED: P2A'+$script:case);exit 1}
