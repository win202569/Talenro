# Synthetic Int32 setup trace ONLY. Never execute the full test or candidate.
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
$script:sequence=0
'@
$module=New-Module -ScriptBlock ([scriptblock]::Create($init+[Environment]::NewLine+($definitions -join [Environment]::NewLine)))
try{
    & $module {
        function Emit([string]$Path,[string]$Stage,[string]$NativeType,$Details){
            $script:sequence++
            $row=[pscustomobject][ordered]@{Sequence=$script:sequence;Path=$Path;Stage=$Stage;NativeType=$NativeType;Details=$Details}
            [Console]::Out.WriteLine('TRACE: '+($row|ConvertTo-Json -Compress))
        }
        # Exact expressions from the failed Int32 fixture's Fresh-Value and Owner.
        function Fresh-Value {return ,(P1R-Fresh ([int]) ([int]7))}
        function Owner($Value){
            $o=[pscustomobject]@{}
            $o.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Value',$Value))
            return ,$o
        }
        # Path 1: same setup/assignment expressions. No observers are invoked.
        $owner=Owner (Fresh-Value)
        $leaf=P1O-Leaf $owner Value other
        $field=$owner.PSObject.Properties['Value']
        $before=(R-Type $field.Value).FullName
        Emit exact before-assignment $before ([ordered]@{CapturedStored=$leaf.Stored.FullName;CapturedBase=$leaf.Type.FullName})
        $field.Value=R-Wrapper $field.Value
        $after=(R-Type $field.Value).FullName
        $base=$owner.Value.PSObject.BaseObject
        Emit exact after-assignment $after ([ordered]@{
            StorageChanged=($after -cne $leaf.Stored.FullName)
            DescriptorSame=(R-Reference $field $leaf.Descriptor)
            BaseSame=(R-Reference $base $leaf.Base)
            BaseType=(R-Type $base).FullName
        })

        # Path 2: separate, deliberately instrumented split boundaries and raw control.
        # These extra locals/calls are diagnostic context, not proof of equivalence.
        $array=[Array]::CreateInstance([int],1);$array.SetValue([int]7,0)
        $raw=$array.GetValue(0)
        Emit split raw-array-box (R-Type $raw).FullName $null
        $fresh=P1R-Fresh ([int]) ([int]7)
        Emit split fresh-helper-return (R-Type $fresh).FullName $null
        $nested=Fresh-Value
        Emit split nested-fresh-return (R-Type $nested).FullName $null
        function Traced-Owner($Value){
            Emit split owner-parameter (R-Type $Value).FullName $null
            $o=[pscustomobject]@{}
            $o.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Value',$Value))
            Emit split owner-slot-inside (R-Type $o.PSObject.Properties['Value'].Value).FullName $null
            return ,$o
        }
        $owner2=Traced-Owner (Fresh-Value)
        $field2=$owner2.PSObject.Properties['Value']
        Emit split owner-slot-returned (R-Type $field2.Value).FullName $null
        $leaf2=P1O-Leaf $owner2 Value other
        Emit split after-leaf-capture (R-Type $field2.Value).FullName ([ordered]@{CapturedStored=$leaf2.Stored.FullName})
        $args2=[object[]]@($field2.Value)
        Emit split constructor-argument (R-Type $args2[0]).FullName ([ordered]@{ArgumentCount=$args2.Count})
        $direct=$script:wrapperConstructor.Invoke($args2)
        Emit split direct-constructor-result (R-Type $direct).FullName $null
        $wrapped=R-Wrapper $field2.Value
        Emit split original-wrapper-return (R-Type $wrapped).FullName $null
        $field2.Value=$wrapped
        Emit split slot-after-local-assignment (R-Type $field2.Value).FullName ([ordered]@{StorageChanged=((R-Type $field2.Value) -ne $leaf2.Stored)})
        # Control: no Fresh/Owner function pipeline before the note-property constructor.
        $control=[pscustomobject]@{}
        $control.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Value',$array.GetValue(0)))
        Emit split direct-slot-control (R-Type $control.PSObject.Properties['Value'].Value).FullName $null
        if($script:sequence -ne 14){throw 'Trace inventory incomplete'}
        [Console]::Out.WriteLine('RESULT: STORAGE_SETUP_TRACE paths=2 records=14 productCalls=0')
    }
}finally{Remove-Module $module}
