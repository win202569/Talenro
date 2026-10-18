param([ValidateSet('all','phase-rounding','deadline-equality')][string]$Group='all')
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:case='ApiRequired';$script:callback=0;$script:metadataSequence=0
$script:diagStage='startup';$script:diagOriginalLine=0;$script:diagField='';$script:diagOtherField='';$script:diagOtherIndex=0;$script:diagAssertion=0
# Failure-only passive inspection. No observation of caller objects before failure.
$script:diagState=$null
function TB-FailureState {
    $state=[ordered]@{Frames=@();CaptureError=$null}
    try{
        for($scopeNumber=1;$scopeNumber -le 3;$scopeNumber++){
            $variables=@(Get-Variable -Name Owner,Field,Selected,other,name,r,a,b -Scope $scopeNumber -ErrorAction SilentlyContinue)
            foreach($variable in $variables){
                $value=$variable.Value
                if($null -eq $value){continue}
                if($variable.Name -cin @('Field','name')){
                    $state.Frames+=,[pscustomobject]@{Scope=$scopeNumber;Variable=$variable.Name;Text=[string]$value}
                    continue
                }
                if($variable.Name -cin @('Selected','other')){$value=$value.Owner}
                $properties=@()
                foreach($property in $value.PSObject.Properties){
                    if($property.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){continue}
                    $v=$property.Value;$typeName=$null;$scalar=$null;$notes=@()
                    if($null -ne $v){
                        $t=$script:typeMethod.Invoke($v.PSObject.BaseObject,$null);$typeName=$t.FullName
                        if($t -eq [int] -or $t -eq [long] -or $t -eq [bool] -or $t -eq [string]){$scalar=$v.PSObject.BaseObject}
                        foreach($member in $v.PSObject.Properties){
                            if($member.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty){$notes+=,$member.Name}
                        }
                    }
                    $properties+=,[pscustomobject]@{Field=$property.Name;Type=$typeName;Scalar=$scalar;NoteProperties=$notes}
                }
                $state.Frames+=,[pscustomobject]@{Scope=$scopeNumber;Variable=$variable.Name;Properties=$properties}
            }
        }
    }catch{$state.CaptureError=$_.Exception.Message}
    $script:diagState=[pscustomobject]$state
}
function TB-Diagnostic([string]$Outcome,$Record=$null){
    $d=[ordered]@{
        Schema=1;Outcome=$Outcome;Case=$script:case;Stage=$script:diagStage
        Field=$script:diagField;OtherField=$script:diagOtherField;OtherIndex=$script:diagOtherIndex
        AssertionOrdinal=$script:diagAssertion;OriginalOperationLine=$script:diagOriginalLine
        ExceptionType=$null;ErrorId=$null;Message=$null;DiagnosticLine=$null;Stack=$null
        State=$script:diagState;ProductChanged=$false;RepairPerformed=$false;Acceptance=$false
    }
    if($null -ne $Record){
        $d.ExceptionType=$Record.Exception.GetType().FullName
        $d.ErrorId=$Record.FullyQualifiedErrorId;$d.Message=$Record.Exception.Message
        $d.DiagnosticLine=$Record.InvocationInfo.ScriptLineNumber;$d.Stack=$Record.ScriptStackTrace
    }
    $json=$d|ConvertTo-Json -Depth 12 -Compress
    if([Text.UTF8Encoding]::new($false).GetByteCount($json) -gt 65536){
        $json='{"Schema":1,"Outcome":"INCOMPLETE","Detail":"diagnostic-record-too-large"}'
    }
    [Console]::Out.WriteLine('DIAG: '+$json)
}
$script:typeMethod=[object].GetMethod('GetType',[type[]]@())
$script:invokeMethod=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
$script:referenceMethod=[object].GetMethod('ReferenceEquals',[type[]]@([object],[object]))
function TB-Type($Value){if($null -eq $Value){return $null};$arguments=[object[]]@($Value,$null);$script:invokeMethod.Invoke($script:typeMethod,$arguments)}
function TB-Ref($Left,$Right){$arguments=[object[]]@($Left,$Right);[bool]$script:referenceMethod.Invoke($null,$arguments)}
function TB-Need([bool]$Value){$script:diagAssertion++;if(-not $Value){TB-FailureState;throw 'Assertion'}}
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
. (Join-Path $PSScriptRoot 'fixture-builder.ps1')
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
    $script:diagStage='leaf-descriptor';$script:diagOriginalLine=49;$script:diagField=$Field;$script:diagOtherField='';$script:diagOtherIndex=0
    $descriptor=$Owner.PSObject.Properties[$Field]
    TB-Need ($null -ne $descriptor -and $null -ne $descriptor.Value)
    $script:diagStage='leaf-context';$script:diagOriginalLine=51
    $stored=TB-Type $descriptor.Value;$base=$Owner.$Field.PSObject.BaseObject
    $view=$Owner.$Field.PSObject;$members=$view.Properties
    $script:diagStage='leaf-initially-plain';$script:diagOriginalLine=53
    foreach($p in $members){TB-Need ($p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty)}
    [pscustomobject]@{Owner=$Owner;Field=$Field;Descriptor=$descriptor;Stored=$stored;Base=$base;Type=(TB-Type $base);View=$view;Members=$members}
}
function TB-Leaves($Value){
    $script:diagStage='enumerate-leaves';$script:diagOriginalLine=57
    foreach($p in $Value.PSObject.Properties){
        if($null -eq $p.Value){continue}
        if((TB-Type $p.Value) -eq [Management.Automation.PSCustomObject]){TB-Leaves $p.Value}else{TB-Leaf $Value $p.Name}
    }
}
function TB-Mutate($Selected,$Others){
    $script:diagStage='mutate-begin';$script:diagOriginalLine=63;$script:diagField=$Selected.Field;$script:diagOtherField='';$script:diagOtherIndex=0
    $script:metadataSequence++;$name='P2AMetadata'+$script:metadataSequence
    $script:diagOtherIndex=0
    foreach($other in $Others){
    $script:diagStage='cross-owner-independence';$script:diagOriginalLine=64;$script:diagOtherField=$other.Field;$script:diagOtherIndex++
        TB-Need (-not (TB-Ref $Selected.Base $other.Base))
    }
    $payload=[object[]]@('initial')
    $script:diagStage='attach-selected';$script:diagOriginalLine=66
    $Selected.Members.Add([Management.Automation.PSNoteProperty]::new($name,$payload))
    $held=$Selected.Members[$name]
    $script:diagStage='held-descriptor';$script:diagOriginalLine=68
    TB-Need ($null -ne $held -and $held.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty -and (TB-Ref $held.Value $payload))
    $script:diagStage='fresh-reread';$script:diagOriginalLine=69
    $read=$Selected.Owner.($Selected.Field).PSObject.Properties[$name];$retained=$null -ne $read
    if($Selected.Stored -ne [string]){TB-Need $retained}
    $script:diagStage='mutate-payload';$script:diagOriginalLine=71
    $held.Value[0]='changed'
    $script:diagStage='held-mutation';$script:diagOriginalLine=72
    TB-Need ((TB-Ref $Selected.View.Properties[$name] $held) -and $held.Value[0] -ceq 'changed' -and $payload[0] -ceq 'changed')
    $script:diagStage='reread-after';$script:diagOriginalLine=73
    $after=$Selected.Owner.($Selected.Field).PSObject.Properties[$name]
    $script:diagStage='reread-consistency';$script:diagOriginalLine=74
    TB-Need (($null -ne $after) -eq $retained)
    $script:diagStage='retained-payload';$script:diagOriginalLine=75
    if($retained){TB-Need ((TB-Ref $after.Value $payload) -and $after.Value[0] -ceq 'changed')}
    $script:diagStage='selected-invariant';$script:diagOriginalLine=76
    TB-Need ((TB-Ref $Selected.Owner.PSObject.Properties[$Selected.Field] $Selected.Descriptor) -and
        (TB-Type $Selected.Descriptor.Value) -eq $Selected.Stored -and
        (TB-Type $Selected.Owner.($Selected.Field).PSObject.BaseObject) -eq $Selected.Type -and
        (TB-Ref $Selected.Owner.($Selected.Field).PSObject.BaseObject $Selected.Base))
    $script:diagOtherIndex=0
    foreach($other in $Others){
    $script:diagStage='cross-owner-clean';$script:diagOriginalLine=80;$script:diagOtherField=$other.Field;$script:diagOtherIndex++
        TB-Need ($null -eq $other.Members[$name] -and $null -eq $other.Owner.($other.Field).PSObject.Properties[$name])
    }
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
    $script:diagStage='construct-request';$script:diagOriginalLine=157;$script:diagField='';$script:diagOtherField='';$script:diagOtherIndex=0
        $r=New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2)
    $script:diagStage='construct-result-a';$script:diagOriginalLine=157;$script:diagField='';$script:diagOtherField='';$script:diagOtherIndex=0
        $a=New-OfflineTimebasePlan $r
    $script:diagStage='construct-result-b';$script:diagOriginalLine=157;$script:diagField='';$script:diagOtherField='';$script:diagOtherIndex=0
        $b=New-OfflineTimebasePlan $r
    $script:diagStage='collect-input-leaves';$script:diagOriginalLine=158;$script:diagField='';$script:diagOtherField='';$script:diagOtherIndex=0
        $inputs=@(TB-Leaves $r)
    $script:diagStage='collect-output-leaves';$script:diagOriginalLine=158;$script:diagField='';$script:diagOtherField='';$script:diagOtherIndex=0
        $outputs=@((TB-Leaves $a);(TB-Leaves $b))
    $script:diagStage='input-mutation-loop';$script:diagOriginalLine=159
        foreach($leaf in $inputs){TB-Mutate $leaf $outputs}
    $script:diagStage='create-root-payload';$script:diagOriginalLine=160;$script:diagField='';$script:diagOtherField='';$script:diagOtherIndex=0
        $payload=[object[]]@('initial')
    $script:diagStage='attach-root-payload';$script:diagOriginalLine=160;$script:diagField='';$script:diagOtherField='';$script:diagOtherIndex=0
        $r.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('RootMutation',$payload))
    $script:diagStage='mutate-root-payload';$script:diagOriginalLine=160;$script:diagField='';$script:diagOtherField='';$script:diagOtherIndex=0
        $payload[0]='changed'
    $script:diagStage='root-retention';$script:diagOriginalLine=161
        TB-Need ((TB-Ref $r.RootMutation $payload) -and $r.RootMutation[0] -ceq 'changed')
    $script:diagStage='replace-input-value';$script:diagOriginalLine=162
        $r.PSObject.Properties['ObservedTick'].Value=[long]5
    $script:diagStage='saved-results-compare';$script:diagOriginalLine=163
        TB-Need ($r.ObservedTick -eq 5 -and (TB-Equal $a (TB-Normal)) -and (TB-Equal $b (TB-Normal)))
    $script:diagStage='next-normal-call';$script:diagOriginalLine=164
        TB-Need (TB-Equal (New-OfflineTimebasePlan (New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2))) (TB-Normal))
    }
    TB-Diagnostic 'NOT_REPRODUCED'
    exit 2
}catch{
    $record=$_
    $outcome=if($script:case -ceq 'InputMutation'){'INPUT_MUTATION_EXCEPTION'}else{'EARLIER_FAILURE'}
    TB-Diagnostic $outcome $record
    [Console]::Out.WriteLine('ASSERTION FAILED: P2A'+$script:case)
    exit 1
}
