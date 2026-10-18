# Evidence-only matrix for the saved scalar-decoration coverage limitation.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$leaf=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../p2b-sdk-structural-compatibility'))
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
$record=[ordered]@{Schema=1;Mode='p2b-scalar-retained-diagnostic-only';HostVersion=$PSVersionTable.PSVersion.ToString();Core=($PSVersionTable.PSEdition -ceq 'Core');X64=[Environment]::Is64BitProcess;Admin=$admin;Stage='Context';Captured=$false;FixtureObservable=$false;CompilePerformed=$false;ProductImported=$false;ProductCalled=$false;FullSuiteRun=$false;FixtureInsertion='exact-sdk-raw-add';ExpectedRows=35;Rows=@();Error=$null;RuntimeApproved=$false}
function SP-Counts {
    [pscustomobject]@{MemberReads=[P2BDescriptorFixture.CounterNote]::MemberReads;ValueReads=[P2BDescriptorFixture.CounterNote]::ValueReads;Copies=[P2BDescriptorFixture.CounterNote]::Copies;Writes=[P2BDescriptorFixture.CounterNote]::Writes;Total=[P2BDescriptorFixture.CounterNote]::Total}
}
function SP-Raw($Value){
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $set=$Value.psextended;$setBase=$set.PSObject.BaseObject
    BC-Need ((BC-Type $setBase) -eq [Management.Automation.PSMemberSet])
    $field=[Management.Automation.PSMemberSet].GetField('internalMembers',$flags)
    $raw=$field.GetValue($setBase)
    BC-Need ((BC-Type $raw) -eq $field.FieldType)
    return ,$raw
}
function SP-InsertRaw($Value,$Descriptor){
    # Test construction only. Exact SDK raw Add keeps coherent hidden counts
    # without the public integrating collection's descriptor Copy path.
    $raw=SP-Raw $Value
    $field=[Management.Automation.PSMemberSet].GetField('internalMembers',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly')
    $method=$field.FieldType.GetMethod('Add',[type[]]@([Management.Automation.PSMemberInfo]))
    BC-Need ($null -ne $method -and $method.DeclaringType -eq $field.FieldType)
    $arguments=[object[]]::new(1);$arguments[0]=$Descriptor.PSObject.BaseObject
    $null=$method.Invoke($raw.PSObject.BaseObject,$arguments)
}
function SP-FindRetained($Value,[string]$Tag){
    $field=[Management.Automation.PSMemberInfo].GetField('name',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly')
    foreach($entry in (SP-Raw $Value)){
        if($field.GetValue($entry.PSObject.BaseObject) -ceq $Tag){return ,$entry}
    }
    return $null
}
function SP-Snapshot($Value,$Descriptor,[string]$Tag){
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $type=[Management.Automation.PSMemberInfo]
    $nameField=$type.GetField('name',$flags)
    $hiddenField=$type.GetField('<IsHidden>k__BackingField',$flags)
    $reservedField=$type.GetField('<IsReservedMember>k__BackingField',$flags)
    $raw=SP-Raw $Value
    $total=0;$ordinary=0;$hits=0;$found=$null
    foreach($entry in $raw){
        $total++;$base=$entry.PSObject.BaseObject
        if(-not $reservedField.GetValue($base)){$ordinary++}
        if($nameField.GetValue($base) -ceq $Tag){
            $hits++
            $found=[pscustomobject]@{Name=$nameField.GetValue($base);CLRType=(BC-Type $entry).FullName;Hidden=$hiddenField.GetValue($base);Reserved=$reservedField.GetValue($base);SameDescriptor=(BC-Ref $base $Descriptor.PSObject.BaseObject)}
        }
    }
    [pscustomobject]@{ValueCLRType=(BC-Type $Value).FullName;RawCount=$total;NonReservedCount=$ordinary;TagCount=$hits;Tag=$found;DescriptorValuesRead=$false}
}
function SP-Result($Result){
    [pscustomobject]@{Schema=$Result.Schema;Mode=$Result.Mode;Accepted=$Result.Accepted;Reason=$Result.Reason;BytesPresent=($null -ne $Result.Bytes);MessagePresent=($null -ne $Result.Message);NextSequence=if($null -eq $Result.NextSequence){$null}else{$Result.NextSequence.ToString([Globalization.CultureInfo]::InvariantCulture)};SourceVerified=$Result.SourceVerified;LiveActionsPerformed=$Result.LiveActionsPerformed;RuntimeApproved=$Result.RuntimeApproved}
}
# Helpers define assertions and independent typed literal-data constructors only.
. (Join-Path $PSScriptRoot 'data-helpers.ps1')
try{
    BC-Need (-not $admin -and $record.HostVersion -ceq '7.6.5' -and $record.Core -and $record.X64)
    $record.Stage='FixtureCompile'
    BC-Need ($null -eq ('P2BDescriptorFixture.CounterNote' -as [type]))
    Add-Type -TypeDefinition ([IO.File]::ReadAllText((Join-Path $leaf 'descriptor-fixture.cs'))) -ReferencedAssemblies ([Management.Automation.PSObject].Assembly.Location) -ErrorAction Stop
    $record.CompilePerformed=$true
    $control=[P2BDescriptorFixture.CounterNote]::new('Observable',[int]1)
    [P2BDescriptorFixture.CounterNote]::Reset();$null=$control.MemberType;$null=$control.Value;$null=$control.Copy()
    $record.FixtureObservable=([P2BDescriptorFixture.CounterNote]::MemberReads -gt 0 -and [P2BDescriptorFixture.CounterNote]::ValueReads -gt 0 -and [P2BDescriptorFixture.CounterNote]::Copies -gt 0)
    BC-Need $record.FixtureObservable
    . (Join-Path $leaf 'control-frame.ps1');$record.ProductImported=$true
    $script:fixtures=@(Get-Content -LiteralPath (Join-Path $leaf 'fixtures.json') -Raw|ConvertFrom-Json)
    $number=0
    foreach($variant in @('clean','visible-note','hidden-note','visible-subclass','hidden-subclass')){
        foreach($mode in @('encode-scalar','encode-array','binding-message-scalar','binding-message-array','binding-expected-scalar','binding-expected-array','decode-array')){
            $number++;$record.Stage=$mode+':'+$variant
            $row=[pscustomobject][ordered]@{Number=$number;Mode=$mode;Variant=$variant;Phase='FixtureConstruction';Field=$null;Before=$null;Preconditions=$null;CreatedAndRetainedSame=$null;CountersBefore=$null;Result=$null;CountersAfterAPI=$null;After=$null;CountersAfterInventory=$null;BaseRetained=$null;InputDescriptorRetained=$null;ContractWouldRequireRefusal=($variant -cne 'clean')}
            $record.Rows+=,$row
            $m=BC-Message 2;$e=BC-Expected $m
            if($mode -ceq 'decode-array'){$owner=BC-Object ([ordered]@{Data=(BC-Bytes 2)});$fieldName='Data'}
            elseif($mode -clike 'binding-expected-*'){$owner=$e;$fieldName=if($mode -ceq 'binding-expected-scalar'){'Schema'}else{'BindingDigest'}}
            else{$owner=$m;$fieldName=if($mode -clike '*-scalar'){'OriginTick'}else{'Payload'}}
            $row.Field=$fieldName
            $rootDescriptor=$owner.PSObject.Properties[$fieldName]
            $beforeBase=$rootDescriptor.Value.PSObject.BaseObject
            $tag='P2BScalarProbe'+$number.ToString('D3')
            $extension=$null
            if($variant -cne 'clean'){
                $extension=if($variant -clike '*subclass'){[P2BDescriptorFixture.CounterNote]::new($tag,[int]42)}else{[Management.Automation.PSNoteProperty]::new($tag,[int]42)}
                if($variant -clike 'hidden-*'){
                    [Management.Automation.PSMemberInfo].GetField('<IsHidden>k__BackingField',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly').SetValue($extension.PSObject.BaseObject,$true)
                }
                SP-InsertRaw $rootDescriptor.Value $extension
            }
            $retained=if($variant -ceq 'clean'){$null}else{SP-FindRetained $rootDescriptor.Value $tag}
            if($null -ne $retained){$row.CreatedAndRetainedSame=BC-Ref $extension.PSObject.BaseObject $retained.PSObject.BaseObject}
            [P2BDescriptorFixture.CounterNote]::Reset()
            $row.Phase='PreSnapshot'
            $pre=SP-Snapshot $rootDescriptor.Value $retained $tag
            $preCounts=SP-Counts
            $row.Before=$pre;$row.CountersBefore=$preCounts
            $baseSame=BC-Ref $beforeBase $rootDescriptor.Value.PSObject.BaseObject
            $details=[ordered]@{InventoryCallbacksZero=($preCounts.Total -eq 0);BaseSame=$baseSame;OrdinaryCountValid=$false;TagCountValid=$false;IdentityValid=$false;ClassValid=$false;FlagsValid=$false}
            if($variant -ceq 'clean'){
                $details.OrdinaryCountValid=$pre.NonReservedCount -eq 0
                $details.TagCountValid=$pre.TagCount -eq 0
                $details.IdentityValid=$true;$details.ClassValid=$true;$details.FlagsValid=$true
            }else{
                $wantType=if($variant -clike '*subclass'){'P2BDescriptorFixture.CounterNote'}else{'System.Management.Automation.PSNoteProperty'}
                $details.OrdinaryCountValid=$pre.NonReservedCount -eq 1
                $details.TagCountValid=$pre.TagCount -eq 1
                if($null -ne $pre.Tag){
                    $details.IdentityValid=$pre.Tag.SameDescriptor
                    $details.ClassValid=$pre.Tag.CLRType -ceq $wantType
                    $details.FlagsValid=-not $pre.Tag.Reserved -and $pre.Tag.Hidden -eq ($variant -clike 'hidden-*')
                }
            }
            $row.Preconditions=[pscustomobject]$details
            $row.Phase='PreconditionsSaved'
            foreach($condition in $details.Values){BC-Need ([bool]$condition)}
            $row.Phase='API'
            $record.ProductCalled=$true
            $result=if($mode -ceq 'decode-array'){ConvertFrom-P2BControlFrame $owner.Data}elseif($mode -clike 'encode-*'){ConvertTo-P2BControlFrame $m}else{Test-P2BControlBinding $m $e}
            $apiCounts=SP-Counts
            $summary=SP-Result $result
            $post=SP-Snapshot $rootDescriptor.Value $retained $tag
            $afterCounts=SP-Counts
            $row.Result=$summary;$row.CountersAfterAPI=$apiCounts
            $row.After=$post;$row.CountersAfterInventory=$afterCounts
            $row.BaseRetained=BC-Ref $beforeBase $rootDescriptor.Value.PSObject.BaseObject
            $row.InputDescriptorRetained=BC-Ref $rootDescriptor $owner.PSObject.Properties[$fieldName]
            $row.Phase='PostconditionsSaved'
            BC-Need ($afterCounts.Total -eq $apiCounts.Total -and $row.BaseRetained -and $row.InputDescriptorRetained)
            if($variant -cne 'clean'){BC-Need ($post.TagCount -eq 1 -and $post.Tag.SameDescriptor -and $post.Tag.Hidden -eq $pre.Tag.Hidden)}
            BC-Need (-not $summary.SourceVerified -and -not $summary.LiveActionsPerformed -and -not $summary.RuntimeApproved)
            if($variant -ceq 'clean'){
                $wantReason=if($mode -ceq 'decode-array'){'decoded'}elseif($mode -clike 'encode-*'){'encoded'}else{'binding-matched'}
                BC-Need ($summary.Accepted -and $summary.Reason -ceq $wantReason -and $apiCounts.Total -eq 0)
            }
            if($variant -ceq 'visible-note'){BC-Need (-not $summary.Accepted -and $summary.Reason -ceq 'invalid-input' -and $apiCounts.Total -eq 0)}
            $row.Phase='Completed'
            # Hidden/subclass outcomes are the question, not an assumed oracle.
            # Record all planned independent vectors; no fix or case is retried.
        }
    }
    $record.Stage='Captured';$record.Captured=$true
}catch{
    $caught=$_;$chain=[Collections.Generic.List[object]]::new();$exception=$caught.Exception
    for($i=0;$null -ne $exception -and $i -lt 8;$i++){$chain.Add([pscustomobject]@{Type=$exception.GetType().FullName;Message=$exception.Message});$exception=$exception.InnerException}
    $record.Error=[pscustomobject]@{FullyQualifiedErrorId=$caught.FullyQualifiedErrorId;Chain=$chain.ToArray();Stack=$caught.ScriptStackTrace;Position=$caught.InvocationInfo.PositionMessage}
    $record.Captured=$true
}
$record|ConvertTo-Json -Depth 18
exit 0
