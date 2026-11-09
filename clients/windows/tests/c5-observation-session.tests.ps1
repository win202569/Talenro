#requires -Version 7.0
$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$script:count=0
function Check($ok,$label){if(-not $ok){throw "ASSERT: $label"};$script:count++;"PASS: $label"}
$module=Join-Path $PSScriptRoot 'c5-observation-session.psm1'
if(-not(Test-Path $module)){throw 'ASSERT: C5 orchestration missing'}
Import-Module $module -Force
Import-Module (Join-Path $PSScriptRoot 'native-process-observation.psm1') -Force
Import-Module (Join-Path $PSScriptRoot 'process-observation.psm1') -Force
function New-Fixture([string]$fault='') {
    $s=@{Fault=$fault;Now=[double]0;NativeCount=0;Launches=0;Calls=[Collections.Generic.List[string]]::new();Rows=[Collections.Generic.List[object]]::new();Progress=[Collections.Generic.List[object]]::new();Sleeps=[Collections.Generic.List[int]]::new()}
    $s.Start=[datetime]::Parse('2026-11-09T18:39:47.4427247Z').ToUniversalTime()
    $s.Context=@{RunId='11111111111141118111111111111111';ExpectedPath='D:\fixture\demo.exe';ExpectedCommandLine='"D:\fixture\demo.exe" --layout-test --layout-language=zh-Hans --layout-profile=unknown-usage'}
    $s.Handle=[object]::new()
    $s.Writer=[pscustomobject]@{State=$s;Disposed=$false;Stream=[IO.MemoryStream]::new()}
    $s.Writer | Add-Member ScriptMethod WriteLine {param($line) if($this.State.Fault -eq 'write'){throw 'private-write'};$this.State.Rows.Add(($line|ConvertFrom-Json -DateKind String));$this.Stream.WriteByte(1)}
    $s.Writer | Add-Member ScriptMethod Dispose {$this.Disposed=$true;$this.Stream.Dispose();if($this.State.Fault -eq 'dispose'){throw 'private-dispose'}}
    $s.Process=[pscustomobject]@{Disposed=$false;Stream=[IO.MemoryStream]::new()}
    $s.Process | Add-Member ScriptMethod Dispose {$this.Disposed=$true;$this.Stream.Dispose()}
    $s.Backend=@{
        Preflight={param($c) $s.Calls.Add('preflight');if($s.Fault -eq 'preflight'){throw 'private-preflight'}}.GetNewClosure()
        AcquireWriter={param($c) $s.Calls.Add('writer');if($s.Fault -eq 'writer'){throw 'private-writer'};return $s.Writer}.GetNewClosure()
        AcquireProcess={param($c) $s.Calls.Add('launch');$s.Launches++;if($s.Fault -eq 'launch'){throw 'private-launch'};return $s.Process}.GetNewClosure()
        GetIdentity={param($p) if(-not[object]::ReferenceEquals($p,$s.Process)){throw 'wrong process'};$s.Calls.Add('identity');if($s.Fault -eq 'identity'){throw 'private-identity'};@{ProcessId=123;StartUtc=$s.Start;Handle=$s.Handle}}.GetNewClosure()
        ReadNative={param($h,$id)
            if(-not[object]::ReferenceEquals($h,$s.Handle)){throw 'wrong handle'}
            $s.Calls.Add('native');$s.NativeCount++
            if($s.Fault -eq 'native'){throw 'private-native'}
            $done=$s.NativeCount -ge 3 -and $s.Fault -notin @('timeout','near-cap')
            if($s.Fault -eq 'overrun' -and $s.NativeCount -ge 2){$s.Now=360001}
            $ft=$s.Start.ToFileTimeUtc(); if($s.Fault -eq 'native-time'){$ft++}
            $exit=if($done){$ft+10000}else{[long]0}
            $times=@{Creation=@{High=[uint32]($ft -shr 32);Low=[uint32]($ft -band 4294967295L)};Exit=@{High=[uint32]($exit -shr 32);Low=[uint32]($exit -band 4294967295L)};Kernel=@{High=[uint32]0;Low=[uint32]0};User=@{High=[uint32]0;Low=[uint32]0}}
            if($s.Fault -eq 'zero-exit'){$times.Exit=@{High=[uint32]0;Low=[uint32]0}}
            $code=if($done){[uint32]0}else{[uint32]259};if($s.Fault -eq 'nonzero' -and $done){$code=[uint32]7}
            $readers=@{Wait={@{RawResult=[uint32]$(if($done){0}else{258})}}.GetNewClosure();ExitCode={@{ApiSucceeded=$true;ExitCode=$code}}.GetNewClosure();Times={@{ApiSucceeded=$true;Times=$times}}.GetNewClosure()}
            $n=Get-TalenroSyntheticNativeProcessObservation -Readers $readers -CorrelationId $id
            if($s.Fault -eq 'api'){$n.Calls.Times.Status='invalid'}
            return $n
        }.GetNewClosure()
        ReadCim={param($id) $s.Calls.Add('cim');if($s.Fault -eq 'cim'){throw 'private-cim'}
            @{ProcessId=$(if($s.Fault -eq 'pid'){124}else{123});ExecutablePath=$(if($s.Fault -eq 'path'){'D:\other.exe'}else{$s.Context.ExpectedPath});CommandLine=$(if($s.Fault -eq 'args'){$s.Context.ExpectedCommandLine+'-extra'}else{$s.Context.ExpectedCommandLine});CreationUtc=$s.Start.AddTicks($(if($s.Fault -eq 'cim-time'){3}else{-7}))}
        }.GetNewClosure()
        ReadManaged={param($p) if(-not[object]::ReferenceEquals($p,$s.Process)){throw 'wrong managed object'};$s.Calls.Add('managed');if($s.Fault -eq 'managed'){throw 'private-managed'}
            $done=$s.NativeCount -ge 3
            $fixtureStart=$s.Start; $disagree=$s.Fault -eq 'disagree'
            $r=@{ProcessId={[int]123};StartTime={$fixtureStart}.GetNewClosure();WaitSignaled={ [bool]($done -and -not $disagree) }.GetNewClosure();HasExited={[bool]$done}.GetNewClosure();ExitCode={[int]$(if($done){0}else{259})}.GetNewClosure();ExitTime={if($done){$fixtureStart.AddTicks(10000)}else{throw 'not exited'}}.GetNewClosure()}
            Get-TalenroSyntheticProcessObservation -Readers $r
        }.GetNewClosure()
        ElapsedMs={if($s.Fault -eq 'clock'){return [double]::NaN};if($s.Fault -eq 'near-cap' -and $s.NativeCount -eq 2 -and $s.Now -eq 0){$s.Now=359500};return $s.Now}.GetNewClosure()
        SleepMs={param($ms) $s.Sleeps.Add($ms);$s.Now+=$ms;if($s.Sleeps.Count -gt 181){throw 'loop unbounded'}}.GetNewClosure()
        Publish={param($r) $s.Progress.Add($r);if($s.Fault -eq 'publish'){throw 'private-publish'}}.GetNewClosure()
    }
    return $s
}
function Run-Fixture($s){Invoke-TalenroC5Observation -Mode Synthetic -Context $s.Context -Backend $s.Backend}
$s=New-Fixture; $r=Run-Fixture $s
Check ($r.Outcome -ceq 'exited' -and $r.ExitVerified) 'valid agreeing exit'
Check ($s.Launches -eq 1 -and $s.Process.Disposed -and $s.Writer.Disposed) 'single acquisition and disposal'
Check (($s.Calls -join ',') -ceq 'preflight,writer,launch,identity,native,cim,native,managed,native,managed') 'identity then ordered same-object pairs'
Check ($s.Rows[1].Kind -ceq 'native-identity' -and -not $s.Rows[1].IdentityVerified) 'initial native receipt not mislabeled'
Check (@($s.Rows|Where-Object {$_.Mode -cne 'Synthetic' -or $_.RunId -cne $s.Context.RunId}).Count -eq 0) 'all receipts synthetic and correlated'
Check (($s.Rows.Sequence -join ',') -ceq '1,2,3,4,5') 'contiguous sequence'
foreach($fault in @('preflight','writer','launch','identity','native','cim','pid','path','args','native-time','cim-time','api','managed','write','publish','dispose','clock')){
    $s=New-Fixture $fault;$failed=$false;$out=@()
    try{$out=@(Run-Fixture $s)}catch{$failed=$true}
    Check ($failed -and $out.Count -eq 0) "fail closed: $fault"
    Check ($s.Launches -le 1) "no retry: $fault"
    if($s.Calls.Contains('writer') -and $fault -ne 'writer'){Check $s.Writer.Disposed "writer cleanup: $fault"}
    if($s.Launches -eq 1 -and $fault -ne 'launch'){Check $s.Process.Disposed "process cleanup: $fault"}
    if($fault -in @('cim','pid','path','args','native-time','cim-time','api')){Check (@($s.Rows|Where-Object Kind -eq 'native-identity').Count -eq 1 -and -not $s.Calls.Contains('managed')) "identity evidence preserved without pair: $fault"}
    Check ([string]($s.Rows|ConvertTo-Json -Depth 20) -notmatch 'private-') "no private error payload: $fault"
}
foreach($fault in @('nonzero','zero-exit','disagree')){$s=New-Fixture $fault;$r=Run-Fixture $s;Check ($r.Outcome -ceq 'exited' -and -not $r.ExitVerified) "not verified exit: $fault"}
foreach($fault in @('timeout','near-cap','overrun')){$s=New-Fixture $fault;$r=Run-Fixture $s;Check ($r.Outcome -ceq 'timed-out' -and -not $r.ExitVerified) "timeout: $fault";Check (@($s.Sleeps|Where-Object {$_ -gt 2000 -or $_ -le 0}).Count -eq 0) "bounded sleeps: $fault";if($fault -eq 'near-cap'){Check ($s.Sleeps[0] -eq 500) 'sleep clipped to remaining budget'}}
$s=New-Fixture;$s.Backend.Remove('ReadNative');$failed=$false;try{Run-Fixture $s}catch{$failed=$true};Check ($failed -and $s.Calls.Count -eq 0) 'missing backend rejects without fallback'
foreach($bad in @([double]-1,[double]::PositiveInfinity,'0')){
    $s=New-Fixture;$s.Backend.ElapsedMs={return $bad}.GetNewClosure();$failed=$false
    try{Run-Fixture $s}catch{$failed=$true};Check ($failed -and $s.NativeCount -eq 0) 'invalid clock stops before native calls'
}
$s=New-Fixture;$ticks=@{N=0};$s.Backend.ElapsedMs={$ticks.N++;if($ticks.N -eq 1){[double]10}else{[double]9}}.GetNewClosure()
$failed=$false;try{Run-Fixture $s}catch{$failed=$true};Check ($failed -and $s.NativeCount -eq 0) 'decreasing clock rejected'
$s=New-Fixture;$s.Now=[double]360000;$r=Run-Fixture $s;Check ($r.Outcome -eq 'timed-out' -and $s.NativeCount -eq 0 -and $s.Sleeps.Count -eq 0) 'exact cap no sample or sleep'
$s=New-Fixture;$s.Backend.Extra={throw 'must not run'};$failed=$false;try{Run-Fixture $s}catch{$failed=$true};Check ($failed -and $s.Calls.Count -eq 0) 'extra backend callback rejected'
Check ($null -eq ('Talenro.Diagnostics.NativeObservationV1.NativeProbe' -as [type])) 'synthetic tests never loaded native probe'
"C5_SYNTHETIC_PASS: $script:count assertions"
