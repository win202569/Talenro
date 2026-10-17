# Diagnostic spike only: no original test-script import, candidate edit or repair.
param([Parameter(Mandatory=$true)][ValidateSet('clean','preflight')][string]$Mode)
$ErrorActionPreference='Stop'
# The following nine definitions are exact static copies of frozen fixture helpers.
function P1I1-Need([bool]$Value){if(-not $Value){throw ('ASSERTION FAILED: '+$script:p1i1Case)}}
function P1I1-Case([string]$Name,[scriptblock]$Body){$script:p1i1Case='P1I1'+$Name;& $Body;Write-Output ('PASS: '+$script:p1i1Case)}
function P1I1-Type($Value){[object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)}
function P1I1-WrapPrimitive([object]$Value){
    if($null -eq $Value){throw 'FIXTURE ERROR: null is not a wrapped primitive'}
    # Decorate owned copies, never cached literals used by plain control scenarios.
    $type=P1I1-Type $Value
    if($type -eq [string]){$fresh=[string]::Copy($Value.PSObject.BaseObject)}
    elseif($type -eq [int] -or $type -eq [long] -or $type -eq [bool]){
        $storage=[Array]::CreateInstance($type,1);$storage.SetValue($Value.PSObject.BaseObject,0);$fresh=$storage.GetValue(0)
    }else{throw 'FIXTURE ERROR: unsupported primitive'}
    $ctor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
    if($null -eq $ctor -or -not $ctor.IsPublic){throw 'FIXTURE ERROR: object constructor unavailable'}
    $parameters=@($ctor.GetParameters())
    if($parameters.Count -ne 1 -or $parameters[0].ParameterType -ne [object]){throw 'FIXTURE ERROR: exact object signature required'}
    return ,($ctor.Invoke([object[]]@($fresh)))
}
function P1I1-Decorate($Owner,[string]$Name){
    $original=$Owner.$Name;$type=P1I1-Type $original
    $wrapped=P1I1-WrapPrimitive $original
    $metadata=[object[]]@('initial')
    $wrapped.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Metadata',$metadata))
    $Owner.$Name=$wrapped;$field=$Owner.$Name;$property=$field.PSObject.Properties['Metadata']
    if((P1I1-Type $field) -ne $type -or $field.PSObject.BaseObject -cne $original.PSObject.BaseObject -or $null -eq $property -or $property.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or -not [object]::ReferenceEquals($property.Value,$metadata)){throw 'FIXTURE ERROR: decorated scalar was not preserved'}
    return ,$metadata
}
function P1I1-FixtureScalar($Value,$Type,$Expected){
    P1I1-Need ((P1I1-Type $Value) -eq $Type -and $Value.PSObject.BaseObject -ceq $Expected)
}
function P1I1-FixtureMetadata($Value,$Reference){
    $p=$Value.PSObject.Properties['Metadata'];P1I1-Need ($null -ne $p -and $p.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty)
    P1I1-Need ([object]::ReferenceEquals($p.Value,$Reference))
}
function P1I1-IsolationVectors {
    @(
        @{Make={ $storage=[Array]::CreateInstance([int],1);return ,($storage.GetValue(0)) };Type=[int];Want=[int]0},
        @{Make={ $storage=[Array]::CreateInstance([long],1);return ,($storage.GetValue(0)) };Type=[long];Want=[long]0},
        @{Make={ $storage=[Array]::CreateInstance([bool],1);$storage.SetValue($true,0);return ,($storage.GetValue(0)) };Type=[bool];Want=$true},
        @{Make={ [string]::new([char[]]@('c','l','e','a','n')) };Type=[string];Want='clean'}
    )
}
function P1I1-FixtureChecks {
    P1I1-Case FixtureConstructor {
        $ctor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]));P1I1-Need ($null -ne $ctor -and $ctor.IsPublic)
        $parameters=@($ctor.GetParameters());P1I1-Need ($parameters.Count -eq 1 -and $parameters[0].ParameterType -eq [object])
        P1I1-FixtureScalar (P1I1-WrapPrimitive ([int]0)) ([int]) ([int]0)
    }
    P1I1-Case FixturePrimitiveMatrix {
        $vectors=@(
            @{Make={ [int]0 };Type=[int];Want=[int]0},@{Make={ [int]1 };Type=[int];Want=[int]1},@{Make={ [int]-1 };Type=[int];Want=[int]-1},
            @{Make={ [int]::MinValue };Type=[int];Want=[int]::MinValue},@{Make={ [int]::MaxValue };Type=[int];Want=[int]::MaxValue},
            @{Make={ [long]0 };Type=[long];Want=[long]0},@{Make={ [long]::MinValue };Type=[long];Want=[long]::MinValue},@{Make={ [long]::MaxValue };Type=[long];Want=[long]::MaxValue},
            @{Make={ $true };Type=[bool];Want=$true},@{Make={ $false };Type=[bool];Want=$false},
            @{Make={ [string]::new([char[]]@()) };Type=[string];Want=''},@{Make={ [string]::new([char[]]@('c','l','e','a','n')) };Type=[string];Want='clean'},
            @{Make={ [string]::new([char[]]@('r','e','f','u','s','e','d')) };Type=[string];Want='refused'},@{Make={ [string]::new([char[]]@('u','n','k','n','o','w','n')) };Type=[string];Want='unknown'}
        )
        foreach($v in $vectors){
            $value=& $v.Make;P1I1-FixtureScalar $value $v.Type $v.Want
            $w=P1I1-WrapPrimitive $value;P1I1-FixtureScalar $w $v.Type $v.Want
            $meta=[object[]]@('initial');$w.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Metadata',$meta))
            P1I1-FixtureScalar $w $v.Type $v.Want;P1I1-FixtureMetadata $w $meta
            $owner=[pscustomobject]@{Value=$null};$owner.Value=$w;P1I1-FixtureScalar $owner.Value $v.Type $v.Want;P1I1-FixtureMetadata $owner.Value $meta
            $retrieved=$owner.Value;P1I1-FixtureScalar $retrieved $v.Type $v.Want;P1I1-FixtureMetadata $retrieved $meta
            $other=[pscustomobject]@{Value=(& $v.Make)};$returned=P1I1-Decorate $other Value
            P1I1-FixtureScalar $other.Value $v.Type $v.Want;P1I1-FixtureMetadata $other.Value $returned
        }
    }
    P1I1-Case FixtureDescriptors {
        $script:p1i1FixtureCallbacks=0
        foreach($kind in @('ScriptProperty','AliasProperty','ScriptMethod')){
            $w=P1I1-WrapPrimitive ([int]0)
            switch($kind){
                ScriptProperty {$member=[Management.Automation.PSScriptProperty]::new('Probe',{ $script:p1i1FixtureCallbacks++;return 'called' })}
                AliasProperty {$member=[Management.Automation.PSAliasProperty]::new('Probe','MissingTarget')}
                ScriptMethod {$member=[Management.Automation.PSScriptMethod]::new('Probe',{ $script:p1i1FixtureCallbacks++;return 'called' })}
            }
            $w.PSObject.Members.Add($member);$owner=[pscustomobject]@{Value=$w};P1I1-FixtureScalar $owner.Value ([int]) ([int]0)
            $descriptor=$owner.Value.PSObject.Members['Probe'];P1I1-Need ($null -ne $descriptor -and $descriptor.Name -ceq 'Probe' -and $descriptor.MemberType.ToString() -ceq $kind)
            P1I1-Need ($script:p1i1FixtureCallbacks -eq 0)
        }
    }
    P1I1-Case FixtureInputIdentity {
        foreach($v in (P1I1-IsolationVectors)){
            $a=[pscustomobject]@{Value=(& $v.Make)};$b=[pscustomobject]@{Value=(& $v.Make)};$c=[pscustomobject]@{Value=(& $v.Make)}
            P1I1-FixtureScalar $a.Value $v.Type $v.Want;P1I1-FixtureScalar $b.Value $v.Type $v.Want;P1I1-FixtureScalar $c.Value $v.Type $v.Want
            P1I1-Need (-not [object]::ReferenceEquals($a.Value.PSObject.BaseObject,$b.Value.PSObject.BaseObject))
            P1I1-Need (-not [object]::ReferenceEquals($a.Value.PSObject.BaseObject,$c.Value.PSObject.BaseObject))
            P1I1-Need (-not [object]::ReferenceEquals($b.Value.PSObject.BaseObject,$c.Value.PSObject.BaseObject))
        }
    }
    P1I1-Case FixtureIsolation {
        foreach($v in (P1I1-IsolationVectors)){
            $make=$v.Make
            $a=[pscustomobject]@{Value=(& $make)};$b=[pscustomobject]@{Value=(& $make)}
            $am=P1I1-Decorate $a Value;$bm=P1I1-Decorate $b Value
            P1I1-Need (-not [object]::ReferenceEquals($am,$bm));$am[0]='changed'
            P1I1-FixtureMetadata $a.Value $am;P1I1-FixtureMetadata $b.Value $bm;P1I1-Need ($bm[0] -ceq 'initial')
            $c=[pscustomobject]@{Value=(& $make)};$cm=P1I1-Decorate $c Value
            P1I1-Need ($cm[0] -ceq 'initial' -and -not [object]::ReferenceEquals($cm,$am) -and -not [object]::ReferenceEquals($cm,$bm))
            P1I1-FixtureMetadata $c.Value $cm
        }
    }
    P1I1-Case FixtureWrappedInputIdentity {
        foreach($v in (P1I1-IsolationVectors)){
            $inputValue=& $v.Make;$w=P1I1-WrapPrimitive $inputValue
            P1I1-FixtureScalar $w $v.Type $v.Want
            P1I1-Need (-not [object]::ReferenceEquals($inputValue.PSObject.BaseObject,$w.PSObject.BaseObject))
            $metadata=[object[]]@('owned');$w.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Metadata',$metadata))
            P1I1-FixtureMetadata $w $metadata
            P1I1-Need ($null -eq $inputValue.PSObject.Properties['Metadata'])
        }
    }
    P1I1-Case FixtureComplete {P1I1-Need ($null -eq (Get-Command Get-OfflineSupervisionDecision -CommandType Function -ErrorAction SilentlyContinue))}
    Write-Output 'LIMIT: fixture construction only; no product import, product assertions, I1 closure or repair-budget renewal.'
}

