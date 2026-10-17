param([Parameter(Mandatory=$true)][ValidateSet('event-decoder','writer-failure')][string]$Group)
$ErrorActionPreference='Stop'
# Coverage extension only: do not change production or the frozen original suite.
. (Join-Path $PSScriptRoot 'synthetic-capture-store.ps1')
function Check-Supplement([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
Check-Supplement ($null -ne (Get-Command Read-OfflineSyntheticCaptureStore -CommandType Function) -and $null -ne (Get-Command Write-OfflineSyntheticCaptureStore -CommandType Function)) 'SyntheticCaptureSupplementRequired'
Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../.superpowers/diagnostics/offline-synthetic-capture-store'))
$null=Get-IsolatedFullPath $root
if(-not (Test-Path -LiteralPath $root)){New-Item -ItemType Directory -Path $root -ErrorAction Stop|Out-Null}
$null=Get-IsolatedFullPath $root
$prefix='supplement-'+[Guid]::NewGuid().ToString('N');$script:supplementNumber=0
function New-SupplementPath{$script:supplementNumber++;Join-Path $root ($prefix+'-'+$script:supplementNumber)}
function Convert-SupplementHex([string]$Hex){$b=New-Object byte[] ($Hex.Length/2);for($i=0;$i -lt $b.Length;$i++){$b[$i]=[Convert]::ToByte($Hex.Substring(2*$i,2),16)};return ,$b}
function Get-SupplementHash([byte[]]$Bytes){$sha=[Security.Cryptography.SHA256]::Create();try{return ,$sha.ComputeHash($Bytes)}finally{$sha.Dispose()}}
function Test-SupplementBytes([byte[]]$A,[byte[]]$B){return [string]::Equals([Convert]::ToBase64String($A),[Convert]::ToBase64String($B),[StringComparison]::Ordinal)}
# Hand-encoded fixed record/lifecycle, not output from the production encoder.
# The binding prefix ends after the empty argument length, before event count.
$bindingPrefix=Convert-SupplementHex '544C524341503031010000000100000001000000730100000078010000006F1C000000323032362D31302D31375430303A30303A30302E303030303030305A1C000000323032362D31302D31375430303A30303A30302E303030303030305A00000000'
$headPrefix=Convert-SupplementHex ('544C5253594E30310100000001000000'+'1D00000073796E7468657469632D636170747572652D636F6E74726163742D7631'+'0900000073796E746865746963'+'01000000730100000078'+'00000000'+'050000007574662D38'+'010000006F'+'1C000000323032362D31302D31375430303A30303A30302E303030303030305A'+'1C000000323032362D31302D31375430303A30303A30302E303030303030305A'+'01')
$lifecyclePrefix=Convert-SupplementHex ('01000000'+'1D0000006F66666C696E652D636170747572652D73696D756C6174696F6E2D7631'+'0A000000696E636F6D706C657465'+'0A000000696E636F6D706C657465')
$lifecycleTail=Convert-SupplementHex ('00'+'0000000000000000'+'000000'+'00'+'0000000000000000'+'00'+'000000')
$recordTail=Convert-SupplementHex ('040000006E6F6E65'+'07000000756E6B6E6F776E'+'00')
function New-SupplementPayload([byte[]]$Events,[int]$Accepted){
    $lifecycle=[byte[]]($lifecyclePrefix+[BitConverter]::GetBytes($Accepted)+$lifecycleTail)
    return ,([byte[]]($headPrefix+$Events+$lifecycle+$recordTail+(Get-SupplementHash ([byte[]]($bindingPrefix+$Events)))+[byte[]]@(0,0,0)))
}
function New-SupplementArchive([byte[]]$Payload){
    $d=New-SupplementPath;New-Item -ItemType Directory -Path $d -ErrorAction Stop|Out-Null
    $marker=[byte[]]((Convert-SupplementHex '544C52535943303101000000')+[BitConverter]::GetBytes([long]$Payload.Length)+(Get-SupplementHash $Payload))
    [IO.File]::WriteAllBytes((Join-Path $d 'record.bin'),$Payload);[IO.File]::WriteAllBytes((Join-Path $d 'complete.bin'),$marker)
    return $d
}
function Assert-SupplementResult($Result,[bool]$Verified,[string]$Reason){
    $names=@('Schema','Verified','Reason','Record','OriginVerified','Acceptance','RuntimeApproved')
    if(@($Result.PSObject.Properties).Count -ne 7 -or @($Result.PSObject.Properties|Where-Object {$_.Name -cnotin $names -or $_.MemberType -ne 'NoteProperty'}).Count -ne 0 -or $Result.Schema -ne 1 -or $Result.Verified -isnot [bool] -or $Result.Verified -ne $Verified -or $Result.Reason -cne $Reason -or $Result.OriginVerified -or $Result.Acceptance -or $Result.RuntimeApproved -or (-not $Verified -and $null -ne $Result.Record)){throw 'Unexpected supplemental archive result'}
}
function Read-SupplementEvents([byte[]]$Bytes,[string]$ExpectedFailure){
    $memory=New-Object IO.MemoryStream(,$Bytes);$reader=New-Object IO.BinaryReader($memory);$caught=$null;$events=$null
    try{try{$events=Read-SyntheticStoreEvents $reader}catch{$caught=$_}
        if($ExpectedFailure.Length -gt 0){if($null -eq $caught -or $caught.Exception.Message -cne $ExpectedFailure){throw 'Expected event rejection branch not observed'}}
        elseif($null -ne $caught -or $memory.Position -ne $memory.Length){throw 'Normal event decoder failed'}
        return ,$events
    }finally{$reader.Dispose();$memory.Dispose()}
}
$emptyEvents=Convert-SupplementHex '00000000'
$minimal=New-SupplementPayload $emptyEvents 0
if($Group -ceq 'event-decoder'){
    # A missing enum/time/total guard must fail these direct-branch assertions;
    # public refusal alone could otherwise be caused by later semantic checks.
    $tick=Convert-SupplementHex '01000000040000000000000000'
    $events=Read-SupplementEvents $tick ''
    $normal=Read-OfflineSyntheticCaptureStore (New-SupplementArchive (New-SupplementPayload $tick 1))
    Assert-SupplementResult $normal $true 'consistent-synthetic-archive-only'
    Check-Supplement ($events.Count -eq 1 -and $events[0].Kind -ceq 'tick' -and $events[0].ElapsedMs -eq 0 -and $normal.Record.Events.Count -eq 1 -and $normal.Record.Events[0].Kind -ceq 'tick' -and $normal.Record.Lifecycle.AcceptedEventCount -eq 1 -and $normal.Record.Lifecycle.Status -ceq 'incomplete') 'IndependentTickArchiveControl'
    $cases=@(
        @{Name='EventKindZeroRefused';Bytes=(Convert-SupplementHex '01000000000000000000000000');Error='Unknown event kind'},
        @{Name='EventKindFiveRefused';Bytes=(Convert-SupplementHex '01000000050000000000000000');Error='Unknown event kind'},
        @{Name='EventChannelZeroRefused';Bytes=(Convert-SupplementHex '0100000002000000000000000000');Error='Unknown event channel'},
        @{Name='EventChannelThreeRefused';Bytes=(Convert-SupplementHex '0100000002000000000000000003');Error='Unknown event channel'},
        @{Name='NegativeEventTimeRefused';Bytes=(Convert-SupplementHex '0100000004FFFFFFFFFFFFFFFF');Error='Negative elapsed time'}
    )
    foreach($case in $cases){
        $null=Read-SupplementEvents $case.Bytes $case.Error
        $r=Read-OfflineSyntheticCaptureStore (New-SupplementArchive (New-SupplementPayload $case.Bytes 0))
        Assert-SupplementResult $r $false 'invalid-payload';Check-Supplement $true $case.Name
    }
    # 33 chunks, each individually legal at65536B; only the aggregate exceeds2MiB.
    $memory=New-Object IO.MemoryStream;$writer=New-Object IO.BinaryWriter($memory)
    try{$writer.Write([int]33);$chunk=New-Object byte[] 65536;for($i=0;$i -lt 33;$i++){$writer.Write([byte]1);$writer.Write([long]0);$writer.Write([byte]1);$writer.Write([int]65536);$writer.Write([byte[]]$chunk)};$writer.Flush();$oversize=$memory.ToArray()}finally{$writer.Dispose();$memory.Dispose()}
    $null=Read-SupplementEvents $oversize 'Whole archive trace exceeds bound'
    $r=Read-OfflineSyntheticCaptureStore (New-SupplementArchive (New-SupplementPayload $oversize 0))
    Assert-SupplementResult $r $false 'invalid-payload';Check-Supplement $true 'CumulativeWireChunksRefused'
    Check-Supplement $true 'SyntheticCaptureSupplementComplete-event-decoder'
}
if($Group -ceq 'writer-failure'){
    function New-SupplementRecord{New-SyntheticCaptureRecord ([pscustomobject]@{Schema=1;SourceMode='synthetic';SessionLabel='s';ExecutableDeclaration='x';ArgumentBytes=[byte[]]@();EncodingName='utf-8';OriginDeclaration='o';StartedUTC='2026-10-17T00:00:00.0000000Z';EndedUTC='2026-10-17T00:00:00.0000000Z';Events=[object[]]@()})}
    function Invoke-SupplementCollision([string]$Mode){
        $d=New-SupplementPath
        $saved=(Get-Command Write-SyntheticStoreNewFile -CommandType Function).ScriptBlock
        $state=[pscustomobject]@{Directory=$d;Mode=$Mode;Injected=$false;Primary=$null;Secondary=$null}
        # Local function shadow only. Its scope unwinds even on exceptions;
        # the script-scope production definition is never overwritten.
        function Write-SyntheticStoreNewFile([string]$Path,[byte[]]$Bytes){
            $leaf=[IO.Path]::GetFileName($Path);$target=if($state.Mode -ceq 'record'){'record.bin'}else{'complete.bin'}
            if(-not $state.Injected -and $leaf -ceq $target){
                if(-not [string]::Equals([IO.Path]::GetDirectoryName($Path),$state.Directory,[StringComparison]::OrdinalIgnoreCase) -or -not (Test-Path -LiteralPath $state.Directory -PathType Container)){throw 'Injection scope mismatch'}
                [IO.File]::WriteAllBytes($Path,[byte[]]@(67,79,76))
                if($state.Mode -ceq 'marker-and-stopped'){$null=[IO.Directory]::CreateDirectory((Join-Path $state.Directory 'stopped'))}
                $state.Injected=$true
            }
            # Actual CreateNew, write, flush and native exception remain real.
            try{& $saved $Path $Bytes}catch{if($leaf -ceq 'stopped'){$state.Secondary=$_}else{$state.Primary=$_};throw}
        }
        $caught=$null
        try{$null=Write-OfflineSyntheticCaptureStore $d (New-SupplementRecord)}catch{$caught=$_}
        if(-not $state.Injected -or $null -eq $caught -or $null -eq $state.Primary){throw 'Writer failure path not exercised'}
        if($caught.Exception.Message -cne $state.Primary.Exception.Message -or $caught.FullyQualifiedErrorId -cne $state.Primary.FullyQualifiedErrorId -or $caught.Exception.GetBaseException().HResult -ne $state.Primary.Exception.GetBaseException().HResult){throw 'Primary writer error was replaced'}
        $collision=if($Mode -ceq 'record'){'record.bin'}else{'complete.bin'}
        if(-not $caught.Exception.Message.Contains($collision)){throw 'Wrong original failure captured'}
        if(-not (Test-SupplementBytes ([IO.File]::ReadAllBytes((Join-Path $d $collision))) ([byte[]]@(67,79,76)))){throw 'Collision artifact was altered'}
        if($Mode -ceq 'record'){
            if(Test-Path -LiteralPath (Join-Path $d 'complete.bin')){throw 'Marker published after record failure'}
        }elseif(-not (Test-SupplementBytes ([IO.File]::ReadAllBytes((Join-Path $d 'record.bin'))) $minimal)){throw 'Written payload not retained'}
        if($Mode -ceq 'marker-and-stopped'){
            if($null -eq $state.Secondary -or -not $state.Secondary.Exception.Message.Contains('stopped') -or -not (Test-Path -LiteralPath (Join-Path $d 'stopped') -PathType Container)){throw 'Sentinel-write failure not exercised'}
        }else{
            if($null -ne $state.Secondary -or -not (Test-SupplementBytes ([IO.File]::ReadAllBytes((Join-Path $d 'stopped'))) ([byte[]]@(1)))){throw 'Failure sentinel not published'}
        }
        $r=Read-OfflineSyntheticCaptureStore $d;Assert-SupplementResult $r $false 'incomplete-archive'
    }
    $original=(Get-Command Write-SyntheticStoreNewFile -CommandType Function).ScriptBlock
    foreach($mode in @('record','marker','marker-and-stopped')){
        Invoke-SupplementCollision $mode
        if(-not [object]::ReferenceEquals($original,(Get-Command Write-SyntheticStoreNewFile -CommandType Function).ScriptBlock)){throw 'Local fault injector escaped scope'}
        Check-Supplement $true ('WriterCollisionPreservesPrimary-'+$mode)
    }
    $d=New-SupplementPath;$r=Write-OfflineSyntheticCaptureStore $d (New-SupplementRecord)
    Assert-SupplementResult $r $true 'consistent-synthetic-archive-only'
    $r=Read-OfflineSyntheticCaptureStore $d;Assert-SupplementResult $r $true 'consistent-synthetic-archive-only'
    Check-Supplement ((Test-SupplementBytes ([IO.File]::ReadAllBytes((Join-Path $d 'record.bin'))) $minimal) -and -not (Test-Path -LiteralPath (Join-Path $d 'stopped'))) 'WriterNormalControlAfterCollisions'
    Check-Supplement $true 'SyntheticCaptureSupplementComplete-writer-failure'
}
Write-Output 'LIMIT: supplemental synthetic coverage only;no live actions,security changes or acceptance.'
