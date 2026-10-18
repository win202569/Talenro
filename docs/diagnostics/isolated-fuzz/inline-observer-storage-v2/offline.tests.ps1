# OFFLINE ONLY: parse source as data; execute only ten observer/helper definitions.
param([Parameter(Mandatory)][string]$Packet)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$names=@('R-Reference','R-Type','R-Base','R-Wrapper','P1O-Fail','P1O-Need','P1O-Leaf','V-Invariant','V-Others','P1R-Fresh')
function Read-Functions($Path){
    $tokens=$null;$errors=$null
    $ast=[Management.Automation.Language.Parser]::ParseFile($Path,[ref]$tokens,[ref]$errors)
    if($errors.Count){throw 'Source parse failed'}
    $defs=@($ast.FindAll({param($a) $a -is [Management.Automation.Language.FunctionDefinitionAst]},$false))
    $selected=@(foreach($name in $names){
        $match=@($defs|Where-Object Name -CEQ $name)
        if($match.Count -ne 1){throw ('Whitelist definition count: '+$name)}
        $match[0].Extent.Text
    })
    return ($selected -join [Environment]::NewLine)
}
$oldPath=Join-Path $PSScriptRoot '../stage-timing-ownership/owned-return.tests.ps1'
$newPath=Join-Path $PSScriptRoot '../inline-ownership-observer/owned-return.tests.ps1'
$oldDefs=Read-Functions $oldPath;$newDefs=Read-Functions $newPath
# Read only the validated setter definition; never execute its file top level.
$setterPath=Join-Path $PSScriptRoot '../reflected-storage-fixture/set-stored.ps1'
$st=$null;$se=$null
$sa=[Management.Automation.Language.Parser]::ParseFile($setterPath,[ref]$st,[ref]$se)
$sf=@($sa.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst]},$false))
if($se.Count -or $sf.Count -ne 1 -or $sf[0].Name -cne 'S-SetStored'){throw 'Setter allowlist'}
$setterDefinition=$sf[0].Extent.Text
# Whole-source equality outside precisely one function, including original timing and assertions.
function Mask-Scan($Path){
    $t=$null;$e=$null
    $a=[Management.Automation.Language.Parser]::ParseFile($Path,[ref]$t,[ref]$e)
    $f=@($a.FindAll({param($n) $n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'V-Others'},$false))
    if($f.Count -ne 1){throw 'Scan count'}
    $s=[IO.File]::ReadAllText($Path).Replace(([string][char]13),'')
    # Parse offsets refer to original bytes-as-characters, so mask before newline normalization.
    $s=[IO.File]::ReadAllText($Path)
    return ($s.Substring(0,$f[0].Extent.StartOffset)+'<SCAN>'+$s.Substring($f[0].Extent.EndOffset)).Replace(([string][char]13),'')
}
if((Mask-Scan $oldPath) -cne (Mask-Scan $newPath)){throw 'Non-scan source changed'}
$init=@'
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:case='fixture'
$script:metadataName='P1PlainOwnershipMetadata'
$script:referenceMethod=[object].GetMethod('ReferenceEquals',[type[]]@([object],[object]))
$script:typeMethod=[object].GetMethod('GetType',[type[]]@())
$script:invokeMethod=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
$script:wrapperConstructor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
$script:baseGetter=[Management.Automation.PSObject].GetProperty('BaseObject').GetGetMethod()
$script:storageSetter=[Management.Automation.PSNoteProperty].GetProperty('Value').GetSetMethod()
'@
$fixtureRunner={
    param([string]$Kind,[string]$Action)
    function Fresh-Value {
        switch -CaseSensitive ($Kind){
            Int32 {return ,(P1R-Fresh ([int]) ([int]7))}
            Int64 {return ,(P1R-Fresh ([long]) ([long]7))}
            Boolean {return ,(P1R-Fresh ([bool]) $true)}
            String {return ,(P1R-Fresh ([string]) 'clean')}
            WrappedString {return ,(R-Wrapper (P1R-Fresh ([string]) 'clean'))}
            default {throw 'Unknown kind'}
        }
    }
    function Owner($Value){
        $o=[pscustomobject]@{}
        $o.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Value',$Value))
        if($Kind -ceq 'WrappedString'){S-SetStored $o.PSObject.Properties['Value'] (R-Wrapper $o.PSObject.Properties['Value'].Value)}
        return ,$o
    }
    function Outcome($Leaves,$Selected){
        try{
            $out=@(V-Others $Leaves $Selected)
            if($out.Count){throw 'Unexpected pipeline output'}
            return 'PASS'
        }catch{
            $message=$_.Exception.Message
            if($message -cin @('FIXTURE_INVALID|fixture|field-base-changed','OWNERSHIP_GAP_OBSERVED|fixture|cross-owner:other')){return $message}
            throw
        }
    }
    function Decorate($Members){
        $payload=[object[]]@('offline-payload',17)
        $Members.Add([Management.Automation.PSNoteProperty]::new($script:metadataName,$payload))
        if($null -eq $Members[$script:metadataName] -or -not (R-Reference $Members[$script:metadataName].Value $payload)){throw 'Decoration setup failed'}
    }
    $selectedOwner=Owner (Fresh-Value)
    $otherOwner=Owner (Fresh-Value)
    if($Action -eq 'shared'){
        # Deliberately one owner slot copied to the other before either capture.
        S-SetStored $otherOwner.PSObject.Properties['Value'] $selectedOwner.PSObject.Properties['Value'].Value
    }
    if($Action -like 'same-owner*'){
        $selectedOwner.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Other',(Fresh-Value)))
        if($Kind -ceq 'WrappedString'){S-SetStored $selectedOwner.PSObject.Properties['Other'] (R-Wrapper $selectedOwner.PSObject.Properties['Other'].Value)}
        $otherOwner=$selectedOwner;$otherField='Other'
    }else{$otherField='Value'}
    $selected=P1O-Leaf $selectedOwner Value selected
    $other=P1O-Leaf $otherOwner $otherField other
    $wantedStored=switch -CaseSensitive ($Kind){Int32 {[int]} Int64 {[long]} Boolean {[bool]} String {[string]} WrappedString {[Management.Automation.PSObject]}}
    if($selected.Stored -ne $wantedStored -or $other.Stored -ne $wantedStored){throw 'Initial storage-kind fixture mismatch'}
    $leaves=[object[]]@($selected,$other)
    if($Action -eq 'late'){
        $middle=P1O-Leaf (Owner (Fresh-Value)) Value middle
        $leaves=[object[]]@($selected,$middle,$other)
    }
    if($Action -eq 'empty'){$leaves=[object[]]@()}
    if($Action -eq 'reread'){
        # Synthetic branch isolation only: substitute an independent real clean collection.
        $decoy=Owner (Fresh-Value)
        $other.Members=$decoy.Value.PSObject.Properties
    }
    $before=Outcome $leaves $selected
    switch -CaseSensitive ($Action){
        clean {}
        empty {}
        'same-owner-clean' {}
        selected {Decorate $selected.Members}
        descriptor {
            $saved=$otherOwner.PSObject.Properties[$otherField].Value
            $otherOwner.PSObject.Properties.Remove($otherField)
            $otherOwner.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($otherField,$saved))
        }
        storage {
            $field=$otherOwner.PSObject.Properties[$otherField]
            if($Kind -eq 'WrappedString'){S-SetStored $field (R-Base $field.Value)}
            else{S-SetStored $field (R-Wrapper $field.Value)}
            if((R-Type $field.Value) -eq $other.Stored){throw 'Storage fixture did not change type'}
        }
        base {
            S-SetStored $otherOwner.PSObject.Properties[$otherField] (Fresh-Value)
            $freshBase=$otherOwner.($otherField).PSObject.BaseObject
            if(R-Reference $freshBase $other.Base){throw 'Base fixture not fresh'}
        }
        type {$other.Type=[decimal]}
        held {Decorate $other.Members}
        reread {
            Decorate $otherOwner.($otherField).PSObject.Properties
            if($null -ne $other.Members[$script:metadataName]){throw 'Reread fixture held decoy contaminated'}
        }
        shared {
            if(-not (R-Reference $selected.Base $other.Base)){throw 'Shared fixture has no shared base'}
            Decorate $selected.Members
            if($null -eq $otherOwner.($otherField).PSObject.Properties[$script:metadataName]){throw 'Shared fixture has no visible contamination'}
        }
        'same-owner-dirty' {Decorate $other.Members}
        late {Decorate $other.Members}
        default {throw 'Unknown action'}
    }
    $after=Outcome $leaves $selected
    return [pscustomobject]@{Before=$before;After=$after}
}
$inventory=[Collections.Generic.List[object]]::new()
$invalid='FIXTURE_INVALID|fixture|field-base-changed'
$gap='OWNERSHIP_GAP_OBSERVED|fixture|cross-owner:other'
foreach($kind in @('Int32','Int64','Boolean','String','WrappedString')){
    foreach($action in @('clean','selected','descriptor','storage','base','type','held')){
        $expected=if($action -in @('clean','selected')){'PASS'}elseif($action -eq 'held'){$gap}else{$invalid}
        $inventory.Add([pscustomobject]@{Kind=$kind;Action=$action;Expected=$expected})
    }
}
foreach($kind in @('Int32','Int64','Boolean','WrappedString')){$inventory.Add([pscustomobject]@{Kind=$kind;Action='reread';Expected=$gap})}
foreach($kind in @('Int32','WrappedString')){$inventory.Add([pscustomobject]@{Kind=$kind;Action='shared';Expected=$gap})}
foreach($action in @('same-owner-clean','same-owner-dirty','late','empty')){
    $inventory.Add([pscustomobject]@{Kind='Int32';Action=$action;Expected=$(if($action -in @('same-owner-clean','empty')){'PASS'}else{$gap})})
}
if($inventory.Count -ne 45){throw 'Inventory count'}
[IO.File]::WriteAllText((Join-Path $Packet 'fixture-inventory.json'),($inventory|ConvertTo-Json -Depth 5),[Text.UTF8Encoding]::new($false))
$passed=0
foreach($case in $inventory){
    $results=@(foreach($defs in @($oldDefs,$newDefs)){
        $module=New-Module -ScriptBlock ([scriptblock]::Create($init+[Environment]::NewLine+$defs+[Environment]::NewLine+$setterDefinition))
        try{& $module $fixtureRunner $case.Kind $case.Action}
        finally{Remove-Module $module}
    })
    if($results.Count -ne 2 -or $results[0].Before -cne 'PASS' -or $results[1].Before -cne 'PASS' -or $results[0].After -cne $case.Expected -or $results[1].After -cne $case.Expected){
        [Console]::Out.WriteLine('FAIL: '+($case|ConvertTo-Json -Compress)+' actual='+($results|ConvertTo-Json -Compress))
        throw 'Observer mismatch; no retry permitted'
    }
    $passed++
    [Console]::Out.WriteLine('PASS: '+$case.Kind+'/'+$case.Action+' before=PASS after='+$case.Expected)
}
[Console]::Out.WriteLine('RESULT: OBSERVER_EQUIVALENCE cases=45 passed='+$passed+' productCalls=0')