function EmptyProbe-Error($Record){
    [pscustomobject]@{Message=$Record.Exception.Message;Position=$Record.InvocationInfo.PositionMessage;Stack=$Record.ScriptStackTrace;Id=$Record.FullyQualifiedErrorId}
}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
    Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
    Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    if($Mode -ceq 'preflight'){P1I1-FixtureChecks}
    . (Join-Path $PSScriptRoot 'offline-supervision-policy.ps1')
    $scenario=[pscustomobject]@{Schema=1;Mode='policy-only';SessionLabel='s';ClockLabel='synthetic-ms';StartMs=0;Observations=[object[]]@()}
    $result=Get-OfflineSupervisionDecision $scenario
    # Match product's strict validation context; catch only for diagnostic transport.
    $shape=& {
        Set-StrictMode -Version Latest
        try{
            Assert-SupervisionPolicyShape $scenario @('Schema','Mode','SessionLabel','ClockLabel','StartMs','Observations')
            [pscustomobject]@{Passed=$true;Error=$null}
        }catch{[pscustomobject]@{Passed=$false;Error=(EmptyProbe-Error $_)}}
    }
    # Descriptor metadata only, for the five owned primitive fields; never member.Value.
    $fields=@(foreach($name in @('Schema','Mode','SessionLabel','ClockLabel','StartMs')){
        $value=$scenario.$name
        $members=@(foreach($member in $value.PSObject.Members){
            [pscustomobject]@{Name=$member.Name;MemberType=$member.MemberType.ToString();IsInstance=$member.IsInstance}
        })
        [pscustomobject]@{Field=$name;BaseType=(Get-SupervisionPolicyType $value).FullName;Members=$members}
    })
    $data=[pscustomobject]@{Schema=1;Mode=$Mode;PreflightRan=($Mode -ceq 'preflight');DecisionCalls=1;ShapeCalls=1;Result=$result;DirectShape=$shape;Fields=$fields;OriginalTestsRun=$false;CandidateChanged=$false}
    Write-Output ('DIAGNOSTIC: '+($data|ConvertTo-Json -Depth 10 -Compress))
    Write-Output 'LIMIT: empty-control diagnosis only; no original tests, candidate edits, retries or budget renewal.'
}catch{Write-Error $_;exit 1}

