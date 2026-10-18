Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:fixtureCase='SchemaIndependent';$script:fixtureCallbacks=0
$script:fixtureType=[object].GetMethod('GetType',[type[]]@())
$script:fixtureInvoke=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
$script:fixtureReference=[object].GetMethod('ReferenceEquals',[type[]]@([object],[object]))
function F-Type($Value){if($null -eq $Value){return $null};$argsLocal=[object[]]@($Value.PSObject.BaseObject,$null);$script:fixtureInvoke.Invoke($script:fixtureType,$argsLocal)}
function F-Ref($Left,$Right){$argsLocal=[object[]]@($Left,$Right);[bool]$script:fixtureReference.Invoke($null,$argsLocal)}
function F-Need([bool]$Value){if(-not $Value){throw 'Assertion'}}
function F-Case([string]$Name,[scriptblock]$Body){$script:fixtureCase=$Name;& $Body;[Console]::Out.WriteLine('PASS: P2AFixture'+$Name)}
function F-Normal {New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2)}
function F-Leaf($Owner,[string]$Field){
    $descriptor=$Owner.PSObject.Properties[$Field]
    F-Need ($null -ne $descriptor -and $descriptor.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty)
    $base=$Owner.$Field.PSObject.BaseObject;$view=$Owner.$Field.PSObject
    [pscustomobject]@{Owner=$Owner;Field=$Field;Descriptor=$descriptor;Base=$base;Members=$view.Properties;View=$view}
}
function F-Plain($Members){foreach($p in $Members){if($p.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty){return $false}};return $true}
function F-Expected {
    [pscustomobject]@{Schema=[int]1;OriginTick=[long]0;Frequency=[long]1000;ObservedTick=[long]4;StopTick=[long]2}
}
function F-Equal($Actual,$Expected){
    F-Need ((F-Type $Actual) -eq [Management.Automation.PSCustomObject] -and @($Actual.PSObject.Properties).Count -eq 5)
    foreach($p in $Expected.PSObject.Properties){
        $got=$Actual.PSObject.Properties[$p.Name]
        F-Need ($null -ne $got -and $got.Name -ceq $p.Name -and $got.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty)
        F-Need ((F-Type $got.Value) -eq (F-Type $p.Value))
        if($null -eq $p.Value){F-Need ($null -eq $got.Value)}else{F-Need ($got.Value.PSObject.BaseObject -ceq $p.Value.PSObject.BaseObject)}
    }
}
function F-Has($Members,[string]$Name,$Descriptor,$Payload){
    $p=$Members[$Name]
    return ($null -ne $p -and $p.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty -and
        (F-Ref $p $Descriptor) -and (F-Ref $p.Value $Payload) -and $p.Value[0] -ceq 'changed')
}
function F-Fresh([type]$Type,$Value){$storage=[Array]::CreateInstance($Type,1);$storage.SetValue($Value,0);return ,($storage.GetValue(0))}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    $principal=[Security.Principal.WindowsPrincipal]::new($identity)
    $admin=$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    [Console]::Out.WriteLine('CONTEXT: '+$identity.Name+' admin='+$admin)
    [Console]::Out.WriteLine('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    F-Need (-not $admin -and $PSVersionTable.PSVersion.ToString() -ceq '7.6.5' -and $PSVersionTable.PSEdition -ceq 'Core' -and [Environment]::Is64BitProcess)
    . (Join-Path $PSScriptRoot 'fixture-builder.ps1')
    F-Case SchemaIndependent {
        $a=F-Normal;$b=F-Normal
        $la=F-Leaf $a Schema;$lb=F-Leaf $b Schema
        F-Need (-not (F-Ref $la.Base $lb.Base))
        F-Equal $a (F-Expected);F-Equal $b (F-Expected)
    }
    F-Case NumericIndependent {
        $a=F-Normal;$b=F-Normal;F-Equal $a (F-Expected);F-Equal $b (F-Expected)
        foreach($field in @('Schema','OriginTick','Frequency','ObservedTick','StopTick')){
            $la=F-Leaf $a $field;$lb=F-Leaf $b $field
            F-Need ((F-Plain $la.Members) -and (F-Plain $lb.Members) -and -not (F-Ref $la.Base $lb.Base))
        }
    }
    F-Case MutationIsolation {
        foreach($field in @('Schema','OriginTick','Frequency','ObservedTick','StopTick')){
            $a=F-Normal;$b=F-Normal;$la=F-Leaf $a $field;$lb=F-Leaf $b $field
            F-Need ((F-Plain $la.Members) -and (F-Plain $lb.Members))
            $name='P2AFixtureIsolation'+$field;$payload=[object[]]@('initial')
            $la.Members.Add([Management.Automation.PSNoteProperty]::new($name,$payload));$held=$la.Members[$name]
            F-Need ($null -ne $held -and (F-Ref $held.Value $payload))
            $held.Value[0]='changed'
            F-Need ((F-Has $la.Members $name $held $payload) -and (F-Has $a.$field.PSObject.Properties $name $held $payload))
            $c=F-Normal;$lc=F-Leaf $c $field
            F-Need (-not (F-Ref $la.Base $lb.Base) -and -not (F-Ref $la.Base $lc.Base))
            F-Need ($null -eq $lb.Members[$name] -and $null -eq $b.$field.PSObject.Properties[$name] -and
                $null -eq $lc.Members[$name] -and $null -eq $c.$field.PSObject.Properties[$name])
            F-Need ((F-Has $la.Members $name $held $payload) -and $payload[0] -ceq 'changed')
            F-Equal $a (F-Expected);F-Equal $b (F-Expected);F-Equal $c (F-Expected)
        }
    }
    F-Case Defaults {
        $a=New-TBRequest;$b=New-TBRequest
        $want=[pscustomobject]@{Schema=[int]1;OriginTick=[long]0;Frequency=[long]1000;ObservedTick=[long]0;StopTick=$null}
        F-Equal $a $want;F-Equal $b $want
        foreach($field in @('Schema','OriginTick','Frequency','ObservedTick')){$la=F-Leaf $a $field;$lb=F-Leaf $b $field;F-Need (-not (F-Ref $la.Base $lb.Base))}
    }
    F-Case WrongTypes {
        foreach($field in @('Schema','OriginTick','Frequency','ObservedTick','StopTick')){
            foreach($value in @([int]2,[long]1,[double]0,'0',$true,$null)){
                $parameters=@{};$parameters[$field]=$value;$r=New-TBRequest @parameters
                F-Need ((F-Type $r.$field) -eq (F-Type $value))
                if($null -eq $value){F-Need ($null -eq $r.$field)}else{F-Need ($r.$field.PSObject.BaseObject -ceq $value.PSObject.BaseObject)}
            }
        }
    }
    F-Case DecoratedInputs {
        foreach($field in @('Schema','OriginTick','Frequency','ObservedTick','StopTick')){
            $type=if($field -ceq 'Schema'){[int]}else{[long]}
            $value=F-Fresh $type 1;$holder=[pscustomobject]@{Value=$value};$members=$holder.Value.PSObject.Properties
            $name='P2AExplicitDecoration'+$field;$payload=[object[]]@('initial')
            $members.Add([Management.Automation.PSNoteProperty]::new($name,$payload));$held=$members[$name];$held.Value[0]='changed'
            F-Need ((F-Has $members $name $held $payload) -and (F-Has $holder.Value.PSObject.Properties $name $held $payload))
            $parameters=@{};$parameters[$field]=$holder.Value;$r=New-TBRequest @parameters
            F-Need ((F-Type $r.$field) -eq $type -and (F-Ref $r.$field.PSObject.BaseObject $holder.Value.PSObject.BaseObject))
            F-Need ((F-Has $r.$field.PSObject.Properties $name $held $payload) -and (F-Has $members $name $held $payload))
        }
    }
    F-Case Ranges {
        foreach($field in @('Schema','OriginTick','Frequency','ObservedTick','StopTick')){
            foreach($value in @([long]-1,[long]999,[long]::MaxValue)){
                $parameters=@{};$parameters[$field]=$value;$r=New-TBRequest @parameters
                F-Need ((F-Type $r.$field) -eq [long] -and $r.$field -eq $value)
            }
        }
    }
    F-Case NoCallbacks {
        $r=[pscustomobject]@{Value=1}
        $r.PSObject.Methods.Add([Management.Automation.PSScriptMethod]::new('GetType',{$script:fixtureCallbacks++;'observable'}))
        $r.PSObject.Methods.Add([Management.Automation.PSScriptMethod]::new('ToString',{$script:fixtureCallbacks++;'observable'}))
        $r.PSObject.Properties.Add([Management.Automation.PSScriptProperty]::new('Probe',{$script:fixtureCallbacks++;1}))
        F-Need ($r.GetType() -ceq 'observable' -and $r.ToString() -ceq 'observable' -and $r.Probe -eq 1 -and $script:fixtureCallbacks -eq 3)
        $script:fixtureCallbacks=0;$result=New-TBRequest -OriginTick $r
        F-Need ((F-Ref $result.OriginTick $r) -and $script:fixtureCallbacks -eq 0)
    }
    F-Case Complete {F-Need ($script:fixtureCallbacks -eq 0)}
    [Console]::Out.WriteLine('LIMIT: fixture regression only; no product import or call')
    exit 0
}catch{[Console]::Out.WriteLine('ASSERTION FAILED: P2AFixture'+$script:fixtureCase);exit 1}
