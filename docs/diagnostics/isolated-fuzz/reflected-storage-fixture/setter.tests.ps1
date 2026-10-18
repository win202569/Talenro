# Fixed synthetic storage-transition tests ONLY; no candidate or observer suite.
param([Parameter(Mandatory)][string]$Packet)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$source=Join-Path $PSScriptRoot '../stage-timing-ownership/owned-return.tests.ps1'
$t=$null;$errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile($source,[ref]$t,[ref]$errors)
if($errors.Count){throw 'Source parse error'}
$allow=@('R-Reference','R-Type','R-Base','R-Wrapper','P1O-Fail','P1O-Need','P1O-Leaf','P1R-Fresh')
$functions=@($ast.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst]},$false))
$definitions=@(foreach($name in $allow){
    $match=@($functions|Where-Object Name -CEQ $name)
    if($match.Count -ne 1){throw ('Allowlist count: '+$name)}
    $match[0].Extent.Text
})
$init=@'
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:case='setup-trace'
$script:metadataName='P1PlainOwnershipMetadata'
$script:referenceMethod=[object].GetMethod('ReferenceEquals',[type[]]@([object],[object]))
$script:typeMethod=[object].GetMethod('GetType',[type[]]@())
$script:invokeMethod=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
$script:wrapperConstructor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
$script:baseGetter=[Management.Automation.PSObject].GetProperty('BaseObject').GetGetMethod()
$script:storageSetter=[Management.Automation.PSNoteProperty].GetProperty('Value').GetSetMethod()
'@
$setterPath=Join-Path $PSScriptRoot 'set-stored.ps1'
$st=$null;$se=$null
$sa=[Management.Automation.Language.Parser]::ParseFile($setterPath,[ref]$st,[ref]$se)
$sf=@($sa.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst]},$false))
if($se.Count -or $sf.Count -ne 1 -or $sf[0].Name -cne 'S-SetStored'){throw 'Setter allowlist'}
$module=New-Module -ScriptBlock ([scriptblock]::Create($init+[Environment]::NewLine+($definitions -join [Environment]::NewLine)+[Environment]::NewLine+$sf[0].Extent.Text))
try{
    & $module {
        $cases=@(
            [pscustomobject]@{Kind='Int32';Type=[int];Value=[int]7;Initial=[int];First=[Management.Automation.PSObject];Second=[int]},
            [pscustomobject]@{Kind='Int64';Type=[long];Value=[long]7;Initial=[long];First=[Management.Automation.PSObject];Second=[long]},
            [pscustomobject]@{Kind='Boolean';Type=[bool];Value=$true;Initial=[bool];First=[Management.Automation.PSObject];Second=[bool]},
            [pscustomobject]@{Kind='String';Type=[string];Value='clean';Initial=[string];First=[Management.Automation.PSObject];Second=[string]},
            [pscustomobject]@{Kind='WrappedString';Type=[string];Value='clean';Initial=[Management.Automation.PSObject];First=[string];Second=[Management.Automation.PSObject]}
        )
        $passed=0
        foreach($case in $cases){
            $value=P1R-Fresh $case.Type $case.Value
            $owner=[pscustomobject]@{}
            $owner.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Value',$value))
            $field=$owner.PSObject.Properties['Value']
            if($case.Kind -ceq 'WrappedString'){
                $out=@(S-SetStored $field (R-Wrapper $field.Value))
                if($out.Count){throw 'Setter leaked pipeline output'}
            }
            $leaf=P1O-Leaf $owner Value fixture
            if($leaf.Stored -ne $case.Initial -or $leaf.Type -ne $case.Type){throw ('Initial storage mismatch '+$case.Kind)}
            foreach($step in @(1,2)){
                $want=if($step -eq 1){$case.First}else{$case.Second}
                $before=(R-Type $field.Value)
                if($want -eq [Management.Automation.PSObject]){$replacement=R-Wrapper $field.Value}
                else{$replacement=R-Base $field.Value}
                $out=@(S-SetStored $field $replacement)
                $actual=R-Type $field.Value
                $base=$owner.Value.PSObject.BaseObject
                if($out.Count -or $actual -ne $want -or $actual -eq $before){throw ('Storage transition failed '+$case.Kind+'/'+$step+' actual='+$actual.FullName+' expected='+$want.FullName)}
                if(-not (R-Reference $owner.PSObject.Properties['Value'] $leaf.Descriptor) -or -not (R-Reference $base $leaf.Base) -or (R-Type $base) -ne $leaf.Type -or $base -cne $leaf.Base){throw ('Identity/value drift '+$case.Kind+'/'+$step)}
                if($null -ne $leaf.Members[$script:metadataName] -or $null -ne $owner.Value.PSObject.Properties[$script:metadataName]){throw 'Unexpected metadata'}
                $passed++
                [Console]::Out.WriteLine('PASS: '+$case.Kind+'/'+$step+' '+$before.FullName+' -> '+$actual.FullName+' descriptorSame=true baseSame=true valueSame=true')
            }
        }
        if($passed -ne 10){throw 'Incomplete transitions'}
        [Console]::Out.WriteLine('RESULT: REFLECTED_STORAGE cases=5 transitions=10 passed=10 productCalls=0')
    }
}finally{Remove-Module $module}
