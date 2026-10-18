# One evidence-only reproduction of SequenceExact; no full-suite replay or fix.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$leaf=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../p2b-passive-descriptor-repair'))
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
$record=[ordered]@{Schema=1;Mode='p2b-sequence-binding-diagnostic-only';HostVersion=$PSVersionTable.PSVersion.ToString();Core=($PSVersionTable.PSEdition -ceq 'Core');X64=[Environment]::Is64BitProcess;Admin=$admin;Stage='Context';Captured=$false;ProductImported=$false;ProductCalled=$false;CompilePerformed=$false;FullSuiteRun=$false;Checkpoints=@();FinalResult=$null;FinalEqual=$null;Inventories=@();Error=$null;RuntimeApproved=$false}
function SQ-Inventory($Value,[string]$Label){
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $memberType=[Management.Automation.PSMemberInfo]
    $fields=@('name','<IsReservedMember>k__BackingField','<IsHidden>k__BackingField','<IsInstance>k__BackingField')|ForEach-Object{$memberType.GetField($_,$flags)}
    foreach($field in $fields){if($null -eq $field -or $field.DeclaringType -ne $memberType){throw 'Fixed SDK metadata unavailable'}}
    $shim=$Value.PSObject;$extended=$Value.psextended
    $shimBase=$shim.PSObject.BaseObject;$setBase=$extended.PSObject.BaseObject
    $field=[Management.Automation.PSMemberSet].GetField('internalMembers',$flags)
    if((BC-Type $setBase) -ne [Management.Automation.PSMemberSet]){throw 'Unexpected SDK receiver'}
    $raw=$field.GetValue($setBase)
    if((BC-Type $raw) -ne $field.FieldType){throw 'Unexpected raw collection'}
    $items=[Collections.Generic.List[object]]::new()
    foreach($descriptor in $raw){
        $base=$descriptor.PSObject.BaseObject
        $items.Add([pscustomobject]@{Name=$fields[0].GetValue($base);CLRType=(BC-Type $descriptor).FullName;Reserved=$fields[1].GetValue($base);Hidden=$fields[2].GetValue($base);Instance=$fields[3].GetValue($base);SameAsPSObject=[object]::ReferenceEquals($base,$shimBase);SameAsPSExtended=[object]::ReferenceEquals($base,$setBase)})
    }
    [pscustomobject]@{Label=$Label;SetType=(BC-Type $setBase).FullName;RawType=(BC-Type $raw).FullName;MemberCount=$items.Count;MemberValuesRead=$false;Members=$items.ToArray()}
}
try{
    if($admin -or $record.HostVersion -cne '7.6.5' -or -not $record.Core -or -not $record.X64){throw 'Ordinary pinned PS7 host required'}
    . (Join-Path $PSScriptRoot 'sequence-helpers.ps1')
    . (Join-Path $leaf 'control-frame.ps1');$record.ProductImported=$true
    $script:fixtures=@(Get-Content -LiteralPath (Join-Path $leaf 'fixtures.json') -Raw|ConvertFrom-Json)
    # Preserve the original SequenceExact operation order, not the preceding groups.
    $record.Stage='LowBinding';$record.ProductCalled=$true
    $m=BC-Message 2;$e=BC-Expected $m;BC-Bind $m $e
    $record.Checkpoints+=,'LowBindingMatched'
    foreach($seq in @([uint64]6,[uint64]8)){$a=BC-Message 2;BC-Put $a Sequence $seq;BC-Bind $a $e sequence-mismatch}
    $record.Checkpoints+=,'LowSequenceMismatches'
    $record.Stage='HighEncode'
    $m=BC-Message 2;BC-Put $m Sequence ([uint64]::Parse('9223372036854775808'))
    $b=BC-Bytes 2;BC-UInt $b 64 8 ([uint64]::Parse('9223372036854775808'))
    BC-Need (BC-Equal (ConvertTo-P2BControlFrame $m) (BC-Result encoded -Bytes $b))
    $record.Checkpoints+=,'HighEncodeEqual'
    $record.Stage='HighDecode'
    BC-Need (BC-Equal (ConvertFrom-P2BControlFrame $b) (BC-Result decoded -Message $m))
    $record.Checkpoints+=,'HighDecodeEqual'
    $record.Stage='HighBinding'
    $e=BC-Expected $m
    $want=BC-Result binding-matched -Next ([uint64]::Parse('9223372036854775809'))
    $actual=Test-P2BControlBinding $m $e
    # Snapshot only the module's owned result; input inventory is deferred until
    # AFTER the actual binding call, so inspection cannot prime that call.
    $record.FinalResult=[pscustomobject]@{Schema=$actual.Schema;Mode=$actual.Mode;Accepted=$actual.Accepted;Reason=$actual.Reason;BytesNull=($null -eq $actual.Bytes);MessageNull=($null -eq $actual.Message);NextSequence=$actual.NextSequence;NextType=if($null -eq $actual.NextSequence){$null}else{(BC-Type $actual.NextSequence).FullName};SourceVerified=$actual.SourceVerified;LiveActionsPerformed=$actual.LiveActionsPerformed;RuntimeApproved=$actual.RuntimeApproved}
    $record.FinalEqual=[bool](BC-Equal $actual $want)
    $record.Stage='PostBindingInventory'
    $record.Inventories=@((SQ-Inventory $m 'Message-after-high-binding'),(SQ-Inventory $e 'Expected-after-high-binding'))
    $record.Stage='Captured';$record.Captured=$true
}catch{
    $caught=$_;$chain=[Collections.Generic.List[object]]::new();$exception=$caught.Exception
    for($i=0;$null -ne $exception -and $i -lt 8;$i++){$chain.Add([pscustomobject]@{Type=$exception.GetType().FullName;Message=$exception.Message});$exception=$exception.InnerException}
    $record.Error=[pscustomobject]@{FullyQualifiedErrorId=$caught.FullyQualifiedErrorId;Chain=$chain.ToArray();Stack=$caught.ScriptStackTrace;Position=$caught.InvocationInfo.PositionMessage}
    $record.Captured=$true
}
$record|ConvertTo-Json -Depth 16
exit 0
