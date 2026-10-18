param([ValidateSet('red','all')][string]$Group='all',[string]$TracePath)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$id=[Security.Principal.WindowsIdentity]::GetCurrent()
$admin=([Security.Principal.WindowsPrincipal]::new($id)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
[Console]::Out.WriteLine('CONTEXT: '+$id.Name+' admin='+$admin)
[Console]::Out.WriteLine('HOST: PS'+$PSVersionTable.PSVersion.ToString()+' Core x64')
. (Join-Path $PSScriptRoot 'test-helpers.ps1')
$script:trRecord=[pscustomobject]@{Schema=1;Mode='p2b-tree-repair-tests';Group=$Group;Groups=@();Representation=$null;FirstActual=$null;Error=$null;Passed=$false;RuntimeApproved=$false}
$script:trGroup='FirstSnapshot';$script:ibGroup='TreeGuard'
function TR-Need([bool]$ok){if(-not $ok){throw ('ASSERTION FAILED: P2BTree'+$script:trGroup)}}
function TR-Same($a,$b){return [bool][object].GetMethod('ReferenceEquals',[type[]]@([object],[object])).Invoke($null,[object[]]@($a,$b))}
function TR-Case([string]$Name,[scriptblock]$Body){if($Group -ceq 'red' -and $Name -cne 'FirstSnapshot'){return};$script:trGroup=$Name;& $Body;$script:trRecord.Groups+=,$Name;[Console]::Out.WriteLine('PASS: P2BTree'+$Name)}
function TR-New {$r=New-P2BInbox (IB-Binding);$null=IB-Result $r created;return $r.Inbox}
function TR-Snapshot($t){$r=Get-P2BInboxSnapshot $t;TR-Need ($r.Accepted -and $r.Reason -ceq 'inspected');$null=IB-Result $r inspected;return $r.Snapshot}
function TR-Initial($t){TR-Need (IB-Eq (TR-Snapshot $t) (IB-Expected (IB-Initial)))}
try{
    $script:trModule=Import-Module (Join-Path $PSScriptRoot 'inbox.psm1') -PassThru -DisableNameChecking
    $script:ibModule=$script:trModule
    $script:ibLiterals=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'fixtures.json') -Raw|ConvertFrom-Json
    TR-Case FirstSnapshot {
        # Independent CLR reflection; not the candidate type/classifier helper.
        $script:trRecord.Representation=& $script:trModule {
            $owner=[pscustomobject][ordered]@{Label='probe';Values=[object[]]::new(0)}
            $getType=[object].GetMethod('GetType',[type[]]@());$invoke=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
            $argsLocal=[object[]]::new(2);$argsLocal[0]=$owner;$argsLocal[1]=$null
            $rawType=$invoke.Invoke($getType,$argsLocal)
            $field=[Management.Automation.PSObject].GetField('_immediateBaseObject',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly')
            $fieldReader=[Reflection.FieldInfo].GetMethod('GetValue',[type[]]@([object]))
            $ownerArg=[object[]]::new(1);$ownerArg[0]=$owner
            $base=$fieldReader.Invoke($field,$ownerArg)
            $argsLocal[0]=$base;$baseType=$invoke.Invoke($getType,$argsLocal)
            [pscustomobject]@{RawType=$rawType.FullName;BaseType=$baseType.FullName;Fields=[string[]]@($owner.PSObject.Properties.Name)}
        }
        TR-Need ($script:trRecord.Representation.RawType -ceq 'System.Management.Automation.PSObject' -and $script:trRecord.Representation.BaseType -ceq 'System.Management.Automation.PSCustomObject' -and (IB-Eq $script:trRecord.Representation.Fields ([string[]]@('Label','Values'))))
        $b=IB-Binding;$r=New-P2BInbox $b;$null=IB-Result $r created;$t=$r.Inbox
        $b.BindingDigest[0]=99;$b.SessionId[0]=99;$b.Frequency=[long]1001
        $r=Get-P2BInboxSnapshot $t;$script:trRecord.FirstActual=$r
        if($Group -ceq 'red'){
            if($r.Accepted -or $r.Reason -cne 'internal-error' -or $null -ne $r.Snapshot){throw 'Unexpected baseline snapshot result'}
            $null=IB-Result $r internal-error
        }
        TR-Need ($r.Accepted -and $r.Reason -ceq 'inspected')
        $null=IB-Result $r inspected;TR-Need (IB-Eq $r.Snapshot (IB-Expected (IB-Initial)))
    }
    TR-Case OwnedObjectTree {
        $pair=& $script:trModule {$source=[pscustomobject][ordered]@{Label='probe';Number=[int]3};$a=[object[]]::new(2);$a[0]=$source;$a[1]=Copy-IBTree $source;return ,$a}
        TR-Need ((IB-Eq $pair[1] ([pscustomobject][ordered]@{Label='probe';Number=[int]3})) -and -not (TR-Same $pair[0] $pair[1]))
    }
    TR-Case OwnedArrayTree {
        $pair=& $script:trModule {$source=[object[]]@([byte[]]@(7,9),[uint64]::Parse('9223372036854775808'));$a=[object[]]::new(2);$a[0]=$source;$a[1]=Copy-IBTree $source;return ,$a}
        $expected=[object[]]@([byte[]]@(7,9),[uint64]::Parse('9223372036854775808'))
        TR-Need ((IB-Eq $pair[1] $expected) -and -not (TR-Same $pair[0] $pair[1]) -and -not (TR-Same $pair[0][0] $pair[1][0]))
    }
    TR-Case EmptyArrays {
        foreach($type in @([object],[byte])){
            $packet=& $script:trModule {param($elementType)$src=[Array]::CreateInstance($elementType,0);$a=[object[]]::new(3);$a[0]=$src;$a[1]=Copy-IBTree $src;$a[2]=Copy-IBTree $src;return ,$a} $type
            TR-Need (-not (TR-Same $packet[0] $packet[1]) -and -not (TR-Same $packet[0] $packet[2]) -and -not (TR-Same $packet[1] $packet[2]))
            $held=$packet[1];$held.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Length',[int]99));$held.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Rank',[int]99))
            TR-Need ($held.Length -eq 99 -and $held.Rank -eq 99 -and $packet[0].GetLength(0) -eq 0 -and $packet[2].GetLength(0) -eq 0 -and $packet[2].Rank -eq 1 -and $packet[2].PSObject.Properties['Rank'].MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty)
        }
    }
    TR-Case NestedValues {
        $pair=& $script:trModule {$src=[pscustomobject][ordered]@{Label='probe';Child=[pscustomobject][ordered]@{Enabled=$true;Bytes=[byte[]]@(7,9)};Items=[object[]]@([int]3,'x')};$a=[object[]]::new(2);$a[0]=$src;$a[1]=Copy-IBTree $src;return ,$a}
        $want=[pscustomobject][ordered]@{Label='probe';Child=[pscustomobject][ordered]@{Enabled=$true;Bytes=[byte[]]@(7,9)};Items=[object[]]@([int]3,'x')}
        TR-Need ((IB-Eq $pair[1] $want) -and -not (TR-Same $pair[0].Child $pair[1].Child) -and -not (TR-Same $pair[0].Items $pair[1].Items))
    }
    TR-Case ScalarLeaves {
        $rows=& $script:trModule {
            $leaves=[object[]]@($null,[byte]7,[uint16]258,[uint32]65537,[uint64]::Parse('9223372036854775808'),[int]-7,[long]-9,$true,[string]::new([char[]]@('p','r','o','b','e')),[byte[]]@(7,9))
            $result=[object[]]::new($leaves.Length)
            for($i=0;$i -lt $leaves.Length;$i++){$row=[object[]]::new(2);$row[0]=$leaves[$i];$row[1]=Copy-IBTree $leaves[$i];$result[$i]=$row};return ,$result
        }
        $expected=[object[]]@($null,[byte]7,[uint16]258,[uint32]65537,[uint64]::Parse('9223372036854775808'),[int]-7,[long]-9,$true,'probe',[byte[]]@(7,9))
        for($i=0;$i -lt 10;$i++){TR-Need ((IB-Eq $rows[$i][1] $expected[$i]) -and ($i -eq 0 -or -not (TR-Same $rows[$i][0] $rows[$i][1])))}
    }
    TR-Case UnsupportedLeaves {
        $rejections=& $script:trModule {$result=[bool[]]::new(3);$values=[object[]]@([object]::new(),[version]::new(1,2),@{Key=1});for($i=0;$i -lt 3;$i++){try{$null=Copy-IBTree $values[$i]}catch{$result[$i]=$true}};return ,$result}
        TR-Need (IB-Eq $rejections ([bool[]]@($true,$true,$true)))
    }
    TR-Case ReturnedMutation {
        $t=TR-New;$f=IB-Frame;$null=IB-Result (Add-P2BInboxFragment $t ([byte]1) $f) queued ([byte]1) 256
        $s=TR-Snapshot $t;$s.Binding.BindingDigest[0]=99;$s.Endpoints[0].Queue[0][0]=99
        $s.Endpoints[0].PartialBytes.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Rank',[int]99))
        $s.Endpoints[0].Count.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Probe',[int]99))
        TR-Need ($s.Endpoints[0].Queue[0][0] -eq 99 -and $s.Endpoints[0].PartialBytes.Rank -eq 99)
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 2 1 ([object[]]@(,$f));TR-Need (IB-Eq (TR-Snapshot $t) (IB-Expected $ep))
    }
    TR-Case TwoSnapshots {
        $t=TR-New;$a=TR-Snapshot $t;$b=TR-Snapshot $t
        TR-Need (-not (TR-Same $a $b) -and -not (TR-Same $a.Endpoints $b.Endpoints) -and -not (TR-Same $a.Endpoints[0] $b.Endpoints[0]) -and -not (TR-Same $a.Endpoints[0].Queue $b.Endpoints[0].Queue))
        $a.Binding.SessionId[0]=99;$a.Endpoints[0].Count.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Probe',[int]99));TR-Initial $t
        TR-Need (IB-Eq $b (IB-Expected (IB-Initial)))
    }
    TR-Case TwoInboxes {
        $a=TR-New;$b=TR-New;TR-Need (-not (TR-Same $a $b))
        $s=TR-Snapshot $a;$s.Binding.Schema=[int]99;$null=IB-Result (Close-P2BInboxInput $a ([byte]1)) peer-closed ([byte]1);TR-Initial $b
    }
    TR-Case StoppedSnapshot {
        $t=TR-New;$f=IB-Frame;$null=IB-Result (Add-P2BInboxFragment $t ([byte]1) $f) queued ([byte]1) 256
        $null=IB-Result (Close-P2BInboxInput $t ([byte]2)) peer-closed ([byte]2)
        $ep=IB-Initial;$ep[0]=IB-Endpoint 1 2 1 ([object[]]@(,$f))
        TR-Need (IB-Eq (TR-Snapshot $t) (IB-Expected $ep (IB-Failure peer-closed ([byte]2) eof)))
        $null=IB-Result (Add-P2BInboxFragment $t ([byte]1) $f) stopped;$null=IB-Result (Receive-P2BInboxFrame $t ([byte]1)) stopped
    }
    TR-Case Complete {
        TR-Need ($script:trRecord.Groups.Count -eq 11 -and (IB-Eq ([string[]]@($script:trModule.ExportedFunctions.Keys|Sort-Object)) ([string[]]@($script:ibNames|Sort-Object))))
    }
    $script:trRecord.Passed=$true
}catch{
    $script:trRecord.Error=[pscustomobject]@{Message=$_.Exception.Message;Type=$_.Exception.GetType().FullName;Position=$_.InvocationInfo.PositionMessage;Stack=$_.ScriptStackTrace}
    [Console]::Out.WriteLine($_.Exception.Message)
}finally{if(-not [string]::IsNullOrEmpty($TracePath)){[IO.File]::WriteAllText($TracePath,($script:trRecord|ConvertTo-Json -Depth 30),[Text.UTF8Encoding]::new($false))}}
if(-not $script:trRecord.Passed){exit 1}
