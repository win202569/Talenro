# Diagnostic only. No product/test imports, file operations, fixes or live adapters.
$ErrorActionPreference='Stop'
$script:fixtureProbeRows=New-Object 'System.Collections.Generic.List[object]'
function FixtureProbe-Scenario([object[]]$Observations=[object[]]@()){
    [pscustomobject]@{Schema=1;Mode='policy-only';SessionLabel='s';ClockLabel='synthetic-ms';StartMs=0;Observations=$Observations}
}
function FixtureProbe-Observation($Sequence,$At,[string]$Kind,[string]$Actor,[hashtable]$Payload=@{}){
    $v=[ordered]@{Sequence=$Sequence;SessionLabel='s';ClockLabel='synthetic-ms';AtMs=$At;Actor=$Actor;Kind=$Kind};foreach($k in $Payload.Keys){$v[$k]=$Payload[$k]};[pscustomobject]$v
}
function FixtureProbe-Normal {
    FixtureProbe-Scenario ([object[]]@(
        (FixtureProbe-Observation 0 0 arm W @{LauncherOwnsControl=$true;SupervisorOwnsControl=$true;ContainerBound=$true;HandleInheritanceBlocked=$true;ReadersReady=$true}),
        (FixtureProbe-Observation 1 1 release W @{InputBindingMatches=$true}),
        (FixtureProbe-Observation 2 2 progress C @{StdoutRetainedBytes=0;StderrRetainedBytes=0;DataEventCount=3;StdoutEOF=$true;StderrEOF=$true;ExitObserved=$true;ExitCode=0}),
        (FixtureProbe-Observation 3 3 cleanup W @{Verified=$true;NaturalCompletion=$true;CountsReconciled=$true}),
        (FixtureProbe-Observation 4 4 persist W @{ReadbackMatches=$true;Classification='clean'})
    ))
}
function FixtureProbe-Type($Value){[object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)}
# The original operations and combined guard, with helper names isolated only.
function FixtureProbe-DecorateOriginal($Owner,[string]$Name){
    $original=$Owner.$Name;$type=FixtureProbe-Type $original
    $wrapped=[Management.Automation.PSObject]::new($original)
    $metadata=[object[]]@('initial')
    $wrapped.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Metadata',$metadata))
    $Owner.$Name=$wrapped;$field=$Owner.$Name;$property=$field.PSObject.Properties['Metadata']
    if((FixtureProbe-Type $field) -ne $type -or $field.PSObject.BaseObject -cne $original.PSObject.BaseObject -or $null -eq $property -or $property.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or -not [object]::ReferenceEquals($property.Value,$metadata)){throw 'FIXTURE ERROR: decorated scalar was not preserved'}
    return ,$metadata
}
function FixtureProbe-Stage([string]$Name,$Value,$Reference,[bool]$Reached=$true){
    $typeName=$null;$valueMatches=$null;$present=$null;$kind=$null;$same=$null
    if($Reached){
        $typeName=(FixtureProbe-Type $Value).FullName;$valueMatches=[bool]($Value.PSObject.BaseObject -ceq [int]0)
        $property=$Value.PSObject.Properties['Metadata'];$present=($null -ne $property)
        if($present){$kind=$property.MemberType.ToString();if($property.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty){$same=[object]::ReferenceEquals($property.Value,$Reference)}}
    }
    $script:fixtureProbeRows.Add([pscustomobject]@{Schema=1;Row=$Name;Status=$(if($Reached){'Observed'}else{'NotReached'});BaseTypeName=$typeName;PrimitiveValueMatches=$valueMatches;MetadataPresent=$present;MetadataKind=$kind;MetadataReferenceSame=$same})
}
function FixtureProbe-DecorateInstrumented($Owner,[string]$Name){
    $original=$Owner.$Name;$type=FixtureProbe-Type $original
    FixtureProbe-Stage original $original $null
    $wrapped=[Management.Automation.PSObject]::new($original)
    FixtureProbe-Stage wrapper-created $wrapped $null
    $metadata=[object[]]@('initial')
    $wrapped.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Metadata',$metadata))
    FixtureProbe-Stage metadata-added $wrapped $metadata
    $Owner.$Name=$wrapped
    FixtureProbe-Stage field-assigned $Owner.$Name $metadata
    $field=$Owner.$Name;$property=$field.PSObject.Properties['Metadata']
    FixtureProbe-Stage field-retrieved $field $metadata
    # Evaluate the original combined expression independently of split predicates.
    $combinedFailure=((FixtureProbe-Type $field) -ne $type -or $field.PSObject.BaseObject -cne $original.PSObject.BaseObject -or $null -eq $property -or $property.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or -not [object]::ReferenceEquals($property.Value,$metadata))
    $predicates=[ordered]@{TypeMatches=((FixtureProbe-Type $field) -eq $type);ValueMatches=($field.PSObject.BaseObject -ceq $original.PSObject.BaseObject);DescriptorPresent=($null -ne $property);DescriptorIsNoteProperty=($null -ne $property -and $property.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty);NestedReferenceSame=$false}
    if($predicates.DescriptorIsNoteProperty){$predicates.NestedReferenceSame=[object]::ReferenceEquals($property.Value,$metadata)}
    [object[]]$falseNames=@(foreach($key in $predicates.Keys){if(-not $predicates[$key]){$key}})
    $script:fixtureProbeInstrumentedGuard=-not [bool]$combinedFailure
    $script:fixtureProbeRows.Add([pscustomobject]@{Schema=1;Row='guard';TypeMatches=[bool]$predicates.TypeMatches;ValueMatches=[bool]$predicates.ValueMatches;DescriptorPresent=[bool]$predicates.DescriptorPresent;DescriptorIsNoteProperty=[bool]$predicates.DescriptorIsNoteProperty;NestedReferenceSame=[bool]$predicates.NestedReferenceSame;OriginalCombinedGuard=$script:fixtureProbeInstrumentedGuard;FalsePredicates=$falseNames})
    if($combinedFailure){throw 'FIXTURE ERROR: decorated scalar was not preserved'}
    return ,$metadata
}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=New-Object Security.Principal.WindowsPrincipal($identity)
    Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
    Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    $originalScenario=FixtureProbe-Normal;$originalScenario.Observations[2].ExitCode=[int]0
    $originalGuard=$true;$originalOutcome='Returned'
    try{$originalMetadata=FixtureProbe-DecorateOriginal $originalScenario.Observations[2] ExitCode}
    catch{if($_.Exception.Message -cne 'FIXTURE ERROR: decorated scalar was not preserved'){throw};$originalGuard=$false;$originalOutcome='FixtureFailure'}
    $script:fixtureProbeRows.Add([pscustomobject]@{Schema=1;Row='original-result';Outcome=$originalOutcome;OriginalCombinedGuard=$originalGuard})
    $instrumentedScenario=FixtureProbe-Normal;$instrumentedScenario.Observations[2].ExitCode=[int]0
    $script:fixtureProbeInstrumentedGuard=$null;$reachedReturn=$true
    try{$returnedMetadata=FixtureProbe-DecorateInstrumented $instrumentedScenario.Observations[2] ExitCode}
    catch{if($_.Exception.Message -cne 'FIXTURE ERROR: decorated scalar was not preserved'){throw};$reachedReturn=$false}
    if($reachedReturn){FixtureProbe-Stage helper-return $instrumentedScenario.Observations[2].ExitCode $returnedMetadata}
    else{FixtureProbe-Stage helper-return $null $null $false}
    $guardRow=$script:fixtureProbeRows[6];$agreement=($originalGuard -eq $script:fixtureProbeInstrumentedGuard)
    $disposition=if($agreement -and -not $originalGuard -and $guardRow.FalsePredicates.Count -eq 1){'SingleFalsePredicateObserved'}else{'Inconclusive'}
    $script:fixtureProbeRows.Add([pscustomobject]@{Schema=1;Row='conclusion';OriginalAndInstrumentedAgree=$agreement;Disposition=$disposition;ProductImported=$false;ProductEvaluated=$false;OriginalTestsRun=$false;FixtureRepaired=$false})
    foreach($row in $script:fixtureProbeRows){Write-Output ('DIAGNOSTIC: '+($row|ConvertTo-Json -Depth 5 -Compress))}
    Write-Output 'LIMIT: fixture-only diagnosis; no product or original test execution, correction, acceptance or repair-budget renewal.'
}catch{Write-Error $_;exit 1}
