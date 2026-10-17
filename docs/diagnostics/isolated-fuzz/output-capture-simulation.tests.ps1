$ErrorActionPreference='Stop'
function Check-Capture([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
$helper=Join-Path $PSScriptRoot 'output-capture-simulation.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
Check-Capture ($null -ne (Get-Command Get-OfflineCaptureSimulation -CommandType Function -ErrorAction SilentlyContinue)) 'OfflineCaptureSimulationRequired'
Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
function Capture-Chunk($Channel,$Bytes,$Time=0){[pscustomobject]@{Kind='chunk';ElapsedMs=$Time;Channel=$Channel;Bytes=$Bytes}}
function Capture-EOF($Channel,$Time=0){[pscustomobject]@{Kind='eof';ElapsedMs=$Time;Channel=$Channel}}
function Capture-Exit($Code=0,$Time=0){[pscustomobject]@{Kind='exit';ElapsedMs=$Time;ExitCode=$Code}}
function Capture-Tick($Time=0){[pscustomobject]@{Kind='tick';ElapsedMs=$Time}}
function Invoke-Capture($Events,$Encoding='utf-8',$Origin='synthetic-only'){
    $r=Get-OfflineCaptureSimulation -Events $Events -EncodingName $Encoding -OriginDeclaration $Origin
    $names=@('Schema','Policy','Status','Reason','AcceptedEventCount','FailureEventIndex','LastObservedElapsedMs','StdoutEOF','StderrEOF','ExitObserved','ExitCode','Stdout','Stderr','Package','OriginVerified','Acceptance','RuntimeApproved')
    if(@($r.PSObject.Properties).Count -ne $names.Count -or @($r.PSObject.Properties|Where-Object{$_.Name -cnotin $names -or $_.MemberType -ne 'NoteProperty'}).Count -ne 0){throw 'Unexpected result fields'}
    if($r.Schema -ne 1 -or $r.Policy -cne 'offline-capture-simulation-v1' -or $r.Status -cnotin @('complete','incomplete','timed-out','refused')){throw 'Unexpected result schema/status'}
    foreach($flag in @('StdoutEOF','StderrEOF','ExitObserved','OriginVerified','Acceptance','RuntimeApproved')){if($r.$flag -isnot [bool]){throw 'Boolean flag required'}}
    if($r.OriginVerified -or $r.Acceptance -or $r.RuntimeApproved){throw 'Simulation granted authority'}
    foreach($stream in @($r.Stdout,$r.Stderr)){
        $b=[Convert]::FromBase64String($stream.Base64);$sha=[Security.Cryptography.SHA256]::Create()
        try{$hash=[BitConverter]::ToString($sha.ComputeHash($b)).Replace('-','')}finally{$sha.Dispose()}
        if($stream.ByteLength -ne $b.Length -or $stream.SHA256 -cne $hash -or $b.Length -gt 1048576){throw 'Snapshot metadata mismatch'}
    }
    if($r.Status -ceq 'complete'){if($null -eq $r.Package -or -not (Test-OfflineOutputEvidencePackage $r.Package).Verified){throw 'Complete package missing/inconsistent'}}elseif($null -ne $r.Package){throw 'Incomplete package published'}
    return $r
}
function Check-CaptureState($R,$Status,$Reason,$Count,$Index,$Time,$OutEOF=$false,$ErrEOF=$false,$ExitSeen=$false,$Code=$null){
    Check-Capture ($R.Status -ceq $Status -and $R.Reason -ceq $Reason -and $R.AcceptedEventCount -eq $Count -and $R.FailureEventIndex -eq $Index -and $R.LastObservedElapsedMs -eq $Time -and $R.StdoutEOF -eq $OutEOF -and $R.StderrEOF -eq $ErrEOF -and $R.ExitObserved -eq $ExitSeen -and $R.ExitCode -eq $Code) ('State-'+$Reason+'-'+$Count+'-'+$Index)
}
# Counter-only reproduction:ETS dispatch must not invoke caller GetType methods.
$script:captureOuterTypeCalls=0;$script:captureEventTypeCalls=0
$outerWithMethod=[pscustomobject]@{Label='not-an-array'}
$outerWithMethod|Add-Member -MemberType ScriptMethod -Name GetType -Force -Value {$script:captureOuterTypeCalls++;[object[]]}
$eventWithMethod=Capture-Tick
$eventWithMethod|Add-Member -MemberType ScriptMethod -Name GetType -Force -Value {$script:captureEventTypeCalls++;[System.Management.Automation.PSCustomObject]}
$outerResult=Invoke-Capture $outerWithMethod
$eventResult=Invoke-Capture ([object[]]@($eventWithMethod))
Check-Capture ($script:captureOuterTypeCalls -eq 0 -and $script:captureEventTypeCalls -eq 0) 'InputGetTypeScriptMethodsNeverInvoked'
Check-CaptureState $outerResult 'refused' 'invalid-input' 0 $null 0
Check-CaptureState $eventResult 'incomplete' 'incomplete' 1 $null 0
# Exact outer runtime type must still refuse covariant string[] rather than coerce.
Check-CaptureState (Invoke-Capture ([string[]]@('tick'))) 'refused' 'invalid-input' 0 $null 0
$abc=[byte[]]@(97,98,99);$empty=[byte[]]@()
$base=[object[]]@((Capture-Chunk 'stdout' $abc),(Capture-Exit 0 1),(Capture-EOF 'stdout' 2),(Capture-EOF 'stderr' 3))
$r=Invoke-Capture $base
Check-CaptureState $r 'complete' 'complete' 4 $null 3 $true $true $true 0
Check-Capture ($r.Stdout.Base64 -ceq 'YWJj' -and $r.Stdout.ByteLength -eq 3 -and $r.Stdout.SHA256 -ceq 'BA7816BF8F01CFEA414140DE5DAE2223B00361A396177A9CB410FF61F20015AD') 'LiteralAbcSnapshot'
Check-Capture (-not $r.Package.Acceptance -and -not $r.Package.OriginVerified -and -not $r.Package.RuntimeApproved) 'PackageNeverGrantsAuthority'
$trailingMethod=Capture-Tick 4
$trailingMethod|Add-Member -MemberType ScriptMethod -Name GetType -Force -Value {$script:captureEventTypeCalls++;[System.Management.Automation.PSCustomObject]}
$trailingResult=Invoke-Capture ([object[]]($base+@($trailingMethod)))
Check-CaptureState $trailingResult 'refused' 'trailing-after-complete' 4 4 4 $true $true $true 0
Check-Capture ($script:captureEventTypeCalls -eq 0 -and $trailingResult.Stdout.Base64 -ceq 'YWJj') 'TrailingTypeValidationNeverInvokesCallerMethod'
Check-CaptureState (Invoke-Capture ([object[]]@())) 'incomplete' 'incomplete' 0 $null 0
$items=@((Capture-Exit 0 3),(Capture-EOF 'stdout' 3),(Capture-EOF 'stderr' 3))
foreach($order in @('012','021','102','120','201','210')){
    $seq=[object[]]@($items[[int]::Parse($order.Substring(0,1))],$items[[int]::Parse($order.Substring(1,1))],$items[[int]::Parse($order.Substring(2,1))])
    Check-CaptureState (Invoke-Capture $seq) 'complete' 'complete' 3 $null 3 $true $true $true 0
}
$r=Invoke-Capture ([object[]]@((Capture-Exit 0 1),(Capture-Chunk 'stdout' $abc 2),(Capture-EOF 'stdout' 3),(Capture-EOF 'stderr' 3)))
Check-Capture ($r.Status -ceq 'complete' -and $r.Stdout.Base64 -ceq 'YWJj' -and $r.AcceptedEventCount -eq 4) 'ExitBeforeBufferedChunksStillComplete'
Check-CaptureState (Invoke-Capture ([object[]]@((Capture-Exit 0 1)))) 'incomplete' 'incomplete' 1 $null 1 $false $false $true 0
Check-CaptureState (Invoke-Capture ([object[]]@((Capture-EOF 'stdout' 1),(Capture-EOF 'stderr' 1)))) 'incomplete' 'incomplete' 2 $null 1 $true $true
Check-CaptureState (Invoke-Capture ([object[]]@((Capture-EOF 'stdout' 1),(Capture-Exit 0 2)))) 'incomplete' 'incomplete' 2 $null 2 $true $false $true 0
foreach($time in @(119999,120000,120001)){
    $r=Invoke-Capture ([object[]]@((Capture-Exit 0 1),(Capture-EOF 'stdout' 2),(Capture-EOF 'stderr' $time)))
    if($time -eq 119999){Check-CaptureState $r 'complete' 'complete' 3 $null $time $true $true $true 0}else{Check-CaptureState $r 'timed-out' 'deadline-reached' 2 2 $time $true $false $true 0}
}
$r=Invoke-Capture ([object[]]@((Capture-Chunk 'stdout' $abc),(Capture-Tick 120000),(Capture-Exit 0 120001)))
Check-CaptureState $r 'timed-out' 'deadline-reached' 1 1 120000
Check-Capture ($r.Stdout.Base64 -ceq 'YWJj') 'TimeoutRetainsOnlyAdmittedPrefix'
foreach($tail in @((Capture-Tick 4),(Capture-Tick 120000),$null,(Capture-Tick 2))){
    $r=Invoke-Capture ([object[]]($base+@($tail)))
    $time=3;if($null -ne $tail -and $tail.ElapsedMs -ge 3){$time=$tail.ElapsedMs}
    Check-CaptureState $r 'refused' 'trailing-after-complete' 4 4 $time $true $true $true 0
    Check-Capture ($r.Stdout.Base64 -ceq 'YWJj') 'TrailingDoesNotPublishPrefixPackage'
}
foreach($time in @(0,-1,'1',$false,[double]1)){
    $r=Invoke-Capture ([object[]]@((Capture-Chunk 'stdout' $abc 1),(Capture-Tick $time)))
    $reason='invalid-event';if($time -is [int] -and $time -eq 0){$reason='time-went-backward'}
    Check-CaptureState $r 'refused' $reason 1 1 1
    Check-Capture ($r.Stdout.Base64 -ceq 'YWJj') 'InvalidTimeRetainsPrefix'
}
Check-CaptureState (Invoke-Capture ([object[]]@((Capture-Tick ([long]::MaxValue))))) 'timed-out' 'deadline-reached' 0 0 ([long]::MaxValue)
Check-CaptureState (Invoke-Capture ([object[]]@((Capture-EOF 'stdout'),(Capture-EOF 'stdout')))) 'refused' 'duplicate-eof' 1 1 0 $true
Check-CaptureState (Invoke-Capture ([object[]]@((Capture-Exit 1),(Capture-Exit 0)))) 'refused' 'duplicate-exit' 1 1 0 $false $false $true 1
Check-CaptureState (Invoke-Capture ([object[]]@((Capture-EOF 'stderr'),(Capture-Chunk 'stderr' $abc)))) 'refused' 'chunk-after-eof' 1 1 0 $false $true
$script:captureGetterCalls=0
$scripted=[pscustomobject]@{Kind='tick'};$scripted|Add-Member ScriptProperty ElapsedMs {$script:captureGetterCalls++;0}
$badEvents=@($null,@{Kind='tick';ElapsedMs=0},[pscustomobject]@{Kind='other';ElapsedMs=0},[pscustomobject]@{Kind='TICK';ElapsedMs=0},[pscustomobject]@{kind='tick';ElapsedMs=0},[pscustomobject]@{Kind='tick'},[pscustomobject]@{Kind='tick';ElapsedMs=0;Extra=0},(Capture-EOF 'other'),(Capture-EOF $true),$scripted)
foreach($bad in $badEvents){Check-CaptureState (Invoke-Capture ([object[]]@($bad))) 'refused' 'invalid-event' 0 0 0}
Check-Capture ($script:captureGetterCalls -eq 0) 'ExecutablePropertiesNeverInvoked'
foreach($wrapper in @([pscustomobject]@{Value=$null},[pscustomobject]@{Value=1},[pscustomobject]@{Value=@{}},[pscustomobject]@{Value=[byte[]]@(1)},[pscustomobject]@{Value=(New-Object 'object[,]' 1,1)},[pscustomobject]@{Value=(New-Object object[] 4097)})){
    Check-CaptureState (Invoke-Capture $wrapper.Value) 'refused' 'invalid-input' 0 $null 0
}
$ticks=New-Object object[] 4096;for($i=0;$i -lt 4096;$i++){$ticks[$i]=Capture-Tick}
Check-CaptureState (Invoke-Capture $ticks) 'incomplete' 'incomplete' 4096 $null 0
foreach($meta in @(@{Encoding='guess';Origin='x'},@{Encoding=1;Origin='x'},@{Encoding='utf-8';Origin=''},@{Encoding='utf-8';Origin=('x'*1025)},@{Encoding='utf-8';Origin=1},@{Encoding='utf-8';Origin=([string][char]0xD800)})){
    Check-CaptureState (Invoke-Capture $base $meta.Encoding $meta.Origin) 'refused' 'invalid-input' 0 $null 0
}
foreach($code in @(-1,[int]::MinValue,[int]::MaxValue)){
    $r=Invoke-Capture ([object[]]@((Capture-Exit $code),(Capture-EOF 'stdout'),(Capture-EOF 'stderr')))
    Check-CaptureState $r 'complete' 'complete' 3 $null 0 $true $true $true $code
    Check-Capture ($r.Package.Classification -ceq 'refused') 'NonzeroExitDoesNotGrantOutputSuccess'
}
foreach($code in @([long]2147483648,'0',$false,[double]0)){
    Check-CaptureState (Invoke-Capture ([object[]]@((Capture-Exit $code)))) 'refused' 'invalid-event' 0 0 0
}
foreach($b in @([pscustomobject]@{Value=$empty},[pscustomobject]@{Value=(New-Object byte[] 65536)})){
    $r=Invoke-Capture ([object[]]@((Capture-Chunk 'stdout' $b.Value)))
    Check-CaptureState $r 'incomplete' 'incomplete' 1 $null 0
    Check-Capture ($r.Stdout.ByteLength -eq $b.Value.Length) 'ValidChunkBoundsRetained'
}
foreach($b in @([pscustomobject]@{Value=$null},[pscustomobject]@{Value='abc'},[pscustomobject]@{Value=[object[]]@(97,98,99)},[pscustomobject]@{Value=(New-Object byte[] 65537)})){
    $r=Invoke-Capture ([object[]]@((Capture-Chunk 'stdout' $b.Value)))
    Check-CaptureState $r 'refused' 'invalid-event' 0 0 0
    Check-Capture ($r.Stdout.ByteLength -eq 0) 'InvalidChunkNeverClippedOrCoerced'
}
$both=New-Object 'System.Collections.Generic.List[object]'
for($i=0;$i -lt 16;$i++){foreach($ch in @('stdout','stderr')){$both.Add((Capture-Chunk $ch (New-Object byte[] 65536)))}}
$r=Invoke-Capture ([object[]]$both.ToArray())
Check-CaptureState $r 'incomplete' 'incomplete' 32 $null 0
Check-Capture ($r.Stdout.ByteLength -eq 1048576 -and $r.Stderr.ByteLength -eq 1048576) 'IndependentExactStreamCaps'
$zeros=New-Object byte[] 1048576;$expectedZeroBase64=[Convert]::ToBase64String($zeros)
foreach($ch in @('stdout','stderr')){
    $r=Invoke-Capture ([object[]]($both.ToArray()+@((Capture-Chunk $ch ([byte[]]@(1))))))
    Check-CaptureState $r 'refused' 'stream-limit' 32 32 0
    Check-Capture ($r.Stdout.Base64 -ceq $expectedZeroBase64 -and $r.Stderr.Base64 -ceq $expectedZeroBase64) 'OverLimitChunkRejectedWhole'
}
foreach($fixture in @(@{Encoding='utf-8';Parts=@([byte[]]@(228),[byte[]]@(184),[byte[]]@(173));Base64='5Lit';Length=3},@{Encoding='utf-16le';Parts=@([byte[]]@(65),[byte[]]@(0,45),[byte[]]@(78));Base64='QQAtTg==';Length=4})){
    $seq=@();foreach($part in $fixture.Parts){$seq+=Capture-Chunk 'stdout' $part};$seq+=Capture-Exit;$seq+=Capture-EOF 'stdout';$seq+=Capture-EOF 'stderr'
    $r=Invoke-Capture ([object[]]$seq) $fixture.Encoding
    Check-Capture ($r.Status -ceq 'complete' -and $r.Stdout.Base64 -ceq $fixture.Base64 -and $r.Stdout.ByteLength -eq $fixture.Length -and $r.Package.EncodingName -ceq $fixture.Encoding -and $r.Package.Classification -ceq 'empty') 'SplitCodepointBytesReassembledBeforeDecoding'
}
$xml='#< CLIXML'+"`n"+'<Objs Version="1.1.0.1" xmlns="http://schemas.microsoft.com/powershell/2004/04"><Obj S="progress" RefId="0"><TN RefId="0"><T>System.Management.Automation.PSCustomObject</T><T>System.Object</T></TN><MS><I64 N="SourceId">1</I64><PR N="Record"><AV>fixture</AV><AI>0</AI><Nil /><PI>-1</PI><PC>-1</PC><T>Completed</T><SR>-1</SR><SD> </SD></PR></MS></Obj></Objs>'
foreach($encoding in @('utf-8','utf-16le')){
    $codec=if($encoding -eq 'utf-8'){[Text.Encoding]::UTF8}else{[Text.Encoding]::Unicode};$b=$codec.GetBytes($xml)
    $seq=[object[]]@((Capture-Chunk 'stderr' ([byte[]]$b[0..2])),(Capture-Chunk 'stderr' ([byte[]]$b[3..($b.Length-1)])),(Capture-Exit),(Capture-EOF 'stdout'),(Capture-EOF 'stderr'))
    $r=Invoke-Capture $seq $encoding
    Check-Capture ($r.Status -ceq 'complete' -and $r.Stderr.Base64 -ceq [Convert]::ToBase64String($b) -and $r.Package.Classification -ceq 'progress-only' -and $r.Package.ProgressCount -eq 1) 'SplitProgressPreservedAndClassified'
}
$r=Invoke-Capture ([object[]]@((Capture-Chunk 'stderr' ([byte[]]@(0,255,13,10))),(Capture-Exit),(Capture-EOF 'stdout'),(Capture-EOF 'stderr')))
Check-Capture ($r.Status -ceq 'complete' -and $r.Stderr.Base64 -ceq 'AP8NCg==' -and $r.Package.Classification -ceq 'refused') 'InvalidTextStillHasCompleteCaptureLifecycle'
$owned=[byte[]]@(97,98,99);$seq=[object[]]@((Capture-Chunk 'stdout' $owned),(Capture-Exit),(Capture-EOF 'stdout'),(Capture-EOF 'stderr'))
$r=Invoke-Capture $seq;$owned[0]=0;$seq[0].Bytes=[byte[]]@(1)
Check-Capture ($r.Stdout.Base64 -ceq 'YWJj' -and $r.Package.Stdout.Base64 -ceq 'YWJj') 'PostReturnMutationCannotChangeSnapshots'
Write-Output 'LIMIT: supplied synthetic event model only;no real capture,clock,watchdog,provenance or acceptance.'
