# Evidence-only matrix for the saved scalar-decoration coverage limitation.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$leaf=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../p2b-sdk-structural-compatibility'))
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
$record=[ordered]@{Schema=1;Mode='p2b-scalar-decoration-diagnostic-only';HostVersion=$PSVersionTable.PSVersion.ToString();Core=($PSVersionTable.PSEdition -ceq 'Core');X64=[Environment]::Is64BitProcess;Admin=$admin;Stage='Context';Captured=$false;FixtureObservable=$false;CompilePerformed=$false;ProductImported=$false;ProductCalled=$false;FullSuiteRun=$false;ExpectedRows=35;Rows=@();Error=$null;RuntimeApproved=$false}
function SP-Counts {
    [pscustomobject]@{MemberReads=[P2BDescriptorFixture.CounterNote]::MemberReads;ValueReads=[P2BDescriptorFixture.CounterNote]::ValueReads;Copies=[P2BDescriptorFixture.CounterNote]::Copies;Writes=[P2BDescriptorFixture.CounterNote]::Writes;Total=[P2BDescriptorFixture.CounterNote]::Total}
}
function SP-Snapshot($Value,$Descriptor,[string]$Tag){
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $type=[Management.Automation.PSMemberInfo]
    $nameField=$type.GetField('name',$flags)
    $hiddenField=$type.GetField('<IsHidden>k__BackingField',$flags)
    $reservedField=$type.GetField('<IsReservedMember>k__BackingField',$flags)
    $set=$Value.psextended;$setBase=$set.PSObject.BaseObject
    BC-Need ((BC-Type $setBase) -eq [Management.Automation.PSMemberSet])
    $field=[Management.Automation.PSMemberSet].GetField('internalMembers',$flags)
    $raw=$field.GetValue($setBase)
    BC-Need ((BC-Type $raw) -eq $field.FieldType)
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
            $m=BC-Message 2;$e=BC-Expected $m
            if($mode -ceq 'decode-array'){$owner=BC-Object ([ordered]@{Data=(BC-Bytes 2)});$fieldName='Data'}
            elseif($mode -clike 'binding-expected-*'){$owner=$e;$fieldName=if($mode -ceq 'binding-expected-scalar'){'Schema'}else{'BindingDigest'}}
            else{$owner=$m;$fieldName=if($mode -clike '*-scalar'){'OriginTick'}else{'Payload'}}
            $rootDescriptor=$owner.PSObject.Properties[$fieldName]
            $beforeBase=$rootDescriptor.Value.PSObject.BaseObject
            $tag='P2BScalarProbe'+$number.ToString('D3')
            $extension=$null
            if($variant -cne 'clean'){
                $extension=if($variant -clike '*subclass'){[P2BDescriptorFixture.CounterNote]::new($tag,[int]42)}else{[Management.Automation.PSNoteProperty]::new($tag,[int]42)}
                if($variant -clike 'hidden-*'){
                    [Management.Automation.PSMemberInfo].GetField('<IsHidden>k__BackingField',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly').SetValue($extension.PSObject.BaseObject,$true)
                }
                $members=$owner.($fieldName).PSObject.Properties
                $members.Add($extension)
            }
            [P2BDescriptorFixture.CounterNote]::Reset()
            $pre=SP-Snapshot $rootDescriptor.Value $extension $tag
            $preCounts=SP-Counts
            BC-Need ($preCounts.Total -eq 0 -and (BC-Ref $beforeBase $rootDescriptor.Value.PSObject.BaseObject))
            if($variant -ceq 'clean'){BC-Need ($pre.NonReservedCount -eq 0 -and $pre.TagCount -eq 0)}
            else{
                $wantType=if($variant -clike '*subclass'){'P2BDescriptorFixture.CounterNote'}else{'System.Management.Automation.PSNoteProperty'}
                BC-Need ($pre.NonReservedCount -eq 1 -and $pre.TagCount -eq 1 -and $pre.Tag.SameDescriptor -and $pre.Tag.CLRType -ceq $wantType -and -not $pre.Tag.Reserved -and $pre.Tag.Hidden -eq ($variant -clike 'hidden-*'))
            }
            $record.ProductCalled=$true
            $result=if($mode -ceq 'decode-array'){ConvertFrom-P2BControlFrame $owner.Data}elseif($mode -clike 'encode-*'){ConvertTo-P2BControlFrame $m}else{Test-P2BControlBinding $m $e}
            $apiCounts=SP-Counts
            $summary=SP-Result $result
            $post=SP-Snapshot $rootDescriptor.Value $extension $tag
            $afterCounts=SP-Counts
            $row=[pscustomobject]@{Number=$number;Mode=$mode;Variant=$variant;Field=$fieldName;Before=$pre;CountersBefore=$preCounts;Result=$summary;CountersAfterAPI=$apiCounts;After=$post;CountersAfterInventory=$afterCounts;BaseRetained=(BC-Ref $beforeBase $rootDescriptor.Value.PSObject.BaseObject);InputDescriptorRetained=(BC-Ref $rootDescriptor $owner.PSObject.Properties[$fieldName]);ContractWouldRequireRefusal=($variant -cne 'clean')}
            $record.Rows+=,$row
            BC-Need ($afterCounts.Total -eq $apiCounts.Total -and $row.BaseRetained -and $row.InputDescriptorRetained)
            if($variant -cne 'clean'){BC-Need ($post.TagCount -eq 1 -and $post.Tag.SameDescriptor -and $post.Tag.Hidden -eq $pre.Tag.Hidden)}
            BC-Need (-not $summary.SourceVerified -and -not $summary.LiveActionsPerformed -and -not $summary.RuntimeApproved)
            if($variant -ceq 'clean'){
                $wantReason=if($mode -ceq 'decode-array'){'decoded'}elseif($mode -clike 'encode-*'){'encoded'}else{'binding-matched'}
                BC-Need ($summary.Accepted -and $summary.Reason -ceq $wantReason -and $apiCounts.Total -eq 0)
            }
            if($variant -ceq 'visible-note'){BC-Need (-not $summary.Accepted -and $summary.Reason -ceq 'invalid-input' -and $apiCounts.Total -eq 0)}
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
