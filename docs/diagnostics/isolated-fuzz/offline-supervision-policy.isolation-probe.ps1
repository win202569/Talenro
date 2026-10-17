# Owned-fixture diagnostic only. No imports, product calls, fixes or warmups.
$ErrorActionPreference='Stop'
$script:isolationProbeIds=@('ArraysABDistinct','ADescriptorNote','AFieldMatchesAArray','BDescriptorNote','BFieldMatchesBArray','BContentsInitial','CContentsInitial','CArrayDistinctA','CArrayDistinctB','CDescriptorNote','CFieldMatchesCArray')
function IsolationProbe-Type($Value){[object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)}
function IsolationProbe-WrapPrimitive([object]$Value){
    if($null -eq $Value){throw 'FIXTURE ERROR: null is not a wrapped primitive'}
    $ctor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
    if($null -eq $ctor -or -not $ctor.IsPublic){throw 'FIXTURE ERROR: object constructor unavailable'}
    $parameters=@($ctor.GetParameters())
    if($parameters.Count -ne 1 -or $parameters[0].ParameterType -ne [object]){throw 'FIXTURE ERROR: exact object signature required'}
    return ,($ctor.Invoke([object[]]@($Value)))
}
function IsolationProbe-Decorate($Owner,[string]$Name){
    $original=$Owner.$Name;$type=IsolationProbe-Type $original
    $wrapped=IsolationProbe-WrapPrimitive $original
    $metadata=[object[]]@('initial')
    $wrapped.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Metadata',$metadata))
    $Owner.$Name=$wrapped;$field=$Owner.$Name;$property=$field.PSObject.Properties['Metadata']
    if((IsolationProbe-Type $field) -ne $type -or $field.PSObject.BaseObject -cne $original.PSObject.BaseObject -or $null -eq $property -or $property.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or -not [object]::ReferenceEquals($property.Value,$metadata)){throw 'FIXTURE ERROR: decorated scalar was not preserved'}
    return ,$metadata
}
function IsolationProbe-Row([string]$Category){
    [object[]]$checks=@(foreach($id in $script:isolationProbeIds){[pscustomobject]@{Id=$id;Reached=$false;Result=$null}})
    [pscustomobject]@{Schema=1;Kind='category';Category=$Category;Status='NotReached';BaseTypeNameA=$null;BaseTypeNameB=$null;BaseObjectSame=$null;Checks=$checks;FirstFailedCheck=$null}
}
function IsolationProbe-Check($Row,[int]$Index,[bool]$Value){
    $Row.Checks[$Index].Reached=$true;$Row.Checks[$Index].Result=$Value
    if(-not $Value){$Row.FirstFailedCheck=$Row.Checks[$Index].Id};return $Value
}
function IsolationProbe-Category([string]$Category,[scriptblock]$Make){
    $row=IsolationProbe-Row $Category;$row.Status='Observed'
    try{
        $a=[pscustomobject]@{Value=(& $Make)};$b=[pscustomobject]@{Value=(& $Make)}
        $row.BaseTypeNameA=(IsolationProbe-Type $a.Value).FullName;$row.BaseTypeNameB=(IsolationProbe-Type $b.Value).FullName
        $row.BaseObjectSame=[object]::ReferenceEquals($a.Value.PSObject.BaseObject,$b.Value.PSObject.BaseObject)
        $am=IsolationProbe-Decorate $a Value;$bm=IsolationProbe-Decorate $b Value
        if(-not (IsolationProbe-Check $row 0 (-not [object]::ReferenceEquals($am,$bm)))){return $row}
        $am[0]='changed'
        $p=$a.Value.PSObject.Properties['Metadata']
        if(-not (IsolationProbe-Check $row 1 ($null -ne $p -and $p.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty))){return $row}
        if(-not (IsolationProbe-Check $row 2 ([object]::ReferenceEquals($p.Value,$am)))){return $row}
        $p=$b.Value.PSObject.Properties['Metadata']
        if(-not (IsolationProbe-Check $row 3 ($null -ne $p -and $p.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty))){return $row}
        if(-not (IsolationProbe-Check $row 4 ([object]::ReferenceEquals($p.Value,$bm)))){return $row}
        if(-not (IsolationProbe-Check $row 5 ($bm[0] -ceq 'initial'))){return $row}
        $c=[pscustomobject]@{Value=(& $Make)};$cm=IsolationProbe-Decorate $c Value
        if(-not (IsolationProbe-Check $row 6 ($cm[0] -ceq 'initial'))){return $row}
        if(-not (IsolationProbe-Check $row 7 (-not [object]::ReferenceEquals($cm,$am)))){return $row}
        if(-not (IsolationProbe-Check $row 8 (-not [object]::ReferenceEquals($cm,$bm)))){return $row}
        $p=$c.Value.PSObject.Properties['Metadata']
        if(-not (IsolationProbe-Check $row 9 ($null -ne $p -and $p.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty))){return $row}
        $null=IsolationProbe-Check $row 10 ([object]::ReferenceEquals($p.Value,$cm))
        return $row
    }catch{
        if($_.Exception.Message -cne 'FIXTURE ERROR: decorated scalar was not preserved'){throw}
        $row.Status='HelperFailure';$row.FirstFailedCheck='HelperFidelity';return $row
    }
}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=New-Object Security.Principal.WindowsPrincipal($identity)
    Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
    Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    $categories=@('Int32','Int64','Boolean','String')
    $factories=@({ [int]0 },{ [long]0 },{ $true },{ [string]::new([char[]]@('c','l','e','a','n')) })
    $halt=$false;$firstCategory=$null;$firstCheck=$null;$disposition='NotReproduced'
    for($i=0;$i -lt 4;$i++){
        $row=if($halt){IsolationProbe-Row $categories[$i]}else{IsolationProbe-Category $categories[$i] $factories[$i]}
        if(-not $halt -and $null -ne $row.FirstFailedCheck){
            $halt=$true;$firstCategory=$row.Category;$firstCheck=$row.FirstFailedCheck
            $disposition=if($row.Status -ceq 'HelperFailure'){'Inconclusive'}else{'ObservedFailure'}
        }
        Write-Output ('DIAGNOSTIC: '+($row|ConvertTo-Json -Depth 7 -Compress))
    }
    $summary=[pscustomobject]@{Schema=1;Kind='summary';Disposition=$disposition;FirstFailedCategory=$firstCategory;FirstFailedCheck=$firstCheck;WarmupReplayed=$false;ProductImported=$false;OriginalTestsRun=$false;FixtureChanged=$false}
    Write-Output ('DIAGNOSTIC: '+($summary|ConvertTo-Json -Compress))
    Write-Output 'LIMIT: isolation fixture diagnosis only; no original tests, warmups, product calls, repairs or budget renewal.'
}catch{Write-Error $_;exit 1}
