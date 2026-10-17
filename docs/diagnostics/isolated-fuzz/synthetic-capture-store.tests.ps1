param([ValidateSet('format','boundary','all','admission','byte-bound','event-bound','reader-paths')][string]$Group='all')
$ErrorActionPreference='Stop'
function Check-CaptureStore([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
$module=Join-Path $PSScriptRoot 'synthetic-capture-store.ps1'
if(Test-Path -LiteralPath $module){. $module}
Check-CaptureStore ($null -ne (Get-Command Write-OfflineSyntheticCaptureStore -CommandType Function -ErrorAction SilentlyContinue) -and $null -ne (Get-Command Read-OfflineSyntheticCaptureStore -CommandType Function -ErrorAction SilentlyContinue)) 'SyntheticCaptureStoreRequired'
Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../.superpowers/diagnostics/offline-synthetic-capture-store'))
$prefix='synthetic-'+[Guid]::NewGuid().ToString('N');$script:fixtureNumber=0;$script:fixtureCallbacks=0
function Fixture-Path{$script:fixtureNumber++;Join-Path $root ($prefix+'-'+$script:fixtureNumber)}
function Fixture-Packet{[pscustomobject]@{Schema=1;SourceMode='synthetic';SessionLabel='s';ExecutableDeclaration='x';ArgumentBytes=[byte[]]@();EncodingName='utf-8';OriginDeclaration='o';StartedUTC='2026-10-17T00:00:00.0000000Z';EndedUTC='2026-10-17T00:00:00.0000000Z';Events=[object[]]@()}}
function Fixture-Chunk([string]$Channel,[byte[]]$Bytes,[long]$Time=0){[pscustomobject]@{Kind='chunk';ElapsedMs=$Time;Channel=$Channel;Bytes=$Bytes}}
function Fixture-EOF([string]$Channel){[pscustomobject]@{Kind='eof';ElapsedMs=[long]0;Channel=$Channel}}
function Fixture-Exit([int]$Code=0){[pscustomobject]@{Kind='exit';ElapsedMs=[long]0;ExitCode=$Code}}
function Fixture-Tick([long]$Time=0){[pscustomobject]@{Kind='tick';ElapsedMs=$Time}}
function Fixture-Record{New-SyntheticCaptureRecord (Fixture-Packet)}
function Fixture-Hex([string]$Hex){$b=New-Object byte[] ($Hex.Length/2);for($i=0;$i -lt $b.Length;$i++){$b[$i]=[Convert]::ToByte($Hex.Substring(2*$i,2),16)};return ,$b}
function Fixture-Hash([byte[]]$Bytes){$sha=[Security.Cryptography.SHA256]::Create();try{return ,$sha.ComputeHash($Bytes)}finally{$sha.Dispose()}}
function Fixture-Equal($Left,$Right){[string]::Equals($Left,$Right,[StringComparison]::Ordinal)}
# Independent hand-encoded minimal record;no production encoder/hash helper.
# Header/schema,then literal length-prefixed ASCII,empty snapshots/events,
# incomplete lifecycle,null index/exit,false flags,none/unknown and binding digest.
$bindingLiteral=Fixture-Hex '544C524341503031010000000100000001000000730100000078010000006F1C000000323032362D31302D31375430303A30303A30302E303030303030305A1C000000323032362D31302D31375430303A30303A30302E303030303030305A0000000000000000'
$baseHead=Fixture-Hex ('544C5253594E30310100000001000000'+'1D00000073796E7468657469632D636170747572652D636F6E74726163742D7631'+'0900000073796E746865746963'+'01000000730100000078'+'00000000'+'050000007574662D38'+'010000006F'+'1C000000323032362D31302D31375430303A30303A30302E303030303030305A'+'1C000000323032362D31302D31375430303A30303A30302E303030303030305A'+'01'+'00000000')
$baseLifecycle=Fixture-Hex ('01000000'+'1D0000006F66666C696E652D636170747572652D73696D756C6174696F6E2D7631'+'0A000000696E636F6D706C657465'+'0A000000696E636F6D706C657465'+'00000000'+'00'+'0000000000000000'+'000000'+'00'+'0000000000000000'+'00'+'000000')
$baseTail=Fixture-Hex ('040000006E6F6E65'+'07000000756E6B6E6F776E'+'00')
$literal=[byte[]]($baseHead+$baseLifecycle+$baseTail+(Fixture-Hash $bindingLiteral)+[byte[]]@(0,0,0))
function Fixture-Marker([byte[]]$Payload){return ,([byte[]]((Fixture-Hex '544C52535943303101000000')+[BitConverter]::GetBytes([long]$Payload.Length)+(Fixture-Hash $Payload)))}
function Fixture-Put([byte[]]$Payload,[byte[]]$Marker){$d=Fixture-Path;New-Item -ItemType Directory -Path $d -ErrorAction Stop|Out-Null;[IO.File]::WriteAllBytes((Join-Path $d 'record.bin'),$Payload);[IO.File]::WriteAllBytes((Join-Path $d 'complete.bin'),$Marker);return $d}
function Fixture-Result($R){
    $fields=@('Schema','Verified','Reason','Record','OriginVerified','Acceptance','RuntimeApproved')
    if(@($R.PSObject.Properties).Count -ne 7 -or @($R.PSObject.Properties|Where-Object{$_.Name -cnotin $fields -or $_.MemberType -ne 'NoteProperty'}).Count -ne 0 -or $R.Schema -ne 1 -or $R.Verified -isnot [bool] -or $R.OriginVerified -or $R.Acceptance -or $R.RuntimeApproved){throw 'Bad archive result'}
    if($R.Verified){if(-not (Fixture-Equal $R.Reason 'consistent-synthetic-archive-only') -or @($R.Record.PSObject.Properties).Count -ne 20){throw 'Bad successful record'}}
    elseif($null -ne $R.Record -or $R.Reason -cnotin @('invalid-path','missing-archive','incomplete-archive','unsupported-format','invalid-payload','integrity-mismatch','invalid-record','io-error')){throw 'Failed archive leaked record/reason'}
}
function Fixture-Write($Record){$d=Fixture-Path;$r=Write-OfflineSyntheticCaptureStore -Directory $d -Record $Record;Fixture-Result $r;if(-not $r.Verified){throw 'Fixture publication failed'};return [pscustomobject]@{Directory=$d;Result=$r}}
function Fixture-Refuse($Record,[string]$Name){$d=Fixture-Path;$before=$script:fixtureCallbacks;$refused=$false;try{$null=Write-OfflineSyntheticCaptureStore -Directory $d -Record $Record}catch{$refused=$true};Check-CaptureStore ($refused -and -not (Test-Path -LiteralPath $d) -and $script:fixtureCallbacks -eq $before) $Name}
function Fixture-ReadRefuse([string]$Directory,[string]$Name){$r=Read-OfflineSyntheticCaptureStore $Directory;Fixture-Result $r;Check-CaptureStore (-not $r.Verified -and $null -eq $r.Record) $Name}
if($Group -in @('format','all')){
    # Missing-root branch runs only when genuinely absent;never delete an old root.
    $existed=Test-Path -LiteralPath $root;Fixture-Refuse $null 'PreflightPreservesRootState';if((Test-Path -LiteralPath $root) -ne $existed){throw 'Invalid record created root'}
    Write-Output ('ROOT_PRECONDITION: existed='+$existed)
    $minimal=Fixture-Record;$published=Fixture-Write $minimal
    $actualBytes=[IO.File]::ReadAllBytes((Join-Path $published.Directory 'record.bin'));$actualMarker=[IO.File]::ReadAllBytes((Join-Path $published.Directory 'complete.bin'))
    $samePayload=Fixture-Equal ([Convert]::ToBase64String($actualBytes)) ([Convert]::ToBase64String($literal))
    $sameMarker=Fixture-Equal ([Convert]::ToBase64String($actualMarker)) ([Convert]::ToBase64String((Fixture-Marker $literal)))
    Check-CaptureStore ($samePayload -and $sameMarker) 'IndependentWireLiteral'
    $fromLiteral=Read-OfflineSyntheticCaptureStore (Fixture-Put $literal (Fixture-Marker $literal));Fixture-Result $fromLiteral
    Check-CaptureStore ($fromLiteral.Verified -and (Fixture-Equal $fromLiteral.Record.SessionLabel 's') -and (Fixture-Equal $fromLiteral.Record.Lifecycle.Status 'incomplete') -and $null -eq $fromLiteral.Record.Lifecycle.Package) 'TwentyFieldsRoundtrip'
    foreach($status in @('incomplete','complete','timed-out','refused')){
        $p=Fixture-Packet
        switch($status){complete{$p.Events=[object[]]@((Fixture-Exit),(Fixture-EOF 'stdout'),(Fixture-EOF 'stderr'))} timed-out{$p.Events=[object[]]@((Fixture-Tick 120000))} refused{$p.Events=[object[]]@((Fixture-EOF 'stdout'),(Fixture-EOF 'stdout'))}}
        $result=(Fixture-Write (New-SyntheticCaptureRecord $p)).Result.Record
        Check-CaptureStore ((Fixture-Equal $result.Lifecycle.Status $status) -and $result.CleanupStatus -ceq 'unknown' -and -not $result.LiveActionsPerformed -and -not $result.OriginVerified -and -not $result.Acceptance -and -not $result.RuntimeApproved) ('FourLifecycleRoundtrips-'+$status)
        if($status -eq 'complete'){Check-CaptureStore ($result.Lifecycle.ExitCode -eq 0 -and $null -eq $result.Lifecycle.FailureEventIndex -and $null -ne $result.Lifecycle.Package) 'NullableExitAndIndexDistinct-complete'}
        if($status -eq 'timed-out'){Check-CaptureStore ($null -eq $result.Lifecycle.ExitCode -and $result.Lifecycle.FailureEventIndex -eq 0 -and $null -eq $result.Lifecycle.Package) 'NullableExitAndIndexDistinct-timeout'}
    }
    $p=Fixture-Packet;$p.Events=[object[]]@((Fixture-Tick ([long]::MaxValue)));$r=(Fixture-Write (New-SyntheticCaptureRecord $p)).Result.Record
    Check-CaptureStore ($r.Events[0].ElapsedMs -is [long] -and $r.Events[0].ElapsedMs -eq [long]::MaxValue -and $r.Lifecycle.LastObservedElapsedMs -eq [long]::MaxValue -and $r.Lifecycle.Status -ceq 'timed-out') 'ExactInt64TimeRetained'
    $p=Fixture-Packet;$p.Events=[object[]]@((Fixture-Chunk 'stdout' ([byte[]]@(97))),(Fixture-Tick 120000),(Fixture-Chunk 'stderr' ([byte[]]@(0,255,13,10)) 120001));$r=(Fixture-Write (New-SyntheticCaptureRecord $p)).Result.Record
    Check-CaptureStore ($r.Events.Count -eq 3 -and $r.Events[2].Data.Base64 -ceq 'AP8NCg==' -and $r.Lifecycle.Stdout.Base64 -ceq 'YQ==' -and $r.Lifecycle.Stderr.Base64 -ceq '' -and $r.Lifecycle.Status -ceq 'timed-out') 'RejectedLateTraceRetained'
    foreach($enc in @('utf-8','utf-16le')){
        $p=Fixture-Packet;$p.EncodingName=$enc;$p.ArgumentBytes=[byte[]]@(65,0,45,78);$p.OriginDeclaration=([string][char]0xE9)+'e'+[char]0x301+[char]0x6A21
        $p.Events=[object[]]@((Fixture-Chunk 'stdout' ([byte[]]@(97,98,99))),(Fixture-Chunk 'stderr' ([byte[]]@(0,255,13,10))),(Fixture-Exit -1),(Fixture-EOF 'stdout'),(Fixture-EOF 'stderr'))
        $r=(Fixture-Write (New-SyntheticCaptureRecord $p)).Result.Record
        Check-CaptureStore ((Fixture-Equal $r.OriginDeclaration $p.OriginDeclaration) -and $r.EncodingName -ceq $enc -and $r.Arguments.Base64 -ceq 'QQAtTg==' -and $r.Lifecycle.Stdout.Base64 -ceq 'YWJj' -and $r.Lifecycle.Stderr.Base64 -ceq 'AP8NCg==' -and $r.Lifecycle.Package.Classification -ceq 'refused' -and $r.Lifecycle.ExitCode -eq -1 -and -not $r.Lifecycle.Package.Acceptance) ('Utf8Utf16UnicodeAndBytes-'+$enc)
    }
    $a=(Read-OfflineSyntheticCaptureStore $published.Directory).Record;$b=(Read-OfflineSyntheticCaptureStore $published.Directory).Record
    Add-Member -InputObject $a.Events -MemberType NoteProperty -Name Length -Value 7 -Force
    Check-CaptureStore ($b.Events.Length -eq 0 -and (Read-OfflineSyntheticCaptureStore $published.Directory).Record.Events.Length -eq 0 -and -not [object]::ReferenceEquals($a.Events,$b.Events)) 'IndependentEmptyEventOwnership'
    $a.Events.PSObject.Properties.Remove('Length');$minimal.SessionLabel='mutated'
    Check-CaptureStore ((Read-OfflineSyntheticCaptureStore $published.Directory).Record.SessionLabel -ceq 's') 'CallerMutationCannotChangeArchive'
    $p=Fixture-Packet;$p.ExecutableDeclaration='throw must-not-run';$p.OriginDeclaration='$(throw must-not-evaluate)';$archive=Fixture-Write (New-SyntheticCaptureRecord $p)
    [IO.File]::WriteAllText((Join-Path $archive.Directory 'extra.ps1'),'throw must-not-run')
    $r=Read-OfflineSyntheticCaptureStore $archive.Directory
    Check-CaptureStore ($r.Verified -and $r.Record.ExecutableDeclaration -ceq 'throw must-not-run' -and $r.Record.OriginDeclaration -ceq '$(throw must-not-evaluate)') 'ArchiveTextNeverExecuted'
    Check-CaptureStore $true 'SyntheticCaptureStoreFormatComplete'
}
if($Group -in @('boundary','all','admission','byte-bound','event-bound','reader-paths')){
    if(-not (Test-Path -LiteralPath $root)){New-Item -ItemType Directory -Path $root -ErrorAction Stop|Out-Null}
    if($Group -in @('boundary','all','admission')){ # PARTITION admission
    foreach($bad in @($null,(New-SyntheticCaptureRecord $null),@{},'record')){Fixture-Refuse $bad 'PreflightCreatesNoDirectory'}
    $p=Fixture-Packet;$p.EndedUTC='2026-10-16T23:59:59.9999999Z';Fixture-Refuse (New-SyntheticCaptureRecord $p) 'ClockReversalAndDescriptorRefused'
    foreach($mutation in @('policy-null','origin-over','mode','flag','shape','derived','binding','snapshot','event-count','event-byte-bound','whole-trace-bound')){
        $r=Fixture-Record
        switch($mutation){policy-null{$r.Policy+=[char]0} origin-over{$r.OriginDeclaration='x'*1025} mode{$r.SourceMode='live'} flag{$r.Acceptance=$true} shape{$r|Add-Member Extra 1} derived{$r.Lifecycle.Status='complete'} binding{$r.BindingSHA256='A'*64} snapshot{$r.Arguments.ByteLength=1} event-count{$r.Events=New-Object object[] 4097} event-byte-bound{$r.Events=[object[]]@([pscustomobject]@{Kind='chunk';ElapsedMs=[long]0;Channel='stdout';Data=[pscustomobject]@{Base64=('A'*87384);ByteLength=65538;SHA256=('A'*64)}})} whole-trace-bound{$data=[pscustomobject]@{Base64=[Convert]::ToBase64String((New-Object byte[] 65536));ByteLength=65536;SHA256=[BitConverter]::ToString((Fixture-Hash (New-Object byte[] 65536))).Replace('-','')};$r.Events=[object[]]@(for($i=0;$i -lt 33;$i++){[pscustomobject]@{Kind='chunk';ElapsedMs=[long]0;Channel='stdout';Data=$data}})}}
        Fixture-Refuse $r ('ExactTokensAndShapes-'+$mutation)
    }
    foreach($target in @('record','lifecycle','arguments','wrong-container')){
        $r=Fixture-Record;$script:fixtureCallbacks=0;$obj=if($target -eq 'record'){$r}elseif($target -eq 'lifecycle'){$r.Lifecycle}elseif($target -eq 'arguments'){$r.Arguments}else{@{}}
        Add-Member -InputObject $obj -MemberType ScriptMethod -Name GetType -Force -Value {$script:fixtureCallbacks++;[Management.Automation.PSCustomObject]}
        if($target -eq 'wrong-container'){Fixture-Refuse $obj 'MethodsAndGettersNeverInvoked-wrong-container'}else{$actual=Fixture-Write $r;Check-CaptureStore ($actual.Result.Verified -and $script:fixtureCallbacks -eq 0) ('MethodsAndGettersNeverInvoked-'+$target)}
    }
    $r=Fixture-Record;$r.PSObject.Properties.Remove('SessionLabel');$r|Add-Member ScriptProperty SessionLabel {$script:fixtureCallbacks++;'s'};Fixture-Refuse $r 'MethodsAndGettersNeverInvoked-script-property'
    Add-Type -TypeDefinition 'using System.Management.Automation; public static class SyntheticStoreFixtureGetter { public static int Count; public static string Read(PSObject value) { Count++; return "s"; } }'
    $r=Fixture-Record;$r.PSObject.Properties.Remove('SessionLabel');$r|Add-Member -MemberType CodeProperty -Name SessionLabel -Value ([SyntheticStoreFixtureGetter].GetMethod('Read'))
    Fixture-Refuse $r 'MethodsAndGettersNeverInvoked-code-property';if([SyntheticStoreFixtureGetter]::Count -ne 0){throw 'Code getter invoked'}
    foreach($target in @('events','origin','base64','scalar-getter')){
        $r=Fixture-Record;$script:fixtureCallbacks=0
        if($target -eq 'events'){Add-Member -InputObject $r.Events -MemberType NoteProperty -Name Rank -Value 1 -Force;$obj=$r.Events;$member='Rank'}
        elseif($target -eq 'origin'){$r.OriginDeclaration=Add-Member -InputObject ('x'*1025) -MemberType NoteProperty -Name Length -Value 1 -Force -PassThru;$obj=$r.OriginDeclaration;$member='Length'}
        elseif($target -eq 'base64'){$r.Arguments.Base64=Add-Member -InputObject $r.Arguments.Base64 -MemberType NoteProperty -Name Length -Value 0 -Force -PassThru;$obj=$r.Arguments.Base64;$member='Length'}
        else{$r.Policy=Add-Member -InputObject $r.Policy -MemberType ScriptProperty -Name Extra -Value {$script:fixtureCallbacks++;'unused'} -PassThru;$obj=$r.Policy;$member='Extra'}
        try{Fixture-Refuse $r ('LengthRankShadowsRefused-'+$target)}finally{$obj.PSObject.Properties.Remove($member)}
    }
    } # PARTITION admission end
    if($Group -in @('boundary','all','byte-bound')){ # PARTITION byte-bound
    # Exact large trace/channel/argument/text bounds;one roundtrip,not a performance test.
    $p=Fixture-Packet;$p.SessionLabel='s'*80;$p.ExecutableDeclaration='x'*1024;$p.OriginDeclaration='o'*1024;$p.ArgumentBytes=New-Object byte[] 65536
    $chunks=New-Object 'System.Collections.Generic.List[object]';for($i=0;$i -lt 16;$i++){foreach($c in @('stdout','stderr')){$chunks.Add((Fixture-Chunk $c (New-Object byte[] 65536)))}}
    $p.Events=[object[]]$chunks.ToArray();$r=(Fixture-Write (New-SyntheticCaptureRecord $p)).Result.Record
    Check-CaptureStore ($r.Arguments.ByteLength -eq 65536 -and $r.Events.Count -eq 32 -and $r.Lifecycle.Stdout.ByteLength -eq 1048576 -and $r.Lifecycle.Stderr.ByteLength -eq 1048576 -and $r.SessionLabel.Length -eq 80 -and $r.OriginDeclaration.Length -eq 1024) 'ExactFieldAndTraceBounds'
    } # PARTITION byte-bound end
    if($Group -in @('boundary','all','event-bound')){ # PARTITION event-bound
    $p=Fixture-Packet;$p.Events=[object[]]@(for($i=0;$i -lt 4096;$i++){Fixture-Tick});$r=(Fixture-Write (New-SyntheticCaptureRecord $p)).Result.Record
    Check-CaptureStore ($r.Events.Count -eq 4096 -and $r.Lifecycle.AcceptedEventCount -eq 4096) 'ExactEventCountBound'
    } # PARTITION event-bound end
    if($Group -in @('boundary','all','reader-paths')){ # PARTITION reader-paths
    # Byte offsets derived from fixed hand-encoded prefix,not production encoder.
    $utcBool=$baseHead.Length-5;$eventCount=$baseHead.Length-4;$lc=$baseHead.Length
    foreach($mutation in @('magic','version','negative-text','huge-text','utf8','bool','count','count-overflow','presence','snapshot-negative','snapshot-oversize','tail','truncate','binding','derived','authority','total-over','total-exact-invalid')){
        $bytes=[byte[]]$literal.Clone()
        switch($mutation){magic{$bytes[0]=0} version{$bytes[8]=2} negative-text{[Array]::Copy([BitConverter]::GetBytes([int]-1),0,$bytes,16,4)} huge-text{[Array]::Copy([BitConverter]::GetBytes([int]::MaxValue),0,$bytes,16,4)} utf8{$bytes[20]=255} bool{$bytes[$utcBool]=2} count{[Array]::Copy([BitConverter]::GetBytes([int]-1),0,$bytes,$eventCount,4)} count-overflow{[Array]::Copy([BitConverter]::GetBytes([int]::MaxValue),0,$bytes,$eventCount,4)} presence{$bytes[$lc+69]=2} snapshot-negative{[Array]::Copy([BitConverter]::GetBytes([int]-1),0,$bytes,72,4)} snapshot-oversize{[Array]::Copy([BitConverter]::GetBytes([int]65537),0,$bytes,72,4)} tail{$bytes=[byte[]]($bytes+0)} truncate{$bytes=$bytes[0..($bytes.Length-2)]} binding{$bytes[$bytes.Length-4]=$bytes[$bytes.Length-4] -bxor 1} derived{$bytes[$lc+65]=1} authority{$bytes[$bytes.Length-1]=1} total-over{$bytes=New-Object byte[] 8388609} total-exact-invalid{$bytes=New-Object byte[] 8388608}}
        $name=if($mutation -in @('derived','authority','binding')){'ValidHashInvalidDerivedStateRefused-'+$mutation}else{'MalformedPrimitiveRefused-'+$mutation}
        Fixture-ReadRefuse (Fixture-Put $bytes (Fixture-Marker $bytes)) $name
    }
    foreach($mutation in @('marker-short','marker-tail','marker-hash','marker-length','marker-version','stopped','missing-record','missing-marker')){
        $marker=Fixture-Marker $literal
        switch($mutation){marker-short{$marker=$marker[0..50]} marker-tail{$marker=[byte[]]($marker+0)} marker-hash{$marker[20]=$marker[20] -bxor 1} marker-length{$marker[12]=0} marker-version{$marker[8]=2}}
        $d=Fixture-Path;New-Item -ItemType Directory -Path $d -ErrorAction Stop|Out-Null
        if($mutation -ne 'missing-record'){[IO.File]::WriteAllBytes((Join-Path $d 'record.bin'),$literal)}
        if($mutation -ne 'missing-marker'){[IO.File]::WriteAllBytes((Join-Path $d 'complete.bin'),$marker)}
        if($mutation -eq 'stopped'){[IO.File]::WriteAllBytes((Join-Path $d 'stopped'),[byte[]]@(1))}
        $name=if($mutation -eq 'stopped'){'StoppedOverridesCompleteMarker'}else{'PartialArchiveNeverPublishesRecord-'+$mutation}
        Fixture-ReadRefuse $d $name
    }
    $d=Fixture-Put $literal (Fixture-Marker $literal);$refused=$false;try{$null=Write-OfflineSyntheticCaptureStore $d (Fixture-Record)}catch{$refused=$true}
    Check-CaptureStore ($refused -and (Fixture-Equal ([Convert]::ToBase64String([IO.File]::ReadAllBytes((Join-Path $d 'record.bin')))) ([Convert]::ToBase64String($literal))) -and -not (Test-Path -LiteralPath (Join-Path $d 'stopped'))) 'ExistingTargetNeverOverwritten'
    foreach($unsafe in @((Join-Path $root '../escape'),'\\localhost\share\fixture',($d+':stream'),(Join-Path $d 'nested'),($d+'.'))){Fixture-ReadRefuse $unsafe 'UnsafePathsAndLinksRefused-read';$rejected=$false;try{$null=Write-OfflineSyntheticCaptureStore $unsafe (Fixture-Record)}catch{$rejected=$true};Check-CaptureStore $rejected 'UnsafePathsAndLinksRefused-write'}
    foreach($leaf in @('record.bin','complete.bin')){$d=Fixture-Put $literal (Fixture-Marker $literal);New-Item -ItemType HardLink -Path (Join-Path $root ($prefix+'-link-'+$leaf)) -Target (Join-Path $d $leaf) -ErrorAction Stop|Out-Null;Fixture-ReadRefuse $d ('UnsafePathsAndLinksRefused-hardlink-'+$leaf)}
    $d=Fixture-Put $literal (Fixture-Marker $literal);$junction=Fixture-Path;New-Item -ItemType Junction -Path $junction -Target $d -ErrorAction Stop|Out-Null;Fixture-ReadRefuse $junction 'UnsafePathsAndLinksRefused-junction'
    $held=[IO.File]::Open((Join-Path $d 'record.bin'),[IO.FileMode]::Open,[IO.FileAccess]::Write,[IO.FileShare]::None);try{Fixture-ReadRefuse $d 'ExclusiveWriterRefused'}finally{$held.Dispose()}
    $old=[byte[]]$literal.Clone();[Array]::Copy((Fixture-Hex '544C524F45563031'),0,$old,0,8);Fixture-ReadRefuse (Fixture-Put $old (Fixture-Marker $old)) 'OldNewFormatCrossReadRefused-new-reader'
    . (Join-Path $PSScriptRoot 'output-evidence-store.ps1');$oldResult=Read-OfflineOutputEvidenceStore $d
    Check-CaptureStore (-not $oldResult.Verified -and $null -eq $oldResult.Package) 'OldNewFormatCrossReadRefused-old-reader-path'
    } # PARTITION reader-paths end
    if($Group -in @('boundary','all')){ # PARTITION legacy completion
    Check-CaptureStore $true 'SyntheticCaptureStoreBoundaryComplete'
    }else{ # PARTITION selected completion
        Check-CaptureStore $true ('SyntheticCaptureStorePartitionComplete-'+$Group) # PARTITION selected completion label
    } # PARTITION completion end
}
Write-Output 'LIMIT: synthetic local archive only;no live capture,origin trust,atomicity,power-loss guarantee or acceptance.'
